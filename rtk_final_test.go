package sidereon

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"
)

func TestRTKFinalConfigDefaultsAndZeroValues(t *testing.T) {
	static, err := DefaultRTKRINEXStaticBaselineConfig()
	if err != nil {
		t.Fatal(err)
	}
	if static.ArcOptions.MinCommonSatellites == 0 || static.FloatOptions.MaxIterations == 0 || static.FixedOptions.MaxIterations == 0 {
		t.Fatalf("unexpected static defaults: %+v", static)
	}
	wide, err := DefaultRTKRINEXWideLaneFixedConfig()
	if err != nil {
		t.Fatal(err)
	}
	if wide.ArcOptions.MinCommonSatellites == 0 || wide.FloatOptions.MaxIterations == 0 || wide.FixedOptions.MaxIterations == 0 {
		t.Fatalf("unexpected wide-lane defaults: %+v", wide)
	}
	static.ReferenceMode = RTKArcReferenceMode(3)
	if _, err := SolveStaticRINEXRTKBaseline(nil, nil, nil, static); err == nil {
		t.Fatal("invalid reference mode was accepted")
	}
	wide.ReferenceMode = RTKArcReferenceMode(3)
	if _, err := SolveWideLaneFixedRINEXRTKBaseline(nil, nil, nil, wide); err == nil {
		t.Fatal("invalid wide-lane reference mode was accepted")
	}
	var staticSolution RTKStaticArcSolution
	if err := staticSolution.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := staticSolution.FixedMetadata(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero static metadata error = %v", err)
	}
	var wideSolution RTKWideLaneFixedRINEXSolution
	if err := wideSolution.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := wideSolution.Metadata(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero wide metadata error = %v", err)
	}
}

