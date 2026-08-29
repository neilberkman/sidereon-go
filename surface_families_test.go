package sidereon

import (
	"errors"
	"math"
	"testing"
)

func TestDeterministicTRLSAndILS(t *testing.T) {
	problem, err := DefaultTRLSProblem(TRLSLinear)
	if err != nil {
		t.Fatal(err)
	}
	problem.A = []float64{1, 0, 1, 1, 1, 2}
	problem.B = []float64{1, 3, 5}
	problem.X0 = []float64{0, 0}
	problem.M, problem.N = 3, 2
	solution, err := SolveTRLS(problem)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = solution.Close() })
	summary, err := solution.Summary()
	if err != nil || !summary.Success || summary.N != 2 || summary.M != 3 || math.Abs(summary.Cost) > 1e-12 {
		t.Fatalf("TRLS summary = %+v, %v", summary, err)
	}
	x, err := solution.X()
	if err != nil || len(x) != 2 || math.Abs(x[0]-1) > 1e-12 || math.Abs(x[1]-2) > 1e-12 {
		t.Fatalf("TRLS x = %#v, %v", x, err)
	}
	for _, read := range []func() ([]float64, error){solution.Residuals, solution.Jacobian, solution.Gradient} {
		if values, readErr := read(); readErr != nil || len(values) == 0 {
			t.Fatalf("TRLS diagnostic = %#v, %v", values, readErr)
		}
	}
	drop, err := SolveTRLSDropOne(problem)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = drop.Close() })
	count, err := drop.Count()
	if err != nil || count != 3 {
		t.Fatalf("TRLS drop count = %d, %v", count, err)
	}

	a := []float64{1585184.171, -6716599.430, 3915742.905, 7627233.455, 9565990.879, 989457273.200}
	q := []float64{
		0.227134, 0.112202, 0.112202, 0.112202, 0.112202, 0.103473,
		0.112202, 0.227134, 0.112202, 0.112202, 0.112202, 0.103473,
		0.112202, 0.112202, 0.227134, 0.112202, 0.112202, 0.103473,
		0.112202, 0.112202, 0.112202, 0.227134, 0.112202, 0.103473,
		0.112202, 0.112202, 0.112202, 0.112202, 0.227134, 0.103473,
		0.103473, 0.103473, 0.103473, 0.103473, 0.103473, 0.434339,
	}
	result, err := LambdaILS(a, q, 3)
	if err != nil || len(result.Fixed) != 6 || result.Fixed[0] != 1585184 || result.Fixed[1] != -6716599 || result.Fixed[2] != 3915743 || result.Fixed[5] != 989457273 {
		t.Fatalf("LAMBDA result = %+v, %v", result, err)
	}
	if result.FixedStatus || !result.SecondBestPresent || math.Abs(result.BestScore-3.5079844392) > 1e-4 || math.Abs(result.SecondBestScore-3.70845619249) > 1e-4 {
		t.Fatalf("LAMBDA diagnostics = %+v", result)
	}
	bounded, err := BoundedILS([]float64{0.1, 0.9}, []float64{1, 0, 0, 1}, BoundedILSOptions{Radius: 1, CandidateLimit: 200000}, 3)
	if err != nil || len(bounded.Fixed) != 2 || bounded.Fixed[0] != 0 || bounded.Fixed[1] != 1 {
		t.Fatalf("bounded ILS = %+v, %v", bounded, err)
	}
	if _, err := LambdaILS(a, q[:35], 3); err == nil {
		t.Fatal("LAMBDA accepted a covariance shape mismatch")
	}
	if _, err := drop.X(-1); err == nil {
		t.Fatal("TRLS accepted a negative drop index")
	}
}

func TestDeterministicFDEOptionsDefault(t *testing.T) {
	options, err := DefaultFDEOptions()
	if err != nil || options.PFA <= 0 || options.PFA >= 1 || !options.UnitWeights {
		t.Fatalf("FDE defaults = %+v, %v", options, err)
	}
}

