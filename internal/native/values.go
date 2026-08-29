//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import "unsafe"

type Geodetic struct {
	LatitudeRad  float64
	LongitudeRad float64
	HeightM      float64
}

type ECEF struct {
	X float64
	Y float64
	Z float64
}

type LineOfSight struct {
	EX float64
	EY float64
	EZ float64
}

type Dop struct {
	GDOP float64
	PDOP float64
	HDOP float64
	VDOP float64
	TDOP float64
}

type GeodesicDirectResult struct {
	LatitudeDeg     float64
	LongitudeDeg    float64
	FinalAzimuthDeg float64
}

type GeodesicInverseResult struct {
	DistanceM         float64
	InitialAzimuthDeg float64
	FinalAzimuthDeg   float64
}

type ErrorEllipse2 struct {
	Confidence     float64
	ChiSquareScale float64
	SemiMajor      float64
	SemiMinor      float64
	OrientationRad float64
}

type ErrorEllipse struct {
	SemiMajorM     float64
	SemiMinorM     float64
	OrientationRad float64
}

type PercentileRadius struct {
	Probability float64
	RadiusM     float64
	ApproxM     float64
	ApproxValid bool
}

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

type TimeScales struct {
	JDWhole     float64
	UT1Fraction float64
	TTFraction  float64
	TDBFraction float64
	JDUT1       float64
	JDTT        float64
	JDTDB       float64
}

type DopplerRangeRate struct {
	RangeRateKmS float64
	DopplerRatio float64
}

type DopplerShift struct {
	RangeRateKmS float64
	DopplerHz    float64
	DopplerRatio float64
}

type CivilDateTime struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second float64
}

type JulianDate struct {
	Whole    float64
	Fraction float64
}

type GNSSWeekSeconds struct {
	Week          float64
	SecondsOfWeek float64
}

type NISGate struct {
	NIS       float64
	Threshold float64
	InGate    bool
	DOF       uint64
}

type CarrierPair struct {
	Band1 int
	Band2 int
}

type matrix3 [3][3]float64

func cGeodetic(value Geodetic) C.SidereonGeodetic {
	return C.SidereonGeodetic{
		lat_rad:  C.double(value.LatitudeRad),
		lon_rad:  C.double(value.LongitudeRad),
		height_m: C.double(value.HeightM),
	}
}

func geodeticFromC(value C.SidereonGeodetic) Geodetic {
	return Geodetic{
		LatitudeRad:  float64(value.lat_rad),
		LongitudeRad: float64(value.lon_rad),
		HeightM:      float64(value.height_m),
	}
}

func ecefFromC(value C.SidereonItrfPosition) ECEF {
	return ECEF{X: float64(value.x_m), Y: float64(value.y_m), Z: float64(value.z_m)}
}

func matrixFromC(values *C.double) matrix3 {
	var out matrix3
	flat := (*[9]C.double)(unsafe.Pointer(values))
	for i := range out {
		for j := range out[i] {
			out[i][j] = float64(flat[i*3+j])
		}
	}
	return out
}

func matrixToC(values matrix3) [9]C.double {
	var out [9]C.double
	for i := range values {
		for j := range values[i] {
			out[i*3+j] = C.double(values[i][j])
		}
	}
	return out
}

func flatToMatrix(values *[9]C.double) matrix3 {
	var out matrix3
	for i := range out {
		for j := range out[i] {
			out[i][j] = float64(values[i*3+j])
		}
	}
	return out
}

func GeodeticToECEF(value Geodetic) (ECEF, error) {
	input := cGeodetic(value)
	var output C.SidereonItrfPosition
	err := callStatus(func() uint32 {
		return C.sidereon_geodetic_to_ecef(&input, &output)
	})
	if err != nil {
		return ECEF{}, err
	}
	return ecefFromC(output), nil
}

func ECEFToGeodetic(value ECEF) (Geodetic, error) {
	input := C.SidereonItrfPosition{x_m: C.double(value.X), y_m: C.double(value.Y), z_m: C.double(value.Z)}
	var output C.SidereonGeodetic
	err := callStatus(func() uint32 {
		return C.sidereon_ecef_to_geodetic(&input, &output)
	})
	if err != nil {
		return Geodetic{}, err
	}
	return geodeticFromC(output), nil
}

