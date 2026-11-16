package sidereon

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCelestrakCatalog(t *testing.T) {
	entry, err := CelestrakGPSourceEntry()
	if err != nil || entry != (CatalogSourceEntry{Protocol: "https", Host: "celestrak.org", RootURL: "https://celestrak.org/NORAD/elements/gp.php"}) {
		t.Fatalf("source: %+v %v", entry, err)
	}
	navEntry, err := NAVCENSourceEntry()
	if err != nil || navEntry != (CatalogSourceEntry{Protocol: "https", Host: "www.navcen.uscg.gov", RootURL: "https://www.navcen.uscg.gov/gps-constellation"}) {
		t.Fatalf("NAVCEN source: %+v %v", navEntry, err)
	}
	hosts, err := AllowedHosts()
	if err != nil {
		t.Fatal(err)
	}
	wantHosts := []string{"www.aiub.unibe.ch", "download.aiub.unibe.ch", "zhw-b.s3.cloud.switch.ch", "navigation-office.esa.int", "isdc-data.gfz.de", "igs.bkg.bund.de", "igs.gnsswhu.cn", "s3.amazonaws.com", "celestrak.org", "cddis.nasa.gov", "urs.earthdata.nasa.gov", "www.navcen.uscg.gov"}
	if !reflect.DeepEqual(hosts, wantHosts) {
		t.Fatalf("allowed hosts: %v", hosts)
	}
	group, err := CelestrakGroup(GNSSSystemSBAS)
	if err != nil || group != "sbas" {
		t.Fatalf("SBAS group: %q %v", group, err)
	}
	for _, tc := range []struct {
		system GNSSSystem
		group  string
	}{{GNSSSystemGPS, "gps-ops"}, {GNSSSystemGalileo, "galileo"}, {GNSSSystemGLONASS, "glo-ops"}, {GNSSSystemBeiDou, "beidou"}, {GNSSSystemQZSS, "gnss"}, {GNSSSystemNavIC, "gnss"}, {GNSSSystemSBAS, "sbas"}} {
		got, err := CelestrakGroup(tc.system)
		if err != nil || got != tc.group {
			t.Fatalf("group %d: %q %v", tc.system, got, err)
		}
		url, err := CelestrakConstellationURL(tc.system, CelestrakGPFormatJSON)
		want := "https://celestrak.org/NORAD/elements/gp.php?GROUP=" + tc.group + "&FORMAT=JSON"
		if err != nil || url != want {
			t.Fatalf("constellation %d: %q %v", tc.system, url, err)
		}
	}
	gps, err := CelestrakConstellationURL(GNSSSystemGPS, CelestrakGPFormatJSON)
	if err != nil || gps != "https://celestrak.org/NORAD/elements/gp.php?GROUP=gps-ops&FORMAT=JSON" {
		t.Fatalf("GPS URL: %q %v", gps, err)
	}
	nav, err := NAVCENConstellationURL()
	if err != nil || nav != "https://www.navcen.uscg.gov/gps-constellation" {
		t.Fatalf("NAVCEN: %q %v", nav, err)
	}
	formats := []string{"TLE", "3LE", "2LE", "XML", "KVN", "JSON", "JSON-PRETTY", "CSV"}
	for i, format := range formats {
		wire, err := CelestrakGPFormat(i).AsStr()
		if err != nil || wire != format {
			t.Fatalf("format token: %q %v", wire, err)
		}
		got, err := CelestrakGPURL(CelestrakGPCatalogNumber(25544), CelestrakGPFormat(i))
		want := "https://celestrak.org/NORAD/elements/gp.php?CATNR=25544&FORMAT=" + format
		if err != nil || got != want {
			t.Fatalf("format %s: %q %v", format, got, err)
		}
	}
	for _, tc := range []struct {
		selector CelestrakGPQuery
		query    string
	}{
		{CelestrakGPGroup("gps-ops"), "GROUP=gps-ops"},
		{CelestrakGPInternationalDesignator("2020-025"), "INTDES=2020-025"},
		{CelestrakGPName("COSMOS 2251 DEB"), "NAME=COSMOS%202251%20DEB"},
	} {
		got, err := CelestrakGPURL(tc.selector, CelestrakGPFormatCSV)
		want := "https://celestrak.org/NORAD/elements/gp.php?" + tc.query + "&FORMAT=CSV"
		if err != nil || got != want {
			t.Fatalf("selector %s: %q %v", tc.query, got, err)
		}
	}
	for i, query := range []string{"GPZ", "GPZ-PLUS", "DECAYING"} {
		wire, err := CelestrakGPSpecial(i).AsStr()
		if err != nil || wire != query {
			t.Fatalf("special token: %q %v", wire, err)
		}
		got, err := CelestrakGPURL(CelestrakGPSpecial(i), CelestrakGPFormatCSV)
		want := "https://celestrak.org/NORAD/elements/gp.php?SPECIAL=" + query + "&FORMAT=CSV"
		if err != nil || got != want {
			t.Fatalf("special %q: %q %v", query, got, err)
		}
	}
	_, err = CelestrakGPURL(CelestrakGPGroup("gps&ops"), CelestrakGPFormatJSON)
	var detail *DataCatalogError
	if !errors.As(err, &detail) || detail.Kind != "invalid_celestrak_gp_selector" || detail.Selector != "GROUP" || detail.Value != "gps&ops" {
		t.Fatalf("invalid group: %#v", err)
	}
	if !detail.Equal(&DataCatalogError{Kind: "invalid_celestrak_gp_selector", Selector: "GROUP", Value: "gps&ops"}) || detail.Equal(&DataCatalogError{Kind: "invalid_celestrak_gp_selector", Selector: "GROUP", Value: "other"}) {
		t.Fatalf("catalog error equality: %+v", detail)
	}
	_, err = CelestrakGPURL(CelestrakGPCatalogNumber(4294967296), CelestrakGPFormatJSON)
	if !errors.As(err, &detail) || detail.Kind != "invalid_celestrak_gp_selector" || detail.Selector != "CATNR" || detail.Value != "4294967296" {
		t.Fatalf("invalid catalog number: %#v", err)
	}
	for _, tc := range []struct {
		number CelestrakGPCatalogNumber
		value  string
	}{{0, "0"}, {1000000000, "1000000000"}} {
		_, err := CelestrakGPURL(tc.number, CelestrakGPFormatJSON)
		if !errors.As(err, &detail) || detail.Kind != "invalid_celestrak_gp_selector" || detail.Selector != "CATNR" || detail.Value != tc.value {
			t.Fatalf("invalid catalog number %s: %v", tc.value, err)
		}
	}
	for _, tc := range []struct {
		query  CelestrakGPQuery
		format CelestrakGPFormat
	}{{CelestrakGPGroup("gps-ops"), CelestrakGPFormat(99)}, {CelestrakGPSpecial(99), CelestrakGPFormatJSON}} {
		_, err := CelestrakGPURL(tc.query, tc.format)
		var status *StatusError
		if !errors.As(err, &status) || status.Code != StatusInvalidArgument {
			t.Fatalf("invalid enum: %v", err)
		}
	}
	for _, call := range []func() (string, error){CelestrakGPFormat(99).AsStr, CelestrakGPSpecial(99).AsStr, func() (string, error) { return CelestrakGPURL(nil, CelestrakGPFormatJSON) }} {
		_, err := call()
		var status *StatusError
		if !errors.As(err, &status) || status.Code != StatusInvalidArgument {
			t.Fatalf("invalid selector: %v", err)
		}
	}
}

