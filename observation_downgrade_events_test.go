package sidereon

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRINEXObservationDowngradePreservesNestedEventListsAndSource(t *testing.T) {
	header := func(body, label string) string { return fmt.Sprintf("%-60s%s", body, label) }
	obsRecord := func(satellite string, values ...float64) string {
		line := satellite
		for _, value := range values {
			line += fmt.Sprintf("%14.3f  ", value)
		}
		return strings.TrimRight(line, " ")
	}
	blankEvent := fmt.Sprintf(">%30s%1d%3d", "", 4, 1)
	lines := []string{
		header("     3.05           OBSERVATION DATA    M (MIXED)", "RINEX VERSION / TYPE"),
		header("G    2 C1C L1C", "SYS / # / OBS TYPES"),
		header("R    1 C1C", "SYS / # / OBS TYPES"),
		header("", "END OF HEADER"),
		"> 2020 01 01 00 00  0.0000000  0  2",
		obsRecord("G01", 20000000, 100000),
		obsRecord("R02", 21000000),
		blankEvent,
		header("G    3 L1C C1C S1C", "SYS / # / OBS TYPES"),
		"> 2020 01 01 00 00 30.0000000  0  2",
		obsRecord("G01", 100030, 20000030, 45),
		obsRecord("R02", 21000030),
	}
	obs, err := ParseRINEXObservation([]byte(strings.Join(lines, "\n")))
	if err != nil {
		t.Fatalf("parse event-list fixture: %v", err)
	}
	before, err := obs.RINEXText()
	if err != nil {
		t.Fatal(err)
	}
	result, err := obs.DowngradeToRINEX2(2.11)
	if err != nil || result.Observation == nil {
		t.Fatalf("event-list downgrade = %+v, %v", result, err)
	}
	after, err := obs.RINEXText()
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("source changed after downgrade: err=%v", err)
	}

	var nested *RINEXObservationDowngradeChange
	var rewritten *RINEXObservationDowngradeChange
	for i := range result.Changes {
		change := &result.Changes[i]
		switch change.Kind {
		case RINEXDowngradeInEventLists:
			if change.HasEpochIndex && change.EpochIndex == 1 && change.HasNestedChange {
				nested = change.Nested
			}
		case RINEXDowngradeEventRecordsRewritten:
			if change.HasEpochIndex && change.EpochIndex == 1 {
				rewritten = change
			}
		}
	}
	if nested == nil || !nested.HasSystem || nested.System != GNSSSystemGLONASS || nested.Kind != RINEXDowngradeCodeAdded || !nested.HasCode || nested.Code != "S1C" {
		t.Fatalf("nested event-list change = %+v", nested)
	}
	if rewritten == nil || len(rewritten.FromRecords) == 0 || len(rewritten.ToRecords) == 0 {
		t.Fatalf("rewritten event records = %+v", rewritten)
	}
	for _, record := range rewritten.ToRecords {
		if !strings.Contains(record, "# / TYPES OF OBSERV") {
			t.Fatalf("rewritten record order/content = %q", record)
		}
	}
	if err := obs.Close(); err != nil {
		t.Fatal(err)
	}
	if epochs, err := result.Observation.EpochCount(); err != nil || epochs != 3 {
		t.Fatalf("independent downgraded product epoch count = %d, %v; want 3 (including flag-4 event)", epochs, err)
	}
	if err := result.Observation.Close(); err != nil {
		t.Fatal(err)
	}
}
