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
	"fmt"
	"runtime"
	"unsafe"
)

// ArcEpoch is one dual-frequency observation epoch for the carrier-phase
// preprocessing functions. The C ABI uses NaN for absent scalar values.
type ArcEpoch struct {
	Phi1Cycles, Phi2Cycles float64
	P1M, P2M               float64
	HasLLI1                bool
	LLI1                   int64
	HasLLI2                bool
	LLI2                   int64
	F1Hz, F2Hz             float64
	GapTimeS               float64
}

type CycleSlipOptions struct {
	GFThresholdM, MWThresholdCycles, MinArcGapS float64
}

type SlipResult struct {
	Slip       bool
	ReasonMask uint32
	GFM, MWM   float64
	Skipped    bool
}

type SmoothCodeResult struct {
	PSmoothM float64
	Window   int
	Reset    bool
}

type IonoFreeSmoothResult struct {
	PSmoothM, PIFM, LIFM float64
	Window               int
	Reset                bool
}

func arcEpochToC(v ArcEpoch) C.SidereonArcEpoch {
	return C.SidereonArcEpoch{
		phi1_cycles: C.double(v.Phi1Cycles), phi2_cycles: C.double(v.Phi2Cycles),
		p1_m: C.double(v.P1M), p2_m: C.double(v.P2M), has_lli1: C.bool(v.HasLLI1),
		lli1: C.int64_t(v.LLI1), has_lli2: C.bool(v.HasLLI2), lli2: C.int64_t(v.LLI2),
		f1_hz: C.double(v.F1Hz), f2_hz: C.double(v.F2Hz), gap_time_s: C.double(v.GapTimeS),
	}
}

func CycleSlipOptionsInit() (CycleSlipOptions, error) {
	var c C.SidereonCycleSlipOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_cycle_slip_options_init(&c)) })
	return CycleSlipOptions{float64(c.gf_threshold_m), float64(c.mw_threshold_cycles), float64(c.min_arc_gap_s)}, err
}

func copyArcEpochs(input []ArcEpoch) ([]C.SidereonArcEpoch, error) {
	if _, err := checkedNativeAllocationSize(len(input), unsafe.Sizeof(C.SidereonArcEpoch{})); err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return nil, nil
	}
	result := make([]C.SidereonArcEpoch, len(input))
	for i, value := range input {
		result[i] = arcEpochToC(value)
	}
	return result, nil
}

func DetectCycleSlips(input []ArcEpoch, options *CycleSlipOptions) ([]SlipResult, error) {
	epochs, err := copyArcEpochs(input)
	if err != nil {
		return nil, err
	}
	var out []C.SidereonSlipResult
	var operationErr error
	withCThread(func() {
		n, sizeErr := checkedNativeSize(len(input))
		if sizeErr != nil {
			operationErr = sizeErr
			return
		}
		var epochPointer *C.SidereonArcEpoch
		if len(epochs) != 0 {
			epochPointer = &epochs[0]
		}
		var option C.SidereonCycleSlipOptions
		var optionPointer *C.SidereonCycleSlipOptions
		if options != nil {
			option = C.SidereonCycleSlipOptions{gf_threshold_m: C.double(options.GFThresholdM), mw_threshold_cycles: C.double(options.MWThresholdCycles), min_arc_gap_s: C.double(options.MinArcGapS)}
			optionPointer = &option
		}
		if _, err := checkedNativeAllocationSize(len(input), unsafe.Sizeof(C.SidereonSlipResult{})); err != nil {
			operationErr = err
			return
		}
		out = make([]C.SidereonSlipResult, len(input))
		var outPointer *C.SidereonSlipResult
		if len(out) != 0 {
			outPointer = &out[0]
		}
		var written, required C.size_t
		status := C.sidereon_detect_cycle_slips(epochPointer, n, optionPointer, nil, 0, &written, &required)
		if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
			return
		}
		count, queryErr := validateNativeQuery("cycle-slip results", uint64(written), uint64(required))
		if queryErr != nil || count != len(input) {
			if queryErr != nil {
				operationErr = queryErr
			} else {
				operationErr = fmt.Errorf("sidereon: native cycle-slip result count %d does not match input count %d", count, len(input))
			}
			return
		}
		written, required = 0, 0
		status = C.sidereon_detect_cycle_slips(epochPointer, n, optionPointer, outPointer, C.size_t(len(out)), &written, &required)
		if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
			return
		}
		var countErr error
		count, countErr = validateTwoPassCounts("cycle-slip results", len(out), len(input), uint64(written), uint64(required))
		if countErr != nil {
			operationErr = countErr
			return
		}
		out = out[:count]
	})
	runtime.KeepAlive(input)
	if operationErr != nil {
		return nil, operationErr
	}
	result := make([]SlipResult, len(out))
	for i, value := range out {
		result[i] = SlipResult{Slip: bool(value.slip), ReasonMask: uint32(value.reason_mask), GFM: float64(value.gf_m), MWM: float64(value.mw_m), Skipped: bool(value.skipped)}
	}
	return result, nil
}

