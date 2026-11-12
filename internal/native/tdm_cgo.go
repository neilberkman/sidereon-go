//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

const (
	TDMObservableRange                = uint32(C.SIDEREON_TDM_OBSERVABLE_RANGE)
	TDMObservableDopplerInstantaneous = uint32(C.SIDEREON_TDM_OBSERVABLE_DOPPLER_INSTANTANEOUS)
	TDMObservableDopplerIntegrated    = uint32(C.SIDEREON_TDM_OBSERVABLE_DOPPLER_INTEGRATED)
	TDMObservableReceiveFreq          = uint32(C.SIDEREON_TDM_OBSERVABLE_RECEIVE_FREQ)
	TDMObservableTransmitFreq         = uint32(C.SIDEREON_TDM_OBSERVABLE_TRANSMIT_FREQ)
	TDMObservableTransmitFreqRate     = uint32(C.SIDEREON_TDM_OBSERVABLE_TRANSMIT_FREQ_RATE)
	TDMObservableAngle1               = uint32(C.SIDEREON_TDM_OBSERVABLE_ANGLE1)
	TDMObservableAngle2               = uint32(C.SIDEREON_TDM_OBSERVABLE_ANGLE2)
	TDMObservableOther                = uint32(C.SIDEREON_TDM_OBSERVABLE_OTHER)
	TDMLeniencyStrict                 = uint32(C.SIDEREON_TDM_LENIENCY_STRICT)
	TDMLeniencyForgive                = uint32(C.SIDEREON_TDM_LENIENCY_FORGIVE)
)

type TDMPolicy struct{ NonPrintable, MissingKeywords, LongLines, EmptyDataSections, RecordOrder, DuplicateRecords, KeywordOrder, FinalTerminator uint32 }
type TDMWritePolicy struct{ NonPrintable, MissingKeywords, LongLines, EmptyDataSections, RecordOrder, DuplicateRecords, KeywordOrder, FinalTerminator, RepeatedKeywords uint32 }
type TDMComment struct {
	SegmentIndex int
	Section      uint32
	BeforeRecord int
	Text         string
}
type TDMField struct {
	SegmentIndex int
	Section      uint32
	Key, Value   string
}
type TDMHeader struct {
	Version                             string
	CreationDate, Originator, MessageID *string
	Comments                            []TDMComment
	Fields                              []TDMField
	SegmentCount                        int
}
type TDMParticipantFull struct {
	SegmentIndex int
	Index        uint8
	Name         string
}
type TDMPathFull struct {
	SegmentIndex int
	Key          string
	HasIndex     bool
	Index        uint8
	Participants []uint8
}
type TDMRecord struct {
	SegmentIndex                         int
	Observable                           uint32
	HasObservableParticipant             bool
	ObservableParticipant                uint8
	ObservableName                       string
	Unit                                 uint32
	UnitLabel, Keyword, Epoch, ValueText string
	Value                                float64
}
type TDMSegment struct {
	SegmentIndex                             int
	Mode, TimetagRef, TimeSystem             *string
	RangeUnit                                uint32
	ParticipantCount, PathCount, RecordCount int
}
type TDMNotice struct {
	Kind    string
	Fields  map[string]EngineJSONValue
	Message string
}

type TDMFull struct {
	_      noCopy
	handle *positioningHandle
}

func releaseTDM(p unsafe.Pointer) { C.sidereon_tdm_free((*C.SidereonTdm)(p)) }
func newTDMFull(p *C.SidereonTdm) (*TDMFull, error) {
	if p == nil {
		return nil, missingNativeHandle("TDMFull")
	}
	return &TDMFull{handle: newPositioningHandle(unsafe.Pointer(p), releaseTDM)}, nil
}
func (t *TDMFull) Close() error {
	if t == nil || t.handle == nil {
		return nil
	}
	return t.handle.close()
}

func (t *TDMFull) Equal(other *TDMFull) (bool, error) {
	if t == nil || t.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	var equal C.bool
	err := withPositioningPair(t.handle, other.handle, func(first, second unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_tdm_equal(
				(*C.SidereonTdm)(first),
				(*C.SidereonTdm)(second),
				&equal,
			))
		})
	})
	return bool(equal), err
}

