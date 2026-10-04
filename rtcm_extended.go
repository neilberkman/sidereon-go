package sidereon

import "sidereon.dev/go/v3/internal/native"

// RTCMTrailingBits retains explicit bits after an RTCM payload's final field.
type RTCMTrailingBits []bool

// RTCMLegacyL1 contains one lossless legacy L1 observation row.
type RTCMLegacyL1 struct {
	CodeIndicator                  bool
	Pseudorange                    uint32
	PhaseRangeMinusPseudorange     int32
	LockTimeIndicator              uint8
	HasPseudorangeModulusAmbiguity bool
	PseudorangeModulusAmbiguity    uint8
	HasCNR                         bool
	CNR                            uint8
}

// RTCMLegacyL2 contains one lossless legacy L2 observation row.
type RTCMLegacyL2 struct {
	CodeIndicator                uint8
	PseudorangeDifference        int16
	PhaseRangeMinusL1Pseudorange int32
	LockTimeIndicator            uint8
	HasCNR                       bool
	CNR                          uint8
}

// RTCMLegacySatellite contains legacy observation data for one satellite.
type RTCMLegacySatellite struct {
	SatelliteID         uint8
	HasFrequencyChannel bool
	FrequencyChannel    uint8
	L1                  RTCMLegacyL1
	HasL2               bool
	L2                  RTCMLegacyL2
}

// RTCMLegacyObservations is a detached legacy RTCM observation message.
type RTCMLegacyObservations struct {
	MessageNumber, ReferenceStationID uint16
	EpochTime                         uint32
	SynchronousGNSS                   bool
	SatelliteCount                    uint8
	DivergenceFreeSmoothing           bool
	SmoothingInterval                 uint8
	Satellites                        []RTCMLegacySatellite
	TrailingBits                      RTCMTrailingBits
}

// RTCMMessageAnnouncement contains one system-parameter announcement.
type RTCMMessageAnnouncement struct {
	MessageNumber uint16
	Synchronous   bool
	Interval      uint16
}

// RTCMSystemParameters contains detached system parameters and announcements.
type RTCMSystemParameters struct {
	ReferenceStationID, MJD        uint16
	SecondsOfDay                   uint32
	AnnouncementCount, LeapSeconds uint8
	Announcements                  []RTCMMessageAnnouncement
	TrailingBits                   RTCMTrailingBits
}

// RTCMText retains a text message's encoded character bytes.
type RTCMText struct {
	ReferenceStationID, MJD uint16
	SecondsOfDay            uint32
	CharacterCount          uint8
	CodeUnits               []byte
	TrailingBits            RTCMTrailingBits
}

// RTCMNetworkAuxiliaryStation contains one auxiliary station record.
type RTCMNetworkAuxiliaryStation struct {
	NetworkID, SubnetworkID, AuxiliaryStationCount uint8
	MasterStationID, AuxiliaryStationID            uint16
	DeltaLatitude, DeltaLongitude, DeltaHeight     int32
	TrailingBits                                   RTCMTrailingBits
}

// RTCMNetworkDifference is one satellite row in a difference message.
type RTCMNetworkDifference struct {
	SatelliteID, AmbiguityStatus, NonSyncCount uint8
	HasGeometric                               bool
	Geometric                                  int32
	HasIOD                                     bool
	IOD                                        uint8
	HasIonospheric                             bool
	Ionospheric                                int32
}

// RTCMNetworkDifferences contains a network difference header and rows.
type RTCMNetworkDifferences struct {
	MessageNumber                       uint16
	NetworkID, SubnetworkID             uint8
	EpochTime                           uint32
	MultipleMessage                     bool
	MasterStationID, AuxiliaryStationID uint16
	SatelliteCount                      uint8
	Satellites                          []RTCMNetworkDifference
	TrailingBits                        RTCMTrailingBits
}

// RTCMNetworkResidual contains one satellite row in a residual message.
type RTCMNetworkResidual struct {
	SatelliteID, SOC uint8
	SOD              uint16
	SOH              uint8
	SLC, SLD         uint16
}

// RTCMNetworkResiduals contains a network residual header and rows.
type RTCMNetworkResiduals struct {
	MessageNumber                         uint16
	EpochTime                             uint32
	ReferenceStationID                    uint16
	ReferenceStationCount, SatelliteCount uint8
	Satellites                            []RTCMNetworkResidual
	TrailingBits                          RTCMTrailingBits
}

// RTCMFKPGradient contains one satellite's FKP gradient coefficients.
type RTCMFKPGradient struct {
	SatelliteID, IOD                                                 uint8
	GeometricNorth, GeometricEast, IonosphericNorth, IonosphericEast int16
}

// RTCMFKPGradients contains an FKP header and satellite rows.
type RTCMFKPGradients struct {
	MessageNumber, ReferenceStationID uint16
	EpochTime                         uint32
	SatelliteCount                    uint8
	Satellites                        []RTCMFKPGradient
	TrailingBits                      RTCMTrailingBits
}

// RTCMGridResidual contains one residual-grid cell.
type RTCMGridResidual struct{ Horizontal1, Horizontal2, Height int16 }

// RTCMResidualGrid contains all sixteen fixed residual-grid cells.
type RTCMResidualGrid struct {
	MessageNumber                                                                      uint16
	SystemID                                                                           uint8
	HorizontalShift, VerticalShift                                                     bool
	Origin1, Origin2                                                                   int32
	Extension1, Extension2                                                             uint16
	MeanOffset1, MeanOffset2, MeanHeightOffset                                         int16
	Residuals                                                                          [16]RTCMGridResidual
	HorizontalInterpolation, VerticalInterpolation, HorizontalQuality, VerticalQuality uint8
	MJD                                                                                uint16
	TrailingBits                                                                       RTCMTrailingBits
}

