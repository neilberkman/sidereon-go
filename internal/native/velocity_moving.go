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

const (
	VelocityObservableRangeRateValue = uint32(C.SIDEREON_VELOCITY_OBSERVABLE_RANGE_RATE)
	VelocityObservableDopplerValue   = uint32(C.SIDEREON_VELOCITY_OBSERVABLE_DOPPLER)

	MovingBaselineStatusFixedValue    = uint32(C.SIDEREON_MOVING_BASELINE_STATUS_FIXED)
	MovingBaselineStatusFloatValue    = uint32(C.SIDEREON_MOVING_BASELINE_STATUS_FLOAT)
	RtkSolveStatusStateToleranceValue = uint32(C.SIDEREON_RTK_SOLVE_STATUS_STATE_TOLERANCE)
	RtkSolveStatusMaxIterationsValue  = uint32(C.SIDEREON_RTK_SOLVE_STATUS_MAX_ITERATIONS)
	RtkIntegerStatusFixedValue        = uint32(C.SIDEREON_RTK_INTEGER_STATUS_FIXED)
	RtkIntegerStatusNotFixedValue     = uint32(C.SIDEREON_RTK_INTEGER_STATUS_NOT_FIXED)
	RtkStochasticSimpleValue          = uint32(C.SIDEREON_RTK_STOCHASTIC_MODEL_SIMPLE)
	RtkStochasticRTKLIBValue          = uint32(C.SIDEREON_RTK_STOCHASTIC_MODEL_RTKLIB)
)

// VelocityObservation is one copied range-rate or Doppler measurement.
type VelocityObservation struct {
	SatelliteID         string
	Value               float64
	CarrierHz           float64
	SatelliteClockDrift float64
}

// VelocityOptions controls the native velocity solve.
type VelocityOptions struct {
	Observable uint32
	LightTime  bool
	Sagnac     bool
}

// VelocitySolution owns a native receiver-velocity solution.
type VelocitySolution struct {
	_      noCopy
	handle *positioningHandle
}

func validateVelocityObservable(value uint32) error {
	if value > VelocityObservableDopplerValue {
		return invalidArgument("velocity observable is not defined")
	}
	return nil
}

func validateMovingBaselineStatus(value uint32) error {
	if value > MovingBaselineStatusFloatValue {
		return invalidArgument("moving-baseline status is not defined")
	}
	return nil
}

func validateRtkSolveStatus(value uint32) error {
	if value > RtkSolveStatusMaxIterationsValue {
		return invalidArgument("RTK solve status is not defined")
	}
	return nil
}

func validateRtkIntegerStatus(value uint32) error {
	if value > RtkIntegerStatusNotFixedValue {
		return invalidArgument("RTK integer status is not defined")
	}
	return nil
}

func validateRtkStochastic(value uint32) error {
	if value > RtkStochasticRTKLIBValue {
		return invalidArgument("RTK stochastic model is not defined")
	}
	return nil
}

func releaseVelocitySolution(pointer unsafe.Pointer) {
	C.sidereon_velocity_solution_free((*C.SidereonVelocitySolution)(pointer))
}

func releaseMovingBaselineSolution(pointer unsafe.Pointer) {
	C.sidereon_moving_baseline_solution_free((*C.SidereonMovingBaselineSolution)(pointer))
}

func VelocityOptionsDefaults() (VelocityOptions, error) {
	var value C.SidereonVelocityOptions
	var err error
	withCThread(func() {
		err = statusErrorLocked(C.sidereon_velocity_options_init(&value))
	})
	if err != nil {
		return VelocityOptions{}, err
	}
	if err := validateVelocityObservable(uint32(value.observable)); err != nil {
		return VelocityOptions{}, err
	}
	return VelocityOptions{Observable: uint32(value.observable), LightTime: bool(value.light_time), Sagnac: bool(value.sagnac)}, nil
}

