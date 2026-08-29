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
	"time"
	"unsafe"
)

type TLEMetadata struct {
	CatalogNumber           string
	Classification          string
	InternationalDesignator string
	EpochYear               int
	EpochDayOfYear          float64
	InclinationDeg          float64
	RAANDeg                 float64
	Eccentricity            float64
	ArgumentOfPerigeeDeg    float64
	MeanAnomalyDeg          float64
	MeanMotionRevPerDay     float64
	MeanMotionDot           float64
	MeanMotionDoubleDot     float64
	BStar                   float64
	EphemerisType           int
	ElementSetNumber        int
	RevolutionNumber        int
}

type TLELines struct {
	Line1 string
	Line2 string
}

type TEMEState struct {
	EpochJ2000S float64
	PositionKm  [3]float64
	VelocityKmS [3]float64
}

type TLE struct {
	_      noCopy
	handle *positioningHandle
}

func releaseTLE(pointer unsafe.Pointer) {
	C.sidereon_tle_free((*C.SidereonTle)(pointer))
}

func ParseTLE(line1, line2 string) (*TLE, error) {
	if err := rejectEmbeddedNUL(line1, "TLE line 1"); err != nil {
		return nil, err
	}
	if err := rejectEmbeddedNUL(line2, "TLE line 2"); err != nil {
		return nil, err
	}
	var pointer *C.SidereonTle
	var err error
	withCThread(func() {
		cLine1 := C.CBytes(append([]byte(line1), 0))
		if cLine1 == nil {
			err = errors.New("sidereon: unable to allocate native TLE line 1")
			return
		}
		defer C.free(cLine1)
		cLine2 := C.CBytes(append([]byte(line2), 0))
		if cLine2 == nil {
			err = errors.New("sidereon: unable to allocate native TLE line 2")
			return
		}
		defer C.free(cLine2)
		err = statusErrorLocked(C.sidereon_tle_load(
			(*C.char)(cLine1), (*C.char)(cLine2), C.uint32_t(0), &pointer,
		))
		if err != nil && pointer != nil {
			releaseTLE(unsafe.Pointer(pointer))
			pointer = nil
		}
	})
	if err != nil {
		return nil, err
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native TLE load returned no handle")
	}
	return &TLE{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseTLE)}, nil
}

func (t *TLE) Close() error {
	if t == nil {
		return nil
	}
	return t.handle.close()
}

func cTLEString(pointer *C.char) string {
	return C.GoString(pointer)
}

func (t *TLE) Metadata() (TLEMetadata, error) {
	if t == nil || t.handle == nil {
		return TLEMetadata{}, ErrClosed
	}
	var metadata C.SidereonTleMetadata
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_metadata((*C.SidereonTle)(pointer), &metadata)
		})
	})
	runtime.KeepAlive(t)
	if err != nil {
		return TLEMetadata{}, err
	}
	return TLEMetadata{
		CatalogNumber:           cTLEString((*C.char)(unsafe.Pointer(&metadata.catalog_number[0]))),
		Classification:          cTLEString((*C.char)(unsafe.Pointer(&metadata.classification[0]))),
		InternationalDesignator: cTLEString((*C.char)(unsafe.Pointer(&metadata.international_designator[0]))),
		EpochYear:               int(metadata.epoch_year),
		EpochDayOfYear:          float64(metadata.epoch_day_of_year),
		InclinationDeg:          float64(metadata.inclination_deg),
		RAANDeg:                 float64(metadata.raan_deg),
		Eccentricity:            float64(metadata.eccentricity),
		ArgumentOfPerigeeDeg:    float64(metadata.arg_perigee_deg),
		MeanAnomalyDeg:          float64(metadata.mean_anomaly_deg),
		MeanMotionRevPerDay:     float64(metadata.mean_motion_rev_per_day),
		MeanMotionDot:           float64(metadata.mean_motion_dot),
		MeanMotionDoubleDot:     float64(metadata.mean_motion_double_dot),
		BStar:                   float64(metadata.bstar),
		EphemerisType:           int(metadata.ephemeris_type),
		ElementSetNumber:        int(metadata.elset_number),
		RevolutionNumber:        int(metadata.rev_number),
	}, nil
}

