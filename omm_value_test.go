package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"sidereon.dev/go/v3/internal/native"
)

const ommValueSnapshotFixture = `CCSDS_OMM_VERS = 3.0
COMMENT header snapshot
CLASSIFICATION = SBU
CREATION_DATE = 2020-065T16:00:00
ORIGINATOR = NOAA
MESSAGE_ID = OMM 202013719185
COMMENT metadata snapshot
OBJECT_NAME = GOES 9
OBJECT_ID = 1995-025A
CENTER_NAME = EARTH
REF_FRAME = TEME
REF_FRAME_EPOCH = 2020-064T00:00:00
TIME_SYSTEM = UTC
MEAN_ELEMENT_THEORY = SGP/SGP4
COMMENT mean-elements snapshot
EPOCH = 2020-064T10:34:41.4264
MEAN_MOTION = 1.00273272
ECCENTRICITY = 0.0005013
INCLINATION = 3.0539
RA_OF_ASC_NODE = 81.7939
ARG_OF_PERICENTER = 249.2363
MEAN_ANOMALY = 150.1602
GM = 398600.8
COMMENT spacecraft snapshot
MASS = 2500 [kg]
DRAG_COEFF = 2.2
COMMENT TLE-parameters snapshot
EPHEMERIS_TYPE = 0
CLASSIFICATION_TYPE = U
NORAD_CAT_ID = 23581
ELEMENT_SET_NO = 925
REV_AT_EPOCH = 4316
BSTAR = 0.0001
MEAN_MOTION_DOT = -0.00000113
MEAN_MOTION_DDOT = 0.0
COMMENT covariance snapshot
COV_REF_FRAME = TEME
CX_X = 3.331349476038534e-04
CY_X = 4.618927349220216e-04
CY_Y = 6.782421679971363e-04
CZ_X = -3.070007847730449e-04
CZ_Y = -4.221234189514228e-04
CZ_Z = 3.231931992380369e-04
CX_DOT_X = -3.349365033922630e-07
CX_DOT_Y = -4.686084221046758e-07
CX_DOT_Z = 2.484949578400095e-07
CX_DOT_X_DOT = 4.296022805587290e-10
CY_DOT_X = -2.211832501084875e-07
CY_DOT_Y = -2.864186892102733e-07
CY_DOT_Z = 1.798098699846038e-07
CY_DOT_X_DOT = 2.608899201686016e-10
CY_DOT_Y_DOT = 1.767514756338532e-10
CZ_DOT_X = -3.041346050686871e-07
CZ_DOT_Y = -4.989496988610662e-07
CZ_DOT_Z = 3.540310904497689e-07
CZ_DOT_X_DOT = 1.869263192954590e-10
CZ_DOT_Y_DOT = 1.008862586240695e-10
CZ_DOT_Z_DOT = 6.224444338635500e-10
COMMENT user-defined snapshot
USER_DEFINED_EARTH_MODEL = WGS-84
USER_DEFINED_C3 = 29.376 [km**2/s**2]
`