func cVelocityOptions(value VelocityOptions) (C.SidereonVelocityOptions, error) {
	if err := validateVelocityObservable(value.Observable); err != nil {
		return C.SidereonVelocityOptions{}, err
	}
	return C.SidereonVelocityOptions{observable: C.uint32_t(value.Observable), light_time: C.bool(value.LightTime), sagnac: C.bool(value.Sagnac)}, nil
}

type velocityObservationMemory struct {
	array   unsafe.Pointer
	strings []unsafe.Pointer
}

func (m *velocityObservationMemory) free() {
	for _, pointer := range m.strings {
		if pointer != nil {
			C.free(pointer)
		}
	}
	if m.array != nil {
		C.free(m.array)
	}
}

func newVelocityObservations(values []VelocityObservation) (*velocityObservationMemory, *C.SidereonVelocityObservation, C.size_t, error) {
	if _, err := checkedNativeSize(len(values)); err != nil {
		return nil, nil, 0, err
	}
	memory := &velocityObservationMemory{}
	if len(values) == 0 {
		return memory, nil, 0, nil
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonVelocityObservation{}))
	if err != nil {
		return nil, nil, 0, err
	}
	memory.array = C.malloc(C.size_t(size))
	if memory.array == nil {
		return nil, nil, 0, errors.New("sidereon: unable to allocate native velocity observations")
	}
	rows := unsafe.Slice((*C.SidereonVelocityObservation)(memory.array), len(values))
	memory.strings = make([]unsafe.Pointer, 0, len(values))
	for i, value := range values {
		pointer, err := copyNativeCString(value.SatelliteID, "velocity satellite ID")
		if err != nil {
			memory.free()
			return nil, nil, 0, err
		}
		memory.strings = append(memory.strings, pointer)
		rows[i].sat_id = (*C.char)(pointer)
		rows[i].value = C.double(value.Value)
		rows[i].carrier_hz = C.double(value.CarrierHz)
		rows[i].sat_clock_drift_s_s = C.double(value.SatelliteClockDrift)
	}
	count, err := checkedNativeSize(len(values))
	if err != nil {
		memory.free()
		return nil, nil, 0, err
	}
	return memory, &rows[0], count, nil
}

func solveVelocityWith(observations []VelocityObservation, receiver [3]float64, tRx float64, options *VelocityOptions, call func(*C.SidereonVelocityObservation, C.size_t, *C.double, C.double, *C.SidereonVelocityOptions, **C.SidereonVelocitySolution) C.enum_SidereonStatus) (*VelocitySolution, error) {
	var nativeOptions *C.SidereonVelocityOptions
	var optionsValue C.SidereonVelocityOptions
	var err error
	if options != nil {
		optionsValue, err = cVelocityOptions(*options)
		if err != nil {
			return nil, err
		}
		nativeOptions = &optionsValue
	}
	var memory *velocityObservationMemory
	var rows *C.SidereonVelocityObservation
	var count C.size_t
	withCThread(func() {
		memory, rows, count, err = newVelocityObservations(observations)
	})
	if err != nil {
		return nil, err
	}
	defer memory.free()
	var solution *C.SidereonVelocitySolution
	withCThread(func() {
		err = statusErrorLocked(call(rows, count, (*C.double)(unsafe.Pointer(&receiver[0])), C.double(tRx), nativeOptions, &solution))
		if err != nil && solution != nil {
			releaseVelocitySolution(unsafe.Pointer(solution))
			solution = nil
		}
	})
	if err != nil {
		return nil, err
	}
	if solution == nil {
		return nil, missingNativeHandle("velocity solve")
	}
	return &VelocitySolution{handle: newPositioningHandle(unsafe.Pointer(solution), releaseVelocitySolution)}, nil
}

func SolveVelocity(sp3 *SP3, observations []VelocityObservation, receiver [3]float64, tRx float64, options *VelocityOptions) (*VelocitySolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var result *VelocitySolution
	var operationErr error
	err := sp3.handle.with(func(pointer unsafe.Pointer) error {
		result, operationErr = solveVelocityWith(observations, receiver, tRx, options, func(rows *C.SidereonVelocityObservation, count C.size_t, position *C.double, epoch C.double, opts *C.SidereonVelocityOptions, output **C.SidereonVelocitySolution) C.enum_SidereonStatus {
			return C.sidereon_solve_velocity((*C.SidereonSp3)(pointer), rows, count, position, epoch, opts, output)
		})
		return operationErr
	})
	runtime.KeepAlive(sp3)
	return result, err
}

