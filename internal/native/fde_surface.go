//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type FDESolution struct {
	_      noCopy
	handle *surfaceHandle
}
type NativeFDEOptions struct {
	PFA              float64
	MaxExclusions    *uint64
	MaxExclusionRMSM *float64
	WeightsMode      uint32
	Weights          []NativeFDERaimWeight
	SystemsEnabled   bool
	Systems          int64
	UseValidation    bool
	HasUseValidation bool
}
type NativeFDEOutput struct {
	Iterations uint64
	Excluded   []string
}

type NativeFDEUnresolvedError struct {
	Reason        uint32
	Solution      SPPSolution
	HasSolution   bool
	Excluded      []string
	HasExcluded   bool
	RAIM          NativeRaimResult
	Normalized    []NativeRaimNormalizedResidual
	HasNormalized bool
	HasRAIM       bool
	Cause         error
	CaptureError  error
}

func (e *NativeFDEUnresolvedError) Error() string {
	return fmt.Sprintf("sidereon: FDE unresolved (reason %d)", e.Reason)
}
func (e *NativeFDEUnresolvedError) Unwrap() error { return e.Cause }

func copyFDERaimRowsLocked(call func(*C.SidereonRaimNormalizedResidual, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]NativeRaimNormalizedResidual, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := sizeTToInt(required, "FDE normalized residual count")
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, "FDE normalized residual first-call count"); err != nil {
		return nil, err
	}
	buffer := make([]C.SidereonRaimNormalizedResidual, count)
	length, err := cSize(len(buffer), "FDE normalized residual output length")
	if err != nil {
		return nil, err
	}
	var output *C.SidereonRaimNormalizedResidual
	if len(buffer) != 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, length, &written, &required))); err != nil {
		return nil, err
	}
	actual, err := validateTwoPassCounts("FDE normalized residuals", len(buffer), count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	rows := make([]NativeRaimNormalizedResidual, actual)
	for i := range rows {
		rows[i] = NativeRaimNormalizedResidual{tokenFromC(buffer[i].sat_id), float64(buffer[i].normalized_residual)}
	}
	return rows, nil
}

// fdeStatusErrorLocked captures quality or unresolved data while still on the
// OS thread that ran the FDE producer.
func fdeStatusErrorLocked(status uint32) error {
	err := qualityStatusErrorLocked(status)
	if err == nil {
		return nil
	}
	if _, ok := err.(*StatusError); !ok {
		return err
	}
	statusErr := err.(*StatusError)
	if statusErr.QualityKind != 0 || statusErr.Code != 5 {
		return err
	}
	unresolved := &NativeFDEUnresolvedError{Cause: err}
	setCaptureError := func(captureErr error) {
		if captureErr == nil {
			return
		}
		if unresolved.CaptureError == nil {
			unresolved.CaptureError = captureErr
		} else {
			unresolved.CaptureError = errors.Join(unresolved.CaptureError, captureErr)
		}
	}
	var info C.SidereonFdeUnresolvedInfo
	infoErr := statusErrorLocked(uint32(C.sidereon_last_fde_unresolved(&info)))
	if infoErr != nil {
		setCaptureError(infoErr)
		return unresolved
	}
	unresolved.Reason = uint32(info.reason)
	if unresolved.Reason == 0 {
		return err
	}
	unresolved.RAIM = nativeRaimResult(info.raim)
	unresolved.HasRAIM = true
	var solution *C.SidereonSppSolution
	solutionErr := statusErrorLocked(uint32(C.sidereon_last_fde_unresolved_solution(&solution)))
	if solutionErr != nil {
		setCaptureError(solutionErr)
	} else if solution == nil {
		setCaptureError(errors.New("sidereon: unresolved FDE solution getter returned nil"))
	} else {
		defer C.sidereon_spp_solution_free(solution)
		nativeSolution, readErr := readSPPSolutionLocked(solution)
		if readErr != nil {
			setCaptureError(readErr)
		} else {
			unresolved.Solution, unresolved.HasSolution = nativeSolution, true
		}
	}
	excluded, readErr := copyNativeTokensLocked("FDE unresolved excluded satellites", func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_last_fde_unresolved_excluded_sats(out, length, written, required)
	})
	if readErr != nil {
		setCaptureError(readErr)
	} else {
		unresolved.Excluded, unresolved.HasExcluded = excluded, true
	}
	rows, readErr := copyFDERaimRowsLocked(func(out *C.SidereonRaimNormalizedResidual, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_last_fde_unresolved_raim_normalized_residuals(out, length, written, required)
	})
	if readErr != nil {
		setCaptureError(readErr)
	} else {
		unresolved.Normalized, unresolved.HasNormalized = rows, true
	}
	return unresolved
}

type NativeSPPRobustConfig struct {
	HuberK, ScaleFloorM float64
	MaxOuter            uint64
	OuterToleranceM     float64
}

