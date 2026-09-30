package sidereon

import (
	"errors"
	"fmt"

	"sidereon.dev/go/v3/internal/native"
)

type InertialNavState struct {
	EpochJ2000S                      float64
	PositionECEFM, VelocityECEFMPerS [3]float64
	AttitudeBodyToECEF               [9]float64
	AccelBiasMPerS2, GyroBiasRadPerS [3]float64
}

type InertialIMUModel struct {
	AccelBiasMPerS2, GyroBiasRadPerS              [3]float64
	AccelScaleMisalignment, GyroScaleMisalignment [9]float64
}

type InertialIncrement struct {
	EpochJ2000S                       float64
	DeltaVelocityMPerS, DeltaThetaRad [3]float64
	DTS                               float64
}

type InertialSimulationOutput uint32

const (
	InertialSimulationRate InertialSimulationOutput = iota
	InertialSimulationIncrement
)

type InertialConingCorrection uint32

const InertialConingCorrectionOff InertialConingCorrection = 0

type InertialSimulationOptions struct {
	Output                                                         InertialSimulationOutput
	Seed                                                           uint64
	InitialAccelBiasMPerS2, InitialGyroBiasRadPerS                 [3]float64
	AccelScaleMisalignment, GyroScaleMisalignment                  [9]float64
	HasRateRandomWalk                                              bool
	AccelRateRandomWalkMPerS2SqrtS, GyroRateRandomWalkRadPerSSqrtS float64
}

type InertialSimulatorState struct {
	AccelBiasMPerS2, GyroBiasRadPerS                     [3]float64
	AccelRateRandomWalkMPerS2, GyroRateRandomWalkRadPerS [3]float64
}

type InertialQuaternion struct{ W, X, Y, Z float64 }
type InertialConstants struct {
	DefaultIMUSimulatorSeed                                          uint64
	NormalGravityEquatorMPerS2, NormalGravityPoleMPerS2, SomiglianaK float64
}
type InertialError struct {
	Kind                uint32
	HasField, HasReason bool
	Field, Reason       string
	Cause               error
}

func (err *InertialError) Error() string {
	if err == nil {
		return "sidereon: inertial operation failed"
	}
	if err.Cause != nil {
		return fmt.Sprintf("sidereon: inertial operation failed: %v", err.Cause)
	}
	return fmt.Sprintf("sidereon: inertial operation failed (kind %d)", err.Kind)
}
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

type InertialMechanizer struct {
	_      noCopy
	handle *native.NativeInertialMechanizer
}

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
func (mechanizer *InertialMechanizer) Propagate(sample FusionIMUSample) (InertialNavState, error) {
	if mechanizer == nil || mechanizer.handle == nil {
		return InertialNavState{}, ErrClosed
	}
	value, err := mechanizer.handle.Propagate(nativeInertialSample(sample))
	return publicInertialState(value), publicInertialError(err)
}
func (mechanizer *InertialMechanizer) State() (InertialNavState, error) {
	if mechanizer == nil || mechanizer.handle == nil {
		return InertialNavState{}, ErrClosed
	}
	value, err := mechanizer.handle.State()
	return publicInertialState(value), publicInertialError(err)
}
func (mechanizer *InertialMechanizer) Close() error {
	if mechanizer == nil || mechanizer.handle == nil {
		return nil
	}
	return publicInertialError(mechanizer.handle.Close())
}

type IMUSimulator struct {
	_      noCopy
	handle *native.NativeIMUSimulator
}

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
func (simulator *IMUSimulator) SampleIncrement(truth InertialIncrement) (FusionIMUSample, InertialSimulatorState, error) {
	if simulator == nil || simulator.handle == nil {
		return FusionIMUSample{}, InertialSimulatorState{}, ErrClosed
	}
	sample, state, err := simulator.handle.SampleIncrement(nativeInertialIncrement(truth))
	return publicInertialSample(sample), publicInertialSimulatorState(state), publicInertialError(err)
}
func (simulator *IMUSimulator) State() (InertialSimulatorState, error) {
	if simulator == nil || simulator.handle == nil {
		return InertialSimulatorState{}, ErrClosed
	}
	value, err := simulator.handle.State()
	return publicInertialSimulatorState(value), publicInertialError(err)
}
func (simulator *IMUSimulator) Close() error {
	if simulator == nil || simulator.handle == nil {
		return nil
	}
	return publicInertialError(simulator.handle.Close())
}