func TestOMMSnapshotRetainsAllFieldsCommentsAndBridge(t *testing.T) {
	omm, err := ParseOMMKVN([]byte(ommValueSnapshotFixture))
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "snapshot OMM", omm.Close)

	snapshot, err := omm.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.CCSDSOMMVERS == nil || *snapshot.CCSDSOMMVERS != "3.0" || snapshot.Classification == nil || *snapshot.Classification != "SBU" || snapshot.MessageID == nil || *snapshot.MessageID != "OMM 202013719185" {
		t.Fatalf("header snapshot lost values: %+v", snapshot)
	}
	if snapshot.Epoch.Year != 2020 || snapshot.Epoch.Month != 3 || snapshot.Epoch.Day != 4 || snapshot.Epoch.Microsecond != 426400 {
		t.Fatalf("epoch snapshot = %+v", snapshot.Epoch)
	}
	if snapshot.Spacecraft == nil || snapshot.Spacecraft.MassKG == nil || *snapshot.Spacecraft.MassKG != 2500 || snapshot.Covariance == nil || snapshot.Covariance.LowerTriangle[20] != 6.2244443386355e-10 || len(snapshot.UserDefined) != 2 {
		t.Fatalf("nested snapshot lost values: spacecraft=%+v covariance=%+v user=%+v", snapshot.Spacecraft, snapshot.Covariance, snapshot.UserDefined)
	}
	if fmt.Sprint(snapshot.Comments.Header) != "[header snapshot]" || fmt.Sprint(snapshot.Comments.Metadata) != "[metadata snapshot]" || fmt.Sprint(snapshot.Comments.MeanElements) != "[mean-elements snapshot]" || fmt.Sprint(snapshot.Comments.TLEParameters) != "[TLE-parameters snapshot]" || fmt.Sprint(snapshot.Comments.UserDefined) != "[user-defined snapshot]" || fmt.Sprint(snapshot.Spacecraft.Comments) != "[spacecraft snapshot]" || fmt.Sprint(snapshot.Covariance.Comments) != "[covariance snapshot]" {
		t.Fatalf("snapshot collapsed block comments: %+v", snapshot.Comments)
	}
	if snapshot.ExactSGP4Epoch != nil || !snapshot.QuantizeTLEDerivedFields {
		t.Fatalf("parsed OMM side channels = exact %v, quantize %t", snapshot.ExactSGP4Epoch, snapshot.QuantizeTLEDerivedFields)
	}
	assertString := func(name string, got *string, want string) {
		t.Helper()
		if got == nil || *got != want {
			t.Fatalf("%s = %v, want %q", name, got, want)
		}
	}
	assertFloat := func(name string, got *float64, want float64) {
		t.Helper()
		if got == nil || *got != want {
			t.Fatalf("%s = %v, want %.17g", name, got, want)
		}
	}
	assertInt32 := func(name string, got *int32, want int32) {
		t.Helper()
		if got == nil || *got != want {
			t.Fatalf("%s = %v, want %d", name, got, want)
		}
	}
	assertString("creation_date", snapshot.CreationDate, "2020-065T16:00:00")
	assertString("originator", snapshot.Originator, "NOAA")
	assertString("object_name", snapshot.ObjectName, "GOES 9")
	assertString("object_id", snapshot.ObjectID, "1995-025A")
	assertString("center_name", snapshot.CenterName, "EARTH")
	assertString("ref_frame", snapshot.RefFrame, "TEME")
	assertString("ref_frame_epoch", snapshot.RefFrameEpoch, "2020-064T00:00:00")
	assertString("time_system", snapshot.TimeSystem, "UTC")
	assertString("mean_element_theory", snapshot.MeanElementTheory, "SGP/SGP4")
	if snapshot.SemiMajorAxisKM != nil {
		t.Fatalf("semi_major_axis_km = %v, want absent when mean_motion is stated", *snapshot.SemiMajorAxisKM)
	}
	if snapshot.InclinationDeg != 3.0539 || snapshot.RAOfAscNodeDeg != 81.7939 || snapshot.ArgOfPericenterDeg != 249.2363 || snapshot.MeanAnomalyDeg != 150.1602 {
		t.Fatalf("angular elements = inclination %.17g, RA %.17g, argument %.17g, anomaly %.17g", snapshot.InclinationDeg, snapshot.RAOfAscNodeDeg, snapshot.ArgOfPericenterDeg, snapshot.MeanAnomalyDeg)
	}
	assertFloat("gm_km3_s2", snapshot.GMKM3S2, 398600.8)
	assertInt32("ephemeris_type", snapshot.EphemerisType, 0)
	assertString("classification_type", snapshot.ClassificationType, "U")
	if snapshot.NORADCatID == nil || *snapshot.NORADCatID != 23581 {
		t.Fatalf("norad_cat_id = %v, want 23581", snapshot.NORADCatID)
	}
	assertInt32("element_set_no", snapshot.ElementSetNo, 925)
	if snapshot.RevAtEpoch == nil || *snapshot.RevAtEpoch != 4316 {
		t.Fatalf("rev_at_epoch = %v, want 4316", snapshot.RevAtEpoch)
	}
	assertFloat("bstar", snapshot.BStar, 0.0001)
	if snapshot.BTermM2KG != nil {
		t.Fatalf("bterm_m2_kg = %v, want absent", *snapshot.BTermM2KG)
	}
	if snapshot.AGOMM2KG != nil {
		t.Fatalf("agom_m2_kg = %v, want absent", *snapshot.AGOMM2KG)
	}
	// Reconstruct through the lossless snapshot constructor and compare every
	// declared field, including all optional presence, comment buckets and the
	// complete 21-value covariance block.
	rebuilt, err := NewOMMFromData(snapshot)
	if err != nil {
		t.Fatalf("construct from complete OMM snapshot: %v", err)
	}
	cleanupClose(t, "rebuilt OMM", rebuilt.Close)
	rebuiltSnapshot, err := rebuilt.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot, rebuiltSnapshot) {
		t.Fatalf("complete snapshot changed during reconstruction:\noriginal=%+v\nrebuilt=%+v", snapshot, rebuiltSnapshot)
	}
	// Mutating source-owned slices after construction cannot mutate the native
	// handle; snapshots are also detached from each other.
	if len(snapshot.Covariance.LowerTriangle) != 21 {
		t.Fatalf("covariance has %d values", len(snapshot.Covariance.LowerTriangle))
	}
	snapshot.Covariance.LowerTriangle[0] = 99
	snapshot.Comments.Header[0] = "changed"
	if again, err := rebuilt.Snapshot(); err != nil || again.Covariance.LowerTriangle[0] == 99 || again.Comments.Header[0] == "changed" {
		t.Fatalf("native snapshot retained caller mutation: %+v, %v", again, err)
	}

	elements, err := omm.ToElementSet()
	if err != nil {
		t.Fatal(err)
	}
	if math.IsNaN(elements.Epoch.Whole) || math.IsInf(elements.Epoch.Whole, 0) || elements.Epoch.Whole+elements.Epoch.Fraction < 2458912 || elements.Epoch.Whole+elements.Epoch.Fraction >= 2458913 || elements.Epoch.Fraction < 0 || elements.Epoch.Fraction >= 1 || elements.OMMEpochDays == nil || elements.MeanMotionDot == nil || *elements.MeanMotionDot != -0.00000113 || elements.MeanMotionDoubleDot == nil || *elements.MeanMotionDoubleDot != 0 {
		t.Fatalf("core element-set conversion lost optional bridge fields: %+v", elements)
	}
	if elements.Eccentricity != snapshot.Eccentricity || elements.MeanMotionRevPerDay != *snapshot.MeanMotion {
		t.Fatalf("element-set values disagree with detached OMM data: %+v / %+v", elements, snapshot)
	}
	rebuiltElements, err := rebuilt.ToElementSet()
	if err != nil || !reflect.DeepEqual(rebuiltElements, elements) {
		t.Fatalf("rebuilt bridge result changed: %+v, err=%v", rebuiltElements, err)
	}

	invalid, err := ParseOMMKVN([]byte(strings.Replace(ommValueSnapshotFixture, "CENTER_NAME = EARTH", "CENTER_NAME = MARS", 1)))
	if err != nil {
		t.Fatalf("parse OMM with bridge-incompatible metadata: %v", err)
	}
	cleanupClose(t, "incompatible OMM", invalid.Close)
	if _, err := invalid.ToElementSet(); err == nil {
		t.Fatal("bridge accepted incompatible CENTER_NAME")
	} else {
		var status *StatusError
		if !errors.As(err, &status) || status.Engine == nil || status.Engine.Family != EngineErrorFamilyOMM || status.Engine.Kind == "" {
			t.Fatalf("bridge refusal lost typed OMM engine error: %#v", err)
		}
	}
	invalidData, err := invalid.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	readable, err := NewOMMFromData(invalidData)
	if err != nil {
		t.Fatalf("snapshot construction incorrectly required SGP4 bridge compatibility: %v", err)
	}
	cleanupClose(t, "readable non-SGP4 OMM", readable.Close)
	if _, err := readable.ToElementSet(); err == nil {
		t.Fatal("rebuilt non-SGP4 OMM unexpectedly bridged")
	}
}

