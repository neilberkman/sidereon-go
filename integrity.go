package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

// ARAIMSatelliteModel contains per-satellite ARAIM integrity model parameters.
type ARAIMSatelliteModel struct {
	// SigmaURAM and SigmaUREM are the nominal URA and URE standard deviations in metres.
	SigmaURAM, SigmaUREM float64
	// HasEffectiveIntegrity and HasEffectiveAccuracy indicate whether the corresponding effective bounds are supplied.
	HasEffectiveIntegrity, HasEffectiveAccuracy bool
	// EffectiveIntegrityM and EffectiveAccuracyM are optional effective integrity and accuracy bounds in metres.
	EffectiveIntegrityM, EffectiveAccuracyM float64
	// NominalBiasM is the nominal range bias in metres; SatelliteFaultProbability is the probability of a satellite fault.
	NominalBiasM, SatelliteFaultProbability float64
}

// ARAIMRow contains one ARAIM observation row, including line-of-sight and integrity terms.
type ARAIMRow struct {
	// SatelliteID identifies the satellite represented by this observation row.
	SatelliteID string
	// LineOfSight is the receiver-to-satellite line-of-sight unit vector in ECEF.
	LineOfSight LineOfSight
	// System identifies the GNSS constellation for the observation.
	System uint32
	// ElevationRad is the elevation rad in radians.
	ElevationRad float64
}

// ARAIMGeometry contains the ARAIM geometry and covariance diagnostics.
type ARAIMGeometry struct {
	// Rows contains a detached copy; nil means this field is absent.
	Rows []ARAIMRow
	// Receiver is the receiver record.
	Receiver Geodetic
	// ClockSystems contains a detached copy; nil means this field is absent.
	ClockSystems []uint32
}

// ARAIMConstellationISM contains constellation-specific integrity support data.
type ARAIMConstellationISM struct {
	// System is the GNSS system identifier.
	System uint32
	// ConstellationFaultProbability is the probability of a constellation fault.
	ConstellationFaultProbability float64
	// DefaultSatellite contains the constellation's default satellite integrity parameters.
	DefaultSatellite ARAIMSatelliteModel
}

// ARAIMSatelliteISM contains satellite-specific integrity support data.
type ARAIMSatelliteISM struct {
	// SatelliteID is the GNSS satellite token for this integrity model row.
	SatelliteID string
	// Model is the selected model.
	Model ARAIMSatelliteModel
}

// ARAIMISM contains the complete ARAIM integrity support message.
type ARAIMISM struct {
	// Constellations contains a detached copy; nil means this field is absent.
	Constellations []ARAIMConstellationISM
	// Satellites contains a detached copy; nil means this field is absent.
	Satellites []ARAIMSatelliteISM
}

// ARAIMAllocation contains allocated protection levels and integrity probabilities.
type ARAIMAllocation struct {
	// PHMITotal, PHMIVertical, and PHMIHorizontal are total, vertical, and horizontal
	// HMI probabilities; PFAVertical and PFAHorizontal are false-alarm probabilities;
	// PThresholdUnmonitored and PEMT are the unmonitored and missed-event probabilities.
	PHMITotal, PHMIVertical, PHMIHorizontal, PFAVertical, PFAHorizontal, PThresholdUnmonitored, PEMT float64
	// MaxFaultOrder is the maximum simultaneous fault order used by the allocation.
	MaxFaultOrder int
}

// ARAIMFaultMode records one modeled ARAIM fault mode and its prior.
type ARAIMFaultMode struct {
	// ExcludedCount is the number of satellites excluded by this fault mode.
	ExcludedCount int
	// HasExcludedConstellation reports whether the has excluded constellation field is present.
	HasExcludedConstellation bool
	// ExcludedConstellation identifies the optionally excluded GNSS constellation.
	ExcludedConstellation uint32
	// Prior is the prior probability assigned to this fault mode.
	Prior float64
	// SigmaIntegrityENU, BiasENU, and ThresholdENU are three-component ENU
	// integrity standard-deviation, bias, and threshold vectors in metres.
	SigmaIntegrityENU, BiasENU, ThresholdENU [3]float64
	// Monitorable reports whether this record is monitorable.
	Monitorable bool
}

