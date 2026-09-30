//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestSP3EncodeTypedOutcomeAndOwnedRefusal(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	product, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, product)
	encoded, outcome, err := product.Encode()
	if err != nil || !outcome.IsOK || len(encoded) == 0 {
		t.Fatalf("Encode() = %d bytes, %+v, %v", len(encoded), outcome, err)
	}
	if !strings.HasPrefix(string(encoded), "#cP2020") {
		t.Fatalf("encoded SP3 header = %q", encoded[:min(len(encoded), 40)])
	}

	refusalData := bytes.Replace(data, []byte("%f  0.0000000"), []byte("%f 1.25000001"), 1)
	if bytes.Equal(refusalData, data) {
		t.Fatal("fixture did not contain the expected base field")
	}
	refused, err := LoadSP3(refusalData)
	if err != nil {
		t.Fatalf("parse finer base fixture: %v", err)
	}
	closeAfterTest(t, refused)
	text, refusedOutcome, err := refused.Encode()
	e := refusedOutcome.Error
	if err != nil || text != nil || refusedOutcome.IsOK || e.Kind != SP3WriteErrorPrecisionNotRepresentable || !e.HasField || e.Field != "pos/vel base" || !e.HasColumns || e.Columns != 10 || !e.HasDecimals || e.Decimals != 7 || !e.HasNumber || e.Number != 1.25000001 || e.Message == "" {
		t.Fatalf("typed writer refusal = %q, %+v, %v", text, refusedOutcome, err)
	}
	owned := e
	if err := refused.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSP3([]byte("invalid")); err == nil {
		t.Fatal("invalid SP3 parse unexpectedly succeeded")
	}
	if owned.Kind != SP3WriteErrorPrecisionNotRepresentable || owned.Field != "pos/vel base" || owned.Number != 1.25000001 || owned.Message == "" {
		t.Fatalf("writer outcome changed after source close and intervening failed parse: %+v", owned)
	}
}
