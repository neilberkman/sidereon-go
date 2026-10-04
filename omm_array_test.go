package sidereon

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func readOMMFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/omm/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestOMMAutodetectAndSingleCSVEntryPoints(t *testing.T) {
	for _, name := range []string{"25544.json", "25544.xml", "25544.kvn"} {
		omm, err := ParseOMM(readOMMFixture(t, name))
		if err != nil {
			t.Fatalf("autodetect %s: %v", name, err)
		}
		data, err := omm.Snapshot()
		closeErr := omm.Close()
		if err != nil || closeErr != nil || data.NORADCatID == nil || *data.NORADCatID != 25544 {
			t.Fatalf("autodetect %s returned %+v, err=%v close=%v", name, data, err, closeErr)
		}
	}
	csv := []byte("OBJECT_NAME,OBJECT_ID,EPOCH,MEAN_MOTION,ECCENTRICITY,INCLINATION,RA_OF_ASC_NODE,ARG_OF_PERICENTER,MEAN_ANOMALY,EPHEMERIS_TYPE,CLASSIFICATION_TYPE,NORAD_CAT_ID,ELEMENT_SET_NO,REV_AT_EPOCH,BSTAR,MEAN_MOTION_DOT,MEAN_MOTION_DDOT\n" +
		"ISS (ZARYA),1998-067A,2026-06-17T04:32:52.099296,15.49273435,0.0004737,51.6332,300.0813,195.1146,164.9702,0,U,25544,999,57175,0.00017172,9.113e-5,0")
	omm, err := ParseOMMCSV(csv)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, omm)
	data, err := omm.Snapshot()
	if err != nil || data.NORADCatID == nil || *data.NORADCatID != 25544 {
		t.Fatalf("single GP CSV parse returned %+v, err=%v", data, err)
	}
}

func TestOMMArrayRetainsSuccessesAndTypedJSONSkips(t *testing.T) {
	object := bytes.TrimSpace(readOMMFixture(t, "25544.json"))
	object = bytes.Trim(object, "[] \r\n\t")
	input := []byte("[" + string(object) + ",null," + string(object) + "]")
	array, err := ParseOMMJSONArray(input)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, array)

	values, err := array.Values()
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[0].NORADCatID == nil || *values[0].NORADCatID != 25544 || values[1].NORADCatID == nil || *values[1].NORADCatID != 25544 {
		t.Fatalf("unexpected successful records: %+v", values)
	}
	skipped, err := array.SkippedRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 || skipped[0].Index != 1 || skipped[0].Error.Kind != "field" || skipped[0].Error.Fields.Message == nil {
		t.Fatalf("unexpected typed skip: %+v", skipped)
	}
	if _, err := ParseOMMJSONArray([]byte("null")); err == nil {
		t.Fatal("malformed top-level JSON was not refused")
	}
	empty, err := ParseOMMJSONArray([]byte("[]"))
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, empty)
	if count, err := empty.Len(); err != nil || count != 0 {
		t.Fatalf("empty input returned count=%d err=%v", count, err)
	}
}

func TestOMMArrayCSVQuotedMultilineAndWriterCommentPolicy(t *testing.T) {
	csv := "OBJECT_NAME,OBJECT_ID,EPOCH,MEAN_MOTION,ECCENTRICITY,INCLINATION,RA_OF_ASC_NODE,ARG_OF_PERICENTER,MEAN_ANOMALY,EPHEMERIS_TYPE,CLASSIFICATION_TYPE,NORAD_CAT_ID,ELEMENT_SET_NO,REV_AT_EPOCH,BSTAR,MEAN_MOTION_DOT,MEAN_MOTION_DDOT\n" +
		"\"ISS,\nZARYA\",1998-067A,2026-06-17T04:32:52.099296,15.49273435,0.0004737,51.6332,300.0813,195.1146,164.9702,0,U,25544,999,57175,0.00017172,9.113e-5,0\n" +
		"bad,row\n"
	array, err := ParseOMMCSVArray([]byte(csv))
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, array)
	values, err := array.Values()
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 1 || values[0].ObjectName == nil || *values[0].ObjectName != "ISS,\nZARYA" {
		t.Fatalf("quoted multiline record was not retained: %+v", values)
	}
	skipped, err := array.SkippedRecords()
	if err != nil || len(skipped) != 1 || skipped[0].Index != 1 || skipped[0].Error.Kind != "csv_column_count" {
		t.Fatalf("CSV row-count refusal lost: skipped=%+v err=%v", skipped, err)
	}

	parsed, err := ParseOMM([]byte(csv[:strings.LastIndex(csv, "bad,row")]))
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, parsed)
	data, err := parsed.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	data.Comments.Metadata = []string{"preserve this comment"}
	commented, err := NewOMMFromData(data)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, commented)
	assembled, err := NewOMMArray(commented)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, assembled)
	if err := commented.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := assembled.JSON(); err == nil {
		t.Fatal("strict JSON writer accepted an unrepresentable comment")
	}
	if _, err := assembled.CSV(); err == nil {
		t.Fatal("strict CSV writer accepted an unrepresentable comment")
	}
	jsonBytes, err := assembled.JSONDiscardingComments()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseOMMJSONArray(jsonBytes); err != nil {
		t.Fatalf("comment-discarding JSON is not parseable: %v", err)
	}
	csvBytes, err := assembled.CSVDiscardingComments()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseOMMCSVArray(csvBytes); err != nil {
		t.Fatalf("comment-discarding CSV is not parseable: %v", err)
	}
}

