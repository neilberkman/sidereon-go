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

type NativeFrequencyChannel struct {
	Slot    uint8
	Channel int8
}

type NativeGlonassRecord struct {
	SatelliteID        string
	ToeUTCJ2000S       float64
	PositionM          [3]float64
	VelocityMPerS      [3]float64
	AccelerationMPerS2 [3]float64
	ClockBiasS         float64
	GammaN             float64
	SVHealth           float64
	FrequencyChannel   int32
}

type NativeSkippedGlonassRecord struct {
	SatelliteID string
}

type RinexGlonassRecords struct {
	_        noCopy
	resource *resource
	cleanup  runtime.Cleanup
}

func newRinexGlonassRecords(pointer *C.SidereonRinexGlonassRecords) (*RinexGlonassRecords, error) {
	if pointer == nil {
		return nil, errNilNativeHandle
	}
	handle := &RinexGlonassRecords{resource: &resource{ptr: unsafe.Pointer(pointer), release: func(value unsafe.Pointer) {
		C.sidereon_rinex_glonass_records_free((*C.SidereonRinexGlonassRecords)(value))
	}}}
	handle.cleanup = runtime.AddCleanup(handle, cleanupResource, handle.resource)
	return handle, nil
}

func ParseRinexGlonassRecords(data []byte) (*RinexGlonassRecords, error) {
	var pointer *C.SidereonRinexGlonassRecords
	err := withInput(data, func(value *C.uint8_t, length C.size_t) uint32 {
		return C.sidereon_parse_rinex_glonass_records(value, length, &pointer)
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_rinex_glonass_records_free(pointer) })
		}
		return nil, err
	}
	handle, err := newRinexGlonassRecords(pointer)
	if err != nil && pointer != nil {
		withCThread(func() { C.sidereon_rinex_glonass_records_free(pointer) })
	}
	return handle, err
}

func (r *RinexGlonassRecords) Close() error {
	if r == nil {
		return nil
	}
	return closeProtocolResource(r, r.resource, &r.cleanup)
}

func (r *RinexGlonassRecords) Count() (int, error) {
	if r == nil || r.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := r.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_glonass_records_count((*C.SidereonRinexGlonassRecords)(pointer), &count)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func glonassRecordFromC(value C.SidereonGlonassRecord) NativeGlonassRecord {
	result := NativeGlonassRecord{SatelliteID: tokenFromC(value.sat_id), ToeUTCJ2000S: float64(value.toe_utc_j2000_s), ClockBiasS: float64(value.clk_bias), GammaN: float64(value.gamma_n), SVHealth: float64(value.sv_health), FrequencyChannel: int32(value.freq_channel)}
	for axis := range result.PositionM {
		result.PositionM[axis] = float64(value.pos_m[axis])
		result.VelocityMPerS[axis] = float64(value.vel_m_s[axis])
		result.AccelerationMPerS2[axis] = float64(value.acc_m_s2[axis])
	}
	return result
}

func (r *RinexGlonassRecords) Record(index int) (NativeGlonassRecord, error) {
	if index < 0 {
		return NativeGlonassRecord{}, errNegativeIndex
	}
	if r == nil || r.resource == nil {
		return NativeGlonassRecord{}, ErrClosed
	}
	var value C.SidereonGlonassRecord
	err := r.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_glonass_records_item((*C.SidereonRinexGlonassRecords)(pointer), C.size_t(index), &value)
		})
	})
	return glonassRecordFromC(value), err
}

func (r *RinexGlonassRecords) SkippedCount() (int, error) {
	if r == nil || r.resource == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := r.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_glonass_records_skipped_count((*C.SidereonRinexGlonassRecords)(pointer), &count)
		})
	})
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (r *RinexGlonassRecords) Skipped(index int) (NativeSkippedGlonassRecord, error) {
	if index < 0 {
		return NativeSkippedGlonassRecord{}, errNegativeIndex
	}
	if r == nil || r.resource == nil {
		return NativeSkippedGlonassRecord{}, ErrClosed
	}
	var value C.SidereonSkippedGlonassRecord
	err := r.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_rinex_glonass_records_skipped_item((*C.SidereonRinexGlonassRecords)(pointer), C.size_t(index), &value)
		})
	})
	return NativeSkippedGlonassRecord{SatelliteID: tokenFromC(value.satellite)}, err
}

func (b *BroadcastEphemeris) GlonassFrequencyChannels() ([]NativeFrequencyChannel, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	var result []NativeFrequencyChannel
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_glonass_frequency_channel_count((*C.SidereonBroadcastEphemeris)(pointer), &count)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonFrequencyChannel{})); err != nil {
			return err
		}
		values := make([]C.SidereonFrequencyChannel, n)
		var output *C.SidereonFrequencyChannel
		if n > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_glonass_frequency_channels((*C.SidereonBroadcastEphemeris)(pointer), output, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		z, err := validateNativeOutput("broadcast GLONASS frequency channels", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeFrequencyChannel, z)
		for i := range result {
			result[i] = NativeFrequencyChannel{Slot: uint8(values[i].slot), Channel: int8(values[i].channel)}
		}
		return nil
	})
	runtime.KeepAlive(b)
	return result, err
}

func (b *BroadcastEphemeris) GlonassRecords() ([]NativeGlonassRecord, error) {
	if b == nil || b.resource == nil {
		return nil, ErrClosed
	}
	var result []NativeGlonassRecord
	err := b.resource.with(func(pointer unsafe.Pointer) error {
		var count C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_glonass_record_count((*C.SidereonBroadcastEphemeris)(pointer), &count)
		}); err != nil {
			return err
		}
		n, err := checkedNativeCount(uint64(count))
		if err != nil {
			return err
		}
		if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonGlonassRecord{})); err != nil {
			return err
		}
		values := make([]C.SidereonGlonassRecord, n)
		var output *C.SidereonGlonassRecord
		if n > 0 {
			output = &values[0]
		}
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_broadcast_ephemeris_glonass_records((*C.SidereonBroadcastEphemeris)(pointer), output, C.size_t(n), &written, &required)
		}); err != nil {
			return err
		}
		z, err := validateNativeOutput("broadcast GLONASS records", n, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		result = make([]NativeGlonassRecord, z)
		for i := range result {
			result[i] = glonassRecordFromC(values[i])
		}
		return nil
	})
	runtime.KeepAlive(b)
	return result, err
}
