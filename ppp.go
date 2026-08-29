package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// PPPTropoMapping selects the tropospheric mapping model used by PPP.
type PPPTropoMapping uint32

const (
	// PPPTropoMappingNiell selects the climatological Niell model.
	PPPTropoMappingNiell PPPTropoMapping = PPPTropoMapping(native.PppTropoMappingNiellValue)
	// PPPTropoMappingVMF1 selects the VMF1 site-sample model.
	PPPTropoMappingVMF1 PPPTropoMapping = PPPTropoMapping(native.PppTropoMappingVMF1Value)
)

// PPPSolveStatus is the terminal status reported by a PPP solve.
type PPPSolveStatus uint32

const (
	// PPPSolveStateTolerance means the state tolerances were reached.
	PPPSolveStateTolerance PPPSolveStatus = PPPSolveStatus(native.PppSolveStateToleranceValue)
	// PPPSolveMaxIterations means the iteration bound was reached.
	PPPSolveMaxIterations PPPSolveStatus = PPPSolveStatus(native.PppSolveMaxIterationsValue)
)

// PPPIntegerStatus reports whether the integer ambiguity search was accepted.
type PPPIntegerStatus uint32

const (
	// PPPIntegerFixed reports an accepted integer fix.
	PPPIntegerFixed PPPIntegerStatus = PPPIntegerStatus(native.PppIntegerFixedValue)
	// PPPIntegerNotFixed reports that no integer fix was accepted.
	PPPIntegerNotFixed PPPIntegerStatus = PPPIntegerStatus(native.PppIntegerNotFixedValue)
)

// PPPAutoInitOptions controls SPP-seeded PPP initialization.
type PPPAutoInitOptions struct {
	// HasInitialGuess reports whether InitialGuessPosition and InitialGuessClockM override the native seed.
	HasInitialGuess bool
	// InitialGuessPosition is an explicit ECEF receiver-position seed in metres when HasInitialGuess is true.
	InitialGuessPosition [3]float64
	// InitialGuessClockM is the initial guess clock m in metres.
	InitialGuessClockM float64
	// SPPInitialGuess is the SPP cold-start state [x_m, y_m, z_m, clock_m] used for per-epoch seed solves.
	SPPInitialGuess [4]float64
	// SPPTroposphere enables troposphere correction during each SPP seed solve; ionosphere is disabled there.
	SPPTroposphere bool
	// SPPPressureHPA is the SPP seed-solve atmospheric pressure in hectopascals.
	SPPPressureHPA float64
	// SPPTemperatureK is the spp temperature k in kelvin.
	SPPTemperatureK float64
	// SPPRelativeHumidity is the SPP seed relative humidity as a fraction in [0, 1].
	SPPRelativeHumidity float64
}

// PPPObservation is one ionosphere-free code/phase measurement. Distances are
// metres and frequencies are hertz.
type PPPObservation struct {
	// SatelliteID is the GNSS satellite identifier for the observation.
	SatelliteID string
	// AmbiguityID identifies the carrier-phase ambiguity associated with the observation.
	AmbiguityID string
	// CodeM is the code m in metres.
	CodeM float64
	// PhaseM is the phase m in metres.
	PhaseM float64
	// Frequency1Hz is the frequency1 hz in hertz.
	Frequency1Hz float64
	// Frequency2Hz is the frequency2 hz in hertz.
	Frequency2Hz float64
}

// PPPEpoch is one static PPP epoch and its observation rows.
type PPPEpoch struct {
	// Civil is the civil-time observable.
	Civil CivilDateTime
	// JDWhole is the associated Julian-date value.
	JDWhole float64
	// JDFraction is the associated Julian-date value.
	JDFraction float64
	// TRxJ2000S is the receiver epoch in seconds from J2000.
	TRxJ2000S float64
	// Observations contains a detached copy; nil means this field is absent.
	Observations []PPPObservation
}

// PPPFloatMapEntry is one ambiguity identifier/value pair.
type PPPFloatMapEntry struct {
	// ID identifies the ambiguity represented by Value.
	ID string
	// Value is the value stored in this record.
	Value float64
}

// PPPFloatState is the initial static-arc PPP state. Clock values are metres
// and must contain one entry per epoch.
type PPPFloatState struct {
	// PositionM is the position m in metres.
	PositionM [3]float64
	// ClocksM is the clocks m in metres.
	ClocksM []float64
	// AmbiguitiesM is the ambiguities m in metres.
	AmbiguitiesM []PPPFloatMapEntry
	// ZTDm is the zenith tropospheric delay in metres.
	ZTDm float64
	// TropoGradientNorth is the north tropospheric gradient.
	TropoGradientNorth float64
	// TropoGradientEast is the east tropospheric gradient.
	TropoGradientEast float64
}

// PPPMeasurementWeights contains inverse measurement sigmas.
type PPPMeasurementWeights struct {
	// Code is the observable code.
	Code float64
	// Phase is the phase observable.
	Phase float64
	// ElevationWeighting enables native elevation-dependent measurement weighting.
	ElevationWeighting bool
}

// PPPVmfSiteSample is one VMF1 site coefficient sample.
type PPPVmfSiteSample struct {
	// MJD is the mjd in modified Julian days.
	MJD float64
	// AH is the VMF1 hydrostatic mapping coefficient.
	AH float64
	// AW is the VMF1 wet mapping coefficient.
	AW float64
}

// PPPTroposphereOptions controls PPP troposphere corrections and states.
type PPPTroposphereOptions struct {
	// Enabled reports whether the option is enabled.
	Enabled bool
	// EstimateZTD enables estimation of the zenith total delay state.
	EstimateZTD bool
	// EstimateTropoGradient enables estimation of north/east tropospheric gradients.
	EstimateTropoGradient bool
	// PressureHPA is the atmospheric pressure in hectopascals.
	PressureHPA float64
	// TemperatureK is the temperature k in kelvin.
	TemperatureK float64
	// RelativeHumidity is the relative humidity fraction.
	RelativeHumidity float64
	// Mapping selects the Niell or VMF1 tropospheric mapping model.
	Mapping PPPTropoMapping
	// VMFSamples contains a detached copy; nil means this field is absent.
	VMFSamples []PPPVmfSiteSample
}

// PPPReceiverAntennaNoaziPCVSample is one receiver antenna no-azimuth PCV
// sample in metres.
type PPPReceiverAntennaNoaziPCVSample struct {
	// ZenithDeg is the zenith deg in degrees.
	ZenithDeg float64
	// ValueM is the value m in metres.
	ValueM float64
}

// PPPReceiverAntennaAzimuthPCVSample is one receiver antenna azimuth PCV
// sample in metres.
type PPPReceiverAntennaAzimuthPCVSample struct {
	// AzimuthDeg is the azimuth deg in degrees.
	AzimuthDeg float64
	// ZenithDeg is the zenith deg in degrees.
	ZenithDeg float64
	// ValueM is the value m in metres.
	ValueM float64
}

