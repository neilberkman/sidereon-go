package sidereon

import (
	"errors"
	"os"
	"unsafe"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// GeoidPointDeg is a geoid query point expressed in degrees.
type GeoidPointDeg struct {
	LatitudeDeg  float64
	LongitudeDeg float64
}

// GeoidPointRad is a geoid query point expressed in radians.
type GeoidPointRad struct {
	LatitudeRad  float64
	LongitudeRad float64
}

// PROJVGridshiftArithmetic selects the floating-point evaluation
// recipe used by PROJ-compatible vertical-grid interpolation.
type PROJVGridshiftArithmetic uint32

const (
	PROJVGridshiftArithmeticSeparateMultiplyAdd PROJVGridshiftArithmetic = PROJVGridshiftArithmetic(native.ProjVgridshiftArithmeticSeparateMultiplyAddValue)
	PROJVGridshiftArithmeticFusedMultiplyAdd    PROJVGridshiftArithmetic = PROJVGridshiftArithmetic(native.ProjVgridshiftArithmeticFusedMultiplyAddValue)
)

// PROJVGridshiftErrorKind identifies a PROJ vertical-grid error.
type PROJVGridshiftErrorKind uint32

const (
	PROJVGridshiftErrorNone                  PROJVGridshiftErrorKind = PROJVGridshiftErrorKind(native.ProjVgridshiftErrorNoneValue)
	PROJVGridshiftErrorNonFiniteCoordinate   PROJVGridshiftErrorKind = PROJVGridshiftErrorKind(native.ProjVgridshiftErrorNonFiniteCoordinateValue)
	PROJVGridshiftErrorCoordinateOutsideGrid PROJVGridshiftErrorKind = PROJVGridshiftErrorKind(native.ProjVgridshiftErrorCoordinateOutsideGridValue)
)

// PROJVGridshiftCoordinate identifies the coordinate in a PROJ error.
type PROJVGridshiftCoordinate uint32

const (
	PROJVGridshiftCoordinateNone      PROJVGridshiftCoordinate = PROJVGridshiftCoordinate(native.ProjVgridshiftCoordinateNoneValue)
	PROJVGridshiftCoordinateLatitude  PROJVGridshiftCoordinate = PROJVGridshiftCoordinate(native.ProjVgridshiftCoordinateLatitudeValue)
	PROJVGridshiftCoordinateLongitude PROJVGridshiftCoordinate = PROJVGridshiftCoordinate(native.ProjVgridshiftCoordinateLongitudeValue)
)

// PROJVGridshiftError contains the C engine's typed interpolation diagnostic.
type PROJVGridshiftError struct {
	Kind       PROJVGridshiftErrorKind
	Coordinate PROJVGridshiftCoordinate
}

// EGM2008GridSpacing selects the global EGM2008 raster spacing.
type EGM2008GridSpacing uint32

const (
	EGM2008OneMinute          EGM2008GridSpacing = EGM2008GridSpacing(native.Egm2008GridSpacingOneMinuteValue)
	EGM2008TwoPointFiveMinute EGM2008GridSpacing = EGM2008GridSpacing(native.Egm2008GridSpacingTwoPointFiveMinuteValue)
)

// EGM2008RasterWindow describes a degree-coordinate crop in an EGM2008
// raster. Counts are the number of latitude and longitude postings.
type EGM2008RasterWindow struct {
	Spacing   EGM2008GridSpacing
	LatMinDeg float64
	LonMinDeg float64
	NLat      int
	NLon      int
}

func checkedEnvironmentAllocation(count int, elementSize uintptr) error {
	if count < 0 || elementSize == 0 {
		return errors.New("sidereon: invalid environment allocation size")
	}
	maxInt := uint64(^uint(0) >> 1)
	if uint64(elementSize) > maxInt || uint64(count) > maxInt/uint64(elementSize) {
		return errors.New("sidereon: environment allocation size overflows")
	}
	return nil
}

func validDTEDInterpolation(value DTEDInterpolation) bool {
	return value == DTEDNearestPosting || value == DTEDBilinear
}

func validVerticalDatum(value VerticalDatum) bool {
	return value == EGM96MSLOrthometric
}

func validTerrainGeoidModel(value TerrainGeoidModel) bool {
	return value == EGM96OneDegree || value == EGM96FifteenMinute
}

func validObservabilityTier(value ObservabilityTier) bool {
	return value == ObservabilityRankDeficient || value == ObservabilityZeroRedundancy || value == ObservabilityWeak || value == ObservabilityNominal
}

func validTropoTimeScale(value TimeScale) bool {
	return value >= UTC && value <= TCB
}

func validGeofenceUncertaintyKind(value GeofenceUncertaintyKind) bool {
	return value == GeofenceENUCovarianceM2 || value == GeofenceECEFCovarianceM2 || value == GeofenceCEPRadiusM
}

func validGeofenceProbabilityMethod(value GeofenceProbabilityMethod) bool {
	return value == GeofenceBoundaryNormal || value == GeofencePlanarQuadrature
}

func validProjArithmetic(value PROJVGridshiftArithmetic) bool {
	return value == PROJVGridshiftArithmeticSeparateMultiplyAdd || value == PROJVGridshiftArithmeticFusedMultiplyAdd
}

func validEGM2008Spacing(value EGM2008GridSpacing) bool {
	return value == EGM2008OneMinute || value == EGM2008TwoPointFiveMinute
}

// GeoidGrid owns a native C geoid-grid handle. Read methods may run
// concurrently. Close waits for active C calls, clears the resource, and is
// idempotent. The handle must not be copied after first use.
type GeoidGrid struct {
	_      noCopy
	native *native.GeoidGrid
}

// GeoidGridFromText parses a geoid grid from bytes using the C parser.
func GeoidGridFromText(data []byte) (*GeoidGrid, error) {
	value, err := native.GeoidGridFromText(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{native: value}, nil
}

// GeoidGridFromPROJEGM96GTX parses a PROJ EGM96 GTX grid from bytes.
func GeoidGridFromPROJEGM96GTX(data []byte) (*GeoidGrid, error) {
	value, err := native.GeoidGridFromPROJEGM96GTX(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{native: value}, nil
}

// GeoidGridFromEGM96DAC parses EGM96 DAC bytes.
func GeoidGridFromEGM96DAC(data []byte) (*GeoidGrid, error) {
	value, err := native.GeoidGridFromEGM96DAC(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{native: value}, nil
}

// GeoidGridFromEGM2008Raster parses a global EGM2008 raster at the supplied
// spacing.
func GeoidGridFromEGM2008Raster(data []byte, spacing EGM2008GridSpacing) (*GeoidGrid, error) {
	if !validEGM2008Spacing(spacing) {
		return nil, errors.New("sidereon: invalid EGM2008 grid spacing")
	}
	value, err := native.GeoidGridFromEGM2008Raster(data, uint32(spacing))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{native: value}, nil
}

// GeoidGridFromEGM2008RasterWindow parses an EGM2008 raster crop.
func GeoidGridFromEGM2008RasterWindow(data []byte, window EGM2008RasterWindow) (*GeoidGrid, error) {
	if !validEGM2008Spacing(window.Spacing) {
		return nil, errors.New("sidereon: invalid EGM2008 grid spacing")
	}
	value, err := native.GeoidGridFromEGM2008RasterWindow(data, native.Egm2008RasterWindow{
		Spacing: uint32(window.Spacing), LatMinDeg: window.LatMinDeg, LonMinDeg: window.LonMinDeg,
		NLat: window.NLat, NLon: window.NLon,
	})
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{native: value}, nil
}

// NewGeoidGrid constructs a grid from copied values. values are ordered by
// the native C row-major grid contract and are not retained by Go.
func NewGeoidGrid(latMinDeg, lonMinDeg, dLatDeg, dLonDeg float64, nLat, nLon int, values []float64) (*GeoidGrid, error) {
	value, err := native.NewGeoidGrid(latMinDeg, lonMinDeg, dLatDeg, dLonDeg, nLat, nLon, values)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native geoid grid constructor returned no handle")
	}
	return &GeoidGrid{native: value}, nil
}

// Close waits for active C calls, clears the native grid, and is idempotent.
func (g *GeoidGrid) Close() error {
	if g == nil || g.native == nil {
		return nil
	}
	return publicError(g.native.Close())
}

// UndulationDeg returns the geoid undulation at a degree-coordinate point.
func (g *GeoidGrid) UndulationDeg(latitudeDeg, longitudeDeg float64) (float64, error) {
	if g == nil || g.native == nil {
		return 0, ErrClosed
	}
	value, err := g.native.UndulationDeg(latitudeDeg, longitudeDeg)
	return value, publicError(err)
}

// UndulationRad returns the geoid undulation at a radian-coordinate point.
func (g *GeoidGrid) UndulationRad(latitudeRad, longitudeRad float64) (float64, error) {
	if g == nil || g.native == nil {
		return 0, ErrClosed
	}
	value, err := g.native.UndulationRad(latitudeRad, longitudeRad)
	return value, publicError(err)
}

// UndulationPROJRad returns the PROJ-compatible interpolation result and
// typed arithmetic/coordinate diagnostic.
func (g *GeoidGrid) UndulationPROJRad(latitudeRad, longitudeRad float64, arithmetic PROJVGridshiftArithmetic) (float64, PROJVGridshiftError, error) {
	if g == nil || g.native == nil {
		return 0, PROJVGridshiftError{}, ErrClosed
	}
	if !validProjArithmetic(arithmetic) {
		return 0, PROJVGridshiftError{}, errors.New("sidereon: invalid PROJ vertical-grid arithmetic")
	}
	value, detail, err := g.native.UndulationPROJRad(latitudeRad, longitudeRad, uint32(arithmetic))
	return value, PROJVGridshiftError{Kind: PROJVGridshiftErrorKind(detail.Kind), Coordinate: PROJVGridshiftCoordinate(detail.Coordinate)}, publicError(err)
}

// UndulationsDeg returns copied degree-coordinate undulations in input order.
func (g *GeoidGrid) UndulationsDeg(points []GeoidPointDeg) ([]float64, error) {
	return g.undulations(points, false)
}

// UndulationsRad returns copied radian-coordinate undulations in input order.
func (g *GeoidGrid) UndulationsRad(points []GeoidPointRad) ([]float64, error) {
	return g.undulations(points, true)
}

func (g *GeoidGrid) undulations(points any, radians bool) ([]float64, error) {
	if g == nil || g.native == nil {
		return nil, ErrClosed
	}
	var nativePoints []native.GeoidPoint
	if radians {
		values := points.([]GeoidPointRad)
		if err := checkedEnvironmentAllocation(len(values), unsafe.Sizeof(native.GeoidPoint{})); err != nil {
			return nil, err
		}
		nativePoints = make([]native.GeoidPoint, len(values))
		for i, point := range values {
			nativePoints[i] = native.GeoidPoint{Latitude: point.LatitudeRad, Longitude: point.LongitudeRad}
		}
	} else {
		values := points.([]GeoidPointDeg)
		if err := checkedEnvironmentAllocation(len(values), unsafe.Sizeof(native.GeoidPoint{})); err != nil {
			return nil, err
		}
		nativePoints = make([]native.GeoidPoint, len(values))
		for i, point := range values {
			nativePoints[i] = native.GeoidPoint{Latitude: point.LatitudeDeg, Longitude: point.LongitudeDeg}
		}
	}
	var value []float64
	var err error
	if radians {
		value, err = g.native.UndulationsRad(nativePoints)
	} else {
		value, err = g.native.UndulationsDeg(nativePoints)
	}
	return value, publicError(err)
}

// OrthometricHeightRad converts an ellipsoidal height to orthometric height.
func (g *GeoidGrid) OrthometricHeightRad(ellipsoidalHeightM, latitudeRad, longitudeRad float64) (float64, error) {
	if g == nil || g.native == nil {
		return 0, ErrClosed
	}
	value, err := g.native.OrthometricHeightRad(ellipsoidalHeightM, latitudeRad, longitudeRad)
	return value, publicError(err)
}

// EllipsoidalHeightRad converts an orthometric height to ellipsoidal height.
func (g *GeoidGrid) EllipsoidalHeightRad(orthometricHeightM, latitudeRad, longitudeRad float64) (float64, error) {
	if g == nil || g.native == nil {
		return 0, ErrClosed
	}
	value, err := g.native.EllipsoidalHeightRad(orthometricHeightM, latitudeRad, longitudeRad)
	return value, publicError(err)
}

// EGM96FifteenMinuteGeoid owns a loaded WW15MGH.DAC C handle. Read access may
// run concurrently. Close waits for active C calls, clears the resource, and
// is idempotent. The handle must not be copied after first use.
type EGM96FifteenMinuteGeoid struct {
	_      noCopy
	native *native.EGM96FifteenMinuteGeoid
}

// EGM96FifteenMinuteGeoidFromBytes loads WW15MGH.DAC bytes.
func EGM96FifteenMinuteGeoidFromBytes(data []byte) (*EGM96FifteenMinuteGeoid, error) {
	value, err := native.EGM96FifteenMinuteGeoidFromBytes(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native EGM96 geoid constructor returned no handle")
	}
	return &EGM96FifteenMinuteGeoid{native: value}, nil
}

// EGM96FifteenMinuteGeoidFromPath loads WW15MGH.DAC from path using Go file
// I/O and the canonical byte constructor.
func EGM96FifteenMinuteGeoidFromPath(path string) (*EGM96FifteenMinuteGeoid, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	value, err := EGM96FifteenMinuteGeoidFromBytes(data)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Close releases the native fifteen-minute geoid and is idempotent.
func (g *EGM96FifteenMinuteGeoid) Close() error {
	if g == nil || g.native == nil {
		return nil
	}
	return publicError(g.native.Close())
}

// GeoidUndulation returns the embedded geoid undulation in metres.
func GeoidUndulation(latitudeRad, longitudeRad float64) (float64, error) {
	value, err := native.GeoidUndulation(latitudeRad, longitudeRad)
	return value, publicError(err)
}

// OrthometricHeight converts ellipsoidal height using the embedded geoid.
func OrthometricHeight(ellipsoidalHeightM, latitudeRad, longitudeRad float64) (float64, error) {
	value, err := native.OrthometricHeight(ellipsoidalHeightM, latitudeRad, longitudeRad)
	return value, publicError(err)
}

// EllipsoidalHeight converts orthometric height using the embedded geoid.
func EllipsoidalHeight(orthometricHeightM, latitudeRad, longitudeRad float64) (float64, error) {
	value, err := native.EllipsoidalHeight(orthometricHeightM, latitudeRad, longitudeRad)
	return value, publicError(err)
}

// EGM96Undulation returns the embedded EGM96 undulation in metres.
func EGM96Undulation(latitudeRad, longitudeRad float64) (float64, error) {
	value, err := native.EGM96Undulation(latitudeRad, longitudeRad)
	return value, publicError(err)
}

// EGM96OrthometricHeight converts ellipsoidal height using EGM96.
func EGM96OrthometricHeight(ellipsoidalHeightM, latitudeRad, longitudeRad float64) (float64, error) {
	value, err := native.EGM96OrthometricHeight(ellipsoidalHeightM, latitudeRad, longitudeRad)
	return value, publicError(err)
}

// EGM96EllipsoidalHeight converts orthometric height using EGM96.
func EGM96EllipsoidalHeight(orthometricHeightM, latitudeRad, longitudeRad float64) (float64, error) {
	value, err := native.EGM96EllipsoidalHeight(orthometricHeightM, latitudeRad, longitudeRad)
	return value, publicError(err)
}

func geoidUndulationsRad(points []GeoidPointRad, egm96 bool) ([]float64, error) {
	if err := checkedEnvironmentAllocation(len(points), unsafe.Sizeof(native.GeoidPoint{})); err != nil {
		return nil, err
	}
	nativePoints := make([]native.GeoidPoint, len(points))
	for i, point := range points {
		nativePoints[i] = native.GeoidPoint{Latitude: point.LatitudeRad, Longitude: point.LongitudeRad}
	}
	var value []float64
	var err error
	if egm96 {
		value, err = native.EGM96UndulationsRad(nativePoints)
	} else {
		value, err = native.GeoidUndulationsRad(nativePoints)
	}
	return value, publicError(err)
}

func geoidUndulationsDeg(points []GeoidPointDeg, egm96 bool) ([]float64, error) {
	if err := checkedEnvironmentAllocation(len(points), unsafe.Sizeof(native.GeoidPoint{})); err != nil {
		return nil, err
	}
	nativePoints := make([]native.GeoidPoint, len(points))
	for i, point := range points {
		nativePoints[i] = native.GeoidPoint{Latitude: point.LatitudeDeg, Longitude: point.LongitudeDeg}
	}
	var value []float64
	var err error
	if egm96 {
		value, err = native.EGM96UndulationsDeg(nativePoints)
	} else {
		value, err = native.GeoidUndulationsDeg(nativePoints)
	}
	return value, publicError(err)
}

// GeoidUndulationsRad returns embedded geoid undulations for copied input
// points in input order.
func GeoidUndulationsRad(points []GeoidPointRad) ([]float64, error) {
	return geoidUndulationsRad(points, false)
}

// GeoidUndulationsDeg returns embedded geoid undulations for degree points.
func GeoidUndulationsDeg(points []GeoidPointDeg) ([]float64, error) {
	return geoidUndulationsDeg(points, false)
}

// EGM96UndulationsRad returns embedded EGM96 undulations for radian points.
func EGM96UndulationsRad(points []GeoidPointRad) ([]float64, error) {
	return geoidUndulationsRad(points, true)
}

// EGM96UndulationsDeg returns embedded EGM96 undulations for degree points.
func EGM96UndulationsDeg(points []GeoidPointDeg) ([]float64, error) {
	return geoidUndulationsDeg(points, true)
}

// DTEDInterpolation selects the terrain interpolation mode.
type DTEDInterpolation uint32

const (
	// DTEDNearestPosting selects the nearest stored posting.
	DTEDNearestPosting DTEDInterpolation = DTEDInterpolation(native.DtedInterpolationNearestPostingValue)
	// DTEDBilinear selects bilinear interpolation.
	DTEDBilinear DTEDInterpolation = DTEDInterpolation(native.DtedInterpolationBilinearValue)
)

// DTEDLookupOptions controls DTED height interpolation.
type DTEDLookupOptions struct {
	Interpolation DTEDInterpolation
}

// DTEDHeightResult contains one batch result. Status preserves the native
// per-point status even when another point in the batch succeeds.
type DTEDHeightResult struct {
	Status     StatusCode
	HasHeightM bool
	HeightM    float64
}

// LonLatDeg is a longitude-first DTED query point in degrees.
type LonLatDeg struct {
	LongitudeDeg float64
	LatitudeDeg  float64
}

// DTEDTerrain owns a mutable native C tile cache. Read calls may run
// concurrently; internally mutating cache calls and Close are serialized.
// Close clears the resource and is idempotent. The handle must not be copied
// after first use.
type DTEDTerrain struct {
	_      noCopy
	native *native.DtedTerrain
}

// DTEDInterpolationLabel returns the native interpolation label.
func DTEDInterpolationLabel(interpolation DTEDInterpolation) (string, error) {
	if !validDTEDInterpolation(interpolation) {
		return "", errors.New("sidereon: invalid DTED interpolation")
	}
	value, err := native.DtedInterpolationLabel(uint32(interpolation))
	return string(value), publicError(err)
}

// DefaultDTEDLookupOptions returns the native default DTED options.
func DefaultDTEDLookupOptions() (DTEDLookupOptions, error) {
	value, err := native.DefaultDTEDLookupOptions()
	return DTEDLookupOptions{Interpolation: DTEDInterpolation(value.Interpolation)}, publicError(err)
}

// NewDTEDTerrain opens a DTED cache rooted at root.
func NewDTEDTerrain(root string) (*DTEDTerrain, error) {
	value, err := native.DtedTerrainNew(root)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native DTED terrain constructor returned no handle")
	}
	return &DTEDTerrain{native: value}, nil
}

// Close releases the DTED cache and is idempotent.
func (t *DTEDTerrain) Close() error {
	if t == nil || t.native == nil {
		return nil
	}
	return publicError(t.native.Close())
}

// HeightM returns an orthometric DTED height in metres.
func (t *DTEDTerrain) HeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.HeightM(longitudeDeg, latitudeDeg)
	return value, publicError(err)
}

// HeightMWithOptions returns an orthometric DTED height using options.
func (t *DTEDTerrain) HeightMWithOptions(longitudeDeg, latitudeDeg float64, options DTEDLookupOptions) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	if !validDTEDInterpolation(options.Interpolation) {
		return 0, errors.New("sidereon: invalid DTED interpolation")
	}
	value, err := t.native.HeightMWithOptions(longitudeDeg, latitudeDeg, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)})
	return value, publicError(err)
}