func TestCelestrakGPResponse(t *testing.T) {
	data := []byte("ISS (ZARYA)             \r\n1 25544U 98067A   26095.55331950  .00008543  00000+0  16420-3 0  9998\r\n2 25544  51.6328 299.5432 0006351 274.8255  85.2008 15.48786980560492\r\n")
	response, err := ParseCelestrakGPResponse(data, CelestrakGPFormatTLE)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "response", response.Close)
	accepted, rejected, err := response.Counts()
	if err != nil || accepted != 1 || rejected != 0 {
		t.Fatalf("counts %d %d %v", accepted, rejected, err)
	}
	for _, index := range []int{-1, 1} {
		_, err := response.Record(index)
		var status *StatusError
		if !errors.As(err, &status) || status.Code != StatusInvalidArgument {
			t.Fatalf("record index %d: %v", index, err)
		}
	}
	_, err = response.RejectedRecord(0)
	var missingStatus *StatusError
	if !errors.As(err, &missingStatus) || missingStatus.Code != StatusInvalidArgument {
		t.Fatalf("rejected index: %v", err)
	}
	record, err := response.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	recordData, err := record.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if recordData.RecordIndex != 1 || recordData.LineNumber == nil || *recordData.LineNumber != 1 || recordData.ObjectName == nil || *recordData.ObjectName != "ISS (ZARYA)" {
		t.Fatalf("record %+v", record)
	}
	if recordData.Payload.Kind != "tle" || recordData.Payload.TLE == nil || recordData.Payload.TLE.Elements.CatalogNumber != "25544" || recordData.Payload.TLE.Elements.EpochYear != 2026 || recordData.Payload.TLE.BStarText == nil || *recordData.Payload.TLE.BStarText != " 16420-3" || len(recordData.Payload.TLE.ChecksumWarnings) != 0 {
		t.Fatalf("TLE payload: %+v", recordData.Payload)
	}
	wantTLE := CelestrakTLERecord{Elements: TLEMetadata{CatalogNumber: "25544", Classification: "U", InternationalDesignator: "98067A", EpochYear: 2026, EpochDayOfYear: 95.5533195, InclinationDeg: 51.6328, RAANDeg: 299.5432, Eccentricity: 0.0006351, ArgumentOfPerigeeDeg: 274.8255, MeanAnomalyDeg: 85.2008, MeanMotionRevPerDay: 15.4878698, MeanMotionDot: 0.00008543, MeanMotionDoubleDot: 0, BStar: 0.0001642, HasEphemerisType: true, EphemerisType: 0, HasElementSetNumber: true, ElementSetNumber: 999, HasRevolutionNumber: true, RevolutionNumber: 56049}, Line1: "1 25544U 98067A   26095.55331950  .00008543  00000+0  16420-3 0  9998", Line2: "2 25544  51.6328 299.5432 0006351 274.8255  85.2008 15.48786980560492", MeanMotionDoubleDotText: fixturePtr(" 00000+0"), BStarText: fixturePtr(" 16420-3"), ChecksumWarnings: []CelestrakChecksumWarning{}}
	if !reflect.DeepEqual(*recordData.Payload.TLE, wantTLE) {
		t.Fatalf("detached TLE: %+v", recordData.Payload.TLE)
	}
	if string(recordData.Raw) != string(data) {
		t.Fatalf("raw %q", recordData.Raw)
	}
	if equal, err := record.Equal(record); err != nil || !equal {
		t.Fatalf("record equality: %v %v", equal, err)
	}
	if equal, err := response.Equal(response); err != nil || !equal {
		t.Fatalf("response equality: %v %v", equal, err)
	}
	otherData := []byte(strings.Replace(string(data), "ISS (ZARYA)", "ISS (OTHER)", 1))
	other, err := ParseCelestrakGPResponse(otherData, CelestrakGPFormatTLE)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "other response", other.Close)
	otherRecord, err := other.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	if equal, err := record.Equal(otherRecord); err != nil || equal {
		t.Fatalf("different record: %v %v", equal, err)
	}
	if equal, err := record.PayloadEqual(otherRecord); err != nil || !equal {
		t.Fatalf("payload equality: %v %v", equal, err)
	}
	if equal, err := record.TLEEqual(otherRecord); err != nil || !equal {
		t.Fatalf("TLE equality: %v %v", equal, err)
	}
	differentTLE, err := os.ReadFile("testdata/celestrak/24876.tle")
	if err != nil {
		t.Fatal(err)
	}
	differentResponse, err := ParseCelestrakGPResponse(differentTLE, CelestrakGPFormatTLE)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "different response", differentResponse.Close)
	differentRecord, err := differentResponse.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	if equal, err := record.PayloadEqual(differentRecord); err != nil || equal {
		t.Fatalf("different payload: %v %v", equal, err)
	}
	if equal, err := record.TLEEqual(differentRecord); err != nil || equal {
		t.Fatalf("different TLE: %v %v", equal, err)
	}
	if equal, err := response.Equal(other); err != nil || equal {
		t.Fatalf("different response: %v %v", equal, err)
	}
	omm, err := record.ToOMM()
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "record OMM", omm.Close)
	snapshot, err := omm.Snapshot()
	if err != nil || snapshot.ObjectID == nil || *snapshot.ObjectID != "1998-067A" {
		t.Fatalf("record OMM: %+v %v", snapshot, err)
	}
	elements, err := record.ToElementSet()
	wantElements := OMMElementSet{Epoch: JulianDate{Whole: 2461135.5, Fraction: 0.5533195}, BStar: 0.0001642, MeanMotionDot: fixturePtr(0.00008543), MeanMotionDoubleDot: fixturePtr(0.0), Eccentricity: 0.0006351, ArgumentOfPerigeeDeg: 274.8255, InclinationDeg: 51.6328, MeanAnomalyDeg: 85.2008, MeanMotionRevPerDay: 15.4878698, RightAscensionDeg: 299.5432, CatalogNumber: fixturePtr(uint32(25544))}
	if err != nil || !reflect.DeepEqual(elements, wantElements) {
		t.Fatalf("record elements: %+v %v", elements, err)
	}
	tle, err := record.TLE()
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "record TLE", tle.Close)
	lines, err := tle.Lines()
	if err != nil || lines.Line1 != "1 25544U 98067A   26095.55331950  .00008543  00000+0  16420-3 0  9998" || lines.Line2 != "2 25544  51.6328 299.5432 0006351 274.8255  85.2008 15.48786980560492" {
		t.Fatalf("record TLE lines: %+v %v", lines, err)
	}
	for _, load := range []func() (*TLE, error){func() (*TLE, error) { return record.TLEWithOpsMode(OpsModeImproved) }, func() (*TLE, error) { return record.TLEWithPolicy(TLEFilePolicyLenient) }, func() (*TLE, error) { return record.TLEWithOpsModeAndPolicy(OpsModeImproved, TLEFilePolicyLenient) }} {
		loaded, err := load()
		if err != nil {
			t.Fatal(err)
		}
		cleanupClose(t, "loaded TLE", loaded.Close)
	}
	for _, load := range []func() (*TLE, error){func() (*TLE, error) { return record.TLEWithOpsMode(OpsMode(99)) }, func() (*TLE, error) { return record.TLEWithPolicy(TLEFilePolicy(99)) }} {
		_, err := load()
		var status *StatusError
		if !errors.As(err, &status) || status.Code != StatusInvalidArgument {
			t.Fatalf("invalid TLE option: %v", err)
		}
	}
}

