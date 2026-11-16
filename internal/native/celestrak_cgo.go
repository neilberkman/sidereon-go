//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
*/
import "C"
import "unsafe"

func CelestrakText(operation, kind uint32, value string, number uint64, format uint32) (string, error) {
	input, err := copyNativeInput([]byte(value))
	if err != nil {
		return "", err
	}
	defer freeNativeInput(input)
	var result []byte
	withCThread(func() {
		result, err = callCatalogString(func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			switch operation {
			case 0:
				return C.sidereon_celestrak_gp_source_entry(out, length, written, required)
			case 1:
				return C.sidereon_navcen_source_entry(out, length, written, required)
			case 2:
				return C.sidereon_navcen_constellation_url(out, length, written, required)
			case 3:
				return C.sidereon_celestrak_group(C.uint32_t(kind), out, length, written, required)
			case 4:
				return C.sidereon_celestrak_constellation_url(C.uint32_t(kind), C.uint32_t(format), out, length, written, required)
			case 5:
				return C.sidereon_celestrak_gp_url(C.uint32_t(kind), (*C.uint8_t)(input), C.size_t(len(value)), C.uint32_t(format), out, length, written, required)
			case 6:
				return C.sidereon_celestrak_gp_catalog_number_url(C.uint64_t(number), C.uint32_t(format), out, length, written, required)
			case 7:
				return C.sidereon_celestrak_gp_special_url(C.uint32_t(kind), C.uint32_t(format), out, length, written, required)
			case 10:
				return C.sidereon_data_allowed_hosts(out, length, written, required)
			case 8:
				return C.sidereon_celestrak_gp_format_as_str(C.uint32_t(format), out, length, written, required)
			default:
				return C.sidereon_celestrak_gp_special_as_str(C.uint32_t(kind), out, length, written, required)
			}
		})
	})
	return string(result), err
}

type CelestrakGpResponse struct{ handle *positioningHandle }

func ParseCelestrakGpResponse(data []byte, format uint32) (*CelestrakGpResponse, error) {
	input, err := copyNativeInput(data)
	if err != nil {
		return nil, err
	}
	defer freeNativeInput(input)
	var out *C.SidereonCelestrakGpResponse
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_parse_celestrak_gp_response((*C.uint8_t)(input), C.size_t(len(data)), C.uint32_t(format), &out))
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_celestrak_gp_response_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("GP response")
	}
	return &CelestrakGpResponse{newPositioningHandle(unsafe.Pointer(out), func(p unsafe.Pointer) { C.sidereon_celestrak_gp_response_free((*C.SidereonCelestrakGpResponse)(p)) })}, nil
}
func (r *CelestrakGpResponse) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return r.handle.close()
}
func (r *CelestrakGpResponse) Counts() (int, int, error) {
	if r == nil || r.handle == nil {
		return 0, 0, ErrClosed
	}
	var a, b C.size_t
	err := r.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_celestrak_gp_response_counts((*C.SidereonCelestrakGpResponse)(p), &a, &b))
		})
	})
	if err != nil {
		return 0, 0, err
	}
	x, err := sizeTToInt(a, "GP count")
	if err != nil {
		return 0, 0, err
	}
	y, err := sizeTToInt(b, "GP rejection count")
	return x, y, err
}
func (r *CelestrakGpResponse) RecordBytes(index int, rejected, raw bool) ([]byte, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("negative GP index")
	}
	var result []byte
	err := r.handle.with(func(p unsafe.Pointer) error {
		var e error
		withCThread(func() {
			result, e = callCatalogString(func(out *C.uint8_t, n C.size_t, w, q *C.size_t) C.enum_SidereonStatus {
				if raw {
					return C.sidereon_celestrak_gp_response_raw((*C.SidereonCelestrakGpResponse)(p), C.bool(rejected), C.size_t(index), out, n, w, q)
				}
				return C.sidereon_celestrak_gp_response_record_json((*C.SidereonCelestrakGpResponse)(p), C.bool(rejected), C.size_t(index), out, n, w, q)
			})
		})
		return e
	})
	return result, err
}
func (r *CelestrakGpResponse) RecordOMM(index int) (*OMM, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("negative GP index")
	}
	var out *C.SidereonOmm
	err := r.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_celestrak_gp_record_to_omm((*C.SidereonCelestrakGpResponse)(p), C.size_t(index), &out))
		})
	})
	if err != nil {
		return nil, err
	}
	return newOMM(out)
}

