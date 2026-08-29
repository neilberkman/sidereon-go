package sidereon

import (
	_ "embed"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// These fixtures are committed public examples. Embedding them keeps tests
// independent of the caller's working directory.
var (
	//go:embed testdata/pass_finder.json
	passFixtureData []byte
	//go:embed testdata/ccsds_example2.kvn
	cdmKVNFixtureData []byte
	//go:embed testdata/ccsds_example2.xml
	cdmXMLFixtureData []byte
)

const (
	fixtureISSLine1          = "1 25544U 98067A   18184.80969102  .00001614  00000-0  31745-4 0  9993"
	fixtureISSLine2          = "2 25544  51.6414 295.8524 0003435 262.6267 204.2868 15.54005638121106"
	fixtureTCAPrimaryLine1   = "1 25544U 98067A   20177.50000000  .00001264  00000-0  29621-4 0  9993"
	fixtureTCAPrimaryLine2   = "2 25544  51.6443 142.0099 0001234  90.0000 270.0000 15.49500000228000"
	fixtureTCASecondaryLine1 = "1 43205U 18015A   20177.50000000  .00000500  00000-0  20000-4 0  9990"
	fixtureTCASecondaryLine2 = "2 43205  51.6400 145.0000 0002000  80.0000 280.0000 15.50000000220000"
)

type passFixture struct {
	TLE struct {
		Line1 string `json:"line1"`
		Line2 string `json:"line2"`
	} `json:"tle"`
	Station struct {
		LatitudeDeg  float64 `json:"latitude_deg"`
		LongitudeDeg float64 `json:"longitude_deg"`
		AltitudeM    float64 `json:"altitude_m"`
	} `json:"station"`
	Window struct {
		StartUnixUS int64 `json:"start_unix_us"`
		EndUnixUS   int64 `json:"end_unix_us"`
	} `json:"window"`
	Options struct {
		ElevationMaskDeg     float64 `json:"elevation_mask_deg"`
		StepSeconds          float64 `json:"coarse_step_seconds"`
		TimeToleranceSeconds float64 `json:"time_tolerance_seconds"`
	} `json:"options"`
	Passes []struct {
		AOSUnixUS          int64   `json:"aos_unix_us"`
		LOSUnixUS          int64   `json:"los_unix_us"`
		CulminationUnixUS  int64   `json:"culmination_unix_us"`
		MaxElevationDeg    float64 `json:"max_elevation_deg"`
		MaxElevationDegHex string  `json:"max_elevation_deg_hex"`
	} `json:"passes"`
}

func fixTLEChecksum(line string) string {
	body := line[:68]
	total := 0
	for _, char := range body {
		if char >= '0' && char <= '9' {
			total += int(char - '0')
		}
		if char == '-' {
			total++
		}
	}
	return body + string(rune('0'+total%10))
}

func cleanupClose(t *testing.T, name string, close func() error) {
	t.Helper()
	t.Cleanup(func() {
		if err := close(); err != nil {
			t.Errorf("%s.Close() = %v", name, err)
		}
	})
}

func TestPassFixtureDeterministic(t *testing.T) {
	var fixture passFixture
	if err := json.Unmarshal(passFixtureData, &fixture); err != nil {
		t.Fatal(err)
	}
	tle, err := ParseTLE(fixture.TLE.Line1, fixture.TLE.Line2)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	station := PassStation{LatitudeDeg: fixture.Station.LatitudeDeg, LongitudeDeg: fixture.Station.LongitudeDeg, AltitudeM: fixture.Station.AltitudeM}
	passes, err := tle.FindPasses(station, time.UnixMicro(fixture.Window.StartUnixUS), time.UnixMicro(fixture.Window.EndUnixUS), &PassFinderOptions{ElevationMaskDeg: fixture.Options.ElevationMaskDeg, StepSeconds: fixture.Options.StepSeconds, TimeToleranceSeconds: fixture.Options.TimeToleranceSeconds})
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "PassList", passes.Close)
	count, err := passes.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != len(fixture.Passes) {
		t.Fatalf("pass count = %d, want %d", count, len(fixture.Passes))
	}
	values, err := passes.Values()
	if err != nil {
		t.Fatal(err)
	}
	for i, got := range values {
		want := fixture.Passes[i]
		bits, err := strconv.ParseUint(strings.TrimPrefix(want.MaxElevationDegHex, "0x"), 16, 64)
		if err != nil {
			t.Fatalf("pass %d max elevation bits %q: %v", i, want.MaxElevationDegHex, err)
		}
		wantMaxElevation := math.Float64frombits(bits)
		if got.AOS.UnixMicro() != want.AOSUnixUS || got.LOS.UnixMicro() != want.LOSUnixUS || got.Culmination.UnixMicro() != want.CulminationUnixUS || math.Abs(got.MaxElevationDeg-wantMaxElevation) > 1e-9 {
			t.Fatalf("pass %d = %+v, want %+v", i, got, want)
		}
		if got.DurationS <= 0 || got.AOS.UnixMicro() > got.Culmination.UnixMicro() || got.Culmination.UnixMicro() > got.LOS.UnixMicro() {
			t.Fatalf("invalid pass ordering: %+v", got)
		}
	}
	original := values[0]
	values[0].MaxElevationDeg = -1
	again, err := passes.Values()
	if err != nil {
		t.Fatal(err)
	}
	if again[0] != original {
		t.Fatalf("C result was not copied independently: got %+v, want %+v", again[0], original)
	}
	if err := passes.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := passes.Values(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Values after Close = %v, want ErrClosed", err)
	}
	if err := passes.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPassSmokeGroundTrackAndLookAngles(t *testing.T) {
	tle, err := ParseTLEWithOpsMode(fixtureISSLine1, fixtureISSLine2, OpsModeImproved)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	metadata, err := tle.Metadata()
	if err != nil || metadata.CatalogNumber != "25544" {
		t.Fatalf("metadata = %+v, %v", metadata, err)
	}
	epochs := []time.Time{time.UnixMicro(1530619200000000), time.UnixMicro(1530672973724542), time.UnixMicro(1530684498320405)}
	station := PassStation{LatitudeDeg: 51.5074, LongitudeDeg: -0.1278, AltitudeM: 80}
	look, err := tle.LookAngles(station, epochs)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "LookAngles", look.Close)
	lookValues, err := look.Values()
	if err != nil || len(lookValues) != len(epochs) {
		t.Fatalf("look angles = %+v, %v", lookValues, err)
	}
	for _, value := range lookValues {
		if !math.IsNaN(value.ElevationDeg) && !math.IsNaN(value.RangeKm) {
			break
		}
	}
	track, err := tle.GroundTrack(epochs)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "GroundTrack", track.Close)
	trackValues, err := track.Values()
	if err != nil || len(trackValues) != len(epochs) {
		t.Fatalf("ground track = %+v, %v", trackValues, err)
	}
	emptyTrack, err := tle.GroundTrack(nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "GroundTrack", emptyTrack.Close)
	emptyValues, err := emptyTrack.Values()
	if err != nil || len(emptyValues) != 0 {
		t.Fatalf("empty ground track = %+v, %v", emptyValues, err)
	}
	if err := tle.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tle.Metadata(); !errors.Is(err, ErrClosed) {
		t.Fatalf("TLE metadata after Close = %v, want ErrClosed", err)
	}
}