func LineOfSightFromAzEl(azimuthDeg, elevationDeg float64, receiver Geodetic) (LineOfSight, error) {
	var output C.SidereonLineOfSight
	input := cGeodetic(receiver)
	err := callStatus(func() uint32 {
		return C.sidereon_line_of_sight_from_az_el_deg(
			C.double(azimuthDeg), C.double(elevationDeg), input, &output,
		)
	})
	if err != nil {
		return LineOfSight{}, err
	}
	return LineOfSight{EX: float64(output.e_x), EY: float64(output.e_y), EZ: float64(output.e_z)}, nil
}

func DOP(los []LineOfSight, weights []float64, receiver Geodetic, convention uint32) (Dop, error) {
	cLos := make([]C.SidereonLineOfSight, len(los))
	for i, value := range los {
		cLos[i] = C.SidereonLineOfSight{e_x: C.double(value.EX), e_y: C.double(value.EY), e_z: C.double(value.EZ)}
	}
	cWeights := make([]C.double, len(weights))
	for i, value := range weights {
		cWeights[i] = C.double(value)
	}
	var losPointer *C.SidereonLineOfSight
	if len(cLos) != 0 {
		losPointer = &cLos[0]
	}
	var weightPointer *C.double
	if len(cWeights) != 0 {
		weightPointer = &cWeights[0]
	}
	input := cGeodetic(receiver)
	var output C.SidereonDop
	err := callStatus(func() uint32 {
		if convention == 0 {
			return C.sidereon_dop(losPointer, weightPointer, C.size_t(len(cLos)), input, &output)
		}
		return C.sidereon_dop_with_convention(
			losPointer, weightPointer, C.size_t(len(cLos)), input, C.uint32_t(convention), &output,
		)
	})
	if err != nil {
		return Dop{}, err
	}
	return Dop{
		GDOP: float64(output.gdop), PDOP: float64(output.pdop), HDOP: float64(output.hdop),
		VDOP: float64(output.vdop), TDOP: float64(output.tdop),
	}, nil
}

func GeodesicDirect(lat1Deg, lon1Deg, azimuthDeg, distanceM float64) (GeodesicDirectResult, error) {
	var output C.SidereonGeodesicDirectResult
	err := callStatus(func() uint32 {
		return C.sidereon_geodesic_direct(
			C.double(lat1Deg), C.double(lon1Deg), C.double(azimuthDeg), C.double(distanceM), &output,
		)
	})
	if err != nil {
		return GeodesicDirectResult{}, err
	}
	return GeodesicDirectResult{
		LatitudeDeg: float64(output.latitude_deg), LongitudeDeg: float64(output.longitude_deg),
		FinalAzimuthDeg: float64(output.final_azimuth_deg),
	}, nil
}

func GeodesicInverse(lat1Deg, lon1Deg, lat2Deg, lon2Deg float64) (GeodesicInverseResult, error) {
	var output C.SidereonGeodesicInverseResult
	err := callStatus(func() uint32 {
		return C.sidereon_geodesic_inverse(
			C.double(lat1Deg), C.double(lon1Deg), C.double(lat2Deg), C.double(lon2Deg), &output,
		)
	})
	if err != nil {
		return GeodesicInverseResult{}, err
	}
	return GeodesicInverseResult{
		DistanceM: float64(output.distance_m), InitialAzimuthDeg: float64(output.initial_azimuth_deg),
		FinalAzimuthDeg: float64(output.final_azimuth_deg),
	}, nil
}

func AngularSeparationCoords(aLonDeg, aLatDeg, bLonDeg, bLatDeg float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_angular_separation_coords_deg(
			C.double(aLonDeg), C.double(aLatDeg), C.double(bLonDeg), C.double(bLatDeg), &output,
		)
	})
	return float64(output), err
}

func AngularSeparation(a, b [3]float64) (float64, error) {
	ca := [3]C.double{C.double(a[0]), C.double(a[1]), C.double(a[2])}
	cb := [3]C.double{C.double(b[0]), C.double(b[1]), C.double(b[2])}
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_angular_separation_deg(&ca[0], &cb[0], &output)
	})
	return float64(output), err
}

func EarthAngularRadius(positionKm [3]float64) (float64, error) {
	input := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_earth_angular_radius_deg(&input[0], &output)
	})
	return float64(output), err
}

