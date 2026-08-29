package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// ScenarioObservation is one copied synthetic observation row.
type ScenarioObservation struct {
	EpochIndex         int
	SatelliteID        string
	CodeObservable     string
	PhaseObservable    string
	DopplerObservable  string
	CarrierHz          float64
	PseudorangeM       float64
	CarrierPhaseCycles float64
	DopplerHz          float64
}

// ScenarioReceiverTruthRow is one copied receiver truth row.
type ScenarioReceiverTruthRow struct {
	TRxJ2000S       float64
	PositionECEFM   [3]float64
	VelocityECEFMPS [3]float64
	ClockM          float64
	ClockRateMPS    float64
}

// ScenarioSummary contains deterministic simulation counts and fingerprint.
type ScenarioSummary struct {
	SchemaVersion          uint32
	Seed                   uint64
	ReceiverTruthCount     int
	ObservationCount       int
	EpochOffsetCount       int
	DeterminismFingerprint uint64
	JSONLength             int
}

// ScenarioTerm is one copied ground-truth error-budget row.
type ScenarioTerm struct {
	GeometricRangeM                       float64
	SatelliteClockM                       float64
	ReceiverClockM                        float64
	SatelliteClockErrorM                  float64
	IonosphereM                           float64
	TroposphereM                          float64
	ThermalNoiseM                         float64
	MultipathM                            float64
	QuantizationM                         float64
	CarrierPhaseGeometricCycles           float64
	CarrierPhaseReceiverClockCycles       float64
	CarrierPhaseSatelliteClockCycles      float64
	CarrierPhaseSatelliteClockErrorCycles float64
	CarrierPhaseIonosphereCycles          float64
	CarrierPhaseTroposphereCycles         float64
	CarrierPhaseThermalNoiseCycles        float64
	CarrierPhaseBiasCycles                float64
	CarrierPhaseQuantizationCycles        float64
	DopplerSatelliteMotionHz              float64
	DopplerReceiverMotionHz               float64
	DopplerSatelliteClockHz               float64
	DopplerReceiverClockHz                float64
	DopplerSatelliteClockErrorHz          float64
	DopplerThermalNoiseHz                 float64
	DopplerQuantizationHz                 float64
}

// ScenarioSimulation owns a deterministic C-generated simulation and must
// not be copied after first use. Read methods are safe to call concurrently
// with Close.
type ScenarioSimulation struct {
	_      noCopy
	handle *native.ScenarioSimulation
}

func publicScenarioSimulation(value *native.ScenarioSimulation) *ScenarioSimulation {
	if value == nil {
		return nil
	}
	return &ScenarioSimulation{handle: value}
}

// ScenarioSimulateJSON runs the synthetic scenario described by JSON bytes.
func ScenarioSimulateJSON(data []byte) (*ScenarioSimulation, error) {
	value, err := native.ScenarioSimulateJSON(data)
	return publicScenarioSimulation(value), publicError(err)
}

// ScenarioSimulateJSONWithBroadcast runs a scenario using a parsed broadcast
// ephemeris product.
func ScenarioSimulateJSONWithBroadcast(data []byte, broadcast *BroadcastEphemeris) (*ScenarioSimulation, error) {
	if broadcast == nil {
		return nil, ErrClosed
	}
	value, err := native.ScenarioSimulateJSONWithBroadcast(data, broadcast.handle)
	return publicScenarioSimulation(value), publicError(err)
}

// ScenarioSimulateJSONWithBroadcastAndIonex runs a scenario using parsed
// broadcast and IONEX products.
func ScenarioSimulateJSONWithBroadcastAndIonex(data []byte, broadcast *BroadcastEphemeris, ionex *IONEX) (*ScenarioSimulation, error) {
	if broadcast == nil || ionex == nil {
		return nil, ErrClosed
	}
	value, err := native.ScenarioSimulateJSONWithBroadcastAndIonex(data, broadcast.handle, ionex.handle)
	return publicScenarioSimulation(value), publicError(err)
}

// ScenarioSimulateJSONWithIonex runs a scenario using a parsed IONEX product.
func ScenarioSimulateJSONWithIonex(data []byte, ionex *IONEX) (*ScenarioSimulation, error) {
	if ionex == nil {
		return nil, ErrClosed
	}
	value, err := native.ScenarioSimulateJSONWithIonex(data, ionex.handle)
	return publicScenarioSimulation(value), publicError(err)
}

// ScenarioSimulateJSONWithSP3 runs a scenario using a parsed SP3 product.
func ScenarioSimulateJSONWithSP3(data []byte, sp3 *SP3) (*ScenarioSimulation, error) {
	if sp3 == nil {
		return nil, ErrClosed
	}
	value, err := native.ScenarioSimulateJSONWithSP3(data, sp3.handle)
	return publicScenarioSimulation(value), publicError(err)
}

// ScenarioSimulateJSONWithSP3AndIonex runs a scenario using parsed SP3 and
// IONEX products.
func ScenarioSimulateJSONWithSP3AndIonex(data []byte, sp3 *SP3, ionex *IONEX) (*ScenarioSimulation, error) {
	if sp3 == nil || ionex == nil {
		return nil, ErrClosed
	}
	value, err := native.ScenarioSimulateJSONWithSP3AndIonex(data, sp3.handle, ionex.handle)
	return publicScenarioSimulation(value), publicError(err)
}

// Close releases the native simulation and is idempotent.
func (s *ScenarioSimulation) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// ScenarioEpochOffsets returns copied epoch offsets.
func ScenarioEpochOffsets(s *ScenarioSimulation) ([]int, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.EpochOffsets()
	return value, publicError(err)
}

