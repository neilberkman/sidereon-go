//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import "unsafe"

func orbitAccelerationPartialsFromC(value C.SidereonAccelerationPartials) AccelerationPartials {
	return AccelerationPartials{
		WrtPosition: *(*[9]float64)(unsafe.Pointer(&value.wrt_position)),
		WrtVelocity: *(*[9]float64)(unsafe.Pointer(&value.wrt_velocity)),
	}
}

func orbitAnalyticCartesianFromC(value C.SidereonAnalyticCartesian) AnalyticCartesian {
	return AnalyticCartesian{
		PositionKm:  *(*[3]float64)(unsafe.Pointer(&value.position_km)),
		VelocityKmS: *(*[3]float64)(unsafe.Pointer(&value.velocity_km_s)),
	}
}

func orbitAnalyticCartesianToC(value AnalyticCartesian) C.SidereonAnalyticCartesian {
	return C.SidereonAnalyticCartesian{
		position_km:   *(*[3]C.double)(unsafe.Pointer(&value.PositionKm)),
		velocity_km_s: *(*[3]C.double)(unsafe.Pointer(&value.VelocityKmS)),
	}
}

func orbitAnalyticKeplerianElementsFromC(value C.SidereonAnalyticKeplerianElements) AnalyticKeplerianElements {
	return AnalyticKeplerianElements{
		AKm:            float64(value.a_km),
		E:              float64(value.e),
		IRad:           float64(value.i_rad),
		ArgPerigeeRad:  float64(value.arg_perigee_rad),
		RaanRad:        float64(value.raan_rad),
		MeanAnomalyRad: float64(value.mean_anomaly_rad),
	}
}

func orbitAnalyticKeplerianElementsToC(value AnalyticKeplerianElements) C.SidereonAnalyticKeplerianElements {
	return C.SidereonAnalyticKeplerianElements{
		a_km:             C.double(value.AKm),
		e:                C.double(value.E),
		i_rad:            C.double(value.IRad),
		arg_perigee_rad:  C.double(value.ArgPerigeeRad),
		raan_rad:         C.double(value.RaanRad),
		mean_anomaly_rad: C.double(value.MeanAnomalyRad),
	}
}

func orbitAttitudeSampleToC(value AttitudeSample) C.SidereonAttitudeSample {
	return C.SidereonAttitudeSample{
		position_m:     *(*[3]C.double)(unsafe.Pointer(&value.PositionM)),
		velocity_m_s:   *(*[3]C.double)(unsafe.Pointer(&value.VelocityMS)),
		sun_position_m: *(*[3]C.double)(unsafe.Pointer(&value.SunPositionM)),
	}
}

func orbitBrouwerLyddaneMeanElementsFromC(value C.SidereonBrouwerLyddaneMeanElements) BrouwerLyddaneMeanElements {
	return BrouwerLyddaneMeanElements{
		Elements:   orbitAnalyticKeplerianElementsFromC(value.elements),
		Iterations: uint32(value.iterations),
	}
}

func orbitCircularElementsFromC(value C.SidereonCircularElements) CircularElements {
	return CircularElements{
		AKm:                     float64(value.a_km),
		Ex:                      float64(value.ex),
		Ey:                      float64(value.ey),
		IRad:                    float64(value.i_rad),
		RaanRad:                 float64(value.raan_rad),
		MeanLatitudeArgumentRad: float64(value.mean_latitude_argument_rad),
	}
}

func orbitCircularElementsToC(value CircularElements) C.SidereonCircularElements {
	return C.SidereonCircularElements{
		a_km:                       C.double(value.AKm),
		ex:                         C.double(value.Ex),
		ey:                         C.double(value.Ey),
		i_rad:                      C.double(value.IRad),
		raan_rad:                   C.double(value.RaanRad),
		mean_latitude_argument_rad: C.double(value.MeanLatitudeArgumentRad),
	}
}

func orbitCosineTurnParametersFromC(value C.SidereonCosineTurnParameters) CosineTurnParameters {
	return CosineTurnParameters{
		BetaLimitRad:       float64(value.beta_limit_rad),
		BetaLimitInclusive: bool(value.beta_limit_inclusive),
		StartAngle:         CosineTurnStartAngle(value.start_angle),
		StartAngleRad:      float64(value.start_angle_rad),
		PeriodS:            float64(value.period_s),
	}
}

