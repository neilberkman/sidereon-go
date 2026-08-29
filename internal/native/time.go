package native

import (
	"fmt"
	"time"
)

const (
	unixMicrosecondsPerSecond = int64(1_000_000)
	maxInt64                  = int64(1<<63 - 1)
	minInt64                  = -maxInt64 - 1
	minInt32                  = -1 << 31
	maxInt32                  = 1<<31 - 1
	maxUint8                  = 1<<8 - 1
)

// unixMicroseconds converts a UTC instant to the signed Unix-microsecond
// representation used by the C ABI. Unix reports whole seconds using floor
// semantics, so adding the nonnegative fractional microseconds preserves the
// same behavior for instants before the Unix epoch as time.Time.UnixMicro.
func unixMicroseconds(value time.Time) (int64, error) {
	utc := value.UTC()
	seconds := utc.Unix()
	fractionalMicroseconds := int64(utc.Nanosecond() / 1_000)
	maxSeconds := maxInt64 / unixMicrosecondsPerSecond
	minSeconds := minInt64 / unixMicrosecondsPerSecond
	if seconds > maxSeconds || seconds < minSeconds-1 {
		return 0, fmt.Errorf("sidereon: UTC instant %v does not fit in signed Unix microseconds", value)
	}
	if seconds == maxSeconds && fractionalMicroseconds > maxInt64%unixMicrosecondsPerSecond {
		return 0, fmt.Errorf("sidereon: UTC instant %v does not fit in signed Unix microseconds", value)
	}
	// The mathematical minimum is one fractional second below the truncating
	// integer quotient of MinInt64. Handle that one boundary without forming an
	// overflowing seconds*microseconds product.
	minimumBoundaryFraction := unixMicrosecondsPerSecond + minInt64%unixMicrosecondsPerSecond
	if seconds == minSeconds-1 {
		if fractionalMicroseconds < minimumBoundaryFraction {
			return 0, fmt.Errorf("sidereon: UTC instant %v does not fit in signed Unix microseconds", value)
		}
		return minInt64 + fractionalMicroseconds - minimumBoundaryFraction, nil
	}
	wholeMicroseconds := seconds * unixMicrosecondsPerSecond
	if wholeMicroseconds > maxInt64-fractionalMicroseconds {
		return 0, fmt.Errorf("sidereon: UTC instant %v does not fit in signed Unix microseconds", value)
	}
	return wholeMicroseconds + fractionalMicroseconds, nil
}

func unixMicrosecondsSlice(values []time.Time) ([]int64, error) {
	out := make([]int64, len(values))
	for i, value := range values {
		converted, err := unixMicroseconds(value)
		if err != nil {
			return nil, err
		}
		out[i] = converted
	}
	return out, nil
}

func checkedInt32(value int, label string) (int32, error) {
	if value < minInt32 || value > maxInt32 {
		return 0, fmt.Errorf("sidereon: %s %d does not fit in C int32", label, value)
	}
	return int32(value), nil
}

func checkedUint8(value int, label string) (uint8, error) {
	if value < 0 || value > maxUint8 {
		return 0, fmt.Errorf("sidereon: %s %d does not fit in C uint8", label, value)
	}
	return uint8(value), nil
}

type civilTimeParts struct {
	year, month, day, hour, minute int32
	second                         float64
}

func checkedCivilTime(value time.Time) (civilTimeParts, error) {
	utc := value.UTC()
	year, month, day := utc.Date()
	checkedYear, err := checkedInt32(year, "UTC year")
	if err != nil {
		return civilTimeParts{}, err
	}
	return civilTimeParts{
		year:   checkedYear,
		month:  int32(month),
		day:    int32(day),
		hour:   int32(utc.Hour()),
		minute: int32(utc.Minute()),
		second: float64(utc.Second()) + float64(utc.Nanosecond()/1_000)/1e6,
	}, nil
}
