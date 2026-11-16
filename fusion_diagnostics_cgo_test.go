//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
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

func fusionGateConfig(t *testing.T, kind FusionFilterKind, tight, joint bool) FusionFilterConfig {
	t.Helper()
	c, err := DefaultFusionFilterConfig()
	if err != nil {
		t.Fatal(err)
	}
	c.FilterKind = kind
	if joint {
		if kind == FusionUKF {
			c.HasUKFJointInnovationGateProbability = true
			c.UKFJointInnovationGateProbability = .01
		} else if tight {
			c.HasTightJointInnovationGateProbability = true
			c.TightJointInnovationGateProbability = .01
		} else {
			c.HasLooseJointInnovationGateProbability = true
			c.LooseJointInnovationGateProbability = .01
		}
	} else {
		g := &FusionInnovationGate{ThresholdSigma: 3, MinRows: 1}
		if kind == FusionUKF {
			c.UKFInnovationGate = g
		} else if tight {
			c.TightInnovationGate = g
		} else {
			c.LooseInnovationGate = g
		}
	}
	return c
}

func fusionLiteralFilter(t *testing.T, c FusionFilterConfig, epoch, variance float64, positions ...[3]float64) *FusionFilter {
	t.Helper()
	diagonal := make([]float64, 15)
	for i := range diagonal {
		diagonal[i] = variance
	}
	position := [3]float64{6378137, 0, 0}
	if len(positions) != 0 {
		position = positions[0]
	}
	f, err := NewFusionFilter(FusionNavState{EpochJ2000S: epoch, PositionECEFM: position, AttitudeBodyToECEF: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}}, diagonal, c)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, f)
	return f
}

func fusionConstantSource(t *testing.T) (*SP3, float64) {
	t.Helper()
	// Axial satellite G08=(0,0,4096000) m, clock=0. SP3 node rotation
	// leaves these coordinates invariant and the relativistic clock term is zero.
	data, err := os.ReadFile("testdata/ppp-static-known-truth.sp3")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "PG08") {
			lines[i] = fmt.Sprintf("PG08%14.6f%14.6f%14.6f%14.6f", 0., 0., 4096., 0.) + "\nVG08      0.000000      0.000000      0.000000      0.000000"
		}
	}
	sp3, err := LoadSP3([]byte(strings.Join(lines, "\n")))
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, sp3)
	epochs, err := sp3.Epochs()
	if err != nil {
		t.Fatal(err)
	}
	return sp3, epochs[4]
}

func fusionClockEpoch(epoch float64, innovation [2]float64) FusionTightEpoch {
	return FusionTightEpoch{EpochJ2000S: epoch, Observations: []FusionTightObservation{{SatelliteID: "G08", PseudorangeM: 32 + innovation[0], SigmaM: 1, HasRangeRate: true, RangeRate: FusionTightRangeRate{MeasuredMPerS: innovation[1], SigmaMPerS: 2}}}}
}

func TestFusionTightJointGateLiteralClockModel(t *testing.T) {
	sp3, epoch := fusionConstantSource(t)
	for _, kind := range []FusionFilterKind{FusionEKF, FusionUKF} {
		for _, innovation := range [][2]float64{{1, 2}, {8, 9}} {
			t.Run(fmt.Sprintf("%d/%v", kind, innovation), func(t *testing.T) {
				c := fusionGateConfig(t, kind, true, true)
				c.TightModel = &FusionTightModelOptions{InitialClockBiasVarianceM2: 2, InitialClockDriftVarianceM2PerS2: 3}
				f := fusionLiteralFilter(t, c, epoch, 0, [3]float64{0, 0, 4095968})
				before, err := f.State()
				if err != nil {
					t.Fatal(err)
				}
				p, err := f.Covariance()
				if err != nil {
					t.Fatal(err)
				}
				update, _, err := f.UpdateTightSP3(fusionClockEpoch(epoch, innovation), sp3, nil, FusionTightDirect)
				if err != nil {
					t.Fatal(err)
				}
				// Core tight.rs joint_gate_tight_clock_rows_match_literal_linear_model:
				// H=I2, P=diag(2,3), R=diag(1,4), S=diag(3,7).
				want := innovation[0]*innovation[0]/3 + innovation[1]*innovation[1]/7
				const threshold = 9.21034037197618
				accepted := want <= threshold
				g := update.JointInnovationGate
				tolerance := 512 * (math.Nextafter(1, 2) - 1) * math.Max(want, 1)
				if g == nil || math.Abs(g.Statistic-want) > tolerance || math.Abs(g.Threshold-threshold) > 8*(math.Nextafter(1, 2)-1)*threshold || g.DOF != 2 || g.Probability != .01 || g.Accepted != accepted || update.Applied != accepted || update.Rows != 2 || update.InnovationGate != nil {
					t.Fatalf("update=%+v joint=%+v wantNIS=%g", update, g, want)
				}
				state, err := f.State()
				if err != nil {
					t.Fatal(err)
				}
				q, err := f.Covariance()
				if err != nil {
					t.Fatal(err)
				}
				if !accepted {
					if state != before || !reflect.DeepEqual(q, p) || update.AcceptedRows != 0 || update.RejectedRows != 2 {
						t.Fatal("rejected tight update changed state or covariance")
					}
				} else {
					if update.AcceptedRows != 2 || update.RejectedRows != 0 || math.Abs(state.TightClockBiasM-2*innovation[0]/3) > tolerance || math.Abs(state.TightClockDriftMPerS-3*innovation[1]/7) > tolerance || math.Abs(state.TightClockCovariance[0]-2./3) > tolerance || math.Abs(state.TightClockCovariance[3]-12./7) > tolerance {
						t.Fatalf("clock=%+v", state)
					}
				}
			})
		}
	}
}

