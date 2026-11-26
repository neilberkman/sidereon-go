package sidereon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const icgemTestHeader = "product_type gravity_field\nmodelname M\nearth_gravity_constant 3.986004415e14\nradius 6378136.3\nmax_degree 2\nerrors formal\ntide_system tide_free\n"
const icgemTestCoefficients = "gfc 2 0 -4.8e-4 0.0 1e-11 0.0\ngfc 2 1 0.0 0.0 1e-11 1e-11\ngfc 2 2 2.4e-6 -1.4e-6 1e-11 1e-11\n"

func TestOrbitalICGEMParseErrors(t *testing.T) {
	base := icgemTestHeader + "end_of_head\n" + icgemTestCoefficients
	timeVariable := "format icgem2.0\n" + strings.Replace(icgemTestHeader, "errors formal", "errors no", 1) + "end_of_head\n" +
		"gfct 2 0 -4.8e-4 0.0 20000101.0000 20100101.0000\n" +
		"gfct 2 0 -4.8e-4 0.0 20100101.0000 20200101.0000\n" +
		"trnd 2 0 1e-12 0.0 20050101.0000 20150101.0000\n" +
		"gfc 2 1 0.0 0.0\ngfc 2 2 2.4e-6 -1.4e-6\n"
	cases := []struct {
		kind, text, fields string
	}{
		{"missing_field", strings.Replace(base, "modelname M", "modelname", 1), `{"line":2,"keyword":"modelname","field":"value"}`},
		{"invalid_field", strings.Replace(base, "errors formal", "errors some", 1), `{"line":6,"keyword":"errors","field":"value","value":"some"}`},
		{"missing_header_keyword", strings.Replace(base, "radius 6378136.3\n", "", 1), `{"keyword":"radius"}`},
		{"repeated_header_keyword", strings.Replace(base, "radius 6378136.3\n", "radius 6378136.3\nradius 6378137\n", 1), `{"line":5,"first_line":4,"keyword":"radius"}`},
		{"keyword_not_in_format", strings.Replace(timeVariable, "trnd 2 0", "dot 2 0", 1), `{"line":12,"keyword":"dot"}`},
		{"unknown_record_keyword", strings.Replace(base, "gfc 2 2", "gfx 2 2", 1), `{"line":11,"keyword":"gfx"}`},
		{"missing_record_keyword", strings.Replace(base, "gfc 2 2", "2 2", 1), `{"line":11}`},
		{"degree_above_maximum", base + "gfc 3 0 1.0e-6 0.0 0 0\n", `{"line":12,"degree":3,"max_degree":2}`},
		{"order_above_degree", base + "gfc 1 2 0.0 0.0 0 0\n", `{"line":12,"degree":1,"order":2}`},
		{"repeated_coefficient", base + "gfc 2 2 2.4e-6 -1.4e-6 1e-11 1e-11\n", `{"line":12,"first_line":11,"degree":2,"order":2}`},
		{"overlapping_term", strings.Replace(timeVariable, "20100101.0000 20200101.0000", "20090101.0000 20200101.0000", 1), `{"line":11,"first_line":10,"keyword":"gfct","degree":2,"order":0}`},
		{"uncovered_term", strings.Replace(timeVariable, "20100101.0000 20200101.0000", "20110101.0000 20200101.0000", 1), `{"line":12,"keyword":"trnd","degree":2,"order":0}`},
		{"missing_coefficient", strings.Replace(base, "gfc 2 1 0.0 0.0 1e-11 1e-11", "# no coefficient", 1) + "# comment\n", `{"degree":2,"order":1}`},
		{"too_few_records", strings.Replace(base, "max_degree 2", "max_degree 60000", 1), `{"max_degree":60000,"required":1800089998,"lines":3}`},
	}
	for _, test := range cases {
		t.Run(test.kind, func(t *testing.T) {
			_, err := ParseICGEMGravityField(test.text)
			requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, test.kind, test.fields)
		})
	}
	t.Run("normalization_out_of_range", func(t *testing.T) {
		var text strings.Builder
		text.WriteString("product_type gravity_field\nmodelname U\nearth_gravity_constant 3.986004415e14\n" +
			"radius 6378136.3\nmax_degree 170\nerrors no\nnorm unnormalized\nend_of_head\n")
		for degree := 2; degree <= 170; degree++ {
			for order := range degree + 1 {
				var c int
				if degree == 170 && order == 170 {
					c = 1
				}
				fmt.Fprintf(&text, "gfc %d %d %d 0\n", degree, order, c)
			}
		}
		_, err := ParseICGEMGravityField(text.String())
		requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "normalization_out_of_range", `{"line":14711,"degree":170,"order":170}`)
	})
}

