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
	"fmt"
	"runtime"
	"sort"
	"unsafe"
)

const (
	ExactSp3CoverageHalfOpen                  = uint32(C.SIDEREON_EXACT_SP3_COVERAGE_HALF_OPEN)
	ExactSp3CoverageInclusive                 = uint32(C.SIDEREON_EXACT_SP3_COVERAGE_INCLUSIVE)
	Sp3FrameReconciliationAssertedEquivalence = uint32(C.SIDEREON_SP3_FRAME_RECONCILIATION_METHOD_ASSERTED_EQUIVALENCE)
	Sp3FrameReconciliationHelmert             = uint32(C.SIDEREON_SP3_FRAME_RECONCILIATION_METHOD_HELMERT)
	Sp3MergeCombineMean                       = uint32(C.SIDEREON_SP3_MERGE_COMBINE_MEAN)
	Sp3MergeCombineMedian                     = uint32(C.SIDEREON_SP3_MERGE_COMBINE_MEDIAN)
	Sp3MergeCombinePrecedence                 = uint32(C.SIDEREON_SP3_MERGE_COMBINE_PRECEDENCE)
	Sp3MergePrecedenceCell                    = uint32(C.SIDEREON_SP3_MERGE_PRECEDENCE_SCOPE_CELL)
	Sp3MergePrecedenceSatelliteArc            = uint32(C.SIDEREON_SP3_MERGE_PRECEDENCE_SCOPE_SATELLITE_ARC)
	Sp3MergeFlagQuarantined                   = uint32(C.SIDEREON_SP3_MERGE_FLAG_KIND_QUARANTINED)
	Sp3MergeFlagSingleSource                  = uint32(C.SIDEREON_SP3_MERGE_FLAG_KIND_SINGLE_SOURCE)
	Sp3MergeFlagPositionOutlier               = uint32(C.SIDEREON_SP3_MERGE_FLAG_KIND_POSITION_OUTLIER)
	Sp3MergeFlagClockOutlier                  = uint32(C.SIDEREON_SP3_MERGE_FLAG_KIND_CLOCK_OUTLIER)
)

func validateExactSp3Coverage(value uint32) error {
	if value > ExactSp3CoverageInclusive {
		return invalidArgument("invalid exact SP3 coverage returned by native code")
	}
	return nil
}

func validateSp3FrameReconciliationMethod(value uint32) error {
	if value > Sp3FrameReconciliationHelmert {
		return invalidArgument("invalid SP3 frame reconciliation method returned by native code")
	}
	return nil
}

func validateSp3MergeCombine(value uint32) error {
	if value > Sp3MergeCombinePrecedence {
		return invalidArgument("invalid SP3 merge combine value")
	}
	return nil
}

func validateSp3MergePrecedenceScope(value uint32) error {
	if value > Sp3MergePrecedenceSatelliteArc {
		return invalidArgument("invalid SP3 merge precedence scope value")
	}
	return nil
}

func validateSp3MergeFlagKind(value uint32) error {
	if value > Sp3MergeFlagClockOutlier {
		return invalidArgument("invalid SP3 merge flag kind")
	}
	return nil
}

func validateSp3TerrestrialFrame(value uint32) error {
	switch value {
	case uint32(C.SIDEREON_TERRESTRIAL_FRAME_ITRF2020),
		uint32(C.SIDEREON_TERRESTRIAL_FRAME_ITRF2014),
		uint32(C.SIDEREON_TERRESTRIAL_FRAME_ITRF2008),
		uint32(C.SIDEREON_TERRESTRIAL_FRAME_ETRF2020):
		return nil
	default:
		return invalidArgument("invalid SP3 terrestrial frame returned by native code")
	}
}

type ExactSp3Request struct {
	_      noCopy
	handle *positioningHandle
}

type Sp3MergeInputIdentity struct {
	_      noCopy
	handle *positioningHandle
}

type Sp3MergeReport struct {
	_      noCopy
	handle *positioningHandle
}

func releaseExactSp3Request(p unsafe.Pointer) {
	withCThread(func() { C.sidereon_sp3_exact_request_free((*C.SidereonExactSp3Request)(p)) })
}

func releaseSp3MergeInputIdentity(p unsafe.Pointer) {
	withCThread(func() { C.sidereon_sp3_merge_input_identity_free((*C.SidereonSp3MergeInputIdentity)(p)) })
}

func releaseSp3MergeReport(p unsafe.Pointer) {
	withCThread(func() { C.sidereon_sp3_merge_report_free((*C.SidereonSp3MergeReport)(p)) })
}

func newExactSp3Request(p *C.SidereonExactSp3Request) (*ExactSp3Request, error) {
	if p == nil {
		return nil, missingNativeHandle("exact SP3 request")
	}
	return &ExactSp3Request{handle: newPositioningHandle(unsafe.Pointer(p), releaseExactSp3Request)}, nil
}

func newSp3MergeInputIdentity(p *C.SidereonSp3MergeInputIdentity) (*Sp3MergeInputIdentity, error) {
	if p == nil {
		return nil, missingNativeHandle("SP3 merge input identity")
	}
	return &Sp3MergeInputIdentity{handle: newPositioningHandle(unsafe.Pointer(p), releaseSp3MergeInputIdentity)}, nil
}

func newSp3MergeReport(p *C.SidereonSp3MergeReport) (*Sp3MergeReport, error) {
	if p == nil {
		return nil, missingNativeHandle("SP3 merge report")
	}
	return &Sp3MergeReport{handle: newPositioningHandle(unsafe.Pointer(p), releaseSp3MergeReport)}, nil
}

func (r *ExactSp3Request) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return r.handle.close()
}

func (r *Sp3MergeInputIdentity) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return r.handle.close()
}

func (r *Sp3MergeReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return r.handle.close()
}

// withPositioningHandleSet acquires unique handles in address order while
// preserving the caller's order in the callback. It is used for C calls that
// consume several SP3 products and therefore also handles repeated inputs.
func withPositioningHandleSet(handles []*positioningHandle, fn func([]unsafe.Pointer) error) error {
	if len(handles) == 0 {
		return fn(nil)
	}
	type indexed struct {
		h *positioningHandle
	}
	ordered := make([]indexed, 0, len(handles))
	seen := make(map[*positioningHandle]struct{}, len(handles))
	for _, h := range handles {
		if h == nil {
			return ErrClosed
		}
		if _, ok := seen[h]; ok {
			continue
		}
		seen[h] = struct{}{}
		ordered = append(ordered, indexed{h: h})
	}
	sort.Slice(ordered, func(i, j int) bool {
		return uintptr(unsafe.Pointer(ordered[i].h)) < uintptr(unsafe.Pointer(ordered[j].h))
	})
	for _, item := range ordered {
		item.h.mu.RLock()
	}
	defer func() {
		for i := len(ordered) - 1; i >= 0; i-- {
			ordered[i].h.mu.RUnlock()
		}
	}()
	for _, item := range ordered {
		if item.h.resource == nil {
			return ErrClosed
		}
	}
	pointers := make([]unsafe.Pointer, len(handles))
	for i, h := range handles {
		if err := h.resource.with(func(p unsafe.Pointer) error {
			pointers[i] = p
			return nil
		}); err != nil {
			return err
		}
	}
	return fn(pointers)
}

type sp3Arena struct {
	values []unsafe.Pointer
}

func (a *sp3Arena) malloc(count int, size uintptr) (unsafe.Pointer, error) {
	bytes, err := checkedNativeAllocationSize(count, size)
	if err != nil {
		return nil, err
	}
	if bytes == 0 {
		return nil, nil
	}
	p := C.calloc(1, C.size_t(bytes))
	if p == nil {
		return nil, errors.New("sidereon: unable to allocate native SP3 temporary")
	}
	a.values = append(a.values, p)
	return p, nil
}

func (a *sp3Arena) string(value, name string) (*C.char, error) {
	p, err := copyNativeCString(value, name)
	if err != nil {
		return nil, err
	}
	a.values = append(a.values, p)
	return (*C.char)(p), nil
}

func (a *sp3Arena) close() {
	for i := len(a.values) - 1; i >= 0; i-- {
		C.free(a.values[i])
	}
	a.values = nil
}

