//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
enum SidereonStatus sidereon_solve_rtk_arc_v2(
    const SidereonRtkArcEpochV2 *, size_t,
    const struct SidereonRtkArcConfig *, struct SidereonRtkArcSolution **);
enum SidereonStatus sidereon_solve_static_rtk_arc_v2(
    const SidereonRtkArcEpochV2 *, size_t,
    const struct SidereonRtkStaticArcConfig *, struct SidereonRtkStaticArcSolution **);
enum SidereonStatus sidereon_fix_wide_lane_rtk_arc_v2(
    const SidereonRtkDualFrequencyArcEpochV2 *, size_t,
    const struct SidereonRtkWideLaneArcConfig *, struct SidereonRtkWideLaneArcSolution **);
enum SidereonStatus sidereon_prepare_ionosphere_free_rtk_arc_v2(
    const SidereonRtkDualFrequencyArcEpochV2 *, size_t,
    const struct SidereonRtkWideLaneCycle *, size_t,
    const struct SidereonRtkIonosphereFreeArcConfig *,
    struct SidereonRtkIonosphereFreeArcSolution **);
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func withRtkExactEpochGroups(groups [][]*ExactEpoch, fn func([][]*C.SidereonExactEpoch) error) error {
	handles := make([]*positioningHandle, 0)
	indices := make([][]int, len(groups))
	for groupIndex, group := range groups {
		indices[groupIndex] = make([]int, len(group))
		for entryIndex, epoch := range group {
			indices[groupIndex][entryIndex] = -1
			if epoch == nil {
				continue
			}
			if epoch.handle == nil {
				return ErrClosed
			}
			indices[groupIndex][entryIndex] = len(handles)
			handles = append(handles, epoch.handle)
		}
	}
	return withPositioningHandles(handles, func(rawPointers []unsafe.Pointer) error {
		pointers := make([][]*C.SidereonExactEpoch, len(groups))
		for groupIndex, group := range groups {
			pointers[groupIndex] = make([]*C.SidereonExactEpoch, len(group))
			for entryIndex, pointerIndex := range indices[groupIndex] {
				if pointerIndex >= 0 {
					pointers[groupIndex][entryIndex] = (*C.SidereonExactEpoch)(rawPointers[pointerIndex])
				}
			}
		}
		return fn(pointers)
	})
}

func copyRtkArcEpochInputsV2(values []RtkArcEpochInput, exactEpochs []*C.SidereonExactEpoch, alloc *cRtkAlloc) (*C.SidereonRtkArcEpochV2, C.size_t, error) {
	legacyRows, count, err := copyRtkArcEpochInputs(values, alloc)
	if err != nil {
		return nil, 0, err
	}
	if len(values) != len(exactEpochs) {
		return nil, 0, invalidArgument("RTK exact prediction-epoch count does not match epoch count")
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkArcEpochV2{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK V2 arc epochs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkArcEpochV2)(memory), len(values))
	legacy := unsafe.Slice(legacyRows, len(values))
	for index, epoch := range exactEpochs {
		rows[index] = C.SidereonRtkArcEpochV2{legacy: legacy[index], has_prediction_epoch: C.bool(epoch != nil), prediction_epoch: epoch}
	}
	return &rows[0], count, nil
}

func copyRtkDualFrequencyEpochsV2(values []RtkDualFrequencyArcEpochInput, exactEpochs [][]*C.SidereonExactEpoch, alloc *cRtkAlloc) (*C.SidereonRtkDualFrequencyArcEpochV2, C.size_t, error) {
	legacyRows, count, err := copyRtkDualFrequencyEpochs(values, alloc)
	if err != nil {
		return nil, 0, err
	}
	if len(values) != len(exactEpochs) {
		return nil, 0, invalidArgument("RTK exact epoch group count does not match epoch count")
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkDualFrequencyArcEpochV2{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK V2 dual-frequency arc epochs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkDualFrequencyArcEpochV2)(memory), len(values))
	legacy := unsafe.Slice(legacyRows, len(values))
	for index, group := range exactEpochs {
		if len(group) != 2 {
			return nil, 0, invalidArgument("RTK V2 dual-frequency epoch requires gap and prediction slots")
		}
		rows[index] = C.SidereonRtkDualFrequencyArcEpochV2{
			legacy: legacy[index], has_gap_epoch: C.bool(group[0] != nil), gap_epoch: group[0],
			has_prediction_epoch: C.bool(group[1] != nil), prediction_epoch: group[1],
		}
	}
	return &rows[0], count, nil
}

