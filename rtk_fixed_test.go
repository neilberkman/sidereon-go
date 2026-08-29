package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func syntheticRTKFixedConfig(t *testing.T) RTKFixedConfig {
	t.Helper()
	model, err := DefaultRTKMeasurementModel()
	if err != nil {
		t.Fatal(err)
	}
	floatOptions, err := DefaultRTKFloatOptions()
	if err != nil {
		t.Fatal(err)
	}
	fixedOptions, err := DefaultRTKFixedOptions()
	if err != nil {
		t.Fatal(err)
	}
	residualOptions, err := DefaultRTKResidualValidationOptions()
	if err != nil {
		t.Fatal(err)
	}
	base := [3]float64{4_000_000, 1_000_000, 4_500_000}
	rover := [3]float64{4_000_001, 999_999, 4_500_002}
	satellites := [][3]float64{
		{15_600_000, 0, 20_180_000},
		{18_700_000, 2_300_000, 19_400_000},
		{21_100_000, -8_000_000, 15_300_000},
		{23_000_000, 11_000_000, 12_700_000},
		{12_000_000, 18_000_000, 21_000_000},
	}
	ids := []string{"G01", "G02", "G03", "G05", "G06"}
	epochs := make([]RTKEpoch, 3)
	for epoch := range epochs {
		rows := make([]RTKSatMeasurement, len(ids))
		for i, satellite := range satellites {
			satellite[0] += float64(epoch) * 1000
			satellite[1] += float64(epoch) * 700
			satellite[2] -= float64(epoch) * 500
			baseRange := math.Sqrt((satellite[0]-base[0])*(satellite[0]-base[0]) + (satellite[1]-base[1])*(satellite[1]-base[1]) + (satellite[2]-base[2])*(satellite[2]-base[2]))
			roverRange := math.Sqrt((satellite[0]-rover[0])*(satellite[0]-rover[0]) + (satellite[1]-rover[1])*(satellite[1]-rover[1]) + (satellite[2]-rover[2])*(satellite[2]-rover[2]))
			ambiguityID := ids[i]
			rows[i] = RTKSatMeasurement{SatelliteID: ids[i], SDAmbiguityID: ambiguityID, BaseCodeM: baseRange, BasePhaseM: baseRange, RoverCodeM: roverRange, RoverPhaseM: roverRange, BaseTXPos: satellite, RoverTXPos: satellite, Pos: satellite}
		}
		epochs[epoch] = RTKEpoch{References: rows[:1], NonReference: rows[1:], DTS: 30}
	}
	return RTKFixedConfig{
		Epochs:              epochs,
		BaseECEFM:           base,
		AmbiguityIDs:        append([]string(nil), ids[1:]...),
		AmbiguitySatellites: []RTKAmbiguitySatellite{{ID: ids[1], SatelliteID: ids[1]}, {ID: ids[2], SatelliteID: ids[2]}, {ID: ids[3], SatelliteID: ids[3]}, {ID: ids[4], SatelliteID: ids[4]}},
		Wavelengths:         []RTKFloatMapEntry{{ID: ids[1], Value: 0.190293672798365}, {ID: ids[2], Value: 0.190293672798365}, {ID: ids[3], Value: 0.190293672798365}, {ID: ids[4], Value: 0.190293672798365}},
		Offsets:             []RTKFloatMapEntry{{ID: ids[1]}, {ID: ids[2]}, {ID: ids[3]}, {ID: ids[4]}},
		Model:               model,
		FloatOptions:        floatOptions,
		FixedOptions:        fixedOptions,
		ResidualOptions:     residualOptions,
		InitialBaselineM:    [3]float64{1, -1, 2},
	}
}

