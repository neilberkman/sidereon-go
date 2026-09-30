package sidereon

import "sidereon.dev/go/v3/internal/native"

// PreciseEphemerisSampleV2 retains the source epoch's tagged instant.
type PreciseEphemerisSampleV2 struct {
	Satellite     string
	Epoch         ClockEpoch
	PositionECEFM [3]float64
	HasClock      bool
	ClockS        float64
	ClockEvent    bool
}

// PreciseEphemerisAccuracySampleV2 retains a tagged epoch and decoded variances.
type PreciseEphemerisAccuracySampleV2 struct {
	Satellite          string
	Epoch              ClockEpoch
	PositionVarianceM2 [3]SP3AccuracyValue
	ClockVarianceM2    SP3AccuracyValue
}

func publicClockEpoch(value native.NativeClockEpoch) ClockEpoch {
	return ClockEpoch{
		Scale: TimeScale(value.Scale), Representation: RINEXClockInstantRepresentation(value.Representation),
		JulianWhole: value.JulianWhole, JulianFraction: value.JulianFraction,
		NanosHigh: value.NanosHigh, NanosLow: value.NanosLow,
	}
}

func nativeClockEpoch(value ClockEpoch) native.NativeClockEpoch {
	return native.NativeClockEpoch{
		Scale: uint32(value.Scale), Representation: uint32(value.Representation),
		JulianWhole: value.JulianWhole, JulianFraction: value.JulianFraction,
		NanosHigh: value.NanosHigh, NanosLow: value.NanosLow,
	}
}

func publicPreciseSampleV2(value native.PreciseEphemerisSampleV2) PreciseEphemerisSampleV2 {
	return PreciseEphemerisSampleV2{
		Satellite: value.Satellite, Epoch: publicClockEpoch(value.Epoch),
		PositionECEFM: value.PositionECEFM, HasClock: value.HasClock,
		ClockS: value.ClockS, ClockEvent: value.ClockEvent,
	}
}

func nativePreciseSampleV2(value PreciseEphemerisSampleV2) native.PreciseEphemerisSampleV2 {
	return native.PreciseEphemerisSampleV2{
		Satellite: value.Satellite, Epoch: nativeClockEpoch(value.Epoch),
		PositionECEFM: value.PositionECEFM, HasClock: value.HasClock,
		ClockS: value.ClockS, ClockEvent: value.ClockEvent,
	}
}

func publicPreciseAccuracySampleV2(value native.PreciseEphemerisAccuracySampleV2) PreciseEphemerisAccuracySampleV2 {
	out := PreciseEphemerisAccuracySampleV2{
		Satellite: value.Satellite, Epoch: publicClockEpoch(value.Epoch),
		ClockVarianceM2: publicSP3AccuracyValue(value.ClockVarianceM2),
	}
	for axis := range out.PositionVarianceM2 {
		out.PositionVarianceM2[axis] = publicSP3AccuracyValue(value.PositionVarianceM2[axis])
	}
	return out
}

func nativePreciseAccuracySampleV2(value PreciseEphemerisAccuracySampleV2) native.PreciseEphemerisAccuracySampleV2 {
	out := native.PreciseEphemerisAccuracySampleV2{
		Satellite: value.Satellite, Epoch: nativeClockEpoch(value.Epoch),
		ClockVarianceM2: native.SP3AccuracyValue{Kind: uint32(value.ClockVarianceM2.Kind), Value: value.ClockVarianceM2.Value},
	}
	for axis := range out.PositionVarianceM2 {
		out.PositionVarianceM2[axis] = native.SP3AccuracyValue{
			Kind: uint32(value.PositionVarianceM2[axis].Kind), Value: value.PositionVarianceM2[axis].Value,
		}
	}
	return out
}

// PreciseSamplesV2 returns canonical SP3 samples without rounding their epochs.
func (s *SP3) PreciseSamplesV2() ([]PreciseEphemerisSampleV2, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.PreciseSamplesV2()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PreciseEphemerisSampleV2, len(values))
	for index, value := range values {
		out[index] = publicPreciseSampleV2(value)
	}
	return out, nil
}

// PreciseAccuracySamplesV2 returns accuracy records with lossless epoch tags.
func (s *SP3) PreciseAccuracySamplesV2() ([]PreciseEphemerisAccuracySampleV2, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.PreciseAccuracySamplesV2()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PreciseEphemerisAccuracySampleV2, len(values))
	for index, value := range values {
		out[index] = publicPreciseAccuracySampleV2(value)
	}
	return out, nil
}

// RecordsV2 returns the source sample records while preserving tagged epochs.
func (s *PreciseEphemerisSamples) RecordsV2() ([]PreciseEphemerisSampleV2, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.RecordsV2()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PreciseEphemerisSampleV2, len(values))
	for index, value := range values {
		out[index] = publicPreciseSampleV2(value)
	}
	return out, nil
}

