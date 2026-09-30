package sidereon

import "testing"

func TestInertialValueAdaptersRetainComponents(t *testing.T) {
	state := InertialNavState{EpochJ2000S: 12.25, PositionECEFM: [3]float64{1, 2, 3}, VelocityECEFMPerS: [3]float64{4, 5, 6}, AttitudeBodyToECEF: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}, AccelBiasMPerS2: [3]float64{7, 8, 9}, GyroBiasRadPerS: [3]float64{10, 11, 12}}
	if roundTrip := publicInertialState(nativeInertialState(state)); roundTrip != state {
		t.Fatalf("state adapter changed fields: %+v", roundTrip)
	}
	increment := InertialIncrement{EpochJ2000S: 20, DeltaVelocityMPerS: [3]float64{1, 2, 3}, DeltaThetaRad: [3]float64{4, 5, 6}, DTS: 0.01}
	if roundTrip := publicInertialIncrement(nativeInertialIncrement(increment)); roundTrip != increment {
		t.Fatalf("increment adapter changed fields: %+v", roundTrip)
	}
}
