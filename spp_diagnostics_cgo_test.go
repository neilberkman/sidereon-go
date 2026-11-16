//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"strings"
	"testing"
)

func TestPseudorangeRejections(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sp3.Close() }()
	for _, bad := range []struct {
		value  float64
		reason SPPRejectionReason
	}{{math.NaN(), SPPRejectionNonFinitePseudorange}, {math.Inf(1), SPPRejectionNonFinitePseudorange}, {0, SPPRejectionNonPositivePseudorange}, {-1, SPPRejectionNonPositivePseudorange}} {
		input := v2Fixture()
		// A valid observation of G08 remains: duplicate validation counts valid rows only.
		input.Base.Observations = append(input.Base.Observations, SPPObservation{SatelliteID: "G08", PseudorangeM: bad.value})
		solution, err := SolveSPPV2(sp3, input)
		if err != nil {
			t.Fatal(err)
		}
		rejected, err := solution.RejectedSatellitesV2()
		_ = solution.Close()
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rejected {
			if row.SatelliteID == "G08" && row.Reason == bad.reason {
				found = true
			}
		}
		if !found {
			t.Fatalf("SPP rejections=%+v want reason=%v", rejected, bad.reason)
		}
		static, _, err := SolveStaticPositionSP3(sp3, []StaticPositionEpoch{{Inputs: input}, {Inputs: input}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		rejections, err := static.RejectedSatellitesV2(0)
		_ = static.Close()
		if err != nil {
			t.Fatal(err)
		}
		found = false
		for _, row := range rejections {
			if row.SatelliteID == "G08" && uint32(row.Reason) == uint32(bad.reason) {
				found = true
			}
		}
		if !found {
			t.Fatalf("static rejections=%+v", rejections)
		}
	}
}

func TestRINEXSPPMaxGDOP(t *testing.T) {
	nav := readPositioningFixture(t, "nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	obsData := readPositioningFixture(t, "obs/ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx")
	broadcast, err := ParseBroadcastEphemeris(nav)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = broadcast.Close() })
	obs, err := ParseRINEXObservation(obsData)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = obs.Close() })
	validation, err := SolutionValidationOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	validation.HasMaxGDOP, validation.MaxGDOP = true, 0.01
	results, err := SolveSPPFromRINEXObs(broadcast, obs, nil, false, &SPPSolvePolicy{UseValidationOptions: true, Validation: validation})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = results.Close() })
	engineError, err := results.EngineError(0)
	if err != nil || engineError == nil {
		t.Fatalf("GDOP error = %+v, %v", engineError, err)
	}
	assertGDOPRefusal(t, engineError)
	cause := engineError.TypedFields["cause"]
	if cause.Kind != EngineJSONObject || cause.Object["kind"].String != "degenerate_geometry_gdop" {
		t.Fatalf("GDOP cause = %+v", cause)
	}
}

