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

type ConstellationDiffCounts struct {
	Added, Removed, NORADReassigned, SP3IDChanged, SVNChanged, FDMAChannelChanged, ActivityChanged, UsabilityChanged int
}

type ConstellationPRN struct {
	System uint32
	PRN    uint16
}
type ConstellationBoolChange struct {
	System   uint32
	PRN      uint16
	From, To bool
}
type ConstellationOptionalU16Change struct {
	System      uint32
	PRN         uint16
	FromPresent bool
	From        uint16
	ToPresent   bool
	To          uint16
}
type ConstellationOptionalI8Change struct {
	System      uint32
	PRN         uint16
	FromPresent bool
	From        int8
	ToPresent   bool
	To          int8
}
type ConstellationU32Change struct {
	System   uint32
	PRN      uint16
	From, To uint32
}
type ConstellationStringChangeMeta struct {
	System         uint32
	PRN            uint16
	FromLen, ToLen int
}
type ConstellationStringChange struct {
	Meta     ConstellationStringChangeMeta
	From, To string
}

type Constellation struct {
	_      noCopy
	handle *positioningHandle
}
type ConstellationDiff struct {
	_      noCopy
	handle *positioningHandle
}
type ConstellationValidation struct {
	_      noCopy
	handle *positioningHandle
}

type NavcenAssessment struct {
	System                uint32
	PRN                   uint16
	SVNPresent            bool
	SVN                   uint16
	Usable                bool
	ActiveNANU            bool
	EvaluatedAtUnixUS     int64
	Timing                uint32
	EffectiveStartPresent bool
	EffectiveStartUnixUS  int64
	EffectiveEndPresent   bool
	EffectiveEndUnixUS    int64
}
type NavcenAssessments struct {
	_      noCopy
	handle *positioningHandle
}

const (
	ConstellationBoolStyleLowerValue = uint32(C.SIDEREON_CONSTELLATION_BOOL_STYLE_LOWER)
	ConstellationBoolStyleTitleValue = uint32(C.SIDEREON_CONSTELLATION_BOOL_STYLE_TITLE)
	NavcenTimingNotApplicableValue   = uint32(C.SIDEREON_NAVCEN_TIMING_NOT_APPLICABLE)
	NavcenTimingParsedValue          = uint32(C.SIDEREON_NAVCEN_TIMING_PARSED)
	NavcenTimingUnparseableValue     = uint32(C.SIDEREON_NAVCEN_TIMING_UNPARSEABLE)
)

func validateNavcenTiming(value uint32) error {
	if value > NavcenTimingUnparseableValue {
		return invalidArgument("native NAVCEN timing is not defined")
	}
	return nil
}

func releaseConstellation(p unsafe.Pointer) {
	C.sidereon_constellation_free((*C.SidereonConstellation)(p))
}
func releaseConstellationDiff(p unsafe.Pointer) {
	C.sidereon_constellation_diff_free((*C.SidereonConstellationDiff)(p))
}
func releaseConstellationValidation(p unsafe.Pointer) {
	C.sidereon_constellation_validation_free((*C.SidereonConstellationValidation)(p))
}
func releaseNavcen(p unsafe.Pointer) {
	C.sidereon_navcen_assessments_free((*C.SidereonNavcenAssessments)(p))
}

func allocNativeArray(count int, elementSize uintptr) (unsafe.Pointer, error) {
	if _, err := checkedNativeAllocationSize(count, elementSize); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	p := C.malloc(C.size_t(count) * C.size_t(elementSize))
	if p == nil {
		return nil, errors.New("sidereon: unable to allocate native array")
	}
	return p, nil
}

