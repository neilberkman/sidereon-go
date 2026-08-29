//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type OrbitFitOptions struct {
	ForceModel                                       uint32
	ForceComponents                                  ForceModelComponents
	MuKm3PerS2Enabled                                bool
	MuKm3PerS2                                       float64
	Integrator                                       uint32
	AbsTol, RelTol, InitialStepS, MinStepS, MaxStepS float64
	MaxSteps                                         uint32
	SolverGTol, SolverFTol, SolverXTol               float64
	SolverMaxNFEV, MinLedgerSamples                  int
	HasDrag                                          bool
	Drag                                             DragParameters
}

type PreciseEphemerisSample struct {
	Satellite     string
	TimeScale     uint32
	EpochJ2000S   float64
	PositionECEFM [3]float64
	HasClock      bool
	ClockS        float64
	ClockEvent    bool
}
type OrbitFitCovariance struct {
	Kind   uint32
	Matrix [36]float64
}
type GeometryQuality struct {
	Tier                               uint32
	Redundancy                         int32
	Rank                               int
	ConditionNumber, GDOP              float64
	RAIMCheckable, CovarianceValidated bool
}
type OrbitFitSolution struct {
	Satellite             string
	InitialState          CartesianState
	Covariance            OrbitFitCovariance
	GeometryQuality       GeometryQuality
	SeedRMS3DM, FitRMS3DM float64
	Iterations            int
}
type OrbitResidualStats struct {
	RadialRMSM, AlongRMSM, CrossRMSM, RMS3DM float64
	N                                        int
	LowSampleCount                           bool
}
type OrbitSatelliteResidualEntry struct {
	Satellite string
	Stats     OrbitResidualStats
}
type OrbitConstellationResidualEntry struct {
	System uint32
	Stats  OrbitResidualStats
}
type OrbitArcSpan struct {
	TimeScale                         uint32
	StartJ2000S, EndJ2000S, DurationS float64
}
type OrbitFitReport struct {
	_      noCopy
	handle *positioningHandle
}

func cOrbitFitOptions(value OrbitFitOptions) (C.SidereonOrbitFitOptions, error) {
	force, err := cPropagationForceModel(value.ForceModel)
	if err != nil {
		return C.SidereonOrbitFitOptions{}, err
	}
	integrator, err := cPropagationIntegrator(value.Integrator)
	if err != nil {
		return C.SidereonOrbitFitOptions{}, err
	}
	maxNFEV, err := checkedNativeSize(value.SolverMaxNFEV)
	if err != nil {
		return C.SidereonOrbitFitOptions{}, err
	}
	minLedger, err := checkedNativeSize(value.MinLedgerSamples)
	if err != nil {
		return C.SidereonOrbitFitOptions{}, err
	}
	return C.SidereonOrbitFitOptions{force_model: force, force_components: cForceModelComponents(value.ForceComponents), mu_km3_s2_enabled: C.bool(value.MuKm3PerS2Enabled), mu_km3_s2: C.double(value.MuKm3PerS2), integrator: integrator, abs_tol: C.double(value.AbsTol), rel_tol: C.double(value.RelTol), initial_step_s: C.double(value.InitialStepS), min_step_s: C.double(value.MinStepS), max_step_s: C.double(value.MaxStepS), max_steps: C.uint32_t(value.MaxSteps), solver_gtol: C.double(value.SolverGTol), solver_ftol: C.double(value.SolverFTol), solver_xtol: C.double(value.SolverXTol), solver_max_nfev: maxNFEV, min_ledger_samples: minLedger, has_drag: C.bool(value.HasDrag), drag: cDragParameters(value.Drag)}, nil
}

