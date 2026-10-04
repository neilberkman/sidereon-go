package sidereon

import (
	"errors"
	"fmt"

	"sidereon.dev/go/v3/internal/native"
)

// InertialNavState stores position, velocity, and body attitude in ECEF at a J2000 epoch, together with IMU biases.
type InertialNavState struct {
	EpochJ2000S                      float64
	PositionECEFM, VelocityECEFMPerS [3]float64
	AttitudeBodyToECEF               [9]float64
	AccelBiasMPerS2, GyroBiasRadPerS [3]float64
}

// InertialIMUModel stores accelerometer and gyroscope bias and scale-misalignment corrections.
type InertialIMUModel struct {
	AccelBiasMPerS2, GyroBiasRadPerS              [3]float64
	AccelScaleMisalignment, GyroScaleMisalignment [9]float64
}

// InertialIncrement stores the integrated velocity and angle increments over one IMU interval.
type InertialIncrement struct {
	EpochJ2000S                       float64
	DeltaVelocityMPerS, DeltaThetaRad [3]float64
	DTS                               float64
}

// InertialSimulationOutput selects whether the simulator emits rates or integrated increments.
type InertialSimulationOutput uint32

const (
	// InertialSimulationRate makes the simulator emit sampled rates.
	InertialSimulationRate InertialSimulationOutput = iota
	// InertialSimulationIncrement makes the simulator emit integrated increments instead of rates.
	InertialSimulationIncrement
)

// InertialConingCorrection selects the coning correction used by the mechanizer.
type InertialConingCorrection uint32

// InertialConingCorrectionOff disables coning correction in the mechanizer.
const InertialConingCorrectionOff InertialConingCorrection = 0

// InertialSimulationOptions sets deterministic simulation output, initial errors, scale terms, and random walks.
type InertialSimulationOptions struct {
	Output                                                         InertialSimulationOutput
	Seed                                                           uint64
	InitialAccelBiasMPerS2, InitialGyroBiasRadPerS                 [3]float64
	AccelScaleMisalignment, GyroScaleMisalignment                  [9]float64
	HasRateRandomWalk                                              bool
	AccelRateRandomWalkMPerS2SqrtS, GyroRateRandomWalkRadPerSSqrtS float64
}

// InertialSimulatorState reports the simulated sensor bias and rate-random-walk state.
type InertialSimulatorState struct {
	AccelBiasMPerS2, GyroBiasRadPerS                     [3]float64
	AccelRateRandomWalkMPerS2, GyroRateRandomWalkRadPerS [3]float64
}

// InertialQuaternion stores a scalar-first rotation quaternion.
type InertialQuaternion struct{ W, X, Y, Z float64 }

// InertialConstants contains simulator and normal-gravity constants used by the inertial routines.
type InertialConstants struct {
	DefaultIMUSimulatorSeed                                          uint64
	NormalGravityEquatorMPerS2, NormalGravityPoleMPerS2, SomiglianaK float64
}

// InertialError preserves a typed inertial failure and its underlying cause.
type InertialError struct {
	Kind                uint32
	HasField, HasReason bool
	Field, Reason       string
	Cause               error
}

// Error returns the InertialError message.
func (err *InertialError) Error() string {
	if err == nil {
		return "sidereon: inertial operation failed"
	}
	if err.Cause != nil {
		return fmt.Sprintf("sidereon: inertial operation failed: %v", err.Cause)
	}
	return fmt.Sprintf("sidereon: inertial operation failed (kind %d)", err.Kind)
}

// Unwrap returns the underlying cause when one is retained.
func (err *InertialError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}
func publicInertialError(err error) error {
	var nativeErr *native.NativeInertialStatusError
	if errors.As(err, &nativeErr) {
		return &InertialError{Kind: nativeErr.Kind, HasField: nativeErr.HasField, HasReason: nativeErr.HasReason, Field: nativeErr.Field, Reason: nativeErr.Reason, Cause: publicError(nativeErr.Cause)}
	}
	return publicError(err)
}

