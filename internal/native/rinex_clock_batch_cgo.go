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
	"unsafe"
)

// RemoveRecords removes several record indices atomically. Duplicate indices
// are accepted by the native API and remove a row only once.
func (clock *RinexClock) RemoveRecords(indices []int) (int, error) {
	if clock == nil || clock.resource == nil {
		return 0, ErrClosed
	}
	for _, index := range indices {
		if index < 0 {
			return 0, errNegativeIndex
		}
	}
	var values unsafe.Pointer
	if len(indices) > 0 {
		size, err := checkedNativeAllocationSize(len(indices), unsafe.Sizeof(C.size_t(0)))
		if err != nil {
			return 0, err
		}
		values = C.malloc(C.size_t(size))
		if values == nil {
			return 0, invalidArgument("unable to allocate RINEX clock removal indices")
		}
		rows := unsafe.Slice((*C.size_t)(values), len(indices))
		for i, index := range indices {
			rows[i] = C.size_t(index)
		}
		defer C.free(values)
	}
	var removed C.size_t
	err := clock.resource.withExclusive(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonRinexClockResult
			status := C.sidereon_rinex_clock_remove_records((*C.SidereonRinexClock)(owner), (*C.size_t)(values), C.size_t(len(indices)), &removed, &result)
			return finishClockMutationLocked(status, result)
		})
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return 0, err
	}
	return sizeTToInt(removed, "RINEX clock records removed")
}

// SetRecordsValues applies a set of row replacements atomically; the native
// result refuses the entire operation when any row is invalid.
func (clock *RinexClock) SetRecordsValues(edits []NativeClockRecordValuesEdit) (int, error) {
	if clock == nil || clock.resource == nil {
		return 0, ErrClosed
	}
	for _, edit := range edits {
		if edit.Index < 0 {
			return 0, errNegativeIndex
		}
		if len(edit.Values) == 0 || len(edit.Values) > int(C.SIDEREON_CLOCK_MAX_VALUES) {
			return 0, invalidArgument("RINEX clock edit must contain one through six values")
		}
	}
	var pointer unsafe.Pointer
	if len(edits) > 0 {
		size, err := checkedNativeAllocationSize(len(edits), unsafe.Sizeof(C.SidereonClockRecordValues{}))
		if err != nil {
			return 0, err
		}
		pointer = C.malloc(C.size_t(size))
		if pointer == nil {
			return 0, invalidArgument("unable to allocate RINEX clock value edits")
		}
		rows := unsafe.Slice((*C.SidereonClockRecordValues)(pointer), len(edits))
		for i, edit := range edits {
			rows[i].index = C.size_t(edit.Index)
			rows[i].value_count = C.size_t(len(edit.Values))
			for j, value := range edit.Values {
				rows[i].values[j] = C.double(value)
			}
		}
		defer C.free(pointer)
	}
	var edited C.size_t
	err := clock.resource.withExclusive(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonRinexClockResult
			status := C.sidereon_rinex_clock_set_records_values((*C.SidereonRinexClock)(owner), (*C.SidereonClockRecordValues)(pointer), C.size_t(len(edits)), &edited, &result)
			return finishClockMutationLocked(status, result)
		})
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return 0, err
	}
	return sizeTToInt(edited, "RINEX clock records edited")
}
