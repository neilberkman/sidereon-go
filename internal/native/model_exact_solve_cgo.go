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

func (store *SBASCorrectionStore) SolveBroadcastV2AtExactEpoch(broadcast *BroadcastEphemeris, geo string, mode uint32, input SppInputsV2, receiveEpoch *ExactEpoch) (SPPSolution, error) {
	if store == nil || store.resource == nil || broadcast == nil || broadcast.resource == nil || receiveEpoch == nil || receiveEpoch.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSBASSolveModeValue(mode); err != nil {
		return SPPSolution{}, err
	}
	if err := validateSPPSatelliteID(geo); err != nil {
		return SPPSolution{}, err
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cinput, err := makeSppV2(input, alloc)
	if err != nil {
		return SPPSolution{}, err
	}
	models, err := makeSppModels(input.Models)
	if err != nil {
		return SPPSolution{}, err
	}
	geoToken := C.CString(geo)
	if geoToken == nil {
		return SPPSolution{}, errors.New("sidereon: unable to allocate native GEO ID")
	}
	defer C.free(unsafe.Pointer(geoToken))
	var result SPPSolution
	err = broadcast.resource.with(func(broadcastPointer unsafe.Pointer) error {
		return store.resource.with(func(storePointer unsafe.Pointer) error {
			return receiveEpoch.handle.with(func(epochPointer unsafe.Pointer) error {
				var operationErr error
				withCThread(func() {
					var solution *C.SidereonSppSolution
					status := C.sidereon_sbas_solve_broadcast_v2_with_models_at_exact_epoch((*C.SidereonBroadcastEphemeris)(broadcastPointer), (*C.SidereonSbasCorrectionStore)(storePointer), geoToken, C.uint32_t(mode), cinput, models, (*C.SidereonExactEpoch)(epochPointer), &solution)
					if err := statusErrorLocked(uint32(status)); err != nil {
						if solution != nil {
							C.sidereon_spp_solution_free(solution)
						}
						operationErr = err
						return
					}
					if solution == nil {
						operationErr = errors.New("sidereon: native SBAS exact-epoch solve returned no solution")
						return
					}
					defer C.sidereon_spp_solution_free(solution)
					result, operationErr = readSPPSolutionLocked(solution)
				})
				return operationErr
			})
		})
	})
	runtime.KeepAlive(store)
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(receiveEpoch)
	return result, err
}

func (store *SSRCorrectionStore) SolveBroadcastV2AtExactEpoch(broadcast *BroadcastEphemeris, input SppInputsV2, staleness float64, missing uint32, allowRegional bool, provider uint16, sizePolicy uint32, receiveEpoch *ExactEpoch) (SPPSolution, error) {
	if store == nil || store.resource == nil || broadcast == nil || broadcast.resource == nil || receiveEpoch == nil || receiveEpoch.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSSRMissingActionValue(missing); err != nil {
		return SPPSolution{}, err
	}
	if sizePolicy > 1 {
		return SPPSolution{}, invalidArgument("invalid SSR correction size policy")
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cinput, err := makeSppV2(input, alloc)
	if err != nil {
		return SPPSolution{}, err
	}
	models, err := makeSppModels(input.Models)
	if err != nil {
		return SPPSolution{}, err
	}
	var result SPPSolution
	err = broadcast.resource.with(func(broadcastPointer unsafe.Pointer) error {
		return store.resource.with(func(storePointer unsafe.Pointer) error {
			return receiveEpoch.handle.with(func(epochPointer unsafe.Pointer) error {
				var operationErr error
				withCThread(func() {
					var solution *C.SidereonSppSolution
					status := C.sidereon_ssr_solve_broadcast_v2_with_models_at_exact_epoch((*C.SidereonBroadcastEphemeris)(broadcastPointer), (*C.SidereonSsrCorrectionStore)(storePointer), C.double(staleness), C.uint32_t(missing), C.bool(allowRegional), C.uint16_t(provider), C.uint32_t(sizePolicy), cinput, models, (*C.SidereonExactEpoch)(epochPointer), &solution)
					if err := statusErrorLocked(uint32(status)); err != nil {
						if solution != nil {
							C.sidereon_spp_solution_free(solution)
						}
						operationErr = err
						return
					}
					if solution == nil {
						operationErr = errors.New("sidereon: native SSR exact-epoch solve returned no solution")
						return
					}
					defer C.sidereon_spp_solution_free(solution)
					result, operationErr = readSPPSolutionLocked(solution)
				})
				return operationErr
			})
		})
	})
	runtime.KeepAlive(store)
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(receiveEpoch)
	return result, err
}
