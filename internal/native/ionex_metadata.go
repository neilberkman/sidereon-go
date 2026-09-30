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
	"unsafe"
)

type TecSample struct {
	TimeScale                                  uint32
	EpochJ2000S, LatDeg, LonDeg, VTECTECU      float64
	VTECPresent, VTECPresenceKnown, RMSPresent bool
	RMSTECU                                    float64
	EpochJ2000WholeS                           int64
	EpochJ2000WholeSPresent                    bool
	HeightOffsetKm                             float64
	HeightOffsetPresent                        bool
}
type TecGridSamples struct {
	TimeScale                                     uint32
	MapEpochsJ2000S                               []float64
	LatNodesDeg                                   []float64
	LonNodesDeg                                   []float64
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	TECMAPsTECU                                   []float64
	TECMAPsPresent                                []bool
	RMSPresent                                    bool
	RMSMAPsTECU                                   []float64
	RMSMAPsPresent                                []bool
	HeightPresent                                 bool
	HeightMapsKm                                  []float64
	HeightMapsPresent                             []bool
	Header                                        *IonexHeaderMetadata
}
type IonexHeaderMetadata struct {
	Version                                                float64
	Date, Program, RunBy, SatelliteSystem, ObservablesUsed string
	HasSatelliteCount, HasStationCount, HasMapsInFile      bool
	SatelliteCount, StationCount, MapsInFile               uint32
	ElevationCutoffDeg                                     float64
	IntervalS                                              uint32
	MappingDeclaration, MappingFunction                    uint32
	MappingFunctionCode                                    string
	Descriptions, Comments                                 []string
}

// TecSamplesOutcome is the detached fixed-width and text result of a sample build.
type TecSamplesOutcome struct {
	// IsOK reports whether construction produced a product.
	IsOK bool
	// HasIndex reports whether the error names an input entry.
	HasIndex bool
	// HasAxisValue reports whether AxisValue carries a refused coordinate or step.
	HasAxisValue bool
	// Status is the native status paired with the outcome.
	Status uint32
	// Kind identifies the typed construction failure.
	Kind uint32
	// Input identifies the caller buffer named by Index.
	Input uint32
	// Index is the zero-based input index when HasIndex is true.
	Index uint64
	// NodeCount is the short-axis length for a too-few-nodes failure.
	NodeCount uint64
	// ValueCount is the supplied value count for a count mismatch.
	ValueCount uint64
	// ExpectedValueCount is the axis-implied count for a count mismatch.
	ExpectedValueCount uint64
	// AxisValue is the refused coordinate or step when HasAxisValue is true.
	AxisValue float64
	// Message is the detached native detail string.
	Message string
}
type TecGridSamplesInfo struct {
	MapEpochCount, LatNodeCount, LonNodeCount     int
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	RMSPresent                                    bool
	TECMAPValueCount, RMSMAPValueCount            int
}

// IonexWarning is a copied parser finding; text remains valid after parse release.
type IonexWarning struct {
	// Kind is the native warning kind.
	Kind uint32
	// Line is the one-based source line.
	Line uint64
	// MapNumber is the associated map number.
	MapNumber uint64
	// SetByLine identifies the exponent source line.
	SetByLine uint64
	// DeclaredCount is a count from the source header.
	DeclaredCount uint64
	// TECMapCount is the number of TEC maps found.
	TECMapCount uint64
	// AllMapCount is the number of all map records found.
	AllMapCount uint64
	// DeclaredIntervalS is the header interval in seconds.
	DeclaredIntervalS uint32
	// ActualSpacingS is the observed interval in seconds.
	ActualSpacingS int64
	// Exponent is the exponent in effect.
	Exponent int32
	// LatDeg is the associated latitude in degrees.
	LatDeg float64
	// LonDeg is the associated longitude in degrees.
	LonDeg float64
	// HasEpochs reports whether epoch values are present.
	HasEpochs bool
	// DeclaredEpochJ2000S is the declared epoch in seconds since J2000.
	DeclaredEpochJ2000S float64
	// MapEpochJ2000S is the map epoch in seconds since J2000.
	MapEpochJ2000S float64
	// DeclaredEpochJ2000WholeS preserves the exact declared whole seconds.
	DeclaredEpochJ2000WholeS int64
	// MapEpochJ2000WholeS preserves the exact map whole seconds.
	MapEpochJ2000WholeS int64
	// Label is the detached warning label.
	Label string
	// Message is the detached warning detail.
	Message string
}
type IonexSlantDelayEvaluation struct {
	DelayM                                           float64
	Status, CoverageError                            uint32
	IsValid, HasHeld, HasDegraded, HasAssumedMapping bool
	Gap                                              IonexNodeGap
	AssumedMapping                                   uint32
}
type IonexNodeGap struct {
	HasGap         bool
	Earlier, Later [4]bool
}
type Ionex struct {
	_      noCopy
	handle *positioningHandle
}

const (
	IONEXCoveragePolicyStrictValue             = uint32(C.SIDEREON_IONEX_COVERAGE_POLICY_STRICT)
	IONEXCoveragePolicyHoldValue               = uint32(C.SIDEREON_IONEX_COVERAGE_POLICY_HOLD)
	IONEXSlantDelayStatusValidValue            = uint32(0)
	IONEXSlantDelayStatusHeldValue             = uint32(1)
	IONEXCoverageErrorNoneValue                = uint32(C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_NONE)
	IONEXCoverageErrorEpochBeforeFirstMapValue = uint32(C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_EPOCH_BEFORE_FIRST_MAP)
	IONEXCoverageErrorEpochAfterLastMapValue   = uint32(C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_EPOCH_AFTER_LAST_MAP)
	IONEXCoverageErrorLatitudeValue            = uint32(C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_LATITUDE_OUT_OF_RANGE)
	IONEXCoverageErrorLongitudeValue           = uint32(C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_LONGITUDE_OUT_OF_RANGE)
)

func releaseIonex(p unsafe.Pointer) { C.sidereon_ionex_free((*C.SidereonIonex)(p)) }
func sampleFromC(v C.SidereonTecSample) (TecSample, error) {
	if err := validTimeScale(uint32(v.time_scale)); err != nil {
		return TecSample{}, err
	}
	return TecSample{TimeScale: uint32(v.time_scale), EpochJ2000S: float64(v.epoch_j2000_s), EpochJ2000WholeS: int64(v.epoch_j2000_whole_s), EpochJ2000WholeSPresent: bool(v.has_epoch_j2000_whole_s), LatDeg: float64(v.lat_deg), LonDeg: float64(v.lon_deg), VTECTECU: float64(v.vtec_tecu), VTECPresent: bool(v.has_vtec_tecu), VTECPresenceKnown: true, RMSPresent: bool(v.has_rms_tecu), RMSTECU: float64(v.rms_tecu), HeightOffsetKm: float64(v.height_offset_km), HeightOffsetPresent: bool(v.has_height_offset_km)}, nil
}

