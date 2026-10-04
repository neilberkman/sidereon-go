package sidereon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRINEXObservationHeaderDetailsPublicProjection(t *testing.T) {
	cases := []struct {
		name     string
		wantHash string
	}{
		{"ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx", "f88ec090270705e874c66fd060d5ce62a7caca3855706e240a3f3062e9d93fa8"},
		{"rinex211_table_a7_example.rnx", "258249313fc2b682bb812d21bdeb782c10315e51c7f0f61f7b22a8d769e7a144"},
		{"header_details_wtzr.rnx", "5079b75f0d3ba734d550247f5ea007b79b68c53730a93c0dff0e95cf584ca175"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "obs", test.name))
			if err != nil {
				t.Fatal(err)
			}
			obs, err := ParseRINEXObservation(data)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := obs.Close(); err != nil {
					t.Error(err)
				}
			}()
			segments, err := obs.HeaderDetails()
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(segments)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(encoded)
			digest := hex.EncodeToString(sum[:])
			if digest != test.wantHash {
				t.Fatalf("header details digest = %s, want %s", digest, test.wantHash)
			}

			switch test.name {
			case "ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx":
				if len(segments) != 1 || segments[0].FirstEpochIndex != 0 {
					t.Fatalf("segments = %+v", segments)
				}
				h := segments[0].Header
				if h.Version != 3.05 || h.ApproxPositionM == nil || *h.ApproxPositionM != [3]RINEXHeaderFloat64{3582105.291, 532589.7313, 5232754.8054} ||
					h.AntennaDeltaHENM == nil || *h.AntennaDeltaHENM != [3]RINEXHeaderFloat64{0.216, 0, 0} {
					t.Fatalf("position/version details = %+v", h)
				}
				if h.ProgramRunByDate == nil || *h.ProgramRunByDate != (RINEXProgramRunByDate{"sbf2rin-13.4.5", "", "20220706 130812 UTC"}) ||
					!reflect.DeepEqual(h.Comments, []string{"gfzrnx-1.16-8177    FILE MERGE          20220706 132211 UTC", "INITIAL_RINEX_VERSION: 3.04", "SEPTENTRIO RECEIVERS OUTPUT ALIGNED CARRIER PHASES.", "NO FURTHER PHASE SHIFT APPLIED IN THE RINEX ENCODER.", "GFZRNX.NUM_EPOCHS: 0"}) {
					t.Fatalf("program/comments = %+v, %+v", h.ProgramRunByDate, h.Comments)
				}
				if h.MarkerName == nil || *h.MarkerName != "ESBC00DNK" || h.MarkerNumber == nil || *h.MarkerNumber != "10118M001" ||
					h.MarkerType == nil || *h.MarkerType != "GEODETIC" || h.Observer == nil || *h.Observer != "SDFE" || h.Agency == nil || *h.Agency != "SDFE" {
					t.Fatalf("marker details = %+v", h)
				}
				if h.Receiver == nil || *h.Receiver != (RINEXReceiverInfo{"3047937", "SEPT POLARX5", "5.2.0"}) ||
					h.Antenna == nil || *h.Antenna != (RINEXAntennaInfo{"CR5200327016", "ASH701945E_M    SCIS"}) {
					t.Fatalf("receiver/antenna = %+v / %+v", h.Receiver, h.Antenna)
				}
				if h.IntervalS == nil || *h.IntervalS != 30 || h.TimeOfFirstObs == nil ||
					*h.TimeOfFirstObs != (RINEXHeaderEpochWithScale{RINEXHeaderEpoch{2020, 6, 25, 0, 0, 0}, "GPST"}) ||
					h.TimeOfLastObs == nil || *h.TimeOfLastObs != (RINEXHeaderEpochWithScale{RINEXHeaderEpoch{2020, 6, 25, 23, 59, 30}, "GPST"}) {
					t.Fatalf("interval/epochs = %+v, %+v, %+v", h.IntervalS, h.TimeOfFirstObs, h.TimeOfLastObs)
				}
				if h.NSatellites == nil || *h.NSatellites != 0 || len(h.PRNObsCounts) != 0 || len(h.RINEX2Types) != 0 ||
					h.RINEX2System != nil || len(h.PhaseShifts) != 22 || len(h.ScaleFactors) != 0 ||
					h.GLONASSCodePhaseEntries != nil || h.SignalStrengthUnit == nil || *h.SignalStrengthUnit != "DBHZ" ||
					h.LeapSeconds != nil || h.UnretainedHeaderLabels == nil || len(h.UnretainedHeaderLabels) != 0 {
					t.Fatalf("optional/count/list details = %+v", h)
				}
				if got := h.ObsCodes["GPS"]; !reflect.DeepEqual(got, []string{"C1C", "C1W", "C2L", "C2W", "C5Q", "D1C", "D2L", "D2W", "D5Q", "L1C", "L2L", "L2W", "L5Q", "S1C", "S1W", "S2L", "S2W", "S5Q"}) ||
					!reflect.DeepEqual(h.ObsCodes, h.DeclaredObsCodes) {
					t.Fatalf("observation code maps = %+v", h.ObsCodes)
				}
				if got := h.PhaseShifts[0]; got.System != "BeiDou" || got.Code == nil || *got.Code != "L2I" || got.CorrectionCycles != nil ||
					len(got.Satellites) != 0 || len(got.UnrepresentableSatellites) != 0 {
					t.Fatalf("phase shift with absent correction = %+v", got)
				}
				if got := h.PhaseShifts[3]; got.CorrectionCycles == nil || *got.CorrectionCycles != 0 {
					t.Fatalf("explicit zero phase shift = %+v", got)
				}
				if h.GLONASSSlots["1"] != 1 || h.GLONASSSlots["10"] != -7 || h.GLONASSSlots["24"] != 2 || len(h.GLONASSSlots) != 23 {
					t.Fatalf("GLONASS slots = %+v", h.GLONASSSlots)
				}
			case "rinex211_table_a7_example.rnx":
				if len(segments) != 2 || segments[0].FirstEpochIndex != 0 || segments[1].FirstEpochIndex != 5 {
					t.Fatalf("event header segment indexes = %+v", segments)
				}
				h := segments[0].Header
				if h.Version != 2.11 || !reflect.DeepEqual(h.RINEX2Types, []string{"P1", "L1", "L2", "P2", "L5"}) ||
					h.RINEX2System != nil || h.MarkerType != nil || h.TimeOfLastObs != nil || h.NSatellites != nil ||
					h.SignalStrengthUnit != nil || h.LeapSeconds != nil || h.GLONASSCodePhaseEntries != nil {
					t.Fatalf("RINEX 2 optional/type details = %+v", h)
				}
				if !reflect.DeepEqual(h.UnretainedHeaderLabels, []string{"WAVELENGTH FACT L1/2", "WAVELENGTH FACT L1/2", "RCV CLOCK OFFS APPL"}) ||
					h.MarkerName == nil || *h.MarkerName != "A 9080" || h.IntervalS == nil || *h.IntervalS != 18 {
					t.Fatalf("RINEX 2 unretained/marker/interval = %+v", h)
				}
			case "header_details_wtzr.rnx":
				h := segments[0].Header
				if len(segments) != 1 || h.NSatellites == nil || *h.NSatellites != 111 || h.LeapSeconds == nil ||
					h.LeapSeconds.Current != 18 || h.LeapSeconds.DeltaFuture == nil || *h.LeapSeconds.DeltaFuture != 18 ||
					h.LeapSeconds.Week == nil || *h.LeapSeconds.Week != 1929 || h.LeapSeconds.Day == nil || *h.LeapSeconds.Day != 7 || h.LeapSeconds.TimeSystem != nil {
					t.Fatalf("satellite/leap data = %+v", h)
				}
				if got := h.PRNObsCounts["C02"]; !reflect.DeepEqual(got, []*uint64{uint64Ptr(1628), uint64Ptr(1266), uint64Ptr(2215), uint64Ptr(1628), uint64Ptr(1266), uint64Ptr(2215), uint64Ptr(1155), uint64Ptr(1261), uint64Ptr(2199), uint64Ptr(1628), uint64Ptr(1266), uint64Ptr(2215)}) {
					t.Fatalf("C02 observation counts = %+v", got)
				}
				wantBiases := []RINEXObservationGLONASSCodePhase{{"C1C", floatPtr(-71.94)}, {"C1P", floatPtr(-71.94)}, {"C2C", floatPtr(-71.94)}, {"C2P", floatPtr(-71.94)}}
				if !reflect.DeepEqual(h.GLONASSCodePhaseEntries, &wantBiases) ||
					h.MarkerName == nil || *h.MarkerName != "WTZR" || h.MarkerType != nil ||
					len(h.UnretainedHeaderLabels) != 1 || h.UnretainedHeaderLabels[0] != "RCV CLOCK OFFS APPL" {
					t.Fatalf("GLONASS bias/marker/unretained details = %+v", h)
				}
			}
		})
	}
}

