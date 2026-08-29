//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type NativeFusionIMUSpec struct {
	AccelVRW, GyroARW, AccelBiasInstability, GyroBiasInstability, AccelBiasTau, GyroBiasTau float64
	HasAccelScale, HasGyroScale                                                             bool
	AccelScalePPM, GyroScalePPM                                                             float64
}
type NativeFusionConfig struct {
	FilterKind, Layout                              uint32
	IMUSpec                                         NativeFusionIMUSpec
	TimeSyncIMUCapacity, TimeSyncCheckpointCapacity uint64
}
type NativeFusionNavState struct {
	Epoch                                      float64
	Position, Velocity                         [3]float64
	Attitude                                   [9]float64
	AccelBias, GyroBias, AccelScale, GyroScale [3]float64
}
type NativeFusionIMUSample struct {
	Epoch                                                 float64
	Kind                                                  uint32
	SpecificForce, AngularRate, DeltaVelocity, DeltaTheta [3]float64
	DTS                                                   float64
}
type NativeFusionState struct {
	Epoch                                      float64
	Position, Velocity                         [3]float64
	Attitude                                   [9]float64
	AccelBias, GyroBias, AccelScale, GyroScale [3]float64
	CovarianceDimension                        uint64
	LastBodyRate                               [3]float64
	ClockBias, ClockDrift                      float64
	ClockCovariance                            [4]float64
}
type NativeFusionTimeSyncStatus struct {
	IMUCapacity, IMULength, CheckpointCapacity, CheckpointLength         uint64
	HasOldestIMU, HasNewestIMU, HasOldestCheckpoint, HasNewestCheckpoint bool
	OldestIMU, NewestIMU, OldestCheckpoint, NewestCheckpoint             float64
}
type NativeFusionLooseMeasurement struct {
	Epoch              float64
	Position, Velocity [3]float64
	HasVelocity        bool
	Covariance         []float64
	SatellitesUsed     uint64
	SolutionValid      bool
	FixStatus          uint32
}
type NativeFusionUpdate struct {
	Applied                          bool
	NIS                              float64
	Rows, AcceptedRows, RejectedRows uint64
}
type NativeFusionTimeSyncUpdate struct {
	Update                      NativeFusionUpdate
	Late                        bool
	Replayed                    uint64
	RestoredEpoch, CurrentEpoch float64
}
type NativeFusionTightRangeRate struct{ Measured, Sigma, ClockDrift float64 }
type NativeFusionTightCarrierPhase struct{ PhaseRange, Sigma, FloatAmbiguity float64 }
type NativeFusionTightObservation struct {
	SatelliteID             string
	Pseudorange, Sigma      float64
	HasRangeRate            bool
	RangeRate               NativeFusionTightRangeRate
	HasCarrierPhase         bool
	CarrierPhase            NativeFusionTightCarrierPhase
	Ionosphere, Troposphere float64
}
type NativeFusionTightEpoch struct {
	Epoch        float64
	Observations []NativeFusionTightObservation
}
type NativeFusionRTSEpoch struct {
	Epoch                                   float64
	CovarianceDimension, AugmentedDimension uint64
	HasTransition                           bool
}
type FusionFilter struct {
	_      noCopy
	handle *surfaceHandle
}
type FusionRTSHistoryBuilder struct {
	_      noCopy
	handle *surfaceHandle
}
type FusionRTSHistory struct {
	_      noCopy
	handle *surfaceHandle
}
type NativeFusionVelocityState struct {
	Epoch              float64
	Position, Velocity [3]float64
}
type NativeFusionVelocityConfig struct{ MaxOutage float64 }
type NativeFusionVelocityTrajectory struct {
	StateCount                         uint64
	EndpointPosition, EndpointVelocity [3]float64
}

