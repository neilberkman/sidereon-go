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
	TimeScale                             uint32
	EpochJ2000S, LatDeg, LonDeg, VTECTECU float64
	RMSPresent                            bool
	RMSTECU                               float64
}
type TecGridSamples struct {
	TimeScale                                     uint32
	MapEpochsJ2000S                               []float64
	LatNodesDeg                                   []float64
	LonNodesDeg                                   []float64
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	TECMAPsTECU                                   []float64
	RMSPresent                                    bool
	RMSMAPsTECU                                   []float64
}
type TecGridSamplesInfo struct {
	MapEpochCount, LatNodeCount, LonNodeCount     int
	DLatDeg, DLonDeg, ShellHeightKm, BaseRadiusKm float64
	Exponent                                      int32
	RMSPresent                                    bool
	TECMAPValueCount, RMSMAPValueCount            int
}
type IonexSlantDelayEvaluation struct {
	DelayM                float64
	Status, CoverageError uint32
}
type Ionex struct {
	_      noCopy
	handle *positioningHandle
}

const (
	IONEXCoveragePolicyStrictValue             = uint32(C.SIDEREON_IONEX_COVERAGE_POLICY_STRICT)
	IONEXCoveragePolicyHoldValue               = uint32(C.SIDEREON_IONEX_COVERAGE_POLICY_HOLD)
	IONEXSlantDelayStatusValidValue            = uint32(C.SIDEREON_IONEX_SLANT_DELAY_STATUS_VALID)
	IONEXSlantDelayStatusHeldValue             = uint32(C.SIDEREON_IONEX_SLANT_DELAY_STATUS_HELD)
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
	return TecSample{TimeScale: uint32(v.time_scale), EpochJ2000S: float64(v.epoch_j2000_s), LatDeg: float64(v.lat_deg), LonDeg: float64(v.lon_deg), VTECTECU: float64(v.vtec_tecu), RMSPresent: bool(v.has_rms_tecu), RMSTECU: float64(v.rms_tecu)}, nil
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
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ionex_slant_delay_with_policy((*C.SidereonIonex)(p), C.double(lat), C.double(lon), C.double(azimuth), C.double(elevation), C.int64_t(epochJ2000S), C.double(frequencyHz), C.uint32_t(policy), &x))
		})
	})
	if e != nil {
		return IonexSlantDelayEvaluation{}, e
	}
	if x.status != C.SIDEREON_IONEX_SLANT_DELAY_STATUS_VALID && x.status != C.SIDEREON_IONEX_SLANT_DELAY_STATUS_HELD {
		return IonexSlantDelayEvaluation{}, invalidArgument("native IONEX slant-delay status is not defined")
	}
	if x.coverage_error < C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_NONE || x.coverage_error > C.SIDEREON_IONEX_COVERAGE_ERROR_KIND_LONGITUDE_OUT_OF_RANGE {
		return IonexSlantDelayEvaluation{}, invalidArgument("native IONEX coverage error is not defined")
	}
	return IonexSlantDelayEvaluation{DelayM: float64(x.delay_m), Status: uint32(x.status), CoverageError: uint32(x.coverage_error)}, e
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
func buildIONEXFromGrid(s TecGridSamples) (*Ionex, error) {
	if err := validTimeScale(s.TimeScale); err != nil {
		return nil, err
	}
	expected, e := checkedIONEXDimensions(len(s.MapEpochsJ2000S), len(s.LatNodesDeg), len(s.LonNodesDeg))
	if e != nil {
		return nil, e
	}
	if len(s.TECMAPsTECU) != expected {
		return nil, invalidArgument("IONEX VTEC map shape does not match axes")
	}
	if s.RMSPresent && len(s.RMSMAPsTECU) != expected {
		return nil, invalidArgument("IONEX RMS map shape does not match axes")
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
		return nil, e
	}
	lats, e := alloc(s.LatNodesDeg)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, e
	}
	lons, e := alloc(s.LonNodesDeg)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, e
	}
	tec, e := alloc(s.TECMAPsTECU)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, e
	}
	rms, e := alloc(s.RMSMAPsTECU)
	if e != nil {
		for _, p := range mem {
			C.free(p)
		}
		return nil, e
	}
	defer func() {
		for _, p := range mem {
			C.free(p)
		}
	}()
	inMem := C.malloc(C.size_t(unsafe.Sizeof(C.SidereonTecGridSamples{})))
	if inMem == nil {
		return nil, errors.New("sidereon: unable to allocate native IONEX grid descriptor")
	}
	defer C.free(inMem)
	in := (*C.SidereonTecGridSamples)(inMem)
	*in = C.SidereonTecGridSamples{time_scale: C.uint32_t(s.TimeScale), map_epochs_j2000_s: (*C.double)(epochs), map_epoch_count: C.size_t(len(s.MapEpochsJ2000S)), lat_nodes_deg: (*C.double)(lats), lat_node_count: C.size_t(len(s.LatNodesDeg)), lon_nodes_deg: (*C.double)(lons), lon_node_count: C.size_t(len(s.LonNodesDeg)), dlat_deg: C.double(s.DLatDeg), dlon_deg: C.double(s.DLonDeg), shell_height_km: C.double(s.ShellHeightKm), base_radius_km: C.double(s.BaseRadiusKm), exponent: C.int32_t(s.Exponent), tec_maps_tecu: (*C.double)(tec), tec_map_value_count: C.size_t(len(s.TECMAPsTECU)), has_rms_maps: C.bool(s.RMSPresent), rms_maps_tecu: (*C.double)(rms), rms_map_value_count: C.size_t(len(s.RMSMAPsTECU))}
	var out *C.SidereonIonex
	e = callStatus(func() uint32 { return uint32(C.sidereon_ionex_from_tec_grid_samples(in, &out)) })
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_ionex_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native IONEX constructor returned no handle")
	}
	return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, nil
}
func BuildIONEXFromTECGridSamples(s TecGridSamples) (*Ionex, error) { return buildIONEXFromGrid(s) }
func BuildIONEXFromTECSamples(samples []TecSample, shellHeightKm, baseRadiusKm float64, exponent int32) (*Ionex, error) {
	for _, sample := range samples {
		if err := validTimeScale(sample.TimeScale); err != nil {
			return nil, err
		}
	}
	if _, e := checkedNativeAllocationSize(len(samples), unsafe.Sizeof(C.SidereonTecSample{})); e != nil {
		return nil, e
	}
	var mem unsafe.Pointer
	if len(samples) > 0 {
		mem = C.malloc(C.size_t(len(samples)) * C.size_t(unsafe.Sizeof(C.SidereonTecSample{})))
		if mem == nil {
			return nil, errors.New("sidereon: unable to allocate native TEC samples")
		}
		defer C.free(mem)
	}
	raw := unsafe.Slice((*C.SidereonTecSample)(mem), len(samples))
	for j, v := range samples {
		raw[j] = C.SidereonTecSample{time_scale: C.uint32_t(v.TimeScale), epoch_j2000_s: C.double(v.EpochJ2000S), lat_deg: C.double(v.LatDeg), lon_deg: C.double(v.LonDeg), vtec_tecu: C.double(v.VTECTECU), has_rms_tecu: C.bool(v.RMSPresent), rms_tecu: C.double(v.RMSTECU)}
	}
	var p *C.SidereonTecSample
	if len(raw) > 0 {
		p = &raw[0]
	}
	var out *C.SidereonIonex
	e := callStatus(func() uint32 {
		return uint32(C.sidereon_ionex_from_tec_samples(p, C.size_t(len(raw)), C.double(shellHeightKm), C.double(baseRadiusKm), C.int32_t(exponent), &out))
	})
	if e != nil {
		if out != nil {
			withCThread(func() { C.sidereon_ionex_free(out) })
		}
		return nil, e
	}
	if out == nil {
		return nil, errors.New("sidereon: native IONEX constructor returned no handle")
	}
	return &Ionex{handle: newPositioningHandle(unsafe.Pointer(out), releaseIonex)}, nil
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
func (i *Ionex) GridEpochsJ2000S() ([]float64, error) {
	return copyIONEXDoubles(i, "IONEX grid epochs", func(p *C.SidereonIonex, o *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ionex_tec_grid_samples_epochs_j2000_s((*C.SidereonIonex)(p), o, n, w, r)
	})
}
