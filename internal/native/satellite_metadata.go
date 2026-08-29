//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"time"
	"unsafe"
)

type SatelliteConstellation struct {
	_      noCopy
	handle *positioningHandle
}
type ConstellationLookAngleArcs struct {
	_      noCopy
	handle *positioningHandle
}
type ConstellationGroundTracks struct {
	_      noCopy
	handle *positioningHandle
}
type FleetPass struct {
	SatelliteIndex int
	Pass           SatellitePass
}
type ConstellationPasses struct {
	_      noCopy
	handle *positioningHandle
}

func releaseSatelliteConstellation(p unsafe.Pointer) {
	C.sidereon_satellite_constellation_free((*C.SidereonSatelliteConstellation)(p))
}
func releaseArcs(p unsafe.Pointer) {
	C.sidereon_satellite_constellation_look_angles_free((*C.SidereonSatelliteConstellationLookAngles)(p))
}
func releaseTracks(p unsafe.Pointer) {
	C.sidereon_satellite_constellation_ground_tracks_free((*C.SidereonSatelliteConstellationGroundTracks)(p))
}
func releaseFleetPasses(p unsafe.Pointer) {
	C.sidereon_satellite_constellation_passes_free((*C.SidereonSatelliteConstellationPasses)(p))
}

func withTLEPointers(tles []*TLE, fn func(**C.SidereonTle) error) error {
	if _, err := checkedNativeAllocationSize(len(tles), unsafe.Sizeof(uintptr(0))); err != nil {
		return err
	}
	var memory unsafe.Pointer
	if len(tles) > 0 {
		memory = C.malloc(C.size_t(len(tles)) * C.size_t(unsafe.Sizeof(uintptr(0))))
		if memory == nil {
			return errors.New("sidereon: unable to allocate satellite pointer array")
		}
		defer C.free(memory)
	}
	ptrs := unsafe.Slice((**C.SidereonTle)(memory), len(tles))
	handles := make([]*positioningHandle, len(tles))
	for i, tle := range tles {
		if tle == nil || tle.handle == nil {
			return ErrClosed
		}
		handles[i] = tle.handle
	}
	return withPositioningHandles(handles, func(pointers []unsafe.Pointer) error {
		for i, pointer := range pointers {
			ptrs[i] = (*C.SidereonTle)(pointer)
		}
		return fn((**C.SidereonTle)(memory))
	})
}
func BuildSatelliteConstellation(tles []*TLE) (*SatelliteConstellation, error) {
	if len(tles) == 0 {
		return nil, invalidArgument("satellite constellation is empty")
	}
	var out *C.SidereonSatelliteConstellation
	e := withTLEPointers(tles, func(p **C.SidereonTle) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_satellite_constellation_build(p, C.size_t(len(tles)), &out)) })
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_satellite_constellation_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native satellite constellation returned no handle")
	}
	return &SatelliteConstellation{handle: newPositioningHandle(unsafe.Pointer(out), releaseSatelliteConstellation)}, nil
}
func (s *SatelliteConstellation) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}
func (s *SatelliteConstellation) Count() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	e := s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_satellite_count((*C.SidereonSatelliteConstellation)(p), &n))
		})
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(n, "satellite constellation count")
}
func (s *SatelliteConstellation) CatalogNumber(index int) (string, error) {
	if s == nil || s.handle == nil {
		return "", ErrClosed
	}
	if index < 0 {
		return "", invalidArgument("negative satellite index")
	}
	var result string
	e := s.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var required C.size_t
			status := C.sidereon_satellite_constellation_catalog_number((*C.SidereonSatelliteConstellation)(p), C.size_t(index), nil, 0, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			n, err := sizeTToInt(required, "catalog number length")
			if err != nil {
				return err
			}
			if n < 1 {
				return errors.New("sidereon: invalid native catalog number length")
			}
			if _, err := checkedNativeAllocationSize(n, 1); err != nil {
				return err
			}
			buffer := C.malloc(C.size_t(n))
			if buffer == nil {
				return errors.New("sidereon: unable to allocate catalog number")
			}
			defer C.free(buffer)
			var written C.size_t
			status = C.sidereon_satellite_constellation_catalog_number((*C.SidereonSatelliteConstellation)(p), C.size_t(index), (*C.char)(buffer), C.size_t(n), &written)
			if err = statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			if written != required {
				return errors.New("sidereon: native catalog number size changed")
			}
			result = C.GoString((*C.char)(buffer))
			return nil
		})
	})
	return result, e
}
func epochsToC(values []int64) (unsafe.Pointer, error) {
	if _, e := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.int64_t(0))); e != nil {
		return nil, e
	}
	if len(values) == 0 {
		return nil, nil
	}
	memory := C.malloc(C.size_t(len(values)) * C.size_t(unsafe.Sizeof(C.int64_t(0))))
	if memory == nil {
		return nil, errors.New("sidereon: unable to allocate epoch array")
	}
	out := unsafe.Slice((*C.int64_t)(memory), len(values))
	for i, value := range values {
		out[i] = C.int64_t(value)
	}
	return memory, nil
}

