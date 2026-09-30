package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

type absentTestError struct{}

func (absentTestError) Error() string { return "absent test error" }

func TestEngineErrorProducerLambdaILSSingular(t *testing.T) {
	ClearEngineError()

	// Singular 2x2 covariance matrix triggers core IlsError::Singular
	floatCycles := []float64{1.0, 2.0}
	singularCov := []float64{0.0, 0.0, 0.0, 0.0}

	// 1. Obtain real error directly from internal/native.LambdaILS
	_, _, nativeErr := native.LambdaILS(floatCycles, singularCov, 2, 3.0, nil)
	if nativeErr == nil {
		t.Fatal("expected native.LambdaILS with singular covariance to fail")
	}

	// 2. Guard the native graph directly with bounded walk before publicError traverses it
	assertAcyclicGraph(t, nativeErr, 50)

	// 3. Translate to public error via publicError
	err := publicError(nativeErr)
	if err == nil {
		t.Fatal("expected public error from native failure")
	}

	// 4. Assert translated graph is acyclic
	assertAcyclicGraph(t, err, 50)

	// 5. Check public error structure
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected *StatusError, got %T: %v", err, err)
	}

	var engineErr *EngineError
	if !errors.As(err, &engineErr) {
		t.Fatalf("expected *EngineError via errors.As, got %v", err)
	}

	if engineErr.Family != EngineErrorFamilyIls {
		t.Errorf("family = %d, expected %d (EngineErrorFamilyIls)", engineErr.Family, EngineErrorFamilyIls)
	}
	if engineErr.FamilyName != "ils" {
		t.Errorf("family name = %q, expected \"ils\"", engineErr.FamilyName)
	}
	if engineErr.Schema != 1 {
		t.Errorf("schema = %d, expected 1", engineErr.Schema)
	}
	if engineErr.Operation != "sidereon_lambda_ils_search" {
		t.Errorf("operation = %q, expected \"sidereon_lambda_ils_search\"", engineErr.Operation)
	}
	if engineErr.Kind != "singular" {
		t.Errorf("kind = %q, expected \"singular\"", engineErr.Kind)
	}
	if len(engineErr.Payload) == 0 {
		t.Error("expected non-empty Payload")
	}

	// Verify statusErr.EngineError() helper
	if statusErr.EngineError() != engineErr {
		t.Errorf("statusErr.EngineError() = %v, expected %v", statusErr.EngineError(), engineErr)
	}

	// Verify unwrap returns nil CaptureError (acyclic, does not point back to statusErr)
	if unwrapped := engineErr.Unwrap(); unwrapped != nil {
		t.Errorf("engineErr.Unwrap() = %v, expected nil", unwrapped)
	}

	// 6. Missing-target Is and As terminate safely without spinning
	if errors.Is(err, ErrClosed) {
		t.Error("unexpected match for ErrClosed")
	}
	var absent *absentTestError
	if errors.As(err, &absent) {
		t.Error("unexpected match for absentTestError")
	}

	// 7. Retain normal public API coverage separately
	res, pubErr := LambdaILS(floatCycles, singularCov, 3.0)
	if pubErr == nil {
		t.Fatalf("expected public LambdaILS to fail, got result: %+v", res)
	}
	var pubEngineErr *EngineError
	if !errors.As(pubErr, &pubEngineErr) {
		t.Fatalf("expected *EngineError from public LambdaILS, got %v", pubErr)
	}
	if pubEngineErr.Kind != "singular" {
		t.Errorf("public LambdaILS kind = %q, expected \"singular\"", pubEngineErr.Kind)
	}
}

func TestEngineErrorPayloadOwnedAcrossLaterOperations(t *testing.T) {
	ClearEngineError()

	floatCycles := []float64{1.0, 2.0}
	singularCov := []float64{0.0, 0.0, 0.0, 0.0}

	_, err := LambdaILS(floatCycles, singularCov, 3.0)
	if err == nil {
		t.Fatal("expected failure")
	}

	var engineErr *EngineError
	if !errors.As(err, &engineErr) {
		t.Fatalf("expected EngineError, got %v", err)
	}

	savedPayload := append([]byte(nil), engineErr.Payload...)
	savedFields := append([]byte(nil), engineErr.Fields...)

	// Perform subsequent ClearEngineError
	ClearEngineError()

	// Perform another operation
	sec, secErr := SecondOfDay(12, 0, 0.0)
	if secErr != nil || sec != 43200.0 {
		t.Fatalf("unexpected SecondOfDay error: %v", secErr)
	}

	// Verify that the captured engine error bytes were not mutated or released
	if !bytes.Equal(engineErr.Payload, savedPayload) {
		t.Errorf("payload was mutated across subsequent operation: before=%s, after=%s", string(savedPayload), string(engineErr.Payload))
	}
	if !bytes.Equal(engineErr.Fields, savedFields) {
		t.Errorf("fields was mutated across subsequent operation: before=%s, after=%s", string(savedFields), string(engineErr.Fields))
	}
}