func TestFusionIGGAllRowsCoast(t *testing.T) {
	for _, kind := range []FusionFilterKind{FusionEKF, FusionUKF} {
		for _, joint := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/joint=%v", kind, joint), func(t *testing.T) {
				c := fusionGateConfig(t, kind, false, joint)
				c.LooseMeasurementReweighting = &FusionIggIIIReweighting{K0Sigma: 1.5, K1Sigma: 3}
				f := fusionLiteralFilter(t, c, 0, 1)
				before, err := f.State()
				if err != nil {
					t.Fatal(err)
				}
				p, err := f.Covariance()
				if err != nil {
					t.Fatal(err)
				}
				// H=[I6 0], P=I15, R=I6, innovation=[50;6]. Each normalized row is 50/sqrt(2)>k1.
				covariance := make([]float64, 36)
				for i := 0; i < 6; i++ {
					covariance[i*6+i] = 1
				}
				u, err := f.UpdateLoose(FusionLooseMeasurement{PositionECEFM: [3]float64{6378187, 50, 50}, HasVelocity: true, VelocityECEFMPerS: [3]float64{50, 50, 50}, Covariance: covariance, SatellitesUsed: 8, SolutionValid: true, FixStatus: FusionFixFixed})
				if err != nil {
					t.Fatal(err)
				}
				if u.Applied || u.Rows != 6 || u.AcceptedRows != 0 || u.RejectedRows != 6 || u.JointInnovationGate != nil {
					t.Fatalf("coast=%+v", u)
				}
				if joint {
					if u.InnovationGate != nil {
						t.Fatalf("unexpected row report %+v", u.InnovationGate)
					}
				} else {
					g := u.InnovationGate
					if g == nil || !g.Coasted || g.InputRows != 6 || g.AcceptedRows != 0 || g.RejectedRows != 6 {
						t.Fatalf("gate=%+v", g)
					}
				}
				after, err := f.State()
				if err != nil {
					t.Fatal(err)
				}
				q, err := f.Covariance()
				if err != nil {
					t.Fatal(err)
				}
				if before != after || !reflect.DeepEqual(p, q) {
					t.Fatal("coast modified state or covariance")
				}
			})
		}
	}
}

