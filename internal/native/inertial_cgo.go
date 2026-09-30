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

type NativeInertialNavState struct {
	Epoch               float64
	Position, Velocity  [3]float64
	AttitudeMatrix      [9]float64
	AccelBias, GyroBias [3]float64
}

type NativeInertialIMUModel struct {
	AccelBias, GyroBias                           [3]float64
	AccelScaleMisalignment, GyroScaleMisalignment [9]float64
}

type NativeInertialIncrement struct {
	Epoch                     float64
	DeltaVelocity, DeltaTheta [3]float64
	DTS                       float64
}

type NativeInertialSimulationOptions struct {
	Output                                        uint32
	Seed                                          uint64
	InitialAccelBias, InitialGyroBias             [3]float64
	AccelScaleMisalignment, GyroScaleMisalignment [9]float64
	HasRateRandomWalk                             bool
	AccelRandomWalk, GyroRandomWalk               float64
}

type NativeInertialSimulatorState struct {
	AccelBias, GyroBias, AccelRateRandomWalk, GyroRateRandomWalk [3]float64
}

type NativeInertialQuaternion struct{ W, X, Y, Z float64 }
type NativeInertialConstants struct {
	DefaultSeed                              uint64
	GravityEquator, GravityPole, SomiglianaK float64
}
type NativeInertialMechanizer struct {
	_      noCopy
	handle *surfaceHandle
}
type NativeIMUSimulator struct {
	_      noCopy
	handle *surfaceHandle
}
type NativeInertialError struct {
	Kind                uint32
	HasField, HasReason bool
	Field, Reason       string
}
type NativeInertialStatusError struct {
	NativeInertialError
	Cause error
}

