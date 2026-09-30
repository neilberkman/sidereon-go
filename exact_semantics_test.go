package sidereon

import (
	"errors"
	"testing"
)

func TestExactEpochComparisonAndQueryEpoch(tester *testing.T) {
	j2000, err := ExactEpochJ2000()
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, j2000)
	components, err := j2000.Components()
	if err != nil || components != (ExactEpochComponents{}) {
		tester.Fatalf("J2000 components = %+v, %v", components, err)
	}
	attosecondsPerSecond, err := ExactAttosecondsPerSecond()
	if err != nil || attosecondsPerSecond != 1_000_000_000_000_000_000 {
		tester.Fatalf("attoseconds per second = %d, %v", attosecondsPerSecond, err)
	}
	earlier, err := NewExactEpoch(1<<53, 1)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, earlier)
	later, err := NewExactEpoch(1<<53, 2)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, later)
	ordering, err := earlier.Compare(later)
	if err != nil || ordering != ExactOrderingLess {
		tester.Fatalf("exact comparison = %v, %v; want less", ordering, err)
	}
	equal, err := earlier.Equal(later)
	if err != nil || equal {
		tester.Fatalf("distinct epochs equal = %t, %v", equal, err)
	}
	query, err := earlier.Query()
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, query)
	queryEpoch, err := query.Epoch()
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, queryEpoch)
	equal, err = queryEpoch.Equal(earlier)
	if err != nil || !equal {
		tester.Fatalf("query epoch equality = %t, %v", equal, err)
	}
	queryEqual, err := query.Equal(query)
	if err != nil || !queryEqual {
		tester.Fatalf("query self-equality = %t, %v", queryEqual, err)
	}
}

func TestStrictSSRRefusalRetainsQueryContext(tester *testing.T) {
	stateEpoch := &ExactEpochQuery{}
	selectionEpoch := &ExactEpochQuery{}
	result := SSRCorrectedState{
		StrictRefusal: true,
		Size:          SSRCorrectionSize{OrbitM: 12.5, ClockM: 0.25},
	}
	returned, err := checkedSSRCorrectionResult(result, "G07", stateEpoch, selectionEpoch)
	var refusal *SSRCorrectionSizeRefusalError
	if !errors.As(err, &refusal) {
		tester.Fatalf("refusal error = %T, %v", err, err)
	}
	if returned != result || refusal.Satellite != "G07" || refusal.StateEpoch != stateEpoch || refusal.SelectionEpoch != selectionEpoch || refusal.Size != result.Size {
		tester.Fatalf("refusal context = %+v, result %+v", refusal, returned)
	}
	result.StrictRefusal = false
	returned, err = checkedSSRCorrectionResult(result, "G07", stateEpoch, selectionEpoch)
	if err != nil || returned != result {
		tester.Fatalf("non-refusal result = %+v, %v", returned, err)
	}
}

func TestPreciseAccuracyMismatchUsesNamedTypedError(tester *testing.T) {
	_, err := BuildPreciseEphemerisSamplesWithAccuracy(
		[]PreciseEphemerisSample{{Satellite: "G01", TimeScale: GPST}}, nil,
	)
	var validation *PreciseSamplesValidationError
	if !errors.As(err, &validation) {
		tester.Fatalf("accuracy mismatch error = %T, %v", err, err)
	}
	if validation.Kind != PreciseSamplesErrorAccuracySamplesMismatch {
		tester.Fatalf("accuracy mismatch kind = %d", validation.Kind)
	}
}
