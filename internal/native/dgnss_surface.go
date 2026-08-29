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

type DgnssCodeObservation struct {
	SatelliteID  string
	PseudorangeM float64
}

type DgnssAppliedRow struct {
	SatelliteID  string
	PseudorangeM float64
}

type DgnssCorrections struct {
	_      noCopy
	handle *surfaceHandle
}

type DgnssApplied struct {
	_      noCopy
	handle *surfaceHandle
}

type DgnssSolution struct {
	_      noCopy
	handle *surfaceHandle
}

func newDgnssCorrections(pointer *C.SidereonDgnssCorrections) (*DgnssCorrections, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_dgnss_corrections_free((*C.SidereonDgnssCorrections)(value))
	})
	if err != nil {
		return nil, err
	}
	return &DgnssCorrections{handle: handle}, nil
}

func newDgnssApplied(pointer *C.SidereonDgnssApplied) (*DgnssApplied, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_dgnss_applied_free((*C.SidereonDgnssApplied)(value))
	})
	if err != nil {
		return nil, err
	}
	return &DgnssApplied{handle: handle}, nil
}

func newDgnssSolution(pointer *C.SidereonDgnssSolution) (*DgnssSolution, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_dgnss_solution_free((*C.SidereonDgnssSolution)(value))
	})
	if err != nil {
		return nil, err
	}
	return &DgnssSolution{handle: handle}, nil
}

func makeDgnssObservations(values []DgnssCodeObservation, alloc *cRtkAlloc) (*C.SidereonCodeObservation, error) {
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonCodeObservation{}))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	memory, err := alloc.malloc(bytes, "DGNSS observations")
	if err != nil {
		return nil, err
	}
	rows := unsafe.Slice((*C.SidereonCodeObservation)(memory), len(values))
	for i, value := range values {
		id, err := alloc.cstring(value.SatelliteID, "DGNSS satellite ID")
		if err != nil {
			return nil, err
		}
		rows[i].sat_id = id
		rows[i].pseudorange_m = C.double(value.PseudorangeM)
	}
	return &rows[0], nil
}

func makeDgnssBaseInputs(config SPPConfig, alloc *cRtkAlloc) (*C.SidereonSppInputsV2, error) {
	count, err := checkedNativeSize(len(config.Observations))
	if err != nil {
		return nil, err
	}
	for _, observation := range config.Observations {
		if err := validateSPPSatelliteID(observation.SatelliteID); err != nil {
			return nil, err
		}
	}
	observationBytes, err := checkedNativeAllocationSize(len(config.Observations), unsafe.Sizeof(C.SidereonObservation{}))
	if err != nil {
		return nil, err
	}
	var observations *C.SidereonObservation
	if len(config.Observations) != 0 {
		memory, e := alloc.malloc(observationBytes, "DGNSS SPP observations")
		if e != nil {
			return nil, e
		}
		rows := unsafe.Slice((*C.SidereonObservation)(memory), len(config.Observations))
		for i, observation := range config.Observations {
			id, e := alloc.cstring(observation.SatelliteID, "DGNSS SPP satellite ID")
			if e != nil {
				return nil, e
			}
			rows[i].sat_id = id
			rows[i].pseudorange_m = C.double(observation.PseudorangeM)
		}
		observations = &rows[0]
	}
	inputBytes, err := checkedNativeAllocationSize(1, unsafe.Sizeof(C.SidereonSppInputsV2{}))
	if err != nil {
		return nil, err
	}
	inputMemory, err := alloc.malloc(inputBytes, "DGNSS SPP inputs")
	if err != nil {
		return nil, err
	}
	inputs := (*C.SidereonSppInputsV2)(inputMemory)
	var initErr error
	withCThread(func() {
		initErr = statusErrorLocked(uint32(C.sidereon_spp_inputs_v2_init(inputs)))
	})
	if initErr != nil {
		return nil, initErr
	}
	inputs.base.observations = observations
	inputs.base.observation_count = count
	inputs.base.t_rx_j2000_s = C.double(config.TRxJ2000S)
	inputs.base.t_rx_second_of_day_s = C.double(config.TRxSecondOfDayS)
	inputs.base.day_of_year = C.double(config.DayOfYear)
	for i := range config.InitialGuess {
		inputs.base.initial_guess[i] = C.double(config.InitialGuess[i])
	}
	inputs.base.ionosphere = C.bool(config.Ionosphere)
	inputs.base.troposphere = C.bool(config.Troposphere)
	inputs.base.with_geodetic = C.bool(config.WithGeodetic)
	return inputs, nil
}