// HeightBatch returns one copied result for every input point.
func (t *DTEDTerrain) HeightBatch(points []LonLatDeg, options DTEDLookupOptions) ([]DTEDHeightResult, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	if !validDTEDInterpolation(options.Interpolation) {
		return nil, errors.New("sidereon: invalid DTED interpolation")
	}
	if err := checkedEnvironmentAllocation(len(points), unsafe.Sizeof(native.LonLatDeg{})); err != nil {
		return nil, err
	}
	nativePoints := make([]native.LonLatDeg, len(points))
	for i, point := range points {
		nativePoints[i] = native.LonLatDeg{LongitudeDeg: point.LongitudeDeg, LatitudeDeg: point.LatitudeDeg}
	}
	value, err := t.native.HeightBatch(nativePoints, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)})
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]DTEDHeightResult, len(value))
	for i, item := range value {
		result[i] = DTEDHeightResult{Status: StatusCode(item.Status), HasHeightM: item.HasHeightM, HeightM: item.HeightM}
	}
	return result, nil
}

// DTEDTile owns one loaded DTED C tile. Read calls may run concurrently. Close
// waits for active C calls, clears the resource, and is idempotent. The handle
// must not be copied after first use.
type DTEDTile struct {
	_      noCopy
	native *native.DtedTile
}

