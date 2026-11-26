package sidereon

import "sidereon.dev/go/v4/internal/native"

// ForceParameter is a named force-model parameter and its configured value.
// Values can be compared with ==.
type ForceParameter struct {
	// Name is the name used to select the parameter for estimation.
	Name string
	// Value is the configured parameter value in its model's units.
	Value float64
}

// Parameters returns the force model's parameters in evaluation order.
func (model *ForceModel) Parameters() ([]ForceParameter, error) {
	values, err := native.ForceModelParameters(nativeForceModelValue(model))
	return orbitMap(values, func(value native.ForceParameter) ForceParameter {
		return ForceParameter(value)
	}), publicError(err)
}

// NewCompositeForceModel adds the supplied forces in evaluation order.
// The composite retains each force independently of the input handles.
func NewCompositeForceModel(models ...Force) (*CompositeForceModel, error) {
	value, err := native.CompositeForceModelNew(orbitMap(models, nativeForceHandle))
	if err != nil {
		return nil, publicError(err)
	}
	return publicCompositeForceModel(value), nil
}

// Constants returns the JPL ephemeris constants keyed by their file names.
func (ephemeris *DEEphemeris) Constants() (map[string]float64, error) {
	value, err := native.DEEphemerisConstants(nativeDEEphemerisHandle(ephemeris))
	return value, publicError(err)
}

// VariationalEquations selects the force and parameter sensitivity columns.
type VariationalEquations struct {
	// Force provides the acceleration and its derivatives.
	Force Force
	// Parameters specifies replacement values in force parameter order.
	Parameters []float64
	// SensitivityParameters lists the parameter indices in output column order.
	SensitivityParameters []uint64
}

// PropagateVariational integrates states, state transitions and parameter sensitivities.
func PropagateVariational(initial CartesianState, epochs []VariationalEpoch, equations VariationalEquations, options IntegratorOptions, context *PropagationContext) (*VariationalTrajectory, error) {
	input := native.VariationalEquations{
		Force:                 nativeForceHandle(equations.Force),
		Parameters:            equations.Parameters,
		SensitivityParameters: equations.SensitivityParameters,
	}
	value, err := native.PropagateVariational(nativeCartesian(initial), orbitMap(epochs, nativeVariationalEpoch), input, nativeIntegratorOptions(options), nativePropagationContext(context))
	if err != nil {
		return nil, publicError(err)
	}
	return publicVariationalTrajectory(value), nil
}

func nativeVariationalEpoch(value VariationalEpoch) native.VariationalEpoch {
	return native.VariationalEpoch(value)
}
