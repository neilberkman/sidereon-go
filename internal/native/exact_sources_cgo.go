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

type EphemerisSourceState struct {
	HasState      bool
	PositionECEFM [3]float64
	ClockS        float64
	HasGroupDelay bool
	GroupDelayS   float64
	Degraded      bool
	DegradeReason uint32
}

type TransmitEpochClock struct {
	HasClock      bool
	ClockS        float64
	Degraded      bool
	DegradeReason uint32
}

type ClockRelativity struct {
	Kind  uint32
	TermS float64
}

type preciseStateQueryCall func(unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.SidereonExactEpochQuery, *C.SidereonEphemerisSourceState) uint32
type preciseClockQueryCall func(unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.SidereonExactEpochQuery, *C.bool, *C.double, *C.bool) uint32
type preciseRelativityQueryCall func(unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.double, *uint32, *C.double) uint32
type preciseVarianceQueryCall func(unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.SidereonExactEpochQuery, *C.double) uint32

func sourceQueryStatusLocked(status uint32) (uint32, error) {
	return sourceQueryStatusWithErrorLocked(status, statusErrorLocked)
}

func rtcmSourceQueryStatusLocked(status uint32) (uint32, error) {
	return sourceQueryStatusWithErrorLocked(status, rtcmErrorLocked)
}

func sourceQueryStatusWithErrorLocked(status uint32, captureError func(uint32) error) (uint32, error) {
	if err := captureError(status); err != nil {
		return 0, err
	}
	var reason C.enum_SidereonDegradeReason
	if getterStatus := C.sidereon_last_degrade_reason(&reason); getterStatus != C.SIDEREON_STATUS_OK {
		return 0, statusErrorLocked(uint32(getterStatus))
	}
	value := uint32(reason)
	if value > 2 {
		return 0, invalidArgument("invalid source-query degradation reason")
	}
	return value, nil
}

func exactSourceCall(fn func() uint32) (uint32, error) {
	var reason uint32
	var operationErr error
	withCThread(func() { reason, operationErr = sourceQueryStatusLocked(fn()) })
	return reason, operationErr
}

func ephemerisSourceStateFromC(value C.SidereonEphemerisSourceState) EphemerisSourceState {
	return EphemerisSourceState{
		HasState: bool(value.has_state),
		PositionECEFM: [3]float64{
			float64(value.position_ecef_m[0]), float64(value.position_ecef_m[1]), float64(value.position_ecef_m[2]),
		},
		ClockS: float64(value.clock_s), HasGroupDelay: bool(value.has_group_delay),
		GroupDelayS: float64(value.group_delay_s), Degraded: bool(value.degraded),
	}
}

func withExactSourceQueries(handle *positioningHandle, stateEpoch, selectionEpoch *ExactEpochQuery, satellite string, call func(unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.SidereonExactEpochQuery) error) error {
	if handle == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return err
	}
	if len(satellite) >= 16 {
		return errTokenTooLong
	}
	satelliteCString := C.CString(satellite)
	if satelliteCString == nil {
		return missingNativeHandle("satellite token")
	}
	defer C.free(unsafe.Pointer(satelliteCString))
	err := withPositioningHandles([]*positioningHandle{handle, stateEpoch.handle, selectionEpoch.handle}, func(pointers []unsafe.Pointer) error {
		return call(pointers[0], satelliteCString, (*C.SidereonExactEpochQuery)(pointers[1]), (*C.SidereonExactEpochQuery)(pointers[2]))
	})
	runtime.KeepAlive(stateEpoch)
	runtime.KeepAlive(selectionEpoch)
	return err
}

func preciseStateQuery(handle *positioningHandle, stateEpoch, selectionEpoch *ExactEpochQuery, satellite string, invoke preciseStateQueryCall) (EphemerisSourceState, error) {
	var output C.SidereonEphemerisSourceState
	var degradeReason C.enum_SidereonDegradeReason
	err := withExactSourceQueries(handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery) error {
		var operationErr error
		withCThread(func() {
			var reason uint32
			reason, operationErr = sourceQueryStatusLocked(invoke(source, sat, state, selection, &output))
			degradeReason = C.enum_SidereonDegradeReason(reason)
		})
		return operationErr
	})
	if err != nil {
		return EphemerisSourceState{}, err
	}
	result := ephemerisSourceStateFromC(output)
	result.DegradeReason = uint32(degradeReason)
	if result.DegradeReason > 2 {
		return EphemerisSourceState{}, invalidArgument("invalid source-state degradation reason")
	}
	return result, nil
}

