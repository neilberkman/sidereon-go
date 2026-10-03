package sidereon

import (
	"errors"
	"math"
	"testing"
)

func TestRINEXObservationHeaderTimelineAndSkippedRecords(t *testing.T) {
	text := []byte("     3.05           OBSERVATION DATA    R                   RINEX VERSION / TYPE\n" +
		"R    1 L1C                                                  SYS / # / OBS TYPES\n" +
		"R L1C  0.25000  03 R01 R28 R00                             SYS / PHASE SHIFT\n" +
		"                                                            END OF HEADER\n")
	obs, err := ParseRINEXObservation(text)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = obs.Close() }()

	skipped, err := obs.SkippedRecords()
	if err != nil || skipped != 1 {
		t.Fatalf("skipped records = %d, %v; want 1", skipped, err)
	}
	timeline, err := obs.HeaderTimeline()
	if err != nil || len(timeline) != 1 {
		t.Fatalf("timeline = %+v, %v", timeline, err)
	}
	header, err := obs.Header()
	if err != nil {
		t.Fatal(err)
	}
	if timeline[0].FirstEpochIndex != 0 || timeline[0].Header != header {
		t.Fatalf("first timeline segment = %+v, header = %+v", timeline[0], header)
	}
	if _, err := obs.HeaderAt(0); err == nil {
		t.Fatal("HeaderAt accepted index equal to the zero epoch count")
	}
	if _, err := obs.HeaderAt(-1); err == nil {
		t.Fatal("HeaderAt accepted a negative index")
	}
}

func TestRINEXObservationDowngradeToRINEX2(t *testing.T) {
	text := []byte("     3.02           OBSERVATION DATA    C                   RINEX VERSION / TYPE\n" +
		"C    2 C1I L1I                                              SYS / # / OBS TYPES\n" +
		"                                                            END OF HEADER\n" +
		"> 2020 06 24 00 00  0.0000000  0  1\n" +
		"C01  22000000.000 7        10.00015\n" +
		"> 2020 06 24 00 00 30.0000000  6  1\n" +
		"C01                       100.000  \n")
	source, err := ParseRINEXObservation(text)
	if err != nil {
		t.Fatal(err)
	}

	first, err := source.HeaderAt(0)
	if err != nil || first.Version != 3.02 {
		t.Fatalf("HeaderAt(0) = %+v, %v", first, err)
	}
	epochs, err := source.EpochCount()
	if err != nil || epochs != 2 {
		t.Fatalf("epoch count = %d, %v", epochs, err)
	}
	if _, err := source.HeaderAt(epochs); err == nil {
		t.Fatal("HeaderAt accepted the epoch count")
	}

	downgrade, err := source.DowngradeToRINEX2(2.11)
	if err != nil {
		t.Fatal(err)
	}
	if downgrade.Observation == nil || len(downgrade.Changes) != 2 {
		t.Fatalf("downgrade = %+v", downgrade)
	}
	wantFrom := []string{"C1I", "L1I"}
	wantTo := []string{"C2I", "L2I"}
	for i, change := range downgrade.Changes {
		if change.Kind != RINEXDowngradeCodeRenamed || !change.HasSystem || !change.HasFromText || change.From != wantFrom[i] || !change.HasToText || change.To != wantTo[i] || change.HasNestedChange || change.Nested != nil {
			t.Fatalf("change %d = %+v", i, change)
		}
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	header, err := downgrade.Observation.Header()
	if err != nil || header.Version != 2.11 {
		t.Fatalf("downgraded header = %+v, %v", header, err)
	}
	if len(downgrade.Changes) != 2 || downgrade.Changes[1].To != "L2I" {
		t.Fatalf("detached changes changed after source close: %+v", downgrade.Changes)
	}
	if err := downgrade.Observation.Close(); err != nil {
		t.Fatal(err)
	}
	if err := downgrade.Observation.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRINEXObservationDowngradeTypedRefusals(t *testing.T) {
	text := []byte("     3.02           OBSERVATION DATA    G                   RINEX VERSION / TYPE\n" +
		"G    1 C1C                                                  SYS / # / OBS TYPES\n" +
		"                                                            END OF HEADER\n")
	obs, err := ParseRINEXObservation(text)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = obs.Close() }()

	for _, version := range []float64{1.99, 3.0, math.NaN(), math.Inf(1), math.Inf(-1)} {
		value, err := obs.DowngradeToRINEX2(version)
		if value.Observation != nil || value.Changes != nil {
			t.Fatalf("refused downgrade returned value: %+v", value)
		}
		var refusal *RINEXObservationWriteError
		if !errors.As(err, &refusal) || refusal.Kind != RINEXWriteErrorNotVersionTwo || !refusal.HasVersion || refusal.Message == "" {
			t.Fatalf("target %v refusal = %#v, %v", version, refusal, err)
		}
	}

	if err := obs.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := obs.HeaderTimeline(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed HeaderTimeline error = %v", err)
	}
	if _, err := obs.HeaderAt(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed HeaderAt error = %v", err)
	}
	if _, err := obs.SkippedRecords(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed SkippedRecords error = %v", err)
	}
	if _, err := obs.DowngradeToRINEX2(2.11); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed DowngradeToRINEX2 error = %v", err)
	}
}
