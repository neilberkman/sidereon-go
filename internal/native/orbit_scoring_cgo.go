//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"slices"
	"unsafe"
)

func BroadcastOrbitNew(ephemeris *BroadcastEphemeris, predictionOrigin NativeClockEpoch) (*OrbitSource, error) {
	if ephemeris == nil || ephemeris.resource == nil {
		return nil, ErrClosed
	}
	var value *C.SidereonOrbitSource
	err := ephemeris.resource.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_broadcast_orbit_new((*C.SidereonBroadcastEphemeris)(pointer), clockEpochToC(predictionOrigin), &value))
		})
	})
	runtime.KeepAlive(ephemeris)
	return newOrbitOrbitSource(value, err)
}

func DefaultVelocityHalfStepS() float64 {
	value := C.sidereon_default_velocity_half_step_s()
	return float64(value)
}

func InertialVelocityFromEarthFixed(positionECEFM [3]float64, velocityECEFMS [3]float64) ([3]float64, error) {
	var value C.SidereonVector3
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_inertial_velocity_from_earth_fixed(orbitVectorToC(positionECEFM), orbitVectorToC(velocityECEFMS), &value))
	})
	return orbitVectorFromC(value), err
}

func MinVelocityHalfStepS() float64 {
	value := C.sidereon_min_velocity_half_step_s()
	return float64(value)
}

func OrbitRegimeFromInertialState(positionM [3]float64, inertialVelocityMS [3]float64) (*OrbitRegime, error) {
	var present C.bool
	var value C.enum_SidereonOrbitRegime
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_orbit_regime_from_inertial_state(orbitVectorToC(positionM), orbitVectorToC(inertialVelocityMS), &present, &value))
	})
	if err != nil || !bool(present) {
		return nil, err
	}
	result := OrbitRegime(value)
	return &result, nil
}

func OrbitRegimeText(value OrbitRegime) (string, error) {
	return orbitText("orbit regime text", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_orbit_regime_text(C.uint32_t(value), out, capacity, written, required)
	})
}

func OrbitScoreReportEqual(left *OrbitScoreReport, right *OrbitScoreReport) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_score_report_equal((*C.SidereonOrbitScoreReport)(pointers[0]), (*C.SidereonOrbitScoreReport)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func OrbitScoreReportHorizonBinEdgesS(report *OrbitScoreReport) ([]float64, error) {
	return withOrbitHandles([]*positioningHandle{report.orbitHandle()}, func(pointers []unsafe.Pointer) ([]float64, error) {
		raw, err := orbitReadRows("orbit score report horizon bin edges s", func(out *C.double, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_score_report_horizon_bin_edges_s((*C.SidereonOrbitScoreReport)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.double) float64 { return float64(item) }), nil
	})
}

func OrbitScoreReportPerSatellite(report *OrbitScoreReport) ([]SatelliteScore, error) {
	return withOrbitHandles([]*positioningHandle{report.orbitHandle()}, func(pointers []unsafe.Pointer) ([]SatelliteScore, error) {
		raw, err := orbitReadRows("orbit score report per satellite", func(out *C.SidereonSatelliteScore, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_score_report_per_satellite((*C.SidereonOrbitScoreReport)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonSatelliteScore) SatelliteScore { return orbitSatelliteScoreFromC(item) }), nil
	})
}

func OrbitScoreReportPerSystem(report *OrbitScoreReport) ([]SystemScore, error) {
	return withOrbitHandles([]*positioningHandle{report.orbitHandle()}, func(pointers []unsafe.Pointer) ([]SystemScore, error) {
		raw, err := orbitReadRows("orbit score report per system", func(out *C.SidereonSystemScore, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_score_report_per_system((*C.SidereonOrbitScoreReport)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonSystemScore) SystemScore { return orbitSystemScoreFromC(item) }), nil
	})
}

func OrbitScoreReportSatelliteBins(report *OrbitScoreReport, index uint64) ([]HorizonBinStats, error) {
	return withOrbitHandles([]*positioningHandle{report.orbitHandle()}, func(pointers []unsafe.Pointer) ([]HorizonBinStats, error) {
		raw, err := orbitReadRows("orbit score report satellite bins", func(out *C.SidereonHorizonBinStats, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_score_report_satellite_bins((*C.SidereonOrbitScoreReport)(pointers[0]), C.size_t(index), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonHorizonBinStats) HorizonBinStats { return orbitHorizonBinStatsFromC(item) }), nil
	})
}

func OrbitScoreReportSummary(report *OrbitScoreReport) (OrbitScoreSummary, error) {
	return withOrbitHandles([]*positioningHandle{report.orbitHandle()}, func(pointers []unsafe.Pointer) (OrbitScoreSummary, error) {
		var value C.SidereonOrbitScoreSummary
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_score_report_summary((*C.SidereonOrbitScoreReport)(pointers[0]), &value))
		})
		return orbitOrbitScoreSummaryFromC(value), err
	})
}

func OrbitScoreReportSystemBins(report *OrbitScoreReport, index uint64) ([]HorizonBinStats, error) {
	return withOrbitHandles([]*positioningHandle{report.orbitHandle()}, func(pointers []unsafe.Pointer) ([]HorizonBinStats, error) {
		raw, err := orbitReadRows("orbit score report system bins", func(out *C.SidereonHorizonBinStats, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_score_report_system_bins((*C.SidereonOrbitScoreReport)(pointers[0]), C.size_t(index), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonHorizonBinStats) HorizonBinStats { return orbitHorizonBinStatsFromC(item) }), nil
	})
}

func OrbitSourceFrameLabel(source *OrbitSource) (string, error) {
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) (string, error) {
		return orbitText("orbit source frame label", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_source_frame_label((*C.SidereonOrbitSource)(pointers[0]), out, capacity, written, required)
		})
	})
}

