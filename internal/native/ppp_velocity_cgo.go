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

type PppVelocityRobustConfig struct {
	HuberK       float64
	ScaleFloorMS float64
	MaxOuter     int
	OuterTolMS   float64
}

type PppVelocityConfig struct {
	MinimumObservations int
	Robust              *PppVelocityRobustConfig
}

type PppReceiverVelocityState struct {
	PositionM    [3]float64
	VelocityMS   [3]float64
	ClockDriftMS float64
}

type PppVelocityObservation struct {
	SatelliteID           string
	SatellitePositionM    [3]float64
	SatelliteVelocityMS   [3]float64
	MeasuredRangeRateMS   float64
	SigmaMS               float64
	SatelliteClockDriftMS float64
}

type PppRangeRatePrediction struct {
	LOSUnit     [3]float64
	RangeRateMS float64
}

type PppVelocitySolutionValues struct {
	VelocityMS       [3]float64
	ClockDriftMS     float64
	ResidualRMSMS    float64
	RobustIterations int
	HasRobustScale   bool
	RobustScaleMS    float64
}

type PppVelocitySolution struct {
	_      noCopy
	handle *positioningHandle
}

func cPppVelocityRobustConfig(value PppVelocityRobustConfig) (C.SidereonPppVelocityRobustConfig, error) {
	maxOuter, err := checkedNativeSize(value.MaxOuter)
	if err != nil {
		return C.SidereonPppVelocityRobustConfig{}, err
	}
	return C.SidereonPppVelocityRobustConfig{huber_k: C.double(value.HuberK), scale_floor_m_s: C.double(value.ScaleFloorMS), max_outer: maxOuter, outer_tol_m_s: C.double(value.OuterTolMS)}, nil
}

func cPppVelocityConfig(value PppVelocityConfig) (C.SidereonPppVelocityConfig, error) {
	minimum, err := checkedNativeSize(value.MinimumObservations)
	if err != nil {
		return C.SidereonPppVelocityConfig{}, err
	}
	out := C.SidereonPppVelocityConfig{minimum_observations: minimum, has_robust: C.bool(value.Robust != nil)}
	if value.Robust != nil {
		out.robust, err = cPppVelocityRobustConfig(*value.Robust)
	}
	return out, err
}

func cPppReceiverVelocityState(value PppReceiverVelocityState) C.SidereonPppReceiverVelocityState {
	var out C.SidereonPppReceiverVelocityState
	for axis := range value.PositionM {
		out.position_m[axis] = C.double(value.PositionM[axis])
		out.velocity_m_s[axis] = C.double(value.VelocityMS[axis])
	}
	out.clock_drift_m_s = C.double(value.ClockDriftMS)
	return out
}

func cPppVelocityObservation(value PppVelocityObservation, arena *pppArena) (C.SidereonPppVelocityObservation, error) {
	id, err := arena.cstring(value.SatelliteID, "PPP velocity satellite ID")
	if err != nil {
		return C.SidereonPppVelocityObservation{}, err
	}
	out := C.SidereonPppVelocityObservation{sat_id: id, measured_range_rate_m_s: C.double(value.MeasuredRangeRateMS), sigma_m_s: C.double(value.SigmaMS), satellite_clock_drift_m_s: C.double(value.SatelliteClockDriftMS)}
	for axis := range value.SatellitePositionM {
		out.satellite_position_m[axis] = C.double(value.SatellitePositionM[axis])
		out.satellite_velocity_m_s[axis] = C.double(value.SatelliteVelocityMS[axis])
	}
	return out, nil
}

func PppVelocityRobustConfigInit() (PppVelocityRobustConfig, error) {
	var value C.SidereonPppVelocityRobustConfig
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_velocity_robust_config_init(&value)) })
	maxOuter, countErr := checkedNativeCount(uint64(value.max_outer))
	if err == nil {
		err = countErr
	}
	return PppVelocityRobustConfig{HuberK: float64(value.huber_k), ScaleFloorMS: float64(value.scale_floor_m_s), MaxOuter: maxOuter, OuterTolMS: float64(value.outer_tol_m_s)}, err
}

