package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// GLONASSChannel supplies the FDMA channel used by the native ionosphere
// correction for one GLONASS slot.
type GLONASSChannel struct {
	// Slot is the GLONASS satellite slot; Channel is its FDMA channel number.
	Slot    uint8
	Channel int8
}

// SPPSolvePolicy controls native validation and coarse-search behavior.
type SPPSolvePolicy struct {
	// UseValidationOptions controls Validation; CoarseSearchSeeds is a native count.
	UseValidationOptions bool
	Validation           SolutionValidationOptions
	CoarseSearchEnabled  bool
	CoarseSearchSeeds    uint64
}

// SPPInputsV2 exposes the extended SPP controls without retaining any native
// pointer. All observations and channel rows are copied by the solve call.
type SPPInputsV2 struct {
	// Base carries metre/second SPP observations and atmospheric settings.
	Base SPPConfig
	// BeidouKlobucharEnabled controls the optional BeiDou alpha/beta arrays.
	BeidouKlobucharEnabled bool
	// BeidouKlobucharAlpha and BeidouKlobucharBeta are optional coefficients.
	BeidouKlobucharAlpha, BeidouKlobucharBeta [4]float64
	// RobustEnabled controls the Robust policy fields.
	RobustEnabled bool
	// Robust contains native robust-loss parameters in metres and iteration counts.
	Robust          SPPRobustConfig
	Policy          SPPSolvePolicy
	GLONASSChannels []GLONASSChannel
}

// RINEXSPPOptions controls native assembly and solving of RINEX epochs.
type RINEXSPPOptions struct {
	// Ionosphere, Troposphere, and InitialGuessEnabled select optional corrections.
	Ionosphere, Troposphere, InitialGuessEnabled bool
	InitialGuess                                 [4]float64
	PressureHPA, TemperatureK, RelativeHumidity  float64
	RobustEnabled                                bool
	Robust                                       SPPRobustConfig
}

// RINEXSPPEpoch is detached RINEX-SPP epoch metadata.
type RINEXSPPEpoch struct {
	// Index and ObservationCount are native counts for the detached epoch.
	Index, ObservationCount int
	Epoch                   CivilDateTime
}

// SPPRejectedSatellite describes one native rejection reason.
type SPPRejectedSatellite struct {
	// SatelliteID identifies the rejected row; Reason is the native reason code.
	SatelliteID string
	Reason      uint32
}

// SPPSystemClock is one detached per-constellation receiver clock.
type SPPSystemClock struct {
	// System identifies the constellation; ReceiverClockS is in seconds.
	System         GNSSSystem
	ReceiverClockS float64
}

// SPPSystemTDOP is one detached per-constellation time DOP.
type SPPSystemTDOP struct {
	// System identifies the constellation; TDOP is dimensionless.
	System GNSSSystem
	TDOP   float64
}

// SPPSolutionHandle owns a native solution and must not be copied after use.
type SPPSolutionHandle struct {
	_      noCopy
	handle *native.SppSolutionHandle
}

// RINEXSPPInputs owns assembled native epoch inputs.
type RINEXSPPInputs struct {
	_      noCopy
	handle *native.RinexSPPInputs
}

// RINEXSPPSolutions owns per-epoch RINEX-SPP results.
type RINEXSPPSolutions struct {
	_      noCopy
	handle *native.RinexSPPSolutions
}

// SPPBatch owns per-epoch batch results.
type SPPBatch struct {
	_      noCopy
	handle *native.SPPBatch
}

