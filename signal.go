package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SignalModulationKind selects the native signal-analysis model.
type SignalModulationKind uint32

const (
	SignalBPSK           SignalModulationKind = SignalModulationKind(native.SignalModulationBPSKValue)
	SignalBOCSine        SignalModulationKind = SignalModulationKind(native.SignalModulationBOCSineValue)
	SignalBOCCosine      SignalModulationKind = SignalModulationKind(native.SignalModulationBOCCosineValue)
	SignalMBOC611Over11  SignalModulationKind = SignalModulationKind(native.SignalModulationMBOCValue)
	SignalTMBOC614Over33 SignalModulationKind = SignalModulationKind(native.SignalModulationTMBOCValue)
	SignalCBOCPlus       SignalModulationKind = SignalModulationKind(native.SignalModulationCBOCPlusValue)
	SignalCBOCMinus      SignalModulationKind = SignalModulationKind(native.SignalModulationCBOCMinusValue)
)

// SignalModulation is a value descriptor consumed by C's signal-analysis
// routines. For BPSK use Order; for BOC use M and N.
type SignalModulation struct {
	Kind        SignalModulationKind
	Order, M, N float64
}

// BPSK describes a binary phase-shift keyed signal with the given order.
func BPSK(order float64) SignalModulation {
	return SignalModulation{Kind: SignalBPSK, Order: order}
}

// BOCSine describes a sine-phased binary offset-carrier signal.
func BOCSine(m, n float64) SignalModulation {
	return SignalModulation{Kind: SignalBOCSine, M: m, N: n}
}

// BOCCosine describes a cosine-phased binary offset-carrier signal.
func BOCCosine(m, n float64) SignalModulation {
	return SignalModulation{Kind: SignalBOCCosine, M: m, N: n}
}
func fixedModulation(kind SignalModulationKind) SignalModulation {
	return SignalModulation{Kind: kind}
}

// MBOC611Over11 returns the fixed MBOC(6,1,1/11) modulation.
func MBOC611Over11() SignalModulation {
	return fixedModulation(SignalMBOC611Over11)
}

// TMBOC614Over33 returns the fixed TMBOC(6,1,4/33) modulation.
func TMBOC614Over33() SignalModulation {
	return fixedModulation(SignalTMBOC614Over33)
}

// CBOCPlus returns the fixed positive CBOC modulation.
func CBOCPlus() SignalModulation {
	return fixedModulation(SignalCBOCPlus)
}

// CBOCMinus returns the fixed negative CBOC modulation.
func CBOCMinus() SignalModulation {
	return fixedModulation(SignalCBOCMinus)
}
func nativeModulation(v SignalModulation) native.NativeSignalModulation {
	return native.NativeSignalModulation{Kind: uint32(v.Kind), Order: v.Order, M: v.M, N: v.N}
}

// IQSample is one complex baseband sample. I and Q are dimensionless sample
// amplitudes.
type IQSample struct {
	I float64
	Q float64
}

// AcquisitionOptions defines the sample rate and Doppler search grid. All
// frequency fields are in hertz.
type AcquisitionOptions struct {
	SampleRateHz  float64
	DopplerMinHz  float64
	DopplerMaxHz  float64
	DopplerStepHz float64
}

// AcquisitionResult contains the selected acquisition peak and the native
// grid dimensions. Code phase is in chips and Doppler is in hertz.
type AcquisitionResult struct {
	CodePhaseChips, DopplerHz, PeakMetric, Metric, PeakPower float64
	GridCodePhaseBins                                        int
	GridDopplerStepHz, GridSamplesPerChip                    float64
	GridDopplerBinCount                                      int
}

func toNativeIQ(v []IQSample) []native.NativeIQSample {
	out := make([]native.NativeIQSample, len(v))
	for i, x := range v {
		out[i] = native.NativeIQSample{I: x.I, Q: x.Q}
	}
	return out
}

// AcquireSignal searches the supplied IQ samples for a PRN and returns the
// copied correlation metric grid.
func AcquireSignal(samples []IQSample, prn int64, options AcquisitionOptions) (AcquisitionResult, []float64, error) {
	r, b, e := native.SignalAcquire(toNativeIQ(samples), prn, native.NativeAcquisitionOptions{SampleRateHz: options.SampleRateHz, DopplerMinHz: options.DopplerMinHz, DopplerMaxHz: options.DopplerMaxHz, DopplerStepHz: options.DopplerStepHz})
	return AcquisitionResult{CodePhaseChips: r.CodePhaseChips, DopplerHz: r.DopplerHz, PeakMetric: r.PeakMetric, Metric: r.Metric, PeakPower: r.PeakPower, GridCodePhaseBins: r.GridCodePhaseBins, GridDopplerStepHz: r.GridDopplerStepHz, GridSamplesPerChip: r.GridSamplesPerChip, GridDopplerBinCount: r.GridDopplerBinCount}, append([]float64(nil), b...), publicError(e)
}

