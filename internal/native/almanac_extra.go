//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"time"
	"unsafe"
)

type AlmanacStation struct{ LatitudeDeg, LongitudeDeg, AltitudeKm float64 }

type BodyAzEl struct{ AzimuthDeg, ElevationDeg, RangeKm float64 }
type MoonIlluminationValue struct{ IlluminatedFraction, PhaseAngleDeg float64 }
type MoonElevationOptions struct{ ElevationThresholdDeg, StepSeconds, TimeToleranceSeconds float64 }
type MoonElevationCrossing struct {
	Time         time.Time
	Kind         uint32
	ElevationDeg float64
}
type MoonTransit struct {
	Time         time.Time
	Kind         uint32
	ElevationDeg float64
}

type EclipseEvent struct {
	Time                              time.Time
	Kind                              uint32
	Magnitude, MoonLatitudeDeg, Gamma float64
	Uncertain                         bool
}
type MeridianTransit struct {
	Time        time.Time
	Kind        uint32
	AltitudeDeg float64
}
type MoonPhaseEvent struct {
	Time time.Time
	Kind uint32
}
type PlanetaryEvent struct {
	Time          time.Time
	Planet, Kind  uint32
	ElongationDeg float64
}
type SeasonEvent struct {
	Time time.Time
	Kind uint32
}

type Refraction struct{ PressureMbar, TemperatureC float64 }
type ObserveOptions struct {
	HasPolarMotion         bool
	XPRad, YPRad           float64
	HasRefraction          bool
	Refraction             Refraction
	Deflection, Aberration bool
}
type Equatorial struct{ RightAscensionDeg, RightAscensionHours, DeclinationDeg, DistanceKm float64 }
type Horizontal struct{ AzimuthDeg, ElevationDeg, RangeKm float64 }
type Ecliptic struct{ LongitudeDeg, LatitudeDeg, DistanceKm float64 }
type BodyObservation struct {
	Astrometric, ApparentICRS, Apparent Equatorial
	Horizontal                          Horizontal
	HourAngleDeg, HourAngleHours        float64
	Ecliptic                            Ecliptic
	Reduced                             bool
}
type SurfacePoint struct{ LatitudeDeg, LongitudeDeg float64 }

const (
	AlmanacEclipseLunarPenumbral uint32 = C.SIDEREON_ALMANAC_ECLIPSE_KIND_LUNAR_PENUMBRAL
	AlmanacEclipseLunarPartial   uint32 = C.SIDEREON_ALMANAC_ECLIPSE_KIND_LUNAR_PARTIAL
	AlmanacEclipseLunarTotal     uint32 = C.SIDEREON_ALMANAC_ECLIPSE_KIND_LUNAR_TOTAL
	AlmanacEclipseSolarPartial   uint32 = C.SIDEREON_ALMANAC_ECLIPSE_KIND_SOLAR_PARTIAL
	AlmanacEclipseSolarAnnular   uint32 = C.SIDEREON_ALMANAC_ECLIPSE_KIND_SOLAR_ANNULAR
	AlmanacEclipseSolarTotal     uint32 = C.SIDEREON_ALMANAC_ECLIPSE_KIND_SOLAR_TOTAL
	AlmanacEclipseSolarHybrid    uint32 = C.SIDEREON_ALMANAC_ECLIPSE_KIND_SOLAR_HYBRID
	TransitBodySun               uint32 = C.SIDEREON_TRANSIT_BODY_KIND_SUN
	TransitBodyMoon              uint32 = C.SIDEREON_TRANSIT_BODY_KIND_MOON
	TransitBodyPlanet            uint32 = C.SIDEREON_TRANSIT_BODY_KIND_PLANET
)

func cAlmanacStation(value AlmanacStation) C.SidereonGeodeticStation {
	return C.SidereonGeodeticStation{latitude_deg: C.double(value.LatitudeDeg), longitude_deg: C.double(value.LongitudeDeg), altitude_km: C.double(value.AltitudeKm)}
}
func unixTime(value C.int64_t) time.Time { return time.UnixMicro(int64(value)).UTC() }

func cMoonOptions(value MoonElevationOptions) C.SidereonMoonElevationOptions {
	return C.SidereonMoonElevationOptions{elevation_threshold_deg: C.double(value.ElevationThresholdDeg), step_seconds: C.double(value.StepSeconds), time_tolerance_seconds: C.double(value.TimeToleranceSeconds)}
}
func moonOptionsFromC(value C.SidereonMoonElevationOptions) MoonElevationOptions {
	return MoonElevationOptions{ElevationThresholdDeg: float64(value.elevation_threshold_deg), StepSeconds: float64(value.step_seconds), TimeToleranceSeconds: float64(value.time_tolerance_seconds)}
}

