package sidereon

import (
	"math"
	"slices"
	"strings"
	"testing"
	"time"
)

func exactOrbitValues(t *testing.T, want []float64, actual ...float64) {
	t.Helper()
	if !slices.Equal(actual, want) {
		t.Fatalf("values = %v, want %v", actual, want)
	}
}

func TestOrbitalMeanElementFixture(t *testing.T) {
	pins := readOrbitalFixture(t).Contracts
	gravity := ZonalHarmonics{MuKm3S2: 398600.4418, ReKm: 6378.137, C20: -1.08262668e-3, C30: 2.5326564853322355e-6, C40: 1.619621591367e-6, C50: 2.272960828686982e-7, C60: -5.406812391070848e-7}
	orbitalEqual(t, gravity)(DefaultZonalHarmonics(), nil)
	orbitalEqual(t, gravity)(ZonalHarmonicsEarth(), nil)
	ehOptions := DefaultMeanConversionOptionsEcksteinHechler()
	blOptions := DefaultMeanConversionOptionsBrouwerLyddane()
	orbitalEqual(t, MeanConversionOptions{Tolerance: 1e-11, MaxIterations: 100})(ehOptions, nil)
	orbitalEqual(t, MeanConversionOptions{Tolerance: 1e-13, MaxIterations: 200})(blOptions, nil)
	orbitalEqual(t, "Eckstein-Hechler")(AnalyticTheoryText(AnalyticTheoryEcksteinHechler))
	circular := func(pin string) CircularElements {
		v := pins[pin]
		return CircularElements{AKm: v[0], Ex: v[1], Ey: v[2], IRad: v[3], RaanRad: v[4], MeanLatitudeArgumentRad: v[5]}
	}
	keplerian := func(pin string) AnalyticKeplerianElements {
		v := pins[pin]
		return AnalyticKeplerianElements{AKm: v[0], E: v[1], IRad: v[2], RaanRad: v[3], ArgPerigeeRad: v[4], MeanAnomalyRad: v[5]}
	}
	epoch := orbitalEpoch + 120
	state := AnalyticCartesian{PositionKm: orbitalPosition, VelocityKmS: orbitalVelocity}
	initial := CartesianState{EpochTDBSeconds: orbitalEpoch, PositionKm: orbitalPosition, VelocityKmPerS: orbitalVelocity}
	orbitalEqual(t, circular("CIRCULAR"))(CircularElementsFromCartesian(state, gravity.MuKm3S2))
	ehMean := circular("EH_MEAN")
	blMean := keplerian("BL_MEAN")
	orbitalEqual(t, MeanElements[CircularElements]{Elements: ehMean, Iterations: uint32(pins["EH_MEAN"][6])})(EcksteinHechlerMeanElements(state, gravity, ehOptions))
	orbitalEqual(t, MeanElements[AnalyticKeplerianElements]{Elements: blMean, Iterations: uint32(pins["BL_MEAN"][6])})(BrouwerLyddaneMeanElements(state, gravity, blOptions))
	orbitalEqual(t, circular("EH_OSCULATING"))(EcksteinHechlerOsculatingElements(ehMean, gravity))
	orbitalEqual(t, keplerian("BL_OSCULATING"))(BrouwerLyddaneOsculatingElements(blMean, gravity))
	eh := orbitalResource[*EcksteinHechlerPropagator](t)(NewEcksteinHechlerPropagator(initial, gravity, ehOptions))
	bl := orbitalResource[*BrouwerLyddanePropagator](t)(NewBrouwerLyddanePropagator(initial, gravity, 0, blOptions))
	orbitalEqual(t, ehMean)(eh.MeanElements())
	orbitalEqual(t, blMean)(bl.MeanElements())
	orbitalEqual(t, circular("EH_AT_EPOCH"))(eh.OsculatingElements(epoch))
	orbitalEqual(t, keplerian("BL_AT_EPOCH"))(bl.OsculatingElements(epoch))
	for _, tc := range []struct {
		pin       string
		ephemeris func([]float64) ([]CartesianState, error)
	}{{"EH_EPHEMERIS", eh.Ephemeris}, {"BL_EPHEMERIS", bl.Ephemeris}} {
		states := orbitalResult[[]CartesianState](t)(tc.ephemeris([]float64{epoch, orbitalEpoch - 60}))
		if len(states) != 2 {
			t.Fatalf("%s returned %d states", tc.pin, len(states))
		}
		for i, state := range states {
			exactOrbitState(t, pins[tc.pin][7*i:7*i+7], state)
		}
	}
}