// PPPReceiverAntennaCalibration is one receiver antenna frequency calibration.
type PPPReceiverAntennaCalibration struct {
	// PCONEUM is the pconeum in metres.
	PCONEUM [3]float64
	// NoaziPCVM is the noazi pcvm in metres.
	NoaziPCVM []PPPReceiverAntennaNoaziPCVSample
	// AzimuthPCVM is the azimuth pcvm in metres.
	AzimuthPCVM []PPPReceiverAntennaAzimuthPCVSample
}

// PPPReceiverAntennaOptions describes the two-frequency receiver antenna
// correction.
type PPPReceiverAntennaOptions struct {
	// Frequency1Label is the receiver antenna calibration label for the first frequency.
	Frequency1Label string
	// Frequency1Hz is the frequency1 hz in hertz.
	Frequency1Hz float64
	// Frequency1 is the first carrier frequency.
	Frequency1 PPPReceiverAntennaCalibration
	// Frequency2Label is the receiver antenna calibration label for the second frequency.
	Frequency2Label string
	// Frequency2Hz is the frequency2 hz in hertz.
	Frequency2Hz float64
	// Frequency2 is the second carrier frequency.
	Frequency2 PPPReceiverAntennaCalibration
}

// PPPSatelliteClockRecord is one precise satellite clock sample in seconds.
type PPPSatelliteClockRecord struct {
	// SatelliteID is the GNSS satellite identifier for the clock record.
	SatelliteID string
	// GPSSeconds is the gps seconds in seconds.
	GPSSeconds float64
	// ClockS is the clock s in seconds.
	ClockS float64
}

// PPPRangeCorrections controls directly representable range corrections.
type PPPRangeCorrections struct {
	// ReceiverAntenna contains a detached copy; nil means this field is absent.
	ReceiverAntenna *PPPReceiverAntennaOptions
	// SatClockRelativity enables the satellite-clock relativistic correction.
	SatClockRelativity bool
	// SatelliteClockRecords contains a detached copy; nil means this field is absent.
	SatelliteClockRecords []PPPSatelliteClockRecord
	// SolidEarthTide enables solid-Earth tide correction.
	SolidEarthTide bool
	// PhaseWindup enables carrier phase-windup correction.
	PhaseWindup bool
	// SatelliteAntenna enables satellite antenna phase-center correction.
	SatelliteAntenna bool
}

// PPPFloatOptions contains float-solve iteration and convergence controls.
type PPPFloatOptions struct {
	// MaxIterations is the maximum number of nonlinear float-solver iterations.
	MaxIterations int
	// PositionToleranceM is the position tolerance m in metres.
	PositionToleranceM float64
	// ClockToleranceM is the clock tolerance m in metres.
	ClockToleranceM float64
	// AmbiguityToleranceM is the ambiguity tolerance m in metres.
	AmbiguityToleranceM float64
	// ZTDToleranceM is the ztd tolerance m in metres.
	ZTDToleranceM float64
}

// PPPFloatConfig is the complete static PPP float input bundle.
type PPPFloatConfig struct {
	// Epochs contains a detached copy; nil means this field is absent.
	Epochs []PPPEpoch
	// InitialState supplies the ECEF position, clock, ambiguity, and troposphere seed states.
	InitialState PPPFloatState
	// Weights is the observation weights.
	Weights PPPMeasurementWeights
	// Troposphere is the troposphere model.
	Troposphere PPPTroposphereOptions
	// Corrections is the applied corrections.
	Corrections PPPRangeCorrections
	// Options is the processing options for this record.
	Options PPPFloatOptions
	// HasElevationCutoff reports whether ElevationCutoffDeg is enforced.
	HasElevationCutoff bool
	// ElevationCutoffDeg is the elevation cutoff deg in degrees.
	ElevationCutoffDeg float64
	// ResidualScreen enables native residual screening during the float solve.
	ResidualScreen bool
}

// PPPFixedAmbiguityOptions controls integer ambiguity fixing.
type PPPFixedAmbiguityOptions struct {
	// WavelengthsM is the wavelengths m in metres.
	WavelengthsM []PPPFloatMapEntry
	// OffsetsM is the offsets m in metres.
	OffsetsM []PPPFloatMapEntry
	// RatioThreshold is the integer-ambiguity acceptance ratio threshold.
	RatioThreshold float64
}

// PPPFixedConfig is the complete static PPP fixed input bundle.
type PPPFixedConfig struct {
	// Epochs contains a detached copy; nil means this field is absent.
	Epochs []PPPEpoch
	// Weights is the observation weights.
	Weights PPPMeasurementWeights
	// Troposphere is the troposphere model.
	Troposphere PPPTroposphereOptions
	// Corrections is the applied corrections.
	Corrections PPPRangeCorrections
	// Options is the processing options for this record.
	Options PPPFloatOptions
	// HasElevationCutoff reports whether ElevationCutoffDeg is enforced.
	HasElevationCutoff bool
	// ElevationCutoffDeg is the elevation cutoff deg in degrees.
	ElevationCutoffDeg float64
	// Ambiguity contains wavelength, offset, and ratio settings for integer fixing.
	Ambiguity PPPFixedAmbiguityOptions
}

// PPPPoleTideOptions contains IERS polar motion in arcseconds.
type PPPPoleTideOptions struct {
	// XPArcsec is the xp arcsec in arcseconds.
	XPArcsec float64
	// YPArcsec is the yp arcsec in arcseconds.
	YPArcsec float64
}

// PPPObservationCorrection is one correction-precompute satellite row.
type PPPObservationCorrection struct {
	// SatelliteID is the GNSS satellite identifier for the correction row.
	SatelliteID string
	// Frequency1Hz is the frequency1 hz in hertz.
	Frequency1Hz float64
	// Frequency2Hz is the frequency2 hz in hertz.
	Frequency2Hz float64
	// HasGLONASSChannel reports whether GLONASSChannel supplies the FDMA channel for this observation.
	HasGLONASSChannel bool
	// GLONASSChannel is the FDMA frequency channel used when HasGLONASSChannel is true.
	GLONASSChannel int8
}

// PPPCorrectionEpoch is one correction-precompute epoch.
type PPPCorrectionEpoch struct {
	// Epoch is the UTC civil epoch associated with the correction rows.
	Epoch CivilDateTime
	// TRxJ2000S is the receiver epoch in seconds from J2000.
	TRxJ2000S float64
	// Observations contains a detached copy; nil means this field is absent.
	Observations []PPPObservationCorrection
}

// PPPCodeBiasSystemPair selects default observables for a GNSS system.
type PPPCodeBiasSystemPair struct {
	// System is the GNSS system identifier.
	System GNSSSystem
	// Obs1 is the first code-observable label for the system.
	Obs1 string
	// Obs2 is the second code-observable label for the system.
	Obs2 string
}

