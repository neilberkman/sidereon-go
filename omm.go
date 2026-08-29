package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

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
		return nil, publicError(err)
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
		return nil, publicError(err)
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
		return nil, publicError(err)
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
	return append([]byte(nil), value...), publicError(err)
}

// KVN returns the serialized KVN representation.
func (o *OMM) KVN() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.KVN()
	return append([]byte(nil), value...), publicError(err)
}

// XML returns the serialized XML representation.
func (o *OMM) XML() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.XML()
	return append([]byte(nil), value...), publicError(err)
}