func cFusionSpec(v NativeFusionIMUSpec) C.SidereonFusionImuSpec {
	return C.SidereonFusionImuSpec{accel_vrw_mps_sqrt_s: C.double(v.AccelVRW), gyro_arw_rad_sqrt_s: C.double(v.GyroARW), accel_bias_instab_mps2: C.double(v.AccelBiasInstability), gyro_bias_instab_rps: C.double(v.GyroBiasInstability), accel_bias_tau_s: C.double(v.AccelBiasTau), gyro_bias_tau_s: C.double(v.GyroBiasTau), has_accel_scale_instab_ppm: C.bool(v.HasAccelScale), accel_scale_instab_ppm: C.double(v.AccelScalePPM), has_gyro_scale_instab_ppm: C.bool(v.HasGyroScale), gyro_scale_instab_ppm: C.double(v.GyroScalePPM)}
}
func cFusionArray3(v [3]float64) [3]C.double {
	var out [3]C.double
	for i := 0; i < 3; i++ {
		out[i] = C.double(v[i])
	}
	return out
}
func cFusionArray9(v [9]float64) [9]C.double {
	var out [9]C.double
	for i := 0; i < 9; i++ {
		out[i] = C.double(v[i])
	}
	return out
}
func cFusionNav(v NativeFusionNavState) C.SidereonFusionNavState {
	return C.SidereonFusionNavState{t_j2000_s: C.double(v.Epoch), position_ecef_m: cFusionArray3(v.Position), velocity_ecef_mps: cFusionArray3(v.Velocity), attitude_body_to_ecef: cFusionArray9(v.Attitude), accel_bias_mps2: cFusionArray3(v.AccelBias), gyro_bias_rps: cFusionArray3(v.GyroBias), accel_scale_factor: cFusionArray3(v.AccelScale), gyro_scale_factor: cFusionArray3(v.GyroScale)}
}
func cFusionSample(v NativeFusionIMUSample) C.SidereonFusionImuSample {
	return C.SidereonFusionImuSample{t_j2000_s: C.double(v.Epoch), kind: C.uint32_t(v.Kind), specific_force_mps2: cFusionArray3(v.SpecificForce), angular_rate_rps: cFusionArray3(v.AngularRate), delta_velocity_mps: cFusionArray3(v.DeltaVelocity), delta_theta_rad: cFusionArray3(v.DeltaTheta), dt_s: C.double(v.DTS)}
}
func fusionState(v C.SidereonFusionState) NativeFusionState {
	var out NativeFusionState
	out.Epoch = float64(v.t_j2000_s)
	for i := 0; i < 3; i++ {
		out.Position[i] = float64(v.position_ecef_m[i])
		out.Velocity[i] = float64(v.velocity_ecef_mps[i])
		out.AccelBias[i] = float64(v.accel_bias_mps2[i])
		out.GyroBias[i] = float64(v.gyro_bias_rps[i])
		out.AccelScale[i] = float64(v.accel_scale_factor[i])
		out.GyroScale[i] = float64(v.gyro_scale_factor[i])
		out.LastBodyRate[i] = float64(v.last_body_rate_wrt_ecef_rps[i])
	}
	for i := 0; i < 9; i++ {
		out.Attitude[i] = float64(v.attitude_body_to_ecef[i])
	}
	out.CovarianceDimension = uint64(v.covariance_dimension)
	out.ClockBias = float64(v.tight_clock_bias_m)
	out.ClockDrift = float64(v.tight_clock_drift_m_s)
	for i := 0; i < 4; i++ {
		out.ClockCovariance[i] = float64(v.tight_clock_covariance[i])
	}
	return out
}
func fusionUpdate(v C.SidereonFusionUpdate) NativeFusionUpdate {
	return NativeFusionUpdate{bool(v.applied), float64(v.nis), uint64(v.rows), uint64(v.accepted_rows), uint64(v.rejected_rows)}
}
func fusionCallRead(h *surfaceHandle, call func(unsafe.Pointer) C.enum_SidereonStatus) error {
	return h.read(func(p unsafe.Pointer) error {
		var e error
		withCThread(func() { e = statusErrorLocked(uint32(call(p))) })
		return e
	})
}
func fusionCallWrite(h *surfaceHandle, call func(unsafe.Pointer) C.enum_SidereonStatus) error {
	return h.write(func(p unsafe.Pointer) error {
		var e error
		withCThread(func() { e = statusErrorLocked(uint32(call(p))) })
		return e
	})
}