func exactOrbitState(t *testing.T, want []float64, state CartesianState) {
	t.Helper()
	orbitalEqual(t, CartesianState{EpochTDBSeconds: want[0], PositionKm: [3]float64(want[1:4]), VelocityKmPerS: [3]float64(want[4:7])})(state, nil)
}

func TestOrbitalEarthOrientationFixture(t *testing.T) {
	pins := readOrbitalFixture(t).Contracts
	utc := CivilDateTime{Year: 2020, Month: 6, Day: 25, Hour: 12}
	scales := orbitalResult[TimeScales](t)(TimeScalesFromUTC(utc))
	forward := orbitalResult[[3][3]float64](t)(GCRSToMeanOfDateMatrix(scales))
	reverse := orbitalResult[[3][3]float64](t)(MeanOfDateToGCRSMatrix(scales))
	for i := range 3 {
		exactOrbitValues(t, pins["GCRS_TO_MOD"][3*i:3*i+3], forward[i][:]...)
		exactOrbitValues(t, pins["MOD_TO_GCRS"][3*i:3*i+3], reverse[i][:]...)
	}
	for pin, value := range map[string]SunMoon{
		"BODIES_AT": orbitalResult[SunMoon](t)(SunMoonGCRSAt(time.Date(2020, 6, 25, 12, 0, 0, 0, time.UTC))),
		"BODIES":    orbitalResult[SunMoon](t)(SunMoonGCRS(0.2)),
	} {
		exactOrbitValues(t, pins[pin][:3], value.SunM[:]...)
		exactOrbitValues(t, pins[pin][3:], value.MoonM[:]...)
	}
	first := IERS2010EopSample{MJDUTC: 59025, XpArcsec: 0.1, YpArcsec: 0.2, Ut1MinusUTCS: -0.2,
		LodS: 0.001, DxArcsec: 0.0001, DyArcsec: -0.0001, HasPoleRates: true, PoleRatesRadS: [2]float64{8e-14, -3e-14}}
	last := first
	last.MJDUTC++
	provider := orbitalResource[*IERS2010EarthOrientationProvider](t)(NewIERS2010EarthOrientationProvider([]IERS2010EopSample{first, last}))
	same := orbitalResource[*IERS2010EarthOrientationProvider](t)(provider.WithValidity(false))
	permissive := orbitalResource[*IERS2010EarthOrientationProvider](t)(provider.WithValidity(true))
	orbitalEqual(t, true)(provider.Equal(same))
	orbitalEqual(t, false)(provider.Equal(permissive))
	orientation := orbitalResult[EarthOrientation](t)(permissive.OrientationAtTDBSeconds(orbitalEpoch + 600))
	angles := pins["IERS_ANGLES"]
	exactOrbitValues(t, angles[:2], orientation.PolarMotionRad[:]...)
	exactOrbitValues(t, angles[2:5], orientation.EarthRotationVectorITRFRadS[:]...)
	exactOrbitValues(t, angles[5:], orientation.IERS2010Quaternion[:]...)
	orbitalEqual(t, true)(orientation.HasIERS2010Quaternion, nil)
	for i, matrix := range [4][3][3]float64{orientation.GCRFToITRFMatrix, orientation.ITRFToGCRFMatrix, orientation.GCRFToITRFRotationRateMatrix, orientation.ITRFToGCRFRotationRateMatrix} {
		for j := range 3 {
			begin := 9*i + 3*j
			exactOrbitValues(t, pins["IERS_MATRICES"][begin:begin+3], matrix[j][:]...)
		}
	}
}