func SolveVelocityBroadcast(broadcast *BroadcastEphemeris, observations []VelocityObservation, receiver [3]float64, tRx float64, options *VelocityOptions) (*VelocitySolution, error) {
	if broadcast == nil || broadcast.resource == nil {
		return nil, ErrClosed
	}
	var result *VelocitySolution
	var operationErr error
	err := broadcast.resource.with(func(pointer unsafe.Pointer) error {
		result, operationErr = solveVelocityWith(observations, receiver, tRx, options, func(rows *C.SidereonVelocityObservation, count C.size_t, position *C.double, epoch C.double, opts *C.SidereonVelocityOptions, output **C.SidereonVelocitySolution) C.enum_SidereonStatus {
			return C.sidereon_solve_velocity_broadcast((*C.SidereonBroadcastEphemeris)(pointer), rows, count, position, epoch, opts, output)
		})
		return operationErr
	})
	runtime.KeepAlive(broadcast)
	return result, err
}

func (s *VelocitySolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *VelocitySolution) ClockDrift() (float64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var value C.double
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = statusErrorLocked(C.sidereon_velocity_solution_clock_drift((*C.SidereonVelocitySolution)(pointer), &value))
		})
		return operationErr
	})
	runtime.KeepAlive(s)
	return float64(value), err
}

func (s *VelocitySolution) Speed() (float64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var value C.double
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = statusErrorLocked(C.sidereon_velocity_solution_speed((*C.SidereonVelocitySolution)(pointer), &value))
		})
		return operationErr
	})
	runtime.KeepAlive(s)
	return float64(value), err
}

func (s *VelocitySolution) Velocity() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	var values [3]C.double
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = statusErrorLocked(C.sidereon_velocity_solution_velocity((*C.SidereonVelocitySolution)(pointer), &values[0], 3))
		})
		return operationErr
	})
	runtime.KeepAlive(s)
	var result [3]float64
	for i := range result {
		result[i] = float64(values[i])
	}
	return result, err
}

func (s *VelocitySolution) StateCovariance() ([16]float64, error) {
	if s == nil || s.handle == nil {
		return [16]float64{}, ErrClosed
	}
	var values [16]C.double
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = statusErrorLocked(C.sidereon_velocity_solution_state_covariance((*C.SidereonVelocitySolution)(pointer), &values[0], 16))
		})
		return operationErr
	})
	runtime.KeepAlive(s)
	var result [16]float64
	for i := range result {
		result[i] = float64(values[i])
	}
	return result, err
}

func (s *VelocitySolution) Residuals() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var values []C.double
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		var queryErr error
		withCThread(func() {
			queryErr = statusErrorLocked(C.sidereon_velocity_solution_residuals((*C.SidereonVelocitySolution)(pointer), nil, 0, &written, &required))
		})
		if queryErr != nil {
			return queryErr
		}
		count, err := validateNativeQuery("velocity residuals", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.double(0))); err != nil {
			return err
		}
		values = make([]C.double, count)
		var output *C.double
		if count != 0 {
			output = &values[0]
		}
		written, required = 0, 0
		var fillErr error
		withCThread(func() {
			fillErr = statusErrorLocked(C.sidereon_velocity_solution_residuals((*C.SidereonVelocitySolution)(pointer), output, C.size_t(count), &written, &required))
		})
		if fillErr != nil {
			return fillErr
		}
		writtenCount, err := validateTwoPassCounts("velocity residuals", len(values), count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		values = values[:writtenCount]
		return nil
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	result := make([]float64, len(values))
	for i := range values {
		result[i] = float64(values[i])
	}
	return result, nil
}

