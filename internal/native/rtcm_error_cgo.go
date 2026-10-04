//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#if SIDEREON_VERSION_MAJOR < 3
typedef enum SidereonRtcmErrorClass {
    SIDEREON_RTCM_ERROR_CLASS_NONE = 0,
    SIDEREON_RTCM_ERROR_CLASS_ENCODE = 1,
    SIDEREON_RTCM_ERROR_CLASS_CONVERSION = 2,
    SIDEREON_RTCM_ERROR_CLASS_OTHER = 3,
    SIDEREON_RTCM_ERROR_CLASS_SBAS_ENCODE = 4
} SidereonRtcmErrorClass;
typedef struct SidereonRtcmErrorInfo {
    SidereonRtcmErrorClass class_;
    uint32_t kind;
    size_t payload_len;
} SidereonRtcmErrorInfo;
enum SidereonStatus sidereon_rtcm_last_error_info(SidereonRtcmErrorInfo *out);
enum SidereonStatus sidereon_rtcm_last_error_payload(uint8_t *out, size_t len, size_t *out_written, size_t *out_required);
#endif
*/
import "C"

import (
	"encoding/json"
	"unsafe"
)

func rtcmErrorLocked(status uint32) error {
	err := statusErrorLocked(status)
	if err == nil {
		return nil
	}
	statusErr, ok := err.(*StatusError)
	if !ok {
		return err
	}
	var info C.SidereonRtcmErrorInfo
	if readErr := statusErrorLocked(uint32(C.sidereon_rtcm_last_error_info(&info))); readErr != nil {
		return statusErr
	}
	class := uint32(info.class_)
	if class == 0 {
		return statusErr
	}
	payloadLength, lengthErr := sizeTToInt(info.payload_len, "RTCM error payload")
	if lengthErr != nil {
		return statusErr
	}
	payload, payloadErr := copyNativeBytesLocked("RTCM error payload", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rtcm_last_error_payload(out, length, written, required)
	})
	if payloadErr != nil || len(payload) != payloadLength || !json.Valid(payload) {
		return statusErr
	}
	statusErr.RTCM = &RTCMError{Class: class, Kind: uint32(info.kind), Payload: append([]byte(nil), payload...)}
	return statusErr
}

func rtcmCallStatus(fn func() uint32) error {
	var err error
	withCThread(func() { err = rtcmErrorLocked(fn()) })
	return err
}

func rtcmInputError(data []byte, fn func(*C.uint8_t, C.size_t) uint32) error {
	return withInputStatus(data, fn, rtcmErrorLocked)
}

func copyRTCMByteOutput(label string, call func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t) uint32) ([]byte, error) {
	var written, required C.size_t
	if err := rtcmCallStatus(func() uint32 { return call(nil, 0, &written, &required) }); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(n, 1); err != nil {
		return nil, err
	}
	buffer := make([]byte, n)
	var output *C.uint8_t
	if n != 0 {
		output = (*C.uint8_t)(unsafe.Pointer(&buffer[0]))
	}
	if err := rtcmCallStatus(func() uint32 { return call(output, C.size_t(n), &written, &required) }); err != nil {
		return nil, err
	}
	w, err := validateNativeOutput(label, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), buffer[:w]...), nil
}
