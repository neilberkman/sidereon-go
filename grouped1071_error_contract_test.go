package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"sidereon.dev/go/v3/internal/native"
	"testing"
)

type groupedErrorCase struct {
	kind   string
	fields map[string]any
}

func groupedReturnedStatus(t *testing.T, family EngineErrorFamily, kind string, fields map[string]any, display string) *StatusError {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"schema_version": 1, "family": family.Name(), "operation": "contract_test", "error": map[string]any{"kind": kind, "fields": fields}})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := native.DecodeSchema1EnginePayload(native.EngineErrorFamily(family), payload)
	if err != nil {
		t.Fatalf("decode %s/%s payload: %v", family, kind, err)
	}
	status, ok := publicError(&native.StatusError{Code: 2, Text: "invalid argument", Detail: display, Engine: decoded}).(*StatusError)
	if !ok {
		t.Fatalf("public status type = %T", publicError(&native.StatusError{Code: 2, Text: "invalid argument", Detail: display, Engine: decoded}))
	}
	if status.Detail != display || status.Error() != "invalid argument: "+display {
		t.Fatalf("same-call display = %q, want %q", status.Error(), display)
	}
	return status
}

func assertReturnedEngineFields(t *testing.T, got *EngineError, want map[string]any) {
	t.Helper()
	if got == nil || got.CaptureError != nil {
		t.Fatalf("engine error = %#v", got)
	}
	var fields map[string]any
	if err := got.UnmarshalFields(&fields); err != nil {
		t.Fatalf("unmarshal exact fields: %v", err)
	}
	gotJSON, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotJSON, wantJSON) {
		t.Fatalf("returned fields = %s, want %s", gotJSON, wantJSON)
	}
}

func TestGrouped1071TDMNativeFailureRoute(t *testing.T) {
	_, err := ParseTDMKVN([]byte("NOT A LINE\n"))
	if err == nil {
		t.Fatal("invalid TDM input unexpectedly succeeded")
	}
	var status *StatusError
	if !errors.As(err, &status) || status.EngineError() == nil {
		t.Fatalf("native TDM status = %T %v", err, err)
	}
	detail := status.TDMError()
	if detail == nil || detail.Kind != "malformed_line" || len(detail.Fields) != 2 || detail.Display == "" {
		t.Fatalf("native TDM detail = %#v", detail)
	}
}

func TestGrouped1071TDMDetailPublicRoute(t *testing.T) {
	cases := []groupedErrorCase{
		{"no_segments", map[string]any{}},
		{"section", map[string]any{"line": 5, "detail": "unexpected section"}},
		{"malformed_line", map[string]any{"line": 12, "text": "NOT A LINE"}},
		{"non_printable_character", map[string]any{"line": 8, "keyword": "COMMENT", "column": 4, "character": "\a"}},
		{"line_too_long", map[string]any{"line": nil, "keyword": "DATA", "length": 255}},
		{"malformed_epoch", map[string]any{"line": 10, "keyword": "RECEIVE_FREQ", "text": "bad-epoch"}},
		{"records_out_of_order", map[string]any{"segment": 1, "keyword": "RANGE", "epoch": "2026-01-01T00:00:00"}},
		{"duplicate_record", map[string]any{"segment": 2, "keyword": "DOPPLER", "epoch": "2026-01-01T00:00:00"}},
		{"unterminated_final_line", map[string]any{"line": 99}},
		{"unwritable", map[string]any{"keyword": "COMMENT", "reason": "non-ascii"}},
		{"keyword_out_of_order", map[string]any{"line": nil, "keyword": "MODE", "section": "header"}},
		{"undefined_participant", map[string]any{"segment": 1, "keyword": "PATH", "index": 3}},
		{"conflicting_keyword", map[string]any{"line": 6, "keyword": "TIME_SYSTEM", "section": "metadata", "first": "UTC", "second": "TAI"}},
		{"repeated_keyword", map[string]any{"line": 7, "keyword": "START_TIME", "section": "metadata"}},
		{"undefined_keyword", map[string]any{"line": 22, "keyword": "UNKNOWN_KW", "section": "header"}},
		{"missing_keyword", map[string]any{"keyword": "TIME_SYSTEM", "segment": 1}},
		{"empty_data_section", map[string]any{"segment": 1}},
		{"empty_value", map[string]any{"line": nil, "keyword": "PARTICIPANT_2"}},
		{"invalid_version", map[string]any{"line": 1, "value": "3.0"}},
		{"keyword_not_assignable", map[string]any{"keyword": "DATA_START"}},
		{"malformed_record", map[string]any{"line": 33, "keyword": "RECEIVE_FREQ"}},
		{"invalid_field", map[string]any{"keyword": "TRANSMIT_FREQ", "kind": "not_positive"}},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			display := "sidereon_tdm_parse_kvn: exact display for " + tc.kind
			status := groupedReturnedStatus(t, EngineErrorFamilyTdm, tc.kind, tc.fields, display)
			detail := status.TDMError()
			if detail == nil || detail.Kind != tc.kind || detail.Display != display || detail.Error() != display {
				t.Fatalf("typed TDM detail = %#v", detail)
			}
			if status.EngineError() == nil || status.EngineError().Family != EngineErrorFamilyTdm || status.EngineError().Kind != tc.kind {
				t.Fatalf("generic C transport changed: %#v", status.EngineError())
			}
			assertReturnedEngineFields(t, status.EngineError(), tc.fields)
			var found *TDMErrorDetail
			if !errors.As(status, &found) || found.Kind != detail.Kind || found.Display != detail.Display {
				t.Fatalf("errors.As TDM detail = %#v, want %#v", found, detail)
			}
		})
	}
	if (&StatusError{Engine: &EngineError{Family: EngineErrorFamilyCatalog}}).TDMError() != nil {
		t.Fatal("non-TDM family exposed a TDM detail")
	}
}

