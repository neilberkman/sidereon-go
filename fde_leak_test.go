package sidereon

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
)

func TestFDEInvalidLaterWeightDoesNotLeakEarlierCString(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	options, err := DefaultFDEOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.UnitWeights = false
	options.Weights = map[string]float64{"A": 1, "Z\x00bad": 1}
	robust := SPPRobustConfig{HuberK: 1.5, ScaleFloorM: 1, MaxOuter: 1, OuterToleranceM: 1e-3}
	for i := 0; i < 4096; i++ {
		result, err := SolveRobustFDE(sp3, SPPConfig{}, robust, options)
		if result != nil {
			_ = result.Close()
			t.Fatalf("iteration %d returned a result for invalid weights", i)
		}
		if err == nil {
			t.Fatalf("iteration %d accepted an embedded-NUL weight key", i)
		}
	}
	runtime.GC()
	debug.FreeOSMemory()

	largeValidKey := "A" + strings.Repeat("x", 1<<20)
	options.Weights = map[string]float64{largeValidKey: 1, "Z\x00bad": 1}
	before, supported, err := reviewMaxRSSBytes()
	if err != nil {
		t.Fatal(err)
	}
	if !supported {
		return
	}
	for i := 0; i < 64; i++ {
		if result, err := SolveRobustFDE(sp3, SPPConfig{}, robust, options); result != nil || err == nil {
			if result != nil {
				_ = result.Close()
			}
			t.Fatalf("large-key iteration %d returned result %v, error %v", i, result, err)
		}
	}
	runtime.GC()
	debug.FreeOSMemory()
	after, _, err := reviewMaxRSSBytes()
	if err != nil {
		t.Fatal(err)
	}
	if growth := after - before; growth > 32<<20 {
		t.Fatalf("invalid FDE weights grew native resident memory by %d bytes", growth)
	}
}
