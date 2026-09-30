//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestANTEXTypedParseAndEncodeOutcomes(t *testing.T) {
	data, err := os.ReadFile("testdata/antex/igs20_wettzell_trim.atx")
	if err != nil {
		t.Fatal(err)
	}
	endRecord := bytes.Index(data, []byte("END OF ANTENNA"))
	if endRecord < 0 {
		t.Fatal("fixture has no antenna block terminator")
	}
	lineEnd := bytes.IndexByte(data[endRecord:], '\n') + endRecord
	if lineEnd < endRecord {
		t.Fatal("fixture has no newline after first antenna block")
	}
	outerLine := []byte(fmt.Sprintf("%-60sCOMMENT\n", "between-block-retention-test"))
	withOuter := make([]byte, 0, len(data)+len(outerLine))
	withOuter = append(withOuter, data[:lineEnd+1]...)
	withOuter = append(withOuter, outerLine...)
	withOuter = append(withOuter, data[lineEnd+1:]...)
	product, parsed, err := ParseANTEXWithOutcome(withOuter)
	if err != nil || product == nil || !parsed.IsOK || parsed.Error.Kind != ANTEXErrorNone {
		t.Fatalf("parse outcome = %v, %+v, %v", product, parsed, err)
	}
	closeAfterTest(t, product)
	encoded, written, err := product.EncodeWithOutcome()
	if err != nil || !written.IsOK || written.Error.Kind != ANTEXErrorNone || len(encoded) == 0 {
		t.Fatalf("encode outcome = %d bytes, %+v, %v", len(encoded), written, err)
	}
	header, err := product.Header()
	if err != nil || !header.HasVersion || header.Version != 1.4 || !header.EndOfHeader || header.HeaderCommentCount == 0 || header.PCVType != ANTEXPCVAbsolute {
		t.Fatalf("retained header = %+v, %v", header, err)
	}
	if comment, err := product.HeaderComment(0); err != nil || comment == "" {
		t.Fatalf("first header comment = %q, %v", comment, err)
	}
	for part := ANTEXHeaderTextType; part <= ANTEXHeaderTextReference; part++ {
		if _, err := product.HeaderText(part); err != nil {
			t.Fatalf("header text part %d: %v", part, err)
		}
	}
	assertInvalidArgument := func(label string, err error) {
		t.Helper()
		var status *StatusError
		if !errors.As(err, &status) || status.Code != StatusInvalidArgument || status.Detail == "" {
			t.Fatalf("%s error = %v, want typed StatusInvalidArgument", label, err)
		}
	}
	if _, err := product.HeaderText(ANTEXHeaderTextPart(3)); err == nil {
		t.Fatal("invalid header text part was accepted")
	} else {
		assertInvalidArgument("header text part", err)
	}
	if _, err := product.HeaderComment(-1); err == nil {
		t.Fatal("negative header comment index was accepted")
	} else {
		assertInvalidArgument("header comment index", err)
	}
	outerCount, err := product.OuterCommentCount()
	if err != nil || outerCount != 1 {
		t.Fatalf("outer comment count = %d, %v", outerCount, err)
	}
	outer, before, err := product.OuterComment(0)
	if err != nil || outer != "between-block-retention-test" || before != 1 {
		t.Fatalf("first outer comment = %q before %d blocks, %v", outer, before, err)
	}
	if _, _, err := product.OuterComment(-1); err == nil {
		t.Fatal("negative outer comment index was accepted")
	} else {
		assertInvalidArgument("outer comment index", err)
	}
	if skipped, err := product.SkippedRecords(); err != nil || skipped < 0 {
		t.Fatalf("skipped record count = %d, %v", skipped, err)
	}
	blocks, err := product.BlockCount()
	if err != nil || blocks == 0 {
		t.Fatalf("block count = %d, %v", blocks, err)
	}
	block, err := product.Block(0)
	if err != nil || block == nil {
		t.Fatalf("first antenna block = %v, %v", block, err)
	}
	if _, err := product.Block(-1); err == nil {
		t.Fatal("negative block index was accepted")
	} else {
		assertInvalidArgument("block index", err)
	}
	closeAfterTest(t, block)
	blockInfo, err := block.Info()
	if err != nil || blockInfo.FrequencyCount == 0 {
		t.Fatalf("antenna info = %+v, %v", blockInfo, err)
	}
	for _, part := range []ANTEXAntennaText{ANTEXAntennaTextID, ANTEXAntennaTextType, ANTEXAntennaTextSerial, ANTEXAntennaTextSINEX} {
		if _, err := block.Text(part); err != nil {
			t.Fatalf("antenna text part %d: %v", part, err)
		}
	}
	if blockInfo.LeadingCommentCount > 0 {
		if _, err := block.Comment(ANTEXAntennaLeadingComments, 0); err != nil {
			t.Fatalf("leading antenna comment: %v", err)
		}
	}
	if blockInfo.CommentCount > 0 {
		if _, err := block.Comment(ANTEXAntennaBlockComments, 0); err != nil {
			t.Fatalf("antenna block comment: %v", err)
		}
	}
	if blockInfo.CalibrationCount > 0 {
		if _, err := block.Calibration(0); err != nil {
			t.Fatalf("antenna calibration: %v", err)
		}
		for _, part := range []ANTEXCalibrationText{ANTEXCalibrationMethod, ANTEXCalibrationAgency, ANTEXCalibrationDate} {
			if _, err := block.CalibrationText(0, part); err != nil {
				t.Fatalf("calibration text part %d: %v", part, err)
			}
		}
	}
	for index := 0; index < blockInfo.FrequencyCount; index++ {
		frequency, err := block.Frequency(index)
		if err != nil {
			t.Fatalf("frequency %d: %v", index, err)
		}
		if _, err := block.FrequencyLabel(index); err != nil {
			t.Fatalf("frequency label %d: %v", index, err)
		}
		samples, err := block.FrequencyPCVSamples(index, false)
		if err != nil || len(samples) != frequency.PCVSampleCount {
			t.Fatalf("frequency samples %d = %d, info=%+v, err=%v", index, len(samples), frequency, err)
		}
		rms, err := block.FrequencyPCVSamples(index, true)
		if err != nil || len(rms) != frequency.RMSPCVSampleCount {
			t.Fatalf("frequency RMS samples %d = %d, info=%+v, err=%v", index, len(rms), frequency, err)
		}
	}
	if _, err := block.PCO("ZZZ"); err == nil {
		t.Fatal("unknown frequency PCO lookup was accepted")
	} else {
		var status *StatusError
		if !errors.As(err, &status) || status.Code != StatusInvalidArgument || status.ANTEX == nil || status.ANTEX.Kind != ANTEXErrorUnknownFrequency || !status.ANTEX.HasFrequency || status.ANTEX.Frequency != "ZZZ" || status.ANTEX.Message == "" {
			t.Fatalf("unknown frequency error lost its typed ANTEX detail: %#v", err)
		}
	}
	valid, err := block.ValidAt(ANTEXDateTime{Year: 2009, Month: 8, Day: 17})
	if err != nil || !valid {
		t.Fatalf("antenna validity at fixture epoch = %v, %v", valid, err)
	}
	epoch := ANTEXDateTime{Year: 2009, Month: 8, Day: 17}
	id := "BLOCK IIR-M         G05                 G050      2009-043A"
	byID, found, err := product.AntennaAt(id, epoch)
	if err != nil || !found || byID == nil {
		t.Fatalf("antenna at validity start = %v, %v, %v", byID, found, err)
	}
	closeAfterTest(t, byID)
	byPRN, found, err := product.SatelliteAntenna("G05", epoch)
	if err != nil || !found || byPRN == nil {
		t.Fatalf("satellite antenna at validity start = %v, %v, %v", byPRN, found, err)
	}
	closeAfterTest(t, byPRN)
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := block.PCO("G01"); err != nil {
		t.Fatalf("detached antenna block after source close: %v", err)
	}
	if _, err := block.Info(); err != nil {
		t.Fatalf("antenna info after source close: %v", err)
	}
	if _, err := block.Text(ANTEXAntennaTextID); err != nil {
		t.Fatalf("antenna text after source close: %v", err)
	}
	if _, err := block.FrequencyPCVSamples(0, false); err != nil {
		t.Fatalf("antenna samples after source close: %v", err)
	}
	if _, err := byID.PCO("G01"); err != nil {
		t.Fatalf("detached timed antenna after source close: %v", err)
	}
	if _, err := byPRN.PCO("G01"); err != nil {
		t.Fatalf("detached satellite antenna after source close: %v", err)
	}
	bad := bytes.Replace(withOuter, []byte("     1.4            M"), []byte("     BAD            M"), 1)
	if bytes.Equal(bad, withOuter) {
		t.Fatal("fixture did not contain the expected version field")
	}
	_, failed, err := ParseANTEXWithOutcome(bad)
	if err != nil || failed.IsOK || failed.Error.Kind != ANTEXErrorInvalidField || !failed.Error.HasField || failed.Error.Field != "version" || !failed.Error.HasValue || failed.Error.Value != "BAD" || failed.Error.Message == "" {
		t.Fatalf("typed malformed-field outcome = %+v, %v", failed, err)
	}
	// A later native operation must not overwrite the detached refusal details.
	validProduct, outcome, err := ParseANTEXWithOutcome(withOuter)
	if err != nil || validProduct == nil || !outcome.IsOK {
		t.Fatalf("follow-up valid parse = %v, %+v, %v", validProduct, outcome, err)
	}
	closeAfterTest(t, validProduct)
	if failed.Error.Kind != ANTEXErrorInvalidField || failed.Error.Field != "version" || failed.Error.Value != "BAD" || failed.Error.Message == "" {
		t.Fatalf("parse refusal changed after subsequent native call: %+v", failed.Error)
	}
}