func nativeSolvePolicy(v SPPSolvePolicy) (native.NativeSPPSolvePolicy, error) {
	seeds, err := nativeCountToInt(v.CoarseSearchSeeds, "SPP coarse-search seeds")
	if err != nil {
		return native.NativeSPPSolvePolicy{}, err
	}
	return native.NativeSPPSolvePolicy{UseValidationOptions: v.UseValidationOptions, Validation: native.NativeSPPValidationOptions{MaxPDOPEnabled: v.Validation.HasMaxPDOP, MaxPDOP: v.Validation.MaxPDOP, MinPlausibleRadiusM: v.Validation.MinPlausibleRadiusM, MaxPlausibleRadiusM: v.Validation.MaxPlausibleRadiusM, MaxConvergedResidualRMSM: v.Validation.MaxConvergedResidualRMSM}, CoarseSearchEnabled: v.CoarseSearchEnabled, CoarseSearchSeeds: seeds}, nil
}
func nativeSppV2(v SPPInputsV2) (native.SppInputsV2, error) {
	channels := make([]native.NativeGlonassChannel, len(v.GLONASSChannels))
	for i, x := range v.GLONASSChannels {
		channels[i] = native.NativeGlonassChannel{Slot: x.Slot, Channel: x.Channel}
	}
	robust, err := nativeRobustConfig(v.Robust)
	if err != nil {
		return native.SppInputsV2{}, err
	}
	policy, err := nativeSolvePolicy(v.Policy)
	if err != nil {
		return native.SppInputsV2{}, err
	}
	return native.SppInputsV2{Base: nativeSPPConfig(v.Base), BeidouEnabled: v.BeidouKlobucharEnabled, BeidouAlpha: v.BeidouKlobucharAlpha, BeidouBeta: v.BeidouKlobucharBeta, RobustEnabled: v.RobustEnabled, Robust: robust, Policy: policy, GlonassChannels: channels}, nil
}
func nativeRinexSppOptions(v RINEXSPPOptions) (native.NativeRinexSPPOptions, error) {
	robust, err := nativeRobustConfig(v.Robust)
	if err != nil {
		return native.NativeRinexSPPOptions{}, err
	}
	return native.NativeRinexSPPOptions{Ionosphere: v.Ionosphere, Troposphere: v.Troposphere, InitialGuessEnabled: v.InitialGuessEnabled, InitialGuess: v.InitialGuess, PressureHPA: v.PressureHPA, TemperatureK: v.TemperatureK, RelativeHumidity: v.RelativeHumidity, RobustEnabled: v.RobustEnabled, Robust: robust}, nil
}
func publicRinexEpoch(v native.NativeRinexSPPEpoch) RINEXSPPEpoch {
	return RINEXSPPEpoch{Index: v.Index, ObservationCount: v.ObservationCount, Epoch: civilFromNative(v.Epoch)}
}
func publicSppV2(v native.SppInputsV2) (SPPInputsV2, error) {
	maxOuter, err := nativeCountToInt(v.Robust.MaxOuter, "SPP robust max outer")
	if err != nil {
		return SPPInputsV2{}, err
	}
	seeds, err := nativeCountToInt(uint64(v.Policy.CoarseSearchSeeds), "SPP coarse-search seeds")
	if err != nil {
		return SPPInputsV2{}, err
	}
	out := SPPInputsV2{Base: SPPConfig{TRxJ2000S: v.Base.TRxJ2000S, TRxSecondOfDayS: v.Base.TRxSecondOfDayS, DayOfYear: v.Base.DayOfYear, InitialGuess: v.Base.InitialGuess, Ionosphere: v.Base.Ionosphere, Troposphere: v.Base.Troposphere, WithGeodetic: v.Base.WithGeodetic, KlobucharAlpha: v.Base.KlobucharAlpha, KlobucharBeta: v.Base.KlobucharBeta, PressureHPA: v.Base.PressureHPA, TemperatureK: v.Base.TemperatureK, RelativeHumidity: v.Base.RelativeHumidity}, BeidouKlobucharEnabled: v.BeidouEnabled, BeidouKlobucharAlpha: v.BeidouAlpha, BeidouKlobucharBeta: v.BeidouBeta, RobustEnabled: v.RobustEnabled, Robust: SPPRobustConfig{HuberK: v.Robust.HuberK, ScaleFloorM: v.Robust.ScaleFloorM, MaxOuter: maxOuter, OuterToleranceM: v.Robust.OuterToleranceM}, Policy: SPPSolvePolicy{UseValidationOptions: v.Policy.UseValidationOptions, CoarseSearchEnabled: v.Policy.CoarseSearchEnabled, CoarseSearchSeeds: uint64(seeds)}}
	out.Base.Observations = make([]SPPObservation, len(v.Base.Observations))
	for i, x := range v.Base.Observations {
		out.Base.Observations[i] = SPPObservation{SatelliteID: x.SatelliteID, PseudorangeM: x.PseudorangeM}
	}
	for _, x := range v.GlonassChannels {
		out.GLONASSChannels = append(out.GLONASSChannels, GLONASSChannel{Slot: x.Slot, Channel: x.Channel})
	}
	out.Policy.Validation = SolutionValidationOptions{HasMaxPDOP: v.Policy.Validation.MaxPDOPEnabled, MaxPDOP: v.Policy.Validation.MaxPDOP, MinPlausibleRadiusM: v.Policy.Validation.MinPlausibleRadiusM, MaxPlausibleRadiusM: v.Policy.Validation.MaxPlausibleRadiusM, MaxConvergedResidualRMSM: v.Policy.Validation.MaxConvergedResidualRMSM}
	return out, nil
}

