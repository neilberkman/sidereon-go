package sidereon

import (
	"bytes"
	"errors"
	"testing"
)

func TestOMMLenientWriters(t *testing.T) {
	data := []byte(`{"OBJECT_NAME":"ISS (ZARYA)","OBJECT_ID":"1998-067A","EPOCH":"2026-06-17T04:32:52.099296","MEAN_MOTION":15.49273435,"ECCENTRICITY":0.0004737,"INCLINATION":51.6332,"RA_OF_ASC_NODE":300.0813,"ARG_OF_PERICENTER":195.1146,"MEAN_ANOMALY":164.9702,"BSTAR":0.00017172}`)
	omm, err := ParseOMMJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "omm", omm.Close)
	for _, write := range []func() ([]byte, error){omm.KVN, omm.XML} {
		_, err := write()
		var detail *OMMParseError
		if !errors.As(err, &detail) || detail.Kind != "missing_field" || detail.Fields.Field == nil || *detail.Fields.Field != "CCSDS_OMM_VERS" {
			t.Fatalf("strict error: %v", err)
		}
	}
	kvn, err := omm.ToKVNLenient()
	if err != nil || !bytes.Contains(kvn, []byte("OBJECT_NAME = ISS (ZARYA)")) {
		t.Fatalf("lenient KVN: %q, %v", kvn, err)
	}
	xml, err := omm.ToXMLLenient()
	if err != nil || !bytes.Contains(xml, []byte("<OBJECT_NAME>ISS (ZARYA)</OBJECT_NAME>")) {
		t.Fatalf("lenient XML: %q, %v", xml, err)
	}
}

func TestOMMFromTLEAndHeader(t *testing.T) {
	tle, err := ParseTLE("1 25544U 98067A   26095.55331950  .00008543  00000+0  16420-3 0  9998", "2 25544  51.6328 299.5432 0006351 274.8255  85.2008 15.48786980560492")
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "tle", tle.Close)
	omm, err := OMMFromTLE(tle, "ISS (ZARYA)")
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "omm", omm.Close)
	_, err = omm.WithCCSDSHeader("2.0", "bad", "TEST")
	var invalidHeader *OMMParseError
	if !errors.As(err, &invalidHeader) || invalidHeader.Kind != "invalid_field" || invalidHeader.Fields.Field == nil || *invalidHeader.Fields.Field != "CREATION_DATE" || invalidHeader.Fields.KindValue == nil || *invalidHeader.Fields.KindValue != "invalid_civil_date" {
		t.Fatalf("invalid header: %v", err)
	}
	data, err := omm.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if data.ObjectName == nil || *data.ObjectName != "ISS (ZARYA)" || data.ObjectID == nil || *data.ObjectID != "1998-067A" {
		t.Fatalf("snapshot %+v", data)
	}
	withHeader, err := omm.WithCCSDSHeader("2.0", "2026-04-06T00:00:00", "TEST")
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "withHeader", withHeader.Close)
	if _, err = withHeader.ToKVN(); err != nil {
		t.Fatal(err)
	}
	same, err := omm.Equal(omm)
	if err != nil || !same {
		t.Fatalf("equal %v %v", same, err)
	}
	different, err := omm.Equal(withHeader)
	if err != nil || different {
		t.Fatalf("different %v %v", different, err)
	}
}

func TestOMMStrictWriterText(t *testing.T) {
	input := []byte("OBJECT_NAME = X\nOBJECT_ID = 1998-067A\nCENTER_NAME = EARTH\nREF_FRAME = TEME\nTIME_SYSTEM = UTC\nMEAN_ELEMENT_THEORY = OTHER\nEPOCH = 2026-01-01T00:00:00.000000\nMEAN_MOTION = 1\nECCENTRICITY = 0\nINCLINATION = 0\nRA_OF_ASC_NODE = 0\nARG_OF_PERICENTER = 0\nMEAN_ANOMALY = 0\n")
	omm, err := ParseOMMKVN(input)
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "partial OMM", omm.Close)
	complete, err := omm.WithCCSDSHeader("2.0", "2026-01-02T00:00:00", "TEST")
	if err != nil {
		t.Fatal(err)
	}
	cleanupClose(t, "complete OMM", complete.Close)
	wantKVN := "CCSDS_OMM_VERS = 2.0\nCREATION_DATE = 2026-01-02T00:00:00\nORIGINATOR = TEST\nOBJECT_NAME = X\nOBJECT_ID = 1998-067A\nCENTER_NAME = EARTH\nREF_FRAME = TEME\nTIME_SYSTEM = UTC\nMEAN_ELEMENT_THEORY = OTHER\nEPOCH = 2026-01-01T00:00:00.000000\nMEAN_MOTION = 1\nECCENTRICITY = 0\nINCLINATION = 0\nRA_OF_ASC_NODE = 0\nARG_OF_PERICENTER = 0\nMEAN_ANOMALY = 0\n"
	kvn, err := complete.ToKVN()
	if err != nil || string(kvn) != wantKVN {
		t.Fatalf("KVN: %q %v", kvn, err)
	}
	wantXML := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<ndm>\n<omm id=\"CCSDS_OMM_VERS\" version=\"2.0\">\n<header><CREATION_DATE>2026-01-02T00:00:00</CREATION_DATE><ORIGINATOR>TEST</ORIGINATOR></header>\n<body><segment>\n<metadata><OBJECT_NAME>X</OBJECT_NAME><OBJECT_ID>1998-067A</OBJECT_ID><CENTER_NAME>EARTH</CENTER_NAME><REF_FRAME>TEME</REF_FRAME><TIME_SYSTEM>UTC</TIME_SYSTEM><MEAN_ELEMENT_THEORY>OTHER</MEAN_ELEMENT_THEORY></metadata>\n<data>\n<meanElements><EPOCH>2026-01-01T00:00:00.000000</EPOCH><MEAN_MOTION>1</MEAN_MOTION><ECCENTRICITY>0</ECCENTRICITY><INCLINATION>0</INCLINATION><RA_OF_ASC_NODE>0</RA_OF_ASC_NODE><ARG_OF_PERICENTER>0</ARG_OF_PERICENTER><MEAN_ANOMALY>0</MEAN_ANOMALY></meanElements>\n</data>\n</segment></body>\n</omm>\n</ndm>\n"
	xml, err := complete.ToXML()
	if err != nil || string(xml) != wantXML {
		t.Fatalf("XML: %q %v", xml, err)
	}
}