// ARAIMSummary contains aggregate ARAIM protection levels and availability.
type ARAIMSummary struct {
	// HPLM and VPLM are horizontal and vertical protection levels in metres;
	// SigmaAccuracyHorizontalM and SigmaAccuracyVerticalM are accuracy sigmas in
	// metres; EMTM is the error bound in metres; PUnmonitored is a probability.
	HPLM, VPLM, SigmaAccuracyHorizontalM, SigmaAccuracyVerticalM, EMTM, PUnmonitored float64
	// Available reports whether protection-level output is available; Availability
	// is the resulting availability probability or fraction.
	Available, Availability bool
	// FaultModeCount is the number of fault modes evaluated for this result.
	FaultModeCount int
}

// ARAIMResult contains the ARAIM solver output and diagnostics.
type ARAIMResult struct {
	_      noCopy
	handle *native.ARAIMResult
}

func nativeARAIMModel(v ARAIMSatelliteModel) native.NativeARAIMSatelliteModel {
	return native.NativeARAIMSatelliteModel{SigmaURA: v.SigmaURAM, SigmaURE: v.SigmaUREM, HasEffectiveInt: v.HasEffectiveIntegrity, EffectiveInt: v.EffectiveIntegrityM, HasEffectiveAcc: v.HasEffectiveAccuracy, EffectiveAcc: v.EffectiveAccuracyM, BNom: v.NominalBiasM, PSat: v.SatelliteFaultProbability}
}
func nativeARAIMGeometry(v ARAIMGeometry) native.NativeARAIMGeometry {
	r := make([]native.NativeARAIMRow, len(v.Rows))
	for i, x := range v.Rows {
		r[i] = native.NativeARAIMRow{SatelliteID: x.SatelliteID, LineOfSight: native.LineOfSight{EX: x.LineOfSight.EX, EY: x.LineOfSight.EY, EZ: x.LineOfSight.EZ}, System: x.System, ElevationRad: x.ElevationRad}
	}
	return native.NativeARAIMGeometry{Rows: r, Receiver: native.Geodetic{LatitudeRad: v.Receiver.LatitudeRad, LongitudeRad: v.Receiver.LongitudeRad, HeightM: v.Receiver.HeightM}, ClockSystems: append([]uint32(nil), v.ClockSystems...)}
}
func nativeARAIMISM(v ARAIMISM) native.NativeARAIMISM {
	c := make([]native.NativeARAIMConstellation, len(v.Constellations))
	for i, x := range v.Constellations {
		c[i] = native.NativeARAIMConstellation{System: x.System, PConst: x.ConstellationFaultProbability, Default: nativeARAIMModel(x.DefaultSatellite)}
	}
	s := make([]native.NativeARAIMSatellite, len(v.Satellites))
	for i, x := range v.Satellites {
		s[i] = native.NativeARAIMSatellite{SatelliteID: x.SatelliteID, Model: nativeARAIMModel(x.Model)}
	}
	return native.NativeARAIMISM{Constellations: c, Satellites: s}
}
func nativeARAIMAllocation(v ARAIMAllocation) native.NativeARAIMAllocation {
	return native.NativeARAIMAllocation{PHMITotal: v.PHMITotal, PHMIVert: v.PHMIVertical, PHMIHor: v.PHMIHorizontal, PFAVert: v.PFAVertical, PFAHor: v.PFAHorizontal, PThresholdUnmonitored: v.PThresholdUnmonitored, PEMT: v.PEMT, MaxFaultOrder: uint64(v.MaxFaultOrder)}
}

