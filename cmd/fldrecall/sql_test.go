package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// GenerateMockSnapshots creates 10 mock snapshot structures with mixed data.
func GenerateMockSnapshots() []*Snapshot {
	// Pre-define a pool of sample directory paths
	dirPool := []string{
		"/dev",
		"/etc/nginx",
		"/home/user/docs",
		"/home/user/downloads",
		"/opt/app",
		"/tmp",
		"/usr/bin",
		"/var/log",
		"/var/www/html",
	}

	snapshots := make([]*Snapshot, 10)

	// Base starting time: 4 days ago at 09:00 AM
	baseTime := time.Now().AddDate(0, 0, -4).Truncate(time.Hour)
	baseTime = time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 9, 0, 0, 0, baseTime.Location())

	// Define specific combinations of indexes from our dirPool for each of the 10 snapshots
	combinations := [][]int{
		{0, 1, 2},    // Use directories 0, 1 and 3
		{1, 3},       // Use directories 1 and 3
		{0, 2, 4, 5}, //
		{2, 5, 6},    //
		{1, 4, 7},    //
		{0, 3, 6},    //
		{4, 8},       //
		{2, 3, 7},    //
		{0, 1, 5, 8}, //
		{1, 6, 7},    //
	}

	// Time offsets to simulate multiple days and different times of day
	// (e.g., increments of several hours to spread across a 3-day window)
	timeOffsets := []time.Duration{
		0 * time.Hour,  // Day 1, 09:00 AM
		4 * time.Hour,  // Day 1, 01:00 PM
		9 * time.Hour,  // Day 1, 06:00 PM
		22 * time.Hour, // Day 2, 07:00 AM
		27 * time.Hour, // Day 2, 12:00 PM
		33 * time.Hour, // Day 2, 06:00 PM
		48 * time.Hour, // Day 3, 09:00 AM
		52 * time.Hour, // Day 3, 01:00 PM
		58 * time.Hour, // Day 3, 07:00 PM
		72 * time.Hour, // Day 4, 09:00 AM
	}

	for i := 0; i < 10; i++ {
		// Populate metadata
		snapshots[i] = &Snapshot{
			Id:        -1,
			Timestamp: baseTime.Add(timeOffsets[i]),
		}

		// Map index combinations to concrete Directory structs
		for _, dirIndex := range combinations[i] {
			dir := Directory{
				Id:   -1, // Set to -1 to force the DB lookup or unique insert
				Path: dirPool[dirIndex],
			}
			snapshots[i].Directories = append(snapshots[i].Directories, dir)
		}
	}

	return snapshots
}

