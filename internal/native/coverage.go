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
	"runtime"
	"unsafe"
)

// CoverageLookAngle is one copied coverage-grid cell. A false OK means C
// could not produce a look angle for that satellite/station pair.
type CoverageLookAngle struct {
	OK           bool
	AzimuthDeg   float64
	ElevationDeg float64
	RangeKm      float64
}

// CoverageGrid owns the C look-angle grid and must not be copied after first
// use. Queries may run concurrently with one another and with Close.
type CoverageGrid struct {
	_      noCopy
	handle *positioningHandle
}

func releaseCoverageGrid(pointer unsafe.Pointer) {
	C.sidereon_coverage_grid_free((*C.SidereonCoverageGrid)(pointer))
}

func cCoverageStations(stations []GroundStation) (unsafe.Pointer, *C.SidereonGroundStation, error) {
	memory, err := allocNativeArray(len(stations), unsafe.Sizeof(C.SidereonGroundStation{}))
	if err != nil {
		return nil, nil, err
	}
	if memory == nil {
		return nil, nil, nil
	}
	values := unsafe.Slice((*C.SidereonGroundStation)(memory), len(stations))
	for i, station := range stations {
		values[i] = cGroundStation(station)
	}
	return memory, &values[0], nil
}

// CoverageLookAngles builds a one-epoch satellite/station grid. C performs
// all propagation and look-angle calculations.
func CoverageLookAngles(tles []*TLE, stations []GroundStation, epochUnixUS int64) (*CoverageGrid, error) {
	if len(tles) == 0 {
		return nil, invalidArgument("coverage satellite list is empty")
	}
	if len(stations) == 0 {
		return nil, invalidArgument("coverage station list is empty")
	}
	tlesCount, err := checkedNativeSize(len(tles))
	if err != nil {
		return nil, err
	}
	stationsCount, err := checkedNativeSize(len(stations))
	if err != nil {
		return nil, err
	}
	var out *C.SidereonCoverageGrid
	err = withTLEPointers(tles, func(tlePointers **C.SidereonTle) error {
		var memory unsafe.Pointer
		var stationPointer *C.SidereonGroundStation
		var allocErr error
		withCThread(func() {
			memory, stationPointer, allocErr = cCoverageStations(stations)
		})
		if allocErr != nil {
			return allocErr
		}
		if memory != nil {
			defer C.free(memory)
		}
		return callStatus(func() uint32 {
			return uint32(C.sidereon_coverage_look_angles(
				tlePointers, tlesCount, stationPointer, stationsCount, C.int64_t(epochUnixUS), &out,
			))
		})
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_coverage_grid_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, errors.New("sidereon: native coverage constructor returned no handle")
	}
	return &CoverageGrid{handle: newPositioningHandle(unsafe.Pointer(out), releaseCoverageGrid)}, nil
}

func (g *CoverageGrid) Close() error {
	if g == nil || g.handle == nil {
		return nil
	}
	return g.handle.close()
}

func (g *CoverageGrid) Dimensions() (satellites, stations int, err error) {
	if g == nil || g.handle == nil {
		return 0, 0, ErrClosed
	}
	var satCount, stationCount C.size_t
	err = g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_coverage_grid_dimensions((*C.SidereonCoverageGrid)(pointer), &satCount, &stationCount))
		})
	})
	runtime.KeepAlive(g)
	if err != nil {
		return 0, 0, err
	}
	satellites, err = sizeTToInt(satCount, "coverage satellite count")
	if err != nil {
		return 0, 0, err
	}
	stations, err = sizeTToInt(stationCount, "coverage station count")
	return satellites, stations, err
}

func (g *CoverageGrid) LookAngle(satelliteIndex, stationIndex int) (CoverageLookAngle, error) {
	if g == nil || g.handle == nil {
		return CoverageLookAngle{}, ErrClosed
	}
	if satelliteIndex < 0 || stationIndex < 0 {
		return CoverageLookAngle{}, invalidArgument("coverage grid indices must not be negative")
	}
	sat, err := checkedNativeSize(satelliteIndex)
	if err != nil {
		return CoverageLookAngle{}, err
	}
	station, err := checkedNativeSize(stationIndex)
	if err != nil {
		return CoverageLookAngle{}, err
	}
	var out C.SidereonCoverageLookAngle
	err = g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_coverage_grid_look_angle((*C.SidereonCoverageGrid)(pointer), sat, station, &out))
		})
	})
	runtime.KeepAlive(g)
	return CoverageLookAngle{OK: bool(out.ok), AzimuthDeg: float64(out.azimuth_deg), ElevationDeg: float64(out.elevation_deg), RangeKm: float64(out.range_km)}, err
}

