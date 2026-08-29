//go:build cgo && !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && ((sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#error sidereon: unsupported platform or libc selection; use a supported cgo target or sidereon_use_system_lib
*/
import "C"
