package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"math"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const orbitalEpoch = 646315200.25

var (
	orbitalPosition = [3]float64{7000, 20, 30}
	orbitalVelocity = [3]float64{0, 7.5, 1}
	orbitalSun      = [3]float64{150000000, 20000000, -10000000}
	orbitalMoon     = [3]float64{-100000, 200000, 300000}
)

func orbitalResult[T any](t *testing.T) func(T, error) T {
	t.Helper()
	return func(value T, err error) T {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
}

func orbitalEqual[T comparable](t *testing.T, expected T) func(T, error) {
	t.Helper()
	return func(value T, err error) {
		t.Helper()
		if err != nil || value != expected {
			t.Fatalf("value = %+v, error = %v, want %+v", value, err, expected)
		}
	}
}

func orbitalResource[T interface{ Close() error }](t *testing.T) func(T, error) T {
	t.Helper()
	return func(value T, err error) T {
		t.Helper()
		resource := orbitalResult[T](t)(value, err)
		cleanupClose(t, "orbital resource", resource.Close)
		return resource
	}
}

type attitudePin struct {
	Epoch    float64
	Sample   AttitudeSample
	Attitude GNSSAttitude
}

type orbitalDynamicsFixture struct {
	Contracts     map[string][]float64   `json:"contracts"`
	Attitude      map[string]attitudePin `json:"attitude"`
	AttitudeFinal CartesianState         `json:"attitude_final"`

	Reentry struct {
		Time     float64 `json:"time"`
		Altitude float64 `json:"altitude"`
	} `json:"reentry"`
	FittedForce            EstimatedForceParameter `json:"fitted_force"`
	FittedForceSigma       float64                 `json:"fitted_force_sigma"`
	SP3Fit                 [2][3]float64           `json:"sp3_fit"`
	SP3FitEpoch            float64                 `json:"sp3_fit_epoch"`
	SP3ParameterFit        [2][3]float64           `json:"sp3_parameter_fit"`
	SP3ParameterFitEpoch   float64                 `json:"sp3_parameter_fit_epoch"`
	Covariance             [6][6]float64           `json:"covariance"`
	CovarianceState        [2][3]float64           `json:"covariance_state"`
	StateTransition        [6][6]float64           `json:"state_transition"`
	ScoreEpochs            uint64                  `json:"score_epochs"`
	ScoreScored            uint64                  `json:"score_scored"`
	ScoreBins              []uint64                `json:"score_bins"`
	ConvertedMeanAnomaly   float64                 `json:"converted_mean_anomaly"`
	Time                   [3]float64              `json:"time"`
	Sun                    [3]float64              `json:"sun"`
	Moon                   [3]float64              `json:"moon"`
	SolarRadiation         [3]float64              `json:"solar_radiation"`
	SolarRadiationPartials [3][3]float64           `json:"solar_radiation_partials"`
	EarthRadiation         [3]float64              `json:"earth_radiation"`
	ThirdBody              [3]float64              `json:"third_body"`
	ECOM1                  [3]float64              `json:"ecom1"`
	GravityAcceleration    [3]float64              `json:"gravity_acceleration"`
	GravityPotential       float64                 `json:"gravity_potential"`
	Drag                   [3]float64              `json:"drag"`
	NominalYaw             NominalYaw              `json:"nominal_yaw"`
	DYB                    DYBAxes                 `json:"dyb"`
	ShadowFraction         float64                 `json:"shadow_fraction"`
	ShadowGradient         [3]float64              `json:"shadow_gradient"`
	EcksteinHechler        [2][3]float64           `json:"eckstein_hechler"`
	BrouwerLyddane         [2][3]float64           `json:"brouwer_lyddane"`
	TwoBody                [2][3]float64           `json:"two_body"`
	TwoBodyJ2              [2][3]float64           `json:"two_body_j2"`
	ICGEMPotential         float64                 `json:"icgem_potential"`
	IERSQuaternion         [4]float64              `json:"iers_quaternion"`
	DE                     struct {
		Version      uint32             `json:"version"`
		RecordSize   uint64             `json:"record_size"`
		Start        float64            `json:"start"`
		End          float64            `json:"end"`
		Constants    map[string]float64 `json:"constants"`
		AU           float64            `json:"astronomical_unit_m"`
		MassRatio    float64            `json:"earth_moon_mass_ratio"`
		GMSun        float64            `json:"gm_sun"`
		SunPosition  [3]float64         `json:"sun_position"`
		SunVelocity  [3]float64         `json:"sun_velocity"`
		MoonPosition [3]float64         `json:"moon_position"`
		MoonVelocity [3]float64         `json:"moon_velocity"`
	} `json:"de"`
	Fit struct {
		Covariance      OrbitParameterCovariance `json:"covariance"`
		Observations    [][3]float64             `json:"observations"`
		Initial         [2][3]float64            `json:"initial"`
		InitialEpoch    float64                  `json:"initial_epoch"`
		ResidualRMSM    float64                  `json:"residual_rms_m"`
		SeedRMSM        float64                  `json:"seed_rms_m"`
		ConditionNumber float64                  `json:"condition_number"`
		Iterations      uint64                   `json:"iterations"`
		SigmaXKm        float64                  `json:"sigma_x_km"`
		CorrelationXY   float64                  `json:"correlation_xy"`
		ParameterNames  []string                 `json:"parameter_names"`
		Residuals       [][5]float64             `json:"residuals"`
		IllConditioned  json.RawMessage          `json:"ill_conditioned"`
		NonConvergent   json.RawMessage          `json:"non_convergent"`
	} `json:"fit"`
}

func readOrbitalFixture(t *testing.T) orbitalDynamicsFixture {
	t.Helper()
	var fixture orbitalDynamicsFixture
	data := readPositioningFixture(t, filepath.Join("orbital", "dynamics.json"))
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestOrbitalForceFixture(t *testing.T) {
	want := readOrbitalFixture(t)
	state := CartesianState{EpochTDBSeconds: orbitalEpoch, PositionKm: orbitalPosition, VelocityKmPerS: orbitalVelocity}
	shadow := DefaultRadiationShadowOptions()
	if shadow != (RadiationShadowOptions{ShadowFunction: ShadowFunctionApparentDisc, EarthShadowModel: EarthShadowWGS84Oblate, MoonEclipseEnabled: true}) {
		t.Fatalf("radiation defaults = %+v", shadow)
	}
	solar := orbitalResource[*SolarRadiationPressureForce](t)(NewSolarRadiationPressureForce(1.25, 0.02, shadow))
	orbitalEqual(t, 1.25)(solar.CR())
	orbitalEqual(t, 0.02)(solar.AreaToMassM2Kg())
	orbitalEqual(t, 149597870.7)(solar.AUKm())
	orbitalEqual(t, ShadowFunctionApparentDisc)(solar.ShadowFunction())
	orbitalEqual(t, EarthShadowWGS84Oblate)(solar.EarthShadowModel())
	orbitalEqual(t, true)(solar.MoonEclipseEnabled())
	orbitalEqual(t, want.SolarRadiation)(solar.Acceleration(state, nil))
	orbitalEqual(t, want.SolarRadiation)(solar.AccelerationWithParameters(state, []float64{1.25}, nil))
	partials := orbitalResult[AccelerationPartials](t)(solar.AccelerationPartials(state, []float64{1.25}, nil))
	for i, row := range want.SolarRadiationPartials {
		for j, value := range row {
			orbitalEqual(t, value)(partials.WrtPosition[i*3+j], nil)
		}
	}
	orbitalEqual(t, [9]float64{})(partials.WrtVelocity, nil)
	orbitalEqual(t, uint64(1))(solar.ParameterCount())
	earth := orbitalResource[*EarthRadiationPressure](t)(NewEarthRadiationPressure(1.25, 0.02, true))
	orbitalEqual(t, want.EarthRadiation)(earth.Acceleration(state, nil))
	orbitalEqual(t, true)(earth.MoonEclipseEnabled())
	noLunar := orbitalResource[*EarthRadiationPressure](t)(earth.WithMoonEclipse(false))
	orbitalEqual(t, false)(noLunar.MoonEclipseEnabled())
	thirdBody := orbitalResource[*ThirdBodyGravity](t)(NewThirdBodyGravity(true, true))
	orbitalEqual(t, want.ThirdBody)(thirdBody.Acceleration(state, nil))
	ecom := orbitalResource[*ECOM1SolarRadiationPressure](t)(NewECOM1SolarRadiationPressure(ECOM1Coefficients{D0: 1e-7}, shadow))
	orbitalEqual(t, want.ECOM1)(ecom.AccelerationWithSunMoon(orbitalPosition, orbitalVelocity, orbitalSun, orbitalMoon))
	orbitalEqual(t, ECOM1Coefficients{D0: 1e-7})(ecom.Coefficients())
	orbitalEqual(t, true)(ecom.Equal(ecom))
	other := orbitalResource[*ECOM1SolarRadiationPressure](t)(ecom.WithMoonEclipse(false))
	orbitalEqual(t, false)(ecom.Equal(other))
	orbitalEqual(t, false)(other.MoonEclipseEnabled())
	gravity := orbitalResource[*SphericalHarmonicGravity](t)(SphericalHarmonicGravityEGM96(36, 36))
	orbitalEqual(t, want.GravityAcceleration)(gravity.BodyFixedAccelerationKmS2(orbitalPosition))
	orbitalEqual(t, want.GravityPotential)(gravity.BodyFixedPotentialKm2S2(orbitalPosition))
	orbitalEqual(t, false)(gravity.IsTimeVariable())
	orbitalEqual(t, true)(gravity.Equal(gravity))
	coarse := orbitalResource[*SphericalHarmonicGravity](t)(SphericalHarmonicGravityEGM96(2, 0))
	orbitalEqual(t, false)(gravity.Equal(coarse))
	orbitalEqual(t, uint64(0))(gravity.ParameterCount())
	drag := orbitalResult[DragParameters](t)(DragParametersFromAreaMass(2.2, 10, 1000, SpaceWeather{F107: 70, F107A: 70, Ap: 4}, 120))
	dragState := CartesianState{EpochTDBSeconds: orbitalEpoch, PositionKm: [3]float64{6778.137, 20, 30}, VelocityKmPerS: [3]float64{0, 4.75, 6}}
	orbitalEqual(t, want.Drag)(DragForceAcceleration(drag, dragState))
	dragForce := orbitalResource[*DragForce](t)(NewDragForce(drag))
	orbitalEqual(t, want.Drag)(dragForce.Acceleration(dragState, nil))
}

func TestOrbitalAnalyticFixture(t *testing.T) {
	want := readOrbitalFixture(t)
	orbitalEqual(t, want.Time)([3]float64{TDBMinusTTSeconds(orbitalEpoch), TDBFromTTJ2000Seconds(orbitalEpoch), TTFromTDBJ2000Seconds(orbitalEpoch)}, nil)
	orbitalEqual(t, SunMoon{SunM: want.Sun, MoonM: want.Moon})(SunMoonGCRSAtTDBSeconds(orbitalEpoch))
	orbitalEqual(t, want.NominalYaw)(ComputeNominalYaw(AttitudeSample{PositionM: orbitalPosition, VelocityMS: orbitalVelocity, SunPositionM: orbitalSun}))
	orbitalEqual(t, want.DYB)(ComputeDYBAxes(orbitalPosition, orbitalSun))
	orbitalEqual(t, ShadowGradient{Fraction: want.ShadowFraction, GradientPerKm: want.ShadowGradient})(
		ShadowFractionWithModelAndGradient([3]float64{-7000, 6378, 0}, [3]float64{150000000, 0, 0}, EarthShadowWGS84Oblate))
	gravity := ZonalHarmonicsEarth()
	eh := orbitalResource[*EcksteinHechlerPropagator](t)(EcksteinHechlerPropagatorFromMeanElements(orbitalEpoch,
		CircularElements{AKm: 7200, Ex: 0.001, Ey: 0.002, IRad: 0.5, RaanRad: 0.25, MeanLatitudeArgumentRad: 0.125}, gravity))
	bl := orbitalResource[*BrouwerLyddanePropagator](t)(BrouwerLyddanePropagatorFromMeanElements(orbitalEpoch,
		AnalyticKeplerianElements{AKm: 7200, E: 0.01, IRad: 0.5, RaanRad: 0.25, ArgPerigeeRad: 0.75, MeanAnomalyRad: 0.125}, gravity, 0))
	for _, tc := range []struct {
		epoch     func() (float64, error)
		gravity   func() (ZonalHarmonics, error)
		propagate func(float64) (CartesianState, error)
		ephemeris func([]float64) ([]CartesianState, error)
		state     [2][3]float64
	}{
		{eh.EpochTDBSeconds, eh.Gravity, eh.Propagate, eh.Ephemeris, want.EcksteinHechler},
		{bl.EpochTDBSeconds, bl.Gravity, bl.Propagate, bl.Ephemeris, want.BrouwerLyddane},
	} {
		end := orbitalEpoch + 600
		state := CartesianState{EpochTDBSeconds: end, PositionKm: tc.state[0], VelocityKmPerS: tc.state[1]}
		orbitalEqual(t, orbitalEpoch)(tc.epoch())
		orbitalEqual(t, gravity)(tc.gravity())
		orbitalEqual(t, state)(tc.propagate(end))
		rows := orbitalResult[[]CartesianState](t)(tc.ephemeris([]float64{end}))
		if !slices.Equal(rows, []CartesianState{state}) {
			t.Fatalf("ephemeris = %+v, want %+v", rows, state)
		}
	}
	orbitalEqual(t, 0.0)(bl.M2RadS2())
	orbitalEqual(t, "Brouwer-Lyddane")(AnalyticTheoryBrouwerLyddane.String(), nil)
	orbitalEqual(t, "Brouwer-Lyddane")(AnalyticTheoryText(AnalyticTheoryBrouwerLyddane))
	start := AnalyticCartesian{PositionKm: [3]float64{4, 0, 0}, VelocityKmS: [3]float64{0, 1, 0}}
	orbitalEqual(t, start)((CircularElements{AKm: 4}).ToCartesian(4))
	orbitalEqual(t, start)((AnalyticKeplerianElements{AKm: 4}).ToCartesian(4))
	orbitalEqual(t, CircularElements{AKm: 4, RaanRad: math.Pi, MeanLatitudeArgumentRad: -math.Pi})(CircularElementsFromCartesian(start, 4))
	orbitalEqual(t, AnalyticKeplerianElements{AKm: 4, RaanRad: math.Pi, ArgPerigeeRad: math.Pi, MeanAnomalyRad: want.ConvertedMeanAnomaly})(AnalyticKeplerianElementsFromCartesian(start, 4))
	orbitalEqual(t, 0.25)((CircularElements{Ex: 0.25}).Eccentricity(), nil)
}

func TestOrbitalDataFixture(t *testing.T) {
	want := readOrbitalFixture(t)
	data := readPositioningFixture(t, filepath.Join("orbital", "lnxp1990.440"))
	de := orbitalResource[*DEEphemeris](t)(DEEphemerisFromBytes(data))
	loaded := orbitalResource[*DEEphemeris](t)(LoadDEEphemeris(filepath.Join("testdata", "orbital", "lnxp1990.440")))
	for _, value := range []*DEEphemeris{de, loaded} {
		orbitalEqual(t, want.DE.Version)(value.Version())
		orbitalEqual(t, want.DE.RecordSize)(value.RecordSize())
		orbitalEqual(t, DECoverage{StartTDBSeconds: want.DE.Start, EndTDBSeconds: want.DE.End})(value.CoverageTDBSeconds())
		orbitalEqual(t, want.DE.AU)(value.AstronomicalUnitM())
		orbitalEqual(t, want.DE.MassRatio)(value.EarthMoonMassRatio())
		orbitalEqual(t, DEGM{Present: true, ValueM3S2: want.DE.GMSun})(value.GMM3S2(DEBodySun))
		orbitalEqual(t, DESunMoon{Sun: DEState{PositionM: want.DE.SunPosition, VelocityMS: want.DE.SunVelocity},
			Moon: DEState{PositionM: want.DE.MoonPosition, VelocityMS: want.DE.MoonVelocity}})(value.SunMoonAtTDBSeconds(want.DE.Start + 3600))
		orbitalEqual(t, SunMoon{SunM: want.DE.SunPosition, MoonM: want.DE.MoonPosition})(value.SunMoonPositionsAtTDBSeconds(want.DE.Start + 3600))
		constants := orbitalResult[map[string]float64](t)(value.Constants())
		if !maps.Equal(constants, want.DE.Constants) {
			t.Fatalf("DE constants = %v, want %v", constants, want.DE.Constants)
		}
	}
	field := orbitalResource[*ICGEMGravityField](t)(ParseICGEMGravityField(string(readPositioningFixture(t, filepath.Join("orbital", "synthetic_icgem2.gfc")))))
	header := orbitalResult[ICGEMHeader](t)(field.Header())
	orbitalEqual(t, ICGEMFormatICGEM2)(header.Format, nil)
	orbitalEqual(t, true)(field.IsTimeVariable())
	orbitalEqual(t, true)(field.Equal(field))
	gravity := orbitalResource[*SphericalHarmonicGravity](t)(field.Gravity(2, 2))
	orbitalEqual(t, want.ICGEMPotential)(gravity.BodyFixedPotentialAtKm2S2(orbitalPosition, 86400))
	orbitalEqual(t, true)(gravity.IsTimeVariable())
	eop := IERS2010EopSample{MJDUTC: 59025, XpArcsec: 0.1, YpArcsec: 0.2, Ut1MinusUTCS: -0.2,
		LodS: 0.001, DxArcsec: 0.0001, DyArcsec: -0.0001, HasPoleRates: true}
	following := eop
	following.MJDUTC++
	orientation := orbitalResource[*IERS2010EarthOrientationProvider](t)(NewIERS2010EarthOrientationProvider([]IERS2010EopSample{eop, following}))
	result := orbitalResult[EarthOrientation](t)(orientation.OrientationAtUTCMJD(59025, 43200))
	orbitalEqual(t, want.IERSQuaternion)(result.IERS2010Quaternion, nil)
	orbitalEqual(t, true)(result.HasIERS2010Quaternion, nil)
	orbitalEqual(t, true)(orientation.Equal(orientation))
	permissive := orbitalResource[*IERS2010EarthOrientationProvider](t)(orientation.WithValidity(true))
	orbitalEqual(t, false)(orientation.Equal(permissive))
}

func TestOrbitalExactRefusals(t *testing.T) {
	_, err := ParseICGEMGravityField("modelname X\n")
	requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "missing_end_of_head", "{}")
	_, err = GNSSBlockFromIGSName("BLOCK I")
	requireOrbitalError[*GNSSBlockError](t, err, StatusInvalidArgument, "unsupported", `{"name":"BLOCK I"}`)
	_, err = GNSSAttitudeLawForBlock(GNSSBlockGPSIIA, 0)
	requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, "no_default_yaw_rate", `{"prn":null}`)
	_, err = ComputeDYBAxes([3]float64{}, orbitalSun)
	requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, "zero_length", `{"field":"position_m"}`)
	_, err = ComputeDYBAxes([3]float64{math.Inf(1)}, orbitalSun)
	requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, "non_finite", `{"field":"position_m"}`)
	_, err = ComputeDYBAxes([3]float64{1}, [3]float64{2})
	requireOrbitalError[*GNSSAttitudeError](t, err, StatusInvalidArgument, "degenerate_geometry", `{"t_s":null}`)
	_, err = ShadowFractionWithGradient(orbitalPosition, [3]float64{})
	requireOrbitalError[*ShadowGradientError](t, err, StatusInvalidArgument, "eclipse", `{"cause":{"fields":{"field":"sun_pos","reason":"zero vector"},"kind":"invalid_input"}}`)
	_, err = (CircularElements{AKm: math.NaN()}).ToCartesian(4)
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "non_finite", `{"field":"elements"}`)
	de := orbitalResource[*DEEphemeris](t)(DEEphemerisFromBytes(readPositioningFixture(t, filepath.Join("orbital", "lnxp1990.440"))))
	_, err = de.SunMoonAtTDBSeconds(math.NaN())
	requireOrbitalError[*DEError](t, err, StatusSolve, "non_finite_epoch", "{}")
	_, err = de.SunMoonAtTDBSeconds(1e20)
	requireOrbitalError[*DEError](t, err, StatusSolve, "outside_coverage", `{"epoch_tdb_seconds":{"bits_hex":"4415af1d78b58c40","decimal":"100000000000000000000"}}`)
}

