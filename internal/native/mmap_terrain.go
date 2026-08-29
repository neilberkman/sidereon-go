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

type MmapTerrainHeightResult struct {
	Status                uint32
	HasOrthometricHeightM bool
	OrthometricHeightM    float64
}

type TerrainStoreTileIndex struct {
	LatIndex        int32
	LonIndex        int32
	MinLongitudeDeg float64
	MinLatitudeDeg  float64
	MaxLongitudeDeg float64
	MaxLatitudeDeg  float64
	LonCount        uint32
	LatCount        uint32
	DataOffset      uint64
	DataLen         uint64
	Checksum64      uint64
	VerticalDatum   uint32
}

type MmapTerrain struct {
	handle *positioningHandle
}

func mmapTerrainFromPointer(pointer *C.SidereonMmapTerrain) (*MmapTerrain, error) {
	if pointer == nil {
		return nil, errors.New("sidereon: native memory-mappable terrain constructor returned no handle")
	}
	return &MmapTerrain{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_mmap_terrain_free((*C.SidereonMmapTerrain)(value))
	})}, nil
}

func mmapTerrainFromBytes(data []byte, fromVec bool) (*MmapTerrain, error) {
	var pointer *C.SidereonMmapTerrain
	err := withInputTerrainDiagnostics(data, func(input *C.uint8_t, length C.size_t) uint32 {
		if fromVec {
			return C.sidereon_mmap_terrain_from_vec(input, length, &pointer)
		}
		return C.sidereon_mmap_terrain_from_bytes(input, length, &pointer)
	}, false, true)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_mmap_terrain_free(pointer) })
		}
		return nil, err
	}
	return mmapTerrainFromPointer(pointer)
}

func MmapTerrainFromBytes(data []byte) (*MmapTerrain, error) {
	return mmapTerrainFromBytes(data, false)
}

func MmapTerrainFromVec(data []byte) (*MmapTerrain, error) {
	return mmapTerrainFromBytes(data, true)
}

func MmapTerrainFromPath(path string) (*MmapTerrain, error) {
	input, err := cString(path)
	if err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(input))
	var pointer *C.SidereonMmapTerrain
	err = callStatusWithTerrainDiagnostics(func() uint32 {
		return C.sidereon_mmap_terrain_from_path(input, &pointer)
	}, false, true)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_mmap_terrain_free(pointer) })
		}
		return nil, err
	}
	return mmapTerrainFromPointer(pointer)
}

func MmapTerrainFromPathAttested(path string, checksum64 uint64) (*MmapTerrain, error) {
	input, err := cString(path)
	if err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(input))
	var pointer *C.SidereonMmapTerrain
	err = callStatusWithTerrainDiagnostics(func() uint32 {
		return C.sidereon_mmap_terrain_from_path_attested(input, C.uint64_t(checksum64), &pointer)
	}, false, true)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_mmap_terrain_free(pointer) })
		}
		return nil, err
	}
	return mmapTerrainFromPointer(pointer)
}

func (t *MmapTerrain) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return t.handle.close()
}

func (t *MmapTerrain) Checksum64() (uint64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output C.uint64_t
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_mmap_terrain_checksum64((*C.SidereonMmapTerrain)(pointer), &output)
		})
	})
	runtime.KeepAlive(t)
	return uint64(output), err
}

func (t *MmapTerrain) DigestProvenance() (uint32, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output uint32
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_mmap_terrain_digest_provenance((*C.SidereonMmapTerrain)(pointer), &output)
		})
	})
	runtime.KeepAlive(t)
	return uint32(output), err
}

func (t *MmapTerrain) HeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output C.SidereonOrthometricHeightM
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_mmap_terrain_height_m((*C.SidereonMmapTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &output)
		})
	})
	runtime.KeepAlive(t)
	return float64(output.value_m), err
}

func (t *MmapTerrain) HeightMWithOptions(longitudeDeg, latitudeDeg float64, options DtedLookupOptions) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	cOptions := cDtedOptions(options)
	var output C.SidereonOrthometricHeightM
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_mmap_terrain_height_m_with_options((*C.SidereonMmapTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &cOptions, &output)
		})
	})
	runtime.KeepAlive(t)
	return float64(output.value_m), err
}

func (t *MmapTerrain) OrthometricHeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output C.SidereonOrthometricHeightM
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_mmap_terrain_orthometric_height_m((*C.SidereonMmapTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &output)
		})
	})
	runtime.KeepAlive(t)
	return float64(output.value_m), err
}