// GenerateMockFullHistorySnapshots creates a full 7-day realistic history of snapshots with mixed data.
func GenerateMockFullHistorySnapshots() []*Snapshot {
	// Diversified directory pools grouped by context
	baseDirs := []string{
		"/home/user/downloads",
		"/home/user/desktop",
		"/home/user/documents/receipts",
		"/home/user/pictures/screenshots",
	}

	projects := [][]string{
		// Microservices API
		{
			"/home/user/src/api-gateway",
			"/home/user/src/api-gateway/internal/router",
			"/home/user/src/api-gateway/configs/prod",
		},

		// Corporate Website Project
		{
			"/home/user/src/corporate-web",
			"/home/user/src/corporate-web/components/ui",
			"/home/user/src/corporate-web/public/assets",
			"/home/user/src/corporate-web/pages/api",
		},

		// Personal Tech Blog (Markdown & Static Hugo site generator)
		{
			"/home/user/personal/blog",
			"/home/user/personal/blog/content/posts",
			"/home/user/personal/blog/themes/minimal",
		},

		// Infrastructure as Code
		{
			"/home/user/src/infra-terraform",
			"/home/user/src/infra-terraform/environments/dev",
			"/home/user/src/infra-terraform/modules/vpc",
		},
	}

	systemDirs := []string{
		"/var/log/nginx",
		"/tmp/build-cache",
		"/etc/docker",
		"/home/user/.config/nvim",
		"/etc/hosts",
	}

	var snapshots []*Snapshot
	rng := rand.New(rand.NewSource(42)) // Use a custom Seed to get a pseudo-random generation but still deterministic

	// Start history exactly 7 days ago
	//baseTime := time.Now().Add(time.Duration(-7*24) * time.Hour)
	baseTime := time.Now().Round(0).AddDate(0, 0, -7)

	// Establish an initial fallback state before the loop starts
	currentActive := []string{"/home/user/src/corporate-web", "/home/user/downloads"}

	// Loop through every single day of the week
	for day := 0; day < 7; day++ {

		// Iterate over 24 hours in 5-minute steps (Continuous timeline)
		for hour := 0; hour < 24; hour++ {

			for minute := 0; minute < 60; minute += 5 {
				snapshotTime := baseTime.
					AddDate(0, 0, day).
					Add(time.Duration(hour) * time.Hour).
					Add(time.Duration(minute) * time.Minute)

				weekday := snapshotTime.Weekday()
				isWeekend := weekday == time.Saturday || weekday == time.Sunday
				isWorkingHours := snapshotTime.Hour() >= 9 && snapshotTime.Hour() < 18

				// Only change state if it's a weekday during [9 AM, 6 PM[
				if !isWeekend && isWorkingHours {
					if hour == 9 && minute == 0 {
						// Morning routine: Shift focus to a new project for the day
						activeProject := projects[rng.Intn(len(projects))]
						currentActive = []string{baseDirs[rng.Intn(len(baseDirs))], activeProject[0]}
					} else {
						// Create micro-adjustments during the workday
						activeProject := projects[rng.Intn(len(projects))] // Can pivot or stay
						currentActive = evolveDirectories(currentActive, baseDirs, activeProject, systemDirs, rng)
					}
				}

				// If it is night or weekend, currentActive remains exactly as it was last set creating identical (repeating) snapshot entries.

				//snapshotTime := time.Date(
				//	dayTime.Year(), dayTime.Month(), dayTime.Day(),
				//	hour, minute, 0, 0, dayTime.Location(),
				//)

				// Map the currentActive string slice to Directory structs
				var dirs []Directory
				for _, p := range currentActive {
					dirs = append(dirs, Directory{Id: InvalidId, Path: p})
				}

				// Create a snapshot
				snapshots = append(snapshots, &Snapshot{
					Id:          InvalidId,
					Timestamp:   snapshotTime,
					Directories: dirs,
				})
			}
		}
	}

	return snapshots
}

// evolveDirectories applies realistic micro-changes to open windows
func evolveDirectories(current []string, base, proj, sys []string, rng *rand.Rand) []string {
	dirMap := make(map[string]struct{})
	for _, p := range current {
		dirMap[p] = struct{}{}
	}

	// 85% chance the user stays focused on current directories (no change)
	if rng.Float64() < 0.85 {
		return current
	}

	// 10% chance to open something new
	if rng.Float64() < 0.70 && len(dirMap) < 6 {
		roll := rng.Float64()
		if roll < 0.50 {
			dirMap[proj[rng.Intn(len(proj))]] = struct{}{} // Drill deeper into active context
		} else if roll < 0.80 {
			dirMap[base[rng.Intn(len(base))]] = struct{}{} // Open a utility folder
		} else {
			dirMap[sys[rng.Intn(len(sys))]] = struct{}{} // Check system/editor configuration
		}
	}

	// 5% chance to close a directory window (keep at least 1 open)
	if rng.Float64() < 0.30 && len(dirMap) > 1 {
		keys := make([]string, 0, len(dirMap))
		for k := range dirMap {
			keys = append(keys, k)
		}
		delete(dirMap, keys[rng.Intn(len(keys))])
	}

	result := make([]string, 0, len(dirMap))
	for k := range dirMap {
		result = append(result, k)
	}
	return result
}

func TestCreateTables(t *testing.T) {
	var err error
	dbConn, err = sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()
}

