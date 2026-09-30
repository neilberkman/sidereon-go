//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func ssrIngestFailureForSatellite(satelliteID uint8) (error, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	message := func(id uint8) (*RTCMMessages, error) {
		return BuildRTCMSSRMessageV2(
			RTCMSSRInfoV2{
				MessageNumber: 1253,
				System:        GNSSSystemSBAS,
				Kind:          RTCMSSRClock,
				Header: RTCMSSRHeader{
					EpochTimeS: 90, UpdateInterval: 2, IODSSR: 3,
					ProviderID: 4, SolutionID: 1, SatelliteCount: 1,
				},
			}, nil,
			[]RTCMSSRClockRecord{{SatelliteID: id, C0: 1}},
			nil, nil, nil,
		)
	}
	invalidMessages, err := message(satelliteID)
	if err != nil {
		return nil, err
	}
	defer invalidMessages.Close()
	validMessages, err := message(1)
	if err != nil {
		return nil, err
	}
	defer validMessages.Close()
	store, err := NewSSRCorrectionStore(SSRReferencePointAntennaPhaseCenter)
	if err != nil {
		return nil, err
	}
	defer store.Close()

	ingestErr := store.Ingest(invalidMessages, GNSSWeekTow{System: GPST, Week: 2425, TOWSeconds: 345000})
	if ingestErr == nil {
		return nil, fmt.Errorf("SSR ingest unexpectedly accepted invalid SBAS satellite id %d", satelliteID)
	}
	encoded, encodeErr := validMessages.Encode(0)
	if encodeErr != nil || len(encoded) == 0 {
		return nil, fmt.Errorf("successful RTCM encode after ingest error: bytes=%d error=%v", len(encoded), encodeErr)
	}
	return ingestErr, nil
}

func TestSSRIngestCapturesOwnedStructuredErrors(t *testing.T) {
	type result struct {
		satelliteID uint8
		err         error
		setupErr    error
	}
	results := make(chan result, 2)
	var workers sync.WaitGroup
	for _, satelliteID := range []uint8{40, 41} {
		workers.Add(1)
		go func(satelliteID uint8) {
			defer workers.Done()
			err, setupErr := ssrIngestFailureForSatellite(satelliteID)
			results <- result{satelliteID: satelliteID, err: err, setupErr: setupErr}
		}(satelliteID)
	}
	workers.Wait()
	close(results)

	for result := range results {
		if result.setupErr != nil {
			t.Fatalf("SSR ingest for SBAS satellite id %d: %v", result.satelliteID, result.setupErr)
		}
		var structured *RTCMError
		if !errors.As(result.err, &structured) {
			t.Fatalf("SSR ingest for SBAS satellite id %d lost structured RTCM error: %v", result.satelliteID, result.err)
		}
		if structured.Class != RTCMErrorClassOther || structured.Kind != RTCMErrorKindUnknown {
			t.Fatalf("unexpected RTCM error for SBAS satellite id %d: class=%d kind=%d", result.satelliteID, structured.Class, structured.Kind)
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(structured.Payload, &payload); err != nil {
			t.Fatalf("invalid RTCM error JSON for SBAS satellite id %d: %s (%v)", result.satelliteID, structured.Payload, err)
		}
		var schemaVersion int
		var detail map[string]json.RawMessage
		if len(payload) != 2 || json.Unmarshal(payload["schema_version"], &schemaVersion) != nil || json.Unmarshal(payload["error"], &detail) != nil || schemaVersion != 1 || len(detail) != 1 {
			t.Fatalf("unexpected RTCM error payload structure for SBAS satellite id %d: %s", result.satelliteID, structured.Payload)
		}
		var variant string
		if json.Unmarshal(detail["variant"], &variant) != nil || variant != "OtherError" {
			t.Fatalf("unexpected RTCM error payload for SBAS satellite id %d: %s", result.satelliteID, structured.Payload)
		}
		if strings.Count(result.err.Error(), fmt.Sprint(result.satelliteID)) == 0 {
			t.Fatalf("SSR ingest error did not identify invalid SBAS satellite id %d: %v", result.satelliteID, result.err)
		}
	}
}
