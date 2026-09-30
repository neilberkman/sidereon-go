package sidereon

import (
	"math"
	"testing"
)

func TestInertialModelConstantsMatchWGS84Reference(t *testing.T) {
	got, err := InertialModelConstants()
	if err != nil {
		t.Fatal(err)
	}
	// USGS Open-File Report 2006-1204, WGS 84 parameter table: gamma_e,
	// gamma_p, and flattening. The API derives k as (1-f)*gamma_p/gamma_e-1;
	// its tolerance includes the report's rounded flattening precision.
	const gammaEquator = 9.7803253359
	const gammaPole = 9.8321849378
	const flattening = 0.00335281066474
	if math.IsNaN(got.NormalGravityEquatorMPerS2) || math.IsInf(got.NormalGravityEquatorMPerS2, 0) ||
		math.Abs(got.NormalGravityEquatorMPerS2-gammaEquator) > 5e-11 {
		t.Errorf("equatorial gravity %.17g differs from NIMA WGS 84 reference %.10g", got.NormalGravityEquatorMPerS2, gammaEquator)
	}
	if math.IsNaN(got.NormalGravityPoleMPerS2) || math.IsInf(got.NormalGravityPoleMPerS2, 0) ||
		math.Abs(got.NormalGravityPoleMPerS2-gammaPole) > 5e-11 {
		t.Errorf("polar gravity %.17g differs from NIMA WGS 84 reference %.10g", got.NormalGravityPoleMPerS2, gammaPole)
	}
	wantK := (1-flattening)*gammaPole/gammaEquator - 1
	kTolerance := 1e-14*(gammaPole/gammaEquator) + 8*2.220446049250313e-16
	if math.IsNaN(got.SomiglianaK) || math.IsInf(got.SomiglianaK, 0) ||
		math.Abs(got.SomiglianaK-wantK) > kTolerance {
		t.Errorf("Somigliana k %.17g differs from independently derived WGS 84 value %.17g", got.SomiglianaK, wantK)
	}
	if got.DefaultIMUSimulatorSeed != 0x4d595df4d0f33173 {
		t.Errorf("default IMU seed = %#x, want pinned deterministic seed %#x", got.DefaultIMUSimulatorSeed, uint64(0x4d595df4d0f33173))
	}
}

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
