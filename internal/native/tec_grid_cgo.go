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

type TecGrid struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func newTecGrid(pointer *C.SidereonTecGrid) (*TecGrid, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h := &TecGrid{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(p unsafe.Pointer) { C.sidereon_tec_grid_free((*C.SidereonTecGrid)(p)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func (g *TecGrid) Close() error {
	if g == nil {
		return nil
	}
	return closeProtocolResource(g, g.resource, &g.cleanup)
}

func cTecGridDoubles(values []float64) (unsafe.Pointer, error) {
	if len(values) == 0 {
		return nil, nil
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.double(0)))
	if err != nil {
		return nil, err
	}
	p := C.malloc(C.size_t(size))
	if p == nil {
		return nil, invalidArgument("unable to allocate TEC grid input values")
	}
	dst := unsafe.Slice((*C.double)(p), len(values))
	for i, v := range values {
		dst[i] = C.double(v)
	}
	return p, nil
}

func (e *NativeTecGridError) nativeResultError() error { return e }

func readTecGridResultErrorLocked(result *C.SidereonTecGridResult, outcome C.SidereonTecGridOutcome) (*NativeTecGridError, error) {
	if bool(outcome.is_ok) {
		return nil, nil
	}
	data := outcome.error
	errorValue := &NativeTecGridError{Kind: uint32(data.kind), Status: uint32(outcome.status), Axis: uint32(data.axis), HasGap: bool(data.has_gap), HasAxisValue: bool(data.has_axis_value), AxisValue: float64(data.axis_value), HasValueIndex: bool(data.has_value_index), ValueIndex: uint64(data.value_index)}
	for i := 0; i < 4; i++ {
		errorValue.Gap.Earlier[i] = bool(data.gap.earlier.missing[i])
		errorValue.Gap.Later[i] = bool(data.gap.later.missing[i])
	}
	errorValue.Gap.HasGap = bool(data.gap.has_gap)
	var err error
	if errorValue.ValueCount, err = sizeTToInt(data.value_count, "TEC grid input value count"); err != nil {
		return nil, err
	}
	if errorValue.ExpectedValueCount, err = sizeTToInt(data.expected_value_count, "TEC grid expected value count"); err != nil {
		return nil, err
	}
	if errorValue.Message, err = copyTecGridResultTextLocked(result, "message", "TEC grid failure message"); err != nil {
		return nil, err
	}
	// Field and reason are separately owned by the result. Their getters use the same two-pass contract.
	if errorValue.Field, err = copyTecGridResultTextLocked(result, "field", "TEC grid failure field"); err != nil {
		return nil, err
	}
	errorValue.HasField = errorValue.Field != ""
	if errorValue.Reason, err = copyTecGridResultTextLocked(result, "reason", "TEC grid failure reason"); err != nil {
		return nil, err
	}
	errorValue.HasReason = errorValue.Reason != ""
	return errorValue, nil
}

func copyTecGridResultTextLocked(result *C.SidereonTecGridResult, part, label string) (string, error) {
	b, err := copyNativeBytesLocked(label, func(dst *C.uint8_t, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		switch part {
		case "field":
			return C.sidereon_tec_grid_result_get_field(result, dst, n, w, r)
		case "reason":
			return C.sidereon_tec_grid_result_get_reason(result, dst, n, w, r)
		default:
			return C.sidereon_tec_grid_result_get_message(result, dst, n, w, r)
		}
	})
	return string(b), err
}

func NewTecGrid(input TecGridInput) (*TecGrid, *NativeTecGridError, error) {
	if len(input.Presence) != 0 && len(input.Presence) != len(input.ValuesTECU) {
		return nil, nil, invalidArgument("TEC grid presence count must be zero or match the values")
	}
	ptrs := make([]unsafe.Pointer, 0, 4)
	free := func() {
		for _, p := range ptrs {
			if p != nil {
				C.free(p)
			}
		}
	}
	defer free()
	epochs, err := cTecGridDoubles(input.EpochsUnixNanos)
	if err != nil {
		return nil, nil, err
	}
	ptrs = append(ptrs, epochs)
	lats, err := cTecGridDoubles(input.LatitudesDeg)
	if err != nil {
		return nil, nil, err
	}
	ptrs = append(ptrs, lats)
	lons, err := cTecGridDoubles(input.LongitudesDeg)
	if err != nil {
		return nil, nil, err
	}
	ptrs = append(ptrs, lons)
	values, err := cTecGridDoubles(input.ValuesTECU)
	if err != nil {
		return nil, nil, err
	}
	ptrs = append(ptrs, values)
	var presence unsafe.Pointer
	if len(input.Presence) > 0 {
		size, err := checkedNativeAllocationSize(len(input.Presence), unsafe.Sizeof(C.bool(false)))
		if err != nil {
			return nil, nil, err
		}
		presence = C.malloc(C.size_t(size))
		if presence == nil {
			return nil, nil, invalidArgument("unable to allocate TEC grid presence flags")
		}
		ptrs = append(ptrs, presence)
		dst := unsafe.Slice((*C.bool)(presence), len(input.Presence))
		for i, v := range input.Presence {
			dst[i] = C.bool(v)
		}
	}
	var grid *C.SidereonTecGrid
	var result *C.SidereonTecGridResult
	err = withCThreadError(func() error {
		status := C.sidereon_tec_grid_new_result((*C.double)(epochs), C.size_t(len(input.EpochsUnixNanos)), (*C.double)(lats), C.size_t(len(input.LatitudesDeg)), (*C.double)(lons), C.size_t(len(input.LongitudesDeg)), (*C.double)(values), (*C.bool)(presence), C.size_t(len(input.ValuesTECU)), &grid, &result)
		if result != nil {
			defer C.sidereon_tec_grid_result_free(result)
		}
		if e := callStatus(func() uint32 { return uint32(status) }); e != nil {
			return e
		}
		if result == nil {
			return missingNativeHandle("TEC grid construction result")
		}
		var outcome C.SidereonTecGridOutcome
		if e := callStatus(func() uint32 { return uint32(C.sidereon_tec_grid_result_get_outcome(result, &outcome)) }); e != nil {
			return e
		}
		failure, e := readTecGridResultErrorLocked(result, outcome)
		if e != nil {
			return e
		}
		if failure != nil {
			return failure
		}
		return nil
	})
	if err != nil {
		if grid != nil {
			C.sidereon_tec_grid_free(grid)
		}
		if failure, ok := err.(*NativeTecGridError); ok {
			return nil, failure, nil
		}
		return nil, nil, err
	}
	h, err := newTecGrid(grid)
	if err != nil {
		if grid != nil {
			C.sidereon_tec_grid_free(grid)
		}
		return nil, nil, err
	}
	return h, nil, nil
}

func (g *TecGrid) Dimensions() (NativeTecGridInfo, error) {
	if g == nil || g.resource == nil {
		return NativeTecGridInfo{}, ErrClosed
	}
	var epochs, lats, lons, values C.size_t
	err := g.resource.with(func(owner unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_tec_grid_dimensions((*C.SidereonTecGrid)(owner), &epochs, &lats, &lons, &values))
		})
	})
	runtime.KeepAlive(g)
	if err != nil {
		return NativeTecGridInfo{}, err
	}
	e, err := sizeTToInt(epochs, "TEC grid epoch count")
	if err != nil {
		return NativeTecGridInfo{}, err
	}
	la, err := sizeTToInt(lats, "TEC grid latitude count")
	if err != nil {
		return NativeTecGridInfo{}, err
	}
	lo, err := sizeTToInt(lons, "TEC grid longitude count")
	if err != nil {
		return NativeTecGridInfo{}, err
	}
	v, err := sizeTToInt(values, "TEC grid value count")
	if err != nil {
		return NativeTecGridInfo{}, err
	}
	return NativeTecGridInfo{EpochCount: e, LatitudeCount: la, LongitudeCount: lo, ValueCount: v}, nil
}

