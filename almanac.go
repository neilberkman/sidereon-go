package sidereon

import (
	"time"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// AlmanacStation uses degrees and kilometres, matching the astronomy C
// station type. It is distinct from GroundStation, whose height is metres.
type AlmanacStation struct {
	LatitudeDeg  float64
	LongitudeDeg float64
	AltitudeKm   float64
}

type BodyAzEl struct {
	AzimuthDeg   float64
	ElevationDeg float64
	RangeKm      float64
}

type MoonIllumination struct {
	IlluminatedFraction float64
	PhaseAngleDeg       float64
}

type MoonElevationOptions struct {
	ElevationThresholdDeg float64
	StepSeconds           float64
	TimeToleranceSeconds  float64
}

type MoonElevationCrossing struct {
	Time         time.Time
	Kind         MoonRiseSetKind
	ElevationDeg float64
}

type MoonRiseSetKind uint32

const (
	MoonRising MoonRiseSetKind = iota
	MoonSetting
)

type MoonTransit struct {
	Time         time.Time
	Kind         MoonTransitKind
	ElevationDeg float64
}

type MoonTransitKind uint32

const (
	MoonUpperCulmination MoonTransitKind = iota
	MoonLowerCulmination
)

type AlmanacEclipseKind uint32

const (
	LunarPenumbralEclipse AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseLunarPenumbral)
	LunarPartialEclipse   AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseLunarPartial)
	LunarTotalEclipse     AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseLunarTotal)
	SolarPartialEclipse   AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarPartial)
	SolarAnnularEclipse   AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarAnnular)
	SolarTotalEclipse     AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarTotal)
	SolarHybridEclipse    AlmanacEclipseKind = AlmanacEclipseKind(native.AlmanacEclipseSolarHybrid)
)

type EclipseEvent struct {
	Time            time.Time
	Kind            AlmanacEclipseKind
	Magnitude       float64
	MoonLatitudeDeg float64
	Gamma           float64
	Uncertain       bool
}

type CulminationKind uint32

const (
	UpperCulmination CulminationKind = iota
	LowerCulmination
)

type MeridianTransit struct {
	Time        time.Time
	Kind        CulminationKind
	AltitudeDeg float64
}

type MoonPhaseKind uint32

const (
	NewMoon MoonPhaseKind = iota
	FirstQuarterMoon
	FullMoon
	LastQuarterMoon
)

type MoonPhaseEvent struct {
	Time time.Time
	Kind MoonPhaseKind
}

type Planet uint32

const (
	Mercury Planet = iota
	Venus
	Mars
	Jupiter
	Saturn
	Uranus
	Neptune
)

type PlanetaryEventKind uint32

const (
	PlanetaryConjunction PlanetaryEventKind = iota
	PlanetaryOpposition
)

type PlanetaryEvent struct {
	Time          time.Time
	Planet        Planet
	Kind          PlanetaryEventKind
	ElongationDeg float64
}

// TransitBodyKind selects the body for a meridian-transit search.
type TransitBodyKind uint32

const (
	TransitBodySun TransitBodyKind = iota
	TransitBodyMoon
	TransitBodyPlanet
)

type SeasonKind uint32

const (
	MarchEquinox SeasonKind = iota
	JuneSolstice
	SeptemberEquinox
	DecemberSolstice
)

type SeasonEvent struct {
	Time time.Time
	Kind SeasonKind
}

type Refraction struct{ PressureMbar, TemperatureC float64 }
type ObserveOptions struct {
	HasPolarMotion bool
	XPRad, YPRad   float64
	HasRefraction  bool
	Refraction     Refraction
	Deflection     bool
	Aberration     bool
}

type ObserveTargetKind uint32

const (
	ObserveSun ObserveTargetKind = iota
	ObserveMoon
	ObserveSPK
	ObserveBarycentricState
)