func EclipseShadowFraction(satelliteKm, sunKm [3]float64, model *uint32) (float64, error) {
	satellite := [3]C.double{C.double(satelliteKm[0]), C.double(satelliteKm[1]), C.double(satelliteKm[2])}
	sun := [3]C.double{C.double(sunKm[0]), C.double(sunKm[1]), C.double(sunKm[2])}
	var output C.double
	err := callStatus(func() uint32 {
		if model == nil {
			return C.sidereon_eclipse_shadow_fraction(&satellite[0], &sun[0], &output)
		}
		return C.sidereon_eclipse_shadow_fraction_with_model(
			&satellite[0], &sun[0], C.uint32_t(*model), &output,
		)
	})
	return float64(output), err
}

func EclipseStatus(satelliteKm, sunKm [3]float64, model *uint32) (int, error) {
	satellite := [3]C.double{C.double(satelliteKm[0]), C.double(satelliteKm[1]), C.double(satelliteKm[2])}
	sun := [3]C.double{C.double(sunKm[0]), C.double(sunKm[1]), C.double(sunKm[2])}
	var output uint32
	err := callStatus(func() uint32 {
		if model == nil {
			return C.sidereon_eclipse_status(&satellite[0], &sun[0], &output)
		}
		return C.sidereon_eclipse_status_with_model(
			&satellite[0], &sun[0], C.uint32_t(*model), &output,
		)
	})
	return int(output), err
}

func cCivil(value CivilDateTime) (C.int32_t, C.int32_t, C.int32_t, C.int32_t, C.int32_t, C.double) {
	return C.int32_t(value.Year), C.int32_t(value.Month), C.int32_t(value.Day),
		C.int32_t(value.Hour), C.int32_t(value.Minute), C.double(value.Second)
}

func CivilToJ2000(value CivilDateTime) (float64, error) {
	year, month, day, hour, minute, second := cCivil(value)
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_civil_to_j2000_seconds(year, month, day, hour, minute, second, &output)
	})
	return float64(output), err
}

func CivilToGPS(value CivilDateTime) (float64, bool, error) {
	year, month, day, hour, minute, second := cCivil(value)
	var output C.double
	var available C.bool
	err := callStatus(func() uint32 {
		return C.sidereon_civil_to_gps_seconds(
			C.int32_t(year), C.uint8_t(month), C.uint8_t(day), C.uint8_t(hour), C.uint8_t(minute),
			second, &output, &available,
		)
	})
	return float64(output), bool(available), err
}

func J2000ToCivil(seconds int64) (CivilDateTime, error) {
	var year, month, day, hour, minute, second C.int64_t
	err := callStatus(func() uint32 {
		return C.sidereon_j2000_seconds_to_civil(
			C.int64_t(seconds), &year, &month, &day, &hour, &minute, &second,
		)
	})
	if err != nil {
		return CivilDateTime{}, err
	}
	return CivilDateTime{
		Year: int(year), Month: int(month), Day: int(day), Hour: int(hour), Minute: int(minute), Second: float64(second),
	}, nil
}

func InstantFromUTC(value CivilDateTime) (JulianDate, float64, error) {
	year, month, day, hour, minute, second := cCivil(value)
	var whole, fraction, j2000 C.double
	err := callStatus(func() uint32 {
		return C.sidereon_instant_from_utc_civil(
			year, month, day, hour, minute, second, &whole, &fraction, &j2000,
		)
	})
	return JulianDate{Whole: float64(whole), Fraction: float64(fraction)}, float64(j2000), err
}

func SplitJulianToJ2000(date JulianDate) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_split_jd_to_j2000_seconds(C.double(date.Whole), C.double(date.Fraction), &output)
	})
	return float64(output), err
}

func LeapSeconds(year, month, day int) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_leap_seconds(C.int32_t(year), C.int32_t(month), C.int32_t(day), &output)
	})
	return float64(output), err
}

func GPSUTCOffset(julianDate float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_gps_utc_offset_s(C.double(julianDate), &output)
	})
	return float64(output), err
}

func TAIUTCOffset(julianDate float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_tai_utc_offset_s(C.double(julianDate), &output)
	})
	return float64(output), err
}

func GNSSSecondsOfWeek(value CivilDateTime) (float64, error) {
	year, month, day, hour, minute, second := cCivil(value)
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_gnss_seconds_of_week_from_calendar(
			C.int64_t(year), C.int64_t(month), C.int64_t(day), C.int64_t(hour), C.int64_t(minute), C.int64_t(second), &output,
		)
	})
	return float64(output), err
}

