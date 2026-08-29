package sidereon

import "fmt"

func validateGNSSSystem(value GNSSSystem) error {
	if value > GNSSSystemSBAS {
		return fmt.Errorf("sidereon: invalid GNSS system %d", value)
	}
	return nil
}

func validateTimeScale(value TimeScale) error {
	if value > TCB {
		return fmt.Errorf("sidereon: invalid time scale %d", value)
	}
	return nil
}

func validateSignalModulationKind(value SignalModulationKind) error {
	if value < SignalBPSK || value > SignalCBOCMinus {
		return fmt.Errorf("sidereon: invalid signal modulation kind %d", value)
	}
	return nil
}

func validateDLLProcessing(value DLLProcessing) error {
	if value != DLLCoherent && value != DLLNonCoherent {
		return fmt.Errorf("sidereon: invalid DLL processing %d", value)
	}
	return nil
}

func validateRTCMMSMKind(value RTCMMSMKind) error {
	if value != RTCMMSM4 && value != RTCMMSM7 {
		return fmt.Errorf("sidereon: invalid RTCM MSM kind %d", value)
	}
	return nil
}

func validateRTCMStringField(value RTCMAntennaStringField) error {
	if value < RTCMAntennaDescriptorField || value > RTCMReceiverSerialNumberField {
		return fmt.Errorf("sidereon: invalid RTCM antenna string field %d", value)
	}
	return nil
}

func validateSBASSolveMode(value SBASSolveMode) error {
	if value != SBASMixedAugmentation && value != SBASOnly {
		return fmt.Errorf("sidereon: invalid SBAS solve mode %d", value)
	}
	return nil
}

func validateSSRReferencePoint(value SSRReferencePoint) error {
	if value != SSRReferencePointAntennaPhaseCenter && value != SSRReferencePointCenterOfMass {
		return fmt.Errorf("sidereon: invalid SSR reference point %d", value)
	}
	return nil
}

func validateSSRMissingAction(value SSRMissingCorrectionAction) error {
	if value != SSRDeclineMissingCorrection && value != SSRFallbackToBroadcast {
		return fmt.Errorf("sidereon: invalid SSR missing-correction action %d", value)
	}
	return nil
}

func validateRINEXLintSeverity(value RINEXLintSeverity) error {
	if value > RINEXLintInfo {
		return fmt.Errorf("sidereon: invalid RINEX lint severity %d", value)
	}
	return nil
}

func validateSBASWireForm(value SBASWireForm) error {
	if value != SBASWireFramed250 && value != SBASWireBody226 {
		return fmt.Errorf("sidereon: invalid SBAS wire form %d", value)
	}
	return nil
}
