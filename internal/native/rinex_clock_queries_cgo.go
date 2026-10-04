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

func CivilToClockEpoch(scale uint32, epoch CivilDateTime) (NativeClockEpoch, bool, error) {
	civil, err := checkedClockCivil(epoch)
	if err != nil {
		return NativeClockEpoch{}, false, err
	}
	var raw C.SidereonClockEpoch
	var available C.bool
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_civil_to_clock_epoch(C.uint32_t(scale), civil.year, civil.month, civil.day, civil.hour, civil.minute, civil.second, &raw, &available))
	})
	if err != nil {
		return NativeClockEpoch{}, false, err
	}
	return clockEpochFromC(raw), bool(available), nil
}

func (clock *RinexClock) BiasAtCivil(satellite string, epoch CivilDateTime) (float64, bool, error) {
	if clock == nil || clock.resource == nil {
		return 0, false, ErrClosed
	}
	civil, err := checkedClockCivil(epoch)
	if err != nil {
		return 0, false, err
	}
	var bias C.double
	var available C.bool
	err = clock.resource.with(func(owner unsafe.Pointer) error {
		return withStringStatus(satellite, func(token *C.char) uint32 {
			return C.sidereon_rinex_clock_bias_at_civil((*C.SidereonRinexClock)(owner), token, &civil, &bias, &available)
		}, statusErrorLocked)
	})
	runtime.KeepAlive(clock)
	return float64(bias), bool(available), err
}

func (clock *RinexClock) BiasAtEpoch(satellite string, epoch NativeClockEpoch) (float64, bool, error) {
	if clock == nil || clock.resource == nil {
		return 0, false, ErrClosed
	}
	instant := clockEpochToC(epoch)
	var bias C.double
	var available C.bool
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withStringStatus(satellite, func(token *C.char) uint32 {
			return C.sidereon_rinex_clock_bias_at_epoch((*C.SidereonRinexClock)(owner), token, &instant, &bias, &available)
		}, statusErrorLocked)
	})
	runtime.KeepAlive(clock)
	return float64(bias), bool(available), err
}