func TestOMMRecordSpanValue(t *testing.T) {
	const document = `[{"OBJECT_NAME":"ISS","OBJECT_ID":"1998-067A","EPOCH":"2026-06-17T04:32:52.099296","MEAN_MOTION":15.49273435,"ECCENTRICITY":0.0004737,"INCLINATION":51.6332,"RA_OF_ASC_NODE":300.0813,"ARG_OF_PERICENTER":195.1146,"MEAN_ANOMALY":164.9702}]`
	array, err := ParseOMMJSONArray([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "array", array.Close)
	spans, err := array.Spans()
	if err != nil {
		t.Fatal(err)
	}
	want := OMMRecordSpan{Index: 0, ByteStart: 1, ByteEnd: uint64(len(document) - 1), Line: 1}
	if len(spans) != 1 || spans[0] != want {
		t.Fatalf("spans %+v, want %+v", spans, want)
	}
	if equal, err := array.Equal(array); err != nil || !equal {
		t.Fatalf("array equality: %v %v", equal, err)
	}
	empty, err := NewOMMArray()
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "empty OMM array", empty.Close)
	if equal, err := array.Equal(empty); err != nil || equal {
		t.Fatalf("different arrays: %v %v", equal, err)
	}
}

func TestCelestrakResponseErrors(t *testing.T) {
	cases := []struct{ data, kind, query, time string }{
		{"No GP data found", "celestrak_gp_data_not_found", "", ""},
		{"GP data has not updated since your last successful\ndownload of GROUP=active at 2026-03-26 08:10:22 UTC.\nData is updated once every 2 hours.\n", "celestrak_gp_not_modified", "GROUP=active", "2026-03-26 08:10:22 UTC"},
		{"<html><body>error</body></html>", "celestrak_gp_html_response", "", ""},
		{"", "empty_celestrak_gp_response", "", ""},
	}
	for _, tc := range cases {
		_, err := ParseCelestrakGPResponse([]byte(tc.data), CelestrakGPFormatJSON)
		var detail *DataCatalogError
		if !errors.As(err, &detail) || detail.Kind != tc.kind || optionalText(detail.Query) != tc.query || optionalText(detail.LastSuccessfulDownload) != tc.time {
			t.Errorf("%q: %#v", tc.kind, err)
		}
	}
}

