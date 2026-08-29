package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// SiderealTemplateMethod selects the residual template estimator.
type SiderealTemplateMethod uint32

const (
	// SiderealTemplateMean uses the ordinary phase-bin mean.
	SiderealTemplateMean SiderealTemplateMethod = iota
	// SiderealTemplateRobustMAD uses a median/MAD robust template.
	SiderealTemplateRobustMAD
	// SiderealTemplateEWMA uses an exponentially weighted template.
	SiderealTemplateEWMA
)

// SiderealFilterOptions controls phase stacking and template estimation.
type SiderealFilterOptions struct {
	// SampleIntervalS and EWMAAlpha use seconds and a dimensionless gain.
	SampleIntervalS float64
	PriorPeriods    int
	MinCoverage     int
	TemplateMethod  SiderealTemplateMethod
	EWMAAlpha       float64
}

// SiderealFilterOutput owns the C phase-stacked residual result.
type SiderealFilterOutput struct {
	_      noCopy
	handle *native.SiderealFilterOutput
}

func nativeSiderealOptions(value SiderealFilterOptions) native.SiderealFilterOptions {
	return native.SiderealFilterOptions{SampleIntervalS: value.SampleIntervalS, PriorPeriods: value.PriorPeriods, MinCoverage: value.MinCoverage, TemplateMethod: uint32(value.TemplateMethod), EWMAAlpha: value.EWMAAlpha}
}
func publicSiderealOptions(value native.SiderealFilterOptions) SiderealFilterOptions {
	return SiderealFilterOptions{SampleIntervalS: value.SampleIntervalS, PriorPeriods: value.PriorPeriods, MinCoverage: value.MinCoverage, TemplateMethod: SiderealTemplateMethod(value.TemplateMethod), EWMAAlpha: value.EWMAAlpha}
}

// DefaultSiderealFilterOptions returns native sidereal-filter defaults.
func DefaultSiderealFilterOptions() (SiderealFilterOptions, error) {
	value, err := native.SiderealFilterOptionsDefaults()
	return publicSiderealOptions(value), publicError(err)
}

// FilterSidereal phase-stacks a residual series over the requested period.
func FilterSidereal(series []float64, periodSeconds float64, options *SiderealFilterOptions) (*SiderealFilterOutput, error) {
	var nativeOptions *native.SiderealFilterOptions
	if options != nil {
		copy := nativeSiderealOptions(*options)
		nativeOptions = &copy
	}
	value, err := native.SiderealFilter(append([]float64(nil), series...), periodSeconds, nativeOptions)
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &SiderealFilterOutput{handle: value}, nil
}

// SiderealFilter applies a phase-stacked repeat template to a residual
// series. It is the direct operation-shaped spelling of FilterSidereal.
func SiderealFilter(series []float64, periodSeconds float64, options *SiderealFilterOptions) (*SiderealFilterOutput, error) {
	return FilterSidereal(series, periodSeconds, options)
}

// Close releases the native sidereal-filter output and is idempotent.
func (o *SiderealFilterOutput) Close() error {
	if o == nil || o.handle == nil {
		return nil
	}
	return publicError(o.handle.Close())
}

// Filtered returns detached residuals after sidereal filtering.
func (o *SiderealFilterOutput) Filtered() ([]float64, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.Filtered()
	return value, publicError(err)
}

// Template returns the detached phase-bin residual template.
func (o *SiderealFilterOutput) Template() ([]float64, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.Template()
	return value, publicError(err)
}

// Coverage returns detached sample counts for each phase bin.
func (o *SiderealFilterOutput) Coverage() ([]int, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.Coverage()
	return value, publicError(err)
}

// UnderCovered returns detached under-coverage flags for each phase bin.
func (o *SiderealFilterOutput) UnderCovered() ([]bool, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.UnderCovered()
	return value, publicError(err)
}

// SiderealPeriodicityScore contains a candidate period in seconds and its strength.
type SiderealPeriodicityScore struct{ PeriodS, Strength float64 }

// SiderealPeriodicityStrength scores candidate repeat periods in seconds.
func SiderealPeriodicityStrength(series, candidatePeriodsS []float64) ([]SiderealPeriodicityScore, error) {
	values, err := native.SiderealPeriodicityStrength(append([]float64(nil), series...), append([]float64(nil), candidatePeriodsS...))
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]SiderealPeriodicityScore, len(values))
	for i := range out {
		out[i] = SiderealPeriodicityScore{PeriodS: values[i].PeriodS, Strength: values[i].Strength}
	}
	return out, nil
}

// SiderealRepeatPeriod returns the native constellation repeat period in seconds.
func SiderealRepeatPeriod(system uint32) (float64, error) {
	value, err := native.SiderealRepeatPeriod(system)
	return value, publicError(err)
}

// PriorPeriods and MinCoverage are native sample-count controls.
// TemplateMethod selects the estimator.