func constellationFromC(v C.SidereonConstellationRecord) (ConstellationRecord, error) {
	if err := validateGNSSSystemValue(uint32(v.system)); err != nil {
		return ConstellationRecord{}, err
	}
	return ConstellationRecord{System: uint32(v.system), PRN: uint16(v.prn), SVNPresent: bool(v.svn_present), SVN: uint16(v.svn), NORADID: uint32(v.norad_id), FDMAChannelPresent: bool(v.fdma_channel_present), FDMAChannel: int8(v.fdma_channel), Active: bool(v.active), Usable: bool(v.usable)}, nil
}
func constellationPRNFromC(v C.SidereonConstellationPrn) (ConstellationPRN, error) {
	if err := validateGNSSSystemValue(uint32(v.system)); err != nil {
		return ConstellationPRN{}, err
	}
	return ConstellationPRN{System: uint32(v.system), PRN: uint16(v.prn)}, nil
}
func boolChangeFromC(v C.SidereonConstellationBoolChange) (ConstellationBoolChange, error) {
	if err := validateGNSSSystemValue(uint32(v.system)); err != nil {
		return ConstellationBoolChange{}, err
	}
	return ConstellationBoolChange{System: uint32(v.system), PRN: uint16(v.prn), From: bool(v.from), To: bool(v.to)}, nil
}
func optionalU16FromC(v C.SidereonConstellationOptionalU16Change) (ConstellationOptionalU16Change, error) {
	if err := validateGNSSSystemValue(uint32(v.system)); err != nil {
		return ConstellationOptionalU16Change{}, err
	}
	return ConstellationOptionalU16Change{System: uint32(v.system), PRN: uint16(v.prn), FromPresent: bool(v.from_present), From: uint16(v.from), ToPresent: bool(v.to_present), To: uint16(v.to)}, nil
}
func optionalI8FromC(v C.SidereonConstellationOptionalI8Change) (ConstellationOptionalI8Change, error) {
	if err := validateGNSSSystemValue(uint32(v.system)); err != nil {
		return ConstellationOptionalI8Change{}, err
	}
	return ConstellationOptionalI8Change{System: uint32(v.system), PRN: uint16(v.prn), FromPresent: bool(v.from_present), From: int8(v.from), ToPresent: bool(v.to_present), To: int8(v.to)}, nil
}
func u32ChangeFromC(v C.SidereonConstellationU32Change) (ConstellationU32Change, error) {
	if err := validateGNSSSystemValue(uint32(v.system)); err != nil {
		return ConstellationU32Change{}, err
	}
	return ConstellationU32Change{System: uint32(v.system), PRN: uint16(v.prn), From: uint32(v.from), To: uint32(v.to)}, nil
}
func stringMetaFromC(v C.SidereonConstellationStringChangeMeta) (ConstellationStringChangeMeta, error) {
	a, e := sizeTToInt(v.from_len, "diff string from length")
	if e != nil {
		return ConstellationStringChangeMeta{}, e
	}
	b, e := sizeTToInt(v.to_len, "diff string to length")
	if e != nil {
		return ConstellationStringChangeMeta{}, e
	}
	if e := validateGNSSSystemValue(uint32(v.system)); e != nil {
		return ConstellationStringChangeMeta{}, e
	}
	return ConstellationStringChangeMeta{System: uint32(v.system), PRN: uint16(v.prn), FromLen: a, ToLen: b}, nil
}