func FusionConfigDefault() (NativeFusionConfig, error) {
	var c C.SidereonFusionFilterConfig
	err := callStatus(func() uint32 { return uint32(C.sidereon_fusion_filter_config_init(&c)) })
	return NativeFusionConfig{uint32(c.filter_kind), uint32(c.error_state_layout), NativeFusionIMUSpec{float64(c.imu_spec.accel_vrw_mps_sqrt_s), float64(c.imu_spec.gyro_arw_rad_sqrt_s), float64(c.imu_spec.accel_bias_instab_mps2), float64(c.imu_spec.gyro_bias_instab_rps), float64(c.imu_spec.accel_bias_tau_s), float64(c.imu_spec.gyro_bias_tau_s), bool(c.imu_spec.has_accel_scale_instab_ppm), bool(c.imu_spec.has_gyro_scale_instab_ppm), float64(c.imu_spec.accel_scale_instab_ppm), float64(c.imu_spec.gyro_scale_instab_ppm)}, uint64(c.time_sync_imu_capacity), uint64(c.time_sync_checkpoint_capacity)}, err
}
func NewFusionFilter(initial NativeFusionNavState, diagonal []float64, config NativeFusionConfig) (*FusionFilter, error) {
	if len(diagonal) == 0 {
		return nil, errors.New("sidereon: fusion covariance diagonal must not be empty")
	}
	imuCapacity, err := cSize64(config.TimeSyncIMUCapacity, "fusion IMU capacity")
	if err != nil {
		return nil, err
	}
	checkpointCapacity, err := cSize64(config.TimeSyncCheckpointCapacity, "fusion checkpoint capacity")
	if err != nil {
		return nil, err
	}
	p, diagonalLength, err := cFloats(diagonal, "fusion covariance diagonal")
	if err != nil {
		return nil, err
	}
	defer C.free(p)
	var c C.SidereonFusionFilterConfig
	var opErr error
	withCThread(func() { opErr = statusErrorLocked(uint32(C.sidereon_fusion_filter_config_init(&c))) })
	if opErr != nil {
		return nil, opErr
	}
	c.filter_kind = C.uint32_t(config.FilterKind)
	c.error_state_layout = C.uint32_t(config.Layout)
	c.imu_spec = cFusionSpec(config.IMUSpec)
	c.time_sync_imu_capacity = imuCapacity
	c.time_sync_checkpoint_capacity = checkpointCapacity
	nav := cFusionNav(initial)
	var out *C.SidereonFusionFilter
	withCThread(func() {
		opErr = statusErrorLocked(uint32(C.sidereon_fusion_filter_create(&nav, (*C.double)(p), diagonalLength, &c, &out)))
	})
	if opErr != nil {
		return nil, opErr
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(x unsafe.Pointer) { C.sidereon_fusion_filter_free((*C.SidereonFusionFilter)(x)) })
	if e != nil {
		return nil, e
	}
	return &FusionFilter{handle: h}, nil
}
func (f *FusionFilter) Close() error {
	if f == nil {
		return nil
	}
	return f.handle.close()
}
func (f *FusionFilter) State() (NativeFusionState, error) {
	var out C.SidereonFusionState
	err := fusionCallRead(f.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_state((*C.SidereonFusionFilter)(p), &out)
	})
	return fusionState(out), err
}
func (f *FusionFilter) TimeSync() (NativeFusionTimeSyncStatus, error) {
	var o C.SidereonFusionTimeSyncStatus
	err := fusionCallRead(f.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_time_sync_status((*C.SidereonFusionFilter)(p), &o)
	})
	return NativeFusionTimeSyncStatus{uint64(o.imu_capacity), uint64(o.imu_len), uint64(o.checkpoint_capacity), uint64(o.checkpoint_len), bool(o.has_oldest_imu_epoch_j2000_s), bool(o.has_newest_imu_epoch_j2000_s), bool(o.has_oldest_checkpoint_epoch_j2000_s), bool(o.has_newest_checkpoint_epoch_j2000_s), float64(o.oldest_imu_epoch_j2000_s), float64(o.newest_imu_epoch_j2000_s), float64(o.oldest_checkpoint_epoch_j2000_s), float64(o.newest_checkpoint_epoch_j2000_s)}, err
}
func (f *FusionFilter) Covariance() ([]float64, error) {
	return copyTrackDoubles(f.handle, "fusion covariance", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_covariance((*C.SidereonFusionFilter)(p), o, l, w, r)
	})
}
func (f *FusionFilter) Propagate(sample NativeFusionIMUSample) error {
	s := cFusionSample(sample)
	return fusionCallWrite(f.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_propagate((*C.SidereonFusionFilter)(p), &s)
	})
}
func (f *FusionFilter) Encode() ([]byte, error) {
	return copySurfaceBytes(f.handle, "fusion encoded state", func(p unsafe.Pointer, o *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_encode_state((*C.SidereonFusionFilter)(p), o, l, w, r)
	})
}
func (f *FusionFilter) Restore(data []byte) error {
	p := C.CBytes(data)
	if len(data) > 0 && p == nil {
		return errors.New("sidereon: unable to allocate fusion state")
	}
	defer C.free(p)
	dataLength, err := cSize(len(data), "fusion encoded state length")
	if err != nil {
		return err
	}
	return fusionCallWrite(f.handle, func(x unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_restore_state((*C.SidereonFusionFilter)(x), (*C.uint8_t)(p), dataLength)
	})
}
func (f *FusionFilter) UpdateLoose(m NativeFusionLooseMeasurement) (NativeFusionUpdate, error) {
	p, covarianceLength, err := cFloats(m.Covariance, "fusion loose covariance")
	if err != nil {
		return NativeFusionUpdate{}, err
	}
	defer C.free(p)
	satellitesUsed, err := cSize64(m.SatellitesUsed, "fusion satellite count")
	if err != nil {
		return NativeFusionUpdate{}, err
	}
	v := C.SidereonFusionLooseMeasurement{t_j2000_s: C.double(m.Epoch), position_ecef_m: cFusionArray3(m.Position), has_velocity: C.bool(m.HasVelocity), velocity_ecef_mps: cFusionArray3(m.Velocity), covariance: (*C.double)(p), covariance_len: covarianceLength, satellites_used: satellitesUsed, solution_valid: C.bool(m.SolutionValid), fix_status: C.uint32_t(m.FixStatus)}
	var o C.SidereonFusionUpdate
	err = fusionCallWrite(f.handle, func(x unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_update_loose((*C.SidereonFusionFilter)(x), &v, &o)
	})
	return fusionUpdate(o), err
}
func (f *FusionFilter) UpdateStationary() (NativeFusionUpdate, bool, error) {
	var o C.SidereonFusionUpdate
	var present C.bool
	err := fusionCallWrite(f.handle, func(x unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_update_stationary((*C.SidereonFusionFilter)(x), &o, &present)
	})
	return fusionUpdate(o), bool(present), err
}
func (f *FusionFilter) UpdateNonHolonomic() (NativeFusionUpdate, bool, error) {
	var o C.SidereonFusionUpdate
	var present C.bool
	err := fusionCallWrite(f.handle, func(x unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_update_non_holonomic((*C.SidereonFusionFilter)(x), &o, &present)
	})
	return fusionUpdate(o), bool(present), err
}