func orbitCosineTurnParametersToC(value CosineTurnParameters) C.SidereonCosineTurnParameters {
	return C.SidereonCosineTurnParameters{
		beta_limit_rad:       C.double(value.BetaLimitRad),
		beta_limit_inclusive: C.bool(value.BetaLimitInclusive),
		start_angle:          C.uint32_t(value.StartAngle),
		start_angle_rad:      C.double(value.StartAngleRad),
		period_s:             C.double(value.PeriodS),
	}
}

func orbitDECoverageFromC(value C.SidereonDeCoverage) DECoverage {
	return DECoverage{
		StartTDBSeconds: float64(value.start_tdb_seconds),
		EndTDBSeconds:   float64(value.end_tdb_seconds),
	}
}

func orbitDEGMFromC(value C.SidereonDeGm) DEGM {
	return DEGM{
		Present:   bool(value.present),
		ValueM3S2: float64(value.value_m3_s2),
	}
}

func orbitDESunMoonFromC(value C.SidereonDeSunMoon) DESunMoon {
	return DESunMoon{
		Sun:  orbitDEStateFromC(value.sun),
		Moon: orbitDEStateFromC(value.moon),
	}
}

func orbitDEStateFromC(value C.SidereonDeState) DEState {
	return DEState{
		PositionM:  *(*[3]float64)(unsafe.Pointer(&value.position_m)),
		VelocityMS: *(*[3]float64)(unsafe.Pointer(&value.velocity_m_s)),
	}
}

func orbitDYBAxesFromC(value C.SidereonDybAxes) DYBAxes {
	return DYBAxes{
		D: *(*[3]float64)(unsafe.Pointer(&value.d)),
		Y: *(*[3]float64)(unsafe.Pointer(&value.y)),
		B: *(*[3]float64)(unsafe.Pointer(&value.b)),
	}
}

func orbitEarthOrientationFromC(value C.SidereonEarthOrientation) EarthOrientation {
	return EarthOrientation{
		TimeScales:                   orbitTimeScalesFromC(value.time_scales),
		PolarMotionRad:               *(*[2]float64)(unsafe.Pointer(&value.polar_motion_rad)),
		EarthRotationVectorITRFRadS:  *(*[3]float64)(unsafe.Pointer(&value.earth_rotation_vector_itrf_rad_s)),
		GCRFToITRFMatrix:             *(*[3][3]float64)(unsafe.Pointer(&value.gcrf_to_itrf_matrix)),
		ITRFToGCRFMatrix:             *(*[3][3]float64)(unsafe.Pointer(&value.itrf_to_gcrf_matrix)),
		GCRFToITRFRotationRateMatrix: *(*[3][3]float64)(unsafe.Pointer(&value.gcrf_to_itrf_rotation_rate_matrix)),
		ITRFToGCRFRotationRateMatrix: *(*[3][3]float64)(unsafe.Pointer(&value.itrf_to_gcrf_rotation_rate_matrix)),
		HasIERS2010Quaternion:        bool(value.has_iers2010_quaternion),
		IERS2010Quaternion:           *(*[4]float64)(unsafe.Pointer(&value.iers2010_quaternion)),
	}
}

func orbitEcksteinHechlerMeanElementsFromC(value C.SidereonEcksteinHechlerMeanElements) EcksteinHechlerMeanElements {
	return EcksteinHechlerMeanElements{
		Elements:   orbitCircularElementsFromC(value.elements),
		Iterations: uint32(value.iterations),
	}
}

func orbitECOM1CoefficientsFromC(value C.SidereonEcom1Coefficients) ECOM1Coefficients {
	return ECOM1Coefficients{
		D0: float64(value.d0),
		Dc: float64(value.dc),
		Ds: float64(value.ds),
		Y0: float64(value.y0),
		Yc: float64(value.yc),
		Ys: float64(value.ys),
		B0: float64(value.b0),
		Bc: float64(value.bc),
		Bs: float64(value.bs),
	}
}

func orbitECOM1CoefficientsToC(value ECOM1Coefficients) C.SidereonEcom1Coefficients {
	return C.SidereonEcom1Coefficients{
		d0: C.double(value.D0),
		dc: C.double(value.Dc),
		ds: C.double(value.Ds),
		y0: C.double(value.Y0),
		yc: C.double(value.Yc),
		ys: C.double(value.Ys),
		b0: C.double(value.B0),
		bc: C.double(value.Bc),
		bs: C.double(value.Bs),
	}
}

