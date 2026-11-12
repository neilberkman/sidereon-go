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

type PppFloatResidual struct {
	EpochIndex  int
	SatelliteID string
	AmbiguityID string
	CodeM       float64
	PhaseM      float64
	CodeWeight  float64
	PhaseWeight float64
}

type PppResidualIonosphere struct {
	AmbiguityID string
	ValueM      float64
}

func pppSolutionScalar(handle *positioningHandle, fixed bool, decorrelation bool) (value float64, err error) {
	if handle == nil {
		return 0, ErrClosed
	}
	err = handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var output C.double
			var status C.enum_SidereonStatus
			switch {
			case decorrelation:
				status = C.sidereon_ppp_float_solution_decorrelation_time_epochs((*C.SidereonPppFloatSolution)(pointer), &output)
			case fixed:
				status = C.sidereon_ppp_fixed_solution_posterior_variance_factor((*C.SidereonPppFixedSolution)(pointer), &output)
			default:
				status = C.sidereon_ppp_float_solution_posterior_variance_factor((*C.SidereonPppFloatSolution)(pointer), &output)
			}
			if statusErr := statusErrorLocked(uint32(status)); statusErr != nil {
				return statusErr
			}
			value = float64(output)
			return nil
		})
	})
	return value, err
}

func (solution *PppFloatSolution) PosteriorVarianceFactor() (float64, error) {
	if solution == nil {
		return 0, ErrClosed
	}
	return pppSolutionScalar(solution.handle, false, false)
}

func (solution *PppFixedSolution) PosteriorVarianceFactor() (float64, error) {
	if solution == nil {
		return 0, ErrClosed
	}
	return pppSolutionScalar(solution.handle, true, false)
}

func (solution *PppFloatSolution) DecorrelationTimeEpochs() (float64, error) {
	if solution == nil {
		return 0, ErrClosed
	}
	return pppSolutionScalar(solution.handle, false, true)
}

func pppSolutionEqual(left, right *positioningHandle, fixed bool) (equal bool, err error) {
	if left == nil || right == nil {
		return false, ErrClosed
	}
	err = withPositioningHandlePair(left, right, func(leftPointer, rightPointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var output C.bool
			var status C.enum_SidereonStatus
			if fixed {
				status = C.sidereon_ppp_fixed_solution_equal((*C.SidereonPppFixedSolution)(leftPointer), (*C.SidereonPppFixedSolution)(rightPointer), &output)
			} else {
				status = C.sidereon_ppp_float_solution_equal((*C.SidereonPppFloatSolution)(leftPointer), (*C.SidereonPppFloatSolution)(rightPointer), &output)
			}
			if statusErr := statusErrorLocked(uint32(status)); statusErr != nil {
				return statusErr
			}
			equal = bool(output)
			return nil
		})
	})
	return equal, err
}

func (solution *PppFloatSolution) Equal(other *PppFloatSolution) (bool, error) {
	if solution == nil || other == nil {
		return false, ErrClosed
	}
	return pppSolutionEqual(solution.handle, other.handle, false)
}

func (solution *PppFixedSolution) Equal(other *PppFixedSolution) (bool, error) {
	if solution == nil || other == nil {
		return false, ErrClosed
	}
	return pppSolutionEqual(solution.handle, other.handle, true)
}

func pppResidualIonosphere(handle *positioningHandle, fixed bool) (result []PppResidualIonosphere, err error) {
	if handle == nil {
		return nil, ErrClosed
	}
	err = handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(output *C.SidereonPppResidualIonosphere, count C.size_t) C.enum_SidereonStatus {
				if fixed {
					return C.sidereon_ppp_fixed_solution_residual_ionosphere((*C.SidereonPppFixedSolution)(pointer), output, count, &written, &required)
				}
				return C.sidereon_ppp_float_solution_residual_ionosphere((*C.SidereonPppFloatSolution)(pointer), output, count, &written, &required)
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP residual ionosphere", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.SidereonPppResidualIonosphere, count)
			var output *C.SidereonPppResidualIonosphere
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP residual ionosphere", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]PppResidualIonosphere, actual)
			for index := range result {
				result[index] = PppResidualIonosphere{AmbiguityID: C.GoString(&values[index].ambiguity_id.bytes[0]), ValueM: float64(values[index].value_m)}
			}
			return nil
		})
	})
	runtime.KeepAlive(handle)
	return result, err
}

func (solution *PppFloatSolution) ResidualIonosphere() ([]PppResidualIonosphere, error) {
	if solution == nil {
		return nil, ErrClosed
	}
	return pppResidualIonosphere(solution.handle, false)
}

func (solution *PppFixedSolution) ResidualIonosphere() ([]PppResidualIonosphere, error) {
	if solution == nil {
		return nil, ErrClosed
	}
	return pppResidualIonosphere(solution.handle, true)
}

func (solution *PppFloatSolution) Residuals() (result []PppFloatResidual, err error) {
	if solution == nil || solution.handle == nil {
		return nil, ErrClosed
	}
	err = solution.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(output *C.SidereonPppFloatResidual, count C.size_t) C.enum_SidereonStatus {
				return C.sidereon_ppp_float_solution_residuals((*C.SidereonPppFloatSolution)(pointer), output, count, &written, &required)
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP float residuals", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.SidereonPppFloatResidual, count)
			var output *C.SidereonPppFloatResidual
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP float residuals", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]PppFloatResidual, actual)
			for index := range result {
				row := values[index]
				epoch, conversionErr := checkedNativeCount(uint64(row.epoch_index))
				if conversionErr != nil {
					return conversionErr
				}
				result[index] = PppFloatResidual{EpochIndex: epoch, SatelliteID: C.GoString(&row.satellite_id.bytes[0]), AmbiguityID: C.GoString(&row.ambiguity_id.bytes[0]), CodeM: float64(row.code_m), PhaseM: float64(row.phase_m), CodeWeight: float64(row.code_weight), PhaseWeight: float64(row.phase_weight)}
			}
			return nil
		})
	})
	runtime.KeepAlive(solution)
	return result, err
}
