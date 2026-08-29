package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

type FusionFilterKind uint32

const (
	FusionEKF FusionFilterKind = FusionFilterKind(native.FusionFilterEKFValue)
	FusionUKF FusionFilterKind = FusionFilterKind(native.FusionFilterUKFValue)
)

type FusionErrorStateLayout uint32

const (
	FusionFifteenState   FusionErrorStateLayout = FusionErrorStateLayout(native.FusionLayoutFifteenValue)
	FusionTwentyOneState FusionErrorStateLayout = FusionErrorStateLayout(native.FusionLayoutTwentyOneValue)
)

type FusionIMUGrade uint32

const (
	FusionIMUMEMS       FusionIMUGrade = FusionIMUGrade(native.FusionImuGradeMEMSValue)
	FusionIMUTactical   FusionIMUGrade = FusionIMUGrade(native.FusionImuGradeTacticalValue)
	FusionIMUNavigation FusionIMUGrade = FusionIMUGrade(native.FusionImuGradeNavigationValue)
)

type FusionIMUSampleKind uint32

const (
	FusionIMURate      FusionIMUSampleKind = FusionIMUSampleKind(native.FusionSampleRateValue)
	FusionIMUIncrement FusionIMUSampleKind = FusionIMUSampleKind(native.FusionSampleIncrementValue)
)

type FusionGNSSFixStatus uint32

const (
	FusionFixSingle FusionGNSSFixStatus = FusionGNSSFixStatus(native.FusionFixSingleValue)
	FusionFixFloat  FusionGNSSFixStatus = FusionGNSSFixStatus(native.FusionFixFloatValue)
	FusionFixFixed  FusionGNSSFixStatus = FusionGNSSFixStatus(native.FusionFixFixedValue)
)

type FusionIMUSpec struct {
	AccelVRWMPerSqrtS, GyroARRadPerSqrtS, AccelBiasInstabilityMPerS2, GyroBiasInstabilityRadPerS, AccelBiasTauS, GyroBiasTauS float64
	HasAccelScaleInstability, HasGyroScaleInstability                                                                         bool
	AccelScaleInstabilityPPM, GyroScaleInstabilityPPM                                                                         float64
}
type FusionFilterConfig struct {
	FilterKind                                      FusionFilterKind
	Layout                                          FusionErrorStateLayout
	IMUSpec                                         FusionIMUSpec
	TimeSyncIMUCapacity, TimeSyncCheckpointCapacity int
}
type FusionNavState struct {
	EpochJ2000S                                                         float64
	PositionECEFM, VelocityECEFMPerS                                    [3]float64
	AttitudeBodyToECEF                                                  [9]float64
	AccelBiasMPerS2, GyroBiasRadPerS, AccelScaleFactor, GyroScaleFactor [3]float64
}
type FusionIMUSample struct {
	EpochJ2000S                                                                float64
	Kind                                                                       FusionIMUSampleKind
	SpecificForceMPerS2, AngularRateRadPerS, DeltaVelocityMPerS, DeltaThetaRad [3]float64
	DTS                                                                        float64
}
type FusionState struct {
	EpochJ2000S                                                         float64
	PositionECEFM, VelocityECEFMPerS                                    [3]float64
	AttitudeBodyToECEF                                                  [9]float64
	AccelBiasMPerS2, GyroBiasRadPerS, AccelScaleFactor, GyroScaleFactor [3]float64
	CovarianceDimension                                                 int
	LastBodyRateRadPerS                                                 [3]float64
	TightClockBiasM, TightClockDriftMPerS                               float64
	TightClockCovariance                                                [4]float64
}
type FusionTimeSyncStatus struct {
	IMUCapacity, IMULength, CheckpointCapacity, CheckpointLength         int
	HasOldestIMU, HasNewestIMU, HasOldestCheckpoint, HasNewestCheckpoint bool
	OldestIMU, NewestIMU, OldestCheckpoint, NewestCheckpoint             float64
}
type FusionLooseMeasurement struct {
	EpochJ2000S                      float64
	PositionECEFM, VelocityECEFMPerS [3]float64
	HasVelocity                      bool
	Covariance                       []float64
	SatellitesUsed                   int
	SolutionValid                    bool
	FixStatus                        FusionGNSSFixStatus
}
type FusionUpdate struct {
	Applied                          bool
	NIS                              float64
	Rows, AcceptedRows, RejectedRows int
}
type FusionTimeSyncUpdate struct {
	Update                                            FusionUpdate
	LateMeasurement                                   bool
	ReplayedIMUSegments                               int
	RestoredCheckpointEpochJ2000S, CurrentEpochJ2000S float64
}