func buildConstellation(system uint32, omm, navcen []byte, evaluatedAt *int64) (*Constellation, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return nil, err
	}
	var out *C.SidereonConstellation
	var err error
	withCThread(func() {
		var ommP, navP unsafe.Pointer
		ommP, err = copyNativeInput(omm)
		if err != nil {
			return
		}
		defer freeNativeInput(ommP)
		navP, err = copyNativeInput(navcen)
		if err != nil {
			return
		}
		defer freeNativeInput(navP)
		var status uint32
		if evaluatedAt == nil {
			status = uint32(C.sidereon_constellation_build(C.uint32_t(system), (*C.uint8_t)(ommP), C.size_t(len(omm)), (*C.uint8_t)(navP), C.size_t(len(navcen)), &out))
		} else {
			status = uint32(C.sidereon_constellation_build_at(C.uint32_t(system), (*C.uint8_t)(ommP), C.size_t(len(omm)), (*C.uint8_t)(navP), C.size_t(len(navcen)), C.int64_t(*evaluatedAt), &out))
		}
		err = statusErrorLocked(status)
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_constellation_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, errors.New("sidereon: native constellation constructor returned no handle")
	}
	return &Constellation{handle: newPositioningHandle(unsafe.Pointer(out), releaseConstellation)}, nil
}
func BuildConstellation(system uint32, omm, navcen []byte) (*Constellation, error) {
	return buildConstellation(system, omm, navcen, nil)
}
func BuildConstellationAt(system uint32, omm, navcen []byte, evaluatedAtUnixUS int64) (*Constellation, error) {
	return buildConstellation(system, omm, navcen, &evaluatedAtUnixUS)
}
func (c *Constellation) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return c.handle.close()
}
func (c *Constellation) RecordCount() (int, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	e := c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_constellation_record_count((*C.SidereonConstellation)(p), &n)) })
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(n, "constellation record count")
}
func (c *Constellation) Record(i int) (ConstellationRecord, error) {
	if c == nil || c.handle == nil {
		return ConstellationRecord{}, ErrClosed
	}
	if i < 0 {
		return ConstellationRecord{}, invalidArgument("negative constellation record index")
	}
	var v C.SidereonConstellationRecord
	e := c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_constellation_record((*C.SidereonConstellation)(p), C.size_t(i), &v))
		})
	})
	runtime.KeepAlive(c)
	if e != nil {
		return ConstellationRecord{}, e
	}
	return constellationFromC(v)
}
func (c *Constellation) Records() ([]ConstellationRecord, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	n, e := c.RecordCount()
	if e != nil {
		return nil, e
	}
	out := make([]ConstellationRecord, n)
	for i := range out {
		out[i], e = c.Record(i)
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
func (c *Constellation) GNSSSP3ID(system uint32, prn uint16) (string, error) {
	return constellationGNSSSP3ID(system, prn)
}
func GNSSSP3ID(system uint32, prn uint16) (string, error) { return constellationGNSSSP3ID(system, prn) }
func constellationGNSSSP3ID(system uint32, prn uint16) (string, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return "", err
	}
	var result []byte
	var e error
	withCThread(func() {
		result, e = copyNativeBytesLocked("constellation GNSS SP3 id", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_constellation_gnss_sp3_id(C.uint32_t(system), C.uint16_t(prn), out, n, w, r)
		})
	})
	return string(result), e
}
func (c *Constellation) ToCSV(boolStyle uint32) ([]byte, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	if boolStyle != ConstellationBoolStyleLowerValue && boolStyle != ConstellationBoolStyleTitleValue {
		return nil, invalidArgument("constellation CSV boolean style is not defined")
	}
	var out []byte
	var e error
	e = c.handle.with(func(p unsafe.Pointer) error {
		withCThread(func() {
			out, e = copyNativeBytesLocked("constellation CSV", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_constellation_to_csv((*C.SidereonConstellation)(p), C.uint32_t(boolStyle), b, n, w, r)
			})
		})
		return e
	})
	return out, e
}
func (c *Constellation) Validate() (*ConstellationValidation, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonConstellationValidation
	e := c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_constellation_validate((*C.SidereonConstellation)(p), &out)) })
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_constellation_validation_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native validation returned no handle")
	}
	return &ConstellationValidation{handle: newPositioningHandle(unsafe.Pointer(out), releaseConstellationValidation)}, nil
}
func (c *Constellation) ValidateAgainstSP3(sp3 *SP3) (*ConstellationValidation, error) {
	if c == nil || c.handle == nil || sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonConstellationValidation
	e := c.handle.with(func(p unsafe.Pointer) error {
		return sp3.handle.with(func(q unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_constellation_validate_against_sp3((*C.SidereonConstellation)(p), (*C.SidereonSp3)(q), &out))
			})
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_constellation_validation_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native validation returned no handle")
	}
	return &ConstellationValidation{handle: newPositioningHandle(unsafe.Pointer(out), releaseConstellationValidation)}, nil
}
func (c *Constellation) ValidateAgainstSP3IDs(ids []string) (*ConstellationValidation, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	ptrs, release, e := cStrings(ids, "SP3 id")
	if e != nil {
		return nil, e
	}
	defer release()
	var out *C.SidereonConstellationValidation
	e = c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			var idPointer **C.char
			if len(ptrs) != 0 {
				idPointer = &ptrs[0]
			}
			return uint32(C.sidereon_constellation_validate_against_sp3_ids((*C.SidereonConstellation)(p), idPointer, C.size_t(len(ids)), &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_constellation_validation_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native validation returned no handle")
	}
	return &ConstellationValidation{handle: newPositioningHandle(unsafe.Pointer(out), releaseConstellationValidation)}, nil
}
func (c *Constellation) ValidateAgainstSP3IDsStrict(ids []string) error {
	if c == nil || c.handle == nil {
		return ErrClosed
	}
	ptrs, release, err := cStrings(ids, "SP3 id")
	if err != nil {
		return err
	}
	defer release()
	return c.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			var idPointer **C.char
			if len(ptrs) != 0 {
				idPointer = &ptrs[0]
			}
			return uint32(C.sidereon_constellation_validate_against_sp3_ids_strict(
				(*C.SidereonConstellation)(p), idPointer, C.size_t(len(ids))))
		})
	})
}