func (t *TLE) Lines() (TLELines, error) {
	if t == nil || t.handle == nil {
		return TLELines{}, ErrClosed
	}
	var lines C.SidereonTleLines
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_to_lines((*C.SidereonTle)(pointer), &lines)
		})
	})
	runtime.KeepAlive(t)
	if err != nil {
		return TLELines{}, err
	}
	return TLELines{
		Line1: cTLEString((*C.char)(unsafe.Pointer(&lines.line1.bytes[0]))),
		Line2: cTLEString((*C.char)(unsafe.Pointer(&lines.line2.bytes[0]))),
	}, nil
}

func civilJ2000Locked(value time.Time) (float64, error) {
	utc := value.UTC()
	year, month, day := utc.Date()
	second := float64(utc.Second()) + float64(utc.Nanosecond()/1000)/1e6
	var output C.double
	status := C.sidereon_civil_to_j2000_seconds(
		C.int32_t(year), C.int32_t(month), C.int32_t(day),
		C.int32_t(utc.Hour()), C.int32_t(utc.Minute()), C.double(second), &output,
	)
	if status != C.SIDEREON_STATUS_OK {
		return 0, statusErrorLocked(uint32(status))
	}
	return float64(output), nil
}

func (t *TLE) Propagate(times []time.Time) ([]TEMEState, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	if len(times) == 0 {
		return []TEMEState{}, nil
	}
	result := make([]TEMEState, len(times))
	var operationErr error
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			size, err := checkedNativeAllocationSize(len(times), unsafe.Sizeof(C.int64_t(0)))
			if err != nil {
				operationErr = err
				return
			}
			epochMemory := C.malloc(C.size_t(size))
			if epochMemory == nil {
				operationErr = errors.New("sidereon: unable to allocate native propagation epochs")
				return
			}
			defer C.free(epochMemory)
			epochs := unsafe.Slice((*C.int64_t)(epochMemory), len(times))
			for i, value := range times {
				utc := value.UTC()
				result[i].EpochJ2000S, operationErr = civilJ2000Locked(utc)
				if operationErr != nil {
					return
				}
				epochs[i] = C.int64_t(utc.UnixMicro())
			}

			var propagation *C.SidereonTlePropagation
			status := C.sidereon_tle_propagate(
				(*C.SidereonTle)(pointer), &epochs[0], C.size_t(len(times)), &propagation,
			)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				if propagation != nil {
					C.sidereon_tle_propagation_free(propagation)
				}
				return
			}
			if propagation == nil {
				operationErr = errors.New("sidereon: native TLE propagation returned no handle")
				return
			}
			defer C.sidereon_tle_propagation_free(propagation)

			var count C.size_t
			status = C.sidereon_tle_propagation_epoch_count(propagation, &count)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			stateCount, err := checkedNativeCount(uint64(count))
			if err != nil {
				operationErr = err
				return
			}
			if stateCount != len(result) {
				operationErr = errors.New("sidereon: native propagation returned an unexpected epoch count")
				return
			}
			if _, err := checkedNativeAllocationSize(stateCount, unsafe.Sizeof(C.SidereonTemeState{})); err != nil {
				operationErr = err
				return
			}
			states := make([]C.SidereonTemeState, stateCount)
			var output *C.SidereonTemeState
			if len(states) != 0 {
				output = &states[0]
			}
			var written, required C.size_t
			status = C.sidereon_tle_propagation_states(
				propagation, output, C.size_t(len(states)), &written, &required,
			)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			_, err = validateTwoPassCounts(
				"TLE propagation states", len(states), stateCount, uint64(written), uint64(required),
			)
			if err != nil {
				operationErr = err
				return
			}
			for i := range states {
				for axis := 0; axis < 3; axis++ {
					result[i].PositionKm[axis] = float64(states[i].position_km[axis])
					result[i].VelocityKmS[axis] = float64(states[i].velocity_km_s[axis])
				}
			}
		})
		return operationErr
	})
	runtime.KeepAlive(t)
	runtime.KeepAlive(times)
	if err != nil {
		return nil, err
	}
	return result, nil
}
