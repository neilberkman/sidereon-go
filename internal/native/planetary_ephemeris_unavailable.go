//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "time"

func DEEphemerisAstronomicalUnitM(_ *DEEphemeris) (float64, error) { return 0, ErrUnavailable }

func DEEphemerisConstantCount(_ *DEEphemeris) (uint64, error) { return 0, ErrUnavailable }

func DEEphemerisCoverageTDBSeconds(_ *DEEphemeris) (DECoverage, error) {
	return DECoverage{}, ErrUnavailable
}

func DEEphemerisEarthMoonMassRatio(_ *DEEphemeris) (float64, error) { return 0, ErrUnavailable }

func DEEphemerisFromBytes(_ []byte) (*DEEphemeris, error) { return nil, ErrUnavailable }

func DEEphemerisGmM3S2(_ *DEEphemeris, _ DEBody) (DEGM, error) { return DEGM{}, ErrUnavailable }

func DEEphemerisLoad(_ string) (*DEEphemeris, error) { return nil, ErrUnavailable }

func DEEphemerisRecordSize(_ *DEEphemeris) (uint64, error) { return 0, ErrUnavailable }

func DEEphemerisSunMoonAtTDBSeconds(_ *DEEphemeris, _ float64) (DESunMoon, error) {
	return DESunMoon{}, ErrUnavailable
}

func DEEphemerisSunMoonPositionsAtTDBSeconds(_ *DEEphemeris, _ float64) (SunMoon, error) {
	return SunMoon{}, ErrUnavailable
}

func DEEphemerisVersion(_ *DEEphemeris) (uint32, error) { return 0, ErrUnavailable }

func FrameGCRSToMeanOfDateMatrix(_ TimeScales) ([3][3]float64, error) {
	return [3][3]float64{}, ErrUnavailable
}

func FrameMeanOfDateToGCRSMatrix(_ TimeScales) ([3][3]float64, error) {
	return [3][3]float64{}, ErrUnavailable
}

func SunMoonGCRS(_ float64) (SunMoon, error) { return SunMoon{}, ErrUnavailable }

func SunMoonGCRSAt(_ time.Time) (SunMoon, error) { return SunMoon{}, ErrUnavailable }

func SunMoonGCRSAtTDBSeconds(_ float64) (SunMoon, error) { return SunMoon{}, ErrUnavailable }

func TDBFromTTJ2000Seconds(_ float64) float64 { return 0 }

func TDBMinusTTSeconds(_ float64) float64 { return 0 }

func TTFromTDBJ2000Seconds(_ float64) float64 { return 0 }
