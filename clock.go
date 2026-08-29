package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// AllanSeriesKind identifies the units and missing-sample policy of a clock
// stability series.
type AllanSeriesKind uint32

const (
	// AllanSeriesKindPhaseSeconds contains phase samples in seconds.
	AllanSeriesKindPhaseSeconds AllanSeriesKind = AllanSeriesKind(native.AllanSeriesPhaseSecondsValue)
	// AllanSeriesKindFractionalFrequency contains dimensionless frequency samples.
	AllanSeriesKindFractionalFrequency AllanSeriesKind = AllanSeriesKind(native.AllanSeriesFractionalFrequencyValue)
	// AllanSeriesKindPhaseSecondsWithGaps contains phase seconds with gaps.
	AllanSeriesKindPhaseSecondsWithGaps AllanSeriesKind = AllanSeriesKind(native.AllanSeriesPhaseSecondsWithGapsValue)
	// AllanSeriesKindFractionalFrequencyWithGaps contains dimensionless samples with gaps.
	AllanSeriesKindFractionalFrequencyWithGaps AllanSeriesKind = AllanSeriesKind(native.AllanSeriesFractionalFrequencyWithGapsValue)
)

// AllanSeries is an immutable copy of a clock-stability sample series.
type AllanSeries struct {
	kind    AllanSeriesKind
	samples []native.AllanSample
}

// AllanSample carries an explicitly present or missing clock sample. Value is
// ignored when Present is false.
type AllanSample struct {
	// Present reports whether Value is observed.
	Present bool
	// Value is seconds for phase series or dimensionless for frequency series.
	Value float64
}

// AllanSeriesFromSamples constructs a series while preserving explicit sample
// presence. It is the lossless constructor for gapped inputs.
func AllanSeriesFromSamples(kind AllanSeriesKind, values []AllanSample) AllanSeries {
	samples := make([]native.AllanSample, len(values))
	for index, value := range values {
		samples[index] = native.AllanSample{Present: value.Present, Value: value.Value}
	}
	return AllanSeries{kind: kind, samples: samples}
}

func newAllanSeries(kind AllanSeriesKind, values []float64, gaps bool) AllanSeries {
	samples := make([]native.AllanSample, len(values))
	for index, value := range values {
		samples[index] = native.AllanSample{Present: !gaps || value == value, Value: value}
	}
	return AllanSeries{kind: kind, samples: samples}
}

// AllanSeriesPhaseSeconds constructs phase deviations in seconds.
func AllanSeriesPhaseSeconds(values []float64) AllanSeries {
	return newAllanSeries(AllanSeriesKindPhaseSeconds, values, false)
}

// AllanSeriesFractionalFrequency constructs dimensionless fractional-frequency samples.
func AllanSeriesFractionalFrequency(values []float64) AllanSeries {
	return newAllanSeries(AllanSeriesKindFractionalFrequency, values, false)
}

// AllanSeriesPhaseSecondsWithGaps constructs a phase series where NaN marks a missing sample.
func AllanSeriesPhaseSecondsWithGaps(values []float64) AllanSeries {
	return newAllanSeries(AllanSeriesKindPhaseSecondsWithGaps, values, true)
}

// AllanSeriesFractionalFrequencyWithGaps constructs a fractional-frequency series where NaN marks a missing sample.
func AllanSeriesFractionalFrequencyWithGaps(values []float64) AllanSeries {
	return newAllanSeries(AllanSeriesKindFractionalFrequencyWithGaps, values, true)
}

// SampleCount returns the number of samples, including missing entries.
func (s AllanSeries) SampleCount() int { return len(s.samples) }

// MissingSampleCount returns the number of explicitly missing samples.
func (s AllanSeries) MissingSampleCount() int {
	count := 0
	for _, sample := range s.samples {
		if !sample.Present {
			count++
		}
	}
	return count
}

