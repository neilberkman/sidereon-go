package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SPK owns a parsed JPL/NAIF binary ephemeris kernel.
type SPK struct {
	_      noCopy
	handle *native.SPK
}

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

func (s *SPK) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

type SPKState struct {
	Target         int32
	Center         int32
	PositionKm     [3]float64
	HasVelocity    bool
	VelocityKmPerS [3]float64
	Frame          int32
}

// State evaluates one target-center path at ET/TDB seconds past J2000.
func (s *SPK) State(target, center int32, etSecondsTDB float64) (SPKState, error) {
	if s == nil || s.handle == nil {
		return SPKState{}, ErrClosed
	}
	value, err := s.handle.State(target, center, etSecondsTDB)
	return SPKState{Target: value.Target, Center: value.Center, PositionKm: value.PositionKm, HasVelocity: value.HasVelocity, VelocityKmPerS: value.VelocityKmPerS, Frame: value.Frame}, publicError(err)
}
