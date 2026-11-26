package sidereon

import "strings"

type orbitalError struct {
	*EngineError
	detail string
}

// Error returns the native failure message.
func (err orbitalError) Error() string {
	if err.detail == "" {
		return err.EngineError.Error()
	}
	return strings.TrimPrefix(err.detail, err.Operation+": ")
}

// EclipseError retains the eclipse error kind and details.
type EclipseError struct{ orbitalError }

// ShadowGradientError retains the shadow gradient error kind and details.
type ShadowGradientError struct{ orbitalError }

// DEError retains the de error kind and details.
type DEError struct{ orbitalError }

// GNSSAttitudeError retains the gnss attitude error kind and details.
type GNSSAttitudeError struct{ orbitalError }

// GNSSBlockError retains the gnss block error kind and details.
type GNSSBlockError struct{ orbitalError }

// AnalyticPropagationError retains the analytic propagation error kind and details.
type AnalyticPropagationError struct{ orbitalError }

// ICGEMError retains the icgem error kind and details.
type ICGEMError struct{ orbitalError }

// OrbitParameterFitError retains the orbit parameter fit error kind and details.
type OrbitParameterFitError struct{ orbitalError }

// OrbitScoreError retains the orbit score error kind and details.
type OrbitScoreError struct{ orbitalError }

func orbitEngineError(engine *EngineError, detail string) error {
	value := orbitalError{EngineError: engine, detail: detail}
	switch engine.Family {
	case EngineErrorFamilyEclipse:
		return &EclipseError{value}
	case EngineErrorFamilyShadowGradient:
		return &ShadowGradientError{value}
	case EngineErrorFamilyDE:
		return &DEError{value}
	case EngineErrorFamilyGNSSAttitude:
		return &GNSSAttitudeError{value}
	case EngineErrorFamilyGNSSBlock:
		return &GNSSBlockError{value}
	case EngineErrorFamilyAnalyticPropagation:
		return &AnalyticPropagationError{value}
	case EngineErrorFamilyICGEM:
		return &ICGEMError{value}
	case EngineErrorFamilyOrbitParameterFit:
		return &OrbitParameterFitError{value}
	case EngineErrorFamilyOrbitScore:
		return &OrbitScoreError{value}
	default:
		return nil
	}
}
