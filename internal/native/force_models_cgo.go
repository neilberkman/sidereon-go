//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

func AccelerationPartialsPositionOnly(wrtPosition [3][3]float64) (AccelerationPartials, error) {
	var value C.SidereonAccelerationPartials
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_acceleration_partials_position_only((*C.double)(unsafe.Pointer(&wrtPosition)), &value))
	})
	return orbitAccelerationPartialsFromC(value), err
}

func AccelerationPartialsZeros() AccelerationPartials {
	value := C.sidereon_acceleration_partials_zeros()
	return orbitAccelerationPartialsFromC(value)
}

func DragForceNew(parameters DragParameters) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_drag_force_new(cDragParameters(parameters), &value)) })
	return newOrbitForceModel(value, err)
}

func EarthRadiationPressureAccelerationWithSunMoon(model *ForceModel, state CartesianState, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_earth_radiation_pressure_acceleration_with_sun_moon((*C.SidereonForceModel)(pointers[0]), cCartesianState(state), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func EarthRadiationPressureMoonEclipseEnabled(model *ForceModel) (bool, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_earth_radiation_pressure_moon_eclipse_enabled((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return bool(value), err
	})
}

func EarthRadiationPressureNew(cr float64, areaToMassM2Kg float64, moonEclipseEnabled bool) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_earth_radiation_pressure_new(C.double(cr), C.double(areaToMassM2Kg), C.bool(moonEclipseEnabled), &value))
	})
	return newOrbitForceModel(value, err)
}

func EarthRadiationPressureWithMoonEclipse(model *ForceModel, enabled bool) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_earth_radiation_pressure_with_moon_eclipse((*C.SidereonForceModel)(pointers[0]), C.bool(enabled), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ECOM1CoefficientCount() uint64 {
	value := C.sidereon_ecom1_coefficient_count()
	return uint64(value)
}

func ECOM1CoefficientsFromArray(values [9]float64) (ECOM1Coefficients, error) {
	var value C.SidereonEcom1Coefficients
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom1_coefficients_from_array((*C.double)(unsafe.Pointer(&values)), &value))
	})
	return orbitECOM1CoefficientsFromC(value), err
}

func ECOM1CoefficientsNames(index uint64) (string, error) {
	return orbitText("ecom1 coefficients names", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ecom1_coefficients_names(C.size_t(index), out, capacity, written, required)
	})
}

func ECOM1CoefficientsParameterNames(index uint64) (string, error) {
	return orbitText("ecom1 coefficients parameter names", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ecom1_coefficients_parameter_names(C.size_t(index), out, capacity, written, required)
	})
}

func ECOM1CoefficientsToArray(coefficients ECOM1Coefficients) ([9]float64, error) {
	var value [9]float64
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom1_coefficients_to_array(orbitECOM1CoefficientsToC(coefficients), (*C.double)(unsafe.Pointer(&value))))
	})
	return value, err
}

func ECOM1SolarRadiationPressureAccelerationPartialsWithSun(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) (AccelerationPartials, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (AccelerationPartials, error) {
		var value C.SidereonAccelerationPartials
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_acceleration_partials_with_sun((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), &value))
		})
		return orbitAccelerationPartialsFromC(value), err
	})
}

func ECOM1SolarRadiationPressureAccelerationPartialsWithSunMoon(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) (AccelerationPartials, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (AccelerationPartials, error) {
		var value C.SidereonAccelerationPartials
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_acceleration_partials_with_sun_moon((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), &value))
		})
		return orbitAccelerationPartialsFromC(value), err
	})
}

