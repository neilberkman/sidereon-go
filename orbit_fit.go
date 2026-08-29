package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// OrbitFitOptions controls precise-ephemeris initial-state fitting. Lengths
// in residual diagnostics are metres; the fitted state uses km and km/s.
type OrbitFitOptions struct {
	ForceModel        PropagationForceModel
	ForceComponents   ForceModelComponents
	MuKm3PerS2Enabled bool
	MuKm3PerS2        float64
	Integrator        PropagationIntegrator
	AbsTol            float64
	RelTol            float64
	InitialStepS      float64
	MinStepS          float64
	MaxStepS          float64
	MaxSteps          uint32
	SolverGTol        float64
	SolverFTol        float64
	SolverXTol        float64
	SolverMaxNFEV     int
	MinLedgerSamples  int
	HasDrag           bool
	Drag              DragParameters
}

type PreciseEphemerisSample struct {
	Satellite     string
	TimeScale     TimeScale
	EpochJ2000S   float64
	PositionECEFM [3]float64
	HasClock      bool
	ClockS        float64
	ClockEvent    bool
}

type OrbitFitCovarianceKind uint32

const (
	OrbitFitCovarianceEstimated OrbitFitCovarianceKind = 0
	OrbitFitCovarianceUnbounded OrbitFitCovarianceKind = 1
)

type OrbitFitCovariance struct {
	Kind   OrbitFitCovarianceKind
	Matrix [36]float64
}

type GeometryQuality struct {
	Tier                ObservabilityTier
	Redundancy          int32
	Rank                int
	ConditionNumber     float64
	GDOP                float64
	RAIMCheckable       bool
	CovarianceValidated bool
}

type OrbitFitSolution struct {
	Satellite       string
	InitialState    CartesianState
	Covariance      OrbitFitCovariance
	GeometryQuality GeometryQuality
	SeedRMS3DM      float64
	FitRMS3DM       float64
	Iterations      int
}

type OrbitResidualStats struct {
	RadialRMSM     float64
	AlongRMSM      float64
	CrossRMSM      float64
	RMS3DM         float64
	N              int
	LowSampleCount bool
}

type OrbitSatelliteResidualEntry struct {
	Satellite string
	Stats     OrbitResidualStats
}

type OrbitConstellationResidualEntry struct {
	System uint32
	Stats  OrbitResidualStats
}

type OrbitArcSpan struct {
	TimeScale   TimeScale
	StartJ2000S float64
	EndJ2000S   float64
	DurationS   float64
}

// OrbitFitReport owns a C fit report and may not be copied after use.
type OrbitFitReport struct {
	_      noCopy
	handle *native.OrbitFitReport
}

func nativeOrbitFitOptions(value OrbitFitOptions) native.OrbitFitOptions {
	return native.OrbitFitOptions{
		ForceModel: uint32(value.ForceModel), ForceComponents: nativeForceComponents(value.ForceComponents),
		MuKm3PerS2Enabled: value.MuKm3PerS2Enabled, MuKm3PerS2: value.MuKm3PerS2,
		Integrator: uint32(value.Integrator), AbsTol: value.AbsTol, RelTol: value.RelTol, InitialStepS: value.InitialStepS, MinStepS: value.MinStepS, MaxStepS: value.MaxStepS, MaxSteps: value.MaxSteps,
		SolverGTol: value.SolverGTol, SolverFTol: value.SolverFTol, SolverXTol: value.SolverXTol, SolverMaxNFEV: value.SolverMaxNFEV, MinLedgerSamples: value.MinLedgerSamples,
		HasDrag: value.HasDrag, Drag: native.DragParameters{BCFactorM2PerKg: value.Drag.BCFactorM2PerKg, Weather: nativeSpaceWeather(value.Drag.Weather), CutoffAltitudeKm: value.Drag.CutoffAltitudeKm},
	}
}

func publicOrbitFitOptions(value native.OrbitFitOptions) OrbitFitOptions {
	return OrbitFitOptions{
		ForceModel: PropagationForceModel(value.ForceModel), ForceComponents: publicForceComponents(value.ForceComponents),
		MuKm3PerS2Enabled: value.MuKm3PerS2Enabled, MuKm3PerS2: value.MuKm3PerS2, Integrator: PropagationIntegrator(value.Integrator),
		AbsTol: value.AbsTol, RelTol: value.RelTol, InitialStepS: value.InitialStepS, MinStepS: value.MinStepS, MaxStepS: value.MaxStepS, MaxSteps: value.MaxSteps,
		SolverGTol: value.SolverGTol, SolverFTol: value.SolverFTol, SolverXTol: value.SolverXTol, SolverMaxNFEV: value.SolverMaxNFEV, MinLedgerSamples: value.MinLedgerSamples,
		HasDrag: value.HasDrag, Drag: publicDragParameters(value.Drag),
	}
}

func nativePreciseSample(value PreciseEphemerisSample) native.PreciseEphemerisSample {
	return native.PreciseEphemerisSample{Satellite: value.Satellite, TimeScale: uint32(value.TimeScale), EpochJ2000S: value.EpochJ2000S, PositionECEFM: value.PositionECEFM, HasClock: value.HasClock, ClockS: value.ClockS, ClockEvent: value.ClockEvent}
}