func TestOrbitalDEContextFixture(t *testing.T) {
	fixture := readOrbitalFixture(t)
	pins := fixture.Contracts
	de := orbitalResource[*DEEphemeris](t)(DEEphemerisFromBytes(readPositioningFixture(t, "orbital/lnxp1990.440")))
	orbitalEqual(t, uint64(len(fixture.DE.Constants)))(de.ConstantCount())
	orbitalEqual(t, DEGM{Present: true, ValueM3S2: pins["DE_EARTH_GM"][0]})(de.GMM3S2(DEBodyEarth))
	config := orbitalResult[PropagationConfig](t)(DefaultPropagationConfig())
	config.EpochTDBSeconds, config.PositionKm, config.VelocityKmPerS = orbitalEpoch+600, orbitalPosition, orbitalVelocity
	config.ForceModel = PropagationForceModelComposite
	earth := EarthRadiationPressureOptions{CR: 1.25, AreaToMassM2PerKg: 0.02}
	solar := SolarRadiationPressure{
		CR: 1.25, AreaToMassM2PerKg: 0.02, HasShadow: true,
		Shadow: RadiationShadowOptions{ShadowFunction: ShadowFunctionConical, EarthShadowModel: EarthShadowSpherical},
	}
	components := ForceModelComponents{HasTwoBody: true, HasSolarRadiationPressure: true, SolarRadiationPressure: solar}
	config.ForceComponents = components.WithEarthRadiationPressure(earth)
	components.HasEarthRadiationPressure, components.EarthRadiationPressure = true, earth
	orbitalEqual(t, components)(config.ForceComponents, nil)
	orbitalEqual(t, solar)(config.ForceComponents.SolarRadiationPressure, nil)
	wantConfig := config
	wantConfig.ForceComponents = components
	orbitalEqual(t, wantConfig)(config, nil)
	orbitalEqual(t, OrbitFitOptions{ForceComponents: components})(OrbitFitOptions{ForceComponents: config.ForceComponents}, nil)
	context := (PropagationContext{}).WithDEEphemeris(de)
	ephemeris := orbitalResource[*Ephemeris](t)(PropagateStateWithForceModel(config, []float64{orbitalEpoch + 660}, PropagationEnvironment{Context: &context}, nil))
	states := orbitalResult[[]CartesianState](t)(ephemeris.States())
	if len(states) != 1 {
		t.Fatalf("state count = %d, want 1", len(states))
	}
	exactOrbitState(t, pins["DE_PROPAGATION"], states[0])
}

func TestOrbitalICGEMFixture(t *testing.T) {
	pins := readOrbitalFixture(t).Contracts
	data := string(readPositioningFixture(t, "orbital/synthetic_icgem2.gfc"))
	field := orbitalResource[*ICGEMGravityField](t)(ParseICGEMGravityFieldTruncated(data, 2, 2))
	orbitalEqual(t, "SYNTHETIC_ICGEM2")(field.ModelName())
	orbitalEqual(t, uint16(2))(field.StoredMaxDegree())
	orbitalEqual(t, uint16(2))(field.StoredMaxOrder())
	orbitalEqual(t, ICGEMHeader{GravityConstantM3S2: 3.986004415e14, RadiusM: 6378136.3, MaxDegree: 6,
		Errors: ICGEMErrorsCalibratedAndFormal, Normalization: ICGEMNormalizationFullyNormalized,
		HasTideSystem: true, TideSystem: GravityTideSystemZeroTide, Format: ICGEMFormatICGEM2})(field.Header())
	full := orbitalResource[*ICGEMGravityField](t)(ParseICGEMGravityField(data))
	orbitalEqual(t, false)(field.Equal(full))
	orbitalEqual(t, OptionalHarmonicCoefficient{Present: true, Coefficient: SphericalHarmonicCoefficient{C: 1}})(field.StaticCoefficient(0, 0))
	orbitalEqual(t, ICGEMSigmas{HasCalibrated: true, HasFormal: true})(field.StaticSigmas(0, 0))
	terms := orbitalResult[[]ICGEMTimeVariableTerm](t)(field.TimeVariableTerms())
	expected := ICGEMTimeVariableTerm{Line: 16, Kind: ICGEMTermKindConstant, Degree: 1, C: 7e-11,
		Sigmas:             ICGEMSigmas{HasCalibrated: true, Calibrated: [2]float64{1e-13, 0}, HasFormal: true, Formal: [2]float64{2e-13, 0}},
		ReferenceTTSeconds: -43200, ValidFromTTSeconds: -43200, ValidUntilTTSeconds: 946728000}
	cosine := expected
	cosine.Line, cosine.Kind, cosine.C, cosine.PeriodYears = 17, ICGEMTermKindCosine, 7.8e-11, 1
	sine := expected
	sine.Line, sine.Kind, sine.Degree, sine.C, sine.PeriodYears = 26, ICGEMTermKindSine, 2, 5.92e-11, 1
	trend := expected
	trend.Line, trend.Kind, trend.Degree, trend.C, trend.ValidUntilTTSeconds = 20, ICGEMTermKindTrend, 2, 1.16e-11, 172110600
	for _, want := range []ICGEMTimeVariableTerm{expected, cosine, sine, trend} {
		if !slices.Contains(terms, want) {
			t.Fatalf("time-variable term missing: %+v", want)
		}
	}
	gravity := orbitalResource[*SphericalHarmonicGravity](t)(field.Gravity(2, 2))
	orbitalEqual(t, [3]float64(pins["ICGEM_ACCELERATION"]))(gravity.BodyFixedAccelerationAtKmS2(orbitalPosition, 86400))
}

