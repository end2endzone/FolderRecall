package recall

import (
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func CreateSnapshotNow() (Snapshot, error) {
	now := time.Now().Round(0)

	snapshot := Snapshot{
		Id:        InvalidId,
		Timestamp: now,
	}

	// Create shell.application object
	unknown, err := oleutil.CreateObject("Shell.Application")
	if err != nil {
		return snapshot, err
	}
	defer unknown.Release()

	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return snapshot, err
	}
	defer shell.Release()

	// Call Windows() method
	windows, err := oleutil.CallMethod(shell, "Windows")
	if err != nil {
		return snapshot, err
	}
	windowsDispatch := windows.ToIDispatch()
	defer windowsDispatch.Release()

	// Get Count of open windows
	countVal, err := oleutil.GetProperty(windowsDispatch, "Count")
	if err != nil {
		return snapshot, err
	}
	count := int(countVal.Val)

	// Loop through each window
	for i := 0; i < count; i++ {
		item, err := oleutil.CallMethod(windowsDispatch, "Item", i)
		if err != nil {
			continue
		}
		itemDispatch := item.ToIDispatch()
		defer itemDispatch.Release()

		// Get Document
		doc, err := oleutil.GetProperty(itemDispatch, "Document")
		if err != nil {
			continue
		}
		docDispatch := doc.ToIDispatch()
		defer docDispatch.Release()

		// Get Folder
		folder, err := oleutil.GetProperty(docDispatch, "Folder")
		if err != nil {
			continue
		}
		folderDispatch := folder.ToIDispatch()
		defer folderDispatch.Release()

		// Get Self
		self, err := oleutil.GetProperty(folderDispatch, "Self")
		if err != nil {
			continue
		}
		selfDispatch := self.ToIDispatch()
		defer selfDispatch.Release()

		// Get Path
		path, err := oleutil.GetProperty(selfDispatch, "Path")
		if err != nil {
			continue
		}

		dir := Directory{
			Id:   InvalidId,
			Path: path.ToString(),
		}
		snapshot.Directories = append(snapshot.Directories, dir)
	}

	// Sort elements for pretty display
	snapshot.Sort()

	// Remove potential duplicates.
	// The databse does not duplicate path.
	snapshot.Unique()

	return snapshot, nil
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
