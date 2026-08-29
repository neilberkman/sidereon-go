package sidereon

import (
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// AlmanacStation uses degrees and kilometres, matching the astronomy C
// station type. It is distinct from GroundStation, whose height is metres.
type AlmanacStation struct {
	// LatitudeDeg is the latitude deg in degrees.
	LatitudeDeg float64
	// LongitudeDeg is the longitude deg in degrees.
	LongitudeDeg float64
	// AltitudeKm is the altitude km in kilometres.
	AltitudeKm float64
}

// BodyAzEl contains azimuth and elevation angles in degrees and range in kilometres.
type BodyAzEl struct {
	// AzimuthDeg is the azimuth deg in degrees.
	AzimuthDeg float64
	// ElevationDeg is the elevation deg in degrees.
	ElevationDeg float64
	// RangeKm is the range km in kilometres.
	RangeKm float64
}

// MoonIllumination contains illuminated fraction and phase angle in degrees.
type MoonIllumination struct {
	// IlluminatedFraction is the illuminated portion of the lunar disk, from 0 to 1.
	IlluminatedFraction float64
	// PhaseAngleDeg is the phase angle deg in degrees.
	PhaseAngleDeg float64
}

// MoonElevationOptions configures a Moon elevation search with degree and second tolerances.
type MoonElevationOptions struct {
	// ElevationThresholdDeg is the elevation threshold deg in degrees.
	ElevationThresholdDeg float64
	// StepSeconds is the step seconds in seconds.
	StepSeconds float64
	// TimeToleranceSeconds is the time tolerance seconds in seconds.
	TimeToleranceSeconds float64
}

// MoonElevationCrossing records a Moon elevation threshold crossing and its timestamp.
type MoonElevationCrossing struct {
	// Time is the timestamp for this record.
	Time time.Time
	// Kind is the event or record kind.
	Kind MoonRiseSetKind
	// ElevationDeg is the elevation deg in degrees.
	ElevationDeg float64
}

// MoonRiseSetKind classifies a Moon rising or setting crossing.
type MoonRiseSetKind uint32

const (
	// MoonRising identifies the moon rising case.
	MoonRising MoonRiseSetKind = iota
	// MoonSetting identifies the moon setting case.
	MoonSetting
)

// MoonTransit records a Moon meridian transit, timestamp, and elevation.
type MoonTransit struct {
	// Time is the timestamp for this record.
	Time time.Time
	// Kind is the event or record kind.
	Kind MoonTransitKind
	// ElevationDeg is the elevation deg in degrees.
	ElevationDeg float64
}

// MoonTransitKind classifies upper and lower Moon culminations.
type MoonTransitKind uint32

const (
	// MoonUpperCulmination identifies the moon upper culmination case.
	MoonUpperCulmination MoonTransitKind = iota
	// MoonLowerCulmination identifies the moon lower culmination case.
	MoonLowerCulmination
)

// AlmanacEclipseKind classifies lunar and solar eclipse events.
type AlmanacEclipseKind uint32

const (
	// LunarPenumbralEclipse identifies the lunar penumbral eclipse case.
	LunarPenumbralEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseLunarPenumbral)
	// LunarPartialEclipse identifies the lunar partial eclipse case.
	LunarPartialEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseLunarPartial)
	// LunarTotalEclipse identifies the lunar total eclipse case.
	LunarTotalEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseLunarTotal)
	// SolarPartialEclipse identifies the solar partial eclipse case.
	SolarPartialEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarPartial)
	// SolarAnnularEclipse identifies the solar annular eclipse case.
	SolarAnnularEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarAnnular)
	// SolarTotalEclipse identifies the solar total eclipse case.
	SolarTotalEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarTotal)
	// SolarHybridEclipse identifies the solar hybrid eclipse case.
	SolarHybridEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarHybrid)
)

// EclipseEvent records an eclipse kind, magnitude, geometry, and timestamp.
type EclipseEvent struct {
	// Time is the timestamp for this record.
	Time time.Time
	// Kind is the event or record kind.
	Kind AlmanacEclipseKind
	// Magnitude is the eclipse magnitude reported by the native eclipse model.
	Magnitude float64
	// MoonLatitudeDeg is the moon latitude deg in degrees.
	MoonLatitudeDeg float64
	// Gamma is the dimensionless distance of the shadow axis from the geocentre in Earth radii.
	Gamma float64
	// Uncertain reports whether the event is uncertain.
	Uncertain bool
}

