package sidereon

import (
	"sidereon.dev/go/v4/internal/native"
	"time"
)

// AstronomicalUnitM returns the astronomical unit m.
func (ephemeris *DEEphemeris) AstronomicalUnitM() (float64, error) {
	value, err := native.DEEphemerisAstronomicalUnitM(nativeDEEphemerisHandle(ephemeris))
	return value, publicError(err)
}

// ConstantCount returns the constant count.
func (ephemeris *DEEphemeris) ConstantCount() (uint64, error) {
	value, err := native.DEEphemerisConstantCount(nativeDEEphemerisHandle(ephemeris))
	return value, publicError(err)
}

// CoverageTDBSeconds returns the coverage TDB seconds.
func (ephemeris *DEEphemeris) CoverageTDBSeconds() (DECoverage, error) {
	value, err := native.DEEphemerisCoverageTDBSeconds(nativeDEEphemerisHandle(ephemeris))
	return publicDECoverage(value), publicError(err)
}

// EarthMoonMassRatio returns the earth moon mass ratio.
func (ephemeris *DEEphemeris) EarthMoonMassRatio() (float64, error) {
	value, err := native.DEEphemerisEarthMoonMassRatio(nativeDEEphemerisHandle(ephemeris))
	return value, publicError(err)
}

// DEEphemerisFromBytes parses binary planetary ephemeris records and owns the resulting data.
func DEEphemerisFromBytes(data []byte) (*DEEphemeris, error) {
	value, err := native.DEEphemerisFromBytes(data)
	if err != nil {
		return nil, publicError(err)
	}
	return publicDEEphemeris(value), nil
}

// GMM3S2 returns the gm m3 s2.
func (ephemeris *DEEphemeris) GMM3S2(body DEBody) (DEGM, error) {
	value, err := native.DEEphemerisGmM3S2(nativeDEEphemerisHandle(ephemeris), native.DEBody(body))
	return publicDEGM(value), publicError(err)
}

// LoadDEEphemeris opens and parses a binary JPL development ephemeris.
func LoadDEEphemeris(path string) (*DEEphemeris, error) {
	value, err := native.DEEphemerisLoad(path)
	if err != nil {
		return nil, publicError(err)
	}
	return publicDEEphemeris(value), nil
}

// RecordSize returns the record size.
func (ephemeris *DEEphemeris) RecordSize() (uint64, error) {
	value, err := native.DEEphemerisRecordSize(nativeDEEphemerisHandle(ephemeris))
	return value, publicError(err)
}

// SunMoonAtTDBSeconds returns the sun moon at TDB seconds.
func (ephemeris *DEEphemeris) SunMoonAtTDBSeconds(epochTDBSeconds float64) (DESunMoon, error) {
	value, err := native.DEEphemerisSunMoonAtTDBSeconds(nativeDEEphemerisHandle(ephemeris), epochTDBSeconds)
	return publicDESunMoon(value), publicError(err)
}

// SunMoonPositionsAtTDBSeconds returns the sun moon positions at TDB seconds.
func (ephemeris *DEEphemeris) SunMoonPositionsAtTDBSeconds(epochTDBSeconds float64) (SunMoon, error) {
	value, err := native.DEEphemerisSunMoonPositionsAtTDBSeconds(nativeDEEphemerisHandle(ephemeris), epochTDBSeconds)
	return publicSunMoon(value), publicError(err)
}

// Version returns the version.
func (ephemeris *DEEphemeris) Version() (uint32, error) {
	value, err := native.DEEphemerisVersion(nativeDEEphemerisHandle(ephemeris))
	return value, publicError(err)
}

// GCRSToMeanOfDateMatrix returns the frame GCRS to mean of date matrix.
func GCRSToMeanOfDateMatrix(ts TimeScales) ([3][3]float64, error) {
	value, err := native.FrameGCRSToMeanOfDateMatrix(nativeTimeScales(ts))
	return value, publicError(err)
}

// MeanOfDateToGCRSMatrix returns the frame mean of date to GCRS matrix.
func MeanOfDateToGCRSMatrix(ts TimeScales) ([3][3]float64, error) {
	value, err := native.FrameMeanOfDateToGCRSMatrix(nativeTimeScales(ts))
	return value, publicError(err)
}

// SunMoonGCRS returns the sun moon GCRS.
func SunMoonGCRS(ttJulianCenturies float64) (SunMoon, error) {
	value, err := native.SunMoonGCRS(ttJulianCenturies)
	return publicSunMoon(value), publicError(err)
}

// SunMoonGCRSAt returns the sun moon GCRS at.
func SunMoonGCRSAt(epoch time.Time) (SunMoon, error) {
	value, err := native.SunMoonGCRSAt(epoch)
	return publicSunMoon(value), publicError(err)
}

// SunMoonGCRSAtTDBSeconds returns the sun moon GCRS at TDB seconds.
func SunMoonGCRSAtTDBSeconds(epochTDBSeconds float64) (SunMoon, error) {
	value, err := native.SunMoonGCRSAtTDBSeconds(epochTDBSeconds)
	return publicSunMoon(value), publicError(err)
}

// TDBFromTTJ2000Seconds returns the TDB from TT j2000 seconds.
func TDBFromTTJ2000Seconds(ttJ2000S float64) float64 {
	return native.TDBFromTTJ2000Seconds(ttJ2000S)
}

// TDBMinusTTSeconds returns the TDB minus TT seconds.
func TDBMinusTTSeconds(ttJ2000S float64) float64 {
	return native.TDBMinusTTSeconds(ttJ2000S)
}

// TTFromTDBJ2000Seconds returns the TT from TDB j2000 seconds.
func TTFromTDBJ2000Seconds(tdbJ2000S float64) float64 {
	return native.TTFromTDBJ2000Seconds(tdbJ2000S)
}
