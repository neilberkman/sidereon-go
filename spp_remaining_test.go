package sidereon

import (
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
	wantPosition := [3]uint64{0x41511b07ff83c7f1, 0x4120cd6b5ee8cafe, 0x41511e62229db724}
	for i, bits := range wantPosition {
		want := math.Float64frombits(bits)
		if !closeTol(detached.PositionM[i], want, toleranceM) {
			t.Fatalf("V2 position[%d] = %.17g, want %.17g (bits %#x)", i, detached.PositionM[i], want, bits)
		}
	}
	if want := math.Float64frombits(0x3f1a3b88360a8d78); !closeTol(detached.ReceiverClockS, want, toleranceS) {
		t.Fatalf("V2 clock = %.17g, want %.17g", detached.ReceiverClockS, want)
	}
	wantECEF := [9]uint64{0x40194e10479e6e2d, 0x3fe2cd9c121466d8, 0x400fca5c00638b84, 0x3fe2cd9c121466d8, 0x3ff7605ef09ccae8, 0x3ff0722b57c6f425, 0x400fca5c00638b84, 0x3ff0722b57c6f425, 0x40163709b8e2fc45}
	for i, bits := range wantECEF {
		want := math.Float64frombits(bits)
		if !closeTol(covECEF[i], want, toleranceM2) {
			t.Fatalf("V2 ECEF covariance[%d] = %.17g, want %.17g (bits %#x)", i, covECEF[i], want, bits)
		}
	}
	wantENU := [9]uint64{0x3ff642156f199a69, 0x3fd9157bfce51e76, 0x3fd76c2f083e23b7, 0x3fd9157bfce51e76, 0x3ffe7cdc8fcc0042, 0xbfdaf4c777cc7450, 0x3fd76c2f083e23b4, 0xbfdaf4c777cc7440, 0x402416ba9e779b40}
	for i, bits := range wantENU {
		want := math.Float64frombits(bits)
		if !closeTol(covENU[i], want, toleranceM2) {
			t.Fatalf("V2 ENU covariance[%d] = %.17g, want %.17g (bits %#x)", i, covENU[i], want, bits)
		}
	}
	if len(rejected) != 0 || len(clocks) != 1 || clocks[0].System != 0 ||
		!closeTol(clocks[0].ReceiverClockS, math.Float64frombits(0x3f1a3b88360a8d78), toleranceS) ||
		len(tdops) != 1 || tdops[0].System != 0 ||
		!closeTol(tdops[0].TDOP, math.Float64frombits(0x4005615e49801311), toleranceRatio) {
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
		if !closeArray3(batchDetached.PositionM, detached.PositionM, toleranceM) ||
			!closeTol(batchDetached.ReceiverClockS, math.Float64frombits(0x3f1a3b88360a8d78), toleranceS) ||
			!reflect.DeepEqual(batchDetached.UsedSatelliteIDs, wantIDs) || len(batchDetached.ResidualsM) != len(wantIDs) {
			t.Fatalf("%s batch differs from frozen fixture: %+v", name, batchDetached)
		}
		wantResiduals := []uint64{0xbe95400000000000, 0xbf46fc0000000000, 0xbf1c068000000000, 0x3f378df000000000, 0xbf1deac000000000, 0xbf24d52000000000, 0x3f43165000000000, 0xbf00f98000000000}
		for i, bits := range wantResiduals {
			want := math.Float64frombits(bits)
			if !closeTol(batchDetached.ResidualsM[i], want, toleranceM) {
				t.Fatalf("%s batch residual[%d] = %.17g, want %.17g (bits %#x)", name, i, batchDetached.ResidualsM[i], want, bits)
			}
		}
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
	if !closeArray3(withVelocity.Receiver.PositionM, detached.PositionM, toleranceM) ||
		!closeTol(withVelocity.Receiver.ReceiverClockS, math.Float64frombits(0x3f1a3b88360a8d78), toleranceS) ||
		!reflect.DeepEqual(withVelocity.Receiver.UsedSatelliteIDs, wantIDs) {
		t.Fatalf("precise Doppler receiver = %#v", withVelocity.Receiver)
	}
	velocity := withVelocity.Velocity
	wantVelocity := [3]uint64{0x407762816c3433ee, 0x407653b0906d6edc, 0x40897f68a2ed638c}
	for i, bits := range wantVelocity {
		want := math.Float64frombits(bits)
		if !closeTol(velocity.VelocityMPerS[i], want, toleranceM) {
			t.Fatalf("precise Doppler velocity[%d] = %.17g, want %.17g (bits %#x)", i, velocity.VelocityMPerS[i], want, bits)
		}
	}
	if !closeTol(velocity.ClockDriftSPerS, math.Float64frombits(0x3ec3a73928b414da), toleranceS) ||
		!closeTol(velocity.SpeedMPerS, math.Float64frombits(0x408e30c56d1c1ecc), toleranceM) ||
		velocity.UsedSatelliteCount != 8 || !reflect.DeepEqual(velocity.UsedSatelliteIDs, wantIDs) || len(velocity.ResidualsMPerS) != 8 {
		t.Fatalf("precise Doppler velocity metadata = %#v", velocity)
	}
	wantCovariance := [16]uint64{0x400282c626323d2a, 0xbfd3062ad69da274, 0x3ff3772a277f1155, 0x3e3a581f579bce26, 0xbfd3062ad69da272, 0x3fe06c071ee2a862, 0x3f9d0b77b1d638fe, 0xbdfb3232e1b75c0c, 0x3ff3772a277f115a, 0x3f9d0b77b1d638fe, 0x40004182baec1d33, 0x3e37ad085814be3d, 0x3e3a581f579bce29, 0xbdfb3232e1b75c09, 0x3e37ad085814be3d, 0x3c78b3e333d6e173}
	for i, bits := range wantCovariance {
		want := math.Float64frombits(bits)
		if !closeTol(velocity.StateCovariance[i], want, toleranceM2) {
			t.Fatalf("precise Doppler covariance[%d] = %.17g, want %.17g (bits %#x)", i, velocity.StateCovariance[i], want, bits)
		}
	}
	wantVelocityResiduals := [8]uint64{0xc051210fe8017ce0, 0x407060d9c41d0ad0, 0x4052423516130ed8, 0xc07221aef275ca36, 0x4071bf47da516754, 0x40502e78233187a8, 0xc081a41144ab2b62, 0x406deb911219d100}
	for i, bits := range wantVelocityResiduals {
		want := math.Float64frombits(bits)
		if !closeTol(velocity.ResidualsMPerS[i], want, toleranceM) {
			t.Fatalf("precise Doppler residual[%d] = %.17g, want %.17g (bits %#x)", i, velocity.ResidualsMPerS[i], want, bits)
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
	if !closeTol(itemValue.PositionM[0], math.Float64frombits(0x414b544c18b32cbb), toleranceM) ||
		!closeTol(itemValue.PositionM[1], math.Float64frombits(0x412040dc27ae50c4), toleranceM) ||
		!closeTol(itemValue.PositionM[2], math.Float64frombits(0x4153f61cf5057880), toleranceM) ||
		!closeTol(itemValue.ReceiverClockS, math.Float64frombits(0x3f3f84a55d21f01a), toleranceS) {
		t.Fatalf("RINEX solution position/clock = %#v %.17g", itemValue.PositionM, itemValue.ReceiverClockS)
	}
	if !reflect.DeepEqual(itemValue.UsedSatelliteIDs, []string{"G05", "G07", "G09", "G13", "G15", "G18", "G27", "G28", "G30", "E01", "E03", "E05", "E09", "E15", "E24", "E31", "C05", "C07", "C10", "C19", "C20", "C23", "C32", "C37"}) || len(itemValue.ResidualsM) != 24 {
		t.Fatalf("RINEX solution IDs = %#v", itemValue.UsedSatelliteIDs)
	}
	wantRINEXResiduals := [24]uint64{0xbfaf6afa60000000, 0x3f8fdc2780000000, 0xbfe73c2090000000, 0xbfc05a03f0000000, 0xbfb0ed9b30000000, 0xbfd30856dc000000, 0xbfc6d2d728000000, 0x3ffd24e4ec000000, 0xbfb11c87a0000000, 0x3fbcbcdb20000000, 0x3fd7b4cdec000000, 0x3fbe634800000000, 0xbfcd1324e8000000, 0x3fa60bb2a0000000, 0xbfd00548b4000000, 0x3fbd7b9cc0000000, 0xbfd9221388000000, 0x3fd3f9a628000000, 0x3fca434090000000, 0xbfa7c6c820000000, 0xbfc7f71090000000, 0xbfce425570000000, 0xbfe936faaa000000, 0x3fded539fc000000}
	for i, bits := range wantRINEXResiduals {
		want := math.Float64frombits(bits)
		if !closeTol(itemValue.ResidualsM[i], want, toleranceM) {
			t.Fatalf("RINEX residual[%d] = %.17g, want %.17g (bits %#x)", i, itemValue.ResidualsM[i], want, bits)
		}
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
	if !closeTol(detached.PositionM[0], math.Float64frombits(0x41511b07ff83c7f1), toleranceM) ||
		!closeTol(detached.PositionM[1], math.Float64frombits(0x4120cd6b5ee8cafe), toleranceM) ||
		!closeTol(detached.PositionM[2], math.Float64frombits(0x41511e62229db724), toleranceM) ||
		!closeTol(detached.ReceiverClockS, math.Float64frombits(0x3f1a3b88360a8d78), toleranceS) {
		t.Fatalf("fallback position/clock = %#v %.17g", detached.PositionM, detached.ReceiverClockS)
	}
	if !reflect.DeepEqual(detached.UsedSatelliteIDs, []string{"G08", "G10", "G16", "G18", "G20", "G21", "G26", "G27"}) || len(detached.ResidualsM) != 8 {
		t.Fatalf("fallback IDs/residuals = %#v %#v", detached.UsedSatelliteIDs, detached.ResidualsM)
	}
	wantResiduals := []uint64{0xbe95400000000000, 0xbf46fc0000000000, 0xbf1c068000000000, 0x3f378df000000000, 0xbf1deac000000000, 0xbf24d52000000000, 0x3f43165000000000, 0xbf00f98000000000}
	for i, bits := range wantResiduals {
		want := math.Float64frombits(bits)
		if !closeTol(detached.ResidualsM[i], want, toleranceM) {
			t.Fatalf("fallback residual[%d] = %.17g, want %.17g (bits %#x)", i, detached.ResidualsM[i], want, bits)
		}
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
