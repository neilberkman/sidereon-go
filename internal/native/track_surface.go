//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

func trackProduct(left, right int) (int, bool) {
	value, err := checkedProduct(left, right, "track matrix shape")
	return value, err == nil
}

type TrackFilter struct {
	_      noCopy
	handle *surfaceHandle
}
type TrackFilterConfig struct {
	_      noCopy
	handle *surfaceHandle
}
type TrackRTSHistoryBuilder struct {
	_      noCopy
	handle *surfaceHandle
}
type TrackRTSHistory struct {
	_      noCopy
	handle *surfaceHandle
}
type SmoothedTrack struct {
	_      noCopy
	handle *surfaceHandle
}

func newTrackHandle(pointer unsafe.Pointer, release func(unsafe.Pointer)) (*surfaceHandle, error) {
	return newSurfaceHandle(pointer, release)
}

func trackStatusRead(handle *surfaceHandle, call func(unsafe.Pointer) C.enum_SidereonStatus) error {
	return handle.read(func(pointer unsafe.Pointer) error {
		var err error
		withCThread(func() { err = statusErrorLocked(uint32(call(pointer))) })
		return err
	})
}

func trackStatusWrite(handle *surfaceHandle, call func(unsafe.Pointer) C.enum_SidereonStatus) error {
	return handle.write(func(pointer unsafe.Pointer) error {
		var err error
		withCThread(func() { err = statusErrorLocked(uint32(call(pointer))) })
		return err
	})
}