func cPolicy(p TDMPolicy) C.SidereonTdmPolicy {
	return C.SidereonTdmPolicy{non_printable: C.uint32_t(p.NonPrintable), missing_keywords: C.uint32_t(p.MissingKeywords), long_lines: C.uint32_t(p.LongLines), empty_data_sections: C.uint32_t(p.EmptyDataSections), record_order: C.uint32_t(p.RecordOrder), duplicate_records: C.uint32_t(p.DuplicateRecords), keyword_order: C.uint32_t(p.KeywordOrder), final_terminator: C.uint32_t(p.FinalTerminator)}
}
func goPolicy(p C.SidereonTdmPolicy) TDMPolicy {
	return TDMPolicy{uint32(p.non_printable), uint32(p.missing_keywords), uint32(p.long_lines), uint32(p.empty_data_sections), uint32(p.record_order), uint32(p.duplicate_records), uint32(p.keyword_order), uint32(p.final_terminator)}
}
func cWritePolicy(p TDMWritePolicy) C.SidereonTdmWritePolicy {
	return C.SidereonTdmWritePolicy{non_printable: C.uint32_t(p.NonPrintable), missing_keywords: C.uint32_t(p.MissingKeywords), long_lines: C.uint32_t(p.LongLines), empty_data_sections: C.uint32_t(p.EmptyDataSections), record_order: C.uint32_t(p.RecordOrder), duplicate_records: C.uint32_t(p.DuplicateRecords), keyword_order: C.uint32_t(p.KeywordOrder), final_terminator: C.uint32_t(p.FinalTerminator), repeated_keywords: C.uint32_t(p.RepeatedKeywords)}
}
func goWritePolicy(p C.SidereonTdmWritePolicy) TDMWritePolicy {
	return TDMWritePolicy{uint32(p.non_printable), uint32(p.missing_keywords), uint32(p.long_lines), uint32(p.empty_data_sections), uint32(p.record_order), uint32(p.duplicate_records), uint32(p.keyword_order), uint32(p.final_terminator), uint32(p.repeated_keywords)}
}
func TDMPolicyDefault() TDMPolicy { return goPolicy(C.sidereon_tdm_policy_default()) }
func TDMPolicyStrict() TDMPolicy  { return goPolicy(C.sidereon_tdm_policy_strict()) }
func TDMPolicyLenient() TDMPolicy { return goPolicy(C.sidereon_tdm_policy_lenient()) }
func TDMWritePolicyDefault() TDMWritePolicy {
	return goWritePolicy(C.sidereon_tdm_write_policy_default())
}
func TDMWritePolicyStrict() TDMWritePolicy {
	return goWritePolicy(C.sidereon_tdm_write_policy_strict())
}
func TDMWritePolicyLenient() TDMWritePolicy {
	return goWritePolicy(C.sidereon_tdm_write_policy_lenient())
}
func (p TDMWritePolicy) AsRead() (TDMPolicy, error) {
	var out C.SidereonTdmPolicy
	e := callStatus(func() uint32 { return uint32(C.sidereon_tdm_write_policy_as_read(cWritePolicy(p), &out)) })
	return goPolicy(out), e
}

func bytesPointer(data []byte) unsafe.Pointer {
	if len(data) == 0 {
		return nil
	}
	return unsafe.Pointer(&data[0])
}
func CreateTDMFull(version string) (*TDMFull, error) {
	data := []byte(version)
	var out *C.SidereonTdm
	e := callStatus(func() uint32 {
		return uint32(C.sidereon_tdm_create((*C.uint8_t)(bytesPointer(data)), C.size_t(len(data)), &out))
	})
	runtime.KeepAlive(data)
	if e != nil {
		return nil, e
	}
	return newTDMFull(out)
}
func ParseTDMFull(data []byte) (*TDMFull, error) {
	copyData := append([]byte(nil), data...)
	var out *C.SidereonTdm
	e := callStatus(func() uint32 {
		return uint32(C.sidereon_tdm_parse_kvn((*C.uint8_t)(bytesPointer(copyData)), C.size_t(len(copyData)), &out))
	})
	runtime.KeepAlive(copyData)
	if e != nil {
		return nil, e
	}
	return newTDMFull(out)
}
func ParseTDMFullWithPolicy(data []byte, p TDMPolicy) (*TDMFull, []TDMNotice, error) {
	copyData := append([]byte(nil), data...)
	var out *C.SidereonTdm
	var warnings *C.SidereonTdmWarnings
	e := callStatus(func() uint32 {
		return uint32(C.sidereon_tdm_parse_kvn_with_policy((*C.uint8_t)(bytesPointer(copyData)), C.size_t(len(copyData)), cPolicy(p), &out, &warnings))
	})
	runtime.KeepAlive(copyData)
	if e != nil {
		return nil, nil, e
	}
	t, ne := newTDMFull(out)
	if ne != nil {
		C.sidereon_tdm_warnings_free(warnings)
		return nil, nil, ne
	}
	values, e := readWarnings(warnings)
	if e != nil {
		_ = t.Close()
		return nil, nil, e
	}
	return t, values, nil
}