func cObserveOptions(value ObserveOptions) C.SidereonObserveOptions {
	return C.SidereonObserveOptions{has_polar_motion: C.bool(value.HasPolarMotion), xp_rad: C.double(value.XPRad), yp_rad: C.double(value.YPRad), has_refraction: C.bool(value.HasRefraction), refraction: C.SidereonRefraction{pressure_mbar: C.double(value.Refraction.PressureMbar), temperature_c: C.double(value.Refraction.TemperatureC)}, deflection: C.bool(value.Deflection), aberration: C.bool(value.Aberration)}
}
func observeOptionsFromC(value C.SidereonObserveOptions) ObserveOptions {
	return ObserveOptions{HasPolarMotion: bool(value.has_polar_motion), XPRad: float64(value.xp_rad), YPRad: float64(value.yp_rad), HasRefraction: bool(value.has_refraction), Refraction: Refraction{PressureMbar: float64(value.refraction.pressure_mbar), TemperatureC: float64(value.refraction.temperature_c)}, Deflection: bool(value.deflection), Aberration: bool(value.aberration)}
}

func bodyObservationFromC(value C.SidereonBodyObservation) BodyObservation {
	return BodyObservation{
		Astrometric:  Equatorial{RightAscensionDeg: float64(value.astrometric.right_ascension_deg), RightAscensionHours: float64(value.astrometric.right_ascension_hours), DeclinationDeg: float64(value.astrometric.declination_deg), DistanceKm: float64(value.astrometric.distance_km)},
		ApparentICRS: Equatorial{RightAscensionDeg: float64(value.apparent_icrs.right_ascension_deg), RightAscensionHours: float64(value.apparent_icrs.right_ascension_hours), DeclinationDeg: float64(value.apparent_icrs.declination_deg), DistanceKm: float64(value.apparent_icrs.distance_km)},
		Apparent:     Equatorial{RightAscensionDeg: float64(value.apparent.right_ascension_deg), RightAscensionHours: float64(value.apparent.right_ascension_hours), DeclinationDeg: float64(value.apparent.declination_deg), DistanceKm: float64(value.apparent.distance_km)},
		Horizontal:   Horizontal{AzimuthDeg: float64(value.horizontal.azimuth_deg), ElevationDeg: float64(value.horizontal.elevation_deg), RangeKm: float64(value.horizontal.range_km)},
		HourAngleDeg: float64(value.hour_angle_deg), HourAngleHours: float64(value.hour_angle_hours),
		Ecliptic: Ecliptic{LongitudeDeg: float64(value.ecliptic.longitude_deg), LatitudeDeg: float64(value.ecliptic.latitude_deg), DistanceKm: float64(value.ecliptic.distance_km)}, Reduced: bool(value.reduced),
	}
}

func queryAlmanacOutput(label string, elementSize uintptr, invoke func(unsafe.Pointer, C.size_t, *C.size_t, *C.size_t) uint32, copyValues func(unsafe.Pointer, int) interface{}) (interface{}, error) {
	var written, required C.size_t
	if err := callStatus(func() uint32 { return invoke(nil, 0, &written, &required) }); err != nil {
		return nil, err
	}
	count, err := validateNativeQuery(label, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, elementSize); err != nil {
		return nil, err
	}
	size, err := checkedNativeAllocationSize(count, elementSize)
	if err != nil {
		return nil, err
	}
	var ptr unsafe.Pointer
	if size != 0 {
		ptr = C.malloc(C.size_t(size))
		if ptr == nil {
			return nil, errors.New("sidereon: unable to allocate native almanac output")
		}
		defer C.free(ptr)
	}
	written, required = 0, 0
	if err := callStatus(func() uint32 { return invoke(ptr, C.size_t(count), &written, &required) }); err != nil {
		return nil, err
	}
	n, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	return copyValues(ptr, n), nil
}

func SunMoonECI(ttJulianCenturies float64) ([3]float64, [3]float64, error) {
	var sun, moon [3]C.double
	err := callStatus(func() uint32 { return C.sidereon_sun_moon_eci(C.double(ttJulianCenturies), &sun[0], &moon[0]) })
	return vectorsFromC(sun), vectorsFromC(moon), err
}