func preciseClockQuery(handle *positioningHandle, stateEpoch, selectionEpoch *ExactEpochQuery, satellite string, invoke preciseClockQueryCall) (TransmitEpochClock, error) {
	var hasClock, degraded C.bool
	var clock C.double
	var degradeReason C.enum_SidereonDegradeReason
	err := withExactSourceQueries(handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery) error {
		var operationErr error
		withCThread(func() {
			var reason uint32
			reason, operationErr = sourceQueryStatusLocked(invoke(source, sat, state, selection, &hasClock, &clock, &degraded))
			degradeReason = C.enum_SidereonDegradeReason(reason)
		})
		return operationErr
	})
	if err != nil {
		return TransmitEpochClock{}, err
	}
	if degradeReason > 2 {
		return TransmitEpochClock{}, invalidArgument("invalid transmit-clock degradation reason")
	}
	return TransmitEpochClock{HasClock: bool(hasClock), ClockS: float64(clock), Degraded: bool(degraded), DegradeReason: uint32(degradeReason)}, nil
}

func preciseRelativityQuery(handle *positioningHandle, epoch *ExactEpochQuery, satellite string, position [3]float64, invoke preciseRelativityQueryCall) (ClockRelativity, error) {
	if handle == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return ClockRelativity{}, err
	}
	if len(satellite) >= 16 {
		return ClockRelativity{}, errTokenTooLong
	}
	satelliteCString := C.CString(satellite)
	if satelliteCString == nil {
		return ClockRelativity{}, missingNativeHandle("satellite token")
	}
	defer C.free(unsafe.Pointer(satelliteCString))
	coordinates := [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])}
	var kind uint32
	var term C.double
	err := withPositioningHandles([]*positioningHandle{handle, epoch.handle}, func(pointers []unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return invoke(pointers[0], satelliteCString, (*C.SidereonExactEpochQuery)(pointers[1]), &coordinates[0], &kind, &term)
		})
	})
	runtime.KeepAlive(epoch)
	if err != nil {
		return ClockRelativity{}, err
	}
	return ClockRelativity{Kind: uint32(kind), TermS: float64(term)}, nil
}

func preciseVarianceQuery(handle *positioningHandle, stateEpoch, selectionEpoch *ExactEpochQuery, satellite string, invoke preciseVarianceQueryCall) (float64, error) {
	var variance C.double
	err := withExactSourceQueries(handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery) error {
		return callStatus(func() uint32 { return invoke(source, sat, state, selection, &variance) })
	})
	return float64(variance), err
}

func broadcastSourceState(b *BroadcastEphemeris, stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	if b == nil || b.resource == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return EphemerisSourceState{}, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return EphemerisSourceState{}, err
	}
	if len(satellite) >= 16 {
		return EphemerisSourceState{}, errTokenTooLong
	}
	satelliteCString := C.CString(satellite)
	if satelliteCString == nil {
		return EphemerisSourceState{}, missingNativeHandle("satellite token")
	}
	defer C.free(unsafe.Pointer(satelliteCString))
	var output C.SidereonEphemerisSourceState
	var degradeReason uint32
	err := withPositioningHandles([]*positioningHandle{stateEpoch.handle, selectionEpoch.handle}, func(epochs []unsafe.Pointer) error {
		return b.resource.with(func(broadcastPointer unsafe.Pointer) error {
			var operationErr error
			withCThread(func() {
				degradeReason, operationErr = sourceQueryStatusLocked(uint32(C.sidereon_broadcast_state_at_epoch_queries(
					(*C.SidereonBroadcastEphemeris)(broadcastPointer), satelliteCString,
					(*C.SidereonExactEpochQuery)(epochs[0]), (*C.SidereonExactEpochQuery)(epochs[1]), &output,
				)))
			})
			return operationErr
		})
	})
	runtime.KeepAlive(b)
	runtime.KeepAlive(stateEpoch)
	runtime.KeepAlive(selectionEpoch)
	if err != nil {
		return EphemerisSourceState{}, err
	}
	result := ephemerisSourceStateFromC(output)
	result.DegradeReason = degradeReason
	return result, nil
}

