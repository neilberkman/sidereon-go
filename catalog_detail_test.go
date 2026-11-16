package sidereon

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestCatalogValueTypes(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b any
	}{
		{"source", CatalogSourceEntry{Host: "a"}, CatalogSourceEntry{Host: "b"}},
		{"payload", CelestrakGPPayload{Kind: "tle"}, CelestrakGPPayload{Kind: "omm"}},
		{"checksum", CelestrakChecksumDetail{Kind: "missing"}, CelestrakChecksumDetail{Kind: "mismatch"}},
		{"element", CelestrakElementSetError{Kind: "invalid_input", Fields: CelestrakElementSetErrorFields{Field: fixturePtr("epoch")}}, CelestrakElementSetError{Kind: "invalid_input", Fields: CelestrakElementSetErrorFields{Field: fixturePtr("bstar")}}},
		{"record", CelestrakGPRecordData{RecordIndex: 1}, CelestrakGPRecordData{RecordIndex: 2}},
		{"rejected", CelestrakGPRejectedRecordData{RecordIndex: 1}, CelestrakGPRejectedRecordData{RecordIndex: 2}},
		{"reason", CelestrakGPRejectionReason{Kind: "tle"}, CelestrakGPRejectionReason{Kind: "omm"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.a, tc.a) || reflect.DeepEqual(tc.a, tc.b) {
				t.Fatalf("value equality: %#v %#v", tc.a, tc.b)
			}
		})
	}
	for _, tc := range []struct {
		value CelestrakGPSpecial
		want  string
	}{{CelestrakGPSpecialGPZ, "GPZ"}, {CelestrakGPSpecialGPZPlus, "GPZ-PLUS"}, {CelestrakGPSpecialDecaying, "DECAYING"}} {
		got, err := tc.value.AsStr()
		if err != nil || got != tc.want {
			t.Fatalf("special %d: %q %v", tc.value, got, err)
		}
	}
	var response *CelestrakGPResponse
	var record CelestrakGPRecord
	if response != nil || record.owner != nil {
		t.Fatal("zero handles")
	}
}

func TestCatalogReasonJSONDetails(t *testing.T) {
	cases := []struct {
		name, input string
		want        CelestrakGPRejectionReason
	}{
		{"tle", `{"kind":"tle","cause":{"kind":"invalid_field","fields":{"value":"x","reason":"bad","field":"epoch","catalog_number":1000000000,"line_label":"line 1","expected":3,"computed":4,"found":"X"}}}`, CelestrakGPRejectionReason{Kind: "tle", Cause: json.RawMessage(`{"kind":"invalid_field","fields":{"value":"x","reason":"bad","field":"epoch","catalog_number":1000000000,"line_label":"line 1","expected":3,"computed":4,"found":"X"}}`), TLE: &CelestrakTLEError{Kind: "invalid_field", Fields: CelestrakTLEErrorFields{Value: fixturePtr("x"), Reason: fixturePtr("bad"), Field: fixturePtr("epoch"), CatalogNumber: fixturePtr(json.Number("1000000000")), LineLabel: fixturePtr("line 1"), Expected: fixturePtr(json.Number("3")), Computed: fixturePtr(json.Number("4")), Found: fixturePtr("X")}}}},
		{"element", `{"kind":"invalid_element_set","cause":{"kind":"invalid_input","fields":{"field":"mean_motion","kind":"not_positive","message":"bad","code":2,"budget":42}}}`, CelestrakGPRejectionReason{Kind: "invalid_element_set", Cause: json.RawMessage(`{"kind":"invalid_input","fields":{"field":"mean_motion","kind":"not_positive","message":"bad","code":2,"budget":42}}`), ElementSet: &CelestrakElementSetError{Kind: "invalid_input", Fields: CelestrakElementSetErrorFields{Field: fixturePtr("mean_motion"), Kind: fixturePtr("not_positive"), Message: fixturePtr("bad"), Code: fixturePtr(json.Number("2")), Budget: fixturePtr(json.Number("42"))}}}},
		{"structural", `{"kind":"incomplete_record","cause":null}`, CelestrakGPRejectionReason{Kind: "incomplete_record", Cause: json.RawMessage(`null`)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got CelestrakGPRejectionReason
			if err := got.UnmarshalJSON([]byte(tc.input)); err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %+v, want %+v: %v", got, tc.want, err)
			}
		})
	}
}

