package sidereon

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func v2Fixture() SPPInputsV2 {
	base := usedSPPConfig()
	return SPPInputsV2{Base: base, Policy: SPPSolvePolicy{UseValidationOptions: false}}
}

func TestSPPV2AndBatchFixture(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	solution, err := SolveSPPV2(sp3, v2Fixture())
	if err != nil {
		t.Fatal(err)
	}
	if solution == nil {
		t.Fatal("SolveSPPV2 returned nil solution")
	}
	t.Cleanup(func() { _ = solution.Close() })
	detached, err := solution.Solution()
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := SolveSPP(sp3, usedSPPConfig())
	if err != nil {
		t.Fatal(err)
	}
	if detached.PositionM != legacy.PositionM || detached.UsedSatelliteIDs[0] != legacy.UsedSatelliteIDs[0] {
		t.Fatalf("V2 solution differs from legacy fixture: %+v / %+v", detached, legacy)
	}
	if _, err := solution.PositionCovarianceECEFM2(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.PositionCovarianceENUM2(); err != nil {
		t.Fatal(err)
	}
	drift, present, err := solution.ReceiverClockDriftSS()
	if err != nil {
		t.Fatal(err)
	}
	if present || drift != 0 {
		t.Fatalf("V2 clock drift = %.17g present=%v", drift, present)
	}
	if _, err := solution.RejectedSatellites(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.SystemClocks(); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.SystemTDOPs(); err != nil {
		t.Fatal(err)
	}
	covECEF, _ := solution.PositionCovarianceECEFM2()
	covENU, _ := solution.PositionCovarianceENUM2()
	rejected, _ := solution.RejectedSatellites()
	clocks, _ := solution.SystemClocks()
	tdops, _ := solution.SystemTDOPs()
	assertSPPIndependentPreciseReference(t, detached.PositionM, detached.ReceiverClockS, detached.ResidualsM)
	// Independently invert the weighted design normal matrix from the pinned
	// RTKLIB transmit states and the documented core variance terms.
	wantECEF := [9]float64{84.03257604763711, -10.2864010948703, 44.412336686725666, -10.286401094870294, 18.6772054268789, 1.371551524068428, 44.41233668672565, 1.3715515240684253, 73.82770776139542}
	for i, want := range wantECEF {
		if math.Abs(covECEF[i]-want) > 1e-3 {
			t.Fatalf("V2 ECEF covariance[%d] = %.17g, independent reference %.17g", i, covECEF[i], want)
		}
	}
	wantENU := [9]float64{22.13633990625892, 9.782899692327842, -15.512058399959887, 9.782899692327844, 32.95212642239692, -3.372867135641211, -15.512058399959889, -3.3728671356411897, 121.44902290725558}
	for i, want := range wantENU {
		if math.Abs(covENU[i]-want) > 1e-3 {
			t.Fatalf("V2 ENU covariance[%d] = %.17g, independent reference %.17g", i, covENU[i], want)
		}
	}
	if len(rejected) != 0 || len(clocks) != 1 || clocks[0].System != 0 || math.Abs(clocks[0].ReceiverClockS-detached.ReceiverClockS) > 1e-15 || len(tdops) != 1 || tdops[0].System != 0 || detached.DOP == nil || math.Abs(tdops[0].TDOP-detached.DOP.TDOP) > 1e-12 {
		t.Fatalf("unexpected V2 per-system outputs: rejected=%#v clocks=%#v tdops=%#v", rejected, clocks, tdops)
	}

	serial, err := SolveSPPBatchSerial(sp3, []SPPInputsV2{v2Fixture()}, false, SPPSolvePolicy{})
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := SolveSPPBatchParallel(sp3, []SPPInputsV2{v2Fixture()}, false, SPPSolvePolicy{})
	if err != nil {
		t.Fatal(err)
	}
	wantIDs := []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}
	for name, batch := range map[string]*SPPBatch{"serial": serial, "parallel": parallel} {
		if batch == nil {
			t.Fatalf("%s batch is nil", name)
		}
		count, err := batch.Count()
		if err != nil || count != 1 {
			t.Fatalf("%s count = %d, %v", name, count, err)
		}
		ok, err := batch.EpochOK(0)
		if err != nil || !ok {
			t.Fatalf("%s epoch ok = %v, %v", name, ok, err)
		}
		message, err := batch.Error(0)
		if err != nil || message != "" {
			t.Fatalf("%s epoch error = %q, %v", name, message, err)
		}
		item, err := batch.Solution(0)
		if err != nil {
			t.Fatalf("%s solution: %v", name, err)
		}
		batchDetached, err := item.Solution()
		if err != nil {
			t.Fatalf("%s detached solution: %v", name, err)
		}
		if batchDetached.PositionM != detached.PositionM || !reflect.DeepEqual(batchDetached.UsedSatelliteIDs, wantIDs) {
			t.Fatalf("%s batch differs from V2 solution: %+v", name, batchDetached)
		}
		assertSPPIndependentPreciseReference(t, batchDetached.PositionM, batchDetached.ReceiverClockS, batchDetached.ResidualsM)
		_ = item.Close()
		_ = batch.Close()
	}
	combined, err := SolveSPPWithDopplerVelocity(sp3, v2Fixture(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if combined.HasVelocity {
		t.Fatalf("empty Doppler result has velocity = %v", combined.HasVelocity)
	}
	if combined.VelocityErrorKind != SPPDopplerVelocityNoError || combined.Velocity != nil {
		t.Fatalf("empty Doppler result = %#v", combined)
	}
	rows := make([]SPPDopplerObservation, 0, len(detached.UsedSatelliteIDs))
	for _, id := range detached.UsedSatelliteIDs {
		rows = append(rows, SPPDopplerObservation{SatelliteID: id, CarrierHz: 1575420000})
	}
	withVelocity, err := SolveSPPWithDopplerVelocity(sp3, v2Fixture(), rows)
	if err != nil {
		t.Fatal(err)
	}
	if !withVelocity.HasVelocity {
		t.Fatalf("fixture Doppler has velocity = %v", withVelocity.HasVelocity)
	}
	if withVelocity.VelocityErrorKind != SPPDopplerVelocityNoError {
		t.Fatalf("fixture Doppler error kind = %d", withVelocity.VelocityErrorKind)
	}
	if withVelocity.Receiver.PositionM == [3]float64{} {
		t.Fatalf("fixture Doppler receiver is empty: %#v", withVelocity.Receiver)
	}
	if withVelocity.Velocity == nil || withVelocity.Velocity.UsedSatelliteCount == 0 {
		t.Fatalf("fixture Doppler velocity is empty: %#v", withVelocity.Velocity)
	}
	if withVelocity.Receiver.PositionM != detached.PositionM || math.Abs(withVelocity.Receiver.ReceiverClockS-detached.ReceiverClockS) > 1e-15 || !reflect.DeepEqual(withVelocity.Receiver.UsedSatelliteIDs, wantIDs) {
		t.Fatalf("precise Doppler receiver = %#v", withVelocity.Receiver)
	}
	velocity := withVelocity.Velocity
	wantVelocity := [3]float64{374.16081325843055, 357.21615594463066, 815.9244622545307}
	for i, want := range wantVelocity {
		if delta := math.Abs(velocity.VelocityMPerS[i] - want); delta > 1e-3 {
			t.Fatalf("precise Doppler velocity[%d] = %.12f m/s, reference %.12f (delta %.6g)", i, velocity.VelocityMPerS[i], want, delta)
		}
	}
	wantClockDrift := 2.342812750204259e-6
	if delta := math.Abs(velocity.ClockDriftSPerS - wantClockDrift); delta > 1e-9 {
		t.Fatalf("precise Doppler clock drift = %.12g s/s, reference %.12g (delta %.6g)", velocity.ClockDriftSPerS, wantClockDrift, delta)
	}
	wantSpeed := 966.0913126363438
	if delta := math.Abs(velocity.SpeedMPerS - wantSpeed); delta > 1e-3 {
		t.Fatalf("precise Doppler speed = %.12f m/s, reference %.12f (delta %.6g)", velocity.SpeedMPerS, wantSpeed, delta)
	}
	if velocity.UsedSatelliteCount != 8 || !reflect.DeepEqual(velocity.UsedSatelliteIDs, wantIDs) || len(velocity.ResidualsMPerS) != 8 {
		t.Fatalf("precise Doppler velocity metadata = %#v", velocity)
	}
	wantCovariance := [16]float64{
		2.3138501017544186, -0.29724975927256736, 1.2165889147076194, 6.133727144216087e-9,
		-0.2972497592725672, 0.5131862947784953, 0.028365868180528164, -3.957493709844634e-10,
		1.2165889147076197, 0.028365868180528025, 2.031984130228662, 5.512462423179669e-9,
		6.133727144216088e-9, -3.9574937098446403e-10, 5.512462423179669e-9, 2.1426104425222244e-17,
	}
	for i, want := range wantCovariance {
		if delta := math.Abs(velocity.StateCovariance[i] - want); delta > 1e-3 {
			t.Fatalf("precise Doppler covariance[%d] = %.12g, reference %.12g (delta %.6g)", i, velocity.StateCovariance[i], want, delta)
		}
	}
	wantVelocityResiduals := [8]float64{-68.51573566095942, 262.04753747815585, 73.03222080511456, -290.10242157914473, 283.9549884028327, 64.72320251877179, -564.4995356713007, 239.35974370652858}
	for i, want := range wantVelocityResiduals {
		if delta := math.Abs(velocity.ResidualsMPerS[i] - want); delta > 0.05 {
			t.Fatalf("precise Doppler residual[%d] = %.12f m/s, reference %.12f (delta %.6g)", i, velocity.ResidualsMPerS[i], want, delta)
		}
	}
}

func TestSPPRINEXAssemblyFixture(t *testing.T) {
	nav := readPositioningFixture(t, "nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	obsData := readPositioningFixture(t, "obs/ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx")
	broadcast, err := ParseBroadcastEphemeris(nav)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = broadcast.Close() }()
	obs, err := ParseRINEXObservation(obsData)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = obs.Close() }()
	inputs, err := SPPInputsFromRINEXObs(obs, broadcast, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = inputs.Close() }()
	count, err := inputs.Count()
	if err != nil || count != 2 {
		t.Fatalf("RINEX SPP input count = %d, want 2 (%v)", count, err)
	}
	inputEpoch, err := inputs.Epoch(0)
	if err != nil {
		t.Fatal(err)
	}
	if inputEpoch.Index != 0 || inputEpoch.ObservationCount != 39 || inputEpoch.Epoch != (CivilDateTime{Year: 2020, Month: 6, Day: 25}) {
		t.Fatalf("RINEX input epoch = %+v", inputEpoch)
	}
	inputValues, err := inputs.EpochInputs(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(inputValues.Base.Observations) != 39 || inputValues.Base.TRxJ2000S == 0 || inputValues.Base.Observations[0].SatelliteID != "G02" || inputValues.Base.Observations[38].SatelliteID != "C37" {
		t.Fatalf("RINEX input values = %+v", inputValues)
	}
	if inputValues.Base.TRxJ2000S != 646315200 || inputValues.Base.TRxSecondOfDayS != 0 || inputValues.Base.DayOfYear != 177 || inputValues.Base.InitialGuess != [4]float64{3582105.291, 532589.7313, 5232754.8054, 0} || !inputValues.Base.Ionosphere || !inputValues.Base.Troposphere || inputValues.Base.WithGeodetic || inputValues.Base.PressureHPA != 1013.25 || inputValues.Base.TemperatureK != 288.15 || inputValues.Base.RelativeHumidity != 0.5 || math.Float64bits(inputValues.Base.Observations[0].PseudorangeM) != 0x4178a663dbeb851f {
		t.Fatalf("RINEX frozen input values = %+v", inputValues.Base)
	}
	inputEpoch1, err := inputs.Epoch(1)
	if err != nil {
		t.Fatal(err)
	}
	if inputEpoch1 != (RINEXSPPEpoch{Index: 1, ObservationCount: 39, Epoch: CivilDateTime{Year: 2020, Month: 6, Day: 25, Second: 30}}) {
		t.Fatalf("RINEX input epoch1 = %+v", inputEpoch1)
	}
	results, err := SolveSPPFromRINEXObs(broadcast, obs, nil, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = results.Close() }()
	resultCount, err := results.Count()
	if err != nil || resultCount != count {
		t.Fatalf("RINEX SPP result count = %d, want %d (%v)", resultCount, count, err)
	}
	resultEpoch, err := results.Epoch(0)
	if err != nil {
		t.Fatal(err)
	}
	if resultEpoch != inputEpoch {
		t.Fatalf("RINEX result epoch = %+v, want %+v", resultEpoch, inputEpoch)
	}
	resultEpoch1, err := results.Epoch(1)
	if err != nil {
		t.Fatal(err)
	}
	if resultEpoch1 != (RINEXSPPEpoch{Index: 1, ObservationCount: 39, Epoch: CivilDateTime{Year: 2020, Month: 6, Day: 25, Second: 30}}) {
		t.Fatalf("RINEX result epoch1 = %+v", resultEpoch1)
	}
	firstSolved := -1
	for i := 0; i < resultCount; i++ {
		ok, err := results.SolutionOK(i)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			firstSolved = i
			break
		}
	}
	if firstSolved >= 0 {
		item, err := results.Solution(firstSolved)
		if err != nil {
			t.Fatal(err)
		}
		_ = item.Close()
	}
	ok, err := results.SolutionOK(0)
	if err != nil {
		t.Fatal(err)
	}
	message, err := results.SolutionError(0)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || message != "" {
		t.Fatalf("RINEX frozen result = ok %v error %q", ok, message)
	}
	engineError, err := results.EngineError(0)
	if err != nil || engineError != nil {
		t.Fatalf("valid RINEX epoch EngineError = %v, %v", engineError, err)
	}
	item, err := results.Solution(0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = item.Close() }()
	itemValue, err := item.Solution()
	if err != nil {
		t.Fatal(err)
	}
	if itemValue.PositionM == [3]float64{} || len(itemValue.UsedSatelliteIDs) == 0 {
		t.Fatalf("RINEX solution = %+v", itemValue)
	}
	// Pinned RTKLIB demo5 on the same 24 selected satellites reports this
	// epoch at [3582103.8919, 532589.8092, 5232755.1223] m. The sub-metre
	// allowance covers the documented core model and solver differences.
	independentRTKLIBPosition := [3]float64{3582103.8919, 532589.8092, 5232755.1223}
	for axis, reference := range independentRTKLIBPosition {
		if delta := math.Abs(itemValue.PositionM[axis] - reference); math.IsNaN(delta) || math.IsInf(delta, 0) || delta > 1 {
			t.Fatalf("RINEX position[%d] = %.12f, pinned RTKLIB reference %.12f (delta %.6g m)", axis, itemValue.PositionM[axis], reference, delta)
		}
	}
	if math.IsNaN(itemValue.ReceiverClockS) || math.IsInf(itemValue.ReceiverClockS, 0) || math.Abs(itemValue.ReceiverClockS) >= 0.001 || !itemValue.Metadata.Converged || itemValue.Metadata.UsedCount != 24 || itemValue.Metadata.SystemCount != 3 {
		t.Fatalf("RINEX solution clock/metadata = %.17g %+v", itemValue.ReceiverClockS, itemValue.Metadata)
	}
	if !reflect.DeepEqual(itemValue.UsedSatelliteIDs, []string{"G05", "G07", "G09", "G13", "G15", "G18", "G27", "G28", "G30", "E01", "E03", "E05", "E09", "E15", "E24", "E31", "C05", "C07", "C10", "C19", "C20", "C23", "C32", "C37"}) || len(itemValue.ResidualsM) != 24 {
		t.Fatalf("RINEX solution IDs = %#v", itemValue.UsedSatelliteIDs)
	}
	for i, residual := range itemValue.ResidualsM {
		if math.IsNaN(residual) || math.IsInf(residual, 0) || math.Abs(residual) > 50 || i >= len(itemValue.PseudorangeVariancesM2) || itemValue.PseudorangeVariancesM2[i] <= 0 {
			t.Fatalf("RINEX residual/variance[%d] = %g / %#v", i, residual, itemValue.PseudorangeVariancesM2)
		}
	}
}

func TestRINEXSPPOwnedRowEngineError(t *testing.T) {
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
		t.Fatalf("SolutionValidationOptionsInit: %v", err)
	}
	validation.HasMaxPDOP = true
	validation.MaxPDOP = 0.01
	policy := &SPPSolvePolicy{UseValidationOptions: true, Validation: validation}
	results, err := SolveSPPFromRINEXObs(broadcast, obs, nil, false, policy)
	if err != nil {
		t.Fatalf("SolveSPPFromRINEXObs with strict validation: %v", err)
	}
	t.Cleanup(func() { _ = results.Close() })
	ok, err := results.SolutionOK(0)
	if err != nil || ok {
		t.Fatalf("strict-validation RINEX epoch success = %v, %v; want failed row", ok, err)
	}
	legacy, err := results.SolutionError(0)
	if err != nil || legacy == "" {
		t.Fatalf("strict-validation RINEX legacy error = %q, %v", legacy, err)
	}
	engineError, err := results.EngineError(0)
	if err != nil {
		t.Fatalf("strict-validation RINEX EngineError: %v", err)
	}
	if engineError == nil || engineError.Family != EngineErrorFamilySppPolicy || engineError.Kind != "validation" || engineError.Operation != "sidereon_rinex_spp_solution_error_payload" || engineError.CaptureError != nil {
		t.Fatalf("strict-validation RINEX EngineError = %+v", engineError)
	}
	cause := engineError.TypedFields["cause"]
	if cause.Kind != EngineJSONObject || cause.Object["kind"].String != "degenerate_geometry_pdop" {
		t.Fatalf("strict-validation RINEX EngineError cause = %+v", cause)
	}
}

func TestSPPFallbackFixtureCall(t *testing.T) {
	nav := readPositioningFixture(t, "nav/ESBC00DNK_R_20201770000_01D_MN.rnx")
	broadcast, err := ParseBroadcastEphemeris(nav)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = broadcast.Close() }()
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sp3.Close() }()
	result, err := SolveWithFallback([]*SP3{sp3}, broadcast, usedSPPConfig(), StalenessPolicyDefault())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = result.Close() }()
	detached, err := result.Solution()
	if err != nil {
		t.Fatal(err)
	}
	assertSPPIndependentPreciseReference(t, detached.PositionM, detached.ReceiverClockS, detached.ResidualsM)
	if !reflect.DeepEqual(detached.UsedSatelliteIDs, []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}) || len(detached.ResidualsM) != 8 {
		t.Fatalf("fallback IDs/residuals = %#v %#v", detached.UsedSatelliteIDs, detached.ResidualsM)
	}
	_, err = SolveWithFallback(nil, broadcast, SPPConfig{}, StalenessPolicyDefault())
	if err == nil {
		t.Fatal("invalid fallback solve unexpectedly succeeded")
	}
	var fallbackErr *FallbackError
	if !errors.As(err, &fallbackErr) {
		t.Fatalf("fallback error = %T %v, want *FallbackError", err, err)
	}
	if fallbackErr.Status != FallbackBroadcastSolve || fallbackErr.Detail == "" || !strings.Contains(fallbackErr.Error(), "broadcast solve") || !strings.Contains(fallbackErr.Error(), "status 6") || !strings.Contains(fallbackErr.Error(), fallbackErr.Detail) {
		t.Fatalf("fallback typed error = %#v (%v)", fallbackErr, fallbackErr)
	}
}

