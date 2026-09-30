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

type IonexEpochError struct {
	Kind, Scale  uint32
	HasUTCJ2000S bool
	UTCJ2000S    int64
}
type IonexSlantError struct {
	Kind, CoverageError                                uint32
	HasGap                                             bool
	Gap                                                IonexNodeGap
	Refusal                                            uint32
	RefusalMapNumber, RefusalLatIndex, RefusalLonIndex uint64
	HasMappingDeclaration                              bool
	MappingDeclaration                                 uint32
	HasMappingFunction                                 bool
	MappingFunction                                    uint32
}
type IonexInstantSlantRequest struct {
	LatDeg, LonDeg, AzimuthDeg, ElevationDeg float64
	Epoch                                    NativeClockEpoch
	FrequencyHz                              float64
}
type IonexInstantSlantRow struct {
	IsOK        bool
	Status      uint32
	Evaluation  IonexSlantDelayEvaluation
	Error       IonexSlantError
	EpochError  IonexEpochError
	Message     string
	MappingCode string
}
type IonexInstantSlantResultList struct{ Rows []IonexInstantSlantRow }
type IonexSlantPolicy struct{ Coverage, MissingNodes, Mapping uint32 }

func clockEpochToC(value NativeClockEpoch) C.SidereonClockEpoch {
	return C.SidereonClockEpoch{scale: C.uint32_t(value.Scale), representation: C.uint32_t(value.Representation), jd_whole: C.double(value.JulianWhole), jd_fraction: C.double(value.JulianFraction), nanos_high: C.int64_t(value.NanosHigh), nanos_low: C.uint64_t(value.NanosLow)}
}
func epochErrorFromC(v C.SidereonIonexEpochError) IonexEpochError {
	return IonexEpochError{Kind: uint32(v.kind), Scale: uint32(v.scale), HasUTCJ2000S: bool(v.has_utc_j2000_s), UTCJ2000S: int64(v.utc_j2000_s)}
}
func slantErrorFromC(v C.SidereonIonexSlantError) IonexSlantError {
	o := IonexSlantError{Kind: uint32(v.kind), CoverageError: uint32(v.coverage_error), HasGap: bool(v.has_gap), Refusal: uint32(v.refusal), RefusalMapNumber: uint64(v.refusal_map_number), RefusalLatIndex: uint64(v.refusal_lat_index), RefusalLonIndex: uint64(v.refusal_lon_index), HasMappingDeclaration: bool(v.has_mapping_declaration), MappingDeclaration: uint32(v.mapping_declaration), HasMappingFunction: bool(v.has_mapping_function), MappingFunction: uint32(v.mapping_function)}
	o.Gap.HasGap = bool(v.gap.has_gap)
	for n := 0; n < 4; n++ {
		o.Gap.Earlier[n] = bool(v.gap.earlier.missing[n])
		o.Gap.Later[n] = bool(v.gap.later.missing[n])
	}
	return o
}
func ionexPolicyToC(v IonexSlantPolicy) C.SidereonIonexSlantPolicy {
	return C.SidereonIonexSlantPolicy{coverage: C.uint32_t(v.Coverage), missing_nodes: C.uint32_t(v.MissingNodes), mapping: C.uint32_t(v.Mapping)}
}
func ionexRequestToC(v IonexInstantSlantRequest) C.SidereonIonexInstantSlantRequest {
	return C.SidereonIonexInstantSlantRequest{lat_deg: C.double(v.LatDeg), lon_deg: C.double(v.LonDeg), azimuth_deg: C.double(v.AzimuthDeg), elevation_deg: C.double(v.ElevationDeg), epoch: clockEpochToC(v.Epoch), frequency_hz: C.double(v.FrequencyHz)}
}
func slantRowFromC(v C.SidereonIonexInstantSlantRowResult) IonexInstantSlantRow {
	return IonexInstantSlantRow{IsOK: bool(v.is_ok), Status: uint32(v.status), Evaluation: ionexEvaluationFromC(v.evaluation), Error: slantErrorFromC(v.error), EpochError: epochErrorFromC(v.epoch_error)}
}
func defaultIonexPolicy(coverage uint32) IonexSlantPolicy {
	return IonexSlantPolicy{Coverage: coverage, MissingNodes: 0, Mapping: 1}
}

func (i *Ionex) SlantDelayAtInstant(lat, lon, azimuth, elevation float64, epoch NativeClockEpoch, frequencyHz float64) (float64, IonexEpochError, error) {
	if i == nil || i.handle == nil {
		return 0, IonexEpochError{}, ErrClosed
	}
	var delay C.double
	var epochError C.SidereonIonexEpochError
	err := i.handle.with(func(p unsafe.Pointer) error {
		ce := clockEpochToC(epoch)
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ionex_slant_delay_at_instant((*C.SidereonIonex)(p), C.double(lat), C.double(lon), C.double(azimuth), C.double(elevation), &ce, C.double(frequencyHz), &delay, &epochError))
		})
	})
	return float64(delay), epochErrorFromC(epochError), err
}
func (i *Ionex) SlantDelayAtInstantWithPolicy(lat, lon, azimuth, elevation float64, epoch NativeClockEpoch, frequencyHz float64, policy IonexSlantPolicy) (IonexInstantSlantRow, error) {
	if i == nil || i.handle == nil {
		return IonexInstantSlantRow{}, ErrClosed
	}
	var row IonexInstantSlantRow
	err := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			ce := clockEpochToC(epoch)
			var out C.SidereonIonexSlantDelayEvaluation
			var detail C.SidereonIonexSlantError
			var epochDetail C.SidereonIonexEpochError
			status := C.sidereon_ionex_slant_delay_at_instant_with_policy((*C.SidereonIonex)(p), C.double(lat), C.double(lon), C.double(azimuth), C.double(elevation), &ce, C.double(frequencyHz), ionexPolicyToC(policy), &out, &detail, &epochDetail)
			row = IonexInstantSlantRow{IsOK: status == C.SIDEREON_STATUS_OK, Status: uint32(status), Evaluation: ionexEvaluationFromC(out), Error: slantErrorFromC(detail), EpochError: epochErrorFromC(epochDetail)}
			return statusErrorLocked(uint32(status))
		})
	})
	return row, err
}