func TestPassConcurrentReadAndClose(t *testing.T) {
	tle, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	passes, err := tle.FindPasses(PassStation{LatitudeDeg: 51.5, LongitudeDeg: -0.1}, time.UnixMicro(1530619200000000), time.UnixMicro(1530705600000000), nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "PassList", passes.Close)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			_, readErr := passes.Values()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent Values: %v", readErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if closeErr := passes.Close(); closeErr != nil {
				t.Errorf("concurrent PassList.Close(): %v", closeErr)
			}
		}()
	}
	close(start)
	group.Wait()
	if err := passes.Close(); err != nil {
		t.Fatal(err)
	}
	if err := passes.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := passes.Values(); !errors.Is(err, ErrClosed) {
		t.Fatalf("PassList Values after Close = %v, want ErrClosed", err)
	}
}

func TestTLEConcurrentReadsAndClose(t *testing.T) {
	tle, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	epochs := []time.Time{time.UnixMicro(1530619200000000), time.UnixMicro(1530684498320405)}
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if _, readErr := tle.Metadata(); readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent TLE.Metadata: %v", readErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if _, readErr := tle.Lines(); readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent TLE.Lines: %v", readErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if _, readErr := tle.Propagate(epochs); readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent TLE.Propagate: %v", readErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if closeErr := tle.Close(); closeErr != nil {
				t.Errorf("concurrent TLE.Close: %v", closeErr)
			}
		}()
	}
	close(start)
	group.Wait()
	if err := tle.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tle.Metadata(); !errors.Is(err, ErrClosed) {
		t.Fatalf("TLE.Metadata after Close = %v, want ErrClosed", err)
	}
	if _, err := tle.Lines(); !errors.Is(err, ErrClosed) {
		t.Fatalf("TLE.Lines after Close = %v, want ErrClosed", err)
	}
	if _, err := tle.Propagate(epochs); !errors.Is(err, ErrClosed) {
		t.Fatalf("TLE.Propagate after Close = %v, want ErrClosed", err)
	}
}

