//go:build !windows

package sidereon

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
)

func TestWriteDTEDStoreNewFileModeHonorsUmask(t *testing.T) {
	if os.Getenv("SIDEREON_DTED_UMASK_CHILD") == "1" {
		syscall.Umask(0)
		if err := WriteDTEDTreeToMMapStore("testdata/dted/tiles", os.Getenv("SIDEREON_DTED_UMASK_PATH")); err != nil {
			t.Fatal(err)
		}
		return
	}

	path := filepath.Join(t.TempDir(), "terrain.store")
	cmd := exec.Command(os.Args[0], "-test.run=^TestWriteDTEDStoreNewFileModeHonorsUmask$")
	cmd.Env = append(os.Environ(), "SIDEREON_DTED_UMASK_CHILD=1", "SIDEREON_DTED_UMASK_PATH="+path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("umask helper failed: %v\n%s", err, output)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o666); got != want {
		t.Fatalf("new DTED store mode = %#o, want %#o after umask 0000", got, want)
	}
}
