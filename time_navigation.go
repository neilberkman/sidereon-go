package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// Ephemeris is an owning, read-only state-propagation result.
type Ephemeris struct {
	_      noCopy
	handle *native.Ephemeris
}

// PropagateState propagates an initial Cartesian state at the requested TDB
// offsets, in seconds. All numerical work is performed by the C engine.
func PropagateState(config PropagationConfig, timesS []float64) (*Ephemeris, error) {
	value, err := native.PropagateState(native.NativePropagationConfig{
		Epoch: config.EpochTDBSeconds, Position: config.PositionKm, Velocity: config.VelocityKmPerS,
		ForceModel: uint32(config.ForceModel), Integrator: uint32(config.Integrator),
		AbsTol: config.AbsTol, RelTol: config.RelTol, InitialStep: config.InitialStepS,
		MinStep: config.MinStepS, MaxStep: config.MaxStepS, MaxSteps: config.MaxSteps,
		MuEnabled: config.MuEnabled, Mu: config.MuKm3S2, HasDrag: config.HasDrag,
		Drag:            native.DragParameters{BCFactorM2PerKg: config.Drag.BCFactorM2PerKg, Weather: nativeSpaceWeather(config.Drag.Weather), CutoffAltitudeKm: config.Drag.CutoffAltitudeKm},
		ForceComponents: nativeForceComponents(config.ForceComponents),
	}, append([]float64(nil), timesS...))
	if err != nil {
		return nil, publicError(err)
	}
	return &Ephemeris{handle: value}, nil
}

// Close releases the native ephemeris. It is safe to call on a nil or already
// closed value.
func (e *Ephemeris) Close() error {
	if e == nil || e.handle == nil {
		return nil
	}
	return publicError(e.handle.Close())
}

// EphemerisEpochCount returns the number of propagated states.
func (e *Ephemeris) EphemerisEpochCount() (int, error) {
	if e == nil || e.handle == nil {
		return 0, ErrClosed
	}
	value, err := e.handle.EpochCount()
	return value, publicError(err)
}

// EpochCount is an alias for EphemerisEpochCount.
func (e *Ephemeris) EpochCount() (int, error) { return e.EphemerisEpochCount() }

// EphemerisStates returns detached propagated Cartesian states.
func (e *Ephemeris) EphemerisStates() ([]CartesianState, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	values, err := e.handle.States()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]CartesianState, len(values))
	for i := range values {
		result[i] = publicCartesian(values[i])
	}
	return result, nil
}

// States returns detached propagated Cartesian states.
func (e *Ephemeris) States() ([]CartesianState, error) { return e.EphemerisStates() }

// EphemerisTimesS returns detached propagated TDB epochs in seconds.
func (e *Ephemeris) EphemerisTimesS() ([]float64, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	values, err := e.handle.TimesS()
	return append([]float64(nil), values...), publicError(err)
}

// TimesS returns detached propagated TDB epochs in seconds.
func (e *Ephemeris) TimesS() ([]float64, error) { return e.EphemerisTimesS() }

// GNSSWeekEpochJulianDayNumber returns a system week epoch JDN when defined.
func GNSSWeekEpochJulianDayNumber(system TimeScale) (int64, bool, error) {
	value, present, err := native.GNSSWeekEpochJulianDayNumber(uint32(system))
	return value, present, publicError(err)
}

// GNSSWeekFromCalendar returns the GNSS week for a calendar date when defined.
func GNSSWeekFromCalendar(system TimeScale, year, month, day int64) (uint32, bool, error) {
	value, present, err := native.GNSSWeekFromCalendar(uint32(system), year, month, day)
	return value, present, publicError(err)
}

// GNSSWeekTowNew constructs a scale-tagged GNSS week and time-of-week value.
func GNSSWeekTowNew(system TimeScale, week uint32, towSeconds float64) (GNSSWeekTow, error) {
	value, err := native.GNSSWeekTowNew(uint32(system), week, towSeconds)
	return GNSSWeekTow{System: TimeScale(value.System), Week: value.Week, TOWSeconds: value.TOWSeconds}, publicError(err)
}

// GNSSWeekTowNormalized normalizes a GNSS time-of-week value through C.
func GNSSWeekTowNormalized(value GNSSWeekTow) (GNSSWeekTow, error) {
	result, err := native.GNSSWeekTowNormalized(native.NativeGnssWeekTow{System: uint32(value.System), Week: value.Week, TOWSeconds: value.TOWSeconds})
	return GNSSWeekTow{System: TimeScale(result.System), Week: result.Week, TOWSeconds: result.TOWSeconds}, publicError(err)
}

// GNSSWeekTowUnrolledWeek applies C's 1024-week rollover interpretation.
func GNSSWeekTowUnrolledWeek(value GNSSWeekTow, rollovers uint32) (uint32, error) {
	result, err := native.GNSSWeekTowUnrolledWeek(native.NativeGnssWeekTow{System: uint32(value.System), Week: value.Week, TOWSeconds: value.TOWSeconds}, rollovers)
	return result, publicError(err)
}