func requireOrbitalError[T error](t *testing.T, err error, code StatusCode, kind, fields string) {
	t.Helper()
	var status *StatusError
	var typed T
	if !errors.As(err, &status) || status.Code != code || !errors.As(err, &typed) ||
		status.Engine == nil || status.Engine.Kind != kind || !orbitalFieldsEqual(t, status.Engine, fields) {
		t.Fatalf("error = %v, status = %+v, fields = %s, want %s %s", err, status, orbitalErrorFields(status), kind, fields)
	}
}

func TestOrbitalVariationalFixture(t *testing.T) {
	gravity := orbitalResource[*TwoBodyGravity](t)(NewTwoBodyGravity(8))
	term := EmpiricalRTNTerm{Axis: RTNAxisRadial, Harmonic: EmpiricalHarmonicConstant, CoefficientKmS2: 0.125}
	empirical := orbitalResource[*EmpiricalRTNAcceleration](t)(NewEmpiricalRTNAcceleration([]EmpiricalRTNTerm{term}))
	model := orbitalResource[*CompositeForceModel](t)(NewCompositeForceModel(gravity, empirical))
	terms := orbitalResult[[]EmpiricalRTNTerm](t)(empirical.Terms())
	if !slices.Equal(terms, []EmpiricalRTNTerm{term}) {
		t.Fatalf("empirical terms = %+v", terms)
	}
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: 0.125})(empirical.CoefficientKmS2(RTNAxisRadial, EmpiricalHarmonicConstant))
	orbitalEqual(t, OptionalOrbitScalar{})(empirical.CoefficientKmS2(RTNAxisAlongTrack, EmpiricalHarmonicConstant))
	orbitalEqual(t, "empirical.radial.constant")(EmpiricalRTNAccelerationParameterName(RTNAxisRadial, EmpiricalHarmonicConstant))
	orbitalEqual(t, true)(empirical.Equal(empirical))
	modified := orbitalResource[*EmpiricalRTNAcceleration](t)(empirical.WithTerm(EmpiricalRTNTerm{Axis: RTNAxisCrossTrack, Harmonic: EmpiricalHarmonicSine, CoefficientKmS2: 0.03125}))
	orbitalEqual(t, false)(empirical.Equal(modified))
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: 0.03125})(modified.CoefficientKmS2(RTNAxisCrossTrack, EmpiricalHarmonicSine))
	orbitalEqual(t, [3]RTNAxis{RTNAxisRadial, RTNAxisAlongTrack, RTNAxisCrossTrack})(RTNAxisAll())
	orbitalEqual(t, [3]EmpiricalHarmonic{EmpiricalHarmonicConstant, EmpiricalHarmonicCosine, EmpiricalHarmonicSine})(EmpiricalHarmonicAll())
	orbitalEqual(t, -1)(RTNAxisRadial.Compare(RTNAxisCrossTrack), nil)
	orbitalEqual(t, 1)(EmpiricalHarmonicSine.Compare(EmpiricalHarmonicConstant), nil)
	constant := orbitalResource[*EmpiricalRTNAcceleration](t)(EmpiricalRTNAccelerationConstantTerms())
	all := orbitalResource[*EmpiricalRTNAcceleration](t)(EmpiricalRTNAccelerationAllTerms())
	zero := orbitalResource[*EmpiricalRTNAcceleration](t)(EmpiricalRTNAccelerationWithZeroTerms([]EmpiricalRTNTermSelector{{Axis: RTNAxisRadial, Harmonic: EmpiricalHarmonicConstant}}))
	orbitalEqual(t, uint64(3))(constant.ParameterCount())
	orbitalEqual(t, uint64(9))(all.ParameterCount())
	orbitalEqual(t, OptionalOrbitScalar{Present: true})(zero.CoefficientKmS2(RTNAxisRadial, EmpiricalHarmonicConstant))
	if err := errors.Join(gravity.Close(), empirical.Close()); err != nil {
		t.Fatal(err)
	}
	state := CartesianState{PositionKm: [3]float64{2, 0, 0}, VelocityKmPerS: [3]float64{0, 2, 0}}
	orbitalEqual(t, [3]float64{-1.875})(model.Acceleration(state, nil))
	orbitalEqual(t, [3]float64{-1.75})(model.AccelerationWithParameters(state, []float64{0.25}, nil))
	parameters := orbitalResult[[]ForceParameter](t)(model.Parameters())
	if !slices.Equal(parameters, []ForceParameter{{Name: "empirical.radial.constant", Value: 0.125}}) {
		t.Fatalf("model parameters = %v", parameters)
	}
	derivatives := orbitalResult[[][3]float64](t)(model.ParameterPartials(state, []float64{0.125}, nil))
	if !slices.Equal(derivatives, [][3]float64{{1, 0, 0}}) {
		t.Fatalf("parameter partials = %v", derivatives)
	}
	initialEpoch := VariationalEpochFromTDBSeconds(0, 0)
	orbitalEqual(t, VariationalEpoch{})(initialEpoch, nil)
	options := DefaultIntegratorOptions()
	result := orbitalResource[*VariationalTrajectory](t)(PropagateVariational(state, []VariationalEpoch{initialEpoch}, VariationalEquations{
		Force: model, Parameters: []float64{0.125}, SensitivityParameters: []uint64{0},
	}, options, nil))
	orbitalEqual(t, uint64(1))(result.Count())
	var identity [36]float64
	for i := 0; i < 36; i += 7 {
		identity[i] = 1
	}
	orbitalEqual(t, VariationalState{State: state, StateTransition: identity, SensitivityColumns: 1})(result.State(0))
	if values := orbitalResult[[]float64](t)(result.Sensitivity(0)); !slices.Equal(values, []float64{0, 0, 0, 0, 0, 0}) {
		t.Fatalf("sensitivity = %v", values)
	}
	orbitalEqual(t, AccelerationPartials{})(AccelerationPartialsZeros(), nil)
	orbitalEqual(t, AccelerationPartials{WrtPosition: [9]float64{1, 2, 3, 4, 5, 6, 7, 8, 9}})(
		AccelerationPartialsPositionOnly([3][3]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}))
}