func TestOrbitalICGEMEvaluationErrors(t *testing.T) {
	base := icgemTestHeader + "end_of_head\n" + icgemTestCoefficients
	field := orbitalResource[*ICGEMGravityField](t)(ParseICGEMGravityField(base))
	_, err := field.Gravity(3, 1)
	requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "truncation_out_of_range", `{"degree":3,"order":1,"max_degree":2}`)
	_, err = field.WithAssumedTideSystem(GravityTideSystemMeanTide)
	requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "tide_system_conflict", `{"stated":0,"assumed":2}`)
	unknown := strings.Replace(base, "tide_system tide_free\n", "", 1)
	withoutTide := orbitalResource[*ICGEMGravityField](t)(ParseICGEMGravityField(unknown))
	_, err = withoutTide.Gravity(2, 2)
	requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "unknown_tide_system", "{}")
	typed := orbitalResource[*ICGEMGravityField](t)(withoutTide.WithAssumedTideSystem(GravityTideSystemTideFree))
	matching := orbitalResource[*SphericalHarmonicGravity](t)(typed.GravityInTideSystem(2, 2, GravityTideSystemTideFree))
	orbitalEqual(t, OptionalHarmonicCoefficient{Present: true, Coefficient: SphericalHarmonicCoefficient{Degree: 2, Order: 0, C: -4.8e-4}})(matching.CoefficientAt(2, 0, 0))
	_, err = field.CoefficientAt(2, 0, math.NaN())
	requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "non_finite_epoch", "{}")
	dynamic := orbitalResource[*ICGEMGravityField](t)(ParseICGEMGravityField(string(readPositioningFixture(t, filepath.Join("orbital", "synthetic_icgem2.gfc")))))
	_, err = dynamic.CoefficientAt(2, 0, 1e12)
	requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "epoch_outside_validity",
		`{"degree":2,"order":0,"epoch_tt_seconds":{"bits_hex":"426d1a94a2000000","decimal":"1000000000000"}}`)
	nonfinite := "format icgem2.0\n" + strings.Replace(icgemTestHeader, "errors formal", "errors no", 1) + "end_of_head\n" +
		"gfct 2 0 -4.8e-4 0.0 20000101.0000 20300101.0000\n" +
		"acos 2 0 1e-11 0.0 20000101.0000 20300101.0000 1e-320\n" +
		"gfc 2 1 0.0 0.0\ngfc 2 2 2.4e-6 -1.4e-6\n"
	tinyPeriod := orbitalResource[*ICGEMGravityField](t)(ParseICGEMGravityField(nonfinite))
	_, err = tinyPeriod.CoefficientAt(2, 0, 1)
	requireOrbitalError[*ICGEMError](t, err, StatusInvalidArgument, "non_finite_coefficient",
		`{"degree":2,"order":0,"epoch_tt_seconds":{"bits_hex":"3ff0000000000000","decimal":"1"}}`)
}

func TestOrbitalDEFileErrors(t *testing.T) {
	_, err := DEEphemerisFromBytes([]byte{1, 2, 3})
	requireOrbitalError[*DEError](t, err, StatusSolve, "invalid_file", `{"reason":"incomplete first header record"}`)
	_, err = LoadDEEphemeris(filepath.Join(t.TempDir(), "absent.440"))
	message := "No such file or directory (os error 2)"
	if runtime.GOOS == "windows" {
		message = "The system cannot find the file specified. (os error 2)"
	}
	requireOrbitalError[*DEError](t, err, StatusSolve, "io", fmt.Sprintf(`{"message":%q,"raw_os_error":2}`, message))
	want := readOrbitalFixture(t).DE
	data := readPositioningFixture(t, filepath.Join("orbital", "lnxp1990.440"))
	for offset := 2*int(want.RecordSize) + 16; offset < 3*int(want.RecordSize); offset += 8 {
		binary.LittleEndian.PutUint64(data[offset:offset+8], math.Float64bits(math.NaN()))
	}
	value := orbitalResource[*DEEphemeris](t)(DEEphemerisFromBytes(data))
	_, err = value.SunMoonAtTDBSeconds(want.Start + 3600)
	requireOrbitalError[*DEError](t, err, StatusSolve, "non_finite_state", "{}")
}