func newTrackFilter(pointer *C.SidereonTrackFilter) (*TrackFilter, error) {
	h, err := newTrackHandle(unsafe.Pointer(pointer), func(pointer unsafe.Pointer) { C.sidereon_track_filter_free((*C.SidereonTrackFilter)(pointer)) })
	if err != nil {
		return nil, err
	}
	return &TrackFilter{handle: h}, nil
}
func newTrackConfig(pointer *C.SidereonTrackFilterConfig) (*TrackFilterConfig, error) {
	h, err := newTrackHandle(unsafe.Pointer(pointer), func(pointer unsafe.Pointer) {
		C.sidereon_track_filter_config_free((*C.SidereonTrackFilterConfig)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &TrackFilterConfig{handle: h}, nil
}
func newTrackBuilder(pointer *C.SidereonTrackRtsHistoryBuilder) (*TrackRTSHistoryBuilder, error) {
	h, err := newTrackHandle(unsafe.Pointer(pointer), func(pointer unsafe.Pointer) {
		C.sidereon_track_rts_history_builder_free((*C.SidereonTrackRtsHistoryBuilder)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &TrackRTSHistoryBuilder{handle: h}, nil
}
func newTrackHistory(pointer *C.SidereonTrackRtsHistory) (*TrackRTSHistory, error) {
	h, err := newTrackHandle(unsafe.Pointer(pointer), func(pointer unsafe.Pointer) { C.sidereon_track_rts_history_free((*C.SidereonTrackRtsHistory)(pointer)) })
	if err != nil {
		return nil, err
	}
	return &TrackRTSHistory{handle: h}, nil
}
func newSmoothedTrack(pointer *C.SidereonSmoothedTrack) (*SmoothedTrack, error) {
	h, err := newTrackHandle(unsafe.Pointer(pointer), func(pointer unsafe.Pointer) { C.sidereon_smoothed_track_free((*C.SidereonSmoothedTrack)(pointer)) })
	if err != nil {
		return nil, err
	}
	return &SmoothedTrack{handle: h}, nil
}

func NewTrackConfigFromPosition(frame uint32, epoch float64, position, covariance []float64, initialVelocityVariance, accelerationVariance float64) (*TrackFilterConfig, error) {
	if len(position) == 0 {
		return nil, errors.New("sidereon: track position must not be empty")
	}
	dimension := len(position)
	if expected, ok := trackProduct(dimension, dimension); !ok || len(covariance) != expected {
		return nil, errors.New("sidereon: track position covariance shape does not match dimension")
	}
	p, _, err := cFloats(position, "track position")
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	c, covarianceLength, err := cFloats(covariance, "track position covariance")
	if err != nil {
		return nil, err
	}
	defer C.free(c)
	dimensionLength, err := cSize(dimension, "track dimension")
	if err != nil {
		return nil, err
	}
	var out *C.SidereonTrackFilterConfig
	var opErr error
	withCThread(func() {
		opErr = statusErrorLocked(uint32(C.sidereon_track_filter_config_from_position(C.uint32_t(frame), C.double(epoch), (*C.double)(p), dimensionLength, (*C.double)(c), covarianceLength, C.double(initialVelocityVariance), C.double(accelerationVariance), &out)))
	})
	if opErr != nil {
		return nil, opErr
	}
	return newTrackConfig(out)
}

func NewTrackConfigFromPositionVelocity(frame uint32, epoch float64, position, velocity, covariance []float64, accelerationVariance float64) (*TrackFilterConfig, error) {
	if len(position) == 0 || len(position) != len(velocity) {
		return nil, errors.New("sidereon: track position and velocity dimensions differ")
	}
	dimension := len(position)
	maxInt := int(^uint(0) >> 1)
	if dimension > maxInt/2 {
		return nil, errors.New("sidereon: track state dimension overflows")
	}
	stateDimension := 2 * dimension
	expected, ok := trackProduct(stateDimension, stateDimension)
	if !ok || len(covariance) != expected {
		return nil, errors.New("sidereon: track covariance shape does not match dimension")
	}
	p, dimensionLength, err := cFloats(position, "track position")
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	v, _, err := cFloats(velocity, "track velocity")
	if err != nil {
		return nil, err
	}
	defer C.free(v)
	c, covarianceLength, err := cFloats(covariance, "track covariance")
	if err != nil {
		return nil, err
	}
	defer C.free(c)
	var out *C.SidereonTrackFilterConfig
	var opErr error
	withCThread(func() {
		opErr = statusErrorLocked(uint32(C.sidereon_track_filter_config_from_position_velocity(C.uint32_t(frame), C.double(epoch), (*C.double)(p), (*C.double)(v), dimensionLength, (*C.double)(c), covarianceLength, C.double(accelerationVariance), &out)))
	})
	if opErr != nil {
		return nil, opErr
	}
	return newTrackConfig(out)
}

func NewTrackFilter(config *TrackFilterConfig) (*TrackFilter, error) {
	if config == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonTrackFilter
	var opErr error
	if err := config.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_track_filter_new((*C.SidereonTrackFilterConfig)(pointer), &out)))
		})
		return nil
	}); err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	return newTrackFilter(out)
}

func NewTrackFilterFromPosition(frame uint32, epoch float64, position, covariance []float64, initialVelocityVariance, accelerationVariance float64) (*TrackFilter, error) {
	if len(position) == 0 {
		return nil, errors.New("sidereon: track position must not be empty")
	}
	dimension := len(position)
	expected, ok := trackProduct(dimension, dimension)
	if !ok || len(covariance) != expected {
		return nil, errors.New("sidereon: track covariance shape does not match dimension")
	}
	p, dimensionLength, err := cFloats(position, "track position")
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	c, covarianceLength, err := cFloats(covariance, "track covariance")
	if err != nil {
		return nil, err
	}
	defer C.free(c)
	var out *C.SidereonTrackFilter
	var opErr error
	withCThread(func() {
		opErr = statusErrorLocked(uint32(C.sidereon_track_filter_new_from_position(C.uint32_t(frame), C.double(epoch), (*C.double)(p), dimensionLength, (*C.double)(c), covarianceLength, C.double(initialVelocityVariance), C.double(accelerationVariance), &out)))
	})
	if opErr != nil {
		return nil, opErr
	}
	return newTrackFilter(out)
}

