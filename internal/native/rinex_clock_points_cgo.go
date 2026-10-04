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
	"strings"
)

func NewRinexClockFromPoints(scale uint32, points []NativeClockSatellitePoint) (*RinexClock, error) {
	rows := make([]C.SidereonClockSatellitePoint, len(points))
	for i, point := range points {
		if strings.IndexByte(point.Satellite, 0) >= 0 || len(point.Satellite) >= len(rows[i].satellite.bytes) {
			return nil, invalidArgument("RINEX clock satellite token is invalid or too long")
		}
		for j, b := range []byte(point.Satellite) {
			rows[i].satellite.bytes[j] = C.char(b)
		}
		rows[i].point.epoch = clockEpochToC(point.Point.Epoch)
		rows[i].point.bias_s = C.double(point.Point.BiasS)
		if len(point.Point.AdditionalValues) > len(rows[i].point.additional_values) {
			return nil, invalidArgument("RINEX clock point has too many additional values")
		}
		rows[i].point.additional_value_count = C.size_t(len(point.Point.AdditionalValues))
		for j, value := range point.Point.AdditionalValues {
			rows[i].point.additional_values[j] = C.double(value)
		}
	}
	var pointer *C.SidereonRinexClock
	var result *C.SidereonRinexClockResult
	err := withCThreadError(func() error {
		var ptr *C.SidereonClockSatellitePoint
		if len(rows) > 0 {
			ptr = &rows[0]
		}
		status := C.sidereon_rinex_clock_from_points(C.uint32_t(scale), ptr, C.size_t(len(rows)), &pointer, &result)
		if result != nil {
			defer C.sidereon_rinex_clock_result_free(result)
			var outcome C.SidereonRinexClockOutcome
			if e := statusErrorLocked(uint32(C.sidereon_rinex_clock_result_get_outcome(result, &outcome))); e != nil {
				return e
			}
			if failure, e := readClockWriteFailureLocked(result, outcome); e != nil {
				return e
			} else if failure != nil {
				return failure
			}
		}
		return statusErrorLocked(uint32(status))
	})
	runtime.KeepAlive(rows)
	runtime.KeepAlive(points)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_rinex_clock_free(pointer) })
		}
		return nil, err
	}
	return newRinexClock(pointer)
}
