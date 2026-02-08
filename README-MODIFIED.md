# VLC Sync Play - Modified Version

## ✅ Modifications Complete

Your modified version of VLC Sync Play has been successfully built! It now supports **synchronized playback of different files** in separate VLC instances.

## 🎯 What's New

**Original Behavior**: All VLC instances play the same file with different audio tracks

**Modified Behavior**: Each VLC instance can play a different file, but they stay perfectly synchronized
- Instance 1: `conductor-version.mp4` (with visual cues)
- Instance 2: `audience-version.mp4` (clean version)
- Playback, pause, and seeking are synchronized across all instances

Perfect for your orchestra performances! 🎻🎬

## 🚀 Quick Start

### 1. Locate Your Binary

Your compiled universal binary is ready:
```
dist/vlc-sync-play-universal
```

This works on both **Apple Silicon (M1/M2/M3)** and **Intel Macs**.

### 2. Basic Usage

```bash
# Run with two different files
./dist/vlc-sync-play-universal conductor-version.mp4 audience-version.mp4
```

This will:
1. Launch VLC instance 1 with `conductor-version.mp4`
2. Launch VLC instance 2 with `audience-version.mp4`
3. Keep them perfectly synchronized

### 3. Recommended Settings

```bash
# Disable video in second instance (audio/cues only)
./dist/vlc-sync-play-universal --no-video conductor.mp4 audience.mp4

# More precise sync (uses more CPU)
./dist/vlc-sync-play-universal --interval 50 conductor.mp4 audience.mp4

# Debug mode to see what's happening
./dist/vlc-sync-play-universal --debug conductor.mp4 audience.mp4
```

## 📋 Available Options

```bash
--instances N          Number of VLC instances (default: 2, max: 4)
--interval MS          Polling interval in ms (default: 100)
                       Lower = more precise but higher CPU
--no-video            Start additional instances without video
--re-seek-src         Re-seek source for precise sync (default: true)
--debug               Enable debug logging
--vlc PATH            Custom VLC executable path
```

## 🔧 Testing

A test script is provided:

```bash
./test-modified.sh your-conductor-file.mp4 your-audience-file.mp4
```

This runs the app in debug mode so you can see the synchronization happening.

## 📁 Files in This Repository

### Binaries
- `dist/vlc-sync-play-universal` - **Main binary** (both ARM64 & Intel)
- `dist/vlc-sync-play-arm64` - Apple Silicon only
- `dist/vlc-sync-play-amd64` - Intel only

### Documentation
- `USAGE.md` - Detailed usage guide
- `MODIFICATIONS.md` - Technical details of changes
- `CLAUDE.md` - Architecture documentation (updated)
- `README-MODIFIED.md` - This file

### Build Scripts
- `build-macos-universal.sh` - **Build universal binary**
- `build-macos.sh` - Build ARM64 only
- `test-modified.sh` - Test with your files

## 🎬 Typical Workflow for Orchestra Performances

1. **Prepare Your Files**
   - Export conductor's version with cues as `conductor.mp4`
   - Export audience version without cues as `audience.mp4`
   - Ensure both have similar durations

2. **Launch Sync Play**
   ```bash
   ./dist/vlc-sync-play-universal conductor.mp4 audience.mp4
   ```

3. **Configure VLC Instances**
   - Set audio output devices for each VLC window
   - Adjust volume levels as needed
   - Position windows on your displays

4. **Control Playback**
   - Use tray icon menu or any VLC window to control
   - Play/pause/seek in one window affects all windows
   - Scrubbing timeline keeps everything in sync

## 🎛️ System Tray Menu

The app adds a tray/menu bar icon with options:
- **VLC Instances** - Change number of instances
- **No video** - Toggle video in additional instances
- **Re-seek source** - Toggle precise sync algorithm
- **Polling interval** - Adjust sync precision
- **Quit** - Stop all instances

Settings persist between sessions in `~/.config/vlc-sync-play/`

## 🔍 How Synchronization Works

1. App polls each VLC instance every 100ms (configurable)
2. Detects state changes (play, pause, seek)
3. Propagates commands to all instances
4. Keeps position synchronized based on timestamps
5. Works with different files because sync is position-based, not content-based

## ⚠️ Important Notes

- **File Duration**: Files should have similar durations for best results
- **Local Files**: Files must be on local storage (not network/streaming)
- **VLC Required**: VLC Media Player must be installed
- **First Launch**: macOS may ask permission to run unsigned app (System Settings → Security)

## 🐛 Troubleshooting

### "VLC not found"
Specify VLC path manually:
```bash
./dist/vlc-sync-play-universal --vlc /Applications/VLC.app/Contents/MacOS/VLC file1.mp4 file2.mp4
```

### Instances out of sync
- Reduce polling interval: `--interval 50`
- Enable re-seek: `--re-seek-src`
- Ensure files are on fast local storage

### App won't open
Allow in System Settings:
1. System Settings → Privacy & Security
2. Find message about blocked app
3. Click "Open Anyway"

### Different file lengths
The app doesn't handle duration mismatches automatically. If files have different lengths:
- Use files with same duration when possible
- Sync will work but may drift at the end
- Consider trimming files to match

## 🔨 Rebuilding

If you need to rebuild after changes:

```bash
# Rebuild universal binary
./build-macos-universal.sh

# Quick rebuild (ARM64 only)
go build -o dist/vlc-sync-play-arm64 ./cmd/trayagent
```

## 📚 Technical Details

See `MODIFICATIONS.md` for:
- Detailed code changes
- Implementation notes
- Architecture decisions
- Testing checklist

## ✨ Success!

You're all set! Your modified VLC Sync Play is ready for:
- ✅ Synchronized playback of different files
- ✅ Works on Apple Silicon Macs (your machine)
- ✅ Works on Intel Macs (universal binary)
- ✅ Perfect for conductor + audience video scenarios

Enjoy your synchronized orchestra performances! 🎼

---

**Build Date**: February 8, 2026
**Modified By**: Claude (Anthropic)
**Original Project**: https://github.com/cardinalby/vlc-sync-play
