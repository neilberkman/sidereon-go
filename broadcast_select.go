package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// CompareEpoch is one broadcast/precise epoch pair for SISRE comparison.
// Julian-date fields preserve the split representation required by the C ABI.
type CompareEpoch struct {
	BroadcastTJ2000S       float64
	PreciseJDWhole         float64
	PreciseJDFraction      float64
	PrecisePlusJDWhole     float64
	PrecisePlusJDFraction  float64
	PreciseMinusJDWhole    float64
	PreciseMinusJDFraction float64
}

// CompareWindow describes a regular broadcast/precise comparison grid.
type CompareWindow struct {
	BroadcastWindowStartJ2000S float64
	BroadcastWindowEndJ2000S   float64
	PreciseStartJDWhole        float64
	PreciseStartJDFraction     float64
	StepS                      float64
	VelocityHalfS              float64
}

// CompareStats contains C-computed broadcast-vs-precise error statistics.
// Distance fields are metres; Count is the number of compared samples.
type CompareStats struct {
	Count                          int
	Orbit3DRMSM, Orbit3DMaxM       float64
	RadialRMSM, RadialMaxM         float64
	AlongRMSM, AlongMaxM           float64
	CrossRMSM, CrossMaxM           float64
	ClockRMSM, ClockMaxM           float64
	ClockDatumRMSM, ClockDatumMaxM float64
}

// BroadcastComparison owns a native comparison report and must not be copied.
// Close is idempotent and safe to call concurrently with read methods.
type BroadcastComparison struct {
	_      noCopy
	handle *native.BroadcastComparison
}

func publicCompareStats(value native.NativeCompareStats) CompareStats {
	return CompareStats{Count: value.Count, Orbit3DRMSM: value.Orbit3DRMSM, Orbit3DMaxM: value.Orbit3DMaxM, RadialRMSM: value.RadialRMSM, RadialMaxM: value.RadialMaxM, AlongRMSM: value.AlongRMSM, AlongMaxM: value.AlongMaxM, CrossRMSM: value.CrossRMSM, CrossMaxM: value.CrossMaxM, ClockRMSM: value.ClockRMSM, ClockMaxM: value.ClockMaxM, ClockDatumRMSM: value.ClockDatumRMSM, ClockDatumMaxM: value.ClockDatumMaxM}
}

// CompareBroadcast compares selected broadcast and SP3 satellites at explicit
// split-Julian epochs using the native numerical implementation.
func CompareBroadcast(broadcast *BroadcastEphemeris, precise *SP3, satellites []string, epochs []CompareEpoch, velocityHalfS float64) (*BroadcastComparison, error) {
	nativeEpochs := make([]native.NativeCompareEpoch, len(epochs))
	for i, value := range epochs {
		nativeEpochs[i] = native.NativeCompareEpoch{BroadcastTJ2000S: value.BroadcastTJ2000S, PreciseJDWhole: value.PreciseJDWhole, PreciseJDFraction: value.PreciseJDFraction, PrecisePlusJDWhole: value.PrecisePlusJDWhole, PrecisePlusJDFraction: value.PrecisePlusJDFraction, PreciseMinusJDWhole: value.PreciseMinusJDWhole, PreciseMinusJDFraction: value.PreciseMinusJDFraction}
	}
	handle, err := native.CompareBroadcast(nativeHandle(broadcast), nativeSP3Handle(precise), append([]string(nil), satellites...), nativeEpochs, velocityHalfS)
	if err != nil {
		return nil, publicError(err)
	}
	return &BroadcastComparison{handle: handle}, nil
}

