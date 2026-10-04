package sidereon

import (
	"bytes"
	"reflect"
	"testing"
)

func checkClockHeaderPayload(t *testing.T, got, want RINEXClockHeaderRecord) {
	t.Helper()
	if got.FieldKind != want.FieldKind || got.Reading != want.Reading || !reflect.DeepEqual(got.TextParts, want.TextParts) {
		t.Fatalf("header kind/reading/text = (%d,%d,%q), want (%d,%d,%q)", got.FieldKind, got.Reading, got.TextParts, want.FieldKind, want.Reading, want.TextParts)
	}
	if got.HasVersion != want.HasVersion || got.HasSystemCode != want.HasSystemCode || got.HasCount != want.HasCount || got.HasInteger != want.HasInteger || got.HasStart != want.HasStart || got.HasStop != want.HasStop || got.HasConstraintS != want.HasConstraintS || got.HasXYZMM != want.HasXYZMM {
		t.Fatalf("header presence flags = %+v, want %+v", got, want)
	}
	if want.HasVersion && got.Version != want.Version {
		t.Fatalf("version %.17g, want %.17g", got.Version, want.Version)
	}
	if want.HasSystemCode && got.SystemCode != want.SystemCode {
		t.Fatalf("system %d, want %d", got.SystemCode, want.SystemCode)
	}
	if want.HasCount && got.Count != want.Count {
		t.Fatalf("count %d, want %d", got.Count, want.Count)
	}
	if want.HasInteger && got.Integer != want.Integer {
		t.Fatalf("integer %d, want %d", got.Integer, want.Integer)
	}
	if want.HasStart && got.Start != want.Start {
		t.Fatalf("start %+v, want %+v", got.Start, want.Start)
	}
	if want.HasStop && got.Stop != want.Stop {
		t.Fatalf("stop %+v, want %+v", got.Stop, want.Stop)
	}
	if want.HasConstraintS && got.ConstraintS != want.ConstraintS {
		t.Fatalf("constraint %.17g, want %.17g", got.ConstraintS, want.ConstraintS)
	}
	if want.HasXYZMM && got.XYZMM != want.XYZMM {
		t.Fatalf("xyz mm %v, want %v", got.XYZMM, want.XYZMM)
	}
}

