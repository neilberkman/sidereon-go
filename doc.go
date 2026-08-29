// Package sidereon is a hand-written Go binding over the complete public
// Sidereon C ABI. Numerical, parsing, validation, and protocol-state operations
// delegate to the bundled native engine; network and filesystem transport stay
// in Go and pass detached bytes across the ABI boundary.
//
// The package requires cgo, CGO_ENABLED=1, and a working C toolchain. Darwin
// arm64 and amd64, Windows amd64 with a GNU C toolchain, and Linux amd64/arm64
// are supported by the build selection files. Bundled-library builds on Linux
// require exactly one of the
// sidereon_linux_glibc or sidereon_linux_musl build tags because Go build
// constraints cannot safely identify the C library ABI. The
// sidereon_use_system_lib tag selects a system library without a bundled-libc
// tag; in that mode the consumer must provide the library through CGO_LDFLAGS.
//
// Returned value structs, strings, and slices are detached Go-owned copies.
// Opaque handle values own native resources, must not be copied after first
// use, and should be released with Close; Close is idempotent. Handle methods
// coordinate concurrent reads and native release internally, so a concurrent
// Close cannot overlap a native read. Callers must still synchronize their own
// mutation of Go slices, option structs, and other input values.
//
// Unless a function documents a narrower error, a failed native operation is
// returned as *StatusError with the stable status code and same-call native
// detail. Methods invoked after their owning handle is closed return ErrClosed,
// and native argument failures use StatusInvalidArgument. Invalid indices,
// Go-side validation, transport, and context cancellation may instead return
// ordinary Go errors. Constructors and methods returning a handle transfer
// ownership to the caller, who should defer Close after checking the error.
package sidereon