func PppVelocityConfigInit() (PppVelocityConfig, error) {
	var value C.SidereonPppVelocityConfig
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_velocity_config_init(&value)) })
	minimum, countErr := checkedNativeCount(uint64(value.minimum_observations))
	if err == nil {
		err = countErr
	}
	result := PppVelocityConfig{MinimumObservations: minimum}
	if bool(value.has_robust) {
		maxOuter, robustErr := checkedNativeCount(uint64(value.robust.max_outer))
		if err == nil {
			err = robustErr
		}
		result.Robust = &PppVelocityRobustConfig{HuberK: float64(value.robust.huber_k), ScaleFloorMS: float64(value.robust.scale_floor_m_s), MaxOuter: maxOuter, OuterTolMS: float64(value.robust.outer_tol_m_s)}
	}
	return result, err
}

func PppReceiverVelocityStateStationaryAt(position [3]float64) (PppReceiverVelocityState, error) {
	input := [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])}
	var out C.SidereonPppReceiverVelocityState
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_receiver_velocity_state_stationary_at(&input[0], &out)) })
	return PppReceiverVelocityState{PositionM: [3]float64{float64(out.position_m[0]), float64(out.position_m[1]), float64(out.position_m[2])}, VelocityMS: [3]float64{float64(out.velocity_m_s[0]), float64(out.velocity_m_s[1]), float64(out.velocity_m_s[2])}, ClockDriftMS: float64(out.clock_drift_m_s)}, err
}

func PppPredictRangeRateMS(observation PppVelocityObservation, receiver PppReceiverVelocityState) (PppRangeRatePrediction, error) {
	var result PppRangeRatePrediction
	err := withCThreadError(func() error {
		arena := &pppArena{}
		defer arena.close()
		input, conversionErr := cPppVelocityObservation(observation, arena)
		if conversionErr != nil {
			return conversionErr
		}
		state := cPppReceiverVelocityState(receiver)
		var out C.SidereonPppRangeRatePrediction
		if statusErr := statusErrorLocked(uint32(C.sidereon_ppp_predict_range_rate_m_s(&input, &state, &out))); statusErr != nil {
			return statusErr
		}
		result = PppRangeRatePrediction{LOSUnit: [3]float64{float64(out.los_unit[0]), float64(out.los_unit[1]), float64(out.los_unit[2])}, RangeRateMS: float64(out.range_rate_m_s)}
		return nil
	})
	return result, err
}

func releasePppVelocitySolution(pointer unsafe.Pointer) {
	withCThread(func() { C.sidereon_ppp_velocity_solution_free((*C.SidereonPppVelocitySolution)(pointer)) })
}

func SolvePppVelocityLeastSquares(observations []PppVelocityObservation, receiverPositionM [3]float64, config PppVelocityConfig) (*PppVelocitySolution, error) {
	var output *C.SidereonPppVelocitySolution
	err := withCThreadError(func() error {
		arena := &pppArena{}
		defer arena.close()
		count, countErr := checkedNativeSize(len(observations))
		if countErr != nil {
			return countErr
		}
		memory, allocationErr := arena.malloc(len(observations), unsafe.Sizeof(C.SidereonPppVelocityObservation{}), "PPP velocity observations")
		if allocationErr != nil {
			return allocationErr
		}
		values := unsafe.Slice((*C.SidereonPppVelocityObservation)(memory), len(observations))
		for index, observation := range observations {
			converted, conversionErr := cPppVelocityObservation(observation, arena)
			if conversionErr != nil {
				return conversionErr
			}
			values[index] = converted
		}
		cConfig, conversionErr := cPppVelocityConfig(config)
		if conversionErr != nil {
			return conversionErr
		}
		position := [3]C.double{C.double(receiverPositionM[0]), C.double(receiverPositionM[1]), C.double(receiverPositionM[2])}
		return statusErrorLocked(uint32(C.sidereon_ppp_solve_velocity_least_squares((*C.SidereonPppVelocityObservation)(memory), count, &position[0], &cConfig, &output)))
	})
	if err != nil {
		if output != nil {
			releasePppVelocitySolution(unsafe.Pointer(output))
		}
		return nil, err
	}
	if output == nil {
		return nil, missingNativeHandle("PPP velocity solution")
	}
	return &PppVelocitySolution{handle: newPositioningHandle(unsafe.Pointer(output), releasePppVelocitySolution)}, nil
}