func TestMaxGDOPForwardedOrRefused(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sp3.Close() }()
	options, err := SolutionValidationOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	if options.HasMaxGDOP {
		t.Fatal("GDOP validation must be opt-in")
	}
	options.HasMaxGDOP = true
	options.MaxGDOP = 0.01
	legacy := usedSPPConfig()
	legacy.Validation = &options
	_, err = SolveSPP(sp3, legacy)
	var engine *EngineError
	if !errors.As(err, &engine) || engine.Kind != "degenerate_geometry_gdop" {
		t.Fatalf("legacy GDOP=%+v / %v", engine, err)
	}
	raw := engine.TypedFields["gdop"]
	value, err := (EngineFloat{BitsHex: raw.Object["bits_hex"].String, Decimal: raw.Object["decimal"].String}).Float64()
	if err != nil || math.IsNaN(value) || value <= 0.01 {
		t.Fatalf("structured GDOP=%+v parsed=%g / %v", raw, value, err)
	}
	input := v2Fixture()
	input.Policy = SPPSolvePolicy{UseValidationOptions: true, Validation: options}
	_, err = SolveSPPV2(sp3, input)
	if err == nil {
		t.Fatal("V2 GDOP ceiling silently ignored")
	}
	// The batch entry point accepts only its separate policy. A GDOP request
	// embedded in a row must be refused by C, rather than ignored.
	batch, err := SolveSPPBatchSerial(sp3, []SPPInputsV2{input}, false, SPPSolvePolicy{})
	if batch != nil {
		_ = batch.Close()
	}
	if err == nil {
		t.Fatal("batch row GDOP silently ignored")
	}
	input = v2Fixture()
	input.Base.Validation = &options
	input.Policy.UseValidationOptions = true
	_, err = SolveSPPV2(sp3, input)
	assertUnsupportedGDOP(t, err)
	input.Policy.UseValidationOptions = false
	_, err = SolveSPPV2(sp3, input)
	assertGDOPRefusal(t, err)
	for _, models := range []SPPModelOptions{{}, {QZSSClock: 1}} {
		legacy.Models = models
		_, err = SolveSPP(sp3, legacy)
		assertGDOPRefusal(t, err)
	}
	epoch, err := ExactEpochFromJ2000Seconds(legacy.TRxJ2000S)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, epoch)
	_, err = SolveSPPV2AtExactEpoch(sp3, input, epoch)
	assertGDOPRefusal(t, err)
	fde, err := DefaultFDEOptions()
	if err != nil {
		t.Fatal(err)
	}
	_, err = SolveFDE(sp3, legacy, fde)
	assertFDEGDOPRefusal(t, err)
	robust := SPPRobustConfig{HuberK: 1.345, ScaleFloorM: 1, MaxOuter: 100, OuterToleranceM: .0001}
	_, err = SolveRobustFDE(sp3, legacy, robust, fde)
	assertFDEGDOPRefusal(t, err)
	batch, err = SolveSPPBatchSerial(sp3, []SPPInputsV2{v2Fixture()}, false, SPPSolvePolicy{UseValidationOptions: true, Validation: options})
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, batch)
	row, err := batch.EngineError(0)
	if err != nil {
		t.Fatal(err)
	}
	assertGDOPRefusal(t, row)
	_, err = SolveSPPWithDopplerVelocity(sp3, input, nil)
	assertUnsupportedGDOP(t, err)
	_, _, err = SolveStaticPositionSP3(sp3, []StaticPositionEpoch{{Inputs: input}}, nil)
	assertUnsupportedGDOP(t, err)
	base := dgnssObservations(legacy.Observations)
	_, err = SolveDGNSSPosition(sp3, [3]float64{legacy.InitialGuess[0], legacy.InitialGuess[1], legacy.InitialGuess[2]}, base, base, legacy)
	assertUnsupportedGDOP(t, err)
}

func assertGDOPRefusal(t *testing.T, err error) {
	t.Helper()
	var engine *EngineError
	if !errors.As(err, &engine) {
		t.Fatalf("expected structured GDOP refusal, got %T: %v", err, err)
	}
	node := EngineJSONValue{Kind: EngineJSONObject, Object: map[string]EngineJSONValue{"kind": {Kind: EngineJSONString, String: engine.Kind}, "fields": {Kind: EngineJSONObject, Object: engine.TypedFields}}}
	for node.Object["kind"].String != "degenerate_geometry_gdop" {
		next, ok := node.Object["fields"].Object["cause"]
		if !ok {
			t.Fatalf("missing GDOP cause: %+v", engine)
		}
		node = next
	}
	field, ok := node.Object["fields"].Object["gdop"]
	if !ok {
		t.Fatalf("missing structured gdop: %+v", node)
	}
	v, e := (EngineFloat{BitsHex: field.Object["bits_hex"].String, Decimal: field.Object["decimal"].String}).Float64()
	if e != nil || math.IsNaN(v) || v <= .01 {
		t.Fatalf("gdop=%g: %v", v, e)
	}
}

func assertUnsupportedGDOP(t *testing.T, err error) {
	t.Helper()
	var status *StatusError
	if !errors.As(err, &status) || status.Code != StatusInvalidArgument || !strings.Contains(status.Detail, "max_gdop") {
		t.Fatalf("expected typed max_gdop refusal: %T %v", err, err)
	}
}

