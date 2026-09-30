package sidereon

import "sidereon.dev/go/v3/internal/native"

type UT1DegradeReason uint32

const (
	UT1NotDegraded UT1DegradeReason = iota
	UT1DegradedBeforeCoverage
	UT1DegradedAfterCoverage
)

// EphemerisSourceState is a detached exact-query source result.
type EphemerisSourceState struct {
	HasState      bool
	PositionECEFM [3]float64
	ClockS        float64
	HasGroupDelay bool
	GroupDelayS   float64
	Degraded      bool
	DegradeReason UT1DegradeReason
}

// TransmitEpochClock carries source clock availability and degradation status.
type TransmitEpochClock struct {
	HasClock      bool
	ClockS        float64
	Degraded      bool
	DegradeReason UT1DegradeReason
}

// ClockRelativityKind identifies the exact-query clock-relativity outcome.
type ClockRelativityKind uint32

const (
	ClockRelativityNotApplicable ClockRelativityKind = iota
	ClockRelativityTerm
	ClockRelativityUnavailable
)

// ClockRelativity carries the source's clock-relativity outcome.
type ClockRelativity struct {
	Kind  ClockRelativityKind
	TermS float64
}

func publicEphemerisSourceState(value native.EphemerisSourceState) EphemerisSourceState {
	return EphemerisSourceState{
		HasState: value.HasState, PositionECEFM: value.PositionECEFM,
		ClockS: value.ClockS, HasGroupDelay: value.HasGroupDelay,
		GroupDelayS: value.GroupDelayS, Degraded: value.Degraded, DegradeReason: UT1DegradeReason(value.DegradeReason),
	}
}

func publicTransmitEpochClock(value native.TransmitEpochClock) TransmitEpochClock {
	return TransmitEpochClock{HasClock: value.HasClock, ClockS: value.ClockS, Degraded: value.Degraded, DegradeReason: UT1DegradeReason(value.DegradeReason)}
}

func publicClockRelativity(value native.ClockRelativity) ClockRelativity {
	return ClockRelativity{Kind: ClockRelativityKind(value.Kind), TermS: value.TermS}
}

// SourceStateAtEpochQueries evaluates the source at the state epoch while
// selecting its record with the separate exact selection epoch.
func (s *SP3) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	if s == nil || s.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return EphemerisSourceState{}, ErrClosed
	}
	value, err := s.handle.SourceStateAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return publicEphemerisSourceState(value), publicError(err)
}

// TransmitEpochClockAtEpochQueries selects the clock arc at the exact selection epoch.
func (s *SP3) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	if s == nil || s.handle == nil || transmitEpoch == nil || transmitEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return TransmitEpochClock{}, ErrClosed
	}
	value, err := s.handle.TransmitEpochClockAtEpochQueries(transmitEpoch.handle, selectionEpoch.handle, satellite)
	return publicTransmitEpochClock(value), publicError(err)
}

// ClockRelativityAtEpochQuery evaluates clock relativity for an exact query and ECEF state.
func (s *SP3) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, positionECEFM [3]float64) (ClockRelativity, error) {
	if s == nil || s.handle == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	value, err := s.handle.ClockRelativityAtEpochQuery(epoch.handle, satellite, positionECEFM)
	return publicClockRelativity(value), publicError(err)
}

// EphemerisVarianceAtEpochQueries evaluates selected uncertainty at exact epochs.
func (s *SP3) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	if s == nil || s.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EphemerisVarianceAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return value, publicError(err)
}

// SourceStateAtEpochQueries evaluates the interpolant using exact state and selection epochs.
func (i *PreciseEphemerisInterpolant) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	if i == nil || i.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return EphemerisSourceState{}, ErrClosed
	}
	value, err := i.handle.SourceStateAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return publicEphemerisSourceState(value), publicError(err)
}

// TransmitEpochClockAtEpochQueries selects the interpolant clock arc at an exact epoch.
func (i *PreciseEphemerisInterpolant) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	if i == nil || i.handle == nil || transmitEpoch == nil || transmitEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return TransmitEpochClock{}, ErrClosed
	}
	value, err := i.handle.TransmitEpochClockAtEpochQueries(transmitEpoch.handle, selectionEpoch.handle, satellite)
	return publicTransmitEpochClock(value), publicError(err)
}

