package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// TerrestrialFrame identifies one built-in ITRF/ETRF realization.
type TerrestrialFrame uint32

const (
	TerrestrialFrameITRF2020 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameITRF2020Value)
	TerrestrialFrameITRF2014 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameITRF2014Value)
	TerrestrialFrameITRF2008 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameITRF2008Value)
	TerrestrialFrameETRF2020 TerrestrialFrame = TerrestrialFrame(native.TerrestrialFrameETRF2020Value)
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

// HelmertTransform is one catalog entry. Translations are mm, rotations are
// mas, scale is ppb, and rates use the corresponding per-year units.
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

func FrameCatalogCount() (int, error) {
	value, err := native.FrameCatalogCount()
	return value, publicError(err)
}

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

func TransformTerrestrialState(state TerrestrialState, from, to TerrestrialFrame, epochYear float64) (TerrestrialState, error) {
	value, err := native.FrameCatalogTransform(nativeTerrestrialState(state), native.TerrestrialFrame(from), native.TerrestrialFrame(to), epochYear)
	return publicTerrestrialState(value), publicError(err)
}

func TransformTerrestrialFromEpoch(position TerrestrialPosition, velocity TerrestrialVelocity, positionYear float64, from, to TerrestrialFrame, transformYear float64) (TerrestrialState, error) {
	value, err := native.FrameCatalogTransformFromEpoch(nativeTerrestrialPosition(position), nativeTerrestrialVelocity(velocity), positionYear, native.TerrestrialFrame(from), native.TerrestrialFrame(to), transformYear)
	return publicTerrestrialState(value), publicError(err)
}
