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

type BLQComment struct {
	Placement uint32
	Row       int
	Text      string
}
type BLQError struct {
	Kind, ParseKind, WriteKind                                                         uint32
	HasLine, HasExpected, HasFound, HasText, HasBlock, HasCoefficient, HasCommentIndex bool
	Line, Expected, Found, Block, Row, CommentIndex                                    int
	Constituent                                                                        uint32
	Message, Text                                                                      string
}
type BLQOutcome struct {
	IsOK  bool
	Error BLQError
}
type BLQBlocks struct{ handle *positioningHandle }

func blqResultErrorLocked(result *C.SidereonBlqResult) (BLQOutcome, error) {
	var raw C.SidereonBlqOutcome
	if err := callStatus(func() uint32 { return uint32(C.sidereon_blq_result_get_outcome(result, &raw)) }); err != nil {
		return BLQOutcome{}, err
	}
	out := BLQOutcome{IsOK: bool(raw.is_ok)}
	e := raw.error
	out.Error = BLQError{Kind: uint32(e.kind), ParseKind: uint32(e.parse_kind), WriteKind: uint32(e.write_kind), HasLine: bool(e.has_line), HasExpected: bool(e.has_expected), HasFound: bool(e.has_found), HasText: bool(e.has_text), HasBlock: bool(e.has_block), HasCoefficient: bool(e.has_coefficient), HasCommentIndex: bool(e.has_comment_index), Constituent: uint32(e.constituent)}
	for _, v := range []struct {
		src  C.size_t
		dst  *int
		name string
	}{{e.line, &out.Error.Line, "line"}, {e.expected, &out.Error.Expected, "expected"}, {e.found, &out.Error.Found, "found"}, {e.block, &out.Error.Block, "block"}, {e.row, &out.Error.Row, "row"}, {e.comment_index, &out.Error.CommentIndex, "comment index"}} {
		n, err := sizeTToInt(v.src, v.name)
		if err != nil {
			return BLQOutcome{}, err
		}
		*v.dst = n
	}
	if out.Error.Kind != 0 {
		msg, err := copyNativeBytes("BLQ error message", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_blq_result_error_text(result, 0, b, n, w, r)
		})
		if err != nil {
			return BLQOutcome{}, err
		}
		out.Error.Message = string(msg)
		if out.Error.HasText {
			txt, err := copyNativeBytes("BLQ error text", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_blq_result_error_text(result, 1, b, n, w, r)
			})
			if err != nil {
				return BLQOutcome{}, err
			}
			out.Error.Text = string(txt)
		}
	}
	return out, nil
}
func newBLQBlocks(p *C.SidereonBlqBlocks) (*BLQBlocks, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	return &BLQBlocks{newPositioningHandle(unsafe.Pointer(p), func(v unsafe.Pointer) { C.sidereon_blq_blocks_free((*C.SidereonBlqBlocks)(v)) })}, nil
}
func ParseBLQ(data []byte) (*BLQBlocks, BLQOutcome, error) {
	var p *C.SidereonBlqBlocks
	var r *C.SidereonBlqResult
	var out BLQOutcome
	var opErr error
	length, sizeErr := checkedNativeSize(len(data))
	if sizeErr != nil {
		return nil, out, sizeErr
	}
	withCThread(func() {
		in, copyErr := copyNativeInput(data)
		if copyErr != nil {
			opErr = copyErr
			return
		}
		defer freeNativeInput(in)
		status := C.sidereon_blq_parse((*C.uint8_t)(in), length, &p, &r)
		if r != nil {
			defer C.sidereon_blq_result_free(r)
			out, opErr = blqResultErrorLocked(r)
		}
		if uint32(status) != 0 && opErr == nil {
			if p != nil {
				C.sidereon_blq_blocks_free(p)
				p = nil
			}
			if r == nil {
				opErr = statusErrorLocked(uint32(status))
			}
		} else if r == nil && opErr == nil {
			opErr = errors.New("sidereon: native BLQ parser returned no result")
		}
	})
	runtime.KeepAlive(data)
	if opErr != nil {
		if p != nil {
			withCThread(func() { C.sidereon_blq_blocks_free(p) })
		}
		return nil, out, opErr
	}
	if !out.IsOK {
		if p != nil {
			withCThread(func() { C.sidereon_blq_blocks_free(p) })
		}
		return nil, out, nil
	}
	b, err := newBLQBlocks(p)
	return b, out, err
}
func NewBLQBlocks() (*BLQBlocks, error) {
	var p *C.SidereonBlqBlocks
	var err error
	withCThread(func() { err = callStatus(func() uint32 { return uint32(C.sidereon_blq_blocks_new(&p)) }) })
	if err != nil {
		return nil, err
	}
	return newBLQBlocks(p)
}
func (b *BLQBlocks) Close() error {
	if b == nil {
		return nil
	}
	return b.handle.close()
}
func (b *BLQBlocks) Count() (int, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	err := b.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_blq_blocks_count((*C.SidereonBlqBlocks)(p), &n)) })
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(n, "BLQ block count")
}
func (b *BLQBlocks) Station(i int) (string, error) {
	if b == nil || b.handle == nil {
		return "", ErrClosed
	}
	idx, err := checkedNativeSize(i)
	if err != nil {
		return "", err
	}
	var out []byte
	err = b.handle.with(func(p unsafe.Pointer) error {
		out, err = copyNativeBytes("BLQ station", func(x *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_blq_blocks_station((*C.SidereonBlqBlocks)(p), idx, x, n, w, r)
		})
		return err
	})
	runtime.KeepAlive(b)
	return string(out), err
}
func (b *BLQBlocks) Coefficients(i int) (OceanLoadingBLQ, error) {
	if b == nil || b.handle == nil {
		return OceanLoadingBLQ{}, ErrClosed
	}
	idx, err := checkedNativeSize(i)
	if err != nil {
		return OceanLoadingBLQ{}, err
	}
	var c C.SidereonOceanLoadingBlq
	err = b.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_blq_blocks_coefficients((*C.SidereonBlqBlocks)(p), idx, &c)) })
	})
	var out OceanLoadingBLQ
	for r := 0; r < 3; r++ {
		for k := 0; k < 11; k++ {
			out.AmplitudeM[r][k] = float64(c.amplitude_m[r][k])
			out.PhaseDeg[r][k] = float64(c.phase_deg[r][k])
		}
	}
	runtime.KeepAlive(b)
	return out, err
}
func (b *BLQBlocks) Push(station string, c OceanLoadingBLQ) error {
	if b == nil || b.handle == nil {
		return ErrClosed
	}
	return b.handle.withExclusive(func(p unsafe.Pointer) error {
		return withString(station, func(s *C.char) uint32 {
			var v C.SidereonOceanLoadingBlq
			for r := 0; r < 3; r++ {
				for k := 0; k < 11; k++ {
					v.amplitude_m[r][k] = C.double(c.AmplitudeM[r][k])
					v.phase_deg[r][k] = C.double(c.PhaseDeg[r][k])
				}
			}
			return uint32(C.sidereon_blq_blocks_push((*C.SidereonBlqBlocks)(p), s, &v))
		})
	})
}
func (b *BLQBlocks) CommentCount(i int) (int, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	idx, err := checkedNativeSize(i)
	if err != nil {
		return 0, err
	}
	var n C.size_t
	err = b.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_blq_blocks_comment_count((*C.SidereonBlqBlocks)(p), idx, &n)) })
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(n, "BLQ comment count")
}
func (b *BLQBlocks) Comment(i, j int) (BLQComment, error) {
	if b == nil || b.handle == nil {
		return BLQComment{}, ErrClosed
	}
	idx, err := checkedNativeSize(i)
	if err != nil {
		return BLQComment{}, err
	}
	ci, err := checkedNativeSize(j)
	if err != nil {
		return BLQComment{}, err
	}
	var out BLQComment
	err = b.handle.with(func(p unsafe.Pointer) error {
		var raw C.SidereonBlqComment
		var text []byte
		text, err = copyNativeBytes("BLQ retained comment", func(x *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_blq_blocks_comment((*C.SidereonBlqBlocks)(p), idx, ci, &raw, x, n, w, r)
		})
		if err == nil {
			out = BLQComment{Placement: uint32(raw.placement), Text: string(text)}
			out.Row, err = sizeTToInt(raw.row, "BLQ comment row")
		}
		return err
	})
	runtime.KeepAlive(b)
	return out, err
}
func (b *BLQBlocks) PushComment(i int, c BLQComment) error {
	if b == nil || b.handle == nil {
		return ErrClosed
	}
	idx, err := checkedNativeSize(i)
	if err != nil {
		return err
	}
	row, err := checkedNativeSize(c.Row)
	if err != nil {
		return err
	}
	return b.handle.withExclusive(func(p unsafe.Pointer) error {
		return withString(c.Text, func(s *C.char) uint32 {
			raw := C.SidereonBlqComment{placement: C.uint32_t(c.Placement), row: row}
			return uint32(C.sidereon_blq_blocks_push_comment((*C.SidereonBlqBlocks)(p), idx, &raw, s))
		})
	})
}
func (b *BLQBlocks) Encode() ([]byte, BLQOutcome, error) {
	if b == nil || b.handle == nil {
		return nil, BLQOutcome{}, ErrClosed
	}
	var text []byte
	var outcome BLQOutcome
	err := b.handle.with(func(p unsafe.Pointer) error {
		var result *C.SidereonBlqResult
		if e := callStatus(func() uint32 { return uint32(C.sidereon_blq_blocks_to_text_result((*C.SidereonBlqBlocks)(p), &result)) }); e != nil {
			return e
		}
		if result == nil {
			return errNilNativeHandle
		}
		defer C.sidereon_blq_result_free(result)
		var e error
		outcome, e = blqResultErrorLocked(result)
		if e != nil || !outcome.IsOK {
			return e
		}
		text, e = copyNativeBytes("BLQ text", func(x *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_blq_result_get_text(result, x, n, w, r)
		})
		return e
	})
	runtime.KeepAlive(b)
	return text, outcome, err
}
func (b *BLQBlocks) BlockToText(i int) ([]byte, BLQOutcome, error) {
	if b == nil || b.handle == nil {
		return nil, BLQOutcome{}, ErrClosed
	}
	idx, err := checkedNativeSize(i)
	if err != nil {
		return nil, BLQOutcome{}, err
	}
	var text []byte
	var outcome BLQOutcome
	err = b.handle.with(func(p unsafe.Pointer) error {
		var result *C.SidereonBlqResult
		if e := callStatus(func() uint32 {
			return uint32(C.sidereon_blq_block_to_text_result((*C.SidereonBlqBlocks)(p), idx, &result))
		}); e != nil {
			return e
		}
		if result == nil {
			return errNilNativeHandle
		}
		defer C.sidereon_blq_result_free(result)
		var e error
		outcome, e = blqResultErrorLocked(result)
		if e != nil || !outcome.IsOK {
			return e
		}
		text, e = copyNativeBytes("BLQ block text", func(x *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_blq_result_get_text(result, x, n, w, r)
		})
		return e
	})
	runtime.KeepAlive(b)
	return text, outcome, err
}