func NewFusionRTSHistoryBuilder() (*FusionRTSHistoryBuilder, error) {
	var out *C.SidereonFusionRtsHistoryBuilder
	err := callStatus(func() uint32 { return uint32(C.sidereon_fusion_rts_history_builder_new(&out)) })
	if err != nil {
		return nil, err
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(p unsafe.Pointer) {
		C.sidereon_fusion_rts_history_builder_free((*C.SidereonFusionRtsHistoryBuilder)(p))
	})
	if e != nil {
		return nil, e
	}
	return &FusionRTSHistoryBuilder{handle: h}, nil
}
func NewFusionRTSHistoryBuilderFromFilter(f *FusionFilter) (*FusionRTSHistoryBuilder, error) {
	var out *C.SidereonFusionRtsHistoryBuilder
	var opErr error
	err := f.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_fusion_rts_history_builder_from_filter((*C.SidereonFusionFilter)(p), &out)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(p unsafe.Pointer) {
		C.sidereon_fusion_rts_history_builder_free((*C.SidereonFusionRtsHistoryBuilder)(p))
	})
	if e != nil {
		return nil, e
	}
	return &FusionRTSHistoryBuilder{handle: h}, nil
}
func (b *FusionRTSHistoryBuilder) Close() error {
	if b == nil {
		return nil
	}
	return b.handle.close()
}
func (b *FusionRTSHistoryBuilder) Finish() (*FusionRTSHistory, error) {
	var out *C.SidereonFusionRtsHistory
	var opErr error
	err := b.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_fusion_rts_history_builder_finish((*C.SidereonFusionRtsHistoryBuilder)(p), &out)))
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if opErr != nil {
		return nil, opErr
	}
	h, e := newSurfaceHandle(unsafe.Pointer(out), func(p unsafe.Pointer) { C.sidereon_fusion_rts_history_free((*C.SidereonFusionRtsHistory)(p)) })
	if e != nil {
		return nil, e
	}
	return &FusionRTSHistory{handle: h}, nil
}
func (h *FusionRTSHistory) Close() error {
	if h == nil {
		return nil
	}
	return h.handle.close()
}
func (h *FusionRTSHistory) EpochCount() (int, error) {
	var o C.size_t
	err := fusionCallRead(h.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_rts_history_epoch_count((*C.SidereonFusionRtsHistory)(p), &o)
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(o, "fusion history count")
}
func (h *FusionRTSHistory) Epoch(index int) (NativeFusionRTSEpoch, error) {
	if index < 0 {
		return NativeFusionRTSEpoch{}, errors.New("sidereon: negative fusion history index")
	}
	var o C.SidereonFusionRtsEpoch
	err := fusionCallRead(h.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_rts_history_epoch((*C.SidereonFusionRtsHistory)(p), C.size_t(index), &o)
	})
	return NativeFusionRTSEpoch{float64(o.t_j2000_s), uint64(o.covariance_dimension), uint64(o.augmented_dimension), bool(o.has_transition_from_previous)}, err
}
func (h *FusionRTSHistory) Values(index int, kind uint32) ([]float64, error) {
	if index < 0 {
		return nil, errors.New("sidereon: negative fusion history index")
	}
	return copyTrackDoubles(h.handle, "fusion history values", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		switch kind {
		case 0:
			return C.sidereon_fusion_rts_history_epoch_predicted_position_ecef_m((*C.SidereonFusionRtsHistory)(p), C.size_t(index), o, l, w, r)
		case 1:
			return C.sidereon_fusion_rts_history_epoch_updated_position_ecef_m((*C.SidereonFusionRtsHistory)(p), C.size_t(index), o, l, w, r)
		default:
			return C.sidereon_fusion_rts_history_epoch_transition_from_previous((*C.SidereonFusionRtsHistory)(p), C.size_t(index), o, l, w, r)
		}
	})
}