// LoadDTEDTile loads a DTED tile from path.
func LoadDTEDTile(path string) (*DTEDTile, error) {
	value, err := native.DtedTileLoad(path)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native DTED tile constructor returned no handle")
	}
	return &DTEDTile{native: value}, nil
}

// Close releases the tile and is idempotent.
func (t *DTEDTile) Close() error {
	if t == nil || t.native == nil {
		return nil
	}
	return publicError(t.native.Close())
}

// Elevation returns the nearest stored orthometric posting in metres.
func (t *DTEDTile) Elevation(longitudeDeg, latitudeDeg float64) (int16, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.Elevation(longitudeDeg, latitudeDeg)
	return value, publicError(err)
}

// TerrainTileID identifies an integer-degree DTED tile.
type TerrainTileID struct {
	LatIndex int32
	LonIndex int32
}

// DTEDTileListEntry maps a tile identifier to a source path.
type DTEDTileListEntry struct {
	TileID TerrainTileID
	Path   string
}

// DTEDTileListToMMapStore converts DTED sources to copied memory-mappable
// terrain-store bytes.
func DTEDTileListToMMapStore(entries []DTEDTileListEntry) ([]byte, error) {
	if err := checkedEnvironmentAllocation(len(entries), unsafe.Sizeof(native.DtedTileListEntry{})); err != nil {
		return nil, err
	}
	nativeEntries := make([]native.DtedTileListEntry, len(entries))
	for i, entry := range entries {
		nativeEntries[i] = native.DtedTileListEntry{TileID: native.TerrainTileID{LatIndex: entry.TileID.LatIndex, LonIndex: entry.TileID.LonIndex}, Path: entry.Path}
	}
	value, err := native.DtedTileListToMmapStore(nativeEntries)
	return value, publicError(err)
}