func broadcastTransmitClock(b *BroadcastEphemeris, transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	if b == nil || b.resource == nil || transmitEpoch == nil || transmitEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return TransmitEpochClock{}, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return TransmitEpochClock{}, err
	}
	if len(satellite) >= 16 {
		return TransmitEpochClock{}, errTokenTooLong
	}
	satelliteCString := C.CString(satellite)
	if satelliteCString == nil {
		return TransmitEpochClock{}, missingNativeHandle("satellite token")
	}
	defer C.free(unsafe.Pointer(satelliteCString))
	var hasClock, degraded C.bool
	var clock C.double
	var degradeReason uint32
	err := withPositioningHandles([]*positioningHandle{transmitEpoch.handle, selectionEpoch.handle}, func(epochs []unsafe.Pointer) error {
		return b.resource.with(func(broadcastPointer unsafe.Pointer) error {
			var operationErr error
			withCThread(func() {
				degradeReason, operationErr = sourceQueryStatusLocked(uint32(C.sidereon_broadcast_transmit_epoch_clock_at_epoch_queries(
					(*C.SidereonBroadcastEphemeris)(broadcastPointer), satelliteCString,
					(*C.SidereonExactEpochQuery)(epochs[0]), (*C.SidereonExactEpochQuery)(epochs[1]),
					&hasClock, &clock, &degraded,
				)))
			})
			return operationErr
		})
	})
	runtime.KeepAlive(b)
	runtime.KeepAlive(transmitEpoch)
	runtime.KeepAlive(selectionEpoch)
	if err != nil {
		return TransmitEpochClock{}, err
	}
	return TransmitEpochClock{HasClock: bool(hasClock), ClockS: float64(clock), Degraded: bool(degraded), DegradeReason: degradeReason}, nil
}

func (b *BroadcastEphemeris) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	return broadcastSourceState(b, stateEpoch, selectionEpoch, satellite)
}

func (b *BroadcastEphemeris) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	return broadcastTransmitClock(b, transmitEpoch, selectionEpoch, satellite)
}

func (b *BroadcastEphemeris) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, position [3]float64) (ClockRelativity, error) {
	if b == nil || b.resource == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return ClockRelativity{}, err
	}
	if len(satellite) >= 16 {
		return ClockRelativity{}, errTokenTooLong
	}
	satelliteCString := C.CString(satellite)
	if satelliteCString == nil {
		return ClockRelativity{}, missingNativeHandle("satellite token")
	}
	defer C.free(unsafe.Pointer(satelliteCString))
	coordinates := [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])}
	var kind uint32
	var term C.double
	err := epoch.handle.with(func(epochPointer unsafe.Pointer) error {
		return b.resource.with(func(broadcastPointer unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_broadcast_clock_relativity_at_epoch_query(
					(*C.SidereonBroadcastEphemeris)(broadcastPointer), satelliteCString,
					(*C.SidereonExactEpochQuery)(epochPointer), &coordinates[0], &kind, &term,
				))
			})
		})
	})
	runtime.KeepAlive(b)
	runtime.KeepAlive(epoch)
	if err != nil {
		return ClockRelativity{}, err
	}
	return ClockRelativity{Kind: uint32(kind), TermS: float64(term)}, nil
}

func (b *BroadcastEphemeris) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	if b == nil || b.resource == nil || stateEpoch == nil || stateEpoch.handle == nil || selectionEpoch == nil || selectionEpoch.handle == nil {
		return 0, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return 0, err
	}
	if len(satellite) >= 16 {
		return 0, errTokenTooLong
	}
	satelliteCString := C.CString(satellite)
	if satelliteCString == nil {
		return 0, missingNativeHandle("satellite token")
	}
	defer C.free(unsafe.Pointer(satelliteCString))
	var variance C.double
	err := withPositioningHandles([]*positioningHandle{stateEpoch.handle, selectionEpoch.handle}, func(epochs []unsafe.Pointer) error {
		return b.resource.with(func(broadcastPointer unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_broadcast_ephemeris_variance_at_epoch_queries(
					(*C.SidereonBroadcastEphemeris)(broadcastPointer), satelliteCString,
					(*C.SidereonExactEpochQuery)(epochs[0]), (*C.SidereonExactEpochQuery)(epochs[1]), &variance,
				))
			})
		})
	})
	runtime.KeepAlive(b)
	runtime.KeepAlive(stateEpoch)
	runtime.KeepAlive(selectionEpoch)
	return float64(variance), err
}

