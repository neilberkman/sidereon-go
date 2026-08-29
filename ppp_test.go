package sidereon

import (
	"errors"
	"math"
	"os"
	"sync"
	"testing"
)

func pppFixture(t *testing.T) (*SP3, PPPFloatConfig, PPPFixedConfig, PPPAutoInitOptions) {
	t.Helper()
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Errorf("SP3.Close() = %v", err)
		}
	})
	auto, err := DefaultPPPAutoInitOptions()
	if err != nil {
		t.Fatal(err)
	}
	weights, err := DefaultPPPMeasurementWeights()
	if err != nil {
		t.Fatal(err)
	}
	tropo, err := DefaultPPPTroposphereOptions()
	if err != nil {
		t.Fatal(err)
	}
	floatOptions, err := DefaultPPPFloatOptions()
	if err != nil {
		t.Fatal(err)
	}
	fixedOptions, err := DefaultPPPFixAmbiguityOptions()
	if err != nil {
		t.Fatal(err)
	}
	receiver := [3]float64{4.5e6, 0.5e6, 4.5e6}
	satellites := []string{"G08", "G10", "G16", "G18", "G20", "G21"}
	observations := make([]PPPObservation, len(satellites))
	ambiguities := make([]PPPFloatMapEntry, len(satellites))
	wavelengths := make([]PPPFloatMapEntry, len(satellites))
	offsets := make([]PPPFloatMapEntry, len(satellites))
	for i, satellite := range satellites {
		state, err := sp3.State(satellite, 1)
		if err != nil {
			t.Fatal(err)
		}
		dx := state.PositionM[0] - receiver[0]
		dy := state.PositionM[1] - receiver[1]
		dz := state.PositionM[2] - receiver[2]
		rangeM := math.Sqrt(dx*dx + dy*dy + dz*dz)
		observations[i] = PPPObservation{SatelliteID: satellite, AmbiguityID: satellite, CodeM: rangeM, PhaseM: rangeM, Frequency1Hz: 1575420000, Frequency2Hz: 1227600000}
		ambiguities[i] = PPPFloatMapEntry{ID: satellite}
		wavelengths[i] = PPPFloatMapEntry{ID: satellite, Value: 0.19029367279836487}
		offsets[i] = PPPFloatMapEntry{ID: satellite}
	}
	epoch := PPPEpoch{Civil: CivilDateTime{Year: 2020, Month: 6, Day: 24, Hour: 12}, TRxJ2000S: 646272000, Observations: observations}
	floatConfig := PPPFloatConfig{Epochs: []PPPEpoch{epoch}, InitialState: PPPFloatState{PositionM: receiver, ClocksM: []float64{0}, AmbiguitiesM: ambiguities}, Weights: weights, Troposphere: tropo, Options: floatOptions}
	fixedConfig := PPPFixedConfig{Epochs: []PPPEpoch{epoch}, Weights: weights, Troposphere: tropo, Options: floatOptions, Ambiguity: PPPFixedAmbiguityOptions{WavelengthsM: wavelengths, OffsetsM: offsets, RatioThreshold: fixedOptions.RatioThreshold}}
	return sp3, floatConfig, fixedConfig, auto
}

