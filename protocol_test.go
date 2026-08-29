package sidereon

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func protocolFixture(t *testing.T, relative ...string) []byte {
	t.Helper()
	path := filepath.Join(append([]string{"testdata"}, relative...)...)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read protocol fixture %s: %v", path, err)
	}
	return data
}

func TestRINEXNavAndClockPublicSmoke(t *testing.T) {
	nav := protocolFixture(t, "nav", "ESBC00DNK_R_20201770000_01D_MN.rnx")
	clockText := protocolFixture(t, "clk", "synthetic_rinex_clock.clk")
	navOriginal := append([]byte(nil), nav...)

	navRecords, err := ParseRINEXNavRecords(nav)
	if err != nil {
		t.Fatal(err)
	}
	nav[0] ^= 0xff
	t.Cleanup(func() {
		if err := navRecords.Close(); err != nil {
			t.Errorf("NAV Close: %v", err)
		}
	})
	count, err := navRecords.Count()
	if err != nil || count != 2216 {
		t.Fatalf("NAV count = %d, %v; want 2216", count, err)
	}
	record, err := navRecords.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	if record.SatelliteID != "C05" || record.Message != 8 || record.Issue != 1 || record.Week != 755 {
		t.Fatalf("unexpected first NAV identity: %+v", record)
	}
	if record.Toe.System != BDT || record.Toe.Week != 755 || math.Float64bits(record.Toe.TOWSeconds) != 0x4114a78000000000 {
		t.Fatalf("unexpected first NAV epoch: %+v", record.Toe)
	}
	if math.Float64bits(record.Elements.SqrtA) != 0x40b95d6102dfffec || math.Float64bits(record.Clock.AF0) != 0xbf40e400000000ca {
		t.Fatalf("unexpected first NAV numeric values: elements=%+v clock=%+v", record.Elements, record.Clock)
	}
	if !record.GroupDelays.HasBeiDouTGD1 || !record.GroupDelays.HasBeiDouTGD2 || record.CNAV.Present {
		t.Fatalf("unexpected first NAV optional values: %+v", record.GroupDelays)
	}

	encoded, err := EncodeRINEXNav([]BroadcastRecord{record})
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 810 || !bytes.HasSuffix(encoded, []byte("\n")) {
		t.Fatalf("encoded NAV length = %d, want 810", len(encoded))
	}
	reparsed, err := ParseRINEXNavRecords(encoded)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := reparsed.Close(); err != nil {
			t.Errorf("reparsed NAV Close: %v", err)
		}
	})
	reparsedRecord, err := reparsed.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(reparsedRecord.Elements.SqrtA) != math.Float64bits(record.Elements.SqrtA) || math.Float64bits(reparsedRecord.Clock.AF0) != math.Float64bits(record.Clock.AF0) {
		t.Fatalf("NAV round trip changed representative values: %+v", reparsedRecord)
	}

	badNAV := bytes.Replace(navOriginal, []byte("C05 2020"), []byte("C05 XXXX"), 1)
	lenient, err := ParseRINEXNavLenient(badNAV)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := lenient.Close(); err != nil {
			t.Errorf("lenient NAV Close: %v", err)
		}
	})
	lenientRecords, err := lenient.Records()
	if err != nil {
		t.Fatal(err)
	}
	skipped, err := lenient.Skipped()
	if err != nil {
		t.Fatal(err)
	}
	if len(lenientRecords) != 2215 || len(skipped) != 1 || skipped[0].SatelliteID != "C05" || skipped[0].Message != "bad/missing toc epoch field in record for C05" {
		t.Fatalf("unexpected lenient NAV result: records=%d skipped=%+v", len(lenientRecords), skipped)
	}
	if count, err := lenient.RecordCount(); err != nil || count != 2215 {
		t.Fatalf("lenient NAV record count = %d, %v", count, err)
	}
	if count, err := lenient.SkippedCount(); err != nil || count != 1 {
		t.Fatalf("lenient NAV skipped count = %d, %v", count, err)
	}
	if _, err := ParseRINEXNavRecords(badNAV); err == nil {
		t.Fatal("strict NAV parse unexpectedly accepted malformed block")
	}

	clock, err := ParseRINEXClock(clockText)
	if err != nil {
		t.Fatal(err)
	}
	clockText[0] ^= 0xff
	t.Cleanup(func() {
		if err := clock.Close(); err != nil {
			t.Errorf("clock Close: %v", err)
		}
	})
	satellites, err := clock.Satellites()
	if err != nil || len(satellites) != 2 || satellites[0] != "G05" || satellites[1] != "G24" {
		t.Fatalf("clock satellites = %v, %v", satellites, err)
	}
	if count, err := clock.SatelliteCount(); err != nil || count != 2 {
		t.Fatalf("clock satellite count = %d, %v", count, err)
	}
	if count, err := clock.SeriesCount(); err != nil || count != 2 {
		t.Fatalf("clock series count = %d, %v; want 2", count, err)
	}
	sampleCount, err := clock.SampleCount()
	if err != nil || sampleCount != 5 {
		t.Fatalf("clock sample count = %d, %v; want 5", sampleCount, err)
	}
	series, err := clock.SeriesFor("G05")
	if err != nil || series == nil {
		t.Fatalf("G05 series = %v, %v", series, err)
	}
	t.Cleanup(func() {
		if err := series.Close(); err != nil {
			t.Errorf("G05 series Close: %v", err)
		}
	})
	points, err := series.Samples()
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 3 || points[0].Epoch.Scale != GPST || math.Float64bits(points[1].BiasS) != 0xbf2a36e36f0d4275 {
		t.Fatalf("unexpected G05 samples: %+v", points)
	}
	if count, err := series.SampleCount(); err != nil || count != 3 {
		t.Fatalf("G05 sample count = %d, %v", count, err)
	}
	allSeries, err := clock.Series()
	if err != nil || len(allSeries) != 2 {
		t.Fatalf("clock series = %v, %v; want 2", allSeries, err)
	}
	for _, value := range allSeries {
		value := value
		t.Cleanup(func() {
			if err := value.Close(); err != nil {
				t.Errorf("clock aggregate series Close: %v", err)
			}
		})
	}
	if err := series.Close(); err != nil {
		t.Fatal(err)
	}
	gpsSeconds := time.Date(2026, 5, 13, 0, 0, 30, 0, time.UTC).Sub(time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)).Seconds()
	bias, available, err := clock.BiasAtGPSSeconds("G05", gpsSeconds)
	if err != nil || !available || math.Float64bits(bias) != math.Float64bits(points[1].BiasS) {
		t.Fatalf("clock interpolation = %.17g, %v, %v", bias, available, err)
	}
	missing, err := clock.SeriesFor("G99")
	if err != nil || missing != nil {
		t.Fatalf("missing clock series = %v, %v", missing, err)
	}
	serialized, err := clock.ToText()
	if err != nil {
		t.Fatal(err)
	}
	roundTripClock, err := ParseRINEXClock(serialized)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := roundTripClock.Close(); err != nil {
			t.Errorf("round-trip clock Close: %v", err)
		}
	})
	roundTripCount, err := roundTripClock.SampleCount()
	if err != nil || roundTripCount != 5 {
		t.Fatalf("clock round-trip sample count = %d, %v", roundTripCount, err)
	}

	malformedClock := []byte("     3.00           C                                       RINEX VERSION / TYPE\n" +
		"                    GPS                                                         TIME SYSTEM ID\n" +
		"                                                                        END OF HEADER\n" +
		"AS G05  2026 05 13 00 00  bad-second  1   2.0e-04\n")
	lossy, err := ParseRINEXClockLossy(malformedClock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := lossy.Close(); err != nil {
			t.Errorf("lossy clock Close: %v", err)
		}
	})
	lossyCount, err := lossy.SampleCount()
	if err != nil || lossyCount != 0 {
		t.Fatalf("lossy malformed clock count = %d, %v", lossyCount, err)
	}
	if _, err := ParseRINEXClock(malformedClock); err == nil {
		t.Fatal("strict clock parse unexpectedly accepted malformed row")
	}
}