func vectorsFromC(value [3]C.double) (out [3]float64) {
	for i := range out {
		out[i] = float64(value[i])
	}
	return
}

func SunMoonECEF(epoch time.Time) ([3]float64, [3]float64, error) {
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return [3]float64{}, [3]float64{}, err
	}
	var sun, moon [3]C.double
	err = callStatus(func() uint32 { return C.sidereon_sun_moon_ecef(C.int64_t(unix), &sun[0], &moon[0]) })
	return vectorsFromC(sun), vectorsFromC(moon), err
}

func sunMoonBatch(times []time.Time, ecef bool) ([][3]float64, [][3]float64, error) {
	if len(times) > int(^uint(0)>>1)/3 {
		return nil, nil, errors.New("sidereon: Sun/Moon batch shape overflows")
	}
	epochs, err := unixMicrosecondsSlice(times)
	if err != nil {
		return nil, nil, err
	}
	n := len(times) * 3
	if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.double(0))); err != nil {
		return nil, nil, err
	}
	sun, moon := make([]C.double, n), make([]C.double, n)
	var pEpoch *C.int64_t
	if len(epochs) > 0 {
		pEpoch = (*C.int64_t)(unsafe.Pointer(&epochs[0]))
	}
	call := func() uint32 {
		if ecef {
			return C.sidereon_sun_moon_ecef_batch(pEpoch, C.size_t(len(times)), &sun[0], C.size_t(n), &moon[0], C.size_t(n))
		}
		return C.sidereon_sun_moon_eci_batch(pEpoch, C.size_t(len(times)), &sun[0], C.size_t(n), &moon[0], C.size_t(n))
	}
	if n == 0 {
		return [][3]float64{}, [][3]float64{}, nil
	}
	if err := callStatus(call); err != nil {
		runtime.KeepAlive(epochs)
		return nil, nil, err
	}
	runtime.KeepAlive(epochs)
	outSun, outMoon := make([][3]float64, len(times)), make([][3]float64, len(times))
	for i := range times {
		for k := 0; k < 3; k++ {
			outSun[i][k], outMoon[i][k] = float64(sun[3*i+k]), float64(moon[3*i+k])
		}
	}
	return outSun, outMoon, nil
}

func SunMoonECIBatch(times []time.Time) ([][3]float64, [][3]float64, error) {
	return sunMoonBatch(times, false)
}
func SunMoonECEFBatch(times []time.Time) ([][3]float64, [][3]float64, error) {
	return sunMoonBatch(times, true)
}

func SunAngle(satelliteKm, sunKm [3]float64) (float64, error) {
	satellite, sun := [3]C.double{}, [3]C.double{}
	for i := range satellite {
		satellite[i], sun[i] = C.double(satelliteKm[i]), C.double(sunKm[i])
	}
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_sun_angle_deg(&satellite[0], &sun[0], &out) })
	return float64(out), err
}

func MoonAngle(satelliteKm, moonKm [3]float64) (float64, error) {
	satellite, moon := [3]C.double{}, [3]C.double{}
	for i := range satellite {
		satellite[i], moon[i] = C.double(satelliteKm[i]), C.double(moonKm[i])
	}
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_moon_angle_deg(&satellite[0], &moon[0], &out) })
	return float64(out), err
}

func SunElevation(satelliteKm, sunKm [3]float64) (float64, error) {
	satellite, sun := [3]C.double{}, [3]C.double{}
	for i := range satellite {
		satellite[i], sun[i] = C.double(satelliteKm[i]), C.double(sunKm[i])
	}
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_sun_elevation_deg(&satellite[0], &sun[0], &out) })
	return float64(out), err
}

