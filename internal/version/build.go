package version

import (
	"runtime/debug"
)

// GetVersionFromMetadata get the version value from the metadata injected at build time.
// For example:
//   - v0.1.1-0.20260815155314-4ad116a8bdf3
//   - v0.2.0-alpha
//
// When compiled from soure code in git version control, the version will be formattated such as `v0.1.1-0.20260815155314-4ad116a8bdf3`.
//
// When compiled using `go install URL@tag`, the version variable contains the name of the tag used.
// For example, when compiled with command
// `go install github.com/username/reponame/cmd/my-app@v0.2.0-alpha`, the returned version will be `v0.2.0-alpha`.
//
// When using `latest` for tag, the version metadata contains the latest tag that is not `alpha`, `beta`, etc.
// For example, when the following tag exists:
// - v0.2.0
// - v0.3.0
// - v0.3.1-alpha
// - v0.3.1-beta
// Go's toolchain/compiler will resolve the latest version as v0.3.0 and will set version metadata to `v0.3.0`.
//
// Warning: version can default to `(devel)` if compilation of source code is make outside of a version control system.
//
// Returns empty values on error or when metadata is not available.
func GetVersionFromMetadata() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// build-info metadata not available
		return ""
	}

	return info.Main.Version
}

// GetGitHashAndDateFromMetadata get the values from git's version control system from the metadata injected at build time.
// For example:
//   - revision: `6cb0ce064a20aeb026d039c12e7ab83b10ad1c63`
//   - datetime: `2026-08-16T15:12:22Z`
//
// Returns empty values on error or when metadata is not available.
func GetGitHashAndDateFromMetadata() (revision, datetime string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Metadata not available
		return
	}

	// Is compiled from source code in Git ?
	isGitVersionControl := false
	var tmp_revision string
	var tmp_datetime string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs":
			isGitVersionControl = (setting.Value == "git")
		case "vcs.revision":
			tmp_revision = setting.Value
		case "vcs.time":
			tmp_datetime = setting.Value
		}
	}

	if isGitVersionControl {
		revision = tmp_revision
		datetime = tmp_datetime
	}

	return
}

// GetBuildMetadata get the full metadata json data as a string.
// Returns empty values on error or when metadata is not available.
func GetBuildMetadata() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Metadata not available
		return ""
	}

	metadata := info.String()
	return metadata
}
