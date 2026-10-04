package sidereon

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Optional query fields use native NaN sentinels. Compare their stored bits,
// including NaN payloads, rather than NaN's arithmetic equality.
func sameTileFailure(a, b DtedTileError) bool {
	left, right := reflect.ValueOf(a), reflect.ValueOf(b)
	for i := 0; i < left.NumField(); i++ {
		if left.Field(i).Kind() == reflect.Float64 {
			if math.Float64bits(left.Field(i).Float()) != math.Float64bits(right.Field(i).Float()) {
				return false
			}
		} else if !reflect.DeepEqual(left.Field(i).Interface(), right.Field(i).Interface()) {
			return false
		}
	}
	return true
}

func preciseArtifactChecksum(data []byte) uint64 {
	const offsetBasis uint64 = 14695981039346656037
	const prime uint64 = 1099511628211
	hash := offsetBasis
	for index, value := range data {
		if index >= 40 && index < 48 {
			value = 0
		}
		hash = (hash ^ uint64(value)) * prime
	}
	return hash
}

func TestStandaloneTerrainTypedFailuresAreOwned(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.dt2")
	_, err := LoadDTEDTile(missing)
	var tileFailure *DtedTileError
	if !errors.As(err, &tileFailure) || tileFailure.Kind == 0 || tileFailure.Path != missing || tileFailure.Message == "" {
		t.Fatalf("DTED load detail=%+v err=%v", tileFailure, err)
	}
	loadRetained := tileFailure
	loadSnapshot := *tileFailure
	tile, err := LoadDTEDTile("testdata/dted/tiles/n36_w107_1arc_v3.dt2")
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, tile)
	_, err = tile.Elevation(0, 0)
	if !errors.As(err, &tileFailure) || !tileFailure.HasQuery || tileFailure.LongitudeDeg != 0 || tileFailure.LatitudeDeg != 0 {
		t.Fatalf("DTED query detail=%+v err=%v", tileFailure, err)
	}
	queryRetained := tileFailure
	querySnapshot := *tileFailure
	_, err = NewGeoidGrid(0, 0, 0, 1, 2, 2, []float64{1, 2, 3, 4})
	var geoidFailure *GeoidError
	if !errors.As(err, &geoidFailure) || geoidFailure.Kind != uint32(GeoidErrorInvalidSpacing) || geoidFailure.Field == "" {
		t.Fatalf("geoid spacing detail=%+v err=%v", geoidFailure, err)
	}
	spacingRetained := geoidFailure
	spacingSnapshot := *geoidFailure
	_, err = NewGeoidGrid(0, 0, 1, 1, 2, 2, []float64{1, 2, 3})
	if !errors.As(err, &geoidFailure) || geoidFailure.Kind != uint32(GeoidErrorInvalidDimensions) || geoidFailure.Expected != 4 || geoidFailure.Found != 3 {
		t.Fatalf("geoid dimensions detail=%+v err=%v", geoidFailure, err)
	}
	_, err = NewGeoidGrid(0, 0, 1, 1, 2, 2, []float64{1, math.NaN(), 3, 4})
	if !errors.As(err, &geoidFailure) || geoidFailure.Kind != uint32(GeoidErrorNonFiniteValue) || geoidFailure.Index != 1 {
		t.Fatalf("geoid value detail=%+v err=%v", geoidFailure, err)
	}
	_, err = LoadDTEDTile("invalid\x00path")
	var status *StatusError
	if err == nil || (errors.As(err, &status) && (status.DtedTile != nil || status.Geoid != nil)) {
		t.Fatalf("marshalling error has stale native detail: %+v %v", status, err)
	}
	if !sameTileFailure(*loadRetained, loadSnapshot) || !sameTileFailure(*queryRetained, querySnapshot) || !reflect.DeepEqual(*spacingRetained, spacingSnapshot) {
		t.Fatal("owned prior terrain failures changed after subsequent operations")
	}
}

