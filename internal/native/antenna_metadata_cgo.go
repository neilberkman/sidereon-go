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

type AntennaInfo struct {
	Kind                                                                uint32
	HasDazi                                                             bool
	DaziDeg                                                             float64
	HasZenithGrid                                                       bool
	ZenithStartDeg, ZenithEndDeg, ZenithStepDeg                         float64
	HasFrequencyCountRecord, HasSinexCode                               bool
	HasValidFrom, HasValidUntil                                         bool
	ValidFrom, ValidUntil                                               AntexDateTime
	CalibrationCount, LeadingCommentCount, CommentCount, FrequencyCount int
}

type AntexCalibration struct {
	HasAntennasCalibrated bool
	AntennasCalibrated    uint32
}
type AntexFrequencyInfo struct {
	PCOM               [3]float64
	PCVSampleCount     int
	HasRMS, HasRMSPCOM bool
	RMSPCOM            [3]float64
	RMSPCVSampleCount  int
}
type AntexPCVSample struct {
	Grid                          uint32
	HasAzimuth                    bool
	AzimuthDeg, ZenithDeg, ValueM float64
}

func antexDateTimeFromC(value C.SidereonAntexDateTime) AntexDateTime {
	return AntexDateTime{Year: int32(value.year), Month: uint8(value.month), Day: uint8(value.day), Hour: uint8(value.hour), Minute: uint8(value.minute), Second: uint8(value.second), FractionDigits: uint64(value.fraction_digits), FractionScale: uint64(value.fraction_scale)}
}

func antexStatusErrorLocked(status uint32) error {
	base := statusErrorLocked(status)
	if base == nil {
		return nil
	}
	var raw C.SidereonAntexError
	if C.sidereon_last_antex_error(&raw) != C.SIDEREON_STATUS_OK {
		return base
	}
	detail := AntexError{Kind: uint32(raw.kind), HasAntennaID: bool(raw.has_antenna_id), HasRecord: bool(raw.has_record), HasField: bool(raw.has_field), HasValue: bool(raw.has_value), HasFrequency: bool(raw.has_frequency), HasReason: bool(raw.has_reason), HasSections: bool(raw.has_sections)}
	if detail.HasSections {
		n, err := sizeTToInt(raw.sections, "ANTEX error section count")
		if err != nil {
			return base
		}
		detail.Sections = n
	}
	read := func(part C.enum_SidereonAntexErrorText, label string) (string, error) {
		b, err := copyNativeBytesLocked(label, func(o *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_last_antex_error_text(C.uint32_t(part), o, n, w, r)
		})
		return string(b), err
	}
	var err error
	if detail.Message, err = read(C.SIDEREON_ANTEX_ERROR_TEXT_MESSAGE, "ANTEX failure message"); err != nil {
		return base
	}
	if detail.HasAntennaID {
		if detail.AntennaID, err = read(C.SIDEREON_ANTEX_ERROR_TEXT_ANTENNA_ID, "ANTEX failure antenna id"); err != nil {
			return base
		}
	}
	if detail.HasRecord {
		if detail.Record, err = read(C.SIDEREON_ANTEX_ERROR_TEXT_RECORD, "ANTEX failure record"); err != nil {
			return base
		}
	}
	if detail.HasField {
		if detail.Field, err = read(C.SIDEREON_ANTEX_ERROR_TEXT_FIELD, "ANTEX failure field"); err != nil {
			return base
		}
	}
	if detail.HasValue {
		if detail.Value, err = read(C.SIDEREON_ANTEX_ERROR_TEXT_VALUE, "ANTEX failure value"); err != nil {
			return base
		}
	}
	if detail.HasFrequency {
		if detail.Frequency, err = read(C.SIDEREON_ANTEX_ERROR_TEXT_FREQUENCY, "ANTEX failure frequency"); err != nil {
			return base
		}
	}
	if detail.HasReason {
		if detail.Reason, err = read(C.SIDEREON_ANTEX_ERROR_TEXT_REASON, "ANTEX failure reason"); err != nil {
			return base
		}
	}
	var statusErr *StatusError
	if errors.As(base, &statusErr) && detail.Kind != 0 {
		statusErr.ANTEX = &detail
	}
	return base
}

