//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func OrbitParameterFitCorrelation(fit *OrbitParameterFit, first string, second string) (OptionalOrbitScalar, error) {
	firstC := C.CBytes([]byte(first))
	defer C.free(firstC)
	secondC := C.CBytes([]byte(second))
	defer C.free(secondC)
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) (OptionalOrbitScalar, error) {
		var value C.SidereonOptionalOrbitScalar
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_parameter_fit_correlation((*C.SidereonOrbitParameterFit)(pointers[0]), C.SidereonOrbitParameterName{data: (*C.uint8_t)(firstC), len: C.size_t(len(first))}, C.SidereonOrbitParameterName{data: (*C.uint8_t)(secondC), len: C.size_t(len(second))}, &value))
		})
		return orbitOptionalOrbitScalarFromC(value), err
	})
}

func OrbitParameterFitCovariance(fit *OrbitParameterFit) (OrbitParameterCovariance, error) {
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) (OrbitParameterCovariance, error) {
		var value C.SidereonOrbitParameterCovariance
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_parameter_fit_covariance((*C.SidereonOrbitParameterFit)(pointers[0]), &value))
		})
		return orbitOrbitParameterCovarianceFromC(value), err
	})
}

func OrbitParameterFitCovarianceMatrix(fit *OrbitParameterFit, correlation bool) ([]float64, error) {
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) ([]float64, error) {
		raw, err := orbitReadRows("orbit parameter fit covariance matrix", func(out *C.double, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_parameter_fit_covariance_matrix((*C.SidereonOrbitParameterFit)(pointers[0]), C.bool(correlation), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.double) float64 { return float64(item) }), nil
	})
}

func OrbitParameterFitEqual(left *OrbitParameterFit, right *OrbitParameterFit) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_parameter_fit_equal((*C.SidereonOrbitParameterFit)(pointers[0]), (*C.SidereonOrbitParameterFit)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func OrbitParameterFitForceParameter(fit *OrbitParameterFit, name string) (EstimatedForceParameter, error) {
	nameC := C.CBytes([]byte(name))
	defer C.free(nameC)
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) (EstimatedForceParameter, error) {
		var value C.SidereonEstimatedForceParameter
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_parameter_fit_force_parameter((*C.SidereonOrbitParameterFit)(pointers[0]), C.SidereonOrbitParameterName{data: (*C.uint8_t)(nameC), len: C.size_t(len(name))}, &value))
		})
		result := orbitEstimatedForceParameterFromC(value)
		if bool(value.present) {
			result.Name = name
		}
		return result, err
	})
}

func OrbitParameterFitForceParameterAt(fit *OrbitParameterFit, index uint64) (EstimatedForceParameter, error) {
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) (EstimatedForceParameter, error) {
		var value C.SidereonEstimatedForceParameter
		name, err := orbitText("orbit fit force parameter name", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_parameter_fit_force_parameter_at((*C.SidereonOrbitParameterFit)(pointers[0]), C.size_t(index), out, capacity, written, required, &value)
		})
		result := orbitEstimatedForceParameterFromC(value)
		result.Name = name
		return result, err
	})
}

func OrbitParameterFitParameterName(fit *OrbitParameterFit, index uint64) (string, error) {
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) (string, error) {
		return orbitText("orbit parameter fit parameter name", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_parameter_fit_parameter_name((*C.SidereonOrbitParameterFit)(pointers[0]), C.size_t(index), out, capacity, written, required)
		})
	})
}

func OrbitParameterFitResiduals(fit *OrbitParameterFit) ([]OrbitRTNResidual, error) {
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) ([]OrbitRTNResidual, error) {
		raw, err := orbitReadRows("orbit parameter fit residuals", func(out *C.SidereonOrbitRtnResidual, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_parameter_fit_residuals((*C.SidereonOrbitParameterFit)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonOrbitRtnResidual) OrbitRTNResidual { return orbitOrbitRTNResidualFromC(item) }), nil
	})
}