// CulminationKind classifies upper and lower meridian culminations.
type CulminationKind uint32

const (
	// UpperCulmination identifies the upper culmination case.
	UpperCulmination CulminationKind = iota
	// LowerCulmination identifies the lower culmination case.
	LowerCulmination
)

// MeridianTransit records a meridian transit and its altitude.
type MeridianTransit struct {
	// Time is the timestamp for this record.
	Time time.Time
	// Kind is the event or record kind.
	Kind CulminationKind
	// AltitudeDeg is the altitude deg in degrees.
	AltitudeDeg float64
}

// MoonPhaseKind classifies the four principal lunar phases.
type MoonPhaseKind uint32

const (
	// NewMoon identifies the new moon case.
	NewMoon MoonPhaseKind = iota
	// FirstQuarterMoon identifies the first quarter moon case.
	FirstQuarterMoon
	// FullMoon identifies the full moon case.
	FullMoon
	// LastQuarterMoon identifies the last quarter moon case.
	LastQuarterMoon
)

// MoonPhaseEvent records a lunar phase kind and timestamp.
type MoonPhaseEvent struct {
	// Time is the timestamp for this record.
	Time time.Time
	// Kind is the event or record kind.
	Kind MoonPhaseKind
}

// Planet identifies one of the supported Solar System planets.
type Planet uint32

const (
	// Mercury identifies the mercury case.
	Mercury Planet = iota
	// Venus identifies the venus case.
	Venus
	// Mars identifies the mars case.
	Mars
	// Jupiter identifies the jupiter case.
	Jupiter
	// Saturn identifies the saturn case.
	Saturn
	// Uranus identifies the uranus case.
	Uranus
	// Neptune identifies the neptune case.
	Neptune
)

// PlanetaryEventKind classifies planetary conjunction and opposition events.
type PlanetaryEventKind uint32

const (
	// PlanetaryConjunction identifies the planetary conjunction case.
	PlanetaryConjunction PlanetaryEventKind = iota
	// PlanetaryOpposition identifies the planetary opposition case.
	PlanetaryOpposition
)

// PlanetaryEvent records a planetary event, elongation, and timestamp.
type PlanetaryEvent struct {
	// Time is the timestamp for this record.
	Time time.Time
	// Planet identifies the planet involved in the event.
	Planet Planet
	// Kind is the event or record kind.
	Kind PlanetaryEventKind
	// ElongationDeg is the elongation deg in degrees.
	ElongationDeg float64
}

// TransitBodyKind selects the body for a meridian-transit search.
type TransitBodyKind uint32

const (
	// TransitBodySun identifies the transit body sun case.
	TransitBodySun TransitBodyKind = iota
	// TransitBodyMoon identifies the transit body moon case.
	TransitBodyMoon
	// TransitBodyPlanet identifies the transit body planet case.
	TransitBodyPlanet
)

// SeasonKind classifies the four annual seasonal events.
type SeasonKind uint32

const (
	// MarchEquinox identifies the march equinox case.
	MarchEquinox SeasonKind = iota
	// JuneSolstice identifies the june solstice case.
	JuneSolstice
	// SeptemberEquinox identifies the september equinox case.
	SeptemberEquinox
	// DecemberSolstice identifies the december solstice case.
	DecemberSolstice
)

// SeasonEvent records a season kind and timestamp.
type SeasonEvent struct {
	// Time is the timestamp for this record.
	Time time.Time
	// Kind is the event or record kind.
	Kind SeasonKind
}

// Refraction contains pressure in millibars and temperature in degrees Celsius for refraction.
type Refraction struct {
	// PressureMbar and TemperatureC are the pressure in millibars and
	// temperature in degrees Celsius used by the refraction model.
	PressureMbar, TemperatureC float64
}

// ObserveOptions configures polar motion, refraction, deflection, and aberration corrections.
type ObserveOptions struct {
	// HasPolarMotion reports whether the has polar motion field is present.
	HasPolarMotion bool
	// XPRad is the xp rad in radians; YPRad is the yp rad in radians.
	XPRad, YPRad float64
	// HasRefraction reports whether the has refraction field is present.
	HasRefraction bool
	// Refraction supplies pressure and temperature for the refraction correction.
	Refraction Refraction
	// Deflection reports whether deflection correction is enabled.
	Deflection bool
	// Aberration reports whether aberration correction is enabled.
	Aberration bool
}