func ECOM1SolarRadiationPressureAccelerationWithSun(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_acceleration_with_sun((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func ECOM1SolarRadiationPressureAccelerationWithSunMoon(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_acceleration_with_sun_moon((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func ECOM1SolarRadiationPressureCoefficients(model *ForceModel) (ECOM1Coefficients, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (ECOM1Coefficients, error) {
		var value C.SidereonEcom1Coefficients
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_coefficients((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return orbitECOM1CoefficientsFromC(value), err
	})
}

func ECOM1SolarRadiationPressureEarthShadowModel(model *ForceModel) (uint32, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (uint32, error) {
		var value C.enum_SidereonEarthShadowModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_earth_shadow_model((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return uint32(value), err
	})
}

func ECOM1SolarRadiationPressureEqual(left *ForceModel, right *ForceModel) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_equal((*C.SidereonForceModel)(pointers[0]), (*C.SidereonForceModel)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func ECOM1SolarRadiationPressureMoonEclipseEnabled(model *ForceModel) (bool, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_moon_eclipse_enabled((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return bool(value), err
	})
}

func ECOM1SolarRadiationPressureNew(coefficients ECOM1Coefficients, shadow RadiationShadowOptions) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom1_solar_radiation_pressure_new(orbitECOM1CoefficientsToC(coefficients), orbitRadiationShadowOptionsToC(shadow), &value))
	})
	return newOrbitForceModel(value, err)
}

func ECOM1SolarRadiationPressureParameterPartialsWithSun(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([][3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([][3]float64, error) {
		raw, err := orbitReadRows("ecom1 solar radiation pressure parameter partials with sun", func(out *C.SidereonVector3, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ecom1_solar_radiation_pressure_parameter_partials_with_sun((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonVector3) [3]float64 { return orbitVectorFromC(item) }), nil
	})
}

func ECOM1SolarRadiationPressureParameterPartialsWithSunMoon(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([][3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([][3]float64, error) {
		raw, err := orbitReadRows("ecom1 solar radiation pressure parameter partials with sun moon", func(out *C.SidereonVector3, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ecom1_solar_radiation_pressure_parameter_partials_with_sun_moon((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonVector3) [3]float64 { return orbitVectorFromC(item) }), nil
	})
}

func ECOM1SolarRadiationPressureShadowFunction(model *ForceModel) (ShadowFunction, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (ShadowFunction, error) {
		var value C.enum_SidereonShadowFunction
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_shadow_function((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return ShadowFunction(value), err
	})
}

func ECOM1SolarRadiationPressureWithEarthShadowModel(model *ForceModel, earthModel uint32) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_with_earth_shadow_model((*C.SidereonForceModel)(pointers[0]), C.uint32_t(earthModel), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ECOM1SolarRadiationPressureWithMoonEclipse(model *ForceModel, enabled bool) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_with_moon_eclipse((*C.SidereonForceModel)(pointers[0]), C.bool(enabled), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ECOM1SolarRadiationPressureWithShadowFunction(model *ForceModel, shadowFunction ShadowFunction) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom1_solar_radiation_pressure_with_shadow_function((*C.SidereonForceModel)(pointers[0]), C.uint32_t(shadowFunction), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ECOM2D2B1CoefficientCount() uint64 {
	value := C.sidereon_ecom2_d2b1_coefficient_count()
	return uint64(value)
}

func ECOM2D2B1CoefficientsFromArray(values [7]float64) (ECOM2D2B1Coefficients, error) {
	var value C.SidereonEcom2D2b1Coefficients
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_d2b1_coefficients_from_array((*C.double)(unsafe.Pointer(&values)), &value))
	})
	return orbitECOM2D2B1CoefficientsFromC(value), err
}

func ECOM2D2B1CoefficientsNames(index uint64) (string, error) {
	return orbitText("ecom2 d2b1 coefficients names", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ecom2_d2b1_coefficients_names(C.size_t(index), out, capacity, written, required)
	})
}

func ECOM2D2B1CoefficientsParameterNames(index uint64) (string, error) {
	return orbitText("ecom2 d2b1 coefficients parameter names", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ecom2_d2b1_coefficients_parameter_names(C.size_t(index), out, capacity, written, required)
	})
}

func ECOM2D2B1CoefficientsToArray(coefficients ECOM2D2B1Coefficients) ([7]float64, error) {
	var value [7]float64
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_d2b1_coefficients_to_array(orbitECOM2D2B1CoefficientsToC(coefficients), (*C.double)(unsafe.Pointer(&value))))
	})
	return value, err
}

func ECOM2D4B1CoefficientCount() uint64 {
	value := C.sidereon_ecom2_d4b1_coefficient_count()
	return uint64(value)
}

func ECOM2D4B1CoefficientsFromArray(values [9]float64) (ECOM2D4B1Coefficients, error) {
	var value C.SidereonEcom2D4b1Coefficients
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_d4b1_coefficients_from_array((*C.double)(unsafe.Pointer(&values)), &value))
	})
	return orbitECOM2D4B1CoefficientsFromC(value), err
}

func ECOM2D4B1CoefficientsNames(index uint64) (string, error) {
	return orbitText("ecom2 d4b1 coefficients names", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ecom2_d4b1_coefficients_names(C.size_t(index), out, capacity, written, required)
	})
}

func ECOM2D4B1CoefficientsParameterNames(index uint64) (string, error) {
	return orbitText("ecom2 d4b1 coefficients parameter names", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ecom2_d4b1_coefficients_parameter_names(C.size_t(index), out, capacity, written, required)
	})
}

func ECOM2D4B1CoefficientsToArray(coefficients ECOM2D4B1Coefficients) ([9]float64, error) {
	var value [9]float64
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_d4b1_coefficients_to_array(orbitECOM2D4B1CoefficientsToC(coefficients), (*C.double)(unsafe.Pointer(&value))))
	})
	return value, err
}

