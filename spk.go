package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SPK owns a parsed JPL/NAIF binary ephemeris kernel.
type SPK struct {
	_      noCopy
	handle *native.SPK
}

// LoadSPK parses detached bytes as a JPL/NAIF binary ephemeris kernel.
func LoadSPK(data []byte) (*SPK, error) {
	handle, err := native.LoadSPK(append([]byte(nil), data...))
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &SPK{handle: handle}, nil
}

// Close releases the SPK kernel and is idempotent.
func (s *SPK) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// SPKState contains a target-center state in kilometres and kilometres per second.
type SPKState struct {
	// Target and Center are NAIF body identifiers.
	Target      int32
	Center      int32
	PositionKm  [3]float64
	HasVelocity bool
	// HasVelocityKmPerS is an explicit unit-bearing alias of HasVelocity.
	HasVelocityKmPerS bool
	VelocityKmPerS    [3]float64
	Frame             int32
}

// State evaluates one target-center path at ET/TDB seconds past J2000.
func (s *SPK) State(target, center int32, etSecondsTDB float64) (SPKState, error) {
	if s == nil || s.handle == nil {
		return SPKState{}, ErrClosed
	}
	value, err := s.handle.State(target, center, etSecondsTDB)
	return SPKState{Target: value.Target, Center: value.Center, PositionKm: value.PositionKm, HasVelocity: value.HasVelocity, HasVelocityKmPerS: value.HasVelocityKmPerS, VelocityKmPerS: value.VelocityKmPerS, Frame: value.Frame}, publicError(err)
}

// PositionKm and VelocityKmPerS are target-center state vectors in km and km/s.
// HasVelocity and HasVelocityKmPerS report optional velocity presence.
// Frame is the native reference-frame identifier.
