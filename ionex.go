package sidereon

import (
	"errors"
	"fmt"

	"sidereon.dev/go/v3/internal/native"
)

// TECSample is one native IONEX TEC sample.
type TECSample struct {
	TimeScale                             TimeScale
	EpochJ2000S, LatDeg, LonDeg, VTECTECU float64
	// EpochJ2000WholeS preserves an exact integer epoch when its presence bit is set.
	EpochJ2000WholeS int64
	// EpochJ2000WholeSPresent selects the exact integer epoch over EpochJ2000S.
	EpochJ2000WholeSPresent bool
	// HeightOffsetKm is the optional per-sample shell-height offset.
	HeightOffsetKm float64
	// HeightOffsetPresent distinguishes a missing height offset from zero.
	HeightOffsetPresent bool
	// VTECPresent reports whether the sample has a VTEC value when presence is known.
	VTECPresent bool
	// VTECPresenceKnown distinguishes an explicit absent value from a legacy zero-value input.
	VTECPresenceKnown bool
	RMSPresent        bool
	RMSTECU           float64
}

// IONEXWarning is a copied parser finding with exact line and epoch fields.
type IONEXWarning struct {
	// Kind is the native warning kind.
	Kind IONEXWarningKind
	// Line is the one-based source line associated with the warning.
	Line uint64
	// MapNumber is the associated one-based map number, when applicable.
	MapNumber uint64
	// SetByLine identifies the line that supplied a carried exponent.
	SetByLine uint64
	// DeclaredCount is the count declared by the relevant header record.
	DeclaredCount uint64
	// TECMapCount is the number of TEC maps found.
	TECMapCount uint64
	// AllMapCount is the number of maps found, including non-TEC maps.
	AllMapCount uint64
	// DeclaredIntervalS is the declared map interval in seconds.
	DeclaredIntervalS uint32
	// ActualSpacingS is the observed map spacing in seconds.
	ActualSpacingS int64
	// Exponent is the exponent in effect for the warning.
	Exponent int32
	// LatDeg is the associated node latitude in degrees.
	LatDeg float64
	// LonDeg is the associated node longitude in degrees.
	LonDeg float64
	// HasEpochs reports whether the epoch fields are present.
	HasEpochs bool
	// DeclaredEpochJ2000S is the declared epoch in seconds since J2000.
	DeclaredEpochJ2000S float64
	// MapEpochJ2000S is the map epoch in seconds since J2000.
	MapEpochJ2000S float64
	// DeclaredEpochJ2000WholeS preserves the exact declared whole-second epoch.
	DeclaredEpochJ2000WholeS int64
	// MapEpochJ2000WholeS preserves the exact map whole-second epoch.
	MapEpochJ2000WholeS int64
	// Label is the detached warning label.
	Label string
	// Message is the detached warning detail.
	Message string
}

// IONEXWarningKind identifies a non-fatal IONEX parser finding.
type IONEXWarningKind uint32

const (
	// IONEXWarningMissingRecord reports an omitted mandatory header record.
	IONEXWarningMissingRecord IONEXWarningKind = 0
	// IONEXWarningVersionRecordNotFirst reports a misplaced version/type record.
	IONEXWarningVersionRecordNotFirst IONEXWarningKind = 1
	// IONEXWarningEpochMismatch reports declared epochs that differ from map epochs.
	IONEXWarningEpochMismatch IONEXWarningKind = 2
	// IONEXWarningMapCountMismatch reports declared and observed map counts that differ.
	IONEXWarningMapCountMismatch IONEXWarningKind = 3
	// IONEXWarningNotANumberValue reports a NaN data token.
	IONEXWarningNotANumberValue IONEXWarningKind = 4
	// IONEXWarningIntervalMismatch reports declared and observed map intervals that differ.
	IONEXWarningIntervalMismatch IONEXWarningKind = 5
	// IONEXWarningExponentCarriedIntoMap reports an exponent carried across map boundaries.
	IONEXWarningExponentCarriedIntoMap IONEXWarningKind = 6
	// IONEXWarningUnknown preserves a native kind this version does not name.
	IONEXWarningUnknown IONEXWarningKind = 999
)

// TECGridSamples describes a complete IONEX TEC grid.
type TECGridSamples struct {
	TimeScale                                     TimeScale
	MapEpochsJ2000S, LatNodesDeg, LonNodesDeg     []float64
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	TECMAPsTECU                                   []float64
	// TECMAPsPresent optionally distinguishes missing VTEC cells from zero values.
	TECMAPsPresent []bool
	RMSPresent     bool
	RMSMAPsTECU    []float64
	// RMSMAPsPresent optionally distinguishes missing RMS cells from zero values.
	RMSMAPsPresent []bool
	// HeightPresent enables the optional per-map shell-height stack.
	HeightPresent bool
	// HeightMapsKm contains flattened shell heights when HeightPresent is true.
	HeightMapsKm []float64
	// HeightMapsPresent optionally records presence for each shell-height cell.
	HeightMapsPresent []bool
	// Header contains optional metadata to write with the grid.
	Header *IONEXHeaderMetadata
}

// IONEXMappingDeclarationKind reports whether the header declares a mapping function.
type IONEXMappingDeclarationKind uint32

const (
	// IONEXMappingDeclarationDeclared means the header has a mapping record.
	IONEXMappingDeclarationDeclared IONEXMappingDeclarationKind = 0
	// IONEXMappingDeclarationAbsent means no mapping record was declared.
	IONEXMappingDeclarationAbsent IONEXMappingDeclarationKind = 1
)

// IONEXMappingFunctionKind identifies standard or future mapping functions.
type IONEXMappingFunctionKind uint32

const (
	// IONEXMappingFunctionNone means no mapping correction was used.
	IONEXMappingFunctionNone IONEXMappingFunctionKind = 0
	// IONEXMappingFunctionCosZ is the 1/cos(z) mapping function.
	IONEXMappingFunctionCosZ IONEXMappingFunctionKind = 1
	// IONEXMappingFunctionQFactor is the Q-factor mapping function.
	IONEXMappingFunctionQFactor IONEXMappingFunctionKind = 2
	// IONEXMappingFunctionOther preserves a custom mapping code.
	IONEXMappingFunctionOther IONEXMappingFunctionKind = 3
	// IONEXMappingFunctionUnknown preserves a future native discriminant.
	IONEXMappingFunctionUnknown IONEXMappingFunctionKind = 999
)

