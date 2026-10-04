package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func isFinitePositiveDefinite3x3(cov [9]float64) bool {
	for i, value := range cov {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
		r, c := i/3, i%3
		if math.Abs(value-cov[c*3+r]) > 1e-9 {
			return false
		}
	}
	minor2 := cov[0]*cov[4] - cov[1]*cov[3]
	determinant := cov[0]*(cov[4]*cov[8]-cov[5]*cov[7]) - cov[1]*(cov[3]*cov[8]-cov[5]*cov[6]) + cov[2]*(cov[3]*cov[7]-cov[4]*cov[6])
	return cov[0] > 0 && minor2 > 0 && determinant > 0
}

func TestStaticPositionSP3PublicFixture(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Error(err)
		}
	})
	cfg := usedSPPConfig()
	epochs := []StaticPositionEpoch{{Inputs: SPPInputsV2{Base: cfg}}, {Inputs: SPPInputsV2{Base: cfg}}}
	solution, kind, err := SolveStaticPositionSP3(sp3, epochs, nil)
	if err != nil {
		t.Fatalf("kind=%d: %v", kind, err)
	}
	t.Cleanup(func() {
		if err := solution.Close(); err != nil {
			t.Error(err)
		}
	})
	position, err := solution.Position()
	if err != nil {
		t.Fatal(err)
	}
	// Independent weighted least squares over the same eight pseudoranges and
	// pinned RTKLIB transmit-time states gives this position. Both static
	// epochs below use those same inputs, so their shared position is the same
	// independent solution within the solver's sub-millimetre tolerance.
	wantPosition := [3]float64{4484137.56180868, 550578.016017378, 4487569.615357698}
	for axis, want := range wantPosition {
		if math.Abs(position[axis]-want) > 5e-5 {
			t.Fatalf("position[%d] = %.17g, independent reference %.17g", axis, position[axis], want)
		}
	}
	ecef, err := solution.PositionCovarianceECEFM2()
	if err != nil {
		t.Fatal(err)
	}
	// The two independent epochs double the normal information relative to
	// the independently inverted one-epoch SPP covariance.
	wantECEF := [9]float64{42.016288023818555, -5.14320054743515, 22.206168343362833, -5.143200547435147, 9.33860271343945, 0.685775762034214, 22.206168343362825, 0.6857757620342127, 36.91385388069771}
	for i, want := range wantECEF {
		if math.Abs(ecef[i]-want) > 1e-3 {
			t.Fatalf("ECEF covariance[%d] = %.17g, independent two-epoch reference %.17g", i, ecef[i], want)
		}
	}
	enu, err := solution.PositionCovarianceENUM2()
	if err != nil {
		t.Fatal(err)
	}
	wantENU := [9]float64{11.06816995312946, 4.891449846163921, -7.756029199979944, 4.891449846163922, 16.47606321119846, -1.6864335678206055, -7.7560291999799445, -1.6864335678205948, 60.72451145362779}
	for i, want := range wantENU {
		if math.Abs(enu[i]-want) > 1e-3 {
			t.Fatalf("ENU covariance[%d] = %.17g, independent two-epoch reference %.17g", i, enu[i], want)
		}
	}
	clocks, err := solution.ClockBiases()
	const independentClockS = 0.00010009082050400748
	if err != nil || len(clocks) != 2 {
		t.Fatalf("clock biases = %#v, %v", clocks, err)
	}
	for i, clock := range clocks {
		if clock.EpochIndex != i || clock.System != GNSSSystemGPS || math.Abs(clock.ClockS-independentClockS) > 1e-12 {
			t.Fatalf("clock bias[%d] = %#v, independent reference %.17g s", i, clock, independentClockS)
		}
	}
	influence, err := solution.EpochInfluence()
	const independentResidualRMSM = 2.5580879486940282
	if err != nil || len(influence) != 2 {
		t.Fatalf("epoch influence = %#v, %v", influence, err)
	}
	for i, value := range influence {
		if value.EpochIndex != i || value.OmittedMeasurements != 8 || value.Status != StaticInfluenceSolved || !value.HasPositionDelta || value.PositionDeltaNormM > 1e-7 || !value.HasResidualRMS || math.Abs(value.ResidualRMSM-independentResidualRMSM) > 5e-3 {
			t.Fatalf("epoch influence[%d] = %#v, independent residual RMS %.12f m", i, value, independentResidualRMSM)
		}
	}
	geo, present, err := solution.Geodetic()
	if err != nil || present || geo != (Geodetic{}) {
		t.Fatalf("geodetic = %#v present=%v err=%v", geo, present, err)
	}
	metadata, err := solution.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	const independentUnitWeightGDOP = 2.116646865130595
	if metadata.Iterations <= 0 || metadata.Iterations > 100 || metadata.OuterIterations != 0 || metadata.UsedMeasurements != 16 || metadata.Parameters != 5 || !metadata.Converged || metadata.Status != StaticPositionSolveSelectionSettled || metadata.Redundancy != 11 || metadata.GeometryQuality.Tier != 3 || metadata.GeometryQuality.Rank != 5 || (math.IsNaN(metadata.GeometryQuality.ConditionNumber) || math.IsInf(metadata.GeometryQuality.ConditionNumber, 0)) || metadata.GeometryQuality.ConditionNumber <= 1 || metadata.GeometryQuality.ConditionNumber >= 100 || math.Abs(metadata.GeometryQuality.GDOP-independentUnitWeightGDOP) > 1e-8 {
		t.Fatalf("metadata = %#v, independent unit-weight GDOP %.15g", metadata, independentUnitWeightGDOP)
	}
	rejected, err := solution.RejectedSatellites(0)
	if err != nil || len(rejected) != 0 {
		t.Fatalf("rejected satellites = %#v, %v", rejected, err)
	}
	if _, err := solution.RejectedSatellites(-1); err == nil {
		t.Fatal("negative static epoch index accepted")
	}
	residuals, err := solution.Residuals()
	if err != nil || len(residuals) != 16 {
		t.Fatalf("residuals = %#v, %v", residuals, err)
	}
	wantIDs := []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}
	wantResiduals := []float64{-0.7862987704575062, -2.770254924893379, -0.6641135476529598, -0.17380670458078384, 3.825963206589222, -4.198393113911152, 2.1111478097736835, 2.6201590932905674}
	for i, value := range residuals {
		ref := i % len(wantIDs)
		if value.EpochIndex != i/len(wantIDs) || value.SatelliteID != wantIDs[ref] || math.Abs(value.ResidualM-wantResiduals[ref]) > 5e-5 || value.BaseWeight <= 0 || value.EffectiveWeight != value.BaseWeight || value.RobustWeightRatio != 1 {
			t.Fatalf("residual[%d] = %#v, independent residual %.12f m", i, value, wantResiduals[ref])
		}
	}
	batchInfluence, err := solution.SatelliteBatchInfluence()
	if err != nil || len(batchInfluence) != 8 {
		t.Fatalf("batch influence = %#v, %v", batchInfluence, err)
	}
	wantBatchDelta := [8][3]float64{
		{3.567821942269802, 1.7024530454073101, 3.0786527767777443},
		{1.0563432946801186, -0.83092282328289, 6.230084664188325},
		{-0.7386459521949291, 0.19758936041034758, -0.6311932364478707},
		{0.5769509514793754, -0.29312286037020385, -0.06301376968622208},
		{-0.5909471698105335, 1.5896472202148288, -0.4200156293809414},
		{-2.843756714835763, -0.4351127319969237, -4.280577372759581},
		{2.0933722890913486, -0.3323887620354071, -0.3552760975435376},
		{0.4980939133092761, -0.9955234124790877, 1.3283235589042306},
	}
	wantBatchRMS := [8]float64{2.6420916296201002, 2.1339377637979875, 2.7150254966926415, 2.731446578426774, 2.1186932639042344, 1.8943474880833497, 2.5320624138597214, 2.4483473901183417}
	for i, value := range batchInfluence {
		if value.SatelliteID != wantIDs[i] || value.OmittedMeasurements != 2 || value.Status != StaticInfluenceSolved || !value.HasPositionDelta || !value.HasResidualRMS || value.MinRobustWeightRatio != 1 || math.Abs(value.ResidualRMSM-wantBatchRMS[i]) > 5e-3 {
			t.Fatalf("batch influence[%d] = %#v, independent residual RMS %.9f m", i, value, wantBatchRMS[i])
		}
		for axis := 0; axis < 3; axis++ {
			if math.Abs(value.PositionDeltaM[axis]-wantBatchDelta[i][axis]) > 1e-3 {
				t.Fatalf("batch influence[%d].delta[%d] = %.12f, independent reference %.12f", i, axis, value.PositionDeltaM[axis], wantBatchDelta[i][axis])
			}
		}
	}
	satInfluence, err := solution.SatelliteInfluence()
	if err != nil || len(satInfluence) != 16 {
		t.Fatalf("satellite influence = %#v, %v", satInfluence, err)
	}
	wantSingleDelta := [16][3]float64{
		{0.5715435137972236, 0.2727226832648739, 0.4931815732270479},
		{0.3134796926751733, -0.24658403731882572, 1.848835232667625},
		{-0.30328692961484194, 0.08112989622168243, -0.25916700530797243},
		{0.1277855047956109, -0.0649220731575042, -0.013956548646092415},
		{-0.26246412191540003, 0.7060280845034868, -0.18654633779078722},
		{-1.2141004065051675, -0.18576504744123667, -1.8275299975648522},
		{0.8509118165820837, -0.13510903890710324, -0.1444122800603509},
		{0.213731087744236, -0.42717710754368454, 0.5699809603393078},
		{0.5715435137972236, 0.27272268349770457, 0.4931815732270479},
		{0.313479695469141, -0.24658403906505555, 1.84883523453027},
		{-0.30328693334013224, 0.08112989482469857, -0.25916700437664986},
		{0.1277855010703206, -0.06492207536939532, -0.013956553302705288},
		{-0.2624641256406903, 0.7060280842706561, -0.18654634058475494},
		{-1.2141004065051675, -0.18576504744123667, -1.8275299975648522},
		{0.8509118175134063, -0.1351090376265347, -0.14441228099167347},
		{0.2137310840189457, -0.4271771074272692, 0.5699809594079852},
	}
	for i, value := range satInfluence {
		ref := i % len(wantIDs)
		if value.SatelliteID != wantIDs[ref] || value.EpochIndex != i/len(wantIDs) || value.Status != StaticInfluenceSolved || !value.HasPositionDelta || math.Abs(value.ResidualM-wantResiduals[ref]) > 5e-5 {
			t.Fatalf("satellite influence[%d] = %#v, independent residual %.12f m", i, value, wantResiduals[ref])
		}
		for axis := 0; axis < 3; axis++ {
			if math.Abs(value.PositionDeltaM[axis]-wantSingleDelta[i][axis]) > 1e-3 {
				t.Fatalf("satellite influence[%d].delta[%d] = %.12f, independent reference %.12f", i, axis, value.PositionDeltaM[axis], wantSingleDelta[i][axis])
			}
		}
	}
	state, err := solution.StateCovarianceM2()
	if err != nil || len(state) != 25 || math.Float64bits(state[0]) != math.Float64bits(ecef[0]) {
		t.Fatalf("state covariance = %#v, %v", state, err)
	}
	wantState := [25]float64{
		42.01636777799649, -5.143020452517948, 22.206181027562145, 33.56307403682808, 33.563074036828084,
		-5.143020452517946, 9.338546269949656, 0.6858985713525081, -1.9034783993484223, -1.9034783993484226,
		22.206181027562153, 0.6858985713525025, 36.91386549251835, 30.188537211785476, 30.18853721178548,
		33.563074036828084, -1.9034783993484266, 30.188537211785473, 37.51750277215632, 32.99986846769158,
		33.563074036828084, -1.9034783993484266, 30.188537211785476, 32.99986846769157, 37.51750277215633,
	}
	for i, want := range wantState {
		if math.Abs(state[i]-want) > 1e-3 {
			t.Fatalf("state covariance[%d] = %.17g, independent reference %.17g", i, state[i], want)
		}
	}
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.Position(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Position after Close = %v", err)
	}
}

