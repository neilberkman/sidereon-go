//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type RtkRinexArcObservation struct {
	SatelliteID, AmbiguityID string
	CodeM, PhaseM            float64
	HasLLI                   bool
	LLI                      int64
}

type RtkRinexArcPosition struct {
	SatelliteID string
	PositionM   [3]float64
}

type RtkRinexArcEpochMetadata struct {
	BaseCount, RoverCount, SatellitePositionCount           int
	BaseSatellitePositionCount, RoverSatellitePositionCount int
	HasVelocityMPS                                          bool
	VelocityMPS                                             [3]float64
	HasPredictionTime                                       bool
	PredictionTimeS                                         float64
}

type RtkRinexMapValue struct {
	ID    string
	Value float64
}

type RtkRinexDualFrequencyObservation struct {
	AmbiguityID                                  string
	P1M, P2M, Phi1Cycles, Phi2Cycles, F1Hz, F2Hz float64
	HasLLI1, HasLLI2                             bool
	LLI1, LLI2                                   int64
}

type RtkRinexDualFrequencySatelliteObservation struct {
	SatelliteID string
	Base, Rover RtkRinexDualFrequencyObservation
}

type RtkRinexDualFrequencyArcEpochMetadata struct {
	JDWhole, JDFraction, GapTimeS, PredictionTimeS float64
	HasGapTimeS, HasVelocityMPS, HasPredictionTime bool
	ObservationCount, SatellitePositionCount       int
	BaseSatellitePositionCount                     int
	RoverSatellitePositionCount                    int
	VelocityMPS                                    [3]float64
}

type RtkRinexArc struct {
	_      noCopy
	handle *surfaceHandle
}

type RtkRinexDualFrequencyArc struct {
	_      noCopy
	handle *surfaceHandle
}

func newRtkRinexArc(pointer *C.SidereonRtkRinexArc) (*RtkRinexArc, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_rinex_arc_free((*C.SidereonRtkRinexArc)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkRinexArc{handle: handle}, nil
}

func newRtkRinexDualFrequencyArc(pointer *C.SidereonRtkRinexDualFrequencyArc) (*RtkRinexDualFrequencyArc, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_rinex_dual_frequency_arc_free((*C.SidereonRtkRinexDualFrequencyArc)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkRinexDualFrequencyArc{handle: handle}, nil
}

func copyRinexCode(dst []C.char, value, label string) error {
	if err := rejectEmbeddedNUL(value, label); err != nil {
		return err
	}
	if len(value) >= len(dst) {
		return fmt.Errorf("sidereon: %s is too long", label)
	}
	for index := range dst {
		dst[index] = 0
	}
	for index, item := range []byte(value) {
		dst[index] = C.char(item)
	}
	return nil
}

func copyRtkRinexSignalPairs(values []RtkRinexSignalPair, alloc *cRtkAlloc) (*C.SidereonRtkRinexSignalPair, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkRinexSignalPair{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK RINEX signal pairs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkRinexSignalPair)(memory), len(values))
	for index, value := range values {
		if err := validateGNSSSystemValue(value.System); err != nil {
			return nil, 0, err
		}
		rows[index] = C.SidereonRtkRinexSignalPair{system: C.uint32_t(value.System)}
		if err := copyRinexCode(rows[index].code_observable[:], value.CodeObservable, "RTK RINEX code observable"); err != nil {
			return nil, 0, err
		}
		if err := copyRinexCode(rows[index].phase_observable[:], value.PhaseObservable, "RTK RINEX phase observable"); err != nil {
			return nil, 0, err
		}
	}
	return &rows[0], count, nil
}

func copyRtkRinexDualSignalPairs(values []RtkRinexDualSignalPair, alloc *cRtkAlloc) (*C.SidereonRtkRinexDualSignalPair, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkRinexDualSignalPair{}))
	if err != nil {
		return nil, 0, err
	}
	memory, err := alloc.malloc(bytes, "RTK dual RINEX signal pairs")
	if err != nil {
		return nil, 0, err
	}
	rows := unsafe.Slice((*C.SidereonRtkRinexDualSignalPair)(memory), len(values))
	for index, value := range values {
		if err := validateGNSSSystemValue(value.System); err != nil {
			return nil, 0, err
		}
		rows[index] = C.SidereonRtkRinexDualSignalPair{system: C.uint32_t(value.System)}
		if err := copyRinexCode(rows[index].code1_observable[:], value.Code1Observable, "RTK dual RINEX band-1 code observable"); err != nil {
			return nil, 0, err
		}
		if err := copyRinexCode(rows[index].phase1_observable[:], value.Phase1Observable, "RTK dual RINEX band-1 phase observable"); err != nil {
			return nil, 0, err
		}
		if err := copyRinexCode(rows[index].code2_observable[:], value.Code2Observable, "RTK dual RINEX band-2 code observable"); err != nil {
			return nil, 0, err
		}
		if err := copyRinexCode(rows[index].phase2_observable[:], value.Phase2Observable, "RTK dual RINEX band-2 phase observable"); err != nil {
			return nil, 0, err
		}
	}
	return &rows[0], count, nil
}