func ParseIONEX(data []byte) (*Ionex, error) {
	var out *C.SidereonIonex
	var e error
	withCThread(func() {
		p, x := copyNativeInput(data)
		if x != nil {
			e = x
			return
		}
		defer freeNativeInput(p)
		e = callStatus(func() uint32 { return uint32(C.sidereon_ionex_parse((*C.uint8_t)(p), C.size_t(len(data)), &out)) })
		if e != nil && out != nil {
			C.sidereon_ionex_free(out)
			out = nil
		}
	})
	if e != nil {
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native IONEX parser returned no handle")
	}
	return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, nil
}

// ParseIONEXWithWarnings retains parser findings alongside the parsed product.
func ParseIONEXWithWarnings(data []byte) (*Ionex, []IonexWarning, error) {
	var out *C.SidereonIonex
	var warnings *C.SidereonIonexWarningList
	var values []IonexWarning
	var e error
	withCThread(func() {
		p, x := copyNativeInput(data)
		if x != nil {
			e = x
			return
		}
		defer freeNativeInput(p)
		e = callStatus(func() uint32 {
			return uint32(C.sidereon_ionex_parse_with_warnings((*C.uint8_t)(p), C.size_t(len(data)), &out, &warnings))
		})
		if e != nil {
			if out != nil {
				C.sidereon_ionex_free(out)
				out = nil
			}
			if warnings != nil {
				C.sidereon_ionex_warning_list_free(warnings)
				warnings = nil
			}
			return
		}
		if out == nil {
			e = errors.New("sidereon: native IONEX parser returned no handle")
			return
		}
		if warnings == nil {
			e = missingNativeHandle("IONEX warning list")
			return
		}
		defer func() { C.sidereon_ionex_warning_list_free(warnings); warnings = nil }()
		var n C.size_t
		if e = callStatus(func() uint32 { return uint32(C.sidereon_ionex_warning_list_count(warnings, &n)) }); e != nil {
			return
		}
		count, err := sizeTToInt(n, "IONEX warning count")
		if err != nil {
			e = err
			return
		}
		values = make([]IonexWarning, count)
		for j := range values {
			idx, err := cSize(j, "IONEX warning index")
			if err != nil {
				e = err
				return
			}
			var info C.SidereonIonexWarningInfo
			if e = callStatus(func() uint32 { return uint32(C.sidereon_ionex_warning_get_info(warnings, idx, &info)) }); e != nil {
				return
			}
			label, err := copyNativeBytesLocked("IONEX warning label", func(o *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_ionex_warning_get_label(warnings, idx, o, l, w, r)
			})
			if err != nil {
				e = err
				return
			}
			message, err := copyNativeBytesLocked("IONEX warning message", func(o *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_ionex_warning_get_message(warnings, idx, o, l, w, r)
			})
			if err != nil {
				e = err
				return
			}
			values[j] = IonexWarning{Kind: uint32(info.kind), Line: uint64(info.line), MapNumber: uint64(info.map_number), SetByLine: uint64(info.set_by_line), DeclaredCount: uint64(info.declared_count), TECMapCount: uint64(info.tec_map_count), AllMapCount: uint64(info.all_map_count), DeclaredIntervalS: uint32(info.declared_interval_s), ActualSpacingS: int64(info.actual_spacing_s), Exponent: int32(info.exponent), LatDeg: float64(info.lat_deg), LonDeg: float64(info.lon_deg), HasEpochs: bool(info.has_epochs), DeclaredEpochJ2000S: float64(info.declared_epoch_j2000_s), MapEpochJ2000S: float64(info.maps_epoch_j2000_s), DeclaredEpochJ2000WholeS: int64(info.declared_epoch_j2000_whole_s), MapEpochJ2000WholeS: int64(info.maps_epoch_j2000_whole_s), Label: string(label), Message: string(message)}
		}
	})
	if e != nil {
		withCThread(func() {
			if out != nil {
				C.sidereon_ionex_free(out)
			}
			if warnings != nil {
				C.sidereon_ionex_warning_list_free(warnings)
			}
		})
		return nil, nil, e
	}
	return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, values, nil
}

// SkippedRecords reports non-map records the parser ignored.
func (i *Ionex) SkippedRecords() (int, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	e := i.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_ionex_skipped_records((*C.SidereonIonex)(p), &n)) })
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(n, "IONEX skipped record count")
}

func copyIONEXBools(i *Ionex, label string, call func(*C.SidereonIonex, *C.bool, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]bool, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.bool
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	err := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			if err := statusErrorLocked(uint32(call((*C.SidereonIonex)(p), nil, 0, &written, &required))); err != nil {
				return err
			}
			n, err := validateNativeQuery(label, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err = allocNativeArray(n, unsafe.Sizeof(C.bool(false)))
			if err != nil {
				return err
			}
			raw = unsafe.Slice((*C.bool)(memory), n)
			var output *C.bool
			if n > 0 {
				output = &raw[0]
			}
			written, required = 0, 0
			if err := statusErrorLocked(uint32(call((*C.SidereonIonex)(p), output, C.size_t(n), &written, &required))); err != nil {
				return err
			}
			_, err = validateTwoPassCounts(label, n, n, uint64(written), uint64(required))
			return err
		})
	})
	if err != nil {
		return nil, err
	}
	out := make([]bool, len(raw))
	for j := range raw {
		out[j] = bool(raw[j])
	}
	return out, nil
}

// TECMapPresence reports which flattened VTEC cells were present in the file.
func (i *Ionex) TECMapPresence() ([]bool, error) {
	return copyIONEXBools(i, "IONEX TEC presence", func(p *C.SidereonIonex, o *C.bool, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_tec_presence(p, o, n, w, r)
	})
}

// RMSMapPresence reports which flattened RMS cells were present in the file.
func (i *Ionex) RMSMapPresence() ([]bool, error) {
	return copyIONEXBools(i, "IONEX RMS presence", func(p *C.SidereonIonex, o *C.bool, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_rms_presence(p, o, n, w, r)
	})
}

