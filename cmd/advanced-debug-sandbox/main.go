package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-ole/go-ole"
)

func IsDirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return true
	}
	return false
}

func IsFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return true
	}
	return false
}

func DirExistsOrError(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

func FileExistsOrError(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("path is not a file: %s", path)
	}
	return nil
}

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
		if IsDirExists(sandboxDir) {
			// This directory exists.
			return projectDir, nil
		}
	}

	// Assume current directory is the advanced-debug-sandbox directory
	{
		advancedDir := currentDir
		projectDir := filepath.Dir(filepath.Dir(advancedDir))
		sandboxDir := filepath.Join(projectDir, ".debug-sandbox")
		if IsDirExists(sandboxDir) {
			// This directory exists.
			return projectDir, nil
		}
	}

	// Assume current directory is .debug-sandbox directory
	{
		projectDir := filepath.Dir(currentDir)
		sandboxDir := filepath.Join(projectDir, ".debug-sandbox")
		if IsDirExists(sandboxDir) {
			// This directory exists.
			return projectDir, nil
		}
	}

	return "", fmt.Errorf("can not find project directory")
}

func GetTempDirectoryPath(sandboxDir string, index int) string {
	tempDir := filepath.Join(sandboxDir, "temp")
	name := fmt.Sprintf("%03d", index)
	subDirPath := filepath.Join(tempDir, name)
	return subDirPath
}

func CreateTempDirectories(sandboxDir string) error {
	// Create 20 subdirectories
	for i := 0; i < 20; i++ {
		subDirPath := GetTempDirectoryPath(sandboxDir, i)
		err := os.MkdirAll(subDirPath, 0o755)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateRecallTestDatabase(sandboxDir string) error {
	testDataDir := filepath.Join(sandboxDir, "testdata")
	err := DirExistsOrError(testDataDir)
	if err != nil {
		return err
	}

	recallDir := filepath.Join(testDataDir, "recall")
	err = os.MkdirAll(recallDir, 0o755)
	if err != nil {
		return err
	}

	dbFilePath := filepath.Join(recallDir, "recall.db")

	dbConn, err = sql.Open("sqlite", dbFilePath)
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
	err = CreateTables(dbConn)
	if err != nil {
		return err
	}

	// Create mock snapshots
	snapshot, err := CreateSnapshotNow()
	if err != nil {
		return err
	}
	snapshots := []*Snapshot{&snapshot}

	// Insert into the database
	err = InsertSnapshots(dbConn, snapshots)

	// Close and save the database
	err = dbConn.Close()
	if err != nil {
		return err
	}
	dbConn = nil

	fmt.Printf("Recall test database created: %s\n", dbFilePath)

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
	err = DirExistsOrError(sandboxDir)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Building advanced debug sandbox in directory %s\n", sandboxDir)

	// Create more directories to use as snapshots
	err = CreateTempDirectories(sandboxDir)
	if err != nil {
		panic(err)
	}

	// Create fake database for recall
	err = CreateRecallTestDatabase(sandboxDir)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Advanced sandbox ready!\n")
}
