package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// TerrestrialFrame identifies one built-in ITRF/ETRF realization.
type TerrestrialFrame uint32

const (
	// TerrestrialFrameITRF2020 identifies the ITRF 2020 terrestrial frame realization.
	TerrestrialFrameITRF2020 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameITRF2020Value)
	// TerrestrialFrameITRF2014 identifies the ITRF 2014 terrestrial frame realization.
	TerrestrialFrameITRF2014 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameITRF2014Value)
	// TerrestrialFrameITRF2008 identifies the ITRF 2008 terrestrial frame realization.
	TerrestrialFrameITRF2008 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameITRF2008Value)
	// TerrestrialFrameETRF2020 identifies the ETRF 2020 terrestrial frame realization.
	TerrestrialFrameETRF2020 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameETRF2020Value)
)

// HelmertParameters contains Helmert translations in millimetres, scale in ppb, and rotations in milliarcseconds.
type HelmertParameters struct {
	// TranslationMm is the translation mm in millimetres.
	TranslationMm [3]float64
	// ScalePPB is the scale ppb in parts per billion.
	ScalePPB float64
	// RotationMAS is the rotation mas in milliarcseconds.
	RotationMAS [3]float64
}

// HelmertRates contains the corresponding Helmert rates per year.
type HelmertRates struct {
	// TranslationMmPerYear is the translation mm per year in millimetres per year.
	TranslationMmPerYear [3]float64
	// ScalePPBPerYear is the Helmert scale rate in parts per billion per year.
	ScalePPBPerYear float64
	// RotationMASPerYear contains the X, Y, and Z rotation rates in milliarcseconds per year.
	RotationMASPerYear [3]float64
}

// HelmertTransform is one catalog entry. Translations are mm, rotations are
// mas, scale is ppb, and rates use the corresponding per-year units.
type HelmertTransform struct {
	// From is the source terrestrial frame realization.
	From TerrestrialFrame
	// To is the target terrestrial frame realization.
	To TerrestrialFrame
	// ReferenceYear is the decimal epoch at which Parameters apply.
	ReferenceYear float64
	// Parameters contains the translation, rotation, and scale offsets at ReferenceYear.
	Parameters HelmertParameters
	// Rates contains the per-year translation, rotation, and scale changes.
	Rates HelmertRates
	// Provenance identifies the catalog source or convention for this transform.
	Provenance string
}

// TerrestrialPosition contains a terrestrial position vector in metres.
type TerrestrialPosition struct {
	// PositionM is the terrestrial position in metres.
	PositionM [3]float64
}

// TerrestrialVelocity contains a terrestrial velocity vector in metres per year.
type TerrestrialVelocity struct {
	// VelocityMPerYear is the terrestrial velocity in metres per year.
	VelocityMPerYear [3]float64
}

// TerrestrialState contains a terrestrial position and optional velocity in the catalog frame.
type TerrestrialState struct {
	// Position is the position value in the containing frame.
	Position TerrestrialPosition
	// HasVelocity reports whether the has velocity field is present.
	HasVelocity bool
	// Velocity is the velocity value in the containing frame.
	Velocity TerrestrialVelocity
}

func publicHelmert(value native.HelmertTransform) HelmertTransform {
	return HelmertTransform{From: TerrestrialFrame(value.From), To: TerrestrialFrame(value.To), ReferenceYear: value.ReferenceYear, Parameters: HelmertParameters{TranslationMm: value.Parameters.TranslationMm, ScalePPB: value.Parameters.ScalePPB, RotationMAS: value.Parameters.RotationMAS}, Rates: HelmertRates{TranslationMmPerYear: value.Rates.TranslationMmPerYear, ScalePPBPerYear: value.Rates.ScalePPBPerYear, RotationMASPerYear: value.Rates.RotationMASPerYear}, Provenance: value.Provenance}
}

func nativeTerrestrialPosition(value TerrestrialPosition) native.TerrestrialPosition {
	return native.TerrestrialPosition{PositionM: value.PositionM}
}
func publicTerrestrialPosition(value native.TerrestrialPosition) TerrestrialPosition {
	return TerrestrialPosition{PositionM: value.PositionM}
}
func nativeTerrestrialVelocity(value TerrestrialVelocity) native.TerrestrialVelocity {
	return native.TerrestrialVelocity{VelocityMPerYear: value.VelocityMPerYear}
}
func nativeTerrestrialState(value TerrestrialState) native.TerrestrialState {
	return native.TerrestrialState{Position: nativeTerrestrialPosition(value.Position), HasVelocity: value.HasVelocity, Velocity: nativeTerrestrialVelocity(value.Velocity)}
}
func publicTerrestrialState(value native.TerrestrialState) TerrestrialState {
	return TerrestrialState{Position: publicTerrestrialPosition(value.Position), HasVelocity: value.HasVelocity, Velocity: TerrestrialVelocity{VelocityMPerYear: value.Velocity.VelocityMPerYear}}
}

// FrameCatalog returns a copied snapshot of the built-in catalog.
func FrameCatalog() ([]HelmertTransform, error) {
	values, err := native.FrameCatalogEntries()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]HelmertTransform, len(values))
	for i := range out {
		out[i] = publicHelmert(values[i])
	}
	return out, nil
}

// FrameCatalogCount returns the number of built-in terrestrial frame transforms.
func FrameCatalogCount() (int, error) {
	value, err := native.FrameCatalogCount()
	return value, publicError(err)
}

// FrameCatalogEntry returns one terrestrial Helmert transform.
func FrameCatalogEntry(from, to TerrestrialFrame) (HelmertTransform, error) {
	value, err := native.FrameCatalogEntry(native.TerrestrialFrame(from), native.TerrestrialFrame(to))
	return publicHelmert(value), publicError(err)
}

// PropagateTerrestrialPosition transports a station position between decimal
// years, using its velocity in m/year.
func PropagateTerrestrialPosition(position TerrestrialPosition, velocity TerrestrialVelocity, fromYear, toYear float64) (TerrestrialPosition, error) {
	value, err := native.FrameCatalogPropagatePosition(nativeTerrestrialPosition(position), nativeTerrestrialVelocity(velocity), fromYear, toYear)
	return publicTerrestrialPosition(value), publicError(err)
}

// TransformTerrestrialState returns a terrestrial state transformed into the target frame.
func TransformTerrestrialState(state TerrestrialState, from, to TerrestrialFrame, epochYear float64) (TerrestrialState, error) {
	value, err := native.FrameCatalogTransform(nativeTerrestrialState(state), native.TerrestrialFrame(from), native.TerrestrialFrame(to), epochYear)
	return publicTerrestrialState(value), publicError(err)
}

// TransformTerrestrialFromEpoch returns a terrestrial state transformed between frames and epochs.
func TransformTerrestrialFromEpoch(position TerrestrialPosition, velocity TerrestrialVelocity, positionYear float64, from, to TerrestrialFrame, transformYear float64) (TerrestrialState, error) {
	value, err := native.FrameCatalogTransformFromEpoch(nativeTerrestrialPosition(position), nativeTerrestrialVelocity(velocity), positionYear, native.TerrestrialFrame(from), native.TerrestrialFrame(to), transformYear)
	return publicTerrestrialState(value), publicError(err)
}
