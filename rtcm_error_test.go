package sidereon

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestPublicStatusErrorPreservesRTCMErrorPayload(t *testing.T) {
	payload := []byte(`{"schema_version":1,"error":{"variant":"FieldOutOfRange","value":"170141183460469231731687303715884105727"}}`)
	err := publicError(&native.StatusError{
		Code: 2,
		Text: "invalid argument",
		RTCM: &native.RTCMError{Class: 4, Kind: 1, Payload: payload},
	})
	payload[0] = ' '
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.RTCM == nil {
		t.Fatalf("expected typed RTCM status error, got %T", err)
	}
	if statusErr.RTCM.Class != RTCMErrorClassSBASEncode || statusErr.RTCM.Kind != 1 {
		t.Fatalf("unexpected RTCM discriminants: class=%d kind=%d", statusErr.RTCM.Class, statusErr.RTCM.Kind)
	}
	var structured *RTCMError
	if !errors.As(err, &structured) || !json.Valid(structured.Payload) {
		t.Fatalf("typed payload was not discoverable or valid JSON: %s", structured.Payload)
	}
	if !strings.Contains(string(structured.Payload), `"170141183460469231731687303715884105727"`) {
		t.Fatalf("large integer payload was not preserved exactly: %s", structured.Payload)
	}
	if string(structured.Payload) == string(payload) {
		t.Fatal("public error payload aliases native storage")
	}
}

func TestPublicStatusErrorPreservesUnknownRTCMKind(t *testing.T) {
	err := publicError(&native.StatusError{
		Code: 2,
		Text: "invalid argument",
		RTCM: &native.RTCMError{Class: 3, Kind: RTCMErrorKindUnknown, Payload: []byte(`{"schema_version":1,"error":{"variant":"Unknown"}}`)},
	})
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.RTCM == nil {
		t.Fatalf("expected typed RTCM status error, got %T", err)
	}
	if statusErr.RTCM.Class != RTCMErrorClassOther || statusErr.RTCM.Kind != RTCMErrorKindUnknown {
		t.Fatalf("unexpected unknown RTCM discriminants: %+v", statusErr.RTCM)
	}
}

