package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// NavMessagePreference returns the navigation-message preference selected by the broadcast ephemeris.
func (b *BroadcastEphemeris) NavMessagePreference() (uint32, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	v, e := b.handle.NavMessagePreference()
	return v, publicError(e)
}

// FDEOptions contains the C RAIM/FDE controls. PFA is dimensionless;
// MaxIterations is a count; satellite weights are dimensionless; Systems is
// the distinct GNSS clock-system count when SystemsEnabled is true.
// FDEOptions configures fault-detection/exclusion thresholds and iteration limits.
type FDEOptions struct {
	PFA           float64
	MaxIterations uint64
	// UnitWeights reports whether unit measurement weights are enabled.
	UnitWeights bool
	Weights     map[string]float64
	// SystemsEnabled reports whether constellation filtering is enabled.
	SystemsEnabled bool
	// Systems identifies the GNSS constellation or constellation set.
	Systems int64
	// UseValidationOptions reports whether validation options are enabled.
	UseValidationOptions bool
}

// DefaultFDEOptions returns the C engine defaults. Weight entries can be
// supplied in the returned options when non-unit weights are requested.
// DefaultFDEOptions returns native defaults for fault-detection thresholds and iteration limits.
func DefaultFDEOptions() (FDEOptions, error) {
	v, err := native.FDEOptionsDefault()
	return FDEOptions{PFA: v.PFA, MaxIterations: v.MaxIterations, UnitWeights: v.UnitWeights, SystemsEnabled: v.SystemsEnabled, Systems: v.Systems, UseValidationOptions: v.UseValidation}, publicError(err)
}

// FDEResult owns fault-detection/exclusion results and the resulting positioning solution.
type FDEResult struct {
	_      noCopy
	handle *native.FDESolution
}

// FDEDiagnostics contains per-satellite exclusions and fault-detection iteration details.
type FDEDiagnostics struct {
	Iterations           uint64
	ExcludedSatelliteIDs []string
}

// SPPRobustConfig contains C's Huber/IRLS controls. huber_k is dimensionless,
// scale_floor_m and outer tolerance are metres, and max_outer includes the
// warm-start solve.
// SPPRobustConfig configures robust weighting for SPP residuals.
type SPPRobustConfig struct {
	HuberK, ScaleFloorM float64
	MaxOuter            int
	// OuterToleranceM contains metres.
	OuterToleranceM float64
}

func nativeFDEOptions(v FDEOptions) native.NativeFDEOptions {
	return native.NativeFDEOptions{PFA: v.PFA, MaxIterations: v.MaxIterations, UnitWeights: v.UnitWeights, Weights: nativeWeightMap(v.Weights), SystemsEnabled: v.SystemsEnabled, Systems: v.Systems, UseValidation: v.UseValidationOptions}
}

// SolveFDE computes fault-detection and exclusion diagnostics using positioning data.
func SolveFDE(sp3 *SP3, config SPPConfig, options FDEOptions) (*FDEResult, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	h, e := native.SolveFDESPP(sp3.handle, native.SPPConfig{Observations: func() []native.SPPObservation {
		v := make([]native.SPPObservation, len(config.Observations))
		for i, o := range config.Observations {
			v[i] = native.SPPObservation{SatelliteID: o.SatelliteID, PseudorangeM: o.PseudorangeM}
		}
		return v
	}(), TRxJ2000S: config.TRxJ2000S, TRxSecondOfDayS: config.TRxSecondOfDayS, DayOfYear: config.DayOfYear, InitialGuess: config.InitialGuess, Ionosphere: config.Ionosphere, Troposphere: config.Troposphere, WithGeodetic: config.WithGeodetic}, nativeFDEOptions(options))
	if e != nil {
		return nil, publicError(e)
	}
	return &FDEResult{handle: h}, nil
}

