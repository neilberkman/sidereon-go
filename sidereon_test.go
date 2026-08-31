package sidereon

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

const nmeaFixture = "$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n" +
	"$GPGGA,123520,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*4D\r\n"

func TestVendoredHeaderMatchesPinnedPublicHeader(t *testing.T) {
	got, err := os.ReadFile(filepath.Join("internal", "native", "include", "sidereon.h"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(got)
	if value := hex.EncodeToString(digest[:]); value != "92a698c0bd124d05f5121c0995dd38645eedd25eb841012bfff5d1fd342010ce" {
		t.Fatalf("vendored header digest = %s, want pinned digest", value)
	}
}

func TestLibraryVersionAndCalendarValues(t *testing.T) {
	version := LibraryVersion()
	if version.Major != 1 || version.Minor != 3 || version.Patch != 3 || version.String != "1.3.3" {
		t.Fatalf("unexpected library version: %+v", version)
	}

	second, err := SecondOfDay(1, 2, 3.5)
	if err != nil || second != 3723.5 {
		t.Fatalf("SecondOfDay = %v, %v", second, err)
	}
	day, err := DayOfYear(2024, 2, 29, 12, 0, 0.25)
	if err != nil || day != 60.50000289351852 {
		t.Fatalf("DayOfYear = %.17g, %v", day, err)
	}
	productDay, err := DataDayOfYear(2020, 3, 1)
	if err != nil || productDay != 61 {
		t.Fatalf("DataDayOfYear = %v, %v", productDay, err)
	}
}

func TestCivilYearOverflowIsRejectedBeforeNativeCall(t *testing.T) {
	tooLarge := int(^uint(0) >> 1)
	if _, err := CivilToJ2000Seconds(CivilDateTime{Year: tooLarge, Month: 1, Day: 1}); err == nil {
		t.Fatal("CivilToJ2000Seconds accepted a year outside C int32")
	} else {
		var statusErr *StatusError
		if errors.As(err, &statusErr) {
			t.Fatalf("CivilToJ2000Seconds crossed the C boundary: %v", err)
		}
	}
	if _, err := DayOfYear(tooLarge, 1, 1, 0, 0, 0); err == nil {
		t.Fatal("DayOfYear accepted a year outside C int32")
	}
	if _, err := DataDayOfYear(tooLarge, 1, 1); err == nil {
		t.Fatal("DataDayOfYear accepted a year outside C int32")
	}
	if _, err := LeapSeconds(tooLarge, 1, 1); err == nil {
		t.Fatal("LeapSeconds accepted a year outside C int32")
	}
}

func TestStatusErrorCarriesCDetail(t *testing.T) {
	_, err := CovarianceFromDiagonal([]float64{1, 2, 3, 4, 5})
	if err == nil {
		t.Fatal("expected invalid covariance error")
	}
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("error type = %T, want *StatusError", err)
	}
	if statusErr.Code != StatusInvalidArgument {
		t.Fatalf("status code = %d, want %d", statusErr.Code, StatusInvalidArgument)
	}
	if statusErr.Text == "" || statusErr.Detail == "" {
		t.Fatalf("status error did not preserve text/detail: %+v", statusErr)
	}
}

func TestCovarianceValuesDelegateToC(t *testing.T) {
	covariance, err := CovarianceFromDiagonal([]float64{1, 4, 9, 16, 25, 36})
	if err != nil {
		t.Fatal(err)
	}
	if covariance.Values[0][0] != 1 || covariance.Values[5][5] != 36 || covariance.Values[0][1] != 0 {
		t.Fatalf("unexpected covariance values: %#v", covariance.Values)
	}
	validation, err := covariance.Validate()
	if err != nil || !validation.Symmetric || !validation.PositiveSemidefinite {
		t.Fatalf("validation = %+v, %v", validation, err)
	}
	scaled, err := covariance.ToMeters()
	if err != nil || scaled.Values[5][5] != 36e6 {
		t.Fatalf("scaled covariance = %#v, %v", scaled.Values, err)
	}
	interpolated, err := covariance.Interpolate(Covariance6{Values: [6][6]float64{
		{4, 0, 0, 0, 0, 0},
		{0, 9, 0, 0, 0, 0},
		{0, 0, 16, 0, 0, 0},
		{0, 0, 0, 25, 0, 0},
		{0, 0, 0, 0, 36, 0},
		{0, 0, 0, 0, 0, 49},
	}}, 0.5)
	if err != nil || math.Abs(interpolated.Values[0][0]-2) > 1e-12 || math.Abs(interpolated.Values[5][5]-42) > 1e-12 {
		t.Fatalf("interpolated covariance = %#v, %v", interpolated.Values, err)
	}
}

func TestNMEAQueryCopiesVariableOutput(t *testing.T) {
	log, err := ParseNMEA([]byte(nmeaFixture))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := log.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	summary, err := log.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.SentenceCount != 2 || summary.EpochCount != 2 || summary.SkipCount != 0 || summary.WarningCount != 0 {
		t.Fatalf("unexpected NMEA summary: %+v", summary)
	}
	epochs, err := log.Epochs()
	if err != nil {
		t.Fatal(err)
	}
	if len(epochs) != 2 || !epochs[0].HasPosition || epochs[0].SentenceCount != 1 {
		t.Fatalf("unexpected NMEA epochs: %+v", epochs)
	}
	if math.Abs(epochs[0].LatitudeRad-48.1173*3.141592653589793/180) > 1e-12 ||
		math.Abs(epochs[0].LongitudeRad-11.516666666666667*3.141592653589793/180) > 1e-12 ||
		epochs[0].HeightM != 592.3 {
		t.Fatalf("unexpected first NMEA position: %+v", epochs[0])
	}
}

func TestNMEADoubleAndConcurrentClose(t *testing.T) {
	log, err := ParseNMEA([]byte(nmeaFixture))
	if err != nil {
		t.Fatal(err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := log.Summary(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Summary after Close = %v, want ErrClosed", err)
	}

	log, err = ParseNMEA([]byte(nmeaFixture))
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if closeErr := log.Close(); closeErr != nil {
				t.Errorf("concurrent NMEA.Close: %v", closeErr)
			}
		}()
	}
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, readErr := log.Summary()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent Summary: %v", readErr)
			}
		}()
	}
	group.Wait()
	if err := log.Close(); err != nil {
		t.Errorf("final NMEA.Close: %v", err)
	}
}
