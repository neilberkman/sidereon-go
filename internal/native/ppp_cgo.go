//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

const (
	PppTropoMappingNiellValue   = uint32(C.SIDEREON_PPP_TROPO_MAPPING_NIELL)
	PppTropoMappingVMF1Value    = uint32(C.SIDEREON_PPP_TROPO_MAPPING_VMF1)
	PppSolveStateToleranceValue = uint32(C.SIDEREON_PPP_SOLVE_STATUS_STATE_TOLERANCE)
	PppSolveMaxIterationsValue  = uint32(C.SIDEREON_PPP_SOLVE_STATUS_MAX_ITERATIONS)
	PppIntegerFixedValue        = uint32(C.SIDEREON_PPP_INTEGER_STATUS_FIXED)
	PppIntegerNotFixedValue     = uint32(C.SIDEREON_PPP_INTEGER_STATUS_NOT_FIXED)
)

type PppAutoInitOptions struct {
	HasInitialGuess      bool
	InitialGuessPosition [3]float64
	InitialGuessClockM   float64
	SPPInitialGuess      [4]float64
	SPPTroposphere       bool
	SPPPressureHPA       float64
	SPPTemperatureK      float64
	SPPRelativeHumidity  float64
}

type PppCorrectionObservation struct {
	SatelliteID       string
	Frequency1Hz      float64
	Frequency2Hz      float64
	HasGLONASSChannel bool
	GLONASSChannel    int8
}

type PppCorrectionEpoch struct {
	Epoch        CivilDateTime
	TRxJ2000S    float64
	Observations []PppCorrectionObservation
}

type PppPoleTideOptions struct {
	XPArcsec float64
	YPArcsec float64
}

type PppCodeBiasSystemPair struct {
	System uint32
	Obs1   string
	Obs2   string
}

type PppCodeBiasSatellitePair struct {
	SatelliteID string
	Obs1        string
	Obs2        string
}

type PppNoaziPcvSample struct {
	A float64
	B float64
}

type PppSatelliteAntennaFrequency struct {
	Label    string
	PCOM     [3]float64
	NoaziPCV []PppNoaziPcvSample
}

type PppSatelliteAntenna struct {
	SatelliteID   string
	HasValidFrom  bool
	ValidFrom     CivilDateTime
	HasValidUntil bool
	ValidUntil    CivilDateTime
	Frequencies   []PppSatelliteAntennaFrequency
}

type PppSatelliteAntennaOptions struct {
	Frequency1Label string
	Frequency1Hz    float64
	Frequency2Label string
	Frequency2Hz    float64
	Antennas        []PppSatelliteAntenna
}

type PppCorrectionsOptions struct {
	SolidEarthTide            bool
	PoleTide                  *PppPoleTideOptions
	OceanLoading              *OceanLoadingBLQ
	PhaseWindup               bool
	SatelliteAntenna          *PppSatelliteAntennaOptions
	CodeBias                  *BiasSet
	CodeBiasSystemPairs       []PppCodeBiasSystemPair
	CodeBiasSatellitePairs    []PppCodeBiasSatellitePair
	CodeBiasClockReference    []PppCodeBiasSystemPair
	HasCodeBiasClockReference bool
}

type PppObservation struct {
	SatelliteID  string
	AmbiguityID  string
	CodeM        float64
	PhaseM       float64
	Frequency1Hz float64
	Frequency2Hz float64
}

type PppEpoch struct {
	Civil        CivilDateTime
	JDWhole      float64
	JDFraction   float64
	TRxJ2000S    float64
	Observations []PppObservation
}

type PppFloatMapEntry struct {
	ID    string
	Value float64
}

type PppFloatState struct {
	PositionM          [3]float64
	ClocksM            []float64
	AmbiguitiesM       []PppFloatMapEntry
	ZTDm               float64
	TropoGradientNorth float64
	TropoGradientEast  float64
}

type PppMeasurementWeights struct {
	Code               float64
	Phase              float64
	ElevationWeighting bool
}

type PppVmfSiteSample struct {
	MJD float64
	AH  float64
	AW  float64
}

type PppTroposphereOptions struct {
	Enabled               bool
	EstimateZTD           bool
	EstimateTropoGradient bool
	PressureHPA           float64
	TemperatureK          float64
	RelativeHumidity      float64
	Mapping               uint32
	VMFSamples            []PppVmfSiteSample
}

type PppReceiverAntennaNoaziPcvSample struct {
	ZenithDeg float64
	ValueM    float64
}

type PppReceiverAntennaAzimuthPcvSample struct {
	AzimuthDeg float64
	ZenithDeg  float64
	ValueM     float64
}

type PppReceiverAntennaCalibration struct {
	PCONEUM     [3]float64
	NoaziPCVM   []PppReceiverAntennaNoaziPcvSample
	AzimuthPCVM []PppReceiverAntennaAzimuthPcvSample
}

type PppReceiverAntennaOptions struct {
	Frequency1Label string
	Frequency1Hz    float64
	Frequency1      PppReceiverAntennaCalibration
	Frequency2Label string
	Frequency2Hz    float64
	Frequency2      PppReceiverAntennaCalibration
}

type PppSatelliteClockRecord struct {
	SatelliteID string
	GPSSeconds  float64
	ClockS      float64
}

type PppRangeCorrections struct {
	ReceiverAntenna       *PppReceiverAntennaOptions
	SatClockRelativity    bool
	SatelliteClockRecords []PppSatelliteClockRecord
	SolidEarthTide        bool
	PhaseWindup           bool
	SatelliteAntenna      bool
}

type PppFloatOptions struct {
	MaxIterations       int
	PositionToleranceM  float64
	ClockToleranceM     float64
	AmbiguityToleranceM float64
	ZTDToleranceM       float64
}

type PppFloatConfig struct {
	Epochs             []PppEpoch
	InitialState       PppFloatState
	Weights            PppMeasurementWeights
	Troposphere        PppTroposphereOptions
	Corrections        PppRangeCorrections
	Options            PppFloatOptions
	HasElevationCutoff bool
	ElevationCutoffDeg float64
	ResidualScreen     bool
}

type PppFixedAmbiguityOptions struct {
	WavelengthsM   []PppFloatMapEntry
	OffsetsM       []PppFloatMapEntry
	RatioThreshold float64
}

type PppFixedConfig struct {
	Epochs             []PppEpoch
	Weights            PppMeasurementWeights
	Troposphere        PppTroposphereOptions
	Corrections        PppRangeCorrections
	Options            PppFloatOptions
	HasElevationCutoff bool
	ElevationCutoffDeg float64
	Ambiguity          PppFixedAmbiguityOptions
}

type PppSatScalarCorrection struct {
	SatelliteID string
	EpochIndex  int
	Value       float64
}

type PppSatVectorCorrection struct {
	SatelliteID string
	EpochIndex  int
	ValueM      [3]float64
}

type PppEpochVectorCorrection struct {
	EpochIndex int
	ValueM     [3]float64
}

type PppFloatAmbiguity struct {
	ID     string
	ValueM float64
}

type PppFixedAmbiguity struct {
	ID     string
	Cycles int64
	ValueM float64
}

type PppPositionCovariance struct {
	ECEFM2 [3][3]float64
	ENUM2  [3][3]float64
}

type PppPositionCovariances struct {
	Posterior      PppPositionCovariance
	Formal         PppPositionCovariance
	Temporal       PppPositionCovariance
	PosteriorScale float64
	TemporalScale  float64
}

type PppTemporalCorrelation struct {
	HasLag1            bool
	Lag1               float64
	HasDecorrelation   bool
	DecorrelationTimeS float64
	NominalSamples     int
	EffectiveSamples   float64
	VarianceInflation  float64
	ArcsUsed           int
}

type PppTropoGradient struct {
	HasGradient         bool
	NorthM              float64
	EastM               float64
	HasCovariance       bool
	CovarianceM2        [2][2]float64
	HasFormalCovariance bool
	FormalCovarianceM2  [2][2]float64
}

type PppFloatMetadata struct {
	Iterations     int
	Converged      bool
	Status         uint32
	HasZTDResidual bool
	ZTDResidualM   float64
	CodeRMSM       float64
	PhaseRMSM      float64
	WeightedRMSM   float64
	AmbiguityCount int
	ResidualCount  int
	UsedSatCount   int
}

type PppFixedMetadata struct {
	Iterations             int
	Converged              bool
	Status                 uint32
	HasZTDResidual         bool
	ZTDResidualM           float64
	CodeRMSM               float64
	PhaseRMSM              float64
	WeightedRMSM           float64
	FixedAmbiguityCount    int
	ResidualCount          int
	UsedSatCount           int
	IntegerStatus          uint32
	IntegerRatio           float64
	IntegerBestScore       float64
	HasIntegerSecondBest   bool
	IntegerSecondBestScore float64
	IntegerCandidates      int
}

type pppArena struct {
	pointers []unsafe.Pointer
}

func (a *pppArena) malloc(count int, elementSize uintptr, label string) (unsafe.Pointer, error) {
	size, err := checkedNativeAllocationSize(count, elementSize)
	if err != nil {
		return nil, fmt.Errorf("sidereon: %s: %w", label, err)
	}
	if size == 0 {
		return nil, nil
	}
	pointer := C.calloc(1, C.size_t(size))
	if pointer == nil {
		return nil, fmt.Errorf("sidereon: unable to allocate native %s", label)
	}
	a.pointers = append(a.pointers, pointer)
	return pointer, nil
}

