package sidereon

import "sidereon.dev/go/v4/internal/native"

// NewBroadcastOrbit constructs a broadcast scoring source that retains the
// navigation data and message preference at construction.
func NewBroadcastOrbit(ephemeris *BroadcastEphemeris, predictionOrigin ClockEpoch) (*BroadcastOrbit, error) {
	value, err := native.BroadcastOrbitNew(nativeBroadcastEphemerisHandle(ephemeris), nativeClockEpoch(predictionOrigin))
	if err != nil {
		return nil, publicError(err)
	}
	return publicBroadcastOrbit(value), nil
}

// DefaultVelocityHalfStepS returns the default velocity half step s.
func DefaultVelocityHalfStepS() float64 {
	return native.DefaultVelocityHalfStepS()
}

// InertialVelocityFromEarthFixed returns the inertial velocity from earth fixed.
func InertialVelocityFromEarthFixed(positionECEFM [3]float64, velocityECEFMS [3]float64) ([3]float64, error) {
	value, err := native.InertialVelocityFromEarthFixed(positionECEFM, velocityECEFMS)
	return value, publicError(err)
}

// MinVelocityHalfStepS returns the min velocity half step s.
func MinVelocityHalfStepS() float64 {
	return native.MinVelocityHalfStepS()
}

// OrbitRegimeFromInertialState classifies an Earth orbit from inertial position and velocity.
func OrbitRegimeFromInertialState(positionM [3]float64, inertialVelocityMS [3]float64) (*OrbitRegime, error) {
	value, err := native.OrbitRegimeFromInertialState(positionM, inertialVelocityMS)
	return orbitOptional(value, func(value native.OrbitRegime) OrbitRegime { return OrbitRegime(value) }), publicError(err)
}

// OrbitRegimeText returns the text.
func OrbitRegimeText(value OrbitRegime) (string, error) {
	text, err := native.OrbitRegimeText(native.OrbitRegime(value))
	return text, publicError(err)
}

// Equal compares the underlying values.
func (left *OrbitScoreReport) Equal(right *OrbitScoreReport) (bool, error) {
	value, err := native.OrbitScoreReportEqual(nativeOrbitScoreReportHandle(left), nativeOrbitScoreReportHandle(right))
	return value, publicError(err)
}

// HorizonBinEdgesS returns the horizon bin edges s.
func (report *OrbitScoreReport) HorizonBinEdgesS() ([]float64, error) {
	value, err := native.OrbitScoreReportHorizonBinEdgesS(nativeOrbitScoreReportHandle(report))
	return value, publicError(err)
}

// PerSatellite returns statistics and coverage for each requested satellite.
func (report *OrbitScoreReport) PerSatellite() ([]SatelliteScore, error) {
	value, err := native.OrbitScoreReportPerSatellite(nativeOrbitScoreReportHandle(report))
	return orbitMap(value, func(value native.SatelliteScore) SatelliteScore { return publicSatelliteScore(value) }), publicError(err)
}

// PerSystem returns aggregate statistics and coverage by constellation.
func (report *OrbitScoreReport) PerSystem() ([]SystemScore, error) {
	value, err := native.OrbitScoreReportPerSystem(nativeOrbitScoreReportHandle(report))
	return orbitMap(value, func(value native.SystemScore) SystemScore { return publicSystemScore(value) }), publicError(err)
}

// SatelliteBins returns the satellite bins.
func (report *OrbitScoreReport) SatelliteBins(index uint64) ([]HorizonBinStats, error) {
	value, err := native.OrbitScoreReportSatelliteBins(nativeOrbitScoreReportHandle(report), index)
	return orbitMap(value, func(value native.HorizonBinStats) HorizonBinStats { return publicHorizonBinStats(value) }), publicError(err)
}

// Summary returns frame agreement, reference point, counts and coverage.
func (report *OrbitScoreReport) Summary() (OrbitScoreSummary, error) {
	value, err := native.OrbitScoreReportSummary(nativeOrbitScoreReportHandle(report))
	return publicOrbitScoreSummary(value), publicError(err)
}

// SystemBins returns the system bins.
func (report *OrbitScoreReport) SystemBins(index uint64) ([]HorizonBinStats, error) {
	value, err := native.OrbitScoreReportSystemBins(nativeOrbitScoreReportHandle(report), index)
	return orbitMap(value, func(value native.HorizonBinStats) HorizonBinStats { return publicHorizonBinStats(value) }), publicError(err)
}

// FrameLabel returns the frame label.
func (source *orbitSourceHandle) FrameLabel() (string, error) {
	value, err := native.OrbitSourceFrameLabel(nativeOrbitSourceValue(source))
	return value, publicError(err)
}