// Kind returns the sample units and gap representation.
func (s AllanSeries) Kind() AllanSeriesKind { return s.kind }

// AllanTauGrid selects the averaging-factor grid.
type AllanTauGrid uint32

const (
	// AllanTauGridOctave selects powers of two while terms remain available.
	AllanTauGridOctave AllanTauGrid = AllanTauGrid(native.AllanTauGridOctaveValue)
	// AllanTauGridAll selects every valid averaging factor.
	AllanTauGridAll AllanTauGrid = AllanTauGrid(native.AllanTauGridAllValue)
	// AllanTauGridExplicit selects TauGrid.AveragingFactors.
	AllanTauGridExplicit AllanTauGrid = AllanTauGrid(native.AllanTauGridExplicitValue)
)

// TauGrid describes the averaging-factor grid and owns a copy of explicit
// factors. Factors are dimensionless positive integers.
type TauGrid struct {
	// Kind selects the averaging-factor grid.
	Kind AllanTauGrid
	// AveragingFactors are positive factors used by the explicit grid.
	AveragingFactors []int
}

// TauGridOctave constructs an octave grid.
func TauGridOctave() TauGrid { return TauGrid{Kind: AllanTauGridOctave} }

// TauGridAll constructs the complete valid-factor grid.
func TauGridAll() TauGrid { return TauGrid{Kind: AllanTauGridAll} }

// TauGridExplicit constructs an explicit grid and copies factors.
func TauGridExplicit(factors []int) TauGrid {
	return TauGrid{Kind: AllanTauGridExplicit, AveragingFactors: append([]int(nil), factors...)}
}

// GapPolicy selects how missing Allan-series samples are handled.
type GapPolicy uint32

const (
	// GapPolicyReject rejects a series containing missing samples.
	GapPolicyReject GapPolicy = GapPolicy(native.GapPolicyRejectValue)
	// GapPolicyOmitTerms omits estimator terms crossing missing samples.
	GapPolicyOmitTerms GapPolicy = GapPolicy(native.GapPolicyOmitTermsValue)
)

// AllanEstimator identifies one Allan-family deviation estimator.
type AllanEstimator uint32

const (
	// AllanEstimatorADEV selects non-overlapping Allan deviation.
	AllanEstimatorADEV AllanEstimator = AllanEstimator(native.AllanEstimatorADEVValue)
	// AllanEstimatorOverlappingADEV selects overlapping Allan deviation.
	AllanEstimatorOverlappingADEV AllanEstimator = AllanEstimator(native.AllanEstimatorOverlappingADEVValue)
	// AllanEstimatorMDEV selects modified Allan deviation.
	AllanEstimatorMDEV AllanEstimator = AllanEstimator(native.AllanEstimatorMDEVValue)
	// AllanEstimatorHDEV selects Hadamard deviation.
	AllanEstimatorHDEV AllanEstimator = AllanEstimator(native.AllanEstimatorHDEVValue)
	// AllanEstimatorTDEV selects time deviation.
	AllanEstimatorTDEV AllanEstimator = AllanEstimator(native.AllanEstimatorTDEVValue)
)

// AllanEstimatorSet selects the combined curves to compute.
type AllanEstimatorSet struct {
	// ADEV requests non-overlapping Allan deviation.
	ADEV bool
	// OverlappingADEV requests overlapping Allan deviation.
	OverlappingADEV bool
	// MDEV requests modified Allan deviation.
	MDEV bool
	// HDEV requests Hadamard deviation.
	HDEV bool
	// TDEV requests time deviation.
	TDEV bool
}

// AllanEstimatorSetNone selects no combined estimators.
func AllanEstimatorSetNone() AllanEstimatorSet { return AllanEstimatorSet{} }

// AllanEstimatorSetStandard selects all five Allan-family estimators.
func AllanEstimatorSetStandard() AllanEstimatorSet {
	return AllanEstimatorSet{ADEV: true, OverlappingADEV: true, MDEV: true, HDEV: true, TDEV: true}
}