// AccuracyRecordsV2 returns source accuracy records with tagged epochs.
func (s *PreciseEphemerisSamples) AccuracyRecordsV2() ([]PreciseEphemerisAccuracySampleV2, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.AccuracyRecordsV2()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PreciseEphemerisAccuracySampleV2, len(values))
	for index, value := range values {
		out[index] = publicPreciseAccuracySampleV2(value)
	}
	return out, nil
}

// BuildPreciseEphemerisSamplesV2 builds a source without reducing epoch labels to float64.
// The source retains each accepted tagged epoch exactly. Its interpolation
// nodes use the core's whole-second SP3 time axis, so distinct subsecond epochs
// that quantize to the same node are rejected as non-monotonic.
func BuildPreciseEphemerisSamplesV2(samples []PreciseEphemerisSampleV2, opts ...SP3Option) (*PreciseEphemerisSamples, error) {
	options := resolveSP3Options(opts)
	nativeSamples := make([]native.PreciseEphemerisSampleV2, len(samples))
	for index, sample := range samples {
		nativeSamples[index] = nativePreciseSampleV2(sample)
	}
	handle, nativeError, err := native.PreciseEphemerisSamplesFromSamplesV2(nativeSamples, options.GapThresholdFactor)
	if err != nil {
		if nativeError.Kind == native.PreciseSamplesErrorNone {
			return nil, publicError(err)
		}
		return nil, &PreciseSamplesValidationError{
			Kind: PreciseSamplesErrorKind(nativeError.Kind), HasSatellite: nativeError.HasSatellite,
			Satellite: nativeError.Satellite, Cause: publicError(err),
		}
	}
	return &PreciseEphemerisSamples{handle: handle}, nil
}

// BuildPreciseEphemerisInterpolantV2 builds an interpolant with lossless epoch labels.
// As with BuildPreciseEphemerisSamplesV2, distinct subsecond epochs that map to
// the same whole-second interpolation node are rejected.
func BuildPreciseEphemerisInterpolantV2(samples []PreciseEphemerisSampleV2, opts ...SP3Option) (*PreciseEphemerisInterpolant, error) {
	options := resolveSP3Options(opts)
	nativeSamples := make([]native.PreciseEphemerisSampleV2, len(samples))
	for index, sample := range samples {
		nativeSamples[index] = nativePreciseSampleV2(sample)
	}
	handle, nativeError, err := native.PreciseEphemerisInterpolantFromSamplesV2(nativeSamples, options.GapThresholdFactor)
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

// BuildPreciseEphemerisSamplesWithAccuracyV2 builds lossless samples with decoded variances.
func BuildPreciseEphemerisSamplesWithAccuracyV2(samples []PreciseEphemerisSampleV2, accuracy []PreciseEphemerisAccuracySampleV2, opts ...SP3Option) (*PreciseEphemerisSamples, error) {
	options := resolveSP3Options(opts)
	nativeSamples := make([]native.PreciseEphemerisSampleV2, len(samples))
	for index, sample := range samples {
		nativeSamples[index] = nativePreciseSampleV2(sample)
	}
	nativeAccuracy := make([]native.PreciseEphemerisAccuracySampleV2, len(accuracy))
	for index, sample := range accuracy {
		nativeAccuracy[index] = nativePreciseAccuracySampleV2(sample)
	}
	handle, nativeError, err := native.PreciseEphemerisSamplesFromSamplesWithAccuracyV2(nativeSamples, nativeAccuracy, options.GapThresholdFactor)
	if err != nil {
		if nativeError.Kind == native.PreciseSamplesErrorNone {
			return nil, publicError(err)
		}
		return nil, &PreciseSamplesValidationError{
			Kind: PreciseSamplesErrorKind(nativeError.Kind), HasSatellite: nativeError.HasSatellite,
			Satellite: nativeError.Satellite, Cause: publicError(err),
		}
	}
	return &PreciseEphemerisSamples{handle: handle}, nil
}

// BuildPreciseEphemerisInterpolantWithAccuracyV2 builds a lossless interpolant with variances.
func BuildPreciseEphemerisInterpolantWithAccuracyV2(samples []PreciseEphemerisSampleV2, accuracy []PreciseEphemerisAccuracySampleV2, opts ...SP3Option) (*PreciseEphemerisInterpolant, error) {
	options := resolveSP3Options(opts)
	nativeSamples := make([]native.PreciseEphemerisSampleV2, len(samples))
	for index, sample := range samples {
		nativeSamples[index] = nativePreciseSampleV2(sample)
	}
	nativeAccuracy := make([]native.PreciseEphemerisAccuracySampleV2, len(accuracy))
	for index, sample := range accuracy {
		nativeAccuracy[index] = nativePreciseAccuracySampleV2(sample)
	}
	handle, nativeError, err := native.PreciseEphemerisInterpolantFromSamplesWithAccuracyV2(nativeSamples, nativeAccuracy, options.GapThresholdFactor)
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
