package sidereon

import (
	"errors"
	"unsafe"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// AntennaPCO is a north/east/up phase-center offset in metres.
type AntennaPCO struct {
	// NorthM is the north m in metres.
	NorthM float64
	// EastM is the east m in metres.
	EastM float64
	// UpM is the up m in metres.
	UpM float64
}

// Antenna owns one parsed ANTEX antenna block. Read calls may run
// concurrently; Close waits for active calls and is idempotent. The value
// must not be copied after first use.
type Antenna struct {
	_      noCopy
	native *native.Antenna
}

// ANTEX owns a parsed ANTEX 1.4 product. Read calls may run concurrently;
// Close waits for active calls and is idempotent. The value must not be copied
// after first use.
type ANTEX struct {
	_      noCopy
	native *native.ANTEX
}

// ParseANTEX parses ANTEX bytes with the native parser.
func ParseANTEX(data []byte) (*ANTEX, error) {
	value, err := native.ParseANTEX(data)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native ANTEX constructor returned no handle")
	}
	return &ANTEX{native: value}, nil
}

// Close releases the parsed ANTEX product and is idempotent.
func (a *ANTEX) Close() error {
	if a == nil || a.native == nil {
		return nil
	}
	return publicError(a.native.Close())
}

// AntennaCount returns the number of antenna blocks in the product.
func (a *ANTEX) AntennaCount() (int, error) {
	if a == nil || a.native == nil {
		return 0, ErrClosed
	}
	value, err := a.native.AntennaCount()
	return value, publicError(err)
}

// Antenna looks up an exact ANTEX TYPE / SERIAL identifier. found is false
// when the C product has no matching block.
func (a *ANTEX) Antenna(id string) (*Antenna, bool, error) {
	if a == nil || a.native == nil {
		return nil, false, ErrClosed
	}
	value, found, err := a.native.Antenna(id)
	if err != nil {
		return nil, false, publicError(err)
	}
	if !found {
		return nil, false, nil
	}
	if value == nil {
		return nil, false, errors.New("sidereon: native ANTEX antenna lookup returned no handle")
	}
	return &Antenna{native: value}, true, nil
}

// Encode serializes the parsed ANTEX product into newly owned bytes.
func (a *ANTEX) Encode() ([]byte, error) {
	if a == nil || a.native == nil {
		return nil, ErrClosed
	}
	value, err := a.native.Encode()
	return value, publicError(err)
}

// Close releases the antenna block and is idempotent.
func (a *Antenna) Close() error {
	if a == nil || a.native == nil {
		return nil
	}
	return publicError(a.native.Close())
}

// PCO returns the frequency-dependent phase-center offset in metres.
func (a *Antenna) PCO(frequency string) (AntennaPCO, error) {
	if a == nil || a.native == nil {
		return AntennaPCO{}, ErrClosed
	}
	value, err := a.native.PCO(frequency)
	return AntennaPCO{NorthM: value.NorthM, EastM: value.EastM, UpM: value.UpM}, publicError(err)
}

// PCV returns the frequency-dependent phase-center variation in metres.
func (a *Antenna) PCV(frequency string, zenithDeg float64, azimuthDeg *float64) (float64, error) {
	if a == nil || a.native == nil {
		return 0, ErrClosed
	}
	var azimuth float64
	if azimuthDeg != nil {
		azimuth = *azimuthDeg
	}
	value, err := a.native.PCV(frequency, zenithDeg, azimuthDeg != nil, azimuth)
	return value, publicError(err)
}

// SpaceWeather contains the three native solar/geomagnetic indices.
type SpaceWeather struct {
	// F107 is the 10.7 cm solar radio flux.
	F107 float64
	// F107A is the adjusted 10.7 cm solar radio flux.
	F107A float64
	// Ap is the planetary geomagnetic index.
	Ap float64
}

// SpaceWeatherObservationClass identifies the provenance of a table record.
// The values mirror the C source enum: observed, interpolated, daily
// predicted, and monthly predicted.
type SpaceWeatherObservationClass uint32

