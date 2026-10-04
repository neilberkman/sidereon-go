package native

// IonexSlantRequest is one whole-UTC-second IONEX query.
type IonexSlantRequest struct {
	LatDeg, LonDeg, AzimuthDeg, ElevationDeg float64
	EpochJ2000S                              int64
	FrequencyHz                              float64
}