func (a *pppArena) cstring(value, label string) (*C.char, error) {
	pointer, err := copyNativeCString(value, label)
	if err != nil {
		return nil, err
	}
	if pointer == nil {
		return nil, fmt.Errorf("sidereon: unable to allocate native %s", label)
	}
	a.pointers = append(a.pointers, pointer)
	return (*C.char)(pointer), nil
}

func (a *pppArena) close() {
	for i := len(a.pointers) - 1; i >= 0; i-- {
		C.free(a.pointers[i])
	}
	a.pointers = nil
}

func checkedPppSum(total, count int, label string) (int, error) {
	if total < 0 || count < 0 || count > int(^uint(0)>>1)-total {
		return 0, fmt.Errorf("sidereon: %s count overflows int", label)
	}
	return total + count, nil
}

func cPppCivil(value CivilDateTime) (C.SidereonPppCivilDateTime, error) {
	year, err := checkedInt32(value.Year, "PPP civil year")
	if err != nil {
		return C.SidereonPppCivilDateTime{}, err
	}
	month, err := checkedUint8(value.Month, "PPP civil month")
	if err != nil {
		return C.SidereonPppCivilDateTime{}, err
	}
	day, err := checkedUint8(value.Day, "PPP civil day")
	if err != nil {
		return C.SidereonPppCivilDateTime{}, err
	}
	hour, err := checkedUint8(value.Hour, "PPP civil hour")
	if err != nil {
		return C.SidereonPppCivilDateTime{}, err
	}
	minute, err := checkedUint8(value.Minute, "PPP civil minute")
	if err != nil {
		return C.SidereonPppCivilDateTime{}, err
	}
	return C.SidereonPppCivilDateTime{year: C.int32_t(year), month: C.uint8_t(month), day: C.uint8_t(day), hour: C.uint8_t(hour), minute: C.uint8_t(minute), second: C.double(value.Second)}, nil
}

func cCorrectionCivil(value CivilDateTime) (C.SidereonCivilDateTime, error) {
	year, err := checkedInt32(value.Year, "PPP correction civil year")
	if err != nil {
		return C.SidereonCivilDateTime{}, err
	}
	month, err := checkedUint8(value.Month, "PPP correction civil month")
	if err != nil {
		return C.SidereonCivilDateTime{}, err
	}
	day, err := checkedUint8(value.Day, "PPP correction civil day")
	if err != nil {
		return C.SidereonCivilDateTime{}, err
	}
	hour, err := checkedUint8(value.Hour, "PPP correction civil hour")
	if err != nil {
		return C.SidereonCivilDateTime{}, err
	}
	minute, err := checkedUint8(value.Minute, "PPP correction civil minute")
	if err != nil {
		return C.SidereonCivilDateTime{}, err
	}
	return C.SidereonCivilDateTime{year: C.int32_t(year), month: C.uint8_t(month), day: C.uint8_t(day), hour: C.uint8_t(hour), minute: C.uint8_t(minute), second: C.double(value.Second)}, nil
}

func cPppCorrectionEpochs(values []PppCorrectionEpoch, arena *pppArena) (*C.SidereonPppCorrectionEpoch, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	var total int
	for _, value := range values {
		total, err = checkedPppSum(total, len(value.Observations), "PPP correction observations")
		if err != nil {
			return nil, 0, err
		}
	}
	epochMemory, err := arena.malloc(len(values), unsafe.Sizeof(C.SidereonPppCorrectionEpoch{}), "PPP correction epochs")
	if err != nil {
		return nil, 0, err
	}
	obsMemory, err := arena.malloc(total, unsafe.Sizeof(C.SidereonPppCorrectionObservation{}), "PPP correction observations")
	if err != nil {
		return nil, 0, err
	}
	epochs := unsafe.Slice((*C.SidereonPppCorrectionEpoch)(epochMemory), len(values))
	observations := unsafe.Slice((*C.SidereonPppCorrectionObservation)(obsMemory), total)
	index := 0
	for i, value := range values {
		civil, civilErr := cCorrectionCivil(value.Epoch)
		if civilErr != nil {
			return nil, 0, civilErr
		}
		epochs[i].epoch = civil
		epochs[i].t_rx_j2000_s = C.double(value.TRxJ2000S)
		observationCount, countErr := checkedNativeSize(len(value.Observations))
		if countErr != nil {
			return nil, 0, countErr
		}
		epochs[i].observation_count = observationCount
		if len(value.Observations) != 0 {
			epochs[i].observations = &observations[index]
		}
		for j, observation := range value.Observations {
			satID, stringErr := arena.cstring(observation.SatelliteID, "PPP correction satellite ID")
			if stringErr != nil {
				return nil, 0, stringErr
			}
			observations[index+j].sat_id = satID
			observations[index+j].freq1_hz = C.double(observation.Frequency1Hz)
			observations[index+j].freq2_hz = C.double(observation.Frequency2Hz)
			observations[index+j].has_glonass_channel = C.bool(observation.HasGLONASSChannel)
			observations[index+j].glonass_channel = C.int8_t(observation.GLONASSChannel)
		}
		index += len(value.Observations)
	}
	return (*C.SidereonPppCorrectionEpoch)(epochMemory), count, nil
}

func cPppMapEntries(values []PppFloatMapEntry, arena *pppArena, label string) (*C.SidereonPppFloatMapEntry, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	memory, err := arena.malloc(len(values), unsafe.Sizeof(C.SidereonPppFloatMapEntry{}), label)
	if err != nil {
		return nil, 0, err
	}
	entries := unsafe.Slice((*C.SidereonPppFloatMapEntry)(memory), len(values))
	for i, value := range values {
		id, stringErr := arena.cstring(value.ID, label+" ID")
		if stringErr != nil {
			return nil, 0, stringErr
		}
		entries[i].id = id
		entries[i].value = C.double(value.Value)
	}
	return (*C.SidereonPppFloatMapEntry)(memory), count, nil
}

func cPppEpochs(values []PppEpoch, arena *pppArena) (*C.SidereonPppEpoch, C.size_t, error) {
	count, err := checkedNativeSize(len(values))
	if err != nil {
		return nil, 0, err
	}
	var total int
	for _, value := range values {
		total, err = checkedPppSum(total, len(value.Observations), "PPP observations")
		if err != nil {
			return nil, 0, err
		}
	}
	epochMemory, err := arena.malloc(len(values), unsafe.Sizeof(C.SidereonPppEpoch{}), "PPP epochs")
	if err != nil {
		return nil, 0, err
	}
	obsMemory, err := arena.malloc(total, unsafe.Sizeof(C.SidereonPppObservation{}), "PPP observations")
	if err != nil {
		return nil, 0, err
	}
	epochs := unsafe.Slice((*C.SidereonPppEpoch)(epochMemory), len(values))
	observations := unsafe.Slice((*C.SidereonPppObservation)(obsMemory), total)
	index := 0
	for i, value := range values {
		civil, civilErr := cPppCivil(value.Civil)
		if civilErr != nil {
			return nil, 0, civilErr
		}
		epochs[i].civil = civil
		epochs[i].jd_whole = C.double(value.JDWhole)
		epochs[i].jd_fraction = C.double(value.JDFraction)
		epochs[i].t_rx_j2000_s = C.double(value.TRxJ2000S)
		observationCount, countErr := checkedNativeSize(len(value.Observations))
		if countErr != nil {
			return nil, 0, countErr
		}
		epochs[i].observation_count = observationCount
		if len(value.Observations) != 0 {
			epochs[i].observations = &observations[index]
		}
		for j, observation := range value.Observations {
			satID, stringErr := arena.cstring(observation.SatelliteID, "PPP satellite ID")
			if stringErr != nil {
				return nil, 0, stringErr
			}
			ambiguityID, stringErr := arena.cstring(observation.AmbiguityID, "PPP ambiguity ID")
			if stringErr != nil {
				return nil, 0, stringErr
			}
			observations[index+j].sat_id = satID
			observations[index+j].ambiguity_id = ambiguityID
			observations[index+j].code_m = C.double(observation.CodeM)
			observations[index+j].phase_m = C.double(observation.PhaseM)
			observations[index+j].freq1_hz = C.double(observation.Frequency1Hz)
			observations[index+j].freq2_hz = C.double(observation.Frequency2Hz)
		}
		index += len(value.Observations)
	}
	return (*C.SidereonPppEpoch)(epochMemory), count, nil
}