// WriteDTEDTileListToMMapStore writes a converted DTED tile list to path.
func WriteDTEDTileListToMMapStore(entries []DTEDTileListEntry, path string) error {
	if err := checkedEnvironmentAllocation(len(entries), unsafe.Sizeof(native.DtedTileListEntry{})); err != nil {
		return err
	}
	nativeEntries := make([]native.DtedTileListEntry, len(entries))
	for i, entry := range entries {
		nativeEntries[i] = native.DtedTileListEntry{TileID: native.TerrainTileID{LatIndex: entry.TileID.LatIndex, LonIndex: entry.TileID.LonIndex}, Path: entry.Path}
	}
	return publicError(native.WriteDtedTileListToMmapStore(nativeEntries, path))
}

// DTEDTreeToMMapStore converts a DTED directory tree to copied store bytes.
func DTEDTreeToMMapStore(root string) ([]byte, error) {
	value, err := native.DtedTreeToMmapStore(root)
	return value, publicError(err)
}

// WriteDTEDTreeToMMapStore writes a converted DTED tree to path.
func WriteDTEDTreeToMMapStore(root, path string) error {
	return publicError(native.WriteDtedTreeToMmapStore(root, path))
}

// MMapTerrainHeightResult contains one memory-mappable terrain batch result.
// Status preserves the per-point native status; OrthometricHeightM is valid
// only when HasOrthometricHeightM is true.
type MMapTerrainHeightResult struct {
	Status                StatusCode
	HasOrthometricHeightM bool
	OrthometricHeightM    float64
}

