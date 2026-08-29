//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type NativeObservablesOptions struct {
	CarrierHz         float64
	LightTime, Sagnac bool
}
type NativeEmissionMediaOptions struct {
	CarrierHz                                   float64
	MinElevationEnabled                         bool
	MinElevationRad                             float64
	TroposphereEnabled                          bool
	PressureHPA, TemperatureK, RelativeHumidity float64
}
type NativeEmissionMediaRow struct {
	Position                [3]float64
	HasPosition             bool
	ClockS                  float64
	HasClock                bool
	IonosphereSlantDelayM   float64
	HasIonosphereSlantDelay bool
	TroposphereDelayM       float64
	HasTroposphereDelay     bool
	Status                  uint32
	ResultStatus            uint32
}

func ObservablesOptionsInit() (NativeObservablesOptions, error) {
	var v C.SidereonObservablesOptions
	err := callStatus(func() uint32 { return C.sidereon_observables_options_init(&v) })
	return NativeObservablesOptions{CarrierHz: float64(v.carrier_hz), LightTime: bool(v.light_time), Sagnac: bool(v.sagnac)}, err
}
func EmissionMediaOptionsInit() (NativeEmissionMediaOptions, error) {
	var v C.SidereonEmissionMediaOptions
	err := callStatus(func() uint32 { return C.sidereon_emission_media_options_init(&v) })
	return NativeEmissionMediaOptions{CarrierHz: float64(v.carrier_hz), MinElevationEnabled: bool(v.min_elevation_enabled), MinElevationRad: float64(v.min_elevation_rad), TroposphereEnabled: bool(v.troposphere_enabled), PressureHPA: float64(v.met.pressure_hpa), TemperatureK: float64(v.met.temperature_k), RelativeHumidity: float64(v.met.relative_humidity)}, err
}
func MissingObservablePosition() ([3]float64, error) {
	var v [3]C.double
	err := callStatus(func() uint32 { return C.sidereon_observable_state_missing_position_ecef_m(&v[0], 3) })
	var out [3]float64
	for i := range out {
		out[i] = float64(v[i])
	}
	return out, err
}
func cEmission(v NativeEmissionMediaOptions) C.SidereonEmissionMediaOptions {
	return C.SidereonEmissionMediaOptions{carrier_hz: C.double(v.CarrierHz), min_elevation_enabled: C.bool(v.MinElevationEnabled), min_elevation_rad: C.double(v.MinElevationRad), troposphere_enabled: C.bool(v.TroposphereEnabled), met: C.SidereonMet{pressure_hpa: C.double(v.PressureHPA), temperature_k: C.double(v.TemperatureK), relative_humidity: C.double(v.RelativeHumidity)}}
}