// RTCMProjection contains raw projection parameters.
type RTCMProjection struct {
	MessageNumber                                             uint16
	SystemID, ProjectionType                                  uint8
	Rectification                                             bool
	Latitude, Longitude, StandardParallel1, StandardParallel2 int64
	Azimuth                                                   uint64
	RectifiedToSkew                                           int32
	AddScale                                                  uint32
	Easting                                                   uint64
	Northing                                                  int64
	TrailingBits                                              RTCMTrailingBits
}

// RTCMHelmertTransformation contains raw Helmert transformation parameters.
type RTCMHelmertTransformation struct {
	MessageNumber                                         uint16
	SourceName, TargetName                                string
	SystemID                                              uint8
	UtilizedMessages                                      uint16
	PlateNumber, ComputationIndicator, HeightIndicator    uint8
	ValidityLatitude, ValidityLongitude                   int32
	ValidityExtensionLatitude, ValidityExtensionLongitude uint16
	DX, DY, DZ, R1, R2, R3, DS                            int32
	HasRotationPoint                                      bool
	RotationPointX, RotationPointY, RotationPointZ        int64
	AddAS, AddBS, AddAT, AddBT                            uint32
	HorizontalQuality, VerticalQuality                    uint8
	TrailingBits                                          RTCMTrailingBits
}

// RTCMPhysicalReferenceStation contains a physical reference-station record.
type RTCMPhysicalReferenceStation struct {
	NonPhysicalStationID, PhysicalStationID uint16
	ITRFRealizationYear                     uint8
	ECEFX, ECEFY, ECEFZ                     int64
	TrailingBits                            RTCMTrailingBits
}

// RTCMGLONASSCodePhaseBiases contains optional GLONASS code and phase biases.
type RTCMGLONASSCodePhaseBiases struct {
	ReferenceStationID               uint16
	Aligned                          bool
	Reserved                         uint8
	HasL1CA, HasL1P, HasL2CA, HasL2P bool
	L1CA, L1P, L2CA, L2P             int16
	TrailingBits                     RTCMTrailingBits
}

// RTCMSSRInfoV2 retains SSR version metadata and the raw padding bits.
type RTCMSSRInfoV2 struct {
	MessageNumber                                                   uint16
	System                                                          GNSSSystem
	Kind                                                            RTCMSSRKind
	Header                                                          RTCMSSRHeader
	HasIGSSSRVersion                                                bool
	IGSSSRVersion                                                   uint8
	OrbitCount, ClockCount, URACount, CodeBiasCount, PhaseBiasCount int
	PaddingBits                                                     RTCMTrailingBits
}

// RTCMSSRVTECInfo contains an SSR VTEC header and polynomial layers.
type RTCMSSRVTECInfo struct {
	MessageNumber    uint16
	HasIGSSSRVersion bool
	IGSSSRVersion    uint8
	EpochTimeS       uint32
	UpdateInterval   uint8
	MultipleMessage  bool
	IODSSR           uint8
	ProviderID       uint16
	SolutionID       uint8
	QualityIndicator uint16
	Layers           []RTCMTecLayer
	TrailingBits     RTCMTrailingBits
}

// RTCMTecLayer contains one VTEC layer's harmonic coefficients.
type RTCMTecLayer struct {
	Height, Degree, Order uint8
	Cosine, Sine          []int16
}

// RTCMUnsupportedBody retains a message number and uninterpreted body bytes.
type RTCMUnsupportedBody struct {
	MessageNumber uint16
	Body          []byte
}

// TrailingBits returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) TrailingBits(index int) (RTCMTrailingBits, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	v, err := m.handle.TrailingBits(index)
	return RTCMTrailingBits(v), publicError(err)
}

// UnsupportedBody returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) UnsupportedBody(index int) (RTCMUnsupportedBody, error) {
	if m == nil || m.handle == nil {
		return RTCMUnsupportedBody{}, ErrClosed
	}
	info, err := m.handle.Kind(index)
	if err != nil {
		return RTCMUnsupportedBody{}, publicError(err)
	}
	body, err := m.handle.UnsupportedBody(index)
	return RTCMUnsupportedBody{MessageNumber: info.MessageNumber, Body: append([]byte(nil), body...)}, publicError(err)
}

// BuildRTCMUnsupported encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMUnsupported(value RTCMUnsupportedBody) (*RTCMMessages, error) {
	h, err := native.BuildRTCMUnsupported(value.MessageNumber, value.Body)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// WithTrailingBits returns a new message collection with the selected message’s trailing bits replaced.
