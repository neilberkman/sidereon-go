//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

type ArcEpoch struct {
	Phi1Cycles, Phi2Cycles float64
	P1M, P2M               float64
	HasLLI1                bool
	LLI1                   int64
	HasLLI2                bool
	LLI2                   int64
	F1Hz, F2Hz             float64
	GapTimeS               float64
	GapEpoch               *ExactEpoch
}

type CycleSlipOptions struct {
	GFThresholdM, MWThresholdCycles, MinArcGapS float64
}

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

func CycleSlipOptionsInit() (CycleSlipOptions, error) {
	return CycleSlipOptions{}, unavailable()
}

func DetectCycleSlips([]ArcEpoch, *CycleSlipOptions) ([]SlipResult, error) {
	return nil, unavailable()
}

func SmoothCode([]ArcEpoch, *CycleSlipOptions, int) ([]SmoothCodeResult, error) {
	return nil, unavailable()
}

func SmoothIonoFreeCode([]ArcEpoch, *CycleSlipOptions, int) ([]IonoFreeSmoothResult, error) {
	return nil, unavailable()
}
