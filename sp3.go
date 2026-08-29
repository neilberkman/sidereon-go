package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

type noCopy struct{}

// Lock acquires the synchronization lock protecting the SP3 selection resource.
func (*noCopy) Lock() {}

// Unlock releases the synchronization lock protecting the SP3 selection resource.
func (*noCopy) Unlock() {}

// SP3State is one copied satellite state from an SP3 epoch. Positions are in
// ECEF meters, velocities are in ECEF meters per second, and clock values are
// seconds or seconds per second as indicated by their field names.
type SP3State struct {
	// PositionM is the position m in metres.
	PositionM [3]float64
	// HasClock reports whether the has clock field is present.
	HasClock bool
	// ClockS is the clock s in seconds.
	ClockS float64
	// HasVelocity reports whether the has velocity field is present.
	HasVelocity bool
	// VelocityMPerS is the velocity m per s in metres per second.
	VelocityMPerS [3]float64
	// HasClockRate reports whether the has clock rate field is present.
	HasClockRate bool
	// ClockRateSPerS is the clock rate s per s in seconds per second.
	ClockRateSPerS float64
	// ClockEvent is the clock event value for SP3State.
	ClockEvent bool
	// ClockPredicted is the clock predicted value for SP3State.
	ClockPredicted bool
	// Maneuver is the maneuver value for SP3State.
	Maneuver bool
	// OrbitPredicted is the orbit predicted value for SP3State.
	OrbitPredicted bool
}

// SP3PredictionSummary contains the product-wide observed/predicted boundary.
type SP3PredictionSummary struct {
	// EpochCount identifies or counts this record.
	EpochCount int
	// ObservedThroughPresent reports whether the observed through present field is present.
	ObservedThroughPresent bool
	// ObservedThroughJ2000S is the observed through j2000 s in seconds.
	ObservedThroughJ2000S float64
}

// SP3 owns a parsed C SP3 handle. It must not be copied after first use. Read
// methods may be called concurrently with one another or with Close.
type SP3 struct {
	_      noCopy
	handle *native.SP3
}

// LoadSP3 parses data through the C ABI. The input is copied for the duration
// of the call and is not retained by the returned handle.
func LoadSP3(data []byte) (*SP3, error) {
	handle, err := native.LoadSP3(data)
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &SP3{handle: handle}, nil
}

// Close releases the native SP3 handle. It is idempotent and safe to call
// concurrently with read-only methods.
func (s *SP3) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// EpochCount returns the number of parsed epoch nodes.
func (s *SP3) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	count, err := s.handle.EpochCount()
	return count, publicError(err)
}

// Epochs returns copied epoch times in the product time scale, expressed as
// seconds since J2000 by the C library.
func (s *SP3) Epochs() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	epochs, err := s.handle.Epochs()
	if err != nil {
		return nil, publicError(err)
	}
	return append([]float64(nil), epochs...), nil
}

// Satellites returns copied satellite identifiers in the product order.
func (s *SP3) Satellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	satellites, err := s.handle.Satellites()
	if err != nil {
		return nil, publicError(err)
	}
	return append([]string(nil), satellites...), nil
}

// State returns one copied satellite state at a parsed epoch index.
func (s *SP3) State(satelliteID string, epochIndex int) (SP3State, error) {
	if s == nil || s.handle == nil {
		return SP3State{}, ErrClosed
	}
	state, err := s.handle.State(satelliteID, epochIndex)
	if err != nil {
		return SP3State{}, publicError(err)
	}
	return SP3State{
		PositionM:      state.PositionM,
		HasClock:       state.HasClock,
		ClockS:         state.ClockS,
		HasVelocity:    state.HasVelocity,
		VelocityMPerS:  state.VelocityMPerS,
		HasClockRate:   state.HasClockRate,
		ClockRateSPerS: state.ClockRateSPerS,
		ClockEvent:     state.ClockEvent,
		ClockPredicted: state.ClockPredicted,
		Maneuver:       state.Maneuver,
		OrbitPredicted: state.OrbitPredicted,
	}, nil
}

// PredictionSummary returns copied observed/predicted boundary metadata.
func (s *SP3) PredictionSummary() (SP3PredictionSummary, error) {
	if s == nil || s.handle == nil {
		return SP3PredictionSummary{}, ErrClosed
	}
	summary, err := s.handle.PredictionSummary()
	if err != nil {
		return SP3PredictionSummary{}, publicError(err)
	}
	return SP3PredictionSummary{
		EpochCount:             summary.EpochCount,
		ObservedThroughPresent: summary.ObservedThroughPresent,
		ObservedThroughJ2000S:  summary.ObservedThroughJ2000S,
	}, nil
}
