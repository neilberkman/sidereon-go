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

type NativeFusionVelocityMatchState struct {
	Epoch              float64
	Position, Velocity [3]float64
}

type NativeFusionVelocityMatchingConfig struct {
	MaxOutageDuration float64
}

type NativeFusionVelocityMatchedTrajectory struct {
	StateCount                                             uint64
	EndpointPositionCorrection, EndpointVelocityCorrection [3]float64
}

func cVelocityMatchState(v NativeFusionVelocityMatchState) C.SidereonFusionVelocityMatchState {
	return C.SidereonFusionVelocityMatchState{
		t_j2000_s: C.double(v.Epoch), position_ecef_m: cFusionArray3(v.Position), velocity_ecef_mps: cFusionArray3(v.Velocity),
	}
}

func velocityMatchState(v C.SidereonFusionVelocityMatchState) NativeFusionVelocityMatchState {
	var out NativeFusionVelocityMatchState
	out.Epoch = float64(v.t_j2000_s)
	for i := 0; i < 3; i++ {
		out.Position[i] = float64(v.position_ecef_m[i])
		out.Velocity[i] = float64(v.velocity_ecef_mps[i])
	}
	return out
}

func velocityMatchedTrajectory(v C.SidereonFusionVelocityMatchedTrajectory) NativeFusionVelocityMatchedTrajectory {
	var out NativeFusionVelocityMatchedTrajectory
	out.StateCount = uint64(v.state_count)
	for i := 0; i < 3; i++ {
		out.EndpointPositionCorrection[i] = float64(v.endpoint_position_correction_ecef_m[i])
		out.EndpointVelocityCorrection[i] = float64(v.endpoint_velocity_correction_ecef_mps[i])
	}
	return out
}

func FusionVelocityMatchOutage(states []NativeFusionVelocityMatchState, first NativeFusionLooseMeasurement, config NativeFusionVelocityMatchingConfig) ([]NativeFusionVelocityMatchState, NativeFusionVelocityMatchedTrajectory, error) {
	stateCount, err := cSize(len(states), "velocity-match state count")
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	stateSize, err := checkedNativeAllocationSize(len(states), unsafe.Sizeof(C.SidereonFusionVelocityMatchState{}))
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	var stateMem unsafe.Pointer
	if len(states) > 0 {
		stateMem = C.malloc(C.size_t(stateSize))
		if stateMem == nil {
			return nil, NativeFusionVelocityMatchedTrajectory{}, errors.New("sidereon: unable to allocate velocity-match states")
		}
		rows := unsafe.Slice((*C.SidereonFusionVelocityMatchState)(stateMem), len(states))
		for i, state := range states {
			rows[i] = cVelocityMatchState(state)
		}
	}
	defer C.free(stateMem)

	cov, covarianceLength, err := cFloats(first.Covariance, "velocity-match covariance")
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	defer C.free(cov)
	satellitesUsed, err := cSize64(first.SatellitesUsed, "velocity-match satellite count")
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	fix := cFusionLoose(first, cov, covarianceLength, satellitesUsed)
	cfg := C.SidereonFusionVelocityMatchingConfig{max_outage_duration_s: C.double(config.MaxOutageDuration)}

	var required C.size_t
	var written C.size_t
	var trajectory C.SidereonFusionVelocityMatchedTrajectory
	var callErr error
	withCThread(func() {
		callErr = statusErrorLocked(uint32(C.sidereon_fusion_velocity_match_outage(
			(*C.SidereonFusionVelocityMatchState)(stateMem), stateCount, &fix, &cfg,
			nil, 0, &written, &required, &trajectory)))
	})
	if callErr != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, callErr
	}
	if _, err := validateTwoPassCounts("velocity-match states query", 0, 0, uint64(written), uint64(required)); err != nil {
		// A query is allowed to report the required output while writing zero rows.
		if required != 0 || written != 0 {
			if _, countErr := checkedNativeCount(uint64(required)); countErr != nil {
				return nil, NativeFusionVelocityMatchedTrajectory{}, countErr
			}
			if written != 0 {
				return nil, NativeFusionVelocityMatchedTrajectory{}, errors.New("sidereon: velocity-match query wrote output")
			}
		}
	}
	n, err := sizeTToInt(required, "velocity-match state count")
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	outMem, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonFusionVelocityMatchState{}))
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	var outPtr unsafe.Pointer
	outputLength, err := cSize(n, "velocity-match output capacity")
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	if n > 0 {
		outPtr = C.malloc(C.size_t(outMem))
		if outPtr == nil {
			return nil, NativeFusionVelocityMatchedTrajectory{}, errors.New("sidereon: unable to allocate velocity-match output")
		}
	}
	defer C.free(outPtr)
	written, required = 0, 0
	callErr = nil
	withCThread(func() {
		callErr = statusErrorLocked(uint32(C.sidereon_fusion_velocity_match_outage(
			(*C.SidereonFusionVelocityMatchState)(stateMem), stateCount, &fix, &cfg,
			(*C.SidereonFusionVelocityMatchState)(outPtr), outputLength, &written, &required, &trajectory)))
	})
	if callErr != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, callErr
	}
	actual, err := validateTwoPassCounts("velocity-match states", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, NativeFusionVelocityMatchedTrajectory{}, err
	}
	result := make([]NativeFusionVelocityMatchState, actual)
	if actual > 0 {
		rows := unsafe.Slice((*C.SidereonFusionVelocityMatchState)(outPtr), actual)
		for i := range result {
			result[i] = velocityMatchState(rows[i])
		}
	}
	return result, velocityMatchedTrajectory(trajectory), nil
}

