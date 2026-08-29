package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// GeodeticTimeSeriesFrame identifies the coordinate frame of a position series.
type GeodeticTimeSeriesFrame uint32

const (
	GeodeticTimeSeriesFrameENU  GeodeticTimeSeriesFrame = GeodeticTimeSeriesFrame(native.GeodeticTimeSeriesFrameENUValue)
	GeodeticTimeSeriesFrameECEF GeodeticTimeSeriesFrame = GeodeticTimeSeriesFrame(native.GeodeticTimeSeriesFrameECEFValue)
)

// GeodeticTimeSeriesQuality describes the quality of a MIDAS estimate.
type GeodeticTimeSeriesQuality uint32

const (
	GeodeticTimeSeriesQualityNominal   GeodeticTimeSeriesQuality = GeodeticTimeSeriesQuality(native.GeodeticTimeSeriesQualityNominalValue)
	GeodeticTimeSeriesQualityShortSpan GeodeticTimeSeriesQuality = GeodeticTimeSeriesQuality(native.GeodeticTimeSeriesQualityShortSpanValue)
)

// GeodeticTrajectoryLoss selects the robust loss used by trajectory fitting.
type GeodeticTrajectoryLoss uint32

const (
	GeodeticTrajectoryLossLinear GeodeticTrajectoryLoss = GeodeticTrajectoryLoss(native.GeodeticTrajectoryLossLinearValue)
	GeodeticTrajectoryLossSoftL1 GeodeticTrajectoryLoss = GeodeticTrajectoryLoss(native.GeodeticTrajectoryLossSoftL1Value)
	GeodeticTrajectoryLossHuber  GeodeticTrajectoryLoss = GeodeticTrajectoryLoss(native.GeodeticTrajectoryLossHuberValue)
	GeodeticTrajectoryLossCauchy GeodeticTrajectoryLoss = GeodeticTrajectoryLoss(native.GeodeticTrajectoryLossCauchyValue)
	GeodeticTrajectoryLossArctan GeodeticTrajectoryLoss = GeodeticTrajectoryLoss(native.GeodeticTrajectoryLossArctanValue)
)

// GeodeticTrajectoryTermKind identifies a fitted trajectory term.
type GeodeticTrajectoryTermKind uint32

const (
	GeodeticTrajectoryTermPosition      GeodeticTrajectoryTermKind = GeodeticTrajectoryTermKind(native.GeodeticTrajectoryTermPositionValue)
	GeodeticTrajectoryTermVelocity      GeodeticTrajectoryTermKind = GeodeticTrajectoryTermKind(native.GeodeticTrajectoryTermVelocityValue)
	GeodeticTrajectoryTermAnnualSin     GeodeticTrajectoryTermKind = GeodeticTrajectoryTermKind(native.GeodeticTrajectoryTermAnnualSinValue)
	GeodeticTrajectoryTermAnnualCos     GeodeticTrajectoryTermKind = GeodeticTrajectoryTermKind(native.GeodeticTrajectoryTermAnnualCosValue)
	GeodeticTrajectoryTermSemiannualSin GeodeticTrajectoryTermKind = GeodeticTrajectoryTermKind(native.GeodeticTrajectoryTermSemiannualSinValue)
	GeodeticTrajectoryTermSemiannualCos GeodeticTrajectoryTermKind = GeodeticTrajectoryTermKind(native.GeodeticTrajectoryTermSemiannualCosValue)
	GeodeticTrajectoryTermOffset        GeodeticTrajectoryTermKind = GeodeticTrajectoryTermKind(native.GeodeticTrajectoryTermOffsetValue)
)

// GeodeticStepDetectionHeuristic identifies the step-detection method.
type GeodeticStepDetectionHeuristic uint32

const GeodeticStepDetectionHeuristicDetrendedSlidingMedian GeodeticStepDetectionHeuristic = GeodeticStepDetectionHeuristic(native.GeodeticStepDetectionHeuristicDetrendedSlidingMedianValue)