func SmoothCode(input []ArcEpoch, options *CycleSlipOptions, hatchWindowCap int) ([]SmoothCodeResult, error) {
	if hatchWindowCap < 0 {
		return nil, invalidArgument("hatch window cap must not be negative")
	}
	epochs, err := copyArcEpochs(input)
	if err != nil {
		return nil, err
	}
	var out []C.SidereonSmoothCodeResult
	var operationErr error
	withCThread(func() {
		var option C.SidereonCycleSlipOptions
		var optionPointer *C.SidereonCycleSlipOptions
		if options != nil {
			option = C.SidereonCycleSlipOptions{gf_threshold_m: C.double(options.GFThresholdM), mw_threshold_cycles: C.double(options.MWThresholdCycles), min_arc_gap_s: C.double(options.MinArcGapS)}
			optionPointer = &option
		}
		var epochPointer *C.SidereonArcEpoch
		if len(epochs) != 0 {
			epochPointer = &epochs[0]
		}
		if _, err := checkedNativeAllocationSize(len(input), unsafe.Sizeof(C.SidereonSmoothCodeResult{})); err != nil {
			operationErr = err
			return
		}
		out = make([]C.SidereonSmoothCodeResult, len(input))
		var output *C.SidereonSmoothCodeResult
		if len(out) != 0 {
			output = &out[0]
		}
		var written, required C.size_t
		status := C.sidereon_smooth_code(epochPointer, C.size_t(len(input)), optionPointer, C.size_t(hatchWindowCap), nil, 0, &written, &required)
		if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
			return
		}
		count, queryErr := validateNativeQuery("smoothed-code results", uint64(written), uint64(required))
		if queryErr != nil || count != len(input) {
			if queryErr != nil {
				operationErr = queryErr
			} else {
				operationErr = fmt.Errorf("sidereon: native smoothed-code result count %d does not match input count %d", count, len(input))
			}
			return
		}
		written, required = 0, 0
		status = C.sidereon_smooth_code(epochPointer, C.size_t(len(input)), optionPointer, C.size_t(hatchWindowCap), output, C.size_t(len(out)), &written, &required)
		if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
			return
		}
		var countErr error
		count, countErr = validateTwoPassCounts("smoothed-code results", len(out), len(input), uint64(written), uint64(required))
		if countErr != nil {
			operationErr = countErr
			return
		}
		out = out[:count]
	})
	runtime.KeepAlive(input)
	if operationErr != nil {
		return nil, operationErr
	}
	result := make([]SmoothCodeResult, len(out))
	for i, value := range out {
		result[i] = SmoothCodeResult{PSmoothM: float64(value.p_smooth_m), Window: int(value.window), Reset: bool(value.reset)}
	}
	return result, nil
}