const (
	// SpaceWeatherObservationObserved identifies the space weather observation observed case.
	SpaceWeatherObservationObserved SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationObservedValue)
	// SpaceWeatherObservationInterpolated identifies the space weather observation interpolated case.
	SpaceWeatherObservationInterpolated SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationInterpolatedValue)
	// SpaceWeatherObservationDailyPredicted identifies the space weather observation daily predicted case.
	SpaceWeatherObservationDailyPredicted SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationDailyPredictedValue)
	// SpaceWeatherObservationMonthlyPredicted identifies the space weather observation monthly predicted case.
	SpaceWeatherObservationMonthlyPredicted SpaceWeatherObservationClass = SpaceWeatherObservationClass(native.SpaceWeatherObservationMonthlyPredictedValue)
)

// DefaultSpaceWeather returns the native quiet-Sun defaults.
func DefaultSpaceWeather() (SpaceWeather, error) {
	value, err := native.DefaultSpaceWeather()
	return SpaceWeather{F107: value.F107, F107A: value.F107A, Ap: value.Ap}, publicError(err)
}

// SpaceWeatherCoverage describes observed and predicted coverage boundaries.
type SpaceWeatherCoverage struct {
	// FirstJ2000S is the first j2000 s in seconds.
	FirstJ2000S float64
	// HasLastObservedJ2000S reports whether the has last observed j2000 s field is present.
	HasLastObservedJ2000S bool
	// LastObservedJ2000S is the last observed j2000 s in seconds.
	LastObservedJ2000S float64
	// HasLastDailyPredictedJ2000S reports whether the has last daily predicted j2000 s field is present.
	HasLastDailyPredictedJ2000S bool
	// LastDailyPredictedJ2000S is the last daily predicted j2000 s in seconds.
	LastDailyPredictedJ2000S float64
	// EndJ2000S is the end j2000 s in seconds.
	EndJ2000S float64
}

// SpaceWeatherDay preserves one complete parsed daily or monthly record,
// including explicit presence for every optional field.
type SpaceWeatherDay struct {
	// Year is the calendar year.
	Year int
	// Month is the calendar month.
	Month uint8
	// Day is the calendar day.
	Day uint8
	// Class is the product or solution class.
	Class SpaceWeatherObservationClass
	// HasBSRN reports whether the has bsrn field is present.
	HasBSRN bool
	// BSRN is the Bartels solar rotation number.
	BSRN uint16
	// HasND reports whether the has nd field is present.
	HasND bool
	// ND is the geomagnetic storm-day count.
	ND uint8
	// HasKp reports whether the has kp field is present.
	HasKp [8]bool
	// Kp10 contains the fixed-size array for this record.
	Kp10 [8]uint16
	// HasKpSum10 reports whether the has kp sum10 field is present.
	HasKpSum10 bool
	// KpSum10 is the 10-index geomagnetic sum.
	KpSum10 uint16
	// HasAp reports whether the has ap field is present.
	HasAp [8]bool
	// Ap8 contains the fixed-size array for this record.
	Ap8 [8]uint16
	// HasApAvg reports whether the has ap avg field is present.
	HasApAvg bool
	// ApAvg is the average planetary geomagnetic index.
	ApAvg uint16
	// HasCP10 reports whether the has cp10 field is present.
	HasCP10 bool
	// CP10 is the ten-centimetre solar flux.
	CP10 uint8
	// HasC9 reports whether the has c9 field is present.
	HasC9 bool
	// C9 is the nine-centimetre solar flux.
	C9 uint8
	// HasISN reports whether the has isn field is present.
	HasISN bool
	// ISN is the international sunspot number.
	ISN uint16
	// HasFluxQualifier reports whether the has flux qualifier field is present.
	HasFluxQualifier bool
	// FluxQualifier is the solar-flux qualifier.
	FluxQualifier uint8
	// HasF107Obs reports whether the has f107 obs field is present.
	HasF107Obs bool
	// F107Obs is the f107 obs value for SpaceWeatherDay.
	F107Obs float64
	// HasF107Adj reports whether the has f107 adj field is present.
	HasF107Adj bool
	// F107Adj is the f107 adj value for SpaceWeatherDay.
	F107Adj float64
	// HasF107ObsCenter81 reports whether the has f107 obs center81 field is present.
	HasF107ObsCenter81 bool
	// F107ObsCenter81 is the f107 obs center81 value for SpaceWeatherDay.
	F107ObsCenter81 float64
	// HasF107ObsLast81 reports whether the has f107 obs last81 field is present.
	HasF107ObsLast81 bool
	// F107ObsLast81 is the f107 obs last81 value for SpaceWeatherDay.
	F107ObsLast81 float64
	// HasF107AdjCenter81 reports whether the has f107 adj center81 field is present.
	HasF107AdjCenter81 bool
	// F107AdjCenter81 is the f107 adj center81 value for SpaceWeatherDay.
	F107AdjCenter81 float64
	// HasF107AdjLast81 reports whether the has f107 adj last81 field is present.
	HasF107AdjLast81 bool
	// F107AdjLast81 is the f107 adj last81 value for SpaceWeatherDay.
	F107AdjLast81 float64
}

