//go:build !darwin && !linux

package sidereon

func reviewMaxRSSBytes() (uint64, bool, error) {
	return 0, false, nil
}
