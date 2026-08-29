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

type DtedLookupOptions struct {
	Interpolation uint32
}

type LonLatDeg struct {
	LongitudeDeg float64
	LatitudeDeg  float64
}

type DtedHeightResult struct {
	Status     uint32
	HasHeightM bool
	HeightM    float64
}

type TerrainTileID struct {
	LatIndex int32
	LonIndex int32
}

type DtedTileListEntry struct {
	TileID TerrainTileID
	Path   string
}

type DtedTerrain struct {
	handle *positioningHandle
}

type DtedTile struct {
	handle *positioningHandle
}

func DefaultDTEDLookupOptions() (DtedLookupOptions, error) {
	var options C.SidereonDtedLookupOptions
	err := callStatus(func() uint32 { return C.sidereon_dted_lookup_options_init(&options) })
	return DtedLookupOptions{Interpolation: uint32(options.interpolation)}, err
}

func dtedTerrainFromPointer(pointer *C.SidereonDtedTerrain) (*DtedTerrain, error) {
	if pointer == nil {
		return nil, errors.New("sidereon: native DTED terrain constructor returned no handle")
	}
	return &DtedTerrain{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_dted_terrain_free((*C.SidereonDtedTerrain)(value))
	})}, nil
}

func DtedTerrainNew(root string) (*DtedTerrain, error) {
	var pointer *C.SidereonDtedTerrain
	err := withString(root, func(input *C.char) uint32 {
		return C.sidereon_dted_terrain_new(input, &pointer)
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_dted_terrain_free(pointer) })
		}
		return nil, err
	}
	return dtedTerrainFromPointer(pointer)
}

func (t *DtedTerrain) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return t.handle.close()
}

func cDtedOptions(options DtedLookupOptions) C.SidereonDtedLookupOptions {
	return C.SidereonDtedLookupOptions{interpolation: C.uint32_t(options.Interpolation)}
}

func cLonLat(point LonLatDeg) C.SidereonLonLatDeg {
	return C.SidereonLonLatDeg{lon_deg: C.double(point.LongitudeDeg), lat_deg: C.double(point.LatitudeDeg)}
}

func (t *DtedTerrain) HeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output C.double
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_dted_terrain_height_m((*C.SidereonDtedTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &output)
		})
	})
	runtime.KeepAlive(t)
	return float64(output), err
}

func (t *DtedTerrain) HeightMWithOptions(longitudeDeg, latitudeDeg float64, options DtedLookupOptions) (float64, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	cOptions := cDtedOptions(options)
	var output C.double
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_dted_terrain_height_m_with_options((*C.SidereonDtedTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &cOptions, &output)
		})
	})
	runtime.KeepAlive(t)
	return float64(output), err
}

func (t *DtedTerrain) HeightBatch(points []LonLatDeg, options DtedLookupOptions) ([]DtedHeightResult, error) {
	if t == nil || t.handle == nil {
		return nil, ErrClosed
	}
	if _, err := checkedNativeAllocationSize(len(points), unsafe.Sizeof(C.SidereonLonLatDeg{})); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(points), unsafe.Sizeof(C.SidereonDtedHeightResult{})); err != nil {
		return nil, err
	}
	cPoints := make([]C.SidereonLonLatDeg, len(points))
	for i, point := range points {
		cPoints[i] = cLonLat(point)
	}
	cResults := make([]C.SidereonDtedHeightResult, len(points))
	cOptions := cDtedOptions(options)
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		var pointsPointer *C.SidereonLonLatDeg
		var resultPointer *C.SidereonDtedHeightResult
		if len(cPoints) != 0 {
			pointsPointer = &cPoints[0]
			resultPointer = &cResults[0]
		}
		return callStatus(func() uint32 {
			return C.sidereon_dted_terrain_height_batch_m((*C.SidereonDtedTerrain)(pointer), pointsPointer, C.size_t(len(cPoints)), &cOptions, resultPointer)
		})
	})
	if err != nil {
		return nil, err
	}
	result := make([]DtedHeightResult, len(cResults))
	for i := range result {
		result[i] = DtedHeightResult{Status: uint32(cResults[i].status), HasHeightM: bool(cResults[i].has_height_m), HeightM: float64(cResults[i].height_m)}
	}
	runtime.KeepAlive(t)
	return result, nil
}