// SpaceWeatherSample is one policy-selected space-weather sample.
type SpaceWeatherSample struct {
	// Weather is the space-weather table.
	Weather SpaceWeather
	// Class is the product or solution class.
	Class SpaceWeatherObservationClass
	// ApDefaulted is the ap defaulted value for SpaceWeatherSample.
	ApDefaulted bool
}

// SpaceWeatherPolicy controls whether interpolated and predicted records are
// acceptable to a sample query.
type SpaceWeatherPolicy struct {
	// AllowInterpolated is the allow interpolated value for SpaceWeatherPolicy.
	AllowInterpolated bool
	// AllowDailyPredicted is the allow daily predicted value for SpaceWeatherPolicy.
	AllowDailyPredicted bool
	// AllowMonthlyPredicted is the allow monthly predicted value for SpaceWeatherPolicy.
	AllowMonthlyPredicted bool
	// RequireGeomagnetic is the require geomagnetic value for SpaceWeatherPolicy.
	RequireGeomagnetic bool
}

// SpaceWeatherTableSummary contains parser record and diagnostic counts.
type SpaceWeatherTableSummary struct {
	// DayCount identifies or counts this record.
	DayCount int
	// MonthlyCount identifies or counts this record.
	MonthlyCount int
	// SkipCount identifies or counts this record.
	SkipCount int
	// WarningCount identifies or counts this record.
	WarningCount int
}

// SpaceWeatherTable owns a parsed space-weather table. Read methods may run
// concurrently; Close waits for active methods and is idempotent. The value
// must not be copied after first use.
type SpaceWeatherTable struct {
	_      noCopy
	native *native.SpaceWeatherTable
}

func spaceWeatherTable(value *native.SpaceWeatherTable, err error) (*SpaceWeatherTable, error) {
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errors.New("sidereon: native space-weather constructor returned no handle")
	}
	return &SpaceWeatherTable{native: value}, nil
}

// ParseSpaceWeatherTable parses the native table format from bytes.
func ParseSpaceWeatherTable(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherTable(native.ParseSpaceWeatherTable(data))
}

// ParseSpaceWeatherCSV parses the CSV table format from bytes.
func ParseSpaceWeatherCSV(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherTable(native.ParseSpaceWeatherCSV(data))
}

// ParseSpaceWeatherTXT parses the text table format from bytes.
func ParseSpaceWeatherTXT(data []byte) (*SpaceWeatherTable, error) {
	return spaceWeatherTable(native.ParseSpaceWeatherTXT(data))
}

// Close releases the table and is idempotent.
func (t *SpaceWeatherTable) Close() error {
	if t == nil || t.native == nil {
		return nil
	}
	return publicError(t.native.Close())
}

// Days returns copied daily records in native table order.
func (t *SpaceWeatherTable) Days() ([]SpaceWeatherDay, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.Days()
	return convertSpaceWeatherDays(value), publicError(err)
}

