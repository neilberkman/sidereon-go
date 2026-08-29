//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type RtkSatMeasurement struct {
	SatelliteID, SDAmbiguityID                     string
	BaseCodeM, BasePhaseM, RoverCodeM, RoverPhaseM float64
	BaseTXPos, RoverTXPos, Pos                     [3]float64
}

type RtkEpoch struct {
	References, NonReference []RtkSatMeasurement
	HasVelocityMPS           bool
	VelocityMPS              [3]float64
	DTS                      float64
}

type RtkAmbiguitySatellite struct{ ID, SatelliteID string }
type RtkFloatMapEntry struct {
	ID    string
	Value float64
}

type RtkFloatConfig struct {
	Epochs           []RtkEpoch
	BaseECEFM        [3]float64
	AmbiguityIDs     []string
	Model            RtkMeasurementModel
	InitialBaselineM [3]float64
	Options          RtkFloatOptions
}

type RtkFloatMetadata struct {
	Iterations, NObservations, AmbiguityCount, ResidualCount, UsedSatCount int
	Converged                                                              bool
	Status                                                                 uint32
	CodeRMSM, PhaseRMSM, WeightedRMSM                                      float64
	GeometryQuality                                                        GeometryQuality
}

type RtkAmbiguity struct {
	ID     string
	ValueM float64
}

type RtkFloatSolution struct {
	_      noCopy
	handle *surfaceHandle
}

func newRtkFloatSolution(pointer *C.SidereonRtkFloatSolution) (*RtkFloatSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_rtk_float_solution_free((*C.SidereonRtkFloatSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &RtkFloatSolution{handle: h}, nil
}

type cRtkAlloc struct{ values []unsafe.Pointer }

func (a *cRtkAlloc) malloc(bytes uintptr, label string) (unsafe.Pointer, error) {
	if bytes == 0 {
		return nil, nil
	}
	p := C.malloc(C.size_t(bytes))
	if p == nil {
		return nil, fmt.Errorf("sidereon: unable to allocate native %s", label)
	}
	a.values = append(a.values, p)
	return p, nil
}
func (a *cRtkAlloc) cstring(value, label string) (*C.char, error) {
	p, err := copyNativeCString(value, label)
	if err != nil {
		return nil, err
	}
	a.values = append(a.values, p)
	return (*C.char)(p), nil
}
func (a *cRtkAlloc) close() {
	for i := len(a.values) - 1; i >= 0; i-- {
		C.free(a.values[i])
	}
}

func fillRtkMeasurement(dst *C.SidereonRtkSatMeasurement, value RtkSatMeasurement, alloc *cRtkAlloc) error {
	sat, err := alloc.cstring(value.SatelliteID, "RTK satellite ID")
	if err != nil {
		return err
	}
	ambiguity, err := alloc.cstring(value.SDAmbiguityID, "RTK ambiguity ID")
	if err != nil {
		return err
	}
	dst.sat_id, dst.sd_ambiguity_id = sat, ambiguity
	dst.base_code_m, dst.base_phase_m = C.double(value.BaseCodeM), C.double(value.BasePhaseM)
	dst.rover_code_m, dst.rover_phase_m = C.double(value.RoverCodeM), C.double(value.RoverPhaseM)
	for i := range value.BaseTXPos {
		dst.base_tx_pos[i], dst.rover_tx_pos[i], dst.pos[i] = C.double(value.BaseTXPos[i]), C.double(value.RoverTXPos[i]), C.double(value.Pos[i])
	}
	return nil
}

func copyRtkEpochs(values []RtkEpoch, alloc *cRtkAlloc) (*C.SidereonRtkEpoch, error) {
	epochBytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonRtkEpoch{}))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	memory, err := alloc.malloc(epochBytes, "RTK epochs")
	if err != nil {
		return nil, err
	}
	dst := unsafe.Slice((*C.SidereonRtkEpoch)(memory), len(values))
	for i, epoch := range values {
		dst[i] = C.SidereonRtkEpoch{}
		referenceCount, err := checkedNativeSize(len(epoch.References))
		if err != nil {
			return nil, err
		}
		nonReferenceCount, err := checkedNativeSize(len(epoch.NonReference))
		if err != nil {
			return nil, err
		}
		all := append(append([]RtkSatMeasurement(nil), epoch.References...), epoch.NonReference...)
		rowBytes, err := checkedNativeAllocationSize(len(all), unsafe.Sizeof(C.SidereonRtkSatMeasurement{}))
		if err != nil {
			return nil, err
		}
		var rows []C.SidereonRtkSatMeasurement
		if len(all) != 0 {
			rowMemory, e := alloc.malloc(rowBytes, "RTK measurements")
			if e != nil {
				return nil, e
			}
			rows = unsafe.Slice((*C.SidereonRtkSatMeasurement)(rowMemory), len(all))
			for j, measurement := range all {
				if e := fillRtkMeasurement(&rows[j], measurement, alloc); e != nil {
					return nil, e
				}
			}
		}
		if len(epoch.References) != 0 {
			dst[i].references = &rows[0]
		}
		dst[i].reference_count = referenceCount
		if len(epoch.NonReference) != 0 {
			dst[i].nonref = &rows[len(epoch.References)]
		}
		dst[i].nonref_count = nonReferenceCount
		dst[i].has_velocity_mps = C.bool(epoch.HasVelocityMPS)
		for j := range epoch.VelocityMPS {
			dst[i].velocity_mps[j] = C.double(epoch.VelocityMPS[j])
		}
		dst[i].dt_s = C.double(epoch.DTS)
	}
	return &dst[0], nil
}

