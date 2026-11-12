//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

func cPppFloatOptionsEqual(value PppFloatOptions) (C.SidereonPppFloatOptions, error) {
	count, err := checkedNativeSize(value.MaxIterations)
	return C.SidereonPppFloatOptions{max_iterations: count, position_tolerance_m: C.double(value.PositionToleranceM), clock_tolerance_m: C.double(value.ClockToleranceM), ambiguity_tolerance_m: C.double(value.AmbiguityToleranceM), ztd_tolerance_m: C.double(value.ZTDToleranceM)}, err
}

func cPppWeightsEqual(value PppMeasurementWeights) C.SidereonPppMeasurementWeights {
	return C.SidereonPppMeasurementWeights{code: C.double(value.Code), phase: C.double(value.Phase), elevation_weighting: C.bool(value.ElevationWeighting)}
}

func cPppFixedAmbiguityEqual(value PppFixedAmbiguityOptions, arena *pppArena) (C.SidereonPppFixedAmbiguityOptions, error) {
	wavelengths, wavelengthCount, err := cPppMapEntries(value.WavelengthsM, arena, "PPP equality wavelengths")
	if err != nil {
		return C.SidereonPppFixedAmbiguityOptions{}, err
	}
	offsets, offsetCount, err := cPppMapEntries(value.OffsetsM, arena, "PPP equality offsets")
	return C.SidereonPppFixedAmbiguityOptions{wavelengths_m: wavelengths, wavelength_count: wavelengthCount, offsets_m: offsets, offset_count: offsetCount, ratio_threshold: C.double(value.RatioThreshold)}, err
}

func pppEqual(call func(*C.bool) C.enum_SidereonStatus) (bool, error) {
	var output C.bool
	err := callStatus(func() uint32 { return uint32(call(&output)) })
	return bool(output), err
}

func PppAutoInitOptionsEqual(left, right PppAutoInitOptions) (bool, error) {
	a, b := cPppAutoInit(left), cPppAutoInit(right)
	return pppEqual(func(out *C.bool) C.enum_SidereonStatus { return C.sidereon_ppp_auto_init_options_equal(&a, &b, out) })
}
func PppFloatOptionsEqual(left, right PppFloatOptions) (bool, error) {
	a, err := cPppFloatOptionsEqual(left)
	if err != nil {
		return false, err
	}
	b, err := cPppFloatOptionsEqual(right)
	if err != nil {
		return false, err
	}
	return pppEqual(func(out *C.bool) C.enum_SidereonStatus { return C.sidereon_ppp_float_options_equal(&a, &b, out) })
}
func PppMeasurementWeightsEqual(left, right PppMeasurementWeights) (bool, error) {
	a := cPppWeightsEqual(left)
	b := cPppWeightsEqual(right)
	return pppEqual(func(out *C.bool) C.enum_SidereonStatus { return C.sidereon_ppp_measurement_weights_equal(&a, &b, out) })
}
func PppVelocityRobustConfigEqual(left, right PppVelocityRobustConfig) (bool, error) {
	a, err := cPppVelocityRobustConfig(left)
	if err != nil {
		return false, err
	}
	b, err := cPppVelocityRobustConfig(right)
	if err != nil {
		return false, err
	}
	return pppEqual(func(out *C.bool) C.enum_SidereonStatus {
		return C.sidereon_ppp_velocity_robust_config_equal(&a, &b, out)
	})
}
func PppVelocityConfigEqual(left, right PppVelocityConfig) (bool, error) {
	a, err := cPppVelocityConfig(left)
	if err != nil {
		return false, err
	}
	b, err := cPppVelocityConfig(right)
	if err != nil {
		return false, err
	}
	return pppEqual(func(out *C.bool) C.enum_SidereonStatus { return C.sidereon_ppp_velocity_config_equal(&a, &b, out) })
}
func PppReceiverVelocityStateEqual(left, right PppReceiverVelocityState) (bool, error) {
	a, b := cPppReceiverVelocityState(left), cPppReceiverVelocityState(right)
	return pppEqual(func(out *C.bool) C.enum_SidereonStatus {
		return C.sidereon_ppp_receiver_velocity_state_equal(&a, &b, out)
	})
}
func PppRangeRatePredictionEqual(left, right PppRangeRatePrediction) (bool, error) {
	var a, b C.SidereonPppRangeRatePrediction
	for i := 0; i < 3; i++ {
		a.los_unit[i], b.los_unit[i] = C.double(left.LOSUnit[i]), C.double(right.LOSUnit[i])
	}
	a.range_rate_m_s, b.range_rate_m_s = C.double(left.RangeRateMS), C.double(right.RangeRateMS)
	return pppEqual(func(out *C.bool) C.enum_SidereonStatus {
		return C.sidereon_ppp_range_rate_prediction_equal(&a, &b, out)
	})
}

func PppVelocityObservationEqual(left, right PppVelocityObservation) (bool, error) {
	return withCThreadValue(func() (bool, error) {
		arena := &pppArena{}
		defer arena.close()
		a, err := cPppVelocityObservation(left, arena)
		if err != nil {
			return false, err
		}
		b, err := cPppVelocityObservation(right, arena)
		if err != nil {
			return false, err
		}
		var output C.bool
		err = statusErrorLocked(uint32(C.sidereon_ppp_velocity_observation_equal(&a, &b, &output)))
		return bool(output), err
	})
}

func withCThreadValue(fn func() (bool, error)) (value bool, err error) {
	withCThread(func() { value, err = fn() })
	return value, err
}

