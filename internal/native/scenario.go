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
	"sort"
	"unsafe"
)

// ScenarioObservation is one copied synthetic observation row.
type ScenarioObservation struct {
	EpochIndex         int
	SatelliteID        string
	CodeObservable     string
	PhaseObservable    string
	DopplerObservable  string
	CarrierHz          float64
	PseudorangeM       float64
	CarrierPhaseCycles float64
	DopplerHz          float64
}

// ScenarioReceiverTruth is one copied receiver truth row.
type ScenarioReceiverTruth struct {
	TRxJ2000S       float64
	PositionECEFM   [3]float64
	VelocityECEFMPS [3]float64
	ClockM          float64
	ClockRateMPS    float64
}

// ScenarioSummary is a copied deterministic simulation summary.
type ScenarioSummary struct {
	SchemaVersion          uint32
	Seed                   uint64
	ReceiverTruthCount     int
	ObservationCount       int
	EpochOffsetCount       int
	DeterminismFingerprint uint64
	JSONLength             int
}

// ScenarioTerms is the copied ground-truth ledger for one synthetic
// observation.
type ScenarioTerms struct {
	GeometricRangeM                       float64
	SatelliteClockM                       float64
	ReceiverClockM                        float64
	SatelliteClockErrorM                  float64
	IonosphereM                           float64
	TroposphereM                          float64
	ThermalNoiseM                         float64
	MultipathM                            float64
	QuantizationM                         float64
	CarrierPhaseGeometricCycles           float64
	CarrierPhaseReceiverClockCycles       float64
	CarrierPhaseSatelliteClockCycles      float64
	CarrierPhaseSatelliteClockErrorCycles float64
	CarrierPhaseIonosphereCycles          float64
	CarrierPhaseTroposphereCycles         float64
	CarrierPhaseThermalNoiseCycles        float64
	CarrierPhaseBiasCycles                float64
	CarrierPhaseQuantizationCycles        float64
	DopplerSatelliteMotionHz              float64
	DopplerReceiverMotionHz               float64
	DopplerSatelliteClockHz               float64
	DopplerReceiverClockHz                float64
	DopplerSatelliteClockErrorHz          float64
	DopplerThermalNoiseHz                 float64
	DopplerQuantizationHz                 float64
}

// ScenarioSimulation owns a C-generated deterministic simulation and must not
// be copied after first use. Queries may run concurrently with Close.
type ScenarioSimulation struct {
	_      noCopy
	handle *positioningHandle
}

func releaseScenarioSimulation(pointer unsafe.Pointer) {
	C.sidereon_scenario_simulation_free((*C.SidereonScenarioSimulation)(pointer))
}

type scenarioInput struct {
	positioning *positioningHandle
	resource    *resource
}

type scenarioTarget struct {
	positioning *positioningHandle
	resource    *resource
	address     uintptr
}