func withCorrectionEpochQueries(broadcast *BroadcastEphemeris, store *SBASCorrectionStore, stateEpoch, selectionEpoch *ExactEpochQuery, satellite string, call func(unsafe.Pointer, unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.SidereonExactEpochQuery) error) error {
	if broadcast == nil || store == nil || stateEpoch == nil || selectionEpoch == nil || stateEpoch.handle == nil || selectionEpoch.handle == nil {
		return ErrClosed
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return err
	}
	return withResources([]*resource{broadcast.resource, store.resource}, func(resources []unsafe.Pointer) error {
		return withPositioningHandles([]*positioningHandle{stateEpoch.handle, selectionEpoch.handle}, func(epochs []unsafe.Pointer) error {
			return call(resources[0], resources[1], (*C.char)(unsafe.Pointer(&sat.bytes[0])), (*C.SidereonExactEpochQuery)(epochs[0]), (*C.SidereonExactEpochQuery)(epochs[1]))
		})
	})
}

func (s *SBASCorrectionStore) SourceStateAtEpochQueries(broadcast *BroadcastEphemeris, geo string, mode uint32, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery) (EphemerisSourceState, error) {
	if err := validateSBASSolveModeValue(mode); err != nil {
		return EphemerisSourceState{}, err
	}
	geoToken, err := tokenToC(geo)
	if err != nil {
		return EphemerisSourceState{}, err
	}
	var out C.SidereonEphemerisSourceState
	var degradeReason uint32
	err = withCorrectionEpochQueries(broadcast, s, stateEpoch, selectionEpoch, satellite, func(b, store unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery) error {
		var operationErr error
		withCThread(func() {
			degradeReason, operationErr = sourceQueryStatusLocked(uint32(C.sidereon_sbas_corrected_state_at_epoch_queries((*C.SidereonBroadcastEphemeris)(b), (*C.SidereonSbasCorrectionStore)(store), (*C.char)(unsafe.Pointer(&geoToken.bytes[0])), C.uint32_t(mode), sat, state, selection, &out)))
		})
		return operationErr
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(s)
	runtime.KeepAlive(stateEpoch)
	runtime.KeepAlive(selectionEpoch)
	if err != nil {
		return EphemerisSourceState{}, err
	}
	result := ephemerisSourceStateFromC(out)
	result.DegradeReason = degradeReason
	return result, nil
}

func (s *SBASCorrectionStore) TransmitEpochClockAtEpochQueries(broadcast *BroadcastEphemeris, geo string, mode uint32, satellite string, transmitEpoch, selectionEpoch *ExactEpochQuery) (TransmitEpochClock, error) {
	if err := validateSBASSolveModeValue(mode); err != nil {
		return TransmitEpochClock{}, err
	}
	geoToken, err := tokenToC(geo)
	if err != nil {
		return TransmitEpochClock{}, err
	}
	var hasClock, degraded C.bool
	var clock C.double
	var degradeReason uint32
	err = withCorrectionEpochQueries(broadcast, s, transmitEpoch, selectionEpoch, satellite, func(b, store unsafe.Pointer, sat *C.char, transmit, selection *C.SidereonExactEpochQuery) error {
		var operationErr error
		withCThread(func() {
			degradeReason, operationErr = sourceQueryStatusLocked(uint32(C.sidereon_sbas_transmit_epoch_clock_at_epoch_queries((*C.SidereonBroadcastEphemeris)(b), (*C.SidereonSbasCorrectionStore)(store), (*C.char)(unsafe.Pointer(&geoToken.bytes[0])), C.uint32_t(mode), sat, transmit, selection, &hasClock, &clock, &degraded)))
		})
		return operationErr
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(s)
	runtime.KeepAlive(transmitEpoch)
	runtime.KeepAlive(selectionEpoch)
	if err != nil {
		return TransmitEpochClock{}, err
	}
	return TransmitEpochClock{HasClock: bool(hasClock), ClockS: float64(clock), Degraded: bool(degraded), DegradeReason: degradeReason}, nil
}

func (s *SBASCorrectionStore) ClockRelativityAtEpochQuery(broadcast *BroadcastEphemeris, geo string, mode uint32, satellite string, epoch *ExactEpochQuery, position [3]float64) (ClockRelativity, error) {
	if broadcast == nil || s == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, ErrClosed
	}
	if err := validateSBASSolveModeValue(mode); err != nil {
		return ClockRelativity{}, err
	}
	geoToken, err := tokenToC(geo)
	if err != nil {
		return ClockRelativity{}, err
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return ClockRelativity{}, err
	}
	coordinates := [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])}
	var kind uint32
	var term C.double
	err = withResources([]*resource{broadcast.resource, s.resource}, func(resources []unsafe.Pointer) error {
		return epoch.handle.with(func(query unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_sbas_clock_relativity_at_epoch_query((*C.SidereonBroadcastEphemeris)(resources[0]), (*C.SidereonSbasCorrectionStore)(resources[1]), (*C.char)(unsafe.Pointer(&geoToken.bytes[0])), C.uint32_t(mode), (*C.char)(unsafe.Pointer(&sat.bytes[0])), (*C.SidereonExactEpochQuery)(query), &coordinates[0], &kind, &term))
			})
		})
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(s)
	runtime.KeepAlive(epoch)
	if err != nil {
		return ClockRelativity{}, err
	}
	return ClockRelativity{Kind: uint32(kind), TermS: float64(term)}, nil
}

