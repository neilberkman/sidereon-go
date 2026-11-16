package sidereon

import (
	"encoding/json"
	"errors"
	"sidereon.dev/go/v3/internal/native"
)

// OMM owns a parsed CCSDS Orbit Mean-Elements Message and must not be copied
// after use.
type OMM struct {
	_      noCopy
	handle *native.OMM
}

func publicOMM(value *native.OMM) *OMM {
	if value == nil {
		return nil
	}
	return &OMM{handle: value}
}

// ParseOMMJSON parses the supplied representation as an OMM JSON document.
func ParseOMMJSON(data []byte) (*OMM, error) {
	value, err := native.ParseOMMJSON(append([]byte(nil), data...))
	if err != nil {
		return nil, publicOMMError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOMM(value), nil
}

// ParseOMMKVN parses the supplied representation as an OMM KVN document.
func ParseOMMKVN(data []byte) (*OMM, error) {
	value, err := native.ParseOMMKVN(append([]byte(nil), data...))
	if err != nil {
		return nil, publicOMMError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOMM(value), nil
}

// ParseOMMXML parses the supplied representation as an OMM XML document.
func ParseOMMXML(data []byte) (*OMM, error) {
	value, err := native.ParseOMMXML(append([]byte(nil), data...))
	if err != nil {
		return nil, publicOMMError(err)
	}
	if value == nil {
		return nil, errNilNativeHandle
	}
	return publicOMM(value), nil
}

// Close releases the native OMM resource and is safe to call repeatedly.
func (o *OMM) Close() error {
	if o == nil || o.handle == nil {
		return nil
	}
	return publicError(o.handle.Close())
}

// JSON returns the serialized JSON representation.
func (o *OMM) JSON() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.JSON()
	return append([]byte(nil), value...), publicOMMError(err)
}

// KVN returns the serialized KVN representation.
func (o *OMM) KVN() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.KVN()
	return append([]byte(nil), value...), publicOMMError(err)
}

// XML returns the serialized XML representation.
func (o *OMM) XML() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.XML()
	return append([]byte(nil), value...), publicOMMError(err)
}

// ToKVNLenient serializes an OMM without requiring a complete CCSDS message.
func (o *OMM) ToKVNLenient() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.KVNLenient()
	return append([]byte(nil), value...), publicOMMError(err)
}

// ToXMLLenient serializes an OMM without requiring a complete CCSDS message.
func (o *OMM) ToXMLLenient() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.XMLLenient()
	return append([]byte(nil), value...), publicOMMError(err)
}

// OMMFromTLE constructs an OMM from a TLE and optional object name.
func OMMFromTLE(tle *TLE, name string) (*OMM, error) {
	if tle == nil || tle.handle == nil {
		return nil, ErrClosed
	}
	value, err := native.OMMFromTLE(tle.handle, name)
	if err != nil {
		return nil, publicOMMError(err)
	}
	return publicOMM(value), nil
}

// WithCCSDSHeader returns a copy with the supplied validated CCSDS header.
func (o *OMM) WithCCSDSHeader(version, creationDate, originator string) (*OMM, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.WithCCSDSHeader(version, creationDate, originator)
	if err != nil {
		return nil, publicOMMError(err)
	}
	return publicOMM(value), nil
}

// Equal compares OMM values using core equality.
func (o *OMM) Equal(other *OMM) (bool, error) {
	if o == nil || o.handle == nil || other == nil || other.handle == nil {
		return false, ErrClosed
	}
	value, err := o.handle.Equal(other.handle)
	return value, publicOMMError(err)
}

func publicOMMError(err error) error {
	result := publicError(err)
	if result == nil {
		return nil
	}
	var engine *EngineError
	if errors.As(result, &engine) && engine.Family == EngineErrorFamilyOMM {
		payload, marshalErr := json.Marshal(struct {
			Kind   string          `json:"kind"`
			Fields json.RawMessage `json:"fields"`
		}{engine.Kind, engine.Fields})
		var detail OMMParseError
		if marshalErr == nil && json.Unmarshal(payload, &detail) == nil {
			return errors.Join(result, &detail)
		}
	}
	return result
}
