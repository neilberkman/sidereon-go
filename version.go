package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// Version is the version reported by the linked Sidereon C library.
type Version struct {
	Major  uint32
	Minor  uint32
	Patch  uint32
	String string
}

// LibraryVersion returns the linked C library version.
func LibraryVersion() Version {
	v := native.LibraryVersion()
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch, String: v.String}
}
