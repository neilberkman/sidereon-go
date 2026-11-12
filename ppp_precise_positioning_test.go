package sidereon

import (
	"errors"
	"testing"
)

func TestPPPStrategyConstantsAndValueEquality(t *testing.T) {
	strategy, err := DefaultPPPAutoInitStrategy()
	if err != nil || strategy != PPPAutoInitReference {
		t.Fatalf("default strategy = %v, %v", strategy, err)
	}
	if PPPVMFSiteMaxSamples != 8 || PPPVMFSiteSingleSampleClampDays != 0.25 {
		t.Fatalf("VMF constants = %d, %g", PPPVMFSiteMaxSamples, PPPVMFSiteSingleSampleClampDays)
	}
	auto, err := DefaultPPPAutoInitOptions()
	if err != nil {
		t.Fatal(err)
	}
	equal, err := auto.Equal(auto)
	if err != nil || !equal {
		t.Fatalf("auto equality = %v, %v", equal, err)
	}
	changed := auto
	changed.SPPPressureHPA++
	equal, err = auto.Equal(changed)
	if err != nil || equal {
		t.Fatalf("auto inequality = %v, %v", equal, err)
	}
	weights, err := DefaultPPPMeasurementWeights()
	if err != nil {
		t.Fatal(err)
	}
	equal, err = weights.Equal(weights)
	if err != nil || !equal {
		t.Fatalf("weights equality = %v, %v", equal, err)
	}
	options, err := DefaultPPPFloatOptions()
	if err != nil {
		t.Fatal(err)
	}
	equal, err = options.Equal(options)
	if err != nil || !equal {
		t.Fatalf("options equality = %v, %v", equal, err)
	}
	lookup, err := DefaultPPPCorrectionLookup()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := lookup.Close(); closeErr != nil {
			t.Errorf("close lookup: %v", closeErr)
		}
	})
	info, err := lookup.Info()
	if err != nil {
		t.Fatal(err)
	}
	if info != (PPPCorrectionLookupInfo{}) {
		t.Fatalf("default lookup info = %#v", info)
	}
	equal, err = lookup.Equal(lookup)
	if err != nil || !equal {
		t.Fatalf("lookup equality = %v, %v", equal, err)
	}
}

func TestPPPAutoInitStrategyRefusal(t *testing.T) {
	_, err := SolvePPPAutoInitFloatWithStrategy(nil, PPPFloatConfig{}, PPPAutoInitOptions{}, PPPAutoInitStrategy(2))
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("nil SP3 error = %v", err)
	}
}