func SmoothIonoFreeCode(input []ArcEpoch, options *CycleSlipOptions, hatchWindowCap int) ([]IonoFreeSmoothResult, error) {
	if hatchWindowCap < 0 {
		return nil, invalidArgument("hatch window cap must not be negative")
	}
	epochs, err := copyArcEpochs(input)
	if err != nil {
		return nil, err
	}
	var out []C.SidereonIonoFreeSmoothResult
	var operationErr error
	withCThread(func() {
		var option C.SidereonCycleSlipOptions
		var optionPointer *C.SidereonCycleSlipOptions
		if options != nil {
			option = C.SidereonCycleSlipOptions{gf_threshold_m: C.double(options.GFThresholdM), mw_threshold_cycles: C.double(options.MWThresholdCycles), min_arc_gap_s: C.double(options.MinArcGapS)}
			optionPointer = &option
		}
		var epochPointer *C.SidereonArcEpoch
		if len(epochs) != 0 {
			epochPointer = &epochs[0]
		}
		if _, err := checkedNativeAllocationSize(len(input), unsafe.Sizeof(C.SidereonIonoFreeSmoothResult{})); err != nil {
			operationErr = err
			return
		}
		out = make([]C.SidereonIonoFreeSmoothResult, len(input))
		var output *C.SidereonIonoFreeSmoothResult
		if len(out) != 0 {
			output = &out[0]
		}
		var written, required C.size_t
		status := C.sidereon_smooth_iono_free_code(epochPointer, C.size_t(len(input)), optionPointer, C.size_t(hatchWindowCap), nil, 0, &written, &required)
		if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
			return
		}
		count, queryErr := validateNativeQuery("ionosphere-free smoothed-code results", uint64(written), uint64(required))
		if queryErr != nil || count != len(input) {
			if queryErr != nil {
				operationErr = queryErr
			} else {
				operationErr = fmt.Errorf("sidereon: native ionosphere-free result count %d does not match input count %d", count, len(input))
			}
			return
		}
		written, required = 0, 0
		status = C.sidereon_smooth_iono_free_code(epochPointer, C.size_t(len(input)), optionPointer, C.size_t(hatchWindowCap), output, C.size_t(len(out)), &written, &required)
		if operationErr = statusErrorLocked(uint32(status)); operationErr != nil {
			return
		}
		var countErr error
		count, countErr = validateTwoPassCounts("ionosphere-free smoothed-code results", len(out), len(input), uint64(written), uint64(required))
		if countErr != nil {
			operationErr = countErr
			return
		}
		out = out[:count]
	})
	runtime.KeepAlive(input)
	if operationErr != nil {
		return nil, operationErr
	}
	result := make([]IonoFreeSmoothResult, len(out))
	for i, value := range out {
		result[i] = IonoFreeSmoothResult{PSmoothM: float64(value.p_smooth_m), PIFM: float64(value.p_if_m), LIFM: float64(value.l_if_m), Window: int(value.window), Reset: bool(value.reset)}
	}
	return result, nil
}

type PseudorangeObservation struct {
	SatelliteID  string
	PseudorangeM float64
}

type IonoFreeOverride struct {
	System byte
	Band1  string
	Band2  string
}

type IonoFreeCombined struct {
	SatelliteID  string
	PseudorangeM float64
}

type IonoFreeDropped struct {
	SatelliteID string
	Reason      uint32
}

func fixedCStringArray(value []C.char) string {
	result := make([]byte, 0, len(value))
	for _, item := range value {
		if item == 0 {
			break
		}
		result = append(result, byte(item))
	}
	return string(result)
}

type IonoFreePseudoranges struct {
	_      noCopy
	handle *surfaceHandle
}

