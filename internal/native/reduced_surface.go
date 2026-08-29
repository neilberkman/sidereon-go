//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type NativeReducedElements struct {
	Model                                                                      uint32
	Epoch                                                                      NativeCalendarEpoch
	AM, E, I, RAAN, RAANRate, RAANRateJ2, ArgLat, MeanMotion, H, K, ArgPerigee float64
}
type NativeECEFSample struct {
	Epoch   NativeCalendarEpoch
	X, Y, Z float64
}
type NativeReducedFitStats struct {
	RMS, Max float64
	NSamples uint64
}
type NativeReducedSampling struct {
	T0, T1  NativeCalendarEpoch
	Cadence float64
}
type NativeReducedSourceFitOptions struct {
	Sampling NativeReducedSampling
	Model    uint32
}
type NativeReducedSourceDriftOptions struct {
	Sampling  NativeReducedSampling
	Threshold float64
}
type NativeReducedDriftEntry struct {
	Epoch NativeCalendarEpoch
	Error float64
}
type NativeReducedDriftSummary struct {
	Max, RMS       float64
	HasCrossing    bool
	ThresholdIndex uint64
}
type NativeReducedPiecewiseInfo struct {
	Model, Scale   uint32
	T0, T1         NativeCalendarEpoch
	SegmentSeconds int64
	NSegments      uint64
}
type NativeReducedPiecewiseSegment struct {
	T0, T1   NativeCalendarEpoch
	Elements NativeReducedElements
	Stats    NativeReducedFitStats
}
type NativeReducedSourceFitStats struct {
	Fit       NativeReducedFitStats
	Requested uint64
}
type NativeReducedPiecewiseSourceStats struct{ Requested, Used uint64 }
type ReducedDriftReport struct {
	_      noCopy
	handle *surfaceHandle
}
type ReducedPiecewise struct {
	_      noCopy
	handle *surfaceHandle
}

func reducedEpoch(v NativeCalendarEpoch) C.SidereonCalendarEpoch {
	return C.SidereonCalendarEpoch{year: C.int32_t(v.Year), month: C.int32_t(v.Month), day: C.int32_t(v.Day), hour: C.int32_t(v.Hour), minute: C.int32_t(v.Minute), second: C.double(v.Second)}
}
func reducedEpochFromC(v C.SidereonCalendarEpoch) NativeCalendarEpoch {
	return NativeCalendarEpoch{int32(v.year), int32(v.month), int32(v.day), int32(v.hour), int32(v.minute), float64(v.second)}
}
func reducedElements(v NativeReducedElements) C.SidereonReducedOrbitElements {
	return C.SidereonReducedOrbitElements{model: C.uint32_t(v.Model), epoch: reducedEpoch(v.Epoch), a_m: C.double(v.AM), e: C.double(v.E), i_rad: C.double(v.I), raan_rad: C.double(v.RAAN), raan_rate_rad_s: C.double(v.RAANRate), raan_rate_j2_rad_s: C.double(v.RAANRateJ2), arg_lat_rad: C.double(v.ArgLat), mean_motion_rad_s: C.double(v.MeanMotion), h: C.double(v.H), k: C.double(v.K), arg_perigee_rad: C.double(v.ArgPerigee)}
}
func reducedElementsFromC(v C.SidereonReducedOrbitElements) NativeReducedElements {
	return NativeReducedElements{uint32(v.model), reducedEpochFromC(v.epoch), float64(v.a_m), float64(v.e), float64(v.i_rad), float64(v.raan_rad), float64(v.raan_rate_rad_s), float64(v.raan_rate_j2_rad_s), float64(v.arg_lat_rad), float64(v.mean_motion_rad_s), float64(v.h), float64(v.k), float64(v.arg_perigee_rad)}
}
func reducedSample(v NativeECEFSample) C.SidereonEcefSample {
	return C.SidereonEcefSample{epoch: reducedEpoch(v.Epoch), x_m: C.double(v.X), y_m: C.double(v.Y), z_m: C.double(v.Z)}
}
func reducedSampling(v NativeReducedSampling) C.SidereonReducedOrbitSourceSampling {
	return C.SidereonReducedOrbitSourceSampling{t0: reducedEpoch(v.T0), t1: reducedEpoch(v.T1), cadence_s: C.double(v.Cadence)}
}
func reducedFitStats(v C.SidereonReducedOrbitFitStats) NativeReducedFitStats {
	return NativeReducedFitStats{float64(v.rms_m), float64(v.max_m), uint64(v.n_samples)}
}
func newReducedDrift(p *C.SidereonReducedOrbitDriftReport) (*ReducedDriftReport, error) {
	h, e := newSurfaceHandle(unsafe.Pointer(p), func(p unsafe.Pointer) {
		C.sidereon_reduced_orbit_drift_report_free((*C.SidereonReducedOrbitDriftReport)(p))
	})
	if e != nil {
		return nil, e
	}
	return &ReducedDriftReport{handle: h}, nil
}
func newReducedPiecewise(p *C.SidereonReducedOrbitPiecewise) (*ReducedPiecewise, error) {
	h, e := newSurfaceHandle(unsafe.Pointer(p), func(p unsafe.Pointer) { C.sidereon_reduced_orbit_piecewise_free((*C.SidereonReducedOrbitPiecewise)(p)) })
	if e != nil {
		return nil, e
	}
	return &ReducedPiecewise{handle: h}, nil
}