func (r *CelestrakGpResponse) Equal(other *CelestrakGpResponse) (bool, error) {
	if r == nil || r.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	var out C.bool
	err := withPositioningPair(r.handle, other.handle, func(a, b unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_celestrak_gp_response_equal((*C.SidereonCelestrakGpResponse)(a), (*C.SidereonCelestrakGpResponse)(b), &out))
		})
	})
	return bool(out), err
}

func (r *CelestrakGpResponse) RecordTLE(index int, opsmode, policy uint32) (*TLE, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("negative GP index")
	}
	if opsmode != TLEOpsModeAFSPCValue && opsmode != TLEOpsModeImprovedValue {
		return nil, invalidArgument("TLE operations mode is not defined by the C ABI")
	}
	if policy != TLEFilePolicyStrictValue && policy != TLEFilePolicyLenientValue {
		return nil, invalidArgument("TLE checksum policy is not defined by the C ABI")
	}
	var out *C.SidereonTle
	err := r.handle.with(func(p unsafe.Pointer) error {
		return callSGP4Status(func() uint32 {
			return uint32(C.sidereon_celestrak_gp_record_tle((*C.SidereonCelestrakGpResponse)(p), C.size_t(index), C.uint32_t(opsmode), C.uint32_t(policy), &out))
		})
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_tle_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("GP TLE")
	}
	return &TLE{handle: newPositioningHandle(unsafe.Pointer(out), releaseTLE)}, nil
}

func (r *CelestrakGpResponse) RecordElementSet(index int) (OMMElementSet, error) {
	if r == nil || r.handle == nil {
		return OMMElementSet{}, ErrClosed
	}
	if index < 0 {
		return OMMElementSet{}, invalidArgument("negative GP index")
	}
	var out C.SidereonOmmElementSet
	err := r.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_celestrak_gp_record_to_element_set((*C.SidereonCelestrakGpResponse)(p), C.size_t(index), &out))
		})
	})
	if err != nil {
		return OMMElementSet{}, err
	}
	return ommElementSetFromC(out), nil
}

func (r *CelestrakGpResponse) CompareRecord(other *CelestrakGpResponse, index, otherIndex int, mode uint32) (bool, error) {
	if r == nil || r.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	if index < 0 || otherIndex < 0 {
		return false, invalidArgument("negative GP index")
	}
	var out C.bool
	err := withPositioningPair(r.handle, other.handle, func(a, b unsafe.Pointer) error {
		return callStatus(func() uint32 {
			x, y := (*C.SidereonCelestrakGpResponse)(a), (*C.SidereonCelestrakGpResponse)(b)
			i, j := C.size_t(index), C.size_t(otherIndex)
			switch mode {
			case 0:
				return uint32(C.sidereon_celestrak_gp_record_equal(x, y, false, i, j, &out))
			case 1:
				return uint32(C.sidereon_celestrak_gp_record_equal(x, y, true, i, j, &out))
			case 2:
				return uint32(C.sidereon_celestrak_gp_payload_equal(x, y, i, j, &out))
			case 3:
				return uint32(C.sidereon_celestrak_tle_record_equal(x, y, i, j, &out))
			case 4:
				return uint32(C.sidereon_celestrak_gp_rejection_reason_equal(x, y, i, j, &out))
			default:
				return uint32(C.sidereon_celestrak_gp_structural_error_equal(x, y, i, j, &out))
			}
		})
	})
	return bool(out), err
}