// RINEXSPPOptionsInit returns C's RINEX-SPP defaults.
func RINEXSPPOptionsInit() (RINEXSPPOptions, error) {
	v, e := native.RINEXSPPOptionsInit()
	if e != nil {
		return RINEXSPPOptions{}, publicError(e)
	}
	maxOuter, err := nativeCountToInt(v.Robust.MaxOuter, "RINEX SPP robust max outer")
	if err != nil {
		return RINEXSPPOptions{}, err
	}
	return RINEXSPPOptions{Ionosphere: v.Ionosphere, Troposphere: v.Troposphere, InitialGuessEnabled: v.InitialGuessEnabled, InitialGuess: v.InitialGuess, PressureHPA: v.PressureHPA, TemperatureK: v.TemperatureK, RelativeHumidity: v.RelativeHumidity, RobustEnabled: v.RobustEnabled, Robust: SPPRobustConfig{HuberK: v.Robust.HuberK, ScaleFloorM: v.Robust.ScaleFloorM, MaxOuter: maxOuter, OuterToleranceM: v.Robust.OuterToleranceM}}, nil
}

// SPPInputsV2Init returns C's V2 defaults.
func SPPInputsV2Init() (SPPInputsV2, error) {
	v, e := native.SPPInputsV2Init()
	if e != nil {
		return SPPInputsV2{}, publicError(e)
	}
	return publicSppV2(v)
}

// Close releases assembled RINEX SPP inputs and is idempotent.
func (r *RINEXSPPInputs) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Count returns the number of assembled RINEX epochs.
func (r *RINEXSPPInputs) Count() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	v, e := r.handle.Count()
	return v, publicError(e)
}

// Epoch returns detached metadata for one assembled RINEX epoch.
func (r *RINEXSPPInputs) Epoch(i int) (RINEXSPPEpoch, error) {
	if r == nil || r.handle == nil {
		return RINEXSPPEpoch{}, ErrClosed
	}
	v, e := r.handle.Epoch(i)
	return publicRinexEpoch(v), publicError(e)
}

// EpochInputs returns detached V2 inputs for one assembled RINEX epoch.
func (r *RINEXSPPInputs) EpochInputs(i int) (SPPInputsV2, error) {
	if r == nil || r.handle == nil {
		return SPPInputsV2{}, ErrClosed
	}
	v, e := r.handle.EpochInputs(i)
	if e != nil {
		return SPPInputsV2{}, publicError(e)
	}
	return publicSppV2(v)
}