func TestFusionGateOptionRefusals(t *testing.T) {
	configs := map[string]FusionFilterConfig{}
	for _, tight := range []bool{false, true} {
		c := fusionGateConfig(t, FusionUKF, tight, true)
		c.HasUKFJointInnovationGateProbability = false
		if tight {
			c.HasTightJointInnovationGateProbability = true
			c.TightJointInnovationGateProbability = .01
		} else {
			c.HasLooseJointInnovationGateProbability = true
			c.LooseJointInnovationGateProbability = .01
		}
		configs[fmt.Sprintf("EKF-joint-on-UKF/%v", tight)] = c
		c = fusionGateConfig(t, FusionEKF, tight, false)
		c.FilterKind = FusionUKF
		configs[fmt.Sprintf("EKF-row-on-UKF/%v", tight)] = c
	}
	c := fusionGateConfig(t, FusionUKF, false, true)
	c.FilterKind = FusionEKF
	configs["UKF-joint-on-EKF"] = c
	c = fusionGateConfig(t, FusionUKF, false, false)
	c.FilterKind = FusionEKF
	configs["UKF-row-on-EKF"] = c
	for _, kind := range []FusionFilterKind{FusionEKF, FusionUKF} {
		for _, tight := range []bool{false, true} {
			for _, probability := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0, 1, -.1, 1.1} {
				c = fusionGateConfig(t, kind, tight, true)
				if kind == FusionUKF {
					c.UKFJointInnovationGateProbability = probability
				} else if tight {
					c.TightJointInnovationGateProbability = probability
				} else {
					c.LooseJointInnovationGateProbability = probability
				}
				configs[fmt.Sprintf("probability/%d/%v/%g", kind, tight, probability)] = c
			}
			c = fusionGateConfig(t, kind, tight, true)
			g := &FusionInnovationGate{ThresholdSigma: 3, MinRows: 1}
			if kind == FusionUKF {
				c.UKFInnovationGate = g
			} else if tight {
				c.TightInnovationGate = g
			} else {
				c.LooseInnovationGate = g
			}
			configs[fmt.Sprintf("exclusive/%d/%v", kind, tight)] = c
		}
	}
	for name, c := range configs {
		t.Run(name, func(t *testing.T) {
			f, err := NewFusionFilter(FusionNavState{PositionECEFM: [3]float64{6378137, 0, 0}, AttitudeBodyToECEF: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1}}, make([]float64, 15), c)
			if f != nil {
				_ = f.Close()
			}
			var engine *EngineError
			if !errors.As(err, &engine) || engine.Family != EngineErrorFamilyFusion || engine.Kind != "invalid_input" {
				t.Fatalf("refusal=%T %v engine=%+v", err, err, engine)
			}
		})
	}
}

func TestFusionRejectedLoosePreservesTightCrossCovariance(t *testing.T) {
	sp3, epoch := fusionConstantSource(t)
	for _, kind := range []FusionFilterKind{FusionEKF, FusionUKF} {
		t.Run(fmt.Sprint(kind), func(t *testing.T) {
			c := fusionGateConfig(t, kind, false, true)
			c.TightModel = &FusionTightModelOptions{InitialClockBiasVarianceM2: 2, InitialClockDriftVarianceM2PerS2: 3}
			control := fusionLiteralFilter(t, c, epoch, 1, [3]float64{0, 0, 4095968})
			screened := fusionLiteralFilter(t, c, epoch, 1, [3]float64{0, 0, 4095968})
			for _, f := range []*FusionFilter{control, screened} {
				if report, _, err := f.UpdateTightSP3(fusionClockEpoch(epoch, [2]float64{1, 2}), sp3, nil, FusionTightDirect); err != nil || !report.Applied {
					t.Fatalf("tight report=%+v error=%v", report, err)
				}
				if err := f.Propagate(FusionIMUSample{EpochJ2000S: epoch + 1./1024, Kind: FusionIMUIncrement, DTS: 1. / 1024}); err != nil {
					t.Fatal(err)
				}
			}
			u, err := screened.UpdateLoose(FusionLooseMeasurement{EpochJ2000S: epoch + 1./1024, PositionECEFM: [3]float64{7378137, 1e6, 1e6}, Covariance: []float64{1, 0, 0, 0, 1, 0, 0, 0, 1}, SatellitesUsed: 8, SolutionValid: true, FixStatus: FusionFixFixed})
			if err != nil {
				t.Fatal(err)
			}
			if u.Applied || u.JointInnovationGate == nil || u.JointInnovationGate.Accepted {
				t.Fatalf("loose=%+v", u)
			}
			for _, f := range []*FusionFilter{control, screened} {
				if err := f.Propagate(FusionIMUSample{EpochJ2000S: epoch + 2./1024, Kind: FusionIMUIncrement, DTS: 1. / 1024}); err != nil {
					t.Fatal(err)
				}
				if report, _, err := f.UpdateTightSP3(fusionClockEpoch(epoch+2./1024, [2]float64{1, 2}), sp3, nil, FusionTightDirect); err != nil || !report.Applied {
					t.Fatalf("tight report=%+v error=%v", report, err)
				}
			}
			a, err := control.State()
			if err != nil {
				t.Fatal(err)
			}
			b, err := screened.State()
			if err != nil {
				t.Fatal(err)
			}
			p, err := control.Covariance()
			if err != nil {
				t.Fatal(err)
			}
			q, err := screened.Covariance()
			if err != nil {
				t.Fatal(err)
			}
			if a != b || !reflect.DeepEqual(p, q) {
				t.Fatalf("rejected loose update changed subsequent tight result: control=%+v screened=%+v", a, b)
			}
		})
	}
}

