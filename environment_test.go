//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package sidereon

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func environmentFixture(t *testing.T, name string) []byte {
	t.Helper()
	value, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return value
}

func floatBits(t *testing.T, value string) uint64 {
	t.Helper()
	if len(value) > 2 && value[:2] == "0x" {
		bits, err := strconv.ParseUint(value[2:], 16, 64)
		if err != nil {
			t.Fatalf("parse float bits %q: %v", value, err)
		}
		return bits
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		t.Fatalf("parse float %q: %v", value, err)
	}
	return math.Float64bits(parsed)
}

func assertFloatBits(t *testing.T, got float64, want string) {
	t.Helper()
	if gotBits := math.Float64bits(got); gotBits != floatBits(t, want) {
		t.Fatalf("got %x (%g), want %s", gotBits, got, want)
	}
}

func assertFloatNear(t *testing.T, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("got %.15g, want %.15g ± %.3g", got, want, tolerance)
	}
}

func runEnvironmentReads(t *testing.T, count int, call func() error) {
	t.Helper()
	var wg sync.WaitGroup
	errCh := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				if err := call(); err != nil {
					errCh <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
}

func runReadCloseRace(t *testing.T, read func() error, closeHandle func() error) {
	t.Helper()
	const readerCount = 8
	start := make(chan struct{})
	ready := make(chan struct{}, readerCount)
	results := make(chan error, readerCount)
	var readers sync.WaitGroup
	for i := 0; i < readerCount; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			<-start
			ready <- struct{}{}
			for j := 0; j < 64; j++ {
				if err := read(); err != nil {
					if errors.Is(err, ErrClosed) {
						return
					}
					results <- err
					return
				}
			}
		}()
	}
	closeResult := make(chan error, 1)
	go func() {
		for i := 0; i < readerCount; i++ {
			<-ready
		}
		closeResult <- closeHandle()
	}()
	close(start)
	readers.Wait()
	if err := <-closeResult; err != nil {
		t.Fatalf("concurrent Close: %v", err)
	}
	close(results)
	for err := range results {
		t.Fatalf("concurrent read: %v", err)
	}
}

func TestEnvironmentFixtureBytesArePinned(t *testing.T) {
	want := map[string]string{
		"antex/igs20_wettzell_trim.atx":          "5c30f41a7cb75564eb379fcbc10e123ebafce86f8c305836413cdeaa129cfa02",
		"dted/tiles/n36_w106_1arc_v3.dt2":        "5c02badbacf28d09d942128add5093c9513def479a94525bfb2e51c6ca9ed1b9",
		"dted/tiles/n36_w107_1arc_v3.dt2":        "afaee9b556eb31c70268561620c8f024ab37bad176f19fb2252b39c2ae7f492f",
		"dted/dted_points.json":                  "1c16fad48c9d2d4a935d63c5cfbb2edb5a53299d4f77ac8db51dcec743f35f10",
		"geoid/egm2008_25_norcal_crop.bin":       "e66da6cbde7bb4015dc8b9c436fd93f16af3734e97017700fa3ab632f71f569d",
		"space_weather/SW-All-20260702-trim.csv": "c9c339aaccc4be3bc20f94fcecfd26a68cc54152159a08e7809050408d7a1310",
		"space_weather/SW-All-20260702-trim.txt": "815cf91852aad32681ba88a41d9a521395b14fa9f855dae684c3ebb214c245a5",
	}
	for name, expected := range want {
		data := environmentFixture(t, name)
		digest := sha256.Sum256(data)
		if got := hex.EncodeToString(digest[:]); got != expected {
			t.Errorf("fixture %s hash %s, want %s", name, got, expected)
		}
	}
}