func TestOrbitalAnalysisRejectsUnixMicrosecondOverflow(t *testing.T) {
	tle, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	overflow := time.Unix(9223372036854, 775808000).UTC()
	valid := time.Unix(0, 0).UTC()
	assertOverflow := func(name string, call func() error) {
		t.Helper()
		if err := call(); err == nil {
			t.Fatalf("%s accepted Unix-microsecond overflow", name)
		} else {
			var statusErr *StatusError
			if errors.As(err, &statusErr) {
				t.Fatalf("%s crossed the C boundary with %T: %v", name, err, err)
			}
		}
	}
	assertOverflow("Propagate", func() error {
		_, err := tle.Propagate([]time.Time{overflow})
		return err
	})
	for _, interval := range [][2]time.Time{{overflow, valid}, {valid, overflow}} {
		assertOverflow("FindPasses", func() error {
			_, err := tle.FindPasses(PassStation{}, interval[0], interval[1], nil)
			return err
		})
	}
	assertOverflow("GroundTrack", func() error {
		_, err := tle.GroundTrack([]time.Time{overflow})
		return err
	})
	assertOverflow("LookAngles", func() error {
		_, err := tle.LookAngles(PassStation{}, []time.Time{overflow})
		return err
	})
}

func TestGroundTrackConcurrentReadAndClose(t *testing.T) {
	tle, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	track, err := tle.GroundTrack([]time.Time{time.UnixMicro(1530619200000000), time.UnixMicro(1530684498320405)})
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "GroundTrack", track.Close)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			_, readErr := track.Values()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent GroundTrack.Values: %v", readErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if closeErr := track.Close(); closeErr != nil {
				t.Errorf("concurrent GroundTrack.Close(): %v", closeErr)
			}
		}()
	}
	close(start)
	group.Wait()
	if err := track.Close(); err != nil {
		t.Fatal(err)
	}
	if err := track.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := track.Values(); !errors.Is(err, ErrClosed) {
		t.Fatalf("GroundTrack Values after Close = %v, want ErrClosed", err)
	}
}

func TestLookAnglesConcurrentReadAndClose(t *testing.T) {
	tle, err := ParseTLE(fixtureISSLine1, fixtureISSLine2)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "TLE", tle.Close)
	look, err := tle.LookAngles(PassStation{LatitudeDeg: 51.5, LongitudeDeg: -0.1}, []time.Time{time.UnixMicro(1530619200000000), time.UnixMicro(1530684498320405)})
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "LookAngles", look.Close)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			_, readErr := look.Values()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent LookAngles.Values: %v", readErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if closeErr := look.Close(); closeErr != nil {
				t.Errorf("concurrent LookAngles.Close(): %v", closeErr)
			}
		}()
	}
	close(start)
	group.Wait()
	if err := look.Close(); err != nil {
		t.Fatal(err)
	}
	if err := look.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := look.Values(); !errors.Is(err, ErrClosed) {
		t.Fatalf("LookAngles Values after Close = %v, want ErrClosed", err)
	}
}