func TestOrbitalExplicitRadiationFixture(t *testing.T) {
	pins := readOrbitalFixture(t).Contracts
	coefficients := [9]float64{1e-8, 2e-8, 3e-8, 4e-8, 5e-8, 6e-8, 7e-8, 8e-8, 9e-8}
	options := DefaultRadiationShadowOptions()
	ecom1Coefficients := orbitalResult[ECOM1Coefficients](t)(ECOM1CoefficientsFromArray(coefficients))
	ecom2Coefficients := orbitalResult[ECOM2D4B1Coefficients](t)(ECOM2D4B1CoefficientsFromArray(coefficients))
	ecom1 := orbitalResource[*ECOM1SolarRadiationPressure](t)(NewECOM1SolarRadiationPressure(ecom1Coefficients, options))
	ecom2 := orbitalResource[*ECOM2SolarRadiationPressure](t)(ECOM2SolarRadiationPressureD4B1(ecom2Coefficients, options))
	ecom1Sun := orbitalResource[*ECOM1SolarRadiationPressure](t)(ecom1.WithMoonEclipse(false))
	ecom2Sun := orbitalResource[*ECOM2SolarRadiationPressure](t)(ecom2.WithMoonEclipse(false))
	comparePartials := func(statePin, parameterPin string, state AccelerationPartials, parameters [][3]float64) {
		t.Helper()
		partials := pins[statePin]
		for i := range 3 {
			exactOrbitValues(t, partials[3*i:3*i+3], state.WrtPosition[3*i:3*i+3]...)
			exactOrbitValues(t, partials[9+3*i:12+3*i], state.WrtVelocity[3*i:3*i+3]...)
		}
		parameterValues := pins[parameterPin]
		if 3*len(parameters) != len(parameterValues) {
			t.Fatalf("partials count = %d, want %d", len(parameters), len(parameterValues)/3)
		}
		for i, vector := range parameters {
			exactOrbitValues(t, parameterValues[3*i:3*i+3], vector[:]...)
		}
	}

	p, v, s, m := orbitalPosition, orbitalVelocity, orbitalSun, orbitalMoon
	for _, test := range []struct {
		prefix    string
		sun       func([3]float64, [3]float64, [3]float64) ([3]float64, error)
		sunMoon   func([3]float64, [3]float64, [3]float64, [3]float64) ([3]float64, error)
		stateSun  func([3]float64, [3]float64, [3]float64) (AccelerationPartials, error)
		stateMoon func([3]float64, [3]float64, [3]float64, [3]float64) (AccelerationPartials, error)
		paramSun  func([3]float64, [3]float64, [3]float64) ([][3]float64, error)
		paramMoon func([3]float64, [3]float64, [3]float64, [3]float64) ([][3]float64, error)
	}{
		{"ECOM1", ecom1Sun.AccelerationWithSun, ecom1.AccelerationWithSunMoon, ecom1Sun.AccelerationPartialsWithSun, ecom1.AccelerationPartialsWithSunMoon, ecom1Sun.ParameterPartialsWithSun, ecom1.ParameterPartialsWithSunMoon},
		{"ECOM2", ecom2Sun.AccelerationWithSun, ecom2.AccelerationWithSunMoon, ecom2Sun.AccelerationPartialsWithSun, ecom2.AccelerationPartialsWithSunMoon, ecom2Sun.ParameterPartialsWithSun, ecom2.ParameterPartialsWithSunMoon},
	} {
		t.Run(test.prefix, func(t *testing.T) {
			orbitalEqual(t, [3]float64(pins[test.prefix+"_SUN"]))(test.sun(p, v, s))
			orbitalEqual(t, [3]float64(pins[test.prefix+"_SUN_MOON"]))(test.sunMoon(p, v, s, m))
			comparePartials(test.prefix+"_STATE_SUN", test.prefix+"_PARAMETER_SUN", orbitalResult[AccelerationPartials](t)(test.stateSun(p, v, s)), orbitalResult[[][3]float64](t)(test.paramSun(p, v, s)))
			comparePartials(test.prefix+"_STATE_SUN_MOON", test.prefix+"_PARAMETER_SUN_MOON", orbitalResult[AccelerationPartials](t)(test.stateMoon(p, v, s, m)), orbitalResult[[][3]float64](t)(test.paramMoon(p, v, s, m)))
		})
	}
	geometry := pins["ECOM_GEOMETRY"]
	expected := ECOMGeometry{DYB: DYBAxes{D: [3]float64(geometry[:3]), Y: [3]float64(geometry[3:6]), B: [3]float64(geometry[6:9])},
		DeltaURad: OptionalOrbitScalar{Present: true, Value: geometry[9]}, ArgumentOfLatitudeRad: OptionalOrbitScalar{Present: true, Value: geometry[10]}, LightingRatio: geometry[11]}
	orbitalEqual(t, expected)(ECOMGeometryWithShadow(p, v, s, &m, EarthShadowSpherical))
	orbitalEqual(t, expected)(ComputeECOMGeometry(p, v, s))
	solar := orbitalResource[*SolarRadiationPressureForce](t)(NewSolarRadiationPressureForce(1.25, 0.02, options))
	earth := orbitalResource[*EarthRadiationPressure](t)(NewEarthRadiationPressure(1.25, 0.02, false))
	orbitalEqual(t, [3]float64(pins["SOLAR"]))(solar.AccelerationWithSunMoon(p, s, m))
	orbitalEqual(t, [3]float64(pins["EARTH"]))(earth.AccelerationWithSunMoon(CartesianState{EpochTDBSeconds: orbitalEpoch, PositionKm: p, VelocityKmPerS: v}, s, m))
	orbitalEqual(t, "srp.cr")(SolarRadiationPressureCRParameter())
}

