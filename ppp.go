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
	// HasInitialGuess reports whether the has initial guess field is present.
	HasInitialGuess bool
	// InitialGuessPosition contains the fixed-size array for this record.
	InitialGuessPosition [3]float64
	// InitialGuessClockM is the initial guess clock m in metres.
	InitialGuessClockM float64
	// SPPInitialGuess contains the fixed-size array for this record.
	SPPInitialGuess [4]float64
	// SPPTroposphere is the spptroposphere value for PPPAutoInitOptions.
	SPPTroposphere bool
	// SPPPressureHPA is the spp pressure hpa in hectopascals.
	SPPPressureHPA float64
	// SPPTemperatureK is the spp temperature k in kelvin.
	SPPTemperatureK float64
	// SPPRelativeHumidity is the spp relative humidity value for PPPAutoInitOptions.
	SPPRelativeHumidity float64
}

// PPPObservation is one ionosphere-free code/phase measurement. Distances are
// metres and frequencies are hertz.
type PPPObservation struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// AmbiguityID identifies or counts this record.
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
	// TRxJ2000S is the t rx j2000 s in seconds.
	TRxJ2000S float64
	// Observations contains a detached copy; nil means this field is absent.
	Observations []PPPObservation
}

// PPPFloatMapEntry is one ambiguity identifier/value pair.
type PPPFloatMapEntry struct {
	// ID identifies or counts this record.
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
	// ElevationWeighting is the elevation weighting value for PPPMeasurementWeights.
	ElevationWeighting bool
}

// PPPVmfSiteSample is one VMF1 site coefficient sample.
type PPPVmfSiteSample struct {
	// MJD is the mjd in modified Julian days.
	MJD float64
	// AH is the ah value for PPPVmfSiteSample.
	AH float64
	// AW is the aw value for PPPVmfSiteSample.
	AW float64
}

// PPPTroposphereOptions controls PPP troposphere corrections and states.
type PPPTroposphereOptions struct {
	// Enabled reports whether the option is enabled.
	Enabled bool
	// EstimateZTD is the estimate ztd value for PPPTroposphereOptions.
	EstimateZTD bool
	// EstimateTropoGradient is the estimate tropo gradient value for PPPTroposphereOptions.
	EstimateTropoGradient bool
	// PressureHPA is the pressure hpa in hectopascals.
	PressureHPA float64
	// TemperatureK is the temperature k in kelvin.
	TemperatureK float64
	// RelativeHumidity is the relative humidity fraction.
	RelativeHumidity float64
	// Mapping is the mapping value for PPPTroposphereOptions.
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
	// Frequency1Label is the frequency1 label value for PPPReceiverAntennaOptions.
	Frequency1Label string
	// Frequency1Hz is the frequency1 hz in hertz.
	Frequency1Hz float64
	// Frequency1 is the first carrier frequency.
	Frequency1 PPPReceiverAntennaCalibration
	// Frequency2Label is the frequency2 label value for PPPReceiverAntennaOptions.
	Frequency2Label string
	// Frequency2Hz is the frequency2 hz in hertz.
	Frequency2Hz float64
	// Frequency2 is the second carrier frequency.
	Frequency2 PPPReceiverAntennaCalibration
}

// PPPSatelliteClockRecord is one precise satellite clock sample in seconds.
type PPPSatelliteClockRecord struct {
	// SatelliteID identifies or counts this record.
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
	// SatClockRelativity is the sat clock relativity value for PPPRangeCorrections.
	SatClockRelativity bool
	// SatelliteClockRecords contains a detached copy; nil means this field is absent.
	SatelliteClockRecords []PPPSatelliteClockRecord
	// SolidEarthTide is the solid earth tide value for PPPRangeCorrections.
	SolidEarthTide bool
	// PhaseWindup is the phase windup value for PPPRangeCorrections.
	PhaseWindup bool
	// SatelliteAntenna is the satellite antenna value for PPPRangeCorrections.
	SatelliteAntenna bool
}

