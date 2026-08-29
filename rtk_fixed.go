package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// RTKAmbiguitySatellite associates an ambiguity identifier with a satellite.
type RTKAmbiguitySatellite struct {
	// ID is the ambiguity key; SatelliteID is its GNSS satellite identifier.
	ID, SatelliteID string
}

// RTKFloatMapEntry is one ambiguity-keyed wavelength or metre offset.
type RTKFloatMapEntry struct {
	// ID is the ambiguity key; Value is the copied metre value.
	ID    string
	Value float64
}

// RTKReceiverAntennaNoAzimuthPCV is one zenith-dependent receiver PCV sample.
type RTKReceiverAntennaNoAzimuthPCV struct {
	// ZenithDeg is the zenith angle; ValueM is the correction in metres.
	ZenithDeg, ValueM float64
}

// RTKReceiverAntennaAzimuthPCV is one azimuth- and zenith-dependent receiver
// PCV sample.
type RTKReceiverAntennaAzimuthPCV struct {
	// AzimuthDeg and ZenithDeg are angles in degrees; ValueM is metres.
	AzimuthDeg, ZenithDeg, ValueM float64
}

// RTKReceiverAntennaCalibration contains one frequency's PCO/PCV correction.
// PCO is in local north/east/up metres.
type RTKReceiverAntennaCalibration struct {
	// PCONEUM is phase-center offset in local north/east/up metres.
	PCONEUM      [3]float64
	NoAzimuthPCV []RTKReceiverAntennaNoAzimuthPCV
	AzimuthPCV   []RTKReceiverAntennaAzimuthPCV
}

// RTKReceiverAntennaCorrections contains base and rover calibrations.
type RTKReceiverAntennaCorrections struct {
	// Base and Rover contain independent copied antenna calibrations.
	Base, Rover RTKReceiverAntennaCalibration
}

// RTKFixedConfig contains all inputs to the static integer-fixed RTK solve.
// Use the DefaultRTK* functions to obtain engine defaults before overriding
// option fields.
type RTKFixedConfig struct {
	// Epochs is the copied RTK observation sequence.
	Epochs              []RTKEpoch
	BaseECEFM           [3]float64
	AmbiguityIDs        []string
	AmbiguitySatellites []RTKAmbiguitySatellite
	Wavelengths         []RTKFloatMapEntry
	Offsets             []RTKFloatMapEntry
	Model               RTKMeasurementModel
	ReceiverAntenna     *RTKReceiverAntennaCorrections
	FloatOptions        RTKFloatOptions
	FixedOptions        RTKFixedOptions
	ResidualOptions     RTKResidualValidationOptions
	FloatOnlySystems    []GNSSSystem
	InitialBaselineM    [3]float64
}

// RTKIntegerStatus is the fixed solver's integer-search verdict.
type RTKIntegerStatus uint32

const (
	// RTKIntegerFixed indicates an integer-fixed solution.
	RTKIntegerFixed RTKIntegerStatus = 0
	// RTKIntegerNotFixed indicates that integer fixing was not achieved.
	RTKIntegerNotFixed RTKIntegerStatus = 1
)

// RTKFixedAmbiguity is one accepted integer ambiguity in cycles and metres.
type RTKFixedAmbiguity struct {
	// ID identifies the ambiguity; Cycles is integer cycles and ValueM is metres.
	ID     string
	Cycles int64
	ValueM float64
}

// RTKFixedMetadata contains copied fixed-solve and integer-search metadata.
type RTKFixedMetadata struct {
	// Iteration, observation, ambiguity, residual, and satellite fields are native counts.
	Iterations, NObservations, FreeAmbiguityCount, FixedAmbiguityCount int
	ResidualCount, UsedSatCount                                        int
	Converged                                                          bool
	Status                                                             uint32
	CodeRMSM, PhaseRMSM, WeightedRMSM                                  float64
	IntegerStatus                                                      RTKIntegerStatus
	HasIntegerRatio                                                    bool
	IntegerRatio                                                       float64
	HasIntegerBestScore                                                bool
	IntegerBestScore                                                   float64
	HasIntegerSecondBestScore                                          bool
	IntegerSecondBestScore                                             float64
	IntegerCandidates                                                  int
	GeometryQuality                                                    SPPGeometryQuality
}

// RTKFixedSolution owns a native static fixed RTK solution. It must not be
// copied after first use.
type RTKFixedSolution struct {
	_      noCopy
	handle *native.RtkFixedSolution
}

