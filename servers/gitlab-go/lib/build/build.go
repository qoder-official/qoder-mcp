// Package build provides build-time information about the binary.
// The version, commit, and date values are set at build time via
// linker flags and default to development values when built locally.
package build

// Default values. Overridden by Goreleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Version returns the version of the binary.
func Version() string {
	return version
}

// Commit returns the Git commit hash at which the binary was built.
func Commit() string {
	return commit
}

// Date returns the date when the binary was built.
func Date() string {
	return date
}
