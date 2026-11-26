//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

func BroadcastOrbitNew(_ *BroadcastEphemeris, _ NativeClockEpoch) (*OrbitSource, error) {
	return nil, ErrUnavailable
}

func DefaultVelocityHalfStepS() float64 { return 0 }

func InertialVelocityFromEarthFixed(_ [3]float64, _ [3]float64) ([3]float64, error) {
	return [3]float64{}, ErrUnavailable
}

func MinVelocityHalfStepS() float64 { return 0 }

func OrbitRegimeFromInertialState(_ [3]float64, _ [3]float64) (*OrbitRegime, error) {
	return nil, ErrUnavailable
}

func OrbitRegimeText(_ OrbitRegime) (string, error) { return "", ErrUnavailable }

func OrbitScoreReportEqual(_ *OrbitScoreReport, _ *OrbitScoreReport) (bool, error) {
	return false, ErrUnavailable
}

func OrbitScoreReportHorizonBinEdgesS(_ *OrbitScoreReport) ([]float64, error) {
	return nil, ErrUnavailable
}

func OrbitScoreReportPerSatellite(_ *OrbitScoreReport) ([]SatelliteScore, error) {
	return nil, ErrUnavailable
}

func OrbitScoreReportPerSystem(_ *OrbitScoreReport) ([]SystemScore, error) {
	return nil, ErrUnavailable
}

func OrbitScoreReportSatelliteBins(_ *OrbitScoreReport, _ uint64) ([]HorizonBinStats, error) {
	return nil, ErrUnavailable
}

func OrbitScoreReportSummary(_ *OrbitScoreReport) (OrbitScoreSummary, error) {
	return OrbitScoreSummary{}, ErrUnavailable
}

func OrbitScoreReportSystemBins(_ *OrbitScoreReport, _ uint64) ([]HorizonBinStats, error) {
	return nil, ErrUnavailable
}

func OrbitSourceFrameLabel(_ *OrbitSource) (string, error) { return "", ErrUnavailable }

func ReadOrbitSourceInfo(_ *OrbitSource) (OrbitSourceInfo, error) {
	return OrbitSourceInfo{}, ErrUnavailable
}

func OrbitSourcePredictionOriginJ2000S(_ *OrbitSource, _ string) (OptionalOrbitScalar, error) {
	return OptionalOrbitScalar{}, ErrUnavailable
}

func OrbitSourceSP3TimeSystem(_ *OrbitSource) (string, error) { return "", ErrUnavailable }

func OrbitSourceStateAtJ2000S(_ *OrbitSource, _ string, _ float64) (OrbitSourceState, error) {
	return OrbitSourceState{}, ErrUnavailable
}

func OrbitTrajectoryEpochs(_ *OrbitSource) ([]NativeClockEpoch, error) { return nil, ErrUnavailable }

func OrbitTrajectoryEqual(_ *OrbitSource, _ *OrbitSource) (bool, error) { return false, ErrUnavailable }

func OrbitTrajectorySatellites(_ *OrbitSource) ([]string, error) { return nil, ErrUnavailable }

func OrbitTrajectoryWithClockRelativityIncluded(_ *OrbitSource, _ bool) (*OrbitSource, error) {
	return nil, ErrUnavailable
}

func PredictedSP3FromPredictionFlags(_ *SP3) (*OrbitSource, error) { return nil, ErrUnavailable }

func PredictedSP3Product(_ *OrbitSource) (*SP3, error) { return nil, ErrUnavailable }

func PredictedSP3WithCutoff(_ *SP3, _ NativeClockEpoch) (*OrbitSource, error) {
	return nil, ErrUnavailable
}

func ComputeRTNComponents(_ [3]float64, _ [3]float64, _ [3]float64) (RTNComponents, error) {
	return RTNComponents{}, ErrUnavailable
}

func ScoredProductText(_ ScoredProduct) (string, error) { return "", ErrUnavailable }

func SISREWeightSelectionEqual(_ *SISREWeights, _ *SISREWeights) (bool, error) {
	return false, ErrUnavailable
}

func SISREWeightsMontenbruck2018(_ uint32, _ OrbitRegime) (*SISREWeights, error) {
	return nil, ErrUnavailable
}

func SISREWeightsOrbitOnlyM(_ SISREWeights, _ RTNComponents) (float64, error) {
	return 0, ErrUnavailable
}

func SISREWeightsWithClockM(_ SISREWeights, _ RTNComponents, _ float64) (float64, error) {
	return 0, ErrUnavailable
}

func OrbitTrajectoryNew(string, NativeClockEpoch, []TrajectorySample) (*OrbitSource, error) {
	return nil, ErrUnavailable
}
func OrbitScoreOptionsDefault(edges []float64) OrbitScoreOptions {
	return OrbitScoreOptions{HorizonBinEdgesS: edges}
}
func ScoreOrbit(*OrbitSource, *SP3, []string, OrbitScoreOptions) (*OrbitScoreReport, error) {
	return nil, ErrUnavailable
}

func TrajectorySampleEqual(TrajectorySample, TrajectorySample) (bool, error) {
	return false, ErrUnavailable
}

func ScoreEpochsEqual([]NativeClockEpoch, []NativeClockEpoch) (bool, error) {
	return false, ErrUnavailable
}

func SP3FrameReconciliationOptionsEqual([][]string, bool, [][]string, bool) (bool, error) {
	return false, ErrUnavailable
}

func OrbitScoreOptionsEqual(OrbitScoreOptions, OrbitScoreOptions) (bool, error) {
	return false, ErrUnavailable
}
