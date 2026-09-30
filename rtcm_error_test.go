package sidereon

import (
	"encoding/json"
	"errors"
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
