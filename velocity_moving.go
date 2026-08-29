package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// VelocityObservable selects the units of velocity observations.
type VelocityObservable uint32

const (
	// VelocityObservableRangeRate interprets observations as metres per second.
	VelocityObservableRangeRate VelocityObservable = VelocityObservable(native.VelocityObservableRangeRateValue)
	// VelocityObservableDoppler interprets observations as Doppler hertz.
	VelocityObservableDoppler VelocityObservable = VelocityObservable(native.VelocityObservableDopplerValue)
)

// VelocityObservation is one pseudorange-rate or Doppler observation.
type VelocityObservation struct {
	// SatelliteID identifies the GNSS satellite.
	SatelliteID         string
	Value               float64
	CarrierHz           float64
	SatelliteClockDrift float64
}

// VelocityOptions controls corrections and observation interpretation.
type VelocityOptions struct {
	// Observable selects range-rate or Doppler interpretation.
	Observable VelocityObservable
	LightTime  bool
	Sagnac     bool
}

// VelocitySolution owns a native receiver ECEF velocity solution.
type VelocitySolution struct {
	_      noCopy
	handle *native.VelocitySolution
}

// VelocityOptionsDefaults returns the native default velocity options.
func VelocityOptionsDefaults() (VelocityOptions, error) {
	value, err := native.VelocityOptionsDefaults()
	return VelocityOptions{Observable: VelocityObservable(value.Observable), LightTime: value.LightTime, Sagnac: value.Sagnac}, publicError(err)
}

func nativeVelocityOptions(value *VelocityOptions) *native.VelocityOptions {
	if value == nil {
		return nil
	}
	return &native.VelocityOptions{Observable: uint32(value.Observable), LightTime: value.LightTime, Sagnac: value.Sagnac}
}

func nativeVelocityObservations(values []VelocityObservation) []native.VelocityObservation {
	out := make([]native.VelocityObservation, len(values))
	for i, value := range values {
		out[i] = native.VelocityObservation{SatelliteID: value.SatelliteID, Value: value.Value, CarrierHz: value.CarrierHz, SatelliteClockDrift: value.SatelliteClockDrift}
	}
	return out
}

// SolveVelocity solves receiver velocity and clock drift against an SP3 source.
func SolveVelocity(sp3 *SP3, observations []VelocityObservation, receiverECEFM [3]float64, tRxJ2000S float64, options *VelocityOptions) (*VelocitySolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	value, err := native.SolveVelocity(sp3.handle, nativeVelocityObservations(observations), receiverECEFM, tRxJ2000S, nativeVelocityOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &VelocitySolution{handle: value}, nil
}

// SolveVelocityBroadcast solves receiver velocity against broadcast ephemeris.
func SolveVelocityBroadcast(broadcast *BroadcastEphemeris, observations []VelocityObservation, receiverECEFM [3]float64, tRxJ2000S float64, options *VelocityOptions) (*VelocitySolution, error) {
	if broadcast == nil || broadcast.handle == nil {
		return nil, ErrClosed
	}
	value, err := native.SolveVelocityBroadcast(broadcast.handle, nativeVelocityObservations(observations), receiverECEFM, tRxJ2000S, nativeVelocityOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &VelocitySolution{handle: value}, nil
}

// Close releases the velocity solution and is idempotent.
func (s *VelocitySolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// ClockDrift returns receiver clock drift in seconds per second.
func (s *VelocitySolution) ClockDrift() (float64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.ClockDrift()
	return value, publicError(err)
}

// Residuals returns detached post-fit range-rate residuals in meters per second.
func (s *VelocitySolution) Residuals() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.Residuals()
	return append([]float64(nil), value...), publicError(err)
}

// Speed returns receiver speed in meters per second.
func (s *VelocitySolution) Speed() (float64, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.Speed()
	return value, publicError(err)
}

// StateCovariance returns the detached row-major 4x4 velocity covariance.
func (s *VelocitySolution) StateCovariance() ([16]float64, error) {
	if s == nil || s.handle == nil {
		return [16]float64{}, ErrClosed
	}
	value, err := s.handle.StateCovariance()
	return value, publicError(err)
}

// UsedSatelliteCount returns the number of satellites used by the solve.
func (s *VelocitySolution) UsedSatelliteCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.UsedSatelliteCount()
	return value, publicError(err)
}

// UsedSatelliteIDs returns detached used-satellite identifiers in solve order.
func (s *VelocitySolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.UsedSatelliteIDs()
	return append([]string(nil), value...), publicError(err)
}

// Velocity returns receiver ECEF velocity in meters per second.
func (s *VelocitySolution) Velocity() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.Velocity()
	return value, publicError(err)
}