func newIonoFreePseudoranges(pointer *C.SidereonIonoFreePseudoranges) (*IonoFreePseudoranges, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle, err := newSurfaceHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_iono_free_pseudoranges_free((*C.SidereonIonoFreePseudoranges)(value))
	})
	if err != nil {
		return nil, err
	}
	return &IonoFreePseudoranges{handle: handle}, nil
}

func makePseudorangeObservations(values []PseudorangeObservation) (unsafe.Pointer, []*C.char, error) {
	bytes, err := checkedNativeAllocationSize(len(values), unsafe.Sizeof(C.SidereonPseudorangeObservation{}))
	if err != nil {
		return nil, nil, err
	}
	if len(values) == 0 {
		return nil, nil, nil
	}
	memory := C.malloc(C.size_t(bytes))
	if memory == nil {
		return nil, nil, errors.New("sidereon: unable to allocate native pseudorange observations")
	}
	result := unsafe.Slice((*C.SidereonPseudorangeObservation)(memory), len(values))
	strings := make([]*C.char, 0, len(values))
	for i, value := range values {
		pointer, err := copyNativeCString(value.SatelliteID, "pseudorange satellite ID")
		if err != nil {
			C.free(memory)
			freeCStrings(strings)
			return nil, nil, err
		}
		strings = append(strings, (*C.char)(pointer))
		result[i] = C.SidereonPseudorangeObservation{sat_id: (*C.char)(pointer), pseudorange_m: C.double(value.PseudorangeM)}
	}
	return memory, strings, nil
}

func CombineIonosphereFreePseudoranges(band1, band2 []PseudorangeObservation, overrides []IonoFreeOverride) (*IonoFreePseudoranges, error) {
	firstMemory, firstStrings, err := makePseudorangeObservations(band1)
	if err != nil {
		return nil, err
	}
	defer freeCStrings(firstStrings)
	defer C.free(firstMemory)
	secondMemory, secondStrings, err := makePseudorangeObservations(band2)
	if err != nil {
		return nil, err
	}
	defer freeCStrings(secondStrings)
	defer C.free(secondMemory)
	overrideBytes, err := checkedNativeAllocationSize(len(overrides), unsafe.Sizeof(C.SidereonIonoFreeOverride{}))
	if err != nil {
		return nil, err
	}
	var overrideMemory unsafe.Pointer
	if len(overrides) != 0 {
		overrideMemory = C.malloc(C.size_t(overrideBytes))
		if overrideMemory == nil {
			return nil, errors.New("sidereon: unable to allocate native ionosphere-free overrides")
		}
	}
	defer C.free(overrideMemory)
	cOverrides := unsafe.Slice((*C.SidereonIonoFreeOverride)(overrideMemory), len(overrides))
	var overrideStrings []*C.char
	defer freeCStrings(overrideStrings)
	for i, value := range overrides {
		if value.System == 0 {
			return nil, invalidArgument("ionosphere-free override system must not be NUL")
		}
		band1Pointer, err := copyNativeCString(value.Band1, "ionosphere-free band 1")
		if err != nil {
			return nil, err
		}
		band2Pointer, err := copyNativeCString(value.Band2, "ionosphere-free band 2")
		if err != nil {
			C.free(band1Pointer)
			return nil, err
		}
		overrideStrings = append(overrideStrings, (*C.char)(band1Pointer), (*C.char)(band2Pointer))
		cOverrides[i] = C.SidereonIonoFreeOverride{system: C.char(value.System), band1: (*C.char)(band1Pointer), band2: (*C.char)(band2Pointer)}
	}
	firstPointer := (*C.SidereonPseudorangeObservation)(firstMemory)
	secondPointer := (*C.SidereonPseudorangeObservation)(secondMemory)
	overridePointer := (*C.SidereonIonoFreeOverride)(overrideMemory)
	var pointer *C.SidereonIonoFreePseudoranges
	var operationErr error
	withCThread(func() {
		status := C.sidereon_combination_ionosphere_free_pseudoranges(firstPointer, C.size_t(len(band1)), secondPointer, C.size_t(len(band2)), overridePointer, C.size_t(len(cOverrides)), &pointer)
		operationErr = statusErrorLocked(uint32(status))
		if operationErr != nil && pointer != nil {
			C.sidereon_iono_free_pseudoranges_free(pointer)
			pointer = nil
		}
	})
	if operationErr != nil {
		return nil, operationErr
	}
	return newIonoFreePseudoranges(pointer)
}