func (c RtkFloatConfig) toC(alloc *cRtkAlloc) (*C.SidereonRtkFloatConfig, error) {
	epochs, err := copyRtkEpochs(c.Epochs, alloc)
	if err != nil {
		return nil, err
	}
	maxIterations, err := checkedNativeSize(c.Options.MaxIterations)
	if err != nil {
		return nil, fmt.Errorf("sidereon: RTK float max iterations: %w", err)
	}
	var ids **C.char
	if len(c.AmbiguityIDs) != 0 {
		idBytes, e := checkedNativeAllocationSize(len(c.AmbiguityIDs), unsafe.Sizeof(uintptr(0)))
		if e != nil {
			return nil, e
		}
		memory, e := alloc.malloc(idBytes, "RTK ambiguity IDs")
		if e != nil {
			return nil, e
		}
		rows := unsafe.Slice((**C.char)(memory), len(c.AmbiguityIDs))
		for i, value := range c.AmbiguityIDs {
			rows[i], e = alloc.cstring(value, "RTK ambiguity ID")
			if e != nil {
				return nil, e
			}
		}
		ids = (**C.char)(memory)
	}
	model := C.SidereonRtkMeasurementModel{code_sigma_m: C.double(c.Model.CodeSigmaM), phase_sigma_m: C.double(c.Model.PhaseSigmaM), sagnac: C.bool(c.Model.Sagnac), stochastic: C.uint32_t(c.Model.Stochastic), elevation_weighting: C.bool(c.Model.ElevationWeighting)}
	floatOptions := C.SidereonRtkFloatOptions{position_tol_m: C.double(c.Options.PositionTolM), ambiguity_tol_m: C.double(c.Options.AmbiguityTolM), max_iterations: maxIterations}
	configBytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonRtkFloatConfig{}))
	if err != nil {
		return nil, err
	}
	memory, err := alloc.malloc(configBytes, "RTK float config")
	if err != nil {
		return nil, err
	}
	config := (*C.SidereonRtkFloatConfig)(memory)
	*config = C.SidereonRtkFloatConfig{epochs: epochs, epoch_count: C.size_t(len(c.Epochs)), base_ecef_m: [3]C.double{C.double(c.BaseECEFM[0]), C.double(c.BaseECEFM[1]), C.double(c.BaseECEFM[2])}, ambiguity_ids: ids, ambiguity_id_count: C.size_t(len(c.AmbiguityIDs)), model: model, initial_baseline_m: [3]C.double{C.double(c.InitialBaselineM[0]), C.double(c.InitialBaselineM[1]), C.double(c.InitialBaselineM[2])}, options: floatOptions}
	return config, nil
}

