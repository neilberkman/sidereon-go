package sidereon

import (
	"bytes"
	"encoding/json"

	"sidereon.dev/go/v3/internal/native"
)

// SGP4FitResult is the complete detached typed fit result. Raw retains the
// exact complete payload for consumers that need its original JSON encoding.
type SGP4FitResult struct {
	Elements SGP4FitElements         `json:"elements"`
	Line1    string                  `json:"line1"`
	Line2    string                  `json:"line2"`
	OMM      SGP4FitOMM              `json:"omm"`
	Stats    SGP4FitResultStatistics `json:"stats"`
	Raw      json.RawMessage         `json:"-"`
}

// SGP4FitElements contains every serialized field of the fitted element set.
type SGP4FitElements struct {
	Epoch                [2]EngineFloat `json:"epoch"`
	BStar                EngineFloat    `json:"bstar"`
	MeanMotionDot        *EngineFloat   `json:"mean_motion_dot"`
	MeanMotionDoubleDot  *EngineFloat   `json:"mean_motion_double_dot"`
	Eccentricity         EngineFloat    `json:"eccentricity"`
	ArgumentOfPerigeeDeg EngineFloat    `json:"argument_of_perigee_deg"`
	InclinationDeg       EngineFloat    `json:"inclination_deg"`
	MeanAnomalyDeg       EngineFloat    `json:"mean_anomaly_deg"`
	MeanMotionRevPerDay  EngineFloat    `json:"mean_motion_rev_per_day"`
	RightAscensionDeg    EngineFloat    `json:"right_ascension_deg"`
	CatalogNumber        *json.Number   `json:"catalog_number"`
	OMMEpochDays         *EngineFloat   `json:"omm_epoch_days"`
}

// SGP4FitOMM is the complete fitted CCSDS mean-elements record.
type SGP4FitOMM struct {
	CCSDSOMMVers             *string                 `json:"ccsds_omm_vers"`
	Classification           *string                 `json:"classification"`
	CreationDate             *string                 `json:"creation_date"`
	Originator               *string                 `json:"originator"`
	MessageID                *string                 `json:"message_id"`
	ObjectName               *string                 `json:"object_name"`
	ObjectID                 *string                 `json:"object_id"`
	CenterName               *string                 `json:"center_name"`
	RefFrame                 *string                 `json:"ref_frame"`
	RefFrameEpoch            *string                 `json:"ref_frame_epoch"`
	TimeSystem               *string                 `json:"time_system"`
	MeanElementTheory        *string                 `json:"mean_element_theory"`
	Epoch                    SGP4FitOMMEpoch         `json:"epoch"`
	MeanMotion               *EngineFloat            `json:"mean_motion"`
	SemiMajorAxisKm          *EngineFloat            `json:"semi_major_axis_km"`
	Eccentricity             EngineFloat             `json:"eccentricity"`
	InclinationDeg           EngineFloat             `json:"inclination_deg"`
	RAOfAscNodeDeg           EngineFloat             `json:"ra_of_asc_node_deg"`
	ArgOfPericenterDeg       EngineFloat             `json:"arg_of_pericenter_deg"`
	MeanAnomalyDeg           EngineFloat             `json:"mean_anomaly_deg"`
	GMKm3S2                  *EngineFloat            `json:"gm_km3_s2"`
	Spacecraft               *SGP4FitOMMSpacecraft   `json:"spacecraft"`
	EphemerisType            *json.Number            `json:"ephemeris_type"`
	ClassificationType       *string                 `json:"classification_type"`
	NORADCatID               *json.Number            `json:"norad_cat_id"`
	ElementSetNo             *json.Number            `json:"element_set_no"`
	RevAtEpoch               *json.Number            `json:"rev_at_epoch"`
	BStar                    *EngineFloat            `json:"bstar"`
	BTermM2Kg                *EngineFloat            `json:"bterm_m2_kg"`
	MeanMotionDot            *EngineFloat            `json:"mean_motion_dot"`
	MeanMotionDDot           *EngineFloat            `json:"mean_motion_ddot"`
	AGOMM2Kg                 *EngineFloat            `json:"agom_m2_kg"`
	Covariance               *SGP4FitOMMCovariance   `json:"covariance"`
	UserDefined              []SGP4FitOMMUserDefined `json:"user_defined"`
	Comments                 SGP4FitOMMComments      `json:"comments"`
	ExactSGP4Epoch           *[2]EngineFloat         `json:"exact_sgp4_epoch"`
	QuantizeTLEDerivedFields bool                    `json:"quantize_tle_derived_fields"`
}

