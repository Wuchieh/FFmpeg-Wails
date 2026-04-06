# Changelog

## v0.0.2 (2026-04-06)

### 🔴 Bug Fixes

- **Fix concurrent runner bug**: Replaced single `runner *ffmpeg.Runner` with `runners map[string]*Runner` keyed by task ID. Previously, launching multiple tasks would overwrite the runner reference, causing `CancelTask` to cancel the wrong process. (#1)
- **Fix `-vf` filter conflict**: When both subtitle burn-in and resolution scaling were enabled, FFmpeg rejected the command because `-vf` was specified twice. Now both filters are merged into a single `-vf subtitles=...,scale=...` chain. (#4)
- **Fix `canceled`/`cancelled` typo**: Backend Go used `"canceled"` (US spelling) but frontend Vue matched `"cancelled"` (UK spelling), causing cancelled tasks to never show the yellow badge. Frontend now uses `"canceled"` to match. (#7)
- **Fix duplicate event listener binding**: `setupListeners()` was called in `onMounted` of every page (convert, stream, tasks), causing events to fire multiple times after page navigation. Moved to `app.vue` with a guard flag to bind only once. (#6)
- **Fix composable state not shared**: Each component calling `useFFmpeg()` got independent `tasks`/`logs`/`progress` refs, so task state was siloed per page. State is now module-level (singleton pattern), shared across all components. (#9)

### 🔄 Refactoring

- **Task ID generation**: Switched from `time.Now().UnixNano()` to `uuid.New().String()` to eliminate collision risk during rapid task creation. (#2)
- **Extract `runTask()` common logic**: `startConvertTask` and `startStreamTask` had ~80% duplicated code for runner initialization, event emission, and goroutine launch. Extracted into a shared `runTask(id, task, args, duration)` method. (#5)
- **Remove dead code**: Deleted unused `StartStream()` function in `ffmpeg/stream.go` — `app.go` was creating runners directly instead. (#12)

### 🛡️ Security & Safety

- **ExtraArgs positioning**: Moved `ExtraArgs` insertion to before `-y` (output overwrite flag), keeping user-supplied args in a more predictable position. (#3)
- **Removed placeholder email**: Cleared `wuchieh@example.com` from `wails.json`. (#11)

### 🎨 UI

- **FileSelector spacing**: Changed `gap-4` to `gap-2` in `FileSelector.vue` for consistent label-input spacing across all form fields. (#13)

### 📦 Build & CI

- **Synced `go.mod`**: Ran `go mod tidy` to ensure dependency format matches Go toolchain expectations, fixing CI lint failures.
- **Go version**: 1.25.8 (aligned across `go.mod`, `.golangci.yml`, and `release.yml`)

---

Full diff: https://github.com/Wuchieh/FFmpeg-Wails/compare/v0.0.1...v0.0.2
