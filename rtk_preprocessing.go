package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// ArcEpoch is one dual-frequency carrier/code epoch. Missing scalar values
// are represented by NaN, as in the native ABI.
type ArcEpoch struct {
	Phi1Cycles, Phi2Cycles float64
	P1M, P2M               float64
	HasLLI1                bool
	LLI1                   int64
	HasLLI2                bool
	LLI2                   int64
	F1Hz, F2Hz             float64
	GapTimeS               float64
}

// CycleSlipOptions controls geometry-free and Melbourne-Wubbena detection.
type CycleSlipOptions struct {
	GFThresholdM, MWThresholdCycles, MinArcGapS float64
}

// SlipResult is the native classification for one input epoch.
type SlipResult struct {
	Slip       bool
	ReasonMask uint32
	GFM, MWM   float64
	Skipped    bool
}

type SmoothCodeResult struct {
	PSmoothM float64
	Window   int
	Reset    bool
}

type IonoFreeSmoothResult struct {
	PSmoothM, PIFM, LIFM float64
	Window               int
	Reset                bool
}

func nativeArcEpoch(value ArcEpoch) native.ArcEpoch {
	return native.ArcEpoch{Phi1Cycles: value.Phi1Cycles, Phi2Cycles: value.Phi2Cycles, P1M: value.P1M, P2M: value.P2M, HasLLI1: value.HasLLI1, LLI1: value.LLI1, HasLLI2: value.HasLLI2, LLI2: value.LLI2, F1Hz: value.F1Hz, F2Hz: value.F2Hz, GapTimeS: value.GapTimeS}
}

func nativeCycleSlipOptions(value *CycleSlipOptions) *native.CycleSlipOptions {
	if value == nil {
		return nil
	}
	return &native.CycleSlipOptions{GFThresholdM: value.GFThresholdM, MWThresholdCycles: value.MWThresholdCycles, MinArcGapS: value.MinArcGapS}
}

func nativeArcEpochs(values []ArcEpoch) []native.ArcEpoch {
	result := make([]native.ArcEpoch, len(values))
	for i, value := range values {
		result[i] = nativeArcEpoch(value)
	}
	return result
}

func CycleSlipOptionsDefault() (CycleSlipOptions, error) {
	value, err := native.CycleSlipOptionsInit()
	return CycleSlipOptions{GFThresholdM: value.GFThresholdM, MWThresholdCycles: value.MWThresholdCycles, MinArcGapS: value.MinArcGapS}, publicError(err)
}

func DetectCycleSlips(epochs []ArcEpoch, options *CycleSlipOptions) ([]SlipResult, error) {
	values, err := native.DetectCycleSlips(nativeArcEpochs(epochs), nativeCycleSlipOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]SlipResult, len(values))
	for i, value := range values {
		result[i] = SlipResult{Slip: value.Slip, ReasonMask: value.ReasonMask, GFM: value.GFM, MWM: value.MWM, Skipped: value.Skipped}
	}
	return result, nil
}

func SmoothCode(epochs []ArcEpoch, options *CycleSlipOptions, hatchWindowCap int) ([]SmoothCodeResult, error) {
	values, err := native.SmoothCode(nativeArcEpochs(epochs), nativeCycleSlipOptions(options), hatchWindowCap)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]SmoothCodeResult, len(values))
	for i, value := range values {
		result[i] = SmoothCodeResult{PSmoothM: value.PSmoothM, Window: value.Window, Reset: value.Reset}
	}
	return result, nil
}

func SmoothIonoFreeCode(epochs []ArcEpoch, options *CycleSlipOptions, hatchWindowCap int) ([]IonoFreeSmoothResult, error) {
	values, err := native.SmoothIonoFreeCode(nativeArcEpochs(epochs), nativeCycleSlipOptions(options), hatchWindowCap)
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]IonoFreeSmoothResult, len(values))
	for i, value := range values {
		result[i] = IonoFreeSmoothResult{PSmoothM: value.PSmoothM, PIFM: value.PIFM, LIFM: value.LIFM, Window: value.Window, Reset: value.Reset}
	}
	return result, nil
}

type PseudorangeObservation struct {
	SatelliteID  string
	PseudorangeM float64
}

type IonoFreeOverride struct {
	System byte
	Band1  string
	Band2  string
}

type IonoFreeCombined struct {
	SatelliteID  string
	PseudorangeM float64
}

type IonoFreeDropped struct {
	SatelliteID string
	Reason      uint32
}

// IonoFreePseudoranges owns the native result until Close. Accessors always
// return detached Go copies and are safe to call concurrently with Close.
type IonoFreePseudoranges struct {
	_      noCopy
	handle *native.IonoFreePseudoranges
}

func CombineIonosphereFreePseudoranges(band1, band2 []PseudorangeObservation, overrides []IonoFreeOverride) (*IonoFreePseudoranges, error) {
	first := make([]native.PseudorangeObservation, len(band1))
	for i, value := range band1 {
		first[i] = native.PseudorangeObservation{SatelliteID: value.SatelliteID, PseudorangeM: value.PseudorangeM}
	}
	second := make([]native.PseudorangeObservation, len(band2))
	for i, value := range band2 {
		second[i] = native.PseudorangeObservation{SatelliteID: value.SatelliteID, PseudorangeM: value.PseudorangeM}
	}
	converted := make([]native.IonoFreeOverride, len(overrides))
	for i, value := range overrides {
		converted[i] = native.IonoFreeOverride{System: value.System, Band1: value.Band1, Band2: value.Band2}
	}
	handle, err := native.CombineIonosphereFreePseudoranges(first, second, converted)
	if err != nil {
		return nil, publicError(err)
	}
	return &IonoFreePseudoranges{handle: handle}, nil
}

func (r *IonoFreePseudoranges) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

func (r *IonoFreePseudoranges) Combined() ([]IonoFreeCombined, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	values, err := r.handle.Combined()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]IonoFreeCombined, len(values))
	for i, value := range values {
		result[i] = IonoFreeCombined{SatelliteID: value.SatelliteID, PseudorangeM: value.PseudorangeM}
	}
	return result, nil
}

func (r *IonoFreePseudoranges) Dropped() ([]IonoFreeDropped, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	values, err := r.handle.Dropped()
	if err != nil {
		return nil, publicError(err)
	}
	result := make([]IonoFreeDropped, len(values))
	for i, value := range values {
		result[i] = IonoFreeDropped{SatelliteID: value.SatelliteID, Reason: value.Reason}
	}
	return result, nil
}
