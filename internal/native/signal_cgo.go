//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

type NativeIQSample struct{ I, Q float64 }
type NativeAcquisitionOptions struct{ SampleRateHz, DopplerMinHz, DopplerMaxHz, DopplerStepHz float64 }
type NativeAcquisitionResult struct {
	CodePhaseChips, DopplerHz, PeakMetric, Metric, PeakPower float64
	GridCodePhaseBins                                        int
	GridDopplerStepHz, GridSamplesPerChip                    float64
	GridDopplerBinCount                                      int
}
type NativeSignalModulation struct {
	Kind        uint32
	Order, M, N float64
}
type NativeDLLTrackingOptions struct{ CN0DBHz, LoopBandwidthHz, IntegrationTimeS, CorrelatorSpacingChips, ReceiverBandwidthHz float64 }
type NativeDLLJitter struct{ Seconds, Chips, Meters, SquaringLoss float64 }
type NativeSignalInterference struct {
	Modulation          NativeSignalModulation
	PowerRatioToCarrier float64
}
type NativeCN0Degradation struct{ EffectiveCN0Hz, EffectiveCN0DBHz, DegradationDB float64 }
type NativeMultipathOptions struct{ MultipathToDirectRatio, CorrelatorSpacingChips, ReceiverBandwidthHz float64 }
type NativeMultipathPoint struct{ DelayChips, DelayS, InPhaseChips, InPhaseS, InPhaseM, AntiPhaseChips, AntiPhaseS, AntiPhaseM, RunningAverageChips, RunningAverageS, RunningAverageM float64 }
type NativeSpectralSeparation struct{ Hz, DBHz float64 }
type NativeCorrelateOptions struct{ SampleRateHz, DopplerHz, CodePhaseChips, CodeDopplerHz float64 }
type NativeCorrelationResult struct{ I, Q, Power float64 }
type NativeReplicaOptions struct {
	SampleRateHz                  float64
	NumSamples                    int
	CodePhaseChips, CodeDopplerHz float64
}

func RINEXBandFrequency(system uint32, band string, hasChannel bool, channel int8) (float64, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return 0, err
	}
	if err := validateRINEXBand(band); err != nil {
		return 0, err
	}
	var out C.double
	err := withString(band, func(value *C.char) uint32 {
		return C.sidereon_rinex_band_frequency_hz(C.uint32_t(system), value, C.bool(hasChannel), C.int8_t(channel), &out)
	})
	return float64(out), err
}

func RINEXBandWavelength(system uint32, band string, hasChannel bool, channel int8) (float64, error) {
	if err := validateGNSSSystemValue(system); err != nil {
		return 0, err
	}
	if err := validateRINEXBand(band); err != nil {
		return 0, err
	}
	var out C.double
	err := withString(band, func(value *C.char) uint32 {
		return C.sidereon_rinex_band_wavelength_m(C.uint32_t(system), value, C.bool(hasChannel), C.int8_t(channel), &out)
	})
	return float64(out), err
}

func validateRINEXBand(value string) error {
	if err := rejectEmbeddedNUL(value, "RINEX band"); err != nil {
		return err
	}
	if len(value) != 1 {
		return errors.New("sidereon: RINEX band must contain exactly one byte")
	}
	return nil
}

func cModulation(v NativeSignalModulation) C.SidereonSignalAnalysisModulation {
	return C.SidereonSignalAnalysisModulation{kind: C.uint32_t(v.Kind), order: C.double(v.Order), m: C.double(v.M), n: C.double(v.N)}
}
func cDLL(v NativeDLLTrackingOptions) C.SidereonSignalAnalysisDllTrackingOptions {
	return C.SidereonSignalAnalysisDllTrackingOptions{cn0_db_hz: C.double(v.CN0DBHz), loop_bandwidth_hz: C.double(v.LoopBandwidthHz), integration_time_s: C.double(v.IntegrationTimeS), correlator_spacing_chips: C.double(v.CorrelatorSpacingChips), receiver_bandwidth_hz: C.double(v.ReceiverBandwidthHz)}
}
func cMultipath(v NativeMultipathOptions) C.SidereonSignalAnalysisMultipathOptions {
	return C.SidereonSignalAnalysisMultipathOptions{multipath_to_direct_ratio: C.double(v.MultipathToDirectRatio), correlator_spacing_chips: C.double(v.CorrelatorSpacingChips), receiver_bandwidth_hz: C.double(v.ReceiverBandwidthHz)}
}