func (v *ConstellationValidation) Close() error {
	if v == nil || v.handle == nil {
		return nil
	}
	return v.handle.close()
}
func (v *ConstellationValidation) IsValid() (bool, error) {
	if v == nil || v.handle == nil {
		return false, ErrClosed
	}
	var x C.bool
	e := v.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_constellation_validation_is_valid((*C.SidereonConstellationValidation)(p), &x))
		})
	})
	return bool(x), e
}
func copyValidationPRNs(v *ConstellationValidation, which int) ([]ConstellationPRN, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	var out []C.SidereonConstellationPrn
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := v.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			var status C.enum_SidereonStatus
			switch which {
			case 0:
				status = C.sidereon_constellation_validation_inactive_unusable_prns((*C.SidereonConstellationValidation)(p), nil, 0, &w, &r)
			case 1:
				status = C.sidereon_constellation_validation_duplicate_prns((*C.SidereonConstellationValidation)(p), nil, 0, &w, &r)
			default:
				return invalidArgument("validation PRN selector")
			}
			if x := statusErrorLocked(uint32(status)); x != nil {
				return x
			}
			n, x := validateNativeQuery("validation PRNs", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.SidereonConstellationPrn{}))
			if x != nil {
				return x
			}
			out = unsafe.Slice((*C.SidereonConstellationPrn)(memory), n)
			w, r = 0, 0
			var q *C.SidereonConstellationPrn
			if n > 0 {
				q = &out[0]
			}
			if which == 0 {
				status = C.sidereon_constellation_validation_inactive_unusable_prns((*C.SidereonConstellationValidation)(p), q, C.size_t(n), &w, &r)
			} else {
				status = C.sidereon_constellation_validation_duplicate_prns((*C.SidereonConstellationValidation)(p), q, C.size_t(n), &w, &r)
			}
			if x := statusErrorLocked(uint32(status)); x != nil {
				return x
			}
			n, x = validateTwoPassCounts("validation PRNs", n, n, uint64(w), uint64(r))
			if x != nil {
				return x
			}
			return nil
		})
	})
	if e != nil {
		return nil, e
	}
	result := make([]ConstellationPRN, len(out))
	for i := range out {
		value, err := constellationPRNFromC(out[i])
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}
func (v *ConstellationValidation) InactiveUnusablePRNs() ([]ConstellationPRN, error) {
	return copyValidationPRNs(v, 0)
}
func (v *ConstellationValidation) DuplicatePRNs() ([]ConstellationPRN, error) {
	return copyValidationPRNs(v, 1)
}
func (v *ConstellationValidation) DuplicateNORADIDs() ([]uint32, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.uint32_t
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	err := v.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := C.sidereon_constellation_validation_duplicate_norad_ids((*C.SidereonConstellationValidation)(p), nil, 0, &written, &required)
			if e := statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			n, e := validateNativeQuery("validation duplicate NORAD ids", uint64(written), uint64(required))
			if e != nil {
				return e
			}
			memory, e = allocNativeArray(n, unsafe.Sizeof(C.uint32_t(0)))
			if e != nil {
				return e
			}
			raw = unsafe.Slice((*C.uint32_t)(memory), n)
			written, required = 0, 0
			var out *C.uint32_t
			if n != 0 {
				out = &raw[0]
			}
			status = C.sidereon_constellation_validation_duplicate_norad_ids((*C.SidereonConstellationValidation)(p), out, C.size_t(n), &written, &required)
			if e = statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			_, e = validateTwoPassCounts("validation duplicate NORAD ids", n, n, uint64(written), uint64(required))
			return e
		})
	})
	if err != nil {
		return nil, err
	}
	out := make([]uint32, len(raw))
	for i := range raw {
		out[i] = uint32(raw[i])
	}
	return out, nil
}
func copyValidationU32(v *ConstellationValidation, missing bool) ([]string, error) {
	if v == nil || v.handle == nil {
		return nil, ErrClosed
	}
	var out []C.SidereonSatelliteToken
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := v.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			var s C.enum_SidereonStatus
			if missing {
				s = C.sidereon_constellation_validation_missing_sp3_ids((*C.SidereonConstellationValidation)(p), nil, 0, &w, &r)
			} else {
				s = C.sidereon_constellation_validation_extra_sp3_ids((*C.SidereonConstellationValidation)(p), nil, 0, &w, &r)
			}
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("validation ids", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.SidereonSatelliteToken{}))
			if x != nil {
				return x
			}
			out = unsafe.Slice((*C.SidereonSatelliteToken)(memory), n)
			w, r = 0, 0
			var q *C.SidereonSatelliteToken
			if n > 0 {
				q = &out[0]
			}
			if missing {
				s = C.sidereon_constellation_validation_missing_sp3_ids((*C.SidereonConstellationValidation)(p), q, C.size_t(n), &w, &r)
			} else {
				s = C.sidereon_constellation_validation_extra_sp3_ids((*C.SidereonConstellationValidation)(p), q, C.size_t(n), &w, &r)
			}
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("validation ids", n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	result := make([]string, len(out))
	for i := range out {
		result[i] = fixedCString((*C.char)(unsafe.Pointer(&out[i].bytes[0])))
	}
	return result, nil
}
func (v *ConstellationValidation) MissingSP3IDs() ([]string, error) {
	return copyValidationU32(v, true)
}
func (v *ConstellationValidation) ExtraSP3IDs() ([]string, error) { return copyValidationU32(v, false) }