func TestSBASAndBareSSRSurface(t *testing.T) {
	sbasBody, err := hex.DecodeString("5306000000000000000000000000000000000000000000000000000040")
	if err != nil {
		t.Fatal(err)
	}
	wantSBASBody := append([]byte(nil), sbasBody...)
	block, err := DecodeSBASBlock(sbasBody, SBASWireBody226)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := block.Close(); err != nil {
			t.Errorf("SBAS Close: %v", err)
		}
	})
	sbasBody[0] ^= 0xff
	if encoded, err := block.Encode(); err != nil || !bytes.Equal(encoded, wantSBASBody) {
		t.Fatalf("SBAS encode = %x, %v", encoded, err)
	}
	info, err := block.Info()
	if err != nil || info.MessageType != 1 || info.Kind != SBASMessagePRNMask || info.Form != SBASWireBody226 {
		t.Fatalf("SBAS info = %+v, %v", info, err)
	}
	mask, err := block.PRNMask()
	if err != nil || mask == nil || mask.Preamble != 0x53 || mask.IODP != 1 {
		t.Fatalf("SBAS PRN mask = %+v, %v", mask, err)
	}
	if _, err := block.FastCorrections(); err != nil {
		t.Fatal(err)
	}
	if err := block.Close(); err != nil {
		t.Fatal(err)
	}
	if err := block.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := block.Info(); !errors.Is(err, ErrClosed) {
		t.Fatalf("SBAS use after close = %v", err)
	}

	logText := []byte("2360 259200 120 1 : 5306000000000000000000000000000000000000000000000000000040\n")
	logs, err := ParseSBASRTKLIBLines(logText)
	if err != nil {
		t.Fatal(err)
	}
	logText[0] ^= 0xff
	t.Cleanup(func() {
		if err := logs.Close(); err != nil {
			t.Errorf("SBAS logs Close: %v", err)
		}
	})
	items, err := logs.Items()
	if err != nil || len(items) != 1 || items[0].SatelliteID != "S20" || items[0].Epoch.Week != 2360 || items[0].ByteCount != len(sbasBody) {
		t.Fatalf("SBAS log items = %+v, %v", items, err)
	}
	if count, err := logs.Count(); err != nil || count != 1 {
		t.Fatalf("SBAS log count = %d, %v", count, err)
	}
	payload, err := logs.Bytes(0)
	if err != nil || !bytes.Equal(payload, wantSBASBody) {
		t.Fatalf("SBAS log payload = %x, %v", payload, err)
	}
	if mapped, present, err := SBASPRNToSatelliteID(120); err != nil || !present || mapped != "S20" {
		t.Fatalf("SBAS PRN mapping = %q, %v, %v", mapped, present, err)
	}
	if prn, present, err := SatelliteIDToSBASPRN("S20"); err != nil || !present || prn != 120 {
		t.Fatalf("SBAS reverse mapping = %d, %v, %v", prn, present, err)
	}
	if _, present, err := SBASPRNToSatelliteID(119); err != nil || present {
		t.Fatalf("absent SBAS PRN mapping = present=%v, err=%v", present, err)
	}

	tests := []struct {
		name string
		body string
	}{
		{name: "code", body: "423546002c803da021883d97249290"},
		{name: "phase", body: "4f1546002c803da408623ffa071f0ee024a1ca2380"},
		{name: "ura", body: "425546002c803da021d2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := hex.DecodeString(test.body)
			if err != nil {
				t.Fatal(err)
			}
			message, err := DecodeSSRMessage(body)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := message.Close(); err != nil {
					t.Errorf("SSR Close: %v", err)
				}
			})
			wantBody := append([]byte(nil), body...)
			body[0] ^= 0xff
			encodedBody, err := message.Encode()
			if err != nil || !bytes.Equal(encodedBody, wantBody) {
				t.Fatalf("SSR Encode = %x, want retained body %x", encodedBody, wantBody)
			}
			encodedBody[0] ^= 0xff
			bodyCopy, err := message.Body()
			if err != nil || !bytes.Equal(bodyCopy, wantBody) {
				t.Fatal("SSR Body did not return an independent copy")
			}
			bodyCopy[0] ^= 0xff
			encodedAgain, err := message.Encode()
			if err != nil || !bytes.Equal(encodedAgain, wantBody) {
				t.Fatalf("SSR Encode changed after Body mutation: %x, %v", encodedAgain, err)
			}
			info, err := message.Info()
			if err != nil {
				t.Fatal(err)
			}
			switch test.name {
			case "code":
				if info.MessageNumber != 1059 || info.Kind != RTCMSSRCodeBias || info.CodeBiasCount != 1 {
					t.Fatalf("code info = %+v", info)
				}
				groups, err := message.CodeBiasGroups()
				if err != nil || len(groups) != 1 || groups[0].Record.SatelliteID != 3 || len(groups[0].Signals) != 2 || groups[0].Signals[0].Bias != -1234 || groups[0].Signals[1].Bias != 2345 {
					t.Fatalf("code groups = %+v, %v", groups, err)
				}
			case "phase":
				if info.MessageNumber != 1265 || info.Kind != RTCMSSRPhaseBias || info.PhaseBiasCount != 1 {
					t.Fatalf("phase info = %+v", info)
				}
				groups, err := message.PhaseBiasGroups()
				if err != nil || len(groups) != 1 || groups[0].Record.YawAngle != 127 || groups[0].Record.YawRate != -12 || len(groups[0].Signals) != 2 || groups[0].Signals[0].Bias != -123456 || groups[0].Signals[1].Bias != 234567 {
					t.Fatalf("phase groups = %+v, %v", groups, err)
				}
			case "ura":
				if info.MessageNumber != 1061 || info.Kind != RTCMSSRURA || info.URACount != 1 {
					t.Fatalf("URA info = %+v", info)
				}
				values, err := message.URA()
				if err != nil || len(values) != 1 || values[0].SatelliteID != 3 || values[0].URAIndex != 41 {
					t.Fatalf("URA values = %+v, %v", values, err)
				}
			}
		})
	}
	if _, err := DecodeSSRMessage([]byte{0x42}); err == nil {
		t.Fatal("malformed bare SSR unexpectedly decoded")
	}
}