// PPPFloatOptions contains float-solve iteration and convergence controls.
type PPPFloatOptions struct {
	// MaxIterations is the max iterations value for PPPFloatOptions.
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
	// InitialState is the initial state value for PPPFloatConfig.
	InitialState PPPFloatState
	// Weights is the observation weights.
	Weights PPPMeasurementWeights
	// Troposphere is the troposphere model.
	Troposphere PPPTroposphereOptions
	// Corrections is the applied corrections.
	Corrections PPPRangeCorrections
	// Options is the processing options for this record.
	Options PPPFloatOptions
	// HasElevationCutoff reports whether the has elevation cutoff field is present.
	HasElevationCutoff bool
	// ElevationCutoffDeg is the elevation cutoff deg in degrees.
	ElevationCutoffDeg float64
	// ResidualScreen is the residual screen value for PPPFloatConfig.
	ResidualScreen bool
}

// PPPFixedAmbiguityOptions controls integer ambiguity fixing.
type PPPFixedAmbiguityOptions struct {
	// WavelengthsM is the wavelengths m in metres.
	WavelengthsM []PPPFloatMapEntry
	// OffsetsM is the offsets m in metres.
	OffsetsM []PPPFloatMapEntry
	// RatioThreshold is the ratio threshold value for PPPFixedAmbiguityOptions.
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
	// HasElevationCutoff reports whether the has elevation cutoff field is present.
	HasElevationCutoff bool
	// ElevationCutoffDeg is the elevation cutoff deg in degrees.
	ElevationCutoffDeg float64
	// Ambiguity is the ambiguity value for PPPFixedConfig.
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
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// Frequency1Hz is the frequency1 hz in hertz.
	Frequency1Hz float64
	// Frequency2Hz is the frequency2 hz in hertz.
	Frequency2Hz float64
	// HasGLONASSChannel reports whether the has glonass channel field is present.
	HasGLONASSChannel bool
	// GLONASSChannel is the glonasschannel value for PPPObservationCorrection.
	GLONASSChannel int8
}

// PPPCorrectionEpoch is one correction-precompute epoch.
type PPPCorrectionEpoch struct {
	// Epoch is the epoch value for PPPCorrectionEpoch.
	Epoch CivilDateTime
	// TRxJ2000S is the t rx j2000 s in seconds.
	TRxJ2000S float64
	// Observations contains a detached copy; nil means this field is absent.
	Observations []PPPObservationCorrection
}

// PPPCodeBiasSystemPair selects default observables for a GNSS system.
type PPPCodeBiasSystemPair struct {
	// System is the GNSS system identifier.
	System GNSSSystem
	// Obs1 is the obs1 value for PPPCodeBiasSystemPair.
	Obs1 string
	// Obs2 is the obs2 value for PPPCodeBiasSystemPair.
	Obs2 string
}

// PPPCodeBiasSatellitePair overrides observables for one satellite.
type PPPCodeBiasSatellitePair struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// Obs1 is the obs1 value for PPPCodeBiasSatellitePair.
	Obs1 string
	// Obs2 is the obs2 value for PPPCodeBiasSatellitePair.
	Obs2 string
}

// PPPSatelliteAntennaFrequency contains one satellite antenna calibration.
type PPPSatelliteAntennaFrequency struct {
	// Label is the label value for PPPSatelliteAntennaFrequency.
	Label string
	// PCOM is the pcom in metres.
	PCOM [3]float64
	// NoaziPCV contains a detached copy; nil means this field is absent.
	NoaziPCV []PPPNoaziPCVSample
}

// PPPNoaziPCVSample is one satellite antenna PCV pair.
type PPPNoaziPCVSample struct {
	// A is the a value for PPPNoaziPCVSample.
	A float64
	// B is the b value for PPPNoaziPCVSample.
	B float64
}

// PPPSatelliteAntenna contains one satellite antenna validity block.
type PPPSatelliteAntenna struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// HasValidFrom reports whether the has valid from field is present.
	HasValidFrom bool
	// ValidFrom is the valid from value for PPPSatelliteAntenna.
	ValidFrom CivilDateTime
	// HasValidUntil reports whether the has valid until field is present.
	HasValidUntil bool
	// ValidUntil is the valid until value for PPPSatelliteAntenna.
	ValidUntil CivilDateTime
	// Frequencies contains a detached copy; nil means this field is absent.
	Frequencies []PPPSatelliteAntennaFrequency
}