func copyBytes(label string, invoke func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t) uint32) ([]byte, error) {
	var w, r C.size_t
	if e := callStatus(func() uint32 { return invoke(nil, 0, &w, &r) }); e != nil {
		return nil, e
	}
	n, e := validateNativeQuery(label, uint64(w), uint64(r))
	if e != nil {
		return nil, e
	}
	out := make([]C.uint8_t, n)
	var p *C.uint8_t
	if n > 0 {
		p = &out[0]
	}
	w, r = 0, 0
	if e = callStatus(func() uint32 { return invoke(p, C.size_t(n), &w, &r) }); e != nil {
		return nil, e
	}
	count, e := validateTwoPassCounts(label, n, n, uint64(w), uint64(r))
	if e != nil {
		return nil, e
	}
	result := make([]byte, count)
	for i := range result {
		result[i] = byte(out[i])
	}
	return result, nil
}
func (t *TDMFull) with(fn func(*C.SidereonTdm) error) error {
	if t == nil || t.handle == nil {
		return ErrClosed
	}
	return t.handle.with(func(p unsafe.Pointer) error { return fn((*C.SidereonTdm)(p)) })
}
func (t *TDMFull) KVN() ([]byte, error) {
	var out []byte
	e := t.with(func(p *C.SidereonTdm) error {
		var x error
		out, x = copyBytes("TDMFull KVN", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_to_kvn(p, b, n, w, r))
		})
		return x
	})
	return out, e
}
func (t *TDMFull) KVNWithPolicy(policy TDMWritePolicy) ([]byte, []TDMNotice, error) {
	var out []byte
	var notices []TDMNotice
	e := t.with(func(p *C.SidereonTdm) error {
		var deps *C.SidereonTdmDepartures
		var w, r C.size_t
		invoke := func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_to_kvn_with_policy(p, cWritePolicy(policy), b, n, w, r, &deps))
		}
		if x := callStatus(func() uint32 { return invoke(nil, 0, &w, &r) }); x != nil {
			return x
		}
		n, x := validateNativeQuery("TDMFull policy KVN", uint64(w), uint64(r))
		if x != nil {
			return x
		}
		if deps != nil {
			C.sidereon_tdm_departures_free(deps)
			deps = nil
		}
		buf := make([]C.uint8_t, n)
		var b *C.uint8_t
		if n > 0 {
			b = &buf[0]
		}
		w, r = 0, 0
		if x = callStatus(func() uint32 { return invoke(b, C.size_t(n), &w, &r) }); x != nil {
			return x
		}
		defer C.sidereon_tdm_departures_free(deps)
		count, x := validateTwoPassCounts("TDMFull policy KVN", n, n, uint64(w), uint64(r))
		if x != nil {
			return x
		}
		out = make([]byte, count)
		for i := range out {
			out[i] = byte(buf[i])
		}
		notices, x = readDepartures(deps)
		return x
	})
	return out, notices, e
}