func withCSignalModulation(value NativeSignalModulation, fn func(*C.SidereonSignalAnalysisModulation) uint32) error {
	var err error
	withCThread(func() {
		memory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisModulation{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(memory)
		modulation := (*C.SidereonSignalAnalysisModulation)(memory)
		*modulation = cModulation(value)
		err = statusErrorLocked(fn(modulation))
	})
	return err
}

func withCSignalModulations(first, second NativeSignalModulation, fn func(*C.SidereonSignalAnalysisModulation, *C.SidereonSignalAnalysisModulation) uint32) error {
	var err error
	withCThread(func() {
		firstMemory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisModulation{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(firstMemory)
		secondMemory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisModulation{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(secondMemory)
		firstValue := (*C.SidereonSignalAnalysisModulation)(firstMemory)
		secondValue := (*C.SidereonSignalAnalysisModulation)(secondMemory)
		*firstValue = cModulation(first)
		*secondValue = cModulation(second)
		err = statusErrorLocked(fn(firstValue, secondValue))
	})
	return err
}

func signalDLL(mod NativeSignalModulation, opt NativeDLLTrackingOptions, processing uint32, fn func(*C.SidereonSignalAnalysisModulation, *C.SidereonSignalAnalysisDllTrackingOptions, *C.SidereonSignalAnalysisDllJitter) uint32) (NativeDLLJitter, error) {
	var value C.SidereonSignalAnalysisDllJitter
	err := withCThreadError(func() error {
		modulationMemory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisModulation{}))
		if err != nil {
			return err
		}
		defer C.free(modulationMemory)
		optionsMemory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisDllTrackingOptions{}))
		if err != nil {
			return err
		}
		defer C.free(optionsMemory)
		modulation := (*C.SidereonSignalAnalysisModulation)(modulationMemory)
		options := (*C.SidereonSignalAnalysisDllTrackingOptions)(optionsMemory)
		*modulation = cModulation(mod)
		*options = cDLL(opt)
		return statusErrorLocked(fn(modulation, options, &value))
	})
	return NativeDLLJitter{Seconds: float64(value.seconds), Chips: float64(value.chips), Meters: float64(value.meters), SquaringLoss: float64(value.squaring_loss)}, err
}

func withCThreadError(fn func() error) error {
	var err error
	withCThread(func() { err = fn() })
	return err
}

func nativeInt8Output(label string, call func(*C.int8_t, C.size_t, *C.size_t, *C.size_t) uint32) ([]int8, error) {
	var w, r C.size_t
	if err := callStatus(func() uint32 { return call(nil, 0, &w, &r) }); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery(label, uint64(w), uint64(r))
	if err != nil {
		return nil, err
	}
	nativeLength, err := checkedNativeSize(n)
	if err != nil {
		return nil, err
	}
	if _, err = checkedNativeAllocationSize(n, unsafe.Sizeof(C.int8_t(0))); err != nil {
		return nil, err
	}
	mem, err := checkedNativeMalloc(n, unsafe.Sizeof(C.int8_t(0)))
	if err != nil {
		return nil, err
	}
	if n > 0 && mem == nil {
		return nil, errors.New("sidereon: unable to allocate native output")
	}
	if mem != nil {
		defer C.free(mem)
	}
	var out *C.int8_t
	if n > 0 {
		out = (*C.int8_t)(mem)
	}
	if err = callStatus(func() uint32 { return call(out, nativeLength, &w, &r) }); err != nil {
		return nil, err
	}
	z, err := validateNativeOutput(label, n, uint64(w), uint64(r))
	if err != nil {
		return nil, err
	}
	values := unsafe.Slice((*C.int8_t)(mem), z)
	outValues := make([]int8, z)
	for i := range outValues {
		outValues[i] = int8(values[i])
	}
	return outValues, nil
}

