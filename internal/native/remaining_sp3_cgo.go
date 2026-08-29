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
	"fmt"
	"runtime"
	"unsafe"
)

type NativeSp3EpochPrediction struct {
	EpochJ2000S                  float64
	Observed                     bool
	OrbitPredictedSatelliteCount int
	ClockPredictedSatelliteCount int
}

type NativeSp3ClockReferenceOffset struct {
	EpochJ2000S float64
	OffsetS     float64
	Satellites  int
}

type NativeSp3Continuity struct {
	Defects          int
	ResidualsChecked int
	ResidualsSkipped int
}

// withPositioningHandlePair acquires two handle read locks in address order.
// Keeping the helper here makes any future same-kind pair operation use the
// same deadlock-free ordering as SP3 clock operations.
func withPositioningHandlePair(left, right *positioningHandle, fn func(unsafe.Pointer, unsafe.Pointer) error) error {
	if left == nil || right == nil {
		return ErrClosed
	}
	if left == right {
		left.mu.RLock()
		defer left.mu.RUnlock()
		if left.resource == nil {
			return ErrClosed
		}
		return left.resource.with(func(pointer unsafe.Pointer) error {
			return fn(pointer, pointer)
		})
	}
	swapped := uintptr(unsafe.Pointer(left)) > uintptr(unsafe.Pointer(right))
	first, second := left, right
	if swapped {
		first, second = second, first
	}
	first.mu.RLock()
	defer first.mu.RUnlock()
	second.mu.RLock()
	defer second.mu.RUnlock()
	if first.resource == nil || second.resource == nil {
		return ErrClosed
	}
	return first.resource.with(func(firstPointer unsafe.Pointer) error {
		return second.resource.with(func(secondPointer unsafe.Pointer) error {
			if swapped {
				return fn(secondPointer, firstPointer)
			}
			return fn(firstPointer, secondPointer)
		})
	})
}

func checkedSp3OrbitClass(value int) (C.int32_t, error) {
	if value < -1<<31 || value > 1<<31-1 {
		return 0, errors.New("sidereon: SP3 orbit class does not fit in int32")
	}
	return C.int32_t(value), nil
}

func (s *SP3) EphemerisSample(satellites []string, start, stop, step float64) ([]NativeEphemerisSampleRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	if len(satellites) == 0 {
		return nil, invalidArgument("SP3 sample satellite list must not be empty")
	}
	var result []NativeEphemerisSampleRow
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satPtr **C.char, count C.size_t) error {
			var written, required C.size_t
			invoke := func(out *C.SidereonEphemerisSampleRow, n C.size_t) uint32 {
				return uint32(C.sidereon_sp3_ephemeris_sample((*C.SidereonSp3)(pointer), (**C.char)(satPtr), count, C.double(start), C.double(stop), C.double(step), out, n, &written, &required))
			}
			if err := callStatus(func() uint32 { return invoke(nil, 0) }); err != nil {
				return err
			}
			countValue, err := validateNativeQuery("SP3 ephemeris sample", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(countValue, unsafe.Sizeof(C.SidereonEphemerisSampleRow{})); err != nil {
				return err
			}
			values := make([]C.SidereonEphemerisSampleRow, countValue)
			var out *C.SidereonEphemerisSampleRow
			if countValue != 0 {
				out = &values[0]
			}
			written, required = 0, 0
			if err := callStatus(func() uint32 { return invoke(out, C.size_t(countValue)) }); err != nil {
				return err
			}
			writtenCount, err := validateNativeOutput("SP3 ephemeris sample", countValue, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			result = make([]NativeEphemerisSampleRow, writtenCount)
			for i := range result {
				if err := validateEphemerisSampleStatusValue(uint32(values[i].status)); err != nil {
					return err
				}
				result[i] = NativeEphemerisSampleRow{SatelliteID: tokenFromC(values[i].sat_id), EpochJ2000S: float64(values[i].epoch_j2000_s), Status: uint32(values[i].status), HasPosition: bool(values[i].has_position_ecef_m), HasClock: bool(values[i].has_clock_s), ClockS: float64(values[i].clock_s)}
				for j := range result[i].Position {
					result[i].Position[j] = float64(values[i].position_ecef_m[j])
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *SP3) DeclaredEpochCount() (uint64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var out C.uint64_t
	err := s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_sp3_declared_epoch_count((*C.SidereonSp3)(p), &out)) })
	})
	runtime.KeepAlive(s)
	return uint64(out), err
}

