package sidereon

import (
	"math"
	"testing"
)

func sppPlacedPseudorange(t *testing.T, sp3 *SP3, satellite string, receiver ECEF, receive *ExactEpochQuery, seedM float64) float64 {
	t.Helper()
	const c = 299792458.0
	placement := seedM
	for iteration := 0; iteration < 12; iteration++ {
		clockEpoch, err := receive.CheckedSubBinarySeconds(placement / c)
		if err != nil {
			t.Fatalf("clock placement %s: %v", satellite, err)
		}
		transmitClock, err := sp3.TransmitEpochClockAtEpochQueries(clockEpoch, receive, satellite)
		if err != nil || !transmitClock.HasClock {
			_ = clockEpoch.Close()
			t.Fatalf("transmit clock %s: %+v, %v", satellite, transmitClock, err)
		}
		transmitEpoch, err := clockEpoch.CheckedSubBinarySeconds(transmitClock.ClockS)
		_ = clockEpoch.Close()
		if err != nil {
			t.Fatalf("transmit epoch %s: %v", satellite, err)
		}
		state, err := sp3.SourceStateAtEpochQueries(transmitEpoch, receive, satellite)
		if err != nil || !state.HasState {
			_ = transmitEpoch.Close()
			t.Fatalf("source state %s: %+v, %v", satellite, state, err)
		}
		clock := state.ClockS
		relativity, err := sp3.ClockRelativityAtEpochQuery(transmitEpoch, satellite, state.PositionECEFM)
		_ = transmitEpoch.Close()
		if err != nil {
			t.Fatalf("clock relativity %s: %v", satellite, err)
		}
		if relativity.Kind == ClockRelativityTerm {
			clock += relativity.TermS
		} else if relativity.Kind != ClockRelativityNotApplicable {
			t.Fatalf("unexpected clock relativity %s: %+v", satellite, relativity)
		}
		if state.HasGroupDelay {
			clock -= state.GroupDelayS
		}
		dx := state.PositionECEFM[0] - receiver.X
		dy := state.PositionECEFM[1] - receiver.Y
		dz := state.PositionECEFM[2] - receiver.Z
		rangeM := math.Hypot(math.Hypot(dx, dy), dz)
		rangeM += 7.2921151467e-5 * (state.PositionECEFM[0]*receiver.Y - state.PositionECEFM[1]*receiver.X) / c
		updated := rangeM - c*clock
		if !math.IsNaN(updated) && !math.IsInf(updated, 0) && math.Abs(updated-placement) <= 1e-9 {
			return updated
		}
		placement = updated
	}
	t.Fatalf("SPP pseudorange placement %s did not converge from %.17g m", satellite, seedM)
	return 0
}

// TestFDEAcceptsConsistentObservations is a same-library forward/inverse
// consistency control at a fixed receiver and epoch. It is not an independent
// external numerical oracle. It constructs observations from the documented
// SPP placement equation (measured-range clock epoch, source clock correction,
// and RTKLIB first-order Sagnac range), then solves them with identical source
// and receive-epoch semantics.
func TestFDEAcceptsConsistentObservations(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	config := usedSPPConfig()
	receiver := ECEF{X: config.InitialGuess[0], Y: config.InitialGuess[1], Z: config.InitialGuess[2]}
	receive, err := ExactEpochQueryFromBinaryJ2000Seconds(config.TRxJ2000S)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, receive)
	for i := range config.Observations {
		observation := &config.Observations[i]
		observation.PseudorangeM = sppPlacedPseudorange(t, sp3, observation.SatelliteID, receiver, receive, observation.PseudorangeM)
	}

	result, err := SolveFDE(sp3, config, FDEOptions{PFA: 1e-3, WeightsMode: FDEWeightsUnit})
	if err != nil {
		t.Fatalf("consistent synthetic FDE: %T %v", err, err)
	}
	t.Cleanup(func() { _ = result.Close() })
	diagnostics, err := result.Diagnostics()
	if err != nil || diagnostics.Iterations != 0 || len(diagnostics.ExcludedSatelliteIDs) != 0 {
		t.Fatalf("consistent synthetic FDE diagnostics = %+v, %v", diagnostics, err)
	}
	solution, err := result.Solution()
	if err != nil || solution.UsedSatelliteCount != len(config.Observations) {
		t.Fatalf("consistent synthetic FDE solution = %+v, %v", solution, err)
	}
	positionError := math.Hypot(math.Hypot(solution.PositionM[0]-receiver.X, solution.PositionM[1]-receiver.Y), solution.PositionM[2]-receiver.Z)
	if !(positionError <= 1e-3) || !(math.Abs(solution.ReceiverClockS) <= 1e-11) {
		t.Fatalf("consistent synthetic FDE recovered position error %.17g m, receiver clock %.17g s", positionError, solution.ReceiverClockS)
	}
	if len(solution.PseudorangeVariancesM2) != solution.UsedSatelliteCount || len(solution.Weights) != solution.UsedSatelliteCount {
		t.Fatalf("consistent synthetic FDE solution lost aligned weights: %+v", solution)
	}
	accepted, rows, err := result.AcceptedRAIM()
	if err != nil || accepted.FaultDetected || len(rows) != solution.UsedSatelliteCount {
		t.Fatalf("consistent synthetic accepted RAIM = %+v rows=%d err=%v", accepted, len(rows), err)
	}
}