func SignalCAChip(prn, index int64) (int8, error) {
	var out C.int8_t
	err := callStatus(func() uint32 { return C.sidereon_signal_ca_chip(C.int64_t(prn), C.int64_t(index), &out) })
	return int8(out), err
}
func SignalCACode(prn int64) ([]int8, error) {
	return nativeInt8Output("C/A code", func(o *C.int8_t, n C.size_t, w, r *C.size_t) uint32 {
		return C.sidereon_signal_ca_code(C.int64_t(prn), o, n, w, r)
	})
}
func SignalReplica(prn int64, options NativeReplicaOptions) ([]int8, error) {
	if options.NumSamples < 0 {
		return nil, errNegativeIndex
	}
	numSamples, err := checkedNativeSize(options.NumSamples)
	if err != nil {
		return nil, err
	}
	optionMemory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonReplicaOptions{}))
	if err != nil {
		return nil, err
	}
	defer C.free(optionMemory)
	v := (*C.SidereonReplicaOptions)(optionMemory)
	v.sample_rate_hz = C.double(options.SampleRateHz)
	v.num_samples = numSamples
	v.code_phase_chips = C.double(options.CodePhaseChips)
	v.code_doppler_hz = C.double(options.CodeDopplerHz)
	return nativeInt8Output("signal replica", func(o *C.int8_t, n C.size_t, w, r *C.size_t) uint32 {
		return C.sidereon_signal_replica(C.int64_t(prn), v, o, n, w, r)
	})
}

func SignalCorrelationAt(a, b []int8, lag int64) (int32, error) {
	if len(a) != len(b) {
		return 0, errors.New("sidereon: code lengths differ")
	}
	var out C.int32_t
	err := withCInt8Pair(a, b, func(x, y *C.int8_t, n C.size_t) error {
		return callStatus(func() uint32 { return C.sidereon_signal_correlation_at(x, y, n, C.int64_t(lag), &out) })
	})
	return int32(out), err
}
func SignalCrossCorrelation(a, b []int8) ([]int32, error) {
	if len(a) != len(b) {
		return nil, errors.New("sidereon: code lengths differ")
	}
	var out []int32
	err := withCInt8Pair(a, b, func(x, y *C.int8_t, n C.size_t) error {
		var w, r C.size_t
		if e := callStatus(func() uint32 { return C.sidereon_signal_cross_correlation(x, y, n, nil, 0, &w, &r) }); e != nil {
			return e
		}
		need, e := validateNativeQuery("signal cross correlation", uint64(w), uint64(r))
		if e != nil {
			return e
		}
		mem, e := checkedNativeMalloc(need, unsafe.Sizeof(C.int32_t(0)))
		if e != nil {
			return e
		}
		defer C.free(mem)
		if e := callStatus(func() uint32 {
			return C.sidereon_signal_cross_correlation(x, y, n, (*C.int32_t)(mem), C.size_t(need), &w, &r)
		}); e != nil {
			return e
		}
		z, e := validateTwoPassCounts("signal cross correlation", need, need, uint64(w), uint64(r))
		if e != nil {
			return e
		}
		out = make([]int32, z)
		vals := unsafe.Slice((*C.int32_t)(mem), z)
		for i := range out {
			out[i] = int32(vals[i])
		}
		return nil
	})
	return out, err
}

func withCInt8Pair(a, b []int8, fn func(*C.int8_t, *C.int8_t, C.size_t) error) error {
	length, lengthErr := checkedNativeSize(len(a))
	if lengthErr != nil {
		return lengthErr
	}
	var err error
	withCThread(func() {
		var pa, pb unsafe.Pointer
		if len(a) > 0 {
			var allocationErr error
			pa, allocationErr = checkedNativeMalloc(len(a), unsafe.Sizeof(C.int8_t(0)))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(pa)
			values := unsafe.Slice((*C.int8_t)(pa), len(a))
			for i, value := range a {
				values[i] = C.int8_t(value)
			}
		}
		if len(b) > 0 {
			var allocationErr error
			pb, allocationErr = checkedNativeMalloc(len(b), unsafe.Sizeof(C.int8_t(0)))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(pb)
			values := unsafe.Slice((*C.int8_t)(pb), len(b))
			for i, value := range b {
				values[i] = C.int8_t(value)
			}
		}
		err = fn((*C.int8_t)(pa), (*C.int8_t)(pb), length)
	})
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	return err
}