// ClockRelativityAtEpochQuery evaluates interpolant clock relativity at an exact epoch.
func (i *PreciseEphemerisInterpolant) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, positionECEFM [3]float64) (ClockRelativity, error) {
	if i == nil || i.handle == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	value, err := i.handle.ClockRelativityAtEpochQuery(epoch.handle, satellite, positionECEFM)
	return publicClockRelativity(value), publicError(err)
}

// EphemerisVarianceAtEpochQueries evaluates interpolant uncertainty at exact epochs.
func (i *PreciseEphemerisInterpolant) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	if i == nil || i.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return 0, ErrClosed
	}
	value, err := i.handle.EphemerisVarianceAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return value, publicError(err)
}

// SourceStateAtEpochQueries evaluates the mapped artifact using exact epochs.
func (a *PreciseInterpolantArtifact) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	if a == nil || a.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return EphemerisSourceState{}, ErrClosed
	}
	value, err := a.handle.SourceStateAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return publicEphemerisSourceState(value), publicError(err)
}

// TransmitEpochClockAtEpochQueries selects the artifact clock arc at an exact epoch.
func (a *PreciseInterpolantArtifact) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	if a == nil || a.handle == nil || transmitEpoch == nil || transmitEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return TransmitEpochClock{}, ErrClosed
	}
	value, err := a.handle.TransmitEpochClockAtEpochQueries(transmitEpoch.handle, selectionEpoch.handle, satellite)
	return publicTransmitEpochClock(value), publicError(err)
}

// ClockRelativityAtEpochQuery evaluates artifact clock relativity at an exact epoch.
func (a *PreciseInterpolantArtifact) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, positionECEFM [3]float64) (ClockRelativity, error) {
	if a == nil || a.handle == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	value, err := a.handle.ClockRelativityAtEpochQuery(epoch.handle, satellite, positionECEFM)
	return publicClockRelativity(value), publicError(err)
}

// EphemerisVarianceAtEpochQueries evaluates artifact uncertainty at exact epochs.
func (a *PreciseInterpolantArtifact) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	if a == nil || a.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return 0, ErrClosed
	}
	value, err := a.handle.EphemerisVarianceAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return value, publicError(err)
}

// SourceStateAtEpochQueries evaluates broadcast state at exact epochs.
func (b *BroadcastEphemeris) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	if b == nil || b.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return EphemerisSourceState{}, ErrClosed
	}
	value, err := b.handle.SourceStateAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return publicEphemerisSourceState(value), publicError(err)
}

// TransmitEpochClockAtEpochQueries evaluates the selected broadcast clock arc.
func (b *BroadcastEphemeris) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	if b == nil || b.handle == nil || transmitEpoch == nil || transmitEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return TransmitEpochClock{}, ErrClosed
	}
	value, err := b.handle.TransmitEpochClockAtEpochQueries(transmitEpoch.handle, selectionEpoch.handle, satellite)
	return publicTransmitEpochClock(value), publicError(err)
}

// ClockRelativityAtEpochQuery evaluates the broadcast clock relativity term.
func (b *BroadcastEphemeris) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, positionECEFM [3]float64) (ClockRelativity, error) {
	if b == nil || b.handle == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	value, err := b.handle.ClockRelativityAtEpochQuery(epoch.handle, satellite, positionECEFM)
	return publicClockRelativity(value), publicError(err)
}

// EphemerisVarianceAtEpochQueries evaluates broadcast uncertainty at exact epochs.
func (b *BroadcastEphemeris) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	if b == nil || b.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return 0, ErrClosed
	}
	value, err := b.handle.EphemerisVarianceAtEpochQueries(stateEpoch.handle, selectionEpoch.handle, satellite)
	return value, publicError(err)
}

// SBASSourceStateAtEpochQueries evaluates corrected state using exact state
// and record-selection epochs.
func (s *SBASCorrectionStore) SourceStateAtEpochQueries(broadcast *BroadcastEphemeris, geo string, mode SBASSolveMode, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery) (EphemerisSourceState, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return EphemerisSourceState{}, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return EphemerisSourceState{}, err
	}
	value, err := s.handle.SourceStateAtEpochQueries(broadcast.handle, geo, uint32(mode), satellite, stateEpoch.handle, selectionEpoch.handle)
	return publicEphemerisSourceState(value), publicError(err)
}