func (a *sp3Arena) cStringArray(values []string, name string) (**C.char, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	p, err := a.malloc(len(values), unsafe.Sizeof(uintptr(0)))
	if err != nil {
		return nil, 0, err
	}
	array := unsafe.Slice((**C.char)(p), len(values))
	for i, value := range values {
		array[i], err = a.string(value, name)
		if err != nil {
			return nil, 0, err
		}
	}
	return (**C.char)(p), count, nil
}

func cBool(value bool) C.uint8_t {
	if value {
		return 1
	}
	return 0
}

type NativeSp3MergeOptions struct {
	PositionToleranceM             float64
	ClockToleranceS                float64
	MinAgree                       int
	ClockMinCommon                 int
	Combine                        uint32
	PrecedenceScope                uint32
	OutlierRejectEnabled           bool
	OutlierRejectPositionTolerance float64
	OutlierRejectClockTolerance    float64
	TargetEpochIntervalEnabled     bool
	TargetEpochIntervalS           float64
	Systems                        []uint32
	AssertedFrameLabelSets         [][]string
	HelmertFrameReconciliation     bool
}

func cSp3MergeOptions(a *sp3Arena, value NativeSp3MergeOptions) (*C.SidereonSp3MergeOptions, error) {
	if err := validateSp3MergeCombine(value.Combine); err != nil {
		return nil, err
	}
	if err := validateSp3MergePrecedenceScope(value.PrecedenceScope); err != nil {
		return nil, err
	}
	if value.MinAgree < 0 || value.ClockMinCommon < 0 {
		return nil, invalidArgument("SP3 merge counts must not be negative")
	}
	minAgree, err := checkedNativeSize(value.MinAgree)
	if err != nil {
		return nil, err
	}
	clockMinCommon, err := checkedNativeSize(value.ClockMinCommon)
	if err != nil {
		return nil, err
	}
	for i, system := range value.Systems {
		if system > uint32(C.SIDEREON_GNSS_SYSTEM_SBAS) {
			return nil, invalidArgument(fmt.Sprintf("invalid SP3 merge system at index %d", i))
		}
	}
	for _, labelSet := range value.AssertedFrameLabelSets {
		if len(labelSet) < 2 {
			return nil, invalidArgument("SP3 asserted frame label sets require at least two labels")
		}
	}
	p, err := a.malloc(1, unsafe.Sizeof(C.SidereonSp3MergeOptions{}))
	if err != nil {
		return nil, err
	}
	out := (*C.SidereonSp3MergeOptions)(p)
	out.position_tolerance_m = C.double(value.PositionToleranceM)
	out.clock_tolerance_s = C.double(value.ClockToleranceS)
	out.min_agree = minAgree
	out.clock_min_common = clockMinCommon
	out.combine = C.uint32_t(value.Combine)
	out.precedence_scope = C.uint32_t(value.PrecedenceScope)
	out.outlier_reject_enabled = cBool(value.OutlierRejectEnabled)
	out.outlier_reject_position_tolerance_m = C.double(value.OutlierRejectPositionTolerance)
	out.outlier_reject_clock_tolerance_s = C.double(value.OutlierRejectClockTolerance)
	out.target_epoch_interval_s_enabled = cBool(value.TargetEpochIntervalEnabled)
	out.target_epoch_interval_s = C.double(value.TargetEpochIntervalS)
	out.helmert_frame_reconciliation = cBool(value.HelmertFrameReconciliation)
	if len(value.Systems) != 0 {
		systemsPointer, err := a.malloc(len(value.Systems), unsafe.Sizeof(C.uint32_t(0)))
		if err != nil {
			return nil, err
		}
		systems := unsafe.Slice((*C.uint32_t)(systemsPointer), len(value.Systems))
		for i, system := range value.Systems {
			systems[i] = C.uint32_t(system)
		}
		out.systems = (*C.uint32_t)(systemsPointer)
		systemCount, err := checkedNativeSize(len(value.Systems))
		if err != nil {
			return nil, err
		}
		out.system_count = systemCount
	}
	if len(value.AssertedFrameLabelSets) != 0 {
		setsPointer, err := a.malloc(len(value.AssertedFrameLabelSets), unsafe.Sizeof(C.SidereonSp3FrameLabelSet{}))
		if err != nil {
			return nil, err
		}
		sets := unsafe.Slice((*C.SidereonSp3FrameLabelSet)(setsPointer), len(value.AssertedFrameLabelSets))
		for i, labels := range value.AssertedFrameLabelSets {
			labelPointer, count, err := a.cStringArray(labels, "SP3 frame label")
			if err != nil {
				return nil, err
			}
			sets[i].labels = (**C.char)(labelPointer)
			sets[i].label_count = count
		}
		out.asserted_frame_label_sets = (*C.SidereonSp3FrameLabelSet)(setsPointer)
		setCount, err := checkedNativeSize(len(value.AssertedFrameLabelSets))
		if err != nil {
			return nil, err
		}
		out.asserted_frame_label_set_count = setCount
	}
	return out, nil
}

func Sp3MergeOptionsInit() (NativeSp3MergeOptions, error) {
	var value C.SidereonSp3MergeOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_sp3_merge_options_init(&value)) })
	if err != nil {
		return NativeSp3MergeOptions{}, err
	}
	minAgree, err := checkedNativeCount(uint64(value.min_agree))
	if err != nil {
		return NativeSp3MergeOptions{}, err
	}
	clockMinCommon, err := checkedNativeCount(uint64(value.clock_min_common))
	if err != nil {
		return NativeSp3MergeOptions{}, err
	}
	if err := validateSp3MergeCombine(uint32(value.combine)); err != nil {
		return NativeSp3MergeOptions{}, err
	}
	if err := validateSp3MergePrecedenceScope(uint32(value.precedence_scope)); err != nil {
		return NativeSp3MergeOptions{}, err
	}
	if value.outlier_reject_enabled > 1 || value.target_epoch_interval_s_enabled > 1 || value.helmert_frame_reconciliation > 1 {
		return NativeSp3MergeOptions{}, invalidArgument("invalid SP3 merge option boolean returned by native code")
	}
	return NativeSp3MergeOptions{
		PositionToleranceM: float64(value.position_tolerance_m), ClockToleranceS: float64(value.clock_tolerance_s), MinAgree: minAgree, ClockMinCommon: clockMinCommon, Combine: uint32(value.combine), PrecedenceScope: uint32(value.precedence_scope), OutlierRejectEnabled: value.outlier_reject_enabled != 0, OutlierRejectPositionTolerance: float64(value.outlier_reject_position_tolerance_m), OutlierRejectClockTolerance: float64(value.outlier_reject_clock_tolerance_s), TargetEpochIntervalEnabled: value.target_epoch_interval_s_enabled != 0, TargetEpochIntervalS: float64(value.target_epoch_interval_s), HelmertFrameReconciliation: value.helmert_frame_reconciliation != 0,
	}, nil
}

func ExactSp3RequestNew(year int, month, day uint8, issue, span, sample, expectedAgency string) (*ExactSp3Request, error) {
	cYear, err := checkedCatalogYear(year)
	if err != nil {
		return nil, err
	}
	var out *C.SidereonExactSp3Request
	withCThread(func() {
		var arena sp3Arena
		defer arena.close()
		var issuePointer, spanPointer, samplePointer, agencyPointer *C.char
		if issue != "" {
			issuePointer, err = arena.string(issue, "SP3 exact issue")
			if err != nil {
				return
			}
		}
		spanPointer, err = arena.string(span, "SP3 exact span")
		if err != nil {
			return
		}
		samplePointer, err = arena.string(sample, "SP3 exact sample")
		if err != nil {
			return
		}
		if expectedAgency != "" {
			agencyPointer, err = arena.string(expectedAgency, "SP3 expected agency")
			if err != nil {
				return
			}
		}
		err = statusErrorLocked(uint32(C.sidereon_sp3_exact_request_new(cYear, C.uint8_t(month), C.uint8_t(day), issuePointer, spanPointer, samplePointer, agencyPointer, &out)))
		if err != nil && out != nil {
			C.sidereon_sp3_exact_request_free(out)
			out = nil
		}
	})
	if err != nil {
		return nil, err
	}
	return newExactSp3Request(out)
}