func (m *RTCMMessages) WithTrailingBits(index int, bits RTCMTrailingBits) (*RTCMMessages, error) {
	if m == nil || m.handle == nil {
		return nil, ErrClosed
	}
	h, err := m.handle.WithTrailingBits(index, []bool(bits))
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// SSRInfoV2 returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) SSRInfoV2(index int) (RTCMSSRInfoV2, error) {
	if m == nil || m.handle == nil {
		return RTCMSSRInfoV2{}, ErrClosed
	}
	v, err := m.handle.SSRInfoV2(index)
	if err != nil {
		return RTCMSSRInfoV2{}, publicError(err)
	}
	bits, err := m.TrailingBits(index)
	if err != nil {
		return RTCMSSRInfoV2{}, err
	}
	return RTCMSSRInfoV2{MessageNumber: v.MessageNumber, System: GNSSSystem(v.System), Kind: RTCMSSRKind(v.Kind), Header: RTCMSSRHeader{EpochTimeS: v.Header.EpochTimeS, UpdateInterval: v.Header.UpdateInterval, MultipleMessage: v.Header.MultipleMessage, IODSSR: v.Header.IODSSR, ProviderID: v.Header.ProviderID, SolutionID: v.Header.SolutionID, HasSatelliteReferenceDatum: v.Header.HasSatelliteReferenceDatum, SatelliteReferenceDatum: v.Header.SatelliteReferenceDatum, HasDispersiveBiasConsistency: v.Header.HasDispersiveBiasConsistency, DispersiveBiasConsistency: v.Header.DispersiveBiasConsistency, HasMWConsistency: v.Header.HasMWConsistency, MWConsistency: v.Header.MWConsistency, SatelliteCount: v.Header.SatelliteCount}, HasIGSSSRVersion: v.HasIGSSSRVersion, IGSSSRVersion: v.IGSSSRVersion, OrbitCount: v.OrbitCount, ClockCount: v.ClockCount, URACount: v.URACount, CodeBiasCount: v.CodeBiasCount, PhaseBiasCount: v.PhaseBiasCount, PaddingBits: bits}, nil
}

// LegacyObservations returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) LegacyObservations(index int) (RTCMLegacyObservations, error) {
	if m == nil || m.handle == nil {
		return RTCMLegacyObservations{}, ErrClosed
	}
	value, err := m.handle.LegacyObservations(index)
	if err != nil {
		return RTCMLegacyObservations{}, publicError(err)
	}
	return rtcmLegacyObservations(value), nil
}

func rtcmLegacyObservations(value native.NativeRTCMLegacyObservations) RTCMLegacyObservations {
	result := RTCMLegacyObservations{MessageNumber: value.MessageNumber, ReferenceStationID: value.ReferenceStationID, EpochTime: value.EpochTime, SynchronousGNSS: value.SynchronousGNSS, SatelliteCount: value.SatelliteCount, DivergenceFreeSmoothing: value.DivergenceFreeSmoothing, SmoothingInterval: value.SmoothingInterval, TrailingBits: RTCMTrailingBits(value.TrailingBits), Satellites: make([]RTCMLegacySatellite, len(value.Satellites))}
	for index, row := range value.Satellites {
		result.Satellites[index] = RTCMLegacySatellite{SatelliteID: row.SatelliteID, HasFrequencyChannel: row.HasFrequencyChannel, FrequencyChannel: row.FrequencyChannel, L1: RTCMLegacyL1{CodeIndicator: row.L1.CodeIndicator, Pseudorange: row.L1.Pseudorange, PhaseRangeMinusPseudorange: row.L1.PhaseRangeMinusPseudorange, LockTimeIndicator: row.L1.LockTimeIndicator, HasPseudorangeModulusAmbiguity: row.L1.HasPseudorangeModulusAmbiguity, PseudorangeModulusAmbiguity: row.L1.PseudorangeModulusAmbiguity, HasCNR: row.L1.HasCNR, CNR: row.L1.CNR}, HasL2: row.HasL2, L2: RTCMLegacyL2{CodeIndicator: row.L2.CodeIndicator, PseudorangeDifference: row.L2.PseudorangeDifference, PhaseRangeMinusL1Pseudorange: row.L2.PhaseRangeMinusL1Pseudorange, LockTimeIndicator: row.L2.LockTimeIndicator, HasCNR: row.L2.HasCNR, CNR: row.L2.CNR}}
	}
	return result
}

// BuildRTCMLegacyObservations encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMLegacyObservations(value RTCMLegacyObservations) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMLegacyObservations{MessageNumber: value.MessageNumber, ReferenceStationID: value.ReferenceStationID, EpochTime: value.EpochTime, SynchronousGNSS: value.SynchronousGNSS, SatelliteCount: value.SatelliteCount, DivergenceFreeSmoothing: value.DivergenceFreeSmoothing, SmoothingInterval: value.SmoothingInterval, TrailingBits: []bool(value.TrailingBits), Satellites: make([]native.NativeRTCMLegacySatellite, len(value.Satellites))}
	for index, row := range value.Satellites {
		nativeValue.Satellites[index] = native.NativeRTCMLegacySatellite{SatelliteID: row.SatelliteID, HasFrequencyChannel: row.HasFrequencyChannel, FrequencyChannel: row.FrequencyChannel, L1: native.NativeRTCMLegacyL1{CodeIndicator: row.L1.CodeIndicator, Pseudorange: row.L1.Pseudorange, PhaseRangeMinusPseudorange: row.L1.PhaseRangeMinusPseudorange, LockTimeIndicator: row.L1.LockTimeIndicator, HasPseudorangeModulusAmbiguity: row.L1.HasPseudorangeModulusAmbiguity, PseudorangeModulusAmbiguity: row.L1.PseudorangeModulusAmbiguity, HasCNR: row.L1.HasCNR, CNR: row.L1.CNR}, HasL2: row.HasL2, L2: native.NativeRTCMLegacyL2{CodeIndicator: row.L2.CodeIndicator, PseudorangeDifference: row.L2.PseudorangeDifference, PhaseRangeMinusL1Pseudorange: row.L2.PhaseRangeMinusL1Pseudorange, LockTimeIndicator: row.L2.LockTimeIndicator, HasCNR: row.L2.HasCNR, CNR: row.L2.CNR}}
	}
	h, err := native.BuildRTCMLegacyObservations(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: h}, nil
}

