package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RTKRINEXSignalPair selects one RINEX code and carrier observable for a GNSS
// system in a single-frequency RTK arc.
type RTKRINEXSignalPair struct {
	System                          GNSSSystem
	CodeObservable, PhaseObservable string
}

// RTKRINEXDualSignalPair selects two RINEX code/carrier pairs for a GNSS
// system in a dual-frequency RTK arc.
type RTKRINEXDualSignalPair struct {
	System                                                               GNSSSystem
	Code1Observable, Phase1Observable, Code2Observable, Phase2Observable string
}

// RTKRINEXArcOptions controls construction of a single-frequency RTK arc.
type RTKRINEXArcOptions struct {
	SignalPairs           []RTKRINEXSignalPair
	HasMaxEpochs          bool
	MaxEpochs             int
	MinCommonSatellites   int
	IncludePredictionTime bool
}

// RTKRINEXDualArcOptions controls construction of a dual-frequency RTK arc.
type RTKRINEXDualArcOptions struct {
	SignalPairs           []RTKRINEXDualSignalPair
	HasMaxEpochs          bool
	MaxEpochs             int
	MinCommonSatellites   int
	IncludePredictionTime bool
}

// RTKRINEXArcObservation is one copied single-frequency epoch observation.
type RTKRINEXArcObservation struct {
	SatelliteID, AmbiguityID string
	CodeM, PhaseM            float64
	HasLLI                   bool
	LLI                      int64
}

// RTKRINEXArcPosition is one copied satellite ECEF position.
type RTKRINEXArcPosition struct {
	SatelliteID string
	PositionM   [3]float64
}

// RTKRINEXArcEpochMetadata contains copied single-frequency epoch shape and
// optional prediction fields.
type RTKRINEXArcEpochMetadata struct {
	BaseCount, RoverCount, SatellitePositionCount           int
	BaseSatellitePositionCount, RoverSatellitePositionCount int
	HasVelocityMPS                                          bool
	VelocityMPS                                             [3]float64
	HasPredictionTime                                       bool
	PredictionTimeS                                         float64
}

// RTKRINEXMapValue is one copied ambiguity-keyed wavelength or offset. It is
// the same shape as the existing RTKFloatMapEntry input type.
type RTKRINEXMapValue = RTKFloatMapEntry

// RTKRINEXDualFrequencyObservation contains one receiver's two-frequency
// observation values.
type RTKRINEXDualFrequencyObservation struct {
	AmbiguityID                                  string
	P1M, P2M, Phi1Cycles, Phi2Cycles, F1Hz, F2Hz float64
	HasLLI1, HasLLI2                             bool
	LLI1, LLI2                                   int64
}

// RTKRINEXDualFrequencySatelliteObservation contains both receivers' values
// for one satellite.
type RTKRINEXDualFrequencySatelliteObservation struct {
	SatelliteID string
	Base, Rover RTKRINEXDualFrequencyObservation
}

// RTKRINEXDualFrequencyArcEpochMetadata contains copied dual-frequency epoch
// time, shape, and optional fields.
type RTKRINEXDualFrequencyArcEpochMetadata struct {
	JDWhole, JDFraction, GapTimeS, PredictionTimeS float64
	HasGapTimeS, HasVelocityMPS, HasPredictionTime bool
	ObservationCount, SatellitePositionCount       int
	BaseSatellitePositionCount                     int
	RoverSatellitePositionCount                    int
	VelocityMPS                                    [3]float64
}

// RTKArcObservation is the shared shape of a copied single-frequency RTK
// observation.
type RTKArcObservation = RTKRINEXArcObservation

// RTKArcPosition is the shared shape of a copied RTK satellite position.
type RTKArcPosition = RTKRINEXArcPosition

// RTKDualFrequencyObservation is the shared shape of one receiver's copied
// dual-frequency observation.
type RTKDualFrequencyObservation = RTKRINEXDualFrequencyObservation

// RTKDualFrequencySatelliteObservation is the shared shape of copied dual
// frequency observations for both receivers.
type RTKDualFrequencySatelliteObservation = RTKRINEXDualFrequencySatelliteObservation

// RTKRINEXArc owns a native single-frequency RINEX RTK arc. It must not be
// copied after first use.
type RTKRINEXArc struct {
	_      noCopy
	handle *native.RtkRinexArc
}

// RTKRINEXDualFrequencyArc owns a native dual-frequency RINEX RTK arc. It
// must not be copied after first use.
type RTKRINEXDualFrequencyArc struct {
	_      noCopy
	handle *native.RtkRinexDualFrequencyArc
}