// TransmitEpochClockAtEpochQueries evaluates the SBAS-corrected clock arc.
func (s *SBASCorrectionStore) TransmitEpochClockAtEpochQueries(broadcast *BroadcastEphemeris, geo string, mode SBASSolveMode, satellite string, transmitEpoch, selectionEpoch *ExactEpochQuery) (TransmitEpochClock, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || transmitEpoch == nil || transmitEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return TransmitEpochClock{}, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return TransmitEpochClock{}, err
	}
	value, err := s.handle.TransmitEpochClockAtEpochQueries(broadcast.handle, geo, uint32(mode), satellite, transmitEpoch.handle, selectionEpoch.handle)
	return publicTransmitEpochClock(value), publicError(err)
}

// ClockRelativityAtEpochQuery evaluates the SBAS-corrected clock relativity term.
func (s *SBASCorrectionStore) ClockRelativityAtEpochQuery(broadcast *BroadcastEphemeris, geo string, mode SBASSolveMode, satellite string, epoch *ExactEpochQuery, positionECEFM [3]float64) (ClockRelativity, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return ClockRelativity{}, err
	}
	value, err := s.handle.ClockRelativityAtEpochQuery(broadcast.handle, geo, uint32(mode), satellite, epoch.handle, positionECEFM)
	return publicClockRelativity(value), publicError(err)
}

// EphemerisVarianceAtEpochQueries evaluates SBAS-selected uncertainty.
func (s *SBASCorrectionStore) EphemerisVarianceAtEpochQueries(broadcast *BroadcastEphemeris, geo string, mode SBASSolveMode, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery) (float64, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return 0, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return 0, err
	}
	value, err := s.handle.EphemerisVarianceAtEpochQueries(broadcast.handle, geo, uint32(mode), satellite, stateEpoch.handle, selectionEpoch.handle)
	return value, publicError(err)
}

func publicSSRSourceState(value native.SSRCorrectedState) SSRCorrectedState {
	return SSRCorrectedState{HasState: value.HasState, PositionECEFM: value.PositionECEFM, ClockS: value.ClockS, HasGroupDelay: value.HasGroupDelay, GroupDelayS: value.GroupDelayS, Degraded: value.Degraded, DegradeReason: UT1DegradeReason(value.DegradeReason), HasSizeEvent: value.HasSizeEvent, StrictRefusal: value.StrictRefusal, Size: SSRCorrectionSize{OrbitM: value.Size.OrbitM, ClockM: value.Size.ClockM}, HasOversizedReport: value.HasOversizedReport, Source: SSRSource(value.Source), ProviderID: value.ProviderID, SolutionID: value.SolutionID, OrbitRefEpochJ2000S: value.OrbitRefEpochJ2000S, ClockRefEpochJ2000S: value.ClockRefEpochJ2000S, FirstAppliedEpochJ2000S: value.FirstAppliedEpochJ2000S}
}

// SourceStateAtEpochQueries evaluates SSR-corrected broadcast state and
// returns selection and size-policy diagnostics.
func (s *SSRCorrectionStore) SourceStateAtEpochQueries(broadcast *BroadcastEphemeris, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy) (SSRCorrectedState, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return SSRCorrectedState{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return SSRCorrectedState{}, err
	}
	if sizePolicy != SSRCorrectionSizeStrict && sizePolicy != SSRCorrectionSizeLenient {
		return SSRCorrectedState{}, invalidArgument("invalid SSR correction size policy")
	}
	value, err := s.handle.SourceStateAtEpochQueries(broadcast.handle, satellite, stateEpoch.handle, selectionEpoch.handle, stalenessSeconds, uint32(missing), allowRegionalProvider, regionalProviderID, uint32(sizePolicy))
	result := publicSSRSourceState(value)
	if err != nil {
		return result, publicError(err)
	}
	return checkedSSRCorrectionResult(result, satellite, stateEpoch, selectionEpoch)
}