// GeodeticPositionSample is one position and optional covariance observation.
type GeodeticPositionSample struct {
	EpochYear         float64
	PositionM         [3]float64
	CovariancePresent bool
	CovarianceM2      [9]float64
}

// GeodeticPositionSeries is a time-ordered station position series.
type GeodeticPositionSeries struct {
	Frame     GeodeticTimeSeriesFrame
	Reference Geodetic
	Samples   []GeodeticPositionSample
}

// MIDASOptions configures the MIDAS velocity estimator.
type MIDASOptions struct {
	DominantPeriodYears, PeriodToleranceYears float64
	MinPairs                                  int
}

// MIDASComponentStats reports pair selection for one velocity component.
type MIDASComponentStats struct {
	PairCount, RetainedPairCount         int
	SlopeSigmaMPerYr, EffectivePairCount float64
}

// MIDASVelocity is the C-derived velocity estimate and diagnostics.
type MIDASVelocity struct {
	RateENU, SigmaENU     [3]float64
	CovarianceENUM2PerYr2 [9]float64
	ComponentStats        [3]MIDASComponentStats
	SampleCount           int
	SpanYears             float64
	Quality               GeodeticTimeSeriesQuality
}

// GeodeticStepDetectionOptions configures displacement-step detection.
type GeodeticStepDetectionOptions struct {
	WindowYears, ScoreThreshold, MinOffsetM float64
	MinSamplesEachSide                      int
	MinSeparationYears                      float64
	MIDAS                                   MIDASOptions
}

// GeodeticStepCandidate is one detected displacement step.
type GeodeticStepCandidate struct {
	EpochYear               float64
	OffsetENU               [3]float64
	Score                   float64
	BeforeCount, AfterCount int
	Heuristic               GeodeticStepDetectionHeuristic
}

// GeodeticTrajectoryModel describes the terms fitted to a position series.
type GeodeticTrajectoryModel struct {
	ReferenceEpochPresent            bool
	ReferenceEpochYear               float64
	IncludeAnnual, IncludeSemiannual bool
	OffsetEpochsYear                 []float64
}

// GeodeticTrajectoryFitOptions configures trajectory fitting.
type GeodeticTrajectoryFitOptions struct {
	Loss           GeodeticTrajectoryLoss
	FScaleM        float64
	MaxNFEVPresent bool
	MaxNFEV        int
}

// GeodeticTrajectoryComponent contains one ENU component's fit coefficients.
type GeodeticTrajectoryComponent struct {
	PositionM, VelocityMPerYr float64
	AnnualSinPresent          bool
	AnnualSinM                float64
	AnnualCosPresent          bool
	AnnualCosM                float64
	SemiannualSinPresent      bool
	SemiannualSinM            float64
	SemiannualCosPresent      bool
	SemiannualCosM            float64
	OffsetCount               int
}

// GeodeticTrajectorySummary contains trajectory fit diagnostics.
type GeodeticTrajectorySummary struct {
	ReferenceEpochYear       float64
	TermCount, CovarianceDim int
	ResidualRMSENU           [3]float64
	GeometryQuality          GeometryQuality
	Status                   int32
	NFEV, NJEV               int
	Cost, Optimality         float64
}

// GeodeticTrajectoryTerm describes one fitted term.
type GeodeticTrajectoryTerm struct {
	Kind        GeodeticTrajectoryTermKind
	OffsetIndex int
	EpochYear   float64
}

// GeodeticTrajectory owns a C-derived fitted trajectory.
type GeodeticTrajectory struct {
	_      noCopy
	handle *native.GeodeticTrajectory
}

// GeodeticNetworkStation supplies one station to network estimation.
type GeodeticNetworkStation struct {
	ID        string
	Reference Geodetic
	Series    GeodeticPositionSeries
}

// GeodeticNetworkFrame configures the network reference frame.
type GeodeticNetworkFrame struct {
	Origin           Geodetic
	RemoveCommonMode bool
}

