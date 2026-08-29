package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// Geodetic is a WGS84 position. Latitude and longitude are radians; height is
// ellipsoidal height in metres.
type Geodetic struct {
	LatitudeRad  float64
	LongitudeRad float64
	HeightM      float64
}

// ECEF is an ITRF/WGS84 Earth-fixed position in metres.
type ECEF struct {
	X float64
	Y float64
	Z float64
}

// LineOfSight is an ECEF unit vector from a receiver towards a satellite.
type LineOfSight struct {
	EX float64
	EY float64
	EZ float64
}

// DOP contains geometric dilution-of-precision scalars.
type DOP struct {
	GDOP float64
	PDOP float64
	HDOP float64
	VDOP float64
	TDOP float64
}

// ENUConvention selects the local horizontal/vertical split used by DOP.
type ENUConvention uint32

const (
	// GeodeticNormal uses the ellipsoid-normal up direction.
	GeodeticNormal ENUConvention = 0
	// GeocentricRadial uses the radial up direction.
	GeocentricRadial ENUConvention = 1
)

// GeodeticToECEF converts a WGS84 geodetic position to ECEF metres.
func GeodeticToECEF(position Geodetic) (ECEF, error) {
	value, err := native.GeodeticToECEF(native.Geodetic{
		LatitudeRad: position.LatitudeRad, LongitudeRad: position.LongitudeRad, HeightM: position.HeightM,
	})
	return ECEF{X: value.X, Y: value.Y, Z: value.Z}, publicError(err)
}

// ToECEF converts the geodetic position to ECEF metres.
func (position Geodetic) ToECEF() (ECEF, error) { return GeodeticToECEF(position) }

// ECEFToGeodetic converts an ECEF position in metres to WGS84 geodetic
// latitude/longitude radians and ellipsoidal height metres.
func ECEFToGeodetic(position ECEF) (Geodetic, error) {
	value, err := native.ECEFToGeodetic(native.ECEF{X: position.X, Y: position.Y, Z: position.Z})
	return Geodetic{LatitudeRad: value.LatitudeRad, LongitudeRad: value.LongitudeRad, HeightM: value.HeightM}, publicError(err)
}

// ToGeodetic converts the ECEF position to WGS84 geodetic coordinates.
func (position ECEF) ToGeodetic() (Geodetic, error) { return ECEFToGeodetic(position) }

// LineOfSightFromAzEl constructs an ECEF line-of-sight unit vector. Azimuth
// and elevation are degrees; azimuth is clockwise from geodetic north.
func LineOfSightFromAzEl(azimuthDeg, elevationDeg float64, receiver Geodetic) (LineOfSight, error) {
	value, err := native.LineOfSightFromAzEl(
		azimuthDeg, elevationDeg,
		native.Geodetic{LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM},
	)
	return LineOfSight{EX: value.EX, EY: value.EY, EZ: value.EZ}, publicError(err)
}

// LineOfSightFromAzimuthElevation is the descriptive form of
// LineOfSightFromAzEl.
func LineOfSightFromAzimuthElevation(azimuthDeg, elevationDeg float64, receiver Geodetic) (LineOfSight, error) {
	return LineOfSightFromAzEl(azimuthDeg, elevationDeg, receiver)
}

// ComputeDOP computes DOP from ECEF line-of-sight rows and inverse-variance
// weights. The slices must have equal lengths; C performs all domain checks.
func ComputeDOP(los []LineOfSight, weights []float64, receiver Geodetic) (DOP, error) {
	return computeDOP(los, weights, receiver, GeodeticNormal)
}

// DOPFromLineOfSight is an alias for ComputeDOP.
func DOPFromLineOfSight(los []LineOfSight, weights []float64, receiver Geodetic) (DOP, error) {
	return ComputeDOP(los, weights, receiver)
}

// ComputeDOPWithConvention computes DOP with an explicit ENU convention.
func ComputeDOPWithConvention(los []LineOfSight, weights []float64, receiver Geodetic, convention ENUConvention) (DOP, error) {
	return computeDOP(los, weights, receiver, convention)
}

func computeDOP(los []LineOfSight, weights []float64, receiver Geodetic, convention ENUConvention) (DOP, error) {
	if len(los) != len(weights) {
		return DOP{}, errors.New("sidereon: line-of-sight and weight lengths differ")
	}
	nativeLOS := make([]native.LineOfSight, len(los))
	for i, value := range los {
		nativeLOS[i] = native.LineOfSight{EX: value.EX, EY: value.EY, EZ: value.EZ}
	}
	value, err := native.DOP(nativeLOS, weights, native.Geodetic{
		LatitudeRad: receiver.LatitudeRad, LongitudeRad: receiver.LongitudeRad, HeightM: receiver.HeightM,
	}, uint32(convention))
	return DOP{GDOP: value.GDOP, PDOP: value.PDOP, HDOP: value.HDOP, VDOP: value.VDOP, TDOP: value.TDOP}, publicError(err)
}
