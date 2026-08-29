package sidereon

import (
	"errors"
	"testing"
)

func TestIONEXSelectionRoutes(t *testing.T) {
	product, err := NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale:       UTC,
		MapEpochsJ2000S: []float64{0},
		LatNodesDeg:     []float64{1, -1},
		LonNodesDeg:     []float64{0, 1},
		DLatDeg:         -2,
		DLonDeg:         1,
		ShellHeightKm:   450,
		BaseRadiusKm:    6371,
		TECMAPsTECU:     []float64{1, 2, 3, 4},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := product.Close(); err != nil {
			t.Errorf("product.Close: %v", err)
		}
	})
	selected, metadata, err := SelectIONEX([]*IONEX{product, product}, 0, StalenessPolicyDefault())
	if err != nil || selected == nil || metadata.Kind != DegradationExact || metadata.RequestedEpochJ2000S != 0 || metadata.SourceEpochJ2000S != 0 || metadata.StalenessS != 0 || metadata.StalenessDays != 0 {
		t.Fatalf("SelectIONEX selected=%v metadata=%+v err=%v", selected, metadata, err)
	}
	t.Cleanup(func() {
		if err := selected.Close(); err != nil {
			t.Errorf("selected.Close: %v", err)
		}
	})
	selectedRange, rangeMetadata, err := SelectIONEXOverRange([]*IONEX{product, product}, 0, 0, StalenessPolicyDefault())
	if err != nil || selectedRange == nil || rangeMetadata.Kind != DegradationExact {
		t.Fatalf("SelectIONEXOverRange selected=%v metadata=%+v err=%v", selectedRange, rangeMetadata, err)
	}
	t.Cleanup(func() {
		if err := selectedRange.Close(); err != nil {
			t.Errorf("selectedRange.Close: %v", err)
		}
	})
	if err := selected.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := selected.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("selected EpochCount after Close=%v", err)
	}
}

func TestIONEXSelectionInvalidInputs(t *testing.T) {
	if _, _, err := SelectIONEX(nil, 0, StalenessPolicyDefault()); err == nil {
		t.Fatal("empty IONEX selection accepted")
	} else {
		var selectionErr *SelectionError
		if !errors.As(err, &selectionErr) || selectionErr.Status != SelectionEmptyProductSet {
			t.Fatalf("empty IONEX selection error=%T %#v", err, err)
		}
	}
	if _, _, err := SelectIONEX([]*IONEX{nil}, 0, StalenessPolicyDefault()); err == nil {
		t.Fatal("nil IONEX selection accepted")
	}
}
