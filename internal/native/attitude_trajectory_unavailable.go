//go:build !cgo || (cgo && !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64)))

package native

type OrbitAttitudeTrajectory struct{}

func (*OrbitAttitudeTrajectory) Close() error { return nil }

type DenseOutput struct{}

func (*DenseOutput) Close() error { return nil }

func OrbitAttitudeTrajectoryFromSP3(*SP3, string, *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	return nil, ErrUnavailable
}

func OrbitAttitudeTrajectoryFromBroadcast(*BroadcastEphemeris, string, *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	return nil, ErrUnavailable
}

func OrbitAttitudeTrajectoryFromOrbitSource(*OrbitSource, string, *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	return nil, ErrUnavailable
}

func OrbitAttitudeTrajectoryFromEcksteinHechler(*EcksteinHechlerPropagator) (*OrbitAttitudeTrajectory, error) {
	return nil, ErrUnavailable
}

func OrbitAttitudeTrajectoryFromBrouwerLyddane(*BrouwerLyddanePropagator) (*OrbitAttitudeTrajectory, error) {
	return nil, ErrUnavailable
}

func OrbitAttitudeTrajectoryFromDenseOutput(*DenseOutput) (*OrbitAttitudeTrajectory, error) {
	return nil, ErrUnavailable
}

func (*OrbitAttitudeTrajectory) WithDEEphemeris(*DEEphemeris) error { return ErrUnavailable }

func (*OrbitAttitudeTrajectory) WithVelocityHalfStepS(float64) error { return ErrUnavailable }

func (*OrbitAttitudeTrajectory) Sample(float64) (AttitudeSample, error) {
	return AttitudeSample{}, ErrUnavailable
}

func GnssAttitude(*GNSSAttitudeLaw, *OrbitAttitudeTrajectory, float64) (GNSSAttitude, error) {
	return GNSSAttitude{}, ErrUnavailable
}

func PropagateStateWithDenseOutput(NativePropagationConfig, float64, PropagationEnvironment, []*ForceModel) (CartesianState, *DenseOutput, error) {
	return CartesianState{}, nil, ErrUnavailable
}
