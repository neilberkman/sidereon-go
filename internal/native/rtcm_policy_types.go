package native

// NativeRTCMDeparture is an owned copy of one encoder departure record.
type NativeRTCMDeparture struct {
	Kind                                        uint32
	MessageNumber                               uint16
	Reserved                                    uint8
	LayerIndex, Declared, Read, Cells, BitCount int
	Degree, Order                               uint8
}

// NativeRTCMStreamDeparture is an owned stream diagnostic.
type NativeRTCMStreamDeparture struct {
	Offset      int
	Description string
}
