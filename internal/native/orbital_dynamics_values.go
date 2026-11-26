package native

type AnalyticTheory uint32

type AttitudeMode uint32

type BeiDouManufacturer uint32

type BeiDouOrbitClass uint32

type ClockDatum uint32

type CosineTurnStartAngle uint32

type DEBody uint32

type EmpiricalHarmonic uint32

type FrameAgreementKind uint32

type OrbitScoreGapPolicy uint32

type GNSSBlock uint32

type GravityTideSystem uint32

type ICGEMErrors uint32

type ICGEMFormat uint32

type ICGEMNormalization uint32

type ICGEMTermKind uint32

type OrbitFrame uint32

type OrbitReferencePoint uint32

type OrbitRegime uint32

type RTNAxis uint32

type ScoredProduct uint32

type ShadowFunction uint32

type TriadVelocity uint32

type Ut1Degradation uint32

type AccelerationPartials struct {
	WrtPosition [9]float64
	WrtVelocity [9]float64
}

type AnalyticCartesian struct {
	PositionKm  [3]float64
	VelocityKmS [3]float64
}

type AnalyticKeplerianElements struct {
	AKm            float64
	E              float64
	IRad           float64
	ArgPerigeeRad  float64
	RaanRad        float64
	MeanAnomalyRad float64
}

type AttitudeSample struct {
	PositionM    [3]float64
	VelocityMS   [3]float64
	SunPositionM [3]float64
}

type BrouwerLyddaneMeanElements struct {
	Elements   AnalyticKeplerianElements
	Iterations uint32
}

type CircularElements struct {
	AKm                     float64
	Ex                      float64
	Ey                      float64
	IRad                    float64
	RaanRad                 float64
	MeanLatitudeArgumentRad float64
}

type CosineTurnParameters struct {
	BetaLimitRad       float64
	BetaLimitInclusive bool
	StartAngle         CosineTurnStartAngle
	StartAngleRad      float64
	PeriodS            float64
}

type DECoverage struct {
	StartTDBSeconds float64
	EndTDBSeconds   float64
}

type DEGM struct {
	Present   bool
	ValueM3S2 float64
}

type DEState struct {
	PositionM  [3]float64
	VelocityMS [3]float64
}

type DESunMoon struct {
	Sun  DEState
	Moon DEState
}

type DYBAxes struct {
	D [3]float64
	Y [3]float64
	B [3]float64
}

type EarthOrientation struct {
	TimeScales                   TimeScales
	PolarMotionRad               [2]float64
	EarthRotationVectorITRFRadS  [3]float64
	GCRFToITRFMatrix             [3][3]float64
	ITRFToGCRFMatrix             [3][3]float64
	GCRFToITRFRotationRateMatrix [3][3]float64
	ITRFToGCRFRotationRateMatrix [3][3]float64
	HasIERS2010Quaternion        bool
	IERS2010Quaternion           [4]float64
}

type EcksteinHechlerMeanElements struct {
	Elements   CircularElements
	Iterations uint32
}

type ECOM1Coefficients struct {
	D0 float64
	Dc float64
	Ds float64
	Y0 float64
	Yc float64
	Ys float64
	B0 float64
	Bc float64
	Bs float64
}

type ECOM2D2B1Coefficients struct {
	B0  float64
	B1c float64
	B1s float64
	D0  float64
	D2c float64
	D2s float64
	Y0  float64
}

type ECOM2D4B1Coefficients struct {
	B0  float64
	B1c float64
	B1s float64
	D0  float64
	D2c float64
	D4c float64
	D2s float64
	D4s float64
	Y0  float64
}

type ECOMGeometry struct {
	DYB                   DYBAxes
	DeltaURad             OptionalOrbitScalar
	ArgumentOfLatitudeRad OptionalOrbitScalar
	LightingRatio         float64
}

type EmpiricalRTNTerm struct {
	Axis            RTNAxis
	Harmonic        EmpiricalHarmonic
	CoefficientKmS2 float64
}

type EmpiricalRTNTermSelector struct {
	Axis     RTNAxis
	Harmonic EmpiricalHarmonic
}

type ErrorStats struct {
	Present bool
	RMSM    float64
	MeanM   float64
	MaxAbsM float64
}

type EstimatedForceParameter struct {
	Name         string
	InitialValue float64
	Value        float64
}

type FrameAgreement struct {
	Kind FrameAgreementKind
	From uint32
	To   uint32
}

type GNSSAttitude struct {
	Mode          AttitudeMode
	YawRad        float64
	NominalYawRad float64
	BetaRad       float64
	OrbitAngleRad float64
	BodyX         [3]float64
	BodyY         [3]float64
	BodyZ         [3]float64
	HasDYB        bool
	DYB           DYBAxes
}

type HorizonBinStats struct {
	HorizonStartS float64
	HorizonEndS   float64
	Count         uint64
	Radial        ErrorStats
	Transverse    ErrorStats
	Normal        ErrorStats
	Orbit3d       ErrorStats
	SISREOrbit    ErrorStats
	Clock         ErrorStats
	SISRE         ErrorStats
	ClockCount    uint64
}