func TestDeterministicTrackSurface(t *testing.T) {
	filter, err := NewTrackFilterFromPositionVelocity(TrackCallerDefinedCartesian, 0, []float64{0}, []float64{1}, []float64{1, 0, 0, 1}, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filter.Close() })
	history, err := NewTrackRTSHistoryBuilderFromFilter(filter)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = history.Close() })
	prediction, err := filter.PredictRecorded(1, history)
	if err != nil || prediction.Predicted.Dimension != 1 || prediction.Predicted.StateDimension != 2 {
		t.Fatalf("track prediction = %+v, %v", prediction, err)
	}
	gated, err := filter.UpdatePositionGatedRecorded([]float64{100}, []float64{0.01}, 0.95, history)
	if err != nil || gated.HasUpdate || gated.Gate.InGate {
		t.Fatalf("track gate = %+v, %v", gated, err)
	}
	position, err := filter.Position()
	stateVector, stateErr := filter.StateVector()
	if err != nil || stateErr != nil || len(position) != 1 || len(stateVector) != 2 || position[0] != 1 || stateVector[0] != 1 || stateVector[1] != 1 {
		t.Fatalf("track state = position %#v, vector %#v, errors %v/%v", position, stateVector, err, stateErr)
	}
	recorded, err := history.Finish()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recorded.Close() })
	smoothed, err := SmoothTrackRTS(recorded)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = smoothed.Close() })
	count, err := smoothed.EpochCount()
	if err != nil || count == 0 {
		t.Fatalf("smoothed count = %d, %v", count, err)
	}
	last, err := smoothed.Position(count - 1)
	if err != nil || len(last) != 1 {
		t.Fatalf("smoothed position = %#v, %v", last, err)
	}
}

func TestDeterministicFusionAndCovarianceSurface(t *testing.T) {
	config, err := DefaultFusionFilterConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.TimeSyncIMUCapacity = 8
	config.TimeSyncCheckpointCapacity = 4
	initial := FusionNavState{PositionECEFM: [3]float64{6378137, 0, 0}, AttitudeBodyToECEF: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}}
	diagonal := make([]float64, 15)
	for i := range diagonal {
		diagonal[i] = 10
	}
	filter, err := NewFusionFilter(initial, diagonal, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filter.Close() })
	if err := filter.Propagate(FusionIMUSample{EpochJ2000S: 1, Kind: FusionIMUIncrement, DTS: 1, DeltaVelocityMPerS: [3]float64{9.7803253359, 0, 0}}); err != nil {
		t.Fatal(err)
	}
	state, err := filter.State()
	if err != nil || state.EpochJ2000S != 1 || state.CovarianceDimension != 15 {
		t.Fatalf("fusion state = %+v, %v", state, err)
	}
	covariance, err := filter.Covariance()
	if err != nil || len(covariance) != 225 {
		t.Fatalf("fusion covariance length = %d, %v", len(covariance), err)
	}
	encoded, err := filter.Encode()
	if err != nil || len(encoded) == 0 {
		t.Fatalf("fusion encode length = %d, %v", len(encoded), err)
	}
	if err := filter.Restore(encoded); err != nil {
		t.Fatal(err)
	}
	if dimension, err := FusionLayoutDimension(FusionFifteenState); err != nil || dimension != 15 {
		t.Fatalf("fusion dimension = %d, %v", dimension, err)
	}
	if _, _, _, err := FusionLabels(FusionEKF, FusionFifteenState, FusionFixSingle); err != nil {
		t.Fatal(err)
	}

	base := Covariance6{}
	base.Values[0][0], base.Values[1][1], base.Values[2][2] = 1, 2, 3
	base.Values[3][3], base.Values[4][4], base.Values[5][5] = 0.01, 0.02, 0.03
	segment := CovarianceTransportSegment{DTSeconds: 10, QRotationState: CartesianState{PositionKm: [3]float64{7000, 0, 0}, VelocityKmPerS: [3]float64{0, 7.5, 0}}}
	for i := range segment.STM {
		segment.STM[i][i] = 1
	}
	transported, err := CovarianceTransport(base, []CovarianceTransportSegment{segment}, ProcessNoise{Kind: ProcessNoiseRTNAccelerationPSD, RadialKm2S3: 1e-6, TransverseKm2S3: 2e-6, NormalKm2S3: 3e-6})
	if err != nil || len(transported) != 2 || math.Abs(transported[1][0][0]-1.0003333333333333) > 1e-15 || math.Abs(transported[1][0][3]-5e-5) > 1e-18 {
		t.Fatalf("covariance transport = %#v, %v", transported, err)
	}
	propagation, err := DefaultPropagationConfig()
	if err != nil {
		t.Fatal(err)
	}
	propagation.EpochTDBSeconds = 0
	propagation.PositionKm = [3]float64{7000, 0, 0}
	propagation.VelocityKmPerS = [3]float64{0, 7.5, 0}
	propagation.ForceModel = PropagationForceModelTwoBody
	ephemeris, err := PropagateCovariance(propagation, base, []float64{0}, CovarianceInertial, CovarianceInertial, ProcessNoise{Kind: ProcessNoiseNone})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ephemeris.Close() })
	if count, err := ephemeris.Count(); err != nil || count != 1 {
		t.Fatalf("covariance ephemeris count = %d, %v", count, err)
	}
	at, err := ephemeris.CovarianceAt(0)
	if err != nil || at.Values[2][2] != 3 {
		t.Fatalf("covariance at epoch = %#v, %v", at, err)
	}
}

