package sidereon

import (
	"errors"
	"math"
	"sync"
	"testing"
)

func TestNavigationRemainingSmoke(t *testing.T) {
	if value, present, err := GNSSWeekEpochJulianDayNumber(GPST); err != nil || !present || value != 2444245 {
		t.Fatalf("GNSS week epoch = %d/%v, %v", value, present, err)
	}
	value, err := GNSSWeekTowNew(GPST, 1, -1)
	if err != nil {
		t.Fatal(err)
	}
	if value.Week != 1 || value.TOWSeconds != -1 {
		t.Fatalf("unexpected GNSS value: %+v", value)
	}
	normalized, err := GNSSWeekTowNormalized(value)
	if err != nil || normalized.Week != 0 || normalized.TOWSeconds != 604799 {
		t.Fatalf("normalized GNSS value: %+v, %v", normalized, err)
	}
	if week, err := GNSSWeekTowUnrolledWeek(GNSSWeekTow{System: GPST, Week: 5}, 2); err != nil || week != 2053 {
		t.Fatalf("unrolled week = %d, %v", week, err)
	}
	if week, present, err := GNSSWeekFromCalendar(GPST, 1980, 1, 6); err != nil || !present || week != 0 {
		t.Fatalf("calendar week = %d/%v, %v", week, present, err)
	}
	if _, err := GNSSWeekTowNew(TimeScale(99), 0, 0); err == nil {
		t.Fatal("invalid time scale unexpectedly accepted")
	}

	leap, err := LeapSecondTableInfo()
	if err != nil || leap.Entries == 0 || leap.SourceLen == 0 || leap.FirstMJD > leap.LastMJD {
		t.Fatalf("leap info = %+v, %v", leap, err)
	}
	leapSource, err := LeapSecondTableSource()
	if err != nil || len(leapSource) != leap.SourceLen {
		t.Fatalf("leap source = %d/%d, %v", len(leapSource), leap.SourceLen, err)
	}
	ut1, err := UT1CoverageInfo()
	if err != nil || ut1.Entries == 0 || ut1.SourceLen == 0 || ut1.FirstMJD > ut1.LastMJD || ut1.FirstJDTT > ut1.LastJDTT {
		t.Fatalf("UT1 info = %+v, %v", ut1, err)
	}
	if covered, err := UT1CoverageCoversJdTt(ut1.FirstJDTT); err != nil || !covered {
		t.Fatalf("UT1 lower boundary = %v, %v", covered, err)
	}
	if covered, err := UT1CoverageCoversJdTt(ut1.LastJDTT + 10000); err != nil || covered {
		t.Fatalf("UT1 outside = %v, %v", covered, err)
	}
	ut1Source, err := UT1CoverageSource()
	if err != nil || len(ut1Source) != ut1.SourceLen {
		t.Fatalf("UT1 source = %d/%d, %v", len(ut1Source), ut1.SourceLen, err)
	}

	parity, err := LNAVParity(make([]byte, 24), 0, 0)
	if err != nil || len(parity) != 6 {
		t.Fatalf("LNAV parity = %v, %v", parity, err)
	}
	if !equalBytes(parity, []byte{0, 0, 0, 0, 0, 0}) {
		t.Fatalf("zero LNAV parity = %v", parity)
	}
	if _, err := LNAVParity(make([]byte, 23), 0, 0); err == nil {
		t.Fatal("short LNAV parity input unexpectedly accepted")
	}
	params := LNAVParams{WeekNumber: 100, L2Code: 0, L2PDataFlag: 0, URAIndex: 0, SVHealth: 0, IODC: 100, TOC: 100000, IODE: 100, Eccentricity: 0.01, SqrtA: 5153.795, TOE: 100000, FitIntervalFlag: 0, AODO: 0, I0: 0.9}
	sf1, sf2, sf3, err := LNAVEncode(params, LNAVOptions{TOW: 100})
	if err != nil || len(sf1) != 300 || len(sf2) != 300 || len(sf3) != 300 {
		t.Fatalf("LNAV encode lengths = %d/%d/%d, %v", len(sf1), len(sf2), len(sf3), err)
	}
	encodedParity, err := LNAVParity(sf1[:24], 0, 0)
	if err != nil || !equalBytes(encodedParity, sf1[24:30]) {
		t.Fatalf("LNAV encoded parity = %v, want %v (err %v)", encodedParity, sf1[24:30], err)
	}
	decoded, err := LNAVDecode(sf1, sf2, sf3)
	if err != nil || decoded.WeekNumber != params.WeekNumber || decoded.IODE != params.IODE {
		t.Fatalf("LNAV decode = %+v, %v", decoded, err)
	}
	if valid, err := LNAVParityValid(sf1[:30], 0, 0); err != nil || !valid {
		t.Fatalf("LNAV parity validation = %v, %v", valid, err)
	}
	if id, err := LNAVSubframeID(sf1); err != nil || id != 1 {
		t.Fatalf("LNAV subframe id = %d, %v", id, err)
	}
	if tow, err := LNAVTOW(sf1); err != nil || tow != 100 {
		t.Fatalf("LNAV TOW = %d, %v", tow, err)
	}

	ray := NequickGRay{Month: 1, UTCHours: 12, StationLonDeg: 0, StationLatDeg: 45, StationHeightM: 100, SatelliteLonDeg: 10, SatelliteLatDeg: 20, SatelliteHeightM: 20200000}
	stec, err := NequickGStecTecu(1, 0, 0, ray)
	if err != nil || math.Abs(stec-1.801691477554427) > 1e-14 {
		t.Fatalf("NeQuick TEC = %v, %v", stec, err)
	}
	delay, err := NequickGDelayM(1, 0, 0, ray, 1575420000)
	if err != nil || math.Abs(delay-0.29254505487201443) > 1e-14 {
		t.Fatalf("NeQuick delay = %v, %v", delay, err)
	}
	galileoDelay, err := GalileoNequickGNative(1, 0, 0, 45, 0, 30, 43200, 100, 1575420000)
	if err != nil || math.Abs(galileoDelay-0.71417989604392706) > 1e-14 {
		t.Fatalf("Galileo NeQuick delay = %v, %v", galileoDelay, err)
	}
	klobucharDelay, err := KlobucharNative([4]float64{0, 0, 0, 0}, [4]float64{0, 0, 0, 0}, 0, 0, 0, 45, 43200, 1575420000)
	if err != nil || math.Abs(klobucharDelay-2.02544581304128) > 1e-14 {
		t.Fatalf("Klobuchar delay = %v, %v", klobucharDelay, err)
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestPropagatedEphemerisOwnershipAndClose(t *testing.T) {
	config, err := DefaultPropagationConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.PositionKm = [3]float64{7000, 0, 0}
	config.VelocityKmPerS = [3]float64{0, 7.5, 0}
	times := []float64{0, 60, 120}
	e, err := PropagateState(config, times)
	if err != nil {
		t.Fatal(err)
	}
	times[1] = 999999
	if count, err := e.EpochCount(); err != nil || count != 3 {
		t.Fatalf("epoch count = %d, %v", count, err)
	}
	actualTimes, err := e.TimesS()
	if err != nil || len(actualTimes) != 3 || actualTimes[1] != 60 {
		t.Fatalf("times = %v, %v", actualTimes, err)
	}
	states, err := e.States()
	if err != nil || len(states) != 3 || states[0].PositionKm != config.PositionKm {
		t.Fatalf("states = %+v, %v", states, err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = e.EpochCount(); _, _ = e.States() }()
	}
	wg.Add(1)
	go func() { defer wg.Done(); _ = e.Close() }()
	wg.Wait()
	if _, err := e.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("EpochCount after Close = %v", err)
	}
}
