package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// TrackCoordinateFrame identifies the metres-based axes used by TrackFilter.
type TrackCoordinateFrame uint32

const (
	TrackECEF                   TrackCoordinateFrame = TrackCoordinateFrame(native.TrackFrameECEFValue)
	TrackENU                    TrackCoordinateFrame = TrackCoordinateFrame(native.TrackFrameENUValue)
	TrackCallerDefinedCartesian TrackCoordinateFrame = TrackCoordinateFrame(native.TrackFrameCallerCartesianValue)
)

// TrackState describes a no-IMU filter state. Epoch is in the caller's
// monotonic seconds; positions are metres and velocities metres per second.
type TrackState struct {
	Frame                     TrackCoordinateFrame
	Epoch                     float64
	Dimension, StateDimension int
}
type TrackInnovation struct {
	Dimension int
	NIS       float64
}
type TrackPrediction struct {
	DT        float64
	Predicted TrackState
}
type TrackUpdate struct {
	Predicted, Updated TrackState
	Innovation         TrackInnovation
}
type TrackGatedUpdate struct {
	Gate      NISGate
	HasUpdate bool
	Update    TrackUpdate
	State     TrackState
}
type TrackRTSEpoch struct {
	Epoch              float64
	Predicted, Updated TrackState
	HasTransition      bool
}
type SmoothedTrackEpoch struct {
	Epoch            float64
	State            TrackState
	HasRTSGainToNext bool
}

type TrackFilterConfig struct {
	_      noCopy
	handle *native.TrackFilterConfig
}
type TrackFilter struct {
	_      noCopy
	handle *native.TrackFilter
}
type TrackRTSHistoryBuilder struct {
	_      noCopy
	handle *native.TrackRTSHistoryBuilder
}
type TrackRTSHistory struct {
	_      noCopy
	handle *native.TrackRTSHistory
}
type SmoothedTrack struct {
	_      noCopy
	handle *native.SmoothedTrack
}

func NewTrackFilterConfigFromPosition(frame TrackCoordinateFrame, epoch float64, position, positionCovariance []float64, initialVelocityVariance, accelerationVariance float64) (*TrackFilterConfig, error) {
	h, err := native.NewTrackConfigFromPosition(uint32(frame), epoch, append([]float64(nil), position...), append([]float64(nil), positionCovariance...), initialVelocityVariance, accelerationVariance)
	if err != nil {
		return nil, publicError(err)
	}
	return &TrackFilterConfig{handle: h}, nil
}
func NewTrackFilterConfigFromPositionVelocity(frame TrackCoordinateFrame, epoch float64, position, velocity, covariance []float64, accelerationVariance float64) (*TrackFilterConfig, error) {
	h, err := native.NewTrackConfigFromPositionVelocity(uint32(frame), epoch, append([]float64(nil), position...), append([]float64(nil), velocity...), append([]float64(nil), covariance...), accelerationVariance)
	if err != nil {
		return nil, publicError(err)
	}
	return &TrackFilterConfig{handle: h}, nil
}
func (c *TrackFilterConfig) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return publicError(c.handle.Close())
}
func (c *TrackFilterConfig) Dimension() (int, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	v, e := c.handle.Dimension()
	return v, publicError(e)
}
func (c *TrackFilterConfig) Frame() (TrackCoordinateFrame, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	v, e := c.handle.Frame()
	return TrackCoordinateFrame(v), publicError(e)
}
func NewTrackFilter(config *TrackFilterConfig) (*TrackFilter, error) {
	if config == nil || config.handle == nil {
		return nil, ErrClosed
	}
	h, e := native.NewTrackFilter(config.handle)
	if e != nil {
		return nil, publicError(e)
	}
	return &TrackFilter{handle: h}, nil
}
func NewTrackFilterFromPosition(frame TrackCoordinateFrame, epoch float64, position, positionCovariance []float64, initialVelocityVariance, accelerationVariance float64) (*TrackFilter, error) {
	h, e := native.NewTrackFilterFromPosition(uint32(frame), epoch, append([]float64(nil), position...), append([]float64(nil), positionCovariance...), initialVelocityVariance, accelerationVariance)
	if e != nil {
		return nil, publicError(e)
	}
	return &TrackFilter{handle: h}, nil
}

