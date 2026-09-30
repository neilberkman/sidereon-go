package sidereon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
)

type fdeCandidateDiagnostic struct {
	ExcludedSatelliteID string       `json:"excluded_satellite_id"`
	Solution            *SPPSolution `json:"solution,omitempty"`
	Error               string       `json:"error,omitempty"`
	ResidualRMSM        float64      `json:"residual_rms_m,omitempty"`
	HasResidualRMS      bool         `json:"has_residual_rms"`
}

type fdeDiagnosticCapture struct {
	FixtureSHA256 string                   `json:"fixture_sha256"`
	Config        SPPConfig                `json:"config"`
	Options       FDEOptions               `json:"options"`
	FDEError      string                   `json:"fde_error"`
	CauseError    string                   `json:"cause_error,omitempty"`
	StatusCause   *StatusError             `json:"status_cause,omitempty"`
	Unresolved    *FDEUnresolvedError      `json:"unresolved,omitempty"`
	Success       *fdeSuccessDiagnostic    `json:"success,omitempty"`
	Candidates    []fdeCandidateDiagnostic `json:"leave_one_out_candidates"`
}

type fdeSuccessDiagnostic struct {
	Solution       SPPSolution            `json:"solution"`
	Diagnostics    FDEDiagnostics         `json:"diagnostics"`
	AcceptedRAIM   RAIMResult             `json:"accepted_raim"`
	NormalizedRows []RAIMNormalizedResidual `json:"normalized_rows"`
}

// TestCaptureDefaultFDECandidateDiagnostics is opt-in instrumentation for a
// host gate. It records the exact public unresolved payload and all one-row
// leave-outs using the fixture and options from TestDeterministicReducedAndFDE.
func TestCaptureDefaultFDECandidateDiagnostics(t *testing.T) {
	dir := os.Getenv("SIDEREON_FDE_DIAGNOSTIC_DIR")
	if dir == "" {
		t.Skip("set SIDEREON_FDE_DIAGNOSTIC_DIR to capture FDE diagnostics")
	}
	data := readPositioningFixture(t, "trimmed.sp3")
	fixtureHash := sha256.Sum256(data)
	sp3, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := sp3.Close(); err != nil {
			t.Error(err)
		}
	}()
	config := usedSPPConfig()
	options := FDEOptions{PFA: 1e-3, WeightsMode: FDEWeightsUnit}
	capture := fdeDiagnosticCapture{
		FixtureSHA256: hex.EncodeToString(fixtureHash[:]),
		Config:        config,
		Options:       options,
		Candidates:    make([]fdeCandidateDiagnostic, 0, len(config.Observations)),
	}
	result, fdeErr := SolveFDE(sp3, config, options)
	if fdeErr != nil {
		capture.FDEError = fdeErr.Error()
		if !errors.As(fdeErr, &capture.Unresolved) {
			t.Fatalf("FDE diagnostic expected typed unresolved error, got %T: %v", fdeErr, fdeErr)
		}
		cause := errors.Unwrap(fdeErr)
		if cause != nil {
			capture.CauseError = cause.Error()
			_ = errors.As(cause, &capture.StatusCause)
		}
	} else if result == nil {
		t.Fatal("FDE returned neither a result nor an error")
	} else {
		solution, solutionErr := result.Solution()
		if solutionErr != nil {
			t.Fatal(solutionErr)
		}
		diagnostics, diagnosticsErr := result.Diagnostics()
		if diagnosticsErr != nil {
			t.Fatal(diagnosticsErr)
		}
		acceptedRAIM, normalizedRows, acceptedErr := result.AcceptedRAIM()
		if acceptedErr != nil {
			t.Fatal(acceptedErr)
		}
		capture.Success = &fdeSuccessDiagnostic{
			Solution: solution, Diagnostics: diagnostics,
			AcceptedRAIM: acceptedRAIM, NormalizedRows: normalizedRows,
		}
		if err := result.Close(); err != nil {
			t.Fatal(err)
		}
	}
	for _, observation := range config.Observations {
		candidate := fdeCandidateDiagnostic{ExcludedSatelliteID: observation.SatelliteID}
		candidateConfig := config
		candidateConfig.Observations = make([]SPPObservation, 0, len(config.Observations)-1)
		for _, remaining := range config.Observations {
			if remaining.SatelliteID != observation.SatelliteID {
				candidateConfig.Observations = append(candidateConfig.Observations, remaining)
			}
		}
		solution, solveErr := SolveSPP(sp3, candidateConfig)
		if solveErr != nil {
			candidate.Error = solveErr.Error()
		} else {
			candidate.Solution = &solution
			if len(solution.ResidualsM) != 0 {
				var sumSquares float64
				finite := true
				for _, residual := range solution.ResidualsM {
					if math.IsNaN(residual) || math.IsInf(residual, 0) {
						finite = false
						break
					}
					sumSquares += residual * residual
				}
				if finite {
					candidate.ResidualRMSM = math.Sqrt(sumSquares / float64(len(solution.ResidualsM)))
					candidate.HasResidualRMS = true
				}
			}
		}
		capture.Candidates = append(capture.Candidates, candidate)
	}
	encoded, err := json.MarshalIndent(capture, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fde-default-candidates.json"), append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
