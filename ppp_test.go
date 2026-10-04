package sidereon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"sync"
	"testing"
)

func pppFixture(t *testing.T) (*SP3, PPPFloatConfig, PPPFixedConfig, PPPAutoInitOptions) {
	t.Helper()
	data, err := os.ReadFile("testdata/ppp-static-known-truth.sp3")
	if err != nil {
		t.Fatal(err)
	}
	manifestData, err := os.ReadFile("testdata/ppp-static-known-truth.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		FixtureSHA256 string    `json:"fixture_sha256"`
		ReceiverM     []float64 `json:"receiver_ecef_m"`
		Satellites    map[string]struct {
			PositionM []float64 `json:"position_ecef_m"`
			RangeM    float64   `json:"code_phase_range_m"`
		} `json:"satellites"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	fixtureHash := sha256.Sum256(data)
	if hex.EncodeToString(fixtureHash[:]) != manifest.FixtureSHA256 {
		t.Fatalf("known-truth SP3 SHA-256 = %x, manifest says %s", fixtureHash, manifest.FixtureSHA256)
	}
	if len(manifest.ReceiverM) != 3 || manifest.ReceiverM[0] != 4.5e6 || manifest.ReceiverM[1] != 0.5e6 || manifest.ReceiverM[2] != 4.5e6 {
		t.Fatalf("known-truth manifest receiver = %v", manifest.ReceiverM)
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
	tropo.Enabled = false
	tropo.EstimateZTD = false
	tropo.EstimateTropoGradient = false
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
		independent, ok := manifest.Satellites[satellite]
		if !ok || len(independent.PositionM) != 3 || math.IsNaN(independent.RangeM) || math.IsInf(independent.RangeM, 0) || independent.RangeM <= 0 {
			t.Fatalf("known-truth manifest row %q = %+v, present=%v", satellite, independent, ok)
		}
		state, err := sp3.State(satellite, 6)
		if err != nil {
			t.Fatal(err)
		}
		for axis := range state.PositionM {
			if math.Float64bits(state.PositionM[axis]) != math.Float64bits(independent.PositionM[axis]) {
				t.Fatalf("SP3 state %s axis %d = %.17g, independently frozen coordinate = %.17g", satellite, axis, state.PositionM[axis], independent.PositionM[axis])
			}
		}
		rangeM := independent.RangeM
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
	directSolution, err := SolvePPPFloat(sp3, floatConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := directSolution.Close(); err != nil {
			t.Errorf("direct solution Close() = %v", err)
		}
	})
	if metadata, err := directSolution.Metadata(); err != nil || !metadata.Converged {
		t.Fatalf("direct float metadata = %+v, err=%v", metadata, err)
	}
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
	floatSolvedEpochs, err := floatSolution.SolvedEpochIndices()
	if err != nil || len(floatSolvedEpochs) != 1 || floatSolvedEpochs[0] != 0 {
		t.Fatalf("float solved epoch indices = %v, err=%v", floatSolvedEpochs, err)
	}
	floatEpochClocks, err := floatSolution.EpochClocksM()
	if err != nil || len(floatEpochClocks) != len(floatSolvedEpochs) || math.IsNaN(floatEpochClocks[0]) || math.IsInf(floatEpochClocks[0], 0) {
		t.Fatalf("float epoch clocks = %v, err=%v", floatEpochClocks, err)
	}
	if removals, err := floatSolution.ResidualScreenRemovals(); err != nil || len(removals) != 0 {
		t.Fatalf("unexpected residual-screen removals = %+v, err=%v", removals, err)
	}
	if exclusions, err := floatSolution.SSRBiasExclusions(); err != nil || len(exclusions) != 0 {
		t.Fatalf("unexpected float SSR-bias exclusions = %+v, err=%v", exclusions, err)
	}
	position, err := floatSolution.Position()
	if err != nil {
		t.Fatal(err)
	}
	wantPosition := [3]float64{4.5e6, 0.5e6, 4.5e6}
	for i := range position {
		if math.IsNaN(position[i]) || math.IsInf(position[i], 0) || math.Abs(position[i]-wantPosition[i]) > 1e-6 {
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
	// The reference observations set code equal to phase and both ambiguity
	// seeds and offsets to zero, so the six exact zero-cycle ambiguities are an
	// admissible integer solution under the configured ratio test.
	if err != nil || fixedMetadata.Status != PPPSolveStateTolerance || fixedMetadata.IntegerStatus != PPPIntegerFixed || fixedMetadata.FixedAmbiguityCount != 6 {
		t.Fatalf("fixed metadata = %+v, err=%v", fixedMetadata, err)
	}
	fixedSolvedEpochs, err := fixedSolution.SolvedEpochIndices()
	if err != nil || len(fixedSolvedEpochs) != 1 || fixedSolvedEpochs[0] != 0 {
		t.Fatalf("fixed solved epoch indices = %v, err=%v", fixedSolvedEpochs, err)
	}
	fixedEpochClocks, err := fixedSolution.EpochClocksM()
	if err != nil || len(fixedEpochClocks) != len(fixedSolvedEpochs) || math.IsNaN(fixedEpochClocks[0]) || math.IsInf(fixedEpochClocks[0], 0) {
		t.Fatalf("fixed epoch clocks = %v, err=%v", fixedEpochClocks, err)
	}
	if exclusions, err := fixedSolution.SSRBiasExclusions(); err != nil || len(exclusions) != 0 {
		t.Fatalf("unexpected fixed SSR-bias exclusions = %+v, err=%v", exclusions, err)
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

func TestPPPLongAmbiguityIDCompatibility(t *testing.T) {
	sp3, config, fixedConfig, _ := pppFixture(t)
	const longID = "G08:LONG-ARC-IDENTIFIER"
	longConfig := config
	longConfig.Epochs = append([]PPPEpoch(nil), config.Epochs...)
	longConfig.Epochs[0].Observations = append([]PPPObservation(nil), config.Epochs[0].Observations...)
	longConfig.Epochs[0].Observations[0].AmbiguityID = longID
	longConfig.InitialState.AmbiguitiesM = append([]PPPFloatMapEntry(nil), config.InitialState.AmbiguitiesM...)
	longConfig.InitialState.AmbiguitiesM[0].ID = longID
	solution, err := SolvePPPFloat(sp3, longConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := solution.Close(); err != nil {
			t.Errorf("solution.Close() = %v", err)
		}
	})
	ids, err := solution.UsedIDs()
	if err != nil || len(ids) != 6 || ids[0] != longID {
		t.Fatalf("UsedIDs() = %#v, err=%v", ids, err)
	}
	if _, err := solution.UsedSatelliteIDs(); err == nil {
		t.Fatal("UsedSatelliteIDs accepted an ambiguity ID longer than the token ABI")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument {
			t.Fatalf("UsedSatelliteIDs() error = %v, want InvalidArgument", err)
		}
	}
	longFixedConfig := fixedConfig
	longFixedConfig.Epochs = append([]PPPEpoch(nil), fixedConfig.Epochs...)
	longFixedConfig.Epochs[0].Observations = append([]PPPObservation(nil), fixedConfig.Epochs[0].Observations...)
	longFixedConfig.Epochs[0].Observations[0].AmbiguityID = longID
	longFixedConfig.Ambiguity.WavelengthsM = append([]PPPFloatMapEntry(nil), fixedConfig.Ambiguity.WavelengthsM...)
	longFixedConfig.Ambiguity.WavelengthsM[0].ID = longID
	longFixedConfig.Ambiguity.OffsetsM = append([]PPPFloatMapEntry(nil), fixedConfig.Ambiguity.OffsetsM...)
	longFixedConfig.Ambiguity.OffsetsM[0].ID = longID
	fixedSolution, err := SolvePPPFixed(sp3, solution, longFixedConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixedSolution.Close(); err != nil {
			t.Errorf("fixed solution.Close() = %v", err)
		}
	})
	fixedIDs, err := fixedSolution.UsedIDs()
	if err != nil || len(fixedIDs) != 6 || fixedIDs[0] != longID {
		t.Fatalf("fixed UsedIDs() = %#v, err=%v", fixedIDs, err)
	}
	if _, err := fixedSolution.UsedSatelliteIDs(); err == nil {
		t.Fatal("fixed UsedSatelliteIDs accepted an ambiguity ID longer than the token ABI")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument {
			t.Fatalf("fixed UsedSatelliteIDs() error = %v, want InvalidArgument", err)
		}
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
