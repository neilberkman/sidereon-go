//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const (
	DtedInterpolationNearestPostingValue = uint32(0)
	DtedInterpolationBilinearValue       = uint32(1)

	DigestProvenanceVerifiedValue = uint32(0)
	DigestProvenanceAttestedValue = uint32(1)

	VerticalDatumEGM96MSLOrthometricValue = uint32(1)

	TerrainGeoidModelEGM96OneDegreeValue     = uint32(0)
	TerrainGeoidModelEGM96FifteenMinuteValue = uint32(1)

	GeofenceErrorNoneValue           = uint32(0)
	GeofenceErrorTooFewVerticesValue = uint32(1)
	GeofenceErrorInvalidInputValue   = uint32(2)
	GeofenceErrorGeodesicValue       = uint32(3)
	GeofenceErrorDOPValue            = uint32(4)
	GeofenceErrorMetricsValue        = uint32(5)

	GeofenceENUCovarianceM2Value  = uint32(0)
	GeofenceECEFCovarianceM2Value = uint32(1)
	GeofenceCEPRadiusMValue       = uint32(2)

	GeofenceBoundaryNormalValue   = uint32(0)
	GeofencePlanarQuadratureValue = uint32(1)

	GeofenceEnteredValue = uint32(0)
	GeofenceLeftValue    = uint32(1)

	ObservabilityRankDeficientValue  = uint32(0)
	ObservabilityZeroRedundancyValue = uint32(1)
	ObservabilityWeakValue           = uint32(2)
	ObservabilityNominalValue        = uint32(3)

	ProjVgridshiftArithmeticSeparateMultiplyAddValue = uint32(0)
	ProjVgridshiftArithmeticFusedMultiplyAddValue    = uint32(1)

	ProjVgridshiftErrorNoneValue                  = uint32(0)
	ProjVgridshiftErrorNonFiniteCoordinateValue   = uint32(1)
	ProjVgridshiftErrorCoordinateOutsideGridValue = uint32(2)

	ProjVgridshiftCoordinateNoneValue      = uint32(0)
	ProjVgridshiftCoordinateLatitudeValue  = uint32(1)
	ProjVgridshiftCoordinateLongitudeValue = uint32(2)

	TerrainDatumErrorNoneValue            = uint32(0)
	TerrainDatumErrorTerrainValue         = uint32(1)
	TerrainDatumErrorGeoidValue           = uint32(2)
	TerrainDatumErrorIOValue              = uint32(3)
	TerrainDatumErrorMissingEGM96DACValue = uint32(4)

	TerrainStoreErrorNoneValue                     = uint32(0)
	TerrainStoreErrorIOValue                       = uint32(1)
	TerrainStoreErrorParseValue                    = uint32(2)
	TerrainStoreErrorUnsupportedVersionValue       = uint32(3)
	TerrainStoreErrorUnsupportedDatumValue         = uint32(4)
	TerrainStoreErrorDuplicateTileValue            = uint32(5)
	TerrainStoreErrorChecksumValue                 = uint32(6)
	TerrainStoreErrorTileIDMismatchValue           = uint32(7)
	TerrainStoreErrorAttestedChecksumMismatchValue = uint32(8)

	TropoMappingErrorNoneValue         = uint32(0)
	TropoMappingErrorLowElevationValue = uint32(1)
	TropoMappingErrorInvalidInputValue = uint32(2)

	Egm2008GridSpacingOneMinuteValue          = uint32(0)
	Egm2008GridSpacingTwoPointFiveMinuteValue = uint32(1)
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

type GeoidGrid struct{}
type EGM96FifteenMinuteGeoid struct{}