// FusionTightUpdateMode selects the native direct, recorded-history, or
// time-synchronizing tight update route.
type FusionTightUpdateMode uint32

const (
	FusionTightDirect   FusionTightUpdateMode = 0
	FusionTightRecorded FusionTightUpdateMode = 1
	FusionTightTimeSync FusionTightUpdateMode = 2
)

type FusionTightRangeRate struct {
	MeasuredMPerS, SigmaMPerS, SatelliteClockDriftMPerS float64
}
type FusionTightCarrierPhase struct {
	PhaseRangeM, SigmaM, FloatAmbiguityM float64
}
type FusionTightObservation struct {
	SatelliteID          string
	PseudorangeM, SigmaM float64
	HasRangeRate         bool
	RangeRate            FusionTightRangeRate
	HasCarrierPhase      bool
	CarrierPhase         FusionTightCarrierPhase
	IonosphereDelayM     float64
	TroposphereDelayM    float64
}
type FusionTightEpoch struct {
	EpochJ2000S  float64
	Observations []FusionTightObservation
}
type FusionRTSEpoch struct {
	EpochJ2000S                             float64
	CovarianceDimension, AugmentedDimension int
	HasTransitionFromPrevious               bool
}
type FusionFilter struct {
	_      noCopy
	handle *native.FusionFilter
}
type FusionRTSHistoryBuilder struct {
	_      noCopy
	handle *native.FusionRTSHistoryBuilder
}
type FusionRTSHistory struct {
	_      noCopy
	handle *native.FusionRTSHistory
}

func DefaultFusionFilterConfig() (FusionFilterConfig, error) {
	v, e := native.FusionConfigDefault()
	imuCapacity, conversionErr := nativeCountToInt(v.TimeSyncIMUCapacity, "fusion IMU capacity")
	if e == nil && conversionErr != nil {
		return FusionFilterConfig{}, conversionErr
	}
	checkpointCapacity, conversionErr := nativeCountToInt(v.TimeSyncCheckpointCapacity, "fusion checkpoint capacity")
	if e == nil && conversionErr != nil {
		return FusionFilterConfig{}, conversionErr
	}
	return FusionFilterConfig{FusionFilterKind(v.FilterKind), FusionErrorStateLayout(v.Layout), FusionIMUSpec{v.IMUSpec.AccelVRW, v.IMUSpec.GyroARW, v.IMUSpec.AccelBiasInstability, v.IMUSpec.GyroBiasInstability, v.IMUSpec.AccelBiasTau, v.IMUSpec.GyroBiasTau, v.IMUSpec.HasAccelScale, v.IMUSpec.HasGyroScale, v.IMUSpec.AccelScalePPM, v.IMUSpec.GyroScalePPM}, imuCapacity, checkpointCapacity}, publicError(e)
}
func nativeFusionConfig(v FusionFilterConfig) native.NativeFusionConfig {
	return native.NativeFusionConfig{FilterKind: uint32(v.FilterKind), Layout: uint32(v.Layout), IMUSpec: native.NativeFusionIMUSpec{AccelVRW: v.IMUSpec.AccelVRWMPerSqrtS, GyroARW: v.IMUSpec.GyroARRadPerSqrtS, AccelBiasInstability: v.IMUSpec.AccelBiasInstabilityMPerS2, GyroBiasInstability: v.IMUSpec.GyroBiasInstabilityRadPerS, AccelBiasTau: v.IMUSpec.AccelBiasTauS, GyroBiasTau: v.IMUSpec.GyroBiasTauS, HasAccelScale: v.IMUSpec.HasAccelScaleInstability, HasGyroScale: v.IMUSpec.HasGyroScaleInstability, AccelScalePPM: v.IMUSpec.AccelScaleInstabilityPPM, GyroScalePPM: v.IMUSpec.GyroScaleInstabilityPPM}, TimeSyncIMUCapacity: uint64(v.TimeSyncIMUCapacity), TimeSyncCheckpointCapacity: uint64(v.TimeSyncCheckpointCapacity)}
}
func nativeFusionNav(v FusionNavState) native.NativeFusionNavState {
	return native.NativeFusionNavState{Epoch: v.EpochJ2000S, Position: v.PositionECEFM, Velocity: v.VelocityECEFMPerS, Attitude: v.AttitudeBodyToECEF, AccelBias: v.AccelBiasMPerS2, GyroBias: v.GyroBiasRadPerS, AccelScale: v.AccelScaleFactor, GyroScale: v.GyroScaleFactor}
}
func NewFusionFilter(initial FusionNavState, covarianceDiagonal []float64, config FusionFilterConfig) (*FusionFilter, error) {
	if config.TimeSyncIMUCapacity < 0 || config.TimeSyncCheckpointCapacity < 0 {
		return nil, errors.New("sidereon: fusion capacities must not be negative")
	}
	if expected, ok := fusionCovarianceDimension(config.Layout); ok && len(covarianceDiagonal) != expected {
		return nil, errors.New("sidereon: fusion covariance diagonal shape does not match error-state layout")
	}
	h, e := native.NewFusionFilter(nativeFusionNav(initial), append([]float64(nil), covarianceDiagonal...), nativeFusionConfig(config))
	if e != nil {
		return nil, publicError(e)
	}
	return &FusionFilter{handle: h}, nil
}

