package sidereon

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func syntheticIONEX(t *testing.T) *IONEX {
	t.Helper()
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
