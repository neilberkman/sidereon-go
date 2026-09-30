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
	"strings"
	"unsafe"
)

func civilClockFromC(value C.SidereonClockCivilEpoch) CivilDateTime {
	return CivilDateTime{Year: int(value.year), Month: int(value.month), Day: int(value.day), Hour: int(value.hour), Minute: int(value.minute), Second: float64(value.second)}
}

func (clock *RinexClock) Records() ([]NativeClockRecord, error) {
	if clock == nil || clock.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeClockRecord
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var snapshot *C.SidereonClockRecords
			if err := callStatus(func() uint32 { return C.sidereon_rinex_clock_records((*C.SidereonRinexClock)(owner), &snapshot) }); err != nil {
				return err
			}
			if snapshot == nil {
				return missingNativeHandle("RINEX clock records")
			}
			defer C.sidereon_clock_records_free(snapshot)
			var count C.size_t
			if err := callStatus(func() uint32 { return C.sidereon_clock_records_count(snapshot, &count) }); err != nil {
				return err
			}
			n, err := sizeTToInt(count, "RINEX clock record count")
			if err != nil {
				return err
			}
			out = make([]NativeClockRecord, n)
			for index := range out {
				var row C.SidereonClockRecord
				idx := C.size_t(index)
				if err := callStatus(func() uint32 { return C.sidereon_clock_records_get(snapshot, idx, &row) }); err != nil {
					return err
				}
				name, err := copyNativeBytesLocked("RINEX clock record name", func(dst *C.uint8_t, len C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_clock_records_name(snapshot, idx, dst, len, w, r)
				})
				if err != nil {
					return err
				}
				valueCount, err := sizeTToInt(row.value_count, "RINEX clock value count")
				if err != nil {
					return err
				}
				if valueCount > len(row.values) {
					return invalidArgument("native RINEX clock value count exceeds its record")
				}
				surplusCount, err := sizeTToInt(row.surplus_count, "RINEX clock surplus count")
				if err != nil {
					return err
				}
				if surplusCount > len(row.surplus) {
					return invalidArgument("native RINEX clock surplus count exceeds its record")
				}
				values := make([]float64, valueCount)
				for j := range values {
					values[j] = float64(row.values[j])
				}
				surplus := make([]NativeClockSurplusValue, surplusCount)
				for j := range surplus {
					surplus[j] = NativeClockSurplusValue{Position: uint64(row.surplus[j].position), Value: float64(row.surplus[j].value)}
				}
				out[index] = NativeClockRecord{RecordType: uint32(row.record_type), Name: string(name), HasSatellite: bool(row.has_satellite), Satellite: tokenFromC(row.satellite), CivilEpoch: civilClockFromC(row.civil_epoch), HasEpoch: bool(row.has_epoch), Epoch: clockEpochFromC(row.epoch), Values: values, Surplus: surplus, HasLine: bool(row.has_line), Line: uint64(row.line), LineCount: uint64(row.line_count), Reading: uint32(row.reading), HasContinuationReading: bool(row.has_continuation_reading), ContinuationReading: uint32(row.continuation_reading), ReadingUnknownVariant: C.GoString(&row.reading_unknown_variant[0]), ContinuationReadingUnknownVariant: C.GoString(&row.continuation_reading_unknown_variant[0])}
			}
			return nil
		})
	})
	runtime.KeepAlive(clock)
	return out, err
}
func (clock *RinexClock) RecordCount() (int, error) {
	if clock == nil || clock.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_rinex_clock_record_count((*C.SidereonRinexClock)(owner), &count) })
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "RINEX clock record count")
}