func (s *SP3) DeclaredStart() (bool, float64, error) {
	if s == nil || s.handle == nil {
		return false, 0, ErrClosed
	}
	var present C.uint8_t
	var seconds C.double
	err := s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_declared_start_j2000_seconds((*C.SidereonSp3)(p), &present, &seconds))
		})
	})
	runtime.KeepAlive(s)
	if present > 1 {
		if err == nil {
			err = invalidArgument("SP3 declared-start presence is not a boolean")
		}
		return false, 0, err
	}
	return present != 0, float64(seconds), err
}

func (s *SP3) EpochPrediction(index int) (NativeSp3EpochPrediction, error) {
	if s == nil || s.handle == nil {
		return NativeSp3EpochPrediction{}, ErrClosed
	}
	if index < 0 {
		return NativeSp3EpochPrediction{}, errors.New("sidereon: epoch index must not be negative")
	}
	count, err := checkedNativeSize(index)
	if err != nil {
		return NativeSp3EpochPrediction{}, err
	}
	var out C.SidereonSp3EpochPrediction
	err = s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_sp3_epoch_prediction((*C.SidereonSp3)(p), count, &out)) })
	})
	runtime.KeepAlive(s)
	if err != nil {
		return NativeSp3EpochPrediction{}, err
	}
	orbitCount, err := checkedNativeCount(uint64(out.orbit_predicted_satellite_count))
	if err != nil {
		return NativeSp3EpochPrediction{}, fmt.Errorf("sidereon: native orbit predicted satellite count: %w", err)
	}
	clockCount, err := checkedNativeCount(uint64(out.clock_predicted_satellite_count))
	if err != nil {
		return NativeSp3EpochPrediction{}, fmt.Errorf("sidereon: native clock predicted satellite count: %w", err)
	}
	return NativeSp3EpochPrediction{EpochJ2000S: float64(out.epoch_j2000_seconds), Observed: bool(out.observed), OrbitPredictedSatelliteCount: orbitCount, ClockPredictedSatelliteCount: clockCount}, nil
}

func (s *SP3) Continuity(orbitClass int, residualTolerance float64) (NativeSp3Continuity, error) {
	if s == nil || s.handle == nil {
		return NativeSp3Continuity{}, ErrClosed
	}
	class, err := checkedSp3OrbitClass(orbitClass)
	if err != nil {
		return NativeSp3Continuity{}, err
	}
	var defects, checked, skipped C.size_t
	err = s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_check_continuity((*C.SidereonSp3)(p), class, C.double(residualTolerance), &defects, &checked, &skipped))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return NativeSp3Continuity{}, err
	}
	defectCount, err := checkedNativeCount(uint64(defects))
	if err != nil {
		return NativeSp3Continuity{}, fmt.Errorf("sidereon: native continuity defects: %w", err)
	}
	checkedCount, err := checkedNativeCount(uint64(checked))
	if err != nil {
		return NativeSp3Continuity{}, fmt.Errorf("sidereon: native continuity checked count: %w", err)
	}
	skippedCount, err := checkedNativeCount(uint64(skipped))
	if err != nil {
		return NativeSp3Continuity{}, fmt.Errorf("sidereon: native continuity skipped count: %w", err)
	}
	return NativeSp3Continuity{Defects: defectCount, ResidualsChecked: checkedCount, ResidualsSkipped: skippedCount}, nil
}

