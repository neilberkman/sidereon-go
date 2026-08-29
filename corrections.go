package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// BroadcastEphemeris owns a C-backed RINEX navigation source. The byte parser
// and all ephemeris selection/correction behavior remain in the native engine.
type BroadcastEphemeris struct {
	_      noCopy
	handle *native.BroadcastEphemeris
}

// AirborneModel contains the airborne-model noise-divergence standard
// deviation in metres.
type AirborneModel struct {
	SigmaNoiseDivergenceM float64
}

// DegradationParams contains SBAS fast-correction degradation terms. Distance
// terms are metres and rate terms are metres per second.
type DegradationParams struct {
	DeltaUDRE, EpsFCM, EpsRRCM, EpsLTCM, EpsERM, EpsIonoM float64
	RSSUDRE                                               bool
}

// SBASKMultipliers contains horizontal and vertical protection-level factors.
type SBASKMultipliers struct {
	KH float64
	KV float64
}

// SBASProtectionRow is one satellite geometry row in the local ENU frame.
type SBASProtectionRow struct {
	SatelliteID  string
	LineOfSight  LineOfSight
	System       GNSSSystem
	ElevationRad float64
}

// SBASSISError contains one satellite-in-space error model in metres.
type SBASSISError struct {
	SatelliteID                                   string
	SigmaFLTM, SigmaUIREM, SigmaAirM, SigmaTropoM float64
}

// SBASProtection contains horizontal and vertical protection levels in metres
// and the associated covariance ellipse terms in metres or square metres.
type SBASProtection struct {
	HPLM, VPLM, DMajorM, SigmaUM, DEastM, DNorthM, DENM2 float64
}

// SBASProtectionError is the C protection-level outcome.
type SBASProtectionError uint32

const (
	SBASProtectionNoError              SBASProtectionError = SBASProtectionError(native.SBASPLNoErrorValue)
	SBASProtectionInsufficientGeometry SBASProtectionError = SBASProtectionError(native.SBASPLInsufficientGeometryValue)
	SBASProtectionNumericalFailure     SBASProtectionError = SBASProtectionError(native.SBASPLNumericalFailureValue)
	SBASProtectionInvalidErrorModel    SBASProtectionError = SBASProtectionError(native.SBASPLInvalidErrorModelValue)
)

// NewAirborneModelAADA returns the native AADA airborne model.
func NewAirborneModelAADA() (AirborneModel, error) {
	v, err := native.SbasAirborneModelAADA()
	return AirborneModel{SigmaNoiseDivergenceM: v.SigmaNoiseDivergenceM}, publicError(err)
}

// AirborneSigma returns airborne sigma in metres at an elevation in radians.
func AirborneSigma(model AirborneModel, elevationRad float64) (float64, error) {
	v, err := native.SbasAirborneSigma(native.NativeAirborneModel{SigmaNoiseDivergenceM: model.SigmaNoiseDivergenceM}, elevationRad)
	return v, publicError(err)
}

// NewSBASDegradationParams returns the native no-degradation parameters.
func NewSBASDegradationParams() (DegradationParams, error) {
	v, err := native.SbasDegradationParamsNone()
	return DegradationParams{DeltaUDRE: v.DeltaUDRE, EpsFCM: v.EpsFCM, EpsRRCM: v.EpsRRCM, EpsLTCM: v.EpsLTCM, EpsERM: v.EpsERM, EpsIonoM: v.EpsIonoM, RSSUDRE: v.RSSUDRE}, publicError(err)
}

// NewSBASKMultipliersEnRouteNPA returns en-route/non-precision factors.
func NewSBASKMultipliersEnRouteNPA() (SBASKMultipliers, error) {
	v, err := native.SbasKMultipliersEnRouteNPA()
	return SBASKMultipliers{KH: v.KH, KV: v.KV}, publicError(err)
}

// NewSBASKMultipliersPrecisionApproach returns precision-approach factors.
func NewSBASKMultipliersPrecisionApproach() (SBASKMultipliers, error) {
	v, err := native.SbasKMultipliersPrecisionApproach()
	return SBASKMultipliers{KH: v.KH, KV: v.KV}, publicError(err)
}

