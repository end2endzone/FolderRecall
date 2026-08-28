package main

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
}