func (c *Constellation) Diff(other *Constellation) (*ConstellationDiff, error) {
	if c == nil || c.handle == nil || other == nil || other.handle == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonConstellationDiff
	e := withPositioningPair(c.handle, other.handle, func(p, q unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_constellation_diff((*C.SidereonConstellation)(p), (*C.SidereonConstellation)(q), &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_constellation_diff_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native diff returned no handle")
	}
	return &ConstellationDiff{handle: newPositioningHandle(unsafe.Pointer(out), releaseConstellationDiff)}, nil
}
func (d *ConstellationDiff) Close() error {
	if d == nil || d.handle == nil {
		return nil
	}
	return d.handle.close()
}
func (d *ConstellationDiff) Changed() (bool, error) {
	if d == nil || d.handle == nil {
		return false, ErrClosed
	}
	var x C.bool
	e := d.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_constellation_diff_changed((*C.SidereonConstellationDiff)(p), &x))
		})
	})
	return bool(x), e
}
func (d *ConstellationDiff) Counts() (ConstellationDiffCounts, error) {
	if d == nil || d.handle == nil {
		return ConstellationDiffCounts{}, ErrClosed
	}
	var x C.SidereonConstellationDiffCounts
	e := d.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_constellation_diff_counts((*C.SidereonConstellationDiff)(p), &x))
		})
	})
	if e != nil {
		return ConstellationDiffCounts{}, e
	}
	values := []*C.size_t{&x.added, &x.removed, &x.norad_reassigned, &x.sp3_id_changed, &x.svn_changed, &x.fdma_channel_changed, &x.activity_changed, &x.usability_changed}
	converted := make([]int, len(values))
	for i, value := range values {
		converted[i], e = sizeTToInt(*value, "constellation diff count")
		if e != nil {
			return ConstellationDiffCounts{}, e
		}
	}
	return ConstellationDiffCounts{Added: converted[0], Removed: converted[1], NORADReassigned: converted[2], SP3IDChanged: converted[3], SVNChanged: converted[4], FDMAChannelChanged: converted[5], ActivityChanged: converted[6], UsabilityChanged: converted[7]}, nil
}

func copyDiffRecords(d *ConstellationDiff, added bool) ([]ConstellationRecord, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonConstellationRecord
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := d.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			var s C.enum_SidereonStatus
			if added {
				s = C.sidereon_constellation_diff_added((*C.SidereonConstellationDiff)(p), nil, 0, &w, &r)
			} else {
				s = C.sidereon_constellation_diff_removed((*C.SidereonConstellationDiff)(p), nil, 0, &w, &r)
			}
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("diff records", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.SidereonConstellationRecord{}))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.SidereonConstellationRecord)(memory), n)
			w, r = 0, 0
			var q *C.SidereonConstellationRecord
			if n > 0 {
				q = &raw[0]
			}
			if added {
				s = C.sidereon_constellation_diff_added((*C.SidereonConstellationDiff)(p), q, C.size_t(n), &w, &r)
			} else {
				s = C.sidereon_constellation_diff_removed((*C.SidereonConstellationDiff)(p), q, C.size_t(n), &w, &r)
			}
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("diff records", n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]ConstellationRecord, len(raw))
	for i := range raw {
		value, err := constellationFromC(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = value
	}
	return out, nil
}
func (d *ConstellationDiff) Added() ([]ConstellationRecord, error) { return copyDiffRecords(d, true) }
func (d *ConstellationDiff) Removed() ([]ConstellationRecord, error) {
	return copyDiffRecords(d, false)
}

