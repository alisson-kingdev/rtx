package utils

import "runtime/debug"

// GetVersion returns the version from go.mod or fallback
func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "1.0.0"
}
