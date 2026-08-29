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

type NativeARAIMRow struct {
	SatelliteID  string
	LineOfSight  LineOfSight
	System       uint32
	ElevationRad float64
}
type NativeARAIMGeometry struct {
	Rows         []NativeARAIMRow
	Receiver     Geodetic
	ClockSystems []uint32
}
type NativeARAIMSatelliteModel struct {
	SigmaURA, SigmaURE float64
	HasEffectiveInt    bool
	EffectiveInt       float64
	HasEffectiveAcc    bool
	EffectiveAcc       float64
	BNom, PSat         float64
}
type NativeARAIMConstellation struct {
	System  uint32
	PConst  float64
	Default NativeARAIMSatelliteModel
}
type NativeARAIMSatellite struct {
	SatelliteID string
	Model       NativeARAIMSatelliteModel
}
type NativeARAIMISM struct {
	Constellations []NativeARAIMConstellation
	Satellites     []NativeARAIMSatellite
}
type NativeARAIMAllocation struct {
	PHMITotal, PHMIVert, PHMIHor, PFAVert, PFAHor, PThresholdUnmonitored, PEMT float64
	MaxFaultOrder                                                              uint64
}
type NativeARAIMFaultMode struct {
	ExcludedCount                      uint64
	HasExcludedConstellation           bool
	ExcludedConstellation              uint32
	Prior                              float64
	SigmaIntENU, BiasENU, ThresholdENU [3]float64
	Monitorable                        bool
}
type NativeARAIMSummary struct {
	HPLM, VPLM, SigmaAccHM, SigmaAccVM, EMTM, PUnmonitored float64
	Available, Availability                                bool
	FaultModeCount                                         uint64
}
type ARAIMResult struct {
	_      noCopy
	handle *surfaceHandle
}