func TestOrbitalParameterFitFixture(t *testing.T) {
	want := readOrbitalFixture(t).Fit
	samples := make([]InertialPositionSample, len(want.Observations))
	for i, position := range want.Observations {
		samples[i] = InertialPositionSample{Epoch: orbitalClockEpoch(TDB, uint64(i)*60), PositionGCRFM: position}
	}
	options := DefaultOrbitParameterFitOptions()
	orbitalEqual(t, 1.0)(options.ObservationSigmaM, nil)
	orbitalEqual(t, uint64(25))(options.MaxIterations, nil)
	orbitalEqual(t, 1e-3)(options.ConvergenceThreshold, nil)
	orbitalEqual(t, 1e12)(options.MaxConditionNumber, nil)
	orbitalEqual(t, uint64(3))(options.MinLedgerSamples, nil)
	orbitalEqual(t, PropagationForceModelEarthPhaseA)(options.ForceModel, nil)
	orbitalEqual(t, GravityTideSystemTideFree)(options.GravityTideSystem, nil)
	orbitalEqual(t, (*float64)(nil))(options.MuKm3S2, nil)
	orbitalEqual(t, (*DragParameters)(nil))(options.Drag, nil)
	orbitalEqual(t, (*SpaceWeatherTable)(nil))(options.SpaceWeather, nil)
	orbitalEqual(t, (*PropagationContext)(nil))(options.PropagationContext, nil)
	if options.AdditionalForces != nil || options.EstimatedParameters != nil {
		t.Fatalf("unexpected default force or parameter selection: %+v", options)
	}
	orbitalEqual(t, IntegratorOptions{Integrator: PropagationIntegratorDP54, AbsTol: 1e-9, RelTol: 1e-12,
		InitialStepS: 60, MinStepS: 1e-6, MaxStepS: 3600, MaxSteps: 1000000})(options.IntegratorOptions, nil)
	options.ForceModel = PropagationForceModelTwoBody
	fit := orbitalResource[*OrbitParameterFit](t)(FitInertialOrbitParameters(samples, "G03", nil, options))
	summary := orbitalResult[OrbitParameterFitSummary](t)(fit.Summary())
	orbitalEqual(t, "G03")(summary.Satellite, nil)
	orbitalEqual(t, want.InitialEpoch)(summary.InitialState.EpochTDBSeconds, nil)
	orbitalEqual(t, want.Initial[0])(summary.InitialState.PositionKm, nil)
	orbitalEqual(t, want.Initial[1])(summary.InitialState.VelocityKmPerS, nil)
	orbitalEqual(t, want.ResidualRMSM)(summary.ResidualStats.RMS3DM, nil)
	orbitalEqual(t, want.SeedRMSM)(summary.SeedRMS3dM, nil)
	orbitalEqual(t, want.ConditionNumber)(summary.ConditionNumber, nil)
	orbitalEqual(t, want.Iterations)(summary.Iterations, nil)
	orbitalEqual(t, want.Covariance)(fit.Covariance())
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: want.SigmaXKm})(fit.Sigma("x_km"))
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: want.CorrelationXY})(fit.Correlation("x_km", "y_km"))
	orbitalEqual(t, EstimatedForceParameter{})(fit.ForceParameter("x_km"))
	if values := orbitalResult[[]EstimatedForceParameter](t)(fit.ForceParameters()); len(values) != 0 {
		t.Fatalf("force parameters = %+v", values)
	}
	for _, read := range []func() ([]string, error){fit.ParameterNames, StateParameterNames} {
		if values := orbitalResult[[]string](t)(read()); !slices.Equal(values, want.ParameterNames) {
			t.Fatalf("parameter names = %v, want %v", values, want.ParameterNames)
		}
	}
	values := orbitalResult[[]OrbitRTNResidual](t)(fit.Residuals())
	if len(values) != len(want.Residuals) {
		t.Fatalf("residual count = %d, want %d", len(values), len(want.Residuals))
	}
	for i, value := range values {
		orbitalEqual(t, want.Residuals[i])([5]float64{value.VariationalEpoch.TDBJ2000S, value.VariationalEpoch.ElapsedTTS, value.RadialM, value.AlongM, value.CrossM}, nil)
	}
	other := orbitalResource[*OrbitParameterFit](t)(FitInertialOrbitParametersWithValidity(samples, "G03", nil, options, false))
	orbitalEqual(t, true)(fit.Equal(other))
	orbitalEqual(t, summary)(other.Summary())
	bounded := orbitalResource[*OrbitParameterFit](t)(FitInertialOrbitParameters(samples[:2], "G03", nil, options))
	orbitalEqual(t, OrbitParameterCovariance{})(bounded.Covariance())
	invalidOptions := options
	invalidOptions.GravityTideSystem = GravityTideSystem(3)
	_, err := FitInertialOrbitParameters(samples, "G03", nil, invalidOptions)
	var status *StatusError
	if !errors.As(err, &status) || status.Code != StatusInvalidArgument ||
		status.Detail != "sidereon_fit_inertial_orbit_parameters: invalid gravity tide-system selector" {
		t.Fatalf("invalid tide system error = %v", err)
	}
	const name = "empirical.along.constant"
	force := orbitalResource[*EmpiricalRTNAcceleration](t)(NewEmpiricalRTNAcceleration([]EmpiricalRTNTerm{
		{Axis: RTNAxisAlongTrack, Harmonic: EmpiricalHarmonicConstant, CoefficientKmS2: 1e-8},
	}))
	forceOptions := options
	forceOptions.AdditionalForces = []Force{force}
	forceOptions.EstimatedParameters = []string{name}
	forceFit := orbitalResource[*OrbitParameterFit](t)(FitInertialOrbitParameters(samples, "G03", nil, forceOptions))
	fitted := readOrbitalFixture(t)
	orbitalEqual(t, fitted.FittedForce)(forceFit.ForceParameter(name))
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: fitted.FittedForceSigma})(forceFit.Sigma(name))
	orbitalEqual(t, OptionalOrbitScalar{})(forceFit.Sigma("unknown"))
	orbitalEqual(t, OptionalOrbitScalar{})(bounded.Sigma("x_km"))
	orbitalEqual(t, false)(fit.Equal(forceFit))
	if values := orbitalResult[[]EstimatedForceParameter](t)(forceFit.ForceParameters()); !slices.Equal(values, []EstimatedForceParameter{fitted.FittedForce}) {
		t.Fatalf("estimated force parameters = %+v", values)
	}

	options.EstimatedParameters = []string{"x_km"}
	_, err = FitInertialOrbitParameters(samples, "G03", nil, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "invalid_parameter", `{"name":"x_km","reason":"is the name of an initial-state parameter"}`)
}