func dtedTileFromPointer(pointer *C.SidereonDtedTile) (*DtedTile, error) {
	if pointer == nil {
		return nil, errors.New("sidereon: native DTED tile constructor returned no handle")
	}
	return &DtedTile{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_dted_tile_free((*C.SidereonDtedTile)(value))
	})}, nil
}

func DtedTileLoad(path string) (*DtedTile, error) {
	var pointer *C.SidereonDtedTile
	err := withString(path, func(input *C.char) uint32 { return C.sidereon_dted_tile_load(input, &pointer) })
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_dted_tile_free(pointer) })
		}
		return nil, err
	}
	return dtedTileFromPointer(pointer)
}

func (t *DtedTile) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return t.handle.close()
}

func (t *DtedTile) Elevation(longitudeDeg, latitudeDeg float64) (int16, error) {
	if t == nil || t.handle == nil {
		return 0, ErrClosed
	}
	var output C.int16_t
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_dted_tile_get_elevation((*C.SidereonDtedTile)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &output)
		})
	})
	runtime.KeepAlive(t)
	return int16(output), err
}

func dtedEntries(entries []DtedTileListEntry) ([]C.SidereonDtedTileListEntry, []unsafe.Pointer, error) {
	entryBytes, err := checkedNativeAllocationSize(len(entries), unsafe.Sizeof(C.SidereonDtedTileListEntry{}))
	if err != nil {
		return nil, nil, err
	}
	if _, err := checkedNativeAllocationSize(len(entries), unsafe.Sizeof(unsafe.Pointer(nil))); err != nil {
		return nil, nil, err
	}
	var entryMemory unsafe.Pointer
	if entryBytes != 0 {
		entryMemory = C.malloc(C.size_t(entryBytes))
		if entryMemory == nil {
			return nil, nil, errors.New("sidereon: unable to allocate native DTED entries")
		}
	}
	cEntries := unsafe.Slice((*C.SidereonDtedTileListEntry)(entryMemory), len(entries))
	paths := make([]unsafe.Pointer, len(entries))
	for i, entry := range entries {
		path, err := cString(entry.Path)
		if err != nil {
			C.free(entryMemory)
			for _, value := range paths {
				if value != nil {
					C.free(value)
				}
			}
			return nil, nil, err
		}
		paths[i] = unsafe.Pointer(path)
		cEntries[i].tile_id = C.SidereonTerrainTileId{lat_index: C.int32_t(entry.TileID.LatIndex), lon_index: C.int32_t(entry.TileID.LonIndex)}
		cEntries[i].path = path
	}
	paths = append(paths, entryMemory)
	return cEntries, paths, nil
}

func freePointers(values []unsafe.Pointer) {
	for _, value := range values {
		if value != nil {
			C.free(value)
		}
	}
}

func DtedTileListToMmapStore(entries []DtedTileListEntry) ([]byte, error) {
	cEntries, paths, err := dtedEntries(entries)
	if err != nil {
		return nil, err
	}
	defer freePointers(paths)
	return copyNativeBytesWithTerrainDiagnostics("DTED tile-list store", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		var pointer *C.SidereonDtedTileListEntry
		if len(cEntries) != 0 {
			pointer = &cEntries[0]
		}
		return C.sidereon_dted_tile_list_to_mmap_store(pointer, C.size_t(len(cEntries)), out, length, written, required)
	}, false, true)
}

func DtedTreeToMmapStore(root string) ([]byte, error) {
	input, err := cString(root)
	if err != nil {
		return nil, err
	}
	defer C.free(unsafe.Pointer(input))
	return copyNativeBytesWithTerrainDiagnostics("DTED tree store", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_dted_tree_to_mmap_store(input, out, length, written, required)
	}, false, true)
}

