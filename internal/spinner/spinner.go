// On platforms/terminals where Unicode Braille Pattern glyphs (U+2800-U+28FF)
// are known to render correctly, the spinner uses the classic Braille
// animation frames ("⠋⠙⠹⠸⠼⠴⠦⠧⠇"). Everywhere else it falls back to a plain
// ASCII animation ("|/-\") that works on every terminal, including legacy
// Windows consoles. See platform_windows.go and platform_other.go for how
// that decision is made on each platform.
package spinner

import (
	"fmt"
	"strings"
	"time"
)

// SupportsBrailleCharacters reports whether the current process's console is
// expected to render Unicode Braille Pattern glyphs correctly. It is
// determined once, at package init time, by platform-specific detection
// logic (see platform_windows.go / platform_other.go). It is exported as a
// variable rather than a function so callers can override the auto-detected
// value if needed, e.g. to force ASCII frames:
//
//	spinner.SupportsBrailleCharacters = false
var SupportsBrailleCharacters bool

// brailleFrames are the classic Unicode Braille Pattern spinner frames.
var brailleFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇"}

// asciiFrames is the bulletproof fallback animation for terminals/fonts that
// do not reliably render the Braille Patterns Unicode block.
var asciiFrames = []string{"|", "/", "-", "\\"}

// Spinner manages a single-threaded console spinner with an optional leading message.
type Spinner struct {
	frames  []string
	index   int
	message string
}

// New creates a spinner that prefixes each animation frame with message.
// The frame set (Braille or ASCII) is chosen based on SupportsBrailleCharacters at the time New is called.
func New(message string) *Spinner {
	frames := asciiFrames
	if SupportsBrailleCharacters {
		frames = brailleFrames
	}

	return &Spinner{
		frames:  frames,
		message: message,
	}
}

// Tick advances the animation by one frame on the current line.
func (s *Spinner) Tick() {
	fmt.Printf("\r%s%s", s.message, s.frames[s.index])
	s.index = (s.index + 1) % len(s.frames)
}

// eraseMessageAndLastFrame erases the spinner message and last frame.
func (s *Spinner) EraseMessageAndLastFrame() {
	// Compute a string to erase the spinner message
	messageOverride := strings.Repeat(" ", len(s.message))

	// Compute a string to erase the last spinner frame
	longestFrameLength := 1
	for _, frame := range s.frames {
		if len(frame) > longestFrameLength {
			longestFrameLength = len(frame)
		}
	}
	frameOverride := strings.Repeat(" ", longestFrameLength)

	fmt.Printf("\r%s%s", messageOverride, frameOverride)
}

// Finish clears the spinner character and ends the message with a dot.
func (s *Spinner) Finish() {
	// Clear the previous message & frame since we will print a truncated version of "message".
	s.EraseMessageAndLastFrame()

	// Truncate the message. The message usually ends with a space or \t to create a space between the message and the spinner frame.
	trimmedMessage := strings.TrimRight(s.message, " \t")
	// Print the message wollowed by a dot (`.`) character.
	fmt.Printf("\r%s.\n", trimmedMessage)
}

// AnimateUntil ticks and sleeps at the given speed in a loop until the endWait time
// then automatically cleans up the line with Finish().
func (s *Spinner) AnimateUntil(speed time.Duration, endWait time.Time) {
	for time.Now().Before(endWait) {
		s.Tick()
		time.Sleep(speed)
	}
	s.Finish()
}
