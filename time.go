package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// CivilDateTime is a UTC-like proleptic-Gregorian calendar instant. Second
// may contain a fractional part.
type CivilDateTime struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second float64
}

// JulianDate is the split Julian date used by the C time-scale APIs.
type JulianDate struct {
	Whole    float64
	Fraction float64
}

// TimeScales contains the split Julian-date values returned by the C frame
// helpers.
type TimeScales struct {
	JDWhole     float64
	UT1Fraction float64
	TTFraction  float64
	TDBFraction float64
	JDUT1       float64
	JDTT        float64
	JDTDB       float64
}

// GNSSWeekSeconds is the result of splitting continuous seconds into a GNSS
// week and seconds within that week.
type GNSSWeekSeconds struct {
	Week          float64
	SecondsOfWeek float64
}

// TimeScale identifies a C time scale.
type TimeScale uint32

const (
	UTC      TimeScale = 0
	TAI      TimeScale = 1
	TT       TimeScale = 2
	TDB      TimeScale = 3
	GPST     TimeScale = 4
	GST      TimeScale = 5
	BDT      TimeScale = 6
	GLONASST TimeScale = 7
	QZSST    TimeScale = 8
	TCG      TimeScale = 9
	TCB      TimeScale = 10
)

// CivilToJ2000Seconds returns continuous UTC-like seconds past J2000.
func CivilToJ2000Seconds(value CivilDateTime) (float64, error) {
	result, err := native.CivilToJ2000(native.CivilDateTime{
		Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second,
	})
	return result, publicError(err)
}

// J2000SecondsToCivil converts integer J2000 seconds to civil calendar fields.
func J2000SecondsToCivil(seconds int64) (CivilDateTime, error) {
	value, err := native.J2000ToCivil(seconds)
	return CivilDateTime{
		Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second,
	}, publicError(err)
}

// CivilToGPSSeconds returns GPS seconds and the C library's availability flag.
// A civil date outside the supported GPS conversion range can return
// available=false with a nil error, as specified by the C ABI.
func CivilToGPSSeconds(value CivilDateTime) (seconds float64, available bool, err error) {
	seconds, available, err = native.CivilToGPS(native.CivilDateTime{
		Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second,
	})
	return seconds, available, publicError(err)
}

// InstantFromUTCCivil returns the split Julian date and continuous J2000
// seconds for a civil UTC instant.
func InstantFromUTCCivil(value CivilDateTime) (JulianDate, float64, error) {
	date, seconds, err := native.InstantFromUTC(native.CivilDateTime{
		Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second,
	})
	return JulianDate{Whole: date.Whole, Fraction: date.Fraction}, seconds, publicError(err)
}

// SplitJulianDateToJ2000Seconds converts a split Julian date to J2000 seconds.
func SplitJulianDateToJ2000Seconds(date JulianDate) (float64, error) {
	seconds, err := native.SplitJulianToJ2000(native.JulianDate{Whole: date.Whole, Fraction: date.Fraction})
	return seconds, publicError(err)
}

// LeapSeconds returns TAI-UTC at UTC midnight for the supplied calendar date.
func LeapSeconds(year, month, day int) (float64, error) {
	seconds, err := native.LeapSeconds(year, month, day)
	return seconds, publicError(err)
}

// GPSUTCOffset returns GPS-UTC seconds at a UTC Julian date.
func GPSUTCOffset(julianDate float64) (float64, error) {
	seconds, err := native.GPSUTCOffset(julianDate)
	return seconds, publicError(err)
}

// TAIUTCOffset returns TAI-UTC seconds at a UTC Julian date.
func TAIUTCOffset(julianDate float64) (float64, error) {
	seconds, err := native.TAIUTCOffset(julianDate)
	return seconds, publicError(err)
}

// GNSSSecondsOfWeek returns whole-second seconds-of-week for a calendar date.
func GNSSSecondsOfWeek(value CivilDateTime) (float64, error) {
	seconds, err := native.GNSSSecondsOfWeek(native.CivilDateTime{
		Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second,
	})
	return seconds, publicError(err)
}

// GNSSWeekAndSecondsOfWeek splits continuous seconds into week and
// seconds-of-week values.
func GNSSWeekAndSecondsOfWeek(continuousSeconds float64) (GNSSWeekSeconds, error) {
	value, err := native.GNSSWeekAndSeconds(continuousSeconds)
	return GNSSWeekSeconds{Week: value.Week, SecondsOfWeek: value.SecondsOfWeek}, publicError(err)
}

// TimeScalesFromUTC resolves all C frame time scales for a civil UTC instant.
func TimeScalesFromUTC(value CivilDateTime) (TimeScales, error) {
	result, err := native.TimeScalesFromUTC(native.CivilDateTime{
		Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second,
	})
	return TimeScales{
		JDWhole: result.JDWhole, UT1Fraction: result.UT1Fraction, TTFraction: result.TTFraction,
		TDBFraction: result.TDBFraction, JDUT1: result.JDUT1, JDTT: result.JDTT, JDTDB: result.JDTDB,
	}, publicError(err)
}

// TimeScaleLabel returns the C label such as "GPST" for a time scale.
func TimeScaleLabel(scale TimeScale) (string, error) {
	value, err := native.TimeScaleLabel(uint32(scale))
	return string(value), publicError(err)
}

// TimeScaleOffset returns the fixed offset to-reading minus from-reading in
// seconds. UTC-based and TDB routes are rejected by the C library because they
// require an epoch or a different model.
func TimeScaleOffset(from, to TimeScale) (float64, error) {
	value, err := native.TimeScaleOffset(uint32(from), uint32(to))
	return value, publicError(err)
}

// TimeScaleOffsetAt returns the leap-aware offset at a UTC Julian date.
func TimeScaleOffsetAt(from, to TimeScale, utcJulianDate float64) (float64, error) {
	value, err := native.TimeScaleOffsetAt(uint32(from), uint32(to), utcJulianDate)
	return value, publicError(err)
}
