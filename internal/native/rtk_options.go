//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

type RtkMeasurementModel struct {
	CodeSigmaM, PhaseSigmaM float64
	Sagnac                  bool
	Stochastic              uint32
	ElevationWeighting      bool
}

type RtkResidualValidationOptions struct {
	ThresholdSigmaEnabled bool
	ThresholdSigma        float64
	MaxExclusions         int
}

type RtkFloatOptions struct {
	PositionTolM, AmbiguityTolM float64
	MaxIterations               int
}

type RtkFixedOptions struct {
	PositionTolM, AmbiguityTolM float64
	MaxIterations               int
	RatioThreshold              float64
	PartialAmbiguityResolution  bool
	PartialMinAmbiguities       int
}

type RtkArcUpdateOptions struct {
	HoldSigmaM, PositionTolM, AmbiguityTolM float64
	MaxIterations                           int
	ProcessNoiseBaselineSigmaM              float64
	DynamicsVelocityPropagated              bool
	FloatOnlySystems                        []uint32
	ReportResiduals                         bool
	HasARArmingSigmaM                       bool
	ARArmingSigmaM                          float64
	RatioThreshold                          float64
	ReceiverAntenna                         *RtkReceiverAntennaCorrections
}

type RtkRinexSignalPair struct {
	System                          uint32
	CodeObservable, PhaseObservable string
}

type RtkRinexDualSignalPair struct {
	System                                                               uint32
	Code1Observable, Phase1Observable, Code2Observable, Phase2Observable string
}

type RtkRinexArcOptions struct {
	SignalPairs           []RtkRinexSignalPair
	HasMaxEpochs          bool
	MaxEpochs             int
	MinCommonSatellites   int
	IncludePredictionTime bool
}

type RtkRinexDualArcOptions struct {
	SignalPairs           []RtkRinexDualSignalPair
	HasMaxEpochs          bool
	MaxEpochs             int
	MinCommonSatellites   int
	IncludePredictionTime bool
}

func RtkMeasurementModelInit() (RtkMeasurementModel, error) {
	var value C.SidereonRtkMeasurementModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtk_measurement_model_init(&value)) })
	return RtkMeasurementModel{float64(value.code_sigma_m), float64(value.phase_sigma_m), bool(value.sagnac), uint32(value.stochastic), bool(value.elevation_weighting)}, err
}

func RtkResidualValidationOptionsInit() (RtkResidualValidationOptions, error) {
	var value C.SidereonRtkResidualValidationOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtk_residual_validation_options_init(&value)) })
	if err != nil {
		return RtkResidualValidationOptions{}, err
	}
	maxExclusions, err := sizeTToInt(value.max_exclusions, "RTK residual max exclusions")
	return RtkResidualValidationOptions{bool(value.threshold_sigma_enabled), float64(value.threshold_sigma), maxExclusions}, err
}

func RtkFloatOptionsInit() (RtkFloatOptions, error) {
	var value C.SidereonRtkFloatOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtk_float_options_init(&value)) })
	if err != nil {
		return RtkFloatOptions{}, err
	}
	maxIterations, err := sizeTToInt(value.max_iterations, "RTK float max iterations")
	return RtkFloatOptions{float64(value.position_tol_m), float64(value.ambiguity_tol_m), maxIterations}, err
}

func RtkFixedOptionsInit() (RtkFixedOptions, error) {
	var value C.SidereonRtkFixedOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtk_fixed_options_init(&value)) })
	if err != nil {
		return RtkFixedOptions{}, err
	}
	maxIterations, err := sizeTToInt(value.max_iterations, "RTK fixed max iterations")
	if err != nil {
		return RtkFixedOptions{}, err
	}
	partialMin, err := sizeTToInt(value.partial_min_ambiguities, "RTK fixed partial ambiguities")
	return RtkFixedOptions{float64(value.position_tol_m), float64(value.ambiguity_tol_m), maxIterations, float64(value.ratio_threshold), bool(value.partial_ambiguity_resolution), partialMin}, err
}

func RtkArcUpdateOptionsInit() (RtkArcUpdateOptions, error) {
	var value C.SidereonRtkArcUpdateOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtk_arc_update_options_init(&value)) })
	if err != nil {
		return RtkArcUpdateOptions{}, err
	}
	maxIterations, err := sizeTToInt(value.max_iterations, "RTK arc max iterations")
	return RtkArcUpdateOptions{HoldSigmaM: float64(value.hold_sigma_m), PositionTolM: float64(value.position_tol_m), AmbiguityTolM: float64(value.ambiguity_tol_m), MaxIterations: maxIterations, ProcessNoiseBaselineSigmaM: float64(value.process_noise_baseline_sigma_m), DynamicsVelocityPropagated: bool(value.dynamics_velocity_propagated), ReportResiduals: bool(value.report_residuals), HasARArmingSigmaM: bool(value.has_ar_arming_sigma_m), ARArmingSigmaM: float64(value.ar_arming_sigma_m), RatioThreshold: float64(value.ratio_threshold)}, err
}

func RtkRinexArcOptionsInit() (RtkRinexArcOptions, error) {
	var value C.SidereonRtkRinexArcOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtk_rinex_arc_options_init(&value)) })
	if err != nil {
		return RtkRinexArcOptions{}, err
	}
	maxEpochs, err := sizeTToInt(value.max_epochs, "RTK RINEX max epochs")
	if err != nil {
		return RtkRinexArcOptions{}, err
	}
	minCommon, err := sizeTToInt(value.min_common_satellites, "RTK RINEX minimum common satellites")
	return RtkRinexArcOptions{HasMaxEpochs: bool(value.has_max_epochs), MaxEpochs: maxEpochs, MinCommonSatellites: minCommon, IncludePredictionTime: bool(value.include_prediction_time)}, err
}

func RtkRinexDualArcOptionsInit() (RtkRinexDualArcOptions, error) {
	var value C.SidereonRtkRinexDualArcOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtk_rinex_dual_arc_options_init(&value)) })
	if err != nil {
		return RtkRinexDualArcOptions{}, err
	}
	maxEpochs, err := sizeTToInt(value.max_epochs, "RTK dual RINEX max epochs")
	if err != nil {
		return RtkRinexDualArcOptions{}, err
	}
	minCommon, err := sizeTToInt(value.min_common_satellites, "RTK dual RINEX minimum common satellites")
	return RtkRinexDualArcOptions{HasMaxEpochs: bool(value.has_max_epochs), MaxEpochs: maxEpochs, MinCommonSatellites: minCommon, IncludePredictionTime: bool(value.include_prediction_time)}, err
}
