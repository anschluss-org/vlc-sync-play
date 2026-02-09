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
   - Open Terminal
   - Run: /Applications/VLC\ Sync\ Play.app/Contents/MacOS/vlc-sync-play /path/to/video.mp4
   - Or use the --debug flag for detailed logging

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

# Calculate size for DMG (app size + 50MB padding)
echo "Calculating DMG size..."
APP_SIZE=$(du -sm "$DMG_TEMP" | cut -f1)
DMG_SIZE=$((APP_SIZE + 50))

echo "Creating temporary DMG (${DMG_SIZE}MB)..."
hdiutil create -srcfolder "$DMG_TEMP" \
               -volname "${APP_NAME}" \
               -fs HFS+ \
               -fsargs "-c c=64,a=16,e=16" \
               -format UDRW \
               -size ${DMG_SIZE}m \
               "dist/temp.dmg"

# Mount the temporary DMG
echo "Mounting temporary DMG..."
MOUNT_OUTPUT=$(hdiutil attach -readwrite -noverify -noautoopen "dist/temp.dmg" 2>&1)
MOUNT_DIR=$(echo "$MOUNT_OUTPUT" | grep "/Volumes/" | awk '{print $3}')

if [ -z "$MOUNT_DIR" ]; then
    echo "Failed to mount DMG. Output:"
    echo "$MOUNT_OUTPUT"
    exit 1
fi

echo "Mounted at: $MOUNT_DIR"

# Set custom icon position (optional, requires Finder scripting)
echo "Configuring DMG appearance..."

# Set background and icon positions using AppleScript
osascript <<EOF 2>/dev/null || echo "Note: Could not set custom DMG layout (Finder may not be available)"
tell application "Finder"
    tell disk "${APP_NAME}"
        open
        set current view of container window to icon view
        set toolbar visible of container window to false
        set statusbar visible of container window to false
        set the bounds of container window to {100, 100, 600, 450}
        set viewOptions to the icon view options of container window
        set arrangement of viewOptions to not arranged
        set icon size of viewOptions to 128
        delay 1
        set position of item "${APP_NAME}.app" of container window to {125, 150}
        set position of item "Applications" of container window to {375, 150}
        set position of item "README.txt" of container window to {250, 300}
        close
        open
        update without registering applications
        delay 2
    end tell
end tell
EOF

# Unmount the temporary DMG
echo "Unmounting temporary DMG..."
sync
sync
sleep 1
hdiutil detach "$MOUNT_DIR" -force || {
    echo "Warning: Could not unmount automatically. Trying alternative method..."
    diskutil unmount force "$MOUNT_DIR" || true
}

# Convert to compressed, read-only final DMG
echo "Creating final compressed DMG..."
hdiutil convert "dist/temp.dmg" \
                -format UDZO \
                -imagekey zlib-level=9 \
                -o "$DMG_FINAL"

# Clean up
echo "Cleaning up temporary files..."
rm -f "dist/temp.dmg"
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
echo "To distribute:"
echo "  1. Copy the DMG to a USB drive or cloud storage"
echo "  2. On the target Mac: Open the DMG"
echo "  3. Drag 'VLC Sync Play.app' to Applications"
echo "  4. Eject the DMG"
echo ""
echo "First launch will require:"
echo "  System Settings → Privacy & Security → 'Open Anyway'"
echo ""