// ARAIMAllocationLPV200 returns the LPV-200 ARAIM allocation parameters.
func ARAIMAllocationLPV200() (ARAIMAllocation, error) {
	v, e := native.ARAIMAllocationLPV200()
	if e != nil {
		return ARAIMAllocation{}, publicError(e)
	}
	maxFaultOrder, conversionErr := nativeCountToInt(v.MaxFaultOrder, "ARAIM maximum fault order")
	if conversionErr != nil {
		return ARAIMAllocation{}, conversionErr
	}
	return ARAIMAllocation{v.PHMITotal, v.PHMIVert, v.PHMIHor, v.PFAVert, v.PFAHor, v.PThresholdUnmonitored, v.PEMT, maxFaultOrder}, nil
}

// RunARAIM returns the ARAIM integrity data.
func RunARAIM(geometry ARAIMGeometry, ism ARAIMISM, allocation ARAIMAllocation) (*ARAIMResult, error) {
	if allocation.MaxFaultOrder < 0 {
		return nil, errors.New("sidereon: ARAIM max fault order must not be negative")
	}
	h, e := native.RunARAIM(nativeARAIMGeometry(geometry), nativeARAIMISM(ism), nativeARAIMAllocation(allocation))
	if e != nil {
		return nil, publicError(e)
	}
	return &ARAIMResult{handle: h}, nil
}

// Close releases the native ARAIMResult resource and is safe to call repeatedly.
func (r *ARAIMResult) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Summary returns ARAIM protection-level, integrity, and availability diagnostics.
func (r *ARAIMResult) Summary() (ARAIMSummary, error) {
	if r == nil || r.handle == nil {
		return ARAIMSummary{}, ErrClosed
	}
	v, e := r.handle.Summary()
	if e != nil {
		return ARAIMSummary{}, publicError(e)
	}
	faultModeCount, conversionErr := nativeCountToInt(v.FaultModeCount, "ARAIM fault-mode count")
	if conversionErr != nil {
		return ARAIMSummary{}, conversionErr
	}
	return ARAIMSummary{v.HPLM, v.VPLM, v.SigmaAccHM, v.SigmaAccVM, v.EMTM, v.PUnmonitored, v.Available, v.Availability, faultModeCount}, nil
}

// FaultModes returns the configured ARAIM fault-mode records.
func (r *ARAIMResult) FaultModes() ([]ARAIMFaultMode, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.FaultModes()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]ARAIMFaultMode, len(v))
	for i, x := range v {
		excludedCount, conversionErr := nativeCountToInt(x.ExcludedCount, "ARAIM excluded satellite count")
		if conversionErr != nil {
			return nil, conversionErr
		}
		out[i] = ARAIMFaultMode{excludedCount, x.HasExcludedConstellation, x.ExcludedConstellation, x.Prior, x.SigmaIntENU, x.BiasENU, x.ThresholdENU, x.Monitorable}
	}
	return out, nil
}

// ExcludedSatellites returns the satellite identifiers excluded by the ARAIM run.
func (r *ARAIMResult) ExcludedSatellites(mode int) ([]string, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.Excluded(mode)
	return append([]string(nil), v...), publicError(e)
}

// ReliabilityOptions contains dimensionless false-alarm/missed-detection
// probabilities and the optional noncentrality override.
type ReliabilityOptions struct {
	// Alpha and Beta are the false-alarm and missed-detection probability thresholds.
	Alpha, Beta float64
	// HasLambda0Override reports whether the has lambda0 override field is present.
	HasLambda0Override bool
	// Lambda0Override is the optional noncentrality threshold when HasLambda0Override is true; MinRedundancy is the minimum redundancy required for testing.
	Lambda0Override, MinRedundancy float64
}

// ReliabilityRow contains one reliability-test row and its residual statistics.
type ReliabilityRow struct {
	// SatelliteID is the satellite token for this reliability row.
	SatelliteID string
	// DesignRow contains a detached copy; nil means this field is absent.
	DesignRow []float64
	// SigmaM is the sigma m in metres.
	SigmaM float64
}

