#!/bin/bash
set -euo pipefail

echo -e "\033[32mBuilding Zipic...\033[0m"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Build frontend
echo -e "\n\033[33m[1/2] Building frontend...\033[0m"
cd web
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    pnpm install
fi
pnpm build
cd ..

# Build backend
echo -e "\n\033[33m[2/2] Building backend...\033[0m"
cd backend
if [ ! -f "go.sum" ]; then
    echo "Downloading Go dependencies..."
    go mod download
fi

# Get version info
VERSION="${VERSION:-v1.0.0}"
BUILD_DATE=$(date +"%Y-%m-%d %H:%M:%S")
GIT_COMMIT="unknown"

if command -v git &> /dev/null; then
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    VERSION=$(git describe --tags --always 2>/dev/null || echo "v1.0.0")
fi

OUTPUT_DIR="bin"
mkdir -p "$OUTPUT_DIR"
OUTPUT="$OUTPUT_DIR/zipic"

go build -trimpath -ldflags="-s -w -X 'main.Version=$VERSION' -X 'main.BuildDate=$BUILD_DATE' -X 'main.GitCommit=$GIT_COMMIT'" -o "$OUTPUT" ./cmd/server

echo -e "\n\033[32mBuild completed successfully!\033[0m"
echo -e "\033[36mOutput: backend/$OUTPUT\033[0m"

cd ..
echo -e "\n\033[36mTo run the server: ./backend/bin/zipic\033[0m"