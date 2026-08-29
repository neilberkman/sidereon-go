//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type Geodetic struct {
	LatitudeRad  float64
	LongitudeRad float64
	HeightM      float64
}

type ECEF struct {
	X float64
	Y float64
	Z float64
}

type LineOfSight struct {
	EX float64
	EY float64
	EZ float64
}

type Dop struct {
	GDOP float64
	PDOP float64
	HDOP float64
	VDOP float64
	TDOP float64
}

type GeodesicDirectResult struct {
	LatitudeDeg     float64
	LongitudeDeg    float64
	FinalAzimuthDeg float64
}

type GeodesicInverseResult struct {
	DistanceM         float64
	InitialAzimuthDeg float64
	FinalAzimuthDeg   float64
}

type ErrorEllipse2 struct {
	Confidence     float64
	ChiSquareScale float64
	SemiMajor      float64
	SemiMinor      float64
	OrientationRad float64
}

type ErrorEllipse struct {
	SemiMajorM     float64
	SemiMinorM     float64
	OrientationRad float64
}

type PercentileRadius struct {
	Probability float64
	RadiusM     float64
	ApproxM     float64
	ApproxValid bool
}

type PositionErrorMetrics struct {
	Ellipse ErrorEllipse
	SigmaEM float64
	SigmaNM float64
	SigmaUM float64
	CEP     PercentileRadius
	R95     PercentileRadius
	R99     PercentileRadius
	DRMS    float64
	TwoDRMS float64
	VEP     float64
	SEP     PercentileRadius
	MRSE    float64
}

type TimeScales struct {
	JDWhole     float64
	UT1Fraction float64
	TTFraction  float64
	TDBFraction float64
	JDUT1       float64
	JDTT        float64
	JDTDB       float64
}

type DopplerRangeRate struct {
	RangeRateKmS float64
	DopplerRatio float64
}

type DopplerShift struct {
	RangeRateKmS float64
	DopplerHz    float64
	DopplerRatio float64
}

type CivilDateTime struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second float64
}

type JulianDate struct {
	Whole    float64
	Fraction float64
}

type GNSSWeekSeconds struct {
	Week          float64
	SecondsOfWeek float64
}

type NISGate struct {
	NIS       float64
	Threshold float64
	InGate    bool
	DOF       uint64
}

type CarrierPair struct {
	Band1 int
	Band2 int
}

func unavailableValue() error { return unavailable() }

