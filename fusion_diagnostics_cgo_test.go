//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"fmt"
	"math"
	"testing"
)

func TestFusionJointInnovationGates(t *testing.T) {
	for _, kind := range []FusionFilterKind{FusionEKF, FusionUKF} {
		for _, delta := range []float64{1, 100} {
			t.Run(fmt.Sprintf("%d/%g", kind, delta), func(t *testing.T) {
				config, err := DefaultFusionFilterConfig()
				if config.HasLooseJointInnovationGateProbability || config.HasUKFJointInnovationGateProbability || config.HasTightJointInnovationGateProbability {
					t.Fatal("joint gates must be opt-in")
				}
				if err != nil {
					t.Fatal(err)
				}
				config.FilterKind = kind
				if kind == FusionEKF {
					config.HasLooseJointInnovationGateProbability = true
					config.LooseJointInnovationGateProbability = 0.05
				} else {
					config.HasUKFJointInnovationGateProbability = true
					config.UKFJointInnovationGateProbability = 0.05
				}
				diagonal := make([]float64, 15)
				for i := range diagonal {
					diagonal[i] = 1
				}
				filter, err := NewFusionFilter(FusionNavState{PositionECEFM: [3]float64{6378137, 0, 0}, AttitudeBodyToECEF: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}}, diagonal, config)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = filter.Close() }()
				// Literal H=[I_3 0], P=I_15, R=diag(2,3,4), innovation=[d,2,3].
				// Thus S=diag(3,4,5), NIS=d²/3+4/4+9/5, independent of filter output.
				wantNIS := delta*delta/3 + 4.0/4 + 9.0/5
				report, err := filter.UpdateLoose(FusionLooseMeasurement{PositionECEFM: [3]float64{6378137 + delta, 2, 3}, Covariance: []float64{2, 0, 0, 0, 3, 0, 0, 0, 4}, SatellitesUsed: 8, SolutionValid: true, FixStatus: FusionFixFixed})
				if err != nil {
					t.Fatal(err)
				}
				gate := report.JointInnovationGate
				// scipy/mpmath chi-square 3 df ppf(0.95), as in core fixtures.
				const threshold = 7.814727903251179
				if gate == nil || gate.DOF != 3 || gate.Probability != 0.05 || math.Abs(gate.Statistic-wantNIS) > 1e-6 || math.Abs(gate.Threshold-threshold) > 1e-12 || gate.Accepted != (wantNIS <= threshold) || report.Applied != (wantNIS <= threshold) {
					t.Fatalf("joint gate=%+v update=%+v wantNIS=%g", gate, report, wantNIS)
				}
			})
		}
	}
}