// ObserveTargetKind selects the target represented by an observation.
type ObserveTargetKind uint32

const (
	// ObserveSun identifies the observe sun case.
	ObserveSun ObserveTargetKind = iota
	// ObserveMoon identifies the observe moon case.
	ObserveMoon
	// ObserveSPK identifies the observe spk case.
	ObserveSPK
	// ObserveBarycentricState identifies the observe barycentric state case.
	ObserveBarycentricState
)

// Equatorial contains equatorial right ascension, declination, and distance.
type Equatorial struct {
	// RightAscensionDeg is the right ascension deg in degrees.
	RightAscensionDeg float64
	// RightAscensionHours is the right ascension hours in hours.
	RightAscensionHours float64
	// DeclinationDeg is the declination deg in degrees.
	DeclinationDeg float64
	// DistanceKm is the distance km in kilometres.
	DistanceKm float64
}

// Horizontal contains azimuth and elevation in degrees and range in kilometres.
type Horizontal struct {
	// AzimuthDeg, ElevationDeg, and RangeKm are the azimuth and elevation in
	// degrees and range in kilometres.
	AzimuthDeg, ElevationDeg, RangeKm float64
}

// Ecliptic contains ecliptic longitude and latitude in degrees and distance in kilometres.
type Ecliptic struct {
	// LongitudeDeg, LatitudeDeg, and DistanceKm are ecliptic longitude and
	// latitude in degrees and distance in kilometres.
	LongitudeDeg, LatitudeDeg, DistanceKm float64
}

// BodyObservation contains the astrometric, apparent, horizontal, and ecliptic views of a body.
type BodyObservation struct {
	// Astrometric is the astrometric value for BodyObservation; ApparentICRS is the apparent icrs value for BodyObservation; Apparent is the apparent value for BodyObservation.
	Astrometric, ApparentICRS, Apparent Equatorial
	// Horizontal contains topocentric azimuth, elevation, and range.
	Horizontal Horizontal
	// HourAngleDeg is the hour angle deg in degrees; HourAngleHours is the hour angle hours in hours.
	HourAngleDeg, HourAngleHours float64
	// Ecliptic contains geocentric ecliptic longitude, latitude, and distance.
	Ecliptic Ecliptic
	// Reduced reports whether the observation is reduced.
	Reduced bool
}

// SurfacePoint contains a surface latitude and longitude in degrees.
type SurfacePoint struct {
	// LatitudeDeg and LongitudeDeg are the surface coordinates in degrees.
	LatitudeDeg, LongitudeDeg float64
}

func nativeAlmanacStation(value AlmanacStation) native.AlmanacStation {
	return native.AlmanacStation{LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg, AltitudeKm: value.AltitudeKm}
}

func publicBodyAzEl(value native.BodyAzEl) BodyAzEl {
	return BodyAzEl{AzimuthDeg: value.AzimuthDeg, ElevationDeg: value.ElevationDeg, RangeKm: value.RangeKm}
}

// SunMoonECI returns geocentric Sun and Moon ECI position vectors in kilometres for the supplied TDB epoch.
func SunMoonECI(ttJulianCenturies float64) ([3]float64, [3]float64, error) {
	sun, moon, err := native.SunMoonECI(ttJulianCenturies)
	return sun, moon, publicError(err)
}

// SunMoonECEF returns geocentric Sun and Moon ECEF position vectors in kilometres for the supplied epoch.
func SunMoonECEF(epoch time.Time) ([3]float64, [3]float64, error) {
	sun, moon, err := native.SunMoonECEF(epoch)
	return sun, moon, publicError(err)
}

// SunMoonECIBatch returns geocentric Sun and Moon ECI position-vector batches in kilometres.
func SunMoonECIBatch(epochs []time.Time) ([][3]float64, [][3]float64, error) {
	sun, moon, err := native.SunMoonECIBatch(append([]time.Time(nil), epochs...))
	return sun, moon, publicError(err)
}