// PPPCodeBiasSatellitePair overrides observables for one satellite.
type PPPCodeBiasSatellitePair struct {
	// SatelliteID is the GNSS satellite whose code-bias labels are overridden.
	SatelliteID string
	// Obs1 is the first code-observable label for the satellite.
	Obs1 string
	// Obs2 is the second code-observable label for the satellite.
	Obs2 string
}

// PPPSatelliteAntennaFrequency contains one satellite antenna calibration.
type PPPSatelliteAntennaFrequency struct {
	// Label is the calibrated carrier label for this antenna-frequency block.
	Label string
	// PCOM is the pcom in metres.
	PCOM [3]float64
	// NoaziPCV contains a detached copy; nil means this field is absent.
	NoaziPCV []PPPNoaziPCVSample
}

// PPPNoaziPCVSample is one satellite antenna PCV pair.
type PPPNoaziPCVSample struct {
	// A is the no-azimuth PCV polynomial coefficient for the sample.
	A float64
	// B is the no-azimuth PCV polynomial coefficient paired with A.
	B float64
}

// PPPSatelliteAntenna contains one satellite antenna validity block.
type PPPSatelliteAntenna struct {
	// SatelliteID is the GNSS satellite covered by the validity block.
	SatelliteID string
	// HasValidFrom reports whether ValidFrom bounds the start of antenna calibration validity.
	HasValidFrom bool
	// ValidFrom is the UTC start of antenna calibration validity when HasValidFrom is true.
	ValidFrom CivilDateTime
	// HasValidUntil reports whether ValidUntil bounds the end of antenna calibration validity.
	HasValidUntil bool
	// ValidUntil is the UTC end of antenna calibration validity when HasValidUntil is true.
	ValidUntil CivilDateTime
	// Frequencies contains a detached copy; nil means this field is absent.
	Frequencies []PPPSatelliteAntennaFrequency
}

// PPPSatelliteAntennaOptions selects satellite antenna calibrations.
type PPPSatelliteAntennaOptions struct {
	// Frequency1Label is the calibrated satellite-antenna label for the first frequency.
	Frequency1Label string
	// Frequency1Hz is the frequency1 hz in hertz.
	Frequency1Hz float64
	// Frequency2Label is the calibrated satellite-antenna label for the second frequency.
	Frequency2Label string
	// Frequency2Hz is the frequency2 hz in hertz.
	Frequency2Hz float64
	// Antennas contains a detached copy; nil means this field is absent.
	Antennas []PPPSatelliteAntenna
}

// PPPCorrectionsOptions controls PPP correction-table construction.
type PPPCorrectionsOptions struct {
	// SolidEarthTide enables solid-Earth tide correction-table construction.
	SolidEarthTide bool
	// PoleTide contains a detached copy; nil means this field is absent.
	PoleTide *PPPPoleTideOptions
	// OceanLoading contains a detached copy; nil means this field is absent.
	OceanLoading *OceanLoadingBLQ
	// PhaseWindup enables phase-windup correction-table construction.
	PhaseWindup bool
	// SatelliteAntenna contains a detached copy; nil means this field is absent.
	SatelliteAntenna *PPPSatelliteAntennaOptions
	// CodeBias contains a detached copy; nil means this field is absent.
	CodeBias *BiasSet
	// CodeBiasSystemPairs contains a detached copy; nil means this field is absent.
	CodeBiasSystemPairs []PPPCodeBiasSystemPair
	// CodeBiasSatellitePairs contains a detached copy; nil means this field is absent.
	CodeBiasSatellitePairs []PPPCodeBiasSatellitePair
	// CodeBiasClockReference contains a detached copy; nil means this field is absent.
	CodeBiasClockReference []PPPCodeBiasSystemPair
	// HasCodeBiasClockReference reports whether CodeBiasClockReference selects a clock-reference observable pair.
	HasCodeBiasClockReference bool
}

// PPPSatScalarCorrection is one detached per-satellite scalar correction in
// metres.
type PPPSatScalarCorrection struct {
	// SatelliteID is the GNSS satellite identifier for the scalar correction.
	SatelliteID string
	// EpochIndex is the zero-based epoch index owning the correction.
	EpochIndex int
	// ValueM is the value m in metres.
	ValueM float64
}

// PPPSatVectorCorrection is one detached per-satellite vector correction in
// metres.
type PPPSatVectorCorrection struct {
	// SatelliteID is the GNSS satellite identifier for the vector correction.
	SatelliteID string
	// EpochIndex is the zero-based epoch index owning the correction.
	EpochIndex int
	// ValueM is the value m in metres.
	ValueM [3]float64
}

// PPPEpochVectorCorrection is one detached epoch vector correction in metres.
type PPPEpochVectorCorrection struct {
	// EpochIndex is the zero-based epoch index owning the correction.
	EpochIndex int
	// ValueM is the value m in metres.
	ValueM [3]float64
}

// PPPFloatAmbiguity is one detached float ambiguity estimate in metres.
type PPPFloatAmbiguity struct {
	// ID identifies the float ambiguity represented by ValueM.
	ID string
	// ValueM is the value m in metres.
	ValueM float64
}

// PPPFixedAmbiguity is one detached fixed ambiguity estimate.
type PPPFixedAmbiguity struct {
	// ID identifies the fixed ambiguity represented by Cycles and ValueM.
	ID string
	// Cycles is the integer ambiguity estimate in carrier cycles.
	Cycles int64
	// ValueM is the value m in metres.
	ValueM float64
}

// PPPPositionCovariance contains ECEF and ENU covariance matrices in square
// metres.
type PPPPositionCovariance struct {
	// ECEFM2 is the ecefm2 in square metres.
	ECEFM2 [3][3]float64
	// ENUM2 is the enum2 in square metres.
	ENUM2 [3][3]float64
}

// PPPPositionCovariances contains the three native PPP covariance products.
type PPPPositionCovariances struct {
	// Posterior is the posterior ECEF/ENU position covariance.
	Posterior PPPPositionCovariance
	// Formal is the formal ECEF/ENU position covariance from the measurement model.
	Formal PPPPositionCovariance
	// Temporal is the covariance product including temporal-correlation effects.
	Temporal PPPPositionCovariance
	// PosteriorScale is the dimensionless posterior covariance scale factor.
	PosteriorScale float64
	// TemporalScale is the dimensionless temporal-correlation scale factor.
	TemporalScale float64
}

// PPPTemporalCorrelation contains residual autocorrelation diagnostics.
type PPPTemporalCorrelation struct {
	// HasLag1 reports whether Lag1 is a valid residual autocorrelation estimate.
	HasLag1 bool
	// Lag1 is the lag-one residual autocorrelation when HasLag1 is true.
	Lag1 float64
	// HasDecorrelation reports whether DecorrelationTimeS is valid.
	HasDecorrelation bool
	// DecorrelationTimeS is the decorrelation time s in seconds.
	DecorrelationTimeS float64
	// NominalSamples is the number of samples used by the nominal correlation estimate.
	NominalSamples int
	// EffectiveSamples is the correlation-adjusted effective sample count.
	EffectiveSamples float64
	// VarianceInflation is the dimensionless variance inflation factor.
	VarianceInflation float64
	// ArcsUsed is the number of independent arcs contributing to the correlation estimate.
	ArcsUsed int
}

