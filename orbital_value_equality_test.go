package sidereon

import (
	"math"
	"testing"
)

func assertOrbitalOptionalEquality[T interface{ Equal(T) bool }](t *testing.T, absent, unused, present, different T) {
	t.Helper()
	if !absent.Equal(unused) || !unused.Equal(absent) || !present.Equal(present) ||
		absent.Equal(present) || present.Equal(absent) || present.Equal(different) || different.Equal(present) {
		t.Fatalf("unexpected %T equality for absent %+v, unused %+v, present %+v, different %+v", absent, absent, unused, present, different)
	}
}

func TestOrbitalOptionalValueEquality(t *testing.T) {
	absentClock := OptionalOrbitScalar{Value: math.NaN()}
	presentClock := OptionalOrbitScalar{Present: true, Value: 1}
	otherClock := OptionalOrbitScalar{Present: true, Value: 2}
	assertOrbitalOptionalEquality(t, OptionalOrbitScalar{}, absentClock, presentClock, otherClock)
	presentNaN := OptionalOrbitScalar{Present: true, Value: math.NaN()}
	if presentNaN.Equal(presentNaN) || !(OptionalOrbitScalar{Present: true, Value: math.Copysign(0, -1)}).Equal(OptionalOrbitScalar{Present: true}) {
		t.Fatal("present scalars must use floating-point equality")
	}
	assertOrbitalOptionalEquality(t, OrbitSourceState{}, OrbitSourceState{ClockS: absentClock},
		OrbitSourceState{ClockS: presentClock}, OrbitSourceState{ClockS: otherClock})
	if (OrbitSourceState{}).Equal(OrbitSourceState{PositionECEFM: [3]float64{1}}) {
		t.Fatal("states with different positions must differ")
	}
	assertOrbitalOptionalEquality(t, DEGM{}, DEGM{ValueM3S2: 1}, DEGM{Present: true}, DEGM{Present: true, ValueM3S2: 1})
	coefficient := SphericalHarmonicCoefficient{Degree: 2, C: 1}
	assertOrbitalOptionalEquality(t, OptionalHarmonicCoefficient{}, OptionalHarmonicCoefficient{Coefficient: coefficient},
		OptionalHarmonicCoefficient{Present: true}, OptionalHarmonicCoefficient{Present: true, Coefficient: coefficient})
	assertOrbitalOptionalEquality(t, ECOMGeometry{}, ECOMGeometry{DeltaURad: absentClock, ArgumentOfLatitudeRad: absentClock},
		ECOMGeometry{DeltaURad: presentClock, ArgumentOfLatitudeRad: presentClock},
		ECOMGeometry{DeltaURad: presentClock, ArgumentOfLatitudeRad: otherClock})
	assertOrbitalOptionalEquality(t, OrbitParameterCovariance{}, OrbitParameterCovariance{Dimension: 2, VarianceFactor: 3, DegreesOfFreedom: 4},
		OrbitParameterCovariance{Estimated: true, Dimension: 2, VarianceFactor: 3, DegreesOfFreedom: 4},
		OrbitParameterCovariance{Estimated: true, Dimension: 2, VarianceFactor: 3, DegreesOfFreedom: 5})
}

