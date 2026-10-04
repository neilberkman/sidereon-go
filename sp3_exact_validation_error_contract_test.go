//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestExactSP3ValidationPayloadPreservesEveryVariantAndField(t *testing.T) {
	cases := []map[string]any{
		{"variant": "parse", "error": "parse error: bad bytes", "debug": `Parse("bad bytes")`},
		{"variant": "catalog", "error": "unknown center bad", "debug": `UnknownCenter("bad")`},
		{"variant": "wrong_product_family", "actual": "clk"},
		{"variant": "invalid_issue", "issue": "2460"},
		{"variant": "unsupported_span_token", "token": "3D"},
		{"variant": "unsupported_sample_token", "token": "7S"},
		{"variant": "non_canonical_span_token", "token": "24H", "canonical": "1D"},
		{"variant": "non_canonical_sample_token", "token": "60S", "canonical": "1M"},
		{"variant": "invalid_expected_agency", "agency": "bad!"},
		{"variant": "agency_mismatch", "expected": "COD", "actual": "GFZ"},
		{"variant": "missing_eof"},
		{"variant": "malformed_eof_record", "line_number": "18446744073709551615", "record_length": "2048"},
		{"variant": "trailing_content_after_eof"},
		{"variant": "mandatory_header_record_count", "record": "%i", "expected": "2", "actual": "1"},
		{"variant": "missing_declared_satellite_count"},
		{"variant": "declared_satellite_count_mismatch", "declared": "18446744073709551615", "tokens": "3"},
		{"variant": "duplicate_declared_satellite", "token": "G01", "first_index": "0", "duplicate_index": "18446744073709551615"},
		{"variant": "no_declared_satellites"},
		{"variant": "satellite_record_sequence_mismatch", "record": "P", "epoch_index": "4", "expected": []any{"G01"}, "actual": []any{"G02"}},
		{"variant": "body_record_interleaving_mismatch", "epoch_index": "5", "expected": []any{"PG01", "VG01"}, "actual": []any{"VG01", "PG01"}},
		{"variant": "non_finite_header_cadence"},
		{"variant": "non_positive_header_cadence", "actual_s": "-inf"},
		{"variant": "unsupported_header_cadence", "actual_s": "100000"},
		{"variant": "cadence_mismatch", "requested_s": "300", "header_s": "600"},
		{"variant": "declared_epoch_count_mismatch", "declared": "18446744073709551615", "parsed": "18446744073709551615"},
		{"variant": "missing_declared_start"},
		{"variant": "declared_start_mismatch", "requested_j2000_s": "1.25", "declared_j2000_s": "1.5", "requested_tick": "170141183460469231731687303715884105727", "declared_tick": "-170141183460469231731687303715884105728"},
		{"variant": "request_before_gps_epoch"},
		{"variant": "non_finite_header_start_metadata", "field": "mjd"},
		{"variant": "invalid_header_start_metadata", "field": "seconds_of_week", "actual": "inf"},
		{"variant": "header_start_metadata_mismatch", "field": "mjd", "requested": "60000", "actual": "60001"},
		{"variant": "empty_epoch_grid"},
		{"variant": "first_epoch_mismatch", "requested_j2000_s": "1", "actual_j2000_s": "2"},
		{"variant": "irregular_epoch_grid", "epoch_index": "18446744073709551615", "requested_s": "300", "actual_s": "299.99999999"},
		{"variant": "span_not_multiple_of_cadence", "span_s": "18446744073709551615", "cadence_s": "300"},
		{"variant": "span_mismatch", "parsed": "20", "half_open": "21", "inclusive": "22"},
		{"variant": "format_version_mismatch", "requested": "d", "actual": "c"},
	}

	if len(cases) != 37 {
		t.Fatalf("exact SP3 variant case count = %d", len(cases))
	}
	seen := make(map[string]struct{}, len(cases))
	for _, wantDetail := range cases {
		variant := wantDetail["variant"].(string)
		t.Run(variant, func(t *testing.T) {
			if _, duplicate := seen[variant]; duplicate {
				t.Fatalf("duplicate exact SP3 variant %q", variant)
			}
			seen[variant] = struct{}{}

			detailJSON, err := json.Marshal(wantDetail)
			if err != nil {
				t.Fatal(err)
			}
			payloadJSON, err := json.Marshal(map[string]any{
				"schema_version": 1,
				"error": map[string]any{
					"kind":       "exact_validation",
					"field":      "other",
					"reason":     "other",
					"details":    json.RawMessage(detailJSON),
					"diagnostic": "exact SP3 fixture " + variant,
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			nativeSP3 := &native.SP3Error{
				Kind:        uint32(SP3ErrorKindExactValidation),
				Field:       uint32(SP3ErrorFieldOther),
				Reason:      uint32(SP3ErrorReasonOther),
				PayloadJSON: string(payloadJSON),
			}
			converted := publicError(&native.StatusError{
				Code:   int(StatusInvalidArgument),
				Text:   "invalid argument",
				Detail: "exact SP3 fixture " + variant,
				SP3:    nativeSP3,
			})
			var statusErr *StatusError
			if !errors.As(converted, &statusErr) || statusErr.SP3 == nil {
				t.Fatalf("public exact SP3 error missing: %T %v", converted, converted)
			}
			if statusErr.SP3.Kind != SP3ErrorKindExactValidation {
				t.Fatalf("public exact SP3 kind = %d", statusErr.SP3.Kind)
			}
			if statusErr.SP3.Error() != "sidereon: SP3 validation kind=1 field=255 reason=255" {
				t.Fatalf("public exact SP3 display = %q", statusErr.SP3.Error())
			}

			var got struct {
				SchemaVersion int `json:"schema_version"`
				Error         struct {
					Details map[string]any `json:"details"`
				} `json:"error"`
			}
			if err := json.Unmarshal(statusErr.SP3.Payload, &got); err != nil {
				t.Fatalf("decode public exact SP3 payload: %v", err)
			}
			if got.SchemaVersion != 1 || !reflect.DeepEqual(got.Error.Details, wantDetail) {
				t.Fatalf("public exact SP3 detail = %#v, want %#v", got.Error.Details, wantDetail)
			}

			owned := append([]byte(nil), statusErr.SP3.Payload...)
			nativeSP3.PayloadJSON = `{"overwritten":true}`
			if !bytes.Equal(statusErr.SP3.Payload, owned) {
				t.Fatalf("public exact SP3 payload aliases native storage: %s", statusErr.SP3.Payload)
			}
		})
	}
	if len(seen) != 37 {
		t.Fatalf("distinct exact SP3 variant count = %d", len(seen))
	}
}