// SystemParameters returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) SystemParameters(index int) (RTCMSystemParameters, error) {
	if m == nil || m.handle == nil {
		return RTCMSystemParameters{}, ErrClosed
	}
	value, err := m.handle.SystemParameters(index)
	if err != nil {
		return RTCMSystemParameters{}, publicError(err)
	}
	result := RTCMSystemParameters{ReferenceStationID: value.ReferenceStationID, MJD: value.MJD, SecondsOfDay: value.SecondsOfDay, AnnouncementCount: value.AnnouncementCount, LeapSeconds: value.LeapSeconds, TrailingBits: RTCMTrailingBits(value.TrailingBits), Announcements: make([]RTCMMessageAnnouncement, len(value.Announcements))}
	for row, item := range value.Announcements {
		result.Announcements[row] = RTCMMessageAnnouncement{MessageNumber: item.MessageNumber, Synchronous: item.Synchronous, Interval: item.Interval}
	}
	return result, nil
}

// Text returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) Text(index int) (RTCMText, error) {
	if m == nil || m.handle == nil {
		return RTCMText{}, ErrClosed
	}
	value, err := m.handle.Text(index)
	if err != nil {
		return RTCMText{}, publicError(err)
	}
	return RTCMText{ReferenceStationID: value.ReferenceStationID, MJD: value.MJD, SecondsOfDay: value.SecondsOfDay, CharacterCount: value.CharacterCount, CodeUnits: append([]byte(nil), value.CodeUnits...), TrailingBits: RTCMTrailingBits(value.TrailingBits)}, nil
}

// BuildRTCMSystemParameters encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMSystemParameters(value RTCMSystemParameters) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMSystemParameters{ReferenceStationID: value.ReferenceStationID, MJD: value.MJD, SecondsOfDay: value.SecondsOfDay, AnnouncementCount: value.AnnouncementCount, LeapSeconds: value.LeapSeconds, TrailingBits: []bool(value.TrailingBits), Announcements: make([]native.NativeRTCMMessageAnnouncement, len(value.Announcements))}
	for index, row := range value.Announcements {
		nativeValue.Announcements[index] = native.NativeRTCMMessageAnnouncement{MessageNumber: row.MessageNumber, Synchronous: row.Synchronous, Interval: row.Interval}
	}
	handle, err := native.BuildRTCMSystemParameters(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// BuildRTCMText encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMText(value RTCMText) (*RTCMMessages, error) {
	handle, err := native.BuildRTCMText(native.NativeRTCMText{ReferenceStationID: value.ReferenceStationID, MJD: value.MJD, SecondsOfDay: value.SecondsOfDay, CharacterCount: value.CharacterCount, CodeUnits: append([]byte(nil), value.CodeUnits...), TrailingBits: []bool(value.TrailingBits)})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// NetworkAuxiliaryStation returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) NetworkAuxiliaryStation(index int) (RTCMNetworkAuxiliaryStation, error) {
	if m == nil || m.handle == nil {
		return RTCMNetworkAuxiliaryStation{}, ErrClosed
	}
	value, err := m.handle.NetworkAuxiliaryStation(index)
	if err != nil {
		return RTCMNetworkAuxiliaryStation{}, publicError(err)
	}
	return RTCMNetworkAuxiliaryStation{NetworkID: value.NetworkID, SubnetworkID: value.SubnetworkID, AuxiliaryStationCount: value.AuxiliaryStationCount, MasterStationID: value.MasterStationID, AuxiliaryStationID: value.AuxiliaryStationID, DeltaLatitude: value.DeltaLatitude, DeltaLongitude: value.DeltaLongitude, DeltaHeight: value.DeltaHeight, TrailingBits: RTCMTrailingBits(value.TrailingBits)}, nil
}

// BuildRTCMNetworkAuxiliaryStation encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMNetworkAuxiliaryStation(value RTCMNetworkAuxiliaryStation) (*RTCMMessages, error) {
	handle, err := native.BuildRTCMNetworkAuxiliaryStation(native.NativeRTCMNetworkAuxiliaryStation{NetworkID: value.NetworkID, SubnetworkID: value.SubnetworkID, AuxiliaryStationCount: value.AuxiliaryStationCount, MasterStationID: value.MasterStationID, AuxiliaryStationID: value.AuxiliaryStationID, DeltaLatitude: value.DeltaLatitude, DeltaLongitude: value.DeltaLongitude, DeltaHeight: value.DeltaHeight, TrailingBits: []bool(value.TrailingBits)})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// SSRVTEC returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) SSRVTEC(index int) (RTCMSSRVTECInfo, error) {
	if m == nil || m.handle == nil {
		return RTCMSSRVTECInfo{}, ErrClosed
	}
	value, err := m.handle.SSRVTEC(index)
	if err != nil {
		return RTCMSSRVTECInfo{}, publicError(err)
	}
	result := RTCMSSRVTECInfo{MessageNumber: value.MessageNumber, HasIGSSSRVersion: value.HasIGSSSRVersion, IGSSSRVersion: value.IGSSSRVersion, EpochTimeS: value.EpochTimeS, UpdateInterval: value.UpdateInterval, MultipleMessage: value.MultipleMessage, IODSSR: value.IODSSR, ProviderID: value.ProviderID, SolutionID: value.SolutionID, QualityIndicator: value.QualityIndicator, TrailingBits: RTCMTrailingBits(value.TrailingBits), Layers: make([]RTCMTecLayer, len(value.Layers))}
	for index, layer := range value.Layers {
		result.Layers[index] = RTCMTecLayer{Height: layer.Height, Degree: layer.Degree, Order: layer.Order, Cosine: append([]int16(nil), layer.Cosine...), Sine: append([]int16(nil), layer.Sine...)}
	}
	return result, nil
}