// IONEXHeaderMetadata is a detached, lossless IONEX header snapshot.
type IONEXHeaderMetadata struct {
	// Version is the IONEX format version.
	Version float64
	// Date is the creation-date header text.
	Date string
	// Program is the program name from the header.
	Program string
	// RunBy is the agency name from the header.
	RunBy string
	// SatelliteSystem retains the header's system text.
	SatelliteSystem string
	// HasSatelliteCount distinguishes an absent count from zero.
	HasSatelliteCount bool
	// SatelliteCount is the declared satellite count.
	SatelliteCount uint32
	// HasStationCount distinguishes an absent count from zero.
	HasStationCount bool
	// StationCount is the declared station count.
	StationCount uint32
	// ElevationCutoffDeg is the elevation cutoff in degrees.
	ElevationCutoffDeg float64
	// IntervalS is the map interval in seconds.
	IntervalS uint32
	// HasMapsInFile distinguishes an absent count from zero.
	HasMapsInFile bool
	// MapsInFile is the declared map count.
	MapsInFile uint32
	// ObservablesUsed is the verbatim observable-description field.
	ObservablesUsed string
	// MappingDeclaration records whether MAPPING FUNCTION was declared.
	MappingDeclaration IONEXMappingDeclarationKind
	// MappingFunction is the recognized or future function kind.
	MappingFunction IONEXMappingFunctionKind
	// MappingFunctionCode supplies a custom code when MappingFunction is Other.
	MappingFunctionCode string
	// Descriptions contains all detached DESCRIPTION lines in order.
	Descriptions []string
	// Comments contains all detached COMMENT lines in order.
	Comments []string
}

// IONEXBuildErrorKind identifies a typed refusal while constructing from TEC samples.
type IONEXBuildErrorKind uint32

const (
	// IONEXBuildErrorNone marks a successful build outcome.
	IONEXBuildErrorNone IONEXBuildErrorKind = 0
	// IONEXBuildErrorEmpty reports that no TEC samples were supplied.
	IONEXBuildErrorEmpty IONEXBuildErrorKind = 1
	// IONEXBuildErrorTooFewNodes reports an axis with fewer than two nodes.
	IONEXBuildErrorTooFewNodes IONEXBuildErrorKind = 2
	// IONEXBuildErrorNonMonotonicLatitude reports latitude nodes in invalid order.
	IONEXBuildErrorNonMonotonicLatitude IONEXBuildErrorKind = 3
	// IONEXBuildErrorNonMonotonicLongitude reports longitude nodes in invalid order.
	IONEXBuildErrorNonMonotonicLongitude IONEXBuildErrorKind = 4
	// IONEXBuildErrorNonMonotonicEpochs reports map epochs in invalid order.
	IONEXBuildErrorNonMonotonicEpochs IONEXBuildErrorKind = 5
	// IONEXBuildErrorEpochNotRepresentable reports a non-whole or out-of-range epoch.
	IONEXBuildErrorEpochNotRepresentable IONEXBuildErrorKind = 6
	// IONEXBuildErrorShapeMismatch reports axis or TEC value counts that disagree.
	IONEXBuildErrorShapeMismatch IONEXBuildErrorKind = 7
	// IONEXBuildErrorRMSCountMismatch reports a mismatched RMS map count.
	IONEXBuildErrorRMSCountMismatch IONEXBuildErrorKind = 8
	// IONEXBuildErrorHeightCountMismatch reports a mismatched shell-height map count.
	IONEXBuildErrorHeightCountMismatch IONEXBuildErrorKind = 9
	// IONEXBuildErrorNonFiniteValue reports a NaN or infinite supplied value.
	IONEXBuildErrorNonFiniteValue IONEXBuildErrorKind = 10
	// IONEXBuildErrorNonPositiveStep reports a zero signed grid step.
	IONEXBuildErrorNonPositiveStep IONEXBuildErrorKind = 11
	// IONEXBuildErrorAxisOutOfRange reports an axis coordinate outside its allowed range.
	IONEXBuildErrorAxisOutOfRange IONEXBuildErrorKind = 12
	// IONEXBuildErrorUnknown retains a future failure discriminant.
	IONEXBuildErrorUnknown IONEXBuildErrorKind = 999
)

// IONEXBuildInput identifies the caller buffer named by a build failure index.
type IONEXBuildInput uint32

const (
	// IONEXBuildInputNone indicates that no input buffer is indexed.
	IONEXBuildInputNone IONEXBuildInput = 0
	// IONEXBuildInputMapEpoch identifies the map-epoch array.
	IONEXBuildInputMapEpoch IONEXBuildInput = 1
	// IONEXBuildInputTECValue identifies the flattened TEC values.
	IONEXBuildInputTECValue IONEXBuildInput = 2
	// IONEXBuildInputRMSValue identifies the flattened RMS values.
	IONEXBuildInputRMSValue IONEXBuildInput = 3
	// IONEXBuildInputHeightValue identifies the flattened height values.
	IONEXBuildInputHeightValue IONEXBuildInput = 4
	// IONEXBuildInputSampleEpoch identifies one sample's epoch.
	IONEXBuildInputSampleEpoch IONEXBuildInput = 5
	// IONEXBuildInputSampleVTEC identifies one sample's VTEC value.
	IONEXBuildInputSampleVTEC IONEXBuildInput = 6
	// IONEXBuildInputSampleRMS identifies one sample's RMS value.
	IONEXBuildInputSampleRMS IONEXBuildInput = 7
	// IONEXBuildInputSampleHeight identifies one sample's height offset.
	IONEXBuildInputSampleHeight IONEXBuildInput = 8
)

// IONEXBuildError retains every fixed-width field of a refused TEC-sample build.
type IONEXBuildError struct {
	// Kind identifies the failure; unknown numeric values are retained.
	Kind IONEXBuildErrorKind
	// Input identifies the indexed caller array, when HasIndex is true.
	Input IONEXBuildInput
	// HasIndex reports whether Index names an entry in a caller array.
	HasIndex bool
	// Index is the zero-based entry in Input.
	Index uint64
	// NodeCount is the short axis length for TooFewNodes failures.
	NodeCount uint64
	// ValueCount is the supplied value count for a mismatch.
	ValueCount uint64
	// ExpectedValueCount is the value count implied by the axes.
	ExpectedValueCount uint64
	// HasAxisValue reports whether AxisValue identifies the refused coordinate or step.
	HasAxisValue bool
	// AxisValue is the refused coordinate or step in degrees.
	AxisValue float64
	// Message is the owned native detail string.
	Message string
}

