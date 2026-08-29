//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "testing"

func TestSourceCovarianceAbsentSkipsShapeDecode(t *testing.T) {
	called := false
	value, present, err := sourceCovarianceResult(false, func() (NativeSourceCovariance, error) {
		called = true
		return NativeSourceCovariance{}, invalidArgument("decode should not run")
	})
	if err != nil {
		t.Fatalf("absent covariance returned error: %v", err)
	}
	if present {
		t.Fatal("absent covariance reported present")
	}
	if value != (NativeSourceCovariance{}) {
		t.Fatalf("absent covariance returned non-zero value: %#v", value)
	}
	if called {
		t.Fatal("absent covariance decoded native output")
	}
}
