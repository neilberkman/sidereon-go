package sidereon

import (
	"fmt"
	"os"

	"sidereon.dev/go/v3/internal/native"
)

// RINEXObservationKind is the C observation classification.
type RINEXObservationKind uint32

// RINEXObservationWriteErrorKind identifies a refused observation serialization.
type RINEXObservationWriteErrorKind uint32

const (
	// RINEXWriteErrorNone indicates a successful serialization.
	RINEXWriteErrorNone RINEXObservationWriteErrorKind = 0
	// RINEXWriteErrorCodeListsNotVersionTwo reports nonuniform version-2 code lists.
	RINEXWriteErrorCodeListsNotVersionTwo RINEXObservationWriteErrorKind = 1
	// RINEXWriteErrorNotVersionTwo reports a downgrade target other than version 2.
	RINEXWriteErrorNotVersionTwo RINEXObservationWriteErrorKind = 2
	// RINEXWriteErrorScaleFactorsInVersionTwo reports scale-factor records in version 2.
	RINEXWriteErrorScaleFactorsInVersionTwo RINEXObservationWriteErrorKind = 3
	// RINEXWriteErrorValuesWithoutCodes reports values beyond declared code lists.
	RINEXWriteErrorValuesWithoutCodes RINEXObservationWriteErrorKind = 4
	// RINEXWriteErrorCountsWithoutCodes reports counts beyond declared code lists.
	RINEXWriteErrorCountsWithoutCodes RINEXObservationWriteErrorKind = 5
	// RINEXWriteErrorCodeListNotStated reports a version-2 code list it cannot write.
	RINEXWriteErrorCodeListNotStated RINEXObservationWriteErrorKind = 6
	// RINEXWriteErrorEpochFlagTooWide reports an epoch flag that does not fit.
	RINEXWriteErrorEpochFlagTooWide RINEXObservationWriteErrorKind = 7
	// RINEXWriteErrorEpochTimeMissing reports an observation epoch without time.
	RINEXWriteErrorEpochTimeMissing RINEXObservationWriteErrorKind = 8
	// RINEXWriteErrorEpochPicosecondsNotInVersion reports unsupported epoch precision.
	RINEXWriteErrorEpochPicosecondsNotInVersion RINEXObservationWriteErrorKind = 9
	// RINEXWriteErrorTooManyObservationTypes reports a code count wider than the format.
	RINEXWriteErrorTooManyObservationTypes RINEXObservationWriteErrorKind = 10
	// RINEXWriteErrorCodeListsNotUnion reports lists inconsistent with event declarations.
	RINEXWriteErrorCodeListsNotUnion RINEXObservationWriteErrorKind = 11
	// RINEXWriteErrorValueOutsideDeclaredList reports a value using an undeclared code.
	RINEXWriteErrorValueOutsideDeclaredList RINEXObservationWriteErrorKind = 12
	// RINEXWriteErrorDeclaredListNotStated reports a version-2 list the format cannot state.
	RINEXWriteErrorDeclaredListNotStated RINEXObservationWriteErrorKind = 13
	// RINEXWriteErrorEventRecordsUnreadable reports an event header that cannot be read.
	RINEXWriteErrorEventRecordsUnreadable RINEXObservationWriteErrorKind = 14
	// RINEXWriteErrorObservableNotRepresentable reports a code unsupported by the target version.
	RINEXWriteErrorObservableNotRepresentable RINEXObservationWriteErrorKind = 15
	// RINEXWriteErrorLeapSecondsTimeSystemNotInVersion reports an unsupported leap-second system.
	RINEXWriteErrorLeapSecondsTimeSystemNotInVersion RINEXObservationWriteErrorKind = 16
	// RINEXWriteErrorInvalidLeapSecondsTimeSystem reports a malformed time-system identifier.
	RINEXWriteErrorInvalidLeapSecondsTimeSystem RINEXObservationWriteErrorKind = 17
	// RINEXWriteErrorReadBackMismatch reports text that does not read back losslessly.
	RINEXWriteErrorReadBackMismatch RINEXObservationWriteErrorKind = 18
)

