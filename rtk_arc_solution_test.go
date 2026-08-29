package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func syntheticRTKArcInput(t *testing.T) ([]RTKArcEpoch, RTKArcConfig) {
	t.Helper()
	fixed := syntheticRTKFixedConfig(t)
	update, err := DefaultRTKArcUpdateOptions()
	if err != nil {
		t.Fatal(err)
	}
	config := RTKArcConfig{BaseM: fixed.BaseECEFM, ReferenceMode: RTKArcReferenceAuto, Model: fixed.Model, BaselinePriorSigmaM: 100, AmbiguityPriorSigmaM: 100, InitialBaselineM: fixed.InitialBaselineM, UpdateOptions: update}
	for _, id := range []string{"G01", "G02", "G03", "G05", "G06"} {
		config.Wavelengths = append(config.Wavelengths, RTKFloatMapEntry{ID: id, Value: 0.190293672798365})
		config.Offsets = append(config.Offsets, RTKFloatMapEntry{ID: id})
	}
	epochs := make([]RTKArcEpoch, len(fixed.Epochs))
	for epochIndex, epoch := range fixed.Epochs {
		rows := append(append([]RTKSatMeasurement(nil), epoch.References...), epoch.NonReference...)
		epochs[epochIndex].HasPredictionTime = true
		epochs[epochIndex].PredictionTimeS = float64(epochIndex * 30)
		for _, row := range rows {
			epochs[epochIndex].Base = append(epochs[epochIndex].Base, RTKArcObservation{SatelliteID: row.SatelliteID, AmbiguityID: row.SDAmbiguityID, CodeM: row.BaseCodeM, PhaseM: row.BasePhaseM})
			epochs[epochIndex].Rover = append(epochs[epochIndex].Rover, RTKArcObservation{SatelliteID: row.SatelliteID, AmbiguityID: row.SDAmbiguityID, CodeM: row.RoverCodeM, PhaseM: row.RoverPhaseM})
			epochs[epochIndex].SatellitePositions = append(epochs[epochIndex].SatellitePositions, RTKArcPositionEntry{SatelliteID: row.SatelliteID, PositionM: row.Pos})
		}
	}
	return epochs, config
}

func TestRTKArcSolutionSyntheticRoutes(t *testing.T) {
	epochs, config := syntheticRTKArcInput(t)
	solution, err := SolveRTKArc(epochs, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = solution.Close() })
	if count, err := solution.EpochCount(); err != nil || count != 3 {
		t.Fatalf("epoch count = %d, %v", count, err)
	}
	if count, err := solution.FinalEpochCount(); err != nil || count != 3 {
		t.Fatalf("final epoch count = %d, %v", count, err)
	}
	metadata, err := solution.EpochMetadata(0)
	if err != nil || metadata.UsedSatelliteCount != 5 || metadata.SDAmbiguityCount != 5 {
		t.Fatalf("metadata = %+v, %v", metadata, err)
	}
	if ambiguities, err := solution.EpochSDAmbiguities(0); err != nil || len(ambiguities) != 5 {
		t.Fatalf("ambiguities = %+v, %v", ambiguities, err)
	}
	if ids, err := solution.EpochStringIDs(0, RTKArcEpochFixedIDs); err != nil {
		t.Fatalf("fixed IDs = %v", err)
	} else if ids == nil {
		t.Fatal("fixed IDs unexpectedly nil")
	}
	if satellites, err := solution.EpochUsedSatellites(0); err != nil || len(satellites) != 5 {
		t.Fatalf("used satellites = %+v, %v", satellites, err)
	}
	if dropped, err := solution.DroppedSatellites(); err != nil {
		t.Fatalf("dropped satellites = %+v, %v", dropped, err)
	}
	if masked, err := solution.ElevationMaskedSatellites(); err != nil {
		t.Fatalf("elevation-masked satellites = %+v, %v", masked, err)
	}
	if baseline, err := solution.FinalBaseline(); err != nil || math.Abs(baseline[0]-1) > 1e-3 || math.Abs(baseline[1]+1) > 1e-3 || math.Abs(baseline[2]-2) > 1e-3 {
		t.Fatalf("final baseline = %+v, %v", baseline, err)
	}
	if covariance, err := solution.MeasurementCovariance(); err != nil || len(covariance) == 0 {
		t.Fatalf("covariance = %d, %v", len(covariance), err)
	}
	if references, err := solution.References(); err != nil || len(references) == 0 {
		t.Fatalf("references = %+v, %v", references, err)
	}
	if _, err := solution.SplitCycleSlipArcs(); err != nil {
		t.Fatalf("split arcs = %v", err)
	}
}

