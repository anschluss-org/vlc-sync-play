#!/bin/bash
set -e

echo "Building VLC Sync Play for macOS..."

# Create output directory
mkdir -p dist

# Build for macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -o dist/vlc-sync-play-darwin-arm64 ./cmd/trayagent
echo "✓ ARM64 binary created: dist/vlc-sync-play-darwin-arm64"

# Build for macOS AMD64 (Intel)
echo "Building for macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -o dist/vlc-sync-play-darwin-amd64 ./cmd/trayagent
echo "✓ AMD64 binary created: dist/vlc-sync-play-darwin-amd64"

# Make binaries executable
chmod +x dist/vlc-sync-play-darwin-arm64
chmod +x dist/vlc-sync-play-darwin-amd64

echo ""
echo "Build complete!"
echo "Binaries are in the dist/ directory:"
ls -lh dist/vlc-sync-play-darwin-*
