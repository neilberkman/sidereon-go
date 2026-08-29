package sidereon

import (
	"errors"
	"sync"
	"testing"
)

func dgnssObservations(values []SPPObservation) []DGNSSObservation {
	result := make([]DGNSSObservation, len(values))
	for i, value := range values {
		result[i] = DGNSSObservation(value)
	}
	return result
}

func TestDGNSSFixtureRoutes(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	config := usedSPPConfig()
	baseSolution, err := SolveSPP(sp3, config)
	if err != nil {
		t.Fatal(err)
	}
	base := dgnssObservations(config.Observations)
	corrections, err := DGNSSPseudorangeCorrections(sp3, baseSolution.PositionM, base, config.TRxJ2000S)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = corrections.Close() })
	count, err := corrections.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != len(base) {
		t.Fatalf("correction count = %d, want %d", count, len(base))
	}
	value, present, err := corrections.Correction("G08")
	if err != nil {
		t.Fatal(err)
	}
	if !present || value == 0 {
		t.Fatalf("G08 correction = (%v, %v), want a present non-zero correction", value, present)
	}
	if value, present, err := corrections.Correction("G01"); err != nil {
		t.Fatal(err)
	} else if present || value != 0 {
		t.Fatalf("G01 correction = (%v, %v), want absent zero", value, present)
	}
	if _, _, err := corrections.Correction("G99"); err == nil {
		t.Fatal("invalid satellite token was accepted")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Detail == "" {
			t.Fatalf("invalid-token error = %T %v, want detailed StatusError", err, err)
		}
	}

	applied, err := ApplyDGNSSCorrections(base, corrections)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = applied.Close() })
	corrected, dropped, err := applied.Counts()
	if err != nil {
		t.Fatal(err)
	}
	if corrected != len(base) || dropped != 0 {
		t.Fatalf("applied counts = (%d, %d), want (%d, 0)", corrected, dropped, len(base))
	}
	row, err := applied.Corrected(0)
	if err != nil {
		t.Fatal(err)
	}
	if row.SatelliteID == "" {
		t.Fatal("corrected row has an empty satellite ID")
	}
	if _, err := applied.Corrected(-1); err == nil {
		t.Fatal("negative corrected index was accepted")
	}
	if _, err := applied.Dropped(-1); err == nil {
		t.Fatal("negative dropped index was accepted")
	}

	solution, err := SolveDGNSSPosition(sp3, baseSolution.PositionM, base, base, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = solution.Close() })
	baseline, length, err := solution.Baseline()
	if err != nil {
		t.Fatal(err)
	}
	if length > 1e-3 || length < 0 || baseline[0] > 1e-3 || baseline[0] < -1e-3 || baseline[1] > 1e-3 || baseline[1] < -1e-3 || baseline[2] > 1e-3 || baseline[2] < -1e-3 {
		t.Fatalf("same-station baseline = (%v, %v), want sub-millimeter", baseline, length)
	}
	if dropped, err := solution.DroppedSatellites(); err != nil {
		t.Fatal(err)
	} else if len(dropped) != 0 {
		t.Fatalf("dropped satellites = %v, want none", dropped)
	}
	if _, err := solution.Solution(); err != nil {
		t.Fatal(err)
	}
}

func TestDGNSSZeroValuesAndBoundaries(t *testing.T) {
	var corrections DGNSSCorrections
	var applied DGNSSApplied
	var solution DGNSSSolution
	if err := corrections.Close(); err != nil {
		t.Fatal(err)
	}
	if err := applied.Close(); err != nil {
		t.Fatal(err)
	}
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := corrections.Count(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero corrections Count() error = %v, want ErrClosed", err)
	}
	if _, _, err := corrections.Correction("G08"); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero corrections Correction() error = %v, want ErrClosed", err)
	}
	if _, err := applied.Corrected(-1); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero applied Corrected() error = %v, want ErrClosed", err)
	}
	if _, _, err := applied.Counts(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero applied Counts() error = %v, want ErrClosed", err)
	}
	if _, err := applied.Dropped(0); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero applied Dropped() error = %v, want ErrClosed", err)
	}
	if _, _, err := solution.Baseline(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero solution Baseline() error = %v, want ErrClosed", err)
	}
	if _, err := solution.DroppedSatellites(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero solution DroppedSatellites() error = %v, want ErrClosed", err)
	}
	if _, err := solution.Solution(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero solution Solution() error = %v, want ErrClosed", err)
	}
	if _, err := DGNSSPseudorangeCorrections(nil, [3]float64{}, nil, 0); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil SP3 corrections error = %v, want ErrClosed", err)
	}
	if _, err := ApplyDGNSSCorrections(nil, nil); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil corrections apply error = %v, want ErrClosed", err)
	}
	if _, err := SolveDGNSSPosition(nil, [3]float64{}, nil, nil, SPPConfig{}); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil SP3 position solve error = %v, want ErrClosed", err)
	}
}

func TestDGNSSCloseReadRace(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	config := usedSPPConfig()
	baseSolution, err := SolveSPP(sp3, config)
	if err != nil {
		_ = sp3.Close()
		t.Fatal(err)
	}
	corrections, err := DGNSSPseudorangeCorrections(sp3, baseSolution.PositionM, dgnssObservations(config.Observations), config.TRxJ2000S)
	if err != nil {
		_ = sp3.Close()
		t.Fatal(err)
	}
	base := dgnssObservations(config.Observations)
	applied, err := ApplyDGNSSCorrections(base, corrections)
	if err != nil {
		_ = corrections.Close()
		_ = sp3.Close()
		t.Fatal(err)
	}
	solution, err := SolveDGNSSPosition(sp3, baseSolution.PositionM, base, base, config)
	_ = sp3.Close()
	if err != nil {
		_ = applied.Close()
		_ = corrections.Close()
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, _ = corrections.Count()
				_, _, _ = corrections.Correction("G08")
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, _, _ = applied.Counts()
				_, _ = applied.Corrected(0)
				_, _ = applied.Dropped(0)
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				_, _, _ = solution.Baseline()
				_, _ = solution.DroppedSatellites()
				_, _ = solution.Solution()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			_ = corrections.Close()
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			_ = applied.Close()
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			_ = solution.Close()
		}
	}()
	wg.Wait()
	if err := corrections.Close(); err != nil {
		t.Fatal(err)
	}
	if err := applied.Close(); err != nil {
		t.Fatal(err)
	}
	if err := solution.Close(); err != nil {
		t.Fatal(err)
	}
}
