//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type NativeSp3AgreementMetric struct {
	Epoch              NativeClockEpoch
	EpochJ2000S        float64
	Satellite          string
	PositionMembers    int
	PositionRMSPresent bool
	PositionRMSM       float64
	PositionMaxPresent bool
	PositionMaxM       float64
	ClockMembers       int
	ClockRMSPresent    bool
	ClockRMSS          float64
	ClockMaxPresent    bool
	ClockMaxS          float64
}
type NativeSp3ClockOmission struct {
	Epoch        NativeClockEpoch
	EpochJ2000S  float64
	Satellite    string
	Source       int
	Reason       uint32
	HasPreferred bool
	Preferred    int
	CellHasClock bool
}
type NativeSp3DroppedInputEpoch struct {
	Source      int
	EpochIndex  int
	Epoch       NativeClockEpoch
	EpochJ2000S float64
	Reason      uint32
}
type NativeSp3MergeEpoch struct {
	Epoch       NativeClockEpoch
	EpochJ2000S float64
}
type NativeSp3ProvenanceInfo struct {
	Recorded        bool
	Mode            uint32
	CellCount       int
	TransitionCount int
	CoverageCount   int
}
type NativeSp3CellSelection struct {
	Kind        uint32
	HasSource   bool
	Source      int
	HasRule     bool
	Rule        uint32
	MemberCount int
}
type NativeSp3CellProvenance struct {
	Epoch       NativeClockEpoch
	EpochJ2000S float64
	Satellite   string
	HasPosition bool
	Position    NativeSp3CellSelection
	HasClock    bool
	Clock       NativeSp3CellSelection
}
type NativeSp3ContributorCoverage struct {
	Source           int
	CellsContributed int
	CellsSelected    int
	HasFirstEpoch    bool
	FirstEpoch       NativeClockEpoch
	FirstEpochJ2000S float64
	HasLastEpoch     bool
	LastEpoch        NativeClockEpoch
	LastEpochJ2000S  float64
	CellsAbsent      int
}
type NativeSp3PrecedenceTransition struct {
	Satellite     string
	Epoch         NativeClockEpoch
	EpochJ2000S   float64
	HasFromSource bool
	FromSource    int
	HasToSource   bool
	ToSource      int
	Reason        uint32
}

func sp3CopySliceLocked[T any](label string, call func(unsafe.Pointer, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]T, error) {
	var written, required C.size_t
	if err := statusSp3ErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(*new(T))); err != nil {
		return nil, err
	}
	values := make([]T, count)
	var out unsafe.Pointer
	if count > 0 {
		out = unsafe.Pointer(&values[0])
	}
	written, required = 0, 0
	if err := statusSp3ErrorLocked(uint32(call(out, C.size_t(count), &written, &required))); err != nil {
		return nil, err
	}
	n, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	return values[:n], nil
}

func nativeSize(value C.size_t, label string) (int, error) { return sizeTToInt(value, label) }
func sp3SelectionFromC(v C.SidereonSp3CellSelection) (NativeSp3CellSelection, error) {
	source, err := nativeSize(v.source, "SP3 provenance selected source")
	if err != nil {
		return NativeSp3CellSelection{}, err
	}
	count, err := nativeSize(v.member_count, "SP3 provenance member count")
	if err != nil {
		return NativeSp3CellSelection{}, err
	}
	return NativeSp3CellSelection{Kind: uint32(v.kind), HasSource: bool(v.has_source), Source: source, HasRule: bool(v.has_rule), Rule: uint32(v.rule), MemberCount: count}, nil
}