// SolveFDEBroadcast computes fault-detection and exclusion diagnostics using positioning data.
func SolveFDEBroadcast(b *BroadcastEphemeris, config SPPConfig, options FDEOptions) (*FDEResult, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	v := make([]native.SPPObservation, len(config.Observations))
	for i, o := range config.Observations {
		v[i] = native.SPPObservation{SatelliteID: o.SatelliteID, PseudorangeM: o.PseudorangeM}
	}
	h, e := native.SolveFDEBroadcast(b.handle, native.SPPConfig{Observations: v, TRxJ2000S: config.TRxJ2000S, TRxSecondOfDayS: config.TRxSecondOfDayS, DayOfYear: config.DayOfYear, InitialGuess: config.InitialGuess, Ionosphere: config.Ionosphere, Troposphere: config.Troposphere, WithGeodetic: config.WithGeodetic}, nativeFDEOptions(options))
	if e != nil {
		return nil, publicError(e)
	}
	return &FDEResult{handle: h}, nil
}

func nativeRobustConfig(v SPPRobustConfig) (native.NativeSPPRobustConfig, error) {
	if v.MaxOuter < 0 {
		return native.NativeSPPRobustConfig{}, errors.New("sidereon: robust outer-iteration count must not be negative")
	}
	return native.NativeSPPRobustConfig{HuberK: v.HuberK, ScaleFloorM: v.ScaleFloorM, MaxOuter: uint64(v.MaxOuter), OuterToleranceM: v.OuterToleranceM}, nil
}

// SolveRobustFDE computes robust fault-detection and exclusion diagnostics using positioning data.
func SolveRobustFDE(sp3 *SP3, config SPPConfig, robust SPPRobustConfig, options FDEOptions) (*FDEResult, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	r, err := nativeRobustConfig(robust)
	if err != nil {
		return nil, err
	}
	h, err := native.SolveRobustFDESPP(sp3.handle, nativeSPPConfig(config), r, nativeFDEOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &FDEResult{handle: h}, nil
}

// SolveRobustFDEBroadcast computes robust fault-detection and exclusion diagnostics using positioning data.
func SolveRobustFDEBroadcast(b *BroadcastEphemeris, config SPPConfig, robust SPPRobustConfig, options FDEOptions) (*FDEResult, error) {
	if b == nil || b.handle == nil {
		return nil, ErrClosed
	}
	r, err := nativeRobustConfig(robust)
	if err != nil {
		return nil, err
	}
	h, err := native.SolveRobustFDEBroadcast(b.handle, nativeSPPConfig(config), r, nativeFDEOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return &FDEResult{handle: h}, nil
}

// Solution returns the positioning solution produced by fault detection and exclusion.
func (f *FDEResult) Solution() (SPPSolution, error) {
	if f == nil || f.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	v, err := f.handle.Solution()
	if err != nil {
		return SPPSolution{}, publicError(err)
	}
	return publicSPPSolution(v), nil
}

// RAIM evaluates the surviving C-backed FDE solution. The returned optional
// threshold and worst-satellite fields retain their native presence flags.
// RAIM runs receiver-autonomous integrity monitoring on positioning data.
func (f *FDEResult) RAIM(options FDEOptions) (RAIMResult, error) {
	if f == nil || f.handle == nil {
		return RAIMResult{}, ErrClosed
	}
	v, err := f.handle.RAIM(options.PFA, options.UnitWeights, nativeFDEOptions(options).Weights, options.SystemsEnabled, options.Systems)
	if err != nil {
		return RAIMResult{}, publicError(err)
	}
	normalizedCount, conversionErr := nativeCountToInt(v.NormalizedResidualCount, "FDE normalized residual count")
	if conversionErr != nil {
		return RAIMResult{}, conversionErr
	}
	return RAIMResult{v.FaultDetected, v.TestStatistic, v.HasThreshold, v.Threshold, v.HasReducedChiSquare, v.ReducedChiSquare, v.RMSM, v.DOF, v.Testable, normalizedCount, v.HasWorstSatellite, v.WorstSatellite}, nil
}

// Close releases the native fault-detection/exclusion report; repeated calls are safe.
func (f *FDEResult) Close() error {
	if f == nil || f.handle == nil {
		return nil
	}
	return publicError(f.handle.Close())
}

// Diagnostics returns per-satellite exclusions and fault-detection iteration details.
func (f *FDEResult) Diagnostics() (FDEDiagnostics, error) {
	if f == nil || f.handle == nil {
		return FDEDiagnostics{}, ErrClosed
	}
	v, e := f.handle.Output()
	return FDEDiagnostics{v.Iterations, append([]string(nil), v.Excluded...)}, publicError(e)
}
