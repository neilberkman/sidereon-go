package native

import (
	"errors"
	"testing"
	"time"
)

func TestUnixMicrosecondsPreservesFloorSemantics(t *testing.T) {
	tests := []struct {
		name string
		when time.Time
		want int64
	}{
		{name: "epoch", when: time.Unix(0, 0), want: 0},
		{name: "positive submicrosecond", when: time.Unix(0, 999), want: 0},
		{name: "negative one microsecond", when: time.Unix(-1, 999999000), want: -1},
		{name: "negative submicrosecond", when: time.Unix(-1, 999999500), want: -1},
		{name: "negative nanosecond argument", when: time.Unix(0, -500), want: -1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := unixMicroseconds(test.when)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("unixMicroseconds(%v) = %d, want %d", test.when, got, test.want)
			}
		})
	}
}

func TestUnixMicrosecondsBoundaries(t *testing.T) {
	maxTime := time.Unix(9223372036854, 775807000)
	if got, err := unixMicroseconds(maxTime); err != nil || got != maxInt64 {
		t.Fatalf("maximum Unix microseconds = %d, %v; want %d, nil", got, err, maxInt64)
	}
	if _, err := unixMicroseconds(time.Unix(9223372036854, 775808000)); err == nil {
		t.Fatal("positive Unix-microsecond overflow was accepted")
	}

	minTime := time.Unix(-9223372036855, 224192000)
	if got, err := unixMicroseconds(minTime); err != nil || got != minInt64 {
		t.Fatalf("minimum Unix microseconds = %d, %v; want %d, nil", got, err, minInt64)
	}
	if _, err := unixMicroseconds(time.Unix(-9223372036855, 224191000)); err == nil {
		t.Fatal("negative Unix-microsecond overflow was accepted")
	}
}

func TestCheckedCivilTimeRejectsYearNarrowing(t *testing.T) {
	if _, err := checkedCivilTime(time.Date(maxInt32+1, time.January, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("civil year outside int32 was accepted")
	}
	if _, err := checkedInt32(minInt32, "year"); err != nil {
		t.Fatal(err)
	}
	if _, err := checkedInt32(maxInt32, "year"); err != nil {
		t.Fatal(err)
	}
}

func TestCivilToGPSRejectsUint8Narrowing(t *testing.T) {
	base := CivilDateTime{Year: 2020, Month: 6, Day: 25, Hour: 12, Minute: 34}
	tests := []struct {
		name  string
		apply func(*CivilDateTime, int)
	}{
		{name: "negative month", apply: func(value *CivilDateTime, hostile int) { value.Month = hostile }},
		{name: "month above uint8", apply: func(value *CivilDateTime, hostile int) { value.Month = hostile }},
		{name: "negative day", apply: func(value *CivilDateTime, hostile int) { value.Day = hostile }},
		{name: "day above uint8", apply: func(value *CivilDateTime, hostile int) { value.Day = hostile }},
		{name: "negative hour", apply: func(value *CivilDateTime, hostile int) { value.Hour = hostile }},
		{name: "hour above uint8", apply: func(value *CivilDateTime, hostile int) { value.Hour = hostile }},
		{name: "negative minute", apply: func(value *CivilDateTime, hostile int) { value.Minute = hostile }},
		{name: "minute above uint8", apply: func(value *CivilDateTime, hostile int) { value.Minute = hostile }},
	}
	for index, test := range tests {
		hostile := -1
		if index%2 == 1 {
			hostile = maxUint8 + 1
		}
		t.Run(test.name, func(t *testing.T) {
			value := base
			test.apply(&value, hostile)
			_, _, err := CivilToGPS(value)
			if err == nil {
				t.Fatal("hostile civil field was accepted")
			}
			var statusErr *StatusError
			if errors.As(err, &statusErr) {
				t.Fatalf("hostile civil field reached C: %v", err)
			}
		})
	}
}

func TestCheckedUint8Boundaries(t *testing.T) {
	for _, value := range []int{0, maxUint8} {
		if got, err := checkedUint8(value, "civil field"); err != nil || int(got) != value {
			t.Fatalf("checkedUint8(%d) = %d, %v", value, got, err)
		}
	}
}

func TestCivilToGPSValidBoundaryConversion(t *testing.T) {
	seconds, available, err := CivilToGPS(CivilDateTime{Year: 1980, Month: 1, Day: 6})
	if err != nil || !available || seconds != 0 {
		t.Fatalf("GPS epoch conversion = %.17g, %t, %v; want 0, true, nil", seconds, available, err)
	}
}
