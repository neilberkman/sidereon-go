package sidereon

import "github.com/neilberkman/sidereon-go/internal/native"

// Version is the version reported by the linked Sidereon C library.
type Version struct {
	// Major is the major value for Version.
	Major uint32
	// Minor is the minor value for Version.
	Minor uint32
	// Patch is the patch value for Version.
	Patch uint32
	// String is the string value for Version.
	String string
}

// LibraryVersion returns the linked C library version.
func LibraryVersion() Version {
	v := native.LibraryVersion()
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch, String: v.String}
}