type NativeTrackState struct {
	Frame                     uint32
	Epoch                     float64
	Dimension, StateDimension uint64
}
type NativeTrackInnovation struct {
	Dimension uint64
	NIS       float64
}
type NativeTrackPrediction struct {
	DTS       float64
	Predicted NativeTrackState
}
type NativeTrackUpdate struct {
	Predicted, Updated NativeTrackState
	Innovation         NativeTrackInnovation
}
type NativeTrackGatedUpdate struct {
	Gate      NISGate
	HasUpdate bool
	Update    NativeTrackUpdate
	State     NativeTrackState
}
type NativeTrackRTSEpoch struct {
	Epoch              float64
	Predicted, Updated NativeTrackState
	HasTransition      bool
}
type NativeSmoothedTrackEpoch struct {
	Epoch   float64
	State   NativeTrackState
	HasGain bool
}

func trackState(value C.SidereonTrackState) NativeTrackState {
	return NativeTrackState{uint32(value.frame), float64(value.t_s), uint64(value.dimension), uint64(value.state_dimension)}
}
func trackInnovation(value C.SidereonTrackInnovation) NativeTrackInnovation {
	return NativeTrackInnovation{uint64(value.dimension), float64(value.nis)}
}
func trackUpdate(value C.SidereonTrackUpdate) NativeTrackUpdate {
	return NativeTrackUpdate{trackState(value.predicted), trackState(value.updated), trackInnovation(value.innovation)}
}
func trackGate(value C.SidereonTrackGatedUpdate) NativeTrackGatedUpdate {
	return NativeTrackGatedUpdate{NISGate{float64(value.gate.nis), float64(value.gate.threshold), bool(value.gate.in_gate), uint64(value.gate.dof)}, bool(value.has_update), trackUpdate(value.update), trackState(value.state)}
}

func (c *TrackFilterConfig) Close() error {
	if c == nil {
		return nil
	}
	return c.handle.close()
}
func (c *TrackFilterConfig) Dimension() (int, error) {
	var out C.size_t
	err := trackStatusRead(c.handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_filter_config_dimension((*C.SidereonTrackFilterConfig)(pointer), &out)
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(out, "track dimension")
}
func (c *TrackFilterConfig) Frame() (uint32, error) {
	var out C.uint32_t
	err := trackStatusRead(c.handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_filter_config_frame((*C.SidereonTrackFilterConfig)(pointer), &out)
	})
	return uint32(out), err
}
func (f *TrackFilter) Close() error {
	if f == nil {
		return nil
	}
	return f.handle.close()
}
func (b *TrackRTSHistoryBuilder) Close() error {
	if b == nil {
		return nil
	}
	return b.handle.close()
}
func (h *TrackRTSHistory) Close() error {
	if h == nil {
		return nil
	}
	return h.handle.close()
}
func (s *SmoothedTrack) Close() error {
	if s == nil {
		return nil
	}
	return s.handle.close()
}

func (f *TrackFilter) State() (NativeTrackState, error) {
	var out C.SidereonTrackState
	err := trackStatusRead(f.handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_filter_state((*C.SidereonTrackFilter)(pointer), &out)
	})
	return trackState(out), err
}
func (f *TrackFilter) Predict(dt float64) (NativeTrackPrediction, error) {
	var out C.SidereonTrackPrediction
	err := trackStatusWrite(f.handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_filter_predict((*C.SidereonTrackFilter)(pointer), C.double(dt), &out)
	})
	return NativeTrackPrediction{float64(out.dt_s), trackState(out.predicted)}, err
}
func (f *TrackFilter) PredictRecorded(dt float64, history *TrackRTSHistoryBuilder) (NativeTrackPrediction, error) {
	if history == nil {
		return NativeTrackPrediction{}, ErrClosed
	}
	var out C.SidereonTrackPrediction
	var opErr error
	err := f.handle.write(func(pointer unsafe.Pointer) error {
		return history.handle.read(func(hp unsafe.Pointer) error {
			withCThread(func() {
				opErr = statusErrorLocked(uint32(C.sidereon_track_filter_predict_recorded((*C.SidereonTrackFilter)(pointer), C.double(dt), (*C.SidereonTrackRtsHistoryBuilder)(hp), &out)))
			})
			return nil
		})
	})
	if err != nil {
		return NativeTrackPrediction{}, err
	}
	return NativeTrackPrediction{float64(out.dt_s), trackState(out.predicted)}, opErr
}
func (f *TrackFilter) RecordPredictionOnly(history *TrackRTSHistoryBuilder) error {
	if history == nil {
		return ErrClosed
	}
	var opErr error
	err := f.handle.read(func(pointer unsafe.Pointer) error {
		return history.handle.write(func(hp unsafe.Pointer) error {
			withCThread(func() {
				opErr = statusErrorLocked(uint32(C.sidereon_track_filter_record_prediction_only((*C.SidereonTrackFilter)(pointer), (*C.SidereonTrackRtsHistoryBuilder)(hp))))
			})
			return nil
		})
	})
	if err != nil {
		return err
	}
	return opErr
}