// TerrainStoreTileIndex describes one copied memory-mappable terrain tile.
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
	VerticalDatum   VerticalDatum
}

// DigestProvenance identifies whether a terrain-store checksum was verified or
// supplied by the caller as an attestation.
type DigestProvenance uint32

const (
	// DigestVerified means the native reader verified the store payload.
	DigestVerified DigestProvenance = DigestProvenance(native.DigestProvenanceVerifiedValue)
	// DigestAttested means the checksum was supplied by the caller and has not
	// yet been verified.
	DigestAttested DigestProvenance = DigestProvenance(native.DigestProvenanceAttestedValue)
)

// MMapTerrain owns a memory-mappable terrain-store C handle. Read calls may
// run concurrently; internally mutating terrain-cache calls and Close are
// serialized. Close clears the resource and is idempotent. The value must not
// be copied after first use.
type MMapTerrain struct {
	_      noCopy
	native *native.MmapTerrain
}

func mmapTerrain(value *native.MmapTerrain, err error) (*MMapTerrain, error) {
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native memory-mappable terrain constructor returned no handle")
	}
	return &MMapTerrain{native: value}, nil
}

// MMapTerrainFromBytes parses and copies a terrain-store byte span.
func MMapTerrainFromBytes(data []byte) (*MMapTerrain, error) {
	return mmapTerrain(native.MmapTerrainFromBytes(data))
}

// MMapTerrainFromPath opens a terrain-store file.
func MMapTerrainFromPath(path string) (*MMapTerrain, error) {
	return mmapTerrain(native.MmapTerrainFromPath(path))
}

// MMapTerrainFromPathAttested opens a terrain-store file with a caller-supplied
// full-store checksum claim.
func MMapTerrainFromPathAttested(path string, checksum64 uint64) (*MMapTerrain, error) {
	return mmapTerrain(native.MmapTerrainFromPathAttested(path, checksum64))
}

// Close releases the memory-mappable terrain reader and is idempotent.
func (t *MMapTerrain) Close() error {
	if t == nil || t.native == nil {
		return nil
	}
	return publicError(t.native.Close())
}

// Checksum64 returns the full terrain-store checksum.
func (t *MMapTerrain) Checksum64() (uint64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.Checksum64()
	return value, publicError(err)
}

// DigestProvenance returns the checksum provenance carried by the reader.
func (t *MMapTerrain) DigestProvenance() (DigestProvenance, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.DigestProvenance()
	return DigestProvenance(value), publicError(err)
}

// HeightM returns bilinear orthometric height in metres.
func (t *MMapTerrain) HeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.HeightM(longitudeDeg, latitudeDeg)
	return value, publicError(err)
}

// HeightMWithOptions returns orthometric height using interpolation options.
func (t *MMapTerrain) HeightMWithOptions(longitudeDeg, latitudeDeg float64, options DTEDLookupOptions) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	if !validDTEDInterpolation(options.Interpolation) {
		return 0, errors.New("sidereon: invalid DTED interpolation")
	}
	value, err := t.native.HeightMWithOptions(longitudeDeg, latitudeDeg, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)})
	return value, publicError(err)
}

// OrthometricHeightM returns a typed orthometric terrain height in metres.
func (t *MMapTerrain) OrthometricHeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.OrthometricHeightM(longitudeDeg, latitudeDeg)
	return value, publicError(err)
}

// OrthometricHeightMWithOptions returns typed orthometric height using options.
func (t *MMapTerrain) OrthometricHeightMWithOptions(longitudeDeg, latitudeDeg float64, options DTEDLookupOptions) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	if !validDTEDInterpolation(options.Interpolation) {
		return 0, errors.New("sidereon: invalid DTED interpolation")
	}
	value, err := t.native.OrthometricHeightMWithOptions(longitudeDeg, latitudeDeg, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)})
	return value, publicError(err)
}

