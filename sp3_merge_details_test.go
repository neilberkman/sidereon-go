package sidereon

import (
	"math"
	"os"
	"reflect"
	"testing"
)

func TestSP3MergeProvenanceAndReportCollections(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	source, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, source)
	options, err := NewSP3MergeOptions()
	if err != nil {
		t.Fatal(err)
	}
	options.MinAgree = 1
	options.ProvenanceMode = SP3ProvenanceFull
	merged, report, err := MergeSP3([]*SP3{source}, &options)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, merged)
	closeAfterTest(t, report)
	info, err := report.ProvenanceInfo()
	if err != nil {
		t.Fatal(err)
	}
	if !info.Recorded || info.Mode != SP3ProvenanceFull || info.CellCount == 0 || info.CoverageCount != 1 {
		t.Fatalf("provenance info = %+v", info)
	}
	cells, err := report.ProvenanceCells()
	if err != nil || len(cells) != info.CellCount {
		t.Fatalf("provenance cells=%d info=%+v err=%v", len(cells), info, err)
	}
	cell := cells[0]
	if !cell.HasPosition || !cell.Position.HasSource || cell.Position.Source != 0 || cell.Position.MemberCount != 1 || cell.Satellite != "G08" || cell.Position.Kind != SP3CellSelectionSingleSource {
		t.Fatalf("first provenance cell = %+v", cell)
	}
	sourceState, err := source.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	sourceEpochs, err := source.Epochs()
	if err != nil || len(sourceEpochs) == 0 {
		t.Fatalf("source epochs=%v err=%v", sourceEpochs, err)
	}
	mergedState, err := merged.State("G08", 0)
	if err != nil {
		t.Fatal(err)
	}
	if cell.EpochJ2000S != sourceEpochs[0] || sourceState.PositionM != mergedState.PositionM || sourceState.ClockS != mergedState.ClockS {
		t.Fatalf("single-source provenance does not match independently read product values: cell=%+v source=%+v merged=%+v", cell, sourceState, mergedState)
	}
	wantFirstPositionM := [3]float64{6_042_284.075, -25_684_525.952, -1_323_814.227}
	for i := range wantFirstPositionM {
		if math.Abs(sourceState.PositionM[i]-wantFirstPositionM[i]) > 2e-9 {
			t.Fatalf("independent first G08 source position = %v, want %v", sourceState.PositionM, wantFirstPositionM)
		}
	}
	positionMembers, err := report.ProvenanceCellMembers(0, SP3MergePositionChannel)
	if err != nil || len(positionMembers) != 1 || positionMembers[0] != 0 {
		t.Fatalf("position members=%v err=%v", positionMembers, err)
	}
	coverage, err := report.ProvenanceCoverage()
	if err != nil || len(coverage) != 1 || coverage[0].Source != 0 || coverage[0].CellsContributed == 0 || !coverage[0].HasFirstEpoch || !coverage[0].HasLastEpoch {
		t.Fatalf("coverage=%+v err=%v", coverage, err)
	}
	metrics, err := report.AgreementMetrics()
	if err != nil || len(metrics) == 0 || metrics[0].Satellite == "" {
		t.Fatalf("agreement metrics=%+v err=%v", metrics, err)
	}
	if dropped, err := report.DroppedInputEpochs(); err != nil || len(dropped) != 0 {
		t.Fatalf("dropped epochs=%+v err=%v", dropped, err)
	}
	if omitted, err := report.OmittedEpochs(); err != nil || len(omitted) != 0 {
		t.Fatalf("omitted epochs=%+v err=%v", omitted, err)
	}
	if clockOmissions, err := report.ClockOmissions(); err != nil || len(clockOmissions) != 0 {
		t.Fatalf("clock omissions=%+v err=%v", clockOmissions, err)
	}
	transitions, err := report.ProvenanceTransitions()
	if err != nil || len(transitions) == 0 {
		t.Fatalf("opening transitions=%+v err=%v", transitions, err)
	}
	for _, transition := range transitions {
		if !transition.HasToSource || transition.ToSource != 0 || transition.HasFromSource {
			t.Fatalf("single-source opening transition = %+v", transition)
		}
	}
	if data, err := report.ContinuityJSON(); err != nil || string(data) != "null" {
		t.Fatalf("continuity JSON=%s err=%v", data, err)
	}
	verified, nodes, err := report.ContinuitySelectedNodes(cell.Satellite, cell.EpochJ2000S, cell.EpochJ2000S)
	if err != nil || verified || len(nodes) != 0 {
		t.Fatalf("unverified continuity nodes: verified=%v nodes=%v err=%v", verified, nodes, err)
	}
	cellsSnapshot := append([]SP3MergeCellProvenance(nil), cells...)
	coverageSnapshot := append([]SP3ContributorCoverage(nil), coverage...)
	if err := report.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSP3([]byte("not an SP3 product")); err == nil {
		t.Fatal("unrelated invalid native parse unexpectedly succeeded")
	}
	if !reflect.DeepEqual(cells, cellsSnapshot) || !reflect.DeepEqual(coverage, coverageSnapshot) || cells[0].Satellite != "G08" {
		t.Fatal("detached provenance rows changed after report close and a failed native parse")
	}
}
