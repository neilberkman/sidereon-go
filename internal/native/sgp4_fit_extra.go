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
	"runtime"
	"unsafe"
)

type SGP4FitEpochKind uint32

const (
	SGP4FitEpochMidpoint SGP4FitEpochKind = iota
	SGP4FitEpochFirst
	SGP4FitEpochLast
	SGP4FitEpochSample
	SGP4FitEpochJD
)

type SGP4Loss uint32

const (
	SGP4LossLinear SGP4Loss = iota
	SGP4LossSoftL1
	SGP4LossHuber
	SGP4LossCauchy
	SGP4LossArctan
)

type SGP4XScaleKind uint32

const (
	SGP4XScaleNone SGP4XScaleKind = iota
	SGP4XScaleUnit
	SGP4XScaleValues
	SGP4XScaleJacobian
)

type SGP4FitSample struct {
	JDWhole             float64
	JDFraction          float64
	PositionTEMEKm      [3]float64
	HasVelocityTEMEKmPS bool
	VelocityTEMEKmPS    [3]float64
}

type SGP4FitConfig struct {
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
	OpsMode                 uint32
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

type SGP4FitStatistics struct {
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

type SGP4TLEFit struct {
	_      noCopy
	handle *positioningHandle
}

func releaseSGP4TLEFit(pointer unsafe.Pointer) {
	C.sidereon_sgp4_tle_fit_free((*C.SidereonSgp4TleFit)(pointer))
}

func SGP4FitConfigDefaults() (SGP4FitConfig, error) {
	var out C.SidereonSgp4FitConfig
	err := callStatus(func() uint32 { return C.sidereon_sgp4_fit_config_init(&out) })
	if err != nil {
		return SGP4FitConfig{}, err
	}
	return sgp4FitConfigFromC(out)
}

func checkSGP4FitConfig(value SGP4FitConfig) error {
	if value.EpochKind > SGP4FitEpochJD {
		return invalidArgument("SGP4 fit epoch kind is not defined by the C ABI")
	}
	if value.EpochKind == SGP4FitEpochSample && value.EpochSampleIndex < 0 {
		return invalidArgument("SGP4 fit epoch sample index must not be negative")
	}
	if err := validOpsMode(value.OpsMode); err != nil {
		return err
	}
	if value.XScaleKind > SGP4XScaleJacobian {
		return invalidArgument("SGP4 fit x-scale kind is not defined by the C ABI")
	}
	if value.Loss > SGP4LossArctan {
		return invalidArgument("SGP4 fit loss is not defined by the C ABI")
	}
	if value.HasMaxNFEV && value.MaxNFEV < 0 {
		return invalidArgument("SGP4 fit maximum evaluations must not be negative")
	}
	if value.XScaleKind == SGP4XScaleValues && len(value.XScaleValues) == 0 {
		return invalidArgument("SGP4 fit value x-scale must not be empty")
	}
	for _, item := range []struct {
		value string
		name  string
	}{{value.Classification, "SGP4 fit classification"}, {value.InternationalDesignator, "SGP4 fit international designator"}, {value.ObjectName, "SGP4 fit object name"}} {
		if err := rejectEmbeddedNUL(item.value, item.name); err != nil {
			return err
		}
	}
	if len(value.Classification) >= 32 || len(value.InternationalDesignator) >= 32 || len(value.ObjectName) >= 65 {
		return invalidArgument("SGP4 fit metadata string is too long")
	}
	return nil
}

func cFixedString32(destination *[32]C.char, value string) {
	for i := range value {
		destination[i] = C.char(value[i])
	}
}

func cFixedString65(destination *[65]C.char, value string) {
	for i := range value {
		destination[i] = C.char(value[i])
	}
}

func cSGP4FitConfig(value SGP4FitConfig, weights, xScale *C.double) (C.SidereonSgp4FitConfig, error) {
	if err := checkSGP4FitConfig(value); err != nil {
		return C.SidereonSgp4FitConfig{}, err
	}
	var epochSampleIndex C.size_t
	var err error
	if value.EpochKind == SGP4FitEpochSample {
		epochSampleIndex, err = checkedNativeSize(value.EpochSampleIndex)
		if err != nil {
			return C.SidereonSgp4FitConfig{}, err
		}
	}
	var maxNFEV C.size_t
	if value.HasMaxNFEV {
		maxNFEV, err = checkedNativeSize(value.MaxNFEV)
		if err != nil {
			return C.SidereonSgp4FitConfig{}, err
		}
	}
	var out C.SidereonSgp4FitConfig
	out.epoch_kind = C.uint32_t(value.EpochKind)
	out.epoch_sample_index = epochSampleIndex
	out.epoch_jd_whole = C.double(value.EpochJDWhole)
	out.epoch_jd_fraction = C.double(value.EpochJDFraction)
	out.fit_bstar = C.bool(value.FitBStar)
	out.bstar_seed = C.double(value.BStarSeed)
	out.use_velocity = C.bool(value.UseVelocity)
	out.has_velocity_weight_s = C.bool(value.HasVelocityWeightS)
	out.velocity_weight_s = C.double(value.VelocityWeightS)
	out.weights = weights
	out.weight_count = C.size_t(len(value.Weights))
	out.opsmode = C.uint32_t(value.OpsMode)
	out.has_ftol = C.bool(value.HasFTol)
	out.ftol = C.double(value.FTol)
	out.has_xtol = C.bool(value.HasXTol)
	out.xtol = C.double(value.XTol)
	out.has_gtol = C.bool(value.HasGTol)
	out.gtol = C.double(value.GTol)
	out.has_max_nfev = C.bool(value.HasMaxNFEV)
	out.max_nfev = maxNFEV
	out.x_scale_kind = C.uint32_t(value.XScaleKind)
	out.x_scale_values = xScale
	out.x_scale_value_count = C.size_t(len(value.XScaleValues))
	out.loss = C.uint32_t(value.Loss)
	out.f_scale = C.double(value.FScale)
	out.catalog_number = C.uint32_t(value.CatalogNumber)
	cFixedString32(&out.classification, value.Classification)
	cFixedString32(&out.international_designator, value.InternationalDesignator)
	out.element_set_number = C.int32_t(value.ElementSetNumber)
	out.rev_at_epoch = C.int64_t(value.RevAtEpoch)
	cFixedString65(&out.object_name, value.ObjectName)
	return out, nil
}

func sgp4FitConfigFromC(value C.SidereonSgp4FitConfig) (SGP4FitConfig, error) {
	epochSampleIndex, err := checkedNativeCount(uint64(value.epoch_sample_index))
	if err != nil {
		return SGP4FitConfig{}, err
	}
	maxNFEV, err := checkedNativeCount(uint64(value.max_nfev))
	if err != nil {
		return SGP4FitConfig{}, err
	}
	return SGP4FitConfig{EpochKind: SGP4FitEpochKind(value.epoch_kind), EpochSampleIndex: epochSampleIndex, EpochJDWhole: float64(value.epoch_jd_whole), EpochJDFraction: float64(value.epoch_jd_fraction), FitBStar: bool(value.fit_bstar), BStarSeed: float64(value.bstar_seed), UseVelocity: bool(value.use_velocity), HasVelocityWeightS: bool(value.has_velocity_weight_s), VelocityWeightS: float64(value.velocity_weight_s), OpsMode: uint32(value.opsmode), HasFTol: bool(value.has_ftol), FTol: float64(value.ftol), HasXTol: bool(value.has_xtol), XTol: float64(value.xtol), HasGTol: bool(value.has_gtol), GTol: float64(value.gtol), HasMaxNFEV: bool(value.has_max_nfev), MaxNFEV: maxNFEV, XScaleKind: SGP4XScaleKind(value.x_scale_kind), Loss: SGP4Loss(value.loss), FScale: float64(value.f_scale), CatalogNumber: uint32(value.catalog_number), Classification: cFixedStringToGo32(value.classification), InternationalDesignator: cFixedStringToGo32(value.international_designator), ElementSetNumber: int32(value.element_set_number), RevAtEpoch: int64(value.rev_at_epoch), ObjectName: cFixedStringToGo65(value.object_name)}, nil
}

func cFixedStringToGo32(value [32]C.char) string {
	length := len(value)
	for i := range value {
		if value[i] == 0 {
			length = i
			break
		}
	}
	result := make([]byte, length)
	for i := range result {
		result[i] = byte(value[i])
	}
	return string(result)
}

func cFixedStringToGo65(value [65]C.char) string {
	length := len(value)
	for i := range value {
		if value[i] == 0 {
			length = i
			break
		}
	}
	result := make([]byte, length)
	for i := range result {
		result[i] = byte(value[i])
	}
	return string(result)
}

func cSGP4FitSample(value SGP4FitSample) C.SidereonSgp4FitSample {
	var out C.SidereonSgp4FitSample
	out.jd_whole = C.double(value.JDWhole)
	out.jd_fraction = C.double(value.JDFraction)
	for i := range out.position_teme_km {
		out.position_teme_km[i] = C.double(value.PositionTEMEKm[i])
		out.velocity_teme_km_s[i] = C.double(value.VelocityTEMEKmPS[i])
	}
	out.has_velocity_teme_km_s = C.bool(value.HasVelocityTEMEKmPS)
	return out
}

func FitSGP4TLE(samples []SGP4FitSample, config SGP4FitConfig) (*SGP4TLEFit, error) {
	if len(samples) == 0 {
		return nil, invalidArgument("SGP4 fit samples must not be empty")
	}
	if err := checkSGP4FitConfig(config); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(samples), unsafe.Sizeof(C.SidereonSgp4FitSample{})); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(config.Weights), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(len(config.XScaleValues), unsafe.Sizeof(C.double(0))); err != nil {
		return nil, err
	}
	cSamples := make([]C.SidereonSgp4FitSample, len(samples))
	for i, sample := range append([]SGP4FitSample(nil), samples...) {
		cSamples[i] = cSGP4FitSample(sample)
	}
	var weightsMemory, xScaleMemory unsafe.Pointer
	if len(config.Weights) != 0 {
		weightsMemory = C.malloc(C.size_t(len(config.Weights)) * C.size_t(unsafe.Sizeof(C.double(0))))
		if weightsMemory == nil {
			return nil, errors.New("sidereon: unable to allocate SGP4 fit weights")
		}
		defer C.free(weightsMemory)
		weights := unsafe.Slice((*C.double)(weightsMemory), len(config.Weights))
		for i, value := range append([]float64(nil), config.Weights...) {
			weights[i] = C.double(value)
		}
	}
	if len(config.XScaleValues) != 0 {
		xScaleMemory = C.malloc(C.size_t(len(config.XScaleValues)) * C.size_t(unsafe.Sizeof(C.double(0))))
		if xScaleMemory == nil {
			return nil, errors.New("sidereon: unable to allocate SGP4 fit x-scale")
		}
		defer C.free(xScaleMemory)
		xScale := unsafe.Slice((*C.double)(xScaleMemory), len(config.XScaleValues))
		for i, value := range append([]float64(nil), config.XScaleValues...) {
			xScale[i] = C.double(value)
		}
	}
	cConfig, err := cSGP4FitConfig(config, (*C.double)(weightsMemory), (*C.double)(xScaleMemory))
	if err != nil {
		return nil, err
	}
	var out *C.SidereonSgp4TleFit
	err = callStatus(func() uint32 {
		return C.sidereon_sgp4_fit_tle(&cSamples[0], C.size_t(len(cSamples)), &cConfig, &out)
	})
	runtime.KeepAlive(cSamples)
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_sgp4_tle_fit_free(out) })
		}
		return nil, err
	}
	if out == nil {
		return nil, missingNativeHandle("SGP4 fit")
	}
	return &SGP4TLEFit{handle: newPositioningHandle(unsafe.Pointer(out), releaseSGP4TLEFit)}, nil
}