// MovingBaselineStatus reports the integer-fix state of one epoch.
type MovingBaselineStatus uint32

const (
	// MovingBaselineStatusFixed indicates an integer-fixed baseline.
	MovingBaselineStatusFixed MovingBaselineStatus = MovingBaselineStatus(native.MovingBaselineStatusFixedValue)
	// MovingBaselineStatusFloat indicates a float baseline.
	MovingBaselineStatusFloat MovingBaselineStatus = MovingBaselineStatus(native.MovingBaselineStatusFloatValue)
)

// MovingBaselineEpochSummary is a detached moving-baseline epoch result.
type MovingBaselineEpochSummary struct {
	// BasePositionM, BaselineM, and BaselineLengthM are ECEF/local metre values.
	BasePositionM   [3]float64
	BaselineM       [3]float64
	BaselineLengthM float64
	Status          MovingBaselineStatus
	Float           RTKFloatMetadata
	Fixed           RTKFixedMetadata
}

// MovingBaselineEpoch is one moving-baseline input epoch.
type MovingBaselineEpoch struct {
	// BasePositionM is the ECEF base position in metres.
	BasePositionM       [3]float64
	Epoch               RTKEpoch
	AmbiguityIDs        []string
	AmbiguitySatellites []RTKAmbiguitySatellite
	WavelengthsM        []RTKFloatMapEntry
	OffsetsM            []RTKFloatMapEntry
	FloatOnlySystems    []GNSSSystem
}

// MovingBaselineConfig describes a complete native moving-baseline solve.
type MovingBaselineConfig struct {
	// Epochs is the copied observation sequence.
	Epochs           []MovingBaselineEpoch
	Model            RTKMeasurementModel
	FloatOptions     RTKFloatOptions
	FixedOptions     RTKFixedOptions
	InitialBaselineM [3]float64
	WarmStart        bool
	ReceiverAntenna  *RTKReceiverAntennaCorrections
}

func publicMovingBaselineSummary(value native.MovingBaselineEpochSummary) MovingBaselineEpochSummary {
	return MovingBaselineEpochSummary{BasePositionM: value.BasePositionM, BaselineM: value.BaselineM, BaselineLengthM: value.BaselineLengthM, Status: MovingBaselineStatus(value.Status), Float: publicRTKFloatMetadata(value.Float), Fixed: publicRTKFixedMetadata(value.Fixed)}
}

func nativeMovingBaselineConfig(value MovingBaselineConfig) native.MovingBaselineConfig {
	out := native.MovingBaselineConfig{Model: native.RtkMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting}, FloatOptions: native.RtkFloatOptions{PositionTolM: value.FloatOptions.PositionTolM, AmbiguityTolM: value.FloatOptions.AmbiguityTolM, MaxIterations: value.FloatOptions.MaxIterations}, FixedOptions: native.RtkFixedOptions{PositionTolM: value.FixedOptions.PositionTolM, AmbiguityTolM: value.FixedOptions.AmbiguityTolM, MaxIterations: value.FixedOptions.MaxIterations, RatioThreshold: value.FixedOptions.RatioThreshold, PartialAmbiguityResolution: value.FixedOptions.PartialAmbiguityResolution, PartialMinAmbiguities: value.FixedOptions.PartialMinAmbiguities}, InitialBaselineM: value.InitialBaselineM, WarmStart: value.WarmStart}
	out.Epochs = make([]native.MovingBaselineEpoch, len(value.Epochs))
	for i, epoch := range value.Epochs {
		floatOnlySystems := make([]uint32, len(epoch.FloatOnlySystems))
		for j, system := range epoch.FloatOnlySystems {
			floatOnlySystems[j] = uint32(system)
		}
		out.Epochs[i] = native.MovingBaselineEpoch{BasePositionM: epoch.BasePositionM, AmbiguityIDs: append([]string(nil), epoch.AmbiguityIDs...), FloatOnlySystems: floatOnlySystems}
		out.Epochs[i].Epoch = native.RtkEpoch{HasVelocityMPS: epoch.Epoch.HasVelocityMPS, VelocityMPS: epoch.Epoch.VelocityMPS, DTS: epoch.Epoch.DTS}
		out.Epochs[i].Epoch.References = nativeRTKMeasurements(epoch.Epoch.References)
		out.Epochs[i].Epoch.NonReference = nativeRTKMeasurements(epoch.Epoch.NonReference)
		out.Epochs[i].AmbiguitySatellites = make([]native.RtkAmbiguitySatellite, len(epoch.AmbiguitySatellites))
		for j, entry := range epoch.AmbiguitySatellites {
			out.Epochs[i].AmbiguitySatellites[j] = native.RtkAmbiguitySatellite{ID: entry.ID, SatelliteID: entry.SatelliteID}
		}
		out.Epochs[i].WavelengthsM = make([]native.RtkFloatMapEntry, len(epoch.WavelengthsM))
		for j, entry := range epoch.WavelengthsM {
			out.Epochs[i].WavelengthsM[j] = native.RtkFloatMapEntry{ID: entry.ID, Value: entry.Value}
		}
		out.Epochs[i].OffsetsM = make([]native.RtkFloatMapEntry, len(epoch.OffsetsM))
		for j, entry := range epoch.OffsetsM {
			out.Epochs[i].OffsetsM[j] = native.RtkFloatMapEntry{ID: entry.ID, Value: entry.Value}
		}
	}
	if value.ReceiverAntenna != nil {
		converted := &native.RtkReceiverAntennaCorrections{}
		converted.Base = nativeRTKReceiverAntennaCalibration(value.ReceiverAntenna.Base)
		converted.Rover = nativeRTKReceiverAntennaCalibration(value.ReceiverAntenna.Rover)
		out.ReceiverAntenna = converted
	}
	return out
}