func GNSSWeekAndSeconds(continuousSeconds float64) (GNSSWeekSeconds, error) {
	var week, sow C.double
	err := callStatus(func() uint32 {
		return C.sidereon_gnss_week_and_seconds_of_week(C.double(continuousSeconds), &week, &sow)
	})
	return GNSSWeekSeconds{Week: float64(week), SecondsOfWeek: float64(sow)}, err
}

func TimeScalesFromUTC(value CivilDateTime) (TimeScales, error) {
	year, month, day, hour, minute, second := cCivil(value)
	var output C.SidereonTimeScales
	err := callStatus(func() uint32 {
		return C.sidereon_timescales_from_utc(year, month, day, hour, minute, second, &output)
	})
	if err != nil {
		return TimeScales{}, err
	}
	return TimeScales{
		JDWhole: float64(output.jd_whole), UT1Fraction: float64(output.ut1_fraction),
		TTFraction: float64(output.tt_fraction), TDBFraction: float64(output.tdb_fraction),
		JDUT1: float64(output.jd_ut1), JDTT: float64(output.jd_tt), JDTDB: float64(output.jd_tdb),
	}, nil
}

func TimeScaleOffset(from, to uint32) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_timescale_offset_s(C.uint32_t(from), C.uint32_t(to), &output)
	})
	return float64(output), err
}

func TimeScaleOffsetAt(from, to uint32, utcJulianDate float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_timescale_offset_at_s(C.uint32_t(from), C.uint32_t(to), C.double(utcJulianDate), &output)
	})
	return float64(output), err
}

func copyLabel(call func(*C.uint8_t, C.size_t, *C.size_t, *C.size_t) uint32) ([]byte, error) {
	return copyNativeBytes("native label", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.enum_SidereonStatus(call(out, length, written, required))
	})
}

func CarrierBandLabel(band uint32) ([]byte, error) {
	return copyLabel(func(out *C.uint8_t, len C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_carrier_band_label(C.uint32_t(band), out, len, written, required)
	})
}

func GNSSSystemLabel(system uint32) ([]byte, error) {
	return copyLabel(func(out *C.uint8_t, len C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_gnss_system_label(C.uint32_t(system), out, len, written, required)
	})
}

func TimeScaleLabel(scale uint32) ([]byte, error) {
	return copyLabel(func(out *C.uint8_t, len C.size_t, written, required *C.size_t) uint32 {
		return C.sidereon_time_scale_abbrev(C.uint32_t(scale), out, len, written, required)
	})
}

func Frequency(system, band uint32) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frequency_hz(C.uint32_t(system), C.uint32_t(band), &output)
	})
	return float64(output), err
}

func Wavelength(system, band uint32) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_wavelength_m(C.uint32_t(system), C.uint32_t(band), &output)
	})
	return float64(output), err
}

func GLONASSG1Frequency(channel int8) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_glonass_g1_frequency_hz(C.int8_t(channel), &output)
	})
	return float64(output), err
}

func DefaultSPPFrequency(system uint32) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_default_spp_frequency_hz(C.uint32_t(system), &output)
	})
	return float64(output), err
}

func DefaultIonosphereFreePair(system uint32) (CarrierPair, bool, error) {
	var output C.SidereonCarrierPair
	var present C.bool
	err := callStatus(func() uint32 {
		return C.sidereon_default_iono_free_pair(C.uint32_t(system), &output, &present)
	})
	return CarrierPair{Band1: int(output.band1), Band2: int(output.band2)}, bool(present), err
}

func PhaseMeters(cycles, frequencyHz float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_carrier_phase_meters(C.double(cycles), C.double(frequencyHz), &output)
	})
	return float64(output), err
}

func CodeMinusCarrier(pMeters, phaseCycles, frequencyHz float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_carrier_code_minus_carrier(C.double(pMeters), C.double(phaseCycles), C.double(frequencyHz), &output)
	})
	return float64(output), err
}

func GeometryFree(l1Meters, l2Meters float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_carrier_geometry_free(C.double(l1Meters), C.double(l2Meters), &output)
	})
	return float64(output), err
}

func MelbourneWubbena(phi1, phi2, p1, p2, f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_carrier_melbourne_wubbena(
			C.double(phi1), C.double(phi2), C.double(p1), C.double(p2), C.double(f1), C.double(f2), &output,
		)
	})
	return float64(output), err
}

