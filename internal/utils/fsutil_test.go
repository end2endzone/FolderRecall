package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func assertFileExists(t *testing.T, path string) {
	require.NotNil(t, path)

	info, err := os.Stat(path)
	require.NoError(t, err, "assertion failed: path does not exists: %v", path)
	require.False(t, info.IsDir(), "assertion failed: path is a directory: %v", path)
}

func assertFileNotExists(t *testing.T, path string) {
	require.NotNil(t, path)

	_, err := os.Stat(path)
	require.Error(t, err, "assertion failed: expected an error for path: %v", path)
	require.True(t, os.IsNotExist(err))
}

func assertDirExists(t *testing.T, path string) {
	require.NotNil(t, path)

	info, err := os.Stat(path)
	require.NoError(t, err, "assertion failed: path does not exists: %v", path)
	require.True(t, info.IsDir(), "assertion failed: path is not a directory: %v", path)
}

func assertDirNotExists(t *testing.T, path string) {
	require.NotNil(t, path)

	_, err := os.Stat(path)
	require.Error(t, err, "assertion failed: expected an error for path: %v", path)
	require.True(t, os.IsNotExist(err))
}

func isFileEquals(file1 *os.File, file2 *os.File) bool {
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

func assertFileEquals(t *testing.T, file1, file2 string) {
	// Do basic tests
	assertFileExists(t, file1)
	assertFileExists(t, file2)

	// Assert same size
	size1 := GetFileSizeSafe(file1)
	size2 := GetFileSizeSafe(file2)
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
	require.True(t, isFileEquals(f1, f2), "assertion failed: the content of the files are not identical (%v,%v)", file1, file2)
}

func TestSanitizeCharactersInPath(t *testing.T) {
	cases := map[string]string{
		"Foobar RP":            "Foobar RP",
		"":                     "",
		"   ":                  "",
		"Weird/Name:Here":      "Weird_Name_Here",
		"  Trim Me  ":          "Trim Me",
		"Path\\With*Bad?Chars": "Path_With_Bad_Chars",
	}
	for in, want := range cases {
		got := sanitizeCharactersInPath(in)
		require.Equal(t, want, got)
	}
}

func TestMoveDir(t *testing.T) {
	serverDir := copyServerFixture(t, "server_empty")
	assertDirExists(t, serverDir)

	//act
	src := serverDir
	dst := t.TempDir()
	require.NoError(t, moveDir(src, dst))

	// assert
	assertDirNotExists(t, src)

	// Get the list of moved files
	actualMovedFiles, err := GetDirectoryContentAbsolutePaths(dst)
	require.NoError(t, err)
	sort.Strings(actualMovedFiles)

	expectedRelativeFilesPath := []string{
		"bedrock_server",
		"bedrock_server.exe",
		"server.properties",
		"worlds/Bedrock level/.keep",
	}
	sort.Strings(expectedRelativeFilesPath)
	filePathsClean(expectedRelativeFilesPath)

	// Compare with the expected content
	require.Equal(t, len(actualMovedFiles), len(actualMovedFiles))
	for i, _ := range expectedRelativeFilesPath {
		expectedAbsPath := filepath.Join(dst, expectedRelativeFilesPath[i])
		actualAbsPath := actualMovedFiles[i]
		expectedAbsPath = filepath.Clean(expectedAbsPath) // normalize expected path to match the file separator on the system

		assertFileExists(t, expectedAbsPath)
		assertFileExists(t, actualAbsPath)
		require.Equal(t, expectedAbsPath, actualAbsPath)
	}
}

func TestCopyFile(t *testing.T) {
	src := getAddonFixturePath(t, "foobar.mcaddon")
	dst := filepath.Join(t.TempDir(), filepath.Base(src)) // put the file directly in TempDir.

	// act
	require.NoError(t, CopyFile(src, dst))

	// assert
	assertFileEquals(t, src, dst)

	// The original file must still exist after a copy (unlike moveDir).
	assertFileExists(t, src)
}

func filePathsClean(paths []string) {
	for i, path := range paths {
		paths[i] = filepath.Clean(path)
	}
}
