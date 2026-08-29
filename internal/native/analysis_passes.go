//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"time"
	"unsafe"
)

// GroundStation is a WGS84 station. Latitude and longitude are degrees and
// altitude is meters.
type GroundStation struct {
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

// LookAngle contains a topocentric look-angle result. Angles are degrees and
// range is kilometers.
type LookAngle struct {
	AzimuthDeg   float64
	ElevationDeg float64
	RangeKm      float64
}

// PassFinderOptions controls C pass finding. The elevation mask is degrees;
// step and time tolerance are seconds.
type PassFinderOptions struct {
	ElevationMaskDeg     float64
	StepSeconds          float64
	TimeToleranceSeconds float64
}

type PassList struct{ handle *positioningHandle }
type GroundTrack struct{ handle *positioningHandle }
type LookAngles struct{ handle *positioningHandle }

func cGroundStation(value GroundStation) C.SidereonGroundStation {
	return C.SidereonGroundStation{
		latitude_deg:  C.double(value.LatitudeDeg),
		longitude_deg: C.double(value.LongitudeDeg),
		altitude_m:    C.double(value.AltitudeM),
	}
}

func cPassFinderOptions(value PassFinderOptions) C.SidereonPassFinderOptions {
	return C.SidereonPassFinderOptions{
		elevation_mask_deg: C.double(value.ElevationMaskDeg),
		step_seconds:       C.double(value.StepSeconds),
		time_tolerance_s:   C.double(value.TimeToleranceSeconds),
	}
}

func releasePassList(pointer unsafe.Pointer) {
	C.sidereon_pass_list_free((*C.SidereonPassList)(pointer))
}

func releaseGroundTrack(pointer unsafe.Pointer) {
	C.sidereon_ground_track_free((*C.SidereonGroundTrack)(pointer))
}

func releaseLookAngles(pointer unsafe.Pointer) {
	C.sidereon_look_angles_free((*C.SidereonLookAngles)(pointer))
}

func unixEpochs(values []time.Time) ([]C.int64_t, error) {
	if _, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.int64_t(0))); err != nil {
		return nil, err
	}
	converted, err := unixMicrosecondsSlice(values)
	if err != nil {
		return nil, err
	}
	out := make([]C.int64_t, len(converted))
	for i, value := range converted {
		out[i] = C.int64_t(value)
	}
	return out, nil
}

// FindPasses returns an owning C pass list for the half-open UTC interval
// [start, end). C receives the times as Unix microseconds.
func (t *TLE) FindPasses(station GroundStation, start, end time.Time, options *PassFinderOptions) (*PassList, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	startUnixUS, err := unixMicroseconds(start)
	if err != nil {
		return nil, err
	}
	endUnixUS, err := unixMicroseconds(end)
	if err != nil {
		return nil, err
	}
	stationC := cGroundStation(station)
	var optionsC C.SidereonPassFinderOptions
	var optionsPointer *C.SidereonPassFinderOptions
	if options != nil {
		optionsC = cPassFinderOptions(*options)
		optionsPointer = &optionsC
	}
	var out *C.SidereonPassList
	err = t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_find_passes(
				(*C.SidereonTle)(pointer), &stationC,
				C.int64_t(startUnixUS), C.int64_t(endUnixUS),
				optionsPointer, &out,
			)
		})
	})
	runtime.KeepAlive(t)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_pass_list_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("pass search")
	}
	return &PassList{handle: newPositioningHandle(unsafe.Pointer(out), releasePassList)}, nil
}

// GroundTrack propagates the TLE at UTC epochs with microsecond precision.
func (t *TLE) GroundTrack(epochs []time.Time) (*GroundTrack, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	cEpochs, err := unixEpochs(epochs)
	if err != nil {
		return nil, err
	}
	var epochPointer *C.int64_t
	if len(cEpochs) > 0 {
		epochPointer = &cEpochs[0]
	}
	var out *C.SidereonGroundTrack
	err = t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_ground_track((*C.SidereonTle)(pointer), epochPointer, C.size_t(len(cEpochs)), &out)
		})
	})
	runtime.KeepAlive(t)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_ground_track_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("ground-track computation")
	}
	return &GroundTrack{handle: newPositioningHandle(unsafe.Pointer(out), releaseGroundTrack)}, nil
}

// LookAngles computes topocentric angles at UTC epochs with microsecond
// precision. Angles are degrees and range is kilometers.
func (t *TLE) LookAngles(station GroundStation, epochs []time.Time) (*LookAngles, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	cEpochs, err := unixEpochs(epochs)
	if err != nil {
		return nil, err
	}
	stationC := cGroundStation(station)
	var epochPointer *C.int64_t
	if len(cEpochs) > 0 {
		epochPointer = &cEpochs[0]
	}
	var out *C.SidereonLookAngles
	err = t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_tle_look_angles((*C.SidereonTle)(pointer), &stationC, epochPointer, C.size_t(len(cEpochs)), &out)
		})
	})
	runtime.KeepAlive(t)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_look_angles_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("look-angle computation")
	}
	return &LookAngles{handle: newPositioningHandle(unsafe.Pointer(out), releaseLookAngles)}, nil
}

