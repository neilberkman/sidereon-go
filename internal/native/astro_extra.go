//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type DragParameters struct {
	BCFactorM2PerKg  float64
	Weather          SpaceWeather
	CutoffAltitudeKm float64
}

type DecayConfig struct {
	ForceModel         uint32
	Integrator         uint32
	AbsTol             float64
	RelTol             float64
	InitialStepS       float64
	MinStepS           float64
	MaxStepS           float64
	MaxSteps           uint32
	MuKm3PerS2Enabled  bool
	MuKm3PerS2         float64
	Drag               DragParameters
	ReentryAltitudeKm  float64
	ScanStepS          float64
	CrossingToleranceS float64
	MaxDurationS       float64
	MaxScanSamples     uint32
}

type DecayEstimate struct {
	TimeToDecayS      float64
	ReentryState      CartesianState
	ReentryAltitudeKm float64
}

type SolarRadiationPressure struct {
	CR                float64
	AreaToMassM2PerKg float64
}

type ForceModelComponents struct {
	HasTwoBody                 bool
	TwoBodyMuKm3PerS2Enabled   bool
	TwoBodyMuKm3PerS2          float64
	HasZonal                   bool
	ZonalMaxDegree             uint32
	HasSphericalHarmonic       bool
	SphericalHarmonicMaxDegree uint32
	SphericalHarmonicMaxOrder  uint32
	HasSolidEarthTide          bool
	HasSolidEarthPoleTide      bool
	HasThirdBody               bool
	ThirdBodySun               bool
	ThirdBodyMoon              bool
	HasSolarRadiationPressure  bool
	SolarRadiationPressure     SolarRadiationPressure
	HasRelativity              bool
}

type TerrestrialFrame uint32

const (
	TerrestrialFrameITRF2020Value TerrestrialFrame = C.SIDEREON_TERRESTRIAL_FRAME_ITRF2020
	TerrestrialFrameITRF2014Value TerrestrialFrame = C.SIDEREON_TERRESTRIAL_FRAME_ITRF2014
	TerrestrialFrameITRF2008Value TerrestrialFrame = C.SIDEREON_TERRESTRIAL_FRAME_ITRF2008
	TerrestrialFrameETRF2020Value TerrestrialFrame = C.SIDEREON_TERRESTRIAL_FRAME_ETRF2020
)

type HelmertParameters struct {
	TranslationMm [3]float64
	ScalePPB      float64
	RotationMAS   [3]float64
}

type HelmertRates struct {
	TranslationMmPerYear [3]float64
	ScalePPBPerYear      float64
	RotationMASPerYear   [3]float64
}

type HelmertTransform struct {
	From          TerrestrialFrame
	To            TerrestrialFrame
	ReferenceYear float64
	Parameters    HelmertParameters
	Rates         HelmertRates
	Provenance    string
}

type TerrestrialPosition struct{ PositionM [3]float64 }
type TerrestrialVelocity struct{ VelocityMPerYear [3]float64 }
type TerrestrialState struct {
	Position    TerrestrialPosition
	HasVelocity bool
	Velocity    TerrestrialVelocity
}

type GibbsResult struct {
	VelocityKmPerS [3]float64
	Theta12Rad     float64
	Theta23Rad     float64
	CoplanarRad    float64
}

type GaussAnglesResult struct {
	PositionKm     [3]float64
	VelocityKmPerS [3]float64
}

func cSpaceWeather(value SpaceWeather) C.SidereonSpaceWeather {
	return C.SidereonSpaceWeather{f107: C.double(value.F107), f107a: C.double(value.F107A), ap: C.double(value.Ap)}
}

func cDragParameters(value DragParameters) C.SidereonDragParameters {
	return C.SidereonDragParameters{
		bc_factor_m2_kg:    C.double(value.BCFactorM2PerKg),
		space_weather:      cSpaceWeather(value.Weather),
		cutoff_altitude_km: C.double(value.CutoffAltitudeKm),
	}
}