// EllipsoidalHeightM returns EGM96 one-degree ellipsoidal height in metres.
func (t *MMapTerrain) EllipsoidalHeightM(longitudeDeg, latitudeDeg float64) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.EllipsoidalHeightM(longitudeDeg, latitudeDeg)
	return value, publicError(err)
}

// EllipsoidalHeightMWithOptions returns one-degree ellipsoidal height using
// interpolation options.
func (t *MMapTerrain) EllipsoidalHeightMWithOptions(longitudeDeg, latitudeDeg float64, options DTEDLookupOptions) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	if !validDTEDInterpolation(options.Interpolation) {
		return 0, errors.New("sidereon: invalid DTED interpolation")
	}
	value, err := t.native.EllipsoidalHeightMWithOptions(longitudeDeg, latitudeDeg, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)})
	return value, publicError(err)
}

// EllipsoidalHeightMWithModel returns ellipsoidal height using the selected
// terrain geoid model. The fifteen-minute model requires geoid to be non-nil.
func (t *MMapTerrain) EllipsoidalHeightMWithModel(longitudeDeg, latitudeDeg float64, options DTEDLookupOptions, model TerrainGeoidModel, geoid *EGM96FifteenMinuteGeoid) (float64, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	if !validDTEDInterpolation(options.Interpolation) {
		return 0, errors.New("sidereon: invalid DTED interpolation")
	}
	if !validTerrainGeoidModel(model) {
		return 0, errors.New("sidereon: invalid terrain geoid model")
	}
	if model == EGM96FifteenMinute && (geoid == nil || geoid.native == nil) {
		return 0, errors.New("sidereon: fifteen-minute terrain model requires a geoid")
	}
	if model == EGM96OneDegree && geoid != nil {
		return 0, errors.New("sidereon: one-degree terrain model does not accept a fifteen-minute geoid")
	}
	var nativeGeoid *native.EGM96FifteenMinuteGeoid
	if geoid != nil {
		nativeGeoid = geoid.native
	}
	value, err := t.native.EllipsoidalHeightMWithModel(longitudeDeg, latitudeDeg, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)}, uint32(model), nativeGeoid)
	return value, publicError(err)
}

// HeightBatch returns copied typed orthometric results in input order.
func (t *MMapTerrain) HeightBatch(points []LonLatDeg, options DTEDLookupOptions) ([]MMapTerrainHeightResult, error) {
	return t.heightBatch(points, options, false)
}

// OrthometricHeightBatch returns copied typed orthometric results in input
// order through the typed C route.
func (t *MMapTerrain) OrthometricHeightBatch(points []LonLatDeg, options DTEDLookupOptions) ([]MMapTerrainHeightResult, error) {
	return t.heightBatch(points, options, true)
}

func (t *MMapTerrain) heightBatch(points []LonLatDeg, options DTEDLookupOptions, typed bool) ([]MMapTerrainHeightResult, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	if !validDTEDInterpolation(options.Interpolation) {
		return nil, errors.New("sidereon: invalid DTED interpolation")
	}
	if err := checkedEnvironmentAllocation(len(points), unsafe.Sizeof(native.LonLatDeg{})); err != nil {
		return nil, err
	}
	nativePoints := make([]native.LonLatDeg, len(points))
	for i, point := range points {
		nativePoints[i] = native.LonLatDeg{LongitudeDeg: point.LongitudeDeg, LatitudeDeg: point.LatitudeDeg}
	}
	var value []native.MmapTerrainHeightResult
	var err error
	if typed {
		value, err = t.native.OrthometricHeightBatch(nativePoints, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)})
	} else {
		value, err = t.native.HeightBatch(nativePoints, native.DtedLookupOptions{Interpolation: uint32(options.Interpolation)})
	}
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]MMapTerrainHeightResult, len(value))
	for i, item := range value {
		result[i] = MMapTerrainHeightResult{Status: StatusCode(item.Status), HasOrthometricHeightM: item.HasOrthometricHeightM, OrthometricHeightM: item.OrthometricHeightM}
	}
	return result, nil
}

// TileIndex returns copied tile-index rows in native store order.
func (t *MMapTerrain) TileIndex() ([]TerrainStoreTileIndex, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.TileIndex()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]TerrainStoreTileIndex, len(value))
	for i, item := range value {
		result[i] = TerrainStoreTileIndex{
			LatIndex: item.LatIndex, LonIndex: item.LonIndex,
			MinLongitudeDeg: item.MinLongitudeDeg, MinLatitudeDeg: item.MinLatitudeDeg,
			MaxLongitudeDeg: item.MaxLongitudeDeg, MaxLatitudeDeg: item.MaxLatitudeDeg,
			LonCount: item.LonCount, LatCount: item.LatCount,
			DataOffset: item.DataOffset, DataLen: item.DataLen, Checksum64: item.Checksum64,
			VerticalDatum: VerticalDatum(item.VerticalDatum),
		}
	}
	return result, nil
}

// ToBytes serializes the complete terrain store into newly owned bytes.
func (t *MMapTerrain) ToBytes() ([]byte, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.ToBytes()
	return value, publicError(err)
}

// Verify hashes and verifies the terrain payload, changing attested provenance
// to verified on success.
func (t *MMapTerrain) Verify() error {
	if t == nil || t.native == nil {
		return ErrClosed
	}
	return publicError(t.native.Verify())
}

// VerticalDatum returns the terrain store's vertical datum.
func (t *MMapTerrain) VerticalDatum() (VerticalDatum, error) {
	if t == nil || t.native == nil {
		return 0, ErrClosed
	}
	value, err := t.native.VerticalDatum()
	return VerticalDatum(value), publicError(err)
}

// VerticalDatum identifies a terrain-store vertical datum.
type VerticalDatum uint32

const (
	// EGM96MSLOrthometric is orthometric height above EGM96 mean sea level.
	EGM96MSLOrthometric VerticalDatum = VerticalDatum(native.VerticalDatumEGM96MSLOrthometricValue)
)