func publicOrbitFit(value *native.OrbitFitReport) *OrbitFitReport {
	if value == nil {
		return nil
	}
	return &OrbitFitReport{handle: value}
}

func OrbitFitOptionsDefaults() (OrbitFitOptions, error) {
	value, err := native.OrbitFitOptionsDefaults()
	return publicOrbitFitOptions(value), publicError(err)
}

func FitSP3PreciseOrbit(sp3 *SP3, satellite string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var nativeOptions *native.OrbitFitOptions
	if options != nil {
		value := nativeOrbitFitOptions(*options)
		nativeOptions = &value
	}
	value, err := native.FitSP3PreciseOrbit(sp3.handle, satellite, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOrbitFit(value), nil
}

func FitSP3ECEFPreciseOrbit(sp3 *SP3, satellite string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var nativeOptions *native.OrbitFitOptions
	if options != nil {
		value := nativeOrbitFitOptions(*options)
		nativeOptions = &value
	}
	value, err := native.FitSP3ECEFPreciseOrbit(sp3.handle, satellite, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOrbitFit(value), nil
}

func FitSP3ECEFPreciseOrbits(sp3 *SP3, satellites []string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	return fitSP3ECEFPreciseOrbits(sp3, satellites, options, false)
}

func FitAllSP3ECEFPreciseOrbits(sp3 *SP3, options *OrbitFitOptions) (*OrbitFitReport, error) {
	return fitSP3ECEFPreciseOrbits(sp3, nil, options, true)
}

func fitSP3ECEFPreciseOrbits(sp3 *SP3, satellites []string, options *OrbitFitOptions, all bool) (*OrbitFitReport, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var nativeOptions *native.OrbitFitOptions
	if options != nil {
		value := nativeOrbitFitOptions(*options)
		nativeOptions = &value
	}
	value, err := native.FitSP3ECEFPreciseOrbits(sp3.handle, append([]string(nil), satellites...), nativeOptions, all)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOrbitFit(value), nil
}

func FitPreciseEphemerisSamples(samples []PreciseEphemerisSample, satellite string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	values := make([]native.PreciseEphemerisSample, len(samples))
	for i := range samples {
		values[i] = nativePreciseSample(samples[i])
	}
	var nativeOptions *native.OrbitFitOptions
	if options != nil {
		value := nativeOrbitFitOptions(*options)
		nativeOptions = &value
	}
	value, err := native.FitPreciseEphemerisSamples(values, satellite, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOrbitFit(value), nil
}

func (r *OrbitFitReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

func (r *OrbitFitReport) Fits() ([]OrbitFitSolution, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	values, err := r.handle.Fits()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]OrbitFitSolution, len(values))
	for i, value := range values {
		out[i] = OrbitFitSolution{Satellite: value.Satellite, InitialState: publicCartesian(value.InitialState), Covariance: OrbitFitCovariance{Kind: OrbitFitCovarianceKind(value.Covariance.Kind), Matrix: value.Covariance.Matrix}, GeometryQuality: GeometryQuality{Tier: ObservabilityTier(value.GeometryQuality.Tier), Redundancy: value.GeometryQuality.Redundancy, Rank: value.GeometryQuality.Rank, ConditionNumber: value.GeometryQuality.ConditionNumber, GDOP: value.GeometryQuality.GDOP, RAIMCheckable: value.GeometryQuality.RAIMCheckable, CovarianceValidated: value.GeometryQuality.CovarianceValidated}, SeedRMS3DM: value.SeedRMS3DM, FitRMS3DM: value.FitRMS3DM, Iterations: value.Iterations}
	}
	return out, nil
}

func (r *OrbitFitReport) SatelliteLedger() ([]OrbitSatelliteResidualEntry, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	values, err := r.handle.SatelliteLedger()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]OrbitSatelliteResidualEntry, len(values))
	for i, value := range values {
		out[i] = OrbitSatelliteResidualEntry{Satellite: value.Satellite, Stats: publicOrbitResidualStats(value.Stats)}
	}
	return out, nil
}

func (r *OrbitFitReport) ConstellationLedger() ([]OrbitConstellationResidualEntry, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	values, err := r.handle.ConstellationLedger()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]OrbitConstellationResidualEntry, len(values))
	for i, value := range values {
		out[i] = OrbitConstellationResidualEntry{System: value.System, Stats: publicOrbitResidualStats(value.Stats)}
	}
	return out, nil
}

func publicOrbitResidualStats(value native.OrbitResidualStats) OrbitResidualStats {
	return OrbitResidualStats{RadialRMSM: value.RadialRMSM, AlongRMSM: value.AlongRMSM, CrossRMSM: value.CrossRMSM, RMS3DM: value.RMS3DM, N: value.N, LowSampleCount: value.LowSampleCount}
}

func (r *OrbitFitReport) ArcSpan() (OrbitArcSpan, error) {
	if r == nil || r.handle == nil {
		return OrbitArcSpan{}, ErrClosed
	}
	value, err := r.handle.ArcSpan()
	return OrbitArcSpan{TimeScale: TimeScale(value.TimeScale), StartJ2000S: value.StartJ2000S, EndJ2000S: value.EndJ2000S, DurationS: value.DurationS}, publicError(err)
}
