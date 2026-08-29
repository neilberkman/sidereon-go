package sidereon

import (
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// PassStation is a WGS84 station for TLE pass and look-angle operations.
// Latitude and longitude are degrees and altitude is meters.
type PassStation struct {
	LatitudeDeg  float64
	LongitudeDeg float64
	AltitudeM    float64
}

// SatellitePass is a C-found visibility event. Event times are UTC with
// microsecond precision; maximum elevation is degrees and duration is seconds.
type SatellitePass struct {
	AOS             time.Time
	LOS             time.Time
	Culmination     time.Time
	MaxElevationDeg float64
	DurationS       float64
}

// PassFinderOptions controls C pass finding. The elevation mask is degrees;
// step and time tolerance are seconds.
type PassFinderOptions struct {
	ElevationMaskDeg     float64
	StepSeconds          float64
	TimeToleranceSeconds float64
}

// PassList owns C-found pass rows and must not be copied after first use.
// Count and Values are read-only and may run concurrently with one another and
// with Close. Close waits for active calls, clears the native resource, and is
// idempotent; methods called after Close return ErrClosed.
type PassList struct {
	_      noCopy
	handle *native.PassList
}

// GroundTrack owns C-propagated sub-satellite points and must not be copied
// after first use. Count and Values are read-only and may run concurrently with
// one another and with Close. Close waits for active calls, clears the native
// resource, and is idempotent; methods called after Close return ErrClosed.
type GroundTrack struct {
	_      noCopy
	handle *native.GroundTrack
}

// LookAngles owns C-propagated topocentric rows and must not be copied after
// first use. Count and Values are read-only and may run concurrently with one
// another and with Close. Close waits for active calls, clears the native
// resource, and is idempotent; methods called after Close return ErrClosed.
type LookAngles struct {
	_      noCopy
	handle *native.LookAngles
}

func nativePassStation(value PassStation) native.GroundStation {
	return native.GroundStation{
		LatitudeDeg:  value.LatitudeDeg,
		LongitudeDeg: value.LongitudeDeg,
		AltitudeM:    value.AltitudeM,
	}
}

func publicPass(value native.SatellitePass) SatellitePass {
	return SatellitePass{
		AOS:             value.AOS,
		LOS:             value.LOS,
		Culmination:     value.Culmination,
		MaxElevationDeg: value.MaxElevationDeg,
		DurationS:       value.DurationS,
	}
}

// DefaultPassFinderOptions returns the C engine defaults.
func DefaultPassFinderOptions() (PassFinderOptions, error) {
	value, err := native.PassFinderDefaults()
	return PassFinderOptions{
		ElevationMaskDeg:     value.ElevationMaskDeg,
		StepSeconds:          value.StepSeconds,
		TimeToleranceSeconds: value.TimeToleranceSeconds,
	}, publicError(err)
}

// FindPasses returns a C-owned pass list for the half-open UTC interval
// [start, end). C receives the input times as Unix microseconds.
func (t *TLE) FindPasses(station PassStation, start, end time.Time, options *PassFinderOptions) (*PassList, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var nativeOptions *native.PassFinderOptions
	if options != nil {
		nativeOptions = &native.PassFinderOptions{
			ElevationMaskDeg:     options.ElevationMaskDeg,
			StepSeconds:          options.StepSeconds,
			TimeToleranceSeconds: options.TimeToleranceSeconds,
		}
	}
	value, err := t.handle.FindPasses(nativePassStation(station), start, end, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	return &PassList{handle: value}, nil
}

// FindPassesValues is the value convenience form of FindPasses. It closes
// the temporary native result after copying it.
func (t *TLE) FindPassesValues(station PassStation, start, end time.Time, options *PassFinderOptions) ([]SatellitePass, error) {
	passes, err := t.FindPasses(station, start, end, options)
	if err != nil {
		return nil, err
	}
	values, valueErr := passes.Values()
	closeErr := passes.Close()
	return values, joinPublicErrors(valueErr, closeErr)
}

// Close releases a pass list and is idempotent.
func (p *PassList) Close() error {
	if p == nil || p.handle == nil {
		return nil
	}
	return publicError(p.handle.Close())
}

// Count returns the number of C-found passes.
func (p *PassList) Count() (int, error) {
	if p == nil || p.handle == nil {
		return 0, ErrClosed
	}
	value, err := p.handle.Count()
	return value, publicError(err)
}

// Values copies pass rows into independent Go memory. Event times preserve
// explicit microsecond precision and are normalized to UTC.
func (p *PassList) Values() ([]SatellitePass, error) {
	if p == nil || p.handle == nil {
		return nil, ErrClosed
	}
	values, err := p.handle.Values()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]SatellitePass, len(values))
	for i := range out {
		out[i] = publicPass(values[i])
	}
	return out, nil
}

// GroundTrack propagates the TLE at caller-supplied UTC epochs with
// microsecond precision.
func (t *TLE) GroundTrack(epochs []time.Time) (*GroundTrack, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	value, err := t.handle.GroundTrack(epochs)
	if err != nil {
		return nil, publicError(err)
	}
	return &GroundTrack{handle: value}, nil
}

// Close releases a ground-track result and is idempotent.
func (g *GroundTrack) Close() error {
	if g == nil || g.handle == nil {
		return nil
	}
	return publicError(g.handle.Close())
}

// Count returns the number of ground-track points.
func (g *GroundTrack) Count() (int, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	value, err := g.handle.Count()
	return value, publicError(err)
}

// Values copies ground-track points into Go-owned geodetic values.
func (g *GroundTrack) Values() ([]Geodetic, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	values, err := g.handle.Values()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]Geodetic, len(values))
	for i := range out {
		out[i] = Geodetic(values[i])
	}
	return out, nil
}

// LookAngles propagates topocentric angles at caller-supplied UTC epochs with
// microsecond precision.
func (t *TLE) LookAngles(station PassStation, epochs []time.Time) (*LookAngles, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	value, err := t.handle.LookAngles(nativePassStation(station), epochs)
	if err != nil {
		return nil, publicError(err)
	}
	return &LookAngles{handle: value}, nil
}

// Close releases a look-angle result and is idempotent.
func (l *LookAngles) Close() error {
	if l == nil || l.handle == nil {
		return nil
	}
	return publicError(l.handle.Close())
}

// Count returns the number of look-angle rows.
func (l *LookAngles) Count() (int, error) {
	if l == nil || l.handle == nil {
		return 0, ErrClosed
	}
	value, err := l.handle.Count()
	return value, publicError(err)
}

// Values copies azimuth/elevation degrees and range kilometers from C.
func (l *LookAngles) Values() ([]Topocentric, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	values, err := l.handle.Values()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]Topocentric, len(values))
	for i, value := range values {
		out[i] = Topocentric{
			AzimuthDeg:   value.AzimuthDeg,
			ElevationDeg: value.ElevationDeg,
			RangeKm:      value.RangeKm,
		}
	}
	return out, nil
}
