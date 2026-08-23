package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ToStringSlice(s *Snapshot) []string {
	var paths []string
	for _, dir := range s.Directories {
		paths = append(paths, dir.Path)
	}
	return paths
}

func TestSort(t *testing.T) {
	now := time.Now().Round(0)
	s := Snapshot{
		Id:        InvalidId,
		TimeStamp: now,
	}

	s.AddDirectory("/tmp")
	s.AddDirectory("/home")
	s.AddDirectory("/var")
	s.AddDirectory("/media")
	s.AddDirectory("/boot")

	s.Sort()

	actual := ToStringSlice(&s)

	// assert
	expected := []string{
		"/boot",
		"/home",
		"/media",
		"/tmp",
		"/var"}

	// Assert element in the list
	require.ElementsMatch(t, expected, actual)
}

func TestUnique(t *testing.T) {
	now := time.Now().Round(0)
	s := Snapshot{
		Id:        InvalidId,
		TimeStamp: now,
	}

	s.AddDirectory("/tmp")
	s.AddDirectory("/home")
	s.AddDirectory("/var")
	s.AddDirectory("/home")
	s.AddDirectory("/media")
	s.AddDirectory("/boot")

	s.Unique()

	actual := ToStringSlice(&s)

	// assert
	expected := []string{
		"/tmp",
		"/home",
		"/var",
		"/media",
		"/boot"}

	// Assert both list are equals
	require.ElementsMatch(t, expected, actual)
}
