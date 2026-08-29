package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SGP4FitEpochKind selects how the fitted TLE epoch is chosen.
type SGP4FitEpochKind uint32

const (
	// SGP4FitEpochMidpoint uses the sample-span midpoint.
	SGP4FitEpochMidpoint SGP4FitEpochKind = iota
	// SGP4FitEpochFirst uses the first sample epoch.
	SGP4FitEpochFirst
	// SGP4FitEpochLast uses the last sample epoch.
	SGP4FitEpochLast
	// SGP4FitEpochSample uses EpochSampleIndex.
	SGP4FitEpochSample
	// SGP4FitEpochJD uses EpochJDWhole and EpochJDFraction.
	SGP4FitEpochJD
)

// SGP4Loss selects the robust loss for TLE fitting.
type SGP4Loss uint32

const (
	// SGP4LossLinear uses a linear least-squares loss.
	SGP4LossLinear SGP4Loss = iota
	// SGP4LossSoftL1 uses a soft-L1 loss.
	SGP4LossSoftL1
	// SGP4LossHuber uses a Huber loss.
	SGP4LossHuber
	// SGP4LossCauchy uses a Cauchy loss.
	SGP4LossCauchy
	// SGP4LossArctan uses an arctangent loss.
	SGP4LossArctan
)

// SGP4XScaleKind selects the parameter scaling policy.
type SGP4XScaleKind uint32

const (
	// SGP4XScaleNone disables parameter scaling.
	SGP4XScaleNone SGP4XScaleKind = iota
	// SGP4XScaleUnit uses unit parameter scales.
	SGP4XScaleUnit
	// SGP4XScaleValues uses XScaleValues.
	SGP4XScaleValues
	// SGP4XScaleJacobian derives scales from the Jacobian.
	SGP4XScaleJacobian
)

// SGP4FitSample is one TEME position and optional velocity observation.
type SGP4FitSample struct {
	// JDWhole and JDFraction are the split TEME observation Julian date.
	JDWhole             float64
	JDFraction          float64
	PositionTEMEKm      [3]float64
	HasVelocityTEMEKmPS bool
	VelocityTEMEKmPS    [3]float64
}

// SGP4FitConfig controls TLE fitting, tolerances, weights, and metadata.
type SGP4FitConfig struct {
	// EpochKind selects the epoch source; sample/JD fields are used as applicable.
	EpochKind               SGP4FitEpochKind
	EpochSampleIndex        int
	EpochJDWhole            float64
	EpochJDFraction         float64
	FitBStar                bool
	BStarSeed               float64
	UseVelocity             bool
	HasVelocityWeightS      bool
	VelocityWeightS         float64
	Weights                 []float64
	OpsMode                 OpsMode
	HasFTol                 bool
	FTol                    float64
	HasXTol                 bool
	XTol                    float64
	HasGTol                 bool
	GTol                    float64
	HasMaxNFEV              bool
	MaxNFEV                 int
	XScaleKind              SGP4XScaleKind
	XScaleValues            []float64
	Loss                    SGP4Loss
	FScale                  float64
	CatalogNumber           uint32
	Classification          string
	InternationalDesignator string
	ElementSetNumber        int32
	RevAtEpoch              int64
	ObjectName              string
}

// SGP4FitStatistics contains detached fitting residual and optimizer results.
type SGP4FitStatistics struct {
	// RMSPositionKm and MaxPositionKm are position residuals in km.
	RMSPositionKm      float64
	MaxPositionKm      float64
	RMSPositionAxesKm  [3]float64
	HasRMSVelocityKmPS bool
	RMSVelocityKmPS    float64
	TLERMSPositionKm   float64
	Status             int32
	NFEV               int
	NJEV               int
	Cost               float64
	Optimality         float64
	BStarObservable    bool
	SeedRefinePasses   int
}

// SGP4TLEFit owns a fitted native TLE result.
type SGP4TLEFit struct {
	_      noCopy
	handle *native.SGP4TLEFit
}

func nativeSGP4FitSample(value SGP4FitSample) native.SGP4FitSample {
	return native.SGP4FitSample{JDWhole: value.JDWhole, JDFraction: value.JDFraction, PositionTEMEKm: value.PositionTEMEKm, HasVelocityTEMEKmPS: value.HasVelocityTEMEKmPS, VelocityTEMEKmPS: value.VelocityTEMEKmPS}
}

func nativeSGP4FitConfig(value SGP4FitConfig) native.SGP4FitConfig {
	return native.SGP4FitConfig{EpochKind: native.SGP4FitEpochKind(value.EpochKind), EpochSampleIndex: value.EpochSampleIndex, EpochJDWhole: value.EpochJDWhole, EpochJDFraction: value.EpochJDFraction, FitBStar: value.FitBStar, BStarSeed: value.BStarSeed, UseVelocity: value.UseVelocity, HasVelocityWeightS: value.HasVelocityWeightS, VelocityWeightS: value.VelocityWeightS, Weights: append([]float64(nil), value.Weights...), OpsMode: uint32(value.OpsMode), HasFTol: value.HasFTol, FTol: value.FTol, HasXTol: value.HasXTol, XTol: value.XTol, HasGTol: value.HasGTol, GTol: value.GTol, HasMaxNFEV: value.HasMaxNFEV, MaxNFEV: value.MaxNFEV, XScaleKind: native.SGP4XScaleKind(value.XScaleKind), XScaleValues: append([]float64(nil), value.XScaleValues...), Loss: native.SGP4Loss(value.Loss), FScale: value.FScale, CatalogNumber: value.CatalogNumber, Classification: value.Classification, InternationalDesignator: value.InternationalDesignator, ElementSetNumber: value.ElementSetNumber, RevAtEpoch: value.RevAtEpoch, ObjectName: value.ObjectName}
}

