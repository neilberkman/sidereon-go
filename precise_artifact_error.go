package sidereon

import (
	"fmt"
	"sidereon.dev/go/v3/internal/native"
)

// OpenPreciseInterpolantArtifactAttestedFile opens a native memory-mapped file
// using a trusted expected checksum. Header and layout validation remain strict;
// payload hashing is deferred until Verify. The native handle owns the mapping.
// A checksum claim that differs from the header is a typed refusal.
func OpenPreciseInterpolantArtifactAttestedFile(path string, checksum uint64) (*PreciseInterpolantArtifact, PreciseInterpolantArtifactError, error) {
	handle, kind, err := native.OpenPreciseInterpolantArtifactAttestedFile(path, checksum)
	if err != nil {
		return nil, PreciseInterpolantArtifactError(kind), publicError(err)
	}
	return &PreciseInterpolantArtifact{handle: handle}, PreciseInterpolantArtifactError(kind), nil
}

// PreciseArtifactError retains every typed and textual native artifact refusal field.
// Fields apply according to Kind; HasSatellite governs Satellite.
type PreciseArtifactError struct {
	// Kind is the artifact refusal category, retaining future numeric values.
	Kind PreciseInterpolantArtifactError
	// Version is the unsupported artifact version.
	Version uint16
	// Tag is an unsupported time-scale or satellite-system tag.
	Tag uint8
	// FoundMagic contains the eight bytes found instead of the artifact magic.
	FoundMagic [8]byte
	// Available is the available byte count for truncation, trailing, or range errors.
	Available uint64
	// Declared is the header-declared byte length for truncation or trailing errors.
	Declared uint64
	// Offset is the start of an out-of-bounds region.
	Offset uint64
	// Length is the length of an out-of-bounds region.
	Length uint64
	// ExpectedChecksum is the expected file or satellite checksum.
	ExpectedChecksum uint64
	// FoundChecksum is the computed file or satellite checksum.
	FoundChecksum uint64
	// ClaimedChecksum is the caller-attested checksum.
	ClaimedChecksum uint64
	// DeclaredChecksum is the header checksum when caller attestation differs.
	DeclaredChecksum uint64
	// HasSatellite reports whether Satellite identifies the affected satellite.
	HasSatellite bool
	// Satellite is the affected satellite token.
	Satellite string
	// Path is the path associated with an I/O failure.
	Path string
	// Message is the retained I/O or future-variant diagnostic.
	Message string
	// Reason is the retained parse reason.
	Reason string
	// Region names the out-of-bounds region.
	Region string
}

// Error returns the retained native reason or the numeric refusal category.
func (e *PreciseArtifactError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return e.Reason
	}
	return fmt.Sprintf("sidereon: precise artifact error %d", e.Kind)
}