func (r *Sp3MergeReport) AgreementMetrics() (result []NativeSp3AgreementMetric, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			rows, e := sp3CopySliceLocked[C.SidereonSp3AgreementMetric]("SP3 merge agreement metric", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_agreement_metrics((*C.SidereonSp3MergeReport)(pointer), (*C.SidereonSp3AgreementMetric)(out), n, w, req)
			})
			if e != nil {
				return e
			}
			result = make([]NativeSp3AgreementMetric, len(rows))
			for i, v := range rows {
				pm, e := nativeSize(v.position_members, "SP3 position member count")
				if e != nil {
					return e
				}
				cm, e := nativeSize(v.clock_members, "SP3 clock member count")
				if e != nil {
					return e
				}
				result[i] = NativeSp3AgreementMetric{Epoch: clockEpochFromC(v.epoch), EpochJ2000S: float64(v.epoch_j2000_seconds), Satellite: tokenFromC(v.sat_id), PositionMembers: pm, PositionRMSPresent: bool(v.has_position_rms_m), PositionRMSM: float64(v.position_rms_m), PositionMaxPresent: bool(v.has_position_max_m), PositionMaxM: float64(v.position_max_m), ClockMembers: cm, ClockRMSPresent: bool(v.has_clock_rms_s), ClockRMSS: float64(v.clock_rms_s), ClockMaxPresent: bool(v.has_clock_max_s), ClockMaxS: float64(v.clock_max_s)}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) ClockOmissions() (result []NativeSp3ClockOmission, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			rows, e := sp3CopySliceLocked[C.SidereonSp3ClockOmission]("SP3 merge clock omission", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_clock_omissions((*C.SidereonSp3MergeReport)(pointer), (*C.SidereonSp3ClockOmission)(out), n, w, req)
			})
			if e != nil {
				return e
			}
			result = make([]NativeSp3ClockOmission, len(rows))
			for i, v := range rows {
				source, e := nativeSize(v.source, "SP3 omitted clock source")
				if e != nil {
					return e
				}
				preferred, e := nativeSize(v.preferred, "SP3 preferred clock source")
				if e != nil {
					return e
				}
				result[i] = NativeSp3ClockOmission{Epoch: clockEpochFromC(v.epoch), EpochJ2000S: float64(v.epoch_j2000_seconds), Satellite: tokenFromC(v.sat_id), Source: source, Reason: uint32(v.reason), HasPreferred: bool(v.has_preferred), Preferred: preferred, CellHasClock: bool(v.cell_has_clock)}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) DroppedInputEpochs() (result []NativeSp3DroppedInputEpoch, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			rows, e := sp3CopySliceLocked[C.SidereonSp3DroppedInputEpoch]("SP3 dropped input epoch", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_dropped_input_epochs((*C.SidereonSp3MergeReport)(pointer), (*C.SidereonSp3DroppedInputEpoch)(out), n, w, req)
			})
			if e != nil {
				return e
			}
			result = make([]NativeSp3DroppedInputEpoch, len(rows))
			for i, v := range rows {
				source, e := nativeSize(v.source, "SP3 dropped epoch source")
				if e != nil {
					return e
				}
				idx, e := nativeSize(v.epoch_index, "SP3 dropped epoch index")
				if e != nil {
					return e
				}
				result[i] = NativeSp3DroppedInputEpoch{Source: source, EpochIndex: idx, Epoch: clockEpochFromC(v.epoch), EpochJ2000S: float64(v.epoch_j2000_seconds), Reason: uint32(v.reason)}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) OmittedEpochs() (result []NativeSp3MergeEpoch, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			rows, e := sp3CopySliceLocked[C.SidereonSp3MergeEpoch]("SP3 merge omitted epoch", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_omitted_epochs((*C.SidereonSp3MergeReport)(pointer), (*C.SidereonSp3MergeEpoch)(out), n, w, req)
			})
			if e != nil {
				return e
			}
			result = make([]NativeSp3MergeEpoch, len(rows))
			for i, v := range rows {
				result[i] = NativeSp3MergeEpoch{Epoch: clockEpochFromC(v.epoch), EpochJ2000S: float64(v.epoch_j2000_seconds)}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) ProvenanceInfo() (result NativeSp3ProvenanceInfo, err error) {
	if r == nil || r.handle == nil {
		return result, ErrClosed
	}
	var value C.SidereonSp3MergeProvenanceInfo
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatusWithSp3Diagnostics(func() uint32 {
			return uint32(C.sidereon_sp3_merge_report_provenance((*C.SidereonSp3MergeReport)(pointer), &value))
		})
	})
	runtime.KeepAlive(r)
	if err != nil {
		return result, err
	}
	cell, e := nativeSize(value.cell_count, "SP3 provenance cell count")
	if e != nil {
		return result, e
	}
	transition, e := nativeSize(value.transition_count, "SP3 provenance transition count")
	if e != nil {
		return result, e
	}
	coverage, e := nativeSize(value.coverage_count, "SP3 provenance coverage count")
	if e != nil {
		return result, e
	}
	return NativeSp3ProvenanceInfo{Recorded: bool(value.recorded), Mode: uint32(value.mode), CellCount: cell, TransitionCount: transition, CoverageCount: coverage}, nil
}