// CAChip returns one GPS C/A-code chip as a signed binary value.
func CAChip(prn, index int64) (int8, error) {
	v, e := native.SignalCAChip(prn, index)
	return v, publicError(e)
}

// CACode returns a detached GPS C/A-code sequence for a PRN.
func CACode(prn int64) ([]int8, error) {
	v, e := native.SignalCACode(prn)
	return append([]int8(nil), v...), publicError(e)
}

// ReplicaOptions controls generated C/A-code samples. Rate is hertz, phase
// is chips, Doppler is hertz, and NumSamples is the requested count.
type ReplicaOptions struct {
	SampleRateHz                  float64
	NumSamples                    int
	CodePhaseChips, CodeDopplerHz float64
}

// SignalReplica generates a detached sampled C/A-code replica.
func SignalReplica(prn int64, options ReplicaOptions) ([]int8, error) {
	v, e := native.SignalReplica(prn, native.NativeReplicaOptions{SampleRateHz: options.SampleRateHz, NumSamples: options.NumSamples, CodePhaseChips: options.CodePhaseChips, CodeDopplerHz: options.CodeDopplerHz})
	return append([]int8(nil), v...), publicError(e)
}

// CorrelateOptions controls one code correlation. Rate and Dopplers are in
// hertz; code phase is in chips.
type CorrelateOptions struct {
	SampleRateHz   float64
	DopplerHz      float64
	CodePhaseChips float64
	CodeDopplerHz  float64
}

// CorrelationResult contains the complex correlation and its power.
type CorrelationResult struct {
	I     float64
	Q     float64
	Power float64
}

// CorrelateSignal correlates IQ samples against a PRN code replica.
func CorrelateSignal(samples []IQSample, prn int64, options CorrelateOptions) (CorrelationResult, error) {
	v, e := native.SignalCorrelate(toNativeIQ(samples), prn, native.NativeCorrelateOptions{SampleRateHz: options.SampleRateHz, DopplerHz: options.DopplerHz, CodePhaseChips: options.CodePhaseChips, CodeDopplerHz: options.CodeDopplerHz})
	return CorrelationResult{I: v.I, Q: v.Q, Power: v.Power}, publicError(e)
}

// CorrelateAgainst correlates IQ samples against an explicit code sequence.
// fs and doppler are in hertz.
func CorrelateAgainst(samples []IQSample, code []int8, fs, doppler float64) (float64, float64, error) {
	i, q, e := native.SignalCorrelateAgainst(toNativeIQ(samples), code, fs, doppler)
	return i, q, publicError(e)
}

// CorrelationAt returns the signed correlation at one chip lag.
func CorrelationAt(a, b []int8, lag int64) (int32, error) {
	v, e := native.SignalCorrelationAt(a, b, lag)
	return v, publicError(e)
}

// CrossCorrelation returns a detached cross-correlation sequence.
func CrossCorrelation(a, b []int8) ([]int32, error) {
	v, e := native.SignalCrossCorrelation(a, b)
	return append([]int32(nil), v...), publicError(e)
}

// Autocorrelation returns a detached autocorrelation sequence.
func Autocorrelation(code []int8) ([]int32, error) {
	v, e := native.SignalAutocorrelation(code)
	return append([]int32(nil), v...), publicError(e)
}

// ReferenceChipRateHz returns the native reference chip rate in hertz.
func ReferenceChipRateHz() (float64, error) {
	v, e := native.SignalReferenceChipRate()
	return v, publicError(e)
}

// BetzL1ReceiverBandwidthHz returns the Betz L1 receiver bandwidth in hertz.
func BetzL1ReceiverBandwidthHz() (float64, error) {
	v, e := native.SignalBetzBandwidth()
	return v, publicError(e)
}

// CoherentLoss returns linear coherent integration loss for hertz and seconds.
func CoherentLoss(freqErrorHz, integrationTimeS float64) (float64, error) {
	v, e := native.SignalCoherentLoss(freqErrorHz, integrationTimeS)
	return v, publicError(e)
}

