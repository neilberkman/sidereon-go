//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
*/
import "C"

// qualityStatusErrorLocked must be called immediately after a quality
// producer, while the producing OS thread is still locked.
func qualityStatusErrorLocked(status uint32) error {
	err := statusErrorLocked(status)
	if err == nil {
		return nil
	}
	statusErr, ok := err.(*StatusError)
	if !ok {
		return err
	}
	kind := uint32(C.sidereon_last_quality_error_kind())
	if kind != 0 {
		statusErr.QualityKind = kind
	}
	return statusErr
}