// CompareBroadcastWindow compares a regular native comparison window.
func CompareBroadcastWindow(broadcast *BroadcastEphemeris, precise *SP3, satellites []string, window CompareWindow) (*BroadcastComparison, error) {
	handle, err := native.CompareBroadcastWindow(nativeHandle(broadcast), nativeSP3Handle(precise), append([]string(nil), satellites...), native.NativeCompareWindow{BroadcastWindowStartJ2000S: window.BroadcastWindowStartJ2000S, BroadcastWindowEndJ2000S: window.BroadcastWindowEndJ2000S, PreciseStartJDWhole: window.PreciseStartJDWhole, PreciseStartJDFraction: window.PreciseStartJDFraction, StepS: window.StepS, VelocityHalfS: window.VelocityHalfS})
	if err != nil {
		return nil, publicError(err)
	}
	return &BroadcastComparison{handle: handle}, nil
}

func nativeHandle(value *BroadcastEphemeris) *native.BroadcastEphemeris {
	if value == nil {
		return nil
	}
	return value.handle
}

func nativeSP3Handle(value *SP3) *native.SP3 {
	if value == nil {
		return nil
	}
	return value.handle
}

// Close releases the comparison report.
func (r *BroadcastComparison) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Overall returns aggregate statistics for the report.
func (r *BroadcastComparison) Overall() (CompareStats, error) {
	if r == nil || r.handle == nil {
		return CompareStats{}, ErrClosed
	}
	value, err := r.handle.Overall()
	return publicCompareStats(value), publicError(err)
}

// SatelliteCount returns the number of per-satellite rows in the report.
func (r *BroadcastComparison) SatelliteCount() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	value, err := r.handle.SatelliteCount()
	return value, publicError(err)
}

// Satellite returns a copied per-satellite token and its statistics.
func (r *BroadcastComparison) Satellite(index int) (string, CompareStats, error) {
	if r == nil || r.handle == nil {
		return "", CompareStats{}, ErrClosed
	}
	satellite, value, err := r.handle.Satellite(index)
	return satellite, publicCompareStats(value), publicError(err)
}

// BroadcastRecordInfo is the compact record metadata returned by native
// issue-selection and record-list routes.
type BroadcastRecordInfo struct {
	SatelliteID       string
	Message, Issue    uint32
	IssueMessage      uint32
	Week, ToeWeek     uint32
	ToeTOWSeconds     float64
	TocWeek           uint32
	TocTOWSeconds     float64
	SVHealth          float64
	SVAccuracyM       float64
	HasFitInterval    bool
	FitIntervalS      float64
	DefaultGroupDelay float64
	CNAV              BroadcastCNAV
}

func publicRecordInfo(value native.NativeBroadcastRecordInfo) BroadcastRecordInfo {
	return BroadcastRecordInfo{SatelliteID: value.SatelliteID, Message: value.Message, Issue: value.Issue, IssueMessage: value.IssueMessage, Week: value.Week, ToeWeek: value.ToeWeek, ToeTOWSeconds: value.ToeTOWSeconds, TocWeek: value.TocWeek, TocTOWSeconds: value.TocTOWSeconds, SVHealth: value.SVHealth, SVAccuracyM: value.SVAccuracyM, HasFitInterval: value.HasFitInterval, FitIntervalS: value.FitIntervalS, DefaultGroupDelay: value.DefaultGroupDelay, CNAV: broadcastRecordFromNative(native.NativeBroadcastRecord{CNAV: value.CNAV}).CNAV}
}

// RecordCNavCorrection returns one native CNAV signal correction and presence.
func (b *BroadcastEphemeris) RecordCNavCorrection(index int, signal uint32) (float64, bool, error) {
	if b == nil || b.handle == nil {
		return 0, false, ErrClosed
	}
	value, present, err := b.handle.RecordCNavCorrection(index, signal)
	return value, present, publicError(err)
}

// RecordGroupDelay returns one native group-delay term and presence.
func (b *BroadcastEphemeris) RecordGroupDelay(index int, term uint32) (float64, bool, error) {
	if b == nil || b.handle == nil {
		return 0, false, ErrClosed
	}
	value, present, err := b.handle.RecordGroupDelay(index, term)
	return value, present, publicError(err)
}

// RecordsInfo returns compact copied metadata for every native NAV record.
func (b *BroadcastEphemeris) RecordsInfo() ([]BroadcastRecordInfo, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.RecordsInfo()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]BroadcastRecordInfo, len(values))
	for i := range values {
		out[i] = publicRecordInfo(values[i])
	}
	return out, nil
}