// CoherentLossDB returns coherent integration loss in dB.
func CoherentLossDB(freqErrorHz, integrationTimeS float64) (float64, error) {
	v, e := native.SignalCoherentLossDB(freqErrorHz, integrationTimeS)
	return v, publicError(e)
}

// PostCorrelationSNRDB returns post-correlation SNR in dB from CN0 in dBHz.
func PostCorrelationSNRDB(cn0DBHz, integrationTimeS float64) (float64, error) {
	v, e := native.SignalSNRPost(cn0DBHz, integrationTimeS)
	return v, publicError(e)
}

// ModulationLabel returns the native display name for a modulation.
func ModulationLabel(mod SignalModulation) (string, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return "", err
	}
	v, e := native.SignalModulationLabel(nativeModulation(mod))
	return v, publicError(e)
}

// ModulationCodeRateHz returns a modulation's code rate in hertz.
func ModulationCodeRateHz(mod SignalModulation) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalModulationCodeRate(nativeModulation(mod))
	return v, publicError(e)
}

// ModulationPSDHz returns the modulation PSD at an offset in hertz.
func ModulationPSDHz(mod SignalModulation, offsetHz float64) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalPSD(nativeModulation(mod), offsetHz)
	return v, publicError(e)
}

// AnalysisPSDHz returns the analysis PSD at an offset in hertz.
func AnalysisPSDHz(mod SignalModulation, offsetHz float64) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalAnalysisPSD(nativeModulation(mod), offsetHz)
	return v, publicError(e)
}

// PowerInBand returns integrated signal power over a bandwidth in hertz.
func PowerInBand(mod SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalPowerInBand(nativeModulation(mod), bandwidthHz)
	return v, publicError(e)
}

// FractionPowerInBand returns the in-band power fraction for a bandwidth in
// hertz.
func FractionPowerInBand(mod SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalFractionPowerInBand(nativeModulation(mod), bandwidthHz)
	return v, publicError(e)
}

// RMSBandwidthHz returns RMS bandwidth in hertz.
func RMSBandwidthHz(mod SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalRMSBandwidth(nativeModulation(mod), bandwidthHz)
	return v, publicError(e)
}

// AnalysisRMSBandwidthHz returns analysis RMS bandwidth in hertz.
func AnalysisRMSBandwidthHz(mod SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalAnalysisRMSBandwidth(nativeModulation(mod), bandwidthHz)
	return v, publicError(e)
}

// DLLProcessing selects coherent or non-coherent DLL processing.
type DLLProcessing uint32

const (
	DLLCoherent    DLLProcessing = DLLProcessing(native.DLLCoherentValue)
	DLLNonCoherent DLLProcessing = DLLProcessing(native.DLLNonCoherentValue)
)

// DLLTrackingOptions describes a delay-lock loop in dBHz, hertz, seconds,
// chips, and hertz respectively.
type DLLTrackingOptions struct {
	CN0DBHz                float64
	LoopBandwidthHz        float64
	IntegrationTimeS       float64
	CorrelatorSpacingChips float64
	ReceiverBandwidthHz    float64
}

// DLLJitter contains equivalent delay jitter in seconds, chips, and metres;
// SquaringLoss is a dimensionless linear factor.
type DLLJitter struct {
	Seconds      float64
	Chips        float64
	Meters       float64
	SquaringLoss float64
}

func nativeDLL(v DLLTrackingOptions) native.NativeDLLTrackingOptions {
	return native.NativeDLLTrackingOptions{CN0DBHz: v.CN0DBHz, LoopBandwidthHz: v.LoopBandwidthHz, IntegrationTimeS: v.IntegrationTimeS, CorrelatorSpacingChips: v.CorrelatorSpacingChips, ReceiverBandwidthHz: v.ReceiverBandwidthHz}
}
func fromDLL(v native.NativeDLLJitter) DLLJitter {
	return DLLJitter{Seconds: v.Seconds, Chips: v.Chips, Meters: v.Meters, SquaringLoss: v.SquaringLoss}
}

// DLLThermalNoiseJitter returns thermal-noise DLL jitter for the selected
// processing mode.
func DLLThermalNoiseJitter(mod SignalModulation, opt DLLTrackingOptions, processing DLLProcessing) (DLLJitter, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return DLLJitter{}, err
	}
	if err := validateDLLProcessing(processing); err != nil {
		return DLLJitter{}, err
	}
	v, e := native.SignalDLL(nativeModulation(mod), nativeDLL(opt), uint32(processing))
	return fromDLL(v), publicError(e)
}