func satelliteEpochsJ2000(values []int64) ([]float64, error) {
	out := make([]float64, len(values))
	var err error
	withCThread(func() {
		for i, value := range values {
			tm := time.UnixMicro(value).UTC()
			out[i], err = civilJ2000Locked(tm)
			if err != nil {
				return
			}
		}
	})
	return out, err
}
func (s *SatelliteConstellation) Propagate(epochs []int64, parallel bool) (*TLEBatchPropagation, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	raw, e := epochsToC(epochs)
	if e != nil {
		return nil, e
	}
	defer C.free(raw)
	p := (*C.int64_t)(raw)
	var out *C.SidereonTleBatchPropagation
	e = s.handle.with(func(q unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_propagate((*C.SidereonSatelliteConstellation)(q), p, C.size_t(len(epochs)), C.bool(parallel), &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_tle_batch_propagation_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native batch propagation returned no handle")
	}
	epochJ2000, err := satelliteEpochsJ2000(epochs)
	if err != nil {
		withCThread(func() { C.sidereon_tle_batch_propagation_free(out) })
		return nil, err
	}
	return &TLEBatchPropagation{handle: newPositioningHandle(unsafe.Pointer(out), releaseTLEBatchPropagation), epochJ2000: epochJ2000}, nil
}
func (s *SatelliteConstellation) Visible(station GroundStation, epochUnixUS int64, minElevationDeg float64) (*VisibleList, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	st := cGroundStation(station)
	var out *C.SidereonVisibleList
	e := s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_visible((*C.SidereonSatelliteConstellation)(p), &st, C.int64_t(epochUnixUS), C.double(minElevationDeg), &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_visible_list_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native visible list returned no handle")
	}
	return &VisibleList{handle: newPositioningHandle(unsafe.Pointer(out), releaseVisibleList)}, nil
}
func (s *SatelliteConstellation) LookAngles(station GroundStation, epochs []int64, parallel bool) (*ConstellationLookAngleArcs, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	raw, e := epochsToC(epochs)
	if e != nil {
		return nil, e
	}
	defer C.free(raw)
	ep := (*C.int64_t)(raw)
	st := cGroundStation(station)
	var out *C.SidereonSatelliteConstellationLookAngles
	e = s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_look_angle_arcs((*C.SidereonSatelliteConstellation)(p), &st, ep, C.size_t(len(epochs)), C.bool(parallel), &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_satellite_constellation_look_angles_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native look-angle arcs returned no handle")
	}
	return &ConstellationLookAngleArcs{handle: newPositioningHandle(unsafe.Pointer(out), releaseArcs)}, nil
}
func (s *SatelliteConstellation) GroundTracks(epochs []int64) (*ConstellationGroundTracks, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	raw, e := epochsToC(epochs)
	if e != nil {
		return nil, e
	}
	defer C.free(raw)
	ep := (*C.int64_t)(raw)
	var out *C.SidereonSatelliteConstellationGroundTracks
	e = s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_ground_tracks((*C.SidereonSatelliteConstellation)(p), ep, C.size_t(len(epochs)), &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_satellite_constellation_ground_tracks_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native ground tracks returned no handle")
	}
	return &ConstellationGroundTracks{handle: newPositioningHandle(unsafe.Pointer(out), releaseTracks)}, nil
}
func (a *ConstellationLookAngleArcs) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}
func (a *ConstellationLookAngleArcs) SatelliteCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	e := a.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_look_angles_satellite_count((*C.SidereonSatelliteConstellationLookAngles)(p), &n))
		})
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(n, "look-angle satellite count")
}
func (a *ConstellationLookAngleArcs) ArcLengths() ([]int, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	return arcLengths(a.handle,
		func(p unsafe.Pointer, n *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_satellite_constellation_look_angles_satellite_count((*C.SidereonSatelliteConstellationLookAngles)(p), n)
		},
		func(p unsafe.Pointer, i C.size_t, n *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_satellite_constellation_look_angles_arc_len((*C.SidereonSatelliteConstellationLookAngles)(p), i, n)
		})
}
func (a *ConstellationLookAngleArcs) Values() ([]LookAngle, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonLookAngle
	e := copyLookAngles(a.handle, &raw, func(p unsafe.Pointer, o *C.SidereonLookAngle, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_satellite_constellation_look_angles_values((*C.SidereonSatelliteConstellationLookAngles)(p), o, n, w, r)
	})
	if e != nil {
		return nil, e
	}
	out := make([]LookAngle, len(raw))
	for j, v := range raw {
		out[j] = LookAngle{AzimuthDeg: float64(v.azimuth_deg), ElevationDeg: float64(v.elevation_deg), RangeKm: float64(v.range_km)}
	}
	return out, nil
}
func (g *ConstellationGroundTracks) Close() error {
	if g == nil || g.handle == nil {
		return nil
	}
	return g.handle.close()
}
func (g *ConstellationGroundTracks) SatelliteCount() (int, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	e := g.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_ground_tracks_satellite_count((*C.SidereonSatelliteConstellationGroundTracks)(p), &n))
		})
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(n, "ground-track satellite count")
}
func (g *ConstellationGroundTracks) TrackLengths() ([]int, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	return arcLengths(g.handle,
		func(p unsafe.Pointer, n *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_satellite_constellation_ground_tracks_satellite_count((*C.SidereonSatelliteConstellationGroundTracks)(p), n)
		},
		func(p unsafe.Pointer, i C.size_t, n *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_satellite_constellation_ground_tracks_track_len((*C.SidereonSatelliteConstellationGroundTracks)(p), i, n)
		})
}
func (g *ConstellationGroundTracks) Values() ([]Geodetic, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonGeodetic
	e := copyGeodeticValues(g.handle, &raw, func(p unsafe.Pointer, o *C.SidereonGeodetic, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_satellite_constellation_ground_tracks_values((*C.SidereonSatelliteConstellationGroundTracks)(p), o, n, w, r)
	})
	if e != nil {
		return nil, e
	}
	out := make([]Geodetic, len(raw))
	for j, v := range raw {
		out[j] = geodeticFromC(v)
	}
	return out, nil
}
func arcLengths(h *positioningHandle, countCall func(unsafe.Pointer, *C.size_t) C.enum_SidereonStatus, lenCall func(unsafe.Pointer, C.size_t, *C.size_t) C.enum_SidereonStatus) ([]int, error) {
	if h == nil {
		return nil, ErrClosed
	}
	var count C.size_t
	e := h.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			s := countCall(p, &count)
			return statusErrorLocked(uint32(s))
		})
	})
	if e != nil {
		return nil, e
	}
	n, e := sizeTToInt(count, "arc satellite count")
	if e != nil {
		return nil, e
	}
	out := make([]int, n)
	for j := range out {
		var length C.size_t
		e = h.with(func(p unsafe.Pointer) error {
			return withCThreadError(func() error {
				return statusErrorLocked(uint32(lenCall(p, C.size_t(j), &length)))
			})
		})
		if e != nil {
			return nil, e
		}
		out[j], e = sizeTToInt(length, "arc length")
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
func copyLookAngles(h *positioningHandle, out *[]C.SidereonLookAngle, call func(unsafe.Pointer, *C.SidereonLookAngle, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) error {
	var w, r C.size_t
	e := h.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			s := call(p, nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("look-angle values", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			mem, x := allocNativeArray(n, unsafe.Sizeof(C.SidereonLookAngle{}))
			if x != nil {
				return x
			}
			defer C.free(mem)
			nativeValues := unsafe.Slice((*C.SidereonLookAngle)(mem), n)
			w, r = 0, 0
			var q *C.SidereonLookAngle
			if n > 0 {
				q = &nativeValues[0]
			}
			s = call(p, q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("look-angle values", n, n, uint64(w), uint64(r))
			if x == nil {
				*out = append([]C.SidereonLookAngle(nil), nativeValues...)
			}
			return x
		})
	})
	return e
}
func copyGeodeticValues(h *positioningHandle, out *[]C.SidereonGeodetic, call func(unsafe.Pointer, *C.SidereonGeodetic, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) error {
	var w, r C.size_t
	return h.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			s := call(p, nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("ground-track values", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			mem, x := allocNativeArray(n, unsafe.Sizeof(C.SidereonGeodetic{}))
			if x != nil {
				return x
			}
			defer C.free(mem)
			nativeValues := unsafe.Slice((*C.SidereonGeodetic)(mem), n)
			w, r = 0, 0
			var q *C.SidereonGeodetic
			if n > 0 {
				q = &nativeValues[0]
			}
			s = call(p, q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("ground-track values", n, n, uint64(w), uint64(r))
			if x == nil {
				*out = append([]C.SidereonGeodetic(nil), nativeValues...)
			}
			return x
		})
	})
}

func (s *SatelliteConstellation) Passes(station GroundStation, startUnixUS, endUnixUS int64, options *PassFinderOptions) (*ConstellationPasses, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	st := cGroundStation(station)
	var co C.SidereonPassFinderOptions
	var cop *C.SidereonPassFinderOptions
	if options != nil {
		co = cPassFinderOptions(*options)
		cop = &co
	}
	var out *C.SidereonSatelliteConstellationPasses
	e := s.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_passes((*C.SidereonSatelliteConstellation)(p), &st, C.int64_t(startUnixUS), C.int64_t(endUnixUS), cop, &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_satellite_constellation_passes_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native constellation passes returned no handle")
	}
	return &ConstellationPasses{handle: newPositioningHandle(unsafe.Pointer(out), releaseFleetPasses)}, nil
}
func (p *ConstellationPasses) Close() error {
	if p == nil || p.handle == nil {
		return nil
	}
	return p.handle.close()
}
func (p *ConstellationPasses) Count() (int, error) {
	if p == nil || p.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	e := p.handle.with(func(q unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_satellite_constellation_passes_count((*C.SidereonSatelliteConstellationPasses)(q), &n))
		})
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(n, "constellation pass count")
}
func (p *ConstellationPasses) Values() ([]FleetPass, error) {
	if p == nil || p.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonFleetPass
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := p.handle.with(func(q unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			s := C.sidereon_satellite_constellation_passes_values((*C.SidereonSatelliteConstellationPasses)(q), nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("fleet passes", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.SidereonFleetPass{}))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.SidereonFleetPass)(memory), n)
			w, r = 0, 0
			var out *C.SidereonFleetPass
			if n > 0 {
				out = &raw[0]
			}
			s = C.sidereon_satellite_constellation_passes_values((*C.SidereonSatelliteConstellationPasses)(q), out, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("fleet passes", n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]FleetPass, len(raw))
	for j, v := range raw {
		satelliteIndex, err := sizeTToInt(v.satellite_index, "fleet pass satellite index")
		if err != nil {
			return nil, err
		}
		out[j] = FleetPass{SatelliteIndex: satelliteIndex, Pass: SatellitePass{AOS: time.UnixMicro(int64(v.pass.aos_unix_us)).UTC(), LOS: time.UnixMicro(int64(v.pass.los_unix_us)).UTC(), Culmination: time.UnixMicro(int64(v.pass.culmination_unix_us)).UTC(), MaxElevationDeg: float64(v.pass.max_elevation_deg), DurationS: float64(v.pass.duration_s)}}
	}
	return out, nil
}

func SatelliteVisualMagnitude(rangeKm, phaseAngleDeg, standardMagnitude, referenceRangeKm float64) (float64, error) {
	var out C.double
	e := callStatus(func() uint32 {
		return uint32(C.sidereon_satellite_visual_magnitude(C.double(rangeKm), C.double(phaseAngleDeg), C.double(standardMagnitude), C.double(referenceRangeKm), &out))
	})
	return float64(out), e
}
