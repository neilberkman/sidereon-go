//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import "unsafe"

type ExactEpochComponents struct {
	Seconds       int64
	Attoseconds   uint64
	ResidueDigits int64
	ResiduePlaces uint16
}

type ExactOrdering int32

const (
	ExactOrderingLess    ExactOrdering = -1
	ExactOrderingEqual   ExactOrdering = 0
	ExactOrderingGreater ExactOrdering = 1
)

type ExactEpoch struct{ handle *positioningHandle }
type ExactEpochQuery struct{ handle *positioningHandle }

func releaseExactEpoch(pointer unsafe.Pointer) {
	C.sidereon_exact_epoch_free((*C.SidereonExactEpoch)(pointer))
}

func releaseExactEpochQuery(pointer unsafe.Pointer) {
	C.sidereon_exact_epoch_query_free((*C.SidereonExactEpochQuery)(pointer))
}

func newExactEpoch(pointer *C.SidereonExactEpoch) (*ExactEpoch, error) {
	if pointer == nil {
		return nil, missingNativeHandle("exact epoch construction")
	}
	return &ExactEpoch{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseExactEpoch)}, nil
}

func newExactEpochQuery(pointer *C.SidereonExactEpochQuery) (*ExactEpochQuery, error) {
	if pointer == nil {
		return nil, missingNativeHandle("exact epoch query construction")
	}
	return &ExactEpochQuery{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseExactEpochQuery)}, nil
}

func NewExactEpoch(seconds int64, attoseconds uint64) (*ExactEpoch, error) {
	var pointer *C.SidereonExactEpoch
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_exact_epoch_new(C.int64_t(seconds), C.uint64_t(attoseconds), &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_epoch_free(pointer) })
		}
		return nil, err
	}
	return newExactEpoch(pointer)
}

func ExactEpochJ2000() (*ExactEpoch, error) {
	var pointer *C.SidereonExactEpoch
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_exact_epoch_j2000(&pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_epoch_free(pointer) })
		}
		return nil, err
	}
	return newExactEpoch(pointer)
}

func ExactAttosecondsPerSecond() (uint64, error) {
	var attoseconds C.uint64_t
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_exact_epoch_attoseconds_per_second(&attoseconds))
	})
	return uint64(attoseconds), err
}

func ExactEpochFromJ2000Seconds(seconds float64) (*ExactEpoch, error) {
	var pointer *C.SidereonExactEpoch
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_exact_epoch_from_j2000_seconds(C.double(seconds), &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_epoch_free(pointer) })
		}
		return nil, err
	}
	return newExactEpoch(pointer)
}

func ExactEpochFromCivil(value CivilDateTime) (*ExactEpoch, error) {
	year, err := checkedInt32(value.Year, "civil year")
	if err != nil {
		return nil, err
	}
	month, err := checkedInt32(value.Month, "civil month")
	if err != nil {
		return nil, err
	}
	day, err := checkedInt32(value.Day, "civil day")
	if err != nil {
		return nil, err
	}
	hour, err := checkedInt32(value.Hour, "civil hour")
	if err != nil {
		return nil, err
	}
	minute, err := checkedInt32(value.Minute, "civil minute")
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonExactEpoch
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_exact_epoch_from_civil(
			C.int32_t(year), C.int32_t(month), C.int32_t(day), C.int32_t(hour),
			C.int32_t(minute), C.double(value.Second), &pointer,
		))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_epoch_free(pointer) })
		}
		return nil, err
	}
	return newExactEpoch(pointer)
}

func ExactEpochQueryFromBinaryJ2000Seconds(seconds float64) (*ExactEpochQuery, error) {
	var pointer *C.SidereonExactEpochQuery
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_exact_epoch_query_from_binary_j2000_seconds(C.double(seconds), &pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_epoch_query_free(pointer) })
		}
		return nil, err
	}
	return newExactEpochQuery(pointer)
}

func (epoch *ExactEpoch) Query() (*ExactEpochQuery, error) {
	if epoch == nil || epoch.handle == nil {
		return nil, ErrClosed
	}
	var result *C.SidereonExactEpochQuery
	err := epoch.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_query((*C.SidereonExactEpoch)(pointer), &result))
		})
	})
	if err != nil {
		if result != nil {
			withCThread(func() { C.sidereon_exact_epoch_query_free(result) })
		}
		return nil, err
	}
	return newExactEpochQuery(result)
}