// Info returns the info.
func (source *orbitSourceHandle) Info() (OrbitSourceInfo, error) {
	value, err := native.ReadOrbitSourceInfo(nativeOrbitSourceValue(source))
	return publicOrbitSourceInfo(value), publicError(err)
}

// PredictionOriginJ2000S returns the prediction origin j2000 s.
func (source *orbitSourceHandle) PredictionOriginJ2000S(satellite string) (OptionalOrbitScalar, error) {
	value, err := native.OrbitSourcePredictionOriginJ2000S(nativeOrbitSourceValue(source), satellite)
	return publicOptionalOrbitScalar(value), publicError(err)
}

// SP3TimeSystem returns the SP3 time system.
func (source *orbitSourceHandle) SP3TimeSystem() (string, error) {
	value, err := native.OrbitSourceSP3TimeSystem(nativeOrbitSourceValue(source))
	return value, publicError(err)
}

// StateAtJ2000S evaluates the Earth-fixed state in the source time scale.
func (source *orbitSourceHandle) StateAtJ2000S(satellite string, epochJ2000S float64) (OrbitSourceState, error) {
	value, err := native.OrbitSourceStateAtJ2000S(nativeOrbitSourceValue(source), satellite, epochJ2000S)
	return publicOrbitSourceState(value), publicError(err)
}

// Epochs returns the epochs.
func (source *OrbitTrajectory) Epochs() ([]ClockEpoch, error) {
	value, err := native.OrbitTrajectoryEpochs(nativeOrbitTrajectoryHandle(source))
	return orbitMap(value, func(value native.NativeClockEpoch) ClockEpoch { return publicClockEpoch(value) }), publicError(err)
}

// Equal compares the underlying values.
func (left *OrbitTrajectory) Equal(right *OrbitTrajectory) (bool, error) {
	value, err := native.OrbitTrajectoryEqual(nativeOrbitTrajectoryHandle(left), nativeOrbitTrajectoryHandle(right))
	return value, publicError(err)
}

// Satellites returns the satellites.
func (source *OrbitTrajectory) Satellites() ([]string, error) {
	value, err := native.OrbitTrajectorySatellites(nativeOrbitTrajectoryHandle(source))
	return value, publicError(err)
}

// WithClockRelativityIncluded returns a copy with the requested change.
func (source *OrbitTrajectory) WithClockRelativityIncluded(included bool) (*OrbitTrajectory, error) {
	value, err := native.OrbitTrajectoryWithClockRelativityIncluded(nativeOrbitTrajectoryHandle(source), included)
	if err != nil {
		return nil, publicError(err)
	}
	return publicOrbitTrajectory(value), nil
}

// PredictedSP3FromPredictionFlags retains the product and reads each prediction
// origin from its flags.
func PredictedSP3FromPredictionFlags(product *SP3) (*PredictedSP3, error) {
	value, err := native.PredictedSP3FromPredictionFlags(nativeSP3Handle(product))
	if err != nil {
		return nil, publicError(err)
	}
	return publicPredictedSP3(value), nil
}

// Product returns the product.
func (source *PredictedSP3) Product() (*SP3, error) {
	value, err := native.PredictedSP3Product(nativePredictedSP3Handle(source))
	if err != nil {
		return nil, publicError(err)
	}
	return publicSP3Handle(value), nil
}

// PredictedSP3WithCutoff retains the product and sets a common prediction
// origin for every satellite.
func PredictedSP3WithCutoff(product *SP3, cutoff ClockEpoch) (*PredictedSP3, error) {
	value, err := native.PredictedSP3WithCutoff(nativeSP3Handle(product), nativeClockEpoch(cutoff))
	if err != nil {
		return nil, publicError(err)
	}
	return publicPredictedSP3(value), nil
}

// ComputeRTNComponents returns the RTN components.
func ComputeRTNComponents(differenceM [3]float64, positionM [3]float64, velocityMS [3]float64) (RTNComponents, error) {
	value, err := native.ComputeRTNComponents(differenceM, positionM, velocityMS)
	return publicRTNComponents(value), publicError(err)
}

// ScoredProductText returns the text.
func ScoredProductText(value ScoredProduct) (string, error) {
	text, err := native.ScoredProductText(native.ScoredProduct(value))
	return text, publicError(err)
}

// Equal compares two published or fixed weighting selections.
func (selection SISREWeightSelection) Equal(other SISREWeightSelection) (bool, error) {
	value, err := native.SISREWeightSelectionEqual(orbitOptional(selection.Fixed, nativeSISREWeights), orbitOptional(other.Fixed, nativeSISREWeights))
	return value, publicError(err)
}

