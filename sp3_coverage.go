package sidereon

import "sidereon.dev/go/v3/internal/native"

// SP3CoverageGrid describes the product epoch grid without discarding irregular epochs.
type SP3CoverageGrid struct {
	// HasInterval reports whether all epochs lie on one regular grid.
	HasInterval bool
	// IntervalS is the inferred grid step when HasInterval is true.
	IntervalS float64
	// AgreesWithHeader reports whether the inferred grid matches the declared interval.
	AgreesWithHeader bool
	// OutOfOrder contains indices whose timestamps are not in file order.
	OutOfOrder []int
	// Unplaced contains indices without an exact SP3 record epoch.
	Unplaced []int
}

// SP3CoverageSpan records one contiguous run carrying a satellite channel.
type SP3CoverageSpan struct {
	// FirstIndex is the first product epoch index in the run.
	FirstIndex int
	// LastIndex is the last product epoch index in the run.
	LastIndex int
	// FirstEpoch is the exact tagged time of the first epoch.
	FirstEpoch ClockEpoch
	// FirstJ2000 is the first epoch in seconds from J2000, or NaN if unreadable.
	FirstJ2000 float64
	// LastEpoch is the exact tagged time of the last epoch.
	LastEpoch ClockEpoch
	// LastJ2000 is the last epoch in seconds from J2000, or NaN if unreadable.
	LastJ2000 float64
}

// SP3CoverageGap records an interval where a satellite channel has no data.
type SP3CoverageGap struct {
	// HasAfterIndex indicates that a prior covered epoch exists.
	HasAfterIndex bool
	// AfterIndex is the last covered epoch before the gap when present.
	AfterIndex int
	// HasBeforeIndex indicates that a following covered epoch exists.
	HasBeforeIndex bool
	// BeforeIndex is the first covered epoch after the gap when present.
	BeforeIndex int
	// MissingEpochs is the number of absent grid epochs represented by the gap.
	MissingEpochs int
}

// SP3ChannelCoverage describes position or clock coverage for one satellite.
type SP3ChannelCoverage struct {
	// Epochs is the number of product epochs carrying this channel.
	Epochs int
	// SpanCount is the number of contiguous spans.
	SpanCount int
	// GapCount is the number of gaps.
	GapCount int
	// Complete reports that the channel is present at each product epoch in one span.
	Complete bool
	// Spans contains the detached channel spans in time order.
	Spans []SP3CoverageSpan
	// Gaps contains the detached channel gaps in time order.
	Gaps []SP3CoverageGap
}

// SP3CoverageSatellite reports the declared and observed channels for one satellite.
type SP3CoverageSatellite struct {
	// Satellite is the canonical satellite token.
	Satellite string
	// Declared reports whether the SP3 header declares this satellite.
	Declared bool
	// Positions is the orbit-position coverage.
	Positions SP3ChannelCoverage
	// Clocks is the satellite-clock coverage.
	Clocks SP3ChannelCoverage
}

// SP3Coverage is an owned, detached summary of the product's per-satellite channels.
type SP3Coverage struct {
	// Grid describes product-wide epoch ordering and interval.
	Grid SP3CoverageGrid
	// Satellites is ordered by canonical satellite identifier; it includes both
	// header-declared satellites and satellites carried only by records.
	Satellites []SP3CoverageSatellite
}

// Coverage returns exact grid diagnostics and per-satellite position and clock coverage.
func (s *SP3) Coverage() (SP3Coverage, error) {
	if s == nil || s.handle == nil {
		return SP3Coverage{}, ErrClosed
	}
	value, err := s.handle.Coverage()
	if err != nil {
		return SP3Coverage{}, publicError(err)
	}
	out := SP3Coverage{Grid: SP3CoverageGrid{HasInterval: value.Grid.HasInterval, IntervalS: value.Grid.IntervalS, AgreesWithHeader: value.Grid.AgreesWithHeader, OutOfOrder: append([]int(nil), value.Grid.OutOfOrder...), Unplaced: append([]int(nil), value.Grid.Unplaced...)}}
	out.Satellites = make([]SP3CoverageSatellite, len(value.Satellites))
	for i, sat := range value.Satellites {
		out.Satellites[i] = SP3CoverageSatellite{Satellite: sat.Satellite, Declared: sat.Declared, Positions: publicSP3ChannelCoverage(sat.Positions), Clocks: publicSP3ChannelCoverage(sat.Clocks)}
	}
	return out, nil
}

// SelectedNodes reports exact product nodes selected by queries in the inclusive time window.
// A satellite or window with no selected nodes returns an empty slice.
func (s *SP3) SelectedNodes(satellite string, fromJ2000S, throughJ2000S float64) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.SelectedNodes(satellite, fromJ2000S, throughJ2000S)
	if err != nil {
		return nil, publicError(err)
	}
	return append([]float64(nil), values...), nil
}

func publicSP3ChannelCoverage(value native.SP3ChannelCoverage) SP3ChannelCoverage {
	out := SP3ChannelCoverage{Epochs: value.Epochs, SpanCount: value.SpanCount, GapCount: value.GapCount, Complete: value.Complete}
	out.Spans = make([]SP3CoverageSpan, len(value.Spans))
	for i, span := range value.Spans {
		out.Spans[i] = SP3CoverageSpan{FirstIndex: span.FirstIndex, LastIndex: span.LastIndex, FirstEpoch: publicClockEpoch(span.FirstEpoch), FirstJ2000: span.FirstJ2000, LastEpoch: publicClockEpoch(span.LastEpoch), LastJ2000: span.LastJ2000}
	}
	out.Gaps = make([]SP3CoverageGap, len(value.Gaps))
	for i, gap := range value.Gaps {
		out.Gaps[i] = SP3CoverageGap{HasAfterIndex: gap.HasAfterIndex, AfterIndex: gap.AfterIndex, HasBeforeIndex: gap.HasBeforeIndex, BeforeIndex: gap.BeforeIndex, MissingEpochs: gap.MissingEpochs}
	}
	return out
}
