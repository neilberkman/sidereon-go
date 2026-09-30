//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "errors"

// This build is intentionally link-free. Use cgo and, on Linux, select one of
// sidereon_linux_glibc or sidereon_linux_musl. Use sidereon_use_system_lib
// with CGO_LDFLAGS when linking a system installation.
var ErrClosed = errors.New("sidereon: handle is closed")

type StatusError struct {
	Code          int
	Text          string
	Detail        string
	Engine        *EngineError
	SP3           *SP3Error
	TerrainDatum  *TerrainDatumError
	TerrainStore  *TerrainStoreError
	TerrainLookup *TerrainLookupError
	Bias          *BiasError
	RTCM          *RTCMError
}

type SelectionError struct {
	Status uint32
	Text   string
	Detail string
}

func (e *SelectionError) Error() string { return e.Text }

type FallbackError struct {
	Status uint32
	Detail string
}

func (e *FallbackError) Error() string {
	if e == nil {
		return "sidereon: fallback solve failed"
	}
	if e.Detail == "" {
		return "sidereon: fallback solve failed"
	}
	return e.Detail
}

func (e *StatusError) Error() string { return e.Text }

func (e *StatusError) Unwrap() error {
	var details []error
	if e.Engine != nil {
		details = append(details, e.Engine)
	}
	if e.TerrainDatum != nil {
		details = append(details, e.TerrainDatum)
	}
	if e.TerrainStore != nil {
		details = append(details, e.TerrainStore)
	}
	if e.SP3 != nil {
		details = append(details, e.SP3)
	}
	if e.Bias != nil {
		details = append(details, e.Bias)
	}
	if e.RTCM != nil {
		details = append(details, e.RTCM)
	}
	return errors.Join(details...)
}

func (e *StatusError) EngineError() *EngineError {
	if e == nil {
		return nil
	}
	return e.Engine
}

func ClearEngineError() {}

func ClearEngineErrorLocked() {}

func CThreadDepth() int { return 0 }

func CurrentEngineErrorLocked() (*EngineError, error) {
	return nil, unavailable()
}

type Version struct {
	Major  uint32
	Minor  uint32
	Patch  uint32
	String string
}

func LibraryVersion() Version { return Version{} }

func unavailable() error {
	return errors.New("sidereon: this platform or build configuration requires cgo and a supported native ABI selection")
}

func SecondOfDay(int, int, float64) (float64, error)              { return 0, unavailable() }
func DayOfYear(int, int, int, int, int, float64) (float64, error) { return 0, unavailable() }
func DataDayOfYear(int, uint8, uint8) (uint16, error)             { return 0, unavailable() }

type CovarianceValidation struct {
	Symmetric            bool
	PositiveSemidefinite bool
}

func CovarianceFromDiagonal([]float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}
func CovarianceValidate([6][6]float64) (CovarianceValidation, error) {
	return CovarianceValidation{}, unavailable()
}
func CovarianceKmToM([6][6]float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}
func CovarianceMToKm([6][6]float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}
func CovarianceInterpolate([6][6]float64, [6][6]float64, float64) ([6][6]float64, error) {
	return [6][6]float64{}, unavailable()
}

type NMEASummary struct {
	SentenceCount uint64
	EpochCount    uint64
	SkipCount     uint64
	WarningCount  uint64
}

type NMEAEpoch struct {
	HasCalendarEpoch   bool
	CalendarEpoch      NativeCalendarEpoch
	HasPosition        bool
	LatitudeRad        float64
	LongitudeRad       float64
	HeightM            float64
	HasInstantJ2000S   bool
	InstantJ2000S      float64
	HasPDOP            bool
	PDOP               float64
	HasHDOP            bool
	HDOP               float64
	HasVDOP            bool
	VDOP               float64
	SentenceCount      uint64
	UsedSatelliteCount uint64
	SatellitesInView   uint64
	SkipCount          uint64
	WarningCount       uint64
	HasGGA             bool
	HasRMC             bool
	HasGLL             bool
	GSACount           uint64
	GSVGroupCount      uint64
}