func TestBroadcastGDOPPolicies(t *testing.T) {
	b, err := ParseBroadcastEphemeris(readPositioningFixture(t, "nav/ESBC00DNK_R_20201770000_01D_MN.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, b)
	obs, err := ParseRINEXObservation(readPositioningFixture(t, "obs/ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, obs)
	assembled, err := SPPInputsFromRINEXObs(obs, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, assembled)
	input, err := assembled.EpochInputs(0)
	if err != nil {
		t.Fatal(err)
	}
	validation, err := SolutionValidationOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	validation.HasMaxGDOP, validation.MaxGDOP = true, .01
	input.Policy = SPPSolvePolicy{UseValidationOptions: true, Validation: validation}
	epoch, err := ExactEpochFromJ2000Seconds(input.Base.TRxJ2000S)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, epoch)
	_, err = SolveBroadcastV2AtExactEpoch(b, input, epoch)
	assertGDOPRefusal(t, err)
	input.Base.Validation = &validation
	for _, models := range []SPPModelOptions{{}, {QZSSClock: 1}} {
		input.Base.Models = models
		_, err = SolveBroadcast(b, input.Base)
		assertGDOPRefusal(t, err)
	}
	sbas, err := NewSBASCorrectionStore()
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, sbas)
	_, err = sbas.SolveBroadcastV2AtExactEpoch(b, "S23", SBASMixedAugmentation, input, epoch)
	assertGDOPRefusal(t, err)
	_, err = sbas.SolveBroadcastAtExactEpoch(b, "S23", SBASMixedAugmentation, input.Base, epoch)
	assertGDOPRefusal(t, err)
	ssr, err := NewSSRCorrectionStore(SSRReferencePointAntennaPhaseCenter)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, ssr)
	_, err = ssr.SolveBroadcastV2AtExactEpoch(b, input, 120, SSRFallbackToBroadcast, false, 0, SSRCorrectionSizeStrict, epoch)
	assertGDOPRefusal(t, err)
	_, err = ssr.SolveBroadcastAtExactEpoch(b, input.Base, 120, SSRFallbackToBroadcast, false, 0, SSRCorrectionSizeStrict, epoch)
	assertGDOPRefusal(t, err)
	fde, err := DefaultFDEOptions()
	if err != nil {
		t.Fatal(err)
	}
	_, err = SolveFDEBroadcast(b, input.Base, fde)
	assertFDEGDOPRefusal(t, err)
	robust := SPPRobustConfig{HuberK: 1.345, ScaleFloorM: 1, MaxOuter: 100, OuterToleranceM: .0001}
	_, err = SolveRobustFDEBroadcast(b, input.Base, robust, fde)
	assertFDEGDOPRefusal(t, err)
	_, err = SolveBroadcastWithDopplerVelocity(b, input.Base, nil)
	assertUnsupportedGDOP(t, err)
	_, _, err = SolveStaticPositionBroadcast(b, []StaticPositionEpoch{{Inputs: input}}, nil)
	assertUnsupportedGDOP(t, err)
}

func assertFDEGDOPRefusal(t *testing.T, err error) {
	t.Helper()
	assertGDOPRefusal(t, err)
	var engine *EngineError
	if !errors.As(err, &engine) {
		t.Fatal(err)
	}
	node := engine.TypedFields
	if cause, ok := node["cause"]; ok {
		node = cause.Object["fields"].Object
	}
	raw, ok := node["max_gdop"]
	if !ok {
		t.Fatalf("missing max_gdop: %+v", node)
	}
	v, e := (EngineFloat{BitsHex: raw.Object["bits_hex"].String, Decimal: raw.Object["decimal"].String}).Float64()
	if e != nil || v != .01 {
		t.Fatalf("max_gdop=%g: %v", v, e)
	}
}

func TestSPPRTKLIBSingleClockDOP(t *testing.T) {
	sp3, cfg, _, _ := pppFixture(t)
	data, err := os.ReadFile("testdata/ppp-static-known-truth.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Receiver   [3]float64 `json:"receiver_ecef_m"`
		Satellites map[string]struct {
			Position [3]float64 `json:"position_ecef_m"`
		} `json:"satellites"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	input := SPPConfig{DayOfYear: 176, TRxSecondOfDayS: 43200, TRxJ2000S: cfg.Epochs[0].TRxJ2000S, InitialGuess: [4]float64{4500000, 500000, 4500000, 0}}
	var normal [4][8]float64
	for _, obs := range cfg.Epochs[0].Observations {
		input.Observations = append(input.Observations, SPPObservation{SatelliteID: obs.SatelliteID, PseudorangeM: obs.CodeM})
		// Hand computation from independent manifest coordinates. Unit-weight
		// shared-clock design H=[-LOS,1], Q=(H^T H)^-1.
		position := manifest.Satellites[obs.SatelliteID].Position
		d := [3]float64{position[0] - manifest.Receiver[0], position[1] - manifest.Receiver[1], position[2] - manifest.Receiver[2]}
		r := math.Sqrt(d[0]*d[0] + d[1]*d[1] + d[2]*d[2])
		h := [4]float64{-d[0] / r, -d[1] / r, -d[2] / r, 1}
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				normal[i][j] += h[i] * h[j]
			}
		}
	}
	for i := 0; i < 4; i++ {
		normal[i][i+4] = 1
	}
	for col := 0; col < 4; col++ {
		pivot := col
		for i := col + 1; i < 4; i++ {
			if math.Abs(normal[i][col]) > math.Abs(normal[pivot][col]) {
				pivot = i
			}
		}
		normal[col], normal[pivot] = normal[pivot], normal[col]
		scale := normal[col][col]
		if math.Abs(scale) < 1e-12 {
			t.Fatal("independent design singular")
		}
		for j := 0; j < 8; j++ {
			normal[col][j] /= scale
		}
		for i := 0; i < 4; i++ {
			if i != col {
				factor := normal[i][col]
				for j := 0; j < 8; j++ {
					normal[i][j] -= factor * normal[col][j]
				}
			}
		}
	}
	pdop := math.Sqrt(normal[0][4] + normal[1][5] + normal[2][6])
	tdop := math.Sqrt(normal[3][7])
	gdop := math.Hypot(pdop, tdop)
	solution, err := SolveSPP(sp3, input)
	if err != nil {
		t.Fatal(err)
	}
	dop := solution.RTKLIBSingleClockDOP
	if dop == nil || math.Abs(dop.PDOP-pdop) > 1e-8 || math.Abs(dop.TDOP-tdop) > 1e-8 || math.Abs(dop.GDOP-gdop) > 1e-8 {
		t.Fatalf("shared DOP=%+v independent=%g,%g,%g", dop, pdop, tdop, gdop)
	}
	validation, err := SolutionValidationOptionsInit()
	if err != nil {
		t.Fatal(err)
	}
	validation.HasMaxGDOP, validation.MaxGDOP = true, .01
	input.Validation = &validation
	fde, err := DefaultFDEOptions()
	if err != nil {
		t.Fatal(err)
	}
	for _, robust := range []bool{false, true} {
		var solveErr error
		if robust {
			_, solveErr = SolveRobustFDE(sp3, input, SPPRobustConfig{HuberK: 1.345, ScaleFloorM: 1, MaxOuter: 100, OuterToleranceM: .0001}, fde)
		} else {
			_, solveErr = SolveFDE(sp3, input, fde)
		}
		assertFDEGDOPRefusal(t, solveErr)
		var engine *EngineError
		if !errors.As(solveErr, &engine) {
			t.Fatal(solveErr)
		}
		raw := engine.TypedFields["gdop"]
		value, e := (EngineFloat{BitsHex: raw.Object["bits_hex"].String, Decimal: raw.Object["decimal"].String}).Float64()
		if e != nil || math.Abs(value-gdop) > 1e-8 {
			t.Fatalf("FDE gdop=%g, independent H normal matrix=%g: %v", value, gdop, e)
		}
	}

}