func FusionLabels(kind, layout, fix uint32) (string, string, string, error) {
	labels := make([]string, 3)
	for i := range labels {
		var value uint32
		var data []byte
		var err error
		switch i {
		case 0:
			value = kind
			data, err = copyNativeBytes("fusion filter label", func(o *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_fusion_filter_kind_label(C.uint32_t(value), o, l, w, r)
			})
		case 1:
			value = layout
			data, err = copyNativeBytes("fusion layout label", func(o *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_fusion_error_state_layout_label(C.uint32_t(value), o, l, w, r)
			})
		default:
			value = fix
			data, err = copyNativeBytes("fusion fix label", func(o *C.uint8_t, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_fusion_gnss_fix_status_label(C.uint32_t(value), o, l, w, r)
			})
		}
		if err != nil {
			return "", "", "", err
		}
		labels[i] = string(data)
	}
	return labels[0], labels[1], labels[2], nil
}
func FusionLayoutDimension(layout uint32) (int, error) {
	var o C.size_t
	err := callStatus(func() uint32 { return uint32(C.sidereon_fusion_error_state_layout_dimension(C.uint32_t(layout), &o)) })
	if err != nil {
		return 0, err
	}
	return sizeTToInt(o, "fusion layout dimension")
}
func FusionIMUPreset(grade uint32) (NativeFusionIMUSpec, error) {
	var o C.SidereonFusionImuSpec
	err := callStatus(func() uint32 { return uint32(C.sidereon_fusion_imu_spec_preset(C.uint32_t(grade), &o)) })
	return NativeFusionIMUSpec{float64(o.accel_vrw_mps_sqrt_s), float64(o.gyro_arw_rad_sqrt_s), float64(o.accel_bias_instab_mps2), float64(o.gyro_bias_instab_rps), float64(o.accel_bias_tau_s), float64(o.gyro_bias_tau_s), bool(o.has_accel_scale_instab_ppm), bool(o.has_gyro_scale_instab_ppm), float64(o.accel_scale_instab_ppm), float64(o.gyro_scale_instab_ppm)}, err
}

func (f *FusionFilter) ConfigureTimeSync(imu, checkpoints uint64) error {
	imuCapacity, err := cSize64(imu, "fusion IMU capacity")
	if err != nil {
		return err
	}
	checkpointCapacity, err := cSize64(checkpoints, "fusion checkpoint capacity")
	if err != nil {
		return err
	}
	return fusionCallWrite(f.handle, func(p unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_configure_time_sync((*C.SidereonFusionFilter)(p), imuCapacity, checkpointCapacity)
	})
}