func (g *TecGrid) copyDoubles(which string, count int) ([]float64, error) {
	if g == nil || g.resource == nil {
		return nil, ErrClosed
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(float64(0))); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	out := make([]float64, count)
	err := g.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			var invoke func(*C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus
			switch which {
			case "epoch":
				invoke = func(p *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_tec_grid_epochs_ns((*C.SidereonTecGrid)(owner), p, n, w, r)
				}
			case "latitude":
				invoke = func(p *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_tec_grid_latitudes_deg((*C.SidereonTecGrid)(owner), p, n, w, r)
				}
			case "longitude":
				invoke = func(p *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_tec_grid_longitudes_deg((*C.SidereonTecGrid)(owner), p, n, w, r)
				}
			default:
				invoke = func(p *C.double, n C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
					return C.sidereon_tec_grid_values_tecu((*C.SidereonTecGrid)(owner), p, n, w, r)
				}
			}
			if err := callStatus(func() uint32 { return uint32(invoke(nil, 0, &written, &required)) }); err != nil {
				return err
			}
			q, err := validateNativeQuery("TEC grid "+which, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if q != count {
				return invalidArgument("TEC grid dimension and " + which + " axis count disagree")
			}
			buf := make([]C.double, count)
			if count > 0 {
				if err := callStatus(func() uint32 { return uint32(invoke(&buf[0], C.size_t(count), &written, &required)) }); err != nil {
					return err
				}
			}
			if _, err := validateTwoPassCounts("TEC grid "+which, count, count, uint64(written), uint64(required)); err != nil {
				return err
			}
			for i, v := range buf {
				out[i] = float64(v)
			}
			return nil
		})
	})
	runtime.KeepAlive(g)
	return out, err
}