func (value RtkRinexArcOptions) toC(alloc *cRtkAlloc) (*C.SidereonRtkRinexArcOptions, error) {
	signalPairs, signalPairCount, err := copyRtkRinexSignalPairs(value.SignalPairs, alloc)
	if err != nil {
		return nil, err
	}
	maxEpochs, err := checkedNativeSize(value.MaxEpochs)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK RINEX max epochs: %w", err)
	}
	minCommon, err := checkedNativeSize(value.MinCommonSatellites)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK RINEX minimum common satellites: %w", err)
	}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkRinexArcOptions{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK RINEX arc options")
	if err != nil {
		return nil, err
	}
	options := (*C.SidereonRtkRinexArcOptions)(memory)
	*options = C.SidereonRtkRinexArcOptions{signal_pairs: signalPairs, signal_pair_count: signalPairCount, has_max_epochs: C.bool(value.HasMaxEpochs), max_epochs: maxEpochs, min_common_satellites: minCommon, include_prediction_time: C.bool(value.IncludePredictionTime)}
	return options, nil
}

func (value RtkRinexDualArcOptions) toC(alloc *cRtkAlloc) (*C.SidereonRtkRinexDualArcOptions, error) {
	signalPairs, signalPairCount, err := copyRtkRinexDualSignalPairs(value.SignalPairs, alloc)
	if err != nil {
		return nil, err
	}
	maxEpochs, err := checkedNativeSize(value.MaxEpochs)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK dual RINEX max epochs: %w", err)
	}
	minCommon, err := checkedNativeSize(value.MinCommonSatellites)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK dual RINEX minimum common satellites: %w", err)
	}
	bytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkRinexDualArcOptions{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(bytes, "RTK dual RINEX arc options")
	if err != nil {
		return nil, err
	}
	options := (*C.SidereonRtkRinexDualArcOptions)(memory)
	*options = C.SidereonRtkRinexDualArcOptions{signal_pairs: signalPairs, signal_pair_count: signalPairCount, has_max_epochs: C.bool(value.HasMaxEpochs), max_epochs: maxEpochs, min_common_satellites: minCommon, include_prediction_time: C.bool(value.IncludePredictionTime)}
	return options, nil
}

type lockedRinexObs struct {
	obs *RinexObs
	key uintptr
	ptr unsafe.Pointer
}

func withRtkRinexObsPair(base, rover *RinexObs, fn func(unsafe.Pointer, unsafe.Pointer) error) error {
	if base == nil || base.resource == nil || rover == nil || rover.resource == nil {
		return ErrClosed
	}
	if base.resource == rover.resource {
		return base.resource.with(func(pointer unsafe.Pointer) error { return fn(pointer, pointer) })
	}
	items := [2]lockedRinexObs{
		{obs: base, key: uintptr(unsafe.Pointer(base.resource))},
		{obs: rover, key: uintptr(unsafe.Pointer(rover.resource))},
	}
	if items[1].key < items[0].key {
		items[0], items[1] = items[1], items[0]
	}
	var lock func(int) error
	lock = func(index int) error {
		if index == len(items) {
			var basePointer, roverPointer unsafe.Pointer
			for _, item := range items {
				if item.obs == base {
					basePointer = item.ptr
				} else {
					roverPointer = item.ptr
				}
			}
			return fn(basePointer, roverPointer)
		}
		return items[index].obs.resource.with(func(pointer unsafe.Pointer) error {
			items[index].ptr = pointer
			return lock(index + 1)
		})
	}
	return lock(0)
}