func TestOrbitDeterministicConversionsAndRelativeMotion(t *testing.T) {
	state := CartesianState{EpochTDBSeconds: 123456.75, PositionKm: [3]float64{7000, 0, 0}, VelocityKmPerS: [3]float64{0, 7.5, 1}}
	coe, err := CartesianToClassical(state, EarthMuKm3PerS2)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := ClassicalToCartesian(coe, EarthMuKm3PerS2)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.EpochTDBSeconds != 0 {
		t.Fatalf("epoch-free classical conversion epoch = %.17g, want zero", recovered.EpochTDBSeconds)
	}
	for i := 0; i < 3; i++ {
		if math.Abs(recovered.PositionKm[i]-state.PositionKm[i]) > 1e-8 || math.Abs(recovered.VelocityKmPerS[i]-state.VelocityKmPerS[i]) > 1e-11 {
			t.Fatalf("classical roundtrip = %+v, want %+v", recovered, state)
		}
	}
	eq, err := CartesianToEquinoctial(state, EarthMuKm3PerS2, RetrogradeFactorPrograde)
	if err != nil {
		t.Fatal(err)
	}
	fromEQ, err := EquinoctialToCartesian(eq, EarthMuKm3PerS2)
	if err != nil {
		t.Fatal(err)
	}
	mee, err := CartesianToModifiedEquinoctial(state, EarthMuKm3PerS2, RetrogradeFactorPrograde)
	if err != nil {
		t.Fatal(err)
	}
	fromMEE, err := ModifiedEquinoctialToCartesian(mee, EarthMuKm3PerS2)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if math.Abs(fromEQ.PositionKm[i]-state.PositionKm[i]) > 1e-7 || math.Abs(fromMEE.PositionKm[i]-state.PositionKm[i]) > 1e-7 {
			t.Fatalf("equinoctial roundtrip mismatch: eq=%+v mee=%+v", fromEQ, fromMEE)
		}
	}
	solved, err := SolveKepler(0.75, 0.1)
	if err != nil || solved.Iterations == 0 {
		t.Fatalf("Kepler = %+v, %v", solved, err)
	}
	eccentric, err := MeanToEccentricAnomaly(0.75, 0.1)
	if err != nil || math.Abs(eccentric-solved.AnomalyRad) > 1e-13 {
		t.Fatalf("Kepler mismatch = %.17g %.17g, %v", eccentric, solved.AnomalyRad, err)
	}
	stm, err := CWStateTransitionMatrix(0.001, 0)
	if err != nil || stm[0][0] != 1 || stm[5][5] != 1 {
		t.Fatalf("CW STM = %+v, %v", stm, err)
	}
	relative, err := RelativeState(state, CartesianState{PositionKm: [3]float64{7001, 2, 0}, VelocityKmPerS: [3]float64{0, 7.51, 1}})
	if err != nil {
		t.Fatal(err)
	}
	absolute, err := AbsoluteFromRelative(state, relative)
	if err != nil || math.Abs(absolute.PositionKm[0]-7001) > 1e-9 {
		t.Fatalf("relative roundtrip = %+v, %v", absolute, err)
	}
	if relative.EpochTDBSeconds != 0 || absolute.EpochTDBSeconds != state.EpochTDBSeconds {
		t.Fatalf("relative epoch propagation = relative %.17g, absolute %.17g, want relative zero and absolute %.17g", relative.EpochTDBSeconds, absolute.EpochTDBSeconds, state.EpochTDBSeconds)
	}
	departure, arrival, err := LambertBattin([3]float64{2.5 * 6378.1363, 0, 0}, [3]float64{1.9151111 * 6378.1363, 1.6069690 * 6378.1363, 0}, [3]float64{0, 4.999792554221911, 0}, LambertShortWay, LambertHighEnergy, 1, 92854.234)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(departure[0]-(-0.8696153795282852)) > 1e-12 || math.Abs(arrival[1]-5.41198791828363) > 1e-12 {
		t.Fatalf("Lambert = %v %v", departure, arrival)
	}
}