func ExactSp3RequestFromIdentity(identity ProductIdentity) (*ExactSp3Request, error) {
	cIdentity, err := cProductIdentity(identity)
	if err != nil {
		return nil, err
	}
	var out *C.SidereonExactSp3Request
	withCThread(func() {
		err = statusErrorLocked(uint32(C.sidereon_sp3_exact_request_from_identity(&cIdentity, &out)))
		if err != nil && out != nil {
			C.sidereon_sp3_exact_request_free(out)
			out = nil
		}
	})
	runtime.KeepAlive(identity)
	if err != nil {
		return nil, err
	}
	return newExactSp3Request(out)
}

func LoadExactSP3(data []byte, request *ExactSp3Request) (*SP3, uint32, error) {
	if request == nil || request.handle == nil {
		return nil, 0, ErrClosed
	}
	var out *C.SidereonSp3
	var coverage C.enum_SidereonExactSp3Coverage
	var operationErr error
	err := request.handle.with(func(requestPointer unsafe.Pointer) error {
		return withInputError(data, func(bytes *C.uint8_t, length C.size_t) error {
			withCThread(func() {
				operationErr = statusErrorLocked(uint32(C.sidereon_sp3_load_exact(bytes, length, (*C.SidereonExactSp3Request)(requestPointer), &out, &coverage)))
				if operationErr != nil && out != nil {
					C.sidereon_sp3_free(out)
					out = nil
				}
			})
			return operationErr
		})
	})
	runtime.KeepAlive(request)
	runtime.KeepAlive(data)
	if err != nil {
		return nil, 0, err
	}
	if err := validateExactSp3Coverage(uint32(coverage)); err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_sp3_free(out) })
		}
		return nil, 0, err
	}
	if out == nil {
		return nil, 0, missingNativeHandle("exact SP3")
	}
	return &SP3{handle: newPositioningHandle(unsafe.Pointer(out), releaseSP3)}, uint32(coverage), nil
}

func ValidateExactSP3(sp3 *SP3, request *ExactSp3Request) (uint32, error) {
	if sp3 == nil || request == nil || sp3.handle == nil || request.handle == nil {
		return 0, ErrClosed
	}
	var coverageValue uint32
	err := withPositioningHandleSet([]*positioningHandle{sp3.handle, request.handle}, func(pointers []unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			var coverage C.enum_SidereonExactSp3Coverage
			operationErr = statusErrorLocked(uint32(C.sidereon_sp3_validate_exact((*C.SidereonSp3)(pointers[0]), (*C.SidereonExactSp3Request)(pointers[1]), &coverage)))
			if operationErr == nil {
				operationErr = validateExactSp3Coverage(uint32(coverage))
				coverageValue = uint32(coverage)
			}
		})
		return operationErr
	})
	return coverageValue, err
}

func MergeSP3(sources []*SP3, options *NativeSp3MergeOptions) (*SP3, *Sp3MergeReport, error) {
	handles := make([]*positioningHandle, len(sources))
	for i, source := range sources {
		if source == nil || source.handle == nil {
			return nil, nil, ErrClosed
		}
		handles[i] = source.handle
	}
	var out *C.SidereonSp3
	var report *C.SidereonSp3MergeReport
	var arena sp3Arena
	defer arena.close()
	var optionPointer *C.SidereonSp3MergeOptions
	var err error
	if options != nil {
		optionPointer, err = cSp3MergeOptions(&arena, *options)
		if err != nil {
			return nil, nil, err
		}
	}
	err = withPositioningHandleSet(handles, func(pointers []unsafe.Pointer) error {
		count, countErr := checkedNativeSize(len(pointers))
		if countErr != nil {
			return countErr
		}
		var sourceMemory unsafe.Pointer
		if len(pointers) != 0 {
			sourceMemory, err = arena.malloc(len(pointers), unsafe.Sizeof(uintptr(0)))
			if err != nil {
				return err
			}
			values := unsafe.Slice((**C.SidereonSp3)(sourceMemory), len(pointers))
			for i, pointer := range pointers {
				values[i] = (*C.SidereonSp3)(pointer)
			}
		}
		var sourcesPointer **C.SidereonSp3
		if sourceMemory != nil {
			sourcesPointer = (**C.SidereonSp3)(sourceMemory)
		}
		withCThread(func() {
			err = statusErrorLocked(uint32(C.sidereon_sp3_merge(sourcesPointer, count, optionPointer, &out, &report)))
			if err != nil {
				if out != nil {
					C.sidereon_sp3_free(out)
					out = nil
				}
				if report != nil {
					C.sidereon_sp3_merge_report_free(report)
					report = nil
				}
			}
		})
		return err
	})
	runtime.KeepAlive(sources)
	if err != nil {
		return nil, nil, err
	}
	if out == nil || report == nil {
		if out != nil {
			withCThread(func() { C.sidereon_sp3_free(out) })
		}
		if report != nil {
			withCThread(func() { C.sidereon_sp3_merge_report_free(report) })
		}
		return nil, nil, missingNativeHandle("SP3 merge output")
	}
	return &SP3{handle: newPositioningHandle(unsafe.Pointer(out), releaseSP3)}, &Sp3MergeReport{handle: newPositioningHandle(unsafe.Pointer(report), releaseSp3MergeReport)}, nil
}

type NativeSp3ArtifactIdentity struct {
	RequestedIdentity ProductIdentity
	ResolvedIdentity  ProductIdentity
	Distribution      uint32
	OfficialFilename  string
	ProductSHA256     string
	ProductByteLength uint64
	ArchiveSHA256     string
	ArchiveByteLength uint64
	Compression       uint32
}

func cSp3ArtifactIdentity(value NativeSp3ArtifactIdentity) (C.SidereonSp3ArtifactIdentity, error) {
	var out C.SidereonSp3ArtifactIdentity
	requested, err := cProductIdentity(value.RequestedIdentity)
	if err != nil {
		return out, err
	}
	resolved, err := cProductIdentity(value.ResolvedIdentity)
	if err != nil {
		return out, err
	}
	out.requested_identity = requested
	out.resolved_identity = resolved
	if value.Distribution > uint32(C.SIDEREON_DISTRIBUTION_SOURCE_IN_MEMORY) {
		return out, invalidArgument("invalid SP3 artifact distribution source")
	}
	if value.Compression > uint32(C.SIDEREON_ARCHIVE_COMPRESSION_UNIX_COMPRESS) {
		return out, invalidArgument("invalid SP3 artifact compression")
	}
	out.distribution_source = C.uint32_t(value.Distribution)
	out.product_byte_length = C.uint64_t(value.ProductByteLength)
	out.archive_byte_length = C.uint64_t(value.ArchiveByteLength)
	out.compression = C.uint32_t(value.Compression)
	if err := copyFixedCString(unsafe.Pointer(&out.official_filename[0]), len(out.official_filename), value.OfficialFilename); err != nil {
		return out, err
	}
	if err := copyFixedCString(unsafe.Pointer(&out.product_sha256[0]), len(out.product_sha256), value.ProductSHA256); err != nil {
		return out, err
	}
	if err := copyFixedCString(unsafe.Pointer(&out.archive_sha256[0]), len(out.archive_sha256), value.ArchiveSHA256); err != nil {
		return out, err
	}
	return out, nil
}

