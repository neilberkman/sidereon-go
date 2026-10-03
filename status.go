package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"sidereon.dev/go/v3/internal/native"
)

func invalidArgument(detail string) error { return fmt.Errorf("sidereon: %s", detail) }

// StatusCode is the numeric status returned by the C ABI.
type StatusCode int

const (
	// StatusOK indicates successful completion.
	StatusOK StatusCode = 0
	// StatusNullPointer indicates a required null pointer.
	StatusNullPointer StatusCode = 1
	// StatusInvalidArgument indicates an invalid argument.
	StatusInvalidArgument StatusCode = 2
	// StatusInvalidToken indicates an invalid textual token.
	StatusInvalidToken StatusCode = 3
	// StatusSP3Parse indicates an SP3 parsing failure.
	StatusSP3Parse StatusCode = 4
	// StatusSolve indicates a numerical solve failure.
	StatusSolve StatusCode = 5
	// StatusPanic indicates a native panic boundary failure.
	StatusPanic StatusCode = 6
	// StatusTimeout indicates a native timeout.
	StatusTimeout StatusCode = 7
)

// SGP4ErrorKind identifies the native SGP4 propagation error category.
type SGP4ErrorKind uint32

const (
	// SGP4ErrorNone indicates no SGP4 failure was captured.
	SGP4ErrorNone SGP4ErrorKind = iota
	// SGP4ErrorInvalidInput indicates invalid propagation input.
	SGP4ErrorInvalidInput
	// SGP4ErrorNonFiniteOutput indicates the propagator produced non-finite state.
	SGP4ErrorNonFiniteOutput
	// SGP4ErrorInvalidTLE indicates the element set is invalid for propagation.
	SGP4ErrorInvalidTLE
	// SGP4ErrorEngine indicates an underlying SGP4 engine error.
	SGP4ErrorEngine
	// SGP4ErrorResonanceStepBudget indicates a resonance propagation budget failure.
	SGP4ErrorResonanceStepBudget
)

// SGP4Error retains the typed SGP4 cause captured on the same native call.
type SGP4Error struct {
	// Kind is the native SGP4 failure category, including unknown future values.
	Kind SGP4ErrorKind
	// HasCode reports whether Code contains an engine-specific SGP4 code.
	HasCode bool
	// Code is the engine-specific SGP4 code when HasCode is true.
	Code int32
	// HasBudget reports whether Budget contains a resonance-step budget.
	HasBudget bool
	// Budget is the configured step budget when HasBudget is true.
	Budget uint64
}

// StatusError reports one failed C call. Text is the stable status name from
// sidereon_status_message; Detail and any structured cause are captured from
// the same native operation.
type StatusError struct {
	// Code is the native status namespace.
	Code StatusCode
	// Text is the stable native status label.
	Text string
	// Detail is same-call thread-local diagnostic text.
	Detail string
	// Engine carries the typed schema1 generic engine error summary and lossless payload.
	Engine *EngineError
	// SP3 carries the typed validation summary and lossless native JSON detail.
	SP3 *SP3Error
	// TerrainDatum and TerrainStore carry optional structured causes.
	TerrainDatum  *TerrainDatumError
	TerrainStore  *TerrainStoreError
	TerrainLookup *TerrainLookupError
	Bias          *BiasError
	// RTCM carries the native RTCM/SBAS discriminant and lossless JSON payload.
	RTCM *RTCMError
	// ANTEX carries a typed antenna lookup or calibration refusal when available.
	ANTEX *ANTEXError
	// DtedTile carries the complete typed refusal from loading or querying one DTED tile.
	DtedTile *DtedTileError
	// Geoid carries the complete typed refusal from a geoid constructor.
	Geoid *GeoidError
	// PreciseArtifact carries the complete typed artifact producer refusal.
	PreciseArtifact *PreciseArtifactError
	// SGP4 carries typed propagation details captured on the same C call.
	SGP4 *SGP4Error
}

// QualityErrorKind identifies the core quality refusal captured from one
// quality-producing native call. Unknown numeric values are retained.
type QualityErrorKind uint32

const (
	// QualityErrorNone means quality evaluation found no error.
	QualityErrorNone QualityErrorKind = iota
	// QualityErrorInvalidElevation identifies an elevation outside the supported range.
	QualityErrorInvalidElevation
	// QualityErrorMissingCN0 identifies a missing carrier-to-noise observation.
	QualityErrorMissingCN0
	// QualityErrorInvalidParameter identifies a non-finite or out-of-range quality parameter.
	QualityErrorInvalidParameter
	// QualityErrorInvalidProbability identifies a probability outside (0, 1).
	QualityErrorInvalidProbability
	// QualityErrorInvalidSystemCount identifies an invalid constellation count.
	QualityErrorInvalidSystemCount
	// QualityErrorInvalidDOF identifies invalid degrees of freedom.
	QualityErrorInvalidDOF
	// QualityErrorInvalidWeight identifies an invalid observation weight.
	QualityErrorInvalidWeight
	// QualityErrorInvalidReliabilityParameter identifies an invalid reliability-monitoring parameter.
	QualityErrorInvalidReliabilityParameter
	// QualityErrorInvalidResiduals identifies invalid or incomplete residual data.
	QualityErrorInvalidResiduals
	// QualityErrorInvalidDesign identifies a design matrix with invalid dimensions or entries.
	QualityErrorInvalidDesign
	// QualityErrorSingularGeometry identifies geometry that cannot support the requested quality statistic.
	QualityErrorSingularGeometry
	// QualityErrorMissingVariances identifies missing observation variances.
	QualityErrorMissingVariances
	// QualityErrorInvalidVariance identifies a non-positive or non-finite variance.
	QualityErrorInvalidVariance
)

// QualityErrorUnknown represents an unrecognized native value.
const QualityErrorUnknown QualityErrorKind = 999

// QualityError is a typed core quality refusal. It unwraps to the status and
// diagnostic captured from the same native operation.
type QualityError struct {
	Kind    QualityErrorKind
	Message string
	cause   error
}

// Error returns the QualityError message.
func (e *QualityError) Error() string {
	if e == nil {
		return "sidereon: quality refusal"
	}
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("sidereon: quality refusal (kind %d)", e.Kind)
}