// SunMoonECEFBatch returns geocentric Sun and Moon ECEF position-vector batches in kilometres.
func SunMoonECEFBatch(epochs []time.Time) ([][3]float64, [][3]float64, error) {
	sun, moon, err := native.SunMoonECEFBatch(append([]time.Time(nil), epochs...))
	return sun, moon, publicError(err)
}

// SunAngle returns the Sun angular separation in radians.
func SunAngle(satelliteKm, sunKm [3]float64) (float64, error) {
	value, err := native.SunAngle(satelliteKm, sunKm)
	return value, publicError(err)
}

// MoonAngle returns the Moon angular separation in radians.
func MoonAngle(satelliteKm, moonKm [3]float64) (float64, error) {
	value, err := native.MoonAngle(satelliteKm, moonKm)
	return value, publicError(err)
}

// SunElevation returns the Sun elevation angle in radians.
func SunElevation(satelliteKm, sunKm [3]float64) (float64, error) {
	value, err := native.SunElevation(satelliteKm, sunKm)
	return value, publicError(err)
}

// NutationIAU2000A returns IAU 2000A nutation offsets in radians.
func NutationIAU2000A(jdTT float64) (dpsi, deps float64, err error) {
	dpsi, deps, err = native.NutationIAU2000A(jdTT)
	return dpsi, deps, publicError(err)
}

// NutationMeanObliquity returns the mean obliquity in radians.
func NutationMeanObliquity(jdTDB float64) (float64, error) {
	value, err := native.NutationMeanObliquity(jdTDB)
	return value, publicError(err)
}

// NutationFundamentalArguments returns the five nutation fundamental arguments in radians.
func NutationFundamentalArguments(tJulianCenturies float64) ([5]float64, error) {
	value, err := native.NutationFundamentalArguments(tJulianCenturies)
	return value, publicError(err)
}

// NutationEquationOfEquinoxesTerms returns the equation-of-equinoxes terms in radians.
func NutationEquationOfEquinoxesTerms(jdTT float64) (float64, error) {
	value, err := native.NutationEquationOfEquinoxesTerms(jdTT)
	return value, publicError(err)
}

// NutationMatrix returns the nutation rotation matrix.
func NutationMatrix(meanObliquityRad, trueObliquityRad, psiRad float64) ([3][3]float64, error) {
	value, err := native.NutationMatrix(meanObliquityRad, trueObliquityRad, psiRad)
	return value, publicError(err)
}

// PrecessionMatrix returns the precession rotation matrix.
func PrecessionMatrix(jdTDB float64) ([3][3]float64, error) {
	value, err := native.PrecessionMatrix(jdTDB)
	return value, publicError(err)
}

// PrecessionICRSToJ2000Matrix returns the ICRS-to-J2000 precession rotation matrix.
func PrecessionICRSToJ2000Matrix() ([3][3]float64, error) {
	value, err := native.PrecessionICRSToJ2000Matrix()
	return value, publicError(err)
}

// SunAzEl returns Sun azimuth and elevation in degrees and range in kilometres.
func SunAzEl(station AlmanacStation, epoch time.Time) (BodyAzEl, error) {
	value, err := native.SunAzEl(nativeAlmanacStation(station), epoch)
	return publicBodyAzEl(value), publicError(err)
}

// MoonAzEl returns Moon azimuth and elevation in degrees and range in kilometres.
func MoonAzEl(station AlmanacStation, epoch time.Time) (BodyAzEl, error) {
	value, err := native.MoonAzEl(nativeAlmanacStation(station), epoch)
	return publicBodyAzEl(value), publicError(err)
}

// MoonElevation returns the Moon elevation angle in degrees.
func MoonElevation(station AlmanacStation, epoch time.Time) (float64, error) {
	value, err := native.MoonElevation(nativeAlmanacStation(station), epoch)
	return value, publicError(err)
}

// MoonIlluminationAt returns the Moon illuminated fraction and phase angle.
func MoonIlluminationAt(station AlmanacStation, epoch time.Time) (MoonIllumination, error) {
	value, err := native.MoonIllumination(nativeAlmanacStation(station), epoch)
	return MoonIllumination{IlluminatedFraction: value.IlluminatedFraction, PhaseAngleDeg: value.PhaseAngleDeg}, publicError(err)
}

