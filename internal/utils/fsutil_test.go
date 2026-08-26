package utils

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/end2endzone/FolderRecall/internal/tests"
	"github.com/stretchr/testify/require"
)

// createMockDirectory creates a temporary directory with sample
// subdirectories and files. The caller should call the returned cleanup
// function when finished.
func createMockDirectory(t *testing.T) string {
	t.Helper()

	// Create a temporary parent directory.
	rootDir := filepath.Join(t.TempDir(), "mock-user-home-dir")

	// Define the sample directory structure and file contents.
	files := map[string]string{
		filepath.Join("documents", "notes.txt"):       "These are some notes.\n",
		filepath.Join("documents", "todo.txt"):        "1. Learn Python\n2. Write tests\n",
		filepath.Join("pictures", "readme.txt"):       "Picture files would go here.\n",
		filepath.Join("projects", "app", "main.py"):   "name = input(\"What is your name? \")\nprint(f\"Hello, {name}!\")\n",
		filepath.Join("projects", "app", "README.md"): "# Sample Python App\n",
	}

	// Create directories and files.
	for relativePath, content := range files {
		fullPath := filepath.Join(rootDir, relativePath)

		err := os.MkdirAll(filepath.Dir(fullPath), 0o755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0o644)
		require.NoError(t, err)
	}

	return rootDir
}

func TestMoveDir(t *testing.T) {
	dir := createMockDirectory(t)
	tests.AssertDirExists(t, dir)

	//act
	src := dir
	dst := t.TempDir()
	require.NoError(t, MoveDir(src, dst))

	// assert
	tests.AssertDirNotExists(t, src)

	// Get the list of moved files
	actualMovedFiles, err := GetDirectoryContentAbsolutePaths(dst)
	require.NoError(t, err)
	sort.Strings(actualMovedFiles)

	expectedRelativeFilesPath := []string{
		"documents/notes.txt",
		"documents/todo.txt",
		"pictures/readme.txt",
		"projects/app/main.py",
		"projects/app/README.md",
	}
	sort.Strings(expectedRelativeFilesPath)
	filePathsClean(expectedRelativeFilesPath)

	// Compare with the expected content
	require.Equal(t, len(actualMovedFiles), len(actualMovedFiles))
	for i, _ := range expectedRelativeFilesPath {
		expectedAbsPath := filepath.Join(dst, expectedRelativeFilesPath[i])
		actualAbsPath := actualMovedFiles[i]
		expectedAbsPath = filepath.Clean(expectedAbsPath) // normalize expected path to match the file separator on the system

		tests.AssertFileExists(t, expectedAbsPath)
		tests.AssertFileExists(t, actualAbsPath)
		require.Equal(t, expectedAbsPath, actualAbsPath)
	}
}

func TestCopyFile(t *testing.T) {
	dir := createMockDirectory(t)
	src := filepath.Join(dir, "documents", "notes.txt")
	dst := filepath.Join(t.TempDir(), filepath.Base(src)) // put the file directly in TempDir.

	// act
	require.NoError(t, CopyFile(src, dst))

	// assert
	tests.AssertFileEquals(t, src, dst)

	// The original file must still exist after a copy (unlike moveDir).
	tests.AssertFileExists(t, src)
}

func filePathsClean(paths []string) {
	for i, path := range paths {
		paths[i] = filepath.Clean(path)
	}
}
