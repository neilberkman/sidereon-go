package sidereon

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestPreciseInterpolantArtifactRoundTripAndCloseRace(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Errorf("close SP3: %v", err)
		}
	})
	interpolant, err := BuildPreciseEphemerisInterpolantFromSP3(sp3)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := interpolant.Close(); err != nil {
			t.Errorf("close precise interpolant: %v", err)
		}
	})
	satellites, err := sp3.Satellites()
	if err != nil || len(satellites) == 0 {
		t.Fatalf("satellites=%v err=%v", satellites, err)
	}
	epochs, err := sp3.Epochs()
	if err != nil || len(epochs) == 0 {
		t.Fatalf("epochs=%v err=%v", epochs, err)
	}
	states, err := interpolant.ObservableStatesShared([]string{satellites[0]}, epochs[0])
	if err != nil || len(states) != 1 {
		t.Fatalf("states=%v err=%v", states, err)
	}
	sampleValues, err := sp3.PreciseSamples()
	if err != nil || len(sampleValues) == 0 {
		t.Fatalf("precise samples=%d err=%v", len(sampleValues), err)
	}
	samples, err := BuildPreciseEphemerisSamples(sampleValues)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := samples.Close(); err != nil {
			t.Errorf("close precise samples: %v", err)
		}
	})
	predictions, err := samples.PredictRanges([]PredictRequest{{SatelliteID: satellites[0], ReceiverECEF: ECEF{}, TRxJ2000S: epochs[0]}}, nil)
	if err != nil || len(predictions) != 1 {
		t.Fatalf("range predictions=%v err=%v", predictions, err)
	}
	artifactBytes, artifactError, err := sp3.ArtifactBytes()
	if err != nil {
		t.Fatalf("artifact error kind=%d: %v", artifactError, err)
	}
	artifact, _, err := OpenPreciseInterpolantArtifact(artifactBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := artifact.Close(); err != nil {
			t.Errorf("close precise artifact: %v", err)
		}
	})
	if _, err := artifact.Satellites(); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = artifact.State(satellites[0], epochs[0]) }()
	}
	if err := artifact.Close(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}

func TestPreciseSamplesRejectInvalidShapes(t *testing.T) {
	if _, err := BuildPreciseEphemerisSamples(nil); err == nil {
		t.Fatal("empty precise samples accepted")
	}
	if _, err := BuildPreciseEphemerisInterpolant([]PreciseEphemerisSample{{Satellite: "G01", TimeScale: TimeScale(999)}}); err == nil {
		t.Fatal("invalid time scale accepted")
	}
}

func TestSP3PairOperationsInverseOrderDoNotDeadlock(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	first, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := first.Close(); err != nil {
			t.Errorf("close first SP3: %v", err)
		}
	})
	second, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := second.Close(); err != nil {
			t.Errorf("close second SP3: %v", err)
		}
	})

	done := make(chan struct{}, 2)
	go func() {
		_, _ = first.ClockReferenceOffsets(second, 1)
		done <- struct{}{}
	}()
	go func() {
		_, _ = second.ClockReferenceOffsets(first, 1)
		done <- struct{}{}
	}()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-deadline.C:
			t.Fatal("inverse-order SP3 pair operation did not complete")
		}
	}
}