// SelectByIssue selects a record by satellite, issue, message family, and
// epoch. The boolean reports C's optional-record presence separately from err.
func (b *BroadcastEphemeris) SelectByIssue(satellite string, issue, message uint32, epochJ2000S float64) (BroadcastRecordInfo, bool, error) {
	if b == nil || b.handle == nil {
		return BroadcastRecordInfo{}, false, ErrClosed
	}
	value, present, err := b.handle.SelectByIssue(satellite, issue, message, epochJ2000S)
	return publicRecordInfo(value), present, publicError(err)
}

// SetNavMessagePreference updates the native navigation-message preference.
func (b *BroadcastEphemeris) SetNavMessagePreference(preference uint32) error {
	if b == nil || b.handle == nil {
		return ErrClosed
	}
	return publicError(b.handle.SetNavMessagePreference(preference))
}

// BroadcastEccentricAnomaly solves Kepler's equation in the native engine.
func BroadcastEccentricAnomaly(meanAnomalyRad, eccentricity float64) (float64, int, error) {
	value, iterations, err := native.EccentricAnomaly(meanAnomalyRad, eccentricity)
	return value, iterations, publicError(err)
}

// ConstellationConstants contains native physical constants used for broadcast
// orbit and clock evaluation. Values use SI units.
type ConstellationConstants struct{ GMM3PerS2, OmegaERadPerS, DTRF float64 }

// ClockOffset contains the native broadcast satellite-clock decomposition.
type ClockOffset struct{ ClockPolynomialS, RelativisticS, GroupDelayS, TotalS float64 }

// OrbitState contains the complete native broadcast orbit evaluation, including
// intermediate Kepler and corrected-orbit values.
type OrbitState struct {
	A, N0, N, TK, MK, EccentricAnomaly                                           float64
	KeplerIterations                                                             int
	SinE, CosE, Nu, Phi, S2, C2, DU, DR, DI, U, R, I, XP, YP, OmegaK, XM, YM, ZM float64
}

// SatelliteState combines an OrbitState and ClockOffset from one C call.
type SatelliteState struct {
	Orbit OrbitState
	Clock ClockOffset
}

// BroadcastSatelliteClockOffset evaluates the native broadcast clock route.
func BroadcastSatelliteClockOffset(clock ClockPolynomial, constants ConstellationConstants, elements KeplerianElements, sinE, tSOWS, tGDS float64) (ClockOffset, error) {
	value, err := native.BroadcastSatelliteClockOffset(native.NativeClockPolynomial{AF0: clock.AF0, AF1: clock.AF1, AF2: clock.AF2, TocSOW: clock.TocSOW}, native.NativeConstellationConstants{GMM3PerS2: constants.GMM3PerS2, OmegaERadPerS: constants.OmegaERadPerS, DTRF: constants.DTRF}, native.NativeKeplerianElements{SqrtA: elements.SqrtA, E: elements.E, M0: elements.M0, DeltaN: elements.DeltaN, Omega0: elements.Omega0, I0: elements.I0, Omega: elements.Omega, OmegaDot: elements.OmegaDot, IDot: elements.IDot, CUC: elements.CUC, CUS: elements.CUS, CRC: elements.CRC, CRS: elements.CRS, CIC: elements.CIC, CIS: elements.CIS, ToeSOW: elements.ToeSOW}, sinE, tSOWS, tGDS)
	return ClockOffset{ClockPolynomialS: value.ClockPolynomialS, RelativisticS: value.RelativisticS, GroupDelayS: value.GroupDelayS, TotalS: value.TotalS}, publicError(err)
}

func nativeBroadcastElements(value KeplerianElements) native.NativeKeplerianElements {
	return native.NativeKeplerianElements{SqrtA: value.SqrtA, E: value.E, M0: value.M0, DeltaN: value.DeltaN, Omega0: value.Omega0, I0: value.I0, Omega: value.Omega, OmegaDot: value.OmegaDot, IDot: value.IDot, CUC: value.CUC, CUS: value.CUS, CRC: value.CRC, CRS: value.CRS, CIC: value.CIC, CIS: value.CIS, ToeSOW: value.ToeSOW}
}

