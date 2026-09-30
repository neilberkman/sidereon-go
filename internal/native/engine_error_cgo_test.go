//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type absentTestError struct{}

func (absentTestError) Error() string { return "absent test error" }

func TestCgoEngineErrorFamilyNames(t *testing.T) {
	cases := []struct {
		family     EngineErrorFamily
		expectedID uint32
		expected   string
	}{
		{EngineErrorFamilyNone, 0, "none"},
		{EngineErrorFamilyRtk, 1, "rtk"},
		{EngineErrorFamilyStaticReference, 2, "static_reference"},
		{EngineErrorFamilyTrls, 3, "trls"},
		{EngineErrorFamilyIls, 4, "ils"},
		{EngineErrorFamilySpk, 5, "spk"},
		{EngineErrorFamilyCdm, 6, "cdm"},
		{EngineErrorFamilyTdm, 7, "tdm"},
		{EngineErrorFamilyFusion, 8, "fusion"},
		{EngineErrorFamilyFusionStateCodec, 9, "fusion_state_codec"},
		{EngineErrorFamilyAllan, 10, "allan"},
		{EngineErrorFamilyPowerLawNoise, 11, "power_law_noise"},
		{EngineErrorFamilyFrameCatalog, 12, "frame_catalog"},
		{EngineErrorFamilySidereal, 13, "sidereal"},
		{EngineErrorFamilyAtmosphere, 14, "atmosphere"},
		{EngineErrorFamilySourceLocalization, 15, "source_localization"},
		{EngineErrorFamilyGeodeticTimeSeries, 16, "geodetic_time_series"},
		{EngineErrorFamilyNormality, 17, "normality"},
		{EngineErrorFamilyTrack, 18, "track"},
		{EngineErrorFamilyPreciseSamples, 19, "precise_samples"},
		{EngineErrorFamilyPreciseInterpolant, 20, "precise_interpolant"},
		{EngineErrorFamilySpaceWeather, 21, "space_weather"},
		{EngineErrorFamilyAraim, 22, "araim"},
		{EngineErrorFamilyReducedOrbit, 23, "reduced_orbit"},
		{EngineErrorFamilyReducedOrbitSource, 24, "reduced_orbit_source"},
		{EngineErrorFamilyPiecewiseOrbit, 25, "piecewise_orbit"},
		{EngineErrorFamilyOrbitFit, 26, "orbit_fit"},
		{EngineErrorFamilyElements, 27, "elements"},
		{EngineErrorFamilyEquinoctial, 28, "equinoctial"},
		{EngineErrorFamilyRtnFrame, 29, "rtn_frame"},
		{EngineErrorFamilyAnomaly, 30, "anomaly"},
		{EngineErrorFamilyPropagation, 31, "propagation"},
		{EngineErrorFamilyDecay, 32, "decay"},
		{EngineErrorFamilyDgnss, 33, "dgnss"},
		{EngineErrorFamilyScenario, 34, "scenario"},
		{EngineErrorFamilyCatalog, 35, "catalog"},
		{EngineErrorFamilyExactCache, 36, "exact_cache"},
		{EngineErrorFamilyTca, 37, "tca"},
		{EngineErrorFamilyAlmanac, 38, "almanac"},
		{EngineErrorFamilyObserve, 39, "observe"},
		{EngineErrorFamilyBodyObservation, 40, "body_observation"},
		{EngineErrorFamilyLookAngle, 41, "look_angle"},
		{EngineErrorFamilyPass, 42, "pass"},
		{EngineErrorFamilyEventFinder, 43, "event_finder"},
		{EngineErrorFamilyFrameTransform, 44, "frame_transform"},
		{EngineErrorFamilyConjunction, 45, "conjunction"},
		{EngineErrorFamilyFacade, 46, "facade"},
		{EngineErrorFamilySpp, 47, "spp"},
		{EngineErrorFamilySppPolicy, 48, "spp_policy"},
		{EngineErrorFamilySunMoon, 49, "sun_moon"},
		{EngineErrorFamilyRinexSpp, 50, "rinex_spp"},
		{EngineErrorFamilySolutionValidation, 51, "solution_validation"},
		{EngineErrorFamilyRF, 52, "rf"},
		{EngineErrorFamilyIonosphereFree, 53, "ionosphere_free"},
		{EngineErrorFamilyDoppler, 54, "doppler"},
		{EngineErrorFamilyOEM, 55, "oem"},
		{EngineErrorFamilyOPM, 56, "opm"},
		{EngineErrorFamilyOMM, 57, "omm"},
		{EngineErrorFamilyDOP, 58, "dop"},
		{EngineErrorFamilyGeofence, 59, "geofence"},
		{EngineErrorFamilySignal, 60, "signal"},
		{EngineErrorFamilyCarrierPhase, 61, "carrier_phase"},
		{EngineErrorFamilySignalAnalysis, 62, "signal_analysis"},
		{EngineErrorFamilyErrorMetrics, 63, "error_metrics"},
		{EngineErrorFamilyObservables, 64, "observables"},
		{EngineErrorFamilyNMEA, 65, "nmea"},
		{EngineErrorFamilyTLEFit, 66, "tle_fit"},
		{EngineErrorFamilyIOD, 67, "iod"},
		{EngineErrorFamilySelection, 68, "selection"},
		{EngineErrorFamilyPppAutoInit, 69, "ppp_auto_init"},
		{EngineErrorFamilyStaticPositioning, 70, "static_positioning"},
		{EngineErrorFamilyTimeOffset, 71, "time_offset"},
		{EngineErrorFamilyTimeModel, 72, "time_model"},
		{EngineErrorFamilyUnknown, 999, "unknown"},
	}

	for _, tc := range cases {
		if uint32(tc.family) != tc.expectedID {
			t.Errorf("native family %s numeric ID = %d, expected %d", tc.expected, uint32(tc.family), tc.expectedID)
		}
		if tc.family != EngineErrorFamily(tc.expectedID) {
			t.Errorf("native family %s cast from ID %d = %d, expected %d", tc.expected, tc.expectedID, tc.family, EngineErrorFamily(tc.expectedID))
		}
		if tc.family.Name() != tc.expected {
			t.Errorf("family %d name = %q, expected %q", tc.family, tc.family.Name(), tc.expected)
		}
		if tc.family.String() != tc.expected {
			t.Errorf("family %d string = %q, expected %q", tc.family, tc.family.String(), tc.expected)
		}
		roundTrip := EngineErrorFamilyFromName(tc.expected)
		if roundTrip != tc.family {
			t.Errorf("round-trip from name %q = %d, expected %d", tc.expected, roundTrip, tc.family)
		}
	}

	for id := uint32(0); id <= 72; id++ {
		fam := EngineErrorFamily(id)
		if fam.Name() == fmt.Sprintf("family_%d", id) {
			t.Errorf("contiguous native family ID %d has unmapped name %q", id, fam.Name())
		}
	}

	if EngineErrorFamilyFromName("nonexistent_subsystem") != EngineErrorFamilyUnknown {
		t.Errorf("expected nonexistent subsystem to map to Unknown")
	}

	// Test future code cases
	future73 := EngineErrorFamily(73)
	if future73.Name() != "family_73" {
		t.Errorf("future family 73 name = %q, expected \"family_73\"", future73.Name())
	}
	if future73.String() != "family_73" {
		t.Errorf("future family 73 string = %q, expected \"family_73\"", future73.String())
	}
	future1000 := EngineErrorFamily(1000)
	if future1000.Name() != "family_1000" {
		t.Errorf("future family 1000 name = %q, expected \"family_1000\"", future1000.Name())
	}
	if EngineErrorFamilyFromName("family_73") != EngineErrorFamilyUnknown {
		t.Errorf("expected unmapped family_73 name to return EngineErrorFamilyUnknown")
	}
}