func nativeRTKMeasurements(values []RTKSatMeasurement) []native.RtkSatMeasurement {
	out := make([]native.RtkSatMeasurement, len(values))
	for i, value := range values {
		out[i] = native.RtkSatMeasurement{SatelliteID: value.SatelliteID, SDAmbiguityID: value.SDAmbiguityID, BaseCodeM: value.BaseCodeM, BasePhaseM: value.BasePhaseM, RoverCodeM: value.RoverCodeM, RoverPhaseM: value.RoverPhaseM, BaseTXPos: value.BaseTXPos, RoverTXPos: value.RoverTXPos, Pos: value.Pos}
	}
	return out
}

// SolveMovingBaseline solves a moving-base RTK arc through the C ABI.
func SolveMovingBaseline(config MovingBaselineConfig) (*MovingBaselineSolution, error) {
	value, err := native.SolveMovingBaseline(nativeMovingBaselineConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &MovingBaselineSolution{handle: value}, nil
}

// MovingBaselineSolution owns a native moving-baseline result.
type MovingBaselineSolution struct {
	_      noCopy
	handle *native.MovingBaselineSolution
}

// Close releases the moving-baseline result and is idempotent.
func (s *MovingBaselineSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// EpochCount returns the number of solved epochs.
func (s *MovingBaselineSolution) EpochCount() (int, error) {
	if s == nil || s.handle == nil {
		return 0, ErrClosed
	}
	value, err := s.handle.EpochCount()
	return value, publicError(err)
}

// Epoch returns one detached moving-baseline epoch summary.
func (s *MovingBaselineSolution) Epoch(index int) (MovingBaselineEpochSummary, error) {
	if s == nil || s.handle == nil {
		return MovingBaselineEpochSummary{}, ErrClosed
	}
	value, err := s.handle.Epoch(index)
	if err != nil {
		return MovingBaselineEpochSummary{}, publicError(err)
	}
	return publicMovingBaselineSummary(value), nil
}

// Value is range rate in m/s or Doppler in Hz according to Observable.
// CarrierHz is the carrier frequency; SatelliteClockDrift is seconds per second.
// LightTime and Sagnac control optional native corrections.
// Status reports fixed versus float; Float and Fixed contain native metadata.
// Epoch contains detached RTK observations and timing.
// AmbiguityIDs and AmbiguitySatellites define the copied ambiguity graph.
// WavelengthsM and OffsetsM are copied ambiguity-keyed metre values.
// FloatOnlySystems selects constellations excluded from fixing.
// Model, FloatOptions, and FixedOptions select native estimation behavior.
// InitialBaselineM is the initial baseline in metres; WarmStart controls reuse.
// ReceiverAntenna optionally supplies copied antenna corrections.
