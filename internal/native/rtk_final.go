//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type RtkRinexStaticBaselineConfig struct {
	BaseM                [3]float64
	ArcOptions           RtkRinexArcOptions
	ReferenceMode        uint32
	ReferenceSatellite   string
	ReferencePerSystem   []RtkArcReferenceEntry
	Model                RtkMeasurementModel
	BaselinePriorSigmaM  float64
	AmbiguityPriorSigmaM float64
	InitialBaselineM     [3]float64
	UpdateOptions        RtkArcUpdateOptions
	Preprocessing        RtkArcPreprocessing
	FloatOptions         RtkFloatOptions
	FixedOptions         RtkFixedOptions
	ResidualOptions      RtkResidualValidationOptions
}

type RtkRinexWideLaneFixedConfig struct {
	BaseM                [3]float64
	ArcOptions           RtkRinexDualArcOptions
	ReferenceMode        uint32
	ReferenceSatellite   string
	ReferencePerSystem   []RtkArcReferenceEntry
	Model                RtkMeasurementModel
	BaselinePriorSigmaM  float64
	AmbiguityPriorSigmaM float64
	InitialBaselineM     [3]float64
	UpdateOptions        RtkArcUpdateOptions
	FloatOptions         RtkFloatOptions
	FixedOptions         RtkFixedOptions
	ResidualOptions      RtkResidualValidationOptions
	ApplyTroposphere     bool
}

type RtkStaticArcConfigInput struct {
	Arc             RtkArcConfigInput
	FloatOptions    RtkFloatOptions
	FixedOptions    RtkFixedOptions
	ResidualOptions RtkResidualValidationOptions
}

type RtkStaticArcSolution struct {
	_      noCopy
	handle *surfaceHandle
}

type RtkWideLaneFixedRinexMetadata struct {
	WideLaneFixed          bool
	WideLaneAmbiguityCount int
	DroppedCycleSlipCount  int
	SplitCycleSlipCount    int
}

type RtkWideLaneFixedRinexSolution struct {
	_      noCopy
	handle *surfaceHandle
}