type ICGEMHeader struct {
	GravityConstantM3S2 float64
	RadiusM             float64
	MaxDegree           uint16
	Errors              ICGEMErrors
	Normalization       ICGEMNormalization
	HasTideSystem       bool
	TideSystem          GravityTideSystem
	Format              ICGEMFormat
}

type ICGEMSigmas struct {
	HasCalibrated bool
	Calibrated    [2]float64
	HasFormal     bool
	Formal        [2]float64
}

type ICGEMTimeVariableTerm struct {
	Line                uint64
	Kind                ICGEMTermKind
	PeriodYears         float64
	Degree              uint16
	Order               uint16
	C                   float64
	S                   float64
	Sigmas              ICGEMSigmas
	ReferenceTTSeconds  float64
	ValidFromTTSeconds  float64
	ValidUntilTTSeconds float64
}

type IERS2010EopSample struct {
	MJDUTC        int32
	XpArcsec      float64
	YpArcsec      float64
	Ut1MinusUTCS  float64
	LodS          float64
	DxArcsec      float64
	DyArcsec      float64
	HasPoleRates  bool
	PoleRatesRadS [2]float64
}

type IntegratorOptions struct {
	Integrator   uint32
	AbsTol       float64
	RelTol       float64
	InitialStepS float64
	MinStepS     float64
	MaxStepS     float64
	MaxSteps     uint32
}

type MeanConversionOptions struct {
	Tolerance     float64
	MaxIterations uint32
}

type NominalYaw struct {
	YawRad  float64
	BetaRad float64
	BodyX   [3]float64
	BodyY   [3]float64
	BodyZ   [3]float64
}

type OptionalHarmonicCoefficient struct {
	Present     bool
	Coefficient SphericalHarmonicCoefficient
}

type OptionalOrbitScalar struct {
	Present bool
	Value   float64
}

type OrbitParameterCovariance struct {
	Estimated        bool
	Dimension        uint64
	VarianceFactor   float64
	DegreesOfFreedom uint64
}

type OrbitParameterFitSummary struct {
	Satellite           string
	InitialState        CartesianState
	ForceParameterCount uint64
	ParameterCount      uint64
	ResidualStats       OrbitResidualStats
	SeedRMS3dM          float64
	Iterations          uint64
	ConditionNumber     float64
	Ut1Degraded         Ut1Degradation
}

type OrbitRTNResidual struct {
	VariationalEpoch VariationalEpoch
	RadialM          float64
	AlongM           float64
	CrossM           float64
}

type OrbitScoreSummary struct {
	Frame                FrameAgreement
	SourceReferencePoint OrbitReferencePoint
	SatelliteCount       uint64
	SystemCount          uint64
	Coverage             ScoreCoverage
}

type OrbitSourceInfo struct {
	TimeScale               uint32
	HasSP3TimeSystem        bool
	Frame                   OrbitFrame
	ReferencePoint          OrbitReferencePoint
	ClockIncludesRelativity bool
}

type OrbitSourceState struct {
	PositionECEFM [3]float64
	ClockS        OptionalOrbitScalar
}

type RadiationShadowOptions struct {
	ShadowFunction     ShadowFunction
	EarthShadowModel   uint32
	MoonEclipseEnabled bool
}

type RTNComponents struct {
	RadialM     float64
	TransverseM float64
	NormalM     float64
}

type SatelliteScore struct {
	Satellite              string
	PredictionOriginJ2000S OptionalOrbitScalar
	HasOrbitRegime         bool
	OrbitRegime            OrbitRegime
	HasSISREWeights        bool
	SISREWeights           SISREWeights
	BinCount               uint64
	Coverage               ScoreCoverage
}

type ScoreCoverage struct {
	Epochs                    uint64
	Scored                    uint64
	Clocked                   uint64
	OutsideHorizonBins        uint64
	NoPredictionOrigin        uint64
	SourceSatelliteMissing    uint64
	SourceEpochMissing        uint64
	ReferenceSatelliteMissing uint64
	ReferenceEpochMissing     uint64
}

type ShadowGradient struct {
	Fraction      float64
	GradientPerKm [3]float64
}

type SISREWeights struct {
	Radial                  float64
	TransverseNormalSquared float64
}

type SphericalHarmonicCoefficient struct {
	Degree uint16
	Order  uint16
	C      float64
	S      float64
}

type SunMoon struct {
	SunM  [3]float64
	MoonM [3]float64
}

type SystemScore struct {
	System   uint32
	BinCount uint64
	Coverage ScoreCoverage
}

type VariationalEpoch struct {
	TDBJ2000S  float64
	ElapsedTTS float64
}

type VariationalState struct {
	State              CartesianState
	StateTransition    [36]float64
	SensitivityColumns uint64
}

type ZonalHarmonics struct {
	MuKm3S2 float64
	ReKm    float64
	C20     float64
	C30     float64
	C40     float64
	C50     float64
	C60     float64
}