func (f *FusionFilter) PropagateRecorded(sample NativeFusionIMUSample, history *FusionRTSHistoryBuilder) error {
	if history == nil {
		return ErrClosed
	}
	s := cFusionSample(sample)
	return f.handle.write(func(fp unsafe.Pointer) error {
		return history.handle.write(func(hp unsafe.Pointer) error {
			var err error
			withCThread(func() {
				err = statusErrorLocked(uint32(C.sidereon_fusion_filter_propagate_recorded(
					(*C.SidereonFusionFilter)(fp), &s, (*C.SidereonFusionRtsHistoryBuilder)(hp))))
			})
			return err
		})
	})
}

func cFusionLoose(v NativeFusionLooseMeasurement, p unsafe.Pointer, covarianceLength, satellitesUsed C.size_t) C.SidereonFusionLooseMeasurement {
	return C.SidereonFusionLooseMeasurement{
		t_j2000_s: C.double(v.Epoch), position_ecef_m: cFusionArray3(v.Position),
		has_velocity: C.bool(v.HasVelocity), velocity_ecef_mps: cFusionArray3(v.Velocity),
		covariance: (*C.double)(p), covariance_len: covarianceLength,
		satellites_used: satellitesUsed, solution_valid: C.bool(v.SolutionValid),
		fix_status: C.uint32_t(v.FixStatus),
	}
}

func fusionTimeSyncUpdate(v C.SidereonFusionTimeSyncUpdate) NativeFusionTimeSyncUpdate {
	return NativeFusionTimeSyncUpdate{
		Update: fusionUpdate(v.update), Late: bool(v.late_measurement),
		Replayed:      uint64(v.replayed_imu_segments),
		RestoredEpoch: float64(v.restored_checkpoint_epoch_j2000_s),
		CurrentEpoch:  float64(v.current_epoch_j2000_s),
	}
}

func (f *FusionFilter) UpdateLooseRecorded(m NativeFusionLooseMeasurement, history *FusionRTSHistoryBuilder) (NativeFusionUpdate, error) {
	if history == nil {
		return NativeFusionUpdate{}, ErrClosed
	}
	p, covarianceLength, err := cFloats(m.Covariance, "fusion loose covariance")
	if err != nil {
		return NativeFusionUpdate{}, err
	}
	defer C.free(p)
	satellitesUsed, err := cSize64(m.SatellitesUsed, "fusion satellite count")
	if err != nil {
		return NativeFusionUpdate{}, err
	}
	v := cFusionLoose(m, p, covarianceLength, satellitesUsed)
	var out C.SidereonFusionUpdate
	err = f.handle.write(func(fp unsafe.Pointer) error {
		return history.handle.write(func(hp unsafe.Pointer) error {
			var callErr error
			withCThread(func() {
				callErr = statusErrorLocked(uint32(C.sidereon_fusion_filter_update_loose_recorded(
					(*C.SidereonFusionFilter)(fp), &v, (*C.SidereonFusionRtsHistoryBuilder)(hp), &out)))
			})
			return callErr
		})
	})
	return fusionUpdate(out), err
}

func (f *FusionFilter) UpdateLooseTimeSync(m NativeFusionLooseMeasurement) (NativeFusionTimeSyncUpdate, error) {
	p, covarianceLength, err := cFloats(m.Covariance, "fusion loose covariance")
	if err != nil {
		return NativeFusionTimeSyncUpdate{}, err
	}
	defer C.free(p)
	satellitesUsed, err := cSize64(m.SatellitesUsed, "fusion satellite count")
	if err != nil {
		return NativeFusionTimeSyncUpdate{}, err
	}
	v := cFusionLoose(m, p, covarianceLength, satellitesUsed)
	var out C.SidereonFusionTimeSyncUpdate
	err = fusionCallWrite(f.handle, func(fp unsafe.Pointer) C.enum_SidereonStatus {
		return C.sidereon_fusion_filter_update_loose_time_sync((*C.SidereonFusionFilter)(fp), &v, &out)
	})
	return fusionTimeSyncUpdate(out), err
}