func TestEngineErrorSuccessAndUnrelatedArgumentFailureClear(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ClearEngineError()

	// 1. Successful producer reset control
	// Induce an engine error
	_, err := LambdaILS([]float64{1.0, 2.0}, []float64{0.0, 0.0, 0.0, 0.0}, 3.0)
	if err == nil {
		t.Fatal("expected failure")
	}
	var engineErr *EngineError
	if !errors.As(err, &engineErr) {
		t.Fatalf("expected EngineError, got %v", err)
	}

	// Perform a successful operation
	sec, err := SecondOfDay(10, 15, 30.0)
	if err != nil {
		t.Fatalf("successful call failed: %v", err)
	}
	if sec != 36930.0 {
		t.Errorf("unexpected SecondOfDay result: %f", sec)
	}

	// Verify generic slot is cleared by successful producer
	clearedNative, cErr := native.CurrentEngineErrorLocked()
	if cErr != nil {
		t.Fatalf("retained reader failed: %v", cErr)
	}
	if clearedNative != nil {
		t.Errorf("expected empty generic slot after successful operation, got: %+v", clearedNative)
	}

	// 2. Perform an unrelated argument failure (failing before C call)
	// Passing mismatched covariance length: len(floatCycles)=2 requires len(covariance)=4, passing 1
	_, argErr := LambdaILS([]float64{1.0, 2.0}, []float64{1.0}, 3.0)
	if argErr == nil {
		t.Fatal("expected argument error")
	}

	var residualEngineErr *EngineError
	if errors.As(argErr, &residualEngineErr) {
		t.Errorf("unrelated argument error inherited previous engine error: %+v", residualEngineErr)
	}

	// 3. Real ILS seed immediately BEFORE unrelated C refusal with same OSThread retained reader
	_, seedErr := LambdaILS([]float64{1.0, 2.0}, []float64{0.0, 0.0, 0.0, 0.0}, 3.0)
	if seedErr == nil {
		t.Fatal("expected seeding failure")
	}
	seededNative, sErr := native.CurrentEngineErrorLocked()
	if sErr != nil || seededNative == nil {
		t.Fatalf("expected seeded engine error before unrelated refusal, got: %v", sErr)
	}

	// 4. Perform an unrelated failing C call: DayOfYear(2024, 13, 1, 0, 0, 0)
	_, dayErr := DayOfYear(2024, 13, 1, 0, 0, 0)
	if dayErr == nil {
		t.Fatal("expected DayOfYear(2024, 13, 1, 0, 0, 0) to fail in C")
	}
	var cStatusErr *StatusError
	if !errors.As(dayErr, &cStatusErr) {
		t.Fatalf("expected StatusError, got %T: %v", dayErr, dayErr)
	}
	if cStatusErr.Engine != nil {
		t.Errorf("unrelated C failure inherited previous engine error: %+v", cStatusErr.Engine)
	}
	postNative, _ := native.CurrentEngineErrorLocked()
	if postNative != nil {
		t.Errorf("expected empty generic slot after unrelated C refusal, got: %+v", postNative)
	}
}