func dragParametersFromC(value C.SidereonDragParameters) DragParameters {
	return DragParameters{
		BCFactorM2PerKg:  float64(value.bc_factor_m2_kg),
		Weather:          SpaceWeather{F107: float64(value.space_weather.f107), F107A: float64(value.space_weather.f107a), Ap: float64(value.space_weather.ap)},
		CutoffAltitudeKm: float64(value.cutoff_altitude_km),
	}
}

func cDecayConfig(value DecayConfig) (C.SidereonDecayConfig, error) {
	for _, item := range []struct {
		value float64
		name  string
	}{
		{value.InitialStepS, "initial step"},
		{value.MinStepS, "minimum step"},
		{value.MaxStepS, "maximum step"},
		{value.ReentryAltitudeKm, "reentry altitude"},
		{value.ScanStepS, "scan step"},
		{value.CrossingToleranceS, "crossing tolerance"},
		{value.MaxDurationS, "maximum duration"},
	} {
		if item.value < 0 {
			return C.SidereonDecayConfig{}, invalidArgument("decay " + item.name + " must not be negative")
		}
	}
	forceModel, err := cPropagationForceModel(value.ForceModel)
	if err != nil {
		return C.SidereonDecayConfig{}, err
	}
	integrator, err := cPropagationIntegrator(value.Integrator)
	if err != nil {
		return C.SidereonDecayConfig{}, err
	}
	return C.SidereonDecayConfig{
		force_model: forceModel, integrator: integrator,
		abs_tol: C.double(value.AbsTol), rel_tol: C.double(value.RelTol),
		initial_step_s: C.double(value.InitialStepS), min_step_s: C.double(value.MinStepS), max_step_s: C.double(value.MaxStepS),
		max_steps: C.uint32_t(value.MaxSteps), mu_km3_s2_enabled: C.bool(value.MuKm3PerS2Enabled), mu_km3_s2: C.double(value.MuKm3PerS2),
		drag: cDragParameters(value.Drag), reentry_altitude_km: C.double(value.ReentryAltitudeKm),
		scan_step_s: C.double(value.ScanStepS), crossing_tolerance_s: C.double(value.CrossingToleranceS),
		max_duration_s: C.double(value.MaxDurationS), max_scan_samples: C.uint32_t(value.MaxScanSamples),
	}, nil
}

func decayConfigFromC(value C.SidereonDecayConfig) DecayConfig {
	return DecayConfig{
		ForceModel: uint32(value.force_model), Integrator: uint32(value.integrator), AbsTol: float64(value.abs_tol), RelTol: float64(value.rel_tol),
		InitialStepS: float64(value.initial_step_s), MinStepS: float64(value.min_step_s), MaxStepS: float64(value.max_step_s), MaxSteps: uint32(value.max_steps),
		MuKm3PerS2Enabled: bool(value.mu_km3_s2_enabled), MuKm3PerS2: float64(value.mu_km3_s2), Drag: dragParametersFromC(value.drag),
		ReentryAltitudeKm: float64(value.reentry_altitude_km), ScanStepS: float64(value.scan_step_s), CrossingToleranceS: float64(value.crossing_tolerance_s),
		MaxDurationS: float64(value.max_duration_s), MaxScanSamples: uint32(value.max_scan_samples),
	}
}

func cSolarRadiationPressure(value SolarRadiationPressure) C.SidereonSolarRadiationPressure {
	return C.SidereonSolarRadiationPressure{cr: C.double(value.CR), area_to_mass_m2_kg: C.double(value.AreaToMassM2PerKg)}
}

func cForceModelComponents(value ForceModelComponents) C.SidereonForceModelComponents {
	return C.SidereonForceModelComponents{
		has_two_body: C.bool(value.HasTwoBody), two_body_mu_km3_s2_enabled: C.bool(value.TwoBodyMuKm3PerS2Enabled), two_body_mu_km3_s2: C.double(value.TwoBodyMuKm3PerS2),
		has_zonal: C.bool(value.HasZonal), zonal_max_degree: C.uint32_t(value.ZonalMaxDegree),
		has_spherical_harmonic: C.bool(value.HasSphericalHarmonic), spherical_harmonic_max_degree: C.uint32_t(value.SphericalHarmonicMaxDegree), spherical_harmonic_max_order: C.uint32_t(value.SphericalHarmonicMaxOrder),
		has_solid_earth_tide: C.bool(value.HasSolidEarthTide), has_solid_earth_pole_tide: C.bool(value.HasSolidEarthPoleTide),
		has_third_body: C.bool(value.HasThirdBody), third_body_sun: C.bool(value.ThirdBodySun), third_body_moon: C.bool(value.ThirdBodyMoon),
		has_solar_radiation_pressure: C.bool(value.HasSolarRadiationPressure), solar_radiation_pressure: cSolarRadiationPressure(value.SolarRadiationPressure), has_relativity: C.bool(value.HasRelativity),
	}
}

