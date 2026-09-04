package recall

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/stretchr/testify/require"
)

func makeSnapshot(dirs ...string) *Snapshot {
	s := &Snapshot{
		Id:        InvalidId,
		Timestamp: time.Now().Round(0),
	}

	for _, path := range dirs {
		s.AddDirectory(path)
	}
	return s
}

func ToStringSlice(s *Snapshot) []string {
	var paths []string
	for _, dir := range s.Directories {
		paths = append(paths, dir.Path)
	}
	return paths
}

func TestSort(t *testing.T) {
	now := time.Now().Round(0)
	s := Snapshot{
		Id:        InvalidId,
		Timestamp: now,
	}

	s.AddDirectory("/tmp")
	s.AddDirectory("/home")
	s.AddDirectory("/var")
	s.AddDirectory("/media")
	s.AddDirectory("/boot")

	s.Sort()

	actual := ToStringSlice(&s)

	// assert
	expected := []string{
		"/boot",
		"/home",
		"/media",
		"/tmp",
		"/var"}

	// Assert element in the list
	require.ElementsMatch(t, expected, actual)
}

func TestUnique(t *testing.T) {
	now := time.Now().Round(0)
	s := Snapshot{
		Id:        InvalidId,
		Timestamp: now,
	}

	s.AddDirectory("/tmp")
	s.AddDirectory("/home")
	s.AddDirectory("/var")
	s.AddDirectory("/home")
	s.AddDirectory("/media")
	s.AddDirectory("/boot")

	s.Unique()

	actual := ToStringSlice(&s)

	// assert
	expected := []string{
		"/tmp",
		"/home",
		"/var",
		"/media",
		"/boot"}

	// Assert both list are equals
	require.ElementsMatch(t, expected, actual)
}

func PathToURL(path string) string {
	path = filepath.Clean(path)

	// Convert backslashes to forward slashes: "C:/Windows/System32/Drivers"
	slashedPath := filepath.ToSlash(path)

	// Create a URL structure
	url := &url.URL{
		Scheme: "file",
		Path:   slashedPath,
	}

	// Output the standardized string
	str := url.String()
	return str
}

// closeFileExplorerWindowForDirectory loops through all File Explorer windows,
// identify the window that is currently opened to the given path directory,
// and closes it.
func closeFileExplorerWindowForDirectory(path string) error {
	// Build a target URL format used by Windows COM (case-insensitive check)
	// Function PathToURL() will return "file://C:/Windows" while Windows COM uses the format
	// "file:///C:/Windows".
	// Another option would be to use the code:
	// ```
	// "file:///" + strings.ReplaceAll(filepath.Clean(path), "\\", "/")
	// ``` but it would not support directory with spaces or special characters.
	targetURL := strings.ReplaceAll(PathToURL(path), "file://", "file:///")

	// Create the Shell.Application object
	unknown, err := oleutil.CreateObject("Shell.Application")
	if err != nil {
		return fmt.Errorf("failed to create Shell.Application: %v", err)
	}
	defer unknown.Release()

	// Get the IDispatch interface
	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("failed to querying IDispatch interface: %v", err)
	}
	defer shell.Release()

	// Call the Windows() method to fetch open File Explorer/IE windows
	windowsResult, err := oleutil.CallMethod(shell, "Windows")
	if err != nil {
		return fmt.Errorf("failed to fetching Windows collection: %v", err)
	}
	windows := windowsResult.ToIDispatch()
	defer windows.Release()

	// Get the count of open windows
	countResult, err := oleutil.GetProperty(windows, "Count")
	if err != nil {
		return fmt.Errorf("failed getting windows count: %v", err)
	}
	count := int(countResult.Val)

	fmt.Printf("Scanning %d open window(s)...\n", count)

	// Iterate through each window to find the target directory
	for i := 0; i < count; i++ {
		itemResult, err := oleutil.CallMethod(windows, "Item", i)
		if err != nil {
			continue
		}
		window := itemResult.ToIDispatch()

		// Fetch the LocationURL property of the window
		urlResult, err := oleutil.GetProperty(window, "LocationURL")
		if err == nil {
			currentURL := urlResult.ToString()

			fmt.Printf("Found window URL: %s\n", currentURL)

			// Check if this window matches our target path
			if strings.EqualFold(currentURL, targetURL) {
				fmt.Printf("Found target window: %s. Closing it...\n", currentURL)

				// Invoke Quit() on the specific window
				oleutil.CallMethod(window, "Quit")
				window.Release()
				return nil
			}
		}
		window.Release()
	}

	return fmt.Errorf("target File Explorer window URL not found: %s", targetURL)
}