// AllanEstimatorSetAll is an alias for AllanEstimatorSetStandard.
func AllanEstimatorSetAll() AllanEstimatorSet { return AllanEstimatorSetStandard() }

// AllanOptions controls combined Allan-family computations.
type AllanOptions struct {
	// Estimators selects the combined output curves.
	Estimators AllanEstimatorSet
	// TauGrid selects averaging factors; its factors are copied at computation.
	TauGrid TauGrid
	// GapPolicy controls missing-sample handling for gapped series.
	GapPolicy GapPolicy
}

// DefaultAllanOptions returns the C library's default options.
func DefaultAllanOptions() (AllanOptions, error) {
	value, err := native.AllanOptionsDefault()
	return AllanOptions{Estimators: AllanEstimatorSet{ADEV: value.Estimators.ADEV, OverlappingADEV: value.Estimators.OverlappingADEV, MDEV: value.Estimators.MDEV, HDEV: value.Estimators.HDEV, TDEV: value.Estimators.TDEV}, TauGrid: TauGrid{Kind: AllanTauGrid(value.TauGrid)}, GapPolicy: GapPolicy(value.GapPolicy)}, publicError(err)
}

func nativeAllanOptions(value AllanOptions) native.AllanOptions {
	return native.AllanOptions{Estimators: native.AllanEstimatorSet{ADEV: value.Estimators.ADEV, OverlappingADEV: value.Estimators.OverlappingADEV, MDEV: value.Estimators.MDEV, HDEV: value.Estimators.HDEV, TDEV: value.Estimators.TDEV}, TauGrid: uint32(value.TauGrid.Kind), GapPolicy: uint32(value.GapPolicy), AveragingFactors: append([]int(nil), value.TauGrid.AveragingFactors...)}
}

// AllanInput combines a series, sample interval, and estimator options.
type AllanInput struct {
	// Series is clock-stability input with seconds or dimensionless units.
	Series AllanSeries
	// Tau0S is the basic sample interval in seconds.
	Tau0S float64
	// Options controls estimators, grid, and gap policy.
	Options AllanOptions
}

// NewAllanInput combines a series, sample interval, and copied options.
func NewAllanInput(series AllanSeries, tau0S float64, options *AllanOptions) AllanInput {
	value, _ := DefaultAllanOptions()
	if options != nil {
		value = AllanOptions{Estimators: options.Estimators, TauGrid: TauGrid{Kind: options.TauGrid.Kind, AveragingFactors: append([]int(nil), options.TauGrid.AveragingFactors...)}, GapPolicy: options.GapPolicy}
	}
	return AllanInput{Series: series, Tau0S: tau0S, Options: value}
}

type AllanPoint struct {
	// TauS is the averaging time in seconds.
	TauS float64
	// Deviation is in the series' natural units.
	Deviation float64
	// N is the number of contributing terms.
	N int
}

// AllanResult contains the three parallel arrays returned by a C estimator.
// Their lengths are always identical.
type AllanResult struct {
	// TauS contains averaging times in seconds.
	TauS []float64
	// Deviation contains values in the series' natural units.
	Deviation []float64
	// N contains contributing-term counts parallel to TauS and Deviation.
	N []int
}

// Len returns the number of parallel result entries.
func (r AllanResult) Len() int { return len(r.TauS) }

// AllanDeviationCurves owns combined C-computed estimator curves. It is
// non-copyable. Reads use shared locking and may race with Close; Close waits
// for an in-flight read, prevents later use, and is idempotent.
type AllanDeviationCurves struct {
	_      noCopy
	native *native.AllanDeviationCurves
}

// ComputeAllanDeviations computes the selected Allan-family curves.
func ComputeAllanDeviations(input AllanInput) (*AllanDeviationCurves, error) {
	value, err := native.ComputeAllanDeviations(native.AllanInputNative{Series: native.AllanSeries{Kind: uint32(input.Series.kind), Samples: append([]native.AllanSample(nil), input.Series.samples...)}, Tau0S: input.Tau0S, Options: nativeAllanOptions(input.Options)})
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &AllanDeviationCurves{native: value}, nil
}

