# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**vlc-sync-play** is a cross-platform application that synchronizes playback across multiple VLC player instances. It enables users to watch the same video with different audio tracks in perfect sync, with all players responding to play/pause/seek commands simultaneously.

The application supports 2-4 synchronized VLC instances and runs as either a system tray application (primary use case) or CLI tool (debug mode).

## Build & Run Commands

### Building
```bash
# Install build tool and build for all platforms
./build.sh

# Generate embedded assets (icons, etc.)
make go_generate
```

The build uses `xgo-pack` for cross-platform compilation. Configuration is in `xgo-pack-config.yaml`. Output goes to `dist/` directory with platform-specific subdirectories.

### Running
```bash
# Run tray application (main entry point)
go run cmd/trayagent/main.go

# Run CLI application (debug/interactive mode)
go run cmd/cliagent/main.go

# Run with debug logging
go run cmd/trayagent/main.go --debug
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific test
go test ./pkg/util/time/ttl_counter_test.go
```

## Architecture Overview

### Entry Points
- **cmd/trayagent**: System tray/menu bar application (primary interface). Manages VLC instances through tray menu
- **cmd/cliagent**: CLI-based debug/interactive interface for development

### Core Components

**internal/app** - Application lifecycle and settings management
- `App`: Main application orchestrator, initializes settings and starts syncer
- `Settings`: Reactive settings using rx.Observable pattern (InstancesNumber, NoVideo, PollingInterval, ClickPause, ReSeekSrc)
- `SettingsStorage`: Persists settings to disk using `github.com/kirsle/configdir`

**pkg/vlc/syncer** - VLC synchronization engine
- `Syncer`: Main synchronization coordinator that polls VLC instances and issues sync commands
- `players`: Manages collection of VLC player instances
- `player`: Wraps individual VLC instance with polling client
- Uses leader/follower pattern: one player is "source of truth" and others follow its state

**pkg/vlc/client** - VLC HTTP API communication layer
- `basic/httpjson/BasicApiClient`: Low-level HTTP JSON API client for VLC (status.json, playlist.json endpoints)
- `extended/client`: Higher-level client with command queuing and repetition logic
- `ConnectionInfo`: Stores host, port, and random password for VLC HTTP interface

**pkg/vlc/instance** - VLC process management
- `Launcher`: Spawns VLC processes with correct HTTP interface flags
- `vlc_path`: Platform-specific VLC binary path detection (darwin/linux/windows)
- Parses stderr for events like mouse clicks (used for click-to-pause feature)

**pkg/vlc/state** - Player state representation
- `PlayerState`: Current playback state (playing/paused, position, duration, file)
- `ChangedProps`: Tracks which properties changed between updates
- `VlcPosition`: Timestamp + time pair for accurate position tracking

**internal/trayagent & internal/cli** - User interface implementations
- `trayagent`: System tray using `fyne.io/systray`, builds dynamic menu
- `cli/interactive`: Terminal UI using `github.com/rivo/tview` for debugging

**pkg/util** - Reusable utilities
- `rx`: Observable/reactive pattern implementation for settings
- `logging`: Simple logger interface with prefix support
- `time/ttl_counter`: TTL-based counter for timeout logic

### Key Design Patterns

**Reactive Settings**: Settings use `rx.Observable[T]` pattern. Changes trigger callbacks automatically (e.g., changing InstancesNumber launches new VLC instances)

**Polling-based Sync**: The syncer polls each VLC instance at `PollingInterval` (default 100ms), compares states, and issues commands to followers to match leader

**Platform-specific Code**: Build tags and separate files handle OS differences:
- VLC binary paths: `vlc_path/path_*.go`
- Static features: `static_features_windows.go` (click-to-pause only on Windows)
- Asset embedding: `embed_*.go`

**Context-driven Lifecycle**: All long-running operations use `context.Context` for cancellation. Settings changes propagate via goroutines with `errgroup.WithContext`

## VLC HTTP Interface

The app communicates with VLC via its built-in HTTP JSON API on localhost with random password. VLC is launched with flags like:
```
--http-host 127.0.0.1 --http-port <random> --http-password <random>
```

API endpoints used:
- `status.json` - Get current state (position, playing status, duration)
- `status.json?command=<cmd>` - Send commands (pl_play, pl_pause, seek)
- `playlist.json` - Get playlist information

## Platform Support

Targets: Windows (amd64), macOS (arm64/amd64), Linux (arm64/amd64)

Platform differences:
- macOS: Builds .app bundle and .dmg
- Linux: Builds .deb package with desktop entry
- Windows: Click-to-pause feature only available here (reads VLC stderr)

## Important Constraints

- Minimum 2 instances, maximum 4 (configurable via settings)
- Sync precision depends on polling interval (lower = more precise but higher CPU)
- VLC HTTP interface has rate limits (~50ms between commands)

## Modifications (February 2026)

**Modified Behavior**: The application now supports playing **different files** in each VLC instance while keeping them synchronized. This is useful for scenarios like:
- Conductor's version with visual cues + Audience version
- Different camera angles of the same performance

**Key Changes**:
- `Syncer` now tracks `filePaths []string` to map instance IDs to specific files
- `launchInstances()` assigns instance-specific files based on index (Instance N → filePaths[N-1])
- `onFileOpened()` opens the correct file for each instance using `getFileURIForInstance()`
- Synchronization still works because it's position-based, not file-based

**Usage**:
```bash
./vlc-sync-play conductor.mp4 audience.mp4  # Plays different files in sync
```

**Build Scripts**:
- `build-macos.sh` - Quick ARM64 build
- `build-macos-universal.sh` - Universal binary (ARM64 + Intel)

See `MODIFICATIONS.md` for complete details.
