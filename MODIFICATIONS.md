# VLC Sync Play Modifications

## Summary

This document describes the modifications made to vlc-sync-play to support synchronized playback of **different files** in separate VLC instances, rather than the same file in all instances.

## Use Case

- **Original**: Play the same file (e.g., `movie.mp4`) in multiple VLC instances with different audio tracks
- **Modified**: Play different related files (e.g., `conductor-version.mp4` and `audience-version.mp4`) in sync

Perfect for live orchestra performances where:
- Instance 1: Plays conductor's version with visual cues
- Instance 2: Plays audience version without extra on-screen information

## Technical Changes

### Files Modified

1. **pkg/vlc/syncer/syncer.go**
   - Added `filePaths []string` field to Syncer struct to track file paths per instance
   - Modified `NewSyncer()` to accept `filePaths` parameter
   - Modified `launchInstances()` to assign instance-specific files based on instance index
   - Modified `onFileOpened()` to open correct file for each instance
   - Added `getFileURIForInstance()` helper function to map instance IDs to file paths

2. **internal/app/app.go**
   - Updated `NewSyncer()` call to pass `settings.FilePaths`

### Key Implementation Details

**Instance ID to File Mapping**:
- Instance IDs start at 1 (assigned by launcher)
- Array indices start at 0
- Mapping: Instance ID N → filePaths[N-1]

**Fallback Behavior**:
- If fewer files than instances: additional instances use first file
- If no files specified: falls back to original behavior

**Synchronization**:
- Play/pause/seek commands synchronized across all instances
- Position sync based on relative timestamps (works even with different files)
- Each instance plays its assigned file, but at synchronized positions

## Building

### Quick Build (ARM64 only)
```bash
./build-macos.sh
```

### Universal Binary (ARM64 + Intel)
```bash
./build-macos-universal.sh
```

This creates:
- `dist/vlc-sync-play-arm64` - Apple Silicon only
- `dist/vlc-sync-play-amd64` - Intel only
- `dist/vlc-sync-play-universal` - Universal binary for both

### Generated Assets

The build requires generated icon assets. If building from scratch:

```bash
# If Docker is available:
make go_generate

# If Docker is not available (workaround):
cp assets/icons/template_tray_icon.png assets/generated/tray_icon.png
```

## Usage Examples

### Basic (2 files)
```bash
./dist/vlc-sync-play-universal conductor.mp4 audience.mp4
```

### With Options
```bash
# Disable video in second instance
./dist/vlc-sync-play-universal --no-video conductor.mp4 audience.mp4

# Custom polling interval (lower = more precise, higher CPU)
./dist/vlc-sync-play-universal --interval 50 conductor.mp4 audience.mp4

# Debug mode
./dist/vlc-sync-play-universal --debug conductor.mp4 audience.mp4

# Specify VLC path
./dist/vlc-sync-play-universal --vlc /Applications/VLC.app/Contents/MacOS/VLC file1.mp4 file2.mp4
```

### Testing
```bash
./test-modified.sh your-file1.mp4 your-file2.mp4
```

## Requirements

- macOS (tested on Apple Silicon, works on Intel via universal binary)
- VLC Media Player installed
- Go 1.21+ (for building)
- Files should have similar duration for best sync results

## How Synchronization Works

1. **Polling**: Each instance polled at interval (default 100ms)
2. **Leader Selection**: First instance with state change becomes temporary "leader"
3. **Command Propagation**: Play/pause/seek commands sent to all instances
4. **Position Sync**: All instances seek to same relative position
5. **File Independence**: Each instance plays its own file, sync is position-based

The sync algorithm doesn't care that files are different - it only cares about:
- Playback state (playing/paused)
- Relative position (timestamp)
- Seeking events

## Limitations

- Files should have similar durations
- Sync precision depends on polling interval
- Maximum 4 instances supported (original limitation)
- Files must be locally accessible (not streaming URLs)

## Testing Checklist

- [ ] Launch with 2 different files
- [ ] Verify both VLC instances open
- [ ] Test play/pause synchronization
- [ ] Test seeking/scrubbing synchronization
- [ ] Test with different file formats
- [ ] Test with files of slightly different durations
- [ ] Verify tray icon and menu work
- [ ] Test settings persistence

## Future Enhancements (Optional)

- Support for duration mismatches (scale playback speed)
- Per-instance volume control
- File selection via GUI
- Real-time file switching
- Support for more than 4 instances

## Rollback

To revert to original behavior, restore from git:
```bash
git checkout pkg/vlc/syncer/syncer.go internal/app/app.go
```