func TestSPPRemainingOwnershipAndBoundaries(t *testing.T) {
	var zero SPPSolutionHandle
	var batch SPPBatch
	var inputs RINEXSPPInputs
	var solutions RINEXSPPSolutions
	if err := zero.Close(); err != nil {
		t.Fatal(err)
	}
	if err := batch.Close(); err != nil {
		t.Fatal(err)
	}
	if err := inputs.Close(); err != nil {
		t.Fatal(err)
	}
	if err := solutions.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := (&SPPBatch{}).EpochOK(-1); err == nil {
		t.Fatal("negative batch index accepted")
	}
	if _, err := (&SPPSolutionHandle{}).PositionCovarianceECEFM2(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero solution covariance error = %v", err)
	}
	if _, err := RINEXSPPOptionsInit(); err != nil {
		t.Fatal(err)
	}
	if _, err := SPPInputsV2Init(); err != nil {
		t.Fatal(err)
	}
}

func TestSPPV2CloseReadRace(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sp3.Close() }()
	solution, err := SolveSPPV2(sp3, v2Fixture())
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = solution.PositionCovarianceECEFM2() }()
	}
	wg.Add(1)
	go func() { defer wg.Done(); _ = solution.Close(); _ = solution.Close() }()
	wg.Wait()
	_, err = solution.SystemClocks()
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("post-close read error = %v", err)
	}
}

