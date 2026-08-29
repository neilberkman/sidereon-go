// Package sidereon is an idiomatic Go binding over the public Sidereon C ABI.
//
// The package requires cgo. Darwin arm64 and amd64, Windows amd64 with a GNU
// C toolchain, and Linux amd64/arm64 are supported by the build selection
// files. Bundled-library builds on Linux require exactly one of the
// sidereon_linux_glibc or sidereon_linux_musl build tags because Go build
// constraints cannot safely identify the C library ABI. The
// sidereon_use_system_lib tag selects a system library without a bundled-libc
// tag; in that mode the consumer must provide the library through CGO_LDFLAGS.
//
// The public surface is intentionally representative rather than a complete
// translation of the C header. Value APIs cover calendar/time scales,
// coordinates and geodesics, carrier combinations, frame/Doppler helpers,
// covariance metrics, innovation statistics, and eclipse geometry. Values
// returned by the API are copied into Go-owned memory. Opaque handles may be
// shared for read-only operations; Close and any future mutating operations
// must be serialized by the handle.
package sidereon