// DefaultMoonElevationOptions returns the default Moon-elevation search options.
func DefaultMoonElevationOptions() (MoonElevationOptions, error) {
	value, err := native.MoonElevationOptionsDefaults()
	return MoonElevationOptions{ElevationThresholdDeg: value.ElevationThresholdDeg, StepSeconds: value.StepSeconds, TimeToleranceSeconds: value.TimeToleranceSeconds}, publicError(err)
}

// FindMoonElevationCrossings returns Moon elevation threshold crossings over the interval.
func FindMoonElevationCrossings(station AlmanacStation, start, end time.Time, options *MoonElevationOptions) ([]MoonElevationCrossing, error) {
	var nativeOptions *native.MoonElevationOptions
	if options != nil {
		copy := native.MoonElevationOptions{ElevationThresholdDeg: options.ElevationThresholdDeg, StepSeconds: options.StepSeconds, TimeToleranceSeconds: options.TimeToleranceSeconds}
		nativeOptions = &copy
	}
	values, err := native.FindMoonCrossings(nativeAlmanacStation(station), start, end, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]MoonElevationCrossing, len(values))
	for i := range out {
		out[i] = MoonElevationCrossing{Time: values[i].Time, Kind: MoonRiseSetKind(values[i].Kind), ElevationDeg: values[i].ElevationDeg}
	}
	return out, nil
}

// FindMoonTransits returns Moon meridian-transit events over the interval.
func FindMoonTransits(station AlmanacStation, start, end time.Time, stepSeconds, toleranceSeconds float64) ([]MoonTransit, error) {
	values, err := native.FindMoonTransits(nativeAlmanacStation(station), start, end, stepSeconds, toleranceSeconds)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]MoonTransit, len(values))
	for i := range out {
		out[i] = MoonTransit{Time: values[i].Time, Kind: MoonTransitKind(values[i].Kind), ElevationDeg: values[i].ElevationDeg}
	}
	return out, nil
}

func publicObserveOptions(value native.ObserveOptions) ObserveOptions {
	return ObserveOptions{HasPolarMotion: value.HasPolarMotion, XPRad: value.XPRad, YPRad: value.YPRad, HasRefraction: value.HasRefraction, Refraction: Refraction{PressureMbar: value.Refraction.PressureMbar, TemperatureC: value.Refraction.TemperatureC}, Deflection: value.Deflection, Aberration: value.Aberration}
}

// DefaultObserveOptions returns the default observation options.
func DefaultObserveOptions() (ObserveOptions, error) {
	value, err := native.ObserveOptionsDefaults()
	return publicObserveOptions(value), publicError(err)
}

func publicObservation(value native.BodyObservation) BodyObservation {
	return BodyObservation{Astrometric: Equatorial{RightAscensionDeg: value.Astrometric.RightAscensionDeg, RightAscensionHours: value.Astrometric.RightAscensionHours, DeclinationDeg: value.Astrometric.DeclinationDeg, DistanceKm: value.Astrometric.DistanceKm}, ApparentICRS: Equatorial{RightAscensionDeg: value.ApparentICRS.RightAscensionDeg, RightAscensionHours: value.ApparentICRS.RightAscensionHours, DeclinationDeg: value.ApparentICRS.DeclinationDeg, DistanceKm: value.ApparentICRS.DistanceKm}, Apparent: Equatorial{RightAscensionDeg: value.Apparent.RightAscensionDeg, RightAscensionHours: value.Apparent.RightAscensionHours, DeclinationDeg: value.Apparent.DeclinationDeg, DistanceKm: value.Apparent.DistanceKm}, Horizontal: Horizontal{AzimuthDeg: value.Horizontal.AzimuthDeg, ElevationDeg: value.Horizontal.ElevationDeg, RangeKm: value.Horizontal.RangeKm}, HourAngleDeg: value.HourAngleDeg, HourAngleHours: value.HourAngleHours, Ecliptic: Ecliptic{LongitudeDeg: value.Ecliptic.LongitudeDeg, LatitudeDeg: value.Ecliptic.LatitudeDeg, DistanceKm: value.Ecliptic.DistanceKm}, Reduced: value.Reduced}
}

