# Run from the app/ directory
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Push-Location $PSScriptRoot

Write-Host "→ Stopping running instance..."
Get-Process -Name chekout -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 300

Write-Host "→ Building..."
go build -o chekout.exe .

Write-Host "→ Launching..."
Start-Process -FilePath ".\chekout.exe"
Write-Host "✓ Done"

Pop-Location
