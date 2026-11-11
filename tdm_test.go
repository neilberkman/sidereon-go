//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

const tdmExpectedKVN = "CCSDS_TDM_VERS = 2.0\nCOMMENT header comment\nCREATION_DATE = 2026-001T00:00:00\nORIGINATOR = TEST\nMETA_START\nCOMMENT metadata comment\nTIME_SYSTEM = UTC\nPARTICIPANT_1 = DSS-55\nMODE = SEQUENTIAL\nPATH = 1\nMETA_STOP\nDATA_START\nCOMMENT data comment\nRANGE = 2026-001T00:00:00 123.5\nDATA_STOP\n"

func testTDMRecordInput() TDMRecordInput {
	return TDMRecordInput{Observable: TDMObservableRange, Unit: TDMUnitKilometers, UnitLabel: "km", Keyword: "RANGE", Epoch: "2026-001T00:00:00", ValueText: "123.5", Value: 123.5}
}

func buildTestTDM(t *testing.T) *TDM {
	t.Helper()
	tdm, err := NewTDM("2.0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := tdm.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := tdm.SetHeader(TDMHeaderValues{
		Version:      "2.0",
		CreationDate: TDMStringField{Present: true, Value: "2026-001T00:00:00"},
		Originator:   TDMStringField{Present: true, Value: "TEST"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := tdm.ReplaceHeaderRaw([]TDMCommentInput{{Text: "header comment", BeforeRecord: 1}}, nil); err != nil {
		t.Fatal(err)
	}
	fields := []TDMFieldInput{{Key: "TIME_SYSTEM", Value: "UTC"}, {Key: "PARTICIPANT_1", Value: "DSS-55"}, {Key: "MODE", Value: "SEQUENTIAL"}, {Key: "PATH", Value: "1"}}
	departures, err := tdm.AppendSegment(fields, []TDMCommentInput{{Text: "metadata comment"}}, []TDMCommentInput{{Text: "data comment"}}, []TDMRecordInput{testTDMRecordInput()}, StrictTDMWritePolicy())
	if err != nil || len(departures) != 0 {
		t.Fatalf("AppendSegment = %#v, %v", departures, err)
	}
	return tdm
}

func TestTDMPolicies(t *testing.T) {
	strict, def, lenient := StrictTDMPolicy(), DefaultTDMPolicy(), LenientTDMPolicy()
	if strict != def {
		t.Fatalf("strict != default: %#v %#v", strict, def)
	}
	if strict != (TDMPolicy{}) {
		t.Fatalf("strict = %#v", strict)
	}
	if lenient != (TDMPolicy{1, 1, 1, 1, 1, 1, 1, 1}) {
		t.Fatalf("lenient = %#v", lenient)
	}
	ws, wd, wl := StrictTDMWritePolicy(), DefaultTDMWritePolicy(), LenientTDMWritePolicy()
	if ws != wd || ws != (TDMWritePolicy{}) {
		t.Fatalf("write strict/default = %#v %#v", ws, wd)
	}
	if wl != (TDMWritePolicy{1, 1, 1, 1, 1, 1, 1, 1, 1}) {
		t.Fatalf("write lenient = %#v", wl)
	}
	read, err := wl.AsRead()
	if err != nil || read != lenient {
		t.Fatalf("AsRead = %#v, %v", read, err)
	}
	bad := ws
	bad.KeywordOrder = TDMLeniency(77)
	_, err = bad.AsRead()
	assertTDMError(t, err, "invalid_policy", "validation")
}

func TestTDMBuildEncodeParseAndAccess(t *testing.T) {
	tdm := buildTestTDM(t)
	encoded, err := tdm.ToKVN()
	if err != nil || string(encoded) != tdmExpectedKVN {
		t.Fatalf("KVN = %q, %v", encoded, err)
	}
	alias, err := tdm.ToKVN()
	if err != nil || string(alias) != tdmExpectedKVN {
		t.Fatalf("ToKVN = %q, %v", alias, err)
	}
	policyText, departures, err := tdm.ToKVNWithPolicy(StrictTDMWritePolicy())
	if err != nil || string(policyText) != tdmExpectedKVN || len(departures) != 0 {
		t.Fatalf("KVNWithPolicy = %q, %#v, %v", policyText, departures, err)
	}
	parsed, err := ParseTDMKVN(encoded)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := parsed.Close(); err != nil {
			t.Error(err)
		}
	})
	header, err := parsed.Header()
	if err != nil || header.Version != "2.0" || header.CreationDate != (TDMStringField{Present: true, Value: "2026-001T00:00:00"}) || header.Originator != (TDMStringField{Present: true, Value: "TEST"}) || header.MessageID != (TDMStringField{}) || header.SegmentCount != 1 {
		t.Fatalf("Header = %#v, %v", header, err)
	}
	if !reflect.DeepEqual(header.Comments, []TDMComment{{Section: TDMSectionHeader, BeforeRecord: 1, Text: "header comment"}}) {
		t.Fatalf("header comments = %#v", header.Comments)
	}
	comments, err := parsed.Comments()
	wantComments := []TDMComment{{Section: TDMSectionHeader, BeforeRecord: 1, Text: "header comment"}, {Section: TDMSectionMetadata, Text: "metadata comment"}, {Section: TDMSectionData, Text: "data comment"}}
	if err != nil || !reflect.DeepEqual(comments, wantComments) {
		t.Fatalf("Comments = %#v, %v", comments, err)
	}
	fields, err := parsed.Fields()
	wantFields := []TDMField{{Section: TDMSectionMetadata, Key: "TIME_SYSTEM", Value: "UTC"}, {Section: TDMSectionMetadata, Key: "PARTICIPANT_1", Value: "DSS-55"}, {Section: TDMSectionMetadata, Key: "MODE", Value: "SEQUENTIAL"}, {Section: TDMSectionMetadata, Key: "PATH", Value: "1"}}
	if err != nil || !reflect.DeepEqual(fields, wantFields) || len(header.Fields) != 0 {
		t.Fatalf("Fields = %#v header=%#v, %v", fields, header.Fields, err)
	}
	participants, err := parsed.Participants()
	if err != nil || !reflect.DeepEqual(participants, []TDMParticipant{{Index: 1, Name: "DSS-55"}}) {
		t.Fatalf("Participants = %#v, %v", participants, err)
	}
	paths, err := parsed.Paths()
	if err != nil || !reflect.DeepEqual(paths, []TDMPath{{Key: "PATH", Participants: []uint8{1}}}) {
		t.Fatalf("Paths = %#v, %v", paths, err)
	}
	segments, err := parsed.Segments()
	wantSegments := []TDMSegmentSummary{{Mode: TDMStringField{true, "SEQUENTIAL"}, TimeSystem: TDMStringField{true, "UTC"}, RangeUnit: TDMUnitKilometers, ParticipantCount: 1, PathCount: 1, RecordCount: 1}}
	if err != nil || !reflect.DeepEqual(segments, wantSegments) {
		t.Fatalf("Segments = %#v, %v", segments, err)
	}
	if n, err := parsed.SegmentCount(); err != nil || n != 1 {
		t.Fatalf("SegmentCount = %d, %v", n, err)
	}
	records, err := parsed.Records()
	recordInput := testTDMRecordInput()
	wantRecord := TDMRecord{
		Observable:               recordInput.Observable,
		HasObservableParticipant: recordInput.HasObservableParticipant,
		ObservableParticipant:    recordInput.ObservableParticipant,
		ObservableName:           recordInput.ObservableName,
		Unit:                     recordInput.Unit,
		UnitLabel:                recordInput.UnitLabel,
		Keyword:                  recordInput.Keyword,
		Epoch:                    recordInput.Epoch,
		ValueText:                recordInput.ValueText,
		Value:                    recordInput.Value,
	}
	if err != nil || !reflect.DeepEqual(records, []TDMRecord{wantRecord}) {
		t.Fatalf("Records = %#v, %v", records, err)
	}
	if n, err := parsed.RecordCount(); err != nil || n != 1 {
		t.Fatalf("RecordCount = %d, %v", n, err)
	}
}

