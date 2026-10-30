//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
)

func TestSP3EncodeTypedOutcomeAndOwnedRefusal(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	product, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, product)
	encoded, outcome, err := product.Encode()
	if err != nil || !outcome.IsOK || len(encoded) == 0 {
		t.Fatalf("Encode() = %d bytes, %+v, %v", len(encoded), outcome, err)
	}
	if !strings.HasPrefix(string(encoded), "#cP2020") {
		t.Fatalf("encoded SP3 header = %q", encoded[:min(len(encoded), 40)])
	}

	refusalData := bytes.Replace(data, []byte("%f  0.0000000"), []byte("%f 1.25000001"), 1)
	if bytes.Equal(refusalData, data) {
		t.Fatal("fixture did not contain the expected base field")
	}
	refused, err := LoadSP3(refusalData)
	if err != nil {
		t.Fatalf("parse finer base fixture: %v", err)
	}
	closeAfterTest(t, refused)
	text, refusedOutcome, err := refused.Encode()
	e := refusedOutcome.Error
	if err != nil || text != nil || refusedOutcome.IsOK || e.Kind != SP3WriteErrorPrecisionNotRepresentable || !e.HasField || e.Field != "pos/vel base" || !e.HasColumns || e.Columns != 10 || !e.HasDecimals || e.Decimals != 7 || !e.HasNumber || e.Number != 1.25000001 || e.Message == "" {
		t.Fatalf("typed writer refusal = %q, %+v, %v", text, refusedOutcome, err)
	}
	owned := e
	if err := refused.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSP3([]byte("invalid")); err == nil {
		t.Fatal("invalid SP3 parse unexpectedly succeeded")
	}
	if owned.Kind != SP3WriteErrorPrecisionNotRepresentable || owned.Field != "pos/vel base" || owned.Number != 1.25000001 || owned.Message == "" {
		t.Fatalf("writer outcome changed after source close and intervening failed parse: %+v", owned)
	}
}

func TestSP3ToSP3StringWithRoundingReportsChangesAndKeepsTypedRefusal(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	product, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, product)
	before, err := product.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	strictText, strictOutcome, err := product.Encode()
	if err != nil || !strictOutcome.IsOK {
		t.Fatalf("strict write = %+v, %v", strictOutcome, err)
	}
	text, report, err := product.ToSP3StringWithRounding()
	if err != nil || report != (SP3RoundingReport{}) {
		t.Fatalf("rounded write = report %+v, err %v", report, err)
	}
	if text != string(strictText) {
		t.Fatal("grid-representable input changed under explicit rounding")
	}
	reparsed, err := LoadSP3([]byte(text))
	if err != nil {
		t.Fatalf("reparse rounded write: %v", err)
	}
	closeAfterTest(t, reparsed)
	after, err := product.State("G08", 0)
	if err != nil || after.ClockS != before.ClockS {
		t.Fatalf("rounded write changed source clock: before %+v, after %+v, err %v", before, after, err)
	}
	written, err := reparsed.State("G08", 0)
	if err != nil || written.ClockS != before.ClockS {
		t.Fatalf("rounded output clock = %+v, original %+v, err %v", written, before, err)
	}

	// Keep the P-record width fixed while adding one source decimal beyond the
	// SP3 clock column precision. The ordinary writer must refuse it; the opt-in
	// writer rounds a copy and reports the actual seconds change.
	lineStart := bytes.Index(data, []byte("PG08   6042.284075"))
	if lineStart < 0 {
		t.Fatal("fixture did not contain the expected first G08 record")
	}
	relativeLineEnd := bytes.IndexByte(data[lineStart:], '\n')
	if relativeLineEnd < 0 {
		t.Fatal("first G08 record had no line ending")
	}
	lineEnd := lineStart + relativeLineEnd
	line := data[lineStart:lineEnd]
	if len(line) < 60 || strings.TrimSpace(string(line[46:60])) != "-38.640412" {
		t.Fatalf("unexpected first G08 clock column: %q", line[46:])
	}
	finerLine := append([]byte(nil), line[:46]...)
	finerLine = append(finerLine, []byte(fmt.Sprintf("%14s", "-38.6404124"))...)
	finerLine = append(finerLine, line[60:]...)
	finerData := append([]byte(nil), data[:lineStart]...)
	finerData = append(finerData, finerLine...)
	finerData = append(finerData, data[lineEnd:]...)
	finerProduct, err := LoadSP3(finerData)
	if err != nil {
		t.Fatalf("parse in-column higher-precision clock: %v", err)
	}
	closeAfterTest(t, finerProduct)
	finerBefore, err := finerProduct.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, strictOutcome, err := finerProduct.Encode()
	if err != nil || strictOutcome.IsOK || strictOutcome.Error.Kind != SP3WriteErrorRecordValueNotRepresentable {
		t.Fatalf("strict writer accepted higher-precision clock: %+v, %v", strictOutcome, err)
	}
	finerText, finerReport, err := finerProduct.ToSP3StringWithRounding()
	if err != nil {
		t.Fatalf("rounded write of higher-precision clock: %v", err)
	}
	if finerReport.ClockS <= 0 || finerReport.PositionXM != 0 || finerReport.PositionYM != 0 || finerReport.PositionZM != 0 || finerReport.VelocityXMS != 0 || finerReport.VelocityYMS != 0 || finerReport.VelocityZMS != 0 || finerReport.ClockRateSS != 0 {
		t.Fatalf("rounding report did not isolate the clock change: %+v", finerReport)
	}
	finerReparsed, err := LoadSP3([]byte(finerText))
	if err != nil {
		t.Fatalf("reparse rounded higher-precision clock: %v", err)
	}
	closeAfterTest(t, finerReparsed)
	finerAfterSource, err := finerProduct.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	finerAfterOutput, err := finerReparsed.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(finerAfterSource.ClockS) != math.Float64bits(finerBefore.ClockS) {
		t.Fatalf("rounded writer changed source clock: before %.17g, after %.17g", finerBefore.ClockS, finerAfterSource.ClockS)
	}
	if delta := math.Abs(finerAfterOutput.ClockS - finerBefore.ClockS); math.Float64bits(delta) != math.Float64bits(finerReport.ClockS) {
		t.Fatalf("clock rounding report %.17g != reparsed change %.17g", finerReport.ClockS, delta)
	}

	baseRefusal := bytes.Replace(data, []byte("%f  0.0000000"), []byte("%f 1.25000001"), 1)
	refused, err := LoadSP3(baseRefusal)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, refused)
	text, _, err = refused.ToSP3StringWithRounding()
	refusal, ok := err.(SP3WriteRefusalError)
	if !ok || text != "" || refusal.Outcome.IsOK || refusal.Outcome.Error.Kind != SP3WriteErrorPrecisionNotRepresentable {
		t.Fatalf("rounded refusal = %q, %T %v", text, err, err)
	}
}