// BuildRTCMSSRVTEC encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMSSRVTEC(value RTCMSSRVTECInfo) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMSSRVTECInfo{MessageNumber: value.MessageNumber, HasIGSSSRVersion: value.HasIGSSSRVersion, IGSSSRVersion: value.IGSSSRVersion, EpochTimeS: value.EpochTimeS, UpdateInterval: value.UpdateInterval, MultipleMessage: value.MultipleMessage, IODSSR: value.IODSSR, ProviderID: value.ProviderID, SolutionID: value.SolutionID, QualityIndicator: value.QualityIndicator, TrailingBits: []bool(value.TrailingBits), Layers: make([]native.NativeRTCMTecLayer, len(value.Layers))}
	for index, layer := range value.Layers {
		nativeValue.Layers[index] = native.NativeRTCMTecLayer{Height: layer.Height, Degree: layer.Degree, Order: layer.Order, Cosine: append([]int16(nil), layer.Cosine...), Sine: append([]int16(nil), layer.Sine...)}
	}
	handle, err := native.BuildRTCMSSRVTEC(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// NetworkDifferences returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) NetworkDifferences(index int) (RTCMNetworkDifferences, error) {
	if m == nil || m.handle == nil {
		return RTCMNetworkDifferences{}, ErrClosed
	}
	value, err := m.handle.NetworkDifferences(index)
	if err != nil {
		return RTCMNetworkDifferences{}, publicError(err)
	}
	result := RTCMNetworkDifferences{MessageNumber: value.MessageNumber, NetworkID: value.NetworkID, SubnetworkID: value.SubnetworkID, EpochTime: value.EpochTime, MultipleMessage: value.MultipleMessage, MasterStationID: value.MasterStationID, AuxiliaryStationID: value.AuxiliaryStationID, SatelliteCount: value.SatelliteCount, TrailingBits: RTCMTrailingBits(value.TrailingBits), Satellites: make([]RTCMNetworkDifference, len(value.Satellites))}
	for row, item := range value.Satellites {
		result.Satellites[row] = RTCMNetworkDifference{SatelliteID: item.SatelliteID, AmbiguityStatus: item.AmbiguityStatus, NonSyncCount: item.NonSyncCount, HasGeometric: item.HasGeometric, Geometric: item.Geometric, HasIOD: item.HasIOD, IOD: item.IOD, HasIonospheric: item.HasIonospheric, Ionospheric: item.Ionospheric}
	}
	return result, nil
}

// BuildRTCMNetworkDifferences encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMNetworkDifferences(value RTCMNetworkDifferences) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMNetworkDifferences{MessageNumber: value.MessageNumber, NetworkID: value.NetworkID, SubnetworkID: value.SubnetworkID, EpochTime: value.EpochTime, MultipleMessage: value.MultipleMessage, MasterStationID: value.MasterStationID, AuxiliaryStationID: value.AuxiliaryStationID, SatelliteCount: value.SatelliteCount, TrailingBits: []bool(value.TrailingBits), Satellites: make([]native.NativeRTCMNetworkDifference, len(value.Satellites))}
	for row, item := range value.Satellites {
		nativeValue.Satellites[row] = native.NativeRTCMNetworkDifference{SatelliteID: item.SatelliteID, AmbiguityStatus: item.AmbiguityStatus, NonSyncCount: item.NonSyncCount, HasGeometric: item.HasGeometric, Geometric: item.Geometric, HasIOD: item.HasIOD, IOD: item.IOD, HasIonospheric: item.HasIonospheric, Ionospheric: item.Ionospheric}
	}
	handle, err := native.BuildRTCMNetworkDifferences(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// NetworkResiduals returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) NetworkResiduals(index int) (RTCMNetworkResiduals, error) {
	if m == nil || m.handle == nil {
		return RTCMNetworkResiduals{}, ErrClosed
	}
	value, err := m.handle.NetworkResiduals(index)
	if err != nil {
		return RTCMNetworkResiduals{}, publicError(err)
	}
	result := RTCMNetworkResiduals{MessageNumber: value.MessageNumber, EpochTime: value.EpochTime, ReferenceStationID: value.ReferenceStationID, ReferenceStationCount: value.ReferenceStationCount, SatelliteCount: value.SatelliteCount, TrailingBits: RTCMTrailingBits(value.TrailingBits), Satellites: make([]RTCMNetworkResidual, len(value.Satellites))}
	for row, item := range value.Satellites {
		result.Satellites[row] = RTCMNetworkResidual{SatelliteID: item.SatelliteID, SOC: item.SOC, SOD: item.SOD, SOH: item.SOH, SLC: item.SLC, SLD: item.SLD}
	}
	return result, nil
}