func ptrUint64(value uint64) *uint64    { return &value }
func ptrFloat64(value float64) *float64 { return &value }

func FDEOptionsDefault() (NativeFDEOptions, error) {
	var options C.SidereonFdeOptions
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_fde_options_init(&options))
	}); err != nil {
		return NativeFDEOptions{}, err
	}
	return NativeFDEOptions{
		PFA:              float64(options.p_fa),
		MaxExclusions:    ptrUint64(uint64(options.max_exclusions)),
		MaxExclusionRMSM: ptrFloat64(float64(options.max_exclusion_rms_m)),
		WeightsMode:      uint32(options.weights_mode),
		SystemsEnabled:   bool(options.n_systems_enabled),
		Systems:          int64(options.n_systems),
		UseValidation:    bool(options.use_validation_options),
	}, nil
}

func newFDESolution(pointer *C.SidereonFdeSolution) (*FDESolution, error) {
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) {
		C.sidereon_fde_solution_free((*C.SidereonFdeSolution)(p))
	})
	if err != nil {
		return nil, err
	}
	return &FDESolution{handle: h}, nil
}
func (b *BroadcastEphemeris) NavMessagePreference() (uint32, error) {
	var out C.uint32_t
	err := b.resource.with(func(p unsafe.Pointer) error {
		var e error
		withCThread(func() {
			e = statusErrorLocked(uint32(C.sidereon_broadcast_ephemeris_nav_message_preference((*C.SidereonBroadcastEphemeris)(p), &out)))
		})
		return e
	})
	return uint32(out), err
}

func nativeFDEOptions(value NativeFDEOptions) (C.SidereonFdeOptions, unsafe.Pointer, []*C.char, error) {
	var options C.SidereonFdeOptions
	var err error
	withCThread(func() { err = statusErrorLocked(uint32(C.sidereon_fde_options_init(&options))) })
	if err != nil {
		return options, nil, nil, err
	}
	if value.MaxExclusions != nil {
		maxExclusions, conversionErr := cSize64(*value.MaxExclusions, "FDE exclusion budget")
		if conversionErr != nil {
			return options, nil, nil, conversionErr
		}
		options.max_exclusions = maxExclusions
	}
	if value.MaxExclusionRMSM != nil {
		options.max_exclusion_rms_m = C.double(*value.MaxExclusionRMSM)
	}
	if value.PFA != 0 {
		options.p_fa = C.double(value.PFA)
	}
	options.weights_mode = C.uint32_t(value.WeightsMode)
	options.n_systems_enabled = C.bool(value.SystemsEnabled)
	options.n_systems = C.int64_t(value.Systems)
	if value.HasUseValidation {
		options.use_validation_options = C.bool(value.UseValidation)
	}
	if len(value.Weights) == 0 {
		return options, nil, nil, nil
	}
	for _, weight := range value.Weights {
		if err := rejectEmbeddedNUL(weight.SatelliteID, "FDE satellite ID"); err != nil {
			return options, nil, nil, err
		}
	}
	weightCount, err := cSize(len(value.Weights), "FDE weight count")
	if err != nil {
		return options, nil, nil, err
	}
	size, err := checkedNativeAllocationSize(len(value.Weights), unsafe.Sizeof(C.SidereonFdeRaimWeight{}))
	if err != nil {
		return options, nil, nil, err
	}
	memory := C.malloc(C.size_t(size))
	if memory == nil {
		return options, nil, nil, errors.New("sidereon: unable to allocate FDE weights")
	}
	entries := unsafe.Slice((*C.SidereonFdeRaimWeight)(memory), len(value.Weights))
	ids := make([]*C.char, len(value.Weights))
	for i, w := range value.Weights {
		ids[i] = C.CString(w.SatelliteID)
		if ids[i] == nil {
			freeCStrings(ids)
			C.free(memory)
			return options, nil, nil, errors.New("sidereon: unable to allocate FDE satellite ID")
		}
		entries[i].sat_id = ids[i]
		entries[i].weight = C.double(w.Weight)
	}
	options.weights = (*C.SidereonFdeRaimWeight)(memory)
	options.weight_count = weightCount
	return options, memory, ids, nil
}
func fdeInputs(value SPPConfig) (C.SidereonSppInputs, unsafe.Pointer, []*C.char, error) {
	observationCount, err := cSize(len(value.Observations), "SPP observation count")
	if err != nil {
		return C.SidereonSppInputs{}, nil, nil, err
	}
	var inputs C.SidereonSppInputs
	size, err := checkedNativeAllocationSize(len(value.Observations), unsafe.Sizeof(C.SidereonObservation{}))
	if err != nil {
		return inputs, nil, nil, err
	}
	var memory unsafe.Pointer
	if len(value.Observations) > 0 {
		memory = C.malloc(C.size_t(size))
		if memory == nil {
			return inputs, nil, nil, errors.New("sidereon: unable to allocate FDE observations")
		}
	}
	entries := unsafe.Slice((*C.SidereonObservation)(memory), len(value.Observations))
	ids := make([]*C.char, len(value.Observations))
	for i, o := range value.Observations {
		if err := rejectEmbeddedNUL(o.SatelliteID, "SPP satellite ID"); err != nil {
			freeCStrings(ids)
			C.free(memory)
			return inputs, nil, nil, err
		}
		ids[i] = C.CString(o.SatelliteID)
		if ids[i] == nil {
			freeCStrings(ids)
			C.free(memory)
			return inputs, nil, nil, errors.New("sidereon: unable to allocate SPP satellite ID")
		}
		entries[i].sat_id = ids[i]
		entries[i].pseudorange_m = C.double(o.PseudorangeM)
	}
	if len(entries) > 0 {
		inputs.observations = &entries[0]
	}
	inputs.observation_count = observationCount
	inputs.t_rx_j2000_s = C.double(value.TRxJ2000S)
	inputs.t_rx_second_of_day_s = C.double(value.TRxSecondOfDayS)
	inputs.day_of_year = C.double(value.DayOfYear)
	for i := range value.InitialGuess {
		inputs.initial_guess[i] = C.double(value.InitialGuess[i])
	}
	inputs.ionosphere = C.bool(value.Ionosphere)
	inputs.troposphere = C.bool(value.Troposphere)
	inputs.with_geodetic = C.bool(value.WithGeodetic)
	return inputs, memory, ids, nil
}
func solveFDE(source unsafe.Pointer, sp3 bool, config SPPConfig, options NativeFDEOptions) (*FDESolution, error) {
	inputs, memory, ids, err := fdeInputs(config)
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	defer freeCStrings(ids)
	coptions, wmemory, wids, err := nativeFDEOptions(options)
	if err != nil {
		return nil, err
	}
	defer C.free(wmemory)
	defer freeCStrings(wids)
	var out *C.SidereonFdeSolution
	var opErr error
	withCThread(func() {
		if sp3 {
			opErr = fdeStatusErrorLocked(uint32(C.sidereon_fde_solve_spp((*C.SidereonSp3)(source), &inputs, &coptions, &out)))
		} else {
			opErr = fdeStatusErrorLocked(uint32(C.sidereon_fde_solve_broadcast((*C.SidereonBroadcastEphemeris)(source), &inputs, &coptions, &out)))
		}
	})
	if opErr != nil {
		return nil, opErr
	}
	return newFDESolution(out)
}