func reducedSamples(values []NativeECEFSample) (unsafe.Pointer, C.size_t, error) {
	count, err := cSize(len(values), "reduced-orbit sample count")
	if err != nil {
		return nil, 0, err
	}
	size, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonEcefSample{}))
	if err != nil {
		return nil, 0, err
	}
	if len(values) == 0 {
		return nil, count, nil
	}
	p := C.malloc(C.size_t(size))
	if p == nil {
		return nil, 0, errors.New("sidereon: unable to allocate reduced-orbit samples")
	}
	v := unsafe.Slice((*C.SidereonEcefSample)(p), len(values))
	for i, x := range values {
		v[i] = reducedSample(x)
	}
	return p, count, nil
}
func ReducedOrbitFit(samples []NativeECEFSample, scale, model uint32) (NativeReducedElements, NativeReducedFitStats, error) {
	p, count, e := reducedSamples(samples)
	if e != nil {
		return NativeReducedElements{}, NativeReducedFitStats{}, e
	}
	defer C.free(p)
	var el C.SidereonReducedOrbitElements
	var st C.SidereonReducedOrbitFitStats
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_reduced_orbit_fit((*C.SidereonEcefSample)(p), count, C.uint32_t(scale), C.uint32_t(model), &el, &st))
	})
	return reducedElementsFromC(el), reducedFitStats(st), err
}
func ReducedOrbitPosition(el NativeReducedElements, epoch NativeCalendarEpoch, scale, frame uint32) ([3]float64, error) {
	ce := reducedElements(el)
	ep := reducedEpoch(epoch)
	var out [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_reduced_orbit_position(&ce, &ep, C.uint32_t(scale), C.uint32_t(frame), &out[0], 3))
	})
	var v [3]float64
	for i := range v {
		v[i] = float64(out[i])
	}
	return v, err
}
func ReducedOrbitPositionVelocity(el NativeReducedElements, epoch NativeCalendarEpoch, scale, frame uint32) ([3]float64, [3]float64, error) {
	ce := reducedElements(el)
	ep := reducedEpoch(epoch)
	var p, v [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_reduced_orbit_position_velocity(&ce, &ep, C.uint32_t(scale), C.uint32_t(frame), &p[0], &v[0]))
	})
	var po, vo [3]float64
	for i := range po {
		po[i] = float64(p[i])
		vo[i] = float64(v[i])
	}
	return po, vo, err
}
func ReducedOrbitDrift(el NativeReducedElements, samples []NativeECEFSample, scale uint32, threshold float64) (*ReducedDriftReport, error) {
	ce := reducedElements(el)
	p, count, e := reducedSamples(samples)
	if e != nil {
		return nil, e
	}
	defer C.free(p)
	var out *C.SidereonReducedOrbitDriftReport
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_reduced_orbit_drift(&ce, (*C.SidereonEcefSample)(p), count, C.uint32_t(scale), C.double(threshold), &out))
	})
	if err != nil {
		return nil, err
	}
	return newReducedDrift(out)
}
func (r *ReducedDriftReport) Close() error {
	if r == nil {
		return nil
	}
	return r.handle.close()
}
func (r *ReducedDriftReport) Output() ([]NativeReducedDriftEntry, NativeReducedDriftSummary, uint64, error) {
	var requested C.size_t
	var summary C.SidereonReducedOrbitDriftSummary
	var opErr error
	err := r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_drift_report_requested_samples((*C.SidereonReducedOrbitDriftReport)(p), &requested)))
			if opErr == nil {
				opErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_drift_report_summary((*C.SidereonReducedOrbitDriftReport)(p), &summary)))
			}
		})
		return nil
	})
	if err != nil {
		return nil, NativeReducedDriftSummary{}, 0, err
	}
	if opErr != nil {
		return nil, NativeReducedDriftSummary{}, 0, opErr
	}
	var written, required C.size_t
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_drift_report_entries((*C.SidereonReducedOrbitDriftReport)(p), nil, 0, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, NativeReducedDriftSummary{}, 0, err
	}
	if opErr != nil {
		return nil, NativeReducedDriftSummary{}, 0, opErr
	}
	n, err := sizeTToInt(required, "reduced drift count")
	if err != nil {
		return nil, NativeReducedDriftSummary{}, 0, err
	}
	if _, err = writtenToInt(written, 0, "reduced drift first-call written count"); err != nil {
		return nil, NativeReducedDriftSummary{}, 0, err
	}
	buffer := make([]C.SidereonReducedOrbitDriftEntry, n)
	outputLength, err := cSize(n, "reduced drift output capacity")
	if err != nil {
		return nil, NativeReducedDriftSummary{}, 0, err
	}
	var output *C.SidereonReducedOrbitDriftEntry
	if n > 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_drift_report_entries((*C.SidereonReducedOrbitDriftReport)(p), output, outputLength, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, NativeReducedDriftSummary{}, 0, err
	}
	if opErr != nil {
		return nil, NativeReducedDriftSummary{}, 0, opErr
	}
	actual, err := validateTwoPassCounts("reduced drift entries", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, NativeReducedDriftSummary{}, 0, err
	}
	entries := make([]NativeReducedDriftEntry, actual)
	for i := range entries {
		entries[i] = NativeReducedDriftEntry{reducedEpochFromC(buffer[i].epoch), float64(buffer[i].error_m)}
	}
	return entries, NativeReducedDriftSummary{float64(summary.max_m), float64(summary.rms_m), bool(summary.has_threshold_crossing), uint64(summary.threshold_index)}, uint64(requested), nil
}

