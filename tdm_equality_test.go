//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"errors"
	"testing"
)

func parseEqualityTDM(t *testing.T, text string) *TDM {
	t.Helper()
	value, err := ParseTDMKVN([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := value.Close(); err != nil {
			t.Error(err)
		}
	})
	return value
}

func TestTDMEqual(t *testing.T) {
	first := parseEqualityTDM(t, tdmExpectedKVN)
	second := parseEqualityTDM(t, tdmExpectedKVN)

	equal, err := first.Equal(second)
	if err != nil || !equal {
		t.Fatalf("Equal(same text) = %t, %v", equal, err)
	}
	if err := second.RemoveRecord(0, 0); err != nil {
		t.Fatal(err)
	}
	equal, err = first.Equal(second)
	if err != nil || equal {
		t.Fatalf("Equal(edited text) = %t, %v", equal, err)
	}

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if equal, err = first.Equal(second); equal || err != ErrClosed {
		t.Fatalf("closed Equal = %t, %v; want false, ErrClosed", equal, err)
	}
	if equal, err = second.Equal(nil); equal || err != ErrClosed {
		t.Fatalf("nil argument Equal = %t, %v; want false, ErrClosed", equal, err)
	}
	var nilTDM *TDM
	if equal, err = nilTDM.Equal(second); equal || err != ErrClosed {
		t.Fatalf("nil receiver Equal = %t, %v; want false, ErrClosed", equal, err)
	}
}