// TerrainGeoidModel selects the geoid tier used for terrain datum conversion.
type TerrainGeoidModel uint32

const (
	// EGM96OneDegree selects the embedded one-degree grid.
	EGM96OneDegree TerrainGeoidModel = TerrainGeoidModel(native.TerrainGeoidModelEGM96OneDegreeValue)
	// EGM96FifteenMinute selects a loaded fifteen-minute WW15MGH.DAC grid.
	EGM96FifteenMinute TerrainGeoidModel = TerrainGeoidModel(native.TerrainGeoidModelEGM96FifteenMinuteValue)
)

// TerrainDatumErrorKind identifies a terrain datum diagnostic.
type TerrainDatumErrorKind uint32

const (
	TerrainDatumErrorNone            TerrainDatumErrorKind = TerrainDatumErrorKind(native.TerrainDatumErrorNoneValue)
	TerrainDatumErrorTerrain         TerrainDatumErrorKind = TerrainDatumErrorKind(native.TerrainDatumErrorTerrainValue)
	TerrainDatumErrorGeoid           TerrainDatumErrorKind = TerrainDatumErrorKind(native.TerrainDatumErrorGeoidValue)
	TerrainDatumErrorIO              TerrainDatumErrorKind = TerrainDatumErrorKind(native.TerrainDatumErrorIOValue)
	TerrainDatumErrorMissingEGM96DAC TerrainDatumErrorKind = TerrainDatumErrorKind(native.TerrainDatumErrorMissingEGM96DACValue)
)

// TerrainStoreErrorKind identifies a terrain-store diagnostic.
type TerrainStoreErrorKind uint32

const (
	TerrainStoreErrorNone                     TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorNoneValue)
	TerrainStoreErrorIO                       TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorIOValue)
	TerrainStoreErrorParse                    TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorParseValue)
	TerrainStoreErrorUnsupportedVersion       TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorUnsupportedVersionValue)
	TerrainStoreErrorUnsupportedDatum         TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorUnsupportedDatumValue)
	TerrainStoreErrorDuplicateTile            TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorDuplicateTileValue)
	TerrainStoreErrorChecksum                 TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorChecksumValue)
	TerrainStoreErrorTileIDMismatch           TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorTileIDMismatchValue)
	TerrainStoreErrorAttestedChecksumMismatch TerrainStoreErrorKind = TerrainStoreErrorKind(native.TerrainStoreErrorAttestedChecksumMismatchValue)
)

// TerrainDatumError is the typed terrain datum detail captured from a failed
// native operation.
type TerrainDatumError struct {
	Kind        TerrainDatumErrorKind
	Path        string
	Message     string
	Remediation string
}

