package main

import (
	"sort"
	"time"
)

type Snapshot struct {
	TimeStamp time.Time
	paths     []string
}

func (ds Snapshot) Sort() {
	sort.Slice(ds.paths, func(i, j int) bool {
		dir1 := ds.paths[i]
		dir2 := ds.paths[j]
		return dir1 < dir2
	})
}
