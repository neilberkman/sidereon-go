package sidereon

import "sidereon.dev/go/v3/internal/native"

// SolveBroadcastV2AtExactEpoch solves broadcast SPP with exact receive time and model selectors.
func SolveBroadcastV2AtExactEpoch(broadcast *BroadcastEphemeris, input SPPInputsV2, receiveEpoch *ExactEpoch) (*SPPSolutionHandle, error) {
	if broadcast == nil || broadcast.handle == nil || receiveEpoch == nil || receiveEpoch.handle == nil {
		return nil, ErrClosed
	}
	nativeInput, err := nativeSppV2(input)
	if err != nil {
		return nil, publicError(err)
	}
	handle, err := native.SolveBroadcastV2AtExactEpoch(broadcast.handle, nativeInput, receiveEpoch.handle)
	if err != nil {
		return nil, publicError(err)
	}
	return &SPPSolutionHandle{handle: handle}, nil
}

func (store *SBASCorrectionStore) SolveBroadcastV2AtExactEpoch(broadcast *BroadcastEphemeris, geo string, mode SBASSolveMode, input SPPInputsV2, receiveEpoch *ExactEpoch) (SPPSolution, error) {
	if store == nil || store.handle == nil || broadcast == nil || broadcast.handle == nil || receiveEpoch == nil || receiveEpoch.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return SPPSolution{}, err
	}
	nativeInput, err := nativeSppV2(input)
	if err != nil {
		return SPPSolution{}, publicError(err)
	}
	value, err := store.handle.SolveBroadcastV2AtExactEpoch(broadcast.handle, geo, uint32(mode), nativeInput, receiveEpoch.handle)
	return publicSPPSolution(value), publicError(err)
}

func (store *SSRCorrectionStore) SolveBroadcastV2AtExactEpoch(broadcast *BroadcastEphemeris, input SPPInputsV2, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy, receiveEpoch *ExactEpoch) (SPPSolution, error) {
	if store == nil || store.handle == nil || broadcast == nil || broadcast.handle == nil || receiveEpoch == nil || receiveEpoch.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return SPPSolution{}, err
	}
	if sizePolicy != SSRCorrectionSizeStrict && sizePolicy != SSRCorrectionSizeLenient {
		return SPPSolution{}, invalidArgument("invalid SSR correction size policy")
	}
	nativeInput, err := nativeSppV2(input)
	if err != nil {
		return SPPSolution{}, publicError(err)
	}
	value, err := store.handle.SolveBroadcastV2AtExactEpoch(broadcast.handle, nativeInput, stalenessSeconds, uint32(missing), allowRegionalProvider, regionalProviderID, uint32(sizePolicy), receiveEpoch.handle)
	return publicSPPSolution(value), publicError(err)
}

// SolveBroadcastAtExactEpoch solves with the supplied exact receive epoch.
// ReceiveEpoch is authoritative; SPPConfig's legacy scalar epoch fields remain
// available for compatibility with SolveBroadcast.
func (s *SBASCorrectionStore) SolveBroadcastAtExactEpoch(broadcast *BroadcastEphemeris, geo string, mode SBASSolveMode, config SPPConfig, receiveEpoch *ExactEpoch) (SPPSolution, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || receiveEpoch == nil || receiveEpoch.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSBASSolveMode(mode); err != nil {
		return SPPSolution{}, err
	}
	if config.Models != (SPPModelOptions{}) {
		input, err := nativeSppV2(SPPInputsV2{Base: config, Models: config.Models})
		if err != nil {
			return SPPSolution{}, publicError(err)
		}
		value, err := s.handle.SolveBroadcastV2AtExactEpoch(broadcast.handle, geo, uint32(mode), input, receiveEpoch.handle)
		return publicSPPSolution(value), publicError(err)
	}
	value, err := s.handle.SolveBroadcastAtExactEpoch(broadcast.handle, geo, uint32(mode), nativeSPPConfig(config), receiveEpoch.handle)
	return publicSPPSolution(value), publicError(err)
}

// SolveBroadcastAtExactEpoch solves with the supplied exact receive epoch.
// ReceiveEpoch is authoritative; SPPConfig's legacy scalar epoch fields remain
// available for compatibility with SolveBroadcast. The current C solve entrypoint
// does not return the typed correction-size refusal payload; this method returns
// its native solve error without synthesizing diagnostics or pre-screening records.
func (s *SSRCorrectionStore) SolveBroadcastAtExactEpoch(broadcast *BroadcastEphemeris, config SPPConfig, stalenessSeconds float64, missing SSRMissingCorrectionAction, allowRegionalProvider bool, regionalProviderID uint16, sizePolicy SSRCorrectionSizePolicy, receiveEpoch *ExactEpoch) (SPPSolution, error) {
	if s == nil || s.handle == nil || broadcast == nil || broadcast.handle == nil || receiveEpoch == nil || receiveEpoch.handle == nil {
		return SPPSolution{}, ErrClosed
	}
	if err := validateSSRMissingAction(missing); err != nil {
		return SPPSolution{}, err
	}
	if sizePolicy != SSRCorrectionSizeStrict && sizePolicy != SSRCorrectionSizeLenient {
		return SPPSolution{}, invalidArgument("invalid SSR correction size policy")
	}
	if config.Models != (SPPModelOptions{}) {
		input, err := nativeSppV2(SPPInputsV2{Base: config, Models: config.Models})
		if err != nil {
			return SPPSolution{}, publicError(err)
		}
		value, err := s.handle.SolveBroadcastV2AtExactEpoch(
			broadcast.handle, input, stalenessSeconds, uint32(missing),
			allowRegionalProvider, regionalProviderID, uint32(sizePolicy), receiveEpoch.handle,
		)
		return publicSPPSolution(value), publicError(err)
	}
	value, err := s.handle.SolveBroadcastAtExactEpoch(
		broadcast.handle, nativeSPPConfig(config), stalenessSeconds,
		uint32(missing), allowRegionalProvider, regionalProviderID, uint32(sizePolicy), receiveEpoch.handle,
	)
	return publicSPPSolution(value), publicError(err)
}