// PPPTropoGradient contains detached north/east gradient estimates and
// covariance products.
type PPPTropoGradient struct {
	// HasGradient reports whether NorthM and EastM contain estimated gradients.
	HasGradient bool
	// NorthM is the north m in metres.
	NorthM float64
	// EastM is the east m in metres.
	EastM float64
	// HasCovariance reports whether CovarianceM2 contains a valid gradient covariance.
	HasCovariance bool
	// CovarianceM2 is the north/east gradient covariance in square metres.
	CovarianceM2 [2][2]float64
	// HasFormalCovariance reports whether FormalCovarianceM2 contains a valid formal covariance.
	HasFormalCovariance bool
	// FormalCovarianceM2 is the formal north/east gradient covariance in square metres.
	FormalCovarianceM2 [2][2]float64
}

// PPPFloatMetadata contains detached float-solver summary values.
type PPPFloatMetadata struct {
	// Iterations is the native solver iteration count.
	Iterations int
	// Converged reports whether the solution converged.
	Converged bool
	// Status is the native status code.
	Status PPPSolveStatus
	// HasZTDResidual reports whether ZTDResidualM contains a valid zenith-delay residual.
	HasZTDResidual bool
	// ZTDResidualM is the ztd residual m in metres.
	ZTDResidualM float64
	// CodeRMSM is the code residual RMS in metres.
	CodeRMSM float64
	// PhaseRMSM is the carrier-phase residual RMS in metres.
	PhaseRMSM float64
	// WeightedRMSM is the weighted residual RMS in metres.
	WeightedRMSM float64
	// AmbiguityCount is the number of ambiguity states in the float solution.
	AmbiguityCount int
	// ResidualCount is the number of residual observations in the float solve.
	ResidualCount int
	// UsedSatCount is the number of satellites contributing to the float solution.
	UsedSatCount int
}

// PPPFixedMetadata contains detached fixed-solver and integer-search summary
// values.
type PPPFixedMetadata struct {
	// Iterations is the native solver iteration count.
	Iterations int
	// Converged reports whether the solution converged.
	Converged bool
	// Status is the native status code.
	Status PPPSolveStatus
	// HasZTDResidual reports whether ZTDResidualM contains a valid zenith-delay residual.
	HasZTDResidual bool
	// ZTDResidualM is the ztd residual m in metres.
	ZTDResidualM float64
	// CodeRMSM is the code residual RMS in metres.
	CodeRMSM float64
	// PhaseRMSM is the carrier-phase residual RMS in metres.
	PhaseRMSM float64
	// WeightedRMSM is the weighted residual RMS in metres.
	WeightedRMSM float64
	// FixedAmbiguityCount is the number of ambiguities accepted as integers.
	FixedAmbiguityCount int
	// ResidualCount is the number of residual observations in the fixed solve.
	ResidualCount int
	// UsedSatCount is the number of satellites contributing to the fixed solution.
	UsedSatCount int
	// IntegerStatus is the integer-solution status.
	IntegerStatus PPPIntegerStatus
	// IntegerRatio is the integer ambiguity ratio.
	IntegerRatio float64
	// IntegerBestScore is the best integer-search objective score.
	IntegerBestScore float64
	// HasIntegerSecondBest reports whether IntegerSecondBestScore is valid.
	HasIntegerSecondBest bool
	// IntegerSecondBestScore is the second-best integer-search objective score when HasIntegerSecondBest is true.
	IntegerSecondBestScore float64
	// IntegerCandidates is the integer candidate count.
	IntegerCandidates int
}

// PPPFloatSolution owns a native static PPP float-solution handle.
type PPPFloatSolution struct {
	_      noCopy
	handle *native.PppFloatSolution
}

// PPPFixedSolution owns a native static PPP fixed-solution handle.
type PPPFixedSolution struct {
	_      noCopy
	handle *native.PppFixedSolution
}

// PPPCorrections owns native precomputed PPP correction tables.
type PPPCorrections struct {
	_      noCopy
	handle *native.PppCorrections
}

func nativePPPAutoInitOptions(value PPPAutoInitOptions) native.PppAutoInitOptions {
	return native.PppAutoInitOptions{HasInitialGuess: value.HasInitialGuess, InitialGuessPosition: value.InitialGuessPosition, InitialGuessClockM: value.InitialGuessClockM, SPPInitialGuess: value.SPPInitialGuess, SPPTroposphere: value.SPPTroposphere, SPPPressureHPA: value.SPPPressureHPA, SPPTemperatureK: value.SPPTemperatureK, SPPRelativeHumidity: value.SPPRelativeHumidity}
}

func nativePPPObservation(value PPPObservation) native.PppObservation {
	return native.PppObservation{SatelliteID: value.SatelliteID, AmbiguityID: value.AmbiguityID, CodeM: value.CodeM, PhaseM: value.PhaseM, Frequency1Hz: value.Frequency1Hz, Frequency2Hz: value.Frequency2Hz}
}

func nativePPPEpoch(value PPPEpoch) native.PppEpoch {
	out := native.PppEpoch{Civil: native.CivilDateTime{Year: value.Civil.Year, Month: value.Civil.Month, Day: value.Civil.Day, Hour: value.Civil.Hour, Minute: value.Civil.Minute, Second: value.Civil.Second}, JDWhole: value.JDWhole, JDFraction: value.JDFraction, TRxJ2000S: value.TRxJ2000S, Observations: make([]native.PppObservation, len(value.Observations))}
	for i, observation := range value.Observations {
		out.Observations[i] = nativePPPObservation(observation)
	}
	return out
}

func nativePPPMapEntries(values []PPPFloatMapEntry) []native.PppFloatMapEntry {
	out := make([]native.PppFloatMapEntry, len(values))
	for i, value := range values {
		out[i] = native.PppFloatMapEntry{ID: value.ID, Value: value.Value}
	}
	return out
}

func nativePPPState(value PPPFloatState) native.PppFloatState {
	return native.PppFloatState{PositionM: value.PositionM, ClocksM: append([]float64(nil), value.ClocksM...), AmbiguitiesM: nativePPPMapEntries(value.AmbiguitiesM), ZTDm: value.ZTDm, TropoGradientNorth: value.TropoGradientNorth, TropoGradientEast: value.TropoGradientEast}
}

func nativePPPWeights(value PPPMeasurementWeights) native.PppMeasurementWeights {
	return native.PppMeasurementWeights{Code: value.Code, Phase: value.Phase, ElevationWeighting: value.ElevationWeighting}
}