func (s *VelocitySolution) UsedSatelliteCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = statusErrorLocked(C.sidereon_velocity_solution_used_sat_count((*C.SidereonVelocitySolution)(pointer), &count))
		})
		return operationErr
	})
	runtime.KeepAlive(s)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (s *VelocitySolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var values []C.SidereonSatelliteToken
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		var queryErr error
		withCThread(func() {
			queryErr = statusErrorLocked(C.sidereon_velocity_solution_used_sat_ids((*C.SidereonVelocitySolution)(pointer), nil, 0, &written, &required))
		})
		if queryErr != nil {
			return queryErr
		}
		count, err := validateNativeQuery("velocity satellite IDs", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
			return err
		}
		values = make([]C.SidereonSatelliteToken, count)
		var output *C.SidereonSatelliteToken
		if count != 0 {
			output = &values[0]
		}
		written, required = 0, 0
		var fillErr error
		withCThread(func() {
			fillErr = statusErrorLocked(C.sidereon_velocity_solution_used_sat_ids((*C.SidereonVelocitySolution)(pointer), output, C.size_t(count), &written, &required))
		})
		if fillErr != nil {
			return fillErr
		}
		writtenCount, err := validateTwoPassCounts("velocity satellite IDs", len(values), count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		values = values[:writtenCount]
		return nil
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(values))
	for i := range values {
		result[i] = tokenFromC(values[i])
	}
	return result, nil
}

// MovingBaselineStatus reports whether an epoch used the float or fixed baseline.
type MovingBaselineStatus uint32

// MovingBaselineEpochSummary is one detached moving-baseline result.
type MovingBaselineEpochSummary struct {
	BasePositionM   [3]float64
	BaselineM       [3]float64
	BaselineLengthM float64
	Status          MovingBaselineStatus
	Float           RtkFloatMetadata
	Fixed           RtkFixedMetadata
}

func movingBaselineSummaryFromC(value C.SidereonMovingBaselineEpochSummary) (MovingBaselineEpochSummary, error) {
	if err := validateMovingBaselineStatus(uint32(value.status)); err != nil {
		return MovingBaselineEpochSummary{}, err
	}
	floatMetadata, err := rtkFloatMetadataFromC(value.float_)
	if err != nil {
		return MovingBaselineEpochSummary{}, err
	}
	fixedMetadata, err := rtkFixedMetadataFromC(value.fixed)
	if err != nil {
		return MovingBaselineEpochSummary{}, err
	}
	var result MovingBaselineEpochSummary
	for i := 0; i < 3; i++ {
		result.BasePositionM[i] = float64(value.base_position_m[i])
		result.BaselineM[i] = float64(value.baseline_m[i])
	}
	result.BaselineLengthM = float64(value.baseline_length_m)
	result.Status = MovingBaselineStatus(value.status)
	result.Float = floatMetadata
	result.Fixed = fixedMetadata
	return result, nil
}

// MovingBaselineSolution owns a native moving-baseline result.
type MovingBaselineSolution struct {
	_      noCopy
	handle *positioningHandle
}

func (s *MovingBaselineSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *MovingBaselineSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = statusErrorLocked(C.sidereon_moving_baseline_solution_epoch_count((*C.SidereonMovingBaselineSolution)(pointer), &count))
		})
		return operationErr
	})
	runtime.KeepAlive(s)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (s *MovingBaselineSolution) Epoch(index int) (MovingBaselineEpochSummary, error) {
	if s == nil || s.handle == nil {
		return MovingBaselineEpochSummary{}, ErrClosed
	}
	if index < 0 {
		return MovingBaselineEpochSummary{}, errors.New("sidereon: moving-baseline epoch index must not be negative")
	}
	position, err := checkedNativeSize(index)
	if err != nil {
		return MovingBaselineEpochSummary{}, err
	}
	var value C.SidereonMovingBaselineEpochSummary
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			err = statusErrorLocked(C.sidereon_moving_baseline_solution_epoch((*C.SidereonMovingBaselineSolution)(pointer), position, &value))
		})
		return err
	})
	runtime.KeepAlive(s)
	if err != nil {
		return MovingBaselineEpochSummary{}, err
	}
	return movingBaselineSummaryFromC(value)
}

