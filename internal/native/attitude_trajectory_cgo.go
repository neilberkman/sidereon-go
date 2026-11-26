//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type OrbitAttitudeTrajectory struct {
	handle *positioningHandle
}

func (trajectory *OrbitAttitudeTrajectory) orbitHandle() *positioningHandle {
	if trajectory == nil {
		return nil
	}
	return trajectory.handle
}

func (trajectory *OrbitAttitudeTrajectory) Close() error {
	return trajectory.orbitHandle().close()
}

type DenseOutput struct {
	handle *positioningHandle
}

func (output *DenseOutput) orbitHandle() *positioningHandle {
	if output == nil {
		return nil
	}
	return output.handle
}

func (output *DenseOutput) Close() error {
	return output.orbitHandle().close()
}

func newAttitudeTrajectory(pointer *C.SidereonOrbitAttitudeTrajectory, err error) (*OrbitAttitudeTrajectory, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_orbit_attitude_trajectory_free((*C.SidereonOrbitAttitudeTrajectory)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &OrbitAttitudeTrajectory{handle: handle}, nil
}

func earthFixedAttitudeTrajectory(source nativeHandleInput, satellite string, orientation *IERS2010EarthOrientationProvider, construct func(unsafe.Pointer, *C.char, *C.SidereonIers2010EarthOrientationProvider, **C.SidereonOrbitAttitudeTrajectory) C.enum_SidereonStatus) (*OrbitAttitudeTrajectory, error) {
	var value *C.SidereonOrbitAttitudeTrajectory
	satelliteToken, err := copyNativeCString(satellite, "satellite")
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(satelliteToken)

	inputs := []nativeHandleInput{source}
	if orientation != nil {
		inputs = append(inputs, nativeHandleInput{positioning: orientation.orbitHandle()})
	}
	err = withNativeHandles(inputs, func(pointers []unsafe.Pointer) error {
		var provider *C.SidereonIers2010EarthOrientationProvider
		if orientation != nil {
			provider = (*C.SidereonIers2010EarthOrientationProvider)(pointers[1])
		}
		return callStatus(func() uint32 {
			return uint32(construct(pointers[0], (*C.char)(satelliteToken), provider, &value))
		})
	})
	return newAttitudeTrajectory(value, err)
}

func OrbitAttitudeTrajectoryFromSP3(product *SP3, satellite string, orientation *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	return earthFixedAttitudeTrajectory(nativeHandleInput{positioning: orbitSP3Handle(product)}, satellite, orientation, func(pointer unsafe.Pointer, satellite *C.char, orientation *C.SidereonIers2010EarthOrientationProvider, out **C.SidereonOrbitAttitudeTrajectory) C.enum_SidereonStatus {
		return C.sidereon_orbit_attitude_trajectory_from_sp3((*C.SidereonSp3)(pointer), satellite, orientation, out)
	})
}

func OrbitAttitudeTrajectoryFromBroadcast(ephemeris *BroadcastEphemeris, satellite string, orientation *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	if ephemeris == nil || ephemeris.resource == nil {
		return nil, ErrClosed
	}
	trajectory, err := earthFixedAttitudeTrajectory(nativeHandleInput{resource: ephemeris.resource}, satellite, orientation, func(pointer unsafe.Pointer, satellite *C.char, orientation *C.SidereonIers2010EarthOrientationProvider, out **C.SidereonOrbitAttitudeTrajectory) C.enum_SidereonStatus {
		return C.sidereon_orbit_attitude_trajectory_from_broadcast((*C.SidereonBroadcastEphemeris)(pointer), satellite, orientation, out)
	})
	runtime.KeepAlive(ephemeris)
	return trajectory, err
}

func OrbitAttitudeTrajectoryFromOrbitSource(source *OrbitSource, satellite string, orientation *IERS2010EarthOrientationProvider) (*OrbitAttitudeTrajectory, error) {
	return earthFixedAttitudeTrajectory(nativeHandleInput{positioning: source.orbitHandle()}, satellite, orientation, func(pointer unsafe.Pointer, satellite *C.char, orientation *C.SidereonIers2010EarthOrientationProvider, out **C.SidereonOrbitAttitudeTrajectory) C.enum_SidereonStatus {
		return C.sidereon_orbit_attitude_trajectory_from_orbit_source((*C.SidereonOrbitSource)(pointer), satellite, orientation, out)
	})
}

func OrbitAttitudeTrajectoryFromBrouwerLyddane(source *BrouwerLyddanePropagator) (*OrbitAttitudeTrajectory, error) {
	var value *C.SidereonOrbitAttitudeTrajectory
	err := source.orbitHandle().with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_attitude_trajectory_from_brouwer_lyddane((*C.SidereonBrouwerLyddanePropagator)(pointer), &value))
		})
	})
	return newAttitudeTrajectory(value, err)
}