// RINEXObservationWriteError retains every optional field and text detail from C.
type RINEXObservationWriteError struct {
	// Kind identifies the native refusal; unrecognized numeric values are retained.
	Kind RINEXObservationWriteErrorKind
	// HasSystem marks System as the affected constellation.
	HasSystem bool
	// System is the affected native GNSS system code.
	System uint32
	// HasSatellite marks SatelliteID as the affected satellite.
	HasSatellite bool
	// SatelliteID is the affected satellite token.
	SatelliteID string
	// HasEpochIndex marks EpochIndex as the affected zero-based epoch.
	HasEpochIndex bool
	// EpochIndex is the affected zero-based epoch index.
	EpochIndex uint64
	// HasPosition marks Position as a code-list position.
	HasPosition bool
	// Position is a zero-based code-list offset or the list length.
	Position uint64
	// HasFlag marks Flag as the affected epoch flag.
	HasFlag bool
	// Flag is the epoch flag value.
	Flag uint8
	// HasVersion marks Version as the source or requested output version.
	HasVersion bool
	// Version is the source or requested RINEX version.
	Version float64
	// HasCount marks Count as the affected count.
	HasCount bool
	// Count is the affected record or observation-type count.
	Count uint64
	// HasCodes marks Codes as the constellation's code count.
	HasCodes bool
	// Codes is the constellation's declared code count.
	Codes uint64
	// HasValues marks Values as the affected value or count length.
	HasValues bool
	// Values is the number of values held by the affected row.
	Values uint64
	// HasCode reports whether Code names the refused observation code.
	HasCode bool
	// Code is the refused observation code when present.
	Code string
	// HasDetail reports whether Detail contains a source-specific explanation.
	HasDetail bool
	// Detail is the reader text, time-system identifier, or changed field.
	Detail string
	// Message is the detached native refusal description.
	Message string
}

