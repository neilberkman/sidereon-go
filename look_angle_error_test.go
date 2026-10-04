package sidereon

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestLookAnglesPreservesTypedInvalidInputDetail(t *testing.T) {
	tle, err := ParseTLE(
		"1 25544U 98067A   18184.80969102  .00001614  00000-0  31745-4 0  9993",
		"2 25544  51.6414 295.8524 0003435 262.6267 204.2868 15.54005638121106",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tle.Close(); err != nil {
			t.Errorf("Close TLE: %v", err)
		}
	}()

	_, err = tle.LookAngles(PassStation{LatitudeDeg: 150}, []time.Time{
		time.Date(2018, time.July, 4, 0, 0, 0, 0, time.UTC),
	})
	assertLookAngleEngineError(t, err, StatusInvalidArgument, "invalid_input", func(fields json.RawMessage) {
		var got struct {
			Field  string `json:"field"`
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal(fields, &got); err != nil {
			t.Fatalf("decode InvalidInput fields %s: %v", fields, err)
		}
		if got.Field != "ground_station.latitude_deg" || got.Reason != "out of range" {
			t.Fatalf("InvalidInput fields = %+v", got)
		}
	})
}

func TestLookAnglesPreservesNestedSGP4PropagationDetail(t *testing.T) {
	tle, err := ParseTLE(
		"1 28872U 05037B   05333.02012661  .25992681  00000-0  24476-3 0  1534",
		"2 28872  96.4736 157.9986 0303955 244.0492 110.6523 16.46015938 10708",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tle.Close(); err != nil {
			t.Errorf("Close TLE: %v", err)
		}
	}()

	epoch := time.Date(2005, time.November, 29, 0, 28, 59, 0, time.UTC)
	_, err = tle.LookAngles(PassStation{}, []time.Time{epoch.Add(24 * time.Hour)})
	assertLookAngleEngineError(t, err, StatusSolve, "propagate", func(fields json.RawMessage) {
		var got struct {
			Cause struct {
				Kind   string `json:"kind"`
				Fields struct {
					Code int `json:"code"`
				} `json:"fields"`
			} `json:"cause"`
		}
		if err := json.Unmarshal(fields, &got); err != nil {
			t.Fatalf("decode Propagate fields %s: %v", fields, err)
		}
		if got.Cause.Kind != "sgp4" || got.Cause.Fields.Code != 6 {
			t.Fatalf("Propagate fields = %+v", got)
		}
	})
}

func assertLookAngleEngineError(t *testing.T, err error, wantCode StatusCode, wantKind string, assertFields func(json.RawMessage)) {
	t.Helper()
	if err == nil {
		t.Fatal("LookAngles unexpectedly succeeded")
	}
	var status *StatusError
	if !errors.As(err, &status) {
		t.Fatalf("error = %T %v, want *StatusError", err, err)
	}
	if status.Code != wantCode || status.Engine == nil {
		t.Fatalf("status = %+v, want code %v and typed engine detail", status, wantCode)
	}
	engine := status.Engine
	if engine.FamilyName != "look_angle" || engine.Operation != "sidereon_tle_look_angles" || engine.Kind != wantKind {
		t.Fatalf("engine error = family %q operation %q kind %q, want look_angle/sidereon_tle_look_angles/%s", engine.FamilyName, engine.Operation, engine.Kind, wantKind)
	}
	if engine.Schema != 1 || len(engine.Payload) == 0 || engine.CaptureError != nil {
		t.Fatalf("engine payload = schema %d payload %s capture error %v", engine.Schema, engine.Payload, engine.CaptureError)
	}
	assertFields(engine.Fields)
}
