//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

const (
	// Values follow #[repr(C)] SidereonSp3Channel in the pinned C binding.
	SP3ChannelPosition = uint32(0)
	SP3ChannelClock    = uint32(1)
)

type SP3CoverageGrid struct {
	HasInterval      bool
	IntervalS        float64
	AgreesWithHeader bool
	OutOfOrder       []int
	Unplaced         []int
}

type SP3ChannelCoverage struct {
	Epochs    int
	SpanCount int
	GapCount  int
	Complete  bool
	Spans     []SP3CoverageSpan
	Gaps      []SP3CoverageGap
}

type SP3CoverageSatellite struct {
	Satellite string
	Declared  bool
	Positions SP3ChannelCoverage
	Clocks    SP3ChannelCoverage
}

type SP3CoverageSpan struct {
	FirstIndex int
	LastIndex  int
	FirstEpoch NativeClockEpoch
	FirstJ2000 float64
	LastEpoch  NativeClockEpoch
	LastJ2000  float64
}

type SP3CoverageGap struct {
	HasAfterIndex  bool
	AfterIndex     int
	HasBeforeIndex bool
	BeforeIndex    int
	MissingEpochs  int
}

type SP3Coverage struct {
	Grid       SP3CoverageGrid
	Satellites []SP3CoverageSatellite
}

func copySp3NativeSlice[T any](label string, copyFn func(unsafe.Pointer, C.size_t, *C.size_t, *C.size_t) uint32) ([]T, error) {
	var written, required C.size_t
	if err := callStatus(func() uint32 { return copyFn(nil, 0, &written, &required) }); err != nil {
		return nil, err
	}
	count, err := checkedNativeCount(uint64(required))
	if err != nil {
		return nil, err
	}
	if _, err := writtenToInt(written, 0, label+" first-call written count"); err != nil {
		return nil, err
	}
	if _, err := checkedNativeAllocationSize(count, unsafe.Sizeof(*new(T))); err != nil {
		return nil, err
	}
	values := make([]T, count)
	var output unsafe.Pointer
	if len(values) != 0 {
		output = unsafe.Pointer(&values[0])
	}
	if err := callStatus(func() uint32 { return copyFn(output, C.size_t(len(values)), &written, &required) }); err != nil {
		return nil, err
	}
	n, err := validateTwoPassCounts(label, len(values), count, uint64(written), uint64(required))
	if err != nil {
		return nil, err
	}
	return values[:n], nil
}