func SignalAutocorrelation(code []int8) ([]int32, error) {
	codeLength, lengthErr := checkedNativeSize(len(code))
	if lengthErr != nil {
		return nil, lengthErr
	}
	var out []int32
	var err error
	withCThread(func() {
		var p unsafe.Pointer
		if len(code) > 0 {
			var allocationErr error
			p, allocationErr = checkedNativeMalloc(len(code), unsafe.Sizeof(C.int8_t(0)))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(p)
			values := unsafe.Slice((*C.int8_t)(p), len(code))
			for i, value := range code {
				values[i] = C.int8_t(value)
			}
		}
		var w, r C.size_t
		if err = statusErrorLocked(C.sidereon_signal_autocorrelation((*C.int8_t)(p), codeLength, nil, 0, &w, &r)); err != nil {
			return
		}
		n, e := validateNativeQuery("signal autocorrelation", uint64(w), uint64(r))
		if e != nil {
			err = e
			return
		}
		mem, e := checkedNativeMalloc(n, unsafe.Sizeof(C.int32_t(0)))
		if e != nil {
			err = e
			return
		}
		if mem != nil {
			defer C.free(mem)
		}
		outputLength, e := checkedNativeSize(n)
		if e != nil {
			err = e
			return
		}
		if err = statusErrorLocked(C.sidereon_signal_autocorrelation((*C.int8_t)(p), codeLength, (*C.int32_t)(mem), outputLength, &w, &r)); err != nil {
			return
		}
		z, e := validateNativeOutput("signal autocorrelation", n, uint64(w), uint64(r))
		if e != nil {
			err = e
			return
		}
		out = make([]int32, z)
		v := unsafe.Slice((*C.int32_t)(mem), z)
		for i := range out {
			out[i] = int32(v[i])
		}
	})
	runtime.KeepAlive(code)
	return out, err
}

func SignalCorrelate(iq []NativeIQSample, prn int64, opt NativeCorrelateOptions) (NativeCorrelationResult, error) {
	var out C.SidereonCorrelationResult
	optionMemory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonCorrelateOptions{}))
	if err != nil {
		return NativeCorrelationResult{}, err
	}
	defer C.free(optionMemory)
	options := (*C.SidereonCorrelateOptions)(optionMemory)
	*options = C.SidereonCorrelateOptions{sample_rate_hz: C.double(opt.SampleRateHz), doppler_hz: C.double(opt.DopplerHz), code_phase_chips: C.double(opt.CodePhaseChips), code_doppler_hz: C.double(opt.CodeDopplerHz)}
	err = withCIQ(iq, func(p *C.SidereonIqSample, n C.size_t) uint32 {
		return C.sidereon_signal_correlate(p, n, C.int64_t(prn), options, &out)
	})
	return NativeCorrelationResult{I: float64(out.i), Q: float64(out.q), Power: float64(out.power)}, err
}
func SignalCorrelateAgainst(iq []NativeIQSample, code []int8, fs, doppler float64) (float64, float64, error) {
	var i, q C.double
	err := withCIQAndCode(iq, code, func(p *C.SidereonIqSample, n C.size_t, c *C.int8_t, m C.size_t) uint32 {
		return C.sidereon_signal_correlate_against(p, n, c, m, C.double(fs), C.double(doppler), &i, &q)
	})
	return float64(i), float64(q), err
}
func withCIQ(iq []NativeIQSample, fn func(*C.SidereonIqSample, C.size_t) uint32) error {
	length, lengthErr := checkedNativeSize(len(iq))
	if lengthErr != nil {
		return lengthErr
	}
	var err error
	withCThread(func() {
		var p unsafe.Pointer
		if len(iq) > 0 {
			var allocationErr error
			p, allocationErr = checkedNativeMalloc(len(iq), unsafe.Sizeof(C.SidereonIqSample{}))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(p)
			v := unsafe.Slice((*C.SidereonIqSample)(p), len(iq))
			for i := range iq {
				v[i].i = C.double(iq[i].I)
				v[i].q = C.double(iq[i].Q)
			}
		}
		err = statusErrorLocked(fn((*C.SidereonIqSample)(p), length))
	})
	runtime.KeepAlive(iq)
	return err
}
func withCIQAndCode(iq []NativeIQSample, code []int8, fn func(*C.SidereonIqSample, C.size_t, *C.int8_t, C.size_t) uint32) error {
	iqLength, iqErr := checkedNativeSize(len(iq))
	if iqErr != nil {
		return iqErr
	}
	codeLength, codeErr := checkedNativeSize(len(code))
	if codeErr != nil {
		return codeErr
	}
	var err error
	withCThread(func() {
		var p unsafe.Pointer
		if len(iq) > 0 {
			var allocationErr error
			p, allocationErr = checkedNativeMalloc(len(iq), unsafe.Sizeof(C.SidereonIqSample{}))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(p)
			v := unsafe.Slice((*C.SidereonIqSample)(p), len(iq))
			for i := range iq {
				v[i].i = C.double(iq[i].I)
				v[i].q = C.double(iq[i].Q)
			}
		}
		var c unsafe.Pointer
		if len(code) > 0 {
			var allocationErr error
			c, allocationErr = checkedNativeMalloc(len(code), unsafe.Sizeof(C.int8_t(0)))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(c)
			values := unsafe.Slice((*C.int8_t)(c), len(code))
			for i, value := range code {
				values[i] = C.int8_t(value)
			}
		}
		err = statusErrorLocked(fn((*C.SidereonIqSample)(p), iqLength, (*C.int8_t)(c), codeLength))
	})
	runtime.KeepAlive(iq)
	runtime.KeepAlive(code)
	return err
}