func (s *SBASCorrectionStore) EphemerisVarianceAtEpochQueries(broadcast *BroadcastEphemeris, geo string, mode uint32, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery) (float64, error) {
	if err := validateSBASSolveModeValue(mode); err != nil {
		return 0, err
	}
	geoToken, err := tokenToC(geo)
	if err != nil {
		return 0, err
	}
	var variance C.double
	err = withCorrectionEpochQueries(broadcast, s, stateEpoch, selectionEpoch, satellite, func(b, store unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_sbas_ephemeris_variance_at_epoch_queries((*C.SidereonBroadcastEphemeris)(b), (*C.SidereonSbasCorrectionStore)(store), (*C.char)(unsafe.Pointer(&geoToken.bytes[0])), C.uint32_t(mode), sat, state, selection, &variance))
		})
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(s)
	runtime.KeepAlive(stateEpoch)
	runtime.KeepAlive(selectionEpoch)
	return float64(variance), err
}

func withSSRSourceQueries(broadcast *BroadcastEphemeris, store *SSRCorrectionStore, stateEpoch, selectionEpoch *ExactEpochQuery, satellite string, staleness float64, missing uint32, allowRegional bool, provider uint16, sizePolicy uint32, call func(unsafe.Pointer, unsafe.Pointer, *C.char, *C.SidereonExactEpochQuery, *C.SidereonExactEpochQuery) error) error {
	if broadcast == nil || store == nil || stateEpoch == nil || selectionEpoch == nil || stateEpoch.handle == nil || selectionEpoch.handle == nil {
		return ErrClosed
	}
	if err := validateSSRMissingActionValue(missing); err != nil {
		return err
	}
	if sizePolicy > 1 {
		return invalidArgument("invalid SSR correction size policy")
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return err
	}
	return withResources([]*resource{broadcast.resource, store.resource}, func(resources []unsafe.Pointer) error {
		return withPositioningHandles([]*positioningHandle{stateEpoch.handle, selectionEpoch.handle}, func(epochs []unsafe.Pointer) error {
			return call(resources[0], resources[1], (*C.char)(unsafe.Pointer(&sat.bytes[0])), (*C.SidereonExactEpochQuery)(epochs[0]), (*C.SidereonExactEpochQuery)(epochs[1]))
		})
	})
}

