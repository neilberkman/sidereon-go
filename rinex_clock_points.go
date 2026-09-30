package sidereon

import (
	"errors"
	"sidereon.dev/go/v3/internal/native"
)

// RINEXClockEpochFromCivil converts civil fields in a continuous time scale to
// the native clock instant representation. UTC leap seconds use the scale's
// civil rules; available is false when that calendar date has no instant.
func RINEXClockEpochFromCivil(scale TimeScale, value CivilDateTime) (ClockEpoch, bool, error) {
	epoch, available, err := native.CivilToClockEpoch(uint32(scale), native.CivilDateTime{Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second})
	if err != nil {
		return ClockEpoch{}, false, publicError(err)
	}
	return ClockEpoch{Scale: TimeScale(epoch.Scale), Representation: RINEXClockInstantRepresentation(epoch.Representation), JulianWhole: epoch.JulianWhole, JulianFraction: epoch.JulianFraction, NanosHigh: epoch.NanosHigh, NanosLow: epoch.NanosLow}, available, nil
}

// RINEXClockPoint is a scale-tagged epoch and the values retained by one AS sample.
type RINEXClockPoint struct {
	// Epoch is the exact instant and time scale of this sample.
	Epoch ClockEpoch
	// BiasS is the satellite clock bias in seconds.
	BiasS float64
	// AdditionalValues are the declared clock terms after bias, in RINEX order.
	AdditionalValues []float64
}

// RINEXClockSatellitePoint associates one sample with its satellite token.
type RINEXClockSatellitePoint struct {
	// Satellite is the satellite identifier as written, such as G01.
	Satellite string
	// Point contains the sample epoch and all declared values.
	Point RINEXClockPoint
}

// NewRINEXClockFromPoints builds an owned clock product from scale-tagged samples.
// Samples for each satellite must appear in strictly increasing epoch order.
func NewRINEXClockFromPoints(scale TimeScale, points []RINEXClockSatellitePoint) (*RINEXClock, error) {
	values := make([]native.NativeClockSatellitePoint, len(points))
	for i, p := range points {
		values[i] = native.NativeClockSatellitePoint{Satellite: p.Satellite, Point: native.NativeClockPoint{Epoch: nativeClockEpoch(p.Point.Epoch), BiasS: p.Point.BiasS, AdditionalValues: append([]float64(nil), p.Point.AdditionalValues...)}}
	}
	handle, err := native.NewRinexClockFromPoints(uint32(scale), values)
	if err != nil {
		return nil, publicClockWriteError(err)
	}
	if handle == nil {
		return nil, errors.New("sidereon: native RINEX clock point constructor returned no handle")
	}
	return &RINEXClock{handle: handle}, nil
}