func (clock *RinexClock) Info() (NativeRinexClockInfo, error) {
	if clock == nil || clock.resource == nil {
		return NativeRinexClockInfo{}, ErrClosed
	}
	var raw C.SidereonRinexClockInfo
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_rinex_clock_info((*C.SidereonRinexClock)(owner), &raw)) })
	})
	runtime.KeepAlive(clock)
	if err != nil {
		return NativeRinexClockInfo{}, err
	}
	counts := []C.size_t{raw.time_system_label_count, raw.header_record_count, raw.record_count, raw.series_count, raw.sample_count, raw.skipped_record_count, raw.diagnostic_count, raw.notice_count}
	converted := make([]int, len(counts))
	for i, count := range counts {
		converted[i], err = sizeTToInt(count, "RINEX clock info count")
		if err != nil {
			return NativeRinexClockInfo{}, err
		}
	}
	return NativeRinexClockInfo{HasVersion: bool(raw.has_version), Version: float64(raw.version), HasLayout: bool(raw.has_layout), Layout: uint32(raw.layout), HasSatelliteSystem: bool(raw.has_satellite_system), SatelliteSystem: uint32(raw.satellite_system), HasTimeSystem: bool(raw.has_time_system), TimeSystem: uint32(raw.time_system), TimeSystemStatus: uint32(raw.time_system_status), TimeSystemLabelCount: converted[0], HasTimeScale: bool(raw.has_time_scale), TimeScale: uint32(raw.time_scale), HeaderRecordCount: converted[1], RecordCount: converted[2], SeriesCount: converted[3], SampleCount: converted[4], SkippedRecordCount: converted[5], DiagnosticCount: converted[6], NoticeCount: converted[7], TimeSystemUnknownVariant: C.GoString(&raw.time_system_unknown_variant[0]), TimeSystemStatusUnknownVariant: C.GoString(&raw.time_system_status_unknown_variant[0])}, nil
}

// SourceLine copies one original source line. Constructed products and line
// numbers outside the parsed input return present=false.
func (clock *RinexClock) SourceLine(line uint64) (string, bool, error) {
	if clock == nil || clock.resource == nil {
		return "", false, ErrClosed
	}
	var text string
	var present bool
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			var has C.bool
			status := C.sidereon_rinex_clock_source_line((*C.SidereonRinexClock)(owner), C.size_t(line), nil, 0, &written, &required, &has)
			if err := callStatus(func() uint32 { return uint32(status) }); err != nil {
				return err
			}
			count, err := validateNativeQuery("RINEX clock source line", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			present = bool(has)
			if !present {
				if count != 0 {
					return invalidArgument("absent RINEX clock source line reported bytes")
				}
				return nil
			}
			buf := make([]C.uint8_t, count)
			if count > 0 {
				status = C.sidereon_rinex_clock_source_line((*C.SidereonRinexClock)(owner), C.size_t(line), &buf[0], C.size_t(count), &written, &required, &has)
				if err := callStatus(func() uint32 { return uint32(status) }); err != nil {
					return err
				}
			}
			if !bool(has) {
				return invalidArgument("RINEX clock source line presence changed during copy")
			}
			if _, err := validateTwoPassCounts("RINEX clock source line", count, count, uint64(written), uint64(required)); err != nil {
				return err
			}
			if count > 0 {
				text = string(unsafe.Slice((*byte)(unsafe.Pointer(&buf[0])), count))
			} else {
				text = ""
			}
			return nil
		})
	})
	runtime.KeepAlive(clock)
	return text, present, err
}

// TimeSystemLabel copies a retained unrecognized or conflicting label.
func (clock *RinexClock) TimeSystemLabel(index int) (string, error) {
	if index < 0 {
		return "", errNegativeIndex
	}
	if clock == nil || clock.resource == nil {
		return "", ErrClosed
	}
	var label string
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			value, err := copyNativeBytesLocked("RINEX clock time-system label", func(dst *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rinex_clock_time_system_label((*C.SidereonRinexClock)(owner), C.size_t(index), dst, n, written, required)
			})
			label = string(value)
			return err
		})
	})
	runtime.KeepAlive(clock)
	return label, err
}

func headerRecordText(snapshot *C.SidereonClockHeaderRecords, index C.size_t, part uint32, label string) (string, error) {
	bytes, err := copyNativeBytesLocked(label, func(dst *C.uint8_t, len C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_clock_header_records_text(snapshot, index, C.uint32_t(part), dst, len, w, r)
	})
	return string(bytes), err
}