func TestTDMEditingLongTextAndDepartures(t *testing.T) {
	tdm := buildTestTDM(t)
	longComment, longValue := strings.Repeat("C", 200), strings.Repeat("P", 200)
	fields := []TDMFieldInput{{Key: "TIME_SYSTEM", Value: "UTC"}, {Key: "PARTICIPANT_1", Value: longValue}, {Key: "MODE", Value: "SEQUENTIAL"}, {Key: "PATH", Value: "1"}}
	d, err := tdm.ReplaceMetadata(0, fields, []TDMCommentInput{{Text: longComment}}, StrictTDMWritePolicy())
	if err != nil || len(d) != 0 {
		t.Fatalf("ReplaceMetadata = %#v, %v", d, err)
	}
	text, err := tdm.ToKVN()
	if err != nil || !strings.Contains(string(text), longComment) || !strings.Contains(string(text), longValue) {
		t.Fatalf("long KVN = %q, %v", text, err)
	}
	round, err := ParseTDMKVN(text)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := round.Close(); err != nil {
			t.Error(err)
		}
	})
	comments, _ := round.Comments()
	gotFields, _ := round.Fields()
	if comments[1].Text != longComment || gotFields[1].Value != longValue {
		t.Fatalf("long round trip = %#v %#v", comments, gotFields)
	}
	if err := round.RemoveRecord(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := round.InsertRecord(0, 0, testTDMRecordInput()); err != nil {
		t.Fatal(err)
	}
	if err := round.ReplaceData(0, []TDMCommentInput{{Text: "data comment"}}, []TDMRecordInput{testTDMRecordInput()}); err != nil {
		t.Fatal(err)
	}
	missing := []TDMFieldInput{{Key: "PARTICIPANT_1", Value: "DSS-55"}}
	d, err = tdm.ReplaceMetadata(0, missing, nil, LenientTDMWritePolicy())
	if err != nil || len(d) == 0 || d[0].Kind != "missing_keyword" || d[0].Fields["keyword"].String != "TIME_SYSTEM" || d[0].Fields["segment"].Number.String() != "1" || d[0].Message == "" {
		t.Fatalf("lenient departure = %#v, %v", d, err)
	}
	_, err = buildTestTDMNoCleanup(t).ReplaceMetadata(0, missing, nil, StrictTDMWritePolicy())
	assertTDMError(t, err, "missing_keyword", "validation")
	if err := round.RemoveSegment(0); err != nil {
		t.Fatal(err)
	}
	if err := round.RemoveSegment(0); err == nil {
		t.Fatal("second RemoveSegment succeeded")
	}
}