// withScenarioInputs acquires mixed product handles in one canonical address
// order. Positioning handles are deduplicated and retain their normal outer
// handle lock; direct protocol resources (broadcast ephemeris) are deduped by
// resource address. Pointers are returned in caller order.
func withScenarioInputs(inputs []scenarioInput, fn func([]unsafe.Pointer) error) error {
	targets := make([]scenarioTarget, 0, len(inputs))
	for _, input := range inputs {
		if input.positioning == nil && input.resource == nil {
			return ErrClosed
		}
		if input.positioning != nil {
			targets = append(targets, scenarioTarget{positioning: input.positioning, address: uintptr(unsafe.Pointer(input.positioning))})
		} else {
			targets = append(targets, scenarioTarget{resource: input.resource, address: uintptr(unsafe.Pointer(input.resource))})
		}
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].address < targets[j].address })
	unique := targets[:0]
	for _, target := range targets {
		duplicate := false
		for _, prior := range unique {
			if target.positioning != nil && target.positioning == prior.positioning || target.resource != nil && target.resource == prior.resource {
				duplicate = true
				break
			}
		}
		if !duplicate {
			unique = append(unique, target)
		}
	}
	locked := make([]scenarioTarget, 0, len(unique))
	unlock := func() {
		for i := len(locked) - 1; i >= 0; i-- {
			if locked[i].resource != nil {
				locked[i].resource.mu.RUnlock()
			}
			if locked[i].positioning != nil {
				locked[i].positioning.mu.RUnlock()
			}
		}
	}
	for _, target := range unique {
		if target.positioning != nil {
			target.positioning.mu.RLock()
			target.resource = target.positioning.resource
			if target.resource == nil {
				target.positioning.mu.RUnlock()
				unlock()
				return ErrClosed
			}
		}
		if target.resource == nil {
			if target.positioning != nil {
				target.positioning.mu.RUnlock()
			}
			unlock()
			return ErrClosed
		}
		target.resource.mu.RLock()
		if target.resource.ptr == nil {
			target.resource.mu.RUnlock()
			if target.positioning != nil {
				target.positioning.mu.RUnlock()
			}
			unlock()
			return ErrClosed
		}
		locked = append(locked, target)
	}
	byPositioning := make(map[*positioningHandle]unsafe.Pointer, len(locked))
	byResource := make(map[*resource]unsafe.Pointer, len(locked))
	for _, target := range locked {
		if target.positioning != nil {
			byPositioning[target.positioning] = target.resource.ptr
		} else {
			byResource[target.resource] = target.resource.ptr
		}
	}
	pointers := make([]unsafe.Pointer, len(inputs))
	for i, input := range inputs {
		if input.positioning != nil {
			pointers[i] = byPositioning[input.positioning]
		} else {
			pointers[i] = byResource[input.resource]
		}
	}
	defer unlock()
	return fn(pointers)
}

