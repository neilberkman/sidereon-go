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
	"fmt"
	"runtime"
	"unsafe"
)

const (
	DtedInterpolationNearestPostingValue = uint32(C.SIDEREON_DTED_INTERPOLATION_NEAREST_POSTING)
	DtedInterpolationBilinearValue       = uint32(C.SIDEREON_DTED_INTERPOLATION_BILINEAR)

	DigestProvenanceVerifiedValue = uint32(C.SIDEREON_DIGEST_PROVENANCE_VERIFIED)
	DigestProvenanceAttestedValue = uint32(C.SIDEREON_DIGEST_PROVENANCE_ATTESTED)

	VerticalDatumEGM96MSLOrthometricValue = uint32(C.SIDEREON_VERTICAL_DATUM_EGM96_MSL_ORTHOMETRIC)

	TerrainGeoidModelEGM96OneDegreeValue     = uint32(C.SIDEREON_TERRAIN_GEOID_MODEL_EGM96_ONE_DEGREE)
	TerrainGeoidModelEGM96FifteenMinuteValue = uint32(C.SIDEREON_TERRAIN_GEOID_MODEL_EGM96_FIFTEEN_MINUTE)

	GeofenceErrorNoneValue           = uint32(C.SIDEREON_GEOFENCE_ERROR_KIND_NONE)
	GeofenceErrorTooFewVerticesValue = uint32(C.SIDEREON_GEOFENCE_ERROR_KIND_TOO_FEW_VERTICES)
	GeofenceErrorInvalidInputValue   = uint32(C.SIDEREON_GEOFENCE_ERROR_KIND_INVALID_INPUT)
	GeofenceErrorGeodesicValue       = uint32(C.SIDEREON_GEOFENCE_ERROR_KIND_GEODESIC)
	GeofenceErrorDOPValue            = uint32(C.SIDEREON_GEOFENCE_ERROR_KIND_DOP)
	GeofenceErrorMetricsValue        = uint32(C.SIDEREON_GEOFENCE_ERROR_KIND_ERROR_METRICS)

	GeofenceENUCovarianceM2Value  = uint32(C.SIDEREON_GEOFENCE_UNCERTAINTY_KIND_ENU_COVARIANCE_M2)
	GeofenceECEFCovarianceM2Value = uint32(C.SIDEREON_GEOFENCE_UNCERTAINTY_KIND_ECEF_COVARIANCE_M2)
	GeofenceCEPRadiusMValue       = uint32(C.SIDEREON_GEOFENCE_UNCERTAINTY_KIND_CEP_RADIUS_M)

	GeofenceBoundaryNormalValue   = uint32(C.SIDEREON_GEOFENCE_PROBABILITY_METHOD_BOUNDARY_NORMAL)
	GeofencePlanarQuadratureValue = uint32(C.SIDEREON_GEOFENCE_PROBABILITY_METHOD_PLANAR_QUADRATURE)

	GeofenceEnteredValue = uint32(C.SIDEREON_GEOFENCE_CROSSING_KIND_ENTERED)
	GeofenceLeftValue    = uint32(C.SIDEREON_GEOFENCE_CROSSING_KIND_LEFT)

	ObservabilityRankDeficientValue  = uint32(C.SIDEREON_OBSERVABILITY_TIER_RANK_DEFICIENT)
	ObservabilityZeroRedundancyValue = uint32(C.SIDEREON_OBSERVABILITY_TIER_ZERO_REDUNDANCY)
	ObservabilityWeakValue           = uint32(C.SIDEREON_OBSERVABILITY_TIER_WEAK)
	ObservabilityNominalValue        = uint32(C.SIDEREON_OBSERVABILITY_TIER_NOMINAL)

	ProjVgridshiftArithmeticSeparateMultiplyAddValue = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_ARITHMETIC_SEPARATE_MULTIPLY_ADD)
	ProjVgridshiftArithmeticFusedMultiplyAddValue    = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_ARITHMETIC_FUSED_MULTIPLY_ADD)

	ProjVgridshiftErrorNoneValue                  = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_ERROR_KIND_NONE)
	ProjVgridshiftErrorNonFiniteCoordinateValue   = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_ERROR_KIND_NON_FINITE_COORDINATE)
	ProjVgridshiftErrorCoordinateOutsideGridValue = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_ERROR_KIND_COORDINATE_OUTSIDE_GRID)

	ProjVgridshiftCoordinateNoneValue      = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_COORDINATE_NONE)
	ProjVgridshiftCoordinateLatitudeValue  = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_COORDINATE_LATITUDE)
	ProjVgridshiftCoordinateLongitudeValue = uint32(C.SIDEREON_PROJ_VGRIDSHIFT_COORDINATE_LONGITUDE)

	TerrainDatumErrorNoneValue            = uint32(C.SIDEREON_TERRAIN_DATUM_ERROR_KIND_NONE)
	TerrainDatumErrorTerrainValue         = uint32(C.SIDEREON_TERRAIN_DATUM_ERROR_KIND_TERRAIN)
	TerrainDatumErrorGeoidValue           = uint32(C.SIDEREON_TERRAIN_DATUM_ERROR_KIND_GEOID)
	TerrainDatumErrorIOValue              = uint32(C.SIDEREON_TERRAIN_DATUM_ERROR_KIND_IO)
	TerrainDatumErrorMissingEGM96DACValue = uint32(C.SIDEREON_TERRAIN_DATUM_ERROR_KIND_MISSING_EGM96_DAC)

	TerrainStoreErrorNoneValue                     = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_NONE)
	TerrainStoreErrorIOValue                       = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_IO)
	TerrainStoreErrorParseValue                    = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_PARSE)
	TerrainStoreErrorUnsupportedVersionValue       = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_UNSUPPORTED_VERSION)
	TerrainStoreErrorUnsupportedDatumValue         = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_UNSUPPORTED_DATUM)
	TerrainStoreErrorDuplicateTileValue            = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_DUPLICATE_TILE)
	TerrainStoreErrorChecksumValue                 = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_CHECKSUM)
	TerrainStoreErrorTileIDMismatchValue           = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_TILE_ID_MISMATCH)
	TerrainStoreErrorAttestedChecksumMismatchValue = uint32(C.SIDEREON_TERRAIN_STORE_ERROR_KIND_ATTESTED_CHECKSUM_MISMATCH)

	TropoMappingErrorNoneValue         = uint32(C.SIDEREON_TROPO_MAPPING_ERROR_KIND_NONE)
	TropoMappingErrorLowElevationValue = uint32(C.SIDEREON_TROPO_MAPPING_ERROR_KIND_LOW_ELEVATION)
	TropoMappingErrorInvalidInputValue = uint32(C.SIDEREON_TROPO_MAPPING_ERROR_KIND_INVALID_INPUT)

	Egm2008GridSpacingOneMinuteValue          = uint32(C.SIDEREON_EGM2008_GRID_SPACING_ONE_MINUTE)
	Egm2008GridSpacingTwoPointFiveMinuteValue = uint32(C.SIDEREON_EGM2008_GRID_SPACING_TWO_POINT_FIVE_MINUTE)
)