func GeodeticToECEF(Geodetic) (ECEF, error) { return ECEF{}, unavailableValue() }
func ECEFToGeodetic(ECEF) (Geodetic, error) { return Geodetic{}, unavailableValue() }
func LineOfSightFromAzEl(float64, float64, Geodetic) (LineOfSight, error) {
	return LineOfSight{}, unavailableValue()
}
func DOP([]LineOfSight, []float64, Geodetic, uint32) (Dop, error) {
	return Dop{}, unavailableValue()
}
func GeodesicDirect(float64, float64, float64, float64) (GeodesicDirectResult, error) {
	return GeodesicDirectResult{}, unavailableValue()
}
func GeodesicInverse(float64, float64, float64, float64) (GeodesicInverseResult, error) {
	return GeodesicInverseResult{}, unavailableValue()
}
func AngularSeparationCoords(float64, float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func AngularSeparation([3]float64, [3]float64) (float64, error) { return 0, unavailableValue() }
func EarthAngularRadius([3]float64) (float64, error)            { return 0, unavailableValue() }
func EclipseShadowFraction([3]float64, [3]float64, *uint32) (float64, error) {
	return 0, unavailableValue()
}
func EclipseStatus([3]float64, [3]float64, *uint32) (int, error) {
	return 0, unavailableValue()
}

func CivilToJ2000(CivilDateTime) (float64, error) { return 0, unavailableValue() }
func CivilToGPS(CivilDateTime) (float64, bool, error) {
	return 0, false, unavailableValue()
}
func J2000ToCivil(int64) (CivilDateTime, error) { return CivilDateTime{}, unavailableValue() }
func InstantFromUTC(CivilDateTime) (JulianDate, float64, error) {
	return JulianDate{}, 0, unavailableValue()
}
func SplitJulianToJ2000(JulianDate) (float64, error) { return 0, unavailableValue() }
func LeapSeconds(int, int, int) (float64, error)     { return 0, unavailableValue() }
func GPSUTCOffset(float64) (float64, error)          { return 0, unavailableValue() }
func TAIUTCOffset(float64) (float64, error)          { return 0, unavailableValue() }
func GNSSSecondsOfWeek(CivilDateTime) (float64, error) {
	return 0, unavailableValue()
}
func GNSSWeekAndSeconds(float64) (GNSSWeekSeconds, error) {
	return GNSSWeekSeconds{}, unavailableValue()
}
func TimeScalesFromUTC(CivilDateTime) (TimeScales, error) {
	return TimeScales{}, unavailableValue()
}
func TimeScaleLabel(uint32) ([]byte, error) { return nil, unavailableValue() }
func TimeScaleOffset(uint32, uint32) (float64, error) {
	return 0, unavailableValue()
}
func TimeScaleOffsetAt(uint32, uint32, float64) (float64, error) {
	return 0, unavailableValue()
}
func CarrierBandLabel(uint32) ([]byte, error) { return nil, unavailableValue() }
func GNSSSystemLabel(uint32) ([]byte, error)  { return nil, unavailableValue() }
func Frequency(uint32, uint32) (float64, error) {
	return 0, unavailableValue()
}
func Wavelength(uint32, uint32) (float64, error) {
	return 0, unavailableValue()
}
func GLONASSG1Frequency(int8) (float64, error) { return 0, unavailableValue() }
func DefaultSPPFrequency(uint32) (float64, error) {
	return 0, unavailableValue()
}
func DefaultIonosphereFreePair(uint32) (CarrierPair, bool, error) {
	return CarrierPair{}, false, unavailableValue()
}
func PhaseMeters(float64, float64) (float64, error) { return 0, unavailableValue() }
func CodeMinusCarrier(float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func GeometryFree(float64, float64) (float64, error) { return 0, unavailableValue() }
func MelbourneWubbena(float64, float64, float64, float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func NarrowLaneCode(float64, float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func WideLaneCycles(float64, float64, float64, float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func WideLaneWavelength(float64, float64) (float64, error) { return 0, unavailableValue() }
func IonosphericGamma(float64, float64) (float64, error)   { return 0, unavailableValue() }
func CombinationNoiseAmplification(float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func IonosphereFreeCode(float64, float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func IonosphereFreePhaseCycles(float64, float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func IonosphereFreePhaseMeters(float64, float64, float64, float64) (float64, error) {
	return 0, unavailableValue()
}
func ComputeDopplerRangeRate([3]float64, [3]float64, float64, float64, float64, TimeScales) (DopplerRangeRate, error) {
	return DopplerRangeRate{}, unavailableValue()
}
func ComputeDopplerShift([3]float64, [3]float64, float64, float64, float64, TimeScales, float64) (DopplerShift, error) {
	return DopplerShift{}, unavailableValue()
}

func ErrorEllipse2x2([4]float64, float64) (ErrorEllipse2, error) {
	return ErrorEllipse2{}, unavailableValue()
}
func ErrorMetricsFromENU([3][3]float64) (PositionErrorMetrics, uint32, error) {
	return PositionErrorMetrics{}, 0, unavailableValue()
}
func ErrorMetricsFromECEF([3][3]float64, Geodetic) (PositionErrorMetrics, uint32, error) {
	return PositionErrorMetrics{}, 0, unavailableValue()
}
func ErrorMetricsFromKinematic([3]float64, [3][3]float64) (PositionErrorMetrics, uint32, error) {
	return PositionErrorMetrics{}, 0, unavailableValue()
}
func ErrorMetricsFromPositionCovariance([3][3]float64, [3][3]float64) (PositionErrorMetrics, uint32, error) {
	return PositionErrorMetrics{}, 0, unavailableValue()
}
func ErrorEllipseFromENU([3][3]float64) (ErrorEllipse, uint32, error) {
	return ErrorEllipse{}, 0, unavailableValue()
}
func HorizontalRadius([3][3]float64, float64) (PercentileRadius, uint32, error) {
	return PercentileRadius{}, 0, unavailableValue()
}
func SphericalRadius([3][3]float64, float64) (PercentileRadius, uint32, error) {
	return PercentileRadius{}, 0, unavailableValue()
}
func VerticalRadius(float64, float64) (float64, uint32, error) {
	return 0, 0, unavailableValue()
}
func NIS(float64, float64) (float64, error)         { return 0, unavailableValue() }
func NISExpected(uint64) (float64, error)           { return 0, unavailableValue() }
func NISThreshold(uint64, float64) (float64, error) { return 0, unavailableValue() }
func ComputeNISGate(float64, float64, uint64, float64) (NISGate, error) {
	return NISGate{}, unavailableValue()
}
func ChiSquareInverse(float64, uint64) (float64, error) { return 0, unavailableValue() }

func PolarMotionMatrix(float64, float64) ([3][3]float64, error) {
	return [3][3]float64{}, unavailableValue()
}
func GCRSToITRSMatrix(TimeScales) ([3][3]float64, error) {
	return [3][3]float64{}, unavailableValue()
}
func ITRSToGCRSMatrix(TimeScales) ([3][3]float64, error) {
	return [3][3]float64{}, unavailableValue()
}
func MatrixVectorMultiply([3][3]float64, [3]float64) ([3]float64, error) {
	return [3]float64{}, unavailableValue()
}
func GCRSToITRS([3]float64, TimeScales, bool) ([3]float64, error) {
	return [3]float64{}, unavailableValue()
}
func ITRSToGCRS([3]float64, TimeScales) ([3]float64, error) {
	return [3]float64{}, unavailableValue()
}
func GCRSToTopocentric([3]float64, float64, float64, float64, TimeScales, bool) ([3]float64, error) {
	return [3]float64{}, unavailableValue()
}