func ReadOrbitSourceInfo(source *OrbitSource) (OrbitSourceInfo, error) {
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) (OrbitSourceInfo, error) {
		var value C.SidereonOrbitSourceInfo
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_source_info((*C.SidereonOrbitSource)(pointers[0]), &value))
		})
		return orbitOrbitSourceInfoFromC(value), err
	})
}

func OrbitSourcePredictionOriginJ2000S(source *OrbitSource, satellite string) (OptionalOrbitScalar, error) {
	satelliteC, err := copyNativeCString(satellite, "satellite")
	if err != nil {
		return OptionalOrbitScalar{}, err
	}
	defer C.free(satelliteC)
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) (OptionalOrbitScalar, error) {
		var value C.SidereonOptionalOrbitScalar
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_source_prediction_origin_j2000_s((*C.SidereonOrbitSource)(pointers[0]), (*C.char)(satelliteC), &value))
		})
		return orbitOptionalOrbitScalarFromC(value), err
	})
}

func OrbitSourceSP3TimeSystem(source *OrbitSource) (string, error) {
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) (string, error) {
		return orbitText("orbit source sp3 time system", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_source_sp3_time_system((*C.SidereonOrbitSource)(pointers[0]), out, capacity, written, required)
		})
	})
}

func OrbitSourceStateAtJ2000S(source *OrbitSource, satellite string, epochJ2000S float64) (OrbitSourceState, error) {
	satelliteC, err := copyNativeCString(satellite, "satellite")
	if err != nil {
		return OrbitSourceState{}, err
	}
	defer C.free(satelliteC)
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) (OrbitSourceState, error) {
		var value C.SidereonOrbitSourceState
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_source_state_at_j2000_s((*C.SidereonOrbitSource)(pointers[0]), (*C.char)(satelliteC), C.double(epochJ2000S), &value))
		})
		return orbitOrbitSourceStateFromC(value), err
	})
}

func OrbitTrajectoryEpochs(source *OrbitSource) ([]NativeClockEpoch, error) {
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) ([]NativeClockEpoch, error) {
		raw, err := orbitReadRows("orbit trajectory epochs", func(out *C.SidereonClockEpoch, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_trajectory_epochs((*C.SidereonOrbitSource)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonClockEpoch) NativeClockEpoch { return clockEpochFromC(item) }), nil
	})
}

