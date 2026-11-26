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

func orbitMap[S, T any](values []S, convert func(S) T) []T {
	if values == nil {
		return nil
	}
	result := make([]T, len(values))
	for i, value := range values {
		result[i] = convert(value)
	}
	return result
}

func withOrbitHandles[T any](handles []*positioningHandle, call func([]unsafe.Pointer) (T, error)) (result T, err error) {
	err = withPositioningHandles(handles, func(pointers []unsafe.Pointer) error {
		var callErr error
		result, callErr = call(pointers)
		return callErr
	})
	return result, err
}

func withOrbitContext[T any](handles []*positioningHandle, context *PropagationContext, call func([]unsafe.Pointer, *C.SidereonPropagationContext) (T, error)) (T, error) {
	if context == nil {
		return withOrbitHandles(handles, func(pointers []unsafe.Pointer) (T, error) {
			return call(pointers, nil)
		})
	}
	de, orientation := -1, -1
	if context.DEEphemeris != nil {
		de = len(handles)
		handles = append(handles, context.DEEphemeris.orbitHandle())
	}
	if context.EarthOrientation != nil {
		orientation = len(handles)
		handles = append(handles, context.EarthOrientation.orbitHandle())
	}
	return withOrbitHandles(handles, func(pointers []unsafe.Pointer) (T, error) {
		var raw C.SidereonPropagationContext
		if de >= 0 {
			raw.de_ephemeris = (*C.SidereonDeEphemeris)(pointers[de])
		}
		if orientation >= 0 {
			raw.earth_orientation = (*C.SidereonIers2010EarthOrientationProvider)(pointers[orientation])
		}
		return call(pointers, &raw)
	})
}

func adoptOrbitHandle(pointer unsafe.Pointer, err error, release func(unsafe.Pointer)) (*positioningHandle, error) {
	if err != nil {
		if pointer != nil {
			withCThread(func() { release(pointer) })
		}
		return nil, err
	}
	if pointer == nil {
		return nil, missingNativeHandle("orbital dynamics resource")
	}
	return newPositioningHandle(pointer, release), nil
}

func orbitReadRows[T any](label string, call func(*T, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]T, error) {
	var result []T
	var err error
	withCThread(func() {
		result, err = orbitReadRowsLocked(label, call)
	})
	return result, err
}

func orbitReadRowsLocked[T any](label string, call func(*T, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]T, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := sizeTToInt(required, label+" required count")
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, label+" first-call written count"); err != nil {
		return nil, err
	}
	var row T
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(row)); err != nil {
		return nil, err
	}
	rows := make([]T, count)
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(unsafe.SliceData(rows), C.size_t(count), &written, &required))); err != nil {
		return nil, err
	}
	if _, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required)); err != nil {
		return nil, err
	}
	return rows, nil
}

func orbitText(label string, call nativeBytesCall) (string, error) {
	value, err := copyNativeBytes(label, call)
	return string(value), err
}

func orbitVectorToC(value [3]float64) C.SidereonVector3 {
	return C.SidereonVector3{xyz: [3]C.double{C.double(value[0]), C.double(value[1]), C.double(value[2])}}
}

func orbitVectorFromC(value C.SidereonVector3) [3]float64 {
	return [3]float64{float64(value.xyz[0]), float64(value.xyz[1]), float64(value.xyz[2])}
}

func orbitTimeScalesFromC(value C.SidereonTimeScales) TimeScales {
	return TimeScales{
		JDWhole: float64(value.jd_whole), UT1Fraction: float64(value.ut1_fraction),
		TTFraction: float64(value.tt_fraction), TDBFraction: float64(value.tdb_fraction),
		JDUT1: float64(value.jd_ut1), JDTT: float64(value.jd_tt), JDTDB: float64(value.jd_tdb),
		UT1Degraded: uint32(value.ut1_degraded),
	}
}

func orbitSP3Handle(value *SP3) *positioningHandle {
	if value == nil {
		return nil
	}
	return value.handle
}

func orbitSpaceWeatherTableHandle(value *SpaceWeatherTable) *positioningHandle {
	if value == nil {
		return nil
	}
	return value.handle
}

func orbitAdoptSP3(value *C.SidereonSp3, err error) (*SP3, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(value), err, releaseSP3)
	if err != nil {
		return nil, err
	}
	return &SP3{handle: handle}, nil
}

func orbitParameterStatsFromC(value C.SidereonOrbitResidualStats) OrbitResidualStats {
	return OrbitResidualStats{
		RadialRMSM:     float64(value.radial_rms_m),
		AlongRMSM:      float64(value.along_rms_m),
		CrossRMSM:      float64(value.cross_rms_m),
		RMS3DM:         float64(value.rms_3d_m),
		N:              int(value.n),
		LowSampleCount: bool(value.low_sample_count),
	}
}

func pinOrbitSlice[T any](pins *runtime.Pinner, values []T) *T {
	if len(values) == 0 {
		return nil
	}
	data := &values[0]
	pins.Pin(data)
	return data
}

func pinOrbitValue[T any](pins *runtime.Pinner, value T) *T {
	pins.Pin(&value)
	return &value
}

func orbitEarthRadiationToC(value EarthRadiationPressureOptions) C.SidereonEarthRadiationPressureOptions {
	return C.SidereonEarthRadiationPressureOptions{
		cr:                   C.double(value.CR),
		area_to_mass_m2_kg:   C.double(value.AreaToMassM2PerKg),
		moon_eclipse_enabled: C.bool(value.MoonEclipseEnabled),
	}
}