// Close releases combined curves and is idempotent.
func (c *AllanDeviationCurves) Close() error {
	if c == nil || c.native == nil {
		return nil
	}
	return publicError(c.native.Close())
}

// Present reports whether an estimator curve is available. False with nil
// error is a valid absence result.
func (c *AllanDeviationCurves) Present(estimator AllanEstimator) (bool, error) {
	if c == nil || c.native == nil {
		return false, ErrClosed
	}
	value, err := c.native.Present(uint32(estimator))
	return value, publicError(err)
}

// Curve returns one copied curve and its presence flag atomically with Close.
func (c *AllanDeviationCurves) Curve(estimator AllanEstimator) (AllanResult, bool, error) {
	if c == nil || c.native == nil {
		return AllanResult{}, false, ErrClosed
	}
	values, present, err := c.native.Curve(uint32(estimator))
	if err != nil || !present {
		return AllanResult{}, present, publicError(err)
	}
	out := AllanResult{TauS: make([]float64, len(values)), Deviation: make([]float64, len(values)), N: make([]int, len(values))}
	for i, value := range values {
		out.TauS[i], out.Deviation[i], out.N[i] = value.TauS, value.Deviation, value.N
	}
	return out, true, nil
}

// ADEV returns the combined non-overlapping Allan deviation curve.
func (c *AllanDeviationCurves) ADEV() (AllanResult, bool, error) { return c.Curve(AllanEstimatorADEV) }

// OverlappingADEV returns the combined overlapping Allan deviation curve.
func (c *AllanDeviationCurves) OverlappingADEV() (AllanResult, bool, error) {
	return c.Curve(AllanEstimatorOverlappingADEV)
}

// MDEV returns the combined modified Allan deviation curve.
func (c *AllanDeviationCurves) MDEV() (AllanResult, bool, error) { return c.Curve(AllanEstimatorMDEV) }

// HDEV returns the combined Hadamard deviation curve.
func (c *AllanDeviationCurves) HDEV() (AllanResult, bool, error) { return c.Curve(AllanEstimatorHDEV) }

// TDEV returns the combined time deviation curve.
func (c *AllanDeviationCurves) TDEV() (AllanResult, bool, error) { return c.Curve(AllanEstimatorTDEV) }

func nativeSeries(value AllanSeries) native.AllanSeries {
	return native.AllanSeries{Kind: uint32(value.kind), Samples: append([]native.AllanSample(nil), value.samples...)}
}
func publicAllanResult(values []native.AllanPoint, err error) (AllanResult, error) {
	if err != nil {
		return AllanResult{}, publicError(err)
	}
	out := AllanResult{TauS: make([]float64, len(values)), Deviation: make([]float64, len(values)), N: make([]int, len(values))}
	for i, v := range values {
		out.TauS[i], out.Deviation[i], out.N[i] = v.TauS, v.Deviation, v.N
	}
	return out, nil
}

// AllanDeviation computes non-overlapping Allan deviation at explicit factors.
func AllanDeviation(series AllanSeries, tau0S float64, factors []int) (AllanResult, error) {
	return publicAllanResult(native.AllanDeviation(nativeSeries(series), tau0S, append([]int(nil), factors...)))
}

// OverlappingADEV computes fully overlapping Allan deviation.
func OverlappingADEV(series AllanSeries, tau0S float64, factors []int) (AllanResult, error) {
	return publicAllanResult(native.OverlappingADEV(nativeSeries(series), tau0S, append([]int(nil), factors...)))
}

// ModifiedADEV computes modified Allan deviation.
func ModifiedADEV(series AllanSeries, tau0S float64, factors []int) (AllanResult, error) {
	return publicAllanResult(native.ModifiedADEV(nativeSeries(series), tau0S, append([]int(nil), factors...)))
}