func OrbitParameterFitSigma(fit *OrbitParameterFit, name string) (OptionalOrbitScalar, error) {
	nameC := C.CBytes([]byte(name))
	defer C.free(nameC)
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) (OptionalOrbitScalar, error) {
		var value C.SidereonOptionalOrbitScalar
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_parameter_fit_sigma((*C.SidereonOrbitParameterFit)(pointers[0]), C.SidereonOrbitParameterName{data: (*C.uint8_t)(nameC), len: C.size_t(len(name))}, &value))
		})
		return orbitOptionalOrbitScalarFromC(value), err
	})
}

func ReadOrbitParameterFitSummary(fit *OrbitParameterFit) (OrbitParameterFitSummary, error) {
	return withOrbitHandles([]*positioningHandle{fit.orbitHandle()}, func(pointers []unsafe.Pointer) (OrbitParameterFitSummary, error) {
		var value C.SidereonOrbitParameterFitSummary
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_parameter_fit_summary((*C.SidereonOrbitParameterFit)(pointers[0]), &value))
		})
		return orbitOrbitParameterFitSummaryFromC(value), err
	})
}

func StateParameterName(index uint64) (string, error) {
	return orbitText("state parameter name", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_state_parameter_name(C.size_t(index), out, capacity, written, required)
	})
}

func StateParameterNameCount() uint64 {
	value := C.sidereon_state_parameter_name_count()
	return uint64(value)
}

func OrbitParameterFitOptionsDefault() OrbitParameterFitOptions {
	value := C.sidereon_orbit_parameter_fit_options_default()
	return OrbitParameterFitOptions{
		GravityTideSystem: uint32(value.gravity_tide_system),
		ForceModel:        uint32(value.force_model), ForceComponents: forceModelComponentsFromC(value.force_components),
		IntegratorOptions: orbitIntegratorOptionsFromC(value.integrator_options), ObservationSigmaM: float64(value.observation_sigma_m),
		MaxIterations: uint64(value.max_iterations), ConvergenceThreshold: float64(value.convergence_threshold),
		MaxConditionNumber: float64(value.max_condition_number), MinLedgerSamples: uint64(value.min_ledger_samples),
	}
}

func FitInertialOrbitParameters(samples []InertialPositionSample, satellite string, initialState *CartesianState, options OrbitParameterFitOptions, permissive *bool) (*OrbitParameterFit, error) {
	rows := orbitMap(samples, func(sample InertialPositionSample) C.SidereonInertialPositionSample {
		return C.SidereonInertialPositionSample{epoch: clockEpochToC(sample.Epoch), position_gcrf_m: orbitVectorToC(sample.PositionGCRFM).xyz}
	})
	return withOrbitParameterFit(nil, satellite, initialState, options, func(_ []unsafe.Pointer, token *C.char, initial *C.SidereonCartesianState, raw *C.SidereonOrbitParameterFitOptions, out **C.SidereonOrbitParameterFit) C.enum_SidereonStatus {
		if permissive != nil {
			return C.sidereon_fit_inertial_orbit_parameters_with_validity(unsafe.SliceData(rows), C.size_t(len(rows)), token, initial, raw, C.bool(*permissive), out)
		}
		return C.sidereon_fit_inertial_orbit_parameters(unsafe.SliceData(rows), C.size_t(len(rows)), token, initial, raw, out)
	})
}

func FitSP3ECEFOrbitParameters(product *SP3, provider *IERS2010EarthOrientationProvider, satellite string, initialState *CartesianState, options OrbitParameterFitOptions, permissive *bool) (*OrbitParameterFit, error) {
	handles := []*positioningHandle{orbitSP3Handle(product)}
	if provider != nil {
		handles = append(handles, provider.orbitHandle())
	}
	return withOrbitParameterFit(handles, satellite, initialState, options, func(pointers []unsafe.Pointer, token *C.char, initial *C.SidereonCartesianState, raw *C.SidereonOrbitParameterFitOptions, out **C.SidereonOrbitParameterFit) C.enum_SidereonStatus {
		var orientation *C.SidereonIers2010EarthOrientationProvider
		if provider != nil {
			orientation = (*C.SidereonIers2010EarthOrientationProvider)(pointers[1])
		}
		product := (*C.SidereonSp3)(pointers[0])
		if permissive != nil {
			return C.sidereon_fit_sp3_ecef_orbit_parameters_with_validity(product, orientation, token, initial, raw, C.bool(*permissive), out)
		}
		return C.sidereon_fit_sp3_ecef_orbit_parameters(product, orientation, token, initial, raw, out)
	})
}

