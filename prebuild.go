// Trigger this execution of this file when go generate
//go:generate go run versioninfo.go --input=versioninfo-template.json --output=versioninfo.json

// Do not build this file when building main application
//go:build ignore
// +build ignore

package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed VERSION
var VersionFileContent string

// VersionDetail holds the 4-part numeric version integers
type VersionDetail struct {
	Major int `json:"Major"`
	Minor int `json:"Minor"`
	Patch int `json:"Patch"`
	Build int `json:"Build"`
}

func (v *VersionDetail) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.Build)
}

// FixedFileInfo holds core metadata used by the Windows OS
type FixedFileInfo struct {
	FileVersion    VersionDetail `json:"FileVersion"`
	ProductVersion VersionDetail `json:"ProductVersion"`
	FileFlagsMask  string        `json:"FileFlagsMask"`
	FileFlags      string        `json:"FileFlags"`
	FileOS         string        `json:"FileOS"`
	FileType       string        `json:"FileType"`
	FileSubType    string        `json:"FileSubType"`
}

// StringFileInfo holds descriptive text strings for properties windows
type StringFileInfo struct {
	Comments         string `json:"Comments"`
	CompanyName      string `json:"CompanyName"`
	FileDescription  string `json:"FileDescription"`
	FileVersion      string `json:"FileVersion"`
	InternalName     string `json:"InternalName"`
	LegalCopyright   string `json:"LegalCopyright"`
	LegalTrademarks  string `json:"LegalTrademarks"`
	OriginalFilename string `json:"OriginalFilename"`
	PrivateBuild     string `json:"PrivateBuild"`
	ProductName      string `json:"ProductName"`
	ProductVersion   string `json:"ProductVersion"`
	SpecialBuild     string `json:"SpecialBuild"`
}

// Translation details language and charset encodings
type Translation struct {
	LangID    string `json:"LangID"`
	CharsetID string `json:"CharsetID"`
}

// VarFileInfo holds internationalization blocks
type VarFileInfo struct {
	Translation Translation `json:"Translation"`
}

// VersionInfo matches your template file blueprint exactly
type VersionInfo struct {
	FixedFileInfo  FixedFileInfo  `json:"FixedFileInfo"`
	StringFileInfo StringFileInfo `json:"StringFileInfo"`
	VarFileInfo    VarFileInfo    `json:"VarFileInfo"`
	IconPath       string         `json:"IconPath"`
	ManifestPath   string         `json:"ManifestPath"`
}

// MapSlice maps a slice of type T to a slice of type U with the help of a conversion function
func MapSlice[T any, U any](input []T, f func(T) (U, error)) ([]U, error) {
	result := make([]U, len(input))
	for i, v := range input {
		res, err := f(v)
		if err != nil {
			return nil, err
		}
		result[i] = res
	}
	return result, nil
}

func GetVersionDetailFromVersionFile() VersionDetail {
	fields := strings.Split(VersionFileContent, ".")

	// Make the version a 4 digits version
	for len(fields) < 4 {
		fields = append(fields, "0")
	}

	// convert []string to []int
	digits, err := MapSlice(fields, strconv.Atoi)
	if err != nil {
		panic(err)
	}

	version := VersionDetail{
		Major: digits[0],
		Minor: digits[1],
		Patch: digits[2],
		Build: digits[3],
	}
	return version
}

func main() {
	// Get the input template and output files names
	inputPathStrPtr := flag.String("input", "", "The input template file path")
	outputPathStrPtr := flag.String("output", "", "The output file path to generate")

	// This parses os.Args[1:] behind the scenes
	flag.Parse()

	inputPath := *inputPathStrPtr
	outputPath := *outputPathStrPtr

	log.Printf("Parsing template %s into %s\n", inputPath, outputPath)

	// Read the template file bytes
	data, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Failed to read template: %v", err)
	}

	// Parse/unmarshal raw JSON into structs
	var info VersionInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	// Inject the content of file VERSION into different structures.
	// Example: info.StringFileInfo.FileVersion = "1.0.0.0"
	actualVersion := GetVersionDetailFromVersionFile()
	info.FixedFileInfo.FileVersion = actualVersion
	info.FixedFileInfo.ProductVersion = actualVersion
	info.StringFileInfo.FileVersion = actualVersion.String()
	info.StringFileInfo.ProductVersion = actualVersion.String()

	// Encode/marshal structs back into formatted JSON bytes
	outputData, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		log.Fatalf("Failed to serialize as JSON: %v", err)
	}

	// Save to versioninfo.json
	err = os.WriteFile(outputPath, outputData, 0644)
	if err != nil {
		log.Fatalf("Failed to save output file: %v", err)
	}

	outputFileName := filepath.Base(outputPath)
	log.Printf("Successfully updated %s to version %s.", outputFileName, actualVersion.String())
}