func OrbitAttitudeTrajectoryFromEcksteinHechler(source *EcksteinHechlerPropagator) (*OrbitAttitudeTrajectory, error) {
	var value *C.SidereonOrbitAttitudeTrajectory
	err := source.orbitHandle().with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_attitude_trajectory_from_eckstein_hechler((*C.SidereonEcksteinHechlerPropagator)(pointer), &value))
		})
	})
	return newAttitudeTrajectory(value, err)
}

func OrbitAttitudeTrajectoryFromDenseOutput(source *DenseOutput) (*OrbitAttitudeTrajectory, error) {
	var value *C.SidereonOrbitAttitudeTrajectory
	err := source.orbitHandle().with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_attitude_trajectory_from_dense_output((*C.SidereonDenseOutput)(pointer), &value))
		})
	})
	return newAttitudeTrajectory(value, err)
}

func (trajectory *OrbitAttitudeTrajectory) WithDEEphemeris(ephemeris *DEEphemeris) error {
	inputs := []nativeHandleInput{{positioning: trajectory.orbitHandle(), exclusive: true}}
	if ephemeris != nil {
		inputs = append(inputs, nativeHandleInput{positioning: ephemeris.orbitHandle()})
	}
	return withNativeHandles(inputs, func(pointers []unsafe.Pointer) error {
		var de *C.SidereonDeEphemeris
		if ephemeris != nil {
			de = (*C.SidereonDeEphemeris)(pointers[1])
		}
		return callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_attitude_trajectory_with_de_ephemeris((*C.SidereonOrbitAttitudeTrajectory)(pointers[0]), de))
		})
	})
}

func (trajectory *OrbitAttitudeTrajectory) WithVelocityHalfStepS(halfStepS float64) error {
	return trajectory.orbitHandle().withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_attitude_trajectory_with_velocity_half_step_s((*C.SidereonOrbitAttitudeTrajectory)(pointer), C.double(halfStepS)))
		})
	})
}

func (trajectory *OrbitAttitudeTrajectory) Sample(tdbSeconds float64) (AttitudeSample, error) {
	var value C.SidereonAttitudeSample
	err := trajectory.orbitHandle().with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_attitude_trajectory_sample((*C.SidereonOrbitAttitudeTrajectory)(pointer), C.double(tdbSeconds), &value))
		})
	})
	return AttitudeSample{
		PositionM:    cDouble3FromC(value.position_m),
		VelocityMS:   cDouble3FromC(value.velocity_m_s),
		SunPositionM: cDouble3FromC(value.sun_position_m),
	}, err
}

func GnssAttitude(law *GNSSAttitudeLaw, trajectory *OrbitAttitudeTrajectory, tdbSeconds float64) (GNSSAttitude, error) {
	var value C.SidereonGnssAttitude
	err := withPositioningPair(law.orbitHandle(), trajectory.orbitHandle(), func(law, trajectory unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_gnss_attitude((*C.SidereonGnssAttitudeLaw)(law), (*C.SidereonOrbitAttitudeTrajectory)(trajectory), C.double(tdbSeconds), &value))
		})
	})
	return GNSSAttitude{
		Mode: AttitudeMode(value.mode), YawRad: float64(value.yaw_rad), NominalYawRad: float64(value.nominal_yaw_rad),
		BetaRad: float64(value.beta_rad), OrbitAngleRad: float64(value.orbit_angle_rad),
		BodyX: cDouble3FromC(value.body_x), BodyY: cDouble3FromC(value.body_y), BodyZ: cDouble3FromC(value.body_z),
		HasDYB: bool(value.has_dyb), DYB: orbitDYBAxesFromC(value.dyb),
	}, err
}

func cDouble3FromC(value [3]C.double) [3]float64 {
	return [3]float64{float64(value[0]), float64(value[1]), float64(value[2])}
}
