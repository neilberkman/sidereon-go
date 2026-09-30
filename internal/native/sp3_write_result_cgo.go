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

type SP3WriteError struct {
	Kind                                                                                uint32
	HasField, HasTextValue, HasSatelliteID, HasEpochIndex, HasCommentIndex              bool
	HasColumns, HasDecimals, HasIntegerValue, HasNumber, HasYear                        bool
	HasFieldSeconds, HasResidualS, HasEpochTimeScale, HasHeaderTimeScale, HasTimeSystem bool
	HasDeclaredEpochs, HasEpochs, HasEntries, HasSatellites, HasCodes                   bool
	HasStored, HasNative, HasColumnValue, HasExponent                                   bool
	SatelliteID, TimeSystem                                                             string
	EpochIndex, CommentIndex, Columns, Decimals                                         int
	IntegerValue                                                                        uint64
	Number                                                                              float64
	Year                                                                                int64
	FieldSeconds, ResidualS                                                             float64
	EpochTimeScale, HeaderTimeScale                                                     uint32
	DeclaredEpochs                                                                      uint64
	Epochs, Entries, Satellites, Codes                                                  int
	Stored, Native, ColumnValue                                                         float64
	Exponent                                                                            int16
	Message, Field, TextValue                                                           string
}
type SP3WriteOutcome struct {
	IsOK   bool
	Status uint32
	Error  SP3WriteError
}

func sp3WriteResultBytes(result *C.SidereonSp3WriteResult, text bool, label string) ([]byte, error) {
	return copyNativeBytes(label, func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		if text {
			return C.sidereon_sp3_write_result_get_text(result, out, n, w, r)
		}
		return C.sidereon_sp3_write_result_get_message(result, out, n, w, r)
	})
}

func (s *SP3) TextResult() ([]byte, SP3WriteOutcome, error) {
	if s == nil || s.handle == nil {
		return nil, SP3WriteOutcome{}, ErrClosed
	}
	var output []byte
	var outcome SP3WriteOutcome
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonSp3WriteResult
			if err := callStatus(func() uint32 { return uint32(C.sidereon_sp3_to_sp3_text_result((*C.SidereonSp3)(pointer), &result)) }); err != nil {
				return err
			}
			if result == nil {
				return errNilNativeHandle
			}
			defer C.sidereon_sp3_write_result_free(result)
			var raw C.SidereonSp3WriteOutcome
			if err := callStatus(func() uint32 { return uint32(C.sidereon_sp3_write_result_get_outcome(result, &raw)) }); err != nil {
				return err
			}
			outcome = SP3WriteOutcome{IsOK: bool(raw.is_ok), Status: uint32(raw.status)}
			e := raw.error
			v := &outcome.Error
			v.Kind = uint32(e.kind)
			v.HasField = bool(e.has_field)
			v.HasTextValue = bool(e.has_text_value)
			v.HasSatelliteID = bool(e.has_sat_id)
			v.HasEpochIndex = bool(e.has_epoch_index)
			v.HasCommentIndex = bool(e.has_comment_index)
			v.HasColumns = bool(e.has_columns)
			v.HasDecimals = bool(e.has_decimals)
			v.HasIntegerValue = bool(e.has_integer_value)
			v.HasNumber = bool(e.has_number)
			v.HasYear = bool(e.has_year)
			v.HasFieldSeconds = bool(e.has_field_seconds)
			v.HasResidualS = bool(e.has_residual_s)
			v.HasEpochTimeScale = bool(e.has_epoch_time_scale)
			v.HasHeaderTimeScale = bool(e.has_header_time_scale)
			v.HasTimeSystem = bool(e.has_time_system)
			v.HasDeclaredEpochs = bool(e.has_declared_epochs)
			v.HasEpochs = bool(e.has_epochs)
			v.HasEntries = bool(e.has_entries)
			v.HasSatellites = bool(e.has_satellites)
			v.HasCodes = bool(e.has_codes)
			v.HasStored = bool(e.has_stored)
			v.HasNative = bool(e.has_native)
			v.HasColumnValue = bool(e.has_column_value)
			v.HasExponent = bool(e.has_exponent)
			v.SatelliteID = C.GoString(&e.sat_id.bytes[0])
			v.TimeSystem = C.GoString(&e.time_system[0])
			v.IntegerValue = uint64(e.integer_value)
			v.Number = float64(e.number)
			v.Year = int64(e.year)
			v.FieldSeconds = float64(e.field_seconds)
			v.ResidualS = float64(e.residual_s)
			v.EpochTimeScale = uint32(e.epoch_time_scale)
			v.HeaderTimeScale = uint32(e.header_time_scale)
			v.DeclaredEpochs = uint64(e.declared_epochs)
			v.Stored = float64(e.stored)
			v.Native = float64(e.native)
			v.ColumnValue = float64(e.column_value)
			v.Exponent = int16(e.exponent)
			for _, item := range []struct {
				src  C.size_t
				dst  *int
				name string
			}{{e.epoch_index, &v.EpochIndex, "SP3 write epoch index"}, {e.comment_index, &v.CommentIndex, "SP3 write comment index"}, {e.columns, &v.Columns, "SP3 write columns"}, {e.decimals, &v.Decimals, "SP3 write decimals"}, {e.epochs, &v.Epochs, "SP3 write epoch count"}, {e.entries, &v.Entries, "SP3 write entry count"}, {e.satellites, &v.Satellites, "SP3 write satellite count"}, {e.codes, &v.Codes, "SP3 write code count"}} {
				n, err := sizeTToInt(item.src, item.name)
				if err != nil {
					return err
				}
				*item.dst = n
			}
			msg, err := sp3WriteResultBytes(result, false, "SP3 write error message")
			if err != nil {
				return err
			}
			v.Message = string(msg)
			if v.HasField {
				b, err := copyNativeBytes("SP3 write field", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_sp3_write_result_get_field(result, out, n, w, r)
				})
				if err != nil {
					return err
				}
				v.Field = string(b)
			}
			if v.HasTextValue {
				b, err := copyNativeBytes("SP3 refused text value", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_sp3_write_result_get_text_value(result, out, n, w, r)
				})
				if err != nil {
					return err
				}
				v.TextValue = string(b)
			}
			if outcome.IsOK {
				output, err = sp3WriteResultBytes(result, true, "SP3 text")
			}
			return err
		})
	})
	runtime.KeepAlive(s)
	return output, outcome, err
}
