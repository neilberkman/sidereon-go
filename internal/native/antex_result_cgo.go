//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type AntexError struct {
	Kind                                                                              uint32
	HasAntennaID, HasRecord, HasField, HasValue, HasFrequency, HasReason, HasSections bool
	Sections                                                                          int
	AntennaID, Record, Field, Value, Frequency, Reason, Message                       string
}

type AntexOutcome struct {
	IsOK   bool
	Status uint32
	Error  AntexError
}

type AntexHeader struct {
	HasVersion          bool
	Version             float64
	HasSystem           bool
	System              uint32
	HasPcvType          bool
	PcvType             uint32
	HasReferenceAntenna bool
	CommentCount        int
	EndOfHeader         bool
}

type AntexDateTime struct {
	Year                             int32
	Month, Day, Hour, Minute, Second uint8
	FractionDigits, FractionScale    uint64
}

func antexDateTimeToC(value AntexDateTime) C.SidereonAntexDateTime {
	return C.SidereonAntexDateTime{year: C.int32_t(value.Year), month: C.uint8_t(value.Month), day: C.uint8_t(value.Day), hour: C.uint8_t(value.Hour), minute: C.uint8_t(value.Minute), second: C.uint8_t(value.Second), fraction_digits: C.uint64_t(value.FractionDigits), fraction_scale: C.uint64_t(value.FractionScale)}
}

func (a *ANTEX) AntennaAt(id string, epoch AntexDateTime) (*Antenna, bool, error) {
	if a == nil || a.handle == nil {
		return nil, false, ErrClosed
	}
	date := antexDateTimeToC(epoch)
	var result *C.SidereonAntenna
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withString(id, func(input *C.char) uint32 {
			return uint32(C.sidereon_antex_antenna_at((*C.SidereonAntex)(pointer), input, &date, &result))
		})
	})
	runtime.KeepAlive(a)
	runtime.KeepAlive(id)
	if err != nil {
		if result != nil {
			withCThread(func() { C.sidereon_antenna_free(result) })
		}
		return nil, false, err
	}
	if result == nil {
		return nil, false, nil
	}
	return antennaFromPointer(result), true, nil
}

func (a *ANTEX) SatelliteAntenna(prn string, epoch AntexDateTime) (*Antenna, bool, error) {
	if a == nil || a.handle == nil {
		return nil, false, ErrClosed
	}
	date := antexDateTimeToC(epoch)
	var result *C.SidereonAntenna
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withString(prn, func(input *C.char) uint32 {
			return uint32(C.sidereon_antex_satellite_antenna((*C.SidereonAntex)(pointer), input, &date, &result))
		})
	})
	runtime.KeepAlive(a)
	runtime.KeepAlive(prn)
	if err != nil {
		if result != nil {
			withCThread(func() { C.sidereon_antenna_free(result) })
		}
		return nil, false, err
	}
	if result == nil {
		return nil, false, nil
	}
	return antennaFromPointer(result), true, nil
}

func (a *ANTEX) Header() (AntexHeader, error) {
	if a == nil || a.handle == nil {
		return AntexHeader{}, ErrClosed
	}
	var result AntexHeader
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var raw C.SidereonAntexHeader
			if err := callStatus(func() uint32 { return uint32(C.sidereon_antex_header((*C.SidereonAntex)(pointer), &raw)) }); err != nil {
				return err
			}
			count, err := sizeTToInt(raw.comment_count, "ANTEX header comment count")
			if err != nil {
				return err
			}
			result = AntexHeader{HasVersion: bool(raw.has_version), Version: float64(raw.version), HasSystem: bool(raw.has_system), System: uint32(raw.system), HasPcvType: bool(raw.has_pcv_type), PcvType: uint32(raw.pcv_type), HasReferenceAntenna: bool(raw.has_reference_antenna), CommentCount: count, EndOfHeader: bool(raw.end_of_header)}
			return nil
		})
	})
	runtime.KeepAlive(a)
	return result, err
}

func (a *ANTEX) HeaderText(part uint32) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	if part > 2 {
		return "", invalidArgument("ANTEX header text part out of range")
	}
	var text []byte
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		var err error
		text, err = copyNativeBytes("ANTEX header text", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_antex_header_text((*C.SidereonAntex)(pointer), C.uint32_t(part), out, n, w, r)
		})
		return err
	})
	runtime.KeepAlive(a)
	return string(text), err
}

func (a *ANTEX) HeaderComment(index int) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	if index < 0 {
		return "", invalidArgument("negative ANTEX header comment index")
	}
	var text []byte
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		var err error
		text, err = copyNativeBytes("ANTEX header comment", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_antex_header_comment((*C.SidereonAntex)(pointer), C.size_t(index), out, n, w, r)
		})
		return err
	})
	runtime.KeepAlive(a)
	return string(text), err
}

func (a *ANTEX) OuterCommentCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_antex_outer_comment_count((*C.SidereonAntex)(pointer), &count))
			})
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "ANTEX outer comment count")
}

func (a *ANTEX) OuterComment(index int) (string, int, error) {
	if a == nil || a.handle == nil {
		return "", 0, ErrClosed
	}
	if index < 0 {
		return "", 0, invalidArgument("negative ANTEX outer comment index")
	}
	var before C.size_t
	var text []byte
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		var err error
		text, err = copyNativeBytes("ANTEX outer comment", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_antex_outer_comment((*C.SidereonAntex)(pointer), C.size_t(index), &before, out, n, w, r)
		})
		return err
	})
	runtime.KeepAlive(a)
	if err != nil {
		return "", 0, err
	}
	count, err := sizeTToInt(before, "ANTEX preceding block count")
	return string(text), count, err
}