func orbitECOM2D2B1CoefficientsFromC(value C.SidereonEcom2D2b1Coefficients) ECOM2D2B1Coefficients {
	return ECOM2D2B1Coefficients{
		B0:  float64(value.b0),
		B1c: float64(value.b1c),
		B1s: float64(value.b1s),
		D0:  float64(value.d0),
		D2c: float64(value.d2c),
		D2s: float64(value.d2s),
		Y0:  float64(value.y0),
	}
}

func orbitECOM2D2B1CoefficientsToC(value ECOM2D2B1Coefficients) C.SidereonEcom2D2b1Coefficients {
	return C.SidereonEcom2D2b1Coefficients{
		b0:  C.double(value.B0),
		b1c: C.double(value.B1c),
		b1s: C.double(value.B1s),
		d0:  C.double(value.D0),
		d2c: C.double(value.D2c),
		d2s: C.double(value.D2s),
		y0:  C.double(value.Y0),
	}
}

func orbitECOM2D4B1CoefficientsFromC(value C.SidereonEcom2D4b1Coefficients) ECOM2D4B1Coefficients {
	return ECOM2D4B1Coefficients{
		B0:  float64(value.b0),
		B1c: float64(value.b1c),
		B1s: float64(value.b1s),
		D0:  float64(value.d0),
		D2c: float64(value.d2c),
		D4c: float64(value.d4c),
		D2s: float64(value.d2s),
		D4s: float64(value.d4s),
		Y0:  float64(value.y0),
	}
}

func orbitECOM2D4B1CoefficientsToC(value ECOM2D4B1Coefficients) C.SidereonEcom2D4b1Coefficients {
	return C.SidereonEcom2D4b1Coefficients{
		b0:  C.double(value.B0),
		b1c: C.double(value.B1c),
		b1s: C.double(value.B1s),
		d0:  C.double(value.D0),
		d2c: C.double(value.D2c),
		d4c: C.double(value.D4c),
		d2s: C.double(value.D2s),
		d4s: C.double(value.D4s),
		y0:  C.double(value.Y0),
	}
}

func orbitECOMGeometryFromC(value C.SidereonEcomGeometry) ECOMGeometry {
	return ECOMGeometry{
		DYB:                   orbitDYBAxesFromC(value.dyb),
		DeltaURad:             orbitOptionalOrbitScalarFromC(value.delta_u_rad),
		ArgumentOfLatitudeRad: orbitOptionalOrbitScalarFromC(value.argument_of_latitude_rad),
		LightingRatio:         float64(value.lighting_ratio),
	}
}

func orbitEmpiricalRTNTermFromC(value C.SidereonEmpiricalRtnTerm) EmpiricalRTNTerm {
	return EmpiricalRTNTerm{
		Axis:            RTNAxis(value.axis),
		Harmonic:        EmpiricalHarmonic(value.harmonic),
		CoefficientKmS2: float64(value.coefficient_km_s2),
	}
}

func orbitEmpiricalRTNTermToC(value EmpiricalRTNTerm) C.SidereonEmpiricalRtnTerm {
	return C.SidereonEmpiricalRtnTerm{
		axis:              C.uint32_t(value.Axis),
		harmonic:          C.uint32_t(value.Harmonic),
		coefficient_km_s2: C.double(value.CoefficientKmS2),
	}
}

func orbitEmpiricalRTNTermSelectorToC(value EmpiricalRTNTermSelector) C.SidereonEmpiricalRtnTermSelector {
	return C.SidereonEmpiricalRtnTermSelector{
		axis:     C.uint32_t(value.Axis),
		harmonic: C.uint32_t(value.Harmonic),
	}
}

func orbitEstimatedForceParameterFromC(value C.SidereonEstimatedForceParameter) EstimatedForceParameter {
	return EstimatedForceParameter{
		InitialValue: float64(value.initial_value),
		Value:        float64(value.value),
	}
}