func publicSGP4FitConfig(value native.SGP4FitConfig) SGP4FitConfig {
	return SGP4FitConfig{EpochKind: SGP4FitEpochKind(value.EpochKind), EpochSampleIndex: value.EpochSampleIndex, EpochJDWhole: value.EpochJDWhole, EpochJDFraction: value.EpochJDFraction, FitBStar: value.FitBStar, BStarSeed: value.BStarSeed, UseVelocity: value.UseVelocity, HasVelocityWeightS: value.HasVelocityWeightS, VelocityWeightS: value.VelocityWeightS, Weights: append([]float64(nil), value.Weights...), OpsMode: OpsMode(value.OpsMode), HasFTol: value.HasFTol, FTol: value.FTol, HasXTol: value.HasXTol, XTol: value.XTol, HasGTol: value.HasGTol, GTol: value.GTol, HasMaxNFEV: value.HasMaxNFEV, MaxNFEV: value.MaxNFEV, XScaleKind: SGP4XScaleKind(value.XScaleKind), XScaleValues: append([]float64(nil), value.XScaleValues...), Loss: SGP4Loss(value.Loss), FScale: value.FScale, CatalogNumber: value.CatalogNumber, Classification: value.Classification, InternationalDesignator: value.InternationalDesignator, ElementSetNumber: value.ElementSetNumber, RevAtEpoch: value.RevAtEpoch, ObjectName: value.ObjectName}
}

// SGP4FitConfigDefaults returns native SGP4 fitting defaults.
func SGP4FitConfigDefaults() (SGP4FitConfig, error) {
	value, err := native.SGP4FitConfigDefaults()
	return publicSGP4FitConfig(value), publicError(err)
}

// FitSGP4TLE fits a TLE to detached TEME observations.
func FitSGP4TLE(samples []SGP4FitSample, config SGP4FitConfig) (*SGP4TLEFit, error) {
	values := make([]native.SGP4FitSample, len(samples))
	for i, value := range append([]SGP4FitSample(nil), samples...) {
		values[i] = nativeSGP4FitSample(value)
	}
	value, err := native.FitSGP4TLE(values, nativeSGP4FitConfig(config))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &SGP4TLEFit{handle: value}, nil
}

// Close releases the fitted TLE and is idempotent.
func (fit *SGP4TLEFit) Close() error {
	if fit == nil || fit.handle == nil {
		return nil
	}
	return publicError(fit.handle.Close())
}

// Lines returns the fitted two-line element text.
func (fit *SGP4TLEFit) Lines() (TLELines, error) {
	if fit == nil || fit.handle == nil {
		return TLELines{}, ErrClosed
	}
	value, err := fit.handle.Lines()
	return TLELines{Line1: value.Line1, Line2: value.Line2}, publicError(err)
}

// OMM returns the fitted detached CCSDS orbit mean-elements record.
func (fit *SGP4TLEFit) OMM() (*OMM, error) {
	if fit == nil || fit.handle == nil {
		return nil, ErrClosed
	}
	value, err := fit.handle.OMM()
	if err != nil {
		return nil, publicError(err)
	}
	return publicOMM(value), nil
}

// Statistics returns detached fitting residual and optimizer statistics.
func (fit *SGP4TLEFit) Statistics() (SGP4FitStatistics, error) {
	if fit == nil || fit.handle == nil {
		return SGP4FitStatistics{}, ErrClosed
	}
	value, err := fit.handle.Statistics()
	return SGP4FitStatistics{RMSPositionKm: value.RMSPositionKm, MaxPositionKm: value.MaxPositionKm, RMSPositionAxesKm: value.RMSPositionAxesKm, HasRMSVelocityKmPS: value.HasRMSVelocityKmPS, RMSVelocityKmPS: value.RMSVelocityKmPS, TLERMSPositionKm: value.TLERMSPositionKm, Status: value.Status, NFEV: value.NFEV, NJEV: value.NJEV, Cost: value.Cost, Optimality: value.Optimality, BStarObservable: value.BStarObservable, SeedRefinePasses: value.SeedRefinePasses}, publicError(err)
}

// PositionTEMEKm is a TEME position in kilometres.
// HasVelocityTEMEKmPS controls the optional velocity vector in km/s.
// EpochSampleIndex selects a sample when EpochKind is SGP4FitEpochSample.
// EpochJDWhole and EpochJDFraction select a split Julian date when requested.
// FitBStar and BStarSeed control the optional B* fit.
// UseVelocity and VelocityWeightS control velocity residuals and weighting.
// Weights is an optional copied per-sample weight vector.
// OpsMode selects the native SGP4 operation mode.
// HasFTol, HasXTol, HasGTol, and HasMaxNFEV guard optimizer limits.
// XScaleValues is used when XScaleKind selects explicit values.
// Loss and FScale select the robust objective and scale.
// CatalogNumber, Classification, and InternationalDesignator are TLE metadata.
// RMSPositionAxesKm contains per-axis TEME residual RMS values in km.
// HasRMSVelocityKmPS controls RMSVelocityKmPS.
// TLERMSPositionKm is the TLE position residual RMS in km.
// Status is the native optimizer status code.
// NFEV and NJEV are native function/Jacobian evaluation counts.
// Cost and Optimality are native optimizer diagnostics.
// BStarObservable reports whether B* was identifiable; SeedRefinePasses is a count.
