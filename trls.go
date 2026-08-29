package sidereon

import (
	"errors"

	"github.com/neilberkman/sidereon-go/internal/native"
)

type TRLSKind uint32

const (
	TRLSLinear      TRLSKind = TRLSKind(native.TrlsLinearValue)
	TRLSPolynomial  TRLSKind = TRLSKind(native.TrlsPolynomialValue)
	TRLSExponential TRLSKind = TRLSKind(native.TrlsExponentialValue)
)

type TRLSLoss uint32

const (
	TRLSLossLinear TRLSLoss = TRLSLoss(native.TrlsLossLinearValue)
	TRLSLossSoftL1 TRLSLoss = TRLSLoss(native.TrlsLossSoftL1Value)
	TRLSLossHuber  TRLSLoss = TRLSLoss(native.TrlsLossHuberValue)
	TRLSLossCauchy TRLSLoss = TRLSLoss(native.TrlsLossCauchyValue)
	TRLSLossArctan TRLSLoss = TRLSLoss(native.TrlsLossArctanValue)
)

type TRLSXScale uint32

const (
	TRLSXScaleUnit   TRLSXScale = TRLSXScale(native.TrlsXScaleUnitValue)
	TRLSXScaleJac    TRLSXScale = TRLSXScale(native.TrlsXScaleJacValue)
	TRLSXScaleValues TRLSXScale = TRLSXScale(native.TrlsXScaleValuesValue)
)

type TRLSBackend uint32

const (
	TRLSBackendNative     TRLSBackend = TRLSBackend(native.TrlsBackendNativeValue)
	TRLSBackendHostLAPACK TRLSBackend = TRLSBackend(native.TrlsBackendHostLAPACKValue)
)

// TRLSProblem describes one C-owned residual model. Arrays are copied at
// SolveTRLS and remain caller-owned thereafter. Linear A is row-major m by n;
// polynomial/exponential T and Y have equal length.
type TRLSProblem struct {
	Kind             TRLSKind
	A, B, T, Y, X0   []float64
	M, N, Degree     int
	Loss             TRLSLoss
	FScale           float64
	XScaleMode       TRLSXScale
	XScaleValues     []float64
	MaxNFEV          int64
	FTOL, XTOL, GTOL float64
	Backend          TRLSBackend
}

// DefaultTRLSProblem returns the C library's SciPy-compatible defaults for a
// residual kind. Populate the kind-specific data arrays before solving.
func DefaultTRLSProblem(kind TRLSKind) (TRLSProblem, error) {
	value, err := native.DefaultTRLSProblem(uint32(kind))
	if err != nil {
		return TRLSProblem{}, publicError(err)
	}
	return TRLSProblem{
		Kind:       TRLSKind(value.Kind),
		Loss:       TRLSLoss(value.Loss),
		FScale:     value.FScale,
		XScaleMode: TRLSXScale(value.XScaleMode),
		MaxNFEV:    value.MaxNFEV,
		FTOL:       value.FTOL,
		XTOL:       value.XTOL,
		GTOL:       value.GTOL,
		Backend:    TRLSBackend(value.Backend),
	}, nil
}

type TRLSSummary struct {
	Cost, Optimality float64
	NFEV, NJEV       uint64
	Status           int32
	Success          bool
	N, M             uint64
}
type TRLSSolution struct {
	_      noCopy
	handle *native.TRLSSolution
}
type TRLSDropOne struct {
	_      noCopy
	handle *native.TRLSDropOne
}

func nativeTRLSProblem(value TRLSProblem) native.NativeTRLSProblem {
	return native.NativeTRLSProblem{Kind: uint32(value.Kind), A: append([]float64(nil), value.A...), B: append([]float64(nil), value.B...), T: append([]float64(nil), value.T...), Y: append([]float64(nil), value.Y...), X0: append([]float64(nil), value.X0...), M: uint64(value.M), N: uint64(value.N), Degree: uint64(value.Degree), Loss: uint32(value.Loss), FScale: value.FScale, XScaleMode: uint32(value.XScaleMode), XScaleValues: append([]float64(nil), value.XScaleValues...), MaxNFEV: value.MaxNFEV, FTOL: value.FTOL, XTOL: value.XTOL, GTOL: value.GTOL, Backend: uint32(value.Backend)}
}
func toTRLSSummary(v native.NativeTRLSSummary) TRLSSummary {
	return TRLSSummary{v.Cost, v.Optimality, v.NFEV, v.NJEV, v.Status, v.Success, v.N, v.M}
}