func cPppSatelliteAntennaOptions(value PppSatelliteAntennaOptions, arena *pppArena) (*C.SidereonSatelliteAntennaOptions, error) {
	memory, err := arena.malloc(1, unsafe.Sizeof(C.SidereonSatelliteAntennaOptions{}), "PPP satellite antenna options")
	if err != nil {
		return nil, err
	}
	out := (*C.SidereonSatelliteAntennaOptions)(memory)
	frequency1Label, err := arena.cstring(value.Frequency1Label, "PPP satellite antenna frequency-1 label")
	if err != nil {
		return nil, err
	}
	frequency2Label, err := arena.cstring(value.Frequency2Label, "PPP satellite antenna frequency-2 label")
	if err != nil {
		return nil, err
	}
	out.freq1_label, out.freq1_hz = frequency1Label, C.double(value.Frequency1Hz)
	out.freq2_label, out.freq2_hz = frequency2Label, C.double(value.Frequency2Hz)
	count, err := checkedNativeSize(len(value.Antennas))
	if err != nil {
		return nil, err
	}
	antennaMemory, err := arena.malloc(len(value.Antennas), unsafe.Sizeof(C.SidereonSatelliteAntenna{}), "PPP satellite antennas")
	if err != nil {
		return nil, err
	}
	out.antennas, out.antenna_count = (*C.SidereonSatelliteAntenna)(antennaMemory), count
	antennas := unsafe.Slice((*C.SidereonSatelliteAntenna)(antennaMemory), len(value.Antennas))
	for i, antenna := range value.Antennas {
		satID, stringErr := arena.cstring(antenna.SatelliteID, "PPP antenna satellite ID")
		if stringErr != nil {
			return nil, stringErr
		}
		antennas[i].sat_id = satID
		antennas[i].has_valid_from = C.bool(antenna.HasValidFrom)
		if antenna.HasValidFrom {
			civil, civilErr := cCorrectionCivil(antenna.ValidFrom)
			if civilErr != nil {
				return nil, civilErr
			}
			antennas[i].valid_from = civil
		}
		antennas[i].has_valid_until = C.bool(antenna.HasValidUntil)
		if antenna.HasValidUntil {
			civil, civilErr := cCorrectionCivil(antenna.ValidUntil)
			if civilErr != nil {
				return nil, civilErr
			}
			antennas[i].valid_until = civil
		}
		frequencyCount, countErr := checkedNativeSize(len(antenna.Frequencies))
		if countErr != nil {
			return nil, countErr
		}
		frequencyMemory, allocationErr := arena.malloc(len(antenna.Frequencies), unsafe.Sizeof(C.SidereonSatelliteAntennaFrequency{}), "PPP antenna frequencies")
		if allocationErr != nil {
			return nil, allocationErr
		}
		antennas[i].frequencies, antennas[i].frequency_count = (*C.SidereonSatelliteAntennaFrequency)(frequencyMemory), frequencyCount
		frequencies := unsafe.Slice((*C.SidereonSatelliteAntennaFrequency)(frequencyMemory), len(antenna.Frequencies))
		for j, frequency := range antenna.Frequencies {
			label, labelErr := arena.cstring(frequency.Label, "PPP antenna frequency label")
			if labelErr != nil {
				return nil, labelErr
			}
			frequencies[j].label = label
			for axis := range frequency.PCOM {
				frequencies[j].pco_m[axis] = C.double(frequency.PCOM[axis])
			}
			noaziCount, noaziErr := checkedNativeSize(len(frequency.NoaziPCV))
			if noaziErr != nil {
				return nil, noaziErr
			}
			noaziMemory, noaziAllocErr := arena.malloc(len(frequency.NoaziPCV), unsafe.Sizeof(C.SidereonNoaziPcvSample{}), "PPP antenna PCV samples")
			if noaziAllocErr != nil {
				return nil, noaziAllocErr
			}
			frequencies[j].noazi_pcv, frequencies[j].noazi_count = (*C.SidereonNoaziPcvSample)(noaziMemory), noaziCount
			noazi := unsafe.Slice((*C.SidereonNoaziPcvSample)(noaziMemory), len(frequency.NoaziPCV))
			for k, sample := range frequency.NoaziPCV {
				noazi[k].a, noazi[k].b = C.double(sample.A), C.double(sample.B)
			}
		}
	}
	return out, nil
}

func cPppReceiverAntennaOptions(value PppReceiverAntennaOptions, arena *pppArena) (*C.SidereonPppReceiverAntennaOptions, error) {
	memory, err := arena.malloc(1, unsafe.Sizeof(C.SidereonPppReceiverAntennaOptions{}), "PPP receiver antenna options")
	if err != nil {
		return nil, err
	}
	out := (*C.SidereonPppReceiverAntennaOptions)(memory)
	frequency1Label, err := arena.cstring(value.Frequency1Label, "PPP receiver antenna frequency-1 label")
	if err != nil {
		return nil, err
	}
	frequency2Label, err := arena.cstring(value.Frequency2Label, "PPP receiver antenna frequency-2 label")
	if err != nil {
		return nil, err
	}
	out.freq1_label, out.freq1_hz = frequency1Label, C.double(value.Frequency1Hz)
	out.freq2_label, out.freq2_hz = frequency2Label, C.double(value.Frequency2Hz)
	for _, spec := range []struct {
		value PppReceiverAntennaCalibration
		out   *C.SidereonReceiverAntennaCalibration
	}{
		{value.Frequency1, &out.freq1},
		{value.Frequency2, &out.freq2},
	} {
		for axis := range spec.value.PCONEUM {
			spec.out.pco_neu_m[axis] = C.double(spec.value.PCONEUM[axis])
		}
		noaziCount, countErr := checkedNativeSize(len(spec.value.NoaziPCVM))
		if countErr != nil {
			return nil, countErr
		}
		noaziMemory, allocationErr := arena.malloc(len(spec.value.NoaziPCVM), unsafe.Sizeof(C.SidereonReceiverAntennaNoaziPcvSample{}), "PPP receiver no-azimuth PCV samples")
		if allocationErr != nil {
			return nil, allocationErr
		}
		spec.out.noazi_pcv_m, spec.out.noazi_pcv_count = (*C.SidereonReceiverAntennaNoaziPcvSample)(noaziMemory), noaziCount
		noazi := unsafe.Slice((*C.SidereonReceiverAntennaNoaziPcvSample)(noaziMemory), len(spec.value.NoaziPCVM))
		for i, sample := range spec.value.NoaziPCVM {
			noazi[i].zenith_deg, noazi[i].value_m = C.double(sample.ZenithDeg), C.double(sample.ValueM)
		}
		azimuthCount, countErr := checkedNativeSize(len(spec.value.AzimuthPCVM))
		if countErr != nil {
			return nil, countErr
		}
		azimuthMemory, allocationErr := arena.malloc(len(spec.value.AzimuthPCVM), unsafe.Sizeof(C.SidereonReceiverAntennaAzimuthPcvSample{}), "PPP receiver azimuth PCV samples")
		if allocationErr != nil {
			return nil, allocationErr
		}
		spec.out.azimuth_pcv_m, spec.out.azimuth_pcv_count = (*C.SidereonReceiverAntennaAzimuthPcvSample)(azimuthMemory), azimuthCount
		azimuth := unsafe.Slice((*C.SidereonReceiverAntennaAzimuthPcvSample)(azimuthMemory), len(spec.value.AzimuthPCVM))
		for i, sample := range spec.value.AzimuthPCVM {
			azimuth[i].azimuth_deg, azimuth[i].zenith_deg, azimuth[i].value_m = C.double(sample.AzimuthDeg), C.double(sample.ZenithDeg), C.double(sample.ValueM)
		}
	}
	return out, nil
}

func cPppRangeCorrections(value PppRangeCorrections, arena *pppArena) (C.SidereonPppRangeCorrections, error) {
	var out C.SidereonPppRangeCorrections
	out.sat_clock_relativity = C.bool(value.SatClockRelativity)
	out.solid_earth_tide = C.bool(value.SolidEarthTide)
	out.phase_windup = C.bool(value.PhaseWindup)
	out.satellite_antenna = C.bool(value.SatelliteAntenna)
	if value.ReceiverAntenna != nil {
		antenna, err := cPppReceiverAntennaOptions(*value.ReceiverAntenna, arena)
		if err != nil {
			return out, err
		}
		out.receiver_antenna = antenna
	}
	count, err := checkedNativeSize(len(value.SatelliteClockRecords))
	if err != nil {
		return out, err
	}
	memory, err := arena.malloc(len(value.SatelliteClockRecords), unsafe.Sizeof(C.SidereonPppSatelliteClockRecord{}), "PPP satellite clock records")
	if err != nil {
		return out, err
	}
	out.satellite_clock_records, out.satellite_clock_record_count = (*C.SidereonPppSatelliteClockRecord)(memory), count
	records := unsafe.Slice((*C.SidereonPppSatelliteClockRecord)(memory), len(value.SatelliteClockRecords))
	for i, record := range value.SatelliteClockRecords {
		satID, stringErr := arena.cstring(record.SatelliteID, "PPP satellite clock ID")
		if stringErr != nil {
			return out, stringErr
		}
		records[i].sat_id, records[i].gps_seconds, records[i].clock_s = satID, C.double(record.GPSSeconds), C.double(record.ClockS)
	}
	return out, nil
}

func cPppTroposphere(value PppTroposphereOptions) (C.SidereonPppTroposphereOptions, error) {
	var out C.SidereonPppTroposphereOptions
	if value.Mapping != PppTropoMappingNiellValue && value.Mapping != PppTropoMappingVMF1Value {
		return out, invalidArgument("invalid PPP troposphere mapping")
	}
	if len(value.VMFSamples) > int(C.SIDEREON_PPP_VMF_SITE_MAX_SAMPLES) {
		return out, invalidArgument("too many PPP VMF samples")
	}
	out.enabled = C.bool(value.Enabled)
	out.estimate_ztd = C.bool(value.EstimateZTD)
	out.estimate_tropo_gradients = C.bool(value.EstimateTropoGradient)
	out.pressure_hpa, out.temperature_k, out.relative_humidity = C.double(value.PressureHPA), C.double(value.TemperatureK), C.double(value.RelativeHumidity)
	out.mapping, out.vmf_sample_count = C.uint32_t(value.Mapping), C.size_t(len(value.VMFSamples))
	for i, sample := range value.VMFSamples {
		out.vmf_samples[i].mjd, out.vmf_samples[i].ah, out.vmf_samples[i].aw = C.double(sample.MJD), C.double(sample.AH), C.double(sample.AW)
	}
	return out, nil
}