func NutationIAU2000A(jdTT float64) (float64, float64, error) {
	var dpsi, deps C.double
	err := callStatus(func() uint32 { return C.sidereon_nutation_iau2000a_radians(C.double(jdTT), &dpsi, &deps) })
	return float64(dpsi), float64(deps), err
}
func NutationMeanObliquity(jdTDB float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_nutation_mean_obliquity_radians(C.double(jdTDB), &out) })
	return float64(out), err
}
func NutationFundamentalArguments(tCenturies float64) ([5]float64, error) {
	var values [5]C.double
	err := callStatus(func() uint32 { return C.sidereon_nutation_fundamental_arguments(C.double(tCenturies), &values[0]) })
	var out [5]float64
	for i := range out {
		out[i] = float64(values[i])
	}
	return out, err
}
func NutationEquationOfEquinoxesTerms(jdTT float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_nutation_equation_of_equinoxes_terms(C.double(jdTT), &out) })
	return float64(out), err
}
func NutationMatrix(mean, trueObliquity, psi float64) ([3][3]float64, error) {
	var values [9]C.double
	err := callStatus(func() uint32 {
		return C.sidereon_nutation_matrix(C.double(mean), C.double(trueObliquity), C.double(psi), &values[0])
	})
	return matrix3FromC(values), err
}
func PrecessionMatrix(jdTDB float64) ([3][3]float64, error) {
	var values [9]C.double
	err := callStatus(func() uint32 { return C.sidereon_precession_matrix(C.double(jdTDB), &values[0]) })
	return matrix3FromC(values), err
}
func PrecessionICRSToJ2000Matrix() ([3][3]float64, error) {
	var values [9]C.double
	err := callStatus(func() uint32 { return C.sidereon_precession_icrs_to_j2000_matrix(&values[0]) })
	return matrix3FromC(values), err
}
func matrix3FromC(value [9]C.double) (out [3][3]float64) {
	for i := range out {
		for j := range out[i] {
			out[i][j] = float64(value[3*i+j])
		}
	}
	return
}

func SunAzEl(station AlmanacStation, epoch time.Time) (BodyAzEl, error) {
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return BodyAzEl{}, err
	}
	var out C.SidereonBodyAzEl
	c := cAlmanacStation(station)
	err = callStatus(func() uint32 { return C.sidereon_sun_az_el(&c, C.int64_t(unix), &out) })
	return BodyAzEl{AzimuthDeg: float64(out.azimuth_deg), ElevationDeg: float64(out.elevation_deg), RangeKm: float64(out.range_km)}, err
}
func MoonAzEl(station AlmanacStation, epoch time.Time) (BodyAzEl, error) {
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return BodyAzEl{}, err
	}
	var out C.SidereonBodyAzEl
	c := cAlmanacStation(station)
	err = callStatus(func() uint32 { return C.sidereon_moon_az_el(&c, C.int64_t(unix), &out) })
	return BodyAzEl{AzimuthDeg: float64(out.azimuth_deg), ElevationDeg: float64(out.elevation_deg), RangeKm: float64(out.range_km)}, err
}
func MoonElevation(station AlmanacStation, epoch time.Time) (float64, error) {
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return 0, err
	}
	var out C.double
	cStation := cAlmanacStation(station)
	err = callStatus(func() uint32 { return C.sidereon_moon_elevation_deg(&cStation, C.int64_t(unix), &out) })
	return float64(out), err
}
func MoonIllumination(station AlmanacStation, epoch time.Time) (MoonIlluminationValue, error) {
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return MoonIlluminationValue{}, err
	}
	var out C.SidereonMoonIllumination
	c := cAlmanacStation(station)
	err = callStatus(func() uint32 { return C.sidereon_moon_illumination(&c, C.int64_t(unix), &out) })
	return MoonIlluminationValue{IlluminatedFraction: float64(out.illuminated_fraction), PhaseAngleDeg: float64(out.phase_angle_deg)}, err
}
func MoonElevationOptionsDefaults() (MoonElevationOptions, error) {
	var out C.SidereonMoonElevationOptions
	err := callStatus(func() uint32 { return C.sidereon_moon_elevation_options_init(&out) })
	return moonOptionsFromC(out), err
}