func orbitalClockEpoch(scale TimeScale, seconds uint64) ClockEpoch {
	return ClockEpoch{Scale: scale, Representation: RINEXClockInstantNanos, NanosLow: seconds * 1_000_000_000}
}

func TestOrbitalScoringFixture(t *testing.T) {
	want := readOrbitalFixture(t)
	epoch := orbitalClockEpoch(GPST, 60)
	node := TrajectorySample{Satellite: "G03", Epoch: epoch, PositionECEFM: [3]float64{10, 20, 30},
		ClockS: OptionalOrbitScalar{Present: true, Value: 0.125}}
	trajectory := orbitalResource[*OrbitTrajectory](t)(NewOrbitTrajectory("IGS20", epoch, []TrajectorySample{node}))
	orbitalEqual(t, OrbitSourceState{PositionECEFM: [3]float64{10, 20, 30}, ClockS: OptionalOrbitScalar{Present: true, Value: 0.125}})(
		trajectory.StateAtJ2000S("G03", 60))
	orbitalEqual(t, OptionalOrbitScalar{Present: true, Value: 60})(trajectory.PredictionOriginJ2000S("G03"))
	orbitalEqual(t, "IGS20")(trajectory.FrameLabel())
	if epochs := orbitalResult[[]ClockEpoch](t)(trajectory.Epochs()); !slices.Equal(epochs, []ClockEpoch{epoch}) {
		t.Fatalf("trajectory epochs = %+v", epochs)
	}
	if satellites := orbitalResult[[]string](t)(trajectory.Satellites()); !slices.Equal(satellites, []string{"G03"}) {
		t.Fatalf("trajectory satellites = %v", satellites)
	}
	marked := orbitalResource[*OrbitTrajectory](t)(trajectory.WithClockRelativityIncluded(true))
	orbitalEqual(t, true)(trajectory.Equal(trajectory))
	orbitalEqual(t, false)(trajectory.Equal(marked))
	info := orbitalResult[OrbitSourceInfo](t)(marked.Info())
	orbitalEqual(t, OrbitSourceInfo{TimeScale: GPST, Frame: OrbitFrameKindLabel, ReferencePoint: OrbitReferencePointCenterOfMass, ClockIncludesRelativity: true})(info, nil)
	_, err := NewOrbitTrajectory("IGS20", orbitalClockEpoch(UTC, 60), []TrajectorySample{node})
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "instant_time_system_mismatch",
		`{"expected":"UTC","field":"samples","found":"GPST","index":0}`)
	orbitalEqual(t, RTNComponents{RadialM: 2, TransverseM: 3, NormalM: 6})(
		ComputeRTNComponents([3]float64{2, 3, 6}, [3]float64{4, 0, 0}, [3]float64{0, 2, 0}))
	weights := SISREWeights{Radial: 1, TransverseNormalSquared: 1}
	orbitalEqual(t, 5.0)(weights.OrbitOnlyM(RTNComponents{RadialM: 3, TransverseM: 4}))
	orbitalEqual(t, 4.0)(weights.WithClockM(RTNComponents{RadialM: 3, TransverseM: 4}, 3))
	orbitalEqual(t, "MEO")(OrbitRegimeText(OrbitRegimeMEO))
	orbitalEqual(t, "reference product")(ScoredProductText(ScoredProductReference))
	product := orbitalResource[*SP3](t)(LoadSP3(readPositioningFixture(t, "trimmed.sp3")))
	source := orbitalResource[*PredictedSP3](t)(PredictedSP3WithCutoff(product, orbitalClockEpoch(GPST, 646266600)))
	reference := orbitalResource[*SP3](t)(source.Product())
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	edges := []float64{0, 1800, 7200}
	options := DefaultOrbitScoreOptions(edges)
	orbitalEqual(t, true)(options.Equal(OrbitScoreOptions{HorizonBinEdgesS: edges, VelocityHalfStepS: 0.5, Triad: TriadVelocityInertial,
		ClockDatum: ClockDatumEpochSystemMedian, Gaps: OrbitScoreGapPolicyCount}))
	options.SISREWeights.Fixed = &SISREWeights{Radial: 1, TransverseNormalSquared: 0.02}
	report := orbitalResource[*OrbitScoreReport](t)(ScoreOrbit(source, reference, []string{"G08"}, options))
	summary := orbitalResult[OrbitScoreSummary](t)(report.Summary())
	coverage := ScoreCoverage{Epochs: want.ScoreEpochs, Scored: want.ScoreScored, OutsideHorizonBins: 5}
	orbitalEqual(t, OrbitScoreSummary{Coverage: coverage, Frame: FrameAgreement{Kind: FrameAgreementKindSameLabel},
		SourceReferencePoint: OrbitReferencePointCenterOfMass, SatelliteCount: 1, SystemCount: 1})(summary, nil)
	zero := ErrorStats{Present: true}
	expectedBins := make([]HorizonBinStats, 2)
	for i := range expectedBins {
		expectedBins[i] = HorizonBinStats{HorizonStartS: edges[i], HorizonEndS: edges[i+1], Count: want.ScoreBins[i],
			Radial: zero, Transverse: zero, Normal: zero, Orbit3d: zero, SISREOrbit: zero}
	}
	for _, read := range []func(uint64) ([]HorizonBinStats, error){report.SatelliteBins, report.SystemBins} {
		bins := orbitalResult[[]HorizonBinStats](t)(read(0))
		if !slices.Equal(bins, expectedBins) {
			t.Fatalf("score bins = %+v, want %+v", bins, expectedBins)
		}
	}
	satellites := orbitalResult[[]SatelliteScore](t)(report.PerSatellite())
	if !slices.Equal(satellites, []SatelliteScore{{Satellite: "G08", PredictionOriginJ2000S: OptionalOrbitScalar{Present: true, Value: 646266600},
		BinCount: 2, HasSISREWeights: true, SISREWeights: SISREWeights{Radial: 1, TransverseNormalSquared: 0.02}, Coverage: coverage}}) {
		t.Fatalf("satellite scores = %+v", satellites)
	}
	systems := orbitalResult[[]SystemScore](t)(report.PerSystem())
	if !slices.Equal(systems, []SystemScore{{System: GNSSSystemGPS, Coverage: coverage, BinCount: 2}}) {
		t.Fatalf("system scores = %+v", systems)
	}
	if got := orbitalResult[[]float64](t)(report.HorizonBinEdgesS()); !slices.Equal(got, edges) {
		t.Fatalf("horizon edges = %v, want %v", got, edges)
	}
	orbitalEqual(t, true)(report.Equal(report))
	options.ClockDatum = ClockDatumUnadjusted
	unadjusted := orbitalResource[*OrbitScoreReport](t)(ScoreOrbit(source, reference, []string{"G08"}, options))
	orbitalEqual(t, false)(report.Equal(unadjusted))
	coverage.Clocked = want.ScoreScored
	orbitalEqual(t, coverage)(orbitalResult[OrbitScoreSummary](t)(unadjusted.Summary()).Coverage, nil)
	orbitalEqual(t, "GPS")(source.SP3TimeSystem())

}

