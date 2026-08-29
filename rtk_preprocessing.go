package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// ArcEpoch is one dual-frequency carrier/code epoch. Missing scalar values
// are represented by NaN, as in the native ABI.
// ArcEpoch contains rover/base observations and satellite positions for one preprocessing epoch.
type ArcEpoch struct {
	// Phi1Cycles and Phi2Cycles are the first-frequency carrier phase in cycles, the second-frequency carrier phase in cycles.
	Phi1Cycles, Phi2Cycles float64
	// P1M and P2M are the two code pseudoranges in metres.
	P1M, P2M float64
	// HasLLI1 reports whether LLI1 is valid.
	HasLLI1 bool
	// LLI1 is the first-frequency loss-of-lock indicator.
	LLI1 int64
	// HasLLI2 reports whether LLI2 is valid.
	HasLLI2 bool
	// LLI2 is the second-frequency loss-of-lock indicator.
	LLI2 int64
	// F1Hz and F2Hz are the first and second carrier frequencies in hertz.
	F1Hz, F2Hz float64
	// GapTimeS is the cycle-slip gap threshold in seconds.
	GapTimeS float64
}

// CycleSlipOptions controls geometry-free and Melbourne-Wubbena detection.
type CycleSlipOptions struct {
	// GFThresholdM and MWThresholdCycles and MinArcGapS are the geometry-free threshold in metres, the Melbourne-Wubbena threshold in cycles, the minimum arc gap in seconds.
	GFThresholdM, MWThresholdCycles, MinArcGapS float64
}

// SlipResult is the native classification for one input epoch.
type SlipResult struct {
	// Slip reports whether a cycle slip was detected.
	Slip bool
	// ReasonMask is the bit mask of cycle-slip reasons.
	ReasonMask uint32
	// GFM is the geometry-free code/phase combination in metres; MWM is the Melbourne-Wübbena combination in cycles.
	GFM, MWM float64
	// Skipped reports whether this observation was skipped.
	Skipped bool
}

// SmoothCodeResult contains one Hatch-smoothed code observation and validity metadata.
type SmoothCodeResult struct {
	// PSmoothM contains metres.
	PSmoothM float64
	// Window is the smoothing window in epochs.
	Window int
	// Reset reports whether smoothing reset at this epoch.
	Reset bool
}

// IonoFreeSmoothResult contains one smoothed ionosphere-free code and carrier observation.
type IonoFreeSmoothResult struct {
	// PSmoothM is the Hatch-smoothed code in metres; PIFM is the ionosphere-free code in metres; LIFM is the ionosphere-free carrier in metres.
	PSmoothM, PIFM, LIFM float64
	// Window is the smoothing window in epochs.
	Window int
	// Reset reports whether smoothing reset at this epoch.
	Reset bool
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

// CycleSlipOptionsDefault returns defaults for cycle-slip detection and smoothing.
func CycleSlipOptionsDefault() (CycleSlipOptions, error) {
	value, err := native.CycleSlipOptionsInit()
	return CycleSlipOptions{GFThresholdM: value.GFThresholdM, MWThresholdCycles: value.MWThresholdCycles, MinArcGapS: value.MinArcGapS}, publicError(err)
}

// DetectCycleSlips identifies cycle-slip events in observation epochs.
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

// SmoothCode computes Hatch-smoothed code observations.
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

// SmoothIonoFreeCode computes Hatch-smoothed ionosphere-free code observations.
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

// PseudorangeObservation supplies one satellite pseudorange and carrier-frequency metadata.
type PseudorangeObservation struct {
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// PseudorangeM contains metres.
	PseudorangeM float64
}

// IonoFreeOverride supplies a per-satellite or per-system ionosphere-free wavelength override.
type IonoFreeOverride struct {
	// System identifies the GNSS constellation or constellation set.
	System byte
	// Band1 is the first signal-band identifier.
	Band1 string
	// Band2 is the second signal-band identifier.
	Band2 string
}

// IonoFreeCombined contains one successfully combined ionosphere-free observation.
type IonoFreeCombined struct {
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	// PseudorangeM contains metres.
	PseudorangeM float64
}

// IonoFreeDropped identifies an observation omitted from ionosphere-free combination and its reason.
type IonoFreeDropped struct {
	// SatelliteID is the GNSS satellite identifier.
	SatelliteID string
	Reason      uint32
}

// IonoFreePseudoranges owns the native result until Close. Accessors always
// return detached Go copies and are safe to call concurrently with Close.
// IonoFreePseudoranges owns native ionosphere-free pseudorange results.
type IonoFreePseudoranges struct {
	_      noCopy
	handle *native.IonoFreePseudoranges
}

// CombineIonosphereFreePseudoranges combines two code bands into ionosphere-free pseudoranges.
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

// Close releases the native ionosphere-free pseudorange result; repeated calls are safe.
func (r *IonoFreePseudoranges) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Combined returns detached successfully combined ionosphere-free observations.
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

// Dropped returns detached observations omitted during ionosphere-free combination.
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
