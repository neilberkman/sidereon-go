package native

// OMMElementSet is the copied core OMM-to-SGP4 conversion result.
type OMMElementSet struct {
	// EpochWhole is the split Julian-date whole component.
	EpochWhole float64
	// EpochFraction is the split Julian-date fractional component.
	EpochFraction float64
	// BStar is the bridged SGP4 drag term.
	BStar float64
	// MeanMotionDotPresent records whether the input field existed.
	MeanMotionDotPresent bool
	// MeanMotionDot is the first derivative when present.
	MeanMotionDot float64
	// MeanMotionDoubleDotPresent records whether the input field existed.
	MeanMotionDoubleDotPresent bool
	// MeanMotionDoubleDot is the second derivative when present.
	MeanMotionDoubleDot float64
	// Eccentricity is the dimensionless orbital eccentricity.
	Eccentricity float64
	// ArgumentOfPerigeeDeg is the argument of perigee in degrees.
	ArgumentOfPerigeeDeg float64
	// InclinationDeg is the inclination in degrees.
	InclinationDeg float64
	// MeanAnomalyDeg is the mean anomaly in degrees.
	MeanAnomalyDeg float64
	// MeanMotionRevPerDay is mean motion in revolutions per day.
	MeanMotionRevPerDay float64
	// RightAscensionDeg is the ascending-node longitude in degrees.
	RightAscensionDeg float64
	// CatalogNumberPresent records whether the OMM stated a catalog number.
	CatalogNumberPresent bool
	// CatalogNumber is the NORAD catalog identifier when present.
	CatalogNumber uint32
	// OMMEpochDaysPresent records whether the compatibility epoch was produced.
	OMMEpochDaysPresent bool
	// OMMEpochDays is the python-sgp4 epoch in days since 1949-12-31.
	OMMEpochDays float64
}