func reducedPiecewiseInfo(v C.SidereonReducedOrbitPiecewiseInfo) NativeReducedPiecewiseInfo {
	return NativeReducedPiecewiseInfo{
		Model: uint32(v.model), Scale: uint32(v.scale), T0: reducedEpochFromC(v.t0), T1: reducedEpochFromC(v.t1),
		SegmentSeconds: int64(v.segment_s), NSegments: uint64(v.n_segments),
	}
}

func reducedPiecewiseSegment(v C.SidereonReducedOrbitPiecewiseSegment) NativeReducedPiecewiseSegment {
	return NativeReducedPiecewiseSegment{
		T0: reducedEpochFromC(v.t0), T1: reducedEpochFromC(v.t1),
		Elements: reducedElementsFromC(v.elements), Stats: reducedFitStats(v.stats),
	}
}

func ReducedOrbitFitPiecewise(samples []NativeECEFSample, scale, model uint32, t0, t1 NativeCalendarEpoch, segmentSeconds int64) (*ReducedPiecewise, error) {
	p, count, err := reducedSamples(samples)
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	a, b := reducedEpoch(t0), reducedEpoch(t1)
	var out *C.SidereonReducedOrbitPiecewise
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_reduced_orbit_fit_piecewise((*C.SidereonEcefSample)(p), count, C.uint32_t(scale), C.uint32_t(model), &a, &b, C.int64_t(segmentSeconds), &out))
	})
	if err != nil {
		return nil, err
	}
	return newReducedPiecewise(out)
}

func (p *ReducedPiecewise) Close() error {
	if p == nil {
		return nil
	}
	return p.handle.close()
}

