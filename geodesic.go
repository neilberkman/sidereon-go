package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// GeodesicDirectResult is a WGS84 direct-geodesic result. Angles are degrees
// and distance is metres.
type GeodesicDirectResult struct {
	// LatitudeDeg is the latitude deg in degrees.
	LatitudeDeg float64
	// LongitudeDeg is the longitude deg in degrees.
	LongitudeDeg float64
	// FinalAzimuthDeg is the final azimuth deg in degrees.
	FinalAzimuthDeg float64
}

// GeodesicInverseResult is a WGS84 inverse-geodesic result. Angles are degrees
// and distance is metres.
type GeodesicInverseResult struct {
	// DistanceM is the distance m in metres.
	DistanceM float64
	// InitialAzimuthDeg is the initial azimuth deg in degrees.
	InitialAzimuthDeg float64
	// FinalAzimuthDeg is the final azimuth deg in degrees.
	FinalAzimuthDeg float64
}

// GeodesicDirect solves the WGS84 direct geodesic from a point, azimuth, and
// distance.
func GeodesicDirect(latitudeDeg, longitudeDeg, initialAzimuthDeg, distanceM float64) (GeodesicDirectResult, error) {
	value, err := native.GeodesicDirect(latitudeDeg, longitudeDeg, initialAzimuthDeg, distanceM)
	return GeodesicDirectResult{
		LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg, FinalAzimuthDeg: value.FinalAzimuthDeg,
	}, publicError(err)
}

// GeodesicInverse solves the WGS84 inverse geodesic between two points.
func GeodesicInverse(latitude1Deg, longitude1Deg, latitude2Deg, longitude2Deg float64) (GeodesicInverseResult, error) {
	value, err := native.GeodesicInverse(latitude1Deg, longitude1Deg, latitude2Deg, longitude2Deg)
	return GeodesicInverseResult{
		DistanceM: value.DistanceM, InitialAzimuthDeg: value.InitialAzimuthDeg, FinalAzimuthDeg: value.FinalAzimuthDeg,
	}, publicError(err)
}