func TestPPPDeterministicSP3Fixture(t *testing.T) {
	sp3, floatConfig, fixedConfig, auto := pppFixture(t)
	zeroStateConfig := floatConfig
	zeroStateConfig.InitialState = PPPFloatState{}
	zeroStateSolution, err := SolvePPPAutoInitFloat(sp3, zeroStateConfig, auto)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := zeroStateSolution.Close(); err != nil {
			t.Errorf("zero-state solution Close() = %v", err)
		}
	})
	floatSolution, err := SolvePPPAutoInitFloat(sp3, floatConfig, auto)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := floatSolution.Close(); err != nil {
			t.Errorf("float solution Close() = %v", err)
		}
	})
	metadata, err := floatSolution.Metadata()
	if err != nil || !metadata.Converged || metadata.Status != PPPSolveStateTolerance || metadata.AmbiguityCount != 6 || metadata.UsedSatCount != 6 {
		t.Fatalf("float metadata = %+v, err=%v", metadata, err)
	}
	position, err := floatSolution.Position()
	if err != nil {
		t.Fatal(err)
	}
	wantPosition := [3]float64{4629603.470231021, 413720.87281577487, 4428197.195378103}
	for i := range position {
		if math.Abs(position[i]-wantPosition[i]) > 1e-6 {
			t.Fatalf("float position[%d] = %.17g, want %.17g", i, position[i], wantPosition[i])
		}
	}
	ambiguities, err := floatSolution.Ambiguities()
	if err != nil || len(ambiguities) != 6 {
		t.Fatalf("float ambiguities = %#v, err=%v", ambiguities, err)
	}
	ids, err := floatSolution.UsedIDs()
	if err != nil || len(ids) != 6 || ids[0] != "G08" {
		t.Fatalf("float IDs = %#v, err=%v", ids, err)
	}
	satelliteIDs, err := floatSolution.UsedSatelliteIDs()
	if err != nil || len(satelliteIDs) != 6 || satelliteIDs[0] != "G08" {
		t.Fatalf("float satellite IDs = %#v, err=%v", satelliteIDs, err)
	}
	if _, err := floatSolution.PositionCovariances(); err != nil {
		t.Fatal(err)
	}
	if _, err := floatSolution.TemporalCorrelation(); err != nil {
		t.Fatal(err)
	}
	if _, err := floatSolution.TropoGradient(); err != nil {
		t.Fatal(err)
	}

	fixedSolution, err := SolvePPPFixed(sp3, floatSolution, fixedConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixedSolution.Close(); err != nil {
			t.Errorf("fixed solution Close() = %v", err)
		}
	})
	fixedMetadata, err := fixedSolution.Metadata()
	if err != nil || fixedMetadata.Status != PPPSolveStateTolerance || fixedMetadata.IntegerStatus != PPPIntegerNotFixed || fixedMetadata.FixedAmbiguityCount != 6 {
		t.Fatalf("fixed metadata = %+v, err=%v", fixedMetadata, err)
	}
	if _, err := fixedSolution.Position(); err != nil {
		t.Fatal(err)
	}
	if _, err := fixedSolution.FloatPosition(); err != nil {
		t.Fatal(err)
	}
	if values, err := fixedSolution.FixedAmbiguities(); err != nil || len(values) != 6 {
		t.Fatalf("fixed ambiguities = %#v, err=%v", values, err)
	}
	if _, err := fixedSolution.UsedIDs(); err != nil {
		t.Fatal(err)
	}
	if _, err := fixedSolution.UsedSatelliteIDs(); err != nil {
		t.Fatal(err)
	}
	if _, err := fixedSolution.PositionCovariances(); err != nil {
		t.Fatal(err)
	}
	if _, err := fixedSolution.TemporalCorrelation(); err != nil {
		t.Fatal(err)
	}
	if _, err := fixedSolution.TropoGradient(); err != nil {
		t.Fatal(err)
	}

	correctionObservations := make([]PPPObservationCorrection, len(floatConfig.Epochs[0].Observations))
	for i, observation := range floatConfig.Epochs[0].Observations {
		correctionObservations[i] = PPPObservationCorrection{SatelliteID: observation.SatelliteID, Frequency1Hz: observation.Frequency1Hz, Frequency2Hz: observation.Frequency2Hz}
	}
	corrections, err := BuildPPPCorrections(sp3, []PPPCorrectionEpoch{{Epoch: floatConfig.Epochs[0].Civil, TRxJ2000S: floatConfig.Epochs[0].TRxJ2000S, Observations: correctionObservations}}, [3]float64{4.5e6, 0.5e6, 4.5e6}, PPPCorrectionsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := corrections.Close(); err != nil {
			t.Errorf("corrections Close() = %v", err)
		}
	})
	if values, err := corrections.CodeBias(); err != nil || len(values) != 0 {
		t.Fatalf("code-bias corrections = %#v, err=%v", values, err)
	}
	if values, err := corrections.OceanLoading(); err != nil || len(values) != 0 {
		t.Fatalf("ocean corrections = %#v, err=%v", values, err)
	}
	if values, err := corrections.PoleTide(); err != nil || len(values) != 0 {
		t.Fatalf("pole-tide corrections = %#v, err=%v", values, err)
	}
	if values, err := corrections.SatPCOECEF(); err != nil || len(values) != 0 {
		t.Fatalf("PCO corrections = %#v, err=%v", values, err)
	}
	if values, err := corrections.SatPCV(); err != nil || len(values) != 0 {
		t.Fatalf("PCV corrections = %#v, err=%v", values, err)
	}
	if values, err := corrections.Tide(); err != nil || len(values) != 0 {
		t.Fatalf("tide corrections = %#v, err=%v", values, err)
	}
	if values, err := corrections.Windup(); err != nil || len(values) != 0 {
		t.Fatalf("windup corrections = %#v, err=%v", values, err)
	}
	autoFixed, err := SolvePPPAutoInitFixed(sp3, floatConfig, fixedConfig, auto)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := autoFixed.Close(); err != nil {
			t.Errorf("auto fixed solution Close() = %v", err)
		}
	})
}