func TestConjunctionAndTCADeterministic(t *testing.T) {
	p1 := [3]float64{7000, 0, 0}
	v1 := [3]float64{0, 7.5, 0}
	p2 := [3]float64{7000.05, 0.02, 0}
	v2 := [3]float64{0, -7.5, 0.1}
	frame, err := BuildEncounterFrame(p1, v1, p2, v2)
	if err != nil || frame.MissKm <= 0 || frame.RelativeSpeedKmPerS <= 0 {
		t.Fatalf("encounter frame = %+v, %v", frame, err)
	}
	plane, err := EncounterPlaneCovariance(frame, [3][3]float64{{0.01, 0, 0}, {0, 0.01, 0}, {0, 0, 0.01}})
	if err != nil || plane[0][0] <= 0 || plane[1][1] <= 0 {
		t.Fatalf("encounter covariance = %+v, %v", plane, err)
	}
	pc, err := CollisionProbability(ConjunctionState{PositionKm: p1, VelocityKmPerS: v1, CovarianceKm2: [3][3]float64{{0.01, 0, 0}, {0, 0.01, 0}, {0, 0, 0.01}}}, ConjunctionState{PositionKm: p2, VelocityKmPerS: v2, CovarianceKm2: [3][3]float64{{0.01, 0, 0}, {0, 0.01, 0}, {0, 0, 0.01}}}, 0.02, PCMethodFosterEqualArea)
	if err != nil || pc.PC < 0 || pc.PC > 1 || pc.MissKm < 0 {
		t.Fatalf("Pc = %+v, %v", pc, err)
	}
	secondaryLine2 := "2 25544  51.6414 100.0000 0003435 010.0000 010.0000 15.54005638121106"
	// The public Python fixture uses this same orbit with a repaired checksum.
	secondaryLine2 = fixTLEChecksum(secondaryLine2)
	candidates, err := FindTCACandidates(fixtureISSLine1, fixtureISSLine2, fixtureISSLine1, secondaryLine2, JulianDate{Whole: 2458303}, JulianDate{Whole: 2458304}, nil)
	if err != nil || len(candidates) == 0 {
		t.Fatalf("TCA candidates = %d, %v", len(candidates), err)
	}
	conjunction, err := TCACollisionProbability(candidates[0], TCAPCOptions{HardBodyRadiusKm: 0.02, Method: PCMethodFosterEqualArea, UseDefaultCovariance: true})
	if err != nil || conjunction.Candidate.MissDistanceKm < 0 || conjunction.CollisionProbability.RelativeSpeedKmPerS <= 0 {
		t.Fatalf("TCA conjunction = %+v, %v", conjunction, err)
	}
	combined, err := FindTCAConjunctions(fixtureISSLine1, fixtureISSLine2, fixtureISSLine1, secondaryLine2, JulianDate{Whole: 2458303}, JulianDate{Whole: 2458304}, nil, TCAPCOptions{HardBodyRadiusKm: 0.02, Method: PCMethodFosterEqualArea, UseDefaultCovariance: true})
	if err != nil || len(combined) != len(candidates) {
		t.Fatalf("combined TCA conjunctions = %d, want %d, %v", len(combined), len(candidates), err)
	}
}