func findMoonCrossings(station AlmanacStation, start, end time.Time, options *MoonElevationOptions, transit bool) (interface{}, error) {
	startUS, err := unixMicroseconds(start)
	if err != nil {
		return nil, err
	}
	endUS, err := unixMicroseconds(end)
	if err != nil {
		return nil, err
	}
	cStation := cAlmanacStation(station)
	var cOptions C.SidereonMoonElevationOptions
	var optionsPtr *C.SidereonMoonElevationOptions
	if options != nil {
		cOptions, optionsPtr = cMoonOptions(*options), &cOptions
	}
	if transit {
		return queryAlmanacOutput("Moon transits", unsafe.Sizeof(C.SidereonMoonTransit{}), func(ptr unsafe.Pointer, length C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_find_moon_transits(&cStation, C.int64_t(startUS), C.int64_t(endUS), func() C.double {
				if options == nil {
					return 0
				}
				return C.double(options.StepSeconds)
			}(), func() C.double {
				if options == nil {
					return 0
				}
				return C.double(options.TimeToleranceSeconds)
			}(), (*C.SidereonMoonTransit)(ptr), length, w, r)
		}, func(ptr unsafe.Pointer, n int) interface{} {
			values := unsafe.Slice((*C.SidereonMoonTransit)(ptr), n)
			out := make([]MoonTransit, n)
			for i := range out {
				out[i] = MoonTransit{Time: unixTime(values[i].time_unix_us), Kind: uint32(values[i].kind), ElevationDeg: float64(values[i].elevation_deg)}
			}
			return out
		})
	}
	return queryAlmanacOutput("Moon elevation crossings", unsafe.Sizeof(C.SidereonMoonElevationCrossing{}), func(ptr unsafe.Pointer, length C.size_t, w, r *C.size_t) uint32 {
		return C.sidereon_find_moon_elevation_crossings(&cStation, C.int64_t(startUS), C.int64_t(endUS), optionsPtr, (*C.SidereonMoonElevationCrossing)(ptr), length, w, r)
	}, func(ptr unsafe.Pointer, n int) interface{} {
		values := unsafe.Slice((*C.SidereonMoonElevationCrossing)(ptr), n)
		out := make([]MoonElevationCrossing, n)
		for i := range out {
			out[i] = MoonElevationCrossing{Time: unixTime(values[i].time_unix_us), Kind: uint32(values[i].kind), ElevationDeg: float64(values[i].elevation_deg)}
		}
		return out
	})
}

func FindMoonCrossings(station AlmanacStation, start, end time.Time, options *MoonElevationOptions) ([]MoonElevationCrossing, error) {
	value, err := findMoonCrossings(station, start, end, options, false)
	if err != nil {
		return nil, err
	}
	return value.([]MoonElevationCrossing), nil
}
func FindMoonTransits(station AlmanacStation, start, end time.Time, stepSeconds, toleranceSeconds float64) ([]MoonTransit, error) {
	value, err := findMoonCrossings(station, start, end, &MoonElevationOptions{StepSeconds: stepSeconds, TimeToleranceSeconds: toleranceSeconds}, true)
	if err != nil {
		return nil, err
	}
	return value.([]MoonTransit), nil
}

func ObserveOptionsDefaults() (ObserveOptions, error) {
	var out C.SidereonObserveOptions
	err := callStatus(func() uint32 { return C.sidereon_observe_options_init(&out) })
	return observeOptionsFromC(out), err
}

func Observe(station AlmanacStation, epoch time.Time, targetKind uint32, spk *SPK, naifID int32, positionKm, velocityKmPerS *[3]float64, options *ObserveOptions) (BodyObservation, error) {
	if targetKind > 3 {
		return BodyObservation{}, invalidArgument("observe target kind is not defined by the C ABI")
	}
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return BodyObservation{}, err
	}
	cStation := cAlmanacStation(station)
	var cOptions C.SidereonObserveOptions
	var optionsPtr *C.SidereonObserveOptions
	if options != nil {
		cOptions, optionsPtr = cObserveOptions(*options), &cOptions
	}
	var position, velocity [3]C.double
	var positionPtr, velocityPtr *C.double
	if positionKm != nil {
		for i := range position {
			position[i] = C.double(positionKm[i])
		}
		positionPtr = &position[0]
	}
	if velocityKmPerS != nil {
		for i := range velocity {
			velocity[i] = C.double(velocityKmPerS[i])
		}
		velocityPtr = &velocity[0]
	}
	var out C.SidereonBodyObservation
	var errCall error
	invoke := func(pointer *C.SidereonSpk) error {
		return callStatus(func() uint32 {
			return C.sidereon_observe(&cStation, C.int64_t(unix), C.uint32_t(targetKind), pointer, C.int32_t(naifID), positionPtr, velocityPtr, optionsPtr, &out)
		})
	}
	if targetKind <= 1 {
		errCall = invoke(nil)
	} else {
		if spk == nil || spk.handle == nil {
			return BodyObservation{}, invalidArgument("selected observation target requires an SPK handle")
		}
		errCall = spk.handle.with(func(pointer unsafe.Pointer) error { return invoke((*C.SidereonSpk)(pointer)) })
	}
	return bodyObservationFromC(out), errCall
}

func ObserveSPKBody(station AlmanacStation, epoch time.Time, spk *SPK, naifID int32) (BodyObservation, error) {
	if spk == nil {
		return BodyObservation{}, invalidArgument("observe requires an SPK handle")
	}
	unix, err := unixMicroseconds(epoch)
	if err != nil {
		return BodyObservation{}, err
	}
	cStation := cAlmanacStation(station)
	var out C.SidereonBodyObservation
	err = spk.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_observe_spk_body(&cStation, C.int64_t(unix), (*C.SidereonSpk)(pointer), C.int32_t(naifID), &out)
		})
	})
	return bodyObservationFromC(out), err
}

