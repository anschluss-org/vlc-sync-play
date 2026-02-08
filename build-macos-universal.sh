#!/bin/bash
set -e

echo "Building VLC Sync Play for macOS (Universal Binary)..."
echo ""

# Create output directory
mkdir -p dist

# Build for macOS ARM64 (Apple Silicon)
echo "Building for macOS ARM64..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
    go build -o dist/vlc-sync-play-arm64 ./cmd/trayagent
echo "✓ ARM64 binary created"

# Try to build for macOS AMD64 (Intel) with CGO
echo ""
echo "Building for macOS AMD64 (Intel)..."
echo "Note: This requires macOS SDK with Intel support"

# Try with CGO enabled and proper architecture
if CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
   go build -o dist/vlc-sync-play-amd64 ./cmd/trayagent 2>&1 | tee /tmp/build-intel.log; then
    echo "✓ AMD64 binary created"

    # Create universal binary
    echo ""
    echo "Creating universal binary..."
    lipo -create \
        dist/vlc-sync-play-arm64 \
        dist/vlc-sync-play-amd64 \
        -output dist/vlc-sync-play-universal

    chmod +x dist/vlc-sync-play-universal
    echo "✓ Universal binary created: dist/vlc-sync-play-universal"

    # Show file info
    echo ""
    echo "Binary architectures:"
    lipo -info dist/vlc-sync-play-universal
else
    echo ""
    echo "⚠ Intel (AMD64) build failed"
    echo "This is expected when cross-compiling CGO apps without proper toolchain"
    echo ""
    echo "Options for Intel Mac build:"
    echo "1. Build on an Intel Mac"
    echo "2. Use Docker with xgo-pack: docker run and xgo-pack build"
    echo "3. Use the ARM64 binary only (works on Apple Silicon Macs)"
    echo ""
    echo "For now, using ARM64 binary only..."
    cp dist/vlc-sync-play-arm64 dist/vlc-sync-play-darwin-arm64
fi

echo ""
echo "Build complete!"
echo "Available binaries:"
ls -lh dist/vlc-sync-play-* 2>/dev/null | grep -v ".dSYM"
