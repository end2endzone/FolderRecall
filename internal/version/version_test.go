package version

import (
	"testing"

	root "github.com/end2endzone/FolderRecall"

	"github.com/stretchr/testify/require"
)

func TestGetProductVersionStringContainsVERSIONFile(t *testing.T) {
	productVersion := GetProductVersionString()
	versionFileContent := root.GetVersionFromVersionFile()

	// Assert the string returned from GetProductVersionString() contains the actual VERSION string.
	require.Contains(t, productVersion, versionFileContent, "GetProductVersionString() which is '%s' does not contains the VERSION which is '%s'. The VERSION file might be outdated.", productVersion, versionFileContent)
}