func TestSPPBatchOwnedRowEngineErrors(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	goodEpoch := v2Fixture()
	badEpoch := v2Fixture()
	badEpoch.Base.Observations = badEpoch.Base.Observations[:2] // Triggers TooFewSatellites { used: 2, required: 4 }

	// Mixed successful and real failed epoch sequence in row order
	epochs := []SPPInputsV2{goodEpoch, badEpoch, goodEpoch}

	serialBatch, err := SolveSPPBatchSerial(sp3, epochs, false, SPPSolvePolicy{})
	if err != nil {
		t.Fatalf("SolveSPPBatchSerial: %v", err)
	}
	t.Cleanup(func() { _ = serialBatch.Close() })

	parallelBatch, err := SolveSPPBatchParallel(sp3, epochs, false, SPPSolvePolicy{})
	if err != nil {
		t.Fatalf("SolveSPPBatchParallel: %v", err)
	}
	t.Cleanup(func() { _ = parallelBatch.Close() })

	for mode, batch := range map[string]*SPPBatch{"serial": serialBatch, "parallel": parallelBatch} {
		count, err := batch.Count()
		if err != nil || count != 3 {
			t.Fatalf("%s: batch count = %d, %v (want 3)", mode, count, err)
		}

		// 1. Successnil detail on successful epoch (index 0)
		ok0, err := batch.EpochOK(0)
		if err != nil || !ok0 {
			t.Fatalf("%s: epoch 0 ok = %v, %v", mode, ok0, err)
		}
		engErr0, err := batch.EngineError(0)
		if err != nil || engErr0 != nil {
			t.Fatalf("%s: epoch 0 EngineError = %v, %v (want nil, nil)", mode, engErr0, err)
		}
		errText0, err := batch.Error(0)
		if err != nil || errText0 != "" {
			t.Fatalf("%s: epoch 0 Error = %q, %v (want empty, nil)", mode, errText0, err)
		}
		sol0, err := batch.Solution(0)
		if err != nil || sol0 == nil {
			t.Fatalf("%s: epoch 0 Solution = %v, %v", mode, sol0, err)
		}
		_ = sol0.Close()

		// Successnil detail on successful epoch (index 2)
		ok2, err := batch.EpochOK(2)
		if err != nil || !ok2 {
			t.Fatalf("%s: epoch 2 ok = %v, %v", mode, ok2, err)
		}
		engErr2, err := batch.EngineError(2)
		if err != nil || engErr2 != nil {
			t.Fatalf("%s: epoch 2 EngineError = %v, %v (want nil, nil)", mode, engErr2, err)
		}

		// 2. Real failed epoch (index 1)
		ok1, err := batch.EpochOK(1)
		if err != nil || ok1 {
			t.Fatalf("%s: epoch 1 ok = %v, %v (want false, nil)", mode, ok1, err)
		}
		originalText, err := batch.Error(1)
		if err != nil {
			t.Fatalf("%s: epoch 1 Error getter failed: %v", mode, err)
		}
		wantOriginalText := "SPP solve failed: only 2 usable satellites; need at least 4 (3 position + 1 clock per GNSS)"
		if originalText != wantOriginalText {
			t.Fatalf("%s: epoch 1 Error text = %q, want %q", mode, originalText, wantOriginalText)
		}
		repeatedErrorText, err := batch.Error(1)
		if err != nil || repeatedErrorText != originalText {
			t.Fatalf("%s: repeated Error getter = %q, want %q (%v)", mode, repeatedErrorText, originalText, err)
		}

		engErr1, err := batch.EngineError(1)
		if err != nil || engErr1 == nil {
			t.Fatalf("%s: epoch 1 EngineError = %v, %v", mode, engErr1, err)
		}

		// Verify failure cause actual fields
		if engErr1.Family != EngineErrorFamilySpp {
			t.Errorf("%s: family = %d (%s), want %d (EngineErrorFamilySpp)", mode, engErr1.Family, engErr1.Family.Name(), EngineErrorFamilySpp)
		}
		if engErr1.FamilyName != "spp" {
			t.Errorf("%s: family name = %q, want \"spp\"", mode, engErr1.FamilyName)
		}
		if engErr1.Schema != 1 {
			t.Errorf("%s: schema = %d, want 1", mode, engErr1.Schema)
		}
		if engErr1.Kind != "too_few_satellites" {
			t.Errorf("%s: kind = %q, want \"too_few_satellites\"", mode, engErr1.Kind)
		}
		if !strings.Contains(engErr1.Operation, "epoch 1") {
			t.Errorf("%s: operation = %q, want operation containing \"epoch 1\"", mode, engErr1.Operation)
		}
		if engErr1.CaptureError != nil {
			t.Errorf("%s: unexpected CaptureError: %v", mode, engErr1.CaptureError)
		}

		var fields struct {
			Used     int `json:"used"`
			Required int `json:"required"`
		}
		if err := engErr1.UnmarshalFields(&fields); err != nil {
			t.Fatalf("%s: UnmarshalFields failed: %v", mode, err)
		}
		if fields.Used != 2 || fields.Required != 4 {
			t.Fatalf("%s: unexpected fields: used=%d required=%d (want 2, 4)", mode, fields.Used, fields.Required)
		}

		// 3. Failed Solution(1) returns existing StatusError with owned row Engine detail attached
		sol1, solErr := batch.Solution(1)
		if solErr == nil {
			t.Fatalf("%s: expected Solution(1) to fail", mode)
		}
		if sol1 != nil {
			_ = sol1.Close()
			t.Fatalf("%s: expected nil solution on failure", mode)
		}

		// Bounded acyclic walk before errors.Is / As
		assertAcyclicGraph(t, solErr, 50)

		var statusErr *StatusError
		if !errors.As(solErr, &statusErr) {
			t.Fatalf("%s: expected *StatusError via errors.As, got %T: %v", mode, solErr, solErr)
		}
		if statusErr.Code != StatusSolve {
			t.Errorf("%s: status code = %d, want %d (StatusSolve)", mode, statusErr.Code, StatusSolve)
		}
		wantPrefix := "sidereon_spp_batch_solution: epoch 1 did not solve: "
		if !strings.HasPrefix(statusErr.Detail, wantPrefix) {
			t.Fatalf("%s: StatusError.Detail %q missing required prefix %q", mode, statusErr.Detail, wantPrefix)
		}
		if statusErr.Detail != wantPrefix+originalText {
			t.Fatalf("%s: StatusError.Detail = %q, want %q", mode, statusErr.Detail, wantPrefix+originalText)
		}

		var fromSolEngErr *EngineError
		if !errors.As(solErr, &fromSolEngErr) {
			t.Fatalf("%s: expected *EngineError via errors.As from failed Solution", mode)
		}
		if fromSolEngErr.Family != EngineErrorFamilySpp || fromSolEngErr.Kind != "too_few_satellites" {
			t.Errorf("%s: attached engine error mismatch: %+v", mode, fromSolEngErr)
		}
		if statusErr.EngineError() != fromSolEngErr {
			t.Errorf("%s: statusErr.EngineError() = %v, want %v", mode, statusErr.EngineError(), fromSolEngErr)
		}

		// Verify missing target terminates safely
		if errors.Is(solErr, ErrClosed) {
			t.Errorf("%s: unexpected errors.Is(solErr, ErrClosed)", mode)
		}
		var absent *absentTestError
		if errors.As(solErr, &absent) {
			t.Errorf("%s: unexpected errors.As for absentTestError", mode)
		}

		// 4. Repeated accessor owned raw byte mutation doesn't change stored record
		savedPayload := append([]byte(nil), engErr1.Payload...)
		savedFields := append([]byte(nil), engErr1.Fields...)

		if len(engErr1.Payload) > 0 {
			engErr1.Payload[0] ^= 0xFF
		}
		if len(engErr1.Fields) > 0 {
			engErr1.Fields[0] ^= 0xFF
		}

		repeatedEngErr, err := batch.EngineError(1)
		if err != nil || repeatedEngErr == nil {
			t.Fatalf("%s: repeated EngineError call failed: %v", mode, err)
		}
		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) {
			t.Fatalf("%s: repeated accessor returned mutated payload bytes", mode)
		}
		if !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeated accessor returned mutated fields bytes", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) {
			t.Fatalf("%s: fromSolEngErr payload bytes mutated", mode)
		}
		if !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr fields bytes mutated", mode)
		}

		// 5. Unrelated new producer / TLS reset then close batch; verify copied full bytes persist after EACH step
		ClearEngineError()
		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) || !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeatedEngErr corrupted after ClearEngineError", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) || !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr corrupted after ClearEngineError", mode)
		}

		floatCycles := []float64{1.0, 2.0}
		singularCov := []float64{0.0, 0.0, 0.0, 0.0}
		ilsRes, ilsErr := LambdaILS(floatCycles, singularCov, 3.0)
		if ilsErr == nil {
			t.Fatalf("%s: expected LambdaILS with singular covariance to fail, got result: %+v", mode, ilsRes)
		}
		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) || !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeatedEngErr corrupted after unrelated LambdaILS failure", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) || !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr corrupted after unrelated LambdaILS failure", mode)
		}

		if err := batch.Close(); err != nil {
			t.Fatalf("%s: batch Close failed: %v", mode, err)
		}

		if !bytes.Equal(repeatedEngErr.Payload, savedPayload) || !bytes.Equal(repeatedEngErr.Fields, savedFields) {
			t.Fatalf("%s: repeatedEngErr corrupted after batch close", mode)
		}
		if !bytes.Equal(fromSolEngErr.Payload, savedPayload) || !bytes.Equal(fromSolEngErr.Fields, savedFields) {
			t.Fatalf("%s: fromSolEngErr corrupted after batch close", mode)
		}
		if repeatedEngErr.Family != EngineErrorFamilySpp || repeatedEngErr.Kind != "too_few_satellites" || repeatedEngErr.Schema != 1 || repeatedEngErr.Operation != engErr1.Operation {
			t.Fatalf("%s: copied record corrupted after batch close: %+v", mode, repeatedEngErr)
		}
		if fromSolEngErr.Family != EngineErrorFamilySpp || fromSolEngErr.Kind != "too_few_satellites" || fromSolEngErr.Schema != 1 || fromSolEngErr.Operation != engErr1.Operation {
			t.Fatalf("%s: attached sol record corrupted after batch close: %+v", mode, fromSolEngErr)
		}

		// Verify closed batch methods return ErrClosed
		if _, err := batch.EngineError(1); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close EngineError error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.Solution(0); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close Solution error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.EpochOK(0); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close EpochOK error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.Error(0); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close Error error = %v, want ErrClosed", mode, err)
		}
		if _, err := batch.Count(); !errors.Is(err, ErrClosed) {
			t.Errorf("%s: post-close Count error = %v, want ErrClosed", mode, err)
		}
	}

	// 6. Test invalid index semantics on a live batch
	liveBatch, err := SolveSPPBatchSerial(sp3, []SPPInputsV2{goodEpoch}, false, SPPSolvePolicy{})
	if err != nil {
		t.Fatalf("SolveSPPBatchSerial: %v", err)
	}
	t.Cleanup(func() { _ = liveBatch.Close() })

	if _, err := liveBatch.EngineError(-1); err == nil {
		t.Errorf("expected error on negative index -1")
	}
	if _, err := liveBatch.EngineError(999); err == nil {
		t.Errorf("expected error on out-of-range index 999")
	}
	var invStatusErr *StatusError
	if _, err := liveBatch.EngineError(999); !errors.As(err, &invStatusErr) || invStatusErr.Code != StatusInvalidArgument {
		t.Errorf("expected StatusInvalidArgument on out-of-range index 999, got %v", err)
	}
}
