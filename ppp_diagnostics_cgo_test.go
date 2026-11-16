//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"testing"
)

func TestPPPVarianceFactorEstimated(t *testing.T) {
	sp3, cfg, fixed, _ := pppFixture(t)
	float, err := SolvePPPFloat(sp3, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = float.Close() }()
	// Six satellites give 12 code/phase rows and 10 states (position, clock, six ambiguities).
	if estimated, err := float.PosteriorVarianceFactorEstimated(); err != nil || !estimated {
		t.Fatalf("float estimated=%v / %v", estimated, err)
	}
	solution, err := SolvePPPFixed(sp3, float, fixed)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = solution.Close() }()
	if estimated, err := solution.PosteriorVarianceFactorEstimated(); err != nil || !estimated {
		t.Fatalf("fixed estimated=%v / %v", estimated, err)
	}
	// Four satellites give eight rows and eight float states: no residual variance estimate.
	cfg.Epochs[0].Observations = cfg.Epochs[0].Observations[:4]
	cfg.InitialState.AmbiguitiesM = cfg.InitialState.AmbiguitiesM[:4]
	exactly, err := SolvePPPFloat(sp3, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = exactly.Close() }()
	if estimated, err := exactly.PosteriorVarianceFactorEstimated(); err != nil || estimated {
		t.Fatalf("exact estimated=%v / %v", estimated, err)
	}
	if factor, err := exactly.PosteriorVarianceFactor(); err != nil || factor != 1 {
		t.Fatalf("a-priori scale=%g / %v", factor, err)
	}
}