func TestOrbitalNumericalPropagationFixture(t *testing.T) {
	want := readOrbitalFixture(t)
	config := orbitalResult[PropagationConfig](t)(DefaultPropagationConfig())
	config.PositionKm = [3]float64{7000, 0, 0}
	config.VelocityKmPerS = [3]float64{0, 7.5, 0}
	for _, test := range []struct {
		force PropagationForceModel
		end   [2][3]float64
	}{
		{PropagationForceModelTwoBody, want.TwoBody},
		{PropagationForceModelTwoBodyJ2, want.TwoBodyJ2},
	} {
		config.ForceModel = test.force
		for _, propagate := range []func(PropagationConfig, []float64) (*Ephemeris, error){
			PropagateState,
			func(config PropagationConfig, epochs []float64) (*Ephemeris, error) {
				return PropagateStateWithForceModel(config, epochs, PropagationEnvironment{}, nil)
			},
		} {
			result := orbitalResource[*Ephemeris](t)(propagate(config, []float64{0, 60, 120}))
			states := orbitalResult[[]CartesianState](t)(result.States())
			if len(states) != 3 {
				t.Fatalf("state count = %d, want 3", len(states))
			}
			orbitalEqual(t, CartesianState{EpochTDBSeconds: 120, PositionKm: test.end[0], VelocityKmPerS: test.end[1]})(states[2], nil)
		}
	}
}

