package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SPPObservation is one measured pseudorange in meters.
type SPPObservation struct {
	SatelliteID  string
	PseudorangeM float64
}

// SPPConfig is the legacy C SPP input surface. Domain validation and all
// numerical behavior remain in the C library.
type SPPConfig struct {
	Observations    []SPPObservation
	TRxJ2000S       float64
	TRxSecondOfDayS float64
	DayOfYear       float64
	InitialGuess    [4]float64
	Ionosphere      bool
	Troposphere     bool
	WithGeodetic    bool
}

// SPPGeometryQuality contains C's geometry diagnostics.
type SPPGeometryQuality struct {
	Tier                uint32
	Redundancy          int32
	Rank                int
	ConditionNumber     float64
	GDOP                float64
	RAIMCheckable       bool
	CovarianceValidated bool
}

// SPPMetadata contains copied C solver and geometry metadata.
type SPPMetadata struct {
	Iterations          int
	Converged           bool
	Status              uint32
	IonosphereApplied   bool
	TroposphereApplied  bool
	OuterIterations     int
	HasFinalRobustScale bool
	FinalRobustScaleM   float64
	UsedCount           int
	SystemCount         int
	Redundancy          int64
	RAIMCheckable       bool
	GeometryQuality     SPPGeometryQuality
}

// SPPSolution owns only Go memory. It contains no live C solution handle and
// can be copied as a value after SolveSPP returns.
type SPPSolution struct {
	PositionM          [3]float64
	ReceiverClockS     float64
	UsedSatelliteCount int
	UsedSatelliteIDs   []string
	ResidualsM         []float64
	DOP                *DOP
	Geodetic           *Geodetic
	Metadata           SPPMetadata
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
		TRxJ2000S:       config.TRxJ2000S,
		TRxSecondOfDayS: config.TRxSecondOfDayS,
		DayOfYear:       config.DayOfYear,
		InitialGuess:    config.InitialGuess,
		Ionosphere:      config.Ionosphere,
		Troposphere:     config.Troposphere,
		WithGeodetic:    config.WithGeodetic,
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
		TRxJ2000S:       config.TRxJ2000S,
		TRxSecondOfDayS: config.TRxSecondOfDayS,
		DayOfYear:       config.DayOfYear,
		InitialGuess:    config.InitialGuess,
		Ionosphere:      config.Ionosphere,
		Troposphere:     config.Troposphere,
		WithGeodetic:    config.WithGeodetic,
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