func orbitHorizonBinStatsFromC(value C.SidereonHorizonBinStats) HorizonBinStats {
	return HorizonBinStats{
		HorizonStartS: float64(value.horizon_start_s),
		HorizonEndS:   float64(value.horizon_end_s),
		Count:         uint64(value.count),
		Radial:        orbitErrorStatsFromC(value.radial),
		Transverse:    orbitErrorStatsFromC(value.transverse),
		Normal:        orbitErrorStatsFromC(value.normal),
		Orbit3d:       orbitErrorStatsFromC(value.orbit_3d),
		SISREOrbit:    orbitErrorStatsFromC(value.sisre_orbit),
		Clock:         orbitErrorStatsFromC(value.clock),
		SISRE:         orbitErrorStatsFromC(value.sisre),
		ClockCount:    uint64(value.clock_count),
	}
}

func orbitErrorStatsFromC(value C.SidereonErrorStats) ErrorStats {
	return ErrorStats{
		Present: bool(value.present),
		RMSM:    float64(value.rms_m),
		MeanM:   float64(value.mean_m),
		MaxAbsM: float64(value.max_abs_m),
	}
}

func orbitICGEMHeaderFromC(value C.SidereonIcgemHeader) ICGEMHeader {
	return ICGEMHeader{
		GravityConstantM3S2: float64(value.gravity_constant_m3_s2),
		RadiusM:             float64(value.radius_m),
		MaxDegree:           uint16(value.max_degree),
		Errors:              ICGEMErrors(value.errors),
		Normalization:       ICGEMNormalization(value.normalization),
		HasTideSystem:       bool(value.has_tide_system),
		TideSystem:          GravityTideSystem(value.tide_system),
		Format:              ICGEMFormat(value.format),
	}
}

func orbitICGEMSigmasFromC(value C.SidereonIcgemSigmas) ICGEMSigmas {
	return ICGEMSigmas{
		HasCalibrated: bool(value.has_calibrated),
		Calibrated:    *(*[2]float64)(unsafe.Pointer(&value.calibrated)),
		HasFormal:     bool(value.has_formal),
		Formal:        *(*[2]float64)(unsafe.Pointer(&value.formal)),
	}
}

func orbitICGEMTimeVariableTermFromC(value C.SidereonIcgemTimeVariableTerm) ICGEMTimeVariableTerm {
	return ICGEMTimeVariableTerm{
		Line:                uint64(value.line),
		Kind:                ICGEMTermKind(value.kind),
		PeriodYears:         float64(value.period_years),
		Degree:              uint16(value.degree),
		Order:               uint16(value.order),
		C:                   float64(value.c),
		S:                   float64(value.s),
		Sigmas:              orbitICGEMSigmasFromC(value.sigmas),
		ReferenceTTSeconds:  float64(value.reference_tt_seconds),
		ValidFromTTSeconds:  float64(value.valid_from_tt_seconds),
		ValidUntilTTSeconds: float64(value.valid_until_tt_seconds),
	}
}

func orbitIERS2010EopSampleToC(value IERS2010EopSample) C.SidereonIers2010EopSample {
	return C.SidereonIers2010EopSample{
		mjd_utc:          C.int32_t(value.MJDUTC),
		xp_arcsec:        C.double(value.XpArcsec),
		yp_arcsec:        C.double(value.YpArcsec),
		ut1_minus_utc_s:  C.double(value.Ut1MinusUTCS),
		lod_s:            C.double(value.LodS),
		dx_arcsec:        C.double(value.DxArcsec),
		dy_arcsec:        C.double(value.DyArcsec),
		has_pole_rates:   C.bool(value.HasPoleRates),
		pole_rates_rad_s: *(*[2]C.double)(unsafe.Pointer(&value.PoleRatesRadS)),
	}
}

func orbitIntegratorOptionsFromC(value C.SidereonIntegratorOptions) IntegratorOptions {
	return IntegratorOptions{
		Integrator:   uint32(value.integrator),
		AbsTol:       float64(value.abs_tol),
		RelTol:       float64(value.rel_tol),
		InitialStepS: float64(value.initial_step_s),
		MinStepS:     float64(value.min_step_s),
		MaxStepS:     float64(value.max_step_s),
		MaxSteps:     uint32(value.max_steps),
	}
}

