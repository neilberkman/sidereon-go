package sidereon

import "testing"

func TestModelSelectorDefaultsAndVariants(t *testing.T) {
	defaults, err := DefaultSPPModelOptions()
	if err != nil || defaults != (SPPModelOptions{QZSSClock: QZSSClockGPS, TroposphereModel: TroposphereRTKLIB}) {
		t.Fatalf("native SPP model defaults = %+v, err=%v", defaults, err)
	}
	if QZSSClockGPS != 0 || QZSSClockSeparate != 1 {
		t.Fatalf("QZSS clock tags = %d,%d", QZSSClockGPS, QZSSClockSeparate)
	}
	if TroposphereRTKLIB != 0 || TroposphereSaastamoinenNiell != 1 {
		t.Fatalf("troposphere tags = %d,%d", TroposphereRTKLIB, TroposphereSaastamoinenNiell)
	}
	options := SPPModelOptions{QZSSClock: QZSSClockSeparate, TroposphereModel: TroposphereSaastamoinenNiell}
	if options.QZSSClock != 1 || options.TroposphereModel != 1 {
		t.Fatalf("model selectors lost: %+v", options)
	}
}

func TestModelSelectorsSurviveInputAdapters(t *testing.T) {
	models := SPPModelOptions{QZSSClock: QZSSClockSeparate, TroposphereModel: TroposphereSaastamoinenNiell}
	input, err := nativeSppV2(SPPInputsV2{Models: models})
	if err != nil {
		t.Fatalf("convert SPP V2 options: %v", err)
	}
	if input.Models.QZSSClock != uint32(models.QZSSClock) || input.Models.TroposphereModel != uint32(models.TroposphereModel) {
		t.Fatalf("SPP adapter lost selectors: %+v", input.Models)
	}
	rinex, err := nativeRinexSppOptions(RINEXSPPOptions{Models: models})
	if err != nil {
		t.Fatalf("convert RINEX options: %v", err)
	}
	if rinex.Models != input.Models {
		t.Fatalf("RINEX adapter lost selectors: %+v", rinex.Models)
	}
	static, err := nativeStaticOptions(&StaticPositionOptions{Models: models})
	if err != nil {
		t.Fatalf("convert static-position options: %v", err)
	}
	if static.Models != input.Models {
		t.Fatalf("static-position adapter lost selectors: %+v", static.Models)
	}
}
