//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
enum SidereonStatus sidereon_bias_set_notice_count(const struct SidereonBiasSet *, size_t *);
enum SidereonStatus sidereon_bias_set_notice(const struct SidereonBiasSet *, size_t, SidereonBiasNotice *);
enum SidereonStatus sidereon_bias_set_notice_text(const struct SidereonBiasSet *, size_t, uint32_t, uint8_t *, size_t, size_t *, size_t *);
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type BiasNotice struct {
	Kind           uint32
	Departure      uint32
	HasLine        bool
	Line           int
	First          int
	Second         int
	DeclaredCount  uint64
	SolutionRows   int
	BiasMode       uint32
	UnknownVariant string
}

func (s *BiasSet) NoticeCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			operationErr = biasErrorLocked(uint32(C.sidereon_bias_set_notice_count((*C.SidereonBiasSet)(pointer), &count)))
		})
		return operationErr
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "bias notice count")
}

func (s *BiasSet) Notice(index int) (BiasNotice, error) {
	if s == nil || s.handle == nil {
		return BiasNotice{}, ErrClosed
	}
	nativeIndex, err := cSize(index, "bias notice index")
	if err != nil {
		return BiasNotice{}, err
	}
	var value C.SidereonBiasNotice
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			err = biasErrorLocked(uint32(C.sidereon_bias_set_notice((*C.SidereonBiasSet)(pointer), nativeIndex, &value)))
		})
		return err
	})
	if err != nil {
		return BiasNotice{}, err
	}
	line, err := sizeTToInt(value.line, "bias notice line")
	if err != nil {
		return BiasNotice{}, err
	}
	first, err := sizeTToInt(value.first, "bias notice first index")
	if err != nil {
		return BiasNotice{}, err
	}
	second, err := sizeTToInt(value.second, "bias notice second index")
	if err != nil {
		return BiasNotice{}, err
	}
	rows, err := sizeTToInt(value.solution_rows, "bias notice solution rows")
	if err != nil {
		return BiasNotice{}, err
	}
	unknown := C.GoString(&value.unknown_variant[0])
	return BiasNotice{Kind: uint32(value.kind), Departure: uint32(value.departure), HasLine: bool(value.has_line), Line: line, First: first, Second: second, DeclaredCount: uint64(value.declared_count), SolutionRows: rows, BiasMode: uint32(value.bias_mode), UnknownVariant: unknown}, nil
}

func (s *BiasSet) NoticeText(index int, part uint32) (string, error) {
	if s == nil || s.handle == nil {
		return "", ErrClosed
	}
	nativeIndex, err := cSize(index, "bias notice index")
	if err != nil {
		return "", err
	}
	var result string
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var written, required C.size_t
			status := C.sidereon_bias_set_notice_text((*C.SidereonBiasSet)(pointer), nativeIndex, C.uint32_t(part), nil, 0, &written, &required)
			if err = biasErrorLocked(uint32(status)); err != nil {
				return
			}
			length, conversionErr := sizeTToInt(required, "bias notice text length")
			if conversionErr != nil {
				err = conversionErr
				return
			}
			if length == 0 {
				result = ""
				return
			}
			if _, conversionErr = checkedNativeAllocationSize(length, 1); conversionErr != nil {
				err = conversionErr
				return
			}
			buffer := make([]byte, length)
			status = C.sidereon_bias_set_notice_text((*C.SidereonBiasSet)(pointer), nativeIndex, C.uint32_t(part), (*C.uint8_t)(unsafe.Pointer(&buffer[0])), C.size_t(length), &written, &required)
			if err = biasErrorLocked(uint32(status)); err != nil {
				return
			}
			writtenCount, countErr := validateTwoPassCounts("bias notice text", length, length, uint64(written), uint64(required))
			if countErr != nil {
				err = countErr
				return
			}
			result = string(buffer[:writtenCount])
		})
		return err
	})
	runtime.KeepAlive(s)
	if err != nil {
		return "", err
	}
	return result, nil
}
