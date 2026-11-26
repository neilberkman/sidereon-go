package sidereon

import (
	"math"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestOrbitalECOMCoefficients(t *testing.T) {
	for _, entry := range []struct {
		names     []string
		count     func() uint64
		label     func(uint64) (string, error)
		parameter func(uint64) (string, error)
		prefix    string
	}{
		{[]string{"D0", "Dc", "Ds", "Y0", "Yc", "Ys", "B0", "Bc", "Bs"}, ECOM1CoefficientCount, ECOM1CoefficientsNames, ECOM1CoefficientsParameterNames, "ecom1."},
		{[]string{"B0", "B1c", "B1s", "D0", "D2c", "D2s", "Y0"}, ECOM2D2B1CoefficientCount, ECOM2D2B1CoefficientsNames, ECOM2D2B1CoefficientsParameterNames, "ecom2."},
		{[]string{"B0", "B1c", "B1s", "D0", "D2c", "D4c", "D2s", "D4s", "Y0"}, ECOM2D4B1CoefficientCount, ECOM2D4B1CoefficientsNames, ECOM2D4B1CoefficientsParameterNames, "ecom2."},
	} {
		orbitalEqual(t, uint64(len(entry.names)))(entry.count(), nil)
		for i, label := range entry.names {
			orbitalEqual(t, label)(entry.label(uint64(i)))
			orbitalEqual(t, entry.prefix+strings.ToLower(label))(entry.parameter(uint64(i)))
		}
	}
	legacy := [9]float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	ecom1 := orbitalResult[ECOM1Coefficients](t)(ECOM1CoefficientsFromArray(legacy))
	orbitalEqual(t, ECOM1Coefficients{D0: 1, Dc: 2, Ds: 3, Y0: 4, Yc: 5, Ys: 6, B0: 7, Bc: 8, Bs: 9})(ecom1, nil)
	orbitalEqual(t, legacy)(ecom1.ToArray())
	short := [7]float64{1, 2, 3, 4, 5, 6, 7}
	d2b1 := orbitalResult[ECOM2D2B1Coefficients](t)(ECOM2D2B1CoefficientsFromArray(short))
	orbitalEqual(t, ECOM2D2B1Coefficients{B0: 1, B1c: 2, B1s: 3, D0: 4, D2c: 5, D2s: 6, Y0: 7})(d2b1, nil)
	orbitalEqual(t, short)(d2b1.ToArray())
	d4b1 := orbitalResult[ECOM2D4B1Coefficients](t)(ECOM2D4B1CoefficientsFromArray(legacy))
	orbitalEqual(t, ECOM2D4B1Coefficients{B0: 1, B1c: 2, B1s: 3, D0: 4, D2c: 5, D4c: 6, D2s: 7, D4s: 8, Y0: 9})(d4b1, nil)
	orbitalEqual(t, legacy)(d4b1.ToArray())
	orbitalEqual(t, uint64(7))(ECOM2SolarRadiationPressureCoefficientCount(1, 1))
	orbitalEqual(t, uint64(9))(ECOM2SolarRadiationPressureCoefficientCount(2, 1))
	orbitalEqual(t, "ecom2.d4s")(ECOM2SolarRadiationPressureParameterName(2, 1, 7))
}

func TestOrbitalECOMForceConfiguration(t *testing.T) {
	shadow := DefaultRadiationShadowOptions()
	zero := orbitalResource[*ECOM2SolarRadiationPressure](t)(ECOM2SolarRadiationPressureD2B1(ECOM2D2B1Coefficients{}, shadow))
	wide := orbitalResource[*ECOM2SolarRadiationPressure](t)(ECOM2SolarRadiationPressureD4B1(ECOM2D4B1Coefficients{}, shadow))
	explicit := orbitalResource[*ECOM2SolarRadiationPressure](t)(NewECOM2SolarRadiationPressure(1, 1, make([]float64, 7), shadow))
	orbitalEqual(t, true)(zero.Equal(explicit))
	orbitalEqual(t, false)(zero.Equal(wide))
	orbitalEqual(t, uint64(1))(zero.ND())
	orbitalEqual(t, uint64(1))(zero.NB())
	orbitalEqual(t, uint64(2))(wide.ND())
	orbitalEqual(t, uint64(1))(wide.NB())
	if values := orbitalResult[[]float64](t)(wide.Coefficients()); !slices.Equal(values, make([]float64, 9)) {
		t.Fatalf("ECOM2 coefficients = %v", values)
	}
	orbitalEqual(t, [3]float64{})(wide.AccelerationWithSunMoon(orbitalPosition, orbitalVelocity, orbitalSun, orbitalMoon))
	orbitalEqual(t, AccelerationPartials{})(wide.AccelerationPartialsWithSunMoon(orbitalPosition, orbitalVelocity, orbitalSun, orbitalMoon))
	orbitalEqual(t, true)(wide.MoonEclipseEnabled())
	orbitalEqual(t, ShadowFunctionApparentDisc)(wide.ShadowFunction())
	orbitalEqual(t, EarthShadowWGS84Oblate)(wide.EarthShadowModel())
	noMoon := orbitalResource[*ECOM2SolarRadiationPressure](t)(wide.WithMoonEclipse(false))
	orbitalEqual(t, false)(noMoon.MoonEclipseEnabled())
	orbitalEqual(t, [3]float64{})(noMoon.AccelerationWithSun(orbitalPosition, orbitalVelocity, orbitalSun))
	orbitalEqual(t, AccelerationPartials{})(noMoon.AccelerationPartialsWithSun(orbitalPosition, orbitalVelocity, orbitalSun))
	conical := orbitalResource[*ECOM2SolarRadiationPressure](t)(wide.WithShadowFunction(ShadowFunctionConical))
	orbitalEqual(t, ShadowFunctionConical)(conical.ShadowFunction())
	orbitalEqual(t, false)(conical.MoonEclipseEnabled())
	spherical := orbitalResource[*ECOM2SolarRadiationPressure](t)(wide.WithEarthShadowModel(EarthShadowSpherical))
	orbitalEqual(t, EarthShadowSpherical)(spherical.EarthShadowModel())
}

func TestOrbitalBroadcastSourceOwnership(t *testing.T) {
	product := orbitalResource[*BroadcastEphemeris](t)(ParseBroadcastEphemeris(readPositioningFixture(t, filepath.Join("nav", "BRD400DLR_S_20261800000_01H_MN_trim.rnx"))))
	source := orbitalResource[*BroadcastOrbit](t)(NewBroadcastOrbit(product, orbitalClockEpoch(GPST, 646315200)))
	orbitalEqual(t, GPST)(source.TimeScale())
	orbitalEqual(t, OrbitReferencePointAntennaPhaseCenter)(source.ReferencePoint())
	orbitalEqual(t, true)(source.ClockIncludesRelativity())
	orbitalEqual(t, OrbitFrame{Kind: OrbitFrameKindBroadcastDatums})(source.Frame())
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: 646315200})(source.PredictionOriginJ2000S("G08"))
}