// Error returns the native refusal detail, or a stable description when it is empty.
func (e *IONEXBuildError) Error() string {
	if e == nil {
		return "sidereon: IONEX build refused"
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("sidereon: IONEX build refused (kind %d)", e.Kind)
}

// IONEXBuildOutcome reports whether the constructor built a product or refused its data.
type IONEXBuildOutcome struct {
	// IsOK reports whether the native constructor built a product.
	IsOK bool
	// Status is the native status associated with the outcome.
	Status StatusCode
	// Error is the complete typed refusal, or nil after a successful build.
	Error *IONEXBuildError
}

func headerMetadataToNative(v *IONEXHeaderMetadata) *native.IonexHeaderMetadata {
	if v == nil {
		return nil
	}
	return &native.IonexHeaderMetadata{Version: v.Version, Date: v.Date, Program: v.Program, RunBy: v.RunBy, SatelliteSystem: v.SatelliteSystem, HasSatelliteCount: v.HasSatelliteCount, SatelliteCount: v.SatelliteCount, HasStationCount: v.HasStationCount, StationCount: v.StationCount, ElevationCutoffDeg: v.ElevationCutoffDeg, IntervalS: v.IntervalS, HasMapsInFile: v.HasMapsInFile, MapsInFile: v.MapsInFile, ObservablesUsed: v.ObservablesUsed, MappingDeclaration: uint32(v.MappingDeclaration), MappingFunction: uint32(v.MappingFunction), MappingFunctionCode: v.MappingFunctionCode, Descriptions: append([]string(nil), v.Descriptions...), Comments: append([]string(nil), v.Comments...)}
}

// TECGridSamplesInfo reports IONEX grid dimensions and metadata.
type TECGridSamplesInfo struct {
	MapEpochCount, LatNodeCount, LonNodeCount     int
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	RMSPresent                                    bool
	TECMAPValueCount, RMSMAPValueCount            int
}

// IONEXSlantDelayEvaluation contains a delay and its coverage status.
type IONEXSlantDelayEvaluation struct {
	DelayM                                           float64
	Status                                           IONEXSlantDelayStatus
	CoverageError                                    IONEXCoverageErrorKind
	IsValid, HasHeld, HasDegraded, HasAssumedMapping bool
	Gap                                              IONEXNodeGap
	AssumedMapping                                   IONEXAssumedMappingKind
}

// IONEXNodeGap preserves all weighted missing-node masks from both maps.
type IONEXNodeGap struct {
	HasGap         bool
	Earlier, Later [4]bool
}

// IONEXEpochErrorKind identifies why an exact instant cannot be mapped to UTC.
type IONEXEpochErrorKind uint32

const (
	// IONEXEpochErrorNone means no error or special condition occurred.
	IONEXEpochErrorNone IONEXEpochErrorKind = 0
	// IONEXEpochErrorNotWholeSecond means the query epoch is not an exact whole second.
	IONEXEpochErrorNotWholeSecond IONEXEpochErrorKind = 1
	// IONEXEpochErrorFractionalUTCSecond means UTC conversion would require a fractional second.
	IONEXEpochErrorFractionalUTCSecond IONEXEpochErrorKind = 2
	// IONEXEpochErrorNoExactUTCOffset means no exact UTC offset is available at the query epoch.
	IONEXEpochErrorNoExactUTCOffset IONEXEpochErrorKind = 3
	// IONEXEpochErrorInsertedLeapSecond means the query lies within an inserted UTC leap second.
	IONEXEpochErrorInsertedLeapSecond IONEXEpochErrorKind = 4
	// IONEXEpochErrorBeforeIntegerLeapSeconds means the query precedes the integer leap-second table.
	IONEXEpochErrorBeforeIntegerLeapSeconds IONEXEpochErrorKind = 5
	// IONEXEpochErrorOutOfRange means the epoch lies outside the IONEX conversion range.
	IONEXEpochErrorOutOfRange IONEXEpochErrorKind = 6
	// IONEXEpochErrorYearOutOfField means the year does not fit the IONEX epoch field.
	IONEXEpochErrorYearOutOfField IONEXEpochErrorKind = 7
	// IONEXEpochErrorUnknown preserves a native value not recognized by this version.
	IONEXEpochErrorUnknown IONEXEpochErrorKind = 999
)

// IONEXEpochError retains the conversion cause independently of other row errors.
type IONEXEpochError struct {
	Kind         IONEXEpochErrorKind
	Scale        TimeScale
	HasUTCJ2000S bool
	UTCJ2000S    int64
}

// IONEXSlantPolicy controls coverage, missing-node, and mapping behavior.
type IONEXSlantPolicy struct {
	Coverage     IONEXCoveragePolicy
	MissingNodes IONEXMissingNodePolicy
	Mapping      IONEXMappingPolicy
}

// IONEXMissingNodePolicy selects how interpolation handles missing grid nodes.
type IONEXMissingNodePolicy uint32

const (
	// IONEXMissingNodesStrict requires every interpolation node to be available.
	IONEXMissingNodesStrict IONEXMissingNodePolicy = 0
	// IONEXMissingNodesRenormalize renormalizes weights across the available interpolation nodes.
	IONEXMissingNodesRenormalize IONEXMissingNodePolicy = 1
)

// IONEXMappingPolicy selects whether the file mapping or a single-layer mapping is used.
type IONEXMappingPolicy uint32

const (
	// IONEXMappingDeclared uses the mapping function declared by the IONEX product.
	IONEXMappingDeclared IONEXMappingPolicy = 0
	// IONEXMappingSingleLayer uses the single-layer mapping function.
	IONEXMappingSingleLayer IONEXMappingPolicy = 1
)

// IONEXAssumedMappingKind records the mapping assumption applied to a slant-delay query.
type IONEXAssumedMappingKind uint32

const (
	// IONEXAssumedMappingNone means no mapping assumption was applied.
	IONEXAssumedMappingNone IONEXAssumedMappingKind = 0
	// IONEXAssumedMappingOther records a mapping assumption outside the named cases.
	IONEXAssumedMappingOther IONEXAssumedMappingKind = 1
	// IONEXAssumedMappingAbsent means the product omitted mapping metadata.
	IONEXAssumedMappingAbsent IONEXAssumedMappingKind = 2
)

// IONEXSlantErrorKind identifies the typed refusal for one slant query.
type IONEXSlantErrorKind uint32

const (
	// IONEXSlantErrorNone means no error or special condition occurred.
	IONEXSlantErrorNone IONEXSlantErrorKind = 0
	// IONEXSlantErrorInvalidInput identifies invalid caller input.
	IONEXSlantErrorInvalidInput IONEXSlantErrorKind = 1
	// IONEXSlantErrorOutOfCoverage identifies a query outside available model coverage.
	IONEXSlantErrorOutOfCoverage IONEXSlantErrorKind = 2
	// IONEXSlantErrorNodesNotAvailable identifies interpolation nodes that are missing or unusable.
	IONEXSlantErrorNodesNotAvailable IONEXSlantErrorKind = 3
	// IONEXSlantErrorUnavailable identifies a result that cannot be computed from the available data.
	IONEXSlantErrorUnavailable IONEXSlantErrorKind = 4
	// IONEXSlantErrorUnknown preserves a native value not recognized by this version.
	IONEXSlantErrorUnknown IONEXSlantErrorKind = 999
)

// IONEXSlantError retains every typed cause field returned by the C ABI.
type IONEXSlantError struct {
	Kind                                               IONEXSlantErrorKind
	CoverageError                                      IONEXCoverageErrorKind
	HasGap                                             bool
	Gap                                                IONEXNodeGap
	Refusal                                            uint32
	RefusalMapNumber, RefusalLatIndex, RefusalLonIndex uint64
	HasMappingDeclaration                              bool
	MappingDeclaration                                 uint32
	HasMappingFunction                                 bool
	MappingFunction                                    uint32
}

// IONEXInstantSlantRequest contains one slant query at a lossless scale-tagged instant.
type IONEXInstantSlantRequest struct {
	LatDeg, LonDeg, AzimuthDeg, ElevationDeg float64
	Epoch                                    ClockEpoch
	FrequencyHz                              float64
}

// IONEXInstantSlantResult retains either the value or full typed row failure.
type IONEXInstantSlantResult struct {
	IsOK        bool
	Status      StatusCode
	Evaluation  IONEXSlantDelayEvaluation
	Error       *IONEXSlantError
	EpochError  *IONEXEpochError
	Message     string
	MappingCode string
}

// IONEXInstantSlantResultList owns detached, input-ordered exact-time results.
type IONEXInstantSlantResultList struct{ Rows []IONEXInstantSlantResult }

// IONEX owns a parsed or synthesized IONEX product.
type IONEX struct {
	_      noCopy
	handle *native.Ionex
}

func publicTECSample(v native.TecSample) TECSample {
	return TECSample{TimeScale: TimeScale(v.TimeScale), EpochJ2000S: v.EpochJ2000S, EpochJ2000WholeS: v.EpochJ2000WholeS, EpochJ2000WholeSPresent: v.EpochJ2000WholeSPresent, LatDeg: v.LatDeg, LonDeg: v.LonDeg, VTECTECU: v.VTECTECU, VTECPresent: v.VTECPresent, VTECPresenceKnown: v.VTECPresenceKnown, RMSPresent: v.RMSPresent, RMSTECU: v.RMSTECU, HeightOffsetKm: v.HeightOffsetKm, HeightOffsetPresent: v.HeightOffsetPresent}
}

// ParseIONEX parses an IONEX product with the native parser.
func ParseIONEX(data []byte) (*IONEX, error) {
	v, e := native.ParseIONEX(data)
	if e != nil {
		return nil, publicError(e)
	}
	if v == nil {
		return nil, errNilNativeHandle
	}
	return &IONEX{handle: v}, nil
}

// ParseIONEXWithWarnings parses a product and retains every parser warning.
// The warning strings and metadata are detached from the native warning list.
func ParseIONEXWithWarnings(data []byte) (*IONEX, []IONEXWarning, error) {
	v, warnings, e := native.ParseIONEXWithWarnings(data)
	if e != nil {
		return nil, nil, publicError(e)
	}
	if v == nil {
		return nil, nil, errNilNativeHandle
	}
	out := make([]IONEXWarning, len(warnings))
	for j, w := range warnings {
		out[j] = IONEXWarning{Kind: IONEXWarningKind(w.Kind), Line: w.Line, MapNumber: w.MapNumber, SetByLine: w.SetByLine, DeclaredCount: w.DeclaredCount, TECMapCount: w.TECMapCount, AllMapCount: w.AllMapCount, DeclaredIntervalS: w.DeclaredIntervalS, ActualSpacingS: w.ActualSpacingS, Exponent: w.Exponent, LatDeg: w.LatDeg, LonDeg: w.LonDeg, HasEpochs: w.HasEpochs, DeclaredEpochJ2000S: w.DeclaredEpochJ2000S, MapEpochJ2000S: w.MapEpochJ2000S, DeclaredEpochJ2000WholeS: w.DeclaredEpochJ2000WholeS, MapEpochJ2000WholeS: w.MapEpochJ2000WholeS, Label: w.Label, Message: w.Message}
	}
	return &IONEX{handle: v}, out, nil
}

// NewIONEXFromTECGridSamples builds an IONEX product from a complete grid.
func NewIONEXFromTECGridSamples(v TECGridSamples) (*IONEX, error) {
	x := native.TecGridSamples{TimeScale: uint32(v.TimeScale), MapEpochsJ2000S: append([]float64(nil), v.MapEpochsJ2000S...), LatNodesDeg: append([]float64(nil), v.LatNodesDeg...), LonNodesDeg: append([]float64(nil), v.LonNodesDeg...), DLatDeg: v.DLatDeg, DLonDeg: v.DLonDeg, ShellHeightKm: v.ShellHeightKm, BaseRadiusKm: v.BaseRadiusKm, Exponent: v.Exponent, TECMAPsTECU: append([]float64(nil), v.TECMAPsTECU...), TECMAPsPresent: append([]bool(nil), v.TECMAPsPresent...), RMSPresent: v.RMSPresent, RMSMAPsTECU: append([]float64(nil), v.RMSMAPsTECU...), RMSMAPsPresent: append([]bool(nil), v.RMSMAPsPresent...), HeightPresent: v.HeightPresent, HeightMapsKm: append([]float64(nil), v.HeightMapsKm...), HeightMapsPresent: append([]bool(nil), v.HeightMapsPresent...), Header: headerMetadataToNative(v.Header)}
	out, e := native.BuildIONEXFromTECGridSamples(x)
	if e != nil {
		return nil, publicError(e)
	}
	return &IONEX{handle: out}, nil
}

// NewIONEXFromTECGridSamplesWithOutcome returns typed sample refusals separately from call errors.
func NewIONEXFromTECGridSamplesWithOutcome(v TECGridSamples) (*IONEX, IONEXBuildOutcome, error) {
	x := native.TecGridSamples{TimeScale: uint32(v.TimeScale), MapEpochsJ2000S: append([]float64(nil), v.MapEpochsJ2000S...), LatNodesDeg: append([]float64(nil), v.LatNodesDeg...), LonNodesDeg: append([]float64(nil), v.LonNodesDeg...), DLatDeg: v.DLatDeg, DLonDeg: v.DLonDeg, ShellHeightKm: v.ShellHeightKm, BaseRadiusKm: v.BaseRadiusKm, Exponent: v.Exponent, TECMAPsTECU: append([]float64(nil), v.TECMAPsTECU...), TECMAPsPresent: append([]bool(nil), v.TECMAPsPresent...), RMSPresent: v.RMSPresent, RMSMAPsTECU: append([]float64(nil), v.RMSMAPsTECU...), RMSMAPsPresent: append([]bool(nil), v.RMSMAPsPresent...), HeightPresent: v.HeightPresent, HeightMapsKm: append([]float64(nil), v.HeightMapsKm...), HeightMapsPresent: append([]bool(nil), v.HeightMapsPresent...), Header: headerMetadataToNative(v.Header)}
	product, outcome, err := native.BuildIONEXFromTECGridSamplesWithOutcome(x)
	if err != nil {
		return nil, IONEXBuildOutcome{}, publicError(err)
	}
	return publicIONEXBuildResult(product, outcome)
}

// NewIONEXFromTECSamples builds an IONEX product from grid-node samples.
func NewIONEXFromTECSamples(v []TECSample, shellHeightKm, baseRadiusKm float64, exponent int32) (*IONEX, error) {
	x := make([]native.TecSample, len(v))
	for i, s := range v {
		x[i] = native.TecSample{TimeScale: uint32(s.TimeScale), EpochJ2000S: s.EpochJ2000S, EpochJ2000WholeS: s.EpochJ2000WholeS, EpochJ2000WholeSPresent: s.EpochJ2000WholeSPresent, LatDeg: s.LatDeg, LonDeg: s.LonDeg, VTECTECU: s.VTECTECU, VTECPresent: s.VTECPresent, VTECPresenceKnown: s.VTECPresenceKnown, RMSPresent: s.RMSPresent, RMSTECU: s.RMSTECU, HeightOffsetKm: s.HeightOffsetKm, HeightOffsetPresent: s.HeightOffsetPresent}
	}
	out, e := native.BuildIONEXFromTECSamples(x, shellHeightKm, baseRadiusKm, exponent)
	if e != nil {
		return nil, publicError(e)
	}
	return &IONEX{handle: out}, nil
}

// NewIONEXFromTECSamplesWithHeader builds an IONEX product from grid-node
// samples and a detached header value. The header is copied into native
// storage during the call; later changes to header do not affect the product.
func NewIONEXFromTECSamplesWithHeader(v []TECSample, shellHeightKm, baseRadiusKm float64, exponent int32, header IONEXHeaderMetadata) (*IONEX, error) {
	x := make([]native.TecSample, len(v))
	for i, s := range v {
		x[i] = native.TecSample{TimeScale: uint32(s.TimeScale), EpochJ2000S: s.EpochJ2000S, EpochJ2000WholeS: s.EpochJ2000WholeS, EpochJ2000WholeSPresent: s.EpochJ2000WholeSPresent, LatDeg: s.LatDeg, LonDeg: s.LonDeg, VTECTECU: s.VTECTECU, VTECPresent: s.VTECPresent, VTECPresenceKnown: s.VTECPresenceKnown, RMSPresent: s.RMSPresent, RMSTECU: s.RMSTECU, HeightOffsetKm: s.HeightOffsetKm, HeightOffsetPresent: s.HeightOffsetPresent}
	}
	h := headerMetadataToNative(&header)
	out, e := native.BuildIONEXFromTECSamplesWithHeader(x, shellHeightKm, baseRadiusKm, exponent, h)
	if e != nil {
		return nil, publicError(e)
	}
	if out == nil {
		return nil, errNilNativeHandle
	}
	return &IONEX{handle: out}, nil
}

// NewIONEXFromTECSamplesWithOutcome returns typed per-node sample refusals separately from call errors.
func NewIONEXFromTECSamplesWithOutcome(v []TECSample, shellHeightKm, baseRadiusKm float64, exponent int32) (*IONEX, IONEXBuildOutcome, error) {
	x := nativeTECSamples(v)
	product, outcome, err := native.BuildIONEXFromTECSamplesWithOutcome(x, shellHeightKm, baseRadiusKm, exponent)
	if err != nil {
		return nil, IONEXBuildOutcome{}, publicError(err)
	}
	return publicIONEXBuildResult(product, outcome)
}

// NewIONEXFromTECSamplesWithHeaderOutcome includes header metadata in a typed build outcome.
func NewIONEXFromTECSamplesWithHeaderOutcome(v []TECSample, shellHeightKm, baseRadiusKm float64, exponent int32, header IONEXHeaderMetadata) (*IONEX, IONEXBuildOutcome, error) {
	x := nativeTECSamples(v)
	h := headerMetadataToNative(&header)
	product, outcome, err := native.BuildIONEXFromTECSamplesWithHeaderOutcome(x, shellHeightKm, baseRadiusKm, exponent, h)
	if err != nil {
		return nil, IONEXBuildOutcome{}, publicError(err)
	}
	return publicIONEXBuildResult(product, outcome)
}

func nativeTECSamples(v []TECSample) []native.TecSample {
	x := make([]native.TecSample, len(v))
	for i, s := range v {
		x[i] = native.TecSample{TimeScale: uint32(s.TimeScale), EpochJ2000S: s.EpochJ2000S, EpochJ2000WholeS: s.EpochJ2000WholeS, EpochJ2000WholeSPresent: s.EpochJ2000WholeSPresent, LatDeg: s.LatDeg, LonDeg: s.LonDeg, VTECTECU: s.VTECTECU, VTECPresent: s.VTECPresent, VTECPresenceKnown: s.VTECPresenceKnown, RMSPresent: s.RMSPresent, RMSTECU: s.RMSTECU, HeightOffsetKm: s.HeightOffsetKm, HeightOffsetPresent: s.HeightOffsetPresent}
	}
	return x
}

func publicIONEXBuildResult(product *native.Ionex, outcome native.TecSamplesOutcome) (*IONEX, IONEXBuildOutcome, error) {
	result := IONEXBuildOutcome{IsOK: outcome.IsOK, Status: StatusCode(outcome.Status)}
	if !outcome.IsOK {
		result.Error = &IONEXBuildError{Kind: IONEXBuildErrorKind(outcome.Kind), Input: IONEXBuildInput(outcome.Input), HasIndex: outcome.HasIndex, Index: outcome.Index, NodeCount: outcome.NodeCount, ValueCount: outcome.ValueCount, ExpectedValueCount: outcome.ExpectedValueCount, HasAxisValue: outcome.HasAxisValue, AxisValue: outcome.AxisValue, Message: outcome.Message}
		if product != nil {
			_ = product.Close()
			return nil, IONEXBuildOutcome{}, errors.New("sidereon: refused IONEX build unexpectedly returned a product")
		}
		return nil, result, nil
	}
	if product == nil {
		return nil, IONEXBuildOutcome{}, errNilNativeHandle
	}
	return &IONEX{handle: product}, result, nil
}

// Close releases the IONEX product; it is safe to call more than once.
func (i *IONEX) Close() error {
	if i == nil || i.handle == nil {
		return nil
	}
	return publicError(i.handle.Close())
}

// EpochCount returns the number of TEC maps.
func (i *IONEX) EpochCount() (int, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, e := i.handle.EpochCount()
	return v, publicError(e)
}

// SkippedRecords reports non-map input records ignored by the parser.
func (i *IONEX) SkippedRecords() (int, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, e := i.handle.SkippedRecords()
	return v, publicError(e)
}

// Exponent returns the IONEX TEC exponent.
func (i *IONEX) Exponent() (int32, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, e := i.handle.Exponent()
	return v, publicError(e)
}

// LatNodesDeg returns the latitude node axis in degrees.
func (i *IONEX) LatNodesDeg() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.LatNodesDeg()
	return v, publicError(e)
}