func (f *FusionFilter) UpdateStationaryRecorded(history *FusionRTSHistoryBuilder) (NativeFusionUpdate, bool, error) {
	if history == nil {
		return NativeFusionUpdate{}, false, ErrClosed
	}
	var out C.SidereonFusionUpdate
	var propagated C.bool
	err := f.handle.write(func(fp unsafe.Pointer) error {
		return history.handle.write(func(hp unsafe.Pointer) error {
			var callErr error
			withCThread(func() {
				callErr = statusErrorLocked(uint32(C.sidereon_fusion_filter_update_stationary_recorded(
					(*C.SidereonFusionFilter)(fp), (*C.SidereonFusionRtsHistoryBuilder)(hp), &out, &propagated)))
			})
			return callErr
		})
	})
	return fusionUpdate(out), bool(propagated), err
}

func (f *FusionFilter) UpdateNonHolonomicRecorded(history *FusionRTSHistoryBuilder) (NativeFusionUpdate, bool, error) {
	if history == nil {
		return NativeFusionUpdate{}, false, ErrClosed
	}
	var out C.SidereonFusionUpdate
	var propagated C.bool
	err := f.handle.write(func(fp unsafe.Pointer) error {
		return history.handle.write(func(hp unsafe.Pointer) error {
			var callErr error
			withCThread(func() {
				callErr = statusErrorLocked(uint32(C.sidereon_fusion_filter_update_non_holonomic_recorded(
					(*C.SidereonFusionFilter)(fp), (*C.SidereonFusionRtsHistoryBuilder)(hp), &out, &propagated)))
			})
			return callErr
		})
	})
	return fusionUpdate(out), bool(propagated), err
}

func cFusionTight(v NativeFusionTightEpoch) (C.SidereonFusionTightEpoch, unsafe.Pointer, []*C.char, error) {
	observationCount, err := cSize(len(v.Observations), "tight observation count")
	if err != nil {
		return C.SidereonFusionTightEpoch{}, nil, nil, err
	}
	size, err := checkedNativeAllocationSize(len(v.Observations), unsafe.Sizeof(C.SidereonFusionTightObservation{}))
	if err != nil {
		return C.SidereonFusionTightEpoch{}, nil, nil, err
	}
	var mem unsafe.Pointer
	if len(v.Observations) > 0 {
		mem = C.malloc(C.size_t(size))
		if mem == nil {
			return C.SidereonFusionTightEpoch{}, nil, nil, errors.New("sidereon: unable to allocate tight observations")
		}
	}
	rows := unsafe.Slice((*C.SidereonFusionTightObservation)(mem), len(v.Observations))
	ids := make([]*C.char, len(v.Observations))
	for i, x := range v.Observations {
		if err := rejectEmbeddedNUL(x.SatelliteID, "tight satellite ID"); err != nil {
			C.free(mem)
			freeCStrings(ids)
			return C.SidereonFusionTightEpoch{}, nil, nil, err
		}
		ids[i] = C.CString(x.SatelliteID)
		if ids[i] == nil {
			C.free(mem)
			freeCStrings(ids)
			return C.SidereonFusionTightEpoch{}, nil, nil, errors.New("sidereon: unable to allocate tight satellite ID")
		}
		rows[i].sat_id = ids[i]
		rows[i].pseudorange_m = C.double(x.Pseudorange)
		rows[i].pseudorange_sigma_m = C.double(x.Sigma)
		rows[i].has_range_rate = C.bool(x.HasRangeRate)
		rows[i].range_rate = C.SidereonFusionTightRangeRate{
			measured_range_rate_m_s: C.double(x.RangeRate.Measured), sigma_m_s: C.double(x.RangeRate.Sigma),
			satellite_clock_drift_m_s: C.double(x.RangeRate.ClockDrift),
		}
		rows[i].has_carrier_phase = C.bool(x.HasCarrierPhase)
		rows[i].carrier_phase = C.SidereonFusionTightCarrierPhase{
			phase_range_m: C.double(x.CarrierPhase.PhaseRange), sigma_m: C.double(x.CarrierPhase.Sigma),
			float_ambiguity_m: C.double(x.CarrierPhase.FloatAmbiguity),
		}
		rows[i].ionosphere_delay_m = C.double(x.Ionosphere)
		rows[i].troposphere_delay_m = C.double(x.Troposphere)
	}
	var pointer *C.SidereonFusionTightObservation
	if len(rows) > 0 {
		pointer = &rows[0]
	}
	return C.SidereonFusionTightEpoch{t_j2000_s: C.double(v.Epoch), observations: pointer, observation_count: observationCount}, mem, ids, nil
}