func (a *Antenna) Info() (AntennaInfo, error) {
	if a == nil || a.handle == nil {
		return AntennaInfo{}, ErrClosed
	}
	var v C.SidereonAntennaInfo
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 { return uint32(C.sidereon_antenna_info((*C.SidereonAntenna)(p), &v)) })
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		return AntennaInfo{}, err
	}
	cal, err := sizeTToInt(v.calibration_count, "ANTEX calibration count")
	if err != nil {
		return AntennaInfo{}, err
	}
	lead, err := sizeTToInt(v.leading_comment_count, "ANTEX leading comment count")
	if err != nil {
		return AntennaInfo{}, err
	}
	comments, err := sizeTToInt(v.comment_count, "ANTEX comment count")
	if err != nil {
		return AntennaInfo{}, err
	}
	freq, err := sizeTToInt(v.frequency_count, "ANTEX frequency count")
	if err != nil {
		return AntennaInfo{}, err
	}
	return AntennaInfo{Kind: uint32(v.kind), HasDazi: bool(v.has_dazi_deg), DaziDeg: float64(v.dazi_deg), HasZenithGrid: bool(v.has_zenith_grid), ZenithStartDeg: float64(v.zenith_start_deg), ZenithEndDeg: float64(v.zenith_end_deg), ZenithStepDeg: float64(v.zenith_step_deg), HasFrequencyCountRecord: bool(v.has_frequency_count_record), HasSinexCode: bool(v.has_sinex_code), HasValidFrom: bool(v.has_valid_from), HasValidUntil: bool(v.has_valid_until), ValidFrom: antexDateTimeFromC(v.valid_from), ValidUntil: antexDateTimeFromC(v.valid_until), CalibrationCount: cal, LeadingCommentCount: lead, CommentCount: comments, FrequencyCount: freq}, nil
}

func (a *Antenna) Text(part uint32) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	var out []byte
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			v, e := copyNativeBytesLocked("ANTEX antenna text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_antenna_text((*C.SidereonAntenna)(p), C.uint32_t(part), o, n, w, r)
			})
			out = v
			return e
		})
	})
	runtime.KeepAlive(a)
	return string(out), err
}

func (a *Antenna) Comment(list uint32, index int) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	if index < 0 {
		return "", invalidArgument("negative ANTEX antenna comment index")
	}
	var out []byte
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			v, e := copyNativeBytesLocked("ANTEX antenna comment", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_antenna_comment((*C.SidereonAntenna)(p), C.uint32_t(list), C.size_t(index), o, n, w, r)
			})
			out = v
			return e
		})
	})
	runtime.KeepAlive(a)
	return string(out), err
}