func TestTDMComparableValues(t *testing.T) {
	if TDMLeniencyStrict != TDMLeniency(0) || TDMLeniencyStrict == TDMLeniencyForgive {
		t.Fatal("TDMLeniency value comparison failed")
	}
	if TDMObservableRange != TDMObservable(0) || TDMObservableRange == TDMObservableOther {
		t.Fatal("TDMObservable value comparison failed")
	}
	if TDMUnitKilometers != TDMUnit(0) || TDMUnitKilometers == TDMUnitUnknown {
		t.Fatal("TDMUnit value comparison failed")
	}
	if TDMSectionHeader != TDMSection(0) || TDMSectionHeader == TDMSectionData {
		t.Fatal("TDMSection value comparison failed")
	}

	strictFirst, strictSecond := StrictTDMPolicy(), StrictTDMPolicy()
	if strictFirst != strictSecond {
		t.Fatalf("independent strict policies differ: %#v %#v", strictFirst, strictSecond)
	}
	changedPolicy := strictSecond
	changedPolicy.LongLines = TDMLeniencyForgive
	if strictFirst == changedPolicy {
		t.Fatal("unequal TDMPolicy values compared equal")
	}

	writeFirst, writeSecond := LenientTDMWritePolicy(), LenientTDMWritePolicy()
	if writeFirst != writeSecond {
		t.Fatalf("independent lenient write policies differ: %#v %#v", writeFirst, writeSecond)
	}
	changedWritePolicy := writeSecond
	changedWritePolicy.RepeatedKeywords = TDMLeniencyStrict
	if writeFirst == changedWritePolicy {
		t.Fatal("unequal TDMWritePolicy values compared equal")
	}

	first := parseEqualityTDM(t, tdmExpectedKVN)
	second := parseEqualityTDM(t, tdmExpectedKVN)
	firstHeader, err := first.Header()
	if err != nil {
		t.Fatal(err)
	}
	secondHeader, err := second.Header()
	if err != nil {
		t.Fatal(err)
	}
	if firstHeader.CreationDate != secondHeader.CreationDate || firstHeader.CreationDate == (TDMStringField{}) {
		t.Fatalf("independent TDMStringField values differ: %#v %#v", firstHeader.CreationDate, secondHeader.CreationDate)
	}
	changedString := secondHeader.CreationDate
	changedString.Value = "different"
	if firstHeader.CreationDate == changedString {
		t.Fatal("unequal TDMStringField values compared equal")
	}

	firstComments, err := first.Comments()
	if err != nil {
		t.Fatal(err)
	}
	secondComments, err := second.Comments()
	if err != nil {
		t.Fatal(err)
	}
	if firstComments[0] != secondComments[0] {
		t.Fatalf("independent TDMComment values differ: %#v %#v", firstComments[0], secondComments[0])
	}
	changedComment := secondComments[0]
	changedComment.Text = "different"
	if firstComments[0] == changedComment {
		t.Fatal("unequal TDMComment values compared equal")
	}

	firstFields, err := first.Fields()
	if err != nil {
		t.Fatal(err)
	}
	secondFields, err := second.Fields()
	if err != nil {
		t.Fatal(err)
	}
	if firstFields[0] != secondFields[0] {
		t.Fatalf("independent TDMField values differ: %#v %#v", firstFields[0], secondFields[0])
	}
	changedField := secondFields[0]
	changedField.Value = "TAI"
	if firstFields[0] == changedField {
		t.Fatal("unequal TDMField values compared equal")
	}

	firstParticipants, err := first.Participants()
	if err != nil {
		t.Fatal(err)
	}
	secondParticipants, err := second.Participants()
	if err != nil {
		t.Fatal(err)
	}
	if firstParticipants[0] != secondParticipants[0] {
		t.Fatalf("independent TDMParticipant values differ: %#v %#v", firstParticipants[0], secondParticipants[0])
	}
	changedParticipant := secondParticipants[0]
	changedParticipant.Name = "different"
	if firstParticipants[0] == changedParticipant {
		t.Fatal("unequal TDMParticipant values compared equal")
	}

	firstRecords, err := first.Records()
	if err != nil {
		t.Fatal(err)
	}
	secondRecords, err := second.Records()
	if err != nil {
		t.Fatal(err)
	}
	if firstRecords[0] != secondRecords[0] {
		t.Fatalf("independent TDMRecord values differ: %#v %#v", firstRecords[0], secondRecords[0])
	}
	changedRecord := secondRecords[0]
	changedRecord.Value = 999
	if firstRecords[0] == changedRecord {
		t.Fatal("unequal TDMRecord values compared equal")
	}

	firstSegments, err := first.Segments()
	if err != nil {
		t.Fatal(err)
	}
	secondSegments, err := second.Segments()
	if err != nil {
		t.Fatal(err)
	}
	if firstSegments[0] != secondSegments[0] {
		t.Fatalf("independent TDMSegmentSummary values differ: %#v %#v", firstSegments[0], secondSegments[0])
	}
	changedSegment := secondSegments[0]
	changedSegment.RecordCount++
	if firstSegments[0] == changedSegment {
		t.Fatal("unequal TDMSegmentSummary values compared equal")
	}

	commentInput := TDMCommentInput{BeforeRecord: 1, Text: "comment"}
	if commentInput != (TDMCommentInput{BeforeRecord: 1, Text: "comment"}) || commentInput == (TDMCommentInput{Text: "different"}) {
		t.Fatal("TDMCommentInput value comparison failed")
	}
	fieldInput := TDMFieldInput{Key: "MODE", Value: "SEQUENTIAL"}
	if fieldInput != (TDMFieldInput{Key: "MODE", Value: "SEQUENTIAL"}) || fieldInput == (TDMFieldInput{Key: "MODE", Value: "SINGLE_DIFF"}) {
		t.Fatal("TDMFieldInput value comparison failed")
	}
	headerValues := TDMHeaderValues{Version: "2.0", Originator: TDMStringField{Present: true, Value: "TEST"}}
	if headerValues != (TDMHeaderValues{Version: "2.0", Originator: TDMStringField{Present: true, Value: "TEST"}}) || headerValues == (TDMHeaderValues{Version: "1.0"}) {
		t.Fatal("TDMHeaderValues value comparison failed")
	}
	recordInput := testTDMRecordInput()
	if recordInput != testTDMRecordInput() {
		t.Fatal("equal TDMRecordInput values differ")
	}
	changedRecordInput := recordInput
	changedRecordInput.ValueText = "999"
	if recordInput == changedRecordInput {
		t.Fatal("unequal TDMRecordInput values compared equal")
	}
}

func TestTDMExplicitValueEquality(t *testing.T) {
	first := parseEqualityTDM(t, tdmExpectedKVN)
	second := parseEqualityTDM(t, tdmExpectedKVN)

	firstHeader, err := first.Header()
	if err != nil {
		t.Fatal(err)
	}
	secondHeader, err := second.Header()
	if err != nil {
		t.Fatal(err)
	}
	if !firstHeader.Equal(secondHeader) {
		t.Fatalf("independent equal headers differ: %#v %#v", firstHeader, secondHeader)
	}
	secondHeader.SegmentCount++
	if firstHeader.Equal(secondHeader) {
		t.Fatal("unequal headers compared equal")
	}

	firstPaths, err := first.Paths()
	if err != nil {
		t.Fatal(err)
	}
	secondPaths, err := second.Paths()
	if err != nil {
		t.Fatal(err)
	}
	if !firstPaths[0].Equal(secondPaths[0]) {
		t.Fatalf("independent equal paths differ: %#v %#v", firstPaths[0], secondPaths[0])
	}
	secondPaths[0].Participants[0]++
	if firstPaths[0].Equal(secondPaths[0]) {
		t.Fatal("unequal paths compared equal")
	}

	unterminated := []byte(tdmExpectedKVN[:len(tdmExpectedKVN)-1])
	warningTDMFirst, warningsFirst, err := ParseTDMKVNWithPolicy(unterminated, LenientTDMPolicy())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := warningTDMFirst.Close(); err != nil {
			t.Error(err)
		}
	})
	warningTDMSecond, warningsSecond, err := ParseTDMKVNWithPolicy(unterminated, LenientTDMPolicy())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := warningTDMSecond.Close(); err != nil {
			t.Error(err)
		}
	})
	if !warningsFirst[0].Equal(warningsSecond[0]) {
		t.Fatalf("independent equal warnings differ: %#v %#v", warningsFirst[0], warningsSecond[0])
	}
	changedWarning := warningsSecond[0]
	changedWarning.Message += " different"
	if warningsFirst[0].Equal(changedWarning) {
		t.Fatal("unequal warnings compared equal")
	}

	departureFirst := testMissingKeywordDeparture(t)
	departureSecond := testMissingKeywordDeparture(t)
	if !departureFirst.Equal(departureSecond) {
		t.Fatalf("independent equal departures differ: %#v %#v", departureFirst, departureSecond)
	}
	changedDeparture := departureSecond
	changedDeparture.Kind = "different"
	if departureFirst.Equal(changedDeparture) {
		t.Fatal("unequal departures compared equal")
	}
}