func (f *FusionFilter) tightUpdate(epoch NativeFusionTightEpoch, source unsafe.Pointer, sourceKind uint32, history *FusionRTSHistoryBuilder, mode uint32) (NativeFusionUpdate, NativeFusionTimeSyncUpdate, error) {
	if sourceKind > 1 || mode > 2 {
		return NativeFusionUpdate{}, NativeFusionTimeSyncUpdate{}, errors.New("sidereon: unknown tight-update route")
	}
	e, mem, ids, err := cFusionTight(epoch)
	if err != nil {
		return NativeFusionUpdate{}, NativeFusionTimeSyncUpdate{}, err
	}
	defer C.free(mem)
	defer freeCStrings(ids)
	var update C.SidereonFusionUpdate
	var syncUpdate C.SidereonFusionTimeSyncUpdate
	var cErr error
	call := func(fp, hp unsafe.Pointer) C.enum_SidereonStatus {
		switch sourceKind*10 + mode {
		case 0:
			return C.sidereon_fusion_filter_update_tight_sp3((*C.SidereonFusionFilter)(fp), (*C.SidereonSp3)(source), &e, &update)
		case 1:
			return C.sidereon_fusion_filter_update_tight_sp3_recorded((*C.SidereonFusionFilter)(fp), (*C.SidereonSp3)(source), &e, (*C.SidereonFusionRtsHistoryBuilder)(hp), &update)
		case 2:
			return C.sidereon_fusion_filter_update_tight_sp3_time_sync((*C.SidereonFusionFilter)(fp), (*C.SidereonSp3)(source), &e, &syncUpdate)
		case 10:
			return C.sidereon_fusion_filter_update_tight_broadcast((*C.SidereonFusionFilter)(fp), (*C.SidereonBroadcastEphemeris)(source), &e, &update)
		case 11:
			return C.sidereon_fusion_filter_update_tight_broadcast_recorded((*C.SidereonFusionFilter)(fp), (*C.SidereonBroadcastEphemeris)(source), &e, (*C.SidereonFusionRtsHistoryBuilder)(hp), &update)
		default:
			return C.sidereon_fusion_filter_update_tight_broadcast_time_sync((*C.SidereonFusionFilter)(fp), (*C.SidereonBroadcastEphemeris)(source), &e, &syncUpdate)
		}
	}
	err = f.handle.write(func(fp unsafe.Pointer) error {
		if mode == 1 {
			if history == nil || history.handle == nil {
				return ErrClosed
			}
			return history.handle.read(func(hp unsafe.Pointer) error {
				withCThread(func() { cErr = statusErrorLocked(uint32(call(fp, hp))) })
				return cErr
			})
		}
		withCThread(func() { cErr = statusErrorLocked(uint32(call(fp, nil))) })
		return cErr
	})
	return fusionUpdate(update), fusionTimeSyncUpdate(syncUpdate), err
}

func (f *FusionFilter) UpdateTightSP3(epoch NativeFusionTightEpoch, sp3 *SP3, history *FusionRTSHistoryBuilder, mode uint32) (NativeFusionUpdate, NativeFusionTimeSyncUpdate, error) {
	if sp3 == nil {
		return NativeFusionUpdate{}, NativeFusionTimeSyncUpdate{}, ErrClosed
	}
	var update NativeFusionUpdate
	var syncUpdate NativeFusionTimeSyncUpdate
	var err error
	err = sp3.handle.with(func(p unsafe.Pointer) error {
		update, syncUpdate, err = f.tightUpdate(epoch, p, 0, history, mode)
		return nil
	})
	if err != nil {
		return NativeFusionUpdate{}, NativeFusionTimeSyncUpdate{}, err
	}
	return update, syncUpdate, err
}

func (f *FusionFilter) UpdateTightBroadcast(epoch NativeFusionTightEpoch, b *BroadcastEphemeris, history *FusionRTSHistoryBuilder, mode uint32) (NativeFusionUpdate, NativeFusionTimeSyncUpdate, error) {
	if b == nil {
		return NativeFusionUpdate{}, NativeFusionTimeSyncUpdate{}, ErrClosed
	}
	var update NativeFusionUpdate
	var syncUpdate NativeFusionTimeSyncUpdate
	err := b.resource.with(func(p unsafe.Pointer) error {
		var inner error
		update, syncUpdate, inner = f.tightUpdate(epoch, p, 1, history, mode)
		return inner
	})
	return update, syncUpdate, err
}
