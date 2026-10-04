//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

func (i *Ionex) SlantDelays(requests []IonexSlantRequest) ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	count, err := checkedNativeSize(len(requests))
	if err != nil {
		return nil, err
	}
	values := make([]float64, len(requests))
	var requestMemory, outputMemory unsafe.Pointer
	if len(requests) != 0 {
		requestBytes, err := checkedNativeAllocationSize(len(requests), unsafe.Sizeof(C.SidereonIonexSlantRequest{}))
		if err != nil {
			return nil, err
		}
		requestMemory = C.calloc(1, C.size_t(requestBytes))
		if requestMemory == nil {
			return nil, errors.New("sidereon: unable to allocate IONEX slant requests")
		}
		defer C.free(requestMemory)
		out := unsafe.Slice((*C.SidereonIonexSlantRequest)(requestMemory), len(requests))
		for index, value := range requests {
			out[index] = C.SidereonIonexSlantRequest{lat_deg: C.double(value.LatDeg), lon_deg: C.double(value.LonDeg), azimuth_deg: C.double(value.AzimuthDeg), elevation_deg: C.double(value.ElevationDeg), epoch_j2000_s: C.int64_t(value.EpochJ2000S), frequency_hz: C.double(value.FrequencyHz)}
		}
		outputBytes, err := checkedNativeAllocationSize(len(requests), unsafe.Sizeof(C.double(0)))
		if err != nil {
			return nil, err
		}
		outputMemory = C.calloc(1, C.size_t(outputBytes))
		if outputMemory == nil {
			return nil, errors.New("sidereon: unable to allocate IONEX slant delays")
		}
		defer C.free(outputMemory)
	}
	err = i.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			status := C.sidereon_ionex_slant_delays((*C.SidereonIonex)(pointer), (*C.SidereonIonexSlantRequest)(requestMemory), count, (*C.double)(outputMemory))
			callErr := statusErrorLocked(uint32(status))
			if len(values) > 0 {
				out := unsafe.Slice((*C.double)(outputMemory), len(values))
				for index := range values {
					values[index] = float64(out[index])
				}
			}
			return callErr
		})
	})
	runtime.KeepAlive(requests)
	runtime.KeepAlive(i)
	return values, err
}
