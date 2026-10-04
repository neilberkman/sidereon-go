package sidereon

import (
	"errors"
	"math"
	"testing"
)

func testStandaloneTECGrid(t *testing.T) *TECGrid {
	t.Helper()
	grid, failure, err := NewTECGrid(TECGridInput{
		EpochsUnixNanos: []float64{0, 1e9}, LatitudesDeg: []float64{0, 10}, LongitudesDeg: []float64{0, 20},
		ValuesTECU: []float64{1, 0, 0, 4, 5, 6, 7, 8},
		Presence:   []bool{true, false, true, true, true, true, true, true},
	})
	if err != nil || failure != nil || grid == nil {
		t.Fatalf("NewTECGrid grid=%v failure=%+v err=%v", grid, failure, err)
	}
	closeAfterTest(t, grid)
	return grid
}

func TestTECGridRetainsAxesValuesPresenceAndInterpolationPolicy(t *testing.T) {
	grid := testStandaloneTECGrid(t)
	info, err := grid.Dimensions()
	if err != nil || info != (TECGridInfo{EpochCount: 2, LatitudeCount: 2, LongitudeCount: 2, ValueCount: 8}) {
		t.Fatalf("Dimensions=%+v err=%v", info, err)
	}
	epochs, err := grid.EpochsUnixNanos()
	if err != nil || len(epochs) != 2 || epochs[0] != 0 || epochs[1] != 1e9 {
		t.Fatalf("epochs=%v err=%v", epochs, err)
	}
	lats, err := grid.LatitudesDeg()
	if err != nil || len(lats) != 2 || lats[0] != 0 || lats[1] != 10 {
		t.Fatalf("latitudes=%v err=%v", lats, err)
	}
	lons, err := grid.LongitudesDeg()
	if err != nil || len(lons) != 2 || lons[0] != 0 || lons[1] != 20 {
		t.Fatalf("longitudes=%v err=%v", lons, err)
	}
	values, err := grid.ValuesTECU()
	if err != nil || len(values) != 8 || values[0] != 1 || !math.IsNaN(values[1]) || values[2] != 0 || values[7] != 8 {
		t.Fatalf("values=%v err=%v", values, err)
	}
	present, err := grid.ValuePresence()
	if err != nil || len(present) != 8 || !present[0] || present[1] || !present[2] {
		t.Fatalf("presence=%v err=%v", present, err)
	}
	// Explicit zero remains present, unlike the deliberately absent second node.
	if !math.IsNaN(values[1]) || !present[2] || values[2] != 0 {
		t.Fatalf("missing/zero authority lost: values=%v presence=%v", values, present)
	}

	point, failure, err := grid.VTECAtPiercePoint(0, 0, 0, IONEXMissingNodesStrict)
	if err != nil || failure != nil || !point.HasVTEC || point.ValueTECU != 1 || point.Degraded.HasGap {
		t.Fatalf("node query=%+v failure=%+v err=%v", point, failure, err)
	}
	_, failure, err = grid.VTECAtPiercePoint(500000000, 5, 10, IONEXMissingNodesStrict)
	if err != nil || failure == nil || failure.Kind != TECGridErrorNodesNotAvailable || !failure.HasGap || !failure.Gap.HasGap {
		t.Fatalf("strict query failure=%+v err=%v", failure, err)
	}
	point, failure, err = grid.VTECAtPiercePoint(500000000, 5, 10, IONEXMissingNodesRenormalize)
	// The core renormalizes the missing corner within its earlier epoch, then
	// blends that epoch with the complete later epoch: (5/3 + 13/2) / 2.
	if err != nil || failure != nil || !point.HasVTEC || math.Abs(point.ValueTECU-49.0/12.0) > 1e-12 || !point.Degraded.HasGap {
		t.Fatalf("renormalized query=%+v failure=%+v err=%v", point, failure, err)
	}
}

func TestTECGridTypedValidationAndClosedHandle(t *testing.T) {
	validGrid, failure, err := NewTECGrid(TECGridInput{EpochsUnixNanos: []float64{0, 1}, LatitudesDeg: []float64{0, 1}, LongitudesDeg: []float64{0, 1}, ValuesTECU: []float64{1, 2, 3, 4, 5, 6, 7, 8}})
	if err != nil || failure != nil {
		t.Fatalf("valid grid: failure=%+v err=%v", failure, err)
	}
	closeAfterTest(t, validGrid)
	_, failure, err = NewTECGrid(TECGridInput{EpochsUnixNanos: []float64{0, 0}, LatitudesDeg: []float64{0, 1}, LongitudesDeg: []float64{0, 1}, ValuesTECU: []float64{1, 2, 3, 4, 5, 6, 7, 8}})
	if err != nil || failure == nil || failure.Kind != TECGridErrorAxesNotIncreasing || failure.Message == "" {
		t.Fatalf("axis refusal: failure=%+v err=%v", failure, err)
	}
	_, failure, err = NewTECGrid(TECGridInput{EpochsUnixNanos: []float64{0, 1}, LatitudesDeg: []float64{0, 1}, LongitudesDeg: []float64{0, 1}, ValuesTECU: []float64{1, math.NaN(), 3, 4, 5, 6, 7, 8}})
	if err != nil || failure == nil || failure.Kind != TECGridErrorValueNotFinite || !failure.HasValueIndex || failure.ValueIndex != 1 {
		t.Fatalf("nonfinite refusal: failure=%+v err=%v", failure, err)
	}
	grid := testStandaloneTECGrid(t)
	_, failure, err = grid.VTECAtPiercePoint(0, math.NaN(), 0, IONEXMissingNodesStrict)
	if err != nil || failure == nil || failure.Kind != TECGridErrorInvalidField || !failure.HasField || failure.Field != "latitude" || !failure.HasReason || failure.Reason != "not finite" || failure.Message == "" {
		t.Fatalf("invalid query field detail: failure=%+v err=%v", failure, err)
	}
	if err := grid.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := grid.ValuesTECU(); !errors.Is(err, ErrClosed) {
		t.Fatalf("ValuesTECU after Close=%v, want ErrClosed", err)
	}
}
