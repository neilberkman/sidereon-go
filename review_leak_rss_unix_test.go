//go:build darwin || linux

package sidereon

import (
	"runtime"
	"syscall"
)

func reviewMaxRSSBytes() (uint64, bool, error) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, true, err
	}
	multiplier := uint64(1)
	if runtime.GOOS == "linux" {
		// Darwin reports bytes and Linux reports kibibytes.
		multiplier = 1024
	}
	return uint64(usage.Maxrss) * multiplier, true, nil
}
