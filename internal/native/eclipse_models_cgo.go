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

func EclipseMoonShadowFraction(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64) (float64, error) {
	var value C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_moon_shadow_fraction((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), (*C.double)(unsafe.Pointer(&moonPosKm)), &value))
	})
	return float64(value), err
}

func EclipseShadowFractionWithGradient(satPosKm [3]float64, sunPosKm [3]float64) (ShadowGradient, error) {
	var value C.SidereonShadowGradient
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_shadow_fraction_with_gradient((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), &value))
	})
	return orbitShadowGradientFromC(value), err
}

func EclipseShadowFractionWithModelAndGradient(satPosKm [3]float64, sunPosKm [3]float64, model uint32) (ShadowGradient, error) {
	var value C.SidereonShadowGradient
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_shadow_fraction_with_model_and_gradient((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), C.uint32_t(model), &value))
	})
	return orbitShadowGradientFromC(value), err
}

func EclipseShadowFractionWithModelAndMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64, model uint32) (float64, error) {
	var value C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_shadow_fraction_with_model_and_moon((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), (*C.double)(unsafe.Pointer(&moonPosKm)), C.uint32_t(model), &value))
	})
	return float64(value), err
}

func EclipseShadowFractionWithModelMoonAndGradient(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64, model uint32) (ShadowGradient, error) {
	var value C.SidereonShadowGradient
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_shadow_fraction_with_model_moon_and_gradient((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), (*C.double)(unsafe.Pointer(&moonPosKm)), C.uint32_t(model), &value))
	})
	return orbitShadowGradientFromC(value), err
}

func EclipseShadowFractionWithMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64) (float64, error) {
	var value C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_shadow_fraction_with_moon((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), (*C.double)(unsafe.Pointer(&moonPosKm)), &value))
	})
	return float64(value), err
}

func EclipseStatusWithModelAndMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64, model uint32) (uint32, error) {
	var value C.enum_SidereonEclipseStatus
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_status_with_model_and_moon((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), (*C.double)(unsafe.Pointer(&moonPosKm)), C.uint32_t(model), &value))
	})
	return uint32(value), err
}

func EclipseStatusWithMoon(satPosKm [3]float64, sunPosKm [3]float64, moonPosKm [3]float64) (uint32, error) {
	var value C.enum_SidereonEclipseStatus
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_eclipse_status_with_moon((*C.double)(unsafe.Pointer(&satPosKm)), (*C.double)(unsafe.Pointer(&sunPosKm)), (*C.double)(unsafe.Pointer(&moonPosKm)), &value))
	})
	return uint32(value), err
}

func IAU2015SolarRadiusKm() float64 {
	value := C.sidereon_iau_2015_solar_radius_km()
	return float64(value)
}