func SignalAcquire(samples []NativeIQSample, prn int64, opt NativeAcquisitionOptions) (NativeAcquisitionResult, []float64, error) {
	sampleLength, lengthErr := checkedNativeSize(len(samples))
	if lengthErr != nil {
		return NativeAcquisitionResult{}, nil, lengthErr
	}
	var result C.SidereonAcquisitionResult
	var bins []float64
	var err error
	withCThread(func() {
		var p unsafe.Pointer
		if len(samples) > 0 {
			var allocationErr error
			p, allocationErr = checkedNativeMalloc(len(samples), unsafe.Sizeof(C.SidereonIqSample{}))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(p)
			v := unsafe.Slice((*C.SidereonIqSample)(p), len(samples))
			for i := range samples {
				v[i].i = C.double(samples[i].I)
				v[i].q = C.double(samples[i].Q)
			}
		}
		optionMemory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonAcquisitionOptions{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(optionMemory)
		options := (*C.SidereonAcquisitionOptions)(optionMemory)
		*options = C.SidereonAcquisitionOptions{sample_rate_hz: C.double(opt.SampleRateHz), doppler_min_hz: C.double(opt.DopplerMinHz), doppler_max_hz: C.double(opt.DopplerMaxHz), doppler_step_hz: C.double(opt.DopplerStepHz)}
		var w, r C.size_t
		if err = statusErrorLocked(C.sidereon_signal_acquire((*C.SidereonIqSample)(p), sampleLength, C.int64_t(prn), options, &result, nil, 0, &w, &r)); err != nil {
			return
		}
		n, e := validateNativeQuery("acquisition Doppler bins", uint64(w), uint64(r))
		if e != nil {
			err = e
			return
		}
		mem, e := checkedNativeMalloc(n, unsafe.Sizeof(C.double(0)))
		if e != nil {
			err = e
			return
		}
		if mem != nil {
			defer C.free(mem)
		}
		outputLength, e := checkedNativeSize(n)
		if e != nil {
			err = e
			return
		}
		if err = statusErrorLocked(C.sidereon_signal_acquire((*C.SidereonIqSample)(p), sampleLength, C.int64_t(prn), options, &result, (*C.double)(mem), outputLength, &w, &r)); err != nil {
			return
		}
		z, e := validateNativeOutput("acquisition Doppler bins", n, uint64(w), uint64(r))
		if e != nil {
			err = e
			return
		}
		bins = make([]float64, z)
		v := unsafe.Slice((*C.double)(mem), z)
		for i := range bins {
			bins[i] = float64(v[i])
		}
	})
	runtime.KeepAlive(samples)
	out := NativeAcquisitionResult{CodePhaseChips: float64(result.code_phase_chips), DopplerHz: float64(result.doppler_hz), PeakMetric: float64(result.peak_metric), Metric: float64(result.metric), PeakPower: float64(result.peak_power), GridDopplerStepHz: float64(result.grid_doppler_step_hz), GridSamplesPerChip: float64(result.grid_samples_per_chip)}
	if err == nil {
		out.GridCodePhaseBins, err = checkedNativeCount(uint64(result.grid_code_phase_bins))
	}
	if err == nil {
		out.GridDopplerBinCount, err = checkedNativeCount(uint64(result.grid_doppler_bin_count))
	}
	return out, bins, err
}

func signalScalar(mod NativeSignalModulation, fn func(*C.SidereonSignalAnalysisModulation, *C.double) uint32) (float64, error) {
	var out C.double
	err := withCSignalModulation(mod, func(value *C.SidereonSignalAnalysisModulation) uint32 { return fn(value, &out) })
	return float64(out), err
}
func SignalPSD(mod NativeSignalModulation, offset float64) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_psd_hz(m, C.double(offset), o)
	})
}
func SignalAnalysisPSD(mod NativeSignalModulation, offset float64) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_analysis_psd(m, C.double(offset), o)
	})
}
func SignalModulationCodeRate(mod NativeSignalModulation) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_modulation_code_rate_hz(m, o)
	})
}
func SignalPowerInBand(mod NativeSignalModulation, bw float64) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_power_in_band(m, C.double(bw), o)
	})
}
func SignalFractionPowerInBand(mod NativeSignalModulation, bw float64) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_fraction_power_in_band(m, C.double(bw), o)
	})
}
func SignalRMSBandwidth(mod NativeSignalModulation, bw float64) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_rms_bandwidth_hz(m, C.double(bw), o)
	})
}
func SignalAnalysisRMSBandwidth(mod NativeSignalModulation, bw float64) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_analysis_rms_bandwidth_hz(m, C.double(bw), o)
	})
}
func SignalReferenceChipRate() (float64, error) {
	var o C.double
	err := callStatus(func() uint32 { return C.sidereon_signal_reference_chip_rate_hz(&o) })
	return float64(o), err
}
func SignalBetzBandwidth() (float64, error) {
	var o C.double
	err := callStatus(func() uint32 { return C.sidereon_signal_betz_l1_receiver_bandwidth_hz(&o) })
	return float64(o), err
}
func SignalCoherentLoss(freq, t float64) (float64, error) {
	var o C.double
	err := callStatus(func() uint32 { return C.sidereon_signal_coherent_loss(C.double(freq), C.double(t), &o) })
	return float64(o), err
}
func SignalCoherentLossDB(freq, t float64) (float64, error) {
	var o C.double
	err := callStatus(func() uint32 { return C.sidereon_signal_coherent_loss_db(C.double(freq), C.double(t), &o) })
	return float64(o), err
}
func SignalSNRPost(cn0, t float64) (float64, error) {
	var o C.double
	err := callStatus(func() uint32 { return C.sidereon_signal_snr_post_db(C.double(cn0), C.double(t), &o) })
	return float64(o), err
}
func SignalModulationLabel(mod NativeSignalModulation) (string, error) {
	memory, err := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisModulation{}))
	if err != nil {
		return "", err
	}
	defer C.free(memory)
	m := (*C.SidereonSignalAnalysisModulation)(memory)
	*m = cModulation(mod)
	b, e := copyByteOutput("signal modulation label", func(o *C.uint8_t, n C.size_t, w, r *C.size_t) uint32 {
		return C.sidereon_signal_modulation_label(m, o, n, w, r)
	})
	return string(b), e
}
func SignalDLL(mod NativeSignalModulation, opt NativeDLLTrackingOptions, processing uint32) (NativeDLLJitter, error) {
	return signalDLL(mod, opt, processing, func(m *C.SidereonSignalAnalysisModulation, o *C.SidereonSignalAnalysisDllTrackingOptions, v *C.SidereonSignalAnalysisDllJitter) uint32 {
		return C.sidereon_signal_dll_thermal_noise_jitter(m, o, C.uint32_t(processing), v)
	})
}
func SignalDLLAnalysis(mod NativeSignalModulation, opt NativeDLLTrackingOptions, processing uint32) (NativeDLLJitter, error) {
	return signalDLL(mod, opt, processing, func(m *C.SidereonSignalAnalysisModulation, o *C.SidereonSignalAnalysisDllTrackingOptions, v *C.SidereonSignalAnalysisDllJitter) uint32 {
		return C.sidereon_signal_analysis_dll_jitter(m, o, C.uint32_t(processing), v)
	})
}
func SignalDLLLowerBound(mod NativeSignalModulation, opt NativeDLLTrackingOptions) (NativeDLLJitter, error) {
	return signalDLL(mod, opt, 0, func(m *C.SidereonSignalAnalysisModulation, o *C.SidereonSignalAnalysisDllTrackingOptions, v *C.SidereonSignalAnalysisDllJitter) uint32 {
		return C.sidereon_signal_dll_lower_bound(m, o, v)
	})
}
func SignalAnalysisDLLLowerBound(mod NativeSignalModulation, opt NativeDLLTrackingOptions) (NativeDLLJitter, error) {
	return signalDLL(mod, opt, 0, func(m *C.SidereonSignalAnalysisModulation, o *C.SidereonSignalAnalysisDllTrackingOptions, v *C.SidereonSignalAnalysisDllJitter) uint32 {
		return C.sidereon_signal_analysis_dll_lower_bound(m, o, v)
	})
}
func SignalSpectralSeparation(a, b NativeSignalModulation, bw float64) (NativeSpectralSeparation, error) {
	var v C.SidereonSignalAnalysisSpectralSeparation
	err := withCSignalModulations(a, b, func(x, y *C.SidereonSignalAnalysisModulation) uint32 {
		return C.sidereon_signal_analysis_spectral_separation(x, y, C.double(bw), &v)
	})
	return NativeSpectralSeparation{Hz: float64(v.hz), DBHz: float64(v.db_hz)}, err
}
func SignalSpectralSeparationHz(a, b NativeSignalModulation, bw float64) (float64, error) {
	var v C.double
	err := withCSignalModulations(a, b, func(x, y *C.SidereonSignalAnalysisModulation) uint32 {
		return C.sidereon_signal_spectral_separation_coefficient_hz(x, y, C.double(bw), &v)
	})
	return float64(v), err
}
func SignalSpectralSeparationDBHz(a, b NativeSignalModulation, bw float64) (float64, error) {
	var v C.double
	err := withCSignalModulations(a, b, func(x, y *C.SidereonSignalAnalysisModulation) uint32 {
		return C.sidereon_signal_spectral_separation_coefficient_db_hz(x, y, C.double(bw), &v)
	})
	return float64(v), err
}
func SignalWhiteNoiseSeparation(a NativeSignalModulation, bw float64) (float64, error) {
	var v C.double
	err := withCSignalModulation(a, func(x *C.SidereonSignalAnalysisModulation) uint32 {
		return C.sidereon_signal_white_noise_spectral_separation_hz(x, C.double(bw), &v)
	})
	return float64(v), err
}
func SignalCN0(a NativeSignalModulation, cn0, bw float64, terms []NativeSignalInterference) (NativeCN0Degradation, error) {
	return signalCN0(a, cn0, bw, terms, false)
}