// RINEXRTKArc is an alternate spelling of RTKRINEXArc.
type RINEXRTKArc = RTKRINEXArc

// RINEXRTKDualFrequencyArc is an alternate spelling of
// RTKRINEXDualFrequencyArc.
type RINEXRTKDualFrequencyArc = RTKRINEXDualFrequencyArc

func nativeRTKRINEXArcOptions(value RTKRINEXArcOptions) native.RtkRinexArcOptions {
	result := native.RtkRinexArcOptions{HasMaxEpochs: value.HasMaxEpochs, MaxEpochs: value.MaxEpochs, MinCommonSatellites: value.MinCommonSatellites, IncludePredictionTime: value.IncludePredictionTime}
	result.SignalPairs = make([]native.RtkRinexSignalPair, len(value.SignalPairs))
	for index, pair := range value.SignalPairs {
		result.SignalPairs[index] = native.RtkRinexSignalPair{System: uint32(pair.System), CodeObservable: pair.CodeObservable, PhaseObservable: pair.PhaseObservable}
	}
	return result
}

func nativeRTKRINEXDualArcOptions(value RTKRINEXDualArcOptions) native.RtkRinexDualArcOptions {
	result := native.RtkRinexDualArcOptions{HasMaxEpochs: value.HasMaxEpochs, MaxEpochs: value.MaxEpochs, MinCommonSatellites: value.MinCommonSatellites, IncludePredictionTime: value.IncludePredictionTime}
	result.SignalPairs = make([]native.RtkRinexDualSignalPair, len(value.SignalPairs))
	for index, pair := range value.SignalPairs {
		result.SignalPairs[index] = native.RtkRinexDualSignalPair{System: uint32(pair.System), Code1Observable: pair.Code1Observable, Phase1Observable: pair.Phase1Observable, Code2Observable: pair.Code2Observable, Phase2Observable: pair.Phase2Observable}
	}
	return result
}

// DefaultRTKRINEXArcOptions returns C's single-frequency arc defaults.
func DefaultRTKRINEXArcOptions() (RTKRINEXArcOptions, error) {
	value, err := native.RtkRinexArcOptionsInit()
	return RTKRINEXArcOptions{HasMaxEpochs: value.HasMaxEpochs, MaxEpochs: value.MaxEpochs, MinCommonSatellites: value.MinCommonSatellites, IncludePredictionTime: value.IncludePredictionTime}, publicError(err)
}

// DefaultRTKRINEXDualArcOptions returns C's dual-frequency arc defaults.
func DefaultRTKRINEXDualArcOptions() (RTKRINEXDualArcOptions, error) {
	value, err := native.RtkRinexDualArcOptionsInit()
	return RTKRINEXDualArcOptions{HasMaxEpochs: value.HasMaxEpochs, MaxEpochs: value.MaxEpochs, MinCommonSatellites: value.MinCommonSatellites, IncludePredictionTime: value.IncludePredictionTime}, publicError(err)
}

