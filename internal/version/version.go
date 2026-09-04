package version

import (
	"fmt"
	"strings"
	"time"

	root "github.com/end2endzone/FolderRecall"

	"golang.org/x/mod/semver"
)

const (
	InvalidVersion    = "0.0.0"
	InvalidCommitHash = "0000000"
	InvalidBuildDate  = "1900-01-01"
)

// These variables are legacy variables that could be populated at build time using `-ldflags`.
// They are normally left unchanged in recent versions of this product.
var (
	Version    = InvalidVersion
	CommitHash = InvalidCommitHash
	BuildDate  = InvalidBuildDate
)

type ProductVersion struct {
	Version    string
	CommitHash string
	BuildDate  string
}

func (p ProductVersion) IsValid() bool {
	if p.Version != InvalidVersion && p.CommitHash != InvalidCommitHash && p.BuildDate != InvalidBuildDate {
		return true
	}
	return false
}

/*
// isNumberInRange validate that the given input str is a number in the range [min;max]
// Returns true when the input is in range. Returns false otherwise.
func isNumberInRange(str string, min int, max int) bool {
	num, err := strconv.Atoi(str)
	if err != nil {
		// Parsing error
		return false
	}

	if num < min || num > max {
		return false
	}
	return true
}

// isValidTimestamp detect if the given string is a timestamp. A timestamp is a string in format `YYYYMMDDhhmmss`.
// Returns true if the given input is a timestamp. Returns false otherwise.
func isValidTimestamp(input string) bool {
	if len(input) != 14 {
		return false
	}

	if input[0:2] != "20" {
		// Not from 2000 era.
		return false
	}

	year := input[0:4]
	month := input[4:6]
	day := input[6:8]
	hours := input[8:10]
	minutes := input[10:12]
	seconds := input[12:14]

	if !isNumberInRange(year, 2000, 2100) ||
		!isNumberInRange(month, 1, 12) ||
		!isNumberInRange(day, 1, 31) ||
		!isNumberInRange(hours, 1, 24) ||
		!isNumberInRange(minutes, 1, 59) ||
		!isNumberInRange(seconds, 1, 59) {
		return false
	}
	return true
}

func isHexString(input string) bool {
	for _, c := range input {
		valid := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		if !valid {
			return false
		}
	}
	return true
}

// isSemVerGitHash detect if the given string is a git hash from a Semantic Version string.
// Returns true if the given input is a git hash. Returns false otherwise.
func isSemVerGitHash(input string) bool {
	// hash are always 12 bytes long
	if len(input) != 12 {
		return false
	}

	// hash are always in hex format
	hex := isHexString(input)
	return hex
}

// findDateTimeInSemVer finds a datetime timestamp in the given semantic version string.
// For example, it returns `20260815155314` from the input string `v0.1.1-0.20260815155314-4ad116a8bdf3`.
// Returns a valid datetime value when found. Returns an empty string otherwise.
func findDateTimeInSemVer(input string) string {
	// Strip out `-`, `.` or `+` from the input string
	replacement := "/"
	input = strings.ReplaceAll(input, ".", replacement)
	input = strings.ReplaceAll(input, "-", replacement)
	input = strings.ReplaceAll(input, "+", replacement)

	// Search in fields
	fields := strings.Split(input, replacement)
	for _, value := range fields {
		if isValidTimestamp(value) {
			return value
		}
	}

	return ""
}

// findGitHashInSemVer finds a git hash in the given semantic version string.
// For example, it returns `4ad116a8bdf3` from the input string `v0.1.1-0.20260815155314-4ad116a8bdf3`.
// Returns a valid git hash value when found. Returns an empty string otherwise.
func findGitHashInSemVer(input string) string {
	// Strip out `-`, `.` or `+` from the input string
	replacement := "/"
	input = strings.ReplaceAll(input, ".", replacement)
	input = strings.ReplaceAll(input, "-", replacement)
	input = strings.ReplaceAll(input, "+", replacement)

	// Search in fields
	fields := strings.Split(input, replacement)
	for _, value := range fields {
		if isSemVerGitHash(value) {
			return value
		}
	}

	return ""
}
*/

func CreateInvalidProductVersion() ProductVersion {
	p := ProductVersion{
		Version:    InvalidVersion,
		CommitHash: InvalidCommitHash,
		BuildDate:  InvalidBuildDate,
	}
	return p
}

