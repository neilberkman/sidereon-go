package sidereon

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func readNavFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/nav/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func detachedFindings(t *testing.T, report *RINEXLintReport) []RINEXLintFinding {
	t.Helper()
	findings, err := report.Findings()
	if err != nil {
		t.Fatal(err)
	}
	if err := report.Close(); err != nil {
		t.Fatal(err)
	}
	return findings
}

func assertPublicFinding(t *testing.T, findings []RINEXLintFinding, want RINEXLintFinding) {
	t.Helper()
	for _, got := range findings {
		if reflect.DeepEqual(got, want) {
			return
		}
	}
	t.Fatalf("finding not present: want %+v; got %+v", want, findings)
}

func TestRINEXLintObservationEventAndHeaderFindingDetails(t *testing.T) {
	lastObs := readObservationFixture(t, "ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx")
	report, err := LintRINEXObservation(lastObs)
	if err != nil {
		t.Fatal(err)
	}
	declared := map[string]any{"year": json.Number("2020"), "month": json.Number("6"), "day": json.Number("25"), "hour": json.Number("23"), "minute": json.Number("59"), "second": json.Number("30.0")}
	observed := map[string]any{"year": json.Number("2020"), "month": json.Number("6"), "day": json.Number("25"), "hour": json.Number("0"), "minute": json.Number("0"), "second": json.Number("30.0")}
	assertPublicFinding(t, detachedFindings(t, report), RINEXLintFinding{
		Kind: "ObsTimeOfLastMismatch", SpecRef: "RINEX 3.05 Table A2, TIME OF LAST OBS",
		Details: map[string]any{"declared": declared, "declared_scale": "GPST", "observed": observed, "observed_scale": "GPST"},
		Code:    "OBS-H08", Severity: RINEXLintError, Repairable: true, HasField: true, Field: "TIME OF LAST OBS",
	})

	table := readObservationFixture(t, "rinex211_table_a7_example.rnx")
	report, err = LintRINEXObservation(table)
	if err != nil {
		t.Fatal(err)
	}
	findings := detachedFindings(t, report)
	assertPublicFinding(t, findings, RINEXLintFinding{
		Kind: "ObsEventEpoch", SpecRef: "RINEX 3.05 Table A3", Details: map[string]any{"flag": json.Number("4")},
		Code: "OBS-B07", Severity: RINEXLintInfo, HasEpochIndex: true, EpochIndex: 1,
	})
	assertPublicFinding(t, findings, RINEXLintFinding{
		Kind: "ObsEpochGap", SpecRef: "RINEX QC policy", Details: map[string]any{"gap_s": json.Number("54.0"), "interval_s": json.Number("18.0")},
		Code: "OBS-B09", Severity: RINEXLintInfo, HasEpochIndex: true, EpochIndex: 4,
	})

	event := readObservationFixture(t, "crinex_event_clocks_v3.rnx")
	report, err = LintRINEXObservation(event)
	if err != nil {
		t.Fatal(err)
	}
	findings = detachedFindings(t, report)
	assertPublicFinding(t, findings, RINEXLintFinding{
		Kind: "ObsMissingHeader", SpecRef: "RINEX 3.05/4.02 Table A2", Details: map[string]any{"label": "MARKER NAME"},
		Code: "OBS-H03", Severity: RINEXLintError, HasField: true, Field: "MARKER NAME",
	})
	assertPublicFinding(t, findings, RINEXLintFinding{
		Kind: "ObsEventEpoch", SpecRef: "RINEX 3.05 Table A3", Details: map[string]any{"flag": json.Number("5")},
		Code: "OBS-B07", Severity: RINEXLintInfo, HasEpochIndex: true, EpochIndex: 1,
	})
}

func TestRINEXLintNavigationFindingDetails(t *testing.T) {
	data := readNavFixture(t, "BRD400DLR_S_20261800000_01H_MN_trim.rnx")
	report, err := LintRINEXNav(data)
	if err != nil {
		t.Fatal(err)
	}
	findings := detachedFindings(t, report)
	assertPublicFinding(t, findings, RINEXLintFinding{
		Kind: "NavUnsortedRecords", SpecRef: "RINEX QC policy", Details: map[string]any{},
		Code: "NAV-B03", Severity: RINEXLintInfo, Repairable: true,
	})
	assertPublicFinding(t, findings, RINEXLintFinding{
		Kind: "NavUnhealthyRecords", SpecRef: "RINEX 3.05 broadcast record layout",
		Details: map[string]any{"system": "GPS", "count": json.Number("2")},
		Code:    "NAV-B05", Severity: RINEXLintInfo,
	})
	assertPublicFinding(t, findings, RINEXLintFinding{
		Kind: "NavOutOfScopeRecords", SpecRef: "RINEX QC parse-scope disclosure",
		Details: map[string]any{"class": "unsupported message CNAV", "count": json.Number("2")},
		Code:    "NAV-B06", Severity: RINEXLintInfo,
	})
}
func TestRINEXLintNavigationHeaderFindingMutations(t *testing.T) {
	data, err := os.ReadFile("testdata/nav/BRDC00GOP_R_20210010000_01D_MN.rnx")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		old, replacement string
		want             RINEXLintFinding
	}{
		{
			old: "LEAP SECONDS", replacement: "COMMENT     ",
			want: RINEXLintFinding{
				Kind: "NavLeapSecondsAbsent", SpecRef: "RINEX 3.05 Table A5", Details: map[string]any{},
				Code: "NAV-H02", Severity: RINEXLintInfo, HasField: true, Field: "LEAP SECONDS",
			},
		},
		{
			old: "7.4506e-09", replacement: "not_a_flt!",
			want: RINEXLintFinding{
				Kind: "NavIonoMalformed", SpecRef: "RINEX 3.05 Table A5",
				Details: map[string]any{"message": "bad/missing ionospheric correction field in navigation header"},
				Code:    "NAV-H03", Severity: RINEXLintWarning, HasField: true, Field: "IONOSPHERIC CORR",
			},
		},
	}
	for _, tc := range cases {
		if len(tc.old) != len(tc.replacement) || bytes.Count(data, []byte(tc.old)) != 1 {
			t.Fatalf("fixture mutation %q is not a unique equal-width field", tc.old)
		}
		mutated := bytes.Replace(data, []byte(tc.old), []byte(tc.replacement), 1)
		report, err := LintRINEXNav(mutated)
		if err != nil {
			t.Fatal(err)
		}
		assertPublicFinding(t, detachedFindings(t, report), tc.want)
	}
}