// PassFinderDefaults returns the native pass-finder defaults.
func PassFinderDefaults() (PassFinderOptions, error) {
	var out C.SidereonPassFinderOptions
	err := callStatus(func() uint32 { return C.sidereon_pass_finder_options_init(&out) })
	return PassFinderOptions{
		ElevationMaskDeg:     float64(out.elevation_mask_deg),
		StepSeconds:          float64(out.step_seconds),
		TimeToleranceSeconds: float64(out.time_tolerance_s),
	}, err
}

// Close releases the native pass list. It is idempotent and safe with Values.
func (p *PassList) Close() error {
	if p == nil {
		return nil
	}
	return p.handle.close()
}

// Close releases the native ground-track result. It is idempotent and safe
// with Values.
func (g *GroundTrack) Close() error {
	if g == nil {
		return nil
	}
	return g.handle.close()
}

// Close releases the native look-angle result. It is idempotent and safe with
// Values.
func (l *LookAngles) Close() error {
	if l == nil {
		return nil
	}
	return l.handle.close()
}

// Values copies pass rows into Go-owned memory.
func (p *PassList) Values() ([]SatellitePass, error) {
	if p == nil || p.handle == nil {
		return nil, ErrClosed
	}
	var values []C.SidereonSatellitePass
	err := p.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_pass_list_values((*C.SidereonPassList)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		expected, err := validateNativeQuery("pass rows", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonSatellitePass{})); err != nil {
			return err
		}
		values = make([]C.SidereonSatellitePass, expected)
		var output *C.SidereonSatellitePass
		if len(values) > 0 {
			output = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 {
			return C.sidereon_pass_list_values((*C.SidereonPassList)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		count, err := validateTwoPassCounts("pass rows", len(values), expected, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		values = values[:count]
		return nil
	})
	runtime.KeepAlive(p)
	if err != nil {
		return nil, err
	}
	out := make([]SatellitePass, len(values))
	for i, value := range values {
		out[i] = SatellitePass{
			AOS:             time.UnixMicro(int64(value.aos_unix_us)).UTC(),
			LOS:             time.UnixMicro(int64(value.los_unix_us)).UTC(),
			Culmination:     time.UnixMicro(int64(value.culmination_unix_us)).UTC(),
			MaxElevationDeg: float64(value.max_elevation_deg),
			DurationS:       float64(value.duration_s),
		}
	}
	return out, nil
}

// Count returns the number of pass rows.
func (p *PassList) Count() (int, error) {
	if p == nil || p.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := p.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_pass_list_count((*C.SidereonPassList)(pointer), &count) })
	})
	runtime.KeepAlive(p)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

// Values copies ground-track points into Go-owned memory.
func (g *GroundTrack) Values() ([]Geodetic, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	var values []C.SidereonGeodetic
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_ground_track_values((*C.SidereonGroundTrack)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		expected, err := validateNativeQuery("ground-track values", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonGeodetic{})); err != nil {
			return err
		}
		values = make([]C.SidereonGeodetic, expected)
		var output *C.SidereonGeodetic
		if len(values) > 0 {
			output = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 {
			return C.sidereon_ground_track_values((*C.SidereonGroundTrack)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		count, err := validateTwoPassCounts("ground-track values", len(values), expected, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		values = values[:count]
		return nil
	})
	runtime.KeepAlive(g)
	if err != nil {
		return nil, err
	}
	out := make([]Geodetic, len(values))
	for i, value := range values {
		out[i] = geodeticFromC(value)
	}
	return out, nil
}

// Count returns the number of ground-track points.
func (g *GroundTrack) Count() (int, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_ground_track_count((*C.SidereonGroundTrack)(pointer), &count) })
	})
	runtime.KeepAlive(g)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

// Values copies look-angle rows into Go-owned memory.
func (l *LookAngles) Values() ([]LookAngle, error) {
	if l == nil || l.handle == nil {
		return nil, ErrClosed
	}
	var values []C.SidereonLookAngle
	err := l.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_look_angles_values((*C.SidereonLookAngles)(pointer), nil, 0, &written, &required)
		}); err != nil {
			return err
		}
		expected, err := validateNativeQuery("look-angle values", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonLookAngle{})); err != nil {
			return err
		}
		values = make([]C.SidereonLookAngle, expected)
		var output *C.SidereonLookAngle
		if len(values) > 0 {
			output = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 {
			return C.sidereon_look_angles_values((*C.SidereonLookAngles)(pointer), output, C.size_t(len(values)), &written, &required)
		}); err != nil {
			return err
		}
		count, err := validateTwoPassCounts("look-angle values", len(values), expected, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		values = values[:count]
		return nil
	})
	runtime.KeepAlive(l)
	if err != nil {
		return nil, err
	}
	out := make([]LookAngle, len(values))
	for i, value := range values {
		out[i] = LookAngle{
			AzimuthDeg:   float64(value.azimuth_deg),
			ElevationDeg: float64(value.elevation_deg),
			RangeKm:      float64(value.range_km),
		}
	}
	return out, nil
}

// Count returns the number of look-angle rows.
func (l *LookAngles) Count() (int, error) {
	if l == nil || l.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := l.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_look_angles_epoch_count((*C.SidereonLookAngles)(pointer), &count) })
	})
	runtime.KeepAlive(l)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}