func araimModel(v NativeARAIMSatelliteModel) C.SidereonAraimSatelliteIsmModel {
	return C.SidereonAraimSatelliteIsmModel{sigma_ura_m: C.double(v.SigmaURA), sigma_ure_m: C.double(v.SigmaURE), has_effective_sigma_int_m: C.bool(v.HasEffectiveInt), effective_sigma_int_m: C.double(v.EffectiveInt), has_effective_sigma_acc_m: C.bool(v.HasEffectiveAcc), effective_sigma_acc_m: C.double(v.EffectiveAcc), b_nom_m: C.double(v.BNom), p_sat: C.double(v.PSat)}
}
func buildARAIM(v NativeARAIMGeometry, ism NativeARAIMISM, allocation NativeARAIMAllocation) (C.SidereonAraimGeometry, C.SidereonAraimIsm, C.SidereonAraimIntegrityAllocation, []unsafe.Pointer, []*C.char, error) {
	var geometry C.SidereonAraimGeometry
	var cism C.SidereonAraimIsm
	var ca C.SidereonAraimIntegrityAllocation
	var allocated []unsafe.Pointer
	var ids []*C.char
	completed := false
	defer func() {
		if !completed {
			for _, pointer := range allocated {
				C.free(pointer)
			}
			freeCStrings(ids)
		}
	}()
	rowCount, err := cSize(len(v.Rows), "ARAIM row count")
	if err != nil {
		return geometry, cism, ca, nil, nil, err
	}
	rowsSize, err := checkedNativeAllocationSize(len(v.Rows), unsafe.Sizeof(C.SidereonAraimRow{}))
	if err != nil {
		return geometry, cism, ca, nil, nil, err
	}
	var rows unsafe.Pointer
	if len(v.Rows) > 0 {
		rows = C.malloc(C.size_t(rowsSize))
		if rows == nil {
			return geometry, cism, ca, nil, nil, errors.New("sidereon: unable to allocate ARAIM rows")
		}
		allocated = append(allocated, rows)
	}
	crows := unsafe.Slice((*C.SidereonAraimRow)(rows), len(v.Rows))
	for i, r := range v.Rows {
		if err := rejectEmbeddedNUL(r.SatelliteID, "ARAIM satellite ID"); err != nil {
			return geometry, cism, ca, allocated, ids, err
		}
		id := C.CString(r.SatelliteID)
		if id == nil {
			return geometry, cism, ca, allocated, ids, errors.New("sidereon: unable to allocate ARAIM satellite ID")
		}
		ids = append(ids, id)
		crows[i].sat_id = id
		crows[i].line_of_sight = C.SidereonLineOfSight{e_x: C.double(r.LineOfSight.EX), e_y: C.double(r.LineOfSight.EY), e_z: C.double(r.LineOfSight.EZ)}
		crows[i].system = C.uint32_t(r.System)
		crows[i].elevation_rad = C.double(r.ElevationRad)
	}
	if len(crows) > 0 {
		geometry.rows = &crows[0]
	}
	geometry.row_count = rowCount
	geometry.receiver = cGeodetic(v.Receiver)
	clockCount, err := cSize(len(v.ClockSystems), "ARAIM clock-system count")
	if err != nil {
		return geometry, cism, ca, allocated, ids, err
	}
	if len(v.ClockSystems) > 0 {
		clockSize, err := checkedNativeAllocationSize(len(v.ClockSystems), unsafe.Sizeof(C.uint32_t(0)))
		if err != nil {
			return geometry, cism, ca, allocated, ids, err
		}
		mem := C.malloc(C.size_t(clockSize))
		if mem == nil {
			return geometry, cism, ca, allocated, ids, errors.New("sidereon: unable to allocate ARAIM clock systems")
		}
		allocated = append(allocated, mem)
		systems := unsafe.Slice((*C.uint32_t)(mem), len(v.ClockSystems))
		for i, x := range v.ClockSystems {
			systems[i] = C.uint32_t(x)
		}
		geometry.clock_systems = &systems[0]
	}
	geometry.clock_system_count = clockCount
	constellationCount, err := cSize(len(ism.Constellations), "ARAIM constellation count")
	if err != nil {
		return geometry, cism, ca, allocated, ids, err
	}
	csize, err := checkedNativeAllocationSize(len(ism.Constellations), unsafe.Sizeof(C.SidereonAraimConstellationIsm{}))
	if err != nil {
		return geometry, cism, ca, allocated, ids, err
	}
	var cmem unsafe.Pointer
	if len(ism.Constellations) > 0 {
		cmem = C.malloc(C.size_t(csize))
		if cmem == nil {
			return geometry, cism, ca, allocated, ids, errors.New("sidereon: unable to allocate ARAIM constellations")
		}
		allocated = append(allocated, cmem)
		values := unsafe.Slice((*C.SidereonAraimConstellationIsm)(cmem), len(ism.Constellations))
		for i, x := range ism.Constellations {
			values[i].system = C.uint32_t(x.System)
			values[i].p_const = C.double(x.PConst)
			values[i].default_sat = araimModel(x.Default)
		}
		cism.constellations = &values[0]
	}
	cism.constellation_count = constellationCount
	satelliteCount, err := cSize(len(ism.Satellites), "ARAIM satellite count")
	if err != nil {
		return geometry, cism, ca, allocated, ids, err
	}
	size, err := checkedNativeAllocationSize(len(ism.Satellites), unsafe.Sizeof(C.SidereonAraimSatelliteIsm{}))
	if err != nil {
		return geometry, cism, ca, allocated, ids, err
	}
	var smem unsafe.Pointer
	if len(ism.Satellites) > 0 {
		smem = C.malloc(C.size_t(size))
		if smem == nil {
			return geometry, cism, ca, allocated, ids, errors.New("sidereon: unable to allocate ARAIM satellites")
		}
		allocated = append(allocated, smem)
		values := unsafe.Slice((*C.SidereonAraimSatelliteIsm)(smem), len(ism.Satellites))
		for i, x := range ism.Satellites {
			if err := rejectEmbeddedNUL(x.SatelliteID, "ARAIM satellite ID"); err != nil {
				return geometry, cism, ca, allocated, ids, err
			}
			id := C.CString(x.SatelliteID)
			if id == nil {
				return geometry, cism, ca, allocated, ids, errors.New("sidereon: unable to allocate ARAIM satellite ID")
			}
			ids = append(ids, id)
			values[i].sat_id = id
			values[i].sigma_ura_m = C.double(x.Model.SigmaURA)
			values[i].sigma_ure_m = C.double(x.Model.SigmaURE)
			values[i].has_effective_sigma_int_m = C.bool(x.Model.HasEffectiveInt)
			values[i].effective_sigma_int_m = C.double(x.Model.EffectiveInt)
			values[i].has_effective_sigma_acc_m = C.bool(x.Model.HasEffectiveAcc)
			values[i].effective_sigma_acc_m = C.double(x.Model.EffectiveAcc)
			values[i].b_nom_m = C.double(x.Model.BNom)
			values[i].p_sat = C.double(x.Model.PSat)
		}
		cism.satellites = &values[0]
	}
	cism.satellite_count = satelliteCount
	maxFaultOrder, err := cSize64(allocation.MaxFaultOrder, "ARAIM maximum fault order")
	if err != nil {
		return geometry, cism, ca, allocated, ids, err
	}
	ca = C.SidereonAraimIntegrityAllocation{phmi_total: C.double(allocation.PHMITotal), phmi_vert: C.double(allocation.PHMIVert), phmi_hor: C.double(allocation.PHMIHor), pfa_vert: C.double(allocation.PFAVert), pfa_hor: C.double(allocation.PFAHor), p_threshold_unmonitored: C.double(allocation.PThresholdUnmonitored), p_emt: C.double(allocation.PEMT), max_fault_order: maxFaultOrder}
	completed = true
	return geometry, cism, ca, allocated, ids, nil
}