// HadamardDeviation computes Hadamard deviation.
func HadamardDeviation(series AllanSeries, tau0S float64, factors []int) (AllanResult, error) {
	return publicAllanResult(native.HadamardDeviation(nativeSeries(series), tau0S, append([]int(nil), factors...)))
}

// TimeDeviation computes time deviation in the series' natural time units.
func TimeDeviation(series AllanSeries, tau0S float64, factors []int) (AllanResult, error) {
	return publicAllanResult(native.TimeDeviation(nativeSeries(series), tau0S, append([]int(nil), factors...)))
}

// PowerLawNoiseType identifies one IEEE 1139 power-law frequency-noise type.
type PowerLawNoiseType uint32

const (
	// PowerLawRandomWalkFM identifies random-walk frequency modulation.
	PowerLawRandomWalkFM PowerLawNoiseType = PowerLawNoiseType(native.PowerLawRandomWalkFMValue)
	// PowerLawFlickerFM identifies flicker frequency modulation.
	PowerLawFlickerFM PowerLawNoiseType = PowerLawNoiseType(native.PowerLawFlickerFMValue)
	// PowerLawWhiteFM identifies white frequency modulation.
	PowerLawWhiteFM PowerLawNoiseType = PowerLawNoiseType(native.PowerLawWhiteFMValue)
	// PowerLawFlickerPM identifies flicker phase modulation.
	PowerLawFlickerPM PowerLawNoiseType = PowerLawNoiseType(native.PowerLawFlickerPMValue)
	// PowerLawWhitePM identifies white phase modulation.
	PowerLawWhitePM PowerLawNoiseType = PowerLawNoiseType(native.PowerLawWhitePMValue)
)

// PowerLawOctaveFlag describes why an octave could not be fully classified.
type PowerLawOctaveFlag uint32

const (
	// PowerLawOctaveUnderSampled indicates too few tau points.
	PowerLawOctaveUnderSampled PowerLawOctaveFlag = PowerLawOctaveFlag(native.PowerLawOctaveUnderSampledValue)
	// PowerLawOctaveDegenerateDeviation indicates a zero deviation.
	PowerLawOctaveDegenerateDeviation PowerLawOctaveFlag = PowerLawOctaveFlag(native.PowerLawOctaveDegenerateValue)
	// PowerLawOctaveMissingMDEV indicates insufficient MDEV points.
	PowerLawOctaveMissingMDEV PowerLawOctaveFlag = PowerLawOctaveFlag(native.PowerLawOctaveMissingMDEVValue)
)

// PowerLawOctaveDominance identifies an octave classification state.
type PowerLawOctaveDominance uint32

const (
	// PowerLawDominant identifies one classified noise type.
	PowerLawDominant PowerLawOctaveDominance = PowerLawOctaveDominance(native.PowerLawDominantValue)
	// PowerLawAmbiguous identifies conflicting or off-table slopes.
	PowerLawAmbiguous PowerLawOctaveDominance = PowerLawOctaveDominance(native.PowerLawAmbiguousValue)
	// PowerLawFlagged identifies an octave lacking required data.
	PowerLawFlagged PowerLawOctaveDominance = PowerLawOctaveDominance(native.PowerLawFlaggedValue)
)

type PowerLawNoiseOptions struct {
	// MinPointsPerOctave is the minimum tau-point count for classification.
	MinPointsPerOctave int
	// SlopeTolerance is the allowed absolute slope error.
	SlopeTolerance float64
	// ScatterTolerance is the allowed robust slope scatter.
	ScatterTolerance float64
	// BasicTauS is the sample interval in seconds.
	BasicTauS float64
	// MeasurementBandwidthHz is the upper measurement bandwidth in hertz.
	MeasurementBandwidthHz float64
}

