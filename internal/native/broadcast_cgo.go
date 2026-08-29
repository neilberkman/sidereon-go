//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type BroadcastEphemeris struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

type NativeEphemerisSampleRow struct {
	SatelliteID string
	EpochJ2000S float64
	Status      uint32
	HasPosition bool
	Position    [3]float64
	HasClock    bool
	ClockS      float64
}

type NativeObservableStateRow struct {
	Position      [3]float64
	ClockS        float64
	HasClock      bool
	ElementStatus uint32
	ResultStatus  uint32
}

type NativePredictedObservables struct {
	GeometricRangeM    float64
	RangeRateMPerS     float64
	DopplerHz          float64
	HasSatelliteClock  bool
	SatelliteClockS    float64
	ElevationDeg       float64
	AzimuthDeg         float64
	TransmitOffsetUS   int64
	TransmitTimeJ2000S float64
	LOSUnit            [3]float64
	SatellitePosition  [3]float64
	SatelliteVelocity  [3]float64
}

type NativePredictRequest struct {
	SatelliteID  string
	ReceiverECEF [3]float64
	TRxJ2000S    float64
}

func validateEphemerisSampleStatusValue(value uint32) error {
	if value > EphemerisSampleGapValue {
		return invalidArgument("invalid ephemeris-sample status returned by native code")
	}
	return nil
}

func validateObservableStateElementStatusValue(value uint32) error {
	if value > ObservableStateErrorValue {
		return invalidArgument("invalid observable-state status returned by native code")
	}
	return nil
}

func validateEmissionMediaStatusValue(value uint32) error {
	if value > EmissionMediaErrorValue {
		return invalidArgument("invalid emission-media status returned by native code")
	}
	return nil
}

func cObservablesOptions(value NativeObservablesOptions) C.SidereonObservablesOptions {
	return C.SidereonObservablesOptions{carrier_hz: C.double(value.CarrierHz), light_time: C.bool(value.LightTime), sagnac: C.bool(value.Sagnac)}
}

func predictedObservablesFromC(value C.SidereonPredictedObservables) NativePredictedObservables {
	out := NativePredictedObservables{GeometricRangeM: float64(value.geometric_range_m), RangeRateMPerS: float64(value.range_rate_m_s), DopplerHz: float64(value.doppler_hz), HasSatelliteClock: bool(value.has_sat_clock_s), SatelliteClockS: float64(value.sat_clock_s), ElevationDeg: float64(value.elevation_deg), AzimuthDeg: float64(value.azimuth_deg), TransmitOffsetUS: int64(value.transmit_offset_us), TransmitTimeJ2000S: float64(value.transmit_time_j2000_s)}
	for i := range out.LOSUnit {
		out.LOSUnit[i] = float64(value.los_unit[i])
		out.SatellitePosition[i] = float64(value.sat_pos_ecef_m[i])
		out.SatelliteVelocity[i] = float64(value.sat_velocity_m_s[i])
	}
	return out
}

func withCStringArray(values []string, fn func(**C.char, C.size_t) error) error {
	var err error
	withCThread(func() {
		count, e := checkedNativeSize(len(values))
		if e != nil {
			err = e
			return
		}
		var memory unsafe.Pointer
		if len(values) > 0 {
			arraySize, e := checkedNativeAllocationSize(len(values), unsafe.Sizeof(uintptr(0)))
			if e != nil {
				err = e
				return
			}
			memory = C.malloc(C.size_t(arraySize))
			if memory == nil {
				err = errors.New("sidereon: unable to allocate native satellite array")
				return
			}
			defer C.free(memory)
		}
		pointers := unsafe.Slice((**C.char)(memory), len(values))
		allocated := make([]unsafe.Pointer, len(values))
		defer func() {
			for _, pointer := range allocated {
				if pointer != nil {
					C.free(pointer)
				}
			}
		}()
		for i, value := range values {
			if e := rejectEmbeddedNUL(value, "satellite token"); e != nil {
				err = e
				return
			}
			if len(value) >= 16 {
				err = errTokenTooLong
				return
			}
			pointer := C.CString(value)
			if pointer == nil {
				err = errors.New("sidereon: unable to allocate native satellite token")
				return
			}
			allocated[i] = unsafe.Pointer(pointer)
			pointers[i] = pointer
		}
		err = fn((**C.char)(memory), count)
	})
	runtime.KeepAlive(values)
	return err
}