func GeoidGridFromText([]byte) (*GeoidGrid, error)         { return nil, unavailable() }
func GeoidGridFromPROJEGM96GTX([]byte) (*GeoidGrid, error) { return nil, unavailable() }
func GeoidGridFromEGM96DAC([]byte) (*GeoidGrid, error)     { return nil, unavailable() }
func GeoidGridFromEGM2008Raster([]byte, uint32) (*GeoidGrid, error) {
	return nil, unavailable()
}
func GeoidGridFromEGM2008RasterWindow([]byte, Egm2008RasterWindow) (*GeoidGrid, error) {
	return nil, unavailable()
}
func NewGeoidGrid(float64, float64, float64, float64, int, int, []float64) (*GeoidGrid, error) {
	return nil, unavailable()
}
func (*GeoidGrid) Close() error { return nil }
func (*GeoidGrid) UndulationDeg(float64, float64) (float64, error) {
	return 0, unavailable()
}
func (*GeoidGrid) UndulationRad(float64, float64) (float64, error) {
	return 0, unavailable()
}
func (*GeoidGrid) UndulationPROJRad(float64, float64, uint32) (float64, ProjVGridshiftError, error) {
	return 0, ProjVGridshiftError{}, unavailable()
}
func (*GeoidGrid) UndulationsDeg([]GeoidPoint) ([]float64, error) { return nil, unavailable() }
func (*GeoidGrid) UndulationsRad([]GeoidPoint) ([]float64, error) { return nil, unavailable() }
func (*GeoidGrid) OrthometricHeightRad(float64, float64, float64) (float64, error) {
	return 0, unavailable()
}
func (*GeoidGrid) EllipsoidalHeightRad(float64, float64, float64) (float64, error) {
	return 0, unavailable()
}
func EGM96FifteenMinuteGeoidFromBytes([]byte) (*EGM96FifteenMinuteGeoid, error) {
	return nil, unavailable()
}
func EGM96FifteenMinuteGeoidFromPath(string) (*EGM96FifteenMinuteGeoid, error) {
	return nil, unavailable()
}
func (*EGM96FifteenMinuteGeoid) Close() error           { return nil }
func GeoidUndulation(float64, float64) (float64, error) { return 0, unavailable() }
func OrthometricHeight(float64, float64, float64) (float64, error) {
	return 0, unavailable()
}
func EllipsoidalHeight(float64, float64, float64) (float64, error) {
	return 0, unavailable()
}
func EGM96Undulation(float64, float64) (float64, error) { return 0, unavailable() }
func EGM96OrthometricHeight(float64, float64, float64) (float64, error) {
	return 0, unavailable()
}
func EGM96EllipsoidalHeight(float64, float64, float64) (float64, error) {
	return 0, unavailable()
}
func GeoidUndulationsRad([]GeoidPoint) ([]float64, error) { return nil, unavailable() }
func GeoidUndulationsDeg([]GeoidPoint) ([]float64, error) { return nil, unavailable() }
func EGM96UndulationsRad([]GeoidPoint) ([]float64, error) { return nil, unavailable() }
func EGM96UndulationsDeg([]GeoidPoint) ([]float64, error) { return nil, unavailable() }

type DtedLookupOptions struct{ Interpolation uint32 }
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
type DtedTerrain struct{}
type DtedTile struct{}

func DefaultDTEDLookupOptions() (DtedLookupOptions, error) { return DtedLookupOptions{}, unavailable() }
func DtedTerrainNew(string) (*DtedTerrain, error)          { return nil, unavailable() }
func (*DtedTerrain) Close() error                          { return nil }
func (*DtedTerrain) HeightM(float64, float64) (float64, error) {
	return 0, unavailable()
}
func (*DtedTerrain) HeightMWithOptions(float64, float64, DtedLookupOptions) (float64, error) {
	return 0, unavailable()
}
func (*DtedTerrain) HeightBatch([]LonLatDeg, DtedLookupOptions) ([]DtedHeightResult, error) {
	return nil, unavailable()
}
func DtedTileLoad(string) (*DtedTile, error) { return nil, unavailable() }
func (*DtedTile) Close() error               { return nil }
func (*DtedTile) Elevation(float64, float64) (int16, error) {
	return 0, unavailable()
}
func DtedTileListToMmapStore([]DtedTileListEntry) ([]byte, error)    { return nil, unavailable() }
func WriteDtedTileListToMmapStore([]DtedTileListEntry, string) error { return unavailable() }
func DtedTreeToMmapStore(string) ([]byte, error)                     { return nil, unavailable() }
func WriteDtedTreeToMmapStore(string, string) error                  { return unavailable() }
func DtedInterpolationLabel(uint32) ([]byte, error)                  { return nil, unavailable() }

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

