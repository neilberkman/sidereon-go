package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// Matrix3 is a row-major 3x3 matrix represented with Go-owned values.
type Matrix3 [3][3]float64

// GroundStation uses the units of the C frame and Doppler helpers: latitude
// and longitude are degrees, and altitude is kilometres.
type GroundStation struct {
	LatitudeDeg  float64
	LongitudeDeg float64
	AltitudeKm   float64
}

// Topocentric contains azimuth in degrees clockwise from north, elevation in
// degrees above the horizon, and slant range in kilometres.
type Topocentric struct {
	AzimuthDeg   float64
	ElevationDeg float64
	RangeKm      float64
}

// PolarMotionMatrix returns the C IERS polar-motion matrix for xp and yp in
// arcseconds.
func PolarMotionMatrix(xpArcsec, ypArcsec float64) (Matrix3, error) {
	value, err := native.PolarMotionMatrix(xpArcsec, ypArcsec)
	return Matrix3(value), publicError(err)
}

// GCRSToITRSMatrix returns the C row-major rotation matrix for the supplied
// time scales.
func GCRSToITRSMatrix(scales TimeScales) (Matrix3, error) {
	value, err := native.GCRSToITRSMatrix(native.TimeScales{
		JDWhole: scales.JDWhole, UT1Fraction: scales.UT1Fraction, TTFraction: scales.TTFraction,
		TDBFraction: scales.TDBFraction, JDUT1: scales.JDUT1, JDTT: scales.JDTT, JDTDB: scales.JDTDB,
	})
	return Matrix3(value), publicError(err)
}

// ITRSToGCRSMatrix returns the inverse-direction C row-major rotation matrix.
func ITRSToGCRSMatrix(scales TimeScales) (Matrix3, error) {
	value, err := native.ITRSToGCRSMatrix(native.TimeScales{
		JDWhole: scales.JDWhole, UT1Fraction: scales.UT1Fraction, TTFraction: scales.TTFraction,
		TDBFraction: scales.TDBFraction, JDUT1: scales.JDUT1, JDTT: scales.JDTT, JDTDB: scales.JDTDB,
	})
	return Matrix3(value), publicError(err)
}

// MatrixVectorMultiply multiplies a row-major 3x3 matrix by a 3-vector in C.
func MatrixVectorMultiply(matrix Matrix3, vector [3]float64) ([3]float64, error) {
	value, err := native.MatrixVectorMultiply(matrix, vector)
	return value, publicError(err)
}

// GCRSToITRS transforms a GCRS position in kilometres to ITRS/ECEF
// kilometres. skyfieldCompatible selects the C compatibility path.
func GCRSToITRS(positionKm [3]float64, scales TimeScales, skyfieldCompatible bool) ([3]float64, error) {
	value, err := native.GCRSToITRS(positionKm, native.TimeScales{
		JDWhole: scales.JDWhole, UT1Fraction: scales.UT1Fraction, TTFraction: scales.TTFraction,
		TDBFraction: scales.TDBFraction, JDUT1: scales.JDUT1, JDTT: scales.JDTT, JDTDB: scales.JDTDB,
	}, skyfieldCompatible)
	return value, publicError(err)
}

// ITRSToGCRS transforms an ITRS/ECEF position in kilometres to GCRS
// kilometres.
func ITRSToGCRS(positionKm [3]float64, scales TimeScales) ([3]float64, error) {
	value, err := native.ITRSToGCRS(positionKm, native.TimeScales{
		JDWhole: scales.JDWhole, UT1Fraction: scales.UT1Fraction, TTFraction: scales.TTFraction,
		TDBFraction: scales.TDBFraction, JDUT1: scales.JDUT1, JDTT: scales.JDTT, JDTDB: scales.JDTDB,
	})
	return value, publicError(err)
}

// GCRSToTopocentric converts a GCRS satellite position to station-relative
// azimuth, elevation, and range.
func GCRSToTopocentric(positionKm [3]float64, station GroundStation, scales TimeScales, skyfieldCompatible bool) (Topocentric, error) {
	value, err := native.GCRSToTopocentric(
		positionKm, station.LatitudeDeg, station.LongitudeDeg, station.AltitudeKm,
		native.TimeScales{
			JDWhole: scales.JDWhole, UT1Fraction: scales.UT1Fraction, TTFraction: scales.TTFraction,
			TDBFraction: scales.TDBFraction, JDUT1: scales.JDUT1, JDTT: scales.JDTT, JDTDB: scales.JDTDB,
		}, skyfieldCompatible,
	)
	return Topocentric{AzimuthDeg: value[0], ElevationDeg: value[1], RangeKm: value[2]}, publicError(err)
}

// DopplerRangeRate is the C range-rate and dimensionless Doppler-ratio result.
type DopplerRangeRate struct {
	RangeRateKmS float64
	DopplerRatio float64
}

// DopplerShift is the C range-rate, carrier Doppler shift, and dimensionless
// Doppler-ratio result.
type DopplerShift struct {
	RangeRateKmS float64
	DopplerHz    float64
	DopplerRatio float64
}

// ComputeDopplerRangeRate computes range rate and Doppler ratio from a GCRS
// state. Position is kilometres, velocity is kilometres per second, and the
// station uses GroundStation units.
func ComputeDopplerRangeRate(positionKm, velocityKmS [3]float64, station GroundStation, scales TimeScales) (DopplerRangeRate, error) {
	value, err := native.ComputeDopplerRangeRate(
		positionKm, velocityKmS, station.LatitudeDeg, station.LongitudeDeg, station.AltitudeKm,
		native.TimeScales{
			JDWhole: scales.JDWhole, UT1Fraction: scales.UT1Fraction, TTFraction: scales.TTFraction,
			TDBFraction: scales.TDBFraction, JDUT1: scales.JDUT1, JDTT: scales.JDTT, JDTDB: scales.JDTDB,
		},
	)
	return DopplerRangeRate{RangeRateKmS: value.RangeRateKmS, DopplerRatio: value.DopplerRatio}, publicError(err)
}

// ComputeDopplerShift computes range rate, Doppler shift in hertz, and ratio
// for a carrier frequency in hertz.
func ComputeDopplerShift(positionKm, velocityKmS [3]float64, station GroundStation, scales TimeScales, frequencyHz float64) (DopplerShift, error) {
	value, err := native.ComputeDopplerShift(
		positionKm, velocityKmS, station.LatitudeDeg, station.LongitudeDeg, station.AltitudeKm,
		native.TimeScales{
			JDWhole: scales.JDWhole, UT1Fraction: scales.UT1Fraction, TTFraction: scales.TTFraction,
			TDBFraction: scales.TDBFraction, JDUT1: scales.JDUT1, JDTT: scales.JDTT, JDTDB: scales.JDTDB,
		}, frequencyHz,
	)
	return DopplerShift{RangeRateKmS: value.RangeRateKmS, DopplerHz: value.DopplerHz, DopplerRatio: value.DopplerRatio}, publicError(err)
}