func ECOM2SolarRadiationPressureAccelerationPartialsWithSun(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) (AccelerationPartials, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (AccelerationPartials, error) {
		var value C.SidereonAccelerationPartials
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_acceleration_partials_with_sun((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), &value))
		})
		return orbitAccelerationPartialsFromC(value), err
	})
}

func ECOM2SolarRadiationPressureAccelerationPartialsWithSunMoon(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) (AccelerationPartials, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (AccelerationPartials, error) {
		var value C.SidereonAccelerationPartials
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_acceleration_partials_with_sun_moon((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), &value))
		})
		return orbitAccelerationPartialsFromC(value), err
	})
}

func ECOM2SolarRadiationPressureAccelerationWithSun(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_acceleration_with_sun((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func ECOM2SolarRadiationPressureAccelerationWithSunMoon(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_acceleration_with_sun_moon((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func ECOM2SolarRadiationPressureCoefficientCount(nD uint64, nB uint64) (uint64, error) {
	var value C.size_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_solar_radiation_pressure_coefficient_count(C.size_t(nD), C.size_t(nB), &value))
	})
	return uint64(value), err
}

func ECOM2SolarRadiationPressureCoefficients(model *ForceModel) ([]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([]float64, error) {
		raw, err := orbitReadRows("ecom2 solar radiation pressure coefficients", func(out *C.double, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ecom2_solar_radiation_pressure_coefficients((*C.SidereonForceModel)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.double) float64 { return float64(item) }), nil
	})
}

func ECOM2SolarRadiationPressureD2B1(coefficients ECOM2D2B1Coefficients, shadow RadiationShadowOptions) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_solar_radiation_pressure_d2b1(orbitECOM2D2B1CoefficientsToC(coefficients), orbitRadiationShadowOptionsToC(shadow), &value))
	})
	return newOrbitForceModel(value, err)
}

func ECOM2SolarRadiationPressureD4B1(coefficients ECOM2D4B1Coefficients, shadow RadiationShadowOptions) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_solar_radiation_pressure_d4b1(orbitECOM2D4B1CoefficientsToC(coefficients), orbitRadiationShadowOptionsToC(shadow), &value))
	})
	return newOrbitForceModel(value, err)
}