func TestOrbitalScoringConventions(t *testing.T) {
	orbitalEqual(t, 0.5)(DefaultVelocityHalfStepS(), nil)
	orbitalEqual(t, 1e-8)(MinVelocityHalfStepS(), nil)
	position := [3]float64{26560000, 0, 0}
	inertial := [3]float64{0, 2000 + 7.2921151467e-5*26560000, 0}
	orbitalEqual(t, inertial)(InertialVelocityFromEarthFixed(position, [3]float64{0, 2000, 0}))
	regime := orbitalResult[*OrbitRegime](t)(OrbitRegimeFromInertialState(position, inertial))
	if regime == nil || *regime != OrbitRegimeMEO {
		t.Fatalf("orbit regime = %v, want MEO", regime)
	}
	weights := orbitalResult[*SISREWeights](t)(SISREWeightsMontenbruck2018(GNSSSystemGPS, OrbitRegimeMEO))
	expected := SISREWeights{Radial: 0.98, TransverseNormalSquared: 1.0 / 49.0}
	if weights == nil {
		t.Fatal("missing GPS MEO weights")
	}
	orbitalEqual(t, expected)(*weights, nil)
	auto := SISREWeightSelection{}
	fixed := SISREWeightSelection{Fixed: weights}
	orbitalEqual(t, true)(auto.Equal(SISREWeightSelection{}))
	orbitalEqual(t, true)(fixed.Equal(SISREWeightSelection{Fixed: &expected}))
	orbitalEqual(t, false)(auto.Equal(fixed))
	orbitalEqual(t, true)((ScoreEpochs{}).Equal(ScoreEpochs{}))
	orbitalEqual(t, false)((ScoreEpochs{}).Equal(ScoreEpochs{Instants: []ClockEpoch{}}))
	orbitalEqual(t, true)((ScoreEpochs{Instants: []ClockEpoch{orbitalClockEpoch(GPST, 60)}}).Equal(
		ScoreEpochs{Instants: []ClockEpoch{{Scale: GPST, Representation: RINEXClockInstantNanos, NanosLow: 60000000000}}}))
	orbitalEqual(t, "MEO")(OrbitRegimeMEO.String(), nil)
	orbitalEqual(t, "orbit source")(ScoredProductSource.String(), nil)
}

func TestOrbitalRadiationOptions(t *testing.T) {
	options := DefaultRadiationShadowOptions()
	ecom1 := orbitalResource[*ECOM1SolarRadiationPressure](t)(NewECOM1SolarRadiationPressure(ECOM1Coefficients{}, options))
	orbitalEqual(t, ECOM1Coefficients{})(ecom1.Coefficients())
	orbitalEqual(t, true)(ecom1.MoonEclipseEnabled())
	orbitalEqual(t, ShadowFunctionApparentDisc)(ecom1.ShadowFunction())
	orbitalEqual(t, EarthShadowWGS84Oblate)(ecom1.EarthShadowModel())
	spherical := orbitalResource[*ECOM1SolarRadiationPressure](t)(ecom1.WithEarthShadowModel(EarthShadowSpherical))
	orbitalEqual(t, EarthShadowSpherical)(spherical.EarthShadowModel())
	conical := orbitalResource[*ECOM1SolarRadiationPressure](t)(ecom1.WithShadowFunction(ShadowFunctionConical))
	orbitalEqual(t, ShadowFunctionConical)(conical.ShadowFunction())
	orbitalEqual(t, false)(conical.MoonEclipseEnabled())
	solar := orbitalResource[*SolarRadiationPressureForce](t)(NewSolarRadiationPressureForce(1.25, 0.02, options))
	sphericalSolar := orbitalResource[*SolarRadiationPressureForce](t)(solar.WithEarthShadowModel(EarthShadowSpherical))
	orbitalEqual(t, EarthShadowSpherical)(sphericalSolar.EarthShadowModel())
	conicalSolar := orbitalResource[*SolarRadiationPressureForce](t)(solar.WithShadowFunction(ShadowFunctionConical))
	orbitalEqual(t, ShadowFunctionConical)(conicalSolar.ShadowFunction())
	withoutMoon := orbitalResource[*SolarRadiationPressureForce](t)(solar.WithMoonEclipse(false))
	orbitalEqual(t, false)(withoutMoon.MoonEclipseEnabled())
}