func TestRINEXHeaderDetailsNumbersAndClosedHandle(t *testing.T) {
	var h RINEXObservationHeaderDetails
	err := json.Unmarshal([]byte(`{"version":3.05,"n_satellites":9007199254740993,"time_of_first_obs":{"epoch":{"year":2020,"month":6,"day":25,"hour":0,"minute":0,"second":"-Infinity"},"time_scale":"GPST"}}`), &h)
	if err != nil {
		t.Fatal(err)
	}
	if h.NSatellites == nil || *h.NSatellites != 9007199254740993 {
		t.Fatalf("large count lost precision: %+v", h.NSatellites)
	}
	if h.TimeOfFirstObs == nil || !math.IsInf(h.TimeOfFirstObs.Epoch.Second.Float64(), -1) {
		t.Fatalf("non-finite seconds not preserved: %+v", h.TimeOfFirstObs)
	}
	closed := &RINEXObservation{}
	if _, err := closed.HeaderDetails(); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil handle error = %v", err)
	}
}

func TestRINEXHeaderFloatJSONRoundTrip(t *testing.T) {
	values := []struct {
		name  string
		value float64
		want  string
	}{
		{"finite fraction", 123.125, "123.125"},
		{"negative zero", math.Copysign(0, -1), "-0"},
		{"NaN", math.NaN(), "\"NaN\""},
		{"positive infinity", math.Inf(1), "\"Infinity\""},
		{"negative infinity", math.Inf(-1), "\"-Infinity\""},
	}
	for _, test := range values {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(RINEXHeaderFloat64(test.value))
			if err != nil || string(encoded) != test.want {
				t.Fatalf("MarshalJSON = %s, %v; want %s", encoded, err, test.want)
			}
			var decoded RINEXHeaderFloat64
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			switch {
			case math.IsNaN(test.value):
				if !math.IsNaN(decoded.Float64()) {
					t.Fatalf("decoded = %v", decoded)
				}
			default:
				if math.Float64bits(decoded.Float64()) != math.Float64bits(test.value) {
					t.Fatalf("decoded bits = %x, want %x", math.Float64bits(decoded.Float64()), math.Float64bits(test.value))
				}
			}
		})
	}
}