func TestPreciseArtifactTypedFailuresPreserveBytesAndCounts(t *testing.T) {
	data, err := os.ReadFile("testdata/trimmed.sp3")
	if err != nil {
		t.Fatal(err)
	}
	source, err := LoadSP3(data)
	if err != nil {
		t.Fatal(err)
	}
	closeAfterTest(t, source)
	artifact, kind, err := source.ArtifactBytes()
	if err != nil || kind != PreciseInterpolantArtifactErrorNone {
		t.Fatalf("artifact producer=%v %v", kind, err)
	}
	badMagic := append([]byte(nil), artifact...)
	copy(badMagic[:8], []byte("BADMAGIC"))
	_, kind, err = OpenPreciseInterpolantArtifact(badMagic)
	var failure *PreciseArtifactError
	if kind != PreciseInterpolantArtifactErrorBadMagic || !errors.As(err, &failure) || failure.Kind != kind || failure.FoundMagic != [8]byte{'B', 'A', 'D', 'M', 'A', 'G', 'I', 'C'} {
		t.Fatalf("artifact magic=%v %+v %v", kind, failure, err)
	}
	retained := failure
	snapshot := *failure
	_, kind, err = OpenPreciseInterpolantArtifact(artifact[:8])
	if kind != PreciseInterpolantArtifactErrorHeaderTruncated || !errors.As(err, &failure) || failure.Kind != kind || failure.Available != 8 {
		t.Fatalf("artifact short header=%v %+v %v", kind, failure, err)
	}
	_, kind, err = OpenPreciseInterpolantArtifact(append(append([]byte(nil), artifact...), 0))
	if kind != PreciseInterpolantArtifactErrorTrailingBytes || !errors.As(err, &failure) || failure.Kind != kind || failure.Declared != uint64(len(artifact)) || failure.Available != uint64(len(artifact)+1) {
		t.Fatalf("artifact trailing bytes=%v %+v %v", kind, failure, err)
	}
	if !reflect.DeepEqual(*retained, snapshot) {
		t.Fatal("owned artifact failure changed after subsequent native failures")
	}

	// The store wire format puts the declared checksum at header bytes 40..48.
	// Read those bytes independently of the native getter for the attestation.
	checksum := binary.LittleEndian.Uint64(artifact[40:48])
	path := filepath.Join(t.TempDir(), "precise.store")
	if err := os.WriteFile(path, artifact, 0600); err != nil {
		t.Fatal(err)
	}
	verified, kind, err := OpenPreciseInterpolantArtifactFile(path)
	if err != nil || kind != PreciseInterpolantArtifactErrorNone {
		t.Fatalf("verified file open=%v %v", kind, err)
	}
	closeAfterTest(t, verified)
	if provenance, err := verified.DigestProvenance(); err != nil || provenance != DigestVerified {
		t.Fatalf("file-open provenance=%v %v", provenance, err)
	}
	verifiedSatellites, err := verified.Satellites()
	if err != nil || len(verifiedSatellites) == 0 {
		t.Fatalf("file-backed satellite read=%v %v", verifiedSatellites, err)
	}
	secondMapping, kind, err := OpenPreciseInterpolantArtifactFile(path)
	if err != nil || kind != PreciseInterpolantArtifactErrorNone {
		t.Fatalf("second file open=%v %v", kind, err)
	}
	closeAfterTest(t, secondMapping)
	if err := verified.Close(); err != nil {
		t.Fatal(err)
	}
	if satellites, err := secondMapping.Satellites(); err != nil || !reflect.DeepEqual(satellites, verifiedSatellites) {
		t.Fatalf("independent mapped handle after peer close=%v %v", satellites, err)
	}
	if err := secondMapping.Close(); err != nil {
		t.Fatal(err)
	}

	missingPath := filepath.Join(t.TempDir(), "missing.store")
	_, kind, err = OpenPreciseInterpolantArtifactFile(missingPath)
	if kind != PreciseInterpolantArtifactErrorIO || !errors.As(err, &failure) || failure.Kind != kind || failure.Path != missingPath || failure.Message == "" {
		t.Fatalf("missing file detail=%v %+v %v", kind, failure, err)
	}

	corrupt := append([]byte(nil), artifact...)
	corrupt[len(corrupt)-1] ^= 1
	binary.LittleEndian.PutUint64(corrupt[40:48], preciseArtifactChecksum(corrupt))
	corruptPath := filepath.Join(t.TempDir(), "bad-payload.store")
	if err := os.WriteFile(corruptPath, corrupt, 0600); err != nil {
		t.Fatal(err)
	}
	_, kind, err = OpenPreciseInterpolantArtifactFile(corruptPath)
	if kind != PreciseInterpolantArtifactErrorSatelliteChecksum || !errors.As(err, &failure) || failure.Kind != kind || !failure.HasSatellite || failure.ExpectedChecksum == failure.FoundChecksum {
		t.Fatalf("satellite checksum detail=%v %+v %v", kind, failure, err)
	}

	_, kind, err = OpenPreciseInterpolantArtifactAttestedFile(path, checksum^1)
	if kind != PreciseInterpolantArtifactErrorAttestedChecksumMismatch || !errors.As(err, &failure) || failure.Kind != kind || failure.ClaimedChecksum != checksum^1 || failure.DeclaredChecksum != checksum {
		t.Fatalf("artifact checksum attestation=%v %+v %v", kind, failure, err)
	}
	mapped, kind, err := OpenPreciseInterpolantArtifactAttestedFile(path, checksum)
	if err != nil || kind != PreciseInterpolantArtifactErrorNone {
		t.Fatalf("attested file open=%v %v", kind, err)
	}
	closeAfterTest(t, mapped)
	if provenance, err := mapped.DigestProvenance(); err != nil || provenance != DigestAttested {
		t.Fatalf("initial provenance=%v %v", provenance, err)
	}
	if kind, err := mapped.Verify(); err != nil || kind != PreciseInterpolantArtifactErrorNone {
		t.Fatalf("mapped verify=%v %v", kind, err)
	}
	if provenance, err := mapped.DigestProvenance(); err != nil || provenance != DigestVerified {
		t.Fatalf("verified provenance=%v %v", provenance, err)
	}
}
