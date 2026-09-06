package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/end2endzone/FolderRecall/internal/recall"
	"github.com/end2endzone/FolderRecall/internal/utils"
	"github.com/go-ole/go-ole"
)

func FindProjectDir() (string, error) {
	// Note:
	// Can not base the resolution on the current executable.
	// Command `go run` resolve as the following: `C:\Users\%USERNAME%\AppData\Local\Temp\go-build119597315\b001`.

	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Assume current directory is the project's root directory
	{
		projectDir := currentDir
		sandboxDir := filepath.Join(projectDir, ".debug-sandbox")
		if utils.DirExists(sandboxDir) {
			// This directory exists.
			return projectDir, nil
		}
	}

	// Assume current directory is the advanced-debug-sandbox directory
	{
		advancedDir := currentDir
		projectDir := filepath.Dir(filepath.Dir(advancedDir))
		sandboxDir := filepath.Join(projectDir, ".debug-sandbox")
		if utils.DirExists(sandboxDir) {
			// This directory exists.
			return projectDir, nil
		}
	}

	// Assume current directory is .debug-sandbox directory
	{
		projectDir := filepath.Dir(currentDir)
		sandboxDir := filepath.Join(projectDir, ".debug-sandbox")
		if utils.DirExists(sandboxDir) {
			// This directory exists.
			return projectDir, nil
		}
	}

	return "", fmt.Errorf("can not find project directory")
}

func FindDebugSandboxDir() string {
	// Find the project's root directory
	projectDir, err := FindProjectDir()
	if err != nil {
		panic(err)
	}

	// Find sandbox directory
	sandboxDir := filepath.Join(projectDir, ".debug-sandbox")
	err = utils.DirExistsOrError(sandboxDir)
	if err != nil {
		panic(err)
	}

	return sandboxDir
}

func FindRecallDir() string {
	sandboxDir := FindDebugSandboxDir()

	testDataDir := filepath.Join(sandboxDir, "testdata")
	err := utils.DirExistsOrError(testDataDir)
	if err != nil {
		panic(err)
	}

	recallDir := filepath.Join(testDataDir, "recall")
	err = os.MkdirAll(recallDir, 0o755)
	if err != nil {
		panic(err)
	}

	return recallDir
}

func GetRecallTempDirectoryFromIndex(recallDir string, index int) string {
	subDirs := filepath.Join(recallDir, "subdirs")
	name := fmt.Sprintf("%03d", index)
	subDirPath := filepath.Join(subDirs, name)
	return subDirPath
}

func CreateRecallTestDirectories() error {
	recallDir := FindRecallDir()

	// Create 20 subdirectories
	for i := 0; i < 20; i++ {
		subDirPath := GetRecallTempDirectoryFromIndex(recallDir, i)
		err := os.MkdirAll(subDirPath, 0o755)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddSingleMockSnapshot(snapshots *[]*recall.Snapshot, age time.Duration, path string) {
	now := time.Now().Round(0)

	snapshot := &recall.Snapshot{
		Id:        recall.InvalidId,
		Timestamp: now.Add(-age),
		Directories: []recall.Directory{
			{
				Id:   recall.InvalidId,
				Path: path,
			},
		},
	}

	*snapshots = append(*snapshots, snapshot)
}

func CreateRecallMockSnapshots() ([]*recall.Snapshot, error) {
	recallDir := FindRecallDir()

	snapshots := []*recall.Snapshot{}

	const oneDay = time.Duration(24 * time.Hour)
	const oneWeek = time.Duration(7 * oneDay)

	// Time offsets to simulate multiple days and different times of day
	// (e.g., increments of several hours to spread across a 3-day window)
	timeOffsets := []time.Duration{
		3 * oneWeek,      // 3 weeks ago
		2 * oneWeek,      // 2 weeks ago
		8 * oneDay,       // 8 days ago
		7 * oneDay,       // 7 days ago
		6 * oneDay,       // 6 days ago
		5 * oneDay,       // 5 days ago
		4 * oneDay,       // 4 days ago
		3 * oneDay,       // 3 days ago
		2 * oneDay,       // 2 days ago
		1 * oneDay,       // 1 days ago
		6 * time.Hour,    // 6 hours ago
		5 * time.Hour,    // 5 hours ago
		4 * time.Hour,    // 4 hours ago
		1 * time.Hour,    // 1 hours ago
		55 * time.Minute, // 55 minutes ago
		15 * time.Minute, // 15 minutes ago
		10 * time.Minute, // 10 minutes ago
		5 * time.Minute,  // 5 minutes ago
		0 * time.Hour,    // now
	}

	for i, timeOffset := range timeOffsets {
		path := GetRecallTempDirectoryFromIndex(recallDir, i)
		AddSingleMockSnapshot(&snapshots, timeOffset, path)
	}

	return snapshots, nil
}

func CreateRecallTestDatabase() error {
	recallDir := FindRecallDir()

	dbFilePath := filepath.Join(recallDir, "recall.db")

	// Delete the database if it alreadu exists
	if utils.FileExists(dbFilePath) {
		err := os.Remove(dbFilePath)
		if err != nil {
			panic(err)
		}
	}

	var dbConn *sql.DB
	dbConn, err := sql.Open("sqlite", dbFilePath)
	defer func() {
		if dbConn != nil {
			dbConn.Close()
		}
	}()

	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	// CreateTables
	err = recall.CreateTables(dbConn)
	if err != nil {
		return err
	}

	// Create mock snapshots
	snapshots, err := CreateRecallMockSnapshots()
	if err != nil {
		return err
	}

	// Insert into the database
	err = recall.InsertSnapshots(dbConn, snapshots)

	// Close and save the database
	err = dbConn.Close()
	if err != nil {
		return err
	}
	dbConn = nil

	fmt.Printf("Recall test database created: %s\n", dbFilePath)

	// Print snapshot details
	fmt.Printf("with the following snapshots:")
	for _, snapshot := range snapshots {
		fmt.Printf("%s\n", snapshot.String())
	}

	return nil
}

func main() {
	// Find the project's root directory
	projectDir, err := FindProjectDir()
	if err != nil {
		panic(err)
	}

	// Find sandbox directory
	sandboxDir := filepath.Join(projectDir, ".debug-sandbox")
	err = utils.DirExistsOrError(sandboxDir)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Building advanced debug sandbox in directory %s\n", sandboxDir)

	// Create more directories to use as snapshots
	err = CreateRecallTestDirectories()
	if err != nil {
		panic(err)
	}

	// Create fake database for recall
	err = CreateRecallTestDatabase()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Advanced sandbox ready!\n")
}
