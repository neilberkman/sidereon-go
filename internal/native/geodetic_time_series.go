//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type GeodeticPositionSample struct {
	EpochYear         float64
	PositionM         [3]float64
	CovariancePresent bool
	CovarianceM2      [9]float64
}
type GeodeticPositionSeries struct {
	Frame     uint32
	Reference Geodetic
	Samples   []GeodeticPositionSample
}
type MidasOptions struct {
	DominantPeriodYears, PeriodToleranceYears float64
	MinPairs                                  int
}
type MidasComponentStats struct {
	PairCount, RetainedPairCount         int
	SlopeSigmaMPerYr, EffectivePairCount float64
}
type MidasVelocity struct {
	RateENU, SigmaENU     [3]float64
	CovarianceENUM2PerYr2 [9]float64
	ComponentStats        [3]MidasComponentStats
	SampleCount           int
	SpanYears             float64
	Quality               uint32
}
type StepDetectionOptions struct {
	WindowYears, ScoreThreshold, MinOffsetM float64
	MinSamplesEachSide                      int
	MinSeparationYears                      float64
	Midas                                   MidasOptions
}
type StepCandidate struct {
	EpochYear               float64
	OffsetENU               [3]float64
	Score                   float64
	BeforeCount, AfterCount int
	Heuristic               uint32
}
type TrajectoryModel struct {
	ReferenceEpochPresent            bool
	ReferenceEpochYear               float64
	IncludeAnnual, IncludeSemiannual bool
	OffsetEpochsYear                 []float64
}
type TrajectoryFitOptions struct {
	Loss           uint32
	FScaleM        float64
	MaxNFEVPresent bool
	MaxNFEV        int
}
type TrajectoryComponent struct {
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
type TrajectorySummary struct {
	ReferenceEpochYear       float64
	TermCount, CovarianceDim int
	ResidualRMSENU           [3]float64
	GeometryQuality          GeometryQuality
	Status                   int32
	NFEV, NJEV               int
	Cost, Optimality         float64
}
type TrajectoryTerm struct {
	Kind        uint32
	OffsetIndex int
	EpochYear   float64
}
type GeodeticTrajectory struct {
	_      noCopy
	handle *positioningHandle
}
type GeodeticStationMotion struct {
	ID                            string
	RateENU, RawRateENU, SigmaENU [3]float64
	LocalVelocity                 MidasVelocity
}
type GeodeticNetworkStation struct {
	ID        string
	Reference Geodetic
	Series    GeodeticPositionSeries
}
type GeodeticNetworkFrame struct {
	Origin           Geodetic
	RemoveCommonMode bool
}
type GeodeticMotionField struct {
	_      noCopy
	handle *positioningHandle
}

const (
	GeodeticTimeSeriesFrameENUValue                           = uint32(C.SIDEREON_GEODETIC_TIME_SERIES_FRAME_ENU)
	GeodeticTimeSeriesFrameECEFValue                          = uint32(C.SIDEREON_GEODETIC_TIME_SERIES_FRAME_ECEF)
	GeodeticTimeSeriesQualityNominalValue                     = uint32(C.SIDEREON_GEODETIC_TIME_SERIES_QUALITY_NOMINAL)
	GeodeticTimeSeriesQualityShortSpanValue                   = uint32(C.SIDEREON_GEODETIC_TIME_SERIES_QUALITY_SHORT_SPAN)
	GeodeticTrajectoryLossLinearValue                         = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_LOSS_LINEAR)
	GeodeticTrajectoryLossSoftL1Value                         = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_LOSS_SOFT_L1)
	GeodeticTrajectoryLossHuberValue                          = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_LOSS_HUBER)
	GeodeticTrajectoryLossCauchyValue                         = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_LOSS_CAUCHY)
	GeodeticTrajectoryLossArctanValue                         = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_LOSS_ARCTAN)
	GeodeticTrajectoryTermPositionValue                       = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_TERM_KIND_POSITION)
	GeodeticTrajectoryTermVelocityValue                       = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_TERM_KIND_VELOCITY)
	GeodeticTrajectoryTermAnnualSinValue                      = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_TERM_KIND_ANNUAL_SIN)
	GeodeticTrajectoryTermAnnualCosValue                      = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_TERM_KIND_ANNUAL_COS)
	GeodeticTrajectoryTermSemiannualSinValue                  = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_TERM_KIND_SEMIANNUAL_SIN)
	GeodeticTrajectoryTermSemiannualCosValue                  = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_TERM_KIND_SEMIANNUAL_COS)
	GeodeticTrajectoryTermOffsetValue                         = uint32(C.SIDEREON_GEODETIC_TRAJECTORY_TERM_KIND_OFFSET)
	GeodeticStepDetectionHeuristicDetrendedSlidingMedianValue = uint32(C.SIDEREON_GEODETIC_STEP_DETECTION_HEURISTIC_DETRENDED_SLIDING_MEDIAN)
)