func cTerrestrialFrame(value TerrestrialFrame) (C.uint32_t, error) {
	switch value {
	case TerrestrialFrameITRF2020Value, TerrestrialFrameITRF2014Value, TerrestrialFrameITRF2008Value, TerrestrialFrameETRF2020Value:
		return C.uint32_t(value), nil
	default:
		return 0, invalidArgument("terrestrial frame is not defined by the C ABI")
	}
}

func cTerrestrialPosition(value TerrestrialPosition) C.SidereonTerrestrialPosition {
	var out C.SidereonTerrestrialPosition
	for i := range value.PositionM {
		out.position_m[i] = C.double(value.PositionM[i])
	}
	return out
}

func terrestrialPositionFromC(value C.SidereonTerrestrialPosition) TerrestrialPosition {
	var out TerrestrialPosition
	for i := range out.PositionM {
		out.PositionM[i] = float64(value.position_m[i])
	}
	return out
}

func cTerrestrialVelocity(value TerrestrialVelocity) C.SidereonTerrestrialVelocity {
	var out C.SidereonTerrestrialVelocity
	for i := range value.VelocityMPerYear {
		out.velocity_m_per_year[i] = C.double(value.VelocityMPerYear[i])
	}
	return out
}

func terrestrialVelocityFromC(value C.SidereonTerrestrialVelocity) TerrestrialVelocity {
	var out TerrestrialVelocity
	for i := range out.VelocityMPerYear {
		out.VelocityMPerYear[i] = float64(value.velocity_m_per_year[i])
	}
	return out
}

func cTerrestrialState(value TerrestrialState) C.SidereonTerrestrialState {
	return C.SidereonTerrestrialState{position: cTerrestrialPosition(value.Position), has_velocity: C.bool(value.HasVelocity), velocity: cTerrestrialVelocity(value.Velocity)}
}

func terrestrialStateFromC(value C.SidereonTerrestrialState) TerrestrialState {
	return TerrestrialState{Position: terrestrialPositionFromC(value.position), HasVelocity: bool(value.has_velocity), Velocity: terrestrialVelocityFromC(value.velocity)}
}

func helmertTransformFromC(value C.SidereonHelmertTransform) HelmertTransform {
	var out HelmertTransform
	out.From, out.To, out.ReferenceYear = TerrestrialFrame(value.from), TerrestrialFrame(value.to), float64(value.reference_epoch_year)
	for i := 0; i < 3; i++ {
		out.Parameters.TranslationMm[i] = float64(value.parameters.translation_mm[i])
		out.Parameters.RotationMAS[i] = float64(value.parameters.rotation_mas[i])
		out.Rates.TranslationMmPerYear[i] = float64(value.rates.translation_mm_per_year[i])
		out.Rates.RotationMASPerYear[i] = float64(value.rates.rotation_mas_per_year[i])
	}
	out.Parameters.ScalePPB, out.Rates.ScalePPBPerYear = float64(value.parameters.scale_ppb), float64(value.rates.scale_ppb_per_year)
	out.Provenance = C.GoString(&value.provenance[0])
	return out
}

type SPK struct {
	_      noCopy
	handle *positioningHandle
}

func releaseSPK(pointer unsafe.Pointer) { C.sidereon_spk_free((*C.SidereonSpk)(pointer)) }