func nativePPPTroposphere(value PPPTroposphereOptions) native.PppTroposphereOptions {
	out := native.PppTroposphereOptions{Enabled: value.Enabled, EstimateZTD: value.EstimateZTD, EstimateTropoGradient: value.EstimateTropoGradient, PressureHPA: value.PressureHPA, TemperatureK: value.TemperatureK, RelativeHumidity: value.RelativeHumidity, Mapping: uint32(value.Mapping), VMFSamples: make([]native.PppVmfSiteSample, len(value.VMFSamples))}
	for i, sample := range value.VMFSamples {
		out.VMFSamples[i] = native.PppVmfSiteSample{MJD: sample.MJD, AH: sample.AH, AW: sample.AW}
	}
	return out
}

func nativePPPReceiverAntennaCalibration(value PPPReceiverAntennaCalibration) native.PppReceiverAntennaCalibration {
	out := native.PppReceiverAntennaCalibration{PCONEUM: value.PCONEUM, NoaziPCVM: make([]native.PppReceiverAntennaNoaziPcvSample, len(value.NoaziPCVM)), AzimuthPCVM: make([]native.PppReceiverAntennaAzimuthPcvSample, len(value.AzimuthPCVM))}
	for i, sample := range value.NoaziPCVM {
		out.NoaziPCVM[i] = native.PppReceiverAntennaNoaziPcvSample{ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	for i, sample := range value.AzimuthPCVM {
		out.AzimuthPCVM[i] = native.PppReceiverAntennaAzimuthPcvSample{AzimuthDeg: sample.AzimuthDeg, ZenithDeg: sample.ZenithDeg, ValueM: sample.ValueM}
	}
	return out
}

func nativePPPRangeCorrections(value PPPRangeCorrections) native.PppRangeCorrections {
	out := native.PppRangeCorrections{SatClockRelativity: value.SatClockRelativity, SolidEarthTide: value.SolidEarthTide, PhaseWindup: value.PhaseWindup, SatelliteAntenna: value.SatelliteAntenna, SatelliteClockRecords: make([]native.PppSatelliteClockRecord, len(value.SatelliteClockRecords))}
	if value.ReceiverAntenna != nil {
		antenna := native.PppReceiverAntennaOptions{Frequency1Label: value.ReceiverAntenna.Frequency1Label, Frequency1Hz: value.ReceiverAntenna.Frequency1Hz, Frequency1: nativePPPReceiverAntennaCalibration(value.ReceiverAntenna.Frequency1), Frequency2Label: value.ReceiverAntenna.Frequency2Label, Frequency2Hz: value.ReceiverAntenna.Frequency2Hz, Frequency2: nativePPPReceiverAntennaCalibration(value.ReceiverAntenna.Frequency2)}
		out.ReceiverAntenna = &antenna
	}
	for i, record := range value.SatelliteClockRecords {
		out.SatelliteClockRecords[i] = native.PppSatelliteClockRecord{SatelliteID: record.SatelliteID, GPSSeconds: record.GPSSeconds, ClockS: record.ClockS}
	}
	return out
}

func nativePPPFloatOptions(value PPPFloatOptions) native.PppFloatOptions {
	return native.PppFloatOptions{MaxIterations: value.MaxIterations, PositionToleranceM: value.PositionToleranceM, ClockToleranceM: value.ClockToleranceM, AmbiguityToleranceM: value.AmbiguityToleranceM, ZTDToleranceM: value.ZTDToleranceM}
}

func nativePPPFloatConfig(value PPPFloatConfig) native.PppFloatConfig {
	out := native.PppFloatConfig{Epochs: make([]native.PppEpoch, len(value.Epochs)), InitialState: nativePPPState(value.InitialState), Weights: nativePPPWeights(value.Weights), Troposphere: nativePPPTroposphere(value.Troposphere), Corrections: nativePPPRangeCorrections(value.Corrections), Options: nativePPPFloatOptions(value.Options), HasElevationCutoff: value.HasElevationCutoff, ElevationCutoffDeg: value.ElevationCutoffDeg, ResidualScreen: value.ResidualScreen}
	for i, epoch := range value.Epochs {
		out.Epochs[i] = nativePPPEpoch(epoch)
	}
	return out
}

func nativePPPFixedConfig(value PPPFixedConfig) native.PppFixedConfig {
	out := native.PppFixedConfig{Epochs: make([]native.PppEpoch, len(value.Epochs)), Weights: nativePPPWeights(value.Weights), Troposphere: nativePPPTroposphere(value.Troposphere), Corrections: nativePPPRangeCorrections(value.Corrections), Options: nativePPPFloatOptions(value.Options), HasElevationCutoff: value.HasElevationCutoff, ElevationCutoffDeg: value.ElevationCutoffDeg, Ambiguity: native.PppFixedAmbiguityOptions{WavelengthsM: nativePPPMapEntries(value.Ambiguity.WavelengthsM), OffsetsM: nativePPPMapEntries(value.Ambiguity.OffsetsM), RatioThreshold: value.Ambiguity.RatioThreshold}}
	for i, epoch := range value.Epochs {
		out.Epochs[i] = nativePPPEpoch(epoch)
	}
	return out
}

func nativePPPCorrectionEpoch(value PPPCorrectionEpoch) native.PppCorrectionEpoch {
	out := native.PppCorrectionEpoch{Epoch: native.CivilDateTime{Year: value.Epoch.Year, Month: value.Epoch.Month, Day: value.Epoch.Day, Hour: value.Epoch.Hour, Minute: value.Epoch.Minute, Second: value.Epoch.Second}, TRxJ2000S: value.TRxJ2000S, Observations: make([]native.PppCorrectionObservation, len(value.Observations))}
	for i, observation := range value.Observations {
		out.Observations[i] = native.PppCorrectionObservation{SatelliteID: observation.SatelliteID, Frequency1Hz: observation.Frequency1Hz, Frequency2Hz: observation.Frequency2Hz, HasGLONASSChannel: observation.HasGLONASSChannel, GLONASSChannel: observation.GLONASSChannel}
	}
	return out
}

func nativePPPCodeBiasSystemPairs(values []PPPCodeBiasSystemPair) []native.PppCodeBiasSystemPair {
	out := make([]native.PppCodeBiasSystemPair, len(values))
	for i, value := range values {
		out[i] = native.PppCodeBiasSystemPair{System: uint32(value.System), Obs1: value.Obs1, Obs2: value.Obs2}
	}
	return out
}

func nativePPPCorrectionsOptions(value PPPCorrectionsOptions) native.PppCorrectionsOptions {
	out := native.PppCorrectionsOptions{SolidEarthTide: value.SolidEarthTide, PhaseWindup: value.PhaseWindup, CodeBiasSystemPairs: nativePPPCodeBiasSystemPairs(value.CodeBiasSystemPairs), CodeBiasSatellitePairs: make([]native.PppCodeBiasSatellitePair, len(value.CodeBiasSatellitePairs)), CodeBiasClockReference: nativePPPCodeBiasSystemPairs(value.CodeBiasClockReference), HasCodeBiasClockReference: value.HasCodeBiasClockReference}
	if value.PoleTide != nil {
		out.PoleTide = &native.PppPoleTideOptions{XPArcsec: value.PoleTide.XPArcsec, YPArcsec: value.PoleTide.YPArcsec}
	}
	if value.OceanLoading != nil {
		loading := native.OceanLoadingBLQ{AmplitudeM: value.OceanLoading.AmplitudeM, PhaseDeg: value.OceanLoading.PhaseDeg}
		out.OceanLoading = &loading
	}
	if value.CodeBias != nil {
		out.CodeBias = value.CodeBias.native
	}
	if value.SatelliteAntenna != nil {
		antenna := &native.PppSatelliteAntennaOptions{Frequency1Label: value.SatelliteAntenna.Frequency1Label, Frequency1Hz: value.SatelliteAntenna.Frequency1Hz, Frequency2Label: value.SatelliteAntenna.Frequency2Label, Frequency2Hz: value.SatelliteAntenna.Frequency2Hz, Antennas: make([]native.PppSatelliteAntenna, len(value.SatelliteAntenna.Antennas))}
		for i, item := range value.SatelliteAntenna.Antennas {
			frequencies := make([]native.PppSatelliteAntennaFrequency, len(item.Frequencies))
			for j, frequency := range item.Frequencies {
				samples := make([]native.PppNoaziPcvSample, len(frequency.NoaziPCV))
				for k, sample := range frequency.NoaziPCV {
					samples[k] = native.PppNoaziPcvSample{A: sample.A, B: sample.B}
				}
				frequencies[j] = native.PppSatelliteAntennaFrequency{Label: frequency.Label, PCOM: frequency.PCOM, NoaziPCV: samples}
			}
			antenna.Antennas[i] = native.PppSatelliteAntenna{SatelliteID: item.SatelliteID, HasValidFrom: item.HasValidFrom, ValidFrom: native.CivilDateTime{Year: item.ValidFrom.Year, Month: item.ValidFrom.Month, Day: item.ValidFrom.Day, Hour: item.ValidFrom.Hour, Minute: item.ValidFrom.Minute, Second: item.ValidFrom.Second}, HasValidUntil: item.HasValidUntil, ValidUntil: native.CivilDateTime{Year: item.ValidUntil.Year, Month: item.ValidUntil.Month, Day: item.ValidUntil.Day, Hour: item.ValidUntil.Hour, Minute: item.ValidUntil.Minute, Second: item.ValidUntil.Second}, Frequencies: frequencies}
		}
		out.SatelliteAntenna = antenna
	}
	for i, value := range value.CodeBiasSatellitePairs {
		out.CodeBiasSatellitePairs[i] = native.PppCodeBiasSatellitePair{SatelliteID: value.SatelliteID, Obs1: value.Obs1, Obs2: value.Obs2}
	}
	return out
}

// DefaultPPPAutoInitOptions returns the C engine's PPP auto-initialization
// defaults.
func DefaultPPPAutoInitOptions() (PPPAutoInitOptions, error) {
	value, err := native.PppAutoInitOptionsInit()
	return PPPAutoInitOptions{HasInitialGuess: value.HasInitialGuess, InitialGuessPosition: value.InitialGuessPosition, InitialGuessClockM: value.InitialGuessClockM, SPPInitialGuess: value.SPPInitialGuess, SPPTroposphere: value.SPPTroposphere, SPPPressureHPA: value.SPPPressureHPA, SPPTemperatureK: value.SPPTemperatureK, SPPRelativeHumidity: value.SPPRelativeHumidity}, publicError(err)
}

// DefaultPPPFloatOptions returns the C engine's PPP float-solver defaults.
func DefaultPPPFloatOptions() (PPPFloatOptions, error) {
	value, err := native.PppFloatOptionsInit()
	return PPPFloatOptions{MaxIterations: value.MaxIterations, PositionToleranceM: value.PositionToleranceM, ClockToleranceM: value.ClockToleranceM, AmbiguityToleranceM: value.AmbiguityToleranceM, ZTDToleranceM: value.ZTDToleranceM}, publicError(err)
}

// DefaultPPPFixAmbiguityOptions returns C's fixed-ambiguity defaults.
func DefaultPPPFixAmbiguityOptions() (PPPFixedAmbiguityOptions, error) {
	value, err := native.PppFixedAmbiguityOptionsInit()
	out := PPPFixedAmbiguityOptions{RatioThreshold: value.RatioThreshold, WavelengthsM: make([]PPPFloatMapEntry, len(value.WavelengthsM)), OffsetsM: make([]PPPFloatMapEntry, len(value.OffsetsM))}
	for i, item := range value.WavelengthsM {
		out.WavelengthsM[i] = PPPFloatMapEntry{ID: item.ID, Value: item.Value}
	}
	for i, item := range value.OffsetsM {
		out.OffsetsM[i] = PPPFloatMapEntry{ID: item.ID, Value: item.Value}
	}
	return out, publicError(err)
}

// DefaultPPPMeasurementWeights returns the C engine's measurement-weight
// defaults.
func DefaultPPPMeasurementWeights() (PPPMeasurementWeights, error) {
	value, err := native.PppMeasurementWeightsInit()
	return PPPMeasurementWeights{Code: value.Code, Phase: value.Phase, ElevationWeighting: value.ElevationWeighting}, publicError(err)
}

// DefaultPPPRangeCorrections returns the C engine's range-correction defaults.
func DefaultPPPRangeCorrections() (PPPRangeCorrections, error) {
	value, err := native.PppRangeCorrectionsInit()
	return PPPRangeCorrections{SatClockRelativity: value.SatClockRelativity, SolidEarthTide: value.SolidEarthTide, PhaseWindup: value.PhaseWindup, SatelliteAntenna: value.SatelliteAntenna}, publicError(err)
}

// DefaultPPPTroposphereOptions returns the C engine's troposphere defaults.
func DefaultPPPTroposphereOptions() (PPPTroposphereOptions, error) {
	value, err := native.PppTroposphereOptionsInit()
	out := PPPTroposphereOptions{Enabled: value.Enabled, EstimateZTD: value.EstimateZTD, EstimateTropoGradient: value.EstimateTropoGradient, PressureHPA: value.PressureHPA, TemperatureK: value.TemperatureK, RelativeHumidity: value.RelativeHumidity, Mapping: PPPTropoMapping(value.Mapping), VMFSamples: make([]PPPVmfSiteSample, len(value.VMFSamples))}
	for i, sample := range value.VMFSamples {
		out.VMFSamples[i] = PPPVmfSiteSample{MJD: sample.MJD, AH: sample.AH, AW: sample.AW}
	}
	return out, publicError(err)
}

// BuildPPPCorrections computes C-owned PPP correction tables from an SP3 arc.
func BuildPPPCorrections(sp3 *SP3, epochs []PPPCorrectionEpoch, receiverECEFM [3]float64, options PPPCorrectionsOptions) (*PPPCorrections, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	for _, pair := range options.CodeBiasSystemPairs {
		if err := validateGNSSSystem(pair.System); err != nil {
			return nil, err
		}
	}
	for _, pair := range options.CodeBiasClockReference {
		if err := validateGNSSSystem(pair.System); err != nil {
			return nil, err
		}
	}
	nativeEpochs := make([]native.PppCorrectionEpoch, len(epochs))
	for i, epoch := range epochs {
		nativeEpochs[i] = nativePPPCorrectionEpoch(epoch)
	}
	value, err := native.PppCorrectionsBuild(sp3.handle, nativeEpochs, receiverECEFM, nativePPPCorrectionsOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &PPPCorrections{handle: value}, nil
}

// SolvePPPFloat solves a static multi-epoch float PPP arc through C.
func SolvePPPFloat(sp3 *SP3, config PPPFloatConfig) (*PPPFloatSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	value, err := native.SolvePppFloat(sp3.handle, nativePPPFloatConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &PPPFloatSolution{handle: value}, nil
}

// SolvePPPAutoInitFloat solves a static float PPP arc using C's SPP auto-seed.
func SolvePPPAutoInitFloat(sp3 *SP3, config PPPFloatConfig, options PPPAutoInitOptions) (*PPPFloatSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	value, err := native.SolvePppAutoInitFloat(sp3.handle, nativePPPFloatConfig(config), nativePPPAutoInitOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &PPPFloatSolution{handle: value}, nil
}

// SolvePPPAutoInitFixed solves and integer-fixes a PPP arc using C's SPP
// auto-seed and LAMBDA route.
func SolvePPPAutoInitFixed(sp3 *SP3, floatConfig PPPFloatConfig, fixedConfig PPPFixedConfig, options PPPAutoInitOptions) (*PPPFixedSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	value, err := native.SolvePppAutoInitFixed(sp3.handle, nativePPPFloatConfig(floatConfig), nativePPPFixedConfig(fixedConfig), nativePPPAutoInitOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &PPPFixedSolution{handle: value}, nil
}

// SolvePPPFixed fixes ambiguities from a previously computed C PPP float
// solution.
func SolvePPPFixed(sp3 *SP3, floatSolution *PPPFloatSolution, config PPPFixedConfig) (*PPPFixedSolution, error) {
	if sp3 == nil || sp3.handle == nil || floatSolution == nil || floatSolution.handle == nil {
		return nil, ErrClosed
	}
	value, err := native.SolvePppFixed(sp3.handle, floatSolution.handle, nativePPPFixedConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &PPPFixedSolution{handle: value}, nil
}

// Close releases the PPP correction tables and is safe to call repeatedly.
func (c *PPPCorrections) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return publicError(c.handle.Close())
}

// Close releases the PPP float solution and is safe to call repeatedly.
func (s *PPPFloatSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// Close releases the PPP fixed solution and is safe to call repeatedly.
func (s *PPPFixedSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}

// CodeBias returns detached per-satellite code-bias corrections in metres.
func (c *PPPCorrections) CodeBias() ([]PPPSatScalarCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.CodeBias()
	out := make([]PPPSatScalarCorrection, len(values))
	for i, value := range values {
		out[i] = PPPSatScalarCorrection{SatelliteID: value.SatelliteID, EpochIndex: value.EpochIndex, ValueM: value.Value}
	}
	return out, publicError(err)
}

// OceanLoading returns detached epoch-indexed ocean-loading vectors in metres.
func (c *PPPCorrections) OceanLoading() ([]PPPEpochVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.OceanLoading()
	return nativePPPEpochVectorCorrections(values), publicError(err)
}

// PoleTide returns detached epoch-indexed pole-tide vectors in metres.
func (c *PPPCorrections) PoleTide() ([]PPPEpochVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.PoleTide()
	return nativePPPEpochVectorCorrections(values), publicError(err)
}

// SatPCOECEF returns detached satellite phase-center offsets in ECEF metres.
func (c *PPPCorrections) SatPCOECEF() ([]PPPSatVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.SatPCOECEF()
	out := make([]PPPSatVectorCorrection, len(values))
	for i, value := range values {
		out[i] = PPPSatVectorCorrection{SatelliteID: value.SatelliteID, EpochIndex: value.EpochIndex, ValueM: value.ValueM}
	}
	return out, publicError(err)
}

// SatPCV returns detached satellite phase-center variations in metres.
func (c *PPPCorrections) SatPCV() ([]PPPSatScalarCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.SatPCV()
	out := make([]PPPSatScalarCorrection, len(values))
	for i, value := range values {
		out[i] = PPPSatScalarCorrection{SatelliteID: value.SatelliteID, EpochIndex: value.EpochIndex, ValueM: value.Value}
	}
	return out, publicError(err)
}

// Tide returns detached epoch-indexed solid-Earth tide vectors in metres.
func (c *PPPCorrections) Tide() ([]PPPEpochVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.Tide()
	return nativePPPEpochVectorCorrections(values), publicError(err)
}

// Windup returns detached satellite phase-windup corrections in metres.
func (c *PPPCorrections) Windup() ([]PPPSatScalarCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.Windup()
	out := make([]PPPSatScalarCorrection, len(values))
	for i, value := range values {
		out[i] = PPPSatScalarCorrection{SatelliteID: value.SatelliteID, EpochIndex: value.EpochIndex, ValueM: value.Value}
	}
	return out, publicError(err)
}

func nativePPPEpochVectorCorrections(values []native.PppEpochVectorCorrection) []PPPEpochVectorCorrection {
	out := make([]PPPEpochVectorCorrection, len(values))
	for i, value := range values {
		out[i] = PPPEpochVectorCorrection{EpochIndex: value.EpochIndex, ValueM: value.ValueM}
	}
	return out
}

// FixedAmbiguities returns detached integer ambiguity estimates.
func (s *PPPFixedSolution) FixedAmbiguities() ([]PPPFixedAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.FixedAmbiguities()
	out := make([]PPPFixedAmbiguity, len(values))
	for i, value := range values {
		out[i] = PPPFixedAmbiguity{ID: value.ID, Cycles: value.Cycles, ValueM: value.ValueM}
	}
	return out, publicError(err)
}

// Ambiguities returns detached float ambiguity estimates in metres.
func (s *PPPFloatSolution) Ambiguities() ([]PPPFloatAmbiguity, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.Ambiguities()
	out := make([]PPPFloatAmbiguity, len(values))
	for i, value := range values {
		out[i] = PPPFloatAmbiguity{ID: value.ID, ValueM: value.ValueM}
	}
	return out, publicError(err)
}

// UsedIDs returns detached full-width ambiguity identifiers from a fixed solve.
func (s *PPPFixedSolution) UsedIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedIDs()
	return values, publicError(err)
}

// UsedIDs returns detached full-width ambiguity identifiers from a float solve.
func (s *PPPFloatSolution) UsedIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedIDs()
	return values, publicError(err)
}

// UsedSatelliteIDs returns detached legacy satellite identifiers from a fixed
// solve. Use UsedIDs when ambiguity identifiers may contain split-arc suffixes.
func (s *PPPFixedSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedSatelliteIDs()
	return values, publicError(err)
}

// UsedSatelliteIDs returns detached legacy satellite identifiers from a float
// solve. Use UsedIDs when ambiguity identifiers may contain split-arc suffixes.
func (s *PPPFloatSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedSatelliteIDs()
	return values, publicError(err)
}

// Position returns the detached float-solution ECEF position in metres.
func (s *PPPFloatSolution) Position() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.Position()
	return value, publicError(err)
}

// Position returns the detached fixed-solution ECEF position in metres.
func (s *PPPFixedSolution) Position() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.Position()
	return value, publicError(err)
}

// FloatPosition returns the fixed solve's unconstrained float ECEF position in
// metres.
func (s *PPPFixedSolution) FloatPosition() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.FloatPosition()
	return value, publicError(err)
}

// Metadata returns detached float-solver summary values.
func (s *PPPFloatSolution) Metadata() (PPPFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return PPPFloatMetadata{}, ErrClosed
	}
	value, err := s.handle.Metadata()
	return PPPFloatMetadata{Iterations: value.Iterations, Converged: value.Converged, Status: PPPSolveStatus(value.Status), HasZTDResidual: value.HasZTDResidual, ZTDResidualM: value.ZTDResidualM, CodeRMSM: value.CodeRMSM, PhaseRMSM: value.PhaseRMSM, WeightedRMSM: value.WeightedRMSM, AmbiguityCount: value.AmbiguityCount, ResidualCount: value.ResidualCount, UsedSatCount: value.UsedSatCount}, publicError(err)
}

// Metadata returns detached fixed-solver and integer-search summary values.
func (s *PPPFixedSolution) Metadata() (PPPFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return PPPFixedMetadata{}, ErrClosed
	}
	value, err := s.handle.Metadata()
	return PPPFixedMetadata{Iterations: value.Iterations, Converged: value.Converged, Status: PPPSolveStatus(value.Status), HasZTDResidual: value.HasZTDResidual, ZTDResidualM: value.ZTDResidualM, CodeRMSM: value.CodeRMSM, PhaseRMSM: value.PhaseRMSM, WeightedRMSM: value.WeightedRMSM, FixedAmbiguityCount: value.FixedAmbiguityCount, ResidualCount: value.ResidualCount, UsedSatCount: value.UsedSatCount, IntegerStatus: PPPIntegerStatus(value.IntegerStatus), IntegerRatio: value.IntegerRatio, IntegerBestScore: value.IntegerBestScore, HasIntegerSecondBest: value.HasIntegerSecondBest, IntegerSecondBestScore: value.IntegerSecondBestScore, IntegerCandidates: value.IntegerCandidates}, publicError(err)
}

