package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

func syntheticIONEX(t *testing.T) *IONEX {
	t.Helper()
	product, err := NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale:         UTC,
		MapEpochsJ2000S:   []float64{0},
		LatNodesDeg:       []float64{1, -1},
		LonNodesDeg:       []float64{0, 1},
		DLatDeg:           -2,
		DLonDeg:           1,
		ShellHeightKm:     450,
		BaseRadiusKm:      6371,
		TECMAPsTECU:       []float64{1, 2, 3, 4},
		HeightPresent:     true,
		HeightMapsKm:      []float64{450, 451, 452, 453},
		HeightMapsPresent: []bool{true, false, true, false},
	})
	if err != nil {
		t.Fatalf("build synthetic IONEX: %v", err)
	}
	if product == nil {
		t.Fatal("build synthetic IONEX returned nil")
	}
	return product
}

func TestMetadataUnixMicrosecondsBoundaries(t *testing.T) {
	const (
		maxSeconds = int64(9223372036854)
		maxMicros  = int64(775807)
	)
	if got, err := metadataUnixMicroseconds(time.Unix(maxSeconds, maxMicros*1000)); err != nil || got != int64(1<<63-1) {
		t.Fatalf("maximum Unix microseconds = %d, %v", got, err)
	}
	if _, err := metadataUnixMicroseconds(time.Unix(maxSeconds, (maxMicros+1)*1000)); err == nil {
		t.Fatal("positive fractional overflow was accepted")
	}
	const minBoundaryMicros = int64(224192)
	if got, err := metadataUnixMicroseconds(time.Unix(-9223372036855, minBoundaryMicros*1000)); err != nil || got != -1<<63 {
		t.Fatalf("minimum Unix microseconds = %d, %v", got, err)
	}
	if _, err := metadataUnixMicroseconds(time.Unix(-9223372036855, (minBoundaryMicros-1)*1000)); err == nil {
		t.Fatal("negative fractional overflow was accepted")
	}
}

func TestIONEXSyntheticGridAndDetachedOutputs(t *testing.T) {
	product := syntheticIONEX(t)
	defer func() { _ = product.Close() }()
	if count, err := product.EpochCount(); err != nil || count != 1 {
		t.Fatalf("epoch count = %d, %v", count, err)
	}
	if values, err := product.TECMAPsTECU(); err != nil || len(values) != 4 || values[0] != 1 || values[3] != 4 {
		t.Fatalf("TEC map = %#v, %v", values, err)
	}
	if samples, err := product.TECSamples(); err != nil || len(samples) != 4 || !samples[0].VTECPresenceKnown || !samples[0].VTECPresent {
		t.Fatalf("TEC sample presence = %#v, %v", samples, err)
	}
	if present, err := product.TECMapPresence(); err != nil || len(present) != 4 || !present[0] || !present[3] {
		t.Fatalf("TEC cell presence = %#v, %v", present, err)
	}
	if present, err := product.RMSMapPresence(); err != nil || len(present) != 0 {
		t.Fatalf("RMS cell presence = %#v, %v", present, err)
	}
	if present, err := product.HeightMapPresence(); err != nil || len(present) != 4 || !present[0] || present[1] || !present[2] || present[3] {
		t.Fatalf("map height presence = %#v, %v", present, err)
	}
	if values, err := product.HeightMapsKm(); err != nil || len(values) != 4 || values[0] != 450 || !math.IsNaN(values[1]) || values[2] != 452 || !math.IsNaN(values[3]) {
		t.Fatalf("height maps = %#v, %v", values, err)
	}
	if values, err := product.LatNodesDeg(); err != nil || len(values) != 2 || values[0] != 1 || values[1] != -1 {
		t.Fatalf("latitude axis = %#v, %v", values, err)
	}
	if err := product.Close(); err != nil {
		t.Fatalf("close IONEX: %v", err)
	}
	if err := product.Close(); err != nil {
		t.Fatalf("idempotent close IONEX: %v", err)
	}
	if _, err := product.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("EpochCount after close = %v, want ErrClosed", err)
	}
}

func TestIONEXInvalidGridShape(t *testing.T) {
	_, err := NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale:       UTC,
		MapEpochsJ2000S: []float64{0},
		LatNodesDeg:     []float64{1, -1},
		LonNodesDeg:     []float64{0, 1},
		TECMAPsTECU:     []float64{1},
	})
	if err == nil {
		t.Fatal("mismatched TEC grid shape was accepted")
	}
}

