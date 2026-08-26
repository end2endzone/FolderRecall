package tests

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func AssertFileExists(t *testing.T, path string) {
	require.NotNil(t, path)

	info, err := os.Stat(path)
	require.NoError(t, err, "assertion failed: path does not exists: %v", path)
	require.False(t, info.IsDir(), "assertion failed: path is a directory: %v", path)
}

func AssertFileNotExists(t *testing.T, path string) {
	require.NotNil(t, path)

	_, err := os.Stat(path)
	require.Error(t, err, "assertion failed: expected an error for path: %v", path)
	require.True(t, os.IsNotExist(err))
}

func AssertDirExists(t *testing.T, path string) {
	require.NotNil(t, path)

	info, err := os.Stat(path)
	require.NoError(t, err, "assertion failed: path does not exists: %v", path)
	require.True(t, info.IsDir(), "assertion failed: path is not a directory: %v", path)
}

func AssertDirNotExists(t *testing.T, path string) {
	require.NotNil(t, path)

	_, err := os.Stat(path)
	require.Error(t, err, "assertion failed: expected an error for path: %v", path)
	require.True(t, os.IsNotExist(err))
}

func IsFileEquals(file1 *os.File, file2 *os.File) bool {
	const chunkSize = 65535
	for {
		b1 := make([]byte, chunkSize)
		_, err1 := file1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := file2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				// EOF reached for both files at the same time
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				// One of the files reached EOF but not the other
				return false
			} else {
				// Unknown error
				return false
			}
		}

		if !bytes.Equal(b1, b2) {
			// buffers are not equal
			return false
		}
	}

	// should not reach this code
	// return false
}

// GetFileSize returns the size of a file in bytes.
// Returns 0 if the file does not exists or there is an error.
func getFileSizeSafe(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func AssertFileEquals(t *testing.T, file1, file2 string) {
	// Do basic tests
	AssertFileExists(t, file1)
	AssertFileExists(t, file2)

	// Assert same size
	size1 := getFileSizeSafe(file1)
	size2 := getFileSizeSafe(file2)
	require.Equal(t, size1, size2, "assertion files: the size of the files is not equal")

	// Open the input files
	f1, err := os.Open(file1)
	require.NotNil(t, f1, "failed to open file: %v", file1)
	require.NoError(t, err)
	defer f1.Close()

	f2, err := os.Open(file2)
	require.NotNil(t, f2, "failed to open file: %v", file2)
	require.NoError(t, err)
	defer f2.Close()

	// Compare the content
	require.True(t, IsFileEquals(f1, f2), "assertion failed: the content of the files are not identical (%v,%v)", file1, file2)
}

func AssertFileContains(t *testing.T, path string, value string) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file %q", path)

	found := strings.Contains(string(data), value)
	require.True(t, found, "file %q does not contain value `%q`", path, value)
}
