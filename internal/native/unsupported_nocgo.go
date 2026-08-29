//go:build !cgo

package native

// The undefined symbol intentionally makes CGO-disabled builds fail with an
// actionable diagnostic at compile time.
var _ = sidereon_requires_cgo