func TestEngineErrorConcurrentGoroutinesIsolation(t *testing.T) {
	const goroutines = 9
	const iterations = 6

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				workerMode := (gid + i) % 3
				switch workerMode {
				case 0:
					// Failing ILS search with singular covariance
					_, err := LambdaILS([]float64{0.0, 0.0}, []float64{0.0, 0.0, 0.0, 0.0}, 3.0)
					if err == nil {
						t.Errorf("goroutine %d iteration %d: expected failure", gid, i)
						return
					}
					var ee *EngineError
					if !errors.As(err, &ee) {
						t.Errorf("goroutine %d iteration %d: expected EngineError", gid, i)
						return
					}
					if ee.Family != EngineErrorFamilyIls || ee.Kind != "singular" {
						t.Errorf("goroutine %d iteration %d: unexpected EngineError: %+v", gid, i, ee)
						return
					}
				case 1:
					// Failing ILS search with valid covariance but negative ratio threshold -> invalid_input
					_, err := LambdaILS([]float64{0.0, 0.0}, []float64{1.0, 0.0, 0.0, 1.0}, -1.0)
					if err == nil {
						t.Errorf("goroutine %d iteration %d: expected failure", gid, i)
						return
					}
					var ee *EngineError
					if !errors.As(err, &ee) {
						t.Errorf("goroutine %d iteration %d: expected EngineError", gid, i)
						return
					}
					if ee.Family != EngineErrorFamilyIls || ee.Kind != "invalid_input" {
						t.Errorf("goroutine %d iteration %d: expected invalid_input, got: %+v", gid, i, ee)
						return
					}
					var fields struct {
						Field  string `json:"field"`
						Reason string `json:"reason"`
					}
					if err := ee.UnmarshalFields(&fields); err != nil {
						t.Errorf("goroutine %d iteration %d: UnmarshalFields failed: %v", gid, i, err)
						return
					}
					if fields.Field != "ils ratio_threshold" || fields.Reason != "negative" {
						t.Errorf("goroutine %d iteration %d: unexpected fields: %+v", gid, i, fields)
						return
					}
				case 2:
					// Successful operation
					sec, err := SecondOfDay(1, 2, 3.0)
					if err != nil {
						t.Errorf("goroutine %d iteration %d: unexpected error: %v", gid, i, err)
						return
					}
					if sec != 3723.0 {
						t.Errorf("goroutine %d iteration %d: unexpected sec: %f", gid, i, sec)
						return
					}
				}
			}
		}(g)
	}

	wg.Wait()
}