func (query *ExactEpochQuery) Epoch() (*ExactEpoch, error) {
	if query == nil || query.handle == nil {
		return nil, ErrClosed
	}
	var result *C.SidereonExactEpoch
	err := query.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_query_epoch(
				(*C.SidereonExactEpochQuery)(pointer), &result,
			))
		})
	})
	if err != nil {
		if result != nil {
			withCThread(func() { C.sidereon_exact_epoch_free(result) })
		}
		return nil, err
	}
	return newExactEpoch(result)
}

func (epoch *ExactEpoch) Compare(other *ExactEpoch) (ExactOrdering, error) {
	if epoch == nil || epoch.handle == nil || other == nil || other.handle == nil {
		return ExactOrderingEqual, ErrClosed
	}
	var ordering int32
	err := withPositioningHandles([]*positioningHandle{epoch.handle, other.handle}, func(pointers []unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_compare(
				(*C.SidereonExactEpoch)(pointers[0]),
				(*C.SidereonExactEpoch)(pointers[1]), &ordering,
			))
		})
	})
	return ExactOrdering(ordering), err
}

func (epoch *ExactEpoch) Equal(other *ExactEpoch) (bool, error) {
	if epoch == nil || epoch.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	var equal C.bool
	err := withPositioningHandles([]*positioningHandle{epoch.handle, other.handle}, func(pointers []unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_equal(
				(*C.SidereonExactEpoch)(pointers[0]),
				(*C.SidereonExactEpoch)(pointers[1]), &equal,
			))
		})
	})
	return bool(equal), err
}

func (query *ExactEpochQuery) Equal(other *ExactEpochQuery) (bool, error) {
	if query == nil || query.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	var equal C.bool
	err := withPositioningHandles([]*positioningHandle{query.handle, other.handle}, func(pointers []unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_query_equal(
				(*C.SidereonExactEpochQuery)(pointers[0]),
				(*C.SidereonExactEpochQuery)(pointers[1]), &equal,
			))
		})
	})
	return bool(equal), err
}

func (epoch *ExactEpoch) CheckedAddSeconds(seconds float64) (*ExactEpoch, error) {
	return epoch.checkedOffset(seconds, true)
}

func (epoch *ExactEpoch) CheckedSubSeconds(seconds float64) (*ExactEpoch, error) {
	return epoch.checkedOffset(seconds, false)
}

func (epoch *ExactEpoch) checkedOffset(seconds float64, add bool) (*ExactEpoch, error) {
	if epoch == nil || epoch.handle == nil {
		return nil, ErrClosed
	}
	var result *C.SidereonExactEpoch
	err := epoch.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			if add {
				return uint32(C.sidereon_exact_epoch_checked_add_seconds(
					(*C.SidereonExactEpoch)(pointer), C.double(seconds), &result,
				))
			}
			return uint32(C.sidereon_exact_epoch_checked_sub_seconds(
				(*C.SidereonExactEpoch)(pointer), C.double(seconds), &result,
			))
		})
	})
	if err != nil {
		if result != nil {
			withCThread(func() { C.sidereon_exact_epoch_free(result) })
		}
		return nil, err
	}
	return newExactEpoch(result)
}

func (query *ExactEpochQuery) CheckedAddBinarySeconds(seconds float64) (*ExactEpochQuery, error) {
	return query.checkedBinaryOffset(seconds, true)
}

func (query *ExactEpochQuery) CheckedSubBinarySeconds(seconds float64) (*ExactEpochQuery, error) {
	return query.checkedBinaryOffset(seconds, false)
}

func (query *ExactEpochQuery) checkedBinaryOffset(seconds float64, add bool) (*ExactEpochQuery, error) {
	if query == nil || query.handle == nil {
		return nil, ErrClosed
	}
	var result *C.SidereonExactEpochQuery
	err := query.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			if add {
				return uint32(C.sidereon_exact_epoch_query_checked_add_binary_seconds(
					(*C.SidereonExactEpochQuery)(pointer), C.double(seconds), &result,
				))
			}
			return uint32(C.sidereon_exact_epoch_query_checked_sub_binary_seconds(
				(*C.SidereonExactEpochQuery)(pointer), C.double(seconds), &result,
			))
		})
	})
	if err != nil {
		if result != nil {
			withCThread(func() { C.sidereon_exact_epoch_query_free(result) })
		}
		return nil, err
	}
	return newExactEpochQuery(result)
}