type GeoidPoint struct {
	Latitude  float64
	Longitude float64
}

type ProjVGridshiftError struct {
	Kind       uint32
	Coordinate uint32
}

type Egm2008RasterWindow struct {
	Spacing   uint32
	LatMinDeg float64
	LonMinDeg float64
	NLat      int
	NLon      int
}

type GeoidGrid struct {
	handle *positioningHandle
}

type EGM96FifteenMinuteGeoid struct {
	handle *positioningHandle
}

func copyDoublesLocked(call func(*C.double, C.size_t, *C.size_t, *C.size_t) uint32) ([]float64, error) {
	var result []float64
	var operationErr error
	withCThread(func() {
		var written, required C.size_t
		status := call(nil, 0, &written, &required)
		if operationErr = statusErrorLocked(status); operationErr != nil {
			return
		}
		queryWritten, err := checkedNativeCount(uint64(written))
		if err != nil {
			operationErr = fmt.Errorf("sidereon: native double output query written count: %w", err)
			return
		}
		if queryWritten != 0 {
			operationErr = errors.New("sidereon: native double output query wrote data")
			return
		}
		requiredCount, err := checkedNativeCount(uint64(required))
		if err != nil {
			operationErr = fmt.Errorf("sidereon: native double output required count: %w", err)
			return
		}
		if _, err := checkedNativeAllocationSize(requiredCount, unsafe.Sizeof(C.double(0))); err != nil {
			operationErr = err
			return
		}
		buffer := make([]C.double, requiredCount)
		var output *C.double
		if len(buffer) != 0 {
			output = &buffer[0]
		}
		status = call(output, C.size_t(len(buffer)), &written, &required)
		if operationErr = statusErrorLocked(status); operationErr != nil {
			return
		}
		count, err := validateTwoPassCounts("double output", len(buffer), len(buffer), uint64(written), uint64(required))
		if err != nil {
			operationErr = err
			return
		}
		result = make([]float64, count)
		for i := range result {
			result[i] = float64(buffer[i])
		}
	})
	return result, operationErr
}