// BroadcastSatellitePositionECEF evaluates the native broadcast orbit route.
func BroadcastSatellitePositionECEF(elements KeplerianElements, constants ConstellationConstants, tSOWS float64, isGeo bool) (OrbitState, error) {
	value, err := native.BroadcastSatellitePosition(nativeBroadcastElements(elements), native.NativeConstellationConstants{GMM3PerS2: constants.GMM3PerS2, OmegaERadPerS: constants.OmegaERadPerS, DTRF: constants.DTRF}, tSOWS, isGeo)
	return publicOrbitState(value), publicError(err)
}

func publicOrbitState(value native.NativeOrbitState) OrbitState {
	return OrbitState{A: value.A, N0: value.N0, N: value.N, TK: value.TK, MK: value.MK, EccentricAnomaly: value.EccentricAnomaly, KeplerIterations: value.KeplerIterations, SinE: value.SinE, CosE: value.CosE, Nu: value.Nu, Phi: value.Phi, S2: value.S2, C2: value.C2, DU: value.DU, DR: value.DR, DI: value.DI, U: value.U, R: value.R, I: value.I, XP: value.XP, YP: value.YP, OmegaK: value.OmegaK, XM: value.XM, YM: value.YM, ZM: value.ZM}
}

// BroadcastSatelliteState evaluates the native combined broadcast state route.
func BroadcastSatelliteState(elements KeplerianElements, clock ClockPolynomial, constants ConstellationConstants, tSOWS, tGDS float64, isGeo bool) (SatelliteState, error) {
	value, err := native.BroadcastSatelliteState(nativeBroadcastElements(elements), native.NativeClockPolynomial{AF0: clock.AF0, AF1: clock.AF1, AF2: clock.AF2, TocSOW: clock.TocSOW}, native.NativeConstellationConstants{GMM3PerS2: constants.GMM3PerS2, OmegaERadPerS: constants.OmegaERadPerS, DTRF: constants.DTRF}, tSOWS, tGDS, isGeo)
	return SatelliteState{Orbit: publicOrbitState(value.Orbit), Clock: ClockOffset{ClockPolynomialS: value.Clock.ClockPolynomialS, RelativisticS: value.Clock.RelativisticS, GroupDelayS: value.Clock.GroupDelayS, TotalS: value.Clock.TotalS}}, publicError(err)
}

// SelectSP3 selects a product usable at one requested J2000 epoch.
func SelectSP3(products []*SP3, requestedEpochJ2000S float64, policy StalenessPolicy) (*SP3, StalenessMetadata, error) {
	nativeProducts := make([]*native.SP3, len(products))
	for i, value := range products {
		nativeProducts[i] = nativeSP3Handle(value)
	}
	selected, metadata, err := native.SelectSP3(nativeProducts, requestedEpochJ2000S, native.StalenessPolicy{MaxStalenessS: policy.MaxStalenessS})
	if err != nil {
		return nil, StalenessMetadata{}, publicError(err)
	}
	return &SP3{handle: selected}, stalenessMetadata(metadata), nil
}

// SelectSP3OverRange selects a product usable across an epoch range.
func SelectSP3OverRange(products []*SP3, startEpochJ2000S, endEpochJ2000S float64, policy StalenessPolicy) (*SP3, StalenessMetadata, error) {
	nativeProducts := make([]*native.SP3, len(products))
	for i, value := range products {
		nativeProducts[i] = nativeSP3Handle(value)
	}
	selected, metadata, err := native.SelectSP3OverRange(nativeProducts, startEpochJ2000S, endEpochJ2000S, native.StalenessPolicy{MaxStalenessS: policy.MaxStalenessS})
	if err != nil {
		return nil, StalenessMetadata{}, publicError(err)
	}
	return &SP3{handle: selected}, stalenessMetadata(metadata), nil
}