func (a *Antenna) Calibration(index int) (AntexCalibration, error) {
	if a == nil || a.handle == nil {
		return AntexCalibration{}, ErrClosed
	}
	if index < 0 {
		return AntexCalibration{}, invalidArgument("negative ANTEX calibration index")
	}
	var v C.SidereonAntexCalibration
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_antenna_calibration((*C.SidereonAntenna)(p), C.size_t(index), &v))
			})
		})
	})
	runtime.KeepAlive(a)
	return AntexCalibration{bool(v.has_antennas_calibrated), uint32(v.antennas_calibrated)}, err
}
func (a *Antenna) CalibrationText(index int, part uint32) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	if index < 0 {
		return "", invalidArgument("negative ANTEX calibration index")
	}
	var out []byte
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			v, e := copyNativeBytesLocked("ANTEX calibration text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_antenna_calibration_text((*C.SidereonAntenna)(p), C.size_t(index), C.uint32_t(part), o, n, w, r)
			})
			out = v
			return e
		})
	})
	runtime.KeepAlive(a)
	return string(out), err
}
func (a *Antenna) Frequency(index int) (AntexFrequencyInfo, error) {
	if a == nil || a.handle == nil {
		return AntexFrequencyInfo{}, ErrClosed
	}
	if index < 0 {
		return AntexFrequencyInfo{}, invalidArgument("negative ANTEX frequency index")
	}
	var v C.SidereonAntexFrequencyInfo
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_antenna_frequency((*C.SidereonAntenna)(p), C.size_t(index), &v))
			})
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		return AntexFrequencyInfo{}, err
	}
	n, e := sizeTToInt(v.pcv_sample_count, "ANTEX PCV sample count")
	if e != nil {
		return AntexFrequencyInfo{}, e
	}
	rn, e := sizeTToInt(v.rms_pcv_sample_count, "ANTEX RMS PCV sample count")
	if e != nil {
		return AntexFrequencyInfo{}, e
	}
	out := AntexFrequencyInfo{PCVSampleCount: n, HasRMS: bool(v.has_rms), HasRMSPCOM: bool(v.has_rms_pco_m), RMSPCVSampleCount: rn}
	for j := range out.PCOM {
		out.PCOM[j] = float64(v.pco_m[j])
		out.RMSPCOM[j] = float64(v.rms_pco_m[j])
	}
	return out, nil
}
func (a *Antenna) FrequencyLabel(index int) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	if index < 0 {
		return "", invalidArgument("negative ANTEX frequency index")
	}
	var out []byte
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			v, e := copyNativeBytesLocked("ANTEX frequency label", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_antenna_frequency_label((*C.SidereonAntenna)(p), C.size_t(index), o, n, w, r)
			})
			out = v
			return e
		})
	})
	runtime.KeepAlive(a)
	return string(out), err
}
func (a *Antenna) FrequencyPCVSamples(index int, rms bool) ([]AntexPCVSample, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	if index < 0 {
		return nil, invalidArgument("negative ANTEX frequency index")
	}
	var out []AntexPCVSample
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var needed, written C.size_t
			invoke := func(dst *C.SidereonAntexPcvSample, n C.size_t) C.enum_SidereonStatus {
				return C.sidereon_antenna_frequency_pcv_samples((*C.SidereonAntenna)(p), C.size_t(index), C.bool(rms), dst, n, &written, &needed)
			}
			if e := callStatus(func() uint32 { return uint32(invoke(nil, 0)) }); e != nil {
				return e
			}
			count, e := validateNativeQuery("ANTEX PCV sample count", uint64(written), uint64(needed))
			if e != nil {
				return e
			}
			if _, e = checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonAntexPcvSample{})); e != nil {
				return e
			}
			rows := make([]C.SidereonAntexPcvSample, count)
			if count > 0 {
				if e = callStatus(func() uint32 { return uint32(invoke(&rows[0], C.size_t(count))) }); e != nil {
					return e
				}
			}
			if _, e = validateTwoPassCounts("ANTEX PCV samples", count, count, uint64(written), uint64(needed)); e != nil {
				return e
			}
			out = make([]AntexPCVSample, count)
			for j, v := range rows {
				out[j] = AntexPCVSample{Grid: uint32(v.grid), HasAzimuth: bool(v.has_azimuth_deg), AzimuthDeg: float64(v.azimuth_deg), ZenithDeg: float64(v.zenith_deg), ValueM: float64(v.value_m)}
			}
			return nil
		})
	})
	runtime.KeepAlive(a)
	return out, err
}
func (a *Antenna) ValidAt(epoch AntexDateTime) (bool, error) {
	if a == nil || a.handle == nil {
		return false, ErrClosed
	}
	date := antexDateTimeToC(epoch)
	var valid C.bool
	err := a.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			return callStatus(func() uint32 { return uint32(C.sidereon_antenna_valid_at((*C.SidereonAntenna)(p), &date, &valid)) })
		})
	})
	runtime.KeepAlive(a)
	return bool(valid), err
}