func cSizeFromInt(value int, name string) (C.size_t, error) {
	if value < 0 {
		return 0, errors.New("sidereon: " + name + " must not be negative")
	}
	if uint64(value) > uint64(^C.size_t(0)) {
		return 0, errors.New("sidereon: " + name + " exceeds native size_t")
	}
	return C.size_t(value), nil
}

func checkedProduct(left, right int, name string) (int, error) {
	if left < 0 || right < 0 {
		return 0, errors.New("sidereon: " + name + " must not be negative")
	}
	maxInt := int(^uint(0) >> 1)
	if right != 0 && left > maxInt/right {
		return 0, errors.New("sidereon: " + name + " overflows")
	}
	return left * right, nil
}

func geoidPointSlice(points []GeoidPoint) ([]C.SidereonGeoidPoint, error) {
	if _, err := checkedNativeAllocationSize(len(points), unsafe.Sizeof(C.SidereonGeoidPoint{})); err != nil {
		return nil, err
	}
	result := make([]C.SidereonGeoidPoint, len(points))
	for i, point := range points {
		result[i] = C.SidereonGeoidPoint{latitude: C.double(point.Latitude), longitude: C.double(point.Longitude)}
	}
	return result, nil
}

func geoidGridFromBytes(data []byte, fn func(*C.uint8_t, C.size_t, **C.SidereonGeoidGrid) uint32) (*GeoidGrid, error) {
	var pointer *C.SidereonGeoidGrid
	operationErr := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return fn(input, length, &pointer)
	})
	if operationErr != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_geoid_grid_free(pointer) })
		}
		return nil, operationErr
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_geoid_grid_free((*C.SidereonGeoidGrid)(value))
	})}, nil
}

func GeoidGridFromText(data []byte) (*GeoidGrid, error) {
	return geoidGridFromBytes(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonGeoidGrid) uint32 {
		return C.sidereon_geoid_grid_from_text(input, length, out)
	})
}

func GeoidGridFromPROJEGM96GTX(data []byte) (*GeoidGrid, error) {
	return geoidGridFromBytes(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonGeoidGrid) uint32 {
		return C.sidereon_geoid_grid_from_proj_egm96_gtx(input, length, out)
	})
}

func GeoidGridFromEGM96DAC(data []byte) (*GeoidGrid, error) {
	return geoidGridFromBytes(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonGeoidGrid) uint32 {
		return C.sidereon_geoid_grid_from_egm96_dac(input, length, out)
	})
}