// Unwrap returns the underlying cause when one is retained.
func (e *QualityError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// FDEUnresolvedReason identifies why FDE stopped with a fault still detected.
type FDEUnresolvedReason uint32

const (
	// FDEUnresolvedNone means the FDE search resolved the detected fault.
	FDEUnresolvedNone FDEUnresolvedReason = iota
	// FDEUnresolvedExclusionBudgetExhausted means no accepted solution was found before the exclusion budget ended.
	FDEUnresolvedExclusionBudgetExhausted
	// FDEUnresolvedNoAdmissibleExclusion means every candidate exclusion violated the configured acceptance rules.
	FDEUnresolvedNoAdmissibleExclusion
)

// FDEUnresolvedUnknown preserves an unrecognized native unresolved-fault value.
const FDEUnresolvedUnknown FDEUnresolvedReason = 999

// FDEUnresolvedError retains the last faulted solution and its actual RAIM
// decision when the exclusion loop cannot produce an accepted fix.
type FDEUnresolvedError struct {
	Reason                 FDEUnresolvedReason
	Solution               SPPSolution
	HasSolution            bool
	Excluded               []string
	HasExcluded            bool
	Iterations             int
	RAIM                   RAIMResult
	HasRAIM                bool
	NormalizedResiduals    []RAIMNormalizedResidual
	HasNormalizedResiduals bool
	CaptureError           error
	cause                  error
}

// Error returns the FDEUnresolvedError message.
func (e *FDEUnresolvedError) Error() string {
	if e == nil {
		return "sidereon: FDE left a detected fault unresolved"
	}
	name := "unknown reason"
	switch e.Reason {
	case FDEUnresolvedExclusionBudgetExhausted:
		name = "exclusion budget exhausted"
	case FDEUnresolvedNoAdmissibleExclusion:
		name = "no admissible exclusion"
	}
	message := "sidereon: FDE fault unresolved: " + name
	if e.CaptureError != nil {
		message += "; some diagnostic data could not be captured: " + e.CaptureError.Error()
	}
	return message
}

// Unwrap returns the underlying cause when one is retained.
func (e *FDEUnresolvedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// SP3ErrorKind identifies the category of a typed SP3 validation failure.
type SP3ErrorKind uint32

const (
	// SP3ErrorKindNone means validation succeeded.
	SP3ErrorKindNone SP3ErrorKind = 0
	// SP3ErrorKindExactValidation classifies an exact-sample validation failure.
	SP3ErrorKindExactValidation SP3ErrorKind = 1
	// SP3ErrorKindEpochInterval classifies an invalid or mismatched epoch interval.
	SP3ErrorKindEpochInterval SP3ErrorKind = 2
	// SP3ErrorKindMergeTolerance classifies a merge-tolerance validation failure.
	SP3ErrorKindMergeTolerance SP3ErrorKind = 3
	// SP3ErrorKindContinuityOptions classifies invalid orbit or clock continuity options.
	SP3ErrorKindContinuityOptions SP3ErrorKind = 4
)

// SP3ErrorField identifies the input or record rejected by SP3 validation.
type SP3ErrorField uint32

const (
	// SP3ErrorFieldNone marks the absence of a field-specific SP3 validation failure.
	SP3ErrorFieldNone SP3ErrorField = 0
	// SP3ErrorFieldDeclaredStart identifies the product-declared start epoch when SP3 validation fails.
	SP3ErrorFieldDeclaredStart SP3ErrorField = 1
	// SP3ErrorFieldTargetEpochInterval identifies the expected target epoch interval when SP3 validation fails.
	SP3ErrorFieldTargetEpochInterval SP3ErrorField = 2
	// SP3ErrorFieldMergedEpochInterval identifies the merged input epoch interval when SP3 validation fails.
	SP3ErrorFieldMergedEpochInterval SP3ErrorField = 3
	// SP3ErrorFieldPositionTolerance identifies the position continuity tolerance in metres when SP3 validation fails.
	SP3ErrorFieldPositionTolerance SP3ErrorField = 4
	// SP3ErrorFieldClockTolerance identifies the clock continuity tolerance in seconds when SP3 validation fails.
	SP3ErrorFieldClockTolerance SP3ErrorField = 5
	// SP3ErrorFieldOutlierPositionTolerance identifies the outlier position tolerance in metres when SP3 validation fails.
	SP3ErrorFieldOutlierPositionTolerance SP3ErrorField = 6
	// SP3ErrorFieldOutlierClockTolerance identifies the outlier clock tolerance in seconds when SP3 validation fails.
	SP3ErrorFieldOutlierClockTolerance SP3ErrorField = 7
	// SP3ErrorFieldSpeedBound identifies the physical speed bound when SP3 validation fails.
	SP3ErrorFieldSpeedBound SP3ErrorField = 8
	// SP3ErrorFieldResidualToleranceM identifies the residual tolerance in metres when SP3 validation fails.
	SP3ErrorFieldResidualToleranceM SP3ErrorField = 9
	// SP3ErrorFieldSP3Content identifies malformed or inconsistent SP3 content when SP3 validation fails.
	SP3ErrorFieldSP3Content SP3ErrorField = 10
	// SP3ErrorFieldCatalogIdentity identifies the catalog identity when SP3 validation fails.
	SP3ErrorFieldCatalogIdentity SP3ErrorField = 11
	// SP3ErrorFieldProductFamily identifies the product family when SP3 validation fails.
	SP3ErrorFieldProductFamily SP3ErrorField = 12
	// SP3ErrorFieldIssueToken identifies the issue-date token when SP3 validation fails.
	SP3ErrorFieldIssueToken SP3ErrorField = 13
	// SP3ErrorFieldSpanToken identifies the span token when SP3 validation fails.
	SP3ErrorFieldSpanToken SP3ErrorField = 14
	// SP3ErrorFieldSampleToken identifies the sample token when SP3 validation fails.
	SP3ErrorFieldSampleToken SP3ErrorField = 15
	// SP3ErrorFieldExpectedAgency identifies the expected producing agency when SP3 validation fails.
	SP3ErrorFieldExpectedAgency SP3ErrorField = 16
	// SP3ErrorFieldProducingAgency identifies the agency declared by the product when SP3 validation fails.
	SP3ErrorFieldProducingAgency SP3ErrorField = 17
	// SP3ErrorFieldTerminalRecord identifies the terminal record when SP3 validation fails.
	SP3ErrorFieldTerminalRecord SP3ErrorField = 18
	// SP3ErrorFieldHeaderRecordCount identifies the header record count when SP3 validation fails.
	SP3ErrorFieldHeaderRecordCount SP3ErrorField = 19
	// SP3ErrorFieldSatelliteCount identifies the satellite count when SP3 validation fails.
	SP3ErrorFieldSatelliteCount SP3ErrorField = 20
	// SP3ErrorFieldSatelliteDeclarations identifies the satellite declarations when SP3 validation fails.
	SP3ErrorFieldSatelliteDeclarations SP3ErrorField = 21
	// SP3ErrorFieldSatelliteRecordSequence identifies the satellite-record sequence when SP3 validation fails.
	SP3ErrorFieldSatelliteRecordSequence SP3ErrorField = 22
	// SP3ErrorFieldBodyRecordOrder identifies the body-record order when SP3 validation fails.
	SP3ErrorFieldBodyRecordOrder SP3ErrorField = 23
	// SP3ErrorFieldHeaderCadence identifies the header cadence when SP3 validation fails.
	SP3ErrorFieldHeaderCadence SP3ErrorField = 24
	// SP3ErrorFieldDeclaredEpochCount identifies the declared epoch count when SP3 validation fails.
	SP3ErrorFieldDeclaredEpochCount SP3ErrorField = 25
	// SP3ErrorFieldGPSStart identifies the GPS start epoch when SP3 validation fails.
	SP3ErrorFieldGPSStart SP3ErrorField = 26
	// SP3ErrorFieldHeaderStartMetadata identifies the header start metadata when SP3 validation fails.
	SP3ErrorFieldHeaderStartMetadata SP3ErrorField = 27
	// SP3ErrorFieldEpochGrid identifies the expected epoch grid when SP3 validation fails.
	SP3ErrorFieldEpochGrid SP3ErrorField = 28
	// SP3ErrorFieldRequestedSpan identifies the requested span when SP3 validation fails.
	SP3ErrorFieldRequestedSpan SP3ErrorField = 29
	// SP3ErrorFieldFormatVersion identifies the format version when SP3 validation fails.
	SP3ErrorFieldFormatVersion SP3ErrorField = 30
	// SP3ErrorFieldOther identifies another SP3 field when SP3 validation fails.
	SP3ErrorFieldOther SP3ErrorField = 255
)

// SP3ErrorReason identifies why a typed SP3 validation field was rejected.
type SP3ErrorReason uint32

const (
	// SP3ErrorReasonNone records that means no validation reason was recorded.
	SP3ErrorReasonNone SP3ErrorReason = 0
	// SP3ErrorReasonMismatch records that means actual data differs from the required value.
	SP3ErrorReasonMismatch SP3ErrorReason = 1
	// SP3ErrorReasonNotFinite records that means a value is NaN or infinite.
	SP3ErrorReasonNotFinite SP3ErrorReason = 2
	// SP3ErrorReasonNegative records that means a value is below zero.
	SP3ErrorReasonNegative SP3ErrorReason = 3
	// SP3ErrorReasonNotPositive records that means a value is zero or below.
	SP3ErrorReasonNotPositive SP3ErrorReason = 4
	// SP3ErrorReasonNotWholeTicks records that means a duration is not an integer number of clock ticks.
	SP3ErrorReasonNotWholeTicks SP3ErrorReason = 5
	// SP3ErrorReasonBeyondTickResolution records that means a duration cannot be represented at the clock tick resolution.
	SP3ErrorReasonBeyondTickResolution SP3ErrorReason = 6
	// SP3ErrorReasonOutsideSpecificationRange records that means a value falls outside the format-specified range.
	SP3ErrorReasonOutsideSpecificationRange SP3ErrorReason = 7
	// SP3ErrorReasonInvalid records that means a value fails semantic validation.
	SP3ErrorReasonInvalid SP3ErrorReason = 8
	// SP3ErrorReasonMissing records that means a required value is absent.
	SP3ErrorReasonMissing SP3ErrorReason = 9
	// SP3ErrorReasonUnsupported records that means the value is valid but unsupported by this operation.
	SP3ErrorReasonUnsupported SP3ErrorReason = 10
	// SP3ErrorReasonMalformed records that means the input token is malformed.
	SP3ErrorReasonMalformed SP3ErrorReason = 11
	// SP3ErrorReasonNonCanonical records that means the token is valid but not in canonical form.
	SP3ErrorReasonNonCanonical SP3ErrorReason = 12
	// SP3ErrorReasonDuplicate records that means a field or record appears more than once.
	SP3ErrorReasonDuplicate SP3ErrorReason = 13
	// SP3ErrorReasonTrailingContent records that means unexpected bytes follow the expected terminal content.
	SP3ErrorReasonTrailingContent SP3ErrorReason = 14
	// SP3ErrorReasonEmpty records that means required content is empty.
	SP3ErrorReasonEmpty SP3ErrorReason = 15
	// SP3ErrorReasonNotMultiple records that means a count or duration is not an allowed multiple.
	SP3ErrorReasonNotMultiple SP3ErrorReason = 16
	// SP3ErrorReasonBeforeGPSEpoch records that means the epoch precedes the GPS epoch.
	SP3ErrorReasonBeforeGPSEpoch SP3ErrorReason = 17
	// SP3ErrorReasonOther records that preserves another validation reason.
	SP3ErrorReasonOther SP3ErrorReason = 255
)

// SP3Error retains the typed native summary and raw schema-versioned JSON.
// The payload keeps exact i128 tick counts as decimal strings.
type SP3Error struct {
	Kind               SP3ErrorKind
	Field              SP3ErrorField
	Reason             SP3ErrorReason
	HasValue           bool
	Value              float64
	HasRequestedTick   bool
	HasDeclaredTick    bool
	HasRequestedJ2000S bool
	HasDeclaredJ2000S  bool
	RequestedJ2000S    float64
	DeclaredJ2000S     float64
	Payload            json.RawMessage
}

// Error returns the SP3Error message.
func (e *SP3Error) Error() string {
	if e == nil {
		return "sidereon: SP3 validation error"
	}
	return fmt.Sprintf("sidereon: SP3 validation kind=%d field=%d reason=%d", e.Kind, e.Field, e.Reason)
}

// RTCMErrorClass identifies the native RTCM or SBAS error category.
type RTCMErrorClass uint32

const (
	// RTCMErrorClassNone means no error or special condition occurred.
	RTCMErrorClassNone RTCMErrorClass = 0
	// RTCMErrorClassEncode classifies a failure while encoding an RTCM message.
	RTCMErrorClassEncode RTCMErrorClass = 1
	// RTCMErrorClassConversion classifies a field-conversion failure in RTCM processing.
	RTCMErrorClassConversion RTCMErrorClass = 2
	// RTCMErrorClassOther preserves a native condition outside the named cases.
	RTCMErrorClassOther RTCMErrorClass = 3
	// RTCMErrorClassSBASEncode classifies a failure while encoding an SBAS message.
	RTCMErrorClassSBASEncode RTCMErrorClass = 4
)

// RTCMErrorKindUnknown marks an unclassified typed error with no variant code.
const RTCMErrorKindUnknown uint32 = 0

// RTCMError preserves the native error discriminant and its schema-versioned
// JSON payload. Integer values in Payload are decimal strings as defined by
// the C ABI, preserving values beyond binary64 precision.
type RTCMError struct {
	Class   RTCMErrorClass
	Kind    uint32
	Payload json.RawMessage
}

// Error returns the RTCMError message.
func (e *RTCMError) Error() string {
	if e == nil {
		return "sidereon: RTCM error"
	}
	return fmt.Sprintf("sidereon: RTCM error class %d kind %d", e.Class, e.Kind)
}

// EngineErrorFamily identifies the core subsystem producing a schema1 generic engine error.
type EngineErrorFamily uint32

const (
	// EngineErrorFamilyNone marks the absence of a family-specific engine error.
	EngineErrorFamilyNone EngineErrorFamily = 0
	// EngineErrorFamilyRtk identifies typed engine errors from RTK positioning.
	EngineErrorFamilyRtk EngineErrorFamily = 1
	// EngineErrorFamilyStaticReference identifies typed engine errors from static reference positioning.
	EngineErrorFamilyStaticReference EngineErrorFamily = 2
	// EngineErrorFamilyTrls identifies typed engine errors from TRLS.
	EngineErrorFamilyTrls EngineErrorFamily = 3
	// EngineErrorFamilyIls identifies typed engine errors from ILS.
	EngineErrorFamilyIls EngineErrorFamily = 4
	// EngineErrorFamilySpk identifies typed engine errors from SPK.
	EngineErrorFamilySpk EngineErrorFamily = 5
	// EngineErrorFamilyCdm identifies typed engine errors from CDM.
	EngineErrorFamilyCdm EngineErrorFamily = 6
	// EngineErrorFamilyTdm identifies typed engine errors from TDM.
	EngineErrorFamilyTdm EngineErrorFamily = 7
	// EngineErrorFamilyFusion identifies typed engine errors from sensor fusion.
	EngineErrorFamilyFusion EngineErrorFamily = 8
	// EngineErrorFamilyFusionStateCodec identifies typed engine errors from fusion-state serialization.
	EngineErrorFamilyFusionStateCodec EngineErrorFamily = 9
	// EngineErrorFamilyAllan identifies typed engine errors from Allan deviation.
	EngineErrorFamilyAllan EngineErrorFamily = 10
	// EngineErrorFamilyPowerLawNoise identifies typed engine errors from power-law noise.
	EngineErrorFamilyPowerLawNoise EngineErrorFamily = 11
	// EngineErrorFamilyFrameCatalog identifies typed engine errors from frame catalog.
	EngineErrorFamilyFrameCatalog EngineErrorFamily = 12
	// EngineErrorFamilySidereal identifies typed engine errors from sidereal filtering.
	EngineErrorFamilySidereal EngineErrorFamily = 13
	// EngineErrorFamilyAtmosphere identifies typed engine errors from atmosphere models.
	EngineErrorFamilyAtmosphere EngineErrorFamily = 14
	// EngineErrorFamilySourceLocalization identifies typed engine errors from source localization.
	EngineErrorFamilySourceLocalization EngineErrorFamily = 15
	// EngineErrorFamilyGeodeticTimeSeries identifies typed engine errors from geodetic time series.
	EngineErrorFamilyGeodeticTimeSeries EngineErrorFamily = 16
	// EngineErrorFamilyNormality identifies typed engine errors from normality testing.
	EngineErrorFamilyNormality EngineErrorFamily = 17
	// EngineErrorFamilyTrack identifies typed engine errors from track propagation.
	EngineErrorFamilyTrack EngineErrorFamily = 18
	// EngineErrorFamilyPreciseSamples identifies typed engine errors from precise-sample validation.
	EngineErrorFamilyPreciseSamples EngineErrorFamily = 19
	// EngineErrorFamilyPreciseInterpolant identifies typed engine errors from precise ephemeris interpolation.
	EngineErrorFamilyPreciseInterpolant EngineErrorFamily = 20
	// EngineErrorFamilySpaceWeather identifies typed engine errors from space weather.
	EngineErrorFamilySpaceWeather EngineErrorFamily = 21
	// EngineErrorFamilyAraim identifies typed engine errors from ARAIM.
	EngineErrorFamilyAraim EngineErrorFamily = 22
	// EngineErrorFamilyReducedOrbit identifies typed engine errors from reduced-orbit models.
	EngineErrorFamilyReducedOrbit EngineErrorFamily = 23
	// EngineErrorFamilyReducedOrbitSource identifies typed engine errors from reduced-orbit sources.
	EngineErrorFamilyReducedOrbitSource EngineErrorFamily = 24
	// EngineErrorFamilyPiecewiseOrbit identifies typed engine errors from piecewise orbits.
	EngineErrorFamilyPiecewiseOrbit EngineErrorFamily = 25
	// EngineErrorFamilyOrbitFit identifies typed engine errors from orbit fitting.
	EngineErrorFamilyOrbitFit EngineErrorFamily = 26
	// EngineErrorFamilyElements identifies typed engine errors from orbital elements.
	EngineErrorFamilyElements EngineErrorFamily = 27
	// EngineErrorFamilyEquinoctial identifies typed engine errors from equinoctial elements.
	EngineErrorFamilyEquinoctial EngineErrorFamily = 28
	// EngineErrorFamilyRtnFrame identifies typed engine errors from RTN frames.
	EngineErrorFamilyRtnFrame EngineErrorFamily = 29
	// EngineErrorFamilyAnomaly identifies typed engine errors from orbital anomaly conversion.
	EngineErrorFamilyAnomaly EngineErrorFamily = 30
	// EngineErrorFamilyPropagation identifies typed engine errors from orbit propagation.
	EngineErrorFamilyPropagation EngineErrorFamily = 31
	// EngineErrorFamilyDecay identifies typed engine errors from orbital decay.
	EngineErrorFamilyDecay EngineErrorFamily = 32
	// EngineErrorFamilyDgnss identifies typed engine errors from differential GNSS.
	EngineErrorFamilyDgnss EngineErrorFamily = 33
	// EngineErrorFamilyScenario identifies typed engine errors from scenario evaluation.
	EngineErrorFamilyScenario EngineErrorFamily = 34
	// EngineErrorFamilyCatalog identifies typed engine errors from catalog access.
	EngineErrorFamilyCatalog EngineErrorFamily = 35
	// EngineErrorFamilyExactCache identifies typed engine errors from exact-epoch caching.
	EngineErrorFamilyExactCache EngineErrorFamily = 36
	// EngineErrorFamilyTca identifies typed engine errors from time of closest approach.
	EngineErrorFamilyTca EngineErrorFamily = 37
	// EngineErrorFamilyAlmanac identifies typed engine errors from almanac data.
	EngineErrorFamilyAlmanac EngineErrorFamily = 38
	// EngineErrorFamilyObserve identifies typed engine errors from observation geometry.
	EngineErrorFamilyObserve EngineErrorFamily = 39
	// EngineErrorFamilyBodyObservation identifies typed engine errors from body observations.
	EngineErrorFamilyBodyObservation EngineErrorFamily = 40
	// EngineErrorFamilyLookAngle identifies typed engine errors from look-angle calculation.
	EngineErrorFamilyLookAngle EngineErrorFamily = 41
	// EngineErrorFamilyPass identifies typed engine errors from pass prediction.
	EngineErrorFamilyPass EngineErrorFamily = 42
	// EngineErrorFamilyEventFinder identifies typed engine errors from event finding.
	EngineErrorFamilyEventFinder EngineErrorFamily = 43
	// EngineErrorFamilyFrameTransform identifies typed engine errors from frame transforms.
	EngineErrorFamilyFrameTransform EngineErrorFamily = 44
	// EngineErrorFamilyConjunction identifies typed engine errors from conjunction analysis.
	EngineErrorFamilyConjunction EngineErrorFamily = 45
	// EngineErrorFamilyFacade identifies typed engine errors from high-level facade calls.
	EngineErrorFamilyFacade EngineErrorFamily = 46
	// EngineErrorFamilySpp identifies typed engine errors from single-point positioning.
	EngineErrorFamilySpp EngineErrorFamily = 47
	// EngineErrorFamilySppPolicy identifies typed engine errors from SPP policies.
	EngineErrorFamilySppPolicy EngineErrorFamily = 48
	// EngineErrorFamilySunMoon identifies typed engine errors from sun and moon ephemerides.
	EngineErrorFamilySunMoon EngineErrorFamily = 49
	// EngineErrorFamilyRinexSpp identifies typed engine errors from RINEX SPP.
	EngineErrorFamilyRinexSpp EngineErrorFamily = 50
	// EngineErrorFamilySolutionValidation identifies typed engine errors from solution validation.
	EngineErrorFamilySolutionValidation EngineErrorFamily = 51
	// EngineErrorFamilyRF identifies typed engine errors from radio-frequency analysis.
	EngineErrorFamilyRF EngineErrorFamily = 52
	// EngineErrorFamilyIonosphereFree identifies typed engine errors from ionosphere-free combinations.
	EngineErrorFamilyIonosphereFree EngineErrorFamily = 53
	// EngineErrorFamilyDoppler identifies typed engine errors from Doppler analysis.
	EngineErrorFamilyDoppler EngineErrorFamily = 54
	// EngineErrorFamilyOEM identifies typed engine errors from OEM codecs.
	EngineErrorFamilyOEM EngineErrorFamily = 55
	// EngineErrorFamilyOPM identifies typed engine errors from OPM codecs.
	EngineErrorFamilyOPM EngineErrorFamily = 56
	// EngineErrorFamilyOMM identifies typed engine errors from OMM codecs.
	EngineErrorFamilyOMM EngineErrorFamily = 57
	// EngineErrorFamilyDOP identifies typed engine errors from DOP calculations.
	EngineErrorFamilyDOP EngineErrorFamily = 58
	// EngineErrorFamilyGeofence identifies typed engine errors from geofencing.
	EngineErrorFamilyGeofence EngineErrorFamily = 59
	// EngineErrorFamilySignal identifies typed engine errors from signal analysis.
	EngineErrorFamilySignal EngineErrorFamily = 60
	// EngineErrorFamilyCarrierPhase identifies typed engine errors from carrier-phase processing.
	EngineErrorFamilyCarrierPhase EngineErrorFamily = 61
	// EngineErrorFamilySignalAnalysis identifies typed engine errors from signal analysis.
	EngineErrorFamilySignalAnalysis EngineErrorFamily = 62
	// EngineErrorFamilyErrorMetrics identifies typed engine errors from error metrics.
	EngineErrorFamilyErrorMetrics EngineErrorFamily = 63
	// EngineErrorFamilyObservables identifies typed engine errors from observable processing.
	EngineErrorFamilyObservables EngineErrorFamily = 64
	// EngineErrorFamilyNMEA identifies typed engine errors from NMEA processing.
	EngineErrorFamilyNMEA EngineErrorFamily = 65
	// EngineErrorFamilyTLEFit identifies typed engine errors from TLE fitting.
	EngineErrorFamilyTLEFit EngineErrorFamily = 66
	// EngineErrorFamilyIOD identifies typed engine errors from initial orbit determination.
	EngineErrorFamilyIOD EngineErrorFamily = 67
	// EngineErrorFamilySelection identifies typed engine errors from selection policies.
	EngineErrorFamilySelection EngineErrorFamily = 68
	// EngineErrorFamilyPppAutoInit identifies typed engine errors from PPP auto-initialization.
	EngineErrorFamilyPppAutoInit EngineErrorFamily = 69
	// EngineErrorFamilyStaticPositioning identifies typed engine errors from static positioning.
	EngineErrorFamilyStaticPositioning EngineErrorFamily = 70
	// EngineErrorFamilyTimeOffset identifies typed engine errors from time-offset conversion.
	EngineErrorFamilyTimeOffset EngineErrorFamily = 71
	// EngineErrorFamilyTimeModel identifies typed engine errors from time models.
	EngineErrorFamilyTimeModel EngineErrorFamily = 72
	// EngineErrorFamilyUnknown identifies typed engine errors from an unrecognized engine subsystem.
	EngineErrorFamilyUnknown EngineErrorFamily = 999
)

// Name returns the stable textual name of this engine error family.
func (f EngineErrorFamily) Name() string {
	return native.EngineErrorFamily(f).Name()
}

// String returns the stable textual name of this engine error family.
func (f EngineErrorFamily) String() string {
	return f.Name()
}

// EngineErrorFamilyFromName parses a stable engine-family name and returns the unknown value for unrecognized names.
func EngineErrorFamilyFromName(name string) EngineErrorFamily {
	return EngineErrorFamily(native.EngineErrorFamilyFromName(name))
}

// EngineFloat represents an exact binary64 value serialized in the C schema1 engine error envelope.
type EngineFloat struct {
	Decimal string `json:"decimal"`
	BitsHex string `json:"bits_hex"`
}

// Float64 parses the bit pattern into an exact float64 value (preserving non-finite values and -0.0).
func (ef EngineFloat) Float64() (float64, error) {
	return native.EngineFloat(ef).Float64()
}

// Bits parses the raw uint64 bit pattern.
func (ef EngineFloat) Bits() (uint64, error) {
	return native.EngineFloat(ef).Bits()
}

// ParseEngineFloat extracts an EngineFloat from raw JSON bytes.
func ParseEngineFloat(data []byte) (EngineFloat, error) {
	ef, err := native.ParseEngineFloat(data)
	if err != nil {
		return EngineFloat{}, err
	}
	return EngineFloat(ef), nil
}

// EngineError preserves the native generic schema1 engine error transport record.
// If diagnostic capture or payload decoding encounters an error, partial data
// (such as Family, FamilyName, and raw Payload) is preserved alongside the
// diagnostic refusal in CaptureError according to the partial-record contract.
type EngineError struct {
	// Family is the numeric family identifier from the C ABI.
	Family EngineErrorFamily
	// FamilyName is the string name of the error family.
	FamilyName string
	// Schema is the schema version from the envelope (schema_version=1).
	Schema uint32
	// Operation is the logical operation name reported by the core engine.
	Operation string
	// Kind is the primary error discriminant string from the error node.
	Kind string
	// Fields contains the raw JSON bytes of the error fields.
	// Preserves arbitrary future fields/codes and all integer lexemes without float64 conversion.
	Fields json.RawMessage
	// TypedFields exposes every field recursively while preserving exact JSON
	// value categories and numeric tokens. Fields remains available for byte-level access.
	TypedFields map[string]EngineJSONValue
	// Payload is the complete versioned JSON payload returned by the C ABI.
	Payload json.RawMessage
	// CaptureError reports an explicit error when C info, payload query, or JSON decoding fails.
	CaptureError error
}

// EngineJSONKind identifies one JSON value category in a typed engine error.
type EngineJSONKind = native.EngineJSONKind

const (
	// EngineJSONNull identifies a JSON null value.
	EngineJSONNull = native.EngineJSONNull
	// EngineJSONString identifies a JSON string value.
	EngineJSONString = native.EngineJSONString
	// EngineJSONNumber identifies a JSON number value.
	EngineJSONNumber = native.EngineJSONNumber
	// EngineJSONBoolean identifies a JSON boolean value.
	EngineJSONBoolean = native.EngineJSONBoolean
	// EngineJSONObject identifies a JSON object value.
	EngineJSONObject = native.EngineJSONObject
	// EngineJSONArray identifies a JSON array value.
	EngineJSONArray = native.EngineJSONArray
)

// EngineJSONValue is a recursively typed engine-error field. Number preserves
// the original JSON token using json.Number; Object and Array retain all nested values.
type EngineJSONValue = native.EngineJSONValue

func publicEngineJSONFields(fields map[string]native.EngineJSONValue) map[string]EngineJSONValue {
	if fields == nil {
		return nil
	}
	result := make(map[string]EngineJSONValue, len(fields))
	for key, value := range fields {
		result[key] = publicEngineJSONValue(value)
	}
	return result
}

func publicEngineJSONValue(value native.EngineJSONValue) EngineJSONValue {
	if value.Object != nil {
		value.Object = publicEngineJSONFields(value.Object)
	}
	if value.Array != nil {
		array := make([]EngineJSONValue, len(value.Array))
		for index, item := range value.Array {
			array[index] = publicEngineJSONValue(item)
		}
		value.Array = array
	}
	return value
}

// Error returns the EngineError message.
func (e *EngineError) Error() string {
	if e == nil {
		return "sidereon: engine error"
	}
	name := e.FamilyName
	if name == "" {
		name = e.Family.Name()
	}
	var msg string
	if e.Operation != "" && e.Kind != "" {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d) op=%s kind=%s", name, e.Family, e.Operation, e.Kind)
	} else if e.Operation != "" {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d) op=%s", name, e.Family, e.Operation)
	} else if e.Kind != "" {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d) kind=%s", name, e.Family, e.Kind)
	} else {
		msg = fmt.Sprintf("sidereon: engine error family=%s (%d)", name, e.Family)
	}
	if e.CaptureError != nil {
		msg += fmt.Sprintf("; capture error: %v", e.CaptureError)
	}
	return msg
}