func decodeNotice(data []byte) (TDMNotice, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var raw map[string]any
	if e := dec.Decode(&raw); e != nil {
		return TDMNotice{}, e
	}
	kind, _ := raw["kind"].(string)
	message, _ := raw["message"].(string)
	fieldObject, ok := raw["fields"].(map[string]any)
	if !ok {
		return TDMNotice{}, errors.New("sidereon: TDM diagnostic fields is not an object")
	}
	fields := make(map[string]EngineJSONValue, len(fieldObject))
	for k, v := range fieldObject {
		x, e := engineJSONValue(v)
		if e != nil {
			return TDMNotice{}, e
		}
		fields[k] = x
	}
	return TDMNotice{kind, fields, message}, nil
}
func readWarnings(p *C.SidereonTdmWarnings) ([]TDMNotice, error) {
	if p == nil {
		return nil, errors.New("sidereon: native TDMFull parse returned no warning handle")
	}
	defer C.sidereon_tdm_warnings_free(p)
	var n C.size_t
	if e := callStatus(func() uint32 { return uint32(C.sidereon_tdm_warnings_count(p, &n)) }); e != nil {
		return nil, e
	}
	count, e := sizeTToInt(n, "TDMFull warning count")
	if e != nil {
		return nil, e
	}
	out := make([]TDMNotice, count)
	for i := range out {
		b, e := copyBytes("TDMFull warning", func(o *C.uint8_t, l C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_warning_detail(p, C.size_t(i), o, l, w, r))
		})
		if e != nil {
			return nil, e
		}
		out[i], e = decodeNotice(b)
		if e != nil {
			return nil, fmt.Errorf("sidereon: decode TDMFull warning: %w", e)
		}
	}
	return out, nil
}
func readDepartures(p *C.SidereonTdmDepartures) ([]TDMNotice, error) {
	if p == nil {
		return nil, errors.New("sidereon: native TDMFull operation returned no departure handle")
	}
	var n C.size_t
	if e := callStatus(func() uint32 { return uint32(C.sidereon_tdm_departures_count(p, &n)) }); e != nil {
		return nil, e
	}
	count, e := sizeTToInt(n, "TDMFull departure count")
	if e != nil {
		return nil, e
	}
	out := make([]TDMNotice, count)
	for i := range out {
		b, e := copyBytes("TDMFull departure", func(o *C.uint8_t, l C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_departure_detail(p, C.size_t(i), o, l, w, r))
		})
		if e != nil {
			return nil, e
		}
		out[i], e = decodeNotice(b)
		if e != nil {
			return nil, fmt.Errorf("sidereon: decode TDMFull departure: %w", e)
		}
	}
	return out, nil
}

func tdmCString(a []C.char) string {
	n := 0
	for n < len(a) && a[n] != 0 {
		n++
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(a[i])
	}
	return string(b)
}
func optionalString(v C.SidereonTdmStringField) *string {
	if !bool(v.has_value) {
		return nil
	}
	x := tdmCString(v.value[:])
	return &x
}
func (t *TDMFull) count(invoke func(*C.SidereonTdm, *C.size_t) C.enum_SidereonStatus) (int, error) {
	var result int
	e := t.with(func(p *C.SidereonTdm) error {
		var n C.size_t
		if x := callStatus(func() uint32 { return uint32(invoke(p, &n)) }); x != nil {
			return x
		}
		var x error
		result, x = sizeTToInt(n, "TDMFull count")
		return x
	})
	return result, e
}
func (t *TDMFull) SegmentCount() (int, error) {
	return t.count(func(p *C.SidereonTdm, n *C.size_t) C.enum_SidereonStatus { return C.sidereon_tdm_segment_count(p, n) })
}
func (t *TDMFull) RecordCount() (int, error) {
	return t.count(func(p *C.SidereonTdm, n *C.size_t) C.enum_SidereonStatus { return C.sidereon_tdm_record_count(p, n) })
}