type constellationBoolDiffCall func(*C.SidereonConstellationDiff, *C.SidereonConstellationBoolChange, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func copyDiffBoolChanges(d *ConstellationDiff, call constellationBoolDiffCall, label string) ([]ConstellationBoolChange, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonConstellationBoolChange
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	err := d.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := call((*C.SidereonConstellationDiff)(p), nil, 0, &written, &required)
			if e := statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			n, e := validateNativeQuery(label, uint64(written), uint64(required))
			if e != nil {
				return e
			}
			memory, e = allocNativeArray(n, unsafe.Sizeof(C.SidereonConstellationBoolChange{}))
			if e != nil {
				return e
			}
			raw = unsafe.Slice((*C.SidereonConstellationBoolChange)(memory), n)
			written, required = 0, 0
			var out *C.SidereonConstellationBoolChange
			if n != 0 {
				out = &raw[0]
			}
			status = call((*C.SidereonConstellationDiff)(p), out, C.size_t(n), &written, &required)
			if e = statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			_, e = validateTwoPassCounts(label, n, n, uint64(written), uint64(required))
			return e
		})
	})
	if err != nil {
		return nil, err
	}
	out := make([]ConstellationBoolChange, len(raw))
	for i := range raw {
		value, err := boolChangeFromC(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = value
	}
	return out, nil
}

func (d *ConstellationDiff) ActivityChanged() ([]ConstellationBoolChange, error) {
	return copyDiffBoolChanges(d, func(diff *C.SidereonConstellationDiff, out *C.SidereonConstellationBoolChange, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_constellation_diff_activity_changed(diff, out, len, written, required)
	}, "diff activity changes")
}
func (d *ConstellationDiff) UsabilityChanged() ([]ConstellationBoolChange, error) {
	return copyDiffBoolChanges(d, func(diff *C.SidereonConstellationDiff, out *C.SidereonConstellationBoolChange, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_constellation_diff_usability_changed(diff, out, len, written, required)
	}, "diff usability changes")
}

type constellationOptionalI8DiffCall func(*C.SidereonConstellationDiff, *C.SidereonConstellationOptionalI8Change, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func (d *ConstellationDiff) FDMAChannelChanged() ([]ConstellationOptionalI8Change, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonConstellationOptionalI8Change
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	err := d.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			call := constellationOptionalI8DiffCall(func(diff *C.SidereonConstellationDiff, out *C.SidereonConstellationOptionalI8Change, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_constellation_diff_fdma_channel_changed(diff, out, len, written, required)
			})
			var written, required C.size_t
			status := call((*C.SidereonConstellationDiff)(p), nil, 0, &written, &required)
			if e := statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			n, e := validateNativeQuery("diff FDMA channel changes", uint64(written), uint64(required))
			if e != nil {
				return e
			}
			memory, e = allocNativeArray(n, unsafe.Sizeof(C.SidereonConstellationOptionalI8Change{}))
			if e != nil {
				return e
			}
			raw = unsafe.Slice((*C.SidereonConstellationOptionalI8Change)(memory), n)
			written, required = 0, 0
			var out *C.SidereonConstellationOptionalI8Change
			if n != 0 {
				out = &raw[0]
			}
			status = call((*C.SidereonConstellationDiff)(p), out, C.size_t(n), &written, &required)
			if e = statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			_, e = validateTwoPassCounts("diff FDMA channel changes", n, n, uint64(written), uint64(required))
			return e
		})
	})
	if err != nil {
		return nil, err
	}
	out := make([]ConstellationOptionalI8Change, len(raw))
	for i := range raw {
		value, err := optionalI8FromC(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = value
	}
	return out, nil
}

type constellationOptionalU16DiffCall func(*C.SidereonConstellationDiff, *C.SidereonConstellationOptionalU16Change, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus

func (d *ConstellationDiff) SVNChanged() ([]ConstellationOptionalU16Change, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonConstellationOptionalU16Change
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	err := d.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			call := constellationOptionalU16DiffCall(func(diff *C.SidereonConstellationDiff, out *C.SidereonConstellationOptionalU16Change, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_constellation_diff_svn_changed(diff, out, len, written, required)
			})
			var written, required C.size_t
			status := call((*C.SidereonConstellationDiff)(p), nil, 0, &written, &required)
			if e := statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			n, e := validateNativeQuery("diff SVN changes", uint64(written), uint64(required))
			if e != nil {
				return e
			}
			memory, e = allocNativeArray(n, unsafe.Sizeof(C.SidereonConstellationOptionalU16Change{}))
			if e != nil {
				return e
			}
			raw = unsafe.Slice((*C.SidereonConstellationOptionalU16Change)(memory), n)
			written, required = 0, 0
			var out *C.SidereonConstellationOptionalU16Change
			if n != 0 {
				out = &raw[0]
			}
			status = call((*C.SidereonConstellationDiff)(p), out, C.size_t(n), &written, &required)
			if e = statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			_, e = validateTwoPassCounts("diff SVN changes", n, n, uint64(written), uint64(required))
			return e
		})
	})
	if err != nil {
		return nil, err
	}
	out := make([]ConstellationOptionalU16Change, len(raw))
	for i := range raw {
		value, err := optionalU16FromC(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = value
	}
	return out, nil
}