func sp3ArtifactIdentityFromC(value C.SidereonSp3ArtifactIdentity) (NativeSp3ArtifactIdentity, error) {
	if uint32(value.distribution_source) > uint32(C.SIDEREON_DISTRIBUTION_SOURCE_IN_MEMORY) {
		return NativeSp3ArtifactIdentity{}, invalidArgument("invalid SP3 artifact distribution source returned by native code")
	}
	if uint32(value.compression) > uint32(C.SIDEREON_ARCHIVE_COMPRESSION_UNIX_COMPRESS) {
		return NativeSp3ArtifactIdentity{}, invalidArgument("invalid SP3 artifact compression returned by native code")
	}
	return NativeSp3ArtifactIdentity{
		RequestedIdentity: identityFromC(&value.requested_identity), ResolvedIdentity: identityFromC(&value.resolved_identity), Distribution: uint32(value.distribution_source), OfficialFilename: C.GoString((*C.char)(unsafe.Pointer(&value.official_filename[0]))), ProductSHA256: C.GoString((*C.char)(unsafe.Pointer(&value.product_sha256[0]))), ProductByteLength: uint64(value.product_byte_length), ArchiveSHA256: C.GoString((*C.char)(unsafe.Pointer(&value.archive_sha256[0]))), ArchiveByteLength: uint64(value.archive_byte_length), Compression: uint32(value.compression),
	}, nil
}

func MergeInputIdentity(contributors []NativeSp3ArtifactIdentity, options *NativeSp3MergeOptions) (*Sp3MergeInputIdentity, error) {
	var arena sp3Arena
	defer arena.close()
	var optionPointer *C.SidereonSp3MergeOptions
	var err error
	if options != nil {
		optionPointer, err = cSp3MergeOptions(&arena, *options)
		if err != nil {
			return nil, err
		}
	}
	var contributorPointer *C.SidereonSp3ArtifactIdentity
	if len(contributors) != 0 {
		memory, allocationErr := arena.malloc(len(contributors), unsafe.Sizeof(C.SidereonSp3ArtifactIdentity{}))
		if allocationErr != nil {
			return nil, allocationErr
		}
		values := unsafe.Slice((*C.SidereonSp3ArtifactIdentity)(memory), len(contributors))
		for i, contributor := range contributors {
			values[i], err = cSp3ArtifactIdentity(contributor)
			if err != nil {
				return nil, fmt.Errorf("sidereon: SP3 merge contributor %d: %w", i, err)
			}
		}
		contributorPointer = (*C.SidereonSp3ArtifactIdentity)(memory)
	}
	count, err := checkedNativeSize(len(contributors))
	if err != nil {
		return nil, err
	}
	var out *C.SidereonSp3MergeInputIdentity
	withCThread(func() {
		err = statusErrorLocked(uint32(C.sidereon_sp3_merge_input_identity(contributorPointer, count, optionPointer, &out)))
		if err != nil && out != nil {
			C.sidereon_sp3_merge_input_identity_free(out)
			out = nil
		}
	})
	runtime.KeepAlive(contributors)
	if err != nil {
		return nil, err
	}
	return newSp3MergeInputIdentity(out)
}

func (i *Sp3MergeInputIdentity) ContributorCount() (int, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	var value C.size_t
	err := i.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_input_identity_contributor_count((*C.SidereonSp3MergeInputIdentity)(pointer), &value))
		})
	})
	runtime.KeepAlive(i)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(value))
}

func (i *Sp3MergeInputIdentity) Contributor(index int) (NativeSp3ArtifactIdentity, error) {
	return i.artifactAt(index, false)
}

func (i *Sp3MergeInputIdentity) PrecedenceContributorCount() (bool, int, error) {
	if i == nil || i.handle == nil {
		return false, 0, ErrClosed
	}
	var present C.uint8_t
	var value C.size_t
	err := i.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_input_identity_precedence_contributor_count((*C.SidereonSp3MergeInputIdentity)(pointer), &present, &value))
		})
	})
	runtime.KeepAlive(i)
	if present > 1 {
		if err == nil {
			err = invalidArgument("invalid SP3 precedence presence returned by native code")
		}
		return false, 0, err
	}
	if err != nil {
		return false, 0, err
	}
	count, countErr := checkedNativeCount(uint64(value))
	return present != 0, count, countErr
}

func (i *Sp3MergeInputIdentity) PrecedenceContributor(index int) (NativeSp3ArtifactIdentity, error) {
	return i.artifactAt(index, true)
}

func (i *Sp3MergeInputIdentity) artifactAt(index int, precedence bool) (NativeSp3ArtifactIdentity, error) {
	if i == nil || i.handle == nil {
		return NativeSp3ArtifactIdentity{}, ErrClosed
	}
	if index < 0 {
		return NativeSp3ArtifactIdentity{}, invalidArgument("SP3 merge contributor index must not be negative")
	}
	valueIndex, err := checkedNativeSize(index)
	if err != nil {
		return NativeSp3ArtifactIdentity{}, err
	}
	var value C.SidereonSp3ArtifactIdentity
	err = i.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			if precedence {
				return uint32(C.sidereon_sp3_merge_input_identity_precedence_contributor((*C.SidereonSp3MergeInputIdentity)(pointer), valueIndex, &value))
			}
			return uint32(C.sidereon_sp3_merge_input_identity_contributor((*C.SidereonSp3MergeInputIdentity)(pointer), valueIndex, &value))
		})
	})
	runtime.KeepAlive(i)
	if err != nil {
		return NativeSp3ArtifactIdentity{}, err
	}
	return sp3ArtifactIdentityFromC(value)
}

func (i *Sp3MergeInputIdentity) SchemaVersion() (uint8, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	var value C.uint8_t
	err := i.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_input_identity_schema_version((*C.SidereonSp3MergeInputIdentity)(pointer), &value))
		})
	})
	runtime.KeepAlive(i)
	return uint8(value), err
}

func (i *Sp3MergeInputIdentity) StableID() ([]byte, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := i.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var copyErr error
			result, copyErr = copyNativeBytesLocked("SP3 merge stable ID", func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_input_identity_stable_id((*C.SidereonSp3MergeInputIdentity)(pointer), out, n, written, required)
			})
			return copyErr
		})
	})
	runtime.KeepAlive(i)
	return result, err
}

type NativeSp3AgreementSummary struct {
	PositionRMSPresent bool
	PositionRMSM       float64
	PositionMaxPresent bool
	PositionMaxM       float64
	ClockRMSPresent    bool
	ClockRMSS          float64
	ClockMaxPresent    bool
	ClockMaxS          float64
}

type NativeSp3EpochAgreement struct {
	EpochJ2000S     float64
	Satellites      int
	PositionRMSM    float64
	PositionMaxM    float64
	ClockRMSPresent bool
	ClockRMSS       float64
	ClockMaxPresent bool
	ClockMaxS       float64
}

type NativeSp3MergeFlag struct {
	EpochJ2000S float64
	Satellite   string
	SourceCount int
}

type NativeSp3FrameReconciliation struct {
	SourceIndex               int
	SourceLabelLen            int
	TargetLabelLen            int
	Method                    uint32
	AssertedLabelCount        int
	SourceFramePresent        bool
	SourceFrame               uint32
	TargetFramePresent        bool
	TargetFrame               uint32
	CatalogFramePresent       bool
	CatalogSourceFrame        uint32
	CatalogTargetFrame        uint32
	CatalogInverse            bool
	ReferenceEpochYearPresent bool
	ReferenceEpochYear        float64
	ParametersPresent         bool
	TranslationMM             [3]float64
	ScalePPB                  float64
	RotationMAS               [3]float64
	RatesPresent              bool
	TranslationMMPerYear      [3]float64
	ScalePPBPerYear           float64
	RotationMASPerYear        [3]float64
	ProvenanceLen             int
	EpochYearSpanPresent      bool
	EpochYearStart            float64
	EpochYearEnd              float64
	RecordsAffected           int
	Identity                  bool
}

func sp3AgreementFromC(value C.SidereonSp3AgreementSummary) NativeSp3AgreementSummary {
	return NativeSp3AgreementSummary{PositionRMSPresent: bool(value.position_rms_present), PositionRMSM: float64(value.position_rms_m), PositionMaxPresent: bool(value.position_max_present), PositionMaxM: float64(value.position_max_m), ClockRMSPresent: bool(value.clock_rms_present), ClockRMSS: float64(value.clock_rms_s), ClockMaxPresent: bool(value.clock_max_present), ClockMaxS: float64(value.clock_max_s)}
}