type NMEALog struct{}

func ParseNMEA([]byte) (*NMEALog, error) { return nil, unavailable() }
func (*NMEALog) Close() error            { return nil }
func (*NMEALog) Summary() (NMEASummary, error) {
	return NMEASummary{}, unavailable()
}
func (*NMEALog) Epochs() ([]NMEAEpoch, error)                   { return nil, unavailable() }
func (*NMEALog) SentenceRecords() ([][]byte, error)             { return nil, unavailable() }
func (*NMEALog) EpochRecords() ([][]byte, error)                { return nil, unavailable() }
func (*NMEALog) Diagnostics() ([]NMEADiagnostic, error)         { return nil, unavailable() }
func (*NMEALog) EpochDiagnostics(int) ([]NMEADiagnostic, error) { return nil, unavailable() }

// SGP4 fitting declarations mirror the native API on unsupported builds.
type SGP4FitEpochKind uint32
type SGP4Loss uint32
type SGP4XScaleKind uint32

const (
	SGP4FitEpochMidpoint SGP4FitEpochKind = iota
	SGP4FitEpochFirst
	SGP4FitEpochLast
	SGP4FitEpochSample
	SGP4FitEpochJD
)

const (
	SGP4LossLinear SGP4Loss = iota
	SGP4LossSoftL1
	SGP4LossHuber
	SGP4LossCauchy
	SGP4LossArctan
)

const (
	SGP4XScaleNone SGP4XScaleKind = iota
	SGP4XScaleUnit
	SGP4XScaleValues
	SGP4XScaleJacobian
)

type SGP4FitSample struct {
	JDWhole, JDFraction float64
	PositionTEMEKm      [3]float64
	HasVelocityTEMEKmPS bool
	VelocityTEMEKmPS    [3]float64
}

type SGP4FitConfig struct {
	EpochKind                               SGP4FitEpochKind
	EpochSampleIndex                        int
	EpochJDWhole, EpochJDFraction           float64
	FitBStar                                bool
	BStarSeed                               float64
	UseVelocity                             bool
	HasVelocityWeightS                      bool
	VelocityWeightS                         float64
	Weights                                 []float64
	OpsMode                                 uint32
	HasFTol                                 bool
	FTol                                    float64
	HasXTol                                 bool
	XTol                                    float64
	HasGTol                                 bool
	GTol                                    float64
	HasMaxNFEV                              bool
	MaxNFEV                                 int
	XScaleKind                              SGP4XScaleKind
	XScaleValues                            []float64
	Loss                                    SGP4Loss
	FScale                                  float64
	CatalogNumber                           uint32
	Classification, InternationalDesignator string
	ElementSetNumber                        int32
	RevAtEpoch                              int64
	ObjectName                              string
}

type SGP4FitStatistics struct {
	RMSPositionKm, MaxPositionKm      float64
	RMSPositionAxesKm                 [3]float64
	HasRMSVelocityKmPS                bool
	RMSVelocityKmPS, TLERMSPositionKm float64
	Status                            int32
	NFEV, NJEV                        int
	Cost, Optimality                  float64
	BStarObservable                   bool
	SeedRefinePasses                  int
}

type SGP4TLEFit struct{}

func SGP4FitConfigDefaults() (SGP4FitConfig, error)                  { return SGP4FitConfig{}, unavailable() }
func FitSGP4TLE([]SGP4FitSample, SGP4FitConfig) (*SGP4TLEFit, error) { return nil, unavailable() }
func (*SGP4TLEFit) Close() error                                     { return nil }
func (*SGP4TLEFit) Lines() (TLELines, error)                         { return TLELines{}, unavailable() }
func (*SGP4TLEFit) OMM() (*OMM, error)                               { return nil, unavailable() }
func (*SGP4TLEFit) Statistics() (SGP4FitStatistics, error)           { return SGP4FitStatistics{}, unavailable() }
func (*SGP4TLEFit) ResultPayload() ([]byte, error)                   { return nil, unavailable() }
