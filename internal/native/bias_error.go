package native

type BiasError struct {
	Kind          uint32
	Line          int
	Record        int
	HasTimeScale  bool
	TimeScale     uint32
	Departure     BiasNotice
	Message       []byte
	Field         []byte
	Reason        []byte
	Code          []byte
	Version       []byte
	DepartureText [9][]byte
}

func (e *BiasError) Error() string {
	if e == nil {
		return "sidereon: bias error"
	}
	if len(e.Message) != 0 {
		return string(e.Message)
	}
	return "sidereon: bias error"
}