// LeapSecondTableInfo describes the compiled leap-second table.
type LeapSecondTableMetadata struct {
	FirstMJD  int32
	LastMJD   int32
	Entries   int
	SourceLen int
}

// LeapSecondTableInfo returns metadata for the native leap-second table.
func LeapSecondTableInfo() (LeapSecondTableMetadata, error) {
	value, err := native.LeapSecondTableInfo()
	return LeapSecondTableMetadata{FirstMJD: value.FirstMJD, LastMJD: value.LastMJD, Entries: value.Entries, SourceLen: value.SourceLen}, publicError(err)
}

// LeapSecondTableSource returns a detached copy of table provenance text.
func LeapSecondTableSource() ([]byte, error) {
	value, err := native.LeapSecondTableSource()
	return append([]byte(nil), value...), publicError(err)
}

// UT1CoverageInfo describes the compiled UT1 coverage table.
type UT1CoverageMetadata struct {
	FirstMJD  int32
	LastMJD   int32
	FirstJDTT float64
	LastJDTT  float64
	Entries   int
	SourceLen int
}

// UT1CoverageCoversJdTt reports whether a TT Julian date is covered.
func UT1CoverageCoversJdTt(jdTT float64) (bool, error) {
	value, err := native.UT1CoverageCoversJDTT(jdTT)
	return value, publicError(err)
}

// UT1CoverageInfo returns metadata for the compiled UT1 table.
func UT1CoverageInfo() (UT1CoverageMetadata, error) {
	value, err := native.UT1CoverageInfo()
	return UT1CoverageMetadata{FirstMJD: value.FirstMJD, LastMJD: value.LastMJD, FirstJDTT: value.FirstJDTT, LastJDTT: value.LastJDTT, Entries: value.Entries, SourceLen: value.SourceLen}, publicError(err)
}

// UT1CoverageSource returns a detached copy of UT1 table provenance text.
func UT1CoverageSource() ([]byte, error) {
	value, err := native.UT1CoverageSource()
	return append([]byte(nil), value...), publicError(err)
}

// LNAVDecoded contains C-decoded GPS LNAV ephemeris and clock fields.
type LNAVDecoded struct {
	WeekNumber, L2Code, URAIndex, SVHealth, IODC int64
	TGD                                          float64
	TOC                                          int64
	AF0, AF1, AF2                                float64
	IODE                                         int64
	CRS, DeltaN, M0, CUC                         float64
	Eccentricity                                 float64
	CUS, SqrtA                                   float64
	TOE, FitIntervalFlag, AODO                   int64
	CIC, Omega0, CIS, I0, CRC, Omega             float64
	OmegaDot, IDot                               float64
}

// LNAVParams contains GPS LNAV encode parameters in C engineering units.
type LNAVParams struct {
	WeekNumber, L2Code, L2PDataFlag, URAIndex, SVHealth, IODC int64
	TGD                                                       float64
	TOC                                                       int64
	AF0, AF1, AF2                                             float64
	IODE                                                      int64
	CRS, DeltaN, M0, CUC                                      float64
	Eccentricity                                              float64
	CUS, SqrtA                                                float64
	TOE, FitIntervalFlag, AODO                                int64
	CIC, Omega0, CIS, I0, CRC, Omega                          float64
	OmegaDot, IDot                                            float64
}

// LNAVOptions contains the TLM/HOW fields accompanying an LNAV encode.
type LNAVOptions struct {
	TOW, Alert, AntiSpoof, Integrity, TLMMessage int64
}

func publicLNAVDecoded(value native.LnavDecoded) LNAVDecoded {
	return LNAVDecoded{WeekNumber: value.WeekNumber, L2Code: value.L2Code, URAIndex: value.URAIndex, SVHealth: value.SVHealth, IODC: value.IODC, TGD: value.TGD, TOC: value.TOC, AF0: value.AF0, AF1: value.AF1, AF2: value.AF2, IODE: value.IODE, CRS: value.CRS, DeltaN: value.DeltaN, M0: value.M0, CUC: value.CUC, Eccentricity: value.Eccentricity, CUS: value.CUS, SqrtA: value.SqrtA, TOE: value.TOE, FitIntervalFlag: value.FitIntervalFlag, AODO: value.AODO, CIC: value.CIC, Omega0: value.Omega0, CIS: value.CIS, I0: value.I0, CRC: value.CRC, Omega: value.Omega, OmegaDot: value.OmegaDot, IDot: value.IDot}
}