func TestGrouped1071CatalogDetailPublicRoute(t *testing.T) {
	cases := []groupedErrorCase{
		{"unknown_center", map[string]any{"center": "bad-center"}},
		{"unknown_product_type", map[string]any{"product_type": "bad-type"}},
		{"unsupported_product", map[string]any{"center": "igs", "product_type": "ionex"}},
		{"unsupported_distribution", map[string]any{"source": "cddis", "product_type": "ionex"}},
		{"unsupported_product_era", map[string]any{"center": "igs", "product_type": "sp3", "year": 2026, "month": 1, "day": 2, "date": "2026-01-02"}},
		{"unsupported_distribution_era", map[string]any{"source": "cddis", "center": "igs", "product_type": "sp3", "year": 2026, "month": 1, "day": 2, "date": "2026-01-02"}},
		{"no_distribution_sources", map[string]any{}},
		{"invalid_official_filename", map[string]any{"filename": "bad..name"}},
		{"inconsistent_product_identity", map[string]any{"field": "center"}},
		{"no_open_mirror", map[string]any{"center": "igs", "product_type": "ionex"}},
		{"invalid_date", map[string]any{"year": 2026, "month": 13, "day": 1}},
		{"date_out_of_range", map[string]any{}},
		{"date_before_gps_epoch", map[string]any{"year": 1979, "month": 12, "day": 31, "date": "1979-12-31"}},
		{"invalid_gps_day_of_week", map[string]any{"day_of_week": 7}},
		{"invalid_sample", map[string]any{"sample": "99X"}},
		{"unsupported_sample", map[string]any{"center": "igs", "product_type": "sp3", "sample": "99X"}},
		{"invalid_span", map[string]any{"span": "99D"}},
		{"invalid_issue", map[string]any{"issue": "9999"}},
		{"missing_issue", map[string]any{"center": "igs"}},
		{"unexpected_issue", map[string]any{"center": "igs"}},
		{"unsupported_issue", map[string]any{"center": "igs", "issue": "9999"}},
		{"invalid_date_time", map[string]any{"hour": 24, "minute": 0, "second": 0}},
		{"no_ultra_issue", map[string]any{}},
		{"no_available_ultra_issue", map[string]any{}},
		{"unsupported_nominal_schedule", map[string]any{"center": "igs", "product_type": "sp3"}},
		{"unrecognized_archive_listing", map[string]any{"reason": "unknown listing"}},
		{"invalid_station", map[string]any{"station": "BADSTATION"}},
		{"invalid_coordinate", map[string]any{"lat_deg_bits": "0x7ff8000000000000", "lon_deg_bits": "0x7ff8000000000000", "lat_deg": map[string]any{"decimal": "NaN", "bits_hex": "7ff8000000000000"}, "lon_deg": map[string]any{"decimal": "NaN", "bits_hex": "7ff8000000000000"}, "latitude_deg": map[string]any{"decimal": "NaN", "bits_hex": "7ff8000000000000"}, "longitude_deg": map[string]any{"decimal": "NaN", "bits_hex": "7ff8000000000000"}}},
		{"invalid_tile_index", map[string]any{"lat_index": 91, "lon_index": -181}},
		{"invalid_tile_id", map[string]any{"tile_id": "invalid_tile"}},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			status := groupedReturnedStatus(t, EngineErrorFamilyCatalog, tc.kind, tc.fields, "catalog: "+tc.kind)
			if status.EngineError() == nil || status.EngineError().Kind != tc.kind {
				t.Fatalf("catalog EngineError = %#v", status.EngineError())
			}
			assertReturnedEngineFields(t, status.EngineError(), tc.fields)
			if status.TDMError() != nil {
				t.Fatal("catalog error was exposed as TDM")
			}
		})
	}
}