func validTimeScale(value uint32) error {
	if value > 10 {
		return invalidArgument("time scale is not defined by the C ABI")
	}
	return nil
}
func orbitFitOptionsFromC(value C.SidereonOrbitFitOptions) (OrbitFitOptions, error) {
	maxNFEV, err := checkedNativeCount(uint64(value.solver_max_nfev))
	if err != nil {
		return OrbitFitOptions{}, err
	}
	minLedgerSamples, err := checkedNativeCount(uint64(value.min_ledger_samples))
	if err != nil {
		return OrbitFitOptions{}, err
	}
	return OrbitFitOptions{ForceModel: uint32(value.force_model), ForceComponents: forceModelComponentsFromC(value.force_components), MuKm3PerS2Enabled: bool(value.mu_km3_s2_enabled), MuKm3PerS2: float64(value.mu_km3_s2), Integrator: uint32(value.integrator), AbsTol: float64(value.abs_tol), RelTol: float64(value.rel_tol), InitialStepS: float64(value.initial_step_s), MinStepS: float64(value.min_step_s), MaxStepS: float64(value.max_step_s), MaxSteps: uint32(value.max_steps), SolverGTol: float64(value.solver_gtol), SolverFTol: float64(value.solver_ftol), SolverXTol: float64(value.solver_xtol), SolverMaxNFEV: maxNFEV, MinLedgerSamples: minLedgerSamples, HasDrag: bool(value.has_drag), Drag: dragParametersFromC(value.drag)}, nil
}
func forceModelComponentsFromC(value C.SidereonForceModelComponents) ForceModelComponents {
	return ForceModelComponents{HasTwoBody: bool(value.has_two_body), TwoBodyMuKm3PerS2Enabled: bool(value.two_body_mu_km3_s2_enabled), TwoBodyMuKm3PerS2: float64(value.two_body_mu_km3_s2), HasZonal: bool(value.has_zonal), ZonalMaxDegree: uint32(value.zonal_max_degree), HasSphericalHarmonic: bool(value.has_spherical_harmonic), SphericalHarmonicMaxDegree: uint32(value.spherical_harmonic_max_degree), SphericalHarmonicMaxOrder: uint32(value.spherical_harmonic_max_order), HasSolidEarthTide: bool(value.has_solid_earth_tide), HasSolidEarthPoleTide: bool(value.has_solid_earth_pole_tide), HasThirdBody: bool(value.has_third_body), ThirdBodySun: bool(value.third_body_sun), ThirdBodyMoon: bool(value.third_body_moon), HasSolarRadiationPressure: bool(value.has_solar_radiation_pressure), SolarRadiationPressure: SolarRadiationPressure{CR: float64(value.solar_radiation_pressure.cr), AreaToMassM2PerKg: float64(value.solar_radiation_pressure.area_to_mass_m2_kg)}, HasRelativity: bool(value.has_relativity)}
}

func releaseOrbitFitReport(pointer unsafe.Pointer) {
	C.sidereon_orbit_fit_report_free((*C.SidereonOrbitFitReport)(pointer))
}

