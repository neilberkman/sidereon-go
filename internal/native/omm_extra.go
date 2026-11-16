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
	"unsafe"
)

type OMM struct {
	_      noCopy
	handle *positioningHandle
}

func releaseOMM(pointer unsafe.Pointer) {
	C.sidereon_omm_free((*C.SidereonOmm)(pointer))
}

func newOMM(pointer *C.SidereonOmm) (*OMM, error) {
	if pointer == nil {
		return nil, missingNativeHandle("OMM parse")
	}
	return &OMM{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseOMM)}, nil
}

func parseOMM(data []byte, kind uint32) (*OMM, error) {
	if _, err := checkedNativeSize(len(data)); err != nil {
		return nil, err
	}
	copyData := append([]byte(nil), data...)
	var pointer unsafe.Pointer
	if len(copyData) != 0 {
		pointer = C.CBytes(copyData)
		if pointer == nil {
			return nil, errors.New("sidereon: unable to allocate native OMM input")
		}
		defer C.free(pointer)
	}
	var out *C.SidereonOmm
	err := callStatus(func() uint32 {
		switch kind {
		case 0:
			return C.sidereon_omm_parse_json((*C.uint8_t)(pointer), C.size_t(len(copyData)), &out)
		case 1:
			return C.sidereon_omm_parse_kvn((*C.uint8_t)(pointer), C.size_t(len(copyData)), &out)
		default:
			return C.sidereon_omm_parse_xml((*C.uint8_t)(pointer), C.size_t(len(copyData)), &out)
		}
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_omm_free(out) })
		}
		return nil, err
	}
	return newOMM(out)
}

func ParseOMMJSON(data []byte) (*OMM, error) { return parseOMM(data, 0) }
func ParseOMMKVN(data []byte) (*OMM, error)  { return parseOMM(data, 1) }
func ParseOMMXML(data []byte) (*OMM, error)  { return parseOMM(data, 2) }

func (o *OMM) Close() error {
	if o == nil || o.handle == nil {
		return nil
	}
	return o.handle.close()
}

func (o *OMM) encode(kind uint32) ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	err := o.handle.with(func(pointer unsafe.Pointer) error {
		invoke := func(out *C.uint8_t, length C.size_t, written, required *C.size_t) uint32 {
			switch kind {
			case 0:
				return C.sidereon_omm_to_json((*C.SidereonOmm)(pointer), out, length, written, required)
			case 1:
				return C.sidereon_omm_to_kvn((*C.SidereonOmm)(pointer), out, length, written, required)
			case 3:
				return C.sidereon_omm_to_kvn_lenient((*C.SidereonOmm)(pointer), out, length, written, required)
			case 4:
				return C.sidereon_omm_to_xml_lenient((*C.SidereonOmm)(pointer), out, length, written, required)
			default:
				return C.sidereon_omm_to_xml((*C.SidereonOmm)(pointer), out, length, written, required)
			}
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 { return invoke(nil, 0, &written, &required) }); err != nil {
			return err
		}
		count, err := validateNativeQuery("OMM encoding", uint64(written), uint64(required))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.uint8_t(0))); err != nil {
			return err
		}
		values := make([]C.uint8_t, count)
		var out *C.uint8_t
		if count != 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(out, C.size_t(count), &written, &required) }); err != nil {
			return err
		}
		n, err := validateTwoPassCounts("OMM encoding", count, count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]byte, n)
		for i := range result {
			result[i] = byte(values[i])
		}
		return nil
	})
	return result, err
}

func (o *OMM) JSON() ([]byte, error)       { return o.encode(0) }
func (o *OMM) KVN() ([]byte, error)        { return o.encode(1) }
func (o *OMM) XML() ([]byte, error)        { return o.encode(2) }
func (o *OMM) KVNLenient() ([]byte, error) { return o.encode(3) }
func (o *OMM) XMLLenient() ([]byte, error) { return o.encode(4) }

func OMMFromTLE(tle *TLE, name string) (*OMM, error) {
	if tle == nil || tle.handle == nil {
		return nil, ErrClosed
	}
	input, err := copyNativeInput([]byte(name))
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(input)
	var out *C.SidereonOmm
	err = tle.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_omm_from_tle((*C.SidereonTle)(p), (*C.uint8_t)(input), C.size_t(len(name)), &out))
		})
	})
	if err != nil {
		return nil, err
	}
	return newOMM(out)
}

func (o *OMM) WithCCSDSHeader(version, creationDate, originator string) (*OMM, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	a, err := copyNativeInput([]byte(version))
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(a)
	b, err := copyNativeInput([]byte(creationDate))
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(b)
	c, err := copyNativeInput([]byte(originator))
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(c)
	var out *C.SidereonOmm
	err = o.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_omm_with_ccsds_header((*C.SidereonOmm)(p), (*C.uint8_t)(a), C.size_t(len(version)), (*C.uint8_t)(b), C.size_t(len(creationDate)), (*C.uint8_t)(c), C.size_t(len(originator)), &out))
		})
	})
	if err != nil {
		return nil, err
	}
	return newOMM(out)
}

func (o *OMM) Equal(other *OMM) (bool, error) {
	if o == nil || o.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	var out C.bool
	err := withPositioningPair(o.handle, other.handle, func(a, b unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_omm_equal((*C.SidereonOmm)(a), (*C.SidereonOmm)(b), &out)) })
	})
	return bool(out), err
}