func withRtkRinexInputs(sp3 *SP3, base, rover *RinexObs, fn func(unsafe.Pointer, unsafe.Pointer, unsafe.Pointer) error) error {
	if sp3 == nil || sp3.handle == nil {
		return ErrClosed
	}
	return sp3.handle.with(func(sp3Pointer unsafe.Pointer) error {
		return withRtkRinexObsPair(base, rover, func(basePointer, roverPointer unsafe.Pointer) error {
			return fn(sp3Pointer, basePointer, roverPointer)
		})
	})
}

func BuildRtkRinexArc(sp3 *SP3, base, rover *RinexObs, options *RtkRinexArcOptions) (*RtkRinexArc, error) {
	if options == nil {
		value, err := RtkRinexArcOptionsInit()
		if err != nil {
			return nil, err
		}
		options = &value
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cOptions, err := options.toC(alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkRinexArc
	var operationErr error
	err = withRtkRinexInputs(sp3, base, rover, func(sp3Pointer, basePointer, roverPointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_build_rinex_rtk_arc((*C.SidereonSp3)(sp3Pointer), (*C.SidereonRinexObs)(basePointer), (*C.SidereonRinexObs)(roverPointer), cOptions, &pointer)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && pointer != nil {
				C.sidereon_rtk_rinex_arc_free(pointer)
				pointer = nil
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkRinexArc(pointer)
}

func BuildRtkRinexDualFrequencyArc(sp3 *SP3, base, rover *RinexObs, options *RtkRinexDualArcOptions) (*RtkRinexDualFrequencyArc, error) {
	if options == nil {
		value, err := RtkRinexDualArcOptionsInit()
		if err != nil {
			return nil, err
		}
		options = &value
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cOptions, err := options.toC(alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkRinexDualFrequencyArc
	var operationErr error
	err = withRtkRinexInputs(sp3, base, rover, func(sp3Pointer, basePointer, roverPointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_build_dual_frequency_rinex_rtk_arc((*C.SidereonSp3)(sp3Pointer), (*C.SidereonRinexObs)(basePointer), (*C.SidereonRinexObs)(roverPointer), cOptions, &pointer)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && pointer != nil {
				C.sidereon_rtk_rinex_dual_frequency_arc_free(pointer)
				pointer = nil
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkRinexDualFrequencyArc(pointer)
}

func (a *RtkRinexArc) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}

func (a *RtkRinexDualFrequencyArc) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}

func (a *RtkRinexArc) EpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_rinex_arc_epoch_count((*C.SidereonRtkRinexArc)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK RINEX arc epoch count")
}

func (a *RtkRinexArc) SkippedEpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_rinex_arc_skipped_epoch_count((*C.SidereonRtkRinexArc)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK RINEX arc skipped epoch count")
}

func (a *RtkRinexDualFrequencyArc) EpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_rinex_dual_frequency_arc_epoch_count((*C.SidereonRtkRinexDualFrequencyArc)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK dual RINEX arc epoch count")
}

func (a *RtkRinexDualFrequencyArc) SkippedEpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_rinex_dual_frequency_arc_skipped_epoch_count((*C.SidereonRtkRinexDualFrequencyArc)(pointer), &count))
		})
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RTK dual RINEX arc skipped epoch count")
}

func checkedRtkRinexIndex(index int, label string) (C.size_t, error) {
	if index < 0 {
		return 0, invalidArgument(label + " index must not be negative")
	}
	return cSize(index, label+" index")
}

type rtkRinexArcObservationCall func(*C.SidereonRtkArcObservationOut, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus
type rtkRinexArcPositionCall func(*C.SidereonRtkArcPositionOut, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus
type rtkRinexMapValueCall func(*C.SidereonRtkMapValue, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus
type rtkRinexDualObservationCall func(*C.SidereonRtkDualFrequencySatelliteObservationOut, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyRtkRinexArcObservations(label string, call rtkRinexArcObservationCall) ([]RtkRinexArcObservation, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkArcObservationOut{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkArcObservationOut, count)
	var output *C.SidereonRtkArcObservationOut
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkRinexArcObservation, writtenCount)
	for index := range result {
		result[index] = RtkRinexArcObservation{SatelliteID: tokenFromC(rows[index].sat_id), AmbiguityID: rtkIDFromC(rows[index].ambiguity_id).Value, CodeM: float64(rows[index].code_m), PhaseM: float64(rows[index].phase_m), HasLLI: bool(rows[index].has_lli), LLI: int64(rows[index].lli)}
	}
	return result, nil
}

func copyRtkRinexArcPositions(label string, call rtkRinexArcPositionCall) ([]RtkRinexArcPosition, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkArcPositionOut{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkArcPositionOut, count)
	var output *C.SidereonRtkArcPositionOut
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkRinexArcPosition, writtenCount)
	for index := range result {
		result[index] = RtkRinexArcPosition{SatelliteID: tokenFromC(rows[index].id)}
		for axis := range result[index].PositionM {
			result[index].PositionM[axis] = float64(rows[index].pos[axis])
		}
	}
	return result, nil
}

func copyRtkRinexMapValues(label string, call rtkRinexMapValueCall) ([]RtkRinexMapValue, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkMapValue{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkMapValue, count)
	var output *C.SidereonRtkMapValue
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkRinexMapValue, writtenCount)
	for index := range result {
		result[index] = RtkRinexMapValue{ID: rtkIDFromC(rows[index].id).Value, Value: float64(rows[index].value)}
	}
	return result, nil
}

func copyRtkRinexDualObservations(label string, call rtkRinexDualObservationCall) ([]RtkRinexDualFrequencySatelliteObservation, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkDualFrequencySatelliteObservationOut{})); err != nil {
		return nil, err
	}
	capacity, err := cSize(count, label+" output capacity")
	if err != nil {
		return nil, err
	}
	rows := make([]C.SidereonRtkDualFrequencySatelliteObservationOut, count)
	var output *C.SidereonRtkDualFrequencySatelliteObservationOut
	if count != 0 {
		output = &rows[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, capacity, &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]RtkRinexDualFrequencySatelliteObservation, writtenCount)
	for index := range result {
		result[index] = RtkRinexDualFrequencySatelliteObservation{SatelliteID: tokenFromC(rows[index].sat_id), Base: dualObservationFromC(rows[index].base), Rover: dualObservationFromC(rows[index].rover)}
	}
	return result, nil
}

func dualObservationFromC(value C.SidereonRtkDualFrequencyObservationOut) RtkRinexDualFrequencyObservation {
	return RtkRinexDualFrequencyObservation{AmbiguityID: rtkIDFromC(value.ambiguity_id).Value, P1M: float64(value.p1_m), P2M: float64(value.p2_m), Phi1Cycles: float64(value.phi1_cycles), Phi2Cycles: float64(value.phi2_cycles), F1Hz: float64(value.f1_hz), F2Hz: float64(value.f2_hz), HasLLI1: bool(value.has_lli1), LLI1: int64(value.lli1), HasLLI2: bool(value.has_lli2), LLI2: int64(value.lli2)}
}

func (a *RtkRinexArc) epochObservations(index int, base bool) ([]RtkRinexArcObservation, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK RINEX epoch")
	if err != nil {
		return nil, err
	}
	var result []RtkRinexArcObservation
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexArcObservations("RTK RINEX arc observations", func(out *C.SidereonRtkArcObservationOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				if base {
					return C.sidereon_rtk_rinex_arc_epoch_base_observations((*C.SidereonRtkRinexArc)(pointer), nativeIndex, out, length, written, required)
				}
				return C.sidereon_rtk_rinex_arc_epoch_rover_observations((*C.SidereonRtkRinexArc)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (a *RtkRinexArc) EpochBaseObservations(index int) ([]RtkRinexArcObservation, error) {
	return a.epochObservations(index, true)
}

func (a *RtkRinexArc) EpochRoverObservations(index int) ([]RtkRinexArcObservation, error) {
	return a.epochObservations(index, false)
}

func (a *RtkRinexArc) epochPositions(index int, kind uint32) ([]RtkRinexArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK RINEX epoch")
	if err != nil {
		return nil, err
	}
	var result []RtkRinexArcPosition
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexArcPositions("RTK RINEX arc positions", func(out *C.SidereonRtkArcPositionOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				switch kind {
				case 0:
					return C.sidereon_rtk_rinex_arc_epoch_satellite_positions((*C.SidereonRtkRinexArc)(pointer), nativeIndex, out, length, written, required)
				case 1:
					return C.sidereon_rtk_rinex_arc_epoch_base_satellite_positions((*C.SidereonRtkRinexArc)(pointer), nativeIndex, out, length, written, required)
				default:
					return C.sidereon_rtk_rinex_arc_epoch_rover_satellite_positions((*C.SidereonRtkRinexArc)(pointer), nativeIndex, out, length, written, required)
				}
			})
		})
		return err
	})
	return result, err
}

func (a *RtkRinexArc) EpochSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	return a.epochPositions(index, 0)
}

func (a *RtkRinexArc) EpochBaseSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	return a.epochPositions(index, 1)
}

func (a *RtkRinexArc) EpochRoverSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	return a.epochPositions(index, 2)
}