// AnalysisDLLThermalNoiseJitter returns analysis thermal-noise DLL jitter.
func AnalysisDLLThermalNoiseJitter(mod SignalModulation, opt DLLTrackingOptions, processing DLLProcessing) (DLLJitter, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return DLLJitter{}, err
	}
	if err := validateDLLProcessing(processing); err != nil {
		return DLLJitter{}, err
	}
	v, e := native.SignalDLLAnalysis(nativeModulation(mod), nativeDLL(opt), uint32(processing))
	return fromDLL(v), publicError(e)
}

// DLLLowerBound returns the native lower-bound DLL jitter.
func DLLLowerBound(mod SignalModulation, opt DLLTrackingOptions) (DLLJitter, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return DLLJitter{}, err
	}
	v, e := native.SignalDLLLowerBound(nativeModulation(mod), nativeDLL(opt))
	return fromDLL(v), publicError(e)
}

// AnalysisDLLLowerBound returns the analysis lower-bound DLL jitter.
func AnalysisDLLLowerBound(mod SignalModulation, opt DLLTrackingOptions) (DLLJitter, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return DLLJitter{}, err
	}
	v, e := native.SignalAnalysisDLLLowerBound(nativeModulation(mod), nativeDLL(opt))
	return fromDLL(v), publicError(e)
}

// SignalInterference describes one interfering signal in a CN0 calculation.
type SignalInterference struct {
	// Modulation identifies the interfering waveform.
	Modulation SignalModulation
	// PowerRatioToCarrier is a dimensionless power ratio.
	PowerRatioToCarrier float64
}

// CN0Degradation contains effective linear CN0 in hertz, effective CN0 in
// dBHz, and degradation in dB.
type CN0Degradation struct {
	EffectiveCN0Hz   float64
	EffectiveCN0DBHz float64
	DegradationDB    float64
}

// EffectiveCN0Degradation computes the effective CN0 after interference.
func EffectiveCN0Degradation(desired SignalModulation, cn0DBHz, bandwidthHz float64, terms []SignalInterference) (CN0Degradation, error) {
	return effectiveCN0Degradation(desired, cn0DBHz, bandwidthHz, terms, false)
}

// AnalysisEffectiveCN0Degradation computes the analysis effective CN0.
func AnalysisEffectiveCN0Degradation(desired SignalModulation, cn0DBHz, bandwidthHz float64, terms []SignalInterference) (CN0Degradation, error) {
	return effectiveCN0Degradation(desired, cn0DBHz, bandwidthHz, terms, true)
}

func effectiveCN0Degradation(desired SignalModulation, cn0DBHz, bandwidthHz float64, terms []SignalInterference, analysis bool) (CN0Degradation, error) {
	if err := validateSignalModulationKind(desired.Kind); err != nil {
		return CN0Degradation{}, err
	}
	v := make([]native.NativeSignalInterference, len(terms))
	for i, x := range terms {
		if err := validateSignalModulationKind(x.Modulation.Kind); err != nil {
			return CN0Degradation{}, err
		}
		v[i] = native.NativeSignalInterference{Modulation: nativeModulation(x.Modulation), PowerRatioToCarrier: x.PowerRatioToCarrier}
	}
	var r native.NativeCN0Degradation
	var e error
	if analysis {
		r, e = native.SignalAnalysisCN0(nativeModulation(desired), cn0DBHz, bandwidthHz, v)
	} else {
		r, e = native.SignalCN0(nativeModulation(desired), cn0DBHz, bandwidthHz, v)
	}
	return CN0Degradation{EffectiveCN0Hz: r.EffectiveCN0Hz, EffectiveCN0DBHz: r.EffectiveCN0DBHz, DegradationDB: r.DegradationDB}, publicError(e)
}

// AnalysisFractionPowerInBand returns analysis in-band power fraction.
func AnalysisFractionPowerInBand(mod SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalAnalysisFractionPowerInBand(nativeModulation(mod), bandwidthHz)
	return v, publicError(e)
}

// MultipathOptions controls multipath-envelope evaluation. Ratio is
// dimensionless, spacing is chips, and receiver bandwidth is hertz.
type MultipathOptions struct {
	MultipathToDirectRatio float64
	CorrelatorSpacingChips float64
	ReceiverBandwidthHz    float64
}