func OrbitTrajectoryEqual(left *OrbitSource, right *OrbitSource) (bool, error) {
	return withOrbitHandles([]*positioningHandle{left.orbitHandle(), right.orbitHandle()}, func(pointers []unsafe.Pointer) (bool, error) {
		var value C.bool
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_trajectory_equal((*C.SidereonOrbitSource)(pointers[0]), (*C.SidereonOrbitSource)(pointers[1]), &value))
		})
		return bool(value), err
	})
}

func OrbitTrajectorySatellites(source *OrbitSource) ([]string, error) {
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) ([]string, error) {
		raw, err := orbitReadRows("orbit trajectory satellites", func(out *C.SidereonSatelliteToken, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_orbit_trajectory_satellites((*C.SidereonOrbitSource)(pointers[0]), out, capacity, written, required)
		})
		if err != nil {
			return nil, err
		}
		return orbitMap(raw, func(item C.SidereonSatelliteToken) string { return tokenFromC(item) }), nil
	})
}

func OrbitTrajectoryWithClockRelativityIncluded(source *OrbitSource, included bool) (*OrbitSource, error) {
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) (*OrbitSource, error) {
		var value *C.SidereonOrbitSource
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_orbit_trajectory_with_clock_relativity_included((*C.SidereonOrbitSource)(pointers[0]), C.bool(included), &value))
		})
		return newOrbitOrbitSource(value, err)
	})
}

func PredictedSP3FromPredictionFlags(product *SP3) (*OrbitSource, error) {
	var value *C.SidereonOrbitSource
	err := orbitSP3Handle(product).with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_predicted_sp3_from_prediction_flags((*C.SidereonSp3)(pointer), &value))
		})
	})
	return newOrbitOrbitSource(value, err)
}

func PredictedSP3Product(source *OrbitSource) (*SP3, error) {
	return withOrbitHandles([]*positioningHandle{source.orbitHandle()}, func(pointers []unsafe.Pointer) (*SP3, error) {
		var value *C.SidereonSp3
		err := callStatus(func() uint32 {
			return uint32(C.sidereon_predicted_sp3_product((*C.SidereonOrbitSource)(pointers[0]), &value))
		})
		return orbitAdoptSP3(value, err)
	})
}

func PredictedSP3WithCutoff(product *SP3, cutoff NativeClockEpoch) (*OrbitSource, error) {
	var value *C.SidereonOrbitSource
	err := orbitSP3Handle(product).with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_predicted_sp3_with_cutoff((*C.SidereonSp3)(pointer), clockEpochToC(cutoff), &value))
		})
	})
	return newOrbitOrbitSource(value, err)
}

func ComputeRTNComponents(differenceM [3]float64, positionM [3]float64, velocityMS [3]float64) (RTNComponents, error) {
	var value C.SidereonRtnComponents
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_rtn_components(orbitVectorToC(differenceM), orbitVectorToC(positionM), orbitVectorToC(velocityMS), &value))
	})
	return orbitRTNComponentsFromC(value), err
}

func ScoredProductText(value ScoredProduct) (string, error) {
	return orbitText("scored product text", func(out *C.uint8_t, capacity C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_scored_product_text(C.uint32_t(value), out, capacity, written, required)
	})
}

func SISREWeightSelectionEqual(left, right *SISREWeights) (bool, error) {
	var leftC, rightC *C.SidereonSisreWeights
	if left != nil {
		value := orbitSISREWeightsToC(*left)
		leftC = &value
	}
	if right != nil {
		value := orbitSISREWeightsToC(*right)
		rightC = &value
	}
	var value C.bool
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_sisre_weight_selection_equal(leftC, rightC, &value))
	})
	return bool(value), err
}

