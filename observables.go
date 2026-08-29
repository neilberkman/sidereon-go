package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// ObservablesOptions controls satellite-state to observable prediction.
// CarrierHz is in hertz; LightTime and Sagnac enable the corresponding
// Earth-fixed propagation corrections.
type ObservablesOptions struct {
	CarrierHz         float64
	LightTime, Sagnac bool
}

// NewObservablesOptions returns the native default prediction options.
func NewObservablesOptions() (ObservablesOptions, error) {
	v, e := native.ObservablesOptionsInit()
	return ObservablesOptions{CarrierHz: v.CarrierHz, LightTime: v.LightTime, Sagnac: v.Sagnac}, publicError(e)
}

// EmissionMediaOptions controls emission-time media calculations. Elevation
// is in radians, pressure in hPa, temperature in kelvin, humidity as a ratio,
// and carrier frequency in hertz.
type EmissionMediaOptions struct {
	CarrierHz                                   float64
	MinElevationEnabled                         bool
	MinElevationRad                             float64
	TroposphereEnabled                          bool
	PressureHPA, TemperatureK, RelativeHumidity float64
}

// NewEmissionMediaOptions returns the native default media options.
func NewEmissionMediaOptions() (EmissionMediaOptions, error) {
	v, e := native.EmissionMediaOptionsInit()
	return EmissionMediaOptions{CarrierHz: v.CarrierHz, MinElevationEnabled: v.MinElevationEnabled, MinElevationRad: v.MinElevationRad, TroposphereEnabled: v.TroposphereEnabled, PressureHPA: v.PressureHPA, TemperatureK: v.TemperatureK, RelativeHumidity: v.RelativeHumidity}, publicError(e)
}

// EmissionMediaStatus identifies the native emission-media result state.
type EmissionMediaStatus uint32

const (
	EmissionMediaValid                EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaValidValue)
	EmissionMediaGap                  EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaGapValue)
	EmissionMediaBelowElevationCutoff EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaBelowElevationValue)
	EmissionMediaError                EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaErrorValue)
)

// EmissionMediaRow contains one satellite/epoch result. Positions are ECEF
// metres and times are J2000 seconds; Has fields distinguish absent values
// from present zero values.
type EmissionMediaRow struct {
	PositionECEFM           [3]float64
	HasPosition             bool
	ClockS                  float64
	HasClock                bool
	IonosphereSlantDelayM   float64
	HasIonosphereSlantDelay bool
	TroposphereDelayM       float64
	HasTroposphereDelay     bool
	Status                  EmissionMediaStatus
	ResultStatus            StatusCode
}

// MissingObservablePositionECEF returns the native missing-position sentinel.
func MissingObservablePositionECEF() ([3]float64, error) {
	v, e := native.MissingObservablePosition()
	return v, publicError(e)
}

// EmissionMediaBatch evaluates aligned satellite IDs and J2000 epochs using
// broadcast ephemerides and an ECEF receiver position in metres.
func (b *BroadcastEphemeris) EmissionMediaBatch(satellites []string, epochs []float64, receiver ECEF, options *EmissionMediaOptions) ([]EmissionMediaRow, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	var o *native.NativeEmissionMediaOptions
	if options != nil {
		o = &native.NativeEmissionMediaOptions{CarrierHz: options.CarrierHz, MinElevationEnabled: options.MinElevationEnabled, MinElevationRad: options.MinElevationRad, TroposphereEnabled: options.TroposphereEnabled, PressureHPA: options.PressureHPA, TemperatureK: options.TemperatureK, RelativeHumidity: options.RelativeHumidity}
	}
	v, e := b.handle.EmissionBatch(satellites, epochs, [3]float64{receiver.X, receiver.Y, receiver.Z}, o)
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]EmissionMediaRow, len(v))
	for i, x := range v {
		out[i] = fromNativeEmissionMedia(x)
	}
	return out, nil
}

func fromNativeEmissionMedia(x native.NativeEmissionMediaRow) EmissionMediaRow {
	return EmissionMediaRow{PositionECEFM: x.Position, HasPosition: x.HasPosition, ClockS: x.ClockS, HasClock: x.HasClock, IonosphereSlantDelayM: x.IonosphereSlantDelayM, HasIonosphereSlantDelay: x.HasIonosphereSlantDelay, TroposphereDelayM: x.TroposphereDelayM, HasTroposphereDelay: x.HasTroposphereDelay, Status: EmissionMediaStatus(x.Status), ResultStatus: StatusCode(x.ResultStatus)}
}

