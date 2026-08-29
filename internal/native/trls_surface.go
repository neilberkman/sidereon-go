//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#include <sidereon.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type NativeTRLSProblem struct {
	Kind             uint32
	A, B, T, Y, X0   []float64
	M, N, Degree     uint64
	Loss             uint32
	FScale           float64
	XScaleMode       uint32
	XScaleValues     []float64
	MaxNFEV          int64
	FTOL, XTOL, GTOL float64
	Backend          uint32
}

type TRLSSolution struct {
	_      noCopy
	handle *surfaceHandle
}
type TRLSDropOne struct {
	_      noCopy
	handle *surfaceHandle
}

type NativeTRLSSummary struct {
	Cost, Optimality float64
	NFEV, NJEV       uint64
	Status           int32
	Success          bool
	N, M             uint64
}

func DefaultTRLSProblem(kind uint32) (NativeTRLSProblem, error) {
	var problem C.SidereonDataProblem
	if err := callStatus(func() uint32 {
		return uint32(C.sidereon_data_problem_init(C.uint32_t(kind), &problem))
	}); err != nil {
		return NativeTRLSProblem{}, err
	}
	return NativeTRLSProblem{
		Kind:       uint32(problem.kind),
		Loss:       uint32(problem.loss),
		FScale:     float64(problem.f_scale),
		XScaleMode: uint32(problem.x_scale_mode),
		MaxNFEV:    int64(problem.max_nfev),
		FTOL:       float64(problem.ftol),
		XTOL:       float64(problem.xtol),
		GTOL:       float64(problem.gtol),
		Backend:    uint32(problem.backend),
	}, nil
}

func trlsSummary(value C.SidereonTrlsSummary) NativeTRLSSummary {
	return NativeTRLSSummary{float64(value.cost), float64(value.optimality), uint64(value.nfev), uint64(value.njev), int32(value.status), bool(value.success), uint64(value.n), uint64(value.m)}
}

func nativeTRLSProblem(value NativeTRLSProblem) (C.SidereonDataProblem, []unsafe.Pointer, error) {
	m, err := cSize64(value.M, "TRLS M")
	if err != nil {
		return C.SidereonDataProblem{}, nil, err
	}
	n, err := cSize64(value.N, "TRLS N")
	if err != nil {
		return C.SidereonDataProblem{}, nil, err
	}
	degree, err := cSize64(value.Degree, "TRLS degree")
	if err != nil {
		return C.SidereonDataProblem{}, nil, err
	}
	var problem C.SidereonDataProblem
	if err := callStatus(func() uint32 { return uint32(C.sidereon_data_problem_init(C.uint32_t(value.Kind), &problem)) }); err != nil {
		return C.SidereonDataProblem{}, nil, err
	}
	problem.kind = C.uint32_t(value.Kind)
	problem.m = m
	problem.n = n
	problem.degree = degree
	problem.loss = C.uint32_t(value.Loss)
	problem.f_scale = C.double(value.FScale)
	problem.x_scale_mode = C.uint32_t(value.XScaleMode)
	problem.max_nfev = C.int64_t(value.MaxNFEV)
	problem.ftol = C.double(value.FTOL)
	problem.xtol = C.double(value.XTOL)
	problem.gtol = C.double(value.GTOL)
	problem.backend = C.uint32_t(value.Backend)
	values := [][]float64{value.A, value.B, value.T, value.Y, value.X0, value.XScaleValues}
	pointers := make([]unsafe.Pointer, len(values))
	for i, data := range values {
		pointer, length, err := cFloats(data, "TRLS problem array")
		if err != nil {
			for _, allocated := range pointers {
				if allocated != nil {
					C.free(allocated)
				}
			}
			return C.SidereonDataProblem{}, nil, err
		}
		pointers[i] = pointer
		switch i {
		case 0:
			problem.a, problem.a_len = (*C.double)(pointer), length
		case 1:
			problem.b, problem.b_len = (*C.double)(pointer), length
		case 2:
			problem.t, problem.t_len = (*C.double)(pointer), length
		case 3:
			problem.y, problem.y_len = (*C.double)(pointer), length
		case 4:
			problem.x0, problem.x0_len = (*C.double)(pointer), length
		case 5:
			problem.x_scale_values, problem.x_scale_values_len = (*C.double)(pointer), length
		}
	}
	return problem, pointers, nil
}

func newTRLSSolution(pointer *C.SidereonTrlsSolution) (*TRLSSolution, error) {
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) { C.sidereon_trls_solution_free((*C.SidereonTrlsSolution)(p)) })
	if err != nil {
		return nil, err
	}
	return &TRLSSolution{handle: h}, nil
}
func newTRLSDropOne(pointer *C.SidereonTrlsDropOne) (*TRLSDropOne, error) {
	h, err := newSurfaceHandle(unsafe.Pointer(pointer), func(p unsafe.Pointer) { C.sidereon_trls_drop_one_free((*C.SidereonTrlsDropOne)(p)) })
	if err != nil {
		return nil, err
	}
	return &TRLSDropOne{handle: h}, nil
}