func SolidEarthTide(station, sun, moon [3]float64, year, month, day int, fractionalHour float64) ([3]float64, error) {
	y, err := checkedInt32(year, "year")
	if err != nil {
		return [3]float64{}, err
	}
	mo, err := checkedInt32(month, "month")
	if err != nil {
		return [3]float64{}, err
	}
	d, err := checkedInt32(day, "day")
	if err != nil {
		return [3]float64{}, err
	}
	var s, su, m, out [3]C.double
	for i := 0; i < 3; i++ {
		s[i], su[i], m[i] = C.double(station[i]), C.double(sun[i]), C.double(moon[i])
	}
	err = callStatus(func() uint32 {
		return C.sidereon_solid_earth_tide(&s[0], C.int32_t(y), C.int32_t(mo), C.int32_t(d), C.double(fractionalHour), &su[0], &m[0], &out[0])
	})
	return vectorsFromC(out), err
}
func SolidEarthPoleTide(station [3]float64, year, month, day int, fractionalHour, xpArcsec, ypArcsec float64) ([3]float64, error) {
	y, err := checkedInt32(year, "year")
	if err != nil {
		return [3]float64{}, err
	}
	mo, err := checkedInt32(month, "month")
	if err != nil {
		return [3]float64{}, err
	}
	d, err := checkedInt32(day, "day")
	if err != nil {
		return [3]float64{}, err
	}
	var s, out [3]C.double
	for i := range s {
		s[i] = C.double(station[i])
	}
	err = callStatus(func() uint32 {
		return C.sidereon_solid_earth_pole_tide(&s[0], C.int32_t(y), C.int32_t(mo), C.int32_t(d), C.double(fractionalHour), C.double(xpArcsec), C.double(ypArcsec), &out[0])
	})
	return vectorsFromC(out), err
}

type OceanLoadingBLQ struct {
	AmplitudeM [3][11]float64
	PhaseDeg   [3][11]float64
}

func cOcean(value OceanLoadingBLQ) C.SidereonOceanLoadingBlq {
	var out C.SidereonOceanLoadingBlq
	for i := 0; i < 3; i++ {
		for j := 0; j < 11; j++ {
			out.amplitude_m[i][j], out.phase_deg[i][j] = C.double(value.AmplitudeM[i][j]), C.double(value.PhaseDeg[i][j])
		}
	}
	return out
}
func OceanTideLoading(station [3]float64, year, month, day int, fractionalHour float64, loading OceanLoadingBLQ) ([3]float64, error) {
	y, err := checkedInt32(year, "year")
	if err != nil {
		return [3]float64{}, err
	}
	mo, err := checkedInt32(month, "month")
	if err != nil {
		return [3]float64{}, err
	}
	d, err := checkedInt32(day, "day")
	if err != nil {
		return [3]float64{}, err
	}
	var s, out [3]C.double
	for i := range s {
		s[i] = C.double(station[i])
	}
	blq := cOcean(loading)
	err = callStatus(func() uint32 {
		return C.sidereon_ocean_tide_loading(&s[0], C.int32_t(y), C.int32_t(mo), C.int32_t(d), C.double(fractionalHour), &blq, &out[0])
	})
	return vectorsFromC(out), err
}

