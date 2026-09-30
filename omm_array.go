package sidereon

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sidereon.dev/go/v3/internal/native"
)

// OMMSkippedRecord preserves the parser's zero-based input index and typed cause.
type OMMSkippedRecord struct {
	// Index is the original zero-based position of the failed input record.
	Index uint64
	// Error preserves the complete typed core parse refusal.
	Error OMMParseError
}

// OMMArray owns successfully parsed OMM values and all skipped-entry details.
// Its writers operate on the successful values in their original order.
type OMMArray struct {
	_      noCopy
	handle *native.OMMArray
}

func publicOMMArray(value *native.OMMArray) (*OMMArray, error) {
	if value == nil {
		return nil, errNilNativeHandle
	}
	return &OMMArray{handle: value}, nil
}

// ParseOMMJSONArray parses a JSON object or array and keeps each malformed element.
func ParseOMMJSONArray(data []byte) (*OMMArray, error) {
	value, err := native.ParseOMMJSONArray(append([]byte(nil), data...))
	if err != nil {
		return nil, publicError(err)
	}
	return publicOMMArray(value)
}

// ParseOMMXMLAll parses all XML OMM messages and records each skipped message.
func ParseOMMXMLAll(data []byte) (*OMMArray, error) {
	value, err := native.ParseOMMXMLAll(append([]byte(nil), data...))
	if err != nil {
		return nil, publicError(err)
	}
	return publicOMMArray(value)
}

// ParseOMMCSVArray parses a GP CSV table and retains malformed row indexes.
func ParseOMMCSVArray(data []byte) (*OMMArray, error) {
	value, err := native.ParseOMMCSVArray(append([]byte(nil), data...))
	if err != nil {
		return nil, publicError(err)
	}
	return publicOMMArray(value)
}

// NewOMMArray clones each supplied value into a caller-assembled writer collection.
func NewOMMArray(values ...*OMM) (*OMMArray, error) {
	value, err := native.NewOMMArray()
	if err != nil {
		return nil, publicError(err)
	}
	array, err := publicOMMArray(value)
	if err != nil {
		return nil, publicError(err)
	}
	for i, omm := range values {
		if omm == nil || omm.handle == nil {
			return nil, errors.Join(fmt.Errorf("sidereon: OMM array element %d is closed", i), publicError(array.Close()))
		}
		if err := value.Append(omm.handle); err != nil {
			return nil, errors.Join(publicError(err), publicError(array.Close()))
		}
	}
	return array, nil
}

// Close releases the collection; repeated calls are safe.
func (a *OMMArray) Close() error {
	if a == nil || a.handle == nil {
		return nil
	}
	return publicError(a.handle.Close())
}

// Len returns the number of successfully parsed or appended records.
func (a *OMMArray) Len() (int, error) {
	if a == nil || a.handle == nil {
		return 0, ErrClosed
	}
	n, err := a.handle.Count()
	return n, publicError(err)
}

// Values returns detached complete snapshots in successful-record order.
func (a *OMMArray) Values() ([]OMMData, error) {
	count, err := a.Len()
	if err != nil {
		return nil, err
	}
	values := make([]OMMData, count)
	for i := 0; i < count; i++ {
		record, err := a.handle.Record(i)
		if err != nil {
			return nil, publicError(err)
		}
		detached, snapshotErr := publicOMM(record).Snapshot()
		closeErr := record.Close()
		if snapshotErr != nil || closeErr != nil {
			return nil, errors.Join(publicError(snapshotErr), publicError(closeErr))
		}
		values[i] = detached
	}
	runtime.KeepAlive(a)
	return values, nil
}

// Append clones one OMM, so later source mutation/close cannot change this collection.
func (a *OMMArray) Append(value *OMM) error {
	if a == nil || a.handle == nil {
		return ErrClosed
	}
	if value == nil || value.handle == nil {
		return ErrClosed
	}
	return publicError(a.handle.Append(value.handle))
}

// SkippedRecords returns detached indexes and complete typed parser errors.
func (a *OMMArray) SkippedRecords() ([]OMMSkippedRecord, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	values, err := a.handle.Skipped()
	if err != nil {
		return nil, publicError(err)
	}
	out := make([]OMMSkippedRecord, len(values))
	for i, value := range values {
		var parsed OMMParseError
		if err := json.Unmarshal(value.Payload, &parsed); err != nil {
			return nil, fmt.Errorf("sidereon: invalid native OMM error payload: %w", err)
		}
		if parsed.Kind == "" || len(parsed.Raw) == 0 {
			return nil, fmt.Errorf("sidereon: incomplete native OMM error payload")
		}
		out[i] = OMMSkippedRecord{Index: value.Index, Error: parsed}
	}
	runtime.KeepAlive(a)
	return out, nil
}

// JSON emits all records as one strict GP JSON array.
func (a *OMMArray) JSON() ([]byte, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.JSON()
	return append([]byte(nil), value...), publicError(err)
}

// JSONDiscardingComments emits the array after discarding only comments JSON cannot carry.
func (a *OMMArray) JSONDiscardingComments() ([]byte, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.JSONDiscardingComments()
	return append([]byte(nil), value...), publicError(err)
}

// CSV emits all records as one strict GP CSV table.
func (a *OMMArray) CSV() ([]byte, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.CSV()
	return append([]byte(nil), value...), publicError(err)
}

// CSVDiscardingComments emits the table after discarding only comments CSV cannot carry.
func (a *OMMArray) CSVDiscardingComments() ([]byte, error) {
	if a == nil || a.handle == nil {
		return nil, ErrClosed
	}
	value, err := a.handle.CSVDiscardingComments()
	return append([]byte(nil), value...), publicError(err)
}

// ParseOMM autodetects XML, JSON, GP CSV, or KVN using the core parser.
func ParseOMM(data []byte) (*OMM, error) {
	value, err := native.ParseOMM(append([]byte(nil), data...))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOMM(value), nil
}

// ParseOMMCSV parses one GP CSV data record.
func ParseOMMCSV(data []byte) (*OMM, error) {
	value, err := native.ParseOMMCSV(append([]byte(nil), data...))
	if err != nil {
		return nil, publicError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOMM(value), nil
}

// ParseOMMEpoch parses a standalone OMM epoch with the core UTC-like policy.
func ParseOMMEpoch(text string) (OMMEpoch, error) {
	value, err := native.ParseOMMEpoch(text)
	if err != nil {
		return OMMEpoch{}, publicError(err)
	}
	return OMMEpoch{Year: value.Year, Month: value.Month, Day: value.Day, Hour: value.Hour, Minute: value.Minute, Second: value.Second, Microsecond: value.Microsecond, Femtosecond: value.Femtosecond}, nil
}
