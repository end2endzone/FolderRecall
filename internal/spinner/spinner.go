package spinner

import (
	"fmt"
	"time"
)

// Spinner manages a single-threaded Unicode console spinner without text.
type Spinner struct {
	frames  []string
	index   int
	message string
}

// New initializes a message-free spinner with Unicode Braille characters.
func New(message string) *Spinner {
	var frames []string

	if SupportsBrailleCharacters {
		// Use premium Braille symbols if terminal supported it perfectly
		frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇"}
	} else {
		// Bulletproof fallback: Will display perfectly on Windows system that do not supports UTF-8 Braille symbols
		frames = []string{"|", "/", "-", "\\"}
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

// Finish clears the spinner character.
func (s *Spinner) Finish() {
	fmt.Print("\r\n")
}

// AnimateUntil loops, ticks, and sleeps at the given speed until the endWait t
// then automatically cleans up the line with Finish().
func (s *Spinner) AnimateUntil(speed time.Duration, endWait time.Time) {
	for time.Now().Before(endWait) {
		s.Tick()
		time.Sleep(speed)
	}
	s.Finish()
}