// Error returns the native refusal message.
func (e *RINEXObservationWriteError) Error() string {
	if e == nil {
		return "sidereon: RINEX observation write refused"
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("sidereon: RINEX observation write refused (kind %d)", e.Kind)
}

// RINEXObservationWriteOutcome retains either written bytes or a typed refusal.
type RINEXObservationWriteOutcome struct {
	// IsOK reports whether native code wrote a lossless product.
	IsOK bool
	// Status is the native status associated with the result.
	Status StatusCode
	// Error is the complete refusal, or nil on success.
	Error *RINEXObservationWriteError
	// Text contains detached RINEX bytes after successful writing.
	Text []byte
}

// RINEXObservationDowngradeChangeKind identifies one ordered RINEX 2 conversion.
// Unknown numeric values are retained for forward compatibility.
type RINEXObservationDowngradeChangeKind uint32

const (
	// RINEXDowngradeCodeRenamed reports a code name replacement.
	RINEXDowngradeCodeRenamed RINEXObservationDowngradeChangeKind = iota
	// RINEXDowngradeCodeMoved reports a code-list index change.
	RINEXDowngradeCodeMoved
	// RINEXDowngradeCodeAdded reports a code inserted for version-2 compatibility.
	RINEXDowngradeCodeAdded
	// RINEXDowngradeCodeListRemoved reports a removed constellation code list.
	RINEXDowngradeCodeListRemoved
	// RINEXDowngradeValueRounded reports a rounded observation value.
	RINEXDowngradeValueRounded
	// RINEXDowngradeCycleSlipRounded reports a rounded cycle-slip value.
	RINEXDowngradeCycleSlipRounded
	// RINEXDowngradeScaleFactorsRemoved reports removed scale-factor records.
	RINEXDowngradeScaleFactorsRemoved
	// RINEXDowngradeEpochPicosecondsRemoved reports discarded epoch picoseconds.
	RINEXDowngradeEpochPicosecondsRemoved
	// RINEXDowngradeClockOffsetRounded reports a rounded receiver clock offset.
	RINEXDowngradeClockOffsetRounded
	// RINEXDowngradeInEventLists wraps a change made inside an event header.
	RINEXDowngradeInEventLists
	// RINEXDowngradeDeprecatedRecordsRemoved reports removed obsolete records.
	RINEXDowngradeDeprecatedRecordsRemoved
	// RINEXDowngradeEventRecordsRewritten reports rewritten event records.
	RINEXDowngradeEventRecordsRewritten
)

// RINEXObservationDowngradeChange retains every field of one ordered downgrade
// change. Presence flags distinguish absent values from valid zero values.
type RINEXObservationDowngradeChange struct {
	Kind            RINEXObservationDowngradeChangeKind
	HasNestedChange bool
	Nested          *RINEXObservationDowngradeChange
	HasSystem       bool
	System          GNSSSystem
	HasEpochIndex   bool
	EpochIndex      int
	HasSatellite    bool
	SatelliteID     string
	HasFromIndex    bool
	FromIndex       int
	HasToIndex      bool
	ToIndex         int
	HasFromValue    bool
	FromValue       float64
	HasToValue      bool
	ToValue         float64
	HasPicoseconds  bool
	Picoseconds     uint32
	HasCount        bool
	Count           int
	HasCode         bool
	Code            string
	HasFromText     bool
	From            string
	HasToText       bool
	To              string
	HasLabel        bool
	Label           string
	Codes           []string
	Records         []string
	FromRecords     []string
	ToRecords       []string
}

// RINEXObservationDowngrade contains an independently owned version-2 product
// and a detached, ordered record of every conversion.
type RINEXObservationDowngrade struct {
	Observation *RINEXObservation
	Changes     []RINEXObservationDowngradeChange
}

const (
	// RINEXObservationPseudorange identifies a pseudorange observable.
	RINEXObservationPseudorange RINEXObservationKind = 0
	// RINEXObservationCarrierPhase identifies a carrier-phase observable.
	RINEXObservationCarrierPhase RINEXObservationKind = 1
	// RINEXObservationDoppler identifies a Doppler observable.
	RINEXObservationDoppler RINEXObservationKind = 2
	// RINEXObservationSignalStrength identifies a signal-strength observable.
	RINEXObservationSignalStrength RINEXObservationKind = 3
	// RINEXObservationUnknown identifies an unrecognized observable.
	RINEXObservationUnknown RINEXObservationKind = 4
)

// RINEXObservationHeader retains optional header presence independently from
// its value. Counts and vectors are copied from C.
type RINEXObservationHeader struct {
	// Version is the RINEX format version.
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

// RINEXObservationHeaderSegment is one detached header snapshot and the first
// epoch index for which it is effective.
type RINEXObservationHeaderSegment struct {
	FirstEpochIndex int
	Header          RINEXObservationHeader
}

// RINEXObservationCode identifies one system/code pair.
type RINEXObservationCode struct {
	// System identifies the constellation; Code is the RINEX observation code.
	System GNSSSystem
	Code   string
}

// RINEXObservationEpoch contains one epoch timestamp and event metadata.
type RINEXObservationEpoch struct {
	// Epoch is the civil observation epoch; Flag is the native event flag.
	Epoch          CivilDateTime
	Flag           uint8
	SatelliteCount int
}

// RINEXObservationValue contains one copied observation value. Value is
// optional; LLI and SSI are loss-of-lock and signal-strength indicators.
type RINEXObservationValue struct {
	// SatelliteID and Code identify the observable; Kind is its classification.
	SatelliteID, Code string
	Kind              RINEXObservationKind
	HasValue          bool
	Value             float64
	LLI, SSI          int32
}

// RINEXPseudorange contains an optional-value-filtered pseudorange in metres.
type RINEXPseudorange struct {
	// SatelliteID identifies the row; PseudorangeM is metres.
	SatelliteID  string
	PseudorangeM float64
}

// RINEXCorrectionStatus reports the status of the phase-shift correction in
// effect for one carrier-phase signal.
type RINEXCorrectionStatus uint32

const (
	// RINEXCorrectionAvailable reports one usable correction, including zero.
	RINEXCorrectionAvailable RINEXCorrectionStatus = 0
	// RINEXCorrectionUnknown reports a header declaration with unknown correction.
	RINEXCorrectionUnknown RINEXCorrectionStatus = 1
	// RINEXCorrectionAmbiguous reports multiple conflicting header corrections.
	RINEXCorrectionAmbiguous RINEXCorrectionStatus = 2
)

// RINEXPhaseShiftCorrection retains one correction from an ambiguous header.
type RINEXPhaseShiftCorrection struct {
	// HasCycles reports whether the header record supplied a correction value.
	HasCycles bool
	// Cycles is the correction in cycles when HasCycles is true.
	Cycles float64
}

// RINEXCarrierPhase contains carrier phase in cycles and derived metres.
// Value, frequency, and wavelength each have independent presence flags.
type RINEXCarrierPhase struct {
	// SatelliteID and Code identify the phase observable.
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
	// PhaseShiftStatus distinguishes available, unknown, and conflicting corrections.
	PhaseShiftStatus RINEXCorrectionStatus
	// PhaseShiftConflictCount is the number of alternatives available by row index.
	PhaseShiftConflictCount int
}

// ReceiverClockPhaseSample contains an optional receiver-clock phase in
// seconds for one observation epoch.
type ReceiverClockPhaseSample struct {
	// HasPhaseS controls PhaseS, which is in seconds.
	HasPhaseS bool
	PhaseS    float64
}

func civilFromNative(value native.NativeCalendarEpoch) CivilDateTime {
	return CivilDateTime{Year: int(value.Year), Month: int(value.Month), Day: int(value.Day), Hour: int(value.Hour), Minute: int(value.Minute), Second: value.Second}
}

func rinexObservationHeaderFromNative(v native.NativeRinexObsHeader) RINEXObservationHeader {
	return RINEXObservationHeader{Version: v.Version, HasApproxPosition: v.HasApproxPosition, ApproxPositionM: v.ApproxPosition, HasAntennaDelta: v.HasAntennaDelta, AntennaDeltaHENM: v.AntennaDelta, HasInterval: v.HasInterval, IntervalS: v.Interval, HasTimeOfFirstObs: v.HasTimeOfFirstObs, TimeOfFirstObs: civilFromNative(v.TimeOfFirstObs), TimeOfFirstObsScale: TimeScale(v.TimeOfFirstObsScale), ObservationCodeCount: v.ObsCodeCount, PhaseShiftCount: v.PhaseShiftCount, ScaleFactorCount: v.ScaleFactorCount, GLONASSSlotCount: v.GLONASSSlotCount, HasMarkerName: v.HasMarkerName, MarkerName: v.MarkerName}
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

// LoadRINEXObservation reads a RINEX observation file with Go and passes only
// its bytes to the native parser.
func LoadRINEXObservation(path string) (*RINEXObservation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseRINEXObservation(data)
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
	return rinexObservationHeaderFromNative(v), nil
}

// HeaderTimeline returns detached header snapshots in file order. The first
// segment always begins at epoch zero; effective event headers add segments.
func (obs *RINEXObservation) HeaderTimeline() ([]RINEXObservationHeaderSegment, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.HeaderTimeline()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXObservationHeaderSegment, len(values))
	for i, value := range values {
		out[i] = RINEXObservationHeaderSegment{FirstEpochIndex: value.FirstEpochIndex, Header: rinexObservationHeaderFromNative(value.Header)}
	}
	return out, nil
}

// HeaderAt returns a detached summary of the header effective at epochIndex.
func (obs *RINEXObservation) HeaderAt(epochIndex int) (RINEXObservationHeader, error) {
	if obs == nil || obs.handle == nil {
		return RINEXObservationHeader{}, ErrClosed
	}
	value, err := obs.handle.HeaderAt(epochIndex)
	if err != nil {
		return RINEXObservationHeader{}, publicError(err)
	}
	return rinexObservationHeaderFromNative(value), nil
}

// SkippedRecords returns the number of input records deliberately skipped by
// the parser because their tokens were not representable.
func (obs *RINEXObservation) SkippedRecords() (int, error) {
	if obs == nil || obs.handle == nil {
		return 0, ErrClosed
	}
	value, err := obs.handle.SkippedRecords()
	return value, publicError(err)
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
		status := RINEXCorrectionStatus(v.PhaseShiftStatus)
		out[i] = RINEXCarrierPhase{SatelliteID: v.SatelliteID, Code: v.Code, HasValueCycles: v.HasValueCycles, ValueCycles: v.ValueCycles, LLI: v.LLI, SSI: v.SSI, HasFrequency: v.HasFrequency, FrequencyHz: v.FrequencyHz, HasWavelength: v.HasWavelength, WavelengthM: v.WavelengthM, HasValueM: v.HasValueM, ValueM: v.ValueM, PhaseShiftCycles: v.PhaseShiftCycles, PhaseShiftStatus: status, PhaseShiftConflictCount: v.ConflictCount}
	}
	return out, nil
}

// CarrierPhaseConflicts returns the ordered correction records for a row
// whose phase-shift status is ambiguous. Other statuses return an empty slice.
func (obs *RINEXObservation) CarrierPhaseConflicts(epochIndex, rowIndex int) ([]RINEXPhaseShiftCorrection, error) {
	if obs == nil || obs.handle == nil {
		return nil, ErrClosed
	}
	values, err := obs.handle.CarrierPhaseConflicts(epochIndex, rowIndex)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]RINEXPhaseShiftCorrection, len(values))
	for i, value := range values {
		out[i] = RINEXPhaseShiftCorrection{HasCycles: value.HasCycles, Cycles: value.Cycles}
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

// RINEXTextWithOutcome returns a lossless write result or its complete typed refusal.
func (obs *RINEXObservation) RINEXTextWithOutcome() (RINEXObservationWriteOutcome, error) {
	if obs == nil || obs.handle == nil {
		return RINEXObservationWriteOutcome{}, ErrClosed
	}
	value, err := obs.handle.RINEXTextWithOutcome()
	if err != nil {
		return RINEXObservationWriteOutcome{}, publicError(err)
	}
	result := RINEXObservationWriteOutcome{IsOK: value.IsOK, Status: StatusCode(value.Status), Text: append([]byte(nil), value.Text...)}
	if !value.IsOK {
		result.Error = rinexObservationWriteErrorFromNative(value)
	}
	return result, nil
}

func rinexObservationWriteErrorFromNative(value native.NativeRinexObsWriteOutcome) *RINEXObservationWriteError {
	e := value.Error
	return &RINEXObservationWriteError{Kind: RINEXObservationWriteErrorKind(e.Kind), HasSystem: e.HasSystem, System: e.System, HasSatellite: e.HasSatellite, SatelliteID: e.SatelliteID, HasEpochIndex: e.HasEpochIndex, EpochIndex: e.EpochIndex, HasPosition: e.HasPosition, Position: e.Position, HasFlag: e.HasFlag, Flag: e.Flag, HasVersion: e.HasVersion, Version: e.Version, HasCount: e.HasCount, Count: e.Count, HasCodes: e.HasCodes, Codes: e.Codes, HasValues: e.HasValues, Values: e.Values, HasCode: e.HasCode, Code: value.Code, HasDetail: e.HasDetail, Detail: value.Detail, Message: value.Message}
}

func rinexObservationDowngradeChangeFromNative(value native.NativeRinexObsDowngradeChange) RINEXObservationDowngradeChange {
	out := RINEXObservationDowngradeChange{Kind: RINEXObservationDowngradeChangeKind(value.Kind), HasNestedChange: value.HasNestedChange, HasSystem: value.HasSystem, System: GNSSSystem(value.System), HasEpochIndex: value.HasEpochIndex, EpochIndex: value.EpochIndex, HasSatellite: value.HasSatellite, SatelliteID: value.SatelliteID, HasFromIndex: value.HasFromIndex, FromIndex: value.FromIndex, HasToIndex: value.HasToIndex, ToIndex: value.ToIndex, HasFromValue: value.HasFromValue, FromValue: value.FromValue, HasToValue: value.HasToValue, ToValue: value.ToValue, HasPicoseconds: value.HasPicoseconds, Picoseconds: value.Picoseconds, HasCount: value.HasCount, Count: value.Count, HasCode: value.HasCode, Code: value.Code, HasFromText: value.HasFromText, From: value.From, HasToText: value.HasToText, To: value.To, HasLabel: value.HasLabel, Label: value.Label, Codes: append([]string(nil), value.Codes...), Records: append([]string(nil), value.Records...), FromRecords: append([]string(nil), value.FromRecords...), ToRecords: append([]string(nil), value.ToRecords...)}
	if value.Nested != nil {
		nested := rinexObservationDowngradeChangeFromNative(*value.Nested)
		out.Nested = &nested
	}
	return out
}

// DowngradeToRINEX2 returns a fresh independently owned version-2 product and
// every ordered conversion. It leaves the source unchanged. Semantic refusals
// are returned as *RINEXObservationWriteError.
func (obs *RINEXObservation) DowngradeToRINEX2(version float64) (RINEXObservationDowngrade, error) {
	if obs == nil || obs.handle == nil {
		return RINEXObservationDowngrade{}, ErrClosed
	}
	product, values, outcome, err := obs.handle.DowngradeToRINEX2(version)
	if err != nil {
		return RINEXObservationDowngrade{}, publicError(err)
	}
	if !outcome.IsOK {
		return RINEXObservationDowngrade{}, rinexObservationWriteErrorFromNative(outcome)
	}
	if product == nil {
		return RINEXObservationDowngrade{}, fmt.Errorf("sidereon: native RINEX downgrade returned no product")
	}
	changes := make([]RINEXObservationDowngradeChange, len(values))
	for i, value := range values {
		changes[i] = rinexObservationDowngradeChangeFromNative(value)
	}
	return RINEXObservationDowngrade{Observation: &RINEXObservation{handle: product}, Changes: changes}, nil
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
	// HasIntervalOverride controls the interval override in IntervalOverrideS.
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
	// The count fields are native record/observation totals.
	TotalEpochRecords, ObservationEpochs, EventRecords, PowerFailureEpochs, SkippedRecords          int
	HasInterval                                                                                     bool
	IntervalS                                                                                       float64
	IntervalSource                                                                                  ObservationQCIntervalSource
	MissingEpochs, DataGapCount, SatelliteCount, SatelliteSignalCount, SystemSignalCount, NoteCount int
}

// ObservationQCDataGap describes a missing-epoch interval; times and deltas
// are civil timestamps and seconds respectively.
type ObservationQCDataGap struct {
	// StartEpoch and EndEpoch bracket the gap; interval fields are seconds.
	StartEpoch, EndEpoch             CivilDateTime
	NominalIntervalS, ObservedDeltaS float64
	MissingEpochs                    int
}

// ObservationQCClockJump describes a receiver-clock discontinuity in seconds.
type ObservationQCClockJump struct {
	// EpochIndex identifies the event; Epoch is its civil time and DeltaS is seconds.
	EpochIndex int
	Epoch      CivilDateTime
	DeltaS     float64
}

// ObservationQCCycleSlips summarizes cycle-slip detection counts.
type ObservationQCCycleSlips struct {
	// Observations, TotalSlips, and SystemCount are native counts.
	Observations, TotalSlips, SystemCount int
	HasObservationsPerSlip                bool
	ObservationsPerSlip                   float64
}

// ObservationQCSystemCycleSlip summarizes cycle slips for one GNSS system.
type ObservationQCSystemCycleSlip struct {
	// System identifies the constellation; count fields are native counts.
	System                 GNSSSystem
	Observations, Slips    int
	HasObservationsPerSlip bool
	ObservationsPerSlip    float64
}

// ObservationQCSatellite summarizes one satellite's observation counts.
type ObservationQCSatellite struct {
	// SatelliteID identifies the satellite; count fields are native counts.
	SatelliteID                               string
	EpochsWithObservations, ValueObservations int
}

// ObservationQCSignal summarizes one signal's values and optional SNR/SSI
// statistics. SNR is in dBHz and SNRN is a sample count.
type ObservationQCSignal struct {
	// SatelliteID, System, and Code identify the signal row.
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
	// N is the sample count; RMSM is multipath RMS in metres.
	N    int
	RMSM float64
}

// ObservationQCSatelliteMultipath contains optional MP1/MP2 statistics for a
// satellite.
type ObservationQCSatelliteMultipath struct {
	// SatelliteID identifies the row; HasMP1 and HasMP2 control metric payloads.
	SatelliteID string
	HasMP1      bool
	MP1         ObservationQCMpStats
	HasMP2      bool
	MP2         ObservationQCMpStats
}

// ObservationQCSystemMultipath contains optional MP1/MP2 statistics for a
// GNSS system.
type ObservationQCSystemMultipath struct {
	// System identifies the constellation; HasMP1 and HasMP2 control payloads.
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

// HasApproxPosition controls ApproxPositionM, which is ECEF metres.
// HasAntennaDelta controls AntennaDeltaHENM in metres.
// HasInterval controls IntervalS in seconds.
// HasTimeOfFirstObs controls TimeOfFirstObs and its scale.
// Counts are native header record counts.
// HasMarkerName controls the optional MarkerName.
// SatelliteCount is the native row count.
// HasValue controls Value; LLI and SSI preserve native indicators.
// HasValueCycles controls ValueCycles, which is in cycles.
// LLI and SSI preserve native indicators.
// HasFrequency and HasWavelength control hertz/metre fields.
// HasValueM controls ValueM; PhaseShiftCycles is in cycles.
// GapFactor and ClockJumpThresholdS are native quality-control thresholds.
// HasInterval controls IntervalS and IntervalSource.
// MissingEpochs through NoteCount are native quality-control counts.
// MissingEpochs is the native missing-epoch count.
// HasObservationsPerSlip controls ObservationsPerSlip.
// HasObservationsPerSlip controls ObservationsPerSlip.
// ValueObservations and SNRN are native counts.
// HasSSI controls SSICounts; HasSNR controls SNR statistics.
// SNRMean, SNRMin, and SNRMax use the native signal-strength units.
// HasSNRStd controls SNRStd.
