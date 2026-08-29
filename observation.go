package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RINEXObservationKind is the C observation classification.
type RINEXObservationKind uint32

const (
	RINEXObservationPseudorange    RINEXObservationKind = 0
	RINEXObservationCarrierPhase   RINEXObservationKind = 1
	RINEXObservationDoppler        RINEXObservationKind = 2
	RINEXObservationSignalStrength RINEXObservationKind = 3
	RINEXObservationUnknown        RINEXObservationKind = 4
)

// RINEXObservationHeader retains optional header presence independently from
// its value. Counts and vectors are copied from C.
type RINEXObservationHeader struct {
	Version              float64
	HasApproxPosition    bool
	ApproxPositionM      [3]float64
	HasAntennaDelta      bool
	AntennaDeltaHENM     [3]float64
	HasInterval          bool
	IntervalS            float64
	HasTimeOfFirstObs    bool
	TimeOfFirstObs       CivilDateTime
	TimeOfFirstObsScale  TimeScale
	ObservationCodeCount int
	PhaseShiftCount      int
	ScaleFactorCount     int
	GLONASSSlotCount     int
	HasMarkerName        bool
	MarkerName           string
}

// RINEXObservationCode identifies one system/code pair.
type RINEXObservationCode struct {
	System GNSSSystem
	Code   string
}

// RINEXObservationEpoch contains one epoch timestamp and event metadata.
type RINEXObservationEpoch struct {
	Epoch          CivilDateTime
	Flag           uint8
	SatelliteCount int
}

// RINEXObservationValue contains one copied observation value. Value is
// optional; LLI and SSI are loss-of-lock and signal-strength indicators.
type RINEXObservationValue struct {
	SatelliteID, Code string
	Kind              RINEXObservationKind
	HasValue          bool
	Value             float64
	LLI, SSI          int32
}

// RINEXPseudorange contains an optional-value-filtered pseudorange in metres.
type RINEXPseudorange struct {
	SatelliteID  string
	PseudorangeM float64
}

// RINEXCarrierPhase contains carrier phase in cycles and derived metres.
// Value, frequency, and wavelength each have independent presence flags.
type RINEXCarrierPhase struct {
	SatelliteID, Code string
	HasValueCycles    bool
	ValueCycles       float64
	LLI, SSI          int32
	HasFrequency      bool
	FrequencyHz       float64
	HasWavelength     bool
	WavelengthM       float64
	HasValueM         bool
	ValueM            float64
	PhaseShiftCycles  float64
}

// ReceiverClockPhaseSample contains an optional receiver-clock phase in
// seconds for one observation epoch.
type ReceiverClockPhaseSample struct {
	HasPhaseS bool
	PhaseS    float64
}

func civilFromNative(value native.NativeCalendarEpoch) CivilDateTime {
	return CivilDateTime{Year: int(value.Year), Month: int(value.Month), Day: int(value.Day), Hour: int(value.Hour), Minute: int(value.Minute), Second: value.Second}
}

// RINEXObservation owns a parsed RINEX 3 observation product. Its read
// methods copy native data. Read-only calls may run concurrently with Close;
// Close waits for active calls, as required by the C handle contract.
type RINEXObservation struct {
	_      noCopy
	handle *native.RinexObs
}

// ParseRINEXObservation parses RINEX observation bytes through the native C
// parser and copies the input into C-owned temporary storage.
func ParseRINEXObservation(data []byte) (*RINEXObservation, error) {
	h, err := native.ParseRinexObs(data)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXObservation{handle: h}, nil
}

// LoadRINEXObservation delegates path loading and parsing to the native C ABI,
// preserving its UTF-8, filesystem, and status-error semantics.
func LoadRINEXObservation(path string) (*RINEXObservation, error) {
	h, err := native.LoadRinexObs(path)
	if err != nil {
		return nil, publicError(err)
	}
	return &RINEXObservation{handle: h}, nil
}

// Close releases the observation handle; repeated calls are safe.
func (obs *RINEXObservation) Close() error {
	if obs == nil || obs.handle == nil {
		return nil
	}
	return publicError(obs.handle.Close())
}

// Version returns the RINEX version of the parsed product.
func (obs *RINEXObservation) Version() (float64, error) {
	if obs == nil || obs.handle == nil {
		return 0, ErrClosed
	}
	v, err := obs.handle.Version()
	return v, publicError(err)
}