func TestOrbitalGNSSConstants(t *testing.T) {
	radians := func(degrees float64) float64 { return degrees * math.Pi / 180 }
	for _, entry := range []struct {
		accessor func() float64
		degrees  float64
	}{
		{BeiDou2OnToYsBetaRad, 3.9},
		{BeiDou2YsToOnBetaRad, 4.1},
		{BeiDou3SECMBeta0Rad, 3},
		{GalileoIOVBetaXRad, 15},
		{GalileoIOVBetaYRad, 2},
		{GLONASSMDefaultYawRateRadS, 0.25},
		{GLONASSMShadowLimitRad, 165.8},
		{GPSIIADefaultYawBiasRad, 0.5},
		{GPSIIFDefaultYawBiasRad, -0.7},
		{GPSIIFDefaultYawRateRadS, 0.11},
		{GPSIIIBetaXRad, 15},
		{GPSIIIBetaYRad, 5.8},
		{GPSIIRDefaultYawRateRadS, 0.2},
		{GPSShadowLimitRad, 166.75},
	} {
		orbitalEqual(t, radians(entry.degrees))(entry.accessor(), nil)
	}
	orbitalEqual(t, 0.1211*float64(math.Pi/180))(GPSIIADefaultYawRateRadS(1))
	orbitalEqual(t, CosineTurnParameters{BetaLimitRad: radians(3), BetaLimitInclusive: true,
		StartAngle: CosineTurnStartAngleOrbitAngle, StartAngleRad: radians(6), PeriodS: 3090})(BeiDou3CASTMEOTurn(), nil)
	orbitalEqual(t, CosineTurnParameters{BetaLimitRad: radians(3), BetaLimitInclusive: true,
		StartAngle: CosineTurnStartAngleOrbitAngle, StartAngleRad: radians(6), PeriodS: 5740})(BeiDou3CASTIGSOTurn(), nil)
	orbitalEqual(t, CosineTurnParameters{BetaLimitRad: radians(4.1),
		StartAngle: CosineTurnStartAngleColinearity, StartAngleRad: radians(10), PeriodS: 5656})(GalileoFOCTurn(), nil)
	_, err := GNSSBlockFromIGSName("UNRECOGNIZED")
	requireOrbitalError[*GNSSBlockError](t, err, StatusInvalidArgument, "unknown", `{"name":"UNRECOGNIZED"}`)
	_, err = GNSSBlockFromConstellationBlockType(GNSSSystemBeiDou, "BDS-3", nil)
	requireOrbitalError[*GNSSBlockError](t, err, StatusInvalidArgument, "beidou_orbit_class_required", `{"block_type":"BDS-3"}`)
}