// Monthly returns copied monthly records in native table order.
func (t *SpaceWeatherTable) Monthly() ([]SpaceWeatherDay, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.Monthly()
	return convertSpaceWeatherDays(value), publicError(err)
}

func convertSpaceWeatherDays(values []native.SpaceWeatherDay) []SpaceWeatherDay {
	result := make([]SpaceWeatherDay, len(values))
	for i, value := range values {
		result[i] = SpaceWeatherDay{Year: value.Year, Month: value.Month, Day: value.Day, Class: SpaceWeatherObservationClass(value.Class), HasBSRN: value.HasBSRN, BSRN: value.BSRN, HasND: value.HasND, ND: value.ND, HasKp: value.HasKp, Kp10: value.Kp10, HasKpSum10: value.HasKpSum10, KpSum10: value.KpSum10, HasAp: value.HasAp, Ap8: value.Ap8, HasApAvg: value.HasApAvg, ApAvg: value.ApAvg, HasCP10: value.HasCP10, CP10: value.CP10, HasC9: value.HasC9, C9: value.C9, HasISN: value.HasISN, ISN: value.ISN, HasFluxQualifier: value.HasFluxQualifier, FluxQualifier: value.FluxQualifier, HasF107Obs: value.HasF107Obs, F107Obs: value.F107Obs, HasF107Adj: value.HasF107Adj, F107Adj: value.F107Adj, HasF107ObsCenter81: value.HasF107ObsCenter81, F107ObsCenter81: value.F107ObsCenter81, HasF107ObsLast81: value.HasF107ObsLast81, F107ObsLast81: value.F107ObsLast81, HasF107AdjCenter81: value.HasF107AdjCenter81, F107AdjCenter81: value.F107AdjCenter81, HasF107AdjLast81: value.HasF107AdjLast81, F107AdjLast81: value.F107AdjLast81}
	}
	return result
}

// Day returns a copied record and whether the requested date is present.
func (t *SpaceWeatherTable) Day(year int, month, day uint8) (SpaceWeatherDay, bool, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherDay{}, false, ErrClosed
	}
	value, present, err := t.native.Day(year, month, day)
	return convertSpaceWeatherDays([]native.SpaceWeatherDay{value})[0], present, publicError(err)
}

// Coverage returns table coverage with explicit presence for optional bounds.
func (t *SpaceWeatherTable) Coverage() (SpaceWeatherCoverage, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherCoverage{}, ErrClosed
	}
	value, err := t.native.Coverage()
	return SpaceWeatherCoverage{FirstJ2000S: value.FirstJ2000S, HasLastObservedJ2000S: value.HasLastObservedJ2000S, LastObservedJ2000S: value.LastObservedJ2000S, HasLastDailyPredictedJ2000S: value.HasLastDailyPredictedS, LastDailyPredictedJ2000S: value.LastDailyPredictedJ2000S, EndJ2000S: value.EndJ2000S}, publicError(err)
}

// Summary returns day, monthly, skipped-record, and warning counts.
func (t *SpaceWeatherTable) Summary() (SpaceWeatherTableSummary, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherTableSummary{}, ErrClosed
	}
	value, err := t.native.Summary()
	return SpaceWeatherTableSummary{DayCount: value.DayCount, MonthlyCount: value.MonthlyCount, SkipCount: value.SkipCount, WarningCount: value.WarningCount}, publicError(err)
}

// APArrayAt returns the seven native Ap-history values at an epoch.
func (t *SpaceWeatherTable) APArrayAt(epochJ2000S float64) ([7]float64, error) {
	if t == nil || t.native == nil {
		return [7]float64{}, ErrClosed
	}
	value, err := t.native.APArrayAt(epochJ2000S)
	return value, publicError(err)
}

// SampleAt samples the native table using default policy.
func (t *SpaceWeatherTable) SampleAt(epochJ2000S float64) (SpaceWeatherSample, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherSample{}, ErrClosed
	}
	value, err := t.native.SampleAt(epochJ2000S)
	return SpaceWeatherSample{Weather: SpaceWeather{F107: value.Weather.F107, F107A: value.Weather.F107A, Ap: value.Weather.Ap}, Class: SpaceWeatherObservationClass(value.Class), ApDefaulted: value.ApDefaulted}, publicError(err)
}