func TestOrbitalGNSSBlockNames(t *testing.T) {
	names := []struct {
		block GNSSBlock
		igs   string
	}{
		{GNSSBlockGPSIIA, "BLOCK IIA"}, {GNSSBlockGPSIIR, "BLOCK IIR-B"}, {GNSSBlockGPSIIRM, "BLOCK IIR-M"},
		{GNSSBlockGPSIIF, "BLOCK IIF"}, {GNSSBlockGPSIII, "BLOCK IIIA"}, {GNSSBlockGLONASSM, "GLONASS-M"},
		{GNSSBlockGalileoIOV, "GALILEO-1"}, {GNSSBlockGalileoFOC, "GALILEO-2"}, {GNSSBlockBeiDou2GEO, "BEIDOU-2G"},
		{GNSSBlockBeiDou2IGSO, "BEIDOU-2I"}, {GNSSBlockBeiDou2MEO, "BEIDOU-2M"}, {GNSSBlockBeiDou3GEO, "BEIDOU-3G-CAST"},
		{GNSSBlockBeiDou3IGSO, "BEIDOU-3I"}, {GNSSBlockBeiDou3MEOUnknown, "BEIDOU-3M"},
		{GNSSBlockBeiDou3MEOCAST, "BEIDOU-3M-CAST"}, {GNSSBlockBeiDou3MEOSECM, "BEIDOU-3M-SECM"},
	}
	for i, entry := range names {
		orbitalEqual(t, uint32(i))(uint32(entry.block), nil)
		orbitalEqual(t, entry.igs)(entry.block.IGSName())
		orbitalEqual(t, entry.igs)(entry.block.String(), nil)
		orbitalEqual(t, entry.block)(GNSSBlockFromIGSName(entry.igs))
	}
	meo := BeiDouOrbitClassMEO
	orbitalEqual(t, GNSSBlockBeiDou3MEOUnknown)(GNSSBlockFromConstellationBlockType(GNSSSystemBeiDou, "BDS-3", &meo))
	orbitalEqual(t, GNSSBlockBeiDou3MEOUnknown)(GNSSBlockBeiDou3MEO(nil))
}