func fusionCovarianceDimension(layout FusionErrorStateLayout) (int, bool) {
	switch layout {
	case FusionFifteenState:
		return 15, true
	case FusionTwentyOneState:
		return 21, true
	default:
		return 0, false
	}
}

func FusionLayoutDimension(layout FusionErrorStateLayout) (int, error) {
	v, e := native.FusionLayoutDimension(uint32(layout))
	return v, publicError(e)
}
func FusionIMUPreset(grade FusionIMUGrade) (FusionIMUSpec, error) {
	v, e := native.FusionIMUPreset(uint32(grade))
	return FusionIMUSpec{v.AccelVRW, v.GyroARW, v.AccelBiasInstability, v.GyroBiasInstability, v.AccelBiasTau, v.GyroBiasTau, v.HasAccelScale, v.HasGyroScale, v.AccelScalePPM, v.GyroScalePPM}, publicError(e)
}

// FusionLabels returns the C labels for a filter kind, error-state layout, and
// GNSS fix status. The returned strings are copied from the native boundary.
func FusionLabels(kind FusionFilterKind, layout FusionErrorStateLayout, fix FusionGNSSFixStatus) (filterLabel, layoutLabel, fixLabel string, err error) {
	filterLabel, layoutLabel, fixLabel, err = native.FusionLabels(uint32(kind), uint32(layout), uint32(fix))
	return filterLabel, layoutLabel, fixLabel, publicError(err)
}
func (f *FusionFilter) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return publicError(f.handle.Close())
}
func checkedFusionState(v native.NativeFusionState) (FusionState, error) {
	dimension, err := nativeCountToInt(v.CovarianceDimension, "fusion covariance dimension")
	if err != nil {
		return FusionState{}, err
	}
	return FusionState{v.Epoch, v.Position, v.Velocity, v.Attitude, v.AccelBias, v.GyroBias, v.AccelScale, v.GyroScale, dimension, v.LastBodyRate, v.ClockBias, v.ClockDrift, v.ClockCovariance}, nil
}
func checkedFusionUpdate(v native.NativeFusionUpdate) (FusionUpdate, error) {
	rows, err := nativeCountToInt(v.Rows, "fusion update rows")
	if err != nil {
		return FusionUpdate{}, err
	}
	accepted, err := nativeCountToInt(v.AcceptedRows, "fusion accepted rows")
	if err != nil {
		return FusionUpdate{}, err
	}
	rejected, err := nativeCountToInt(v.RejectedRows, "fusion rejected rows")
	if err != nil {
		return FusionUpdate{}, err
	}
	return FusionUpdate{v.Applied, v.NIS, rows, accepted, rejected}, nil
}
func checkedFusionTimeSyncUpdate(v native.NativeFusionTimeSyncUpdate) (FusionTimeSyncUpdate, error) {
	update, err := checkedFusionUpdate(v.Update)
	if err != nil {
		return FusionTimeSyncUpdate{}, err
	}
	replayed, err := nativeCountToInt(v.Replayed, "fusion replayed IMU segments")
	if err != nil {
		return FusionTimeSyncUpdate{}, err
	}
	return FusionTimeSyncUpdate{update, v.Late, replayed, v.RestoredEpoch, v.CurrentEpoch}, nil
}
func (f *FusionFilter) State() (FusionState, error) {
	if f == nil || f.handle == nil {
		return FusionState{}, ErrClosed
	}
	v, e := f.handle.State()
	state, conversionErr := checkedFusionState(v)
	if e != nil {
		return FusionState{}, publicError(e)
	}
	return state, conversionErr
}
func (f *FusionFilter) TimeSyncStatus() (FusionTimeSyncStatus, error) {
	if f == nil || f.handle == nil {
		return FusionTimeSyncStatus{}, ErrClosed
	}
	v, e := f.handle.TimeSync()
	imuCapacity, conversionErr := nativeCountToInt(v.IMUCapacity, "fusion IMU capacity")
	if e == nil && conversionErr != nil {
		return FusionTimeSyncStatus{}, conversionErr
	}
	imuLength, conversionErr := nativeCountToInt(v.IMULength, "fusion IMU length")
	if e == nil && conversionErr != nil {
		return FusionTimeSyncStatus{}, conversionErr
	}
	checkpointCapacity, conversionErr := nativeCountToInt(v.CheckpointCapacity, "fusion checkpoint capacity")
	if e == nil && conversionErr != nil {
		return FusionTimeSyncStatus{}, conversionErr
	}
	checkpointLength, conversionErr := nativeCountToInt(v.CheckpointLength, "fusion checkpoint length")
	if e == nil && conversionErr != nil {
		return FusionTimeSyncStatus{}, conversionErr
	}
	return FusionTimeSyncStatus{imuCapacity, imuLength, checkpointCapacity, checkpointLength, v.HasOldestIMU, v.HasNewestIMU, v.HasOldestCheckpoint, v.HasNewestCheckpoint, v.OldestIMU, v.NewestIMU, v.OldestCheckpoint, v.NewestCheckpoint}, publicError(e)
}
func (f *FusionFilter) Covariance() ([]float64, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Covariance()
	return append([]float64(nil), v...), publicError(e)
}
func (f *FusionFilter) Propagate(s FusionIMUSample) error {
	if f == nil || f.handle == nil {
		return ErrClosed
	}
	v := native.NativeFusionIMUSample{Epoch: s.EpochJ2000S, Kind: uint32(s.Kind), SpecificForce: s.SpecificForceMPerS2, AngularRate: s.AngularRateRadPerS, DeltaVelocity: s.DeltaVelocityMPerS, DeltaTheta: s.DeltaThetaRad, DTS: s.DTS}
	return publicError(f.handle.Propagate(v))
}
func (f *FusionFilter) Encode() ([]byte, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	v, e := f.handle.Encode()
	return append([]byte(nil), v...), publicError(e)
}
func (f *FusionFilter) Restore(data []byte) error {
	if f == nil || f.handle == nil {
		return ErrClosed
	}
	return publicError(f.handle.Restore(append([]byte(nil), data...)))
}
func nativeFusionLoose(v FusionLooseMeasurement) native.NativeFusionLooseMeasurement {
	return native.NativeFusionLooseMeasurement{Epoch: v.EpochJ2000S, Position: v.PositionECEFM, Velocity: v.VelocityECEFMPerS, HasVelocity: v.HasVelocity, Covariance: append([]float64(nil), v.Covariance...), SatellitesUsed: uint64(v.SatellitesUsed), SolutionValid: v.SolutionValid, FixStatus: uint32(v.FixStatus)}
}