func GeoidGridFromEGM2008Raster(data []byte, spacing uint32) (*GeoidGrid, error) {
	return geoidGridFromBytes(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonGeoidGrid) uint32 {
		return C.sidereon_geoid_grid_from_egm2008_raster(input, length, C.uint32_t(spacing), out)
	})
}

func GeoidGridFromEGM2008RasterWindow(data []byte, window Egm2008RasterWindow) (*GeoidGrid, error) {
	if _, err := checkedProduct(window.NLat, window.NLon, "raster window dimensions"); err != nil {
		return nil, err
	}
	nLat, err := cSizeFromInt(window.NLat, "raster window latitude count")
	if err != nil {
		return nil, err
	}
	nLon, err := cSizeFromInt(window.NLon, "raster window longitude count")
	if err != nil {
		return nil, err
	}
	cWindow := C.SidereonEgm2008RasterWindow{
		spacing:     C.uint32_t(window.Spacing),
		lat_min_deg: C.double(window.LatMinDeg), lon_min_deg: C.double(window.LonMinDeg),
		n_lat: nLat, n_lon: nLon,
	}
	return geoidGridFromBytes(data, func(input *C.uint8_t, length C.size_t, out **C.SidereonGeoidGrid) uint32 {
		return C.sidereon_geoid_grid_from_egm2008_raster_window(input, length, &cWindow, out)
	})
}

func NewGeoidGrid(latMinDeg, lonMinDeg, dLatDeg, dLonDeg float64, nLat, nLon int, values []float64) (*GeoidGrid, error) {
	latCount, err := cSizeFromInt(nLat, "geoid latitude count")
	if err != nil {
		return nil, err
	}
	lonCount, err := cSizeFromInt(nLon, "geoid longitude count")
	if err != nil {
		return nil, err
	}
	valueCount, err := checkedProduct(nLat, nLon, "geoid grid dimensions")
	if err != nil {
		return nil, err
	}
	if valueCount != len(values) {
		return nil, errors.New("sidereon: geoid grid value count does not match dimensions")
	}
	if _, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	cValues := make([]C.double, len(values))
	for i, value := range values {
		cValues[i] = C.double(value)
	}
	var pointer *C.SidereonGeoidGrid
	var operationErr error
	withCThread(func() {
		var input *C.double
		if len(cValues) != 0 {
			input = &cValues[0]
		}
		operationErr = statusErrorLocked(C.sidereon_geoid_grid_new(
			C.double(latMinDeg), C.double(lonMinDeg), C.double(dLatDeg), C.double(dLonDeg),
			latCount, lonCount, input, C.size_t(len(cValues)), &pointer,
		))
	})
	if operationErr != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_geoid_grid_free(pointer) })
		}
		return nil, operationErr
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_geoid_grid_free((*C.SidereonGeoidGrid)(value))
	})}, nil
}

func (g *GeoidGrid) Close() error {
	if g == nil || g.handle == nil {
		return nil
	}
	return g.handle.close()
}

func (g *GeoidGrid) UndulationDeg(latDeg, lonDeg float64) (float64, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	var output C.double
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_geoid_grid_undulation_deg((*C.SidereonGeoidGrid)(pointer), C.double(latDeg), C.double(lonDeg), &output)
		})
	})
	runtime.KeepAlive(g)
	return float64(output), err
}

func (g *GeoidGrid) UndulationRad(latRad, lonRad float64) (float64, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	var output C.double
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_geoid_grid_undulation_rad((*C.SidereonGeoidGrid)(pointer), C.double(latRad), C.double(lonRad), &output)
		})
	})
	runtime.KeepAlive(g)
	return float64(output), err
}