func TestOrbitalAttitudeAndGravityValueEquality(t *testing.T) {
	axes := DYBAxes{D: [3]float64{1}, Y: [3]float64{0, 1}, B: [3]float64{0, 0, 1}}
	assertOrbitalOptionalEquality(t, GNSSAttitude{}, GNSSAttitude{DYB: axes},
		GNSSAttitude{HasDYB: true}, GNSSAttitude{HasDYB: true, DYB: axes})
	quaternion := [4]float64{1}
	assertOrbitalOptionalEquality(t, EarthOrientation{}, EarthOrientation{IERS2010Quaternion: quaternion},
		EarthOrientation{HasIERS2010Quaternion: true}, EarthOrientation{HasIERS2010Quaternion: true, IERS2010Quaternion: quaternion})
	sample := IERS2010EopSample{MJDUTC: 59025}
	assertOrbitalOptionalEquality(t, sample, IERS2010EopSample{MJDUTC: 59025, PoleRatesRadS: [2]float64{1, 2}},
		IERS2010EopSample{MJDUTC: 59025, HasPoleRates: true},
		IERS2010EopSample{MJDUTC: 59025, HasPoleRates: true, PoleRatesRadS: [2]float64{1, 2}})

	assertOrbitalOptionalEquality(t, ICGEMHeader{}, ICGEMHeader{TideSystem: GravityTideSystemMeanTide},
		ICGEMHeader{HasTideSystem: true, TideSystem: GravityTideSystemTideFree},
		ICGEMHeader{HasTideSystem: true, TideSystem: GravityTideSystemMeanTide})
	uncertainty := [2]float64{1, 2}
	absentSigmas := ICGEMSigmas{Calibrated: uncertainty, Formal: uncertainty}
	calibrated := ICGEMSigmas{HasCalibrated: true, Calibrated: uncertainty}
	formal := ICGEMSigmas{HasFormal: true, Formal: uncertainty}
	assertOrbitalOptionalEquality(t, ICGEMSigmas{}, absentSigmas, calibrated, ICGEMSigmas{HasCalibrated: true})
	assertOrbitalOptionalEquality(t, ICGEMSigmas{}, absentSigmas, formal, ICGEMSigmas{HasFormal: true})
	if !(ICGEMTimeVariableTerm{Kind: ICGEMTermKindTrend, PeriodYears: 1, Sigmas: absentSigmas}).Equal(ICGEMTimeVariableTerm{Kind: ICGEMTermKindTrend}) {
		t.Fatal("trend comparison must ignore the unused period and uncertainties")
	}
	term := ICGEMTimeVariableTerm{Kind: ICGEMTermKindSine, PeriodYears: 1, Sigmas: formal}
	different := term
	different.PeriodYears = 2
	if !term.Equal(term) || term.Equal(different) || different.Equal(term) {
		t.Fatal("sine term comparison must include the period")
	}
	different = term
	different.Sigmas = calibrated
	if term.Equal(different) {
		t.Fatal("term comparison must include the uncertainties")
	}
}

func TestOrbitalScoreSummaryValueEquality(t *testing.T) {
	absent := ErrorStats{RMSM: 1, MeanM: 2, MaxAbsM: 3}
	present := ErrorStats{Present: true, RMSM: 1, MeanM: 2, MaxAbsM: 3}
	different := ErrorStats{Present: true, RMSM: 1, MeanM: 2, MaxAbsM: 4}
	assertOrbitalOptionalEquality(t, ErrorStats{}, absent, present, different)
	assertOrbitalOptionalEquality(t, HorizonBinStats{},
		HorizonBinStats{Radial: absent, Transverse: absent, Normal: absent, Orbit3d: absent, SISREOrbit: absent, Clock: absent, SISRE: absent},
		HorizonBinStats{Count: 1, Radial: present}, HorizonBinStats{Count: 1, Radial: different})

	weights := SISREWeights{Radial: 1, TransverseNormalSquared: 2}
	absentSatellite := SatelliteScore{Satellite: "G03"}
	withRegime := SatelliteScore{Satellite: "G03", HasOrbitRegime: true, OrbitRegime: 1}
	otherRegime := withRegime
	otherRegime.OrbitRegime = 2
	assertOrbitalOptionalEquality(t, absentSatellite, SatelliteScore{Satellite: "G03", PredictionOriginJ2000S: OptionalOrbitScalar{Value: 1}, OrbitRegime: 1, SISREWeights: weights},
		withRegime, otherRegime)
	if absentSatellite.Equal(SatelliteScore{Satellite: "G03", HasSISREWeights: true, SISREWeights: weights}) ||
		absentSatellite.Equal(SatelliteScore{Satellite: "G03", PredictionOriginJ2000S: OptionalOrbitScalar{Present: true}}) {
		t.Fatal("satellite comparison must include the optional weights and prediction origin")
	}

	broadcast := OrbitFrame{Kind: OrbitFrameKindBroadcastDatums}
	assertOrbitalOptionalEquality(t, broadcast, OrbitFrame{Kind: OrbitFrameKindBroadcastDatums, FrameLabel: "IGS20"},
		OrbitFrame{Kind: OrbitFrameKindLabel, FrameLabel: "IGS20"}, OrbitFrame{Kind: OrbitFrameKindLabel, FrameLabel: "IGS14"})
	unused := FrameAgreement{Kind: FrameAgreementKindSameLabel, From: 1, To: 2}
	helmert := FrameAgreement{Kind: FrameAgreementKindHelmert, From: 1, To: 2}
	otherHelmert := FrameAgreement{Kind: FrameAgreementKindHelmert, From: 1, To: 3}
	assertOrbitalOptionalEquality(t, FrameAgreement{}, unused, helmert, otherHelmert)
	assertOrbitalOptionalEquality(t, OrbitScoreSummary{}, OrbitScoreSummary{Frame: unused},
		OrbitScoreSummary{Frame: helmert}, OrbitScoreSummary{Frame: otherHelmert})
}