func InertialIMUPreset(grade FusionIMUGrade) (FusionIMUSpec, error) {
	value, err := native.InertialIMUPreset(uint32(grade))
	if err != nil {
		return FusionIMUSpec{}, publicInertialError(err)
	}
	return FusionIMUSpec{AccelVRWMPerSqrtS: value.AccelVRW, GyroARRadPerSqrtS: value.GyroARW, AccelBiasInstabilityMPerS2: value.AccelBiasInstability, GyroBiasInstabilityRadPerS: value.GyroBiasInstability, AccelBiasTauS: value.AccelBiasTau, GyroBiasTauS: value.GyroBiasTau, HasAccelScaleInstability: value.HasAccelScale, AccelScaleInstabilityPPM: value.AccelScalePPM, HasGyroScaleInstability: value.HasGyroScale, GyroScaleInstabilityPPM: value.GyroScalePPM}, nil
}
func ValidateInertialIMUSpec(spec FusionIMUSpec) error {
	return publicInertialError(native.InertialIMUSpecValidate(nativeFusionIMUSpec(spec)))
}
func ValidateInertialIMUModel(model InertialIMUModel) error {
	return publicInertialError(native.InertialIMUModelValidate(nativeInertialModel(model)))
}
func ValidateInertialRateRandomWalk(accel, gyro float64) error {
	return publicInertialError(native.InertialRateRandomWalkValidate(accel, gyro))
}
func InertialIMUBiasStatistics(spec FusionIMUSpec, deltaSeconds float64) ([4]float64, error) {
	value, err := native.InertialIMUSpecBiasStatistics(nativeFusionIMUSpec(spec), deltaSeconds)
	return value, publicInertialError(err)
}
func InertialModelConstants() (InertialConstants, error) {
	value, err := native.InertialConstants()
	if err != nil {
		return InertialConstants{}, publicInertialError(err)
	}
	return InertialConstants{DefaultIMUSimulatorSeed: value.DefaultSeed, NormalGravityEquatorMPerS2: value.GravityEquator, NormalGravityPoleMPerS2: value.GravityPole, SomiglianaK: value.SomiglianaK}, nil
}
func InertialCalibrationFromScalePPM(accel, gyro [3]float64) (InertialIMUModel, error) {
	value, err := native.InertialCalibrationFromScalePPM(accel, gyro)
	if err != nil {
		return InertialIMUModel{}, publicInertialError(err)
	}
	return InertialIMUModel{AccelBiasMPerS2: value.AccelBias, GyroBiasRadPerS: value.GyroBias, AccelScaleMisalignment: value.AccelScaleMisalignment, GyroScaleMisalignment: value.GyroScaleMisalignment}, nil
}
func InertialCorrectSample(sample FusionIMUSample, previousEpoch float64, model InertialIMUModel) (InertialIncrement, error) {
	value, err := native.InertialCorrectSample(nativeInertialSample(sample), previousEpoch, nativeInertialModel(model))
	return publicInertialIncrement(value), publicInertialError(err)
}
func InertialMechanizeECEF(state InertialNavState, increment InertialIncrement) (InertialNavState, error) {
	value, err := native.InertialMechanizeECEF(nativeInertialState(state), nativeInertialIncrement(increment), 0)
	return publicInertialState(value), publicInertialError(err)
}
func InertialNormalGravity(latitudeRad, heightM float64) (float64, error) {
	value, err := native.InertialNormalGravity(latitudeRad, heightM)
	return value, publicInertialError(err)
}
func InertialGravityECEF(positionECEFM [3]float64) ([3]float64, error) {
	value, err := native.InertialGravityECEF(positionECEFM)
	return value, publicInertialError(err)
}
func InertialTrueIncrementBetween(start, end InertialNavState) (InertialIncrement, error) {
	value, err := native.InertialTrueIncrementBetween(nativeInertialState(start), nativeInertialState(end))
	return publicInertialIncrement(value), publicInertialError(err)
}
func InertialQuaternionFromDCM(dcm [9]float64) (InertialQuaternion, error) {
	value, err := native.InertialQuaternionFromDCM(dcm)
	return InertialQuaternion{W: value.W, X: value.X, Y: value.Y, Z: value.Z}, publicInertialError(err)
}
func InertialDCMFromQuaternion(value InertialQuaternion) ([9]float64, error) {
	dcm, err := native.InertialDCMFromQuaternion(native.NativeInertialQuaternion{W: value.W, X: value.X, Y: value.Y, Z: value.Z})
	return dcm, publicInertialError(err)
}
func InertialReorthonormalizeDCM(dcm [9]float64) ([9]float64, error) {
	value, err := native.InertialReorthonormalizeDCM(dcm)
	return value, publicInertialError(err)
}
func InertialYawPitchRoll(dcm [9]float64) ([3]float64, error) {
	value, err := native.InertialYawPitchRoll(dcm)
	return value, publicInertialError(err)
}
func InertialRodrigues(deltaThetaRad [3]float64) ([9]float64, error) {
	value, err := native.InertialRodrigues(deltaThetaRad)
	return value, publicInertialError(err)
}
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