func TestOrbitalEnumValues(t *testing.T) {
	for _, entry := range []struct {
		value uint32
		code  uint32
	}{
		{uint32(AttitudeModeMidnightTurn), 2},
		{uint32(AttitudeModeModifiedYawSteering), 5},
		{uint32(AttitudeModeNominal), 0},
		{uint32(AttitudeModeNoonTurn), 1},
		{uint32(AttitudeModeOrbitNormal), 4},
		{uint32(AttitudeModeShadow), 3},
		{uint32(CosineTurnStartAngleColinearity), 0},
		{uint32(CosineTurnStartAngleOrbitAngle), 1},
		{uint32(BeiDouManufacturerCAST), 0},
		{uint32(BeiDouManufacturerSECM), 1},
		{uint32(BeiDouOrbitClassGEO), 0},
		{uint32(BeiDouOrbitClassIGSO), 1},
		{uint32(BeiDouOrbitClassMEO), 2},
		{uint32(DEBodyEarth), 4},
		{uint32(DEBodyEarthMoon), 3},
		{uint32(DEBodyJupiter), 7},
		{uint32(DEBodyMars), 6},
		{uint32(DEBodyMercury), 1},
		{uint32(DEBodyMoon), 5},
		{uint32(DEBodyNeptune), 10},
		{uint32(DEBodyPluto), 11},
		{uint32(DEBodySaturn), 8},
		{uint32(DEBodySolarSystem), 12},
		{uint32(DEBodySun), 0},
		{uint32(DEBodyUranus), 9},
		{uint32(DEBodyVenus), 2},
		{uint32(ShadowFunctionApparentDisc), 0},
		{uint32(ShadowFunctionConical), 1},
		{uint32(EmpiricalHarmonicConstant), 0},
		{uint32(EmpiricalHarmonicCosine), 1},
		{uint32(EmpiricalHarmonicSine), 2},
		{uint32(RTNAxisAlongTrack), 1},
		{uint32(RTNAxisCrossTrack), 2},
		{uint32(RTNAxisRadial), 0},
		{uint32(ICGEMErrorsCalibrated), 1},
		{uint32(ICGEMErrorsCalibratedAndFormal), 3},
		{uint32(ICGEMErrorsFormal), 2},
		{uint32(ICGEMErrorsNo), 0},
		{uint32(ICGEMFormatICGEM1), 0},
		{uint32(ICGEMFormatICGEM2), 1},
		{uint32(ICGEMNormalizationFullyNormalized), 0},
		{uint32(ICGEMNormalizationUnnormalized), 1},
		{uint32(ICGEMTermKindConstant), 0},
		{uint32(ICGEMTermKindTrend), 1},
		{uint32(AnalyticTheoryBrouwerLyddane), 1},
		{uint32(AnalyticTheoryEcksteinHechler), 0},
		{uint32(ClockDatumEpochSystemMedian), 0},
		{uint32(ClockDatumUnadjusted), 1},
		{uint32(FrameAgreementKindAssertedEquivalent), 1},
		{uint32(FrameAgreementKindBroadcastDatums), 3},
		{uint32(FrameAgreementKindSameLabel), 0},
		{uint32(OrbitScoreGapPolicyCount), 0},
		{uint32(OrbitScoreGapPolicyRefuse), 1},
		{uint32(OrbitReferencePointAntennaPhaseCenter), 1},
		{uint32(OrbitReferencePointCenterOfMass), 0},
		{uint32(OrbitRegimeGEO), 2},
		{uint32(OrbitRegimeIGSO), 1},
		{uint32(OrbitRegimeMEO), 0},
		{uint32(ScoredProductReference), 1},
		{uint32(ScoredProductSource), 0},
		{uint32(TriadVelocityEarthFixed), 1},
		{uint32(TriadVelocityInertial), 0},
	} {
		orbitalEqual(t, entry.code)(entry.value, nil)
	}
}

func TestOrbitalScoreFrameFixtures(t *testing.T) {
	data := string(readPositioningFixture(t, "trimmed.sp3"))
	reference := orbitalResource[*SP3](t)(LoadSP3([]byte(data)))
	newFrame := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.Replace(data, "IGb14", "IGS20", 1))))
	source := orbitalResource[*PredictedSP3](t)(PredictedSP3WithCutoff(newFrame, orbitalClockEpoch(GPST, 646266600)))
	options := DefaultOrbitScoreOptions([]float64{0, 1800, 7200})
	// 2020-06-24 10:45:00 GPST is 900 seconds into the prediction horizon.
	options.Epochs = ScoreEpochs{Instants: []ClockEpoch{orbitalClockEpoch(GPST, 646267500)}}
	options.SISREWeights.Fixed = &SISREWeights{Radial: 1, TransverseNormalSquared: 0.02}
	options.FrameReconciliation = SP3FrameReconciliationOptions{AssertedEquivalentLabelSets: [][]string{{"IGS20", "IGb14"}}}
	result := orbitalResource[*OrbitScoreReport](t)(ScoreOrbit(source, reference, []string{"G08"}, options))
	summary := orbitalResult[OrbitScoreSummary](t)(result.Summary())
	orbitalEqual(t, FrameAgreement{Kind: FrameAgreementKindAssertedEquivalent})(summary.Frame, nil)
	orbitalEqual(t, ScoreCoverage{Epochs: 1, Scored: 1})(summary.Coverage, nil)
	options.FrameReconciliation = SP3FrameReconciliationOptions{Helmert: true}
	result = orbitalResource[*OrbitScoreReport](t)(ScoreOrbit(source, reference, []string{"G08"}, options))
	summary = orbitalResult[OrbitScoreSummary](t)(result.Summary())
	orbitalEqual(t, FrameAgreement{Kind: FrameAgreementKindHelmert, From: TerrestrialFrameITRF2020, To: TerrestrialFrameITRF2014})(summary.Frame, nil)
	orbitalEqual(t, ScoreCoverage{Epochs: 1, Scored: 1})(summary.Coverage, nil)
}