func optionalText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func TestCelestrakResponseFormats(t *testing.T) {
	read := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile("testdata/celestrak/25544." + name)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	tle := read("tle")
	lines := strings.Split(strings.TrimSpace(string(tle)), "\n")
	cases := []struct {
		format CelestrakGPFormat
		data   []byte
	}{
		{CelestrakGPFormatTLE, tle},
		{CelestrakGPFormatThreeLine, []byte("0 " + strings.Join(lines, "\n") + "\n")},
		{CelestrakGPFormatTwoLine, []byte(lines[1] + "\n" + lines[2] + "\n")},
		{CelestrakGPFormatXML, read("xml")}, {CelestrakGPFormatKVN, read("kvn")},
		{CelestrakGPFormatJSON, read("json")}, {CelestrakGPFormatJSONPretty, read("json")},
		{CelestrakGPFormatCSV, []byte("OBJECT_NAME,EPOCH,MEAN_MOTION,ECCENTRICITY,INCLINATION,RA_OF_ASC_NODE,ARG_OF_PERICENTER,MEAN_ANOMALY,BSTAR\nISS,2026-01-01T00:00:00,1,0,0,0,0,0,0\n")},
	}
	for _, tc := range cases {
		response, err := ParseCelestrakGPResponse(tc.data, tc.format)
		if err != nil {
			t.Fatalf("format %d: %v", tc.format, err)
		}
		accepted, rejected, err := response.Counts()
		if err != nil || accepted != 1 || rejected != 0 {
			t.Errorf("format %d counts: %d %d %v", tc.format, accepted, rejected, err)
		}
		record, err := response.Record(0)
		recordData, snapshotErr := record.Snapshot()
		if snapshotErr != nil {
			t.Fatal(snapshotErr)
		}
		if err != nil || string(recordData.Raw) != string(tc.data) {
			t.Errorf("format %d raw: %q %v", tc.format, recordData.Raw, err)
		}
		if err := response.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCelestrakRejectedTLE(t *testing.T) {
	data := []byte("1 25544U 98067A   26168.18949189  .00009113  00000+0  17172-3 0  9996\n2 25544  51.6332 300.0813 0004737 195.1146 164.9702 15.49273435571753")
	response, err := ParseCelestrakGPResponse(data, CelestrakGPFormatTwoLine)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "rejected response", response.Close)
	accepted, rejected, err := response.Counts()
	if err != nil || accepted != 0 || rejected != 1 {
		t.Fatalf("counts %d %d %v", accepted, rejected, err)
	}
	record, err := response.RejectedRecord(0)
	if err != nil {
		t.Fatal(err)
	}
	recordData, err := record.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if recordData.RecordIndex != 1 || recordData.LineNumber == nil || *recordData.LineNumber != 2 || string(recordData.Raw) != string(data) || recordData.Reason.Kind != "tle" || recordData.Reason.TLE == nil || recordData.Reason.TLE.Kind != "checksum_mismatch" {
		t.Fatalf("rejection %+v", record)
	}
	fields := recordData.Reason.TLE.Fields
	if fields.LineLabel == nil || *fields.LineLabel != "line 2" || fields.Expected == nil || fields.Expected.String() != "3" || fields.Computed == nil || fields.Computed.String() != "2" {
		t.Fatalf("checksum detail %+v", fields)
	}
	if equal, err := record.Equal(record); err != nil || !equal {
		t.Fatalf("equal %v %v", equal, err)
	}
	if equal, err := record.ReasonEqual(record); err != nil || !equal {
		t.Fatalf("reason equality: %v %v", equal, err)
	}
}

func TestOMMMissingSizeAlternatives(t *testing.T) {
	data := []byte(`[{"OBJECT_NAME":"ISS","OBJECT_ID":"1998-067A","EPOCH":"2026-06-17T04:32:52","ECCENTRICITY":0.1,"INCLINATION":1,"RA_OF_ASC_NODE":2,"ARG_OF_PERICENTER":3,"MEAN_ANOMALY":4}]`)
	array, err := ParseOMMJSONArray(data)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "OMM array", array.Close)
	skipped, err := array.SkippedRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) != 1 || skipped[0].Error.Kind != "missing_one_of" || len(skipped[0].Error.Fields.Fields) != 2 || skipped[0].Error.Fields.Fields[0] != "SEMI_MAJOR_AXIS" || skipped[0].Error.Fields.Fields[1] != "MEAN_MOTION" {
		t.Fatalf("skip %+v", skipped)
	}
}

