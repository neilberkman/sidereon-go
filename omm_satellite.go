package sidereon

import (
	"errors"
	"runtime"
	"sidereon.dev/go/v3/internal/native"
)

// SGP4Satellite is an independently owned propagation-ready satellite created
// directly by the core OMM-to-SGP4 constructor.
type SGP4Satellite struct {
	_      noCopy
	handle *native.Sgp4Satellite
	epoch  JulianDate
}

// NewSGP4SatelliteFromOMM invokes the core Satellite::from_omm path and returns
// an independently owned propagation handle.
func NewSGP4SatelliteFromOMM(omm *OMM) (*SGP4Satellite, error) {
	if omm == nil || omm.handle == nil {
		return nil, ErrClosed
	}
	handle, err := native.Sgp4SatelliteFromOMM(omm.handle)
	if err != nil {
		return nil, publicError(err)
	}
	epoch, err := handle.Epoch()
	if err != nil {
		return nil, errors.Join(publicError(err), publicError(handle.Close()))
	}
	runtime.KeepAlive(omm)
	return &SGP4Satellite{handle: handle, epoch: JulianDate{Whole: epoch.Whole, Fraction: epoch.Fraction}}, nil
}

// Close releases the initialized SGP4 satellite; repeated calls are safe.
func (s *SGP4Satellite) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Epoch returns the exact split Julian date retained by the initialized SGP4 satellite.
func (s *SGP4Satellite) Epoch() (JulianDate, error) {
	if s == nil || s.handle == nil {
		return JulianDate{}, ErrClosed
	}
	epoch, err := s.handle.Epoch()
	runtime.KeepAlive(s)
	if err != nil {
		return JulianDate{}, publicError(err)
	}
	return JulianDate{Whole: epoch.Whole, Fraction: epoch.Fraction}, nil
}

// EpochJ2000S returns the split epoch as seconds from J2000 for convenience.
func (s *SGP4Satellite) EpochJ2000S() (float64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EpochJ2000S()
	runtime.KeepAlive(s)
	if err != nil {
		return 0, publicError(err)
	}
	return value, nil
}

// PropagateMinutesSinceEpoch returns a detached TEME state at the supplied
// minutes from this OMM's SGP4 element epoch.
func (s *SGP4Satellite) PropagateMinutesSinceEpoch(minutes float64) (TEMEState, error) {
	if s == nil || s.handle == nil {
		return TEMEState{}, ErrClosed
	}
	value, err := s.handle.PropagateMinutes(minutes)
	runtime.KeepAlive(s)
	if err != nil {
		return TEMEState{}, publicError(err)
	}
	value.EpochJ2000S = (s.epoch.Whole-2451545.0)*86400.0 + s.epoch.Fraction*86400.0 + minutes*60.0
	return TEMEState{EpochJ2000S: value.EpochJ2000S, PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS}, nil
}

// PropagateJulianDate returns a detached TEME state at an exact split Julian date.
func (s *SGP4Satellite) PropagateJulianDate(epoch JulianDate) (TEMEState, error) {
	if s == nil || s.handle == nil {
		return TEMEState{}, ErrClosed
	}
	value, err := s.handle.PropagateJD(native.JulianDate{Whole: epoch.Whole, Fraction: epoch.Fraction})
	runtime.KeepAlive(s)
	if err != nil {
		return TEMEState{}, publicError(err)
	}
	value.EpochJ2000S = (epoch.Whole-2451545.0)*86400.0 + epoch.Fraction*86400.0
	return TEMEState{EpochJ2000S: value.EpochJ2000S, PositionKm: value.PositionKm, VelocityKmPerS: value.VelocityKmPerS}, nil
}
