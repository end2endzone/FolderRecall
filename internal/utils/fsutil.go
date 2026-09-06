package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MoveDir moves a directory from `src` to `dst`.
// It first attempts an os.Rename which is the fastest way when `src` and `dst` are on the same filesystem.
// It falls back to a recursive copy-then-remove when that fails.
func MoveDir(src string, dst string) error {
	// Try to rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fall back to a recursive copy-then-remove
	err = os.MkdirAll(filepath.Dir(dst), 0o755)
	if err != nil {
		return err
	}

	err = CopyDir(src, dst)
	if err != nil {
		return err
	}

	return os.RemoveAll(src)
}

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		err = os.MkdirAll(filepath.Dir(target), 0o755)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode()|0o600)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		return err
	})
}

// SanitizeCharactersInPath converts the given string into a filesystem-safe directory name or file name.
// Characters that are not supported by filesystems are replaeced by an underscore.
func SanitizeCharactersInPath(name string) string {
	// Remove non-friendly filesystem characters
	name = ReplaceInvalidFileSystemCharacters(name, "_")

	// Remove emojis, or unreadable characters
	name = SanitizeString(name, "_")

	// and trim
	name = strings.TrimSpace(name)

	return name
}

// CopyFile copies a a file from `src` to `dst`.
// It first attempts an os.Rename which is the fastest way when `src` and `dst` are on the same filesystem.
// It falls back to a recursive copy-then-remove when that fails.
func CopyFile(src string, dst string) error {
	dstDir := filepath.Dir(dst)

	// Create the directory tree (does nothing if it already exists)
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// FileExists checks if a file exists for the given path
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return true
	}
	return false
}

// DirExists checks if a file exists for the given path
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return true
	}
	return false
}

// FileExistsOrError checks if a file exists for the given path.
// Returns nil if the file exists, return an error otherwise.
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

// FileExistsOrError checks if a file exists for the given path.
// Returns nil if the file exists, return an error otherwise.
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

// GetFileSize returns the size of a file in bytes or an error.
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileSize returns the size of a file in bytes.
// Returns 0 if the file does not exists or there is an error.
func GetFileSizeSafe(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// GetDirectoryContentAbsolutePaths recursively walks through the given directory to find all the files under the given directory.
// Returns a list of all the files found in absolute paths.
func GetDirectoryContentAbsolutePaths(dirPath string) ([]string, error) {
	var fileNames []string

	// ReadDir reads the directory named by dirname and returns a list of directory entries sorted by filename.
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Construct the full path of the entry
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// If it's a directory, recursively call the function
			subDirFiles, err := GetDirectoryContentAbsolutePaths(fullPath)
			if err != nil {
				return nil, err
			}
			fileNames = append(fileNames, subDirFiles...)
		} else {
			// If it's a file, append its name to the list
			fileNames = append(fileNames, fullPath)
		}
	}

	return fileNames, nil
}
