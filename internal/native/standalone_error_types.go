package native

import "fmt"

func (e *AntexError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("sidereon: ANTEX error %d", e.Kind)
}

// PreciseArtifactError owns the typed and textual detail of an artifact refusal.
type PreciseArtifactError struct {
	Kind                                                               uint32
	Version                                                            uint16
	Tag                                                                uint8
	FoundMagic                                                         [8]byte
	Available, Declared, Offset, Length                                uint64
	ExpectedChecksum, FoundChecksum, ClaimedChecksum, DeclaredChecksum uint64
	HasSatellite                                                       bool
	Satellite                                                          string
	Path, Message, Reason, Region                                      string
}

func (e *PreciseArtifactError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Reason != "" {
		return e.Reason
	}
	return fmt.Sprintf("sidereon: precise artifact error %d", e.Kind)
}

func (e *DtedTileError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Text != "" {
		return e.Text
	}
	return fmt.Sprintf("sidereon: DTED tile error %d", e.Kind)
}

func (e *GeoidError) Error() string {
	if e.Reason != "" {
		return e.Reason
	}
	return fmt.Sprintf("sidereon: geoid error %d", e.Kind)
}
