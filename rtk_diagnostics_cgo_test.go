//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
)

type holdFixtureEpoch struct {
	Rover        [3]float64   `json:"rover"`
	Positions    [][3]float64 `json:"positions"`
	Reference    string       `json:"reference"`
	Observations []struct {
		ID         string  `json:"id"`
		BaseCode   float64 `json:"base_code"`
		RoverCode  float64 `json:"rover_code"`
		BasePhase  float64 `json:"base_phase"`
		RoverPhase float64 `json:"rover_phase"`
	} `json:"observations"`
}
type holdFixtureCase struct {
	Offset      float64            `json:"offset_m"`
	Wavelength  float64            `json:"wavelength_m"`
	Wavelengths map[string]float64 `json:"wavelengths_m"`
	Epochs      []holdFixtureEpoch `json:"epochs"`
}

func TestRTKMaskReferencesAndReleasedHolds(t *testing.T) {
	data, err := os.ReadFile("testdata/rtk/rtk_hold_transport.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Base      [3]float64        `json:"base_m"`
		Mask      float64           `json:"mask_deg"`
		Scenarios []holdFixtureCase `json:"scenarios"`
		FDMA      holdFixtureCase   `json:"fdma_scenario"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, fdma := range []bool{false, true} {
		c := fixture.Scenarios[0]
		reason := RTKHoldReleaseUnheldReference
		if fdma {
			c = fixture.FDMA
			reason = RTKHoldReleaseIncompatibleWavelengths
		}
		model, err := DefaultRTKMeasurementModel()
		if err != nil {
			t.Fatal(err)
		}
		model.Sagnac = false
		update, err := DefaultRTKArcUpdateOptions()
		if err != nil {
			t.Fatal(err)
		}
		update.HoldSigmaM = 1e-4
		update.DynamicsVelocityPropagated = true
		config := RTKArcConfig{BaseM: fixture.Base, ReferenceMode: RTKArcReferencePerSystem, ReferencePerSystem: []RTKArcReferenceEntry{{System: GNSSSystemGPS, SatelliteID: "G01"}}, Model: model, BaselinePriorSigmaM: 100, AmbiguityPriorSigmaM: 1000, UpdateOptions: update, Preprocessing: RTKArcPreprocessing{HasElevationMaskDeg: true, ElevationMaskDeg: fixture.Mask}}
		for a := 0; a < 3; a++ {
			config.InitialBaselineM[a] = c.Epochs[0].Rover[a] - fixture.Base[a]
		}
		for i := 0; i < 6; i++ {
			id := fmt.Sprintf("G%02d", i+1)
			if i > 0 {
				config.Offsets = append(config.Offsets, RTKFloatMapEntry{ID: id, Value: c.Offset})
			}
			if fdma || i > 0 {
				wave := c.Wavelength
				if fdma {
					wave = c.Wavelengths[id]
				}
				config.Wavelengths = append(config.Wavelengths, RTKFloatMapEntry{ID: id, Value: wave})
			}
		}
		epochs := make([]RTKArcEpoch, 2)
		for e := range epochs {
			epoch := &epochs[e]
			epoch.HasVelocityMPS = true
			epoch.HasPredictionTime = true
			epoch.PredictionTimeS = float64(e * 30)
			if e > 0 {
				for a := 0; a < 3; a++ {
					epoch.VelocityMPS[a] = (c.Epochs[e].Rover[a] - c.Epochs[e-1].Rover[a]) / 30
				}
			}
			for i, row := range c.Epochs[e].Observations {
				if !fdma && e == 0 && row.ID == "G04" {
					continue
				}
				epoch.Base = append(epoch.Base, RTKArcObservation{SatelliteID: row.ID, AmbiguityID: row.ID, CodeM: row.BaseCode, PhaseM: row.BasePhase})
				epoch.Rover = append(epoch.Rover, RTKArcObservation{SatelliteID: row.ID, AmbiguityID: row.ID, CodeM: row.RoverCode, PhaseM: row.RoverPhase})
				epoch.SatellitePositions = append(epoch.SatellitePositions, RTKArcPositionEntry{SatelliteID: row.ID, PositionM: c.Epochs[e].Positions[i]})
			}
		}
		solution, err := SolveRTKArc(epochs, config)
		if err != nil {
			t.Fatal(err)
		}
		meta, err := solution.EpochMetadata(1)
		if err != nil {
			t.Fatal(err)
		}
		refs, err := solution.EpochReferenceSatellites(1)
		if err != nil {
			t.Fatal(err)
		}
		masked, err := solution.EpochElevationMaskedSatellites(1)
		if err != nil {
			t.Fatal(err)
		}
		released, err := solution.EpochReleasedHolds(1)
		if err != nil {
			t.Fatal(err)
		}
		_ = solution.Close()
		// Core hold-transport tests and RTKLIB mask fixture: G01 is masked; G04 takes over.
		if !meta.ElevationMaskConverged || !reflect.DeepEqual(refs, []string{"G04"}) || !reflect.DeepEqual(masked, []string{"G01"}) || len(released) < 2 || meta.ReleasedHoldCount < 2 || meta.ReferenceSatelliteCount != 1 || meta.ElevationMaskedSatelliteCount != 1 {
			t.Fatalf("fdma=%v meta=%+v refs=%v masked=%v released=%v", fdma, meta, refs, masked, released)
		}
		for _, hold := range released {
			if hold.Reason != reason || hold.AmbiguityID == "" || hold.AmbiguityID == "G01" || (!fdma && hold.AmbiguityID == "G04") {
				t.Fatalf("hold=%+v want reason=%v", hold, reason)
			}
		}
	}
}

func TestRTKStaticMaskConvergence(t *testing.T) {
	epochs, arc := syntheticRTKArcInput(t)
	float, err := DefaultRTKFloatOptions()
	if err != nil {
		t.Fatal(err)
	}
	fixed, err := DefaultRTKFixedOptions()
	if err != nil {
		t.Fatal(err)
	}
	residual, err := DefaultRTKResidualValidationOptions()
	if err != nil {
		t.Fatal(err)
	}
	solution, err := SolveStaticRTKArc(epochs, RTKStaticArcConfig{Arc: arc, FloatOptions: float, FixedOptions: fixed, ResidualOptions: residual})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = solution.Close() }()
	if converged, err := solution.ElevationMaskConverged(); err != nil || !converged {
		t.Fatalf("static convergence=%v / %v", converged, err)
	}
}

func TestCombinedMultiConstellationRTK(t *testing.T) {
	read := func(name string) []byte {
		data, err := os.ReadFile("testdata/rtk/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	sp3, err := LoadSP3(read("GBM0MGXRAP_20201770000_01D_05M_ORB_120epoch.sp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sp3.Close() }()
	base, err := ParseRINEXObservation(read("WTZR00DEU_R_20201770000_01D_30S_MO_120epoch.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = base.Close() }()
	rover, err := ParseRINEXObservation(read("WTZZ00DEU_R_20201770000_01D_30S_MO_120epoch.rnx"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rover.Close() }()
	// IGS marker coordinates and RINEX antenna heights from the core's real-arc
	// fixture. Compute ARP along the WGS84 geodetic normal, independently of C.
	arp := func(marker [3]float64, height float64) [3]float64 {
		const a = 6378137.0
		const e2 = 6.6943799901413165e-3
		rho := math.Hypot(marker[0], marker[1])
		lat := math.Atan2(marker[2], rho*(1-e2))
		lon := math.Atan2(marker[1], marker[0])
		for i := 0; i < 10; i++ {
			n := a / math.Sqrt(1-e2*math.Sin(lat)*math.Sin(lat))
			lat = math.Atan2(marker[2]+e2*n*math.Sin(lat), rho)
		}
		return [3]float64{marker[0] + height*math.Cos(lat)*math.Cos(lon), marker[1] + height*math.Cos(lat)*math.Sin(lon), marker[2] + height*math.Sin(lat)}
	}
	baseARP := arp([3]float64{4075580.3111, 931854.0543, 4801568.2808}, 0.071)
	roverARP := arp([3]float64{4075579.1913, 931853.3696, 4801569.1897}, 0.284)
	config, err := DefaultRTKRINEXWideLaneFixedConfig()
	if err != nil {
		t.Fatal(err)
	}
	config.BaseM = baseARP
	config.ArcOptions.HasMaxEpochs = true
	config.ArcOptions.MaxEpochs = 120
	config.ArcOptions.IncludePredictionTime = false
	config.Model.CodeSigmaM = 2
	config.Model.PhaseSigmaM = 0.01
	config.Model.Sagnac = true
	config.Model.Stochastic = 0
	config.Model.ElevationWeighting = true
	config.ArcOptions.SignalPairs = []RTKRINEXDualSignalPair{
		{System: GNSSSystemGPS, Code1Observable: "C1C", Phase1Observable: "L1C", Code2Observable: "C2W", Phase2Observable: "L2W"},
		{System: GNSSSystemBeiDou, Code1Observable: "C2I", Phase1Observable: "L2I", Code2Observable: "C6I", Phase2Observable: "L6I"},
	}
	solution, err := SolveWideLaneFixedRINEXRTKBaseline(sp3, base, rover, config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = solution.Close() }()
	cycles, err := solution.WideLaneCycles()
	if err != nil {
		t.Fatal(err)
	}
	gps, bds := false, false
	for _, row := range cycles {
		if len(row.ID) > 0 {
			gps = gps || row.ID[0] == 'G'
			bds = bds || row.ID[0] == 'C'
		}
	}
	if !gps || !bds {
		t.Fatalf("missing constellation cycles=%v", cycles)
	}
	baseline, err := solution.FloatBaselineECEF()
	if err != nil {
		t.Fatal(err)
	}
	squared := 0.0
	for i := 0; i < 3; i++ {
		delta := baseline[i] - (roverARP[i] - baseARP[i])
		squared += delta * delta
	}
	if math.Sqrt(squared) >= 0.1 {
		t.Fatalf("baseline=%v IGS difference=%g", baseline, math.Sqrt(squared))
	}
}
