package sidereon

import (
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

func TestPPPSSRExclusionCopiesTypedTagsAndNestedPayload(t *testing.T) {
	input := []native.PppSsrBiasExclusion{{
		EpochIndex: 17, SatelliteID: "G06", AmbiguityID: "G06:L1C", CodeBiasMissing: true, PhaseBiasMissing: true,
		TransmitTimeFailure: native.PppTransmitTimeFailure{Kind: 999, HasTransmitTime: true, TransmitTimeJ2000S: 123.5, HasApplied: true, Applied: native.PppSsrSolutionID{Source: 2, ProviderID: 42, SolutionID: 7}, HasSignal: true, Signal: native.PppSsrSignalKey{IsPhysical: true, System: uint32(GNSSSystemGPS), Code: "1C", Source: 2, Index: 9}, BiasStatus: 4242, UnknownVariant: "FutureFailure"},
		Application:         native.PppSsrApplication{Present: true, HasTransmitTime: true, TransmitTimeJ2000S: 123.5, HasAppliedOrbitClockSolution: true, AppliedOrbitClockSolution: native.PppSsrSolutionID{Source: 2, ProviderID: 42, SolutionID: 7}, HasObservationSignals: true, Code1Signal: "1C", Code2Signal: "2W", Phase1Signal: "1C", Phase2Signal: "2W", CodeStatus: 999, CodeStatusUT1: 2, HasAppliedCodeIFM: true, AppliedCodeIFM: 0.12, Code1: native.PppSsrSignalReport{Present: true, Signal: native.PppSsrSignalKey{IsPhysical: true, System: uint32(GNSSSystemGPS), Code: "1C"}, Status: 999, HasSourceSignal: true, SourceSignal: native.PppSsrSignalKey{Source: 2, Index: 3}, HasBiasM: true, BiasM: 0.25, HasBiasCycles: true, BiasCycles: -0.5, HasSolution: true, Solution: native.PppSsrSolutionID{Source: 2, ProviderID: 42, SolutionID: 7}, HasIODSSR: true, IODSSR: 19, UnknownVariant: "FutureBias"}, PhaseStatus: 14, PhaseStatusUT1: 1, HasAppliedPhaseIFM: true, AppliedPhaseIFM: -0.03, UnknownVariant: "FutureCombination"},
		ErrorText:           "retained source failure",
	}}
	got := publicPPPSSRExclusions(input)
	if len(got) != 1 {
		t.Fatalf("exclusions = %+v", got)
	}
	row := got[0]
	if row.EpochIndex != 17 || row.SatelliteID != "G06" || row.AmbiguityID != "G06:L1C" || !row.CodeBiasMissing || !row.PhaseBiasMissing || row.ErrorText != "retained source failure" {
		t.Fatalf("exclusion identity/details = %+v", row)
	}
	if row.TransmitTimeFailure.Kind != PPPTransmitTimeFailureUnknown || uint32(row.TransmitTimeFailure.Kind) != 999 || row.TransmitTimeFailure.BiasStatus != SSRBiasStatus(4242) || row.TransmitTimeFailure.UnknownVariant != "FutureFailure" {
		t.Fatalf("typed failure tags = %+v", row.TransmitTimeFailure)
	}
	if row.TransmitTimeFailure.Applied.Source != SSRSourceIGS || row.TransmitTimeFailure.Signal.System != GNSSSystemGPS || row.TransmitTimeFailure.Signal.Source != SSRSourceIGS {
		t.Fatalf("typed source/signal identity = %+v", row.TransmitTimeFailure)
	}
	app := row.Application
	if app.CodeStatus != PPPSSRIFUnknown || uint32(app.CodeStatus) != 999 || app.CodeStatusUT1 != UT1DegradeReason(2) || app.PhaseStatus != PPPSSRIFUT1OutsideCoverage || app.PhaseStatusUT1 != UT1DegradeReason(1) || app.UnknownVariant != "FutureCombination" {
		t.Fatalf("application tags = %+v", app)
	}
	if app.Code1.Status != SSRBiasUnknown || uint32(app.Code1.Status) != 999 || app.Code1.UnknownVariant != "FutureBias" || !app.Code1.HasBiasCycles || app.Code1.BiasCycles != -0.5 || app.Code1.Solution.Source != SSRSourceIGS {
		t.Fatalf("nested signal report = %+v", app.Code1)
	}
}
