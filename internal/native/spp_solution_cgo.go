//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

func withSPPInputs(config SPPConfig, fn func(*C.SidereonSppInputs) error) error {
	observationCount, err := checkedNativeSize(len(config.Observations))
	if err != nil {
		return err
	}
	for _, observation := range config.Observations {
		if err := rejectEmbeddedNUL(observation.SatelliteID, "SPP satellite ID"); err != nil {
			return err
		}
		if len(observation.SatelliteID) >= 16 {
			return fmt.Errorf("sidereon: SPP satellite ID is too long: %d bytes", len(observation.SatelliteID))
		}
	}
	var operationErr error
	withCThread(func() {
		var observationMemory unsafe.Pointer
		if len(config.Observations) != 0 {
			size, err := checkedNativeAllocationSize(len(config.Observations), unsafe.Sizeof(C.SidereonObservation{}))
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
		observations := unsafe.Slice((*C.SidereonObservation)(observationMemory), len(config.Observations))
		ids := make([]unsafe.Pointer, 0, len(config.Observations))
		defer func() {
			for _, id := range ids {
				C.free(id)
			}
		}()
		for i, observation := range config.Observations {
			id := C.CBytes(append([]byte(observation.SatelliteID), 0))
			if id == nil {
				operationErr = errors.New("sidereon: unable to allocate native satellite ID")
				return
			}
			ids = append(ids, id)
			observations[i].sat_id = (*C.char)(id)
			observations[i].pseudorange_m = C.double(observation.PseudorangeM)
		}

		inputMemory := C.calloc(1, C.size_t(unsafe.Sizeof(C.SidereonSppInputs{})))
		if inputMemory == nil {
			operationErr = errors.New("sidereon: unable to allocate native SPP inputs")
			return
		}
		defer C.free(inputMemory)
		inputs := (*C.SidereonSppInputs)(inputMemory)
		if len(observations) != 0 {
			inputs.observations = &observations[0]
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
		for i := 0; i < 4; i++ {
			inputs.klobuchar_alpha[i] = C.double(config.KlobucharAlpha[i])
			inputs.klobuchar_beta[i] = C.double(config.KlobucharBeta[i])
		}
		inputs.pressure_hpa = C.double(config.PressureHPA)
		inputs.temperature_k = C.double(config.TemperatureK)
		inputs.relative_humidity = C.double(config.RelativeHumidity)
		operationErr = fn(inputs)
	})
	runtime.KeepAlive(config.Observations)
	return operationErr
}

func validateSPPSatelliteID(value string) error {
	if err := rejectEmbeddedNUL(value, "SPP satellite ID"); err != nil {
		return err
	}
	if len(value) >= 16 {
		return fmt.Errorf("sidereon: SPP satellite ID is too long: %d bytes", len(value))
	}
	return nil
}

func (s *SBASCorrectionStore) SolveBroadcast(b *BroadcastEphemeris, geo string, mode uint32, config SPPConfig) (SPPSolution, error) {
	if s == nil || s.resource == nil || b == nil || b.resource == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSBASSolveModeValue(mode); err != nil {
		return SPPSolution{}, err
	}
	var result SPPSolution
	err := b.resource.with(func(broadcastPointer unsafe.Pointer) error {
		return s.resource.with(func(storePointer unsafe.Pointer) error {
			return withSPPInputs(config, func(inputs *C.SidereonSppInputs) error {
				if err := validateSPPSatelliteID(geo); err != nil {
					return err
				}
				geoPointer := C.CString(geo)
				if geoPointer == nil {
					return errors.New("sidereon: unable to allocate native GEO ID")
				}
				defer C.free(unsafe.Pointer(geoPointer))
				var solution *C.SidereonSppSolution
				status := C.sidereon_sbas_solve_broadcast(
					(*C.SidereonBroadcastEphemeris)(broadcastPointer),
					(*C.SidereonSbasCorrectionStore)(storePointer), geoPointer,
					C.uint32_t(mode), inputs, &solution,
				)
				if err := statusErrorLocked(uint32(status)); err != nil {
					if solution != nil {
						C.sidereon_spp_solution_free(solution)
					}
					return err
				}
				if solution == nil {
					return errors.New("sidereon: native SBAS solve returned no solution")
				}
				defer C.sidereon_spp_solution_free(solution)
				var readErr error
				result, readErr = readSPPSolutionLocked(solution)
				return readErr
			})
		})
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(b)
	return result, err
}

func (s *SSRCorrectionStore) SolveBroadcast(b *BroadcastEphemeris, config SPPConfig, staleness float64, missing uint32, allowRegionalProvider bool, regionalProviderID uint16) (SPPSolution, error) {
	if s == nil || s.resource == nil || b == nil || b.resource == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSSRMissingActionValue(missing); err != nil {
		return SPPSolution{}, err
	}
	var result SPPSolution
	err := b.resource.with(func(broadcastPointer unsafe.Pointer) error {
		return s.resource.with(func(storePointer unsafe.Pointer) error {
			return withSPPInputs(config, func(inputs *C.SidereonSppInputs) error {
				var solution *C.SidereonSppSolution
				status := C.sidereon_ssr_solve_broadcast(
					(*C.SidereonBroadcastEphemeris)(broadcastPointer),
					(*C.SidereonSsrCorrectionStore)(storePointer), C.double(staleness),
					C.uint32_t(missing), C.bool(allowRegionalProvider), C.uint16_t(regionalProviderID),
					inputs, &solution,
				)
				if err := statusErrorLocked(uint32(status)); err != nil {
					if solution != nil {
						C.sidereon_spp_solution_free(solution)
					}
					return err
				}
				if solution == nil {
					return errors.New("sidereon: native SSR solve returned no solution")
				}
				defer C.sidereon_spp_solution_free(solution)
				var readErr error
				result, readErr = readSPPSolutionLocked(solution)
				return readErr
			})
		})
	})
	runtime.KeepAlive(s)
	runtime.KeepAlive(b)
	return result, err
}

// readSPPSolutionLocked copies the complete public SPP solution while the
// caller holds the native source locks and the OS thread is locked. It is
// shared by source-specific correction solves so they expose the same exact
// aggregate as the existing SP3 SPP route.
func readSPPSolutionLocked(solution *C.SidereonSppSolution) (SPPSolution, error) {
	if solution == nil {
		return SPPSolution{}, errors.New("sidereon: native SPP solve returned no solution")
	}
	var result SPPSolution
	var position [3]C.double
	if err := statusErrorLocked(C.sidereon_spp_solution_position(solution, &position[0], 3)); err != nil {
		return SPPSolution{}, err
	}
	for i := range result.PositionM {
		result.PositionM[i] = float64(position[i])
	}
	var receiverClock C.double
	if err := statusErrorLocked(C.sidereon_spp_solution_rx_clock_s(solution, &receiverClock)); err != nil {
		return SPPSolution{}, err
	}
	result.ReceiverClockS = float64(receiverClock)
	var usedCount C.size_t
	if err := statusErrorLocked(C.sidereon_spp_solution_used_sat_count(solution, &usedCount)); err != nil {
		return SPPSolution{}, err
	}
	var err error
	result.UsedSatelliteCount, err = checkedNativeCount(uint64(usedCount))
	if err != nil {
		return SPPSolution{}, err
	}

	var written, required C.size_t
	if err := statusErrorLocked(C.sidereon_spp_solution_used_sat_ids(solution, nil, 0, &written, &required)); err != nil {
		return SPPSolution{}, err
	}
	idCount, err := validateNativeQuery("SPP used satellite IDs", uint64(written), uint64(required))
	if err != nil {
		return SPPSolution{}, err
	}
	if _, err := checkedNativeAllocationSize(idCount, unsafe.Sizeof(C.SidereonSatelliteToken{})); err != nil {
		return SPPSolution{}, err
	}
	ids := make([]C.SidereonSatelliteToken, idCount)
	idLength, err := checkedNativeSize(len(ids))
	if err != nil {
		return SPPSolution{}, err
	}
	var idOutput *C.SidereonSatelliteToken
	if len(ids) > 0 {
		idOutput = &ids[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(C.sidereon_spp_solution_used_sat_ids(solution, idOutput, idLength, &written, &required)); err != nil {
		return SPPSolution{}, err
	}
	writtenIDs, err := validateTwoPassCounts("SPP used satellite IDs", len(ids), idCount, uint64(written), uint64(required))
	if err != nil {
		return SPPSolution{}, err
	}
	result.UsedSatelliteIDs = make([]string, writtenIDs)
	for i := range result.UsedSatelliteIDs {
		result.UsedSatelliteIDs[i] = tokenFromC(ids[i])
	}

	written, required = 0, 0
	if err := statusErrorLocked(C.sidereon_spp_solution_residuals(solution, nil, 0, &written, &required)); err != nil {
		return SPPSolution{}, err
	}
	residualCount, err := validateNativeQuery("SPP residuals", uint64(written), uint64(required))
	if err != nil {
		return SPPSolution{}, err
	}
	if _, err := checkedNativeAllocationSize(residualCount, unsafe.Sizeof(C.double(0))); err != nil {
		return SPPSolution{}, err
	}
	residuals := make([]C.double, residualCount)
	residualLength, err := checkedNativeSize(len(residuals))
	if err != nil {
		return SPPSolution{}, err
	}
	var residualOutput *C.double
	if len(residuals) > 0 {
		residualOutput = &residuals[0]
	}
	written, required = 0, 0
	if err := statusErrorLocked(C.sidereon_spp_solution_residuals(solution, residualOutput, residualLength, &written, &required)); err != nil {
		return SPPSolution{}, err
	}
	writtenResiduals, err := validateTwoPassCounts("SPP residuals", len(residuals), residualCount, uint64(written), uint64(required))
	if err != nil {
		return SPPSolution{}, err
	}
	result.ResidualsM = make([]float64, writtenResiduals)
	for i := range result.ResidualsM {
		result.ResidualsM[i] = float64(residuals[i])
	}

	var metadata C.SidereonSppMetadata
	if err := statusErrorLocked(C.sidereon_spp_solution_metadata(solution, &metadata)); err != nil {
		return SPPSolution{}, err
	}
	if uint32(metadata.status) > uint32(C.SIDEREON_SPP_SOLVE_STATUS_MAX_EVALUATIONS) {
		return SPPSolution{}, invalidArgument("invalid SPP solve status returned by native code")
	}
	if uint32(metadata.geometry_quality.tier) > uint32(C.SIDEREON_OBSERVABILITY_TIER_NOMINAL) {
		return SPPSolution{}, invalidArgument("invalid SPP observability tier returned by native code")
	}
	iterations, err := checkedNativeCount(uint64(metadata.iterations))
	if err != nil {
		return SPPSolution{}, err
	}
	outerIterations, err := checkedNativeCount(uint64(metadata.outer_iterations))
	if err != nil {
		return SPPSolution{}, err
	}
	metadataUsedCount, err := checkedNativeCount(uint64(metadata.used_count))
	if err != nil {
		return SPPSolution{}, err
	}
	metadataSystemCount, err := checkedNativeCount(uint64(metadata.system_count))
	if err != nil {
		return SPPSolution{}, err
	}
	qualityRank, err := checkedNativeCount(uint64(metadata.geometry_quality.rank))
	if err != nil {
		return SPPSolution{}, err
	}
	result.Metadata = SPPMetadata{Iterations: iterations, Converged: bool(metadata.converged), Status: uint32(metadata.status), IonosphereApplied: bool(metadata.ionosphere_applied), TroposphereApplied: bool(metadata.troposphere_applied), OuterIterations: outerIterations, HasFinalRobustScale: bool(metadata.has_final_robust_scale_m), FinalRobustScaleM: float64(metadata.final_robust_scale_m), UsedCount: metadataUsedCount, SystemCount: metadataSystemCount, Redundancy: int64(metadata.redundancy), RAIMCheckable: bool(metadata.raim_checkable), GeometryQuality: SPPGeometryQuality{Tier: uint32(metadata.geometry_quality.tier), Redundancy: int32(metadata.geometry_quality.redundancy), Rank: qualityRank, ConditionNumber: float64(metadata.geometry_quality.condition_number), GDOP: float64(metadata.geometry_quality.gdop), RAIMCheckable: bool(metadata.geometry_quality.raim_checkable), CovarianceValidated: bool(metadata.geometry_quality.covariance_validated)}}

	var dop C.SidereonDop
	status := C.sidereon_spp_solution_dop(solution, &dop)
	if status == C.SIDEREON_STATUS_OK {
		result.DOP = &Dop{GDOP: float64(dop.gdop), PDOP: float64(dop.pdop), HDOP: float64(dop.hdop), VDOP: float64(dop.vdop), TDOP: float64(dop.tdop)}
	} else if status != C.SIDEREON_STATUS_INVALID_ARGUMENT {
		return SPPSolution{}, statusErrorLocked(uint32(status))
	}

	var geodetic C.SidereonGeodetic
	var present C.bool
	if err := statusErrorLocked(C.sidereon_spp_solution_geodetic(solution, &geodetic, &present)); err != nil {
		return SPPSolution{}, err
	}
	if bool(present) {
		result.Geodetic = &Geodetic{LatitudeRad: float64(geodetic.lat_rad), LongitudeRad: float64(geodetic.lon_rad), HeightM: float64(geodetic.height_m)}
	}
	return result, nil
}