func TestRTKArcSolutionBoundariesAndOwnership(t *testing.T) {
	var zero RTKArcSolution
	if err := zero.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := zero.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero EpochCount = %v", err)
	}
	epochs, config := syntheticRTKArcInput(t)
	epochs[0].Base[0].SatelliteID = "G01\x00"
	if _, err := SolveRTKArc(epochs, config); err == nil {
		t.Fatal("embedded NUL was accepted")
	}
	epochs, config = syntheticRTKArcInput(t)
	config.UpdateOptions.MaxIterations = -1
	if _, err := SolveRTKArc(epochs, config); err == nil {
		t.Fatal("negative iteration count was accepted")
	}
	epochs, config = syntheticRTKArcInput(t)
	config.ReferenceMode = RTKArcReferenceMode(99)
	if _, err := SolveRTKArc(epochs, config); err == nil {
		t.Fatal("invalid reference mode was accepted")
	}
	epochs, config = syntheticRTKArcInput(t)
	solution, err := SolveRTKArc(epochs, config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := solution.EpochMetadata(-1); err == nil {
		t.Fatal("negative epoch index was accepted")
	}
	if _, err := solution.EpochMetadata(3); err == nil {
		t.Fatal("out-of-range epoch index was accepted")
	}
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.FinalBaseline(); !errors.Is(err, ErrClosed) {
		t.Fatalf("FinalBaseline after Close = %v", err)
	}
}

func TestRTKArcSolutionCloseReadRace(t *testing.T) {
	epochs, config := syntheticRTKArcInput(t)
	solution, err := SolveRTKArc(epochs, config)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 3; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 30; j++ {
				_, _ = solution.EpochCount()
				_, _ = solution.EpochMetadata(0)
				_, _ = solution.EpochUsedSatellites(0)
				_, _ = solution.MeasurementCovariance()
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		for i := 0; i < 20; i++ {
			_ = solution.Close()
		}
	}()
	group.Wait()
	_ = solution.Close()
}

func syntheticRTKDualArcEpochs(t *testing.T) []RTKDualFrequencyArcEpoch {
	t.Helper()
	fixed := syntheticRTKFixedConfig(t)
	result := make([]RTKDualFrequencyArcEpoch, len(fixed.Epochs))
	for epochIndex, epoch := range fixed.Epochs {
		rows := append(append([]RTKSatMeasurement(nil), epoch.References...), epoch.NonReference...)
		result[epochIndex] = RTKDualFrequencyArcEpoch{JDWhole: 2459000, JDFraction: float64(epochIndex) / 86400, HasEpochSortKey: true, EpochSortKey: "synthetic", HasPredictionTime: true, PredictionTimeS: float64(epochIndex * 30)}
		for _, row := range rows {
			base := RTKDualFrequencyObservation{AmbiguityID: row.SDAmbiguityID, P1M: row.BaseCodeM, P2M: row.BaseCodeM, Phi1Cycles: row.BasePhaseM / 0.190293672798365, Phi2Cycles: row.BasePhaseM / 0.244210213424568, F1Hz: 1575420000, F2Hz: 1227600000}
			rover := RTKDualFrequencyObservation{AmbiguityID: row.SDAmbiguityID, P1M: row.RoverCodeM, P2M: row.RoverCodeM, Phi1Cycles: row.RoverPhaseM / 0.190293672798365, Phi2Cycles: row.RoverPhaseM / 0.244210213424568, F1Hz: 1575420000, F2Hz: 1227600000}
			result[epochIndex].Observations = append(result[epochIndex].Observations, RTKDualFrequencySatelliteObservation{SatelliteID: row.SatelliteID, Base: base, Rover: rover})
			result[epochIndex].SatellitePositions = append(result[epochIndex].SatellitePositions, RTKArcPositionEntry{SatelliteID: row.SatelliteID, PositionM: row.Pos})
		}
	}
	return result
}

