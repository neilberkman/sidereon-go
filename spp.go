package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SPPObservation is one measured pseudorange in meters.
type SPPObservation struct {
	// SatelliteID is the GNSS satellite identifier for the measured pseudorange.
	SatelliteID string
	// PseudorangeM is the pseudorange m in metres.
	PseudorangeM float64
}

// SPPConfig is the legacy C SPP input surface. Domain validation and all
// numerical behavior remain in the C library.
type SPPConfig struct {
	// Observations contains a detached copy; nil means this field is absent.
	Observations []SPPObservation
	// TRxJ2000S is the receiver epoch in seconds from J2000.
	TRxJ2000S float64
	// TRxSecondOfDayS is the receiver UTC seconds since midnight.
	TRxSecondOfDayS float64
	// DayOfYear is the one-based fractional day of year used by atmospheric models.
	DayOfYear float64
	// InitialGuess is [ECEF x, y, z, receiver clock range bias], all in metres.
	InitialGuess [4]float64
	// Ionosphere requests the Klobuchar ionosphere correction.
	Ionosphere bool
	// Troposphere requests the configured troposphere correction.
	Troposphere bool
	// WithGeodetic requests latitude, longitude, and height alongside the ECEF result.
	WithGeodetic bool
	// KlobucharAlpha contains the four Klobuchar alpha coefficients [a0, a1, a2, a3].
	KlobucharAlpha [4]float64
	// KlobucharBeta contains the four Klobuchar beta coefficients [b0, b1, b2, b3].
	KlobucharBeta [4]float64
	// PressureHPA is the atmospheric pressure in hectopascals.
	PressureHPA float64
	// TemperatureK is the temperature k in kelvin.
	TemperatureK float64
	// RelativeHumidity is the relative humidity fraction.
	RelativeHumidity float64
	// Validation optionally applies C's receiver plausibility gates before the
	// detached solution is returned.
	Validation *SolutionValidationOptions
}

// SPPGeometryQuality contains C's geometry diagnostics.
type SPPGeometryQuality struct {
	// Tier is the geometry-quality tier.
	Tier uint32
	// Redundancy is the residual degrees of freedom.
	Redundancy int32
	// Rank is the matrix rank.
	Rank int
	// ConditionNumber is the dimensionless geometry-matrix condition number.
	ConditionNumber float64
	// GDOP is the dimensionless geometric dilution of precision.
	GDOP float64
	// RAIMCheckable reports whether the available redundancy supports a RAIM check.
	RAIMCheckable bool
	// CovarianceValidated reports whether the native covariance passed validation.
	CovarianceValidated bool
}

// SPPMetadata contains detached C solver and geometry metadata.
type SPPMetadata struct {
	// Iterations is the native solver iteration count.
	Iterations int
	// Converged reports whether the solution converged.
	Converged bool
	// Status is the native status code.
	Status uint32
	// IonosphereApplied reports whether the ionosphere correction was applied.
	IonosphereApplied bool
	// TroposphereApplied reports whether the troposphere correction was applied.
	TroposphereApplied bool
	// OuterIterations is the number of robust outer-loop iterations.
	OuterIterations int
	// HasFinalRobustScale reports whether FinalRobustScale is valid for the solve.
	HasFinalRobustScale bool
	// FinalRobustScaleM is the final robust scale m in metres.
	FinalRobustScaleM float64
	// UsedCount is the number of observations retained in the solution.
	UsedCount int
	// SystemCount is the number of GNSS systems represented by retained observations.
	SystemCount int
	// Redundancy is the redundancy value.
	Redundancy int64
	// RAIMCheckable reports whether the solution has enough redundancy for RAIM.
	RAIMCheckable bool
	// GeometryQuality is the geometry-quality diagnostics.
	GeometryQuality SPPGeometryQuality
}

// SPPSolution owns only Go memory. It contains no live C solution handle and
// can be copied as a value after SolveSPP returns.
type SPPSolution struct {
	// PositionM is the position m in metres.
	PositionM [3]float64
	// ReceiverClockS is the receiver clock s in seconds.
	ReceiverClockS float64
	// UsedSatelliteCount is the number of GNSS satellites contributing to the solution.
	UsedSatelliteCount int
	// UsedSatelliteIDs contains a detached copy; nil means this field is absent.
	UsedSatelliteIDs []string
	// ResidualsM is the residuals m in metres.
	ResidualsM []float64
	// DOP contains a detached copy; nil means this field is absent.
	DOP *DOP
	// Geodetic contains a detached copy; nil means this field is absent.
	Geodetic *Geodetic
	// Metadata is the metadata for this record.
	Metadata SPPMetadata
}