func TestRINEXClockHeaderFieldPublicPayloads(t *testing.T) {
	cases := []struct {
		fixture string
		want    []RINEXClockHeaderRecord
	}{
		{fixture: "a17", want: []RINEXClockHeaderRecord{
			{FieldKind: RINEXClockHeaderFieldVersionType, TextParts: []string{"C", "G"}, HasVersion: true, Version: 3.04},
			{FieldKind: RINEXClockHeaderFieldProgramRunByDate, TextParts: []string{"TORINEXC V9.9", "USNO", "19960403  001000 UTC"}},
			{FieldKind: RINEXClockHeaderFieldComment, TextParts: []string{"EXAMPLE OF A CLOCK DATA ANALYSIS FILE"}},
			{FieldKind: RINEXClockHeaderFieldComment, TextParts: []string{"IN THIS CASE ANALYSIS RESULTS FROM GPS ONLY ARE INCLUDED"}},
			{FieldKind: RINEXClockHeaderFieldComment, TextParts: []string{"No re-alignment of the clocks has been applied."}},
			{FieldKind: RINEXClockHeaderFieldObservationTypes, TextParts: []string{"C1W", "L1W", "C2W", "L2W"}, HasSystemCode: true, SystemCode: 71, HasCount: true, Count: 4},
			{FieldKind: RINEXClockHeaderFieldTimeSystem, TextParts: []string{"GPS"}},
			{FieldKind: RINEXClockHeaderFieldLeapSeconds, HasInteger: true, Integer: 10},
			{FieldKind: RINEXClockHeaderFieldDCBSApplied, TextParts: []string{"G", "CC2NONCC", "p1c1bias.hist @ goby.nrl.navy.mil"}, Reading: RINEXClockHeaderReadingOtherVersionColumns},
			{FieldKind: RINEXClockHeaderFieldPCVSApplied, TextParts: []string{"G", "PAGES", "igs05.atx @ igscb.jpl.nasa.gov"}, Reading: RINEXClockHeaderReadingOtherVersionColumns},
			{FieldKind: RINEXClockHeaderFieldTypesOfData, TextParts: []string{"AS", "AR"}, HasCount: true, Count: 2},
			{FieldKind: RINEXClockHeaderFieldAnalysisCenter, TextParts: []string{"USN", "USNO USING GIPSY/OASIS-II"}},
			{FieldKind: RINEXClockHeaderFieldClockRefCount, HasCount: true, Count: 1, HasStart: true, Start: CivilDateTime{Year: 1994, Month: 7, Day: 14}, HasStop: true, Stop: CivilDateTime{Year: 1994, Month: 7, Day: 14, Hour: 20, Minute: 59}, Reading: RINEXClockHeaderReadingOtherVersionColumns},
			{FieldKind: RINEXClockHeaderFieldAnalysisClockRef, TextParts: []string{"USNO", "40451S003"}, HasConstraintS: true, ConstraintS: -0.123456789012},
			{FieldKind: RINEXClockHeaderFieldClockRefCount, HasCount: true, Count: 1, HasStart: true, Start: CivilDateTime{Year: 1994, Month: 7, Day: 14, Hour: 21}, HasStop: true, Stop: CivilDateTime{Year: 1994, Month: 7, Day: 14, Hour: 21, Minute: 59}, Reading: RINEXClockHeaderReadingOtherVersionColumns},
			{FieldKind: RINEXClockHeaderFieldAnalysisClockRef, TextParts: []string{"TIDB", "50103M108"}, HasConstraintS: true, ConstraintS: -0.123456789012},
			{FieldKind: RINEXClockHeaderFieldSolutionStationCount, HasCount: true, Count: 4, TextParts: []string{"ITRF96"}},
			{FieldKind: RINEXClockHeaderFieldSolutionStation, TextParts: []string{"GOLD", "40405S031"}, HasXYZMM: true, XYZMM: [3]int64{1234567890, -1234567890, -1234567890}},
			{FieldKind: RINEXClockHeaderFieldSolutionStation, TextParts: []string{"AREQ", "42202M005"}, HasXYZMM: true, XYZMM: [3]int64{-1234567890, 1234567890, -1234567890}},
			{FieldKind: RINEXClockHeaderFieldSolutionStation, TextParts: []string{"TIDB", "50103M108"}, HasXYZMM: true, XYZMM: [3]int64{1234567890, -1234567890, 1234567890}},
			{FieldKind: RINEXClockHeaderFieldSolutionStation, TextParts: []string{"HARK", "30302M007"}, HasXYZMM: true, XYZMM: [3]int64{-1234567890, 1234567890, -1234567890}},
			{FieldKind: RINEXClockHeaderFieldSolutionStation, TextParts: []string{"USNO", "40451S003"}, HasXYZMM: true, XYZMM: [3]int64{1234567890, -1234567890, -1234567890}},
			{FieldKind: RINEXClockHeaderFieldSolutionSatelliteCount, HasCount: true, Count: 27},
			{FieldKind: RINEXClockHeaderFieldPRNList, TextParts: []string{"G01", "G02", "G03", "G04", "G05", "G06", "G07", "G08", "G09", "G10", "G13", "G14", "G15", "G16", "G17", "G18"}},
			{FieldKind: RINEXClockHeaderFieldPRNList, TextParts: []string{"G19", "G21", "G22", "G23", "G24", "G25", "G26", "G27", "G29", "G30", "G31"}},
			{FieldKind: RINEXClockHeaderFieldEndOfHeader},
		}},
		{fixture: "a18", want: []RINEXClockHeaderRecord{
			{FieldKind: RINEXClockHeaderFieldVersionType, TextParts: []string{"C", ""}, HasVersion: true, Version: 3.04},
			{FieldKind: RINEXClockHeaderFieldProgramRunByDate, TextParts: []string{"TORINEXC V9.9", "USNO", "19960403  001000 UTC"}},
			{FieldKind: RINEXClockHeaderFieldComment, TextParts: []string{"EXAMPLE OF A CLOCK DATA FILE"}},
			{FieldKind: RINEXClockHeaderFieldComment, TextParts: []string{"IN THIS CASE CALIBRATION/DISCONTINUITY DATA GIVEN"}},
			{FieldKind: RINEXClockHeaderFieldLeapSecondsGNSS, HasInteger: true, Integer: 10},
			{FieldKind: RINEXClockHeaderFieldTypesOfData, TextParts: []string{"CR", "DR"}, HasCount: true, Count: 2},
			{FieldKind: RINEXClockHeaderFieldStationNameNum, TextParts: []string{"USNO", "40451S003"}, Reading: RINEXClockHeaderReadingOtherVersionColumns},
			{FieldKind: RINEXClockHeaderFieldStationClockRef, TextParts: []string{"UTC(USNO) MASTER CLOCK VIA CONTINUOUS CABLE MONITOR"}},
			{FieldKind: RINEXClockHeaderFieldEndOfHeader},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			data := protocolFixture(t, "clk", "lossless", "rinex_clock304_table_"+tc.fixture+".clk")
			clock, err := ParseRINEXClock(data)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := clock.Close(); err != nil {
					t.Error(err)
				}
			})
			got, err := clock.HeaderRecords()
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("record count %d, want %d", len(got), len(tc.want))
			}
			for i := range got {
				checkClockHeaderPayload(t, got[i], tc.want[i])
			}
		})
	}
}

