package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SecondOfDay returns the fractional seconds since midnight. Validation and
// arithmetic are performed by the C library.
func SecondOfDay(hour, minute int, second float64) (float64, error) {
	value, err := native.SecondOfDay(hour, minute, second)
	return value, publicError(err)
}

// DayOfYear returns the fractional, one-based day of a civil instant.
func DayOfYear(year, month, day, hour, minute int, second float64) (float64, error) {
	value, err := native.DayOfYear(year, month, day, hour, minute, second)
	return value, publicError(err)
}

// DataDayOfYear returns the one-based integer day of a product date.
func DataDayOfYear(year int, month, day uint8) (uint16, error) {
	value, err := native.DataDayOfYear(year, month, day)
	return value, publicError(err)
}
