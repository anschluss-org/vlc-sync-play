# Usage Guide for Modified VLC Sync Play

## What Changed

This modified version allows you to synchronize playback of **different but related files** in separate VLC instances. Perfect for scenarios like:
- Conductor's version with visual cues + Audience version without cues
- Different audio tracks or camera angles of the same performance
- Any synchronized multi-file playback scenario

## How to Use

### Basic Usage (2 files)

```bash
./vlc-sync-play conductor-version.mp4 audience-version.mp4
```

This will:
1. Launch first VLC instance with `conductor-version.mp4`
2. Launch second VLC instance with `audience-version.mp4`
3. Keep them synchronized

### With More Options

```bash
# Specify number of instances and custom settings
./vlc-sync-play --instances 2 --interval 100 conductor.mp4 audience.mp4

# Disable video in second instance
./vlc-sync-play --no-video conductor.mp4 audience.mp4

# Debug mode
./vlc-sync-play --debug conductor.mp4 audience.mp4
```

### Command Line Options

- `--instances N` - Number of VLC instances (default: 2)
- `--interval MS` - Polling interval in milliseconds (default: 100)
- `--no-video` - Start additional instances without video
- `--re-seek-src` - Re-seek source player for precise sync (default: true)
- `--click-pause` - Click to pause/resume (Windows only)
- `--debug` - Enable debug logging

### File Mapping

Files are mapped to instances in order:
- Instance 1 (ID 1) → First file argument
- Instance 2 (ID 2) → Second file argument
- Instance 3 (ID 3) → Third file argument (if specified)
- Instance 4 (ID 4) → Fourth file argument (if specified)

If you request more instances than files provided, additional instances will use the first file.

## Synchronization Behavior

- **Play/Pause**: Synchronized across all instances
- **Seek**: Synchronized to same relative position
- **File Opening**: When you open a file in one instance, corresponding files open in others
- **Scrubbing**: Real-time position sync while dragging

## System Requirements

- macOS (ARM64 - Apple Silicon)
- VLC Media Player installed
- Files should have similar duration for best results

## Troubleshooting

### VLC not found
If you get an error about VLC not being found, specify the path:
```bash
./vlc-sync-play --vlc /Applications/VLC.app/Contents/MacOS/VLC conductor.mp4 audience.mp4
```

### Out of sync
- Try adjusting the polling interval: `--interval 50` (lower = more precise, higher CPU)
- Ensure files have similar durations
- Check that files are on fast storage (not network drives)

### Settings Persistence
Settings are saved in `~/.config/vlc-sync-play/` and persist between sessions.