// EmissionMediaBatch evaluates one index-aligned satellite/epoch batch from a
// loaded SP3 product through the native engine. Epochs are J2000 seconds and
// the receiver position is ECEF metres.
func (s *SP3) EmissionMediaBatch(satellites []string, epochs []float64, receiver ECEF, options *EmissionMediaOptions) ([]EmissionMediaRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	var o *native.NativeEmissionMediaOptions
	if options != nil {
		o = &native.NativeEmissionMediaOptions{CarrierHz: options.CarrierHz, MinElevationEnabled: options.MinElevationEnabled, MinElevationRad: options.MinElevationRad, TroposphereEnabled: options.TroposphereEnabled, PressureHPA: options.PressureHPA, TemperatureK: options.TemperatureK, RelativeHumidity: options.RelativeHumidity}
	}
	v, err := s.handle.EmissionBatch(satellites, epochs, [3]float64{receiver.X, receiver.Y, receiver.Z}, o)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]EmissionMediaRow, len(v))
	for i, x := range v {
		out[i] = fromNativeEmissionMedia(x)
	}
	return out, nil
}

// EphemerisSampleStatus identifies whether a sampled ephemeris row is valid
// or represents a gap.
type EphemerisSampleStatus uint32

const (
	EphemerisSampleValid EphemerisSampleStatus = EphemerisSampleStatus(native.EphemerisSampleValidValue)
	EphemerisSampleGap   EphemerisSampleStatus = EphemerisSampleStatus(native.EphemerisSampleGapValue)
)

// EphemerisSampleRow contains a sampled satellite state at a J2000 epoch.
// Position is ECEF metres; ClockS is seconds and is optional via HasClock.
type EphemerisSampleRow struct {
	SatelliteID   string
	EpochJ2000S   float64
	Status        EphemerisSampleStatus
	PositionECEFM [3]float64
	HasPosition   bool
	ClockS        float64
	HasClock      bool
}

// ObservableStateElementStatus identifies one native observable-state result.
type ObservableStateElementStatus uint32

const (
	ObservableStateValid ObservableStateElementStatus = ObservableStateElementStatus(native.ObservableStateValidValue)
	ObservableStateGap   ObservableStateElementStatus = ObservableStateElementStatus(native.ObservableStateGapValue)
	ObservableStateError ObservableStateElementStatus = ObservableStateElementStatus(native.ObservableStateErrorValue)
)

// ObservableStateRow contains a copied ECEF position in metres and an
// optional satellite clock in seconds for one requested state.
type ObservableStateRow struct {
	PositionECEFM [3]float64
	ClockS        float64
	HasClock      bool
	ElementStatus ObservableStateElementStatus
	ResultStatus  StatusCode
}

// PredictedObservables contains geometric and signal observables. Distances
// are metres, rates metres per second, Doppler hertz, angles degrees, and
// epochs J2000 seconds; satellite clock presence is reported by HasSatelliteClock.
type PredictedObservables struct {
	GeometricRangeM        float64
	RangeRateMPerS         float64
	DopplerHz              float64
	HasSatelliteClock      bool
	SatelliteClockS        float64
	ElevationDeg           float64
	AzimuthDeg             float64
	TransmitOffsetUS       int64
	TransmitTimeJ2000S     float64
	LOSUnit                [3]float64
	SatellitePositionECEFM [3]float64
	SatelliteVelocityMPerS [3]float64
}

// PredictRequest describes one satellite prediction at a receiver ECEF
// position in metres and a receive epoch in J2000 seconds.
type PredictRequest struct {
	SatelliteID  string
	ReceiverECEF ECEF
	TRxJ2000S    float64
}

func fromNativeSample(value native.NativeEphemerisSampleRow) EphemerisSampleRow {
	return EphemerisSampleRow{SatelliteID: value.SatelliteID, EpochJ2000S: value.EpochJ2000S, Status: EphemerisSampleStatus(value.Status), PositionECEFM: value.Position, HasPosition: value.HasPosition, ClockS: value.ClockS, HasClock: value.HasClock}
}
func fromNativeState(value native.NativeObservableStateRow) ObservableStateRow {
	return ObservableStateRow{PositionECEFM: value.Position, ClockS: value.ClockS, HasClock: value.HasClock, ElementStatus: ObservableStateElementStatus(value.ElementStatus), ResultStatus: StatusCode(value.ResultStatus)}
}
func fromNativePredicted(value native.NativePredictedObservables) PredictedObservables {
	return PredictedObservables{GeometricRangeM: value.GeometricRangeM, RangeRateMPerS: value.RangeRateMPerS, DopplerHz: value.DopplerHz, HasSatelliteClock: value.HasSatelliteClock, SatelliteClockS: value.SatelliteClockS, ElevationDeg: value.ElevationDeg, AzimuthDeg: value.AzimuthDeg, TransmitOffsetUS: value.TransmitOffsetUS, TransmitTimeJ2000S: value.TransmitTimeJ2000S, LOSUnit: value.LOSUnit, SatellitePositionECEFM: value.SatellitePosition, SatelliteVelocityMPerS: value.SatelliteVelocity}
}