func (err *NativeInertialStatusError) Error() string {
	if err == nil {
		return "sidereon: inertial operation failed"
	}
	if err.Cause != nil {
		return fmt.Sprintf("sidereon: inertial operation failed: %v", err.Cause)
	}
	return fmt.Sprintf("sidereon: inertial operation failed (kind %d)", err.Kind)
}
func (err *NativeInertialStatusError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

func inertialErrorTextLocked(part uint32) (string, error) {
	var written, required C.size_t
	status := C.sidereon_inertial_last_error_text(C.uint32_t(part), nil, 0, &written, &required)
	if err := statusErrorLocked(uint32(status)); err != nil {
		return "", err
	}
	n, err := sizeTToInt(required, "inertial error text")
	if err != nil || n == 0 {
		return "", err
	}
	buffer := make([]byte, n)
	status = C.sidereon_inertial_last_error_text(C.uint32_t(part), (*C.uint8_t)(unsafe.Pointer(&buffer[0])), C.size_t(n), &written, &required)
	if err := statusErrorLocked(uint32(status)); err != nil {
		return "", err
	}
	count, err := validateTwoPassCounts("inertial error text", n, n, uint64(written), uint64(required))
	if err != nil {
		return "", err
	}
	return string(buffer[:count]), nil
}

func inertialStatusLocked(status C.enum_SidereonStatus) (NativeInertialError, error) {
	if status == C.SIDEREON_STATUS_OK {
		return NativeInertialError{}, nil
	}
	var raw C.SidereonInertialError
	if getStatus := C.sidereon_inertial_last_error(&raw); getStatus != C.SIDEREON_STATUS_OK {
		return NativeInertialError{}, statusErrorLocked(uint32(getStatus))
	}
	field, err := inertialErrorTextLocked(0)
	if err != nil {
		return NativeInertialError{}, err
	}
	reason, err := inertialErrorTextLocked(1)
	if err != nil {
		return NativeInertialError{}, err
	}
	detail := NativeInertialError{Kind: uint32(raw.kind), HasField: bool(raw.has_field), HasReason: bool(raw.has_reason), Field: field, Reason: reason}
	return detail, statusErrorLocked(uint32(status))
}

func inertialCall(fn func() C.enum_SidereonStatus) (NativeInertialError, error) {
	var detail NativeInertialError
	var operationErr error
	withCThread(func() { detail, operationErr = inertialStatusLocked(fn()) })
	if operationErr != nil && detail.Kind != 0 {
		return detail, &NativeInertialStatusError{NativeInertialError: detail, Cause: operationErr}
	}
	return detail, operationErr
}

func inertialStatusError(status C.enum_SidereonStatus) error {
	detail, err := inertialStatusLocked(status)
	if err != nil && detail.Kind != 0 {
		return &NativeInertialStatusError{NativeInertialError: detail, Cause: err}
	}
	return err
}

func cInertialState(value NativeInertialNavState) C.SidereonInertialNavState {
	var out C.SidereonInertialNavState
	out.t_j2000_s = C.double(value.Epoch)
	for i := 0; i < 3; i++ {
		out.position_ecef_m[i] = C.double(value.Position[i])
		out.velocity_ecef_mps[i] = C.double(value.Velocity[i])
		out.accel_bias_mps2[i] = C.double(value.AccelBias[i])
		out.gyro_bias_rps[i] = C.double(value.GyroBias[i])
	}
	for i := 0; i < 9; i++ {
		out.attitude_body_to_ecef[i] = C.double(value.AttitudeMatrix[i])
	}
	return out
}

func inertialStateFromC(value C.SidereonInertialNavState) NativeInertialNavState {
	var out NativeInertialNavState
	out.Epoch = float64(value.t_j2000_s)
	for i := 0; i < 3; i++ {
		out.Position[i] = float64(value.position_ecef_m[i])
		out.Velocity[i] = float64(value.velocity_ecef_mps[i])
		out.AccelBias[i] = float64(value.accel_bias_mps2[i])
		out.GyroBias[i] = float64(value.gyro_bias_rps[i])
	}
	for i := 0; i < 9; i++ {
		out.AttitudeMatrix[i] = float64(value.attitude_body_to_ecef[i])
	}
	return out
}

func cInertialModel(value NativeInertialIMUModel) C.SidereonInertialImuModel {
	var out C.SidereonInertialImuModel
	for i := 0; i < 3; i++ {
		out.accel_bias_mps2[i] = C.double(value.AccelBias[i])
		out.gyro_bias_rps[i] = C.double(value.GyroBias[i])
	}
	for i := 0; i < 9; i++ {
		out.accel_scale_misalignment[i] = C.double(value.AccelScaleMisalignment[i])
		out.gyro_scale_misalignment[i] = C.double(value.GyroScaleMisalignment[i])
	}
	return out
}

func cInertialIncrement(value NativeInertialIncrement) C.SidereonInertialIncrement {
	var out C.SidereonInertialIncrement
	out.t_j2000_s, out.dt_s = C.double(value.Epoch), C.double(value.DTS)
	for i := 0; i < 3; i++ {
		out.delta_velocity_mps[i] = C.double(value.DeltaVelocity[i])
		out.delta_theta_rad[i] = C.double(value.DeltaTheta[i])
	}
	return out
}

func inertialIncrementFromC(value C.SidereonInertialIncrement) NativeInertialIncrement {
	out := NativeInertialIncrement{Epoch: float64(value.t_j2000_s), DTS: float64(value.dt_s)}
	for i := 0; i < 3; i++ {
		out.DeltaVelocity[i] = float64(value.delta_velocity_mps[i])
		out.DeltaTheta[i] = float64(value.delta_theta_rad[i])
	}
	return out
}

func cInertialSimulationOptions(value NativeInertialSimulationOptions) C.SidereonInertialSimulationOptions {
	var out C.SidereonInertialSimulationOptions
	out.output, out.seed = C.uint32_t(value.Output), C.uint64_t(value.Seed)
	out.has_rate_random_walk = C.bool(value.HasRateRandomWalk)
	out.accel_random_walk_mps2_sqrt_s, out.gyro_random_walk_rps_sqrt_s = C.double(value.AccelRandomWalk), C.double(value.GyroRandomWalk)
	for i := 0; i < 3; i++ {
		out.initial_accel_bias_mps2[i] = C.double(value.InitialAccelBias[i])
		out.initial_gyro_bias_rps[i] = C.double(value.InitialGyroBias[i])
	}
	for i := 0; i < 9; i++ {
		out.accel_scale_misalignment[i] = C.double(value.AccelScaleMisalignment[i])
		out.gyro_scale_misalignment[i] = C.double(value.GyroScaleMisalignment[i])
	}
	return out
}

func inertialSimulatorStateFromC(value C.SidereonInertialSimulatorState) NativeInertialSimulatorState {
	var out NativeInertialSimulatorState
	for i := 0; i < 3; i++ {
		out.AccelBias[i] = float64(value.accel_bias_mps2[i])
		out.GyroBias[i] = float64(value.gyro_bias_rps[i])
		out.AccelRateRandomWalk[i] = float64(value.accel_rate_random_walk_mps2[i])
		out.GyroRateRandomWalk[i] = float64(value.gyro_rate_random_walk_rps[i])
	}
	return out
}

func newInertialMechanizer(pointer *C.SidereonInertialMechanizer) (*NativeInertialMechanizer, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) { C.sidereon_inertial_mechanizer_free((*C.SidereonInertialMechanizer)(p)) })
	if err != nil {
		return nil, err
	}
	return &NativeInertialMechanizer{handle: handle}, nil
}