// LonNodesDeg returns the longitude node axis in degrees.
func (i *IONEX) LonNodesDeg() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.LonNodesDeg()
	return v, publicError(e)
}

// MapEpochsJ2000S returns map epochs as integer J2000 seconds.
func (i *IONEX) MapEpochsJ2000S() ([]int64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.MapEpochsJ2000S()
	return v, publicError(e)
}

// ToIONEXText serializes the product using the native IONEX writer.
func (i *IONEX) ToIONEXText() ([]byte, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.ToIONEXText()
	return v, publicError(e)
}

// HeaderMetadata returns detached header values and records from the product.
func (i *IONEX) HeaderMetadata() (IONEXHeaderMetadata, error) {
	if i == nil || i.handle == nil {
		return IONEXHeaderMetadata{}, ErrClosed
	}
	v, e := i.handle.HeaderMetadata()
	if e != nil {
		return IONEXHeaderMetadata{}, publicError(e)
	}
	return IONEXHeaderMetadata{Version: v.Version, Date: v.Date, Program: v.Program, RunBy: v.RunBy, SatelliteSystem: v.SatelliteSystem, HasSatelliteCount: v.HasSatelliteCount, SatelliteCount: v.SatelliteCount, HasStationCount: v.HasStationCount, StationCount: v.StationCount, ElevationCutoffDeg: v.ElevationCutoffDeg, IntervalS: v.IntervalS, HasMapsInFile: v.HasMapsInFile, MapsInFile: v.MapsInFile, ObservablesUsed: v.ObservablesUsed, MappingDeclaration: IONEXMappingDeclarationKind(v.MappingDeclaration), MappingFunction: IONEXMappingFunctionKind(v.MappingFunction), MappingFunctionCode: v.MappingFunctionCode, Descriptions: append([]string(nil), v.Descriptions...), Comments: append([]string(nil), v.Comments...)}, nil
}