func ECOM2SolarRadiationPressureEarthShadowModel(model *ForceModel) (uint32, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (uint32, error) {
		var value C.enum_SidereonEarthShadowModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_earth_shadow_model((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return uint32(value), err
	})
}

func ECOM2SolarRadiationPressureEqual(left *ForceModel, right *ForceModel) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_equal((*C.SidereonForceModel)(pointers[0]), (*C.SidereonForceModel)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func ECOM2SolarRadiationPressureMoonEclipseEnabled(model *ForceModel) (bool, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_moon_eclipse_enabled((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return bool(value), err
	})
}

func ECOM2SolarRadiationPressureNB(model *ForceModel) (uint64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (uint64, error) {
		var value C.size_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_n_b((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return uint64(value), err
	})
}

func ECOM2SolarRadiationPressureND(model *ForceModel) (uint64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (uint64, error) {
		var value C.size_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_n_d((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return uint64(value), err
	})
}

func ECOM2SolarRadiationPressureNew(nD uint64, nB uint64, coefficients []float64, shadow RadiationShadowOptions) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom2_solar_radiation_pressure_new(C.size_t(nD), C.size_t(nB), (*C.double)(unsafe.Pointer(unsafe.SliceData(coefficients))), C.size_t(len(coefficients)), orbitRadiationShadowOptionsToC(shadow), &value))
	})
	return newOrbitForceModel(value, err)
}

func ECOM2SolarRadiationPressureParameterName(nD uint64, nB uint64, index uint64) (string, error) {
	return orbitText("ecom2 solar radiation pressure parameter name", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ecom2_solar_radiation_pressure_parameter_name(C.size_t(nD), C.size_t(nB), C.size_t(index), out, capacity, written, required)
	})
}

func ECOM2SolarRadiationPressureParameterPartialsWithSun(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) ([][3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([][3]float64, error) {
		raw, err := orbitReadRows("ecom2 solar radiation pressure parameter partials with sun", func(out *C.SidereonVector3, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ecom2_solar_radiation_pressure_parameter_partials_with_sun((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonVector3) [3]float64 { return orbitVectorFromC(item) }), nil
	})
}

func ECOM2SolarRadiationPressureParameterPartialsWithSunMoon(model *ForceModel, positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([][3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([][3]float64, error) {
		raw, err := orbitReadRows("ecom2 solar radiation pressure parameter partials with sun moon", func(out *C.SidereonVector3, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ecom2_solar_radiation_pressure_parameter_partials_with_sun_moon((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonVector3) [3]float64 { return orbitVectorFromC(item) }), nil
	})
}

func ECOM2SolarRadiationPressureShadowFunction(model *ForceModel) (ShadowFunction, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (ShadowFunction, error) {
		var value C.enum_SidereonShadowFunction
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_shadow_function((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return ShadowFunction(value), err
	})
}

func ECOM2SolarRadiationPressureWithEarthShadowModel(model *ForceModel, earthModel uint32) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_with_earth_shadow_model((*C.SidereonForceModel)(pointers[0]), C.uint32_t(earthModel), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ECOM2SolarRadiationPressureWithMoonEclipse(model *ForceModel, enabled bool) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_with_moon_eclipse((*C.SidereonForceModel)(pointers[0]), C.bool(enabled), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ECOM2SolarRadiationPressureWithShadowFunction(model *ForceModel, shadowFunction ShadowFunction) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_ecom2_solar_radiation_pressure_with_shadow_function((*C.SidereonForceModel)(pointers[0]), C.uint32_t(shadowFunction), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ComputeECOMGeometry(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64) (ECOMGeometry, error) {
	var value C.SidereonEcomGeometry
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom_geometry(orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), &value))
	})
	return orbitECOMGeometryFromC(value), err
}

func ECOMGeometryWithShadow(positionKm [3]float64, velocityKmS [3]float64, sunPositionKm [3]float64, moonPositionKm *[3]float64, earthShadowModel uint32) (ECOMGeometry, error) {
	var moonPositionKmC *C.SidereonVector3
	if moonPositionKm != nil {
		value := orbitVectorToC(*moonPositionKm)
		moonPositionKmC = &value
	}
	var value C.SidereonEcomGeometry
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_ecom_geometry_with_shadow(orbitVectorToC(positionKm), orbitVectorToC(velocityKmS), orbitVectorToC(sunPositionKm), moonPositionKmC, C.uint32_t(earthShadowModel), &value))
	})
	return orbitECOMGeometryFromC(value), err
}

func EmpiricalHarmonicAll() ([3]EmpiricalHarmonic, error) {
	var raw [3]C.enum_SidereonEmpiricalHarmonic
	err := callStatus(func() uint32 { return uint32(C.sidereon_empirical_harmonic_all(&raw[0])) })
	var value [3]EmpiricalHarmonic
	for i, item := range raw {
		value[i] = EmpiricalHarmonic(item)
	}
	return value, err
}

func EmpiricalRTNAccelerationAllTerms() (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_empirical_rtn_acceleration_all_terms(&value)) })
	return newOrbitForceModel(value, err)
}