func SignalAnalysisCN0(a NativeSignalModulation, cn0, bw float64, terms []NativeSignalInterference) (NativeCN0Degradation, error) {
	return signalCN0(a, cn0, bw, terms, true)
}

func signalCN0(a NativeSignalModulation, cn0, bw float64, terms []NativeSignalInterference, analysis bool) (NativeCN0Degradation, error) {
	termLength, lengthErr := checkedNativeSize(len(terms))
	if lengthErr != nil {
		return NativeCN0Degradation{}, lengthErr
	}
	var v C.SidereonSignalAnalysisCn0Degradation
	var err error
	withCThread(func() {
		modulationMemory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisModulation{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(modulationMemory)
		x := (*C.SidereonSignalAnalysisModulation)(modulationMemory)
		*x = cModulation(a)
		var p unsafe.Pointer
		if len(terms) > 0 {
			p, allocationErr = checkedNativeMalloc(len(terms), unsafe.Sizeof(C.SidereonSignalAnalysisInterference{}))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(p)
			out := unsafe.Slice((*C.SidereonSignalAnalysisInterference)(p), len(terms))
			for i, t := range terms {
				out[i].modulation = cModulation(t.Modulation)
				out[i].power_ratio_to_carrier = C.double(t.PowerRatioToCarrier)
			}
		}
		if analysis {
			err = statusErrorLocked(C.sidereon_signal_analysis_effective_cn0_degradation(x, C.double(cn0), C.double(bw), (*C.SidereonSignalAnalysisInterference)(p), termLength, &v))
		} else {
			err = statusErrorLocked(C.sidereon_signal_effective_cn0_degradation(x, C.double(cn0), C.double(bw), (*C.SidereonSignalAnalysisInterference)(p), termLength, &v))
		}
	})
	return NativeCN0Degradation{EffectiveCN0Hz: float64(v.effective_cn0_hz), EffectiveCN0DBHz: float64(v.effective_cn0_db_hz), DegradationDB: float64(v.degradation_db)}, err
}

func SignalAnalysisFractionPowerInBand(mod NativeSignalModulation, bw float64) (float64, error) {
	return signalScalar(mod, func(m *C.SidereonSignalAnalysisModulation, o *C.double) uint32 {
		return C.sidereon_signal_analysis_fraction_power(m, C.double(bw), o)
	})
}

func SignalAnalysisMultipath(mod NativeSignalModulation, opt NativeMultipathOptions, delays []float64) ([]NativeMultipathPoint, error) {
	return signalMultipath(mod, opt, delays, true)
}

func SignalMultipath(mod NativeSignalModulation, opt NativeMultipathOptions, delays []float64) ([]NativeMultipathPoint, error) {
	return signalMultipath(mod, opt, delays, false)
}

func signalMultipath(mod NativeSignalModulation, opt NativeMultipathOptions, delays []float64, analysis bool) ([]NativeMultipathPoint, error) {
	delayLength, lengthErr := checkedNativeSize(len(delays))
	if lengthErr != nil {
		return nil, lengthErr
	}
	var result []NativeMultipathPoint
	var err error
	withCThread(func() {
		modulationMemory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisModulation{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(modulationMemory)
		optionsMemory, allocationErr := checkedNativeMalloc(1, unsafe.Sizeof(C.SidereonSignalAnalysisMultipathOptions{}))
		if allocationErr != nil {
			err = allocationErr
			return
		}
		defer C.free(optionsMemory)
		m := (*C.SidereonSignalAnalysisModulation)(modulationMemory)
		o := (*C.SidereonSignalAnalysisMultipathOptions)(optionsMemory)
		*m = cModulation(mod)
		*o = cMultipath(opt)
		var dp unsafe.Pointer
		if len(delays) > 0 {
			dp, allocationErr = checkedNativeMalloc(len(delays), unsafe.Sizeof(C.double(0)))
			if allocationErr != nil {
				err = allocationErr
				return
			}
			defer C.free(dp)
			v := unsafe.Slice((*C.double)(dp), len(delays))
			for i, x := range delays {
				v[i] = C.double(x)
			}
		}
		var w, r C.size_t
		call := func(out *C.SidereonSignalAnalysisMultipathEnvelopePoint, length C.size_t) uint32 {
			if analysis {
				return C.sidereon_signal_analysis_multipath_envelope(m, o, (*C.double)(dp), delayLength, out, length, &w, &r)
			}
			return C.sidereon_signal_multipath_error_envelope(m, o, (*C.double)(dp), delayLength, out, length, &w, &r)
		}
		if err = statusErrorLocked(call(nil, 0)); err != nil {
			return
		}
		n, e := validateNativeQuery("signal multipath envelope", uint64(w), uint64(r))
		if e != nil {
			err = e
			return
		}
		mem, e := checkedNativeMalloc(n, unsafe.Sizeof(C.SidereonSignalAnalysisMultipathEnvelopePoint{}))
		if e != nil {
			err = e
			return
		}
		if mem != nil {
			defer C.free(mem)
		}
		if err = statusErrorLocked(call((*C.SidereonSignalAnalysisMultipathEnvelopePoint)(mem), C.size_t(n))); err != nil {
			return
		}
		z, e := validateNativeOutput("signal multipath envelope", n, uint64(w), uint64(r))
		if e != nil {
			err = e
			return
		}
		v := unsafe.Slice((*C.SidereonSignalAnalysisMultipathEnvelopePoint)(mem), z)
		result = make([]NativeMultipathPoint, z)
		for i := range result {
			x := v[i]
			result[i] = NativeMultipathPoint{DelayChips: float64(x.delay_chips), DelayS: float64(x.delay_s), InPhaseChips: float64(x.in_phase_chips), InPhaseS: float64(x.in_phase_s), InPhaseM: float64(x.in_phase_m), AntiPhaseChips: float64(x.anti_phase_chips), AntiPhaseS: float64(x.anti_phase_s), AntiPhaseM: float64(x.anti_phase_m), RunningAverageChips: float64(x.running_average_chips), RunningAverageS: float64(x.running_average_s), RunningAverageM: float64(x.running_average_m)}
		}
	})
	runtime.KeepAlive(delays)
	return result, err
}
