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

# Ensures your binary does not depend on host operating system C libraries, making the binary completely portable.
$env:CGO_ENABLED = "0"

# Read current version
$Version = Get-Content -Path "VERSION" -Raw

Write-Host ""

Write-Host "Generating..."
# Run generators recursively across your entire project.
# Generators must running in the local CPU architecture disregarding what ever GOARCH is set/override to.
# If we don't we risk trying to run an arm64 executable on a amd64 CPU.
# This would result in the following error:
# ```
# fork/exec C:\Users\%USERNAME%\AppData\Local\Temp\go-build3692463925\b001\exe\prebuild.exe:
# This version of %1 is not compatible with the version of Windows you're running.
# Check your computer's system information and then contact the software publisher.
$oldGOARCH = $Env:GOARCH
try {
    $Env:GOARCH = $(go env GOHOSTARCH);
    go generate ./...
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to run go generators: exit code $LASTEXITCODE"
    }
}
finally {
    $Env:GOARCH = $oldGOARCH
}
Write-Host "done"
Write-Host ""

########################
# advanced-debug-sandbox
########################

# Define the target binary file path based on the environment
$Target="$ProjectRoot\cmd\advanced-debug-sandbox\advanced-debug-sandbox.exe"
if ($env:CI -eq "true") {
    $Target="$ProjectRoot\cmd\advanced-debug-sandbox-$env:GOOS-$env:GOARCH.exe"
    Write-Host "Building on CI/CD server. Changing the target file name to '$Target'"
}

Write-Host "Building $(Split-Path -Leaf $Target) version $Version..."
# Run build from the root, pointing to the main package directory
go build -o $Target ./cmd/advanced-debug-sandbox
if ($LASTEXITCODE -ne 0) {
    throw "Failed to build go code: exit code $LASTEXITCODE"
}
Write-Host "done"
Write-Host ""

############
# fldrecall
############

# Define the target binary file path based on the environment
$Target="$ProjectRoot\bin\fldrecall.exe"
if ($env:CI -eq "true") {
    $Target="$ProjectRoot\bin\fldrecall-$env:GOOS-$env:GOARCH.exe"
    Write-Host "Building on CI/CD server. Changing the target file name to '$Target'"
}

Write-Host "Building $(Split-Path -Leaf $Target) version $Version..."
# Run build from the root, pointing to the main package directory
go build -o $Target ./cmd/fldrecall
if ($LASTEXITCODE -ne 0) {
    throw "Failed to build go code: exit code $LASTEXITCODE"
}
Write-Host "done"
Write-Host ""

Write-Host "Build complete!"
Write-Host

Pop-Location