func publicSPPSolution(result native.SPPSolution) SPPSolution {
	out := SPPSolution{
		PositionM: result.PositionM, ReceiverClockS: result.ReceiverClockS,
		UsedSatelliteCount: result.UsedSatelliteCount,
		UsedSatelliteIDs:   append([]string(nil), result.UsedSatelliteIDs...),
		ResidualsM:         append([]float64(nil), result.ResidualsM...),
		Metadata: SPPMetadata{
			Iterations: result.Metadata.Iterations, Converged: result.Metadata.Converged, Status: result.Metadata.Status,
			IonosphereApplied: result.Metadata.IonosphereApplied, TroposphereApplied: result.Metadata.TroposphereApplied,
			OuterIterations: result.Metadata.OuterIterations, HasFinalRobustScale: result.Metadata.HasFinalRobustScale,
			FinalRobustScaleM: result.Metadata.FinalRobustScaleM, UsedCount: result.Metadata.UsedCount,
			SystemCount: result.Metadata.SystemCount, Redundancy: result.Metadata.Redundancy,
			RAIMCheckable: result.Metadata.RAIMCheckable,
			GeometryQuality: SPPGeometryQuality{
				Tier: result.Metadata.GeometryQuality.Tier, Redundancy: result.Metadata.GeometryQuality.Redundancy,
				Rank: result.Metadata.GeometryQuality.Rank, ConditionNumber: result.Metadata.GeometryQuality.ConditionNumber,
				GDOP: result.Metadata.GeometryQuality.GDOP, RAIMCheckable: result.Metadata.GeometryQuality.RAIMCheckable,
				CovarianceValidated: result.Metadata.GeometryQuality.CovarianceValidated,
			},
		},
	}
	if result.DOP != nil {
		out.DOP = &DOP{GDOP: result.DOP.GDOP, PDOP: result.DOP.PDOP, HDOP: result.DOP.HDOP, VDOP: result.DOP.VDOP, TDOP: result.DOP.TDOP}
	}
	if result.Geodetic != nil {
		out.Geodetic = &Geodetic{LatitudeRad: result.Geodetic.LatitudeRad, LongitudeRad: result.Geodetic.LongitudeRad, HeightM: result.Geodetic.HeightM}
	}
	return out
}

// SolveSPP solves one single-point positioning epoch through the C ABI.
func SolveSPP(sp3 *SP3, config SPPConfig) (SPPSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	nativeConfig := native.SPPConfig{
		TRxJ2000S:        config.TRxJ2000S,
		TRxSecondOfDayS:  config.TRxSecondOfDayS,
		DayOfYear:        config.DayOfYear,
		InitialGuess:     config.InitialGuess,
		Ionosphere:       config.Ionosphere,
		Troposphere:      config.Troposphere,
		WithGeodetic:     config.WithGeodetic,
		KlobucharAlpha:   config.KlobucharAlpha,
		KlobucharBeta:    config.KlobucharBeta,
		PressureHPA:      config.PressureHPA,
		TemperatureK:     config.TemperatureK,
		RelativeHumidity: config.RelativeHumidity,
	}
	if config.Validation != nil {
		value := native.NativeSolutionValidationOptions{HasMaxPDOP: config.Validation.HasMaxPDOP, MaxPDOP: config.Validation.MaxPDOP, MinPlausibleRadiusM: config.Validation.MinPlausibleRadiusM, MaxPlausibleRadiusM: config.Validation.MaxPlausibleRadiusM, MaxConvergedResidualRMSM: config.Validation.MaxConvergedResidualRMSM}
		nativeConfig.Validation = &value
	}
	nativeConfig.Observations = make([]native.SPPObservation, len(config.Observations))
	for i, observation := range config.Observations {
		nativeConfig.Observations[i] = native.SPPObservation{
			SatelliteID:  observation.SatelliteID,
			PseudorangeM: observation.PseudorangeM,
		}
	}
	result, err := sp3.handle.Solve(nativeConfig)
	if err != nil {
		return SPPSolution{}, publicError(err)
	}
	return publicSPPSolution(result), nil
}