func TestOrbitalPredictionFlagOrigin(t *testing.T) {
	rows := strings.Split(string(readPositioningFixture(t, "trimmed.sp3")), "\n")
	for i := len(rows) - 1; i >= 0; i-- {
		if strings.HasPrefix(rows[i], "PG08") {
			rows[i] = rows[i][:60] + strings.Repeat(" ", 19) + "P"
			break
		}
	}
	product := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.Join(rows, "\n"))))
	predicted := orbitalResource[*PredictedSP3](t)(PredictedSP3FromPredictionFlags(product))
	// The last unpredicted G08 record is at 2020-06-24 13:15:00 GPST.
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: 646276500})(predicted.PredictionOriginJ2000S("G08"))
	orbitalEqual(t, OrbitFrame{Kind: OrbitFrameKindLabel, FrameLabel: "IGb14"})(predicted.Frame())
}

func TestOrbitalScoringValueEquality(t *testing.T) {
	leftFrame := SP3FrameReconciliationOptions{AssertedEquivalentLabelSets: [][]string{{"IGS14", " ITRF14 "}}}
	rightFrame := SP3FrameReconciliationOptions{AssertedEquivalentLabelSets: [][]string{{"ITRF14", "IGS14", "ITRF14"}}}
	orbitalEqual(t, true)(leftFrame.Equal(rightFrame))
	rightFrame.Helmert = true
	orbitalEqual(t, false)(leftFrame.Equal(rightFrame))

	left := DefaultOrbitScoreOptions([]float64{0, 60})
	left.FrameReconciliation = leftFrame
	left.Epochs = ScoreEpochs{Instants: []ClockEpoch{orbitalClockEpoch(GPST, 60)}}
	right := left
	right.HorizonBinEdgesS = []float64{0, 60}
	right.Epochs = ScoreEpochs{Instants: []ClockEpoch{orbitalClockEpoch(GPST, 60)}}
	right.FrameReconciliation = rightFrame
	right.FrameReconciliation.Helmert = false
	left.SISREWeights.Fixed = &SISREWeights{Radial: 1, TransverseNormalSquared: 0.02}
	right.SISREWeights.Fixed = &SISREWeights{Radial: 1, TransverseNormalSquared: 0.02}
	orbitalEqual(t, true)(left.Equal(right))
	right.Epochs.Instants[0] = orbitalClockEpoch(GPST, 61)
	orbitalEqual(t, false)(left.Equal(right))
	orbitalEqual(t, false)(left.Epochs.Equal(right.Epochs))

	sample := TrajectorySample{Satellite: "G03", Epoch: orbitalClockEpoch(GPST, 60), PositionECEFM: [3]float64{1, 2, 3}}
	other := sample
	other.ClockS.Value = 1
	orbitalEqual(t, true)(sample.Equal(other))
	other.ClockS.Present = true
	orbitalEqual(t, false)(sample.Equal(other))
	sample.ClockS = OptionalOrbitScalar{Present: true, Value: 1}
	orbitalEqual(t, true)(sample.Equal(other))
}

func TestOrbitalJ2ForceFixture(t *testing.T) {
	pins := readOrbitalFixture(t).Contracts
	field := ZonalHarmonicsEarth()
	gravity := orbitalResource[*J2Gravity](t)(NewJ2Gravity(field.MuKm3S2, field.ReKm, -field.C20))
	unusedVelocity := [3]float64{math.NaN(), math.NaN(), math.NaN()}
	state := CartesianState{EpochTDBSeconds: math.NaN(), PositionKm: orbitalPosition, VelocityKmPerS: unusedVelocity}
	want := [3]float64(pins["J2_ACCELERATION"])
	orbitalEqual(t, want)(gravity.Acceleration(state, nil))
	orbitalEqual(t, want)(gravity.AccelerationWithParameters(state, nil, nil))
	orbitalEqual(t, want)(ForceJ2Acceleration(state.PositionKm, unusedVelocity))
	distant := CartesianState{PositionKm: [3]float64{1e80, 2e80, 3e80}}
	orbitalEqual(t, [3]float64(pins["J2_DISTANT_ACCELERATION"]))(gravity.Acceleration(distant, nil))
	orbitalEqual(t, AccelerationPartials{
		WrtPosition: [9]float64(pins["J2_PARTIALS"][:9]), WrtVelocity: [9]float64(pins["J2_PARTIALS"][9:]),
	})(gravity.AccelerationPartials(state, nil, nil))
	if params := orbitalResult[[]ForceParameter](t)(gravity.Parameters()); len(params) != 0 {
		t.Fatalf("J2 parameters = %v", params)
	}
}