// Unwrap returns the underlying cause when one is retained.
func (e *EngineError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.CaptureError
}

// UnmarshalFields decodes e.Fields into target using json.Number so that integer
// lexemes are preserved without lossy float64 conversion.
func (e *EngineError) UnmarshalFields(target any) error {
	if len(e.Fields) == 0 {
		return errors.New("sidereon: engine error has no fields")
	}
	dec := json.NewDecoder(bytes.NewReader(e.Fields))
	dec.UseNumber()
	return dec.Decode(target)
}

// ClearEngineError explicitly clears the versioned generic engine error slot on
// the current OS thread.
func ClearEngineError() {
	native.ClearEngineError()
}

// FallbackStatus is the distinct status namespace returned by the
// precise/broadcast fallback solver.
type FallbackStatus uint32

const (
	// FallbackOK indicates successful fallback completion.
	FallbackOK FallbackStatus = 0
	// FallbackNullPointer indicates a required null pointer.
	FallbackNullPointer FallbackStatus = 1
	// FallbackInvalidArgument indicates an invalid argument.
	FallbackInvalidArgument FallbackStatus = 2
	// FallbackInvalidToken indicates an invalid textual token.
	FallbackInvalidToken FallbackStatus = 3
	// FallbackPanic indicates a native panic boundary failure.
	FallbackPanic FallbackStatus = 4
	// FallbackPreciseSolve indicates failure in precise solving.
	FallbackPreciseSolve FallbackStatus = 5
	// FallbackBroadcastSolve indicates failure in broadcast solving.
	FallbackBroadcastSolve FallbackStatus = 6
)

