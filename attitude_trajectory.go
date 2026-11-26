package sidereon

import "sidereon.dev/go/v4/internal/native"

// OrbitAttitudeTrajectory owns an orbit and Sun source for GNSS attitude evaluation.
// Its methods may run concurrently with Close. It must not be copied after first use.
type OrbitAttitudeTrajectory struct {
	_      noCopy
	handle *native.OrbitAttitudeTrajectory
}

func (trajectory *OrbitAttitudeTrajectory) nativeHandle() *native.OrbitAttitudeTrajectory {
	if trajectory == nil {
		return nil
	}
	return trajectory.handle
}

// Close releases the trajectory. It is safe to call more than once.
func (trajectory *OrbitAttitudeTrajectory) Close() error {
	return publicError(trajectory.nativeHandle().Close())
}

// Sample returns the GCRF satellite position and geocentric Sun position in
// metres and satellite velocity in metres per second. The epoch is in TDB
// seconds past J2000.
func (trajectory *OrbitAttitudeTrajectory) Sample(tdbSeconds float64) (AttitudeSample, error) {
	value, err := trajectory.nativeHandle().Sample(tdbSeconds)
	return AttitudeSample(value), publicError(err)
}

// WithDEEphemeris sets the Sun ephemeris. Pass nil to restore the analytic
// Sun model. The trajectory retains the data after the ephemeris is closed.
func (trajectory *OrbitAttitudeTrajectory) WithDEEphemeris(ephemeris *DEEphemeris) error {
	return publicError(trajectory.nativeHandle().WithDEEphemeris(nativeDEEphemerisHandle(ephemeris)))
}

// WithVelocityHalfStepS sets the Earth-fixed velocity difference half step
// in seconds of the source time scale. On refusal, it returns an OrbitScoreError
// and retains the previous half step.
func (trajectory *OrbitAttitudeTrajectory) WithVelocityHalfStepS(halfStepS float64) error {
	return publicError(trajectory.nativeHandle().WithVelocityHalfStepS(halfStepS))
}

// GnssAttitude evaluates a GNSS yaw law at TDB seconds past J2000 and returns
// body and DYB axes in GCRF. Some laws also require coverage around the epoch.
func GnssAttitude(law *GNSSAttitudeLaw, trajectory *OrbitAttitudeTrajectory, tdbSeconds float64) (GNSSAttitude, error) {
	value, err := native.GnssAttitude(nativeGNSSAttitudeLawHandle(law), trajectory.nativeHandle(), tdbSeconds)
	return GNSSAttitude{
		Mode: AttitudeMode(value.Mode), YawRad: value.YawRad, NominalYawRad: value.NominalYawRad,
		BetaRad: value.BetaRad, OrbitAngleRad: value.OrbitAngleRad,
		BodyX: value.BodyX, BodyY: value.BodyY, BodyZ: value.BodyZ,
		HasDYB: value.HasDYB, DYB: publicDYBAxes(value.DYB),
	}, publicError(err)
}

// DenseOutput owns the continuous output of a DP5(4) numerical propagation.
// It must not be copied after first use.
type DenseOutput struct {
	_      noCopy
	handle *native.DenseOutput
}

func (output *DenseOutput) nativeHandle() *native.DenseOutput {
	if output == nil {
		return nil
	}
	return output.handle
}

// Close releases the continuous output. It is safe to call more than once.
func (output *DenseOutput) Close() error {
	return publicError(output.nativeHandle().Close())
}

// OrbitAttitudeTrajectoryFromSP3 selects a satellite from a parsed SP3 product.
// Pass nil for the default TDB Earth-orientation provider. The source frame
// must align with the provider’s ITRF. The trajectory retains its inputs.
func OrbitAttitudeTrajectoryFromSP3(product *SP3, satellite string, orientation *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	value, err := native.OrbitAttitudeTrajectoryFromSP3(nativeSP3Handle(product), satellite, nativeIERS2010EarthOrientationProviderHandle(orientation))
	if err != nil {
		return nil, publicError(err)
	}
	return &OrbitAttitudeTrajectory{handle: value}, nil
}

// OrbitAttitudeTrajectoryFromBroadcast selects a satellite from a broadcast navigation set.
// Pass nil for the default TDB Earth-orientation provider. The source frame
// must align with the provider’s ITRF. The trajectory retains its inputs.
func OrbitAttitudeTrajectoryFromBroadcast(ephemeris *BroadcastEphemeris, satellite string, orientation *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	value, err := native.OrbitAttitudeTrajectoryFromBroadcast(nativeBroadcastEphemerisHandle(ephemeris), satellite, nativeIERS2010EarthOrientationProviderHandle(orientation))
	if err != nil {
		return nil, publicError(err)
	}
	return &OrbitAttitudeTrajectory{handle: value}, nil
}

// OrbitAttitudeTrajectoryFromOrbitSource selects a satellite from a predicted SP3, broadcast orbit or sampled orbit scoring source.
// Pass nil for the default TDB Earth-orientation provider. The source frame
// must align with the provider’s ITRF. The trajectory retains its inputs.
func OrbitAttitudeTrajectoryFromOrbitSource(source OrbitSource, satellite string, orientation *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	value, err := native.OrbitAttitudeTrajectoryFromOrbitSource(nativeOrbitSourceHandle(source), satellite, nativeIERS2010EarthOrientationProviderHandle(orientation))
	if err != nil {
		return nil, publicError(err)
	}
	return &OrbitAttitudeTrajectory{handle: value}, nil
}

// OrbitAttitudeTrajectoryFromEcksteinHechler uses an Eckstein–Hechler propagator defined in GCRF.
// The trajectory retains its input.
func OrbitAttitudeTrajectoryFromEcksteinHechler(propagator *EcksteinHechlerPropagator) (*OrbitAttitudeTrajectory, error) {
	value, err := native.OrbitAttitudeTrajectoryFromEcksteinHechler(nativeEcksteinHechlerPropagatorHandle(propagator))
	if err != nil {
		return nil, publicError(err)
	}
	return &OrbitAttitudeTrajectory{handle: value}, nil
}

// OrbitAttitudeTrajectoryFromBrouwerLyddane uses a Brouwer–Lyddane propagator defined in GCRF.
// The trajectory retains its input.
func OrbitAttitudeTrajectoryFromBrouwerLyddane(propagator *BrouwerLyddanePropagator) (*OrbitAttitudeTrajectory, error) {
	value, err := native.OrbitAttitudeTrajectoryFromBrouwerLyddane(nativeBrouwerLyddanePropagatorHandle(propagator))
	if err != nil {
		return nil, publicError(err)
	}
	return &OrbitAttitudeTrajectory{handle: value}, nil
}

// OrbitAttitudeTrajectoryFromDenseOutput uses continuous numerical propagation output defined in GCRF.
// The trajectory retains its input.
func OrbitAttitudeTrajectoryFromDenseOutput(output *DenseOutput) (*OrbitAttitudeTrajectory, error) {
	value, err := native.OrbitAttitudeTrajectoryFromDenseOutput(output.nativeHandle())
	if err != nil {
		return nil, publicError(err)
	}
	return &OrbitAttitudeTrajectory{handle: value}, nil
}