// IONEXSlantRequest contains one whole-second UTC IONEX slant-delay query.
type IONEXSlantRequest struct {
	// LatDeg is the receiver geodetic latitude in degrees.
	LatDeg float64
	// LonDeg is the receiver geodetic longitude in degrees.
	LonDeg float64
	// AzimuthDeg is the satellite azimuth in degrees.
	AzimuthDeg float64
	// ElevationDeg is the satellite elevation above the horizon in degrees.
	ElevationDeg float64
	// EpochJ2000S is UTC seconds from J2000.
	EpochJ2000S int64
	// FrequencyHz is the carrier frequency in hertz.
	FrequencyHz float64
}

// SlantDelay evaluates one IONEX slant delay under the default strict policy.
func (i *IONEX) SlantDelay(lat, lon, azimuth, elevation float64, epochJ2000S int64, frequencyHz float64) (float64, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	v, e := i.handle.SlantDelay(lat, lon, azimuth, elevation, epochJ2000S, frequencyHz)
	return v, publicError(e)
}

// SlantDelays evaluates a batch under the default strict policy. Results keep
// request order; if any request is refused, all returned delay values are zero.
func (i *IONEX) SlantDelays(requests []IONEXSlantRequest) ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	nativeRequests := make([]native.IonexSlantRequest, len(requests))
	for index, value := range requests {
		nativeRequests[index] = native.IonexSlantRequest{LatDeg: value.LatDeg, LonDeg: value.LonDeg, AzimuthDeg: value.AzimuthDeg, ElevationDeg: value.ElevationDeg, EpochJ2000S: value.EpochJ2000S, FrequencyHz: value.FrequencyHz}
	}
	values, err := i.handle.SlantDelays(nativeRequests)
	return values, publicError(err)
}

