package sidereon

import (
	"errors"
	"math"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestRAIMPreservesQualityErrorKinds(t *testing.T) {
	ids := []string{"G01", "G02", "G03", "G04", "G05"}
	residuals := []float64{0, 0, 0, 0, 0}
	_, _, err := RAIM(RAIMInput{SatelliteIDs: ids, ResidualsM: residuals}, RAIMOptions{})
	var quality *QualityError
	if !errors.As(err, &quality) || quality.Kind != QualityErrorMissingVariances {
		t.Fatalf("missing-variance error = %T %v", err, err)
	}
	_, _, err = RAIM(RAIMInput{SatelliteIDs: ids, ResidualsM: residuals, VariancesM2: []float64{1, 1, math.NaN(), 1, 1}}, RAIMOptions{PFA: 1e-3})
	if !errors.As(err, &quality) || quality.Kind != QualityErrorInvalidVariance {
		t.Fatalf("invalid-variance error = %T %v", err, err)
	}
	if _, _, err := RAIM(RAIMInput{SatelliteIDs: ids, ResidualsM: residuals}, RAIMOptions{WeightsMode: FDEWeightsUnit}); err != nil {
		t.Fatalf("unit-weight RAIM without variances: %v", err)
	}
	invalidVariances := []float64{math.NaN()}
	if _, _, err := RAIM(RAIMInput{SatelliteIDs: ids, ResidualsM: residuals, VariancesM2: invalidVariances}, RAIMOptions{WeightsMode: FDEWeightsUnit}); err != nil {
		t.Fatalf("unit-weight RAIM should ignore variance input: %v", err)
	}
	if _, _, err := RAIM(RAIMInput{SatelliteIDs: ids, ResidualsM: residuals, VariancesM2: invalidVariances}, RAIMOptions{WeightsMode: FDEWeightsBySatellite, Weights: map[string]float64{"G01": 1}}); err != nil {
		t.Fatalf("by-satellite RAIM should ignore variance input: %v", err)
	}
	_, _, err = RAIM(RAIMInput{SatelliteIDs: ids, ResidualsM: residuals, VariancesM2: invalidVariances}, RAIMOptions{WeightsMode: FDEWeightsSolution})
	if !errors.As(err, &quality) || quality.Kind != QualityErrorInvalidVariance {
		t.Fatalf("solution-weight variance-length error = %T %v", err, err)
	}
	_, _, err = RAIM(RAIMInput{SatelliteIDs: ids, ResidualsM: residuals, VariancesM2: invalidVariances}, RAIMOptions{PFA: 1, WeightsMode: FDEWeightsSolution})
	if !errors.As(err, &quality) || quality.Kind != QualityErrorInvalidProbability {
		t.Fatalf("invalid probability precedence = %T %v", err, err)
	}
}

func TestPublicErrorKeepsUnresolvedRecordWhenRAIMCountCannotConvert(t *testing.T) {
	input := &native.NativeFDEUnresolvedError{
		Reason:      1,
		HasExcluded: true,
		Excluded:    []string{"G08"},
		HasRAIM:     true,
		RAIM: native.NativeRaimResult{
			FaultDetected: true, TestStatistic: 12.5, HasThreshold: true, Threshold: 4.0,
			HasReducedChiSquare: true, ReducedChiSquare: 3.25, RMSM: 2.5, DOF: 4,
			Testable: true, NormalizedResidualCount: ^uint64(0),
			HasWorstSatellite: true, WorstSatellite: "G08",
		},
		Cause: &native.StatusError{Code: 5, Text: "fault unresolved"},
	}
	got := publicError(input)
	var unresolved *FDEUnresolvedError
	if !errors.As(got, &unresolved) {
		t.Fatalf("translated error = %T %v, want FDEUnresolvedError", got, got)
	}
	if unresolved.Reason != FDEUnresolvedExclusionBudgetExhausted || !unresolved.HasExcluded || len(unresolved.Excluded) != 1 || !unresolved.HasRAIM || unresolved.CaptureError == nil {
		t.Fatalf("translated partial unresolved record = %+v", unresolved)
	}
	if !unresolved.RAIM.FaultDetected || unresolved.RAIM.TestStatistic != 12.5 || !unresolved.RAIM.HasThreshold || unresolved.RAIM.Threshold != 4 || !unresolved.RAIM.HasReducedChiSquare || unresolved.RAIM.ReducedChiSquare != 3.25 || unresolved.RAIM.RMSM != 2.5 || unresolved.RAIM.DOF != 4 || !unresolved.RAIM.Testable || !unresolved.RAIM.HasWorstSatellite || unresolved.RAIM.WorstSatellite != "G08" {
		t.Fatalf("RAIM summary lost when row count conversion failed: %+v", unresolved.RAIM)
	}
}