func (s *SP3) ClockReferenceOffsets(other *SP3, minCommon int) ([]NativeSp3ClockReferenceOffset, error) {
	if s == nil || s.handle == nil || other == nil || other.handle == nil {
		return nil, ErrClosed
	}
	if minCommon < 0 {
		return nil, errors.New("sidereon: min common must not be negative")
	}
	common, err := checkedNativeSize(minCommon)
	if err != nil {
		return nil, err
	}
	var result []NativeSp3ClockReferenceOffset
	// Hold both read locks for the complete query/fill sequence. This also
	// prevents Close from freeing either product between the two C calls.
	err = withPositioningHandlePair(s.handle, other.handle, func(reference unsafe.Pointer, candidate unsafe.Pointer) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonSp3ClockReferenceOffset, n C.size_t) uint32 {
			return uint32(C.sidereon_sp3_clock_reference_offsets((*C.SidereonSp3)(reference), (*C.SidereonSp3)(candidate), common, out, n, &written, &required))
		}
		if err := callStatus(func() uint32 { return invoke(nil, 0) }); err != nil {
			return err
		}
		count, err := validateNativeQuery("SP3 clock reference offsets", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonSp3ClockReferenceOffset{})); err != nil {
			return err
		}
		values := make([]C.SidereonSp3ClockReferenceOffset, count)
		var out *C.SidereonSp3ClockReferenceOffset
		if count != 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(out, C.size_t(count)) }); err != nil {
			return err
		}
		writtenCount, err := validateNativeOutput("SP3 clock reference offsets", count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeSp3ClockReferenceOffset, writtenCount)
		for i := range result {
			satellites, countErr := checkedNativeCount(uint64(values[i].satellites))
			if countErr != nil {
				return fmt.Errorf("sidereon: native clock reference satellite count: %w", countErr)
			}
			result[i] = NativeSp3ClockReferenceOffset{EpochJ2000S: float64(values[i].epoch_j2000_seconds), OffsetS: float64(values[i].offset_s), Satellites: satellites}
		}
		return nil
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(other)
	return result, err
}

func (s *SP3) AlignClockReference(other *SP3, minCommon int) (*SP3, error) {
	if s == nil || s.handle == nil || other == nil || other.handle == nil {
		return nil, ErrClosed
	}
	if minCommon < 0 {
		return nil, errors.New("sidereon: min common must not be negative")
	}
	common, err := checkedNativeSize(minCommon)
	if err != nil {
		return nil, err
	}
	var out *C.SidereonSp3
	err = withPositioningHandlePair(s.handle, other.handle, func(reference unsafe.Pointer, candidate unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_align_clock_reference((*C.SidereonSp3)(reference), (*C.SidereonSp3)(candidate), common, &out))
		})
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(other)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_sp3_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("aligned SP3")
	}
	return &SP3{handle: newPositioningHandle(unsafe.Pointer(out), releaseSP3)}, nil
}