func (t *MmapTerrain) OrthometricHeightMWithOptions(longitudeDeg, latitudeDeg float64, options DtedLookupOptions) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	cOptions := cDtedOptions(options)
	var output C.SidereonOrthometricHeightM
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_mmap_terrain_orthometric_height_m_with_options((*C.SidereonMmapTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &cOptions, &output)
		})
	})
	runtime.KeepAlive(t)
	return float64(output.value_m), err
}

func (t *MmapTerrain) EllipsoidalHeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output C.SidereonEllipsoidalHeightM
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatusWithTerrainDiagnostics(func() uint32 {
			return C.sidereon_mmap_terrain_ellipsoidal_height_m((*C.SidereonMmapTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &output)
		}, true, false)
	})
	runtime.KeepAlive(t)
	return float64(output.value_m), err
}

func (t *MmapTerrain) EllipsoidalHeightMWithOptions(longitudeDeg, latitudeDeg float64, options DtedLookupOptions) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	cOptions := cDtedOptions(options)
	var output C.SidereonEllipsoidalHeightM
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatusWithTerrainDiagnostics(func() uint32 {
			return C.sidereon_mmap_terrain_ellipsoidal_height_m_with_options((*C.SidereonMmapTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &cOptions, &output)
		}, true, false)
	})
	runtime.KeepAlive(t)
	return float64(output.value_m), err
}

func (t *MmapTerrain) EllipsoidalHeightMWithModel(longitudeDeg, latitudeDeg float64, options DtedLookupOptions, model uint32, geoid *EGM96FifteenMinuteGeoid) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	cOptions := cDtedOptions(options)
	var output C.SidereonEllipsoidalHeightM
	call := func(terrainPointer unsafe.Pointer, geoidPointer unsafe.Pointer) error {
		return callStatusWithTerrainDiagnostics(func() uint32 {
			return C.sidereon_mmap_terrain_ellipsoidal_height_m_with_model(
				(*C.SidereonMmapTerrain)(terrainPointer), C.double(longitudeDeg), C.double(latitudeDeg), &cOptions,
				C.uint32_t(model), (*C.SidereonEgm96FifteenMinuteGeoid)(geoidPointer), &output,
			)
		}, true, false)
	}
	var err error
	if geoid == nil || geoid.handle == nil {
		err = t.handle.with(func(pointer unsafe.Pointer) error { return call(pointer, nil) })
	} else {
		err = t.handle.with(func(pointer unsafe.Pointer) error {
			return geoid.handle.with(func(geoidPointer unsafe.Pointer) error { return call(pointer, geoidPointer) })
		})
	}
	runtime.KeepAlive(t)
	runtime.KeepAlive(geoid)
	return float64(output.value_m), err
}

func (t *MmapTerrain) heightBatch(points []LonLatDeg, options DtedLookupOptions, typed bool) ([]MmapTerrainHeightResult, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	if _, err := checkedNativeAllocationSize(len(points), unsafe.Sizeof(C.SidereonLonLatDeg{})); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(points), unsafe.Sizeof(C.SidereonTerrainHeightResult{})); err != nil {
		return nil, err
	}
	cPoints := make([]C.SidereonLonLatDeg, len(points))
	cResults := make([]C.SidereonTerrainHeightResult, len(points))
	for i, point := range points {
		cPoints[i] = cLonLat(point)
	}
	cOptions := cDtedOptions(options)
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		var pointsPointer *C.SidereonLonLatDeg
		var resultPointer *C.SidereonTerrainHeightResult
		if len(cPoints) != 0 {
			pointsPointer = &cPoints[0]
			resultPointer = &cResults[0]
		}
		return callStatus(func() uint32 {
			if typed {
				return C.sidereon_mmap_terrain_orthometric_height_batch((*C.SidereonMmapTerrain)(pointer), pointsPointer, C.size_t(len(cPoints)), &cOptions, resultPointer)
			}
			return C.sidereon_mmap_terrain_height_batch((*C.SidereonMmapTerrain)(pointer), pointsPointer, C.size_t(len(cPoints)), &cOptions, resultPointer)
		})
	})
	runtime.KeepAlive(t)
	if err != nil {
		return nil, err
	}
	result := make([]MmapTerrainHeightResult, len(cResults))
	for i, value := range cResults {
		result[i] = MmapTerrainHeightResult{Status: uint32(value.status), HasOrthometricHeightM: bool(value.has_orthometric_height_m), OrthometricHeightM: float64(value.orthometric_height_m.value_m)}
	}
	return result, nil
}

