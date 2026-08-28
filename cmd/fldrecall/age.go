package main

import (
	"fmt"
	"math"
	"time"
)

// FormatAge calculates how old a given time is from "now" in a humain readable format.
func FormatAge(t time.Time) string {
	duration := time.Since(t)

	// Handle future times or exact matches
	if duration <= 0 {
		return "just now"
	}

	// Calculate seconds, minutes, hours, and days
	minutes := duration.Minutes()
	hours := duration.Hours()
	days := math.Floor(hours / 24)

	// Return high-detail minutes for short durations
	if minutes < 60 {
		m := int(math.Floor(minutes))
		if m <= 0 {
			return "just now"
		}
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	}

	// Return hours for medium durations
	if hours < 24 {
		h := int(math.Floor(hours))
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}

	// Return days for long durations
	d := int(days)
	if d == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", d)
}
