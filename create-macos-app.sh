#!/bin/bash

set -e

APP_NAME="VLC Sync Play"
APP_BUNDLE="dist/${APP_NAME}.app"
BINARY_NAME="vlc-sync-play-universal"
VERSION="1.0.0"

echo "Creating macOS application bundle..."

# Clean up old bundle if it exists
if [ -d "$APP_BUNDLE" ]; then
    echo "Removing old app bundle..."
    rm -rf "$APP_BUNDLE"
fi

# Create app bundle structure
echo "Creating app bundle structure..."
mkdir -p "$APP_BUNDLE/Contents/MacOS"
mkdir -p "$APP_BUNDLE/Contents/Resources"

# Copy binary
echo "Copying binary..."
cp "dist/$BINARY_NAME" "$APP_BUNDLE/Contents/MacOS/vlc-sync-play"
chmod +x "$APP_BUNDLE/Contents/MacOS/vlc-sync-play"

# Create icon (convert PNG to icns)
echo "Creating application icon..."
ICON_SOURCE="assets/generated/tray_icon.png"

if [ -f "$ICON_SOURCE" ]; then
    # Create iconset directory
    ICONSET="$APP_BUNDLE/Contents/Resources/AppIcon.iconset"
    mkdir -p "$ICONSET"

    # Generate all required icon sizes
    sips -z 16 16     "$ICON_SOURCE" --out "$ICONSET/icon_16x16.png" 2>/dev/null
    sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET/icon_16x16@2x.png" 2>/dev/null
    sips -z 32 32     "$ICON_SOURCE" --out "$ICONSET/icon_32x32.png" 2>/dev/null
    sips -z 64 64     "$ICON_SOURCE" --out "$ICONSET/icon_32x32@2x.png" 2>/dev/null
    sips -z 128 128   "$ICON_SOURCE" --out "$ICONSET/icon_128x128.png" 2>/dev/null
    sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET/icon_128x128@2x.png" 2>/dev/null
    sips -z 256 256   "$ICON_SOURCE" --out "$ICONSET/icon_256x256.png" 2>/dev/null
    sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET/icon_256x256@2x.png" 2>/dev/null
    sips -z 512 512   "$ICON_SOURCE" --out "$ICONSET/icon_512x512.png" 2>/dev/null
    sips -z 1024 1024 "$ICON_SOURCE" --out "$ICONSET/icon_512x512@2x.png" 2>/dev/null

    # Convert iconset to icns
    iconutil -c icns "$ICONSET" -o "$APP_BUNDLE/Contents/Resources/AppIcon.icns"

    # Clean up iconset directory
    rm -rf "$ICONSET"

    echo "✓ Icon created"
else
    echo "⚠ Icon source not found at $ICON_SOURCE, skipping icon creation"
fi

# Create Info.plist
echo "Creating Info.plist..."
cat > "$APP_BUNDLE/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>vlc-sync-play</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.github.cardinalby.vlc-sync-play</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <false/>
    <key>CFBundleDocumentTypes</key>
    <array>
        <dict>
            <key>CFBundleTypeExtensions</key>
            <array>
                <string>mp4</string>
                <string>mov</string>
                <string>avi</string>
                <string>mkv</string>
                <string>m4v</string>
            </array>
            <key>CFBundleTypeName</key>
            <string>Video File</string>
            <key>CFBundleTypeRole</key>
            <string>Viewer</string>
            <key>LSHandlerRank</key>
            <string>Alternate</string>
        </dict>
    </array>
</dict>
</plist>
EOF

echo "✓ Info.plist created"

# Create PkgInfo
echo "APPL????" > "$APP_BUNDLE/Contents/PkgInfo"

echo ""
echo "=========================================="
echo "✅ macOS Application Bundle Created!"
echo "=========================================="
echo ""
echo "Location: $APP_BUNDLE"
echo ""
echo "To use:"
echo "  1. Double-click to launch (opens both VLC instances)"
echo "  2. Or drag a video file onto the app icon"
echo "  3. Or right-click a video → Open With → ${APP_NAME}"
echo ""
echo "To install:"
echo "  mv \"$APP_BUNDLE\" /Applications/"
echo ""
