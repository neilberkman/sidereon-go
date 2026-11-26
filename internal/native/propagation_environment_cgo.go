//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func withPropagationEnvironment[T any](handles []*positioningHandle, environment PropagationEnvironment, call func([]unsafe.Pointer, *C.SidereonStatePropagationEnvironment) (T, error)) (T, error) {
	weather := len(handles)
	if environment.SpaceWeather != nil {
		handles = append(handles, orbitSpaceWeatherTableHandle(environment.SpaceWeather))
	}
	return withOrbitContext(handles, environment.Context, func(pointers []unsafe.Pointer, context *C.SidereonPropagationContext) (T, error) {
		var pins runtime.Pinner
		defer pins.Unpin()
		raw := C.SidereonStatePropagationEnvironment{gravity_tide_system: C.uint32_t(environment.GravityTideSystem)}
		if environment.SolarShadow != nil {
			raw.solar_shadow = pinOrbitValue(&pins, orbitRadiationShadowOptionsToC(*environment.SolarShadow))
		}
		if environment.EarthRadiationPressure != nil {
			raw.earth_radiation_pressure = pinOrbitValue(&pins, orbitEarthRadiationToC(*environment.EarthRadiationPressure))
		}
		if context != nil {
			pins.Pin(context)
			raw.context = context
		}
		if environment.SpaceWeather != nil {
			raw.space_weather = (*C.SidereonSpaceWeatherTable)(pointers[weather])
		}
		return call(pointers, &raw)
	})
}

func withStatePropagationForces[T any](environment PropagationEnvironment, additionalForces []*ForceModel, call func(*C.SidereonStatePropagationEnvironment, **C.SidereonForceModel, C.size_t) (T, error)) (T, error) {
	handles := orbitMap(additionalForces, (*ForceModel).orbitHandle)
	return withPropagationEnvironment(handles, environment, func(pointers []unsafe.Pointer, environment *C.SidereonStatePropagationEnvironment) (T, error) {
		forces := orbitMap(pointers[:len(handles)], func(pointer unsafe.Pointer) *C.SidereonForceModel {
			return (*C.SidereonForceModel)(pointer)
		})
		return call(environment, unsafe.SliceData(forces), C.size_t(len(forces)))
	})
}

func PropagateStateWithForceModel(config NativePropagationConfig, epochs []float64, environment PropagationEnvironment, additionalForces []*ForceModel) (*Ephemeris, error) {
	raw := propagationConfig(config)
	return withStatePropagationForces(environment, additionalForces, func(environment *C.SidereonStatePropagationEnvironment, forces **C.SidereonForceModel, count C.size_t) (*Ephemeris, error) {
		var value *C.SidereonEphemeris
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_propagate_state_with_force_model(&raw, (*C.double)(unsafe.Pointer(unsafe.SliceData(epochs))), C.size_t(len(epochs)), environment, forces, count, &value))
		})
		handle, err := adoptOrbitSurface(unsafe.Pointer(value), err, releaseEphemeris)
		if err != nil {
			return nil, err
		}
		return &Ephemeris{handle: handle}, nil
	})
}

func PropagateCovarianceWithContext(config NativePropagationConfig, covariance [6][6]float64, epochs []float64, inputFrame, outputFrame uint32, noise NativeProcessNoise, environment PropagationEnvironment) (*CovarianceEphemeris, error) {
	raw := propagationConfig(config)
	base := covarianceToC(covariance)
	options := C.SidereonCovariancePropagationOptions{input_frame: C.uint32_t(inputFrame), output_frame: C.uint32_t(outputFrame), process_noise: C.SidereonProcessNoise{
		kind: C.uint32_t(noise.Kind), q_radial_km2_s3: C.double(noise.RadialKm2S3), q_transverse_km2_s3: C.double(noise.TransverseKm2S3), q_normal_km2_s3: C.double(noise.NormalKm2S3),
	}}
	return withPropagationEnvironment(nil, environment, func(_ []unsafe.Pointer, environment *C.SidereonStatePropagationEnvironment) (*CovarianceEphemeris, error) {
		var value *C.SidereonCovarianceEphemeris
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_propagate_covariance_with_context(&raw, &base, (*C.double)(unsafe.Pointer(unsafe.SliceData(epochs))), C.size_t(len(epochs)), options, environment, &value))
		})
		handle, err := adoptOrbitSurface(unsafe.Pointer(value), err, func(pointer unsafe.Pointer) {
			C.sidereon_covariance_ephemeris_free((*C.SidereonCovarianceEphemeris)(pointer))
		})
		if err != nil {
			return nil, err
		}
		return &CovarianceEphemeris{handle: handle}, nil
	})
}