func LoadSPK(data []byte) (*SPK, error) {
	if len(data) == 0 {
		return nil, invalidArgument("SPK data must not be empty")
	}
	if _, err := checkedNativeSize(len(data)); err != nil {
		return nil, err
	}
	p := C.CBytes(data)
	if p == nil {
		return nil, errors.New("sidereon: unable to allocate native SPK input")
	}
	defer C.free(p)
	var out *C.SidereonSpk
	err := callStatus(func() uint32 { return C.sidereon_spk_load((*C.uint8_t)(p), C.size_t(len(data)), &out) })
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_spk_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("SPK load")
	}
	return &SPK{handle: newPositioningHandle(unsafe.Pointer(out), releaseSPK)}, nil
}

func (s *SPK) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

type SPKState struct {
	Target, Center    int32
	PositionKm        [3]float64
	HasVelocity       bool
	HasVelocityKmPerS bool
	VelocityKmPerS    [3]float64
	Frame             int32
}

func (s *SPK) State(target, center int32, etSecondsTDB float64) (SPKState, error) {
	if s == nil || s.handle == nil {
		return SPKState{}, ErrClosed
	}
	var out C.SidereonSpkState
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return C.sidereon_spk_state((*C.SidereonSpk)(pointer), C.int32_t(target), C.int32_t(center), C.double(etSecondsTDB), &out)
		})
	})
	hasVelocity := bool(out.has_velocity_km_s)
	result := SPKState{Target: int32(out.target), Center: int32(out.center), HasVelocity: hasVelocity, HasVelocityKmPerS: hasVelocity, Frame: int32(out.frame)}
	for i := 0; i < 3; i++ {
		result.PositionKm[i], result.VelocityKmPerS[i] = float64(out.position_km[i]), float64(out.velocity_km_s[i])
	}
	return result, err
}

func DragParametersFromAreaMass(cd, areaM2, massKg float64, weather SpaceWeather, cutoffAltitudeKm float64) (DragParameters, error) {
	var out C.SidereonDragParameters
	err := callStatus(func() uint32 {
		return C.sidereon_drag_parameters_from_area_mass(C.double(cd), C.double(areaM2), C.double(massKg), cSpaceWeather(weather), C.double(cutoffAltitudeKm), &out)
	})
	return dragParametersFromC(out), err
}

func DragParametersFromBCFactor(factor float64, weather SpaceWeather, cutoffAltitudeKm float64) (DragParameters, error) {
	var out C.SidereonDragParameters
	err := callStatus(func() uint32 {
		return C.sidereon_drag_parameters_from_bc_factor(C.double(factor), cSpaceWeather(weather), C.double(cutoffAltitudeKm), &out)
	})
	return dragParametersFromC(out), err
}

func DragParametersFromBallisticCoefficient(bcKgM2 float64, weather SpaceWeather, cutoffAltitudeKm float64) (DragParameters, error) {
	var out C.SidereonDragParameters
	err := callStatus(func() uint32 {
		return C.sidereon_drag_parameters_from_ballistic_coefficient(C.double(bcKgM2), cSpaceWeather(weather), C.double(cutoffAltitudeKm), &out)
	})
	return dragParametersFromC(out), err
}

func DragForceAcceleration(drag DragParameters, state CartesianState) ([3]float64, error) {
	var out [3]C.double
	cState := cCartesianState(state)
	cDrag := cDragParameters(drag)
	err := callStatus(func() uint32 { return C.sidereon_drag_force_acceleration(&cDrag, &cState, &out[0]) })
	var result [3]float64
	for i := range result {
		result[i] = float64(out[i])
	}
	return result, err
}

func ForceTwoBodyAcceleration(positionKm, velocityKmPerS [3]float64) ([3]float64, error) {
	var position, velocity, out [3]C.double
	for i := 0; i < 3; i++ {
		position[i], velocity[i] = C.double(positionKm[i]), C.double(velocityKmPerS[i])
	}
	err := callStatus(func() uint32 { return C.sidereon_force_twobody_acceleration(&position[0], &velocity[0], &out[0]) })
	var result [3]float64
	for i := range result {
		result[i] = float64(out[i])
	}
	return result, err
}

