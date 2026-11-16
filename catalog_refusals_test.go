package sidereon

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCatalogTLERefusalVariants(t *testing.T) {
	fixture, err := os.ReadFile("testdata/celestrak/25544.tle")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(fixture)), "\n")
	a, b := strings.TrimSpace(lines[1]), strings.TrimSpace(lines[2])
	edit := func(s string, i, j int, v string) string { return s[:i] + v + s[j:] }
	checksum := func(s string) string {
		sum := 0
		for _, c := range s[:68] {
			if c >= '0' && c <= '9' {
				sum += int(c - '0')
			}
			if c == '-' {
				sum++
			}
		}
		return s[:68] + string(rune('0'+sum%10))
	}
	cases := []struct{ name, a, b, kind, cause string }{
		{"bad length", a[:60], b, "tle", `{"kind":"format","fields":{}}`},
		{"bad line number", edit(a, 0, 1, "3"), b, "invalid_tle_structure", ""},
		{"nonascii", edit(a, 8, 9, "é"), b, "tle", `{"kind":"non_ascii","fields":{}}`},
		{"mismatch", a, checksum(edit(b, 2, 7, "25545")), "invalid_tle_structure", ""},
		{"catalog", checksum(edit(a, 2, 7, "I1234")), checksum(edit(b, 2, 7, "I1234")), "tle", `{"kind":"invalid_catalog_number","fields":{"value":"I1234","reason":"invalid Alpha-5 leading letter"}}`},
		{"invalid field", checksum(edit(a, 18, 20, "XX")), b, "tle", `{"kind":"invalid_field","fields":{"field":"epoch_year","reason":"must be two digits"}}`},
		{"numeric field", a, checksum(edit(b, 8, 16, " XX.6328")), "tle", `{"kind":"field","fields":{"value":"invalid float \"XX.6328\""}}`},
		{"checksum not digit", a, edit(b, 68, 69, "X"), "tle", `{"kind":"checksum_not_digit","fields":{"line_label":"line 2","found":"X","computed":2}}`},
		{"checksum mismatch", a, edit(b, 68, 69, "3"), "tle", `{"kind":"checksum_mismatch","fields":{"line_label":"line 2","expected":3,"computed":2}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.a + "\n" + tc.b + "\n" + a + "\n" + b
			response, err := ParseCelestrakGPResponse([]byte(body), CelestrakGPFormatTwoLine)
			if err != nil {
				t.Fatal(err)
			}
			cleanupClose(t, "response", response.Close)
			accepted, rejected, err := response.Counts()
			if err != nil || accepted != 1 || rejected != 1 {
				t.Fatalf("counts %d %d %v", accepted, rejected, err)
			}
			record, err := response.RejectedRecord(0)
			if err != nil {
				t.Fatal(err)
			}
			got, err := record.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			wantLines := map[string]uint64{"bad length": 1, "bad line number": 1, "nonascii": 1, "mismatch": 1, "catalog": 1, "invalid field": 1, "numeric field": 2, "checksum not digit": 2, "checksum mismatch": 2}
			if got.RecordIndex != 1 || got.LineNumber == nil || *got.LineNumber != wantLines[tc.name] || got.Reason.Kind != tc.kind || string(got.Reason.Cause) != tc.cause || string(got.Raw) != tc.a+"\n"+tc.b+"\n" {
				encoded, _ := json.Marshal(got.Reason)
				t.Fatalf("rejection index=%d line=%v raw=%q reason=%s", got.RecordIndex, got.LineNumber, got.Raw, encoded)
			}
		})
	}
}

func TestCatalogOMMRefusalVariants(t *testing.T) {
	data, err := os.ReadFile("testdata/celestrak/25544.kvn")
	if err != nil {
		t.Fatal(err)
	}
	fixture := strings.ReplaceAll(string(data), "\r\n", "\n")
	cases := []struct {
		name, old, replacement, kind, cause string
		line                                uint64
	}{
		{"missing", "EPOCH          = 2026-06-17T04:32:52.099296\n", "", "omm", `{"kind":"missing_field","fields":{"field":"EPOCH"}}`, 1},
		{"alternatives", "MEAN_MOTION    = 15.49273435\n", "", "omm", `{"kind":"missing_one_of","fields":{"fields":["SEMI_MAJOR_AXIS","MEAN_MOTION"]}}`, 1},
		{"invalid element", "ECCENTRICITY   = .0004737", "ECCENTRICITY   = 2", "invalid_element_set", `{"kind":"invalid_input","fields":{"field":"eccentricity","kind":"out_of_range"}}`, 1},
		{"invalid field", "INCLINATION    = 51.6332", "INCLINATION    = abc", "omm", `{"kind":"invalid_field","fields":{"field":"INCLINATION","kind":"float_parse"}}`, 15},
		{"epoch", "EPOCH          = 2026-06-17T04:32:52.099296", "EPOCH          = nonsense", "omm", `{"kind":"epoch","fields":{"message":"invalid seconds in \"nonsense\""}}`, 12},
		{"duplicate", "INCLINATION    = 51.6332", "INCLINATION    = 51.6332\nINCLINATION = 10", "omm", `{"kind":"duplicate_field","fields":{"field":"INCLINATION","first":"51.6332","second":"10"}}`, 16},
		{"unknown", "INCLINATION    = 51.6332", "UNKNOWN_KEY = bad\nINCLINATION    = 51.6332", "omm", `{"kind":"unknown_field","fields":{"field":"UNKNOWN_KEY"}}`, 1},
		{"malformed line", "INCLINATION    = 51.6332", "bad line\nINCLINATION    = 51.6332", "omm", `{"kind":"malformed_line","fields":{"line":15,"text":"bad line"}}`, 15},
		{"unit", "INCLINATION    = 51.6332", "INCLINATION    = 51.6332 [km]", "omm", `{"kind":"unit_mismatch","fields":{"field":"INCLINATION","unit":"km","expected":"deg"}}`, 15},
		{"metadata", "CENTER_NAME    = EARTH", "CENTER_NAME    = MARS", "invalid_element_set", `{"kind":"invalid_input","fields":{"field":"CENTER_NAME","kind":"out_of_range"}}`, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := strings.Replace(fixture, tc.old, tc.replacement, 1)
			if body == fixture {
				t.Fatal("fixture edit not applied")
			}
			r, err := ParseCelestrakGPResponse([]byte(body), CelestrakGPFormatKVN)
			if err != nil {
				t.Fatal(err)
			}
			cleanupClose(t, "response", r.Close)
			a, b, err := r.Counts()
			if err != nil || a != 0 || b != 1 {
				t.Fatalf("counts %d %d %v", a, b, err)
			}
			record, err := r.RejectedRecord(0)
			if err != nil {
				t.Fatal(err)
			}
			got, err := record.Snapshot()
			if err != nil || got.RecordIndex != 1 || got.LineNumber == nil || *got.LineNumber != tc.line || got.Reason.Kind != tc.kind || string(got.Reason.Cause) != tc.cause || string(got.Raw) != body {
				t.Fatalf("rejection %+v, %v", got, err)
			}
		})
	}
}

func TestCatalogMalformedDocumentLocation(t *testing.T) {
	_, err := ParseOMMJSONArray([]byte("["))
	var engine *EngineError
	if !errors.As(err, &engine) || engine.Kind != "malformed_document" || string(engine.Fields) != `{"format":"JSON","message":"EOF while parsing a list at line 1 column 1","byte_offset":0}` {
		t.Fatalf("OMM malformed document: %v", err)
	}
	_, err = ParseCelestrakGPResponse([]byte("["), CelestrakGPFormatJSON)
	var detail *DataCatalogError
	want := DataCatalogError{Kind: "invalid_celestrak_gp_response", Reason: "malformed JSON: EOF while parsing a list at line 1 column 1; response excerpt: \"[\"", Line: fixturePtr(uint64(1)), Column: fixturePtr(uint64(1))}
	if !errors.As(err, &detail) || !detail.Equal(&want) {
		t.Fatalf("GP malformed document: %+v %v", detail, err)
	}
}

func TestCatalogStructuralRefusalVariants(t *testing.T) {
	data, err := os.ReadFile("testdata/celestrak/25544.tle")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	a, b := strings.TrimSpace(lines[1]), strings.TrimSpace(lines[2])
	cases := []struct {
		name            string
		format          CelestrakGPFormat
		body, raw, kind string
		index, line     uint64
	}{
		{"unexpected", CelestrakGPFormatTwoLine, a + "\n" + b + "\nSTRAY\n" + a + "\n" + b, "STRAY\n", "unexpected_content", 2, 3},
		{"incomplete", CelestrakGPFormatTLE, "ISS\n" + a + "\n" + b + "\nNEXT SAT", "NEXT SAT", "incomplete_record", 2, 4},
		{"wrong line", CelestrakGPFormatTwoLine, "3" + a[1:] + "\n" + b + "\n" + a + "\n" + b, "3" + a[1:] + "\n" + b + "\n", "invalid_tle_structure", 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ParseCelestrakGPResponse([]byte(tc.body), tc.format)
			if err != nil {
				t.Fatal(err)
			}
			cleanupClose(t, "response", r.Close)
			record, err := r.RejectedRecord(0)
			if err != nil {
				t.Fatal(err)
			}
			got, err := record.Snapshot()
			if err != nil || got.RecordIndex != tc.index || got.LineNumber == nil || *got.LineNumber != tc.line || got.Reason.Kind != tc.kind || got.Reason.Cause != nil || string(got.Raw) != tc.raw {
				t.Fatalf("rejection %+v: %v", got, err)
			}
		})
	}
}

func TestCatalogOMMDocumentRejectionVariants(t *testing.T) {
	fixture, err := os.ReadFile("testdata/celestrak/25544.json")
	if err != nil {
		t.Fatal(err)
	}
	valid := strings.TrimSpace(string(fixture))
	cases := []struct {
		name, body, raw, cause string
		format                 CelestrakGPFormat
		line                   uint64
	}{
		{"csv columns", "OBJECT_NAME,EPOCH,MEAN_MOTION\nISS,2026-01-01T00:00:00\n", "OBJECT_NAME,EPOCH,MEAN_MOTION\nISS,2026-01-01T00:00:00\n", `{"kind":"csv_column_count","fields":{"found":2,"expected":3}}`, CelestrakGPFormatCSV, 2},
		{"json element", "[42," + valid[1:], "[42,", `{"kind":"field","fields":{"message":"expected a JSON object"}}`, CelestrakGPFormatJSON, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ParseCelestrakGPResponse([]byte(tc.body), tc.format)
			if err != nil {
				t.Fatal(err)
			}
			cleanupClose(t, "response", r.Close)
			record, err := r.RejectedRecord(0)
			if err != nil {
				t.Fatal(err)
			}
			got, err := record.Snapshot()
			if err != nil || got.RecordIndex != 1 || got.LineNumber == nil || *got.LineNumber != tc.line || got.Reason.Kind != "omm" || string(got.Reason.Cause) != tc.cause || string(got.Raw) != tc.raw {
				t.Fatalf("rejection %+v %v", got, err)
			}
		})
	}
}
