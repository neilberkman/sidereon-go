package sidereon

import (
	"os"
	"testing"
)

func TestSP3RemainingRoutesFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	epochs, err := sp3.Epochs()
	if err != nil || len(epochs) == 0 {
		t.Fatalf("epochs=%v err=%v", epochs, err)
	}
	satellites, err := sp3.Satellites()
	if err != nil || len(satellites) == 0 {
		t.Fatalf("satellites=%v err=%v", satellites, err)
	}

	request, requestErr := NewExactSP3Request(2020, 6, 24, "1145", "15M", "01D", "TEST")
	if requestErr != nil {
		t.Fatal(requestErr)
	}
	t.Cleanup(func() {
		if err := request.Close(); err != nil {
			t.Errorf("close exact request: %v", err)
		}
	})
	exactData := append(append([]byte(nil), data...), []byte("\nEOF\n")...)
	loaded, coverageErr, loadErr := LoadExactSP3(exactData, request)
	if loadErr == nil {
		t.Errorf("exact load unexpectedly succeeded with coverage %v", coverageErr)
	}
	if loaded != nil {
		t.Cleanup(func() {
			if err := loaded.Close(); err != nil {
				t.Errorf("close exact SP3: %v", err)
			}
		})
	}
	_, validateErr := sp3.ValidateExact(request)
	if validateErr == nil {
		t.Error("exact validate unexpectedly succeeded for fixture without EOF")
	}
	if _, err := ExactSP3RequestFromIdentity(ProductIdentity{}); err == nil {
		t.Error("empty product identity unexpectedly accepted")
	}

	if _, _, _, err := sp3.ObservableState(satellites[0], epochs[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.ObservableStates([]string{satellites[0]}, []float64{epochs[0]}); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.ObservableStatesShared([]string{satellites[0]}, epochs[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.PredictObservables(satellites[0], ECEF{}, epochs[0], nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sp3.PredictObservablesBatch([]PredictRequest{{SatelliteID: satellites[0], TRxJ2000S: epochs[0]}}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.PredictRanges([]PredictRequest{{SatelliteID: satellites[0], TRxJ2000S: epochs[0]}}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.GeometryVisible(ECEF{}, epochs[0], -90, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.GeometryPasses(ECEF{}, epochs[0], epochs[len(epochs)-1], 900, -90, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.GeometryVisibilitySeries(ECEF{}, epochs[0], epochs[len(epochs)-1], 900, -90, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := sp3.StencilExtent(); err != nil {
		t.Fatal(err)
	}

	options, err := NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.MinAgree = 1
	duplicate, duplicateReport, _ := MergeSP3([]*SP3{sp3, sp3}, &options)
	if duplicate != nil {
		_ = duplicate.Close()
	}
	if duplicateReport != nil {
		_ = duplicateReport.Close()
	}
	merged, report, err := MergeSP3([]*SP3{sp3}, &options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := merged.Close(); err != nil {
			t.Errorf("close merged SP3: %v", err)
		}
	})
	t.Cleanup(func() {
		if err := report.Close(); err != nil {
			t.Errorf("close merge report: %v", err)
		}
	})
	if _, err := report.AgreementSummary(); err != nil {
		t.Fatal(err)
	}
	if _, err := report.EpochAgreementCount(); err != nil {
		t.Fatal(err)
	}
	_, _ = report.EpochAgreement(0)
	for _, kind := range []SP3MergeFlagKind{SP3MergeFlagQuarantined, SP3MergeFlagSingleSource, SP3MergeFlagPositionOutlier, SP3MergeFlagClockOutlier} {
		if _, err := report.FlagCount(kind); err != nil {
			t.Fatal(err)
		}
		_, _ = report.Flag(kind, 0)
		_, _ = report.FlagSources(kind, 0)
	}
	if _, err := report.FrameReconciliationCount(); err != nil {
		t.Fatal(err)
	}
	_, _ = report.FrameReconciliation(0)
	_, _ = report.SourceLabel(0)
	_, _ = report.TargetLabel(0)
	_, _ = report.Provenance(0)
	_, _ = report.AssertedLabel(0, 0)
	if _, err := report.ContinuityVerdictJSON(merged, epochs[0], epochs[len(epochs)-1]); err != nil {
		t.Fatal(err)
	}

	if _, err := BuildSP3MergeInputIdentity(nil, nil); err == nil {
		t.Fatal("empty SP3 merge identity unexpectedly succeeded")
	}
}

func TestSP3RemainingCloseRaceAndInvalidBoundaries(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	epochs, err := sp3.Epochs()
	if err != nil || len(epochs) == 0 {
		t.Fatal(err)
	}
	satellites, err := sp3.Satellites()
	if err != nil || len(satellites) == 0 {
		t.Fatal(err)
	}
	assertConcurrentClose(t, func() error {
		_, _, _, err := sp3.ObservableState(satellites[0], epochs[0])
		return err
	}, sp3.Close)
	if _, err := sp3.GeometryVisible(ECEF{}, epochs[0], -90, []GNSSSystem{GNSSSystem(99)}); err == nil {
		t.Fatal("invalid geometry system accepted")
	}
	if _, _, _, err := sp3.ObservableState(satellites[0]+"\x00", epochs[0]); err == nil {
		t.Fatal("embedded NUL satellite accepted")
	}
}