// BuildRINEXRTKArc builds a single-frequency RTK arc from parsed paired RINEX
// observations and an SP3 source.
func BuildRINEXRTKArc(sp3 *SP3, base, rover *RINEXObservation, options *RTKRINEXArcOptions) (*RTKRINEXArc, error) {
	var nativeOptions *native.RtkRinexArcOptions
	if options != nil {
		value := nativeRTKRINEXArcOptions(*options)
		nativeOptions = &value
	}
	handle, err := native.BuildRtkRinexArc(nativeSP3(sp3), nativeRinexObs(base), nativeRinexObs(rover), nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKRINEXArc{handle: handle}, nil
}

// BuildRTKRINEXArc is an alias for BuildRINEXRTKArc.
func BuildRTKRINEXArc(sp3 *SP3, base, rover *RINEXObservation, options *RTKRINEXArcOptions) (*RTKRINEXArc, error) {
	return BuildRINEXRTKArc(sp3, base, rover, options)
}

// BuildDualFrequencyRINEXRTKArc builds a dual-frequency RTK arc from parsed
// paired RINEX observations and an SP3 source.
func BuildDualFrequencyRINEXRTKArc(sp3 *SP3, base, rover *RINEXObservation, options *RTKRINEXDualArcOptions) (*RTKRINEXDualFrequencyArc, error) {
	var nativeOptions *native.RtkRinexDualArcOptions
	if options != nil {
		value := nativeRTKRINEXDualArcOptions(*options)
		nativeOptions = &value
	}
	handle, err := native.BuildRtkRinexDualFrequencyArc(nativeSP3(sp3), nativeRinexObs(base), nativeRinexObs(rover), nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKRINEXDualFrequencyArc{handle: handle}, nil
}

// BuildRINEXRTKDualFrequencyArc is an alias for
// BuildDualFrequencyRINEXRTKArc.
func BuildRINEXRTKDualFrequencyArc(sp3 *SP3, base, rover *RINEXObservation, options *RTKRINEXDualArcOptions) (*RTKRINEXDualFrequencyArc, error) {
	return BuildDualFrequencyRINEXRTKArc(sp3, base, rover, options)
}

// BuildDualFrequencyRTKRINEXArc is an alias for
// BuildDualFrequencyRINEXRTKArc.
func BuildDualFrequencyRTKRINEXArc(sp3 *SP3, base, rover *RINEXObservation, options *RTKRINEXDualArcOptions) (*RTKRINEXDualFrequencyArc, error) {
	return BuildDualFrequencyRINEXRTKArc(sp3, base, rover, options)
}

func nativeSP3(value *SP3) *native.SP3 {
	if value == nil {
		return nil
	}
	return value.handle
}

func nativeRinexObs(value *RINEXObservation) *native.RinexObs {
	if value == nil {
		return nil
	}
	return value.handle
}

// Close releases the single-frequency arc. It is idempotent and safe to call
// concurrently with accessors.
func (a *RTKRINEXArc) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return publicError(a.handle.Close())
}

// Close releases the dual-frequency arc. It is idempotent and safe to call
// concurrently with accessors.
func (a *RTKRINEXDualFrequencyArc) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return publicError(a.handle.Close())
}

func (a *RTKRINEXArc) EpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	value, err := a.handle.EpochCount()
	return value, publicError(err)
}

func (a *RTKRINEXArc) SkippedEpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	value, err := a.handle.SkippedEpochCount()
	return value, publicError(err)
}

func (a *RTKRINEXArc) EpochBaseObservations(index int) ([]RTKRINEXArcObservation, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochBaseObservations(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcObservations(value), nil
}

func (a *RTKRINEXArc) EpochRoverObservations(index int) ([]RTKRINEXArcObservation, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochRoverObservations(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcObservations(value), nil
}

func publicRTKRINEXArcObservations(value []native.RtkRinexArcObservation) []RTKRINEXArcObservation {
	result := make([]RTKRINEXArcObservation, len(value))
	for index, row := range value {
		result[index] = RTKRINEXArcObservation{SatelliteID: row.SatelliteID, AmbiguityID: row.AmbiguityID, CodeM: row.CodeM, PhaseM: row.PhaseM, HasLLI: row.HasLLI, LLI: row.LLI}
	}
	return result
}

func publicRTKRINEXArcPositions(value []native.RtkRinexArcPosition) []RTKRINEXArcPosition {
	result := make([]RTKRINEXArcPosition, len(value))
	for index, row := range value {
		result[index] = RTKRINEXArcPosition{SatelliteID: row.SatelliteID, PositionM: row.PositionM}
	}
	return result
}

func (a *RTKRINEXArc) EpochSatellitePositions(index int) ([]RTKRINEXArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcPositions(value), nil
}

func (a *RTKRINEXArc) EpochBaseSatellitePositions(index int) ([]RTKRINEXArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochBaseSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcPositions(value), nil
}

func (a *RTKRINEXArc) EpochRoverSatellitePositions(index int) ([]RTKRINEXArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochRoverSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcPositions(value), nil
}

func (a *RTKRINEXArc) EpochMetadata(index int) (RTKRINEXArcEpochMetadata, error) {
	if a == nil || a.handle == nil {
		return RTKRINEXArcEpochMetadata{}, ErrClosed
	}
	value, err := a.handle.EpochMetadata(index)
	if err != nil {
		return RTKRINEXArcEpochMetadata{}, publicError(err)
	}
	return RTKRINEXArcEpochMetadata{BaseCount: value.BaseCount, RoverCount: value.RoverCount, SatellitePositionCount: value.SatellitePositionCount, BaseSatellitePositionCount: value.BaseSatellitePositionCount, RoverSatellitePositionCount: value.RoverSatellitePositionCount, HasVelocityMPS: value.HasVelocityMPS, VelocityMPS: value.VelocityMPS, HasPredictionTime: value.HasPredictionTime, PredictionTimeS: value.PredictionTimeS}, nil
}

func publicRTKRINEXMapValues(value []native.RtkRinexMapValue) []RTKRINEXMapValue {
	result := make([]RTKRINEXMapValue, len(value))
	for index, row := range value {
		result[index] = RTKRINEXMapValue{ID: row.ID, Value: row.Value}
	}
	return result
}

func (a *RTKRINEXArc) OffsetsM() ([]RTKRINEXMapValue, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.OffsetsM()
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXMapValues(value), nil
}

func (a *RTKRINEXArc) WavelengthsM() ([]RTKRINEXMapValue, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.WavelengthsM()
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXMapValues(value), nil
}

func (a *RTKRINEXDualFrequencyArc) EpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	value, err := a.handle.EpochCount()
	return value, publicError(err)
}

func (a *RTKRINEXDualFrequencyArc) SkippedEpochCount() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	value, err := a.handle.SkippedEpochCount()
	return value, publicError(err)
}

