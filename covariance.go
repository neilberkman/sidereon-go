package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// Covariance6 is a value-owned row-major six-by-six covariance matrix.
// Methods delegate validation and transformations to the C library.
type Covariance6 struct {
	Values [6][6]float64
}

// CovarianceValidation contains the C library's matrix validation results.
type CovarianceValidation struct {
	Symmetric            bool
	PositiveSemidefinite bool
}

// CovarianceFrame identifies inertial or radial/transverse/normal covariance
// coordinates. The C header carries these as the repr(C) discriminants.
type CovarianceFrame uint32

const (
	CovarianceInertial CovarianceFrame = 0
	CovarianceRTN      CovarianceFrame = 1
)

// ProcessNoiseKind selects the C process-noise interpretation. Values are
// variances in km²/s³ for the RTN acceleration-PSD form.
type ProcessNoiseKind uint32

const (
	ProcessNoiseNone               ProcessNoiseKind = 0
	ProcessNoiseRTNAccelerationPSD ProcessNoiseKind = 1
)

// CovarianceFromDiagonal builds a covariance from exactly six diagonal
// variances. The input is copied before crossing the cgo boundary.
func CovarianceFromDiagonal(diagonal []float64) (Covariance6, error) {
	values, err := native.CovarianceFromDiagonal(diagonal)
	return Covariance6{Values: values}, publicError(err)
}

// Validate checks symmetry and positive semidefiniteness.
func (c Covariance6) Validate() (CovarianceValidation, error) {
	validation, err := native.CovarianceValidate(c.Values)
	return CovarianceValidation{
		Symmetric:            validation.Symmetric,
		PositiveSemidefinite: validation.PositiveSemidefinite,
	}, publicError(err)
}

// ToMeters converts a kilometre-based covariance to metre units.
func (c Covariance6) ToMeters() (Covariance6, error) {
	values, err := native.CovarianceKmToM(c.Values)
	return Covariance6{Values: values}, publicError(err)
}

// ToKilometers converts a metre-based covariance to kilometre units.
func (c Covariance6) ToKilometers() (Covariance6, error) {
	values, err := native.CovarianceMToKm(c.Values)
	return Covariance6{Values: values}, publicError(err)
}

// Interpolate returns the C library's PSD-safe interpolation between c and
// other at parameter u.
func (c Covariance6) Interpolate(other Covariance6, u float64) (Covariance6, error) {
	values, err := native.CovarianceInterpolate(c.Values, other.Values, u)
	return Covariance6{Values: values}, publicError(err)
}

func (c Covariance6) ECIToRTN(state CartesianState) (Covariance6, error) {
	v, err := native.CovarianceECIToRTN(c.Values, native.NativeCartesianState{EpochS: state.EpochTDBSeconds, PositionKm: state.PositionKm, VelocityKmS: state.VelocityKmPerS})
	return Covariance6{Values: v}, publicError(err)
}
func (c Covariance6) RTNToECI(state CartesianState) (Covariance6, error) {
	v, err := native.CovarianceRTNToECI(c.Values, native.NativeCartesianState{EpochS: state.EpochTDBSeconds, PositionKm: state.PositionKm, VelocityKmS: state.VelocityKmPerS})
	return Covariance6{Values: v}, publicError(err)
}

// CovarianceTransportSegment contains one C-provided state transition matrix
// and the ECI state used to rotate radial/transverse/normal process noise.
type CovarianceTransportSegment struct {
	STM            [6][6]float64
	DTSeconds      float64
	QRotationState CartesianState
}

// ProcessNoise describes C's optional RTN acceleration power spectral density
// in km²/s³. It is applied during covariance transport in the inertial frame.
type ProcessNoise struct {
	Kind                                      ProcessNoiseKind
	RadialKm2S3, TransverseKm2S3, NormalKm2S3 float64
}

// PropagationConfig describes an ECI/TDB propagation request. Position is km,
// velocity is km/s, EpochTDBSeconds is seconds since J2000 TDB, tolerances are
// dimensionless, and step limits are seconds.
type PropagationConfig struct {
	EpochTDBSeconds                                  float64
	PositionKm, VelocityKmPerS                       [3]float64
	ForceModel                                       PropagationForceModel
	Integrator                                       PropagationIntegrator
	AbsTol, RelTol, InitialStepS, MinStepS, MaxStepS float64
	MaxSteps                                         uint32
	MuEnabled                                        bool
	MuKm3S2                                          float64
	HasDrag                                          bool
	Drag                                             DragParameters
	ForceComponents                                  ForceModelComponents
}
type CovarianceNode struct {
	State      CartesianState
	Covariance Covariance6
	Frame      CovarianceFrame
}
type CovarianceEphemeris struct {
	_      noCopy
	handle *native.CovarianceEphemeris
}

