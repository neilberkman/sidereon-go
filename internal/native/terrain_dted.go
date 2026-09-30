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

func cFixedString(value *C.char, capacity C.size_t) string {
	if value == nil || capacity == 0 {
		return ""
	}
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(value)), int(capacity))
	for index, item := range bytes {
		if item == 0 {
			return string(bytes[:index])
		}
	}
	return string(bytes)
}

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
	Error      TerrainLookupError
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
		return callStatusWithTerrainLookupDiagnostics(func() uint32 {
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
		call := func() uint32 {
			return C.sidereon_dted_terrain_height_m_with_options((*C.SidereonDtedTerrain)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &cOptions, &output)
		}
		if !validDtedLookupOptions(options) {
			return callStatus(call)
		}
		return callStatusWithTerrainLookupDiagnostics(call)
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
	typedErrors := make([]TerrainLookupError, len(points))
	cOptions := cDtedOptions(options)
	err := t.handle.withExclusive(func(pointer unsafe.Pointer) error {
		var pointsPointer *C.SidereonLonLatDeg
		var resultPointer *C.SidereonDtedHeightResult
		if len(cPoints) != 0 {
			pointsPointer = &cPoints[0]
			resultPointer = &cResults[0]
		}
		var operationErr error
		withCThread(func() {
			status := uint32(C.sidereon_dted_terrain_height_batch_m((*C.SidereonDtedTerrain)(pointer), pointsPointer, C.size_t(len(cPoints)), &cOptions, resultPointer))
			operationErr = statusErrorLocked(status)
			if operationErr != nil {
				return
			}
			for i := range cResults {
				if uint32(cResults[i].status) == uint32(C.SIDEREON_STATUS_OK) {
					continue
				}
				typedErrors[i], operationErr = terrainLookupBatchErrorLocked(i, cResults[i].error)
				if operationErr != nil {
					return
				}
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	result := make([]DtedHeightResult, len(cResults))
	for i := range result {
		result[i] = DtedHeightResult{Status: uint32(cResults[i].status), HasHeightM: bool(cResults[i].has_height_m), HeightM: float64(cResults[i].height_m), Error: typedErrors[i]}
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
	err := withStringStatus(path, func(input *C.char) uint32 { return C.sidereon_dted_tile_load(input, &pointer) }, statusDtedTileErrorLocked)
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
		return withCThreadError(func() error {
			return statusDtedTileErrorLocked(uint32(C.sidereon_dted_tile_get_elevation((*C.SidereonDtedTile)(pointer), C.double(longitudeDeg), C.double(latitudeDeg), &output)))
		})
	})
	runtime.KeepAlive(t)
	return int16(output), err
}

func (t *DtedTile) HorizontalDatum() (HorizontalDatum, error) {
	if t == nil || t.handle == nil {
		return HorizontalDatum{}, ErrClosed
	}
	var output C.SidereonDtedHorizontalDatumValue
	err := t.handle.with(func(pointer unsafe.Pointer) error {
		return callStatusWithTerrainDiagnostics(func() uint32 {
			return uint32(C.sidereon_dted_tile_horizontal_datum((*C.SidereonDtedTile)(pointer), &output))
		}, false, false)
	})
	runtime.KeepAlive(t)
	return terrainHorizontalDatum(output), err
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

func terrainHorizontalDatum(value C.SidereonDtedHorizontalDatumValue) HorizontalDatum {
	return HorizontalDatum{Kind: uint32(value.kind), Text: cFixedString(&value.text[0], C.SIDEREON_DTED_DATUM_TEXT_C_BYTES), WGS84Compatible: bool(value.wgs84_compatible)}
}

func terrainTileError(value C.SidereonDtedTileError) DtedTileError {
	return DtedTileError{
		Kind: uint32(value.kind), Field: cFixedString(&value.field[0], C.SIDEREON_TERRAIN_ERROR_FIELD_C_BYTES), Text: cFixedString(&value.text[0], C.SIDEREON_TERRAIN_ERROR_FIELD_C_BYTES),
		HasCounts: bool(value.has_counts), LongitudeCount: uint64(value.lon_count), LatitudeCount: uint64(value.lat_count),
		HasLengths: bool(value.has_lengths), ActualBytes: uint64(value.actual_bytes), ExpectedBytes: uint64(value.expected_bytes),
		HasQuery: bool(value.has_query), LongitudeDeg: float64(value.longitude_deg), LatitudeDeg: float64(value.latitude_deg), OriginLongitudeDeg: float64(value.origin_longitude_deg), OriginLatitudeDeg: float64(value.origin_latitude_deg),
		HasLongitudeIndex: bool(value.has_longitude_index), HasLatitudeIndex: bool(value.has_latitude_index), LongitudeIndex: uint64(value.longitude_index), LatitudeIndex: uint64(value.latitude_index),
		HasChecksum: bool(value.has_checksum), Checksum: int32(value.checksum), Sum: int32(value.sum), HasHemisphere: bool(value.has_hemisphere), Hemisphere: uint32(value.hemisphere), ExpectedHemispheres: cFixedString(&value.expected_hemispheres[0], C.SIDEREON_TERRAIN_ERROR_FIELD_C_BYTES),
		HasNegativeIndex: bool(value.has_negative_index), NegativeIndex: int64(value.negative_index), HasInterval: bool(value.has_interval), IntervalTenthsArcsec: uint32(value.interval_tenths_arcsec), Count: uint64(value.count), HasDeclared: bool(value.has_declared), Declared: int32(value.declared),
	}
}

func terrainLookupError(value C.SidereonTerrainLookupError) TerrainLookupError {
	return TerrainLookupError{
		Kind: uint32(value.kind), HasTile: bool(value.has_tile), LatIndex: int32(value.lat_index), LonIndex: int32(value.lon_index), HasPosting: bool(value.has_posting), LatitudePosting: uint64(value.latitude_posting), LongitudePosting: uint64(value.longitude_posting),
		HasHorizontalDatum: bool(value.has_horizontal_datum), HorizontalDatum: terrainHorizontalDatum(value.horizontal_datum), HasOrigin: bool(value.has_origin), OriginLatitudeDeg: int32(value.origin_latitude_deg), OriginLongitudeDeg: int32(value.origin_longitude_deg), HasTileError: bool(value.has_tile_error), TileError: terrainTileError(value.tile_error),
	}
}

func terrainErrorTextLocked(family, part uint32) (string, error) {
	var written, required C.size_t
	status := uint32(C.sidereon_last_terrain_error_text(C.uint32_t(family), C.uint32_t(part), nil, 0, &written, &required))
	if err := statusErrorLocked(status); err != nil {
		return "", err
	}
	length, err := sizeTToInt(required, "terrain error text")
	if err != nil {
		return "", err
	}
	buffer := make([]byte, length)
	if length != 0 {
		status = uint32(C.sidereon_last_terrain_error_text(C.uint32_t(family), C.uint32_t(part), (*C.uint8_t)(unsafe.Pointer(&buffer[0])), required, &written, &required))
		if err := statusErrorLocked(status); err != nil {
			return "", err
		}
	}
	count, err := sizeTToInt(written, "terrain error text written length")
	if err != nil {
		return "", err
	}
	return string(buffer[:count]), nil
}

func terrainTileErrorTextsLocked(value *DtedTileError) error {
	const family = uint32(C.SIDEREON_TERRAIN_ERROR_FAMILY_DTED_TILE)
	for _, item := range []struct {
		part   uint32
		target *string
	}{
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_PATH), &value.Path},
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_MESSAGE), &value.Message},
	} {
		text, err := terrainErrorTextLocked(family, item.part)
		if err != nil {
			return err
		}
		*item.target = text
	}
	return nil
}

func terrainBatchErrorTextLocked(row int, part uint32) (string, error) {
	var written, required C.size_t
	status := uint32(C.sidereon_last_terrain_batch_error_text(C.size_t(row), C.uint32_t(part), nil, 0, &written, &required))
	if err := statusErrorLocked(status); err != nil {
		return "", err
	}
	length, err := sizeTToInt(required, "terrain batch error text")
	if err != nil {
		return "", err
	}
	buffer := make([]byte, length)
	if length != 0 {
		status = uint32(C.sidereon_last_terrain_batch_error_text(C.size_t(row), C.uint32_t(part), (*C.uint8_t)(unsafe.Pointer(&buffer[0])), required, &written, &required))
		if err := statusErrorLocked(status); err != nil {
			return "", err
		}
	}
	count, err := sizeTToInt(written, "terrain batch error text written length")
	if err != nil {
		return "", err
	}
	return string(buffer[:count]), nil
}

func terrainLookupBatchErrorLocked(row int, value C.SidereonTerrainLookupError) (TerrainLookupError, error) {
	detail := terrainLookupError(value)
	for _, item := range []struct {
		part   uint32
		target *string
	}{
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_PATH), &detail.Path},
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_MESSAGE), &detail.Message},
	} {
		text, err := terrainBatchErrorTextLocked(row, item.part)
		if err != nil {
			return detail, err
		}
		*item.target = text
	}
	if detail.HasTileError {
		detail.TileError.Path = detail.Path
		detail.TileError.Message = detail.Message
	}
	return detail, nil
}