func SubSolarPoint(sunECEF [3]float64) (SurfacePoint, error) {
	var s [3]C.double
	for i := range s {
		s[i] = C.double(sunECEF[i])
	}
	var out C.SidereonSurfacePoint
	err := callStatus(func() uint32 { return C.sidereon_sub_solar_point(&s[0], &out) })
	return SurfacePoint{LatitudeDeg: float64(out.latitude_deg), LongitudeDeg: float64(out.longitude_deg)}, err
}
func SubObserverPoint(observer [3]float64, poleRA, poleDec, primeMeridian float64) (SurfacePoint, error) {
	var o [3]C.double
	for i := range o {
		o[i] = C.double(observer[i])
	}
	var out C.SidereonSurfacePoint
	err := callStatus(func() uint32 {
		return C.sidereon_sub_observer_point(&o[0], C.double(poleRA), C.double(poleDec), C.double(primeMeridian), &out)
	})
	return SurfacePoint{LatitudeDeg: float64(out.latitude_deg), LongitudeDeg: float64(out.longitude_deg)}, err
}
func TerminatorLatitude(subSolarLat, subSolarLon, longitude float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 {
		return C.sidereon_terminator_latitude_deg(C.double(subSolarLat), C.double(subSolarLon), C.double(longitude), &out)
	})
	return float64(out), err
}
func ParallacticAngle(observerLat, hourAngle, declination float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 {
		return C.sidereon_parallactic_angle_deg(C.double(observerLat), C.double(hourAngle), C.double(declination), &out)
	})
	return float64(out), err
}
func PhaseAngle(sat, sun, observer [3]float64) (float64, error) {
	var a, b, c [3]C.double
	for i := range a {
		a[i], b[i], c[i] = C.double(sat[i]), C.double(sun[i]), C.double(observer[i])
	}
	var out C.double
	err := callStatus(func() uint32 { return C.sidereon_phase_angle_deg(&a[0], &b[0], &c[0], &out) })
	return float64(out), err
}
func PositionAngle(fromLon, fromLat, toLon, toLat float64) (float64, error) {
	var out C.double
	err := callStatus(func() uint32 {
		return C.sidereon_position_angle_deg(C.double(fromLon), C.double(fromLat), C.double(toLon), C.double(toLat), &out)
	})
	return float64(out), err
}

func almanacEvents(spk *SPK, start, end time.Time, step, tolerance float64, kind string, body, planet, eventKind uint32) (interface{}, error) {
	if spk == nil {
		return nil, invalidArgument("almanac events require an SPK handle")
	}
	startUS, err := unixMicroseconds(start)
	if err != nil {
		return nil, err
	}
	endUS, err := unixMicroseconds(end)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = spk.handle.with(func(pointer unsafe.Pointer) error {
		invoke := func(ptr unsafe.Pointer, n C.size_t, w, r *C.size_t) uint32 {
			source := (*C.SidereonSpk)(pointer)
			switch kind {
			case "seasons":
				return C.sidereon_almanac_seasons(source, C.int64_t(startUS), C.int64_t(endUS), C.double(step), C.double(tolerance), (*C.SidereonSeasonEvent)(ptr), n, w, r)
			case "moon phases":
				return C.sidereon_almanac_moon_phases(source, C.int64_t(startUS), C.int64_t(endUS), C.double(step), C.double(tolerance), (*C.SidereonMoonPhaseEvent)(ptr), n, w, r)
			case "planetary events":
				return C.sidereon_almanac_planetary_events(source, C.uint32_t(planet), C.uint32_t(eventKind), C.int64_t(startUS), C.int64_t(endUS), C.double(step), C.double(tolerance), (*C.SidereonPlanetaryEvent)(ptr), n, w, r)
			case "eclipses":
				return C.sidereon_almanac_lunar_solar_eclipses(source, C.int64_t(startUS), C.int64_t(endUS), C.double(step), C.double(tolerance), (*C.SidereonAlmanacEclipseEvent)(ptr), n, w, r)
			default:
				return uint32(C.SIDEREON_STATUS_INVALID_ARGUMENT)
			}
		}
		var size uintptr
		var converter func(unsafe.Pointer, int) interface{}
		switch kind {
		case "seasons":
			size = unsafe.Sizeof(C.SidereonSeasonEvent{})
			converter = func(ptr unsafe.Pointer, n int) interface{} {
				values := unsafe.Slice((*C.SidereonSeasonEvent)(ptr), n)
				out := make([]SeasonEvent, n)
				for i := range out {
					out[i] = SeasonEvent{Time: unixTime(values[i].time_unix_us), Kind: uint32(values[i].kind)}
				}
				return out
			}
		case "moon phases":
			size = unsafe.Sizeof(C.SidereonMoonPhaseEvent{})
			converter = func(ptr unsafe.Pointer, n int) interface{} {
				values := unsafe.Slice((*C.SidereonMoonPhaseEvent)(ptr), n)
				out := make([]MoonPhaseEvent, n)
				for i := range out {
					out[i] = MoonPhaseEvent{Time: unixTime(values[i].time_unix_us), Kind: uint32(values[i].kind)}
				}
				return out
			}
		case "planetary events":
			size = unsafe.Sizeof(C.SidereonPlanetaryEvent{})
			converter = func(ptr unsafe.Pointer, n int) interface{} {
				values := unsafe.Slice((*C.SidereonPlanetaryEvent)(ptr), n)
				out := make([]PlanetaryEvent, n)
				for i := range out {
					out[i] = PlanetaryEvent{Time: unixTime(values[i].time_unix_us), Planet: uint32(values[i].planet), Kind: uint32(values[i].kind), ElongationDeg: float64(values[i].elongation_deg)}
				}
				return out
			}
		case "eclipses":
			size = unsafe.Sizeof(C.SidereonAlmanacEclipseEvent{})
			converter = func(ptr unsafe.Pointer, n int) interface{} {
				values := unsafe.Slice((*C.SidereonAlmanacEclipseEvent)(ptr), n)
				out := make([]EclipseEvent, n)
				for i := range out {
					out[i] = EclipseEvent{Time: unixTime(values[i].time_maximum_unix_us), Kind: uint32(values[i].kind), Magnitude: float64(values[i].magnitude), MoonLatitudeDeg: float64(values[i].moon_latitude_deg), Gamma: float64(values[i].gamma), Uncertain: bool(values[i].uncertain)}
				}
				return out
			}
		default:
			return invalidArgument("unsupported almanac event")
		}
		value, queryErr := queryAlmanacOutput(kind, size, invoke, converter)
		if queryErr == nil {
			result = value
		}
		return queryErr
	})
	return result, err
}

