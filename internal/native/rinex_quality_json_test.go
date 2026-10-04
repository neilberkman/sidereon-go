package native

import (
	"encoding/json"
	"math"
	"testing"
)

func TestDecodeRINEXLintFindingDetailPreservesNumbers(t *testing.T) {
	payload := []byte(`{"kind":"NavUnhealthyRecords","spec_ref":"fixture","details":{"count":18446744073709551615,"none":null,"zero":0,"signed_zero":-0.0}}`)
	detail, err := decodeRINEXLintFindingDetail(payload)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Kind != "NavUnhealthyRecords" || detail.SpecRef != "fixture" {
		t.Fatalf("header=%+v", detail)
	}
	if got, ok := detail.Details["count"].(json.Number); !ok || got.String() != "18446744073709551615" {
		t.Fatalf("count=%T(%v)", detail.Details["count"], detail.Details["count"])
	}
	if detail.Details["none"] != nil {
		t.Fatalf("none=%#v", detail.Details["none"])
	}
	if got, ok := detail.Details["zero"].(json.Number); !ok || got.String() != "0" {
		t.Fatalf("zero=%T(%v)", detail.Details["zero"], detail.Details["zero"])
	}
	got, ok := detail.Details["signed_zero"].(json.Number)
	if !ok {
		t.Fatalf("signed_zero=%T", detail.Details["signed_zero"])
	}
	value, err := got.Float64()
	if err != nil || value != 0 || !math.Signbit(value) {
		t.Fatalf("signed_zero=%v err=%v", value, err)
	}
}

func TestDecodeRINEXLintFindingDetailRequiresOneCompleteObject(t *testing.T) {
	for _, payload := range [][]byte{
		[]byte(`{"kind":"K","spec_ref":"S","details":{}} {}`),
		[]byte(`{"kind":"K","spec_ref":"S","details":null}`),
	} {
		if _, err := decodeRINEXLintFindingDetail(payload); err == nil {
			t.Fatalf("accepted %s", payload)
		}
	}
}