// ObserveBody evaluates one typed observation source. For SPK and
// barycentric-state targets, spk must be non-nil; position and velocity are
// copied before entering C and may be nil when the selected target does not
// use them.
func ObserveBody(station AlmanacStation, epoch time.Time, target ObserveTargetKind, spk *SPK, naifID int32, positionKm, velocityKmPerS *[3]float64, options *ObserveOptions) (BodyObservation, error) {
	var nativeOptions *native.ObserveOptions
	if options != nil {
		copy := native.ObserveOptions{HasPolarMotion: options.HasPolarMotion, XPRad: options.XPRad, YPRad: options.YPRad, HasRefraction: options.HasRefraction, Refraction: native.Refraction{PressureMbar: options.Refraction.PressureMbar, TemperatureC: options.Refraction.TemperatureC}, Deflection: options.Deflection, Aberration: options.Aberration}
		nativeOptions = &copy
	}
	value, err := native.Observe(nativeAlmanacStation(station), epoch, uint32(target), func() *native.SPK {
		if spk == nil {
			return nil
		}
		return spk.handle
	}(), naifID, positionKm, velocityKmPerS, nativeOptions)
	return publicObservation(value), publicError(err)
}

// ObserveSPKBody returns the observation of one SPK body from the station and epoch.
func ObserveSPKBody(station AlmanacStation, epoch time.Time, spk *SPK, naifID int32) (BodyObservation, error) {
	if spk == nil || spk.handle == nil {
		return BodyObservation{}, ErrClosed
	}
	value, err := native.ObserveSPKBody(nativeAlmanacStation(station), epoch, spk.handle, naifID)
	return publicObservation(value), publicError(err)
}

// AlmanacSeasons returns seasonal events over the interval.
func AlmanacSeasons(spk *SPK, start, end time.Time, stepSeconds, toleranceSeconds float64) ([]SeasonEvent, error) {
	if spk == nil || spk.handle == nil {
		return nil, ErrClosed
	}
	values, err := native.AlmanacSeasons(spk.handle, start, end, stepSeconds, toleranceSeconds)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]SeasonEvent, len(values))
	for i := range out {
		out[i] = SeasonEvent{Time: values[i].Time, Kind: SeasonKind(values[i].Kind)}
	}
	return out, nil
}

// AlmanacMoonPhases returns lunar phase events over the interval.
func AlmanacMoonPhases(spk *SPK, start, end time.Time, stepSeconds, toleranceSeconds float64) ([]MoonPhaseEvent, error) {
	if spk == nil || spk.handle == nil {
		return nil, ErrClosed
	}
	values, err := native.AlmanacMoonPhases(spk.handle, start, end, stepSeconds, toleranceSeconds)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]MoonPhaseEvent, len(values))
	for i := range out {
		out[i] = MoonPhaseEvent{Time: values[i].Time, Kind: MoonPhaseKind(values[i].Kind)}
	}
	return out, nil
}

// AlmanacPlanetaryEvents returns planetary events over the interval.
func AlmanacPlanetaryEvents(spk *SPK, planet Planet, kind PlanetaryEventKind, start, end time.Time, stepSeconds, toleranceSeconds float64) ([]PlanetaryEvent, error) {
	if spk == nil || spk.handle == nil {
		return nil, ErrClosed
	}
	values, err := native.AlmanacPlanetaryEvents(spk.handle, uint32(planet), uint32(kind), start, end, stepSeconds, toleranceSeconds)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]PlanetaryEvent, len(values))
	for i := range out {
		out[i] = PlanetaryEvent{Time: values[i].Time, Planet: Planet(values[i].Planet), Kind: PlanetaryEventKind(values[i].Kind), ElongationDeg: values[i].ElongationDeg}
	}
	return out, nil
}

// AlmanacEclipses returns eclipse events over the interval.
func AlmanacEclipses(spk *SPK, start, end time.Time, stepSeconds, toleranceSeconds float64) ([]EclipseEvent, error) {
	if spk == nil || spk.handle == nil {
		return nil, ErrClosed
	}
	values, err := native.AlmanacEclipses(spk.handle, start, end, stepSeconds, toleranceSeconds)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]EclipseEvent, len(values))
	for i := range out {
		out[i] = EclipseEvent{Time: values[i].Time, Kind: AlmanacEclipseKind(values[i].Kind), Magnitude: values[i].Magnitude, MoonLatitudeDeg: values[i].MoonLatitudeDeg, Gamma: values[i].Gamma, Uncertain: values[i].Uncertain}
	}
	return out, nil
}