func (s *SP3) Interpolate(satellite string, epochs []float64) ([][3]float64, []float64, int, error) {
	if s == nil || s.handle == nil {
		return nil, nil, 0, ErrClosed
	}
	if len(epochs) == 0 {
		return nil, nil, 0, nil
	}
	if _, err := checkedNativeAllocationSize(len(epochs), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, nil, 0, err
	}
	epochMemory, err := checkedNativeMalloc(len(epochs), unsafe.Sizeof(C.double(0)))
	if err != nil {
		return nil, nil, 0, err
	}
	defer C.free(epochMemory)
	positionMemory, err := checkedNativeMalloc(len(epochs), unsafe.Sizeof(C.double(0))*3)
	if err != nil {
		return nil, nil, 0, err
	}
	defer C.free(positionMemory)
	clockMemory, err := checkedNativeMalloc(len(epochs), unsafe.Sizeof(C.double(0)))
	if err != nil {
		return nil, nil, 0, err
	}
	defer C.free(clockMemory)
	epochValues := unsafe.Slice((*C.double)(epochMemory), len(epochs))
	for i := range epochs {
		epochValues[i] = C.double(epochs[i])
	}
	var positions [][3]float64
	var clocks []float64
	var written C.size_t
	var countErr error
	err = s.handle.with(func(p unsafe.Pointer) error {
		callErr := withString(satellite, func(id *C.char) uint32 {
			status := C.sidereon_sp3_interpolate((*C.SidereonSp3)(p), id, (*C.double)(epochMemory), C.size_t(len(epochs)), (*C.double)(positionMemory), C.size_t(len(epochs)*3), (*C.double)(clockMemory), C.size_t(len(epochs)), &written)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return uint32(status)
			}
			writtenInt, conversionErr := writtenToInt(written, len(epochs), "SP3 interpolation written count")
			if conversionErr != nil {
				countErr = conversionErr
				return uint32(C.SIDEREON_STATUS_OK)
			}
			positions = make([][3]float64, writtenInt)
			clocks = make([]float64, writtenInt)
			pv := unsafe.Slice((*C.double)(positionMemory), writtenInt*3)
			cv := unsafe.Slice((*C.double)(clockMemory), writtenInt)
			for i := range positions {
				for j := 0; j < 3; j++ {
					positions[i][j] = float64(pv[i*3+j])
				}
				clocks[i] = float64(cv[i])
			}
			return uint32(C.SIDEREON_STATUS_OK)
		})
		if countErr != nil {
			return countErr
		}
		return callErr
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(epochs)
	if err != nil {
		return positions, clocks, 0, err
	}
	writtenCount, conversionErr := writtenToInt(written, len(epochs), "SP3 interpolation written count")
	if conversionErr != nil {
		return positions, clocks, 0, conversionErr
	}
	return positions, clocks, writtenCount, nil
}

func (s *SP3) ContinuityVerdictJSON(orbitClass int, residualTolerance, from, through float64) ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	class, err := checkedSp3OrbitClass(orbitClass)
	if err != nil {
		return nil, err
	}
	var result []byte
	err = s.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var err error
			result, err = copyNativeBytesLocked("SP3 continuity verdict", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_continuity_verdict_json((*C.SidereonSp3)(p), class, C.double(residualTolerance), C.double(from), C.double(through), out, n, w, r)
			})
			return err
		})
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *SP3) Text() ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := s.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var err error
			result, err = copyNativeBytesLocked("SP3 text", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_to_sp3_text((*C.SidereonSp3)(p), out, n, w, r)
			})
			return err
		})
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *SP3) ArtifactBytes() ([]byte, uint32, error) {
	if s == nil || s.handle == nil {
		return nil, 0, ErrClosed
	}
	var artifactError C.enum_SidereonPreciseInterpolantArtifactErrorKind
	var result []byte
	err := s.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var err error
			result, err = copyNativeBytesLocked("SP3 precise artifact", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_precise_interpolant_artifact_bytes((*C.SidereonSp3)(p), &artifactError, out, n, w, r)
			})
			return err
		})
	})
	runtime.KeepAlive(s)
	if enumErr := validatePreciseInterpolantArtifactError(uint32(artifactError)); enumErr != nil {
		artifactError = C.enum_SidereonPreciseInterpolantArtifactErrorKind(PreciseInterpolantArtifactErrorNoneValue)
		if err == nil {
			err = enumErr
		}
	}
	return result, uint32(artifactError), err
}

func (s *SP3) PreciseSamples() ([]PreciseEphemerisSample, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []PreciseEphemerisSample
	err := s.handle.with(func(p unsafe.Pointer) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonPreciseEphemerisSample, n C.size_t) uint32 {
			return uint32(C.sidereon_sp3_precise_ephemeris_samples((*C.SidereonSp3)(p), out, n, &written, &required))
		}
		if err := callStatus(func() uint32 { return invoke(nil, 0) }); err != nil {
			return err
		}
		count, err := validateNativeQuery("SP3 precise samples", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonPreciseEphemerisSample{})); err != nil {
			return err
		}
		values := make([]C.SidereonPreciseEphemerisSample, count)
		var out *C.SidereonPreciseEphemerisSample
		if count != 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(out, C.size_t(count)) }); err != nil {
			return err
		}
		writtenCount, err := validateNativeOutput("SP3 precise samples", count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]PreciseEphemerisSample, writtenCount)
		for i := range result {
			value := values[i]
			if err := validTimeScale(uint32(value.time_scale)); err != nil {
				return err
			}
			result[i].Satellite = tokenFromC(value.sat)
			result[i].TimeScale = uint32(value.time_scale)
			result[i].EpochJ2000S = float64(value.epoch_j2000_s)
			result[i].HasClock = bool(value.has_clock_s)
			result[i].ClockS = float64(value.clock_s)
			result[i].ClockEvent = bool(value.clock_event)
			for axis := range result[i].PositionECEFM {
				result[i].PositionECEFM[axis] = float64(value.position_ecef_m[axis])
			}
		}
		return nil
	})
	runtime.KeepAlive(s)
	return result, err
}
