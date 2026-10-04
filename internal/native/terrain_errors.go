package native

// Terrain errors mirror the typed values exposed by the C34 ABI.
type TerrainDatumError struct {
	Kind    uint32
	Terrain TerrainLookupError
	Geoid   GeoidError
	// Legacy text fields are retained for source compatibility and are populated
	// from C34's per-thread terrain text side channel when available.
	Path, Message, Remediation string
}

type TerrainStoreError struct {
	Kind                            uint32
	Version                         uint16
	Tag                             uint8
	LatIndex, LonIndex              int32
	ExpectedTileID, FoundTileID     TerrainTileID
	ExpectedChecksum, FoundChecksum uint64
	Field                           string
	HasHorizontalDatum              bool
	HorizontalDatum                 HorizontalDatum
	HasTileError                    bool
	TileError                       DtedTileError
	// Legacy text fields are retained for source compatibility and are populated
	// from C34's per-thread terrain text side channel when available.
	Path, Message, Reason string
}

type TerrainLookupError struct {
	Kind                                  uint32
	HasTile                               bool
	LatIndex, LonIndex                    int32
	HasPosting                            bool
	LatitudePosting, LongitudePosting     uint64
	HasHorizontalDatum                    bool
	HorizontalDatum                       HorizontalDatum
	HasOrigin                             bool
	OriginLatitudeDeg, OriginLongitudeDeg int32
	HasTileError                          bool
	TileError                             DtedTileError
	// Path and Message are copied from the C thread-local text side channel.
	Path, Message string
}

type GeoidError struct {
	Kind                   uint32
	Expected, Found, Index uint64
	Field, Reason          string
}

type HorizontalDatum struct {
	Kind            uint32
	Text            string
	WGS84Compatible bool
}

type DtedTileError struct {
	Kind        uint32
	Field, Text string
	// Path and Message come from the DTED tile text family, not the fixed C struct.
	Path, Message                                                    string
	HasCounts                                                        bool
	LongitudeCount, LatitudeCount                                    uint64
	HasLengths                                                       bool
	ActualBytes, ExpectedBytes                                       uint64
	HasQuery                                                         bool
	LongitudeDeg, LatitudeDeg, OriginLongitudeDeg, OriginLatitudeDeg float64
	HasLongitudeIndex, HasLatitudeIndex                              bool
	LongitudeIndex, LatitudeIndex                                    uint64
	HasChecksum                                                      bool
	Checksum, Sum                                                    int32
	HasHemisphere                                                    bool
	Hemisphere                                                       uint32
	ExpectedHemispheres                                              string
	HasNegativeIndex                                                 bool
	NegativeIndex                                                    int64
	HasInterval                                                      bool
	IntervalTenthsArcsec                                             uint32
	Count                                                            uint64
	HasDeclared                                                      bool
	Declared                                                         int32
}

func validDtedLookupOptions(options DtedLookupOptions) bool {
	return options.Interpolation == DtedInterpolationNearestPostingValue || options.Interpolation == DtedInterpolationBilinearValue
}

func (e *TerrainDatumError) Error() string {
	if e == nil {
		return "sidereon: terrain datum error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Remediation != "" {
		return e.Remediation
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain datum error"
}

func (e *TerrainStoreError) Error() string {
	if e == nil {
		return "sidereon: terrain store error"
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return e.Reason
	}
	if e.Path != "" {
		return e.Path
	}
	return "sidereon: terrain store error"
}

func (e *TerrainLookupError) Error() string {
	if e == nil {
		return "sidereon: terrain lookup error"
	}
	if e.Message != "" {
		return e.Message
	}
	return "sidereon: terrain lookup error"
}
