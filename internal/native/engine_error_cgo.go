//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
)

// currentEngineErrorLocked queries and copies the retained generic schema1
// engine error on the caller's locked OS thread. Uses bare status conversion to
// avoid recursive generic error capture if secondary diagnostic calls fail.
func currentEngineErrorLocked() (*EngineError, error) {
	var info C.SidereonEngineErrorInfo
	infoStatus := uint32(C.sidereon_last_engine_error_info(&info))
	if infoErr := bareStatusErrorLocked(infoStatus); infoErr != nil {
		return &EngineError{
			Family:       EngineErrorFamilyUnknown,
			FamilyName:   "unknown",
			CaptureError: fmt.Errorf("sidereon: engine error info query failed: %w", infoErr),
		}, infoErr
	}

	family := EngineErrorFamily(info.family)
	payloadLen := uint64(info.payload_len)

	if family == EngineErrorFamilyNone {
		if payloadLen == 0 {
			return nil, nil
		}
		// Family None requires length 0. Non-loss policy retains raw payload with diagnostic.
	}

	var preCaptureErr error
	if family == EngineErrorFamilyNone && payloadLen > 0 {
		preCaptureErr = errors.New("sidereon: family none reported non-zero payload length")
	}

	payload, payloadErr := copyNativeBytesLockedWithStatus(
		"engine error payload",
		func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_last_engine_error_payload(out, capacity, written, required)
		},
		bareStatusErrorLocked,
	)
	if payloadErr != nil {
		err := fmt.Errorf("sidereon: engine error payload capture failed: %w", payloadErr)
		if preCaptureErr != nil {
			err = errors.Join(preCaptureErr, err)
		}
		return &EngineError{
			Family:       family,
			FamilyName:   family.Name(),
			CaptureError: err,
		}, err
	}

	expectedLength, lengthErr := sizeTToInt(info.payload_len, "engine error payload length")
	if lengthErr == nil && expectedLength != len(payload) {
		lengthErr = fmt.Errorf("sidereon: engine error payload length mismatch: summary=%d payload=%d", expectedLength, len(payload))
	}
	if lengthErr != nil {
		if preCaptureErr != nil {
			lengthErr = errors.Join(preCaptureErr, lengthErr)
		}
		return &EngineError{
			Family:       family,
			FamilyName:   family.Name(),
			Payload:      append(json.RawMessage(nil), payload...),
			CaptureError: lengthErr,
		}, lengthErr
	}

	decoded, decodeErr := DecodeSchema1EnginePayload(family, payload)
	if preCaptureErr != nil {
		if decoded.CaptureError != nil {
			decoded.CaptureError = errors.Join(preCaptureErr, decoded.CaptureError)
		} else {
			decoded.CaptureError = preCaptureErr
		}
	}
	return decoded, decodeErr
}

// ClearEngineError explicitly clears the versioned generic engine error slot on
// the caller's thread.
func ClearEngineError() {
	withCThread(func() {
		C.sidereon_clear_engine_error()
	})
}

// ClearEngineErrorLocked clears the generic engine error slot on the currently locked C thread.
func ClearEngineErrorLocked() {
	C.sidereon_clear_engine_error()
}

// CurrentEngineErrorLocked queries the retained generic engine error on the caller's locked OS thread.
func CurrentEngineErrorLocked() (*EngineError, error) {
	return currentEngineErrorLocked()
}