func scenarioSimulate(data []byte, inputs []scenarioInput, kind uint32) (*ScenarioSimulation, error) {
	length, err := checkedNativeSize(len(data))
	if err != nil {
		return nil, err
	}
	var out *C.SidereonScenarioSimulation
	err = withScenarioInputs(inputs, func(pointers []unsafe.Pointer) error {
		return withCThreadError(func() error {
			memory, copyErr := copyNativeInput(data)
			if copyErr != nil {
				return copyErr
			}
			if memory != nil {
				defer C.free(memory)
			}
			status := C.enum_SidereonStatus(C.SIDEREON_STATUS_OK)
			dataPointer := (*C.uint8_t)(memory)
			switch kind {
			case 0:
				status = C.sidereon_scenario_simulate_json(dataPointer, length, &out)
			case 1:
				status = C.sidereon_scenario_simulate_json_with_broadcast(dataPointer, length, (*C.SidereonBroadcastEphemeris)(pointers[0]), &out)
			case 2:
				status = C.sidereon_scenario_simulate_json_with_broadcast_and_ionex(dataPointer, length, (*C.SidereonBroadcastEphemeris)(pointers[0]), (*C.SidereonIonex)(pointers[1]), &out)
			case 3:
				status = C.sidereon_scenario_simulate_json_with_ionex(dataPointer, length, (*C.SidereonIonex)(pointers[0]), &out)
			case 4:
				status = C.sidereon_scenario_simulate_json_with_sp3(dataPointer, length, (*C.SidereonSp3)(pointers[0]), &out)
			case 5:
				status = C.sidereon_scenario_simulate_json_with_sp3_and_ionex(dataPointer, length, (*C.SidereonSp3)(pointers[0]), (*C.SidereonIonex)(pointers[1]), &out)
			default:
				return invalidArgument("scenario simulation route is not defined")
			}
			err := statusErrorLocked(uint32(status))
			if err != nil && out != nil {
				C.sidereon_scenario_simulation_free(out)
				out = nil
			}
			return err
		})
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_scenario_simulation_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, errors.New("sidereon: native scenario simulation returned no handle")
	}
	return &ScenarioSimulation{handle: newPositioningHandle(unsafe.Pointer(out), releaseScenarioSimulation)}, nil
}

func ScenarioSimulateJSON(data []byte) (*ScenarioSimulation, error) {
	return scenarioSimulate(data, nil, 0)
}

func ScenarioSimulateJSONWithBroadcast(data []byte, broadcast *BroadcastEphemeris) (*ScenarioSimulation, error) {
	if broadcast == nil {
		return nil, ErrClosed
	}
	return scenarioSimulate(data, []scenarioInput{{resource: broadcast.resource}}, 1)
}

func ScenarioSimulateJSONWithBroadcastAndIonex(data []byte, broadcast *BroadcastEphemeris, ionex *Ionex) (*ScenarioSimulation, error) {
	if broadcast == nil || ionex == nil || ionex.handle == nil {
		return nil, ErrClosed
	}
	return scenarioSimulate(data, []scenarioInput{{resource: broadcast.resource}, {positioning: ionex.handle}}, 2)
}

func ScenarioSimulateJSONWithIonex(data []byte, ionex *Ionex) (*ScenarioSimulation, error) {
	if ionex == nil || ionex.handle == nil {
		return nil, ErrClosed
	}
	return scenarioSimulate(data, []scenarioInput{{positioning: ionex.handle}}, 3)
}

func ScenarioSimulateJSONWithSP3(data []byte, sp3 *SP3) (*ScenarioSimulation, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	return scenarioSimulate(data, []scenarioInput{{positioning: sp3.handle}}, 4)
}

func ScenarioSimulateJSONWithSP3AndIonex(data []byte, sp3 *SP3, ionex *Ionex) (*ScenarioSimulation, error) {
	if sp3 == nil || ionex == nil || sp3.handle == nil || ionex.handle == nil {
		return nil, ErrClosed
	}
	return scenarioSimulate(data, []scenarioInput{{positioning: sp3.handle}, {positioning: ionex.handle}}, 5)
}

func (s *ScenarioSimulation) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func copyScenarioEpochOffsets(s *ScenarioSimulation) ([]int, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.size_t
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(out *C.size_t, length C.size_t) C.enum_SidereonStatus {
				return C.sidereon_scenario_epoch_offsets((*C.SidereonScenarioSimulation)(pointer), out, length, &written, &required)
			}
			if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
				return err
			}
			n, err := validateNativeQuery("scenario epoch offsets", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err := allocNativeArray(n, unsafe.Sizeof(C.size_t(0)))
			if err != nil {
				return err
			}
			if memory != nil {
				defer C.free(memory)
			}
			written, required = 0, 0
			var out *C.size_t
			if n > 0 {
				out = (*C.size_t)(memory)
			}
			if err := statusErrorLocked(uint32(call(out, C.size_t(n)))); err != nil {
				return err
			}
			count, err := validateTwoPassCounts("scenario epoch offsets", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if count > 0 {
				raw = append(raw, unsafe.Slice((*C.size_t)(memory), count)...)
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	out := make([]int, len(raw))
	for i, value := range raw {
		out[i], err = sizeTToInt(value, "scenario epoch offset")
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *ScenarioSimulation) EpochOffsets() ([]int, error) {
	return copyScenarioEpochOffsets(s)
}

func scenarioString(pointer *C.char, length int) (string, error) {
	if length <= 0 {
		return "", errors.New("sidereon: invalid native scenario string length")
	}
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(pointer)), length)
	for i, value := range bytes {
		if value == 0 {
			return string(bytes[:i]), nil
		}
	}
	return "", errors.New("sidereon: native scenario string is not terminated")
}

func scenarioObservationFromC(value C.SidereonScenarioObservation) (ScenarioObservation, error) {
	code, err := scenarioString(&value.code_observable[0], len(value.code_observable))
	if err != nil {
		return ScenarioObservation{}, err
	}
	phase, err := scenarioString(&value.phase_observable[0], len(value.phase_observable))
	if err != nil {
		return ScenarioObservation{}, err
	}
	doppler, err := scenarioString(&value.doppler_observable[0], len(value.doppler_observable))
	if err != nil {
		return ScenarioObservation{}, err
	}
	epoch, err := sizeTToInt(value.epoch_index, "scenario observation epoch index")
	if err != nil {
		return ScenarioObservation{}, err
	}
	return ScenarioObservation{EpochIndex: epoch, SatelliteID: tokenFromC(value.sat_id), CodeObservable: code, PhaseObservable: phase, DopplerObservable: doppler, CarrierHz: float64(value.carrier_hz), PseudorangeM: float64(value.pseudorange_m), CarrierPhaseCycles: float64(value.carrier_phase_cycles), DopplerHz: float64(value.doppler_hz)}, nil
}

func (s *ScenarioSimulation) Observations() ([]ScenarioObservation, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonScenarioObservation
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(out *C.SidereonScenarioObservation, length C.size_t) C.enum_SidereonStatus {
				return C.sidereon_scenario_observations((*C.SidereonScenarioSimulation)(pointer), out, length, &written, &required)
			}
			if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
				return err
			}
			n, err := validateNativeQuery("scenario observations", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err := allocNativeArray(n, unsafe.Sizeof(C.SidereonScenarioObservation{}))
			if err != nil {
				return err
			}
			if memory != nil {
				defer C.free(memory)
			}
			written, required = 0, 0
			var out *C.SidereonScenarioObservation
			if n > 0 {
				out = (*C.SidereonScenarioObservation)(memory)
			}
			if err := statusErrorLocked(uint32(call(out, C.size_t(n)))); err != nil {
				return err
			}
			count, err := validateTwoPassCounts("scenario observations", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if count > 0 {
				raw = append(raw, unsafe.Slice((*C.SidereonScenarioObservation)(memory), count)...)
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	out := make([]ScenarioObservation, len(raw))
	for i, value := range raw {
		out[i], err = scenarioObservationFromC(value)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *ScenarioSimulation) ReceiverTruth() ([]ScenarioReceiverTruth, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonScenarioReceiverTruth
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(out *C.SidereonScenarioReceiverTruth, length C.size_t) C.enum_SidereonStatus {
				return C.sidereon_scenario_receiver_truth((*C.SidereonScenarioSimulation)(pointer), out, length, &written, &required)
			}
			if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
				return err
			}
			n, err := validateNativeQuery("scenario receiver truth", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err := allocNativeArray(n, unsafe.Sizeof(C.SidereonScenarioReceiverTruth{}))
			if err != nil {
				return err
			}
			if memory != nil {
				defer C.free(memory)
			}
			written, required = 0, 0
			var out *C.SidereonScenarioReceiverTruth
			if n > 0 {
				out = (*C.SidereonScenarioReceiverTruth)(memory)
			}
			if err := statusErrorLocked(uint32(call(out, C.size_t(n)))); err != nil {
				return err
			}
			count, err := validateTwoPassCounts("scenario receiver truth", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if count > 0 {
				raw = append(raw, unsafe.Slice((*C.SidereonScenarioReceiverTruth)(memory), count)...)
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	out := make([]ScenarioReceiverTruth, len(raw))
	for i, value := range raw {
		out[i] = ScenarioReceiverTruth{TRxJ2000S: float64(value.t_rx_j2000_s), ClockM: float64(value.clock_m), ClockRateMPS: float64(value.clock_rate_m_s)}
		for j := range out[i].PositionECEFM {
			out[i].PositionECEFM[j] = float64(value.position_ecef_m[j])
			out[i].VelocityECEFMPS[j] = float64(value.velocity_ecef_m_s[j])
		}
	}
	return out, nil
}

func (s *ScenarioSimulation) Summary() (ScenarioSummary, error) {
	if s == nil || s.handle == nil {
		return ScenarioSummary{}, ErrClosed
	}
	var value C.SidereonScenarioSummary
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			return statusErrorLocked(uint32(C.sidereon_scenario_simulation_summary((*C.SidereonScenarioSimulation)(pointer), &value)))
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return ScenarioSummary{}, err
	}
	receiverCount, err := sizeTToInt(value.receiver_truth_count, "scenario receiver truth count")
	if err != nil {
		return ScenarioSummary{}, err
	}
	observationCount, err := sizeTToInt(value.observation_count, "scenario observation count")
	if err != nil {
		return ScenarioSummary{}, err
	}
	offsetCount, err := sizeTToInt(value.epoch_offset_count, "scenario epoch offset count")
	if err != nil {
		return ScenarioSummary{}, err
	}
	jsonLength, err := sizeTToInt(value.json_len, "scenario JSON length")
	if err != nil {
		return ScenarioSummary{}, err
	}
	return ScenarioSummary{SchemaVersion: uint32(value.schema_version), Seed: uint64(value.seed), ReceiverTruthCount: receiverCount, ObservationCount: observationCount, EpochOffsetCount: offsetCount, DeterminismFingerprint: uint64(value.determinism_fingerprint), JSONLength: jsonLength}, nil
}

func (s *ScenarioSimulation) JSON() ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var output []byte
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var err error
			output, err = copyNativeBytesLocked("scenario JSON", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_scenario_simulation_json((*C.SidereonScenarioSimulation)(pointer), out, length, written, required)
			})
			return err
		})
	})
	runtime.KeepAlive(s)
	return output, err
}

