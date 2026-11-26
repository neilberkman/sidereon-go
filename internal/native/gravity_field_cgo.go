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

func ICGEMGravityFieldCoefficientAt(field *ICGEMGravityField, degree uint16, order uint16, epochTTSeconds float64) (OptionalHarmonicCoefficient, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (OptionalHarmonicCoefficient, error) {
		var value C.SidereonOptionalHarmonicCoefficient
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_coefficient_at((*C.SidereonIcgemGravityField)(pointers[0]), C.uint16_t(degree), C.uint16_t(order), C.double(epochTTSeconds), &value))
		})
		return orbitOptionalHarmonicCoefficientFromC(value), err
	})
}

func ICGEMGravityFieldEqual(left *ICGEMGravityField, right *ICGEMGravityField) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_equal((*C.SidereonIcgemGravityField)(pointers[0]), (*C.SidereonIcgemGravityField)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func ICGEMGravityFieldGravity(field *ICGEMGravityField, maxDegree uint16, maxOrder uint16) (*SphericalHarmonicGravity, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (*SphericalHarmonicGravity, error) {
		var value *C.SidereonSphericalHarmonicGravity
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_gravity((*C.SidereonIcgemGravityField)(pointers[0]), C.uint16_t(maxDegree), C.uint16_t(maxOrder), &value))
		})
		return newOrbitSphericalHarmonicGravity(value, err)
	})
}

func ICGEMGravityFieldGravityInTideSystem(field *ICGEMGravityField, maxDegree uint16, maxOrder uint16, tideSystem GravityTideSystem) (*SphericalHarmonicGravity, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (*SphericalHarmonicGravity, error) {
		var value *C.SidereonSphericalHarmonicGravity
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_gravity_in_tide_system((*C.SidereonIcgemGravityField)(pointers[0]), C.uint16_t(maxDegree), C.uint16_t(maxOrder), C.uint32_t(tideSystem), &value))
		})
		return newOrbitSphericalHarmonicGravity(value, err)
	})
}

func ICGEMGravityFieldHeader(field *ICGEMGravityField) (ICGEMHeader, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (ICGEMHeader, error) {
		var value C.SidereonIcgemHeader
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_header((*C.SidereonIcgemGravityField)(pointers[0]), &value))
		})
		return orbitICGEMHeaderFromC(value), err
	})
}

func ICGEMGravityFieldIsTimeVariable(field *ICGEMGravityField) (bool, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_is_time_variable((*C.SidereonIcgemGravityField)(pointers[0]), &value))
		})
		return bool(value), err
	})
}

func ICGEMGravityFieldModelName(field *ICGEMGravityField) (string, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (string, error) {
		return orbitText("icgem gravity field model name", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_icgem_gravity_field_model_name((*C.SidereonIcgemGravityField)(pointers[0]), out, capacity, written, required)
		})
	})
}

func ICGEMGravityFieldParse(data []byte) (*ICGEMGravityField, error) {
	var value *C.SidereonIcgemGravityField
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_icgem_gravity_field_parse((*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), &value))
	})
	return newOrbitICGEMGravityField(value, err)
}

func ICGEMGravityFieldParseTruncated(data []byte, maxDegree uint16, maxOrder uint16) (*ICGEMGravityField, error) {
	var value *C.SidereonIcgemGravityField
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_icgem_gravity_field_parse_truncated((*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), C.uint16_t(maxDegree), C.uint16_t(maxOrder), &value))
	})
	return newOrbitICGEMGravityField(value, err)
}

func ICGEMGravityFieldStaticCoefficient(field *ICGEMGravityField, degree uint16, order uint16) (OptionalHarmonicCoefficient, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (OptionalHarmonicCoefficient, error) {
		var value C.SidereonOptionalHarmonicCoefficient
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_static_coefficient((*C.SidereonIcgemGravityField)(pointers[0]), C.uint16_t(degree), C.uint16_t(order), &value))
		})
		return orbitOptionalHarmonicCoefficientFromC(value), err
	})
}

func ICGEMGravityFieldStaticSigmas(field *ICGEMGravityField, degree uint16, order uint16) (ICGEMSigmas, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (ICGEMSigmas, error) {
		var value C.SidereonIcgemSigmas
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_static_sigmas((*C.SidereonIcgemGravityField)(pointers[0]), C.uint16_t(degree), C.uint16_t(order), &value))
		})
		return orbitICGEMSigmasFromC(value), err
	})
}

func ICGEMGravityFieldStoredMaxDegree(field *ICGEMGravityField) (uint16, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (uint16, error) {
		var value C.uint16_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_stored_max_degree((*C.SidereonIcgemGravityField)(pointers[0]), &value))
		})
		return uint16(value), err
	})
}

func ICGEMGravityFieldStoredMaxOrder(field *ICGEMGravityField) (uint16, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (uint16, error) {
		var value C.uint16_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_stored_max_order((*C.SidereonIcgemGravityField)(pointers[0]), &value))
		})
		return uint16(value), err
	})
}