// SampleAtWithPolicy samples the table under explicit acceptance policy.
func (t *SpaceWeatherTable) SampleAtWithPolicy(epochJ2000S float64, policy SpaceWeatherPolicy) (SpaceWeatherSample, error) {
	if t == nil || t.native == nil {
		return SpaceWeatherSample{}, ErrClosed
	}
	value, err := t.native.SampleAtWithPolicy(epochJ2000S, native.SpaceWeatherPolicy{AllowInterpolated: policy.AllowInterpolated, AllowDailyPredicted: policy.AllowDailyPredicted, AllowMonthlyPredicted: policy.AllowMonthlyPredicted, RequireGeomagnetic: policy.RequireGeomagnetic})
	return SpaceWeatherSample{Weather: SpaceWeather{F107: value.Weather.F107, F107A: value.Weather.F107A, Ap: value.Weather.Ap}, Class: SpaceWeatherObservationClass(value.Class), ApDefaulted: value.ApDefaulted}, publicError(err)
}

// SpaceWeatherAt samples only the three weather indices.
func (t *SpaceWeatherTable) SpaceWeatherAt(epochJ2000S float64) (SpaceWeather, error) {
	if t == nil || t.native == nil {
		return SpaceWeather{}, ErrClosed
	}
	value, err := t.native.SpaceWeatherAt(epochJ2000S)
	return SpaceWeather{F107: value.F107, F107A: value.F107A, Ap: value.Ap}, publicError(err)
}

// ToCSV serializes the table to newly owned CSV bytes.
func (t *SpaceWeatherTable) ToCSV() ([]byte, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.ToCSV()
	return value, publicError(err)
}

// ToTXT serializes the table to newly owned text bytes.
func (t *SpaceWeatherTable) ToTXT() ([]byte, error) {
	if t == nil || t.native == nil {
		return nil, ErrClosed
	}
	value, err := t.native.ToTXT()
	return value, publicError(err)
}

// GeofenceErrorKind identifies the native diagnostic category accompanying a
// geofence operation.
type GeofenceErrorKind uint32

const (
	// GeofenceErrorNone identifies the geofence error none case.
	GeofenceErrorNone GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorNoneValue)
	// GeofenceErrorTooFewVertices identifies the geofence error too few vertices case.
	GeofenceErrorTooFewVertices GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorTooFewVerticesValue)
	// GeofenceErrorInvalidInput identifies the geofence error invalid input case.
	GeofenceErrorInvalidInput GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorInvalidInputValue)
	// GeofenceErrorGeodesic identifies the geofence error geodesic case.
	GeofenceErrorGeodesic GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorGeodesicValue)
	// GeofenceErrorDOP identifies the geofence error dop case.
	GeofenceErrorDOP GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorDOPValue)
	// GeofenceErrorMetrics identifies the geofence error metrics case.
	GeofenceErrorMetrics GeofenceErrorKind = GeofenceErrorKind(native.GeofenceErrorMetricsValue)
)

// GeofenceUncertaintyKind identifies an uncertainty representation.
type GeofenceUncertaintyKind uint32

const (
	// GeofenceENUCovarianceM2 identifies the geofence enu covariance m2 case.
	GeofenceENUCovarianceM2 GeofenceUncertaintyKind = GeofenceUncertaintyKind(native.GeofenceENUCovarianceM2Value)
	// GeofenceECEFCovarianceM2 identifies the geofence ecef covariance m2 case.
	GeofenceECEFCovarianceM2 GeofenceUncertaintyKind = GeofenceUncertaintyKind(native.GeofenceECEFCovarianceM2Value)
	// GeofenceCEPRadiusM identifies the geofence cep radius m case.
	GeofenceCEPRadiusM GeofenceUncertaintyKind = GeofenceUncertaintyKind(native.GeofenceCEPRadiusMValue)
)

