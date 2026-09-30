package sidereon

import (
	"runtime"

	"sidereon.dev/go/v3/internal/native"
)

// PPPTransmitTimeFailureKind identifies why a bias was unusable at transmission time.
type PPPTransmitTimeFailureKind uint32

const (
	// PPPTransmitTimeFailureNone means transmission-time handling added no failure.
	PPPTransmitTimeFailureNone PPPTransmitTimeFailureKind = 0
	// PPPTransmitTimeFailureSourceWithoutSSR means the source applies no SSR corrections.
	PPPTransmitTimeFailureSourceWithoutSSR PPPTransmitTimeFailureKind = 1
	// PPPTransmitTimeFailureUnavailable means transmission time could not be predicted.
	PPPTransmitTimeFailureUnavailable PPPTransmitTimeFailureKind = 2
	// PPPTransmitTimeFailureOrbitClockSolution means the source applied another solution.
	PPPTransmitTimeFailureOrbitClockSolution PPPTransmitTimeFailureKind = 3
	// PPPTransmitTimeFailureBiasRecord means the recorded bias was not valid at transmission time.
	PPPTransmitTimeFailureBiasRecord PPPTransmitTimeFailureKind = 4
	// PPPTransmitTimeFailureSource means the ephemeris source returned an error.
	PPPTransmitTimeFailureSource PPPTransmitTimeFailureKind = 5
	// PPPTransmitTimeFailureUnknown is the engine's explicit unknown tag.
	PPPTransmitTimeFailureUnknown PPPTransmitTimeFailureKind = 999
)

// SSRBiasStatus identifies a source's status for one bias query.
type SSRBiasStatus uint32

const (
	// SSRBiasAvailable reports that a bias record was found and valid.
	SSRBiasAvailable SSRBiasStatus = iota
	// SSRBiasMissing reports that no matching bias record exists.
	SSRBiasMissing
	// SSRBiasUnavailable reports that the source cannot provide this bias.
	SSRBiasUnavailable
	// SSRBiasNotYetValid reports that the record starts after the query epoch.
	SSRBiasNotYetValid
	// SSRBiasExpired reports that the record ended before the query epoch.
	SSRBiasExpired
	// SSRBiasExcluded reports that source policy excludes the record.
	SSRBiasExcluded
	// SSRBiasInvalidEpoch reports that the query epoch is not usable.
	SSRBiasInvalidEpoch
	// SSRBiasPhaseDiscontinuityNeedsReset reports a phase arc discontinuity.
	SSRBiasPhaseDiscontinuityNeedsReset
	// SSRBiasUnknownSignal reports that the signal is not known to the source.
	SSRBiasUnknownSignal
	// SSRBiasUnknown is the engine's explicit unknown tag.
	SSRBiasUnknown SSRBiasStatus = 999
)

// PPPSSRIFCombinationStatus identifies an ionosphere-free bias combination result.
type PPPSSRIFCombinationStatus uint32

