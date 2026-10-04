package sidereon

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestSP3FullContinuityPolicyRoutes(t *testing.T) {
	policy, err := DefaultSP3ContinuityPolicy(SP3OrbitClassMEOGNSS)
	if err != nil {
		t.Fatal(err)
	}
	if policy.SpeedBoundKind != SP3SpeedBoundOrbitClass || policy.OrbitClass != SP3OrbitClassMEOGNSS || !policy.ResidualToleranceEnabled || policy.ResidualToleranceM != 1 || policy.GapThresholdFactor != 1.5 {
		t.Fatalf("native continuity defaults = %+v", policy)
	}
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, sp3)
	report, err := sp3.ContinuityReportJSON(policy)
	if err != nil {
		t.Fatal(err)
	}
	var reportValue map[string]json.RawMessage
	if err := json.Unmarshal(report, &reportValue); err != nil {
		t.Fatalf("continuity report is invalid JSON: %v", err)
	}
	for _, field := range []string{"attested", "defects", "pairs_checked", "residuals_checked", "residuals_skipped"} {
		if _, ok := reportValue[field]; !ok {
			t.Errorf("continuity report is missing %q: %s", field, report)
		}
	}
	verdict, err := sp3.ContinuityVerdictJSONWithPolicy(policy, 646260200, 646260400)
	if err != nil {
		t.Fatal(err)
	}
	var verdictValue map[string]json.RawMessage
	if err := json.Unmarshal(verdict, &verdictValue); err != nil {
		t.Fatalf("continuity verdict is invalid JSON: %v", err)
	}
	for _, field := range []string{"decision", "accepted", "influencing_defects", "all_defects"} {
		if _, ok := verdictValue[field]; !ok {
			t.Errorf("continuity verdict is missing %q: %s", field, verdict)
		}
	}
	epochs, err := sp3.Epochs()
	if err != nil || len(epochs) < 2 {
		t.Fatalf("continuity epochs = %v, err=%v", epochs, err)
	}
	strict := policy
	strict.SpeedBoundKind = SP3SpeedBoundExplicit
	strict.ExplicitMaxSpeedMPS = 0
	strict.ResidualToleranceEnabled = false
	strictReport, err := sp3.ContinuityReportJSON(strict)
	if err != nil {
		t.Fatal(err)
	}
	var strictValue struct {
		Attested bool `json:"attested"`
		Defects  []struct {
			Kind string `json:"kind"`
		} `json:"defects"`
	}
	if err := json.Unmarshal(strictReport, &strictValue); err != nil {
		t.Fatal(err)
	}
	if strictValue.Attested || len(strictValue.Defects) == 0 || strictValue.Defects[0].Kind != "speed_bound" {
		t.Fatalf("zero-speed policy report = %s", strictReport)
	}
	noSpeed := policy
	noSpeed.SpeedBoundKind = SP3SpeedBoundNone
	noSpeed.ResidualToleranceEnabled = false
	noSpeedReport, err := sp3.ContinuityReportJSON(noSpeed)
	if err != nil {
		t.Fatal(err)
	}
	var noSpeedValue struct {
		Attested         bool `json:"attested"`
		PairsChecked     int  `json:"pairs_checked"`
		ResidualsChecked int  `json:"residuals_checked"`
		Defects          []struct {
			Kind string `json:"kind"`
		} `json:"defects"`
	}
	if err := json.Unmarshal(noSpeedReport, &noSpeedValue); err != nil {
		t.Fatal(err)
	}
	if !noSpeedValue.Attested || noSpeedValue.PairsChecked != 0 || noSpeedValue.ResidualsChecked != 0 || len(noSpeedValue.Defects) != 0 {
		t.Fatalf("disabled continuity checks report = %s", noSpeedReport)
	}
	strictVerdict, err := sp3.ContinuityVerdictJSONWithPolicy(strict, epochs[0], epochs[len(epochs)-1])
	if err != nil {
		t.Fatal(err)
	}
	var strictDecision struct {
		Decision string `json:"decision"`
		Accepted bool   `json:"accepted"`
	}
	if err := json.Unmarshal(strictVerdict, &strictDecision); err != nil {
		t.Fatal(err)
	}
	if strictDecision.Accepted || strictDecision.Decision != "refuse" {
		t.Fatalf("strict continuity verdict = %s", strictVerdict)
	}
	relaxed := strict
	relaxed.ExplicitMaxSpeedMPS = 100000
	relaxedReport, err := sp3.ContinuityReportJSON(relaxed)
	if err != nil {
		t.Fatal(err)
	}
	var relaxedValue struct {
		Attested bool `json:"attested"`
		Defects  []struct {
			Kind string `json:"kind"`
		} `json:"defects"`
	}
	if err := json.Unmarshal(relaxedReport, &relaxedValue); err != nil {
		t.Fatal(err)
	}
	if !relaxedValue.Attested || len(relaxedValue.Defects) != 0 {
		t.Fatalf("relaxed speed policy report = %s", relaxedReport)
	}
	policy.SpeedBoundKind = SP3SpeedBoundExplicit
	policy.ExplicitMaxSpeedMPS = -1
	if _, err := sp3.ContinuityReportJSON(policy); err == nil {
		t.Fatal("negative explicit speed bound unexpectedly succeeded")
	} else {
		var statusErr *StatusError
		if !errors.As(err, &statusErr) || statusErr.Code != StatusInvalidArgument || statusErr.SP3 == nil || statusErr.SP3.Field != SP3ErrorFieldSpeedBound || statusErr.SP3.Reason != SP3ErrorReasonNegative || !statusErr.SP3.HasValue || statusErr.SP3.Value != -1 {
			t.Fatalf("negative bound error = %#v (%v), want typed speed-bound negative value", statusErr, err)
		}
	}
	reportSnapshot := append([]byte(nil), report...)
	verdictSnapshot := append([]byte(nil), verdict...)
	if err := sp3.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSP3([]byte("not an SP3 product")); err == nil {
		t.Fatal("invalid native parse unexpectedly succeeded")
	}
	if !bytes.Equal(report, reportSnapshot) || !bytes.Equal(verdict, verdictSnapshot) {
		t.Fatal("detached continuity JSON changed after close and a failing native call")
	}
}
