package sidereon

import (
	"encoding/json"
	"time"

	"sidereon.dev/go/v3/internal/native"
)

// CoverageLookAngle is one satellite/station cell from a coverage grid. OK
// is false when C could not compute a look angle for that pair.
type CoverageLookAngle struct {
	// OK reports whether C computed a look angle for this pair.
	OK bool
	// AzimuthDeg is the azimuth deg in degrees.
	AzimuthDeg float64
	// ElevationDeg is the elevation deg in degrees.
	ElevationDeg float64
	// RangeKm is the range km in kilometres.
	RangeKm float64
}

// CoverageGrid owns a C-computed one-epoch coverage grid. It must not be
// copied after first use. Queries may run concurrently with Close; Close is
// idempotent and waits for in-flight queries.
type CoverageGrid struct {
	_      noCopy
	handle *native.CoverageGrid
}

// CoverageLookAngles computes a one-epoch grid for TLEs and WGS84 stations.
// Station latitude/longitude are degrees and altitude is metres.
func CoverageLookAngles(tles []*TLE, stations []PassStation, epoch time.Time) (*CoverageGrid, error) {
	nativeStations := make([]native.GroundStation, len(stations))
	for i, station := range stations {
		nativeStations[i] = native.GroundStation{LatitudeDeg: station.LatitudeDeg, LongitudeDeg: station.LongitudeDeg, AltitudeM: station.AltitudeM}
	}
	epochUS, err := metadataUnixMicroseconds(epoch)
	if err != nil {
		return nil, err
	}
	value, err := native.CoverageLookAngles(nativeTLEs(tles), nativeStations, epochUS)
	if err != nil {
		return nil, publicError(err)
	}
	return &CoverageGrid{handle: value}, nil
}

func nativeTLEs(values []*TLE) []*native.TLE {
	out := make([]*native.TLE, len(values))
	for i, value := range values {
		if value != nil {
			out[i] = value.handle
		}
	}
	return out
}

// Close releases the C coverage grid and is safe to call repeatedly.
func (g *CoverageGrid) Close() error {
	if g == nil || g.handle == nil {
		return nil
	}
	return publicError(g.handle.Close())
}

// Dimensions returns the satellite and station dimensions of the grid.
func (g *CoverageGrid) Dimensions() (satellites, stations int, err error) {
	if g == nil || g.handle == nil {
		return 0, 0, ErrClosed
	}
	satellites, stations, err = g.handle.Dimensions()
	return satellites, stations, publicError(err)
}

// CoverageGridDimensions returns the satellite and station dimensions of a
// coverage grid.
func CoverageGridDimensions(g *CoverageGrid) (satellites, stations int, err error) {
	return g.Dimensions()
}

// LookAngle returns one copied grid cell.
func (g *CoverageGrid) LookAngle(satelliteIndex, stationIndex int) (CoverageLookAngle, error) {
	if g == nil || g.handle == nil {
		return CoverageLookAngle{}, ErrClosed
	}
	value, err := g.handle.LookAngle(satelliteIndex, stationIndex)
	return CoverageLookAngle{OK: value.OK, AzimuthDeg: value.AzimuthDeg, ElevationDeg: value.ElevationDeg, RangeKm: value.RangeKm}, publicError(err)
}

// LookAngleErrorPayload copies the complete typed JSON error for a failed
// cell. The bytes remain usable after the grid is closed.
func (g *CoverageGrid) LookAngleErrorPayload(satelliteIndex, stationIndex int) ([]byte, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	value, err := g.handle.LookAngleErrorPayload(satelliteIndex, stationIndex)
	return value, publicError(err)
}

// LookAngleError returns the tagged, complete error details for a failed cell.
// Raw retains the exact JSON so future fields remain accessible.
func (g *CoverageGrid) LookAngleError(satelliteIndex, stationIndex int) (CoverageLookAngleError, error) {
	payload, err := g.LookAngleErrorPayload(satelliteIndex, stationIndex)
	if err != nil {
		return CoverageLookAngleError{}, err
	}
	return decodeCoverageLookAngleError(payload)
}

