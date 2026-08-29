//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import (
	"strings"
	"testing"
)

func TestNativeEnumValidation(t *testing.T) {
	if err := validateStaticPositionErrorKind(staticPositionErrorKindMax); err != nil {
		t.Fatalf("validateStaticPositionErrorKind(max): %v", err)
	}
	if err := validateStaticPositionErrorKind(staticPositionErrorKindMax + 1); err == nil {
		t.Fatal("validateStaticPositionErrorKind accepted an undefined value")
	}
	if err := validateSPPMetadataEnums(staticSPPSolveStatusMax, staticObservabilityTierMax); err != nil {
		t.Fatalf("validateSPPMetadataEnums(max, max): %v", err)
	}
	if err := validateSPPMetadataEnums(staticSPPSolveStatusMax+1, staticObservabilityTierMax); err == nil {
		t.Fatal("validateSPPMetadataEnums accepted an undefined solve status")
	}
	if err := validateSPPMetadataEnums(staticSPPSolveStatusMax, staticObservabilityTierMax+1); err == nil {
		t.Fatal("validateSPPMetadataEnums accepted an undefined observability tier")
	}
	if err := validateRtkStochastic(RtkStochasticRTKLIBValue); err != nil {
		t.Fatalf("validateRtkStochastic(max): %v", err)
	}
	if err := validateRtkStochastic(RtkStochasticRTKLIBValue + 1); err == nil {
		t.Fatal("validateRtkStochastic accepted an undefined value")
	}
}

func TestCheckedNativeAllocationSize(t *testing.T) {
	size, err := checkedNativeAllocationSize(3, 8)
	if err != nil || size != 24 {
		t.Fatalf("checkedNativeAllocationSize(3, 8) = %d, %v; want 24, nil", size, err)
	}

	maxInt := int(^uint(0) >> 1)
	if _, err := checkedNativeAllocationSize(maxInt/2+1, 2); err == nil {
		t.Fatal("checkedNativeAllocationSize accepted a product beyond Go's maximum int")
	}
	if _, err := checkedNativeAllocationSize(1, 0); err == nil {
		t.Fatal("checkedNativeAllocationSize accepted a zero element size")
	}
	if _, err := checkedNativeSize(-1); err == nil {
		t.Fatal("checkedNativeSize accepted a negative count")
	}
	if count, err := checkedNativeSize(3); err != nil || uint64(count) != 3 {
		t.Fatalf("checkedNativeSize(3) = %d, %v; want 3, nil", count, err)
	}
}

func TestCheckedNativeProduct(t *testing.T) {
	if product, err := checkedNativeProduct(4, 3, "test shape"); err != nil || product != 12 {
		t.Fatalf("checkedNativeProduct(4, 3) = %d, %v; want 12, nil", product, err)
	}
	if _, err := checkedNativeProduct(-1, 3, "test shape"); err == nil {
		t.Fatal("checkedNativeProduct accepted a negative count")
	}
	maxInt := int(^uint(0) >> 1)
	if _, err := checkedNativeProduct(maxInt/2+1, 2, "test shape"); err == nil {
		t.Fatal("checkedNativeProduct accepted an overflowing product")
	}
}

func TestCheckedNativeCount(t *testing.T) {
	if count, err := checkedNativeCount(7); err != nil || count != 7 {
		t.Fatalf("checkedNativeCount(7) = %d, %v; want 7, nil", count, err)
	}
	tooLarge := uint64(^uint(0)>>1) + 1
	if _, err := checkedNativeCount(tooLarge); err == nil {
		t.Fatal("checkedNativeCount accepted a value larger than int")
	}
}

func TestValidateTwoPassCounts(t *testing.T) {
	tooLarge := uint64(^uint(0)>>1) + 1
	tests := []struct {
		name     string
		capacity int
		expected int
		written  uint64
		required uint64
		want     int
		wantErr  string
	}{
		{
			name:     "complete copy",
			capacity: 3,
			expected: 3,
			written:  3,
			required: 3,
			want:     3,
		},
		{
			name:     "empty copy",
			capacity: 0,
			expected: 0,
			written:  0,
			required: 0,
			want:     0,
		},
		{
			name:     "written count overflow",
			capacity: 3,
			expected: 3,
			written:  tooLarge,
			required: 3,
			wantErr:  "written count",
		},
		{
			name:     "required count overflow",
			capacity: 3,
			expected: 3,
			written:  3,
			required: tooLarge,
			wantErr:  "required count",
		},
		{
			name:     "required count exceeds capacity",
			capacity: 3,
			expected: 3,
			written:  3,
			required: 4,
			wantErr:  "exceeds allocated capacity",
		},
		{
			name:     "written count exceeds capacity",
			capacity: 3,
			expected: 3,
			written:  4,
			required: 3,
			wantErr:  "wrote 4 entries into capacity 3",
		},
		{
			name:     "required count changed",
			capacity: 3,
			expected: 3,
			written:  2,
			required: 2,
			wantErr:  "required count changed",
		},
		{
			name:     "partial copy",
			capacity: 3,
			expected: 3,
			written:  2,
			required: 3,
			wantErr:  "wrote 2 entries, required 3",
		},
		{
			name:     "written count exceeds capacity",
			capacity: 3,
			expected: 3,
			written:  4,
			required: 3,
			wantErr:  "wrote 4 entries into capacity 3",
		},
		{
			name:     "written count exceeds required",
			capacity: 4,
			expected: 3,
			written:  4,
			required: 3,
			wantErr:  "wrote 4 entries, required 3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateTwoPassCounts("test list", test.capacity, test.expected, test.written, test.required)
			if test.wantErr == "" {
				if err != nil || got != test.want {
					t.Fatalf("validateTwoPassCounts() = %d, %v; want %d, nil", got, err, test.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("validateTwoPassCounts() error = %v; want substring %q", err, test.wantErr)
			}
		})
	}
}

func TestValidateNativeCopyShapes(t *testing.T) {
	if count, err := validateNativeQuery("test query", 0, 4); err != nil || count != 4 {
		t.Fatalf("validateNativeQuery(0, 4) = %d, %v; want 4, nil", count, err)
	}
	if _, err := validateNativeQuery("test query", 1, 4); err == nil || !strings.Contains(err.Error(), "query wrote") {
		t.Fatalf("validateNativeQuery accepted a non-empty written count: %v", err)
	}
	tooLarge := uint64(^uint(0)>>1) + 1
	if _, err := validateNativeQuery("test query", 0, tooLarge); err == nil || !strings.Contains(err.Error(), "required count") {
		t.Fatalf("validateNativeQuery accepted an overflowing required count: %v", err)
	}
	if _, err := validateNativeOutput("test copy", 2, 1, 2); err == nil {
		t.Fatal("validateNativeOutput accepted a count mismatch")
	}
}