// SISREWeightsMontenbruck2018 returns the montenbruck 2018.
func SISREWeightsMontenbruck2018(system GNSSSystem, regime OrbitRegime) (*SISREWeights, error) {
	value, err := native.SISREWeightsMontenbruck2018(uint32(system), native.OrbitRegime(regime))
	return orbitOptional(value, func(value native.SISREWeights) SISREWeights { return publicSISREWeights(value) }), publicError(err)
}

// OrbitOnlyM returns the orbit only m.
func (weights SISREWeights) OrbitOnlyM(components RTNComponents) (float64, error) {
	value, err := native.SISREWeightsOrbitOnlyM(nativeSISREWeights(weights), nativeRTNComponents(components))
	return value, publicError(err)
}

// WithClockM returns a copy with the requested change.
func (weights SISREWeights) WithClockM(components RTNComponents, clockM float64) (float64, error) {
	value, err := native.SISREWeightsWithClockM(nativeSISREWeights(weights), nativeRTNComponents(components), clockM)
	return value, publicError(err)
}

// ScoreEpochs selects reference record epochs or explicitly supplied instants.
type ScoreEpochs struct {
	// Instants supplies evaluation epochs. A nil slice selects the reference record epochs.
	Instants []ClockEpoch
}

// Equal compares two epoch selections using the engine's value comparison.
func (epochs ScoreEpochs) Equal(other ScoreEpochs) (bool, error) {
	equal, err := native.ScoreEpochsEqual(orbitMap(epochs.Instants, nativeClockEpoch), orbitMap(other.Instants, nativeClockEpoch))
	return equal, publicError(err)
}

// SISREWeightSelection selects published constellation weights or fixed weights.
type SISREWeightSelection struct {
	// Fixed supplies the same weights for every satellite. Nil selects published weights.
	Fixed *SISREWeights
}

// SP3FrameReconciliationOptions controls how terrestrial frames are reconciled.
type SP3FrameReconciliationOptions struct {
	// AssertedEquivalentLabelSets groups labels that describe the same frame.
	AssertedEquivalentLabelSets [][]string
	// Helmert enables catalog Helmert transforms.
	Helmert bool
}

// Equal compares frame reconciliation controls by value.
func (options SP3FrameReconciliationOptions) Equal(other SP3FrameReconciliationOptions) (bool, error) {
	equal, err := native.SP3FrameReconciliationOptionsEqual(options.AssertedEquivalentLabelSets, options.Helmert, other.AssertedEquivalentLabelSets, other.Helmert)
	return equal, publicError(err)
}

// OrbitScoreOptions configures orbit and clock comparisons against an SP3 product.
type OrbitScoreOptions struct {
	// Epochs selects the evaluation instants.
	Epochs ScoreEpochs
	// HorizonBinEdgesS gives increasing prediction horizon bin boundaries in seconds.
	HorizonBinEdgesS []float64
	// Triad selects the velocity used to form the reference RTN axes.
	Triad TriadVelocity
	// VelocityHalfStepS sets the reference velocity difference half step in seconds.
	VelocityHalfStepS float64
	// SISREWeights selects the weighting applied to orbital errors.
	SISREWeights SISREWeightSelection
	// ClockDatum controls clock datum adjustment.
	ClockDatum ClockDatum
	// Gaps determines how missing source or reference states are handled.
	Gaps OrbitScoreGapPolicy
	// FrameReconciliation determines how frame labels are matched and transformed.
	FrameReconciliation SP3FrameReconciliationOptions
}

// Equal compares orbit scoring controls by value.
func (options OrbitScoreOptions) Equal(other OrbitScoreOptions) (bool, error) {
	equal, err := native.OrbitScoreOptionsEqual(nativeOrbitScoreOptions(options), nativeOrbitScoreOptions(other))
	return equal, publicError(err)
}

// DefaultOrbitScoreOptions constructs the scoring defaults for the supplied horizon bins.
func DefaultOrbitScoreOptions(horizonBinEdgesS []float64) OrbitScoreOptions {
	value := native.OrbitScoreOptionsDefault(horizonBinEdgesS)
	return OrbitScoreOptions{
		HorizonBinEdgesS: value.HorizonBinEdgesS,
		Triad:            TriadVelocity(value.Triad), VelocityHalfStepS: value.VelocityHalfStepS,
		ClockDatum: ClockDatum(value.ClockDatum), Gaps: OrbitScoreGapPolicy(value.Gaps),
	}
}

// TrajectorySample is one satellite's Earth-fixed position and optional clock offset.
type TrajectorySample struct {
	// Satellite is the satellite token.
	Satellite string
	// Epoch is the time-scale-tagged observation instant.
	Epoch ClockEpoch
	// PositionECEFM gives the position in the trajectory's frame in metres.
	PositionECEFM [3]float64
	// ClockS contains the clock offset in seconds when present.
	ClockS OptionalOrbitScalar
}