func cPppCorrectionsOptions(value PppCorrectionsOptions, arena *pppArena, bias *C.SidereonBiasSet) (C.SidereonPppCorrectionsOptions, error) {
	var out C.SidereonPppCorrectionsOptions
	out.solid_earth_tide = C.bool(value.SolidEarthTide)
	out.phase_windup = C.bool(value.PhaseWindup)
	if value.PoleTide != nil {
		out.has_pole_tide = true
		out.pole_tide.xp_arcsec, out.pole_tide.yp_arcsec = C.double(value.PoleTide.XPArcsec), C.double(value.PoleTide.YPArcsec)
	}
	if value.OceanLoading != nil {
		out.has_ocean_loading = true
		for i := range value.OceanLoading.AmplitudeM {
			for j := range value.OceanLoading.AmplitudeM[i] {
				out.ocean_loading.amplitude_m[i][j] = C.double(value.OceanLoading.AmplitudeM[i][j])
				out.ocean_loading.phase_deg[i][j] = C.double(value.OceanLoading.PhaseDeg[i][j])
			}
		}
	}
	if value.SatelliteAntenna != nil {
		antenna, err := cPppSatelliteAntennaOptions(*value.SatelliteAntenna, arena)
		if err != nil {
			return out, err
		}
		out.has_satellite_antenna, out.satellite_antenna = true, antenna
	}
	if value.CodeBias != nil {
		if bias == nil {
			return out, ErrClosed
		}
		out.has_code_bias, out.code_bias = true, bias
	}
	if err := cPppCodeBiasPairs(value.CodeBiasSystemPairs, &out.code_bias_system_pairs, &out.code_bias_system_pair_count, arena); err != nil {
		return out, err
	}
	if err := cPppCodeBiasSatellitePairs(value.CodeBiasSatellitePairs, &out.code_bias_satellite_pairs, &out.code_bias_satellite_pair_count, arena); err != nil {
		return out, err
	}
	if value.HasCodeBiasClockReference {
		out.has_code_bias_clock_reference = true
	}
	if err := cPppCodeBiasPairs(value.CodeBiasClockReference, &out.code_bias_clock_reference_pairs, &out.code_bias_clock_reference_pair_count, arena); err != nil {
		return out, err
	}
	return out, nil
}

func cPppCodeBiasPairs(values []PppCodeBiasSystemPair, output **C.SidereonPppCodeBiasSystemPair, count *C.size_t, arena *pppArena) error {
	nativeCount, err := checkedNativeSize(len(values))
	if err != nil {
		return err
	}
	memory, err := arena.malloc(len(values), unsafe.Sizeof(C.SidereonPppCodeBiasSystemPair{}), "PPP code-bias system pairs")
	if err != nil {
		return err
	}
	*output, *count = (*C.SidereonPppCodeBiasSystemPair)(memory), nativeCount
	pairs := unsafe.Slice((*C.SidereonPppCodeBiasSystemPair)(memory), len(values))
	for i, value := range values {
		obs1, stringErr := arena.cstring(value.Obs1, "PPP code-bias first observable")
		if stringErr != nil {
			return stringErr
		}
		obs2, stringErr := arena.cstring(value.Obs2, "PPP code-bias second observable")
		if stringErr != nil {
			return stringErr
		}
		pairs[i].system, pairs[i].obs1, pairs[i].obs2 = C.uint32_t(value.System), obs1, obs2
	}
	return nil
}

func cPppCodeBiasSatellitePairs(values []PppCodeBiasSatellitePair, output **C.SidereonPppCodeBiasSatellitePair, count *C.size_t, arena *pppArena) error {
	nativeCount, err := checkedNativeSize(len(values))
	if err != nil {
		return err
	}
	memory, err := arena.malloc(len(values), unsafe.Sizeof(C.SidereonPppCodeBiasSatellitePair{}), "PPP satellite code-bias pairs")
	if err != nil {
		return err
	}
	*output, *count = (*C.SidereonPppCodeBiasSatellitePair)(memory), nativeCount
	pairs := unsafe.Slice((*C.SidereonPppCodeBiasSatellitePair)(memory), len(values))
	for i, value := range values {
		satID, stringErr := arena.cstring(value.SatelliteID, "PPP code-bias satellite ID")
		if stringErr != nil {
			return stringErr
		}
		obs1, stringErr := arena.cstring(value.Obs1, "PPP satellite code-bias first observable")
		if stringErr != nil {
			return stringErr
		}
		obs2, stringErr := arena.cstring(value.Obs2, "PPP satellite code-bias second observable")
		if stringErr != nil {
			return stringErr
		}
		pairs[i].sat_id, pairs[i].obs1, pairs[i].obs2 = satID, obs1, obs2
	}
	return nil
}

func cPppState(value PppFloatState, arena *pppArena, epochCount int) (C.SidereonPppFloatState, error) {
	var out C.SidereonPppFloatState
	if len(value.ClocksM) != epochCount {
		return out, invalidArgument("PPP initial clock count must equal epoch count")
	}
	clockCount, err := checkedNativeSize(len(value.ClocksM))
	if err != nil {
		return out, err
	}
	clockMemory, err := arena.malloc(len(value.ClocksM), unsafe.Sizeof(C.double(0)), "PPP initial clocks")
	if err != nil {
		return out, err
	}
	clocks := unsafe.Slice((*C.double)(clockMemory), len(value.ClocksM))
	for i, clock := range value.ClocksM {
		clocks[i] = C.double(clock)
	}
	ambiguityMemory, ambiguityCount, err := cPppMapEntries(value.AmbiguitiesM, arena, "PPP initial ambiguities")
	if err != nil {
		return out, err
	}
	for axis := range value.PositionM {
		out.position_m[axis] = C.double(value.PositionM[axis])
	}
	out.clocks_m, out.clock_count = (*C.double)(clockMemory), clockCount
	out.ambiguities_m, out.ambiguity_count = ambiguityMemory, ambiguityCount
	out.ztd_m, out.tropo_gradient_north_m, out.tropo_gradient_east_m = C.double(value.ZTDm), C.double(value.TropoGradientNorth), C.double(value.TropoGradientEast)
	return out, nil
}

func cPppFloatConfig(value PppFloatConfig, arena *pppArena, includeInitialState bool) (C.SidereonPppFloatConfig, error) {
	var out C.SidereonPppFloatConfig
	epochs, epochCount, err := cPppEpochs(value.Epochs, arena)
	if err != nil {
		return out, err
	}
	var state C.SidereonPppFloatState
	if includeInitialState {
		state, err = cPppState(value.InitialState, arena, len(value.Epochs))
		if err != nil {
			return out, err
		}
	}
	weights := C.SidereonPppMeasurementWeights{code: C.double(value.Weights.Code), phase: C.double(value.Weights.Phase), elevation_weighting: C.bool(value.Weights.ElevationWeighting)}
	tropo, err := cPppTroposphere(value.Troposphere)
	if err != nil {
		return out, err
	}
	corrections, err := cPppRangeCorrections(value.Corrections, arena)
	if err != nil {
		return out, err
	}
	maxIterations, err := checkedNativeSize(value.Options.MaxIterations)
	if err != nil {
		return out, err
	}
	options := C.SidereonPppFloatOptions{max_iterations: maxIterations, position_tolerance_m: C.double(value.Options.PositionToleranceM), clock_tolerance_m: C.double(value.Options.ClockToleranceM), ambiguity_tolerance_m: C.double(value.Options.AmbiguityToleranceM), ztd_tolerance_m: C.double(value.Options.ZTDToleranceM)}
	out.epochs, out.epoch_count = epochs, epochCount
	out.initial_state, out.weights, out.tropo, out.corrections, out.options = state, weights, tropo, corrections, options
	out.has_elevation_cutoff_deg, out.elevation_cutoff_deg, out.residual_screen = C.bool(value.HasElevationCutoff), C.double(value.ElevationCutoffDeg), C.bool(value.ResidualScreen)
	return out, nil
}

func cPppFixedConfig(value PppFixedConfig, arena *pppArena) (C.SidereonPppFixedConfig, error) {
	var out C.SidereonPppFixedConfig
	epochs, epochCount, err := cPppEpochs(value.Epochs, arena)
	if err != nil {
		return out, err
	}
	weights := C.SidereonPppMeasurementWeights{code: C.double(value.Weights.Code), phase: C.double(value.Weights.Phase), elevation_weighting: C.bool(value.Weights.ElevationWeighting)}
	tropo, err := cPppTroposphere(value.Troposphere)
	if err != nil {
		return out, err
	}
	corrections, err := cPppRangeCorrections(value.Corrections, arena)
	if err != nil {
		return out, err
	}
	maxIterations, err := checkedNativeSize(value.Options.MaxIterations)
	if err != nil {
		return out, err
	}
	wavelengths, wavelengthCount, err := cPppMapEntries(value.Ambiguity.WavelengthsM, arena, "PPP wavelength map")
	if err != nil {
		return out, err
	}
	offsets, offsetCount, err := cPppMapEntries(value.Ambiguity.OffsetsM, arena, "PPP offset map")
	if err != nil {
		return out, err
	}
	options := C.SidereonPppFloatOptions{max_iterations: maxIterations, position_tolerance_m: C.double(value.Options.PositionToleranceM), clock_tolerance_m: C.double(value.Options.ClockToleranceM), ambiguity_tolerance_m: C.double(value.Options.AmbiguityToleranceM), ztd_tolerance_m: C.double(value.Options.ZTDToleranceM)}
	ambiguity := C.SidereonPppFixedAmbiguityOptions{wavelengths_m: wavelengths, wavelength_count: wavelengthCount, offsets_m: offsets, offset_count: offsetCount, ratio_threshold: C.double(value.Ambiguity.RatioThreshold)}
	out.epochs, out.epoch_count = epochs, epochCount
	out.weights, out.tropo, out.corrections, out.options, out.ambiguity = weights, tropo, corrections, options, ambiguity
	out.has_elevation_cutoff_deg, out.elevation_cutoff_deg = C.bool(value.HasElevationCutoff), C.double(value.ElevationCutoffDeg)
	return out, nil
}

