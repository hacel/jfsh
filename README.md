# jfsh

A terminal-based client for [Jellyfin](https://jellyfin.org) that lets you browse your media library and play videos via [mpv](https://mpv.io).
Inspired by [jftui](https://github.com/Aanok/jftui).

![Demo](demo/demo.gif)

## Features

- Uses _your_ mpv config!
- Resumes playback!
- Tracks playback progress and updates jellyfin!
- Automatic segment (intro, etc.) skipping!
- No mouse required!

## Installation

### Prerequisites

- A running [Jellyfin](https://jellyfin.org) instance.
- [mpv](https://mpv.io) available in PATH.

#### Install from AUR

On Arch Linux, jfsh can be installed from the [AUR](https://aur.archlinux.org/packages/jfsh)

```sh
yay -S jfsh
```

#### Download a release

Download the latest pre-built binary for your platform from the [releases page](https://github.com/hacel/jfsh/releases/latest).

#### Install via go

```sh
go install github.com/hacel/jfsh@latest
```

## Usage

1. **Start jfsh**

   ```sh
   jfsh
   ```

2. **Login**

   On first launch, you'll be prompted to enter:

   - **Host**: e.g., `http://localhost:8096`
   - **Username**
   - **Password**

3. **Play Media**

   - Select an item and press **Enter** or **Space** to play it.
   - `mpv` will launch and begin streaming.

4. **Quit**

   - Press **`q`** at any time to exit jfsh.

## Configuration

By default, the configuration file is stored in `$XDG_CONFIG_HOME/jfsh/jfsh.yaml`. If `$XDG_CONFIG_HOME` is not set it defaults to:

- **Linux**: `~/.config/jfsh/jfsh.yaml`
- **macOS**: `~/Library/Application Support/jfsh/jfsh.yaml`
- **Windows**: `%APPDATA%/jfsh/jfsh.yaml`

```yaml
host: http://localhost:8096
username: me
password: hunter2
device: mycomputer # Device name to report to jellyfin (default: hostname)
skip_segments: # Segments to automatically skip (default: [])
  - Recap
  - Preview
  - Intro
  - Outro
```

### Segment skipping

By default, no segments are automatically skipped. To enable skipping segments you must add `skip_segments` to the configuration file. Possible values for `skip_segments` are the segment types in Jellyfin which are: `Unknown`, `Commercial`, `Preview`, `Recap`, `Outro` and `Intro`.

## Plans

- Configuration through TUI
- Complete library browsing
- Sorting
- Better search: Filter by media type, watched status, and metadata