type Equatorial struct {
	RightAscensionDeg   float64
	RightAscensionHours float64
	DeclinationDeg      float64
	DistanceKm          float64
}
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

func nativeAlmanacStation(value AlmanacStation) native.AlmanacStation {
	return native.AlmanacStation{LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg, AltitudeKm: value.AltitudeKm}
}

func publicBodyAzEl(value native.BodyAzEl) BodyAzEl {
	return BodyAzEl{AzimuthDeg: value.AzimuthDeg, ElevationDeg: value.ElevationDeg, RangeKm: value.RangeKm}
}

func SunMoonECI(ttJulianCenturies float64) ([3]float64, [3]float64, error) {
	sun, moon, err := native.SunMoonECI(ttJulianCenturies)
	return sun, moon, publicError(err)
}
func SunMoonECEF(epoch time.Time) ([3]float64, [3]float64, error) {
	sun, moon, err := native.SunMoonECEF(epoch)
	return sun, moon, publicError(err)
}
func SunMoonECIBatch(epochs []time.Time) ([][3]float64, [][3]float64, error) {
	sun, moon, err := native.SunMoonECIBatch(append([]time.Time(nil), epochs...))
	return sun, moon, publicError(err)
}
func SunMoonECEFBatch(epochs []time.Time) ([][3]float64, [][3]float64, error) {
	sun, moon, err := native.SunMoonECEFBatch(append([]time.Time(nil), epochs...))
	return sun, moon, publicError(err)
}

func SunAngle(satelliteKm, sunKm [3]float64) (float64, error) {
	value, err := native.SunAngle(satelliteKm, sunKm)
	return value, publicError(err)
}

func MoonAngle(satelliteKm, moonKm [3]float64) (float64, error) {
	value, err := native.MoonAngle(satelliteKm, moonKm)
	return value, publicError(err)
}

func SunElevation(satelliteKm, sunKm [3]float64) (float64, error) {
	value, err := native.SunElevation(satelliteKm, sunKm)
	return value, publicError(err)
}

func NutationIAU2000A(jdTT float64) (dpsi, deps float64, err error) {
	dpsi, deps, err = native.NutationIAU2000A(jdTT)
	return dpsi, deps, publicError(err)
}
func NutationMeanObliquity(jdTDB float64) (float64, error) {
	value, err := native.NutationMeanObliquity(jdTDB)
	return value, publicError(err)
}
func NutationFundamentalArguments(tJulianCenturies float64) ([5]float64, error) {
	value, err := native.NutationFundamentalArguments(tJulianCenturies)
	return value, publicError(err)
}
func NutationEquationOfEquinoxesTerms(jdTT float64) (float64, error) {
	value, err := native.NutationEquationOfEquinoxesTerms(jdTT)
	return value, publicError(err)
}
func NutationMatrix(meanObliquityRad, trueObliquityRad, psiRad float64) ([3][3]float64, error) {
	value, err := native.NutationMatrix(meanObliquityRad, trueObliquityRad, psiRad)
	return value, publicError(err)
}
func PrecessionMatrix(jdTDB float64) ([3][3]float64, error) {
	value, err := native.PrecessionMatrix(jdTDB)
	return value, publicError(err)
}
func PrecessionICRSToJ2000Matrix() ([3][3]float64, error) {
	value, err := native.PrecessionICRSToJ2000Matrix()
	return value, publicError(err)
}