func TestMergedGridSP3WritesAndRoundTrips(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	// Change one valid six-decimal clock column by one column unit, so the
	// two-source mean falls halfway between representable clock columns.
	secondData := bytes.Replace(data,
		[]byte("PG08   6042.284075 -25684.525952  -1323.814227    -38.640412"),
		[]byte("PG08   6042.284075 -25684.525952  -1323.814227    -38.640413"), 1)
	if bytes.Equal(secondData, data) {
		t.Fatal("fixture did not contain the expected first G08 record")
	}
	first, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, first)
	second, err := LoadSP3(secondData)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, second)
	for label, source := range map[string]*SP3{"first": first, "second": second} {
		if _, outcome, err := source.Encode(); err != nil || !outcome.IsOK {
			t.Fatalf("%s writable SP3 input = %+v, %v", label, outcome, err)
		}
	}

	options, err := NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.MinAgree = 1
	options.ClockMinCommon = 1
	merged, report, err := MergeSP3([]*SP3{first, second}, &options)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, merged)
	closeAfterTest(t, report)
	encoded, outcome, err := merged.Encode()
	if err != nil || !outcome.IsOK || len(encoded) == 0 {
		t.Fatalf("strict write of merged grid product = %+v, %v", outcome, err)
	}
	roundedText, rounding, err := merged.ToSP3StringWithRounding()
	if err != nil {
		t.Fatal(err)
	}
	if rounding != (SP3RoundingReport{}) || roundedText != string(encoded) {
		t.Fatalf("explicit writer changed already-quantized merge: report=%+v", rounding)
	}
	reparsed, err := LoadSP3(encoded)
	if err != nil {
		t.Fatalf("reparse merged output: %v", err)
	}
	closeAfterTest(t, reparsed)
	mergedState, err := merged.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	reparsedState, err := reparsed.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Float64bits(mergedState.ClockS) != math.Float64bits(reparsedState.ClockS) {
		t.Fatalf("merged/read-back clock differs: merged %.17g, read-back %.17g", mergedState.ClockS, reparsedState.ClockS)
	}
	if _, outcome, err := reparsed.Encode(); err != nil || !outcome.IsOK {
		t.Fatalf("strict write after merge/read-back = %+v, %v", outcome, err)
	}
}
