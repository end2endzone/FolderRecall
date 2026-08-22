$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = Convert-Path("$PSScriptRoot/../..")

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
}

# Show compiled version
& $Target --version --verbose