// Header returns a copied observation header with independent optional flags.
func (obs *RINEXObservation) Header() (RINEXObservationHeader, error) {
	if obs == nil || obs.handle == nil {
		return RINEXObservationHeader{}, ErrClosed
	}
	v, err := obs.handle.Header()
	if err != nil {
		return RINEXObservationHeader{}, publicError(err)
	}
	return RINEXObservationHeader{Version: v.Version, HasApproxPosition: v.HasApproxPosition, ApproxPositionM: v.ApproxPosition, HasAntennaDelta: v.HasAntennaDelta, AntennaDeltaHENM: v.AntennaDelta, HasInterval: v.HasInterval, IntervalS: v.Interval, HasTimeOfFirstObs: v.HasTimeOfFirstObs, TimeOfFirstObs: civilFromNative(v.TimeOfFirstObs), TimeOfFirstObsScale: TimeScale(v.TimeOfFirstObsScale), ObservationCodeCount: v.ObsCodeCount, PhaseShiftCount: v.PhaseShiftCount, ScaleFactorCount: v.ScaleFactorCount, GLONASSSlotCount: v.GLONASSSlotCount, HasMarkerName: v.HasMarkerName, MarkerName: v.MarkerName}, nil
}

// EpochCount returns the number of observation epochs.
func (obs *RINEXObservation) EpochCount() (int, error) {
	if obs == nil || obs.handle == nil {
		return 0, ErrClosed
	}
	v, err := obs.handle.EpochCount()
	return v, publicError(err)
}

// Codes returns detached observation-code descriptors.
func (obs *RINEXObservation) Codes() ([]RINEXObservationCode, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.Codes()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXObservationCode, len(values))
	for i, v := range values {
		out[i] = RINEXObservationCode{System: GNSSSystem(v.System), Code: v.Code}
	}
	return out, nil
}

// Epochs returns detached epoch metadata and civil timestamps.
func (obs *RINEXObservation) Epochs() ([]RINEXObservationEpoch, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.Epochs()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXObservationEpoch, len(values))
	for i, v := range values {
		out[i] = RINEXObservationEpoch{Epoch: civilFromNative(v.Epoch), Flag: v.Flag, SatelliteCount: v.SatelliteCount}
	}
	return out, nil
}

// Values returns detached values for an epoch index.
func (obs *RINEXObservation) Values(epoch int) ([]RINEXObservationValue, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.Values(epoch)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXObservationValue, len(values))
	for i, v := range values {
		out[i] = RINEXObservationValue{SatelliteID: v.SatelliteID, Code: v.Code, Kind: RINEXObservationKind(v.Kind), HasValue: v.HasValue, Value: v.Value, LLI: v.LLI, SSI: v.SSI}
	}
	return out, nil
}

// CarrierPhase returns detached carrier-phase values for an epoch index.
func (obs *RINEXObservation) CarrierPhase(epoch int) ([]RINEXCarrierPhase, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.CarrierPhase(epoch)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXCarrierPhase, len(values))
	for i, v := range values {
		out[i] = RINEXCarrierPhase{SatelliteID: v.SatelliteID, Code: v.Code, HasValueCycles: v.HasValueCycles, ValueCycles: v.ValueCycles, LLI: v.LLI, SSI: v.SSI, HasFrequency: v.HasFrequency, FrequencyHz: v.FrequencyHz, HasWavelength: v.HasWavelength, WavelengthM: v.WavelengthM, HasValueM: v.HasValueM, ValueM: v.ValueM, PhaseShiftCycles: v.PhaseShiftCycles}
	}
	return out, nil
}

// Pseudoranges returns detached pseudoranges in metres for an epoch index.
func (obs *RINEXObservation) Pseudoranges(epoch int) ([]RINEXPseudorange, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.Pseudoranges(epoch)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXPseudorange, len(values))
	for i, v := range values {
		out[i] = RINEXPseudorange{SatelliteID: v.SatelliteID, PseudorangeM: v.PseudorangeM}
	}
	return out, nil
}

// Observation returns one value, presence flag, LLI, and SSI by epoch, token,
// and code.
func (obs *RINEXObservation) Observation(epoch int, satellite, code string) (float64, bool, int32, int32, error) {
	if obs == nil || obs.handle == nil {
		return 0, false, -1, -1, ErrClosed
	}
	v, present, lli, ssi, err := obs.handle.Observation(epoch, satellite, code)
	return v, present, lli, ssi, publicError(err)
}

// ReceiverClockPhaseDeviations returns detached receiver-clock phase samples
// in seconds, preserving absent values through HasPhaseS.
func (obs *RINEXObservation) ReceiverClockPhaseDeviations() ([]ReceiverClockPhaseSample, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.ReceiverClockPhase()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ReceiverClockPhaseSample, len(values))
	for i, v := range values {
		out[i] = ReceiverClockPhaseSample{HasPhaseS: v.HasPhaseS, PhaseS: v.PhaseS}
	}
	return out, nil
}