func TestRINEXObservationHeaderDetailsJSONRoundTrip(t *testing.T) {
	count := uint64(9007199254740993)
	populated := RINEXHeaderFloat64(12.5)
	negativeZero := RINEXHeaderFloat64(math.Copysign(0, -1))
	nan := RINEXHeaderFloat64(math.NaN())
	entries := []RINEXObservationGLONASSCodePhase{
		{Code: "C1C", Bias: &populated},
		{Code: "C2C", Bias: nil},
	}
	header := RINEXObservationHeaderDetails{
		Version:                 3.05,
		NSatellites:             &count,
		ApproxPositionM:         &[3]RINEXHeaderFloat64{123.125, negativeZero, nan},
		GLONASSCodePhaseEntries: &entries,
	}
	encoded, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), "\"glonass_cod_phs_bis\":[[\"C1C\",12.5],[\"C2C\",null]]") {
		t.Fatalf("GLONASS entries are not serialized as native tuples: %s", encoded)
	}
	var decoded RINEXObservationHeaderDetails
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.NSatellites == nil || *decoded.NSatellites != count {
		t.Fatalf("large count round trip = %v", decoded.NSatellites)
	}
	if decoded.ApproxPositionM == nil ||
		float64((*decoded.ApproxPositionM)[0]) != 123.125 ||
		math.Float64bits(float64((*decoded.ApproxPositionM)[1])) != math.Float64bits(math.Copysign(0, -1)) ||
		!math.IsNaN(float64((*decoded.ApproxPositionM)[2])) {
		t.Fatalf("floating-point tuple round trip = %+v", decoded.ApproxPositionM)
	}
	if decoded.GLONASSCodePhaseEntries == nil || len(*decoded.GLONASSCodePhaseEntries) != 2 ||
		(*decoded.GLONASSCodePhaseEntries)[0].Code != "C1C" ||
		(*decoded.GLONASSCodePhaseEntries)[0].Bias == nil ||
		float64(*(*decoded.GLONASSCodePhaseEntries)[0].Bias) != 12.5 ||
		(*decoded.GLONASSCodePhaseEntries)[1].Code != "C2C" ||
		(*decoded.GLONASSCodePhaseEntries)[1].Bias != nil {
		t.Fatalf("GLONASS tuple round trip = %+v", decoded.GLONASSCodePhaseEntries)
	}
}