func (a *RtkRinexArc) EpochMetadata(index int) (RtkRinexArcEpochMetadata, error) {
	if a == nil || a.handle == nil {
		return RtkRinexArcEpochMetadata{}, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK RINEX epoch")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	var value C.SidereonRtkArcEpochOutMetadata
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_rinex_arc_epoch_metadata((*C.SidereonRtkRinexArc)(pointer), nativeIndex, &value))
		})
	})
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	baseCount, err := sizeTToInt(value.base_count, "RTK RINEX base observation count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	roverCount, err := sizeTToInt(value.rover_count, "RTK RINEX rover observation count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	positions, err := sizeTToInt(value.satellite_position_count, "RTK RINEX satellite position count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	basePositions, err := sizeTToInt(value.base_satellite_position_count, "RTK RINEX base position count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	roverPositions, err := sizeTToInt(value.rover_satellite_position_count, "RTK RINEX rover position count")
	if err != nil {
		return RtkRinexArcEpochMetadata{}, err
	}
	return RtkRinexArcEpochMetadata{BaseCount: baseCount, RoverCount: roverCount, SatellitePositionCount: positions, BaseSatellitePositionCount: basePositions, RoverSatellitePositionCount: roverPositions, HasVelocityMPS: bool(value.has_velocity_mps), VelocityMPS: [3]float64{float64(value.velocity_mps[0]), float64(value.velocity_mps[1]), float64(value.velocity_mps[2])}, HasPredictionTime: bool(value.has_prediction_time), PredictionTimeS: float64(value.prediction_time_s)}, nil
}

func (a *RtkRinexArc) mapValues(offsets bool) ([]RtkRinexMapValue, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkRinexMapValue
	var operationErr error
	err := a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyRtkRinexMapValues("RTK RINEX map values", func(out *C.SidereonRtkMapValue, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				if offsets {
					return C.sidereon_rtk_rinex_arc_offsets_m((*C.SidereonRtkRinexArc)(pointer), out, length, written, required)
				}
				return C.sidereon_rtk_rinex_arc_wavelengths_m((*C.SidereonRtkRinexArc)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}

func (a *RtkRinexArc) OffsetsM() ([]RtkRinexMapValue, error)     { return a.mapValues(true) }
func (a *RtkRinexArc) WavelengthsM() ([]RtkRinexMapValue, error) { return a.mapValues(false) }

func (a *RtkRinexDualFrequencyArc) EpochMetadata(index int) (RtkRinexDualFrequencyArcEpochMetadata, error) {
	if a == nil || a.handle == nil {
		return RtkRinexDualFrequencyArcEpochMetadata{}, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK dual RINEX epoch")
	if err != nil {
		return RtkRinexDualFrequencyArcEpochMetadata{}, err
	}
	var value C.SidereonRtkRinexDualFrequencyArcEpochOutMetadata
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_rinex_dual_frequency_arc_epoch_metadata((*C.SidereonRtkRinexDualFrequencyArc)(pointer), nativeIndex, &value))
		})
	})
	if err != nil {
		return RtkRinexDualFrequencyArcEpochMetadata{}, err
	}
	observations, err := sizeTToInt(value.observation_count, "RTK dual RINEX observation count")
	if err != nil {
		return RtkRinexDualFrequencyArcEpochMetadata{}, err
	}
	positions, err := sizeTToInt(value.satellite_position_count, "RTK dual RINEX satellite position count")
	if err != nil {
		return RtkRinexDualFrequencyArcEpochMetadata{}, err
	}
	basePositions, err := sizeTToInt(value.base_satellite_position_count, "RTK dual RINEX base position count")
	if err != nil {
		return RtkRinexDualFrequencyArcEpochMetadata{}, err
	}
	roverPositions, err := sizeTToInt(value.rover_satellite_position_count, "RTK dual RINEX rover position count")
	if err != nil {
		return RtkRinexDualFrequencyArcEpochMetadata{}, err
	}
	return RtkRinexDualFrequencyArcEpochMetadata{JDWhole: float64(value.jd_whole), JDFraction: float64(value.jd_fraction), HasGapTimeS: bool(value.has_gap_time_s), GapTimeS: float64(value.gap_time_s), ObservationCount: observations, SatellitePositionCount: positions, BaseSatellitePositionCount: basePositions, RoverSatellitePositionCount: roverPositions, HasVelocityMPS: bool(value.has_velocity_mps), VelocityMPS: [3]float64{float64(value.velocity_mps[0]), float64(value.velocity_mps[1]), float64(value.velocity_mps[2])}, HasPredictionTime: bool(value.has_prediction_time), PredictionTimeS: float64(value.prediction_time_s)}, nil
}