func TestOMMMalformedDocument(t *testing.T) {
	_, err := ParseOMMJSONArray([]byte("["))
	var engine *EngineError
	if !errors.As(err, &engine) || engine.Kind != "malformed_document" {
		t.Fatalf("malformed error %v", err)
	}
	var fields OMMParseErrorFields
	if err := json.Unmarshal(engine.Fields, &fields); err != nil {
		t.Fatal(err)
	}
	if fields.Format == nil || *fields.Format != "JSON" || fields.ByteOffset == nil {
		t.Fatalf("malformed detail %+v", fields)
	}
}

func TestCelestrakInvalidSelectors(t *testing.T) {
	cases := []struct {
		query           CelestrakGPQuery
		value, selector string
	}{
		{CelestrakGPGroup(""), "", "GROUP"}, {CelestrakGPInternationalDesignator("2020/025"), "2020/025", "INTDES"}, {CelestrakGPName(" "), " ", "NAME"},
	}
	for _, tc := range cases {
		_, err := CelestrakGPURL(tc.query, CelestrakGPFormatJSON)
		var detail *DataCatalogError
		if !errors.As(err, &detail) || detail.Kind != "invalid_celestrak_gp_selector" || detail.Selector != tc.selector || detail.Value != tc.value {
			t.Errorf("%s: %v", tc.selector, err)
		}
	}
}

