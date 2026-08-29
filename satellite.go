package sidereon

import (
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

func satelliteUnixMicroseconds(value time.Time) (int64, error) {
	return metadataUnixMicroseconds(value)
}

// SatelliteConstellation owns a borrowed-TLE constellation computation handle.
type SatelliteConstellation struct {
	_      noCopy
	handle *native.SatelliteConstellation
}

// ConstellationLookAngleArcs owns flattened look-angle arcs.
type ConstellationLookAngleArcs struct {
	_      noCopy
	handle *native.ConstellationLookAngleArcs
}

// ConstellationGroundTracks owns flattened ground-track arcs.
type ConstellationGroundTracks struct {
	_      noCopy
	handle *native.ConstellationGroundTracks
}

// FleetPass associates a pass with its fleet-order satellite index.
type FleetPass struct {
	// SatelliteIndex is the zero-based index of the satellite in the fleet input.
	SatelliteIndex int
	// Pass contains the visibility interval and peak look-angle data.
	Pass SatellitePass
}

// ConstellationPasses owns flattened constellation pass results.
type ConstellationPasses struct {
	_      noCopy
	handle *native.ConstellationPasses
}

// NewSatelliteConstellation builds a constellation from parsed TLE handles.
func NewSatelliteConstellation(tles []*TLE) (*SatelliteConstellation, error) {
	raw := make([]*native.TLE, len(tles))
	for i, t := range tles {
		if t == nil || t.handle == nil {
			return nil, ErrClosed
		}
		raw[i] = t.handle
	}
	v, e := native.BuildSatelliteConstellation(raw)
	if e != nil {
		return nil, publicError(e)
	}
	return &SatelliteConstellation{handle: v}, nil
}

// BuildSatelliteConstellation is an alias for NewSatelliteConstellation.
func BuildSatelliteConstellation(tles []*TLE) (*SatelliteConstellation, error) {
	return NewSatelliteConstellation(tles)
}

// Close releases the constellation; it is safe to call more than once.
func (s *SatelliteConstellation) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Count returns the number of satellites in the constellation.
func (s *SatelliteConstellation) Count() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	v, e := s.handle.Count()
	return v, publicError(e)
}

// CatalogNumber returns the native catalog number at fleet index i.
func (s *SatelliteConstellation) CatalogNumber(i int) (string, error) {
	if s == nil || s.handle == nil {
		return "", ErrClosed
	}
	v, e := s.handle.CatalogNumber(i)
	return v, publicError(e)
}
func timeEpochs(values []time.Time) ([]int64, error) {
	out := make([]int64, len(values))
	for i, v := range values {
		x, e := satelliteUnixMicroseconds(v)
		if e != nil {
			return nil, e
		}
		out[i] = x
	}
	return out, nil
}

// Propagate computes the shared-epoch TLE state matrix.
func (s *SatelliteConstellation) Propagate(epochs []time.Time, parallel bool) (*TLEBatchPropagation, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	x, e := timeEpochs(epochs)
	if e != nil {
		return nil, e
	}
	v, e := s.handle.Propagate(x, parallel)
	if e != nil {
		return nil, publicError(e)
	}
	return &TLEBatchPropagation{handle: v}, nil
}

// Visible returns satellites above the elevation mask at one epoch.
func (s *SatelliteConstellation) Visible(station PassStation, epoch time.Time, minElevationDeg float64) (*VisibleList, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	us, e := satelliteUnixMicroseconds(epoch)
	if e != nil {
		return nil, e
	}
	v, e := s.handle.Visible(native.GroundStation{LatitudeDeg: station.LatitudeDeg, LongitudeDeg: station.LongitudeDeg, AltitudeM: station.AltitudeM}, us, minElevationDeg)
	if e != nil {
		return nil, publicError(e)
	}
	return &VisibleList{handle: v}, nil
}

// LookAngles computes flattened topocentric arcs for all satellites.
func (s *SatelliteConstellation) LookAngles(station PassStation, epochs []time.Time, parallel bool) (*ConstellationLookAngleArcs, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	x, e := timeEpochs(epochs)
	if e != nil {
		return nil, e
	}
	v, e := s.handle.LookAngles(native.GroundStation{LatitudeDeg: station.LatitudeDeg, LongitudeDeg: station.LongitudeDeg, AltitudeM: station.AltitudeM}, x, parallel)
	if e != nil {
		return nil, publicError(e)
	}
	return &ConstellationLookAngleArcs{handle: v}, nil
}

// GroundTracks computes flattened sub-satellite geodetic arcs.
func (s *SatelliteConstellation) GroundTracks(epochs []time.Time) (*ConstellationGroundTracks, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	x, e := timeEpochs(epochs)
	if e != nil {
		return nil, e
	}
	v, e := s.handle.GroundTracks(x)
	if e != nil {
		return nil, publicError(e)
	}
	return &ConstellationGroundTracks{handle: v}, nil
}