func TestEnvironmentTerrainDiagnosticsStayWithFailedOperation(t *testing.T) {
	const goroutineCount = 64
	var wg sync.WaitGroup
	results := make(chan error, goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := EGM96FifteenMinuteGeoidFromBytes([]byte{0})
			if err == nil {
				results <- errors.New("malformed EGM96 DAC unexpectedly succeeded")
				return
			}
			var detail *TerrainDatumError
			if !errors.As(err, &detail) {
				results <- errors.New("malformed EGM96 DAC did not preserve terrain datum detail")
				return
			}
			if detail.Kind != TerrainDatumErrorGeoid || detail.Message == "" {
				results <- errors.New("malformed EGM96 DAC returned incomplete terrain datum detail")
				return
			}
			var status *StatusError
			if !errors.As(err, &status) || status.Code != StatusInvalidArgument || status.Detail == "" {
				results <- errors.New("malformed EGM96 DAC did not preserve general status detail")
			}
		}()
	}
	wg.Wait()
	close(results)
	for err := range results {
		t.Fatal(err)
	}

	storeBytes, err := DTEDTreeToMMapStore("testdata/dted/tiles")
	if err != nil {
		t.Fatal(err)
	}
	corrupt := append([]byte(nil), storeBytes...)
	corrupt[len(corrupt)-1] ^= 1
	if _, err := MMapTerrainFromBytes(corrupt); err == nil {
		t.Fatal("corrupt terrain store unexpectedly succeeded")
	} else {
		var detail *TerrainStoreError
		if !errors.As(err, &detail) {
			t.Fatalf("corrupt terrain store error = %v, want terrain store detail", err)
		}
		if detail.Kind != TerrainStoreErrorChecksum || detail.ExpectedChecksum == detail.FoundChecksum {
			t.Fatalf("corrupt terrain store detail = %#v", detail)
		}
		var status *StatusError
		if !errors.As(err, &status) || status.Code != StatusInvalidArgument || status.TerrainStore == nil {
			t.Fatalf("corrupt terrain store status = %#v, %v", status, err)
		}
	}

	path := filepath.Join(t.TempDir(), "terrain.store")
	if err := os.WriteFile(path, storeBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	claimed := uint64(0xff514a676a94d478)
	terrain, err := MMapTerrainFromPathAttested(path, claimed)
	if err != nil {
		t.Fatal(err)
	}
	if err := terrain.Verify(); err == nil {
		t.Fatal("attested checksum mismatch unexpectedly verified")
	} else {
		var detail *TerrainStoreError
		if !errors.As(err, &detail) {
			t.Fatalf("attested verification error = %v, want terrain store detail", err)
		}
		if detail.Kind != TerrainStoreErrorAttestedChecksumMismatch || detail.ExpectedChecksum != claimed || detail.FoundChecksum != 0xff514a676a94d479 {
			t.Fatalf("attested verification detail = %#v", detail)
		}
	}
	if err := terrain.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentANTEXFixture(t *testing.T) {
	fixture := environmentFixture(t, "antex/igs20_wettzell_trim.atx")
	product, err := ParseANTEX(fixture)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := product.Close(); err != nil {
			t.Error(err)
		}
	}()
	fixture[0] ^= 0xff
	count, err := product.AntennaCount()
	if err != nil || count != 10 {
		t.Fatalf("AntennaCount = %d, %v; want 10", count, err)
	}
	id := "BLOCK IIR-M         G05                 G050      2009-043A"
	antenna, found, err := product.Antenna(id)
	if err != nil || !found || antenna == nil {
		t.Fatalf("Antenna(%q) = %v, %v, %v", id, antenna, found, err)
	}
	defer func() {
		if err := antenna.Close(); err != nil {
			t.Error(err)
		}
	}()
	pco, err := antenna.PCO("G01")
	if err != nil {
		t.Fatal(err)
	}
	assertFloatBits(t, pco.NorthM, "-0.0033")
	assertFloatBits(t, pco.EastM, "-0.0003")
	assertFloatBits(t, pco.UpM, "0.74263")
	pcv, err := antenna.PCV("G01", 9, nil)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatBits(t, pcv, "-0.0095")
	azimuth := 0.0
	pcv, err = antenna.PCV("G01", 9, &azimuth)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatBits(t, pcv, "-0.0095")
	runEnvironmentReads(t, 8, func() error {
		_, err := antenna.PCO("G01")
		return err
	})
	receiver, found, err := product.Antenna("LEIAR25.R3      LEIT")
	if err != nil || !found || receiver == nil {
		t.Fatalf("receiver lookup = %v, %v, %v", receiver, found, err)
	}
	receiverPCO, err := receiver.PCO("G01")
	if err != nil {
		t.Fatal(err)
	}
	assertFloatNear(t, receiverPCO.UpM, 0.16096, 1e-15)
	if err := receiver.Close(); err != nil {
		t.Fatal(err)
	}
	if err := receiver.Close(); err != nil {
		t.Fatal(err)
	}
	encoded, err := product.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) == 0 {
		t.Fatal("ANTEX encode returned no bytes")
	}
	roundTrip, err := ParseANTEX(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if roundTripCount, err := roundTrip.AntennaCount(); err != nil || roundTripCount != 10 {
		t.Fatalf("ANTEX round-trip count = %d, %v", roundTripCount, err)
	}
	if err := roundTrip.Close(); err != nil {
		t.Fatal(err)
	}
	if err := roundTrip.Close(); err != nil {
		t.Fatal(err)
	}
	runEnvironmentReads(t, 8, func() error {
		_, err := product.AntennaCount()
		return err
	})
	runReadCloseRace(t, func() error {
		_, err := antenna.PCO("G01")
		return err
	}, antenna.Close)
	if err := antenna.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := antenna.PCO("G01"); !errors.Is(err, ErrClosed) {
		t.Fatalf("antenna use after close = %v, want ErrClosed", err)
	}
	runReadCloseRace(t, func() error {
		_, err := product.AntennaCount()
		return err
	}, product.Close)
	if err := product.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := product.AntennaCount(); !errors.Is(err, ErrClosed) {
		t.Fatalf("ANTEX use after close = %v, want ErrClosed", err)
	}
}