func SISREWeightsMontenbruck2018(system uint32, regime OrbitRegime) (*SISREWeights, error) {
	var present C.bool
	var value C.SidereonSisreWeights
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_sisre_weights_montenbruck_2018(C.uint32_t(system), C.uint32_t(regime), &present, &value))
	})
	if err != nil || !bool(present) {
		return nil, err
	}
	result := orbitSISREWeightsFromC(value)
	return &result, nil
}

func SISREWeightsOrbitOnlyM(weights SISREWeights, components RTNComponents) (float64, error) {
	var value C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_sisre_weights_orbit_only_m(orbitSISREWeightsToC(weights), orbitRTNComponentsToC(components), &value))
	})
	return float64(value), err
}

func SISREWeightsWithClockM(weights SISREWeights, components RTNComponents, clockM float64) (float64, error) {
	var value C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_sisre_weights_with_clock_m(orbitSISREWeightsToC(weights), orbitRTNComponentsToC(components), C.double(clockM), &value))
	})
	return float64(value), err
}

func OrbitTrajectoryNew(frameLabel string, predictionOrigin NativeClockEpoch, samples []TrajectorySample) (*OrbitSource, error) {
	rows := make([]C.SidereonTrajectorySample, len(samples))
	for i, sample := range samples {
		row, err := orbitTrajectorySampleToC(sample)
		if err != nil {
			return nil, err
		}
		rows[i] = row
	}
	label := []byte(frameLabel)
	var value *C.SidereonOrbitSource
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_orbit_trajectory_new((*C.uint8_t)(unsafe.Pointer(unsafe.SliceData(label))), C.size_t(len(label)), clockEpochToC(predictionOrigin), unsafe.SliceData(rows), C.size_t(len(rows)), &value))
	})
	return newOrbitOrbitSource(value, err)
}

func OrbitScoreOptionsDefault(horizonBinEdgesS []float64) OrbitScoreOptions {
	value := C.sidereon_orbit_score_options_default(nil, 0)
	return OrbitScoreOptions{
		HorizonBinEdgesS: slices.Clone(horizonBinEdgesS), Triad: TriadVelocity(value.triad),
		VelocityHalfStepS: float64(value.velocity_half_step_s), ClockDatum: ClockDatum(value.clock_datum),
		Gaps: OrbitScoreGapPolicy(value.gaps),
	}
}

func ScoreOrbit(source *OrbitSource, reference *SP3, satellites []string, options OrbitScoreOptions) (*OrbitScoreReport, error) {
	var pins runtime.Pinner
	defer pins.Unpin()
	raw := orbitScoreOptionsToC(options, &pins)
	return withOrbitHandles([]*positioningHandle{source.orbitHandle(), orbitSP3Handle(reference)}, func(pointers []unsafe.Pointer) (*OrbitScoreReport, error) {
		var value *C.SidereonOrbitScoreReport
		err := withCStringArray(satellites, func(tokens **C.char, count C.size_t) error {
			return callStatus(func() uint32 {
				return uint32(C.sidereon_score_orbit((*C.SidereonOrbitSource)(pointers[0]), (*C.SidereonSp3)(pointers[1]), tokens, count, &raw, &value))
			})
		})
		return newOrbitOrbitScoreReport(value, err)
	})
}

func orbitTrajectorySampleToC(sample TrajectorySample) (C.SidereonTrajectorySample, error) {
	token, err := tokenToC(sample.Satellite)
	if err != nil {
		return C.SidereonTrajectorySample{}, err
	}
	return C.SidereonTrajectorySample{
		satellite: token, epoch: clockEpochToC(sample.Epoch), position_ecef_m: orbitVectorToC(sample.PositionECEFM).xyz,
		has_clock_s: C.bool(sample.ClockS.Present), clock_s: C.double(sample.ClockS.Value),
	}, nil
}

func TrajectorySampleEqual(left, right TrajectorySample) (bool, error) {
	leftC, err := orbitTrajectorySampleToC(left)
	if err != nil {
		return false, err
	}
	rightC, err := orbitTrajectorySampleToC(right)
	if err != nil {
		return false, err
	}
	var equal C.bool
	err = callStatus(func() uint32 {
		return uint32(C.sidereon_trajectory_sample_equal(leftC, rightC, &equal))
	})
	return bool(equal), err
}

