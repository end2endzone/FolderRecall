package age

import (
	"fmt"
	"math"
	"time"
)

// ElapsedTimeInDays returns the absolute number of calendar days between two times.
func ElapsedTimeInDays(t1, t2 time.Time) float64 {
	// Subtract the dates to get the duration
	duration := t2.Sub(t1)

	// Convert duration to hours, divide by 24, and take the absolute value
	days := duration.Hours() / 24

	return math.Abs(days)
}

// ElapsedTimeInMonths returns the absolute number of calendar months between two times.
func ElapsedTimeInMonths(t1, t2 time.Time) float64 {
	// Normalize both times to midnight UTC to strip time-of-day and DST impacts
	// (We drop hours/minutes/seconds becuase we do not want to check if late's hours are >= to early's hours, same for minutes, same for seconds)
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC)
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.UTC)

	// Sort the times chronologically so `early` always comes before `late`.
	// This prevents us from having to deal with math.Abs() function
	early, late := t1, t2
	if early.After(late) {
		early, late = late, early
	}

	// Calculate the month difference without looking at the day
	months := (late.Year()-early.Year())*12 + int(late.Month()-early.Month())

	// If late's day of the month hasn't reached early's day, a full month hasn't passed.
	// In other words, if we are the 25th, late must be set on the 25th or after in order to be a full month.
	// If late is "the 12th next month", that is not a full month.
	if late.Day() < early.Day() {
		months--
	}

	// Compute the fractionnal part.

	// Find the current anniversary date for the current number of month identified
	elapsedMonthAnniversary := early.AddDate(0, months, 0)

	// Find the next anniversary date (exactly 1 month after the current one)
	nextMonthAnniversary := early.AddDate(0, months+1, 0)

	// Calculate total days in this specific month-long window whcih spans [elapsedMonth, nextMonth]
	totalDaysInMonthWindow := ElapsedTimeInDays(elapsedMonthAnniversary, nextMonthAnniversary)

	// Calculate how many days have passed since the the elapsed months anniversary
	daysElapsedInWindow := ElapsedTimeInDays(elapsedMonthAnniversary, late)

	// Compute fraction
	fractionalMonth := daysElapsedInWindow / totalDaysInMonthWindow

	// Combine whole months with the precise fractional month
	totalMonth := float64(months) + fractionalMonth

	return totalMonth
}

// FormatAge calculates how old a given time is from "now" in a humain readable format.
func FormatAge(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t) // time.Since(t)

	// Handle future times or exact matches
	if duration <= 0 {
		return "just now"
	}

	// Calculate base time units
	minutes := duration.Minutes()
	hours := duration.Hours()

	// Compute average days & month estimation
	days := ElapsedTimeInDays(t, now)
	months := ElapsedTimeInMonths(t, now)
	weeks := days / 7
	years := months / 12

	// Minutes (up to 59m)
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

	// Hours (up to 23h)
	if hours < 24 {
		h := int(math.Floor(hours))
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}

	// Days (up to 6d)
	if days < 7 {
		d := int(math.Floor(days))
		if d == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", d)
	}

	// Special case for 1 month that must take priority over 4.428 weeks
	if int(months) == 1 {
		return "1 month ago"
	}

	// Weeks (up to 4 weeks)
	if weeks < 5 {
		w := int(math.Floor(weeks))
		if w == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", w)
	}

	// Months (up to 11 months)
	if months < 12 {
		m := int(months)
		if m == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", m)
	}

	// Years
	y := int(years)
	if y <= 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", y)
}
