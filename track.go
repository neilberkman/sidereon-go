package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// TrackCoordinateFrame identifies the Cartesian frame used by the tracking state.
type TrackCoordinateFrame uint32

const (
	// TrackECEF selects the Earth-centered, Earth-fixed (ECEF) frame.
	TrackECEF TrackCoordinateFrame = TrackCoordinateFrame(native.TrackFrameECEFValue)
	// TrackENU selects the local east-north-up (ENU) frame.
	TrackENU TrackCoordinateFrame = TrackCoordinateFrame(native.TrackFrameENUValue)
	// TrackCallerDefinedCartesian selects the caller-defined Cartesian frame.
	TrackCallerDefinedCartesian TrackCoordinateFrame = TrackCoordinateFrame(native.TrackFrameCallerCartesianValue)
)

// TrackState describes a no-IMU filter state. Epoch is in the caller's
// monotonic seconds; positions are metres and velocities metres per second.
// TrackState contains the current position/velocity state dimensions and caller-monotonic epoch.
type TrackState struct {
	// Frame identifies the coordinate frame for the values.
	Frame                     TrackCoordinateFrame
	Epoch                     float64
	Dimension, StateDimension int
}

// TrackInnovation contains innovation residual, covariance, and normalized-innovation diagnostics.
type TrackInnovation struct {
	Dimension int
	NIS       float64
}

// TrackPrediction contains the predicted tracking state and covariance.
type TrackPrediction struct {
	// DT is the prediction interval in seconds.
	DT float64
	// Predicted is the predicted state.
	Predicted TrackState
}

// TrackUpdate contains the updated tracking state and covariance.
type TrackUpdate struct {
	// Predicted and Updated are the predicted state, the updated state.
	Predicted, Updated TrackState
	// Innovation is the measurement innovation.
	Innovation TrackInnovation
}

// TrackGatedUpdate contains the gated update decision and its optional update/state.
type TrackGatedUpdate struct {
	// Gate is the NIS gate decision.
	Gate      NISGate
	HasUpdate bool
	Update    TrackUpdate
	State     TrackState
}

// TrackRTSEpoch contains one recorded RTS epoch and transition metadata.
type TrackRTSEpoch struct {
	Epoch float64
	// Predicted and Updated are the predicted state, the updated state.
	Predicted, Updated TrackState
	HasTransition      bool
}

// SmoothedTrackEpoch contains one RTS-smoothed state, covariance, and transition metadata.
type SmoothedTrackEpoch struct {
	Epoch            float64
	State            TrackState
	HasRTSGainToNext bool
}

// TrackFilterConfig owns the native filter configuration; Close releases it and readers are synchronized.
type TrackFilterConfig struct {
	_      noCopy
	handle *native.TrackFilterConfig
}

// TrackFilter owns the native tracking filter; Close synchronizes with prediction, update, and readers.
type TrackFilter struct {
	_      noCopy
	handle *native.TrackFilter
}

// TrackRTSHistoryBuilder owns the native RTS history recorder until Finish or Close.
type TrackRTSHistoryBuilder struct {
	_      noCopy
	handle *native.TrackRTSHistoryBuilder
}

// TrackRTSHistory owns recorded filter history and provides detached epoch readers.
type TrackRTSHistory struct {
	_      noCopy
	handle *native.TrackRTSHistory
}

// SmoothedTrack owns the native RTS-smoothed track and its detached epoch results.
type SmoothedTrack struct {
	_      noCopy
	handle *native.SmoothedTrack
}

// NewTrackFilterConfigFromPosition creates a native filter configuration for position-only state.
func NewTrackFilterConfigFromPosition(frame TrackCoordinateFrame, epoch float64, position, positionCovariance []float64, initialVelocityVariance, accelerationVariance float64) (*TrackFilterConfig, error) {
	h, err := native.NewTrackConfigFromPosition(uint32(frame), epoch, append([]float64(nil), position...), append([]float64(nil), positionCovariance...), initialVelocityVariance, accelerationVariance)
	if err != nil {
		return nil, publicError(err)
	}
	return &TrackFilterConfig{handle: h}, nil
}

// NewTrackFilterConfigFromPositionVelocity creates a native filter configuration for position-and-velocity state.
func NewTrackFilterConfigFromPositionVelocity(frame TrackCoordinateFrame, epoch float64, position, velocity, covariance []float64, accelerationVariance float64) (*TrackFilterConfig, error) {
	h, err := native.NewTrackConfigFromPositionVelocity(uint32(frame), epoch, append([]float64(nil), position...), append([]float64(nil), velocity...), append([]float64(nil), covariance...), accelerationVariance)
	if err != nil {
		return nil, publicError(err)
	}
	return &TrackFilterConfig{handle: h}, nil
}

// Close releases the native tracking configuration; repeated calls are safe.
func (c *TrackFilterConfig) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return publicError(c.handle.Close())
}

// Dimension returns the configured state dimension of the tracking filter.
func (c *TrackFilterConfig) Dimension() (int, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	v, e := c.handle.Dimension()
	return v, publicError(e)
}