func nativeInertialState(value InertialNavState) native.NativeInertialNavState {
	return native.NativeInertialNavState{Epoch: value.EpochJ2000S, Position: value.PositionECEFM, Velocity: value.VelocityECEFMPerS, AttitudeMatrix: value.AttitudeBodyToECEF, AccelBias: value.AccelBiasMPerS2, GyroBias: value.GyroBiasRadPerS}
}
func publicInertialState(value native.NativeInertialNavState) InertialNavState {
	return InertialNavState{EpochJ2000S: value.Epoch, PositionECEFM: value.Position, VelocityECEFMPerS: value.Velocity, AttitudeBodyToECEF: value.AttitudeMatrix, AccelBiasMPerS2: value.AccelBias, GyroBiasRadPerS: value.GyroBias}
}
func nativeInertialIncrement(value InertialIncrement) native.NativeInertialIncrement {
	return native.NativeInertialIncrement{Epoch: value.EpochJ2000S, DeltaVelocity: value.DeltaVelocityMPerS, DeltaTheta: value.DeltaThetaRad, DTS: value.DTS}
}
func publicInertialIncrement(value native.NativeInertialIncrement) InertialIncrement {
	return InertialIncrement{EpochJ2000S: value.Epoch, DeltaVelocityMPerS: value.DeltaVelocity, DeltaThetaRad: value.DeltaTheta, DTS: value.DTS}
}
func nativeInertialModel(value InertialIMUModel) native.NativeInertialIMUModel {
	return native.NativeInertialIMUModel{AccelBias: value.AccelBiasMPerS2, GyroBias: value.GyroBiasRadPerS, AccelScaleMisalignment: value.AccelScaleMisalignment, GyroScaleMisalignment: value.GyroScaleMisalignment}
}
func nativeInertialSimulationOptions(value InertialSimulationOptions) native.NativeInertialSimulationOptions {
	return native.NativeInertialSimulationOptions{Output: uint32(value.Output), Seed: value.Seed, InitialAccelBias: value.InitialAccelBiasMPerS2, InitialGyroBias: value.InitialGyroBiasRadPerS, AccelScaleMisalignment: value.AccelScaleMisalignment, GyroScaleMisalignment: value.GyroScaleMisalignment, HasRateRandomWalk: value.HasRateRandomWalk, AccelRandomWalk: value.AccelRateRandomWalkMPerS2SqrtS, GyroRandomWalk: value.GyroRateRandomWalkRadPerSSqrtS}
}
func publicInertialSimulatorState(value native.NativeInertialSimulatorState) InertialSimulatorState {
	return InertialSimulatorState{AccelBiasMPerS2: value.AccelBias, GyroBiasRadPerS: value.GyroBias, AccelRateRandomWalkMPerS2: value.AccelRateRandomWalk, GyroRateRandomWalkRadPerS: value.GyroRateRandomWalk}
}
func nativeInertialSample(value FusionIMUSample) native.NativeFusionIMUSample {
	return native.NativeFusionIMUSample{Epoch: value.EpochJ2000S, Kind: uint32(value.Kind), SpecificForce: value.SpecificForceMPerS2, AngularRate: value.AngularRateRadPerS, DeltaVelocity: value.DeltaVelocityMPerS, DeltaTheta: value.DeltaThetaRad, DTS: value.DTS}
}
func nativeFusionIMUSpec(value FusionIMUSpec) native.NativeFusionIMUSpec {
	return native.NativeFusionIMUSpec{AccelVRW: value.AccelVRWMPerSqrtS, GyroARW: value.GyroARRadPerSqrtS, AccelBiasInstability: value.AccelBiasInstabilityMPerS2, GyroBiasInstability: value.GyroBiasInstabilityRadPerS, AccelBiasTau: value.AccelBiasTauS, GyroBiasTau: value.GyroBiasTauS, HasAccelScale: value.HasAccelScaleInstability, AccelScalePPM: value.AccelScaleInstabilityPPM, HasGyroScale: value.HasGyroScaleInstability, GyroScalePPM: value.GyroScaleInstabilityPPM}
}
func publicInertialSample(value native.NativeFusionIMUSample) FusionIMUSample {
	return FusionIMUSample{EpochJ2000S: value.Epoch, Kind: FusionIMUSampleKind(value.Kind), SpecificForceMPerS2: value.SpecificForce, AngularRateRadPerS: value.AngularRate, DeltaVelocityMPerS: value.DeltaVelocity, DeltaThetaRad: value.DeltaTheta, DTS: value.DTS}
}

