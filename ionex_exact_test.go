package sidereon

import (
	"math"
	"reflect"
	"testing"
)

const ionexExactFirstMapJ2000S int64 = 631108800

func ionexExactClock(scale TimeScale, j2000Nanos int64) ClockEpoch {
	return ClockEpoch{Scale: scale, Representation: RINEXClockInstantNanos, NanosLow: uint64(j2000Nanos)}
}

func newIONEXExactTestProduct(t *testing.T) *IONEX {
	t.Helper()
	value, err := NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale:       UTC,
		MapEpochsJ2000S: []float64{float64(ionexExactFirstMapJ2000S), float64(ionexExactFirstMapJ2000S + 3600)},
		LatNodesDeg:     []float64{40, -40}, LonNodesDeg: []float64{-20, 20}, DLatDeg: -80, DLonDeg: 40,
		ShellHeightKm: 450, BaseRadiusKm: 6371, Exponent: 0,
		TECMAPsTECU: []float64{10, 10, 10, 10, 20, 20, 20, 20},
	})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestIONEXExactInstantFractionScaleBatchAndOwnership(t *testing.T) {
	product := newIONEXExactTestProduct(t)
	t.Cleanup(func() { _ = product.Close() })
	first := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000)
	last := ionexExactClock(UTC, (ionexExactFirstMapJ2000S+3600)*1_000_000_000)
	middle := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000+1_800_500_000_000)
	gpstEquivalent := ionexExactClock(GPST, ionexExactFirstMapJ2000S*1_000_000_000+1_818_500_000_000)
	frequency := 1_575_420_000.0
	query := func(epoch ClockEpoch) float64 {
		value, epochError, err := product.SlantDelayAtInstant(0, 0, 0, 90, epoch, frequency)
		if err != nil || epochError != nil {
			t.Fatalf("exact scalar epoch error=%+v err=%v", epochError, err)
		}
		return value
	}
	d0, d1, dm := query(first), query(last), query(middle)
	want := d0 + (d1-d0)*(1800.5/3600.0)
	if math.Abs(dm-want) > math.Max(1e-12, math.Abs(want)*1e-12) {
		t.Fatalf("fractional interpolation %.17g, independent linear reference %.17g", dm, want)
	}
	requests := []IONEXInstantSlantRequest{
		{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: middle, FrequencyHz: frequency},
		{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: gpstEquivalent, FrequencyHz: frequency},
	}
	out := make([]IONEXInstantSlantResult, len(requests))
	if err := product.SlantDelayResultsAtInstants(requests, out, DefaultIONEXSlantPolicy()); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || !out[0].IsOK || !out[1].IsOK || math.Float64bits(out[0].Evaluation.DelayM) != math.Float64bits(out[1].Evaluation.DelayM) {
		t.Fatalf("UTC/GPST batch results=%+v", out)
	}
	if math.Float64bits(out[0].Evaluation.DelayM) != math.Float64bits(dm) {
		t.Fatalf("batch delay bits differ from exact scalar: %.17g %.17g", out[0].Evaluation.DelayM, dm)
	}
	tdb := ionexExactClock(TDB, ionexExactFirstMapJ2000S*1_000_000_000+1_800_500_000_000)
	epochRows := make([]IONEXInstantSlantResult, 1)
	epochErr := product.SlantDelayResultsAtInstants([]IONEXInstantSlantRequest{{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: tdb, FrequencyHz: frequency}}, epochRows, DefaultIONEXSlantPolicy())
	if epochErr != nil || len(epochRows) != 1 || epochRows[0].IsOK || epochRows[0].EpochError == nil || epochRows[0].EpochError.Kind != IONEXEpochErrorNoExactUTCOffset || epochRows[0].Error != nil {
		t.Fatalf("typed TDB epoch refusal=%+v err=%v", epochRows, epochErr)
	}
	owned, err := product.SlantDelayResultsAtInstantsOwned(requests, DefaultIONEXSlantPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(owned.Rows) != 2 || owned.Rows[0].Message != "" {
		t.Fatalf("owned successful rows=%+v", owned.Rows)
	}
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(owned.Rows[0].Evaluation, out[0].Evaluation) {
		t.Fatal("owned row changed after source close")
	}
}

func TestIONEXExactInstantKeepsStrictFractionalBoundaryAndSelection(t *testing.T) {
	product := newIONEXExactTestProduct(t)
	defer product.Close()
	after := ionexExactClock(UTC, (ionexExactFirstMapJ2000S+3600)*1_000_000_000+500_000_000)
	row, err := product.SlantDelayAtInstantWithPolicy(0, 0, 0, 90, after, 1_575_420_000, DefaultIONEXSlantPolicy())
	if err == nil || row.IsOK || row.EpochError != nil || row.Error == nil || row.Error.Kind != IONEXSlantErrorOutOfCoverage || row.Error.CoverageError != IONEXCoverageErrorEpochAfterLastMap {
		t.Fatalf("strict fractional boundary row=%+v err=%v", row, err)
	}
	owned, ownedErr := product.SlantDelayResultsAtInstantsOwned([]IONEXInstantSlantRequest{{LatDeg: 0, LonDeg: 0, ElevationDeg: 90, Epoch: after, FrequencyHz: 1_575_420_000}}, DefaultIONEXSlantPolicy())
	if ownedErr != nil || len(owned.Rows) != 1 || owned.Rows[0].Error == nil || owned.Rows[0].Error.Kind != IONEXSlantErrorOutOfCoverage || owned.Rows[0].Message == "" {
		t.Fatalf("owned strict refusal=%+v err=%v", owned, ownedErr)
	}
	requested := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000+1_800_500_000_000)
	selected, metadata, epochError, err := SelectIONEXAtInstant([]*IONEX{product}, requested, StalenessPolicyDefault())
	if err != nil || selected == nil || epochError != nil || metadata.RequestedEpochJ2000S != float64(ionexExactFirstMapJ2000S)+1800.5 {
		t.Fatalf("exact selection=%v metadata=%+v epoch=%+v err=%v", selected, metadata, epochError, err)
	}
	defer selected.Close()
	start := ionexExactClock(UTC, ionexExactFirstMapJ2000S*1_000_000_000+1_800_250_000_000)
	end := ionexExactClock(GPST, ionexExactFirstMapJ2000S*1_000_000_000+1_836_750_000_000)
	rangeProduct, rangeMetadata, epochError, err := SelectIONEXOverInstantRange([]*IONEX{product}, start, end, StalenessPolicyDefault())
	if err != nil || rangeProduct == nil || epochError != nil || rangeMetadata.RequestedEpochJ2000S != float64(ionexExactFirstMapJ2000S)+1818.75 {
		t.Fatalf("exact range selection=%v metadata=%+v epoch=%+v err=%v", rangeProduct, rangeMetadata, epochError, err)
	}
	defer rangeProduct.Close()
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	if len(owned.Rows) != 1 || owned.Rows[0].Error == nil || owned.Rows[0].Message == "" {
		t.Fatalf("owned strict refusal lost details after source close: %+v", owned.Rows)
	}
}