func ICGEMGravityFieldTimeVariableTerms(field *ICGEMGravityField) ([]ICGEMTimeVariableTerm, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) ([]ICGEMTimeVariableTerm, error) {
		raw, err := orbitReadRows("icgem gravity field time variable terms", func(out *C.SidereonIcgemTimeVariableTerm, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_icgem_gravity_field_time_variable_terms((*C.SidereonIcgemGravityField)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonIcgemTimeVariableTerm) ICGEMTimeVariableTerm {
			return orbitICGEMTimeVariableTermFromC(item)
		}), nil
	})
}

func ICGEMGravityFieldWithAssumedTideSystem(field *ICGEMGravityField, tideSystem GravityTideSystem) (*ICGEMGravityField, error) {
	return withOrbitHandles([]*positioningHandle{field.orbitHandle()}, func(pointers []unsafe.Pointer) (*ICGEMGravityField, error) {
		var value *C.SidereonIcgemGravityField
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_icgem_gravity_field_with_assumed_tide_system((*C.SidereonIcgemGravityField)(pointers[0]), C.uint32_t(tideSystem), &value))
		})
		return newOrbitICGEMGravityField(value, err)
	})
}

func SphericalHarmonicGravityAsForceModel(gravity *SphericalHarmonicGravity) (*ForceModel, error) {
	return withOrbitHandles([]*positioningHandle{gravity.orbitHandle()}, func(pointers []unsafe.Pointer) (*ForceModel, error) {
		var value *C.SidereonForceModel
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_as_force_model((*C.SidereonSphericalHarmonicGravity)(pointers[0]), &value))
		})
		return newOrbitForceModel(value, err)
	})
}

func SphericalHarmonicGravityBodyFixedAccelerationAtKmS2(gravity *SphericalHarmonicGravity, positionKm [3]float64, epochTTSeconds float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{gravity.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_body_fixed_acceleration_at_km_s2((*C.SidereonSphericalHarmonicGravity)(pointers[0]), orbitVectorToC(positionKm), C.double(epochTTSeconds), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func SphericalHarmonicGravityBodyFixedAccelerationKmS2(gravity *SphericalHarmonicGravity, positionKm [3]float64) ([3]float64, error) {
	return withOrbitHandles([]*positioningHandle{gravity.orbitHandle()}, func(pointers []unsafe.Pointer) ([3]float64, error) {
		var value C.SidereonVector3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_body_fixed_acceleration_km_s2((*C.SidereonSphericalHarmonicGravity)(pointers[0]), orbitVectorToC(positionKm), &value))
		})
		return orbitVectorFromC(value), err
	})
}

func SphericalHarmonicGravityBodyFixedPotentialAtKm2S2(gravity *SphericalHarmonicGravity, positionKm [3]float64, epochTTSeconds float64) (float64, error) {
	return withOrbitHandles([]*positioningHandle{gravity.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_body_fixed_potential_at_km2_s2((*C.SidereonSphericalHarmonicGravity)(pointers[0]), orbitVectorToC(positionKm), C.double(epochTTSeconds), &value))
		})
		return float64(value), err
	})
}

func SphericalHarmonicGravityBodyFixedPotentialKm2S2(gravity *SphericalHarmonicGravity, positionKm [3]float64) (float64, error) {
	return withOrbitHandles([]*positioningHandle{gravity.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_body_fixed_potential_km2_s2((*C.SidereonSphericalHarmonicGravity)(pointers[0]), orbitVectorToC(positionKm), &value))
		})
		return float64(value), err
	})
}

func SphericalHarmonicGravityCoefficientAt(gravity *SphericalHarmonicGravity, degree uint16, order uint16, epochTTSeconds float64) (OptionalHarmonicCoefficient, error) {
	return withOrbitHandles([]*positioningHandle{gravity.orbitHandle()}, func(pointers []unsafe.Pointer) (OptionalHarmonicCoefficient, error) {
		var value C.SidereonOptionalHarmonicCoefficient
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_coefficient_at((*C.SidereonSphericalHarmonicGravity)(pointers[0]), C.uint16_t(degree), C.uint16_t(order), C.double(epochTTSeconds), &value))
		})
		return orbitOptionalHarmonicCoefficientFromC(value), err
	})
}

func SphericalHarmonicGravityEgm96(maxDegree uint16, maxOrder uint16) (*SphericalHarmonicGravity, error) {
	var value *C.SidereonSphericalHarmonicGravity
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_spherical_harmonic_gravity_egm96(C.uint16_t(maxDegree), C.uint16_t(maxOrder), &value))
	})
	return newOrbitSphericalHarmonicGravity(value, err)
}

func SphericalHarmonicGravityEqual(left *SphericalHarmonicGravity, right *SphericalHarmonicGravity) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_equal((*C.SidereonSphericalHarmonicGravity)(pointers[0]), (*C.SidereonSphericalHarmonicGravity)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func SphericalHarmonicGravityIsTimeVariable(gravity *SphericalHarmonicGravity) (bool, error) {
	return withOrbitHandles([]*positioningHandle{gravity.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_spherical_harmonic_gravity_is_time_variable((*C.SidereonSphericalHarmonicGravity)(pointers[0]), &value))
		})
		return bool(value), err
	})
}
