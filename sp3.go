package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

type noCopy struct{}

// Lock acquires the synchronization lock protecting the SP3 selection resource.
func (*noCopy) Lock() {}

// Unlock releases the synchronization lock protecting the SP3 selection resource.
func (*noCopy) Unlock() {}

// SP3State is one detached satellite state from an SP3 epoch. Positions are in
// ECEF metres, velocities are in ECEF metres per second, and clock values are
// seconds or seconds per second as indicated by their field names.
type SP3State struct {
	// PositionM is the position m in metres.
	PositionM [3]float64
	// HasClock reports whether ClockS is valid.
	HasClock bool
	// ClockS is the clock s in seconds.
	ClockS float64
	// HasVelocity reports whether VelocityMPerS is valid.
	HasVelocity bool
	// VelocityMPerS is the velocity m per s in metres per second.
	VelocityMPerS [3]float64
	// HasClockRate reports whether ClockRateSPerS is valid.
	HasClockRate bool
	// ClockRateSPerS is the clock rate s per s in seconds per second.
	ClockRateSPerS float64
	// ClockEvent reports whether the sample is marked as a clock event.
	ClockEvent bool
	// ClockPredicted reports whether the clock value was predicted rather than observed.
	ClockPredicted bool
	// Maneuver reports whether the sample is marked as a maneuver.
	Maneuver bool
	// OrbitPredicted reports whether the orbit position was predicted rather than observed.
	OrbitPredicted bool
}

// SP3PredictionSummary contains the product-wide observed/predicted boundary.
type SP3PredictionSummary struct {
	// EpochCount is the number of epochs in the parsed SP3 product.
	EpochCount int
	// ObservedThroughPresent reports whether ObservedThroughJ2000S is valid.
	ObservedThroughPresent bool
	// ObservedThroughJ2000S is the last observed epoch in seconds from J2000.
	ObservedThroughJ2000S float64
}

// SP3InterpolationOptions configures the coverage-gap threshold factor for SP3 interpolation.
type SP3InterpolationOptions struct {
	// GapThresholdFactor is the multiple of nominal node spacing above which
	// consecutive nodes are treated as a coverage gap. A non-positive value
	// (<= 0.0) selects the engine default of 1.5. Values in (0.0, 1.0] or non-finite
	// values are rejected by the engine with an invalid argument error.
	GapThresholdFactor float64
}

// SP3LoadOptions configures SP3 product loading options.
type SP3LoadOptions = SP3InterpolationOptions

// SP3ContinuityOptions configures SP3 product continuity check options.
type SP3ContinuityOptions = SP3InterpolationOptions

// SP3Option configures optional SP3 loading, continuity, and ephemeris policy.
type SP3Option func(*SP3InterpolationOptions)

// WithGapThresholdFactor returns an SP3Option configuring the interpolation gap threshold factor.
func WithGapThresholdFactor(factor float64) SP3Option {
	return func(o *SP3InterpolationOptions) {
		o.GapThresholdFactor = factor
	}
}

// NewSP3InterpolationOptions builds an SP3InterpolationOptions struct from functional options.
func NewSP3InterpolationOptions(opts ...SP3Option) SP3InterpolationOptions {
	var options SP3InterpolationOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func resolveSP3Options(opts []SP3Option) SP3InterpolationOptions {
	return NewSP3InterpolationOptions(opts...)
}

// SP3 owns a parsed C SP3 handle. It must not be copied after first use. Read
// methods may be called concurrently with one another or with Close.
type SP3 struct {
	_      noCopy
	handle *native.SP3
}

// LoadSP3 parses data through the C ABI. The input is copied for the duration
// of the call and is not retained by the returned handle.
func LoadSP3(data []byte, opts ...SP3Option) (*SP3, error) {
	if len(opts) == 0 {
		handle, err := native.LoadSP3(data)
		if err != nil {
			return nil, publicError(err)
		}
		if handle == nil {
			return nil, errNilNativeHandle
		}
		return &SP3{handle: handle}, nil
	}
	return LoadSP3WithOptions(data, resolveSP3Options(opts))
}

// LoadSP3WithOptions parses data through the C ABI using explicit interpolation options.
func LoadSP3WithOptions(data []byte, options SP3InterpolationOptions) (*SP3, error) {
	handle, err := native.LoadSP3WithGapThresholdFactor(data, options.GapThresholdFactor)
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &SP3{handle: handle}, nil
}

// GapThresholdFactor returns the SP3 interpolation gap threshold factor carried by this product.
func (s *SP3) GapThresholdFactor() (float64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	v, err := s.handle.GapThresholdFactor()
	return v, publicError(err)
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

// Epochs returns detached epoch times in the product time scale, expressed as
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

// Satellites returns detached satellite identifiers in the product order.
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

// State returns one detached satellite state at a parsed epoch index.
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

// PredictionSummary returns detached observed/predicted boundary metadata.
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