func cPppAutoInit(value PppAutoInitOptions) C.SidereonPppAutoInitOptions {
	var out C.SidereonPppAutoInitOptions
	out.has_initial_guess = C.bool(value.HasInitialGuess)
	for i := range value.InitialGuessPosition {
		out.initial_guess_position_m[i] = C.double(value.InitialGuessPosition[i])
	}
	out.initial_guess_clock_m = C.double(value.InitialGuessClockM)
	for i := range value.SPPInitialGuess {
		out.spp_initial_guess[i] = C.double(value.SPPInitialGuess[i])
	}
	out.spp_troposphere = C.bool(value.SPPTroposphere)
	out.spp_pressure_hpa, out.spp_temperature_k, out.spp_relative_humidity = C.double(value.SPPPressureHPA), C.double(value.SPPTemperatureK), C.double(value.SPPRelativeHumidity)
	return out
}

type PppCorrections struct {
	_      noCopy
	handle *positioningHandle
}

type PppFloatSolution struct {
	_      noCopy
	handle *positioningHandle
}

type PppFixedSolution struct {
	_      noCopy
	handle *positioningHandle
}

func releasePppCorrections(pointer unsafe.Pointer) {
	withCThread(func() { C.sidereon_ppp_corrections_free((*C.SidereonPppCorrections)(pointer)) })
}

func releasePppFloatSolution(pointer unsafe.Pointer) {
	withCThread(func() { C.sidereon_ppp_float_solution_free((*C.SidereonPppFloatSolution)(pointer)) })
}

func releasePppFixedSolution(pointer unsafe.Pointer) {
	withCThread(func() { C.sidereon_ppp_fixed_solution_free((*C.SidereonPppFixedSolution)(pointer)) })
}

func newPppCorrections(pointer *C.SidereonPppCorrections) (*PppCorrections, error) {
	if pointer == nil {
		return nil, missingNativeHandle("PPP corrections")
	}
	return &PppCorrections{handle: newPositioningHandle(unsafe.Pointer(pointer), releasePppCorrections)}, nil
}

func newPppFloatSolution(pointer *C.SidereonPppFloatSolution) (*PppFloatSolution, error) {
	if pointer == nil {
		return nil, missingNativeHandle("PPP float solution")
	}
	return &PppFloatSolution{handle: newPositioningHandle(unsafe.Pointer(pointer), releasePppFloatSolution)}, nil
}

func newPppFixedSolution(pointer *C.SidereonPppFixedSolution) (*PppFixedSolution, error) {
	if pointer == nil {
		return nil, missingNativeHandle("PPP fixed solution")
	}
	return &PppFixedSolution{handle: newPositioningHandle(unsafe.Pointer(pointer), releasePppFixedSolution)}, nil
}

func (c *PppCorrections) Close() error {
	if c == nil || c.handle == nil {
		return nil
	}
	return c.handle.close()
}
func (s *PppFloatSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}
func (s *PppFixedSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return s.handle.close()
}

func PppAutoInitOptionsInit() (PppAutoInitOptions, error) {
	var value C.SidereonPppAutoInitOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_auto_init_options_init(&value)) })
	return PppAutoInitOptions{HasInitialGuess: bool(value.has_initial_guess), InitialGuessPosition: [3]float64{float64(value.initial_guess_position_m[0]), float64(value.initial_guess_position_m[1]), float64(value.initial_guess_position_m[2])}, InitialGuessClockM: float64(value.initial_guess_clock_m), SPPInitialGuess: [4]float64{float64(value.spp_initial_guess[0]), float64(value.spp_initial_guess[1]), float64(value.spp_initial_guess[2]), float64(value.spp_initial_guess[3])}, SPPTroposphere: bool(value.spp_troposphere), SPPPressureHPA: float64(value.spp_pressure_hpa), SPPTemperatureK: float64(value.spp_temperature_k), SPPRelativeHumidity: float64(value.spp_relative_humidity)}, err
}

func PppFloatOptionsInit() (PppFloatOptions, error) {
	var value C.SidereonPppFloatOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_float_options_init(&value)) })
	out := PppFloatOptions{PositionToleranceM: float64(value.position_tolerance_m), ClockToleranceM: float64(value.clock_tolerance_m), AmbiguityToleranceM: float64(value.ambiguity_tolerance_m), ZTDToleranceM: float64(value.ztd_tolerance_m)}
	if err != nil {
		return out, err
	}
	maxIterations, countErr := checkedNativeCount(uint64(value.max_iterations))
	if countErr != nil {
		return out, countErr
	}
	out.MaxIterations = maxIterations
	return out, nil
}

func PppFixedAmbiguityOptionsInit() (PppFixedAmbiguityOptions, error) {
	var value C.SidereonPppFixedAmbiguityOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_fixed_ambiguity_options_init(&value)) })
	if err != nil {
		return PppFixedAmbiguityOptions{}, err
	}
	read := func(pointer *C.SidereonPppFloatMapEntry, count C.size_t) ([]PppFloatMapEntry, error) {
		n, countErr := checkedNativeCount(uint64(count))
		if countErr != nil {
			return nil, countErr
		}
		if n == 0 {
			return nil, nil
		}
		if pointer == nil {
			return nil, invalidArgument("native PPP ambiguity map has a nonzero count and nil pointer")
		}
		values := unsafe.Slice(pointer, n)
		out := make([]PppFloatMapEntry, n)
		for i, item := range values {
			if item.id == nil {
				return nil, invalidArgument("native PPP ambiguity map contains a nil ID")
			}
			out[i] = PppFloatMapEntry{ID: C.GoString(item.id), Value: float64(item.value)}
		}
		return out, nil
	}
	wavelengths, readErr := read(value.wavelengths_m, value.wavelength_count)
	if readErr != nil {
		return PppFixedAmbiguityOptions{}, readErr
	}
	offsets, readErr := read(value.offsets_m, value.offset_count)
	if readErr != nil {
		return PppFixedAmbiguityOptions{}, readErr
	}
	return PppFixedAmbiguityOptions{WavelengthsM: wavelengths, OffsetsM: offsets, RatioThreshold: float64(value.ratio_threshold)}, nil
}

func PppMeasurementWeightsInit() (PppMeasurementWeights, error) {
	var value C.SidereonPppMeasurementWeights
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_measurement_weights_init(&value)) })
	return PppMeasurementWeights{Code: float64(value.code), Phase: float64(value.phase), ElevationWeighting: bool(value.elevation_weighting)}, err
}

func PppRangeCorrectionsInit() (PppRangeCorrections, error) {
	var value C.SidereonPppRangeCorrections
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_range_corrections_init(&value)) })
	return PppRangeCorrections{SatClockRelativity: bool(value.sat_clock_relativity), SolidEarthTide: bool(value.solid_earth_tide), PhaseWindup: bool(value.phase_windup), SatelliteAntenna: bool(value.satellite_antenna)}, err
}

func PppTroposphereOptionsInit() (PppTroposphereOptions, error) {
	var value C.SidereonPppTroposphereOptions
	err := callStatus(func() uint32 { return uint32(C.sidereon_ppp_troposphere_options_init(&value)) })
	out := PppTroposphereOptions{Enabled: bool(value.enabled), EstimateZTD: bool(value.estimate_ztd), EstimateTropoGradient: bool(value.estimate_tropo_gradients), PressureHPA: float64(value.pressure_hpa), TemperatureK: float64(value.temperature_k), RelativeHumidity: float64(value.relative_humidity), Mapping: uint32(value.mapping)}
	if err != nil {
		return out, err
	}
	if out.Mapping != PppTropoMappingNiellValue && out.Mapping != PppTropoMappingVMF1Value {
		return out, invalidArgument("native PPP troposphere options returned an invalid mapping")
	}
	if count, countErr := checkedNativeCount(uint64(value.vmf_sample_count)); countErr != nil {
		return out, countErr
	} else if count <= len(value.vmf_samples) {
		out.VMFSamples = make([]PppVmfSiteSample, count)
		for i := range out.VMFSamples {
			out.VMFSamples[i] = PppVmfSiteSample{MJD: float64(value.vmf_samples[i].mjd), AH: float64(value.vmf_samples[i].ah), AW: float64(value.vmf_samples[i].aw)}
		}
	} else {
		return out, invalidArgument("native PPP VMF sample count exceeds ABI bound")
	}
	return out, nil
}