// Close releases look-angle arcs; it is safe to call more than once.
func (a *ConstellationLookAngleArcs) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return publicError(a.handle.Close())
}

// SatelliteCount returns the number of look-angle arcs.
func (a *ConstellationLookAngleArcs) SatelliteCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	v, e := a.handle.SatelliteCount()
	return v, publicError(e)
}

// ArcLengths returns each satellite's look-angle arc length.
func (a *ConstellationLookAngleArcs) ArcLengths() ([]int, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	v, e := a.handle.ArcLengths()
	return v, publicError(e)
}

// Values returns detached flattened look-angle rows.
func (a *ConstellationLookAngleArcs) Values() ([]Topocentric, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	v, e := a.handle.Values()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]Topocentric, len(v))
	for i, x := range v {
		out[i] = Topocentric{AzimuthDeg: x.AzimuthDeg, ElevationDeg: x.ElevationDeg, RangeKm: x.RangeKm}
	}
	return out, nil
}

// Close releases ground tracks; it is safe to call more than once.
func (g *ConstellationGroundTracks) Close() error {
	if g == nil || g.handle == nil {
		return nil
	}
	return publicError(g.handle.Close())
}

// SatelliteCount returns the number of ground tracks.
func (g *ConstellationGroundTracks) SatelliteCount() (int, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	v, e := g.handle.SatelliteCount()
	return v, publicError(e)
}

// TrackLengths returns each satellite's ground-track length.
func (g *ConstellationGroundTracks) TrackLengths() ([]int, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	v, e := g.handle.TrackLengths()
	return v, publicError(e)
}

// Values returns detached flattened geodetic track rows.
func (g *ConstellationGroundTracks) Values() ([]Geodetic, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	x, e := g.handle.Values()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]Geodetic, len(x))
	for i, v := range x {
		out[i] = Geodetic{LatitudeRad: v.LatitudeRad, LongitudeRad: v.LongitudeRad, HeightM: v.HeightM}
	}
	return out, nil
}

// Passes finds native dense passes in the requested UTC interval.
func (s *SatelliteConstellation) Passes(station PassStation, start, end time.Time, options *PassFinderOptions) (*ConstellationPasses, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	a, e := satelliteUnixMicroseconds(start)
	if e != nil {
		return nil, e
	}
	b, e := satelliteUnixMicroseconds(end)
	if e != nil {
		return nil, e
	}
	var no *native.PassFinderOptions
	if options != nil {
		x := native.PassFinderOptions{ElevationMaskDeg: options.ElevationMaskDeg, StepSeconds: options.StepSeconds, TimeToleranceSeconds: options.TimeToleranceSeconds}
		no = &x
	}
	v, e := s.handle.Passes(native.GroundStation{LatitudeDeg: station.LatitudeDeg, LongitudeDeg: station.LongitudeDeg, AltitudeM: station.AltitudeM}, a, b, no)
	if e != nil {
		return nil, publicError(e)
	}
	return &ConstellationPasses{handle: v}, nil
}

// Close releases constellation passes; it is safe to call more than once.
func (p *ConstellationPasses) Close() error {
	if p == nil || p.handle == nil {
		return nil
	}
	return publicError(p.handle.Close())
}

// Count returns the number of flattened passes.
func (p *ConstellationPasses) Count() (int, error) {
	if p == nil || p.handle == nil {
		return 0, ErrClosed
	}
	v, e := p.handle.Count()
	return v, publicError(e)
}

// Values returns detached fleet pass rows.
func (p *ConstellationPasses) Values() ([]FleetPass, error) {
	if p == nil || p.handle == nil {
		return nil, ErrClosed
	}
	x, e := p.handle.Values()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]FleetPass, len(x))
	for i, v := range x {
		out[i] = FleetPass{SatelliteIndex: v.SatelliteIndex, Pass: SatellitePass{AOS: v.Pass.AOS, LOS: v.Pass.LOS, Culmination: v.Pass.Culmination, MaxElevationDeg: v.Pass.MaxElevationDeg, DurationS: v.Pass.DurationS}}
	}
	return out, nil
}

// SatelliteVisualMagnitude evaluates the native diffuse-sphere phase law.
func SatelliteVisualMagnitude(rangeKm, phaseAngleDeg, standardMagnitude, referenceRangeKm float64) (float64, error) {
	v, e := native.SatelliteVisualMagnitude(rangeKm, phaseAngleDeg, standardMagnitude, referenceRangeKm)
	return v, publicError(e)
}
