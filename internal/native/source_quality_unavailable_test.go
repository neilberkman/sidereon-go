//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "testing"

func TestOwnedDiagnosticsReturnUnavailableWithoutCGO(t *testing.T) {
	if _, err := new(SourcedSolution).BroadcastReasonDetail(); err == nil {
		t.Fatal("sourced detail getter did not report unsupported native build")
	}
	if _, err := new(CoverageGrid).LookAngleErrorPayload(0, 0); err == nil {
		t.Fatal("coverage diagnostic getter did not report unsupported native build")
	}
}