func nativeRTKFixedConfig(value RTKFixedConfig) native.RtkFixedConfig {
	result := native.RtkFixedConfig{
		BaseECEFM:        value.BaseECEFM,
		AmbiguityIDs:     append([]string(nil), value.AmbiguityIDs...),
		Model:            native.RtkMeasurementModel{CodeSigmaM: value.Model.CodeSigmaM, PhaseSigmaM: value.Model.PhaseSigmaM, Sagnac: value.Model.Sagnac, Stochastic: value.Model.Stochastic, ElevationWeighting: value.Model.ElevationWeighting},
		FloatOptions:     native.RtkFloatOptions{PositionTolM: value.FloatOptions.PositionTolM, AmbiguityTolM: value.FloatOptions.AmbiguityTolM, MaxIterations: value.FloatOptions.MaxIterations},
		FixedOptions:     native.RtkFixedOptions{PositionTolM: value.FixedOptions.PositionTolM, AmbiguityTolM: value.FixedOptions.AmbiguityTolM, MaxIterations: value.FixedOptions.MaxIterations, RatioThreshold: value.FixedOptions.RatioThreshold, PartialAmbiguityResolution: value.FixedOptions.PartialAmbiguityResolution, PartialMinAmbiguities: value.FixedOptions.PartialMinAmbiguities},
		ResidualOptions:  native.RtkResidualValidationOptions{ThresholdSigmaEnabled: value.ResidualOptions.ThresholdSigmaEnabled, ThresholdSigma: value.ResidualOptions.ThresholdSigma, MaxExclusions: value.ResidualOptions.MaxExclusions},
		FloatOnlySystems: make([]uint32, len(value.FloatOnlySystems)),
		InitialBaselineM: value.InitialBaselineM,
	}
	result.Epochs = make([]native.RtkEpoch, len(value.Epochs))
	for i, epoch := range value.Epochs {
		result.Epochs[i] = native.RtkEpoch{HasVelocityMPS: epoch.HasVelocityMPS, VelocityMPS: epoch.VelocityMPS, DTS: epoch.DTS}
		result.Epochs[i].References = make([]native.RtkSatMeasurement, len(epoch.References))
		result.Epochs[i].NonReference = make([]native.RtkSatMeasurement, len(epoch.NonReference))
		for j, row := range epoch.References {
			result.Epochs[i].References[j] = native.RtkSatMeasurement{SatelliteID: row.SatelliteID, SDAmbiguityID: row.SDAmbiguityID, BaseCodeM: row.BaseCodeM, BasePhaseM: row.BasePhaseM, RoverCodeM: row.RoverCodeM, RoverPhaseM: row.RoverPhaseM, BaseTXPos: row.BaseTXPos, RoverTXPos: row.RoverTXPos, Pos: row.Pos}
		}
		for j, row := range epoch.NonReference {
			result.Epochs[i].NonReference[j] = native.RtkSatMeasurement{SatelliteID: row.SatelliteID, SDAmbiguityID: row.SDAmbiguityID, BaseCodeM: row.BaseCodeM, BasePhaseM: row.BasePhaseM, RoverCodeM: row.RoverCodeM, RoverPhaseM: row.RoverPhaseM, BaseTXPos: row.BaseTXPos, RoverTXPos: row.RoverTXPos, Pos: row.Pos}
		}
	}
	result.AmbiguitySatellites = make([]native.RtkAmbiguitySatellite, len(value.AmbiguitySatellites))
	for i, row := range value.AmbiguitySatellites {
		result.AmbiguitySatellites[i] = native.RtkAmbiguitySatellite{ID: row.ID, SatelliteID: row.SatelliteID}
	}
	result.Wavelengths = make([]native.RtkFloatMapEntry, len(value.Wavelengths))
	for i, row := range value.Wavelengths {
		result.Wavelengths[i] = native.RtkFloatMapEntry{ID: row.ID, Value: row.Value}
	}
	result.Offsets = make([]native.RtkFloatMapEntry, len(value.Offsets))
	for i, row := range value.Offsets {
		result.Offsets[i] = native.RtkFloatMapEntry{ID: row.ID, Value: row.Value}
	}
	for i, system := range value.FloatOnlySystems {
		result.FloatOnlySystems[i] = uint32(system)
	}
	if value.ReceiverAntenna != nil {
		result.ReceiverAntenna = &native.RtkReceiverAntennaCorrections{
			Base:  nativeRTKReceiverAntennaCalibration(value.ReceiverAntenna.Base),
			Rover: nativeRTKReceiverAntennaCalibration(value.ReceiverAntenna.Rover),
		}
	}
	return result
}