func copyTrackDoubles(handle *surfaceHandle, label string, call func(unsafe.Pointer, *C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]float64, error) {
	return copySurfaceDoubles(handle, label, call)
}
func (f *TrackFilter) Position() ([]float64, error) {
	return copyTrackDoubles(f.handle, "track position", func(pointer unsafe.Pointer, out *C.double, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_track_filter_position_m((*C.SidereonTrackFilter)(pointer), out, len, written, required)
	})
}
func (f *TrackFilter) Velocity() ([]float64, error) {
	return copyTrackDoubles(f.handle, "track velocity", func(pointer unsafe.Pointer, out *C.double, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_track_filter_velocity_m_s((*C.SidereonTrackFilter)(pointer), out, len, written, required)
	})
}
func (f *TrackFilter) StateVector() ([]float64, error) {
	return copyTrackDoubles(f.handle, "track state vector", func(pointer unsafe.Pointer, out *C.double, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_track_filter_state_vector((*C.SidereonTrackFilter)(pointer), out, len, written, required)
	})
}
func (f *TrackFilter) Covariance() ([]float64, error) {
	return copyTrackDoubles(f.handle, "track covariance", func(pointer unsafe.Pointer, out *C.double, len C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_track_filter_covariance((*C.SidereonTrackFilter)(pointer), out, len, written, required)
	})
}

