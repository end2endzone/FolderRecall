$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = Convert-Path("$PSScriptRoot/../..")

Push-Location
Set-Location $ProjectRoot

# GOOS
if (-not [string]::IsNullOrEmpty($env:GOOS)) {
    Write-Output "GOOS is set to: $env:GOOS"
} else {
    $env:GOOS = (go env GOHOSTOS).Trim()
    Write-Output "GOOS is not set and is forced to: $env:GOOS"
}

# GOARCH
if (-not [string]::IsNullOrEmpty($env:GOARCH)) {
    Write-Output "GOARCH is set to: $env:GOARCH"
} else {
    $env:GOARCH = (go env GOHOSTARCH).Trim()
    Write-Output "GOARCH is not set and is forced to: $env:GOARCH"
}

# Define the target binary file path based on the environment
$Target="$ProjectRoot\bin\fldrecall.exe"
if ($env:CI -eq "true") {
    $Target="$ProjectRoot\bin\fldrecall-$env:GOOS-$env:GOARCH.exe"
    echo "Building on CI/CD server. Changing the target file name to '$Target'"
}

# Ensures your binary does not depend on host operating system C libraries, making the binary completely portable.
$env:CGO_ENABLED = "0"

# Point this to the exact package path where your main function lives
$Pkg = "main"

Write-Host "Generating..."
go generate ./cmd/fldrecall

Write-Host "Building $(Split-Path -Leaf $Target) version $Version..."
# Run build from the root, pointing to the main package directory
go build -o $Target ./cmd/fldrecall

Write-Host "Build complete!"
Write-Host

Pop-Location