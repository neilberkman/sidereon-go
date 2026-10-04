//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Sgp4Satellite owns a propagation-ready core satellite initialized from OMM.
type Sgp4Satellite struct {
	_      noCopy
	handle *positioningHandle
}

func releaseSgp4Satellite(pointer unsafe.Pointer) {
	C.sidereon_sgp4_satellite_free((*C.SidereonSgp4Satellite)(pointer))
}

// Sgp4SatelliteFromOMM uses the core Satellite::from_omm constructor.
func Sgp4SatelliteFromOMM(omm *OMM) (*Sgp4Satellite, error) {
	if omm == nil || omm.handle == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonSgp4Satellite
	err := omm.handle.with(func(pointer unsafe.Pointer) error {
		return callSGP4Status(func() uint32 { return uint32(C.sidereon_sgp4_satellite_from_omm((*C.SidereonOmm)(pointer), &out)) })
	})
	runtime.KeepAlive(omm)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_sgp4_satellite_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("OMM SGP4 initialization")
	}
	return &Sgp4Satellite{handle: newPositioningHandle(unsafe.Pointer(out), releaseSgp4Satellite)}, nil
}

// PropagateMinutes returns a detached TEME state at minutes since element epoch.
func (s *Sgp4Satellite) PropagateMinutes(minutes float64) (TEMEState, error) {
	if s == nil || s.handle == nil {
		return TEMEState{}, ErrClosed
	}
	var value C.SidereonTemeState
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callSGP4Status(func() uint32 {
			return uint32(C.sidereon_sgp4_satellite_propagate_minutes((*C.SidereonSgp4Satellite)(pointer), C.double(minutes), &value))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return TEMEState{}, err
	}
	return TEMEState{PositionKm: [3]float64{float64(value.position_km[0]), float64(value.position_km[1]), float64(value.position_km[2])}, VelocityKmPerS: [3]float64{float64(value.velocity_km_s[0]), float64(value.velocity_km_s[1]), float64(value.velocity_km_s[2])}}, nil
}

// PropagateJD returns a detached TEME state at a split Julian date.
func (s *Sgp4Satellite) PropagateJD(epoch JulianDate) (TEMEState, error) {
	if s == nil || s.handle == nil {
		return TEMEState{}, ErrClosed
	}
	var value C.SidereonTemeState
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callSGP4Status(func() uint32 {
			return uint32(C.sidereon_sgp4_satellite_propagate_jd((*C.SidereonSgp4Satellite)(pointer), C.double(epoch.Whole), C.double(epoch.Fraction), &value))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return TEMEState{}, err
	}
	return TEMEState{PositionKm: [3]float64{float64(value.position_km[0]), float64(value.position_km[1]), float64(value.position_km[2])}, VelocityKmPerS: [3]float64{float64(value.velocity_km_s[0]), float64(value.velocity_km_s[1]), float64(value.velocity_km_s[2])}}, nil
}

// EpochJ2000S returns the initialized satellite epoch from the core split date.
func (s *Sgp4Satellite) EpochJ2000S() (float64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var value C.double
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callSGP4Status(func() uint32 {
			return uint32(C.sidereon_sgp4_satellite_epoch_j2000_s((*C.SidereonSgp4Satellite)(pointer), &value))
		})
	})
	runtime.KeepAlive(s)
	return float64(value), err
}

// Epoch returns the exact split Julian date retained by the core satellite.
func (s *Sgp4Satellite) Epoch() (JulianDate, error) {
	if s == nil || s.handle == nil {
		return JulianDate{}, ErrClosed
	}
	var whole, fraction C.double
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callSGP4Status(func() uint32 {
			return uint32(C.sidereon_sgp4_satellite_epoch_jd((*C.SidereonSgp4Satellite)(pointer), &whole, &fraction))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return JulianDate{}, err
	}
	return JulianDate{Whole: float64(whole), Fraction: float64(fraction)}, nil
}

// Close releases the initialized satellite; repeated calls are safe.
func (s *Sgp4Satellite) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}