// PositionCovariances returns detached ECEF/ENU covariance products in square
// metres.
func (s *PPPFloatSolution) PositionCovariances() (PPPPositionCovariances, error) {
	if s == nil || s.handle == nil {
		return PPPPositionCovariances{}, ErrClosed
	}
	value, err := s.handle.PositionCovariances()
	return nativePPPPositionCovariances(value), publicError(err)
}

// PositionCovariances returns detached ECEF/ENU covariance products in square
// metres.
func (s *PPPFixedSolution) PositionCovariances() (PPPPositionCovariances, error) {
	if s == nil || s.handle == nil {
		return PPPPositionCovariances{}, ErrClosed
	}
	value, err := s.handle.PositionCovariances()
	return nativePPPPositionCovariances(value), publicError(err)
}

func nativePPPPositionCovariances(value native.PppPositionCovariances) PPPPositionCovariances {
	return PPPPositionCovariances{Posterior: nativePPPPositionCovariance(value.Posterior), Formal: nativePPPPositionCovariance(value.Formal), Temporal: nativePPPPositionCovariance(value.Temporal), PosteriorScale: value.PosteriorScale, TemporalScale: value.TemporalScale}
}

func nativePPPPositionCovariance(value native.PppPositionCovariance) PPPPositionCovariance {
	return PPPPositionCovariance{ECEFM2: value.ECEFM2, ENUM2: value.ENUM2}
}

