package native

type PppSsrSolutionID struct {
	Source     uint32
	ProviderID uint16
	SolutionID uint8
}
type PppSsrSignalKey struct {
	IsPhysical bool
	System     uint32
	Code       string
	Source     uint32
	Index      uint8
}
type PppTransmitTimeFailure struct {
	Kind               uint32
	HasTransmitTime    bool
	TransmitTimeJ2000S float64
	HasApplied         bool
	Applied            PppSsrSolutionID
	HasSignal          bool
	Signal             PppSsrSignalKey
	BiasStatus         uint32
	UnknownVariant     string
}
type PppSsrSignalReport struct {
	Present         bool
	Signal          PppSsrSignalKey
	Status          uint32
	HasSourceSignal bool
	SourceSignal    PppSsrSignalKey
	HasBiasM        bool
	BiasM           float64
	HasBiasCycles   bool
	BiasCycles      float64
	HasSolution     bool
	Solution        PppSsrSolutionID
	HasIODSSR       bool
	IODSSR          uint8
	UnknownVariant  string
}
type PppSsrApplication struct {
	Present                                              bool
	HasTransmitTime                                      bool
	TransmitTimeJ2000S                                   float64
	HasAppliedOrbitClockSolution                         bool
	AppliedOrbitClockSolution                            PppSsrSolutionID
	HasObservationSignals                                bool
	Code1Signal, Code2Signal, Phase1Signal, Phase2Signal string
	CodeStatus, CodeStatusUT1                            uint32
	HasAppliedCodeIFM                                    bool
	AppliedCodeIFM                                       float64
	Code1, Code2                                         PppSsrSignalReport
	PhaseStatus, PhaseStatusUT1                          uint32
	HasAppliedPhaseIFM                                   bool
	AppliedPhaseIFM                                      float64
	Phase1, Phase2                                       PppSsrSignalReport
	UnknownVariant                                       string
}
type PppSsrBiasExclusion struct {
	EpochIndex                        int
	SatelliteID, AmbiguityID          string
	CodeBiasMissing, PhaseBiasMissing bool
	TransmitTimeFailure               PppTransmitTimeFailure
	Application                       PppSsrApplication
	ErrorText                         string
}