// DefaultPowerLawNoiseOptions returns native IEEE 1139 fitting defaults.
func DefaultPowerLawNoiseOptions(basicTauS, bandwidthHz float64) (PowerLawNoiseOptions, error) {
	value, err := native.PowerLawNoiseOptionsDefault(basicTauS, bandwidthHz)
	return PowerLawNoiseOptions{MinPointsPerOctave: value.MinPointsPerOctave, SlopeTolerance: value.SlopeTolerance, ScatterTolerance: value.ScatterTolerance, BasicTauS: value.BasicTauS, MeasurementBandwidthHz: value.MeasurementBandwidthHz}, publicError(err)
}

type PowerLawOctave struct {
	// TauStartS is the first tau in seconds.
	TauStartS float64
	// TauEndS is the last tau in seconds.
	TauEndS float64
	// PointCount is the number of ADEV points used.
	PointCount int
	// HasADEVSlope reports whether ADEVSlope is present.
	HasADEVSlope bool
	// ADEVSlope is the fitted ADEV log-log slope.
	ADEVSlope float64
	// HasMDEVSlope reports whether MDEVSlope is present.
	HasMDEVSlope bool
	// MDEVSlope is the fitted MDEV log-log slope.
	MDEVSlope float64
	// HasSlopeScatter reports whether SlopeScatter is present.
	HasSlopeScatter bool
	// SlopeScatter is robust adjacent-slope scatter.
	SlopeScatter float64
	// Dominance is the classification state.
	Dominance PowerLawOctaveDominance
	// NoiseType is meaningful when Dominance is PowerLawDominant.
	NoiseType PowerLawNoiseType
	// Flag is meaningful when Dominance is PowerLawFlagged.
	Flag PowerLawOctaveFlag
}

// PowerLawNoiseRegion is a consecutive tau span with one fitted coefficient.
type PowerLawNoiseRegion struct {
	// NoiseType is the classified PSD noise type.
	NoiseType PowerLawNoiseType
	// TauStartS is the first tau in seconds.
	TauStartS float64
	// TauEndS is the last tau in seconds.
	TauEndS float64
	// OctaveCount is the number of merged octaves.
	OctaveCount int
	// PointCount is the number of points used by the coefficient fit.
	PointCount int
	// MeanSlope is the mean local log-log slope.
	MeanSlope float64
	// Coefficient is the fitted PSD coefficient.
	Coefficient float64
}

// PowerLawNoiseFit owns a native IEEE 1139 power-law classification. It is
// non-copyable; reads may race with Close, which is synchronized and idempotent.
type PowerLawNoiseFit struct {
	_      noCopy
	native *native.PowerLawNoiseFit
}