func orbitIntegratorOptionsToC(value IntegratorOptions) C.SidereonIntegratorOptions {
	return C.SidereonIntegratorOptions{
		integrator:     C.uint32_t(value.Integrator),
		abs_tol:        C.double(value.AbsTol),
		rel_tol:        C.double(value.RelTol),
		initial_step_s: C.double(value.InitialStepS),
		min_step_s:     C.double(value.MinStepS),
		max_step_s:     C.double(value.MaxStepS),
		max_steps:      C.uint32_t(value.MaxSteps),
	}
}

func orbitMeanConversionOptionsFromC(value C.SidereonMeanConversionOptions) MeanConversionOptions {
	return MeanConversionOptions{
		Tolerance:     float64(value.tolerance),
		MaxIterations: uint32(value.max_iterations),
	}
}

func orbitMeanConversionOptionsToC(value MeanConversionOptions) C.SidereonMeanConversionOptions {
	return C.SidereonMeanConversionOptions{
		tolerance:      C.double(value.Tolerance),
		max_iterations: C.uint32_t(value.MaxIterations),
	}
}

func orbitNominalYawFromC(value C.SidereonNominalYaw) NominalYaw {
	return NominalYaw{
		YawRad:  float64(value.yaw_rad),
		BetaRad: float64(value.beta_rad),
		BodyX:   *(*[3]float64)(unsafe.Pointer(&value.body_x)),
		BodyY:   *(*[3]float64)(unsafe.Pointer(&value.body_y)),
		BodyZ:   *(*[3]float64)(unsafe.Pointer(&value.body_z)),
	}
}

func orbitOptionalHarmonicCoefficientFromC(value C.SidereonOptionalHarmonicCoefficient) OptionalHarmonicCoefficient {
	return OptionalHarmonicCoefficient{
		Present:     bool(value.present),
		Coefficient: orbitSphericalHarmonicCoefficientFromC(value.coefficient),
	}
}

func orbitOptionalOrbitScalarFromC(value C.SidereonOptionalOrbitScalar) OptionalOrbitScalar {
	return OptionalOrbitScalar{
		Present: bool(value.present),
		Value:   float64(value.value),
	}
}

func orbitOrbitParameterCovarianceFromC(value C.SidereonOrbitParameterCovariance) OrbitParameterCovariance {
	return OrbitParameterCovariance{
		Estimated:        bool(value.estimated),
		Dimension:        uint64(value.dimension),
		VarianceFactor:   float64(value.variance_factor),
		DegreesOfFreedom: uint64(value.degrees_of_freedom),
	}
}

func orbitOrbitParameterFitSummaryFromC(value C.SidereonOrbitParameterFitSummary) OrbitParameterFitSummary {
	return OrbitParameterFitSummary{
		Satellite:           tokenFromC(value.satellite),
		InitialState:        cartesianStateFromC(value.initial_state),
		ForceParameterCount: uint64(value.force_parameter_count),
		ParameterCount:      uint64(value.parameter_count),
		ResidualStats:       orbitParameterStatsFromC(value.residual_stats),
		SeedRMS3dM:          float64(value.seed_rms_3d_m),
		Iterations:          uint64(value.iterations),
		ConditionNumber:     float64(value.condition_number),
		Ut1Degraded:         Ut1Degradation(value.ut1_degraded),
	}
}

func orbitOrbitRTNResidualFromC(value C.SidereonOrbitRtnResidual) OrbitRTNResidual {
	return OrbitRTNResidual{
		VariationalEpoch: orbitVariationalEpochFromC(value.variational_epoch),
		RadialM:          float64(value.radial_m),
		AlongM:           float64(value.along_m),
		CrossM:           float64(value.cross_m),
	}
}

func orbitOrbitScoreSummaryFromC(value C.SidereonOrbitScoreSummary) OrbitScoreSummary {
	return OrbitScoreSummary{
		Frame:                orbitFrameAgreementFromC(value.frame),
		SourceReferencePoint: OrbitReferencePoint(value.source_reference_point),
		SatelliteCount:       uint64(value.satellite_count),
		SystemCount:          uint64(value.system_count),
		Coverage:             orbitScoreCoverageFromC(value.coverage),
	}
}

func orbitFrameAgreementFromC(value C.SidereonFrameAgreement) FrameAgreement {
	return FrameAgreement{
		Kind: FrameAgreementKind(value.kind),
		From: uint32(value.from),
		To:   uint32(value.to),
	}
}