// SBASSigmaFLTMForUDREI returns fast/long-term sigma in metres for a UDREI.
func SBASSigmaFLTMForUDREI(udrei uint8, params DegradationParams) (float64, error) {
	v, err := native.SbasSigmaFLTM(udrei, native.NativeDegradationParams{DeltaUDRE: params.DeltaUDRE, EpsFCM: params.EpsFCM, EpsRRCM: params.EpsRRCM, EpsLTCM: params.EpsLTCM, EpsERM: params.EpsERM, EpsIonoM: params.EpsIonoM, RSSUDRE: params.RSSUDRE})
	return v, publicError(err)
}

// SBASSigmaTropoM returns tropospheric sigma in metres at an elevation in
// radians.
func SBASSigmaTropoM(elevationRad float64) (float64, error) {
	v, err := native.SbasSigmaTropo(elevationRad)
	return v, publicError(err)
}

// SBASSISErrorSigmaM returns the combined satellite-in-space sigma in metres.
func SBASSISErrorSigmaM(value SBASSISError) (float64, error) {
	v, err := native.SbasSISErrorSigma(native.NativeSbasSISError{SatelliteID: value.SatelliteID, SigmaFLTM: value.SigmaFLTM, SigmaUIREM: value.SigmaUIREM, SigmaAirM: value.SigmaAirM, SigmaTropoM: value.SigmaTropoM})
	return v, publicError(err)
}

// SBASProtectionLevels computes ECEF-to-local protection levels from rows,
// receiver geodetic coordinates, and the selected system error model.
func SBASProtectionLevels(rows []SBASProtectionRow, receiver Geodetic, clockSystems []GNSSSystem, model []SBASSISError, multipliers SBASKMultipliers) (SBASProtection, SBASProtectionError, error) {
	nativeRows := make([]native.NativeSbasProtectionRow, len(rows))
	for i, value := range rows {
		if err := validateGNSSSystem(value.System); err != nil {
			return SBASProtection{}, 0, err
		}
		nativeRows[i] = native.NativeSbasProtectionRow{SatelliteID: value.SatelliteID, LineOfSight: [3]float64{value.LineOfSight.EX, value.LineOfSight.EY, value.LineOfSight.EZ}, System: uint32(value.System), ElevationRad: value.ElevationRad}
	}
	nativeModel := make([]native.NativeSbasSISError, len(model))
	for i, value := range model {
		nativeModel[i] = native.NativeSbasSISError{SatelliteID: value.SatelliteID, SigmaFLTM: value.SigmaFLTM, SigmaUIREM: value.SigmaUIREM, SigmaAirM: value.SigmaAirM, SigmaTropoM: value.SigmaTropoM}
	}
	nativeSystems := make([]uint32, len(clockSystems))
	for i, value := range clockSystems {
		if err := validateGNSSSystem(value); err != nil {
			return SBASProtection{}, 0, err
		}
		nativeSystems[i] = uint32(value)
	}
	v, outcome, err := native.SbasProtectionLevels(nativeRows, [3]float64{receiver.LatitudeRad, receiver.LongitudeRad, receiver.HeightM}, nativeSystems, nativeModel, native.NativeSbasKMultipliers{KH: multipliers.KH, KV: multipliers.KV})
	return SBASProtection{HPLM: v.HPLM, VPLM: v.VPLM, DMajorM: v.DMajorM, SigmaUM: v.SigmaUM, DEastM: v.DEastM, DNorthM: v.DNorthM, DENM2: v.DENM2}, SBASProtectionError(outcome), publicError(err)
}

// SBASSolveMode selects how the native SBAS correction store participates in a
// broadcast SPP solve.
type SBASSolveMode uint32

const (
	// SBASMixedAugmentation combines broadcast and SBAS corrections.
	SBASMixedAugmentation SBASSolveMode = SBASSolveMode(native.SBASSolveMixedValue)
	// SBASOnly uses SBAS corrections without the mixed fallback mode.
	SBASOnly SBASSolveMode = SBASSolveMode(native.SBASSolveSBASOnlyValue)
)