func decodeCoverageLookAngleError(payload []byte) (CoverageLookAngleError, error) {
	var value CoverageLookAngleError
	if err := json.Unmarshal(payload, &value); err != nil {
		return CoverageLookAngleError{}, err
	}
	value.Raw = append(json.RawMessage(nil), payload...)
	var fields struct {
		Fields struct {
			Cause json.RawMessage `json:"cause"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(payload, &fields); err != nil {
		return CoverageLookAngleError{}, err
	}
	if value.Fields.Cause != nil {
		value.Fields.Cause.Raw = append(json.RawMessage(nil), fields.Fields.Cause...)
	}
	return value, nil
}

// CoverageLookAngleError is the typed error for one failed coverage cell.
type CoverageLookAngleError struct {
	Kind   string                       `json:"kind"`
	Fields CoverageLookAngleErrorFields `json:"fields"`
	Raw    json.RawMessage              `json:"-"`
}

// CoverageLookAngleErrorFields holds fields used by all current cell errors.
type CoverageLookAngleErrorFields struct {
	Field  string                       `json:"field,omitempty"`
	Reason string                       `json:"reason,omitempty"`
	Cause  *CoverageLookAngleErrorCause `json:"cause,omitempty"`
}

// CoverageLookAngleErrorCause is a typed nested SGP4 or frame-transform error.
type CoverageLookAngleErrorCause struct {
	Kind   string                            `json:"kind"`
	Fields CoverageLookAngleErrorCauseFields `json:"fields"`
	Raw    json.RawMessage                   `json:"-"`
}

// CoverageLookAngleErrorCauseFields contains all currently serialized nested
// diagnostic fields.
type CoverageLookAngleErrorCauseFields struct {
	Field     string  `json:"field,omitempty"`
	InputKind string  `json:"kind,omitempty"`
	Message   string  `json:"message,omitempty"`
	Code      *int32  `json:"code,omitempty"`
	Budget    *uint64 `json:"budget,omitempty"`
	Reason    string  `json:"reason,omitempty"`
}

// CoverageGridLookAngle returns one copied cell from a coverage grid.
func CoverageGridLookAngle(g *CoverageGrid, satelliteIndex, stationIndex int) (CoverageLookAngle, error) {
	return g.LookAngle(satelliteIndex, stationIndex)
}

// CoverageGridAccessCounts returns the number of successful accesses per
// station above the supplied elevation mask in degrees.
func CoverageGridAccessCounts(g *CoverageGrid, minElevationDeg float64) ([]int, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	value, err := g.handle.AccessCounts(minElevationDeg)
	return value, publicError(err)
}

// AccessCounts is the method form of CoverageGridAccessCounts.
func (g *CoverageGrid) AccessCounts(minElevationDeg float64) ([]int, error) {
	return CoverageGridAccessCounts(g, minElevationDeg)
}

// CoverageGridMaxElevationDeg returns each station's maximum successful
// elevation in degrees. Stations without a successful cell contain NaN.
func CoverageGridMaxElevationDeg(g *CoverageGrid) ([]float64, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	value, err := g.handle.MaxElevationDeg()
	return value, publicError(err)
}

// MaxElevationDeg is the method form of CoverageGridMaxElevationDeg.
func (g *CoverageGrid) MaxElevationDeg() ([]float64, error) {
	return CoverageGridMaxElevationDeg(g)
}

// CoverageGridVisibleMask returns the flattened satellite-major visibility
// mask for the supplied elevation threshold.
func CoverageGridVisibleMask(g *CoverageGrid, minElevationDeg float64) ([]bool, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	value, err := g.handle.VisibleMask(minElevationDeg)
	return value, publicError(err)
}

// VisibleMask is the method form of CoverageGridVisibleMask.
func (g *CoverageGrid) VisibleMask(minElevationDeg float64) ([]bool, error) {
	return CoverageGridVisibleMask(g, minElevationDeg)
}