func TestOMMXMLAllRetainsMessageOrderAndSkippedMessage(t *testing.T) {
	document := string(readOMMFixture(t, "25544.xml"))
	start := strings.Index(document, "<omm ")
	end := strings.Index(document, "</omm>")
	if start < 0 || end < start {
		t.Fatal("fixture has no complete OMM element")
	}
	message := document[start : end+len("</omm>")]
	combined := "<ndm>" + message + "<omm><header/><body/></omm>" + message + "</ndm>"
	array, err := ParseOMMXMLAll([]byte(combined))
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, array)
	values, err := array.Values()
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[0].NORADCatID == nil || *values[0].NORADCatID != 25544 || values[1].NORADCatID == nil || *values[1].NORADCatID != 25544 {
		t.Fatalf("XML successes changed order or content: %+v", values)
	}
	values[0].ObjectName = nil
	retained, err := array.Values()
	if err != nil || retained[0].ObjectName == nil || *retained[0].ObjectName != "ISS (ZARYA)" {
		t.Fatalf("returned snapshots alias retained OMM data: values=%+v err=%v", retained, err)
	}
	skipped, err := array.SkippedRecords()
	if err != nil || len(skipped) != 1 || skipped[0].Index != 1 || skipped[0].Error.Kind == "" {
		t.Fatalf("XML skipped message lost: skipped=%+v err=%v", skipped, err)
	}
}

func TestOMMEpochPreservesFemtosecondsAndLeapSecond(t *testing.T) {
	value, err := ParseOMMEpoch("2016-12-31T23:59:60.123456789123456")
	if err != nil {
		t.Fatal(err)
	}
	want := OMMEpoch{Year: 2016, Month: 12, Day: 31, Hour: 23, Minute: 59, Second: 60, Microsecond: 123456, Femtosecond: 789123456}
	if value != want {
		t.Fatalf("epoch precision lost: got %+v want %+v", value, want)
	}
}

func TestOMMFromOMMPropagationOwnsSatelliteAfterSourceClose(t *testing.T) {
	omm, err := ParseOMM(readOMMFixture(t, "25544.json"))
	if err != nil {
		t.Fatal(err)
	}
	satellite, err := NewSGP4SatelliteFromOMM(omm)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, satellite)
	elements, err := omm.ToElementSet()
	if err != nil {
		t.Fatal(err)
	}
	satelliteEpoch, err := satellite.Epoch()
	if err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(satelliteEpoch.Whole) != math.Float64bits(elements.Epoch.Whole) || math.Float64bits(satelliteEpoch.Fraction) != math.Float64bits(elements.Epoch.Fraction) {
		t.Fatalf("owned SGP4 epoch lost split bits: got %+v want %+v", satelliteEpoch, elements.Epoch)
	}
	if err := omm.Close(); err != nil {
		t.Fatal(err)
	}
	minutes, err := satellite.PropagateMinutesSinceEpoch(0)
	if err != nil {
		t.Fatal(err)
	}
	state, err := satellite.PropagateJulianDate(elements.Epoch)
	if err != nil {
		t.Fatal(err)
	}
	if minutes != state {
		t.Fatalf("minutes and split-date routes disagree at the exact element epoch: %+v %+v", minutes, state)
	}
	wantEpoch := (elements.Epoch.Whole-2451545.0)*86400.0 + elements.Epoch.Fraction*86400.0
	if math.Float64bits(minutes.EpochJ2000S) != math.Float64bits(wantEpoch) {
		t.Fatalf("stored core epoch differs by bits: got %.17g want %.17g", minutes.EpochJ2000S, wantEpoch)
	}
	if _, err := satellite.PropagateMinutesSinceEpoch(math.NaN()); err == nil {
		t.Fatal("non-finite SGP4 offset was accepted")
	}
	if err := satellite.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := satellite.Epoch(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed satellite returned its split epoch: %v", err)
	}
}