// ParseBroadcastEphemeris parses a RINEX navigation byte slice in C-owned
// storage and returns an owning broadcast handle.
func ParseBroadcastEphemeris(data []byte) (*BroadcastEphemeris, error) {
	h, e := native.ParseBroadcast(data)
	if e != nil {
		return nil, publicError(e)
	}
	return &BroadcastEphemeris{handle: h}, nil
}

// LoadBroadcastEphemeris delegates path loading and parsing to the native C
// ABI, preserving its UTF-8, filesystem, and status-error semantics.
func LoadBroadcastEphemeris(path string) (*BroadcastEphemeris, error) {
	h, e := native.LoadBroadcastEphemeris(path)
	if e != nil {
		return nil, publicError(e)
	}
	return &BroadcastEphemeris{handle: h}, nil
}

// Close releases the broadcast handle; repeated calls are safe.
func (b *BroadcastEphemeris) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}

// RecordCount returns the number of copied broadcast records.
func (b *BroadcastEphemeris) RecordCount() (int, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	v, e := b.handle.RecordCount()
	return v, publicError(e)
}

// Records returns detached broadcast records with native optional fields
// represented by their corresponding presence flags.
func (b *BroadcastEphemeris) Records() ([]BroadcastRecord, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	v, e := b.handle.Records()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]BroadcastRecord, len(v))
	for i, x := range v {
		out[i] = broadcastRecordFromNative(x)
	}
	return out, nil
}

// RINEXText returns a detached copy of the native RINEX navigation text.
func (b *BroadcastEphemeris) RINEXText() ([]byte, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	v, e := b.handle.Text()
	return v, publicError(e)
}

// IonosphereCorrections returns copied broadcast ionosphere coefficients.
func (b *BroadcastEphemeris) IonosphereCorrections() (IONOCorrections, error) {
	if b == nil || b.handle == nil {
		return IONOCorrections{}, ErrClosed
	}
	v, e := b.handle.IonoCorrections()
	return IONOCorrections{GPSPresent: v.GPSPresent, GPSAlpha: v.GPSAlpha, GPSBeta: v.GPSBeta, BeiDouPresent: v.BeiDouPresent, BeiDouAlpha: v.BeiDouAlpha, BeiDouBeta: v.BeiDouBeta, GalileoPresent: v.GalileoPresent, GalileoAI0: v.GalileoAI0, GalileoAI1: v.GalileoAI1, GalileoAI2: v.GalileoAI2}, publicError(e)
}

// LeapSeconds returns the UTC leap-second value and its presence flag.
func (b *BroadcastEphemeris) LeapSeconds() (float64, bool, error) {
	if b == nil || b.handle == nil {
		return 0, false, ErrClosed
	}
	v, p, e := b.handle.LeapSeconds()
	return v, p, publicError(e)
}

// SBASFastCorrection contains fast corrections in metres and metres per
// second, with TOfJ2000S in J2000 seconds.
type SBASFastCorrection struct {
	PRCM, RRCMPerS float64
	UDREI          uint8
	TOfJ2000S      float64
	IODF           uint8
}

// SBASLongTermCorrection contains ECEF metre corrections and rates in metres
// per second; T0J2000S is the J2000 reference epoch.
type SBASLongTermCorrection struct {
	IODE                               uint8
	DeltaECEFM, DeltaECEFRateMPerS     [3]float64
	DeltaAF0S, DeltaAF1SPerS, T0J2000S float64
}

// SBASGeoState contains a GEO ECEF state in metres, metres per second, and
// metres per second squared; clock fields are seconds-based.
type SBASGeoState struct {
	PositionECEFM, VelocityECEFMPerS, AccelerationECEFMPerS2 [3]float64
	ClockOffsetS, ClockDriftSS, T0J2000S                     float64
}

// SBASIGP contains an ionospheric grid point in degrees and metres. GIVE
// variance is optional and is present only when HasGIVEVariance is true.
type SBASIGP struct {
	LatitudeDeg, LongitudeDeg, VerticalDelayM float64
	HasGIVEVariance                           bool
	GIVEVarianceM2                            float64
}