func orbitOrbitSourceInfoFromC(value C.SidereonOrbitSourceInfo) OrbitSourceInfo {
	return OrbitSourceInfo{
		TimeScale:               uint32(value.time_scale),
		HasSP3TimeSystem:        bool(value.has_sp3_time_system),
		Frame:                   OrbitFrame(value.frame),
		ReferencePoint:          OrbitReferencePoint(value.reference_point),
		ClockIncludesRelativity: bool(value.clock_includes_relativity),
	}
}

func orbitOrbitSourceStateFromC(value C.SidereonOrbitSourceState) OrbitSourceState {
	return OrbitSourceState{
		PositionECEFM: *(*[3]float64)(unsafe.Pointer(&value.position_ecef_m)),
		ClockS:        orbitOptionalOrbitScalarFromC(value.clock_s),
	}
}

func orbitRadiationShadowOptionsFromC(value C.SidereonRadiationShadowOptions) RadiationShadowOptions {
	return RadiationShadowOptions{
		ShadowFunction:     ShadowFunction(value.shadow_function),
		EarthShadowModel:   uint32(value.earth_shadow_model),
		MoonEclipseEnabled: bool(value.moon_eclipse_enabled),
	}
}

func orbitRadiationShadowOptionsToC(value RadiationShadowOptions) C.SidereonRadiationShadowOptions {
	return C.SidereonRadiationShadowOptions{
		shadow_function:      C.uint32_t(value.ShadowFunction),
		earth_shadow_model:   C.uint32_t(value.EarthShadowModel),
		moon_eclipse_enabled: C.bool(value.MoonEclipseEnabled),
	}
}

func orbitRTNComponentsFromC(value C.SidereonRtnComponents) RTNComponents {
	return RTNComponents{
		RadialM:     float64(value.radial_m),
		TransverseM: float64(value.transverse_m),
		NormalM:     float64(value.normal_m),
	}
}

func orbitRTNComponentsToC(value RTNComponents) C.SidereonRtnComponents {
	return C.SidereonRtnComponents{
		radial_m:     C.double(value.RadialM),
		transverse_m: C.double(value.TransverseM),
		normal_m:     C.double(value.NormalM),
	}
}

func orbitSatelliteScoreFromC(value C.SidereonSatelliteScore) SatelliteScore {
	return SatelliteScore{
		Satellite:              tokenFromC(value.satellite),
		PredictionOriginJ2000S: orbitOptionalOrbitScalarFromC(value.prediction_origin_j2000_s),
		HasOrbitRegime:         bool(value.has_orbit_regime),
		OrbitRegime:            OrbitRegime(value.orbit_regime),
		HasSISREWeights:        bool(value.has_sisre_weights),
		SISREWeights:           orbitSISREWeightsFromC(value.sisre_weights),
		BinCount:               uint64(value.bin_count),
		Coverage:               orbitScoreCoverageFromC(value.coverage),
	}
}

func orbitScoreCoverageFromC(value C.SidereonScoreCoverage) ScoreCoverage {
	return ScoreCoverage{
		Epochs:                    uint64(value.epochs),
		Scored:                    uint64(value.scored),
		Clocked:                   uint64(value.clocked),
		OutsideHorizonBins:        uint64(value.outside_horizon_bins),
		NoPredictionOrigin:        uint64(value.no_prediction_origin),
		SourceSatelliteMissing:    uint64(value.source_satellite_missing),
		SourceEpochMissing:        uint64(value.source_epoch_missing),
		ReferenceSatelliteMissing: uint64(value.reference_satellite_missing),
		ReferenceEpochMissing:     uint64(value.reference_epoch_missing),
	}
}

func orbitShadowGradientFromC(value C.SidereonShadowGradient) ShadowGradient {
	return ShadowGradient{
		Fraction:      float64(value.fraction),
		GradientPerKm: *(*[3]float64)(unsafe.Pointer(&value.gradient_per_km)),
	}
}

func orbitSISREWeightsFromC(value C.SidereonSisreWeights) SISREWeights {
	return SISREWeights{
		Radial:                  float64(value.radial),
		TransverseNormalSquared: float64(value.transverse_normal_squared),
	}
}

func orbitSISREWeightsToC(value SISREWeights) C.SidereonSisreWeights {
	return C.SidereonSisreWeights{
		radial:                    C.double(value.Radial),
		transverse_normal_squared: C.double(value.TransverseNormalSquared),
	}
}