func fitReportWithSP3(sp3 *SP3, satID string, options *OrbitFitOptions, mode uint32) (*OrbitFitReport, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	if err := rejectEmbeddedNUL(satID, "orbit-fit satellite ID"); err != nil {
		return nil, err
	}
	if len(satID) == 0 {
		return nil, invalidArgument("orbit-fit satellite ID must not be empty")
	}
	var cOptions C.SidereonOrbitFitOptions
	var optionsPtr *C.SidereonOrbitFitOptions
	var err error
	if options != nil {
		cOptions, err = cOrbitFitOptions(*options)
		if err != nil {
			return nil, err
		}
		optionsPtr = &cOptions
	}
	var out *C.SidereonOrbitFitReport
	err = sp3.handle.with(func(pointer unsafe.Pointer) error {
		return withString(satID, func(value *C.char) uint32 {
			switch mode {
			case 0:
				return C.sidereon_fit_sp3_precise_orbit((*C.SidereonSp3)(pointer), value, optionsPtr, &out)
			default:
				return C.sidereon_fit_sp3_ecef_precise_orbit((*C.SidereonSp3)(pointer), value, optionsPtr, &out)
			}
		})
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_orbit_fit_report_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("orbit-fit report")
	}
	return &OrbitFitReport{handle: newPositioningHandle(unsafe.Pointer(out), releaseOrbitFitReport)}, nil
}
func FitSP3PreciseOrbit(sp3 *SP3, satID string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	return fitReportWithSP3(sp3, satID, options, 0)
}
func FitSP3ECEFPreciseOrbit(sp3 *SP3, satID string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	return fitReportWithSP3(sp3, satID, options, 1)
}

func FitSP3ECEFPreciseOrbits(sp3 *SP3, satellites []string, options *OrbitFitOptions, all bool) (*OrbitFitReport, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	satCopy := append([]string(nil), satellites...)
	if all && len(satCopy) != 0 {
		return nil, invalidArgument("all-satellites fit does not accept a satellite list")
	}
	if !all && len(satCopy) == 0 {
		return nil, invalidArgument("selected-satellite fit requires at least one satellite")
	}
	if _, err := checkedNativeAllocationSize(len(satCopy), unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
		return nil, err
	}
	tokens := make([]C.SidereonSatelliteToken, len(satCopy))
	for i, value := range satCopy {
		if value == "" {
			return nil, invalidArgument("orbit-fit satellite ID must not be empty")
		}
		var err error
		tokens[i], err = tokenToC(value)
		if err != nil {
			return nil, err
		}
		if err := rejectEmbeddedNUL(value, "orbit-fit satellite ID"); err != nil {
			return nil, err
		}
	}
	var cOptions C.SidereonOrbitFitOptions
	var optionsPtr *C.SidereonOrbitFitOptions
	var err error
	if options != nil {
		cOptions, err = cOrbitFitOptions(*options)
		if err != nil {
			return nil, err
		}
		optionsPtr = &cOptions
	}
	var out *C.SidereonOrbitFitReport
	var tokenPtr *C.SidereonSatelliteToken
	if len(tokens) > 0 {
		tokenPtr = &tokens[0]
	}
	err = sp3.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			if all {
				return C.sidereon_fit_all_sp3_ecef_precise_orbits((*C.SidereonSp3)(pointer), optionsPtr, &out)
			}
			return C.sidereon_fit_sp3_ecef_precise_orbits((*C.SidereonSp3)(pointer), tokenPtr, C.size_t(len(tokens)), optionsPtr, &out)
		})
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_orbit_fit_report_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("orbit-fit report")
	}
	return &OrbitFitReport{handle: newPositioningHandle(unsafe.Pointer(out), releaseOrbitFitReport)}, nil
}