func newBroadcastEphemeris(p *C.SidereonBroadcastEphemeris) (*BroadcastEphemeris, error) {
	if p == nil {
		return nil, errNilNativeHandle
	}
	h := &BroadcastEphemeris{resource: &resource{ptr: unsafe.Pointer(p), release: func(x unsafe.Pointer) { C.sidereon_broadcast_ephemeris_free((*C.SidereonBroadcastEphemeris)(x)) }}}
	h.cleanup = runtime.AddCleanup(h, cleanupResource, h.resource)
	return h, nil
}
func ParseBroadcastEphemeris(data []byte) (*BroadcastEphemeris, error) {
	var p *C.SidereonBroadcastEphemeris
	err := withInput(data, func(b *C.uint8_t, n C.size_t) uint32 { return C.sidereon_broadcast_ephemeris_parse_nav(b, n, &p) })
	if err != nil {
		if p != nil {
			withCThread(func() { C.sidereon_broadcast_ephemeris_free(p) })
		}
		return nil, err
	}
	handle, err := newBroadcastEphemeris(p)
	if err != nil && p != nil {
		withCThread(func() { C.sidereon_broadcast_ephemeris_free(p) })
	}
	return handle, err
}

func (b *BroadcastEphemeris) Close() error {
	if b == nil {
		return nil
	}
	return closeProtocolResource(b, b.resource, &b.cleanup)
}

func (b *BroadcastEphemeris) RecordCount() (int, error) {
	var count C.size_t
	err := b.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_record_count((*C.SidereonBroadcastEphemeris)(p), &count)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}
func (b *BroadcastEphemeris) Records() ([]NativeBroadcastRecord, error) {
	var result []NativeBroadcastRecord
	err := b.resource.with(func(p unsafe.Pointer) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_records_full((*C.SidereonBroadcastEphemeris)(p), nil, 0, &w, &r)
		}); e != nil {
			return e
		}
		n, e := validateNativeQuery("broadcast records", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		if _, e := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonBroadcastRecord{})); e != nil {
			return e
		}
		v := make([]C.SidereonBroadcastRecord, n)
		var q *C.SidereonBroadcastRecord
		if n > 0 {
			q = &v[0]
		}
		if e = callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_records_full((*C.SidereonBroadcastEphemeris)(p), q, C.size_t(n), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateNativeOutput("broadcast records", n, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		result = make([]NativeBroadcastRecord, z)
		for i := range result {
			result[i] = broadcastRecordFromC(v[i])
		}
		return nil
	})
	runtime.KeepAlive(b)
	return result, err
}
func (b *BroadcastEphemeris) IonoCorrections() (NativeIonoCorrections, error) {
	var v C.SidereonIonoCorrections
	err := b.resource.with(func(p unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_iono_corrections((*C.SidereonBroadcastEphemeris)(p), &v)
		})
	})
	return NativeIonoCorrections{GPSPresent: bool(v.gps.present), GPSAlpha: copyFloat4(&v.gps.alpha), GPSBeta: copyFloat4(&v.gps.beta), BeiDouPresent: bool(v.beidou.present), BeiDouAlpha: copyFloat4(&v.beidou.alpha), BeiDouBeta: copyFloat4(&v.beidou.beta), GalileoPresent: bool(v.galileo.present), GalileoAI0: float64(v.galileo.ai0), GalileoAI1: float64(v.galileo.ai1), GalileoAI2: float64(v.galileo.ai2)}, err
}
func (b *BroadcastEphemeris) LeapSeconds() (float64, bool, error) {
	var v C.double
	var p C.bool
	err := b.resource.with(func(x unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_leap_seconds((*C.SidereonBroadcastEphemeris)(x), &v, &p)
		})
	})
	return float64(v), bool(p), err
}
func (b *BroadcastEphemeris) Text() ([]byte, error) {
	var out []byte
	err := b.resource.with(func(p unsafe.Pointer) error {
		var e error
		out, e = copyByteOutput("RINEX navigation text", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_rinex_encode_nav((*C.SidereonBroadcastEphemeris)(p), o, n, w, r)
		})
		return e
	})
	return out, err
}

