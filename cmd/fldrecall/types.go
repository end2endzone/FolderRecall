package main

import (
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

func (ds Snapshot) Sort() {
	sort.Slice(ds.Directories, func(i, j int) bool {
		dir1 := ds.Directories[i].Path
		dir2 := ds.Directories[j].Path
		return dir1 < dir2
	})
}