func sp3EpochAgreementFromC(value C.SidereonSp3EpochAgreement) (NativeSp3EpochAgreement, error) {
	satellites, err := checkedNativeCount(uint64(value.satellites))
	if err != nil {
		return NativeSp3EpochAgreement{}, err
	}
	return NativeSp3EpochAgreement{EpochJ2000S: float64(value.epoch_j2000_seconds), Satellites: satellites, PositionRMSM: float64(value.position_rms_m), PositionMaxM: float64(value.position_max_m), ClockRMSPresent: bool(value.clock_rms_present), ClockRMSS: float64(value.clock_rms_s), ClockMaxPresent: bool(value.clock_max_present), ClockMaxS: float64(value.clock_max_s)}, nil
}

func sp3MergeFlagFromC(value C.SidereonSp3MergeFlag) (NativeSp3MergeFlag, error) {
	sourceCount, err := checkedNativeCount(uint64(value.source_count))
	if err != nil {
		return NativeSp3MergeFlag{}, err
	}
	return NativeSp3MergeFlag{EpochJ2000S: float64(value.epoch_j2000_seconds), Satellite: tokenFromC(value.sat_id), SourceCount: sourceCount}, nil
}

func sp3FrameReconciliationFromC(value C.SidereonSp3FrameReconciliation) (NativeSp3FrameReconciliation, error) {
	if err := validateSp3FrameReconciliationMethod(uint32(value.method)); err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	for _, frame := range []uint32{uint32(value.source_frame), uint32(value.target_frame), uint32(value.catalog_source_frame), uint32(value.catalog_target_frame)} {
		if err := validateSp3TerrestrialFrame(frame); err != nil {
			return NativeSp3FrameReconciliation{}, err
		}
	}
	convert := func(v C.size_t, label string) (int, error) {
		out, err := checkedNativeCount(uint64(v))
		if err != nil {
			return 0, fmt.Errorf("sidereon: native SP3 reconciliation %s: %w", label, err)
		}
		return out, nil
	}
	sourceIndex, err := convert(value.source_index, "source index")
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	sourceLabelLen, err := convert(value.source_label_len, "source label length")
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	targetLabelLen, err := convert(value.target_label_len, "target label length")
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	assertedCount, err := convert(value.asserted_label_count, "asserted label count")
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	provenanceLen, err := convert(value.provenance_len, "provenance length")
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	records, err := convert(value.records_affected, "records affected")
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	return NativeSp3FrameReconciliation{SourceIndex: sourceIndex, SourceLabelLen: sourceLabelLen, TargetLabelLen: targetLabelLen, Method: uint32(value.method), AssertedLabelCount: assertedCount, SourceFramePresent: bool(value.source_frame_present), SourceFrame: uint32(value.source_frame), TargetFramePresent: bool(value.target_frame_present), TargetFrame: uint32(value.target_frame), CatalogFramePresent: bool(value.catalog_frame_present), CatalogSourceFrame: uint32(value.catalog_source_frame), CatalogTargetFrame: uint32(value.catalog_target_frame), CatalogInverse: bool(value.catalog_inverse), ReferenceEpochYearPresent: bool(value.reference_epoch_year_present), ReferenceEpochYear: float64(value.reference_epoch_year), ParametersPresent: bool(value.parameters_present), ScalePPB: float64(value.scale_ppb), RatesPresent: bool(value.rates_present), ScalePPBPerYear: float64(value.scale_ppb_per_year), ProvenanceLen: provenanceLen, EpochYearSpanPresent: bool(value.epoch_year_span_present), EpochYearStart: float64(value.epoch_year_start), EpochYearEnd: float64(value.epoch_year_end), RecordsAffected: records, Identity: bool(value.identity), TranslationMM: [3]float64{float64(value.translation_mm[0]), float64(value.translation_mm[1]), float64(value.translation_mm[2])}, RotationMAS: [3]float64{float64(value.rotation_mas[0]), float64(value.rotation_mas[1]), float64(value.rotation_mas[2])}, TranslationMMPerYear: [3]float64{float64(value.translation_mm_per_year[0]), float64(value.translation_mm_per_year[1]), float64(value.translation_mm_per_year[2])}, RotationMASPerYear: [3]float64{float64(value.rotation_mas_per_year[0]), float64(value.rotation_mas_per_year[1]), float64(value.rotation_mas_per_year[2])}}, nil
}

func (r *Sp3MergeReport) AgreementSummary() (NativeSp3AgreementSummary, error) {
	if r == nil || r.handle == nil {
		return NativeSp3AgreementSummary{}, ErrClosed
	}
	var value C.SidereonSp3AgreementSummary
	err := r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_agreement_summary((*C.SidereonSp3MergeReport)(pointer), &value))
		})
	})
	runtime.KeepAlive(r)
	return sp3AgreementFromC(value), err
}

func (r *Sp3MergeReport) ContinuityVerdictJSON(merged *SP3, from, through float64) ([]byte, error) {
	if r == nil || r.handle == nil || merged == nil || merged.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := withPositioningHandleSet([]*positioningHandle{r.handle, merged.handle}, func(pointers []unsafe.Pointer) error {
		return withCThreadError(func() error {
			var copyErr error
			result, copyErr = copyNativeBytesLocked("SP3 merge continuity verdict", func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_continuity_verdict_json((*C.SidereonSp3MergeReport)(pointers[0]), (*C.SidereonSp3)(pointers[1]), C.double(from), C.double(through), out, n, written, required)
			})
			return copyErr
		})
	})
	runtime.KeepAlive(r)
	runtime.KeepAlive(merged)
	return result, err
}

func (r *Sp3MergeReport) EpochAgreementCount() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	var value C.size_t
	err := r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_epoch_agreement_count((*C.SidereonSp3MergeReport)(pointer), &value))
		})
	})
	runtime.KeepAlive(r)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(value))
}

func (r *Sp3MergeReport) EpochAgreement(index int) (NativeSp3EpochAgreement, error) {
	if r == nil || r.handle == nil {
		return NativeSp3EpochAgreement{}, ErrClosed
	}
	if index < 0 {
		return NativeSp3EpochAgreement{}, invalidArgument("SP3 merge epoch agreement index must not be negative")
	}
	valueIndex, err := checkedNativeSize(index)
	if err != nil {
		return NativeSp3EpochAgreement{}, err
	}
	var value C.SidereonSp3EpochAgreement
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_epoch_agreement((*C.SidereonSp3MergeReport)(pointer), valueIndex, &value))
		})
	})
	runtime.KeepAlive(r)
	if err != nil {
		return NativeSp3EpochAgreement{}, err
	}
	return sp3EpochAgreementFromC(value)
}

func (r *Sp3MergeReport) FlagCount(kind uint32) (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	if err := validateSp3MergeFlagKind(kind); err != nil {
		return 0, err
	}
	var value C.size_t
	err := r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_flag_count((*C.SidereonSp3MergeReport)(pointer), C.uint32_t(kind), &value))
		})
	})
	runtime.KeepAlive(r)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(value))
}

func (r *Sp3MergeReport) Flag(kind uint32, index int) (NativeSp3MergeFlag, error) {
	if r == nil || r.handle == nil {
		return NativeSp3MergeFlag{}, ErrClosed
	}
	if err := validateSp3MergeFlagKind(kind); err != nil {
		return NativeSp3MergeFlag{}, err
	}
	if index < 0 {
		return NativeSp3MergeFlag{}, invalidArgument("SP3 merge flag index must not be negative")
	}
	valueIndex, err := checkedNativeSize(index)
	if err != nil {
		return NativeSp3MergeFlag{}, err
	}
	var value C.SidereonSp3MergeFlag
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_flag((*C.SidereonSp3MergeReport)(pointer), C.uint32_t(kind), valueIndex, &value))
		})
	})
	runtime.KeepAlive(r)
	if err != nil {
		return NativeSp3MergeFlag{}, err
	}
	return sp3MergeFlagFromC(value)
}