// ReliabilityObservation contains one observation reliability and minimum-detectable-bias result.
type ReliabilityObservation struct {
	// SatelliteID is the satellite token for this observation.
	SatelliteID string
	// Redundancy is the redundancy value.
	Redundancy float64
	// HasMDB reports whether the has mdb field is present.
	HasMDB bool
	// MDBM is the mdbm in metres.
	MDBM float64
	// HasExternalENU reports whether the has external enu field is present.
	HasExternalENU bool
	// ExternalENUM is the external enum in metres.
	ExternalENUM [3]float64
	// HasBiasToNoise reports whether the has bias to noise field is present.
	HasBiasToNoise bool
	// BiasToNoise is the dimensionless bias-to-noise ratio when HasBiasToNoise is true.
	BiasToNoise float64
	// Uncheckable reports whether this record is uncheckable.
	Uncheckable bool
}

// ReliabilitySummary contains aggregate reliability counts and test statistics.
type ReliabilitySummary struct {
	// ObservationCount and ParameterCount are the numbers of observations and estimated parameters; DOF is the resulting degrees of freedom.
	ObservationCount, ParameterCount, DOF int
	// SumRedundancy is the total redundancy; Lambda0 is the noncentrality parameter used by the reliability test.
	SumRedundancy, Lambda0 float64
	// HasMaxMDB reports whether the has max mdb field is present.
	HasMaxMDB bool
	// MaxMDBSatelliteID is the satellite token attaining the maximum minimum-detectable bias when HasMaxMDB is true.
	MaxMDBSatelliteID string
	// MaxMDBM is the max mdbm in metres.
	MaxMDBM float64
	// MinRedundancySatelliteID is the satellite token with the minimum redundancy.
	MinRedundancySatelliteID string
	// MinRedundancy is the minimum redundancy.
	MinRedundancy float64
	// UncheckableCount is the number of observations that could not be tested.
	UncheckableCount int
}

// ReliabilityReport contains the complete reliability report and detached observation rows.
type ReliabilityReport struct {
	_      noCopy
	handle *native.ReliabilityReport
}

// DefaultReliabilityOptions returns the default reliability options configuration.
func DefaultReliabilityOptions() (ReliabilityOptions, error) {
	v, e := native.ReliabilityOptionsDefault()
	return ReliabilityOptions{v.Alpha, v.Beta, v.HasLambda0, v.Lambda0, v.MinRedundancy}, publicError(e)
}

// WTestNoncentrality contains the W-test noncentrality parameters.
type WTestNoncentrality struct {
	// Delta0 and Lambda0 are the noncentrality parameters of the W-test.
	Delta0, Lambda0 float64
}

// WTestNoncentralityFor returns the W-test noncentrality parameters.
func WTestNoncentralityFor(alpha, beta float64) (WTestNoncentrality, error) {
	v, e := native.WTestNoncentrality(alpha, beta)
	return WTestNoncentrality{v.Delta0, v.Lambda0}, publicError(e)
}

// ReliabilityFromDesign returns a reliability report from the supplied design.
func ReliabilityFromDesign(rows []ReliabilityRow, options ReliabilityOptions) (*ReliabilityReport, error) {
	r := make([]native.NativeReliabilityRow, len(rows))
	for i, x := range rows {
		r[i] = native.NativeReliabilityRow{ID: x.SatelliteID, Design: append([]float64(nil), x.DesignRow...), Sigma: x.SigmaM}
	}
	h, e := native.ReliabilityDesign(r, native.NativeReliabilityOptions{Alpha: options.Alpha, Beta: options.Beta, HasLambda0: options.HasLambda0Override, Lambda0: options.Lambda0Override, MinRedundancy: options.MinRedundancy})
	if e != nil {
		return nil, publicError(e)
	}
	return &ReliabilityReport{handle: h}, nil
}

// ReliabilityFromARAIM returns a reliability report from the supplied ARAIM result.
func ReliabilityFromARAIM(geometry ARAIMGeometry, ism ARAIMISM, options ReliabilityOptions) (*ReliabilityReport, error) {
	h, e := native.ReliabilityARAIM(nativeARAIMGeometry(geometry), nativeARAIMISM(ism), native.NativeReliabilityOptions{Alpha: options.Alpha, Beta: options.Beta, HasLambda0: options.HasLambda0Override, Lambda0: options.Lambda0Override, MinRedundancy: options.MinRedundancy})
	if e != nil {
		return nil, publicError(e)
	}
	return &ReliabilityReport{handle: h}, nil
}