// EpochOffsets is the method form of ScenarioEpochOffsets.
func (s *ScenarioSimulation) EpochOffsets() ([]int, error) { return ScenarioEpochOffsets(s) }

// ScenarioObservations returns copied synthetic observation rows.
func ScenarioObservations(s *ScenarioSimulation) ([]ScenarioObservation, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.Observations()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ScenarioObservation, len(values))
	for i, value := range values {
		out[i] = ScenarioObservation{EpochIndex: value.EpochIndex, SatelliteID: value.SatelliteID, CodeObservable: value.CodeObservable, PhaseObservable: value.PhaseObservable, DopplerObservable: value.DopplerObservable, CarrierHz: value.CarrierHz, PseudorangeM: value.PseudorangeM, CarrierPhaseCycles: value.CarrierPhaseCycles, DopplerHz: value.DopplerHz}
	}
	return out, nil
}

// Observations is the method form of ScenarioObservations.
func (s *ScenarioSimulation) Observations() ([]ScenarioObservation, error) {
	return ScenarioObservations(s)
}

// ScenarioReceiverTruth returns copied receiver truth rows.
func ScenarioReceiverTruth(s *ScenarioSimulation) ([]ScenarioReceiverTruthRow, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.ReceiverTruth()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ScenarioReceiverTruthRow, len(values))
	for i, value := range values {
		out[i] = ScenarioReceiverTruthRow{TRxJ2000S: value.TRxJ2000S, PositionECEFM: value.PositionECEFM, VelocityECEFMPS: value.VelocityECEFMPS, ClockM: value.ClockM, ClockRateMPS: value.ClockRateMPS}
	}
	return out, nil
}

// ReceiverTruth is the method form of ScenarioReceiverTruth.
func (s *ScenarioSimulation) ReceiverTruth() ([]ScenarioReceiverTruthRow, error) {
	return ScenarioReceiverTruth(s)
}

// ScenarioSimulationJSON returns C's detached JSON serialization.
func ScenarioSimulationJSON(s *ScenarioSimulation) ([]byte, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	value, err := s.handle.JSON()
	return value, publicError(err)
}

// JSON is the method form of ScenarioSimulationJSON.
func (s *ScenarioSimulation) JSON() ([]byte, error) { return ScenarioSimulationJSON(s) }

// ScenarioSimulationSummary returns copied simulation summary metadata.
func ScenarioSimulationSummary(s *ScenarioSimulation) (ScenarioSummary, error) {
	if s == nil || s.handle == nil {
		return ScenarioSummary{}, ErrClosed
	}
	value, err := s.handle.Summary()
	return ScenarioSummary{SchemaVersion: value.SchemaVersion, Seed: value.Seed, ReceiverTruthCount: value.ReceiverTruthCount, ObservationCount: value.ObservationCount, EpochOffsetCount: value.EpochOffsetCount, DeterminismFingerprint: value.DeterminismFingerprint, JSONLength: value.JSONLength}, publicError(err)
}

// Summary is the method form of ScenarioSimulationSummary.
func (s *ScenarioSimulation) Summary() (ScenarioSummary, error) { return ScenarioSimulationSummary(s) }

// ScenarioTerms returns copied ground-truth term rows.
func ScenarioTerms(s *ScenarioSimulation) ([]ScenarioTerm, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.Terms()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]ScenarioTerm, len(values))
	for i, value := range values {
		out[i] = ScenarioTerm{
			GeometricRangeM: value.GeometricRangeM, SatelliteClockM: value.SatelliteClockM, ReceiverClockM: value.ReceiverClockM, SatelliteClockErrorM: value.SatelliteClockErrorM,
			IonosphereM: value.IonosphereM, TroposphereM: value.TroposphereM, ThermalNoiseM: value.ThermalNoiseM, MultipathM: value.MultipathM, QuantizationM: value.QuantizationM,
			CarrierPhaseGeometricCycles: value.CarrierPhaseGeometricCycles, CarrierPhaseReceiverClockCycles: value.CarrierPhaseReceiverClockCycles, CarrierPhaseSatelliteClockCycles: value.CarrierPhaseSatelliteClockCycles, CarrierPhaseSatelliteClockErrorCycles: value.CarrierPhaseSatelliteClockErrorCycles,
			CarrierPhaseIonosphereCycles: value.CarrierPhaseIonosphereCycles, CarrierPhaseTroposphereCycles: value.CarrierPhaseTroposphereCycles, CarrierPhaseThermalNoiseCycles: value.CarrierPhaseThermalNoiseCycles, CarrierPhaseBiasCycles: value.CarrierPhaseBiasCycles, CarrierPhaseQuantizationCycles: value.CarrierPhaseQuantizationCycles,
			DopplerSatelliteMotionHz: value.DopplerSatelliteMotionHz, DopplerReceiverMotionHz: value.DopplerReceiverMotionHz, DopplerSatelliteClockHz: value.DopplerSatelliteClockHz, DopplerReceiverClockHz: value.DopplerReceiverClockHz, DopplerSatelliteClockErrorHz: value.DopplerSatelliteClockErrorHz, DopplerThermalNoiseHz: value.DopplerThermalNoiseHz, DopplerQuantizationHz: value.DopplerQuantizationHz,
		}
	}
	return out, nil
}

// Terms is the method form of ScenarioTerms.
func (s *ScenarioSimulation) Terms() ([]ScenarioTerm, error) { return ScenarioTerms(s) }