// TemporalCorrelation returns detached residual-correlation diagnostics.
func (s *PPPFloatSolution) TemporalCorrelation() (PPPTemporalCorrelation, error) {
	if s == nil || s.handle == nil {
		return PPPTemporalCorrelation{}, ErrClosed
	}
	value, err := s.handle.TemporalCorrelation()
	return nativePPPTemporalCorrelation(value), publicError(err)
}

// TemporalCorrelation returns detached residual-correlation diagnostics.
func (s *PPPFixedSolution) TemporalCorrelation() (PPPTemporalCorrelation, error) {
	if s == nil || s.handle == nil {
		return PPPTemporalCorrelation{}, ErrClosed
	}
	value, err := s.handle.TemporalCorrelation()
	return nativePPPTemporalCorrelation(value), publicError(err)
}

func nativePPPTemporalCorrelation(value native.PppTemporalCorrelation) PPPTemporalCorrelation {
	return PPPTemporalCorrelation{HasLag1: value.HasLag1, Lag1: value.Lag1, HasDecorrelation: value.HasDecorrelation, DecorrelationTimeS: value.DecorrelationTimeS, NominalSamples: value.NominalSamples, EffectiveSamples: value.EffectiveSamples, VarianceInflation: value.VarianceInflation, ArcsUsed: value.ArcsUsed}
}

// TropoGradient returns detached north/east troposphere estimates and
// covariance products.
func (s *PPPFloatSolution) TropoGradient() (PPPTropoGradient, error) {
	if s == nil || s.handle == nil {
		return PPPTropoGradient{}, ErrClosed
	}
	value, err := s.handle.TropoGradient()
	return nativePPPTropoGradient(value), publicError(err)
}

// TropoGradient returns detached north/east troposphere estimates and
// covariance products.
func (s *PPPFixedSolution) TropoGradient() (PPPTropoGradient, error) {
	if s == nil || s.handle == nil {
		return PPPTropoGradient{}, ErrClosed
	}
	value, err := s.handle.TropoGradient()
	return nativePPPTropoGradient(value), publicError(err)
}

func nativePPPTropoGradient(value native.PppTropoGradient) PPPTropoGradient {
	return PPPTropoGradient{HasGradient: value.HasGradient, NorthM: value.NorthM, EastM: value.EastM, HasCovariance: value.HasCovariance, CovarianceM2: value.CovarianceM2, HasFormalCovariance: value.HasFormalCovariance, FormalCovarianceM2: value.FormalCovarianceM2}
}
