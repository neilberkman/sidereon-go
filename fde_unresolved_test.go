package sidereon

import (
	"errors"
	"math"
	"reflect"
	"runtime"
	"slices"
	"testing"
)

func TestFDEUnresolvedOwnsCompletePayloadAcrossNextProducer(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	clean := usedSPPConfig()
	initial, err := SolveSPP(sp3, clean)
	if err != nil {
		t.Fatal(err)
	}
	if len(initial.UsedSatelliteIDs) == 0 {
		t.Fatal("clean SPP returned no used satellites")
	}
	faulted := usedSPPConfig()
	injected := false
	for i := range faulted.Observations {
		if faulted.Observations[i].SatelliteID == initial.UsedSatelliteIDs[0] {
			faulted.Observations[i].PseudorangeM += 1000
			injected = true
			break
		}
	}
	if !injected {
		t.Fatalf("selected SPP satellite %q is absent from the faulted input", initial.UsedSatelliteIDs[0])
	}
	zero := uint64(0)
	options := FDEOptions{PFA: 1e-3, MaxExclusions: &zero, WeightsMode: FDEWeightsUnit}
	for _, cap := range []float64{0, math.NaN()} {
		_, capErr := SolveFDE(sp3, clean, FDEOptions{
			PFA: 1e-3, MaxExclusions: &zero, MaxExclusionRMSM: &cap,
			WeightsMode: FDEWeightsUnit,
		})
		var quality *QualityError
		if !errors.As(capErr, &quality) || quality.Kind != QualityErrorInvalidParameter {
			t.Fatalf("invalid cap %v with zero exclusion budget = %T %v", cap, capErr, capErr)
		}
	}
	infiniteCap := math.Inf(1)
	_, err = SolveFDE(sp3, clean, FDEOptions{
		PFA: 1e-3, MaxExclusions: &zero, MaxExclusionRMSM: &infiniteCap,
		WeightsMode: FDEWeightsUnit,
	})
	var infiniteCapUnresolved *FDEUnresolvedError
	if !errors.As(err, &infiniteCapUnresolved) || infiniteCapUnresolved.Reason != FDEUnresolvedExclusionBudgetExhausted {
		t.Fatalf("positive-infinity cap should pass option validation and retain the resulting unresolved error: %T %v", err, err)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	_, err = SolveFDE(sp3, faulted, options)
	var unresolved *FDEUnresolvedError
	if !errors.As(err, &unresolved) {
		t.Fatalf("faulted zero-budget solve = %T %v, want *FDEUnresolvedError", err, err)
	}
	if unresolved.Reason != FDEUnresolvedExclusionBudgetExhausted || unresolved.CaptureError != nil {
		t.Fatalf("unresolved reason/capture = %d, %v", unresolved.Reason, unresolved.CaptureError)
	}
	if !unresolved.HasSolution || !unresolved.HasExcluded || !unresolved.HasRAIM || !unresolved.HasNormalizedResiduals {
		t.Fatalf("unresolved payload presence = solution:%t excluded:%t RAIM:%t rows:%t", unresolved.HasSolution, unresolved.HasExcluded, unresolved.HasRAIM, unresolved.HasNormalizedResiduals)
	}
	if unresolved.Solution.UsedSatelliteCount < 4 || unresolved.Solution.UsedSatelliteCount != len(unresolved.Solution.UsedSatelliteIDs) || len(unresolved.Excluded) != 0 || unresolved.Iterations != 0 {
		t.Fatalf("unresolved solution/exclusions = used:%d IDs:%d excluded:%v iterations:%d", unresolved.Solution.UsedSatelliteCount, len(unresolved.Solution.UsedSatelliteIDs), unresolved.Excluded, unresolved.Iterations)
	}
	if len(unresolved.Solution.PseudorangeVariancesM2) != unresolved.Solution.UsedSatelliteCount || len(unresolved.Solution.Weights) != unresolved.Solution.UsedSatelliteCount {
		t.Fatalf("unresolved solution arrays are not aligned: used=%d variances=%d weights=%d", unresolved.Solution.UsedSatelliteCount, len(unresolved.Solution.PseudorangeVariancesM2), len(unresolved.Solution.Weights))
	}
	if !unresolved.RAIM.FaultDetected || len(unresolved.NormalizedResiduals) != unresolved.Solution.UsedSatelliteCount {
		t.Fatalf("unresolved RAIM = %+v rows:%d", unresolved.RAIM, len(unresolved.NormalizedResiduals))
	}
	saved := *unresolved
	saved.Excluded = slices.Clone(unresolved.Excluded)
	saved.NormalizedResiduals = slices.Clone(unresolved.NormalizedResiduals)
	saved.Solution.UsedSatelliteIDs = slices.Clone(unresolved.Solution.UsedSatelliteIDs)
	saved.Solution.ResidualsM = slices.Clone(unresolved.Solution.ResidualsM)
	saved.Solution.PseudorangeVariancesM2 = slices.Clone(unresolved.Solution.PseudorangeVariancesM2)
	saved.Solution.Weights = slices.Clone(unresolved.Solution.Weights)
	if unresolved.Solution.DOP != nil {
		dop := *unresolved.Solution.DOP
		saved.Solution.DOP = &dop
	}
	if unresolved.Solution.Geodetic != nil {
		geodetic := *unresolved.Solution.Geodetic
		saved.Solution.Geodetic = &geodetic
	}
	if !reflect.DeepEqual(saved, *unresolved) {
		t.Fatal("detached unresolved payload copy differs before the next producer")
	}
	cleanSolution, err := SolveSPP(sp3, clean)
	if err != nil {
		t.Fatalf("subsequent successful SPP producer on the same thread: %v", err)
	}
	if len(cleanSolution.UsedSatelliteIDs) == 0 {
		t.Fatal("subsequent SPP producer returned no used satellites")
	}
	if !reflect.DeepEqual(saved, *unresolved) {
		t.Fatalf("owned unresolved payload changed after successful SPP producer: before=%+v after=%+v", saved, *unresolved)
	}
}