// PPPSatelliteAntennaOptions selects satellite antenna calibrations.
type PPPSatelliteAntennaOptions struct {
	// Frequency1Label is the frequency1 label value for PPPSatelliteAntennaOptions.
	Frequency1Label string
	// Frequency1Hz is the frequency1 hz in hertz.
	Frequency1Hz float64
	// Frequency2Label is the frequency2 label value for PPPSatelliteAntennaOptions.
	Frequency2Label string
	// Frequency2Hz is the frequency2 hz in hertz.
	Frequency2Hz float64
	// Antennas contains a detached copy; nil means this field is absent.
	Antennas []PPPSatelliteAntenna
}

// PPPCorrectionsOptions controls PPP correction-table construction.
type PPPCorrectionsOptions struct {
	// SolidEarthTide is the solid earth tide value for PPPCorrectionsOptions.
	SolidEarthTide bool
	// PoleTide contains a detached copy; nil means this field is absent.
	PoleTide *PPPPoleTideOptions
	// OceanLoading contains a detached copy; nil means this field is absent.
	OceanLoading *OceanLoadingBLQ
	// PhaseWindup is the phase windup value for PPPCorrectionsOptions.
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
	// HasCodeBiasClockReference reports whether the has code bias clock reference field is present.
	HasCodeBiasClockReference bool
}

// PPPSatScalarCorrection is one copied per-satellite scalar correction in
// metres.
type PPPSatScalarCorrection struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// EpochIndex identifies or counts this record.
	EpochIndex int
	// ValueM is the value m in metres.
	ValueM float64
}

// PPPSatVectorCorrection is one copied per-satellite vector correction in
// metres.
type PPPSatVectorCorrection struct {
	// SatelliteID identifies or counts this record.
	SatelliteID string
	// EpochIndex identifies or counts this record.
	EpochIndex int
	// ValueM is the value m in metres.
	ValueM [3]float64
}

// PPPEpochVectorCorrection is one copied epoch vector correction in metres.
type PPPEpochVectorCorrection struct {
	// EpochIndex identifies or counts this record.
	EpochIndex int
	// ValueM is the value m in metres.
	ValueM [3]float64
}

// PPPFloatAmbiguity is one copied float ambiguity estimate in metres.
type PPPFloatAmbiguity struct {
	// ID identifies or counts this record.
	ID string
	// ValueM is the value m in metres.
	ValueM float64
}

// PPPFixedAmbiguity is one copied fixed ambiguity estimate.
type PPPFixedAmbiguity struct {
	// ID identifies or counts this record.
	ID string
	// Cycles is the cycles value for PPPFixedAmbiguity.
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
	// Posterior is the posterior value for PPPPositionCovariances.
	Posterior PPPPositionCovariance
	// Formal is the formal value for PPPPositionCovariances.
	Formal PPPPositionCovariance
	// Temporal is the temporal value for PPPPositionCovariances.
	Temporal PPPPositionCovariance
	// PosteriorScale is the posterior scale value for PPPPositionCovariances.
	PosteriorScale float64
	// TemporalScale is the temporal scale value for PPPPositionCovariances.
	TemporalScale float64
}

// PPPTemporalCorrelation contains residual autocorrelation diagnostics.
type PPPTemporalCorrelation struct {
	// HasLag1 reports whether the has lag1 field is present.
	HasLag1 bool
	// Lag1 is the lag1 value for PPPTemporalCorrelation.
	Lag1 float64
	// HasDecorrelation reports whether the has decorrelation field is present.
	HasDecorrelation bool
	// DecorrelationTimeS is the decorrelation time s in seconds.
	DecorrelationTimeS float64
	// NominalSamples is the nominal samples value for PPPTemporalCorrelation.
	NominalSamples int
	// EffectiveSamples is the effective samples value for PPPTemporalCorrelation.
	EffectiveSamples float64
	// VarianceInflation is the variance inflation value for PPPTemporalCorrelation.
	VarianceInflation float64
	// ArcsUsed is the arcs used value for PPPTemporalCorrelation.
	ArcsUsed int
}

// PPPTropoGradient contains copied north/east gradient estimates and
// covariance products.
type PPPTropoGradient struct {
	// HasGradient reports whether the has gradient field is present.
	HasGradient bool
	// NorthM is the north m in metres.
	NorthM float64
	// EastM is the east m in metres.
	EastM float64
	// HasCovariance reports whether the has covariance field is present.
	HasCovariance bool
	// CovarianceM2 is the covariance m2 in square metres.
	CovarianceM2 [2][2]float64
	// HasFormalCovariance reports whether the has formal covariance field is present.
	HasFormalCovariance bool
	// FormalCovarianceM2 is the formal covariance m2 in square metres.
	FormalCovarianceM2 [2][2]float64
}

