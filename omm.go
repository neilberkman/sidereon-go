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

func (o *OMM) Close() error {
	if o == nil || o.handle == nil {
		return nil
	}
	return publicError(o.handle.Close())
}

func (o *OMM) JSON() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.JSON()
	return append([]byte(nil), value...), publicError(err)
}

func (o *OMM) KVN() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.KVN()
	return append([]byte(nil), value...), publicError(err)
}

func (o *OMM) XML() ([]byte, error) {
	if o == nil || o.handle == nil {
		return nil, ErrClosed
	}
	value, err := o.handle.XML()
	return append([]byte(nil), value...), publicError(err)
}