func readArray[T any](label string, invoke func(*T, C.size_t, *C.size_t, *C.size_t) uint32) ([]T, error) {
	var w, r C.size_t
	if e := callStatus(func() uint32 { return invoke(nil, 0, &w, &r) }); e != nil {
		return nil, e
	}
	n, e := validateNativeQuery(label, uint64(w), uint64(r))
	if e != nil {
		return nil, e
	}
	out := make([]T, n)
	var p *T
	if n > 0 {
		p = &out[0]
	}
	w, r = 0, 0
	if e = callStatus(func() uint32 { return invoke(p, C.size_t(n), &w, &r) }); e != nil {
		return nil, e
	}
	count, e := validateTwoPassCounts(label, n, n, uint64(w), uint64(r))
	return out[:count], e
}
func (t *TDMFull) text(label string, invoke func(*C.SidereonTdm, *C.uint8_t, C.size_t, *C.size_t, *C.size_t) uint32) (string, error) {
	var out []byte
	e := t.with(func(p *C.SidereonTdm) error {
		var x error
		out, x = copyBytes(label, func(b *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 { return invoke(p, b, n, w, r) })
		return x
	})
	return string(out), e
}
func (t *TDMFull) Header() (TDMHeader, error) {
	var result TDMHeader
	e := t.with(func(p *C.SidereonTdm) error {
		var h C.SidereonTdmHeader
		if x := callStatus(func() uint32 { return uint32(C.sidereon_tdm_header(p, &h)) }); x != nil {
			return x
		}
		get := func(part uint32) (string, error) {
			b, x := copyBytes("TDMFull header text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
				return uint32(C.sidereon_tdm_header_text(p, C.uint32_t(part), o, n, w, r))
			})
			return string(b), x
		}
		var x error
		if result.Version, x = get(0); x != nil {
			return x
		}
		if bool(h.has_creation_date) {
			v, x := get(1)
			if x != nil {
				return x
			}
			result.CreationDate = &v
		}
		if bool(h.has_originator) {
			v, x := get(2)
			if x != nil {
				return x
			}
			result.Originator = &v
		}
		if bool(h.has_message_id) {
			v, x := get(3)
			if x != nil {
				return x
			}
			result.MessageID = &v
		}
		result.SegmentCount = int(h.segment_count)
		return nil
	})
	if e != nil {
		return TDMHeader{}, e
	}
	comments, e := t.Comments()
	if e != nil {
		return TDMHeader{}, e
	}
	fields, e := t.Fields()
	if e != nil {
		return TDMHeader{}, e
	}
	for _, v := range comments {
		if v.Section == 0 {
			result.Comments = append(result.Comments, v)
		}
	}
	for _, v := range fields {
		if v.Section == 0 {
			result.Fields = append(result.Fields, v)
		}
	}
	return result, nil
}
func (t *TDMFull) Comments() ([]TDMComment, error) {
	var result []TDMComment
	e := t.with(func(p *C.SidereonTdm) error {
		rows, x := readArray("TDMFull comments", func(o *C.SidereonTdmComment, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_comments(p, o, n, w, r))
		})
		if x != nil {
			return x
		}
		result = make([]TDMComment, len(rows))
		for i, v := range rows {
			txt, x := copyBytes("TDMFull comment text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
				return uint32(C.sidereon_tdm_comment_text(p, C.size_t(i), o, n, w, r))
			})
			if x != nil {
				return x
			}
			result[i] = TDMComment{int(v.segment_index), uint32(v.section), int(v.before_record), string(txt)}
		}
		return nil
	})
	return result, e
}
func (t *TDMFull) Fields() ([]TDMField, error) {
	var result []TDMField
	e := t.with(func(p *C.SidereonTdm) error {
		rows, x := readArray("TDMFull fields", func(o *C.SidereonTdmField, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_fields(p, o, n, w, r))
		})
		if x != nil {
			return x
		}
		result = make([]TDMField, len(rows))
		for i, v := range rows {
			get := func(part uint32) (string, error) {
				b, x := copyBytes("TDMFull field text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
					return uint32(C.sidereon_tdm_field_text(p, C.size_t(i), C.uint32_t(part), o, n, w, r))
				})
				return string(b), x
			}
			k, x := get(0)
			if x != nil {
				return x
			}
			val, x := get(1)
			if x != nil {
				return x
			}
			result[i] = TDMField{int(v.segment_index), uint32(v.section), k, val}
		}
		return nil
	})
	return result, e
}
func (t *TDMFull) Participants() ([]TDMParticipantFull, error) {
	var result []TDMParticipantFull
	e := t.with(func(p *C.SidereonTdm) error {
		rows, x := readArray("TDMFull participants", func(o *C.SidereonTdmParticipant, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_participants(p, o, n, w, r))
		})
		if x != nil {
			return x
		}
		result = make([]TDMParticipantFull, len(rows))
		for i, v := range rows {
			result[i] = TDMParticipantFull{int(v.segment_index), uint8(v.index), tdmCString(v.name[:])}
		}
		return nil
	})
	return result, e
}
func (t *TDMFull) Paths() ([]TDMPathFull, error) {
	var result []TDMPathFull
	e := t.with(func(p *C.SidereonTdm) error {
		rows, x := readArray("TDMFull paths", func(o *C.SidereonTdmPath, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_paths(p, o, n, w, r))
		})
		if x != nil {
			return x
		}
		result = make([]TDMPathFull, len(rows))
		for i, v := range rows {
			pc := int(v.participant_count)
			ps := make([]uint8, pc)
			for j := range ps {
				ps[j] = uint8(v.participants[j])
			}
			result[i] = TDMPathFull{int(v.segment_index), tdmCString(v.key[:]), bool(v.has_index), uint8(v.index), ps}
		}
		return nil
	})
	return result, e
}
func (t *TDMFull) Segments() ([]TDMSegment, error) {
	var result []TDMSegment
	e := t.with(func(p *C.SidereonTdm) error {
		rows, x := readArray("TDMFull segments", func(o *C.SidereonTdmSegmentSummary, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_segments(p, o, n, w, r))
		})
		if x != nil {
			return x
		}
		result = make([]TDMSegment, len(rows))
		for i, v := range rows {
			result[i] = TDMSegment{int(v.segment_index), optionalString(v.mode), optionalString(v.timetag_ref), optionalString(v.time_system), uint32(v.range_unit), int(v.participant_count), int(v.path_count), int(v.record_count)}
		}
		return nil
	})
	return result, e
}
func (t *TDMFull) Records() ([]TDMRecord, error) {
	var result []TDMRecord
	e := t.with(func(p *C.SidereonTdm) error {
		rows, x := readArray("TDMFull records", func(o *C.SidereonTdmDataRecord, n C.size_t, w, r *C.size_t) uint32 {
			return uint32(C.sidereon_tdm_records(p, o, n, w, r))
		})
		if x != nil {
			return x
		}
		result = make([]TDMRecord, len(rows))
		for i, v := range rows {
			get := func(part uint32) (string, error) {
				b, x := copyBytes("TDMFull record text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
					return uint32(C.sidereon_tdm_record_text(p, C.size_t(i), C.uint32_t(part), o, n, w, r))
				})
				return string(b), x
			}
			texts := make([]string, 5)
			for j := range texts {
				texts[j], x = get(uint32(j))
				if x != nil {
					return x
				}
			}
			result[i] = TDMRecord{int(v.segment_index), uint32(v.observable), bool(v.has_observable_participant), uint8(v.observable_participant), texts[3], uint32(v.unit), texts[4], texts[0], texts[1], texts[2], float64(v.value)}
		}
		return nil
	})
	return result, e
}