func TestStaticPositionInvalidAndClose(t *testing.T) {
	defaults, err := DefaultStaticPositionOptions()
	if err != nil {
		t.Fatal(err)
	}
	if defaults != (StaticPositionOptions{Robust: SPPRobustConfig{HuberK: 1.345, ScaleFloorM: 1, MaxOuter: 100, OuterToleranceM: 0.0001}}) {
		t.Fatalf("static defaults = %#v", defaults)
	}
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	if _, kind, err := SolveStaticPositionSP3(sp3, nil, nil); err == nil || kind != StaticPositionErrorEmptyEpochs {
		t.Fatalf("empty solve = kind %d err %v", kind, err)
	}
	if err := sp3.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := SolveStaticPositionSP3(sp3, nil, nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed solve = %v", err)
	}
	var zero StaticPositionSolution
	if err := zero.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := zero.Position(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero position = %v", err)
	}
	badWeights := []StaticPositionEpoch{{Inputs: SPPInputsV2{Base: usedSPPConfig()}, Weights: []float64{1}}}
	open, kind, err := SolveStaticPositionSP3(nil, badWeights, nil)
	if open != nil || kind != 0 || !errors.Is(err, ErrClosed) {
		t.Fatalf("nil static solve = handle %v kind %d err %v", open, kind, err)
	}
	sp3, err = LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	if _, _, err := SolveStaticPositionSP3(sp3, badWeights, nil); err == nil {
		t.Fatal("mismatched static weights accepted")
	}
	if _, _, err := SolveStaticPositionSP3(sp3, []StaticPositionEpoch{{Inputs: SPPInputsV2{Base: usedSPPConfig()}}}, &StaticPositionOptions{Robust: SPPRobustConfig{MaxOuter: -1}}); err == nil {
		t.Fatal("negative static robust count accepted")
	}
}