func TestBasic(t *testing.T) {
	dbFilePath := createTempDatabaseFileName(t)

	var err error
	dbConn, err = sql.Open("sqlite", dbFilePath)
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	{
		err = CreateTables(dbConn)
		require.NoError(t, err)
	}

	snapshot, err := CreateSnapshotNow()

	// SaveSnapshot
	{
		require.NoError(t, err)

		err = SaveSnapshot(dbConn, &snapshot)
		require.NoError(t, err)

		// assert the snapshot has its id set
		require.NotEqual(t, InvalidId, snapshot.Id)
	}

	// FindSnapshotId
	{
		existingId, err := FindSnapshotId(dbConn, snapshot.Timestamp)
		require.NoError(t, err)
		require.NotEqual(t, InvalidId, existingId)
	}

	// LoadSnapshot
	{
		tmp, err := LoadSnapshot(dbConn, 1)
		require.NoError(t, err)
		require.Equal(t, 1, tmp.Id)
	}

	// Create 3 more snapshots in the database
	{
		snapshot1, err := CreateSnapshotNow()
		snapshot2, err := CreateSnapshotNow()
		snapshot3, err := CreateSnapshotNow()
		err = SaveSnapshot(dbConn, &snapshot1)
		require.NoError(t, err)
		err = SaveSnapshot(dbConn, &snapshot2)
		require.NoError(t, err)
		err = SaveSnapshot(dbConn, &snapshot3)
		require.NoError(t, err)
	}

	// GetLatestSnapshotId, GetLatestSnapshot
	{
		id, err := GetLatestSnapshotId(dbConn)
		require.NoError(t, err)
		require.Equal(t, 4, id)

		latestSnapshot, err := GetLatestSnapshot(dbConn)
		require.NoError(t, err)
		require.LessOrEqual(t, snapshot.Timestamp, latestSnapshot.Timestamp)
	}
}

func TestGetLastNSnapshots(t *testing.T) {
	dbFilePath := createTempDatabaseFileName(t)

	var err error
	dbConn, err = sql.Open("sqlite", dbFilePath)
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	{
		err = CreateTables(dbConn)
		require.NoError(t, err)
	}

	// Create mock snapshots
	mockSnapshots := GenerateMockSnapshots()
	err = InsertSnapshots(dbConn, mockSnapshots)
	require.NoError(t, err)

	// Ask too any snapshots
	snapshots, err := GetLastNSnapshots(dbConn, 70)
	require.NoError(t, err)
	require.Len(t, snapshots, len(mockSnapshots)) // all snapshots are returned

	// Ask for the last 7
	snapshots, err = GetLastNSnapshots(dbConn, 7)
	require.NoError(t, err)
	require.Len(t, snapshots, 7)
}

func TestDeleteSnapshotsFunctions(t *testing.T) {
	dbFilePath := createTempDatabaseFileName(t)

	var err error
	dbConn, err = sql.Open("sqlite", dbFilePath)
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	{
		err = CreateTables(dbConn)
		require.NoError(t, err)
	}

	// Create mock snapshots
	mockSnapshots := GenerateMockSnapshots()
	err = InsertSnapshots(dbConn, mockSnapshots)
	require.NoError(t, err)

	// DeleteSnapshotsInInterval()
	{
		// Delete snapshots [7, 8], leaving 0, 1 and 9
		err = DeleteSnapshotsInInterval(dbConn, mockSnapshots[2].Timestamp, mockSnapshots[8].Timestamp)
		require.NoError(t, err)

		// Get all remaining snapshots
		snapshots, err := GetLastNSnapshots(dbConn, 999)
		require.NoError(t, err)
		require.Len(t, snapshots, 3)
	}

	// Delete snapshot 1, leaving 0 and 9
	{
		err = DeleteSnapshotById(dbConn, mockSnapshots[1].Id)
		require.NoError(t, err)

		// Get all remaining snapshots
		snapshots, err := GetLastNSnapshots(dbConn, 999)
		require.NoError(t, err)
		require.Len(t, snapshots, 2)
	}
}