func TestOrbitalShadowGradientSingularity(t *testing.T) {
	_, err := ShadowFractionWithGradient([3]float64{-2e6, 0, 0}, [3]float64{149597870.7, 0, 0})
	requireOrbitalError[*ShadowGradientError](t, err, StatusInvalidArgument, "numerical_failure", "{}")
}

func TestOrbitalAnalyticDomainErrors(t *testing.T) {
	gravity := ZonalHarmonicsEarth()
	gravity.ReKm = 6500
	for _, test := range []struct {
		mean   CircularElements
		kind   string
		fields string
	}{
		{CircularElements{AKm: 6400, IRad: 0.5}, "inside_reference_sphere", `{"semi_major_axis_km":{"bits_hex":"40b9000000000000","decimal":"6400"},"reference_radius_km":{"bits_hex":"40b9640000000000","decimal":"6500"}}`},
		{CircularElements{AKm: 7200}, "nearly_equatorial", `{"inclination_rad":{"bits_hex":"0000000000000000","decimal":"0"}}`},
		{CircularElements{AKm: 7200, IRad: 1.107421875}, "near_critical_inclination", `{"inclination_rad":{"bits_hex":"3ff1b80000000000","decimal":"1.107421875"}}`},
		{CircularElements{AKm: 7200, Ex: 0.5, IRad: 0.5}, "eccentricity_too_large", `{"eccentricity":{"bits_hex":"3fe0000000000000","decimal":"0.5"},"limit":{"bits_hex":"3fb999999999999a","decimal":"0.1"}}`},
	} {
		t.Run(test.kind, func(t *testing.T) {
			_, err := EcksteinHechlerPropagatorFromMeanElements(0, test.mean, gravity)
			requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, test.kind, test.fields)
		})
	}
	_, err := (CircularElements{AKm: 7200}).ToCartesian(0)
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "invalid_parameter", `{"field":"mu_km3_s2","reason":"must be positive"}`)
	_, err = CircularElementsFromCartesian(AnalyticCartesian{PositionKm: [3]float64{4, 0, 0}}, 4)
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "not_elliptic", `{"field":"state"}`)
	_, err = (CircularElements{AKm: math.NaN()}).ToCartesian(gravity.MuKm3S2)
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "non_finite", `{"field":"elements"}`)
	retrograde := AnalyticCartesian{PositionKm: [3]float64{7000, 0, 0}, VelocityKmS: [3]float64{0, -7.546, 0}}
	_, err = BrouwerLyddaneMeanElements(retrograde, gravity, DefaultMeanConversionOptionsBrouwerLyddane())
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "retrograde_equatorial",
		`{"inclination_rad":{"bits_hex":"400921fb54442d18","decimal":"3.141592653589793"}}`)
	state := AnalyticCartesian{PositionKm: orbitalPosition, VelocityKmS: orbitalVelocity}
	_, err = EcksteinHechlerMeanElements(state, gravity, MeanConversionOptions{Tolerance: 1e-11, MaxIterations: 1})
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "mean_elements_not_converged",
		`{"theory":"Eckstein-Hechler","iterations":1}`)
	maxEpoch := math.MaxFloat64
	eh := orbitalResource[*EcksteinHechlerPropagator](t)(EcksteinHechlerPropagatorFromMeanElements(-maxEpoch,
		CircularElements{AKm: 7200, IRad: 0.5}, gravity))
	_, err = eh.OsculatingElements(maxEpoch)
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "non_finite_result",
		`{"theory":"Eckstein-Hechler","field":"a_km"}`)
	bl := orbitalResource[*BrouwerLyddanePropagator](t)(BrouwerLyddanePropagatorFromMeanElements(0,
		AnalyticKeplerianElements{AKm: 7200, E: 0.01, IRad: 0.5}, gravity, -1))
	_, err = bl.OsculatingElements(maxEpoch)
	requireOrbitalError[*AnalyticPropagationError](t, err, StatusInvalidArgument, "kepler_not_converged",
		`{"field":"eccentric anomaly"}`)
}