func TestRINEXObservationHeaderDetailsPopulatedScaleAndV2System(t *testing.T) {
	parse := func(lines ...string) RINEXObservationHeaderDetails {
		lines = append(lines, rinexHeaderLine("", "END OF HEADER"))
		obs, err := ParseRINEXObservation([]byte(strings.Join(lines, "\n")))
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := obs.Close(); err != nil {
				t.Error(err)
			}
		}()
		segments, err := obs.HeaderDetails()
		if err != nil {
			t.Fatal(err)
		}
		if len(segments) != 1 {
			t.Fatalf("segments = %d", len(segments))
		}
		return segments[0].Header
	}

	t.Run("scale factor", func(t *testing.T) {
		h := parse(
			rinexHeaderLine("     3.05           OBSERVATION DATA    M (MIXED)", "RINEX VERSION / TYPE"),
			rinexHeaderLine("G    1 C1C", "SYS / # / OBS TYPES"),
			rinexHeaderLine("G   10  1 C1C", "SYS / SCALE FACTOR"),
		)
		if len(h.ScaleFactors) != 1 || h.ScaleFactors[0].System != "GPS" ||
			h.ScaleFactors[0].Factor != 10 || !reflect.DeepEqual(h.ScaleFactors[0].Codes, []string{"C1C"}) {
			t.Fatalf("scale factors = %+v", h.ScaleFactors)
		}
	})
	t.Run("phase shift satellite details", func(t *testing.T) {
		h := parse(
			rinexHeaderLine("     3.05           OBSERVATION DATA    M (MIXED)", "RINEX VERSION / TYPE"),
			rinexHeaderLine("G    1 C1C", "SYS / # / OBS TYPES"),
			rinexHeaderLine("G L1C  0.25000  01 G01", "SYS / PHASE SHIFT"),
			rinexHeaderLine("R L1C  0.50000  01 R00", "SYS / PHASE SHIFT"),
		)
		if len(h.PhaseShifts) != 2 ||
			h.PhaseShifts[0].System != "GPS" || h.PhaseShifts[0].Code == nil || *h.PhaseShifts[0].Code != "L1C" ||
			h.PhaseShifts[0].CorrectionCycles == nil || *h.PhaseShifts[0].CorrectionCycles != .25 ||
			!reflect.DeepEqual(h.PhaseShifts[0].Satellites, []string{"G01"}) || len(h.PhaseShifts[0].UnrepresentableSatellites) != 0 ||
			h.PhaseShifts[1].System != "GLONASS" || h.PhaseShifts[1].Code == nil || *h.PhaseShifts[1].Code != "L1C" ||
			h.PhaseShifts[1].CorrectionCycles == nil || *h.PhaseShifts[1].CorrectionCycles != .5 ||
			len(h.PhaseShifts[1].Satellites) != 0 || !reflect.DeepEqual(h.PhaseShifts[1].UnrepresentableSatellites, []string{"R00"}) {
			t.Fatalf("phase shift details = %+v", h.PhaseShifts)
		}
	})
	t.Run("RINEX 2 system", func(t *testing.T) {
		h := parse(
			rinexHeaderLine("     2.11           OBSERVATION DATA    G (GPS)", "RINEX VERSION / TYPE"),
			rinexHeaderLine("     2    C1    L1", "# / TYPES OF OBSERV"),
		)
		if h.RINEX2System == nil || *h.RINEX2System != "GPS" ||
			!reflect.DeepEqual(h.RINEX2Types, []string{"C1", "L1"}) {
			t.Fatalf("RINEX 2 system/types = %v, %+v", h.RINEX2System, h.RINEX2Types)
		}
	})
}