// SBASCorrectionStore owns a C-backed SBAS correction state machine.
type SBASCorrectionStore struct {
	_      noCopy
	handle *native.SBASCorrectionStore
}

// NewSBASCorrectionStore creates an empty SBAS correction store.
func NewSBASCorrectionStore() (*SBASCorrectionStore, error) {
	h, e := native.NewSBASStore()
	if e != nil {
		return nil, publicError(e)
	}
	return &SBASCorrectionStore{handle: h}, nil
}

// Close releases the SBAS store; repeated calls are safe.
func (s *SBASCorrectionStore) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}
func nativeWeek(v GNSSWeekTow) (native.NativeGnssWeekTow, error) {
	if err := validateTimeScale(v.System); err != nil {
		return native.NativeGnssWeekTow{}, err
	}
	return native.NativeGnssWeekTow{System: uint32(v.System), Week: v.Week, TOWSeconds: v.TOWSeconds}, nil
}

// Ingest adds one SBAS block at the supplied GNSS week/TOW epoch.
func (s *SBASCorrectionStore) Ingest(block *SBASBlock, geo string, epoch GNSSWeekTow) error {
	if s == nil || s.handle == nil || block == nil || block.handle == nil {
		return ErrClosed
	}
	week, err := nativeWeek(epoch)
	if err != nil {
		return err
	}
	return publicError(s.handle.Ingest(block.handle, geo, week))
}

// PreferredGEO returns the preferred GEO at a J2000 epoch and its presence.
func (s *SBASCorrectionStore) PreferredGEO(tJ2000S float64) (string, bool, error) {
	if s == nil || s.handle == nil {
		return "", false, ErrClosed
	}
	v, p, e := s.handle.PreferredGeo(tJ2000S)
	return v, p, publicError(e)
}

// ReadyGEOs returns detached GEO identifiers ready at a J2000 epoch.
func (s *SBASCorrectionStore) ReadyGEOs(tJ2000S float64) ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.ReadyGeos(tJ2000S)
	return append([]string(nil), v...), publicError(e)
}

// FastCorrection returns a copied optional fast correction.
func (s *SBASCorrectionStore) FastCorrection(geo, satellite string) (SBASFastCorrection, bool, error) {
	if s == nil || s.handle == nil {
		return SBASFastCorrection{}, false, ErrClosed
	}
	v, p, e := s.handle.Fast(geo, satellite)
	return SBASFastCorrection{PRCM: v.PRCM, RRCMPerS: v.RRCMPerS, UDREI: v.UDREI, TOfJ2000S: v.TOfJ2000S, IODF: v.IODF}, p, publicError(e)
}

// LongTermCorrection returns a copied optional long-term ECEF correction.
func (s *SBASCorrectionStore) LongTermCorrection(geo, satellite string) (SBASLongTermCorrection, bool, error) {
	if s == nil || s.handle == nil {
		return SBASLongTermCorrection{}, false, ErrClosed
	}
	v, p, e := s.handle.LongTerm(geo, satellite)
	return SBASLongTermCorrection{IODE: v.IODE, DeltaECEFM: v.DeltaECEFM, DeltaECEFRateMPerS: v.DeltaECEFRateMPerS, DeltaAF0S: v.DeltaAF0S, DeltaAF1SPerS: v.DeltaAF1SPerS, T0J2000S: v.T0J2000S}, p, publicError(e)
}

// GeoNavigation returns a copied optional GEO navigation state.
func (s *SBASCorrectionStore) GeoNavigation(geo string) (SBASGeoState, bool, error) {
	if s == nil || s.handle == nil {
		return SBASGeoState{}, false, ErrClosed
	}
	v, p, e := s.handle.GeoNav(geo)
	return SBASGeoState{PositionECEFM: v.Position, VelocityECEFMPerS: v.Velocity, AccelerationECEFMPerS2: v.Acceleration, ClockOffsetS: v.ClockOffsetS, ClockDriftSS: v.ClockDriftSS, T0J2000S: v.T0J2000S}, p, publicError(e)
}