// HeightMapPresence reports which per-map shell-height values were present.
func (i *Ionex) HeightMapPresence() ([]bool, error) {
	return copyIONEXBools(i, "IONEX shell-height presence", func(p *C.SidereonIonex, o *C.bool, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_height_presence(p, o, n, w, r)
	})
}
func ionexHeaderString(header *C.SidereonIonexHeader, field, label string, index C.size_t) (string, error) {
	bytes, err := copyNativeBytesLocked(label, func(out *C.uint8_t, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		switch field {
		case "date":
			return C.sidereon_ionex_header_get_date(header, out, len, written, required)
		case "description":
			return C.sidereon_ionex_header_get_description(header, index, out, len, written, required)
		case "mapping":
			return C.sidereon_ionex_header_get_mapping_function_code(header, out, len, written, required)
		case "observables":
			return C.sidereon_ionex_header_get_observables_used(header, out, len, written, required)
		case "program":
			return C.sidereon_ionex_header_get_program(header, out, len, written, required)
		case "run_by":
			return C.sidereon_ionex_header_get_run_by(header, out, len, written, required)
		case "satellite_system":
			return C.sidereon_ionex_header_get_satellite_system(header, out, len, written, required)
		case "comment":
			return C.sidereon_ionex_header_get_comment(header, index, out, len, written, required)
		default:
			return C.SIDEREON_STATUS_INVALID_ARGUMENT
		}
	})
	return string(bytes), err
}
func (i *Ionex) HeaderMetadata() (IonexHeaderMetadata, error) {
	if i == nil || i.handle == nil {
		return IonexHeaderMetadata{}, ErrClosed
	}
	var result IonexHeaderMetadata
	err := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var header *C.SidereonIonexHeader
			if err := statusErrorLocked(uint32(C.sidereon_ionex_get_header((*C.SidereonIonex)(p), &header))); err != nil {
				return err
			}
			if header == nil {
				return missingNativeHandle("IONEX header")
			}
			defer C.sidereon_ionex_header_free(header)
			text := func(field, label string, index C.size_t) (string, error) {
				return ionexHeaderString(header, field, label, index)
			}
			var err error
			if result.Version, err = ionexHeaderVersion(header); err != nil {
				return err
			}
			if result.Date, err = text("date", "IONEX header date", 0); err != nil {
				return err
			}
			if result.Program, err = text("program", "IONEX header program", 0); err != nil {
				return err
			}
			if result.RunBy, err = text("run_by", "IONEX header agency", 0); err != nil {
				return err
			}
			if result.SatelliteSystem, err = text("satellite_system", "IONEX satellite system", 0); err != nil {
				return err
			}
			if result.ObservablesUsed, err = text("observables", "IONEX observables", 0); err != nil {
				return err
			}
			if result.MappingFunctionCode, err = text("mapping", "IONEX mapping function", 0); err != nil {
				return err
			}
			var decl C.enum_SidereonIonexMappingDeclarationKind
			var mapping C.enum_SidereonIonexMappingFunctionKind
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_get_mapping_declaration(header, &decl, &mapping))); err != nil {
				return err
			}
			result.MappingDeclaration, result.MappingFunction = uint32(decl), uint32(mapping)
			var satellite, station, maps C.uint32_t
			var hasSatellite, hasStation, hasMaps C.bool
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_get_satellite_count(header, &satellite, &hasSatellite))); err != nil {
				return err
			}
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_get_station_count(header, &station, &hasStation))); err != nil {
				return err
			}
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_get_maps_in_file(header, &maps, &hasMaps))); err != nil {
				return err
			}
			result.SatelliteCount, result.HasSatelliteCount = uint32(satellite), bool(hasSatellite)
			result.StationCount, result.HasStationCount = uint32(station), bool(hasStation)
			result.MapsInFile, result.HasMapsInFile = uint32(maps), bool(hasMaps)
			var cutoff C.double
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_get_elevation_cutoff_deg(header, &cutoff))); err != nil {
				return err
			}
			result.ElevationCutoffDeg = float64(cutoff)
			var interval C.uint32_t
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_get_interval_s(header, &interval))); err != nil {
				return err
			}
			result.IntervalS = uint32(interval)
			var count C.size_t
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_description_count(header, &count))); err != nil {
				return err
			}
			descriptionCount, err := sizeTToInt(count, "IONEX description count")
			if err != nil {
				return err
			}
			result.Descriptions = make([]string, descriptionCount)
			for n := range result.Descriptions {
				result.Descriptions[n], err = text("description", "IONEX description", C.size_t(n))
				if err != nil {
					return err
				}
			}
			if err = statusErrorLocked(uint32(C.sidereon_ionex_header_comment_count(header, &count))); err != nil {
				return err
			}
			commentCount, err := sizeTToInt(count, "IONEX comment count")
			if err != nil {
				return err
			}
			result.Comments = make([]string, commentCount)
			for n := range result.Comments {
				result.Comments[n], err = text("comment", "IONEX comment", C.size_t(n))
				if err != nil {
					return err
				}
			}
			return nil
		})
	})
	if err != nil {
		return IonexHeaderMetadata{}, err
	}
	return result, nil
}
func ionexHeaderVersion(header *C.SidereonIonexHeader) (float64, error) {
	var value C.double
	err := statusErrorLocked(uint32(C.sidereon_ionex_header_get_version(header, &value)))
	return float64(value), err
}
func (i *Ionex) Close() error {
	if i == nil || i.handle == nil {
		return nil
	}
	return i.handle.close()
}
func (i *Ionex) EpochCount() (int, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	var n C.size_t
	e := i.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_ionex_epoch_count((*C.SidereonIonex)(p), &n)) })
	})
	if e != nil {
		return 0, e
	}
	return sizeTToInt(n, "IONEX epoch count")
}
func (i *Ionex) Exponent() (int32, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	var n C.int32_t
	e := i.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_ionex_exponent((*C.SidereonIonex)(p), &n)) })
	})
	return int32(n), e
}
func copyIONEXDoubles(i *Ionex, label string, call func(*C.SidereonIonex, *C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]float64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.double
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			s := call((*C.SidereonIonex)(p), nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery(label, uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.double(0)))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.double)(memory), n)
			w, r = 0, 0
			var q *C.double
			if n > 0 {
				q = &raw[0]
			}
			s = call((*C.SidereonIonex)(p), q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts(label, n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]float64, len(raw))
	for j := range raw {
		out[j] = float64(raw[j])
	}
	return out, nil
}
func (i *Ionex) LatNodesDeg() ([]float64, error) {
	return copyIONEXDoubles(i, "IONEX latitude nodes", func(p *C.SidereonIonex, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_lat_nodes_deg((*C.SidereonIonex)(p), o, n, w, r)
	})
}
func (i *Ionex) LonNodesDeg() ([]float64, error) {
	return copyIONEXDoubles(i, "IONEX longitude nodes", func(p *C.SidereonIonex, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_lon_nodes_deg((*C.SidereonIonex)(p), o, n, w, r)
	})
}
func (i *Ionex) MapEpochsJ2000S() ([]int64, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.int64_t
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			s := C.sidereon_ionex_map_epochs_j2000_s((*C.SidereonIonex)(p), nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("IONEX map epochs", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.int64_t(0)))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.int64_t)(memory), n)
			w, r = 0, 0
			var q *C.int64_t
			if n > 0 {
				q = &raw[0]
			}
			s = C.sidereon_ionex_map_epochs_j2000_s((*C.SidereonIonex)(p), q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("IONEX map epochs", n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]int64, len(raw))
	for j := range raw {
		out[j] = int64(raw[j])
	}
	return out, nil
}
func (i *Ionex) ToIONEXText() ([]byte, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	var out []byte
	var e error
	e = i.handle.with(func(p unsafe.Pointer) error {
		withCThread(func() {
			out, e = copyNativeBytesLocked("IONEX text", func(b *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_ionex_to_ionex_text((*C.SidereonIonex)(p), b, n, w, r)
			})
		})
		return e
	})
	return out, e
}
func (i *Ionex) SlantDelay(lat, lon, azimuth, elevation float64, epochJ2000S int64, frequencyHz float64) (float64, error) {
	if i == nil || i.handle == nil {
		return 0, ErrClosed
	}
	var x C.double
	e := i.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ionex_slant_delay((*C.SidereonIonex)(p), C.double(lat), C.double(lon), C.double(azimuth), C.double(elevation), C.int64_t(epochJ2000S), C.double(frequencyHz), &x))
		})
	})
	return float64(x), e
}
func (i *Ionex) SlantDelayWithPolicy(lat, lon, azimuth, elevation float64, epochJ2000S int64, frequencyHz float64, policy uint32) (IonexSlantDelayEvaluation, error) {
	if i == nil || i.handle == nil {
		return IonexSlantDelayEvaluation{}, ErrClosed
	}
	if policy != IONEXCoveragePolicyStrictValue && policy != IONEXCoveragePolicyHoldValue {
		return IonexSlantDelayEvaluation{}, invalidArgument("IONEX coverage policy is not defined")
	}
	var x C.SidereonIonexSlantDelayEvaluation
	e := i.handle.with(func(p unsafe.Pointer) error {
		var detail C.SidereonIonexSlantError
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ionex_slant_delay_with_coverage_policy((*C.SidereonIonex)(p), C.double(lat), C.double(lon), C.double(azimuth), C.double(elevation), C.int64_t(epochJ2000S), C.double(frequencyHz), C.uint32_t(policy), &x, &detail))
		})
	})
	if e != nil {
		return IonexSlantDelayEvaluation{}, e
	}
	if x.status.coverage_error < C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_NONE || x.status.coverage_error > C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_LONGITUDE_OUT_OF_RANGE {
		return IonexSlantDelayEvaluation{}, invalidArgument("native IONEX coverage error is not defined")
	}
	return ionexEvaluationFromC(x), e
}

