//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

func (m *RtcmMessages) EncodeWithPolicy(index int, policy uint32) ([]byte, []NativeRTCMDeparture, error) {
	return m.bytesWithPolicy(index, policy, false)
}

func (m *RtcmMessages) FrameWithPolicy(index int, policy uint32) ([]byte, []NativeRTCMDeparture, error) {
	return m.bytesWithPolicy(index, policy, true)
}

func (m *RtcmMessages) bytesWithPolicy(index int, policy uint32, frame bool) ([]byte, []NativeRTCMDeparture, error) {
	if index < 0 {
		return nil, nil, errNegativeIndex
	}
	if policy > 1 {
		return nil, nil, invalidArgument("invalid RTCM policy")
	}
	var output []byte
	var departures []NativeRTCMDeparture
	err := m.resource.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			idx, err := checkedNativeSize(index)
			if err != nil {
				return err
			}
			fn := func(out *C.uint8_t, outLen C.size_t, outWritten, outRequired *C.size_t, dep *C.SidereonRtcmDeparture, depCap C.size_t, depWritten, depRequired *C.size_t) C.enum_SidereonStatus {
				if frame {
					return C.sidereon_rtcm_message_to_frame_with_policy((*C.SidereonRtcmMessages)(p), idx, C.uint32_t(policy), out, outLen, outWritten, outRequired, dep, depCap, depWritten, depRequired)
				}
				return C.sidereon_rtcm_message_encode_with_policy((*C.SidereonRtcmMessages)(p), idx, C.uint32_t(policy), out, outLen, outWritten, outRequired, dep, depCap, depWritten, depRequired)
			}
			var written, required, depWritten, depRequired C.size_t
			status := fn(nil, 0, &written, &required, nil, 0, &depWritten, &depRequired)
			if status != C.SIDEREON_STATUS_OK && (status != C.SIDEREON_STATUS_INVALID_ARGUMENT || written != 0 || required == 0) {
				return statusErrorLocked(uint32(status))
			}
			n, err := sizeTToInt(required, "RTCM encoded output length")
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, 1); err != nil {
				return err
			}
			output = make([]byte, n)
			var out *C.uint8_t
			if n > 0 {
				out = (*C.uint8_t)(unsafe.Pointer(&output[0]))
			}
			written, required, depWritten, depRequired = 0, 0, 0, 0
			status = fn(out, C.size_t(n), &written, &required, nil, 0, &depWritten, &depRequired)
			if status != C.SIDEREON_STATUS_OK && (status != C.SIDEREON_STATUS_INVALID_ARGUMENT || written != required || depWritten != 0 || depRequired == 0) {
				return statusErrorLocked(uint32(status))
			}
			if _, err := validateTwoPassCounts("RTCM encoded bytes", n, n, uint64(written), uint64(required)); err != nil {
				return err
			}
			depCount, err := sizeTToInt(depRequired, "RTCM departure count")
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(depCount, unsafe.Sizeof(C.SidereonRtcmDeparture{})); err != nil {
				return err
			}
			var depMem unsafe.Pointer
			if depCount > 0 {
				depMem = C.calloc(C.size_t(depCount), C.size_t(unsafe.Sizeof(C.SidereonRtcmDeparture{})))
				if depMem == nil {
					return errors.New("sidereon: unable to allocate RTCM departures")
				}
				defer C.free(depMem)
			}
			var depPtr *C.SidereonRtcmDeparture
			if depMem != nil {
				depPtr = (*C.SidereonRtcmDeparture)(depMem)
			}
			written, required, depWritten, depRequired = 0, 0, 0, 0
			status = fn(out, C.size_t(n), &written, &required, depPtr, C.size_t(depCount), &depWritten, &depRequired)
			if status != C.SIDEREON_STATUS_OK {
				return statusErrorLocked(uint32(status))
			}
			if _, err := validateTwoPassCounts("RTCM encoded bytes", n, n, uint64(written), uint64(required)); err != nil {
				return err
			}
			depWrittenInt, err := validateNativeOutput("RTCM departures", depCount, uint64(depWritten), uint64(depRequired))
			if err != nil {
				return err
			}
			var values []C.SidereonRtcmDeparture
			if depCount > 0 {
				values = unsafe.Slice((*C.SidereonRtcmDeparture)(depMem), depCount)
			}
			departures = make([]NativeRTCMDeparture, depWrittenInt)
			for j := range departures {
				v := values[j]
				fields := []struct {
					name   string
					value  C.size_t
					target *int
				}{
					{"layer index", v.layer_index, &departures[j].LayerIndex}, {"declared", v.declared, &departures[j].Declared}, {"read", v.read, &departures[j].Read}, {"cells", v.cells, &departures[j].Cells}, {"bit count", v.bit_count, &departures[j].BitCount},
				}
				for _, field := range fields {
					*field.target, err = checkedNativeCount(uint64(field.value))
					if err != nil {
						return err
					}
				}
				departures[j].Kind = uint32(v.kind)
				departures[j].MessageNumber = uint16(v.message_number)
				departures[j].Reserved = uint8(v.reserved)
				departures[j].Degree = uint8(v.degree)
				departures[j].Order = uint8(v.order)
			}
			output = output[:written]
			return nil
		})
	})
	runtime.KeepAlive(m)
	return output, departures, err
}