func ForceJ2Acceleration(positionKm, velocityKmPerS [3]float64) ([3]float64, error) {
	var position, velocity, out [3]C.double
	for i := 0; i < 3; i++ {
		position[i], velocity[i] = C.double(positionKm[i]), C.double(velocityKmPerS[i])
	}
	err := callStatus(func() uint32 { return C.sidereon_force_j2_acceleration(&position[0], &velocity[0], &out[0]) })
	var result [3]float64
	for i := range result {
		result[i] = float64(out[i])
	}
	return result, err
}

func DecayConfigDefaults() (DecayConfig, error) {
	var out C.SidereonDecayConfig
	err := callStatus(func() uint32 { return C.sidereon_decay_config_init(&out) })
	return decayConfigFromC(out), err
}

func EstimateDecay(initial CartesianState, config DecayConfig) (DecayEstimate, error) {
	return estimateDecay(initial, config, nil)
}

func EstimateDecayWithSpaceWeatherTable(initial CartesianState, config DecayConfig, table *SpaceWeatherTable) (DecayEstimate, error) {
	if table == nil || table.handle == nil {
		return DecayEstimate{}, ErrClosed
	}
	return estimateDecay(initial, config, table)
}

func estimateDecay(initial CartesianState, config DecayConfig, table *SpaceWeatherTable) (DecayEstimate, error) {
	cConfig, err := cDecayConfig(config)
	if err != nil {
		return DecayEstimate{}, err
	}
	cInitial := cCartesianState(initial)
	var out C.SidereonDecayEstimate
	if table == nil {
		err = callStatus(func() uint32 { return C.sidereon_estimate_decay(&cInitial, &cConfig, &out) })
	} else {
		err = table.handle.with(func(pointer unsafe.Pointer) error {
			return callStatus(func() uint32 {
				return C.sidereon_estimate_decay_with_space_weather_table(&cInitial, &cConfig, (*C.SidereonSpaceWeatherTable)(pointer), &out)
			})
		})
	}
	return DecayEstimate{TimeToDecayS: float64(out.time_to_decay_s), ReentryState: cartesianStateFromC(out.reentry_state), ReentryAltitudeKm: float64(out.reentry_altitude_km)}, err
}

func IODGibbs(r1, r2, r3 [3]float64) (GibbsResult, error) {
	var a, b, c, v [3]C.double
	for i := 0; i < 3; i++ {
		a[i], b[i], c[i] = C.double(r1[i]), C.double(r2[i]), C.double(r3[i])
	}
	var t12, t23, cop C.double
	err := callStatus(func() uint32 { return C.sidereon_iod_gibbs(&a[0], &b[0], &c[0], &v[0], &t12, &t23, &cop) })
	var out GibbsResult
	for i := range out.VelocityKmPerS {
		out.VelocityKmPerS[i] = float64(v[i])
	}
	out.Theta12Rad, out.Theta23Rad, out.CoplanarRad = float64(t12), float64(t23), float64(cop)
	return out, err
}

func IODHGibbs(r1, r2, r3 [3]float64, jd1, jd2, jd3 float64) (GibbsResult, error) {
	var a, b, c, v [3]C.double
	for i := 0; i < 3; i++ {
		a[i], b[i], c[i] = C.double(r1[i]), C.double(r2[i]), C.double(r3[i])
	}
	var t12, t23, cop C.double
	err := callStatus(func() uint32 {
		return C.sidereon_iod_hgibbs(&a[0], &b[0], &c[0], C.double(jd1), C.double(jd2), C.double(jd3), &v[0], &t12, &t23, &cop)
	})
	var out GibbsResult
	for i := range out.VelocityKmPerS {
		out.VelocityKmPerS[i] = float64(v[i])
	}
	out.Theta12Rad, out.Theta23Rad, out.CoplanarRad = float64(t12), float64(t23), float64(cop)
	return out, err
}

