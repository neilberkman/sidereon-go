package sidereon

import (
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// TEMEState is one TEME state copied from C TLE propagation. Position is
// in kilometers, velocity is in kilometers per second, and EpochJ2000S is the
// corresponding UTC epoch in seconds since J2000 computed by C.
type TEMEState struct {
	// EpochJ2000S is the epoch j2000 s in seconds.
	EpochJ2000S float64
	// PositionKm is the position km in kilometres.
	PositionKm [3]float64
	// VelocityKmPerS is the velocity km per s in kilometres per second.
	VelocityKmPerS [3]float64
}

// TLEMetadata contains read-only parsed element fields copied from C.
type TLEMetadata struct {
	// CatalogNumber identifies or counts this record.
	CatalogNumber string
	// Classification is the classification value for TLEMetadata.
	Classification string
	// InternationalDesignator is the international designator value for TLEMetadata.
	InternationalDesignator string
	// EpochYear is the epoch year value for TLEMetadata.
	EpochYear int
	// EpochDayOfYear is the epoch day of year value for TLEMetadata.
	EpochDayOfYear float64
	// InclinationDeg is the inclination deg in degrees.
	InclinationDeg float64
	// RAANDeg is the raan deg in degrees.
	RAANDeg float64
	// Eccentricity is the eccentricity value for TLEMetadata.
	Eccentricity float64
	// ArgumentOfPerigeeDeg is the argument of perigee deg in degrees.
	ArgumentOfPerigeeDeg float64
	// MeanAnomalyDeg is the mean anomaly deg in degrees.
	MeanAnomalyDeg float64
	// MeanMotionRevPerDay is the mean motion rev per day in revolutions per day.
	MeanMotionRevPerDay float64
	// MeanMotionDot is the mean motion dot value for TLEMetadata.
	MeanMotionDot float64
	// MeanMotionDoubleDot is the mean motion double dot value for TLEMetadata.
	MeanMotionDoubleDot float64
	// BStar is the bstar value for TLEMetadata.
	BStar float64
	// EphemerisType is the ephemeris type value for TLEMetadata.
	EphemerisType int
	// ElementSetNumber identifies or counts this record.
	ElementSetNumber int
	// RevolutionNumber identifies or counts this record.
	RevolutionNumber int
}

// TLELines contains C's re-encoded TLE lines.
type TLELines struct {
	// Line1 is the line1 value for TLELines.
	Line1 string
	// Line2 is the line2 value for TLELines.
	Line2 string
}

// TLE owns a parsed C TLE handle. It must not be copied after first use.
// Read operations may run concurrently with one another and with Close.
// Close waits for in-flight reads, clears the native state, and is idempotent;
// operations attempted after Close return ErrClosed.
type TLE struct {
	_      noCopy
	handle *native.TLE
}

// OpsMode selects the C SGP4 operations mode.
type OpsMode uint32

const (
	// OpsModeAFSPC selects the historical AFSPC SGP4 mode.
	OpsModeAFSPC OpsMode = OpsMode(native.TLEOpsModeAFSPCValue)
	// OpsModeImproved selects the improved SGP4 operations mode.
	OpsModeImproved OpsMode = OpsMode(native.TLEOpsModeImprovedValue)
)

// ParseTLE parses two TLE lines through the C ABI using OpsModeAFSPC.
func ParseTLE(line1, line2 string) (*TLE, error) {
	return ParseTLEWithOpsMode(line1, line2, OpsModeAFSPC)
}

// ParseTLEWithOpsMode parses two TLE lines using an explicitly selected SGP4
// operations mode.
func ParseTLEWithOpsMode(line1, line2 string, mode OpsMode) (*TLE, error) {
	handle, err := native.LoadTLE(line1, line2, uint32(mode))
	if err != nil {
		return nil, publicError(err)
	}
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &TLE{handle: handle}, nil
}

// Close releases the native TLE handle. It waits for in-flight read operations,
// clears the native state, is idempotent, and is safe to call concurrently with
// read operations.
func (t *TLE) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return publicError(t.handle.Close())
}

// Metadata returns copied parsed TLE metadata.
func (t *TLE) Metadata() (TLEMetadata, error) {
	if t == nil || t.handle == nil {
		return TLEMetadata{}, ErrClosed
	}
	metadata, err := t.handle.Metadata()
	if err != nil {
		return TLEMetadata{}, publicError(err)
	}
	return TLEMetadata{
		CatalogNumber:           metadata.CatalogNumber,
		Classification:          metadata.Classification,
		InternationalDesignator: metadata.InternationalDesignator,
		EpochYear:               metadata.EpochYear,
		EpochDayOfYear:          metadata.EpochDayOfYear,
		InclinationDeg:          metadata.InclinationDeg,
		RAANDeg:                 metadata.RAANDeg,
		Eccentricity:            metadata.Eccentricity,
		ArgumentOfPerigeeDeg:    metadata.ArgumentOfPerigeeDeg,
		MeanAnomalyDeg:          metadata.MeanAnomalyDeg,
		MeanMotionRevPerDay:     metadata.MeanMotionRevPerDay,
		MeanMotionDot:           metadata.MeanMotionDot,
		MeanMotionDoubleDot:     metadata.MeanMotionDoubleDot,
		BStar:                   metadata.BStar,
		EphemerisType:           metadata.EphemerisType,
		ElementSetNumber:        metadata.ElementSetNumber,
		RevolutionNumber:        metadata.RevolutionNumber,
	}, nil
}

// Lines returns C's re-encoded lines, suitable for a round trip comparison.
func (t *TLE) Lines() (TLELines, error) {
	if t == nil || t.handle == nil {
		return TLELines{}, ErrClosed
	}
	lines, err := t.handle.Lines()
	if err != nil {
		return TLELines{}, publicError(err)
	}
	return TLELines{Line1: lines.Line1, Line2: lines.Line2}, nil
}

// Propagate returns copied TEME Cartesian states for the supplied UTC times.
// C performs the civil-to-J2000 conversion and all propagation mathematics.
// The C propagation input is UTC Unix microseconds, so sub-microsecond input
// precision is not represented in the returned state epoch.
func (t *TLE) Propagate(times []time.Time) ([]TEMEState, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	states, err := t.handle.Propagate(times)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]TEMEState, len(states))
	for i, state := range states {
		out[i] = TEMEState{
			EpochJ2000S:    state.EpochJ2000S,
			PositionKm:     state.PositionKm,
			VelocityKmPerS: state.VelocityKmPerS,
		}
	}
	return out, nil
}
