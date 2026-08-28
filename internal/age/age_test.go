package age

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatAge(t *testing.T) {

	t.Run("less than 1 minute", func(t *testing.T) {
		timestamp := time.Now().Add(time.Duration(-53 * time.Second))
		str := FormatAge(timestamp)
		require.Equal(t, "just now", str)
	})

	t.Run("less than 1 hour", func(t *testing.T) {
		timestamp := time.Now().Add(time.Duration(-53 * time.Minute))
		str := FormatAge(timestamp)
		require.Equal(t, "53 minutes ago", str)
	})

	t.Run("exactly 1 hour ago", func(t *testing.T) {
		timestamp := time.Now().Add(time.Duration(-1 * time.Hour))
		str := FormatAge(timestamp)
		require.Equal(t, "1 hour ago", str)
	})

	t.Run("less than 24 hours ago", func(t *testing.T) {
		timestamp := time.Now().Add(time.Duration(-16 * time.Hour))
		str := FormatAge(timestamp)
		require.Equal(t, "16 hours ago", str)
	})

	t.Run("exactly 1 day ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(0, 0, -1)
		str := FormatAge(timestamp)
		require.Equal(t, "1 day ago", str)
	})

	t.Run("3 days ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(0, 0, -3)
		str := FormatAge(timestamp)
		require.Equal(t, "3 days ago", str)
	})

	t.Run("exactly 7 days ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(0, 0, -7)
		str := FormatAge(timestamp)
		require.Equal(t, "1 week ago", str)
	})

	t.Run("less than 30 days ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(0, 0, -26)
		str := FormatAge(timestamp)
		require.Equal(t, "3 weeks ago", str)

		timestamp = time.Now().AddDate(0, 0, -28) // 4*7 = 28
		str = FormatAge(timestamp)
		require.Equal(t, "4 weeks ago", str)
	})

	t.Run("exactly 1 month ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(0, -1, 0)
		str := FormatAge(timestamp)
		require.Equal(t, "1 month ago", str)
	})

	t.Run("less than 12 months ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(0, -5, 0)
		str := FormatAge(timestamp)
		require.Equal(t, "5 months ago", str)
	})

	t.Run("exactly 1 year ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(-1, 0, 0)
		str := FormatAge(timestamp)
		require.Equal(t, "1 year ago", str)
	})

	t.Run("less than 10 years ago", func(t *testing.T) {
		timestamp := time.Now().AddDate(-3, 0, 0)
		str := FormatAge(timestamp)
		require.Equal(t, "3 years ago", str)
	})
}

func TestElapsedTimeInMonths(t *testing.T) {

	t.Run("same day", func(t *testing.T) {
		t1 := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
		t2 := t1

		months := ElapsedTimeInMonths(t1, t2)
		require.InDelta(t, 0.0, months, 0.0001)
	})

	t.Run("exactly 1 month", func(t *testing.T) {
		t1 := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
		t2 := t1.AddDate(0, 1, 0)

		months := ElapsedTimeInMonths(t1, t2)
		require.InEpsilon(t, 1.0, months, 0.01)
	})

	t.Run("exactly n months", func(t *testing.T) {
		t1 := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
		for i := 1; i < 15; i++ {
			t2 := t1.AddDate(0, i, 0)

			months := ElapsedTimeInMonths(t1, t2)
			require.InEpsilon(t, float64(i), months, 0.01)
		}
	})

	t.Run("exactly 1 month and 1 day", func(t *testing.T) {
		t1 := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
		t2 := t1.AddDate(0, 1, 1)

		months := ElapsedTimeInMonths(t1, t2)
		require.Greater(t, months, 1.0)
	})

	t.Run("exactly close but not exactly 3 months (1 day less than 3 months)", func(t *testing.T) {
		t1 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		t2 := t1.AddDate(0, 3, -1) // expecting a little lower than 3.0

		months := ElapsedTimeInMonths(t1, t2)
		require.Less(t, months, 3.0)
		require.InEpsilon(t, 2.9, months, 0.1)
	})

	t.Run("fractions", func(t *testing.T) {
		// August 2026 is 31 days long.
		// Each day should be 1/31 (0.032258064516129) month long
		delta := 0.032258064516129

		t1 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		for days := 1; days < 30; days++ { // loop trhough all days of August 2026
			t2 := t1.AddDate(0, 0, days)

			months := ElapsedTimeInMonths(t1, t2)

			expected := float64(days) * delta

			require.InEpsilon(t, expected, months, 0.0001)
		}
	})

	t.Run("swap/commutative/symmetric", func(t *testing.T) {
		t1 := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
		t2 := t1.AddDate(0, 1, 0)

		months1 := ElapsedTimeInMonths(t1, t2)
		months2 := ElapsedTimeInMonths(t2, t1)
		require.Equal(t, months1, months2)
	})
}

func TestElapsedTimeInDays(t *testing.T) {
	// Base time for calculations
	baseTime := time.Date(2026, 1, 1, 12, 34, 56, 7890, time.UTC)

	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected float64
	}{
		{
			name:     "zero time difference",
			t1:       baseTime,
			t2:       baseTime,
			expected: 0.0,
		},
		{
			name:     "exactly one day later",
			t1:       baseTime,
			t2:       baseTime.Add(24 * time.Hour),
			expected: 1.0,
		},
		{
			name:     "exactly half a day later (fractional)",
			t1:       baseTime,
			t2:       baseTime.Add(12 * time.Hour),
			expected: 0.5,
		},
		{
			name:     "Granular precise duration (1 hour and 30 minutes)",
			t1:       baseTime,
			t2:       baseTime.Add(90 * time.Minute),
			expected: 90.0 / (60 * 24.0), // 90 minutes of 24 hours
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ElapsedTimeInDays(tt.t1, tt.t2)

			// Use an epsilon delta of 1e-9 to handle underlying float64 rounding
			epsilon := 1e-9
			require.InDelta(t, tt.expected, actual, epsilon, "ElapsedTimeInDays(%v, %v) got unexpected result", tt.t1, tt.t2)
		})
	}

	t.Run("swap/commutative/symmetric", func(t *testing.T) {
		t1 := time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
		t2 := t1.AddDate(0, 1, 0)

		days1 := ElapsedTimeInDays(t1, t2)
		days2 := ElapsedTimeInDays(t2, t1)
		require.Equal(t, days1, days2)
	})

}
