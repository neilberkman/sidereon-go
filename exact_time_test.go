package sidereon

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

func TestExactEpochPreservesIntegerComponents(tester *testing.T) {
	const seconds int64 = 1<<53 + 1
	const attoseconds uint64 = 999999999999999999
	epoch, err := NewExactEpoch(seconds, attoseconds)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, epoch)
	components, err := epoch.Components()
	if err != nil {
		tester.Fatal(err)
	}
	want := ExactEpochComponents{Seconds: seconds, Attoseconds: attoseconds}
	if components != want {
		tester.Fatalf("components = %+v, want %+v", components, want)
	}
	difference, err := epoch.SecondsSince(epoch)
	if err != nil || difference != 0 {
		tester.Fatalf("self difference = %g, %v", difference, err)
	}
	if invalid, err := NewExactEpoch(0, 1000000000000000000); err == nil {
		if invalid != nil {
			closeAfterTest(tester, invalid)
		}
		tester.Fatal("accepted an attosecond count of one second")
	}
}

func TestExactEpochQueryRetainsBinaryOffsetDifference(tester *testing.T) {
	decimal, err := ExactEpochFromJ2000Seconds(0.1)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, decimal)
	binary, err := ExactEpochQueryFromBinaryJ2000Seconds(0.1)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, binary)
	rational := new(big.Rat).SetFloat64(0.1)
	rational.Sub(rational, big.NewRat(1, 10))
	want, _ := rational.Float64()
	difference, err := binary.SecondsSince(decimal)
	if err != nil || math.Float64bits(difference) != math.Float64bits(want) {
		tester.Fatalf("binary minus decimal = %.17g, want %.17g, error %v", difference, want, err)
	}
	decimalQuery, err := decimal.Query()
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, decimalQuery)
	difference, err = binary.SecondsSinceQuery(decimalQuery)
	if err != nil || math.Float64bits(difference) != math.Float64bits(want) {
		tester.Fatalf("query difference = %.17g, want %.17g, error %v", difference, want, err)
	}
	shifted, err := decimalQuery.CheckedAddBinarySeconds(0.1)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, shifted)
	restored, err := shifted.CheckedSubBinarySeconds(0.1)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, restored)
	difference, err = restored.SecondsSinceQuery(decimalQuery)
	if err != nil || difference != 0 {
		tester.Fatalf("exact offset cancellation = %g, %v", difference, err)
	}
}

func TestExactEpochQueryOwnsItsValueAfterEpochClose(tester *testing.T) {
	epoch, err := NewExactEpoch(1, 1)
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, epoch)
	query, err := epoch.Query()
	if err != nil {
		tester.Fatal(err)
	}
	closeAfterTest(tester, query)
	if err := epoch.Close(); err != nil {
		tester.Fatal(err)
	}
	if _, err := epoch.Components(); !errors.Is(err, ErrClosed) {
		tester.Fatalf("closed epoch error = %v, want ErrClosed", err)
	}
	if seconds, err := query.J2000Seconds(); err != nil || seconds != 1 {
		tester.Fatalf("independently owned query = %g, %v", seconds, err)
	}
	difference, err := query.SecondsSinceQuery(query)
	if err != nil || difference != 0 {
		tester.Fatalf("query self difference = %g, %v", difference, err)
	}
	if err := query.Close(); err != nil {
		tester.Fatal(err)
	}
	if _, err := query.J2000Seconds(); !errors.Is(err, ErrClosed) {
		tester.Fatalf("closed query error = %v, want ErrClosed", err)
	}
	var emptyEpoch *ExactEpoch
	var emptyQuery *ExactEpochQuery
	if err := emptyEpoch.Close(); err != nil {
		tester.Fatal(err)
	}
	if err := emptyQuery.Close(); err != nil {
		tester.Fatal(err)
	}
}