func ionexEvaluationFromC(x C.SidereonIonexSlantDelayEvaluation) IonexSlantDelayEvaluation {
	status := IONEXSlantDelayStatusValidValue
	if bool(x.status.has_held) {
		status = IONEXSlantDelayStatusHeldValue
	}
	value := IonexSlantDelayEvaluation{DelayM: float64(x.delay_m), Status: status, CoverageError: uint32(x.status.coverage_error), IsValid: bool(x.status.is_valid), HasHeld: bool(x.status.has_held), HasDegraded: bool(x.status.has_degraded), HasAssumedMapping: bool(x.status.has_assumed_mapping), AssumedMapping: uint32(x.status.assumed_mapping)}
	value.Gap.HasGap = bool(x.status.gap.has_gap)
	for index := 0; index < 4; index++ {
		value.Gap.Earlier[index] = bool(x.status.gap.earlier.missing[index])
		value.Gap.Later[index] = bool(x.status.gap.later.missing[index])
	}
	return value
}

func checkedIONEXDimensions(values ...int) (int, error) {
	total := 1
	for _, v := range values {
		if v < 0 {
			return 0, invalidArgument("negative dimension")
		}
		if total != 0 && v > (int(^uint(0)>>1))/total {
			return 0, invalidArgument("array dimension multiplication overflows")
		}
		total *= v
	}
	return total, nil
}
func cDoubleArray(values []float64) (unsafe.Pointer, error) {
	if len(values) == 0 {
		return nil, nil
	}
	size, e := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.double(0)))
	if e != nil {
		return nil, e
	}
	p := C.malloc(C.size_t(size))
	if p == nil {
		return nil, errors.New("sidereon: unable to allocate native double array")
	}
	dst := unsafe.Slice((*C.double)(p), len(values))
	for j, v := range values {
		dst[j] = C.double(v)
	}
	return p, nil
}
func cBoolArray(values []bool) (unsafe.Pointer, error) {
	if len(values) == 0 {
		return nil, nil
	}
	size, e := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.bool(false)))
	if e != nil {
		return nil, e
	}
	p := C.malloc(C.size_t(size))
	if p == nil {
		return nil, errors.New("sidereon: unable to allocate native bool array")
	}
	dst := unsafe.Slice((*C.bool)(p), len(values))
	for j, v := range values {
		dst[j] = C.bool(v)
	}
	return p, nil
}

const (
	ionexHeaderSetDate = iota
	ionexHeaderSetProgram
	ionexHeaderSetRunBy
	ionexHeaderSetSatelliteSystem
	ionexHeaderSetObservables
	ionexHeaderAddDescription
	ionexHeaderAddComment
)

