# FFmpeg-Wails

A cross-platform FFmpeg GUI built with [Wails](https://wails.io/) (Go) + [Nuxt 4](https://nuxt.com/) (Vue 3).

Convert video/audio, stream to RTMP/SRT platforms, and manage tasks — all from a native desktop app with a modern dark UI.

![License](https://img.shields.io/badge/license-MIT-blue)

---

## Features

### 🎬 Convert

Video and audio conversion with full codec control:

| Option | Choices |
|--------|---------|
| Video Codec | H.264, H.265, VP9, AV1, Copy |
| Audio Codec | AAC, MP3, Opus, Copy |
| Resolution | Original, 1080p, 720p, 480p, 360p |
| CRF | 0–51 (lower = better quality) |
| FPS | Custom frame rate |
| Bitrate | Video & audio bitrate control |
| Subtitles | Burn-in from .srt / .ass files |
| Extra Args | Raw FFmpeg flags for advanced use |

Real-time progress bar with FPS, bitrate, speed, and time tracking.

### 📡 Stream

Push to live platforms via RTMP or SRT:

- **RTMP** — YouTube, Twitch, custom servers (with one-click URL presets)
- **SRT** — Low-latency transport with configurable latency
- Supports local files and live sources (RTSP/HTTP)

### 📋 Tasks

Unified task management panel:

- View all tasks (convert + stream) sorted by time
- Real-time status badges: running, completed, failed, cancelled
- Collapsible log viewer per task
- Cancel running tasks

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, Wails v2 |
| Frontend | Nuxt 4, Vue 3, UnoCSS |
| Engine | FFmpeg / ffprobe |
| CI/CD | GitHub Actions (test → build → release) |

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/) 22+
- [pnpm](https://pnpm.io/) 9+
- [FFmpeg](https://ffmpeg.org/) installed and in PATH

Install FFmpeg:

```bash
# macOS
brew install ffmpeg

# Ubuntu / Debian
sudo apt install ffmpeg

# Windows — download from https://ffmpeg.org/download.html
```

Verify:

```bash
ffmpeg -version
ffprobe -version
```

### Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Run in Development

```bash
git clone https://github.com/Wuchieh/FFmpeg-Wails.git
cd FFmpeg-Wails

# Install frontend dependencies
cd frontend && pnpm install && cd ..

# Start dev server (Go backend + Nuxt frontend with hot reload)
wails dev
```

### Build

```bash
wails build
```

Output binaries in `build/bin/`:

- **macOS** — `.app`
- **Windows** — `.exe`
- **Linux** — binary

---

## Project Structure

```
FFmpeg-Wails/
├── main.go                  # App entry, FFmpeg availability check
├── backend/
│   └── app.go               # Wails-bound App (tasks, file dialogs, events)
├── ffmpeg/
│   ├── command.go           # Build FFmpeg args (convert + stream)
│   ├── runner.go            # Execute FFmpeg, parse progress, cancel support
│   └── stream.go            # High-level stream launcher
├── frontend/
│   ├── app.vue              # Layout with navigation
│   ├── nuxt.config.ts       # Nuxt 4 + UnoCSS config
│   ├── pages/
│   │   ├── convert.vue      # Conversion UI
│   │   ├── stream.vue       # Streaming UI
│   │   └── tasks.vue        # Task management UI
│   ├── components/
│   │   ├── FileSelector.vue # Native file/directory picker
│   │   ├── ProgressBar.vue  # Animated progress bar
│   │   └── LogViewer.vue    # Scrollable log output
│   └── composables/
│       └── useFFmpeg.ts     # Wails bindings + event handling
├── .github/workflows/
│   └── release.yml          # CI: test → build → GitHub Release
├── wails.json
└── go.mod
```

---

## Architecture

### Backend (Go)

The `App` struct is bound to Wails and exposed to the frontend:

```go
// Key methods exposed to frontend:
app.StartTask(payloadJSON)    // Auto-detects convert vs stream
app.CancelTask(id)            // Kill running FFmpeg process
app.GetTaskStatus(id)         // Poll task state
app.ListTasks()               // List all tasks
app.SelectFile()              // Native file dialog
app.SelectDirectory()         // Native directory dialog
app.GetFFmpegVersion()        // Probe installed FFmpeg
```

FFmpeg commands are built by `ffmpeg.BuildConvertArgs()` / `ffmpeg.BuildStreamArgs()` and executed by `ffmpeg.Runner`. Progress is parsed from FFmpeg's stderr using regex (frame, fps, bitrate, time, speed) and emitted to the frontend via Wails events.

Duration probing via `ffprobe` enables accurate percentage calculation.

### Frontend (Vue 3 + Nuxt 4)

Three pages connected by `useFFmpeg()` composable:

- **`/convert`** — Form with codec/resolution/CRF controls, native file picker, live progress bar
- **`/stream`** — RTMP/SRT protocol toggle, platform presets (YouTube/Twitch), live source toggle
- **`/tasks`** — Sorted task list with status badges, collapsible logs, cancel buttons

Real-time updates via `window.runtime.EventsOn()` (Wails event system).

### Events

| Event | Data | Description |
|-------|------|-------------|
| `task:progress` | `{id, progress, fps, bitrate, time, frame, speed}` | Live progress update |
| `task:log` | `{id, line}` | FFmpeg stderr line |
| `task:done` | `{id, status, error}` | Task completed/failed/cancelled |

---

## Releasing

Pushing a tag triggers the full CI pipeline:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will:

1. **Test** — `go vet`, `golangci-lint`, `go test`
2. **Build** — Cross-compile for Windows, macOS (amd64 + arm64), Linux
3. **Release** — Create a GitHub Release with all binaries attached

---

## API Reference (Backend → Frontend)

### `StartTask(payloadJSON: string): Task`

Auto-detects task type from payload:

- Has `url` field → **stream** task
- Otherwise → **convert** task

**Convert payload:**

```json
{
  "input": "/path/to/input.mp4",
  "output": "/path/to/output.mp4",
  "videoCodec": "libx264",
  "crf": 23,
  "resolution": "1280:720",
  "fps": 30,
  "audioCodec": "aac",
  "audioBitrate": "128k",
  "subtitleFile": "/path/to/sub.srt",
  "extraArgs": "-vf eq=brightness=0.1"
}
```

**Stream payload:**

```json
{
  "source": "/path/to/input.mp4",
  "protocol": "rtmp",
  "url": "rtmp://a.rtmp.youtube.com/live2/STREAM-KEY",
  "videoCodec": "libx264",
  "preset": "veryfast",
  "bitrate": "3000k",
  "isLive": false
}
```

---

## License

[MIT](./LICENSE)

---

## Credits

- [FFmpeg](https://ffmpeg.org/) — The engine behind it all
- [Wails](https://wails.io/) — Go + Web frontend → native desktop
- [Nuxt](https://nuxt.com/) — Vue 3 meta-framework