func newIMUSimulator(pointer *C.SidereonImuSimulator) (*NativeIMUSimulator, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) { C.sidereon_imu_simulator_free((*C.SidereonImuSimulator)(p)) })
	if err != nil {
		return nil, err
	}
	return &NativeIMUSimulator{handle: handle}, nil
}

func InertialMechanizerNew(initial NativeInertialNavState, model *NativeInertialIMUModel, coning uint32) (*NativeInertialMechanizer, error) {
	if coning != 0 {
		return nil, invalidArgument("unsupported inertial coning-correction selector")
	}
	cInitial := cInertialState(initial)
	var pointer *C.SidereonInertialMechanizer
	var operationErr error
	withCThread(func() {
		var status C.enum_SidereonStatus
		if model == nil {
			status = C.sidereon_inertial_mechanizer_new(&cInitial, &pointer)
		} else {
			cModel := cInertialModel(*model)
			status = C.sidereon_inertial_mechanizer_new_with_model(&cInitial, &cModel, &pointer)
		}
		operationErr = inertialStatusError(status)
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newInertialMechanizer(pointer)
}

func InertialMechanizerNewWithConfig(initial NativeInertialNavState, coning uint32) (*NativeInertialMechanizer, error) {
	if coning != 0 {
		return nil, invalidArgument("unsupported inertial coning-correction selector")
	}
	cInitial := cInertialState(initial)
	config := C.SidereonFusionMechanizationConfig{coning_correction: C.uint32_t(coning)}
	var pointer *C.SidereonInertialMechanizer
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_mechanizer_new_with_config(&cInitial, &config, &pointer)
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_inertial_mechanizer_free(pointer) })
		}
		return nil, err
	}
	return newInertialMechanizer(pointer)
}

func (mechanizer *NativeInertialMechanizer) Propagate(sample NativeFusionIMUSample) (NativeInertialNavState, error) {
	if mechanizer == nil || mechanizer.handle == nil {
		return NativeInertialNavState{}, ErrClosed
	}
	cSample := cFusionSample(sample)
	var output C.SidereonInertialNavState
	err := mechanizer.handle.write(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return inertialStatusError(C.sidereon_inertial_mechanizer_propagate((*C.SidereonInertialMechanizer)(pointer), &cSample, &output))
		})
	})
	return inertialStateFromC(output), err
}

func (mechanizer *NativeInertialMechanizer) State() (NativeInertialNavState, error) {
	if mechanizer == nil || mechanizer.handle == nil {
		return NativeInertialNavState{}, ErrClosed
	}
	var output C.SidereonInertialNavState
	err := mechanizer.handle.read(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return inertialStatusError(C.sidereon_inertial_mechanizer_state((*C.SidereonInertialMechanizer)(pointer), &output))
		})
	})
	return inertialStateFromC(output), err
}

func (mechanizer *NativeInertialMechanizer) Close() error {
	if mechanizer == nil || mechanizer.handle == nil {
		return nil
	}
	return mechanizer.handle.close()
}

func InertialMechanizeECEF(state NativeInertialNavState, increment NativeInertialIncrement, coning uint32) (NativeInertialNavState, error) {
	if coning != 0 {
		return NativeInertialNavState{}, invalidArgument("unsupported inertial coning-correction selector")
	}
	cState, cIncrement := cInertialState(state), cInertialIncrement(increment)
	config := C.SidereonFusionMechanizationConfig{coning_correction: C.uint32_t(coning)}
	var output C.SidereonInertialNavState
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_mechanize_ecef(&cState, &cIncrement, &config, &output)
	})
	return inertialStateFromC(output), err
}