func setIONEXHeaderText(header *C.SidereonIonexHeader, text string, setter int) error {
	n, err := checkedNativeSize(len(text))
	if err != nil {
		return err
	}
	p, err := copyNativeInput([]byte(text))
	if err != nil {
		return err
	}
	defer freeNativeInput(p)
	bytes := (*C.uint8_t)(p)
	return callStatus(func() uint32 {
		switch setter {
		case ionexHeaderSetDate:
			return uint32(C.sidereon_ionex_header_set_date(header, bytes, n))
		case ionexHeaderSetProgram:
			return uint32(C.sidereon_ionex_header_set_program(header, bytes, n))
		case ionexHeaderSetRunBy:
			return uint32(C.sidereon_ionex_header_set_run_by(header, bytes, n))
		case ionexHeaderSetSatelliteSystem:
			return uint32(C.sidereon_ionex_header_set_satellite_system(header, bytes, n))
		case ionexHeaderSetObservables:
			return uint32(C.sidereon_ionex_header_set_observables_used(header, bytes, n))
		case ionexHeaderAddDescription:
			return uint32(C.sidereon_ionex_header_add_description(header, bytes, n))
		case ionexHeaderAddComment:
			return uint32(C.sidereon_ionex_header_add_comment(header, bytes, n))
		default:
			return uint32(C.SIDEREON_STATUS_INVALID_ARGUMENT)
		}
	})
}
func validateIONEXHeaderMetadata(value *IonexHeaderMetadata) error {
	if value == nil {
		return nil
	}
	if value.MappingDeclaration != 0 && value.MappingDeclaration != 1 {
		return invalidArgument("IONEX mapping declaration kind is not defined")
	}
	return nil
}

func buildIONEXHeaderMetadata(value *IonexHeaderMetadata) (*C.SidereonIonexHeader, error) {
	if value == nil {
		return nil, nil
	}
	if err := validateIONEXHeaderMetadata(value); err != nil {
		return nil, err
	}
	var header *C.SidereonIonexHeader
	if err := callStatus(func() uint32 { return uint32(C.sidereon_ionex_header_new(&header)) }); err != nil {
		return nil, err
	}
	if header == nil {
		return nil, errors.New("sidereon: native IONEX header constructor returned no handle")
	}
	freeOnError := true
	defer func() {
		if freeOnError {
			withCThread(func() { C.sidereon_ionex_header_free(header) })
		}
	}()
	if err := callStatus(func() uint32 { return uint32(C.sidereon_ionex_header_set_version(header, C.double(value.Version))) }); err != nil {
		return nil, err
	}
	if err := setIONEXHeaderText(header, value.Date, ionexHeaderSetDate); err != nil {
		return nil, err
	}
	if err := setIONEXHeaderText(header, value.Program, ionexHeaderSetProgram); err != nil {
		return nil, err
	}
	if err := setIONEXHeaderText(header, value.RunBy, ionexHeaderSetRunBy); err != nil {
		return nil, err
	}
	if err := setIONEXHeaderText(header, value.SatelliteSystem, ionexHeaderSetSatelliteSystem); err != nil {
		return nil, err
	}
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_ionex_header_set_satellite_count(header, C.uint32_t(value.SatelliteCount), C.bool(value.HasSatelliteCount)))
	}); err != nil {
		return nil, err
	}
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_ionex_header_set_station_count(header, C.uint32_t(value.StationCount), C.bool(value.HasStationCount)))
	}); err != nil {
		return nil, err
	}
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_ionex_header_set_elevation_cutoff_deg(header, C.double(value.ElevationCutoffDeg)))
	}); err != nil {
		return nil, err
	}
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_ionex_header_set_interval_s(header, C.uint32_t(value.IntervalS)))
	}); err != nil {
		return nil, err
	}
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_ionex_header_set_maps_in_file(header, C.uint32_t(value.MapsInFile), C.bool(value.HasMapsInFile)))
	}); err != nil {
		return nil, err
	}
	if err := setIONEXHeaderText(header, value.ObservablesUsed, ionexHeaderSetObservables); err != nil {
		return nil, err
	}
	if value.MappingDeclaration == 1 {
		if err := callStatus(func() uint32 { return uint32(C.sidereon_ionex_header_clear_mapping_function(header)) }); err != nil {
			return nil, err
		}
	} else if err := setIONEXHeaderMappingFunction(header, value.MappingFunction, value.MappingFunctionCode); err != nil {
		return nil, err
	}
	for _, text := range value.Descriptions {
		if err := setIONEXHeaderText(header, text, ionexHeaderAddDescription); err != nil {
			return nil, err
		}
	}
	for _, text := range value.Comments {
		if err := setIONEXHeaderText(header, text, ionexHeaderAddComment); err != nil {
			return nil, err
		}
	}
	freeOnError = false
	return header, nil
}
func setIONEXHeaderMappingFunction(header *C.SidereonIonexHeader, kind uint32, code string) error {
	if kind != 3 {
		code = ""
	}
	n, err := checkedNativeSize(len(code))
	if err != nil {
		return err
	}
	p, err := copyNativeInput([]byte(code))
	if err != nil {
		return err
	}
	defer freeNativeInput(p)
	return callStatus(func() uint32 {
		return uint32(C.sidereon_ionex_header_set_mapping_function(header, C.uint32_t(kind), (*C.uint8_t)(p), n))
	})
}
func buildIONEXFromGrid(s TecGridSamples) (*Ionex, error) {
	value, _, err := buildIONEXFromGridWithOutcome(s, false)
	return value, err
}