func DecodeRTCMStreamWithPolicy(data []byte, policy uint32) (*RtcmMessages, *RtcmDiagnostics, error) {
	if policy > 1 {
		return nil, nil, invalidArgument("invalid RTCM policy")
	}
	length, err := checkedNativeSize(len(data))
	if err != nil {
		return nil, nil, err
	}
	var messages *C.SidereonRtcmMessages
	var diagnostics *C.SidereonRtcmStreamDiagnostics
	var operationErr error
	operationErr = withCThreadError(func() error {
		var input unsafe.Pointer
		if len(data) != 0 {
			input = C.CBytes(data)
			if input == nil {
				return errors.New("sidereon: unable to allocate native input buffer")
			}
			defer C.free(input)
		}
		operationErr = statusErrorLocked(uint32(C.sidereon_rtcm_decode_stream_with_policy((*C.uint8_t)(input), length, C.uint32_t(policy), &messages, &diagnostics)))
		return operationErr
	})
	runtime.KeepAlive(data)
	if operationErr != nil {
		withCThread(func() {
			if messages != nil {
				C.sidereon_rtcm_messages_free(messages)
				messages = nil
			}
			if diagnostics != nil {
				C.sidereon_rtcm_stream_diagnostics_free(diagnostics)
				diagnostics = nil
			}
		})
		return nil, nil, operationErr
	}
	messageHandle, err := newRtcmMessages(messages)
	if err != nil {
		withCThread(func() {
			if messages != nil {
				C.sidereon_rtcm_messages_free(messages)
			}
			if diagnostics != nil {
				C.sidereon_rtcm_stream_diagnostics_free(diagnostics)
			}
		})
		return nil, nil, err
	}
	diagnosticHandle, err := newRtcmDiagnostics(diagnostics)
	if err != nil {
		_ = messageHandle.Close()
		return nil, nil, err
	}
	return messageHandle, diagnosticHandle, nil
}

func (d *RtcmDiagnostics) CRCFailures() (int, error) {
	var value C.size_t
	err := d.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_stream_diagnostics_crc_failures((*C.SidereonRtcmStreamDiagnostics)(p), &value)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(value))
}

func (d *RtcmDiagnostics) DepartureCount() (int, error) {
	var value C.size_t
	err := d.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rtcm_stream_diagnostics_departure_count((*C.SidereonRtcmStreamDiagnostics)(p), &value)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(value))
}

func (d *RtcmDiagnostics) Departure(index int) (NativeRTCMStreamDeparture, error) {
	if index < 0 {
		return NativeRTCMStreamDeparture{}, errNegativeIndex
	}
	var result NativeRTCMStreamDeparture
	err := d.resource.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			idx, err := checkedNativeSize(index)
			if err != nil {
				return err
			}
			var offset C.size_t
			text, err := copyNativeBytesLocked("RTCM stream departure", func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtcm_stream_diagnostics_departure((*C.SidereonRtcmStreamDiagnostics)(p), idx, &offset, out, n, written, required)
			})
			if err != nil {
				return err
			}
			result.Offset, err = checkedNativeCount(uint64(offset))
			if err != nil {
				return err
			}
			result.Description = string(text)
			return nil
		})
	})
	runtime.KeepAlive(d)
	return result, err
}