// GeodeticStationMotion contains one estimated station motion.
type GeodeticStationMotion struct {
	ID                            string
	RateENU, RawRateENU, SigmaENU [3]float64
	LocalVelocity                 MIDASVelocity
}

// GeodeticMotionField owns a C-derived network motion field.
type GeodeticMotionField struct {
	_      noCopy
	handle *native.GeodeticMotionField
}

func nativePositionSeries(v GeodeticPositionSeries) native.GeodeticPositionSeries {
	out := native.GeodeticPositionSeries{Frame: uint32(v.Frame), Reference: native.Geodetic{LatitudeRad: v.Reference.LatitudeRad, LongitudeRad: v.Reference.LongitudeRad, HeightM: v.Reference.HeightM}, Samples: make([]native.GeodeticPositionSample, len(v.Samples))}
	for i, s := range v.Samples {
		out.Samples[i] = native.GeodeticPositionSample{EpochYear: s.EpochYear, PositionM: s.PositionM, CovariancePresent: s.CovariancePresent, CovarianceM2: s.CovarianceM2}
	}
	return out
}
func nativeMIDAS(v MIDASOptions) native.MidasOptions {
	return native.MidasOptions{DominantPeriodYears: v.DominantPeriodYears, PeriodToleranceYears: v.PeriodToleranceYears, MinPairs: v.MinPairs}
}

// DefaultMIDASOptions returns the native MIDAS defaults.
func DefaultMIDASOptions() (MIDASOptions, error) {
	v, e := native.MidasDefaults()
	return MIDASOptions{DominantPeriodYears: v.DominantPeriodYears, PeriodToleranceYears: v.PeriodToleranceYears, MinPairs: v.MinPairs}, publicError(e)
}

// DefaultGeodeticStepDetectionOptions returns native step-detection defaults.
func DefaultGeodeticStepDetectionOptions() (GeodeticStepDetectionOptions, error) {
	v, e := native.StepDetectionDefaults()
	return GeodeticStepDetectionOptions{WindowYears: v.WindowYears, ScoreThreshold: v.ScoreThreshold, MinOffsetM: v.MinOffsetM, MinSamplesEachSide: v.MinSamplesEachSide, MinSeparationYears: v.MinSeparationYears, MIDAS: MIDASOptions{DominantPeriodYears: v.Midas.DominantPeriodYears, PeriodToleranceYears: v.Midas.PeriodToleranceYears, MinPairs: v.Midas.MinPairs}}, publicError(e)
}

// DefaultGeodeticTrajectoryFitOptions returns native trajectory-fit defaults.
func DefaultGeodeticTrajectoryFitOptions() (GeodeticTrajectoryFitOptions, error) {
	v, e := native.TrajectoryFitDefaults()
	return GeodeticTrajectoryFitOptions{Loss: GeodeticTrajectoryLoss(v.Loss), FScaleM: v.FScaleM, MaxNFEVPresent: v.MaxNFEVPresent, MaxNFEV: v.MaxNFEV}, publicError(e)
}

// VelocityMIDAS estimates station velocity with the native MIDAS solver.
func VelocityMIDAS(v GeodeticPositionSeries, o *MIDASOptions) (MIDASVelocity, error) {
	var no *native.MidasOptions
	if o != nil {
		x := nativeMIDAS(*o)
		no = &x
	}
	x, e := native.VelocityMIDAS(nativePositionSeries(v), no)
	return publicMIDAS(x), publicError(e)
}
func publicMIDAS(v native.MidasVelocity) MIDASVelocity {
	out := MIDASVelocity{SampleCount: v.SampleCount, SpanYears: v.SpanYears, Quality: GeodeticTimeSeriesQuality(v.Quality), RateENU: v.RateENU, SigmaENU: v.SigmaENU, CovarianceENUM2PerYr2: v.CovarianceENUM2PerYr2}
	for i := range out.ComponentStats {
		out.ComponentStats[i] = MIDASComponentStats{PairCount: v.ComponentStats[i].PairCount, RetainedPairCount: v.ComponentStats[i].RetainedPairCount, SlopeSigmaMPerYr: v.ComponentStats[i].SlopeSigmaMPerYr, EffectivePairCount: v.ComponentStats[i].EffectivePairCount}
	}
	return out
}

