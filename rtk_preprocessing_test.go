package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func TestRTKFloatZeroValueCloseAndNativeFailure(t *testing.T) {
	var zero RTKFloatSolution
	if err := zero.Close(); err != nil {
		t.Fatalf("zero-value Close: %v", err)
	}
	if _, err := zero.BaselineENU(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero-value BaselineENU: %v, want ErrClosed", err)
	}
	options, err := DefaultRTKFloatOptions()
	if err != nil {
		t.Fatalf("DefaultRTKFloatOptions: %v", err)
	}
	if _, err := SolveRTKFloat(RTKFloatConfig{Options: options}); err == nil {
		t.Fatal("empty RTK float configuration unexpectedly solved")
	}
	if _, err := SolveRTKFloat(RTKFloatConfig{Options: RTKFloatOptions{MaxIterations: -1}}); err == nil {
		t.Fatal("negative RTK float iteration count was accepted")
	}
}

func TestRTKPreprocessingBoundaryAndDefaults(t *testing.T) {
	options, err := CycleSlipOptionsDefault()
	if err != nil {
		t.Fatalf("CycleSlipOptionsDefault: %v", err)
	}
	if options.GFThresholdM <= 0 || options.MWThresholdCycles <= 0 || options.MinArcGapS <= 0 {
		t.Fatalf("unexpected cycle-slip defaults: %#v", options)
	}
	if _, err := SmoothCode(nil, nil, -1); err == nil {
		t.Fatal("negative hatch window was accepted")
	}
	if _, err := SmoothIonoFreeCode(nil, nil, -1); err == nil {
		t.Fatal("negative ionosphere-free hatch window was accepted")
	}
	if _, err := DetectCycleSlips([]ArcEpoch{{F1Hz: math.NaN(), F2Hz: math.NaN()}}, &options); err != nil {
		t.Fatalf("single skipped epoch: %v", err)
	}
}

func TestIonoFreeOwnershipAndNULValidation(t *testing.T) {
	_, err := CombineIonosphereFreePseudoranges(
		[]PseudorangeObservation{{SatelliteID: "G01\x00", PseudorangeM: 2e7}},
		[]PseudorangeObservation{{SatelliteID: "G01", PseudorangeM: 2e7}}, nil,
	)
	if err == nil {
		t.Fatal("embedded NUL was accepted")
	}
	var statusErr *StatusError
	if !errors.As(err, &statusErr) && err.Error() == "" {
		t.Fatalf("unexpected validation error: %v", err)
	}

	result, err := CombineIonosphereFreePseudoranges(
		[]PseudorangeObservation{{SatelliteID: "G01", PseudorangeM: 2e7}},
		[]PseudorangeObservation{{SatelliteID: "G01", PseudorangeM: 2e7}}, nil,
	)
	if err != nil {
		t.Fatalf("combine: %v", err)
	}
	if err := result.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := result.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if _, err := result.Combined(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Combined after Close = %v, want ErrClosed", err)
	}
}

func TestIonoFreeConcurrentClose(t *testing.T) {
	result, err := CombineIonosphereFreePseudoranges(
		[]PseudorangeObservation{{SatelliteID: "G01", PseudorangeM: 2e7}},
		[]PseudorangeObservation{{SatelliteID: "G01", PseudorangeM: 2e7}}, nil,
	)
	if err != nil {
		t.Fatalf("combine: %v", err)
	}
	var group sync.WaitGroup
	for i := 0; i < 4; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 50; j++ {
				_, _ = result.Combined()
			}
		}()
	}
	group.Add(1)
	go func() { defer group.Done(); _ = result.Close() }()
	group.Wait()
}