func PppCorrectionsBuild(sp3 *SP3, epochs []PppCorrectionEpoch, receiver [3]float64, options PppCorrectionsOptions) (*PppCorrections, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	if options.CodeBias != nil && (options.CodeBias.handle == nil) {
		return nil, ErrClosed
	}
	var output *C.SidereonPppCorrections
	build := func(sp3Pointer unsafe.Pointer, biasPointer unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			arena := &pppArena{}
			defer arena.close()
			cEpochs, epochCount, err := cPppCorrectionEpochs(epochs, arena)
			if err != nil {
				operationErr = err
				return
			}
			var bias *C.SidereonBiasSet
			if biasPointer != nil {
				bias = (*C.SidereonBiasSet)(biasPointer)
			}
			cOptions, err := cPppCorrectionsOptions(options, arena, bias)
			if err != nil {
				operationErr = err
				return
			}
			var receiverMemory [3]C.double
			for i := range receiverMemory {
				receiverMemory[i] = C.double(receiver[i])
			}
			status := C.sidereon_ppp_corrections_build(
				(*C.SidereonSp3)(sp3Pointer), cEpochs, epochCount, &receiverMemory[0], &cOptions, &output,
			)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && output != nil {
				C.sidereon_ppp_corrections_free(output)
				output = nil
			}
		})
		return operationErr
	}
	var err error
	if options.CodeBias != nil {
		err = withPositioningHandlePair(sp3.handle, options.CodeBias.handle, build)
	} else {
		err = sp3.handle.with(func(pointer unsafe.Pointer) error { return build(pointer, nil) })
	}
	runtime.KeepAlive(sp3)
	runtime.KeepAlive(options.CodeBias)
	if err != nil {
		return nil, err
	}
	return newPppCorrections(output)
}

func solvePppFloat(sp3 *SP3, config PppFloatConfig, auto *PppAutoInitOptions) (*PppFloatSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var output *C.SidereonPppFloatSolution
	err := sp3.handle.with(func(sp3Pointer unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			arena := &pppArena{}
			defer arena.close()
			cConfig, conversionErr := cPppFloatConfig(config, arena, auto == nil)
			if conversionErr != nil {
				operationErr = conversionErr
				return
			}
			status := C.enum_SidereonStatus(0)
			if auto == nil {
				status = C.sidereon_solve_ppp_float((*C.SidereonSp3)(sp3Pointer), &cConfig, &output)
			} else {
				cAuto := cPppAutoInit(*auto)
				status = C.sidereon_solve_ppp_auto_init_float((*C.SidereonSp3)(sp3Pointer), &cConfig, &cAuto, &output)
			}
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && output != nil {
				C.sidereon_ppp_float_solution_free(output)
				output = nil
			}
		})
		return operationErr
	})
	runtime.KeepAlive(sp3)
	if err != nil {
		return nil, err
	}
	return newPppFloatSolution(output)
}

func SolvePppFloat(sp3 *SP3, config PppFloatConfig) (*PppFloatSolution, error) {
	return solvePppFloat(sp3, config, nil)
}

func SolvePppAutoInitFloat(sp3 *SP3, config PppFloatConfig, options PppAutoInitOptions) (*PppFloatSolution, error) {
	return solvePppFloat(sp3, config, &options)
}

func SolvePppAutoInitFixed(sp3 *SP3, floatConfig PppFloatConfig, fixedConfig PppFixedConfig, options PppAutoInitOptions) (*PppFixedSolution, error) {
	if sp3 == nil || sp3.handle == nil {
		return nil, ErrClosed
	}
	var output *C.SidereonPppFixedSolution
	err := sp3.handle.with(func(sp3Pointer unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			arena := &pppArena{}
			defer arena.close()
			cFloat, floatErr := cPppFloatConfig(floatConfig, arena, false)
			if floatErr != nil {
				operationErr = floatErr
				return
			}
			cFixed, fixedErr := cPppFixedConfig(fixedConfig, arena)
			if fixedErr != nil {
				operationErr = fixedErr
				return
			}
			cAuto := cPppAutoInit(options)
			status := C.sidereon_solve_ppp_auto_init_fixed((*C.SidereonSp3)(sp3Pointer), &cFloat, &cFixed, &cAuto, &output)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && output != nil {
				C.sidereon_ppp_fixed_solution_free(output)
				output = nil
			}
		})
		return operationErr
	})
	runtime.KeepAlive(sp3)
	if err != nil {
		return nil, err
	}
	return newPppFixedSolution(output)
}

func SolvePppFixed(sp3 *SP3, floatSolution *PppFloatSolution, config PppFixedConfig) (*PppFixedSolution, error) {
	if sp3 == nil || sp3.handle == nil || floatSolution == nil || floatSolution.handle == nil {
		return nil, ErrClosed
	}
	var output *C.SidereonPppFixedSolution
	err := withPositioningHandlePair(sp3.handle, floatSolution.handle, func(sp3Pointer, solutionPointer unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			arena := &pppArena{}
			defer arena.close()
			cConfig, conversionErr := cPppFixedConfig(config, arena)
			if conversionErr != nil {
				operationErr = conversionErr
				return
			}
			status := C.sidereon_solve_ppp_fixed((*C.SidereonSp3)(sp3Pointer), (*C.SidereonPppFloatSolution)(solutionPointer), &cConfig, &output)
			operationErr = statusErrorLocked(uint32(status))
			if operationErr != nil && output != nil {
				C.sidereon_ppp_fixed_solution_free(output)
				output = nil
			}
		})
		return operationErr
	})
	runtime.KeepAlive(sp3)
	runtime.KeepAlive(floatSolution)
	if err != nil {
		return nil, err
	}
	return newPppFixedSolution(output)
}

func pppCopyScalarLocked(pointer unsafe.Pointer, call func(*C.SidereonPppCorrections, *C.SidereonSatScalarCorrection, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]PppSatScalarCorrection, error) {
	var written, required C.size_t
	corrections := (*C.SidereonPppCorrections)(pointer)
	if err := statusErrorLocked(uint32(call(corrections, nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("PPP scalar correction", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(n, unsafe.Sizeof(C.SidereonSatScalarCorrection{})); err != nil {
		return nil, err
	}
	arena := &pppArena{}
	defer arena.close()
	memory, err := arena.malloc(n, unsafe.Sizeof(C.SidereonSatScalarCorrection{}), "PPP scalar correction output")
	if err != nil {
		return nil, err
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(corrections, (*C.SidereonSatScalarCorrection)(memory), C.size_t(n), &written, &required))); err != nil {
		return nil, err
	}
	w, err := validateTwoPassCounts("PPP scalar correction", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	values := unsafe.Slice((*C.SidereonSatScalarCorrection)(memory), w)
	out := make([]PppSatScalarCorrection, w)
	for i, value := range values {
		epochIndex, countErr := checkedNativeCount(uint64(value.epoch_index))
		if countErr != nil {
			return nil, countErr
		}
		out[i] = PppSatScalarCorrection{SatelliteID: fixedCString((*C.char)(unsafe.Pointer(&value.sat_id[0]))), EpochIndex: epochIndex, Value: float64(value.value_m)}
	}
	return out, nil
}

func pppCopyVectorLocked(pointer unsafe.Pointer, call func(*C.SidereonPppCorrections, *C.SidereonEpochVectorCorrection, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]PppEpochVectorCorrection, error) {
	var written, required C.size_t
	corrections := (*C.SidereonPppCorrections)(pointer)
	if err := statusErrorLocked(uint32(call(corrections, nil, 0, &written, &required))); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("PPP epoch vector correction", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	arena := &pppArena{}
	defer arena.close()
	memory, err := arena.malloc(n, unsafe.Sizeof(C.SidereonEpochVectorCorrection{}), "PPP epoch vector correction output")
	if err != nil {
		return nil, err
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call(corrections, (*C.SidereonEpochVectorCorrection)(memory), C.size_t(n), &written, &required))); err != nil {
		return nil, err
	}
	w, err := validateTwoPassCounts("PPP epoch vector correction", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	values := unsafe.Slice((*C.SidereonEpochVectorCorrection)(memory), w)
	out := make([]PppEpochVectorCorrection, w)
	for i, value := range values {
		epochIndex, countErr := checkedNativeCount(uint64(value.epoch_index))
		if countErr != nil {
			return nil, countErr
		}
		for axis := range out[i].ValueM {
			out[i].ValueM[axis] = float64(value.vector_m[axis])
		}
		out[i].EpochIndex = epochIndex
	}
	return out, nil
}

func (c *PppCorrections) scalar(call func(*C.SidereonPppCorrections, *C.SidereonSatScalarCorrection, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) (result []PppSatScalarCorrection, err error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	err = c.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppCopyScalarLocked(pointer, call) })
		return err
	})
	runtime.KeepAlive(c)
	return result, err
}

func (c *PppCorrections) vector(call func(*C.SidereonPppCorrections, *C.SidereonEpochVectorCorrection, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) (result []PppEpochVectorCorrection, err error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	err = c.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppCopyVectorLocked(pointer, call) })
		return err
	})
	runtime.KeepAlive(c)
	return result, err
}

func (c *PppCorrections) CodeBias() ([]PppSatScalarCorrection, error) {
	return c.scalar(func(pointer *C.SidereonPppCorrections, output *C.SidereonSatScalarCorrection, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_corrections_code_bias(pointer, output, length, written, required)
	})
}

func (c *PppCorrections) OceanLoading() ([]PppEpochVectorCorrection, error) {
	return c.vector(func(pointer *C.SidereonPppCorrections, output *C.SidereonEpochVectorCorrection, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_corrections_ocean_loading(pointer, output, length, written, required)
	})
}

func (c *PppCorrections) PoleTide() ([]PppEpochVectorCorrection, error) {
	return c.vector(func(pointer *C.SidereonPppCorrections, output *C.SidereonEpochVectorCorrection, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_corrections_pole_tide(pointer, output, length, written, required)
	})
}