func (d *ConstellationDiff) NORADReassigned() ([]ConstellationU32Change, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonConstellationU32Change
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	err := d.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := C.sidereon_constellation_diff_norad_reassigned((*C.SidereonConstellationDiff)(p), nil, 0, &written, &required)
			if e := statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			n, e := validateNativeQuery("diff NORAD changes", uint64(written), uint64(required))
			if e != nil {
				return e
			}
			memory, e = allocNativeArray(n, unsafe.Sizeof(C.SidereonConstellationU32Change{}))
			if e != nil {
				return e
			}
			raw = unsafe.Slice((*C.SidereonConstellationU32Change)(memory), n)
			written, required = 0, 0
			var out *C.SidereonConstellationU32Change
			if n != 0 {
				out = &raw[0]
			}
			status = C.sidereon_constellation_diff_norad_reassigned((*C.SidereonConstellationDiff)(p), out, C.size_t(n), &written, &required)
			if e = statusErrorLocked(uint32(status)); e != nil {
				return e
			}
			_, e = validateTwoPassCounts("diff NORAD changes", n, n, uint64(written), uint64(required))
			return e
		})
	})
	if err != nil {
		return nil, err
	}
	out := make([]ConstellationU32Change, len(raw))
	for i := range raw {
		value, err := u32ChangeFromC(raw[i])
		if err != nil {
			return nil, err
		}
		out[i] = value
	}
	return out, nil
}

func (d *ConstellationDiff) SP3IDChanged() ([]ConstellationStringChange, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	counts, err := d.Counts()
	if err != nil {
		return nil, err
	}
	if counts.SP3IDChanged < 0 {
		return nil, errors.New("sidereon: negative SP3 id change count")
	}
	out := make([]ConstellationStringChange, counts.SP3IDChanged)
	for index := range out {
		var meta C.SidereonConstellationStringChangeMeta
		var from, to []byte
		err = d.handle.with(func(p unsafe.Pointer) error {
			return withCThreadError(func() error {
				status := C.sidereon_constellation_diff_sp3_id_changed_meta((*C.SidereonConstellationDiff)(p), C.size_t(index), &meta)
				if err := statusErrorLocked(uint32(status)); err != nil {
					return err
				}
				var err error
				from, err = copyNativeBytesLocked("constellation SP3 id change", func(b *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_constellation_diff_sp3_id_changed_from((*C.SidereonConstellationDiff)(p), C.size_t(index), b, n, written, required)
				})
				if err != nil {
					return err
				}
				to, err = copyNativeBytesLocked("constellation SP3 id change", func(b *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_constellation_diff_sp3_id_changed_to((*C.SidereonConstellationDiff)(p), C.size_t(index), b, n, written, required)
				})
				return err
			})
		})
		if err != nil {
			return nil, err
		}
		converted, err := stringMetaFromC(meta)
		if err != nil {
			return nil, err
		}
		fromLen, err := sizeTToInt(meta.from_len, "SP3 id change from length")
		if err != nil {
			return nil, err
		}
		toLen, err := sizeTToInt(meta.to_len, "SP3 id change to length")
		if err != nil {
			return nil, err
		}
		if len(from) != fromLen || len(to) != toLen {
			return nil, errors.New("sidereon: native SP3 id change length changed")
		}
		out[index] = ConstellationStringChange{Meta: converted, From: string(from), To: string(to)}
	}
	return out, nil
}

func galileanOptional(system uint32, arg uint16) (bool, uint16, error) {
	var p C.bool
	var out C.uint16_t
	var e error
	withCThread(func() {
		e = callStatus(func() uint32 { return uint32(C.sidereon_constellation_galileo_prn_for_gsat(C.uint16_t(arg), &p, &out)) })
	})
	_ = system
	return bool(p), uint16(out), e
}
func GalileoPRNForGSAT(gsat uint16) (bool, uint16, error) { return galileanOptional(0, gsat) }
func GLONASSSlotForNumber(number uint16) (bool, uint16, error) {
	var p C.bool
	var out C.uint16_t
	e := callStatus(func() uint32 {
		return uint32(C.sidereon_constellation_glonass_slot_for_number(C.uint16_t(number), &p, &out))
	})
	return bool(p), uint16(out), e
}
func GLONASSFDMAChannel(slot uint16) (bool, int8, error) {
	var p C.bool
	var out C.int8_t
	e := callStatus(func() uint32 {
		return uint32(C.sidereon_constellation_glonass_fdma_channel(C.uint16_t(slot), &p, &out))
	})
	return bool(p), int8(out), e
}

