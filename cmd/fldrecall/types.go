package main

import "sort"

type DirectorySnapshot struct {
	directories []string
}

func (ds DirectorySnapshot) Sort() {
	sort.Slice(ds.directories, func(i, j int) bool {
		dir1 := ds.directories[i]
		dir2 := ds.directories[j]
		return dir1 < dir2
	})
}