func validateFusionLoose(v FusionLooseMeasurement) error {
	if v.SatellitesUsed < 0 {
		return errors.New("sidereon: fusion satellite count must not be negative")
	}
	return nil
}
func (f *FusionFilter) UpdateLoose(m FusionLooseMeasurement) (FusionUpdate, error) {
	if f == nil || f.handle == nil {
		return FusionUpdate{}, ErrClosed
	}
	if err := validateFusionLoose(m); err != nil {
		return FusionUpdate{}, err
	}
	v, e := f.handle.UpdateLoose(nativeFusionLoose(m))
	update, conversionErr := checkedFusionUpdate(v)
	if e != nil {
		return FusionUpdate{}, publicError(e)
	}
	return update, conversionErr
}
func (f *FusionFilter) UpdateStationary() (FusionUpdate, bool, error) {
	if f == nil || f.handle == nil {
		return FusionUpdate{}, false, ErrClosed
	}
	v, p, e := f.handle.UpdateStationary()
	update, conversionErr := checkedFusionUpdate(v)
	if e != nil {
		return FusionUpdate{}, p, publicError(e)
	}
	return update, p, conversionErr
}
func (f *FusionFilter) UpdateNonHolonomic() (FusionUpdate, bool, error) {
	if f == nil || f.handle == nil {
		return FusionUpdate{}, false, ErrClosed
	}
	v, p, e := f.handle.UpdateNonHolonomic()
	update, conversionErr := checkedFusionUpdate(v)
	if e != nil {
		return FusionUpdate{}, p, publicError(e)
	}
	return update, p, conversionErr
}