const (
	// PPPSSRIFApplied reports that the ionosphere-free bias was applied.
	PPPSSRIFApplied PPPSSRIFCombinationStatus = iota
	// PPPSSRIFOptedOut reports that options disable this bias family.
	PPPSSRIFOptedOut
	// PPPSSRIFSignalUnavailable reports that a required signal is absent.
	PPPSSRIFSignalUnavailable
	// PPPSSRIFInvalidFrequencies reports that the frequencies cannot form the combination.
	PPPSSRIFInvalidFrequencies
	// PPPSSRIFObservationSignalsUnknown reports that the observation signal IDs are absent.
	PPPSSRIFObservationSignalsUnknown
	// PPPSSRIFCarrierUnresolved reports that the carrier could not be resolved.
	PPPSSRIFCarrierUnresolved
	// PPPSSRIFObservationFrequencyMismatch reports that the observed frequency differs.
	PPPSSRIFObservationFrequencyMismatch
	// PPPSSRIFIncompatibleSourceOrSolution reports source or solution disagreement.
	PPPSSRIFIncompatibleSourceOrSolution
	// PPPSSRIFIncompatibleIOD reports that the issue-of-data values do not match.
	PPPSSRIFIncompatibleIOD
	// PPPSSRIFOrbitClockSolutionUnavailable reports no applicable orbit-clock correction.
	PPPSSRIFOrbitClockSolutionUnavailable
	// PPPSSRIFOrbitClockSolutionMismatch reports a different orbit-clock solution.
	PPPSSRIFOrbitClockSolutionMismatch
	// PPPSSRIFSatelliteExcluded reports that source policy excludes the satellite.
	PPPSSRIFSatelliteExcluded
	// PPPSSRIFTransmitTimeUnavailable reports that the transmission time is unavailable.
	PPPSSRIFTransmitTimeUnavailable
	// PPPSSRIFPhaseDiscontinuityNeedsReset reports a phase-bias arc discontinuity.
	PPPSSRIFPhaseDiscontinuityNeedsReset
	// PPPSSRIFUT1OutsideCoverage reports that the evaluated epoch is outside UT1 coverage.
	PPPSSRIFUT1OutsideCoverage
	// PPPSSRIFUnknown is the engine's explicit unknown tag.
	PPPSSRIFUnknown PPPSSRIFCombinationStatus = 999
)

// PPPSSRExclusion preserves why a PPP observation lacked a required SSR bias.
type PPPSSRExclusion struct {
	// EpochIndex is the observation's zero-based input epoch index.
	EpochIndex int
	// SatelliteID is the satellite token.
	SatelliteID string
	// AmbiguityID is the observation's ambiguity identifier.
	AmbiguityID string
	// CodeBiasMissing reports a missing required code bias.
	CodeBiasMissing bool
	// PhaseBiasMissing reports a missing required phase bias.
	PhaseBiasMissing bool
	// TransmitTimeFailure retains the typed reason a source did not apply a bias.
	TransmitTimeFailure PPPSSRTransmitTimeFailure
	// Application retains the complete bias lookup row and signal reports.
	Application PPPSSRApplication
	// ErrorText is the source diagnostic when the failure is a source error.
	ErrorText string
}

// PPPSSRTransmitTimeFailure records the typed transmission-time failure.
type PPPSSRTransmitTimeFailure struct {
	// Kind is the native transmit-time failure tag.
	Kind PPPTransmitTimeFailureKind
	// HasTransmitTime controls TransmitTimeJ2000S.
	HasTransmitTime bool
	// TransmitTimeJ2000S is the transmission time in seconds from J2000.
	TransmitTimeJ2000S float64
	// HasApplied controls Applied.
	HasApplied bool
	// Applied identifies the applied orbit-clock source solution.
	Applied PPPSSRSolutionID
	// HasSignal controls Signal and BiasStatus.
	HasSignal bool
	// Signal identifies the bias record signal.
	Signal PPPSSRSignalKey
	// BiasStatus is the native status of the signal-bias query.
	BiasStatus SSRBiasStatus
	// UnknownVariant retains a future native kind label.
	UnknownVariant string
}

// PPPSSRSolutionID identifies an SSR/HAS correction solution.
type PPPSSRSolutionID struct {
	// Source identifies RTCM SSR, Galileo HAS, or IGS SSR.
	Source SSRSource
	// ProviderID is the SSR provider identifier.
	ProviderID uint16
	// SolutionID is the SSR solution identifier.
	SolutionID uint8
}

// PPPSSRSignalKey identifies a physical signal or a source-specific raw index.
type PPPSSRSignalKey struct {
	// IsPhysical reports whether Code names a physical signal.
	IsPhysical bool
	// System is the GNSS system tag.
	System GNSSSystem
	// Code is the RINEX band and tracking attribute when IsPhysical.
	Code string
	// Source identifies the source of a raw signal index.
	Source SSRSource
	// Index is the raw signal index when the source has no physical mapping.
	Index uint8
}