func TestOrbitalParameterFitRefusals(t *testing.T) {
	_, err := FitInertialOrbitParameters(nil, "G03", nil, DefaultOrbitParameterFitOptions())
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "fit",
		`{"cause":{"kind":"too_few_samples","fields":{"satellite":"G03","got":0,"required":2}}}`)
	options := DefaultOrbitParameterFitOptions()
	options.ForceModel = PropagationForceModelTwoBody
	want := readOrbitalFixture(t).Fit
	seed := CartesianState{EpochTDBSeconds: want.InitialEpoch, PositionKm: orbitalPosition, VelocityKmPerS: orbitalVelocity}
	samples := make([]InertialPositionSample, 0, len(want.Observations))
	for i, position := range want.Observations {
		samples = append(samples, InertialPositionSample{
			Epoch: orbitalClockEpoch(TDB, uint64(i)*60), PositionGCRFM: position,
		})
	}
	overflow := append([]InertialPositionSample(nil), samples...)
	overflow[2].PositionGCRFM[0] = 1e200
	_, err = FitInertialOrbitParameters(overflow, "G03", &seed, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "numerical_failure",
		`{"satellite":"G03","quantity":"sum of squared residuals"}`)
	illConditioned := options
	illConditioned.MaxConditionNumber = 1.5
	_, err = FitInertialOrbitParameters(samples, "G03", nil, illConditioned)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "ill_conditioned_normal_equations", string(want.IllConditioned))
	nonConvergent := options
	nonConvergent.MaxIterations = 1
	nonConvergent.ConvergenceThreshold = 1e-12
	displaced := seed
	displaced.PositionKm[0] += 0.01
	_, err = FitInertialOrbitParameters(samples, "G03", &displaced, nonConvergent)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "did_not_converge", string(want.NonConvergent))

	empirical := orbitalResource[*EmpiricalRTNAcceleration](t)(NewEmpiricalRTNAcceleration([]EmpiricalRTNTerm{
		{Axis: RTNAxisAlongTrack, Harmonic: EmpiricalHarmonicConstant, CoefficientKmS2: 1e-8},
	}))
	options.AdditionalForces = []Force{empirical}
	options.EstimatedParameters = []string{"empirical.along.constant"}
	_, err = FitInertialOrbitParameters(samples[:2], "G03", &seed, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "singular_normal_equations",
		`{"satellite":"G03","condition_number":{"bits_hex":"7ff0000000000000","decimal":"inf"},"weakest":[]}`)

	tiny := orbitalResource[*SolarRadiationPressureForce](t)(NewSolarRadiationPressureForce(1, math.SmallestNonzeroFloat64, DefaultRadiationShadowOptions()))
	options.AdditionalForces = []Force{tiny}
	options.EstimatedParameters = []string{"srp.cr"}
	_, err = FitInertialOrbitParameters(samples, "G03", &seed, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "unobservable_parameter",
		`{"satellite":"G03","name":"srp.cr"}`)
}

func TestOrbitalScoreRefusals(t *testing.T) {
	reference := orbitalResource[*SP3](t)(LoadSP3(readPositioningFixture(t, "trimmed.sp3")))
	epoch := orbitalClockEpoch(GPST, 646266600)
	prediction := orbitalResource[*PredictedSP3](t)(PredictedSP3WithCutoff(reference, epoch))
	opts := DefaultOrbitScoreOptions([]float64{0})
	_, err := ScoreOrbit(prediction, reference, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "invalid_input",
		`{"field":"horizon_bin_edges_s","reason":"fewer than two edges"}`)
	opts = DefaultOrbitScoreOptions([]float64{0, 1800, 7200})
	opts.Gaps = OrbitScoreGapPolicyRefuse
	_, err = ScoreOrbit(prediction, reference, []string{"G03"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "missing_satellite", `{"satellite":"G03","product":1}`)
	utc := orbitalClockEpoch(UTC, 646266600)
	_, err = PredictedSP3WithCutoff(reference, utc)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "instant_time_system_mismatch",
		`{"field":"cutoff","index":0,"found":"UTC","expected":"GPST"}`)
	opts.HorizonBinEdgesS = []float64{0, 200000000}
	opts.Epochs.Instants = []ClockEpoch{orbitalClockEpoch(GPST, 746266600)}
	_, err = ScoreOrbit(prediction, reference, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "missing_epoch",
		`{"satellite":"G08","product":1,"epoch_index":0,"cause":{"kind":"epoch_out_of_range","fields":{}}}`)
	opts.Epochs.Instants = nil
	opts.HorizonBinEdgesS = []float64{0, 1800, 7200}
	otherTime := orbitalResource[*OrbitTrajectory](t)(NewOrbitTrajectory("IGb14", utc, []TrajectorySample{
		{Satellite: "G08", Epoch: utc, PositionECEFM: [3]float64{2e7, 3e7, 4e7}},
	}))
	_, err = ScoreOrbit(otherTime, reference, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "time_system_mismatch", `{"orbit_source":"UTC","reference":"GPST"}`)
	otherFrame := orbitalResource[*OrbitTrajectory](t)(NewOrbitTrajectory("IGS20", epoch, []TrajectorySample{
		{Satellite: "G08", Epoch: epoch, PositionECEFM: [3]float64{2e7, 3e7, 4e7}},
	}))
	_, err = ScoreOrbit(otherFrame, reference, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "frame_mismatch", `{"orbit_source":"IGS20","reference":"IGb14"}`)

	sp3 := string(readPositioningFixture(t, "trimmed.sp3"))
	irn := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.Replace(sp3, "cc GPS ccc", "cc IRN ccc", 1))))
	_, err = ScoreOrbit(prediction, irn, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "time_system_label_mismatch",
		`{"orbit_source":"GPS","reference":"IRN"}`)
	rows := strings.Split(sp3, "\n")
	for i, row := range rows {
		if strings.HasPrefix(row, "PG08") {
			rows[i] = row[:60] + strings.Repeat(" ", 19) + "P"
		}
	}
	allPredicted := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.Join(rows, "\n"))))
	withoutOrigin := orbitalResource[*PredictedSP3](t)(PredictedSP3FromPredictionFlags(allPredicted))
	_, err = ScoreOrbit(withoutOrigin, reference, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "no_prediction_origin", `{"satellite":"G08"}`)
}