func TestTDMDetailFieldEqualityIgnoresEmptyContainers(t *testing.T) {
	withNil := TDMWarning{Kind: "UNTERMINATED_FINAL_LINE", Message: "text"}
	withEmpty := TDMWarning{Kind: "UNTERMINATED_FINAL_LINE", Fields: map[string]EngineJSONValue{}, Message: "text"}
	if !withNil.Equal(withEmpty) {
		t.Fatal("nil and empty warning fields compared unequal")
	}

	nested := func(array []EngineJSONValue) map[string]EngineJSONValue {
		return map[string]EngineJSONValue{
			"line":  {Kind: EngineJSONNumber, Number: "54"},
			"items": {Kind: EngineJSONArray, Array: array},
		}
	}
	first := TDMDeparture{Kind: "K", Fields: nested(nil), Message: "m"}
	second := TDMDeparture{Kind: "K", Fields: nested([]EngineJSONValue{}), Message: "m"}
	if !first.Equal(second) {
		t.Fatal("nil and empty nested arrays compared unequal")
	}
	third := TDMDeparture{Kind: "K", Fields: nested([]EngineJSONValue{{Kind: EngineJSONBoolean, Bool: true}}), Message: "m"}
	if first.Equal(third) {
		t.Fatal("different nested arrays compared equal")
	}
	token := TDMDeparture{Kind: "K", Fields: map[string]EngineJSONValue{"line": {Kind: EngineJSONNumber, Number: "54.0"}}, Message: "m"}
	plain := TDMDeparture{Kind: "K", Fields: map[string]EngineJSONValue{"line": {Kind: EngineJSONNumber, Number: "54"}}, Message: "m"}
	if token.Equal(plain) {
		t.Fatal("different number tokens compared equal")
	}

	detail := TDMErrorDetail{Kind: "K", OperationClass: "parse", Display: "d"}
	if !detail.Equal(TDMErrorDetail{Kind: "K", Fields: map[string]EngineJSONValue{}, OperationClass: "parse", Display: "d"}) {
		t.Fatal("nil and empty error detail fields compared unequal")
	}
}

func testMissingKeywordDeparture(t *testing.T) TDMDeparture {
	t.Helper()
	value := buildTestTDM(t)
	departures, err := value.ReplaceMetadata(
		0,
		[]TDMFieldInput{{Key: "PARTICIPANT_1", Value: "DSS-55"}},
		nil,
		LenientTDMWritePolicy(),
	)
	if err != nil || len(departures) != 1 {
		t.Fatalf("ReplaceMetadata departures = %#v, %v", departures, err)
	}
	return departures[0]
}

func TestTDMErrorDetailEqual(t *testing.T) {
	first := parseTDMErrorDetail(t, []byte("NOT A LINE\n"))
	second := parseTDMErrorDetail(t, []byte("NOT A LINE\n"))
	if !first.Equal(second) {
		t.Fatalf("independent equal details differ: %#v %#v", first, second)
	}
	second.Kind = "different"
	if first.Equal(second) {
		t.Fatal("unequal error details compared equal")
	}
}

func parseTDMErrorDetail(t *testing.T, data []byte) TDMErrorDetail {
	t.Helper()
	_, err := ParseTDMKVN(data)
	if err == nil {
		t.Fatal("invalid TDM parsed successfully")
	}
	var detail *TDMErrorDetail
	if !errors.As(err, &detail) {
		t.Fatalf("error detail = %T, %v", err, err)
	}
	return *detail
}
