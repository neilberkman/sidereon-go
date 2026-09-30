package sidereon

import (
	"math"
	"strings"
	"testing"
)

const blqRoundTripFixture = "$$ file header\nTEST\n$$ station note\n" +
	"$$ COLUMN ORDER: M2 S2 N2 K2 K1 O1 P1 Q1 MF MM SSA\n" +
	"1 2 3 4 5 6 7 8 9 10 11\n" +
	"12 13 14 15 16 17 18 19 20 21 22\n" +
	"23 24 25 26 27 28 29 30 31 32 33\n" +
	"34 35 36 37 38 39 40 41 42 43 44\n" +
	"45 46 47 48 49 50 51 52 53 54 55\n" +
	"56 57 58 59 60 61 62 63 64 65 66\n" +
	"$$ trailing note\n"

func TestBLQRetainsCommentsAndRoundTrips(t *testing.T) {
	blocks, outcome, err := ParseBLQ([]byte(blqRoundTripFixture))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, blocks)
	if !outcome.IsOK {
		t.Fatalf("parse outcome = %+v", outcome)
	}
	if n, err := blocks.Count(); err != nil || n != 1 {
		t.Fatalf("Count() = %d, %v", n, err)
	}
	station, err := blocks.Station(0)
	if err != nil || station != "TEST" {
		t.Fatalf("Station() = %q, %v", station, err)
	}
	count, err := blocks.CommentCount(0)
	if err != nil || count != 4 {
		t.Fatalf("CommentCount() = %d, %v", count, err)
	}
	want := []BLQComment{{Placement: BLQBeforeStation, Text: "$$ file header"}, {Placement: BLQBeforeRow, Row: 0, Text: "$$ station note"}, {Placement: BLQBeforeRow, Row: 0, Text: "$$ COLUMN ORDER: M2 S2 N2 K2 K1 O1 P1 Q1 MF MM SSA"}, {Placement: BLQAfterRows, Text: "$$ trailing note"}}
	comments := make([]BLQComment, len(want))
	for i, expected := range want {
		got, e := blocks.Comment(0, i)
		if e != nil || got != expected {
			t.Fatalf("Comment(%d) = %+v, %v; want %+v", i, got, e, expected)
		}
		comments[i] = got
	}
	coeff, err := blocks.Coefficients(0)
	if err != nil || coeff.AmplitudeM[0][0] != 1 || coeff.PhaseDeg[2][10] != 66 {
		t.Fatalf("Coefficients() = %+v, %v", coeff, err)
	}
	encoded, writeOutcome, err := blocks.Encode()
	if err != nil || !writeOutcome.IsOK {
		t.Fatalf("Encode() = %q, %+v, %v", encoded, writeOutcome, err)
	}
	if !strings.Contains(string(encoded), "$$ file header\n") || !strings.Contains(string(encoded), "$$ trailing note\n") || !strings.Contains(string(encoded), "$$ COLUMN ORDER: M2 S2 N2 K2 K1 O1 P1 Q1 MF MM SSA\n") {
		t.Fatalf("encoded text lost retained comments/header: %s", encoded)
	}
	if err := blocks.Close(); err != nil {
		t.Fatal(err)
	}
	if comments[0].Text != "$$ file header" || comments[2].Placement != BLQBeforeRow || comments[2].Row != 0 {
		t.Fatalf("detached comments changed after Close: %+v", comments)
	}
	reparsed, parseOutcome, err := ParseBLQ(encoded)
	if err != nil || !parseOutcome.IsOK {
		t.Fatalf("ParseBLQ(encoded) = %+v, %v", parseOutcome, err)
	}
	closeAfterTest(t, reparsed)
	if got, e := reparsed.Station(0); e != nil || got != station {
		t.Fatalf("round-trip station = %q, %v", got, e)
	}
	got, err := reparsed.Coefficients(0)
	if err != nil || got != coeff {
		t.Fatalf("round-trip coefficients differ: %+v, %v", got, err)
	}
	built, err := NewBLQBlocks()
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, built)
	if err := built.Push(station, coeff); err != nil {
		t.Fatal(err)
	}
	if err := built.PushComment(0, BLQComment{Placement: BLQBeforeStation, Text: "$$ constructed"}); err != nil {
		t.Fatal(err)
	}
	builtText, builtOutcome, err := built.Encode()
	if err != nil || !builtOutcome.IsOK {
		t.Fatalf("constructed Encode() = %q, %+v, %v", builtText, builtOutcome, err)
	}
	builtAgain, builtParseOutcome, err := ParseBLQ(builtText)
	if err != nil || !builtParseOutcome.IsOK {
		t.Fatalf("ParseBLQ(constructed text) = %+v, %v", builtParseOutcome, err)
	}
	closeAfterTest(t, builtAgain)
	retained, err := builtAgain.Comment(0, 0)
	if err != nil || retained.Text != "$$ constructed" || retained.Placement != BLQBeforeStation {
		t.Fatalf("constructed comment = %+v, %v", retained, err)
	}
}

func TestBLQTypedParseAndWriterRefusals(t *testing.T) {
	blocks, outcome, err := ParseBLQ([]byte("TEST\n1 2\n"))
	if err != nil || blocks != nil || outcome.IsOK || outcome.Error.Kind != BLQErrorParse || outcome.Error.ParseKind != BLQParseWrongColumnCount || !outcome.Error.HasLine || outcome.Error.Line != 2 || !outcome.Error.HasExpected || outcome.Error.Expected != 11 || !outcome.Error.HasFound || outcome.Error.Found != 2 || outcome.Error.Message == "" {
		t.Fatalf("typed parse refusal = blocks %v, outcome %+v, err %v", blocks, outcome, err)
	}

	blocks, err = NewBLQBlocks()
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, blocks)
	coeff := OceanLoadingBLQ{}
	coeff.AmplitudeM[1][3] = math.Inf(1)
	if err := blocks.Push("TEST", coeff); err != nil {
		t.Fatal(err)
	}
	text, writeOutcome, err := blocks.Encode()
	if err := blocks.Close(); err != nil {
		t.Fatal(err)
	}
	if err != nil || text != nil || writeOutcome.IsOK || writeOutcome.Error.Kind != BLQErrorWrite || writeOutcome.Error.WriteKind != BLQWriteNonFiniteCoefficient || !writeOutcome.Error.HasBlock || writeOutcome.Error.Block != 0 || !writeOutcome.Error.HasCoefficient || writeOutcome.Error.Row != 1 || writeOutcome.Error.Constituent != BLQTideK2 || writeOutcome.Error.Message == "" {
		t.Fatalf("typed write refusal after source close = %q, %+v, %v", text, writeOutcome, err)
	}
}