func TestOMMFromOMMPropagationMatchesPythonSGP4Reference(t *testing.T) {
	type state struct {
		Minutes     float64   `json:"minutes_since_epoch"`
		Error       int       `json:"error"`
		Position    []float64 `json:"position_km"`
		PositionHex []string  `json:"position_km_hex"`
		Velocity    []float64 `json:"velocity_km_s"`
		VelocityHex []string  `json:"velocity_km_s_hex"`
	}
	var reference struct {
		Generator  string  `json:"generator"`
		Version    string  `json:"sgp4_version"`
		Fixture    string  `json:"input_fixture"`
		FixtureSHA string  `json:"input_fixture_sha256"`
		States     []state `json:"states"`
	}
	referenceBytes, err := os.ReadFile("testdata/omm/python_sgp4_25544_state.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(referenceBytes, &reference); err != nil {
		t.Fatal(err)
	}
	fixture := readOMMFixture(t, "25544.json")
	fixtureDigest := sha256.Sum256(fixture)
	if reference.Generator != "python-sgp4 2.22 accelerated Satrec.sgp4_tsince" || reference.Version != "2.22" || reference.Fixture != "25544.json" || reference.FixtureSHA != hex.EncodeToString(fixtureDigest[:]) {
		t.Fatalf("Python-SGP4 reference provenance does not match its checked-in input: %+v", reference)
	}
	if len(reference.States) != 2 || reference.States[0].Minutes != 0 || reference.States[1].Minutes != 10 {
		t.Fatalf("unexpected Python-SGP4 reference epochs: %+v", reference.States)
	}
	omm, err := ParseOMM(fixture)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, omm)
	satellite, err := NewSGP4SatelliteFromOMM(omm)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, satellite)
	if err := omm.Close(); err != nil {
		t.Fatal(err)
	}

	const positionToleranceKM = 1e-3
	const velocityToleranceKMPerS = 1e-6
	var maxPositionDelta, maxVelocityDelta float64
	epochPositionX := 0.0
	for _, want := range reference.States {
		if want.Error != 0 || len(want.Position) != 3 || len(want.PositionHex) != 3 || len(want.Velocity) != 3 || len(want.VelocityHex) != 3 {
			t.Fatalf("invalid external reference state: %+v", want)
		}
		for i := 0; i < 3; i++ {
			position, err := strconv.ParseFloat(want.PositionHex[i], 64)
			if err != nil || math.Float64bits(position) != math.Float64bits(want.Position[i]) {
				t.Fatalf("reference position decimal/hex mismatch at component %d: %q %.17g err=%v", i, want.PositionHex[i], want.Position[i], err)
			}
			velocity, err := strconv.ParseFloat(want.VelocityHex[i], 64)
			if err != nil || math.Float64bits(velocity) != math.Float64bits(want.Velocity[i]) {
				t.Fatalf("reference velocity decimal/hex mismatch at component %d: %q %.17g err=%v", i, want.VelocityHex[i], want.Velocity[i], err)
			}
		}
		got, err := satellite.PropagateMinutesSinceEpoch(want.Minutes)
		if err != nil {
			t.Fatal(err)
		}
		if want.Minutes == 0 {
			epochPositionX = got.PositionKm[0]
		}
		for i := 0; i < 3; i++ {
			positionDelta := math.Abs(got.PositionKm[i] - want.Position[i])
			velocityDelta := math.Abs(got.VelocityKmPerS[i] - want.Velocity[i])
			maxPositionDelta = math.Max(maxPositionDelta, positionDelta)
			maxVelocityDelta = math.Max(maxVelocityDelta, velocityDelta)
			if positionDelta > positionToleranceKM || velocityDelta > velocityToleranceKMPerS {
				t.Fatalf("python-sgp4 state differs at %+g min: position delta %.9g km, velocity delta %.9g km/s", want.Minutes, positionDelta, velocityDelta)
			}
		}
	}
	t.Logf("Python-SGP4 2.22 maximum observed state residuals: %.9g km, %.9g km/s", maxPositionDelta, maxVelocityDelta)
	perturbed := reference.States[0].Position[0] + 0.01
	if math.Abs(epochPositionX-perturbed) <= positionToleranceKM {
		t.Fatal("reference tolerance accepted a materially perturbed position")
	}
}

func TestOMMToSGP4IncompatibleMetadataRetainsTypedCause(t *testing.T) {
	omm, err := ParseOMM(readOMMFixture(t, "25544.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, omm)
	data, err := omm.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	center := "MARS"
	data.CenterName = &center
	incompatible, err := NewOMMFromData(data)
	if err != nil {
		t.Fatal(err)
	}
	defer closeAfterTest(t, incompatible)
	if _, err := NewSGP4SatelliteFromOMM(incompatible); err == nil {
		t.Fatal("non-Earth OMM initialized an SGP4 satellite")
	} else {
		var status *StatusError
		if !errors.As(err, &status) || status.SGP4 == nil || status.SGP4.Kind == SGP4ErrorNone {
			t.Fatalf("incompatible OMM refusal lost typed SGP4 details: %T %v", err, err)
		}
	}
}
