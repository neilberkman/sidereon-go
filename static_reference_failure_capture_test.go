package sidereon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type staticReferenceCapture struct {
	SP3InputSHA256 string          `json:"sp3_input_sha256"`
	ObsInputSHA256 string          `json:"obs_input_sha256"`
	RetimedEpochs  []float64       `json:"retimed_sp3_epochs_j2000_s"`
	ReferenceM     [3]float64      `json:"reference_position_m"`
	Config         any             `json:"config"`
	Success        bool            `json:"success"`
	Error          string          `json:"error,omitempty"`
	Status         *StatusError    `json:"status,omitempty"`
	Solution       any             `json:"solution,omitempty"`
}

// TestCaptureStaticReferenceModeFailure is host-gate instrumentation for the
// exact static reference-station fixture. It records the full typed engine
// payload so mode-level errors remain inspectable when no mode yields a result.
func TestCaptureStaticReferenceModeFailure(t *testing.T) {
	dir := os.Getenv("SIDEREON_STATIC_REFERENCE_CAPTURE_DIR")
	if dir == "" {
		t.Skip("set SIDEREON_STATIC_REFERENCE_CAPTURE_DIR to capture static reference result")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	sp3Input := readPositioningFixture(t, "trimmed.sp3")
	obsInput := readObservationFixture(t, "ESBC00DNK_R_20201770000_01D_30S_MO_trim.rnx")
	sp3, reference, rover := rinexRTKFixture(t)
	config, err := DefaultStaticReferenceStationRinexConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.ReferencePositionM = [3]float64{3582105.291, 532589.7313, 5232754.8054}
	epochs, err := sp3.Epochs()
	if err != nil {
		t.Fatal(err)
	}
	capture := staticReferenceCapture{
		SP3InputSHA256: sha256Text(sp3Input),
		ObsInputSHA256: sha256Text(obsInput),
		RetimedEpochs:  epochs,
		ReferenceM:     config.ReferencePositionM,
		Config:         config,
	}
	solution, solveErr := SolveStaticReferenceStationRINEX(sp3, reference, rover, config)
	if solveErr != nil {
		capture.Error = solveErr.Error()
		var status *StatusError
		if errors.As(solveErr, &status) {
			copyStatus := *status
			if status.Engine != nil {
				copyEngine := *status.Engine
				copyEngine.Fields = append([]byte(nil), status.Engine.Fields...)
				copyEngine.Payload = append([]byte(nil), status.Engine.Payload...)
				copyStatus.Engine = &copyEngine
			}
			capture.Status = &copyStatus
		}
	} else {
		capture.Success = true
		defer solution.Close()
		metadata, metadataErr := solution.Metadata()
		if metadataErr != nil {
			t.Fatal(metadataErr)
		}
		capture.Solution = metadata
	}
	encoded, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "static-reference-mode-failure.json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sha256Text(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
