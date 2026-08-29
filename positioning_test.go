package sidereon

import (
	"bytes"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func readPositioningFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func oneEpochSPPConfig() SPPConfig {
	return SPPConfig{
		Observations: []SPPObservation{
			{SatelliteID: "G08", PseudorangeM: 23825519.8},
			{SatelliteID: "G10", PseudorangeM: 22717690.1},
			{SatelliteID: "G16", PseudorangeM: 20478653.4},
			{SatelliteID: "G18", PseudorangeM: 21768335.2},
			{SatelliteID: "G20", PseudorangeM: 21248327.7},
			{SatelliteID: "G21", PseudorangeM: 20808709.8},
		},
		TRxJ2000S:       646272000.0,
		TRxSecondOfDayS: 43200.0,
		DayOfYear:       176.5,
		InitialGuess:    [4]float64{4.5e6, 0.5e6, 4.5e6, 0},
		WithGeodetic:    true,
	}
}

func publicSPPConfig() SPPConfig {
	config := oneEpochSPPConfig()
	ids := []string{
		"G01", "G02", "G03", "G05", "G06", "G07", "G08", "G09", "G10", "G11",
		"G12", "G13", "G14", "G15", "G16", "G17", "G18", "G19", "G20", "G21",
		"G22", "G24", "G25", "G26", "G27", "G28", "G29", "G30", "G31", "G32",
	}
	values := []uint64{
		0x417b004ff10ef0a2, 0x417dc8eb4a61a3c6, 0x417d7794e8d4f542, 0x417a86abdf14b1ed,
		0x417f5a58cab632f7, 0x417856f3fdaa90ab, 0x4176b8c6fd82e861, 0x417ac4d73d52338c,
		0x4175aa4fa1a0c21f, 0x417912ee07a3e2f7, 0x417d71f24cff269c, 0x4178cd0483a92706,
		0x417ac2f98647e0ea, 0x417816e92e0128ad, 0x417387abd6052c3b, 0x417e9e575c4eb87b,
		0x4174c288f3bd1166, 0x417f9cc78862439f, 0x417443947bd00bd6, 0x4173d8405cd09f84,
		0x417c1b70fd3eab2c, 0x417d0ddeba10a71c, 0x417b4abb86be2ed5, 0x417425d51967e798,
		0x41745a4b78a81707, 0x417d5146915dd476, 0x41787ae825e18d44, 0x4179eaeb14a44f8f,
		0x4179031a36a10fbc, 0x417a6e35835d6753,
	}
	config.Observations = make([]SPPObservation, len(ids))
	for i := range ids {
		config.Observations[i] = SPPObservation{SatelliteID: ids[i], PseudorangeM: math.Float64frombits(values[i])}
	}
	return config
}

func usedSPPConfig() SPPConfig {
	config := publicSPPConfig()
	observations := config.Observations
	used := []int{6, 8, 14, 16, 18, 19, 23, 24}
	config.Observations = make([]SPPObservation, len(used))
	for i, index := range used {
		config.Observations[i] = observations[index]
	}
	return config
}

func TestDeterministicSPPFixture(t *testing.T) {
	// This compact SP3 is an excerpt of the public GRG fixture used by
	// bindings/c/tests/spp_fixture.h at the pinned sidereon-c revision.
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Errorf("SP3.Close() = %v", err)
		}
	})

	solution, err := SolveSPP(sp3, usedSPPConfig())
	if err != nil {
		t.Fatal(err)
	}
	wantPosition := [3]uint64{0x41511b07ff83c7f1, 0x4120cd6b5ee8cafe, 0x41511e62229db724}
	for axis := range wantPosition {
		if math.Float64bits(solution.PositionM[axis]) != wantPosition[axis] {
			t.Fatalf("position[%d] = %.17g, want bits %#x", axis, solution.PositionM[axis], wantPosition[axis])
		}
	}
	if math.Float64bits(solution.ReceiverClockS) != 0x3f1a3b88360a8d78 {
		t.Fatalf("receiver clock = %.17g, want frozen C value", solution.ReceiverClockS)
	}
	wantIDs := []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}
	if solution.UsedSatelliteCount != len(wantIDs) || len(solution.UsedSatelliteIDs) != len(wantIDs) || len(solution.ResidualsM) != len(wantIDs) {
		t.Fatalf("unexpected solution sizes: %+v", solution)
	}
	for i, id := range wantIDs {
		if solution.UsedSatelliteIDs[i] != id {
			t.Fatalf("used satellite[%d] = %q, want %q", i, solution.UsedSatelliteIDs[i], id)
		}
	}
	wantResiduals := []uint64{
		0xbe95400000000000, 0xbf46fc0000000000, 0xbf1c068000000000, 0x3f378df000000000,
		0xbf1deac000000000, 0xbf24d52000000000, 0x3f43165000000000, 0xbf00f98000000000,
	}
	for i, bits := range wantResiduals {
		if math.Float64bits(solution.ResidualsM[i]) != bits {
			t.Fatalf("residual[%d] = %.17g, want bits %#x", i, solution.ResidualsM[i], bits)
		}
	}
	if solution.DOP == nil || solution.Geodetic == nil || !solution.Metadata.Converged {
		t.Fatalf("missing C solution outputs: %+v", solution)
	}
}

