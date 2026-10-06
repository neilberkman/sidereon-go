package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"
)

func TestNMEAFieldErrorPreservesSignedWideCivilFields(t *testing.T) {
	const payload = `{"kind":"invalid_civil_components","fields":{"year":-120000,"month":4097,"day":900000,"hour":-33,"minute":900000,"second":{"decimal":"1.25","bits_hex":"3ff4000000000000"}}}`
	var got NMEAFieldError
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatal(err)
	}
	if got.Fields.Year == nil || *got.Fields.Year != -120000 || got.Fields.Month == nil || *got.Fields.Month != 4097 || got.Fields.Day == nil || *got.Fields.Day != 900000 || got.Fields.Hour == nil || *got.Fields.Hour != -33 || got.Fields.Minute == nil || *got.Fields.Minute != 900000 {
		t.Fatalf("civil diagnostic fields lost source integers: %+v", got.Fields)
	}
}

func TestNMEAAccumulatorChunkingFinishAndCopies(t *testing.T) {
	accumulator, err := NewNMEAAccumulator()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := accumulator.Close(); err != nil {
			t.Error(err)
		}
	})

	data := []byte(nmeaFixture)
	first, err := accumulator.Push(data[:18])
	if err != nil {
		t.Fatal(err)
	}
	if first.SentenceCount != 0 || first.CompletedEpochCount != 0 || first.RetainedLength != 18 {
		t.Fatalf("first chunk = %+v", first)
	}
	for i := range data[:18] {
		data[i] = 'X'
	}
	second, err := accumulator.Push(data[18:])
	if err != nil {
		t.Fatal(err)
	}
	if second.SentenceCount != 2 || second.CompletedEpochCount != 1 || second.RetainedLength != 0 {
		t.Fatalf("second chunk = %+v", second)
	}
	retained, err := accumulator.RetainedLength()
	if err != nil || retained != 0 {
		t.Fatalf("retained length = %d, %v", retained, err)
	}
	finished, err := accumulator.Finish()
	if err != nil {
		t.Fatal(err)
	}
	if finished.CompletedEpochCount != 1 || finished.RetainedLength != 0 {
		t.Fatalf("finish = %+v", finished)
	}
	summary, err := accumulator.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.SentenceCount != 2 || summary.EpochCount != 2 || summary.SkipCount != 0 || summary.WarningCount != 0 {
		t.Fatalf("summary = %+v", summary)
	}
	epochs, err := accumulator.Epochs()
	if err != nil {
		t.Fatal(err)
	}
	if len(epochs) != 2 || !epochs[0].HasGGA || !epochs[0].HasPosition {
		t.Fatalf("epochs = %+v", epochs)
	}
}

func TestWriteNMEAGGA(t *testing.T) {
	options := DefaultNMEAGGAOptions()
	options.UTCSecondsOfDay = 3661.239
	options.Position = Geodetic{LatitudeRad: 40 * math.Pi / 180, LongitudeRad: -105 * math.Pi / 180, HeightM: 1600}
	value, err := WriteNMEAGGA(options)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "$GPGGA,010101.23,4000.0000000,N,10500.0000000,W,1,10,1.00,1600.0,M,0.0,M,,*49\r\n"
	if string(value) != expected {
		t.Fatalf("GGA = %q, want %q", value, expected)
	}
	options.Talker = "G"
	if _, err := WriteNMEAGGA(options); err == nil {
		t.Fatal("one-byte talker was accepted")
	}
}

