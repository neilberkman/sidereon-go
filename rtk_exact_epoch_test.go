package sidereon

import "testing"

func TestRTKExactEpochInputsPreserveOptionalHandles(t *testing.T) {
	prediction, err := NewExactEpoch(11, 0)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, prediction)
	gap, err := NewExactEpoch(12, 0)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, gap)
	raw := nativeRTKArcEpoch(RTKArcEpoch{PredictionEpoch: prediction})
	if raw.PredictionEpoch != prediction.handle {
		t.Fatal("raw RTK adapter dropped exact prediction epoch")
	}
	legacy := nativeRTKArcEpoch(RTKArcEpoch{HasPredictionTime: true, PredictionTimeS: 12.5})
	if legacy.PredictionEpoch != nil || !legacy.HasPredictionTime || legacy.PredictionTimeS != 12.5 {
		t.Fatalf("legacy raw RTK inputs changed: %+v", legacy)
	}

	epochs, useExact := nativeRTKDualEpochsWithExactEpochs([]RTKDualFrequencyArcEpoch{{GapEpoch: gap, PredictionEpoch: prediction}, {}})
	if !useExact || epochs[0].GapEpoch != gap.handle || epochs[0].PredictionEpoch != prediction.handle {
		t.Fatalf("dual-frequency exact epochs not preserved: %+v", epochs)
	}
	if epochs[1].GapEpoch != nil || epochs[1].PredictionEpoch != nil {
		t.Fatalf("absent exact epochs were populated: %+v", epochs[1])
	}
	legacyEpochs, useExact := nativeRTKDualEpochsWithExactEpochs([]RTKDualFrequencyArcEpoch{{HasGapTimeS: true, GapTimeS: 3.25}})
	if useExact || !legacyEpochs[0].HasGapTimeS || legacyEpochs[0].GapTimeS != 3.25 {
		t.Fatalf("legacy dual-frequency inputs changed: %+v", legacyEpochs[0])
	}
}

func TestCarrierPhaseExactGapInputPreservesHandleAndLegacyFallback(t *testing.T) {
	gap, err := NewExactEpoch(13, 0)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, gap)
	converted := nativeArcEpoch(ArcEpoch{GapTimeS: 2.75, GapEpoch: gap})
	if converted.GapEpoch != gap.handle || converted.GapTimeS != 2.75 {
		t.Fatalf("carrier-phase exact gap adapter = %+v", converted)
	}
	legacy := nativeArcEpoch(ArcEpoch{GapTimeS: 4.5})
	if legacy.GapEpoch != nil || legacy.GapTimeS != 4.5 {
		t.Fatalf("carrier-phase legacy gap adapter = %+v", legacy)
	}
}
