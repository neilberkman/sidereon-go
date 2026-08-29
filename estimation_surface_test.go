package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func TestEstimationSurfaceFrozenOutputs(t *testing.T) {
	step, err := AlphaBetaFilterStep(AlphaBetaState{Level: 1, Rate: 2}, 10, 0.5, AlphaBetaGains{Alpha: 0.2, Beta: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	if step.Predicted != (AlphaBetaState{Level: 2, Rate: 2}) || step.Updated != (AlphaBetaState{Level: 3.6, Rate: 3.6}) || step.Innovation != 8 {
		t.Fatalf("alpha-beta output = %+v, want frozen C output", step)
	}
	if value, err := NormalizedInnovation(3, 2); err != nil || value != 3/math.Sqrt2 {
		t.Fatalf("normalized innovation = %.17g, %v; want 3/sqrt(2)", value, err)
	}
	if value, err := EWMA(2, 6, 0.25); err != nil || value != 3 {
		t.Fatalf("EWMA = %.17g, %v; want 3", value, err)
	}
	if value, err := NIS(3, 2); err != nil || math.Abs(value-4.5) > 1e-14 {
		t.Fatalf("NIS = %.17g, %v; want 4.5", value, err)
	}
	symmetric, err := CovarianceIsSymmetric([3][3]float64{{1, 0, 0}, {0, 2, 0}, {0, 0, 3}})
	if err != nil || !symmetric {
		t.Fatalf("symmetric = %v, %v", symmetric, err)
	}
	positive, err := CovarianceIsPositiveSemidefinite([3][3]float64{{1, 0, 0}, {0, 2, 0}, {0, 0, 3}})
	if err != nil || !positive {
		t.Fatalf("positive semidefinite = %v, %v", positive, err)
	}
	covariance, err := NormalCovariance([]float64{1, 0, 0, 0, 2}, 2, 2, 3)
	if err == nil || covariance != nil {
		// The shape seam is intentionally checked before C; the valid frozen
		// matrix is exercised below.
		t.Fatalf("malformed Jacobian result = %#v, %v", covariance, err)
	}
	covariance, err = NormalCovariance([]float64{1, 0, 0, 2}, 2, 2, 3)
	if err != nil || len(covariance) != 4 || covariance[0] != 3 || covariance[3] != 0.75 {
		t.Fatalf("normal covariance = %#v, %v", covariance, err)
	}
	if _, err := HessianTrace([]float64{1, 0, 0, 2}, 2, 2); err != nil {
		t.Fatal(err)
	}
}

func TestEstimationValidationAndOwnership(t *testing.T) {
	if _, err := NormalCovariance([]float64{1, 2}, -1, 2, 1); err == nil {
		t.Fatal("negative matrix shape was accepted")
	}
	if _, err := NormalCovariance(nil, int(^uint(0)>>1), 2, 1); err == nil {
		t.Fatal("overflowing matrix shape was accepted")
	}
	if _, err := WeightVector([]WeightEntry{{SatelliteID: "G\x00 1"}}, PseudorangeVarianceOptions{}); err == nil {
		t.Fatal("embedded NUL satellite ID was accepted")
	}
	if _, _, err := FusionVelocityMatchOutage(nil, FusionLooseMeasurement{SatellitesUsed: -1}, FusionVelocityMatchingConfig{}); err == nil {
		t.Fatal("negative fusion satellite count was accepted")
	}
	if _, err := LambdaILS([]float64{1, 2}, []float64{1, 0, 0, 1}, 3); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RAIM(RAIMInput{SatelliteIDs: []string{"G01", "G02", "G03", "G04", "G05"}, ResidualsM: []float64{0, 0, 0, 0, 0}}, RAIMOptions{PFA: 0.001, UnitWeights: true}); err != nil {
		t.Fatal(err)
	}
	sigmas, err := PseudorangeSigmas([]WeightEntry{{SatelliteID: "G01", ElevationDeg: 90}}, PseudorangeVarianceOptions{AM: 1, BM: 1, Model: PseudorangeVarianceElevation})
	if err != nil || len(sigmas) != 1 || !sigmas[0].Present || sigmas[0].Value == 0 {
		t.Fatalf("present pseudorange sigma = %#v, %v", sigmas, err)
	}
	values := []float64{1, 2, 3}
	copyOfValues := append([]float64(nil), values...)
	if _, err := MADSpread(values, 1); err != nil {
		t.Fatal(err)
	}
	values[0] = 99
	if !equalFloatSlices(values, []float64{99, 2, 3}) || !equalFloatSlices(copyOfValues, []float64{1, 2, 3}) {
		t.Fatal("test ownership setup changed unexpectedly")
	}
}

func TestEstimationHandleConcurrencyAndClose(t *testing.T) {
	position := []float64{1, 2, 3}
	covariance := []float64{1, 0, 0, 0, 1, 0, 0, 0, 1}
	filter, err := NewTrackFilterFromPosition(TrackECEF, 0, position, covariance, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	position[0], covariance[0] = 99, 99
	if got, readErr := filter.Position(); readErr != nil || len(got) != 3 || got[0] != 1 {
		t.Fatalf("constructor input ownership = %#v, %v", got, readErr)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, readErr := filter.Position()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("Position: %v", readErr)
			}
		}()
	}
	for i := 0; i < 2; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if closeErr := filter.Close(); closeErr != nil {
				t.Errorf("Close: %v", closeErr)
			}
		}()
	}
	group.Wait()
	if err := filter.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := filter.Position(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Position after Close = %v, want ErrClosed", err)
	}
}

func equalFloatSlices(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