func TestLegacySPPExtendedAtmosphereFixture(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	config := usedSPPConfig()
	config.Ionosphere = true
	config.Troposphere = true
	config.KlobucharAlpha = [4]float64{1e-8, 1e-8, -1e-7, 0}
	config.KlobucharBeta = [4]float64{90000, 0, -100000, 0}
	config.PressureHPA = 900
	config.TemperatureK = 280
	config.RelativeHumidity = 0.2
	solution, err := SolveSPP(sp3, config)
	if err != nil {
		t.Fatal(err)
	}
	wantPosition := [3]uint64{0x41511b067925e57d, 0x4120cd6993456913, 0x41511e607e84895e}
	for axis, bits := range wantPosition {
		if math.Float64bits(solution.PositionM[axis]) != bits {
			t.Fatalf("extended position[%d] bits = %#x, want %#x", axis, math.Float64bits(solution.PositionM[axis]), bits)
		}
	}
	if math.Float64bits(solution.ReceiverClockS) != 0x3f1a38751a0c0e58 {
		t.Fatalf("extended receiver clock bits = %#x", math.Float64bits(solution.ReceiverClockS))
	}
	if !solution.Metadata.IonosphereApplied || !solution.Metadata.TroposphereApplied || solution.Metadata.Iterations != 11 || solution.Metadata.UsedCount != 8 {
		t.Fatalf("extended metadata = %+v", solution.Metadata)
	}
}

func TestSPPErrorPreservesTypedDetail(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Errorf("SP3.Close() = %v", err)
		}
	})

	_, err = SolveSPP(sp3, SPPConfig{})
	if err == nil {
		t.Fatal("SolveSPP unexpectedly succeeded")
	}
	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("SolveSPP error = %T, want *StatusError", err)
	}
	if statusErr.Code != StatusSolve || statusErr.Detail == "" {
		t.Fatalf("SolveSPP status = %+v", statusErr)
	}
}