func emissionBatchWith(withHandle func(func(unsafe.Pointer) error) error, satellites []string, epochs []float64, receiver [3]float64, opt *NativeEmissionMediaOptions, broadcast bool) ([]NativeEmissionMediaRow, error) {
	if len(satellites) != len(epochs) {
		return nil, errors.New("sidereon: satellites and epochs have different lengths")
	}
	var result []NativeEmissionMediaRow
	err := withHandle(func(bp unsafe.Pointer) error {
		var optMemory unsafe.Pointer
		var opp *C.SidereonEmissionMediaOptions
		if opt != nil {
			var allocationErr error
			optMemory, allocationErr = checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonEmissionMediaOptions{}))
			if allocationErr != nil {
				return allocationErr
			}
			defer C.free(optMemory)
			opp = (*C.SidereonEmissionMediaOptions)(optMemory)
			*opp = cEmission(*opt)
		}
		var satMem, epochMem unsafe.Pointer
		n := len(satellites)
		nativeCount, err := checkedNativeSize(n)
		if err != nil {
			return err
		}
		if n > 0 {
			satSize, err := checkedNativeAllocationSize(n, unsafe.Sizeof((*C.char)(nil)))
			if err != nil {
				return err
			}
			satMem = C.malloc(C.size_t(satSize))
			if satMem == nil {
				return errors.New("sidereon: unable to allocate native satellite pointers")
			}
			defer C.free(satMem)
			epochSize, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.double(0)))
			if err != nil {
				return err
			}
			epochMem = C.malloc(C.size_t(epochSize))
			if epochMem == nil {
				return errors.New("sidereon: unable to allocate native epochs")
			}
			defer C.free(epochMem)
		}
		satPtrs := unsafe.Slice((**C.char)(satMem), n)
		epochPtrs := unsafe.Slice((*C.double)(epochMem), n)
		allocated := make([]unsafe.Pointer, n)
		defer func() {
			for _, p := range allocated {
				if p != nil {
					C.free(p)
				}
			}
		}()
		for i, s := range satellites {
			if e := rejectEmbeddedNUL(s, "satellite token"); e != nil {
				return e
			}
			p := C.CString(s)
			if p == nil {
				return errors.New("sidereon: unable to allocate native satellite token")
			}
			allocated[i] = unsafe.Pointer(p)
			satPtrs[i] = p
			epochPtrs[i] = C.double(epochs[i])
		}
		var posMem, clockMem, ionoMem, tropoMem, hasPosMem, hasClockMem, hasIonoMem, hasTropoMem, statusMem, resultStatusMem unsafe.Pointer
		outputMemory := make([]unsafe.Pointer, 0, 10)
		defer func() {
			for _, pointer := range outputMemory {
				C.free(pointer)
			}
		}()
		if n > 0 {
			var err error
			if posMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.double(0))*3); err != nil {
				return err
			}
			outputMemory = append(outputMemory, posMem)
			if clockMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.double(0))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, clockMem)
			if ionoMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.double(0))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, ionoMem)
			if tropoMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.double(0))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, tropoMem)
			if hasPosMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.bool(false))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, hasPosMem)
			if hasClockMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.bool(false))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, hasClockMem)
			if hasIonoMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.bool(false))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, hasIonoMem)
			if hasTropoMem, err = checkedNativeMalloc(n, unsafe.Sizeof(C.bool(false))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, hasTropoMem)
			if statusMem, err = checkedNativeMalloc(n, unsafe.Sizeof(uint32(0))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, statusMem)
			if resultStatusMem, err = checkedNativeMalloc(n, unsafe.Sizeof(uint32(0))); err != nil {
				return err
			}
			outputMemory = append(outputMemory, resultStatusMem)
		}
		if n > 0 && (posMem == nil || clockMem == nil || ionoMem == nil || tropoMem == nil || hasPosMem == nil || hasClockMem == nil || hasIonoMem == nil || hasTropoMem == nil || statusMem == nil || resultStatusMem == nil) {
			return errors.New("sidereon: unable to allocate native emission output")
		}
		receiverMemory, allocationErr := checkedNativeMalloc(3, unsafe.Sizeof(C.double(0)))
		if allocationErr != nil {
			return allocationErr
		}
		defer C.free(receiverMemory)
		receiverValues := unsafe.Slice((*C.double)(receiverMemory), 3)
		for i := range receiverValues {
			receiverValues[i] = C.double(receiver[i])
		}
		var status C.enum_SidereonStatus
		if broadcast {
			status = C.sidereon_broadcast_emission_media_batch_at_j2000_s((*C.SidereonBroadcastEphemeris)(bp), (**C.char)(satMem), (*C.double)(epochMem), nativeCount, (*C.double)(receiverMemory), opp, (*C.double)(posMem), (*C.bool)(hasPosMem), (*C.double)(clockMem), (*C.bool)(hasClockMem), (*C.double)(ionoMem), (*C.bool)(hasIonoMem), (*C.double)(tropoMem), (*C.bool)(hasTropoMem), (*C.enum_SidereonEmissionMediaStatus)(statusMem), (*C.enum_SidereonStatus)(resultStatusMem))
		} else {
			status = C.sidereon_sp3_emission_media_batch_at_j2000_s((*C.SidereonSp3)(bp), (**C.char)(satMem), (*C.double)(epochMem), nativeCount, (*C.double)(receiverMemory), opp, (*C.double)(posMem), (*C.bool)(hasPosMem), (*C.double)(clockMem), (*C.bool)(hasClockMem), (*C.double)(ionoMem), (*C.bool)(hasIonoMem), (*C.double)(tropoMem), (*C.bool)(hasTropoMem), (*C.enum_SidereonEmissionMediaStatus)(statusMem), (*C.enum_SidereonStatus)(resultStatusMem))
		}
		if status != C.SIDEREON_STATUS_OK {
			return statusErrorLocked(uint32(status))
		}
		result = make([]NativeEmissionMediaRow, n)
		positionCount, err := checkedNativeProduct(n, 3, "emission-media position")
		if err != nil {
			return err
		}
		pv := unsafe.Slice((*C.double)(posMem), positionCount)
		cv := unsafe.Slice((*C.double)(clockMem), n)
		iv := unsafe.Slice((*C.double)(ionoMem), n)
		tv := unsafe.Slice((*C.double)(tropoMem), n)
		pp := unsafe.Slice((*C.bool)(hasPosMem), n)
		pc := unsafe.Slice((*C.bool)(hasClockMem), n)
		pi := unsafe.Slice((*C.bool)(hasIonoMem), n)
		pt := unsafe.Slice((*C.bool)(hasTropoMem), n)
		ps := unsafe.Slice((*uint32)(statusMem), n)
		pr := unsafe.Slice((*uint32)(resultStatusMem), n)
		for i := range result {
			for j := 0; j < 3; j++ {
				result[i].Position[j] = float64(pv[i*3+j])
			}
			result[i].HasPosition = bool(pp[i])
			result[i].ClockS = float64(cv[i])
			result[i].HasClock = bool(pc[i])
			result[i].IonosphereSlantDelayM = float64(iv[i])
			result[i].HasIonosphereSlantDelay = bool(pi[i])
			result[i].TroposphereDelayM = float64(tv[i])
			result[i].HasTroposphereDelay = bool(pt[i])
			if e := validateEmissionMediaStatusValue(ps[i]); e != nil {
				return e
			}
			result[i].Status = ps[i]
			result[i].ResultStatus = pr[i]
		}
		return nil
	})
	runtime.KeepAlive(satellites)
	runtime.KeepAlive(epochs)
	return result, err
}

func checkedNativeMalloc(count int, elementSize uintptr) (unsafe.Pointer, error) {
	size, err := checkedNativeAllocationSize(count, elementSize)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}
	pointer := C.malloc(C.size_t(size))
	if pointer == nil {
		return nil, errors.New("sidereon: unable to allocate native emission output")
	}
	return pointer, nil
}

func emissionBatch(b *BroadcastEphemeris, satellites []string, epochs []float64, receiver [3]float64, opt *NativeEmissionMediaOptions) ([]NativeEmissionMediaRow, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	return emissionBatchWith(b.resource.with, satellites, epochs, receiver, opt, true)
}

func (s *SP3) EmissionBatch(satellites []string, epochs []float64, receiver [3]float64, opt *NativeEmissionMediaOptions) ([]NativeEmissionMediaRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	return emissionBatchWith(s.handle.with, satellites, epochs, receiver, opt, false)
}
func (b *BroadcastEphemeris) EmissionBatch(satellites []string, epochs []float64, receiver [3]float64, opt *NativeEmissionMediaOptions) ([]NativeEmissionMediaRow, error) {
	return emissionBatch(b, satellites, epochs, receiver, opt)
}