func (p *ReducedPiecewise) Info() (NativeReducedPiecewiseInfo, error) {
	var out C.SidereonReducedOrbitPiecewiseInfo
	err := p.handle.read(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_info((*C.SidereonReducedOrbitPiecewise)(pointer), &out)))
		})
		return callErr
	})
	return reducedPiecewiseInfo(out), err
}

func (p *ReducedPiecewise) Segments() ([]NativeReducedPiecewiseSegment, error) {
	var written, required C.size_t
	var callErr error
	err := p.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_segments((*C.SidereonReducedOrbitPiecewise)(pointer), nil, 0, &written, &required)))
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if callErr != nil {
		return nil, callErr
	}
	n, err := sizeTToInt(required, "reduced piecewise segment count")
	if err != nil {
		return nil, err
	}
	if written != 0 {
		return nil, errors.New("sidereon: reduced piecewise query wrote output")
	}
	buffer := make([]C.SidereonReducedOrbitPiecewiseSegment, n)
	outputLength, err := cSize(n, "reduced piecewise segment output capacity")
	if err != nil {
		return nil, err
	}
	var out *C.SidereonReducedOrbitPiecewiseSegment
	if n > 0 {
		out = &buffer[0]
	}
	written, required = 0, 0
	err = p.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_segments((*C.SidereonReducedOrbitPiecewise)(pointer), out, outputLength, &written, &required)))
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if callErr != nil {
		return nil, callErr
	}
	actual, err := validateTwoPassCounts("reduced piecewise segments", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	values := make([]C.SidereonReducedOrbitPiecewiseSegment, actual)
	copy(values, buffer[:actual])
	if err != nil {
		return nil, err
	}
	result := make([]NativeReducedPiecewiseSegment, len(values))
	for i := range values {
		result[i] = reducedPiecewiseSegment(values[i])
	}
	return result, nil
}

func (p *ReducedPiecewise) SelectSegment(epoch NativeCalendarEpoch) (uint64, NativeReducedPiecewiseSegment, error) {
	ce := reducedEpoch(epoch)
	var index C.size_t
	var segment C.SidereonReducedOrbitPiecewiseSegment
	err := p.handle.read(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_select_segment((*C.SidereonReducedOrbitPiecewise)(pointer), &ce, &index, &segment)))
		})
		return callErr
	})
	return uint64(index), reducedPiecewiseSegment(segment), err
}

func (p *ReducedPiecewise) Position(epoch NativeCalendarEpoch, frame uint32) ([3]float64, error) {
	ce := reducedEpoch(epoch)
	var out [3]C.double
	err := p.handle.read(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_position((*C.SidereonReducedOrbitPiecewise)(pointer), &ce, C.uint32_t(frame), &out[0], 3)))
		})
		return callErr
	})
	var result [3]float64
	for i := range result {
		result[i] = float64(out[i])
	}
	return result, err
}

func (p *ReducedPiecewise) PositionVelocity(epoch NativeCalendarEpoch, frame uint32) ([3]float64, [3]float64, error) {
	ce := reducedEpoch(epoch)
	var position, velocity [3]C.double
	err := p.handle.read(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_position_velocity((*C.SidereonReducedOrbitPiecewise)(pointer), &ce, C.uint32_t(frame), &position[0], &velocity[0])))
		})
		return callErr
	})
	var pOut, vOut [3]float64
	for i := range pOut {
		pOut[i], vOut[i] = float64(position[i]), float64(velocity[i])
	}
	return pOut, vOut, err
}

func (p *ReducedPiecewise) Drift(samples []NativeECEFSample, threshold float64) (*ReducedDriftReport, error) {
	mem, count, err := reducedSamples(samples)
	if err != nil {
		return nil, err
	}
	defer C.free(mem)
	var out *C.SidereonReducedOrbitDriftReport
	err = p.handle.read(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_drift((*C.SidereonReducedOrbitPiecewise)(pointer), (*C.SidereonEcefSample)(mem), count, C.double(threshold), &out)))
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return newReducedDrift(out)
}

func reducedSourceFitOptions(v NativeReducedSourceFitOptions) C.SidereonReducedOrbitSourceFitOptions {
	return C.SidereonReducedOrbitSourceFitOptions{sampling: reducedSampling(v.Sampling), model: C.uint32_t(v.Model)}
}