func TestRINEXClockHeaderFieldPublicOptionalPresence(t *testing.T) {
	data := protocolFixture(t, "clk", "lossless", "rinex_clock304_table_a17.clk")
	lines := bytes.Split(data, []byte("\n"))
	for i, line := range lines {
		if bytes.Contains(line, []byte("# OF CLK REF")) {
			lines[i] = []byte("     1" + string(bytes.Repeat([]byte{32}, 59)) + "# OF CLK REF")
			break
		}
	}
	clock, err := ParseRINEXClock(bytes.Join(lines, []byte("\n")))
	if err != nil {
		t.Fatal(err)
	}
	records, err := clock.HeaderRecords()
	if err != nil {
		t.Fatal(err)
	}
	if err := clock.Close(); err != nil {
		t.Fatal(err)
	}
	var refs []RINEXClockHeaderRecord
	for _, r := range records {
		if r.FieldKind == RINEXClockHeaderFieldClockRefCount {
			refs = append(refs, r)
		}
	}
	if len(refs) != 2 || !refs[0].HasCount || refs[0].Count != 1 || refs[0].HasStart || refs[0].HasStop || !refs[1].HasStart || !refs[1].HasStop {
		t.Fatalf("optional clock-ref fields = %+v", refs)
	}
	blank := bytes.Replace(data, []byte("USNO      40451S003                           -.123456789012E+00 ANALYSIS CLK REF"), []byte("USNO      40451S003"+string(bytes.Repeat([]byte{32}, 47))+"ANALYSIS CLK REF"), 1)
	clock, err = ParseRINEXClock(blank)
	if err != nil {
		t.Fatal(err)
	}
	records, err = clock.HeaderRecords()
	if err != nil {
		t.Fatal(err)
	}
	if err := clock.Close(); err != nil {
		t.Fatal(err)
	}
	for _, r := range records {
		if r.FieldKind == RINEXClockHeaderFieldAnalysisClockRef && r.TextParts[0] == "USNO" && r.TextParts[1] == "40451S003" {
			if r.HasConstraintS {
				t.Fatalf("blank constraint present: %+v", r)
			}
			return
		}
	}
	t.Fatal("USNO analysis clock reference missing")
}