func (b *BroadcastEphemeris) EphemerisSample(satellites []string, start, stop, step float64) ([]NativeEphemerisSampleRow, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	var result []NativeEphemerisSampleRow
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satellitesPointer **C.char, count C.size_t) error {
			var written, required C.size_t
			call := func(out *C.SidereonEphemerisSampleRow, length C.size_t) uint32 {
				return C.sidereon_broadcast_ephemeris_sample((*C.SidereonBroadcastEphemeris)(pointer), satellitesPointer, count, C.double(start), C.double(stop), C.double(step), out, length, &written, &required)
			}
			if err := callStatus(func() uint32 { return call(nil, 0) }); err != nil {
				return err
			}
			n, err := validateNativeQuery("broadcast ephemeris sample", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonEphemerisSampleRow{})); err != nil {
				return err
			}
			values := make([]C.SidereonEphemerisSampleRow, n)
			var output *C.SidereonEphemerisSampleRow
			if n > 0 {
				output = &values[0]
			}
			if err := callStatus(func() uint32 { return call(output, C.size_t(n)) }); err != nil {
				return err
			}
			writtenCount, err := validateNativeOutput("broadcast ephemeris sample", n, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			result = make([]NativeEphemerisSampleRow, writtenCount)
			for i := range result {
				value := values[i]
				if err := validateEphemerisSampleStatusValue(uint32(value.status)); err != nil {
					return err
				}
				row := NativeEphemerisSampleRow{SatelliteID: tokenFromC(value.sat_id), EpochJ2000S: float64(value.epoch_j2000_s), Status: uint32(value.status), HasPosition: bool(value.has_position_ecef_m), HasClock: bool(value.has_clock_s), ClockS: float64(value.clock_s)}
				for axis := range row.Position {
					row.Position[axis] = float64(value.position_ecef_m[axis])
				}
				result[i] = row
			}
			return nil
		})
	})
	runtime.KeepAlive(b)
	return result, err
}

func (b *BroadcastEphemeris) ObservableState(satellite string, epoch float64) ([3]float64, float64, bool, error) {
	if b == nil || b.resource == nil {
		return [3]float64{}, 0, false, ErrClosed
	}
	var position [3]C.double
	var clock C.double
	var hasClock C.bool
	var err error
	if err = rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return [3]float64{}, 0, false, err
	}
	if len(satellite) >= 16 {
		return [3]float64{}, 0, false, errTokenTooLong
	}
	err = b.resource.with(func(pointer unsafe.Pointer) error {
		return withToken(satellite, "satellite token", func(value *C.char) uint32 {
			return C.sidereon_broadcast_observable_state((*C.SidereonBroadcastEphemeris)(pointer), value, C.double(epoch), &position[0], &clock, &hasClock)
		})
	})
	var out [3]float64
	for i := range out {
		out[i] = float64(position[i])
	}
	runtime.KeepAlive(b)
	return out, float64(clock), bool(hasClock), err
}