func (f *FusionFilter) ConfigureTimeSync(imuCapacity, checkpointCapacity int) error {
	if f == nil || f.handle == nil {
		return ErrClosed
	}
	if imuCapacity < 0 || checkpointCapacity < 0 {
		return errors.New("sidereon: fusion capacities must not be negative")
	}
	return publicError(f.handle.ConfigureTimeSync(uint64(imuCapacity), uint64(checkpointCapacity)))
}

func nativeFusionTight(v FusionTightEpoch) native.NativeFusionTightEpoch {
	out := native.NativeFusionTightEpoch{Epoch: v.EpochJ2000S, Observations: make([]native.NativeFusionTightObservation, len(v.Observations))}
	for i, observation := range v.Observations {
		out.Observations[i] = native.NativeFusionTightObservation{
			SatelliteID: observation.SatelliteID, Pseudorange: observation.PseudorangeM, Sigma: observation.SigmaM,
			HasRangeRate:    observation.HasRangeRate,
			RangeRate:       native.NativeFusionTightRangeRate{Measured: observation.RangeRate.MeasuredMPerS, Sigma: observation.RangeRate.SigmaMPerS, ClockDrift: observation.RangeRate.SatelliteClockDriftMPerS},
			HasCarrierPhase: observation.HasCarrierPhase,
			CarrierPhase:    native.NativeFusionTightCarrierPhase{PhaseRange: observation.CarrierPhase.PhaseRangeM, Sigma: observation.CarrierPhase.SigmaM, FloatAmbiguity: observation.CarrierPhase.FloatAmbiguityM},
			Ionosphere:      observation.IonosphereDelayM, Troposphere: observation.TroposphereDelayM,
		}
	}
	return out
}

func (f *FusionFilter) PropagateRecorded(sample FusionIMUSample, history *FusionRTSHistoryBuilder) error {
	if f == nil || f.handle == nil || history == nil || history.handle == nil {
		return ErrClosed
	}
	return publicError(f.handle.PropagateRecorded(native.NativeFusionIMUSample{Epoch: sample.EpochJ2000S, Kind: uint32(sample.Kind), SpecificForce: sample.SpecificForceMPerS2, AngularRate: sample.AngularRateRadPerS, DeltaVelocity: sample.DeltaVelocityMPerS, DeltaTheta: sample.DeltaThetaRad, DTS: sample.DTS}, history.handle))
}

func (f *FusionFilter) UpdateLooseRecorded(m FusionLooseMeasurement, history *FusionRTSHistoryBuilder) (FusionUpdate, error) {
	if f == nil || f.handle == nil || history == nil || history.handle == nil {
		return FusionUpdate{}, ErrClosed
	}
	if err := validateFusionLoose(m); err != nil {
		return FusionUpdate{}, err
	}
	v, err := f.handle.UpdateLooseRecorded(nativeFusionLoose(m), history.handle)
	update, conversionErr := checkedFusionUpdate(v)
	if err != nil {
		return FusionUpdate{}, publicError(err)
	}
	return update, conversionErr
}