func solveRobustFDE(source unsafe.Pointer, sp3 bool, config SPPConfig, robust NativeSPPRobustConfig, options NativeFDEOptions) (*FDESolution, error) {
	inputs, memory, ids, err := fdeInputs(config)
	if err != nil {
		return nil, err
	}
	defer C.free(memory)
	defer freeCStrings(ids)
	coptions, weightMemory, weightIDs, err := nativeFDEOptions(options)
	if err != nil {
		return nil, err
	}
	defer C.free(weightMemory)
	defer freeCStrings(weightIDs)
	maxOuter, err := cSize64(robust.MaxOuter, "robust outer-iteration count")
	if err != nil {
		return nil, err
	}
	cr := C.SidereonSppRobustConfig{
		huber_k: C.double(robust.HuberK), scale_floor_m: C.double(robust.ScaleFloorM),
		max_outer: maxOuter, outer_tol_m: C.double(robust.OuterToleranceM),
	}
	var out *C.SidereonFdeSolution
	var callErr error
	withCThread(func() {
		if sp3 {
			callErr = fdeStatusErrorLocked(uint32(C.sidereon_robust_fde_solve_spp((*C.SidereonSp3)(source), &inputs, &cr, &coptions, &out)))
		} else {
			callErr = fdeStatusErrorLocked(uint32(C.sidereon_robust_fde_solve_broadcast((*C.SidereonBroadcastEphemeris)(source), &inputs, &cr, &coptions, &out)))
		}
	})
	if callErr != nil {
		return nil, callErr
	}
	return newFDESolution(out)
}
func SolveFDESPP(sp3 *SP3, config SPPConfig, options NativeFDEOptions) (*FDESolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var out *FDESolution
	err := sp3.handle.with(func(p unsafe.Pointer) error {
		var inner error
		out, inner = solveFDE(p, true, config, options)
		return inner
	})
	return out, err
}
func SolveFDEBroadcast(b *BroadcastEphemeris, config SPPConfig, options NativeFDEOptions) (*FDESolution, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	var out *FDESolution
	err := b.resource.with(func(p unsafe.Pointer) error {
		var inner error
		out, inner = solveFDE(p, false, config, options)
		return inner
	})
	return out, err
}