func TestCgoEngineFloatExactNonFiniteBits(t *testing.T) {
	cases := []struct {
		name       string
		decimal    string
		bitsHex    string
		expectBits uint64
		check      func(t *testing.T, val float64)
	}{
		{
			name:       "positive zero",
			decimal:    "0",
			bitsHex:    "0000000000000000",
			expectBits: 0x0000000000000000,
			check: func(t *testing.T, val float64) {
				if val != 0.0 || math.Signbit(val) {
					t.Errorf("expected +0.0, got %v (signbit=%v)", val, math.Signbit(val))
				}
			},
		},
		{
			name:       "negative zero",
			decimal:    "-0",
			bitsHex:    "8000000000000000",
			expectBits: 0x8000000000000000,
			check: func(t *testing.T, val float64) {
				if val != 0.0 || !math.Signbit(val) {
					t.Errorf("expected -0.0, got %v (signbit=%v)", val, math.Signbit(val))
				}
			},
		},
		{
			name:       "normal float 1.5",
			decimal:    "1.5",
			bitsHex:    "3ff8000000000000",
			expectBits: 0x3ff8000000000000,
			check: func(t *testing.T, val float64) {
				if val != 1.5 {
					t.Errorf("expected 1.5, got %v", val)
				}
			},
		},
		{
			name:       "positive infinity",
			decimal:    "inf",
			bitsHex:    "7ff0000000000000",
			expectBits: 0x7ff0000000000000,
			check: func(t *testing.T, val float64) {
				if !math.IsInf(val, 1) {
					t.Errorf("expected +Inf, got %v", val)
				}
			},
		},
		{
			name:       "negative infinity",
			decimal:    "-inf",
			bitsHex:    "fff0000000000000",
			expectBits: 0xfff0000000000000,
			check: func(t *testing.T, val float64) {
				if !math.IsInf(val, -1) {
					t.Errorf("expected -Inf, got %v", val)
				}
			},
		},
		{
			name:       "quiet NaN",
			decimal:    "NaN",
			bitsHex:    "7ff8000000000000",
			expectBits: 0x7ff8000000000000,
			check: func(t *testing.T, val float64) {
				if !math.IsNaN(val) {
					t.Errorf("expected NaN, got %v", val)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ef := EngineFloat{Decimal: tc.decimal, BitsHex: tc.bitsHex}
			bits, err := ef.Bits()
			if err != nil {
				t.Fatalf("unexpected error parsing bits: %v", err)
			}
			if bits != tc.expectBits {
				t.Fatalf("bits = 0x%016x, expected 0x%016x", bits, tc.expectBits)
			}

			val, err := ef.Float64()
			if err != nil {
				t.Fatalf("unexpected error parsing float64: %v", err)
			}
			tc.check(t, val)

			rawJSON := []byte(`{"decimal":"` + tc.decimal + `","bits_hex":"` + tc.bitsHex + `"}`)
			parsed, err := ParseEngineFloat(rawJSON)
			if err != nil {
				t.Fatalf("unexpected error from ParseEngineFloat: %v", err)
			}
			if parsed.BitsHex != tc.bitsHex || parsed.Decimal != tc.decimal {
				t.Fatalf("parsed %+v does not match expected %+v", parsed, ef)
			}
		})
	}

	malformedCases := []struct {
		name    string
		rawJSON string
	}{
		{"null literal", "null"},
		{"empty object", "{}"},
		{"missing bits_hex", `{"decimal":"1.5"}`},
		{"missing decimal", `{"bits_hex":"3ff8000000000000"}`},
		{"short bits_hex", `{"decimal":"1.5","bits_hex":"1"}`},
		{"long bits_hex", `{"decimal":"1.5","bits_hex":"3ff80000000000000"}`},
		{"non-hex character in bits_hex", `{"decimal":"1.5","bits_hex":"3ff800000000000z"}`},
		{"numeric decimal instead of string", `{"decimal":1.5,"bits_hex":"3ff8000000000000"}`},
		{"duplicate keys", `{"decimal":"1.5","decimal":"2.5","bits_hex":"3ff8000000000000"}`},
		{"uppercase decimal only", `{"DECIMAL":"1.5","bits_hex":"3ff8000000000000"}`},
		{"uppercase bits_hex only", `{"decimal":"1.5","BITS_HEX":"3ff8000000000000"}`},
	}
	for _, tc := range malformedCases {
		t.Run("malformed_"+tc.name, func(t *testing.T) {
			_, err := ParseEngineFloat([]byte(tc.rawJSON))
			if err == nil {
				t.Fatalf("expected ParseEngineFloat to fail on %s, got nil error", tc.name)
			}
		})
	}

	t.Run("float alias key does not override exact canonical key", func(t *testing.T) {
		raw := []byte(`{"decimal":"1.5","DECIMAL":"2.5","bits_hex":"3ff8000000000000"}`)
		ef, err := ParseEngineFloat(raw)
		if err != nil {
			t.Fatalf("unexpected error parsing float with alias key: %v", err)
		}
		if ef.Decimal != "1.5" {
			t.Errorf("decimal = %q, expected \"1.5\"", ef.Decimal)
		}
	})

	t.Run("Bits and Float64 refuse invalid width", func(t *testing.T) {
		efShort := EngineFloat{Decimal: "1.0", BitsHex: "1"}
		if _, err := efShort.Bits(); err == nil {
			t.Fatal("expected Bits() to fail on short hex")
		}
		if _, err := efShort.Float64(); err == nil {
			t.Fatal("expected Float64() to fail on short hex")
		}
	})
}

func TestCgoDecodeSchema1PayloadExactIntegerLexemes(t *testing.T) {
	// 9007199254740993 is 2^53 + 1, an integer that cannot be exactly represented in binary64 float64.
	payload := []byte(`{
		"schema_version": 1,
		"family": "ils",
		"operation": "test_ils_search",
		"error": {
			"kind": "too_many_candidates",
			"fields": {
				"evaluated": 9007199254740993,
				"limit": 9007199254740992,
				"label": "exact_integer_check"
			}
		}
	}`)

	engineErr, err := DecodeSchema1EnginePayload(EngineErrorFamilyIls, payload)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if engineErr.Schema != 1 {
		t.Errorf("schema = %d, expected 1", engineErr.Schema)
	}
	if engineErr.Family != EngineErrorFamilyIls {
		t.Errorf("family = %d, expected %d", engineErr.Family, EngineErrorFamilyIls)
	}
	if engineErr.FamilyName != "ils" {
		t.Errorf("family name = %q, expected ils", engineErr.FamilyName)
	}
	if engineErr.Operation != "test_ils_search" {
		t.Errorf("operation = %q, expected test_ils_search", engineErr.Operation)
	}
	if engineErr.Kind != "too_many_candidates" {
		t.Errorf("kind = %q, expected too_many_candidates", engineErr.Kind)
	}

	type fieldsTarget struct {
		Evaluated json.Number `json:"evaluated"`
		Limit     json.Number `json:"limit"`
		Label     string      `json:"label"`
	}

	var target fieldsTarget
	if err := engineErr.UnmarshalFields(&target); err != nil {
		t.Fatalf("UnmarshalFields failed: %v", err)
	}
	if target.Evaluated.String() != "9007199254740993" {
		t.Errorf("evaluated = %s, expected exact integer 9007199254740993", target.Evaluated.String())
	}
	if target.Limit.String() != "9007199254740992" {
		t.Errorf("limit = %s, expected exact integer 9007199254740992", target.Limit.String())
	}
	if target.Label != "exact_integer_check" {
		t.Errorf("label = %q, expected exact_integer_check", target.Label)
	}
}

func TestCgoDecodeSchema1NonLossPolicy(t *testing.T) {
	t.Run("unknown schema version retained", func(t *testing.T) {
		raw := []byte(`{"schema_version": 99, "family": "rtk", "operation": "future_op", "error": {"kind": "future_kind", "fields": {}}}`)
		decoded, err := DecodeSchema1EnginePayload(EngineErrorFamilyRtk, raw)
		if err == nil {
			t.Fatal("expected error on unsupported schema version")
		}
		if decoded == nil {
			t.Fatal("expected non-nil decoded error retaining raw payload")
		}
		if decoded.Schema != 99 {
			t.Errorf("schema = %d, expected 99", decoded.Schema)
		}
		if string(decoded.Payload) != string(raw) {
			t.Errorf("payload not retained: %s", string(decoded.Payload))
		}
		if decoded.CaptureError == nil {
			t.Error("expected CaptureError to be populated")
		}
	})

	t.Run("missing schema version retained", func(t *testing.T) {
		raw := []byte(`{"family": "rtk", "operation": "bad_op", "error": {"kind": "test", "fields": {}}}`)
		decoded, err := DecodeSchema1EnginePayload(EngineErrorFamilyRtk, raw)
		if err == nil {
			t.Fatal("expected error on missing schema version")
		}
		if decoded == nil || len(decoded.Payload) == 0 {
			t.Fatal("expected non-nil decoded error retaining raw payload")
		}
		if decoded.CaptureError == nil {
			t.Error("expected CaptureError to be populated")
		}
	})

	t.Run("malformed JSON retained", func(t *testing.T) {
		raw := []byte(`{"schema_version": 1, "unterminated: "bad"`)
		decoded, err := DecodeSchema1EnginePayload(EngineErrorFamilyRtk, raw)
		if err == nil {
			t.Fatal("expected error on malformed JSON")
		}
		if decoded == nil || len(decoded.Payload) == 0 {
			t.Fatal("expected non-nil decoded error retaining raw payload")
		}
		if decoded.CaptureError == nil {
			t.Error("expected CaptureError to be populated")
		}
	})

	t.Run("non-standard error body retained as fields", func(t *testing.T) {
		raw := []byte(`{"schema_version": 1, "family": "rtk", "operation": "op", "error": "string_error"}`)
		decoded, err := DecodeSchema1EnginePayload(EngineErrorFamilyRtk, raw)
		if err == nil {
			t.Fatal("expected diagnostic error on non-standard error body")
		}
		if decoded == nil || len(decoded.Fields) == 0 {
			t.Fatal("expected raw error string retained in Fields")
		}
		if string(decoded.Fields) != `"string_error"` {
			t.Errorf("fields = %s, expected \"string_error\"", string(decoded.Fields))
		}
	})
}

func TestCgoDecodeSchema1IsolatedMalformed(t *testing.T) {
	isolatedCases := []struct {
		name    string
		family  EngineErrorFamily
		payload string
	}{
		{
			name:    "missing family operation error",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1}`,
		},
		{
			name:    "null error field",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "error": null}`,
		},
		{
			name:    "empty operation",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "empty kind",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "error": {"kind": "", "fields": {}}}`,
		},
		{
			name:    "fields is array",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "error": {"kind": "singular", "fields": []}}`,
		},
		{
			name:    "missing fields",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "error": {"kind": "singular"}}`,
		},
		{
			name:    "numeric summary family ILS with envelope family rtk",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "rtk", "operation": "search", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "duplicate schema_version",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 99, "schema_version": 1, "family": "ils", "operation": "search", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "duplicate family key",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "family": "ils", "operation": "search", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "duplicate kind key in error body",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "error": {"kind": "singular", "kind": "duplicate", "fields": {}}}`,
		},
		{
			name:    "schema override attempt with uppercase key",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 99, "SCHEMA_VERSION": 1, "family": "ils", "operation": "search", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "uppercase schema_version only",
			family:  EngineErrorFamilyIls,
			payload: `{"SCHEMA_VERSION": 1, "family": "ils", "operation": "search", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "uppercase family only",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "FAMILY": "ils", "operation": "search", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "uppercase operation only",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "OPERATION": "search", "error": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "uppercase error only",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "ERROR": {"kind": "singular", "fields": {}}}`,
		},
		{
			name:    "uppercase kind only in body",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "error": {"KIND": "singular", "fields": {}}}`,
		},
		{
			name:    "uppercase fields only in body",
			family:  EngineErrorFamilyIls,
			payload: `{"schema_version": 1, "family": "ils", "operation": "search", "error": {"kind": "singular", "FIELDS": {}}}`,
		},
	}

	for _, tc := range isolatedCases {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(tc.payload)
			decoded, err := DecodeSchema1EnginePayload(tc.family, raw)
			if err == nil {
				t.Fatalf("expected error on malformed payload %s, but got nil", tc.name)
			}
			if decoded == nil {
				t.Fatal("expected decoded non-nil error")
			}
			if !bytes.Equal(decoded.Payload, raw) {
				t.Errorf("raw payload not preserved: got %s, expected %s", string(decoded.Payload), string(raw))
			}
			if decoded.CaptureError == nil {
				t.Error("expected CaptureError to be populated")
			}
		})
	}

	t.Run("valid unknown future error kind allowed", func(t *testing.T) {
		raw := []byte(`{"schema_version": 1, "family": "ils", "operation": "search", "error": {"kind": "future_unknown_kind", "fields": {"arbitrary": 42}}}`)
		decoded, err := DecodeSchema1EnginePayload(EngineErrorFamilyIls, raw)
		if err != nil {
			t.Fatalf("unexpected error for valid unknown future kind: %v", err)
		}
		if decoded.Kind != "future_unknown_kind" {
			t.Errorf("kind = %q, expected future_unknown_kind", decoded.Kind)
		}
		if decoded.CaptureError != nil {
			t.Errorf("unexpected CaptureError: %v", decoded.CaptureError)
		}
	})

	t.Run("valid future extra fields in envelope and body allowed", func(t *testing.T) {
		raw := []byte(`{"schema_version": 1, "family": "ils", "operation": "search", "future_env_extra": "meta", "error": {"kind": "singular", "fields": {"future_body_extra": true}}}`)
		decoded, err := DecodeSchema1EnginePayload(EngineErrorFamilyIls, raw)
		if err != nil {
			t.Fatalf("unexpected error for valid future extra fields: %v", err)
		}
		if decoded.Kind != "singular" {
			t.Errorf("kind = %q, expected singular", decoded.Kind)
		}
		if decoded.CaptureError != nil {
			t.Errorf("unexpected CaptureError: %v", decoded.CaptureError)
		}
	})

	t.Run("unknown future numeric family and name preserved", func(t *testing.T) {
		futureFamily := EngineErrorFamily(88)
		raw := []byte(`{"schema_version": 1, "family": "deep_space_tracking", "operation": "track", "error": {"kind": "lock_lost", "fields": {"snr": 1.2}}}`)
		decoded, err := DecodeSchema1EnginePayload(futureFamily, raw)
		if err != nil {
			t.Fatalf("unexpected error for future numeric family: %v", err)
		}
		if decoded.Family != futureFamily {
			t.Errorf("family = %d, expected 88", decoded.Family)
		}
		if decoded.FamilyName != "deep_space_tracking" {
			t.Errorf("family name = %q, expected deep_space_tracking", decoded.FamilyName)
		}
		if decoded.CaptureError != nil {
			t.Errorf("unexpected CaptureError: %v", decoded.CaptureError)
		}
	})

	t.Run("future numeric values large integer and range exceeding binary64 preserved", func(t *testing.T) {
		raw := []byte(`{"schema_version": 1, "family": "ils", "operation": "search", "future_env_num": 1e1000, "error": {"kind": "singular", "fields": {"future_value": 1e1000, "large_int": 9223372036854775807999999999}}}`)
		decoded, err := DecodeSchema1EnginePayload(EngineErrorFamilyIls, raw)
		if err != nil {
			t.Fatalf("unexpected error for payload with large numbers: %v", err)
		}
		if decoded == nil {
			t.Fatal("expected non-nil decoded error")
		}
		if !bytes.Equal(decoded.Payload, raw) {
			t.Errorf("raw payload not preserved: got %s, expected %s", string(decoded.Payload), string(raw))
		}
		if decoded.CaptureError != nil {
			t.Errorf("unexpected CaptureError: %v", decoded.CaptureError)
		}

		var fieldsTarget map[string]json.Number
		if err := decoded.UnmarshalFields(&fieldsTarget); err != nil {
			t.Fatalf("UnmarshalFields failed: %v", err)
		}
		if fieldsTarget["future_value"].String() != "1e1000" {
			t.Errorf("future_value = %s, expected 1e1000", fieldsTarget["future_value"])
		}
		if fieldsTarget["large_int"].String() != "9223372036854775807999999999" {
			t.Errorf("large_int = %s, expected 9223372036854775807999999999", fieldsTarget["large_int"])
		}
	})
}

func assertAcyclicGraph(t *testing.T, root error, maxDepth int) {
	t.Helper()
	if root == nil {
		return
	}
	visiting := make(map[error]bool)
	var walk func(err error, depth int)
	walk = func(err error, depth int) {
		if err == nil {
			return
		}
		if depth > maxDepth {
			t.Fatalf("error graph depth exceeded limit %d (cycle suspected): %v", maxDepth, err)
		}
		if visiting[err] {
			t.Fatalf("cycle detected in error graph at node %T: %v", err, err)
		}
		visiting[err] = true
		defer delete(visiting, err)

		type singleUnwrapper interface {
			Unwrap() error
		}
		type multiUnwrapper interface {
			Unwrap() []error
		}

		if u, ok := err.(singleUnwrapper); ok {
			walk(u.Unwrap(), depth+1)
		}
		if u, ok := err.(multiUnwrapper); ok {
			for _, child := range u.Unwrap() {
				walk(child, depth+1)
			}
		}
	}
	walk(root, 0)
}

func TestCgoActualProducerNativeLambdaILS(t *testing.T) {
	ClearEngineError()

	// Covariance of all zeros is degenerate/singular for ILS
	floatCycles := []float64{1.0, 2.0}
	singularCov := []float64{0.0, 0.0, 0.0, 0.0}

	_, _, err := LambdaILS(floatCycles, singularCov, 2, 3.0, nil)
	if err == nil {
		t.Fatal("expected LambdaILS with singular covariance to fail")
	}

	// 1. Verify acyclicity with bounded walk of both Unwrap forms and visited set BEFORE errors.Is/As
	assertAcyclicGraph(t, err, 50)

	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected StatusError, got %T: %v", err, err)
	}

	var engineErr *EngineError
	if !errors.As(err, &engineErr) {
		t.Fatalf("expected EngineError via errors.As, got %v", err)
	}

	if engineErr.Family != EngineErrorFamilyIls {
		t.Errorf("expected family Ils (4), got %d (%s)", engineErr.Family, engineErr.FamilyName)
	}
	if engineErr.FamilyName != "ils" {
		t.Errorf("expected family name ils, got %q", engineErr.FamilyName)
	}
	if engineErr.Schema != 1 {
		t.Errorf("expected schema 1, got %d", engineErr.Schema)
	}
	if engineErr.Operation != "sidereon_lambda_ils_search" {
		t.Errorf("expected operation sidereon_lambda_ils_search, got %q", engineErr.Operation)
	}
	if engineErr.Kind != "singular" {
		t.Errorf("expected kind singular, got %q", engineErr.Kind)
	}
	if len(engineErr.Payload) == 0 {
		t.Error("expected non-empty Payload JSON")
	}

	// Verify unwrap returns CaptureError (which is nil here), not cyclic parent statusErr
	if unwrapped := engineErr.Unwrap(); unwrapped != nil {
		t.Errorf("expected EngineError.Unwrap() to return nil CaptureError, got %v", unwrapped)
	}

	// Unmatched Is and missing-type As terminate safely without spinning
	if errors.Is(err, ErrClosed) {
		t.Error("unexpected match for ErrClosed")
	}
	var absent *absentTestError
	if errors.As(err, &absent) {
		t.Error("unexpected match for absentTestError")
	}
}

func TestCgoThreadDepthNestingAndRetention(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ClearEngineErrorLocked()

	floatCycles := []float64{0.0, 0.0}
	singularCov := []float64{1.0, 1.0, 1.0, 1.0}

	// 1. Enter outer withCThread first, seed real failure INSIDE outer scope
	var savedPayload string
	var retainedErr *EngineError
	withCThread(func() {
		outerDepth := CThreadDepth()
		if outerDepth != 1 {
			t.Fatalf("expected outer depth 1, got %d", outerDepth)
		}

		// Trigger real ILS failure inside active outer scope
		_, _, err := LambdaILS(floatCycles, singularCov, 2, 3.0, nil)
		if err == nil {
			t.Fatal("expected LambdaILS failure")
		}

		// 2. Direct retained summary/payload read inside outer scope
		var qErr error
		retainedErr, qErr = currentEngineErrorLocked()
		if qErr != nil || retainedErr == nil {
			t.Fatalf("direct retained read failed: %v", qErr)
		}
		if retainedErr.Family != EngineErrorFamilyIls || retainedErr.Kind != "singular" {
			t.Errorf("unexpected retained error: %+v", retainedErr)
		}
		if len(retainedErr.Payload) == 0 {
			t.Fatal("expected non-empty retained payload")
		}
		savedPayload = string(retainedErr.Payload)

		// 3. Nested withCThread call without a new producer and equal payload
		withCThread(func() {
			innerDepth := CThreadDepth()
			if innerDepth != 2 {
				t.Fatalf("expected inner depth 2, got %d", innerDepth)
			}
			nestedErr, nErr := currentEngineErrorLocked()
			if nErr != nil || nestedErr == nil {
				t.Fatalf("nested direct read failed: %v", nErr)
			}
			if nestedErr.Family != retainedErr.Family {
				t.Errorf("nested family = %d, expected %d", nestedErr.Family, retainedErr.Family)
			}
			if string(nestedErr.Payload) != savedPayload {
				t.Errorf("nested payload = %s, expected %s", string(nestedErr.Payload), savedPayload)
			}
		})
	})

	// After leaving outer scope, depth is 0.
	// Separately verify an independent call clears the slot.
	withCThread(func() {
		clearedErr, cErr := currentEngineErrorLocked()
		if cErr != nil {
			t.Fatalf("direct read after leaving outer scope returned error: %v", cErr)
		}
		if clearedErr != nil {
			t.Errorf("expected empty generic slot after independent outer entry, got: %+v", clearedErr)
		}
	})

	// 4. Subsequent independent successful producer with empty generic slot
	goodFloats := []float64{0.0, 0.0}
	goodCov := []float64{1.0, 0.0, 0.0, 1.0}
	fixed, res, succErr := LambdaILS(goodFloats, goodCov, 2, 3.0, nil)
	if succErr != nil {
		t.Fatalf("expected successful LambdaILS, got: %v", succErr)
	}
	if len(fixed) != 2 || !res.FixedStatus {
		t.Fatalf("unexpected LambdaILS result: fixed=%v res=%+v", fixed, res)
	}
	clearedErr, cErr := currentEngineErrorLocked()
	if cErr != nil {
		t.Fatalf("direct read after success returned error: %v", cErr)
	}
	if clearedErr != nil {
		t.Errorf("expected empty generic slot after independent success, got: %+v", clearedErr)
	}

	// 5. Seed an actual ILS failure immediately BEFORE unrelated C refusal with same OSThread retained reader
	_, _, seedErr := LambdaILS(floatCycles, singularCov, 2, 3.0, nil)
	if seedErr == nil {
		t.Fatal("expected seeding failure")
	}
	seededErr, sErr := currentEngineErrorLocked()
	if sErr != nil || seededErr == nil {
		t.Fatalf("expected seeded engine error before unrelated refusal, got: %v", sErr)
	}

	// Unrelated C failure with no stale Engine attachment
	_, dayErr := DayOfYear(2024, 13, 1, 0, 0, 0)
	if dayErr == nil {
		t.Fatal("expected DayOfYear(2024, 13, 1, 0, 0, 0) to fail in C")
	}
	statusErr, ok := dayErr.(*StatusError)
	if !ok {
		t.Fatalf("expected *StatusError, got %T: %v", dayErr, dayErr)
	}
	if statusErr.Engine != nil {
		t.Errorf("unrelated C failure inherited stale engine error: %+v", statusErr.Engine)
	}
	postErr, _ := currentEngineErrorLocked()
	if postErr != nil {
		t.Errorf("expected empty generic slot after unrelated failure, got: %+v", postErr)
	}
}

func TestCgoConcurrentDistinguishableFailures(t *testing.T) {
	const goroutines = 9
	const iterations = 6

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				workerMode := (id + i) % 3
				switch workerMode {
				case 0:
					// Distinguishable genuine failure 1: singular covariance
					floatCycles := []float64{0.0, 0.0}
					singularCov := []float64{1.0, 1.0, 1.0, 1.0}
					_, _, err := LambdaILS(floatCycles, singularCov, 2, 3.0, nil)
					if err == nil {
						t.Errorf("worker %d iteration %d: expected failure", id, i)
						return
					}
					var engineErr *EngineError
					if !errors.As(err, &engineErr) {
						t.Errorf("worker %d iteration %d: expected EngineError", id, i)
						return
					}
					if engineErr.Family != EngineErrorFamilyIls || engineErr.Kind != "singular" {
						t.Errorf("worker %d iteration %d: expected singular ILS error, got: %+v", id, i, engineErr)
						return
					}
				case 1:
					// Distinguishable genuine failure 2: valid covariance with negative ratio threshold -> invalid_input
					floatCycles := []float64{0.0, 0.0}
					goodCov := []float64{1.0, 0.0, 0.0, 1.0}
					_, _, err := LambdaILS(floatCycles, goodCov, 2, -1.0, nil)
					if err == nil {
						t.Errorf("worker %d iteration %d: expected failure", id, i)
						return
					}
					var engineErr *EngineError
					if !errors.As(err, &engineErr) {
						t.Errorf("worker %d iteration %d: expected EngineError", id, i)
						return
					}
					if engineErr.Family != EngineErrorFamilyIls || engineErr.Kind != "invalid_input" {
						t.Errorf("worker %d iteration %d: expected invalid_input ILS error, got: %+v", id, i, engineErr)
						return
					}
					var fields struct {
						Field  string `json:"field"`
						Reason string `json:"reason"`
					}
					if err := engineErr.UnmarshalFields(&fields); err != nil {
						t.Errorf("worker %d iteration %d: UnmarshalFields failed: %v", id, i, err)
						return
					}
					if fields.Field != "ils ratio_threshold" || fields.Reason != "negative" {
						t.Errorf("worker %d iteration %d: unexpected fields: %+v", id, i, fields)
						return
					}
				case 2:
					// Valid control: well-conditioned covariance and valid threshold -> Ok
					floatCycles := []float64{0.0, 0.0}
					goodCov := []float64{1.0, 0.0, 0.0, 1.0}
					fixed, res, err := LambdaILS(floatCycles, goodCov, 2, 3.0, nil)
					if err != nil {
						t.Errorf("worker %d iteration %d: expected success, got error: %v", id, i, err)
						return
					}
					if len(fixed) != 2 || !res.FixedStatus {
						t.Errorf("worker %d iteration %d: unexpected valid control result: fixed=%v, res=%+v", id, i, fixed, res)
						return
					}
				}
			}
		}(g)
	}

	wg.Wait()
}

func TestCgoDirectNativeSPPBatchOwnedRowEngineErrors(t *testing.T) {
	sp3Data, err := os.ReadFile("../../testdata/trimmed.sp3")
	if err != nil {
		t.Fatalf("failed to read testdata/trimmed.sp3: %v", err)
	}
	sp3, err := LoadSP3(sp3Data)
	if err != nil {
		t.Fatalf("LoadSP3 failed: %v", err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	obs8 := []SPPObservation{
		{SatelliteID: "G08", PseudorangeM: math.Float64frombits(0x4176b8c6fd82e861)},
		{SatelliteID: "G10", PseudorangeM: math.Float64frombits(0x4175aa4fa1a0c21f)},
		{SatelliteID: "G16", PseudorangeM: math.Float64frombits(0x417387abd6052c3b)},
		{SatelliteID: "G18", PseudorangeM: math.Float64frombits(0x4174c288f3bd1166)},
		{SatelliteID: "G20", PseudorangeM: math.Float64frombits(0x417443947bd00bd6)},
		{SatelliteID: "G21", PseudorangeM: math.Float64frombits(0x4173d8405cd09f84)},
		{SatelliteID: "G26", PseudorangeM: math.Float64frombits(0x417425d51967e798)},
		{SatelliteID: "G27", PseudorangeM: math.Float64frombits(0x41745a4b78a81707)},
	}

	goodConfig := SPPConfig{
		Observations:    obs8,
		TRxJ2000S:       646272000.0,
		TRxSecondOfDayS: 43200.0,
		DayOfYear:       176.5,
		InitialGuess:    [4]float64{4.5e6, 0.5e6, 4.5e6, 0},
	}
	badConfig := goodConfig
	badConfig.Observations = obs8[:2] // "G08" and "G10", both GPS

	goodInput := SppInputsV2{Base: goodConfig}
	badInput := SppInputsV2{Base: badConfig}
	epochs := []SppInputsV2{goodInput, badInput, goodInput}

	serialBatch, err := SolveSPPBatchSerial(sp3, epochs, false, NativeSPPSolvePolicy{})
	if err != nil {
		t.Fatalf("SolveSPPBatchSerial failed: %v", err)
	}
	t.Cleanup(func() { _ = serialBatch.Close() })

	parallelBatch, err := SolveSPPBatchParallel(sp3, epochs, false, NativeSPPSolvePolicy{})
	if err != nil {
		t.Fatalf("SolveSPPBatchParallel failed: %v", err)
	}
	t.Cleanup(func() { _ = parallelBatch.Close() })

	for mode, batch := range map[string]*SPPBatch{"serial": serialBatch, "parallel": parallelBatch} {
		count, err := batch.Count()
		if err != nil || count != 3 {
			t.Fatalf("%s: batch count = %d, %v (want 3)", mode, count, err)
		}

		// 1. Successful epoch 0
		ok0, err := batch.EpochOK(0)
		if err != nil || !ok0 {
			t.Fatalf("%s: epoch 0 ok = %v, %v", mode, ok0, err)
		}
		errText0, err := batch.Error(0)
		if err != nil || errText0 != "" {
			t.Fatalf("%s: epoch 0 Error = %q, %v (want empty, nil)", mode, errText0, err)
		}
		engErr0, err := batch.EngineError(0)
		if err != nil || engErr0 != nil {
			t.Fatalf("%s: epoch 0 EngineError = %v, %v (want nil, nil)", mode, engErr0, err)
		}
		sol0, err := batch.Solution(0)
		if err != nil || sol0 == nil {
			t.Fatalf("%s: epoch 0 Solution = %v, %v", mode, sol0, err)
		}
		_ = sol0.Close()

		// Successful epoch 2
		ok2, err := batch.EpochOK(2)
		if err != nil || !ok2 {
			t.Fatalf("%s: epoch 2 ok = %v, %v", mode, ok2, err)
		}
		engErr2, err := batch.EngineError(2)
		if err != nil || engErr2 != nil {
			t.Fatalf("%s: epoch 2 EngineError = %v, %v (want nil, nil)", mode, engErr2, err)
		}

		// 2. Failed epoch 1
		ok1, err := batch.EpochOK(1)
		if err != nil || ok1 {
			t.Fatalf("%s: epoch 1 ok = %v, %v (want false, nil)", mode, ok1, err)
		}
		originalText, err := batch.Error(1)
		if err != nil {
			t.Fatalf("%s: epoch 1 Error getter failed: %v", mode, err)
		}
		wantOriginalText := "SPP solve failed: only 2 usable satellites; need at least 4 (3 position + 1 clock per GNSS)"
		if originalText != wantOriginalText {
			t.Fatalf("%s: epoch 1 Error text = %q, want %q", mode, originalText, wantOriginalText)
		}
		repText, err := batch.Error(1)
		if err != nil || repText != originalText {
			t.Fatalf("%s: repeated Error getter = %q, want %q (%v)", mode, repText, originalText, err)
		}

		engErr1, err := batch.EngineError(1)
		if err != nil || engErr1 == nil {
			t.Fatalf("%s: epoch 1 EngineError = %v, %v", mode, engErr1, err)
		}
		if engErr1.Family != EngineErrorFamilySpp {
			t.Errorf("%s: family = %d (%s), want %d", mode, engErr1.Family, engErr1.Family.Name(), EngineErrorFamilySpp)
		}
		if engErr1.FamilyName != "spp" {
			t.Errorf("%s: family name = %q, want \"spp\"", mode, engErr1.FamilyName)
		}
		if engErr1.Schema != 1 {
			t.Errorf("%s: schema = %d, want 1", mode, engErr1.Schema)
		}
		if engErr1.Kind != "too_few_satellites" {
			t.Errorf("%s: kind = %q, want \"too_few_satellites\"", mode, engErr1.Kind)
		}
		if !strings.Contains(engErr1.Operation, "epoch 1") {
			t.Errorf("%s: operation = %q, want operation containing \"epoch 1\"", mode, engErr1.Operation)
		}
		if engErr1.CaptureError != nil {
			t.Errorf("%s: unexpected CaptureError: %v", mode, engErr1.CaptureError)
		}

		var fields struct {
			Used     int `json:"used"`
			Required int `json:"required"`
		}
		if err := engErr1.UnmarshalFields(&fields); err != nil {
			t.Fatalf("%s: UnmarshalFields failed: %v", mode, err)
		}
		if fields.Used != 2 || fields.Required != 4 {
			t.Fatalf("%s: unexpected fields: used=%d required=%d (want 2, 4)", mode, fields.Used, fields.Required)
		}

		// 3. Failed Solution(1)
		sol1, solErr := batch.Solution(1)
		if solErr == nil {
			t.Fatalf("%s: expected Solution(1) to fail", mode)
		}
		if sol1 != nil {
			_ = sol1.Close()
			t.Fatalf("%s: expected nil solution on failure", mode)
		}
		assertAcyclicGraph(t, solErr, 50)

		var statusErr *StatusError
		if !errors.As(solErr, &statusErr) {
			t.Fatalf("%s: expected *StatusError, got %T: %v", mode, solErr, solErr)
		}
		if statusErr.Code != 5 { // StatusSolve
			t.Errorf("%s: status code = %d, want 5 (StatusSolve)", mode, statusErr.Code)
		}
		wantPrefix := "sidereon_spp_batch_solution: epoch 1 did not solve: "
		if !strings.HasPrefix(statusErr.Detail, wantPrefix) {
			t.Fatalf("%s: StatusError.Detail %q missing required prefix %q", mode, statusErr.Detail, wantPrefix)
		}
		if statusErr.Detail != wantPrefix+originalText {
			t.Fatalf("%s: StatusError.Detail = %q, want %q", mode, statusErr.Detail, wantPrefix+originalText)
		}

		var fromSolEngErr *EngineError
		if !errors.As(solErr, &fromSolEngErr) {
			t.Fatalf("%s: expected *EngineError via errors.As from failed Solution", mode)
		}
		if fromSolEngErr.Family != EngineErrorFamilySpp || fromSolEngErr.Kind != "too_few_satellites" {
			t.Errorf("%s: attached engine error mismatch: %+v", mode, fromSolEngErr)
		}

		// 4. Mutation and retention checks
		savedPayload := append([]byte(nil), engErr1.Payload...)
		savedFields := append([]byte(nil), engErr1.Fields...)

		if len(engErr1.Payload) > 0 {
			engErr1.Payload[0] ^= 0xFF
		}
		if len(engErr1.Fields) > 0 {
			engErr1.Fields[0] ^= 0xFF
		}

		repeatedEngErr, err := batch.EngineError(1)
		if err != nil || repeatedEngErr == nil {
			t.Fatalf("%s: repeated EngineError call failed: %v", mode, err)
		}
		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) {
			t.Fatalf("%s: repeated accessor returned mutated payload bytes", mode)
		}
		if !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeated accessor returned mutated fields bytes", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) {
			t.Fatalf("%s: fromSolEngErr payload bytes mutated", mode)
		}
		if !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr fields bytes mutated", mode)
		}

		// Clear generic TLS and assert retention
		ClearEngineError()
		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) || !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeatedEngErr corrupted after ClearEngineError", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) || !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr corrupted after ClearEngineError", mode)
		}

		// Unrelated producer (LambdaILS failure)
		floatCycles := []float64{1.0, 2.0}
		singularCov := []float64{0.0, 0.0, 0.0, 0.0}
		_, _, ilsErr := LambdaILS(floatCycles, singularCov, 2, 3.0, nil)
		if ilsErr == nil {
			t.Fatalf("%s: expected LambdaILS failure", mode)
		}
		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) || !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeatedEngErr corrupted after unrelated LambdaILS failure", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) || !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr corrupted after unrelated LambdaILS failure", mode)
		}

		// Close batch and assert retention
		if err := batch.Close(); err != nil {
			t.Fatalf("%s: batch Close failed: %v", mode, err)
		}
		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) || !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeatedEngErr corrupted after batch close", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) || !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr corrupted after batch close", mode)
		}
		if repeatedEngErr.Family != EngineErrorFamilySpp || repeatedEngErr.Kind != "too_few_satellites" || repeatedEngErr.Schema != 1 || repeatedEngErr.Operation != engErr1.Operation {
			t.Fatalf("%s: copied record corrupted after batch close: %+v", mode, repeatedEngErr)
		}
		if fromSolEngErr.Family != EngineErrorFamilySpp || fromSolEngErr.Kind != "too_few_satellites" || fromSolEngErr.Schema != 1 || fromSolEngErr.Operation != engErr1.Operation {
			t.Fatalf("%s: attached sol record corrupted after batch close: %+v", mode, fromSolEngErr)
		}

		// Closed batch semantics
		if _, err := batch.EngineError(1); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close EngineError error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.Solution(0); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close Solution error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.EpochOK(0); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close EpochOK error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.Error(0); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close Error error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.Count(); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close Count error = %v, want ErrClosed", mode, err)
		}
	}

	// Live batch invalid index semantics
	liveBatch, err := SolveSPPBatchSerial(sp3, []SppInputsV2{goodInput}, false, NativeSPPSolvePolicy{})
	if err != nil {
		t.Fatalf("SolveSPPBatchSerial: %v", err)
	}
	t.Cleanup(func() { _ = liveBatch.Close() })

	if _, err := liveBatch.EngineError(-1); err == nil {
		t.Errorf("expected error on negative index -1")
	}
	if _, err := liveBatch.EngineError(999); err == nil {
		t.Errorf("expected error on out-of-range index 999")
	}
	var invStatusErr *StatusError
	if _, err := liveBatch.EngineError(999); !errors.As(err, &invStatusErr) || invStatusErr.Code != 2 { // StatusInvalidArgument == 2
		t.Errorf("expected StatusInvalidArgument on out-of-range index 999, got %v", err)
	}
}