func TestOrbitalGNSSBlockLawSelection(t *testing.T) {
	rad := func(degrees float64) float64 { return degrees * math.Pi / 180 }
	castTurn := CosineTurnParameters{BetaLimitRad: rad(3), BetaLimitInclusive: true,
		StartAngle: CosineTurnStartAngleOrbitAngle, StartAngleRad: rad(6), PeriodS: 3090}
	focTurn := CosineTurnParameters{BetaLimitRad: rad(4.1), StartAngle: CosineTurnStartAngleColinearity,
		StartAngleRad: rad(10), PeriodS: 5656}
	tests := []struct {
		block GNSSBlock
		prn   uint8
		law   func() (*GNSSAttitudeLaw, error)
	}{
		{GNSSBlockGPSIIA, 1, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawGPSIIA(0.1211*float64(math.Pi/180), rad(0.5), rad(166.75))
		}},
		{GNSSBlockGPSIIF, 0, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawGPSIIF(rad(0.11), rad(-0.7), rad(166.75))
		}},
		{GNSSBlockGLONASSM, 0, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawGLONASSM(rad(0.25), rad(165.8))
		}},
		{GNSSBlockGalileoIOV, 0, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawGalileoIOV(rad(15), rad(2))
		}},
		{GNSSBlockGalileoFOC, 0, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawCosineTurn(focTurn)
		}},
		{GNSSBlockBeiDou2IGSO, 0, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawBeiDouOrbitNormalSwitch(rad(4.1), rad(3.9))
		}},
		{GNSSBlockBeiDou3MEOSECM, 0, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawFixedBetaYawSteering(rad(3))
		}},
		{GNSSBlockBeiDou3MEOUnknown, 0, func() (*GNSSAttitudeLaw, error) {
			return GNSSAttitudeLawBeiDou3MEOUnknownManufacturer(castTurn, rad(3))
		}},
	}
	for _, test := range tests {
		t.Run(test.block.String(), func(t *testing.T) {
			byBlock := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawForBlock(test.block, test.prn))
			expected := orbitalResource[*GNSSAttitudeLaw](t)(test.law())
			orbitalEqual(t, true)(byBlock.Equal(expected))
		})
	}
}

func TestOrbitalDecayFixture(t *testing.T) {
	want := readOrbitalFixture(t).Reentry
	config := orbitalResult[DecayConfig](t)(DefaultDecayConfig())
	config.ReentryAltitudeKm = 125
	config.Drag = orbitalResult[DragParameters](t)(DragParametersFromAreaMass(2.2, 10, 1000,
		SpaceWeather{F107: 70, F107A: 70, Ap: 4}, 120))
	state := CartesianState{EpochTDBSeconds: orbitalEpoch, PositionKm: [3]float64{6468.137, 0, 0}, VelocityKmPerS: [3]float64{0, 7.8, 0}}
	expected := DecayEstimate{TimeToDecayS: want.Time, ReentryState: state, ReentryAltitudeKm: want.Altitude}
	orbitalEqual(t, expected)(EstimateDecayWithContext(state, config, &PropagationContext{}))
	table := orbitalResource[*SpaceWeatherTable](t)(ParseSpaceWeatherTable(readPositioningFixture(t, filepath.Join("space_weather", "SW-All-20260702-trim.csv"))))
	orbitalEqual(t, expected)(EstimateDecayWithSourceAndContext(state, config, table, nil))
}