func trackInput(position, covariance []float64) (unsafe.Pointer, C.size_t, unsafe.Pointer, C.size_t, error) {
	p, positionLength, err := cFloats(position, "track measurement")
	if err != nil {
		return nil, 0, nil, 0, err
	}
	c, covarianceLength, err := cFloats(covariance, "track measurement covariance")
	if err != nil {
		C.free(p)
		return nil, 0, nil, 0, err
	}
	return p, positionLength, c, covarianceLength, nil
}
func (f *TrackFilter) UpdatePosition(position, covariance []float64) (NativeTrackUpdate, error) {
	if len(position) == 0 {
		return NativeTrackUpdate{}, errors.New("sidereon: track measurement must not be empty")
	}
	expected, ok := trackProduct(len(position), len(position))
	if !ok || len(covariance) != expected {
		return NativeTrackUpdate{}, errors.New("sidereon: track measurement covariance shape does not match dimension")
	}
	p, positionLength, c, covarianceLength, err := trackInput(position, covariance)
	if err != nil {
		return NativeTrackUpdate{}, err
	}
	defer C.free(p)
	defer C.free(c)
	var out C.SidereonTrackUpdate
	err = trackStatusWrite(f.handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_filter_update_position((*C.SidereonTrackFilter)(pointer), (*C.double)(p), positionLength, (*C.double)(c), covarianceLength, &out)
	})
	return trackUpdate(out), err
}
func (f *TrackFilter) UpdateState(state, covariance []float64) (NativeTrackUpdate, error) {
	if len(state) == 0 {
		return NativeTrackUpdate{}, errors.New("sidereon: track state must not be empty")
	}
	expected, ok := trackProduct(len(state), len(state))
	if !ok || len(covariance) != expected {
		return NativeTrackUpdate{}, errors.New("sidereon: track state covariance shape does not match dimension")
	}
	p, stateLength, c, covarianceLength, err := trackInput(state, covariance)
	if err != nil {
		return NativeTrackUpdate{}, err
	}
	defer C.free(p)
	defer C.free(c)
	var out C.SidereonTrackUpdate
	err = trackStatusWrite(f.handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_filter_update_state((*C.SidereonTrackFilter)(pointer), (*C.double)(p), stateLength, (*C.double)(c), covarianceLength, &out)
	})
	return trackUpdate(out), err
}
func (f *TrackFilter) UpdatePositionGated(position, covariance []float64, confidence float64) (NativeTrackGatedUpdate, error) {
	if len(position) == 0 {
		return NativeTrackGatedUpdate{}, errors.New("sidereon: track measurement must not be empty")
	}
	expected, ok := trackProduct(len(position), len(position))
	if !ok || len(covariance) != expected {
		return NativeTrackGatedUpdate{}, errors.New("sidereon: track measurement covariance shape does not match dimension")
	}
	p, positionLength, c, covarianceLength, err := trackInput(position, covariance)
	if err != nil {
		return NativeTrackGatedUpdate{}, err
	}
	defer C.free(p)
	defer C.free(c)
	var out C.SidereonTrackGatedUpdate
	err = trackStatusWrite(f.handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_filter_update_position_gated((*C.SidereonTrackFilter)(pointer), (*C.double)(p), positionLength, (*C.double)(c), covarianceLength, C.double(confidence), &out)
	})
	return trackGate(out), err
}
func (f *TrackFilter) UpdatePositionRecorded(position, covariance []float64, history *TrackRTSHistoryBuilder) (NativeTrackUpdate, error) {
	if history == nil {
		return NativeTrackUpdate{}, ErrClosed
	}
	if len(position) == 0 {
		return NativeTrackUpdate{}, errors.New("sidereon: track measurement must not be empty")
	}
	expected, ok := trackProduct(len(position), len(position))
	if !ok || len(covariance) != expected {
		return NativeTrackUpdate{}, errors.New("sidereon: track measurement covariance shape does not match dimension")
	}
	p, positionLength, c, covarianceLength, err := trackInput(position, covariance)
	if err != nil {
		return NativeTrackUpdate{}, err
	}
	defer C.free(p)
	defer C.free(c)
	var out C.SidereonTrackUpdate
	var opErr error
	err = f.handle.write(func(pointer unsafe.Pointer) error {
		return history.handle.write(func(hp unsafe.Pointer) error {
			withCThread(func() {
				opErr = statusErrorLocked(uint32(C.sidereon_track_filter_update_position_recorded((*C.SidereonTrackFilter)(pointer), (*C.double)(p), positionLength, (*C.double)(c), covarianceLength, (*C.SidereonTrackRtsHistoryBuilder)(hp), &out)))
			})
			return nil
		})
	})
	if err != nil {
		return NativeTrackUpdate{}, err
	}
	return trackUpdate(out), opErr
}

func (f *TrackFilter) UpdatePositionGatedRecorded(position, covariance []float64, confidence float64, history *TrackRTSHistoryBuilder) (NativeTrackGatedUpdate, error) {
	if history == nil {
		return NativeTrackGatedUpdate{}, ErrClosed
	}
	if len(position) == 0 {
		return NativeTrackGatedUpdate{}, errors.New("sidereon: track measurement must not be empty")
	}
	expected, ok := trackProduct(len(position), len(position))
	if !ok || len(covariance) != expected {
		return NativeTrackGatedUpdate{}, errors.New("sidereon: track measurement covariance shape does not match dimension")
	}
	p, positionLength, c, covarianceLength, err := trackInput(position, covariance)
	if err != nil {
		return NativeTrackGatedUpdate{}, err
	}
	defer C.free(p)
	defer C.free(c)
	var out C.SidereonTrackGatedUpdate
	var opErr error
	err = f.handle.write(func(pointer unsafe.Pointer) error {
		return history.handle.write(func(hp unsafe.Pointer) error {
			withCThread(func() {
				opErr = statusErrorLocked(uint32(C.sidereon_track_filter_update_position_gated_recorded((*C.SidereonTrackFilter)(pointer), (*C.double)(p), positionLength, (*C.double)(c), covarianceLength, C.double(confidence), (*C.SidereonTrackRtsHistoryBuilder)(hp), &out)))
			})
			return nil
		})
	})
	if err != nil {
		return NativeTrackGatedUpdate{}, err
	}
	return trackGate(out), opErr
}

