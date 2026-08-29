package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

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