func SolveRtkArcV2(epochs []RtkArcEpochInput, config RtkArcConfigInput) (*RtkArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := copyRtkArcConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	groups := make([][]*ExactEpoch, len(epochs))
	for index, epoch := range epochs {
		groups[index] = []*ExactEpoch{epoch.PredictionEpoch}
	}
	var pointer *C.SidereonRtkArcSolution
	err = withRtkExactEpochGroups(groups, func(exact [][]*C.SidereonExactEpoch) error {
		exactEpochs := make([]*C.SidereonExactEpoch, len(exact))
		for index := range exact {
			exactEpochs[index] = exact[index][0]
		}
		cEpochs, count, copyErr := copyRtkArcEpochInputsV2(epochs, exactEpochs, alloc)
		if copyErr != nil {
			return copyErr
		}
		withCThread(func() {
			status := C.sidereon_solve_rtk_arc_v2(cEpochs, count, cConfig, &pointer)
			copyErr = statusErrorLocked(uint32(status))
			if copyErr != nil && pointer != nil {
				C.sidereon_rtk_arc_solution_free(pointer)
				pointer = nil
			}
		})
		return copyErr
	})
	runtime.KeepAlive(epochs)
	if err != nil {
		return nil, err
	}
	return newRtkArcSolution(pointer)
}

func SolveStaticRtkArcV2(epochs []RtkArcEpochInput, config RtkStaticArcConfigInput) (*RtkStaticArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := copyRtkStaticArcConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	groups := make([][]*ExactEpoch, len(epochs))
	for index, epoch := range epochs {
		groups[index] = []*ExactEpoch{epoch.PredictionEpoch}
	}
	var pointer *C.SidereonRtkStaticArcSolution
	err = withRtkExactEpochGroups(groups, func(exact [][]*C.SidereonExactEpoch) error {
		exactEpochs := make([]*C.SidereonExactEpoch, len(exact))
		for index := range exact {
			exactEpochs[index] = exact[index][0]
		}
		cEpochs, count, copyErr := copyRtkArcEpochInputsV2(epochs, exactEpochs, alloc)
		if copyErr != nil {
			return copyErr
		}
		withCThread(func() {
			status := C.sidereon_solve_static_rtk_arc_v2(cEpochs, count, cConfig, &pointer)
			copyErr = statusErrorLocked(uint32(status))
			if copyErr != nil && pointer != nil {
				C.sidereon_rtk_static_arc_solution_free(pointer)
				pointer = nil
			}
		})
		return copyErr
	})
	runtime.KeepAlive(epochs)
	if err != nil {
		return nil, err
	}
	return newRtkStaticArcSolution(pointer)
}

func FixWideLaneRtkArcV2(epochs []RtkDualFrequencyArcEpochInput, config RtkWideLaneConfigInput) (*RtkWideLaneArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := copyRtkWideLaneConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	groups := make([][]*ExactEpoch, len(epochs))
	for index, epoch := range epochs {
		groups[index] = []*ExactEpoch{epoch.GapEpoch, epoch.PredictionEpoch}
	}
	var pointer *C.SidereonRtkWideLaneArcSolution
	err = withRtkExactEpochGroups(groups, func(exact [][]*C.SidereonExactEpoch) error {
		cEpochs, count, copyErr := copyRtkDualFrequencyEpochsV2(epochs, exact, alloc)
		if copyErr != nil {
			return copyErr
		}
		withCThread(func() {
			status := C.sidereon_fix_wide_lane_rtk_arc_v2(cEpochs, count, cConfig, &pointer)
			copyErr = statusErrorLocked(uint32(status))
			if copyErr != nil && pointer != nil {
				C.sidereon_rtk_wide_lane_arc_solution_free(pointer)
				pointer = nil
			}
		})
		return copyErr
	})
	runtime.KeepAlive(epochs)
	if err != nil {
		return nil, err
	}
	return newRtkWideLaneArcSolution(pointer)
}

func PrepareIonosphereFreeRtkArcV2(epochs []RtkDualFrequencyArcEpochInput, cycles []RtkWideLaneCycleInput, config RtkIonosphereFreeConfigInput) (*RtkIonosphereFreeArcSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := copyRtkIonosphereFreeConfig(config, alloc)
	if err != nil {
		return nil, err
	}
	cCycles, cycleCount, err := copyRtkWideLaneCycles(cycles, alloc)
	if err != nil {
		return nil, err
	}
	groups := make([][]*ExactEpoch, len(epochs))
	for index, epoch := range epochs {
		groups[index] = []*ExactEpoch{epoch.GapEpoch, epoch.PredictionEpoch}
	}
	var pointer *C.SidereonRtkIonosphereFreeArcSolution
	err = withRtkExactEpochGroups(groups, func(exact [][]*C.SidereonExactEpoch) error {
		cEpochs, count, copyErr := copyRtkDualFrequencyEpochsV2(epochs, exact, alloc)
		if copyErr != nil {
			return copyErr
		}
		withCThread(func() {
			status := C.sidereon_prepare_ionosphere_free_rtk_arc_v2(cEpochs, count, cCycles, cycleCount, cConfig, &pointer)
			copyErr = statusErrorLocked(uint32(status))
			if copyErr != nil && pointer != nil {
				C.sidereon_rtk_ionosphere_free_arc_solution_free(pointer)
				pointer = nil
			}
		})
		return copyErr
	})
	runtime.KeepAlive(epochs)
	if err != nil {
		return nil, err
	}
	return newRtkIonosphereFreeArcSolution(pointer)
}
