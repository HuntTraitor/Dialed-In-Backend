package vcs

import (
	"fmt"
	"runtime/debug"
)

var (
	time     string
	revision string
	modified bool
)

// Version reads the build info during runtime and determines if this version is the current committed version or
// an uncommitted version
func Version() string {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.time":
				time = s.Value
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					modified = true
				}
			}
		}
	}

	if modified {
		return fmt.Sprintf("%s-%s-dirty", time, revision)
	}
	return revision
}