func reducedSourceDriftOptions(v NativeReducedSourceDriftOptions) C.SidereonReducedOrbitSourceDriftOptions {
	return C.SidereonReducedOrbitSourceDriftOptions{sampling: reducedSampling(v.Sampling), threshold_m: C.double(v.Threshold)}
}

func ReducedOrbitFitSP3Source(sp3 *SP3, satellite string, options NativeReducedSourceFitOptions) (NativeReducedElements, NativeReducedSourceFitStats, error) {
	if sp3 == nil {
		return NativeReducedElements{}, NativeReducedSourceFitStats{}, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "reduced-orbit satellite ID"); err != nil {
		return NativeReducedElements{}, NativeReducedSourceFitStats{}, err
	}
	id := C.CString(satellite)
	if id == nil {
		return NativeReducedElements{}, NativeReducedSourceFitStats{}, errors.New("sidereon: unable to allocate reduced-orbit satellite ID")
	}
	defer C.free(unsafe.Pointer(id))
	o := reducedSourceFitOptions(options)
	var elements C.SidereonReducedOrbitElements
	var stats C.SidereonReducedOrbitSourceFitStats
	err := sp3.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_fit_sp3_source((*C.SidereonSp3)(pointer), id, &o, &elements, &stats)))
		})
		return callErr
	})
	return reducedElementsFromC(elements), NativeReducedSourceFitStats{reducedFitStats(stats.fit), uint64(stats.requested_samples)}, err
}

func ReducedOrbitFitTLESource(tle *TLE, options NativeReducedSourceFitOptions) (NativeReducedElements, NativeReducedSourceFitStats, error) {
	if tle == nil {
		return NativeReducedElements{}, NativeReducedSourceFitStats{}, ErrClosed
	}
	o := reducedSourceFitOptions(options)
	var elements C.SidereonReducedOrbitElements
	var stats C.SidereonReducedOrbitSourceFitStats
	err := tle.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_fit_tle_source((*C.SidereonTle)(pointer), &o, &elements, &stats)))
		})
		return callErr
	})
	return reducedElementsFromC(elements), NativeReducedSourceFitStats{reducedFitStats(stats.fit), uint64(stats.requested_samples)}, err
}

func ReducedOrbitFitPiecewiseSP3Source(sp3 *SP3, satellite string, options NativeReducedSourceFitOptions, segmentSeconds float64) (*ReducedPiecewise, NativeReducedPiecewiseSourceStats, error) {
	if sp3 == nil {
		return nil, NativeReducedPiecewiseSourceStats{}, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "reduced-orbit satellite ID"); err != nil {
		return nil, NativeReducedPiecewiseSourceStats{}, err
	}
	id := C.CString(satellite)
	if id == nil {
		return nil, NativeReducedPiecewiseSourceStats{}, errors.New("sidereon: unable to allocate reduced-orbit satellite ID")
	}
	defer C.free(unsafe.Pointer(id))
	o := reducedSourceFitOptions(options)
	var out *C.SidereonReducedOrbitPiecewise
	var stats C.SidereonReducedOrbitPiecewiseSourceFitStats
	err := sp3.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_fit_piecewise_sp3_source((*C.SidereonSp3)(pointer), id, &o, C.double(segmentSeconds), &out, &stats)))
		})
		return callErr
	})
	if err != nil {
		return nil, NativeReducedPiecewiseSourceStats{}, err
	}
	piecewise, err := newReducedPiecewise(out)
	return piecewise, NativeReducedPiecewiseSourceStats{uint64(stats.requested_samples), uint64(stats.used_samples)}, err
}

func ReducedOrbitFitPiecewiseTLESource(tle *TLE, options NativeReducedSourceFitOptions, segmentSeconds float64) (*ReducedPiecewise, NativeReducedPiecewiseSourceStats, error) {
	if tle == nil {
		return nil, NativeReducedPiecewiseSourceStats{}, ErrClosed
	}
	o := reducedSourceFitOptions(options)
	var out *C.SidereonReducedOrbitPiecewise
	var stats C.SidereonReducedOrbitPiecewiseSourceFitStats
	err := tle.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_fit_piecewise_tle_source((*C.SidereonTle)(pointer), &o, C.double(segmentSeconds), &out, &stats)))
		})
		return callErr
	})
	if err != nil {
		return nil, NativeReducedPiecewiseSourceStats{}, err
	}
	piecewise, err := newReducedPiecewise(out)
	return piecewise, NativeReducedPiecewiseSourceStats{uint64(stats.requested_samples), uint64(stats.used_samples)}, err
}