func newRtkStaticArcSolution(pointer *C.SidereonRtkStaticArcSolution) (*RtkStaticArcSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_static_arc_solution_free((*C.SidereonRtkStaticArcSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkStaticArcSolution{handle: handle}, nil
}

func newRtkWideLaneFixedRinexSolution(pointer *C.SidereonRtkWideLaneFixedRinexSolution) (*RtkWideLaneFixedRinexSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_wide_lane_fixed_rinex_solution_free((*C.SidereonRtkWideLaneFixedRinexSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkWideLaneFixedRinexSolution{handle: handle}, nil
}

func rtkMeasurementModelToC(value RtkMeasurementModel) C.SidereonRtkMeasurementModel {
	return C.SidereonRtkMeasurementModel{code_sigma_m: C.double(value.CodeSigmaM), phase_sigma_m: C.double(value.PhaseSigmaM), sagnac: C.bool(value.Sagnac), stochastic: C.uint32_t(value.Stochastic), elevation_weighting: C.bool(value.ElevationWeighting)}
}

func rtkFloatOptionsToC(value RtkFloatOptions) (C.SidereonRtkFloatOptions, error) {
	maxIterations, err := checkedNativeSize(value.MaxIterations)
	if err != nil {
		return C.SidereonRtkFloatOptions{}, fmt.Errorf("sidereon: RTK float max iterations: %w", err)
	}
	return C.SidereonRtkFloatOptions{position_tol_m: C.double(value.PositionTolM), ambiguity_tol_m: C.double(value.AmbiguityTolM), max_iterations: maxIterations}, nil
}

func rtkFixedOptionsToC(value RtkFixedOptions) (C.SidereonRtkFixedOptions, error) {
	maxIterations, err := checkedNativeSize(value.MaxIterations)
	if err != nil {
		return C.SidereonRtkFixedOptions{}, fmt.Errorf("sidereon: RTK fixed max iterations: %w", err)
	}
	partialMin, err := checkedNativeSize(value.PartialMinAmbiguities)
	if err != nil {
		return C.SidereonRtkFixedOptions{}, fmt.Errorf("sidereon: RTK fixed partial ambiguities: %w", err)
	}
	return C.SidereonRtkFixedOptions{position_tol_m: C.double(value.PositionTolM), ambiguity_tol_m: C.double(value.AmbiguityTolM), max_iterations: maxIterations, ratio_threshold: C.double(value.RatioThreshold), partial_ambiguity_resolution: C.bool(value.PartialAmbiguityResolution), partial_min_ambiguities: partialMin}, nil
}

func rtkResidualOptionsToC(value RtkResidualValidationOptions) (C.SidereonRtkResidualValidationOptions, error) {
	maxExclusions, err := checkedNativeSize(value.MaxExclusions)
	if err != nil {
		return C.SidereonRtkResidualValidationOptions{}, fmt.Errorf("sidereon: RTK residual max exclusions: %w", err)
	}
	return C.SidereonRtkResidualValidationOptions{threshold_sigma_enabled: C.bool(value.ThresholdSigmaEnabled), threshold_sigma: C.double(value.ThresholdSigma), max_exclusions: maxExclusions}, nil
}

func copyRtkArcPreprocessing(value RtkArcPreprocessing) (C.SidereonRtkArcPreprocessing, error) {
	if value.CycleSlip > 2 {
		return C.SidereonRtkArcPreprocessing{}, invalidArgument("RTK arc cycle-slip policy is out of range")
	}
	hatchWindow, err := checkedNativeSize(value.HatchWindowCap)
	if err != nil {
		return C.SidereonRtkArcPreprocessing{}, fmt.Errorf("sidereon: RTK arc hatch window cap: %w", err)
	}
	return C.SidereonRtkArcPreprocessing{has_cycle_slip: C.bool(value.HasCycleSlip), cycle_slip: C.uint32_t(value.CycleSlip), has_hatch_window_cap: C.bool(value.HasHatchWindowCap), hatch_window_cap: hatchWindow, has_elevation_mask_deg: C.bool(value.HasElevationMaskDeg), elevation_mask_deg: C.double(value.ElevationMaskDeg)}, nil
}

func copyRtkArcUpdateOptions(value RtkArcUpdateOptions, alloc *cRtkAlloc) (C.SidereonRtkArcUpdateOptions, error) {
	maxIterations, err := checkedNativeSize(value.MaxIterations)
	if err != nil {
		return C.SidereonRtkArcUpdateOptions{}, fmt.Errorf("sidereon: RTK arc max iterations: %w", err)
	}
	floatOnly, floatOnlyCount, err := copyRtkFloatOnlySystemsForArc(value.FloatOnlySystems, alloc)
	if err != nil {
		return C.SidereonRtkArcUpdateOptions{}, err
	}
	antenna, err := copyRtkReceiverAntenna(value.ReceiverAntenna, alloc)
	if err != nil {
		return C.SidereonRtkArcUpdateOptions{}, err
	}
	return C.SidereonRtkArcUpdateOptions{hold_sigma_m: C.double(value.HoldSigmaM), position_tol_m: C.double(value.PositionTolM), ambiguity_tol_m: C.double(value.AmbiguityTolM), max_iterations: maxIterations, process_noise_baseline_sigma_m: C.double(value.ProcessNoiseBaselineSigmaM), dynamics_velocity_propagated: C.bool(value.DynamicsVelocityPropagated), float_only_systems: floatOnly, float_only_system_count: floatOnlyCount, report_residuals: C.bool(value.ReportResiduals), has_ar_arming_sigma_m: C.bool(value.HasARArmingSigmaM), ar_arming_sigma_m: C.double(value.ARArmingSigmaM), ratio_threshold: C.double(value.RatioThreshold), receiver_antenna: antenna}, nil
}

func copyRtkStaticArcConfig(value RtkStaticArcConfigInput, alloc *cRtkAlloc) (*C.SidereonRtkStaticArcConfig, error) {
	arc, err := copyRtkArcConfig(value.Arc, alloc)
	if err != nil {
		return nil, err
	}
	floatOptions, err := rtkFloatOptionsToC(value.FloatOptions)
	if err != nil {
		return nil, err
	}
	fixedOptions, err := rtkFixedOptionsToC(value.FixedOptions)
	if err != nil {
		return nil, err
	}
	residualOptions, err := rtkResidualOptionsToC(value.ResidualOptions)
	if err != nil {
		return nil, err
	}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkStaticArcConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK static arc config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkStaticArcConfig)(memory)
	*config = C.SidereonRtkStaticArcConfig{arc: *arc, float_options: floatOptions, fixed_options: fixedOptions, residual_options: residualOptions}
	return config, nil
}

func copyRtkRinexStaticBaselineConfig(value RtkRinexStaticBaselineConfig, alloc *cRtkAlloc) (*C.SidereonRtkRinexStaticBaselineConfig, error) {
	if value.ReferenceMode > 2 {
		return nil, invalidArgument("RTK RINEX reference mode is out of range")
	}
	if value.Model.Stochastic > 1 {
		return nil, invalidArgument("RTK RINEX measurement stochastic model is out of range")
	}
	arcOptions, err := value.ArcOptions.toC(alloc)
	if err != nil {
		return nil, err
	}
	refSatellite, err := optionalCString(value.ReferenceSatellite, alloc, "RTK RINEX reference satellite ID")
	if err != nil {
		return nil, err
	}
	references, referenceCount, err := copyRtkArcReferences(value.ReferencePerSystem, alloc, "RTK RINEX per-system references")
	if err != nil {
		return nil, err
	}
	updateOptions, err := copyRtkArcUpdateOptions(value.UpdateOptions, alloc)
	if err != nil {
		return nil, err
	}
	preprocessing, err := copyRtkArcPreprocessing(value.Preprocessing)
	if err != nil {
		return nil, err
	}
	floatOptions, err := rtkFloatOptionsToC(value.FloatOptions)
	if err != nil {
		return nil, err
	}
	fixedOptions, err := rtkFixedOptionsToC(value.FixedOptions)
	if err != nil {
		return nil, err
	}
	residualOptions, err := rtkResidualOptionsToC(value.ResidualOptions)
	if err != nil {
		return nil, err
	}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkRinexStaticBaselineConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK RINEX static baseline config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkRinexStaticBaselineConfig)(memory)
	*config = C.SidereonRtkRinexStaticBaselineConfig{base_m: [3]C.double{C.double(value.BaseM[0]), C.double(value.BaseM[1]), C.double(value.BaseM[2])}, arc_options: *arcOptions, reference_mode: C.uint32_t(value.ReferenceMode), reference_satellite: refSatellite, reference_per_system: references, reference_per_system_count: referenceCount, model: rtkMeasurementModelToC(value.Model), baseline_prior_sigma_m: C.double(value.BaselinePriorSigmaM), ambiguity_prior_sigma_m: C.double(value.AmbiguityPriorSigmaM), initial_baseline_m: [3]C.double{C.double(value.InitialBaselineM[0]), C.double(value.InitialBaselineM[1]), C.double(value.InitialBaselineM[2])}, update_options: updateOptions, preprocessing: preprocessing, float_options: floatOptions, fixed_options: fixedOptions, residual_options: residualOptions}
	return config, nil
}

func copyRtkRinexWideLaneFixedConfig(value RtkRinexWideLaneFixedConfig, alloc *cRtkAlloc) (*C.SidereonRtkRinexWideLaneFixedConfig, error) {
	if value.ReferenceMode > 2 {
		return nil, invalidArgument("RTK RINEX reference mode is out of range")
	}
	if value.Model.Stochastic > 1 {
		return nil, invalidArgument("RTK RINEX measurement stochastic model is out of range")
	}
	arcOptions, err := value.ArcOptions.toC(alloc)
	if err != nil {
		return nil, err
	}
	refSatellite, err := optionalCString(value.ReferenceSatellite, alloc, "RTK RINEX reference satellite ID")
	if err != nil {
		return nil, err
	}
	references, referenceCount, err := copyRtkArcReferences(value.ReferencePerSystem, alloc, "RTK RINEX per-system references")
	if err != nil {
		return nil, err
	}
	updateOptions, err := copyRtkArcUpdateOptions(value.UpdateOptions, alloc)
	if err != nil {
		return nil, err
	}
	floatOptions, err := rtkFloatOptionsToC(value.FloatOptions)
	if err != nil {
		return nil, err
	}
	fixedOptions, err := rtkFixedOptionsToC(value.FixedOptions)
	if err != nil {
		return nil, err
	}
	residualOptions, err := rtkResidualOptionsToC(value.ResidualOptions)
	if err != nil {
		return nil, err
	}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkRinexWideLaneFixedConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK RINEX wide-lane fixed config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkRinexWideLaneFixedConfig)(memory)
	*config = C.SidereonRtkRinexWideLaneFixedConfig{base_m: [3]C.double{C.double(value.BaseM[0]), C.double(value.BaseM[1]), C.double(value.BaseM[2])}, arc_options: *arcOptions, reference_mode: C.uint32_t(value.ReferenceMode), reference_satellite: refSatellite, reference_per_system: references, reference_per_system_count: referenceCount, model: rtkMeasurementModelToC(value.Model), baseline_prior_sigma_m: C.double(value.BaselinePriorSigmaM), ambiguity_prior_sigma_m: C.double(value.AmbiguityPriorSigmaM), initial_baseline_m: [3]C.double{C.double(value.InitialBaselineM[0]), C.double(value.InitialBaselineM[1]), C.double(value.InitialBaselineM[2])}, update_options: updateOptions, float_options: floatOptions, fixed_options: fixedOptions, residual_options: residualOptions, apply_troposphere: C.bool(value.ApplyTroposphere)}
	return config, nil
}

func optionalCString(value string, alloc *cRtkAlloc, label string) (*C.char, error) {
	if value == "" {
		return nil, nil
	}
	return alloc.cstring(value, label)
}

func RtkRinexStaticBaselineConfigInit() (RtkRinexStaticBaselineConfig, error) {
	var value C.SidereonRtkRinexStaticBaselineConfig
	var err error
	withCThread(func() { err = statusErrorLocked(uint32(C.sidereon_rtk_rinex_static_baseline_config_init(&value))) })
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	return rtkRinexStaticBaselineConfigFromC(value)
}

func RtkRinexWideLaneFixedConfigInit() (RtkRinexWideLaneFixedConfig, error) {
	var value C.SidereonRtkRinexWideLaneFixedConfig
	var err error
	withCThread(func() { err = statusErrorLocked(uint32(C.sidereon_rtk_rinex_wide_lane_fixed_config_init(&value))) })
	if err != nil {
		return RtkRinexWideLaneFixedConfig{}, err
	}
	return rtkRinexWideLaneFixedConfigFromC(value)
}

func rtkRinexArcOptionsFromC(value C.SidereonRtkRinexArcOptions) (RtkRinexArcOptions, error) {
	maxEpochs, err := sizeTToInt(value.max_epochs, "RTK RINEX max epochs")
	if err != nil {
		return RtkRinexArcOptions{}, err
	}
	minCommon, err := sizeTToInt(value.min_common_satellites, "RTK RINEX minimum common satellites")
	if err != nil {
		return RtkRinexArcOptions{}, err
	}
	result := RtkRinexArcOptions{HasMaxEpochs: bool(value.has_max_epochs), MaxEpochs: maxEpochs, MinCommonSatellites: minCommon, IncludePredictionTime: bool(value.include_prediction_time)}
	count, err := sizeTToInt(value.signal_pair_count, "RTK RINEX signal-pair count")
	if err != nil {
		return RtkRinexArcOptions{}, err
	}
	if count != 0 && value.signal_pairs == nil {
		return RtkRinexArcOptions{}, fmt.Errorf("sidereon: native RTK RINEX signal-pair pointer is nil")
	}
	if count != 0 {
		rows := unsafe.Slice(value.signal_pairs, count)
		result.SignalPairs = make([]RtkRinexSignalPair, count)
		for i := range rows {
			if err := validateGNSSSystemValue(uint32(rows[i].system)); err != nil {
				return RtkRinexArcOptions{}, err
			}
			result.SignalPairs[i] = RtkRinexSignalPair{System: uint32(rows[i].system), CodeObservable: observationFixedString(rows[i].code_observable[:]), PhaseObservable: observationFixedString(rows[i].phase_observable[:])}
		}
	}
	return result, nil
}

func rtkRinexDualArcOptionsFromC(value C.SidereonRtkRinexDualArcOptions) (RtkRinexDualArcOptions, error) {
	maxEpochs, err := sizeTToInt(value.max_epochs, "RTK dual RINEX max epochs")
	if err != nil {
		return RtkRinexDualArcOptions{}, err
	}
	minCommon, err := sizeTToInt(value.min_common_satellites, "RTK dual RINEX minimum common satellites")
	if err != nil {
		return RtkRinexDualArcOptions{}, err
	}
	result := RtkRinexDualArcOptions{HasMaxEpochs: bool(value.has_max_epochs), MaxEpochs: maxEpochs, MinCommonSatellites: minCommon, IncludePredictionTime: bool(value.include_prediction_time)}
	count, err := sizeTToInt(value.signal_pair_count, "RTK dual RINEX signal-pair count")
	if err != nil {
		return RtkRinexDualArcOptions{}, err
	}
	if count != 0 && value.signal_pairs == nil {
		return RtkRinexDualArcOptions{}, fmt.Errorf("sidereon: native RTK dual RINEX signal-pair pointer is nil")
	}
	if count != 0 {
		rows := unsafe.Slice(value.signal_pairs, count)
		result.SignalPairs = make([]RtkRinexDualSignalPair, count)
		for i := range rows {
			if err := validateGNSSSystemValue(uint32(rows[i].system)); err != nil {
				return RtkRinexDualArcOptions{}, err
			}
			result.SignalPairs[i] = RtkRinexDualSignalPair{System: uint32(rows[i].system), Code1Observable: observationFixedString(rows[i].code1_observable[:]), Phase1Observable: observationFixedString(rows[i].phase1_observable[:]), Code2Observable: observationFixedString(rows[i].code2_observable[:]), Phase2Observable: observationFixedString(rows[i].phase2_observable[:])}
		}
	}
	return result, nil
}

func rtkArcUpdateOptionsFromC(value C.SidereonRtkArcUpdateOptions) (RtkArcUpdateOptions, error) {
	maxIterations, err := sizeTToInt(value.max_iterations, "RTK arc max iterations")
	if err != nil {
		return RtkArcUpdateOptions{}, err
	}
	result := RtkArcUpdateOptions{HoldSigmaM: float64(value.hold_sigma_m), PositionTolM: float64(value.position_tol_m), AmbiguityTolM: float64(value.ambiguity_tol_m), MaxIterations: maxIterations, ProcessNoiseBaselineSigmaM: float64(value.process_noise_baseline_sigma_m), DynamicsVelocityPropagated: bool(value.dynamics_velocity_propagated), ReportResiduals: bool(value.report_residuals), HasARArmingSigmaM: bool(value.has_ar_arming_sigma_m), ARArmingSigmaM: float64(value.ar_arming_sigma_m), RatioThreshold: float64(value.ratio_threshold)}
	count, err := sizeTToInt(value.float_only_system_count, "RTK arc float-only system count")
	if err != nil {
		return RtkArcUpdateOptions{}, err
	}
	if count != 0 && value.float_only_systems == nil {
		return RtkArcUpdateOptions{}, fmt.Errorf("sidereon: native RTK float-only system pointer is nil")
	}
	if count != 0 {
		rows := unsafe.Slice(value.float_only_systems, count)
		result.FloatOnlySystems = make([]uint32, count)
		for i := range rows {
			if err := validateGNSSSystemValue(uint32(rows[i])); err != nil {
				return RtkArcUpdateOptions{}, err
			}
			result.FloatOnlySystems[i] = uint32(rows[i])
		}
	}
	if value.receiver_antenna != nil {
		antenna, err := rtkReceiverAntennaFromC(value.receiver_antenna)
		if err != nil {
			return RtkArcUpdateOptions{}, err
		}
		result.ReceiverAntenna = antenna
	}
	return result, nil
}

func rtkReceiverAntennaCalibrationFromC(value C.SidereonReceiverAntennaCalibration) (RtkReceiverAntennaCalibration, error) {
	noAzimuthCount, err := sizeTToInt(value.noazi_pcv_count, "RTK receiver no-azimuth PCV count")
	if err != nil {
		return RtkReceiverAntennaCalibration{}, err
	}
	azimuthCount, err := sizeTToInt(value.azimuth_pcv_count, "RTK receiver azimuth PCV count")
	if err != nil {
		return RtkReceiverAntennaCalibration{}, err
	}
	if noAzimuthCount != 0 && value.noazi_pcv_m == nil {
		return RtkReceiverAntennaCalibration{}, fmt.Errorf("sidereon: native RTK no-azimuth PCV pointer is nil")
	}
	if azimuthCount != 0 && value.azimuth_pcv_m == nil {
		return RtkReceiverAntennaCalibration{}, fmt.Errorf("sidereon: native RTK azimuth PCV pointer is nil")
	}
	result := RtkReceiverAntennaCalibration{PCONEUM: [3]float64{float64(value.pco_neu_m[0]), float64(value.pco_neu_m[1]), float64(value.pco_neu_m[2])}}
	if noAzimuthCount != 0 {
		rows := unsafe.Slice(value.noazi_pcv_m, noAzimuthCount)
		result.NoAzimuthPCV = make([]RtkReceiverAntennaNoAzimuthPCV, noAzimuthCount)
		for i := range rows {
			result.NoAzimuthPCV[i] = RtkReceiverAntennaNoAzimuthPCV{ZenithDeg: float64(rows[i].zenith_deg), ValueM: float64(rows[i].value_m)}
		}
	}
	if azimuthCount != 0 {
		rows := unsafe.Slice(value.azimuth_pcv_m, azimuthCount)
		result.AzimuthPCV = make([]RtkReceiverAntennaAzimuthPCV, azimuthCount)
		for i := range rows {
			result.AzimuthPCV[i] = RtkReceiverAntennaAzimuthPCV{AzimuthDeg: float64(rows[i].azimuth_deg), ZenithDeg: float64(rows[i].zenith_deg), ValueM: float64(rows[i].value_m)}
		}
	}
	return result, nil
}

func rtkReceiverAntennaFromC(value *C.SidereonRtkReceiverAntennaCorrections) (*RtkReceiverAntennaCorrections, error) {
	if value == nil {
		return nil, nil
	}
	base, err := rtkReceiverAntennaCalibrationFromC(value.base)
	if err != nil {
		return nil, err
	}
	rover, err := rtkReceiverAntennaCalibrationFromC(value.rover)
	if err != nil {
		return nil, err
	}
	return &RtkReceiverAntennaCorrections{Base: base, Rover: rover}, nil
}

func rtkReferenceEntriesFromC(value *C.SidereonRtkArcReferenceEntry, count C.size_t) ([]RtkArcReferenceEntry, error) {
	n, err := sizeTToInt(count, "RTK reference count")
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	if value == nil {
		return nil, fmt.Errorf("sidereon: native RTK reference pointer is nil")
	}
	rows := unsafe.Slice(value, n)
	result := make([]RtkArcReferenceEntry, n)
	for i := range rows {
		if err := validateGNSSSystemValue(uint32(rows[i].system)); err != nil {
			return nil, err
		}
		if rows[i].sat_id == nil {
			return nil, fmt.Errorf("sidereon: native RTK reference satellite pointer is nil")
		}
		result[i] = RtkArcReferenceEntry{System: uint32(rows[i].system), SatelliteID: C.GoString(rows[i].sat_id)}
	}
	return result, nil
}

func rtkFloatOptionsFromC(value C.SidereonRtkFloatOptions) (RtkFloatOptions, error) {
	maxIterations, err := sizeTToInt(value.max_iterations, "RTK float max iterations")

	if err != nil {
		return RtkFloatOptions{}, err
	}
	return RtkFloatOptions{PositionTolM: float64(value.position_tol_m), AmbiguityTolM: float64(value.ambiguity_tol_m), MaxIterations: maxIterations}, nil
}

func rtkFixedOptionsFromC(value C.SidereonRtkFixedOptions) (RtkFixedOptions, error) {
	maxIterations, err := sizeTToInt(value.max_iterations, "RTK fixed max iterations")
	if err != nil {
		return RtkFixedOptions{}, err
	}
	partialMin, err := sizeTToInt(value.partial_min_ambiguities, "RTK fixed partial ambiguities")
	if err != nil {
		return RtkFixedOptions{}, err
	}
	return RtkFixedOptions{PositionTolM: float64(value.position_tol_m), AmbiguityTolM: float64(value.ambiguity_tol_m), MaxIterations: maxIterations, RatioThreshold: float64(value.ratio_threshold), PartialAmbiguityResolution: bool(value.partial_ambiguity_resolution), PartialMinAmbiguities: partialMin}, nil
}

func rtkResidualOptionsFromC(value C.SidereonRtkResidualValidationOptions) (RtkResidualValidationOptions, error) {
	maxExclusions, err := sizeTToInt(value.max_exclusions, "RTK residual max exclusions")
	if err != nil {
		return RtkResidualValidationOptions{}, err
	}
	return RtkResidualValidationOptions{ThresholdSigmaEnabled: bool(value.threshold_sigma_enabled), ThresholdSigma: float64(value.threshold_sigma), MaxExclusions: maxExclusions}, nil
}

func copyRtkPreprocessingFromC(value C.SidereonRtkArcPreprocessing) (RtkArcPreprocessing, error) {
	if value.cycle_slip > 2 {
		return RtkArcPreprocessing{}, invalidArgument("native RTK cycle-slip policy is out of range")
	}
	hatchWindow, err := sizeTToInt(value.hatch_window_cap, "RTK arc hatch window cap")
	if err != nil {
		return RtkArcPreprocessing{}, err
	}
	return RtkArcPreprocessing{HasCycleSlip: bool(value.has_cycle_slip), CycleSlip: uint32(value.cycle_slip), HasHatchWindowCap: bool(value.has_hatch_window_cap), HatchWindowCap: hatchWindow, HasElevationMaskDeg: bool(value.has_elevation_mask_deg), ElevationMaskDeg: float64(value.elevation_mask_deg)}, nil
}

func rtkRinexStaticBaselineConfigFromC(value C.SidereonRtkRinexStaticBaselineConfig) (RtkRinexStaticBaselineConfig, error) {
	if value.reference_mode > 2 {
		return RtkRinexStaticBaselineConfig{}, invalidArgument("native RTK RINEX reference mode is out of range")
	}
	arcOptions, err := rtkRinexArcOptionsFromC(value.arc_options)
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	updateOptions, err := rtkArcUpdateOptionsFromC(value.update_options)
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	preprocessing, err := copyRtkPreprocessingFromC(value.preprocessing)
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	floatOptions, err := rtkFloatOptionsFromC(value.float_options)
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	fixedOptions, err := rtkFixedOptionsFromC(value.fixed_options)
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	residualOptions, err := rtkResidualOptionsFromC(value.residual_options)
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	references, err := rtkReferenceEntriesFromC(value.reference_per_system, value.reference_per_system_count)
	if err != nil {
		return RtkRinexStaticBaselineConfig{}, err
	}
	result := RtkRinexStaticBaselineConfig{BaseM: [3]float64{float64(value.base_m[0]), float64(value.base_m[1]), float64(value.base_m[2])}, ArcOptions: arcOptions, ReferenceMode: uint32(value.reference_mode), ReferenceSatellite: cStringOrEmpty(value.reference_satellite), Model: RtkMeasurementModel{CodeSigmaM: float64(value.model.code_sigma_m), PhaseSigmaM: float64(value.model.phase_sigma_m), Sagnac: bool(value.model.sagnac), Stochastic: uint32(value.model.stochastic), ElevationWeighting: bool(value.model.elevation_weighting)}, BaselinePriorSigmaM: float64(value.baseline_prior_sigma_m), AmbiguityPriorSigmaM: float64(value.ambiguity_prior_sigma_m), InitialBaselineM: [3]float64{float64(value.initial_baseline_m[0]), float64(value.initial_baseline_m[1]), float64(value.initial_baseline_m[2])}, UpdateOptions: updateOptions, Preprocessing: preprocessing, FloatOptions: floatOptions, FixedOptions: fixedOptions, ResidualOptions: residualOptions}
	result.ReferencePerSystem = references
	return result, nil
}

func rtkRinexWideLaneFixedConfigFromC(value C.SidereonRtkRinexWideLaneFixedConfig) (RtkRinexWideLaneFixedConfig, error) {
	if value.reference_mode > 2 {
		return RtkRinexWideLaneFixedConfig{}, invalidArgument("native RTK RINEX reference mode is out of range")
	}
	arcOptions, err := rtkRinexDualArcOptionsFromC(value.arc_options)
	if err != nil {
		return RtkRinexWideLaneFixedConfig{}, err
	}
	updateOptions, err := rtkArcUpdateOptionsFromC(value.update_options)
	if err != nil {
		return RtkRinexWideLaneFixedConfig{}, err
	}
	floatOptions, err := rtkFloatOptionsFromC(value.float_options)
	if err != nil {
		return RtkRinexWideLaneFixedConfig{}, err
	}
	fixedOptions, err := rtkFixedOptionsFromC(value.fixed_options)
	if err != nil {
		return RtkRinexWideLaneFixedConfig{}, err
	}
	residualOptions, err := rtkResidualOptionsFromC(value.residual_options)
	if err != nil {
		return RtkRinexWideLaneFixedConfig{}, err
	}
	references, err := rtkReferenceEntriesFromC(value.reference_per_system, value.reference_per_system_count)
	if err != nil {
		return RtkRinexWideLaneFixedConfig{}, err
	}
	result := RtkRinexWideLaneFixedConfig{BaseM: [3]float64{float64(value.base_m[0]), float64(value.base_m[1]), float64(value.base_m[2])}, ArcOptions: arcOptions, ReferenceMode: uint32(value.reference_mode), ReferenceSatellite: cStringOrEmpty(value.reference_satellite), Model: RtkMeasurementModel{CodeSigmaM: float64(value.model.code_sigma_m), PhaseSigmaM: float64(value.model.phase_sigma_m), Sagnac: bool(value.model.sagnac), Stochastic: uint32(value.model.stochastic), ElevationWeighting: bool(value.model.elevation_weighting)}, BaselinePriorSigmaM: float64(value.baseline_prior_sigma_m), AmbiguityPriorSigmaM: float64(value.ambiguity_prior_sigma_m), InitialBaselineM: [3]float64{float64(value.initial_baseline_m[0]), float64(value.initial_baseline_m[1]), float64(value.initial_baseline_m[2])}, UpdateOptions: updateOptions, FloatOptions: floatOptions, FixedOptions: fixedOptions, ResidualOptions: residualOptions, ApplyTroposphere: bool(value.apply_troposphere)}
	result.ReferencePerSystem = references
	return result, nil
}

func cStringOrEmpty(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}

func SolveStaticRtkArc(epochs []RtkArcEpochInput, config RtkStaticArcConfigInput) (*RtkStaticArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cEpochs, epochCount, err := copyRtkArcEpochInputs(epochs, alloc)
	if err != nil {
		return nil, err
	}
	cConfig, err := copyRtkStaticArcConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkStaticArcSolution
	var operationErr error
	withCThread(func() {
		status := C.sidereon_solve_static_rtk_arc(cEpochs, epochCount, cConfig, &pointer)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && pointer != nil {
			C.sidereon_rtk_static_arc_solution_free(pointer)
			pointer = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkStaticArcSolution(pointer)
}

func SolveStaticRinexRtkBaseline(sp3 *SP3, base, rover *RinexObs, config RtkRinexStaticBaselineConfig) (*RtkStaticArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := copyRtkRinexStaticBaselineConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkStaticArcSolution
	var operationErr error
	err = withRtkRinexInputs(sp3, base, rover, func(sp3Pointer, basePointer, roverPointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_solve_static_rinex_rtk_baseline((*C.SidereonSp3)(sp3Pointer), (*C.SidereonRinexObs)(basePointer), (*C.SidereonRinexObs)(roverPointer), cConfig, &pointer)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && pointer != nil {
				C.sidereon_rtk_static_arc_solution_free(pointer)
				pointer = nil
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkStaticArcSolution(pointer)
}

func SolveWideLaneFixedRinexRtkBaseline(sp3 *SP3, base, rover *RinexObs, config RtkRinexWideLaneFixedConfig) (*RtkWideLaneFixedRinexSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := copyRtkRinexWideLaneFixedConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkWideLaneFixedRinexSolution
	var operationErr error
	err = withRtkRinexInputs(sp3, base, rover, func(sp3Pointer, basePointer, roverPointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_solve_wide_lane_fixed_rinex_rtk_baseline((*C.SidereonSp3)(sp3Pointer), (*C.SidereonRinexObs)(basePointer), (*C.SidereonRinexObs)(roverPointer), cConfig, &pointer)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && pointer != nil {
				C.sidereon_rtk_wide_lane_fixed_rinex_solution_free(pointer)
				pointer = nil
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkWideLaneFixedRinexSolution(pointer)
}

func (s *RtkStaticArcSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *RtkWideLaneFixedRinexSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

type rtkStaticAmbiguitySatelliteOutputCall func(*C.SidereonRtkAmbiguitySatelliteOut, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkStaticAmbiguitySatellites(label string, call rtkStaticAmbiguitySatelliteOutputCall) ([]RtkAmbiguitySatellite, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkAmbiguitySatelliteOut{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkAmbiguitySatelliteOut, count)
	var output *C.SidereonRtkAmbiguitySatelliteOut
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkAmbiguitySatellite, writtenCount)
	for i := range result {
		result[i] = RtkAmbiguitySatellite{ID: rtkIDFromC(rows[i].id).Value, SatelliteID: tokenFromC(rows[i].sat_id)}
	}
	return result, nil
}

type rtkFixedAmbiguityOutputCall func(*C.SidereonRtkFixedAmbiguity, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkFixedAmbiguitiesOutput(label string, call rtkFixedAmbiguityOutputCall) ([]RtkFixedAmbiguity, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkFixedAmbiguity{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkFixedAmbiguity, count)
	var output *C.SidereonRtkFixedAmbiguity
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkFixedAmbiguity, writtenCount)
	for i := range result {
		result[i] = RtkFixedAmbiguity{ID: rtkIDFromC(rows[i].id).Value, Cycles: int64(rows[i].cycles), ValueM: float64(rows[i].value_m)}
	}
	return result, nil
}

type rtkStaticAmbiguityOutputCall func(*C.SidereonRtkAmbiguity, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkStaticAmbiguities(label string, call rtkStaticAmbiguityOutputCall) ([]RtkAmbiguity, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkAmbiguity{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkAmbiguity, count)
	var output *C.SidereonRtkAmbiguity
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkAmbiguity, writtenCount)
	for i := range result {
		result[i] = RtkAmbiguity{ID: rtkIDFromC(rows[i].id).Value, ValueM: float64(rows[i].value_m)}
	}
	return result, nil
}

func (s *RtkStaticArcSolution) baseline(call func(*C.SidereonRtkStaticArcSolution, *C.double, C.size_t) C.enum_SidereonStatus) ([3]float64, error) {
	var result [3]float64
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	var values [3]C.double
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(call((*C.SidereonRtkStaticArcSolution)(pointer), &values[0], 3)) })
	})
	if err != nil {
		return result, err
	}
	for i := range result {
		result[i] = float64(values[i])
	}
	return result, nil
}

func (s *RtkStaticArcSolution) AmbiguityIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []string
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkIDs("RTK static arc ambiguity IDs", func(out *C.SidereonRtkId, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_static_arc_solution_ambiguity_ids((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkStaticArcSolution) AmbiguitySatellites() ([]RtkAmbiguitySatellite, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkAmbiguitySatellite
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkStaticAmbiguitySatellites("RTK static arc ambiguity satellites", func(out *C.SidereonRtkAmbiguitySatelliteOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_static_arc_solution_ambiguity_satellites((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkStaticArcSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return copySurfaceTokens(s.handle, "RTK static arc dropped satellites", func(pointer unsafe.Pointer, out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtk_static_arc_solution_dropped_sats((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
	})
}

func (s *RtkStaticArcSolution) ElevationMaskedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return copySurfaceTokens(s.handle, "RTK static arc elevation-masked satellites", func(pointer unsafe.Pointer, out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtk_static_arc_solution_elevation_masked_sats((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
	})
}

func (s *RtkStaticArcSolution) FixedAmbiguities() ([]RtkFixedAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkFixedAmbiguity
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkFixedAmbiguitiesOutput("RTK static arc fixed ambiguities", func(out *C.SidereonRtkFixedAmbiguity, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_static_arc_solution_fixed_ambiguities((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkStaticArcSolution) FixedBaselineECEF() ([3]float64, error) {
	return s.baseline(func(solution *C.SidereonRtkStaticArcSolution, out *C.double, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtk_static_arc_solution_fixed_baseline_ecef(solution, out, length)
	})
}

func (s *RtkStaticArcSolution) FixedFreeAmbiguities() ([]RtkAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkAmbiguity
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkStaticAmbiguities("RTK static arc fixed free ambiguities", func(out *C.SidereonRtkAmbiguity, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_static_arc_solution_fixed_free_ambiguities((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func rtkFixedMetadataFromC(value C.SidereonRtkFixedMetadata) (RtkFixedMetadata, error) {
	iterations, err := sizeTToInt(value.iterations, "RTK static fixed metadata iterations")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	nObservations, err := sizeTToInt(value.n_observations, "RTK static fixed metadata observations")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	freeAmbiguities, err := sizeTToInt(value.free_ambiguity_count, "RTK static fixed metadata free ambiguities")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	fixedAmbiguities, err := sizeTToInt(value.fixed_ambiguity_count, "RTK static fixed metadata fixed ambiguities")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	residuals, err := sizeTToInt(value.residual_count, "RTK static fixed metadata residuals")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	usedSatellites, err := sizeTToInt(value.used_sat_count, "RTK static fixed metadata satellites")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	candidates, err := sizeTToInt(value.integer_candidates, "RTK static fixed metadata integer candidates")
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	if value.status != C.SIDEREON_RTK_SOLVE_STATUS_STATE_TOLERANCE && value.status != C.SIDEREON_RTK_SOLVE_STATUS_MAX_ITERATIONS {
		return RtkFixedMetadata{}, fmt.Errorf("sidereon: native RTK static fixed solve status %d is invalid", uint32(value.status))
	}
	if value.integer_status != C.SIDEREON_RTK_INTEGER_STATUS_FIXED && value.integer_status != C.SIDEREON_RTK_INTEGER_STATUS_NOT_FIXED {
		return RtkFixedMetadata{}, fmt.Errorf("sidereon: native RTK static integer status %d is invalid", uint32(value.integer_status))
	}
	geometry, err := geometryFromC(value.geometry_quality)
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	return RtkFixedMetadata{Iterations: iterations, NObservations: nObservations, FreeAmbiguityCount: freeAmbiguities, FixedAmbiguityCount: fixedAmbiguities, ResidualCount: residuals, UsedSatCount: usedSatellites, Converged: bool(value.converged), Status: uint32(value.status), IntegerStatus: uint32(value.integer_status), CodeRMSM: float64(value.code_rms_m), PhaseRMSM: float64(value.phase_rms_m), WeightedRMSM: float64(value.weighted_rms_m), HasIntegerRatio: bool(value.has_integer_ratio), IntegerRatio: float64(value.integer_ratio), HasIntegerBestScore: bool(value.has_integer_best_score), IntegerBestScore: float64(value.integer_best_score), HasIntegerSecondBestScore: bool(value.has_integer_second_best_score), IntegerSecondBestScore: float64(value.integer_second_best_score), IntegerCandidates: candidates, GeometryQuality: geometry}, nil
}

func rtkFloatMetadataFromC(value C.SidereonRtkFloatMetadata) (RtkFloatMetadata, error) {
	if value.status != C.SIDEREON_RTK_SOLVE_STATUS_STATE_TOLERANCE && value.status != C.SIDEREON_RTK_SOLVE_STATUS_MAX_ITERATIONS {
		return RtkFloatMetadata{}, fmt.Errorf("sidereon: native RTK static float solve status %d is invalid", uint32(value.status))
	}
	iterations, err := sizeTToInt(value.iterations, "RTK static float metadata iterations")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	nObservations, err := sizeTToInt(value.n_observations, "RTK static float metadata observations")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	ambiguities, err := sizeTToInt(value.ambiguity_count, "RTK static float metadata ambiguities")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	residuals, err := sizeTToInt(value.residual_count, "RTK static float metadata residuals")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	usedSatellites, err := sizeTToInt(value.used_sat_count, "RTK static float metadata satellites")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	geometry, err := geometryFromC(value.geometry_quality)
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	return RtkFloatMetadata{Iterations: iterations, NObservations: nObservations, AmbiguityCount: ambiguities, ResidualCount: residuals, UsedSatCount: usedSatellites, Converged: bool(value.converged), Status: uint32(value.status), CodeRMSM: float64(value.code_rms_m), PhaseRMSM: float64(value.phase_rms_m), WeightedRMSM: float64(value.weighted_rms_m), GeometryQuality: geometry}, nil
}

func (s *RtkStaticArcSolution) FixedMetadata() (RtkFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkFixedMetadata{}, ErrClosed
	}
	var value C.SidereonRtkFixedMetadata
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_static_arc_solution_fixed_metadata((*C.SidereonRtkStaticArcSolution)(pointer), &value))
		})
	})
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	return rtkFixedMetadataFromC(value)
}

func (s *RtkStaticArcSolution) FloatAmbiguities() ([]RtkAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkAmbiguity
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkStaticAmbiguities("RTK static arc float ambiguities", func(out *C.SidereonRtkAmbiguity, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_static_arc_solution_float_ambiguities((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkStaticArcSolution) FloatBaselineECEF() ([3]float64, error) {
	return s.baseline(func(solution *C.SidereonRtkStaticArcSolution, out *C.double, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtk_static_arc_solution_float_baseline_ecef(solution, out, length)
	})
}

func (s *RtkStaticArcSolution) FloatMetadata() (RtkFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkFloatMetadata{}, ErrClosed
	}
	var value C.SidereonRtkFloatMetadata
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_static_arc_solution_float_metadata((*C.SidereonRtkStaticArcSolution)(pointer), &value))
		})
	})
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	return rtkFloatMetadataFromC(value)
}

func (s *RtkStaticArcSolution) GeometryQuality() (GeometryQuality, error) {
	if s == nil || s.handle == nil {
		return GeometryQuality{}, ErrClosed
	}
	var value C.SidereonGeometryQuality
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_static_arc_solution_geometry_quality((*C.SidereonRtkStaticArcSolution)(pointer), &value))
		})
	})
	if err != nil {
		return GeometryQuality{}, err
	}
	return geometryFromC(value)
}

func (s *RtkStaticArcSolution) References() ([]RtkArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkArcReference
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkReferences("RTK static arc references", func(out *C.SidereonRtkArcReferenceOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_static_arc_solution_references((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkStaticArcSolution) SplitCycleSlipArcs() ([]RtkArcSplitArc, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkArcSplitArc
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkSplitArcs("RTK static arc split cycle-slip arcs", func(out *C.SidereonRtkArcSplitArc, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_static_arc_solution_split_cycle_slip_arcs((*C.SidereonRtkStaticArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkWideLaneArcSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return copySurfaceTokens(s.handle, "RTK wide-lane dropped satellites", func(pointer unsafe.Pointer, out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtk_wide_lane_arc_solution_dropped_sats((*C.SidereonRtkWideLaneArcSolution)(pointer), out, length, written, required)
	})
}

func (s *RtkWideLaneArcSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_wide_lane_arc_solution_epoch_count((*C.SidereonRtkWideLaneArcSolution)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK wide-lane solution epoch count")
}

func (s *RtkWideLaneArcSolution) GeometryQuality() (GeometryQuality, error) {
	if s == nil || s.handle == nil {
		return GeometryQuality{}, ErrClosed
	}
	var value C.SidereonGeometryQuality
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_wide_lane_arc_solution_geometry_quality((*C.SidereonRtkWideLaneArcSolution)(pointer), &value))
		})
	})
	if err != nil {
		return GeometryQuality{}, err
	}
	return geometryFromC(value)
}

func (s *RtkWideLaneArcSolution) References() ([]RtkArcReference, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkArcReference
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkReferences("RTK wide-lane references", func(out *C.SidereonRtkArcReferenceOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_wide_lane_arc_solution_references((*C.SidereonRtkWideLaneArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkWideLaneArcSolution) SplitCycleSlipArcs() ([]RtkArcSplitArc, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkArcSplitArc
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkSplitArcs("RTK wide-lane split cycle-slip arcs", func(out *C.SidereonRtkArcSplitArc, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_wide_lane_arc_solution_split_cycle_slip_arcs((*C.SidereonRtkWideLaneArcSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (s *RtkWideLaneArcSolution) WideLaneCycles() ([]RtkWideLaneCycleInput, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkWideLaneCycleInput
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var written, required C.size_t
			status := C.sidereon_rtk_wide_lane_arc_solution_wide_lane_cycles((*C.SidereonRtkWideLaneArcSolution)(pointer), nil, 0, &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			count, e := validateNativeQuery("RTK wide-lane cycles", uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			if _, e = checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkWideLaneCycle{})); e != nil {
				operationErr = e
				return
			}
			rows := make([]C.SidereonRtkWideLaneCycle, count)
			var output *C.SidereonRtkWideLaneCycle
			if count != 0 {
				output = &rows[0]
			}
			written, required = 0, 0
			status = C.sidereon_rtk_wide_lane_arc_solution_wide_lane_cycles((*C.SidereonRtkWideLaneArcSolution)(pointer), output, C.size_t(count), &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			writtenCount, e := validateTwoPassCounts("RTK wide-lane cycles", count, count, uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			result = make([]RtkWideLaneCycleInput, writtenCount)
			for i := range result {
				result[i] = RtkWideLaneCycleInput{ID: rtkIDFromC(rows[i].id).Value, Cycles: int64(rows[i].cycles)}
			}
		})
		return operationErr
	})
	return result, err
}

func (s *RtkWideLaneFixedRinexSolution) baseline(call func(*C.SidereonRtkWideLaneFixedRinexSolution, *C.double, C.size_t) C.enum_SidereonStatus) ([3]float64, error) {
	var result [3]float64
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	var values [3]C.double
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(call((*C.SidereonRtkWideLaneFixedRinexSolution)(pointer), &values[0], 3)) })
	})
	if err != nil {
		return result, err
	}
	for i := range result {
		result[i] = float64(values[i])
	}
	return result, nil
}

func (s *RtkWideLaneFixedRinexSolution) FixedBaselineECEF() ([3]float64, error) {
	return s.baseline(func(solution *C.SidereonRtkWideLaneFixedRinexSolution, out *C.double, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtk_wide_lane_fixed_rinex_solution_fixed_baseline_ecef(solution, out, length)
	})
}

func (s *RtkWideLaneFixedRinexSolution) FloatBaselineECEF() ([3]float64, error) {
	return s.baseline(func(solution *C.SidereonRtkWideLaneFixedRinexSolution, out *C.double, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtk_wide_lane_fixed_rinex_solution_float_baseline_ecef(solution, out, length)
	})
}

func (s *RtkWideLaneFixedRinexSolution) FixedMetadata() (RtkFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkFixedMetadata{}, ErrClosed
	}
	var value C.SidereonRtkFixedMetadata
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_wide_lane_fixed_rinex_solution_fixed_metadata((*C.SidereonRtkWideLaneFixedRinexSolution)(pointer), &value))
		})
	})
	if err != nil {
		return RtkFixedMetadata{}, err
	}
	return rtkFixedMetadataFromC(value)
}

func (s *RtkWideLaneFixedRinexSolution) FloatMetadata() (RtkFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkFloatMetadata{}, ErrClosed
	}
	var value C.SidereonRtkFloatMetadata
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_wide_lane_fixed_rinex_solution_float_metadata((*C.SidereonRtkWideLaneFixedRinexSolution)(pointer), &value))
		})
	})
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	return rtkFloatMetadataFromC(value)
}

func (s *RtkWideLaneFixedRinexSolution) Metadata() (RtkWideLaneFixedRinexMetadata, error) {
	if s == nil || s.handle == nil {
		return RtkWideLaneFixedRinexMetadata{}, ErrClosed
	}
	var value C.SidereonRtkWideLaneFixedRinexMetadata
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_wide_lane_fixed_rinex_solution_metadata((*C.SidereonRtkWideLaneFixedRinexSolution)(pointer), &value))
		})
	})
	if err != nil {
		return RtkWideLaneFixedRinexMetadata{}, err
	}
	wideLaneCount, err := sizeTToInt(value.wide_lane_ambiguity_count, "RTK wide-lane fixed ambiguity count")
	if err != nil {
		return RtkWideLaneFixedRinexMetadata{}, err
	}
	droppedCount, err := sizeTToInt(value.dropped_cycle_slip_sat_count, "RTK wide-lane dropped cycle-slip count")
	if err != nil {
		return RtkWideLaneFixedRinexMetadata{}, err
	}
	splitCount, err := sizeTToInt(value.split_cycle_slip_arc_count, "RTK wide-lane split cycle-slip count")
	if err != nil {
		return RtkWideLaneFixedRinexMetadata{}, err
	}
	return RtkWideLaneFixedRinexMetadata{WideLaneFixed: bool(value.wide_lane_fixed), WideLaneAmbiguityCount: wideLaneCount, DroppedCycleSlipCount: droppedCount, SplitCycleSlipCount: splitCount}, nil
}

func (s *RtkWideLaneFixedRinexSolution) WideLaneCycles() ([]RtkWideLaneCycleInput, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkWideLaneCycleInput
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var written, required C.size_t
			status := C.sidereon_rtk_wide_lane_fixed_rinex_solution_wide_lane_cycles((*C.SidereonRtkWideLaneFixedRinexSolution)(pointer), nil, 0, &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			count, e := validateNativeQuery("RTK wide-lane fixed RINEX cycles", uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			if _, e = checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkWideLaneCycle{})); e != nil {
				operationErr = e
				return
			}
			rows := make([]C.SidereonRtkWideLaneCycle, count)
			var output *C.SidereonRtkWideLaneCycle
			if count != 0 {
				output = &rows[0]
			}
			written, required = 0, 0
			status = C.sidereon_rtk_wide_lane_fixed_rinex_solution_wide_lane_cycles((*C.SidereonRtkWideLaneFixedRinexSolution)(pointer), output, C.size_t(count), &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			writtenCount, e := validateTwoPassCounts("RTK wide-lane fixed RINEX cycles", count, count, uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			result = make([]RtkWideLaneCycleInput, writtenCount)
			for i := range result {
				result[i] = RtkWideLaneCycleInput{ID: rtkIDFromC(rows[i].id).Value, Cycles: int64(rows[i].cycles)}
			}
		})
		return operationErr
	})
	return result, err
}