func (clock *RinexClock) HeaderRecords() ([]NativeClockHeaderRecord, error) {
	if clock == nil || clock.resource == nil {
		return nil, ErrClosed
	}
	var out []NativeClockHeaderRecord
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var snapshot *C.SidereonClockHeaderRecords
			if err := callStatus(func() uint32 { return C.sidereon_rinex_clock_header_records((*C.SidereonRinexClock)(owner), &snapshot) }); err != nil {
				return err
			}
			if snapshot == nil {
				return missingNativeHandle("RINEX clock header records")
			}
			defer C.sidereon_clock_header_records_free(snapshot)
			var count C.size_t
			if err := callStatus(func() uint32 { return C.sidereon_clock_header_records_count(snapshot, &count) }); err != nil {
				return err
			}
			n, err := sizeTToInt(count, "RINEX clock header record count")
			if err != nil {
				return err
			}
			out = make([]NativeClockHeaderRecord, n)
			for index := range out {
				idx := C.size_t(index)
				var row C.SidereonClockHeaderRecord
				if err := callStatus(func() uint32 { return C.sidereon_clock_header_records_get(snapshot, idx, &row) }); err != nil {
					return err
				}
				line, err := headerRecordText(snapshot, idx, uint32(C.SIDEREON_CLOCK_HEADER_TEXT_LINE), "RINEX clock header line")
				if err != nil {
					return err
				}
				label, err := headerRecordText(snapshot, idx, uint32(C.SIDEREON_CLOCK_HEADER_TEXT_LABEL), "RINEX clock header label")
				if err != nil {
					return err
				}
				payload, err := headerRecordText(snapshot, idx, uint32(C.SIDEREON_CLOCK_HEADER_TEXT_PAYLOAD), "RINEX clock header payload")
				if err != nil {
					return err
				}
				partCount, err := sizeTToInt(row.text_part_count, "RINEX clock header text-part count")
				if err != nil {
					return err
				}
				parts := make([]string, partCount)
				for part := range parts {
					pidx := C.size_t(part)
					bytes, err := copyNativeBytesLocked("RINEX clock header typed text", func(dst *C.uint8_t, len C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
						return C.sidereon_clock_header_records_field_text(snapshot, idx, pidx, dst, len, w, r)
					})
					if err != nil {
						return err
					}
					parts[part] = string(bytes)
				}
				var xyz [3]int64
				for j := range xyz {
					xyz[j] = int64(row.xyz_mm[j])
				}
				out[index] = NativeClockHeaderRecord{HasLine: bool(row.has_line), Line: uint64(row.line), LabelColumn: uint64(row.label_column), Reading: uint32(row.reading), FieldKind: uint32(row.field_kind), TextParts: parts, HasVersion: bool(row.has_version), Version: float64(row.version), HasSystemCode: bool(row.has_system_code), SystemCode: uint32(row.system_code), HasCount: bool(row.has_count), Count: uint64(row.count), HasInteger: bool(row.has_integer), Integer: int64(row.integer), HasStart: bool(row.has_start), HasStop: bool(row.has_stop), Start: civilClockFromC(row.start), Stop: civilClockFromC(row.stop), HasConstraintS: bool(row.has_constraint_s), ConstraintS: float64(row.constraint_s), HasXYZMM: bool(row.has_xyz_mm), XYZMM: xyz, ReadingUnknownVariant: C.GoString(&row.reading_unknown_variant[0]), FieldKindUnknownVariant: C.GoString(&row.field_kind_unknown_variant[0]), LineText: line, Label: label, Payload: payload}
			}
			return nil
		})
	})
	runtime.KeepAlive(clock)
	return out, err
}

func cClockCivil(value CivilDateTime) C.SidereonClockCivilEpoch {
	return C.SidereonClockCivilEpoch{year: C.int32_t(value.Year), month: C.uint8_t(value.Month), day: C.uint8_t(value.Day), hour: C.uint8_t(value.Hour), minute: C.uint8_t(value.Minute), second: C.double(value.Second)}
}

func checkedClockCivil(value CivilDateTime) (C.SidereonClockCivilEpoch, error) {
	if int64(value.Year) < -2147483648 || int64(value.Year) > 2147483647 || value.Month < 0 || value.Month > 255 || value.Day < 0 || value.Day > 255 || value.Hour < 0 || value.Hour > 255 || value.Minute < 0 || value.Minute > 255 {
		return C.SidereonClockCivilEpoch{}, invalidArgument("RINEX clock civil epoch field is outside its native representation")
	}
	return cClockCivil(value), nil
}