func nativeLNAVParams(value LNAVParams) native.LnavParams {
	return native.LnavParams{WeekNumber: value.WeekNumber, L2Code: value.L2Code, L2PDataFlag: value.L2PDataFlag, URAIndex: value.URAIndex, SVHealth: value.SVHealth, IODC: value.IODC, TGD: value.TGD, TOC: value.TOC, AF0: value.AF0, AF1: value.AF1, AF2: value.AF2, IODE: value.IODE, CRS: value.CRS, DeltaN: value.DeltaN, M0: value.M0, CUC: value.CUC, Eccentricity: value.Eccentricity, CUS: value.CUS, SqrtA: value.SqrtA, TOE: value.TOE, FitIntervalFlag: value.FitIntervalFlag, AODO: value.AODO, CIC: value.CIC, Omega0: value.Omega0, CIS: value.CIS, I0: value.I0, CRC: value.CRC, Omega: value.Omega, OmegaDot: value.OmegaDot, IDot: value.IDot}
}

// LNAVDecode decodes three 300-bit GPS LNAV subframes through C.
func LNAVDecode(sf1, sf2, sf3 []byte) (LNAVDecoded, error) {
	value, err := native.LnavDecode(append([]byte(nil), sf1...), append([]byte(nil), sf2...), append([]byte(nil), sf3...))
	return publicLNAVDecoded(value), publicError(err)
}

// LNAVEncode encodes three detached 300-bit GPS LNAV subframes through C.
func LNAVEncode(params LNAVParams, options LNAVOptions) (sf1, sf2, sf3 []byte, err error) {
	sf1, sf2, sf3, err = native.LnavEncode(nativeLNAVParams(params), native.LnavOptions{TOW: options.TOW, Alert: options.Alert, AntiSpoof: options.AntiSpoof, Integrity: options.Integrity, TLMMessage: options.TLMMessage})
	return sf1, sf2, sf3, publicError(err)
}

// LNAVParity computes the six transmitted parity bits for 24 data bits.
func LNAVParity(data24 []byte, d29Previous, d30Previous byte) ([]byte, error) {
	value, err := native.LnavParity(append([]byte(nil), data24...), d29Previous, d30Previous)
	return append([]byte(nil), value...), publicError(err)
}

// LNAVParityValid verifies one 30-bit LNAV word.
func LNAVParityValid(word30 []byte, d29Previous, d30Previous byte) (bool, error) {
	value, err := native.LnavParityValid(append([]byte(nil), word30...), d29Previous, d30Previous)
	return value, publicError(err)
}

// LNAVSubframeID extracts the subframe identifier from a 30- or 300-bit input.
func LNAVSubframeID(bits []byte) (uint64, error) {
	value, err := native.LnavSubframeID(append([]byte(nil), bits...))
	return value, publicError(err)
}

// LNAVTOW extracts the time-of-week count from a 30- or 300-bit input.
func LNAVTOW(bits []byte) (uint64, error) {
	value, err := native.LnavTOW(append([]byte(nil), bits...))
	return value, publicError(err)
}

// NequickGRay contains the receiver-to-satellite geometry for NeQuick-G.
type NequickGRay struct {
	Month                                        uint8
	UTCHours                                     float64
	StationLonDeg, StationLatDeg, StationHeightM float64
	SatelliteLonDeg, SatelliteLatDeg             float64
	SatelliteHeightM                             float64
}

func nativeNequickGRay(value NequickGRay) native.NequickGRay {
	return native.NequickGRay{Month: value.Month, UTCHours: value.UTCHours, StationLonDeg: value.StationLonDeg, StationLatDeg: value.StationLatDeg, StationHeightM: value.StationHeightM, SatelliteLonDeg: value.SatelliteLonDeg, SatelliteLatDeg: value.SatelliteLatDeg, SatelliteHeightM: value.SatelliteHeightM}
}

// GalileoNequickGNative evaluates the coefficient-driven Galileo delay.
func GalileoNequickGNative(ai0, ai1, ai2, latDeg, lonDeg, elDeg, tGalS, dayOfYear, frequencyHz float64) (float64, error) {
	value, err := native.GalileoNequickGNative(ai0, ai1, ai2, latDeg, lonDeg, elDeg, tGalS, dayOfYear, frequencyHz)
	return value, publicError(err)
}

// NequickGDelayM evaluates full NeQuick-G slant delay in metres.
func NequickGDelayM(ai0, ai1, ai2 float64, ray NequickGRay, frequencyHz float64) (float64, error) {
	value, err := native.NequickGDelayM(ai0, ai1, ai2, nativeNequickGRay(ray), frequencyHz)
	return value, publicError(err)
}

// NequickGStecTecu evaluates full NeQuick-G slant TEC in TECU.
func NequickGStecTecu(ai0, ai1, ai2 float64, ray NequickGRay) (float64, error) {
	value, err := native.NequickGStecTecu(ai0, ai1, ai2, nativeNequickGRay(ray))
	return value, publicError(err)
}

// KlobucharNative evaluates the coefficient-driven GPS Klobuchar delay.
func KlobucharNative(alpha, beta [4]float64, latDeg, lonDeg, azDeg, elDeg, tGPSS, frequencyHz float64) (float64, error) {
	value, err := native.KlobucharNative(alpha, beta, latDeg, lonDeg, azDeg, elDeg, tGPSS, frequencyHz)
	return value, publicError(err)
}