func (b *BroadcastEphemeris) ObservableStates(satellites []string, epochs []float64) ([]NativeObservableStateRow, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	if len(satellites) != len(epochs) {
		return nil, errors.New("sidereon: satellite and epoch lengths differ")
	}
	if _, err := checkedNativeAllocationSize(len(satellites), unsafe.Sizeof(C.SidereonObservableStateElementStatus(0))); err != nil {
		return nil, err
	}
	var result []NativeObservableStateRow
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satellitesPointer **C.char, count C.size_t) error {
			n := len(satellites)
			positionCount, err := checkedNativeProduct(n, 3, "observable-state position")
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, 3*unsafe.Sizeof(C.double(0))); err != nil {
				return err
			}
			positions := make([]C.double, positionCount)
			clocks := make([]C.double, n)
			hasClocks := make([]C.bool, n)
			elements := make([]C.enum_SidereonObservableStateElementStatus, n)
			statuses := make([]C.enum_SidereonStatus, n)
			epochMemory, err := checkedNativeMalloc(n, unsafe.Sizeof(C.double(0)))
			if err != nil {
				return err
			}
			if epochMemory != nil {
				defer C.free(epochMemory)
			}
			nativeEpochs := unsafe.Slice((*C.double)(epochMemory), n)
			for i, epoch := range epochs {
				nativeEpochs[i] = C.double(epoch)
			}
			var positionPointer, clockPointer *C.double
			var hasClockPointer *C.bool
			var elementPointer *C.enum_SidereonObservableStateElementStatus
			var statusPointer *C.enum_SidereonStatus
			var epochPointer *C.double
			if n > 0 {
				positionPointer = &positions[0]
				clockPointer = &clocks[0]
				hasClockPointer = &hasClocks[0]
				elementPointer = &elements[0]
				statusPointer = &statuses[0]
				epochPointer = &nativeEpochs[0]
			}
			if err := callStatus(func() uint32 {
				return C.sidereon_broadcast_observable_states_at_j2000_s((*C.SidereonBroadcastEphemeris)(pointer), satellitesPointer, epochPointer, count, positionPointer, clockPointer, hasClockPointer, elementPointer, statusPointer)
			}); err != nil {
				return err
			}
			result = make([]NativeObservableStateRow, n)
			for i := range result {
				result[i].ClockS = float64(clocks[i])
				result[i].HasClock = bool(hasClocks[i])
				if err := validateObservableStateElementStatusValue(uint32(elements[i])); err != nil {
					return err
				}
				result[i].ElementStatus = uint32(elements[i])
				result[i].ResultStatus = uint32(statuses[i])
				for axis := range result[i].Position {
					result[i].Position[axis] = float64(positions[i*3+axis])
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(b)
	runtime.KeepAlive(epochs)
	return result, err
}

func (b *BroadcastEphemeris) ObservableStatesShared(satellites []string, epoch float64) ([]NativeObservableStateRow, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	var result []NativeObservableStateRow
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withCStringArray(satellites, func(satellitesPointer **C.char, count C.size_t) error {
			n := len(satellites)
			positionCount, err := checkedNativeProduct(n, 3, "shared observable-state position")
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(n, 3*unsafe.Sizeof(C.double(0))); err != nil {
				return err
			}
			positions := make([]C.double, positionCount)
			clocks := make([]C.double, n)
			hasClocks := make([]C.bool, n)
			elements := make([]C.enum_SidereonObservableStateElementStatus, n)
			statuses := make([]C.enum_SidereonStatus, n)
			var pp, cp *C.double
			var hp *C.bool
			var ep *C.enum_SidereonObservableStateElementStatus
			var sp *C.enum_SidereonStatus
			if n > 0 {
				pp, cp, hp, ep, sp = &positions[0], &clocks[0], &hasClocks[0], &elements[0], &statuses[0]
			}
			if err := callStatus(func() uint32 {
				return C.sidereon_broadcast_observable_states_at_shared_j2000_s((*C.SidereonBroadcastEphemeris)(pointer), satellitesPointer, count, C.double(epoch), pp, cp, hp, ep, sp)
			}); err != nil {
				return err
			}
			result = make([]NativeObservableStateRow, n)
			for i := range result {
				result[i].ClockS, result[i].HasClock = float64(clocks[i]), bool(hasClocks[i])
				if err := validateObservableStateElementStatusValue(uint32(elements[i])); err != nil {
					return err
				}
				result[i].ElementStatus, result[i].ResultStatus = uint32(elements[i]), uint32(statuses[i])
				for axis := range result[i].Position {
					result[i].Position[axis] = float64(positions[i*3+axis])
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(b)
	return result, err
}

func (b *BroadcastEphemeris) PredictObservables(satellite string, receiver [3]float64, epoch float64, options *NativeObservablesOptions) (NativePredictedObservables, error) {
	if b == nil || b.resource == nil {
		return NativePredictedObservables{}, ErrClosed
	}
	if err := rejectEmbeddedNUL(satellite, "satellite token"); err != nil {
		return NativePredictedObservables{}, err
	}
	if len(satellite) >= 16 {
		return NativePredictedObservables{}, errTokenTooLong
	}
	var value C.SidereonPredictedObservables
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		return withTokenError(satellite, "satellite token", func(sat *C.char) error {
			var optionMemory unsafe.Pointer
			var optionPointer *C.SidereonObservablesOptions
			if options != nil {
				var err error
				optionMemory, err = checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservablesOptions{}))
				if err != nil {
					return err
				}
				defer C.free(optionMemory)
				optionPointer = (*C.SidereonObservablesOptions)(optionMemory)
				*optionPointer = cObservablesOptions(*options)
			}
			receiverMemory, err := checkedNativeMalloc(3, unsafe.Sizeof(C.double(0)))
			if err != nil {
				return err
			}
			defer C.free(receiverMemory)
			receiverValue := unsafe.Slice((*C.double)(receiverMemory), 3)
			for i := range receiverValue {
				receiverValue[i] = C.double(receiver[i])
			}
			return statusErrorLocked(C.sidereon_broadcast_observables((*C.SidereonBroadcastEphemeris)(pointer), sat, (*C.double)(receiverMemory), C.double(epoch), optionPointer, &value))
		})
	})
	return predictedObservablesFromC(value), err
}

func (b *BroadcastEphemeris) PredictObservablesBatch(requests []NativePredictRequest, options *NativeObservablesOptions) ([]NativePredictedObservables, []bool, error) {
	if b == nil || b.resource == nil {
		return nil, nil, ErrClosed
	}
	var result []NativePredictedObservables
	var ok []bool
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		requestCount, err := checkedNativeSize(len(requests))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(len(requests), unsafe.Sizeof(C.SidereonPredictRequest{})); err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(len(requests), unsafe.Sizeof(C.SidereonPredictedObservables{})); err != nil {
			return err
		}
		satellites := make([]string, len(requests))
		for i, request := range requests {
			satellites[i] = request.SatelliteID
		}
		return withCStringArray(satellites, func(satellitesPointer **C.char, count C.size_t) error {
			requestMemory, err := checkedNativeMalloc(len(requests), unsafe.Sizeof(C.SidereonPredictRequest{}))
			if err != nil {
				return err
			}
			if requestMemory != nil {
				defer C.free(requestMemory)
			}
			nativeRequests := unsafe.Slice((*C.SidereonPredictRequest)(requestMemory), len(requests))
			pointers := unsafe.Slice(satellitesPointer, len(requests))
			for i, request := range requests {
				nativeRequests[i].sat_id = pointers[i]
				nativeRequests[i].receiver_ecef_m[0] = C.double(request.ReceiverECEF[0])
				nativeRequests[i].receiver_ecef_m[1] = C.double(request.ReceiverECEF[1])
				nativeRequests[i].receiver_ecef_m[2] = C.double(request.ReceiverECEF[2])
				nativeRequests[i].t_rx_j2000_s = C.double(request.TRxJ2000S)
			}
			values := make([]C.SidereonPredictedObservables, len(requests))
			accepted := make([]C.bool, len(requests))
			var requestPointer *C.SidereonPredictRequest
			var valuePointer *C.SidereonPredictedObservables
			var acceptedPointer *C.bool
			if len(requests) > 0 {
				requestPointer = &nativeRequests[0]
				valuePointer = &values[0]
				acceptedPointer = &accepted[0]
			}
			var optionMemory unsafe.Pointer
			var optionPointer *C.SidereonObservablesOptions
			if options != nil {
				optionMemory, err = checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonObservablesOptions{}))
				if err != nil {
					return err
				}
				defer C.free(optionMemory)
				optionPointer = (*C.SidereonObservablesOptions)(optionMemory)
				*optionPointer = cObservablesOptions(*options)
			}
			if err := callStatus(func() uint32 {
				return C.sidereon_broadcast_observables_batch((*C.SidereonBroadcastEphemeris)(pointer), requestPointer, requestCount, optionPointer, valuePointer, acceptedPointer)
			}); err != nil {
				return err
			}
			result = make([]NativePredictedObservables, len(values))
			ok = make([]bool, len(values))
			for i := range result {
				result[i] = predictedObservablesFromC(values[i])
				ok[i] = bool(accepted[i])
			}
			return nil
		})
	})
	runtime.KeepAlive(b)
	return result, ok, err
}