// SlantDelayWithPolicy evaluates slant delay with an explicit coverage policy.
func (i *IONEX) SlantDelayWithPolicy(lat, lon, azimuth, elevation float64, epochJ2000S int64, frequencyHz float64, policy IONEXCoveragePolicy) (IONEXSlantDelayEvaluation, error) {
	if i == nil || i.handle == nil {
		return IONEXSlantDelayEvaluation{}, ErrClosed
	}
	v, e := i.handle.SlantDelayWithPolicy(lat, lon, azimuth, elevation, epochJ2000S, frequencyHz, uint32(policy))
	return publicIONEXEvaluation(v), publicError(e)
}

// SlantDelayAtInstant evaluates a delay without reducing the supplied epoch to whole seconds.
func (i *IONEX) SlantDelayAtInstant(lat, lon, azimuth, elevation float64, epoch ClockEpoch, frequencyHz float64) (float64, *IONEXEpochError, error) {
	if i == nil || i.handle == nil {
		return 0, nil, ErrClosed
	}
	delay, detail, err := i.handle.SlantDelayAtInstant(lat, lon, azimuth, elevation, nativeClockEpoch(epoch), frequencyHz)
	return delay, publicIONEXEpochError(detail), publicError(err)
}

// SlantDelayAtInstantWithPolicy applies the exact composite policy and returns complete typed detail.
func (i *IONEX) SlantDelayAtInstantWithPolicy(lat, lon, azimuth, elevation float64, epoch ClockEpoch, frequencyHz float64, policy IONEXSlantPolicy) (IONEXInstantSlantResult, error) {
	if i == nil || i.handle == nil {
		return IONEXInstantSlantResult{}, ErrClosed
	}
	row, err := i.handle.SlantDelayAtInstantWithPolicy(lat, lon, azimuth, elevation, nativeClockEpoch(epoch), frequencyHz, nativeIONEXPolicy(policy))
	return publicIONEXInstantRow(row), publicError(err)
}

