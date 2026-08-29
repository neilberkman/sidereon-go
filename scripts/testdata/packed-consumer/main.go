package main

import (
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/neilberkman/sidereon-go"
)

// This is the only implementation-dependent part of the packed consumer.
// It performs a numerical SPP solve over in-memory bytes from the archived
// module fixture.
func runSolve() (err error) {
	sp3Bytes, err := os.ReadFile("trimmed.sp3")
	if err != nil {
		return err
	}
	sp3, err := sidereon.LoadSP3(sp3Bytes)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sp3.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close SP3: %w", closeErr))
		}
	}()

	observations := []struct {
		id   string
		bits uint64
	}{
		{"G08", 0x4176b8c6fd82e861},
		{"G10", 0x4175aa4fa1a0c21f},
		{"G16", 0x417387abd6052c3b},
		{"G18", 0x4174c288f3bd1166},
		{"G20", 0x417443947bd00bd6},
		{"G21", 0x4173d8405cd09f84},
		{"G26", 0x417425d51967e798},
		{"G27", 0x41745a4b78a81707},
	}
	input := make([]sidereon.SPPObservation, len(observations))
	for i, observation := range observations {
		input[i] = sidereon.SPPObservation{
			SatelliteID:  observation.id,
			PseudorangeM: math.Float64frombits(observation.bits),
		}
	}
	solution, err := sidereon.SolveSPP(sp3, sidereon.SPPConfig{
		Observations:    input,
		TRxJ2000S:       646272000.0,
		TRxSecondOfDayS: 43200.0,
		DayOfYear:       176.5,
		InitialGuess:    [4]float64{4.5e6, 0.5e6, 4.5e6, 0},
		Ionosphere:      false,
		Troposphere:     false,
		WithGeodetic:    true,
	})
	if err != nil {
		return err
	}
	wantPosition := [3]uint64{0x41511b07ff83c7f1, 0x4120cd6b5ee8cafe, 0x41511e62229db724}
	for axis, bits := range wantPosition {
		if math.Float64bits(solution.PositionM[axis]) != bits {
			return fmt.Errorf("position[%d] = %.17g, want frozen C value", axis, solution.PositionM[axis])
		}
	}
	if math.Float64bits(solution.ReceiverClockS) != 0x3f1a3b88360a8d78 {
		return fmt.Errorf("receiver clock = %.17g, want frozen C value", solution.ReceiverClockS)
	}
	if solution.UsedSatelliteCount != len(observations) {
		return fmt.Errorf("used satellite count = %d, want %d", solution.UsedSatelliteCount, len(observations))
	}
	fmt.Printf("numerical solve: %v\n", solution)
	return nil
}

func main() {
	if err := runSolve(); err != nil {
		panic(err)
	}
}