func TestFusionPerRowGateReports(t *testing.T) {
	for _, kind := range []FusionFilterKind{FusionEKF, FusionUKF} {
		for _, delta := range []float64{1, 100} {
			t.Run(fmt.Sprintf("%d/%g", kind, delta), func(t *testing.T) {
				c := fusionGateConfig(t, kind, false, false)
				if kind == FusionEKF {
					c.LooseInnovationGate.MinRows = 3
				} else {
					c.UKFInnovationGate.MinRows = 3
				}
				f := fusionLiteralFilter(t, c, 0, 1)
				// H=[I3 0], P=I15, R=diag(2,3,4), S=diag(3,4,5).
				// Only row 0 exceeds three sigma when delta=100; MinRows=3 then requires coasting.
				accepted := 3
				rejected := 0
				coasted := delta/math.Sqrt(3) > 3
				if coasted {
					accepted = 2
					rejected = 1
				}
				before, err := f.State()
				if err != nil {
					t.Fatal(err)
				}
				p, err := f.Covariance()
				if err != nil {
					t.Fatal(err)
				}
				u, err := f.UpdateLoose(FusionLooseMeasurement{PositionECEFM: [3]float64{6378137 + delta, 2, 3}, Covariance: []float64{2, 0, 0, 0, 3, 0, 0, 0, 4}, SatellitesUsed: 8, SolutionValid: true, FixStatus: FusionFixFixed})
				if err != nil {
					t.Fatal(err)
				}
				g := u.InnovationGate
				if g == nil || g.Coasted != coasted || g.InputRows != 3 || g.AcceptedRows != accepted || g.RejectedRows != rejected || u.Applied == coasted || u.JointInnovationGate != nil {
					t.Fatalf("update=%+v row=%+v", u, g)
				}
				if coasted {
					after, err := f.State()
					if err != nil {
						t.Fatal(err)
					}
					q, err := f.Covariance()
					if err != nil {
						t.Fatal(err)
					}
					if before != after || !reflect.DeepEqual(p, q) {
						t.Fatal("per-row rejection modified state or covariance")
					}
				}
			})
		}
	}
}

func TestFusionTightPerRowGateReports(t *testing.T) {
	sp3, epoch := fusionConstantSource(t)
	for _, kind := range []FusionFilterKind{FusionEKF, FusionUKF} {
		for _, innovation := range [][2]float64{{1, 2}, {8, 9}} {
			t.Run(fmt.Sprintf("%d/%v", kind, innovation), func(t *testing.T) {
				c := fusionGateConfig(t, kind, true, false)
				c.TightModel = &FusionTightModelOptions{InitialClockBiasVarianceM2: 2, InitialClockDriftVarianceM2PerS2: 3}
				f := fusionLiteralFilter(t, c, epoch, 0, [3]float64{0, 0, 4095968})
				before, err := f.State()
				if err != nil {
					t.Fatal(err)
				}
				p, err := f.Covariance()
				if err != nil {
					t.Fatal(err)
				}
				u, _, err := f.UpdateTightSP3(fusionClockEpoch(epoch, innovation), sp3, nil, FusionTightDirect)
				if err != nil {
					t.Fatal(err)
				}
				// H=I2, P=diag(2,3), R=diag(1,4); screen each innovation against 3*sqrt(S_ii).
				rejected := 0
				if math.Abs(innovation[0])/math.Sqrt(3) > 3 {
					rejected++
				}
				if math.Abs(innovation[1])/math.Sqrt(7) > 3 {
					rejected++
				}
				g := u.InnovationGate
				coasted := rejected == 2
				if g == nil || g.InputRows != 2 || g.AcceptedRows != 2-rejected || g.RejectedRows != rejected || g.Coasted != coasted || u.Applied == coasted || u.JointInnovationGate != nil {
					t.Fatalf("update=%+v gate=%+v", u, g)
				}
				if coasted {
					after, err := f.State()
					if err != nil {
						t.Fatal(err)
					}
					q, err := f.Covariance()
					if err != nil {
						t.Fatal(err)
					}
					if before != after || !reflect.DeepEqual(p, q) {
						t.Fatal("tight per-row rejection modified state or covariance")
					}
				}
			})
		}
	}
}
