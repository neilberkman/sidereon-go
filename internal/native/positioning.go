//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <stddef.h>
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

// positioningHandle serializes ownership changes with native read calls. The
// resource lock also protects the native pointer itself, so both explicit
// Close and the cleanup callback clear the pointer before releasing it.
type positioningHandle struct {
	mu       sync.RWMutex
	resource *resource
	cleanup  runtime.Cleanup
}

func newPositioningHandle(pointer unsafe.Pointer, release func(unsafe.Pointer)) *positioningHandle {
	r := &resource{ptr: pointer, release: release}
	handle := &positioningHandle{resource: r}
	handle.cleanup = runtime.AddCleanup(handle, func(r *resource) { r.close() }, r)
	return handle
}

func (h *positioningHandle) with(fn func(unsafe.Pointer) error) error {
	if h == nil {
		return ErrClosed
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.resource == nil {
		return ErrClosed
	}
	return h.resource.with(fn)
}

// withExclusive serializes calls that may mutate a native cache with Close.
func (h *positioningHandle) withExclusive(fn func(unsafe.Pointer) error) error {
	if h == nil {
		return ErrClosed
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.resource == nil {
		return ErrClosed
	}
	return h.resource.with(fn)
}

func (h *positioningHandle) close() error {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.resource == nil {
		return nil
	}
	resource := h.resource
	h.resource = nil
	h.cleanup.Stop()
	resource.close()
	return nil
}

type positioningReadLock struct {
	h *positioningHandle
	r *resource
}

// withPositioningHandles acquires a set of positioning handles in one
// canonical address order. Duplicate handles are locked once, while the
// callback receives pointers in the caller's original order.
func withPositioningHandles(handles []*positioningHandle, fn func([]unsafe.Pointer) error) error {
	unique := make([]*positioningHandle, 0, len(handles))
	seen := make(map[*positioningHandle]struct{}, len(handles))
	for _, handle := range handles {
		if handle == nil {
			return ErrClosed
		}
		if _, ok := seen[handle]; ok {
			continue
		}
		seen[handle] = struct{}{}
		unique = append(unique, handle)
	}
	sort.Slice(unique, func(i, j int) bool {
		return uintptr(unsafe.Pointer(unique[i])) < uintptr(unsafe.Pointer(unique[j]))
	})
	locked := make([]positioningReadLock, 0, len(unique))
	unlock := func() {
		for i := len(locked) - 1; i >= 0; i-- {
			locked[i].r.mu.RUnlock()
			locked[i].h.mu.RUnlock()
		}
	}
	for _, handle := range unique {
		handle.mu.RLock()
		resource := handle.resource
		if resource == nil {
			handle.mu.RUnlock()
			unlock()
			return ErrClosed
		}
		resource.mu.RLock()
		if resource.ptr == nil {
			resource.mu.RUnlock()
			handle.mu.RUnlock()
			unlock()
			return ErrClosed
		}
		locked = append(locked, positioningReadLock{h: handle, r: resource})
	}
	byHandle := make(map[*positioningHandle]unsafe.Pointer, len(unique))
	for _, item := range locked {
		byHandle[item.h] = item.r.ptr
	}
	pointers := make([]unsafe.Pointer, len(handles))
	for i, handle := range handles {
		pointers[i] = byHandle[handle]
	}
	defer unlock()
	return fn(pointers)
}

func withPositioningPair(first, second *positioningHandle, fn func(unsafe.Pointer, unsafe.Pointer) error) error {
	return withPositioningHandles([]*positioningHandle{first, second}, func(pointers []unsafe.Pointer) error {
		return fn(pointers[0], pointers[1])
	})
}

func checkedNativeCount(value uint64) (int, error) {
	maxInt := uint64(^uint(0) >> 1)
	if value > maxInt {
		return 0, errors.New("sidereon: native result is too large")
	}
	return int(value), nil
}

func checkedNativeProduct(value, factor int, label string) (int, error) {
	if value < 0 || factor < 0 {
		return 0, fmt.Errorf("sidereon: native %s has a negative shape", label)
	}
	if factor != 0 && value > int(^uint(0)>>1)/factor {
		return 0, fmt.Errorf("sidereon: native %s shape overflows int", label)
	}
	return value * factor, nil
}

func invalidArgument(detail string) error {
	return &StatusError{
		Code:   int(C.SIDEREON_STATUS_INVALID_ARGUMENT),
		Text:   "invalid argument",
		Detail: detail,
	}
}

func missingNativeHandle(operation string) error {
	return fmt.Errorf("sidereon: native %s returned no handle", operation)
}

// validateTwoPassCounts accepts a native copy only when both reported counts
// fit the allocation and agree with the initial size query. A successful copy
// with a query-sized buffer must not produce a partial or ambiguous result.
func validateTwoPassCounts(label string, capacity, expected int, written, required uint64) (int, error) {
	writtenCount, err := checkedNativeCount(written)
	if err != nil {
		return 0, fmt.Errorf("sidereon: native %s written count: %w", label, err)
	}
	requiredCount, err := checkedNativeCount(required)
	if err != nil {
		return 0, fmt.Errorf("sidereon: native %s required count: %w", label, err)
	}
	if capacity < 0 || expected < 0 || expected > capacity {
		return 0, fmt.Errorf("sidereon: invalid native %s allocation contract", label)
	}
	if requiredCount > capacity {
		return 0, fmt.Errorf("sidereon: native %s required count %d exceeds allocated capacity %d", label, requiredCount, capacity)
	}
	if requiredCount != expected {
		return 0, fmt.Errorf("sidereon: native %s required count changed from %d to %d", label, expected, requiredCount)
	}
	if writtenCount > capacity {
		return 0, fmt.Errorf("sidereon: native %s wrote %d entries into capacity %d", label, writtenCount, capacity)
	}
	if writtenCount != requiredCount {
		return 0, fmt.Errorf("sidereon: native %s wrote %d entries, required %d", label, writtenCount, requiredCount)
	}
	return writtenCount, nil
}

// validateNativeOutput applies the common query/copy contract to a buffer
// sized from the first call's required count.
func validateNativeOutput(label string, capacity int, written, required uint64) (int, error) {
	return validateTwoPassCounts(label, capacity, capacity, written, required)
}

// validateNativeQuery checks the count pair returned by a NULL/zero-length
// query. The query must write no elements, while required is the allocation
// count for the subsequent copy.
func validateNativeQuery(label string, written, required uint64) (int, error) {
	writtenCount, err := checkedNativeCount(written)
	if err != nil {
		return 0, fmt.Errorf("sidereon: native %s written count: %w", label, err)
	}
	requiredCount, err := checkedNativeCount(required)
	if err != nil {
		return 0, fmt.Errorf("sidereon: native %s required count: %w", label, err)
	}
	if writtenCount != 0 {
		return 0, fmt.Errorf("sidereon: native %s query wrote %d entries", label, writtenCount)
	}
	return requiredCount, nil
}

func checkedNativeAllocationSize(count int, elementSize uintptr) (uintptr, error) {
	if count < 0 || elementSize == 0 {
		return 0, errors.New("sidereon: invalid native allocation size")
	}

	maxGoInt := uint64(^uint(0) >> 1)
	maxSizeT := uint64(^C.size_t(0))
	maxBytes := maxGoInt
	if maxSizeT < maxBytes {
		maxBytes = maxSizeT
	}
	elementSize64 := uint64(elementSize)
	if elementSize64 > maxBytes || uint64(count) > maxBytes/elementSize64 {
		return 0, errors.New("sidereon: native allocation size overflows")
	}
	return uintptr(uint64(count) * elementSize64), nil
}

func checkedNativeSize(count int) (C.size_t, error) {
	if _, err := checkedNativeAllocationSize(count, 1); err != nil {
		return 0, err
	}
	return C.size_t(count), nil
}

func rejectEmbeddedNUL(value, name string) error {
	if strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("sidereon: %s contains an embedded NUL byte", name)
	}
	return nil
}
