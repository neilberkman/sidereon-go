package sidereon

import (
	"fmt"

	"sidereon.dev/go/v3/internal/native"
)

// StationTideConstants selects the diurnal-band coefficients for solid Earth tide displacement.
type StationTideConstants uint32

const (
	// StationTideConventions uses the IERS Conventions (2010) coefficients with the corrected P1 sign.
	StationTideConventions StationTideConstants = iota
	// StationTideIERSRoutine uses the distributed DEHANTTIDEINEL diurnal-band coefficients.
	StationTideIERSRoutine
)

// StationTideValidityMode selects strict validation or permitted degradation for tide calculations.
type StationTideValidityMode uint32

const (
	// StationTideValidityStrict requires the requested epoch to satisfy model-coverage checks.
	StationTideValidityStrict StationTideValidityMode = iota
	// StationTideValidityPermissive allows supported degradation outside model coverage.
	StationTideValidityPermissive
)

// StationTideDegradeReason records why a station-tide result used degraded inputs.
type StationTideDegradeReason uint32

const (
	// StationTideNotDegraded means the tide calculation used full-coverage inputs.
	StationTideNotDegraded StationTideDegradeReason = iota
	// StationTideBeforeCoverage means the epoch precedes the available tide-model coverage.
	StationTideBeforeCoverage
	// StationTideAfterCoverage means the epoch follows the available tide-model coverage.
	StationTideAfterCoverage
)

// StationTideEpoch identifies one tide-evaluation epoch and its time scale.
type StationTideEpoch struct {
	Year, Month, Day, Hour, Minute int
	Second                         float64
	HasPolarMotion                 bool
	XPArcsec, YPArcsec             float64
}

// StationTideOptions configures station coordinates, conventions, and validity handling.
type StationTideOptions struct {
	SolidEarthTide, PoleTide bool
	OceanLoading             *OceanLoadingBLQ
	Constants                StationTideConstants
	Validity                 StationTideValidityMode
}

// StationTideDisplacement contains the predicted three-component station displacement.
type StationTideDisplacement struct {
	ECEFM, SolidEarthTideECEFM, PoleTideECEFM, OceanLoadingECEFM [3]float64
	HasSolidEarthTide, HasPoleTide, HasOceanLoading              bool
	DegradeReason                                                StationTideDegradeReason
}

// StationTideError preserves a typed station-tide failure and its cause.
type StationTideError struct {
	Kind, NestedKind, SunMoonCause, InputKind uint32
	DegradeReason                             StationTideDegradeReason
	HasField, HasReason                       bool
	Field, Reason                             string
	Cause                                     error
}

// Error returns the StationTideError message.
func (err *StationTideError) Error() string {
	if err == nil {
		return "sidereon: station tide failed"
	}
	if err.Cause != nil {
		return fmt.Sprintf("sidereon: station tide failed: %v", err.Cause)
	}
	return fmt.Sprintf("sidereon: station tide failed (kind %d)", err.Kind)
}

