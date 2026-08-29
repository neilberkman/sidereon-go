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
	"unsafe"
)

type AntennaPco struct {
	NorthM float64
	EastM  float64
	UpM    float64
}

type Antenna struct {
	handle *positioningHandle
}

type ANTEX struct {
	handle *positioningHandle
}

func ParseANTEX(data []byte) (*ANTEX, error) {
	var pointer *C.SidereonAntex
	operationErr := withInput(data, func(input *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_antex_parse(input, length, &pointer)
	})
	if operationErr != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_antex_free(pointer) })
		}
		return nil, operationErr
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native ANTEX constructor returned no handle")
	}
	return &ANTEX{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_antex_free((*C.SidereonAntex)(value))
	})}, nil
}

func (a *ANTEX) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}

func (a *ANTEX) AntennaCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_antex_antenna_count((*C.SidereonAntex)(pointer), &count) })
	})
	runtime.KeepAlive(a)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func antennaFromPointer(pointer *C.SidereonAntenna) *Antenna {
	if pointer == nil {
		return nil
	}
	return &Antenna{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_antenna_free((*C.SidereonAntenna)(value))
	})}
}

func (a *ANTEX) Antenna(id string) (*Antenna, bool, error) {
	if a == nil || a.handle == nil {
		return nil, false, ErrClosed
	}
	var pointer *C.SidereonAntenna
	err := a.handle.with(func(owner unsafe.Pointer) error {
		return withString(id, func(input *C.char) uint32 {
			return C.sidereon_antex_antenna((*C.SidereonAntex)(owner), input, &pointer)
		})
	})
	runtime.KeepAlive(a)
	if err != nil {
		return nil, false, err
	}
	return antennaFromPointer(pointer), pointer != nil, nil
}

func (a *ANTEX) Encode() ([]byte, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	var output []byte
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		output, operationErr = copyNativeBytes("ANTEX encoding", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_antex_encode((*C.SidereonAntex)(pointer), out, length, written, required)
		})
		return operationErr
	})
	runtime.KeepAlive(a)
	return output, err
}

func (a *Antenna) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return a.handle.close()
}

func (a *Antenna) PCO(frequency string) (AntennaPco, error) {
	if a == nil || a.handle == nil {
		return AntennaPco{}, ErrClosed
	}
	var output [3]C.double
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withString(frequency, func(input *C.char) uint32 {
			return C.sidereon_antenna_pco((*C.SidereonAntenna)(pointer), input, &output[0])
		})
	})
	runtime.KeepAlive(a)
	return AntennaPco{NorthM: float64(output[0]), EastM: float64(output[1]), UpM: float64(output[2])}, err
}

func (a *Antenna) PCV(frequency string, zenithDeg float64, hasAzimuth bool, azimuthDeg float64) (float64, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	var output C.double
	err := a.handle.with(func(pointer unsafe.Pointer) error {
		return withString(frequency, func(input *C.char) uint32 {
			return C.sidereon_antenna_pcv((*C.SidereonAntenna)(pointer), input, C.double(zenithDeg), C.bool(hasAzimuth), C.double(azimuthDeg), &output)
		})
	})
	runtime.KeepAlive(a)
	return float64(output), err
}

type AtmosphereInput struct {
	Year         int
	DayOfYear    int
	Second       float64
	AltitudeKm   float64
	LatitudeDeg  float64
	LongitudeDeg float64
	HasLST       bool
	LST          float64
	F107         float64
	F107A        float64
	Ap           float64
	HasApArray   bool
	ApArray      [7]float64
}

type AtmosphereOutput struct {
	DensityKgM3  float64
	TemperatureK float64
}

func atmosphereInputToC(input AtmosphereInput) (C.SidereonAtmosphereInput, error) {
	var output C.SidereonAtmosphereInput
	year, err := checkedInt32(input.Year, "NRLMSISE year")
	if err != nil {
		return output, err
	}
	doy, err := checkedInt32(input.DayOfYear, "NRLMSISE day of year")
	if err != nil {
		return output, err
	}
	output.year = C.int32_t(year)
	output.doy = C.int32_t(doy)
	output.sec = C.double(input.Second)
	output.alt_km = C.double(input.AltitudeKm)
	output.lat_deg = C.double(input.LatitudeDeg)
	output.lon_deg = C.double(input.LongitudeDeg)
	output.has_lst = C.bool(input.HasLST)
	output.lst = C.double(input.LST)
	output.f107 = C.double(input.F107)
	output.f107a = C.double(input.F107A)
	output.ap = C.double(input.Ap)
	output.has_ap_array = C.bool(input.HasApArray)
	for i := range input.ApArray {
		output.ap_array[i] = C.double(input.ApArray[i])
	}
	return output, nil
}

func atmosphereInputFromC(input C.SidereonAtmosphereInput) AtmosphereInput {
	var output AtmosphereInput
	output.Year = int(input.year)
	output.DayOfYear = int(input.doy)
	output.Second = float64(input.sec)
	output.AltitudeKm = float64(input.alt_km)
	output.LatitudeDeg = float64(input.lat_deg)
	output.LongitudeDeg = float64(input.lon_deg)
	output.HasLST = bool(input.has_lst)
	output.LST = float64(input.lst)
	output.F107 = float64(input.f107)
	output.F107A = float64(input.f107a)
	output.Ap = float64(input.ap)
	output.HasApArray = bool(input.has_ap_array)
	for i := range output.ApArray {
		output.ApArray[i] = float64(input.ap_array[i])
	}
	return output
}

func AtmosphereInputDefault() AtmosphereInput {
	return atmosphereInputFromC(C.sidereon_atmosphere_input_default())
}

func AtmosphereNRLMSISE00(input AtmosphereInput) (AtmosphereOutput, error) {
	cInput, err := atmosphereInputToC(input)
	if err != nil {
		return AtmosphereOutput{}, err
	}
	var output C.SidereonAtmosphereOutput
	err = callStatus(func() uint32 { return C.sidereon_atmosphere_nrlmsise00(&cInput, &output) })
	return AtmosphereOutput{DensityKgM3: float64(output.density_kg_m3), TemperatureK: float64(output.temperature_k)}, err
}