func TestProtocolBoundaryValidationAndOwnership(t *testing.T) {
	nav := protocolFixture(t, "nav", "ESBC00DNK_R_20201770000_01D_MN.rnx")
	navRecords, err := ParseRINEXNavRecords(nav)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := navRecords.Close(); err != nil {
			t.Errorf("NAV Close: %v", err)
		}
	})
	if _, err := navRecords.Record(-1); err == nil {
		t.Fatal("negative NAV record index was accepted")
	}
	if err := navRecords.Close(); err != nil {
		t.Fatal(err)
	}
	if err := navRecords.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := navRecords.Count(); !errors.Is(err, ErrClosed) {
		t.Fatalf("NAV Count after Close = %v, want ErrClosed", err)
	}

	clockText := protocolFixture(t, "clk", "synthetic_rinex_clock.clk")
	clock, err := ParseRINEXClock(clockText)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := clock.Close(); err != nil {
			t.Errorf("clock Close: %v", err)
		}
	})
	if _, err := clock.SeriesAt(-1); err == nil {
		t.Fatal("negative clock series index was accepted")
	}
	if _, err := clock.SeriesFor("G05\x00suffix"); err == nil {
		t.Fatal("embedded NUL in SeriesFor was accepted")
	}
	if _, _, err := clock.BiasAtGPSSeconds("G05\x00suffix", 0); err == nil {
		t.Fatal("embedded NUL in BiasAtGPSSeconds was accepted")
	}
	if err := clock.Close(); err != nil {
		t.Fatal(err)
	}
	if err := clock.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := clock.SampleCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("clock SampleCount after Close = %v, want ErrClosed", err)
	}

	var record BroadcastRecord
	record.SatelliteID = "C05\x00suffix"
	if _, err := EncodeRINEXNav([]BroadcastRecord{record}); err == nil {
		t.Fatal("embedded NUL in NAV satellite token was accepted")
	}
	record.SatelliteID = strings.Repeat("X", 17)
	if _, err := EncodeRINEXNav([]BroadcastRecord{record}); err == nil {
		t.Fatal("overlong NAV satellite token was accepted")
	}

	sbasBody, err := hex.DecodeString("5306000000000000000000000000000000000000000000000000000040")
	if err != nil {
		t.Fatal(err)
	}
	block, err := DecodeSBASBlock(sbasBody, SBASWireBody226)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := block.Close(); err != nil {
			t.Errorf("SBAS Close: %v", err)
		}
	})
	if _, err := block.LongTermHalf(-1); err == nil {
		t.Fatal("negative SBAS half index was accepted")
	}
	if _, err := block.LongTermRecords(-1); err == nil {
		t.Fatal("negative SBAS record index was accepted")
	}
	if err := block.Close(); err != nil {
		t.Fatal(err)
	}

	logText := []byte("2360 259200 120 1 : 5306000000000000000000000000000000000000000000000000000040\n")
	logs, err := ParseSBASRTKLIBLines(logText)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := logs.Close(); err != nil {
			t.Errorf("SBAS logs Close: %v", err)
		}
	})
	if _, err := logs.Bytes(-1); err == nil {
		t.Fatal("negative SBAS log index was accepted")
	}
	if err := logs.Close(); err != nil {
		t.Fatal(err)
	}

	body, err := hex.DecodeString("423546002c803da021883d97249290")
	if err != nil {
		t.Fatal(err)
	}
	message, err := DecodeSSRMessage(body)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := message.Close(); err != nil {
			t.Errorf("SSR Close: %v", err)
		}
	})
	if _, err := message.CodeBiasSignals(-1); err == nil {
		t.Fatal("negative SSR code-bias index was accepted")
	}
	if _, err := message.PhaseBiasSignals(-1); err == nil {
		t.Fatal("negative SSR phase-bias index was accepted")
	}
	if err := message.Close(); err != nil {
		t.Fatal(err)
	}
	if err := message.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := message.Info(); !errors.Is(err, ErrClosed) {
		t.Fatalf("SSR Info after Close = %v, want ErrClosed", err)
	}
	if _, err := message.Body(); !errors.Is(err, ErrClosed) {
		t.Fatalf("SSR Body after Close = %v, want ErrClosed", err)
	}
	if _, err := message.Encode(); !errors.Is(err, ErrClosed) {
		t.Fatalf("SSR Encode after Close = %v, want ErrClosed", err)
	}
}