func TestPublicStatusErrorPreservesEveryRTCMEncodeVariantPayload(t *testing.T) {
	cases := []struct {
		kind    uint32
		variant string
		error   string
	}{
		{1, "FieldOutOfRange", `{"variant":"FieldOutOfRange","message_number":"1005","field":"ecef_x","value":"-2","width":"38","encoding":"TwosComplement"}`},
		{2, "NegativeZeroWithValue", `{"variant":"NegativeZeroWithValue","message_number":"1020","field":"tau_n","value":"7"}`},
		{3, "NegativeZeroMask", `{"variant":"NegativeZeroMask","message_number":"1020","mask":"9"}`},
		{4, "MessageNumber", `{"variant":"MessageNumber","message_number":"999","record":{"variant":"StationCoordinates"}}`},
		{5, "FieldPresence", `{"variant":"FieldPresence","message_number":"1005","record":{"variant":"StationCoordinates"},"field":"antenna_height","carried":false}`},
		{6, "SatelliteFieldPresence", `{"variant":"SatelliteFieldPresence","message_number":"1074","record":{"variant":"Msm","system":"GPS","kind":"MSM4"},"satellite":"7","field":"extended_info","carried":false}`},
		{7, "CountMismatch", `{"variant":"CountMismatch","message_number":"1015","field":"satellites","expected":"2","actual":"1"}`},
		{8, "ValueOutOfRange", `{"variant":"ValueOutOfRange","message_number":"1005","field":"itrf","value":"64","minimum":"0","maximum":"63"}`},
		{9, "NonLatin1Character", `{"variant":"NonLatin1Character","field":"descriptor","character":"λ","codepoint":"955"}`},
		{10, "SatelliteIdOutOfRange", `{"variant":"SatelliteIdOutOfRange","message_number":"1019","field":"GPS PRN","value":"64","width":"6"}`},
		{11, "SsrSatelliteIdOutOfRange", `{"variant":"SsrSatelliteIdOutOfRange","message_number":"1057","value":"64","width":"6"}`},
		{12, "SsrRecordsNotCarried", `{"variant":"SsrRecordsNotCarried","message_number":"1058","kind":"Clock","records":"orbit","count":"2"}`},
		{13, "SsrCombinedRecordCounts", `{"variant":"SsrCombinedRecordCounts","message_number":"1060","orbit":"2","clock":"1"}`},
		{14, "SsrCombinedSatelliteMismatch", `{"variant":"SsrCombinedSatelliteMismatch","message_number":"1060","index":"1","orbit_satellite":"4","clock_satellite":"5"}`},
		{15, "SsrHighRateClockTerms", `{"variant":"SsrHighRateClockTerms","message_number":"1062","satellite":"3","c1":"-4","c2":"5"}`},
		{16, "SsrSatelliteCount", `{"variant":"SsrSatelliteCount","message_number":"1057","declared":"2","records":"1"}`},
		{17, "MsmMask", `{"variant":"MsmMask","message_number":"1074","problem":{"variant":"SignalNotInMask","signal":"3","mask":"5"}}`},
		{18, "MsmOptional", `{"variant":"MsmOptional","message_number":"1077","kind":"MSM7","satellite":"4","signal":"6","field":"FinePhaseRangeRate","problem":{"variant":"InvalidValue","value":"-16384"}}`},
		{19, "TrailingZeroBits", `{"variant":"TrailingZeroBits","message_number":"1006","bits":"3"}`},
		{20, "StrictDeparture", `{"variant":"StrictDeparture","departure":{"variant":"FrameReservedBits","reserved":"5"}}`},
		{21, "UnsupportedBodyTooShort", `{"variant":"UnsupportedBodyTooShort","message_number":"4090"}`},
		{22, "UnsupportedBodyNumber", `{"variant":"UnsupportedBodyNumber","message_number":"4090","carried":"4089"}`},
		{23, "UnsupportedDecodedNumber", `{"variant":"UnsupportedDecodedNumber","message_number":"4090"}`},
		{24, "FrameBodyTooLong", `{"variant":"FrameBodyTooLong","len":"1024"}`},
		{25, "FrameReservedOutOfRange", `{"variant":"FrameReservedOutOfRange","value":"64"}`},
	}

	for _, tc := range cases {
		t.Run(tc.variant, func(t *testing.T) {
			want := `{"schema_version":1,"error":` + tc.error + `}`
			nativePayload := []byte(want)
			err := publicError(&native.StatusError{
				Code: 2,
				Text: "invalid argument",
				RTCM: &native.RTCMError{Class: uint32(RTCMErrorClassEncode), Kind: tc.kind, Payload: nativePayload},
			})
			var structured *RTCMError
			if !errors.As(err, &structured) {
				t.Fatalf("expected public RTCMError, got %T", err)
			}
			if structured.Class != RTCMErrorClassEncode || structured.Kind != tc.kind {
				t.Fatalf("unexpected discriminants: class=%d kind=%d", structured.Class, structured.Kind)
			}
			if got, want := structured.Error(), fmt.Sprintf("sidereon: RTCM error class %d kind %d", RTCMErrorClassEncode, tc.kind); got != want {
				t.Fatalf("Error()=%q, want %q", got, want)
			}
			if !json.Valid(structured.Payload) || string(structured.Payload) != want {
				t.Fatalf("payload changed across the public route: %s", structured.Payload)
			}
			var payload struct {
				Error struct {
					Variant string `json:"variant"`
				} `json:"error"`
			}
			if unmarshalErr := json.Unmarshal(structured.Payload, &payload); unmarshalErr != nil {
				t.Fatal(unmarshalErr)
			}
			if payload.Error.Variant != tc.variant {
				t.Fatalf("variant=%q, want %q", payload.Error.Variant, tc.variant)
			}
			nativePayload[0] = ' '
			if string(structured.Payload) != want {
				t.Fatal("public payload aliases native storage")
			}
		})
	}
}