// Close releases RINEX SPP solutions and is idempotent.
func (r *RINEXSPPSolutions) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Count returns the number of RINEX SPP solution epochs.
func (r *RINEXSPPSolutions) Count() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	v, e := r.handle.Count()
	return v, publicError(e)
}

// Epoch returns detached metadata for one solution epoch.
func (r *RINEXSPPSolutions) Epoch(i int) (RINEXSPPEpoch, error) {
	if r == nil || r.handle == nil {
		return RINEXSPPEpoch{}, ErrClosed
	}
	v, e := r.handle.Epoch(i)
	return publicRinexEpoch(v), publicError(e)
}

// SolutionOK reports whether one RINEX epoch solved successfully.
func (r *RINEXSPPSolutions) SolutionOK(i int) (bool, error) {
	if r == nil || r.handle == nil {
		return false, ErrClosed
	}
	v, e := r.handle.SolutionOK(i)
	return v, publicError(e)
}

// SolutionError returns native error text for one failed epoch.
func (r *RINEXSPPSolutions) SolutionError(i int) (string, error) {
	if r == nil || r.handle == nil {
		return "", ErrClosed
	}
	v, e := r.handle.SolutionError(i)
	return v, publicError(e)
}

// Solution returns an owning handle for one successful epoch.
func (r *RINEXSPPSolutions) Solution(i int) (*SPPSolutionHandle, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.Solution(i)
	if e != nil {
		return nil, publicError(e)
	}
	return &SPPSolutionHandle{handle: v}, nil
}

// Close releases batch results and is idempotent.
func (b *SPPBatch) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}

// Count returns the number of SPP batch epochs.
func (b *SPPBatch) Count() (int, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	v, e := b.handle.Count()
	return v, publicError(e)
}

// EpochOK reports whether one batch epoch solved successfully.
func (b *SPPBatch) EpochOK(i int) (bool, error) {
	if b == nil || b.handle == nil {
		return false, ErrClosed
	}
	v, e := b.handle.EpochOK(i)
	return v, publicError(e)
}

// Error returns native error text for one failed batch epoch.
func (b *SPPBatch) Error(i int) (string, error) {
	if b == nil || b.handle == nil {
		return "", ErrClosed
	}
	v, e := b.handle.Error(i)
	return v, publicError(e)
}

// Solution returns an owning handle for one successful batch epoch.
func (b *SPPBatch) Solution(i int) (*SPPSolutionHandle, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	v, e := b.handle.Solution(i)
	if e != nil {
		return nil, publicError(e)
	}
	return &SPPSolutionHandle{handle: v}, nil
}

// Close releases the SPP solution and is idempotent.
func (s *SPPSolutionHandle) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Solution returns the detached position/time solution.
func (s *SPPSolutionHandle) Solution() (SPPSolution, error) {
	if s == nil || s.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	v, e := s.handle.Solution()
	return publicSPPSolution(v), publicError(e)
}

// PositionCovarianceECEFM2 returns ECEF position covariance in square metres.
func (s *SPPSolutionHandle) PositionCovarianceECEFM2() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	v, e := s.handle.CovarianceECEFM2()
	return v, publicError(e)
}

// PositionCovarianceENUM2 returns ENU position covariance in square metres.
func (s *SPPSolutionHandle) PositionCovarianceENUM2() ([9]float64, error) {
	if s == nil || s.handle == nil {
		return [9]float64{}, ErrClosed
	}
	v, e := s.handle.CovarianceENUM2()
	return v, publicError(e)
}

// RejectedSatellites returns detached rejected-satellite rows.
func (s *SPPSolutionHandle) RejectedSatellites() ([]SPPRejectedSatellite, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.RejectedSatellites()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SPPRejectedSatellite, len(v))
	for i, x := range v {
		out[i] = SPPRejectedSatellite{SatelliteID: x.SatelliteID, Reason: x.Reason}
	}
	return out, nil
}