func TestNMEAEpochInstantUsesCanonicalUTCValidation(t *testing.T) {
	readEpoch := func(body string) NMEAEpoch {
		t.Helper()
		log, err := ParseNMEA([]byte(nmeaTestSentence(body) + "\n"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := log.Close(); err != nil {
				t.Error(err)
			}
		})
		epochs, err := log.Epochs()
		if err != nil {
			t.Fatal(err)
		}
		if len(epochs) != 1 {
			t.Fatalf("epochs = %+v, want one", epochs)
		}
		return epochs[0]
	}

	ordinary := readEpoch("GPRMC,120000.123456789,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if !ordinary.HasInstantJ2000S || ordinary.InstantJ2000S != 0.12345678900000001 {
		t.Fatalf("ordinary instant = present %v, %.17g", ordinary.HasInstantJ2000S, ordinary.InstantJ2000S)
	}

	wholeLeap := readEpoch("GPRMC,235960,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if !wholeLeap.HasInstantJ2000S || wholeLeap.InstantJ2000S != 43_200 {
		t.Fatalf("whole leap instant = present %v, %.17g", wholeLeap.HasInstantJ2000S, wholeLeap.InstantJ2000S)
	}

	fractionalLeap := readEpoch("GPRMC,235960.123456789,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if !fractionalLeap.HasCalendarEpoch || fractionalLeap.CalendarEpoch.Second != 60.123456789 {
		t.Fatalf("fractional leap calendar = %+v", fractionalLeap)
	}
	if fractionalLeap.HasInstantJ2000S || fractionalLeap.InstantJ2000S != 0 {
		t.Fatalf("fractional leap instant = present %v, %.17g", fractionalLeap.HasInstantJ2000S, fractionalLeap.InstantJ2000S)
	}

	missingDate := readEpoch("GPGGA,120000,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,")
	if missingDate.HasInstantJ2000S || missingDate.InstantJ2000S != 0 {
		t.Fatalf("missing-date instant = present %v, %.17g", missingDate.HasInstantJ2000S, missingDate.InstantJ2000S)
	}

	missingTime := readEpoch("GPRMC,,A,4807.038,N,01131.000,E,0.0,0.0,010100,,,A")
	if missingTime.HasInstantJ2000S || missingTime.InstantJ2000S != 0 {
		t.Fatalf("missing-time instant = present %v, %.17g", missingTime.HasInstantJ2000S, missingTime.InstantJ2000S)
	}
}

func TestNMEAAccumulatorCloseIsIdempotent(t *testing.T) {
	accumulator, err := NewNMEAAccumulator()
	if err != nil {
		t.Fatal(err)
	}
	if err := accumulator.Close(); err != nil {
		t.Fatal(err)
	}
	if err := accumulator.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := accumulator.Summary(); !errors.Is(err, ErrClosed) {
		t.Fatalf("summary after close = %v, want ErrClosed", err)
	}
}