func TestCatalogResponseRefusalDetails(t *testing.T) {
	cases := []struct {
		name, body, text string
		want             DataCatalogError
	}{
		{"no data", "No GP data found", "invalid argument: sidereon_parse_celestrak_gp_response: CelesTrak found no GP data", DataCatalogError{Kind: "celestrak_gp_data_not_found"}},
		{"repeat", "GP data has not updated since your last successful\ndownload of GROUP=active at 2026-03-26 08:10:22 UTC.\nData is updated once every 2 hours.\n", "invalid argument: sidereon_parse_celestrak_gp_response: CelesTrak GP data has not changed since the last successful download of GROUP=active at 2026-03-26 08:10:22 UTC", DataCatalogError{Kind: "celestrak_gp_not_modified", Query: fixturePtr("GROUP=active"), LastSuccessfulDownload: fixturePtr("2026-03-26 08:10:22 UTC")}},
		{"html", "<html><body>error</body></html>", "invalid argument: sidereon_parse_celestrak_gp_response: CelesTrak GP response is HTML", DataCatalogError{Kind: "celestrak_gp_html_response"}},
		{"empty", "", "invalid argument: sidereon_parse_celestrak_gp_response: empty CelesTrak GP response", DataCatalogError{Kind: "empty_celestrak_gp_response"}},
		{"malformed", "not json", "invalid argument: sidereon_parse_celestrak_gp_response: invalid CelesTrak GP response at line 1, column 2: malformed JSON: expected ident at line 1 column 2; response excerpt: \"not json\"", DataCatalogError{Kind: "invalid_celestrak_gp_response", Reason: "malformed JSON: expected ident at line 1 column 2; response excerpt: \"not json\"", Line: fixturePtr(uint64(1)), Column: fixturePtr(uint64(2))}},
		{"utf8", string([]byte{0xff}), "invalid argument: sidereon_parse_celestrak_gp_response: invalid CelesTrak GP response: invalid utf-8 sequence of 1 bytes from index 0", DataCatalogError{Kind: "invalid_celestrak_gp_response", Reason: "invalid utf-8 sequence of 1 bytes from index 0"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseCelestrakGPResponse([]byte(tc.body), CelestrakGPFormatJSON)
			var got *DataCatalogError
			if !errors.As(err, &got) || !got.Equal(&tc.want) || err.Error() != tc.text {
				t.Fatalf("got %#v, %v; want %#v, %q", got, err, tc.want, tc.text)
			}
		})
	}
}

func TestCatalogInvalidCodesAndText(t *testing.T) {
	cases := []struct {
		name string
		call func() error
		text string
	}{
		{"format", func() error { _, e := CelestrakGPFormat(99).AsStr(); return e }, "invalid argument: invalid CelesTrak GP format: 99"},
		{"special", func() error { _, e := CelestrakGPSpecial(99).AsStr(); return e }, "invalid argument: sidereon_celestrak_gp_special_as_str: invalid CelesTrak GP SPECIAL selector \"99\""},
		{"system", func() error { _, e := CelestrakGroup(GNSSSystem(99)); return e }, "invalid argument: sidereon_celestrak_group: invalid system GNSS system"},
		{"query", func() error { _, e := CelestrakGPURL(nil, CelestrakGPFormatJSON); return e }, "invalid argument: GP query is nil"},
		{"selector utf8", func() error {
			_, e := CelestrakGPURL(CelestrakGPName(string([]byte{0xff})), CelestrakGPFormatJSON)
			return e
		}, "invalid UTF-8 or satellite token: sidereon_celestrak_gp_url: value is not valid UTF-8"},
		{"response format", func() error { _, e := ParseCelestrakGPResponse([]byte("x"), CelestrakGPFormat(99)); return e }, "invalid argument: invalid CelesTrak GP format: 99"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil || err.Error() != tc.text {
				t.Fatalf("got %v, want %q", err, tc.text)
			}
		})
	}
	if got := (&DataCatalogError{Kind: "example"}).Error(); got != "sidereon: catalog error example" {
		t.Fatal(got)
	}
	if got := (&CelestrakGPRecordError{Kind: "example"}).Error(); got != "sidereon: GP record error example" {
		t.Fatal(got)
	}
}

func TestCatalogErrorUnwrap(t *testing.T) {
	if got := (&DataCatalogError{Kind: "x"}).Unwrap(); got != nil {
		t.Fatalf("catalog cause %v", got)
	}
	if got := (&CelestrakGPRecordError{Kind: "x"}).Unwrap(); got != nil {
		t.Fatalf("record cause %v", got)
	}
}