// IonosphereGrid returns detached IGP rows with optional GIVE variance.
func (s *SBASCorrectionStore) IonosphereGrid(geo string) ([]SBASIGP, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.IGPs(geo)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]SBASIGP, len(v))
	for i, x := range v {
		out[i] = SBASIGP{LatitudeDeg: x.LatDeg, LongitudeDeg: x.LonDeg, VerticalDelayM: x.VerticalDelayM, HasGIVEVariance: x.HasGIVEVariance, GIVEVarianceM2: x.GIVEVarianceM2}
	}
	return out, nil
}

// IonosphereSlantDelay returns an optional slant delay in metres. Angles are
// radians, frequency is hertz, and receiver coordinates are geodetic.
func (s *SBASCorrectionStore) IonosphereSlantDelay(geo string, receiver Geodetic, elevationRad, azimuthRad, frequencyHz float64) (float64, bool, error) {
	if s == nil || s.handle == nil {
		return 0, false, ErrClosed
	}
	v, p, e := s.handle.SlantDelay(geo, native.Geodetic{LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM}, elevationRad, azimuthRad, frequencyHz)
	return v, p, publicError(e)
}

// CorrectedState returns the SBAS-corrected ECEF position in metres and clock
// correction in seconds. The boolean reports correction presence.
func (s *SBASCorrectionStore) CorrectedState(b *BroadcastEphemeris, geo string, mode SBASSolveMode, satellite string, tJ2000S float64) ([3]float64, float64, bool, error) {
	if s == nil || s.handle == nil || b == nil || b.handle == nil {
		return [3]float64{}, 0, false, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return [3]float64{}, 0, false, err
	}
	p, c, present, e := s.handle.CorrectedState(b.handle, geo, uint32(mode), satellite, tJ2000S)
	return p, c, present, publicError(e)
}

// SolveBroadcast solves one epoch through the native SBAS-corrected broadcast
// route and returns a detached, value-oriented SPP solution.
func (s *SBASCorrectionStore) SolveBroadcast(b *BroadcastEphemeris, geo string, mode SBASSolveMode, config SPPConfig) (SPPSolution, error) {
	if s == nil || s.handle == nil || b == nil || b.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return SPPSolution{}, err
	}
	v, err := s.handle.SolveBroadcast(b.handle, geo, uint32(mode), nativeSPPConfig(config))
	if err != nil {
		return SPPSolution{}, publicError(err)
	}
	return fromNativeSPPSolution(v), nil
}

// SSRSource identifies the correction source defined by the public C API.
type SSRSource uint32

const (
	// SSRSourceRTCM identifies RTCM SSR corrections.
	SSRSourceRTCM SSRSource = SSRSource(native.SSRSourceRTCMValue)
	// SSRSourceGalileoHAS identifies Galileo High Accuracy Service corrections.
	SSRSourceGalileoHAS SSRSource = SSRSource(native.SSRSourceGalileoHASValue)
)

// SSRReferencePoint selects the datum for orbit corrections.
type SSRReferencePoint uint32

const (
	// SSRReferencePointAntennaPhaseCenter uses the antenna phase center datum.
	SSRReferencePointAntennaPhaseCenter SSRReferencePoint = SSRReferencePoint(native.SSRReferencePointAntennaValue)
	// SSRReferencePointCenterOfMass uses the satellite center-of-mass datum.
	SSRReferencePointCenterOfMass SSRReferencePoint = SSRReferencePoint(native.SSRReferencePointCenterOfMassValue)
)

// SSRClockCorrection contains copied clock-polynomial corrections. Distances
// are metres, rates are metres per second, and epochs are J2000 seconds.
type SSRClockCorrection struct {
	Source                                                       SSRSource
	ProviderID                                                   uint16
	SolutionID, IODSSR                                           uint8
	C0M, C1MPerS, C2MPerS2, RefEpochJ2000S, UpdateIntervalS      float64
	HasHighRate                                                  bool
	HighRateC0M, HighRateRefEpochJ2000S, HighRateUpdateIntervalS float64
}

