//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"time"
	"unsafe"
)

func DEEphemerisAstronomicalUnitM(ephemeris *DEEphemeris) (float64, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_astronomical_unit_m((*C.SidereonDeEphemeris)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func DEEphemerisConstantCount(ephemeris *DEEphemeris) (uint64, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (uint64, error) {
		var value C.size_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_constant_count((*C.SidereonDeEphemeris)(pointers[0]), &value))
		})
		return uint64(value), err
	})
}

func DEEphemerisCoverageTDBSeconds(ephemeris *DEEphemeris) (DECoverage, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (DECoverage, error) {
		var value C.SidereonDeCoverage
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_coverage_tdb_seconds((*C.SidereonDeEphemeris)(pointers[0]), &value))
		})
		return orbitDECoverageFromC(value), err
	})
}

func DEEphemerisEarthMoonMassRatio(ephemeris *DEEphemeris) (float64, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (float64, error) {
		var value C.double
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_earth_moon_mass_ratio((*C.SidereonDeEphemeris)(pointers[0]), &value))
		})
		return float64(value), err
	})
}

func DEEphemerisFromBytes(data []byte) (*DEEphemeris, error) {
	var value *C.SidereonDeEphemeris
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_de_ephemeris_from_bytes((*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(data))), C.size_t(len(data)), &value))
	})
	return newOrbitDEEphemeris(value, err)
}

func DEEphemerisGmM3S2(ephemeris *DEEphemeris, body DEBody) (DEGM, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (DEGM, error) {
		var value C.SidereonDeGm
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_gm_m3_s2((*C.SidereonDeEphemeris)(pointers[0]), C.uint32_t(body), &value))
		})
		return orbitDEGMFromC(value), err
	})
}

func DEEphemerisLoad(path string) (*DEEphemeris, error) {
	pathC, err := copyNativeCString(path, "path")
	if err != nil {
		return nil, err
	}
	defer C.free(pathC)
	var value *C.SidereonDeEphemeris
	err = callStatus(func() uint32 { return uint32(C.sidereon_de_ephemeris_load((*C.char)(pathC), &value)) })
	return newOrbitDEEphemeris(value, err)
}

func DEEphemerisRecordSize(ephemeris *DEEphemeris) (uint64, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (uint64, error) {
		var value C.size_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_record_size((*C.SidereonDeEphemeris)(pointers[0]), &value))
		})
		return uint64(value), err
	})
}

func DEEphemerisSunMoonAtTDBSeconds(ephemeris *DEEphemeris, epochTDBSeconds float64) (DESunMoon, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (DESunMoon, error) {
		var value C.SidereonDeSunMoon
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_sun_moon_at_tdb_seconds((*C.SidereonDeEphemeris)(pointers[0]), C.double(epochTDBSeconds), &value))
		})
		return orbitDESunMoonFromC(value), err
	})
}

func DEEphemerisSunMoonPositionsAtTDBSeconds(ephemeris *DEEphemeris, epochTDBSeconds float64) (SunMoon, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (SunMoon, error) {
		var value C.SidereonSunMoon
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_sun_moon_positions_at_tdb_seconds((*C.SidereonDeEphemeris)(pointers[0]), C.double(epochTDBSeconds), &value))
		})
		return orbitSunMoonFromC(value), err
	})
}

func DEEphemerisVersion(ephemeris *DEEphemeris) (uint32, error) {
	return withOrbitHandles([]*positioningHandle{ephemeris.orbitHandle()}, func(pointers []unsafe.Pointer) (uint32, error) {
		var value C.uint32_t
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_de_ephemeris_version((*C.SidereonDeEphemeris)(pointers[0]), &value))
		})
		return uint32(value), err
	})
}

func FrameGCRSToMeanOfDateMatrix(ts TimeScales) ([3][3]float64, error) {
	tsC := cTimeScales(ts)
	var value [3][3]float64
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_gcrs_to_mean_of_date_matrix(&tsC, (*C.double)(unsafe.Pointer(&value))))
	})
	return value, err
}

func FrameMeanOfDateToGCRSMatrix(ts TimeScales) ([3][3]float64, error) {
	tsC := cTimeScales(ts)
	var value [3][3]float64
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_mean_of_date_to_gcrs_matrix(&tsC, (*C.double)(unsafe.Pointer(&value))))
	})
	return value, err
}

func SunMoonGCRS(ttJulianCenturies float64) (SunMoon, error) {
	var value C.SidereonSunMoon
	err := callStatus(func() uint32 { return uint32(C.sidereon_sun_moon_gcrs(C.double(ttJulianCenturies), &value)) })
	return orbitSunMoonFromC(value), err
}

func SunMoonGCRSAt(epoch time.Time) (SunMoon, error) {
	epochUnixUs, err := unixMicroseconds(epoch)
	if err != nil {
		return SunMoon{}, err
	}
	var value C.SidereonSunMoon
	err = callStatus(func() uint32 { return uint32(C.sidereon_sun_moon_gcrs_at(C.int64_t(epochUnixUs), &value)) })
	return orbitSunMoonFromC(value), err
}

func SunMoonGCRSAtTDBSeconds(epochTDBSeconds float64) (SunMoon, error) {
	var value C.SidereonSunMoon
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_sun_moon_gcrs_at_tdb_seconds(C.double(epochTDBSeconds), &value))
	})
	return orbitSunMoonFromC(value), err
}

func TDBFromTTJ2000Seconds(ttJ2000S float64) float64 {
	value := C.sidereon_tdb_from_tt_j2000_seconds(C.double(ttJ2000S))
	return float64(value)
}

func TDBMinusTTSeconds(ttJ2000S float64) float64 {
	value := C.sidereon_tdb_minus_tt_seconds(C.double(ttJ2000S))
	return float64(value)
}

func TTFromTDBJ2000Seconds(tdbJ2000S float64) float64 {
	value := C.sidereon_tt_from_tdb_j2000_seconds(C.double(tdbJ2000S))
	return float64(value)
}