func TestCelestrakStructuralRejection(t *testing.T) {
	first := "1 25544U 98067A   26168.18949189  .00009113  00000+0  17172-3 0  9996"
	second := "2 25544  51.6332 300.0813 0004737 195.1146 164.9702 15.49273435571752"
	rejected := "STRAY\n" + first + "\n" + second[:68] + "X\n\n"
	response, err := ParseCelestrakGPResponse([]byte(rejected+first+"\n"+second), CelestrakGPFormatTwoLine)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "structural response", response.Close)
	record, err := response.RejectedRecord(0)
	if err != nil {
		t.Fatal(err)
	}
	recordData, err := record.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if recordData.Reason.Kind != "invalid_tle_structure" || string(recordData.Raw) != rejected || recordData.LineNumber == nil || *recordData.LineNumber != 1 {
		t.Fatalf("structural rejection %+v", record)
	}
	if equal, err := record.StructuralErrorEqual(record); err != nil || !equal {
		t.Fatalf("structural equality %v %v", equal, err)
	}
}

func TestOMMFixtureSpans(t *testing.T) {
	jsonData, err := os.ReadFile("testdata/celestrak/25544.json")
	if err != nil {
		t.Fatal(err)
	}
	xmlData, err := os.ReadFile("testdata/celestrak/25544.xml")
	if err != nil {
		t.Fatal(err)
	}
	csvData := []byte("OBJECT_NAME,EPOCH,MEAN_MOTION,ECCENTRICITY,INCLINATION,RA_OF_ASC_NODE,ARG_OF_PERICENTER,MEAN_ANOMALY,BSTAR\nISS,2026-01-01T00:00:00,1,0,0,0,0,0,0\n")
	cases := []struct {
		name  string
		parse func([]byte) (*OMMArray, error)
		data  []byte
		want  OMMRecordSpan
	}{
		{"json", ParseOMMJSONArray, jsonData, OMMRecordSpan{Index: 0, ByteStart: 1, ByteEnd: 418, Line: 1}},
		{"xml", ParseOMMXMLAll, xmlData, OMMRecordSpan{Index: 0, ByteStart: 207, ByteEnd: 1194, Line: 3}},
		{"csv", ParseOMMCSVArray, csvData, OMMRecordSpan{Index: 0, ByteStart: 107, ByteEnd: 144, Line: 2}},
	}
	for _, tc := range cases {
		array, err := tc.parse(tc.data)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		spans, err := array.Spans()
		if err != nil || len(spans) != 1 || spans[0] != tc.want {
			t.Errorf("%s spans %+v %v", tc.name, spans, err)
		}
		if err := array.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCelestrakInvalidResponseDetail(t *testing.T) {
	_, err := ParseCelestrakGPResponse([]byte("not json"), CelestrakGPFormatJSON)
	var detail *DataCatalogError
	if !errors.As(err, &detail) || detail.Kind != "invalid_celestrak_gp_response" || detail.Reason != "malformed JSON: expected ident at line 1 column 2; response excerpt: \"not json\"" || detail.Line == nil || *detail.Line != 1 || detail.Column == nil || *detail.Column != 2 {
		t.Fatalf("response detail: %+v (%v)", detail, err)
	}
}

func fixturePtr[T any](value T) *T { return &value }

func TestCelestrakJSONConversions(t *testing.T) {
	data, err := os.ReadFile("testdata/celestrak/25544.json")
	if err != nil {
		t.Fatal(err)
	}
	response, err := ParseCelestrakGPResponse(data, CelestrakGPFormatJSON)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "GP response", response.Close)
	record, err := response.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	_, err = record.TLE()
	var tleStatus *StatusError
	if !errors.As(err, &tleStatus) || tleStatus.Code != StatusInvalidArgument || tleStatus.Detail != "sidereon_celestrak_gp_record_tle: payload is not TLE" {
		t.Fatalf("OMM payload TLE refusal: %v", err)
	}
	got, err := record.ToElementSet()
	if err != nil {
		t.Fatal(err)
	}
	want := OMMElementSet{Epoch: JulianDate{Whole: math.Float64frombits(0x4142c70c40000000), Fraction: math.Float64frombits(0x3fc841452f342018)}, BStar: 0.00017172, MeanMotionDot: fixturePtr(0.00009113), MeanMotionDoubleDot: fixturePtr(0.0), Eccentricity: 0.0004737, ArgumentOfPerigeeDeg: 195.1146, InclinationDeg: 51.6332, MeanAnomalyDeg: 164.9702, MeanMotionRevPerDay: 15.49273435, RightAscensionDeg: 300.0813, CatalogNumber: fixturePtr(uint32(25544)), OMMEpochDays: fixturePtr(math.Float64frombits(0x40db45cc20a2979a))}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("element set: %+v; want %+v", got, want)
	}
	omm, err := record.ToOMM()
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "GP OMM", omm.Close)
	snapshot, err := omm.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	expected := OMMData{ObjectName: fixturePtr("ISS (ZARYA)"), ObjectID: fixturePtr("1998-067A"), Epoch: OMMEpoch{Year: 2026, Month: 6, Day: 17, Hour: 4, Minute: 32, Second: 52, Microsecond: 99296}, MeanMotion: fixturePtr(15.49273435), Eccentricity: 0.0004737, InclinationDeg: 51.6332, RAOfAscNodeDeg: 300.0813, ArgOfPericenterDeg: 195.1146, MeanAnomalyDeg: 164.9702, EphemerisType: fixturePtr(int32(0)), ClassificationType: fixturePtr("U"), NORADCatID: fixturePtr(uint32(25544)), ElementSetNo: fixturePtr(int32(999)), RevAtEpoch: fixturePtr(int64(57175)), BStar: fixturePtr(0.00017172), MeanMotionDot: fixturePtr(0.00009113), MeanMotionDDot: fixturePtr(0.0), UserDefined: []OMMUserDefined{}, Comments: OMMComments{Header: []string{}, Metadata: []string{}, MeanElements: []string{}, TLEParameters: []string{}, UserDefined: []string{}}, QuantizeTLEDerivedFields: true}
	if !reflect.DeepEqual(snapshot, expected) {
		t.Fatalf("OMM snapshot: %+v; want %+v", snapshot, expected)
	}
}

