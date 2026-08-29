//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type FDESolution struct {
	_      noCopy
	handle *surfaceHandle
}
type NativeFDEOptions struct {
	PFA            float64
	MaxIterations  uint64
	UnitWeights    bool
	Weights        []NativeFDERaimWeight
	SystemsEnabled bool
	Systems        int64
	UseValidation  bool
}
type NativeFDEOutput struct {
	Iterations uint64
	Excluded   []string
}

type NativeSPPRobustConfig struct {
	HuberK, ScaleFloorM float64
	MaxOuter            uint64
	OuterToleranceM     float64
}

func FDEOptionsDefault() (NativeFDEOptions, error) {
	var options C.SidereonFdeOptions
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_fde_options_init(&options))
	}); err != nil {
		return NativeFDEOptions{}, err
	}
	return NativeFDEOptions{
		PFA:            float64(options.p_fa),
		MaxIterations:  uint64(options.max_iterations),
		UnitWeights:    bool(options.unit_weights),
		SystemsEnabled: bool(options.n_systems_enabled),
		Systems:        int64(options.n_systems),
		UseValidation:  bool(options.use_validation_options),
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
	maxIterations, err := cSize64(value.MaxIterations, "FDE iteration count")
	if err != nil {
		return C.SidereonFdeOptions{}, nil, nil, err
	}
	var options C.SidereonFdeOptions
	withCThread(func() { err = statusErrorLocked(uint32(C.sidereon_fde_options_init(&options))) })
	if err != nil {
		return options, nil, nil, err
	}
	options.p_fa = C.double(value.PFA)
	options.max_iterations = maxIterations
	options.unit_weights = C.bool(value.UnitWeights)
	options.n_systems_enabled = C.bool(value.SystemsEnabled)
	options.n_systems = C.int64_t(value.Systems)
	options.use_validation_options = C.bool(value.UseValidation)
	if len(value.Weights) == 0 {
		return options, nil, nil, nil
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
		if err := rejectEmbeddedNUL(w.SatelliteID, "FDE satellite ID"); err != nil {
			C.free(memory)
			return options, nil, nil, err
		}
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
			opErr = statusErrorLocked(uint32(C.sidereon_fde_solve_spp((*C.SidereonSp3)(source), &inputs, &coptions, &out)))
		} else {
			opErr = statusErrorLocked(uint32(C.sidereon_fde_solve_broadcast((*C.SidereonBroadcastEphemeris)(source), &inputs, &coptions, &out)))
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
			callErr = statusErrorLocked(uint32(C.sidereon_robust_fde_solve_spp((*C.SidereonSp3)(source), &inputs, &cr, &coptions, &out)))
		} else {
			callErr = statusErrorLocked(uint32(C.sidereon_robust_fde_solve_broadcast((*C.SidereonBroadcastEphemeris)(source), &inputs, &cr, &coptions, &out)))
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

func (f *FDESolution) RAIM(pfa float64, unitWeights bool, weights []NativeFDERaimWeight, systemsEnabled bool, systems int64) (NativeRaimResult, error) {
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
			callErr = statusErrorLocked(uint32(C.sidereon_raim_for_solution(solution, C.double(pfa), C.bool(unitWeights), weightPointer, weightCount, C.bool(systemsEnabled), C.int64_t(systems), &result)))
		})
		return callErr
	})
	if err != nil {
		return NativeRaimResult{}, err
	}
	return NativeRaimResult{
		FaultDetected: bool(result.fault_detected), TestStatistic: float64(result.test_statistic),
		HasThreshold: bool(result.has_threshold), Threshold: float64(result.threshold),
		HasReducedChiSquare: bool(result.has_reduced_chi_square), ReducedChiSquare: float64(result.reduced_chi_square),
		RMSM: float64(result.rms_m), DOF: int64(result.dof), Testable: bool(result.testable),
		NormalizedResidualCount: uint64(result.normalized_residual_count), HasWorstSatellite: bool(result.has_worst_sat),
		WorstSatellite: tokenChars(result.worst_sat[:]),
	}, nil
}