// ReceiverClockDriftSS returns receiver clock drift in seconds per second.
func (s *SPPSolutionHandle) ReceiverClockDriftSS() (float64, bool, error) {
	if s == nil || s.handle == nil {
		return 0, false, ErrClosed
	}
	v, p, e := s.handle.ReceiverClockDrift()
	return v, p, publicError(e)
}

// SystemClocks returns detached per-constellation receiver clocks in seconds.
func (s *SPPSolutionHandle) SystemClocks() ([]SPPSystemClock, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.SystemClocks()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SPPSystemClock, len(v))
	for i, x := range v {
		out[i] = SPPSystemClock{System: GNSSSystem(x.System), ReceiverClockS: x.ReceiverS}
	}
	return out, nil
}

// SystemTDOPs returns detached per-constellation time DOP values.
func (s *SPPSolutionHandle) SystemTDOPs() ([]SPPSystemTDOP, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.SystemTDOPs()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SPPSystemTDOP, len(v))
	for i, x := range v {
		out[i] = SPPSystemTDOP{System: GNSSSystem(x.System), TDOP: x.TDOP}
	}
	return out, nil
}

// SolveSPPV2 solves one extended SPP epoch through the native solver.
func SolveSPPV2(sp3 *SP3, input SPPInputsV2) (*SPPSolutionHandle, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	ninput, e := nativeSppV2(input)
	if e != nil {
		return nil, publicError(e)
	}
	v, e := native.SolveSPPV2(sp3.handle, ninput)
	if e != nil {
		return nil, publicError(e)
	}
	return &SPPSolutionHandle{handle: v}, nil
}

// SolveSPPBatchSerial solves SPP epochs serially through the native solver.
func SolveSPPBatchSerial(sp3 *SP3, inputs []SPPInputsV2, withGeodetic bool, policy SPPSolvePolicy) (*SPPBatch, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	values := make([]native.SppInputsV2, len(inputs))
	for i, x := range inputs {
		value, e := nativeSppV2(x)
		if e != nil {
			return nil, publicError(e)
		}
		values[i] = value
	}
	npolicy, e := nativeSolvePolicy(policy)
	if e != nil {
		return nil, publicError(e)
	}
	v, e := native.SolveSPPBatchSerial(sp3.handle, values, withGeodetic, npolicy)
	if e != nil {
		return nil, publicError(e)
	}
	return &SPPBatch{handle: v}, nil
}

// SolveSPPBatchParallel solves SPP epochs in native parallel mode.
func SolveSPPBatchParallel(sp3 *SP3, inputs []SPPInputsV2, withGeodetic bool, policy SPPSolvePolicy) (*SPPBatch, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	values := make([]native.SppInputsV2, len(inputs))
	for i, x := range inputs {
		value, e := nativeSppV2(x)
		if e != nil {
			return nil, publicError(e)
		}
		values[i] = value
	}
	npolicy, e := nativeSolvePolicy(policy)
	if e != nil {
		return nil, publicError(e)
	}
	v, e := native.SolveSPPBatchParallel(sp3.handle, values, withGeodetic, npolicy)
	if e != nil {
		return nil, publicError(e)
	}
	return &SPPBatch{handle: v}, nil
}