func validateGeodeticQuality(value uint32) error {
	if value != GeodeticTimeSeriesQualityNominalValue && value != GeodeticTimeSeriesQualityShortSpanValue {
		return invalidArgument("native geodetic time-series quality is not defined")
	}
	return nil
}

func validateGeodeticStepHeuristic(value uint32) error {
	if value != GeodeticStepDetectionHeuristicDetrendedSlidingMedianValue {
		return invalidArgument("native geodetic step heuristic is not defined")
	}
	return nil
}

func validateGeodeticTrajectoryTerm(value uint32) error {
	switch value {
	case GeodeticTrajectoryTermPositionValue, GeodeticTrajectoryTermVelocityValue,
		GeodeticTrajectoryTermAnnualSinValue, GeodeticTrajectoryTermAnnualCosValue,
		GeodeticTrajectoryTermSemiannualSinValue, GeodeticTrajectoryTermSemiannualCosValue,
		GeodeticTrajectoryTermOffsetValue:
		return nil
	default:
		return invalidArgument("native geodetic trajectory term kind is not defined")
	}
}

func cPositionSeries(s GeodeticPositionSeries) (*C.SidereonGeodeticPositionSeries, unsafe.Pointer, error) {
	if s.Frame != GeodeticTimeSeriesFrameENUValue && s.Frame != GeodeticTimeSeriesFrameECEFValue {
		return nil, nil, invalidArgument("geodetic time-series frame is not defined")
	}
	if len(s.Samples) > int(^uint(0)>>1) {
		return nil, nil, invalidArgument("position series is too large")
	}
	size, e := checkedNativeAllocationSize(len(s.Samples), unsafe.Sizeof(C.SidereonGeodeticPositionSample{}))
	if e != nil {
		return nil, nil, e
	}
	outerSize := unsafe.Sizeof(C.SidereonGeodeticPositionSeries{})
	if size > ^uintptr(0)-outerSize {
		return nil, nil, invalidArgument("position series allocation overflows")
	}
	p := C.malloc(C.size_t(outerSize + size))
	if p == nil {
		return nil, nil, errors.New("sidereon: unable to allocate position series")
	}
	outer := (*C.SidereonGeodeticPositionSeries)(p)
	var samples *C.SidereonGeodeticPositionSample
	if len(s.Samples) > 0 {
		samples = (*C.SidereonGeodeticPositionSample)(unsafe.Add(p, unsafe.Sizeof(C.SidereonGeodeticPositionSeries{})))
		dst := unsafe.Slice(samples, len(s.Samples))
		for j, v := range s.Samples {
			dst[j] = C.SidereonGeodeticPositionSample{epoch_year: C.double(v.EpochYear), has_covariance_m2: C.bool(v.CovariancePresent)}
			for k := 0; k < 3; k++ {
				dst[j].position_m[k] = C.double(v.PositionM[k])
			}
			for k := 0; k < 9; k++ {
				dst[j].covariance_m2[k] = C.double(v.CovarianceM2[k])
			}
		}
	}
	*outer = C.SidereonGeodeticPositionSeries{frame: C.uint32_t(s.Frame), reference: cGeodetic(s.Reference), samples: samples, sample_count: C.size_t(len(s.Samples))}
	return outer, p, nil
}
func freeNativePointers(ps []unsafe.Pointer) {
	for _, p := range ps {
		if p != nil {
			C.free(p)
		}
	}
}
func cMidas(v MidasOptions) (C.SidereonMidasOptions, error) {
	if v.MinPairs < 0 {
		return C.SidereonMidasOptions{}, invalidArgument("negative MIDAS pair count")
	}
	minPairs, err := checkedNativeSize(v.MinPairs)
	if err != nil {
		return C.SidereonMidasOptions{}, err
	}
	return C.SidereonMidasOptions{dominant_period_years: C.double(v.DominantPeriodYears), period_tolerance_years: C.double(v.PeriodToleranceYears), min_pairs: minPairs}, nil
}
func cStep(v StepDetectionOptions) (C.SidereonGeodeticStepDetectionOptions, error) {
	if v.MinSamplesEachSide < 0 {
		return C.SidereonGeodeticStepDetectionOptions{}, invalidArgument("negative step sample count")
	}
	minSamples, err := checkedNativeSize(v.MinSamplesEachSide)
	if err != nil {
		return C.SidereonGeodeticStepDetectionOptions{}, err
	}
	midas, err := cMidas(v.Midas)
	if err != nil {
		return C.SidereonGeodeticStepDetectionOptions{}, err
	}
	return C.SidereonGeodeticStepDetectionOptions{window_years: C.double(v.WindowYears), score_threshold: C.double(v.ScoreThreshold), min_offset_m: C.double(v.MinOffsetM), min_samples_each_side: minSamples, min_separation_years: C.double(v.MinSeparationYears), midas: midas}, nil
}
func cTrajectoryModel(v TrajectoryModel) (*C.SidereonGeodeticTrajectoryModel, unsafe.Pointer, error) {
	size, e := checkedNativeAllocationSize(len(v.OffsetEpochsYear), unsafe.Sizeof(C.double(0)))
	if e != nil {
		return nil, nil, e
	}
	outerSize := unsafe.Sizeof(C.SidereonGeodeticTrajectoryModel{})
	if size > ^uintptr(0)-outerSize {
		return nil, nil, invalidArgument("trajectory model allocation overflows")
	}
	p := C.malloc(C.size_t(outerSize + size))
	if p == nil {
		return nil, nil, errors.New("sidereon: unable to allocate trajectory model")
	}
	outer := (*C.SidereonGeodeticTrajectoryModel)(p)
	var epochs *C.double
	if len(v.OffsetEpochsYear) > 0 {
		epochs = (*C.double)(unsafe.Add(p, unsafe.Sizeof(C.SidereonGeodeticTrajectoryModel{})))
		values := unsafe.Slice(epochs, len(v.OffsetEpochsYear))
		for i, value := range v.OffsetEpochsYear {
			values[i] = C.double(value)
		}
	}
	*outer = C.SidereonGeodeticTrajectoryModel{has_reference_epoch_year: C.bool(v.ReferenceEpochPresent), reference_epoch_year: C.double(v.ReferenceEpochYear), include_annual: C.bool(v.IncludeAnnual), include_semiannual: C.bool(v.IncludeSemiannual), offset_epochs_year: epochs, offset_count: C.size_t(len(v.OffsetEpochsYear))}
	return outer, p, nil
}
func cTrajectoryOptions(v TrajectoryFitOptions) (C.SidereonGeodeticTrajectoryFitOptions, error) {
	if v.Loss > GeodeticTrajectoryLossArctanValue {
		return C.SidereonGeodeticTrajectoryFitOptions{}, invalidArgument("geodetic trajectory loss is not defined")
	}
	if v.MaxNFEV < 0 {
		return C.SidereonGeodeticTrajectoryFitOptions{}, invalidArgument("negative trajectory evaluation count")
	}
	maxNFEV, err := checkedNativeSize(v.MaxNFEV)
	if err != nil {
		return C.SidereonGeodeticTrajectoryFitOptions{}, err
	}
	return C.SidereonGeodeticTrajectoryFitOptions{loss: C.uint32_t(v.Loss), f_scale_m: C.double(v.FScaleM), has_max_nfev: C.bool(v.MaxNFEVPresent), max_nfev: maxNFEV}, nil
}
func releaseTrajectory(p unsafe.Pointer) {
	C.sidereon_geodetic_trajectory_free((*C.SidereonGeodeticTrajectory)(p))
}
func releaseMotionField(p unsafe.Pointer) {
	C.sidereon_geodetic_motion_field_free((*C.SidereonGeodeticMotionField)(p))
}
func MidasDefaults() (MidasOptions, error) {
	var x C.SidereonMidasOptions
	e := callStatus(func() uint32 { return uint32(C.sidereon_geodetic_midas_options_init(&x)) })
	if e != nil {
		return MidasOptions{}, e
	}
	minPairs, e := sizeTToInt(x.min_pairs, "MIDAS default pair count")
	return MidasOptions{DominantPeriodYears: float64(x.dominant_period_years), PeriodToleranceYears: float64(x.period_tolerance_years), MinPairs: minPairs}, e
}
func StepDetectionDefaults() (StepDetectionOptions, error) {
	var x C.SidereonGeodeticStepDetectionOptions
	e := callStatus(func() uint32 { return uint32(C.sidereon_geodetic_step_detection_options_init(&x)) })
	if e != nil {
		return StepDetectionOptions{}, e
	}
	minSamples, e := sizeTToInt(x.min_samples_each_side, "step default sample count")
	if e != nil {
		return StepDetectionOptions{}, e
	}
	minPairs, e := sizeTToInt(x.midas.min_pairs, "step default MIDAS pair count")
	return StepDetectionOptions{WindowYears: float64(x.window_years), ScoreThreshold: float64(x.score_threshold), MinOffsetM: float64(x.min_offset_m), MinSamplesEachSide: minSamples, MinSeparationYears: float64(x.min_separation_years), Midas: MidasOptions{DominantPeriodYears: float64(x.midas.dominant_period_years), PeriodToleranceYears: float64(x.midas.period_tolerance_years), MinPairs: minPairs}}, e
}
func TrajectoryFitDefaults() (TrajectoryFitOptions, error) {
	var x C.SidereonGeodeticTrajectoryFitOptions
	e := callStatus(func() uint32 { return uint32(C.sidereon_geodetic_trajectory_fit_options_init(&x)) })
	if e != nil {
		return TrajectoryFitOptions{}, e
	}
	maxNFEV, e := sizeTToInt(x.max_nfev, "trajectory default evaluation count")
	return TrajectoryFitOptions{Loss: uint32(x.loss), FScaleM: float64(x.f_scale_m), MaxNFEVPresent: bool(x.has_max_nfev), MaxNFEV: maxNFEV}, e
}
func VelocityMIDAS(s GeodeticPositionSeries, o *MidasOptions) (MidasVelocity, error) {
	cs, p, e := cPositionSeries(s)
	if e != nil {
		return MidasVelocity{}, e
	}
	defer func() {
		if p != nil {
			C.free(p)
		}
	}()
	var co C.SidereonMidasOptions
	var cop *C.SidereonMidasOptions
	if o != nil {
		co, e = cMidas(*o)
		if e != nil {
			return MidasVelocity{}, e
		}
		cop = &co
	}
	var x C.SidereonMidasVelocity
	e = callStatus(func() uint32 { return uint32(C.sidereon_geodetic_velocity_midas(cs, cop, &x)) })
	if e != nil {
		return MidasVelocity{}, e
	}
	if e := validateGeodeticQuality(uint32(x.quality)); e != nil {
		return MidasVelocity{}, e
	}
	sampleCount, e := sizeTToInt(x.sample_count, "MIDAS sample count")
	if e != nil {
		return MidasVelocity{}, e
	}
	out := MidasVelocity{SampleCount: sampleCount, SpanYears: float64(x.span_years), Quality: uint32(x.quality)}
	for j := 0; j < 3; j++ {
		out.RateENU[j] = float64(x.rate_enu_m_per_yr[j])
		out.SigmaENU[j] = float64(x.sigma_enu_m_per_yr[j])
		pairCount, err := sizeTToInt(x.component_stats[j].pair_count, "MIDAS pair count")
		if err != nil {
			return MidasVelocity{}, err
		}
		retained, err := sizeTToInt(x.component_stats[j].retained_pair_count, "MIDAS retained pair count")
		if err != nil {
			return MidasVelocity{}, err
		}
		out.ComponentStats[j] = MidasComponentStats{PairCount: pairCount, RetainedPairCount: retained, SlopeSigmaMPerYr: float64(x.component_stats[j].slope_sigma_m_per_yr), EffectivePairCount: float64(x.component_stats[j].effective_pair_count)}
	}
	for j := 0; j < 9; j++ {
		out.CovarianceENUM2PerYr2[j] = float64(x.covariance_enu_m2_per_yr2[j])
	}
	return out, e
}
func DetectGeodeticSteps(s GeodeticPositionSeries, o *StepDetectionOptions) ([]StepCandidate, error) {
	cs, p, e := cPositionSeries(s)
	if e != nil {
		return nil, e
	}
	defer func() {
		if p != nil {
			C.free(p)
		}
	}()
	var co C.SidereonGeodeticStepDetectionOptions
	var cop *C.SidereonGeodeticStepDetectionOptions
	if o != nil {
		co, e = cStep(*o)
		if e != nil {
			return nil, e
		}
		cop = &co
	}
	var raw []C.SidereonGeodeticStepCandidate
	e = copyGeodeticSteps(cs, cop, &raw)
	if e != nil {
		return nil, e
	}
	out := make([]StepCandidate, len(raw))
	for j, v := range raw {
		if err := validateGeodeticStepHeuristic(uint32(v.heuristic)); err != nil {
			return nil, err
		}
		before, err := sizeTToInt(v.before_count, "step candidate before count")
		if err != nil {
			return nil, err
		}
		after, err := sizeTToInt(v.after_count, "step candidate after count")
		if err != nil {
			return nil, err
		}
		out[j] = StepCandidate{EpochYear: float64(v.epoch_year), Score: float64(v.score), BeforeCount: before, AfterCount: after, Heuristic: uint32(v.heuristic)}
		for k := 0; k < 3; k++ {
			out[j].OffsetENU[k] = float64(v.offset_enu_m[k])
		}
	}
	return out, nil
}
func copyGeodeticSteps(cs *C.SidereonGeodeticPositionSeries, o *C.SidereonGeodeticStepDetectionOptions, out *[]C.SidereonGeodeticStepCandidate) error {
	return withCThreadError(func() error {
		var w, r C.size_t
		s := C.sidereon_geodetic_detect_steps(cs, o, nil, 0, &w, &r)
		if e := statusErrorLocked(uint32(s)); e != nil {
			return e
		}
		n, e := validateNativeQuery("geodetic steps", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		memory, e := allocNativeArray(n, unsafe.Sizeof(C.SidereonGeodeticStepCandidate{}))
		if e != nil {
			return e
		}
		defer func() {
			if memory != nil {
				C.free(memory)
			}
		}()
		nativeValues := unsafe.Slice((*C.SidereonGeodeticStepCandidate)(memory), n)
		w, r = 0, 0
		var q *C.SidereonGeodeticStepCandidate
		if n > 0 {
			q = &nativeValues[0]
		}
		s = C.sidereon_geodetic_detect_steps(cs, o, q, C.size_t(n), &w, &r)
		if e = statusErrorLocked(uint32(s)); e != nil {
			return e
		}
		_, e = validateTwoPassCounts("geodetic steps", n, n, uint64(w), uint64(r))
		if e == nil {
			*out = append([]C.SidereonGeodeticStepCandidate(nil), nativeValues...)
		}
		return e
	})
}
func FitGeodeticTrajectory(s GeodeticPositionSeries, m TrajectoryModel, o *TrajectoryFitOptions) (*GeodeticTrajectory, error) {
	cs, p, e := cPositionSeries(s)
	if e != nil {
		return nil, e
	}
	defer func() {
		if p != nil {
			C.free(p)
		}
	}()
	cm, mp, e := cTrajectoryModel(m)
	if e != nil {
		return nil, e
	}
	defer func() {
		if mp != nil {
			C.free(mp)
		}
	}()
	var co C.SidereonGeodeticTrajectoryFitOptions
	var cop *C.SidereonGeodeticTrajectoryFitOptions
	if o != nil {
		co, e = cTrajectoryOptions(*o)
		if e != nil {
			return nil, e
		}
		cop = &co
	}
	var out *C.SidereonGeodeticTrajectory
	e = callStatus(func() uint32 { return uint32(C.sidereon_geodetic_fit_trajectory(cs, cm, cop, &out)) })
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_geodetic_trajectory_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native trajectory returned no handle")
	}
	return &GeodeticTrajectory{handle: newPositioningHandle(unsafe.Pointer(out), releaseTrajectory)}, nil
}
func trajectoryComponent(v C.SidereonGeodeticTrajectoryComponent) (TrajectoryComponent, error) {
	offsetCount, err := sizeTToInt(v.offset_count, "trajectory offset count")
	if err != nil {
		return TrajectoryComponent{}, err
	}
	return TrajectoryComponent{PositionM: float64(v.position_m), VelocityMPerYr: float64(v.velocity_m_per_yr), AnnualSinPresent: bool(v.has_annual_sin_m), AnnualSinM: float64(v.annual_sin_m), AnnualCosPresent: bool(v.has_annual_cos_m), AnnualCosM: float64(v.annual_cos_m), SemiannualSinPresent: bool(v.has_semiannual_sin_m), SemiannualSinM: float64(v.semiannual_sin_m), SemiannualCosPresent: bool(v.has_semiannual_cos_m), SemiannualCosM: float64(v.semiannual_cos_m), OffsetCount: offsetCount}, nil
}
func (t *GeodeticTrajectory) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return t.handle.close()
}
func (t *GeodeticTrajectory) Components() ([3]TrajectoryComponent, error) {
	var raw [3]C.SidereonGeodeticTrajectoryComponent
	if t == nil || t.handle == nil {
		return [3]TrajectoryComponent{}, ErrClosed
	}
	e := t.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_geodetic_trajectory_components((*C.SidereonGeodeticTrajectory)(p), &raw[0]))
		})
	})
	if e != nil {
		return [3]TrajectoryComponent{}, e
	}
	var out [3]TrajectoryComponent
	for j := range out {
		out[j], e = trajectoryComponent(raw[j])
		if e != nil {
			return [3]TrajectoryComponent{}, e
		}
	}
	return out, nil
}
func (t *GeodeticTrajectory) Summary() (TrajectorySummary, error) {
	if t == nil || t.handle == nil {
		return TrajectorySummary{}, ErrClosed
	}
	var x C.SidereonGeodeticTrajectorySummary
	e := t.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_geodetic_trajectory_summary((*C.SidereonGeodeticTrajectory)(p), &x))
		})
	})
	if e != nil {
		return TrajectorySummary{}, e
	}
	termCount, e := sizeTToInt(x.term_count, "trajectory term count")
	if e != nil {
		return TrajectorySummary{}, e
	}
	covarianceDim, e := sizeTToInt(x.covariance_dim, "trajectory covariance dimension")
	if e != nil {
		return TrajectorySummary{}, e
	}
	nfev, e := sizeTToInt(x.nfev, "trajectory function evaluation count")
	if e != nil {
		return TrajectorySummary{}, e
	}
	njev, e := sizeTToInt(x.njev, "trajectory Jacobian evaluation count")
	if e != nil {
		return TrajectorySummary{}, e
	}
	geometry, e := geometryFromC(x.geometry_quality)
	if e != nil {
		return TrajectorySummary{}, e
	}
	out := TrajectorySummary{ReferenceEpochYear: float64(x.reference_epoch_year), TermCount: termCount, CovarianceDim: covarianceDim, GeometryQuality: geometry, Status: int32(x.status), NFEV: nfev, NJEV: njev, Cost: float64(x.cost), Optimality: float64(x.optimality)}
	for j := 0; j < 3; j++ {
		out.ResidualRMSENU[j] = float64(x.residual_rms_enu_m[j])
	}
	return out, nil
}
func (t *GeodeticTrajectory) Terms() ([]TrajectoryTerm, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonGeodeticTrajectoryTerm
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := t.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			s := C.sidereon_geodetic_trajectory_terms((*C.SidereonGeodeticTrajectory)(p), nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("trajectory terms", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.SidereonGeodeticTrajectoryTerm{}))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.SidereonGeodeticTrajectoryTerm)(memory), n)
			w, r = 0, 0
			var q *C.SidereonGeodeticTrajectoryTerm
			if n > 0 {
				q = &raw[0]
			}
			s = C.sidereon_geodetic_trajectory_terms((*C.SidereonGeodeticTrajectory)(p), q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("trajectory terms", n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]TrajectoryTerm, len(raw))
	for j, v := range raw {
		if err := validateGeodeticTrajectoryTerm(uint32(v.kind)); err != nil {
			return nil, err
		}
		offsetIndex, err := sizeTToInt(v.offset_index, "trajectory offset index")
		if err != nil {
			return nil, err
		}
		out[j] = TrajectoryTerm{Kind: uint32(v.kind), OffsetIndex: offsetIndex, EpochYear: float64(v.epoch_year)}
	}
	return out, nil
}
func (t *GeodeticTrajectory) Offsets(axis int) ([]float64, error) {
	if axis < 0 {
		return nil, invalidArgument("negative trajectory axis")
	}
	return copyTrajectoryDoubles(t, "trajectory offsets", func(p *C.SidereonGeodeticTrajectory, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_geodetic_trajectory_offsets(p, C.size_t(axis), o, n, w, r)
	})
}
func (t *GeodeticTrajectory) ParameterCovariance() ([]float64, error) {
	return copyTrajectoryDoubles(t, "trajectory covariance", func(p *C.SidereonGeodeticTrajectory, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_geodetic_trajectory_parameter_covariance(p, o, n, w, r)
	})
}
func copyTrajectoryDoubles(t *GeodeticTrajectory, label string, call func(*C.SidereonGeodeticTrajectory, *C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]float64, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.double
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := t.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			s := call((*C.SidereonGeodeticTrajectory)(p), nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery(label, uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.double(0)))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.double)(memory), n)
			w, r = 0, 0
			var q *C.double
			if n > 0 {
				q = &raw[0]
			}
			s = call((*C.SidereonGeodeticTrajectory)(p), q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts(label, n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]float64, len(raw))
	for j := range raw {
		out[j] = float64(raw[j])
	}
	return out, nil
}
func NetworkField(stations []GeodeticNetworkStation, frame GeodeticNetworkFrame) (*GeodeticMotionField, error) {
	if len(stations) == 0 {
		return nil, invalidArgument("network station list is empty")
	}
	size, e := checkedNativeAllocationSize(len(stations), unsafe.Sizeof(C.SidereonGeodeticNetworkStation{}))
	if e != nil {
		return nil, e
	}
	mem := C.malloc(C.size_t(size))
	if mem == nil {
		return nil, errors.New("sidereon: unable to allocate network stations")
	}
	defer C.free(mem)
	raw := unsafe.Slice((*C.SidereonGeodeticNetworkStation)(mem), len(stations))
	var allocated []unsafe.Pointer
	defer func() { freeNativePointers(allocated) }()
	for j, s := range stations {
		id, e := cString(s.ID)
		if e != nil {
			return nil, e
		}
		allocated = append(allocated, unsafe.Pointer(id))
		series, p, e := cPositionSeries(s.Series)
		if e != nil {
			return nil, e
		}
		if p != nil {
			allocated = append(allocated, p)
		}
		raw[j] = C.SidereonGeodeticNetworkStation{id: id, reference: cGeodetic(s.Reference), series: *series}
	}
	in := C.SidereonGeodeticNetworkFrame{origin: cGeodetic(frame.Origin), remove_common_mode: C.bool(frame.RemoveCommonMode)}
	var out *C.SidereonGeodeticMotionField
	e = callStatus(func() uint32 {
		return uint32(C.sidereon_geodetic_network_field((*C.SidereonGeodeticNetworkStation)(mem), C.size_t(len(stations)), in, &out))
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_geodetic_motion_field_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native motion field returned no handle")
	}
	return &GeodeticMotionField{handle: newPositioningHandle(unsafe.Pointer(out), releaseMotionField)}, nil
}
func (m *GeodeticMotionField) Close() error {
	if m == nil || m.handle == nil {
		return nil
	}
	return m.handle.close()
}
func (m *GeodeticMotionField) CommonMode() ([3]float64, error) {
	if m == nil || m.handle == nil {
		return [3]float64{}, ErrClosed
	}
	var x [3]C.double
	e := m.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_geodetic_motion_field_common_mode((*C.SidereonGeodeticMotionField)(p), &x[0]))
		})
	})
	var out [3]float64
	for j := range out {
		out[j] = float64(x[j])
	}
	return out, e
}
func (m *GeodeticMotionField) Stations() ([]GeodeticStationMotion, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonGeodeticStationMotion
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := m.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			s := C.sidereon_geodetic_motion_field_stations((*C.SidereonGeodeticMotionField)(p), nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("motion field stations", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.SidereonGeodeticStationMotion{}))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.SidereonGeodeticStationMotion)(memory), n)
			w, r = 0, 0
			var q *C.SidereonGeodeticStationMotion
			if n > 0 {
				q = &raw[0]
			}
			s = C.sidereon_geodetic_motion_field_stations((*C.SidereonGeodeticMotionField)(p), q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("motion field stations", n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]GeodeticStationMotion, len(raw))
	for j, v := range raw {
		out[j] = GeodeticStationMotion{ID: fixedCString((*C.char)(unsafe.Pointer(&v.id.bytes[0]))), RateENU: [3]float64{float64(v.rate_enu_m_per_yr[0]), float64(v.rate_enu_m_per_yr[1]), float64(v.rate_enu_m_per_yr[2])}, RawRateENU: [3]float64{float64(v.raw_rate_enu_m_per_yr[0]), float64(v.raw_rate_enu_m_per_yr[1]), float64(v.raw_rate_enu_m_per_yr[2])}, SigmaENU: [3]float64{float64(v.sigma_enu_m_per_yr[0]), float64(v.sigma_enu_m_per_yr[1]), float64(v.sigma_enu_m_per_yr[2])}}
		for k := 0; k < 3; k++ {
			out[j].LocalVelocity.RateENU[k] = float64(v.local_velocity.rate_enu_m_per_yr[k])
			out[j].LocalVelocity.SigmaENU[k] = float64(v.local_velocity.sigma_enu_m_per_yr[k])
		}
		sampleCount, err := sizeTToInt(v.local_velocity.sample_count, "network local MIDAS sample count")
		if err != nil {
			return nil, err
		}
		out[j].LocalVelocity.SampleCount = sampleCount
		out[j].LocalVelocity.SpanYears = float64(v.local_velocity.span_years)
		if err := validateGeodeticQuality(uint32(v.local_velocity.quality)); err != nil {
			return nil, err
		}
		out[j].LocalVelocity.Quality = uint32(v.local_velocity.quality)
	}
	return out, nil
}
