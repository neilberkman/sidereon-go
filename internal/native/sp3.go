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

type SP3State struct {
	PositionM      [3]float64
	HasClock       bool
	ClockS         float64
	HasVelocity    bool
	VelocityMPerS  [3]float64
	HasClockRate   bool
	ClockRateSPerS float64
	ClockEvent     bool
	ClockPredicted bool
	Maneuver       bool
	OrbitPredicted bool
}

type SP3PredictionSummary struct {
	EpochCount             int
	ObservedThroughPresent bool
	ObservedThroughJ2000S  float64
}

type SP3 struct {
	handle *positioningHandle
}

func releaseSP3(pointer unsafe.Pointer) {
	C.sidereon_sp3_free((*C.SidereonSp3)(pointer))
}

func LoadSP3(data []byte) (*SP3, error) {
	var pointer *C.SidereonSp3
	var err error
	withCThread(func() {
		var cdata unsafe.Pointer
		if len(data) != 0 {
			cdata = C.CBytes(data)
			if cdata == nil {
				err = errors.New("sidereon: unable to allocate native input buffer")
				return
			}
			defer C.free(cdata)
		}
		err = statusErrorLocked(C.sidereon_sp3_load(
			(*C.uint8_t)(cdata), C.size_t(len(data)), &pointer,
		))
	})
	if err != nil {
		return nil, err
	}
	if pointer == nil {
		return nil, errors.New("sidereon: native SP3 load returned no handle")
	}
	return &SP3{handle: newPositioningHandle(unsafe.Pointer(pointer), releaseSP3)}, nil
}

func (s *SP3) Close() error {
	if s == nil {
		return nil
	}
	return s.handle.close()
}

func (s *SP3) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	var count C.size_t
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sp3_epoch_count((*C.SidereonSp3)(pointer), &count)
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(count))
}

func (s *SP3) Epochs() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var values []C.double
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_sp3_epochs_j2000_seconds(
				(*C.SidereonSp3)(pointer), nil, 0, &written, &required,
			)
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		values = make([]C.double, count)
		var output *C.double
		if len(values) != 0 {
			output = &values[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_sp3_epochs_j2000_seconds(
				(*C.SidereonSp3)(pointer), output, C.size_t(len(values)), &written, &required,
			)
		}); err != nil {
			return err
		}
		writtenCount, err := validateTwoPassCounts(
			"SP3 epochs", len(values), count, uint64(written), uint64(required),
		)
		if err != nil {
			return err
		}
		values = values[:writtenCount]
		return nil
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	out := make([]float64, len(values))
	for i := range values {
		out[i] = float64(values[i])
	}
	return out, nil
}

func (s *SP3) Satellites() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var tokens []C.SidereonSatelliteToken
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		if err := callStatus(func() uint32 {
			return C.sidereon_sp3_satellites(
				(*C.SidereonSp3)(pointer), nil, 0, &written, &required,
			)
		}); err != nil {
			return err
		}
		count, err := checkedNativeCount(uint64(required))
		if err != nil {
			return err
		}
		tokens = make([]C.SidereonSatelliteToken, count)
		var output *C.SidereonSatelliteToken
		if len(tokens) != 0 {
			output = &tokens[0]
		}
		if err := callStatus(func() uint32 {
			return C.sidereon_sp3_satellites(
				(*C.SidereonSp3)(pointer), output, C.size_t(len(tokens)), &written, &required,
			)
		}); err != nil {
			return err
		}
		writtenCount, err := validateTwoPassCounts(
			"SP3 satellites", len(tokens), count, uint64(written), uint64(required),
		)
		if err != nil {
			return err
		}
		tokens = tokens[:writtenCount]
		return nil
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(tokens))
	for i := range tokens {
		out[i] = tokenFromC(tokens[i])
	}
	return out, nil
}

func (s *SP3) State(satelliteID string, epochIndex int) (SP3State, error) {
	if s == nil || s.handle == nil {
		return SP3State{}, ErrClosed
	}
	if epochIndex < 0 {
		return SP3State{}, errors.New("sidereon: epoch index must not be negative")
	}
	if err := rejectEmbeddedNUL(satelliteID, "SP3 satellite ID"); err != nil {
		return SP3State{}, err
	}
	cid := C.CBytes(append([]byte(satelliteID), 0))
	if cid == nil {
		return SP3State{}, errors.New("sidereon: unable to allocate native satellite id")
	}
	defer C.free(cid)
	var state C.SidereonSp3State
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sp3_state(
				(*C.SidereonSp3)(pointer), (*C.char)(cid), C.size_t(epochIndex), &state,
			)
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return SP3State{}, err
	}
	var out SP3State
	for i := 0; i < 3; i++ {
		out.PositionM[i] = float64(state.position_m[i])
		out.VelocityMPerS[i] = float64(state.velocity_m_s[i])
	}
	out.HasClock = bool(state.has_clock_s)
	out.ClockS = float64(state.clock_s)
	out.HasVelocity = bool(state.has_velocity_m_s)
	out.HasClockRate = bool(state.has_clock_rate_s_s)
	out.ClockRateSPerS = float64(state.clock_rate_s_s)
	out.ClockEvent = bool(state.clock_event)
	out.ClockPredicted = bool(state.clock_predicted)
	out.Maneuver = bool(state.maneuver)
	out.OrbitPredicted = bool(state.orbit_predicted)
	return out, nil
}

func (s *SP3) PredictionSummary() (SP3PredictionSummary, error) {
	if s == nil || s.handle == nil {
		return SP3PredictionSummary{}, ErrClosed
	}
	var summary C.SidereonSp3PredictionSummary
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_sp3_prediction_summary((*C.SidereonSp3)(pointer), &summary)
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return SP3PredictionSummary{}, err
	}
	count, err := checkedNativeCount(uint64(summary.epoch_count))
	if err != nil {
		return SP3PredictionSummary{}, err
	}
	return SP3PredictionSummary{
		EpochCount:             count,
		ObservedThroughPresent: bool(summary.observed_through_present),
		ObservedThroughJ2000S:  float64(summary.observed_through_j2000_seconds),
	}, nil
}
