package native

// OMMEpoch is an exact detached civil timestamp returned by parse_epoch.
type OMMEpoch struct {
	// Year is the signed civil year.
	Year int32
	// Month, Day, Hour, Minute, and Second are civil epoch components.
	Month, Day, Hour, Minute, Second uint32
	// Microsecond and Femtosecond preserve all parsed decimal digits.
	Microsecond, Femtosecond uint32
}

// OMMSkippedRecord is one detached skipped OMM input and its typed payload.
type OMMSkippedRecord struct {
	// Index is the zero-based source-record position.
	Index uint64
	// Payload is the complete tagged OMM error JSON.
	Payload []byte
}
