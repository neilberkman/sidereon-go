package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// Version is the version reported by the linked Sidereon C library.
type Version struct {
	// Major is the semantic-version major component.
	Major uint32
	// Minor is the semantic-version minor component.
	Minor uint32
	// Patch is the semantic-version patch component.
	Patch uint32
	// String is the native formatted semantic-version string.
	String string
}

// LibraryVersion returns the linked C library version.
func LibraryVersion() Version {
	v := native.LibraryVersion()
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch, String: v.String}
}