// MovingBaselineEpoch is one moving-baseline input epoch.
type MovingBaselineEpoch struct {
	BasePositionM       [3]float64
	Epoch               RtkEpoch
	AmbiguityIDs        []string
	AmbiguitySatellites []RtkAmbiguitySatellite
	WavelengthsM        []RtkFloatMapEntry
	OffsetsM            []RtkFloatMapEntry
	FloatOnlySystems    []uint32
}

// MovingBaselineConfig describes a complete native moving-baseline solve.
type MovingBaselineConfig struct {
	Epochs           []MovingBaselineEpoch
	Model            RtkMeasurementModel
	FloatOptions     RtkFloatOptions
	FixedOptions     RtkFixedOptions
	InitialBaselineM [3]float64
	WarmStart        bool
	ReceiverAntenna  *RtkReceiverAntennaCorrections
}

func callocRtk(alloc *cRtkAlloc, size uintptr, label string) (unsafe.Pointer, error) {
	if size == 0 {
		return nil, nil
	}
	p := C.calloc(1, C.size_t(size))
	if p == nil {
		return nil, fmt.Errorf("sidereon: unable to allocate native %s", label)
	}
	alloc.values = append(alloc.values, p)
	return p, nil
}

func movingBaselineInput(value MovingBaselineConfig) (*cRtkAlloc, *C.SidereonMovingBaselineConfig, error) {
	alloc := new(cRtkAlloc)
	configMemory, err := callocRtk(alloc, unsafe.Sizeof(C.SidereonMovingBaselineConfig{}), "moving-baseline config")
	if err != nil {
		return nil, nil, err
	}
	config := (*C.SidereonMovingBaselineConfig)(configMemory)
	if err := validateRtkStochastic(value.Model.Stochastic); err != nil {
		alloc.close()
		return nil, nil, err
	}
	config.model = C.SidereonRtkMeasurementModel{code_sigma_m: C.double(value.Model.CodeSigmaM), phase_sigma_m: C.double(value.Model.PhaseSigmaM), sagnac: C.bool(value.Model.Sagnac), stochastic: C.uint32_t(value.Model.Stochastic), elevation_weighting: C.bool(value.Model.ElevationWeighting)}
	floatIterations, err := checkedNativeSize(value.FloatOptions.MaxIterations)
	if err != nil {
		alloc.close()
		return nil, nil, err
	}
	fixedIterations, err := checkedNativeSize(value.FixedOptions.MaxIterations)
	if err != nil {
		alloc.close()
		return nil, nil, err
	}
	partialMinimum, err := checkedNativeSize(value.FixedOptions.PartialMinAmbiguities)
	if err != nil {
		alloc.close()
		return nil, nil, err
	}
	config.float_options = C.SidereonRtkFloatOptions{position_tol_m: C.double(value.FloatOptions.PositionTolM), ambiguity_tol_m: C.double(value.FloatOptions.AmbiguityTolM), max_iterations: floatIterations}
	config.fixed_options = C.SidereonRtkFixedOptions{position_tol_m: C.double(value.FixedOptions.PositionTolM), ambiguity_tol_m: C.double(value.FixedOptions.AmbiguityTolM), max_iterations: fixedIterations, ratio_threshold: C.double(value.FixedOptions.RatioThreshold), partial_ambiguity_resolution: C.bool(value.FixedOptions.PartialAmbiguityResolution), partial_min_ambiguities: partialMinimum}
	for i := 0; i < 3; i++ {
		config.initial_baseline_m[i] = C.double(value.InitialBaselineM[i])
	}
	config.warm_start = C.bool(value.WarmStart)
	epochCount, err := checkedNativeSize(len(value.Epochs))
	if err != nil {
		alloc.close()
		return nil, nil, err
	}
	if len(value.Epochs) > 0 {
		rtkEpochs := make([]RtkEpoch, len(value.Epochs))
		for i, input := range value.Epochs {
			rtkEpochs[i] = input.Epoch
		}
		rtkEpochPointer, err := copyRtkEpochs(rtkEpochs, alloc)
		if err != nil {
			alloc.close()
			return nil, nil, err
		}
		rtkEpochRows := unsafe.Slice(rtkEpochPointer, len(value.Epochs))
		epochBytes, err := checkedNativeAllocationSize(len(value.Epochs), unsafe.Sizeof(C.SidereonMovingBaselineEpoch{}))
		if err != nil {
			alloc.close()
			return nil, nil, err
		}
		epochMemory, err := callocRtk(alloc, epochBytes, "moving-baseline epochs")
		if err != nil {
			alloc.close()
			return nil, nil, err
		}
		epochs := unsafe.Slice((*C.SidereonMovingBaselineEpoch)(epochMemory), len(value.Epochs))
		for i, input := range value.Epochs {
			ambiguityIDs, ambiguityIDCount, err := copyRtkAmbiguityIDs(input.AmbiguityIDs, alloc)
			if err != nil {
				alloc.close()
				return nil, nil, err
			}
			ambiguitySatellites, ambiguitySatelliteCount, err := copyRtkAmbiguitySatellites(input.AmbiguitySatellites, alloc)
			if err != nil {
				alloc.close()
				return nil, nil, err
			}
			wavelengths, wavelengthCount, err := copyRtkFloatMapEntries(input.WavelengthsM, alloc, "moving-baseline wavelengths")
			if err != nil {
				alloc.close()
				return nil, nil, err
			}
			offsets, offsetCount, err := copyRtkFloatMapEntries(input.OffsetsM, alloc, "moving-baseline offsets")
			if err != nil {
				alloc.close()
				return nil, nil, err
			}
			floatOnlySystems := make([]uint32, len(input.FloatOnlySystems))
			for j, system := range input.FloatOnlySystems {
				floatOnlySystems[j] = uint32(system)
			}
			floatOnlySystemPointer, floatOnlySystemCount, err := copyRtkFloatOnlySystems(floatOnlySystems, alloc)
			if err != nil {
				alloc.close()
				return nil, nil, err
			}
			epochs[i] = C.SidereonMovingBaselineEpoch{
				base_position_m: [3]C.double{C.double(input.BasePositionM[0]), C.double(input.BasePositionM[1]), C.double(input.BasePositionM[2])},
				epoch:           rtkEpochRows[i], ambiguity_ids: ambiguityIDs, ambiguity_id_count: ambiguityIDCount,
				ambiguity_satellites: ambiguitySatellites, ambiguity_satellite_count: ambiguitySatelliteCount,
				wavelengths_m: wavelengths, wavelength_count: wavelengthCount, offsets_m: offsets, offset_count: offsetCount,
				float_only_systems: floatOnlySystemPointer, float_only_system_count: floatOnlySystemCount,
			}
		}
		config.epochs, config.epoch_count = &epochs[0], epochCount
	}
	if value.ReceiverAntenna != nil {
		antenna, err := copyRtkReceiverAntenna(value.ReceiverAntenna, alloc)
		if err != nil {
			alloc.close()
			return nil, nil, err
		}
		config.receiver_antenna = antenna
	}
	return alloc, config, nil
}

func SolveMovingBaseline(value MovingBaselineConfig) (*MovingBaselineSolution, error) {
	var memory *cRtkAlloc
	var config *C.SidereonMovingBaselineConfig
	var err error
	withCThread(func() { memory, config, err = movingBaselineInput(value) })
	if err != nil {
		return nil, err
	}
	defer memory.close()
	var solution *C.SidereonMovingBaselineSolution
	withCThread(func() {
		err = statusErrorLocked(C.sidereon_solve_moving_baseline(config, &solution))
		if err != nil && solution != nil {
			releaseMovingBaselineSolution(unsafe.Pointer(solution))
			solution = nil
		}
	})
	if err != nil {
		return nil, err
	}
	if solution == nil {
		return nil, missingNativeHandle("moving-baseline solve")
	}
	return &MovingBaselineSolution{handle: newPositioningHandle(unsafe.Pointer(solution), releaseMovingBaselineSolution)}, nil
}