func cPreciseSample(value PreciseEphemerisSample) (C.SidereonPreciseEphemerisSample, error) {
	if value.Satellite == "" {
		return C.SidereonPreciseEphemerisSample{}, invalidArgument("precise-ephemeris satellite ID must not be empty")
	}
	sat, err := tokenToC(value.Satellite)
	if err != nil {
		return C.SidereonPreciseEphemerisSample{}, err
	}
	if err := validTimeScale(value.TimeScale); err != nil {
		return C.SidereonPreciseEphemerisSample{}, err
	}
	var out C.SidereonPreciseEphemerisSample
	out.sat, out.time_scale, out.epoch_j2000_s, out.has_clock_s, out.clock_s, out.clock_event = sat, C.uint32_t(value.TimeScale), C.double(value.EpochJ2000S), C.bool(value.HasClock), C.double(value.ClockS), C.bool(value.ClockEvent)
	for i := range out.position_ecef_m {
		out.position_ecef_m[i] = C.double(value.PositionECEFM[i])
	}
	return out, nil
}
func FitPreciseEphemerisSamples(samples []PreciseEphemerisSample, satID string, options *OrbitFitOptions) (*OrbitFitReport, error) {
	if err := rejectEmbeddedNUL(satID, "orbit-fit satellite ID"); err != nil {
		return nil, err
	}
	if satID == "" {
		return nil, invalidArgument("orbit-fit satellite ID must not be empty")
	}
	if len(samples) == 0 {
		return nil, invalidArgument("precise-ephemeris samples must not be empty")
	}
	if _, err := checkedNativeAllocationSize(len(samples), unsafe.Sizeof(C.SidereonPreciseEphemerisSample{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonPreciseEphemerisSample, len(samples))
	for i, sample := range samples {
		var err error
		values[i], err = cPreciseSample(sample)
		if err != nil {
			return nil, err
		}
	}
	var cOptions C.SidereonOrbitFitOptions
	var optionsPtr *C.SidereonOrbitFitOptions
	var err error
	if options != nil {
		cOptions, err = cOrbitFitOptions(*options)
		if err != nil {
			return nil, err
		}
		optionsPtr = &cOptions
	}
	var out *C.SidereonOrbitFitReport
	err = withString(satID, func(value *C.char) uint32 {
		return C.sidereon_fit_precise_ephemeris_sample_orbit(&values[0], C.size_t(len(values)), value, optionsPtr, &out)
	})
	runtime.KeepAlive(values)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_orbit_fit_report_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("orbit-fit report")
	}
	return &OrbitFitReport{handle: newPositioningHandle(unsafe.Pointer(out), releaseOrbitFitReport)}, nil
}

func (r *OrbitFitReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return r.handle.close()
}
func orbitStatsFromC(value C.SidereonOrbitResidualStats) (OrbitResidualStats, error) {
	n, err := checkedNativeCount(uint64(value.n))
	if err != nil {
		return OrbitResidualStats{}, err
	}
	return OrbitResidualStats{RadialRMSM: float64(value.radial_rms_m), AlongRMSM: float64(value.along_rms_m), CrossRMSM: float64(value.cross_rms_m), RMS3DM: float64(value.rms_3d_m), N: n, LowSampleCount: bool(value.low_sample_count)}, nil
}
func geometryFromC(value C.SidereonGeometryQuality) (GeometryQuality, error) {
	rank, err := checkedNativeCount(uint64(value.rank))
	if err != nil {
		return GeometryQuality{}, err
	}
	return GeometryQuality{Tier: uint32(value.tier), Redundancy: int32(value.redundancy), Rank: rank, ConditionNumber: float64(value.condition_number), GDOP: float64(value.gdop), RAIMCheckable: bool(value.raim_checkable), CovarianceValidated: bool(value.covariance_validated)}, nil
}
func solutionFromC(value C.SidereonOrbitFitSolution) (OrbitFitSolution, error) {
	covariance := OrbitFitCovariance{Kind: uint32(value.covariance.kind)}
	for i := range covariance.Matrix {
		covariance.Matrix[i] = float64(value.covariance.matrix[i])
	}
	iterations, err := checkedNativeCount(uint64(value.iterations))
	if err != nil {
		return OrbitFitSolution{}, err
	}
	geometry, err := geometryFromC(value.geometry_quality)
	if err != nil {
		return OrbitFitSolution{}, err
	}
	return OrbitFitSolution{Satellite: tokenFromC(value.satellite), InitialState: cartesianStateFromC(value.initial_state), Covariance: covariance, GeometryQuality: geometry, SeedRMS3DM: float64(value.seed_rms_3d_m), FitRMS3DM: float64(value.fit_rms_3d_m), Iterations: iterations}, nil
}

func reportValues(r *OrbitFitReport, label string, kind uint32) (interface{}, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	var result interface{}
	err := r.handle.with(func(pointer unsafe.Pointer) error {
		var written, required C.size_t
		invoke := func(out unsafe.Pointer, n C.size_t, w, q *C.size_t) uint32 {
			switch kind {
			case 0:
				return C.sidereon_orbit_fit_report_fits((*C.SidereonOrbitFitReport)(pointer), (*C.SidereonOrbitFitSolution)(out), n, w, q)
			case 1:
				return C.sidereon_orbit_fit_report_satellite_ledger((*C.SidereonOrbitFitReport)(pointer), (*C.SidereonOrbitSatelliteResidualEntry)(out), n, w, q)
			default:
				return C.sidereon_orbit_fit_report_constellation_ledger((*C.SidereonOrbitFitReport)(pointer), (*C.SidereonOrbitConstellationResidualEntry)(out), n, w, q)
			}
		}
		if err := callStatus(func() uint32 { return invoke(nil, 0, &written, &required) }); err != nil {
			return err
		}
		count, err := validateNativeQuery(label, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		var size uintptr
		switch kind {
		case 0:
			size = unsafe.Sizeof(C.SidereonOrbitFitSolution{})
		case 1:
			size = unsafe.Sizeof(C.SidereonOrbitSatelliteResidualEntry{})
		default:
			size = unsafe.Sizeof(C.SidereonOrbitConstellationResidualEntry{})
		}
		if _, err := checkedNativeAllocationSize(count, size); err != nil {
			return err
		}
		if kind == 0 {
			values := make([]C.SidereonOrbitFitSolution, count)
			var out *C.SidereonOrbitFitSolution
			if count > 0 {
				out = &values[0]
			}
			written, required = 0, 0
			if err := callStatus(func() uint32 { return invoke(unsafe.Pointer(out), C.size_t(count), &written, &required) }); err != nil {
				return err
			}
			n, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			outValues := make([]OrbitFitSolution, n)
			for i := range outValues {
				outValues[i], err = solutionFromC(values[i])
				if err != nil {
					return err
				}
			}
			result = outValues
			return nil
		}
		if kind == 1 {
			values := make([]C.SidereonOrbitSatelliteResidualEntry, count)
			var out *C.SidereonOrbitSatelliteResidualEntry
			if count > 0 {
				out = &values[0]
			}
			written, required = 0, 0
			if err := callStatus(func() uint32 { return invoke(unsafe.Pointer(out), C.size_t(count), &written, &required) }); err != nil {
				return err
			}
			n, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
			if err != nil {
				return err
			}
			outValues := make([]OrbitSatelliteResidualEntry, n)
			for i := range outValues {
				stats, err := orbitStatsFromC(values[i].stats)
				if err != nil {
					return err
				}
				outValues[i] = OrbitSatelliteResidualEntry{Satellite: tokenFromC(values[i].satellite), Stats: stats}
			}
			result = outValues
			return nil
		}
		values := make([]C.SidereonOrbitConstellationResidualEntry, count)
		var out *C.SidereonOrbitConstellationResidualEntry
		if count > 0 {
			out = &values[0]
		}
		written, required = 0, 0
		if err := callStatus(func() uint32 { return invoke(unsafe.Pointer(out), C.size_t(count), &written, &required) }); err != nil {
			return err
		}
		n, err := validateTwoPassCounts(label, count, count, uint64(written), uint64(required))
		if err != nil {
			return err
		}
		outValues := make([]OrbitConstellationResidualEntry, n)
		for i := range outValues {
			stats, err := orbitStatsFromC(values[i].stats)
			if err != nil {
				return err
			}
			outValues[i] = OrbitConstellationResidualEntry{System: uint32(values[i].system), Stats: stats}
		}
		result = outValues
		return nil
	})
	return result, err
}
func (r *OrbitFitReport) Fits() ([]OrbitFitSolution, error) {
	value, err := reportValues(r, "orbit fit solutions", 0)
	if err != nil {
		return nil, err
	}
	return value.([]OrbitFitSolution), nil
}
func (r *OrbitFitReport) SatelliteLedger() ([]OrbitSatelliteResidualEntry, error) {
	value, err := reportValues(r, "orbit satellite residual ledger", 1)
	if err != nil {
		return nil, err
	}
	return value.([]OrbitSatelliteResidualEntry), nil
}
func (r *OrbitFitReport) ConstellationLedger() ([]OrbitConstellationResidualEntry, error) {
	value, err := reportValues(r, "orbit constellation residual ledger", 2)
	if err != nil {
		return nil, err
	}
	return value.([]OrbitConstellationResidualEntry), nil
}
func (r *OrbitFitReport) ArcSpan() (OrbitArcSpan, error) {
	if r == nil || r.handle == nil {
		return OrbitArcSpan{}, ErrClosed
	}
	var out C.SidereonOrbitArcSpan
	err := r.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_orbit_fit_report_arc_span((*C.SidereonOrbitFitReport)(pointer), &out) })
	})
	return OrbitArcSpan{TimeScale: uint32(out.time_scale), StartJ2000S: float64(out.start_j2000_s), EndJ2000S: float64(out.end_j2000_s), DurationS: float64(out.duration_s)}, err
}
func OrbitFitOptionsDefaults() (OrbitFitOptions, error) {
	var out C.SidereonOrbitFitOptions
	err := callStatus(func() uint32 { return C.sidereon_orbit_fit_options_init(&out) })
	if err != nil {
		return OrbitFitOptions{}, err
	}
	return orbitFitOptionsFromC(out)
}