type dtedFixture struct {
	MultiTileCases []struct {
		CaseID        string `json:"case_id"`
		LatitudeBits  string `json:"latitude_bits"`
		LongitudeBits string `json:"longitude_bits"`
		NearestBits   string `json:"nearest_bits"`
		BilinearBits  string `json:"bilinear_bits"`
	} `json:"multi_tile_cases"`
}

func dtedFloat(bits string) float64 {
	value, _ := strconv.ParseUint(bits[2:], 16, 64)
	return math.Float64frombits(value)
}

func TestEnvironmentDTEDAndMMapFixtures(t *testing.T) {
	var fixture dtedFixture
	if err := json.Unmarshal(environmentFixture(t, "dted/dted_points.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	points := make([]LonLatDeg, len(fixture.MultiTileCases))
	for i, item := range fixture.MultiTileCases {
		points[i] = LonLatDeg{LongitudeDeg: dtedFloat(item.LongitudeBits), LatitudeDeg: dtedFloat(item.LatitudeBits)}
	}
	terrain, err := NewDTEDTerrain("testdata/dted/tiles")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := terrain.Close(); err != nil {
			t.Error(err)
		}
	}()
	options := DTEDLookupOptions{Interpolation: DTEDBilinear}
	results, err := terrain.HeightBatch(points, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != len(points) {
		t.Fatalf("DTED batch length = %d, want %d", len(results), len(points))
	}
	for i, item := range fixture.MultiTileCases {
		if !results[i].HasHeightM {
			t.Fatalf("DTED %s has no height", item.CaseID)
		}
		assertFloatBits(t, results[i].HeightM, item.BilinearBits)
		got, err := terrain.HeightMWithOptions(points[i].LongitudeDeg, points[i].LatitudeDeg, options)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatBits(t, got, item.BilinearBits)
	}
	options.Interpolation = DTEDNearestPosting
	nearest, err := terrain.HeightBatch(points, options)
	if err != nil {
		t.Fatal(err)
	}
	for i, item := range fixture.MultiTileCases {
		assertFloatBits(t, nearest[i].HeightM, item.NearestBits)
	}
	runEnvironmentReads(t, 8, func() error {
		_, err := terrain.HeightMWithOptions(points[0].LongitudeDeg, points[0].LatitudeDeg, DTEDLookupOptions{Interpolation: DTEDBilinear})
		return err
	})
	runReadCloseRace(t, func() error {
		_, err := terrain.HeightMWithOptions(points[0].LongitudeDeg, points[0].LatitudeDeg, DTEDLookupOptions{Interpolation: DTEDBilinear})
		return err
	}, terrain.Close)
	tile, err := LoadDTEDTile("testdata/dted/tiles/n36_w107_1arc_v3.dt2")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := tile.Elevation(points[0].LongitudeDeg, points[0].LatitudeDeg); err != nil || got != -20 {
		t.Fatalf("DTED tile elevation = %d, %v; want -20", got, err)
	}
	runEnvironmentReads(t, 8, func() error {
		_, err := tile.Elevation(points[0].LongitudeDeg, points[0].LatitudeDeg)
		return err
	})
	runReadCloseRace(t, func() error {
		_, err := tile.Elevation(points[0].LongitudeDeg, points[0].LatitudeDeg)
		return err
	}, tile.Close)
	if err := tile.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tile.Elevation(points[0].LongitudeDeg, points[0].LatitudeDeg); !errors.Is(err, ErrClosed) {
		t.Fatalf("DTED tile use after close = %v, want ErrClosed", err)
	}
	storeBytes, err := DTEDTreeToMMapStore("testdata/dted/tiles")
	if err != nil {
		t.Fatal(err)
	}
	entries := []DTEDTileListEntry{
		{TileID: TerrainTileID{LatIndex: 36, LonIndex: -106}, Path: "testdata/dted/tiles/n36_w106_1arc_v3.dt2"},
		{TileID: TerrainTileID{LatIndex: 36, LonIndex: -107}, Path: "testdata/dted/tiles/n36_w107_1arc_v3.dt2"},
	}
	listBytes, err := DTEDTileListToMMapStore(entries)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(listBytes, storeBytes) {
		t.Fatal("DTED list and tree conversions produced different stores")
	}
	writtenPath := filepath.Join(t.TempDir(), "terrain.store")
	if err := WriteDTEDTileListToMMapStore(entries, writtenPath); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(writtenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written, storeBytes) {
		t.Fatal("Go-owned DTED writer changed the native store bytes")
	}
	checksum, err := TerrainStoreChecksum64(storeBytes)
	if err != nil || checksum != 0xff514a676a94d479 {
		t.Fatalf("terrain store checksum = %#x, %v", checksum, err)
	}
	storeInput := append([]byte(nil), storeBytes...)
	store, err := MMapTerrainFromBytes(storeInput)
	if err != nil {
		t.Fatal(err)
	}
	storeInput[0] ^= 0xff
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	if datum, err := store.VerticalDatum(); err != nil || datum != EGM96MSLOrthometric {
		t.Fatalf("vertical datum = %v, %v", datum, err)
	}
	if provenance, err := store.DigestProvenance(); err != nil || provenance != DigestVerified {
		t.Fatalf("digest provenance = %v, %v", provenance, err)
	}
	index, err := store.TileIndex()
	if err != nil || len(index) != 2 {
		t.Fatalf("tile index = %d, %v; want 2", len(index), err)
	}
	serialized, err := store.ToBytes()
	if err != nil || !bytes.Equal(serialized, storeBytes) {
		t.Fatalf("store round trip mismatch: %v", err)
	}
	batch, err := store.HeightBatch(points, DTEDLookupOptions{Interpolation: DTEDBilinear})
	if err != nil {
		t.Fatal(err)
	}
	for i, item := range fixture.MultiTileCases {
		if !batch[i].HasOrthometricHeightM {
			t.Fatalf("mmap %s has no height", item.CaseID)
		}
		assertFloatBits(t, batch[i].OrthometricHeightM, item.BilinearBits)
	}
	zeroDAC := make([]byte, 721*1440*2)
	fifteenInput := append([]byte(nil), zeroDAC...)
	fifteen, err := EGM96FifteenMinuteGeoidFromBytes(fifteenInput)
	if err != nil {
		t.Fatal(err)
	}
	fifteenInput[0] ^= 0xff
	runEnvironmentReads(t, 8, func() error {
		_, err := store.EllipsoidalHeightMWithModel(points[0].LongitudeDeg, points[0].LatitudeDeg, DTEDLookupOptions{Interpolation: DTEDBilinear}, EGM96FifteenMinute, fifteen)
		return err
	})
	runReadCloseRace(t, func() error {
		_, err := store.EllipsoidalHeightMWithModel(points[0].LongitudeDeg, points[0].LatitudeDeg, DTEDLookupOptions{Interpolation: DTEDBilinear}, EGM96FifteenMinute, fifteen)
		return err
	}, fifteen.Close)
	if err := fifteen.Close(); err != nil {
		t.Fatal(err)
	}
	if err := fifteen.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.EllipsoidalHeightMWithModel(points[0].LongitudeDeg, points[0].LatitudeDeg, DTEDLookupOptions{Interpolation: DTEDBilinear}, EGM96FifteenMinute, fifteen); !errors.Is(err, ErrClosed) {
		t.Fatalf("closed 15-minute geoid use = %v, want ErrClosed", err)
	}
	runReadCloseRace(t, func() error {
		_, err := store.Checksum64()
		return err
	}, store.Close)
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Checksum64(); !errors.Is(err, ErrClosed) {
		t.Fatalf("mmap use after close = %v, want ErrClosed", err)
	}
	if err := terrain.Close(); err != nil {
		t.Fatal(err)
	}
	if err := terrain.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := terrain.HeightM(0, 0); !errors.Is(err, ErrClosed) {
		t.Fatalf("DTED use after close = %v, want ErrClosed", err)
	}
}

func TestEnvironmentGeoidFixtures(t *testing.T) {
	values := []float64{10, 20, 30, 40}
	grid, err := NewGeoidGrid(0, 0, 1, 1, 2, 2, values)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = -1000
	got, err := grid.UndulationsDeg([]GeoidPointDeg{{LatitudeDeg: 0, LongitudeDeg: 0}, {LatitudeDeg: 1, LongitudeDeg: 1}})
	if err != nil || len(got) != 2 || got[0] != 10 || got[1] != 40 {
		t.Fatalf("geoid degree batch = %#v, %v", got, err)
	}
	rad, err := grid.UndulationsRad([]GeoidPointRad{{LatitudeRad: 0, LongitudeRad: 0}})
	if err != nil || len(rad) != 1 || rad[0] != 10 {
		t.Fatalf("geoid radian batch = %#v, %v", rad, err)
	}
	if _, detail, err := grid.UndulationPROJRad(0.5*math.Pi/180, 0.5*math.Pi/180, PROJVGridshiftArithmeticFusedMultiplyAdd); err != nil || detail.Kind != PROJVGridshiftErrorNone {
		t.Fatalf("PROJ geoid lookup = %#v, %v", detail, err)
	}
	runEnvironmentReads(t, 8, func() error {
		_, err := grid.UndulationDeg(0, 0)
		return err
	})
	runReadCloseRace(t, func() error {
		_, err := grid.UndulationDeg(0, 0)
		return err
	}, grid.Close)
	if err := grid.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := grid.UndulationDeg(0, 0); !errors.Is(err, ErrClosed) {
		t.Fatalf("geoid use after close = %v, want ErrClosed", err)
	}
	cropInput := environmentFixture(t, "geoid/egm2008_25_norcal_crop.bin")
	crop, err := GeoidGridFromEGM2008RasterWindow(cropInput, EGM2008RasterWindow{Spacing: EGM2008TwoPointFiveMinute, LatMinDeg: 37, LonMinDeg: -123, NLat: 25, NLon: 25})
	if err != nil {
		t.Fatal(err)
	}
	cropInput[0] ^= 0xff
	for _, item := range []struct {
		lat, lon, want float64
	}{
		{37.7749, -122.4194, -32.163558372373},
		{37.5, -122.75, -33.605857849121},
		{37.875, -122.125, -31.847370147705},
		{38, -122, -31.767843246460},
		{37, -123, -36.499370574951},
	} {
		value, err := crop.UndulationDeg(item.lat, item.lon)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatNear(t, value, item.want, 5e-9)
	}
	if err := crop.Close(); err != nil {
		t.Fatal(err)
	}
	if err := crop.Close(); err != nil {
		t.Fatal(err)
	}
	zeroDAC := make([]byte, 721*1440*2)
	fifteenInput := append([]byte(nil), zeroDAC...)
	fifteen, err := EGM96FifteenMinuteGeoidFromBytes(fifteenInput)
	if err != nil {
		t.Fatal(err)
	}
	fifteenInput[0] ^= 0xff
	if err := fifteen.Close(); err != nil {
		t.Fatal(err)
	}
	if err := fifteen.Close(); err != nil {
		t.Fatal(err)
	}
	deg := []GeoidPointDeg{{LatitudeDeg: 37, LongitudeDeg: -122}, {LatitudeDeg: 0.125, LongitudeDeg: -179.875}}
	batch, err := EGM96UndulationsDeg(deg)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatBits(t, batch[0], "-33.35")
	assertFloatBits(t, batch[1], "0x4034cbf5c28f5c29")
	runEnvironmentReads(t, 8, func() error {
		_, err := GeoidUndulationsDeg([]GeoidPointDeg{{LatitudeDeg: 0, LongitudeDeg: 0}})
		return err
	})
}

func TestEnvironmentSpaceWeatherFixtures(t *testing.T) {
	csv := environmentFixture(t, "space_weather/SW-All-20260702-trim.csv")
	txt := environmentFixture(t, "space_weather/SW-All-20260702-trim.txt")
	for name, data := range map[string][]byte{"csv": csv, "txt": txt} {
		input := append([]byte(nil), data...)
		var table *SpaceWeatherTable
		var err error
		if name == "csv" {
			table, err = ParseSpaceWeatherCSV(input)
		} else {
			table, err = ParseSpaceWeatherTXT(input)
		}
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		input[len(input)-1] ^= 0xff
		days, err := table.Days()
		if err != nil || len(days) != 14 {
			t.Fatalf("%s days = %d, %v; want 14", name, len(days), err)
		}
		monthly, err := table.Monthly()
		if err != nil || len(monthly) != 2 {
			t.Fatalf("%s monthly = %d, %v; want 2", name, len(monthly), err)
		}
		summary, err := table.Summary()
		if err != nil || summary.DayCount != 14 || summary.MonthlyCount != 2 {
			t.Fatalf("%s summary = %#v, %v", name, summary, err)
		}
		if days[0].Class != SpaceWeatherObservationObserved || days[11].Class != SpaceWeatherObservationDailyPredicted || monthly[0].Class != SpaceWeatherObservationMonthlyPredicted {
			t.Fatalf("%s observation classes = %v, %v, %v", name, days[0].Class, days[11].Class, monthly[0].Class)
		}
		day, present, err := table.Day(2026, 7, 1)
		if err != nil || !present || day.Kp10[0] != 40 || day.Kp10[7] != 17 || day.ApAvg != 12 || day.F107Obs != 249.7 {
			t.Fatalf("%s day = %#v, %v, %v", name, day, present, err)
		}
		if name == "csv" {
			encoded, err := table.ToCSV()
			if err != nil || !bytes.Equal(encoded, data) {
				t.Fatalf("CSV round trip mismatch: %v", err)
			}
		} else {
			encoded, err := table.ToTXT()
			if err != nil || !bytes.Equal(encoded, data) {
				t.Fatalf("TXT round trip mismatch: %v", err)
			}
		}
		epoch, err := CivilToJ2000Seconds(CivilDateTime{Year: 2026, Month: 7, Day: 1, Hour: 12})
		if err != nil {
			t.Fatal(err)
		}
		sample, err := table.SampleAt(epoch)
		if err != nil {
			t.Fatal(err)
		}
		assertFloatBits(t, sample.Weather.F107, "202.6")
		assertFloatBits(t, sample.Weather.F107A, "145.9")
		assertFloatBits(t, sample.Weather.Ap, "12")
		if sample.Class != SpaceWeatherObservationObserved {
			t.Fatalf("%s sample class = %v, want observed", name, sample.Class)
		}
		runEnvironmentReads(t, 8, func() error {
			_, err := table.Coverage()
			return err
		})
		runReadCloseRace(t, func() error {
			_, err := table.Coverage()
			return err
		}, table.Close)
		if err := table.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := table.Days(); !errors.Is(err, ErrClosed) {
			t.Fatalf("%s use after close = %v, want ErrClosed", name, err)
		}
	}
	parsed, err := ParseSpaceWeatherTable(txt)
	if err != nil {
		t.Fatal(err)
	}
	if err := parsed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := parsed.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentAtmosphereTropoRFAndLabels(t *testing.T) {
	input := DefaultAtmosphereInput()
	input.Year = 0
	input.DayOfYear = 172
	input.Second = 29000
	input.AltitudeKm = 0
	input.LatitudeDeg = 0
	input.LongitudeDeg = 0
	input.F107 = 150
	input.F107A = 150
	input.Ap = 4
	atmosphere, err := NRLMSISE00(input)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatNear(t, atmosphere.DensityKgM3, 1.16827744668804, 1e-12)
	if atmosphere.TemperatureK < 270 || atmosphere.TemperatureK > 300 {
		t.Fatalf("atmosphere temperature = %g", atmosphere.TemperatureK)
	}
	met := Met{PressureHPa: 1013, TemperatureK: 288.15, RelativeHumidity: 0.5}
	receiver := Geodetic{LatitudeRad: math.Pi / 4, LongitudeRad: 8 * math.Pi / 180}
	date, _, err := InstantFromUTCCivil(CivilDateTime{Year: 2020, Month: 1, Day: 28})
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := TropoMappingFactors(math.Pi/2, receiver, UTC, date.Whole, date.Fraction)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatBits(t, mapping.Dry, "0x3ff0000000000000")
	assertFloatBits(t, mapping.Wet, "0x3ff0000000000000")
	zenith, err := TropoZenithDelay(receiver, met)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatNear(t, zenith.DryM, 2.307, 0.01)
	if zenith.WetM <= 0 {
		t.Fatalf("wet zenith delay = %g", zenith.WetM)
	}
	slant, err := TropoSlantDelay(math.Pi/2, receiver, met, UTC, date.Whole, date.Fraction)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatNear(t, slant, zenith.DryM+zenith.WetM, 1e-12)
	_, detail, err := TropoMappingFactorsChecked(0, receiver, UTC, date.Whole, date.Fraction)
	if err == nil || detail.Kind != TropoMappingErrorLowElevation {
		t.Fatalf("checked low elevation = %#v, %v", detail, err)
	}
	fspl, err := RFFSPL(1200, 1616)
	if err != nil {
		t.Fatal(err)
	}
	assertFloatNear(t, fspl, 158.201, 0.01)
	eirp, err := RFEIRP(27, 3)
	if err != nil || eirp != 0 {
		t.Fatalf("EIRP = %g, %v; want 0", eirp, err)
	}
	cn0, err := RFCN0(0, 165, -12, 3)
	if err != nil {
		t.Fatal(err)
	}
	margin, err := RFLinkMargin(LinkBudget{EIRPdBW: 0, FSPLdB: 165, ReceiverGTdBK: -12, OtherLossesdB: 3, RequiredCN0dBHz: 35})
	if err != nil || margin != cn0-35 {
		t.Fatalf("link margin = %g, %v; CN0 = %g", margin, err, cn0)
	}
	if wavelength, err := RFWavelength(1616e6); err != nil || wavelength <= 0 {
		t.Fatalf("wavelength = %g, %v", wavelength, err)
	}
	if gain, err := RFDishGain(1, 1616e6, 0.55); err != nil || gain <= 0 {
		t.Fatalf("dish gain = %g, %v", gain, err)
	}
	labels := map[ObservabilityTier]string{ObservabilityRankDeficient: "rank_deficient", ObservabilityZeroRedundancy: "zero_redundancy", ObservabilityWeak: "weak", ObservabilityNominal: "nominal"}
	for tier, want := range labels {
		got, err := ObservabilityTierLabel(tier)
		if err != nil || got != want {
			t.Fatalf("observability label %v = %q, %v; want %q", tier, got, err, want)
		}
	}
	if _, err := DTEDInterpolationLabel(DTEDInterpolation(99)); err == nil {
		t.Fatal("invalid DTED selector unexpectedly accepted")
	}
	if _, err := VerticalDatumLabel(VerticalDatum(99)); err == nil {
		t.Fatal("invalid datum unexpectedly accepted")
	}
	if _, err := TerrainGeoidModelLabel(TerrainGeoidModel(99)); err == nil {
		t.Fatal("invalid terrain geoid model unexpectedly accepted")
	}
}

func TestEnvironmentGeofenceFixture(t *testing.T) {
	fence, detail, err := NewGeofence([]Geodetic{
		{LatitudeRad: -0.01, LongitudeRad: 0},
		{LatitudeRad: -0.01, LongitudeRad: 0.02},
		{LatitudeRad: 0.01, LongitudeRad: 0.02},
		{LatitudeRad: 0.01, LongitudeRad: 0},
	})
	if err != nil || detail != GeofenceErrorNone {
		t.Fatalf("NewGeofence = %v, %v", detail, err)
	}
	defer func() {
		if err := fence.Close(); err != nil {
			t.Error(err)
		}
	}()
	inside := Geodetic{LatitudeRad: 0, LongitudeRad: 0.01}
	outside := Geodetic{LatitudeRad: 0, LongitudeRad: 0.03}
	if got, detail, err := fence.Contains(inside); err != nil || !got || detail != GeofenceErrorNone {
		t.Fatalf("inside = %v, %v, %v", got, detail, err)
	}
	if got, detail, err := fence.Contains(outside); err != nil || got || detail != GeofenceErrorNone {
		t.Fatalf("outside = %v, %v, %v", got, detail, err)
	}
	if distance, detail, err := fence.DistanceToBoundary(inside); err != nil || distance <= 0 || detail != GeofenceErrorNone {
		t.Fatalf("inside distance = %g, %v, %v", distance, detail, err)
	}
	if distance, detail, err := fence.DistanceToBoundary(outside); err != nil || distance >= 0 || detail != GeofenceErrorNone {
		t.Fatalf("outside distance = %g, %v, %v", distance, detail, err)
	}
	uncertainty := GeofenceUncertainty{Kind: GeofenceENUCovarianceM2, CovarianceM2: Matrix3{{400, 0, 0}, {0, 400, 0}, {0, 0, 0}}}
	probability, detail, err := fence.ContainmentProbability(Geodetic{LatitudeRad: 0, LongitudeRad: 0}, uncertainty)
	if err != nil || detail != GeofenceErrorNone || probability < 0.45 || probability > 0.55 {
		t.Fatalf("boundary probability = %g, %v, %v", probability, detail, err)
	}
	options := GeofenceProbabilityOptions{Method: GeofencePlanarQuadrature}
	corner, detail, err := fence.ContainmentProbabilityWithOptions(Geodetic{LatitudeRad: 0, LongitudeRad: 0}, uncertainty, options)
	if err != nil || detail != GeofenceErrorNone || corner < 0 || corner > 1 {
		t.Fatalf("corner probability = %g, %v, %v", corner, detail, err)
	}
	hysteresis := GeofenceHysteresis{EnterConfidence: 0.8, LeaveConfidence: 0.8}
	metersPerRad := 6378137.0
	samples := []GeofencePositionEstimate{
		{Position: Geodetic{LatitudeRad: 0, LongitudeRad: -20 / metersPerRad}, Uncertainty: uncertainty},
		{Position: Geodetic{LatitudeRad: 0, LongitudeRad: 20 / metersPerRad}, Uncertainty: uncertainty},
		{Position: Geodetic{LatitudeRad: 0, LongitudeRad: -20 / metersPerRad}, Uncertainty: uncertainty},
	}
	events, detail, err := fence.CrossingProbability(samples, hysteresis)
	if err != nil || detail != GeofenceErrorNone || len(events) != 2 {
		t.Fatalf("crossing events = %#v, %v, %v", events, detail, err)
	}
	if events[0].SampleIndex != 1 || events[0].Kind != GeofenceEntered || events[1].SampleIndex != 2 || events[1].Kind != GeofenceLeft {
		t.Fatalf("crossing events = %#v", events)
	}
	runEnvironmentReads(t, 8, func() error {
		_, _, err := fence.Contains(inside)
		return err
	})
	runReadCloseRace(t, func() error {
		_, _, err := fence.Contains(inside)
		return err
	}, fence.Close)
	if err := fence.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := fence.Contains(inside); !errors.Is(err, ErrClosed) {
		t.Fatalf("geofence use after close = %v, want ErrClosed", err)
	}
}

func TestEnvironmentRejectsInvalidInputsBeforeC(t *testing.T) {
	checks := []func() error{
		func() error { _, err := NewGeoidGrid(0, 0, 1, 1, -1, 2, nil); return err },
		func() error { _, err := GeoidGridFromEGM2008Raster(nil, EGM2008GridSpacing(99)); return err },
		func() error { _, err := EGM96FifteenMinuteGeoidFromPath("bad\x00path"); return err },
		func() error { _, err := NewDTEDTerrain("bad\x00root"); return err },
		func() error { _, err := LoadDTEDTile("bad\x00tile"); return err },
		func() error {
			_, err := DTEDTileListToMMapStore([]DTEDTileListEntry{{Path: "bad\x00tile"}})
			return err
		},
		func() error { _, err := MMapTerrainFromPath("bad\x00store"); return err },
		func() error { _, err := MMapTerrainFromPathAttested("bad\x00store", 1); return err },
	}
	for i, check := range checks {
		if err := check(); err == nil {
			t.Fatalf("invalid input %d unexpectedly succeeded", i)
		}
	}
	if _, _, err := NewGeofence(nil); err == nil {
		t.Fatal("empty geofence unexpectedly succeeded")
	}
	if _, err := ParseSpaceWeatherCSV(nil); err == nil {
		t.Fatal("empty space-weather CSV unexpectedly succeeded")
	}
	maxInt := int(^uint(0) >> 1)
	atmosphereInput := DefaultAtmosphereInput()
	atmosphereInput.Year = maxInt
	if _, err := NRLMSISE00(atmosphereInput); err == nil || !strings.Contains(err.Error(), "C int32") {
		t.Fatalf("out-of-range NRLMSISE year = %v", err)
	}
	atmosphereInput.Year = 0
	atmosphereInput.DayOfYear = maxInt
	if _, err := NRLMSISE00(atmosphereInput); err == nil || !strings.Contains(err.Error(), "C int32") {
		t.Fatalf("out-of-range NRLMSISE day of year = %v", err)
	}
	table, err := ParseSpaceWeatherCSV(environmentFixture(t, "space_weather/SW-All-20260702-trim.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := table.Day(maxInt, 7, 1); err == nil || !strings.Contains(err.Error(), "C int32") {
		t.Fatalf("out-of-range space-weather year = %v", err)
	}
	if err := table.Close(); err != nil {
		t.Fatal(err)
	}
	if err := table.Close(); err != nil {
		t.Fatal(err)
	}
}
