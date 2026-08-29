//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type NMEAChunkSummary struct {
	SentenceCount       uint64
	CompletedEpochCount uint64
	SkipCount           uint64
	WarningCount        uint64
	RetainedLength      uint64
}

type NMEAGGAOptions struct {
	Talker             string
	UTCSecondsOfDay    float64
	Position           Geodetic
	Quality            uint32
	SatellitesUsed     uint8
	HDOP               float64
	CoordinateDecimals uint8
}

type NMEAAccumulator struct {
	handle *positioningHandle
}

func nmeaChunkSummaryFromC(value C.SidereonNmeaChunkSummary) NMEAChunkSummary {
	return NMEAChunkSummary{
		SentenceCount:       uint64(value.sentence_count),
		CompletedEpochCount: uint64(value.completed_epoch_count),
		SkipCount:           uint64(value.skip_count),
		WarningCount:        uint64(value.warning_count),
		RetainedLength:      uint64(value.retained_len),
	}
}

func NewNMEAAccumulator() (*NMEAAccumulator, error) {
	var pointer *C.SidereonNmeaAccumulator
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_nmea_accumulator_new(&pointer))
	})
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_nmea_accumulator_free(pointer) })
		}
		return nil, err
	}
	if pointer == nil {
		return nil, missingNativeHandle("NMEA accumulator")
	}
	return &NMEAAccumulator{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_nmea_accumulator_free((*C.SidereonNmeaAccumulator)(value))
	})}, nil
}

func (accumulator *NMEAAccumulator) Close() error {
	if accumulator == nil || accumulator.handle == nil {
		return nil
	}
	return accumulator.handle.close()
}

func (accumulator *NMEAAccumulator) Push(data []byte) (NMEAChunkSummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEAChunkSummary{}, ErrClosed
	}
	length, err := checkedNativeSize(len(data))
	if err != nil {
		return NMEAChunkSummary{}, err
	}
	var output C.SidereonNmeaChunkSummary
	err = accumulator.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return withInputError(data, func(input *C.uint8_t, inputLength C.size_t) error {
			if inputLength != length {
				return errors.New("sidereon: inconsistent NMEA input length")
			}
			return callStatus(func() uint32 {
				return uint32(C.sidereon_nmea_accumulator_push((*C.SidereonNmeaAccumulator)(pointer), input, inputLength, &output))
			})
		})
	})
	runtime.KeepAlive(accumulator)
	runtime.KeepAlive(data)
	return nmeaChunkSummaryFromC(output), err
}

func (accumulator *NMEAAccumulator) Finish() (NMEAChunkSummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEAChunkSummary{}, ErrClosed
	}
	var output C.SidereonNmeaChunkSummary
	err := accumulator.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_nmea_accumulator_finish((*C.SidereonNmeaAccumulator)(pointer), &output))
		})
	})
	runtime.KeepAlive(accumulator)
	return nmeaChunkSummaryFromC(output), err
}

func (accumulator *NMEAAccumulator) Summary() (NMEASummary, error) {
	if accumulator == nil || accumulator.handle == nil {
		return NMEASummary{}, ErrClosed
	}
	var output C.SidereonNmeaSummary
	err := accumulator.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_nmea_accumulator_summary((*C.SidereonNmeaAccumulator)(pointer), &output))
		})
	})
	runtime.KeepAlive(accumulator)
	return NMEASummary{
		SentenceCount: uint64(output.sentence_count),
		EpochCount:    uint64(output.epoch_count),
		SkipCount:     uint64(output.skip_count),
		WarningCount:  uint64(output.warning_count),
	}, err
}

func (accumulator *NMEAAccumulator) RetainedLength() (uint64, error) {
	if accumulator == nil || accumulator.handle == nil {
		return 0, ErrClosed
	}
	var output C.size_t
	err := accumulator.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_nmea_accumulator_retained_len((*C.SidereonNmeaAccumulator)(pointer), &output))
		})
	})
	runtime.KeepAlive(accumulator)
	return uint64(output), err
}

func (accumulator *NMEAAccumulator) Epochs() ([]NMEAEpoch, error) {
	if accumulator == nil || accumulator.handle == nil {
		return nil, ErrClosed
	}
	var result []NMEAEpoch
	err := accumulator.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var written, required C.size_t
			status := C.sidereon_nmea_accumulator_epochs((*C.SidereonNmeaAccumulator)(pointer), nil, 0, &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			count, err := validateNativeQuery("NMEA accumulator epochs", uint64(written), uint64(required))
			if err != nil {
				return err
			}
			if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonNmeaEpochSummary{})); err != nil {
				return err
			}
			values := make([]C.SidereonNmeaEpochSummary, count)
			var output *C.SidereonNmeaEpochSummary
			if len(values) != 0 {
				output = &values[0]
			}
			status = C.sidereon_nmea_accumulator_epochs((*C.SidereonNmeaAccumulator)(pointer), output, C.size_t(len(values)), &written, &required)
			if err := statusErrorLocked(uint32(status)); err != nil {
				return err
			}
			writtenCount, err := validateNativeOutput("NMEA accumulator epochs", len(values), uint64(written), uint64(required))
			if err != nil {
				return err
			}
			result = make([]NMEAEpoch, writtenCount)
			for i := range result {
				result[i] = nmeaEpochFromC(&values[i])
			}
			return nil
		})
	})
	runtime.KeepAlive(accumulator)
	return result, err
}

func WriteNMEAGGA(options NMEAGGAOptions) ([]byte, error) {
	if len(options.Talker) != 2 {
		return nil, invalidArgument("NMEA talker must contain exactly two bytes")
	}
	if err := rejectEmbeddedNUL(options.Talker, "NMEA talker"); err != nil {
		return nil, err
	}
	var input C.SidereonNmeaGgaOptions
	input.talker[0] = C.char(options.Talker[0])
	input.talker[1] = C.char(options.Talker[1])
	input.talker[2] = 0
	input.utc_seconds_of_day = C.double(options.UTCSecondsOfDay)
	input.position = C.SidereonGeodetic{
		lat_rad:  C.double(options.Position.LatitudeRad),
		lon_rad:  C.double(options.Position.LongitudeRad),
		height_m: C.double(options.Position.HeightM),
	}
	input.quality = C.uint32_t(options.Quality)
	input.satellites_used = C.uint8_t(options.SatellitesUsed)
	input.hdop = C.double(options.HDOP)
	input.coordinate_decimals = C.uint8_t(options.CoordinateDecimals)
	return copyNativeBytes("NMEA GGA sentence", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_nmea_write_gga(&input, out, length, written, required)
	})
}