func (r *Sp3MergeReport) ProvenanceCells() (result []NativeSp3CellProvenance, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			rows, e := sp3CopySliceLocked[C.SidereonSp3CellProvenance]("SP3 provenance cell", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_provenance_cells((*C.SidereonSp3MergeReport)(pointer), (*C.SidereonSp3CellProvenance)(out), n, w, req)
			})
			if e != nil {
				return e
			}
			result = make([]NativeSp3CellProvenance, len(rows))
			for i, v := range rows {
				p, e := sp3SelectionFromC(v.position)
				if e != nil {
					return e
				}
				c, e := sp3SelectionFromC(v.clock)
				if e != nil {
					return e
				}
				result[i] = NativeSp3CellProvenance{Epoch: clockEpochFromC(v.epoch), EpochJ2000S: float64(v.epoch_j2000_seconds), Satellite: tokenFromC(v.sat_id), HasPosition: bool(v.has_position), Position: p, HasClock: bool(v.has_clock), Clock: c}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) ProvenanceCellMembers(index int, channel uint32) (result []int, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("SP3 provenance cell index must not be negative")
	}
	idx, err := checkedNativeSize(index)
	if err != nil {
		return nil, err
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			values, err := sp3CopySliceLocked[C.size_t]("SP3 provenance cell members", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_provenance_cell_members((*C.SidereonSp3MergeReport)(pointer), idx, C.uint32_t(channel), (*C.size_t)(out), n, w, req)
			})
			if err != nil {
				return err
			}
			result = make([]int, len(values))
			for i, value := range values {
				result[i], err = sizeTToInt(value, "SP3 provenance cell member")
				if err != nil {
					return err
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) ProvenanceCoverage() (result []NativeSp3ContributorCoverage, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			rows, e := sp3CopySliceLocked[C.SidereonSp3ContributorCoverage]("SP3 contributor coverage", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_provenance_coverage((*C.SidereonSp3MergeReport)(pointer), (*C.SidereonSp3ContributorCoverage)(out), n, w, req)
			})
			if e != nil {
				return e
			}
			result = make([]NativeSp3ContributorCoverage, len(rows))
			for i, v := range rows {
				source, e := nativeSize(v.source, "SP3 contributor source")
				if e != nil {
					return e
				}
				contributed, e := nativeSize(v.cells_contributed, "SP3 contributed cell count")
				if e != nil {
					return e
				}
				selected, e := nativeSize(v.cells_selected, "SP3 selected cell count")
				if e != nil {
					return e
				}
				absent, e := nativeSize(v.cells_absent, "SP3 absent cell count")
				if e != nil {
					return e
				}
				result[i] = NativeSp3ContributorCoverage{Source: source, CellsContributed: contributed, CellsSelected: selected, HasFirstEpoch: bool(v.has_first_epoch), FirstEpoch: clockEpochFromC(v.first_epoch), FirstEpochJ2000S: float64(v.first_epoch_j2000_seconds), HasLastEpoch: bool(v.has_last_epoch), LastEpoch: clockEpochFromC(v.last_epoch), LastEpochJ2000S: float64(v.last_epoch_j2000_seconds), CellsAbsent: absent}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) ProvenanceTransitions() (result []NativeSp3PrecedenceTransition, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			rows, e := sp3CopySliceLocked[C.SidereonSp3PrecedenceTransition]("SP3 precedence transition", func(out unsafe.Pointer, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_provenance_transitions((*C.SidereonSp3MergeReport)(pointer), (*C.SidereonSp3PrecedenceTransition)(out), n, w, req)
			})
			if e != nil {
				return e
			}
			result = make([]NativeSp3PrecedenceTransition, len(rows))
			for i, v := range rows {
				from, e := nativeSize(v.from_source, "SP3 transition prior source")
				if e != nil {
					return e
				}
				to, e := nativeSize(v.to_source, "SP3 transition next source")
				if e != nil {
					return e
				}
				result[i] = NativeSp3PrecedenceTransition{Satellite: tokenFromC(v.sat_id), Epoch: clockEpochFromC(v.epoch), EpochJ2000S: float64(v.epoch_j2000_seconds), HasFromSource: bool(v.has_from_source), FromSource: from, HasToSource: bool(v.has_to_source), ToSource: to, Reason: uint32(v.reason)}
			}
			return nil
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) ContinuityJSON() (result []byte, err error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	err = r.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var e error
			result, e = copyNativeBytesLockedWithStatus("SP3 merge continuity report", func(out *C.uint8_t, n C.size_t, w, req *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_sp3_merge_report_continuity_json((*C.SidereonSp3MergeReport)(pointer), out, n, w, req)
			}, statusSp3ErrorLocked)
			return e
		})
	})
	runtime.KeepAlive(r)
	return result, err
}

func (r *Sp3MergeReport) ContinuitySelectedNodes(satellite string, from, through float64) (verified bool, result []float64, err error) {
	if r == nil || r.handle == nil {
		return false, nil, ErrClosed
	}
	err = r.handle.with(func(reportPointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return withTokenError(satellite, "satellite token", func(id *C.char) error {
				var written, required C.size_t
				var copied []C.double
				var verifiedC C.bool
				copyCall := func(out *C.double, n C.size_t) C.enum_SidereonStatus {
					return C.sidereon_sp3_merge_report_continuity_selected_nodes((*C.SidereonSp3MergeReport)(reportPointer), id, C.double(from), C.double(through), &verifiedC, out, n, &written, &required)
				}
				if err := statusSp3ErrorLocked(uint32(copyCall(nil, 0))); err != nil {
					return err
				}
				count, err := validateNativeQuery("SP3 merge continuity selected nodes", uint64(written), uint64(required))
				if err != nil {
					return err
				}
				if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.double(0))); err != nil {
					return err
				}
				copied = make([]C.double, count)
				var out *C.double
				if count > 0 {
					out = &copied[0]
				}
				written, required = 0, 0
				if err := statusSp3ErrorLocked(uint32(copyCall(out, C.size_t(count)))); err != nil {
					return err
				}
				n, err := validateTwoPassCounts("SP3 merge continuity selected nodes", count, count, uint64(written), uint64(required))
				if err != nil {
					return err
				}
				verified = bool(verifiedC)
				result = make([]float64, n)
				for i, value := range copied[:n] {
					result[i] = float64(value)
				}
				return nil
			})
		})
	})
	runtime.KeepAlive(r)
	return verified, result, err
}