// SSROrbitCorrection contains copied orbit corrections in the radial,
// along-track, and cross-track frame. Distances are metres and epochs are
// J2000 seconds.
type SSROrbitCorrection struct {
	Source                                                                                                    SSRSource
	ProviderID                                                                                                uint16
	SolutionID                                                                                                uint8
	IODE                                                                                                      uint32
	IODSSR                                                                                                    uint8
	CRSRegional                                                                                               bool
	ReferencePoint                                                                                            SSRReferencePoint
	RadialM, AlongM, CrossM, RadialRateMPerS, AlongRateMPerS, CrossRateMPerS, RefEpochJ2000S, UpdateIntervalS float64
}

// SSRCorrectionStore owns a C-backed SSR correction state machine.
type SSRCorrectionStore struct {
	_      noCopy
	handle *native.SSRCorrectionStore
}

// NewSSRCorrectionStore creates an empty C-owned SSR correction store using
// the selected orbit reference datum.
func NewSSRCorrectionStore(referencePoint SSRReferencePoint) (*SSRCorrectionStore, error) {
	if err := validateSSRReferencePoint(referencePoint); err != nil {
		return nil, err
	}
	h, e := native.NewSSRStore(uint32(referencePoint))
	if e != nil {
		return nil, publicError(e)
	}
	return &SSRCorrectionStore{handle: h}, nil
}

// NewSSRCorrectionStoreFromRTCM creates an SSR store by decoding RTCM SSR
// bytes at a GNSS week/TOW epoch.
func NewSSRCorrectionStoreFromRTCM(data []byte, epoch GNSSWeekTow) (*SSRCorrectionStore, error) {
	week, err := nativeWeek(epoch)
	if err != nil {
		return nil, err
	}
	h, e := native.NewSSRStoreFromRTCM(data, week)
	if e != nil {
		return nil, publicError(e)
	}
	return &SSRCorrectionStore{handle: h}, nil
}

// Close releases the SSR store; repeated calls are safe.
func (s *SSRCorrectionStore) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Ingest adds RTCM messages to the SSR store at a GNSS week/TOW epoch.
func (s *SSRCorrectionStore) Ingest(messages *RTCMMessages, epoch GNSSWeekTow) error {
	if s == nil || s.handle == nil || messages == nil || messages.handle == nil {
		return ErrClosed
	}
	week, err := nativeWeek(epoch)
	if err != nil {
		return err
	}
	return publicError(s.handle.Ingest(messages.handle, week))
}

// Orbit returns a copied optional radial/along-track/cross-track correction.
func (s *SSRCorrectionStore) Orbit(satellite string) (SSROrbitCorrection, bool, error) {
	if s == nil || s.handle == nil {
		return SSROrbitCorrection{}, false, ErrClosed
	}
	v, p, e := s.handle.Orbit(satellite)
	return SSROrbitCorrection{Source: SSRSource(v.Source), ProviderID: v.ProviderID, SolutionID: v.SolutionID, IODE: v.IODE, IODSSR: v.IODSSR, CRSRegional: v.CRSRegional, ReferencePoint: SSRReferencePoint(v.ReferencePoint), RadialM: v.RadialM, AlongM: v.AlongM, CrossM: v.CrossM, RadialRateMPerS: v.RadialRateMPerS, AlongRateMPerS: v.AlongRateMPerS, CrossRateMPerS: v.CrossRateMPerS, RefEpochJ2000S: v.RefEpochJ2000S, UpdateIntervalS: v.UpdateIntervalS}, p, publicError(e)
}

// Clock returns a copied optional clock correction in metres and metre-based
// rates; reference epochs are J2000 seconds.
func (s *SSRCorrectionStore) Clock(satellite string) (SSRClockCorrection, bool, error) {
	if s == nil || s.handle == nil {
		return SSRClockCorrection{}, false, ErrClosed
	}
	v, p, e := s.handle.Clock(satellite)
	return SSRClockCorrection{Source: SSRSource(v.Source), ProviderID: v.ProviderID, SolutionID: v.SolutionID, IODSSR: v.IODSSR, C0M: v.C0M, C1MPerS: v.C1MPerS, C2MPerS2: v.C2MPerS2, RefEpochJ2000S: v.RefEpochJ2000S, UpdateIntervalS: v.UpdateIntervalS, HasHighRate: v.HasHighRate, HighRateC0M: v.HighRateC0M, HighRateRefEpochJ2000S: v.HighRateRefEpochJ2000S, HighRateUpdateIntervalS: v.HighRateUpdateIntervalS}, p, publicError(e)
}

