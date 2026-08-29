//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

func newIonexFromPointer(pointer *C.SidereonIonex) (*Ionex, error) {
	if pointer == nil {
		return nil, missingNativeHandle("IONEX selection")
	}
	return &Ionex{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseIonex)}, nil
}

// SelectIONEX selects a product usable at one integer J2000 epoch.
func SelectIONEX(products []*Ionex, requested int64, policy StalenessPolicy) (*Ionex, NativeStalenessMetadata, error) {
	return selectIONEX(products, requested, requested, policy, false)
}

// SelectIONEXOverRange selects a product usable across an integer J2000 range.
func SelectIONEXOverRange(products []*Ionex, start, end int64, policy StalenessPolicy) (*Ionex, NativeStalenessMetadata, error) {
	return selectIONEX(products, start, end, policy, true)
}

func selectIONEX(products []*Ionex, start, end int64, policy StalenessPolicy, overRange bool) (*Ionex, NativeStalenessMetadata, error) {
	handles := make([]*positioningHandle, len(products))
	for index, product := range products {
		if product == nil || product.handle == nil {
			return nil, NativeStalenessMetadata{}, ErrClosed
		}
		handles[index] = product.handle
	}
	productCount, err := checkedNativeSize(len(products))
	if err != nil {
		return nil, NativeStalenessMetadata{}, err
	}
	var selected *C.SidereonIonex
	var result *Ionex
	var metadata NativeStalenessMetadata
	err = withPositioningHandleSet(handles, func(pointers []unsafe.Pointer) error {
		return withCThreadError(func() error {
			arrayBytes, err := checkedNativeAllocationSize(len(pointers), unsafe.Sizeof(uintptr(0)))
			if err != nil {
				return err
			}
			var memory unsafe.Pointer
			if arrayBytes != 0 {
				memory = C.calloc(1, C.size_t(arrayBytes))
				if memory == nil {
					return errors.New("sidereon: unable to allocate native IONEX selection array")
				}
				defer C.free(memory)
			}
			array := unsafe.Slice((**C.SidereonIonex)(memory), len(pointers))
			for index, pointer := range pointers {
				array[index] = (*C.SidereonIonex)(pointer)
			}
			var cMetadata C.SidereonStalenessMetadata
			var status C.SidereonSelectionStatus
			if overRange {
				status = C.SidereonSelectionStatus(C.sidereon_select_ionex_over_range((**C.SidereonIonex)(memory), productCount, C.int64_t(start), C.int64_t(end), C.SidereonStalenessPolicy{max_staleness_s: C.double(policy.MaxStalenessS)}, &selected, &cMetadata))
			} else {
				status = C.SidereonSelectionStatus(C.sidereon_select_ionex((**C.SidereonIonex)(memory), productCount, C.int64_t(start), C.SidereonStalenessPolicy{max_staleness_s: C.double(policy.MaxStalenessS)}, &selected, &cMetadata))
			}
			if err := validateSelectionStatus(uint32(status)); err != nil {
				if selected != nil {
					C.sidereon_ionex_free(selected)
					selected = nil
				}
				return err
			}
			if status != C.SIDEREON_SELECTION_STATUS_OK {
				err := selectionStatusErrorLocked(uint32(status))
				if selected != nil {
					C.sidereon_ionex_free(selected)
					selected = nil
				}
				return err
			}
			if selected == nil {
				return errors.New("sidereon: native IONEX selection returned no handle")
			}
			converted, err := stalenessMetadataFromC(cMetadata)
			if err != nil {
				C.sidereon_ionex_free(selected)
				selected = nil
				return err
			}
			result, err = newIonexFromPointer(selected)
			if err != nil {
				C.sidereon_ionex_free(selected)
				selected = nil
				return err
			}
			selected = nil
			metadata = converted
			return nil
		})
	})
	for _, product := range products {
		runtime.KeepAlive(product)
	}
	if err != nil {
		if selected != nil {
			withCThread(func() { releaseIonex(unsafe.Pointer(selected)) })
		}
		return nil, NativeStalenessMetadata{}, err
	}
	if result == nil {
		return nil, NativeStalenessMetadata{}, errors.New("sidereon: native IONEX selection returned no wrapper")
	}
	return result, metadata, nil
}