func NewTrackRTSHistoryBuilder() (*TrackRTSHistoryBuilder, error) {
	var out *C.SidereonTrackRtsHistoryBuilder
	var err error
	withCThread(func() { err = statusErrorLocked(uint32(C.sidereon_track_rts_history_builder_new(&out))) })
	if err != nil {
		return nil, err
	}
	return newTrackBuilder(out)
}
func NewTrackRTSHistoryBuilderFromFilter(filter *TrackFilter) (*TrackRTSHistoryBuilder, error) {
	if filter == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonTrackRtsHistoryBuilder
	var opErr error
	err := filter.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_track_rts_history_builder_from_filter((*C.SidereonTrackFilter)(pointer), &out)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	return newTrackBuilder(out)
}
func (b *TrackRTSHistoryBuilder) Finish() (*TrackRTSHistory, error) {
	if b == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonTrackRtsHistory
	var opErr error
	err := b.handle.read(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_track_rts_history_builder_finish((*C.SidereonTrackRtsHistoryBuilder)(pointer), &out)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	return newTrackHistory(out)
}
func historyCount(handle *surfaceHandle, call func(unsafe.Pointer, *C.size_t) C.enum_SidereonStatus) (int, error) {
	var out C.size_t
	err := trackStatusRead(handle, func(pointer unsafe.Pointer) C.enum_SidereonStatus { return call(pointer, &out) })
	if err != nil {
		return 0, err
	}
	return sizeTToInt(out, "history epoch count")
}
func (h *TrackRTSHistory) EpochCount() (int, error) {
	return historyCount(h.handle, func(p unsafe.Pointer, o *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_track_rts_history_epoch_count((*C.SidereonTrackRtsHistory)(p), o)
	})
}
func (h *TrackRTSHistory) Epoch(index int) (NativeTrackRTSEpoch, error) {
	if index < 0 {
		return NativeTrackRTSEpoch{}, errors.New("sidereon: negative history index")
	}
	var out C.SidereonTrackRtsEpoch
	err := trackStatusRead(h.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_track_rts_history_epoch((*C.SidereonTrackRtsHistory)(p), C.size_t(index), &out)
	})
	return NativeTrackRTSEpoch{float64(out.t_s), trackState(out.predicted), trackState(out.updated), bool(out.has_transition_from_previous)}, err
}
func (h *TrackRTSHistory) Position(index int, updated bool) ([]float64, error) {
	if index < 0 {
		return nil, errors.New("sidereon: negative history index")
	}
	return copyTrackDoubles(h.handle, "history position", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		if updated {
			return C.sidereon_track_rts_history_epoch_updated_position_m((*C.SidereonTrackRtsHistory)(p), C.size_t(index), o, l, w, r)
		}
		return C.sidereon_track_rts_history_epoch_predicted_position_m((*C.SidereonTrackRtsHistory)(p), C.size_t(index), o, l, w, r)
	})
}
func (h *TrackRTSHistory) Transition(index int) ([]float64, error) {
	if index < 0 {
		return nil, errors.New("sidereon: negative history index")
	}
	return copyTrackDoubles(h.handle, "history transition", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_track_rts_history_epoch_transition_from_previous((*C.SidereonTrackRtsHistory)(p), C.size_t(index), o, l, w, r)
	})
}
func SmoothTrackRTS(history *TrackRTSHistory) (*SmoothedTrack, error) {
	if history == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonSmoothedTrack
	var opErr error
	err := history.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_smooth_track_rts((*C.SidereonTrackRtsHistory)(p), &out)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	return newSmoothedTrack(out)
}
func (s *SmoothedTrack) EpochCount() (int, error) {
	return historyCount(s.handle, func(p unsafe.Pointer, o *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_smoothed_track_epoch_count((*C.SidereonSmoothedTrack)(p), o)
	})
}
func (s *SmoothedTrack) Epoch(index int) (NativeSmoothedTrackEpoch, error) {
	if index < 0 {
		return NativeSmoothedTrackEpoch{}, errors.New("sidereon: negative smoothed index")
	}
	var out C.SidereonSmoothedTrackEpoch
	err := trackStatusRead(s.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_smoothed_track_epoch((*C.SidereonSmoothedTrack)(p), C.size_t(index), &out)
	})
	return NativeSmoothedTrackEpoch{float64(out.t_s), trackState(out.state), bool(out.has_rts_gain_to_next)}, err
}
func (s *SmoothedTrack) Values(index int, kind uint32) ([]float64, error) {
	if index < 0 {
		return nil, errors.New("sidereon: negative smoothed index")
	}
	return copyTrackDoubles(s.handle, "smoothed track values", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		switch kind {
		case 0:
			return C.sidereon_smoothed_track_epoch_position_m((*C.SidereonSmoothedTrack)(p), C.size_t(index), o, l, w, r)
		case 1:
			return C.sidereon_smoothed_track_epoch_covariance((*C.SidereonSmoothedTrack)(p), C.size_t(index), o, l, w, r)
		default:
			return C.sidereon_smoothed_track_epoch_rts_gain_to_next((*C.SidereonSmoothedTrack)(p), C.size_t(index), o, l, w, r)
		}
	})
}