func LastTerrainDatumError() (TerrainDatumError, error) { return TerrainDatumError{}, unavailable() }
func LastTerrainStoreError() (TerrainStoreError, error) { return TerrainStoreError{}, unavailable() }
func TerrainStoreChecksum64([]byte) (uint64, error)     { return 0, unavailable() }
func VerticalDatumLabel(uint32) ([]byte, error)         { return nil, unavailable() }
func TerrainGeoidModelLabel(uint32) ([]byte, error)     { return nil, unavailable() }

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
type MmapTerrain struct{}

func MmapTerrainFromBytes([]byte) (*MmapTerrain, error) { return nil, unavailable() }
func MmapTerrainFromVec([]byte) (*MmapTerrain, error)   { return nil, unavailable() }
func MmapTerrainFromPath(string) (*MmapTerrain, error)  { return nil, unavailable() }
func MmapTerrainFromPathAttested(string, uint64) (*MmapTerrain, error) {
	return nil, unavailable()
}
func (*MmapTerrain) Close() error                              { return nil }
func (*MmapTerrain) Checksum64() (uint64, error)               { return 0, unavailable() }
func (*MmapTerrain) DigestProvenance() (uint32, error)         { return 0, unavailable() }
func (*MmapTerrain) HeightM(float64, float64) (float64, error) { return 0, unavailable() }
func (*MmapTerrain) HeightMWithOptions(float64, float64, DtedLookupOptions) (float64, error) {
	return 0, unavailable()
}
func (*MmapTerrain) OrthometricHeightM(float64, float64) (float64, error) {
	return 0, unavailable()
}
func (*MmapTerrain) OrthometricHeightMWithOptions(float64, float64, DtedLookupOptions) (float64, error) {
	return 0, unavailable()
}
func (*MmapTerrain) EllipsoidalHeightM(float64, float64) (float64, error) {
	return 0, unavailable()
}
func (*MmapTerrain) EllipsoidalHeightMWithOptions(float64, float64, DtedLookupOptions) (float64, error) {
	return 0, unavailable()
}
func (*MmapTerrain) EllipsoidalHeightMWithModel(float64, float64, DtedLookupOptions, uint32, *EGM96FifteenMinuteGeoid) (float64, error) {
	return 0, unavailable()
}
func (*MmapTerrain) HeightBatch([]LonLatDeg, DtedLookupOptions) ([]MmapTerrainHeightResult, error) {
	return nil, unavailable()
}
func (*MmapTerrain) OrthometricHeightBatch([]LonLatDeg, DtedLookupOptions) ([]MmapTerrainHeightResult, error) {
	return nil, unavailable()
}
func (*MmapTerrain) TileIndex() ([]TerrainStoreTileIndex, error) { return nil, unavailable() }
func (*MmapTerrain) ToBytes() ([]byte, error)                    { return nil, unavailable() }
func (*MmapTerrain) Verify() error                               { return unavailable() }
func (*MmapTerrain) VerticalDatum() (uint32, error)              { return 0, unavailable() }

type AntennaPco struct {
	NorthM float64
	EastM  float64
	UpM    float64
}
type Antenna struct{}
type ANTEX struct{}

func ParseANTEX([]byte) (*ANTEX, error)               { return nil, unavailable() }
func (*ANTEX) Close() error                           { return nil }
func (*ANTEX) AntennaCount() (int, error)             { return 0, unavailable() }
func (*ANTEX) Antenna(string) (*Antenna, bool, error) { return nil, false, unavailable() }
func (*ANTEX) Encode() ([]byte, error)                { return nil, unavailable() }
func (*Antenna) Close() error                         { return nil }
func (*Antenna) PCO(string) (AntennaPco, error)       { return AntennaPco{}, unavailable() }
func (*Antenna) PCV(string, float64, bool, float64) (float64, error) {
	return 0, unavailable()
}

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
type AtmosphereOutput struct {
	DensityKgM3  float64
	TemperatureK float64
}

func AtmosphereInputDefault() AtmosphereInput { return AtmosphereInput{} }
func AtmosphereNRLMSISE00(AtmosphereInput) (AtmosphereOutput, error) {
	return AtmosphereOutput{}, unavailable()
}

