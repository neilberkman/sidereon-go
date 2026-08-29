//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Product family values mirror SidereonProductFamily in the C ABI.
type ProductFamily uint32

const (
	ProductFamilySP3             ProductFamily = C.SIDEREON_PRODUCT_FAMILY_SP3
	ProductFamilyIONEX           ProductFamily = C.SIDEREON_PRODUCT_FAMILY_IONEX
	ProductFamilyRINEXClock      ProductFamily = C.SIDEREON_PRODUCT_FAMILY_RINEX_CLOCK
	ProductFamilyRINEXNavigation ProductFamily = C.SIDEREON_PRODUCT_FAMILY_RINEX_NAVIGATION
)

// DistributionSource values mirror SidereonDistributionSource in the C ABI.
type DistributionSource uint32

const (
	DistributionSourceDirect    DistributionSource = C.SIDEREON_DISTRIBUTION_SOURCE_DIRECT
	DistributionSourceNASACDDIS DistributionSource = C.SIDEREON_DISTRIBUTION_SOURCE_NASA_CDDIS
	DistributionSourceLocalFile DistributionSource = C.SIDEREON_DISTRIBUTION_SOURCE_LOCAL_FILE
	DistributionSourceInMemory  DistributionSource = C.SIDEREON_DISTRIBUTION_SOURCE_IN_MEMORY
)

// ArchiveCompression values mirror SidereonArchiveCompression in the C ABI.
type ArchiveCompression uint32

const (
	ArchiveCompressionNone         ArchiveCompression = C.SIDEREON_ARCHIVE_COMPRESSION_NONE
	ArchiveCompressionGZIP         ArchiveCompression = C.SIDEREON_ARCHIVE_COMPRESSION_GZIP
	ArchiveCompressionUnixCompress ArchiveCompression = C.SIDEREON_ARCHIVE_COMPRESSION_UNIX_COMPRESS
)

// SolutionClass values mirror SidereonSolutionClass in the C ABI.
type SolutionClass uint32

const (
	SolutionClassFinal        SolutionClass = C.SIDEREON_SOLUTION_CLASS_FINAL
	SolutionClassRapid        SolutionClass = C.SIDEREON_SOLUTION_CLASS_RAPID
	SolutionClassUltraRapid   SolutionClass = C.SIDEREON_SOLUTION_CLASS_ULTRA_RAPID
	SolutionClassPredicted    SolutionClass = C.SIDEREON_SOLUTION_CLASS_PREDICTED
	SolutionClassBroadcast    SolutionClass = C.SIDEREON_SOLUTION_CLASS_BROADCAST
	SolutionClassNearRealTime SolutionClass = C.SIDEREON_SOLUTION_CLASS_NEAR_REAL_TIME
)

// ProductPublisher values mirror SidereonProductPublisher in the C ABI.
type ProductPublisher uint32

const (
	ProductPublisherIGS  ProductPublisher = C.SIDEREON_PRODUCT_PUBLISHER_IGS
	ProductPublisherCODE ProductPublisher = C.SIDEREON_PRODUCT_PUBLISHER_CODE
	ProductPublisherESA  ProductPublisher = C.SIDEREON_PRODUCT_PUBLISHER_ESA
	ProductPublisherGFZ  ProductPublisher = C.SIDEREON_PRODUCT_PUBLISHER_GFZ
	ProductPublisherWHU  ProductPublisher = C.SIDEREON_PRODUCT_PUBLISHER_WHU
)

// ProductCampaign values mirror SidereonProductCampaign in the C ABI.
type ProductCampaign uint32

const (
	ProductCampaignOperational         ProductCampaign = C.SIDEREON_PRODUCT_CAMPAIGN_OPERATIONAL
	ProductCampaignMultiGNSS           ProductCampaign = C.SIDEREON_PRODUCT_CAMPAIGN_MULTI_GNSS
	ProductCampaignMultiGNSSExperiment ProductCampaign = C.SIDEREON_PRODUCT_CAMPAIGN_MULTI_GNSS_EXPERIMENT
	ProductCampaignBroadcast           ProductCampaign = C.SIDEREON_PRODUCT_CAMPAIGN_BROADCAST
)

