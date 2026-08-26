package utils

import (
	"strings"
	"unicode"
)

// RemoveNonPrintableCharacters removes non-printable characters from a string.
// It returns a string which all characters can be printed on the console or a terminal.
func RemoveNonPrintableCharacters(s string) string {
	out := ReplaceNonPrintableCharacters(s, "")
	return out
}

// ReplaceNonPrintableCharacters takes an input string and replaces any non-printable characters
// (characters not writable on the console or a terminal) with the specified replacement string.
func ReplaceNonPrintableCharacters(s string, replacement string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, rune := range s {
		if unicode.IsPrint(rune) {
			b.WriteRune(rune)
		} else {
			b.WriteString(replacement)
		}
	}
	out := b.String()
	return out
}

// RemoveNonAsciiCharacters removes non-ascii characters from a string
func RemoveNonAsciiCharacters(s string) string {
	out := ReplaceNonAsciiCharacters(s, "")
	return out
}

// ReplaceNonAsciiCharacters takes an input string and replaces any non-ASCII characters
// (characters with a code point value greater than 127) with the specified replacement string.
func ReplaceNonAsciiCharacters(s string, replacement string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, rune := range s {
		if rune <= 127 {
			b.WriteRune(rune)
		} else {
			b.WriteString(replacement)
		}
	}
	out := b.String()
	return out
}

// RemoveInvalidFileSystemCharacters removes characters from a string that are not compatible with most filesystems.
// It returns a string which can safely be used to create files or directories.
func RemoveInvalidFileSystemCharacters(s string) string {
	out := ReplaceInvalidFileSystemCharacters(s, "")
	return out
}

// ReplaceNonAsciiCharacters takes an input string and replaces any characters that are not filesystem compatible
// (characters such as control characters in a filesystem) with the specified replacement string.
func ReplaceInvalidFileSystemCharacters(s string, replacement string) string {
	// Replace invalid  by underscore
	replacer := strings.NewReplacer(
		// Windows invalid characters
		"\\", replacement, "/", replacement, ":", replacement, "*", replacement, "?", replacement,
		"\"", replacement, "<", replacement, ">", replacement, "|", replacement,
		// "&", replacement,
		// ";", replacement,
		// "'", replacement,
		// "[", replacement,
		// "]", replacement,

		// Linux invalid characters
		"/", replacement, "\000", replacement,
	)
	out := replacer.Replace(s)
	return out
}

// SanitizeString takes an input string and replaces all non-ASCII characters from a string, except for Latin alphabet letters
// (characters with accents such as ó, ò, ô, ö, õ, and ō) with the specified replacement string.
func SanitizeString(s string, replacement string) string {
	var b strings.Builder
	b.Grow(len(s)) // Pre-allocate memory for performance

	for _, r := range s {
		// Allow all standard ASCII characters (0 - 127)
		if r <= 127 {
			b.WriteRune(r)
			continue
		}

		// Allow non-ASCII characters ONLY if they are Latin letters (covers accents)
		if unicode.Is(unicode.Latin, r) && unicode.IsLetter(r) {
			b.WriteRune(r)
			continue
		}

		// Non allowed, replace by replacement string.
		b.WriteString(replacement)
	}

	out := b.String()
	return out
}