func (g *TecGrid) Epochs() ([]float64, error) {
	d, e := g.Dimensions()
	if e != nil {
		return nil, e
	}
	return g.copyDoubles("epoch", d.EpochCount)
}
func (g *TecGrid) Latitudes() ([]float64, error) {
	d, e := g.Dimensions()
	if e != nil {
		return nil, e
	}
	return g.copyDoubles("latitude", d.LatitudeCount)
}
func (g *TecGrid) Longitudes() ([]float64, error) {
	d, e := g.Dimensions()
	if e != nil {
		return nil, e
	}
	return g.copyDoubles("longitude", d.LongitudeCount)
}
func (g *TecGrid) Values() ([]float64, error) {
	d, e := g.Dimensions()
	if e != nil {
		return nil, e
	}
	return g.copyDoubles("value", d.ValueCount)
}

func (g *TecGrid) Presence() ([]bool, error) {
	if g == nil || g.resource == nil {
		return nil, ErrClosed
	}
	d, err := g.Dimensions()
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(d.ValueCount, unsafe.Sizeof(bool(false))); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(d.ValueCount, unsafe.Sizeof(C.bool(false))); err != nil {
		return nil, err
	}
	out := make([]bool, d.ValueCount)
	err = g.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			invoke := func(dst *C.bool, n C.size_t) C.enum_SidereonStatus {
				return C.sidereon_tec_grid_value_presence((*C.SidereonTecGrid)(owner), dst, n, &written, &required)
			}
			if err := callStatus(func() uint32 { return uint32(invoke(nil, 0)) }); err != nil {
				return err
			}
			q, err := validateNativeQuery("TEC grid presence", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if q != len(out) {
				return invalidArgument("TEC grid dimension and presence count disagree")
			}
			buf := make([]C.bool, len(out))
			if len(buf) > 0 {
				if err := callStatus(func() uint32 { return uint32(invoke(&buf[0], C.size_t(len(buf)))) }); err != nil {
					return err
				}
			}
			if _, err := validateTwoPassCounts("TEC grid presence", len(buf), len(out), uint64(written), uint64(required)); err != nil {
				return err
			}
			for i, v := range buf {
				out[i] = bool(v)
			}
			return nil
		})
	})
	runtime.KeepAlive(g)
	return out, err
}

func (g *TecGrid) VTECAt(unixNanos int64, lat, lon float64, policy uint32) (NativeTecGridVTEC, error) {
	if g == nil || g.resource == nil {
		return NativeTecGridVTEC{}, ErrClosed
	}
	var out NativeTecGridVTEC
	err := g.resource.with(func(owner unsafe.Pointer) error {
		return withCThreadError(func() error {
			var result *C.SidereonTecGridResult
			status := C.sidereon_tec_grid_vtec_at_pierce_point_result((*C.SidereonTecGrid)(owner), C.int64_t(unixNanos), C.double(lat), C.double(lon), C.uint32_t(policy), &result)
			if result != nil {
				defer C.sidereon_tec_grid_result_free(result)
			}
			if e := callStatus(func() uint32 { return uint32(status) }); e != nil {
				return e
			}
			if result == nil {
				return missingNativeHandle("TEC grid query result")
			}
			var outcome C.SidereonTecGridOutcome
			if e := callStatus(func() uint32 { return uint32(C.sidereon_tec_grid_result_get_outcome(result, &outcome)) }); e != nil {
				return e
			}
			failure, e := readTecGridResultErrorLocked(result, outcome)
			if e != nil {
				return e
			}
			if failure != nil {
				return failure
			}
			out.HasVTEC = bool(outcome.has_vtec)
			if out.HasVTEC {
				out.ValueTECU = float64(outcome.vtec_tecu)
			}
			for i := 0; i < 4; i++ {
				out.Degraded.Earlier[i] = bool(outcome.degraded.earlier.missing[i])
				out.Degraded.Later[i] = bool(outcome.degraded.later.missing[i])
			}
			out.Degraded.HasGap = bool(outcome.degraded.has_gap)
			return nil
		})
	})
	runtime.KeepAlive(g)
	return out, err
}
