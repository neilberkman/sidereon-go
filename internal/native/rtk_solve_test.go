//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "testing"

func TestRtkFloatConfigOwnsOuterAndNestedStorage(t *testing.T) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	config, err := (RtkFloatConfig{
		Epochs:       []RtkEpoch{{References: []RtkSatMeasurement{{SatelliteID: "G01", SDAmbiguityID: "G01"}}}},
		AmbiguityIDs: []string{"G01"},
	}).toC(alloc)
	if err != nil {
		t.Fatalf("toC: %v", err)
	}
	if config == nil || config.epochs == nil || config.ambiguity_ids == nil {
		t.Fatalf("config did not contain C-owned outer/nested pointers: %#v", config)
	}
	if len(alloc.values) < 5 {
		t.Fatalf("expected outer config, arrays, and strings in C-owned allocations; got %d", len(alloc.values))
	}
}