func (epoch *ExactEpoch) Components() (ExactEpochComponents, error) {
	if epoch == nil || epoch.handle == nil {
		return ExactEpochComponents{}, ErrClosed
	}
	var seconds C.int64_t
	var attoseconds C.uint64_t
	var residueDigits C.int64_t
	var residuePlaces C.uint16_t
	err := epoch.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_components(
				(*C.SidereonExactEpoch)(pointer), &seconds, &attoseconds, &residueDigits, &residuePlaces,
			))
		})
	})
	return ExactEpochComponents{
		Seconds: int64(seconds), Attoseconds: uint64(attoseconds),
		ResidueDigits: int64(residueDigits), ResiduePlaces: uint16(residuePlaces),
	}, err
}

func (epoch *ExactEpoch) J2000Seconds() (float64, error) {
	if epoch == nil || epoch.handle == nil {
		return 0, ErrClosed
	}
	var seconds C.double
	err := epoch.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_j2000_seconds((*C.SidereonExactEpoch)(pointer), &seconds))
		})
	})
	return float64(seconds), err
}

func (epoch *ExactEpoch) SplitJulianDate() (JulianDate, error) {
	if epoch == nil || epoch.handle == nil {
		return JulianDate{}, ErrClosed
	}
	var whole, fraction C.double
	err := epoch.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_split_julian_date(
				(*C.SidereonExactEpoch)(pointer), &whole, &fraction,
			))
		})
	})
	return JulianDate{Whole: float64(whole), Fraction: float64(fraction)}, err
}

func (epoch *ExactEpoch) SecondsSince(earlier *ExactEpoch) (float64, error) {
	if epoch == nil || epoch.handle == nil || earlier == nil || earlier.handle == nil {
		return 0, ErrClosed
	}
	var seconds C.double
	err := withPositioningHandles([]*positioningHandle{epoch.handle, earlier.handle}, func(pointers []unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_seconds_since(
				(*C.SidereonExactEpoch)(pointers[0]), (*C.SidereonExactEpoch)(pointers[1]), &seconds,
			))
		})
	})
	return float64(seconds), err
}

func (query *ExactEpochQuery) SecondsSince(earlier *ExactEpoch) (float64, error) {
	if query == nil || query.handle == nil || earlier == nil || earlier.handle == nil {
		return 0, ErrClosed
	}
	var seconds C.double
	err := withPositioningHandles([]*positioningHandle{query.handle, earlier.handle}, func(pointers []unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_query_seconds_since(
				(*C.SidereonExactEpochQuery)(pointers[0]), (*C.SidereonExactEpoch)(pointers[1]), &seconds,
			))
		})
	})
	return float64(seconds), err
}

func (query *ExactEpochQuery) SecondsSinceQuery(earlier *ExactEpochQuery) (float64, error) {
	if query == nil || query.handle == nil || earlier == nil || earlier.handle == nil {
		return 0, ErrClosed
	}
	var seconds C.double
	err := withPositioningHandles([]*positioningHandle{query.handle, earlier.handle}, func(pointers []unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_query_seconds_since_query(
				(*C.SidereonExactEpochQuery)(pointers[0]), (*C.SidereonExactEpochQuery)(pointers[1]), &seconds,
			))
		})
	})
	return float64(seconds), err
}

func (query *ExactEpochQuery) J2000Seconds() (float64, error) {
	if query == nil || query.handle == nil {
		return 0, ErrClosed
	}
	var seconds C.double
	err := query.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_epoch_query_j2000_seconds((*C.SidereonExactEpochQuery)(pointer), &seconds))
		})
	})
	return float64(seconds), err
}

func (epoch *ExactEpoch) Close() error {
	if epoch == nil || epoch.handle == nil {
		return nil
	}
	return epoch.handle.close()
}

func (query *ExactEpochQuery) Close() error {
	if query == nil || query.handle == nil {
		return nil
	}
	return query.handle.close()
}