func (g *GeoidGrid) UndulationPROJRad(latRad, lonRad float64, arithmetic uint32) (float64, ProjVGridshiftError, error) {
	if g == nil || g.handle == nil {
		return 0, ProjVGridshiftError{}, ErrClosed
	}
	var output C.double
	var detail C.SidereonProjVgridshiftError
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_geoid_grid_undulation_proj_rad(
				(*C.SidereonGeoidGrid)(pointer), C.double(latRad), C.double(lonRad), C.uint32_t(arithmetic), &detail, &output,
			)
		})
	})
	runtime.KeepAlive(g)
	return float64(output), ProjVGridshiftError{Kind: uint32(detail.kind), Coordinate: uint32(detail.coordinate)}, err
}

func (g *GeoidGrid) UndulationsDeg(points []GeoidPoint) ([]float64, error) {
	return g.undulations(points, false)
}

func (g *GeoidGrid) UndulationsRad(points []GeoidPoint) ([]float64, error) {
	return g.undulations(points, true)
}

func (g *GeoidGrid) undulations(points []GeoidPoint, radians bool) ([]float64, error) {
	if g == nil || g.handle == nil {
		return nil, ErrClosed
	}
	cPoints, err := geoidPointSlice(points)
	if err != nil {
		return nil, err
	}
	var output []float64
	err = g.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		output, operationErr = copyDoublesLocked(func(out *C.double, length C.size_t, written, required *C.size_t) uint32 {
			var input *C.SidereonGeoidPoint
			if len(cPoints) != 0 {
				input = &cPoints[0]
			}
			if radians {
				return C.sidereon_geoid_grid_undulations_rad((*C.SidereonGeoidGrid)(pointer), input, C.size_t(len(cPoints)), out, length, written, required)
			}
			return C.sidereon_geoid_grid_undulations_deg((*C.SidereonGeoidGrid)(pointer), input, C.size_t(len(cPoints)), out, length, written, required)
		})
		return operationErr
	})
	runtime.KeepAlive(g)
	return output, err
}

func (g *GeoidGrid) OrthometricHeightRad(ellipsoidalHeightM, latRad, lonRad float64) (float64, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	var output C.double
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_geoid_grid_orthometric_height_rad((*C.SidereonGeoidGrid)(pointer), C.double(ellipsoidalHeightM), C.double(latRad), C.double(lonRad), &output)
		})
	})
	runtime.KeepAlive(g)
	return float64(output), err
}

func (g *GeoidGrid) EllipsoidalHeightRad(orthometricHeightM, latRad, lonRad float64) (float64, error) {
	if g == nil || g.handle == nil {
		return 0, ErrClosed
	}
	var output C.double
	err := g.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_geoid_grid_ellipsoidal_height_rad((*C.SidereonGeoidGrid)(pointer), C.double(orthometricHeightM), C.double(latRad), C.double(lonRad), &output)
		})
	})
	runtime.KeepAlive(g)
	return float64(output), err
}

func EGM96FifteenMinuteGeoidFromBytes(data []byte) (*EGM96FifteenMinuteGeoid, error) {
	var pointer *C.SidereonEgm96FifteenMinuteGeoid
	operationErr := withInputTerrainDiagnostics(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_egm96_15m_geoid_from_ww15mgh_dac_bytes(input, length, &pointer)
	}, true, false)
	if operationErr != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_egm96_15m_geoid_free(pointer) })
		}
		return nil, operationErr
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native EGM96 geoid constructor returned no handle")
	}
	return &EGM96FifteenMinuteGeoid{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_egm96_15m_geoid_free((*C.SidereonEgm96FifteenMinuteGeoid)(value))
	})}, nil
}