func DefaultPropagationConfig() (PropagationConfig, error) {
	v, e := native.PropagationConfigDefault()
	return PropagationConfig{EpochTDBSeconds: v.Epoch, PositionKm: v.Position, VelocityKmPerS: v.Velocity,
		ForceModel: PropagationForceModel(v.ForceModel), Integrator: PropagationIntegrator(v.Integrator),
		AbsTol: v.AbsTol, RelTol: v.RelTol, InitialStepS: v.InitialStep, MinStepS: v.MinStep, MaxStepS: v.MaxStep,
		MaxSteps: v.MaxSteps, MuEnabled: v.MuEnabled, MuKm3S2: v.Mu, HasDrag: v.HasDrag,
		Drag:            DragParameters{BCFactorM2PerKg: v.Drag.BCFactorM2PerKg, Weather: SpaceWeather{F107: v.Drag.Weather.F107, F107A: v.Drag.Weather.F107A, Ap: v.Drag.Weather.Ap}, CutoffAltitudeKm: v.Drag.CutoffAltitudeKm},
		ForceComponents: publicForceComponents(v.ForceComponents)}, publicError(e)
}
func PropagateCovariance(config PropagationConfig, covariance Covariance6, epochs []float64, inputFrame, outputFrame CovarianceFrame, noise ProcessNoise) (*CovarianceEphemeris, error) {
	v, e := native.PropagateCovariance(native.NativePropagationConfig{Epoch: config.EpochTDBSeconds, Position: config.PositionKm, Velocity: config.VelocityKmPerS, ForceModel: uint32(config.ForceModel), Integrator: uint32(config.Integrator), AbsTol: config.AbsTol, RelTol: config.RelTol, InitialStep: config.InitialStepS, MinStep: config.MinStepS, MaxStep: config.MaxStepS, MaxSteps: config.MaxSteps, MuEnabled: config.MuEnabled, Mu: config.MuKm3S2, HasDrag: config.HasDrag, Drag: native.DragParameters{BCFactorM2PerKg: config.Drag.BCFactorM2PerKg, Weather: nativeSpaceWeather(config.Drag.Weather), CutoffAltitudeKm: config.Drag.CutoffAltitudeKm}, ForceComponents: nativeForceComponents(config.ForceComponents)}, covariance.Values, append([]float64(nil), epochs...), uint32(inputFrame), uint32(outputFrame), native.NativeProcessNoise{Kind: uint32(noise.Kind), RadialKm2S3: noise.RadialKm2S3, TransverseKm2S3: noise.TransverseKm2S3, NormalKm2S3: noise.NormalKm2S3})
	if e != nil {
		return nil, publicError(e)
	}
	return &CovarianceEphemeris{handle: v}, nil
}
func (e *CovarianceEphemeris) Close() error {
	if e == nil || e.handle == nil {
		return nil
	}
	return publicError(e.handle.Close())
}
func (e *CovarianceEphemeris) Count() (int, error) {
	if e == nil || e.handle == nil {
		return 0, ErrClosed
	}
	v, x := e.handle.Count()
	return v, publicError(x)
}
func (e *CovarianceEphemeris) CovarianceAt(epoch float64) (Covariance6, error) {
	if e == nil || e.handle == nil {
		return Covariance6{}, ErrClosed
	}
	v, x := e.handle.CovarianceAt(epoch)
	return Covariance6{v}, publicError(x)
}
func (e *CovarianceEphemeris) Nodes() ([]CovarianceNode, error) {
	if e == nil || e.handle == nil {
		return nil, ErrClosed
	}
	v, x := e.handle.Nodes()
	if x != nil {
		return nil, publicError(x)
	}
	out := make([]CovarianceNode, len(v))
	for i, n := range v {
		out[i] = CovarianceNode{CartesianState{EpochTDBSeconds: n.State.EpochS, PositionKm: n.State.PositionKm, VelocityKmPerS: n.State.VelocityKmS}, Covariance6{n.Covariance}, CovarianceFrame(n.Frame)}
	}
	return out, nil
}

func CovarianceTransport(c Covariance6, segments []CovarianceTransportSegment, noise ProcessNoise) ([][6][6]float64, error) {
	s := make([]native.NativeCovarianceTransportSegment, len(segments))
	for i, x := range segments {
		s[i] = native.NativeCovarianceTransportSegment{STM: x.STM, DTSeconds: x.DTSeconds, QRotationState: native.NativeCartesianState{EpochS: x.QRotationState.EpochTDBSeconds, PositionKm: x.QRotationState.PositionKm, VelocityKmS: x.QRotationState.VelocityKmPerS}}
	}
	v, e := native.CovarianceTransport(c.Values, s, native.NativeProcessNoise{Kind: uint32(noise.Kind), RadialKm2S3: noise.RadialKm2S3, TransverseKm2S3: noise.TransverseKm2S3, NormalKm2S3: noise.NormalKm2S3})
	return v, publicError(e)
}