func (c *DgnssCorrections) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return c.handle.close()
}

func (c *DgnssCorrections) Count() (int, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	var operationErr error
	err := c.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_dgnss_corrections_count((*C.SidereonDgnssCorrections)(pointer), &count)
			operationErr = statusErrorLocked(uint32(status))
		})
		return operationErr
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(count, "DGNSS correction count")
}

func (c *DgnssCorrections) Correction(satelliteID string) (float64, bool, error) {
	if c == nil || c.handle == nil {
		return 0, false, ErrClosed
	}
	id, err := copyNativeCString(satelliteID, "DGNSS satellite ID")
	if err != nil {
		return 0, false, err
	}
	defer C.free(id)
	var value C.double
	var present C.bool
	var operationErr error
	err = c.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_dgnss_correction((*C.SidereonDgnssCorrections)(pointer), (*C.char)(id), &value, &present)
			operationErr = statusErrorLocked(uint32(status))
		})
		return operationErr
	})
	return float64(value), bool(present), err
}

func DgnssPseudorangeCorrections(sp3 *SP3, basePosition [3]float64, observations []DgnssCodeObservation, receiveTimeJ2000S float64) (*DgnssCorrections, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	base, err := makeDgnssObservations(observations, alloc)
	if err != nil {
		return nil, err
	}
	baseCount, err := checkedNativeSize(len(observations))
	if err != nil {
		return nil, err
	}
	var basePositionC [3]C.double
	for i := range basePosition {
		basePositionC[i] = C.double(basePosition[i])
	}
	var pointer *C.SidereonDgnssCorrections
	var operationErr error
	err = sp3.handle.with(func(sp3Pointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_dgnss_pseudorange_corrections((*C.SidereonSp3)(sp3Pointer), &basePositionC[0], base, baseCount, C.double(receiveTimeJ2000S), &pointer)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && pointer != nil {
				C.sidereon_dgnss_corrections_free(pointer)
				pointer = nil
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	return newDgnssCorrections(pointer)
}

func (a *DgnssApplied) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}

func (a *DgnssApplied) Counts() (int, int, error) {
	if a == nil || a.handle == nil {
		return 0, 0, ErrClosed
	}
	var corrected, dropped C.size_t
	var operationErr error
	err := a.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			status := C.sidereon_dgnss_applied_counts((*C.SidereonDgnssApplied)(pointer), &corrected, &dropped)
			operationErr = statusErrorLocked(uint32(status))
		})
		return operationErr
	})
	if err != nil {
		return 0, 0, err
	}
	correctedCount, err := sizeTToInt(corrected, "DGNSS corrected count")
	if err != nil {
		return 0, 0, err
	}
	droppedCount, err := sizeTToInt(dropped, "DGNSS dropped count")
	return correctedCount, droppedCount, err
}

func (a *DgnssApplied) Corrected(index int) (DgnssAppliedRow, error) {
	if a == nil || a.handle == nil {
		return DgnssAppliedRow{}, ErrClosed
	}
	if index < 0 {
		return DgnssAppliedRow{}, invalidArgument("DGNSS corrected index must not be negative")
	}
	nativeIndex, err := cSize(index, "DGNSS corrected index")
	if err != nil {
		return DgnssAppliedRow{}, err
	}
	var id [17]C.char
	idLength, err := cSize(len(id), "DGNSS corrected satellite ID output length")
	if err != nil {
		return DgnssAppliedRow{}, err
	}
	var pseudorange C.double
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_dgnss_applied_corrected((*C.SidereonDgnssApplied)(pointer), nativeIndex, &id[0], idLength, &pseudorange))
		})
	})
	return DgnssAppliedRow{SatelliteID: fixedCStringArray(id[:]), PseudorangeM: float64(pseudorange)}, err
}