func (s FallbackStatus) name() string {
	switch s {
	case FallbackOK:
		return "ok"
	case FallbackNullPointer:
		return "null pointer"
	case FallbackInvalidArgument:
		return "invalid argument"
	case FallbackInvalidToken:
		return "invalid token"
	case FallbackPanic:
		return "panic"
	case FallbackPreciseSolve:
		return "precise solve"
	case FallbackBroadcastSolve:
		return "broadcast solve"
	default:
		return "unknown"
	}
}

// FallbackError reports one failed fallback solve. Detail is the native
// thread-local reason captured during the same call.
type FallbackError struct {
	// Status is the fallback solver's distinct status namespace.
	Status FallbackStatus
	// Detail is same-call native diagnostic text.
	Detail string
}

// Error describes a fallback solve failure including its status detail.
func (e *FallbackError) Error() string {
	if e == nil {
		return "sidereon fallback solve failed"
	}
	message := fmt.Sprintf("sidereon fallback solve failed: %s (status %d)", e.Status.name(), e.Status)
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	return message
}

// Error formats the native status and thread-local detail.
func (e *StatusError) Error() string {
	if e.Detail == "" {
		return e.Text
	}
	return fmt.Sprintf("%s: %s", e.Text, e.Detail)
}

// Unwrap exposes structured SP3, terrain, bias, RTCM, and engine causes attached to
// the status.
func (e *StatusError) Unwrap() error {
	var details []error
	if e.Engine != nil {
		details = append(details, e.Engine)
	}
	if detail := e.TDMError(); detail != nil {
		details = append(details, detail)
	}
	if e.TerrainDatum != nil {
		details = append(details, e.TerrainDatum)
	}
	if e.TerrainStore != nil {
		details = append(details, e.TerrainStore)
	}
	if e.TerrainLookup != nil {
		details = append(details, e.TerrainLookup)
	}
	if e.Bias != nil {
		details = append(details, e.Bias)
	}
	if e.RTCM != nil {
		details = append(details, e.RTCM)
	}
	if e.SP3 != nil {
		details = append(details, e.SP3)
	}
	if e.ANTEX != nil {
		details = append(details, e.ANTEX)
	}
	if e.DtedTile != nil {
		details = append(details, e.DtedTile)
	}
	if e.Geoid != nil {
		details = append(details, e.Geoid)
	}
	if e.PreciseArtifact != nil {
		details = append(details, e.PreciseArtifact)
	}
	return errors.Join(details...)
}