// NewTrackFilterFromPositionVelocity creates a constant-velocity filter from
// a position/velocity state. Positions are metres, velocities metres per
// second, and covariance is row-major over the 2*dimension state.
func NewTrackFilterFromPositionVelocity(frame TrackCoordinateFrame, epoch float64, position, velocity, covariance []float64, accelerationVariance float64) (*TrackFilter, error) {
	config, e := native.NewTrackConfigFromPositionVelocity(uint32(frame), epoch, append([]float64(nil), position...), append([]float64(nil), velocity...), append([]float64(nil), covariance...), accelerationVariance)
	if e != nil {
		return nil, publicError(e)
	}
	defer func() { _ = config.Close() }()
	h, e := native.NewTrackFilter(config)
	if e != nil {
		return nil, publicError(e)
	}
	return &TrackFilter{handle: h}, nil
}
func (f *TrackFilter) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return publicError(f.handle.Close())
}
func checkedTrackState(v native.NativeTrackState) (TrackState, error) {
	dimension, err := nativeCountToInt(v.Dimension, "track dimension")
	if err != nil {
		return TrackState{}, err
	}
	stateDimension, err := nativeCountToInt(v.StateDimension, "track state dimension")
	if err != nil {
		return TrackState{}, err
	}
	return TrackState{TrackCoordinateFrame(v.Frame), v.Epoch, dimension, stateDimension}, nil
}
func checkedTrackInnovation(v native.NativeTrackInnovation) (TrackInnovation, error) {
	dimension, err := nativeCountToInt(v.Dimension, "track innovation dimension")
	if err != nil {
		return TrackInnovation{}, err
	}
	return TrackInnovation{dimension, v.NIS}, nil
}
func checkedTrackUpdate(v native.NativeTrackUpdate) (TrackUpdate, error) {
	predicted, err := checkedTrackState(v.Predicted)
	if err != nil {
		return TrackUpdate{}, err
	}
	updated, err := checkedTrackState(v.Updated)
	if err != nil {
		return TrackUpdate{}, err
	}
	innovation, err := checkedTrackInnovation(v.Innovation)
	if err != nil {
		return TrackUpdate{}, err
	}
	return TrackUpdate{predicted, updated, innovation}, nil
}
func checkedTrackGate(v native.NativeTrackGatedUpdate) (TrackGatedUpdate, error) {
	update, err := checkedTrackUpdate(v.Update)
	if err != nil {
		return TrackGatedUpdate{}, err
	}
	state, err := checkedTrackState(v.State)
	if err != nil {
		return TrackGatedUpdate{}, err
	}
	return TrackGatedUpdate{NISGate{v.Gate.NIS, v.Gate.Threshold, v.Gate.InGate, v.Gate.DOF}, v.HasUpdate, update, state}, nil
}
func (f *TrackFilter) State() (TrackState, error) {
	if f == nil || f.handle == nil {
		return TrackState{}, ErrClosed
	}
	v, e := f.handle.State()
	state, conversionErr := checkedTrackState(v)
	if e != nil {
		return TrackState{}, publicError(e)
	}
	return state, conversionErr
}
func (f *TrackFilter) Predict(dt float64) (TrackPrediction, error) {
	if f == nil || f.handle == nil {
		return TrackPrediction{}, ErrClosed
	}
	v, e := f.handle.Predict(dt)
	state, conversionErr := checkedTrackState(v.Predicted)
	if e != nil {
		return TrackPrediction{}, publicError(e)
	}
	return TrackPrediction{v.DTS, state}, conversionErr
}
func (f *TrackFilter) Position() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Position()
	return append([]float64(nil), v...), publicError(e)
}
func (f *TrackFilter) Velocity() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Velocity()
	return append([]float64(nil), v...), publicError(e)
}
func (f *TrackFilter) StateVector() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.StateVector()
	return append([]float64(nil), v...), publicError(e)
}
func (f *TrackFilter) Covariance() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Covariance()
	return append([]float64(nil), v...), publicError(e)
}
func (f *TrackFilter) UpdatePosition(position, covariance []float64) (TrackUpdate, error) {
	if f == nil || f.handle == nil {
		return TrackUpdate{}, ErrClosed
	}
	v, e := f.handle.UpdatePosition(append([]float64(nil), position...), append([]float64(nil), covariance...))
	update, conversionErr := checkedTrackUpdate(v)
	if e != nil {
		return TrackUpdate{}, publicError(e)
	}
	return update, conversionErr
}
func (f *TrackFilter) UpdateState(state, covariance []float64) (TrackUpdate, error) {
	if f == nil || f.handle == nil {
		return TrackUpdate{}, ErrClosed
	}
	v, e := f.handle.UpdateState(append([]float64(nil), state...), append([]float64(nil), covariance...))
	update, conversionErr := checkedTrackUpdate(v)
	if e != nil {
		return TrackUpdate{}, publicError(e)
	}
	return update, conversionErr
}
func (f *TrackFilter) UpdatePositionGated(position, covariance []float64, confidence float64) (TrackGatedUpdate, error) {
	if f == nil || f.handle == nil {
		return TrackGatedUpdate{}, ErrClosed
	}
	v, e := f.handle.UpdatePositionGated(append([]float64(nil), position...), append([]float64(nil), covariance...), confidence)
	gate, conversionErr := checkedTrackGate(v)
	if e != nil {
		return TrackGatedUpdate{}, publicError(e)
	}
	return gate, conversionErr
}
func NewTrackRTSHistoryBuilder() (*TrackRTSHistoryBuilder, error) {
	h, e := native.NewTrackRTSHistoryBuilder()
	if e != nil {
		return nil, publicError(e)
	}
	return &TrackRTSHistoryBuilder{handle: h}, nil
}
func NewTrackRTSHistoryBuilderFromFilter(f *TrackFilter) (*TrackRTSHistoryBuilder, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	h, e := native.NewTrackRTSHistoryBuilderFromFilter(f.handle)
	if e != nil {
		return nil, publicError(e)
	}
	return &TrackRTSHistoryBuilder{handle: h}, nil
}
func (b *TrackRTSHistoryBuilder) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}
func (b *TrackRTSHistoryBuilder) Finish() (*TrackRTSHistory, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	h, e := b.handle.Finish()
	if e != nil {
		return nil, publicError(e)
	}
	return &TrackRTSHistory{handle: h}, nil
}
func (f *TrackFilter) PredictRecorded(dt float64, b *TrackRTSHistoryBuilder) (TrackPrediction, error) {
	if f == nil || f.handle == nil || b == nil || b.handle == nil {
		return TrackPrediction{}, ErrClosed
	}
	v, e := f.handle.PredictRecorded(dt, b.handle)
	state, conversionErr := checkedTrackState(v.Predicted)
	if e != nil {
		return TrackPrediction{}, publicError(e)
	}
	return TrackPrediction{v.DTS, state}, conversionErr
}
func (f *TrackFilter) RecordPredictionOnly(b *TrackRTSHistoryBuilder) error {
	if f == nil || f.handle == nil || b == nil || b.handle == nil {
		return ErrClosed
	}
	return publicError(f.handle.RecordPredictionOnly(b.handle))
}
func (f *TrackFilter) UpdatePositionRecorded(p, c []float64, b *TrackRTSHistoryBuilder) (TrackUpdate, error) {
	if f == nil || f.handle == nil || b == nil || b.handle == nil {
		return TrackUpdate{}, ErrClosed
	}
	v, e := f.handle.UpdatePositionRecorded(append([]float64(nil), p...), append([]float64(nil), c...), b.handle)
	update, conversionErr := checkedTrackUpdate(v)
	if e != nil {
		return TrackUpdate{}, publicError(e)
	}
	return update, conversionErr
}
func (f *TrackFilter) UpdatePositionGatedRecorded(p, c []float64, confidence float64, b *TrackRTSHistoryBuilder) (TrackGatedUpdate, error) {
	if f == nil || f.handle == nil || b == nil || b.handle == nil {
		return TrackGatedUpdate{}, ErrClosed
	}
	v, e := f.handle.UpdatePositionGatedRecorded(append([]float64(nil), p...), append([]float64(nil), c...), confidence, b.handle)
	gate, conversionErr := checkedTrackGate(v)
	if e != nil {
		return TrackGatedUpdate{}, publicError(e)
	}
	return gate, conversionErr
}
func (f *TrackFilter) PositionInnovation(position, covariance []float64) ([]float64, []float64, TrackInnovation, error) {
	if f == nil || f.handle == nil {
		return nil, nil, TrackInnovation{}, ErrClosed
	}
	a, b, v, e := native.TrackInnovation(f.handle, append([]float64(nil), position...), append([]float64(nil), covariance...), true)
	innovation, conversionErr := checkedTrackInnovation(v)
	if e != nil {
		return a, b, TrackInnovation{}, publicError(e)
	}
	return a, b, innovation, conversionErr
}
func (f *TrackFilter) StateInnovation(state, covariance []float64) ([]float64, []float64, TrackInnovation, error) {
	if f == nil || f.handle == nil {
		return nil, nil, TrackInnovation{}, ErrClosed
	}
	a, b, v, e := native.TrackInnovation(f.handle, append([]float64(nil), state...), append([]float64(nil), covariance...), false)
	innovation, conversionErr := checkedTrackInnovation(v)
	if e != nil {
		return a, b, TrackInnovation{}, publicError(e)
	}
	return a, b, innovation, conversionErr
}
func (h *TrackRTSHistory) Close() error {
	if h == nil || h.handle == nil {
		return nil
	}
	return publicError(h.handle.Close())
}
func (h *TrackRTSHistory) EpochCount() (int, error) {
	if h == nil || h.handle == nil {
		return 0, ErrClosed
	}
	v, e := h.handle.EpochCount()
	return v, publicError(e)
}
func (h *TrackRTSHistory) Epoch(index int) (TrackRTSEpoch, error) {
	if h == nil || h.handle == nil {
		return TrackRTSEpoch{}, ErrClosed
	}
	v, e := h.handle.Epoch(index)
	predicted, conversionErr := checkedTrackState(v.Predicted)
	if e != nil {
		return TrackRTSEpoch{}, publicError(e)
	}
	if conversionErr != nil {
		return TrackRTSEpoch{}, conversionErr
	}
	updated, conversionErr := checkedTrackState(v.Updated)
	if conversionErr != nil {
		return TrackRTSEpoch{}, conversionErr
	}
	return TrackRTSEpoch{v.Epoch, predicted, updated, v.HasTransition}, nil
}
func (h *TrackRTSHistory) PredictedPosition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Position(index, false)
	return append([]float64(nil), v...), publicError(e)
}
func (h *TrackRTSHistory) UpdatedPosition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Position(index, true)
	return append([]float64(nil), v...), publicError(e)
}
func (h *TrackRTSHistory) Transition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Transition(index)
	return append([]float64(nil), v...), publicError(e)
}
func SmoothTrackRTS(h *TrackRTSHistory) (*SmoothedTrack, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := native.SmoothTrackRTS(h.handle)
	if e != nil {
		return nil, publicError(e)
	}
	return &SmoothedTrack{handle: v}, nil
}
func (s *SmoothedTrack) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}
func (s *SmoothedTrack) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	v, e := s.handle.EpochCount()
	return v, publicError(e)
}
func (s *SmoothedTrack) Epoch(index int) (SmoothedTrackEpoch, error) {
	if s == nil || s.handle == nil {
		return SmoothedTrackEpoch{}, ErrClosed
	}
	v, e := s.handle.Epoch(index)
	state, conversionErr := checkedTrackState(v.State)
	if e != nil {
		return SmoothedTrackEpoch{}, publicError(e)
	}
	return SmoothedTrackEpoch{v.Epoch, state, v.HasGain}, conversionErr
}
func (s *SmoothedTrack) Position(index int) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(index, 0)
	return append([]float64(nil), v...), publicError(e)
}
func (s *SmoothedTrack) Covariance(index int) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(index, 1)
	return append([]float64(nil), v...), publicError(e)
}
func (s *SmoothedTrack) RTSGainToNext(index int) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(index, 2)
	return append([]float64(nil), v...), publicError(e)
}
