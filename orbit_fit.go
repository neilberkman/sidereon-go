package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// OrbitFitOptions controls precise-ephemeris initial-state fitting. Lengths
// in residual diagnostics are metres; the fitted state uses km and km/s.
// OrbitFitOptions configures force model, integrator, tolerances, and residual-ledger limits.
type OrbitFitOptions struct {
	ForceModel      PropagationForceModel
	ForceComponents ForceModelComponents
	// MuKm3PerS2Enabled reports whether MuKm3PerS2 is supplied.
	MuKm3PerS2Enabled bool
	MuKm3PerS2        float64
	Integrator        PropagationIntegrator
	// AbsTol is the absolute integration tolerance.
	AbsTol float64
	// RelTol is the relative integration tolerance.
	RelTol float64
	// InitialStepS is the initial integration step in seconds.
	InitialStepS float64
	// MinStepS is the minimum integration step in seconds.
	MinStepS float64
	// MaxStepS is the maximum integration step in seconds.
	MaxStepS float64
	// MaxSteps is the maximum integration step count.
	MaxSteps uint32
	// SolverGTol is the gradient stopping tolerance.
	SolverGTol float64
	// SolverFTol is the cost stopping tolerance.
	SolverFTol float64
	// SolverXTol is the step stopping tolerance.
	SolverXTol float64
	// SolverMaxNFEV is the solver function-evaluation limit.
	SolverMaxNFEV    int
	MinLedgerSamples int
	// HasDrag reports whether drag parameters are supplied.
	HasDrag bool
	Drag    DragParameters
}

// PreciseEphemerisSample is one precise-orbit sample with epoch, ECEF position, and clock value.
type PreciseEphemerisSample struct {
	// Satellite identifies the GNSS satellite associated with this record.
	Satellite string
	TimeScale TimeScale
	// EpochJ2000S is the sample epoch in seconds from J2000.
	EpochJ2000S float64
	// PositionECEFM contains metres.
	PositionECEFM [3]float64
	// HasClock reports whether a clock value is supplied.
	HasClock bool
	// ClockS is the optional satellite-clock value in seconds when HasClock is true.
	ClockS float64
	// ClockEvent reports whether the sample marks a clock event.
	ClockEvent bool
}

// OrbitFitCovarianceKind identifies whether the orbit-fit covariance is estimated or unbounded.
type OrbitFitCovarianceKind uint32

const (
	// OrbitFitCovarianceEstimated marks a covariance matrix estimated by the orbit fit.
	OrbitFitCovarianceEstimated OrbitFitCovarianceKind = 0
	// OrbitFitCovarianceUnbounded marks an unbounded covariance returned when estimation is unavailable.
	OrbitFitCovarianceUnbounded OrbitFitCovarianceKind = 1
)

// OrbitFitCovariance contains the row-major 6x6 orbit-fit covariance and its mixed position/velocity units.
type OrbitFitCovariance struct {
	Kind OrbitFitCovarianceKind
	// Matrix is the row-major 6x6 covariance for [x, y, z, vx, vy, vz], with position and velocity units.
	Matrix [36]float64
}

// GeometryQuality reports orbit-fit observability tier and redundancy.
type GeometryQuality struct {
	Tier       ObservabilityTier
	Redundancy int32
	// Rank is the fitted design-matrix rank.
	Rank int
	// ConditionNumber is the design-matrix condition number.
	ConditionNumber float64
	// GDOP is the geometric dilution of precision.
	GDOP float64
	// RAIMCheckable reports whether RAIM can test the solution.
	RAIMCheckable bool
	// CovarianceValidated reports whether the covariance passed validation.
	CovarianceValidated bool
}

// OrbitFitSolution contains the fitted orbit initial state, covariance, geometry quality, and iteration count.
type OrbitFitSolution struct {
	// Satellite identifies the GNSS satellite associated with this record.
	Satellite       string
	InitialState    CartesianState
	Covariance      OrbitFitCovariance
	GeometryQuality GeometryQuality
	// SeedRMS3DM contains metres.
	SeedRMS3DM float64
	// FitRMS3DM contains metres.
	FitRMS3DM  float64
	Iterations int
}

// OrbitResidualStats contains position residual statistics in metres for an orbit-fit ledger.
type OrbitResidualStats struct {
	// RadialRMSM contains metres.
	RadialRMSM float64
	// AlongRMSM contains metres.
	AlongRMSM float64
	// CrossRMSM contains metres.
	CrossRMSM float64
	// RMS3DM contains metres.
	RMS3DM float64
	// N is the number of residual samples contributing to the statistics.
	N int
	// LowSampleCount reports whether too few samples were available.
	LowSampleCount bool
}

// OrbitSatelliteResidualEntry associates one satellite with its orbit-fit residual statistics.
type OrbitSatelliteResidualEntry struct {
	// Satellite identifies the GNSS satellite associated with this record.
	Satellite string
	Stats     OrbitResidualStats
}

// OrbitConstellationResidualEntry associates one GNSS constellation with its orbit-fit residual statistics.
type OrbitConstellationResidualEntry struct {
	// System identifies the GNSS constellation or constellation set.
	System uint32
	Stats  OrbitResidualStats
}

// OrbitArcSpan contains the start/end epochs and duration of an orbit-fit arc.
type OrbitArcSpan struct {
	TimeScale TimeScale
	// StartJ2000S is the arc start in seconds from J2000.
	StartJ2000S float64
	// EndJ2000S is the arc end in seconds from J2000.
	EndJ2000S float64
	// DurationS is the arc duration in seconds.
	DurationS float64
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

// OrbitFitOptionsDefaults returns native default force-model, integrator, and tolerance settings.
func OrbitFitOptionsDefaults() (OrbitFitOptions, error) {
	value, err := native.OrbitFitOptionsDefaults()
	return publicOrbitFitOptions(value), publicError(err)
}

// FitSP3PreciseOrbit fits a propagated Cartesian orbit to the supplied SP3 samples.
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

// FitSP3ECEFPreciseOrbit fits a precise ECEF orbit to one SP3 satellite.
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

// FitSP3ECEFPreciseOrbits fits propagated Cartesian orbits to the supplied SP3 samples for all selected satellites.
func FitSP3ECEFPreciseOrbits(sp3 *SP3, satellites []string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	return fitSP3ECEFPreciseOrbits(sp3, satellites, options, false)
}

// FitAllSP3ECEFPreciseOrbits fits propagated Cartesian orbits to every satellite in the supplied SP3 product.
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

// FitPreciseEphemerisSamples fits a propagated Cartesian orbit to precise ephemeris samples.
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

// Close releases the native orbit-fit report; repeated calls are safe.
func (r *OrbitFitReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Fits returns detached fitted-orbit solutions from the report.
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

// SatelliteLedger returns detached residual statistics grouped by satellite.
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

// ConstellationLedger returns detached residual statistics grouped by GNSS constellation.
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

// ArcSpan returns the fitted orbit arc time span.
func (r *OrbitFitReport) ArcSpan() (OrbitArcSpan, error) {
	if r == nil || r.handle == nil {
		return OrbitArcSpan{}, ErrClosed
	}
	value, err := r.handle.ArcSpan()
	return OrbitArcSpan{TimeScale: TimeScale(value.TimeScale), StartJ2000S: value.StartJ2000S, EndJ2000S: value.EndJ2000S, DurationS: value.DurationS}, publicError(err)
}