func nativeIONEXRequests(values []IonexInstantSlantRequest) (unsafe.Pointer, *C.SidereonIonexInstantSlantRequest, error) {
	_, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, nil, err
	}
	if len(values) == 0 {
		return nil, nil, nil
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonIonexInstantSlantRequest{}))
	if err != nil {
		return nil, nil, err
	}
	p := C.malloc(C.size_t(size))
	if p == nil {
		return nil, nil, errors.New("sidereon: unable to allocate IONEX instant requests")
	}
	out := unsafe.Slice((*C.SidereonIonexInstantSlantRequest)(p), len(values))
	for j, v := range values {
		out[j] = ionexRequestToC(v)
	}
	return p, &out[0], nil
}
func (i *Ionex) SlantDelayResultsAtInstants(values []IonexInstantSlantRequest, policy IonexSlantPolicy) ([]IonexInstantSlantRow, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	var rows []IonexInstantSlantRow
	err := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			requests, ptr, err := nativeIONEXRequests(values)
			if err != nil {
				return err
			}
			if requests != nil {
				defer C.free(requests)
			}
			var outMem unsafe.Pointer
			if len(values) > 0 {
				bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonIonexInstantSlantRowResult{}))
				if err != nil {
					return err
				}
				outMem = C.calloc(1, C.size_t(bytes))
				if outMem == nil {
					return errors.New("sidereon: unable to allocate IONEX instant rows")
				}
				defer C.free(outMem)
			}
			status := C.sidereon_ionex_slant_delay_results_at_instants((*C.SidereonIonex)(p), ptr, C.size_t(len(values)), ionexPolicyToC(policy), (*C.SidereonIonexInstantSlantRowResult)(outMem))
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			crows := unsafe.Slice((*C.SidereonIonexInstantSlantRowResult)(outMem), len(values))
			rows = make([]IonexInstantSlantRow, len(values))
			for j := range rows {
				rows[j] = slantRowFromC(crows[j])
			}
			return nil
		})
	})
	runtime.KeepAlive(values)
	return rows, err
}
func (i *Ionex) SlantDelayResultsAtInstantsOwned(values []IonexInstantSlantRequest, policy IonexSlantPolicy) (IonexInstantSlantResultList, error) {
	if i == nil || i.handle == nil {
		return IonexInstantSlantResultList{}, ErrClosed
	}
	var list *C.SidereonIonexInstantSlantResultList
	var rows []IonexInstantSlantRow
	err := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			requests, ptr, err := nativeIONEXRequests(values)
			if err != nil {
				return err
			}
			if requests != nil {
				defer C.free(requests)
			}
			status := C.sidereon_ionex_slant_delay_results_at_instants_owned((*C.SidereonIonex)(p), ptr, C.size_t(len(values)), ionexPolicyToC(policy), &list)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			if list == nil {
				return missingNativeHandle("IONEX instant result list")
			}
			defer C.sidereon_ionex_instant_slant_result_list_free(list)
			var count C.size_t
			if err := statusErrorLocked(uint32(C.sidereon_ionex_instant_slant_result_list_count(list, &count))); err != nil {
				return err
			}
			n, err := sizeTToInt(count, "IONEX instant result count")
			if err != nil {
				return err
			}
			if n != len(values) {
				return errors.New("sidereon: native IONEX instant result count differs from input")
			}
			rows = make([]IonexInstantSlantRow, n)
			for j := range rows {
				index, err := cSize(j, "IONEX instant row index")
				if err != nil {
					return err
				}
				var value C.SidereonIonexInstantSlantRowResult
				if err := statusErrorLocked(uint32(C.sidereon_ionex_instant_slant_result_get_row(list, index, &value))); err != nil {
					return err
				}
				rows[j] = slantRowFromC(value)
				message, err := copyNativeBytes("IONEX instant row message", func(out *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_ionex_instant_slant_result_get_message(list, index, out, l, w, r)
				})
				if err != nil {
					return err
				}
				mapping, err := copyNativeBytes("IONEX instant mapping code", func(out *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_ionex_instant_slant_result_get_mapping_code(list, index, out, l, w, r)
				})
				if err != nil {
					return err
				}
				rows[j].Message = string(message)
				rows[j].MappingCode = string(mapping)
			}
			return nil
		})
	})
	runtime.KeepAlive(values)
	return IonexInstantSlantResultList{Rows: rows}, err
}
