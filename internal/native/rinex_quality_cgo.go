//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type NativeRINEXLintSummary struct {
	FindingCount, FatalCount, ErrorCount, WarningCount, InfoCount int
	IsClean, DecodedFromCRINEX                                    bool
}
type NativeRINEXLintFinding struct {
	Code          string
	Severity      uint32
	Repairable    bool
	HasEpochIndex bool
	EpochIndex    int
	HasSatellite  bool
	Satellite     string
	HasField      bool
	Field         string
}
type RinexLintReport struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}
type NativeRINEXRepairOptions struct {
	HasFileStamp                                                                                        bool
	Program, RunBy, Date                                                                                string
	SetInterval, SetTimeOfLastObs, SetObservationCounts, DropEmptyRecords, SortRecords, DropUnsupported bool
}
type NativeRINEXRepairAction struct{ ID, Message string }
type RinexRepair struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func validateRINEXLintSeverityValue(value uint32) error {
	if value > 3 {
		return invalidArgument("invalid RINEX lint severity returned by native code")
	}
	return nil
}

func newRinexLintReport(p *C.SidereonRinexLintReport) (*RinexLintReport, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &RinexLintReport{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_rinex_lint_report_free((*C.SidereonRinexLintReport)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func newRinexRepair(p *C.SidereonRinexRepair) (*RinexRepair, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &RinexRepair{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_rinex_repair_free((*C.SidereonRinexRepair)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func ParseRINEXLint(data []byte, observation bool) (*RinexLintReport, error) {
	var p *C.SidereonRinexLintReport
	err := withInput(data, func(b *C.uint8_t, n C.size_t) uint32 {
		if observation {
			return C.sidereon_rinex_lint_obs(b, n, &p)
		}
		return C.sidereon_rinex_lint_nav(b, n, &p)
	})
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_rinex_lint_report_free(p) })
		}
		return nil, err
	}
	handle, err := newRinexLintReport(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_rinex_lint_report_free(p) })
	}
	return handle, err
}
func (r *RinexLintReport) Close() error {
	if r == nil {
		return nil
	}
	return closeProtocolResource(r, r.resource, &r.cleanup)
}
func (r *RinexLintReport) Summary() (NativeRINEXLintSummary, error) {
	var v C.SidereonRinexLintSummary
	err := r.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_rinex_lint_summary((*C.SidereonRinexLintReport)(p), &v) })
	})
	if err != nil {
		return NativeRINEXLintSummary{}, err
	}
	o := NativeRINEXLintSummary{IsClean: bool(v.is_clean), DecodedFromCRINEX: bool(v.decoded_from_crinex)}
	counts := []*int{&o.FindingCount, &o.FatalCount, &o.ErrorCount, &o.WarningCount, &o.InfoCount}
	nativeCounts := []C.size_t{v.finding_count, v.fatal_count, v.error_count, v.warning_count, v.info_count}
	for i := range counts {
		*counts[i], err = checkedNativeCount(uint64(nativeCounts[i]))
		if err != nil {
			return NativeRINEXLintSummary{}, err
		}
	}
	return o, nil
}
func (r *RinexLintReport) Findings() ([]NativeRINEXLintFinding, error) {
	var out []NativeRINEXLintFinding
	err := r.resource.with(func(p unsafe.Pointer) error {
		var w, req C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_rinex_lint_findings((*C.SidereonRinexLintReport)(p), nil, 0, &w, &req)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("RINEX lint findings", uint64(w), uint64(req))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRinexLintFinding{})); e != nil {
			return e
		}
		v := make([]C.SidereonRinexLintFinding, n)
		var q *C.SidereonRinexLintFinding
		if n > 0 {
			q = &v[0]
		}
		if e := callStatus(func() uint32 {
			return C.sidereon_rinex_lint_findings((*C.SidereonRinexLintReport)(p), q, C.size_t(n), &w, &req)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("RINEX lint findings", n, uint64(w), uint64(req))
		if e != nil {
			return e
		}
		out = make([]NativeRINEXLintFinding, z)
		for i := range out {
			if e := validateRINEXLintSeverityValue(uint32(v[i].severity)); e != nil {
				return e
			}
			out[i] = NativeRINEXLintFinding{Code: observationFixedString(v[i].code[:]), Severity: uint32(v[i].severity), Repairable: bool(v[i].repairable), HasEpochIndex: bool(v[i].has_epoch_index), HasSatellite: bool(v[i].has_satellite), Satellite: tokenFromC(v[i].satellite), HasField: bool(v[i].has_field), Field: observationFixedString(v[i].field[:])}
			out[i].EpochIndex, e = checkedNativeCount(uint64(v[i].epoch_index))
			if e != nil {
				return e
			}
		}
		return nil
	})
	runtime.KeepAlive(r)
	return out, err
}

