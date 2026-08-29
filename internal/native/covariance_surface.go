//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type NativePropagationConfig struct {
	Epoch                                         float64
	Position, Velocity                            [3]float64
	ForceModel, Integrator                        uint32
	AbsTol, RelTol, InitialStep, MinStep, MaxStep float64
	MaxSteps                                      uint32
	MuEnabled                                     bool
	Mu                                            float64
	HasDrag                                       bool
	Drag                                          DragParameters
	ForceComponents                               ForceModelComponents
}
type NativeCovarianceNode struct {
	State      NativeCartesianState
	Covariance [6][6]float64
	Frame      uint32
}
type CovarianceEphemeris struct {
	_      noCopy
	handle *surfaceHandle
}

func propagationConfig(v NativePropagationConfig) C.SidereonStatePropagationConfig {
	var out C.SidereonStatePropagationConfig
	out.epoch_s = C.double(v.Epoch)
	for i := 0; i < 3; i++ {
		out.position_km[i] = C.double(v.Position[i])
		out.velocity_km_s[i] = C.double(v.Velocity[i])
	}
	out.force_model = C.uint32_t(v.ForceModel)
	out.integrator = C.uint32_t(v.Integrator)
	out.abs_tol = C.double(v.AbsTol)
	out.rel_tol = C.double(v.RelTol)
	out.initial_step_s = C.double(v.InitialStep)
	out.min_step_s = C.double(v.MinStep)
	out.max_step_s = C.double(v.MaxStep)
	out.max_steps = C.uint32_t(v.MaxSteps)
	out.mu_km3_s2_enabled = C.bool(v.MuEnabled)
	out.mu_km3_s2 = C.double(v.Mu)
	out.has_drag = C.bool(v.HasDrag)
	out.drag = cDragParameters(v.Drag)
	out.force_components = cForceModelComponents(v.ForceComponents)
	return out
}
func PropagationConfigDefault() (NativePropagationConfig, error) {
	var c C.SidereonStatePropagationConfig
	err := callStatus(func() uint32 { return uint32(C.sidereon_state_propagation_config_init(&c)) })
	var p, v [3]float64
	for i := 0; i < 3; i++ {
		p[i] = float64(c.position_km[i])
		v[i] = float64(c.velocity_km_s[i])
	}
	return NativePropagationConfig{Epoch: float64(c.epoch_s), Position: p, Velocity: v, ForceModel: uint32(c.force_model), Integrator: uint32(c.integrator), AbsTol: float64(c.abs_tol), RelTol: float64(c.rel_tol), InitialStep: float64(c.initial_step_s), MinStep: float64(c.min_step_s), MaxStep: float64(c.max_step_s), MaxSteps: uint32(c.max_steps), MuEnabled: bool(c.mu_km3_s2_enabled), Mu: float64(c.mu_km3_s2), HasDrag: bool(c.has_drag), Drag: dragParametersFromC(c.drag), ForceComponents: forceModelComponentsFromC(c.force_components)}, err
}
func PropagateCovariance(config NativePropagationConfig, covariance [6][6]float64, epochs []float64, inputFrame, outputFrame uint32, noise NativeProcessNoise) (*CovarianceEphemeris, error) {
	if len(epochs) == 0 {
		return nil, errors.New("sidereon: covariance epoch list must not be empty")
	}
	p, epochLength, err := cFloats(epochs, "covariance epochs")
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	c := propagationConfig(config)
	co := C.SidereonCovariancePropagationOptions{input_frame: C.uint32_t(inputFrame), output_frame: C.uint32_t(outputFrame), process_noise: C.SidereonProcessNoise{kind: C.uint32_t(noise.Kind), q_radial_km2_s3: C.double(noise.RadialKm2S3), q_transverse_km2_s3: C.double(noise.TransverseKm2S3), q_normal_km2_s3: C.double(noise.NormalKm2S3)}}
	base := covarianceToC(covariance)
	var out *C.SidereonCovarianceEphemeris
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_propagate_covariance(&c, &base, (*C.double)(p), epochLength, co, &out))
	})
	if err != nil {
		return nil, err
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(x unsafe.Pointer) { C.sidereon_covariance_ephemeris_free((*C.SidereonCovarianceEphemeris)(x)) })
	if e != nil {
		return nil, e
	}
	return &CovarianceEphemeris{handle: h}, nil
}
func (e *CovarianceEphemeris) Close() error {
	if e == nil {
		return nil
	}
	return e.handle.close()
}
func (e *CovarianceEphemeris) Count() (int, error) {
	var o C.size_t
	err := fusionCallRead(e.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_covariance_ephemeris_count((*C.SidereonCovarianceEphemeris)(p), &o)
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(o, "covariance ephemeris count")
}
func (e *CovarianceEphemeris) CovarianceAt(epoch float64) ([6][6]float64, error) {
	var o C.SidereonCovarianceMatrix6
	err := fusionCallRead(e.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_covariance_ephemeris_covariance_at((*C.SidereonCovarianceEphemeris)(p), C.double(epoch), &o)
	})
	if err != nil {
		return [6][6]float64{}, err
	}
	return covarianceFromC(&o), nil
}
func (e *CovarianceEphemeris) Nodes() ([]NativeCovarianceNode, error) {
	var written, required C.size_t
	var opErr error
	err := e.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_covariance_ephemeris_nodes((*C.SidereonCovarianceEphemeris)(p), nil, 0, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	n, err := sizeTToInt(required, "covariance node count")
	if err != nil {
		return nil, err
	}
	if _, err = writtenToInt(written, 0, "covariance node first-call written count"); err != nil {
		return nil, err
	}
	buffer := make([]C.SidereonCovarianceNode, n)
	outputLength, err := cSize(n, "covariance node output capacity")
	if err != nil {
		return nil, err
	}
	var out *C.SidereonCovarianceNode
	if n > 0 {
		out = &buffer[0]
	}
	written, required = 0, 0
	err = e.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_covariance_ephemeris_nodes((*C.SidereonCovarianceEphemeris)(p), out, outputLength, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	actual, err := validateTwoPassCounts("covariance nodes", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]NativeCovarianceNode, actual)
	for i := range result {
		state := NativeCartesianState{EpochS: float64(buffer[i].state.epoch_s)}
		for k := 0; k < 3; k++ {
			state.PositionKm[k] = float64(buffer[i].state.position_km[k])
			state.VelocityKmS[k] = float64(buffer[i].state.velocity_km_s[k])
		}
		result[i] = NativeCovarianceNode{state, covarianceFromC(&buffer[i].covariance), uint32(buffer[i].frame)}
	}
	return result, nil
}
