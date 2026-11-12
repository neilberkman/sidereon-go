package sidereon

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestPPPVelocityDefaultsPredictionAndEquality(t *testing.T) {
	robust, err := DefaultPPPVelocityRobustConfig()
	if err != nil {
		t.Fatal(err)
	}
	if robust != (PPPVelocityRobustConfig{HuberK: 1.345, ScaleFloorMS: 0.05, MaxOuter: 2, OuterToleranceMS: 0.000001}) {
		t.Fatalf("robust defaults = %#v", robust)
	}
	config, err := DefaultPPPVelocityConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.MinimumObservations != 4 || config.Robust != nil {
		t.Fatalf("velocity defaults = %#v", config)
	}
	state, err := NewStationaryPPPReceiverVelocityState([3]float64{})
	if err != nil {
		t.Fatal(err)
	}
	state.VelocityMS = [3]float64{3, -2, 0.5}
	state.ClockDriftMS = 0.25
	observation := PPPVelocityObservation{SatelliteID: "G01", SatellitePositionM: [3]float64{20200000, 0, 0}, SatelliteVelocityMS: [3]float64{10, 2, 3}, SigmaMS: 0.1, SatelliteClockDriftMS: 0.1}
	prediction, err := PredictPPPRangeRateMS(observation, state)
	if err != nil {
		t.Fatal(err)
	}
	if prediction != (PPPRangeRatePrediction{LOSUnit: [3]float64{1, 0, 0}, RangeRateMS: 7.15}) {
		t.Fatalf("prediction = %#v", prediction)
	}
	equal, err := prediction.Equal(prediction)
	if err != nil || !equal {
		t.Fatalf("prediction equality = %v, %v", equal, err)
	}
	unequal, err := prediction.Equal(PPPRangeRatePrediction{RangeRateMS: 8})
	if err != nil || unequal {
		t.Fatalf("prediction inequality = %v, %v", unequal, err)
	}
}

func TestPPPVelocitySolveAndTypedError(t *testing.T) {
	positions := [][3]float64{{20200000, 0, 0}, {0, 21100000, 0}, {0, 0, 22300000}, {-20900000, 19700000, 21500000}, {18600000, -20400000, 23100000}}
	measured := []float64{-2.75, 2.25, -0.25, 2.796208544343477, -2.754479649598877}
	observations := make([]PPPVelocityObservation, 5)
	for index := range observations {
		observations[index] = PPPVelocityObservation{SatelliteID: "G0" + string(rune('1'+index)), SatellitePositionM: positions[index], MeasuredRangeRateMS: measured[index], SigmaMS: 0.1}
	}
	config, err := DefaultPPPVelocityConfig()
	if err != nil {
		t.Fatal(err)
	}
	solution, err := SolvePPPVelocityLeastSquares(observations, [3]float64{}, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if closeErr := solution.Close(); closeErr != nil {
			t.Errorf("close solution: %v", closeErr)
		}
	})
	values, err := solution.Values()
	if err != nil {
		t.Fatal(err)
	}
	for axis, want := range [3]float64{3, -2, 0.5} {
		if math.Abs(values.VelocityMS[axis]-want) > 1e-12 {
			t.Fatalf("velocity[%d] = %.17g", axis, values.VelocityMS[axis])
		}
	}
	if math.Abs(values.ClockDriftMS-0.25) > 1e-12 || values.ResidualRMSMS >= 1e-12 || values.RobustIterations != 0 || values.HasRobustScale {
		t.Fatalf("values = %#v", values)
	}
	used, err := solution.UsedSatelliteIDs()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(used, []string{"G01", "G02", "G03", "G04", "G05"}) {
		t.Fatalf("used = %v", used)
	}
	downweighted, err := solution.DownweightedSatelliteIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(downweighted) != 0 {
		t.Fatalf("downweighted = %v", downweighted)
	}
	equal, err := solution.Equal(solution)
	if err != nil || !equal {
		t.Fatalf("solution equality = %v, %v", equal, err)
	}
	_, err = SolvePPPVelocityLeastSquares(observations[:3], [3]float64{}, config)
	var status *StatusError
	if !errors.As(err, &status) || status.Engine == nil {
		t.Fatalf("error = %#v", err)
	}
	if status.Engine.Family != EngineErrorFamilyPPPVelocity || status.Engine.Kind != "too_few_observations" {
		t.Fatalf("engine error = %#v", status.Engine)
	}
}
