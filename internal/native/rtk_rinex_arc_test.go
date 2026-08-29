//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "testing"

func TestRtkRinexArcOptionsOwnPointerStorage(t *testing.T) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	options, err := (RtkRinexArcOptions{SignalPairs: []RtkRinexSignalPair{{System: 0, CodeObservable: "C1C", PhaseObservable: "L1C"}}, HasMaxEpochs: true, MaxEpochs: 2, MinCommonSatellites: 4}).toC(alloc)
	if err != nil {
		t.Fatalf("toC: %v", err)
	}
	if options == nil || options.signal_pairs == nil || options.signal_pair_count != 1 {
		t.Fatalf("options did not contain C-owned signal pairs: %#v", options)
	}
	if len(alloc.values) < 2 {
		t.Fatalf("expected C-owned outer and array, got %d allocations", len(alloc.values))
	}
}
