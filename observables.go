package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// ObservablesOptions controls satellite-state to observable prediction.
// CarrierHz is in hertz; LightTime and Sagnac enable the corresponding
// Earth-fixed propagation corrections.
type ObservablesOptions struct {
	// CarrierHz is the carrier hz in hertz.
	CarrierHz float64
	// LightTime reports whether light-time correction is enabled; Sagnac reports whether Sagnac correction is enabled.
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
	// CarrierHz is the carrier hz in hertz.
	CarrierHz float64
	// MinElevationEnabled is the min elevation enabled value for EmissionMediaOptions.
	MinElevationEnabled bool
	// MinElevationRad is the min elevation rad in radians.
	MinElevationRad float64
	// TroposphereEnabled is the troposphere enabled value for EmissionMediaOptions.
	TroposphereEnabled bool
	// PressureHPA is the pressure hpa in hectopascals; TemperatureK is the temperature k in kelvin; RelativeHumidity is the relative humidity fraction.
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
	// EmissionMediaValid identifies the emission media valid case.
	EmissionMediaValid EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaValidValue)
	// EmissionMediaGap identifies the emission media gap case.
	EmissionMediaGap EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaGapValue)
	// EmissionMediaBelowElevationCutoff identifies the emission media below elevation cutoff case.
	EmissionMediaBelowElevationCutoff EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaBelowElevationValue)
	// EmissionMediaError identifies the emission media error case.
	EmissionMediaError EmissionMediaStatus = EmissionMediaStatus(native.EmissionMediaErrorValue)
)

// EmissionMediaRow contains one satellite/epoch result. Positions are ECEF
// metres and times are J2000 seconds; Has fields distinguish absent values
// from present zero values.
type EmissionMediaRow struct {
	// PositionECEFM is the position ecefm in metres.
	PositionECEFM [3]float64
	// HasPosition reports whether the has position field is present.
	HasPosition bool
	// ClockS is the clock s in seconds.
	ClockS float64
	// HasClock reports whether the has clock field is present.
	HasClock bool
	// IonosphereSlantDelayM is the ionosphere slant delay m in metres.
	IonosphereSlantDelayM float64
	// HasIonosphereSlantDelay reports whether the has ionosphere slant delay field is present.
	HasIonosphereSlantDelay bool
	// TroposphereDelayM is the troposphere delay m in metres.
	TroposphereDelayM float64
	// HasTroposphereDelay reports whether the has troposphere delay field is present.
	HasTroposphereDelay bool
	// Status is the native status code.
	Status EmissionMediaStatus
	// ResultStatus is the native result status.
	ResultStatus StatusCode
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
	// EphemerisSampleValid identifies the ephemeris sample valid case.
	EphemerisSampleValid EphemerisSampleStatus = EphemerisSampleStatus(native.EphemerisSampleValidValue)
	// EphemerisSampleGap identifies the ephemeris sample gap case.
	EphemerisSampleGap EphemerisSampleStatus = EphemerisSampleStatus(native.EphemerisSampleGapValue)
)

// EphemerisSampleRow contains a sampled satellite state at a J2000 epoch.
// Position is ECEF metres; ClockS is seconds and is optional via HasClock.
type EphemerisSampleRow struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// EpochJ2000S is the epoch j2000 s in seconds.
	EpochJ2000S float64
	// Status is the native status code.
	Status EphemerisSampleStatus
	// PositionECEFM is the position ecefm in metres.
	PositionECEFM [3]float64
	// HasPosition reports whether the has position field is present.
	HasPosition bool
	// ClockS is the clock s in seconds.
	ClockS float64
	// HasClock reports whether the has clock field is present.
	HasClock bool
}

// ObservableStateElementStatus identifies one native observable-state result.
type ObservableStateElementStatus uint32

const (
	// ObservableStateValid identifies the observable state valid case.
	ObservableStateValid ObservableStateElementStatus = ObservableStateElementStatus(native.ObservableStateValidValue)
	// ObservableStateGap identifies the observable state gap case.
	ObservableStateGap ObservableStateElementStatus = ObservableStateElementStatus(native.ObservableStateGapValue)
	// ObservableStateError identifies the observable state error case.
	ObservableStateError ObservableStateElementStatus = ObservableStateElementStatus(native.ObservableStateErrorValue)
)

// ObservableStateRow contains a copied ECEF position in metres and an
// optional satellite clock in seconds for one requested state.
type ObservableStateRow struct {
	// PositionECEFM is the position ecefm in metres.
	PositionECEFM [3]float64
	// ClockS is the clock s in seconds.
	ClockS float64
	// HasClock reports whether the has clock field is present.
	HasClock bool
	// ElementStatus is the element status value for ObservableStateRow.
	ElementStatus ObservableStateElementStatus
	// ResultStatus is the native result status.
	ResultStatus StatusCode
}

// PredictedObservables contains geometric and signal observables. Distances
// are metres, rates metres per second, Doppler hertz, angles degrees, and
// epochs J2000 seconds; satellite clock presence is reported by HasSatelliteClock.
type PredictedObservables struct {
	// GeometricRangeM is the geometric range m in metres.
	GeometricRangeM float64
	// RangeRateMPerS is the range rate m per s in metres per second.
	RangeRateMPerS float64
	// DopplerHz is the doppler hz in hertz.
	DopplerHz float64
	// HasSatelliteClock reports whether the has satellite clock field is present.
	HasSatelliteClock bool
	// SatelliteClockS is the satellite clock s in seconds.
	SatelliteClockS float64
	// ElevationDeg is the elevation deg in degrees.
	ElevationDeg float64
	// AzimuthDeg is the azimuth deg in degrees.
	AzimuthDeg float64
	// TransmitOffsetUS is the transmit offset us value for PredictedObservables.
	TransmitOffsetUS int64
	// TransmitTimeJ2000S is the transmit time j2000 s in seconds.
	TransmitTimeJ2000S float64
	// LOSUnit contains the fixed-size array for this record.
	LOSUnit [3]float64
	// SatellitePositionECEFM is the satellite position ecefm in metres.
	SatellitePositionECEFM [3]float64
	// SatelliteVelocityMPerS is the satellite velocity m per s in metres per second.
	SatelliteVelocityMPerS [3]float64
}

// PredictRequest describes one satellite prediction at a receiver ECEF
// position in metres and a receive epoch in J2000 seconds.
type PredictRequest struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// ReceiverECEF is the receiver ecef value for PredictRequest.
	ReceiverECEF ECEF
	// TRxJ2000S is the t rx j2000 s in seconds.
	TRxJ2000S float64
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
