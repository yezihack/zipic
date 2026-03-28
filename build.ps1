#!/usr/bin/env pwsh
# Build script for Windows
# Usage: ./build.ps1 [-Target win|linux] [-SkipFrontend]

param(
    [ValidateSet("win", "linux")]
    [string]$Target = "win",
    [switch]$SkipFrontend
)

Write-Host "Building Zipic ($Target)..." -ForegroundColor Green

$ErrorActionPreference = "Stop"

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Build frontend
if (-not $SkipFrontend) {
    Write-Host "`n[1/2] Building frontend..." -ForegroundColor Yellow
    Set-Location web
    if (-not (Test-Path "node_modules")) {
        Write-Host "Installing frontend dependencies..."
        pnpm install
    }
    pnpm build
    Set-Location ..
} else {
    Write-Host "`n[1/2] Skipping frontend build" -ForegroundColor Yellow
}

# Build backend
Write-Host "`n[2/2] Building backend ($Target)..." -ForegroundColor Yellow
Set-Location backend
if (-not (Test-Path "go.sum")) {
    Write-Host "Downloading Go dependencies..."
    go mod download
}

# Get version from git tag or use default
$Version = "v1.0.0"
$BuildDate = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
$GitCommit = "unknown"

try {
    $GitCommit = git rev-parse --short HEAD 2>$null
    $Version = git describe --tags --always 2>$null
} catch {}

$OutputDir = "bin"
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

# Set output filename based on target
if ($Target -eq "linux") {
    $Output = Join-Path $OutputDir "zipic"
    $GoOs = "linux"
    $GoArch = "amd64"
} else {
    $Output = Join-Path $OutputDir "zipic.exe"
    $GoOs = "windows"
    $GoArch = "amd64"
}

$Env:GOOS = $GoOs
$Env:GOARCH = $GoArch

go build -trimpath -ldflags="-s -w -X 'main.Version=$Version' -X 'main.BuildDate=$BuildDate' -X 'main.GitCommit=$GitCommit'" -o $Output ./cmd/server

# Clear env vars
$Env:GOOS = ""
$Env:GOARCH = ""

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild completed successfully!" -ForegroundColor Green
    Write-Host "Target: $Target ($GoOs/$GoArch)" -ForegroundColor Cyan
    Write-Host "Output: backend\$Output" -ForegroundColor Cyan
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}

Set-Location ..

if ($Target -eq "win") {
    Write-Host "`nTo run the server: .\backend\bin\zipic.exe" -ForegroundColor Cyan
} else {
    Write-Host "`nTo run on Linux: chmod +x backend/bin/zipic && ./backend/bin/zipic" -ForegroundColor Cyan
}