func InertialNormalGravity(latitude, height float64) (float64, error) {
	var output C.double
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_normal_gravity(C.double(latitude), C.double(height), &output)
	})
	return float64(output), err
}

func InertialGravityECEF(position [3]float64) ([3]float64, error) {
	cPosition := cFusionArray3(position)
	var output [3]C.double
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_gravity_ecef(&cPosition[0], &output[0]) })
	return [3]float64{float64(output[0]), float64(output[1]), float64(output[2])}, err
}

func InertialTrueIncrementBetween(start, end NativeInertialNavState) (NativeInertialIncrement, error) {
	cStart, cEnd := cInertialState(start), cInertialState(end)
	var output C.SidereonInertialIncrement
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_true_increment_between(&cStart, &cEnd, &output)
	})
	return inertialIncrementFromC(output), err
}

func InertialIMUPreset(grade uint32) (NativeFusionIMUSpec, error) {
	var value C.SidereonFusionImuSpec
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_imu_spec_preset(C.uint32_t(grade), &value) })
	return NativeFusionIMUSpec{AccelVRW: float64(value.accel_vrw_mps_sqrt_s), GyroARW: float64(value.gyro_arw_rad_sqrt_s), AccelBiasInstability: float64(value.accel_bias_instab_mps2), GyroBiasInstability: float64(value.gyro_bias_instab_rps), AccelBiasTau: float64(value.accel_bias_tau_s), GyroBiasTau: float64(value.gyro_bias_tau_s), HasAccelScale: bool(value.has_accel_scale_instab_ppm), AccelScalePPM: float64(value.accel_scale_instab_ppm), HasGyroScale: bool(value.has_gyro_scale_instab_ppm), GyroScalePPM: float64(value.gyro_scale_instab_ppm)}, err
}

func InertialIMUSpecBiasStatistics(spec NativeFusionIMUSpec, deltaSeconds float64) ([4]float64, error) {
	cSpec := cFusionSpec(spec)
	var output [4]C.double
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_imu_spec_bias_statistics(&cSpec, C.double(deltaSeconds), &output[0])
	})
	return [4]float64{float64(output[0]), float64(output[1]), float64(output[2]), float64(output[3])}, err
}

func InertialIMUSpecValidate(spec NativeFusionIMUSpec) error {
	value := cFusionSpec(spec)
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_imu_spec_validate(&value) })
	return err
}

func InertialRateRandomWalkValidate(accel, gyro float64) error {
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_rate_random_walk_validate(C.double(accel), C.double(gyro))
	})
	return err
}

func InertialConstants() (NativeInertialConstants, error) {
	var value C.SidereonInertialConstants
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_constants(&value) })
	return NativeInertialConstants{DefaultSeed: uint64(value.default_imu_sim_seed), GravityEquator: float64(value.normal_gravity_equator_mps2), GravityPole: float64(value.normal_gravity_pole_mps2), SomiglianaK: float64(value.somigliana_k)}, err
}

func InertialCalibrationFromScalePPM(accel, gyro [3]float64) (NativeInertialIMUModel, error) {
	cAccel, cGyro := cFusionArray3(accel), cFusionArray3(gyro)
	var value C.SidereonInertialImuModel
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_calibration_from_scale_ppm(&cAccel[0], &cGyro[0], &value)
	})
	out := NativeInertialIMUModel{}
	for i := 0; i < 3; i++ {
		out.AccelBias[i] = float64(value.accel_bias_mps2[i])
		out.GyroBias[i] = float64(value.gyro_bias_rps[i])
	}
	for i := 0; i < 9; i++ {
		out.AccelScaleMisalignment[i] = float64(value.accel_scale_misalignment[i])
		out.GyroScaleMisalignment[i] = float64(value.gyro_scale_misalignment[i])
	}
	return out, err
}

func InertialIMUModelValidate(model NativeInertialIMUModel) error {
	value := cInertialModel(model)
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_imu_model_validate(&value) })
	return err
}

