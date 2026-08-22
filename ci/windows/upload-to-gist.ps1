$ErrorActionPreference = "Stop"

# Assert required variables are defined
$requiredVars = @("GH_TOKEN", "GIST_ID", "FILE_PATH", "GIST_FILENAME")

foreach ($var in $requiredVars) {
    $val = [Environment]::GetEnvironmentVariable($var)
    if ([string]::IsNullOrWhiteSpace($val)) {
        Write-Error "Error: Mandatory environment variable '$var' is not set."
        exit 1
    }
}

# Assert local file actually exists before trying to upload
if (-not (Test-Path -Path $env:FILE_PATH -PathType Leaf)) {
    Write-Error "Error: Local file path '$env:FILE_PATH' does not exist."
    exit 1
}

# Upload (create new file or override existing)
Write-Host "Checking Gist ($env:GIST_ID) for file: $env:GIST_FILENAME..."

# Retrieve files list from the Gist using gh CLI
$existingFiles = gh gist view $env:GIST_ID --files

# Check if the filename already exists in the Gist
if ($existingFiles -contains $env:GIST_FILENAME) {
    Write-Host "File '$env:GIST_FILENAME' exists in Gist. Overwriting..."
    gh gist edit $env:GIST_ID $env:FILE_PATH --filename $env:GIST_FILENAME
} else {
    Write-Host "File '$env:GIST_FILENAME' does not exist in Gist. Adding..."
    gh gist edit $env:GIST_ID $env:FILE_PATH --add $env:GIST_FILENAME
}

Write-Host "Successfully updated Gist!"