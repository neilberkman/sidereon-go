//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import (
	"strings"
	"testing"
)

func TestCheckedCatalogYearRejectsOverflowBeforeNativeCall(t *testing.T) {
	const (
		minCInt32 = -1 << 31
		maxCInt32 = 1<<31 - 1
	)
	if uint64(^uint(0)>>1) <= uint64(maxCInt32) {
		t.Skip("int cannot represent a year outside C.int32_t")
	}

	years := []int{int(int64(maxCInt32) + 1), int(int64(minCInt32) - 1)}
	tests := []struct {
		name string
		call func(int) error
	}{
		{name: "default sample", call: func(year int) error {
			_, err := CatalogDefaultSample("gfz_ult", ProductFamilySP3, year, 1, 1)
			return err
		}},
		{name: "supported samples", call: func(year int) error {
			_, err := CatalogSupportedSamples("gfz_ult", ProductFamilySP3, year, 1, 1, "0000")
			return err
		}},
		{name: "identity", call: func(year int) error {
			_, err := CatalogIdentity("gfz_ult", ProductFamilySP3, year, 1, 1, "", "0000")
			return err
		}},
		{name: "location", call: func(year int) error {
			_, err := CatalogLocation("gfz_ult", ProductFamilySP3, year, 1, 1, "", "0000", DistributionSourceDirect)
			return err
		}},
		{name: "listing URLs", call: func(year int) error {
			_, err := CatalogListingURLs("gfz_ult", ProductFamilySP3, year, 1, 1)
			return err
		}},
		{name: "next issue", call: func(year int) error {
			_, err := CatalogNextIssueDue("gfz_ult", ProductFamilySP3, year, 1, 1, 0, 0, 0)
			return err
		}},
		{name: "predicted IONEX", call: func(year int) error {
			_, err := CatalogPredictedIONEXCandidates(year, 1, 1, "01H")
			return err
		}},
	}
	for _, year := range years {
		year := year
		yearName := "above_maximum"
		if year < 0 {
			yearName = "below_minimum"
		}
		for _, test := range tests {
			test := test
			t.Run(test.name+"/"+yearName, func(t *testing.T) {
				err := test.call(year)
				if err == nil || !strings.Contains(err.Error(), "catalog year") {
					t.Fatalf("year %d returned error %v; want checked catalog-year error", year, err)
				}
			})
		}
	}
}

func TestCheckedCatalogYearAcceptsCInt32Bounds(t *testing.T) {
	for _, year := range []int{-1 << 31, 1<<31 - 1} {
		converted, err := checkedCatalogYear(year)
		if err != nil || int(converted) != year {
			t.Fatalf("checkedCatalogYear(%d) = %d, %v", year, converted, err)
		}
	}
}
