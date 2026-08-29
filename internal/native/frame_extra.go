//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

// FrameGeodeticFromECEF is the degree/metre result of the PROJ-compatible
// Earth-fixed conversion. The fields follow the C ABI's longitude,
// latitude, altitude ordering.
type FrameGeodeticFromECEF struct {
	LongitudeDeg float64
	LatitudeDeg  float64
	AltitudeM    float64
}

// FrameGeodeticFromITRS is the degree/kilometre result of the ITRS
// conversion. The fields follow the C ABI's latitude, longitude, altitude
// ordering.
type FrameGeodeticFromITRS struct {
	LatitudeDeg  float64
	LongitudeDeg float64
	AltitudeKm   float64
}

// FrameTEMEToGCRS is a copied position/velocity pair in kilometres and
// kilometres per second.
type FrameTEMEToGCRSResult struct {
	PositionKm     [3]float64
	VelocityKmPerS [3]float64
}

func FrameGASTRadians(scales TimeScales) (float64, error) {
	ts := cTimeScales(scales)
	var out C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_gast_radians(&ts, &out))
	})
	return float64(out), err
}

func FrameGCRSToITRSMatrixWithPolarMotion(scales TimeScales, xpArcsec, ypArcsec float64) (matrix3, error) {
	ts := cTimeScales(scales)
	var out [9]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_gcrs_to_itrs_matrix_with_polar_motion(&ts, C.double(xpArcsec), C.double(ypArcsec), &out[0]))
	})
	if err != nil {
		return matrix3{}, err
	}
	return flatToMatrix(&out), nil
}

func FrameGCRSToITRSWithPolarMotion(positionKm [3]float64, scales TimeScales, skyfieldCompatible bool, xpArcsec, ypArcsec float64) ([3]float64, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	ts := cTimeScales(scales)
	var out [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_gcrs_to_itrs_with_polar_motion(&position[0], &ts, C.bool(skyfieldCompatible), C.double(xpArcsec), C.double(ypArcsec), &out[0]))
	})
	return [3]float64{float64(out[0]), float64(out[1]), float64(out[2])}, err
}

func FrameGeodeticFromECEFProj(ecefM [3]float64) (FrameGeodeticFromECEF, error) {
	ecef := [3]C.double{C.double(ecefM[0]), C.double(ecefM[1]), C.double(ecefM[2])}
	var out [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_geodetic_from_ecef_proj(&ecef[0], &out[0]))
	})
	return FrameGeodeticFromECEF{LongitudeDeg: float64(out[0]), LatitudeDeg: float64(out[1]), AltitudeM: float64(out[2])}, err
}

func FrameGeodeticToITRS(latitudeDeg, longitudeDeg, altitudeKm float64) ([3]float64, error) {
	var out [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_geodetic_to_itrs(C.double(latitudeDeg), C.double(longitudeDeg), C.double(altitudeKm), &out[0]))
	})
	return [3]float64{float64(out[0]), float64(out[1]), float64(out[2])}, err
}

func FrameGMSTRadians(scales TimeScales) (float64, error) {
	ts := cTimeScales(scales)
	var out C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_gmst_radians(&ts, &out))
	})
	return float64(out), err
}

func FrameITRSToGCRSMatrixWithPolarMotion(scales TimeScales, xpArcsec, ypArcsec float64) (matrix3, error) {
	ts := cTimeScales(scales)
	var out [9]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_itrs_to_gcrs_matrix_with_polar_motion(&ts, C.double(xpArcsec), C.double(ypArcsec), &out[0]))
	})
	if err != nil {
		return matrix3{}, err
	}
	return flatToMatrix(&out), nil
}

func FrameITRSToGCRSWithPolarMotion(positionKm [3]float64, scales TimeScales, xpArcsec, ypArcsec float64) ([3]float64, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	ts := cTimeScales(scales)
	var out [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_itrs_to_gcrs_with_polar_motion(&position[0], &ts, C.double(xpArcsec), C.double(ypArcsec), &out[0]))
	})
	return [3]float64{float64(out[0]), float64(out[1]), float64(out[2])}, err
}

func FrameITRSToGeodetic(positionKm [3]float64) (FrameGeodeticFromITRS, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	var out [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_itrs_to_geodetic(&position[0], &out[0]))
	})
	return FrameGeodeticFromITRS{LatitudeDeg: float64(out[0]), LongitudeDeg: float64(out[1]), AltitudeKm: float64(out[2])}, err
}

func FrameMeanOfDateToITRSMatrix(scales TimeScales) (matrix3, error) {
	ts := cTimeScales(scales)
	var out [9]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_mean_of_date_to_itrs_matrix(&ts, &out[0]))
	})
	if err != nil {
		return matrix3{}, err
	}
	return flatToMatrix(&out), nil
}

func FrameMeanOfDateToITRSMatrixWithPolarMotion(scales TimeScales, xpArcsec, ypArcsec float64) (matrix3, error) {
	ts := cTimeScales(scales)
	var out [9]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_mean_of_date_to_itrs_matrix_with_polar_motion(&ts, C.double(xpArcsec), C.double(ypArcsec), &out[0]))
	})
	if err != nil {
		return matrix3{}, err
	}
	return flatToMatrix(&out), nil
}

func FrameTEMEToGCRS(positionKm, velocityKmS [3]float64, scales TimeScales, skyfieldCompatible bool) (FrameTEMEToGCRSResult, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	velocity := [3]C.double{C.double(velocityKmS[0]), C.double(velocityKmS[1]), C.double(velocityKmS[2])}
	ts := cTimeScales(scales)
	var outPosition, outVelocity [3]C.double
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_frame_teme_to_gcrs(&position[0], &velocity[0], &ts, C.bool(skyfieldCompatible), &outPosition[0], &outVelocity[0]))
	})
	return FrameTEMEToGCRSResult{
		PositionKm:     [3]float64{float64(outPosition[0]), float64(outPosition[1]), float64(outPosition[2])},
		VelocityKmPerS: [3]float64{float64(outVelocity[0]), float64(outVelocity[1]), float64(outVelocity[2])},
	}, err
}
