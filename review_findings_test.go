package sidereon

import (
	"errors"
	"testing"
)

func newReviewFusionFilter(t *testing.T) *FusionFilter {
	t.Helper()
	config, err := DefaultFusionFilterConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.TimeSyncIMUCapacity = 8
	config.TimeSyncCheckpointCapacity = 4
	initial := FusionNavState{
		PositionECEFM:      [3]float64{6378137, 0, 0},
		AttitudeBodyToECEF: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1},
	}
	diagonal := make([]float64, 15)
	for i := range diagonal {
		diagonal[i] = 10
	}
	filter, err := NewFusionFilter(initial, diagonal, config)
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

func reviewFusionHistory(t *testing.T, mode FusionTightUpdateMode) *FusionRTSHistoryBuilder {
	t.Helper()
	if mode != FusionTightRecorded {
		return nil
	}
	history, err := NewFusionRTSHistoryBuilder()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = history.Close() })
	return history
}

func TestUpdateTightSP3PropagatesErrorsInEveryMode(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })

	modes := []FusionTightUpdateMode{FusionTightDirect, FusionTightRecorded, FusionTightTimeSync}
	for _, mode := range modes {
		t.Run(mode.String()+"/closed filter", func(t *testing.T) {
			filter := newReviewFusionFilter(t)
			if err := filter.Close(); err != nil {
				t.Fatal(err)
			}
			history := reviewFusionHistory(t, mode)
			_, _, err := filter.UpdateTightSP3(FusionTightEpoch{}, sp3, history, mode)
			if !errors.Is(err, ErrClosed) {
				t.Fatalf("UpdateTightSP3 after Close = %v, want ErrClosed", err)
			}
		})

		t.Run(mode.String()+"/embedded NUL", func(t *testing.T) {
			filter := newReviewFusionFilter(t)
			t.Cleanup(func() { _ = filter.Close() })
			history := reviewFusionHistory(t, mode)
			epoch := FusionTightEpoch{Observations: []FusionTightObservation{{SatelliteID: "G01\x00suffix"}}}
			_, _, err := filter.UpdateTightSP3(epoch, sp3, history, mode)
			if err == nil {
				t.Fatal("UpdateTightSP3 accepted an embedded-NUL satellite ID")
			}
		})
	}
}

func (m FusionTightUpdateMode) String() string {
	switch m {
	case FusionTightDirect:
		return "direct"
	case FusionTightRecorded:
		return "recorded"
	case FusionTightTimeSync:
		return "time sync"
	default:
		return "unknown"
	}
}

func TestNonRobustFDEPropagatesInputAndNativeErrors(t *testing.T) {
	sp3, err := LoadSP3(readPositioningFixture(t, "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	broadcast, err := ParseBroadcastEphemeris(readPositioningFixture(t, "nav/ESBC00DNK_R_20201770000_01D_MN.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = broadcast.Close() })
	options, err := DefaultFDEOptions()
	if err != nil {
		t.Fatal(err)
	}

	sources := []struct {
		name  string
		solve func(SPPConfig, FDEOptions) (*FDEResult, error)
	}{
		{"SP3", func(config SPPConfig, options FDEOptions) (*FDEResult, error) {
			return SolveFDE(sp3, config, options)
		}},
		{"broadcast", func(config SPPConfig, options FDEOptions) (*FDEResult, error) {
			return SolveFDEBroadcast(broadcast, config, options)
		}},
	}
	for _, source := range sources {
		t.Run(source.name+"/embedded NUL", func(t *testing.T) {
			config := usedSPPConfig()
			config.Observations[0].SatelliteID += "\x00suffix"
			result, err := source.solve(config, options)
			if result != nil {
				_ = result.Close()
				t.Fatalf("solve result = %#v, want nil after validation failure", result)
			}
			if err == nil {
				t.Fatal("solve accepted an embedded-NUL observation")
			}
		})

		t.Run(source.name+"/native validation", func(t *testing.T) {
			result, err := source.solve(SPPConfig{}, options)
			if result != nil {
				_ = result.Close()
				t.Fatalf("solve result = %#v, want nil after native validation failure", result)
			}
			var statusErr *StatusError
			if !errors.As(err, &statusErr) {
				t.Fatalf("solve error = %T %v, want *StatusError", err, err)
			}
			if statusErr.Code != StatusSolve || statusErr.Detail == "" {
				t.Fatalf("solve status = %+v, want detailed solve failure", statusErr)
			}
		})
	}
}

func TestNewFDEResultRejectsNilNativeHandle(t *testing.T) {
	result, err := newFDEResult(nil)
	if result != nil || !errors.Is(err, errNilNativeHandle) {
		t.Fatalf("newFDEResult(nil) = %#v, %v; want nil, errNilNativeHandle", result, err)
	}
}
