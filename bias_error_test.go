package sidereon

import (
	"errors"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestBiasErrorPublicConversionPreservesTypedPayloadAndBytes(t *testing.T) {
	message := []byte{0xff, 0x00, 'x'}
	status := &native.StatusError{
		Code: 2, Text: "invalid argument", Bias: &native.BiasError{
			Kind: 9, Line: 14, HasTimeScale: true, TimeScale: 3,
			Departure: native.BiasNotice{Kind: 0, Departure: 2, HasLine: true, Line: 14},
			Message:   message, Version: []byte("2.00"),
			DepartureText: [9][]byte{nil, nil, nil, nil, nil, nil, []byte("2.00")},
		},
	}
	err := publicError(status)
	var publicStatus *StatusError
	if !errors.As(err, &publicStatus) || publicStatus.Bias == nil {
		t.Fatalf("typed bias payload missing from status error: %T %v", err, err)
	}
	var biasErr *BiasError
	if !errors.As(err, &biasErr) {
		t.Fatalf("BiasError not discoverable through errors.As: %T %v", err, err)
	}
	if biasErr.Kind != BiasErrorDeparture || biasErr.Line != 14 || !biasErr.HasTimeScale || biasErr.TimeScale != 3 || biasErr.Departure.Departure != BiasDepartureOtherVersion {
		t.Fatalf("typed bias fields = %+v", biasErr)
	}
	message[0] = 0
	got, err := biasErr.TextBytes(BiasErrorTextMessage, 0)
	if err != nil || len(got) != 3 || got[0] != 0xff || got[1] != 0 || got[2] != 'x' {
		t.Fatalf("message bytes = %v, %v", got, err)
	}
	got[0] = 0
	again, err := biasErr.TextBytes(BiasErrorTextDepartureNotice, BiasNoticeTextVersion)
	if err != nil || string(again) != "2.00" {
		t.Fatalf("departure version bytes = %q, %v", again, err)
	}
}
