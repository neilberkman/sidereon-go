//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

const recordedLockProbe = 200 * time.Millisecond

func recordedCallResult(done <-chan error) (error, bool) {
	select {
	case err := <-done:
		return err, true
	case <-time.After(recordedLockProbe):
		return nil, false
	}
}

func awaitRecordedCall(t *testing.T, name string, done <-chan error) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("%s did not finish after the builder lock was released", name)
	}
}

func newNativeReviewTrackFilter(t *testing.T) *TrackFilter {
	t.Helper()
	filter, err := NewTrackFilterFromPosition(0, 0, []float64{0}, []float64{1}, 1, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filter.Close() })
	return filter
}

func TestPredictRecordedRequestsExclusiveBuilderLock(t *testing.T) {
	filter := newNativeReviewTrackFilter(t)
	history, err := NewTrackRTSHistoryBuilderFromFilter(filter)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = history.Close() })

	history.handle.mu.RLock()
	done := make(chan error, 1)
	go func() {
		_, err := filter.PredictRecorded(1, history)
		done <- err
	}()
	callErr, completed := recordedCallResult(done)
	history.handle.mu.RUnlock()
	if completed {
		t.Fatalf("PredictRecorded entered while a builder reader was active: %v", callErr)
	}
	awaitRecordedCall(t, "PredictRecorded", done)
}

func newNativeReviewFusionFilter(t *testing.T) *FusionFilter {
	t.Helper()
	config, err := FusionConfigDefault()
	if err != nil {
		t.Fatal(err)
	}
	config.TimeSyncIMUCapacity = 8
	config.TimeSyncCheckpointCapacity = 4
	diagonal := make([]float64, 15)
	for i := range diagonal {
		diagonal[i] = 10
	}
	filter, err := NewFusionFilter(NativeFusionNavState{
		Position: [3]float64{6378137, 0, 0},
		Attitude: [9]float64{1, 0, 0, 0, 1, 0, 0, 0, 1},
	}, diagonal, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = filter.Close() })
	return filter
}

func TestFusionTightRecordedRequestsExclusiveBuilderLock(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "trimmed.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp3.Close() })
	filter := newNativeReviewFusionFilter(t)
	history, err := NewFusionRTSHistoryBuilder()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = history.Close() })

	history.handle.mu.RLock()
	done := make(chan error, 1)
	go func() {
		_, _, err := filter.UpdateTightSP3(NativeFusionTightEpoch{}, sp3, history, 1)
		done <- err
	}()
	callErr, completed := recordedCallResult(done)
	history.handle.mu.RUnlock()
	if completed {
		t.Fatalf("UpdateTightSP3 recorded entered while a builder reader was active: %v", callErr)
	}
	awaitRecordedCall(t, "UpdateTightSP3 recorded", done)
}

func TestRecordedMutationExcludesFinishAndSecondMutation(t *testing.T) {
	first := newNativeReviewTrackFilter(t)
	second := newNativeReviewTrackFilter(t)
	history, err := NewTrackRTSHistoryBuilderFromFilter(first)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = history.Close() })

	// The exclusive lock represents the native callback of the first recorded
	// mutation. The tests above verify that both recorded entry points request
	// this lock rather than a shared reader lock.
	history.handle.mu.Lock()
	finishDone := make(chan error, 1)
	go func() {
		finished, err := history.Finish()
		if finished != nil {
			_ = finished.Close()
		}
		finishDone <- err
	}()
	mutationDone := make(chan error, 1)
	go func() {
		_, err := second.PredictRecorded(1, history)
		mutationDone <- err
	}()
	finishErr, finishCompleted := recordedCallResult(finishDone)
	mutationErr, mutationCompleted := recordedCallResult(mutationDone)
	history.handle.mu.Unlock()
	if finishCompleted {
		t.Fatalf("Finish entered while a recorded mutation was active: %v", finishErr)
	}
	if mutationCompleted {
		t.Fatalf("second recorded mutation entered while the first was active: %v", mutationErr)
	}
	awaitRecordedCall(t, "Finish", finishDone)
	awaitRecordedCall(t, "second recorded mutation", mutationDone)
}