func TestIONEXDisabledHeightStackIsIgnored(t *testing.T) {
	product, err := NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale: UTC, MapEpochsJ2000S: []float64{0}, LatNodesDeg: []float64{1, -1}, LonNodesDeg: []float64{0, 1},
		DLatDeg: -2, DLonDeg: 1, ShellHeightKm: 450, BaseRadiusKm: 6371, TECMAPsTECU: []float64{1, 2, 3, 4},
		HeightMapsKm: []float64{123}, HeightMapsPresent: []bool{false, true},
	})
	if err != nil {
		t.Fatalf("disabled malformed height stack should be ignored: %v", err)
	}
	closeAfterTest(t, product)
	if present, err := product.HeightMapPresence(); err != nil || len(present) != 0 {
		t.Fatalf("disabled height stack presence = %#v, %v", present, err)
	}
}

func TestIONEXHeaderMetadataWritesAndReadsDetachedValues(t *testing.T) {
	header := &IONEXHeaderMetadata{
		Version: 1.0, Date: "2026-09-29 00:00 UTC", Program: "sidereon", RunBy: "tests", SatelliteSystem: "GPS",
		HasSatelliteCount: true, SatelliteCount: 3, HasStationCount: true, StationCount: 2,
		ElevationCutoffDeg: 10, IntervalS: 3600, HasMapsInFile: true, MapsInFile: 2,
		ObservablesUsed: "GPS", MappingDeclaration: IONEXMappingDeclarationDeclared,
		MappingFunction: IONEXMappingFunctionOther, MappingFunctionCode: "XMAP",
		Descriptions: []string{"header description"}, Comments: []string{"header comment"},
	}
	product, err := NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale: UTC, MapEpochsJ2000S: []float64{0, 3600}, LatNodesDeg: []float64{1, -1}, LonNodesDeg: []float64{0, 1},
		DLatDeg: -2, DLonDeg: 1, ShellHeightKm: 450, BaseRadiusKm: 6371,
		TECMAPsTECU: []float64{1, 2, 3, 4, 5, 6, 7, 8}, Header: header,
	})
	if err != nil {
		t.Fatalf("build product with header: %v", err)
	}
	closeAfterTest(t, product)
	text, err := product.ToIONEXText()
	if err != nil {
		t.Fatalf("write product: %v", err)
	}
	parsed, err := ParseIONEX(text)
	if err != nil {
		t.Fatalf("parse written product: %v", err)
	}
	closeAfterTest(t, parsed)
	got, err := parsed.HeaderMetadata()
	if err != nil {
		t.Fatalf("read header metadata: %v", err)
	}
	if got.Version != header.Version || got.Date != header.Date || got.Program != header.Program || got.RunBy != header.RunBy || got.SatelliteSystem != header.SatelliteSystem || !got.HasSatelliteCount || got.SatelliteCount != header.SatelliteCount || !got.HasStationCount || got.StationCount != header.StationCount || got.ElevationCutoffDeg != header.ElevationCutoffDeg || got.IntervalS != header.IntervalS || !got.HasMapsInFile || got.MapsInFile != header.MapsInFile || got.ObservablesUsed != header.ObservablesUsed || got.MappingDeclaration != header.MappingDeclaration || got.MappingFunction != header.MappingFunction || got.MappingFunctionCode != header.MappingFunctionCode || len(got.Descriptions) != 1 || got.Descriptions[0] != header.Descriptions[0] || len(got.Comments) != 1 || got.Comments[0] != header.Comments[0] {
		t.Fatalf("header metadata changed across write/read: got %#v want %#v", got, *header)
	}
	if err := parsed.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseIONEX([]byte("not an IONEX product")); err == nil {
		t.Fatal("malformed IONEX text unexpectedly parsed")
	}
	if got.Descriptions[0] != "header description" || got.Comments[0] != "header comment" || got.MappingDeclaration != IONEXMappingDeclarationDeclared || got.MappingFunction != IONEXMappingFunctionOther || got.MappingFunctionCode != "XMAP" {
		t.Fatalf("detached custom header snapshot changed after source close and parse failure: %#v", got)
	}
	invalidHeader := *header
	invalidHeader.MappingDeclaration = IONEXMappingDeclarationKind(99)
	_, err = NewIONEXFromTECGridSamples(TECGridSamples{
		TimeScale: UTC, MapEpochsJ2000S: []float64{0, 3600}, LatNodesDeg: []float64{1, -1}, LonNodesDeg: []float64{0, 1},
		DLatDeg: -2, DLonDeg: 1, ShellHeightKm: 450, BaseRadiusKm: 6371,
		TECMAPsTECU: []float64{1, 2, 3, 4, 5, 6, 7, 8}, Header: &invalidHeader,
	})
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument {
		t.Fatalf("invalid mapping declaration error = %v, want StatusInvalidArgument", err)
	}
}

func TestIONEXReadCloseRace(t *testing.T) {
	product := syntheticIONEX(t)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_, _ = product.EpochCount()
				_, _ = product.GridInfo()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			_ = product.Close()
		}
	}()
	wg.Wait()
}