func NarrowLaneCode(p1, p2, f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_carrier_narrow_lane_code(C.double(p1), C.double(p2), C.double(f1), C.double(f2), &output)
	})
	return float64(output), err
}

func WideLaneCycles(phi1, phi2, p1, p2, f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_carrier_wide_lane_cycles(
			C.double(phi1), C.double(phi2), C.double(p1), C.double(p2), C.double(f1), C.double(f2), &output,
		)
	})
	return float64(output), err
}

func WideLaneWavelength(f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_carrier_wide_lane_wavelength(C.double(f1), C.double(f2), &output)
	})
	return float64(output), err
}

func IonosphericGamma(f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_combination_gamma(C.double(f1), C.double(f2), &output)
	})
	return float64(output), err
}

func CombinationNoiseAmplification(f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_combination_noise_amplification(C.double(f1), C.double(f2), &output)
	})
	return float64(output), err
}

func IonosphereFreeCode(obs1, obs2, f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_combination_ionosphere_free(C.double(obs1), C.double(obs2), C.double(f1), C.double(f2), &output)
	})
	return float64(output), err
}

func IonosphereFreePhaseCycles(phi1, phi2, f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_combination_ionosphere_free_phase_cycles(
			C.double(phi1), C.double(phi2), C.double(f1), C.double(f2), &output,
		)
	})
	return float64(output), err
}

func IonosphereFreePhaseMeters(phase1, phase2, f1, f2 float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_combination_ionosphere_free_phase_m(
			C.double(phase1), C.double(phase2), C.double(f1), C.double(f2), &output,
		)
	})
	return float64(output), err
}

func ComputeDopplerRangeRate(positionKm, velocityKmS [3]float64, stationLatDeg, stationLonDeg, stationAltKm float64, scales TimeScales) (DopplerRangeRate, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	velocity := [3]C.double{C.double(velocityKmS[0]), C.double(velocityKmS[1]), C.double(velocityKmS[2])}
	timescales := cTimeScales(scales)
	var output C.SidereonDopplerRangeRate
	err := callStatus(func() uint32 {
		return C.sidereon_doppler_range_rate_and_ratio(
			&position[0], &velocity[0], C.double(stationLatDeg), C.double(stationLonDeg), C.double(stationAltKm), &timescales, &output,
		)
	})
	return DopplerRangeRate{RangeRateKmS: float64(output.range_rate_km_s), DopplerRatio: float64(output.doppler_ratio)}, err
}

func ComputeDopplerShift(positionKm, velocityKmS [3]float64, stationLatDeg, stationLonDeg, stationAltKm float64, scales TimeScales, frequencyHz float64) (DopplerShift, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	velocity := [3]C.double{C.double(velocityKmS[0]), C.double(velocityKmS[1]), C.double(velocityKmS[2])}
	timescales := cTimeScales(scales)
	var output C.SidereonDopplerShift
	err := callStatus(func() uint32 {
		return C.sidereon_doppler_shift(
			&position[0], &velocity[0], C.double(stationLatDeg), C.double(stationLonDeg), C.double(stationAltKm), &timescales, C.double(frequencyHz), &output,
		)
	})
	return DopplerShift{RangeRateKmS: float64(output.range_rate_km_s), DopplerHz: float64(output.doppler_hz), DopplerRatio: float64(output.doppler_ratio)}, err
}

func cTimeScales(value TimeScales) C.SidereonTimeScales {
	return C.SidereonTimeScales{
		jd_whole: C.double(value.JDWhole), ut1_fraction: C.double(value.UT1Fraction),
		tt_fraction: C.double(value.TTFraction), tdb_fraction: C.double(value.TDBFraction),
		jd_ut1: C.double(value.JDUT1), jd_tt: C.double(value.JDTT), jd_tdb: C.double(value.JDTDB),
	}
}

func ErrorEllipse2x2(covariance [4]float64, confidence float64) (ErrorEllipse2, error) {
	input := [4]C.double{C.double(covariance[0]), C.double(covariance[1]), C.double(covariance[2]), C.double(covariance[3])}
	var output C.SidereonErrorEllipse2
	err := callStatus(func() uint32 {
		return C.sidereon_error_ellipse_2x2(&input[0], C.double(confidence), &output)
	})
	return ErrorEllipse2{
		Confidence: float64(output.confidence), ChiSquareScale: float64(output.chi_square_scale),
		SemiMajor: float64(output.semi_major), SemiMinor: float64(output.semi_minor),
		OrientationRad: float64(output.orientation_rad),
	}, err
}