func lastTerrainDatumErrorLocked() (TerrainDatumError, error) {
	var value C.SidereonTerrainDatumError
	err := statusErrorLocked(uint32(C.sidereon_last_terrain_datum_error(&value)))
	detail := TerrainDatumError{Kind: uint32(value.kind), Terrain: terrainLookupError(value.terrain), Geoid: GeoidError{Kind: uint32(value.geoid.kind), Expected: uint64(value.geoid.expected), Found: uint64(value.geoid.found), Index: uint64(value.geoid.index)}}
	if err != nil {
		return detail, err
	}
	const family = uint32(C.SIDEREON_TERRAIN_ERROR_FAMILY_TERRAIN_DATUM)
	for _, item := range []struct {
		part   uint32
		target *string
	}{
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_PATH), &detail.Path},
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_MESSAGE), &detail.Message},
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_REMEDIATION), &detail.Remediation},
	} {
		text, textErr := terrainErrorTextLocked(family, item.part)
		if textErr != nil {
			return detail, textErr
		}
		*item.target = text
	}
	if detail.Terrain.HasTileError {
		if textErr := terrainTileErrorTextsLocked(&detail.Terrain.TileError); textErr != nil {
			return detail, textErr
		}
	}
	if detail.Kind == TerrainDatumErrorTerrainValue {
		detail.Terrain.Path = detail.Path
		detail.Terrain.Message = detail.Message
	}
	if detail.Kind == TerrainDatumErrorGeoidValue {
		const geoidFamily = uint32(C.SIDEREON_TERRAIN_ERROR_FAMILY_GEOID)
		for _, item := range []struct {
			part   uint32
			target *string
		}{
			{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_FIELD), &detail.Geoid.Field},
			{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_REASON), &detail.Geoid.Reason},
		} {
			text, textErr := terrainErrorTextLocked(geoidFamily, item.part)
			if textErr != nil {
				return detail, textErr
			}
			*item.target = text
		}
	}
	return detail, nil
}