func (r *Sp3MergeReport) FlagSources(kind uint32, index int) ([]int, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	if err := validateSp3MergeFlagKind(kind); err != nil {
		return nil, err
	}
	if index < 0 {
		return nil, invalidArgument("SP3 merge flag index must not be negative")
	}
	valueIndex, err := checkedNativeSize(index)
	if err != nil {
		return nil, err
	}
	var result []int
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var copyErr error
			result, copyErr = copyNativeSizeTsLocked("SP3 merge flag sources", func(out *C.size_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_flag_sources((*C.SidereonSp3MergeReport)(pointer), C.uint32_t(kind), valueIndex, out, n, written, required)
			})
			return copyErr
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func copyNativeSizeTsLocked(label string, call func(*C.size_t, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]int, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.size_t(0))); err != nil {
		return nil, err
	}
	values := make([]C.size_t, count)
	var out *C.size_t
	if count != 0 {
		out = &values[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(out, C.size_t(count), &written, &required))); err != nil {
		return nil, err
	}
	writtenCount, err := validateNativeOutput(label, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]int, writtenCount)
	for i := range result {
		result[i], err = checkedNativeCount(uint64(values[i]))
		if err != nil {
			return nil, fmt.Errorf("sidereon: native %s index: %w", label, err)
		}
	}
	return result, nil
}

func (r *Sp3MergeReport) FrameReconciliationCount() (int, error) {
	if r == nil || r.handle == nil {
		return 0, ErrClosed
	}
	var value C.size_t
	err := r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_frame_reconciliation_count((*C.SidereonSp3MergeReport)(pointer), &value))
		})
	})
	runtime.KeepAlive(r)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(value))
}

func (r *Sp3MergeReport) FrameReconciliation(index int) (NativeSp3FrameReconciliation, error) {
	if r == nil || r.handle == nil {
		return NativeSp3FrameReconciliation{}, ErrClosed
	}
	if index < 0 {
		return NativeSp3FrameReconciliation{}, invalidArgument("SP3 frame reconciliation index must not be negative")
	}
	valueIndex, err := checkedNativeSize(index)
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	var value C.SidereonSp3FrameReconciliation
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_frame_reconciliation((*C.SidereonSp3MergeReport)(pointer), valueIndex, &value))
		})
	})
	runtime.KeepAlive(r)
	if err != nil {
		return NativeSp3FrameReconciliation{}, err
	}
	return sp3FrameReconciliationFromC(value)
}

func (r *Sp3MergeReport) frameBytes(index, labelIndex int, kind string) ([]byte, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 || labelIndex < 0 {
		return nil, invalidArgument("SP3 frame reconciliation index must not be negative")
	}
	valueIndex, err := checkedNativeSize(index)
	if err != nil {
		return nil, err
	}
	labelValue, err := checkedNativeSize(labelIndex)
	if err != nil {
		return nil, err
	}
	var result []byte
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var copyErr error
			result, copyErr = copyNativeBytesLocked(kind, func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_frame_reconciliation_asserted_label((*C.SidereonSp3MergeReport)(pointer), valueIndex, labelValue, out, n, written, required)
			})
			return copyErr
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) AssertedLabel(index, labelIndex int) ([]byte, error) {
	return r.frameBytes(index, labelIndex, "SP3 asserted frame label")
}

func (r *Sp3MergeReport) Provenance(index int) ([]byte, error) {
	return r.singleFrameBytes(index, "SP3 frame provenance", func(report *C.SidereonSp3MergeReport, i C.size_t, out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_sp3_merge_report_frame_reconciliation_provenance(report, i, out, n, written, required)
	})
}

func (r *Sp3MergeReport) SourceLabel(index int) ([]byte, error) {
	return r.singleFrameBytes(index, "SP3 source frame label", func(report *C.SidereonSp3MergeReport, i C.size_t, out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_sp3_merge_report_frame_reconciliation_source_label(report, i, out, n, written, required)
	})
}

func (r *Sp3MergeReport) TargetLabel(index int) ([]byte, error) {
	return r.singleFrameBytes(index, "SP3 target frame label", func(report *C.SidereonSp3MergeReport, i C.size_t, out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_sp3_merge_report_frame_reconciliation_target_label(report, i, out, n, written, required)
	})
}

func (r *Sp3MergeReport) singleFrameBytes(index int, label string, call func(*C.SidereonSp3MergeReport, C.size_t, *C.uint8_t, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]byte, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("SP3 frame reconciliation index must not be negative")
	}
	valueIndex, err := checkedNativeSize(index)
	if err != nil {
		return nil, err
	}
	var result []byte
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var copyErr error
			result, copyErr = copyNativeBytesLocked(label, func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return call((*C.SidereonSp3MergeReport)(pointer), valueIndex, out, n, written, required)
			})
			return copyErr
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (s *SP3) ObservableState(satellite string, epoch float64) ([3]float64, float64, bool, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, 0, false, ErrClosed
	}
	var position [3]C.double
	var clock C.double
	var hasClock C.bool
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withTokenError(satellite, "satellite token", func(id *C.char) error {
			return statusErrorLocked(uint32(C.sidereon_sp3_observable_state((*C.SidereonSp3)(pointer), id, C.double(epoch), &position[0], &clock, &hasClock)))
		})
	})
	runtime.KeepAlive(s)
	return [3]float64{float64(position[0]), float64(position[1]), float64(position[2])}, float64(clock), bool(hasClock), err
}

