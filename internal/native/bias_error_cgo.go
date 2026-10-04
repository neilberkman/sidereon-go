//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

const biasErrorKindDeparture = 9

func biasErrorTextBytesLocked(part, departurePart uint32) ([]byte, error) {
	return copyNativeBytesLockedWithStatus("bias error text", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_last_bias_error_text(C.uint32_t(part), C.uint32_t(departurePart), out, length, written, required)
	}, statusErrorLocked)
}

func biasErrorInfoLocked(raw C.SidereonBiasErrorInfo) (*BiasError, error) {
	line, err := sizeTToInt(raw.line, "bias error line")
	if err != nil {
		return nil, err
	}
	record, err := sizeTToInt(raw.record, "bias error record")
	if err != nil {
		return nil, err
	}
	departureLine, err := sizeTToInt(raw.departure.line, "bias error departure line")
	if err != nil {
		return nil, err
	}
	first, err := sizeTToInt(raw.departure.first, "bias error first record")
	if err != nil {
		return nil, err
	}
	second, err := sizeTToInt(raw.departure.second, "bias error second record")
	if err != nil {
		return nil, err
	}
	rows, err := sizeTToInt(raw.departure.solution_rows, "bias error solution rows")
	if err != nil {
		return nil, err
	}
	result := BiasError{
		Kind: uint32(raw.kind), Line: line, Record: record, HasTimeScale: bool(raw.has_time_scale), TimeScale: uint32(raw.time_scale),
		Departure: BiasNotice{Kind: uint32(raw.departure.kind), Departure: uint32(raw.departure.departure), HasLine: bool(raw.departure.has_line), Line: departureLine, First: first, Second: second, DeclaredCount: uint64(raw.departure.declared_count), SolutionRows: rows, BiasMode: uint32(raw.departure.bias_mode), UnknownVariant: C.GoString(&raw.departure.unknown_variant[0])},
	}
	texts := []*[]byte{&result.Message, &result.Field, &result.Reason, &result.Code, &result.Version}
	for index, target := range texts {
		value, textErr := biasErrorTextBytesLocked(uint32(index), 0)
		if textErr != nil {
			return nil, textErr
		}
		*target = value
	}
	if result.Kind == biasErrorKindDeparture {
		for index := range result.DepartureText {
			value, textErr := biasErrorTextBytesLocked(5, uint32(index))
			if textErr != nil {
				return nil, textErr
			}
			result.DepartureText[index] = value
		}
	}
	return &result, nil
}

func biasErrorLocked(status uint32) error {
	statusErr := statusErrorLocked(status)
	if statusErr == nil {
		return nil
	}
	var raw C.SidereonBiasErrorInfo
	if readErr := statusErrorLocked(uint32(C.sidereon_last_bias_error(&raw))); readErr != nil || uint32(raw.kind) == 0 {
		return statusErr
	}
	value, captureErr := biasErrorInfoLocked(raw)
	if captureErr != nil {
		return statusErr
	}
	if typedStatus, ok := statusErr.(*StatusError); ok {
		typedStatus.Bias = value
	}
	return statusErr
}

func biasCallStatus(fn func() uint32) error {
	var err error
	withCThread(func() { err = biasErrorLocked(fn()) })
	return err
}