// SlantDelayResultsAtInstants fills caller-owned Go rows from exact-time requests.
func (i *IONEX) SlantDelayResultsAtInstants(requests []IONEXInstantSlantRequest, out []IONEXInstantSlantResult, policy IONEXSlantPolicy) error {
	if i == nil || i.handle == nil {
		return ErrClosed
	}
	if len(out) != len(requests) {
		return invalidArgument("IONEX request and output lengths differ")
	}
	nativeRequests := nativeIONEXRequests(requests)
	rows, err := i.handle.SlantDelayResultsAtInstants(nativeRequests, nativeIONEXPolicy(policy))
	if err != nil {
		return publicError(err)
	}
	if len(rows) != len(out) {
		return errors.New("sidereon: native IONEX result count differs from request count")
	}
	for index := range out {
		out[index] = publicIONEXInstantRow(rows[index])
	}
	return nil
}

// SlantDelayResultsAtInstantsOwned returns independent rows with their messages and mapping codes.
func (i *IONEX) SlantDelayResultsAtInstantsOwned(requests []IONEXInstantSlantRequest, policy IONEXSlantPolicy) (IONEXInstantSlantResultList, error) {
	if i == nil || i.handle == nil {
		return IONEXInstantSlantResultList{}, ErrClosed
	}
	values, err := i.handle.SlantDelayResultsAtInstantsOwned(nativeIONEXRequests(requests), nativeIONEXPolicy(policy))
	if err != nil {
		return IONEXInstantSlantResultList{}, publicError(err)
	}
	out := IONEXInstantSlantResultList{Rows: make([]IONEXInstantSlantResult, len(values.Rows))}
	for index := range values.Rows {
		out.Rows[index] = publicIONEXInstantRow(values.Rows[index])
	}
	return out, nil
}

// DefaultIONEXSlantPolicy returns strict coverage/nodes and single-layer mapping.
func DefaultIONEXSlantPolicy() IONEXSlantPolicy {
	return IONEXSlantPolicy{Coverage: IONEXCoveragePolicyStrict, MissingNodes: IONEXMissingNodesStrict, Mapping: IONEXMappingSingleLayer}
}

// IONEXSlantPolicyFromCoverage returns a default policy with the requested
// coverage behavior and strict missing-node/single-layer defaults.
func IONEXSlantPolicyFromCoverage(coverage IONEXCoveragePolicy) IONEXSlantPolicy {
	policy := DefaultIONEXSlantPolicy()
	policy.Coverage = coverage
	return policy
}

// NewIONEXSlantPolicy constructs a composite policy from its three numeric
// policy tags. Unknown values remain intact and are handled by native calls.
func NewIONEXSlantPolicy(coverage IONEXCoveragePolicy, missingNodes IONEXMissingNodePolicy, mapping IONEXMappingPolicy) IONEXSlantPolicy {
	return IONEXSlantPolicy{Coverage: coverage, MissingNodes: missingNodes, Mapping: mapping}
}