// TerrainStoreError is the typed terrain-store detail captured from a failed
// native operation.
type TerrainStoreError struct {
	Kind             TerrainStoreErrorKind
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

// TerrainStoreChecksum64 returns the native FNV-1a checksum of data.
func TerrainStoreChecksum64(data []byte) (uint64, error) {
	value, err := native.TerrainStoreChecksum64(data)
	return value, publicError(err)
}

// VerticalDatumLabel returns the native vertical-datum label.
func VerticalDatumLabel(datum VerticalDatum) (string, error) {
	if !validVerticalDatum(datum) {
		return "", errors.New("sidereon: invalid vertical datum")
	}
	value, err := native.VerticalDatumLabel(uint32(datum))
	return string(value), publicError(err)
}

// TerrainGeoidModelLabel returns the native terrain geoid-model label.
func TerrainGeoidModelLabel(model TerrainGeoidModel) (string, error) {
	if !validTerrainGeoidModel(model) {
		return "", errors.New("sidereon: invalid terrain geoid model")
	}
	value, err := native.TerrainGeoidModelLabel(uint32(model))
	return string(value), publicError(err)
}

// AtmosphereInput contains the complete NRLMSISE-00 input, including explicit
// presence for local solar time and Ap history.
type AtmosphereInput struct {
	Year         int
	DayOfYear    int
	Second       float64
	AltitudeKm   float64
	LatitudeDeg  float64
	LongitudeDeg float64
	HasLST       bool
	LST          float64
	F107         float64
	F107A        float64
	Ap           float64
	HasApArray   bool
	ApArray      [7]float64
}

// AtmosphereOutput is the NRLMSISE-00 density and temperature result.
type AtmosphereOutput struct {
	DensityKgM3  float64
	TemperatureK float64
}

func atmosphereInputToNative(value AtmosphereInput) native.AtmosphereInput {
	return native.AtmosphereInput{Year: value.Year, DayOfYear: value.DayOfYear, Second: value.Second, AltitudeKm: value.AltitudeKm, LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg, HasLST: value.HasLST, LST: value.LST, F107: value.F107, F107A: value.F107A, Ap: value.Ap, HasApArray: value.HasApArray, ApArray: value.ApArray}
}

// DefaultAtmosphereInput returns the native quiet-Sun defaults.
func DefaultAtmosphereInput() AtmosphereInput {
	value := native.AtmosphereInputDefault()
	return AtmosphereInput{Year: value.Year, DayOfYear: value.DayOfYear, Second: value.Second, AltitudeKm: value.AltitudeKm, LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg, HasLST: value.HasLST, LST: value.LST, F107: value.F107, F107A: value.F107A, Ap: value.Ap, HasApArray: value.HasApArray, ApArray: value.ApArray}
}

// NRLMSISE00 evaluates the native neutral-atmosphere model.
func NRLMSISE00(input AtmosphereInput) (AtmosphereOutput, error) {
	value, err := native.AtmosphereNRLMSISE00(atmosphereInputToNative(input))
	return AtmosphereOutput{DensityKgM3: value.DensityKgM3, TemperatureK: value.TemperatureK}, publicError(err)
}

// Met is surface meteorology for troposphere calculations.
type Met struct {
	PressureHPa      float64
	TemperatureK     float64
	RelativeHumidity float64
}

// MappingFactors contains dry and wet troposphere mapping factors.
type MappingFactors struct {
	Dry float64
	Wet float64
}

// TropoMappingErrorKind identifies a checked troposphere mapping failure.
type TropoMappingErrorKind uint32

const (
	TropoMappingErrorNone         TropoMappingErrorKind = TropoMappingErrorKind(native.TropoMappingErrorNoneValue)
	TropoMappingErrorLowElevation TropoMappingErrorKind = TropoMappingErrorKind(native.TropoMappingErrorLowElevationValue)
	TropoMappingErrorInvalidInput TropoMappingErrorKind = TropoMappingErrorKind(native.TropoMappingErrorInvalidInputValue)
)

// TropoMappingError contains typed low-elevation diagnostic fields.
type TropoMappingError struct {
	Kind            TropoMappingErrorKind
	ElevationRad    float64
	MinElevationRad float64
}

// ZenithDelay contains hydrostatic and wet zenith delays in metres.
type ZenithDelay struct {
	DryM float64
	WetM float64
}

// DefaultMet returns the native standard-atmosphere meteorology.
func DefaultMet() (Met, error) {
	value, err := native.DefaultMet()
	return Met{PressureHPa: value.PressureHPa, TemperatureK: value.TemperatureK, RelativeHumidity: value.RelativeHumidity}, publicError(err)
}

// TropoMappingFactors returns native Niell dry and wet mapping factors.
func TropoMappingFactors(elevationRad float64, receiver Geodetic, scale TimeScale, jdWhole, jdFraction float64) (MappingFactors, error) {
	if !validTropoTimeScale(scale) {
		return MappingFactors{}, errors.New("sidereon: invalid troposphere time scale")
	}
	value, err := native.TropoMappingFactors(elevationRad, native.Geodetic{LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM}, uint32(scale), jdWhole, jdFraction)
	return MappingFactors{Dry: value.Dry, Wet: value.Wet}, publicError(err)
}

// TropoMappingFactorsChecked returns mapping factors and low-elevation detail.
func TropoMappingFactorsChecked(elevationRad float64, receiver Geodetic, scale TimeScale, jdWhole, jdFraction float64) (MappingFactors, TropoMappingError, error) {
	if !validTropoTimeScale(scale) {
		return MappingFactors{}, TropoMappingError{Kind: TropoMappingErrorInvalidInput}, errors.New("sidereon: invalid troposphere time scale")
	}
	value, detail, err := native.TropoMappingFactorsChecked(elevationRad, native.Geodetic{LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM}, uint32(scale), jdWhole, jdFraction)
	return MappingFactors{Dry: value.Dry, Wet: value.Wet}, TropoMappingError{Kind: TropoMappingErrorKind(detail.Kind), ElevationRad: detail.ElevationRad, MinElevationRad: detail.MinElevationRad}, publicError(err)
}

// TropoZenithDelay returns native Saastamoinen zenith delays.
func TropoZenithDelay(receiver Geodetic, met Met) (ZenithDelay, error) {
	value, err := native.TropoZenithDelay(native.Geodetic{LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM}, native.Met{PressureHPa: met.PressureHPa, TemperatureK: met.TemperatureK, RelativeHumidity: met.RelativeHumidity})
	return ZenithDelay{DryM: value.DryM, WetM: value.WetM}, publicError(err)
}

// TropoSlantDelay returns total native slant delay in metres.
func TropoSlantDelay(elevationRad float64, receiver Geodetic, met Met, scale TimeScale, jdWhole, jdFraction float64) (float64, error) {
	if !validTropoTimeScale(scale) {
		return 0, errors.New("sidereon: invalid troposphere time scale")
	}
	value, err := native.TropoSlantDelay(elevationRad, native.Geodetic{LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM}, native.Met{PressureHPa: met.PressureHPa, TemperatureK: met.TemperatureK, RelativeHumidity: met.RelativeHumidity}, uint32(scale), jdWhole, jdFraction)
	return value, publicError(err)
}

// LinkBudget contains the native RF link-budget terms.
type LinkBudget struct {
	EIRPdBW         float64
	FSPLdB          float64
	ReceiverGTdBK   float64
	OtherLossesdB   float64
	RequiredCN0dBHz float64
}

// RFFSPL computes free-space path loss in dB.
func RFFSPL(distanceKm, frequencyMHz float64) (float64, error) {
	value, err := native.RFFSPL(distanceKm, frequencyMHz)
	return value, publicError(err)
}

// RFEIRP computes effective isotropic radiated power in dBW.
func RFEIRP(txPowerdBm, txAntennaGaindBi float64) (float64, error) {
	value, err := native.RFEIRP(txPowerdBm, txAntennaGaindBi)
	return value, publicError(err)
}

// RFCN0 computes received carrier-to-noise density in dB-Hz.
func RFCN0(eirpdBW, fspldB, receiverGTdBK, otherLossesdB float64) (float64, error) {
	value, err := native.RFCN0(eirpdBW, fspldB, receiverGTdBK, otherLossesdB)
	return value, publicError(err)
}

// RFLinkMargin computes received minus required C/N0 in dB.
func RFLinkMargin(budget LinkBudget) (float64, error) {
	value, err := native.RFLinkMargin(native.LinkBudget{EIRPdBW: budget.EIRPdBW, FSPLdB: budget.FSPLdB, ReceiverGTdBK: budget.ReceiverGTdBK, OtherLossesdB: budget.OtherLossesdB, RequiredCN0dBHz: budget.RequiredCN0dBHz})
	return value, publicError(err)
}

// RFWavelength returns wavelength in metres for frequencyHz.
func RFWavelength(frequencyHz float64) (float64, error) {
	value, err := native.RFWavelength(frequencyHz)
	return value, publicError(err)
}

// RFDishGain computes parabolic-dish gain in dBi.
func RFDishGain(diameterM, frequencyHz, efficiency float64) (float64, error) {
	value, err := native.RFDishGain(diameterM, frequencyHz, efficiency)
	return value, publicError(err)
}