// PPPFloatMetadata contains copied float-solver summary values.
type PPPFloatMetadata struct {
	// Iterations is the native solver iteration count.
	Iterations int
	// Converged reports whether the solution converged.
	Converged bool
	// Status is the native status code.
	Status PPPSolveStatus
	// HasZTDResidual reports whether the has ztd residual field is present.
	HasZTDResidual bool
	// ZTDResidualM is the ztd residual m in metres.
	ZTDResidualM float64
	// CodeRMSM is the code rmsm in metres.
	CodeRMSM float64
	// PhaseRMSM is the phase rmsm in metres.
	PhaseRMSM float64
	// WeightedRMSM is the weighted rmsm in metres.
	WeightedRMSM float64
	// AmbiguityCount identifies or counts this record.
	AmbiguityCount int
	// ResidualCount identifies or counts this record.
	ResidualCount int
	// UsedSatCount identifies or counts this record.
	UsedSatCount int
}

// PPPFixedMetadata contains copied fixed-solver and integer-search summary
// values.
type PPPFixedMetadata struct {
	// Iterations is the native solver iteration count.
	Iterations int
	// Converged reports whether the solution converged.
	Converged bool
	// Status is the native status code.
	Status PPPSolveStatus
	// HasZTDResidual reports whether the has ztd residual field is present.
	HasZTDResidual bool
	// ZTDResidualM is the ztd residual m in metres.
	ZTDResidualM float64
	// CodeRMSM is the code rmsm in metres.
	CodeRMSM float64
	// PhaseRMSM is the phase rmsm in metres.
	PhaseRMSM float64
	// WeightedRMSM is the weighted rmsm in metres.
	WeightedRMSM float64
	// FixedAmbiguityCount identifies or counts this record.
	FixedAmbiguityCount int
	// ResidualCount identifies or counts this record.
	ResidualCount int
	// UsedSatCount identifies or counts this record.
	UsedSatCount int
	// IntegerStatus is the integer-solution status.
	IntegerStatus PPPIntegerStatus
	// IntegerRatio is the integer ambiguity ratio.
	IntegerRatio float64
	// IntegerBestScore is the integer best score value for PPPFixedMetadata.
	IntegerBestScore float64
	// HasIntegerSecondBest reports whether the has integer second best field is present.
	HasIntegerSecondBest bool
	// IntegerSecondBestScore is the integer second best score value for PPPFixedMetadata.
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

// CodeBias returns copied per-satellite code-bias corrections in metres.
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

// OceanLoading returns copied epoch-indexed ocean-loading vectors in metres.
func (c *PPPCorrections) OceanLoading() ([]PPPEpochVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.OceanLoading()
	return nativePPPEpochVectorCorrections(values), publicError(err)
}

// PoleTide returns copied epoch-indexed pole-tide vectors in metres.
func (c *PPPCorrections) PoleTide() ([]PPPEpochVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.PoleTide()
	return nativePPPEpochVectorCorrections(values), publicError(err)
}

// SatPCOECEF returns copied satellite phase-center offsets in ECEF metres.
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

// SatPCV returns copied satellite phase-center variations in metres.
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

// Tide returns copied epoch-indexed solid-Earth tide vectors in metres.
func (c *PPPCorrections) Tide() ([]PPPEpochVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	values, err := c.handle.Tide()
	return nativePPPEpochVectorCorrections(values), publicError(err)
}

// Windup returns copied satellite phase-windup corrections in metres.
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

// FixedAmbiguities returns copied integer ambiguity estimates.
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

// Ambiguities returns copied float ambiguity estimates in metres.
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

// UsedIDs returns copied full-width ambiguity identifiers from a fixed solve.
func (s *PPPFixedSolution) UsedIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedIDs()
	return values, publicError(err)
}

// UsedIDs returns copied full-width ambiguity identifiers from a float solve.
func (s *PPPFloatSolution) UsedIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedIDs()
	return values, publicError(err)
}

