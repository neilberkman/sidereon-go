//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "testing"

func TestRtkFixedConfigOwnsPointerBearingStorage(t *testing.T) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	config, err := (RtkFixedConfig{
		Epochs:              []RtkEpoch{{References: []RtkSatMeasurement{{SatelliteID: "G01", SDAmbiguityID: "G01"}}}},
		AmbiguityIDs:        []string{"G01"},
		AmbiguitySatellites: []RtkAmbiguitySatellite{{ID: "G01", SatelliteID: "G01"}},
		Wavelengths:         []RtkFloatMapEntry{{ID: "G01", Value: 0.19}},
		Offsets:             []RtkFloatMapEntry{{ID: "G01", Value: 0.01}},
		ReceiverAntenna: &RtkReceiverAntennaCorrections{
			Base: RtkReceiverAntennaCalibration{
				NoAzimuthPCV: []RtkReceiverAntennaNoAzimuthPCV{{ZenithDeg: 10, ValueM: 0.001}},
				AzimuthPCV:   []RtkReceiverAntennaAzimuthPCV{{AzimuthDeg: 20, ZenithDeg: 10, ValueM: 0.002}},
			},
		},
	}).toC(alloc)
	if err != nil {
		t.Fatalf("toC: %v", err)
	}
	if config == nil || config.epochs == nil || config.ambiguity_ids == nil || config.ambiguity_satellites == nil || config.wavelengths_m == nil || config.offsets_m == nil || config.receiver_antenna == nil {
		t.Fatalf("config did not contain C-owned pointer-bearing storage: %#v", config)
	}
	if config.receiver_antenna.base.noazi_pcv_m == nil || config.receiver_antenna.base.azimuth_pcv_m == nil {
		t.Fatal("receiver antenna PCV arrays were not C-owned")
	}
	if len(alloc.values) < 12 {
		t.Fatalf("expected C-owned outer, arrays, and strings; got %d allocations", len(alloc.values))
	}
}