// BuildRTCMNetworkResiduals encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMNetworkResiduals(value RTCMNetworkResiduals) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMNetworkResiduals{MessageNumber: value.MessageNumber, EpochTime: value.EpochTime, ReferenceStationID: value.ReferenceStationID, ReferenceStationCount: value.ReferenceStationCount, SatelliteCount: value.SatelliteCount, TrailingBits: []bool(value.TrailingBits), Satellites: make([]native.NativeRTCMNetworkResidual, len(value.Satellites))}
	for row, item := range value.Satellites {
		nativeValue.Satellites[row] = native.NativeRTCMNetworkResidual{SatelliteID: item.SatelliteID, SOC: item.SOC, SOD: item.SOD, SOH: item.SOH, SLC: item.SLC, SLD: item.SLD}
	}
	handle, err := native.BuildRTCMNetworkResiduals(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// FKPGradients returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) FKPGradients(index int) (RTCMFKPGradients, error) {
	if m == nil || m.handle == nil {
		return RTCMFKPGradients{}, ErrClosed
	}
	value, err := m.handle.FKPGradients(index)
	if err != nil {
		return RTCMFKPGradients{}, publicError(err)
	}
	result := RTCMFKPGradients{MessageNumber: value.MessageNumber, ReferenceStationID: value.ReferenceStationID, EpochTime: value.EpochTime, SatelliteCount: value.SatelliteCount, TrailingBits: RTCMTrailingBits(value.TrailingBits), Satellites: make([]RTCMFKPGradient, len(value.Satellites))}
	for row, item := range value.Satellites {
		result.Satellites[row] = RTCMFKPGradient{SatelliteID: item.SatelliteID, IOD: item.IOD, GeometricNorth: item.GeometricNorth, GeometricEast: item.GeometricEast, IonosphericNorth: item.IonosphericNorth, IonosphericEast: item.IonosphericEast}
	}
	return result, nil
}

// BuildRTCMFKPGradients encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMFKPGradients(value RTCMFKPGradients) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMFKPGradients{MessageNumber: value.MessageNumber, ReferenceStationID: value.ReferenceStationID, EpochTime: value.EpochTime, SatelliteCount: value.SatelliteCount, TrailingBits: []bool(value.TrailingBits), Satellites: make([]native.NativeRTCMFKPGradient, len(value.Satellites))}
	for row, item := range value.Satellites {
		nativeValue.Satellites[row] = native.NativeRTCMFKPGradient{SatelliteID: item.SatelliteID, IOD: item.IOD, GeometricNorth: item.GeometricNorth, GeometricEast: item.GeometricEast, IonosphericNorth: item.IonosphericNorth, IonosphericEast: item.IonosphericEast}
	}
	handle, err := native.BuildRTCMFKPGradients(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// ResidualGrid returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) ResidualGrid(index int) (RTCMResidualGrid, error) {
	if m == nil || m.handle == nil {
		return RTCMResidualGrid{}, ErrClosed
	}
	value, err := m.handle.ResidualGrid(index)
	if err != nil {
		return RTCMResidualGrid{}, publicError(err)
	}
	result := RTCMResidualGrid{MessageNumber: value.MessageNumber, SystemID: value.SystemID, HorizontalShift: value.HorizontalShift, VerticalShift: value.VerticalShift, Origin1: value.Origin1, Origin2: value.Origin2, Extension1: value.Extension1, Extension2: value.Extension2, MeanOffset1: value.MeanOffset1, MeanOffset2: value.MeanOffset2, MeanHeightOffset: value.MeanHeightOffset, HorizontalInterpolation: value.HorizontalInterpolation, VerticalInterpolation: value.VerticalInterpolation, HorizontalQuality: value.HorizontalQuality, VerticalQuality: value.VerticalQuality, MJD: value.MJD, TrailingBits: RTCMTrailingBits(value.TrailingBits)}
	for row, item := range value.Residuals {
		result.Residuals[row] = RTCMGridResidual{Horizontal1: item.Horizontal1, Horizontal2: item.Horizontal2, Height: item.Height}
	}
	return result, nil
}

// BuildRTCMResidualGrid encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMResidualGrid(value RTCMResidualGrid) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMResidualGrid{MessageNumber: value.MessageNumber, SystemID: value.SystemID, HorizontalShift: value.HorizontalShift, VerticalShift: value.VerticalShift, Origin1: value.Origin1, Origin2: value.Origin2, Extension1: value.Extension1, Extension2: value.Extension2, MeanOffset1: value.MeanOffset1, MeanOffset2: value.MeanOffset2, MeanHeightOffset: value.MeanHeightOffset, HorizontalInterpolation: value.HorizontalInterpolation, VerticalInterpolation: value.VerticalInterpolation, HorizontalQuality: value.HorizontalQuality, VerticalQuality: value.VerticalQuality, MJD: value.MJD, TrailingBits: []bool(value.TrailingBits)}
	for row, item := range value.Residuals {
		nativeValue.Residuals[row] = native.NativeRTCMGridResidual{Horizontal1: item.Horizontal1, Horizontal2: item.Horizontal2, Height: item.Height}
	}
	handle, err := native.BuildRTCMResidualGrid(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// Projection returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) Projection(index int) (RTCMProjection, error) {
	if m == nil || m.handle == nil {
		return RTCMProjection{}, ErrClosed
	}
	value, err := m.handle.Projection(index)
	if err != nil {
		return RTCMProjection{}, publicError(err)
	}
	return RTCMProjection{MessageNumber: value.MessageNumber, SystemID: value.SystemID, ProjectionType: value.ProjectionType, Rectification: value.Rectification, Latitude: value.Latitude, Longitude: value.Longitude, StandardParallel1: value.StandardParallel1, StandardParallel2: value.StandardParallel2, Azimuth: value.Azimuth, RectifiedToSkew: value.RectifiedToSkew, AddScale: value.AddScale, Easting: value.Easting, Northing: value.Northing, TrailingBits: RTCMTrailingBits(value.TrailingBits)}, nil
}

