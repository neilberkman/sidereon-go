//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

func OpenPreciseInterpolantArtifactFile(path string) (*PreciseInterpolantArtifact, uint32, error) {
	return openPreciseInterpolantArtifactPath(path, false, 0)
}

func OpenPreciseInterpolantArtifactAttestedFile(path string, checksum uint64) (*PreciseInterpolantArtifact, uint32, error) {
	return openPreciseInterpolantArtifactPath(path, true, checksum)
}

func openPreciseInterpolantArtifactPath(path string, attested bool, checksum uint64) (*PreciseInterpolantArtifact, uint32, error) {
	var output *C.SidereonPreciseInterpolantArtifact
	var kind C.enum_SidereonPreciseInterpolantArtifactErrorKind
	err := withStringStatus(path, func(value *C.char) uint32 {
		if attested {
			return uint32(C.sidereon_precise_interpolant_artifact_from_path_attested(value, C.uint64_t(checksum), &kind, &output))
		}
		return uint32(C.sidereon_precise_interpolant_artifact_from_path(value, &kind, &output))
	}, statusPreciseArtifactErrorLocked)
	if err != nil {
		if output != nil {
			withCThread(func() { C.sidereon_precise_interpolant_artifact_free(output) })
		}
		return nil, uint32(kind), err
	}
	if output == nil {
		return nil, uint32(kind), missingNativeHandle("precise artifact")
	}
	return &PreciseInterpolantArtifact{handle: newPositioningHandle(unsafe.Pointer(output), releasePreciseArtifact)}, uint32(kind), nil
}

func standaloneTerrainTextLocked(family, part C.uint32_t) (string, error) {
	text, err := copyNativeBytesLocked("terrain error text", func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_last_terrain_error_text(family, part, out, n, written, required)
	})
	return string(text), err
}

func statusDtedTileErrorLocked(status uint32) error {
	err := statusErrorLocked(status)
	if err == nil {
		return nil
	}
	var primary *StatusError
	if !errors.As(err, &primary) {
		return err
	}
	var raw C.SidereonDtedTileError
	if detailErr := statusErrorLocked(uint32(C.sidereon_last_dted_tile_error(&raw))); detailErr != nil {
		return errors.Join(err, detailErr)
	}
	if uint32(raw.kind) == 0 {
		return err
	}
	value := terrainTileError(raw)
	var detailErr error
	value.Path, detailErr = standaloneTerrainTextLocked(C.SIDEREON_TERRAIN_ERROR_FAMILY_DTED_TILE, C.SIDEREON_TERRAIN_ERROR_TEXT_PATH)
	if detailErr != nil {
		return errors.Join(err, detailErr)
	}
	value.Message, detailErr = standaloneTerrainTextLocked(C.SIDEREON_TERRAIN_ERROR_FAMILY_DTED_TILE, C.SIDEREON_TERRAIN_ERROR_TEXT_MESSAGE)
	if detailErr != nil {
		return errors.Join(err, detailErr)
	}
	primary.DtedTile = &value
	return err
}

func statusGeoidErrorLocked(status uint32) error {
	err := statusErrorLocked(status)
	if err == nil {
		return nil
	}
	var primary *StatusError
	if !errors.As(err, &primary) {
		return err
	}
	var raw C.SidereonGeoidError
	if detailErr := statusErrorLocked(uint32(C.sidereon_last_geoid_error(&raw))); detailErr != nil {
		return errors.Join(err, detailErr)
	}
	if uint32(raw.kind) == 0 {
		return err
	}
	value := GeoidError{Kind: uint32(raw.kind), Expected: uint64(raw.expected), Found: uint64(raw.found), Index: uint64(raw.index)}
	var detailErr error
	value.Field, detailErr = standaloneTerrainTextLocked(C.SIDEREON_TERRAIN_ERROR_FAMILY_GEOID, C.SIDEREON_TERRAIN_ERROR_TEXT_FIELD)
	if detailErr != nil {
		return errors.Join(err, detailErr)
	}
	value.Reason, detailErr = standaloneTerrainTextLocked(C.SIDEREON_TERRAIN_ERROR_FAMILY_GEOID, C.SIDEREON_TERRAIN_ERROR_TEXT_REASON)
	if detailErr != nil {
		return errors.Join(err, detailErr)
	}
	primary.Geoid = &value
	return err
}

func statusPreciseArtifactErrorLocked(status uint32) error {
	return attachPreciseArtifactErrorLocked(statusErrorLocked(status))
}

func attachPreciseArtifactErrorLocked(err error) error {
	if err == nil {
		return nil
	}
	var primary *StatusError
	if !errors.As(err, &primary) {
		return err
	}
	var raw C.SidereonPreciseInterpolantArtifactError
	if detailErr := statusErrorLocked(uint32(C.sidereon_last_precise_interpolant_artifact_error(&raw))); detailErr != nil {
		return errors.Join(err, detailErr)
	}
	if raw.kind == 0 {
		return err
	}
	value := PreciseArtifactError{Kind: uint32(raw.kind), Version: uint16(raw.version), Tag: uint8(raw.tag), Available: uint64(raw.available), Declared: uint64(raw.declared), Offset: uint64(raw.offset), Length: uint64(raw.len), ExpectedChecksum: uint64(raw.expected_checksum64), FoundChecksum: uint64(raw.found_checksum64), ClaimedChecksum: uint64(raw.claimed_checksum64), DeclaredChecksum: uint64(raw.declared_checksum64), HasSatellite: bool(raw.has_satellite), Satellite: tokenFromC(raw.satellite)}
	for i := range value.FoundMagic {
		value.FoundMagic[i] = byte(raw.found_magic[i])
	}
	for _, part := range []struct {
		code        C.uint32_t
		destination *string
	}{
		{C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_TEXT_PATH, &value.Path},
		{C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_TEXT_MESSAGE, &value.Message},
		{C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_TEXT_REASON, &value.Reason},
		{C.SIDEREON_PRECISE_INTERPOLANT_ARTIFACT_ERROR_TEXT_REGION, &value.Region},
	} {
		text, detailErr := copyNativeBytesLocked("precise artifact error text", func(out *C.uint8_t, n C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_last_precise_interpolant_artifact_error_text(part.code, out, n, written, required)
		})
		if detailErr != nil {
			return errors.Join(err, detailErr)
		}
		*part.destination = string(text)
	}
	primary.PreciseArtifact = &value
	return err
}