// TransmitEpochClockAtEpochQueries evaluates the SSR-corrected transmit clock.
func (s *SSRCorrectionStore) TransmitEpochClockAtEpochQueries(broadcast *BroadcastEphemeris, satellite string, transmitEpoch, selectionEpoch *ExactEpochQuery, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy) (TransmitEpochClock, SSRCorrectedState, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || transmitEpoch == nil || transmitEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return TransmitEpochClock{}, SSRCorrectedState{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return TransmitEpochClock{}, SSRCorrectedState{}, err
	}
	if sizePolicy != SSRCorrectionSizeStrict && sizePolicy != SSRCorrectionSizeLenient {
		return TransmitEpochClock{}, SSRCorrectedState{}, invalidArgument("invalid SSR correction size policy")
	}
	value, policy, err := s.handle.TransmitEpochClockAtEpochQueries(broadcast.handle, satellite, transmitEpoch.handle, selectionEpoch.handle, stalenessSeconds, uint32(missing), allowRegionalProvider, regionalProviderID, uint32(sizePolicy))
	result := publicSSRSourceState(policy)
	if err != nil {
		return TransmitEpochClock{}, result, publicError(err)
	}
	clock := publicTransmitEpochClock(value)
	if policy.StrictRefusal {
		return clock, result, &SSRCorrectionSizeRefusalError{Satellite: satellite, StateEpoch: transmitEpoch, SelectionEpoch: selectionEpoch, Size: result.Size}
	}
	return clock, result, nil
}

// ClockRelativityAtEpochQuery evaluates SSR-corrected clock relativity and
// returns its correction-policy diagnostics.
func (s *SSRCorrectionStore) ClockRelativityAtEpochQuery(broadcast *BroadcastEphemeris, satellite string, epoch *ExactEpochQuery, positionECEFM [3]float64, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy) (ClockRelativity, SSRCorrectedState, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, SSRCorrectedState{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return ClockRelativity{}, SSRCorrectedState{}, err
	}
	if sizePolicy != SSRCorrectionSizeStrict && sizePolicy != SSRCorrectionSizeLenient {
		return ClockRelativity{}, SSRCorrectedState{}, invalidArgument("invalid SSR correction size policy")
	}
	value, policy, err := s.handle.ClockRelativityAtEpochQuery(broadcast.handle, satellite, epoch.handle, positionECEFM, stalenessSeconds, uint32(missing), allowRegionalProvider, regionalProviderID, uint32(sizePolicy))
	result := publicSSRSourceState(policy)
	if err != nil {
		return ClockRelativity{}, result, publicError(err)
	}
	if policy.StrictRefusal {
		return ClockRelativity{Kind: ClockRelativityKind(value.Kind), TermS: value.TermS}, result, &SSRCorrectionSizeRefusalError{Satellite: satellite, StateEpoch: epoch, SelectionEpoch: epoch, Size: result.Size}
	}
	return publicClockRelativity(value), result, nil
}

// EphemerisVarianceAtEpochQueries evaluates SSR-selected uncertainty and
// returns its correction-policy diagnostics.
func (s *SSRCorrectionStore) EphemerisVarianceAtEpochQueries(broadcast *BroadcastEphemeris, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy) (float64, SSRCorrectedState, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return 0, SSRCorrectedState{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return 0, SSRCorrectedState{}, err
	}
	if sizePolicy != SSRCorrectionSizeStrict && sizePolicy != SSRCorrectionSizeLenient {
		return 0, SSRCorrectedState{}, invalidArgument("invalid SSR correction size policy")
	}
	value, policy, err := s.handle.EphemerisVarianceAtEpochQueries(broadcast.handle, satellite, stateEpoch.handle, selectionEpoch.handle, stalenessSeconds, uint32(missing), allowRegionalProvider, regionalProviderID, uint32(sizePolicy))
	result := publicSSRSourceState(policy)
	if err != nil {
		return 0, result, publicError(err)
	}
	if policy.StrictRefusal {
		return value, result, &SSRCorrectionSizeRefusalError{Satellite: satellite, StateEpoch: stateEpoch, SelectionEpoch: selectionEpoch, Size: result.Size}
	}
	return value, result, nil
}