// DetectGeodeticSteps returns native displacement-step candidates.
func DetectGeodeticSteps(v GeodeticPositionSeries, o *GeodeticStepDetectionOptions) ([]GeodeticStepCandidate, error) {
	var no *native.StepDetectionOptions
	if o != nil {
		x := native.StepDetectionOptions{WindowYears: o.WindowYears, ScoreThreshold: o.ScoreThreshold, MinOffsetM: o.MinOffsetM, MinSamplesEachSide: o.MinSamplesEachSide, MinSeparationYears: o.MinSeparationYears, Midas: nativeMIDAS(o.MIDAS)}
		no = &x
	}
	x, e := native.DetectGeodeticSteps(nativePositionSeries(v), no)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]GeodeticStepCandidate, len(x))
	for i, s := range x {
		out[i] = GeodeticStepCandidate{EpochYear: s.EpochYear, OffsetENU: s.OffsetENU, Score: s.Score, BeforeCount: s.BeforeCount, AfterCount: s.AfterCount, Heuristic: GeodeticStepDetectionHeuristic(s.Heuristic)}
	}
	return out, nil
}

// FitGeodeticTrajectory fits and returns a native trajectory handle.
func FitGeodeticTrajectory(v GeodeticPositionSeries, m GeodeticTrajectoryModel, o *GeodeticTrajectoryFitOptions) (*GeodeticTrajectory, error) {
	nm := native.TrajectoryModel{ReferenceEpochPresent: m.ReferenceEpochPresent, ReferenceEpochYear: m.ReferenceEpochYear, IncludeAnnual: m.IncludeAnnual, IncludeSemiannual: m.IncludeSemiannual, OffsetEpochsYear: append([]float64(nil), m.OffsetEpochsYear...)}
	var no *native.TrajectoryFitOptions
	if o != nil {
		x := native.TrajectoryFitOptions{Loss: uint32(o.Loss), FScaleM: o.FScaleM, MaxNFEVPresent: o.MaxNFEVPresent, MaxNFEV: o.MaxNFEV}
		no = &x
	}
	x, e := native.FitGeodeticTrajectory(nativePositionSeries(v), nm, no)
	if e != nil {
		return nil, publicError(e)
	}
	return &GeodeticTrajectory{handle: x}, nil
}

// Close releases the trajectory; it is safe to call more than once.
func (t *GeodeticTrajectory) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return publicError(t.handle.Close())
}

// Components returns the three fitted ENU components.
func (t *GeodeticTrajectory) Components() ([3]GeodeticTrajectoryComponent, error) {
	if t == nil || t.handle == nil {
		return [3]GeodeticTrajectoryComponent{}, ErrClosed
	}
	x, e := t.handle.Components()
	var out [3]GeodeticTrajectoryComponent
	for i := range out {
		out[i] = GeodeticTrajectoryComponent{PositionM: x[i].PositionM, VelocityMPerYr: x[i].VelocityMPerYr, AnnualSinPresent: x[i].AnnualSinPresent, AnnualSinM: x[i].AnnualSinM, AnnualCosPresent: x[i].AnnualCosPresent, AnnualCosM: x[i].AnnualCosM, SemiannualSinPresent: x[i].SemiannualSinPresent, SemiannualSinM: x[i].SemiannualSinM, SemiannualCosPresent: x[i].SemiannualCosPresent, SemiannualCosM: x[i].SemiannualCosM, OffsetCount: x[i].OffsetCount}
	}
	return out, publicError(e)
}

