package sidereon

import "sidereon.dev/go/v3/internal/native"

// ExactEpochComponents preserves the integer and decimal-residue components of an epoch.
type ExactEpochComponents struct {
	Seconds       int64
	Attoseconds   uint64
	ResidueDigits int64
	ResiduePlaces uint16
}

// ExactOrdering is the result of an exact epoch comparison.
type ExactOrdering int32

const (
	// ExactOrderingLess means the receiver is earlier than the argument.
	ExactOrderingLess ExactOrdering = ExactOrdering(native.ExactOrderingLess)
	// ExactOrderingEqual means both values represent the same instant.
	ExactOrderingEqual ExactOrdering = ExactOrdering(native.ExactOrderingEqual)
	// ExactOrderingGreater means the receiver is later than the argument.
	ExactOrderingGreater ExactOrdering = ExactOrdering(native.ExactOrderingGreater)
)

// ExactEpoch stores a J2000-relative instant without reducing it to float64.
// Close releases its native handle; a nil receiver is safe to close.
type ExactEpoch struct{ handle *native.ExactEpoch }

// ExactEpochQuery stores an exact instant for binary-computed offsets.
// Close releases its native handle; a nil receiver is safe to close.
type ExactEpochQuery struct{ handle *native.ExactEpochQuery }