func TestOrbitalScoreStateRefusals(t *testing.T) {
	sp3 := string(readPositioningFixture(t, "trimmed.sp3"))
	reference := orbitalResource[*SP3](t)(LoadSP3([]byte(sp3)))
	origin := orbitalClockEpoch(GPST, 646266600)
	epoch := orbitalClockEpoch(GPST, 646267500)
	node := TrajectorySample{Satellite: "G08", Epoch: epoch, PositionECEFM: [3]float64{2e7, 3e7, 4e7}}
	opts := DefaultOrbitScoreOptions([]float64{0, 1800})
	opts.Epochs.Instants = []ClockEpoch{epoch}
	opts.Gaps = OrbitScoreGapPolicyRefuse
	node.ClockS = OptionalOrbitScalar{Present: true, Value: 1e308}
	source := orbitalResource[*OrbitTrajectory](t)(NewOrbitTrajectory("IGb14", origin, []TrajectorySample{node}))
	_, err := ScoreOrbit(source, reference, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "non_finite_result", `{"satellite":"G08"}`)

	rows := strings.Split(sp3, "\n")
	for i, row := range rows {
		if strings.HasPrefix(row, "PG08") {
			rows[i] = fmt.Sprintf("PG08%14.6f%14.6f%14.6f%s", 100000.0, 0.0, 0.0, row[46:])
		}
	}
	static := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.Join(rows, "\n"))))
	node.ClockS = OptionalOrbitScalar{}
	source = orbitalResource[*OrbitTrajectory](t)(NewOrbitTrajectory("IGb14", origin, []TrajectorySample{node}))
	_, err = ScoreOrbit(source, static, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "undetermined_orbit_regime", `{"satellite":"G08"}`)
	for i, row := range rows {
		if strings.HasPrefix(row, "PG08") {
			rows[i] = fmt.Sprintf("PG08%14.6f%14.6f%14.6f%s", 0.0, 0.0, 100000.0, row[46:])
		}
	}
	polar := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.Join(rows, "\n"))))
	_, err = ScoreOrbit(source, polar, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "degenerate_reference_state",
		`{"satellite":"G08","epoch_index":0}`)

	qzss := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.ReplaceAll(sp3, "G08", "J01"))))
	qzssSource := orbitalResource[*PredictedSP3](t)(PredictedSP3WithCutoff(qzss, origin))
	opts.Triad = TriadVelocityInertial
	_, err = ScoreOrbit(qzssSource, qzss, []string{"J01"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "no_sisre_weights",
		`{"satellite":"J01","regime":"MEO"}`)

	itrf2020 := orbitalResource[*SP3](t)(LoadSP3([]byte(strings.Replace(sp3, "IGb14", "IGS20", 1))))
	node.PositionECEFM = [3]float64{math.MaxFloat64, 0, 0}
	source = orbitalResource[*OrbitTrajectory](t)(NewOrbitTrajectory("IGb14", origin, []TrajectorySample{node}))
	opts.FrameReconciliation.Helmert = true
	_, err = ScoreOrbit(source, itrf2020, []string{"G08"}, opts)
	requireOrbitalError[*OrbitScoreError](t, err, StatusInvalidArgument, "frame_transform",
		`{"cause":{"kind":"invalid_input","fields":{"field":"x_m","reason":"must be finite"}}}`)
}