func (t *MmapTerrain) HeightBatch(points []LonLatDeg, options DtedLookupOptions) ([]MmapTerrainHeightResult, error) {
	return t.heightBatch(points, options, false)
}

func (t *MmapTerrain) OrthometricHeightBatch(points []LonLatDeg, options DtedLookupOptions) ([]MmapTerrainHeightResult, error) {
	return t.heightBatch(points, options, true)
}

func mmapTileIndexFromC(value C.SidereonTerrainStoreTileIndex) TerrainStoreTileIndex {
	return TerrainStoreTileIndex{
		LatIndex: int32(value.lat_index), LonIndex: int32(value.lon_index),
		MinLongitudeDeg: float64(value.min_longitude_deg), MinLatitudeDeg: float64(value.min_latitude_deg),
		MaxLongitudeDeg: float64(value.max_longitude_deg), MaxLatitudeDeg: float64(value.max_latitude_deg),
		LonCount: uint32(value.lon_count), LatCount: uint32(value.lat_count),
		DataOffset: uint64(value.data_offset), DataLen: uint64(value.data_len),
		Checksum64: uint64(value.checksum64), VerticalDatum: uint32(value.vertical_datum),
	}
}

func copyMmapTileIndexLocked(call func(*C.SidereonTerrainStoreTileIndex, C.size_t, *C.size_t, *C.size_t) uint32) ([]TerrainStoreTileIndex, error) {
	var result []TerrainStoreTileIndex
	var operationErr error
	withCThread(func() {
		var written, required C.size_t
		if operationErr = statusErrorLocked(call(nil, 0, &written, &required)); operationErr != nil {
			return
		}
		queryWritten, err := checkedNativeCount(uint64(written))
		if err != nil {
			operationErr = err
			return
		}
		if queryWritten != 0 {
			operationErr = errors.New("sidereon: native terrain tile-index query wrote data")
			return
		}
		expected, err := checkedNativeCount(uint64(required))
		if err != nil {
			operationErr = err
			return
		}
		if _, err := checkedNativeAllocationSize(expected, unsafe.Sizeof(C.SidereonTerrainStoreTileIndex{})); err != nil {
			operationErr = err
			return
		}
		buffer := make([]C.SidereonTerrainStoreTileIndex, expected)
		var output *C.SidereonTerrainStoreTileIndex
		if len(buffer) != 0 {
			output = &buffer[0]
		}
		if operationErr = statusErrorLocked(call(output, C.size_t(len(buffer)), &written, &required)); operationErr != nil {
			return
		}
		count, err := validateTwoPassCounts("terrain tile-index output", len(buffer), expected, uint64(written), uint64(required))
		if err != nil {
			operationErr = err
			return
		}
		result = make([]TerrainStoreTileIndex, count)
		for i := range result {
			result[i] = mmapTileIndexFromC(buffer[i])
		}
	})
	return result, operationErr
}

func (t *MmapTerrain) TileIndex() ([]TerrainStoreTileIndex, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []TerrainStoreTileIndex
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		var copyErr error
		result, copyErr = copyMmapTileIndexLocked(func(out *C.SidereonTerrainStoreTileIndex, length C.size_t, written, required *C.size_t) uint32 {
			return C.sidereon_mmap_terrain_tile_index((*C.SidereonMmapTerrain)(pointer), out, length, written, required)
		})
		return copyErr
	})
	runtime.KeepAlive(t)
	return result, err
}

func (t *MmapTerrain) ToBytes() ([]byte, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		var copyErr error
		result, copyErr = copyNativeBytes("terrain store bytes", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_mmap_terrain_to_bytes((*C.SidereonMmapTerrain)(pointer), out, length, written, required)
		})
		return copyErr
	})
	runtime.KeepAlive(t)
	return result, err
}

func (t *MmapTerrain) Verify() error {
	if t == nil || t.handle == nil {
		return ErrClosed
	}
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatusWithTerrainDiagnostics(func() uint32 { return C.sidereon_mmap_terrain_verify((*C.SidereonMmapTerrain)(pointer)) }, false, true)
	})
	runtime.KeepAlive(t)
	return err
}

func (t *MmapTerrain) VerticalDatum() (uint32, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output uint32
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_mmap_terrain_vertical_datum((*C.SidereonMmapTerrain)(pointer), &output)
		})
	})
	runtime.KeepAlive(t)
	return uint32(output), err
}