func EmpiricalRTNAccelerationCoefficientKmS2(model *ForceModel, axis RTNAxis, harmonic EmpiricalHarmonic) (OptionalOrbitScalar, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (OptionalOrbitScalar, error) {
		var value C.SidereonOptionalOrbitScalar
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_empirical_rtn_acceleration_coefficient_km_s2((*C.SidereonForceModel)(pointers[0]), C.uint32_t(axis), C.uint32_t(harmonic), &value))
		})
		return orbitOptionalOrbitScalarFromC(value), err
	})
}

func EmpiricalRTNAccelerationConstantTerms() (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_empirical_rtn_acceleration_constant_terms(&value)) })
	return newOrbitForceModel(value, err)
}

func EmpiricalRTNAccelerationEqual(left *ForceModel, right *ForceModel) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_empirical_rtn_acceleration_equal((*C.SidereonForceModel)(pointers[0]), (*C.SidereonForceModel)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func EmpiricalRTNAccelerationNew(terms []EmpiricalRTNTerm) (*ForceModel, error) {
	termsC := orbitMap(terms, func(item EmpiricalRTNTerm) C.SidereonEmpiricalRtnTerm { return orbitEmpiricalRTNTermToC(item) })
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_empirical_rtn_acceleration_new(unsafe.SliceData(termsC), C.size_t(len(terms)), &value))
	})
	return newOrbitForceModel(value, err)
}

func EmpiricalRTNAccelerationParameterName(axis RTNAxis, harmonic EmpiricalHarmonic) (string, error) {
	return orbitText("empirical rtn acceleration parameter name", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_empirical_rtn_acceleration_parameter_name(C.uint32_t(axis), C.uint32_t(harmonic), out, capacity, written, required)
	})
}

func EmpiricalRTNAccelerationTerms(model *ForceModel) ([]EmpiricalRTNTerm, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([]EmpiricalRTNTerm, error) {
		raw, err := orbitReadRows("empirical rtn acceleration terms", func(out *C.SidereonEmpiricalRtnTerm, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_empirical_rtn_acceleration_terms((*C.SidereonForceModel)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonEmpiricalRtnTerm) EmpiricalRTNTerm { return orbitEmpiricalRTNTermFromC(item) }), nil
	})
}

func EmpiricalRTNAccelerationWithTerm(model *ForceModel, term EmpiricalRTNTerm) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_empirical_rtn_acceleration_with_term((*C.SidereonForceModel)(pointers[0]), orbitEmpiricalRTNTermToC(term), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func EmpiricalRTNAccelerationWithZeroTerms(terms []EmpiricalRTNTermSelector) (*ForceModel, error) {
	termsC := orbitMap(terms, func(item EmpiricalRTNTermSelector) C.SidereonEmpiricalRtnTermSelector {
		return orbitEmpiricalRTNTermSelectorToC(item)
	})
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_empirical_rtn_acceleration_with_zero_terms(unsafe.SliceData(termsC), C.size_t(len(terms)), &value))
	})
	return newOrbitForceModel(value, err)
}