// GeofenceProbabilityMethod selects native probability integration.
type GeofenceProbabilityMethod uint32

const (
	// GeofenceBoundaryNormal identifies the geofence boundary normal case.
	GeofenceBoundaryNormal GeofenceProbabilityMethod = GeofenceProbabilityMethod(native.GeofenceBoundaryNormalValue)
	// GeofencePlanarQuadrature identifies the geofence planar quadrature case.
	GeofencePlanarQuadrature GeofenceProbabilityMethod = GeofenceProbabilityMethod(native.GeofencePlanarQuadratureValue)
)

// GeofenceCrossingKind identifies an entered or left event.
type GeofenceCrossingKind uint32

const (
	// GeofenceEntered identifies the geofence entered case.
	GeofenceEntered GeofenceCrossingKind = GeofenceCrossingKind(native.GeofenceEnteredValue)
	// GeofenceLeft identifies the geofence left case.
	GeofenceLeft GeofenceCrossingKind = GeofenceCrossingKind(native.GeofenceLeftValue)
)

// GeofenceUncertainty describes covariance or CEP-radius uncertainty.
type GeofenceUncertainty struct {
	// Kind is the event or record kind.
	Kind GeofenceUncertaintyKind
	// CovarianceM2 is the covariance m2 in square metres.
	CovarianceM2 Matrix3
	// RadiusM is the radius m in metres.
	RadiusM float64
}

// GeofenceProbabilityOptions controls probability integration.
type GeofenceProbabilityOptions struct {
	// Method is the selected method.
	Method GeofenceProbabilityMethod
}

// GeofencePositionEstimate is one probabilistic trajectory sample.
type GeofencePositionEstimate struct {
	// Position is the position value in the containing frame.
	Position Geodetic
	// Uncertainty is the uncertainty value for GeofencePositionEstimate.
	Uncertainty GeofenceUncertainty
}

// GeofenceHysteresis contains entered and left confidence thresholds.
type GeofenceHysteresis struct {
	// EnterConfidence is the enter confidence value for GeofenceHysteresis.
	EnterConfidence float64
	// LeaveConfidence is the leave confidence value for GeofenceHysteresis.
	LeaveConfidence float64
}

// GeofenceCrossingEvent is one copied native crossing event.
type GeofenceCrossingEvent struct {
	// SampleIndex identifies or counts this record.
	SampleIndex int
	// Kind is the event or record kind.
	Kind GeofenceCrossingKind
	// InsideProbability is the inside probability value for GeofenceCrossingEvent.
	InsideProbability float64
}

// Geofence owns a native WGS84 geodesic polygon. Read calls may run
// concurrently; Close waits for active calls and is idempotent. The value
// must not be copied after first use.
type Geofence struct {
	_      noCopy
	native *native.Geofence
}

// NewGeofence creates a geodesic polygon and returns its native diagnostic
// category as well as any status error.
func NewGeofence(vertices []Geodetic) (*Geofence, GeofenceErrorKind, error) {
	if err := checkedEnvironmentAllocation(len(vertices), unsafe.Sizeof(native.Geodetic{})); err != nil {
		return nil, GeofenceErrorNone, err
	}
	nativeVertices := make([]native.Geodetic, len(vertices))
	for i, vertex := range vertices {
		nativeVertices[i] = native.Geodetic{LatitudeRad: vertex.LatitudeRad, LongitudeRad: vertex.LongitudeRad, HeightM: vertex.HeightM}
	}
	value, detail, err := native.GeofenceCreate(nativeVertices)
	if err != nil {
		return nil, GeofenceErrorKind(detail), publicError(err)
	}
	if value == nil {
		return nil, GeofenceErrorKind(detail), errors.New("sidereon: native geofence constructor returned no handle")
	}
	return &Geofence{native: value}, GeofenceErrorKind(detail), nil
}

// Close releases the native polygon and is idempotent.
func (f *Geofence) Close() error {
	if f == nil || f.native == nil {
		return nil
	}
	return publicError(f.native.Close())
}

