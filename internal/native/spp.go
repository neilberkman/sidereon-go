//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

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

type SPPGeometryQuality struct {
	Tier                uint32
	Redundancy          int32
	Rank                int
	ConditionNumber     float64
	GDOP                float64
	RAIMCheckable       bool
	CovarianceValidated bool
}

type SPPMetadata struct {
	Iterations          int
	Converged           bool
	Status              uint32
	IonosphereApplied   bool
	TroposphereApplied  bool
	OuterIterations     int
	HasFinalRobustScale bool
	FinalRobustScaleM   float64
	UsedCount           int
	SystemCount         int
	Redundancy          int64
	RAIMCheckable       bool
	GeometryQuality     SPPGeometryQuality
}

type SPPSolution struct {
	PositionM          [3]float64
	ReceiverClockS     float64
	UsedSatelliteCount int
	UsedSatelliteIDs   []string
	ResidualsM         []float64
	DOP                *Dop
	Geodetic           *Geodetic
	Metadata           SPPMetadata
}

func (s *SP3) Solve(config SPPConfig) (SPPSolution, error) {
	if s == nil || s.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	for _, observation := range config.Observations {
		if err := rejectEmbeddedNUL(observation.SatelliteID, "SPP satellite ID"); err != nil {
			return SPPSolution{}, err
		}
	}
	var result SPPSolution
	var operationErr error
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		withCThread(func() {
			observationCount, err := checkedNativeSize(len(config.Observations))
			if err != nil {
				operationErr = err
				return
			}
			var observationMemory unsafe.Pointer
			if len(config.Observations) != 0 {
				size, err := checkedNativeAllocationSize(
					len(config.Observations), unsafe.Sizeof(C.SidereonObservation{}),
				)
				if err != nil {
					operationErr = err
					return
				}
				observationMemory = C.malloc(C.size_t(size))
				if observationMemory == nil {
					operationErr = errors.New("sidereon: unable to allocate native observations")
					return
				}
				defer C.free(observationMemory)
			}
			cObservations := unsafe.Slice((*C.SidereonObservation)(observationMemory), len(config.Observations))
			cIDs := make([]unsafe.Pointer, 0, len(config.Observations))
			defer func() {
				for _, id := range cIDs {
					C.free(id)
				}
			}()
			for i, observation := range config.Observations {
				id := C.CBytes(append([]byte(observation.SatelliteID), 0))
				if id == nil {
					operationErr = errors.New("sidereon: unable to allocate native satellite id")
					return
				}
				cIDs = append(cIDs, id)
				cObservations[i].sat_id = (*C.char)(id)
				cObservations[i].pseudorange_m = C.double(observation.PseudorangeM)
			}

			var inputs C.SidereonSppInputs
			if len(cObservations) != 0 {
				inputs.observations = &cObservations[0]
			}
			inputs.observation_count = observationCount
			inputs.t_rx_j2000_s = C.double(config.TRxJ2000S)
			inputs.t_rx_second_of_day_s = C.double(config.TRxSecondOfDayS)
			inputs.day_of_year = C.double(config.DayOfYear)
			for i := range config.InitialGuess {
				inputs.initial_guess[i] = C.double(config.InitialGuess[i])
			}
			inputs.ionosphere = C.bool(config.Ionosphere)
			inputs.troposphere = C.bool(config.Troposphere)
			inputs.with_geodetic = C.bool(config.WithGeodetic)

			var solution *C.SidereonSppSolution
			status := C.sidereon_solve_spp(
				(*C.SidereonSp3)(pointer), &inputs, &solution,
			)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				if solution != nil {
					C.sidereon_spp_solution_free(solution)
				}
				return
			}
			if solution == nil {
				operationErr = errors.New("sidereon: native SPP solve returned no solution")
				return
			}
			defer C.sidereon_spp_solution_free(solution)

			var position [3]C.double
			status = C.sidereon_spp_solution_position(solution, &position[0], 3)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			for i := range position {
				result.PositionM[i] = float64(position[i])
			}

			var receiverClock C.double
			status = C.sidereon_spp_solution_rx_clock_s(solution, &receiverClock)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			result.ReceiverClockS = float64(receiverClock)

			var usedCount C.size_t
			status = C.sidereon_spp_solution_used_sat_count(solution, &usedCount)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			used, err := checkedNativeCount(uint64(usedCount))
			if err != nil {
				operationErr = err
				return
			}
			result.UsedSatelliteCount = used

			var written, required C.size_t
			status = C.sidereon_spp_solution_used_sat_ids(
				solution, nil, 0, &written, &required,
			)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			idCount, err := checkedNativeCount(uint64(required))
			if err != nil {
				operationErr = err
				return
			}
			if _, err := writtenToInt(written, 0, "SPP used satellite ID first-call written count"); err != nil {
				operationErr = err
				return
			}
			if _, err := checkedNativeAllocationSize(idCount, unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
				operationErr = err
				return
			}
			ids := make([]C.SidereonSatelliteToken, idCount)
			var idOutput *C.SidereonSatelliteToken
			if len(ids) != 0 {
				idOutput = &ids[0]
			}
			status = C.sidereon_spp_solution_used_sat_ids(
				solution, idOutput, C.size_t(len(ids)), &written, &required,
			)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			writtenIDs, err := validateTwoPassCounts(
				"SPP used satellite IDs", len(ids), idCount, uint64(written), uint64(required),
			)
			if err != nil {
				operationErr = err
				return
			}
			result.UsedSatelliteIDs = make([]string, writtenIDs)
			for i := 0; i < writtenIDs; i++ {
				result.UsedSatelliteIDs[i] = C.GoString((*C.char)(unsafe.Pointer(&ids[i].bytes[0])))
			}

			written, required = 0, 0
			status = C.sidereon_spp_solution_residuals(
				solution, nil, 0, &written, &required,
			)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			residualCount, err := checkedNativeCount(uint64(required))
			if err != nil {
				operationErr = err
				return
			}
			if _, err := writtenToInt(written, 0, "SPP residual first-call written count"); err != nil {
				operationErr = err
				return
			}
			if _, err := checkedNativeAllocationSize(residualCount, unsafe.Sizeof(C.double(0))); err != nil {
				operationErr = err
				return
			}
			residuals := make([]C.double, residualCount)
			var residualOutput *C.double
			if len(residuals) != 0 {
				residualOutput = &residuals[0]
			}
			status = C.sidereon_spp_solution_residuals(
				solution, residualOutput, C.size_t(len(residuals)), &written, &required,
			)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			writtenResiduals, err := validateTwoPassCounts(
				"SPP residuals", len(residuals), residualCount, uint64(written), uint64(required),
			)
			if err != nil {
				operationErr = err
				return
			}
			result.ResidualsM = make([]float64, writtenResiduals)
			for i := 0; i < writtenResiduals; i++ {
				result.ResidualsM[i] = float64(residuals[i])
			}

			var metadata C.SidereonSppMetadata
			status = C.sidereon_spp_solution_metadata(solution, &metadata)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			metadataIterations, err := checkedNativeCount(uint64(metadata.iterations))
			if err != nil {
				operationErr = err
				return
			}
			metadataOuterIterations, err := checkedNativeCount(uint64(metadata.outer_iterations))
			if err != nil {
				operationErr = err
				return
			}
			metadataUsedCount, err := checkedNativeCount(uint64(metadata.used_count))
			if err != nil {
				operationErr = err
				return
			}
			metadataSystemCount, err := checkedNativeCount(uint64(metadata.system_count))
			if err != nil {
				operationErr = err
				return
			}
			qualityRank, err := checkedNativeCount(uint64(metadata.geometry_quality.rank))
			if err != nil {
				operationErr = err
				return
			}
			result.Metadata = SPPMetadata{
				Iterations:          metadataIterations,
				Converged:           bool(metadata.converged),
				Status:              uint32(metadata.status),
				IonosphereApplied:   bool(metadata.ionosphere_applied),
				TroposphereApplied:  bool(metadata.troposphere_applied),
				OuterIterations:     metadataOuterIterations,
				HasFinalRobustScale: bool(metadata.has_final_robust_scale_m),
				FinalRobustScaleM:   float64(metadata.final_robust_scale_m),
				UsedCount:           metadataUsedCount,
				SystemCount:         metadataSystemCount,
				Redundancy:          int64(metadata.redundancy),
				RAIMCheckable:       bool(metadata.raim_checkable),
				GeometryQuality: SPPGeometryQuality{
					Tier:                uint32(metadata.geometry_quality.tier),
					Redundancy:          int32(metadata.geometry_quality.redundancy),
					Rank:                qualityRank,
					ConditionNumber:     float64(metadata.geometry_quality.condition_number),
					GDOP:                float64(metadata.geometry_quality.gdop),
					RAIMCheckable:       bool(metadata.geometry_quality.raim_checkable),
					CovarianceValidated: bool(metadata.geometry_quality.covariance_validated),
				},
			}

			var dop C.SidereonDop
			status = C.sidereon_spp_solution_dop(solution, &dop)
			if status == C.SIDEREON_STATUS_OK {
				result.DOP = &Dop{
					GDOP: float64(dop.gdop),
					PDOP: float64(dop.pdop),
					HDOP: float64(dop.hdop),
					VDOP: float64(dop.vdop),
					TDOP: float64(dop.tdop),
				}
			} else if status != C.SIDEREON_STATUS_INVALID_ARGUMENT {
				operationErr = statusErrorLocked(uint32(status))
				return
			}

			var geodetic C.SidereonGeodetic
			var present C.bool
			status = C.sidereon_spp_solution_geodetic(solution, &geodetic, &present)
			if status != C.SIDEREON_STATUS_OK {
				operationErr = statusErrorLocked(uint32(status))
				return
			}
			if bool(present) {
				result.Geodetic = &Geodetic{
					LatitudeRad:  float64(geodetic.lat_rad),
					LongitudeRad: float64(geodetic.lon_rad),
					HeightM:      float64(geodetic.height_m),
				}
			}
		})
		return operationErr
	})
	runtime.KeepAlive(s)
	if err != nil {
		return SPPSolution{}, err
	}
	return result, nil
}