func TestSP3FixtureQueriesCopyAndOwnInput(t *testing.T) {
	data := readPositioningFixture(t, "trimmed.sp3")
	original := append([]byte(nil), data...)
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	for i := range data {
		data[i] = 'x'
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Errorf("SP3.Close() = %v", err)
		}
	})

	count, err := sp3.EpochCount()
	if err != nil || count != 5 {
		t.Fatalf("EpochCount = %d, %v", count, err)
	}
	epochs, err := sp3.Epochs()
	if err != nil || len(epochs) != 5 {
		t.Fatalf("Epochs = %#v, %v", epochs, err)
	}
	satellites, err := sp3.Satellites()
	if err != nil || len(satellites) != 8 || satellites[0] != "G08" {
		t.Fatalf("Satellites = %#v, %v", satellites, err)
	}
	state, err := sp3.State("G08", 0)
	if err != nil || !state.HasClock || state.PositionM == [3]float64{} {
		t.Fatalf("State = %#v, %v", state, err)
	}
	summary, err := sp3.PredictionSummary()
	if err != nil || summary.EpochCount != 5 {
		t.Fatalf("PredictionSummary = %#v, %v", summary, err)
	}
	if !bytes.Equal(original, readPositioningFixture(t, "trimmed.sp3")) {
		t.Fatal("fixture unexpectedly changed")
	}
}

func TestInvalidSP3AndEmptyInputPreserveTypedDetail(t *testing.T) {
	for _, input := range [][]byte{nil, {}, []byte("not an SP3")} {
		_, err := LoadSP3(input)
		if err == nil {
			t.Fatalf("LoadSP3(%q) unexpectedly succeeded", input)
		}
		var statusErr *StatusError
		if !errors.As(err, &statusErr) {
			t.Fatalf("LoadSP3(%q) error = %T, want *StatusError", input, err)
		}
		if statusErr.Code != StatusSP3Parse || statusErr.Text == "" || statusErr.Detail == "" {
			t.Fatalf("LoadSP3(%q) status = %+v", input, statusErr)
		}
	}
}

func TestSP3DoubleConcurrentCloseAndUseAfterClose(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	if err := sp3.Close(); err != nil {
		t.Fatal(err)
	}
	if err := sp3.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := sp3.EpochCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("EpochCount after Close = %v, want ErrClosed", err)
	}
	if _, err := sp3.State("G08", 0); !errors.Is(err, ErrClosed) {
		t.Fatalf("State after Close = %v, want ErrClosed", err)
	}

	sp3, err = LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if closeErr := sp3.Close(); closeErr != nil {
				t.Errorf("concurrent SP3.Close: %v", closeErr)
			}
		}()
	}
	for i := 0; i < 16; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, readErr := sp3.Epochs()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent Epochs: %v", readErr)
			}
		}()
	}
	group.Wait()
	if err := sp3.Close(); err != nil {
		t.Errorf("final SP3.Close: %v", err)
	}
}

func TestPositioningRejectsEmbeddedNULWithoutTruncating(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sp3.Close(); err != nil {
			t.Errorf("SP3.Close() = %v", err)
		}
	})
	if _, err := sp3.State("G08\x00suffix", 0); err == nil {
		t.Fatal("SP3.State accepted an embedded NUL and a truncated satellite ID")
	}

	config := usedSPPConfig()
	config.Observations[0].SatelliteID = "G08\x00suffix"
	if _, err := SolveSPP(sp3, config); err == nil {
		t.Fatal("SolveSPP accepted an embedded NUL and a truncated satellite ID")
	}

	line1, line2, ok := splitTLELines(string(readPositioningFixture(t, "iss.tle")))
	if !ok {
		t.Fatal("invalid TLE fixture")
	}
	for name, lines := range map[string][2]string{
		"line 1": {line1 + "\x00suffix", line2},
		"line 2": {line1, line2 + "\x00suffix"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseTLE(lines[0], lines[1]); err == nil {
				t.Fatalf("ParseTLE accepted an embedded NUL and a truncated %s", name)
			}
		})
	}
}