func nativeRTKReceiverAntennaCalibration(value RTKReceiverAntennaCalibration) native.RtkReceiverAntennaCalibration {
	result := native.RtkReceiverAntennaCalibration{PCONEUM: value.PCONEUM, NoAzimuthPCV: make([]native.RtkReceiverAntennaNoAzimuthPCV, len(value.NoAzimuthPCV)), AzimuthPCV: make([]native.RtkReceiverAntennaAzimuthPCV, len(value.AzimuthPCV))}
	for i, sample := range value.NoAzimuthPCV {
		result.NoAzimuthPCV[i] = native.RtkReceiverAntennaNoAzimuthPCV{ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	for i, sample := range value.AzimuthPCV {
		result.AzimuthPCV[i] = native.RtkReceiverAntennaAzimuthPCV{AzimuthDeg: sample.AzimuthDeg, ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	return result
}

// SolveRTKFixed solves a static integer-fixed RTK baseline through the C ABI.
func SolveRTKFixed(config RTKFixedConfig) (*RTKFixedSolution, error) {
	handle, err := native.SolveRtkFixed(nativeRTKFixedConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTKFixedSolution{handle: handle}, nil
}

// Close releases the fixed solution. It is idempotent and safe to call
// concurrently with read methods.
func (s *RTKFixedSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// FixedAmbiguities returns copied integer-fixed ambiguities.
func (s *RTKFixedSolution) FixedAmbiguities() ([]RTKFixedAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.FixedAmbiguities()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKFixedAmbiguity, len(value))
	for i, row := range value {
		result[i] = RTKFixedAmbiguity{ID: row.ID, Cycles: row.Cycles, ValueM: row.ValueM}
	}
	return result, nil
}

// FixedBaselineECEF returns the fixed rover-minus-base ECEF baseline.
func (s *RTKFixedSolution) FixedBaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FixedBaselineECEF()
	return value, publicError(err)
}

// FixedBaselineENU returns the fixed baseline in local East-North-Up axes.
func (s *RTKFixedSolution) FixedBaselineENU() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FixedBaselineENU()
	return value, publicError(err)
}

// BaselineECEF is the fixed rover-minus-base ECEF baseline. It is an alias
// for FixedBaselineECEF matching RTKFloatSolution's naming.
func (s *RTKFixedSolution) BaselineECEF() ([3]float64, error) {
	return s.FixedBaselineECEF()
}

// BaselineENU is the fixed baseline in local East-North-Up axes. It is an
// alias for FixedBaselineENU matching RTKFloatSolution's naming.
func (s *RTKFixedSolution) BaselineENU() ([3]float64, error) {
	return s.FixedBaselineENU()
}

// FloatBaselineECEF returns the float baseline used by the fixed solve.
func (s *RTKFixedSolution) FloatBaselineECEF() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FloatBaselineECEF()
	return value, publicError(err)
}

// FreeAmbiguities returns copied ambiguities left float by the fixed solve.
func (s *RTKFixedSolution) FreeAmbiguities() ([]RTKAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.FreeAmbiguities()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]RTKAmbiguity, len(value))
	for i, row := range value {
		result[i] = RTKAmbiguity{ID: row.ID, ValueM: row.ValueM}
	}
	return result, nil
}

// Metadata returns copied solver and integer-search metadata.
func (s *RTKFixedSolution) Metadata() (RTKFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return RTKFixedMetadata{}, ErrClosed
	}
	value, err := s.handle.Metadata()
	return RTKFixedMetadata{Iterations: value.Iterations, NObservations: value.NObservations, FreeAmbiguityCount: value.FreeAmbiguityCount, FixedAmbiguityCount: value.FixedAmbiguityCount, ResidualCount: value.ResidualCount, UsedSatCount: value.UsedSatCount, Converged: value.Converged, Status: value.Status, CodeRMSM: value.CodeRMSM, PhaseRMSM: value.PhaseRMSM, WeightedRMSM: value.WeightedRMSM, IntegerStatus: RTKIntegerStatus(value.IntegerStatus), HasIntegerRatio: value.HasIntegerRatio, IntegerRatio: value.IntegerRatio, HasIntegerBestScore: value.HasIntegerBestScore, IntegerBestScore: value.IntegerBestScore, HasIntegerSecondBestScore: value.HasIntegerSecondBestScore, IntegerSecondBestScore: value.IntegerSecondBestScore, IntegerCandidates: value.IntegerCandidates, GeometryQuality: SPPGeometryQuality{Tier: value.GeometryQuality.Tier, Redundancy: value.GeometryQuality.Redundancy, Rank: value.GeometryQuality.Rank, ConditionNumber: value.GeometryQuality.ConditionNumber, GDOP: value.GeometryQuality.GDOP, RAIMCheckable: value.GeometryQuality.RAIMCheckable, CovarianceValidated: value.GeometryQuality.CovarianceValidated}}, publicError(err)
}

// UsedSatelliteIDs returns copied satellite identifiers in native order.
func (s *RTKFixedSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.UsedSatelliteIDs()
	return append([]string(nil), value...), publicError(err)
}

// NoAzimuthPCV and AzimuthPCV are copied optional correction tables.
// BaseECEFM and InitialBaselineM are ECEF baseline inputs in metres.
// AmbiguityIDs and AmbiguitySatellites define the copied ambiguity graph.
// Wavelengths and Offsets are copied ambiguity-keyed metre values.
// Model and option fields select native float/fixed validation behavior.
// ReceiverAntenna is optional copied antenna calibration data.
// FloatOnlySystems excludes selected constellations from integer fixing.
// Converged and Status preserve the native fixed-solve status.
// RMS fields are metre-valued residual diagnostics.
// IntegerStatus and optional score fields describe ambiguity search results.