func lastTerrainLookupErrorLocked() (TerrainLookupError, error) {
	var value C.SidereonTerrainLookupError
	err := statusErrorLocked(uint32(C.sidereon_last_terrain_lookup_error(&value)))
	detail := terrainLookupError(value)
	if err != nil {
		return detail, err
	}
	const family = uint32(C.SIDEREON_TERRAIN_ERROR_FAMILY_TERRAIN_LOOKUP)
	for _, item := range []struct {
		part   uint32
		target *string
	}{
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_PATH), &detail.Path},
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_MESSAGE), &detail.Message},
	} {
		text, textErr := terrainErrorTextLocked(family, item.part)
		if textErr != nil {
			return detail, textErr
		}
		*item.target = text
	}
	if detail.HasTileError {
		if textErr := terrainTileErrorTextsLocked(&detail.TileError); textErr != nil {
			return detail, textErr
		}
	}
	return detail, nil
}

func LastTerrainStoreError() (TerrainStoreError, error) {
	var value TerrainStoreError
	var err error
	withCThread(func() { value, err = lastTerrainStoreErrorLocked() })
	return value, err
}

func LastTerrainDatumError() (TerrainDatumError, error) {
	var value TerrainDatumError
	var err error
	withCThread(func() { value, err = lastTerrainDatumErrorLocked() })
	return value, err
}

func lastTerrainStoreErrorLocked() (TerrainStoreError, error) {
	var value C.SidereonTerrainStoreError
	err := statusErrorLocked(uint32(C.sidereon_last_terrain_store_error(&value)))
	detail := TerrainStoreError{Kind: uint32(value.kind), Version: uint16(value.version), Tag: uint8(value.tag), LatIndex: int32(value.lat_index), LonIndex: int32(value.lon_index), ExpectedTileID: TerrainTileID{LatIndex: int32(value.expected_tile_id.lat_index), LonIndex: int32(value.expected_tile_id.lon_index)}, FoundTileID: TerrainTileID{LatIndex: int32(value.found_tile_id.lat_index), LonIndex: int32(value.found_tile_id.lon_index)}, ExpectedChecksum: uint64(value.expected_checksum64), FoundChecksum: uint64(value.found_checksum64), Field: cFixedString(&value.field[0], C.SIDEREON_TERRAIN_ERROR_FIELD_C_BYTES), HasHorizontalDatum: bool(value.has_horizontal_datum), HorizontalDatum: terrainHorizontalDatum(value.horizontal_datum), HasTileError: bool(value.has_tile_error), TileError: terrainTileError(value.tile_error)}
	if err != nil {
		return detail, err
	}
	const family = uint32(C.SIDEREON_TERRAIN_ERROR_FAMILY_TERRAIN_STORE)
	for _, item := range []struct {
		part   uint32
		target *string
	}{
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_PATH), &detail.Path},
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_MESSAGE), &detail.Message},
		{uint32(C.SIDEREON_TERRAIN_ERROR_TEXT_REASON), &detail.Reason},
	} {
		text, textErr := terrainErrorTextLocked(family, item.part)
		if textErr != nil {
			return detail, textErr
		}
		*item.target = text
	}
	if detail.HasTileError {
		if textErr := terrainTileErrorTextsLocked(&detail.TileError); textErr != nil {
			return detail, textErr
		}
	}
	return detail, nil
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
