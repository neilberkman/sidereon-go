package sidereon

import (
	"math"
	"testing"
)

// referenceEccentricAnomaly solves M=E-e*sin(E) independently of the binding.
func referenceEccentricAnomaly(mean, eccentricity float64) float64 {
	e := mean
	for i := 0; i < 32; i++ {
		e -= (e - eccentricity*math.Sin(e) - mean) / (1 - eccentricity*math.Cos(e))
	}
	return e
}

func TestAnomalyConversionsAgainstIndependentOrbitEquations(t *testing.T) {
	const mean, eccentricity = 0.81, 0.23
	eccentric := referenceEccentricAnomaly(mean, eccentricity)
	trueAnomaly := math.Atan2(
		math.Sqrt(1-eccentricity*eccentricity)*math.Sin(eccentric),
		math.Cos(eccentric)-eccentricity,
	)
	// Core's documented Newton residual bound is 1e-12 + 1e-14*|M|.
	// Since dM/dE = 1-e*cos(E) >= 1-e, residual/(1-e) bounds the induced
	// eccentric-anomaly error. For true anomaly, |dν/dM| =
	// sqrt(1-e²)/(1-e*cos(E))² < 2 at this e, so twice the residual bound is
	// sufficient. Add ULPs only for formula evaluation; closed-form conversions
	// use that ULP bound alone.
	const machineEpsilon = 2.220446049250313e-16
	residualTolerance := 1e-12 + 1e-14*math.Abs(mean)
	eccentricSolverTolerance := residualTolerance/(1-eccentricity) + 32*machineEpsilon
	trueSolverTolerance := 2*residualTolerance + 32*machineEpsilon
	closedFormTolerance := 32 * machineEpsilon

	checks := []struct {
		name string
		call func() (float64, error)
		want float64
		tol  float64
	}{
		{
			name: "mean to eccentric",
			call: func() (float64, error) { return MeanToEccentricAnomaly(mean, eccentricity) },
			want: eccentric,
			tol:  eccentricSolverTolerance,
		},
		{
			name: "eccentric to mean",
			call: func() (float64, error) { return EccentricToMeanAnomaly(eccentric, eccentricity) },
			want: eccentric - eccentricity*math.Sin(eccentric),
			tol:  closedFormTolerance,
		},
		{
			name: "eccentric to true",
			call: func() (float64, error) { return EccentricToTrueAnomaly(eccentric, eccentricity) },
			want: trueAnomaly,
			tol:  closedFormTolerance,
		},
		{
			name: "true to eccentric",
			call: func() (float64, error) { return TrueToEccentricAnomaly(trueAnomaly, eccentricity) },
			want: eccentric,
			tol:  closedFormTolerance,
		},
		{
			name: "mean to true",
			call: func() (float64, error) { return MeanToTrueAnomaly(mean, eccentricity) },
			want: trueAnomaly,
			tol:  trueSolverTolerance,
		},
		{
			name: "true to mean",
			call: func() (float64, error) { return TrueToMeanAnomaly(trueAnomaly, eccentricity) },
			want: eccentric - eccentricity*math.Sin(eccentric),
			tol:  closedFormTolerance,
		},
	}
	for _, check := range checks {
		got, err := check.call()
		if err != nil {
			t.Errorf("%s returned error: %v", check.name, err)
			continue
		}
		if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-check.want) > check.tol {
			t.Errorf("%s = %.17g, want finite independent orbital result %.17g (tolerance %.3g)", check.name, got, check.want, check.tol)
		}
	}
}