// Equal compares trajectory samples by value, including their optional clocks.
func (sample TrajectorySample) Equal(other TrajectorySample) (bool, error) {
	equal, err := native.TrajectorySampleEqual(nativeTrajectorySample(sample), nativeTrajectorySample(other))
	return equal, publicError(err)
}

func nativeTrajectorySample(value TrajectorySample) native.TrajectorySample {
	return native.TrajectorySample{
		Satellite: value.Satellite, Epoch: nativeClockEpoch(value.Epoch),
		PositionECEFM: value.PositionECEFM, ClockS: nativeOptionalOrbitScalar(value.ClockS),
	}
}

// NewOrbitTrajectory creates an orbit source from sampled Earth-fixed states.
func NewOrbitTrajectory(frameLabel string, predictionOrigin ClockEpoch, samples []TrajectorySample) (*OrbitTrajectory, error) {
	values := orbitMap(samples, nativeTrajectorySample)
	handle, err := native.OrbitTrajectoryNew(frameLabel, nativeClockEpoch(predictionOrigin), values)
	if err != nil {
		return nil, publicError(err)
	}
	return publicOrbitTrajectory(handle), nil
}

func nativeOrbitScoreOptions(options OrbitScoreOptions) native.OrbitScoreOptions {
	return native.OrbitScoreOptions{
		Epochs:           orbitMap(options.Epochs.Instants, nativeClockEpoch),
		HorizonBinEdgesS: options.HorizonBinEdgesS,
		Triad:            native.TriadVelocity(options.Triad), VelocityHalfStepS: options.VelocityHalfStepS,
		SISREWeights: orbitOptional(options.SISREWeights.Fixed, nativeSISREWeights),
		ClockDatum:   native.ClockDatum(options.ClockDatum), Gaps: native.OrbitScoreGapPolicy(options.Gaps),
		FrameLabelSets: options.FrameReconciliation.AssertedEquivalentLabelSets,
		Helmert:        options.FrameReconciliation.Helmert,
	}
}

// ScoreOrbit compares an orbit source against a reference SP3 product.
func ScoreOrbit(source OrbitSource, reference *SP3, satellites []string, options OrbitScoreOptions) (*OrbitScoreReport, error) {
	value, err := native.ScoreOrbit(nativeOrbitSourceHandle(source), nativeSP3Handle(reference), satellites, nativeOrbitScoreOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	return publicOrbitScoreReport(value), nil
}

func nativeOptionalOrbitScalar(value OptionalOrbitScalar) native.OptionalOrbitScalar {
	return native.OptionalOrbitScalar(value)
}

// String returns the display name of the orbital regime.
func (regime OrbitRegime) String() string {
	value, _ := OrbitRegimeText(regime)
	return value
}

// String returns the role of the product in orbit scoring.
func (product ScoredProduct) String() string {
	value, _ := ScoredProductText(product)
	return value
}

// OrbitFrame identifies the terrestrial frame of an orbit source.
// Use Equal to compare values.
type OrbitFrame struct {
	// Kind distinguishes named frames from broadcast datums.
	Kind OrbitFrameKind
	// FrameLabel is the full label of a named frame.
	FrameLabel string
}

// Equal compares the frame kind and the label of a named frame.
func (frame OrbitFrame) Equal(other OrbitFrame) bool {
	return frame.Kind == other.Kind && (frame.Kind == OrbitFrameKindBroadcastDatums || frame.FrameLabel == other.FrameLabel)
}

// Frame returns the terrestrial frame of the source.
func (source *orbitSourceHandle) Frame() (OrbitFrame, error) {
	info, err := source.Info()
	if err != nil {
		return OrbitFrame{}, err
	}
	if info.Frame == OrbitFrameKindBroadcastDatums {
		return OrbitFrame{Kind: info.Frame}, nil
	}
	label, err := source.FrameLabel()
	return OrbitFrame{Kind: info.Frame, FrameLabel: label}, err
}

// TimeScale returns the time scale used to evaluate source states.
func (source *orbitSourceHandle) TimeScale() (TimeScale, error) {
	info, err := source.Info()
	return info.TimeScale, err
}

// ReferencePoint identifies the spacecraft reference point of source positions.
func (source *orbitSourceHandle) ReferencePoint() (OrbitReferencePoint, error) {
	info, err := source.Info()
	return info.ReferencePoint, err
}

// ClockIncludesRelativity reports whether the source clock includes the periodic relativistic term.
func (source *orbitSourceHandle) ClockIncludesRelativity() (bool, error) {
	info, err := source.Info()
	return info.ClockIncludesRelativity, err
}