func readClockWriteFailureLocked(result *C.SidereonRinexClockResult, outcome C.SidereonRinexClockOutcome) (*NativeClockWriteFailure, error) {
	if bool(outcome.is_ok) {
		return nil, nil
	}
	failure := &NativeClockWriteFailure{Kind: uint32(outcome.error.kind), HasLine: bool(outcome.error.has_line), Line: uint64(outcome.error.line), HasTimeScale: bool(outcome.error.has_time_scale), TimeScale: uint32(outcome.error.time_scale), HasField: bool(outcome.error.has_field), HasReason: bool(outcome.error.has_reason), HasRecord: bool(outcome.error.has_record), HasRecordType: bool(outcome.error.has_record_type), HasValue: bool(outcome.error.has_value)}
	readPart := func(part C.enum_SidereonRinexClockErrorText, label string) (string, error) {
		value, err := copyNativeBytesLocked(label, func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_rinex_clock_result_error_text(result, C.uint32_t(part), out, n, written, required)
		})
		return string(value), err
	}
	var err error
	if failure.Message, err = readPart(C.SIDEREON_RINEX_CLOCK_ERROR_TEXT_MESSAGE, "RINEX clock write error message"); err != nil {
		return nil, err
	}
	if failure.HasField {
		if failure.Field, err = readPart(C.SIDEREON_RINEX_CLOCK_ERROR_TEXT_FIELD, "RINEX clock write error field"); err != nil {
			return nil, err
		}
	}
	if failure.HasReason {
		if failure.Reason, err = readPart(C.SIDEREON_RINEX_CLOCK_ERROR_TEXT_REASON, "RINEX clock write error reason"); err != nil {
			return nil, err
		}
	}
	if failure.HasRecord {
		if failure.Record, err = readPart(C.SIDEREON_RINEX_CLOCK_ERROR_TEXT_RECORD, "RINEX clock write error record"); err != nil {
			return nil, err
		}
	}
	if failure.HasRecordType {
		if failure.RecordType, err = readPart(C.SIDEREON_RINEX_CLOCK_ERROR_TEXT_RECORD_TYPE, "RINEX clock write error record type"); err != nil {
			return nil, err
		}
	}
	if failure.HasValue {
		if failure.Value, err = readPart(C.SIDEREON_RINEX_CLOCK_ERROR_TEXT_VALUE, "RINEX clock write error value"); err != nil {
			return nil, err
		}
	}
	return failure, nil
}

func finishClockMutationLocked(status C.enum_SidereonStatus, result *C.SidereonRinexClockResult) error {
	if result != nil {
		defer C.sidereon_rinex_clock_result_free(result)
		var outcome C.SidereonRinexClockOutcome
		if err := statusErrorLocked(uint32(C.sidereon_rinex_clock_result_get_outcome(result, &outcome))); err != nil {
			return err
		}
		if failure, err := readClockWriteFailureLocked(result, outcome); err != nil {
			return err
		} else if failure != nil {
			return failure
		}
	}
	return statusErrorLocked(uint32(status))
}
func cClockValues(values []float64) (unsafe.Pointer, error) {
	if len(values) == 0 {
		return nil, nil
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.double(0)))
	if err != nil {
		return nil, err
	}
	p := C.malloc(C.size_t(size))
	if p == nil {
		return nil, errors.New("sidereon: unable to allocate clock record values")
	}
	out := unsafe.Slice((*C.double)(p), len(values))
	for i, v := range values {
		out[i] = C.double(v)
	}
	return p, nil
}
func (clock *RinexClock) InsertRecord(index int, recordType uint32, name string, epoch CivilDateTime, values []float64) error {
	if index < 0 {
		return errNegativeIndex
	}
	if clock == nil || clock.resource == nil {
		return ErrClosed
	}
	if strings.IndexByte(name, 0) >= 0 {
		return invalidArgument("RINEX clock record name contains NUL")
	}
	civil, err := checkedClockCivil(epoch)
	if err != nil {
		return err
	}
	if _, err := checkedNativeSize(len(name)); err != nil {
		return err
	}
	namePointer := C.CString(name)
	if namePointer == nil {
		return errors.New("sidereon: unable to allocate clock record name")
	}
	defer C.free(unsafe.Pointer(namePointer))
	valuesPointer, err := cClockValues(values)
	if err != nil {
		return err
	}
	if valuesPointer != nil {
		defer C.free(valuesPointer)
	}
	err = clock.resource.withExclusive(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonRinexClockResult
			status := C.sidereon_rinex_clock_insert_record((*C.SidereonRinexClock)(owner), C.size_t(index), C.uint32_t(recordType), namePointer, &civil, (*C.double)(valuesPointer), C.size_t(len(values)), &result)
			return finishClockMutationLocked(status, result)
		})
	})
	runtime.KeepAlive(clock)
	return err
}
func (clock *RinexClock) SetRecordValues(index int, values []float64) error {
	if index < 0 {
		return errNegativeIndex
	}
	if clock == nil || clock.resource == nil {
		return ErrClosed
	}
	p, err := cClockValues(values)
	if err != nil {
		return err
	}
	if p != nil {
		defer C.free(p)
	}
	err = clock.resource.withExclusive(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonRinexClockResult
			status := C.sidereon_rinex_clock_set_record_values((*C.SidereonRinexClock)(owner), C.size_t(index), (*C.double)(p), C.size_t(len(values)), &result)
			return finishClockMutationLocked(status, result)
		})
	})
	runtime.KeepAlive(clock)
	return err
}
func (clock *RinexClock) RemoveRecord(index int) error {
	if index < 0 {
		return errNegativeIndex
	}
	if clock == nil || clock.resource == nil {
		return ErrClosed
	}
	err := clock.resource.withExclusive(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonRinexClockResult
			status := C.sidereon_rinex_clock_remove_record((*C.SidereonRinexClock)(owner), C.size_t(index), &result)
			return finishClockMutationLocked(status, result)
		})
	})
	runtime.KeepAlive(clock)
	return err
}