func (fit *SGP4TLEFit) Close() error {
	if fit == nil || fit.handle == nil {
		return nil
	}
	return fit.handle.close()
}

func (fit *SGP4TLEFit) Lines() (TLELines, error) {
	if fit == nil || fit.handle == nil {
		return TLELines{}, ErrClosed
	}
	var out C.SidereonTleLines
	err := fit.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_sgp4_tle_fit_lines((*C.SidereonSgp4TleFit)(pointer), &out) })
	})
	if err != nil {
		return TLELines{}, err
	}
	return TLELines{Line1: cTLEString((*C.char)(unsafe.Pointer(&out.line1.bytes[0]))), Line2: cTLEString((*C.char)(unsafe.Pointer(&out.line2.bytes[0])))}, nil
}

func (fit *SGP4TLEFit) OMM() (*OMM, error) {
	if fit == nil || fit.handle == nil {
		return nil, ErrClosed
	}
	var out *C.SidereonOmm
	err := fit.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_sgp4_tle_fit_omm((*C.SidereonSgp4TleFit)(pointer), &out) })
	})
	if err != nil {
		if out != nil {
			withCThread(func() { C.sidereon_omm_free(out) })
		}
		return nil, err
	}
	return newOMM(out)
}

func (fit *SGP4TLEFit) Statistics() (SGP4FitStatistics, error) {
	if fit == nil || fit.handle == nil {
		return SGP4FitStatistics{}, ErrClosed
	}
	var value C.SidereonSgp4FitStatistics
	err := fit.handle.with(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 { return C.sidereon_sgp4_tle_fit_statistics((*C.SidereonSgp4TleFit)(pointer), &value) })
	})
	if err != nil {
		return SGP4FitStatistics{}, err
	}
	var axes [3]float64
	for i := range axes {
		axes[i] = float64(value.rms_position_axes_km[i])
	}
	nfev, err := checkedNativeCount(uint64(value.nfev))
	if err != nil {
		return SGP4FitStatistics{}, err
	}
	njev, err := checkedNativeCount(uint64(value.njev))
	if err != nil {
		return SGP4FitStatistics{}, err
	}
	passes, err := checkedNativeCount(uint64(value.seed_refine_passes))
	if err != nil {
		return SGP4FitStatistics{}, err
	}
	return SGP4FitStatistics{RMSPositionKm: float64(value.rms_position_km), MaxPositionKm: float64(value.max_position_km), RMSPositionAxesKm: axes, HasRMSVelocityKmPS: bool(value.has_rms_velocity_km_s), RMSVelocityKmPS: float64(value.rms_velocity_km_s), TLERMSPositionKm: float64(value.tle_rms_position_km), Status: int32(value.status), NFEV: nfev, NJEV: njev, Cost: float64(value.cost), Optimality: float64(value.optimality), BStarObservable: bool(value.bstar_observable), SeedRefinePasses: passes}, nil
}