// Unwrap returns the underlying cause when one is retained.
func (err *StationTideError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

func checkedStationTideInt32(value int, label string) (int32, error) {
	if int64(value) < -1<<31 || int64(value) > 1<<31-1 {
		return 0, fmt.Errorf("sidereon: %s %d does not fit in C int32", label, value)
	}
	return int32(value), nil
}

func nativeStationTideEpoch(value StationTideEpoch) (native.NativeStationTideEpoch, error) {
	year, err := checkedStationTideInt32(value.Year, "station tide year")
	if err != nil {
		return native.NativeStationTideEpoch{}, err
	}
	month, err := checkedStationTideInt32(value.Month, "station tide month")
	if err != nil {
		return native.NativeStationTideEpoch{}, err
	}
	day, err := checkedStationTideInt32(value.Day, "station tide day")
	if err != nil {
		return native.NativeStationTideEpoch{}, err
	}
	hour, err := checkedStationTideInt32(value.Hour, "station tide hour")
	if err != nil {
		return native.NativeStationTideEpoch{}, err
	}
	minute, err := checkedStationTideInt32(value.Minute, "station tide minute")
	if err != nil {
		return native.NativeStationTideEpoch{}, err
	}
	return native.NativeStationTideEpoch{Year: year, Month: month, Day: day, Hour: hour, Minute: minute, Second: value.Second, HasPolarMotion: value.HasPolarMotion, XPArcsec: value.XPArcsec, YPArcsec: value.YPArcsec}, nil
}

func nativeStationTideOptions(value StationTideOptions) native.NativeStationTideOptions {
	var loading *native.OceanLoadingBLQ
	if value.OceanLoading != nil {
		loading = &native.OceanLoadingBLQ{AmplitudeM: value.OceanLoading.AmplitudeM, PhaseDeg: value.OceanLoading.PhaseDeg}
	}
	return native.NativeStationTideOptions{SolidEarthTide: value.SolidEarthTide, PoleTide: value.PoleTide, OceanLoading: loading, Constants: uint32(value.Constants), ValidityMode: uint32(value.Validity)}
}

func stationTideError(value native.NativeStationTideError, cause error) error {
	if cause == nil && value.Kind == 0 {
		return nil
	}
	return &StationTideError{Kind: value.Kind, NestedKind: value.NestedKind, SunMoonCause: value.SunMoonCause, InputKind: value.InputKind, DegradeReason: StationTideDegradeReason(value.DegradeReason), HasField: value.HasField, HasReason: value.HasReason, Field: value.Field, Reason: value.Reason, Cause: cause}
}

func publicStationTideDisplacement(value native.NativeStationTideDisplacement) StationTideDisplacement {
	return StationTideDisplacement{ECEFM: value.ECEFM, SolidEarthTideECEFM: value.SolidEarthTideECEFM, PoleTideECEFM: value.PoleTideECEFM, OceanLoadingECEFM: value.OceanLoadingECEFM, HasSolidEarthTide: value.HasSolidEarthTide, HasPoleTide: value.HasPoleTide, HasOceanLoading: value.HasOceanLoading, DegradeReason: StationTideDegradeReason(value.DegradeReason)}
}

// StationTideConstantsDefault returns the default pinned tide coefficients and convention.
func StationTideConstantsDefault() StationTideConstants {
	return StationTideConstants(native.StationTideConstantsDefault())
}

// StationTideDisplace calculates solid Earth tide displacement for one station and epoch.
func StationTideDisplace(positionECEFM [3]float64, epoch StationTideEpoch, options StationTideOptions) (StationTideDisplacement, error) {
	nativeEpoch, err := nativeStationTideEpoch(epoch)
	if err != nil {
		return StationTideDisplacement{}, err
	}
	value, typed, operationErr := native.StationTideDisplacement(positionECEFM, nativeEpoch, nativeStationTideOptions(options))
	return publicStationTideDisplacement(value), stationTideError(typed, publicError(operationErr))
}

// StationTideBatchRow pairs a requested station/epoch with its calculated displacement.
type StationTideBatchRow struct {
	Status       uint32
	Displacement StationTideDisplacement
	Error        *StationTideError
}

// StationTideDisplaceBatch evaluates station displacement for a batch of requested epochs.
func StationTideDisplaceBatch(positionECEFM [3]float64, epochs []StationTideEpoch, options StationTideOptions) ([]StationTideBatchRow, error) {
	nativeEpochs := make([]native.NativeStationTideEpoch, len(epochs))
	for index, epoch := range epochs {
		value, err := nativeStationTideEpoch(epoch)
		if err != nil {
			return nil, err
		}
		nativeEpochs[index] = value
	}
	values, err := native.StationTideDisplacementBatch(positionECEFM, nativeEpochs, nativeStationTideOptions(options))
	if err != nil {
		return nil, publicError(err)
	}
	rows := make([]StationTideBatchRow, len(values))
	for index, value := range values {
		rows[index] = StationTideBatchRow{Status: value.Status, Displacement: publicStationTideDisplacement(value.Displacement)}
		if value.Error.Kind != 0 {
			rows[index].Error = &StationTideError{Kind: value.Error.Kind, NestedKind: value.Error.NestedKind, SunMoonCause: value.Error.SunMoonCause, InputKind: value.Error.InputKind, DegradeReason: StationTideDegradeReason(value.Error.DegradeReason), HasField: value.Error.HasField, HasReason: value.Error.HasReason, Field: value.Error.Field, Reason: value.Error.Reason}
		}
	}
	return rows, nil
}

// SolidEarthTideWithConstants calculates displacement using the supplied tide coefficients.
func SolidEarthTideWithConstants(stationECEFM [3]float64, year, month, day int, fractionalHour float64, sunECEFM, moonECEFM [3]float64, constants StationTideConstants) ([3]float64, error) {
	y, err := checkedStationTideInt32(year, "year")
	if err != nil {
		return [3]float64{}, err
	}
	m, err := checkedStationTideInt32(month, "month")
	if err != nil {
		return [3]float64{}, err
	}
	d, err := checkedStationTideInt32(day, "day")
	if err != nil {
		return [3]float64{}, err
	}
	value, err := native.SolidEarthTideWithConstants(stationECEFM, y, m, d, fractionalHour, sunECEFM, moonECEFM, uint32(constants))
	return value, publicError(err)
}