func cMatrix3(value matrix3) [9]C.double { return matrixToC(value) }

func percentileFromC(value C.SidereonPercentileRadius) PercentileRadius {
	return PercentileRadius{Probability: float64(value.probability), RadiusM: float64(value.radius_m), ApproxM: float64(value.approx_m), ApproxValid: bool(value.approx_valid)}
}

func metricsFromC(value C.SidereonPositionErrorMetrics) PositionErrorMetrics {
	return PositionErrorMetrics{
		Ellipse: ErrorEllipse{SemiMajorM: float64(value.ellipse.semi_major_m), SemiMinorM: float64(value.ellipse.semi_minor_m), OrientationRad: float64(value.ellipse.orientation_rad)},
		SigmaEM: float64(value.sigma_e_m), SigmaNM: float64(value.sigma_n_m), SigmaUM: float64(value.sigma_u_m),
		CEP: percentileFromC(value.cep_m), R95: percentileFromC(value.r95_m), R99: percentileFromC(value.r99_m),
		DRMS: float64(value.drms_m), TwoDRMS: float64(value.two_drms_m), VEP: float64(value.vep_m),
		SEP: percentileFromC(value.sep_m), MRSE: float64(value.mrse_m),
	}
}

func metricsErrorKind(value uint32) uint32 { return value }

func ErrorMetricsFromENU(covariance matrix3) (PositionErrorMetrics, uint32, error) {
	input := cMatrix3(covariance)
	var output C.SidereonPositionErrorMetrics
	var kind uint32
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_from_enu_covariance_m2(&input[0], &output, &kind)
	})
	return metricsFromC(output), metricsErrorKind(kind), err
}

func ErrorMetricsFromECEF(covariance matrix3, receiver Geodetic) (PositionErrorMetrics, uint32, error) {
	input := cMatrix3(covariance)
	var output C.SidereonPositionErrorMetrics
	var kind uint32
	receiverC := cGeodetic(receiver)
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_from_ecef_covariance_m2(&input[0], receiverC, &output, &kind)
	})
	return metricsFromC(output), metricsErrorKind(kind), err
}

func ErrorMetricsFromKinematic(position [3]float64, covariance matrix3) (PositionErrorMetrics, uint32, error) {
	input := C.SidereonKinematicSolutionMetricsInput{
		position_m: [3]C.double{C.double(position[0]), C.double(position[1]), C.double(position[2])},
	}
	flat := cMatrix3(covariance)
	input.position_covariance_m2 = flat
	var output C.SidereonPositionErrorMetrics
	var kind uint32
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_from_kinematic_solution(&input, &output, &kind)
	})
	return metricsFromC(output), metricsErrorKind(kind), err
}

func ErrorMetricsFromPositionCovariance(ecef, enu matrix3) (PositionErrorMetrics, uint32, error) {
	var input C.SidereonPositionCovariance
	ecefC := cMatrix3(ecef)
	enuC := cMatrix3(enu)
	input.ecef_m2 = ecefC
	input.enu_m2 = enuC
	var output C.SidereonPositionErrorMetrics
	var kind uint32
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_from_position_covariance(&input, &output, &kind)
	})
	return metricsFromC(output), metricsErrorKind(kind), err
}

func ErrorEllipseFromENU(covariance matrix3) (ErrorEllipse, uint32, error) {
	input := cMatrix3(covariance)
	var output C.SidereonErrorEllipse
	var kind uint32
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_error_ellipse_from_enu_m2(&input[0], &output, &kind)
	})
	return ErrorEllipse{SemiMajorM: float64(output.semi_major_m), SemiMinorM: float64(output.semi_minor_m), OrientationRad: float64(output.orientation_rad)}, metricsErrorKind(kind), err
}

func HorizontalRadius(covariance matrix3, probability float64) (PercentileRadius, uint32, error) {
	input := cMatrix3(covariance)
	var output C.SidereonPercentileRadius
	var kind uint32
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_horizontal_radius_at(&input[0], C.double(probability), &output, &kind)
	})
	return percentileFromC(output), metricsErrorKind(kind), err
}