func TestOrbitalParityRoutesAndCatalogScreening(t *testing.T) {
	if beta, err := BetaAngleDeg([3]float64{0, 0, 1}, [3]float64{1, 0, 0}); err != nil || math.Abs(beta) > 1e-14 {
		t.Fatalf("BetaAngleDeg = %.17g, %v", beta, err)
	}
	if beta, err := BetaAngleFromStateDeg([3]float64{7000, 0, 0}, [3]float64{0, 7.5, 0}, [3]float64{1.5e8, 0, 0}); err != nil || math.Abs(beta) > 1e-14 {
		t.Fatalf("BetaAngleFromStateDeg = %.17g, %v", beta, err)
	}
	covariance, err := RTNToECICovariance([3][3]float64{{1, 0, 0}, {0, 2, 0}, {0, 0, 3}}, [3]float64{7000, 0, 0}, [3]float64{0, 7.5, 0})
	if err != nil {
		t.Fatal(err)
	}
	for row := range covariance {
		for column := range covariance[row] {
			if math.IsNaN(covariance[row][column]) || math.IsInf(covariance[row][column], 0) {
				t.Fatalf("RTNToECICovariance[%d][%d] = %v", row, column, covariance[row][column])
			}
		}
	}

	tcaOptions, err := DefaultTCAFinderOptions()
	if err != nil {
		t.Fatal(err)
	}
	start, end := JulianDate{Whole: 2459023}, JulianDate{Whole: 2459023, Fraction: 1}
	catalog := []TCATLEPair{
		{Line1: fixtureTCASecondaryLine1, Line2: fixtureTCASecondaryLine2},
		{Line1: fixtureTCASecondaryLine1, Line2: fixtureTCASecondaryLine2},
	}
	candidateHits, err := ScreenTCACandidatesFromTLECatalog(fixtureTCAPrimaryLine1, fixtureTCAPrimaryLine2, catalog, start, end, 1e9, &tcaOptions)
	if err != nil || len(candidateHits) == 0 {
		t.Fatalf("catalog candidates = %d, %v", len(candidateHits), err)
	}
	lastIndex := -1
	for _, hit := range candidateHits {
		if hit.SecondaryIndex < lastIndex || hit.SecondaryIndex < 0 || hit.SecondaryIndex >= len(catalog) {
			t.Fatalf("catalog candidate index/order = %d after %d", hit.SecondaryIndex, lastIndex)
		}
		lastIndex = hit.SecondaryIndex
	}
	if lastIndex != len(catalog)-1 {
		t.Fatalf("catalog candidate indices ended at %d, want %d", lastIndex, len(catalog)-1)
	}

	pcOptions := TCAPCOptions{HardBodyRadiusKm: 0.02, Method: PCMethodFosterEqualArea, UseDefaultCovariance: true}
	conjunctionHits, err := ScreenTCAConjunctionsFromTLECatalog(fixtureTCAPrimaryLine1, fixtureTCAPrimaryLine2, catalog, start, end, 1e9, &tcaOptions, pcOptions)
	if err != nil || len(conjunctionHits) != len(candidateHits) {
		t.Fatalf("catalog conjunctions = %d, candidates = %d, %v", len(conjunctionHits), len(candidateHits), err)
	}
	for i, hit := range conjunctionHits {
		if hit.SecondaryIndex != candidateHits[i].SecondaryIndex || !math.IsNaN(hit.Conjunction.CollisionProbability.PC) && hit.Conjunction.CollisionProbability.PC < 0 {
			t.Fatalf("catalog conjunction[%d] = %+v, candidate index = %d", i, hit, candidateHits[i].SecondaryIndex)
		}
	}

	var propagated [6][6]float64
	for i := range propagated {
		propagated[i][i] = 1e-2
	}
	propagatedOptions := TCAPropagatedCovariancePCOptions{
		HardBodyRadiusKm:     0.02,
		Method:               PCMethodFosterEqualArea,
		PrimaryCovariance0:   propagated,
		SecondaryCovariance0: propagated,
		ForceModel:           PropagationForceModelTwoBodyJ2,
		Integrator:           PropagationIntegratorDP54,
		AbsTol:               1e-9,
		RelTol:               1e-9,
		InitialStepSeconds:   1,
		MinStepSeconds:       1e-3,
		MaxStepSeconds:       60,
		MaxSteps:             100000,
	}
	propagatedConjunctions, err := FindTCAConjunctionsWithPropagatedCovarianceFromTLEs(fixtureTCAPrimaryLine1, fixtureTCAPrimaryLine2, fixtureTCASecondaryLine1, fixtureTCASecondaryLine2, start, end, &tcaOptions, propagatedOptions)
	if err != nil || len(propagatedConjunctions) == 0 {
		t.Fatalf("propagated-covariance conjunctions = %d, %v", len(propagatedConjunctions), err)
	}
	for _, conjunction := range propagatedConjunctions {
		if conjunction.Candidate.MissDistanceKm < 0 || math.IsNaN(conjunction.CollisionProbability.PC) || math.IsInf(conjunction.CollisionProbability.PC, 0) {
			t.Fatalf("invalid propagated conjunction = %+v", conjunction)
		}
	}

	ownedLine := fixtureTCASecondaryLine1
	ownedCatalog := []TCATLEPair{{Line1: ownedLine, Line2: fixtureTCASecondaryLine2}}
	if _, err := ScreenTCACandidatesFromTLECatalog(fixtureTCAPrimaryLine1, fixtureTCAPrimaryLine2, ownedCatalog, start, end, 1e9, &tcaOptions); err != nil {
		t.Fatalf("catalog input ownership = %v", err)
	}
	ownedLine = fixtureISSLine1
	ownedCatalog[0].Line1 = ownedLine
	ownedCatalog[0].Line2 = fixtureISSLine2
	if _, err := ScreenTCACandidatesFromTLECatalog(fixtureTCAPrimaryLine1, fixtureTCAPrimaryLine2, ownedCatalog, start, end, 1e9, &tcaOptions); err != nil {
		t.Fatalf("catalog retained caller input = %v", err)
	}
	ownedCatalog[0].Line1 = fixtureTCASecondaryLine1 + "\x00"
	if _, err := ScreenTCACandidatesFromTLECatalog(fixtureTCAPrimaryLine1, fixtureTCAPrimaryLine2, ownedCatalog, start, end, 1e9, &tcaOptions); err == nil {
		t.Fatal("catalog accepted an embedded NUL")
	}
}