type orbitParameterFitCall func([]unsafe.Pointer, *C.char, *C.SidereonCartesianState, *C.SidereonOrbitParameterFitOptions, **C.SidereonOrbitParameterFit) C.enum_SidereonStatus

func withOrbitParameterFit(handles []*positioningHandle, satellite string, initialState *CartesianState, options OrbitParameterFitOptions, call orbitParameterFitCall) (*OrbitParameterFit, error) {
	token, err := copyNativeCString(satellite, "satellite")
	if err != nil {
		return nil, err
	}
	defer C.free(token)
	var initial *C.SidereonCartesianState
	if initialState != nil {
		value := cCartesianState(*initialState)
		initial = &value
	}
	firstForce := len(handles)
	for _, force := range options.AdditionalForces {
		handles = append(handles, force.orbitHandle())
	}
	afterForces := len(handles)
	if options.SpaceWeather != nil {
		handles = append(handles, orbitSpaceWeatherTableHandle(options.SpaceWeather))
	}
	return withOrbitContext(handles, options.Context, func(pointers []unsafe.Pointer, context *C.SidereonPropagationContext) (*OrbitParameterFit, error) {
		var pins runtime.Pinner
		defer pins.Unpin()
		raw := orbitParameterFitOptionsToC(options, &pins)
		forces := orbitMap(pointers[firstForce:afterForces], func(pointer unsafe.Pointer) *C.SidereonForceModel {
			return (*C.SidereonForceModel)(pointer)
		})
		raw.additional_forces = pinOrbitSlice(&pins, forces)
		raw.additional_force_count = C.size_t(len(forces))
		if options.SpaceWeather != nil {
			raw.space_weather = (*C.SidereonSpaceWeatherTable)(pointers[afterForces])
		}
		if context != nil {
			pins.Pin(context)
			raw.propagation_context = context
		}
		var value *C.SidereonOrbitParameterFit
		err := callStatus(func() uint32 {
			return uint32(call(pointers, (*C.char)(token), initial, &raw, &value))
		})
		return newOrbitOrbitParameterFit(value, err)
	})
}

func orbitParameterFitOptionsToC(value OrbitParameterFitOptions, pins *runtime.Pinner) C.SidereonOrbitParameterFitOptions {
	names := make([]C.SidereonOrbitParameterName, len(value.EstimatedParameters))
	for i, name := range value.EstimatedParameters {
		bytes := []byte(name)
		names[i] = C.SidereonOrbitParameterName{
			data: (*C.uint8_t)(unsafe.Pointer(pinOrbitSlice(pins, bytes))), len: C.size_t(len(bytes)),
		}
	}
	raw := C.SidereonOrbitParameterFitOptions{
		gravity_tide_system: C.uint32_t(value.GravityTideSystem),
		force_model:         C.uint32_t(value.ForceModel), force_components: cForceModelComponents(value.ForceComponents),
		estimated_parameters: pinOrbitSlice(pins, names), estimated_parameter_count: C.size_t(len(names)),
		integrator_options: orbitIntegratorOptionsToC(value.IntegratorOptions), observation_sigma_m: C.double(value.ObservationSigmaM),
		max_iterations: C.size_t(value.MaxIterations), convergence_threshold: C.double(value.ConvergenceThreshold),
		max_condition_number: C.double(value.MaxConditionNumber), min_ledger_samples: C.size_t(value.MinLedgerSamples),
	}
	if value.MuKm3S2 != nil {
		raw.mu_km3_s2 = pinOrbitValue(pins, C.double(*value.MuKm3S2))
	}
	if value.Drag != nil {
		raw.drag = pinOrbitValue(pins, cDragParameters(*value.Drag))
	}
	if value.SolarShadow != nil {
		raw.solar_shadow = pinOrbitValue(pins, orbitRadiationShadowOptionsToC(*value.SolarShadow))
	}
	if value.EarthRadiationPressure != nil {
		raw.earth_radiation_pressure = pinOrbitValue(pins, orbitEarthRadiationToC(*value.EarthRadiationPressure))
	}
	return raw
}