func (r *IonoFreePseudoranges) Close() error {
	if r == nil {
		return nil
	}
	return r.handle.close()
}

func copyIonoFreeRows(r *IonoFreePseudoranges, dropped bool) (combined []IonoFreeCombined, droppedRows []IonoFreeDropped, err error) {
	if r == nil || r.handle == nil {
		return nil, nil, ErrClosed
	}
	readErr := r.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var written, required C.size_t
			if dropped {
				status := C.sidereon_iono_free_pseudoranges_dropped((*C.SidereonIonoFreePseudoranges)(pointer), nil, 0, &written, &required)
				if err = statusErrorLocked(uint32(status)); err != nil {
					return
				}
				count, e := validateNativeQuery("ionosphere-free dropped rows", uint64(written), uint64(required))
				if e != nil {
					err = e
					return
				}
				rows := make([]C.SidereonIonoFreeDropped, count)
				var output *C.SidereonIonoFreeDropped
				if count != 0 {
					output = &rows[0]
				}
				written, required = 0, 0
				status = C.sidereon_iono_free_pseudoranges_dropped((*C.SidereonIonoFreePseudoranges)(pointer), output, C.size_t(count), &written, &required)
				if err = statusErrorLocked(uint32(status)); err != nil {
					return
				}
				w, e := validateTwoPassCounts("ionosphere-free dropped rows", count, count, uint64(written), uint64(required))
				if e != nil {
					err = e
					return
				}
				droppedRows = make([]IonoFreeDropped, w)
				for i := range droppedRows {
					droppedRows[i] = IonoFreeDropped{SatelliteID: fixedCStringArray(rows[i].sat_id[:]), Reason: uint32(rows[i].reason)}
				}
				return
			}
			status := C.sidereon_iono_free_pseudoranges_combined((*C.SidereonIonoFreePseudoranges)(pointer), nil, 0, &written, &required)
			if err = statusErrorLocked(uint32(status)); err != nil {
				return
			}
			count, e := validateNativeQuery("ionosphere-free combined rows", uint64(written), uint64(required))
			if e != nil {
				err = e
				return
			}
			rows := make([]C.SidereonIonoFreeCombined, count)
			var output *C.SidereonIonoFreeCombined
			if count != 0 {
				output = &rows[0]
			}
			written, required = 0, 0
			status = C.sidereon_iono_free_pseudoranges_combined((*C.SidereonIonoFreePseudoranges)(pointer), output, C.size_t(count), &written, &required)
			if err = statusErrorLocked(uint32(status)); err != nil {
				return
			}
			w, e := validateTwoPassCounts("ionosphere-free combined rows", count, count, uint64(written), uint64(required))
			if e != nil {
				err = e
				return
			}
			combined = make([]IonoFreeCombined, w)
			for i := range combined {
				combined[i] = IonoFreeCombined{SatelliteID: fixedCStringArray(rows[i].sat_id[:]), PseudorangeM: float64(rows[i].pseudorange_m)}
			}
		})
		return err
	})
	if readErr != nil {
		return nil, nil, readErr
	}
	return combined, droppedRows, err
}

func (r *IonoFreePseudoranges) Combined() ([]IonoFreeCombined, error) {
	combined, _, err := copyIonoFreeRows(r, false)
	return combined, err
}

func (r *IonoFreePseudoranges) Dropped() ([]IonoFreeDropped, error) {
	_, dropped, err := copyIonoFreeRows(r, true)
	return dropped, err
}