func freeARAIMAllocations(values []unsafe.Pointer) {
	for _, pointer := range values {
		if pointer != nil {
			C.free(pointer)
		}
	}
}

func RunARAIM(geometry NativeARAIMGeometry, ism NativeARAIMISM, allocation NativeARAIMAllocation) (*ARAIMResult, error) {
	g, i, a, allocations, ids, err := buildARAIM(geometry, ism, allocation)
	if err != nil {
		return nil, err
	}
	defer freeARAIMAllocations(allocations)
	defer freeCStrings(ids)
	var out *C.SidereonAraimResult
	var opErr error
	withCThread(func() { opErr = statusErrorLocked(uint32(C.sidereon_araim(&g, &i, &a, &out))) })
	if opErr != nil {
		return nil, opErr
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(p unsafe.Pointer) { C.sidereon_araim_result_free((*C.SidereonAraimResult)(p)) })
	if e != nil {
		return nil, e
	}
	return &ARAIMResult{handle: h}, nil
}
func ARAIMAllocationLPV200() (NativeARAIMAllocation, error) {
	var out C.SidereonAraimIntegrityAllocation
	err := callStatus(func() uint32 { return uint32(C.sidereon_araim_allocation_lpv_200(&out)) })
	return NativeARAIMAllocation{float64(out.phmi_total), float64(out.phmi_vert), float64(out.phmi_hor), float64(out.pfa_vert), float64(out.pfa_hor), float64(out.p_threshold_unmonitored), float64(out.p_emt), uint64(out.max_fault_order)}, err
}
func (s *ARAIMResult) Close() error {
	if s == nil {
		return nil
	}
	return s.handle.close()
}
func (s *ARAIMResult) Summary() (NativeARAIMSummary, error) {
	var out C.SidereonAraimSummary
	var opErr error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_araim_result_summary((*C.SidereonAraimResult)(p), &out)))
		})
		return nil
	})
	if err != nil {
		return NativeARAIMSummary{}, err
	}
	if opErr != nil {
		return NativeARAIMSummary{}, opErr
	}
	return NativeARAIMSummary{float64(out.hpl_m), float64(out.vpl_m), float64(out.sigma_acc_h_m), float64(out.sigma_acc_v_m), float64(out.emt_m), float64(out.p_unmonitored), bool(out.available), bool(out.availability), uint64(out.fault_mode_count)}, nil
}
func (s *ARAIMResult) FaultModes() ([]NativeARAIMFaultMode, error) {
	var written, required C.size_t
	var opErr error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_araim_result_fault_modes((*C.SidereonAraimResult)(p), nil, 0, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	n, err := sizeTToInt(required, "ARAIM fault-mode count")
	if err != nil {
		return nil, err
	}
	if _, err = writtenToInt(written, 0, "ARAIM fault-mode first-call written count"); err != nil {
		return nil, err
	}
	if _, err = checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonAraimFaultMode{})); err != nil {
		return nil, err
	}
	buffer := make([]C.SidereonAraimFaultMode, n)
	outputLength, err := cSize(n, "ARAIM fault-mode output capacity")
	if err != nil {
		return nil, err
	}
	var output *C.SidereonAraimFaultMode
	if n > 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	err = s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_araim_result_fault_modes((*C.SidereonAraimResult)(p), output, outputLength, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	actual, err := validateTwoPassCounts("ARAIM fault modes", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	out := make([]NativeARAIMFaultMode, actual)
	for j := range out {
		v := buffer[j]
		value := NativeARAIMFaultMode{ExcludedCount: uint64(v.excluded_count), HasExcludedConstellation: bool(v.has_excluded_constellation), ExcludedConstellation: uint32(v.excluded_constellation), Prior: float64(v.prior), Monitorable: bool(v.monitorable)}
		for k := 0; k < 3; k++ {
			value.SigmaIntENU[k] = float64(v.sigma_int_enu_m[k])
			value.BiasENU[k] = float64(v.bias_enu_m[k])
			value.ThresholdENU[k] = float64(v.threshold_enu_m[k])
		}
		out[j] = value
	}
	return out, nil
}
func (s *ARAIMResult) Excluded(mode int) ([]string, error) {
	if mode < 0 {
		return nil, errors.New("sidereon: negative ARAIM fault-mode index")
	}
	return copySurfaceTokens(s.handle, "ARAIM excluded satellites", func(p unsafe.Pointer, o *C.SidereonSatelliteToken, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_araim_result_fault_mode_excluded_sats((*C.SidereonAraimResult)(p), C.size_t(mode), o, l, w, r)
	})
}

type NativeReliabilityOptions struct {
	Alpha, Beta            float64
	HasLambda0             bool
	Lambda0, MinRedundancy float64
}
type NativeReliabilityRow struct {
	ID     string
	Design []float64
	Sigma  float64
}
type NativeReliabilityObservation struct {
	ID             string
	Redundancy     float64
	HasMDB         bool
	MDB            float64
	HasExternal    bool
	External       [3]float64
	HasBiasToNoise bool
	BiasToNoise    float64
	Uncheckable    bool
}
type NativeReliabilitySummary struct {
	NObs, NParams, DOF     uint64
	SumRedundancy, Lambda0 float64
	HasMaxMDB              bool
	MaxMDBID               string
	MaxMDB                 float64
	MinRedundancyID        string
	MinRedundancy          float64
	NUncheckable           uint64
}
type ReliabilityReport struct {
	_      noCopy
	handle *surfaceHandle
}
type RangeFDEResult struct {
	_      noCopy
	handle *surfaceHandle
}
type NativeRangeFDEOptions struct {
	PFA                          float64
	MaxExclusions, MinRedundancy uint64
}
type NativeRangeFDERow struct {
	ID       string
	Residual float64
	Design   []float64
	Weight   float64
}
type NativeRangeFDEGlobalTest struct {
	WeightedSumSquares float64
	DOF                int64
	HasThreshold       bool
	Threshold          float64
	Testable           bool
	FaultDetected      bool
}
type NativeRangeFDEDiagnostic struct {
	ID                                  string
	Excluded                            bool
	PostFitResidual, NormalizedResidual float64
}
type NativeRangeFDEOutput struct {
	StateDimension         uint64
	Correction, Covariance []float64
	Global                 NativeRangeFDEGlobalTest
	Iterations             uint64
	Excluded               []string
	Diagnostics            []NativeRangeFDEDiagnostic
}

func copyNativeRtkIDs(call func(*C.SidereonRtkId, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus, label string) ([]string, error) {
	var written, required C.size_t
	if err := statusErrorLocked(uint32(call(nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	n, err := sizeTToInt(required, label+" required count")
	if err != nil {
		return nil, err
	}
	if _, err = writtenToInt(written, 0, label+" first-call written count"); err != nil {
		return nil, err
	}
	if _, err = checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonRtkId{})); err != nil {
		return nil, err
	}
	buffer := make([]C.SidereonRtkId, n)
	outputLength, err := cSize(n, label+" output capacity")
	if err != nil {
		return nil, err
	}
	var output *C.SidereonRtkId
	if n > 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(output, outputLength, &written, &required))); err != nil {
		return nil, err
	}
	actual, err := validateTwoPassCounts(label, n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	result := make([]string, actual)
	for i := range result {
		result[i] = rtkIDFromC(buffer[i]).Value
	}
	return result, nil
}

func ReliabilityOptionsDefault() (NativeReliabilityOptions, error) {
	var o C.SidereonReliabilityOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_reliability_options_init(&o)) })
	return NativeReliabilityOptions{float64(o.alpha), float64(o.beta), bool(o.has_lambda0_override), float64(o.lambda0_override), float64(o.min_redundancy)}, err
}
func ReliabilityDesign(rows []NativeReliabilityRow, options NativeReliabilityOptions) (*ReliabilityReport, error) {
	rowCount, err := cSize(len(rows), "reliability row count")
	if err != nil {
		return nil, err
	}
	size, err := checkedNativeAllocationSize(len(rows), unsafe.Sizeof(C.SidereonRangeReliabilityRow{}))
	if err != nil {
		return nil, err
	}
	var mem unsafe.Pointer
	if len(rows) > 0 {
		mem = C.malloc(C.size_t(size))
		if mem == nil {
			return nil, errors.New("sidereon: unable to allocate reliability rows")
		}
		defer C.free(mem)
	}
	cr := unsafe.Slice((*C.SidereonRangeReliabilityRow)(mem), len(rows))
	ids := make([]*C.char, len(rows))
	defer freeCStrings(ids)
	designPointers := make([]unsafe.Pointer, len(rows))
	defer func() {
		for _, p := range designPointers {
			if p != nil {
				C.free(p)
			}
		}
	}()
	for j, r := range rows {
		if err := rejectEmbeddedNUL(r.ID, "reliability observation ID"); err != nil {
			return nil, err
		}
		ids[j] = C.CString(r.ID)
		if ids[j] == nil {
			return nil, errors.New("sidereon: unable to allocate reliability observation ID")
		}
		p, designLength, err := cFloats(r.Design, "reliability design row")
		if err != nil {
			return nil, err
		}
		designPointers[j] = p
		cr[j].id = ids[j]
		cr[j].design_row = (*C.double)(p)
		cr[j].design_dim = designLength
		cr[j].sigma_m = C.double(r.Sigma)
	}
	co := C.SidereonReliabilityOptions{alpha: C.double(options.Alpha), beta: C.double(options.Beta), has_lambda0_override: C.bool(options.HasLambda0), lambda0_override: C.double(options.Lambda0), min_redundancy: C.double(options.MinRedundancy)}
	var out *C.SidereonReliabilityReport
	var opErr error
	withCThread(func() {
		opErr = statusErrorLocked(uint32(C.sidereon_reliability_design((*C.SidereonRangeReliabilityRow)(mem), rowCount, &co, &out)))
	})
	if opErr != nil {
		return nil, opErr
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(p unsafe.Pointer) { C.sidereon_reliability_report_free((*C.SidereonReliabilityReport)(p)) })
	if e != nil {
		return nil, e
	}
	return &ReliabilityReport{handle: h}, nil
}

