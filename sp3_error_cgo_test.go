//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestSP3TypedMergeErrorOwnsLosslessPayload(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	resolved, err := ResolveProductIdentity(gfzUltraRequest())
	if err != nil {
		t.Fatal(err)
	}
	resolved.HasFormatVersion = true
	resolved.FormatVersion = "c"
	identity := SP3ArtifactIdentity{
		RequestedIdentity: resolved,
		ResolvedIdentity:  resolved,
		Distribution:      DistributionSourceInMemory,
		OfficialFilename:  resolved.OfficialFilename,
		ProductSHA256:     strings.Repeat("a", 64),
		ProductByteLength: uint64(len(data)),
		ArchiveSHA256:     strings.Repeat("b", 64),
		ArchiveByteLength: uint64(len(data)),
		Compression:       ArchiveCompressionNone,
	}
	options, err := NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.PositionToleranceM = -1
	if _, err := BuildSP3MergeInputIdentity([]SP3ArtifactIdentity{identity}, &options); err == nil {
		t.Fatal("negative merge tolerance unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.SP3 == nil {
			t.Fatalf("typed SP3 detail missing: %T %v", err, err)
		}
		if statusErr.SP3.Kind != SP3ErrorKindMergeTolerance || statusErr.SP3.Field != SP3ErrorFieldPositionTolerance || statusErr.SP3.Reason != SP3ErrorReasonNegative {
			t.Fatalf("typed SP3 summary = %+v", statusErr.SP3)
		}
		if !statusErr.SP3.HasValue || statusErr.SP3.Value != -1 {
			t.Fatalf("typed SP3 value = present:%t value:%v", statusErr.SP3.HasValue, statusErr.SP3.Value)
		}
		var payloadFields struct {
			Error struct {
				Value string `json:"value"`
			} `json:"error"`
		}
		if !json.Valid(statusErr.SP3.Payload) || json.Unmarshal(statusErr.SP3.Payload, &payloadFields) != nil || payloadFields.Error.Value != "-1" {
			t.Fatalf("lossless SP3 payload = %s", statusErr.SP3.Payload)
		}
		ownedPayload := append([]byte(nil), statusErr.SP3.Payload...)

		// The public payload remains owned after later calls overwrite native TLS.
		if _, err := sp3.Epochs(); err != nil {
			t.Fatal(err)
		}
		if _, err := NewSP3MergeOptions(); err != nil {
			t.Fatal(err)
		}
		if string(statusErr.SP3.Payload) != string(ownedPayload) {
			t.Fatalf("captured payload changed after later native calls: %s", statusErr.SP3.Payload)
		}
	}
}