func TestPublicCleanupErrorComposition(t *testing.T) {
	copyErr := errors.New("copy failed")
	closeErr := errors.New("close failed")
	joined := joinPublicErrors(copyErr, closeErr)
	if !errors.Is(joined, copyErr) || !errors.Is(joined, closeErr) {
		t.Fatalf("joined cleanup errors = %v", joined)
	}
}

func TestCDMFixtureDeterministicAndCopyIndependent(t *testing.T) {
	cdm, err := ParseCDMKVN(cdmKVNFixtureData)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "CDM", cdm.Close)
	numbers, err := cdm.Numbers()
	if err != nil {
		t.Fatal(err)
	}
	if !numbers.MissDistanceM.Present || numbers.MissDistanceM.Value != 715 || !numbers.RelativeSpeedMPerS.Present || numbers.RelativeSpeedMPerS.Value != 14762 || !numbers.CollisionProbability.Present {
		t.Fatalf("CDM numbers = %+v", numbers)
	}
	if numbers.HardBodyRadiusM.Present || !math.IsNaN(numbers.HardBodyRadiusM.Value) {
		t.Fatalf("absent hard-body radius = %+v", numbers.HardBodyRadiusM)
	}
	object, err := cdm.Object(1)
	if err != nil {
		t.Fatal(err)
	}
	if object.ObjectName != "SATELLITE A" || object.RefFrame != "EME2000" || object.PositionKm[0] != 2570.097065 || object.CovarianceRTN[0] != 41.42 || !object.HasVelocityCovariance {
		t.Fatalf("CDM object = %+v", object)
	}
	wantCovariance := [6]float64{41.42, -8.579, 2533, -23.13, 13.36, 70.98}
	wantVelocityCovariance := [15]float64{2.520e-3, -5.476, 8.626e-4, 5.744e-3, -1.006e-2, 4.041e-3, -1.359e-3, -1.502e-5, 1.049e-5, 1.053e-3, -3.412e-3, 1.213e-2, -3.004e-6, -1.091e-6, 5.529e-5}
	if object.CovarianceRTN != wantCovariance || object.VelocityCovarianceRTN != wantVelocityCovariance {
		t.Fatalf("CDM covariance = %v / %v, want %v / %v", object.CovarianceRTN, object.VelocityCovarianceRTN, wantCovariance, wantVelocityCovariance)
	}
	object2, err := cdm.Object(2)
	if err != nil {
		t.Fatal(err)
	}
	if object2.CovarianceRTN != [6]float64{1337, -48060, 2492000, -32.98, -758.88, 71.05} || !object2.HasVelocityCovariance || object2.VelocityCovarianceRTN[14] != 5.178e-5 {
		t.Fatalf("CDM object 2 covariance = %+v, want full public covariance", object2)
	}
	object.PositionKm[0] = -1
	again, err := cdm.Object(1)
	if err != nil || again.PositionKm[0] != 2570.097065 {
		t.Fatalf("CDM object copy = %+v, %v", again, err)
	}
	encoded, err := cdm.ToKVN()
	if err != nil || len(encoded) == 0 {
		t.Fatalf("CDM KVN = %d, %v", len(encoded), err)
	}
	encodedCopy := append([]byte(nil), encoded...)
	encodedCopy[0] = 'x'
	encodedAgain, err := cdm.ToKVN()
	if err != nil || string(encodedAgain) != string(encoded) {
		t.Fatal("CDM serialization was not stable after copying")
	}
	reparsed, err := ParseCDMKVN(encoded)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "CDM", reparsed.Close)
	if got, _, err := reparsed.StringField(CDMTCA); err != nil || got != "2010-03-13T22:37:52.618" {
		t.Fatalf("TCA string = %q, %v", got, err)
	}
	fromPublicXML, err := ParseCDMXML(cdmXMLFixtureData)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "CDM", fromPublicXML.Close)
	if got, _, err := fromPublicXML.StringField(CDMOriginator); err != nil || got != "JSPOC" {
		t.Fatalf("public XML originator = %q, %v", got, err)
	}
	xmlObject, err := fromPublicXML.Object(1)
	if err != nil {
		t.Fatal(err)
	}
	if xmlObject.CovarianceRTN != wantCovariance {
		t.Fatalf("public XML covariance = %+v, want public covariance content", xmlObject)
	}
	xmlObject2, err := fromPublicXML.Object(2)
	if err != nil {
		t.Fatal(err)
	}
	if xmlObject2.CovarianceRTN != [6]float64{1337, -48060, 2492000, -32.98, -758.88, 71.05} || xmlObject2.HasVelocityCovariance {
		t.Fatalf("public XML object 2 covariance = %+v, want its complete position covariance", xmlObject2)
	}
	xml, err := cdm.ToXML()
	if err != nil || len(xml) == 0 {
		t.Fatalf("CDM XML = %d, %v", len(xml), err)
	}
	publicEncodedXML, err := fromPublicXML.ToXML()
	if err != nil || len(publicEncodedXML) == 0 {
		t.Fatalf("public CDM XML serialization failed: got=%d err=%v", len(publicEncodedXML), err)
	}
	fromXML, err := ParseCDMXML(xml)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "CDM", fromXML.Close)
	if got, _, err := fromXML.ObjectStringField(2, CDMObjectName); err != nil || got != "FENGYUN 1C DEB" {
		t.Fatalf("object name = %q, %v", got, err)
	}
	if err := cdm.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cdm.ToKVN(); !errors.Is(err, ErrClosed) {
		t.Fatalf("CDM ToKVN after Close = %v", err)
	}
	if err := cdm.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestMalformedAnalysisInputsHaveTypedCDetail(t *testing.T) {
	if _, err := ParseTLE("bad", "bad"); err == nil {
		t.Fatal("bad TLE unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Detail == "" {
			t.Fatalf("bad TLE error = %T %v", err, err)
		}
	}
	if _, err := ParseCDMKVN([]byte("OBJECT = OBJECT1\nX = 1 [km]\n")); err == nil {
		t.Fatal("bad CDM unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Detail == "" {
			t.Fatalf("bad CDM error = %T %v", err, err)
		}
	}
	if _, err := CartesianToClassical(CartesianState{}, EarthMuKm3PerS2); err == nil {
		t.Fatal("degenerate state unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Detail == "" {
			t.Fatalf("degenerate state error = %T %v", err, err)
		}
	}
	if _, err := RelativeRotation(RelativeFrame(99), CartesianState{}); err == nil {
		t.Fatal("invalid relative-frame selector unexpectedly succeeded")
	}
	if _, err := CollisionProbability(ConjunctionState{}, ConjunctionState{}, 0.02, PCMethod(99)); err == nil {
		t.Fatal("invalid PC selector unexpectedly succeeded")
	}
	if _, _, err := LambertBattin([3]float64{}, [3]float64{}, [3]float64{}, LambertDirection(99), LambertLowEnergy, 0, 1); err == nil {
		t.Fatal("invalid Lambert direction unexpectedly succeeded")
	}
	if _, _, err := LambertBattin([3]float64{}, [3]float64{}, [3]float64{}, LambertShortWay, LambertEnergy(99), 0, 1); err == nil {
		t.Fatal("invalid Lambert energy unexpectedly succeeded")
	}
	cdm, err := ParseCDMKVN(cdmKVNFixtureData)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "CDM", cdm.Close)
	if _, _, err := cdm.StringField(CDMStringField(99)); err == nil {
		t.Fatal("invalid CDM selector unexpectedly succeeded")
	}
	if _, err := cdm.Object(0); err == nil {
		t.Fatal("invalid CDM object index unexpectedly succeeded")
	}
	if _, _, err := cdm.ObjectStringField(3, CDMObjectName); err == nil {
		t.Fatal("invalid CDM object index unexpectedly succeeded")
	}
}

func TestCDMConcurrentReadAndClose(t *testing.T) {
	cdm, err := ParseCDMKVN(cdmKVNFixtureData)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "CDM", cdm.Close)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			_, readErr := cdm.Numbers()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent CDM.Numbers: %v", readErr)
			}
		}()
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			if closeErr := cdm.Close(); closeErr != nil {
				t.Errorf("concurrent CDM.Close(): %v", closeErr)
			}
		}()
	}
	close(start)
	group.Wait()
	if err := cdm.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cdm.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := cdm.ToKVN(); !errors.Is(err, ErrClosed) {
		t.Fatalf("CDM ToKVN after Close = %v, want ErrClosed", err)
	}
}
