#!/bin/bash

echo "Testing modified VLC Sync Play..."
echo ""

# Check if test files are provided
if [ $# -lt 2 ]; then
    echo "Usage: $0 <file1> <file2>"
    echo ""
    echo "Example:"
    echo "  $0 conductor-version.mp4 audience-version.mp4"
    echo ""
    echo "This will launch two VLC instances with synchronized playback"
    echo "of the two different files you specify."
    exit 1
fi

FILE1="$1"
FILE2="$2"

# Check if files exist
if [ ! -f "$FILE1" ]; then
    echo "Error: File not found: $FILE1"
    exit 1
fi

if [ ! -f "$FILE2" ]; then
    echo "Error: File not found: $FILE2"
    exit 1
fi

echo "File 1: $FILE1"
echo "File 2: $FILE2"
echo ""
echo "Launching VLC Sync Play..."
echo "- Instance 1 will play: $FILE1"
echo "- Instance 2 will play: $FILE2"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Run the application
./dist/vlc-sync-play-darwin-arm64 --debug "$FILE1" "$FILE2"