func (a *ANTEX) SkippedRecords() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 { return uint32(C.sidereon_antex_skipped_records((*C.SidereonAntex)(pointer), &count)) })
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "ANTEX skipped record count")
}

func (a *ANTEX) BlockCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 { return uint32(C.sidereon_antex_block_count((*C.SidereonAntex)(pointer), &count)) })
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "ANTEX block count")
}

func (a *ANTEX) Block(index int) (*Antenna, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("negative ANTEX block index")
	}
	var result *C.SidereonAntenna
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_antex_block((*C.SidereonAntex)(pointer), C.size_t(index), &result))
			})
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		if result != nil {
			withCThread(func() { C.sidereon_antenna_free(result) })
		}
		return nil, err
	}
	if result == nil {
		return nil, errNilNativeHandle
	}
	return antennaFromPointer(result), nil
}

func antexResultText(result *C.SidereonAntexResult, part uint32, label string) ([]byte, error) {
	return copyNativeBytes(label, func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_antex_result_error_text(result, C.uint32_t(part), out, n, w, r)
	})
}

func antexOutcome(result *C.SidereonAntexResult) (AntexOutcome, error) {
	var raw C.SidereonAntexOutcome
	if err := callStatus(func() uint32 { return uint32(C.sidereon_antex_result_get_outcome(result, &raw)) }); err != nil {
		return AntexOutcome{}, err
	}
	e := raw.error
	out := AntexOutcome{IsOK: bool(raw.is_ok), Status: uint32(raw.status), Error: AntexError{Kind: uint32(e.kind), HasAntennaID: bool(e.has_antenna_id), HasRecord: bool(e.has_record), HasField: bool(e.has_field), HasValue: bool(e.has_value), HasFrequency: bool(e.has_frequency), HasReason: bool(e.has_reason), HasSections: bool(e.has_sections)}}
	if out.Error.HasSections {
		n, err := sizeTToInt(e.sections, "ANTEX error section count")
		if err != nil {
			return AntexOutcome{}, err
		}
		out.Error.Sections = n
	}
	for _, part := range []struct {
		id      uint32
		present bool
		dst     *string
		name    string
	}{{1, out.Error.HasAntennaID, &out.Error.AntennaID, "ANTEX error antenna id"}, {2, out.Error.HasRecord, &out.Error.Record, "ANTEX error record"}, {3, out.Error.HasField, &out.Error.Field, "ANTEX error field"}, {4, out.Error.HasValue, &out.Error.Value, "ANTEX error value"}, {5, out.Error.HasFrequency, &out.Error.Frequency, "ANTEX error frequency"}, {6, out.Error.HasReason, &out.Error.Reason, "ANTEX error reason"}, {0, true, &out.Error.Message, "ANTEX error message"}} {
		if !part.present {
			continue
		}
		b, err := antexResultText(result, part.id, part.name)
		if err != nil {
			return AntexOutcome{}, err
		}
		*part.dst = string(b)
	}
	return out, nil
}

func ParseANTEXWithOutcome(data []byte) (*ANTEX, AntexOutcome, error) {
	var pointer *C.SidereonAntex
	var result *C.SidereonAntexResult
	err := withInput(data, func(input *C.uint8_t, n C.size_t) uint32 {
		return C.sidereon_antex_parse_result(input, n, &pointer, &result)
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_antex_free(pointer) })
		}
		if result != nil {
			withCThread(func() { C.sidereon_antex_result_free(result) })
		}
		return nil, AntexOutcome{}, err
	}
	if result == nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_antex_free(pointer) })
		}
		return nil, AntexOutcome{}, errNilNativeHandle
	}
	defer withCThread(func() { C.sidereon_antex_result_free(result) })
	out, err := withOutcomeError(result)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_antex_free(pointer) })
		}
		return nil, AntexOutcome{}, err
	}
	if !out.IsOK {
		if pointer != nil {
			withCThread(func() { C.sidereon_antex_free(pointer) })
		}
		return nil, out, nil
	}
	if pointer == nil {
		return nil, AntexOutcome{}, errNilNativeHandle
	}
	runtime.KeepAlive(data)
	return &ANTEX{handle: newPositioningHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) { C.sidereon_antex_free((*C.SidereonAntex)(p)) })}, out, nil
}

func withOutcomeError(result *C.SidereonAntexResult) (AntexOutcome, error) {
	var out AntexOutcome
	err := withCThreadError(func() error { var e error; out, e = antexOutcome(result); return e })
	return out, err
}

func (a *ANTEX) EncodeWithOutcome() ([]byte, AntexOutcome, error) {
	if a == nil || a.handle == nil {
		return nil, AntexOutcome{}, ErrClosed
	}
	var output []byte
	var out AntexOutcome
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonAntexResult
			if err := callStatus(func() uint32 { return uint32(C.sidereon_antex_encode_result((*C.SidereonAntex)(pointer), &result)) }); err != nil {
				return err
			}
			if result == nil {
				return errNilNativeHandle
			}
			defer C.sidereon_antex_result_free(result)
			var err error
			out, err = antexOutcome(result)
			if err != nil {
				return err
			}
			if out.IsOK {
				output, err = copyNativeBytes("ANTEX text", func(p *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_antex_result_get_text(result, p, n, w, r)
				})
			}
			return err
		})
	})
	runtime.KeepAlive(a)
	return output, out, err
}