// EphemerisSample samples broadcast states from startJ2000S through
// stopJ2000S at stepS, with all times expressed in J2000 seconds.
func (b *BroadcastEphemeris) EphemerisSample(satellites []string, startJ2000S, stopJ2000S, stepS float64) ([]EphemerisSampleRow, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.EphemerisSample(satellites, startJ2000S, stopJ2000S, stepS)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]EphemerisSampleRow, len(values))
	for i, value := range values {
		out[i] = fromNativeSample(value)
	}
	return out, nil
}

// ObservableState returns one broadcast ECEF position in metres, clock in
// seconds, and a presence flag at a J2000 epoch.
func (b *BroadcastEphemeris) ObservableState(satelliteID string, epochJ2000S float64) ([3]float64, float64, bool, error) {
	if b == nil || b.handle == nil {
		return [3]float64{}, 0, false, ErrClosed
	}
	position, clock, present, err := b.handle.ObservableState(satelliteID, epochJ2000S)
	return position, clock, present, publicError(err)
}

// ObservableStates evaluates the index-aligned satellite and J2000 epoch
// vectors through the broadcast state route.
func (b *BroadcastEphemeris) ObservableStates(satellites []string, epochsJ2000S []float64) ([]ObservableStateRow, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.ObservableStates(satellites, epochsJ2000S)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ObservableStateRow, len(values))
	for i, value := range values {
		out[i] = fromNativeState(value)
	}
	return out, nil
}

// ObservableStatesShared evaluates one shared J2000 epoch for all satellites.
func (b *BroadcastEphemeris) ObservableStatesShared(satellites []string, epochJ2000S float64) ([]ObservableStateRow, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	values, err := b.handle.ObservableStatesShared(satellites, epochJ2000S)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ObservableStateRow, len(values))
	for i, value := range values {
		out[i] = fromNativeState(value)
	}
	return out, nil
}

// PredictObservables computes one satellite's geometric observables from an
// ECEF receiver position in metres at a J2000 receive epoch.
func (b *BroadcastEphemeris) PredictObservables(satelliteID string, receiver ECEF, epochJ2000S float64, options *ObservablesOptions) (PredictedObservables, error) {
	if b == nil || b.handle == nil {
		return PredictedObservables{}, ErrClosed
	}
	var nativeOptions *native.NativeObservablesOptions
	if options != nil {
		nativeOptions = &native.NativeObservablesOptions{CarrierHz: options.CarrierHz, LightTime: options.LightTime, Sagnac: options.Sagnac}
	}
	value, err := b.handle.PredictObservables(satelliteID, [3]float64{receiver.X, receiver.Y, receiver.Z}, epochJ2000S, nativeOptions)
	return fromNativePredicted(value), publicError(err)
}

// PredictObservablesBatch computes copied results for multiple prediction
// requests. The boolean slice reports per-request presence.
func (b *BroadcastEphemeris) PredictObservablesBatch(requests []PredictRequest, options *ObservablesOptions) ([]PredictedObservables, []bool, error) {
	if b == nil || b.handle == nil {
		return nil, nil, ErrClosed
	}
	nativeRequests := make([]native.NativePredictRequest, len(requests))
	for i, request := range requests {
		nativeRequests[i] = native.NativePredictRequest{SatelliteID: request.SatelliteID, ReceiverECEF: [3]float64{request.ReceiverECEF.X, request.ReceiverECEF.Y, request.ReceiverECEF.Z}, TRxJ2000S: request.TRxJ2000S}
	}
	var nativeOptions *native.NativeObservablesOptions
	if options != nil {
		nativeOptions = &native.NativeObservablesOptions{CarrierHz: options.CarrierHz, LightTime: options.LightTime, Sagnac: options.Sagnac}
	}
	values, ok, err := b.handle.PredictObservablesBatch(nativeRequests, nativeOptions)
	if err != nil {
		return nil, nil, publicError(err)
	}
	out := make([]PredictedObservables, len(values))
	for i, value := range values {
		out[i] = fromNativePredicted(value)
	}
	return out, append([]bool(nil), ok...), nil
}