// ProductFormat values mirror SidereonProductFormat in the C ABI.
type ProductFormat uint32

const (
	ProductFormatSP3             ProductFormat = C.SIDEREON_PRODUCT_FORMAT_SP3
	ProductFormatIONEX           ProductFormat = C.SIDEREON_PRODUCT_FORMAT_IONEX
	ProductFormatRINEXClock      ProductFormat = C.SIDEREON_PRODUCT_FORMAT_RINEX_CLOCK
	ProductFormatRINEXNavigation ProductFormat = C.SIDEREON_PRODUCT_FORMAT_RINEX_NAVIGATION
)

// ProductIdentity is a copy of SidereonProductIdentity. Text is copied from
// C's fixed, null-terminated buffers before the native call returns.
type ProductIdentity struct {
	Family                   ProductFamily
	AnalysisCenter           string
	Publisher                ProductPublisher
	SolutionClass            SolutionClass
	Campaign                 ProductCampaign
	FilenameVersion          uint8
	Year                     int32
	Month                    uint8
	Day                      uint8
	HasIssue                 bool
	Issue                    string
	Span                     string
	Sample                   string
	OfficialFilename         string
	Format                   ProductFormat
	HasFormatVersion         bool
	FormatVersion            string
	HasPredictionHorizonDays bool
	PredictionHorizonDays    uint8
}

// DistributionLocation is a copy of SidereonDistributionLocation.
type DistributionLocation struct {
	Source          DistributionSource
	HasOriginalURL  bool
	OriginalURL     string
	ArchiveFilename string
	Compression     ArchiveCompression
}

func cString(value string) (*C.char, error) {
	if containsNUL(value) {
		return nil, errors.New("sidereon: string contains NUL")
	}
	result := C.CString(value)
	if result == nil {
		return nil, errors.New("sidereon: unable to allocate native string")
	}
	return result, nil
}

func containsNUL(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] == 0 {
			return true
		}
	}
	return false
}

func freeCString(value *C.char) {
	if value != nil {
		C.free(unsafe.Pointer(value))
	}
}

func fixedCString(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}

func identityFromC(value *C.SidereonProductIdentity) ProductIdentity {
	return ProductIdentity{
		Family:                   ProductFamily(value.family),
		AnalysisCenter:           fixedCString((*C.char)(unsafe.Pointer(&value.analysis_center[0]))),
		Publisher:                ProductPublisher(value.publisher),
		SolutionClass:            SolutionClass(value.solution_class),
		Campaign:                 ProductCampaign(value.campaign),
		FilenameVersion:          uint8(value.filename_version),
		Year:                     int32(value.year),
		Month:                    uint8(value.month),
		Day:                      uint8(value.day),
		HasIssue:                 value.has_issue != 0,
		Issue:                    fixedCString((*C.char)(unsafe.Pointer(&value.issue[0]))),
		Span:                     fixedCString((*C.char)(unsafe.Pointer(&value.span[0]))),
		Sample:                   fixedCString((*C.char)(unsafe.Pointer(&value.sample[0]))),
		OfficialFilename:         fixedCString((*C.char)(unsafe.Pointer(&value.official_filename[0]))),
		Format:                   ProductFormat(value.format),
		HasFormatVersion:         value.has_format_version != 0,
		FormatVersion:            fixedCString((*C.char)(unsafe.Pointer(&value.format_version[0]))),
		HasPredictionHorizonDays: value.has_prediction_horizon_days != 0,
		PredictionHorizonDays:    uint8(value.prediction_horizon_days),
	}
}

func locationFromC(value *C.SidereonDistributionLocation) DistributionLocation {
	return DistributionLocation{
		Source:          DistributionSource(value.source),
		HasOriginalURL:  bool(value.has_original_url),
		OriginalURL:     fixedCString((*C.char)(unsafe.Pointer(&value.original_url[0]))),
		ArchiveFilename: fixedCString((*C.char)(unsafe.Pointer(&value.archive_filename[0]))),
		Compression:     ArchiveCompression(value.compression),
	}
}

