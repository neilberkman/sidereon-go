//go:build !cgo || !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

import "time"

type SP3State struct{}

type SP3PredictionSummary struct{}

type SP3 struct{}

func LoadSP3([]byte) (*SP3, error) { return nil, unavailable() }
func (*SP3) Close() error          { return nil }
func (*SP3) EpochCount() (int, error) {
	return 0, unavailable()
}
func (*SP3) Epochs() ([]float64, error)          { return nil, unavailable() }
func (*SP3) Satellites() ([]string, error)       { return nil, unavailable() }
func (*SP3) State(string, int) (SP3State, error) { return SP3State{}, unavailable() }
func (*SP3) PredictionSummary() (SP3PredictionSummary, error) {
	return SP3PredictionSummary{}, unavailable()
}

type SPPObservation struct {
	SatelliteID  string
	PseudorangeM float64
}

type SPPConfig struct {
	Observations    []SPPObservation
	TRxJ2000S       float64
	TRxSecondOfDayS float64
	DayOfYear       float64
	InitialGuess    [4]float64
	Ionosphere      bool
	Troposphere     bool
	WithGeodetic    bool
}

type SPPGeometryQuality struct{}
type SPPMetadata struct{}
type SPPSolution struct{}

func (*SP3) Solve(SPPConfig) (SPPSolution, error) { return SPPSolution{}, unavailable() }

type TLEMetadata struct{}
type TLELines struct{}
type TEMEState struct{}
type TLE struct{}

func ParseTLE(string, string) (*TLE, error) { return nil, unavailable() }
func (*TLE) Close() error                   { return nil }
func (*TLE) Metadata() (TLEMetadata, error) { return TLEMetadata{}, unavailable() }
func (*TLE) Lines() (TLELines, error)       { return TLELines{}, unavailable() }
func (*TLE) Propagate([]time.Time) ([]TEMEState, error) {
	return nil, unavailable()
}