func requireOrbitalFamily(t *testing.T, err error, code StatusCode, family EngineErrorFamily, kind, fields string) {
	t.Helper()
	requireOrbitalError[*EngineError](t, err, code, kind, fields)
	var engine *EngineError
	if !errors.As(err, &engine) || engine.Family != family {
		t.Fatalf("engine error = %+v, want family %v", engine, family)
	}
}

func TestOrbitalEnvironmentRefusals(t *testing.T) {
	_, err := GCRSToMeanOfDateMatrix(TimeScales{JDTDB: math.NaN()})
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilyFrameTransform, "invalid_input",
		`{"field":"jd_tdb","reason":"must be finite"}`)
	_, err = SunMoonGCRS(math.NaN())
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilySunMoon, "invalid_input",
		`{"field":"t","reason":"not finite"}`)
	_, err = SunMoonGCRSAt(time.UnixMilli(math.MaxInt64 / 1000))
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilySunMoon, "frame_transform",
		`{"cause":{"kind":"invalid_input","fields":{"field":"datetime","reason":"invalid UTC instant"}}}`)
	position, velocity, sun := [3]float64{7000, 0, 0}, [3]float64{0, 7.5, 0}, [3]float64{150000000, 0, 0}
	_, err = ShadowFractionWithMoon(position, sun, [3]float64{})
	requireOrbitalError[*EclipseError](t, err, StatusInvalidArgument, "invalid_input", `{"field":"moon_pos","reason":"zero vector"}`)
	_, err = ComputeECOMGeometry(position, velocity, sun)
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyPropagation, "numerical_failure",
		`{"message":"ECOM DYB frame is undefined or overflows","reason":"ECOM DYB frame is undefined or overflows"}`)
	srp := orbitalResource[*SolarRadiationPressureForce](t)(NewSolarRadiationPressureForce(1.25, 0.02, DefaultRadiationShadowOptions()))
	_, err = srp.AccelerationWithSunMoon(position, sun, [3]float64{})
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyPropagation, "force_model_failure",
		`{"message":"eclipse: invalid eclipse input moon_pos: zero vector","reason":"eclipse: invalid eclipse input moon_pos: zero vector"}`)

	config := orbitalResult[PropagationConfig](t)(DefaultPropagationConfig())
	config.EpochTDBSeconds, config.PositionKm, config.VelocityKmPerS = orbitalEpoch, orbitalPosition, orbitalVelocity
	config.HasDrag = true
	orientation := orbitalResource[*IERS2010EarthOrientationProvider](t)(NewIERS2010EarthOrientationProvider([]IERS2010EopSample{{MJDUTC: 59020}, {MJDUTC: 59021}}))
	_, _, err = PropagateStateWithDenseOutput(config, orbitalEpoch+600, PropagationEnvironment{Context: &PropagationContext{EarthOrientation: orientation}}, nil)
	requireOrbitalFamily(t, err, StatusUT1OutsideCoverage, EngineErrorFamilyPropagation, "ut1_outside_coverage", `{"reason":"after_coverage"}`)
	_, err = orientation.OrientationAtTDBSeconds(orbitalEpoch)
	requireOrbitalFamily(t, err, StatusUT1OutsideCoverage, EngineErrorFamilyFrameTransform, "ut1_outside_coverage", `{"reason":"after_coverage"}`)
	config.HasDrag, config.MaxSteps = false, 1
	_, _, err = PropagateStateWithDenseOutput(config, orbitalEpoch+600, PropagationEnvironment{}, nil)
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyPropagation, "max_steps_exceeded", "{}")
}

func TestOrbitalDecayRefusals(t *testing.T) {
	config := orbitalResult[DecayConfig](t)(DefaultDecayConfig())
	initial := CartesianState{EpochTDBSeconds: orbitalEpoch + 600, PositionKm: [3]float64{7000, 0, 0}, VelocityKmPerS: [3]float64{0, 7.5, 0}}
	config.ScanStepS = 0
	_, err := EstimateDecayWithContext(initial, config, nil)
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilyDecay, "invalid_config",
		`{"field":"scan_step_s","reason":"scan_step_s"}`)
	config.ScanStepS, config.MaxDurationS = 1, 1
	_, err = EstimateDecayWithContext(initial, config, nil)
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyDecay, "no_decay_within_horizon",
		`{"horizon_s":{"decimal":"1","bits_hex":"3ff0000000000000"}}`)
	config.MaxDurationS, config.MaxScanSamples = 5, 1
	_, err = EstimateDecayWithContext(initial, config, nil)
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyDecay, "scan_budget_exhausted",
		`{"scanned_s":{"decimal":"1","bits_hex":"3ff0000000000000"},"samples":1}`)
	initial.PositionKm[0] = math.NaN()
	_, err = EstimateDecayWithContext(initial, config, nil)
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyDecay, "propagation",
		`{"cause":{"kind":"invalid_input","fields":{"message":"position_km not finite","reason":"position_km not finite"}}}`)
}

