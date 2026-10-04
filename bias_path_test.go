package sidereon

import (
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func gzipMembers(t *testing.T, members ...[]byte) []byte {
	t.Helper()
	var output bytes.Buffer
	for _, member := range members {
		writer := gzip.NewWriter(&output)
		if _, err := writer.Write(member); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return output.Bytes()
}

func TestReadBiasPathUsesSuffixAndBounds(t *testing.T) {
	root := t.TempDir()
	plainPath := filepath.Join(root, "bias.bia")
	if err := os.WriteFile(plainPath, []byte("plain"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readBiasPathWithLimits(plainPath, 4, 5)
	if err != nil || string(got) != "plain" {
		t.Fatalf("plain read=%q %v", got, err)
	}
	_, err = readBiasPathWithLimits(plainPath, 4, 4)
	var sizeLimit *SizeLimitError
	if !errors.As(err, &sizeLimit) || sizeLimit.Kind != "Bias-SINEX product" || sizeLimit.Limit != 4 {
		t.Fatalf("plain limit detail=%+v err=%v", sizeLimit, err)
	}

	payload := gzipMembers(t, []byte("multi "), []byte("member"))
	gzipPath := filepath.Join(root, "bias.bia.gz")
	if err := os.WriteFile(gzipPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err = readBiasPathWithLimits(gzipPath, int64(len(payload)), 12)
	if err != nil || string(got) != "multi member" {
		t.Fatalf("multi-member gzip=%q %v", got, err)
	}
	_, err = readBiasPathWithLimits(gzipPath, int64(len(payload)-1), 12)
	if !errors.As(err, &sizeLimit) || sizeLimit.Kind != "compressed Bias-SINEX product" || sizeLimit.Limit != int64(len(payload)-1) {
		t.Fatalf("compressed limit detail=%+v err=%v", sizeLimit, err)
	}
	_, err = readBiasPathWithLimits(gzipPath, int64(len(payload)), 11)
	if !errors.As(err, &sizeLimit) || sizeLimit.Kind != "decompressed Bias-SINEX product" || sizeLimit.Limit != 11 {
		t.Fatalf("expanded limit detail=%+v err=%v", sizeLimit, err)
	}

	magicOnlyPath := filepath.Join(root, "bias.not-gzip")
	if err := os.WriteFile(magicOnlyPath, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err = readBiasPathWithLimits(magicOnlyPath, int64(len(payload)), 100)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("non-.gz path was not kept as plain bytes: %x %v", got, err)
	}
}