// Summary returns native trajectory diagnostics.
func (t *GeodeticTrajectory) Summary() (GeodeticTrajectorySummary, error) {
	if t == nil || t.handle == nil {
		return GeodeticTrajectorySummary{}, ErrClosed
	}
	x, e := t.handle.Summary()
	out := GeodeticTrajectorySummary{ReferenceEpochYear: x.ReferenceEpochYear, TermCount: x.TermCount, CovarianceDim: x.CovarianceDim, ResidualRMSENU: x.ResidualRMSENU, Status: x.Status, NFEV: x.NFEV, NJEV: x.NJEV, Cost: x.Cost, Optimality: x.Optimality, GeometryQuality: GeometryQuality{Tier: ObservabilityTier(x.GeometryQuality.Tier), Redundancy: x.GeometryQuality.Redundancy, Rank: x.GeometryQuality.Rank, ConditionNumber: x.GeometryQuality.ConditionNumber, GDOP: x.GeometryQuality.GDOP, RAIMCheckable: x.GeometryQuality.RAIMCheckable, CovarianceValidated: x.GeometryQuality.CovarianceValidated}}
	return out, publicError(e)
}

// Terms returns the native fitted-term list.
func (t *GeodeticTrajectory) Terms() ([]GeodeticTrajectoryTerm, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	x, e := t.handle.Terms()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]GeodeticTrajectoryTerm, len(x))
	for i, v := range x {
		out[i] = GeodeticTrajectoryTerm{Kind: GeodeticTrajectoryTermKind(v.Kind), OffsetIndex: v.OffsetIndex, EpochYear: v.EpochYear}
	}
	return out, nil
}

// Offsets returns one component's fitted offset coefficients.
func (t *GeodeticTrajectory) Offsets(axis int) ([]float64, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	x, e := t.handle.Offsets(axis)
	return x, publicError(e)
}

// ParameterCovariance returns the flattened fitted-parameter covariance.
func (t *GeodeticTrajectory) ParameterCovariance() ([]float64, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	x, e := t.handle.ParameterCovariance()
	return x, publicError(e)
}

// NetworkMotionField estimates station motions with the native solver.
func NetworkMotionField(stations []GeodeticNetworkStation, frame GeodeticNetworkFrame) (*GeodeticMotionField, error) {
	x := make([]native.GeodeticNetworkStation, len(stations))
	for i, s := range stations {
		x[i] = native.GeodeticNetworkStation{ID: s.ID, Reference: native.Geodetic{LatitudeRad: s.Reference.LatitudeRad, LongitudeRad: s.Reference.LongitudeRad, HeightM: s.Reference.HeightM}, Series: nativePositionSeries(s.Series)}
	}
	v, e := native.NetworkField(x, native.GeodeticNetworkFrame{Origin: native.Geodetic{LatitudeRad: frame.Origin.LatitudeRad, LongitudeRad: frame.Origin.LongitudeRad, HeightM: frame.Origin.HeightM}, RemoveCommonMode: frame.RemoveCommonMode})
	if e != nil {
		return nil, publicError(e)
	}
	return &GeodeticMotionField{handle: v}, nil
}

// Close releases the motion field; it is safe to call more than once.
func (m *GeodeticMotionField) Close() error {
	if m == nil || m.handle == nil {
		return nil
	}
	return publicError(m.handle.Close())
}

// CommonMode returns the native common-mode velocity.
func (m *GeodeticMotionField) CommonMode() ([3]float64, error) {
	if m == nil || m.handle == nil {
		return [3]float64{}, ErrClosed
	}
	v, e := m.handle.CommonMode()
	return v, publicError(e)
}

// Stations returns detached station-motion results.
func (m *GeodeticMotionField) Stations() ([]GeodeticStationMotion, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	x, e := m.handle.Stations()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]GeodeticStationMotion, len(x))
	for i, v := range x {
		out[i] = GeodeticStationMotion{ID: v.ID, RateENU: v.RateENU, RawRateENU: v.RawRateENU, SigmaENU: v.SigmaENU, LocalVelocity: publicMIDAS(v.LocalVelocity)}
	}
	return out, nil
}
