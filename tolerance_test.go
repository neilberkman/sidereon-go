package sidereon

import "math"

// Tolerances for cross-architecture floating-point comparisons in frozen
// fixture tests. The engine documents only same-machine bit reproducibility
// ("cross-platform bit identity holds only where a test pins it" -
// github.com/neilberkman/sidereon README); these fixtures were captured on
// arm64 and differ from x86_64 by a few ULP in transcendental-heavy solves
// (least squares, SGP4, Kepler iteration, velocity/covariance propagation).
//
// closeTol mirrors numpy.allclose: |got-want| <= absTol + relTol*|want|. A
// pure absolute bound is wrong here because these results include covariance
// matrices whose entries span many orders of magnitude (down to ~1e-17);
// relTol carries large-magnitude values (position, velocity in meters or
// m/s), absTol is a floor for near-zero entries. relTol=1e-9 gives roughly
// four million ULP of headroom above the observed few-ULP divergence while
// staying far tighter than any real regression, which moves answers by at
// least 1e-6 relative (e.g. a wrong satellite shifts position by meters out
// of a ~1e6 m baseline).
const (
	relTol = 1e-9

	toleranceM     = 1e-6  // meters: position, velocity(m/s), baseline, residual, delta, height
	toleranceKm    = 1e-9  // kilometers: SGP4 position/velocity
	toleranceM2    = 1e-9  // meters^2 (or (m/s)^2): covariance
	toleranceRad   = 1e-12 // radians: geodetic latitude/longitude
	toleranceS     = 1e-12 // seconds: clock bias, time-of-flight
	toleranceRatio = 1e-9  // dimensionless: GDOP, condition number, weights, RMS
)

func closeTol(got, want, absTol float64) bool {
	if math.IsNaN(want) {
		return math.IsNaN(got)
	}
	return math.Abs(got-want) <= absTol+relTol*math.Abs(want)
}

func closeArray3(got, want [3]float64, absTol float64) bool {
	return closeTol(got[0], want[0], absTol) && closeTol(got[1], want[1], absTol) && closeTol(got[2], want[2], absTol)
}

func closeArray4(got, want [4]float64, absTol float64) bool {
	return closeTol(got[0], want[0], absTol) && closeTol(got[1], want[1], absTol) &&
		closeTol(got[2], want[2], absTol) && closeTol(got[3], want[3], absTol)
}

func closeArray9(got, want [9]float64, absTol float64) bool {
	for i := range got {
		if !closeTol(got[i], want[i], absTol) {
			return false
		}
	}
	return true
}

func closeSlice(got, want []float64, absTol float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !closeTol(got[i], want[i], absTol) {
			return false
		}
	}
	return true
}