func TestRINEXObservationGLONASSHeaderBiasLookup(t *testing.T) {
	parseHeader := func(version string, records ...string) RINEXObservationHeaderDetails {
		lines := []string{
			rinexHeaderLine("     "+version+"           OBSERVATION DATA    M (MIXED)", "RINEX VERSION / TYPE"),
			rinexHeaderLine("G    1 C1C", "SYS / # / OBS TYPES"),
		}
		for _, record := range records {
			lines = append(lines, rinexHeaderLine(record, "GLONASS COD/PHS/BIS"))
		}
		lines = append(lines, rinexHeaderLine("", "END OF HEADER"))
		obs, err := ParseRINEXObservation([]byte(strings.Join(lines, "\n")))
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := obs.Close(); err != nil {
				t.Error(err)
			}
		}()
		segments, err := obs.HeaderDetails()
		if err != nil {
			t.Fatal(err)
		}
		if len(segments) != 1 {
			t.Fatalf("header segments = %d", len(segments))
		}
		return segments[0].Header
	}

	t.Run("populated value", func(t *testing.T) {
		header := parseHeader("3.05", " C1C  -71.940")
		got, present, err := header.GLONASSCodePhaseBias("C1C")
		if err != nil || !present || got != -71.94 {
			t.Fatalf("lookup = %v, %t, %v", got, present, err)
		}
		if _, present, err := header.GLONASSCodePhaseBias("C2C"); err != nil || present {
			t.Fatalf("missing code lookup present=%t err=%v", present, err)
		}
	})
	t.Run("blank table and blank bias are unknown", func(t *testing.T) {
		table := parseHeader("3.05", "")
		if table.GLONASSCodePhaseEntries == nil || len(*table.GLONASSCodePhaseEntries) != 0 {
			t.Fatalf("blank table details = %+v", table.GLONASSCodePhaseEntries)
		}
		blank := parseHeader("3.05", " C1C")
		if blank.GLONASSCodePhaseEntries == nil || len(*blank.GLONASSCodePhaseEntries) != 1 ||
			(*blank.GLONASSCodePhaseEntries)[0].Code != "C1C" || (*blank.GLONASSCodePhaseEntries)[0].Bias != nil {
			t.Fatalf("blank bias details = %+v", blank.GLONASSCodePhaseEntries)
		}
		for _, header := range []RINEXObservationHeaderDetails{table, blank} {
			_, present, err := header.GLONASSCodePhaseBias("C1C")
			var unavailable *RINEXCorrectionUnavailableError
			if present || !errors.As(err, &unavailable) || unavailable.Kind != "Unknown" {
				t.Fatalf("unknown lookup present=%t err=%#v", present, err)
			}
		}
	})
	t.Run("conflicting duplicate is ambiguous", func(t *testing.T) {
		header := parseHeader("3.05", " C1C  -71.940 C1C  -72.000")
		_, present, err := header.GLONASSCodePhaseBias("C1C")
		var unavailable *RINEXCorrectionUnavailableError
		if present || !errors.As(err, &unavailable) || unavailable.Kind != "Ambiguous" ||
			len(unavailable.Corrections) != 2 || unavailable.Corrections[0] == nil ||
			*unavailable.Corrections[0] != -71.94 || unavailable.Corrections[1] == nil ||
			*unavailable.Corrections[1] != -72 {
			t.Fatalf("ambiguous lookup present=%t err=%#v", present, err)
		}
	})
	t.Run("RINEX 4 ignores legacy bias records", func(t *testing.T) {
		header := parseHeader("4.00", " C1C  -71.940")
		if got, present, err := header.GLONASSCodePhaseBias("C1C"); err != nil || present || got != 0 {
			t.Fatalf("RINEX 4 lookup = %v, %t, %v", got, present, err)
		}
	})
}

func rinexHeaderLine(body, label string) string {
	if len(body) > 60 {
		panic("RINEX header body exceeds 60 columns")
	}
	return body + strings.Repeat(" ", 60-len(body)) + label
}

func uint64Ptr(value uint64) *uint64 { return &value }
func floatPtr(value float64) *RINEXHeaderFloat64 {
	converted := RINEXHeaderFloat64(value)
	return &converted
}
