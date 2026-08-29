package sidereon

import (
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// TEMEState is one TEME state detached from C TLE propagation. Position is
// in kilometers, velocity is in kilometers per second, and EpochJ2000S is the
// corresponding UTC epoch in seconds since J2000 computed by C.
type TEMEState struct {
	// EpochJ2000S is the corresponding UTC epoch in seconds from J2000.
	EpochJ2000S float64
	// PositionKm is the position km in kilometres.
	PositionKm [3]float64
	// VelocityKmPerS is the velocity km per s in kilometres per second.
	VelocityKmPerS [3]float64
}

// TLEMetadata contains read-only parsed element fields detached from C.
type TLEMetadata struct {
	// CatalogNumber is the NORAD catalog identifier.
	CatalogNumber string
	// Classification is the one-character TLE object classification.
	Classification string
	// InternationalDesignator is the launch-year, launch-piece international designator.
	InternationalDesignator string
	// EpochYear is the two-digit TLE epoch year resolved by the native parser.
	EpochYear int
	// EpochDayOfYear is the fractional day of year at the TLE epoch.
	EpochDayOfYear float64
	// InclinationDeg is the inclination deg in degrees.
	InclinationDeg float64
	// RAANDeg is the raan deg in degrees.
	RAANDeg float64
	// Eccentricity is the dimensionless TLE eccentricity.
	Eccentricity float64
	// ArgumentOfPerigeeDeg is the argument of perigee deg in degrees.
	ArgumentOfPerigeeDeg float64
	// MeanAnomalyDeg is the mean anomaly deg in degrees.
	MeanAnomalyDeg float64
	// MeanMotionRevPerDay is the mean motion rev per day in revolutions per day.
	MeanMotionRevPerDay float64
	// MeanMotionDot is the first mean-motion derivative in revolutions per day squared.
	MeanMotionDot float64
	// MeanMotionDoubleDot is the second mean-motion derivative in revolutions per day cubed.
	MeanMotionDoubleDot float64
	// BStar is the parsed SGP4 B* drag term.
	BStar float64
	// EphemerisType is the TLE ephemeris-type code.
	EphemerisType int
	// ElementSetNumber is the TLE element-set sequence number.
	ElementSetNumber int
	// RevolutionNumber is the revolution number at the TLE epoch.
	RevolutionNumber int
}

// TLELines contains C's re-encoded TLE lines.
type TLELines struct {
	// Line1 is the reconstructed first TLE line, including its checksum.
	Line1 string
	// Line2 is the reconstructed second TLE line, including its checksum.
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

// Metadata returns detached parsed TLE metadata.
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

// Propagate returns detached TEME Cartesian states for the supplied UTC times.
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