// Contains reports whether a position is inside and returns the native
// diagnostic category.
func (f *Geofence) Contains(position Geodetic) (bool, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return false, GeofenceErrorNone, ErrClosed
	}
	value, detail, err := f.native.Contains(native.Geodetic{LatitudeRad: position.LatitudeRad, LongitudeRad: position.LongitudeRad, HeightM: position.HeightM})
	return value, GeofenceErrorKind(detail), publicError(err)
}

// DistanceToBoundary returns signed metres, positive inside, and the native
// diagnostic category.
func (f *Geofence) DistanceToBoundary(position Geodetic) (float64, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return 0, GeofenceErrorNone, ErrClosed
	}
	value, detail, err := f.native.DistanceToBoundary(native.Geodetic{LatitudeRad: position.LatitudeRad, LongitudeRad: position.LongitudeRad, HeightM: position.HeightM})
	return value, GeofenceErrorKind(detail), publicError(err)
}

func geofenceUncertainty(value GeofenceUncertainty) native.GeofenceUncertainty {
	var covariance [9]float64
	for row := range value.CovarianceM2 {
		for column := range value.CovarianceM2[row] {
			covariance[row*3+column] = value.CovarianceM2[row][column]
		}
	}
	return native.GeofenceUncertainty{Kind: uint32(value.Kind), Covariance: covariance, RadiusM: value.RadiusM}
}

// ContainmentProbability returns the native probability and diagnostic.
func (f *Geofence) ContainmentProbability(position Geodetic, uncertainty GeofenceUncertainty) (float64, GeofenceErrorKind, error) {
	return f.containmentProbability(position, uncertainty, nil)
}

// ContainmentProbabilityWithOptions returns probability under explicit
// integration options.
func (f *Geofence) ContainmentProbabilityWithOptions(position Geodetic, uncertainty GeofenceUncertainty, options GeofenceProbabilityOptions) (float64, GeofenceErrorKind, error) {
	return f.containmentProbability(position, uncertainty, &options)
}

func (f *Geofence) containmentProbability(position Geodetic, uncertainty GeofenceUncertainty, options *GeofenceProbabilityOptions) (float64, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return 0, GeofenceErrorNone, ErrClosed
	}
	if !validGeofenceUncertaintyKind(uncertainty.Kind) {
		return 0, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence uncertainty kind")
	}
	if options != nil && !validGeofenceProbabilityMethod(options.Method) {
		return 0, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence probability method")
	}
	nativePosition := native.Geodetic{LatitudeRad: position.LatitudeRad, LongitudeRad: position.LongitudeRad, HeightM: position.HeightM}
	nativeUncertainty := geofenceUncertainty(uncertainty)
	var value float64
	var detail uint32
	var err error
	if options == nil {
		value, detail, err = f.native.ContainmentProbability(nativePosition, nativeUncertainty)
	} else {
		value, detail, err = f.native.ContainmentProbabilityWithOptions(nativePosition, nativeUncertainty, native.GeofenceProbabilityOptions{Method: uint32(options.Method)})
	}
	return value, GeofenceErrorKind(detail), publicError(err)
}

func geofenceEvents(values []native.GeofenceCrossingEvent) []GeofenceCrossingEvent {
	if len(values) == 0 {
		return nil
	}
	result := make([]GeofenceCrossingEvent, len(values))
	for i, value := range values {
		result[i] = GeofenceCrossingEvent{SampleIndex: value.SampleIndex, Kind: GeofenceCrossingKind(value.Kind), InsideProbability: value.InsideProbability}
	}
	return result
}

// DefaultGeofenceProbabilityOptions returns native probability defaults.
func DefaultGeofenceProbabilityOptions() (GeofenceProbabilityOptions, error) {
	value, err := native.DefaultGeofenceProbabilityOptions()
	return GeofenceProbabilityOptions{Method: GeofenceProbabilityMethod(value.Method)}, publicError(err)
}

