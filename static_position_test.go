package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

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
	if position != [3]float64{4484127.991055579, 550581.6856632147, 4487560.540095652} {
		t.Fatalf("position = %#v", position)
	}
	ecef, err := solution.PositionCovarianceECEFM2()
	if err != nil {
		t.Fatal(err)
	}
	if ecef != [9]float64{5.877755853743985, 0.051250589926605916, 2.622850226493485, 0.051250589926605916, 1.418834303773507, 0.8716461355572004, 2.622850226493485, 0.8716461355572004, 3.0021283778640564} {
		t.Fatalf("ECEF covariance = %#v", ecef)
	}
	enu, err := solution.PositionCovarianceENUM2()
	if err != nil {
		t.Fatal(err)
	}
	if enu != [9]float64{1.4726601768633898, 0.7319482419316552, 0.03951088901828653, 0.7319482419316552, 1.7035023383074528, -1.410900801045587, 0.03951088901828648, -1.4109008010455883, 7.122556020210707} {
		t.Fatalf("ENU covariance = %#v", enu)
	}
	clocks, err := solution.ClockBiases()
	if err != nil || len(clocks) != 2 || clocks[0] != (StaticPositionClockBias{EpochIndex: 0, System: GNSSSystemGPS, ClockS: 0.00010006922168398482}) || clocks[1] != (StaticPositionClockBias{EpochIndex: 1, System: GNSSSystemGPS, ClockS: 0.00010006922168398482}) {
		t.Fatalf("clock biases = %#v, %v", clocks, err)
	}
	influence, err := solution.EpochInfluence()
	if err != nil || len(influence) != 2 || influence[0].OmittedMeasurements != 8 || influence[0].Status != StaticInfluenceSolved || !influence[0].HasPositionDelta || influence[0].PositionDeltaM != [3]float64{-2.514570951461792e-08, -3.958120942115784e-09, -1.210719347000122e-08} || influence[0].ResidualRMSM != 0.0004936508030193884 {
		t.Fatalf("epoch influence = %#v, %v", influence, err)
	}
	geo, present, err := solution.Geodetic()
	if err != nil || present || geo != (Geodetic{}) {
		t.Fatalf("geodetic = %#v present=%v err=%v", geo, present, err)
	}
	metadata, err := solution.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Iterations != 9 || metadata.OuterIterations != 0 || metadata.UsedMeasurements != 16 || metadata.Parameters != 5 || !metadata.Converged || metadata.Status != StaticPositionSolveStepTolerance || metadata.Redundancy != 11 || metadata.GeometryQuality.Tier != 3 || metadata.GeometryQuality.Rank != 5 || metadata.GeometryQuality.ConditionNumber != 12.138211809822929 || metadata.GeometryQuality.GDOP != 4.584145017556228 {
		t.Fatalf("metadata = %#v", metadata)
	}
	rejected, err := solution.RejectedSatellites(0)
	if err != nil || len(rejected) != 0 {
		t.Fatalf("rejected satellites = %#v, %v", rejected, err)
	}
	if _, err := solution.RejectedSatellites(-1); err == nil {
		t.Fatal("negative static epoch index accepted")
	}
	residuals, err := solution.Residuals()
	if err != nil || len(residuals) != 16 || residuals[0] != (StaticPositionResidual{EpochIndex: 0, SatelliteID: "G08", ResidualM: 0.0005127266049385071, BaseWeight: 0.07900945480129141, EffectiveWeight: 0.07900945480129141, RobustWeightRatio: 1}) {
		t.Fatalf("residuals = %#v, %v", residuals, err)
	}
	batchInfluence, err := solution.SatelliteBatchInfluence()
	if err != nil || len(batchInfluence) != 8 || batchInfluence[0].SatelliteID != "G08" || batchInfluence[0].OmittedMeasurements != 2 || batchInfluence[0].PositionDeltaM != [3]float64{-0.0006722398102283478, -0.00029017007909715176, -0.00044688954949378967} {
		t.Fatalf("batch influence = %#v, %v", batchInfluence, err)
	}
	satInfluence, err := solution.SatelliteInfluence()
	if err != nil || len(satInfluence) != 16 || satInfluence[0].SatelliteID != "G08" || satInfluence[0].EpochIndex != 0 || satInfluence[0].ResidualM != 0.0005127266049385071 || satInfluence[0].PositionDeltaM != [3]float64{-0.00020091980695724487, -8.672429248690605e-05, -0.0001335684210062027} {
		t.Fatalf("satellite influence = %#v, %v", satInfluence, err)
	}
	state, err := solution.StateCovarianceM2()
	if err != nil || len(state) != 25 || state[0] != 5.877755853743985 || state[24] != 5.357833503302017 || math.Float64bits(state[0]) != math.Float64bits(ecef[0]) {
		t.Fatalf("state covariance = %#v, %v", state, err)
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
	if defaults != (StaticPositionOptions{Robust: SPPRobustConfig{HuberK: 1.345, ScaleFloorM: 1, MaxOuter: 5, OuterToleranceM: 0.0001}}) {
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
	if position != [3]float64{3582103.9938147026, 532590.01012966689, 5232755.532730571} {
		t.Fatalf("broadcast position = %#v", position)
	}
	ecef, err := solution.PositionCovarianceECEFM2()
	if err != nil {
		t.Fatal(err)
	}
	if ecef != [9]float64{0.56010094372035801, 0.06789245706390809, 0.39786682612054358, 0.06789245706390809, 0.26437804716697677, 0.14208537464863213, 0.39786682612054358, 0.14208537464863213, 1.4765846379904126} {
		t.Fatalf("broadcast ECEF covariance = %#v", ecef)
	}
	clocks, err := solution.ClockBiases()
	if err != nil || len(clocks) != 6 || clocks[0] != (StaticPositionClockBias{EpochIndex: 0, System: GNSSSystemGPS, ClockS: 0.00048092782500347268}) || clocks[1] != (StaticPositionClockBias{EpochIndex: 0, System: GNSSSystemGalileo, ClockS: 0.00048092832196602794}) || clocks[2] != (StaticPositionClockBias{EpochIndex: 0, System: GNSSSystemBeiDou, ClockS: 0.00048093290113135361}) {
		t.Fatalf("broadcast clock biases = %#v, %v", clocks, err)
	}
	metadata, err := solution.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Iterations != 8 || metadata.UsedMeasurements != 48 || metadata.Parameters != 9 || !metadata.Converged || metadata.Status != StaticPositionSolveStepTolerance || metadata.Redundancy != 39 || metadata.GeometryQuality.Tier != 3 || metadata.GeometryQuality.Rank != 9 || metadata.GeometryQuality.ConditionNumber != 10.717263798595297 || metadata.GeometryQuality.GDOP != 3.1307928447329547 {
		t.Fatalf("broadcast metadata = %#v", metadata)
	}
	enu, err := solution.PositionCovarianceENUM2()
	if err != nil {
		t.Fatal(err)
	}
	if enu != [9]float64{0.25102193297322734, 0.028390470896108508, 0.080024687429818581, 0.028390470896108494, 0.47634632889543976, 0.273173416267884, 0.080024687429818553, 0.27317341626788394, 1.5736953670090803} {
		t.Fatalf("broadcast ENU covariance = %#v", enu)
	}
	influence, err := solution.EpochInfluence()
	if err != nil {
		t.Fatal(err)
	}
	if len(influence) != 2 || influence[0].OmittedMeasurements != 39 || influence[0].Status != StaticInfluenceSolved || influence[0].PositionDeltaM != [3]float64{-0.016123553737998009, 0.0040332247735932469, -0.00062527135014533997} || influence[0].ResidualRMSM != 0.62091273850711959 {
		t.Fatalf("broadcast epoch influence = %#v", influence)
	}
	geo, present, err := solution.Geodetic()
	if err != nil {
		t.Fatal(err)
	}
	if !present || geo != (Geodetic{LatitudeRad: 0.96854560893027519, LongitudeRad: 0.14759950631954899, HeightM: 59.372217842657093}) {
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
	if len(residuals) != 48 || residuals[0] != (StaticPositionResidual{EpochIndex: 0, SatelliteID: "G05", ResidualM: -0.079147819429636, BaseWeight: 0.7633750022938937, EffectiveWeight: 0.7633750022938937, RobustWeightRatio: 1}) {
		t.Fatalf("broadcast residuals = %#v", residuals)
	}
	batch, err := solution.SatelliteBatchInfluence()
	if err != nil {
		t.Fatal(err)
	}
	if len(batch) != 24 || batch[0].SatelliteID != "G05" || batch[0].OmittedMeasurements != 2 || batch[0].PositionDeltaM != [3]float64{-0.047507710754871368, 0.026683527394197881, 0.014262265525758266} || batch[0].ResidualRMSM != 0.56206807488704968 {
		t.Fatalf("broadcast batch influence = %#v", batch)
	}
	sat, err := solution.SatelliteInfluence()
	if err != nil {
		t.Fatal(err)
	}
	if len(sat) != 48 || sat[0].SatelliteID != "G05" || sat[0].EpochIndex != 0 || sat[0].ResidualM != -0.079147819429636 || sat[0].PositionDeltaM != [3]float64{-0.0099466284736990929, 0.0056296463590115309, 0.002839202992618084} {
		t.Fatalf("broadcast satellite influence = %#v", sat)
	}
	state, err := solution.StateCovarianceM2()
	if err != nil {
		t.Fatal(err)
	}
	if len(state) != 81 || state[0] != 0.56010094372035801 || state[len(state)-1] != 1.2142188450309574 {
		t.Fatalf("broadcast state covariance = %#v", state)
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
	if ecef != [9]float64{6.677160239881516e-05, 9.720380944920287e-06, -6.529210197339873e-05, 9.720380944920286e-06, 1.9831553100283693e-05, -6.207956522582059e-05, -6.529210197339872e-05, -6.207956522582057e-05, 0.00029206895833734935} {
		t.Fatalf("reference ECEF covariance = %#v", ecef)
	}
	enu, err := solution.CovarianceENU()
	if err != nil {
		t.Fatal(err)
	}
	if enu != [9]float64{1.8018814006830666e-05, -3.1382933637363e-05, -4.128819755877781e-05, -3.1382933637362997e-05, 0.00020912710500505113, 0.00013072992853920755, -4.12881975587778e-05, 0.0001307299285392075, 0.0001515261948245664} {
		t.Fatalf("reference ENU covariance = %#v", enu)
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