// NewExactEpoch constructs an epoch from whole seconds and attoseconds.
func NewExactEpoch(seconds int64, attoseconds uint64) (*ExactEpoch, error) {
	value, err := native.NewExactEpoch(seconds, attoseconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpoch{handle: value}, nil
}

// ExactEpochJ2000 returns a new owned handle to the J2000 epoch.
func ExactEpochJ2000() (*ExactEpoch, error) {
	value, err := native.ExactEpochJ2000()
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpoch{handle: value}, nil
}

// ExactAttosecondsPerSecond returns the integer attosecond scale used by ExactEpoch.
func ExactAttosecondsPerSecond() (uint64, error) {
	value, err := native.ExactAttosecondsPerSecond()
	return value, publicError(err)
}

// ExactEpochFromJ2000Seconds parses the shortest-decimal representation of seconds.
func ExactEpochFromJ2000Seconds(seconds float64) (*ExactEpoch, error) {
	value, err := native.ExactEpochFromJ2000Seconds(seconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpoch{handle: value}, nil
}

// ExactEpochFromCivil constructs an epoch from civil fields and a decimal second label.
func ExactEpochFromCivil(value CivilDateTime) (*ExactEpoch, error) {
	nativeValue, err := native.ExactEpochFromCivil(native.CivilDateTime{
		Year: value.Year, Month: value.Month, Day: value.Day,
		Hour: value.Hour, Minute: value.Minute, Second: value.Second,
	})
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpoch{handle: nativeValue}, nil
}

// ExactEpochQueryFromBinaryJ2000Seconds preserves the exact binary float value.
func ExactEpochQueryFromBinaryJ2000Seconds(seconds float64) (*ExactEpochQuery, error) {
	value, err := native.ExactEpochQueryFromBinaryJ2000Seconds(seconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpochQuery{handle: value}, nil
}

// Query returns a query at the exact civil-label epoch without a binary offset.
func (epoch *ExactEpoch) Query() (*ExactEpochQuery, error) {
	if epoch == nil || epoch.handle == nil {
		return nil, ErrClosed
	}
	value, err := epoch.handle.Query()
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpochQuery{handle: value}, nil
}

// Epoch returns a newly owned exact epoch with the query's exact value.
func (query *ExactEpochQuery) Epoch() (*ExactEpoch, error) {
	if query == nil || query.handle == nil {
		return nil, ErrClosed
	}
	value, err := query.handle.Epoch()
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpoch{handle: value}, nil
}

// Compare orders epochs without converting them to rounded seconds.
func (epoch *ExactEpoch) Compare(other *ExactEpoch) (ExactOrdering, error) {
	if epoch == nil || epoch.handle == nil || other == nil || other.handle == nil {
		return ExactOrderingEqual, ErrClosed
	}
	value, err := epoch.handle.Compare(other.handle)
	return ExactOrdering(value), publicError(err)
}

// Equal reports whether two epochs represent the same instant.
func (epoch *ExactEpoch) Equal(other *ExactEpoch) (bool, error) {
	if epoch == nil || epoch.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	value, err := epoch.handle.Equal(other.handle)
	return value, publicError(err)
}

// Equal reports whether two queries represent the same instant, independent
// of their epoch origin or binary-offset representation.
func (query *ExactEpochQuery) Equal(other *ExactEpochQuery) (bool, error) {
	if query == nil || query.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	value, err := query.handle.Equal(other.handle)
	return value, publicError(err)
}

// CheckedAddSeconds adds a shortest-decimal seconds value to the epoch.
func (epoch *ExactEpoch) CheckedAddSeconds(seconds float64) (*ExactEpoch, error) {
	if epoch == nil || epoch.handle == nil {
		return nil, ErrClosed
	}
	value, err := epoch.handle.CheckedAddSeconds(seconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpoch{handle: value}, nil
}

// CheckedSubSeconds subtracts a shortest-decimal seconds value from the epoch.
func (epoch *ExactEpoch) CheckedSubSeconds(seconds float64) (*ExactEpoch, error) {
	if epoch == nil || epoch.handle == nil {
		return nil, ErrClosed
	}
	value, err := epoch.handle.CheckedSubSeconds(seconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpoch{handle: value}, nil
}

// CheckedAddBinarySeconds adds the exact binary value of seconds to the query.
func (query *ExactEpochQuery) CheckedAddBinarySeconds(seconds float64) (*ExactEpochQuery, error) {
	if query == nil || query.handle == nil {
		return nil, ErrClosed
	}
	value, err := query.handle.CheckedAddBinarySeconds(seconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpochQuery{handle: value}, nil
}

// CheckedSubBinarySeconds subtracts the exact binary value of seconds from the query.
func (query *ExactEpochQuery) CheckedSubBinarySeconds(seconds float64) (*ExactEpochQuery, error) {
	if query == nil || query.handle == nil {
		return nil, ErrClosed
	}
	value, err := query.handle.CheckedSubBinarySeconds(seconds)
	if err != nil {
		return nil, publicError(err)
	}
	return &ExactEpochQuery{handle: value}, nil
}

// Components returns the exact integer and sub-attosecond decimal components.
func (epoch *ExactEpoch) Components() (ExactEpochComponents, error) {
	if epoch == nil || epoch.handle == nil {
		return ExactEpochComponents{}, ErrClosed
	}
	value, err := epoch.handle.Components()
	return ExactEpochComponents{
		Seconds: value.Seconds, Attoseconds: value.Attoseconds,
		ResidueDigits: value.ResidueDigits, ResiduePlaces: value.ResiduePlaces,
	}, publicError(err)
}

// J2000Seconds returns the exact epoch rounded once to float64.
func (epoch *ExactEpoch) J2000Seconds() (float64, error) {
	if epoch == nil || epoch.handle == nil {
		return 0, ErrClosed
	}
	value, err := epoch.handle.J2000Seconds()
	return value, publicError(err)
}

// SplitJulianDate returns a Julian date rounded to a whole/fraction pair.
func (epoch *ExactEpoch) SplitJulianDate() (JulianDate, error) {
	if epoch == nil || epoch.handle == nil {
		return JulianDate{}, ErrClosed
	}
	value, err := epoch.handle.SplitJulianDate()
	return JulianDate{Whole: value.Whole, Fraction: value.Fraction}, publicError(err)
}

// SecondsSince returns epoch minus earlier, rounded once to float64.
func (epoch *ExactEpoch) SecondsSince(earlier *ExactEpoch) (float64, error) {
	if epoch == nil || epoch.handle == nil || earlier == nil || earlier.handle == nil {
		return 0, ErrClosed
	}
	value, err := epoch.handle.SecondsSince(earlier.handle)
	return value, publicError(err)
}

// SecondsSince returns query minus earlier, rounded once to float64.
func (query *ExactEpochQuery) SecondsSince(earlier *ExactEpoch) (float64, error) {
	if query == nil || query.handle == nil || earlier == nil || earlier.handle == nil {
		return 0, ErrClosed
	}
	value, err := query.handle.SecondsSince(earlier.handle)
	return value, publicError(err)
}

// SecondsSinceQuery returns the exact query difference rounded once to float64.
func (query *ExactEpochQuery) SecondsSinceQuery(earlier *ExactEpochQuery) (float64, error) {
	if query == nil || query.handle == nil || earlier == nil || earlier.handle == nil {
		return 0, ErrClosed
	}
	value, err := query.handle.SecondsSinceQuery(earlier.handle)
	return value, publicError(err)
}

// J2000Seconds returns the query rounded once to float64.
func (query *ExactEpochQuery) J2000Seconds() (float64, error) {
	if query == nil || query.handle == nil {
		return 0, ErrClosed
	}
	value, err := query.handle.J2000Seconds()
	return value, publicError(err)
}

// Close releases the epoch's native handle.
func (epoch *ExactEpoch) Close() error {
	if epoch == nil || epoch.handle == nil {
		return nil
	}
	return publicError(epoch.handle.Close())
}

// Close releases the query's native handle.
func (query *ExactEpochQuery) Close() error {
	if query == nil || query.handle == nil {
		return nil
	}
	return publicError(query.handle.Close())
}
