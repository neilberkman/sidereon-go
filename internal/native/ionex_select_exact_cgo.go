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

func SelectIONEXAtInstant(products []*Ionex, requested NativeClockEpoch, policy StalenessPolicy) (*Ionex, NativeStalenessMetadata, IonexEpochError, error) {
	return selectIONEXAtInstantRange(products, requested, requested, policy, false)
}
func SelectIONEXOverInstantRange(products []*Ionex, start, end NativeClockEpoch, policy StalenessPolicy) (*Ionex, NativeStalenessMetadata, IonexEpochError, error) {
	return selectIONEXAtInstantRange(products, start, end, policy, true)
}
func selectIONEXAtInstantRange(products []*Ionex, start, end NativeClockEpoch, policy StalenessPolicy, overRange bool) (*Ionex, NativeStalenessMetadata, IonexEpochError, error) {
	handles := make([]*positioningHandle, len(products))
	for j, p := range products {
		if p == nil || p.handle == nil {
			return nil, NativeStalenessMetadata{}, IonexEpochError{}, ErrClosed
		}
		handles[j] = p.handle
	}
	count, err := checkedNativeSize(len(products))
	if err != nil {
		return nil, NativeStalenessMetadata{}, IonexEpochError{}, err
	}
	var selected *C.SidereonIonex
	var result *Ionex
	var metadata NativeStalenessMetadata
	var epochError IonexEpochError
	err = withPositioningHandleSet(handles, func(pointers []unsafe.Pointer) error {
		return withCThreadError(func() error {
			bytes, err := checkedNativeAllocationSize(len(pointers), unsafe.Sizeof(uintptr(0)))
			if err != nil {
				return err
			}
			var memory unsafe.Pointer
			if bytes > 0 {
				memory = C.calloc(1, C.size_t(bytes))
				if memory == nil {
					return errors.New("sidereon: unable to allocate native IONEX selection array")
				}
				defer C.free(memory)
			}
			array := unsafe.Slice((**C.SidereonIonex)(memory), len(pointers))
			for j, p := range pointers {
				array[j] = (*C.SidereonIonex)(p)
			}
			a, b := clockEpochToC(start), clockEpochToC(end)
			var cmeta C.SidereonStalenessMetadata
			var ce C.SidereonIonexEpochError
			var status C.enum_SidereonSelectionStatus
			if overRange {
				status = C.sidereon_select_ionex_over_instant_range((**C.SidereonIonex)(memory), count, &a, &b, C.SidereonStalenessPolicy{max_staleness_s: C.double(policy.MaxStalenessS)}, &selected, &cmeta, &ce)
			} else {
				status = C.sidereon_select_ionex_at_instant((**C.SidereonIonex)(memory), count, &a, C.SidereonStalenessPolicy{max_staleness_s: C.double(policy.MaxStalenessS)}, &selected, &cmeta, &ce)
			}
			epochError = epochErrorFromC(ce)
			if err := validateSelectionStatus(uint32(status)); err != nil {
				if selected != nil {
					C.sidereon_ionex_free(selected)
					selected = nil
				}
				return err
			}
			if status != C.SIDEREON_SELECTION_STATUS_OK {
				failure := selectionStatusErrorLocked(uint32(status))
				if selected != nil {
					C.sidereon_ionex_free(selected)
					selected = nil
				}
				return failure
			}
			if selected == nil {
				return errors.New("sidereon: native exact IONEX selection returned no handle")
			}
			metadata, err = stalenessMetadataFromC(cmeta)
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
			return nil
		})
	})
	for _, p := range products {
		runtime.KeepAlive(p)
	}
	if err != nil {
		if selected != nil {
			withCThread(func() { releaseIonex(unsafe.Pointer(selected)) })
		}
		return nil, NativeStalenessMetadata{}, epochError, err
	}
	if result == nil {
		return nil, NativeStalenessMetadata{}, epochError, errors.New("sidereon: exact IONEX selection returned no wrapper")
	}
	return result, metadata, epochError, nil
}