func TestTLEFixtureDeterministicPropagationAndMetadata(t *testing.T) {
	lines := string(readPositioningFixture(t, "iss.tle"))
	line1, line2, ok := splitTLELines(lines)
	if !ok {
		t.Fatal("invalid TLE fixture")
	}
	timestamps := []int64{
		1530646200000000,
		1530646260000000,
		1530646320000000,
		1530646380000000,
		1530646440000000,
		1530646500000000,
		1530646560000000,
		1530646620000000,
		1530646680000000,
		1530646740000000,
	}
	times := make([]time.Time, len(timestamps))
	for i, timestamp := range timestamps {
		times[i] = time.UnixMicro(timestamp).UTC()
	}

	tle, err := ParseTLE(line1, line2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := tle.Close(); err != nil {
			t.Errorf("TLE.Close() = %v", err)
		}
	})
	metadata, err := tle.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	if metadata.CatalogNumber != "25544" || metadata.EpochYear != 2018 || metadata.RevolutionNumber != 12110 {
		t.Fatalf("metadata = %+v", metadata)
	}
	encoded, err := tle.Lines()
	if err != nil || encoded.Line1 != line1 || encoded.Line2 != line2 {
		t.Fatalf("Lines = %+v, %v", encoded, err)
	}

	states, err := tle.Propagate(times)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != len(times) {
		t.Fatalf("Propagate returned %d states, want %d", len(states), len(times))
	}
	// Values are copied from the public C binding's prop_fixture.h, generated
	// from its committed propagation fixture at the pinned C revision.
	wantPosition := [3]uint64{0x4098ea1be4cb4974, 0x40b2e5565b1d73e0, 0x40b17a14ef3fa337}
	wantVelocity := [3]uint64{0xc0147e8d3aa3fa34, 0x4012c73c3e761c93, 0xc009f3378fdc48e0}
	for axis := 0; axis < 3; axis++ {
		if math.Float64bits(states[0].PositionKm[axis]) != wantPosition[axis] {
			t.Fatalf("first position[%d] = %.17g, bits %#x", axis, states[0].PositionKm[axis], math.Float64bits(states[0].PositionKm[axis]))
		}
		if math.Float64bits(states[0].VelocityKmPerS[axis]) != wantVelocity[axis] {
			t.Fatalf("first velocity[%d] = %.17g, bits %#x", axis, states[0].VelocityKmPerS[axis], math.Float64bits(states[0].VelocityKmPerS[axis]))
		}
	}
	if math.IsNaN(states[0].EpochJ2000S) || math.IsInf(states[0].EpochJ2000S, 0) || states[0].EpochJ2000S == 0 {
		t.Fatalf("EpochJ2000S = %.17g", states[0].EpochJ2000S)
	}
	if empty, err := tle.Propagate(nil); err != nil || len(empty) != 0 {
		t.Fatalf("empty Propagate = %#v, %v", empty, err)
	}
}

func TestTLEDoubleConcurrentCloseAndUseAfterClose(t *testing.T) {
	tokens := readPositioningFixture(t, "iss.tle")
	line1, line2, ok := splitTLELines(string(tokens))
	if !ok {
		t.Fatal("invalid TLE fixture")
	}
	tle, err := ParseTLE(line1, line2)
	if err != nil {
		t.Fatal(err)
	}
	if err := tle.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tle.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tle.Metadata(); !errors.Is(err, ErrClosed) {
		t.Fatalf("Metadata after Close = %v, want ErrClosed", err)
	}

	tle, err = ParseTLE(line1, line2)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if closeErr := tle.Close(); closeErr != nil {
				t.Errorf("concurrent TLE.Close: %v", closeErr)
			}
		}()
	}
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, readErr := tle.Metadata()
			if readErr != nil && !errors.Is(readErr, ErrClosed) {
				t.Errorf("concurrent Metadata: %v", readErr)
			}
		}()
	}
	group.Wait()
	if err := tle.Close(); err != nil {
		t.Errorf("final TLE.Close: %v", err)
	}
}

func splitTLELines(value string) (string, string, bool) {
	first := ""
	second := ""
	for _, line := range bytes.Split([]byte(value), []byte{'\n'}) {
		line = bytes.TrimSuffix(line, []byte{'\r'})
		if len(line) == 0 {
			continue
		}
		if first == "" {
			first = string(line)
		} else if second == "" {
			second = string(line)
		}
	}
	return first, second, first != "" && second != ""
}
