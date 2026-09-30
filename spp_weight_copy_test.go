package sidereon

import (
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestSPPSolutionConvertersDetachVarianceAndWeightArrays(t *testing.T) {
	makeNative := func() native.SPPSolution {
		return native.SPPSolution{
			UsedSatelliteCount:     2,
			UsedSatelliteIDs:       []string{"G01", "E02"},
			ResidualsM:             []float64{1, -2},
			PseudorangeVariancesM2: []float64{4, 9},
			Weights:                []float64{0.25, 1.0 / 9.0},
		}
	}
	assertDetached := func(name string, convert func(native.SPPSolution) SPPSolution) {
		t.Helper()
		input := makeNative()
		got := convert(input)
		input.PseudorangeVariancesM2[0] = 100
		input.Weights[0] = 100
		if got.PseudorangeVariancesM2[0] != 4 || got.Weights[0] != 0.25 {
			t.Fatalf("%s reused native arrays: variances=%v weights=%v", name, got.PseudorangeVariancesM2, got.Weights)
		}
		got.PseudorangeVariancesM2[1] = 200
		got.Weights[1] = 200
		if input.PseudorangeVariancesM2[1] != 9 || input.Weights[1] != 1.0/9.0 {
			t.Fatalf("%s mutation leaked to native arrays: variances=%v weights=%v", name, input.PseudorangeVariancesM2, input.Weights)
		}
	}
	assertDetached("publicSPPSolution", publicSPPSolution)
	assertDetached("fromNativeSPPSolution", fromNativeSPPSolution)
}