// AlmanacMeridianTransits returns meridian-transit events over the interval.
func AlmanacMeridianTransits(spk *SPK, station AlmanacStation, body TransitBodyKind, planet Planet, start, end time.Time, stepSeconds, toleranceSeconds float64) ([]MeridianTransit, error) {
	if spk == nil || spk.handle == nil {
		return nil, ErrClosed
	}
	values, err := native.AlmanacMeridianTransits(spk.handle, nativeAlmanacStation(station), uint32(body), uint32(planet), start, end, stepSeconds, toleranceSeconds)
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]MeridianTransit, len(values))
	for i := range out {
		out[i] = MeridianTransit{Time: values[i].Time, Kind: CulminationKind(values[i].Kind), AltitudeDeg: values[i].AltitudeDeg}
	}
	return out, nil
}

// OceanTideLoading returns the ocean-loading displacement in metres.
func OceanTideLoading(stationECEFM [3]float64, year, month, day int, fractionalHour float64, loading OceanLoadingBLQ) ([3]float64, error) {
	value, err := native.OceanTideLoading(stationECEFM, year, month, day, fractionalHour, native.OceanLoadingBLQ{AmplitudeM: loading.AmplitudeM, PhaseDeg: loading.PhaseDeg})
	return value, publicError(err)
}

// OceanLoadingBLQ contains BLQ ocean-loading amplitudes in metres and phases in degrees.
type OceanLoadingBLQ struct {
	// AmplitudeM is the amplitude m in metres.
	AmplitudeM [3][11]float64
	// PhaseDeg is the phase deg in degrees.
	PhaseDeg [3][11]float64
}

// SolidEarthTide returns the solid-Earth-tide displacement in metres.
func SolidEarthTide(stationECEFM, sunECEFM, moonECEFM [3]float64, year, month, day int, fractionalHour float64) ([3]float64, error) {
	value, err := native.SolidEarthTide(stationECEFM, sunECEFM, moonECEFM, year, month, day, fractionalHour)
	return value, publicError(err)
}

// SolidEarthPoleTide returns the pole-tide displacement in metres.
func SolidEarthPoleTide(stationECEFM [3]float64, year, month, day int, fractionalHour, xpArcsec, ypArcsec float64) ([3]float64, error) {
	value, err := native.SolidEarthPoleTide(stationECEFM, year, month, day, fractionalHour, xpArcsec, ypArcsec)
	return value, publicError(err)
}

// SubSolarPoint returns the subsolar latitude and longitude in degrees.
func SubSolarPoint(sunECEF [3]float64) (SurfacePoint, error) {
	value, err := native.SubSolarPoint(sunECEF)
	return SurfacePoint{LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg}, publicError(err)
}

// SubObserverPoint returns the subobserver latitude and longitude in degrees.
func SubObserverPoint(observerFromBody [3]float64, poleRA, poleDec, primeMeridian float64) (SurfacePoint, error) {
	value, err := native.SubObserverPoint(observerFromBody, poleRA, poleDec, primeMeridian)
	return SurfacePoint{LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg}, publicError(err)
}

// TerminatorLatitude returns the terminator latitude in radians.
func TerminatorLatitude(subSolarLat, subSolarLon, longitude float64) (float64, error) {
	value, err := native.TerminatorLatitude(subSolarLat, subSolarLon, longitude)
	return value, publicError(err)
}

// ParallacticAngle returns the parallactic angle in radians.
func ParallacticAngle(observerLat, hourAngle, declination float64) (float64, error) {
	value, err := native.ParallacticAngle(observerLat, hourAngle, declination)
	return value, publicError(err)
}

// PhaseAngle returns the phase angle in radians.
func PhaseAngle(satelliteKm, sunKm, observerKm [3]float64) (float64, error) {
	value, err := native.PhaseAngle(satelliteKm, sunKm, observerKm)
	return value, publicError(err)
}

// PositionAngle returns the position angle in radians.
func PositionAngle(fromLon, fromLat, toLon, toLat float64) (float64, error) {
	value, err := native.PositionAngle(fromLon, fromLat, toLon, toLat)
	return value, publicError(err)
}
