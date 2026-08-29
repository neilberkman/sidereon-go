//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

func copyFixedCString(dst unsafe.Pointer, width int, value string) error {
	if err := rejectEmbeddedNUL(value, "catalog text"); err != nil {
		return err
	}
	if len(value) >= width {
		return errors.New("sidereon: catalog text is too long")
	}
	if width != 0 {
		bytes := unsafe.Slice((*C.char)(dst), width)
		for i := range bytes {
			bytes[i] = 0
		}
		for i := range value {
			bytes[i] = C.char(value[i])
		}
	}
	return nil
}

func cProductIdentity(value ProductIdentity) (C.SidereonProductIdentity, error) {
	var out C.SidereonProductIdentity
	out.family, out.publisher, out.solution_class, out.campaign = C.uint32_t(value.Family), C.uint32_t(value.Publisher), C.uint32_t(value.SolutionClass), C.uint32_t(value.Campaign)
	out.filename_version, out.year, out.month, out.day = C.uint8_t(value.FilenameVersion), C.int32_t(value.Year), C.uint8_t(value.Month), C.uint8_t(value.Day)
	out.has_issue, out.has_format_version, out.has_prediction_horizon_days = C.uint8_t(0), C.uint8_t(0), C.uint8_t(0)
	if value.HasIssue {
		out.has_issue = 1
	}
	if value.HasFormatVersion {
		out.has_format_version = 1
	}
	if value.HasPredictionHorizonDays {
		out.has_prediction_horizon_days = 1
	}
	out.prediction_horizon_days, out.format = C.uint8_t(value.PredictionHorizonDays), C.uint32_t(value.Format)
	if err := copyFixedCString(unsafe.Pointer(&out.analysis_center[0]), len(out.analysis_center), value.AnalysisCenter); err != nil {
		return out, err
	}
	if err := copyFixedCString(unsafe.Pointer(&out.issue[0]), len(out.issue), value.Issue); err != nil {
		return out, err
	}
	if err := copyFixedCString(unsafe.Pointer(&out.span[0]), len(out.span), value.Span); err != nil {
		return out, err
	}
	if err := copyFixedCString(unsafe.Pointer(&out.sample[0]), len(out.sample), value.Sample); err != nil {
		return out, err
	}
	if err := copyFixedCString(unsafe.Pointer(&out.official_filename[0]), len(out.official_filename), value.OfficialFilename); err != nil {
		return out, err
	}
	if err := copyFixedCString(unsafe.Pointer(&out.format_version[0]), len(out.format_version), value.FormatVersion); err != nil {
		return out, err
	}
	return out, nil
}

func CatalogIdentityCacheKey(value ProductIdentity) ([]byte, error) {
	identity, err := cProductIdentity(value)
	if err != nil {
		return nil, err
	}
	var result []byte
	withCThread(func() {
		result, err = copyNativeBytesLocked("catalog identity cache key", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_data_product_identity_cache_key(&identity, out, n, w, r)
		})
	})
	runtime.KeepAlive(value)
	return result, err
}

func CatalogSP3ContentStartConvention(center string, year int, month, day uint8, issue string) (uint32, int64, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return 0, 0, err
	}
	var convention C.enum_SidereonSp3ContentStartConvention
	var offset C.int64_t
	var callErr error
	withCThread(func() {
		var centerPointer, issuePointer *C.char
		centerPointer, err = cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(centerPointer)
		if issue != "" {
			issuePointer, err = cString(issue)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(issuePointer)
		}
		callErr = callStatus(func() uint32 {
			return uint32(C.sidereon_data_sp3_content_start_convention(centerPointer, cYear, C.uint8_t(month), C.uint8_t(day), issuePointer, &convention, &offset))
		})
	})
	return uint32(convention), int64(offset), callErr
}

func CatalogValidateExactProductSet(expected, available []ProductIdentity) error {
	if len(expected) == 0 {
		return invalidArgument("expected product set must not be empty")
	}
	expectedValues := make([]C.SidereonProductIdentity, len(expected))
	availableValues := make([]C.SidereonProductIdentity, len(available))
	for i := range expected {
		var err error
		expectedValues[i], err = cProductIdentity(expected[i])
		if err != nil {
			return err
		}
	}
	for i := range available {
		var err error
		availableValues[i], err = cProductIdentity(available[i])
		if err != nil {
			return err
		}
	}
	expectedCount, err := checkedNativeSize(len(expectedValues))
	if err != nil {
		return err
	}
	availableCount, err := checkedNativeSize(len(availableValues))
	if err != nil {
		return err
	}
	withCThread(func() {
		var expectedPointer, availablePointer *C.SidereonProductIdentity
		if len(expectedValues) != 0 {
			expectedPointer = &expectedValues[0]
		}
		if len(availableValues) != 0 {
			availablePointer = &availableValues[0]
		}
		err = callStatus(func() uint32 {
			return uint32(C.sidereon_data_validate_exact_product_set(expectedPointer, expectedCount, availablePointer, availableCount))
		})
	})
	runtime.KeepAlive(expected)
	runtime.KeepAlive(available)
	return err
}