// RINEXText returns a detached copy of the native observation text.
func (obs *RINEXObservation) RINEXText() ([]byte, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	out, err := obs.handle.Text()
	return out, publicError(err)
}

// RINEXObservationFrequency returns a signal frequency in hertz. A nil
// GLONASS channel means that no channel is present.
func RINEXObservationFrequency(system GNSSSystem, code string, version float64, glonassChannel *int8) (float64, error) {
	if err := validateGNSSSystem(system); err != nil {
		return 0, err
	}
	var channel int8
	has := glonassChannel != nil
	if has {
		channel = *glonassChannel
	}
	v, err := native.RinexObservationFrequency(uint32(system), code, version, has, channel)
	return v, publicError(err)
}

// RINEXObservationWavelength returns a signal wavelength in metres. A nil
// GLONASS channel means that no channel is present.
func RINEXObservationWavelength(system GNSSSystem, code string, version float64, glonassChannel *int8) (float64, error) {
	if err := validateGNSSSystem(system); err != nil {
		return 0, err
	}
	var channel int8
	has := glonassChannel != nil
	if has {
		channel = *glonassChannel
	}
	v, err := native.RinexObservationWavelength(uint32(system), code, version, has, channel)
	return v, publicError(err)
}

// ObservationQCOptions controls interval, gap, and receiver-clock checks.
// IntervalOverrideS and ClockJumpThresholdS are seconds.
type ObservationQCOptions struct {
	HasIntervalOverride                               bool
	IntervalOverrideS, GapFactor, ClockJumpThresholdS float64
}

// ObservationQCIntervalSource identifies the native provenance of a QC
// interval. The C ABI may add source values; callers should preserve unknown
// values rather than treating them as zero.
type ObservationQCIntervalSource uint32

// ObservationQCSummary contains copied QC counts and optional interval
// provenance. All count fields are checked Go ints.
type ObservationQCSummary struct {
	TotalEpochRecords, ObservationEpochs, EventRecords, PowerFailureEpochs, SkippedRecords          int
	HasInterval                                                                                     bool
	IntervalS                                                                                       float64
	IntervalSource                                                                                  ObservationQCIntervalSource
	MissingEpochs, DataGapCount, SatelliteCount, SatelliteSignalCount, SystemSignalCount, NoteCount int
}

// ObservationQCDataGap describes a missing-epoch interval; times and deltas
// are civil timestamps and seconds respectively.
type ObservationQCDataGap struct {
	StartEpoch, EndEpoch             CivilDateTime
	NominalIntervalS, ObservedDeltaS float64
	MissingEpochs                    int
}

// ObservationQCClockJump describes a receiver-clock discontinuity in seconds.
type ObservationQCClockJump struct {
	EpochIndex int
	Epoch      CivilDateTime
	DeltaS     float64
}

// ObservationQCCycleSlips summarizes cycle-slip detection counts.
type ObservationQCCycleSlips struct {
	Observations, TotalSlips, SystemCount int
	HasObservationsPerSlip                bool
	ObservationsPerSlip                   float64
}

// ObservationQCSystemCycleSlip summarizes cycle slips for one GNSS system.
type ObservationQCSystemCycleSlip struct {
	System                 GNSSSystem
	Observations, Slips    int
	HasObservationsPerSlip bool
	ObservationsPerSlip    float64
}

// ObservationQCSatellite summarizes one satellite's observation counts.
type ObservationQCSatellite struct {
	SatelliteID                               string
	EpochsWithObservations, ValueObservations int
}

// ObservationQCSignal summarizes one signal's values and optional SNR/SSI
// statistics. SNR is in dBHz and SNRN is a sample count.
type ObservationQCSignal struct {
	SatelliteID             string
	System                  GNSSSystem
	Code                    string
	ValueObservations       int
	HasSSI                  bool
	SSICounts               [10]uint64
	HasSNR                  bool
	SNRN                    int
	SNRMean, SNRMin, SNRMax float64
	HasSNRStd               bool
	SNRStd                  float64
}

// ObservationQCMpStats contains a multipath sample count and RMS in metres.
type ObservationQCMpStats struct {
	N    int
	RMSM float64
}

// ObservationQCSatelliteMultipath contains optional MP1/MP2 statistics for a
// satellite.
type ObservationQCSatelliteMultipath struct {
	SatelliteID string
	HasMP1      bool
	MP1         ObservationQCMpStats
	HasMP2      bool
	MP2         ObservationQCMpStats
}