func (r *ReducedDriftReport) sourceOutput() ([]NativeReducedDriftEntry, NativeReducedDriftSummary, uint64, error) {
	return r.Output()
}

func reducedSourceDriftReport(out **C.SidereonReducedOrbitDriftReport, call func(**C.SidereonReducedOrbitDriftReport) C.enum_SidereonStatus) (*ReducedDriftReport, error) {
	var callErr error
	withCThread(func() { callErr = statusErrorLocked(uint32(call(out))) })
	if callErr != nil {
		return nil, callErr
	}
	return newReducedDrift(*out)
}

func (p *ReducedPiecewise) DriftSP3Source(sp3 *SP3, satellite string, options NativeReducedSourceDriftOptions) (*ReducedDriftReport, error) {
	if sp3 == nil {
		return nil, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "reduced-orbit satellite ID"); err != nil {
		return nil, err
	}
	id := C.CString(satellite)
	if id == nil {
		return nil, errors.New("sidereon: unable to allocate reduced-orbit satellite ID")
	}
	defer C.free(unsafe.Pointer(id))
	o := reducedSourceDriftOptions(options)
	var out *C.SidereonReducedOrbitDriftReport
	var callErr error
	err := p.handle.read(func(pp unsafe.Pointer) error {
		return sp3.handle.with(func(sp unsafe.Pointer) error {
			withCThread(func() {
				callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_drift_sp3_source((*C.SidereonReducedOrbitPiecewise)(pp), (*C.SidereonSp3)(sp), id, &o, &out)))
			})
			return callErr
		})
	})
	if err != nil {
		return nil, err
	}
	return newReducedDrift(out)
}

func (p *ReducedPiecewise) DriftTLESource(tle *TLE, options NativeReducedSourceDriftOptions) (*ReducedDriftReport, error) {
	if tle == nil {
		return nil, ErrClosed
	}
	o := reducedSourceDriftOptions(options)
	var out *C.SidereonReducedOrbitDriftReport
	err := p.handle.read(func(pp unsafe.Pointer) error {
		return tle.handle.with(func(tp unsafe.Pointer) error {
			var callErr error
			withCThread(func() {
				callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_piecewise_drift_tle_source((*C.SidereonReducedOrbitPiecewise)(pp), (*C.SidereonTle)(tp), &o, &out)))
			})
			return callErr
		})
	})
	if err != nil {
		return nil, err
	}
	return newReducedDrift(out)
}

func ReducedOrbitDriftSP3Source(elements NativeReducedElements, sp3 *SP3, satellite string, options NativeReducedSourceDriftOptions) (*ReducedDriftReport, error) {
	if sp3 == nil {
		return nil, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "reduced-orbit satellite ID"); err != nil {
		return nil, err
	}
	id := C.CString(satellite)
	if id == nil {
		return nil, errors.New("sidereon: unable to allocate reduced-orbit satellite ID")
	}
	defer C.free(unsafe.Pointer(id))
	ce := reducedElements(elements)
	o := reducedSourceDriftOptions(options)
	var out *C.SidereonReducedOrbitDriftReport
	err := sp3.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_drift_sp3_source(&ce, (*C.SidereonSp3)(pointer), id, &o, &out)))
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return newReducedDrift(out)
}

func ReducedOrbitDriftTLESource(elements NativeReducedElements, tle *TLE, options NativeReducedSourceDriftOptions) (*ReducedDriftReport, error) {
	if tle == nil {
		return nil, ErrClosed
	}
	ce := reducedElements(elements)
	o := reducedSourceDriftOptions(options)
	var out *C.SidereonReducedOrbitDriftReport
	err := tle.handle.with(func(pointer unsafe.Pointer) error {
		var callErr error
		withCThread(func() {
			callErr = statusErrorLocked(uint32(C.sidereon_reduced_orbit_drift_tle_source(&ce, (*C.SidereonTle)(pointer), &o, &out)))
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return newReducedDrift(out)
}