// InertialMechanizer owns a stateful native ECEF mechanizer and must be closed when no longer needed.
type InertialMechanizer struct {
	_      noCopy
	handle *native.NativeInertialMechanizer
}

// NewInertialMechanizer creates an ECEF mechanizer from an initial navigation state and optional sensor model.
func NewInertialMechanizer(initial InertialNavState, model *InertialIMUModel) (*InertialMechanizer, error) {
	var nativeModel *native.NativeInertialIMUModel
	if model != nil {
		value := nativeInertialModel(*model)
		nativeModel = &value
	}
	handle, err := native.InertialMechanizerNew(nativeInertialState(initial), nativeModel, 0)
	if err != nil {
		return nil, publicInertialError(err)
	}
	return &InertialMechanizer{handle: handle}, nil
}

// NewInertialMechanizerWithConfig creates a mechanizer with the selected supported coning configuration.
func NewInertialMechanizerWithConfig(initial InertialNavState, coning InertialConingCorrection) (*InertialMechanizer, error) {
	if coning != InertialConingCorrectionOff {
		return nil, fmt.Errorf("sidereon: unsupported inertial coning-correction selector %d", coning)
	}
	handle, err := native.InertialMechanizerNewWithConfig(nativeInertialState(initial), uint32(coning))
	if err != nil {
		return nil, publicInertialError(err)
	}
	return &InertialMechanizer{handle: handle}, nil
}

// Propagate applies one IMU sample and returns the updated navigation state.
func (mechanizer *InertialMechanizer) Propagate(sample FusionIMUSample) (InertialNavState, error) {
	if mechanizer == nil || mechanizer.handle == nil {
		return InertialNavState{}, ErrClosed
	}
	value, err := mechanizer.handle.Propagate(nativeInertialSample(sample))
	return publicInertialState(value), publicInertialError(err)
}

// State returns a copy of the current navigation or simulator state.
func (mechanizer *InertialMechanizer) State() (InertialNavState, error) {
	if mechanizer == nil || mechanizer.handle == nil {
		return InertialNavState{}, ErrClosed
	}
	value, err := mechanizer.handle.State()
	return publicInertialState(value), publicInertialError(err)
}

// Close releases the owned native resource; repeated calls are safe.
func (mechanizer *InertialMechanizer) Close() error {
	if mechanizer == nil || mechanizer.handle == nil {
		return nil
	}
	return publicInertialError(mechanizer.handle.Close())
}

// IMUSimulator owns a stateful native IMU simulator and must be closed when no longer needed.
type IMUSimulator struct {
	_      noCopy
	handle *native.NativeIMUSimulator
}

// NewIMUSimulator creates a deterministic IMU simulator from a sensor specification and options.
func NewIMUSimulator(spec FusionIMUSpec, options InertialSimulationOptions) (*IMUSimulator, error) {
	if options.Output > InertialSimulationIncrement {
		return nil, fmt.Errorf("sidereon: invalid inertial simulation output %d", options.Output)
	}
	handle, err := native.InertialIMUSimulatorNew(nativeFusionIMUSpec(spec), nativeInertialSimulationOptions(options))
	if err != nil {
		return nil, publicInertialError(err)
	}
	return &IMUSimulator{handle: handle}, nil
}

// SampleIncrement simulates sensor output for one truth increment and returns the resulting simulator state.
func (simulator *IMUSimulator) SampleIncrement(truth InertialIncrement) (FusionIMUSample, InertialSimulatorState, error) {
	if simulator == nil || simulator.handle == nil {
		return FusionIMUSample{}, InertialSimulatorState{}, ErrClosed
	}
	sample, state, err := simulator.handle.SampleIncrement(nativeInertialIncrement(truth))
	return publicInertialSample(sample), publicInertialSimulatorState(state), publicInertialError(err)
}

// State returns a copy of the current navigation or simulator state.
func (simulator *IMUSimulator) State() (InertialSimulatorState, error) {
	if simulator == nil || simulator.handle == nil {
		return InertialSimulatorState{}, ErrClosed
	}
	value, err := simulator.handle.State()
	return publicInertialSimulatorState(value), publicInertialError(err)
}

// Close releases the owned native resource; repeated calls are safe.
func (simulator *IMUSimulator) Close() error {
	if simulator == nil || simulator.handle == nil {
		return nil
	}
	return publicInertialError(simulator.handle.Close())
}

