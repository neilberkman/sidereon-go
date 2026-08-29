//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import (
	"sync"
	"testing"
	"time"
	"unsafe"
)

func testPositioningHandle(t *testing.T) *positioningHandle {
	t.Helper()
	value := new(int)
	handle := newPositioningHandle(unsafe.Pointer(value), func(unsafe.Pointer) {})
	t.Cleanup(func() { _ = handle.close() })
	return handle
}

func TestWithPositioningHandlesPreservesOrderAndDeduplicates(t *testing.T) {
	first := testPositioningHandle(t)
	second := testPositioningHandle(t)
	var got []unsafe.Pointer
	if err := withPositioningHandles([]*positioningHandle{second, first, second}, func(pointers []unsafe.Pointer) error {
		got = append([]unsafe.Pointer(nil), pointers...)
		return nil
	}); err != nil {
		t.Fatalf("withPositioningHandles: %v", err)
	}
	if len(got) != 3 || got[0] == nil || got[1] == nil || got[2] != got[0] {
		t.Fatalf("pointers = %#v, want original order with duplicate pointer", got)
	}
}

func TestWithPositioningHandlesInverseOrderAndDuplicateNoDeadlock(t *testing.T) {
	first := testPositioningHandle(t)
	second := testPositioningHandle(t)
	start := make(chan struct{})
	done := make(chan error, 2)
	var group sync.WaitGroup
	group.Add(2)
	go func() {
		defer group.Done()
		<-start
		done <- withPositioningHandles([]*positioningHandle{first, second, first}, func([]unsafe.Pointer) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}()
	go func() {
		defer group.Done()
		<-start
		done <- withPositioningHandles([]*positioningHandle{second, first, second}, func([]unsafe.Pointer) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}()
	close(start)
	finished := make(chan struct{})
	go func() {
		group.Wait()
		close(finished)
	}()
	select {
	case <-finished:
		for range 2 {
			if err := <-done; err != nil {
				t.Fatalf("withPositioningHandles inverse list: %v", err)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("withPositioningHandles inverse lists deadlocked")
	}
}

func TestAstroOutputEnumValidation(t *testing.T) {
	if err := validateGNSSSystemValue(7); err == nil {
		t.Fatal("invalid GNSS system accepted")
	}
	if err := validateNavcenTiming(NavcenTimingUnparseableValue + 1); err == nil {
		t.Fatal("invalid NAVCEN timing accepted")
	}
	if err := validTimeScale(11); err == nil {
		t.Fatal("invalid IONEX time scale accepted")
	}
	if err := validateGeodeticQuality(2); err == nil {
		t.Fatal("invalid geodetic quality accepted")
	}
	if err := validateGeodeticStepHeuristic(1); err == nil {
		t.Fatal("invalid geodetic step heuristic accepted")
	}
	if err := validateGeodeticTrajectoryTerm(7); err == nil {
		t.Fatal("invalid geodetic trajectory term accepted")
	}
	if err := validateObservabilityTier(4); err == nil {
		t.Fatal("invalid observability tier accepted")
	}
}
