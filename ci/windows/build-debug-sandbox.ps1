$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = Convert-Path("$PSScriptRoot/../..")

Set-Location $ProjectRoot

if (Test-Path .debug-sandbox) { Remove-Item -Recurse -Force .debug-sandbox }
New-Item -ItemType Directory -Path .debug-sandbox | Out-Null
Copy-Item -Recurse testdata .debug-sandbox/testdata

Write-Host "Sandbox ready!"