// UsedSatelliteIDs returns copied legacy satellite identifiers from a fixed
// solve. Use UsedIDs when ambiguity identifiers may contain split-arc suffixes.
func (s *PPPFixedSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedSatelliteIDs()
	return values, publicError(err)
}

// UsedSatelliteIDs returns copied legacy satellite identifiers from a float
// solve. Use UsedIDs when ambiguity identifiers may contain split-arc suffixes.
func (s *PPPFloatSolution) UsedSatelliteIDs() ([]string, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.UsedSatelliteIDs()
	return values, publicError(err)
}

// Position returns the copied float-solution ECEF position in metres.
func (s *PPPFloatSolution) Position() ([3]float64, error) {
	if s == nil || s.handle == nil {
		return [3]float64{}, ErrClosed
	}
	value, err := s.handle.Position()
	return value, publicError(err)
}

// Position returns the copied fixed-solution ECEF position in metres.
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

// Metadata returns copied float-solver summary values.
func (s *PPPFloatSolution) Metadata() (PPPFloatMetadata, error) {
	if s == nil || s.handle == nil {
		return PPPFloatMetadata{}, ErrClosed
	}
	value, err := s.handle.Metadata()
	return PPPFloatMetadata{Iterations: value.Iterations, Converged: value.Converged, Status: PPPSolveStatus(value.Status), HasZTDResidual: value.HasZTDResidual, ZTDResidualM: value.ZTDResidualM, CodeRMSM: value.CodeRMSM, PhaseRMSM: value.PhaseRMSM, WeightedRMSM: value.WeightedRMSM, AmbiguityCount: value.AmbiguityCount, ResidualCount: value.ResidualCount, UsedSatCount: value.UsedSatCount}, publicError(err)
}

// Metadata returns copied fixed-solver and integer-search summary values.
func (s *PPPFixedSolution) Metadata() (PPPFixedMetadata, error) {
	if s == nil || s.handle == nil {
		return PPPFixedMetadata{}, ErrClosed
	}
	value, err := s.handle.Metadata()
	return PPPFixedMetadata{Iterations: value.Iterations, Converged: value.Converged, Status: PPPSolveStatus(value.Status), HasZTDResidual: value.HasZTDResidual, ZTDResidualM: value.ZTDResidualM, CodeRMSM: value.CodeRMSM, PhaseRMSM: value.PhaseRMSM, WeightedRMSM: value.WeightedRMSM, FixedAmbiguityCount: value.FixedAmbiguityCount, ResidualCount: value.ResidualCount, UsedSatCount: value.UsedSatCount, IntegerStatus: PPPIntegerStatus(value.IntegerStatus), IntegerRatio: value.IntegerRatio, IntegerBestScore: value.IntegerBestScore, HasIntegerSecondBest: value.HasIntegerSecondBest, IntegerSecondBestScore: value.IntegerSecondBestScore, IntegerCandidates: value.IntegerCandidates}, publicError(err)
}

// PositionCovariances returns copied ECEF/ENU covariance products in square
// metres.
func (s *PPPFloatSolution) PositionCovariances() (PPPPositionCovariances, error) {
	if s == nil || s.handle == nil {
		return PPPPositionCovariances{}, ErrClosed
	}
	value, err := s.handle.PositionCovariances()
	return nativePPPPositionCovariances(value), publicError(err)
}

// PositionCovariances returns copied ECEF/ENU covariance products in square
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

// TemporalCorrelation returns copied residual-correlation diagnostics.
func (s *PPPFloatSolution) TemporalCorrelation() (PPPTemporalCorrelation, error) {
	if s == nil || s.handle == nil {
		return PPPTemporalCorrelation{}, ErrClosed
	}
	value, err := s.handle.TemporalCorrelation()
	return nativePPPTemporalCorrelation(value), publicError(err)
}

// TemporalCorrelation returns copied residual-correlation diagnostics.
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

// TropoGradient returns copied north/east troposphere estimates and
// covariance products.
func (s *PPPFloatSolution) TropoGradient() (PPPTropoGradient, error) {
	if s == nil || s.handle == nil {
		return PPPTropoGradient{}, ErrClosed
	}
	value, err := s.handle.TropoGradient()
	return nativePPPTropoGradient(value), publicError(err)
}

// TropoGradient returns copied north/east troposphere estimates and
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
