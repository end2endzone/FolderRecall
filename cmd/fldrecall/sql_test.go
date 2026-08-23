package main

import (
	"database/sql"
	"testing"

	ole "github.com/go-ole/go-ole"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

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
