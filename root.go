package root

import (
	_ "embed"
)

//go:embed VERSION
var FileVersion string

func GetVersionFromVersionFile() string {
	return FileVersion
}