func TestOrbitalAttitudeLaws(t *testing.T) {
	block := orbitalResult[GNSSBlock](t)(GNSSBlockFromIGSName(" block iir "))
	orbitalEqual(t, GNSSBlockGPSIIR)(block, nil)
	orbitalEqual(t, "BLOCK IIR-B")(block.IGSName())
	orbitalEqual(t, "BLOCK IIR-B")(block.String(), nil)
	orbitalEqual(t, GNSSSystemGPS)(block.System())
	orbitalEqual(t, int32(-1))(GNSSBlockGPSIIA.Compare(GNSSBlockGPSIIR))
	orbitalEqual(t, int32(-1))(BeiDouManufacturerCAST.Compare(BeiDouManufacturerSECM))
	orbitalEqual(t, int32(-1))(BeiDouOrbitClassGEO.Compare(BeiDouOrbitClassMEO))
	cast := BeiDouManufacturerCAST
	orbitalEqual(t, GNSSBlockBeiDou3MEOCAST)(GNSSBlockBeiDou3MEO(&cast))
	first := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawForBlock(block, 0))
	manual := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawGPSIIR(0.2 * math.Pi / 180))
	normal := orbitalResource[*GNSSAttitudeLaw](t)(GNSSAttitudeLawOrbitNormal())
	orbitalEqual(t, true)(first.Equal(manual))
	orbitalEqual(t, false)(first.Equal(normal))
}