type SmoothedFusionTrajectory struct {
	_      noCopy
	handle *surfaceHandle
}

func newSmoothedFusionTrajectory(p *C.SidereonSmoothedFusionTrajectory) (*SmoothedFusionTrajectory, error) {
	h, err := newSurfaceHandle(unsafe.Pointer(p), func(p unsafe.Pointer) {
		C.sidereon_smoothed_fusion_trajectory_free((*C.SidereonSmoothedFusionTrajectory)(p))
	})
	if err != nil {
		return nil, err
	}
	return &SmoothedFusionTrajectory{handle: h}, nil
}

func (h *FusionRTSHistory) Smooth() (*SmoothedFusionTrajectory, error) {
	var out *C.SidereonSmoothedFusionTrajectory
	err := h.handle.read(func(p unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_smooth_fusion_rts((*C.SidereonFusionRtsHistory)(p), &out)))
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return newSmoothedFusionTrajectory(out)
}

func (h *SmoothedFusionTrajectory) Close() error {
	if h == nil {
		return nil
	}
	return h.handle.close()
}

func (h *SmoothedFusionTrajectory) EpochCount() (int, error) {
	var count C.size_t
	err := h.handle.read(func(p unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_smoothed_fusion_trajectory_epoch_count((*C.SidereonSmoothedFusionTrajectory)(p), &count)))
		})
		return callErr
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "smoothed fusion epoch count")
}

func (h *SmoothedFusionTrajectory) Epoch(index int) (NativeFusionRTSEpoch, error) {
	if index < 0 {
		return NativeFusionRTSEpoch{}, errors.New("sidereon: negative smoothed fusion index")
	}
	var out C.SidereonSmoothedFusionEpoch
	err := h.handle.read(func(p unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_smoothed_fusion_trajectory_epoch((*C.SidereonSmoothedFusionTrajectory)(p), C.size_t(index), &out)))
		})
		return callErr
	})
	return NativeFusionRTSEpoch{float64(out.t_j2000_s), uint64(out.covariance_dimension), uint64(out.correction_len), bool(out.has_rts_gain_to_next)}, err
}

func (h *SmoothedFusionTrajectory) Values(index int, kind uint32) ([]float64, error) {
	if index < 0 {
		return nil, errors.New("sidereon: negative smoothed fusion index")
	}
	return copyTrackDoubles(h.handle, "smoothed fusion values", func(p unsafe.Pointer, out *C.double, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		switch kind {
		case 0:
			return C.sidereon_smoothed_fusion_trajectory_epoch_covariance((*C.SidereonSmoothedFusionTrajectory)(p), C.size_t(index), out, length, written, required)
		case 1:
			return C.sidereon_smoothed_fusion_trajectory_epoch_error_state_correction((*C.SidereonSmoothedFusionTrajectory)(p), C.size_t(index), out, length, written, required)
		case 2:
			return C.sidereon_smoothed_fusion_trajectory_epoch_position_ecef_m((*C.SidereonSmoothedFusionTrajectory)(p), C.size_t(index), out, length, written, required)
		default:
			return C.sidereon_smoothed_fusion_trajectory_epoch_rts_gain_to_next((*C.SidereonSmoothedFusionTrajectory)(p), C.size_t(index), out, length, written, required)
		}
	})
}