// SPPDopplerObservation is one Doppler row in hertz and carrier frequency.
type SPPDopplerObservation struct {
	SatelliteID                                    string
	DopplerHz, CarrierHz, SatelliteClockDriftSPerS float64
}

// SPPDopplerVelocityErrorKind is the native typed Doppler outcome.
type SPPDopplerVelocityErrorKind uint32

const (
	SPPDopplerVelocityNoError SPPDopplerVelocityErrorKind = iota
	SPPDopplerVelocityNoObservations
	SPPDopplerVelocityTooFewSatellites
	SPPDopplerVelocitySingularGeometry
	SPPDopplerVelocityDuplicateObservation
	SPPDopplerVelocityInvalidCarrier
	SPPDopplerVelocityInvalidInput
	SPPDopplerVelocityInvalidObservation
	SPPDopplerVelocityInvalidReceiverState
)

// SPPDopplerSolution is a detached receiver SPP result plus Doppler status.
type SPPDopplerSolution struct {
	Receiver          SPPSolution
	HasVelocity       bool
	VelocityErrorKind SPPDopplerVelocityErrorKind
	Velocity          *SPPDopplerVelocitySolution
}

// SPPDopplerVelocitySolution is the detached native velocity result. Velocity
// and residuals use metres per second; ClockDriftSPerS uses seconds per second.
// StateCovariance is the native row-major unit-variance 4x4 matrix.
type SPPDopplerVelocitySolution struct {
	VelocityMPerS      [3]float64
	ClockDriftSPerS    float64
	SpeedMPerS         float64
	StateCovariance    [16]float64
	UsedSatelliteCount int
	UsedSatelliteIDs   []string
	ResidualsMPerS     []float64
}

// SolveBroadcast solves one SPP epoch using broadcast navigation messages.
func SolveBroadcast(broadcast *BroadcastEphemeris, config SPPConfig) (SPPSolution, error) {
	if broadcast == nil || broadcast.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	result, err := broadcast.handle.SolveBroadcast(nativeSPPConfig(config))
	return publicSPPSolution(result), publicError(err)
}

// SolveBroadcastWithDopplerVelocity solves broadcast SPP and its native Doppler
// velocity extension. The returned receiver solution is always detached.
func SolveBroadcastWithDopplerVelocity(broadcast *BroadcastEphemeris, config SPPConfig, observations []SPPDopplerObservation) (SPPDopplerSolution, error) {
	if broadcast == nil || broadcast.handle == nil {
		return SPPDopplerSolution{}, ErrClosed
	}
	nativeObservations := make([]native.NativeSppDopplerObservation, len(observations))
	for i, value := range observations {
		nativeObservations[i] = native.NativeSppDopplerObservation{SatelliteID: value.SatelliteID, DopplerHz: value.DopplerHz, CarrierHz: value.CarrierHz, SatelliteClockDriftSS: value.SatelliteClockDriftSPerS}
	}
	value, err := broadcast.handle.SolveBroadcastWithDopplerVelocity(nativeSPPConfig(config), nativeObservations)
	out := SPPDopplerSolution{Receiver: publicSPPSolution(value.Receiver), HasVelocity: value.HasVelocity, VelocityErrorKind: SPPDopplerVelocityErrorKind(value.VelocityErrorKind)}
	if value.Velocity != nil {
		out.Velocity = &SPPDopplerVelocitySolution{VelocityMPerS: value.Velocity.VelocityMPerS, ClockDriftSPerS: value.Velocity.ClockDriftSPerS, SpeedMPerS: value.Velocity.SpeedMPerS, StateCovariance: value.Velocity.StateCovariance, UsedSatelliteCount: value.Velocity.UsedSatelliteCount, UsedSatelliteIDs: append([]string(nil), value.Velocity.UsedSatelliteIDs...), ResidualsMPerS: append([]float64(nil), value.Velocity.ResidualsMPerS...)}
	}
	return out, publicError(err)
}
