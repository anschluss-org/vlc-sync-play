#!/bin/bash

set -e

APP_NAME="VLC Sync Play"
APP_BUNDLE="dist/${APP_NAME}.app"
DMG_NAME="VLC-Sync-Play"
VERSION="1.0.0"
DMG_TEMP="dist/dmg_temp"
DMG_FINAL="dist/${DMG_NAME}-${VERSION}.dmg"

echo "Creating DMG installer for ${APP_NAME}..."

# Check if app bundle exists
if [ ! -d "$APP_BUNDLE" ]; then
    echo "Error: App bundle not found at $APP_BUNDLE"
    echo "Please run ./create-macos-app.sh first"
    exit 1
fi

# Clean up old files
echo "Cleaning up old files..."
rm -rf "$DMG_TEMP"
rm -f "$DMG_FINAL"

# Create temporary DMG directory
echo "Creating temporary directory structure..."
mkdir -p "$DMG_TEMP"

# Copy app bundle to temp directory
echo "Copying app bundle..."
cp -R "$APP_BUNDLE" "$DMG_TEMP/"

# Create symbolic link to Applications folder
echo "Creating Applications folder link..."
ln -s /Applications "$DMG_TEMP/Applications"

# Create a README file
echo "Creating README..."
cat > "$DMG_TEMP/README.txt" <<EOF
VLC Sync Play - Synchronized Multi-Instance Video Playback
===========================================================

Version: ${VERSION}
Modified for different files per instance

Installation:
  Drag "VLC Sync Play.app" to the Applications folder

Usage:

1. WITHOUT FILES (Manual Selection):
   - Double-click the app
   - In the first VLC window, open a video file
   - The app automatically finds and opens the paired file
   - Both videos remain paused - arrange windows as needed
   - Click play in either window to start synchronized playback

2. WITH FILES (Command Line):
   Open Terminal and run:
   /Applications/VLC\ Sync\ Play.app/Contents/MacOS/vlc-sync-play /path/to/video.mp4

3. DRAG & DROP:
   - Drag a video file onto the app icon
   - Auto-detects paired file based on naming convention

File Naming Convention:
  For automatic pairing, name your files:
    movie-conductor.mp4  +  movie-audience.mp4
    concert-conductor.mov  +  concert-audience.mov

Features:
  ✓ Synchronized playback across instances
  ✓ Different files per VLC window
  ✓ Auto-detection of file pairs
  ✓ Synchronized play/pause/seek
  ✓ No playlist windows
  ✓ Both instances start paused

Requirements:
  - macOS 10.13 or later
  - VLC Media Player installed

For more information:
  https://github.com/cardinalby/vlc-sync-play

Modified: February 2026
EOF

# Create compressed DMG directly
echo "Creating compressed DMG..."
hdiutil create -volname "${APP_NAME}" \
               -srcfolder "$DMG_TEMP" \
               -ov \
               -format UDZO \
               -imagekey zlib-level=9 \
               "$DMG_FINAL"

# Clean up
echo "Cleaning up temporary files..."
rm -rf "$DMG_TEMP"

# Get final DMG size
DMG_SIZE_FINAL=$(du -h "$DMG_FINAL" | cut -f1)

echo ""
echo "=========================================="
echo "✅ DMG Created Successfully!"
echo "=========================================="
echo ""
echo "Location: $DMG_FINAL"
echo "Size: $DMG_SIZE_FINAL"
echo ""
echo "The DMG contains:"
echo "  • VLC Sync Play.app"
echo "  • Applications folder link (for easy installation)"
echo "  • README.txt (usage instructions)"
echo ""
echo "To install on another Mac:"
echo "  1. Copy the DMG file to the target machine"
echo "  2. Double-click the DMG to mount it"
echo "  3. Drag 'VLC Sync Play.app' to the Applications folder"
echo "  4. Eject the DMG"
echo "  5. First launch: System Settings → Privacy & Security → 'Open Anyway'"
echo ""