func TestOrbitalForceConfigurationValueEquality(t *testing.T) {
	shadow := RadiationShadowOptions{MoonEclipseEnabled: true}
	assertOrbitalOptionalEquality(t, SolarRadiationPressure{}, SolarRadiationPressure{Shadow: shadow},
		SolarRadiationPressure{HasShadow: true}, SolarRadiationPressure{HasShadow: true, Shadow: shadow})

	unused := ForceModelComponents{
		TwoBodyMuKm3PerS2Enabled: true, TwoBodyMuKm3PerS2: 1, ZonalMaxDegree: 2,
		SphericalHarmonicMaxDegree: 3, SphericalHarmonicMaxOrder: 2, ThirdBodySun: true, ThirdBodyMoon: true,
		SolarRadiationPressure: SolarRadiationPressure{HasShadow: true, Shadow: shadow},
		EarthRadiationPressure: EarthRadiationPressureOptions{CR: 1},
	}
	assertOrbitalOptionalEquality(t, ForceModelComponents{}, unused,
		ForceModelComponents{HasEarthRadiationPressure: true}, ForceModelComponents{HasEarthRadiationPressure: true, EarthRadiationPressure: unused.EarthRadiationPressure})

	components := ForceModelComponents{HasTwoBody: true, HasSolarRadiationPressure: true}
	withUnused := components
	withUnused.TwoBodyMuKm3PerS2 = 1
	withUnused.SolarRadiationPressure.Shadow = shadow
	withOverride := withUnused
	withOverride.TwoBodyMuKm3PerS2Enabled = true
	different := withOverride
	different.SolarRadiationPressure.HasShadow = true
	assertOrbitalOptionalEquality(t, components, withUnused, withOverride, different)
}