func ParseNAVCENAt(html []byte, evaluatedAtUnixUS int64) (*NavcenAssessments, error) {
	var out *C.SidereonNavcenAssessments
	var e error
	withCThread(func() {
		p, copyErr := copyNativeInput(html)
		if copyErr != nil {
			e = copyErr
			return
		}
		defer freeNativeInput(p)
		e = callStatus(func() uint32 {
			return uint32(C.sidereon_navcen_parse_at((*C.uint8_t)(p), C.size_t(len(html)), C.int64_t(evaluatedAtUnixUS), &out))
		})
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_navcen_assessments_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native NAVCEN parser returned no handle")
	}
	return &NavcenAssessments{handle: newPositioningHandle(unsafe.Pointer(out), releaseNavcen)}, nil
}
func (n *NavcenAssessments) Close() error {
	if n == nil || n.handle == nil {
		return nil
	}
	return n.handle.close()
}
func (n *NavcenAssessments) Count() (int, error) {
	if n == nil || n.handle == nil {
		return 0, ErrClosed
	}
	var x C.size_t
	e := n.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_navcen_assessment_count((*C.SidereonNavcenAssessments)(p), &x))
		})
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(x, "NAVCEN assessment count")
}
func (n *NavcenAssessments) Assessment(i int) (NavcenAssessment, error) {
	if n == nil || n.handle == nil {
		return NavcenAssessment{}, ErrClosed
	}
	if i < 0 {
		return NavcenAssessment{}, invalidArgument("negative NAVCEN assessment index")
	}
	var x C.SidereonNavcenAssessment
	e := n.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_navcen_assessment((*C.SidereonNavcenAssessments)(p), C.size_t(i), &x))
		})
	})
	if e != nil {
		return NavcenAssessment{}, e
	}
	if err := validateGNSSSystemValue(uint32(x.system)); err != nil {
		return NavcenAssessment{}, err
	}
	if err := validateNavcenTiming(uint32(x.timing)); err != nil {
		return NavcenAssessment{}, err
	}
	return NavcenAssessment{System: uint32(x.system), PRN: uint16(x.prn), SVNPresent: bool(x.svn_present), SVN: uint16(x.svn), Usable: bool(x.usable), ActiveNANU: bool(x.active_nanu), EvaluatedAtUnixUS: int64(x.evaluated_at_unix_us), Timing: uint32(x.timing), EffectiveStartPresent: bool(x.effective_start_present), EffectiveStartUnixUS: int64(x.effective_start_unix_us), EffectiveEndPresent: bool(x.effective_end_present), EffectiveEndUnixUS: int64(x.effective_end_unix_us)}, e
}
func (n *NavcenAssessments) text(i int, kind uint32) (string, error) {
	if n == nil || n.handle == nil {
		return "", ErrClosed
	}
	var out []byte
	var e error
	e = n.handle.with(func(p unsafe.Pointer) error {
		withCThread(func() {
			out, e = copyNativeBytesLocked("NAVCEN text", func(b *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				switch kind {
				case 0:
					return C.sidereon_navcen_assessment_nanu_type((*C.SidereonNavcenAssessments)(p), C.size_t(i), b, l, w, r)
				case 1:
					return C.sidereon_navcen_assessment_nanu_subject((*C.SidereonNavcenAssessments)(p), C.size_t(i), b, l, w, r)
				default:
					return C.sidereon_navcen_assessment_outage_start((*C.SidereonNavcenAssessments)(p), C.size_t(i), b, l, w, r)
				}
			})
		})
		return e
	})
	return string(out), e
}
func (n *NavcenAssessments) NANUType(i int) (string, error) {
	if i < 0 {
		return "", invalidArgument("negative NAVCEN assessment index")
	}
	return n.text(i, 0)
}
func (n *NavcenAssessments) NANUSubject(i int) (string, error) {
	if i < 0 {
		return "", invalidArgument("negative NAVCEN assessment index")
	}
	return n.text(i, 1)
}
func (n *NavcenAssessments) OutageStart(i int) (string, error) {
	if i < 0 {
		return "", invalidArgument("negative NAVCEN assessment index")
	}
	return n.text(i, 2)
}
