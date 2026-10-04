//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import "fmt"

// statusSp3ErrorLocked captures the typed SP3 summary and payload immediately
// after its producer, while still on the producer's locked OS thread.
func statusSp3ErrorLocked(status uint32) error {
	err := statusErrorLocked(status)
	if err == nil {
		return nil
	}
	statusErr, ok := err.(*StatusError)
	if !ok {
		return err
	}
	sp3, captureErr := currentSp3ErrorLocked()
	statusErr.SP3 = sp3
	if captureErr != nil {
		statusErr.Detail = appendDiagnostic(statusErr.Detail, "SP3 error capture failed: "+captureErr.Error())
		return statusErr
	}
	return statusErr
}

// currentSp3ErrorLocked copies the retaining SP3 diagnostic accessors on the
// same locked OS thread as the producer.
func currentSp3ErrorLocked() (*SP3Error, error) {
	var info C.SidereonSp3ErrorInfo
	infoStatus := uint32(C.sidereon_sp3_last_error_info(&info))
	if infoErr := statusErrorLocked(infoStatus); infoErr != nil {
		return nil, infoErr
	}
	if uint32(info.kind) == 0 {
		return nil, nil
	}
	sp3 := &SP3Error{
		Kind:               uint32(info.kind),
		Field:              uint32(info.field),
		Reason:             uint32(info.reason),
		HasValue:           bool(info.has_value),
		Value:              float64(info.value),
		HasRequestedTick:   bool(info.has_requested_tick),
		HasDeclaredTick:    bool(info.has_declared_tick),
		HasRequestedJ2000S: bool(info.has_requested_j2000_s),
		HasDeclaredJ2000S:  bool(info.has_declared_j2000_s),
		RequestedJ2000S:    float64(info.requested_j2000_s),
		DeclaredJ2000S:     float64(info.declared_j2000_s),
	}
	payload, payloadErr := copyNativeBytesLockedWithStatus(
		"SP3 error payload",
		func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_sp3_last_error_payload(out, capacity, written, required)
		},
		statusErrorLocked,
	)
	if payloadErr != nil {
		return sp3, fmt.Errorf("SP3 error payload capture failed: %w", payloadErr)
	}
	payloadLength, lengthErr := sizeTToInt(info.payload_len, "SP3 error payload length")
	if lengthErr == nil && payloadLength != len(payload) {
		lengthErr = fmt.Errorf("SP3 error payload length mismatch: summary=%d payload=%d", payloadLength, len(payload))
	}
	if lengthErr != nil {
		return sp3, fmt.Errorf("SP3 error payload capture failed: %w", lengthErr)
	}
	sp3.PayloadJSON = string(payload)
	return sp3, nil
}

func appendDiagnostic(current, extra string) string {
	if current == "" {
		return extra
	}
	return current + "; " + extra
}