func TrackInnovation(filter *TrackFilter, state, covariance []float64, positionOnly bool) ([]float64, []float64, NativeTrackInnovation, error) {
	if filter == nil {
		return nil, nil, NativeTrackInnovation{}, ErrClosed
	}
	if len(state) == 0 {
		return nil, nil, NativeTrackInnovation{}, errors.New("sidereon: innovation state must not be empty")
	}
	expected, ok := trackProduct(len(state), len(state))
	if !ok || len(covariance) != expected {
		return nil, nil, NativeTrackInnovation{}, errors.New("sidereon: innovation covariance shape does not match dimension")
	}
	p, stateLength, c, covarianceLength, err := trackInput(state, covariance)
	if err != nil {
		return nil, nil, NativeTrackInnovation{}, err
	}
	defer C.free(p)
	defer C.free(c)
	innovation := make([]C.double, len(state))
	innovationCov := make([]C.double, len(covariance))
	innovationLength, err := cSize(len(innovation), "innovation count")
	if err != nil {
		return nil, nil, NativeTrackInnovation{}, err
	}
	innovationCovLength, err := cSize(len(innovationCov), "innovation covariance count")
	if err != nil {
		return nil, nil, NativeTrackInnovation{}, err
	}
	var report C.SidereonTrackInnovation
	var opErr error
	err = filter.handle.read(func(fp unsafe.Pointer) error {
		withCThread(func() {
			if positionOnly {
				opErr = statusErrorLocked(uint32(C.sidereon_track_filter_position_innovation((*C.SidereonTrackFilter)(fp), (*C.double)(p), stateLength, (*C.double)(c), covarianceLength, &innovation[0], innovationLength, &innovationCov[0], innovationCovLength, &report)))
			} else {
				opErr = statusErrorLocked(uint32(C.sidereon_track_filter_state_innovation((*C.SidereonTrackFilter)(fp), (*C.double)(p), stateLength, (*C.double)(c), covarianceLength, &innovation[0], innovationLength, &innovationCov[0], innovationCovLength, &report)))
			}
		})
		return nil
	})
	if err != nil {
		return nil, nil, NativeTrackInnovation{}, err
	}
	if opErr != nil {
		return nil, nil, NativeTrackInnovation{}, opErr
	}
	a := make([]float64, len(innovation))
	b := make([]float64, len(innovationCov))
	for i := range a {
		a[i] = float64(innovation[i])
	}
	for i := range b {
		b[i] = float64(innovationCov[i])
	}
	return a, b, trackInnovation(report), nil
}