func TestEngineErrorExactIntegerAndNonFiniteBits(t *testing.T) {
	// 1. Non-finite and signed-zero float bits
	cases := []struct {
		name    string
		decimal string
		bitsHex string
		check   func(t *testing.T, val float64)
	}{
		{
			name:    "positive zero",
			decimal: "0",
			bitsHex: "0000000000000000",
			check: func(t *testing.T, val float64) {
				if val != 0.0 || math.Signbit(val) {
					t.Errorf("expected +0.0, got %v", val)
				}
			},
		},
		{
			name:    "negative zero",
			decimal: "-0",
			bitsHex: "8000000000000000",
			check: func(t *testing.T, val float64) {
				if val != 0.0 || !math.Signbit(val) {
					t.Errorf("expected -0.0, got %v", val)
				}
			},
		},
		{
			name:    "positive infinity",
			decimal: "inf",
			bitsHex: "7ff0000000000000",
			check: func(t *testing.T, val float64) {
				if !math.IsInf(val, 1) {
					t.Errorf("expected +Inf, got %v", val)
				}
			},
		},
		{
			name:    "negative infinity",
			decimal: "-inf",
			bitsHex: "fff0000000000000",
			check: func(t *testing.T, val float64) {
				if !math.IsInf(val, -1) {
					t.Errorf("expected -Inf, got %v", val)
				}
			},
		},
		{
			name:    "quiet NaN",
			decimal: "NaN",
			bitsHex: "7ff8000000000000",
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
			val, err := ef.Float64()
			if err != nil {
				t.Fatalf("unexpected Float64 error: %v", err)
			}
			tc.check(t, val)

			raw := []byte(`{"decimal":"` + tc.decimal + `","bits_hex":"` + tc.bitsHex + `"}`)
			parsed, err := ParseEngineFloat(raw)
			if err != nil {
				t.Fatalf("unexpected ParseEngineFloat error: %v", err)
			}
			if parsed.BitsHex != tc.bitsHex {
				t.Errorf("BitsHex = %q, expected %q", parsed.BitsHex, tc.bitsHex)
			}
		})
	}

	malformedFloatCases := []struct {
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
	for _, tc := range malformedFloatCases {
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

	// 2. Exact integer lexeme preservation via UnmarshalFields
	rawFields := []byte(`{
		"evaluated": 9223372036854775807,
		"limit": 9223372036854775806,
		"future_num": 1e1000,
		"name": "large_int_test"
	}`)

	ee := &EngineError{
		Family:     EngineErrorFamilyIls,
		FamilyName: "ils",
		Fields:     rawFields,
	}

	type target struct {
		Evaluated   json.Number `json:"evaluated"`
		Limit       json.Number `json:"limit"`
		FutureValue json.Number `json:"future_num"`
		Name        string      `json:"name"`
	}

	var tgt target
	if err := ee.UnmarshalFields(&tgt); err != nil {
		t.Fatalf("UnmarshalFields failed: %v", err)
	}

	if tgt.Evaluated.String() != "9223372036854775807" {
		t.Errorf("evaluated = %s, expected 9223372036854775807", tgt.Evaluated.String())
	}
	if tgt.Limit.String() != "9223372036854775806" {
		t.Errorf("limit = %s, expected 9223372036854775806", tgt.Limit.String())
	}
	if tgt.FutureValue.String() != "1e1000" {
		t.Errorf("future_num = %s, expected 1e1000", tgt.FutureValue.String())
	}
}

func TestEngineErrorRecursiveTypedFieldsPreserveCompletePayload(t *testing.T) {
	payload := []byte(`{"schema_version":1,"family":"observables","operation":"typed_fields_control","error":{"kind":"failure","fields":{"count":18446744073709551615,"future":1e1000,"nullable":null,"ok":true,"names":["one","two"],"cause":{"kind":"unknown_satellite","fields":{"satellite_id":"G99","measurement":1e-300}}}}}`)
	nativeValue, err := native.DecodeSchema1EnginePayload(native.EngineErrorFamilyObservables, payload)
	if err != nil {
		t.Fatalf("decode typed engine payload: %v", err)
	}
	value := publicEngineError(nativeValue)
	if value == nil || value.TypedFields["count"].Kind != EngineJSONNumber || value.TypedFields["count"].Number.String() != "18446744073709551615" {
		t.Fatalf("exact unsigned count = %+v", value)
	}
	if value.TypedFields["future"].Number.String() != "1e1000" {
		t.Fatalf("huge exponent token = %+v", value.TypedFields["future"])
	}
	if value.TypedFields["nullable"].Kind != EngineJSONNull || value.TypedFields["ok"].Kind != EngineJSONBoolean || !value.TypedFields["ok"].Bool {
		t.Fatalf("null/bool typed fields = %+v", value.TypedFields)
	}
	names := value.TypedFields["names"]
	if names.Kind != EngineJSONArray || len(names.Array) != 2 || names.Array[1].String != "two" {
		t.Fatalf("typed string array = %+v", names)
	}
	cause := value.TypedFields["cause"]
	if cause.Kind != EngineJSONObject || cause.Object["kind"].String != "unknown_satellite" || cause.Object["fields"].Object["satellite_id"].String != "G99" || cause.Object["fields"].Object["measurement"].Number.String() != "1e-300" {
		t.Fatalf("nested typed cause = %+v", cause)
	}
	if string(value.Fields) != `{"count":18446744073709551615,"future":1e1000,"nullable":null,"ok":true,"names":["one","two"],"cause":{"kind":"unknown_satellite","fields":{"satellite_id":"G99","measurement":1e-300}}}` {
		t.Fatalf("raw fields changed: %s", value.Fields)
	}
	nativeCause := nativeValue.TypedFields["cause"]
	nativeCause.Object["fields"].Object["satellite_id"] = native.EngineJSONValue{Kind: native.EngineJSONString, String: "G00"}
	nativeValue.TypedFields["cause"] = nativeCause
	nativeNames := nativeValue.TypedFields["names"]
	nativeNames.Array[0].String = "mutated"
	nativeValue.TypedFields["names"] = nativeNames
	if value.TypedFields["cause"].Object["fields"].Object["satellite_id"].String != "G99" || value.TypedFields["names"].Array[0].String != "one" {
		t.Fatalf("public typed payload aliases native values: cause=%+v names=%+v", value.TypedFields["cause"], value.TypedFields["names"])
	}
}

func TestStatusErrorRetainsAllExistingAttachments(t *testing.T) {
	ee := &EngineError{Family: EngineErrorFamilyRtk, FamilyName: "rtk", Kind: "empty_epochs"}
	sp3 := &SP3Error{Kind: SP3ErrorKindExactValidation}
	td := &TerrainDatumError{Message: "datum error"}
	ts := &TerrainStoreError{Message: "store error"}
	bias := &BiasError{Kind: BiasErrorInvalidInput}
	rtcm := &RTCMError{Class: RTCMErrorClassEncode}

	statusErr := &StatusError{
		Code:         StatusInvalidArgument,
		Text:         "invalid argument",
		Detail:       "test detail",
		Engine:       ee,
		SP3:          sp3,
		TerrainDatum: td,
		TerrainStore: ts,
		Bias:         bias,
		RTCM:         rtcm,
	}

	// 1. First validate graph acyclicity with bounded walk of both Unwrap forms and visited set
	assertAcyclicGraph(t, statusErr, 50)

	// 2. Verify errors.As works for each attached sibling type
	var foundEE *EngineError
	if !errors.As(statusErr, &foundEE) || foundEE != ee {
		t.Errorf("errors.As for EngineError failed: got %v", foundEE)
	}

	var foundSP3 *SP3Error
	if !errors.As(statusErr, &foundSP3) || foundSP3 != sp3 {
		t.Errorf("errors.As for SP3Error failed: got %v", foundSP3)
	}

	var foundTD *TerrainDatumError
	if !errors.As(statusErr, &foundTD) || foundTD != td {
		t.Errorf("errors.As for TerrainDatumError failed: got %v", foundTD)
	}

	var foundTS *TerrainStoreError
	if !errors.As(statusErr, &foundTS) || foundTS != ts {
		t.Errorf("errors.As for TerrainStoreError failed: got %v", foundTS)
	}

	var foundBias *BiasError
	if !errors.As(statusErr, &foundBias) || foundBias != bias {
		t.Errorf("errors.As for BiasError failed: got %v", foundBias)
	}

	var foundRTCM *RTCMError
	if !errors.As(statusErr, &foundRTCM) || foundRTCM != rtcm {
		t.Errorf("errors.As for RTCMError failed: got %v", foundRTCM)
	}

	// 3. Verify unmatched Is and missing-type As terminate safely
	if errors.Is(statusErr, ErrClosed) {
		t.Error("unexpected match for ErrClosed")
	}
	var absent *absentTestError
	if errors.As(statusErr, &absent) {
		t.Error("unexpected match for absentTestError")
	}
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

func TestPublicEngineErrorFamilyRegistry(t *testing.T) {
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
			t.Errorf("public family %s numeric ID = %d, expected %d", tc.expected, uint32(tc.family), tc.expectedID)
		}
		if tc.family != EngineErrorFamily(tc.expectedID) {
			t.Errorf("public family %s cast from ID %d = %d, expected %d", tc.expected, tc.expectedID, tc.family, EngineErrorFamily(tc.expectedID))
		}
		if tc.family.Name() != tc.expected {
			t.Errorf("public family %d name = %q, expected %q", tc.family, tc.family.Name(), tc.expected)
		}
		if tc.family.String() != tc.expected {
			t.Errorf("public family %d string = %q, expected %q", tc.family, tc.family.String(), tc.expected)
		}
		roundTrip := EngineErrorFamilyFromName(tc.expected)
		if roundTrip != tc.family {
			t.Errorf("public round-trip from name %q = %d, expected %d", tc.expected, roundTrip, tc.family)
		}
	}

	for id := uint32(0); id <= 72; id++ {
		fam := EngineErrorFamily(id)
		if fam.Name() == fmt.Sprintf("family_%d", id) {
			t.Errorf("contiguous public family ID %d has unmapped name %q", id, fam.Name())
		}
	}

	if EngineErrorFamilyFromName("unrecognized_subsystem") != EngineErrorFamilyUnknown {
		t.Errorf("expected unrecognized subsystem to map to EngineErrorFamilyUnknown")
	}

	// Verify future codes produce raw family_<N> names and string representation.
	futureCode73 := EngineErrorFamily(73)
	if futureCode73.Name() != "family_73" {
		t.Errorf("future family 73 name = %q, expected \"family_73\"", futureCode73.Name())
	}
	if futureCode73.String() != "family_73" {
		t.Errorf("future family 73 string = %q, expected \"family_73\"", futureCode73.String())
	}
	futureCode9999 := EngineErrorFamily(9999)
	if futureCode9999.Name() != "family_9999" {
		t.Errorf("future family 9999 name = %q, expected \"family_9999\"", futureCode9999.Name())
	}
}