func TestDeterministicReducedAndFDE(t *testing.T) {
	samples := []ECEFSample{
		{CalendarEpoch{2020, 6, 24, 0, 0, 0}, 7093820.244, -22286025.531, 12351111.011},
		{CalendarEpoch{2020, 6, 24, 0, 15, 0}, 7438042.916, -20762704.119, 14621192.800},
		{CalendarEpoch{2020, 6, 24, 0, 30, 0}, 7899334.503, -19019862.023, 16636833.344},
		{CalendarEpoch{2020, 6, 24, 0, 45, 0}, 8498085.463, -17097635.989, 18363427.105},
		{CalendarEpoch{2020, 6, 24, 1, 0, 0}, 9247369.015, -15040654.467, 19771547.156},
	}
	elements, stats, err := ReducedOrbitFit(samples, GPST, ReducedOrbitCircularSecular)
	if err != nil || stats.SampleCount != len(samples) || !math.IsInf(elements.SemiMajorAxisM, 0) && !(elements.SemiMajorAxisM > 0) {
		t.Fatalf("reduced fit = %+v, %+v, %v", elements, stats, err)
	}
	position, err := elements.Position(samples[0].Epoch, GPST, ReducedOrbitECEF)
	if err != nil || !finite3(position) {
		t.Fatalf("reduced position = %#v, %v", position, err)
	}
	report, err := ReducedOrbitDrift(elements, samples, GPST, 1e9)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = report.Close() })
	entries, summary, requested, err := report.Output()
	if err != nil || requested != len(samples) || len(entries) != len(samples) || summary.HasThresholdCrossing {
		t.Fatalf("reduced drift = %#v, %#v, %d, %v", entries, summary, requested, err)
	}

	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	fde, err := SolveFDE(sp3, usedSPPConfig(), FDEOptions{PFA: 1e-3, MaxIterations: 8, UnitWeights: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = fde.Close() })
	diagnostics, err := fde.Diagnostics()
	if err != nil || diagnostics.Iterations != 0 || len(diagnostics.ExcludedSatelliteIDs) != 0 {
		t.Fatalf("FDE diagnostics = %+v, %v", diagnostics, err)
	}
	solution, err := fde.Solution()
	if err != nil || solution.UsedSatelliteCount < 4 || solution.PositionM == [3]float64{} {
		t.Fatalf("FDE solution = %+v, %v", solution, err)
	}
}