// SGP4FitOMMEpoch retains exact civil components, including sub-microseconds.
type SGP4FitOMMEpoch struct {
	Year        json.Number `json:"year"`
	Month       json.Number `json:"month"`
	Day         json.Number `json:"day"`
	Hour        json.Number `json:"hour"`
	Minute      json.Number `json:"minute"`
	Second      json.Number `json:"second"`
	Microsecond json.Number `json:"microsecond"`
	Femtosecond json.Number `json:"femtosecond"`
}

// SGP4FitOMMSpacecraft retains every optional spacecraft field.
type SGP4FitOMMSpacecraft struct {
	Comments       []string     `json:"comments"`
	MassKg         *EngineFloat `json:"mass_kg"`
	SolarRadAreaM2 *EngineFloat `json:"solar_rad_area_m2"`
	SolarRadCoeff  *EngineFloat `json:"solar_rad_coeff"`
	DragAreaM2     *EngineFloat `json:"drag_area_m2"`
	DragCoeff      *EngineFloat `json:"drag_coeff"`
}

// SGP4FitOMMCovariance retains the source-order lower triangle.
type SGP4FitOMMCovariance struct {
	Comments      []string        `json:"comments"`
	CovRefFrame   *string         `json:"cov_ref_frame"`
	LowerTriangle [21]EngineFloat `json:"lower_triangle"`
}

// SGP4FitOMMUserDefined is one ordered OMM user parameter.
type SGP4FitOMMUserDefined struct {
	Parameter string `json:"parameter"`
	Value     string `json:"value"`
}

// SGP4FitOMMComments retains comments by OMM block and source order.
type SGP4FitOMMComments struct {
	Header        []string `json:"header"`
	Metadata      []string `json:"metadata"`
	MeanElements  []string `json:"mean_elements"`
	TLEParameters []string `json:"tle_parameters"`
	UserDefined   []string `json:"user_defined"`
}

// SGP4FitResultStatistics keeps every exact floating-point encoding and integer result.
type SGP4FitResultStatistics struct {
	RMSPositionKm     EngineFloat    `json:"rms_position_km"`
	MaxPositionKm     EngineFloat    `json:"max_position_km"`
	RMSPositionAxesKm [3]EngineFloat `json:"rms_position_axes_km"`
	RMSVelocityKmPS   *EngineFloat   `json:"rms_velocity_km_s"`
	TLERMSPositionKm  EngineFloat    `json:"tle_rms_position_km"`
	Status            json.Number    `json:"status"`
	NFEV              json.Number    `json:"nfev"`
	NJEV              json.Number    `json:"njev"`
	Cost              EngineFloat    `json:"cost"`
	Optimality        EngineFloat    `json:"optimality"`
	BStarObservable   bool           `json:"bstar_observable"`
	SeedRefinePasses  json.Number    `json:"seed_refine_passes"`
}

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

// ResultPayload returns a detached lossless JSON copy of the full fitted
// result, including fields beyond Lines, OMM, and Statistics.
func (fit *SGP4TLEFit) ResultPayload() ([]byte, error) {
	if fit == nil || fit.handle == nil {
		return nil, ErrClosed
	}
	value, err := fit.handle.ResultPayload()
	return append([]byte(nil), value...), publicError(err)
}

// Result returns the full typed top-level result and lossless nested fields.
func (fit *SGP4TLEFit) Result() (SGP4FitResult, error) {
	var result SGP4FitResult
	payload, err := fit.ResultPayload()
	if err != nil {
		return result, err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return SGP4FitResult{}, err
	}
	result.Raw = append(json.RawMessage(nil), payload...)
	return result, nil
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
