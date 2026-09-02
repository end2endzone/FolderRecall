//go:build !windows

package spinner

import "os"

// On Linux and Darwin, terminal run with a UTF-8 locale and a font that supports Braille Pattern glyphs.
// So we mostly defaults to true. The one notable exception is the bare Linux virtual console (the VGA text-mode tty).
// It does not support the Braille Patterns block. It can be identified with `TERM=linux`.
func init() {
	SupportsBrailleCharacters = detectBrailleSupport()
}

func detectBrailleSupport() bool {
	if os.Getenv("TERM") == "linux" {
		return false // bare Linux VGA text-mode console
	}
	return true
}