func TestRTKFinalStaticFixtureAndCloseRace(t *testing.T) {
	sp3, base, rover := rinexRTKFixture(t)
	config, err := DefaultRTKRINEXStaticBaselineConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.BaseM = [3]float64{3582105.291, 532589.7313, 5232754.8054}
	config.ArcOptions.HasMaxEpochs = true
	config.ArcOptions.MaxEpochs = 2
	config.ArcOptions.MinCommonSatellites = 4
	config.ArcOptions.SignalPairs = []RTKRINEXSignalPair{{System: GNSSSystemGPS, CodeObservable: "C1C", PhaseObservable: "L1C"}}
	solution, err := SolveStaticRINEXRTKBaseline(sp3, base, rover, config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = solution.Close() }()
	fixedMetadata, err := solution.FixedMetadata()
	if err != nil {
		t.Fatal(err)
	}
	ambiguityIDs, err := solution.AmbiguityIDs()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := solution.AmbiguitySatellites(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.ElevationMaskedSatellites(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.FixedAmbiguities(); err != nil {
		t.Fatal(err)
	}
	fixedBaseline, err := solution.FixedBaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := solution.FixedFreeAmbiguities(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.FloatAmbiguities(); err != nil {
		t.Fatal(err)
	}
	floatMetadata, err := solution.FloatMetadata()
	if err != nil {
		t.Fatal(err)
	}
	geometry, err := solution.GeometryQuality()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := solution.References(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.SplitCycleSlipArcs(); err != nil {
		t.Fatal(err)
	}
	floatBaseline, err := solution.FloatBaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ambiguityIDs, []string{"G08", "G18", "G27"}) {
		t.Fatalf("static ambiguity ids = %v", ambiguityIDs)
	}
	if fixedMetadata.Iterations != 1 || fixedMetadata.NObservations != 12 || fixedMetadata.FixedAmbiguityCount != 3 || fixedMetadata.ResidualCount != 6 || fixedMetadata.UsedSatCount != 3 || !fixedMetadata.Converged {
		t.Fatalf("static fixed metadata = %+v", fixedMetadata)
	}
	if floatMetadata.Iterations != 1 || floatMetadata.NObservations != 12 || floatMetadata.AmbiguityCount != 3 || floatMetadata.ResidualCount != 6 || floatMetadata.UsedSatCount != 3 || !floatMetadata.Converged {
		t.Fatalf("static float metadata = %+v", floatMetadata)
	}
	if fixedBaseline != [3]float64{} || floatBaseline != [3]float64{} {
		t.Fatalf("static baselines = fixed %v float %v", fixedBaseline, floatBaseline)
	}
	if geometry.Tier != 3 || geometry.Redundancy != 6 || geometry.Rank != 6 || math.Abs(geometry.ConditionNumber-11.883787660307695) > 1e-12 || math.Abs(geometry.GDOP-0.2825426946625203) > 1e-12 || !geometry.RAIMCheckable || !geometry.CovarianceValidated {
		t.Fatalf("static geometry = %+v", geometry)
	}
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		for i := 0; i < 20; i++ {
			_, _ = solution.FixedMetadata()
			_, _ = solution.DroppedSatellites()
		}
	}()
	go func() {
		defer group.Done()
		for i := 0; i < 10; i++ {
			_ = solution.Close()
		}
	}()
	group.Wait()
}

func TestRTKFinalWideLaneFixture(t *testing.T) {
	sp3, base, rover := rinexRTKFixture(t)
	config, err := DefaultRTKRINEXWideLaneFixedConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.BaseM = [3]float64{3582105.291, 532589.7313, 5232754.8054}
	config.ArcOptions.HasMaxEpochs = true
	config.ArcOptions.MaxEpochs = 2
	config.ArcOptions.MinCommonSatellites = 4
	config.ArcOptions.SignalPairs = []RTKRINEXDualSignalPair{{System: GNSSSystemGPS, Code1Observable: "C1C", Phase1Observable: "L1C", Code2Observable: "C2W", Phase2Observable: "L2W"}}
	solution, err := SolveWideLaneFixedRINEXRTKBaseline(sp3, base, rover, config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = solution.Close() }()
	metadata, err := solution.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	fixedBaseline, err := solution.FixedBaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	fixedMetadata, err := solution.FixedMetadata()
	if err != nil {
		t.Fatal(err)
	}
	floatMetadata, err := solution.FloatMetadata()
	if err != nil {
		t.Fatal(err)
	}
	floatBaseline, err := solution.FloatBaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	cycles, err := solution.WideLaneCycles()
	if err != nil {
		t.Fatal(err)
	}
	if metadata != (RTKWideLaneFixedRINEXMetadata{WideLaneFixed: true, WideLaneAmbiguityCount: 3}) {
		t.Fatalf("wide metadata = %+v", metadata)
	}
	if fixedMetadata.FixedAmbiguityCount != 3 || fixedMetadata.NObservations != 12 || fixedMetadata.ResidualCount != 6 || !fixedMetadata.Converged || floatMetadata.AmbiguityCount != 3 || floatMetadata.NObservations != 12 || floatMetadata.ResidualCount != 6 || !floatMetadata.Converged {
		t.Fatalf("wide solution metadata = fixed %+v float %+v", fixedMetadata, floatMetadata)
	}
	if fixedBaseline != [3]float64{} || floatBaseline != [3]float64{} {
		t.Fatalf("wide baselines = fixed %v float %v", fixedBaseline, floatBaseline)
	}
	if !reflect.DeepEqual(cycles, []RTKWideLaneCycle{{ID: "G08", Cycles: 0}, {ID: "G18", Cycles: 0}, {ID: "G27", Cycles: 0}}) {
		t.Fatalf("wide-lane cycles = %v", cycles)
	}
}

func TestRTKFinalStaticRawFixture(t *testing.T) {
	sp3, base, rover := rinexRTKFixture(t)
	arc, err := BuildRINEXRTKArc(sp3, base, rover, &RTKRINEXArcOptions{HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = arc.Close() }()
	model, err := DefaultRTKMeasurementModel()
	if err != nil {
		t.Fatal(err)
	}
	update, err := DefaultRTKArcUpdateOptions()
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
	wavelengths, err := arc.WavelengthsM()
	if err != nil {
		t.Fatal(err)
	}
	offsets, err := arc.OffsetsM()
	if err != nil {
		t.Fatal(err)
	}
	epochs := make([]RTKArcEpoch, 2)
	for i := range epochs {
		baseObservations, e := arc.EpochBaseObservations(i)
		if e != nil {
			t.Fatal(e)
		}
		roverObservations, e := arc.EpochRoverObservations(i)
		if e != nil {
			t.Fatal(e)
		}
		positions, e := arc.EpochSatellitePositions(i)
		if e != nil {
			t.Fatal(e)
		}
		entries := make([]RTKArcPositionEntry, len(positions))
		for j, position := range positions {
			entries[j] = RTKArcPositionEntry(position)
		}
		epochs[i] = RTKArcEpoch{Base: baseObservations, Rover: roverObservations, SatellitePositions: entries}
	}
	solution, err := SolveStaticRTKArc(epochs, RTKStaticArcConfig{Arc: RTKArcConfig{BaseM: [3]float64{3582105.291, 532589.7313, 5232754.8054}, Model: model, Wavelengths: wavelengths, Offsets: offsets, UpdateOptions: update}, FloatOptions: floatOptions, FixedOptions: fixedOptions, ResidualOptions: residualOptions})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = solution.Close() }()
	metadata, err := solution.FloatMetadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Iterations != 1 || metadata.NObservations != 12 || metadata.AmbiguityCount != 3 || metadata.ResidualCount != 6 || metadata.UsedSatCount != 3 || !metadata.Converged || metadata.GeometryQuality.Tier != 3 || metadata.GeometryQuality.Redundancy != 6 || metadata.GeometryQuality.Rank != 6 || math.Abs(metadata.GeometryQuality.ConditionNumber-11.888200237918669) > 1e-12 || math.Abs(metadata.GeometryQuality.GDOP-0.2825120571884736) > 1e-12 {
		t.Fatalf("raw static float metadata = %+v", metadata)
	}
}