// InertialIMUPreset returns the sensor specification for a named IMU grade.
func InertialIMUPreset(grade FusionIMUGrade) (FusionIMUSpec, error) {
	value, err := native.InertialIMUPreset(uint32(grade))
	if err != nil {
		return FusionIMUSpec{}, publicInertialError(err)
	}
	return FusionIMUSpec{AccelVRWMPerSqrtS: value.AccelVRW, GyroARRadPerSqrtS: value.GyroARW, AccelBiasInstabilityMPerS2: value.AccelBiasInstability, GyroBiasInstabilityRadPerS: value.GyroBiasInstability, AccelBiasTauS: value.AccelBiasTau, GyroBiasTauS: value.GyroBiasTau, HasAccelScaleInstability: value.HasAccelScale, AccelScaleInstabilityPPM: value.AccelScalePPM, HasGyroScaleInstability: value.HasGyroScale, GyroScaleInstabilityPPM: value.GyroScalePPM}, nil
}

// ValidateInertialIMUSpec checks the IMU noise and bias specification.
func ValidateInertialIMUSpec(spec FusionIMUSpec) error {
	return publicInertialError(native.InertialIMUSpecValidate(nativeFusionIMUSpec(spec)))
}

// ValidateInertialIMUModel checks bias and scale-misalignment corrections.
func ValidateInertialIMUModel(model InertialIMUModel) error {
	return publicInertialError(native.InertialIMUModelValidate(nativeInertialModel(model)))
}

// ValidateInertialRateRandomWalk checks accelerometer and gyroscope rate-random-walk values.
func ValidateInertialRateRandomWalk(accel, gyro float64) error {
	return publicInertialError(native.InertialRateRandomWalkValidate(accel, gyro))
}

// InertialIMUBiasStatistics evaluates expected bias statistics over the requested interval.
func InertialIMUBiasStatistics(spec FusionIMUSpec, deltaSeconds float64) ([4]float64, error) {
	value, err := native.InertialIMUSpecBiasStatistics(nativeFusionIMUSpec(spec), deltaSeconds)
	return value, publicInertialError(err)
}

// InertialModelConstants returns the pinned simulator seed and normal-gravity constants.
func InertialModelConstants() (InertialConstants, error) {
	value, err := native.InertialConstants()
	if err != nil {
		return InertialConstants{}, publicInertialError(err)
	}
	return InertialConstants{DefaultIMUSimulatorSeed: value.DefaultSeed, NormalGravityEquatorMPerS2: value.GravityEquator, NormalGravityPoleMPerS2: value.GravityPole, SomiglianaK: value.SomiglianaK}, nil
}

// InertialCalibrationFromScalePPM converts axis scale errors in ppm into an IMU correction model.
func InertialCalibrationFromScalePPM(accel, gyro [3]float64) (InertialIMUModel, error) {
	value, err := native.InertialCalibrationFromScalePPM(accel, gyro)
	if err != nil {
		return InertialIMUModel{}, publicInertialError(err)
	}
	return InertialIMUModel{AccelBiasMPerS2: value.AccelBias, GyroBiasRadPerS: value.GyroBias, AccelScaleMisalignment: value.AccelScaleMisalignment, GyroScaleMisalignment: value.GyroScaleMisalignment}, nil
}

// InertialCorrectSample removes modeled sensor errors and integrates one IMU observation.
func InertialCorrectSample(sample FusionIMUSample, previousEpoch float64, model InertialIMUModel) (InertialIncrement, error) {
	value, err := native.InertialCorrectSample(nativeInertialSample(sample), previousEpoch, nativeInertialModel(model))
	return publicInertialIncrement(value), publicInertialError(err)
}

// InertialMechanizeECEF propagates an ECEF navigation state by one corrected increment.
func InertialMechanizeECEF(state InertialNavState, increment InertialIncrement) (InertialNavState, error) {
	value, err := native.InertialMechanizeECEF(nativeInertialState(state), nativeInertialIncrement(increment), 0)
	return publicInertialState(value), publicInertialError(err)
}

// InertialNormalGravity returns normal gravity at geodetic latitude and height.
func InertialNormalGravity(latitudeRad, heightM float64) (float64, error) {
	value, err := native.InertialNormalGravity(latitudeRad, heightM)
	return value, publicInertialError(err)
}