func SphericalRadius(covariance matrix3, probability float64) (PercentileRadius, uint32, error) {
	input := cMatrix3(covariance)
	var output C.SidereonPercentileRadius
	var kind uint32
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_spherical_radius_at(&input[0], C.double(probability), &output, &kind)
	})
	return percentileFromC(output), metricsErrorKind(kind), err
}

func VerticalRadius(sigmaUM2, probability float64) (float64, uint32, error) {
	var output C.double
	var kind uint32
	err := callStatus(func() uint32 {
		return C.sidereon_error_metrics_vertical_radius_at(C.double(sigmaUM2), C.double(probability), &output, &kind)
	})
	return float64(output), metricsErrorKind(kind), err
}

func NIS(innovation, variance float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_nis(C.double(innovation), C.double(variance), &output)
	})
	return float64(output), err
}

func NISExpected(dof uint64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_nis_expected_value(C.size_t(dof), &output)
	})
	return float64(output), err
}

func NISThreshold(dof uint64, confidence float64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_nis_gate_threshold(C.size_t(dof), C.double(confidence), &output)
	})
	return float64(output), err
}

func ComputeNISGate(innovation, variance float64, dof uint64, confidence float64) (NISGate, error) {
	var output C.SidereonNisGate
	err := callStatus(func() uint32 {
		return C.sidereon_nis_gate_test(C.double(innovation), C.double(variance), C.size_t(dof), C.double(confidence), &output)
	})
	return NISGate{NIS: float64(output.nis), Threshold: float64(output.threshold), InGate: bool(output.in_gate), DOF: uint64(output.dof)}, err
}

func ChiSquareInverse(probability float64, dof uint64) (float64, error) {
	var output C.double
	err := callStatus(func() uint32 {
		return C.sidereon_chi2_inv(C.double(probability), C.size_t(dof), &output)
	})
	return float64(output), err
}

func PolarMotionMatrix(xpArcsec, ypArcsec float64) (matrix3, error) {
	var output [9]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frame_polar_motion_matrix(C.double(xpArcsec), C.double(ypArcsec), &output[0])
	})
	return flatToMatrix(&output), err
}

func GCRSToITRSMatrix(scales TimeScales) (matrix3, error) {
	input := cTimeScales(scales)
	var output [9]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frame_gcrs_to_itrs_matrix(&input, &output[0])
	})
	return flatToMatrix(&output), err
}

func ITRSToGCRSMatrix(scales TimeScales) (matrix3, error) {
	input := cTimeScales(scales)
	var output [9]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frame_itrs_to_gcrs_matrix(&input, &output[0])
	})
	return flatToMatrix(&output), err
}

func MatrixVectorMultiply(matrix [3][3]float64, vector [3]float64) ([3]float64, error) {
	inputMatrix := cMatrix3(matrix3(matrix))
	inputVector := [3]C.double{C.double(vector[0]), C.double(vector[1]), C.double(vector[2])}
	var output [3]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frame_mat3_vec3_mul(&inputMatrix[0], &inputVector[0], &output[0])
	})
	return [3]float64{float64(output[0]), float64(output[1]), float64(output[2])}, err
}

func GCRSToITRS(positionKm [3]float64, scales TimeScales, skyfieldCompatible bool) ([3]float64, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	input := cTimeScales(scales)
	var output [3]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frame_gcrs_to_itrs(&position[0], &input, C.bool(skyfieldCompatible), &output[0])
	})
	return [3]float64{float64(output[0]), float64(output[1]), float64(output[2])}, err
}

func ITRSToGCRS(positionKm [3]float64, scales TimeScales) ([3]float64, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	input := cTimeScales(scales)
	var output [3]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frame_itrs_to_gcrs(&position[0], &input, &output[0])
	})
	return [3]float64{float64(output[0]), float64(output[1]), float64(output[2])}, err
}

func GCRSToTopocentric(positionKm [3]float64, stationLatDeg, stationLonDeg, stationAltKm float64, scales TimeScales, skyfieldCompatible bool) ([3]float64, error) {
	position := [3]C.double{C.double(positionKm[0]), C.double(positionKm[1]), C.double(positionKm[2])}
	input := cTimeScales(scales)
	var output [3]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_frame_gcrs_to_topocentric(&position[0], C.double(stationLatDeg), C.double(stationLonDeg), C.double(stationAltKm), &input, C.bool(skyfieldCompatible), &output[0])
	})
	return [3]float64{float64(output[0]), float64(output[1]), float64(output[2])}, err
}
