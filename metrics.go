package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// ErrorEllipse2 is a confidence ellipse from a 2x2 covariance block. Axes
// have the covariance's units and orientation is radians.
type ErrorEllipse2 struct {
	Confidence     float64
	ChiSquareScale float64
	SemiMajor      float64
	SemiMinor      float64
	OrientationRad float64
}

// ErrorEllipse is the one-sigma horizontal ellipse in metres.
type ErrorEllipse struct {
	SemiMajorM     float64
	SemiMinorM     float64
	OrientationRad float64
}

// PercentileRadius is a C-computed circle or sphere radius in metres.
type PercentileRadius struct {
	Probability float64
	RadiusM     float64
	ApproxM     float64
	ApproxValid bool
}

// PositionErrorMetrics contains standard C-computed position-error measures.
type PositionErrorMetrics struct {
	Ellipse ErrorEllipse
	SigmaEM float64
	SigmaNM float64
	SigmaUM float64
	CEP     PercentileRadius
	R95     PercentileRadius
	R99     PercentileRadius
	DRMS    float64
	TwoDRMS float64
	VEP     float64
	SEP     PercentileRadius
	MRSE    float64
}

// ErrorMetricsErrorKind is the detailed metric-domain diagnostic written by C.
type ErrorMetricsErrorKind uint32

const (
	ErrorMetricsNone                    ErrorMetricsErrorKind = 0
	ErrorMetricsNonFinite               ErrorMetricsErrorKind = 1
	ErrorMetricsNotPositiveSemidefinite ErrorMetricsErrorKind = 2
	ErrorMetricsInvalidProbability      ErrorMetricsErrorKind = 3
	ErrorMetricsRotation                ErrorMetricsErrorKind = 4
)

// ErrorEllipse2x2 computes a confidence ellipse from a row-major 2x2
// covariance block. Confidence is in [0, 1).
func ErrorEllipse2x2(covariance [4]float64, confidence float64) (ErrorEllipse2, error) {
	value, err := native.ErrorEllipse2x2(covariance, confidence)
	return ErrorEllipse2{
		Confidence: value.Confidence, ChiSquareScale: value.ChiSquareScale,
		SemiMajor: value.SemiMajor, SemiMinor: value.SemiMinor, OrientationRad: value.OrientationRad,
	}, publicError(err)
}

func nativeMatrix3(value [3][3]float64) [3][3]float64 { return value }

func publicMetrics(value native.PositionErrorMetrics) PositionErrorMetrics {
	return PositionErrorMetrics{
		Ellipse: ErrorEllipse{
			SemiMajorM: value.Ellipse.SemiMajorM, SemiMinorM: value.Ellipse.SemiMinorM, OrientationRad: value.Ellipse.OrientationRad,
		},
		SigmaEM: value.SigmaEM, SigmaNM: value.SigmaNM, SigmaUM: value.SigmaUM,
		CEP:  PercentileRadius{Probability: value.CEP.Probability, RadiusM: value.CEP.RadiusM, ApproxM: value.CEP.ApproxM, ApproxValid: value.CEP.ApproxValid},
		R95:  PercentileRadius{Probability: value.R95.Probability, RadiusM: value.R95.RadiusM, ApproxM: value.R95.ApproxM, ApproxValid: value.R95.ApproxValid},
		R99:  PercentileRadius{Probability: value.R99.Probability, RadiusM: value.R99.RadiusM, ApproxM: value.R99.ApproxM, ApproxValid: value.R99.ApproxValid},
		DRMS: value.DRMS, TwoDRMS: value.TwoDRMS, VEP: value.VEP,
		SEP:  PercentileRadius{Probability: value.SEP.Probability, RadiusM: value.SEP.RadiusM, ApproxM: value.SEP.ApproxM, ApproxValid: value.SEP.ApproxValid},
		MRSE: value.MRSE,
	}
}