func sp3ObservableStates(h *positioningHandle, satellites []string, epochs []float64, shared bool, epoch float64) ([]NativeObservableStateRow, error) {
	if !shared && len(satellites) != len(epochs) {
		return nil, errors.New("sidereon: satellite and epoch lengths differ")
	}
	var result []NativeObservableStateRow
	err := h.with(func(pointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(ids **C.char, count C.size_t) error {
			n := len(satellites)
			positionCount, err := checkedNativeProduct(n, 3, "SP3 observable-state position")
			if err != nil {
				return err
			}
			for _, item := range []struct {
				count int
				size  uintptr
				name  string
			}{{positionCount, unsafe.Sizeof(C.double(0)), "SP3 observable-state positions"}, {n, unsafe.Sizeof(C.double(0)), "SP3 observable-state clocks"}, {n, unsafe.Sizeof(C.bool(false)), "SP3 observable-state clock flags"}, {n, unsafe.Sizeof(C.enum_SidereonObservableStateElementStatus(0)), "SP3 observable-state element statuses"}, {n, unsafe.Sizeof(C.enum_SidereonStatus(0)), "SP3 observable-state result statuses"}} {
				if _, err := checkedNativeAllocationSize(item.count, item.size); err != nil {
					return fmt.Errorf("sidereon: %s: %w", item.name, err)
				}
			}
			positions := make([]C.double, positionCount)
			clocks := make([]C.double, n)
			hasClocks := make([]C.bool, n)
			elements := make([]C.enum_SidereonObservableStateElementStatus, n)
			statuses := make([]C.enum_SidereonStatus, n)
			var epochMemory unsafe.Pointer
			if !shared {
				epochMemory, err = checkedNativeMalloc(n, unsafe.Sizeof(C.double(0)))
				if err != nil {
					return err
				}
				defer C.free(epochMemory)
				values := unsafe.Slice((*C.double)(epochMemory), n)
				for i, value := range epochs {
					values[i] = C.double(value)
				}
			}
			var positionPointer, clockPointer *C.double
			var hasClockPointer *C.bool
			var elementPointer *C.enum_SidereonObservableStateElementStatus
			var statusPointer *C.enum_SidereonStatus
			var epochPointer *C.double
			if n != 0 {
				positionPointer, clockPointer, hasClockPointer = &positions[0], &clocks[0], &hasClocks[0]
				elementPointer, statusPointer = &elements[0], &statuses[0]
				if !shared {
					epochPointer = (*C.double)(epochMemory)
				}
			}
			var status uint32
			if shared {
				status = uint32(C.sidereon_sp3_observable_states_at_shared_j2000_s((*C.SidereonSp3)(pointer), ids, count, C.double(epoch), positionPointer, clockPointer, hasClockPointer, elementPointer, statusPointer))
			} else {
				status = uint32(C.sidereon_sp3_observable_states_at_j2000_s((*C.SidereonSp3)(pointer), ids, (*C.double)(epochPointer), count, positionPointer, clockPointer, hasClockPointer, elementPointer, statusPointer))
			}
			if err := statusErrorLocked(status); err != nil {
				return err
			}
			result = make([]NativeObservableStateRow, n)
			for i := range result {
				if err := validateObservableStateElementStatusValue(uint32(elements[i])); err != nil {
					return err
				}
				if err := validateObservableStateResultStatusValue(uint32(statuses[i])); err != nil {
					return err
				}
				result[i].ClockS, result[i].HasClock = float64(clocks[i]), bool(hasClocks[i])
				result[i].ElementStatus, result[i].ResultStatus = uint32(elements[i]), uint32(statuses[i])
				for axis := range result[i].Position {
					result[i].Position[axis] = float64(positions[i*3+axis])
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(satellites)
	runtime.KeepAlive(epochs)
	return result, err
}

func (s *SP3) ObservableStates(satellites []string, epochs []float64) ([]NativeObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return sp3ObservableStates(s.handle, append([]string(nil), satellites...), append([]float64(nil), epochs...), false, 0)
}

func (s *SP3) ObservableStatesShared(satellites []string, epoch float64) ([]NativeObservableStateRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return sp3ObservableStates(s.handle, append([]string(nil), satellites...), nil, true, epoch)
}

func (s *SP3) PredictObservables(satellite string, receiver [3]float64, epoch float64, options *NativeObservablesOptions) (NativePredictedObservables, error) {
	if s == nil || s.handle == nil {
		return NativePredictedObservables{}, ErrClosed
	}
	var value C.SidereonPredictedObservables
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withTokenError(satellite, "satellite token", func(id *C.char) error {
			var receiverMemory unsafe.Pointer
			var err error
			receiverMemory, err = checkedNativeMalloc(3, unsafe.Sizeof(C.double(0)))
			if err != nil {
				return err
			}
			defer C.free(receiverMemory)
			receiverValues := unsafe.Slice((*C.double)(receiverMemory), 3)
			for i := range receiver {
				receiverValues[i] = C.double(receiver[i])
			}
			var optionsMemory unsafe.Pointer
			var optionsPointer *C.SidereonObservablesOptions
			if options != nil {
				optionsMemory, err = checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservablesOptions{}))
				if err != nil {
					return err
				}
				defer C.free(optionsMemory)
				optionsPointer = (*C.SidereonObservablesOptions)(optionsMemory)
				*optionsPointer = cObservablesOptions(*options)
			}
			return statusErrorLocked(uint32(C.sidereon_sp3_observables((*C.SidereonSp3)(pointer), id, (*C.double)(receiverMemory), C.double(epoch), optionsPointer, &value)))
		})
	})
	runtime.KeepAlive(s)
	return predictedObservablesFromC(value), err
}

func (s *SP3) PredictObservablesBatch(requests []NativePredictRequest, options *NativeObservablesOptions) ([]NativePredictedObservables, []bool, error) {
	if s == nil || s.handle == nil {
		return nil, nil, ErrClosed
	}
	var result []NativePredictedObservables
	var accepted []bool
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		count, err := checkedNativeSize(len(requests))
		if err != nil {
			return err
		}
		return withCStringArray(func() []string {
			ids := make([]string, len(requests))
			for i := range requests {
				ids[i] = requests[i].SatelliteID
			}
			return ids
		}(), func(ids **C.char, _ C.size_t) error {
			for _, item := range []struct {
				count int
				size  uintptr
				name  string
			}{{len(requests), unsafe.Sizeof(C.SidereonPredictRequest{}), "SP3 observable batch requests"}, {len(requests), unsafe.Sizeof(C.SidereonPredictedObservables{}), "SP3 observable batch outputs"}, {len(requests), unsafe.Sizeof(C.bool(false)), "SP3 observable batch status flags"}} {
				if _, err := checkedNativeAllocationSize(item.count, item.size); err != nil {
					return fmt.Errorf("sidereon: %s: %w", item.name, err)
				}
			}
			requestMemory, err := checkedNativeMalloc(len(requests), unsafe.Sizeof(C.SidereonPredictRequest{}))
			if err != nil {
				return err
			}
			defer C.free(requestMemory)
			requestValues := unsafe.Slice((*C.SidereonPredictRequest)(requestMemory), len(requests))
			idValues := unsafe.Slice(ids, len(requests))
			for i, request := range requests {
				requestValues[i].sat_id = idValues[i]
				for axis := range request.ReceiverECEF {
					requestValues[i].receiver_ecef_m[axis] = C.double(request.ReceiverECEF[axis])
				}
				requestValues[i].t_rx_j2000_s = C.double(request.TRxJ2000S)
			}
			values := make([]C.SidereonPredictedObservables, len(requests))
			okValues := make([]C.bool, len(requests))
			var requestPointer *C.SidereonPredictRequest
			var valuePointer *C.SidereonPredictedObservables
			var okPointer *C.bool
			if len(requests) != 0 {
				requestPointer, valuePointer, okPointer = &requestValues[0], &values[0], &okValues[0]
			}
			var optionsMemory unsafe.Pointer
			var optionsPointer *C.SidereonObservablesOptions
			if options != nil {
				optionsMemory, err = checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservablesOptions{}))
				if err != nil {
					return err
				}
				defer C.free(optionsMemory)
				optionsPointer = (*C.SidereonObservablesOptions)(optionsMemory)
				*optionsPointer = cObservablesOptions(*options)
			}
			if err := statusErrorLocked(uint32(C.sidereon_sp3_observables_batch((*C.SidereonSp3)(pointer), requestPointer, count, optionsPointer, valuePointer, okPointer))); err != nil {
				return err
			}
			result = make([]NativePredictedObservables, len(values))
			accepted = make([]bool, len(values))
			for i := range values {
				result[i], accepted[i] = predictedObservablesFromC(values[i]), bool(okValues[i])
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(requests)
	return result, accepted, err
}

func (s *SP3) PredictRanges(requests []NativePredictRequest, options *NativeObservablesOptions) ([]NativeRangePrediction, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []NativeRangePrediction
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		count, err := checkedNativeSize(len(requests))
		if err != nil {
			return err
		}
		ids := make([]string, len(requests))
		for i := range requests {
			ids[i] = requests[i].SatelliteID
		}
		return withCStringArray(ids, func(idPointers **C.char, _ C.size_t) error {
			for _, item := range []struct {
				count int
				size  uintptr
				name  string
			}{{len(requests), unsafe.Sizeof(C.SidereonRangePredictionRequest{}), "SP3 range requests"}, {len(requests), unsafe.Sizeof(C.SidereonRangePrediction{}), "SP3 range outputs"}} {
				if _, err := checkedNativeAllocationSize(item.count, item.size); err != nil {
					return fmt.Errorf("sidereon: %s: %w", item.name, err)
				}
			}
			requestMemory, err := checkedNativeMalloc(len(requests), unsafe.Sizeof(C.SidereonRangePredictionRequest{}))
			if err != nil {
				return err
			}
			defer C.free(requestMemory)
			requestValues := unsafe.Slice((*C.SidereonRangePredictionRequest)(requestMemory), len(requests))
			idValues := unsafe.Slice(idPointers, len(requests))
			for i, request := range requests {
				requestValues[i].sat_id = idValues[i]
				for axis := range request.ReceiverECEF {
					requestValues[i].receiver_ecef_m[axis] = C.double(request.ReceiverECEF[axis])
				}
				requestValues[i].t_rx_j2000_s = C.double(request.TRxJ2000S)
			}
			values := make([]C.SidereonRangePrediction, len(requests))
			var requestPointer *C.SidereonRangePredictionRequest
			var outPointer *C.SidereonRangePrediction
			if len(requests) != 0 {
				requestPointer, outPointer = &requestValues[0], &values[0]
			}
			var optionsMemory unsafe.Pointer
			var optionsPointer *C.SidereonObservablesOptions
			if options != nil {
				optionsMemory, err = checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservablesOptions{}))
				if err != nil {
					return err
				}
				defer C.free(optionsMemory)
				optionsPointer = (*C.SidereonObservablesOptions)(optionsMemory)
				*optionsPointer = cObservablesOptions(*options)
			}
			if err := statusErrorLocked(uint32(C.sidereon_sp3_predict_ranges((*C.SidereonSp3)(pointer), requestPointer, count, optionsPointer, outPointer))); err != nil {
				return err
			}
			result = make([]NativeRangePrediction, len(values))
			for i, value := range values {
				result[i].GeometricRangeM = float64(value.geometric_range_m)
				result[i].HasSatelliteClock, result[i].SatelliteClockS = bool(value.has_sat_clock_s), float64(value.sat_clock_s)
				result[i].TransmitTimeJ2000S = float64(value.transmit_time_j2000_s)
				for axis := range result[i].SatellitePositionECEF {
					result[i].SatellitePositionECEF[axis] = float64(value.sat_pos_ecef_m[axis])
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(requests)
	return result, err
}

type NativeVisibilityPass struct {
	Satellite     string
	RiseStepIndex int
	SetStepIndex  int
	PeakElevation float64
	PeakStepIndex int
}

type NativeVisibilitySeriesPoint struct {
	StepIndex int
	Visible   int
}

type NativeGeometryVisible struct {
	Satellite    string
	ElevationDeg float64
	AzimuthDeg   float64
}

func withSp3GeometryInputs(h *positioningHandle, receiver [3]float64, systems []uint32, fn func(unsafe.Pointer, *C.double, *C.uint32_t, C.size_t) error) error {
	if h == nil {
		return ErrClosed
	}
	for i, system := range systems {
		if system > uint32(C.SIDEREON_GNSS_SYSTEM_SBAS) {
			return invalidArgument(fmt.Sprintf("invalid SP3 geometry system at index %d", i))
		}
	}
	return h.with(func(pointer unsafe.Pointer) error {
		var arena sp3Arena
		defer arena.close()
		receiverMemory, err := arena.malloc(3, unsafe.Sizeof(C.double(0)))
		if err != nil {
			return err
		}
		receiverValues := unsafe.Slice((*C.double)(receiverMemory), 3)
		for i, value := range receiver {
			receiverValues[i] = C.double(value)
		}
		var systemsPointer *C.uint32_t
		if len(systems) != 0 {
			systemsMemory, allocationErr := arena.malloc(len(systems), unsafe.Sizeof(C.uint32_t(0)))
			if allocationErr != nil {
				return allocationErr
			}
			values := unsafe.Slice((*C.uint32_t)(systemsMemory), len(systems))
			for i, value := range systems {
				values[i] = C.uint32_t(value)
			}
			systemsPointer = (*C.uint32_t)(systemsMemory)
		}
		systemsCount, err := checkedNativeSize(len(systems))
		if err != nil {
			return err
		}
		var resultErr error
		withCThread(func() {
			resultErr = fn(pointer, (*C.double)(receiverMemory), systemsPointer, systemsCount)
		})
		runtime.KeepAlive(systems)
		return resultErr
	})
}

func (s *SP3) GeometryPasses(receiver [3]float64, windowStart, windowEnd float64, stepSeconds uint64, elevationMask float64, systems []uint32) ([]NativeVisibilityPass, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []NativeVisibilityPass
	err := withSp3GeometryInputs(s.handle, receiver, systems, func(pointer unsafe.Pointer, receiverPointer *C.double, systemsPointer *C.uint32_t, systemsCount C.size_t) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonVisibilityPass, count C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_geometry_passes((*C.SidereonSp3)(pointer), receiverPointer, C.double(windowStart), C.double(windowEnd), C.uint64_t(stepSeconds), C.double(elevationMask), systemsPointer, systemsCount, out, count, &written, &required)
		}
		if err := statusErrorLocked(uint32(invoke(nil, 0))); err != nil {
			return err
		}
		count, err := validateNativeQuery("SP3 geometry passes", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonVisibilityPass{})); err != nil {
			return err
		}
		values := make([]C.SidereonVisibilityPass, count)
		var out *C.SidereonVisibilityPass
		if count != 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := statusErrorLocked(uint32(invoke(out, C.size_t(count)))); err != nil {
			return err
		}
		writtenCount, err := validateNativeOutput("SP3 geometry passes", count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeVisibilityPass, writtenCount)
		for i := range result {
			rise, err := checkedNativeCount(uint64(values[i].rise_step_index))
			if err != nil {
				return err
			}
			set, err := checkedNativeCount(uint64(values[i].set_step_index))
			if err != nil {
				return err
			}
			peak, err := checkedNativeCount(uint64(values[i].peak_step_index))
			if err != nil {
				return err
			}
			result[i] = NativeVisibilityPass{Satellite: tokenFromC(values[i].satellite), RiseStepIndex: rise, SetStepIndex: set, PeakElevation: float64(values[i].peak_elevation_deg), PeakStepIndex: peak}
		}
		return nil
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *SP3) GeometryVisibilitySeries(receiver [3]float64, windowStart, windowEnd float64, stepSeconds uint64, elevationMask float64, systems []uint32) ([]NativeVisibilitySeriesPoint, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []NativeVisibilitySeriesPoint
	err := withSp3GeometryInputs(s.handle, receiver, systems, func(pointer unsafe.Pointer, receiverPointer *C.double, systemsPointer *C.uint32_t, systemsCount C.size_t) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonVisibilitySeriesPoint, count C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_geometry_visibility_series((*C.SidereonSp3)(pointer), receiverPointer, C.double(windowStart), C.double(windowEnd), C.uint64_t(stepSeconds), C.double(elevationMask), systemsPointer, systemsCount, out, count, &written, &required)
		}
		if err := statusErrorLocked(uint32(invoke(nil, 0))); err != nil {
			return err
		}
		count, err := validateNativeQuery("SP3 visibility series", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonVisibilitySeriesPoint{})); err != nil {
			return err
		}
		values := make([]C.SidereonVisibilitySeriesPoint, count)
		var out *C.SidereonVisibilitySeriesPoint
		if count != 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := statusErrorLocked(uint32(invoke(out, C.size_t(count)))); err != nil {
			return err
		}
		writtenCount, err := validateNativeOutput("SP3 visibility series", count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeVisibilitySeriesPoint, writtenCount)
		for i := range result {
			step, err := checkedNativeCount(uint64(values[i].step_index))
			if err != nil {
				return err
			}
			visible, err := checkedNativeCount(uint64(values[i].n_visible))
			if err != nil {
				return err
			}
			result[i] = NativeVisibilitySeriesPoint{StepIndex: step, Visible: visible}
		}
		return nil
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *SP3) GeometryVisible(receiver [3]float64, epoch, elevationMask float64, systems []uint32) ([]NativeGeometryVisible, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []NativeGeometryVisible
	err := withSp3GeometryInputs(s.handle, receiver, systems, func(pointer unsafe.Pointer, receiverPointer *C.double, systemsPointer *C.uint32_t, systemsCount C.size_t) error {
		var written, required C.size_t
		invoke := func(out *C.SidereonGeometryVisible, count C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_geometry_visible((*C.SidereonSp3)(pointer), receiverPointer, C.double(epoch), C.double(elevationMask), systemsPointer, systemsCount, out, count, &written, &required)
		}
		if err := statusErrorLocked(uint32(invoke(nil, 0))); err != nil {
			return err
		}
		count, err := validateNativeQuery("SP3 visible geometry", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonGeometryVisible{})); err != nil {
			return err
		}
		values := make([]C.SidereonGeometryVisible, count)
		var out *C.SidereonGeometryVisible
		if count != 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := statusErrorLocked(uint32(invoke(out, C.size_t(count)))); err != nil {
			return err
		}
		writtenCount, err := validateNativeOutput("SP3 visible geometry", count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeGeometryVisible, writtenCount)
		for i := range result {
			result[i] = NativeGeometryVisible{Satellite: tokenFromC(values[i].satellite), ElevationDeg: float64(values[i].elevation_deg), AzimuthDeg: float64(values[i].azimuth_deg)}
		}
		return nil
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *SP3) StencilExtent() (float64, float64, error) {
	if s == nil || s.handle == nil {
		return 0, 0, ErrClosed
	}
	var before, after C.double
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sp3_stencil_extent((*C.SidereonSp3)(pointer), &before, &after))
		})
	})
	runtime.KeepAlive(s)
	return float64(before), float64(after), err
}