func (f *FusionFilter) UpdateLooseTimeSync(m FusionLooseMeasurement) (FusionTimeSyncUpdate, error) {
	if f == nil || f.handle == nil {
		return FusionTimeSyncUpdate{}, ErrClosed
	}
	if err := validateFusionLoose(m); err != nil {
		return FusionTimeSyncUpdate{}, err
	}
	v, err := f.handle.UpdateLooseTimeSync(nativeFusionLoose(m))
	update, conversionErr := checkedFusionTimeSyncUpdate(v)
	if err != nil {
		return FusionTimeSyncUpdate{}, publicError(err)
	}
	return update, conversionErr
}

func (f *FusionFilter) UpdateStationaryRecorded(history *FusionRTSHistoryBuilder) (FusionUpdate, bool, error) {
	if f == nil || f.handle == nil || history == nil || history.handle == nil {
		return FusionUpdate{}, false, ErrClosed
	}
	v, propagated, err := f.handle.UpdateStationaryRecorded(history.handle)
	update, conversionErr := checkedFusionUpdate(v)
	if err != nil {
		return FusionUpdate{}, propagated, publicError(err)
	}
	return update, propagated, conversionErr
}

func (f *FusionFilter) UpdateNonHolonomicRecorded(history *FusionRTSHistoryBuilder) (FusionUpdate, bool, error) {
	if f == nil || f.handle == nil || history == nil || history.handle == nil {
		return FusionUpdate{}, false, ErrClosed
	}
	v, propagated, err := f.handle.UpdateNonHolonomicRecorded(history.handle)
	update, conversionErr := checkedFusionUpdate(v)
	if err != nil {
		return FusionUpdate{}, propagated, publicError(err)
	}
	return update, propagated, conversionErr
}

func (f *FusionFilter) UpdateTightSP3(epoch FusionTightEpoch, sp3 *SP3, history *FusionRTSHistoryBuilder, mode FusionTightUpdateMode) (FusionUpdate, FusionTimeSyncUpdate, error) {
	if f == nil || f.handle == nil || sp3 == nil || sp3.handle == nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, ErrClosed
	}
	if mode > FusionTightTimeSync {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, errors.New("sidereon: unknown tight-update mode")
	}
	if mode == FusionTightRecorded && (history == nil || history.handle == nil) {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, ErrClosed
	}
	var nativeHistory *native.FusionRTSHistoryBuilder
	if history != nil {
		nativeHistory = history.handle
	}
	update, syncUpdate, err := f.handle.UpdateTightSP3(nativeFusionTight(epoch), sp3.handle, nativeHistory, uint32(mode))
	convertedUpdate, conversionErr := checkedFusionUpdate(update)
	if err != nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, publicError(err)
	}
	if conversionErr != nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, conversionErr
	}
	convertedSync, conversionErr := checkedFusionTimeSyncUpdate(syncUpdate)
	if conversionErr != nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, conversionErr
	}
	return convertedUpdate, convertedSync, nil
}