// GetProductVersionFromMetadata parses the go pseudo-version into a ProductVersion.
// A pseudo-version is a string in format "v[version]-[datetime]-[revision][dirtybit]".
// For example: "v1.2.3-20260815131415-de3798c09c08+dirty".
func GetProductVersionFromMetadata() ProductVersion {
	p := CreateInvalidProductVersion()

	version := GetVersionFromMetadata()
	revision, datetime := GetGitHashAndDateFromMetadata()
	if version == "" || revision == "" || datetime == "" {
		// Metadata not available or not compiled from git version control.
		// Return an invalid ProductVersion
		return p
	}

	// Version: if installed via `go install` tagged release (for example `@v1.2.3`)
	// Warning: version can default to `(devel)` if run locally or built without version flags.
	// For example, if you run your app using `go run .` or compile it locally without tags, Go sets info.Main.Version to the string "(devel)".
	p.Version = version

	// Remove `+dirty` from version and standardizes it
	p.Version = semver.Canonical(p.Version)

	// Strip the leading "v" to only get the digits and pre-release labels
	p.Version = strings.TrimPrefix(p.Version, "v")

	// Parse revision
	if revision != "" {
		// Truncate revision to 12 characters or less
		p.CommitHash = revision[0:min(12, len(revision))]
	}

	// Parse datetime
	{
		p.BuildDate = datetime

		// Example value: `2026-08-15T13:14:15Z`. Try to truncate for keeping only the date

		// Parse the string directly into a time.Time object
		buildTime, err := time.Parse(time.RFC3339, datetime)
		if err == nil {
			year, month, day := buildTime.Date()
			p.BuildDate = fmt.Sprintf("%d-%02d-%02d", year, month, day)
		}
	}

	// Remove datetime from version string
	{
		// datetime == "2026-08-16T15:12:22Z"
		// version == "v1.2.3-20260816151222-de3798c09c08"

		pattern := datetime
		pattern = strings.ReplaceAll(pattern, "-", "")
		pattern = strings.ReplaceAll(pattern, "T", "")
		pattern = strings.ReplaceAll(pattern, ":", "")
		pattern = strings.ReplaceAll(pattern, "Z", "")

		// datetime can sometime show as `-20260816154451` or as `.20260816154451`.
		pattern1 := "-" + pattern
		pattern2 := "." + pattern

		// Remove the pattern from the version
		p.Version = strings.ReplaceAll(p.Version, pattern1, "")
		p.Version = strings.ReplaceAll(p.Version, pattern2, "")
	}

	// Remove git hash from version string
	{
		// revision == "de3798c09c08aeb026d039c12e7ab83b10ad1c63"
		// version == "v1.2.3-20260816151222-de3798c09c08"

		for i := len(revision) - 1; i >= 0; i-- {
			pattern := "-" + revision[0:i+1]

			// Remove the pattern from the version
			before := len(p.Version)
			p.Version = strings.ReplaceAll(p.Version, pattern, "")
			after := len(p.Version)

			if before != after {
				// We found the pattern, stop trying
				break
			}
		}
	}

	return p
}

func GetProductVersion() ProductVersion {
	// Create a product version from local legacy variables.
	local := ProductVersion{
		Version:    Version,
		CommitHash: CommitHash,
		BuildDate:  BuildDate,
	}

	// If the product version was set with the legacy `-ldflags` command line at build time, use that.
	if local.IsValid() {
		return local
	}

	// Get a valid the product version from metadata
	metadata := GetProductVersionFromMetadata()
	if metadata.IsValid() {
		return metadata
	}

	// Try to build a ProductVersion from the version metadata.
	tagName := GetVersionFromMetadata()
	if tagName != "" && tagName != "(devel)" {
		// Strip the leading "v" to only get the digits and pre-release labels
		tagName = strings.TrimPrefix(tagName, "v")

		// This product version is still invalid because it has no CommitHash and no BuildDate.
		// It will require a different formatting
		p := CreateInvalidProductVersion()
		p.Version = tagName
		return p
	}

	// Build a ProductVersion from the VERSION file instead.
	// This product version is still invalid because it has no CommitHash and no BuildDate.
	// It will require a different formatting
	p := CreateInvalidProductVersion()
	p.Version = root.GetVersionFromVersionFile()
	return p
}

func GetProductVersionString() string {
	p := GetProductVersion()

	var msg string
	if p.IsValid() {
		// Version 0.3.1-rc1 (820b508a0f53, released 2026-08-18)
		msg = fmt.Sprintf("%s (%s, released %s)", p.Version, p.CommitHash, p.BuildDate)
	} else {
		// Invalid product version. Assume only p.Version is valid.
		msg = fmt.Sprintf("%s", p.Version)
	}

	return msg
}