// EngineError returns the typed engine error attached to the current thread-local native response.
func (e *StatusError) EngineError() *EngineError {
	if e == nil {
		return nil
	}
	return e.Engine
}

// TDMError returns the typed TDM detail captured with this native failure.
func (e *StatusError) TDMError() *TDMErrorDetail {
	if e == nil {
		return nil
	}
	if e.Engine == nil || e.Engine.Family != EngineErrorFamilyTdm || e.Engine.CaptureError != nil {
		return nil
	}
	return &TDMErrorDetail{Kind: e.Engine.Kind, Fields: publicEngineJSONFields(e.Engine.TypedFields), Display: e.Detail}
}

// ErrClosed is returned when an operation uses a handle after Close.
var ErrClosed = errors.New("sidereon: handle is closed")

func nativeCountToInt(value uint64, field string) (int, error) {
	if value > uint64(^uint(0)>>1) {
		return 0, fmt.Errorf("sidereon: native %s %d does not fit in int", field, value)
	}
	return int(value), nil
}

var errNilNativeHandle = errors.New("sidereon: native constructor returned no handle")

// Error formats a terrain-datum acquisition failure.
func (e *TerrainDatumError) Error() string {
	if e == nil {
		return "sidereon: terrain datum error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Remediation != "" {
		return e.Remediation
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain datum error"
}

// Error formats a terrain-store acquisition failure.
func (e *TerrainStoreError) Error() string {
	if e == nil {
		return "sidereon: terrain store error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return e.Reason
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain store error"
}

func publicEngineError(value *native.EngineError) *EngineError {
	if value == nil {
		return nil
	}
	return &EngineError{
		Family:       EngineErrorFamily(value.Family),
		FamilyName:   value.FamilyName,
		Schema:       value.Schema,
		Operation:    value.Operation,
		Kind:         value.Kind,
		Fields:       append(json.RawMessage(nil), value.Fields...),
		TypedFields:  publicEngineJSONFields(value.TypedFields),
		Payload:      append(json.RawMessage(nil), value.Payload...),
		CaptureError: value.CaptureError,
	}
}

func publicError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, native.ErrClosed) {
		return ErrClosed
	}
	var unresolved *native.NativeFDEUnresolvedError
	if errors.As(err, &unresolved) {
		var normalized []RAIMNormalizedResidual
		if unresolved.HasNormalized {
			normalized = make([]RAIMNormalizedResidual, len(unresolved.Normalized))
			for i, row := range unresolved.Normalized {
				normalized[i] = RAIMNormalizedResidual{row.SatelliteID, row.NormalizedResidual}
			}
		}
		var raim RAIMResult
		captureErr := unresolved.CaptureError
		if unresolved.HasRAIM {
			var conversionErr error
			raim, conversionErr = publicRAIMResult(unresolved.RAIM)
			if conversionErr != nil {
				captureErr = errors.Join(captureErr, conversionErr)
			}
		}
		public := &FDEUnresolvedError{
			Reason: FDEUnresolvedReason(unresolved.Reason), HasSolution: unresolved.HasSolution,
			HasExcluded: unresolved.HasExcluded, HasRAIM: unresolved.HasRAIM,
			HasNormalizedResiduals: unresolved.HasNormalized, RAIM: raim,
			NormalizedResiduals: normalized, CaptureError: captureErr,
			cause: publicError(unresolved.Cause),
		}
		if unresolved.HasSolution {
			public.Solution = publicSPPSolution(unresolved.Solution)
		}
		if unresolved.HasExcluded {
			public.Excluded = append([]string(nil), unresolved.Excluded...)
			public.Iterations = len(unresolved.Excluded)
		}
		return public
	}
	var selectionErr *native.SelectionError
	if errors.As(err, &selectionErr) {
		return &SelectionError{
			Status: SelectionStatus(selectionErr.Status),
			Text:   selectionErr.Text,
			Detail: selectionErr.Detail,
		}
	}
	var statusErr *native.StatusError
	if errors.As(err, &statusErr) {
		result := &StatusError{
			Code:   StatusCode(statusErr.Code),
			Text:   statusErr.Text,
			Detail: statusErr.Detail,
		}
		if statusErr.Engine != nil {
			result.Engine = publicEngineError(statusErr.Engine)
		}
		if statusErr.SGP4 != nil {
			value := statusErr.SGP4
			result.SGP4 = &SGP4Error{Kind: SGP4ErrorKind(value.Kind), HasCode: value.HasCode, Code: value.Code, HasBudget: value.HasBudget, Budget: value.Budget}
		}
		if statusErr.SP3 != nil {
			value := statusErr.SP3
			result.SP3 = &SP3Error{
				Kind: SP3ErrorKind(value.Kind), Field: SP3ErrorField(value.Field), Reason: SP3ErrorReason(value.Reason),
				HasValue: value.HasValue, Value: value.Value, HasRequestedTick: value.HasRequestedTick,
				HasDeclaredTick: value.HasDeclaredTick, HasRequestedJ2000S: value.HasRequestedJ2000S,
				HasDeclaredJ2000S: value.HasDeclaredJ2000S, RequestedJ2000S: value.RequestedJ2000S,
				DeclaredJ2000S: value.DeclaredJ2000S, Payload: append(json.RawMessage(nil), value.PayloadJSON...),
			}
		}
		if statusErr.TerrainDatum != nil {
			value := *statusErr.TerrainDatum
			result.TerrainDatum = &TerrainDatumError{
				Kind:        TerrainDatumErrorKind(value.Kind),
				Path:        value.Path,
				Message:     value.Message,
				Remediation: value.Remediation,
				Terrain:     value.Terrain,
				Geoid:       value.Geoid,
			}
		}
		if statusErr.TerrainStore != nil {
			value := *statusErr.TerrainStore
			result.TerrainStore = &TerrainStoreError{
				Kind:               TerrainStoreErrorKind(value.Kind),
				Path:               value.Path,
				Message:            value.Message,
				Reason:             value.Reason,
				Version:            value.Version,
				Tag:                value.Tag,
				LatIndex:           value.LatIndex,
				LonIndex:           value.LonIndex,
				ExpectedChecksum:   value.ExpectedChecksum,
				FoundChecksum:      value.FoundChecksum,
				ExpectedTileID:     TerrainTileID{LatIndex: value.ExpectedTileID.LatIndex, LonIndex: value.ExpectedTileID.LonIndex},
				FoundTileID:        TerrainTileID{LatIndex: value.FoundTileID.LatIndex, LonIndex: value.FoundTileID.LonIndex},
				Field:              value.Field,
				HasHorizontalDatum: value.HasHorizontalDatum,
				HorizontalDatum:    value.HorizontalDatum,
				HasTileError:       value.HasTileError,
				TileError:          value.TileError,
			}
		}
		if statusErr.TerrainLookup != nil {
			value := *statusErr.TerrainLookup
			result.TerrainLookup = &value
		}
		if statusErr.DtedTile != nil {
			value := *statusErr.DtedTile
			result.DtedTile = &value
		}
		if statusErr.Geoid != nil {
			value := *statusErr.Geoid
			result.Geoid = &value
		}
		if statusErr.PreciseArtifact != nil {
			value := statusErr.PreciseArtifact
			result.PreciseArtifact = &PreciseArtifactError{Kind: PreciseInterpolantArtifactError(value.Kind), Version: value.Version, Tag: value.Tag, FoundMagic: value.FoundMagic, Available: value.Available, Declared: value.Declared, Offset: value.Offset, Length: value.Length, ExpectedChecksum: value.ExpectedChecksum, FoundChecksum: value.FoundChecksum, ClaimedChecksum: value.ClaimedChecksum, DeclaredChecksum: value.DeclaredChecksum, HasSatellite: value.HasSatellite, Satellite: value.Satellite, Path: value.Path, Message: value.Message, Reason: value.Reason, Region: value.Region}
		}
		if statusErr.Bias != nil {
			value := statusErr.Bias
			result.Bias = &BiasError{
				Kind: BiasErrorKind(value.Kind), Line: value.Line, Record: value.Record,
				HasTimeScale: value.HasTimeScale, TimeScale: TimeScale(value.TimeScale),
				Departure: BiasNotice{Kind: BiasNoticeKind(value.Departure.Kind), Departure: BiasDepartureKind(value.Departure.Departure), HasLine: value.Departure.HasLine, Line: value.Departure.Line, First: value.Departure.First, Second: value.Departure.Second, DeclaredCount: value.Departure.DeclaredCount, SolutionRows: value.Departure.SolutionRows, BiasMode: BiasMode(value.Departure.BiasMode), UnknownVariant: value.Departure.UnknownVariant},
				Message:   append([]byte(nil), value.Message...), Field: append([]byte(nil), value.Field...),
				Reason: append([]byte(nil), value.Reason...), Code: append([]byte(nil), value.Code...), Version: append([]byte(nil), value.Version...),
			}
			for index := range value.DepartureText {
				result.Bias.DepartureText[index] = append([]byte(nil), value.DepartureText[index]...)
			}
		}
		if statusErr.RTCM != nil {
			value := statusErr.RTCM
			result.RTCM = &RTCMError{Class: RTCMErrorClass(value.Class), Kind: value.Kind, Payload: append(json.RawMessage(nil), value.Payload...)}
		}
		if statusErr.ANTEX != nil {
			value := statusErr.ANTEX
			result.ANTEX = &ANTEXError{Kind: ANTEXErrorKind(value.Kind), HasAntennaID: value.HasAntennaID, HasRecord: value.HasRecord, HasField: value.HasField, HasValue: value.HasValue, HasFrequency: value.HasFrequency, HasReason: value.HasReason, HasSections: value.HasSections, Sections: value.Sections, AntennaID: value.AntennaID, Record: value.Record, Field: value.Field, Value: value.Value, Frequency: value.Frequency, Reason: value.Reason, Message: value.Message}
		}
		if statusErr.QualityKind != 0 {
			message := result.Detail
			if message == "" {
				message = result.Error()
			}
			return &QualityError{Kind: QualityErrorKind(statusErr.QualityKind), Message: message, cause: result}
		}
		return result
	}
	var fallbackErr *native.FallbackError
	if errors.As(err, &fallbackErr) {
		return &FallbackError{Status: FallbackStatus(fallbackErr.Status), Detail: fallbackErr.Detail}
	}
	return err
}

func publicRAIMResult(value native.NativeRaimResult) (RAIMResult, error) {
	result := RAIMResult{
		FaultDetected: value.FaultDetected, TestStatistic: value.TestStatistic,
		HasThreshold: value.HasThreshold, Threshold: value.Threshold,
		HasReducedChiSquare: value.HasReducedChiSquare, ReducedChiSquare: value.ReducedChiSquare,
		RMSM: value.RMSM, DOF: value.DOF, Testable: value.Testable,
		HasWorstSatellite: value.HasWorstSatellite, WorstSatellite: value.WorstSatellite,
	}
	count, err := nativeCountToInt(value.NormalizedResidualCount, "RAIM normalized residual count")
	if err != nil {
		return result, err
	}
	result.NormalizedResidualCount = count
	return result, nil
}

func joinPublicErrors(errs ...error) error {
	translated := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			translated = append(translated, publicError(err))
		}
	}
	return errors.Join(translated...)
}