func ScoreEpochsEqual(left, right []NativeClockEpoch) (bool, error) {
	var pins runtime.Pinner
	defer pins.Unpin()
	leftC := orbitScoreEpochsToC(left, &pins)
	rightC := orbitScoreEpochsToC(right, &pins)
	var equal C.bool
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_score_epochs_equal(leftC, rightC, &equal))
	})
	return bool(equal), err
}

func SP3FrameReconciliationOptionsEqual(leftSets [][]string, leftHelmert bool, rightSets [][]string, rightHelmert bool) (bool, error) {
	var pins runtime.Pinner
	defer pins.Unpin()
	leftC := orbitFrameReconciliationOptionsToC(leftSets, leftHelmert, &pins)
	rightC := orbitFrameReconciliationOptionsToC(rightSets, rightHelmert, &pins)
	var equal C.bool
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_sp3_frame_reconciliation_options_equal(leftC, rightC, &equal))
	})
	return bool(equal), err
}

func OrbitScoreOptionsEqual(left, right OrbitScoreOptions) (bool, error) {
	var pins runtime.Pinner
	defer pins.Unpin()
	leftC := orbitScoreOptionsToC(left, &pins)
	rightC := orbitScoreOptionsToC(right, &pins)
	var equal C.bool
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_orbit_score_options_equal(leftC, rightC, &equal))
	})
	return bool(equal), err
}

func orbitScoreEpochsToC(epochs []NativeClockEpoch, pins *runtime.Pinner) C.SidereonScoreEpochs {
	instants := orbitMap(epochs, clockEpochToC)
	return C.SidereonScoreEpochs{explicit_instants: C.bool(epochs != nil), instants: pinOrbitSlice(pins, instants), count: C.size_t(len(instants))}
}

func orbitFrameReconciliationOptionsToC(labelSets [][]string, helmert bool, pins *runtime.Pinner) C.SidereonSp3FrameReconciliationOptions {
	sets := make([]C.SidereonFrameLabelSet, len(labelSets))
	for i, group := range labelSets {
		labels := make([]C.SidereonFrameLabel, len(group))
		for j, label := range group {
			bytes := []byte(label)
			labels[j] = C.SidereonFrameLabel{data: (*C.uint8_t)(unsafe.Pointer(pinOrbitSlice(pins, bytes))), len: C.size_t(len(bytes))}
		}
		sets[i] = C.SidereonFrameLabelSet{labels: pinOrbitSlice(pins, labels), label_count: C.size_t(len(labels))}
	}
	return C.SidereonSp3FrameReconciliationOptions{
		asserted_equivalent_label_sets: pinOrbitSlice(pins, sets), label_set_count: C.size_t(len(sets)), helmert: C.bool(helmert),
	}
}

func orbitScoreOptionsToC(options OrbitScoreOptions, pins *runtime.Pinner) C.SidereonOrbitScoreOptions {
	var weights *C.SidereonSisreWeights
	if options.SISREWeights != nil {
		value := orbitSISREWeightsToC(*options.SISREWeights)
		weights = &value
		pins.Pin(weights)
	}
	return C.SidereonOrbitScoreOptions{
		epochs:                 orbitScoreEpochsToC(options.Epochs, pins),
		horizon_bin_edges_s:    (*C.double)(unsafe.Pointer(pinOrbitSlice(pins, options.HorizonBinEdgesS))),
		horizon_bin_edge_count: C.size_t(len(options.HorizonBinEdgesS)),
		triad:                  C.uint32_t(options.Triad), velocity_half_step_s: C.double(options.VelocityHalfStepS),
		sisre_weights: weights, clock_datum: C.uint32_t(options.ClockDatum), gaps: C.uint32_t(options.Gaps),
		frame_reconciliation: orbitFrameReconciliationOptionsToC(options.FrameLabelSets, options.Helmert, pins),
	}
}