// PPPSSRSignalReport retains one signal's bias query result.
type PPPSSRSignalReport struct {
	// Present reports whether the query row exists.
	Present bool
	// Signal is the requested signal.
	Signal PPPSSRSignalKey
	// Status is the native query status.
	Status SSRBiasStatus
	// HasSourceSignal controls SourceSignal.
	HasSourceSignal bool
	// SourceSignal is the raw signal found in the source row.
	SourceSignal PPPSSRSignalKey
	// HasBiasM controls BiasM.
	HasBiasM bool
	// BiasM is the bias in metres.
	BiasM float64
	// HasBiasCycles controls BiasCycles.
	HasBiasCycles bool
	// BiasCycles is the phase bias in cycles.
	BiasCycles float64
	// HasSolution controls Solution.
	HasSolution bool
	// Solution is the source solution for the row.
	Solution PPPSSRSolutionID
	// HasIODSSR controls IODSSR.
	HasIODSSR bool
	// IODSSR is the SSR issue of data.
	IODSSR uint8
	// UnknownVariant retains an unrecognized native status label.
	UnknownVariant string
}

// PPPSSRApplication is the complete SSR/HAS lookup result for one observation.
type PPPSSRApplication struct {
	// Present reports whether a lookup row exists.
	Present bool
	// HasTransmitTime controls TransmitTimeJ2000S.
	HasTransmitTime bool
	// TransmitTimeJ2000S is the time used to evaluate the biases.
	TransmitTimeJ2000S float64
	// HasAppliedOrbitClockSolution controls AppliedOrbitClockSolution.
	HasAppliedOrbitClockSolution bool
	// AppliedOrbitClockSolution identifies the applied orbit-clock correction.
	AppliedOrbitClockSolution PPPSSRSolutionID
	// HasObservationSignals reports whether observation signal codes are present.
	HasObservationSignals bool
	// Code1Signal is the first code tracking identifier.
	Code1Signal string
	// Code2Signal is the second code tracking identifier.
	Code2Signal string
	// Phase1Signal is the first phase tracking identifier.
	Phase1Signal string
	// Phase2Signal is the second phase tracking identifier.
	Phase2Signal string
	// CodeStatus is the ionosphere-free code combination status.
	CodeStatus PPPSSRIFCombinationStatus
	// CodeStatusUT1 is the code UT1 coverage side tag.
	CodeStatusUT1 UT1DegradeReason
	// HasAppliedCodeIFM controls AppliedCodeIFM.
	HasAppliedCodeIFM bool
	// AppliedCodeIFM is the applied ionosphere-free code bias in metres.
	AppliedCodeIFM float64
	// Code1 is the first code signal lookup report.
	Code1 PPPSSRSignalReport
	// Code2 is the second code signal lookup report.
	Code2 PPPSSRSignalReport
	// PhaseStatus is the ionosphere-free phase combination status.
	PhaseStatus PPPSSRIFCombinationStatus
	// PhaseStatusUT1 is the phase UT1 coverage side tag.
	PhaseStatusUT1 UT1DegradeReason
	// HasAppliedPhaseIFM controls AppliedPhaseIFM.
	HasAppliedPhaseIFM bool
	// AppliedPhaseIFM is the applied ionosphere-free phase bias in metres.
	AppliedPhaseIFM float64
	// Phase1 is the first phase signal lookup report.
	Phase1 PPPSSRSignalReport
	// Phase2 is the second phase signal lookup report.
	Phase2 PPPSSRSignalReport
	// UnknownVariant retains an unrecognized native status label.
	UnknownVariant string
}

