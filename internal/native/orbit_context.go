package native

type PropagationContext struct {
	DEEphemeris      *DEEphemeris
	EarthOrientation *IERS2010EarthOrientationProvider
}

type ForceParameter struct {
	Name  string
	Value float64
}

type VariationalEquations struct {
	Force                 *ForceModel
	Parameters            []float64
	SensitivityParameters []uint64
}

type TrajectorySample struct {
	Satellite     string
	Epoch         NativeClockEpoch
	PositionECEFM [3]float64
	ClockS        OptionalOrbitScalar
}

type OrbitScoreOptions struct {
	Epochs            []NativeClockEpoch
	HorizonBinEdgesS  []float64
	Triad             TriadVelocity
	VelocityHalfStepS float64
	SISREWeights      *SISREWeights
	ClockDatum        ClockDatum
	Gaps              OrbitScoreGapPolicy
	FrameLabelSets    [][]string
	Helmert           bool
}

type EarthRadiationPressureOptions struct {
	CR                 float64
	AreaToMassM2PerKg  float64
	MoonEclipseEnabled bool
}

type InertialPositionSample struct {
	Epoch         NativeClockEpoch
	PositionGCRFM [3]float64
}

type OrbitParameterFitOptions struct {
	GravityTideSystem      uint32
	ForceModel             uint32
	ForceComponents        ForceModelComponents
	SolarShadow            *RadiationShadowOptions
	EarthRadiationPressure *EarthRadiationPressureOptions
	MuKm3S2                *float64
	Drag                   *DragParameters
	SpaceWeather           *SpaceWeatherTable
	AdditionalForces       []*ForceModel
	EstimatedParameters    []string
	IntegratorOptions      IntegratorOptions
	Context                *PropagationContext
	ObservationSigmaM      float64
	MaxIterations          uint64
	ConvergenceThreshold   float64
	MaxConditionNumber     float64
	MinLedgerSamples       uint64
}

type PropagationEnvironment struct {
	GravityTideSystem      uint32
	SolarShadow            *RadiationShadowOptions
	EarthRadiationPressure *EarthRadiationPressureOptions
	Context                *PropagationContext
	SpaceWeather           *SpaceWeatherTable
}

type StateWithCovariance struct {
	State      CartesianState
	Covariance [6][6]float64
}

type TCAForceOptions struct {
	ForceComponents ForceModelComponents
	Environment     PropagationEnvironment
}
