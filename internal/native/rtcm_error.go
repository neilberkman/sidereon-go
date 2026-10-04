package native

type RTCMError struct {
	Class   uint32
	Kind    uint32
	Payload []byte
}

func (e *RTCMError) Error() string {
	if e == nil {
		return "sidereon: RTCM error"
	}
	return "sidereon: structured RTCM or SBAS error"
}
