package recall

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// testdataDir returns the absolute path to the testdata directory located at the project's root directory.
// It returns the directory location regardless of which package's test binary is running.
func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	// This file lives in <module>/minecraftbedrock, so testdata is a sibling of that directory.
	return filepath.Join(wd, "..", "testdata")
}

// createTempDatabaseFileName generate a file name in a temp directory.
// The temporary directory will be deleted at the end of the test session.
func createTempDatabaseFileName(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), t.Name()+".db")
	return path
}