func orbitSphericalHarmonicCoefficientFromC(value C.SidereonSphericalHarmonicCoefficient) SphericalHarmonicCoefficient {
	return SphericalHarmonicCoefficient{
		Degree: uint16(value.degree),
		Order:  uint16(value.order),
		C:      float64(value.c),
		S:      float64(value.s),
	}
}

func orbitSunMoonFromC(value C.SidereonSunMoon) SunMoon {
	return SunMoon{
		SunM:  *(*[3]float64)(unsafe.Pointer(&value.sun_m)),
		MoonM: *(*[3]float64)(unsafe.Pointer(&value.moon_m)),
	}
}

func orbitSystemScoreFromC(value C.SidereonSystemScore) SystemScore {
	return SystemScore{
		System:   uint32(value.system),
		BinCount: uint64(value.bin_count),
		Coverage: orbitScoreCoverageFromC(value.coverage),
	}
}

func orbitVariationalEpochFromC(value C.SidereonVariationalEpoch) VariationalEpoch {
	return VariationalEpoch{
		TDBJ2000S:  float64(value.tdb_j2000_s),
		ElapsedTTS: float64(value.elapsed_tt_s),
	}
}

func orbitVariationalStateFromC(value C.SidereonVariationalState) VariationalState {
	return VariationalState{
		State:              cartesianStateFromC(value.state),
		StateTransition:    *(*[36]float64)(unsafe.Pointer(&value.state_transition)),
		SensitivityColumns: uint64(value.sensitivity_columns),
	}
}

func orbitZonalHarmonicsFromC(value C.SidereonZonalHarmonics) ZonalHarmonics {
	return ZonalHarmonics{
		MuKm3S2: float64(value.mu_km3_s2),
		ReKm:    float64(value.re_km),
		C20:     float64(value.c20),
		C30:     float64(value.c30),
		C40:     float64(value.c40),
		C50:     float64(value.c50),
		C60:     float64(value.c60),
	}
}

func orbitZonalHarmonicsToC(value ZonalHarmonics) C.SidereonZonalHarmonics {
	return C.SidereonZonalHarmonics{
		mu_km3_s2: C.double(value.MuKm3S2),
		re_km:     C.double(value.ReKm),
		c20:       C.double(value.C20),
		c30:       C.double(value.C30),
		c40:       C.double(value.C40),
		c50:       C.double(value.C50),
		c60:       C.double(value.C60),
	}
}

type BrouwerLyddanePropagator struct {
	_      noCopy
	handle *positioningHandle
}