type Met struct {
	PressureHPa      float64
	TemperatureK     float64
	RelativeHumidity float64
}
type MappingFactors struct {
	Dry float64
	Wet float64
}
type TropoMappingError struct {
	Kind            uint32
	ElevationRad    float64
	MinElevationRad float64
}
type ZenithDelay struct {
	DryM float64
	WetM float64
}

func DefaultMet() (Met, error) { return Met{}, unavailable() }
func TropoMappingFactors(float64, Geodetic, uint32, float64, float64) (MappingFactors, error) {
	return MappingFactors{}, unavailable()
}
func TropoMappingFactorsChecked(float64, Geodetic, uint32, float64, float64) (MappingFactors, TropoMappingError, error) {
	return MappingFactors{}, TropoMappingError{}, unavailable()
}
func TropoZenithDelay(Geodetic, Met) (ZenithDelay, error) { return ZenithDelay{}, unavailable() }
func TropoSlantDelay(float64, Geodetic, Met, uint32, float64, float64) (float64, error) {
	return 0, unavailable()
}

type LinkBudget struct {
	EIRPdBW         float64
	FSPLdB          float64
	ReceiverGTdBK   float64
	OtherLossesdB   float64
	RequiredCN0dBHz float64
}

func RFFSPL(float64, float64) (float64, error)                  { return 0, unavailable() }
func RFEIRP(float64, float64) (float64, error)                  { return 0, unavailable() }
func RFCN0(float64, float64, float64, float64) (float64, error) { return 0, unavailable() }
func RFLinkMargin(LinkBudget) (float64, error)                  { return 0, unavailable() }
func RFWavelength(float64) (float64, error)                     { return 0, unavailable() }
func RFDishGain(float64, float64, float64) (float64, error)     { return 0, unavailable() }

type SpaceWeather struct {
	F107  float64
	F107A float64
	Ap    float64
}
type SpaceWeatherObservationClass uint32

const (
	SpaceWeatherObservationObservedValue         SpaceWeatherObservationClass = 0
	SpaceWeatherObservationInterpolatedValue     SpaceWeatherObservationClass = 1
	SpaceWeatherObservationDailyPredictedValue   SpaceWeatherObservationClass = 2
	SpaceWeatherObservationMonthlyPredictedValue SpaceWeatherObservationClass = 3
)

type SpaceWeatherCoverage struct {
	FirstJ2000S              float64
	HasLastObservedJ2000S    bool
	LastObservedJ2000S       float64
	HasLastDailyPredictedS   bool
	LastDailyPredictedJ2000S float64
	EndJ2000S                float64
}
type SpaceWeatherDay struct {
	Year               int
	Month              uint8
	Day                uint8
	Class              SpaceWeatherObservationClass
	HasBSRN            bool
	BSRN               uint16
	HasND              bool
	ND                 uint8
	HasKp              [8]bool
	Kp10               [8]uint16
	HasKpSum10         bool
	KpSum10            uint16
	HasAp              [8]bool
	Ap8                [8]uint16
	HasApAvg           bool
	ApAvg              uint16
	HasCP10            bool
	CP10               uint8
	HasC9              bool
	C9                 uint8
	HasISN             bool
	ISN                uint16
	HasFluxQualifier   bool
	FluxQualifier      uint8
	HasF107Obs         bool
	F107Obs            float64
	HasF107Adj         bool
	F107Adj            float64
	HasF107ObsCenter81 bool
	F107ObsCenter81    float64
	HasF107ObsLast81   bool
	F107ObsLast81      float64
	HasF107AdjCenter81 bool
	F107AdjCenter81    float64
	HasF107AdjLast81   bool
	F107AdjLast81      float64
}
type SpaceWeatherSample struct {
	Weather     SpaceWeather
	Class       SpaceWeatherObservationClass
	ApDefaulted bool
}
type SpaceWeatherPolicy struct {
	AllowInterpolated     bool
	AllowDailyPredicted   bool
	AllowMonthlyPredicted bool
	RequireGeomagnetic    bool
}
type SpaceWeatherTableSummary struct {
	DayCount     int
	MonthlyCount int
	SkipCount    int
	WarningCount int
}
type SpaceWeatherTable struct{}