func InertialQuaternionFromDCM(dcm [9]float64) (NativeInertialQuaternion, error) {
	cValue := cFusionArray9(dcm)
	var output C.SidereonInertialQuaternion
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_dcm_to_quaternion(&cValue[0], &output) })
	return NativeInertialQuaternion{W: float64(output.w), X: float64(output.x), Y: float64(output.y), Z: float64(output.z)}, err
}

func InertialDCMFromQuaternion(value NativeInertialQuaternion) ([9]float64, error) {
	cValue := C.SidereonInertialQuaternion{w: C.double(value.W), x: C.double(value.X), y: C.double(value.Y), z: C.double(value.Z)}
	var output [9]C.double
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_quaternion_to_dcm(&cValue, &output[0]) })
	var result [9]float64
	for i := range result {
		result[i] = float64(output[i])
	}
	return result, err
}

func InertialReorthonormalizeDCM(value [9]float64) ([9]float64, error) {
	cValue := cFusionArray9(value)
	var output [9]C.double
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_reorthonormalize_dcm(&cValue[0], &output[0]) })
	var result [9]float64
	for i := range result {
		result[i] = float64(output[i])
	}
	return result, err
}

func InertialYawPitchRoll(value [9]float64) ([3]float64, error) {
	cValue := cFusionArray9(value)
	var output [3]C.double
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_attitude_yaw_pitch_roll(&cValue[0], &output[0])
	})
	return [3]float64{float64(output[0]), float64(output[1]), float64(output[2])}, err
}

func InertialRodrigues(delta [3]float64) ([9]float64, error) {
	cDelta := cFusionArray3(delta)
	var output [9]C.double
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_inertial_rodrigues_delta_dcm(&cDelta[0], &output[0]) })
	var result [9]float64
	for i := range result {
		result[i] = float64(output[i])
	}
	return result, err
}

func InertialCorrectSample(sample NativeFusionIMUSample, previousEpoch float64, model NativeInertialIMUModel) (NativeInertialIncrement, error) {
	cSample, cModel := cFusionSample(sample), cInertialModel(model)
	var output C.SidereonInertialIncrement
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_correct_sample(&cSample, C.double(previousEpoch), &cModel, &output)
	})
	return inertialIncrementFromC(output), err
}

func InertialIMUSimulatorNew(spec NativeFusionIMUSpec, options NativeInertialSimulationOptions) (*NativeIMUSimulator, error) {
	cSpec, cOptions := cFusionSpec(spec), cInertialSimulationOptions(options)
	var pointer *C.SidereonImuSimulator
	_, err := inertialCall(func() C.enum_SidereonStatus { return C.sidereon_imu_simulator_new(&cSpec, &cOptions, &pointer) })
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_imu_simulator_free(pointer) })
		}
		return nil, err
	}
	return newIMUSimulator(pointer)
}

func (simulator *NativeIMUSimulator) SampleIncrement(truth NativeInertialIncrement) (NativeFusionIMUSample, NativeInertialSimulatorState, error) {
	if simulator == nil || simulator.handle == nil {
		return NativeFusionIMUSample{}, NativeInertialSimulatorState{}, ErrClosed
	}
	cTruth := cInertialIncrement(truth)
	var sample C.SidereonFusionImuSample
	var state C.SidereonInertialSimulatorState
	err := simulator.handle.write(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return inertialStatusError(C.sidereon_imu_simulator_sample_increment((*C.SidereonImuSimulator)(pointer), &cTruth, &sample, &state))
		})
	})
	if err != nil {
		return NativeFusionIMUSample{}, NativeInertialSimulatorState{}, err
	}
	output, states := inertialSequenceFromC([]C.SidereonFusionImuSample{sample}, []C.SidereonInertialSimulatorState{state})
	return output[0], states[0], nil
}

func (simulator *NativeIMUSimulator) State() (NativeInertialSimulatorState, error) {
	if simulator == nil || simulator.handle == nil {
		return NativeInertialSimulatorState{}, ErrClosed
	}
	var output C.SidereonInertialSimulatorState
	err := simulator.handle.read(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return inertialStatusError(C.sidereon_imu_simulator_state((*C.SidereonImuSimulator)(pointer), &output))
		})
	})
	return inertialSimulatorStateFromC(output), err
}

