package sidereon

import (
	"errors"

	"sidereon.dev/go/v3/internal/native"
)

// NavMessagePreference returns the navigation-message preference selected by the broadcast ephemeris.
func (b *BroadcastEphemeris) NavMessagePreference() (uint32, error) {
	if b == nil || b.handle == nil {
		return 0, ErrClosed
	}
	v, e := b.handle.NavMessagePreference()
	return v, publicError(e)
}

// FDEWeightsMode selects the variances or weights used by the RAIM statistic.
type FDEWeightsMode uint32

const (
	// FDEWeightsSolution uses the estimator pseudorange variances to normalize residuals.
	FDEWeightsSolution FDEWeightsMode = iota
	// FDEWeightsUnit assigns unit weight to each residual.
	FDEWeightsUnit
	// FDEWeightsBySatellite uses the configured per-satellite residual weights.
	FDEWeightsBySatellite
)

// FDEOptions configures fault detection, exclusion and validation. Nil
// MaxExclusions and MaxExclusionRMSM use core defaults; non-nil zero values are
// passed explicitly. Positive infinity disables only the candidate RMS cap.
type FDEOptions struct {
	// PFA is the dimensionless false-alarm probability; zero uses the core default.
	PFA float64
	// MaxExclusions is optional; the core default is one and explicit zero disables exclusion.
	MaxExclusions *uint64
	// MaxExclusionRMSM is optional; the core default is 100 metres.
	MaxExclusionRMSM *float64
	// WeightsMode selects Solution, Unit, or BySatellite weights.
	WeightsMode FDEWeightsMode
	// Weights maps satellite IDs to inverse-variance weights in BySatellite mode.
	Weights map[string]float64
	// SystemsEnabled enables the explicit receiver clock-system count override.
	SystemsEnabled bool
	// Systems is the positive number of receiver clock parameters when enabled.
	Systems int64
	// UseValidationOptions reports whether validation options are enabled.
	UseValidationOptions bool
	// HasUseValidationOptions distinguishes an explicit false from the native default.
	HasUseValidationOptions bool
}

// DefaultFDEOptions returns native defaults for FDE thresholds, weights,
// exclusion limits, and solution validation.
func DefaultFDEOptions() (FDEOptions, error) {
	v, err := native.FDEOptionsDefault()
	return FDEOptions{PFA: v.PFA, MaxExclusions: v.MaxExclusions, MaxExclusionRMSM: v.MaxExclusionRMSM, WeightsMode: FDEWeightsMode(v.WeightsMode), SystemsEnabled: v.SystemsEnabled, Systems: v.Systems, UseValidationOptions: v.UseValidation, HasUseValidationOptions: true}, publicError(err)
}

// FDEResult owns fault-detection/exclusion results and the resulting positioning solution.
type FDEResult struct {
	_      noCopy
	handle *native.FDESolution
}

func newFDEResult(handle *native.FDESolution) (*FDEResult, error) {
	if handle == nil {
		return nil, errNilNativeHandle
	}
	return &FDEResult{handle: handle}, nil
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
	return native.NativeFDEOptions{PFA: v.PFA, MaxExclusions: v.MaxExclusions, MaxExclusionRMSM: v.MaxExclusionRMSM, WeightsMode: uint32(v.WeightsMode), Weights: nativeWeightMap(v.Weights), SystemsEnabled: v.SystemsEnabled, Systems: v.Systems, UseValidation: v.UseValidationOptions, HasUseValidation: v.HasUseValidationOptions}
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
	return newFDEResult(h)
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
	return newFDEResult(h)
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
	return newFDEResult(h)
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
	return newFDEResult(h)
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

// RAIM recomputes a test for the surviving solution using the supplied options.
// Use AcceptedRAIM to retrieve the decision that the FDE search actually used.
func (f *FDEResult) RAIM(options FDEOptions) (RAIMResult, error) {
	if f == nil || f.handle == nil {
		return RAIMResult{}, ErrClosed
	}
	pfa := options.PFA
	if pfa == 0 {
		pfa = 1e-3
	}
	v, err := f.handle.RecomputeRAIM(pfa, uint32(options.WeightsMode), nativeFDEOptions(options).Weights, options.SystemsEnabled, options.Systems)
	if err != nil {
		return RAIMResult{}, publicError(err)
	}
	normalizedCount, conversionErr := nativeCountToInt(v.NormalizedResidualCount, "FDE normalized residual count")
	if conversionErr != nil {
		return RAIMResult{}, conversionErr
	}
	return RAIMResult{v.FaultDetected, v.TestStatistic, v.HasThreshold, v.Threshold, v.HasReducedChiSquare, v.ReducedChiSquare, v.RMSM, v.DOF, v.Testable, normalizedCount, v.HasWorstSatellite, v.WorstSatellite}, nil
}

// AcceptedRAIM returns the exact detection summary and weighted residual rows
// retained by the FDE solve that selected the reported solution.
func (f *FDEResult) AcceptedRAIM() (RAIMResult, []RAIMNormalizedResidual, error) {
	if f == nil || f.handle == nil {
		return RAIMResult{}, nil, ErrClosed
	}
	v, rows, err := f.handle.AcceptedRAIM()
	if err != nil {
		return RAIMResult{}, nil, publicError(err)
	}
	result, err := publicRAIMResult(v)
	if err != nil {
		return RAIMResult{}, nil, err
	}
	out := make([]RAIMNormalizedResidual, len(rows))
	for i, row := range rows {
		out[i] = RAIMNormalizedResidual{row.SatelliteID, row.NormalizedResidual}
	}
	return result, out, nil
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