func publicPPPSSRSolutionID(value native.PppSsrSolutionID) PPPSSRSolutionID {
	return PPPSSRSolutionID{Source: SSRSource(value.Source), ProviderID: value.ProviderID, SolutionID: value.SolutionID}
}
func publicPPPSSRSignalKey(value native.PppSsrSignalKey) PPPSSRSignalKey {
	return PPPSSRSignalKey{IsPhysical: value.IsPhysical, System: GNSSSystem(value.System), Code: value.Code, Source: SSRSource(value.Source), Index: value.Index}
}
func publicPPPSSRSignalReport(value native.PppSsrSignalReport) PPPSSRSignalReport {
	return PPPSSRSignalReport{Present: value.Present, Signal: publicPPPSSRSignalKey(value.Signal), Status: SSRBiasStatus(value.Status), HasSourceSignal: value.HasSourceSignal, SourceSignal: publicPPPSSRSignalKey(value.SourceSignal), HasBiasM: value.HasBiasM, BiasM: value.BiasM, HasBiasCycles: value.HasBiasCycles, BiasCycles: value.BiasCycles, HasSolution: value.HasSolution, Solution: publicPPPSSRSolutionID(value.Solution), HasIODSSR: value.HasIODSSR, IODSSR: value.IODSSR, UnknownVariant: value.UnknownVariant}
}
func publicPPPSSRApplication(value native.PppSsrApplication) PPPSSRApplication {
	return PPPSSRApplication{Present: value.Present, HasTransmitTime: value.HasTransmitTime, TransmitTimeJ2000S: value.TransmitTimeJ2000S, HasAppliedOrbitClockSolution: value.HasAppliedOrbitClockSolution, AppliedOrbitClockSolution: publicPPPSSRSolutionID(value.AppliedOrbitClockSolution), HasObservationSignals: value.HasObservationSignals, Code1Signal: value.Code1Signal, Code2Signal: value.Code2Signal, Phase1Signal: value.Phase1Signal, Phase2Signal: value.Phase2Signal, CodeStatus: PPPSSRIFCombinationStatus(value.CodeStatus), CodeStatusUT1: UT1DegradeReason(value.CodeStatusUT1), HasAppliedCodeIFM: value.HasAppliedCodeIFM, AppliedCodeIFM: value.AppliedCodeIFM, Code1: publicPPPSSRSignalReport(value.Code1), Code2: publicPPPSSRSignalReport(value.Code2), PhaseStatus: PPPSSRIFCombinationStatus(value.PhaseStatus), PhaseStatusUT1: UT1DegradeReason(value.PhaseStatusUT1), HasAppliedPhaseIFM: value.HasAppliedPhaseIFM, AppliedPhaseIFM: value.AppliedPhaseIFM, Phase1: publicPPPSSRSignalReport(value.Phase1), Phase2: publicPPPSSRSignalReport(value.Phase2), UnknownVariant: value.UnknownVariant}
}

func publicPPPSSRExclusions(values []native.PppSsrBiasExclusion) []PPPSSRExclusion {
	out := make([]PPPSSRExclusion, len(values))
	for i, value := range values {
		failure := value.TransmitTimeFailure
		out[i] = PPPSSRExclusion{EpochIndex: value.EpochIndex, SatelliteID: value.SatelliteID, AmbiguityID: value.AmbiguityID, CodeBiasMissing: value.CodeBiasMissing, PhaseBiasMissing: value.PhaseBiasMissing, TransmitTimeFailure: PPPSSRTransmitTimeFailure{Kind: PPPTransmitTimeFailureKind(failure.Kind), HasTransmitTime: failure.HasTransmitTime, TransmitTimeJ2000S: failure.TransmitTimeJ2000S, HasApplied: failure.HasApplied, Applied: publicPPPSSRSolutionID(failure.Applied), HasSignal: failure.HasSignal, Signal: publicPPPSSRSignalKey(failure.Signal), BiasStatus: SSRBiasStatus(failure.BiasStatus), UnknownVariant: failure.UnknownVariant}, Application: publicPPPSSRApplication(value.Application), ErrorText: value.ErrorText}
	}
	return out
}

// SSRBiasExclusions returns detached, typed observations excluded by missing SSR/HAS biases.
func (s *PPPFloatSolution) SSRBiasExclusions() ([]PPPSSRExclusion, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.SSRBiasExclusions()
	runtime.KeepAlive(s)
	if err != nil {
		return nil, publicError(err)
	}
	return publicPPPSSRExclusions(values), nil
}

// SSRBiasExclusions returns detached, typed observations excluded by missing SSR/HAS biases.
func (s *PPPFixedSolution) SSRBiasExclusions() ([]PPPSSRExclusion, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	values, err := s.handle.SSRBiasExclusions()
	runtime.KeepAlive(s)
	if err != nil {
		return nil, publicError(err)
	}
	return publicPPPSSRExclusions(values), nil
}