func ReliabilityARAIM(geometry NativeARAIMGeometry, ism NativeARAIMISM, options NativeReliabilityOptions) (*ReliabilityReport, error) {
	g, i, _, allocations, ids, err := buildARAIM(geometry, ism, NativeARAIMAllocation{})
	if err != nil {
		return nil, err
	}
	defer freeARAIMAllocations(allocations)
	defer freeCStrings(ids)
	co := C.SidereonReliabilityOptions{alpha: C.double(options.Alpha), beta: C.double(options.Beta), has_lambda0_override: C.bool(options.HasLambda0), lambda0_override: C.double(options.Lambda0), min_redundancy: C.double(options.MinRedundancy)}
	var out *C.SidereonReliabilityReport
	var opErr error
	withCThread(func() { opErr = statusErrorLocked(uint32(C.sidereon_reliability_araim(&g, &i, &co, &out))) })
	if opErr != nil {
		return nil, opErr
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(p unsafe.Pointer) { C.sidereon_reliability_report_free((*C.SidereonReliabilityReport)(p)) })
	if e != nil {
		return nil, e
	}
	return &ReliabilityReport{handle: h}, nil
}
func (r *ReliabilityReport) Close() error {
	if r == nil {
		return nil
	}
	return r.handle.close()
}
func (r *ReliabilityReport) Summary() (NativeReliabilitySummary, error) {
	var out C.SidereonReliabilitySummary
	var opErr error
	err := r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_reliability_report_summary((*C.SidereonReliabilityReport)(p), &out)))
		})
		return nil
	})
	if err != nil {
		return NativeReliabilitySummary{}, err
	}
	if opErr != nil {
		return NativeReliabilitySummary{}, opErr
	}
	return NativeReliabilitySummary{uint64(out.n_obs), uint64(out.n_params), uint64(out.dof), float64(out.sum_redundancy), float64(out.lambda0), bool(out.has_max_mdb_m), rtkIDFromC(out.max_mdb_id).Value, float64(out.max_mdb_m), rtkIDFromC(out.min_redundancy_id).Value, float64(out.min_redundancy), uint64(out.n_uncheckable)}, nil
}
func (r *ReliabilityReport) Observations() ([]NativeReliabilityObservation, error) {
	var written, required C.size_t
	var opErr error
	err := r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_reliability_report_observations((*C.SidereonReliabilityReport)(p), nil, 0, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	n, err := sizeTToInt(required, "reliability observation count")
	if err != nil {
		return nil, err
	}
	if _, err = writtenToInt(written, 0, "reliability observation first-call written count"); err != nil {
		return nil, err
	}
	if _, err = checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonObservationReliability{})); err != nil {
		return nil, err
	}
	buffer := make([]C.SidereonObservationReliability, n)
	outputLength, err := cSize(n, "reliability observation output capacity")
	if err != nil {
		return nil, err
	}
	var output *C.SidereonObservationReliability
	if n > 0 {
		output = &buffer[0]
	}
	written, required = 0, 0
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_reliability_report_observations((*C.SidereonReliabilityReport)(p), output, outputLength, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	actual, err := validateTwoPassCounts("reliability observations", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	out := make([]NativeReliabilityObservation, actual)
	for j := range out {
		v := buffer[j]
		out[j] = NativeReliabilityObservation{ID: rtkIDFromC(v.id).Value, Redundancy: float64(v.redundancy), HasMDB: bool(v.has_mdb_m), MDB: float64(v.mdb_m), HasExternal: bool(v.has_external_enu_m), HasBiasToNoise: bool(v.has_bias_to_noise), BiasToNoise: float64(v.bias_to_noise), Uncheckable: bool(v.uncheckable)}
		for k := 0; k < 3; k++ {
			out[j].External[k] = float64(v.external_enu_m[k])
		}
	}
	return out, nil
}

func RangeFDEOptionsDefault() (NativeRangeFDEOptions, error) {
	var o C.SidereonRangeFdeOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_range_fde_options_init(&o)) })
	return NativeRangeFDEOptions{float64(o.p_fa), uint64(o.max_exclusions), uint64(o.min_redundancy)}, err
}

