//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import "unsafe"

func SiderealOrbitRepeatLag(b *BroadcastEphemeris, satellite string, nearEpochJ2000S float64) (float64, error) {
	if b == nil || b.resource == nil {
		return 0, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "broadcast satellite ID"); err != nil {
		return 0, err
	}
	if satellite == "" {
		return 0, invalidArgument("broadcast satellite ID must not be empty")
	}
	var result C.double
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withString(satellite, func(value *C.char) uint32 {
			return C.sidereon_sidereal_orbit_repeat_lag((*C.SidereonBroadcastEphemeris)(pointer), value, C.double(nearEpochJ2000S), &result)
		})
	})
	return float64(result), err
}