func (g *CoverageGrid) AccessCounts(minElevationDeg float64) ([]int, error) {
	return coverageGridIntValues(g, "coverage access counts", func(pointer unsafe.Pointer, out *C.size_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_coverage_grid_access_counts((*C.SidereonCoverageGrid)(pointer), C.double(minElevationDeg), out, length, written, required)
	})
}

func (g *CoverageGrid) MaxElevationDeg() ([]float64, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.double
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(out *C.double, length C.size_t) C.enum_SidereonStatus {
				return C.sidereon_coverage_grid_max_elevation_deg((*C.SidereonCoverageGrid)(pointer), out, length, &written, &required)
			}
			if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
				return err
			}
			n, err := validateNativeQuery("coverage maximum elevations", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err := allocNativeArray(n, unsafe.Sizeof(C.double(0)))
			if err != nil {
				return err
			}
			if memory != nil {
				defer C.free(memory)
			}
			if n > 0 {
				raw = append(raw, unsafe.Slice((*C.double)(memory), n)...)
			}
			written, required = 0, 0
			var out *C.double
			if n > 0 {
				out = (*C.double)(memory)
			}
			if err := statusErrorLocked(uint32(call(out, C.size_t(n)))); err != nil {
				return err
			}
			if _, err := validateTwoPassCounts("coverage maximum elevations", n, n, uint64(written), uint64(required)); err != nil {
				return err
			}
			return nil
		})
	})
	runtime.KeepAlive(g)
	if err != nil {
		return nil, err
	}
	out := make([]float64, len(raw))
	for i := range raw {
		out[i] = float64(raw[i])
	}
	return out, nil
}

func (g *CoverageGrid) VisibleMask(minElevationDeg float64) ([]bool, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.bool
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(out *C.bool, length C.size_t) C.enum_SidereonStatus {
				return C.sidereon_coverage_grid_visible_mask((*C.SidereonCoverageGrid)(pointer), C.double(minElevationDeg), out, length, &written, &required)
			}
			if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
				return err
			}
			n, err := validateNativeQuery("coverage visible mask", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err := allocNativeArray(n, unsafe.Sizeof(C.bool(false)))
			if err != nil {
				return err
			}
			if memory != nil {
				defer C.free(memory)
			}
			if n > 0 {
				raw = append(raw, unsafe.Slice((*C.bool)(memory), n)...)
			}
			written, required = 0, 0
			var out *C.bool
			if n > 0 {
				out = (*C.bool)(memory)
			}
			if err := statusErrorLocked(uint32(call(out, C.size_t(n)))); err != nil {
				return err
			}
			_, err = validateTwoPassCounts("coverage visible mask", n, n, uint64(written), uint64(required))
			return err
		})
	})
	runtime.KeepAlive(g)
	if err != nil {
		return nil, err
	}
	out := make([]bool, len(raw))
	for i := range raw {
		out[i] = bool(raw[i])
	}
	return out, nil
}

func coverageGridIntValues(g *CoverageGrid, label string, call func(unsafe.Pointer, *C.size_t, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]int, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.size_t
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			if err := statusErrorLocked(uint32(call(pointer, nil, 0, &written, &required))); err != nil {
				return err
			}
			n, err := validateNativeQuery(label, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err := allocNativeArray(n, unsafe.Sizeof(C.size_t(0)))
			if err != nil {
				return err
			}
			if memory != nil {
				defer C.free(memory)
			}
			written, required = 0, 0
			var out *C.size_t
			if n > 0 {
				out = (*C.size_t)(memory)
			}
			if err := statusErrorLocked(uint32(call(pointer, out, C.size_t(n), &written, &required))); err != nil {
				return err
			}
			count, err := validateTwoPassCounts(label, n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if count > 0 {
				raw = append(raw, unsafe.Slice((*C.size_t)(memory), count)...)
			}
			return nil
		})
	})
	runtime.KeepAlive(g)
	if err != nil {
		return nil, err
	}
	out := make([]int, len(raw))
	for i, value := range raw {
		out[i], err = sizeTToInt(value, label+" value")
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
