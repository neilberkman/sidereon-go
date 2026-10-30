package sidereon

import (
	"errors"
	"testing"
)

func TestInitialOrbitGaussAnglesInvalidInputReturnsTypedError(t *testing.T) {
	_, err := InitialOrbitGaussAngles(
		[3]float64{},
		[3]float64{},
		[3]float64{2_451_545, 2_451_545, 2_451_545},
		[3]float64{},
		[3][3]float64{},
	)
	if err == nil {
		t.Fatal("InitialOrbitGaussAngles accepted zero line-of-sight and site vectors")
	}
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("InitialOrbitGaussAngles error type = %T, want *StatusError", err)
	}
	engineErr := statusErr.EngineError()
	if engineErr == nil || engineErr.Family != EngineErrorFamilyIOD || engineErr.Kind != "invalid_time_geometry" {
		t.Fatalf("InitialOrbitGaussAngles engine error = %#v, want iod/invalid_time_geometry", engineErr)
	}
}