func (s *ScenarioSimulation) Terms() ([]ScenarioTerms, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var raw []C.SidereonScenarioTerms
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			call := func(out *C.SidereonScenarioTerms, length C.size_t) C.enum_SidereonStatus {
				return C.sidereon_scenario_terms((*C.SidereonScenarioSimulation)(pointer), out, length, &written, &required)
			}
			if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
				return err
			}
			n, err := validateNativeQuery("scenario terms", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			memory, err := allocNativeArray(n, unsafe.Sizeof(C.SidereonScenarioTerms{}))
			if err != nil {
				return err
			}
			if memory != nil {
				defer C.free(memory)
			}
			written, required = 0, 0
			var out *C.SidereonScenarioTerms
			if n > 0 {
				out = (*C.SidereonScenarioTerms)(memory)
			}
			if err := statusErrorLocked(uint32(call(out, C.size_t(n)))); err != nil {
				return err
			}
			count, err := validateTwoPassCounts("scenario terms", n, n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if count > 0 {
				raw = append(raw, unsafe.Slice((*C.SidereonScenarioTerms)(memory), count)...)
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	out := make([]ScenarioTerms, len(raw))
	for i, value := range raw {
		out[i] = ScenarioTerms{
			GeometricRangeM: float64(value.geometric_range_m), SatelliteClockM: float64(value.satellite_clock_m), ReceiverClockM: float64(value.receiver_clock_m), SatelliteClockErrorM: float64(value.satellite_clock_error_m),
			IonosphereM: float64(value.ionosphere_m), TroposphereM: float64(value.troposphere_m), ThermalNoiseM: float64(value.thermal_noise_m), MultipathM: float64(value.multipath_m), QuantizationM: float64(value.quantization_m),
			CarrierPhaseGeometricCycles: float64(value.carrier_phase_geometric_cycles), CarrierPhaseReceiverClockCycles: float64(value.carrier_phase_receiver_clock_cycles), CarrierPhaseSatelliteClockCycles: float64(value.carrier_phase_satellite_clock_cycles), CarrierPhaseSatelliteClockErrorCycles: float64(value.carrier_phase_satellite_clock_error_cycles),
			CarrierPhaseIonosphereCycles: float64(value.carrier_phase_ionosphere_cycles), CarrierPhaseTroposphereCycles: float64(value.carrier_phase_troposphere_cycles), CarrierPhaseThermalNoiseCycles: float64(value.carrier_phase_thermal_noise_cycles), CarrierPhaseBiasCycles: float64(value.carrier_phase_bias_cycles), CarrierPhaseQuantizationCycles: float64(value.carrier_phase_quantization_cycles),
			DopplerSatelliteMotionHz: float64(value.doppler_satellite_motion_hz), DopplerReceiverMotionHz: float64(value.doppler_receiver_motion_hz), DopplerSatelliteClockHz: float64(value.doppler_satellite_clock_hz), DopplerReceiverClockHz: float64(value.doppler_receiver_clock_hz), DopplerSatelliteClockErrorHz: float64(value.doppler_satellite_clock_error_hz), DopplerThermalNoiseHz: float64(value.doppler_thermal_noise_hz), DopplerQuantizationHz: float64(value.doppler_quantization_hz),
		}
	}
	return out, nil
}