// ObservationQCSystemMultipath contains optional MP1/MP2 statistics for a
// GNSS system.
type ObservationQCSystemMultipath struct {
	System GNSSSystem
	HasMP1 bool
	MP1    ObservationQCMpStats
	HasMP2 bool
	MP2    ObservationQCMpStats
}

// ObservationQCReport owns a C-backed observation quality report.
type ObservationQCReport struct {
	_      noCopy
	handle *native.ObservationQcReport
}

func nativeQCOptions(v *ObservationQCOptions) *native.NativeObservationQcOptions {
	if v == nil {
		return nil
	}
	return &native.NativeObservationQcOptions{HasIntervalOverride: v.HasIntervalOverride, IntervalOverride: v.IntervalOverrideS, GapFactor: v.GapFactor, ClockJumpThreshold: v.ClockJumpThresholdS}
}

// NewObservationQCOptions returns native default QC options.
func NewObservationQCOptions() (ObservationQCOptions, error) {
	v, err := native.QcOptionsInit()
	return ObservationQCOptions{HasIntervalOverride: v.HasIntervalOverride, IntervalOverrideS: v.IntervalOverride, GapFactor: v.GapFactor, ClockJumpThresholdS: v.ClockJumpThreshold}, publicError(err)
}

// Quality computes a C-backed QC report from the observation product.
func (obs *RINEXObservation) Quality(options *ObservationQCOptions) (*ObservationQCReport, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	h, err := native.ObservationQCFromObs(obs.handle, nativeQCOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &ObservationQCReport{handle: h}, nil
}

// ParseObservationQC computes a C-backed QC report directly from observation
// bytes copied into native temporary storage.
func ParseObservationQC(data []byte, options *ObservationQCOptions) (*ObservationQCReport, error) {
	h, err := native.ParseObservationQC(data, nativeQCOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &ObservationQCReport{handle: h}, nil
}

// Close releases the QC report; repeated calls are safe.
func (report *ObservationQCReport) Close() error {
	if report == nil || report.handle == nil {
		return nil
	}
	return publicError(report.handle.Close())
}

// Summary returns copied QC counts and interval metadata.
func (report *ObservationQCReport) Summary() (ObservationQCSummary, error) {
	if report == nil || report.handle == nil {
		return ObservationQCSummary{}, ErrClosed
	}
	v, err := report.handle.Summary()
	if err != nil {
		return ObservationQCSummary{}, publicError(err)
	}
	return ObservationQCSummary{TotalEpochRecords: v.TotalEpochRecords, ObservationEpochs: v.ObservationEpochs, EventRecords: v.EventRecords, PowerFailureEpochs: v.PowerFailureEpochs, SkippedRecords: v.SkippedRecords, HasInterval: v.HasInterval, IntervalS: v.Interval, IntervalSource: ObservationQCIntervalSource(v.IntervalSource), MissingEpochs: v.MissingEpochs, DataGapCount: v.DataGapCount, SatelliteCount: v.SatelliteCount, SatelliteSignalCount: v.SatelliteSignalCount, SystemSignalCount: v.SystemSignalCount, NoteCount: v.NoteCount}, nil
}

// Gaps returns detached data-gap records.
func (report *ObservationQCReport) Gaps() ([]ObservationQCDataGap, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	values, err := report.handle.Gaps()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ObservationQCDataGap, len(values))
	for i, v := range values {
		out[i] = ObservationQCDataGap{StartEpoch: civilFromNative(v.Start), EndEpoch: civilFromNative(v.End), NominalIntervalS: v.NominalInterval, ObservedDeltaS: v.ObservedDelta, MissingEpochs: v.MissingEpochs}
	}
	return out, nil
}

// ClockJumps returns detached receiver-clock jumps in seconds.
func (report *ObservationQCReport) ClockJumps() ([]ObservationQCClockJump, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	values, err := report.handle.ClockJumps()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ObservationQCClockJump, len(values))
	for i, v := range values {
		out[i] = ObservationQCClockJump{EpochIndex: v.EpochIndex, Epoch: civilFromNative(v.Epoch), DeltaS: v.DeltaS}
	}
	return out, nil
}

// CycleSlips returns the copied overall cycle-slip summary.
func (report *ObservationQCReport) CycleSlips() (ObservationQCCycleSlips, error) {
	if report == nil || report.handle == nil {
		return ObservationQCCycleSlips{}, ErrClosed
	}
	v, err := report.handle.CycleSlips()
	return ObservationQCCycleSlips{Observations: v.Observations, TotalSlips: v.TotalSlips, SystemCount: v.SystemCount, HasObservationsPerSlip: v.HasObservationsPerSlip, ObservationsPerSlip: v.ObservationsPerSlip}, publicError(err)
}

