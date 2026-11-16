//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"math"
	"testing"
)

func TestChiSquareInverseSurvival(t *testing.T) {
	// For two degrees of freedom the survival function is exp(-x/2).
	value, err := ChiSquareInverseSurvival(math.Exp(-2), 2)
	if err != nil || math.Abs(value-4) > 1e-12 {
		t.Fatalf("quantile = %g, %v", value, err)
	}
}
