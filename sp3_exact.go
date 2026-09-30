package sidereon

import (
	"fmt"

	"sidereon.dev/go/v3/internal/native"
)

// SP3AccuracyValueKind identifies whether an accuracy value is usable.
type SP3AccuracyValueKind uint32

const (
	// SP3AccuracyKnown means the decoded accuracy value is within the supported limit.
	SP3AccuracyKnown SP3AccuracyValueKind = iota
	// SP3AccuracyUnknown means the record does not provide a usable accuracy value.
	SP3AccuracyUnknown
	// SP3AccuracyTooLarge means the accuracy exponent exceeds the supported limit.
	SP3AccuracyTooLarge
	// SP3AccuracyInvalidBase means the accuracy base is invalid.
	SP3AccuracyInvalidBase
	// SP3AccuracyOverflow means decoding overflowed the numeric range.
	SP3AccuracyOverflow
	// SP3AccuracyOther preserves another native accuracy status.
	SP3AccuracyOther
)

// SP3AccuracyValue carries a variance or sigma and its interpretation status.
type SP3AccuracyValue struct {
	Kind  SP3AccuracyValueKind
	Value float64
}

// SP3AccuracyCodeGroup preserves raw record-level accuracy exponents and bases.
type SP3AccuracyCodeGroup struct {
	HasAxisExponents        [3]bool
	AxisExponents           [3]int16
	HasClockExponent        bool
	ClockExponent           int16
	HasPositionVelocityBase bool
	PositionVelocityBase    float64
	HasClockRateBase        bool
	ClockRateBase           float64
}

// SP3RawRecordAccuracy preserves raw P- and V-record accuracy fields.
type SP3RawRecordAccuracy struct {
	HasP bool
	P    SP3AccuracyCodeGroup
	HasV bool
	V    SP3AccuracyCodeGroup
}