func TestCelestrakRecordErrorValueEquality(t *testing.T) {
	one := &CelestrakGPRecordError{Kind: "tle", TLE: &CelestrakTLEError{Kind: "invalid_field", Fields: CelestrakTLEErrorFields{Field: fixturePtr("epoch"), Reason: fixturePtr("bad")}}}
	same := &CelestrakGPRecordError{Kind: "tle", TLE: &CelestrakTLEError{Kind: "invalid_field", Fields: CelestrakTLEErrorFields{Field: fixturePtr("epoch"), Reason: fixturePtr("bad")}}}
	different := &CelestrakGPRecordError{Kind: "tle", TLE: &CelestrakTLEError{Kind: "invalid_field", Fields: CelestrakTLEErrorFields{Field: fixturePtr("epoch"), Reason: fixturePtr("other")}}}
	if !one.Equal(same) || one.Equal(different) {
		t.Fatalf("record error equality failed")
	}
	omm := &CelestrakGPRecordError{Kind: "omm", OMM: &OMMParseError{Kind: "missing_field", Fields: OMMParseErrorFields{Field: fixturePtr("BSTAR")}}}
	ommEqual := &CelestrakGPRecordError{Kind: "omm", OMM: &OMMParseError{Kind: "missing_field", Fields: OMMParseErrorFields{Field: fixturePtr("BSTAR")}}}
	ommDifferent := &CelestrakGPRecordError{Kind: "omm", OMM: &OMMParseError{Kind: "missing_field", Fields: OMMParseErrorFields{Field: fixturePtr("MEAN_MOTION")}}}
	if !omm.Equal(ommEqual) || omm.Equal(ommDifferent) || omm.Equal(one) {
		t.Fatal("OMM record error equality failed")
	}
}