func TestDeterministicReliabilityARAIMAndRangeFDE(t *testing.T) {
	reliabilityOptions, err := DefaultReliabilityOptions()
	if err != nil {
		t.Fatal(err)
	}
	reliability, err := ReliabilityFromDesign([]ReliabilityRow{
		{SatelliteID: "rx_clock", DesignRow: []float64{1, 0}, SigmaM: 1},
		{SatelliteID: "range_a", DesignRow: []float64{0, 1}, SigmaM: 1},
		{SatelliteID: "range_b", DesignRow: []float64{0, 1}, SigmaM: 1},
	}, reliabilityOptions)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reliability.Close() })
	summary, err := reliability.Summary()
	if err != nil || summary.DOF != 1 || summary.UncheckableCount != 1 {
		t.Fatalf("reliability summary = %+v, %v", summary, err)
	}
	observations, err := reliability.Observations()
	if err != nil || len(observations) != 3 || !observations[0].Uncheckable || observations[0].HasMDB || observations[0].MDBM != 0 || observations[1].Uncheckable || !observations[1].HasMDB {
		t.Fatalf("reliability observations = %#v, %v", observations, err)
	}
	if _, err := WTestNoncentralityFor(0.001, 0.1); err != nil {
		t.Fatal(err)
	}

	rangeResult, err := RunRangeFDE([]RangeFDERow{
		{ID: "rx_clock", ResidualM: 0, DesignRow: []float64{1, 0}, Weight: 1},
		{ID: "range_a", ResidualM: 0.1, DesignRow: []float64{0, 1}, Weight: 1},
		{ID: "range_b", ResidualM: -0.1, DesignRow: []float64{0, 1}, Weight: 1},
	}, RangeFDEOptions{PFA: 1e-3, MaxExclusions: 1, MinRedundancy: 1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rangeResult.Close() })
	rangeOutput, err := rangeResult.Output()
	if err != nil || rangeOutput.StateDimension != 2 || len(rangeOutput.Diagnostics) != 3 || len(rangeOutput.Covariance) != 4 {
		t.Fatalf("range FDE output = %+v, %v", rangeOutput, err)
	}

	geometry := ARAIMGeometry{Rows: []ARAIMRow{
		{SatelliteID: "G01", LineOfSight: LineOfSight{0.0966, -0.0225, -0.9951}, System: 0, ElevationRad: math.Pi / 2},
		{SatelliteID: "G02", LineOfSight: LineOfSight{0.2612, -0.6750, 0.6900}, System: 0, ElevationRad: math.Pi / 2},
		{SatelliteID: "G03", LineOfSight: LineOfSight{0.7477, -0.0723, 0.6601}, System: 0, ElevationRad: math.Pi / 2},
		{SatelliteID: "G04", LineOfSight: LineOfSight{0.2269, 0.9398, -0.2553}, System: 0, ElevationRad: math.Pi / 2},
		{SatelliteID: "G05", LineOfSight: LineOfSight{0.2877, 0.5907, 0.7539}, System: 0, ElevationRad: math.Pi / 2},
		{SatelliteID: "E01", LineOfSight: LineOfSight{0.9455, 0.3236, 0.0354}, System: 2, ElevationRad: math.Pi / 2},
		{SatelliteID: "E02", LineOfSight: LineOfSight{0.5957, 0.6748, -0.4356}, System: 2, ElevationRad: math.Pi / 2},
		{SatelliteID: "E03", LineOfSight: LineOfSight{0.7075, -0.0938, 0.7004}, System: 2, ElevationRad: math.Pi / 2},
		{SatelliteID: "E04", LineOfSight: LineOfSight{0.7709, -0.5571, -0.3088}, System: 2, ElevationRad: math.Pi / 2},
		{SatelliteID: "E05", LineOfSight: LineOfSight{0.2780, -0.6622, -0.6958}, System: 2, ElevationRad: math.Pi / 2},
	}, ClockSystems: []uint32{0, 2}}
	model := ARAIMSatelliteModel{SigmaURAM: 0.75, SigmaUREM: 0.5, NominalBiasM: 0.5, SatelliteFaultProbability: 1e-5}
	ism := ARAIMISM{Constellations: []ARAIMConstellationISM{{System: 0, ConstellationFaultProbability: 1e-4, DefaultSatellite: model}, {System: 2, ConstellationFaultProbability: 1e-4, DefaultSatellite: model}}}
	effective := []struct {
		id        string
		integrity float64
		accuracy  float64
	}{
		{"G01", 3.8865, 3.5740}, {"G02", 1.4377, 1.1252}, {"G03", 0.8604, 0.5479},
		{"G04", 1.6383, 1.3258}, {"G05", 1.3229, 1.0104}, {"E01", 0.8434, 0.5309},
		{"E02", 0.8963, 0.5838}, {"E03", 0.8669, 0.5544}, {"E04", 0.8573, 0.5448},
		{"E05", 1.3616, 1.0491},
	}
	for _, value := range effective {
		satelliteModel := model
		satelliteModel.HasEffectiveIntegrity = true
		satelliteModel.EffectiveIntegrityM = math.Sqrt(value.integrity)
		satelliteModel.HasEffectiveAccuracy = true
		satelliteModel.EffectiveAccuracyM = math.Sqrt(value.accuracy)
		ism.Satellites = append(ism.Satellites, ARAIMSatelliteISM{SatelliteID: value.id, Model: satelliteModel})
	}
	allocation, err := ARAIMAllocationLPV200()
	if err != nil {
		t.Fatal(err)
	}
	result, err := RunARAIM(geometry, ism, allocation)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = result.Close() })
	araimSummary, err := result.Summary()
	if err != nil || !araimSummary.Available || !araimSummary.Availability || araimSummary.FaultModeCount != 13 || math.Abs(araimSummary.VPLM-19.2) > 0.05 {
		t.Fatalf("ARAIM summary = %+v, %v", araimSummary, err)
	}
	faultModes, err := result.FaultModes()
	if err != nil || len(faultModes) != 13 || !faultModes[0].Monitorable || faultModes[0].ExcludedCount != 0 || faultModes[0].HasExcludedConstellation {
		t.Fatalf("ARAIM fault modes = %#v, %v", faultModes, err)
	}
	if _, err := result.ExcludedSatellites(-1); err == nil {
		t.Fatal("ARAIM accepted a negative fault-mode index")
	}
}

func TestSurfaceClosedCopyRoutes(t *testing.T) {
	config, err := DefaultFusionFilterConfig()
	if err != nil {
		t.Fatal(err)
	}
	filter, err := NewFusionFilter(FusionNavState{AttitudeBodyToECEF: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}}, make([]float64, 15), config)
	if err != nil {
		t.Fatal(err)
	}
	if err := filter.Close(); err != nil || filter.Close() != nil {
		t.Fatalf("double close = %v", err)
	}
	if _, err := filter.Encode(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Encode after Close = %v, want ErrClosed", err)
	}
}

func finite3(value [3]float64) bool {
	for _, component := range value {
		if math.IsInf(component, 0) || math.IsNaN(component) {
			return false
		}
	}
	return true
}