func TestRTKWideLaneAndIonosphereFreeRoutes(t *testing.T) {
	epochs := syntheticRTKDualArcEpochs(t)
	wide, err := FixWideLaneRTKArc(epochs, RTKWideLaneArcConfig{BaseM: [3]float64{4_000_000, 1_000_000, 4_500_000}, ReferenceMode: RTKArcReferenceAuto, Options: RTKWideLaneOptions{MinEpochs: 1, ToleranceCycles: 0.5}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wide.Close() })
	cycles := make([]RTKWideLaneCycle, 0, 5)
	for _, id := range []string{"G01", "G02", "G03", "G05", "G06"} {
		cycles = append(cycles, RTKWideLaneCycle{ID: id})
	}
	free, err := PrepareIonosphereFreeRTKArc(epochs, cycles, RTKIonosphereFreeArcConfig{BaseM: [3]float64{4_000_000, 1_000_000, 4_500_000}, InitialBaselineM: [3]float64{1, -1, 2}, ReferenceMode: RTKArcReferenceAuto})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = free.Close() })
	if count, err := free.EpochCount(); err != nil || count != len(epochs) {
		t.Fatalf("ionosphere-free epoch count = %d, %v", count, err)
	}
	if metadata, err := free.EpochMetadata(0); err != nil || metadata.BaseCount != 5 || metadata.RoverCount != 5 {
		t.Fatalf("ionosphere-free metadata = %+v, %v", metadata, err)
	}
	if observations, err := free.EpochBaseObservations(0); err != nil || len(observations) != 5 {
		t.Fatalf("ionosphere-free observations = %d, %v", len(observations), err)
	}
	if positions, err := free.EpochSatellitePositions(0); err != nil || len(positions) != 5 {
		t.Fatalf("ionosphere-free positions = %d, %v", len(positions), err)
	}
	if _, err := free.EpochBaseSatellitePositions(0); err != nil {
		t.Fatal(err)
	}
	if _, err := free.EpochRoverSatellitePositions(0); err != nil {
		t.Fatal(err)
	}
	if _, err := free.EpochRoverObservations(0); err != nil {
		t.Fatal(err)
	}
	if _, err := free.OffsetsM(); err != nil {
		t.Fatal(err)
	}
	if _, err := free.WavelengthsM(); err != nil {
		t.Fatal(err)
	}
	if _, err := free.References(); err != nil {
		t.Fatal(err)
	}
}

func TestRTKIonosphereFreeZeroValueAndBoundaries(t *testing.T) {
	var solution RTKIonosphereFreeArcSolution
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero EpochCount = %v", err)
	}
	if _, err := solution.EpochMetadata(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero EpochMetadata = %v", err)
	}
	epochs := syntheticRTKDualArcEpochs(t)
	config := RTKWideLaneArcConfig{ReferenceMode: RTKArcReferenceMode(9), Options: RTKWideLaneOptions{MinEpochs: 1}}
	if _, err := FixWideLaneRTKArc(epochs, config); err == nil {
		t.Fatal("invalid wide-lane reference mode was accepted")
	}
	config = RTKWideLaneArcConfig{Options: RTKWideLaneOptions{MinEpochs: -1}}
	if _, err := FixWideLaneRTKArc(epochs, config); err == nil {
		t.Fatal("negative wide-lane minimum epochs was accepted")
	}
	config = RTKWideLaneArcConfig{ReferenceSatellite: "G01\x00"}
	if _, err := FixWideLaneRTKArc(epochs, config); err == nil {
		t.Fatal("embedded NUL wide-lane reference was accepted")
	}
	if _, err := PrepareIonosphereFreeRTKArc(epochs, []RTKWideLaneCycle{{ID: "G01\x00"}}, RTKIonosphereFreeArcConfig{}); err == nil {
		t.Fatal("embedded NUL wide-lane cycle was accepted")
	}
}

func TestRTKDerivedArcCloseReadRace(t *testing.T) {
	epochs := syntheticRTKDualArcEpochs(t)
	solution, err := PrepareIonosphereFreeRTKArc(epochs, []RTKWideLaneCycle{{ID: "G01"}, {ID: "G02"}, {ID: "G03"}, {ID: "G05"}, {ID: "G06"}}, RTKIonosphereFreeArcConfig{BaseM: [3]float64{4_000_000, 1_000_000, 4_500_000}, InitialBaselineM: [3]float64{1, -1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 3; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 30; j++ {
				_, _ = solution.EpochCount()
				_, _ = solution.EpochMetadata(0)
				_, _ = solution.EpochBaseObservations(0)
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		for i := 0; i < 20; i++ {
			_ = solution.Close()
		}
	}()
	group.Wait()
	_ = solution.Close()
}
