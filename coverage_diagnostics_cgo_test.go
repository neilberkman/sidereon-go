//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"math"
	"testing"
	"time"
)

func TestCheckedCoverage(t *testing.T) {
	tle, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tle.Close() }()
	grid, err := CoverageLookAngles([]*TLE{tle}, []PassStation{{LatitudeDeg: 51.5, LongitudeDeg: -0.1, AltitudeM: 80}}, time.Unix(1530000000, 0))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = grid.Close() }()
	if values, err := grid.TryAccessCounts(-90); err != nil || len(values) != 1 || values[0] != 1 {
		t.Fatalf("checked coverage counts = %v, %v", values, err)
	}
	if values, err := grid.TryVisibleMask(-90); err != nil || len(values) != 1 || !values[0] {
		t.Fatalf("checked coverage mask = %v, %v", values, err)
	}
	if _, err := grid.TryAccessCounts(math.NaN()); err == nil {
		t.Fatal("non-finite coverage threshold accepted")
	}
	if _, err := grid.TryVisibleMask(math.NaN()); err == nil {
		t.Fatal("non-finite coverage threshold accepted")
	}
}