func SolveRtkFloat(config RtkFloatConfig) (*RtkFloatSolution, error) {
	alloc := new(cRtkAlloc)
	defer alloc.close()
	cConfig, err := config.toC(alloc)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonRtkFloatSolution
	var operationErr error
	withCThread(func() {
		status := C.sidereon_solve_rtk_float(cConfig, &pointer)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && pointer != nil {
			C.sidereon_rtk_float_solution_free(pointer)
			pointer = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newRtkFloatSolution(pointer)
}

func (s *RtkFloatSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *RtkFloatSolution) BaselineECEF() ([3]float64, error) {
	var result [3]float64
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	var values [3]C.double
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_float_solution_baseline_ecef((*C.SidereonRtkFloatSolution)(pointer), &values[0], 3))
		})
	})
	if err != nil {
		return result, err
	}
	for i := range result {
		result[i] = float64(values[i])
	}
	return result, err
}

func (s *RtkFloatSolution) BaselineENU() ([3]float64, error) {
	var result [3]float64
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	var values [3]C.double
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_float_solution_baseline_enu((*C.SidereonRtkFloatSolution)(pointer), &values[0], 3))
		})
	})
	if err != nil {
		return result, err
	}
	for i := range result {
		result[i] = float64(values[i])
	}
	return result, nil
}

func (s *RtkFloatSolution) Metadata() (RtkFloatMetadata, error) {
	var value C.SidereonRtkFloatMetadata
	if s == nil || s.handle == nil {
		return RtkFloatMetadata{}, ErrClosed
	}
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_rtk_float_solution_metadata((*C.SidereonRtkFloatSolution)(pointer), &value))
		})
	})
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	iterations, err := sizeTToInt(value.iterations, "RTK float metadata iterations")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	nObservations, err := sizeTToInt(value.n_observations, "RTK float metadata observations")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	ambiguityCount, err := sizeTToInt(value.ambiguity_count, "RTK float metadata ambiguities")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	residualCount, err := sizeTToInt(value.residual_count, "RTK float metadata residuals")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	usedSatCount, err := sizeTToInt(value.used_sat_count, "RTK float metadata satellites")
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	geometry, err := geometryFromC(value.geometry_quality)
	if err != nil {
		return RtkFloatMetadata{}, err
	}
	return RtkFloatMetadata{Iterations: iterations, NObservations: nObservations, AmbiguityCount: ambiguityCount, ResidualCount: residualCount, UsedSatCount: usedSatCount, Converged: bool(value.converged), Status: uint32(value.status), CodeRMSM: float64(value.code_rms_m), PhaseRMSM: float64(value.phase_rms_m), WeightedRMSM: float64(value.weighted_rms_m), GeometryQuality: geometry}, nil
}

func (s *RtkFloatSolution) Ambiguities() ([]RtkAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []RtkAmbiguity
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var written, required C.size_t
			status := C.sidereon_rtk_float_solution_ambiguities((*C.SidereonRtkFloatSolution)(pointer), nil, 0, &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			count, e := validateNativeQuery("RTK float ambiguities", uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			if _, e = checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonRtkAmbiguity{})); e != nil {
				operationErr = e
				return
			}
			rows := make([]C.SidereonRtkAmbiguity, count)
			var output *C.SidereonRtkAmbiguity
			if count != 0 {
				output = &rows[0]
			}
			written, required = 0, 0
			status = C.sidereon_rtk_float_solution_ambiguities((*C.SidereonRtkFloatSolution)(pointer), output, C.size_t(count), &written, &required)
			if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
				return
			}
			w, e := validateTwoPassCounts("RTK float ambiguities", count, count, uint64(written), uint64(required))
			if e != nil {
				operationErr = e
				return
			}
			result = make([]RtkAmbiguity, w)
			for i := range result {
				result[i] = RtkAmbiguity{ID: rtkIDFromC(rows[i].id).Value, ValueM: float64(rows[i].value_m)}
			}
		})
		return operationErr
	})
	return result, err
}

func (s *RtkFloatSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []string
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeTokensLocked("RTK float used satellites", func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_rtk_float_solution_used_sat_ids((*C.SidereonRtkFloatSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	return result, err
}