func TestOrbitalEclipseComposition(t *testing.T) {
	satellite := [3]float64{7000, 1, 0}
	sun := [3]float64{150000000, 0, 0}
	moon := [3]float64{-100000, 200000, 300000}
	orbitalEqual(t, 0.0)(MoonShadowFraction(satellite, sun, moon))
	orbitalEqual(t, 0.0)(ShadowFractionWithMoon(satellite, sun, moon))
	orbitalEqual(t, 0.0)(ShadowFractionWithModelAndMoon(satellite, sun, moon, EarthShadowWGS84Oblate))
	orbitalEqual(t, ShadowGradient{})(ShadowFractionWithGradient(satellite, sun))
	orbitalEqual(t, ShadowGradient{})(ShadowFractionWithModelMoonAndGradient(satellite, sun, moon, EarthShadowWGS84Oblate))
	orbitalEqual(t, EclipseSunlit)(EclipseStatusWithMoon(satellite, sun, moon))
	orbitalEqual(t, EclipseSunlit)(EclipseStatusWithModelAndMoon(satellite, sun, moon, EarthShadowWGS84Oblate))
	orbitalEqual(t, 695700.0)(IAU2015SolarRadiusKm(), nil)
}

func orbitalFieldsEqual(t *testing.T, engine *EngineError, fields string) bool {
	t.Helper()
	decoder := json.NewDecoder(strings.NewReader(fields))
	decoder.UseNumber()
	var expected, actual map[string]any
	if err := decoder.Decode(&expected); err != nil {
		t.Fatal(err)
	}
	if err := engine.UnmarshalFields(&actual); err != nil {
		t.Fatal(err)
	}
	return reflect.DeepEqual(actual, expected)
}

func orbitalErrorFields(status *StatusError) string {
	if status != nil && status.Engine != nil {
		return string(status.Engine.Fields)
	}
	return ""
}

func TestOrbitalCovarianceFixture(t *testing.T) {
	want := readOrbitalFixture(t)
	config := orbitalResult[PropagationConfig](t)(DefaultPropagationConfig())
	config.PositionKm = [3]float64{7000, 0, 0}
	config.VelocityKmPerS = [3]float64{0, 7.5, 0}
	config.ForceModel = PropagationForceModelTwoBodyJ2
	covariance := orbitalResult[Covariance6](t)(CovarianceFromDiagonal([]float64{1, 2, 3, 0.01, 0.02, 0.03}))
	env := PropagationEnvironment{}
	state := CartesianState{EpochTDBSeconds: 120, PositionKm: want.CovarianceState[0], VelocityKmPerS: want.CovarianceState[1]}
	orbitalEqual(t, StateWithCovariance{State: state, Covariance: Covariance6{Values: want.Covariance}})(
		PropagateStateWithCovarianceWithContext(config, covariance, 120, env))
	orbitalEqual(t, want.StateTransition)(StateTransitionMatrixToWithContext(config, 120, env))
	ephemeris := orbitalResource[*CovarianceEphemeris](t)(PropagateCovarianceWithContext(config, covariance, []float64{120}, CovarianceInertial, CovarianceInertial, ProcessNoise{}, env))
	orbitalEqual(t, 1)(ephemeris.Count())
	orbitalEqual(t, Covariance6{Values: want.Covariance})(ephemeris.CovarianceAt(120))
}