func buildTestTDMNoCleanup(t *testing.T) *TDM { tdm := buildTestTDM(t); return tdm }

func TestTDMPolicyWarningsAndInvalidCodes(t *testing.T) {
	input := []byte(strings.TrimSuffix(tdmExpectedKVN, "\n"))
	tdm, warnings, err := ParseTDMKVNWithPolicy(input, LenientTDMPolicy())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := tdm.Close(); err != nil {
			t.Error(err)
		}
	})
	if len(warnings) != 1 || warnings[0].Kind != "unterminated_final_line" || warnings[0].Fields["line"].Number.String() == "" || warnings[0].Message == "" {
		t.Fatalf("warnings = %#v", warnings)
	}
	bad := DefaultTDMPolicy()
	bad.RecordOrder = TDMLeniency(99)
	value, warnings, err := ParseTDMKVNWithPolicy([]byte(tdmExpectedKVN), bad)
	if value != nil || warnings != nil {
		t.Fatalf("invalid policy outputs = %#v %#v", value, warnings)
	}
	assertTDMError(t, err, "invalid_policy", "parse")
}

func TestTDMClosed(t *testing.T) {
	tdm := buildTestTDM(t)
	if err := tdm.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tdm.ToKVN(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed error = %v", err)
	}
	if err := tdm.Close(); err != nil {
		t.Fatal(err)
	}
}

func assertTDMError(t *testing.T, err error, kind, operationClass string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	var status *StatusError
	if !errors.As(err, &status) {
		t.Fatalf("error type = %T: %v", err, err)
	}
	var detail *TDMErrorDetail
	if !errors.As(err, &detail) || detail.Kind != kind {
		t.Fatalf("TDM detail = %#v, error=%v", detail, err)
	}
	if detail.OperationClass != operationClass {
		t.Fatalf("operation_class = %q", detail.OperationClass)
	}
}