func (f *FusionFilter) UpdateTightBroadcast(epoch FusionTightEpoch, broadcast *BroadcastEphemeris, history *FusionRTSHistoryBuilder, mode FusionTightUpdateMode) (FusionUpdate, FusionTimeSyncUpdate, error) {
	if f == nil || f.handle == nil || broadcast == nil || broadcast.handle == nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, ErrClosed
	}
	if mode > FusionTightTimeSync {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, errors.New("sidereon: unknown tight-update mode")
	}
	if mode == FusionTightRecorded && (history == nil || history.handle == nil) {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, ErrClosed
	}
	var nativeHistory *native.FusionRTSHistoryBuilder
	if history != nil {
		nativeHistory = history.handle
	}
	update, syncUpdate, err := f.handle.UpdateTightBroadcast(nativeFusionTight(epoch), broadcast.handle, nativeHistory, uint32(mode))
	convertedUpdate, conversionErr := checkedFusionUpdate(update)
	if err != nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, publicError(err)
	}
	if conversionErr != nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, conversionErr
	}
	convertedSync, conversionErr := checkedFusionTimeSyncUpdate(syncUpdate)
	if conversionErr != nil {
		return FusionUpdate{}, FusionTimeSyncUpdate{}, conversionErr
	}
	return convertedUpdate, convertedSync, nil
}
func NewFusionRTSHistoryBuilder() (*FusionRTSHistoryBuilder, error) {
	h, e := native.NewFusionRTSHistoryBuilder()
	if e != nil {
		return nil, publicError(e)
	}
	return &FusionRTSHistoryBuilder{handle: h}, nil
}
func NewFusionRTSHistoryBuilderFromFilter(f *FusionFilter) (*FusionRTSHistoryBuilder, error) {
	if f == nil || f.handle == nil {
		return nil, ErrClosed
	}
	h, e := native.NewFusionRTSHistoryBuilderFromFilter(f.handle)
	if e != nil {
		return nil, publicError(e)
	}
	return &FusionRTSHistoryBuilder{handle: h}, nil
}
func (b *FusionRTSHistoryBuilder) Close() error {
	if b == nil || b.handle == nil {
		return nil
	}
	return publicError(b.handle.Close())
}
func (b *FusionRTSHistoryBuilder) Finish() (*FusionRTSHistory, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	h, e := b.handle.Finish()
	if e != nil {
		return nil, publicError(e)
	}
	return &FusionRTSHistory{handle: h}, nil
}
func (h *FusionRTSHistory) Close() error {
	if h == nil || h.handle == nil {
		return nil
	}
	return publicError(h.handle.Close())
}
func (h *FusionRTSHistory) EpochCount() (int, error) {
	if h == nil || h.handle == nil {
		return 0, ErrClosed
	}
	v, e := h.handle.EpochCount()
	return v, publicError(e)
}
func (h *FusionRTSHistory) Epoch(index int) (FusionRTSEpoch, error) {
	if h == nil || h.handle == nil {
		return FusionRTSEpoch{}, ErrClosed
	}
	v, e := h.handle.Epoch(index)
	covarianceDimension, conversionErr := nativeCountToInt(v.CovarianceDimension, "fusion history covariance dimension")
	if e == nil && conversionErr != nil {
		return FusionRTSEpoch{}, conversionErr
	}
	augmentedDimension, conversionErr := nativeCountToInt(v.AugmentedDimension, "fusion history augmented dimension")
	if e == nil && conversionErr != nil {
		return FusionRTSEpoch{}, conversionErr
	}
	return FusionRTSEpoch{v.Epoch, covarianceDimension, augmentedDimension, v.HasTransition}, publicError(e)
}
func (h *FusionRTSHistory) PredictedPosition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Values(index, 0)
	return append([]float64(nil), v...), publicError(e)
}
func (h *FusionRTSHistory) UpdatedPosition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Values(index, 1)
	return append([]float64(nil), v...), publicError(e)
}
func (h *FusionRTSHistory) Transition(index int) ([]float64, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, e := h.handle.Values(index, 2)
	return append([]float64(nil), v...), publicError(e)
}

// FusionVelocityMatchState is an ECEF state in metres, metres per second, and
// seconds since J2000. It is used to repair a short GNSS outage.
type FusionVelocityMatchState struct {
	EpochJ2000S       float64
	PositionECEFM     [3]float64
	VelocityECEFMPerS [3]float64
}

// FusionVelocityMatchingConfig limits the outage duration eligible for
// matching, in seconds.
type FusionVelocityMatchingConfig struct{ MaxOutageDurationS float64 }

// FusionVelocityMatchedTrajectory contains the native matched-state count and
// endpoint corrections in ECEF metres and metres per second.
type FusionVelocityMatchedTrajectory struct {
	StateCount                          int
	EndpointPositionCorrectionECEFM     [3]float64
	EndpointVelocityCorrectionECEFMPerS [3]float64
}