// BuildRTCMProjection encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMProjection(value RTCMProjection) (*RTCMMessages, error) {
	handle, err := native.BuildRTCMProjection(native.NativeRTCMProjection{MessageNumber: value.MessageNumber, SystemID: value.SystemID, ProjectionType: value.ProjectionType, Rectification: value.Rectification, Latitude: value.Latitude, Longitude: value.Longitude, StandardParallel1: value.StandardParallel1, StandardParallel2: value.StandardParallel2, Azimuth: value.Azimuth, RectifiedToSkew: value.RectifiedToSkew, AddScale: value.AddScale, Easting: value.Easting, Northing: value.Northing, TrailingBits: []bool(value.TrailingBits)})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// PhysicalReferenceStation returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) PhysicalReferenceStation(index int) (RTCMPhysicalReferenceStation, error) {
	if m == nil || m.handle == nil {
		return RTCMPhysicalReferenceStation{}, ErrClosed
	}
	value, err := m.handle.PhysicalReferenceStation(index)
	if err != nil {
		return RTCMPhysicalReferenceStation{}, publicError(err)
	}
	return RTCMPhysicalReferenceStation{NonPhysicalStationID: value.NonPhysicalStationID, PhysicalStationID: value.PhysicalStationID, ITRFRealizationYear: value.ITRFRealizationYear, ECEFX: value.ECEFX, ECEFY: value.ECEFY, ECEFZ: value.ECEFZ, TrailingBits: RTCMTrailingBits(value.TrailingBits)}, nil
}

// BuildRTCMPhysicalReferenceStation encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMPhysicalReferenceStation(value RTCMPhysicalReferenceStation) (*RTCMMessages, error) {
	handle, err := native.BuildRTCMPhysicalReferenceStation(native.NativeRTCMPhysicalReferenceStation{NonPhysicalStationID: value.NonPhysicalStationID, PhysicalStationID: value.PhysicalStationID, ITRFRealizationYear: value.ITRFRealizationYear, ECEFX: value.ECEFX, ECEFY: value.ECEFY, ECEFZ: value.ECEFZ, TrailingBits: []bool(value.TrailingBits)})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// HelmertTransformation returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) HelmertTransformation(index int) (RTCMHelmertTransformation, error) {
	if m == nil || m.handle == nil {
		return RTCMHelmertTransformation{}, ErrClosed
	}
	value, err := m.handle.HelmertTransformation(index)
	if err != nil {
		return RTCMHelmertTransformation{}, publicError(err)
	}
	return RTCMHelmertTransformation{MessageNumber: value.MessageNumber, SourceName: value.SourceName, TargetName: value.TargetName, SystemID: value.SystemID, UtilizedMessages: value.UtilizedMessages, PlateNumber: value.PlateNumber, ComputationIndicator: value.ComputationIndicator, HeightIndicator: value.HeightIndicator, ValidityLatitude: value.ValidityLatitude, ValidityLongitude: value.ValidityLongitude, ValidityExtensionLatitude: value.ValidityExtensionLatitude, ValidityExtensionLongitude: value.ValidityExtensionLongitude, DX: value.DX, DY: value.DY, DZ: value.DZ, R1: value.R1, R2: value.R2, R3: value.R3, DS: value.DS, HasRotationPoint: value.HasRotationPoint, RotationPointX: value.RotationPointX, RotationPointY: value.RotationPointY, RotationPointZ: value.RotationPointZ, AddAS: value.AddAS, AddBS: value.AddBS, AddAT: value.AddAT, AddBT: value.AddBT, HorizontalQuality: value.HorizontalQuality, VerticalQuality: value.VerticalQuality, TrailingBits: RTCMTrailingBits(value.TrailingBits)}, nil
}

// BuildRTCMHelmertTransformation encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMHelmertTransformation(value RTCMHelmertTransformation) (*RTCMMessages, error) {
	handle, err := native.BuildRTCMHelmertTransformation(native.NativeRTCMHelmertTransformation{MessageNumber: value.MessageNumber, SourceName: value.SourceName, TargetName: value.TargetName, SystemID: value.SystemID, UtilizedMessages: value.UtilizedMessages, PlateNumber: value.PlateNumber, ComputationIndicator: value.ComputationIndicator, HeightIndicator: value.HeightIndicator, ValidityLatitude: value.ValidityLatitude, ValidityLongitude: value.ValidityLongitude, ValidityExtensionLatitude: value.ValidityExtensionLatitude, ValidityExtensionLongitude: value.ValidityExtensionLongitude, DX: value.DX, DY: value.DY, DZ: value.DZ, R1: value.R1, R2: value.R2, R3: value.R3, DS: value.DS, HasRotationPoint: value.HasRotationPoint, RotationPointX: value.RotationPointX, RotationPointY: value.RotationPointY, RotationPointZ: value.RotationPointZ, AddAS: value.AddAS, AddBS: value.AddBS, AddAT: value.AddAT, AddBT: value.AddBT, HorizontalQuality: value.HorizontalQuality, VerticalQuality: value.VerticalQuality, TrailingBits: []bool(value.TrailingBits)})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// GLONASSCodePhaseBiases returns the decoded typed payload for this RTCM message; it reports an error when the message type differs.