func (c *PppCorrections) SatPCOECEF() ([]PppSatVectorCorrection, error) {
	if c == nil || c.handle == nil {
		return nil, ErrClosed
	}
	var result []PppSatVectorCorrection
	err := c.handle.with(func(pointer unsafe.Pointer) error {
		var operationErr error
		withCThread(func() {
			var written, required C.size_t
			call := func(output *C.SidereonSatVectorCorrection, length C.size_t) C.enum_SidereonStatus {
				return C.sidereon_ppp_corrections_sat_pco_ecef((*C.SidereonPppCorrections)(pointer), output, length, &written, &required)
			}
			if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
				result = nil
				operationErr = err
				return
			}
			n, err := validateNativeQuery("PPP satellite PCO correction", uint64(written), uint64(required))
			if err != nil {
				result = nil
				operationErr = err
				return
			}
			arena := &pppArena{}
			defer arena.close()
			memory, err := arena.malloc(n, unsafe.Sizeof(C.SidereonSatVectorCorrection{}), "PPP satellite PCO correction output")
			if err != nil {
				result = nil
				operationErr = err
				return
			}
			written, required = 0, 0
			if err := statusErrorLocked(uint32(call((*C.SidereonSatVectorCorrection)(memory), C.size_t(n)))); err != nil {
				result = nil
				operationErr = err
				return
			}
			w, err := validateTwoPassCounts("PPP satellite PCO correction", n, n, uint64(written), uint64(required))
			if err != nil {
				result = nil
				operationErr = err
				return
			}
			values := unsafe.Slice((*C.SidereonSatVectorCorrection)(memory), w)
			result = make([]PppSatVectorCorrection, w)
			for i, value := range values {
				epochIndex, countErr := checkedNativeCount(uint64(value.epoch_index))
				if countErr != nil {
					result = nil
					operationErr = countErr
					return
				}
				result[i].SatelliteID = fixedCString((*C.char)(unsafe.Pointer(&value.sat_id[0])))
				result[i].EpochIndex = epochIndex
				for axis := range result[i].ValueM {
					result[i].ValueM[axis] = float64(value.vector_m[axis])
				}
			}
		})
		return operationErr
	})
	runtime.KeepAlive(c)
	return result, err
}

func (c *PppCorrections) SatPCV() ([]PppSatScalarCorrection, error) {
	return c.scalar(func(pointer *C.SidereonPppCorrections, output *C.SidereonSatScalarCorrection, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_corrections_sat_pcv(pointer, output, length, written, required)
	})
}

func (c *PppCorrections) Tide() ([]PppEpochVectorCorrection, error) {
	return c.vector(func(pointer *C.SidereonPppCorrections, output *C.SidereonEpochVectorCorrection, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_corrections_tide(pointer, output, length, written, required)
	})
}

func (c *PppCorrections) Windup() ([]PppSatScalarCorrection, error) {
	return c.scalar(func(pointer *C.SidereonPppCorrections, output *C.SidereonSatScalarCorrection, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_corrections_windup(pointer, output, length, written, required)
	})
}

func validatePppSolveStatus(value uint32) error {
	if value != PppSolveStateToleranceValue && value != PppSolveMaxIterationsValue {
		return invalidArgument("invalid PPP solve status returned by native code")
	}
	return nil
}

func validatePppIntegerStatus(value uint32) error {
	if value != PppIntegerFixedValue && value != PppIntegerNotFixedValue {
		return invalidArgument("invalid PPP integer status returned by native code")
	}
	return nil
}

func pppIDFromC(value C.SidereonPppId) (string, error) {
	bytes := make([]byte, 0, len(value.bytes))
	for i := range value.bytes {
		if value.bytes[i] == 0 {
			return string(bytes), nil
		}
		bytes = append(bytes, byte(value.bytes[i]))
	}
	return "", invalidArgument("native PPP ambiguity ID is not NUL terminated")
}

func pppPositionCovarianceFromC(value C.SidereonPositionCovariance) PppPositionCovariance {
	var out PppPositionCovariance
	for row := range out.ECEFM2 {
		for column := range out.ECEFM2[row] {
			out.ECEFM2[row][column] = float64(value.ecef_m2[row*3+column])
			out.ENUM2[row][column] = float64(value.enu_m2[row*3+column])
		}
	}
	return out
}

func pppPositionCovariancesFromC(value C.SidereonPppPositionCovariances) PppPositionCovariances {
	return PppPositionCovariances{Posterior: pppPositionCovarianceFromC(value.posterior), Formal: pppPositionCovarianceFromC(value.formal), Temporal: pppPositionCovarianceFromC(value.temporal), PosteriorScale: float64(value.posterior_scale_factor), TemporalScale: float64(value.temporal_scale_factor)}
}

func pppTemporalCorrelationFromC(value C.SidereonPppTemporalCorrelation) (PppTemporalCorrelation, error) {
	nominal, err := checkedNativeCount(uint64(value.nominal_sample_count))
	if err != nil {
		return PppTemporalCorrelation{}, err
	}
	arcs, err := checkedNativeCount(uint64(value.arcs_used))
	if err != nil {
		return PppTemporalCorrelation{}, err
	}
	return PppTemporalCorrelation{HasLag1: bool(value.has_lag1_autocorrelation), Lag1: float64(value.lag1_autocorrelation), HasDecorrelation: bool(value.has_decorrelation_time_s), DecorrelationTimeS: float64(value.decorrelation_time_s), NominalSamples: nominal, EffectiveSamples: float64(value.effective_sample_count), VarianceInflation: float64(value.variance_inflation_factor), ArcsUsed: arcs}, nil
}

func pppTropoGradientFromC(value C.SidereonPppTropoGradientEstimate) PppTropoGradient {
	var out PppTropoGradient
	out.HasGradient, out.NorthM, out.EastM = bool(value.has_gradient), float64(value.north_m), float64(value.east_m)
	out.HasCovariance, out.HasFormalCovariance = bool(value.has_covariance_m2), bool(value.has_formal_covariance_m2)
	for row := range out.CovarianceM2 {
		for column := range out.CovarianceM2[row] {
			out.CovarianceM2[row][column] = float64(value.covariance_m2[row*2+column])
			out.FormalCovarianceM2[row][column] = float64(value.formal_covariance_m2[row*2+column])
		}
	}
	return out
}

func pppFixedMetadataFromC(value C.SidereonPppFixedMetadata) (PppFixedMetadata, error) {
	iterations, err := checkedNativeCount(uint64(value.iterations))
	if err != nil {
		return PppFixedMetadata{}, err
	}
	fixedCount, err := checkedNativeCount(uint64(value.fixed_ambiguity_count))
	if err != nil {
		return PppFixedMetadata{}, err
	}
	residualCount, err := checkedNativeCount(uint64(value.residual_count))
	if err != nil {
		return PppFixedMetadata{}, err
	}
	usedCount, err := checkedNativeCount(uint64(value.used_sat_count))
	if err != nil {
		return PppFixedMetadata{}, err
	}
	if err := validatePppSolveStatus(uint32(value.status)); err != nil {
		return PppFixedMetadata{}, err
	}
	if err := validatePppIntegerStatus(uint32(value.integer_status)); err != nil {
		return PppFixedMetadata{}, err
	}
	candidates, err := checkedNativeCount(uint64(value.integer_candidates))
	if err != nil {
		return PppFixedMetadata{}, err
	}
	return PppFixedMetadata{Iterations: iterations, Converged: bool(value.converged), Status: uint32(value.status), HasZTDResidual: bool(value.has_ztd_residual_m), ZTDResidualM: float64(value.ztd_residual_m), CodeRMSM: float64(value.code_rms_m), PhaseRMSM: float64(value.phase_rms_m), WeightedRMSM: float64(value.weighted_rms_m), FixedAmbiguityCount: fixedCount, ResidualCount: residualCount, UsedSatCount: usedCount, IntegerStatus: uint32(value.integer_status), IntegerRatio: float64(value.integer_ratio), IntegerBestScore: float64(value.integer_best_score), HasIntegerSecondBest: bool(value.has_integer_second_best_score), IntegerSecondBestScore: float64(value.integer_second_best_score), IntegerCandidates: candidates}, nil
}

func pppFloatMetadataFromC(value C.SidereonPppFloatMetadata) (PppFloatMetadata, error) {
	iterations, err := checkedNativeCount(uint64(value.iterations))
	if err != nil {
		return PppFloatMetadata{}, err
	}
	ambiguityCount, err := checkedNativeCount(uint64(value.ambiguity_count))
	if err != nil {
		return PppFloatMetadata{}, err
	}
	residualCount, err := checkedNativeCount(uint64(value.residual_count))
	if err != nil {
		return PppFloatMetadata{}, err
	}
	usedCount, err := checkedNativeCount(uint64(value.used_sat_count))
	if err != nil {
		return PppFloatMetadata{}, err
	}
	if err := validatePppSolveStatus(uint32(value.status)); err != nil {
		return PppFloatMetadata{}, err
	}
	return PppFloatMetadata{Iterations: iterations, Converged: bool(value.converged), Status: uint32(value.status), HasZTDResidual: bool(value.has_ztd_residual_m), ZTDResidualM: float64(value.ztd_residual_m), CodeRMSM: float64(value.code_rms_m), PhaseRMSM: float64(value.phase_rms_m), WeightedRMSM: float64(value.weighted_rms_m), AmbiguityCount: ambiguityCount, ResidualCount: residualCount, UsedSatCount: usedCount}, nil
}

