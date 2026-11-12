//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import "unsafe"

type PppFloatAmbiguityCycles struct {
	ID          string
	ValueCycles float64
}
type PppAmbiguitySearch struct {
	_      noCopy
	handle *positioningHandle
}

func releasePppAmbiguitySearch(pointer unsafe.Pointer) {
	withCThread(func() { C.sidereon_ppp_ambiguity_search_free((*C.SidereonPppAmbiguitySearch)(pointer)) })
}
func (search *PppAmbiguitySearch) Close() error {
	if search == nil || search.handle == nil {
		return nil
	}
	return search.handle.close()
}

func (solution *PppFixedSolution) AmbiguitySearch() (*PppAmbiguitySearch, error) {
	if solution == nil || solution.handle == nil {
		return nil, ErrClosed
	}
	var output *C.SidereonPppAmbiguitySearch
	err := solution.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_ppp_fixed_solution_ambiguity_search((*C.SidereonPppFixedSolution)(pointer), &output)))
		})
	})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, missingNativeHandle("PPP ambiguity search")
	}
	return &PppAmbiguitySearch{handle: newPositioningHandle(unsafe.Pointer(output), releasePppAmbiguitySearch)}, nil
}

func (search *PppAmbiguitySearch) Equal(other *PppAmbiguitySearch) (equal bool, err error) {
	if search == nil || search.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	err = withPositioningHandlePair(search.handle, other.handle, func(left, right unsafe.Pointer) error {
		return withCThreadError(func() error {
			var output C.bool
			if statusErr := statusErrorLocked(uint32(C.sidereon_ppp_ambiguity_search_equal((*C.SidereonPppAmbiguitySearch)(left), (*C.SidereonPppAmbiguitySearch)(right), &output))); statusErr != nil {
				return statusErr
			}
			equal = bool(output)
			return nil
		})
	})
	return equal, err
}

func pppAmbiguityMatrix(search *PppAmbiguitySearch, inverse bool) (result []float64, err error) {
	if search == nil || search.handle == nil {
		return nil, ErrClosed
	}
	err = search.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(output *C.double, count C.size_t) C.enum_SidereonStatus {
				if inverse {
					return C.sidereon_ppp_ambiguity_search_covariance_inverse_cycles((*C.SidereonPppAmbiguitySearch)(pointer), output, count, &written, &required)
				}
				return C.sidereon_ppp_ambiguity_search_covariance_cycles((*C.SidereonPppAmbiguitySearch)(pointer), output, count, &written, &required)
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP ambiguity matrix", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.double, count)
			var output *C.double
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP ambiguity matrix", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]float64, actual)
			for index := range result {
				result[index] = float64(values[index])
			}
			return nil
		})
	})
	return result, err
}
func (search *PppAmbiguitySearch) CovarianceCycles() ([]float64, error) {
	return pppAmbiguityMatrix(search, false)
}
func (search *PppAmbiguitySearch) CovarianceInverseCycles() ([]float64, error) {
	return pppAmbiguityMatrix(search, true)
}

func (search *PppAmbiguitySearch) FloatCycles() (result []PppFloatAmbiguityCycles, err error) {
	if search == nil || search.handle == nil {
		return nil, ErrClosed
	}
	err = search.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(output *C.SidereonPppFloatAmbiguityCycles, count C.size_t) C.enum_SidereonStatus {
				return C.sidereon_ppp_ambiguity_search_float_cycles((*C.SidereonPppAmbiguitySearch)(pointer), output, count, &written, &required)
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP float ambiguity cycles", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.SidereonPppFloatAmbiguityCycles, count)
			var output *C.SidereonPppFloatAmbiguityCycles
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP float ambiguity cycles", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]PppFloatAmbiguityCycles, actual)
			for index := range result {
				result[index] = PppFloatAmbiguityCycles{ID: C.GoString(&values[index].id.bytes[0]), ValueCycles: float64(values[index].value_cycles)}
			}
			return nil
		})
	})
	return result, err
}