// MultipathEnvelopePoint contains delay in chips and seconds and error
// envelopes in chips, seconds, and metres.
type MultipathEnvelopePoint struct {
	DelayChips          float64
	DelayS              float64
	InPhaseChips        float64
	InPhaseS            float64
	InPhaseM            float64
	AntiPhaseChips      float64
	AntiPhaseS          float64
	AntiPhaseM          float64
	RunningAverageChips float64
	RunningAverageS     float64
	RunningAverageM     float64
}

// MultipathErrorEnvelope evaluates the native multipath error envelope.
func MultipathErrorEnvelope(mod SignalModulation, opt MultipathOptions, delays []float64) ([]MultipathEnvelopePoint, error) {
	return multipathErrorEnvelope(mod, opt, delays, false)
}

// AnalysisMultipathErrorEnvelope evaluates the analysis multipath envelope.
func AnalysisMultipathErrorEnvelope(mod SignalModulation, opt MultipathOptions, delays []float64) ([]MultipathEnvelopePoint, error) {
	return multipathErrorEnvelope(mod, opt, delays, true)
}

func multipathErrorEnvelope(mod SignalModulation, opt MultipathOptions, delays []float64, analysis bool) ([]MultipathEnvelopePoint, error) {
	if err := validateSignalModulationKind(mod.Kind); err != nil {
		return nil, err
	}
	options := native.NativeMultipathOptions{MultipathToDirectRatio: opt.MultipathToDirectRatio, CorrelatorSpacingChips: opt.CorrelatorSpacingChips, ReceiverBandwidthHz: opt.ReceiverBandwidthHz}
	var v []native.NativeMultipathPoint
	var e error
	if analysis {
		v, e = native.SignalAnalysisMultipath(nativeModulation(mod), options, delays)
	} else {
		v, e = native.SignalMultipath(nativeModulation(mod), options, delays)
	}
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]MultipathEnvelopePoint, len(v))
	for i, x := range v {
		out[i] = MultipathEnvelopePoint{DelayChips: x.DelayChips, DelayS: x.DelayS, InPhaseChips: x.InPhaseChips, InPhaseS: x.InPhaseS, InPhaseM: x.InPhaseM, AntiPhaseChips: x.AntiPhaseChips, AntiPhaseS: x.AntiPhaseS, AntiPhaseM: x.AntiPhaseM, RunningAverageChips: x.RunningAverageChips, RunningAverageS: x.RunningAverageS, RunningAverageM: x.RunningAverageM}
	}
	return out, nil
}

// SpectralSeparation contains spectral separation in hertz and dBHz.
type SpectralSeparation struct {
	Hz   float64
	DBHz float64
}

// SpectralSeparationCoefficient returns linear and logarithmic separation.
func SpectralSeparationCoefficient(desired, interference SignalModulation, bandwidthHz float64) (SpectralSeparation, error) {
	if err := validateSignalModulationKind(desired.Kind); err != nil {
		return SpectralSeparation{}, err
	}
	if err := validateSignalModulationKind(interference.Kind); err != nil {
		return SpectralSeparation{}, err
	}
	v, e := native.SignalSpectralSeparation(nativeModulation(desired), nativeModulation(interference), bandwidthHz)
	return SpectralSeparation{Hz: v.Hz, DBHz: v.DBHz}, publicError(e)
}

// SpectralSeparationHz returns linear spectral separation in hertz.
func SpectralSeparationHz(desired, interference SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(desired.Kind); err != nil {
		return 0, err
	}
	if err := validateSignalModulationKind(interference.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalSpectralSeparationHz(nativeModulation(desired), nativeModulation(interference), bandwidthHz)
	return v, publicError(e)
}

// SpectralSeparationDBHz returns logarithmic spectral separation in dBHz.
func SpectralSeparationDBHz(desired, interference SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(desired.Kind); err != nil {
		return 0, err
	}
	if err := validateSignalModulationKind(interference.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalSpectralSeparationDBHz(nativeModulation(desired), nativeModulation(interference), bandwidthHz)
	return v, publicError(e)
}

// WhiteNoiseSpectralSeparationHz returns white-noise separation in hertz.
func WhiteNoiseSpectralSeparationHz(desired SignalModulation, bandwidthHz float64) (float64, error) {
	if err := validateSignalModulationKind(desired.Kind); err != nil {
		return 0, err
	}
	v, e := native.SignalWhiteNoiseSeparation(nativeModulation(desired), bandwidthHz)
	return v, publicError(e)
}