func (simulator *NativeIMUSimulator) Close() error {
	if simulator == nil || simulator.handle == nil {
		return nil
	}
	return simulator.handle.close()
}

func InertialSimulateTrajectory(trajectory []NativeInertialNavState, spec NativeFusionIMUSpec, options NativeInertialSimulationOptions) ([]NativeFusionIMUSample, []NativeInertialSimulatorState, error) {
	cStates := make([]C.SidereonInertialNavState, len(trajectory))
	for index, state := range trajectory {
		cStates[index] = cInertialState(state)
	}
	outputCount := 0
	if len(trajectory) > 0 {
		outputCount = len(trajectory) - 1
	}
	samples := make([]C.SidereonFusionImuSample, outputCount)
	states := make([]C.SidereonInertialSimulatorState, outputCount)
	var statePointer *C.SidereonInertialNavState
	if len(cStates) > 0 {
		statePointer = &cStates[0]
	}
	var samplePointer *C.SidereonFusionImuSample
	var historyPointer *C.SidereonInertialSimulatorState
	if outputCount > 0 {
		samplePointer = &samples[0]
		historyPointer = &states[0]
	}
	cSpec, cOptions := cFusionSpec(spec), cInertialSimulationOptions(options)
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_simulate_trajectory(statePointer, C.size_t(len(cStates)), &cSpec, &cOptions, samplePointer, historyPointer)
	})
	if err != nil {
		return nil, nil, err
	}
	outSamples, outStates := inertialSequenceFromC(samples, states)
	return outSamples, outStates, nil
}

func InertialSimulateIncrements(increments []NativeInertialIncrement, spec NativeFusionIMUSpec, options NativeInertialSimulationOptions) ([]NativeFusionIMUSample, []NativeInertialSimulatorState, error) {
	cIncrements := make([]C.SidereonInertialIncrement, len(increments))
	for i, value := range increments {
		cIncrements[i] = cInertialIncrement(value)
	}
	samples := make([]C.SidereonFusionImuSample, len(increments))
	states := make([]C.SidereonInertialSimulatorState, len(increments))
	cSpec, cOptions := cFusionSpec(spec), cInertialSimulationOptions(options)
	var incrementPointer *C.SidereonInertialIncrement
	if len(cIncrements) != 0 {
		incrementPointer = &cIncrements[0]
	}
	var outputPointer *C.SidereonFusionImuSample
	var statePointer *C.SidereonInertialSimulatorState
	if len(samples) != 0 {
		outputPointer = &samples[0]
		statePointer = &states[0]
	}
	_, err := inertialCall(func() C.enum_SidereonStatus {
		return C.sidereon_inertial_simulate_increments(incrementPointer, C.size_t(len(increments)), &cSpec, &cOptions, outputPointer, statePointer)
	})
	if err != nil {
		return nil, nil, err
	}
	outSamples, outStates := inertialSequenceFromC(samples, states)
	return outSamples, outStates, nil
}

func inertialSequenceFromC(samples []C.SidereonFusionImuSample, states []C.SidereonInertialSimulatorState) ([]NativeFusionIMUSample, []NativeInertialSimulatorState) {
	outSamples := make([]NativeFusionIMUSample, len(samples))
	outStates := make([]NativeInertialSimulatorState, len(states))
	for i, value := range samples {
		outSamples[i] = NativeFusionIMUSample{Epoch: float64(value.t_j2000_s), Kind: uint32(value.kind), DTS: float64(value.dt_s)}
		for axis := 0; axis < 3; axis++ {
			outSamples[i].SpecificForce[axis] = float64(value.specific_force_mps2[axis])
			outSamples[i].AngularRate[axis] = float64(value.angular_rate_rps[axis])
			outSamples[i].DeltaVelocity[axis] = float64(value.delta_velocity_mps[axis])
			outSamples[i].DeltaTheta[axis] = float64(value.delta_theta_rad[axis])
		}
		outStates[i] = inertialSimulatorStateFromC(states[i])
	}
	return outSamples, outStates
}

func inertialHandle(pointer unsafe.Pointer, release func(unsafe.Pointer)) (*surfaceHandle, error) {
	return newSurfaceHandle(pointer, release)
}