func buildIONEXFromGridWithOutcome(s TecGridSamples, captureOutcome bool) (*Ionex, TecSamplesOutcome, error) {
	var emptyOutcome TecSamplesOutcome
	if err := validateIONEXHeaderMetadata(s.Header); err != nil {
		return nil, emptyOutcome, err
	}
	if err := validTimeScale(s.TimeScale); err != nil {
		return nil, emptyOutcome, err
	}
	expected, e := checkedIONEXDimensions(len(s.MapEpochsJ2000S), len(s.LatNodesDeg), len(s.LonNodesDeg))
	if e != nil {
		return nil, emptyOutcome, e
	}
	if !captureOutcome && len(s.TECMAPsTECU) != expected {
		return nil, emptyOutcome, invalidArgument("IONEX VTEC map shape does not match axes")
	}
	if !captureOutcome && s.RMSPresent && len(s.RMSMAPsTECU) != expected {
		return nil, emptyOutcome, invalidArgument("IONEX RMS map shape does not match axes")
	}
	// C has no independent count for the VTEC/RMS presence buffers. Require
	// each mask to match its values whenever C may read it. For an outcome call
	// with mismatched values, C returns before reading the mask, preserving its
	// typed count refusal.
	if (len(s.TECMAPsTECU) == expected || !captureOutcome) && len(s.TECMAPsPresent) != 0 && len(s.TECMAPsPresent) != len(s.TECMAPsTECU) {
		return nil, emptyOutcome, invalidArgument("IONEX VTEC presence count does not match values")
	}
	if s.RMSPresent && (len(s.RMSMAPsTECU) == expected || !captureOutcome) && len(s.RMSMAPsPresent) != 0 && len(s.RMSMAPsPresent) != len(s.RMSMAPsTECU) {
		return nil, emptyOutcome, invalidArgument("IONEX RMS presence count does not match values")
	}
	if s.HeightPresent {
		if !captureOutcome && len(s.HeightMapsKm) != expected {
			return nil, emptyOutcome, invalidArgument("IONEX shell-height map count does not match axes")
		}
		// C has no independent count for the height-presence buffer. When the
		// height value count is valid, require an equally sized mask to prevent
		// an out-of-bounds read. For a mismatched height count C returns before
		// reading either buffer, so let the result API retain its typed refusal.
		if (len(s.HeightMapsKm) == expected || !captureOutcome) && len(s.HeightMapsPresent) != 0 && len(s.HeightMapsPresent) != len(s.HeightMapsKm) {
			return nil, emptyOutcome, invalidArgument("IONEX shell-height presence count does not match axes")
		}
	}
	var mem []unsafe.Pointer
	alloc := func(v []float64) (unsafe.Pointer, error) {
		p, x := cDoubleArray(v)
		if p != nil {
			mem = append(mem, p)
		}
		return p, x
	}
	epochs, e := alloc(s.MapEpochsJ2000S)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, emptyOutcome, e
	}
	lats, e := alloc(s.LatNodesDeg)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, emptyOutcome, e
	}
	lons, e := alloc(s.LonNodesDeg)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, emptyOutcome, e
	}
	tec, e := alloc(s.TECMAPsTECU)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, emptyOutcome, e
	}
	var tecPresence unsafe.Pointer
	if len(s.TECMAPsTECU) == expected && len(s.TECMAPsPresent) != 0 {
		tecPresence, e = cBoolArray(s.TECMAPsPresent)
		if tecPresence != nil {
			mem = append(mem, tecPresence)
		}
		if e != nil {
			for _, p := range mem {
				C.free(p)
			}
			return nil, emptyOutcome, e
		}
	}
	rms, e := alloc(s.RMSMAPsTECU)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, emptyOutcome, e
	}
	var rmsPresence unsafe.Pointer
	if s.RMSPresent && len(s.RMSMAPsTECU) == expected && len(s.RMSMAPsPresent) != 0 {
		rmsPresence, e = cBoolArray(s.RMSMAPsPresent)
		if rmsPresence != nil {
			mem = append(mem, rmsPresence)
		}
		if e != nil {
			for _, p := range mem {
				C.free(p)
			}
			return nil, emptyOutcome, e
		}
	}
	var heights, heightPresence unsafe.Pointer
	heightCount := 0
	if s.HeightPresent {
		heightCount = len(s.HeightMapsKm)
		heights, e = alloc(s.HeightMapsKm)
		if e != nil {
			for _, p := range mem {
				C.free(p)
			}
			return nil, emptyOutcome, e
		}
		if len(s.HeightMapsPresent) == heightCount {
			heightPresence, e = cBoolArray(s.HeightMapsPresent)
		}
		if heightPresence != nil {
			mem = append(mem, heightPresence)
		}
		if e != nil {
			for _, p := range mem {
				C.free(p)
			}
			return nil, emptyOutcome, e
		}
	}
	defer func() {
		for _, p := range mem {
			C.free(p)
		}
	}()
	header, e := buildIONEXHeaderMetadata(s.Header)
	if e != nil {
		return nil, emptyOutcome, e
	}
	if header != nil {
		defer withCThread(func() { C.sidereon_ionex_header_free(header) })
	}
	inMem := C.malloc(C.size_t(unsafe.Sizeof(C.SidereonTecGridSamples{})))
	if inMem == nil {
		return nil, emptyOutcome, errors.New("sidereon: unable to allocate native IONEX grid descriptor")
	}
	defer C.free(inMem)
	in := (*C.SidereonTecGridSamples)(inMem)
	*in = C.SidereonTecGridSamples{time_scale: C.uint32_t(s.TimeScale), map_epochs_j2000_s: (*C.double)(epochs), map_epoch_count: C.size_t(len(s.MapEpochsJ2000S)), lat_nodes_deg: (*C.double)(lats), lat_node_count: C.size_t(len(s.LatNodesDeg)), lon_nodes_deg: (*C.double)(lons), lon_node_count: C.size_t(len(s.LonNodesDeg)), dlat_deg: C.double(s.DLatDeg), dlon_deg: C.double(s.DLonDeg), shell_height_km: C.double(s.ShellHeightKm), base_radius_km: C.double(s.BaseRadiusKm), exponent: C.int32_t(s.Exponent), tec_maps_tecu: (*C.double)(tec), tec_maps_present: (*C.bool)(tecPresence), tec_map_value_count: C.size_t(len(s.TECMAPsTECU)), has_rms_maps: C.bool(s.RMSPresent), rms_maps_tecu: (*C.double)(rms), rms_maps_present: (*C.bool)(rmsPresence), rms_map_value_count: C.size_t(len(s.RMSMAPsTECU)), has_height_maps: C.bool(s.HeightPresent), height_maps_km: (*C.double)(heights), height_maps_present: (*C.bool)(heightPresence), height_map_value_count: C.size_t(heightCount), header: header}
	var out *C.SidereonIonex
	var result *C.SidereonTecSamplesResult
	if captureOutcome {
		e = callStatus(func() uint32 { return uint32(C.sidereon_ionex_from_tec_grid_samples_result(in, &out, &result)) })
	} else {
		e = callStatus(func() uint32 { return uint32(C.sidereon_ionex_from_tec_grid_samples(in, &out)) })
	}
	if result != nil {
		defer withCThread(func() { C.sidereon_tec_samples_result_free(result) })
	}
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_ionex_free(out) })
		}
		return nil, emptyOutcome, e
	}
	if captureOutcome {
		outcome, outcomeErr := readTecSamplesOutcome(result)
		if outcomeErr != nil {
			if out != nil {
				withCThread(func() { C.sidereon_ionex_free(out) })
			}
			return nil, emptyOutcome, outcomeErr
		}
		if !outcome.IsOK {
			if out != nil {
				withCThread(func() { C.sidereon_ionex_free(out) })
				return nil, emptyOutcome, errors.New("sidereon: refused IONEX build unexpectedly returned a product")
			}
			return nil, outcome, nil
		}
		if out == nil {
			return nil, emptyOutcome, errors.New("sidereon: successful IONEX result returned no handle")
		}
		return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, outcome, nil
	}
	if out == nil {
		return nil, emptyOutcome, errors.New("sidereon: native IONEX constructor returned no handle")
	}
	return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, emptyOutcome, nil
}
func readTecSamplesOutcome(result *C.SidereonTecSamplesResult) (TecSamplesOutcome, error) {
	if result == nil {
		return TecSamplesOutcome{}, errors.New("sidereon: native IONEX result returned no outcome handle")
	}
	var value TecSamplesOutcome
	err := withCThreadError(func() error {
		var raw C.SidereonTecSamplesOutcome
		if err := statusErrorLocked(uint32(C.sidereon_tec_samples_result_get_outcome(result, &raw))); err != nil {
			return err
		}
		value = TecSamplesOutcome{IsOK: bool(raw.is_ok), Status: uint32(raw.status), Kind: uint32(raw.error.kind), Input: uint32(raw.error.input), HasIndex: bool(raw.error.has_index), Index: uint64(raw.error.index), NodeCount: uint64(raw.error.node_count), ValueCount: uint64(raw.error.value_count), ExpectedValueCount: uint64(raw.error.expected_value_count), HasAxisValue: bool(raw.error.has_axis_value), AxisValue: float64(raw.error.axis_value)}
		message, err := copyNativeBytesLocked("IONEX build outcome message", func(out *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_tec_samples_result_get_message(result, out, n, w, r)
		})
		if err != nil {
			return err
		}
		value.Message = string(message)
		return nil
	})
	return value, err
}
func BuildIONEXFromTECGridSamples(s TecGridSamples) (*Ionex, error) { return buildIONEXFromGrid(s) }

