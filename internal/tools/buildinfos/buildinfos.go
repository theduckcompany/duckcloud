package buildinfos

import (
	"errors"
)

var (
	version string = "unknown"
	// buildTime string = "unknown"
	isRelease string = "false"
)

var ErrNotSet = errors.New("not set")

// IsRelease is set to true if the binary is an official release.
// All the other builds will return false.
func IsRelease() bool {
	return isRelease == "true"
}

// Version of the release or "unknown"
func Version() string {
	return version
}

// XXX: Unused
//
// // BuildTime is ISO-8601 UTC string representation of the time of
// // the build or "time.Time{}"
// func BuildTime() (time.Time, error) {
// 	if buildTime == "unknown" {
// 		return time.Time{}, ErrNotSet
// 	}

// 	raw, err := time.Parse(time.RFC3339, buildTime)

// 	return raw, err
// }
