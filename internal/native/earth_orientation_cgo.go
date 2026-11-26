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

func IERS2010EarthOrientationProviderEqual(left *IERS2010EarthOrientationProvider, right *IERS2010EarthOrientationProvider) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_iers2010_earth_orientation_provider_equal((*C.SidereonIers2010EarthOrientationProvider)(pointers[0]), (*C.SidereonIers2010EarthOrientationProvider)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func IERS2010EarthOrientationProviderNew(samples []IERS2010EopSample) (*IERS2010EarthOrientationProvider, error) {
	samplesC := orbitMap(samples, func(item IERS2010EopSample) C.SidereonIers2010EopSample { return orbitIERS2010EopSampleToC(item) })
	var value *C.SidereonIers2010EarthOrientationProvider
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_iers2010_earth_orientation_provider_new(unsafe.SliceData(samplesC), C.size_t(len(samples)), &value))
	})
	return newOrbitIERS2010EarthOrientationProvider(value, err)
}

func IERS2010EarthOrientationProviderOrientationAtTDBSeconds(provider *IERS2010EarthOrientationProvider, epochTDBSeconds float64) (EarthOrientation, error) {
	return withOrbitHandles([]*positioningHandle{provider.orbitHandle()}, func(pointers []unsafe.Pointer) (EarthOrientation, error) {
		var value C.SidereonEarthOrientation
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_iers2010_earth_orientation_provider_orientation_at_tdb_seconds((*C.SidereonIers2010EarthOrientationProvider)(pointers[0]), C.double(epochTDBSeconds), &value))
		})
		return orbitEarthOrientationFromC(value), err
	})
}

func IERS2010EarthOrientationProviderOrientationAtUTCMJD(provider *IERS2010EarthOrientationProvider, mjdUTC int32, secondsInDay float64) (EarthOrientation, error) {
	return withOrbitHandles([]*positioningHandle{provider.orbitHandle()}, func(pointers []unsafe.Pointer) (EarthOrientation, error) {
		var value C.SidereonEarthOrientation
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_iers2010_earth_orientation_provider_orientation_at_utc_mjd((*C.SidereonIers2010EarthOrientationProvider)(pointers[0]), C.int32_t(mjdUTC), C.double(secondsInDay), &value))
		})
		return orbitEarthOrientationFromC(value), err
	})
}

func IERS2010EarthOrientationProviderWithValidity(provider *IERS2010EarthOrientationProvider, permissive bool) (*IERS2010EarthOrientationProvider, error) {
	return withOrbitHandles([]*positioningHandle{provider.orbitHandle()}, func(pointers []unsafe.Pointer) (*IERS2010EarthOrientationProvider, error) {
		var value *C.SidereonIers2010EarthOrientationProvider
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_iers2010_earth_orientation_provider_with_validity((*C.SidereonIers2010EarthOrientationProvider)(pointers[0]), C.bool(permissive), &value))
		})
		return newOrbitIERS2010EarthOrientationProvider(value, err)
	})
}