func TestCelestrakRejectionVariants(t *testing.T) {
	source, err := os.ReadFile("testdata/celestrak/25544.tle")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(source)), "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	jsonData, err := os.ReadFile("testdata/celestrak/25544.json")
	if err != nil {
		t.Fatal(err)
	}
	invalid := strings.Replace(string(jsonData), "\"MEAN_MOTION\":15.49273435", "\"MEAN_MOTION\":0.0", 1)
	cases := []struct {
		name       string
		format     CelestrakGPFormat
		body, kind string
		line       uint64
	}{
		{"stray", CelestrakGPFormatTwoLine, lines[1] + "\n" + lines[2] + "\nSTRAY\n" + lines[1] + "\n" + lines[2], "unexpected_content", 3},
		{"incomplete", CelestrakGPFormatTLE, lines[0] + "\n" + lines[1] + "\n" + lines[2] + "\nNEXT SAT", "incomplete_record", 4},
		{"element", CelestrakGPFormatJSON, invalid, "invalid_element_set", 1},
		{"omm", CelestrakGPFormatJSON, `[{"OBJECT_NAME":"X","EPOCH":"2026-01-01T00:00:00","ECCENTRICITY":0,"INCLINATION":0,"RA_OF_ASC_NODE":0,"ARG_OF_PERICENTER":0,"MEAN_ANOMALY":0}]`, "omm", 1},
	}
	for _, tc := range cases {
		response, err := ParseCelestrakGPResponse([]byte(tc.body), tc.format)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		rec, err := response.RejectedRecord(0)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		data, err := rec.Snapshot()
		if err != nil || data.Reason.Kind != tc.kind || data.LineNumber == nil || *data.LineNumber != tc.line {
			t.Errorf("%s: %+v %v", tc.name, data, err)
		}
		if tc.kind == "invalid_element_set" && (data.Reason.ElementSet == nil || data.Reason.ElementSet.Kind != "invalid_input" || data.Reason.ElementSet.Fields.Field == nil || *data.Reason.ElementSet.Fields.Field != "mean_motion" || data.Reason.ElementSet.Fields.Kind == nil || *data.Reason.ElementSet.Fields.Kind != "not_positive") {
			t.Errorf("element detail %+v", data.Reason)
		}
		if tc.kind == "omm" && (data.Reason.OMM == nil || data.Reason.OMM.Kind != "missing_one_of") {
			t.Errorf("OMM detail %+v", data.Reason)
		}
		if err := response.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCelestrakHighEccentricityRecord(t *testing.T) {
	data := []byte("1 25544U 98067A   26095.55331950  .00008543  00000+0  16420-3 0  9998\n2 25544  51.6328 299.5432 9999999 274.8255  85.2008 15.48786980560490\n")
	response, err := ParseCelestrakGPResponse(data, CelestrakGPFormatTwoLine)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "response", response.Close)
	accepted, rejected, err := response.Counts()
	if err != nil || accepted != 1 || rejected != 0 {
		t.Fatalf("counts %d %d %v", accepted, rejected, err)
	}
	record, err := response.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := record.Snapshot()
	if err != nil || snapshot.Payload.TLE == nil || snapshot.Payload.TLE.Elements.Eccentricity != 0.9999999 {
		t.Fatalf("snapshot %+v %v", snapshot, err)
	}
}

func TestCelestrakRejectedEquality(t *testing.T) {
	first := "1 25544U 98067A   26168.18949189  .00009113  00000+0  17172-3 0  9996"
	second := "2 25544  51.6332 300.0813 0004737 195.1146 164.9702 15.49273435571752"
	cases := []struct {
		format CelestrakGPFormat
		body   string
	}{
		{CelestrakGPFormatTwoLine, first + "\n" + second + "\nSTRAY\n" + first + "\n" + second},
		{CelestrakGPFormatTLE, "ISS\n" + first + "\n" + second + "\nNEXT SAT"},
	}
	var records [2]CelestrakGPRejectedRecord
	for i, tc := range cases {
		response, err := ParseCelestrakGPResponse([]byte(tc.body), tc.format)
		if err != nil {
			t.Fatal(err)
		}
		cleanupClose(t, "response", response.Close)
		records[i], err = response.RejectedRecord(0)
		if err != nil {
			t.Fatal(err)
		}
	}
	if equal, err := records[0].Equal(records[1]); err != nil || equal {
		t.Fatalf("rejected equality %v %v", equal, err)
	}
	if equal, err := records[0].ReasonEqual(records[1]); err != nil || equal {
		t.Fatalf("reason equality %v %v", equal, err)
	}
	if equal, err := records[0].StructuralErrorEqual(records[1]); err != nil || equal {
		t.Fatalf("structural equality %v %v", equal, err)
	}
	if equal, err := records[0].ReasonEqual(records[0]); err != nil || !equal {
		t.Fatalf("equal reason %v %v", equal, err)
	}
}

func TestCelestrakClosedResponse(t *testing.T) {
	response, err := ParseCelestrakGPResponse([]byte("1 25544U 98067A   26095.55331950  .00008543  00000+0  16420-3 0  9998\n2 25544  51.6328 299.5432 0006351 274.8255  85.2008 15.48786980560492\n"), CelestrakGPFormatTwoLine)
	if err != nil {
		t.Fatal(err)
	}
	record, err := response.Record(0)
	if err != nil {
		t.Fatal(err)
	}
	if err := response.Close(); err != nil {
		t.Fatal(err)
	}
	if err := response.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := response.Counts(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed count: %v", err)
	}
	if _, err := record.Snapshot(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed record: %v", err)
	}
	if _, err := record.TLE(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed TLE: %v", err)
	}
	if _, err := record.ToOMM(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed OMM: %v", err)
	}
	if _, err := record.ToElementSet(); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed elements: %v", err)
	}
	if _, err := response.Equal(response); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed equality: %v", err)
	}
}
