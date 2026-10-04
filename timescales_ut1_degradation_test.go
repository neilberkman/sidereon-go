package sidereon

import (
	"errors"
	"strings"
	"testing"
)

func TestUT1DegradationSurvivesTimeScaleAndFrameRoutes(t *testing.T) {
	tests := []struct {
		name   string
		date   CivilDateTime
		want   uint32
		reason string
	}{
		{name: "before coverage", date: CivilDateTime{Year: 1900, Month: 1, Day: 1}, want: 1, reason: "before_coverage"},
		{name: "after coverage", date: CivilDateTime{Year: 2500, Month: 1, Day: 1}, want: 2, reason: "after_coverage"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scales, err := TimeScalesFromUTC(test.date)
			if err != nil {
				t.Fatalf("TimeScalesFromUTC(%+v): %v", test.date, err)
			}
			if scales.UT1Degraded != test.want {
				t.Fatalf("UT1Degraded = %d, want %d", scales.UT1Degraded, test.want)
			}
			_, err = GCRSToITRSMatrix(scales)
			assertUT1CoverageCause(t, err, test.reason)
			if _, err := GCRSToITRS([3]float64{1000, 2000, 3000}, scales, false); err == nil {
				t.Fatal("GCRSToITRS accepted UT1 outside table coverage")
			}
			if _, err := ComputeDopplerRangeRate([3]float64{1000, 2000, 3000}, [3]float64{1, 2, 3}, GroundStation{}, scales); err == nil {
				t.Fatal("ComputeDopplerRangeRate accepted UT1 outside table coverage")
			}
		})
	}
}

func TestUT1DegradationAcceptsNoneAndRejectsUnknown(t *testing.T) {
	scales, err := TimeScalesFromUTC(CivilDateTime{Year: 2020, Month: 6, Day: 24})
	if err != nil {
		t.Fatal(err)
	}
	if scales.UT1Degraded != 0 {
		t.Fatalf("in-coverage UT1Degraded = %d, want 0", scales.UT1Degraded)
	}
	if _, err := GCRSToITRSMatrix(scales); err != nil {
		t.Fatalf("GCRSToITRSMatrix rejected table-backed UT1: %v", err)
	}

	scales.UT1Degraded = 99
	if _, err := GCRSToITRSMatrix(scales); err == nil {
		t.Fatal("GCRSToITRSMatrix accepted an unknown UT1 degradation value")
	}
}

func assertUT1CoverageCause(t *testing.T, err error, wantReason string) {
	t.Helper()
	var status *StatusError
	if !errors.As(err, &status) || status.Detail == "" {
		t.Fatalf("expected status error with UT1 diagnostic, got %T: %v", err, err)
	}
	wantText := "UT1 table coverage"
	switch wantReason {
	case "before_coverage":
		wantText = "precedes the UT1 table coverage"
	case "after_coverage":
		wantText = "follows the UT1 table coverage"
	}
	if !strings.Contains(status.Detail, wantText) {
		t.Fatalf("frame-transform detail = %q, want UT1 coverage refusal %q", status.Detail, wantText)
	}
}
