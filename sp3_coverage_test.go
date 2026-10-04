package sidereon

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestSP3CoverageRetainsDetachedGridAndChannelData(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	product, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	coverage, err := product.Coverage()
	if err != nil {
		t.Fatal(err)
	}
	epochCount, err := product.EpochCount()
	if err != nil {
		t.Fatal(err)
	}
	satellites, err := product.Satellites()
	if err != nil {
		t.Fatal(err)
	}
	if len(coverage.Satellites) != len(satellites) || len(coverage.Satellites) == 0 {
		t.Fatalf("coverage satellites=%d, product satellites=%d", len(coverage.Satellites), len(satellites))
	}
	expectedSatellites := []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}
	if !reflect.DeepEqual(satellites, expectedSatellites) || len(coverage.Grid.OutOfOrder) != 0 || len(coverage.Grid.Unplaced) != 0 || !coverage.Grid.HasInterval || coverage.Grid.IntervalS != 900 || !coverage.Grid.AgreesWithHeader {
		t.Fatalf("fixture grid/satellites = %+v / %v", coverage.Grid, satellites)
	}
	productSatellites := make(map[string]bool, len(satellites))
	for _, satellite := range satellites {
		productSatellites[satellite] = true
	}
	seen := make(map[string]bool, len(coverage.Satellites))
	for _, sat := range coverage.Satellites {
		if !productSatellites[sat.Satellite] || seen[sat.Satellite] {
			t.Fatalf("coverage returned unknown or duplicate satellite %q", sat.Satellite)
		}
		seen[sat.Satellite] = true
		if !sat.Declared || sat.Positions.Epochs != 13 || sat.Positions.SpanCount != 1 || sat.Positions.GapCount != 0 || !sat.Positions.Complete || sat.Clocks.Epochs != 13 || sat.Clocks.SpanCount != 1 || sat.Clocks.GapCount != 0 || !sat.Clocks.Complete {
			t.Fatalf("unexpected fixture channel coverage for %s: %+v", sat.Satellite, sat)
		}
		for label, channel := range map[string]SP3ChannelCoverage{"position": sat.Positions, "clock": sat.Clocks} {
			if channel.SpanCount != len(channel.Spans) || channel.GapCount != len(channel.Gaps) {
				t.Fatalf("%s counts disagree for %s: %+v", label, sat.Satellite, channel)
			}
			for _, span := range channel.Spans {
				if span.FirstIndex < 0 || span.LastIndex < span.FirstIndex || span.LastIndex >= epochCount {
					t.Fatalf("invalid %s span for %s: %+v", label, sat.Satellite, span)
				}
			}
		}
	}
	if coverage.Grid.HasInterval && coverage.Grid.IntervalS <= 0 {
		t.Fatalf("invalid grid interval: %+v", coverage.Grid)
	}
	epochs, err := product.Epochs()
	if err != nil {
		t.Fatal(err)
	}
	selected, err := product.SelectedNodes(satellites[0], epochs[0], epochs[len(epochs)-1])
	if err != nil {
		t.Fatal(err)
	}
	if unknown, err := product.SelectedNodes("G99", epochs[0], epochs[len(epochs)-1]); err != nil || len(unknown) != 0 {
		t.Fatalf("unknown-satellite selected nodes=%v err=%v", unknown, err)
	}
	if outside, err := product.SelectedNodes(satellites[0], epochs[len(epochs)-1]+86400, epochs[len(epochs)-1]+90000); err != nil || len(outside) != 0 {
		t.Fatalf("out-of-product selected nodes=%v err=%v", outside, err)
	}
	if reversed, err := product.SelectedNodes(satellites[0], epochs[1], epochs[0]); err == nil || len(reversed) != 0 {
		t.Fatalf("reversed-window selected nodes=%v err=%v", reversed, err)
	}
	if len(selected) == 0 {
		t.Fatal("selected-node query over the product returned no nodes")
	}
	for _, node := range selected {
		found := false
		for _, epoch := range epochs {
			if node == epoch {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("selected node %v is not a product epoch", node)
		}
	}
	snapshot, err := json.Marshal(coverage)
	if err != nil {
		t.Fatal(err)
	}
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	bad, parseErr := LoadSP3([]byte("invalid SP3 text"))
	if parseErr == nil {
		if bad != nil {
			closeAfterTest(t, bad)
		}
		t.Fatal("invalid SP3 unexpectedly parsed")
	}
	after, err := json.Marshal(coverage)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot, after) {
		t.Fatal("detached coverage changed after source close and subsequent parser failure")
	}
	if got, err := product.Coverage(); !errors.Is(err, ErrClosed) || len(got.Satellites) != 0 {
		t.Fatalf("closed coverage=%+v err=%v", got, err)
	}
	if len(coverage.Satellites) == 0 || coverage.Satellites[0].Satellite != satellites[0] {
		t.Fatalf("coverage was not detached from closed product: %+v", coverage)
	}
}