func SolveRobustFDESPP(sp3 *SP3, config SPPConfig, robust NativeSPPRobustConfig, options NativeFDEOptions) (*FDESolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var out *FDESolution
	err := sp3.handle.with(func(pointer unsafe.Pointer) error {
		var inner error
		out, inner = solveRobustFDE(pointer, true, config, robust, options)
		return inner
	})
	return out, err
}

func SolveRobustFDEBroadcast(b *BroadcastEphemeris, config SPPConfig, robust NativeSPPRobustConfig, options NativeFDEOptions) (*FDESolution, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	var out *FDESolution
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		var inner error
		out, inner = solveRobustFDE(pointer, false, config, robust, options)
		return inner
	})
	return out, err
}
func (f *FDESolution) Close() error {
	if f == nil {
		return nil
	}
	return f.handle.close()
}
func (f *FDESolution) Output() (NativeFDEOutput, error) {
	var iterations C.size_t
	var err error
	err = f.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			err = statusErrorLocked(uint32(C.sidereon_fde_solution_iterations((*C.SidereonFdeSolution)(p), &iterations)))
		})
		return err
	})
	if err != nil {
		return NativeFDEOutput{}, err
	}
	ids, err := copySurfaceTokens(f.handle, "FDE excluded satellites", func(p unsafe.Pointer, o *C.SidereonSatelliteToken, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_fde_solution_excluded_sats((*C.SidereonFdeSolution)(p), o, l, w, r)
	})
	return NativeFDEOutput{uint64(iterations), ids}, err
}

func (f *FDESolution) Solution() (SPPSolution, error) {
	var solution *C.SidereonSppSolution
	var result SPPSolution
	err := f.handle.read(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_fde_solution_solution((*C.SidereonFdeSolution)(pointer), &solution)))
			if callErr == nil {
				if solution == nil {
					callErr = errors.New("sidereon: FDE solution accessor returned nil")
					return
				}
				defer C.sidereon_spp_solution_free(solution)
				result, callErr = readSPPSolutionLocked(solution)
			}
		})
		return callErr
	})
	return result, err
}

func (f *FDESolution) RecomputeRAIM(pfa float64, weightsMode uint32, weights []NativeFDERaimWeight, systemsEnabled bool, systems int64) (NativeRaimResult, error) {
	weightMemory, weightCount, weightIDs, err := makeFDERaimWeights(weights)
	if err != nil {
		return NativeRaimResult{}, err
	}
	defer C.free(weightMemory)
	defer freeCStrings(weightIDs)
	var weightPointer *C.SidereonFdeRaimWeight
	if weightMemory != nil {
		weightPointer = (*C.SidereonFdeRaimWeight)(weightMemory)
	}
	var result C.SidereonRaimResult
	err = f.handle.read(func(pointer unsafe.Pointer) error {
		var solution *C.SidereonSppSolution
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_fde_solution_solution((*C.SidereonFdeSolution)(pointer), &solution)))
			if callErr != nil {
				return
			}
			if solution == nil {
				callErr = errors.New("sidereon: FDE solution accessor returned nil")
				return
			}
			defer C.sidereon_spp_solution_free(solution)
			callErr = qualityStatusErrorLocked(uint32(C.sidereon_raim_for_solution(solution, C.double(pfa), C.uint32_t(weightsMode), weightPointer, weightCount, C.bool(systemsEnabled), C.int64_t(systems), &result)))
		})
		return callErr
	})
	if err != nil {
		return NativeRaimResult{}, err
	}
	return nativeRaimResult(result), nil
}

func (f *FDESolution) AcceptedRAIM() (NativeRaimResult, []NativeRaimNormalizedResidual, error) {
	var result C.SidereonRaimResult
	var operationErr error
	err := f.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = statusErrorLocked(uint32(C.sidereon_fde_solution_raim((*C.SidereonFdeSolution)(pointer), &result)))
		})
		return operationErr
	})
	if err != nil {
		return NativeRaimResult{}, nil, err
	}
	rows, err := copySurfaceFDERaimRows(f.handle, "accepted FDE normalized residuals", func(pointer unsafe.Pointer, out *C.SidereonRaimNormalizedResidual, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_fde_solution_raim_normalized_residuals((*C.SidereonFdeSolution)(pointer), out, length, written, required)
	})
	return nativeRaimResult(result), rows, err
}

func copySurfaceFDERaimRows(handle *surfaceHandle, label string, call func(unsafe.Pointer, *C.SidereonRaimNormalizedResidual, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]NativeRaimNormalizedResidual, error) {
	var rows []NativeRaimNormalizedResidual
	var opErr error
	err := handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			rows, opErr = copyFDERaimRowsLocked(func(out *C.SidereonRaimNormalizedResidual, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return call(pointer, out, length, written, required)
			})
		})
		return opErr
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}