// DefaultGeofenceHysteresis returns native hysteresis defaults.
func DefaultGeofenceHysteresis() (GeofenceHysteresis, error) {
	value, err := native.DefaultGeofenceHysteresis()
	return GeofenceHysteresis{EnterConfidence: value.EnterConfidence, LeaveConfidence: value.LeaveConfidence}, publicError(err)
}

// CrossingProbability returns copied native crossing events and a diagnostic.
func (f *Geofence) CrossingProbability(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis) ([]GeofenceCrossingEvent, GeofenceErrorKind, error) {
	return f.crossingProbability(samples, hysteresis, nil)
}

// CrossingProbabilityWithOptions returns crossing events under explicit
// probability options.
func (f *Geofence) CrossingProbabilityWithOptions(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis, options GeofenceProbabilityOptions) ([]GeofenceCrossingEvent, GeofenceErrorKind, error) {
	return f.crossingProbability(samples, hysteresis, &options)
}

func (f *Geofence) crossingProbability(samples []GeofencePositionEstimate, hysteresis GeofenceHysteresis, options *GeofenceProbabilityOptions) ([]GeofenceCrossingEvent, GeofenceErrorKind, error) {
	if f == nil || f.native == nil {
		return nil, GeofenceErrorNone, ErrClosed
	}
	if err := checkedEnvironmentAllocation(len(samples), unsafe.Sizeof(native.GeofencePositionEstimate{})); err != nil {
		return nil, GeofenceErrorInvalidInput, err
	}
	if options != nil && !validGeofenceProbabilityMethod(options.Method) {
		return nil, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence probability method")
	}
	nativeSamples := make([]native.GeofencePositionEstimate, len(samples))
	for i, sample := range samples {
		if !validGeofenceUncertaintyKind(sample.Uncertainty.Kind) {
			return nil, GeofenceErrorInvalidInput, errors.New("sidereon: invalid geofence uncertainty kind")
		}
		nativeSamples[i] = native.GeofencePositionEstimate{Position: native.Geodetic{LatitudeRad: sample.Position.LatitudeRad, LongitudeRad: sample.Position.LongitudeRad, HeightM: sample.Position.HeightM}, Uncertainty: geofenceUncertainty(sample.Uncertainty)}
	}
	nativeHysteresis := native.GeofenceHysteresis{EnterConfidence: hysteresis.EnterConfidence, LeaveConfidence: hysteresis.LeaveConfidence}
	var value []native.GeofenceCrossingEvent
	var detail uint32
	var err error
	if options == nil {
		value, detail, err = f.native.CrossingProbability(nativeSamples, nativeHysteresis)
	} else {
		value, detail, err = f.native.CrossingProbabilityWithOptions(nativeSamples, nativeHysteresis, native.GeofenceProbabilityOptions{Method: uint32(options.Method)})
	}
	return geofenceEvents(value), GeofenceErrorKind(detail), publicError(err)
}

// ObservabilityTier identifies a geometry-quality tier.
type ObservabilityTier uint32

const (
	// ObservabilityRankDeficient identifies the observability rank deficient case.
	ObservabilityRankDeficient ObservabilityTier = ObservabilityTier(native.ObservabilityRankDeficientValue)
	// ObservabilityZeroRedundancy identifies the observability zero redundancy case.
	ObservabilityZeroRedundancy ObservabilityTier = ObservabilityTier(native.ObservabilityZeroRedundancyValue)
	// ObservabilityWeak identifies the observability weak case.
	ObservabilityWeak ObservabilityTier = ObservabilityTier(native.ObservabilityWeakValue)
	// ObservabilityNominal identifies the observability nominal case.
	ObservabilityNominal ObservabilityTier = ObservabilityTier(native.ObservabilityNominalValue)
)

// ObservabilityTierLabel returns the native stable lowercase tier label.
func ObservabilityTierLabel(tier ObservabilityTier) (string, error) {
	if !validObservabilityTier(tier) {
		return "", errors.New("sidereon: invalid observability tier")
	}
	value, err := native.ObservabilityTierLabel(uint32(tier))
	return string(value), publicError(err)
}