func ForceModelAcceleration(model *ForceModel, state CartesianState, context *PropagationContext) ([3]float64, error) {
	return withOrbitContext([]*positioningHandle{model.orbitHandle()}, context, func(pointers []unsafe.Pointer, contextC *C.SidereonPropagationContext) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_force_model_acceleration((*C.SidereonForceModel)(pointers[0]), cCartesianState(state), contextC, &value))
		})
		return orbitVectorFromC(value), err
	})
}

func ForceModelAccelerationPartials(model *ForceModel, state CartesianState, parameters []float64, context *PropagationContext) (AccelerationPartials, error) {
	return withOrbitContext([]*positioningHandle{model.orbitHandle()}, context, func(pointers []unsafe.Pointer, contextC *C.SidereonPropagationContext) (AccelerationPartials, error) {
		var value C.SidereonAccelerationPartials
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_force_model_acceleration_partials((*C.SidereonForceModel)(pointers[0]), cCartesianState(state), (*C.double)(unsafe.Pointer(unsafe.SliceData(parameters))), C.size_t(len(parameters)), contextC, &value))
		})
		return orbitAccelerationPartialsFromC(value), err
	})
}

func ForceModelAccelerationWithParameters(model *ForceModel, state CartesianState, parameters []float64, context *PropagationContext) ([3]float64, error) {
	return withOrbitContext([]*positioningHandle{model.orbitHandle()}, context, func(pointers []unsafe.Pointer, contextC *C.SidereonPropagationContext) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_force_model_acceleration_with_parameters((*C.SidereonForceModel)(pointers[0]), cCartesianState(state), (*C.double)(unsafe.Pointer(unsafe.SliceData(parameters))), C.size_t(len(parameters)), contextC, &value))
		})
		return orbitVectorFromC(value), err
	})
}

func ForceModelParameterCount(model *ForceModel) (uint64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (uint64, error) {
		var value C.size_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_force_model_parameter_count((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return uint64(value), err
	})
}

func ForceModelParameterPartials(model *ForceModel, state CartesianState, parameters []float64, context *PropagationContext) ([][3]float64, error) {
	return withOrbitContext([]*positioningHandle{model.orbitHandle()}, context, func(pointers []unsafe.Pointer, contextC *C.SidereonPropagationContext) ([][3]float64, error) {
		raw, err := orbitReadRows("force model parameter partials", func(out *C.SidereonVector3, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_force_model_parameter_partials((*C.SidereonForceModel)(pointers[0]), cCartesianState(state), (*C.double)(unsafe.Pointer(unsafe.SliceData(parameters))), C.size_t(len(parameters)), contextC, out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonVector3) [3]float64 { return orbitVectorFromC(item) }), nil
	})
}

func J2GravityNew(muKm3S2 float64, referenceRadiusKm float64, j2 float64) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_j2_gravity_new(C.double(muKm3S2), C.double(referenceRadiusKm), C.double(j2), &value))
	})
	return newOrbitForceModel(value, err)
}

func RadiationShadowOptionsDefault() RadiationShadowOptions {
	value := C.sidereon_radiation_shadow_options_default()
	return orbitRadiationShadowOptionsFromC(value)
}

func RTNAxisAll() ([3]RTNAxis, error) {
	var raw [3]C.enum_SidereonRtnAxis
	err := callStatus(func() uint32 { return uint32(C.sidereon_rtn_axis_all(&raw[0])) })
	var value [3]RTNAxis
	for i, item := range raw {
		value[i] = RTNAxis(item)
	}
	return value, err
}

func SchwarzschildRelativityNew(muKm3S2 float64, cKmS float64) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_schwarzschild_relativity_new(C.double(muKm3S2), C.double(cKmS), &value))
	})
	return newOrbitForceModel(value, err)
}