func ssrSourceStateFromC(value C.SidereonSsrCorrectedStateResult) SSRCorrectedState {
	result := SSRCorrectedState{HasState: bool(value.has_state), ClockS: float64(value.clock_s), HasGroupDelay: bool(value.has_group_delay), GroupDelayS: float64(value.group_delay_s), Degraded: bool(value.degraded), HasSizeEvent: bool(value.has_size_event), StrictRefusal: bool(value.strict_refusal), Size: SSRCorrectionSize{OrbitM: float64(value.size.orbit_m), ClockM: float64(value.size.clock_m)}, HasOversizedReport: bool(value.has_oversized_report), Source: uint32(value.source), ProviderID: uint16(value.provider_id), SolutionID: uint8(value.solution_id), OrbitRefEpochJ2000S: float64(value.orbit_ref_epoch_j2000_s), ClockRefEpochJ2000S: float64(value.clock_ref_epoch_j2000_s), FirstAppliedEpochJ2000S: float64(value.first_applied_epoch_j2000_s)}
	for axis := 0; axis < 3; axis++ {
		result.PositionECEFM[axis] = float64(value.position_ecef_m[axis])
	}
	return result
}

func (s *SSRCorrectionStore) SourceStateAtEpochQueries(broadcast *BroadcastEphemeris, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery, staleness float64, missing uint32, allowRegional bool, provider uint16, sizePolicy uint32) (SSRCorrectedState, error) {
	var output C.SidereonSsrCorrectedStateResult
	var degradeReason uint32
	err := withSSRSourceQueries(broadcast, s, stateEpoch, selectionEpoch, satellite, staleness, missing, allowRegional, provider, sizePolicy, func(b, store unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery) error {
		var operationErr error
		withCThread(func() {
			degradeReason, operationErr = rtcmSourceQueryStatusLocked(uint32(C.sidereon_ssr_corrected_state_at_epoch_queries((*C.SidereonBroadcastEphemeris)(b), (*C.SidereonSsrCorrectionStore)(store), sat, state, selection, C.double(staleness), C.uint32_t(missing), C.bool(allowRegional), C.uint16_t(provider), C.uint32_t(sizePolicy), &output)))
		})
		return operationErr
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(s)
	runtime.KeepAlive(stateEpoch)
	runtime.KeepAlive(selectionEpoch)
	if err != nil {
		return SSRCorrectedState{}, err
	}
	result := ssrSourceStateFromC(output)
	result.DegradeReason = degradeReason
	return result, nil
}

func (s *SSRCorrectionStore) TransmitEpochClockAtEpochQueries(broadcast *BroadcastEphemeris, satellite string, transmitEpoch, selectionEpoch *ExactEpochQuery, staleness float64, missing uint32, allowRegional bool, provider uint16, sizePolicy uint32) (TransmitEpochClock, SSRCorrectedState, error) {
	var hasClock, degraded C.bool
	var clock C.double
	var policy C.SidereonSsrCorrectedStateResult
	var degradeReason uint32
	err := withSSRSourceQueries(broadcast, s, transmitEpoch, selectionEpoch, satellite, staleness, missing, allowRegional, provider, sizePolicy, func(b, store unsafe.Pointer, sat *C.char, transmit, selection *C.SidereonExactEpochQuery) error {
		var operationErr error
		withCThread(func() {
			degradeReason, operationErr = rtcmSourceQueryStatusLocked(uint32(C.sidereon_ssr_transmit_epoch_clock_at_epoch_queries((*C.SidereonBroadcastEphemeris)(b), (*C.SidereonSsrCorrectionStore)(store), sat, transmit, selection, C.double(staleness), C.uint32_t(missing), C.bool(allowRegional), C.uint16_t(provider), C.uint32_t(sizePolicy), &hasClock, &clock, &degraded, &policy)))
		})
		return operationErr
	})
	if err != nil {
		return TransmitEpochClock{}, SSRCorrectedState{}, err
	}
	state := ssrSourceStateFromC(policy)
	state.DegradeReason = degradeReason
	return TransmitEpochClock{HasClock: bool(hasClock), ClockS: float64(clock), Degraded: bool(degraded), DegradeReason: degradeReason}, state, nil
}

func (s *SSRCorrectionStore) ClockRelativityAtEpochQuery(broadcast *BroadcastEphemeris, satellite string, epoch *ExactEpochQuery, position [3]float64, staleness float64, missing uint32, allowRegional bool, provider uint16, sizePolicy uint32) (ClockRelativity, SSRCorrectedState, error) {
	if broadcast == nil || s == nil || epoch == nil || epoch.handle == nil {
		return ClockRelativity{}, SSRCorrectedState{}, ErrClosed
	}
	if err := validateSSRMissingActionValue(missing); err != nil {
		return ClockRelativity{}, SSRCorrectedState{}, err
	}
	if sizePolicy > 1 {
		return ClockRelativity{}, SSRCorrectedState{}, invalidArgument("invalid SSR correction size policy")
	}
	sat, err := tokenToC(satellite)
	if err != nil {
		return ClockRelativity{}, SSRCorrectedState{}, err
	}
	coordinates := [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])}
	var kind uint32
	var term C.double
	var policy C.SidereonSsrCorrectedStateResult
	err = withResources([]*resource{broadcast.resource, s.resource}, func(resources []unsafe.Pointer) error {
		return epoch.handle.with(func(query unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_ssr_clock_relativity_at_epoch_query((*C.SidereonBroadcastEphemeris)(resources[0]), (*C.SidereonSsrCorrectionStore)(resources[1]), (*C.char)(unsafe.Pointer(&sat.bytes[0])), (*C.SidereonExactEpochQuery)(query), &coordinates[0], C.double(staleness), C.uint32_t(missing), C.bool(allowRegional), C.uint16_t(provider), C.uint32_t(sizePolicy), &kind, &term, &policy))
			})
		})
	})
	runtime.KeepAlive(broadcast)
	runtime.KeepAlive(s)
	runtime.KeepAlive(epoch)
	if err != nil {
		return ClockRelativity{}, SSRCorrectedState{}, err
	}
	return ClockRelativity{Kind: uint32(kind), TermS: float64(term)}, ssrSourceStateFromC(policy), nil
}

