//go:build cgo && linux && arm64 && sidereon_linux_glibc && !sidereon_linux_musl && sidereon_use_system_lib

package native

// In system-library mode the consumer supplies CGO_LDFLAGS.