// BuildIONEXFromTECGridSamplesWithOutcome preserves typed data refusals.
func BuildIONEXFromTECGridSamplesWithOutcome(s TecGridSamples) (*Ionex, TecSamplesOutcome, error) {
	return buildIONEXFromGridWithOutcome(s, true)
}
func BuildIONEXFromTECSamples(samples []TecSample, shellHeightKm, baseRadiusKm float64, exponent int32) (*Ionex, error) {
	value, _, err := buildIONEXFromTECSamples(samples, shellHeightKm, baseRadiusKm, exponent, nil, false, false)
	return value, err
}

func BuildIONEXFromTECSamplesWithHeader(samples []TecSample, shellHeightKm, baseRadiusKm float64, exponent int32, value *IonexHeaderMetadata) (*Ionex, error) {
	product, _, err := buildIONEXFromTECSamplesWithHeader(samples, shellHeightKm, baseRadiusKm, exponent, value, false)
	return product, err
}

// BuildIONEXFromTECSamplesWithOutcome preserves typed per-node data refusals.
func BuildIONEXFromTECSamplesWithOutcome(samples []TecSample, shellHeightKm, baseRadiusKm float64, exponent int32) (*Ionex, TecSamplesOutcome, error) {
	return buildIONEXFromTECSamplesWithHeader(samples, shellHeightKm, baseRadiusKm, exponent, nil, true)
}

// BuildIONEXFromTECSamplesWithHeaderOutcome preserves typed data refusals with header input.
func BuildIONEXFromTECSamplesWithHeaderOutcome(samples []TecSample, shellHeightKm, baseRadiusKm float64, exponent int32, value *IonexHeaderMetadata) (*Ionex, TecSamplesOutcome, error) {
	return buildIONEXFromTECSamplesWithHeader(samples, shellHeightKm, baseRadiusKm, exponent, value, true)
}

func buildIONEXFromTECSamplesWithHeader(samples []TecSample, shellHeightKm, baseRadiusKm float64, exponent int32, value *IonexHeaderMetadata, captureOutcome bool) (*Ionex, TecSamplesOutcome, error) {
	header, err := buildIONEXHeaderMetadata(value)
	if err != nil {
		return nil, TecSamplesOutcome{}, err
	}
	if header != nil {
		defer withCThread(func() { C.sidereon_ionex_header_free(header) })
	}
	return buildIONEXFromTECSamples(samples, shellHeightKm, baseRadiusKm, exponent, header, true, captureOutcome)
}

