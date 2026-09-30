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

// OMMArray owns successful OMM clones and typed per-record skip payloads.
type OMMArray struct {
	_      noCopy
	handle *positioningHandle
}

func releaseOMMArray(pointer unsafe.Pointer) {
	C.sidereon_omm_array_free((*C.SidereonOmmArray)(pointer))
}

func newOMMArray(pointer *C.SidereonOmmArray) (*OMMArray, error) {
	if pointer == nil {
		return nil, missingNativeHandle("OMM array")
	}
	return &OMMArray{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseOMMArray)}, nil
}

func parseOMMArray(data []byte, kind uint32) (*OMMArray, error) {
	input, err := copyNativeInput(data)
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(input)
	var out *C.SidereonOmmArray
	err = callStatus(func() uint32 {
		switch kind {
		case 0:
			return uint32(C.sidereon_omm_parse_json_array((*C.uint8_t)(input), C.size_t(len(data)), &out))
		case 1:
			return uint32(C.sidereon_omm_parse_xml_all((*C.uint8_t)(input), C.size_t(len(data)), &out))
		default:
			return uint32(C.sidereon_omm_parse_csv_array((*C.uint8_t)(input), C.size_t(len(data)), &out))
		}
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_omm_array_free(out) })
		}
		return nil, err
	}
	return newOMMArray(out)
}

// ParseOMMJSONArray parses a JSON OMM object or array.
func ParseOMMJSONArray(data []byte) (*OMMArray, error) { return parseOMMArray(data, 0) }

// ParseOMMXMLAll parses all OMM messages in an NDM XML document.
func ParseOMMXMLAll(data []byte) (*OMMArray, error) { return parseOMMArray(data, 1) }

// ParseOMMCSVArray parses a GP CSV table into successful and skipped rows.
func ParseOMMCSVArray(data []byte) (*OMMArray, error) { return parseOMMArray(data, 2) }

func NewOMMArray() (*OMMArray, error) {
	var out *C.SidereonOmmArray
	if err := callStatus(func() uint32 { return uint32(C.sidereon_omm_array_new(&out)) }); err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_omm_array_free(out) })
		}
		return nil, err
	}
	return newOMMArray(out)
}

// Close releases the owned collection; repeated calls are safe.
func (a *OMMArray) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}

// Append clones an OMM into the owned array.
func (a *OMMArray) Append(omm *OMM) error {
	if a == nil || a.handle == nil || omm == nil || omm.handle == nil {
		return ErrClosed
	}
	// Lock both resources while the native call clones the immutable OMM value.
	err := a.handle.withExclusive(func(arrayPointer unsafe.Pointer) error {
		return omm.handle.with(func(ommPointer unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_omm_array_push((*C.SidereonOmmArray)(arrayPointer), (*C.SidereonOmm)(ommPointer)))
			})
		})
	})
	runtime.KeepAlive(a)
	runtime.KeepAlive(omm)
	return err
}

// Count returns the number of successful records.
func (a *OMMArray) Count() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_omm_array_count((*C.SidereonOmmArray)(pointer), &count)) })
	})
	runtime.KeepAlive(a)
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "OMM array count")
}

// Record returns a detached OMM clone at index.
func (a *OMMArray) Record(index int) (*OMM, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("OMM record index is negative")
	}
	var out *C.SidereonOmm
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		i, e := cSize(index, "OMM record index")
		if e != nil {
			return e
		}
		return callStatus(func() uint32 { return uint32(C.sidereon_omm_array_record((*C.SidereonOmmArray)(pointer), i, &out)) })
	})
	runtime.KeepAlive(a)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_omm_free(out) })
		}
		return nil, err
	}
	return newOMM(out)
}

