#!/usr/bin/env pwsh
# Build script for Windows
# Usage: ./build.ps1 [target]
#   target format: [steps][:platform]
#   steps: web, backend, web:backend (default: web:backend)
#   platform: win (default), linux
# Examples:
#   ./build.ps1              - build web + backend for windows
#   ./build.ps1 web          - build frontend only
#   ./build.ps1 backend      - build backend only for windows
#   ./build.ps1 web:backend  - build web + backend for windows
#   ./build.ps1 backend:linux - build backend only for linux
#   ./build.ps1 web:backend:linux - build web + backend for linux

param(
    [string]$Target = ""
)

function Show-Help {
    Write-Host @"
Build script for Zipic

Usage: ./build.ps1 <target>

Target format: [steps][:platform]

Steps (choose one or combine):
  web              - Build frontend only
  backend          - Build backend only
  web:backend      - Build frontend then backend

Platform (optional, default: win):
  :win             - Windows output
  :linux           - Linux output

Examples:
  ./build.ps1 web             - Build frontend only
  ./build.ps1 backend         - Build backend only for windows
  ./build.ps1 web:backend     - Build web + backend for windows
  ./build.ps1 backend:linux   - Build backend only for linux
  ./build.ps1 web:backend:linux - Build web + backend for linux
  ./build.ps1 -h              - Show this help
"@
}

# Show help if no target or -h/--help
if ($Target -eq "" -or $Target -eq "-h" -or $Target -eq "--help") {
    Show-Help
    exit 0
}

# Parse target
$Platform = "win"
$BuildWeb = $false
$BuildBackend = $false

$Parts = $Target -split ":"

foreach ($Part in $Parts) {
    switch ($Part) {
        "web" { $BuildWeb = $true }
        "backend" { $BuildBackend = $true }
        "win" { $Platform = "win" }
        "linux" { $Platform = "linux" }
        "" { } # skip empty parts
        default {
            Write-Host "Unknown target part: $Part" -ForegroundColor Red
            Show-Help
            exit 1
        }
    }
}

# Default: build both if nothing specified
if (-not $BuildWeb -and -not $BuildBackend) {
    $BuildWeb = $true
    $BuildBackend = $true
}

Write-Host "Building Zipic..." -ForegroundColor Green
Write-Host "Target: $Target (Platform: $Platform, Web: $BuildWeb, Backend: $BuildBackend)" -ForegroundColor Cyan

$ErrorActionPreference = "Stop"

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

$StepNum = 0
$TotalSteps = if ($BuildWeb -and $BuildBackend) { 2 } elseif ($BuildWeb -or $BuildBackend) { 1 } else { 0 }

# Build frontend
if ($BuildWeb) {
    $StepNum++
    Write-Host "`n[$StepNum/$TotalSteps] Building frontend..." -ForegroundColor Yellow
    Set-Location web
    if (-not (Test-Path "node_modules")) {
        Write-Host "Installing frontend dependencies..."
        pnpm install
    }
    pnpm build
    Set-Location ..
}

# Build backend
if ($BuildBackend) {
    $StepNum++
    Write-Host "`n[$StepNum/$TotalSteps] Building backend ($Platform)..." -ForegroundColor Yellow
    Set-Location backend
    if (-not (Test-Path "go.sum")) {
        Write-Host "Downloading Go dependencies..."
        go mod download
    }

    # Get version from git tag or use default
    $Version = "v1.0.0"
    $BuildDate = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
    $GitCommit = "unknown"

    try {
        $GitCommit = git rev-parse HEAD 2>$null
        if ($GitCommit) {
            $GitCommit = $GitCommit.Trim()
        }
        $TagVersion = git describe --tags --always 2>$null
        if ($TagVersion) {
            $Version = $TagVersion.Trim()
        }
    } catch {}

    $OutputDir = "bin"
    if (-not (Test-Path $OutputDir)) {
        New-Item -ItemType Directory -Path $OutputDir | Out-Null
    }

    # Set output filename based on platform
    if ($Platform -eq "linux") {
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

    $LdFlags = "-s -w -X 'zipic/internal/version.Version=$Version' -X 'zipic/internal/version.BuildDate=$BuildDate' -X 'zipic/internal/version.GitCommit=$GitCommit'"
    go build -trimpath -ldflags="$LdFlags" -o $Output ./cmd/server

    # Clear env vars
    $Env:GOOS = ""
    $Env:GOARCH = ""

    if ($LASTEXITCODE -eq 0) {
        Write-Host "`nBackend build completed!" -ForegroundColor Green
        Write-Host "Platform: $Platform ($GoOs/$GoArch)" -ForegroundColor Cyan
        Write-Host "Output: backend\$Output" -ForegroundColor Cyan
    } else {
        Write-Host "`nBackend build failed!" -ForegroundColor Red
        Set-Location ..
        exit 1
    }

    Set-Location ..
}

Write-Host "`nAll builds completed successfully!" -ForegroundColor Green

if ($BuildBackend) {
    if ($Platform -eq "win") {
        Write-Host "To run the server: .\backend\bin\zipic.exe" -ForegroundColor Cyan
    } else {
        Write-Host "To run on Linux: chmod +x backend/bin/zipic && ./backend/bin/zipic" -ForegroundColor Cyan
    }
}