package main

import (
	"database/sql"
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