func trlsProduct(a, b int) (int, bool) {
	if a < 0 || b < 0 || (a != 0 && b > int(^uint(0)>>1)/a) {
		return 0, false
	}
	return a * b, true
}

func validateTRLSProblem(problem TRLSProblem) error {
	if problem.M < 0 || problem.N < 0 || problem.Degree < 0 {
		return errors.New("sidereon: TRLS dimensions must not be negative")
	}
	parameters := problem.N
	switch problem.Kind {
	case TRLSLinear:
		expected, ok := trlsProduct(problem.M, problem.N)
		if !ok || len(problem.A) != expected || len(problem.B) != problem.M || len(problem.X0) != problem.N {
			return errors.New("sidereon: linear TRLS arrays do not match M and N")
		}
	case TRLSPolynomial:
		if problem.Degree == int(^uint(0)>>1) {
			return errors.New("sidereon: polynomial degree is too large")
		}
		parameters = problem.Degree + 1
		if len(problem.T) != len(problem.Y) || len(problem.X0) != parameters {
			return errors.New("sidereon: polynomial TRLS arrays do not match degree")
		}
	case TRLSExponential:
		parameters = 3
		if len(problem.T) != len(problem.Y) || len(problem.X0) != parameters {
			return errors.New("sidereon: exponential TRLS arrays do not match three parameters")
		}
	default:
		return errors.New("sidereon: unknown TRLS residual kind")
	}
	if problem.XScaleMode == TRLSXScaleValues && len(problem.XScaleValues) != parameters {
		return errors.New("sidereon: TRLS x-scale values do not match parameter count")
	}
	return nil
}

func SolveTRLS(problem TRLSProblem) (*TRLSSolution, error) {
	if err := validateTRLSProblem(problem); err != nil {
		return nil, err
	}
	v, e := native.SolveTRLS(nativeTRLSProblem(problem), false)
	if e != nil {
		return nil, publicError(e)
	}
	return &TRLSSolution{handle: v.(*native.TRLSSolution)}, nil
}
func SolveTRLSDropOne(problem TRLSProblem) (*TRLSDropOne, error) {
	if err := validateTRLSProblem(problem); err != nil {
		return nil, err
	}
	v, e := native.SolveTRLS(nativeTRLSProblem(problem), true)
	if e != nil {
		return nil, publicError(e)
	}
	return &TRLSDropOne{handle: v.(*native.TRLSDropOne)}, nil
}
func (s *TRLSSolution) Close() error {
	if s == nil || s.handle == nil {
		return nil
	}
	return publicError(s.handle.Close())
}
func (s *TRLSSolution) Summary() (TRLSSummary, error) {
	if s == nil || s.handle == nil {
		return TRLSSummary{}, ErrClosed
	}
	v, e := s.handle.Summary()
	return toTRLSSummary(v), publicError(e)
}
func (s *TRLSSolution) X() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(0)
	return append([]float64(nil), v...), publicError(e)
}
func (s *TRLSSolution) Residuals() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(1)
	return append([]float64(nil), v...), publicError(e)
}
func (s *TRLSSolution) Jacobian() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(2)
	return append([]float64(nil), v...), publicError(e)
}
func (s *TRLSSolution) Gradient() ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	v, e := s.handle.Values(3)
	return append([]float64(nil), v...), publicError(e)
}
func (d *TRLSDropOne) Close() error {
	if d == nil || d.handle == nil {
		return nil
	}
	return publicError(d.handle.Close())
}
func (d *TRLSDropOne) Count() (int, error) {
	if d == nil || d.handle == nil {
		return 0, ErrClosed
	}
	v, e := d.handle.Count()
	return v, publicError(e)
}
func (d *TRLSDropOne) CostDeltas() ([]float64, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	v, e := d.handle.Values()
	return append([]float64(nil), v...), publicError(e)
}
func (d *TRLSDropOne) X(index int) ([]float64, error) {
	if d == nil || d.handle == nil {
		return nil, ErrClosed
	}
	v, e := d.handle.X(index)
	return append([]float64(nil), v...), publicError(e)
}
func (d *TRLSDropOne) BaseSummary() (TRLSSummary, error) {
	if d == nil || d.handle == nil {
		return TRLSSummary{}, ErrClosed
	}
	v, e := d.handle.BaseSummary()
	return toTRLSSummary(v), publicError(e)
}
func (d *TRLSDropOne) Summary(index int) (TRLSSummary, error) {
	if d == nil || d.handle == nil {
		return TRLSSummary{}, ErrClosed
	}
	v, e := d.handle.Summary(index)
	return toTRLSSummary(v), publicError(e)
}