// FitPowerLawNoise classifies matching ADEV and MDEV result arrays. Inputs
// and options are copied before the native call.
func FitPowerLawNoise(adev, mdev AllanResult, options *PowerLawNoiseOptions) (*PowerLawNoiseFit, error) {
	if len(adev.TauS) != len(adev.Deviation) || len(adev.TauS) != len(adev.N) {
		return nil, errors.New("sidereon: Allan ADEV result arrays have different lengths")
	}
	if len(mdev.TauS) != len(mdev.Deviation) || len(mdev.TauS) != len(mdev.N) {
		return nil, errors.New("sidereon: Allan MDEV result arrays have different lengths")
	}
	if options != nil && options.MinPointsPerOctave < 0 {
		return nil, errors.New("sidereon: power-law minimum points must not be negative")
	}
	var nativeOptions *native.PowerLawNoiseOptions
	if options != nil {
		nativeOptions = &native.PowerLawNoiseOptions{MinPointsPerOctave: options.MinPointsPerOctave, SlopeTolerance: options.SlopeTolerance, ScatterTolerance: options.ScatterTolerance, BasicTauS: options.BasicTauS, MeasurementBandwidthHz: options.MeasurementBandwidthHz}
	}
	a := make([]native.AllanPoint, len(adev.TauS))
	for i := range a {
		if adev.N[i] < 0 {
			return nil, errors.New("sidereon: Allan ADEV term count must not be negative")
		}
		a[i] = native.AllanPoint{TauS: adev.TauS[i], Deviation: adev.Deviation[i], N: adev.N[i]}
	}
	m := make([]native.AllanPoint, len(mdev.TauS))
	for i := range m {
		if mdev.N[i] < 0 {
			return nil, errors.New("sidereon: Allan MDEV term count must not be negative")
		}
		m[i] = native.AllanPoint{TauS: mdev.TauS[i], Deviation: mdev.Deviation[i], N: mdev.N[i]}
	}
	value, err := native.FitPowerLawNoise(a, m, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &PowerLawNoiseFit{native: value}, nil
}

// Close releases the power-law fit and is idempotent.
func (f *PowerLawNoiseFit) Close() error {
	if f == nil || f.native == nil {
		return nil
	}
	return publicError(f.native.Close())
}

// Coefficients returns five independent PSD coefficients in native order.
func (f *PowerLawNoiseFit) Coefficients() ([5]float64, error) {
	if f == nil || f.native == nil {
		return [5]float64{}, ErrClosed
	}
	value, err := f.native.Coefficients()
	return value, publicError(err)
}

// Octaves returns independent per-octave classifications and presence flags.
func (f *PowerLawNoiseFit) Octaves() ([]PowerLawOctave, error) {
	if f == nil || f.native == nil {
		return nil, ErrClosed
	}
	values, err := f.native.Octaves()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PowerLawOctave, len(values))
	for i, v := range values {
		out[i] = PowerLawOctave{TauStartS: v.TauStartS, TauEndS: v.TauEndS, PointCount: v.PointCount, HasADEVSlope: v.HasADEVSlope, ADEVSlope: v.ADEVSlope, HasMDEVSlope: v.HasMDEVSlope, MDEVSlope: v.MDEVSlope, HasSlopeScatter: v.HasSlopeScatter, SlopeScatter: v.SlopeScatter, Dominance: PowerLawOctaveDominance(v.DominanceKind), NoiseType: PowerLawNoiseType(v.NoiseType), Flag: PowerLawOctaveFlag(v.Flag)}
	}
	return out, nil
}

// Regions returns independent consecutive classified tau regions.
func (f *PowerLawNoiseFit) Regions() ([]PowerLawNoiseRegion, error) {
	if f == nil || f.native == nil {
		return nil, ErrClosed
	}
	values, err := f.native.Regions()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PowerLawNoiseRegion, len(values))
	for i, v := range values {
		out[i] = PowerLawNoiseRegion{NoiseType: PowerLawNoiseType(v.NoiseType), TauStartS: v.TauStartS, TauEndS: v.TauEndS, OctaveCount: v.OctaveCount, PointCount: v.PointCount, MeanSlope: v.MeanSlope, Coefficient: v.Coefficient}
	}
	return out, nil
}

// AllanDeviationPowerLawSlope returns the exact-table ADEV slope.
func AllanDeviationPowerLawSlope(noiseType PowerLawNoiseType) (float64, error) {
	a, _, _, err := native.PowerLawNoiseSlopes(uint32(noiseType))
	return a, publicError(err)
}

// ModifiedAllanDeviationPowerLawSlope returns the exact-table MDEV slope.
func ModifiedAllanDeviationPowerLawSlope(noiseType PowerLawNoiseType) (float64, error) {
	_, m, _, err := native.PowerLawNoiseSlopes(uint32(noiseType))
	return m, publicError(err)
}

// AllanVariancePowerLawTauExponent returns the variance tau exponent.
func AllanVariancePowerLawTauExponent(noiseType PowerLawNoiseType) (int, error) {
	_, _, e, err := native.PowerLawNoiseSlopes(uint32(noiseType))
	return e, publicError(err)
}
