//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"runtime"
	"testing"
)

func TestSP3DiagnosticRetainsAcrossGetterAndClearsOnProducer(t *testing.T) {
	data, err := os.ReadFile("../../testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// The legacy C entrypoint documents finite negative residual tolerance as
	// disabling the residual check; preserve that established contract.
	negativeResult, err := sp3.Continuity(0, -1)
	if err != nil {
		t.Fatalf("finite negative continuity tolerance should disable residual checking: %v", err)
	}
	if negativeResult.ResidualsChecked != 0 {
		t.Fatalf("negative tolerance checked %d residuals; want disabled residual check", negativeResult.ResidualsChecked)
	}
	if _, err := sp3.Continuity(0, math.NaN()); err == nil {
		t.Fatal("NaN continuity tolerance unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.SP3 == nil {
			t.Fatalf("typed SP3 continuity error missing: %T %v", err, err)
		}
	}

	before := currentSP3Diagnostic(t)
	if before == nil || before.Kind == 0 || before.PayloadJSON == "" {
		t.Fatalf("SP3 diagnostic missing after producer: %+v", before)
	}
	if before.Kind != 4 || before.Field != 9 || before.Reason != 2 || !before.HasValue || math.Float64bits(before.Value) != math.Float64bits(math.NaN()) {
		t.Fatalf("unexpected nonfinite continuity diagnostic: %+v", before)
	}
	var payload struct {
		Error struct {
			Value string `json:"value"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(before.PayloadJSON), &payload); err != nil || payload.Error.Value != "NaN" {
		t.Fatalf("continuity payload value = %q, %v; want exact NaN text", payload.Error.Value, err)
	}
	if _, err := sp3.EpochCount(); err != nil {
		t.Fatal(err)
	}
	afterGetter := currentSP3Diagnostic(t)
	if !sameSP3Diagnostic(afterGetter, before) {
		t.Fatalf("read-only getter changed retained SP3 diagnostic: before=%+v after=%+v", before, afterGetter)
	}

	if _, err := Sp3MergeOptionsInit(); err != nil {
		t.Fatal(err)
	}
	afterSuccess := currentSP3Diagnostic(t)
	if afterSuccess != nil {
		t.Fatalf("successful producer left stale SP3 diagnostic: %+v", afterSuccess)
	}
}

func TestSP3InterpolateReturnsCapturedNativeStatus(t *testing.T) {
	data, err := os.ReadFile("../../testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	_, _, _, err = sp3.Interpolate("G99", []float64{0})
	if err == nil {
		t.Fatal("interpolation for an absent satellite unexpectedly succeeded")
	}
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.Text == "" || statusErr.Detail == "" {
		t.Fatalf("native interpolation diagnostic was lost: %T %v", err, err)
	}
}

func currentSP3Diagnostic(t *testing.T) *SP3Error {
	t.Helper()
	var value *SP3Error
	var err error
	withCThread(func() { value, err = currentSp3ErrorLocked() })
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func sameSP3Diagnostic(left, right *SP3Error) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Kind == right.Kind && left.Field == right.Field && left.Reason == right.Reason &&
		left.HasValue == right.HasValue && math.Float64bits(left.Value) == math.Float64bits(right.Value) &&
		left.HasRequestedTick == right.HasRequestedTick && left.HasDeclaredTick == right.HasDeclaredTick &&
		left.HasRequestedJ2000S == right.HasRequestedJ2000S && left.HasDeclaredJ2000S == right.HasDeclaredJ2000S &&
		math.Float64bits(left.RequestedJ2000S) == math.Float64bits(right.RequestedJ2000S) &&
		math.Float64bits(left.DeclaredJ2000S) == math.Float64bits(right.DeclaredJ2000S) &&
		left.PayloadJSON == right.PayloadJSON
}