func (v *BrouwerLyddanePropagator) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *BrouwerLyddanePropagator) Close() error { return v.orbitHandle().close() }
func newOrbitBrouwerLyddanePropagator(pointer *C.SidereonBrouwerLyddanePropagator, err error) (*BrouwerLyddanePropagator, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_brouwer_lyddane_propagator_free((*C.SidereonBrouwerLyddanePropagator)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &BrouwerLyddanePropagator{handle: handle}, nil
}

type DEEphemeris struct {
	_      noCopy
	handle *positioningHandle
}

func (v *DEEphemeris) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *DEEphemeris) Close() error { return v.orbitHandle().close() }
func newOrbitDEEphemeris(pointer *C.SidereonDeEphemeris, err error) (*DEEphemeris, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_de_ephemeris_free((*C.SidereonDeEphemeris)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &DEEphemeris{handle: handle}, nil
}

type EcksteinHechlerPropagator struct {
	_      noCopy
	handle *positioningHandle
}

func (v *EcksteinHechlerPropagator) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *EcksteinHechlerPropagator) Close() error { return v.orbitHandle().close() }
func newOrbitEcksteinHechlerPropagator(pointer *C.SidereonEcksteinHechlerPropagator, err error) (*EcksteinHechlerPropagator, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_eckstein_hechler_propagator_free((*C.SidereonEcksteinHechlerPropagator)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &EcksteinHechlerPropagator{handle: handle}, nil
}

type ForceModel struct {
	_      noCopy
	handle *positioningHandle
}

func (v *ForceModel) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *ForceModel) Close() error { return v.orbitHandle().close() }
func newOrbitForceModel(pointer *C.SidereonForceModel, err error) (*ForceModel, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_force_model_free((*C.SidereonForceModel)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &ForceModel{handle: handle}, nil
}

type GNSSAttitudeLaw struct {
	_      noCopy
	handle *positioningHandle
}

func (v *GNSSAttitudeLaw) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *GNSSAttitudeLaw) Close() error { return v.orbitHandle().close() }
func newOrbitGNSSAttitudeLaw(pointer *C.SidereonGnssAttitudeLaw, err error) (*GNSSAttitudeLaw, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_gnss_attitude_law_free((*C.SidereonGnssAttitudeLaw)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &GNSSAttitudeLaw{handle: handle}, nil
}

type ICGEMGravityField struct {
	_      noCopy
	handle *positioningHandle
}

func (v *ICGEMGravityField) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *ICGEMGravityField) Close() error { return v.orbitHandle().close() }
func newOrbitICGEMGravityField(pointer *C.SidereonIcgemGravityField, err error) (*ICGEMGravityField, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_icgem_gravity_field_free((*C.SidereonIcgemGravityField)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &ICGEMGravityField{handle: handle}, nil
}

type IERS2010EarthOrientationProvider struct {
	_      noCopy
	handle *positioningHandle
}

func (v *IERS2010EarthOrientationProvider) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *IERS2010EarthOrientationProvider) Close() error { return v.orbitHandle().close() }
func newOrbitIERS2010EarthOrientationProvider(pointer *C.SidereonIers2010EarthOrientationProvider, err error) (*IERS2010EarthOrientationProvider, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_iers2010_earth_orientation_provider_free((*C.SidereonIers2010EarthOrientationProvider)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &IERS2010EarthOrientationProvider{handle: handle}, nil
}

type OrbitParameterFit struct {
	_      noCopy
	handle *positioningHandle
}

func (v *OrbitParameterFit) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *OrbitParameterFit) Close() error { return v.orbitHandle().close() }
func newOrbitOrbitParameterFit(pointer *C.SidereonOrbitParameterFit, err error) (*OrbitParameterFit, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_orbit_parameter_fit_free((*C.SidereonOrbitParameterFit)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &OrbitParameterFit{handle: handle}, nil
}

type OrbitScoreReport struct {
	_      noCopy
	handle *positioningHandle
}

func (v *OrbitScoreReport) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *OrbitScoreReport) Close() error { return v.orbitHandle().close() }
func newOrbitOrbitScoreReport(pointer *C.SidereonOrbitScoreReport, err error) (*OrbitScoreReport, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_orbit_score_report_free((*C.SidereonOrbitScoreReport)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &OrbitScoreReport{handle: handle}, nil
}

type OrbitSource struct {
	_      noCopy
	handle *positioningHandle
}

func (v *OrbitSource) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *OrbitSource) Close() error { return v.orbitHandle().close() }
func newOrbitOrbitSource(pointer *C.SidereonOrbitSource, err error) (*OrbitSource, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_orbit_source_free((*C.SidereonOrbitSource)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &OrbitSource{handle: handle}, nil
}

type SphericalHarmonicGravity struct {
	_      noCopy
	handle *positioningHandle
}

func (v *SphericalHarmonicGravity) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *SphericalHarmonicGravity) Close() error { return v.orbitHandle().close() }
func newOrbitSphericalHarmonicGravity(pointer *C.SidereonSphericalHarmonicGravity, err error) (*SphericalHarmonicGravity, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_spherical_harmonic_gravity_free((*C.SidereonSphericalHarmonicGravity)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &SphericalHarmonicGravity{handle: handle}, nil
}

type VariationalTrajectory struct {
	_      noCopy
	handle *positioningHandle
}

func (v *VariationalTrajectory) orbitHandle() *positioningHandle {
	if v == nil {
		return nil
	}
	return v.handle
}
func (v *VariationalTrajectory) Close() error { return v.orbitHandle().close() }
func newOrbitVariationalTrajectory(pointer *C.SidereonVariationalTrajectory, err error) (*VariationalTrajectory, error) {
	handle, err := adoptOrbitHandle(unsafe.Pointer(pointer), err, func(pointer unsafe.Pointer) {
		C.sidereon_variational_trajectory_free((*C.SidereonVariationalTrajectory)(pointer))
	})
	if err != nil {
		return nil, err
	}
	return &VariationalTrajectory{handle: handle}, nil
}