// Close releases the native ReliabilityReport resource and is safe to call repeatedly.
func (r *ReliabilityReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Summary returns reliability counts and aggregate test statistics.
func (r *ReliabilityReport) Summary() (ReliabilitySummary, error) {
	if r == nil || r.handle == nil {
		return ReliabilitySummary{}, ErrClosed
	}
	v, e := r.handle.Summary()
	if e != nil {
		return ReliabilitySummary{}, publicError(e)
	}
	nobs, conversionErr := nativeCountToInt(v.NObs, "reliability observation count")
	if conversionErr != nil {
		return ReliabilitySummary{}, conversionErr
	}
	nparams, conversionErr := nativeCountToInt(v.NParams, "reliability parameter count")
	if conversionErr != nil {
		return ReliabilitySummary{}, conversionErr
	}
	dof, conversionErr := nativeCountToInt(v.DOF, "reliability degrees of freedom")
	if conversionErr != nil {
		return ReliabilitySummary{}, conversionErr
	}
	uncheckable, conversionErr := nativeCountToInt(v.NUncheckable, "reliability uncheckable count")
	if conversionErr != nil {
		return ReliabilitySummary{}, conversionErr
	}
	return ReliabilitySummary{nobs, nparams, dof, v.SumRedundancy, v.Lambda0, v.HasMaxMDB, v.MaxMDBID, v.MaxMDB, v.MinRedundancyID, v.MinRedundancy, uncheckable}, nil
}

// Observations returns per-observation reliability diagnostics.
func (r *ReliabilityReport) Observations() ([]ReliabilityObservation, error) {
	if r == nil || r.handle == nil {
		return nil, ErrClosed
	}
	v, e := r.handle.Observations()
	if e != nil {
		return nil, publicError(e)
	}
	out := make([]ReliabilityObservation, len(v))
	for i, x := range v {
		out[i] = ReliabilityObservation{x.ID, x.Redundancy, x.HasMDB, x.MDB, x.HasExternal, x.External, x.HasBiasToNoise, x.BiasToNoise, x.Uncheckable}
	}
	return out, nil
}

// RangeFDEOptions contains dimensionless PFA and integer exclusion/redundancy
// limits for the C range-domain FDE solve.
type RangeFDEOptions struct {
	// PFA is the probability of false alarm.
	PFA float64
	// MaxExclusions is the maximum number of exclusions allowed; MinRedundancy is the minimum redundancy required by RangeFDEOptions.
	MaxExclusions, MinRedundancy int
}

// RangeFDERow contains one range fault-detection and exclusion row.
type RangeFDERow struct {
	// ID identifies the range observation represented by this diagnostic row.
	ID string
	// ResidualM is the residual m in metres.
	ResidualM float64
	// DesignRow contains a detached copy; nil means this field is absent.
	DesignRow []float64
	// Weight is the observation weight.
	Weight float64
}

// RangeFDEGlobalTest contains the global range FDE test statistic and threshold.
type RangeFDEGlobalTest struct {
	// WeightedSumSquares is the weighted residual sum of squares.
	WeightedSumSquares float64
	// DOF is the degrees of freedom.
	DOF int64
	// HasThreshold reports whether the has threshold field is present.
	HasThreshold bool
	// Threshold is the configured threshold.
	Threshold float64
	// Testable reports whether the record can be tested; FaultDetected reports whether a fault was detected.
	Testable, FaultDetected bool
}

// RangeFDEDiagnostic contains one range FDE diagnostic and residual.
type RangeFDEDiagnostic struct {
	// ID identifies the range observation represented by this diagnostic row.
	ID string
	// Excluded reports whether this record was excluded.
	Excluded bool
	// PostFitResidualM is the post fit residual m in metres; NormalizedResidual is the dimensionless normalized residual.
	PostFitResidualM, NormalizedResidual float64
}

// RangeFDEOutput contains range FDE corrections and covariance output.
type RangeFDEOutput struct {
	// StateDimension is the state dimension.
	StateDimension int
	// Correction contains a detached copy; nil means this field is absent; Covariance contains a detached copy; nil means this field is absent.
	Correction, Covariance []float64
	// Global reports whether this is a global test.
	Global RangeFDEGlobalTest
	// Iterations is the native solver iteration count.
	Iterations int
	// Excluded reports whether this record was excluded.
	Excluded []string
	// Diagnostics contains a detached copy; nil means this field is absent.
	Diagnostics []RangeFDEDiagnostic
}

// RangeFDEResult contains the complete range FDE result and diagnostics.
type RangeFDEResult struct {
	_      noCopy
	handle *native.RangeFDEResult
}

// DefaultRangeFDEOptions returns the default Range FDE options configuration.
func DefaultRangeFDEOptions() (RangeFDEOptions, error) {
	v, e := native.RangeFDEOptionsDefault()
	if e != nil {
		return RangeFDEOptions{}, publicError(e)
	}
	maxExclusions, conversionErr := nativeCountToInt(v.MaxExclusions, "range FDE maximum exclusions")
	if conversionErr != nil {
		return RangeFDEOptions{}, conversionErr
	}
	minRedundancy, conversionErr := nativeCountToInt(v.MinRedundancy, "range FDE minimum redundancy")
	if conversionErr != nil {
		return RangeFDEOptions{}, conversionErr
	}
	return RangeFDEOptions{v.PFA, maxExclusions, minRedundancy}, nil
}

// RunRangeFDE returns the range fault-detection and exclusion data.
func RunRangeFDE(rows []RangeFDERow, options RangeFDEOptions) (*RangeFDEResult, error) {
	if options.MaxExclusions < 0 || options.MinRedundancy < 0 {
		return nil, errors.New("sidereon: range FDE limits must not be negative")
	}
	r := make([]native.NativeRangeFDERow, len(rows))
	for i, x := range rows {
		r[i] = native.NativeRangeFDERow{ID: x.ID, Residual: x.ResidualM, Design: append([]float64(nil), x.DesignRow...), Weight: x.Weight}
	}
	h, e := native.RangeFDE(r, native.NativeRangeFDEOptions{PFA: options.PFA, MaxExclusions: uint64(options.MaxExclusions), MinRedundancy: uint64(options.MinRedundancy)})
	if e != nil {
		return nil, publicError(e)
	}
	return &RangeFDEResult{handle: h}, nil
}

// Close releases the native RangeFDEResult resource and is safe to call repeatedly.
func (r *RangeFDEResult) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}