// Skipped returns detached original indexes and typed error payloads.
func (a *OMMArray) Skipped() ([]OMMSkippedRecord, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	var records []OMMSkippedRecord
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return uint32(C.sidereon_omm_array_skipped_count((*C.SidereonOmmArray)(pointer), &count))
		}); err != nil {
			return err
		}
		n, err := sizeTToInt(count, "OMM skipped record count")
		if err != nil {
			return err
		}
		records = make([]OMMSkippedRecord, n)
		for i := range records {
			var index C.size_t
			payload, err := copyNativeBytesLocked("OMM skipped error", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				j, convErr := cSize(i, "OMM skipped index")
				if convErr != nil {
					return C.enum_SidereonStatus(C.SIDEREON_STATUS_INVALID_ARGUMENT)
				}
				status := C.sidereon_omm_array_skipped((*C.SidereonOmmArray)(pointer), j, &index, out, length, written, required)
				return status
			})
			if err != nil {
				return err
			}
			records[i] = OMMSkippedRecord{Index: uint64(index), Payload: append([]byte(nil), payload...)}
		}
		return nil
	})
	runtime.KeepAlive(a)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (a *OMMArray) encode(kind uint32) ([]byte, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	var result []byte
	var err error
	err = a.handle.with(func(pointer unsafe.Pointer) error {
		result, err = copyNativeBytesLocked("OMM array encoding", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			switch kind {
			case 0:
				return C.sidereon_omm_array_to_json((*C.SidereonOmmArray)(pointer), out, length, written, required)
			case 1:
				return C.sidereon_omm_array_to_json_discarding_comments((*C.SidereonOmmArray)(pointer), out, length, written, required)
			case 2:
				return C.sidereon_omm_array_to_csv((*C.SidereonOmmArray)(pointer), out, length, written, required)
			default:
				return C.sidereon_omm_array_to_csv_discarding_comments((*C.SidereonOmmArray)(pointer), out, length, written, required)
			}
		})
		return err
	})
	runtime.KeepAlive(a)
	return result, err
}
func (a *OMMArray) JSON() ([]byte, error)                   { return a.encode(0) }
func (a *OMMArray) JSONDiscardingComments() ([]byte, error) { return a.encode(1) }
func (a *OMMArray) CSV() ([]byte, error)                    { return a.encode(2) }
func (a *OMMArray) CSVDiscardingComments() ([]byte, error)  { return a.encode(3) }

// ParseOMM invokes the core format autodetector.
func ParseOMM(data []byte) (*OMM, error) { return parseOMMKind(data, 3) }

// ParseOMMCSV parses one GP CSV data record.
func ParseOMMCSV(data []byte) (*OMM, error) { return parseOMMKind(data, 4) }

func parseOMMKind(data []byte, kind uint32) (*OMM, error) {
	input, err := copyNativeInput(data)
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(input)
	var out *C.SidereonOmm
	err = callStatus(func() uint32 {
		if kind == 3 {
			return uint32(C.sidereon_omm_parse((*C.uint8_t)(input), C.size_t(len(data)), &out))
		}
		return uint32(C.sidereon_omm_parse_csv((*C.uint8_t)(input), C.size_t(len(data)), &out))
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_omm_free(out) })
		}
		return nil, err
	}
	return newOMM(out)
}

// ParseOMMEpoch parses an OMM civil epoch with femtosecond precision.
func ParseOMMEpoch(text string) (OMMEpoch, error) {
	input, err := copyNativeInput([]byte(text))
	if err != nil {
		return OMMEpoch{}, err
	}
	defer freeNativeInput(input)
	var value C.SidereonOmmEpoch
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_omm_parse_epoch((*C.uint8_t)(input), C.size_t(len(text)), &value))
	})
	if err != nil {
		return OMMEpoch{}, err
	}
	return OMMEpoch{Year: int32(value.year), Month: uint32(value.month), Day: uint32(value.day), Hour: uint32(value.hour), Minute: uint32(value.minute), Second: uint32(value.second), Microsecond: uint32(value.microsecond), Femtosecond: uint32(value.femtosecond)}, nil
}
