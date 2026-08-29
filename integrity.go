package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

type ARAIMSatelliteModel struct {
	SigmaURAM, SigmaUREM                        float64
	HasEffectiveIntegrity, HasEffectiveAccuracy bool
	EffectiveIntegrityM, EffectiveAccuracyM     float64
	NominalBiasM, SatelliteFaultProbability     float64
}
type ARAIMRow struct {
	SatelliteID  string
	LineOfSight  LineOfSight
	System       uint32
	ElevationRad float64
}
type ARAIMGeometry struct {
	Rows         []ARAIMRow
	Receiver     Geodetic
	ClockSystems []uint32
}
type ARAIMConstellationISM struct {
	System                        uint32
	ConstellationFaultProbability float64
	DefaultSatellite              ARAIMSatelliteModel
}
type ARAIMSatelliteISM struct {
	SatelliteID string
	Model       ARAIMSatelliteModel
}
type ARAIMISM struct {
	Constellations []ARAIMConstellationISM
	Satellites     []ARAIMSatelliteISM
}
type ARAIMAllocation struct {
	PHMITotal, PHMIVertical, PHMIHorizontal, PFAVertical, PFAHorizontal, PThresholdUnmonitored, PEMT float64
	MaxFaultOrder                                                                                    int
}
type ARAIMFaultMode struct {
	ExcludedCount                            int
	HasExcludedConstellation                 bool
	ExcludedConstellation                    uint32
	Prior                                    float64
	SigmaIntegrityENU, BiasENU, ThresholdENU [3]float64
	Monitorable                              bool
}
type ARAIMSummary struct {
	HPLM, VPLM, SigmaAccuracyHorizontalM, SigmaAccuracyVerticalM, EMTM, PUnmonitored float64
	Available, Availability                                                          bool
	FaultModeCount                                                                   int
}
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
func (r *ARAIMResult) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}
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
	Alpha, Beta                    float64
	HasLambda0Override             bool
	Lambda0Override, MinRedundancy float64
}
type ReliabilityRow struct {
	SatelliteID string
	DesignRow   []float64
	SigmaM      float64
}
type ReliabilityObservation struct {
	SatelliteID    string
	Redundancy     float64
	HasMDB         bool
	MDBM           float64
	HasExternalENU bool
	ExternalENUM   [3]float64
	HasBiasToNoise bool
	BiasToNoise    float64
	Uncheckable    bool
}
type ReliabilitySummary struct {
	ObservationCount, ParameterCount, DOF int
	SumRedundancy, Lambda0                float64
	HasMaxMDB                             bool
	MaxMDBSatelliteID                     string
	MaxMDBM                               float64
	MinRedundancySatelliteID              string
	MinRedundancy                         float64
	UncheckableCount                      int
}
type ReliabilityReport struct {
	_      noCopy
	handle *native.ReliabilityReport
}

func DefaultReliabilityOptions() (ReliabilityOptions, error) {
	v, e := native.ReliabilityOptionsDefault()
	return ReliabilityOptions{v.Alpha, v.Beta, v.HasLambda0, v.Lambda0, v.MinRedundancy}, publicError(e)
}

type WTestNoncentrality struct{ Delta0, Lambda0 float64 }

func WTestNoncentralityFor(alpha, beta float64) (WTestNoncentrality, error) {
	v, e := native.WTestNoncentrality(alpha, beta)
	return WTestNoncentrality{v.Delta0, v.Lambda0}, publicError(e)
}
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
func ReliabilityFromARAIM(geometry ARAIMGeometry, ism ARAIMISM, options ReliabilityOptions) (*ReliabilityReport, error) {
	h, e := native.ReliabilityARAIM(nativeARAIMGeometry(geometry), nativeARAIMISM(ism), native.NativeReliabilityOptions{Alpha: options.Alpha, Beta: options.Beta, HasLambda0: options.HasLambda0Override, Lambda0: options.Lambda0Override, MinRedundancy: options.MinRedundancy})
	if e != nil {
		return nil, publicError(e)
	}
	return &ReliabilityReport{handle: h}, nil
}
func (r *ReliabilityReport) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}
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
	PFA                          float64
	MaxExclusions, MinRedundancy int
}
type RangeFDERow struct {
	ID        string
	ResidualM float64
	DesignRow []float64
	Weight    float64
}
type RangeFDEGlobalTest struct {
	WeightedSumSquares      float64
	DOF                     int64
	HasThreshold            bool
	Threshold               float64
	Testable, FaultDetected bool
}
type RangeFDEDiagnostic struct {
	ID                                   string
	Excluded                             bool
	PostFitResidualM, NormalizedResidual float64
}
type RangeFDEOutput struct {
	StateDimension         int
	Correction, Covariance []float64
	Global                 RangeFDEGlobalTest
	Iterations             int
	Excluded               []string
	Diagnostics            []RangeFDEDiagnostic
}
type RangeFDEResult struct {
	_      noCopy
	handle *native.RangeFDEResult
}

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
func (r *RangeFDEResult) Close() error {
	if r == nil || r.handle == nil {
		return nil
	}
	return publicError(r.handle.Close())
}
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