func TestProtocolCloseReadRace(t *testing.T) {
	body, err := hex.DecodeString("423546002c803da021883d97249290")
	if err != nil {
		t.Fatal(err)
	}
	message, err := DecodeSSRMessage(body)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := message.Close(); err != nil {
			t.Errorf("SSR Close: %v", err)
		}
	})
	var group sync.WaitGroup
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for repeat := 0; repeat < 20; repeat++ {
				_, readErr := message.Info()
				if readErr != nil && !errors.Is(readErr, ErrClosed) {
					t.Errorf("concurrent SSR Info: %v", readErr)
				}
			}
		}()
	}
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for repeat := 0; repeat < 20; repeat++ {
				if closeErr := message.Close(); closeErr != nil {
					t.Errorf("concurrent SSR Close: %v", closeErr)
				}
			}
		}()
	}
	group.Wait()
	if _, err := message.Info(); !errors.Is(err, ErrClosed) {
		t.Fatalf("SSR Info after concurrent Close = %v", err)
	}
}

func runProtocolCloseReadRace(t *testing.T, read func() error, closeHandle func() error) {
	t.Helper()
	var group sync.WaitGroup
	errorsCh := make(chan error, 256)
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for repeat := 0; repeat < 20; repeat++ {
				if err := read(); err != nil && !errors.Is(err, ErrClosed) {
					errorsCh <- err
				}
			}
		}()
	}
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for repeat := 0; repeat < 20; repeat++ {
				if err := closeHandle(); err != nil {
					errorsCh <- err
				}
			}
		}()
	}
	group.Wait()
	close(errorsCh)
	for err := range errorsCh {
		t.Errorf("concurrent read/Close: %v", err)
	}
}