func SunAzEl(station AlmanacStation, epoch time.Time) (BodyAzEl, error) {
	value, err := native.SunAzEl(nativeAlmanacStation(station), epoch)
	return publicBodyAzEl(value), publicError(err)
}
func MoonAzEl(station AlmanacStation, epoch time.Time) (BodyAzEl, error) {
	value, err := native.MoonAzEl(nativeAlmanacStation(station), epoch)
	return publicBodyAzEl(value), publicError(err)
}
func MoonElevation(station AlmanacStation, epoch time.Time) (float64, error) {
	value, err := native.MoonElevation(nativeAlmanacStation(station), epoch)
	return value, publicError(err)
}
func MoonIlluminationAt(station AlmanacStation, epoch time.Time) (MoonIllumination, error) {
	value, err := native.MoonIllumination(nativeAlmanacStation(station), epoch)
	return MoonIllumination{IlluminatedFraction: value.IlluminatedFraction, PhaseAngleDeg: value.PhaseAngleDeg}, publicError(err)
}
func DefaultMoonElevationOptions() (MoonElevationOptions, error) {
	value, err := native.MoonElevationOptionsDefaults()
	return MoonElevationOptions{ElevationThresholdDeg: value.ElevationThresholdDeg, StepSeconds: value.StepSeconds, TimeToleranceSeconds: value.TimeToleranceSeconds}, publicError(err)
}

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

func ObserveSPKBody(station AlmanacStation, epoch time.Time, spk *SPK, naifID int32) (BodyObservation, error) {
	if spk == nil || spk.handle == nil {
		return BodyObservation{}, ErrClosed
	}
	value, err := native.ObserveSPKBody(nativeAlmanacStation(station), epoch, spk.handle, naifID)
	return publicObservation(value), publicError(err)
}

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

func OceanTideLoading(stationECEFM [3]float64, year, month, day int, fractionalHour float64, loading OceanLoadingBLQ) ([3]float64, error) {
	value, err := native.OceanTideLoading(stationECEFM, year, month, day, fractionalHour, native.OceanLoadingBLQ{AmplitudeM: loading.AmplitudeM, PhaseDeg: loading.PhaseDeg})
	return value, publicError(err)
}

type OceanLoadingBLQ struct {
	AmplitudeM [3][11]float64
	PhaseDeg   [3][11]float64
}

func SolidEarthTide(stationECEFM, sunECEFM, moonECEFM [3]float64, year, month, day int, fractionalHour float64) ([3]float64, error) {
	value, err := native.SolidEarthTide(stationECEFM, sunECEFM, moonECEFM, year, month, day, fractionalHour)
	return value, publicError(err)
}
func SolidEarthPoleTide(stationECEFM [3]float64, year, month, day int, fractionalHour, xpArcsec, ypArcsec float64) ([3]float64, error) {
	value, err := native.SolidEarthPoleTide(stationECEFM, year, month, day, fractionalHour, xpArcsec, ypArcsec)
	return value, publicError(err)
}
func SubSolarPoint(sunECEF [3]float64) (SurfacePoint, error) {
	value, err := native.SubSolarPoint(sunECEF)
	return SurfacePoint{LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg}, publicError(err)
}
func SubObserverPoint(observerFromBody [3]float64, poleRA, poleDec, primeMeridian float64) (SurfacePoint, error) {
	value, err := native.SubObserverPoint(observerFromBody, poleRA, poleDec, primeMeridian)
	return SurfacePoint{LatitudeDeg: value.LatitudeDeg, LongitudeDeg: value.LongitudeDeg}, publicError(err)
}
func TerminatorLatitude(subSolarLat, subSolarLon, longitude float64) (float64, error) {
	value, err := native.TerminatorLatitude(subSolarLat, subSolarLon, longitude)
	return value, publicError(err)
}
func ParallacticAngle(observerLat, hourAngle, declination float64) (float64, error) {
	value, err := native.ParallacticAngle(observerLat, hourAngle, declination)
	return value, publicError(err)
}
func PhaseAngle(satelliteKm, sunKm, observerKm [3]float64) (float64, error) {
	value, err := native.PhaseAngle(satelliteKm, sunKm, observerKm)
	return value, publicError(err)
}
func PositionAngle(fromLon, fromLat, toLon, toLat float64) (float64, error) {
	value, err := native.PositionAngle(fromLon, fromLat, toLon, toLat)
	return value, publicError(err)
}
