//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

// These value types mirror the cgo-backed fusion surface so public wrappers
// remain type-checkable when the native backend is unavailable.
type NativeFusionIMUSpec struct {
	AccelVRW, GyroARW, AccelBiasInstability, GyroBiasInstability, AccelBiasTau, GyroBiasTau float64
	HasAccelScale, HasGyroScale                                                             bool
	AccelScalePPM, GyroScalePPM                                                             float64
}

type NativeFusionIMUSample struct {
	Epoch                                                 float64
	Kind                                                  uint32
	SpecificForce, AngularRate, DeltaVelocity, DeltaTheta [3]float64
	DTS                                                   float64
}

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
type NativeInertialSimulatorState struct{ AccelBias, GyroBias, AccelRateRandomWalk, GyroRateRandomWalk [3]float64 }
type NativeInertialQuaternion struct{ W, X, Y, Z float64 }
type NativeInertialConstants struct {
	DefaultSeed                              uint64
	GravityEquator, GravityPole, SomiglianaK float64
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
	return "sidereon: inertial operation unavailable"
}
func (err *NativeInertialStatusError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

type NativeInertialMechanizer struct{}
type NativeIMUSimulator struct{}

func InertialMechanizerNew(NativeInertialNavState, *NativeInertialIMUModel, uint32) (*NativeInertialMechanizer, error) {
	return nil, unavailable()
}
func InertialMechanizerNewWithConfig(NativeInertialNavState, uint32) (*NativeInertialMechanizer, error) {
	return nil, unavailable()
}
func (*NativeInertialMechanizer) Propagate(NativeFusionIMUSample) (NativeInertialNavState, error) {
	return NativeInertialNavState{}, unavailable()
}
func (*NativeInertialMechanizer) State() (NativeInertialNavState, error) {
	return NativeInertialNavState{}, unavailable()
}
func (*NativeInertialMechanizer) Close() error { return nil }
func InertialMechanizeECEF(NativeInertialNavState, NativeInertialIncrement, uint32) (NativeInertialNavState, error) {
	return NativeInertialNavState{}, unavailable()
}
func InertialNormalGravity(float64, float64) (float64, error) { return 0, unavailable() }
func InertialGravityECEF([3]float64) ([3]float64, error)      { return [3]float64{}, unavailable() }
func InertialTrueIncrementBetween(NativeInertialNavState, NativeInertialNavState) (NativeInertialIncrement, error) {
	return NativeInertialIncrement{}, unavailable()
}
func InertialIMUPreset(uint32) (NativeFusionIMUSpec, error) {
	return NativeFusionIMUSpec{}, unavailable()
}
func InertialIMUSpecValidate(NativeFusionIMUSpec) error { return unavailable() }
func InertialIMUSpecBiasStatistics(NativeFusionIMUSpec, float64) ([4]float64, error) {
	return [4]float64{}, unavailable()
}
func InertialIMUModelValidate(NativeInertialIMUModel) error { return unavailable() }
func InertialRateRandomWalkValidate(float64, float64) error { return unavailable() }
func InertialConstants() (NativeInertialConstants, error) {
	return NativeInertialConstants{}, unavailable()
}
func InertialCalibrationFromScalePPM([3]float64, [3]float64) (NativeInertialIMUModel, error) {
	return NativeInertialIMUModel{}, unavailable()
}
func InertialQuaternionFromDCM([9]float64) (NativeInertialQuaternion, error) {
	return NativeInertialQuaternion{}, unavailable()
}
func InertialDCMFromQuaternion(NativeInertialQuaternion) ([9]float64, error) {
	return [9]float64{}, unavailable()
}
func InertialReorthonormalizeDCM([9]float64) ([9]float64, error) { return [9]float64{}, unavailable() }
func InertialYawPitchRoll([9]float64) ([3]float64, error)        { return [3]float64{}, unavailable() }
func InertialRodrigues([3]float64) ([9]float64, error)           { return [9]float64{}, unavailable() }
func InertialCorrectSample(NativeFusionIMUSample, float64, NativeInertialIMUModel) (NativeInertialIncrement, error) {
	return NativeInertialIncrement{}, unavailable()
}
func InertialIMUSimulatorNew(NativeFusionIMUSpec, NativeInertialSimulationOptions) (*NativeIMUSimulator, error) {
	return nil, unavailable()
}
func (*NativeIMUSimulator) SampleIncrement(NativeInertialIncrement) (NativeFusionIMUSample, NativeInertialSimulatorState, error) {
	return NativeFusionIMUSample{}, NativeInertialSimulatorState{}, unavailable()
}
func (*NativeIMUSimulator) State() (NativeInertialSimulatorState, error) {
	return NativeInertialSimulatorState{}, unavailable()
}
func (*NativeIMUSimulator) Close() error { return nil }
func InertialSimulateIncrements([]NativeInertialIncrement, NativeFusionIMUSpec, NativeInertialSimulationOptions) ([]NativeFusionIMUSample, []NativeInertialSimulatorState, error) {
	return nil, nil, unavailable()
}
func InertialSimulateTrajectory([]NativeInertialNavState, NativeFusionIMUSpec, NativeInertialSimulationOptions) ([]NativeFusionIMUSample, []NativeInertialSimulatorState, error) {
	return nil, nil, unavailable()
}
