# Smart File Pairing

VLC Sync Play now supports automatic file pairing based on a naming convention!

## Naming Convention

Files should be named: **`[basename]-[variant].[extension]`**

Where `[variant]` is either:
- `conductor` - for the conductor's version with visual cues
- `audience` - for the audience version without extra information

### Examples

- `beethoven-symphony-conductor.mp4` + `beethoven-symphony-audience.mp4`
- `silent-movie-conductor.mov` + `silent-movie-audience.mov`
- `orchestra-performance-conductor.mkv` + `orchestra-performance-audience.mkv`

## Usage

### Option 1: Specify Both Files (Explicit)
```bash
./vlc-sync-play movie-conductor.mp4 movie-audience.mp4
```

### Option 2: Specify One File (Auto-Detection) ✨
```bash
# Specify the conductor version - audience version is found automatically
./vlc-sync-play movie-conductor.mp4

# Or specify the audience version - conductor version is found automatically
./vlc-sync-play movie-audience.mp4

# Or use the base name - both versions are found automatically
./vlc-sync-play movie.mp4
```

If a matching pair is found, you'll see:
```
Auto-detected file pair:
  Conductor: /path/to/movie-conductor.mp4
  Audience:  /path/to/movie-audience.mp4
```

## Benefits

- **Convenience**: No need to type both file paths
- **Less Error-Prone**: Ensures correct pairing
- **Flexible**: Works from any version or the base name
- **Automatic**: Just drag one file to the app

## Fallback Behavior

If a matching file isn't found, the app will:
- Use the file you specified for the first instance
- Fall back to same file for second instance (original behavior)

## File Organization

Recommended directory structure:
```
performances/
  ├── beethoven-symphony-conductor.mp4
  ├── beethoven-symphony-audience.mp4
  ├── mozart-requiem-conductor.mp4
  ├── mozart-requiem-audience.mp4
  └── ...
```

## Implementation Details

The pairing logic:
1. Parses the input filename
2. Detects if it contains `-conductor` or `-audience`
3. Constructs the complementary filename
4. Checks if the complementary file exists
5. If found, uses both files; otherwise uses original input

Files must:
- Be in the same directory
- Have the same file extension
- Follow the naming convention exactly