func nativeSPPConfig(config SPPConfig) native.SPPConfig {
	out := native.SPPConfig{
		TRxJ2000S:        config.TRxJ2000S,
		TRxSecondOfDayS:  config.TRxSecondOfDayS,
		DayOfYear:        config.DayOfYear,
		InitialGuess:     config.InitialGuess,
		Ionosphere:       config.Ionosphere,
		Troposphere:      config.Troposphere,
		WithGeodetic:     config.WithGeodetic,
		KlobucharAlpha:   config.KlobucharAlpha,
		KlobucharBeta:    config.KlobucharBeta,
		PressureHPA:      config.PressureHPA,
		TemperatureK:     config.TemperatureK,
		RelativeHumidity: config.RelativeHumidity,
	}
	if config.Validation != nil {
		value := native.NativeSolutionValidationOptions{HasMaxPDOP: config.Validation.HasMaxPDOP, MaxPDOP: config.Validation.MaxPDOP, MinPlausibleRadiusM: config.Validation.MinPlausibleRadiusM, MaxPlausibleRadiusM: config.Validation.MaxPlausibleRadiusM, MaxConvergedResidualRMSM: config.Validation.MaxConvergedResidualRMSM}
		out.Validation = &value
	}
	out.Observations = make([]native.SPPObservation, len(config.Observations))
	for i, observation := range config.Observations {
		out.Observations[i] = native.SPPObservation{
			SatelliteID:  observation.SatelliteID,
			PseudorangeM: observation.PseudorangeM,
		}
	}
	return out
}

func fromNativeSPPSolution(result native.SPPSolution) SPPSolution {
	out := SPPSolution{
		PositionM:          result.PositionM,
		ReceiverClockS:     result.ReceiverClockS,
		UsedSatelliteCount: result.UsedSatelliteCount,
		UsedSatelliteIDs:   append([]string(nil), result.UsedSatelliteIDs...),
		ResidualsM:         append([]float64(nil), result.ResidualsM...),
		Metadata: SPPMetadata{
			Iterations:          result.Metadata.Iterations,
			Converged:           result.Metadata.Converged,
			Status:              result.Metadata.Status,
			IonosphereApplied:   result.Metadata.IonosphereApplied,
			TroposphereApplied:  result.Metadata.TroposphereApplied,
			OuterIterations:     result.Metadata.OuterIterations,
			HasFinalRobustScale: result.Metadata.HasFinalRobustScale,
			FinalRobustScaleM:   result.Metadata.FinalRobustScaleM,
			UsedCount:           result.Metadata.UsedCount,
			SystemCount:         result.Metadata.SystemCount,
			Redundancy:          result.Metadata.Redundancy,
			RAIMCheckable:       result.Metadata.RAIMCheckable,
			GeometryQuality: SPPGeometryQuality{
				Tier:                result.Metadata.GeometryQuality.Tier,
				Redundancy:          result.Metadata.GeometryQuality.Redundancy,
				Rank:                result.Metadata.GeometryQuality.Rank,
				ConditionNumber:     result.Metadata.GeometryQuality.ConditionNumber,
				GDOP:                result.Metadata.GeometryQuality.GDOP,
				RAIMCheckable:       result.Metadata.GeometryQuality.RAIMCheckable,
				CovarianceValidated: result.Metadata.GeometryQuality.CovarianceValidated,
			},
		},
	}
	if result.DOP != nil {
		out.DOP = &DOP{
			GDOP: result.DOP.GDOP,
			PDOP: result.DOP.PDOP,
			HDOP: result.DOP.HDOP,
			VDOP: result.DOP.VDOP,
			TDOP: result.DOP.TDOP,
		}
	}
	if result.Geodetic != nil {
		out.Geodetic = &Geodetic{
			LatitudeRad:  result.Geodetic.LatitudeRad,
			LongitudeRad: result.Geodetic.LongitudeRad,
			HeightM:      result.Geodetic.HeightM,
		}
	}
	return out
}