func TestRTKFixedSyntheticRoutes(t *testing.T) {
	config := syntheticRTKFixedConfig(t)
	solution, err := SolveRTKFixed(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = solution.Close() })
	metadata, err := solution.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata.NObservations != 24 || metadata.FreeAmbiguityCount != 0 || metadata.FixedAmbiguityCount != 4 || metadata.UsedSatCount != 4 || !metadata.Converged || metadata.IntegerStatus != RTKIntegerFixed {
		t.Fatalf("unexpected fixed metadata: %+v", metadata)
	}
	fixedAmbiguities, err := solution.FixedAmbiguities()
	if err != nil {
		t.Fatal(err)
	}
	if len(fixedAmbiguities) != 4 {
		t.Fatalf("fixed ambiguity count = %d, want 4", len(fixedAmbiguities))
	}
	for i, want := range []string{"G02", "G03", "G05", "G06"} {
		if fixedAmbiguities[i].ID != want || fixedAmbiguities[i].Cycles != 0 {
			t.Fatalf("fixed ambiguity[%d] = %+v, want %s with zero cycles", i, fixedAmbiguities[i], want)
		}
	}
	if _, err := solution.FreeAmbiguities(); err != nil {
		t.Fatal(err)
	}
	fixedECEF, err := solution.FixedBaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	fixedENU, err := solution.FixedBaselineENU()
	if err != nil {
		t.Fatal(err)
	}
	floatECEF, err := solution.FloatBaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range [3]float64{1, -1, 2} {
		if math.Abs(fixedECEF[i]-want) > 1e-4 || math.Abs(floatECEF[i]-want) > 1e-4 {
			t.Fatalf("baseline[%d] fixed=%v float=%v, want %v", i, fixedECEF[i], floatECEF[i], want)
		}
	}
	for i, want := range [3]float64{-1.2126826726591378, 0.8146458612589099, 1.966155085108602} {
		if math.Abs(fixedENU[i]-want) > 1e-9 {
			t.Fatalf("fixed ENU[%d] = %.17g, want %.17g", i, fixedENU[i], want)
		}
	}
	ids, err := solution.UsedSatelliteIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Fatal("fixed solution has no used satellites")
	}
}

func TestRTKFixedZeroValueAndBoundaries(t *testing.T) {
	var solution RTKFixedSolution
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.Metadata(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero Metadata() error = %v, want ErrClosed", err)
	}
	if _, err := solution.FixedAmbiguities(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero FixedAmbiguities() error = %v, want ErrClosed", err)
	}
	if _, err := solution.FixedBaselineECEF(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero FixedBaselineECEF() error = %v, want ErrClosed", err)
	}
	if _, err := solution.FixedBaselineENU(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero FixedBaselineENU() error = %v, want ErrClosed", err)
	}
	if _, err := solution.FloatBaselineECEF(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero FloatBaselineECEF() error = %v, want ErrClosed", err)
	}
	if _, err := solution.FreeAmbiguities(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero FreeAmbiguities() error = %v, want ErrClosed", err)
	}
	if _, err := solution.UsedSatelliteIDs(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero UsedSatelliteIDs() error = %v, want ErrClosed", err)
	}
	config := syntheticRTKFixedConfig(t)
	config.FixedOptions.MaxIterations = -1
	if _, err := SolveRTKFixed(config); err == nil {
		t.Fatal("negative max iterations was accepted")
	}
	config = syntheticRTKFixedConfig(t)
	config.FloatOnlySystems = []GNSSSystem{GNSSSystem(99)}
	if _, err := SolveRTKFixed(config); err == nil {
		t.Fatal("invalid float-only system was accepted")
	}
	config = syntheticRTKFixedConfig(t)
	config.AmbiguityIDs[0] = "G01\x00"
	if _, err := SolveRTKFixed(config); err == nil {
		t.Fatal("embedded NUL ambiguity ID was accepted")
	}
}

func TestRTKFixedCloseReadRace(t *testing.T) {
	solution, err := SolveRTKFixed(syntheticRTKFixedConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 4; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 20; j++ {
				_, _ = solution.Metadata()
				_, _ = solution.FixedAmbiguities()
				_, _ = solution.FreeAmbiguities()
				_, _ = solution.FixedBaselineECEF()
				_, _ = solution.FixedBaselineENU()
				_, _ = solution.FloatBaselineECEF()
				_, _ = solution.UsedSatelliteIDs()
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		for i := 0; i < 8; i++ {
			_ = solution.Close()
		}
	}()
	group.Wait()
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
}
