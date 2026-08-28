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

	// Calculate base time units
	minutes := duration.Minutes()
	hours := duration.Hours()
	days := math.Floor(hours / 24)
	weeks := math.Floor(days / 7)
	months := math.Floor(days / 30.44) // Average days in a month

	// High detail: Minutes (up to 59m)
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

	// Medium detail: Hours (up to 23h)
	if hours < 24 {
		h := int(math.Floor(hours))
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}

	// Coarse detail: Days (up to 6d)
	if days < 7 {
		d := int(days)
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	}

	// Coarse detail: Weeks (up to ~4 weeks)
	if days < 30 {
		w := int(weeks)
		if w == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", w)
	}

	// Low detail: Months
	m := int(months)
	if m <= 1 {
		return "1 month ago"
	}
	return fmt.Sprintf("%d months ago", m)
}