func TestNMEATypedSentenceAndEpochRecords(t *testing.T) {
	input := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\n" +
		"$GPRMC,123520,A,4807.038,N,01131.000,E,22.4,84.4,230394,3.1,W,A,S*72\n" +
		"$GNGSA,A,3,01,02,03,04,,,,,,,,,1.5,0.9,1.2,1*3B\n" +
		"$GPGSV,2,1,05,01,45,083,42,02,17,308,40,03,25,120,39,04,10,200,35,1*68\n" +
		"$GPGSV,2,2,05,05,05,010,30,,,,,,,,,1*53\n" +
		"$GPGST,123520,1.2,3.4,2.3,45.0,0.5,0.6,0.7*4E\n" +
		"$GPVTG,84.4,T,83.1,M,22.4,N,41.5,K,A*25\n" +
		"$GPGLL,4807.038,N,01131.000,E,123520,A,A*42\n" +
		"$GPZDA,123520,23,03,1994,00,00*48\n"
	log, err := ParseNMEA([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = log.Close() })
	var nilLog *NMEALog
	if _, err := nilLog.SentenceRecords(); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil log records = %v", err)
	}
	if _, err := log.EpochDiagnostics(-1); err == nil {
		t.Fatal("negative log epoch index succeeded")
	}
	if _, err := log.EpochDiagnostics(99); err == nil {
		t.Fatal("out-of-range epoch diagnostics succeeded")
	}
	records, err := log.SentenceRecords()
	if err != nil {
		t.Fatal(err)
	}
	wantKinds := []string{"gga", "rmc", "gsa", "gsv", "gsv", "gst", "vtg", "gll", "zda"}
	if len(records) != len(wantKinds) {
		t.Fatalf("sentence records = %d, want %d", len(records), len(wantKinds))
	}
	for i, want := range wantKinds {
		if records[i].Body.Kind != want {
			t.Fatalf("sentence %d kind = %q, want %q", i, records[i].Body.Kind, want)
		}
	}
	if records[0].Body.GGA == nil || records[1].Body.RMC == nil || records[2].Body.GSA == nil || records[3].Body.GSV == nil || records[5].Body.GST == nil || records[6].Body.VTG == nil || records[7].Body.GLL == nil || records[8].Body.ZDA == nil {
		t.Fatal("one of the eight typed NMEA body variants was not decoded")
	}
	if records[0].Body.GGA.Quality == nil || records[0].Body.GGA.Quality.Value != 1 {
		t.Fatalf("GGA quality = %+v", records[0].Body.GGA.Quality)
	}
	epochs, err := log.EpochRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(epochs) != 2 || len(epochs[1].GSA) != 1 || len(epochs[1].GSV) != 1 || len(epochs[1].GSV[0].Satellites) != 7 {
		t.Fatalf("epoch records = %+v", epochs)
	}
	if epochs[1].GSV[0].ClaimedInView == nil || *epochs[1].GSV[0].ClaimedInView != 5 {
		t.Fatalf("claimed in-view = %+v", epochs[1].GSV[0].ClaimedInView)
	}
	if epochs[1].GSV[0].Satellites[5].SatNumber != nil || epochs[1].GSV[0].Satellites[6].SatNumber != nil {
		t.Fatal("empty trailing GSV slots were not retained")
	}
	firstRaw := append([]byte(nil), records[0].Raw...)
	records[0].Raw[0] = '!'
	again, err := log.SentenceRecords()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again[0].Raw, firstRaw) {
		t.Fatal("record payload shares storage with a prior result")
	}
	encoded, err := json.Marshal(records[3].Body)
	if err != nil {
		t.Fatal(err)
	}
	var roundTrip NMEASentenceBody
	if err := json.Unmarshal(encoded, &roundTrip); err != nil || roundTrip.GSV == nil || roundTrip.GSV.SatellitesInView == nil {
		t.Fatalf("GSV typed round trip: %s, %v", encoded, err)
	}
	rmcBody, err := json.Marshal(records[1].Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rmcBody, &roundTrip); err != nil || roundTrip.GSV != nil || roundTrip.RMC == nil {
		t.Fatalf("variant reuse retained stale fields: %+v, %v", roundTrip, err)
	}
}