func TestPPPInvalidShapesAndText(t *testing.T) {
	sp3, config, _, _ := pppFixture(t)
	invalid := config
	invalid.Troposphere.Mapping = PPPTropoMapping(99)
	if _, err := SolvePPPFloat(sp3, invalid); err == nil {
		t.Fatal("invalid PPP mapping accepted")
	}
	invalid = config
	invalid.InitialState.ClocksM = nil
	if _, err := SolvePPPFloat(sp3, invalid); err == nil {
		t.Fatal("mismatched PPP clocks accepted")
	}
	invalid = config
	invalid.Options.MaxIterations = -1
	if _, err := SolvePPPFloat(sp3, invalid); err == nil {
		t.Fatal("negative PPP iteration count accepted")
	}
	invalid = config
	invalid.Epochs[0].Observations[0].SatelliteID = "G08\x00suffix"
	if _, err := SolvePPPFloat(sp3, invalid); err == nil {
		t.Fatal("embedded-NUL PPP satellite ID accepted")
	}
	invalid = config
	invalid.Troposphere.VMFSamples = make([]PPPVmfSiteSample, 9)
	invalid.Troposphere.Mapping = PPPTropoMappingVMF1
	if _, err := SolvePPPFloat(sp3, invalid); err == nil {
		t.Fatal("oversized VMF sample series accepted")
	}
	correction := PPPCorrectionEpoch{Epoch: config.Epochs[0].Civil, TRxJ2000S: config.Epochs[0].TRxJ2000S, Observations: []PPPObservationCorrection{{SatelliteID: "G08\x00suffix", Frequency1Hz: 1575420000, Frequency2Hz: 1227600000}}}
	if _, err := BuildPPPCorrections(sp3, []PPPCorrectionEpoch{correction}, [3]float64{}, PPPCorrectionsOptions{}); err == nil {
		t.Fatal("embedded-NUL correction satellite ID accepted")
	}
	if _, err := BuildPPPCorrections(sp3, nil, [3]float64{}, PPPCorrectionsOptions{CodeBiasSystemPairs: []PPPCodeBiasSystemPair{{System: GNSSSystem(99)}}}); err == nil {
		t.Fatal("invalid PPP code-bias GNSS system accepted")
	}
}

func TestPPPEmptyNestedInputs(t *testing.T) {
	sp3, config, _, _ := pppFixture(t)
	emptyEpoch := PPPCorrectionEpoch{Epoch: config.Epochs[0].Civil, TRxJ2000S: config.Epochs[0].TRxJ2000S}
	corrections, err := BuildPPPCorrections(sp3, []PPPCorrectionEpoch{emptyEpoch}, config.InitialState.PositionM, PPPCorrectionsOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := corrections.Close(); err != nil {
			t.Errorf("corrections.Close() = %v", err)
		}
	})
	if _, err := corrections.CodeBias(); err != nil {
		t.Fatal(err)
	}
}

func TestPPPSolutionsCloseReadRace(t *testing.T) {
	sp3, config, _, auto := pppFixture(t)
	solution, err := SolvePPPAutoInitFloat(sp3, config, auto)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := solution.Close(); err != nil {
			t.Errorf("solution.Close() = %v", err)
		}
	})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 32; j++ {
				_, _ = solution.Metadata()
				_, _ = solution.Position()
				_, _ = solution.Ambiguities()
			}
		}()
	}
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
	group.Wait()
	if _, err := solution.Position(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Position after Close = %v, want ErrClosed", err)
	}
}