func TestOMMSnapshotConstructorRejectsNonFiniteAndClosedAccess(t *testing.T) {
	var nilOMM *OMM
	if _, err := nilOMM.Snapshot(); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil snapshot error = %v", err)
	}
	parsed, err := ParseOMMKVN([]byte(ommValueSnapshotFixture))
	if err != nil {
		t.Fatal(err)
	}
	data, err := parsed.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	data.InclinationDeg = math.NaN()
	if _, err := NewOMMFromData(data); err == nil {
		t.Fatal("constructor accepted NaN numeric field")
	}
	if err := parsed.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := parsed.Snapshot(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed snapshot error = %v", err)
	}
}

func TestOMMFromDataRetainsExplicitEmptyUnicodeAndOrder(t *testing.T) {
	parsed, err := ParseOMMKVN([]byte(ommValueSnapshotFixture))
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "source OMM", parsed.Close)
	data, err := parsed.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := parsed.Close(); err != nil {
		t.Fatal(err)
	}
	empty := ""
	data.MessageID = &empty
	data.Comments.Header = []string{"orbit 🌍", "second header comment"}
	data.UserDefined = []OMMUserDefined{{Parameter: "EMPTY", Value: ""}, {Parameter: "TEXT", Value: "μ-value"}}
	data.Covariance.LowerTriangle = [21]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}
	zero := float64(0)
	data.MeanMotionDDot = &zero
	constructed, err := NewOMMFromData(data)
	if err != nil {
		t.Fatalf("construct exact OMM data: %v", err)
	}
	cleanupClose(t, "constructed OMM", constructed.Close)
	got, err := constructed.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(data, got) {
		t.Fatalf("constructor lost an explicit value, order, or presence:\nwant=%+v\ngot=%+v", data, got)
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	for label, malformed := range map[string][]byte{
		"duplicate top key":    bytes.Replace(encoded, []byte(`"classification":`), []byte(`"classification":null,"classification":`), 1),
		"duplicate nested key": bytes.Replace(encoded, []byte(`"header":`), []byte(`"header":[],"header":`), 1),
	} {
		if _, err := native.OMMFromSnapshotJSON(malformed); err == nil {
			t.Errorf("native constructor accepted %s", label)
		}
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	raw["eccentricity"] = json.RawMessage("null")
	missingNumeric, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := native.OMMFromSnapshotJSON(missingNumeric); err == nil {
		t.Fatal("native constructor accepted null required numeric field")
	}
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	var covariance map[string]json.RawMessage
	if err := json.Unmarshal(raw["covariance"], &covariance); err != nil {
		t.Fatal(err)
	}
	var triangle []float64
	if err := json.Unmarshal(covariance["lower_triangle"], &triangle); err != nil {
		t.Fatal(err)
	}
	covariance["lower_triangle"], _ = json.Marshal(triangle[:20])
	raw["covariance"], _ = json.Marshal(covariance)
	malformedArray, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := native.OMMFromSnapshotJSON(malformedArray); err == nil {
		t.Fatal("native constructor accepted malformed covariance length")
	}
}

func TestOMMSnapshotDuplicateKeyScanner(t *testing.T) {
	for _, input := range []string{`{"a":1,"a":2}`, `{"nested":{"x":1,"x":2}}`} {
		if err := validateUniqueJSONKeys([]byte(input)); err == nil {
			t.Errorf("accepted duplicate JSON key in %s", input)
		}
	}
}