// RemoveRecordWithValue removes a row and returns the owned row snapshot that
// the native mutation result retained before changing the product.
func (clock *RinexClock) RemoveRecordWithValue(index int) (NativeClockRecord, bool, error) {
	if index < 0 {
		return NativeClockRecord{}, false, errNegativeIndex
	}
	if clock == nil || clock.resource == nil {
		return NativeClockRecord{}, false, ErrClosed
	}
	var removed NativeClockRecord
	var present bool
	err := clock.resource.withExclusive(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonRinexClockResult
			status := C.sidereon_rinex_clock_remove_record((*C.SidereonRinexClock)(owner), C.size_t(index), &result)
			if result == nil {
				return statusErrorLocked(uint32(status))
			}
			defer C.sidereon_rinex_clock_result_free(result)
			var outcome C.SidereonRinexClockOutcome
			if err := statusErrorLocked(uint32(C.sidereon_rinex_clock_result_get_outcome(result, &outcome))); err != nil {
				return err
			}
			if failure, err := readClockWriteFailureLocked(result, outcome); err != nil {
				return err
			} else if failure != nil {
				return failure
			}
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			var row C.SidereonClockRecord
			var has C.bool
			if err := callStatus(func() uint32 { return uint32(C.sidereon_rinex_clock_result_record(result, &has, &row)) }); err != nil {
				return err
			}
			if !bool(has) {
				return nil
			}
			name, err := copyNativeBytesLocked("removed RINEX clock record name", func(dst *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rinex_clock_result_record_name(result, dst, n, w, r)
			})
			if err != nil {
				return err
			}
			valueCount, err := sizeTToInt(row.value_count, "removed RINEX clock value count")
			if err != nil {
				return err
			}
			if valueCount > len(row.values) {
				return invalidArgument("native removed RINEX clock value count exceeds its record")
			}
			surplusCount, err := sizeTToInt(row.surplus_count, "removed RINEX clock surplus count")
			if err != nil {
				return err
			}
			if surplusCount > len(row.surplus) {
				return invalidArgument("native removed RINEX clock surplus count exceeds its record")
			}
			values := make([]float64, valueCount)
			for i := range values {
				values[i] = float64(row.values[i])
			}
			surplus := make([]NativeClockSurplusValue, surplusCount)
			for i := range surplus {
				surplus[i] = NativeClockSurplusValue{Position: uint64(row.surplus[i].position), Value: float64(row.surplus[i].value)}
			}
			removed = NativeClockRecord{RecordType: uint32(row.record_type), Name: string(name), HasSatellite: bool(row.has_satellite), Satellite: tokenFromC(row.satellite), CivilEpoch: civilClockFromC(row.civil_epoch), HasEpoch: bool(row.has_epoch), Epoch: clockEpochFromC(row.epoch), Values: values, Surplus: surplus, HasLine: bool(row.has_line), Line: uint64(row.line), LineCount: uint64(row.line_count), Reading: uint32(row.reading), HasContinuationReading: bool(row.has_continuation_reading), ContinuationReading: uint32(row.continuation_reading), ReadingUnknownVariant: C.GoString(&row.reading_unknown_variant[0]), ContinuationReadingUnknownVariant: C.GoString(&row.continuation_reading_unknown_variant[0])}
			present = true
			return nil
		})
	})
	runtime.KeepAlive(clock)
	return removed, present, err
}
func (clock *RinexClock) SetTimeSystem(system uint32) error {
	if clock == nil || clock.resource == nil {
		return ErrClosed
	}
	err := clock.resource.withExclusive(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonRinexClockResult
			status := C.sidereon_rinex_clock_set_time_system((*C.SidereonRinexClock)(owner), C.uint32_t(system), &result)
			return finishClockMutationLocked(status, result)
		})
	})
	runtime.KeepAlive(clock)
	return err
}
func clockDepartureText(result *C.SidereonRinexClockResult, index C.size_t, part uint32, label string) (string, error) {
	b, e := copyNativeBytesLocked(label, func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_rinex_clock_result_departure_text(result, index, C.uint32_t(part), out, n, w, r)
	})
	return string(b), e
}
func (clock *RinexClock) TextWithPolicy(allowNearestMicrosecond bool) (NativeClockWriteResult, error) {
	if clock == nil || clock.resource == nil {
		return NativeClockWriteResult{}, ErrClosed
	}
	var output NativeClockWriteResult
	err := clock.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var policy C.SidereonClockWritePolicy
			var policyPointer *C.SidereonClockWritePolicy
			if allowNearestMicrosecond {
				if e := callStatus(func() uint32 { return C.sidereon_clock_write_policy_init(&policy) }); e != nil {
					return e
				}
				policy.nearest_microsecond_epochs = C.SIDEREON_CLOCK_WRITE_LENIENCY_ALLOW
				policyPointer = &policy
			}
			var result *C.SidereonRinexClockResult
			if e := callStatus(func() uint32 {
				return C.sidereon_rinex_clock_to_text_result((*C.SidereonRinexClock)(owner), policyPointer, &result)
			}); e != nil {
				return e
			}
			if result == nil {
				return missingNativeHandle("RINEX clock write result")
			}
			defer C.sidereon_rinex_clock_result_free(result)
			var outcome C.SidereonRinexClockOutcome
			if e := callStatus(func() uint32 { return C.sidereon_rinex_clock_result_get_outcome(result, &outcome) }); e != nil {
				return e
			}
			if failure, e := readClockWriteFailureLocked(result, outcome); e != nil {
				return e
			} else if failure != nil {
				return failure
			}
			text, e := copyNativeBytesLocked("RINEX clock output", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rinex_clock_result_get_text(result, out, n, w, r)
			})
			if e != nil {
				return e
			}
			output.Text = text
			var written, required C.size_t
			if e = callStatus(func() uint32 { return C.sidereon_rinex_clock_result_departures(result, nil, 0, &written, &required) }); e != nil {
				return e
			}
			count, e := validateNativeQuery("RINEX clock write departures", uint64(written), uint64(required))
			if e != nil {
				return e
			}
			rawMem, e := allocNativeArray(count, unsafe.Sizeof(C.SidereonClockWriteDeparture{}))
			if e != nil {
				return e
			}
			defer C.free(rawMem)
			raw := unsafe.Slice((*C.SidereonClockWriteDeparture)(rawMem), count)
			var dst *C.SidereonClockWriteDeparture
			if count > 0 {
				dst = &raw[0]
			}
			written, required = 0, 0
			if e = callStatus(func() uint32 {
				return C.sidereon_rinex_clock_result_departures(result, dst, C.size_t(count), &written, &required)
			}); e != nil {
				return e
			}
			if _, e = validateTwoPassCounts("RINEX clock write departures", count, count, uint64(written), uint64(required)); e != nil {
				return e
			}
			output.Departures = make([]NativeClockWriteDeparture, count)
			for i, v := range raw {
				name, e := clockDepartureText(result, C.size_t(i), uint32(C.SIDEREON_CLOCK_DEPARTURE_TEXT_NAME), "RINEX clock departure name")
				if e != nil {
					return e
				}
				writtenText, e := clockDepartureText(result, C.size_t(i), uint32(C.SIDEREON_CLOCK_DEPARTURE_TEXT_WRITTEN), "RINEX clock departure epoch")
				if e != nil {
					return e
				}
				output.Departures[i] = NativeClockWriteDeparture{Kind: uint32(v.kind), Record: uint64(v.record), HasEpoch: bool(v.has_epoch), Epoch: clockEpochFromC(v.epoch), UnknownVariant: C.GoString(&v.unknown_variant[0]), Name: name, Written: writtenText}
			}
			return nil
		})
	})
	runtime.KeepAlive(clock)
	return output, err
}