// Output returns the range-FDE corrections and covariance output.
func (r *RangeFDEResult) Output() (RangeFDEOutput, error) {
	if r == nil || r.handle == nil {
		return RangeFDEOutput{}, ErrClosed
	}
	v, e := r.handle.Output()
	if e != nil {
		return RangeFDEOutput{}, publicError(e)
	}
	stateDimension, conversionErr := nativeCountToInt(v.StateDimension, "range FDE state dimension")
	if conversionErr != nil {
		return RangeFDEOutput{}, conversionErr
	}
	iterations, conversionErr := nativeCountToInt(v.Iterations, "range FDE iteration count")
	if conversionErr != nil {
		return RangeFDEOutput{}, conversionErr
	}
	out := RangeFDEOutput{stateDimension, append([]float64(nil), v.Correction...), append([]float64(nil), v.Covariance...), RangeFDEGlobalTest{v.Global.WeightedSumSquares, v.Global.DOF, v.Global.HasThreshold, v.Global.Threshold, v.Global.Testable, v.Global.FaultDetected}, iterations, append([]string(nil), v.Excluded...), make([]RangeFDEDiagnostic, len(v.Diagnostics))}
	for i, x := range v.Diagnostics {
		out.Diagnostics[i] = RangeFDEDiagnostic{x.ID, x.Excluded, x.PostFitResidual, x.NormalizedResidual}
	}
	return out, nil
}