func (a *RTKRINEXDualFrequencyArc) EpochMetadata(index int) (RTKRINEXDualFrequencyArcEpochMetadata, error) {
	if a == nil || a.handle == nil {
		return RTKRINEXDualFrequencyArcEpochMetadata{}, ErrClosed
	}
	value, err := a.handle.EpochMetadata(index)
	if err != nil {
		return RTKRINEXDualFrequencyArcEpochMetadata{}, publicError(err)
	}
	return RTKRINEXDualFrequencyArcEpochMetadata{JDWhole: value.JDWhole, JDFraction: value.JDFraction, HasGapTimeS: value.HasGapTimeS, GapTimeS: value.GapTimeS, ObservationCount: value.ObservationCount, SatellitePositionCount: value.SatellitePositionCount, BaseSatellitePositionCount: value.BaseSatellitePositionCount, RoverSatellitePositionCount: value.RoverSatellitePositionCount, HasVelocityMPS: value.HasVelocityMPS, VelocityMPS: value.VelocityMPS, HasPredictionTime: value.HasPredictionTime, PredictionTimeS: value.PredictionTimeS}, nil
}

func publicRTKRINEXDualObservations(value []native.RtkRinexDualFrequencySatelliteObservation) []RTKRINEXDualFrequencySatelliteObservation {
	result := make([]RTKRINEXDualFrequencySatelliteObservation, len(value))
	for index, row := range value {
		result[index] = RTKRINEXDualFrequencySatelliteObservation{SatelliteID: row.SatelliteID, Base: RTKRINEXDualFrequencyObservation{AmbiguityID: row.Base.AmbiguityID, P1M: row.Base.P1M, P2M: row.Base.P2M, Phi1Cycles: row.Base.Phi1Cycles, Phi2Cycles: row.Base.Phi2Cycles, F1Hz: row.Base.F1Hz, F2Hz: row.Base.F2Hz, HasLLI1: row.Base.HasLLI1, LLI1: row.Base.LLI1, HasLLI2: row.Base.HasLLI2, LLI2: row.Base.LLI2}, Rover: RTKRINEXDualFrequencyObservation{AmbiguityID: row.Rover.AmbiguityID, P1M: row.Rover.P1M, P2M: row.Rover.P2M, Phi1Cycles: row.Rover.Phi1Cycles, Phi2Cycles: row.Rover.Phi2Cycles, F1Hz: row.Rover.F1Hz, F2Hz: row.Rover.F2Hz, HasLLI1: row.Rover.HasLLI1, LLI1: row.Rover.LLI1, HasLLI2: row.Rover.HasLLI2, LLI2: row.Rover.LLI2}}
	}
	return result
}

func (a *RTKRINEXDualFrequencyArc) EpochObservations(index int) ([]RTKRINEXDualFrequencySatelliteObservation, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochObservations(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXDualObservations(value), nil
}

func (a *RTKRINEXDualFrequencyArc) EpochSatellitePositions(index int) ([]RTKRINEXArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcPositions(value), nil
}

func (a *RTKRINEXDualFrequencyArc) EpochBaseSatellitePositions(index int) ([]RTKRINEXArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochBaseSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcPositions(value), nil
}

func (a *RTKRINEXDualFrequencyArc) EpochRoverSatellitePositions(index int) ([]RTKRINEXArcPosition, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.EpochRoverSatellitePositions(index)
	if err != nil {
		return nil, publicError(err)
	}
	return publicRTKRINEXArcPositions(value), nil
}

func (a *RTKRINEXDualFrequencyArc) EpochSortKey(index int) (string, error) {
	if a == nil || a.handle == nil {
		return "", ErrClosed
	}
	value, err := a.handle.EpochSortKey(index)
	return value, publicError(err)
}
