//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import (
	"errors"
	"testing"
)

func TestSatelliteMetadataUnavailableDiagnostics(t *testing.T) {
	if _, err := BuildSatelliteConstellation(nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("BuildSatelliteConstellation error = %v", err)
	}
	if _, err := new(ConstellationLookAngleArcs).ErrorPayload(0); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("look-angle ErrorPayload error = %v", err)
	}
	if _, err := new(ConstellationGroundTracks).ErrorPayload(0); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("ground-track ErrorPayload error = %v", err)
	}
	if _, err := new(ConstellationPasses).ErrorPayload(0); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("pass ErrorPayload error = %v", err)
	}
}