func PropagateStateWithCovarianceWithContext(config NativePropagationConfig, covariance [6][6]float64, spanSeconds float64, environment PropagationEnvironment) (StateWithCovariance, error) {
	raw := propagationConfig(config)
	base := covarianceToC(covariance)
	return withPropagationEnvironment(nil, environment, func(_ []unsafe.Pointer, environment *C.SidereonStatePropagationEnvironment) (StateWithCovariance, error) {
		var value C.SidereonStateWithCovariance
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_propagate_state_with_covariance_with_context(&raw, &base, C.double(spanSeconds), environment, &value))
		})
		return StateWithCovariance{State: cartesianStateFromC(value.state), Covariance: covarianceFromC(&value.covariance)}, err
	})
}

func StateTransitionMatrixToWithContext(config NativePropagationConfig, epochTDBSeconds float64, environment PropagationEnvironment) ([6][6]float64, error) {
	raw := propagationConfig(config)
	return withPropagationEnvironment(nil, environment, func(_ []unsafe.Pointer, environment *C.SidereonStatePropagationEnvironment) ([6][6]float64, error) {
		var value [6][6]float64
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_state_transition_matrix_to_with_context(&raw, C.double(epochTDBSeconds), environment, (*C.double)(unsafe.Pointer(&value))))
		})
		return value, err
	})
}

func EstimateDecayWithContext(initial CartesianState, config DecayConfig, table *SpaceWeatherTable, context *PropagationContext) (DecayEstimate, error) {
	raw, err := cDecayConfig(config)
	if err != nil {
		return DecayEstimate{}, err
	}
	var handles []*positioningHandle
	if table != nil {
		handles = append(handles, orbitSpaceWeatherTableHandle(table))
	}
	state := cCartesianState(initial)
	return withOrbitContext(handles, context, func(pointers []unsafe.Pointer, context *C.SidereonPropagationContext) (DecayEstimate, error) {
		var value C.SidereonDecayEstimate
		err := callStatus(func() uint32 {
			if table != nil {
				return uint32(C.sidereon_estimate_decay_with_source_and_context(&state, &raw, (*C.SidereonSpaceWeatherTable)(pointers[0]), context, &value))
			}
			return uint32(C.sidereon_estimate_decay_with_context(&state, &raw, context, &value))
		})
		return DecayEstimate{
			TimeToDecayS: float64(value.time_to_decay_s), ReentryState: cartesianStateFromC(value.reentry_state),
			ReentryAltitudeKm: float64(value.reentry_altitude_km),
		}, err
	})
}

func adoptOrbitSurface(pointer unsafe.Pointer, err error, release func(unsafe.Pointer)) (*surfaceHandle, error) {
	if err != nil {
		if pointer != nil {
			withCThread(func() { release(pointer) })
		}
		return nil, err
	}
	handle, err := newSurfaceHandle(pointer, release)
	if err != nil && pointer != nil {
		withCThread(func() { release(pointer) })
	}
	return handle, err
}

func findTCAConjunctionsWithForceOptions(lines [4]*C.char, epochs [4]float64, tcaOptions *C.SidereonTcaFinderOptions, pcOptions *C.SidereonTcaPropagatedCovariancePcOptions, input TCAForceOptions) ([]TCAConjunction, error) {
	return withPropagationEnvironment(nil, input.Environment, func(_ []unsafe.Pointer, environment *C.SidereonStatePropagationEnvironment) ([]TCAConjunction, error) {
		options := C.SidereonTcaForceOptions{
			force_components:         cForceModelComponents(input.ForceComponents),
			gravity_tide_system:      environment.gravity_tide_system,
			solar_shadow:             environment.solar_shadow,
			earth_radiation_pressure: environment.earth_radiation_pressure,
		}
		rows, err := orbitReadRows("TCA conjunctions", func(out *C.SidereonTcaConjunction, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_find_tca_conjunctions_with_propagated_covariance_from_tles_with_force_options(
				lines[0], lines[1], lines[2], lines[3], C.double(epochs[0]), C.double(epochs[1]), C.double(epochs[2]), C.double(epochs[3]),
				tcaOptions, pcOptions, &options, out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(rows, tcaConjunctionFromC), nil
	})
}

func PropagateStateWithDenseOutput(config NativePropagationConfig, targetTDBSeconds float64, environment PropagationEnvironment, additionalForces []*ForceModel) (state CartesianState, dense *DenseOutput, err error) {
	raw := propagationConfig(config)
	dense, err = withStatePropagationForces(environment, additionalForces, func(environment *C.SidereonStatePropagationEnvironment, forces **C.SidereonForceModel, count C.size_t) (*DenseOutput, error) {
		var final C.SidereonCartesianState
		var value *C.SidereonDenseOutput
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_propagate_state_with_dense_output(&raw, C.double(targetTDBSeconds), environment, forces, count, &final, &value))
		})
		if err == nil {
			state = cartesianStateFromC(final)
		}
		if value == nil {
			return nil, err
		}
		handle, err := adoptOrbitHandle(unsafe.Pointer(value), err, func(pointer unsafe.Pointer) {
			C.sidereon_dense_output_free((*C.SidereonDenseOutput)(pointer))
		})
		if err != nil {
			return nil, err
		}
		return &DenseOutput{handle: handle}, nil
	})
	return state, dense, err
}