func (s *SP3) Coverage() (SP3Coverage, error) {
	if s == nil || s.handle == nil {
		return SP3Coverage{}, ErrClosed
	}
	var out SP3Coverage
	err := s.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			var raw *C.SidereonSp3Coverage
			if err := callStatus(func() uint32 { return uint32(C.sidereon_sp3_satellite_coverage((*C.SidereonSp3)(pointer), &raw)) }); err != nil {
				return err
			}
			if raw == nil {
				return missingNativeHandle("SP3 coverage")
			}
			defer C.sidereon_sp3_coverage_free(raw)
			var grid C.SidereonSp3EpochGrid
			if err := callStatus(func() uint32 { return uint32(C.sidereon_sp3_coverage_grid(raw, &grid)) }); err != nil {
				return err
			}
			out.Grid = SP3CoverageGrid{HasInterval: bool(grid.has_interval), IntervalS: float64(grid.interval_s), AgreesWithHeader: bool(grid.agrees_with_header)}
			outOfOrderCount, err := sizeTToInt(grid.out_of_order_count, "SP3 out-of-order epoch count")
			if err != nil {
				return err
			}
			unplacedCount, err := sizeTToInt(grid.unplaced_count, "SP3 unplaced epoch count")
			if err != nil {
				return err
			}
			indices := func(fn func(unsafe.Pointer, C.size_t, *C.size_t, *C.size_t) uint32, label string) ([]int, error) {
				rawValues, err := copySp3NativeSlice[C.size_t](label, fn)
				if err != nil {
					return nil, err
				}
				values := make([]int, len(rawValues))
				for i, value := range rawValues {
					values[i], err = sizeTToInt(value, label)
					if err != nil {
						return nil, err
					}
				}
				return values, nil
			}
			out.Grid.OutOfOrder, err = indices(func(dst unsafe.Pointer, n C.size_t, written, required *C.size_t) uint32 {
				return uint32(C.sidereon_sp3_coverage_grid_out_of_order(raw, (*C.size_t)(dst), n, written, required))
			}, "SP3 out-of-order epoch index")
			if err != nil {
				return err
			}
			if len(out.Grid.OutOfOrder) != outOfOrderCount {
				return invalidArgument("SP3 out-of-order epoch count disagrees with copied indices")
			}
			out.Grid.Unplaced, err = indices(func(dst unsafe.Pointer, n C.size_t, written, required *C.size_t) uint32 {
				return uint32(C.sidereon_sp3_coverage_grid_unplaced(raw, (*C.size_t)(dst), n, written, required))
			}, "SP3 unplaced epoch index")
			if err != nil {
				return err
			}
			if len(out.Grid.Unplaced) != unplacedCount {
				return invalidArgument("SP3 unplaced epoch count disagrees with copied indices")
			}
			rawSats, err := copySp3NativeSlice[C.SidereonSp3SatelliteCoverage]("SP3 satellite coverage", func(dst unsafe.Pointer, n C.size_t, written, required *C.size_t) uint32 {
				return uint32(C.sidereon_sp3_coverage_satellites(raw, (*C.SidereonSp3SatelliteCoverage)(dst), n, written, required))
			})
			if err != nil {
				return err
			}
			out.Satellites = make([]SP3CoverageSatellite, len(rawSats))
			for i, satellite := range rawSats {
				out.Satellites[i] = SP3CoverageSatellite{Satellite: tokenFromC(satellite.sat_id), Declared: bool(satellite.declared)}
				for channelIndex, channel := range []uint32{SP3ChannelPosition, SP3ChannelClock} {
					rawChannel := satellite.positions
					if channelIndex == 1 {
						rawChannel = satellite.clocks
					}
					epochs, e := sizeTToInt(rawChannel.epochs, "SP3 channel epoch count")
					if e != nil {
						return e
					}
					spanCount, e := sizeTToInt(rawChannel.span_count, "SP3 channel span count")
					if e != nil {
						return e
					}
					gapCount, e := sizeTToInt(rawChannel.gap_count, "SP3 channel gap count")
					if e != nil {
						return e
					}
					converted := SP3ChannelCoverage{Epochs: epochs, SpanCount: spanCount, GapCount: gapCount, Complete: bool(rawChannel.complete)}
					spans, e := copySp3NativeSlice[C.SidereonSp3CoverageSpan]("SP3 coverage span", func(dst unsafe.Pointer, n C.size_t, written, required *C.size_t) uint32 {
						return uint32(C.sidereon_sp3_coverage_spans(raw, C.size_t(i), C.uint32_t(channel), (*C.SidereonSp3CoverageSpan)(dst), n, written, required))
					})
					if e != nil {
						return e
					}
					if len(spans) != spanCount {
						return invalidArgument("SP3 channel span count disagrees with copied spans")
					}
					converted.Spans = make([]SP3CoverageSpan, len(spans))
					for j, span := range spans {
						first, e := sizeTToInt(span.first_index, "SP3 coverage first index")
						if e != nil {
							return e
						}
						last, e := sizeTToInt(span.last_index, "SP3 coverage last index")
						if e != nil {
							return e
						}
						converted.Spans[j] = SP3CoverageSpan{FirstIndex: first, LastIndex: last, FirstEpoch: clockEpochFromC(span.first_epoch), FirstJ2000: float64(span.first_epoch_j2000_seconds), LastEpoch: clockEpochFromC(span.last_epoch), LastJ2000: float64(span.last_epoch_j2000_seconds)}
					}
					gaps, e := copySp3NativeSlice[C.SidereonSp3CoverageGap]("SP3 coverage gap", func(dst unsafe.Pointer, n C.size_t, written, required *C.size_t) uint32 {
						return uint32(C.sidereon_sp3_coverage_gaps(raw, C.size_t(i), C.uint32_t(channel), (*C.SidereonSp3CoverageGap)(dst), n, written, required))
					})
					if e != nil {
						return e
					}
					if len(gaps) != gapCount {
						return invalidArgument("SP3 channel gap count disagrees with copied gaps")
					}
					converted.Gaps = make([]SP3CoverageGap, len(gaps))
					for j, gap := range gaps {
						missingEpochs, e := sizeTToInt(gap.missing_epochs, "SP3 missing epoch count")
						if e != nil {
							return e
						}
						converted.Gaps[j] = SP3CoverageGap{HasAfterIndex: bool(gap.has_after_index), HasBeforeIndex: bool(gap.has_before_index), MissingEpochs: missingEpochs}
						if gap.has_after_index {
							converted.Gaps[j].AfterIndex, e = sizeTToInt(gap.after_index, "SP3 coverage gap after index")
							if e != nil {
								return e
							}
						}
						if gap.has_before_index {
							converted.Gaps[j].BeforeIndex, e = sizeTToInt(gap.before_index, "SP3 coverage gap before index")
							if e != nil {
								return e
							}
						}
					}
					if channelIndex == 0 {
						out.Satellites[i].Positions = converted
					} else {
						out.Satellites[i].Clocks = converted
					}
				}
			}
			return nil
		})
	})
	runtime.KeepAlive(s)
	if err != nil {
		return SP3Coverage{}, err
	}
	return out, nil
}

// SelectedNodes copies the product nodes selected for at least one query in an inclusive window.
func (s *SP3) SelectedNodes(satellite string, fromJ2000S, throughJ2000S float64) ([]float64, error) {
	if s == nil || s.handle == nil {
		return nil, ErrClosed
	}
	token, err := tokenToC(satellite)
	if err != nil {
		return nil, err
	}
	var result []float64
	err = s.handle.with(func(pointer unsafe.Pointer) error {
		values, err := copySp3NativeSlice[C.double]("SP3 selected nodes", func(dst unsafe.Pointer, n C.size_t, written, required *C.size_t) uint32 {
			return uint32(C.sidereon_sp3_selected_nodes((*C.SidereonSp3)(pointer), &token.bytes[0], C.double(fromJ2000S), C.double(throughJ2000S), (*C.double)(dst), n, written, required))
		})
		if err != nil {
			return err
		}
		result = make([]float64, len(values))
		for i, value := range values {
			result[i] = float64(value)
		}
		return nil
	})
	runtime.KeepAlive(s)
	if err != nil {
		return nil, err
	}
	return result, nil
}