// InertialGravityECEF returns the gravity vector at an ECEF position.
func InertialGravityECEF(positionECEFM [3]float64) ([3]float64, error) {
	value, err := native.InertialGravityECEF(positionECEFM)
	return value, publicInertialError(err)
}

// InertialTrueIncrementBetween derives the ideal IMU increment between two navigation states.
func InertialTrueIncrementBetween(start, end InertialNavState) (InertialIncrement, error) {
	value, err := native.InertialTrueIncrementBetween(nativeInertialState(start), nativeInertialState(end))
	return publicInertialIncrement(value), publicInertialError(err)
}

// InertialQuaternionFromDCM converts a direction-cosine matrix to a normalized quaternion.
func InertialQuaternionFromDCM(dcm [9]float64) (InertialQuaternion, error) {
	value, err := native.InertialQuaternionFromDCM(dcm)
	return InertialQuaternion{W: value.W, X: value.X, Y: value.Y, Z: value.Z}, publicInertialError(err)
}

// InertialDCMFromQuaternion converts a quaternion to a direction-cosine matrix.
func InertialDCMFromQuaternion(value InertialQuaternion) ([9]float64, error) {
	dcm, err := native.InertialDCMFromQuaternion(native.NativeInertialQuaternion{W: value.W, X: value.X, Y: value.Y, Z: value.Z})
	return dcm, publicInertialError(err)
}

// InertialReorthonormalizeDCM projects a near-rotation matrix onto an orthonormal rotation matrix.
func InertialReorthonormalizeDCM(dcm [9]float64) ([9]float64, error) {
	value, err := native.InertialReorthonormalizeDCM(dcm)
	return value, publicInertialError(err)
}

// InertialYawPitchRoll extracts yaw, pitch, and roll angles from a direction-cosine matrix.
func InertialYawPitchRoll(dcm [9]float64) ([3]float64, error) {
	value, err := native.InertialYawPitchRoll(dcm)
	return value, publicInertialError(err)
}

// InertialRodrigues computes the rotation matrix for a rotation vector.
func InertialRodrigues(deltaThetaRad [3]float64) ([9]float64, error) {
	value, err := native.InertialRodrigues(deltaThetaRad)
	return value, publicInertialError(err)
}

// InertialSimulateIncrements generates noisy IMU samples for supplied truth increments.
func InertialSimulateIncrements(increments []InertialIncrement, spec FusionIMUSpec, options InertialSimulationOptions) ([]FusionIMUSample, []InertialSimulatorState, error) {
	values := make([]native.NativeInertialIncrement, len(increments))
	for i, value := range increments {
		values[i] = nativeInertialIncrement(value)
	}
	samples, states, err := native.InertialSimulateIncrements(values, nativeFusionIMUSpec(spec), nativeInertialSimulationOptions(options))
	if err != nil {
		return nil, nil, publicInertialError(err)
	}
	outSamples, outStates := publicInertialSequence(samples, states)
	return outSamples, outStates, nil
}

// InertialSimulateTrajectory generates noisy IMU samples from successive navigation states.
func InertialSimulateTrajectory(trajectory []InertialNavState, spec FusionIMUSpec, options InertialSimulationOptions) ([]FusionIMUSample, []InertialSimulatorState, error) {
	values := make([]native.NativeInertialNavState, len(trajectory))
	for i, value := range trajectory {
		values[i] = nativeInertialState(value)
	}
	samples, states, err := native.InertialSimulateTrajectory(values, nativeFusionIMUSpec(spec), nativeInertialSimulationOptions(options))
	if err != nil {
		return nil, nil, publicInertialError(err)
	}
	outSamples, outStates := publicInertialSequence(samples, states)
	return outSamples, outStates, nil
}
func publicInertialSequence(samples []native.NativeFusionIMUSample, states []native.NativeInertialSimulatorState) ([]FusionIMUSample, []InertialSimulatorState) {
	outSamples := make([]FusionIMUSample, len(samples))
	outStates := make([]InertialSimulatorState, len(states))
	for i := range samples {
		outSamples[i] = publicInertialSample(samples[i])
		outStates[i] = publicInertialSimulatorState(states[i])
	}
	return outSamples, outStates
}