func RangeFDE(rows []NativeRangeFDERow, options NativeRangeFDEOptions) (*RangeFDEResult, error) {
	rowCount, err := cSize(len(rows), "range FDE row count")
	if err != nil {
		return nil, err
	}
	maxExclusions, err := cSize64(options.MaxExclusions, "range FDE maximum exclusions")
	if err != nil {
		return nil, err
	}
	minRedundancy, err := cSize64(options.MinRedundancy, "range FDE minimum redundancy")
	if err != nil {
		return nil, err
	}
	size, err := checkedNativeAllocationSize(len(rows), unsafe.Sizeof(C.SidereonRangeFdeRow{}))
	if err != nil {
		return nil, err
	}
	var memory unsafe.Pointer
	if len(rows) > 0 {
		memory = C.malloc(C.size_t(size))
		if memory == nil {
			return nil, errors.New("sidereon: unable to allocate range FDE rows")
		}
		defer C.free(memory)
	}
	input := unsafe.Slice((*C.SidereonRangeFdeRow)(memory), len(rows))
	ids := make([]*C.char, len(rows))
	defer freeCStrings(ids)
	designs := make([]unsafe.Pointer, len(rows))
	defer func() {
		for _, p := range designs {
			if p != nil {
				C.free(p)
			}
		}
	}()
	for i, row := range rows {
		if err := rejectEmbeddedNUL(row.ID, "range FDE observation ID"); err != nil {
			return nil, err
		}
		ids[i] = C.CString(row.ID)
		if ids[i] == nil {
			return nil, errors.New("sidereon: unable to allocate range FDE observation ID")
		}
		p, designLength, err := cFloats(row.Design, "range FDE design row")
		if err != nil {
			return nil, err
		}
		designs[i] = p
		input[i] = C.SidereonRangeFdeRow{id: ids[i], residual_m: C.double(row.Residual), design_row: (*C.double)(p), design_dim: designLength, weight: C.double(row.Weight)}
	}
	optionsC := C.SidereonRangeFdeOptions{p_fa: C.double(options.PFA), max_exclusions: maxExclusions, min_redundancy: minRedundancy}
	var out *C.SidereonRangeFdeResult
	var opErr error
	withCThread(func() {
		opErr = statusErrorLocked(uint32(C.sidereon_raim_fde_design((*C.SidereonRangeFdeRow)(memory), rowCount, &optionsC, &out)))
	})
	if opErr != nil {
		return nil, opErr
	}
	h, err := newSurfaceHandle(unsafe.Pointer(out), func(p unsafe.Pointer) { C.sidereon_range_fde_result_free((*C.SidereonRangeFdeResult)(p)) })
	if err != nil {
		return nil, err
	}
	return &RangeFDEResult{handle: h}, nil
}