// SolveSPPWithDopplerVelocity solves SPP and Doppler velocity together.
func SolveSPPWithDopplerVelocity(sp3 *SP3, input SPPInputsV2, rows []SPPDopplerObservation) (SPPDopplerSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return SPPDopplerSolution{}, ErrClosed
	}
	values := make([]native.NativeSppDopplerObservation, len(rows))
	for i, x := range rows {
		values[i] = native.NativeSppDopplerObservation{SatelliteID: x.SatelliteID, DopplerHz: x.DopplerHz, CarrierHz: x.CarrierHz, SatelliteClockDriftSS: x.SatelliteClockDriftSPerS}
	}
	ninput, e := nativeSppV2(input)
	if e != nil {
		return SPPDopplerSolution{}, publicError(e)
	}
	v, e := native.SolveSPPWithDoppler(sp3.handle, ninput, values)
	if e != nil {
		return SPPDopplerSolution{}, publicError(e)
	}
	out := SPPDopplerSolution{Receiver: publicSPPSolution(v.Receiver), HasVelocity: v.HasVelocity, VelocityErrorKind: SPPDopplerVelocityErrorKind(v.VelocityErrorKind)}
	if v.Velocity != nil {
		out.Velocity = &SPPDopplerVelocitySolution{VelocityMPerS: v.Velocity.VelocityMPerS, ClockDriftSPerS: v.Velocity.ClockDriftSPerS, SpeedMPerS: v.Velocity.SpeedMPerS, StateCovariance: v.Velocity.StateCovariance, UsedSatelliteCount: v.Velocity.UsedSatelliteCount, UsedSatelliteIDs: append([]string(nil), v.Velocity.UsedSatelliteIDs...), ResidualsMPerS: append([]float64(nil), v.Velocity.ResidualsMPerS...)}
	}
	return out, nil
}

// SolveSPPFromRINEXObs assembles and solves native RINEX SPP epochs.
func SolveSPPFromRINEXObs(broadcast *BroadcastEphemeris, obs *RINEXObservation, options *RINEXSPPOptions, withGeodetic bool, policy *SPPSolvePolicy) (*RINEXSPPSolutions, error) {
	if broadcast == nil || broadcast.handle == nil || obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	var no *native.NativeRinexSPPOptions
	if options != nil {
		v, e := nativeRinexSppOptions(*options)
		if e != nil {
			return nil, publicError(e)
		}
		no = &v
	}
	var np *native.NativeSPPSolvePolicy
	if policy != nil {
		v, e := nativeSolvePolicy(*policy)
		if e != nil {
			return nil, publicError(e)
		}
		np = &v
	}
	v, e := native.SolveSPPFromRINEXObs(broadcast.handle, obs.handle, no, withGeodetic, np)
	if e != nil {
		return nil, publicError(e)
	}
	return &RINEXSPPSolutions{handle: v}, nil
}

// SPPInputsFromRINEXObs assembles native SPP inputs from RINEX observations.
func SPPInputsFromRINEXObs(obs *RINEXObservation, broadcast *BroadcastEphemeris, options *RINEXSPPOptions) (*RINEXSPPInputs, error) {
	if broadcast == nil || broadcast.handle == nil || obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	var no *native.NativeRinexSPPOptions
	if options != nil {
		v, e := nativeRinexSppOptions(*options)
		if e != nil {
			return nil, publicError(e)
		}
		no = &v
	}
	v, e := native.SPPInputsFromRINEXObs(obs.handle, broadcast.handle, no)
	if e != nil {
		return nil, publicError(e)
	}
	return &RINEXSPPInputs{handle: v}, nil
}

// SolveWithFallback tries precise and broadcast native SPP routes in policy order.
func SolveWithFallback(precise []*SP3, broadcast *BroadcastEphemeris, input SPPConfig, policy StalenessPolicy) (*SourcedSolution, error) {
	if broadcast == nil || broadcast.handle == nil {
		return nil, ErrClosed
	}
	products := make([]*native.SP3, len(precise))
	for i, x := range precise {
		if x == nil || x.handle == nil {
			return nil, ErrClosed
		}
		products[i] = x.handle
	}
	v, e := native.SolveWithFallback(products, broadcast.handle, nativeSPPConfig(input), native.StalenessPolicy{MaxStalenessS: policy.MaxStalenessS})
	if e != nil {
		return nil, publicError(e)
	}
	return &SourcedSolution{handle: v}, nil
}

// Policy contains native validation and coarse-search options.
// GLONASSChannels is copied channel metadata for GLONASS corrections.
// InitialGuess is an ECEF/geodetic native initial state in metres as configured.
// PressureHPA, TemperatureK, and RelativeHumidity are atmospheric inputs.
// RobustEnabled controls Robust.
