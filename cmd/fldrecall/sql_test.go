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

	t.Run("create tables", func(t *testing.T) {
		err = CreateTables(dbConn)
		require.NoError(t, err)
	})

	t.Run("save snapshot", func(t *testing.T) {
		snapshot, err := GetCurrentDirectoriesSnapshot()
		require.NoError(t, err)

		err = SaveSnapshot(dbConn, &snapshot)
		require.NoError(t, err)
	})
}