// Frame returns the coordinate frame configured for the tracking state.
func (c *TrackFilterConfig) Frame() (TrackCoordinateFrame, error) {
	if c == nil || c.handle == nil {
		return 0, ErrClosed
	}
	v, e := c.handle.Frame()
	return TrackCoordinateFrame(v), publicError(e)
}

// NewTrackFilter creates a native tracking filter from its configuration.
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

// NewTrackFilterFromPosition creates a native position-only tracking filter.
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
// NewTrackFilterFromPositionVelocity creates a native position-and-velocity tracking filter.
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

// Close releases the native tracking filter; repeated calls are safe.
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

// State returns the detached current tracking state and covariance metadata.
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

// Predict propagates the filter state by the supplied interval in seconds.
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

// Position returns detached position components from the selected state.
func (f *TrackFilter) Position() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Position()
	return append([]float64(nil), v...), publicError(e)
}

// Velocity returns detached velocity components from the current tracking state.
func (f *TrackFilter) Velocity() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Velocity()
	return append([]float64(nil), v...), publicError(e)
}

// StateVector returns the detached current state vector in filter order.
func (f *TrackFilter) StateVector() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.StateVector()
	return append([]float64(nil), v...), publicError(e)
}

// Covariance returns detached state covariance in row-major order.
func (f *TrackFilter) Covariance() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Covariance()
	return append([]float64(nil), v...), publicError(e)
}

// UpdatePosition applies a position measurement and returns the updated state.
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

// UpdateState applies a full state measurement and returns the updated state.
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

// UpdatePositionGated applies a position measurement and returns the updated state.
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

// NewTrackRTSHistoryBuilder creates a native RTS history recorder.
func NewTrackRTSHistoryBuilder() (*TrackRTSHistoryBuilder, error) {
	h, e := native.NewTrackRTSHistoryBuilder()
	if e != nil {
		return nil, publicError(e)
	}
	return &TrackRTSHistoryBuilder{handle: h}, nil
}

// NewTrackRTSHistoryBuilderFromFilter creates an RTS history recorder associated with the filter.
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

// Close releases the native RTS history builder; repeated calls are safe.
func (b *TrackRTSHistoryBuilder) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}

// Finish finalizes recorded RTS history and returns its native history handle.
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

// PredictRecorded propagates the filter by dt seconds and records the prediction.
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

// RecordPredictionOnly records the filter prediction in the RTS builder.
func (f *TrackFilter) RecordPredictionOnly(b *TrackRTSHistoryBuilder) error {
	if f == nil || f.handle == nil || b == nil || b.handle == nil {
		return ErrClosed
	}
	return publicError(f.handle.RecordPredictionOnly(b.handle))
}

// UpdatePositionRecorded applies a position measurement and returns the updated state.
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

// UpdatePositionGatedRecorded applies a position measurement and returns the updated state.
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

// PositionInnovation computes position innovation, covariance, and diagnostics.
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

// StateInnovation computes state innovation, covariance, and diagnostics.
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

// Close releases the native RTS history; repeated calls are safe.
func (h *TrackRTSHistory) Close() error {
	if h == nil || h.handle == nil {
		return nil
	}
	return publicError(h.handle.Close())
}

// EpochCount returns the number of recorded epochs.
func (h *TrackRTSHistory) EpochCount() (int, error) {
	if h == nil || h.handle == nil {
		return 0, ErrClosed
	}
	v, e := h.handle.EpochCount()
	return v, publicError(e)
}

// Epoch returns detached state and transition data for the selected history epoch.
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

// PredictedPosition returns the detached predicted position for the selected epoch.
func (h *TrackRTSHistory) PredictedPosition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Position(index, false)
	return append([]float64(nil), v...), publicError(e)
}

// UpdatedPosition returns the detached updated position for the selected epoch.
func (h *TrackRTSHistory) UpdatedPosition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Position(index, true)
	return append([]float64(nil), v...), publicError(e)
}

// Transition returns the detached state-transition matrix for the selected epoch.
func (h *TrackRTSHistory) Transition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Transition(index)
	return append([]float64(nil), v...), publicError(e)
}

// SmoothTrackRTS computes a Rauch–Tung–Striebel smoothed track from recorded history.
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

// Close releases the native smoothed track; repeated calls are safe.
func (s *SmoothedTrack) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// EpochCount returns the number of recorded epochs.
func (s *SmoothedTrack) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	v, e := s.handle.EpochCount()
	return v, publicError(e)
}

// Epoch returns detached state and transition data for the selected history epoch.
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

// Position returns detached position components from the selected state.
func (s *SmoothedTrack) Position(index int) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(index, 0)
	return append([]float64(nil), v...), publicError(e)
}

// Covariance returns detached state covariance in row-major order.
func (s *SmoothedTrack) Covariance(index int) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(index, 1)
	return append([]float64(nil), v...), publicError(e)
}

// RTSGainToNext returns the smoother gain to the next recorded epoch.
func (s *SmoothedTrack) RTSGainToNext(index int) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(index, 2)
	return append([]float64(nil), v...), publicError(e)
}