func TestOrbitalParameterFitNestedRefusals(t *testing.T) {
	options := DefaultOrbitParameterFitOptions()
	options.ForceModel = PropagationForceModelTwoBody
	first := InertialPositionSample{Epoch: orbitalClockEpoch(TDB, 0), PositionGCRFM: [3]float64{7000000, 0, 0}}
	second := InertialPositionSample{Epoch: orbitalClockEpoch(TDB, 60), PositionGCRFM: [3]float64{7000000, 0, 0}}
	samples := []InertialPositionSample{first, second}
	invalidSigma := options
	invalidSigma.ObservationSigmaM = 0
	_, err := FitInertialOrbitParameters(samples, "G03", nil, invalidSigma)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "fit",
		`{"cause":{"kind":"invalid_option","fields":{"field":"observation_sigma_m","reason":"not finite and positive"}}}`)
	_, err = FitInertialOrbitParameters([]InertialPositionSample{first, first}, "G03", nil, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "fit",
		`{"cause":{"kind":"non_monotonic_epochs","fields":{"satellite":"G03"}}}`)
	seed := CartesianState{EpochTDBSeconds: 1, PositionKm: [3]float64{7000, 0, 0}, VelocityKmPerS: [3]float64{0, 7.5, 0}}
	_, err = FitInertialOrbitParameters(samples, "G03", &seed, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "fit",
		`{"cause":{"kind":"invalid_epoch","fields":{"satellite":"G03","reason":"initial-state seed epoch must match the first sample"}}}`)
	second.PositionGCRFM[0] = math.NaN()
	_, err = FitInertialOrbitParameters([]InertialPositionSample{first, second}, "G03", nil, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "fit",
		`{"cause":{"kind":"invalid_observation","fields":{"satellite":"G03","reason":"position components must be finite"}}}`)
	seed.EpochTDBSeconds = readOrbitalFixture(t).Fit.InitialEpoch
	stepLimit := options
	stepLimit.IntegratorOptions.MaxSteps = 0
	_, err = FitInertialOrbitParameters(samples, "G03", &seed, stepLimit)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "fit",
		`{"cause":{"kind":"propagation","fields":{"satellite":"G03","cause":{"kind":"max_steps_exceeded","fields":{}}}}}`)
	options.ForceModel = PropagationForceModelComposite
	options.ForceComponents.HasTwoBody = false
	seed.VelocityKmPerS = [3]float64{}
	_, err = FitInertialOrbitParameters(samples, "G03", &seed, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "fit",
		`{"cause":{"kind":"rtn_frame","fields":{"satellite":"G03","cause":{"kind":"parallel_position_velocity","fields":{}}}}}`)
	seed.PositionKm = [3]float64{}
	samples[0].PositionGCRFM, samples[1].PositionGCRFM = [3]float64{}, [3]float64{}
	_, err = FitInertialOrbitParameters(samples, "G03", &seed, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusSolve, "fit",
		`{"cause":{"kind":"rtn_frame","fields":{"satellite":"G03","cause":{"kind":"zero_position","fields":{}}}}}`)
}

func TestOrbitalParameterFitOrientationRefusals(t *testing.T) {
	options := DefaultOrbitParameterFitOptions()
	options.ForceModel = PropagationForceModelTwoBody
	product := orbitalResource[*SP3](t)(LoadSP3(readPositioningFixture(t, "rtk/GBM0MGXRAP_20201770000_01D_05M_ORB_120epoch.sp3")))
	orientation := orbitalResource[*IERS2010EarthOrientationProvider](t)(NewIERS2010EarthOrientationProvider([]IERS2010EopSample{{MJDUTC: 59020}, {MJDUTC: 59021}}))
	_, err := FitSP3ECEFOrbitParameters(product, orientation, "G03", nil, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "fit",
		`{"cause":{"kind":"frame","fields":{"satellite":"G03","cause":{"kind":"ut1_outside_coverage","fields":{"reason":"after_coverage"}}}}}`)
	permissive := orbitalResource[*IERS2010EarthOrientationProvider](t)(orientation.WithValidity(true))
	_, err = FitSP3ECEFOrbitParameters(product, permissive, "G03", nil, options)
	requireOrbitalError[*OrbitParameterFitError](t, err, StatusInvalidArgument, "fit",
		`{"cause":{"kind":"ut1_validity_mismatch","fields":{"fit":"strict","provider":"orientation","provider_mode":"permissive"}}}`)
}