// CodeBias returns an optional code bias in metres.
func (s *SSRCorrectionStore) CodeBias(satellite string, signalID uint8) (float64, bool, error) {
	if s == nil || s.handle == nil {
		return 0, false, ErrClosed
	}
	v, p, e := s.handle.CodeBias(satellite, signalID)
	return v, p, publicError(e)
}

// PhaseBias returns an optional phase bias in metres.
func (s *SSRCorrectionStore) PhaseBias(satellite string, signalID uint8) (float64, bool, error) {
	if s == nil || s.handle == nil {
		return 0, false, ErrClosed
	}
	v, p, e := s.handle.PhaseBias(satellite, signalID)
	return v, p, publicError(e)
}

// URAIndex returns an optional SSR user-range-accuracy index.
func (s *SSRCorrectionStore) URAIndex(satellite string) (uint8, bool, error) {
	if s == nil || s.handle == nil {
		return 0, false, ErrClosed
	}
	v, p, e := s.handle.URA(satellite)
	return v, p, publicError(e)
}

// SSRMissingCorrectionAction selects behavior when an SSR correction is
// absent or stale.
type SSRMissingCorrectionAction uint32

const (
	// SSRDeclineMissingCorrection reports missing correction data.
	SSRDeclineMissingCorrection SSRMissingCorrectionAction = SSRMissingCorrectionAction(native.SSRMissingDeclineValue)
	// SSRFallbackToBroadcast uses the broadcast state when SSR is unavailable.
	SSRFallbackToBroadcast SSRMissingCorrectionAction = SSRMissingCorrectionAction(native.SSRMissingFallbackValue)
)

// CorrectedState returns an optional SSR-corrected ECEF position in metres and
// clock correction in seconds at a J2000 epoch.
func (s *SSRCorrectionStore) CorrectedState(b *BroadcastEphemeris, satellite string, tJ2000S, stalenessS float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16) ([3]float64, float64, bool, error) {
	if s == nil || s.handle == nil || b == nil || b.handle == nil {
		return [3]float64{}, 0, false, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return [3]float64{}, 0, false, err
	}
	p, c, present, e := s.handle.CorrectedState(b.handle, satellite, tJ2000S, stalenessS, uint32(missing), allowRegionalProvider, regionalProviderID)
	return p, c, present, publicError(e)
}

// SolveBroadcast solves one epoch through the native SSR-corrected broadcast
// route and returns a detached, value-oriented SPP solution.
func (s *SSRCorrectionStore) SolveBroadcast(b *BroadcastEphemeris, config SPPConfig, stalenessS float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16) (SPPSolution, error) {
	if s == nil || s.handle == nil || b == nil || b.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return SPPSolution{}, err
	}
	v, err := s.handle.SolveBroadcast(b.handle, nativeSPPConfig(config), stalenessS, uint32(missing), allowRegionalProvider, regionalProviderID)
	if err != nil {
		return SPPSolution{}, publicError(err)
	}
	return fromNativeSPPSolution(v), nil
}

// EphemerisSample samples SSR-corrected broadcast states over J2000 seconds.
func (s *SSRCorrectionStore) EphemerisSample(b *BroadcastEphemeris, satellites []string, stalenessS float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, startJ2000S, stopJ2000S, stepS float64) ([]EphemerisSampleRow, error) {
	if s == nil || s.handle == nil || b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return nil, err
	}
	values, err := s.handle.EphemerisSample(b.handle, satellites, stalenessS, uint32(missing), allowRegionalProvider, regionalProviderID, startJ2000S, stopJ2000S, stepS)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]EphemerisSampleRow, len(values))
	for i, value := range values {
		out[i] = fromNativeSample(value)
	}
	return out, nil
}