func TestGrouped1071CoreErrorNestedReturnContract(t *testing.T) {
	cases := []groupedErrorCase{
		{"parse", map[string]any{"message": "bad input"}},
		{"unknown_satellite", map[string]any{"satellite_id": "G03"}},
		{"missing_glonass_channel", map[string]any{}},
		{"missing_terrain_tile", map[string]any{"lat_index": 45, "lon_index": -73}},
		{"unknown_terrain_elevation", map[string]any{"lat_index": 45, "lon_index": -73, "latitude_posting": 4, "longitude_posting": 8}},
		{"non_wgs84_terrain_tile", map[string]any{"lat_index": 45, "lon_index": -73, "datum": map[string]any{"kind": "other", "value": "ED50"}}},
		{"terrain_tile", map[string]any{"lat_index": 45, "lon_index": -73, "cause": map[string]any{"kind": "io", "fields": map[string]any{"path": "/tiles/n45.dt2", "message": "permission denied"}}}},
		{"terrain_tile_origin", map[string]any{"path": "/tiles/n45.dt2", "lat_index": 45, "lon_index": -73, "origin_latitude": 44, "origin_longitude": -73}},
		{"ionex_out_of_coverage", map[string]any{"cause": map[string]any{"kind": "before_start", "fields": map[string]any{}}}},
		{"ionex_nodes_not_available", map[string]any{"cause": map[string]any{"kind": "node_gap", "fields": map[string]any{"latitude_posting": 1, "longitude_posting": 2}}}},
		{"ionex_slant_unavailable", map[string]any{"cause": map[string]any{"kind": "unavailable", "fields": map[string]any{}}}},
		{"ionex_epoch", map[string]any{"cause": map[string]any{"kind": "fractional_second", "fields": map[string]any{}}}},
		{"epoch_out_of_range", map[string]any{}},
		{"insufficient_precise_nodes", map[string]any{"satellite_id": "G03", "nodes": 3, "required": 4}},
		{"invalid_input", map[string]any{"message": "incompatible inputs"}},
		{"sp3_epoch_interval", map[string]any{"cause": map[string]any{"kind": "invalid", "fields": map[string]any{"field": "interval"}}}},
		{"sp3_merge_tolerance", map[string]any{"cause": map[string]any{"kind": "invalid", "fields": map[string]any{"field": "tolerance"}}}},
		{"continuity_options", map[string]any{"cause": map[string]any{"kind": "invalid", "fields": map[string]any{"field": "speed_bound"}}}},
		{"sbas_encode", map[string]any{"cause": map[string]any{"kind": "invalid", "fields": map[string]any{"field": "message"}}}},
		{"rtcm_encode", map[string]any{"cause": map[string]any{"kind": "invalid", "fields": map[string]any{"field": "message"}}}},
		{"rtcm_conversion", map[string]any{"cause": map[string]any{"kind": "unsupported", "fields": map[string]any{"message_type": 999}}}},
		{"ut1_outside_coverage", map[string]any{"reason": "outside_table"}},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			nested := map[string]any{"kind": tc.kind, "fields": tc.fields}
			status := groupedReturnedStatus(t, EngineErrorFamilyRtk, "observation", map[string]any{"cause": nested}, "observation: "+tc.kind)
			var top map[string]json.RawMessage
			if err := status.EngineError().UnmarshalFields(&top); err != nil {
				t.Fatal(err)
			}
			var got struct {
				Kind   string         `json:"kind"`
				Fields map[string]any `json:"fields"`
			}
			if err := json.Unmarshal(top["cause"], &got); err != nil {
				t.Fatal(err)
			}
			gotFields, err := json.Marshal(got.Fields)
			if err != nil {
				t.Fatal(err)
			}
			wantFields, err := json.Marshal(tc.fields)
			if err != nil {
				t.Fatal(err)
			}
			if got.Kind != tc.kind || !bytes.Equal(gotFields, wantFields) {
				t.Fatalf("nested core returned value = kind %q fields %s, want %q %s", got.Kind, gotFields, tc.kind, wantFields)
			}
		})
	}
}