func DefaultSpaceWeather() (SpaceWeather, error)                { return SpaceWeather{}, unavailable() }
func ParseSpaceWeatherTable([]byte) (*SpaceWeatherTable, error) { return nil, unavailable() }
func ParseSpaceWeatherCSV([]byte) (*SpaceWeatherTable, error)   { return nil, unavailable() }
func ParseSpaceWeatherTXT([]byte) (*SpaceWeatherTable, error)   { return nil, unavailable() }
func (*SpaceWeatherTable) Close() error                         { return nil }
func (*SpaceWeatherTable) Days() ([]SpaceWeatherDay, error)     { return nil, unavailable() }
func (*SpaceWeatherTable) Monthly() ([]SpaceWeatherDay, error)  { return nil, unavailable() }
func (*SpaceWeatherTable) Day(int, uint8, uint8) (SpaceWeatherDay, bool, error) {
	return SpaceWeatherDay{}, false, unavailable()
}
func (*SpaceWeatherTable) Coverage() (SpaceWeatherCoverage, error) {
	return SpaceWeatherCoverage{}, unavailable()
}
func (*SpaceWeatherTable) Summary() (SpaceWeatherTableSummary, error) {
	return SpaceWeatherTableSummary{}, unavailable()
}
func (*SpaceWeatherTable) APArrayAt(float64) ([7]float64, error) {
	return [7]float64{}, unavailable()
}
func (*SpaceWeatherTable) SampleAt(float64) (SpaceWeatherSample, error) {
	return SpaceWeatherSample{}, unavailable()
}
func (*SpaceWeatherTable) SampleAtWithPolicy(float64, SpaceWeatherPolicy) (SpaceWeatherSample, error) {
	return SpaceWeatherSample{}, unavailable()
}
func (*SpaceWeatherTable) SpaceWeatherAt(float64) (SpaceWeather, error) {
	return SpaceWeather{}, unavailable()
}
func (*SpaceWeatherTable) ToCSV() ([]byte, error) { return nil, unavailable() }
func (*SpaceWeatherTable) ToTXT() ([]byte, error) { return nil, unavailable() }

type GeofenceUncertainty struct {
	Kind       uint32
	Covariance [9]float64
	RadiusM    float64
}
type GeofenceProbabilityOptions struct{ Method uint32 }
type GeofencePositionEstimate struct {
	Position    Geodetic
	Uncertainty GeofenceUncertainty
}
type GeofenceHysteresis struct {
	EnterConfidence float64
	LeaveConfidence float64
}
type GeofenceCrossingEvent struct {
	SampleIndex       int
	Kind              uint32
	InsideProbability float64
}
type Geofence struct{}

func DefaultGeofenceProbabilityOptions() (GeofenceProbabilityOptions, error) {
	return GeofenceProbabilityOptions{}, unavailable()
}
func DefaultGeofenceHysteresis() (GeofenceHysteresis, error) {
	return GeofenceHysteresis{}, unavailable()
}
func GeofenceCreate([]Geodetic) (*Geofence, uint32, error) { return nil, 0, unavailable() }
func (*Geofence) Close() error                             { return nil }
func (*Geofence) Contains(Geodetic) (bool, uint32, error)  { return false, 0, unavailable() }
func (*Geofence) DistanceToBoundary(Geodetic) (float64, uint32, error) {
	return 0, 0, unavailable()
}
func (*Geofence) ContainmentProbability(Geodetic, GeofenceUncertainty) (float64, uint32, error) {
	return 0, 0, unavailable()
}
func (*Geofence) ContainmentProbabilityWithOptions(Geodetic, GeofenceUncertainty, GeofenceProbabilityOptions) (float64, uint32, error) {
	return 0, 0, unavailable()
}
func (*Geofence) CrossingProbability([]GeofencePositionEstimate, GeofenceHysteresis) ([]GeofenceCrossingEvent, uint32, error) {
	return nil, 0, unavailable()
}
func (*Geofence) CrossingProbabilityWithOptions([]GeofencePositionEstimate, GeofenceHysteresis, GeofenceProbabilityOptions) ([]GeofenceCrossingEvent, uint32, error) {
	return nil, 0, unavailable()
}
func ObservabilityTierLabel(uint32) ([]byte, error) { return nil, unavailable() }