func callCatalogString(call func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]byte, error) {
	return copyNativeBytesLocked("catalog", nativeBytesCall(call))
}

func checkedCatalogYear(year int) (C.int32_t, error) {
	const (
		minCInt32 = -1 << 31
		maxCInt32 = 1<<31 - 1
	)
	if year < minCInt32 || year > maxCInt32 {
		return 0, fmt.Errorf("sidereon: catalog year %d does not fit in C.int32_t", year)
	}
	return C.int32_t(year), nil
}

func CatalogDefaultSample(center string, family ProductFamily, year int, month, day uint8) (string, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return "", err
	}
	var result []byte
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		result, callErr = callCatalogString(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_data_default_sample_for_date(cCenter, C.uint32_t(family), cYear, C.uint8_t(month), C.uint8_t(day), out, length, written, required)
		})
	})
	return string(result), callErr
}

func CatalogSupportedSamples(center string, family ProductFamily, year int, month, day uint8, issue string) ([]string, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return nil, err
	}
	var result []string
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		var cIssue *C.char
		if issue != "" {
			cIssue, err = cString(issue)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(cIssue)
		}
		var written, required C.size_t
		callErr = statusErrorLocked(uint32(C.sidereon_data_supported_samples(cCenter, C.uint32_t(family), cYear, C.uint8_t(month), C.uint8_t(day), cIssue, nil, 0, &written, &required)))
		if callErr != nil {
			return
		}
		if _, callErr = writtenToInt(written, 0, "catalog sample first-call written count"); callErr != nil {
			return
		}
		requiredInt, err := sizeTToInt(required, "catalog sample required count")
		if err != nil {
			callErr = err
			return
		}
		if _, callErr = checkedNativeAllocationSize(requiredInt, unsafe.Sizeof(C.SidereonProductSample{})); callErr != nil {
			return
		}
		samples := make([]C.SidereonProductSample, requiredInt)
		var output *C.SidereonProductSample
		if len(samples) != 0 {
			output = &samples[0]
		}
		callErr = statusErrorLocked(uint32(C.sidereon_data_supported_samples(cCenter, C.uint32_t(family), cYear, C.uint8_t(month), C.uint8_t(day), cIssue, output, C.size_t(len(samples)), &written, &required)))
		if callErr != nil {
			return
		}
		writtenInt, countErr := validateTwoPassCounts(
			"catalog samples", len(samples), requiredInt, uint64(written), uint64(required),
		)
		if countErr != nil {
			callErr = countErr
			return
		}
		result = make([]string, writtenInt)
		for i := range result {
			result[i] = fixedCString((*C.char)(unsafe.Pointer(&samples[i].token[0])))
		}
	})
	return result, callErr
}

func CatalogIdentity(center string, family ProductFamily, year int, month, day uint8, sample, issue string) (ProductIdentity, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return ProductIdentity{}, err
	}
	var result ProductIdentity
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		var cSample, cIssue *C.char
		if sample != "" {
			cSample, err = cString(sample)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(cSample)
		}
		if issue != "" {
			cIssue, err = cString(issue)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(cIssue)
		}
		var output C.SidereonProductIdentity
		callErr = callStatus(func() uint32 {
			return uint32(C.sidereon_data_product_identity(cCenter, C.uint32_t(family), cYear, C.uint8_t(month), C.uint8_t(day), cSample, cIssue, &output))
		})
		if callErr == nil {
			result = identityFromC(&output)
		}
	})
	return result, callErr
}