func DtedInterpolationLabel(interpolation uint32) ([]byte, error) {
	return copyLabel(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_dted_interpolation_label(C.uint32_t(interpolation), out, length, written, required)
	})
}

type TerrainDatumError struct {
	Kind        uint32
	Path        string
	Message     string
	Remediation string
}

type TerrainStoreError struct {
	Kind             uint32
	Path             string
	Message          string
	Reason           string
	Version          uint16
	Tag              uint8
	LatIndex         int32
	LonIndex         int32
	ExpectedChecksum uint64
	FoundChecksum    uint64
}

func (e *TerrainDatumError) Error() string {
	if e == nil {
		return "sidereon: terrain datum error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Remediation != "" {
		return e.Remediation
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain datum error"
}

func (e *TerrainStoreError) Error() string {
	if e == nil {
		return "sidereon: terrain store error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return e.Reason
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain store error"
}

func cFixedString(value *C.char, length int) string {
	if value == nil {
		return ""
	}
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(value)), length)
	for i, item := range bytes {
		if item == 0 {
			return string(bytes[:i])
		}
	}
	return string(bytes)
}

func lastTerrainDatumErrorLocked() (TerrainDatumError, error) {
	var value C.SidereonTerrainDatumError
	err := statusErrorLocked(uint32(C.sidereon_last_terrain_datum_error(&value)))
	return TerrainDatumError{Kind: uint32(value.kind), Path: cFixedString(&value.path[0], C.SIDEREON_TERRAIN_ERROR_TEXT_C_BYTES), Message: cFixedString(&value.message[0], C.SIDEREON_TERRAIN_ERROR_TEXT_C_BYTES), Remediation: cFixedString(&value.remediation[0], C.SIDEREON_TERRAIN_ERROR_TEXT_C_BYTES)}, err
}

func LastTerrainStoreError() (TerrainStoreError, error) {
	var value TerrainStoreError
	var err error
	withCThread(func() {
		value, err = lastTerrainStoreErrorLocked()
	})
	return value, err
}

func LastTerrainDatumError() (TerrainDatumError, error) {
	var value TerrainDatumError
	var err error
	withCThread(func() {
		value, err = lastTerrainDatumErrorLocked()
	})
	return value, err
}

func lastTerrainStoreErrorLocked() (TerrainStoreError, error) {
	var value C.SidereonTerrainStoreError
	err := statusErrorLocked(uint32(C.sidereon_last_terrain_store_error(&value)))
	return TerrainStoreError{Kind: uint32(value.kind), Path: cFixedString(&value.path[0], C.SIDEREON_TERRAIN_ERROR_TEXT_C_BYTES), Message: cFixedString(&value.message[0], C.SIDEREON_TERRAIN_ERROR_TEXT_C_BYTES), Reason: cFixedString(&value.reason[0], C.SIDEREON_TERRAIN_ERROR_TEXT_C_BYTES), Version: uint16(value.version), Tag: uint8(value.tag), LatIndex: int32(value.lat_index), LonIndex: int32(value.lon_index), ExpectedChecksum: uint64(value.expected_checksum64), FoundChecksum: uint64(value.found_checksum64)}, err
}

func TerrainStoreChecksum64(data []byte) (uint64, error) {
	var checksum C.uint64_t
	operationErr := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_terrain_store_checksum64(input, length, &checksum)
	})
	return uint64(checksum), operationErr
}

func VerticalDatumLabel(datum uint32) ([]byte, error) {
	return copyLabel(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_vertical_datum_label(C.uint32_t(datum), out, length, written, required)
	})
}

func TerrainGeoidModelLabel(model uint32) ([]byte, error) {
	return copyLabel(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_terrain_geoid_model_label(C.uint32_t(model), out, length, written, required)
	})
}