func pppFixedAmbiguitiesLocked(pointer unsafe.Pointer) ([]PppFixedAmbiguity, error) {
	var written, required C.size_t
	call := func(output *C.SidereonPppFixedAmbiguity, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_fixed_solution_fixed_ambiguities((*C.SidereonPppFixedSolution)(pointer), output, length, &written, &required)
	}
	if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("PPP fixed ambiguities", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	arena := &pppArena{}
	defer arena.close()
	memory, err := arena.malloc(n, unsafe.Sizeof(C.SidereonPppFixedAmbiguity{}), "PPP fixed ambiguity output")
	if err != nil {
		return nil, err
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call((*C.SidereonPppFixedAmbiguity)(memory), C.size_t(n)))); err != nil {
		return nil, err
	}
	w, err := validateTwoPassCounts("PPP fixed ambiguities", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	values := unsafe.Slice((*C.SidereonPppFixedAmbiguity)(memory), w)
	out := make([]PppFixedAmbiguity, w)
	for i, value := range values {
		id, idErr := pppIDFromC(value.id)
		if idErr != nil {
			return nil, idErr
		}
		out[i] = PppFixedAmbiguity{ID: id, Cycles: int64(value.cycles), ValueM: float64(value.value_m)}
	}
	return out, nil
}

func pppFloatAmbiguitiesLocked(pointer unsafe.Pointer) ([]PppFloatAmbiguity, error) {
	var written, required C.size_t
	call := func(output *C.SidereonPppAmbiguity, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_float_solution_ambiguities((*C.SidereonPppFloatSolution)(pointer), output, length, &written, &required)
	}
	if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("PPP float ambiguities", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	arena := &pppArena{}
	defer arena.close()
	memory, err := arena.malloc(n, unsafe.Sizeof(C.SidereonPppAmbiguity{}), "PPP float ambiguity output")
	if err != nil {
		return nil, err
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call((*C.SidereonPppAmbiguity)(memory), C.size_t(n)))); err != nil {
		return nil, err
	}
	w, err := validateTwoPassCounts("PPP float ambiguities", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	values := unsafe.Slice((*C.SidereonPppAmbiguity)(memory), w)
	out := make([]PppFloatAmbiguity, w)
	for i, value := range values {
		id, idErr := pppIDFromC(value.id)
		if idErr != nil {
			return nil, idErr
		}
		out[i] = PppFloatAmbiguity{ID: id, ValueM: float64(value.value_m)}
	}
	return out, nil
}

func pppIDsLocked(pointer unsafe.Pointer, fixed bool) ([]string, error) {
	var written, required C.size_t
	call := func(output *C.SidereonPppId, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_fixed_solution_used_ids((*C.SidereonPppFixedSolution)(pointer), output, length, &written, &required)
	}
	if !fixed {
		call = func(output *C.SidereonPppId, length C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ppp_float_solution_used_ids((*C.SidereonPppFloatSolution)(pointer), output, length, &written, &required)
		}
	}
	if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("PPP used IDs", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	arena := &pppArena{}
	defer arena.close()
	memory, err := arena.malloc(n, unsafe.Sizeof(C.SidereonPppId{}), "PPP used ID output")
	if err != nil {
		return nil, err
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call((*C.SidereonPppId)(memory), C.size_t(n)))); err != nil {
		return nil, err
	}
	w, err := validateTwoPassCounts("PPP used IDs", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	values := unsafe.Slice((*C.SidereonPppId)(memory), w)
	out := make([]string, w)
	for i, value := range values {
		out[i], err = pppIDFromC(value)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *PppFixedSolution) FixedAmbiguities() (result []PppFixedAmbiguity, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppFixedAmbiguitiesLocked(pointer) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) Ambiguities() (result []PppFloatAmbiguity, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppFloatAmbiguitiesLocked(pointer) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFixedSolution) UsedIDs() (result []string, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppIDsLocked(pointer, true) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) UsedIDs() (result []string, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppIDsLocked(pointer, false) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func pppUsedSatIDsLocked(pointer unsafe.Pointer, fixed bool) ([]string, error) {
	var written, required C.size_t
	call := func(output *C.SidereonSatelliteToken, length C.size_t) C.enum_SidereonStatus {
		return C.sidereon_ppp_fixed_solution_used_sat_ids((*C.SidereonPppFixedSolution)(pointer), output, length, &written, &required)
	}
	if !fixed {
		call = func(output *C.SidereonSatelliteToken, length C.size_t) C.enum_SidereonStatus {
			return C.sidereon_ppp_float_solution_used_sat_ids((*C.SidereonPppFloatSolution)(pointer), output, length, &written, &required)
		}
	}
	if err := statusErrorLocked(uint32(call(nil, 0))); err != nil {
		return nil, err
	}
	n, err := validateNativeQuery("PPP used satellite IDs", uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	arena := &pppArena{}
	defer arena.close()
	memory, err := arena.malloc(n, unsafe.Sizeof(C.SidereonSatelliteToken{}), "PPP used satellite ID output")
	if err != nil {
		return nil, err
	}
	written, required = 0, 0
	if err := statusErrorLocked(uint32(call((*C.SidereonSatelliteToken)(memory), C.size_t(n)))); err != nil {
		return nil, err
	}
	w, err := validateTwoPassCounts("PPP used satellite IDs", n, n, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	values := unsafe.Slice((*C.SidereonSatelliteToken)(memory), w)
	out := make([]string, w)
	for i, value := range values {
		out[i] = tokenFromC(value)
	}
	return out, nil
}

func (s *PppFixedSolution) UsedSatelliteIDs() (result []string, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppUsedSatIDsLocked(pointer, true) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) UsedSatelliteIDs() (result []string, err error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppUsedSatIDsLocked(pointer, false) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func pppPositionLocked(pointer unsafe.Pointer, fixed bool) ([3]float64, error) {
	var out [3]C.double
	arena := &pppArena{}
	defer arena.close()
	memory, err := arena.malloc(3, unsafe.Sizeof(C.double(0)), "PPP position output")
	if err != nil {
		return [3]float64{}, err
	}
	if fixed {
		err = statusErrorLocked(uint32(C.sidereon_ppp_fixed_solution_position((*C.SidereonPppFixedSolution)(pointer), (*C.double)(memory), 3)))
	} else {
		err = statusErrorLocked(uint32(C.sidereon_ppp_float_solution_position((*C.SidereonPppFloatSolution)(pointer), (*C.double)(memory), 3)))
	}
	if err != nil {
		return [3]float64{}, err
	}
	values := unsafe.Slice((*C.double)(memory), 3)
	for i := range out {
		out[i] = values[i]
	}
	return [3]float64{float64(out[0]), float64(out[1]), float64(out[2])}, nil
}

func (s *PppFixedSolution) Position() (result [3]float64, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppPositionLocked(pointer, true) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) Position() (result [3]float64, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() { result, err = pppPositionLocked(pointer, false) })
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFixedSolution) FloatPosition() (result [3]float64, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var values [3]C.double
			arena := &pppArena{}
			defer arena.close()
			memory, allocationErr := arena.malloc(3, unsafe.Sizeof(C.double(0)), "PPP float position output")
			if allocationErr != nil {
				err = allocationErr
				return
			}
			if status := C.sidereon_ppp_fixed_solution_float_position((*C.SidereonPppFixedSolution)(pointer), (*C.double)(memory), 3); status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			out := unsafe.Slice((*C.double)(memory), 3)
			for i := range values {
				values[i] = out[i]
			}
			for i := range result {
				result[i] = float64(values[i])
			}
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFixedSolution) Metadata() (result PppFixedMetadata, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppFixedMetadata
			status := C.sidereon_ppp_fixed_solution_metadata((*C.SidereonPppFixedSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result, err = pppFixedMetadataFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) Metadata() (result PppFloatMetadata, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppFloatMetadata
			status := C.sidereon_ppp_float_solution_metadata((*C.SidereonPppFloatSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result, err = pppFloatMetadataFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFixedSolution) PositionCovariances() (result PppPositionCovariances, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppPositionCovariances
			status := C.sidereon_ppp_fixed_solution_position_covariances((*C.SidereonPppFixedSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result = pppPositionCovariancesFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) PositionCovariances() (result PppPositionCovariances, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppPositionCovariances
			status := C.sidereon_ppp_float_solution_position_covariances((*C.SidereonPppFloatSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result = pppPositionCovariancesFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFixedSolution) TemporalCorrelation() (result PppTemporalCorrelation, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppTemporalCorrelation
			status := C.sidereon_ppp_fixed_solution_temporal_correlation((*C.SidereonPppFixedSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result, err = pppTemporalCorrelationFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) TemporalCorrelation() (result PppTemporalCorrelation, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppTemporalCorrelation
			status := C.sidereon_ppp_float_solution_temporal_correlation((*C.SidereonPppFloatSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result, err = pppTemporalCorrelationFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFixedSolution) TropoGradient() (result PppTropoGradient, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppTropoGradientEstimate
			status := C.sidereon_ppp_fixed_solution_tropo_gradient((*C.SidereonPppFixedSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result = pppTropoGradientFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}

func (s *PppFloatSolution) TropoGradient() (result PppTropoGradient, err error) {
	if s == nil || s.handle == nil {
		return result, ErrClosed
	}
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			var value C.SidereonPppTropoGradientEstimate
			status := C.sidereon_ppp_float_solution_tropo_gradient((*C.SidereonPppFloatSolution)(pointer), &value)
			if status != C.SIDEREON_STATUS_OK {
				err = statusErrorLocked(uint32(status))
				return
			}
			result = pppTropoGradientFromC(value)
		})
		return err
	})
	runtime.KeepAlive(s)
	return result, err
}