func (s *SSRCorrectionStore) EphemerisVarianceAtEpochQueries(broadcast *BroadcastEphemeris, satellite string, stateEpoch, selectionEpoch *ExactEpochQuery, staleness float64, missing uint32, allowRegional bool, provider uint16, sizePolicy uint32) (float64, SSRCorrectedState, error) {
	var variance C.double
	var policy C.SidereonSsrCorrectedStateResult
	err := withSSRSourceQueries(broadcast, s, stateEpoch, selectionEpoch, satellite, staleness, missing, allowRegional, provider, sizePolicy, func(b, store unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_ssr_ephemeris_variance_at_epoch_queries((*C.SidereonBroadcastEphemeris)(b), (*C.SidereonSsrCorrectionStore)(store), sat, state, selection, C.double(staleness), C.uint32_t(missing), C.bool(allowRegional), C.uint16_t(provider), C.uint32_t(sizePolicy), &variance, &policy))
		})
	})
	if err != nil {
		return 0, SSRCorrectedState{}, err
	}
	return float64(variance), ssrSourceStateFromC(policy), nil
}

func (s *SP3) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	return preciseStateQuery(s.handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery, output *C.SidereonEphemerisSourceState) uint32 {
		return uint32(C.sidereon_sp3_source_state_at_epoch_queries((*C.SidereonSp3)(source), sat, state, selection, output))
	})
}

func (s *SP3) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	return preciseClockQuery(s.handle, transmitEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, transmit, selection *C.SidereonExactEpochQuery, hasClock *C.bool, clock *C.double, degraded *C.bool) uint32 {
		return uint32(C.sidereon_sp3_source_transmit_epoch_clock_at_epoch_queries((*C.SidereonSp3)(source), sat, transmit, selection, hasClock, clock, degraded))
	})
}

func (s *SP3) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, position [3]float64) (ClockRelativity, error) {
	return preciseRelativityQuery(s.handle, epoch, satellite, position, func(source unsafe.Pointer, sat *C.char, query *C.SidereonExactEpochQuery, coordinates *C.double, kind *uint32, term *C.double) uint32 {
		return uint32(C.sidereon_sp3_source_clock_relativity_at_epoch_query((*C.SidereonSp3)(source), sat, query, coordinates, kind, term))
	})
}