func SolveTRLS(value NativeTRLSProblem, dropOne bool) (any, error) {
	problem, pointers, err := nativeTRLSProblem(value)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, p := range pointers {
			if p != nil {
				C.free(p)
			}
		}
	}()
	var solution *C.SidereonTrlsSolution
	var drop *C.SidereonTrlsDropOne
	var opErr error
	withCThread(func() {
		if dropOne {
			opErr = statusErrorLocked(uint32(C.sidereon_solve_data_problem_drop_one(&problem, &drop)))
		} else {
			opErr = statusErrorLocked(uint32(C.sidereon_solve_data_problem(&problem, &solution)))
		}
	})
	if opErr != nil {
		return nil, opErr
	}
	if dropOne {
		return newTRLSDropOne(drop)
	}
	return newTRLSSolution(solution)
}

func trlsCopy(handle *surfaceHandle, label string, call func(unsafe.Pointer, *C.double, C.size_t, *C.size_t, *C.size_t) C.enum_SidereonStatus) ([]float64, error) {
	return copySurfaceDoubles(handle, label, call)
}
func (s *TRLSSolution) Close() error {
	if s == nil {
		return nil
	}
	return s.handle.close()
}
func (s *TRLSSolution) Summary() (NativeTRLSSummary, error) {
	var out C.SidereonTrlsSummary
	var opErr error
	err := s.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			opErr = statusErrorLocked(uint32(C.sidereon_trls_solution_summary((*C.SidereonTrlsSolution)(p), &out)))
		})
		return opErr
	})
	if err != nil {
		return NativeTRLSSummary{}, err
	}
	if opErr != nil {
		return NativeTRLSSummary{}, opErr
	}
	return trlsSummary(out), err
}
func (s *TRLSSolution) Values(kind uint32) ([]float64, error) {
	switch kind {
	case 0:
		return trlsCopy(s.handle, "TRLS x", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_trls_solution_x((*C.SidereonTrlsSolution)(p), o, l, w, r)
		})
	case 1:
		return trlsCopy(s.handle, "TRLS residuals", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_trls_solution_residuals((*C.SidereonTrlsSolution)(p), o, l, w, r)
		})
	case 2:
		return trlsCopy(s.handle, "TRLS Jacobian", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_trls_solution_jacobian((*C.SidereonTrlsSolution)(p), o, l, w, r)
		})
	default:
		return trlsCopy(s.handle, "TRLS gradient", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
			return C.sidereon_trls_solution_gradient((*C.SidereonTrlsSolution)(p), o, l, w, r)
		})
	}
}
func (d *TRLSDropOne) Close() error {
	if d == nil {
		return nil
	}
	return d.handle.close()
}
func (d *TRLSDropOne) Count() (int, error) {
	var out C.size_t
	var err error
	err = d.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			err = statusErrorLocked(uint32(C.sidereon_trls_drop_one_count((*C.SidereonTrlsDropOne)(p), &out)))
		})
		return err
	})
	if err != nil {
		return 0, err
	}
	return sizeTToInt(out, "TRLS drop count")
}
func (d *TRLSDropOne) Values() ([]float64, error) {
	return trlsCopy(d.handle, "TRLS cost deltas", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_trls_drop_one_cost_delta((*C.SidereonTrlsDropOne)(p), o, l, w, r)
	})
}
func (d *TRLSDropOne) X(index int) ([]float64, error) {
	if index < 0 {
		return nil, errors.New("sidereon: negative TRLS drop index")
	}
	return trlsCopy(d.handle, "TRLS dropped x", func(p unsafe.Pointer, o *C.double, l C.size_t, w, r *C.size_t) C.enum_SidereonStatus {
		return C.sidereon_trls_drop_one_drop_x((*C.SidereonTrlsDropOne)(p), C.size_t(index), o, l, w, r)
	})
}
func (d *TRLSDropOne) Summary(index int) (NativeTRLSSummary, error) {
	if index < 0 {
		return NativeTRLSSummary{}, errors.New("sidereon: negative TRLS drop index")
	}
	var out C.SidereonTrlsSummary
	var err error
	err = d.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			err = statusErrorLocked(uint32(C.sidereon_trls_drop_one_drop_summary((*C.SidereonTrlsDropOne)(p), C.size_t(index), &out)))
		})
		return err
	})
	return trlsSummary(out), err
}
func (d *TRLSDropOne) BaseSummary() (NativeTRLSSummary, error) {
	var out C.SidereonTrlsSummary
	var err error
	err = d.handle.read(func(p unsafe.Pointer) error {
		withCThread(func() {
			err = statusErrorLocked(uint32(C.sidereon_trls_drop_one_base_summary((*C.SidereonTrlsDropOne)(p), &out)))
		})
		return err
	})
	return trlsSummary(out), err
}