func TestNMEAGSVPreservesSignedElevationAndWideOptionalCount(t *testing.T) {
	line := nmeaTestSentence("GPGSV,1,1,300,05,-45,359,99")
	log, err := ParseNMEA([]byte(line + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, log)
	records, err := log.SentenceRecords()
	if err != nil {
		t.Fatal(err)
	}
	value := records[0].Body.GSV
	if value == nil || value.SatellitesInView == nil || *value.SatellitesInView != 300 {
		t.Fatalf("GSV count = %+v", value)
	}
	if len(value.Satellites) != 1 || value.Satellites[0].ElevationDeg == nil || *value.Satellites[0].ElevationDeg != -45 {
		t.Fatalf("GSV satellite = %+v", value.Satellites)
	}
	if value.Satellites[0].AzimuthDeg == nil || *value.Satellites[0].AzimuthDeg != 359 || value.Satellites[0].CN0DBHz == nil || *value.Satellites[0].CN0DBHz != 99 {
		t.Fatalf("GSV satellite values = %+v", value.Satellites[0])
	}
	nullLog, err := ParseNMEA([]byte(nmeaTestSentence("GPGSV,1,1,,05,-45,359,99") + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, nullLog)
	nullRecords, err := nullLog.SentenceRecords()
	if err != nil {
		t.Fatal(err)
	}
	if nullRecords[0].Body.GSV == nil || nullRecords[0].Body.GSV.SatellitesInView != nil {
		t.Fatalf("absent in-view count lost its null value: %+v", nullRecords[0].Body.GSV)
	}
}

func TestNMEATypedDiagnosticsAndFinishTailRetention(t *testing.T) {
	good := nmeaTestSentence("GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,")
	log, err := ParseNMEA([]byte("prefix" + good + "\nbad line\n" + good + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	diagnostics, err := log.Diagnostics()
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 3 || diagnostics[0].Kind != NMEADiagnosticSkip || diagnostics[0].Skip == nil || diagnostics[0].Skip.Fields.Reason.Kind != "unknown_block" || diagnostics[0].Skip.Fields.Reason.Fields.Block == nil || *diagnostics[0].Skip.Fields.Reason.Fields.Block != "no NMEA start delimiter" {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	if diagnostics[0].Skip.Fields.At.Line == nil || *diagnostics[0].Skip.Fields.At.Line != 2 {
		t.Fatalf("skip record reference = %+v", diagnostics[0].Skip.Fields.At)
	}
	if diagnostics[1].Warning == nil || diagnostics[1].Warning.Fields.WarningKind != "mismatch" {
		t.Fatalf("parser warning = %+v", diagnostics[1])
	}
	if diagnostics[0].DecodeError != nil || diagnostics[1].DecodeError != nil {
		t.Fatalf("typed diagnostics failed to decode: %v %v", diagnostics[0].DecodeError, diagnostics[1].DecodeError)
	}
	if !json.Valid(diagnostics[0].Payload) {
		t.Fatalf("diagnostic payload is not valid JSON: %s", diagnostics[0].Payload)
	}
	savedDiagnostic := append([]byte(nil), diagnostics[0].Payload...)
	_ = log.Close()
	if !bytes.Equal(diagnostics[0].Payload, savedDiagnostic) {
		t.Fatal("diagnostic payload changed after closing its source log")
	}
	if _, err := log.Diagnostics(); !errors.Is(err, ErrClosed) {
		t.Fatalf("diagnostics after close = %v", err)
	}

	accumulator, err := NewNMEAAccumulator()
	if err != nil {
		t.Fatal(err)
	}
	first := "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n"
	final := "$GPGGA,123520,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,"
	if _, err := accumulator.EpochDiagnostics(-1); err == nil {
		t.Fatal("negative accumulator epoch index succeeded")
	}
	if _, err := accumulator.Push([]byte(first + final)); err != nil {
		t.Fatal(err)
	}
	finished, err := accumulator.Finish()
	if err != nil {
		t.Fatal(err)
	}
	if finished.SentenceCount != 1 || finished.CompletedEpochCount != 2 || finished.WarningCount != 1 {
		t.Fatalf("finish summary = %+v", finished)
	}
	sentences, err := accumulator.SentenceRecords()
	if err != nil {
		t.Fatal(err)
	}
	epochs, err := accumulator.EpochRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(sentences) != 2 || sentences[1].Body.GGA == nil || len(epochs) != 2 || epochs[1].Time == nil || epochs[1].Time.Second != 20 {
		t.Fatalf("final tail records = %d sentences %d epochs", len(sentences), len(epochs))
	}
	detail, err := accumulator.Diagnostics()
	if err != nil {
		t.Fatal(err)
	}
	if len(detail) != 1 || detail[0].Warning == nil || detail[0].Warning.Fields.WarningKind != "missing_metadata" {
		t.Fatalf("final tail diagnostics = %+v", detail)
	}
	if detail[0].Warning.Fields.At.Line == nil || *detail[0].Warning.Fields.At.Line != 2 {
		t.Fatalf("final warning line = %+v", detail[0].Warning.Fields.At.Line)
	}
	if err := accumulator.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := accumulator.EpochRecords(); !errors.Is(err, ErrClosed) {
		t.Fatalf("records after close = %v", err)
	}
	malformed, err := NewNMEAAccumulator()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := malformed.Push([]byte("bad line")); err != nil {
		t.Fatal(err)
	}
	malformedFinish, err := malformed.Finish()
	if err != nil {
		t.Fatal(err)
	}
	if malformedFinish.SkipCount != 1 {
		t.Fatalf("malformed final tail = %+v", malformedFinish)
	}
	malformedDiagnostics, err := malformed.Diagnostics()
	if err != nil {
		t.Fatal(err)
	}
	if len(malformedDiagnostics) != 1 || malformedDiagnostics[0].Skip == nil || malformedDiagnostics[0].Skip.Fields.Reason.Kind != "unknown_block" {
		t.Fatalf("malformed final-tail diagnostics = %+v", malformedDiagnostics)
	}
	if err := malformed.Close(); err != nil {
		t.Fatal(err)
	}
}

func nmeaTestSentence(body string) string {
	var checksum byte
	for _, value := range []byte(body) {
		checksum ^= value
	}
	return fmt.Sprintf("$%s*%02X", body, checksum)
}
