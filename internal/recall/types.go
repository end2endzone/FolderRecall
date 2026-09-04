package recall

import (
	"fmt"
	"os/exec"
	"slices"
	"sort"
	"time"

	"github.com/end2endzone/FolderRecall/internal/age"
)

const InvalidId = -1

type Directory struct {
	Id   int    `json:"id"`
	Path string `json:"path"`
}

type Snapshot struct {
	Id          int         `json:"id"`
	Timestamp   time.Time   `json:"timestamp"`
	Directories []Directory `json:"directories"`
}

// DirectoryDiff holds the lists of added and removed directories resulting in comparing 2 snapshots.
type DirectoryDiff struct {
	Added   []string
	Removed []string
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

// String returns a string description for the current snapshot.
func (s *Snapshot) String() string {
	age := age.FormatAge(s.Timestamp)
	str := fmt.Sprintf("%s (%s) with %d directories", s.Timestamp, age, len(s.Directories))
	return str
}

// Restore restores a snapshot by opening File Explorer for each of the directories in the snapshot.
func (s *Snapshot) Restore() error {
	for _, dir := range s.Directories {
		// Run the explorer command
		cmd := exec.Command("explorer", dir.Path)

		// cmd.Start() launches the explorer process asynchronously and returns immediately without waiting for the File Explorer window to close.
		err := cmd.Start()
		if err != nil {
			return err
		}
	}

	return nil
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

// HasDirectoryEquals checks if the directory list is equal to another snapshot's directory list.
// Note: This checks for strict positional equality (order matters).
// To compare using unordered lists, use CompareDirectories().
func (s *Snapshot) HasDirectoryEquals(other *Snapshot) bool {
	if len(s.Directories) != len(other.Directories) {
		return false
	}

	for i := range s.Directories {
		if s.Directories[i] != other.Directories[i] {
			return false
		}
	}

	return true
}

// Less reports whether the snapshot's timestamp is before another snapshot's timestamp.
// Use for sorting slices of Snapshots.
func (s *Snapshot) Less(other *Snapshot) bool {
	return s.Timestamp.Before(other.Timestamp)
}

// Equals compares 2 snapshots to see if they are equals.
func (s *Snapshot) Equals(other *Snapshot) bool {
	if s.Timestamp != other.Timestamp {
		return false
	}

	sameDirectories := s.HasDirectoryEquals(other)
	return sameDirectories
}

// CompareDirectories compares the current snapshot against another snapshot, returning which directories were added or removed.
// The current snapshot is treated as the base/older state and the other snapshot is treated as the newer state.
func (s *Snapshot) CompareDirectories(other *Snapshot) DirectoryDiff {
	// Move local/other directories to a map.
	thisDirMap := make(map[string]struct{})
	for _, dir := range s.Directories {
		thisDirMap[dir.Path] = struct{}{}
	}

	otherDirMap := make(map[string]struct{})
	for _, dir := range other.Directories {
		otherDirMap[dir.Path] = struct{}{}
	}

	var added []string
	var removed []string

	// Find directories present in 'other' but not in 's' (Added)
	for _, dir := range other.Directories {
		_, exists := thisDirMap[dir.Path]
		if !exists {
			added = append(added, dir.Path)
		}
	}

	// Find directories present in 's' but not in 'other' (Removed)
	for _, dir := range s.Directories {
		_, exists := otherDirMap[dir.Path]
		if !exists {
			removed = append(removed, dir.Path)
		}
	}

	return DirectoryDiff{
		Added:   added,
		Removed: removed,
	}
}

// SimplifySnapshotsByDirectory removes consecutive snapshots in a slice that are similar to the following one.
// When multiple consecutive snapshots are similar, only the last one (bigger timestamp) is preserved.
func SimplifySnapshotsByDirectory(snapshots []*Snapshot) []*Snapshot {
	var lastMeaningfulSnapshot *Snapshot

	meaningfulSnapshots := []*Snapshot{}

	// Check each snapshots one by one to see if they are meaningful
	for _, s := range snapshots {

		// No existing meaningful snapshot yet
		if lastMeaningfulSnapshot == nil || len(meaningfulSnapshots) == 0 {
			// keep the first snapshot
			lastMeaningfulSnapshot = s
			meaningfulSnapshots = append(meaningfulSnapshots, s)
			continue
		}

		// Is this snapshot different than the last one ?
		diff := lastMeaningfulSnapshot.CompareDirectories(s)
		if len(diff.Added) == 0 && len(diff.Removed) == 0 {
			// This snapshot is identical to the last meaningful one (same directories / no changes.)
			// It must replace the lastMeaningfulSnapshot
			lastMeaningfulSnapshot = s
			meaningfulSnapshots[len(meaningfulSnapshots)-1] = s
			continue
		}

		// This one is meaningful
		lastMeaningfulSnapshot = s
		meaningfulSnapshots = append(meaningfulSnapshots, s)
	}

	return meaningfulSnapshots
}