func assertOrbitalErrorText[T error](t *testing.T, err error, want string) {
	t.Helper()
	var typed T
	if !errors.As(err, &typed) {
		t.Fatalf("error = %v, want %T", err, typed)
	}
	orbitalEqual(t, want)(typed.Error(), nil)
}

func TestOrbitalErrorText(t *testing.T) {
	_, err := ComputeDYBAxes([3]float64{}, orbitalSun)
	assertOrbitalErrorText[*GNSSAttitudeError](t, err, "position_m has zero length")
	_, err = GNSSBlockFromIGSName("X")
	assertOrbitalErrorText[*GNSSBlockError](t, err, `unknown GNSS block name "X"`)
	_, err = DEEphemerisFromBytes([]byte("short"))
	assertOrbitalErrorText[*DEError](t, err, "Invalid DE ephemeris: incomplete first header record")
	_, err = ShadowFractionWithGradient([3]float64{-2e6, 0, 0}, [3]float64{149597870.7, 0, 0})
	assertOrbitalErrorText[*ShadowGradientError](t, err, "shadow fraction gradient is not finite")
	_, err = ParseICGEMGravityField(strings.Replace(icgemTestHeader, "modelname M", "modelname", 1) + "end_of_head\n" + icgemTestCoefficients)
	assertOrbitalErrorText[*ICGEMError](t, err, "ICGEM line 2: modelname record has no value")
	_, err = (CircularElements{AKm: 7200}).ToCartesian(0)
	assertOrbitalErrorText[*AnalyticPropagationError](t, err, "invalid mu_km3_s2: must be positive")
	_, err = FitInertialOrbitParameters(nil, "G03", nil, DefaultOrbitParameterFitOptions())
	assertOrbitalErrorText[*OrbitParameterFitError](t, err, "satellite G03 has 0 samples; need at least 2")
	item := ecksteinHechlerAttitude(t)
	err = item.WithVelocityHalfStepS(1e-9)
	assertOrbitalErrorText[*OrbitScoreError](t, err, "invalid orbit scoring velocity_half_step_s: below the 1e-8 s resolution of SP3 epochs")
}

func TestOrbitalFrameCatalogSingularity(t *testing.T) {
	// On reversing the 2020-to-2008 transform, this epoch cancels its unit
	// scale with the catalog's scale rate of 0.03 ppb per year.
	const year = 2015 + (-1e9+0.29)/0.03
	_, err := TransformTerrestrialState(TerrestrialState{}, TerrestrialFrameITRF2008, TerrestrialFrameITRF2020, year)
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyFrameCatalog, "singular_transform",
		`{"from":"itrf2020","to":"itrf2008","epoch_year":{"decimal":"-33333331308.666668","bits_hex":"c21f0b4915b2aaab"}}`)
}

func TestOrbitalForceInputRefusals(t *testing.T) {
	earth := ZonalHarmonicsEarth()
	j2 := orbitalResource[*J2Gravity](t)(NewJ2Gravity(earth.MuKm3S2, earth.ReKm, -earth.C20))
	_, err := j2.Acceleration(CartesianState{PositionKm: [3]float64{math.NaN(), 1, 1}}, nil)
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilyPropagation, "invalid_input",
		`{"message":"position must be finite","reason":"position must be finite"}`)
	_, err = j2.Acceleration(CartesianState{PositionKm: [3]float64{1e-100, 0, 0}}, nil)
	requireOrbitalFamily(t, err, StatusSolve, EngineErrorFamilyPropagation, "numerical_failure",
		`{"message":"acceleration not finite","reason":"acceleration not finite"}`)

	srp := orbitalResource[*SolarRadiationPressureForce](t)(NewSolarRadiationPressureForce(1.25, 0.02, DefaultRadiationShadowOptions()))
	_, err = srp.Acceleration(CartesianState{EpochTDBSeconds: math.NaN(), PositionKm: orbitalPosition}, nil)
	requireOrbitalFamily(t, err, StatusInvalidArgument, EngineErrorFamilyPropagation, "invalid_input",
		`{"message":"epoch must be finite","reason":"epoch must be finite"}`)
}