func TestStaticPositionCloseReadRace(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	solution, _, err := SolveStaticPositionSP3(sp3, []StaticPositionEpoch{{Inputs: SPPInputsV2{Base: usedSPPConfig()}}, {Inputs: SPPInputsV2{Base: usedSPPConfig()}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 20; j++ {
				_, _ = solution.Position()
				_, _ = solution.Metadata()
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		for i := 0; i < 10; i++ {
			_ = solution.Close()
		}
	}()
	group.Wait()
	if _, err := solution.Position(); !errors.Is(err, ErrClosed) {
		t.Fatalf("post-race position = %v", err)
	}
}

func TestStaticPositionBroadcastFixture(t *testing.T) {
	broadcast, err := ParseBroadcastEphemeris(readPositioningFixture(t, "nav/ESBC00DNK_R_20201770000_01D_MN.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = broadcast.Close() })
	obs, err := ParseRINEXObservation(readPositioningFixture(t, "obs/ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = obs.Close() })
	assembled, err := SPPInputsFromRINEXObs(obs, broadcast, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = assembled.Close() })
	count, err := assembled.Count()
	if err != nil || count != 2 {
		t.Fatalf("assembled input count = %d, %v", count, err)
	}
	epochs := make([]StaticPositionEpoch, count)
	for i := range epochs {
		inputs, err := assembled.EpochInputs(i)
		if err != nil {
			t.Fatal(err)
		}
		epochs[i] = StaticPositionEpoch{Inputs: inputs}
	}
	options := &StaticPositionOptions{InitialPositionM: [3]float64{3582105.291, 532589.7313, 5232754.8054}, WithGeodetic: true}
	solution, kind, err := SolveStaticPositionBroadcast(broadcast, epochs, options)
	if err != nil {
		t.Fatalf("kind=%d: %v", kind, err)
	}
	t.Cleanup(func() { _ = solution.Close() })
	position, err := solution.Position()
	if err != nil {
		t.Fatal(err)
	}
	// The RINEX marker coordinates provide an independent, metre-scale
	// reference; the broadcast solve is expected to remain within a few metres.
	marker := [3]float64{3582105.291, 532589.7313, 5232754.8054}
	deltaM := 0.0
	for axis := range position {
		delta := position[axis] - marker[axis]
		deltaM += delta * delta
	}
	if math.Sqrt(deltaM) > 5 {
		t.Fatalf("broadcast position %#v is more than 5 m from RINEX marker %#v", position, marker)
	}
	ecef, err := solution.PositionCovarianceECEFM2()
	if err != nil {
		t.Fatal(err)
	}
	if !isFinitePositiveDefinite3x3(ecef) {
		t.Fatalf("broadcast ECEF covariance is not finite positive-definite: %#v", ecef)
	}
	clocks, err := solution.ClockBiases()
	if err != nil || len(clocks) != 6 {
		t.Fatalf("broadcast clock biases = %#v, %v", clocks, err)
	}
	systems := []GNSSSystem{GNSSSystemGPS, GNSSSystemGalileo, GNSSSystemBeiDou}
	for i, clock := range clocks {
		if clock.EpochIndex != i/3 || clock.System != systems[i%3] || math.IsNaN(clock.ClockS) || math.IsInf(clock.ClockS, 0) || math.Abs(clock.ClockS) >= 0.001 {
			t.Fatalf("broadcast clock bias[%d] = %#v", i, clock)
		}
	}
	metadata, err := solution.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Iterations <= 0 || metadata.Iterations > 100 || metadata.UsedMeasurements != 48 || metadata.Parameters != 9 || !metadata.Converged || metadata.Status > StaticPositionSolveSelectionSettled || metadata.Redundancy != 39 || metadata.GeometryQuality.Tier != 3 || metadata.GeometryQuality.Rank != 9 || math.IsNaN(metadata.GeometryQuality.ConditionNumber) || math.IsInf(metadata.GeometryQuality.ConditionNumber, 0) || metadata.GeometryQuality.ConditionNumber <= 1 || metadata.GeometryQuality.ConditionNumber >= 1000 || math.IsNaN(metadata.GeometryQuality.GDOP) || math.IsInf(metadata.GeometryQuality.GDOP, 0) || metadata.GeometryQuality.GDOP <= 1 || metadata.GeometryQuality.GDOP >= 10 {
		t.Fatalf("broadcast metadata = %#v", metadata)
	}
	enu, err := solution.PositionCovarianceENUM2()
	if err != nil {
		t.Fatal(err)
	}
	if !isFinitePositiveDefinite3x3(enu) {
		t.Fatalf("broadcast ENU covariance is not finite positive-definite: %#v", enu)
	}
	influence, err := solution.EpochInfluence()
	if err != nil {
		t.Fatal(err)
	}
	if len(influence) != 2 {
		t.Fatalf("broadcast epoch influence = %#v", influence)
	}
	for i, value := range influence {
		if value.EpochIndex != i || value.OmittedMeasurements <= 0 || value.Status != StaticInfluenceSolved || !value.HasPositionDelta || math.IsNaN(value.PositionDeltaNormM) || math.IsInf(value.PositionDeltaNormM, 0) || !value.HasResidualRMS || math.IsNaN(value.ResidualRMSM) || math.IsInf(value.ResidualRMSM, 0) {
			t.Fatalf("broadcast epoch influence[%d] = %#v", i, value)
		}
	}
	geo, present, err := solution.Geodetic()
	if err != nil {
		t.Fatal(err)
	}
	if !present || math.Abs(geo.LatitudeRad-0.9685) > 1e-4 || math.Abs(geo.LongitudeRad-0.1476) > 1e-4 || math.Abs(geo.HeightM-59.4) > 10 {
		t.Fatalf("broadcast geodetic = %#v present=%v", geo, present)
	}
	rejected, err := solution.RejectedSatellites(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rejected) != 15 || rejected[0] != (StaticPositionRejectedSatellite{SatelliteID: "G02", Reason: StaticPositionRejectionLowElevation}) || rejected[3] != (StaticPositionRejectedSatellite{SatelliteID: "R01", Reason: StaticPositionRejectionNoEphemeris}) {
		t.Fatalf("broadcast rejected satellites = %#v", rejected)
	}
	residuals, err := solution.Residuals()
	if err != nil {
		t.Fatal(err)
	}
	if len(residuals) != 48 || residuals[0].EpochIndex != 0 || residuals[0].SatelliteID != "G05" {
		t.Fatalf("broadcast residuals = %#v", residuals)
	}
	for i, value := range residuals {
		if value.EpochIndex != i/24 || value.SatelliteID == "" || math.IsNaN(value.ResidualM) || math.IsInf(value.ResidualM, 0) || value.BaseWeight <= 0 || math.IsNaN(value.BaseWeight) || math.IsInf(value.BaseWeight, 0) || value.EffectiveWeight <= 0 || value.RobustWeightRatio <= 0 || value.RobustWeightRatio > 1 {
			t.Fatalf("broadcast residual[%d] = %#v", i, value)
		}
	}
	batch, err := solution.SatelliteBatchInfluence()
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) != 24 || batch[0].SatelliteID != "G05" || batch[0].OmittedMeasurements != 2 {
		t.Fatalf("broadcast batch influence = %#v", batch)
	}
	for i, value := range batch {
		if value.SatelliteID == "" || value.OmittedMeasurements != 2 || value.Status != StaticInfluenceSolved || !value.HasPositionDelta || math.IsNaN(value.PositionDeltaNormM) || math.IsInf(value.PositionDeltaNormM, 0) || !value.HasResidualRMS || math.IsNaN(value.ResidualRMSM) || math.IsInf(value.ResidualRMSM, 0) {
			t.Fatalf("broadcast batch influence[%d] = %#v", i, value)
		}
	}
	sat, err := solution.SatelliteInfluence()
	if err != nil {
		t.Fatal(err)
	}
	if len(sat) != 48 || sat[0].SatelliteID != "G05" || sat[0].EpochIndex != 0 {
		t.Fatalf("broadcast satellite influence = %#v", sat)
	}
	for i, value := range sat {
		if value.SatelliteID == "" || value.EpochIndex != i/24 || !value.HasPositionDelta || math.IsNaN(value.ResidualM) || math.IsInf(value.ResidualM, 0) || value.BaseWeight <= 0 || value.EffectiveWeight <= 0 || value.RobustWeightRatio <= 0 || value.RobustWeightRatio > 1 {
			t.Fatalf("broadcast satellite influence[%d] = %#v", i, value)
		}
	}
	state, err := solution.StateCovarianceM2()
	if err != nil {
		t.Fatal(err)
	}
	if len(state) != 81 || state[len(state)-1] <= 0 {
		t.Fatalf("broadcast state covariance = %#v", state)
	}
	for i, value := range state {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			t.Fatalf("broadcast state covariance[%d] is non-finite: %g", i, value)
		}
	}
	for row := 0; row < 3; row++ {
		for column := 0; column < 3; column++ {
			if math.Abs(state[row*9+column]-ecef[row*3+column]) > 1e-12 {
				t.Fatalf("broadcast state position covariance[%d,%d] differs from ECEF covariance", row, column)
			}
		}
	}
}

func TestStaticReferenceStationRoute(t *testing.T) {
	sp3, reference, rover := rinexRTKFixture(t)
	cfg, err := DefaultStaticReferenceStationRinexConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ReferencePositionM != [3]float64{} || !cfg.EnableCodeDGNSS || !cfg.EnableCarrierRTK || !cfg.WithGeodetic || cfg.Carrier.ArcOptions.MinCommonSatellites != 4 || !cfg.Carrier.ArcOptions.IncludePredictionTime || cfg.Carrier.Model.CodeSigmaM != 0.3 || cfg.Carrier.Model.PhaseSigmaM != 0.003 || !cfg.Carrier.Model.Sagnac || cfg.Carrier.FloatOptions.MaxIterations != 10 || cfg.Carrier.FixedOptions.MaxIterations != 10 {
		t.Fatalf("reference defaults = %#v", cfg)
	}
	if _, err := SolveStaticReferenceStationRINEX(nil, nil, nil, cfg); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil reference solve = %v", err)
	}
	cfg.ReferencePositionM = [3]float64{3582105.291, 532589.7313, 5232754.8054}
	cfg.EnableCodeDGNSS = true
	solution, err := SolveStaticReferenceStationRINEX(sp3, reference, rover, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = solution.Close() })
	baseline, err := solution.BaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	if baseline != [3]float64{} {
		t.Fatalf("baseline = %#v", baseline)
	}
	position, err := solution.PositionECEF()
	if err != nil {
		t.Fatal(err)
	}
	if position != [3]float64{3582105.291, 532589.7313, 5232754.8054} {
		t.Fatalf("reference position = %#v", position)
	}
	ecef, err := solution.CovarianceECEF()
	if err != nil {
		t.Fatal(err)
	}
	if !isFinitePositiveDefinite3x3(ecef) {
		t.Fatalf("reference ECEF covariance is not finite positive-definite: %#v", ecef)
	}
	enu, err := solution.CovarianceENU()
	if err != nil {
		t.Fatal(err)
	}
	if !isFinitePositiveDefinite3x3(enu) {
		t.Fatalf("reference ENU covariance is not finite positive-definite: %#v", enu)
	}
	diagnostics, err := solution.Diagnostics()
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 2 || diagnostics[0] != (StaticReferenceEpochDiagnostic{Mode: StaticReferenceModeCarrierFixed, EpochIndex: 0, UsedSatelliteCount: 3, HasCodeResidualRMS: true, HasPhaseResidualRMS: true, HasResidualRMS: true}) || diagnostics[1] != (StaticReferenceEpochDiagnostic{Mode: StaticReferenceModeCarrierFixed, EpochIndex: 1, UsedSatelliteCount: 3, HasCodeResidualRMS: true, HasPhaseResidualRMS: true, HasResidualRMS: true}) {
		t.Fatalf("reference diagnostics = %#v", diagnostics)
	}
	metadata, err := solution.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Mode != StaticReferenceModeCarrierFixed || metadata.FixStatus != StaticReferenceFixCarrierFixed || !metadata.HasGeodetic || metadata.Geodetic != (Geodetic{LatitudeRad: 0.9685453838806702, LongitudeRad: 0.14759937748625832, HeightM: 59.47648589304751}) || metadata.BaselineM != 0 || !metadata.HasCodeSolution || !metadata.HasCarrierSolution || metadata.DiagnosticCount != 2 || metadata.ModeReportCount != 2 || metadata.CarrierIntegerStatus != RTKIntegerFixed || !metadata.HasCarrierIntegerRatio || metadata.CarrierIntegerRatio != math.MaxFloat64 || metadata.CodeDiagnosticCount != 2 || metadata.CarrierDiagnosticCount != 2 {
		t.Fatalf("reference metadata = %#v", metadata)
	}
	reports, err := solution.ModeReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 || reports[0] != (StaticReferenceModeReport{Mode: StaticReferenceModeCodeDGNSS, Status: StaticReferenceModeSolved, UsedEpochs: 2, UsedMeasurements: 8}) || reports[1] != (StaticReferenceModeReport{Mode: StaticReferenceModeCarrierFixed, Status: StaticReferenceModeSolved, UsedEpochs: 2, UsedMeasurements: 12}) {
		t.Fatalf("reference reports = %#v", reports)
	}
}

func TestStaticReferenceCloseReadRace(t *testing.T) {
	var zero StaticReferenceStationSolution
	if err := zero.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := zero.PositionECEF(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero reference position = %v", err)
	}
	sp3, reference, rover := rinexRTKFixture(t)
	config, err := DefaultStaticReferenceStationRinexConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.ReferencePositionM = [3]float64{3582105.291, 532589.7313, 5232754.8054}
	solution, err := SolveStaticReferenceStationRINEX(sp3, reference, rover, config)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 20; j++ {
				_, _ = solution.PositionECEF()
				_, _ = solution.Metadata()
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		for i := 0; i < 10; i++ {
			_ = solution.Close()
		}
	}()
	group.Wait()
	if _, err := solution.PositionECEF(); !errors.Is(err, ErrClosed) {
		t.Fatalf("post-race reference position = %v", err)
	}
}