// CycleSlipSystems returns detached per-system cycle-slip summaries.
func (report *ObservationQCReport) CycleSlipSystems() ([]ObservationQCSystemCycleSlip, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	v, e := report.handle.CycleSlipSystems()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]ObservationQCSystemCycleSlip, len(v))
	for i, x := range v {
		out[i] = ObservationQCSystemCycleSlip{System: GNSSSystem(x.System), Observations: x.Observations, Slips: x.Slips, HasObservationsPerSlip: x.HasObservationsPerSlip, ObservationsPerSlip: x.ObservationsPerSlip}
	}
	return out, nil
}

// Satellites returns detached per-satellite QC summaries.
func (report *ObservationQCReport) Satellites() ([]ObservationQCSatellite, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	v, e := report.handle.Satellites()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]ObservationQCSatellite, len(v))
	for i, x := range v {
		out[i] = ObservationQCSatellite{SatelliteID: x.SatelliteID, EpochsWithObservations: x.EpochsWithObservations, ValueObservations: x.ValueObservations}
	}
	return out, nil
}

// SatelliteSignals returns detached per-satellite signal summaries.
func (report *ObservationQCReport) SatelliteSignals() ([]ObservationQCSignal, error) {
	return report.signals(false)
}

// SystemSignals returns detached per-system signal summaries.
func (report *ObservationQCReport) SystemSignals() ([]ObservationQCSignal, error) {
	return report.signals(true)
}
func (report *ObservationQCReport) signals(system bool) ([]ObservationQCSignal, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	v, e := report.handle.Signals(system)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]ObservationQCSignal, len(v))
	for i, x := range v {
		out[i] = ObservationQCSignal{SatelliteID: x.SatelliteID, System: GNSSSystem(x.System), Code: x.Code, ValueObservations: x.ValueObservations, HasSSI: x.HasSSI, SSICounts: x.SSICounts, HasSNR: x.HasSNR, SNRN: x.SNRN, SNRMean: x.SNRMean, SNRMin: x.SNRMin, SNRMax: x.SNRMax, HasSNRStd: x.HasSNRStd, SNRStd: x.SNRStd}
	}
	return out, nil
}

// SatelliteMultipath returns detached per-satellite multipath summaries.
func (report *ObservationQCReport) SatelliteMultipath() ([]ObservationQCSatelliteMultipath, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	v, e := report.handle.Multipath(false)
	if e != nil {
		return nil, publicError(e)
	}
	values := v.([]native.NativeObservationQcSatelliteMultipath)
	out := make([]ObservationQCSatelliteMultipath, len(values))
	for i, x := range values {
		out[i] = ObservationQCSatelliteMultipath{SatelliteID: x.SatelliteID, HasMP1: x.HasMP1, MP1: ObservationQCMpStats{N: x.MP1.N, RMSM: x.MP1.RMSM}, HasMP2: x.HasMP2, MP2: ObservationQCMpStats{N: x.MP2.N, RMSM: x.MP2.RMSM}}
	}
	return out, nil
}

// SystemMultipath returns detached per-system multipath summaries.
func (report *ObservationQCReport) SystemMultipath() ([]ObservationQCSystemMultipath, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	v, e := report.handle.Multipath(true)
	if e != nil {
		return nil, publicError(e)
	}
	values := v.([]native.NativeObservationQcSystemMultipath)
	out := make([]ObservationQCSystemMultipath, len(values))
	for i, x := range values {
		out[i] = ObservationQCSystemMultipath{System: GNSSSystem(x.System), HasMP1: x.HasMP1, MP1: ObservationQCMpStats{N: x.MP1.N, RMSM: x.MP1.RMSM}, HasMP2: x.HasMP2, MP2: ObservationQCMpStats{N: x.MP2.N, RMSM: x.MP2.RMSM}}
	}
	return out, nil
}

// Text returns a detached plain-text QC report.
func (report *ObservationQCReport) Text() ([]byte, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	out, err := report.handle.Text()
	return out, publicError(err)
}

// HTML returns a detached HTML QC report.
func (report *ObservationQCReport) HTML() ([]byte, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	out, err := report.handle.HTML()
	return out, publicError(err)
}

// JSON returns a detached JSON QC report.
func (report *ObservationQCReport) JSON() ([]byte, error) {
	if report == nil || report.handle == nil {
		return nil, ErrClosed
	}
	out, err := report.handle.JSON()
	return out, publicError(err)
}
