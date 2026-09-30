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

type PppResidualScreenRemoval struct {
	EpochIndex  int
	AmbiguityID string
}

func pppCopySlice[T any](label string, fn func(unsafe.Pointer, C.size_t, *C.size_t, *C.size_t) uint32) ([]T, error) {
	var written, required C.size_t
	if err := callStatus(func() uint32 { return fn(nil, 0, &written, &required) }); err != nil {
		return nil, err
	}
	count, err := checkedNativeCount(uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, label+" query count"); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(*new(T))); err != nil {
		return nil, err
	}
	values := make([]T, count)
	var output unsafe.Pointer
	if len(values) > 0 {
		output = unsafe.Pointer(&values[0])
	}
	if err := callStatus(func() uint32 { return fn(output, C.size_t(count), &written, &required) }); err != nil {
		return nil, err
	}
	n, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	return values[:n], nil
}

func pppSolvedEpochIndices(handle *positioningHandle, fixed bool) (result []int, err error) {
	if handle == nil {
		return nil, ErrClosed
	}
	err = handle.with(func(pointer unsafe.Pointer) error {
		values, e := pppCopySlice[C.size_t]("PPP solved epoch index", func(out unsafe.Pointer, n C.size_t, w, r *C.size_t) uint32 {
			if fixed {
				return uint32(C.sidereon_ppp_fixed_solution_solved_epochs((*C.SidereonPppFixedSolution)(pointer), (*C.size_t)(out), n, w, r))
			}
			return uint32(C.sidereon_ppp_float_solution_solved_epochs((*C.SidereonPppFloatSolution)(pointer), (*C.size_t)(out), n, w, r))
		})
		if e != nil {
			return e
		}
		result = make([]int, len(values))
		for i, value := range values {
			result[i], e = sizeTToInt(value, "PPP solved input epoch index")
			if e != nil {
				return e
			}
		}
		return nil
	})
	return result, err
}

func pppEpochClocks(handle *positioningHandle, fixed bool) (result []float64, err error) {
	if handle == nil {
		return nil, ErrClosed
	}
	err = handle.with(func(pointer unsafe.Pointer) error {
		values, e := pppCopySlice[C.double]("PPP solved epoch receiver clock", func(out unsafe.Pointer, n C.size_t, w, r *C.size_t) uint32 {
			if fixed {
				return uint32(C.sidereon_ppp_fixed_solution_epoch_clocks((*C.SidereonPppFixedSolution)(pointer), (*C.double)(out), n, w, r))
			}
			return uint32(C.sidereon_ppp_float_solution_epoch_clocks((*C.SidereonPppFloatSolution)(pointer), (*C.double)(out), n, w, r))
		})
		if e != nil {
			return e
		}
		result = make([]float64, len(values))
		for i, value := range values {
			result[i] = float64(value)
		}
		return nil
	})
	return result, err
}

func (s *PppFloatSolution) SolvedEpochIndices() ([]int, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	result, err := pppSolvedEpochIndices(s.handle, false)
	runtime.KeepAlive(s)
	return result, err
}
func (s *PppFixedSolution) SolvedEpochIndices() ([]int, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	result, err := pppSolvedEpochIndices(s.handle, true)
	runtime.KeepAlive(s)
	return result, err
}
func (s *PppFloatSolution) EpochClocksM() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	result, err := pppEpochClocks(s.handle, false)
	runtime.KeepAlive(s)
	return result, err
}
func (s *PppFixedSolution) EpochClocksM() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	result, err := pppEpochClocks(s.handle, true)
	runtime.KeepAlive(s)
	return result, err
}
func (s *PppFloatSolution) ResidualScreenRemovals() (result []PppResidualScreenRemoval, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		values, e := pppCopySlice[C.SidereonPppEpochObservation]("PPP residual-screen removal", func(out unsafe.Pointer, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_ppp_float_solution_residual_screen_removals((*C.SidereonPppFloatSolution)(pointer), (*C.SidereonPppEpochObservation)(out), n, w, r))
		})
		if e != nil {
			return e
		}
		result = make([]PppResidualScreenRemoval, len(values))
		for i, value := range values {
			index, e := sizeTToInt(value.epoch_index, "PPP residual-screen epoch index")
			if e != nil {
				return e
			}
			result[i] = PppResidualScreenRemoval{EpochIndex: index, AmbiguityID: C.GoString(&value.ambiguity_id.bytes[0])}
		}
		return nil
	})
	runtime.KeepAlive(s)
	return result, err
}