func TestSSRBodyAndEncodeCloseRace(t *testing.T) {
	body, err := hex.DecodeString("423546002c803da021883d97249290")
	if err != nil {
		t.Fatal(err)
	}
	operations := []struct {
		name string
		read func(*SSRMessage) error
	}{
		{name: "Body", read: func(message *SSRMessage) error {
			_, err := message.Body()
			return err
		}},
		{name: "Encode", read: func(message *SSRMessage) error {
			_, err := message.Encode()
			return err
		}},
	}
	for _, operation := range operations {
		operation := operation
		t.Run(operation.name, func(t *testing.T) {
			message, err := DecodeSSRMessage(body)
			if err != nil {
				t.Fatal(err)
			}
			runProtocolCloseReadRace(t, func() error {
				return operation.read(message)
			}, message.Close)
			if err := operation.read(message); !errors.Is(err, ErrClosed) {
				t.Fatalf("%s after concurrent Close = %v, want ErrClosed", operation.name, err)
			}
		})
	}
}

func TestProtocolOwnedHandlesCloseReadRace(t *testing.T) {
	nav := protocolFixture(t, "nav", "ESBC00DNK_R_20201770000_01D_MN.rnx")
	navRecords, err := ParseRINEXNavRecords(nav)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := navRecords.Close(); err != nil {
			t.Errorf("NAV Close: %v", err)
		}
	})
	t.Run("RINEX NAV records", func(t *testing.T) {
		runProtocolCloseReadRace(t, func() error {
			_, err := navRecords.Count()
			return err
		}, navRecords.Close)
	})

	navParse, err := ParseRINEXNavLenient(nav)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := navParse.Close(); err != nil {
			t.Errorf("lenient NAV Close: %v", err)
		}
	})
	t.Run("RINEX NAV parse", func(t *testing.T) {
		runProtocolCloseReadRace(t, func() error {
			_, err := navParse.RecordCount()
			return err
		}, navParse.Close)
	})

	clockText := protocolFixture(t, "clk", "synthetic_rinex_clock.clk")
	clock, err := ParseRINEXClock(clockText)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := clock.Close(); err != nil {
			t.Errorf("clock Close: %v", err)
		}
	})
	t.Run("RINEX clock", func(t *testing.T) {
		runProtocolCloseReadRace(t, func() error {
			_, err := clock.SampleCount()
			return err
		}, clock.Close)
	})

	series, err := ParseRINEXClock(clockText)
	if err != nil {
		t.Fatal(err)
	}
	clockSeries, err := series.SeriesFor("G05")
	if err != nil || clockSeries == nil {
		if series != nil {
			if closeErr := series.Close(); closeErr != nil {
				t.Errorf("temporary clock Close: %v", closeErr)
			}
		}
		t.Fatalf("G05 series = %v, %v", clockSeries, err)
	}
	t.Cleanup(func() {
		if err := clockSeries.Close(); err != nil {
			t.Errorf("clock series Close: %v", err)
		}
		if err := series.Close(); err != nil {
			t.Errorf("series owner Close: %v", err)
		}
	})
	t.Run("RINEX clock series", func(t *testing.T) {
		runProtocolCloseReadRace(t, func() error {
			_, err := clockSeries.SampleCount()
			return err
		}, clockSeries.Close)
	})

	sbasBody, err := hex.DecodeString("5306000000000000000000000000000000000000000000000000000040")
	if err != nil {
		t.Fatal(err)
	}
	block, err := DecodeSBASBlock(sbasBody, SBASWireBody226)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := block.Close(); err != nil {
			t.Errorf("SBAS Close: %v", err)
		}
	})
	t.Run("SBAS block", func(t *testing.T) {
		runProtocolCloseReadRace(t, func() error {
			_, err := block.Info()
			return err
		}, block.Close)
	})

	logText := []byte("2360 259200 120 1 : 5306000000000000000000000000000000000000000000000000000040\n")
	logs, err := ParseSBASRTKLIBLines(logText)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := logs.Close(); err != nil {
			t.Errorf("SBAS logs Close: %v", err)
		}
	})
	t.Run("SBAS logs", func(t *testing.T) {
		runProtocolCloseReadRace(t, func() error {
			_, err := logs.Count()
			return err
		}, logs.Close)
	})

	body, err := hex.DecodeString("423546002c803da021883d97249290")
	if err != nil {
		t.Fatal(err)
	}
	message, err := DecodeSSRMessage(body)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := message.Close(); err != nil {
			t.Errorf("SSR Close: %v", err)
		}
	})
	t.Run("SSR message", func(t *testing.T) {
		runProtocolCloseReadRace(t, func() error {
			_, err := message.Info()
			return err
		}, message.Close)
	})
}
