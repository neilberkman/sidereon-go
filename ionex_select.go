package sidereon

import "sidereon.dev/go/v3/internal/native"

// SelectIONEX selects an IONEX product usable at an integer J2000 epoch.
func SelectIONEX(products []*IONEX, requestedEpochJ2000S int64, policy StalenessPolicy) (*IONEX, StalenessMetadata, error) {
	nativeProducts := make([]*native.Ionex, len(products))
	for index, product := range products {
		if product != nil {
			nativeProducts[index] = product.handle
		}
	}
	selected, metadata, err := native.SelectIONEX(nativeProducts, requestedEpochJ2000S, native.StalenessPolicy{MaxStalenessS: policy.MaxStalenessS})
	if err != nil {
		return nil, StalenessMetadata{}, publicError(err)
	}
	return &IONEX{handle: selected}, stalenessMetadata(metadata), nil
}

// SelectIONEXOverRange selects an IONEX product usable across an integer J2000
// epoch range.
func SelectIONEXOverRange(products []*IONEX, startEpochJ2000S, endEpochJ2000S int64, policy StalenessPolicy) (*IONEX, StalenessMetadata, error) {
	nativeProducts := make([]*native.Ionex, len(products))
	for index, product := range products {
		if product != nil {
			nativeProducts[index] = product.handle
		}
	}
	selected, metadata, err := native.SelectIONEXOverRange(nativeProducts, startEpochJ2000S, endEpochJ2000S, native.StalenessPolicy{MaxStalenessS: policy.MaxStalenessS})
	if err != nil {
		return nil, StalenessMetadata{}, publicError(err)
	}
	return &IONEX{handle: selected}, stalenessMetadata(metadata), nil
}

// SelectIONEXAtInstant selects a product against a lossless scale-tagged instant.
func SelectIONEXAtInstant(products []*IONEX, requested ClockEpoch, policy StalenessPolicy) (*IONEX, StalenessMetadata, *IONEXEpochError, error) {
	nativeProducts := make([]*native.Ionex, len(products))
	for index, product := range products {
		if product != nil {
			nativeProducts[index] = product.handle
		}
	}
	selected, metadata, epochError, err := native.SelectIONEXAtInstant(nativeProducts, nativeClockEpoch(requested), native.StalenessPolicy{MaxStalenessS: policy.MaxStalenessS})
	if err != nil {
		return nil, StalenessMetadata{}, publicIONEXEpochError(epochError), publicError(err)
	}
	return &IONEX{handle: selected}, stalenessMetadata(metadata), publicIONEXEpochError(epochError), nil
}

// SelectIONEXOverInstantRange selects a product usable across an exact instant range.
func SelectIONEXOverInstantRange(products []*IONEX, start, end ClockEpoch, policy StalenessPolicy) (*IONEX, StalenessMetadata, *IONEXEpochError, error) {
	nativeProducts := make([]*native.Ionex, len(products))
	for index, product := range products {
		if product != nil {
			nativeProducts[index] = product.handle
		}
	}
	selected, metadata, epochError, err := native.SelectIONEXOverInstantRange(nativeProducts, nativeClockEpoch(start), nativeClockEpoch(end), native.StalenessPolicy{MaxStalenessS: policy.MaxStalenessS})
	if err != nil {
		return nil, StalenessMetadata{}, publicIONEXEpochError(epochError), publicError(err)
	}
	return &IONEX{handle: selected}, stalenessMetadata(metadata), publicIONEXEpochError(epochError), nil
}