func EGM96FifteenMinuteGeoidFromPath(path string) (*EGM96FifteenMinuteGeoid, error) {
	var pointer *C.SidereonEgm96FifteenMinuteGeoid
	err := withStringTerrainDiagnostics(path, func(input *C.char) uint32 {
		return C.sidereon_egm96_15m_geoid_from_ww15mgh_dac_path(input, &pointer)
	}, true, false)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_egm96_15m_geoid_free(pointer) })
		}
		return nil, err
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native EGM96 geoid constructor returned no handle")
	}
	return &EGM96FifteenMinuteGeoid{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_egm96_15m_geoid_free((*C.SidereonEgm96FifteenMinuteGeoid)(value))
	})}, nil
}

func (g *EGM96FifteenMinuteGeoid) Close() error {
	if g == nil || g.handle == nil {
		return nil
	}
	return g.handle.close()
}

func scalarGeoidCall(fn func(*C.double) uint32) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 { return fn(&output) })
	return float64(output), err
}

func GeoidUndulation(latRad, lonRad float64) (float64, error) {
	return scalarGeoidCall(func(out *C.double) uint32 {
		return C.sidereon_geoid_undulation(C.double(latRad), C.double(lonRad), out)
	})
}

func OrthometricHeight(ellipsoidalHeightM, latRad, lonRad float64) (float64, error) {
	return scalarGeoidCall(func(out *C.double) uint32 {
		return C.sidereon_orthometric_height_m(C.double(ellipsoidalHeightM), C.double(latRad), C.double(lonRad), out)
	})
}

func EllipsoidalHeight(orthometricHeightM, latRad, lonRad float64) (float64, error) {
	return scalarGeoidCall(func(out *C.double) uint32 {
		return C.sidereon_ellipsoidal_height_m(C.double(orthometricHeightM), C.double(latRad), C.double(lonRad), out)
	})
}

func EGM96Undulation(latRad, lonRad float64) (float64, error) {
	return scalarGeoidCall(func(out *C.double) uint32 {
		return C.sidereon_egm96_undulation(C.double(latRad), C.double(lonRad), out)
	})
}

func EGM96OrthometricHeight(ellipsoidalHeightM, latRad, lonRad float64) (float64, error) {
	return scalarGeoidCall(func(out *C.double) uint32 {
		return C.sidereon_egm96_orthometric_height_m(C.double(ellipsoidalHeightM), C.double(latRad), C.double(lonRad), out)
	})
}

func EGM96EllipsoidalHeight(orthometricHeightM, latRad, lonRad float64) (float64, error) {
	return scalarGeoidCall(func(out *C.double) uint32 {
		return C.sidereon_egm96_ellipsoidal_height_m(C.double(orthometricHeightM), C.double(latRad), C.double(lonRad), out)
	})
}

func geoidUndulations(points []GeoidPoint, kind uint32) ([]float64, error) {
	cPoints, err := geoidPointSlice(points)
	if err != nil {
		return nil, err
	}
	return copyDoublesLocked(func(out *C.double, length C.size_t, written, required *C.size_t) uint32 {
		var input *C.SidereonGeoidPoint
		if len(cPoints) != 0 {
			input = &cPoints[0]
		}
		switch kind {
		case 0:
			return C.sidereon_geoid_undulations_rad(input, C.size_t(len(cPoints)), out, length, written, required)
		case 1:
			return C.sidereon_geoid_undulations_deg(input, C.size_t(len(cPoints)), out, length, written, required)
		case 2:
			return C.sidereon_egm96_undulations_rad(input, C.size_t(len(cPoints)), out, length, written, required)
		default:
			return C.sidereon_egm96_undulations_deg(input, C.size_t(len(cPoints)), out, length, written, required)
		}
	})
}

func GeoidUndulationsRad(points []GeoidPoint) ([]float64, error) { return geoidUndulations(points, 0) }
func GeoidUndulationsDeg(points []GeoidPoint) ([]float64, error) { return geoidUndulations(points, 1) }
func EGM96UndulationsRad(points []GeoidPoint) ([]float64, error) { return geoidUndulations(points, 2) }
func EGM96UndulationsDeg(points []GeoidPoint) ([]float64, error) { return geoidUndulations(points, 3) }