func nativeIONEXPolicy(value IONEXSlantPolicy) native.IonexSlantPolicy {
	return native.IonexSlantPolicy{Coverage: uint32(value.Coverage), MissingNodes: uint32(value.MissingNodes), Mapping: uint32(value.Mapping)}
}
func nativeIONEXRequests(values []IONEXInstantSlantRequest) []native.IonexInstantSlantRequest {
	out := make([]native.IonexInstantSlantRequest, len(values))
	for i, v := range values {
		out[i] = native.IonexInstantSlantRequest{LatDeg: v.LatDeg, LonDeg: v.LonDeg, AzimuthDeg: v.AzimuthDeg, ElevationDeg: v.ElevationDeg, Epoch: nativeClockEpoch(v.Epoch), FrequencyHz: v.FrequencyHz}
	}
	return out
}
func publicIONEXEvaluation(v native.IonexSlantDelayEvaluation) IONEXSlantDelayEvaluation {
	return IONEXSlantDelayEvaluation{DelayM: v.DelayM, Status: IONEXSlantDelayStatus(v.Status), CoverageError: IONEXCoverageErrorKind(v.CoverageError), IsValid: v.IsValid, HasHeld: v.HasHeld, HasDegraded: v.HasDegraded, HasAssumedMapping: v.HasAssumedMapping, Gap: IONEXNodeGap{HasGap: v.Gap.HasGap, Earlier: v.Gap.Earlier, Later: v.Gap.Later}, AssumedMapping: IONEXAssumedMappingKind(v.AssumedMapping)}
}
func publicIONEXEpochError(v native.IonexEpochError) *IONEXEpochError {
	if v.Kind == 0 {
		return nil
	}
	return &IONEXEpochError{Kind: IONEXEpochErrorKind(v.Kind), Scale: TimeScale(v.Scale), HasUTCJ2000S: v.HasUTCJ2000S, UTCJ2000S: v.UTCJ2000S}
}
func publicIONEXSlantError(v native.IonexSlantError) *IONEXSlantError {
	if v.Kind == 0 {
		return nil
	}
	return &IONEXSlantError{Kind: IONEXSlantErrorKind(v.Kind), CoverageError: IONEXCoverageErrorKind(v.CoverageError), HasGap: v.HasGap, Gap: IONEXNodeGap{HasGap: v.Gap.HasGap, Earlier: v.Gap.Earlier, Later: v.Gap.Later}, Refusal: v.Refusal, RefusalMapNumber: v.RefusalMapNumber, RefusalLatIndex: v.RefusalLatIndex, RefusalLonIndex: v.RefusalLonIndex, HasMappingDeclaration: v.HasMappingDeclaration, MappingDeclaration: v.MappingDeclaration, HasMappingFunction: v.HasMappingFunction, MappingFunction: v.MappingFunction}
}
func publicIONEXInstantRow(v native.IonexInstantSlantRow) IONEXInstantSlantResult {
	return IONEXInstantSlantResult{IsOK: v.IsOK, Status: StatusCode(v.Status), Evaluation: publicIONEXEvaluation(v.Evaluation), Error: publicIONEXSlantError(v.Error), EpochError: publicIONEXEpochError(v.EpochError), Message: v.Message, MappingCode: v.MappingCode}
}

// TECSamples returns detached TEC samples from the product.
func (i *IONEX) TECSamples() ([]TECSample, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.TECSamples()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]TECSample, len(v))
	for j := range v {
		out[j] = publicTECSample(v[j])
	}
	return out, nil
}

// GridInfo returns native grid dimensions and metadata.
func (i *IONEX) GridInfo() (TECGridSamplesInfo, error) {
	if i == nil || i.handle == nil {
		return TECGridSamplesInfo{}, ErrClosed
	}
	v, e := i.handle.GridInfo()
	return TECGridSamplesInfo{MapEpochCount: v.MapEpochCount, LatNodeCount: v.LatNodeCount, LonNodeCount: v.LonNodeCount, DLatDeg: v.DLatDeg, DLonDeg: v.DLonDeg, ShellHeightKm: v.ShellHeightKm, BaseRadiusKm: v.BaseRadiusKm, Exponent: v.Exponent, RMSPresent: v.RMSPresent, TECMAPValueCount: v.TECMAPValueCount, RMSMAPValueCount: v.RMSMAPValueCount}, publicError(e)
}

// TECMAPsTECU returns flattened VTEC map values.
func (i *IONEX) TECMAPsTECU() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.TECMAPsTECU()
	return v, publicError(e)
}

// RMSMAPsTECU returns flattened RMS map values.
func (i *IONEX) RMSMAPsTECU() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.RMSMAPsTECU()
	return v, publicError(e)
}

// GridEpochsJ2000S returns grid map epochs as J2000 seconds.
func (i *IONEX) GridEpochsJ2000S() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.GridEpochsJ2000S()
	return v, publicError(e)
}

// TECMapPresence reports per-cell VTEC presence in flattened map order.
func (i *IONEX) TECMapPresence() ([]bool, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.TECMapPresence()
	return v, publicError(e)
}

// RMSMapPresence reports per-cell RMS presence in flattened map order.
func (i *IONEX) RMSMapPresence() ([]bool, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.RMSMapPresence()
	return v, publicError(e)
}

// HeightMapPresence reports shell-height presence for each flattened cell.
func (i *IONEX) HeightMapPresence() ([]bool, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.HeightMapPresence()
	return v, publicError(e)
}

// HeightMapsKm returns the flattened per-cell shell-height maps.
func (i *IONEX) HeightMapsKm() ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	v, e := i.handle.HeightMapsKm()
	return v, publicError(e)
}

// IONEXCoveragePolicy controls behavior outside map coverage.
type IONEXCoveragePolicy uint32

const (
	// IONEXCoveragePolicyStrict rejects epochs outside map coverage.
	IONEXCoveragePolicyStrict IONEXCoveragePolicy = IONEXCoveragePolicy(native.IONEXCoveragePolicyStrictValue)
	// IONEXCoveragePolicyHold holds the nearest map outside coverage.
	IONEXCoveragePolicyHold IONEXCoveragePolicy = IONEXCoveragePolicy(native.IONEXCoveragePolicyHoldValue)
)

// IONEXSlantDelayStatus identifies the returned delay status.
type IONEXSlantDelayStatus uint32

const (
	// IONEXSlantDelayStatusValid indicates a directly interpolated delay.
	IONEXSlantDelayStatusValid IONEXSlantDelayStatus = IONEXSlantDelayStatus(native.IONEXSlantDelayStatusValidValue)
	// IONEXSlantDelayStatusHeld indicates a held boundary delay.
	IONEXSlantDelayStatusHeld IONEXSlantDelayStatus = IONEXSlantDelayStatus(native.IONEXSlantDelayStatusHeldValue)
)

// IONEXCoverageErrorKind identifies a held-value coverage miss.
type IONEXCoverageErrorKind uint32

const (
	// IONEXCoverageErrorNone indicates no coverage error.
	IONEXCoverageErrorNone IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorNoneValue)
	// IONEXCoverageErrorEpochBeforeFirstMap indicates an early epoch.
	IONEXCoverageErrorEpochBeforeFirstMap IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorEpochBeforeFirstMapValue)
	// IONEXCoverageErrorEpochAfterLastMap indicates a late epoch.
	IONEXCoverageErrorEpochAfterLastMap IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorEpochAfterLastMapValue)
	// IONEXCoverageErrorLatitude indicates an invalid latitude.
	IONEXCoverageErrorLatitude IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorLatitudeValue)
	// IONEXCoverageErrorLongitude indicates an invalid longitude.
	IONEXCoverageErrorLongitude IONEXCoverageErrorKind = IONEXCoverageErrorKind(native.IONEXCoverageErrorLongitudeValue)
)