// ErrorMetricsFromENU computes standard metrics from an ENU covariance in
// square metres. The returned kind is the C diagnostic detail.
func ErrorMetricsFromENU(covariance [3][3]float64) (PositionErrorMetrics, ErrorMetricsErrorKind, error) {
	value, kind, err := native.ErrorMetricsFromENU(nativeMatrix3(covariance))
	return publicMetrics(value), ErrorMetricsErrorKind(kind), publicError(err)
}

// ErrorMetricsFromECEF rotates an ECEF covariance into the local ENU frame
// at receiver and computes standard metrics. Covariance entries are m².
func ErrorMetricsFromECEF(covariance [3][3]float64, receiver Geodetic) (PositionErrorMetrics, ErrorMetricsErrorKind, error) {
	value, kind, err := native.ErrorMetricsFromECEF(nativeMatrix3(covariance), native.Geodetic{
		LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM,
	})
	return publicMetrics(value), ErrorMetricsErrorKind(kind), publicError(err)
}

// ErrorMetricsFromKinematicSolution computes metrics from an ECEF position and
// its 3x3 position covariance.
func ErrorMetricsFromKinematicSolution(position [3]float64, covariance [3][3]float64) (PositionErrorMetrics, ErrorMetricsErrorKind, error) {
	value, kind, err := native.ErrorMetricsFromKinematic(position, nativeMatrix3(covariance))
	return publicMetrics(value), ErrorMetricsErrorKind(kind), publicError(err)
}

// ErrorMetricsFromPositionCovariance computes metrics from an already paired
// ECEF and ENU covariance, both in m².
func ErrorMetricsFromPositionCovariance(ecef, enu [3][3]float64) (PositionErrorMetrics, ErrorMetricsErrorKind, error) {
	value, kind, err := native.ErrorMetricsFromPositionCovariance(nativeMatrix3(ecef), nativeMatrix3(enu))
	return publicMetrics(value), ErrorMetricsErrorKind(kind), publicError(err)
}

// ErrorEllipseFromENU computes the one-sigma horizontal ellipse from an ENU
// covariance in m².
func ErrorEllipseFromENU(covariance [3][3]float64) (ErrorEllipse, ErrorMetricsErrorKind, error) {
	value, kind, err := native.ErrorEllipseFromENU(nativeMatrix3(covariance))
	return ErrorEllipse{SemiMajorM: value.SemiMajorM, SemiMinorM: value.SemiMinorM, OrientationRad: value.OrientationRad}, ErrorMetricsErrorKind(kind), publicError(err)
}

// HorizontalProtectionRadius returns the exact C percentile circle radius for
// an ENU covariance in m².
func HorizontalProtectionRadius(covariance [3][3]float64, probability float64) (PercentileRadius, ErrorMetricsErrorKind, error) {
	value, kind, err := native.HorizontalRadius(nativeMatrix3(covariance), probability)
	return PercentileRadius{Probability: value.Probability, RadiusM: value.RadiusM, ApproxM: value.ApproxM, ApproxValid: value.ApproxValid}, ErrorMetricsErrorKind(kind), publicError(err)
}

// SphericalProtectionRadius returns the exact C percentile sphere radius for
// an ENU covariance in m².
func SphericalProtectionRadius(covariance [3][3]float64, probability float64) (PercentileRadius, ErrorMetricsErrorKind, error) {
	value, kind, err := native.SphericalRadius(nativeMatrix3(covariance), probability)
	return PercentileRadius{Probability: value.Probability, RadiusM: value.RadiusM, ApproxM: value.ApproxM, ApproxValid: value.ApproxValid}, ErrorMetricsErrorKind(kind), publicError(err)
}

// VerticalProtectionRadius returns the C percentile vertical radius in metres
// for an up variance in m².
func VerticalProtectionRadius(sigmaUM2, probability float64) (float64, ErrorMetricsErrorKind, error) {
	value, kind, err := native.VerticalRadius(sigmaUM2, probability)
	return value, ErrorMetricsErrorKind(kind), publicError(err)
}