func (a *DgnssApplied) Dropped(index int) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	if index < 0 {
		return "", invalidArgument("DGNSS dropped index must not be negative")
	}
	nativeIndex, err := cSize(index, "DGNSS dropped index")
	if err != nil {
		return "", err
	}
	var id [17]C.char
	idLength, err := cSize(len(id), "DGNSS dropped satellite ID output length")
	if err != nil {
		return "", err
	}
	err = a.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_dgnss_applied_dropped((*C.SidereonDgnssApplied)(pointer), nativeIndex, &id[0], idLength))
		})
	})
	return fixedCStringArray(id[:]), err
}

func ApplyDgnssCorrections(observations []DgnssCodeObservation, corrections *DgnssCorrections) (*DgnssApplied, error) {
	if corrections == nil || corrections.handle == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	rover, err := makeDgnssObservations(observations, alloc)
	if err != nil {
		return nil, err
	}
	roverCount, err := checkedNativeSize(len(observations))
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonDgnssApplied
	err = corrections.handle.read(func(correctionsPointer unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			status := C.sidereon_dgnss_apply_corrections(rover, roverCount, (*C.SidereonDgnssCorrections)(correctionsPointer), &pointer)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && pointer != nil {
				C.sidereon_dgnss_applied_free(pointer)
				pointer = nil
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	return newDgnssApplied(pointer)
}

func SolveDgnssPosition(sp3 *SP3, basePosition [3]float64, base, rover []DgnssCodeObservation, config SPPConfig) (*DgnssSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	alloc := new(cRtkAlloc)
	defer alloc.close()
	baseObservations, err := makeDgnssObservations(base, alloc)
	if err != nil {
		return nil, err
	}
	roverObservations, err := makeDgnssObservations(rover, alloc)
	if err != nil {
		return nil, err
	}
	inputs, err := makeDgnssBaseInputs(config, alloc)
	if err != nil {
		return nil, err
	}
	baseCount, err := checkedNativeSize(len(base))
	if err != nil {
		return nil, err
	}
	roverCount, err := checkedNativeSize(len(rover))
	if err != nil {
		return nil, err
	}
	var basePositionC [3]C.double
	for i := range basePosition {
		basePositionC[i] = C.double(basePosition[i])
	}
	var pointer *C.SidereonDgnssSolution
	err = sp3.handle.with(func(sp3Pointer unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			status := C.sidereon_dgnss_position_solve((*C.SidereonSp3)(sp3Pointer), &basePositionC[0], baseObservations, baseCount, roverObservations, roverCount, inputs, &pointer)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && pointer != nil {
				C.sidereon_dgnss_solution_free(pointer)
				pointer = nil
			}
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	return newDgnssSolution(pointer)
}

func (s *DgnssSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func (s *DgnssSolution) Baseline() ([3]float64, float64, error) {
	var vector [3]float64
	if s == nil || s.handle == nil {
		return vector, 0, ErrClosed
	}
	var nativeVector [3]C.double
	var length C.double
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_dgnss_solution_baseline((*C.SidereonDgnssSolution)(pointer), &nativeVector[0], C.size_t(len(nativeVector)), &length))
		})
	})
	for i := range vector {
		vector[i] = float64(nativeVector[i])
	}
	return vector, float64(length), err
}

func (s *DgnssSolution) DroppedSatellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var result []string
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			result, operationErr = copyNativeTokensLocked("DGNSS dropped satellites", func(out *C.SidereonSatelliteToken, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_dgnss_solution_dropped_sats((*C.SidereonDgnssSolution)(pointer), out, length, written, required)
			})
		})
		return operationErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *DgnssSolution) Solution() (SPPSolution, error) {
	if s == nil || s.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	var result SPPSolution
	var operationErr error
	err := s.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var nested *C.SidereonSppSolution
			status := C.sidereon_dgnss_solution_solution((*C.SidereonDgnssSolution)(pointer), &nested)
			if callErr := statusErrorLocked(uint32(status)); callErr != nil {
				operationErr = callErr
				if nested != nil {
					C.sidereon_spp_solution_free(nested)
				}
				return
			}
			if nested == nil {
				operationErr = errors.New("sidereon: native DGNSS solution returned no SPP solution")
				return
			}
			defer C.sidereon_spp_solution_free(nested)
			result, operationErr = readSPPSolutionLocked(nested)
		})
		return operationErr
	})
	return result, err
}
