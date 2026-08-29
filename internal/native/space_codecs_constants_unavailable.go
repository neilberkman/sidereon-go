//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

const (
	OMMFormatKVNValue  uint32 = 0
	OMMFormatXMLValue  uint32 = 1
	OMMFormatJSONValue uint32 = 2

	BiasModeAbsoluteValue            uint32 = 0
	BiasModeRelativeValue            uint32 = 1
	BiasModeUnspecifiedValue         uint32 = 2
	BiasKindOSBValue                 uint32 = 0
	BiasKindDSBValue                 uint32 = 1
	BiasKindISBValue                 uint32 = 2
	BiasTargetSystemValue            uint32 = 0
	BiasTargetSatelliteValue         uint32 = 1
	BiasTargetReceiverValue          uint32 = 2
	BiasTargetSatelliteReceiverValue uint32 = 3

	GNSSSystemGPS     uint32 = 0
	GNSSSystemGLONASS uint32 = 1
	GNSSSystemGalileo uint32 = 2
	GNSSSystemBeiDou  uint32 = 3
	GNSSSystemQZSS    uint32 = 4
	GNSSSystemNavIC   uint32 = 5
	GNSSSystemSBAS    uint32 = 6

	CarrierBandL1  uint32 = 0
	CarrierBandL2  uint32 = 1
	CarrierBandL5  uint32 = 2
	CarrierBandE1  uint32 = 3
	CarrierBandE5A uint32 = 4
	CarrierBandE5B uint32 = 5
	CarrierBandE5  uint32 = 6
	CarrierBandE6  uint32 = 7
	CarrierBandB1C uint32 = 8
	CarrierBandB1I uint32 = 9
	CarrierBandB2A uint32 = 10
	CarrierBandB2B uint32 = 11
	CarrierBandB2  uint32 = 12
	CarrierBandB3I uint32 = 13
	CarrierBandG1  uint32 = 14
	CarrierBandG2  uint32 = 15

	TimeScaleUTC      uint32 = 0
	TimeScaleTAI      uint32 = 1
	TimeScaleTT       uint32 = 2
	TimeScaleTDB      uint32 = 3
	TimeScaleGPST     uint32 = 4
	TimeScaleGST      uint32 = 5
	TimeScaleBDT      uint32 = 6
	TimeScaleGLONASST uint32 = 7
	TimeScaleQZSST    uint32 = 8
	TimeScaleTCG      uint32 = 9
	TimeScaleTCB      uint32 = 10

	AllanSeriesPhaseSecondsValue                uint32 = 0
	AllanSeriesFractionalFrequencyValue         uint32 = 1
	AllanSeriesPhaseSecondsWithGapsValue        uint32 = 2
	AllanSeriesFractionalFrequencyWithGapsValue uint32 = 3
	AllanTauGridOctaveValue                     uint32 = 0
	AllanTauGridAllValue                        uint32 = 1
	AllanTauGridExplicitValue                   uint32 = 2
	GapPolicyRejectValue                        uint32 = 0
	GapPolicyOmitTermsValue                     uint32 = 1
	AllanEstimatorADEVValue                     uint32 = 0
	AllanEstimatorOverlappingADEVValue          uint32 = 1
	AllanEstimatorMDEVValue                     uint32 = 2
	AllanEstimatorHDEVValue                     uint32 = 3
	AllanEstimatorTDEVValue                     uint32 = 4

	PowerLawRandomWalkFMValue       uint32 = 0
	PowerLawFlickerFMValue          uint32 = 1
	PowerLawWhiteFMValue            uint32 = 2
	PowerLawFlickerPMValue          uint32 = 3
	PowerLawWhitePMValue            uint32 = 4
	PowerLawOctaveUnderSampledValue uint32 = 0
	PowerLawOctaveDegenerateValue   uint32 = 1
	PowerLawOctaveMissingMDEVValue  uint32 = 2
	PowerLawDominantValue           uint32 = 0
	PowerLawAmbiguousValue          uint32 = 1
	PowerLawFlaggedValue            uint32 = 2

	TDMObservableRangeValue                uint32 = 0
	TDMObservableDopplerInstantaneousValue uint32 = 1
	TDMObservableDopplerIntegratedValue    uint32 = 2
	TDMObservableReceiveFrequencyValue     uint32 = 3
	TDMObservableTransmitFrequencyValue    uint32 = 4
	TDMObservableTransmitRateValue         uint32 = 5
	TDMObservableAngle1Value               uint32 = 6
	TDMObservableAngle2Value               uint32 = 7
	TDMObservableOtherValue                uint32 = 255
	TDMUnitKilometersValue                 uint32 = 0
	TDMUnitSecondsValue                    uint32 = 1
	TDMUnitRangeUnitsValue                 uint32 = 2
	TDMUnitKilometersPerSecondValue        uint32 = 3
	TDMUnitHertzValue                      uint32 = 4
	TDMUnitHertzPerSecondValue             uint32 = 5
	TDMUnitDegreesValue                    uint32 = 6
	TDMUnitDecibelWattsValue               uint32 = 7
	TDMUnitDecibelHertzValue               uint32 = 8
	TDMUnitSquareMetersValue               uint32 = 9
	TDMUnitMetersValue                     uint32 = 10
	TDMUnitSecondsPerSecondValue           uint32 = 11
	TDMUnitPercentValue                    uint32 = 12
	TDMUnitKelvinValue                     uint32 = 13
	TDMUnitHectopascalsValue               uint32 = 14
	TDMUnitTotalElectronContentUnitsValue  uint32 = 15
	TDMUnitDimensionlessValue              uint32 = 16
	TDMUnitUnknownValue                    uint32 = 255
)