func (m *RTCMMessages) GLONASSCodePhaseBiases(index int) (RTCMGLONASSCodePhaseBiases, error) {
	if m == nil || m.handle == nil {
		return RTCMGLONASSCodePhaseBiases{}, ErrClosed
	}
	value, err := m.handle.GLONASSCodePhaseBiases(index)
	if err != nil {
		return RTCMGLONASSCodePhaseBiases{}, publicError(err)
	}
	return RTCMGLONASSCodePhaseBiases{ReferenceStationID: value.ReferenceStationID, Aligned: value.Aligned, Reserved: value.Reserved, HasL1CA: value.HasL1CA, L1CA: value.L1CA, HasL1P: value.HasL1P, L1P: value.L1P, HasL2CA: value.HasL2CA, L2CA: value.L2CA, HasL2P: value.HasL2P, L2P: value.L2P, TrailingBits: RTCMTrailingBits(value.TrailingBits)}, nil
}

// BuildRTCMGLONASSCodePhaseBiases encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMGLONASSCodePhaseBiases(value RTCMGLONASSCodePhaseBiases) (*RTCMMessages, error) {
	handle, err := native.BuildRTCMGLONASSCodePhaseBiases(native.NativeRTCMGLONASSCodePhaseBiases{ReferenceStationID: value.ReferenceStationID, Aligned: value.Aligned, Reserved: value.Reserved, HasL1CA: value.HasL1CA, L1CA: value.L1CA, HasL1P: value.HasL1P, L1P: value.L1P, HasL2CA: value.HasL2CA, L2CA: value.L2CA, HasL2P: value.HasL2P, L2P: value.L2P, TrailingBits: []bool(value.TrailingBits)})
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}

// BuildRTCMSSRMessageV2 encodes the supplied typed payload as an RTCM message and validates its field ranges.
func BuildRTCMSSRMessageV2(info RTCMSSRInfoV2, orbits []RTCMSSROrbitRecord, clocks []RTCMSSRClockRecord, ura []RTCMSSRURARecord, codeBiases []RTCMSSRCodeBiasGroup, phaseBiases []RTCMSSRPhaseBiasGroup) (*RTCMMessages, error) {
	nativeValue := native.NativeRTCMSSRMessageV2{Info: native.NativeRTCMSSRInfoV2{MessageNumber: info.MessageNumber, System: uint32(info.System), Kind: uint32(info.Kind), HasIGSSSRVersion: info.HasIGSSSRVersion, IGSSSRVersion: info.IGSSSRVersion, Header: native.NativeSsrHeader{EpochTimeS: info.Header.EpochTimeS, UpdateInterval: info.Header.UpdateInterval, MultipleMessage: info.Header.MultipleMessage, IODSSR: info.Header.IODSSR, ProviderID: info.Header.ProviderID, SolutionID: info.Header.SolutionID, HasSatelliteReferenceDatum: info.Header.HasSatelliteReferenceDatum, SatelliteReferenceDatum: info.Header.SatelliteReferenceDatum, HasDispersiveBiasConsistency: info.Header.HasDispersiveBiasConsistency, DispersiveBiasConsistency: info.Header.DispersiveBiasConsistency, HasMWConsistency: info.Header.HasMWConsistency, MWConsistency: info.Header.MWConsistency, SatelliteCount: info.Header.SatelliteCount}}, PaddingBits: []bool(info.PaddingBits)}
	for _, row := range orbits {
		nativeValue.Orbits = append(nativeValue.Orbits, native.NativeSsrOrbitRecord{SatelliteID: row.SatelliteID, IODE: row.IODE, DeltaRadial: row.DeltaRadial, DeltaAlong: row.DeltaAlong, DeltaCross: row.DeltaCross, DotDeltaRadial: row.DotDeltaRadial, DotDeltaAlong: row.DotDeltaAlong, DotDeltaCross: row.DotDeltaCross})
	}
	for _, row := range clocks {
		nativeValue.Clocks = append(nativeValue.Clocks, native.NativeSsrClockRecord{SatelliteID: row.SatelliteID, C0: row.C0, C1: row.C1, C2: row.C2})
	}
	for _, row := range ura {
		nativeValue.URA = append(nativeValue.URA, native.NativeSsrUraRecord{SatelliteID: row.SatelliteID, URAIndex: row.URAIndex})
	}
	for _, group := range codeBiases {
		nativeValue.CodeBiases = append(nativeValue.CodeBiases, native.NativeSsrCodeBiasRecord{SatelliteID: group.Record.SatelliteID, SignalCount: len(group.Signals)})
		for _, signal := range group.Signals {
			nativeValue.CodeBiasSignals = append(nativeValue.CodeBiasSignals, native.NativeSsrCodeBiasSignal{SignalID: signal.SignalID, Bias: signal.Bias})
		}
	}
	for _, group := range phaseBiases {
		nativeValue.PhaseBiases = append(nativeValue.PhaseBiases, native.NativeSsrPhaseBiasRecord{SatelliteID: group.Record.SatelliteID, YawAngle: group.Record.YawAngle, YawRate: group.Record.YawRate, SignalCount: len(group.Signals)})
		for _, signal := range group.Signals {
			nativeValue.PhaseBiasSignals = append(nativeValue.PhaseBiasSignals, native.NativeSsrPhaseBiasSignal{SignalID: signal.SignalID, IntegerIndicator: signal.IntegerIndicator, WideLaneIntegerIndicator: signal.WideLaneIntegerIndicator, DiscontinuityCounter: signal.DiscontinuityCounter, Bias: signal.Bias})
		}
	}
	handle, err := native.BuildRTCMSSRMessageV2(nativeValue)
	if err != nil {
		return nil, publicError(err)
	}
	return &RTCMMessages{handle: handle}, nil
}