func TestRestore(t *testing.T) {
	//Lock this goroutine to the current OS thread so that COM initializations (which are thread-bound) do not change.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM library
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	now := time.Now().Round(0)
	s := Snapshot{
		Id:        InvalidId,
		Timestamp: now,
	}

	path := "C:\\Windows"
	s.AddDirectory(path)

	err := s.Restore()

	// Assert File Explorer directories are now restored
	require.NoError(t, err)

	// Cleanup
	time.Sleep(5 * time.Second)
	err = closeFileExplorerWindowForDirectory(path)
	require.NoError(t, err)
}

func TestString(t *testing.T) {
	s := Snapshot{
		Id:        InvalidId,
		Timestamp: time.Now().AddDate(0, 0, -3), // 3 days ago
	}

	s.AddDirectory("/tmp")
	s.AddDirectory("/home")
	s.AddDirectory("/var")
	s.AddDirectory("/home")
	s.AddDirectory("/media")
	s.AddDirectory("/boot")

	actual := s.String()

	// Assert
	timeOffsetPos := strings.Index(actual, "(3 days ago)")
	directoryCountPos := strings.Index(actual, "6 directories")

	// Assert both substrings were found
	require.NotEqual(t, -1, timeOffsetPos)
	require.NotEqual(t, -1, directoryCountPos)

	// Assert offset is before directory count
	require.Less(t, timeOffsetPos, directoryCountPos)
}

func TestCompareDirectories(t *testing.T) {
	t.Run("same content", func(t *testing.T) {
		s1 := makeSnapshot("/tmp", "/home", "/var", "/media", "/boot")
		s2 := makeSnapshot("/tmp", "/home", "/var", "/media", "/boot")

		expectedAdded := []string{}
		expectedRemoved := []string{}

		diff := s1.CompareDirectories(s2)

		require.ElementsMatch(t, expectedAdded, diff.Added)
		require.ElementsMatch(t, expectedRemoved, diff.Removed)
	})

	t.Run("added 1 dir", func(t *testing.T) {
		s1 := makeSnapshot("/tmp", "/home", "/var", "/boot")
		s2 := makeSnapshot("/tmp", "/home", "/var", "/media", "/boot") // media was added

		expectedAdded := []string{"/media"}
		expectedRemoved := []string{}

		diff := s1.CompareDirectories(s2)

		require.ElementsMatch(t, expectedAdded, diff.Added)
		require.ElementsMatch(t, expectedRemoved, diff.Removed)
	})

	t.Run("removed 1 dir", func(t *testing.T) {
		s1 := makeSnapshot("/tmp", "/home", "/var", "/media", "/boot")
		s2 := makeSnapshot("/tmp", "/home", "/var", "/boot") // media was removed

		expectedAdded := []string{}
		expectedRemoved := []string{"/media"}

		diff := s1.CompareDirectories(s2)

		require.ElementsMatch(t, expectedAdded, diff.Added)
		require.ElementsMatch(t, expectedRemoved, diff.Removed)
	})

	t.Run("mix", func(t *testing.T) {
		s1 := makeSnapshot("/home", "/tmp", "/media")
		s2 := makeSnapshot("/var", "/home", "/tmp", "/boot") // media was removed, /var was added, /boot was added

		expectedAdded := []string{"/var", "/boot"}
		expectedRemoved := []string{"/media"}

		diff := s1.CompareDirectories(s2)

		require.ElementsMatch(t, expectedAdded, diff.Added)
		require.ElementsMatch(t, expectedRemoved, diff.Removed)
	})
}

func TestSimplifySnapshotsByDirectory(t *testing.T) {
	s0 := makeSnapshot()
	s1 := makeSnapshot("/tmp", "/home")
	s2 := makeSnapshot("/tmp", "/home")
	s3 := makeSnapshot("/tmp", "/home")
	s4 := makeSnapshot("/tmp", "/home")
	s5 := makeSnapshot("/tmp")
	s6 := makeSnapshot("/var", "/media", "/boot")
	s7 := makeSnapshot("/var", "/media", "/boot")
	s8 := makeSnapshot("/var", "/media", "/boot")
	s9 := makeSnapshot("/var", "/media", "/boot")

	snapshots := []*Snapshot{s0, s1, s2, s3, s4, s5, s6, s7, s8, s9}

	simplified := SimplifySnapshotsByDirectory(snapshots)

	require.Len(t, simplified, 4)

	require.Equal(t, simplified[0], s0)
	require.Equal(t, simplified[1], s4)
	require.Equal(t, simplified[2], s5)
	require.Equal(t, simplified[3], s9)
}