func TestOrbitalSP3FitFixture(t *testing.T) {
	want := readOrbitalFixture(t)
	product := orbitalResource[*SP3](t)(LoadSP3(readPositioningFixture(t, "orbitfit-5epoch.sp3")))
	for _, test := range []struct {
		fit   func() (*OrbitFitReport, error)
		count int
	}{
		{func() (*OrbitFitReport, error) {
			return FitSP3ECEFPreciseOrbit(product, "G08", nil)
		}, 1},
		{func() (*OrbitFitReport, error) {
			return FitSP3ECEFPreciseOrbitWithContext(product, "G08", nil, PropagationEnvironment{})
		}, 1},
		{func() (*OrbitFitReport, error) {
			return FitSP3ECEFPreciseOrbitsWithContext(product, []string{"G08"}, nil, PropagationEnvironment{})
		}, 1},
		{func() (*OrbitFitReport, error) {
			return FitAllSP3ECEFPreciseOrbitsWithContext(product, nil, PropagationEnvironment{})
		}, 8},
	} {
		result := orbitalResource[*OrbitFitReport](t)(test.fit())
		solutions := orbitalResult[[]OrbitFitSolution](t)(result.Fits())
		if len(solutions) != test.count || solutions[0].Satellite != "G08" {
			t.Fatalf("solution count = %d, want %d with G08 first", len(solutions), test.count)
		}
		orbitalEqual(t, CartesianState{EpochTDBSeconds: want.SP3FitEpoch, PositionKm: want.SP3Fit[0], VelocityKmPerS: want.SP3Fit[1]})(solutions[0].InitialState, nil)
	}
	options := DefaultOrbitParameterFitOptions()
	options.ForceModel = PropagationForceModelTwoBody
	parameters := orbitalResource[*OrbitParameterFit](t)(FitSP3ECEFOrbitParameters(product, nil, "G08", nil, options))
	summary := orbitalResult[OrbitParameterFitSummary](t)(parameters.Summary())
	orbitalEqual(t, CartesianState{EpochTDBSeconds: want.SP3ParameterFitEpoch, PositionKm: want.SP3ParameterFit[0], VelocityKmPerS: want.SP3ParameterFit[1]})(summary.InitialState, nil)
	other := orbitalResource[*OrbitParameterFit](t)(FitSP3ECEFOrbitParametersWithValidity(product, nil, "G08", nil, options, false))
	orbitalEqual(t, true)(parameters.Equal(other))
	orbitalEqual(t, summary)(other.Summary())
}

func TestOrbitalAdditionalForcesFixture(t *testing.T) {
	pins := readOrbitalFixture(t).Contracts
	epoch := orbitalResult[float64](t)(CivilToJ2000Seconds(CivilDateTime{Year: 2003, Month: 10, Day: 29, Hour: 12}))
	config := orbitalResult[PropagationConfig](t)(DefaultPropagationConfig())
	config.EpochTDBSeconds, config.ForceModel = epoch, PropagationForceModelTwoBodyJ2
	config.PositionKm, config.VelocityKmPerS = [3]float64{6778.137, 20, 30}, [3]float64{0, 7.5, 1}
	config.HasDrag = true
	config.Drag = orbitalResult[DragParameters](t)(DragParametersFromAreaMass(2.2, 10, 1000, SpaceWeather{F107: 70, F107A: 70, Ap: 4}, 120))
	along := orbitalResource[*EmpiricalRTNAcceleration](t)(NewEmpiricalRTNAcceleration([]EmpiricalRTNTerm{
		{Axis: RTNAxisAlongTrack, Harmonic: EmpiricalHarmonicConstant, CoefficientKmS2: 1e-6},
	}))
	cross := orbitalResource[*EmpiricalRTNAcceleration](t)(NewEmpiricalRTNAcceleration([]EmpiricalRTNTerm{
		{Axis: RTNAxisCrossTrack, Harmonic: EmpiricalHarmonicConstant, CoefficientKmS2: 3e-7},
	}))
	forces := []Force{along, cross}
	table := orbitalResource[*SpaceWeatherTable](t)(ParseSpaceWeatherCSV(readPositioningFixture(t, "space_weather/SW-All-20260702-trim.csv")))
	epochs := []float64{epoch, epoch + 250, epoch + 640}
	for _, tc := range []struct {
		pin     string
		weather *SpaceWeatherTable
	}{{"FIXED_DRAG", nil}, {"SOURCED_DRAG", table}} {
		t.Run(tc.pin, func(t *testing.T) {
			env := PropagationEnvironment{SpaceWeather: tc.weather}
			want := pins[tc.pin]
			eph := orbitalResource[*Ephemeris](t)(PropagateStateWithForceModel(config, epochs, env, forces))
			if got := orbitalResult[[]float64](t)(eph.TimesS()); !slices.Equal(got, epochs) {
				t.Fatalf("ephemeris epochs = %v, want %v", got, epochs)
			}
			states := orbitalResult[[]CartesianState](t)(eph.States())
			if len(states) != len(epochs) {
				t.Fatalf("ephemeris state count = %d, want %d", len(states), len(epochs))
			}
			for i, state := range states {
				exactOrbitState(t, want[7*i:7*i+7], state)
			}
			final, dense, err := PropagateStateWithDenseOutput(config, epochs[2], env, forces)
			dense = orbitalResource[*DenseOutput](t)(dense, err)
			exactOrbitState(t, want[21:28], final)
			trajectory := orbitalResource[*OrbitAttitudeTrajectory](t)(OrbitAttitudeTrajectoryFromDenseOutput(dense))
			sample := orbitalResult[AttitudeSample](t)(trajectory.Sample(epoch + 320))
			orbitalEqual(t, [3]float64{want[29] * 1e3, want[30] * 1e3, want[31] * 1e3})(sample.PositionM, nil)
			orbitalEqual(t, [3]float64{want[32] * 1e3, want[33] * 1e3, want[34] * 1e3})(sample.VelocityMS, nil)
		})
	}
	config.HasDrag = false
	_, err := PropagateStateWithForceModel(config, epochs, PropagationEnvironment{SpaceWeather: table}, forces)
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilyPropagation, "invalid_input",
		`{"message":"space weather source without drag","reason":"space weather source without drag"}`)
	if err := cross.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = PropagateStateWithForceModel(config, epochs, PropagationEnvironment{}, forces)
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("closed force error = %v, want %v", err, ErrClosed)
	}
	_, dense, err := PropagateStateWithDenseOutput(config, epochs[2], PropagationEnvironment{}, forces)
	if dense != nil || !errors.Is(err, ErrClosed) {
		t.Fatalf("closed force produced dense output %p and error %v", dense, err)
	}
}