func (r *RangeFDEResult) Close() error {
	if r == nil {
		return nil
	}
	return r.handle.close()
}
func (r *RangeFDEResult) Values(kind uint32) ([]float64, error) {
	if kind == 0 {
		return copyTrackDoubles(r.handle, "range FDE correction", func(p unsafe.Pointer, o *C.double, l C.size_t, w, q *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_range_fde_result_state_correction((*C.SidereonRangeFdeResult)(p), o, l, w, q)
		})
	}
	return copyTrackDoubles(r.handle, "range FDE covariance", func(p unsafe.Pointer, o *C.double, l C.size_t, w, q *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_range_fde_result_covariance((*C.SidereonRangeFdeResult)(p), o, l, w, q)
	})
}
func (r *RangeFDEResult) Output() (NativeRangeFDEOutput, error) {
	var dimension, iterations C.size_t
	var global C.SidereonRangeChiSquareTest
	var opErr error
	err := r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_range_fde_result_state_dim((*C.SidereonRangeFdeResult)(p), &dimension)))
			if opErr == nil {
				opErr = statusErrorLocked(uint32(C.sidereon_range_fde_result_iterations((*C.SidereonRangeFdeResult)(p), &iterations)))
			}
			if opErr == nil {
				opErr = statusErrorLocked(uint32(C.sidereon_range_fde_result_global_test((*C.SidereonRangeFdeResult)(p), &global)))
			}
		})
		return nil
	})
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	if opErr != nil {
		return NativeRangeFDEOutput{}, opErr
	}
	correction, err := r.Values(0)
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	covariance, err := r.Values(1)
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	var excluded []string
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			excluded, err = copyNativeRtkIDs(func(o *C.SidereonRtkId, l C.size_t, w, q *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_range_fde_result_excluded((*C.SidereonRangeFdeResult)(p), o, l, w, q)
			}, "range FDE excluded")
		})
		return nil
	})
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	var written, required C.size_t
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_range_fde_result_diagnostics((*C.SidereonRangeFdeResult)(p), nil, 0, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	if opErr != nil {
		return NativeRangeFDEOutput{}, opErr
	}
	n, err := sizeTToInt(required, "range FDE diagnostics count")
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	if _, err = writtenToInt(written, 0, "range FDE diagnostics first-call written count"); err != nil {
		return NativeRangeFDEOutput{}, err
	}
	diagnostics := make([]C.SidereonRangeFdeDiagnostic, n)
	outputLength, err := cSize(n, "range FDE diagnostics output capacity")
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	var output *C.SidereonRangeFdeDiagnostic
	if n > 0 {
		output = &diagnostics[0]
	}
	written, required = 0, 0
	err = r.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_range_fde_result_diagnostics((*C.SidereonRangeFdeResult)(p), output, outputLength, &written, &required)))
		})
		return nil
	})
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	if opErr != nil {
		return NativeRangeFDEOutput{}, opErr
	}
	actual, err := validateTwoPassCounts("range FDE diagnostics", n, n, uint64(written), uint64(required))
	if err != nil {
		return NativeRangeFDEOutput{}, err
	}
	outd := make([]NativeRangeFDEDiagnostic, actual)
	for i := range outd {
		outd[i] = NativeRangeFDEDiagnostic{rtkIDFromC(diagnostics[i].id).Value, bool(diagnostics[i].excluded), float64(diagnostics[i].post_fit_residual_m), float64(diagnostics[i].normalized_residual)}
	}
	return NativeRangeFDEOutput{uint64(dimension), correction, covariance, NativeRangeFDEGlobalTest{float64(global.weighted_sum_squares), int64(global.dof), bool(global.has_threshold), float64(global.threshold), bool(global.testable), bool(global.fault_detected)}, uint64(iterations), excluded, outd}, nil
}