type cArena struct{ ptrs []unsafe.Pointer }

func (a *cArena) text(s string) C.SidereonTdmText {
	b := []byte(s)
	var p unsafe.Pointer
	if len(b) > 0 {
		p = C.CBytes(b)
		a.ptrs = append(a.ptrs, p)
	}
	return C.SidereonTdmText{data: (*C.uint8_t)(p), len: C.size_t(len(b))}
}
func (a *cArena) free() {
	for _, p := range a.ptrs {
		C.free(p)
	}
}
func allocSlice[T any](n int) ([]T, unsafe.Pointer) {
	if n == 0 {
		return nil, nil
	}
	p := C.calloc(C.size_t(n), C.size_t(unsafe.Sizeof(*new(T))))
	return unsafe.Slice((*T)(p), n), p
}
func cInputs(fields []TDMField, comments []TDMComment, records []TDMRecord) ([]C.SidereonTdmFieldInput, []C.SidereonTdmCommentInput, []C.SidereonTdmDataRecordInput, func()) {
	a := &cArena{}
	fs, fp := allocSlice[C.SidereonTdmFieldInput](len(fields))
	cs, cp := allocSlice[C.SidereonTdmCommentInput](len(comments))
	rs, rp := allocSlice[C.SidereonTdmDataRecordInput](len(records))
	for i, v := range fields {
		fs[i] = C.SidereonTdmFieldInput{key: a.text(v.Key), value: a.text(v.Value)}
	}
	for i, v := range comments {
		cs[i] = C.SidereonTdmCommentInput{text: a.text(v.Text), before_record: C.size_t(v.BeforeRecord)}
	}
	for i, v := range records {
		rs[i] = C.SidereonTdmDataRecordInput{observable: C.uint32_t(v.Observable), has_observable_participant: C.bool(v.HasObservableParticipant), observable_participant: C.uint8_t(v.ObservableParticipant), observable_name: a.text(v.ObservableName), unit: C.uint32_t(v.Unit), unit_label: a.text(v.UnitLabel), keyword: a.text(v.Keyword), epoch: a.text(v.Epoch), value_text: a.text(v.ValueText), value: C.double(v.Value)}
	}
	return fs, cs, rs, func() { a.free(); C.free(fp); C.free(cp); C.free(rp) }
}
func firstPtr[T any](s []T) *T {
	if len(s) == 0 {
		return nil
	}
	return &s[0]
}
func (t *TDMFull) SetHeader(h TDMHeader) error {
	a := &cArena{}
	defer a.free()
	opt := func(v *string) C.SidereonTdmOptionalText {
		if v == nil {
			return C.SidereonTdmOptionalText{}
		}
		return C.SidereonTdmOptionalText{has_value: true, value: a.text(*v)}
	}
	in := C.SidereonTdmHeaderInput{version: a.text(h.Version), creation_date: opt(h.CreationDate), originator: opt(h.Originator), message_id: opt(h.MessageID)}
	return t.with(func(p *C.SidereonTdm) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_tdm_set_header(p, in)) })
	})
}
func (t *TDMFull) ReplaceHeaderRaw(comments []TDMComment, fields []TDMField) error {
	fs, cs, _, done := cInputs(fields, comments, nil)
	defer done()
	return t.with(func(p *C.SidereonTdm) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_tdm_replace_header_raw(p, firstPtr(cs), C.size_t(len(cs)), firstPtr(fs), C.size_t(len(fs))))
		})
	})
}
func (t *TDMFull) AppendSegment(fields []TDMField, metadataComments, dataComments []TDMComment, records []TDMRecord, policy TDMWritePolicy) ([]TDMNotice, error) {
	fs, mc, rs, done := cInputs(fields, metadataComments, records)
	defer done()
	_, dc, _, done2 := cInputs(nil, dataComments, nil)
	defer done2()
	var out []TDMNotice
	e := t.with(func(p *C.SidereonTdm) error {
		var deps *C.SidereonTdmDepartures
		x := callStatus(func() uint32 {
			return uint32(C.sidereon_tdm_append_segment(p, firstPtr(fs), C.size_t(len(fs)), firstPtr(mc), C.size_t(len(mc)), firstPtr(dc), C.size_t(len(dc)), firstPtr(rs), C.size_t(len(rs)), cWritePolicy(policy), &deps))
		})
		if x != nil {
			return x
		}
		defer C.sidereon_tdm_departures_free(deps)
		out, x = readDepartures(deps)
		return x
	})
	return out, e
}
func (t *TDMFull) ReplaceMetadata(segment int, fields []TDMField, comments []TDMComment, policy TDMWritePolicy) ([]TDMNotice, error) {
	fs, cs, _, done := cInputs(fields, comments, nil)
	defer done()
	var out []TDMNotice
	e := t.with(func(p *C.SidereonTdm) error {
		var deps *C.SidereonTdmDepartures
		x := callStatus(func() uint32 {
			return uint32(C.sidereon_tdm_replace_metadata(p, C.size_t(segment), firstPtr(fs), C.size_t(len(fs)), firstPtr(cs), C.size_t(len(cs)), cWritePolicy(policy), &deps))
		})
		if x != nil {
			return x
		}
		defer C.sidereon_tdm_departures_free(deps)
		out, x = readDepartures(deps)
		return x
	})
	return out, e
}
func (t *TDMFull) ReplaceData(segment int, comments []TDMComment, records []TDMRecord) error {
	_, cs, rs, done := cInputs(nil, comments, records)
	defer done()
	return t.with(func(p *C.SidereonTdm) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_tdm_replace_data(p, C.size_t(segment), firstPtr(cs), C.size_t(len(cs)), firstPtr(rs), C.size_t(len(rs))))
		})
	})
}
func (t *TDMFull) InsertRecord(segment, index int, record TDMRecord) error {
	_, _, rs, done := cInputs(nil, nil, []TDMRecord{record})
	defer done()
	return t.with(func(p *C.SidereonTdm) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_tdm_insert_record(p, C.size_t(segment), C.size_t(index), rs[0]))
		})
	})
}
func (t *TDMFull) RemoveRecord(segment, index int) error {
	return t.with(func(p *C.SidereonTdm) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_tdm_remove_record(p, C.size_t(segment), C.size_t(index))) })
	})
}
func (t *TDMFull) RemoveSegment(index int) error {
	return t.with(func(p *C.SidereonTdm) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_tdm_remove_segment(p, C.size_t(index))) })
	})
}
