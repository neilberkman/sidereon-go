//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

func SPPModelOptionsInit() (NativeSPPModelOptions, error) {
	var value C.SidereonSppModelOptions
	if err := callStatus(func() uint32 { return uint32(C.sidereon_spp_model_options_init(&value)) }); err != nil {
		return NativeSPPModelOptions{}, err
	}
	return NativeSPPModelOptions{QZSSClock: uint32(value.qzss_clock), TroposphereModel: uint32(value.troposphere_model)}, nil
}