func TestDeleteSnapshotsOlderThanDays(t *testing.T) {
	dbFilePath := createTempDatabaseFileName(t)

	var err error
	dbConn, err = sql.Open("sqlite", dbFilePath)
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	{
		err = CreateTables(dbConn)
		require.NoError(t, err)
	}

	// Create mock snapshots
	mockSnapshots := GenerateMockSnapshots()
	err = InsertSnapshots(dbConn, mockSnapshots)
	require.NoError(t, err)

	// DeleteSnapshotsOlderThanDays()
	{
		// Delete snapshots [1, 5], leaving [6,10]
		err = DeleteSnapshotsOlderThanDays(dbConn, 3)
		require.NoError(t, err)

		// Get all remaining snapshots
		snapshots, err := GetLastNSnapshots(dbConn, 999)
		require.NoError(t, err)
		require.Len(t, snapshots, 5)
	}
}

func TestGetSnapshotsInInterval(t *testing.T) {
	dbFilePath := createTempDatabaseFileName(t)

	var err error
	dbConn, err = sql.Open("sqlite", dbFilePath)
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	{
		err = CreateTables(dbConn)
		require.NoError(t, err)
	}

	// Create mock snapshots
	mockSnapshots := GenerateMockSnapshots()
	err = InsertSnapshots(dbConn, mockSnapshots)
	require.NoError(t, err)

	// GetSnapshotsInInterval()
	{
		// Delete snapshots [3, 6]
		matches, err := GetSnapshotsInInterval(dbConn, mockSnapshots[3].Timestamp, mockSnapshots[6].Timestamp)
		require.NoError(t, err)
		require.Len(t, matches, 4)
	}
}

func TestExportSnapshotsToJson(t *testing.T) {
	dbFilePath := createTempDatabaseFileName(t)

	var err error
	dbConn, err = sql.Open("sqlite", dbFilePath)
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	{
		err = CreateTables(dbConn)
		require.NoError(t, err)
	}

	// Create mock snapshots
	mockSnapshots := GenerateMockSnapshots()
	err = InsertSnapshots(dbConn, mockSnapshots)
	require.NoError(t, err)

	// ExportSnapshotsToJson()
	{
		jsonFilePath := dbFilePath + ".json"
		err := ExportSnapshotsToJson(dbConn, jsonFilePath)
		require.NoError(t, err)
	}
}

func TestGetSnapshotRecallCandidates(t *testing.T) {
	dbFilePath := createTempDatabaseFileName(t)

	var err error
	dbConn, err = sql.Open("sqlite", dbFilePath)
	require.NoError(t, err)
	require.NotNil(t, dbConn)
	defer dbConn.Close()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	{
		err = CreateTables(dbConn)
		require.NoError(t, err)
	}

	// Create mock snapshots
	mockSnapshots := GenerateMockFullHistorySnapshots()
	err = InsertSnapshots(dbConn, mockSnapshots)
	require.NoError(t, err)
	count, err := GetSnapshotsCount(dbConn)
	require.Equal(t, len(mockSnapshots), count)

	// GetSnapshotRecallCandidates()
	{
		candidates, err := GetSnapshotRecallCandidates(dbConn)
		require.NoError(t, err)
		require.NotNil(t, candidates)

		debug := true
		if debug {
			jsonFilePath := dbFilePath + ".json"
			err := ExportSnapshotsToJson(dbConn, jsonFilePath)
			require.NoError(t, err)
			fmt.Printf("Snapshots exported to file '%s'\n", jsonFilePath)

			fmt.Printf("Daily snapshots:\n")
			for i, snapshot := range candidates.daily {
				fmt.Printf("%2d: %s\n", i, snapshot.String())
			}
			fmt.Printf("\n")

			fmt.Printf("Hourly snapshots:\n")
			for i, snapshot := range candidates.hourly {
				fmt.Printf("%2d: %s\n", i, snapshot.String())
			}
			fmt.Printf("\n")

			fmt.Printf("Latest snapshots:\n")
			for i, snapshot := range candidates.latest {
				fmt.Printf("%2d: %s\n", i, snapshot.String())
			}
			fmt.Printf("\n")
		}
	}
}
