#!/usr/bin/env pwsh
# Build script for Windows

Write-Host "Building Zipic..." -ForegroundColor Green

$ErrorActionPreference = "Stop"

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Build frontend
Write-Host "`n[1/2] Building frontend..." -ForegroundColor Yellow
Set-Location web
if (-not (Test-Path "node_modules")) {
    Write-Host "Installing frontend dependencies..."
    pnpm install
}
pnpm build
Set-Location ..

# Build backend
Write-Host "`n[2/2] Building backend..." -ForegroundColor Yellow
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

$Output = Join-Path $OutputDir "zipic.exe"

go build -trimpath -ldflags="-s -w -X 'main.Version=$Version' -X 'main.BuildDate=$BuildDate' -X 'main.GitCommit=$GitCommit'" -o $Output ./cmd/server

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nBuild completed successfully!" -ForegroundColor Green
    Write-Host "Output: backend\$Output" -ForegroundColor Cyan
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}

Set-Location ..
Write-Host "`nTo run the server: .\backend\bin\zipic.exe" -ForegroundColor Cyan