func pppArenaEqual(left, right func(*pppArena) (any, error), call func(any, any, *C.bool) C.enum_SidereonStatus) (bool, error) {
	var equal bool
	var resultErr error
	withCThread(func() {
		arena := &pppArena{}
		defer arena.close()
		a, err := left(arena)
		if err != nil {
			resultErr = err
			return
		}
		b, err := right(arena)
		if err != nil {
			resultErr = err
			return
		}
		var output C.bool
		resultErr = statusErrorLocked(uint32(call(a, b, &output)))
		equal = bool(output)
	})
	return equal, resultErr
}

func PppObservationEqual(left, right PppObservation) (bool, error) {
	return pppArenaEqual(func(a *pppArena) (any, error) {
		values, _, err := cPppEpochs([]PppEpoch{{Observations: []PppObservation{left}}}, a)
		if err != nil {
			return nil, err
		}
		return values.observations, nil
	}, func(a *pppArena) (any, error) {
		values, _, err := cPppEpochs([]PppEpoch{{Observations: []PppObservation{right}}}, a)
		if err != nil {
			return nil, err
		}
		return values.observations, nil
	}, func(a, b any, out *C.bool) C.enum_SidereonStatus {
		return C.sidereon_ppp_observation_equal(a.(*C.SidereonPppObservation), b.(*C.SidereonPppObservation), out)
	})
}
func PppEpochEqual(left, right PppEpoch) (bool, error) {
	return pppArenaEqual(func(a *pppArena) (any, error) { value, _, err := cPppEpochs([]PppEpoch{left}, a); return value, err }, func(a *pppArena) (any, error) { value, _, err := cPppEpochs([]PppEpoch{right}, a); return value, err }, func(a, b any, out *C.bool) C.enum_SidereonStatus {
		return C.sidereon_ppp_epoch_equal(a.(*C.SidereonPppEpoch), b.(*C.SidereonPppEpoch), out)
	})
}
func PppTroposphereOptionsEqual(left, right PppTroposphereOptions) (bool, error) {
	return pppArenaEqual(func(a *pppArena) (any, error) { value, err := cPppTroposphere(left); return value, err }, func(a *pppArena) (any, error) { value, err := cPppTroposphere(right); return value, err }, func(a, b any, out *C.bool) C.enum_SidereonStatus {
		av, bv := a.(C.SidereonPppTroposphereOptions), b.(C.SidereonPppTroposphereOptions)
		return C.sidereon_ppp_troposphere_options_equal(&av, &bv, out)
	})
}
func PppRangeCorrectionsEqual(left, right PppRangeCorrections) (bool, error) {
	return pppArenaEqual(func(a *pppArena) (any, error) { value, err := cPppRangeCorrections(left, a); return value, err }, func(a *pppArena) (any, error) { value, err := cPppRangeCorrections(right, a); return value, err }, func(a, b any, out *C.bool) C.enum_SidereonStatus {
		av, bv := a.(C.SidereonPppRangeCorrections), b.(C.SidereonPppRangeCorrections)
		return C.sidereon_ppp_range_corrections_equal(&av, &bv, out)
	})
}
func PppFixedAmbiguityOptionsEqual(left, right PppFixedAmbiguityOptions) (bool, error) {
	return pppArenaEqual(func(a *pppArena) (any, error) { value, err := cPppFixedAmbiguityEqual(left, a); return value, err }, func(a *pppArena) (any, error) { value, err := cPppFixedAmbiguityEqual(right, a); return value, err }, func(a, b any, out *C.bool) C.enum_SidereonStatus {
		av, bv := a.(C.SidereonPppFixedAmbiguityOptions), b.(C.SidereonPppFixedAmbiguityOptions)
		return C.sidereon_ppp_fixed_ambiguity_options_equal(&av, &bv, out)
	})
}
func PppFloatConfigEqual(left, right PppFloatConfig) (bool, error) {
	return pppArenaEqual(func(a *pppArena) (any, error) {
		legacy, err := cPppFloatConfig(left, a, true)
		return C.SidereonPppFloatConfigV2{legacy: legacy, estimate_residual_ionosphere: C.bool(left.EstimateResidualIonosphere)}, err
	}, func(a *pppArena) (any, error) {
		legacy, err := cPppFloatConfig(right, a, true)
		return C.SidereonPppFloatConfigV2{legacy: legacy, estimate_residual_ionosphere: C.bool(right.EstimateResidualIonosphere)}, err
	}, func(a, b any, out *C.bool) C.enum_SidereonStatus {
		av, bv := a.(C.SidereonPppFloatConfigV2), b.(C.SidereonPppFloatConfigV2)
		return C.sidereon_ppp_float_config_v2_equal(&av, &bv, out)
	})
}
func PppFixedConfigEqual(left, right PppFixedConfig) (bool, error) {
	return pppArenaEqual(func(a *pppArena) (any, error) {
		legacy, err := cPppFixedConfig(left, a)
		return C.SidereonPppFixedConfigV2{legacy: legacy, estimate_residual_ionosphere: C.bool(left.EstimateResidualIonosphere)}, err
	}, func(a *pppArena) (any, error) {
		legacy, err := cPppFixedConfig(right, a)
		return C.SidereonPppFixedConfigV2{legacy: legacy, estimate_residual_ionosphere: C.bool(right.EstimateResidualIonosphere)}, err
	}, func(a, b any, out *C.bool) C.enum_SidereonStatus {
		av, bv := a.(C.SidereonPppFixedConfigV2), b.(C.SidereonPppFixedConfigV2)
		return C.sidereon_ppp_fixed_config_v2_equal(&av, &bv, out)
	})
}