func AlmanacSeasons(spk *SPK, start, end time.Time, step, tolerance float64) ([]SeasonEvent, error) {
	value, err := almanacEvents(spk, start, end, step, tolerance, "seasons", 0, 0, 0)
	if err != nil {
		return nil, err
	}
	return value.([]SeasonEvent), nil
}
func AlmanacMoonPhases(spk *SPK, start, end time.Time, step, tolerance float64) ([]MoonPhaseEvent, error) {
	value, err := almanacEvents(spk, start, end, step, tolerance, "moon phases", 0, 0, 0)
	if err != nil {
		return nil, err
	}
	return value.([]MoonPhaseEvent), nil
}
func AlmanacPlanetaryEvents(spk *SPK, planet, eventKind uint32, start, end time.Time, step, tolerance float64) ([]PlanetaryEvent, error) {
	if planet > 6 || eventKind > 1 {
		return nil, invalidArgument("planetary event enum is not defined by the C ABI")
	}
	value, err := almanacEvents(spk, start, end, step, tolerance, "planetary events", 0, planet, eventKind)
	if err != nil {
		return nil, err
	}
	return value.([]PlanetaryEvent), nil
}
func AlmanacEclipses(spk *SPK, start, end time.Time, step, tolerance float64) ([]EclipseEvent, error) {
	value, err := almanacEvents(spk, start, end, step, tolerance, "eclipses", 0, 0, 0)
	if err != nil {
		return nil, err
	}
	return value.([]EclipseEvent), nil
}

func AlmanacMeridianTransits(spk *SPK, station AlmanacStation, body, planet uint32, start, end time.Time, step, tolerance float64) ([]MeridianTransit, error) {
	if body > TransitBodyPlanet || (body == TransitBodyPlanet && planet > 6) {
		return nil, invalidArgument("transit enum is not defined by the C ABI")
	}
	if spk == nil {
		return nil, invalidArgument("almanac events require an SPK handle")
	}
	startUS, err := unixMicroseconds(start)
	if err != nil {
		return nil, err
	}
	endUS, err := unixMicroseconds(end)
	if err != nil {
		return nil, err
	}
	cStation := cAlmanacStation(station)
	var result interface{}
	err = spk.handle.with(func(pointer unsafe.Pointer) error {
		invoke := func(ptr unsafe.Pointer, n C.size_t, w, r *C.size_t) uint32 {
			return C.sidereon_almanac_meridian_transits((*C.SidereonSpk)(pointer), C.uint32_t(body), C.uint32_t(planet), &cStation, C.int64_t(startUS), C.int64_t(endUS), C.double(step), C.double(tolerance), (*C.SidereonMeridianTransit)(ptr), n, w, r)
		}
		value, qErr := queryAlmanacOutput("meridian transits", unsafe.Sizeof(C.SidereonMeridianTransit{}), invoke, func(ptr unsafe.Pointer, n int) interface{} {
			values := unsafe.Slice((*C.SidereonMeridianTransit)(ptr), n)
			out := make([]MeridianTransit, n)
			for i := range out {
				out[i] = MeridianTransit{Time: unixTime(values[i].time_unix_us), Kind: uint32(values[i].kind), AltitudeDeg: float64(values[i].altitude_deg)}
			}
			return out
		})
		if qErr == nil {
			result = value
		}
		return qErr
	})
	if err != nil {
		return nil, err
	}
	return result.([]MeridianTransit), nil
}
