//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"testing"
)

func TestADDV30AndUnsupportedConstellation(t *testing.T) {
	data, err := os.ReadFile("testdata/araim/araim_oracles.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Sigma []struct {
			Elevation float64 `json:"elevation_deg"`
			Variance  float64 `json:"v3_0_variance_m2"`
		} `json:"sigma_model"`
		Galileo []struct {
			Elevation float64 `json:"elevation_deg"`
			Sigma     float64 `json:"user_sigma_m"`
		} `json:"galileo_user_sigma_model"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, row := range fixture.Sigma {
		model, err := ARAIMSatelliteModelADDV30(uint32(GNSSSystemGPS), row.Elevation*math.Pi/180, 0.75, 0.5, 0.5, 1e-5)
		if err != nil {
			t.Fatal(err)
		}
		if !model.HasEffectiveIntegrity || !model.HasEffectiveAccuracy || math.Abs(model.EffectiveIntegrityM*model.EffectiveIntegrityM-(0.75*0.75+row.Variance)) > 3e-15 || math.Abs(model.EffectiveAccuracyM*model.EffectiveAccuracyM-(0.5*0.5+row.Variance)) > 3e-15 {
			t.Fatalf("GPS ADD model = %+v", model)
		}
	}
	for _, row := range fixture.Galileo {
		el := row.Elevation * math.Pi / 180
		model, err := ARAIMSatelliteModelADDV30(uint32(GNSSSystemGalileo), el, 0, 0, 0, 1e-5)
		if err != nil {
			t.Fatal(err)
		}
		tropo := 0.12 * 1.001 / math.Sqrt(0.002001+math.Sin(el)*math.Sin(el))
		want := math.Sqrt(row.Sigma*row.Sigma + tropo*tropo)
		if math.Abs(model.EffectiveIntegrityM-want) > 2e-15 {
			t.Fatalf("Galileo sigma=%g want=%g", model.EffectiveIntegrityM, want)
		}
	}
	_, err = ARAIMSatelliteModelADDV30(uint32(GNSSSystemBeiDou), math.Pi/6, 0.75, 0.5, 0.5, 1e-5)
	var engine *EngineError
	if !errors.As(err, &engine) || engine.Kind != "unsupported_constellation" || engine.TypedFields["system"].Number != "3" {
		t.Fatalf("unsupported system = %+v / %v", engine, err)
	}
}