// FusionVelocityMatchOutage delegates outage repair and all interpolation to
// the C engine. Inputs and returned state rows are copied at the boundary.
func FusionVelocityMatchOutage(states []FusionVelocityMatchState, firstGoodFix FusionLooseMeasurement, config FusionVelocityMatchingConfig) ([]FusionVelocityMatchState, FusionVelocityMatchedTrajectory, error) {
	if err := validateFusionLoose(firstGoodFix); err != nil {
		return nil, FusionVelocityMatchedTrajectory{}, err
	}
	nativeStates := make([]native.NativeFusionVelocityMatchState, len(states))
	for i, state := range states {
		nativeStates[i] = native.NativeFusionVelocityMatchState{Epoch: state.EpochJ2000S, Position: state.PositionECEFM, Velocity: state.VelocityECEFMPerS}
	}
	matched, trajectory, err := native.FusionVelocityMatchOutage(nativeStates, nativeFusionLoose(firstGoodFix), native.NativeFusionVelocityMatchingConfig{MaxOutageDuration: config.MaxOutageDurationS})
	if err != nil {
		return nil, FusionVelocityMatchedTrajectory{}, publicError(err)
	}
	out := make([]FusionVelocityMatchState, len(matched))
	for i, state := range matched {
		out[i] = FusionVelocityMatchState{state.Epoch, state.Position, state.Velocity}
	}
	stateCount, err := nativeCountToInt(trajectory.StateCount, "fusion velocity-match state count")
	if err != nil {
		return nil, FusionVelocityMatchedTrajectory{}, err
	}
	return out, FusionVelocityMatchedTrajectory{stateCount, trajectory.EndpointPositionCorrection, trajectory.EndpointVelocityCorrection}, nil
}

// SmoothedFusionEpoch describes one fixed-interval RTS-smoothed epoch. The
// covariance and correction vectors are copied by the corresponding methods;
// the RTS gain is absent on the final epoch.
type SmoothedFusionEpoch struct {
	EpochJ2000S                        float64
	CovarianceDimension, CorrectionLen int
	HasRTSGainToNext                   bool
}

// SmoothedFusionTrajectory owns a C-backed fixed-interval RTS result.
type SmoothedFusionTrajectory struct {
	_      noCopy
	handle *native.SmoothedFusionTrajectory
}

// Smooth applies the C-backed fixed-interval RTS smoother to a finished fusion
// history. The history remains independently owned.
func (h *FusionRTSHistory) Smooth() (*SmoothedFusionTrajectory, error) {
	if h == nil || h.handle == nil {
		return nil, ErrClosed
	}
	v, err := h.handle.Smooth()
	if err != nil {
		return nil, publicError(err)
	}
	return &SmoothedFusionTrajectory{handle: v}, nil
}

func (s *SmoothedFusionTrajectory) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

func (s *SmoothedFusionTrajectory) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	v, err := s.handle.EpochCount()
	return v, publicError(err)
}

func (s *SmoothedFusionTrajectory) Epoch(index int) (SmoothedFusionEpoch, error) {
	if s == nil || s.handle == nil {
		return SmoothedFusionEpoch{}, ErrClosed
	}
	v, err := s.handle.Epoch(index)
	covarianceDimension, conversionErr := nativeCountToInt(v.CovarianceDimension, "smoothed fusion covariance dimension")
	if err == nil && conversionErr != nil {
		return SmoothedFusionEpoch{}, conversionErr
	}
	correctionDimension, conversionErr := nativeCountToInt(v.AugmentedDimension, "smoothed fusion correction dimension")
	if err == nil && conversionErr != nil {
		return SmoothedFusionEpoch{}, conversionErr
	}
	return SmoothedFusionEpoch{v.Epoch, covarianceDimension, correctionDimension, v.HasTransition}, publicError(err)
}

// Covariance returns the row-major smoothed covariance for an epoch.
func (s *SmoothedFusionTrajectory) Covariance(index int) ([]float64, error) {
	return s.values(index, 0)
}

// ErrorStateCorrection returns the smoothed error-state correction vector.
func (s *SmoothedFusionTrajectory) ErrorStateCorrection(index int) ([]float64, error) {
	return s.values(index, 1)
}

// PositionECEF returns the smoothed ECEF position in metres.
func (s *SmoothedFusionTrajectory) PositionECEF(index int) ([]float64, error) {
	return s.values(index, 2)
}

// RTSGainToNext returns the row-major gain to the next epoch, or an empty
// slice when the native epoch reports that no gain is present.
func (s *SmoothedFusionTrajectory) RTSGainToNext(index int) ([]float64, error) {
	return s.values(index, 3)
}

func (s *SmoothedFusionTrajectory) values(index, kind int) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, err := s.handle.Values(index, uint32(kind))
	return append([]float64(nil), v...), publicError(err)
}