func TestOrbitalOptionsForceEquality(t *testing.T) {
	unused := ForceModelComponents{ZonalMaxDegree: 2, SolarRadiationPressure: SolarRadiationPressure{CR: math.NaN()}}
	enabled := ForceModelComponents{HasZonal: true, ZonalMaxDegree: 2}
	different := ForceModelComponents{HasZonal: true, ZonalMaxDegree: 3}
	drag := DragParameters{BCFactorM2PerKg: 1}

	assertOrbitalOptionalEquality(t, PropagationConfig{},
		PropagationConfig{ForceComponents: unused, MuKm3S2: math.NaN(), Drag: drag},
		PropagationConfig{ForceComponents: enabled}, PropagationConfig{ForceComponents: different})
	assertOrbitalOptionalEquality(t, OrbitFitOptions{},
		OrbitFitOptions{ForceComponents: unused, MuKm3PerS2: math.NaN(), Drag: drag},
		OrbitFitOptions{ForceComponents: enabled}, OrbitFitOptions{ForceComponents: different})
	assertOrbitalOptionalEquality(t, OrbitParameterFitOptions{},
		OrbitParameterFitOptions{ForceComponents: unused},
		OrbitParameterFitOptions{ForceComponents: enabled}, OrbitParameterFitOptions{ForceComponents: different})
	assertOrbitalOptionalEquality(t, TCAForceOptions{},
		TCAForceOptions{ForceComponents: unused},
		TCAForceOptions{ForceComponents: enabled}, TCAForceOptions{ForceComponents: different})

	if (PropagationConfig{HasDrag: true}).Equal(PropagationConfig{HasDrag: true, Drag: drag}) ||
		(PropagationConfig{MuEnabled: true}).Equal(PropagationConfig{MuEnabled: true, MuKm3S2: 1}) ||
		(PropagationConfig{}).Equal(PropagationConfig{MaxSteps: 1}) {
		t.Fatal("propagation equality must compare supplied drag, mu and integration settings")
	}
	if (OrbitFitOptions{HasDrag: true}).Equal(OrbitFitOptions{HasDrag: true, Drag: drag}) ||
		(OrbitFitOptions{MuKm3PerS2Enabled: true}).Equal(OrbitFitOptions{MuKm3PerS2Enabled: true, MuKm3PerS2: 1}) ||
		(OrbitFitOptions{}).Equal(OrbitFitOptions{SolverMaxNFEV: 1}) {
		t.Fatal("orbit fit equality must compare supplied drag, mu and solver settings")
	}
	if (TCAForceOptions{}).Equal(TCAForceOptions{GravityTideSystem: GravityTideSystemZeroTide}) {
		t.Fatal("TCA force equality must compare the tide convention")
	}
}

func TestOrbitParameterFitOptionsEquality(t *testing.T) {
	mu, equalMu := 1.0, 1.0
	drag := DragParameters{BCFactorM2PerKg: 2}
	equalDrag := drag
	left := OrbitParameterFitOptions{MuKm3S2: &mu, Drag: &drag, EstimatedParameters: []string{"a", "b"}}
	right := OrbitParameterFitOptions{
		MuKm3S2: &equalMu, Drag: &equalDrag, EstimatedParameters: []string{"a", "b"}, PropagationContext: &PropagationContext{},
	}
	if !left.Equal(right) || !right.Equal(left) {
		t.Fatal("fit options with equal supplied values must compare equal")
	}
	for _, right := range []OrbitParameterFitOptions{
		{MuKm3S2: &equalMu, EstimatedParameters: []string{"a", "b"}},
		{Drag: &equalDrag, EstimatedParameters: []string{"a", "b"}},
		{MuKm3S2: &equalMu, Drag: &equalDrag, EstimatedParameters: []string{"b", "a"}},
		{MuKm3S2: &equalMu, Drag: &equalDrag, EstimatedParameters: []string{"a", "b"}, MaxIterations: 1},
	} {
		if left.Equal(right) || right.Equal(left) {
			t.Fatalf("fit options with different supplied values compare equal: %+v", right)
		}
	}
	if !(OrbitParameterFitOptions{}).Equal(OrbitParameterFitOptions{AdditionalForces: []Force{}, EstimatedParameters: []string{}}) {
		t.Fatal("empty force and parameter lists must compare equal")
	}
}

func TestOrbitParameterFitForceIdentity(t *testing.T) {
	first := orbitalResource[*TwoBodyGravity](t)(NewTwoBodyGravity(398600.4418))
	second := orbitalResource[*TwoBodyGravity](t)(NewTwoBodyGravity(398600.4418))
	options := OrbitParameterFitOptions{AdditionalForces: []Force{first, second}}
	if !options.Equal(OrbitParameterFitOptions{AdditionalForces: []Force{first.ForceModel, second}}) {
		t.Fatal("the same force handles must compare equal")
	}
	if options.Equal(OrbitParameterFitOptions{AdditionalForces: []Force{second, first}}) ||
		options.Equal(OrbitParameterFitOptions{AdditionalForces: []Force{first, first}}) {
		t.Fatal("fit options must compare additional force identity in order")
	}
}