func (s *SP3) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	return preciseVarianceQuery(s.handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery, variance *C.double) uint32 {
		return uint32(C.sidereon_sp3_source_ephemeris_variance_at_epoch_queries((*C.SidereonSp3)(source), sat, state, selection, variance))
	})
}

func (i *PreciseEphemerisInterpolant) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	return preciseStateQuery(i.handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery, output *C.SidereonEphemerisSourceState) uint32 {
		return uint32(C.sidereon_precise_interpolant_state_at_epoch_queries((*C.SidereonPreciseEphemerisInterpolant)(source), sat, state, selection, output))
	})
}

func (i *PreciseEphemerisInterpolant) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	return preciseClockQuery(i.handle, transmitEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, transmit, selection *C.SidereonExactEpochQuery, hasClock *C.bool, clock *C.double, degraded *C.bool) uint32 {
		return uint32(C.sidereon_precise_interpolant_transmit_epoch_clock_at_epoch_queries((*C.SidereonPreciseEphemerisInterpolant)(source), sat, transmit, selection, hasClock, clock, degraded))
	})
}

func (i *PreciseEphemerisInterpolant) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, position [3]float64) (ClockRelativity, error) {
	return preciseRelativityQuery(i.handle, epoch, satellite, position, func(source unsafe.Pointer, sat *C.char, query *C.SidereonExactEpochQuery, coordinates *C.double, kind *uint32, term *C.double) uint32 {
		return uint32(C.sidereon_precise_interpolant_clock_relativity_at_epoch_query((*C.SidereonPreciseEphemerisInterpolant)(source), sat, query, coordinates, kind, term))
	})
}

func (i *PreciseEphemerisInterpolant) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	return preciseVarianceQuery(i.handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery, variance *C.double) uint32 {
		return uint32(C.sidereon_precise_interpolant_ephemeris_variance_at_epoch_queries((*C.SidereonPreciseEphemerisInterpolant)(source), sat, state, selection, variance))
	})
}

func (a *PreciseInterpolantArtifact) SourceStateAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (EphemerisSourceState, error) {
	return preciseStateQuery(a.handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery, output *C.SidereonEphemerisSourceState) uint32 {
		return uint32(C.sidereon_precise_artifact_source_state_at_epoch_queries((*C.SidereonPreciseInterpolantArtifact)(source), sat, state, selection, output))
	})
}

func (a *PreciseInterpolantArtifact) TransmitEpochClockAtEpochQueries(transmitEpoch, selectionEpoch *ExactEpochQuery, satellite string) (TransmitEpochClock, error) {
	return preciseClockQuery(a.handle, transmitEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, transmit, selection *C.SidereonExactEpochQuery, hasClock *C.bool, clock *C.double, degraded *C.bool) uint32 {
		return uint32(C.sidereon_precise_artifact_source_transmit_epoch_clock_at_epoch_queries((*C.SidereonPreciseInterpolantArtifact)(source), sat, transmit, selection, hasClock, clock, degraded))
	})
}

func (a *PreciseInterpolantArtifact) ClockRelativityAtEpochQuery(epoch *ExactEpochQuery, satellite string, position [3]float64) (ClockRelativity, error) {
	return preciseRelativityQuery(a.handle, epoch, satellite, position, func(source unsafe.Pointer, sat *C.char, query *C.SidereonExactEpochQuery, coordinates *C.double, kind *uint32, term *C.double) uint32 {
		return uint32(C.sidereon_precise_artifact_source_clock_relativity_at_epoch_query((*C.SidereonPreciseInterpolantArtifact)(source), sat, query, coordinates, kind, term))
	})
}

func (a *PreciseInterpolantArtifact) EphemerisVarianceAtEpochQueries(stateEpoch, selectionEpoch *ExactEpochQuery, satellite string) (float64, error) {
	return preciseVarianceQuery(a.handle, stateEpoch, selectionEpoch, satellite, func(source unsafe.Pointer, sat *C.char, state, selection *C.SidereonExactEpochQuery, variance *C.double) uint32 {
		return uint32(C.sidereon_precise_artifact_source_ephemeris_variance_at_epoch_queries((*C.SidereonPreciseInterpolantArtifact)(source), sat, state, selection, variance))
	})
}