func IODGaussAngles(declRad, rightAscensionRad, jd, jdFraction [3]float64, siteECIKm [3][3]float64) (GaussAnglesResult, error) {
	var d, r, j, f, sites [3]C.double
	for i := 0; i < 3; i++ {
		d[i], r[i], j[i], f[i] = C.double(declRad[i]), C.double(rightAscensionRad[i]), C.double(jd[i]), C.double(jdFraction[i])
		for k := 0; k < 3; k++ {
			sites[i*3+k] = C.double(siteECIKm[i][k])
		}
	}
	var p, v [3]C.double
	err := callStatus(func() uint32 { return C.sidereon_iod_gauss_angles(&d[0], &r[0], &j[0], &f[0], &sites[0], &p[0], &v[0]) })
	var out GaussAnglesResult
	for i := 0; i < 3; i++ {
		out.PositionKm[i], out.VelocityKmPerS[i] = float64(p[i]), float64(v[i])
	}
	return out, err
}

func FrameCatalogCount() (int, error) {
	var out C.size_t
	err := callStatus(func() uint32 { return C.sidereon_frame_catalog_count(&out) })
	if err != nil {
		return 0, err
	}
	return checkedNativeCount(uint64(out))
}

func FrameCatalogEntries() ([]HelmertTransform, error) {
	count, err := FrameCatalogCount()
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(C.SidereonHelmertTransform{})); err != nil {
		return nil, err
	}
	values := make([]C.SidereonHelmertTransform, count)
	var written, required C.size_t
	var ptr *C.SidereonHelmertTransform
	if count > 0 {
		ptr = &values[0]
	}
	err = callStatus(func() uint32 { return C.sidereon_frame_catalog_entries(ptr, C.size_t(count), &written, &required) })
	if err != nil {
		return nil, err
	}
	n, err := validateTwoPassCounts("frame catalog entries", count, count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	out := make([]HelmertTransform, n)
	for i := range out {
		out[i] = helmertTransformFromC(values[i])
	}
	return out, nil
}

func FrameCatalogEntry(from, to TerrestrialFrame) (HelmertTransform, error) {
	f, err := cTerrestrialFrame(from)
	if err != nil {
		return HelmertTransform{}, err
	}
	t, err := cTerrestrialFrame(to)
	if err != nil {
		return HelmertTransform{}, err
	}
	var out C.SidereonHelmertTransform
	err = callStatus(func() uint32 { return C.sidereon_frame_catalog_entry(f, t, &out) })
	return helmertTransformFromC(out), err
}

func FrameCatalogPropagatePosition(position TerrestrialPosition, velocity TerrestrialVelocity, fromYear, toYear float64) (TerrestrialPosition, error) {
	p, v := cTerrestrialPosition(position), cTerrestrialVelocity(velocity)
	var out C.SidereonTerrestrialPosition
	err := callStatus(func() uint32 {
		return C.sidereon_frame_catalog_propagate_position(&p, &v, C.double(fromYear), C.double(toYear), &out)
	})
	return terrestrialPositionFromC(out), err
}

func FrameCatalogTransform(state TerrestrialState, from, to TerrestrialFrame, epochYear float64) (TerrestrialState, error) {
	f, err := cTerrestrialFrame(from)
	if err != nil {
		return TerrestrialState{}, err
	}
	t, err := cTerrestrialFrame(to)
	if err != nil {
		return TerrestrialState{}, err
	}
	s := cTerrestrialState(state)
	var out C.SidereonTerrestrialState
	err = callStatus(func() uint32 { return C.sidereon_frame_catalog_transform(&s, f, t, C.double(epochYear), &out) })
	return terrestrialStateFromC(out), err
}

func FrameCatalogTransformFromEpoch(position TerrestrialPosition, velocity TerrestrialVelocity, positionYear float64, from, to TerrestrialFrame, transformYear float64) (TerrestrialState, error) {
	f, err := cTerrestrialFrame(from)
	if err != nil {
		return TerrestrialState{}, err
	}
	t, err := cTerrestrialFrame(to)
	if err != nil {
		return TerrestrialState{}, err
	}
	p, v := cTerrestrialPosition(position), cTerrestrialVelocity(velocity)
	var out C.SidereonTerrestrialState
	err = callStatus(func() uint32 {
		return C.sidereon_frame_catalog_transform_from_epoch(&p, &v, C.double(positionYear), f, t, C.double(transformYear), &out)
	})
	return terrestrialStateFromC(out), err
}
