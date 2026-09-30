package sidereon

import (
	"math"
	"math/big"
	"testing"
)

func ionexUTCWholeSecondEpoch(seconds int64) ClockEpoch {
	nanos := new(big.Int).Mul(big.NewInt(seconds), big.NewInt(1_000_000_000))
	if nanos.Sign() < 0 {
		nanos.Add(nanos, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	lowMask := new(big.Int).SetUint64(^uint64(0))
	low := new(big.Int).And(new(big.Int).Set(nanos), lowMask).Uint64()
	high := int64(new(big.Int).Rsh(new(big.Int).Set(nanos), 64).Uint64())
	return ClockEpoch{Scale: UTC, Representation: RINEXClockInstantNanos, NanosHigh: high, NanosLow: low}
}

func TestIONEXWholeUTCSecondEpochUsesExactSigned128BitNanos(t *testing.T) {
	for _, tc := range []struct {
		seconds int64
		high    int64
		low     uint64
	}{
		{-1, -1, 18446744072709551616},
		{-1 << 63, -500000000, 0},
		{1<<63 - 1, 499999999, 18446744072709551616},
		{1<<53 + 1, 488281, 4611686019427387904},
	} {
		got := ionexUTCWholeSecondEpoch(tc.seconds)
		if got.Scale != UTC || got.Representation != RINEXClockInstantNanos || got.NanosHigh != tc.high || got.NanosLow != tc.low {
			t.Errorf("whole UTC second %d encoded as %+v; expected high=%d low=%d", tc.seconds, got, tc.high, tc.low)
		}
	}
}

func TestIONEXLegacyWholeSecondMatchesExactUTCInstant(t *testing.T) {
	product := newIONEXExactTestProduct(t)
	closeAfterTest(t, product)
	seconds := ionexExactFirstMapJ2000S + 1800
	legacy, err := product.SlantDelay(0, 0, 0, 90, seconds, 1_575_420_000)
	if err != nil {
		t.Fatal(err)
	}
	instant := ionexUTCWholeSecondEpoch(seconds)
	exact, epochError, err := product.SlantDelayAtInstant(0, 0, 0, 90, instant, 1_575_420_000)
	if err != nil || epochError != nil || math.Float64bits(legacy) != math.Float64bits(exact) {
		t.Fatalf("whole-second UTC composition differs: legacy=%.17g exact=%.17g epoch=%+v err=%v", legacy, exact, epochError, err)
	}
	legacyBatch, err := product.SlantDelays([]IONEXSlantRequest{{LatDeg: 0, LonDeg: 0, AzimuthDeg: 0, ElevationDeg: 90, EpochJ2000S: seconds, FrequencyHz: 1_575_420_000}})
	if err != nil || len(legacyBatch) != 1 || math.Float64bits(legacyBatch[0]) != math.Float64bits(legacy) {
		t.Fatalf("whole-second UTC plain batch differs: delays=%+v legacy=%.17g err=%v", legacyBatch, legacy, err)
	}
	refused, err := product.SlantDelays([]IONEXSlantRequest{
		{LatDeg: 0, LonDeg: 0, AzimuthDeg: 0, ElevationDeg: 90, EpochJ2000S: seconds, FrequencyHz: 1_575_420_000},
		{LatDeg: 91, LonDeg: 0, AzimuthDeg: 0, ElevationDeg: 90, EpochJ2000S: seconds, FrequencyHz: 1_575_420_000},
	})
	if err == nil || len(refused) != 2 || math.Float64bits(refused[0]) != 0 || math.Float64bits(refused[1]) != 0 {
		t.Fatalf("plain batch did not preserve all-zero refusal output: delays=%+v err=%v", refused, err)
	}
	defaults := DefaultIONEXSlantPolicy()
	if defaults != (IONEXSlantPolicy{Coverage: IONEXCoveragePolicyStrict, MissingNodes: IONEXMissingNodesStrict, Mapping: IONEXMappingSingleLayer}) {
		t.Fatalf("default composite policy=%+v", defaults)
	}
	coveragePolicy := IONEXSlantPolicyFromCoverage(IONEXCoveragePolicyHold)
	if coveragePolicy != (IONEXSlantPolicy{Coverage: IONEXCoveragePolicyHold, MissingNodes: IONEXMissingNodesStrict, Mapping: IONEXMappingSingleLayer}) {
		t.Fatalf("coverage policy defaults=%+v", coveragePolicy)
	}
	policy := NewIONEXSlantPolicy(IONEXCoveragePolicyStrict, IONEXMissingNodesStrict, IONEXMappingSingleLayer)
	owned, err := product.SlantDelayResultsAtInstantsOwned([]IONEXInstantSlantRequest{{LatDeg: 0, LonDeg: 0, AzimuthDeg: 0, ElevationDeg: 90, Epoch: instant, FrequencyHz: 1_575_420_000}}, policy)
	if err != nil || len(owned.Rows) != 1 || !owned.Rows[0].IsOK || math.Float64bits(owned.Rows[0].Evaluation.DelayM) != math.Float64bits(legacy) {
		t.Fatalf("whole-second UTC owned batch composition differs: rows=%+v err=%v", owned.Rows, err)
	}
	legacyEvaluation, err := product.SlantDelayWithPolicy(0, 0, 0, 90, seconds, 1_575_420_000, IONEXCoveragePolicyStrict)
	if err != nil {
		t.Fatal(err)
	}
	row, err := product.SlantDelayAtInstantWithPolicy(0, 0, 0, 90, instant, 1_575_420_000, policy)
	if err != nil || !row.IsOK || row.EpochError != nil || math.Float64bits(legacyEvaluation.DelayM) != math.Float64bits(row.Evaluation.DelayM) {
		t.Fatalf("composite UTC policy composition differs: legacy=%+v exact=%+v err=%v", legacyEvaluation, row, err)
	}
}