func SolarRadiationPressureAccelerationWithSunMoon(model *ForceModel, positionKm [3]float64, sunPositionKm [3]float64, moonPositionKm [3]float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_acceleration_with_sun_moon((*C.SidereonForceModel)(pointers[0]), orbitVectorToC(positionKm), orbitVectorToC(sunPositionKm), orbitVectorToC(moonPositionKm), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func SolarRadiationPressureAreaToMassM2Kg(model *ForceModel) (float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_area_to_mass_m2_kg((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func SolarRadiationPressureAUKm(model *ForceModel) (float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_au_km((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func SolarRadiationPressureCR(model *ForceModel) (float64, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_cr((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func SolarRadiationPressureCRParameter() (string, error) {
	return orbitText("solar radiation pressure cr parameter", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_solar_radiation_pressure_cr_parameter(out, capacity, written, required)
	})
}

func SolarRadiationPressureEarthShadowModel(model *ForceModel) (uint32, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (uint32, error) {
		var value C.enum_SidereonEarthShadowModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_earth_shadow_model((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return uint32(value), err
	})
}

func SolarRadiationPressureMoonEclipseEnabled(model *ForceModel) (bool, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_moon_eclipse_enabled((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return bool(value), err
	})
}

func SolarRadiationPressureNew(cr float64, areaToMassM2Kg float64, shadow RadiationShadowOptions) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_solar_radiation_pressure_new(C.double(cr), C.double(areaToMassM2Kg), orbitRadiationShadowOptionsToC(shadow), &value))
	})
	return newOrbitForceModel(value, err)
}

func SolarRadiationPressureShadowFunction(model *ForceModel) (ShadowFunction, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (ShadowFunction, error) {
		var value C.enum_SidereonShadowFunction
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_shadow_function((*C.SidereonForceModel)(pointers[0]), &value))
		})
		return ShadowFunction(value), err
	})
}

func SolarRadiationPressureWithEarthShadowModel(model *ForceModel, earthModel uint32) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_with_earth_shadow_model((*C.SidereonForceModel)(pointers[0]), C.uint32_t(earthModel), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func SolarRadiationPressureWithMoonEclipse(model *ForceModel, enabled bool) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_with_moon_eclipse((*C.SidereonForceModel)(pointers[0]), C.bool(enabled), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func SolarRadiationPressureWithShadowFunction(model *ForceModel, shadowFunction ShadowFunction) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{model.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_solar_radiation_pressure_with_shadow_function((*C.SidereonForceModel)(pointers[0]), C.uint32_t(shadowFunction), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func SolidEarthPoleTideGravityNew() (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_solid_earth_pole_tide_gravity_new(&value)) })
	return newOrbitForceModel(value, err)
}

func SolidEarthTideGravityNew(tideSystem GravityTideSystem) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_solid_earth_tide_gravity_new(C.uint32_t(tideSystem), &value)) })
	return newOrbitForceModel(value, err)
}

func SourcedDragForceNew(bcFactorM2Kg float64, cutoffAltitudeKm float64, table *SpaceWeatherTable) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{orbitSpaceWeatherTableHandle(table)}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_sourced_drag_force_new(C.double(bcFactorM2Kg), C.double(cutoffAltitudeKm), (*C.SidereonSpaceWeatherTable)(pointers[0]), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func ThirdBodyGravityNew(sun bool, moon bool) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_third_body_gravity_new(C.bool(sun), C.bool(moon), &value)) })
	return newOrbitForceModel(value, err)
}

func TwoBodyGravityNew(muKm3S2 float64) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 { return uint32(C.sidereon_two_body_gravity_new(C.double(muKm3S2), &value)) })
	return newOrbitForceModel(value, err)
}

func ZonalGravityNew(muKm3S2 float64, referenceRadiusKm float64, maxDegree uint8, tideSystem GravityTideSystem) (*ForceModel, error) {
	var value *C.SidereonForceModel
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_zonal_gravity_new(C.double(muKm3S2), C.double(referenceRadiusKm), C.uint8_t(maxDegree), C.uint32_t(tideSystem), &value))
	})
	return newOrbitForceModel(value, err)
}