func cRepairOptions(v NativeRINEXRepairOptions) (C.SidereonRinexRepairOptions, error) {
	var o C.SidereonRinexRepairOptions
	o.has_file_stamp = C.bool(v.HasFileStamp)
	o.set_interval = C.bool(v.SetInterval)
	o.set_time_of_last_obs = C.bool(v.SetTimeOfLastObs)
	o.set_obs_counts = C.bool(v.SetObservationCounts)
	o.drop_empty_records = C.bool(v.DropEmptyRecords)
	o.sort_records = C.bool(v.SortRecords)
	o.drop_unsupported = C.bool(v.DropUnsupported)
	fields := []struct {
		dst         []C.char
		value, name string
	}{{o.file_stamp_program[:], v.Program, "file-stamp program"}, {o.file_stamp_run_by[:], v.RunBy, "file-stamp run-by"}, {o.file_stamp_date[:], v.Date, "file-stamp date"}}
	for _, f := range fields {
		if err := rejectEmbeddedNUL(f.value, f.name); err != nil {
			return o, err
		}
		if len(f.value) >= len(f.dst) {
			return o, errors.New("sidereon: " + f.name + " is too long")
		}
		for i, b := range []byte(f.value) {
			f.dst[i] = C.char(b)
		}
	}
	return o, nil
}
func NewRINEXRepairOptions() (NativeRINEXRepairOptions, error) {
	var v C.SidereonRinexRepairOptions
	err := callStatus(func() uint32 { return C.sidereon_rinex_repair_options_init(&v) })
	return NativeRINEXRepairOptions{HasFileStamp: bool(v.has_file_stamp), Program: observationFixedString(v.file_stamp_program[:]), RunBy: observationFixedString(v.file_stamp_run_by[:]), Date: observationFixedString(v.file_stamp_date[:]), SetInterval: bool(v.set_interval), SetTimeOfLastObs: bool(v.set_time_of_last_obs), SetObservationCounts: bool(v.set_obs_counts), DropEmptyRecords: bool(v.drop_empty_records), SortRecords: bool(v.sort_records), DropUnsupported: bool(v.drop_unsupported)}, err
}
func ParseRINEXRepair(data []byte, observation bool, opt *NativeRINEXRepairOptions) (*RinexRepair, error) {
	length, lengthErr := checkedNativeSize(len(data))
	if lengthErr != nil {
		return nil, lengthErr
	}
	var p *C.SidereonRinexRepair
	var err error
	withCThread(func() {
		var b unsafe.Pointer
		if len(data) > 0 {
			b = C.CBytes(data)
			if b == nil {
				err = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(b)
		}
		var op *C.SidereonRinexRepairOptions
		if opt != nil {
			var e error
			value, e := cRepairOptions(*opt)
			if e != nil {
				err = e
				return
			}
			memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonRinexRepairOptions{}))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(memory)
			op = (*C.SidereonRinexRepairOptions)(memory)
			*op = value
		}
		if observation {
			err = statusErrorLocked(C.sidereon_rinex_repair_obs((*C.uint8_t)(b), length, op, &p))
		} else {
			err = statusErrorLocked(C.sidereon_rinex_repair_nav((*C.uint8_t)(b), length, op, &p))
		}
		if err != nil && p != nil {
			C.sidereon_rinex_repair_free(p)
			p = nil
		}
	})
	runtime.KeepAlive(data)
	if err != nil {
		return nil, err
	}
	handle, err := newRinexRepair(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_rinex_repair_free(p) })
	}
	return handle, err
}
func (r *RinexRepair) Close() error {
	if r == nil {
		return nil
	}
	return closeProtocolResource(r, r.resource, &r.cleanup)
}
func (r *RinexRepair) Summary() (NativeRINEXLintSummary, error) {
	var v C.SidereonRinexLintSummary
	err := r.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_rinex_repair_summary((*C.SidereonRinexRepair)(p), &v) })
	})
	if err != nil {
		return NativeRINEXLintSummary{}, err
	}
	o := NativeRINEXLintSummary{IsClean: bool(v.is_clean), DecodedFromCRINEX: bool(v.decoded_from_crinex)}
	counts := []*int{&o.FindingCount, &o.FatalCount, &o.ErrorCount, &o.WarningCount, &o.InfoCount}
	nativeCounts := []C.size_t{v.finding_count, v.fatal_count, v.error_count, v.warning_count, v.info_count}
	for i := range counts {
		*counts[i], err = checkedNativeCount(uint64(nativeCounts[i]))
		if err != nil {
			return NativeRINEXLintSummary{}, err
		}
	}
	return o, nil
}
func (r *RinexRepair) Actions() ([]NativeRINEXRepairAction, error) {
	var out []NativeRINEXRepairAction
	err := r.resource.with(func(p unsafe.Pointer) error {
		var w, req C.size_t
		if e := callStatus(func() uint32 { return C.sidereon_rinex_repair_actions((*C.SidereonRinexRepair)(p), nil, 0, &w, &req) }); e != nil {
			return e
		}
		n, e := validateNativeQuery("RINEX repair actions", uint64(w), uint64(req))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRinexRepairAction{})); e != nil {
			return e
		}
		v := make([]C.SidereonRinexRepairAction, n)
		var q *C.SidereonRinexRepairAction
		if n > 0 {
			q = &v[0]
		}
		if e := callStatus(func() uint32 {
			return C.sidereon_rinex_repair_actions((*C.SidereonRinexRepair)(p), q, C.size_t(n), &w, &req)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("RINEX repair actions", n, uint64(w), uint64(req))
		if e != nil {
			return e
		}
		out = make([]NativeRINEXRepairAction, z)
		for i := range out {
			out[i] = NativeRINEXRepairAction{ID: observationFixedString(v[i].id[:]), Message: observationFixedString(v[i].message[:])}
		}
		return nil
	})
	runtime.KeepAlive(r)
	return out, err
}
func (r *RinexRepair) Text() ([]byte, error) {
	var out []byte
	err := r.resource.with(func(p unsafe.Pointer) error {
		var e error
		out, e = copyByteOutput("RINEX repair text", func(b *C.uint8_t, n C.size_t, w, q *C.size_t) uint32 {
			return C.sidereon_rinex_repair_text((*C.SidereonRinexRepair)(p), b, n, w, q)
		})
		return e
	})
	return out, err
}
func (r *RinexRepair) CRINEXText() ([]byte, error) {
	var out []byte
	err := r.resource.with(func(p unsafe.Pointer) error {
		var e error
		out, e = copyByteOutput("RINEX repair CRINEX text", func(b *C.uint8_t, n C.size_t, w, q *C.size_t) uint32 {
			return C.sidereon_rinex_repair_crinex_text((*C.SidereonRinexRepair)(p), b, n, w, q)
		})
		return e
	})
	return out, err
}
func transformCRINEX(data []byte, encode bool) ([]byte, error) {
	var out []byte
	err := withInputError(data, func(in *C.uint8_t, n C.size_t) error {
		var w, r C.size_t
		fn := func(b *C.uint8_t, z C.size_t) uint32 {
			if encode {
				return C.sidereon_crinex_encode(in, n, b, z, &w, &r)
			}
			return C.sidereon_crinex_decode(in, n, b, z, &w, &r)
		}
		if err := callStatus(func() uint32 { return fn(nil, 0) }); err != nil {
			return err
		}
		need, err := validateNativeQuery("CRINEX transform", uint64(w), uint64(r))
		if err != nil {
			return err
		}
		mem, err := checkedNativeMalloc(need, 1)
		if err != nil {
			return err
		}
		if mem != nil {
			defer C.free(mem)
		}
		if err := callStatus(func() uint32 { return fn((*C.uint8_t)(mem), C.size_t(need)) }); err != nil {
			return err
		}
		z, err := validateNativeOutput("CRINEX transform", need, uint64(w), uint64(r))
		if err != nil {
			return err
		}
		out = append([]byte(nil), unsafe.Slice((*byte)(mem), z)...)
		return nil
	})
	return out, err
}
func DecodeCRINEX(data []byte) ([]byte, error) { return transformCRINEX(data, false) }
func EncodeCRINEX(data []byte) ([]byte, error) { return transformCRINEX(data, true) }