func CatalogLocation(center string, family ProductFamily, year int, month, day uint8, sample, issue string, source DistributionSource) (DistributionLocation, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return DistributionLocation{}, err
	}
	var result DistributionLocation
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		var cSample, cIssue *C.char
		if sample != "" {
			cSample, err = cString(sample)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(cSample)
		}
		if issue != "" {
			cIssue, err = cString(issue)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(cIssue)
		}
		var output C.SidereonDistributionLocation
		callErr = callStatus(func() uint32 {
			return uint32(C.sidereon_data_distribution_location(cCenter, C.uint32_t(family), cYear, C.uint8_t(month), C.uint8_t(day), cSample, cIssue, C.uint32_t(source), &output))
		})
		if callErr == nil {
			result = locationFromC(&output)
		}
	})
	return result, callErr
}

func CatalogSolutionClass(center string, family ProductFamily) (SolutionClass, error) {
	var result SolutionClass
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		var output C.enum_SidereonSolutionClass
		callErr = callStatus(func() uint32 {
			return uint32(C.sidereon_data_product_solution_class(cCenter, C.uint32_t(family), &output))
		})
		result = SolutionClass(output)
	})
	return result, callErr
}

func CatalogListingURLs(center string, family ProductFamily, year int, month, day uint8) ([]byte, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return nil, err
	}
	var result []byte
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		result, callErr = callCatalogString(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_data_publication_listing_urls_json(cCenter, C.uint32_t(family), cYear, C.uint8_t(month), C.uint8_t(day), out, length, written, required)
		})
	})
	return result, callErr
}

func CatalogNewestPublished(center string, family ProductFamily, body []byte) ([]byte, error) {
	var result []byte
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		if containsNUL(string(body)) {
			callErr = errors.New("sidereon: listing body contains NUL")
			return
		}
		cBody, err := cString(string(body))
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cBody)
		result, callErr = callCatalogString(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_data_newest_published_product_json(cCenter, C.uint32_t(family), cBody, out, length, written, required)
		})
	})
	return result, callErr
}

func CatalogNextIssueDue(center string, family ProductFamily, year int, month, day, hour, minute, second uint8) ([]byte, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return nil, err
	}
	var result []byte
	var callErr error
	withCThread(func() {
		cCenter, err := cString(center)
		if err != nil {
			callErr = err
			return
		}
		defer freeCString(cCenter)
		result, callErr = callCatalogString(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_data_next_issue_due_json(cCenter, C.uint32_t(family), cYear, C.uint8_t(month), C.uint8_t(day), C.uint8_t(hour), C.uint8_t(minute), C.uint8_t(second), out, length, written, required)
		})
	})
	return result, callErr
}

func CatalogPredictedIONEXCandidates(year int, month, day uint8, sample string) ([]byte, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return nil, err
	}
	var result []byte
	var callErr error
	withCThread(func() {
		var cSample *C.char
		var err error
		if sample != "" {
			cSample, err = cString(sample)
			if err != nil {
				callErr = err
				return
			}
			defer freeCString(cSample)
		}
		result, callErr = callCatalogString(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_data_predicted_ionex_line_candidates_json(cYear, C.uint8_t(month), C.uint8_t(day), cSample, out, length, written, required)
		})
	})
	return result, callErr
}

// ntripHandle serializes exclusive access to an owning C handle and permits
// shared read access where the native object is read-only. It owns no Go
// pointer in C; every release function runs on the locked C thread.
type ntripHandle struct {
	mu      sync.RWMutex
	ptr     unsafe.Pointer
	release func(unsafe.Pointer)
}

// noCopy makes go vet flag accidental copies of native owning values.
type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

func (h *ntripHandle) with(fn func(unsafe.Pointer) error) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ptr == nil {
		return ErrClosed
	}
	return fn(h.ptr)
}

func (h *ntripHandle) read(fn func(unsafe.Pointer) error) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.ptr == nil {
		return ErrClosed
	}
	return fn(h.ptr)
}

func (h *ntripHandle) close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.ptr == nil {
		return
	}
	ptr := h.ptr
	h.ptr = nil
	withCThread(func() { h.release(ptr) })
}

func addNTRIPCleanup[T any](owner *T, handle *ntripHandle) runtime.Cleanup {
	return runtime.AddCleanup(owner, func(value *ntripHandle) { value.close() }, handle)
}