func buildIONEXFromTECSamples(samples []TecSample, shellHeightKm, baseRadiusKm float64, exponent int32, header *C.SidereonIonexHeader, withHeader, captureOutcome bool) (*Ionex, TecSamplesOutcome, error) {
	var emptyOutcome TecSamplesOutcome
	for _, sample := range samples {
		if err := validTimeScale(sample.TimeScale); err != nil {
			return nil, emptyOutcome, err
		}
	}
	if _, e := checkedNativeAllocationSize(len(samples), unsafe.Sizeof(C.SidereonTecSample{})); e != nil {
		return nil, emptyOutcome, e
	}
	var mem unsafe.Pointer
	if len(samples) > 0 {
		mem = C.malloc(C.size_t(len(samples)) * C.size_t(unsafe.Sizeof(C.SidereonTecSample{})))
		if mem == nil {
			return nil, emptyOutcome, errors.New("sidereon: unable to allocate native TEC samples")
		}
		defer C.free(mem)
	}
	raw := unsafe.Slice((*C.SidereonTecSample)(mem), len(samples))
	for j, v := range samples {
		hasVTEC := v.VTECPresent || !v.VTECPresenceKnown
		raw[j] = C.SidereonTecSample{time_scale: C.uint32_t(v.TimeScale), epoch_j2000_s: C.double(v.EpochJ2000S), has_epoch_j2000_whole_s: C.bool(v.EpochJ2000WholeSPresent), epoch_j2000_whole_s: C.int64_t(v.EpochJ2000WholeS), lat_deg: C.double(v.LatDeg), lon_deg: C.double(v.LonDeg), has_vtec_tecu: C.bool(hasVTEC), vtec_tecu: C.double(v.VTECTECU), has_rms_tecu: C.bool(v.RMSPresent), rms_tecu: C.double(v.RMSTECU), has_height_offset_km: C.bool(v.HeightOffsetPresent), height_offset_km: C.double(v.HeightOffsetKm)}
	}
	var p *C.SidereonTecSample
	if len(raw) > 0 {
		p = &raw[0]
	}
	var out *C.SidereonIonex
	var result *C.SidereonTecSamplesResult
	e := callStatus(func() uint32 {
		if withHeader {
			if captureOutcome {
				return uint32(C.sidereon_ionex_from_tec_samples_result(p, C.size_t(len(raw)), C.double(shellHeightKm), C.double(baseRadiusKm), C.int32_t(exponent), header, &out, &result))
			}
			return uint32(C.sidereon_ionex_from_tec_samples_with_header(p, C.size_t(len(raw)), C.double(shellHeightKm), C.double(baseRadiusKm), C.int32_t(exponent), header, &out))
		}
		if captureOutcome {
			return uint32(C.sidereon_ionex_from_tec_samples_result(p, C.size_t(len(raw)), C.double(shellHeightKm), C.double(baseRadiusKm), C.int32_t(exponent), nil, &out, &result))
		}
		return uint32(C.sidereon_ionex_from_tec_samples(p, C.size_t(len(raw)), C.double(shellHeightKm), C.double(baseRadiusKm), C.int32_t(exponent), &out))
	})
	if result != nil {
		defer withCThread(func() { C.sidereon_tec_samples_result_free(result) })
	}
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_ionex_free(out) })
		}
		return nil, emptyOutcome, e
	}
	if captureOutcome {
		outcome, outcomeErr := readTecSamplesOutcome(result)
		if outcomeErr != nil {
			if out != nil {
				withCThread(func() { C.sidereon_ionex_free(out) })
			}
			return nil, emptyOutcome, outcomeErr
		}
		if !outcome.IsOK {
			if out != nil {
				withCThread(func() { C.sidereon_ionex_free(out) })
				return nil, emptyOutcome, errors.New("sidereon: refused IONEX build unexpectedly returned a product")
			}
			return nil, outcome, nil
		}
		if out == nil {
			return nil, emptyOutcome, errors.New("sidereon: successful IONEX result returned no handle")
		}
		return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, outcome, nil
	}
	if out == nil {
		return nil, emptyOutcome, errors.New("sidereon: native IONEX constructor returned no handle")
	}
	return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, emptyOutcome, nil
}
func (i *Ionex) TECSamples() ([]TecSample, error) {
	if i == nil || i.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonTecSample
	var memory unsafe.Pointer
	defer func() {
		if memory != nil {
			C.free(memory)
		}
	}()
	e := i.handle.with(func(p unsafe.Pointer) error {
		return withCThreadError(func() error {
			var w, r C.size_t
			s := C.sidereon_ionex_tec_samples((*C.SidereonIonex)(p), nil, 0, &w, &r)
			if x := statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			n, x := validateNativeQuery("IONEX TEC samples", uint64(w), uint64(r))
			if x != nil {
				return x
			}
			memory, x = allocNativeArray(n, unsafe.Sizeof(C.SidereonTecSample{}))
			if x != nil {
				return x
			}
			raw = unsafe.Slice((*C.SidereonTecSample)(memory), n)
			w, r = 0, 0
			var q *C.SidereonTecSample
			if n > 0 {
				q = &raw[0]
			}
			s = C.sidereon_ionex_tec_samples((*C.SidereonIonex)(p), q, C.size_t(n), &w, &r)
			if x = statusErrorLocked(uint32(s)); x != nil {
				return x
			}
			_, x = validateTwoPassCounts("IONEX TEC samples", n, n, uint64(w), uint64(r))
			return x
		})
	})
	if e != nil {
		return nil, e
	}
	out := make([]TecSample, len(raw))
	for j := range raw {
		value, err := sampleFromC(raw[j])
		if err != nil {
			return nil, err
		}
		out[j] = value
	}
	return out, nil
}
func (i *Ionex) GridInfo() (TecGridSamplesInfo, error) {
	if i == nil || i.handle == nil {
		return TecGridSamplesInfo{}, ErrClosed
	}
	var x C.SidereonTecGridSamplesInfo
	e := i.handle.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 { return uint32(C.sidereon_ionex_tec_grid_samples_info((*C.SidereonIonex)(p), &x)) })
	})
	if e != nil {
		return TecGridSamplesInfo{}, e
	}
	mapEpochCount, e := sizeTToInt(x.map_epoch_count, "IONEX map epoch count")
	if e != nil {
		return TecGridSamplesInfo{}, e
	}
	latNodeCount, e := sizeTToInt(x.lat_node_count, "IONEX latitude node count")
	if e != nil {
		return TecGridSamplesInfo{}, e
	}
	lonNodeCount, e := sizeTToInt(x.lon_node_count, "IONEX longitude node count")
	if e != nil {
		return TecGridSamplesInfo{}, e
	}
	tecCount, e := sizeTToInt(x.tec_map_value_count, "IONEX TEC map value count")
	if e != nil {
		return TecGridSamplesInfo{}, e
	}
	rmsCount, e := sizeTToInt(x.rms_map_value_count, "IONEX RMS map value count")
	if e != nil {
		return TecGridSamplesInfo{}, e
	}
	return TecGridSamplesInfo{MapEpochCount: mapEpochCount, LatNodeCount: latNodeCount, LonNodeCount: lonNodeCount, DLatDeg: float64(x.dlat_deg), DLonDeg: float64(x.dlon_deg), ShellHeightKm: float64(x.shell_height_km), BaseRadiusKm: float64(x.base_radius_km), Exponent: int32(x.exponent), RMSPresent: bool(x.has_rms_maps), TECMAPValueCount: tecCount, RMSMAPValueCount: rmsCount}, nil
}
func (i *Ionex) TECMAPsTECU() ([]float64, error) {
	return copyIONEXDoubles(i, "IONEX VTEC maps", func(p *C.SidereonIonex, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_tec_maps_tecu((*C.SidereonIonex)(p), o, n, w, r)
	})
}
func (i *Ionex) RMSMAPsTECU() ([]float64, error) {
	return copyIONEXDoubles(i, "IONEX RMS maps", func(p *C.SidereonIonex, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_rms_maps_tecu((*C.SidereonIonex)(p), o, n, w, r)
	})
}
func (i *Ionex) HeightMapsKm() ([]float64, error) {
	return copyIONEXDoubles(i, "IONEX shell-height maps", func(p *C.SidereonIonex, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_height_maps_km((*C.SidereonIonex)(p), o, n, w, r)
	})
}
func (i *Ionex) GridEpochsJ2000S() ([]float64, error) {
	return copyIONEXDoubles(i, "IONEX grid epochs", func(p *C.SidereonIonex, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_epochs_j2000_s((*C.SidereonIonex)(p), o, n, w, r)
	})
}