func (solution *PppVelocitySolution) Close() error {
	if solution == nil || solution.handle == nil {
		return nil
	}
	return solution.handle.close()
}

func (solution *PppVelocitySolution) Values() (result PppVelocitySolutionValues, err error) {
	if solution == nil || solution.handle == nil {
		return result, ErrClosed
	}
	err = solution.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var velocity [3]C.double
			var clock, rms, scale C.double
			var iterations C.size_t
			var hasScale C.bool
			if statusErr := statusErrorLocked(uint32(C.sidereon_ppp_velocity_solution_values((*C.SidereonPppVelocitySolution)(pointer), &velocity[0], &clock, &rms, &iterations, &hasScale, &scale))); statusErr != nil {
				return statusErr
			}
			count, countErr := checkedNativeCount(uint64(iterations))
			if countErr != nil {
				return countErr
			}
			result = PppVelocitySolutionValues{VelocityMS: [3]float64{float64(velocity[0]), float64(velocity[1]), float64(velocity[2])}, ClockDriftMS: float64(clock), ResidualRMSMS: float64(rms), RobustIterations: count, HasRobustScale: bool(hasScale), RobustScaleMS: float64(scale)}
			return nil
		})
	})
	runtime.KeepAlive(solution)
	return result, err
}

func pppVelocitySatelliteIDs(solution *PppVelocitySolution, downweighted bool) (result []string, err error) {
	if solution == nil || solution.handle == nil {
		return nil, ErrClosed
	}
	err = solution.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(out *C.SidereonSatelliteToken, count C.size_t) C.enum_SidereonStatus {
				if downweighted {
					return C.sidereon_ppp_velocity_solution_downweighted_sats((*C.SidereonPppVelocitySolution)(pointer), out, count, &written, &required)
				}
				return C.sidereon_ppp_velocity_solution_used_sats((*C.SidereonPppVelocitySolution)(pointer), out, count, &written, &required)
			}
			if statusErr := statusErrorLocked(uint32(call(nil, 0))); statusErr != nil {
				return statusErr
			}
			count, countErr := validateNativeQuery("PPP velocity satellite IDs", uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			values := make([]C.SidereonSatelliteToken, count)
			var output *C.SidereonSatelliteToken
			if count != 0 {
				output = &values[0]
			}
			written, required = 0, 0
			if statusErr := statusErrorLocked(uint32(call(output, C.size_t(count)))); statusErr != nil {
				return statusErr
			}
			actual, countErr := validateTwoPassCounts("PPP velocity satellite IDs", count, count, uint64(written), uint64(required))
			if countErr != nil {
				return countErr
			}
			result = make([]string, actual)
			for index := range result {
				result[index] = tokenFromC(values[index])
			}
			return nil
		})
	})
	runtime.KeepAlive(solution)
	return result, err
}

func (solution *PppVelocitySolution) UsedSatelliteIDs() ([]string, error) {
	return pppVelocitySatelliteIDs(solution, false)
}

func (solution *PppVelocitySolution) DownweightedSatelliteIDs() ([]string, error) {
	return pppVelocitySatelliteIDs(solution, true)
}

func (solution *PppVelocitySolution) Equal(other *PppVelocitySolution) (equal bool, err error) {
	if solution == nil || solution.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	err = withPositioningHandlePair(solution.handle, other.handle, func(left, right unsafe.Pointer) error {
		return withCThreadError(func() error {
			var value C.bool
			if statusErr := statusErrorLocked(uint32(C.sidereon_ppp_velocity_solution_equal((*C.SidereonPppVelocitySolution)(left), (*C.SidereonPppVelocitySolution)(right), &value))); statusErr != nil {
				return statusErr
			}
			equal = bool(value)
			return nil
		})
	})
	runtime.KeepAlive(solution)
	runtime.KeepAlive(other)
	return equal, err
}