// SP3PositionClockAccuracy contains decoded position and clock accuracy values.
type SP3PositionClockAccuracy struct {
	PositionSigmaM     [3]SP3AccuracyValue
	ClockSigmaM        SP3AccuracyValue
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

// SP3VelocityAccuracy contains decoded velocity and clock-rate accuracy values.
type SP3VelocityAccuracy struct {
	VelocitySigmaMPerS       [3]SP3AccuracyValue
	ClockRateSigmaMPerS      SP3AccuracyValue
	VelocityVarianceM2PerS2  [3]SP3AccuracyValue
	ClockRateVarianceM2PerS2 SP3AccuracyValue
}

// SP3RecordAccuracy contains decoded record-level P and V accuracy values.
type SP3RecordAccuracy struct {
	HasP bool
	P    SP3PositionClockAccuracy
	HasV bool
	V    SP3VelocityAccuracy
}

// PreciseEphemerisAccuracySample pairs one epoch label with decoded variances.
type PreciseEphemerisAccuracySample struct {
	Satellite string
	TimeScale TimeScale
	// EpochJ2000S is the legacy rounded scalar label, not a lossless Instant.
	EpochJ2000S        float64
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

// PreciseSamplesErrorKind identifies native precise-sample validation failures.
type PreciseSamplesErrorKind uint32

const (
	// PreciseSamplesErrorNone means sample validation succeeded.
	PreciseSamplesErrorNone PreciseSamplesErrorKind = iota
	// PreciseSamplesErrorEmpty means no precise samples were supplied.
	PreciseSamplesErrorEmpty
	// PreciseSamplesErrorSingleSampleSatellite means a satellite has only one sample.
	PreciseSamplesErrorSingleSampleSatellite
	// PreciseSamplesErrorNonMonotonicEpochs means sample epochs are not strictly increasing.
	PreciseSamplesErrorNonMonotonicEpochs
	// PreciseSamplesErrorMixedTimeScales means samples use incompatible time scales.
	PreciseSamplesErrorMixedTimeScales
	// PreciseSamplesErrorEpochNotRepresentable means an epoch cannot be represented by the exact time type.
	PreciseSamplesErrorEpochNotRepresentable
	// PreciseSamplesErrorNonFiniteSample means a sample contains a non-finite value.
	PreciseSamplesErrorNonFiniteSample
	// PreciseSamplesErrorAccuracySamplesMismatch means accuracy samples do not align with the ephemeris samples.
	PreciseSamplesErrorAccuracySamplesMismatch
	// PreciseSamplesErrorInvalidAccuracyValue means an accuracy value violates its encoding constraints.
	PreciseSamplesErrorInvalidAccuracyValue
	// PreciseSamplesErrorOther preserves another native sample-validation failure.
	PreciseSamplesErrorOther
)

// PreciseSamplesValidationError retains the native validation kind and context.
type PreciseSamplesValidationError struct {
	Kind         PreciseSamplesErrorKind
	HasSatellite bool
	Satellite    string
	Cause        error
}

// Error returns the PreciseSamplesValidationError message.
func (e *PreciseSamplesValidationError) Error() string {
	if e == nil {
		return "sidereon: precise sample validation failed"
	}
	if e.Cause != nil {
		return fmt.Sprintf("sidereon: precise sample validation kind %d: %v", e.Kind, e.Cause)
	}
	if e.HasSatellite {
		return fmt.Sprintf("sidereon: precise sample validation kind %d for %s", e.Kind, e.Satellite)
	}
	return fmt.Sprintf("sidereon: precise sample validation kind %d", e.Kind)
}

// Unwrap returns the underlying cause when one is retained.
func (e *PreciseSamplesValidationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func publicSP3AccuracyValue(value native.SP3AccuracyValue) SP3AccuracyValue {
	return SP3AccuracyValue{Kind: SP3AccuracyValueKind(value.Kind), Value: value.Value}
}

func publicSP3State(value native.SP3State) SP3State {
	return SP3State{
		PositionM: value.PositionM, HasClock: value.HasClock, ClockS: value.ClockS,
		HasVelocity: value.HasVelocity, VelocityMPerS: value.VelocityMPerS,
		HasClockRate: value.HasClockRate, ClockRateSPerS: value.ClockRateSPerS,
		ClockEvent: value.ClockEvent, ClockPredicted: value.ClockPredicted,
		Maneuver: value.Maneuver, OrbitPredicted: value.OrbitPredicted,
	}
}

// StateAt evaluates one SP3 satellite at an exact query epoch.
func (s *SP3) StateAt(query *ExactEpochQuery, satellite string) (SP3State, error) {
	if s == nil || s.handle == nil || query == nil || query.handle == nil {
		return SP3State{}, ErrClosed
	}
	value, err := s.handle.StateAtEpochQuery(query.handle, satellite)
	if err != nil {
		return SP3State{}, publicError(err)
	}
	return publicSP3State(value), nil
}

// RecordAccuracyCodes returns raw SP3 accuracy codes for one satellite epoch.
func (s *SP3) RecordAccuracyCodes(satellite string, epochIndex int) (SP3RawRecordAccuracy, error) {
	if s == nil || s.handle == nil {
		return SP3RawRecordAccuracy{}, ErrClosed
	}
	value, err := s.handle.RecordAccuracyCodes(satellite, epochIndex)
	if err != nil {
		return SP3RawRecordAccuracy{}, publicError(err)
	}
	return SP3RawRecordAccuracy{
		HasP: value.HasP,
		P: SP3AccuracyCodeGroup{
			HasAxisExponents: value.P.HasAxisExponents, AxisExponents: value.P.AxisExponents,
			HasClockExponent: value.P.HasClockExponent, ClockExponent: value.P.ClockExponent,
			HasPositionVelocityBase: value.P.HasPositionVelocityBase,
			PositionVelocityBase:    value.P.PositionVelocityBase,
			HasClockRateBase:        value.P.HasClockRateBase, ClockRateBase: value.P.ClockRateBase,
		},
		HasV: value.HasV,
		V: SP3AccuracyCodeGroup{
			HasAxisExponents: value.V.HasAxisExponents, AxisExponents: value.V.AxisExponents,
			HasClockExponent: value.V.HasClockExponent, ClockExponent: value.V.ClockExponent,
			HasPositionVelocityBase: value.V.HasPositionVelocityBase,
			PositionVelocityBase:    value.V.PositionVelocityBase,
			HasClockRateBase:        value.V.HasClockRateBase, ClockRateBase: value.V.ClockRateBase,
		},
	}, nil
}

// RecordAccuracy returns decoded SP3 sigma and variance values for one record.
func (s *SP3) RecordAccuracy(satellite string, epochIndex int) (SP3RecordAccuracy, error) {
	if s == nil || s.handle == nil {
		return SP3RecordAccuracy{}, ErrClosed
	}
	value, err := s.handle.RecordAccuracy(satellite, epochIndex)
	if err != nil {
		return SP3RecordAccuracy{}, publicError(err)
	}
	out := SP3RecordAccuracy{HasP: value.HasP, HasV: value.HasV}
	for axis := 0; axis < 3; axis++ {
		out.P.PositionSigmaM[axis] = publicSP3AccuracyValue(value.P.PositionSigmaM[axis])
		out.P.PositionVarianceM2[axis] = publicSP3AccuracyValue(value.P.PositionVarianceM2[axis])
		out.V.VelocitySigmaMPerS[axis] = publicSP3AccuracyValue(value.V.VelocitySigmaMPerS[axis])
		out.V.VelocityVarianceM2PerS2[axis] = publicSP3AccuracyValue(value.V.VelocityVarianceM2PerS2[axis])
	}
	out.P.ClockSigmaM = publicSP3AccuracyValue(value.P.ClockSigmaM)
	out.P.ClockVarianceM2 = publicSP3AccuracyValue(value.P.ClockVarianceM2)
	out.V.ClockRateSigmaMPerS = publicSP3AccuracyValue(value.V.ClockRateSigmaMPerS)
	out.V.ClockRateVarianceM2PerS2 = publicSP3AccuracyValue(value.V.ClockRateVarianceM2PerS2)
	return out, nil
}

// PreciseAccuracySamples copies decoded variance samples from the SP3 product.
func (s *SP3) PreciseAccuracySamples() ([]PreciseEphemerisAccuracySample, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.PreciseAccuracySamples()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PreciseEphemerisAccuracySample, len(values))
	for i, value := range values {
		out[i] = PreciseEphemerisAccuracySample{
			Satellite: value.Satellite, TimeScale: TimeScale(value.TimeScale), EpochJ2000S: value.EpochJ2000S,
			ClockVarianceM2: publicSP3AccuracyValue(value.ClockVarianceM2),
		}
		for axis := 0; axis < 3; axis++ {
			out[i].PositionVarianceM2[axis] = publicSP3AccuracyValue(value.PositionVarianceM2[axis])
		}
	}
	return out, nil
}

func nativeAccuracySample(value PreciseEphemerisAccuracySample) native.PreciseEphemerisAccuracySample {
	out := native.PreciseEphemerisAccuracySample{
		Satellite: value.Satellite, TimeScale: uint32(value.TimeScale), EpochJ2000S: value.EpochJ2000S,
		ClockVarianceM2: native.SP3AccuracyValue{Kind: uint32(value.ClockVarianceM2.Kind), Value: value.ClockVarianceM2.Value},
	}
	for axis := 0; axis < 3; axis++ {
		out.PositionVarianceM2[axis] = native.SP3AccuracyValue{
			Kind: uint32(value.PositionVarianceM2[axis].Kind), Value: value.PositionVarianceM2[axis].Value,
		}
	}
	return out
}

// BuildPreciseEphemerisSamplesWithAccuracy builds samples with decoded accuracy.
func BuildPreciseEphemerisSamplesWithAccuracy(samples []PreciseEphemerisSample, accuracy []PreciseEphemerisAccuracySample, opts ...SP3Option) (*PreciseEphemerisSamples, error) {
	options := resolveSP3Options(opts)
	nativeSamples := make([]native.PreciseEphemerisSample, len(samples))
	for i := range samples {
		nativeSamples[i] = nativePreciseSample(samples[i])
	}
	nativeAccuracy := make([]native.PreciseEphemerisAccuracySample, len(accuracy))
	for i := range accuracy {
		nativeAccuracy[i] = nativeAccuracySample(accuracy[i])
	}
	handle, nativeError, err := native.PreciseEphemerisSamplesFromSamplesWithAccuracy(nativeSamples, nativeAccuracy, options.GapThresholdFactor)
	if err != nil {
		if nativeError.Kind == native.PreciseSamplesErrorNone {
			return nil, publicError(err)
		}
		publicValidationError := &PreciseSamplesValidationError{
			Kind: PreciseSamplesErrorKind(nativeError.Kind), HasSatellite: nativeError.HasSatellite,
			Satellite: nativeError.Satellite, Cause: publicError(err),
		}
		return nil, publicValidationError
	}
	return &PreciseEphemerisSamples{handle: handle}, nil
}

// BuildPreciseEphemerisInterpolantWithAccuracy builds an interpolant with decoded accuracy.
func BuildPreciseEphemerisInterpolantWithAccuracy(samples []PreciseEphemerisSample, accuracy []PreciseEphemerisAccuracySample, opts ...SP3Option) (*PreciseEphemerisInterpolant, error) {
	options := resolveSP3Options(opts)
	nativeSamples := make([]native.PreciseEphemerisSample, len(samples))
	for i := range samples {
		nativeSamples[i] = nativePreciseSample(samples[i])
	}
	nativeAccuracy := make([]native.PreciseEphemerisAccuracySample, len(accuracy))
	for i := range accuracy {
		nativeAccuracy[i] = nativeAccuracySample(accuracy[i])
	}
	handle, nativeError, err := native.PreciseEphemerisInterpolantFromSamplesWithAccuracy(nativeSamples, nativeAccuracy, options.GapThresholdFactor)
	if err != nil {
		if nativeError.Kind == native.PreciseSamplesErrorNone {
			return nil, publicError(err)
		}
		return nil, &PreciseSamplesValidationError{
			Kind: PreciseSamplesErrorKind(nativeError.Kind), HasSatellite: nativeError.HasSatellite,
			Satellite: nativeError.Satellite, Cause: publicError(err),
		}
	}
	return &PreciseEphemerisInterpolant{handle: handle}, nil
}

// StateAt evaluates one interpolant satellite at an exact query epoch.
func (i *PreciseEphemerisInterpolant) StateAt(query *ExactEpochQuery, satellite string) (SP3State, error) {
	if i == nil || i.handle == nil || query == nil || query.handle == nil {
		return SP3State{}, ErrClosed
	}
	value, err := i.handle.StateAtEpochQuery(query.handle, satellite)
	if err != nil {
		return SP3State{}, publicError(err)
	}
	return publicSP3State(value), nil
}

// StateAt evaluates the artifact at an exact query epoch.
func (a *PreciseInterpolantArtifact) StateAt(query *ExactEpochQuery, satellite string) (SP3State, error) {
	if a == nil || a.handle == nil || query == nil || query.handle == nil {
		return SP3State{}, ErrClosed
	}
	value, err := a.handle.StateAtEpochQuery(query.handle, satellite)
	if err != nil {
		return SP3State{}, publicError(err)
	}
	return publicSP3State(value), nil
}

// SSRCorrectionSizePolicy selects refusal or reporting for oversized SSR corrections.
type SSRCorrectionSizePolicy uint32

const (
	// SSRCorrectionSizeStrict refuses oversized corrections.
	SSRCorrectionSizeStrict SSRCorrectionSizePolicy = iota
	// SSRCorrectionSizeLenient reports oversized corrections while applying them.
	SSRCorrectionSizeLenient
)

// SSRCorrectionSize contains orbit and clock correction magnitudes in metres.
type SSRCorrectionSize struct {
	OrbitM float64
	ClockM float64
}

// SSRCorrectionSizeRefusalError reports a strict refusal for an oversized SSR
// correction. StateEpoch and SelectionEpoch borrow the query handles supplied
// to CorrectedStateAtEpochQueriesChecked and remain owned by the caller.
type SSRCorrectionSizeRefusalError struct {
	Satellite      string
	StateEpoch     *ExactEpochQuery
	SelectionEpoch *ExactEpochQuery
	Size           SSRCorrectionSize
}

// Error describes the strict SSR size refusal and its correction magnitudes.
func (refusal *SSRCorrectionSizeRefusalError) Error() string {
	if refusal == nil {
		return "sidereon: strict SSR correction-size refusal"
	}
	return fmt.Sprintf("sidereon: strict SSR correction-size refusal for %s (orbit %.9g m, clock %.9g m)",
		refusal.Satellite, refusal.Size.OrbitM, refusal.Size.ClockM)
}

// SSRCorrectedState contains the corrected state and SSR selection diagnostics.
type SSRCorrectedState struct {
	HasState                bool
	PositionECEFM           [3]float64
	ClockS                  float64
	HasGroupDelay           bool
	GroupDelayS             float64
	Degraded                bool
	DegradeReason           UT1DegradeReason
	HasSizeEvent            bool
	StrictRefusal           bool
	Size                    SSRCorrectionSize
	HasOversizedReport      bool
	Source                  SSRSource
	ProviderID              uint16
	SolutionID              uint8
	OrbitRefEpochJ2000S     float64
	ClockRefEpochJ2000S     float64
	FirstAppliedEpochJ2000S float64
}

// CorrectedStateAtEpochQueries evaluates corrected states for the supplied exact epoch queries.
func (s *SSRCorrectionStore) CorrectedStateAtEpochQueries(broadcast *BroadcastEphemeris, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy) (SSRCorrectedState, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return SSRCorrectedState{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return SSRCorrectedState{}, err
	}
	if sizePolicy != SSRCorrectionSizeStrict && sizePolicy != SSRCorrectionSizeLenient {
		return SSRCorrectedState{}, fmt.Errorf("sidereon: invalid SSR correction size policy %d", sizePolicy)
	}
	value, err := native.CorrectedStateAtEpochQueries(
		broadcast.handle, s.handle, satellite, stateEpoch.handle, selectionEpoch.handle,
		stalenessSeconds, uint32(missing), allowRegionalProvider, regionalProviderID, uint32(sizePolicy),
	)
	if err != nil {
		return SSRCorrectedState{}, publicError(err)
	}
	return SSRCorrectedState{
		HasState: value.HasState, PositionECEFM: value.PositionECEFM, ClockS: value.ClockS,
		HasGroupDelay: value.HasGroupDelay, GroupDelayS: value.GroupDelayS, Degraded: value.Degraded,
		HasSizeEvent: value.HasSizeEvent, StrictRefusal: value.StrictRefusal,
		Size:               SSRCorrectionSize{OrbitM: value.Size.OrbitM, ClockM: value.Size.ClockM},
		HasOversizedReport: value.HasOversizedReport, Source: SSRSource(value.Source),
		ProviderID: value.ProviderID, SolutionID: value.SolutionID,
		OrbitRefEpochJ2000S:     value.OrbitRefEpochJ2000S,
		ClockRefEpochJ2000S:     value.ClockRefEpochJ2000S,
		FirstAppliedEpochJ2000S: value.FirstAppliedEpochJ2000S,
	}, nil
}

// CorrectedStateAtEpochQueriesChecked applies the query-time SSR correction
// policy and returns a typed error when strict size policy refuses a result.
// The result still carries refusal flags and size diagnostics alongside the
// error. Error query fields borrow the supplied handles.
func (s *SSRCorrectionStore) CorrectedStateAtEpochQueriesChecked(broadcast *BroadcastEphemeris, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy) (SSRCorrectedState, error) {
	result, err := s.CorrectedStateAtEpochQueries(
		broadcast, satellite, stateEpoch, selectionEpoch, stalenessSeconds,
		missing, allowRegionalProvider, regionalProviderID, sizePolicy,
	)
	if err != nil {
		return result, err
	}
	return checkedSSRCorrectionResult(result, satellite, stateEpoch, selectionEpoch)
}

func checkedSSRCorrectionResult(result SSRCorrectedState, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery) (SSRCorrectedState, error) {
	if !result.StrictRefusal {
		return result, nil
	}
	return result, &SSRCorrectionSizeRefusalError{
		Satellite: satellite, StateEpoch: stateEpoch, SelectionEpoch: selectionEpoch, Size: result.Size,
	}
}