func (a *RtkRinexDualFrequencyArc) EpochBaseSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK dual RINEX epoch")
	if err != nil {
		return nil, err
	}
	var result []RtkRinexArcPosition
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexArcPositions("RTK dual RINEX base positions", func(out *C.SidereonRtkArcPositionOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_rinex_dual_frequency_arc_epoch_base_satellite_positions((*C.SidereonRtkRinexDualFrequencyArc)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (a *RtkRinexDualFrequencyArc) epochPositions(index int, base bool) ([]RtkRinexArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK dual RINEX epoch")
	if err != nil {
		return nil, err
	}
	var result []RtkRinexArcPosition
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexArcPositions("RTK dual RINEX positions", func(out *C.SidereonRtkArcPositionOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				if base {
					return C.sidereon_rtk_rinex_dual_frequency_arc_epoch_base_satellite_positions((*C.SidereonRtkRinexDualFrequencyArc)(pointer), nativeIndex, out, length, written, required)
				}
				return C.sidereon_rtk_rinex_dual_frequency_arc_epoch_rover_satellite_positions((*C.SidereonRtkRinexDualFrequencyArc)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (a *RtkRinexDualFrequencyArc) EpochRoverSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	return a.epochPositions(index, false)
}

func (a *RtkRinexDualFrequencyArc) EpochSatellitePositions(index int) ([]RtkRinexArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK dual RINEX epoch")
	if err != nil {
		return nil, err
	}
	var result []RtkRinexArcPosition
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexArcPositions("RTK dual RINEX positions", func(out *C.SidereonRtkArcPositionOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_rinex_dual_frequency_arc_epoch_satellite_positions((*C.SidereonRtkRinexDualFrequencyArc)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (a *RtkRinexDualFrequencyArc) EpochObservations(index int) ([]RtkRinexDualFrequencySatelliteObservation, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK dual RINEX epoch")
	if err != nil {
		return nil, err
	}
	var result []RtkRinexDualFrequencySatelliteObservation
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyRtkRinexDualObservations("RTK dual RINEX observations", func(out *C.SidereonRtkDualFrequencySatelliteObservationOut, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_rinex_dual_frequency_arc_epoch_observations((*C.SidereonRtkRinexDualFrequencyArc)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return result, err
}

func (a *RtkRinexDualFrequencyArc) EpochSortKey(index int) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	nativeIndex, err := checkedRtkRinexIndex(index, "RTK dual RINEX epoch")
	if err != nil {
		return "", err
	}
	var result []byte
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, err = copyNativeBytesLocked("RTK dual RINEX epoch sort key", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_rinex_dual_frequency_arc_epoch_sort_key((*C.SidereonRtkRinexDualFrequencyArc)(pointer), nativeIndex, out, length, written, required)
			})
		})
		return err
	})
	return string(result), err
}
