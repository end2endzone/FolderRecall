package main

import (
	"slices"
	"sort"
	"time"
)

const InvalidId = -1

type Directory struct {
	Id   int
	Path string
}

type Snapshot struct {
	Id          int
	TimeStamp   time.Time
	Directories []Directory
}

func (s *Snapshot) AddDirectory(path string) {
	dir := Directory{
		Id:   InvalidId,
		Path: path,
	}
	s.Directories = append(s.Directories, dir)
}

func (s *Snapshot) FindDirectoryIndex(path string) int {
	index := slices.IndexFunc(s.Directories, func(d Directory) bool {
		return path == d.Path
	})
	return index
}

func (s *Snapshot) HasDirectory(path string) bool {
	index := s.FindDirectoryIndex(path)
	if index == -1 {
		return false
	}
	return true
}

// Sort sorts directory entries alphabetically
func (s *Snapshot) Sort() {
	sort.Slice(s.Directories, func(i, j int) bool {
		dir1 := s.Directories[i].Path
		dir2 := s.Directories[j].Path
		return dir1 < dir2
	})
}

// Unique removes duplicate directories from the snapshot.
func (s *Snapshot) Unique() {
	// create a Set that contains <path> to know if a given directory was already seen.
	// Go does not supports sets, instead you must create a the following map where `struct{}` consumes 0 bytes of memory.
	seen := make(map[string]struct{})

	var duplicatesIndex []int

	// Identify the index of duplicate entries
	for i, dir := range s.Directories {
		_, exists := seen[dir.Path]
		if exists {
			duplicatesIndex = append(duplicatesIndex, i)
		} else {
			// remember this directory
			seen[dir.Path] = struct{}{}
		}
	}

	// Loop through duplicate indices backwards to safely remove elements
	for _, idx := range slices.Backward(duplicatesIndex) {
		// Remove the element at 'idx' by splicing it out to preserve the order
		s.Directories = append(s.Directories[:idx], s.Directories[idx+1:]...)
	}
}
