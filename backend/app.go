package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"ffmpeg-wails/ffmpeg"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Task represents a single FFmpeg operation tracked by the application.
type Task struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Command   string    `json:"command"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`
	Input     string    `json:"input"`
	Output    string    `json:"output"`
	CreatedAt time.Time `json:"createdAt"`
	Error     string    `json:"error,omitempty"`
}

// ConvertPayload is the JSON payload sent from the frontend for conversion tasks.
type ConvertPayload struct {
	Input        string `json:"input"`
	Output       string `json:"output"`
	VideoCodec   string `json:"videoCodec,omitempty"`
	AudioCodec   string `json:"audioCodec,omitempty"`
	Resolution   string `json:"resolution,omitempty"`
	FPS          int    `json:"fps,omitempty"`
	CRF          int    `json:"crf,omitempty"`
	Bitrate      string `json:"bitrate,omitempty"`
	AudioBitrate string `json:"audioBitrate,omitempty"`
	SubtitleFile string `json:"subtitleFile,omitempty"`
	Format       string `json:"format,omitempty"`
	ExtraArgs    string `json:"extraArgs,omitempty"`
}

// StreamPayload is the JSON payload sent from the frontend for streaming tasks.
type StreamPayload struct {
	Source     string `json:"source"`
	Protocol   string `json:"protocol"`
	URL        string `json:"url"`
	VideoCodec string `json:"videoCodec,omitempty"`
	AudioCodec string `json:"audioCodec,omitempty"`
	Bitrate    string `json:"bitrate,omitempty"`
	Preset     string `json:"preset,omitempty"`
	Latency    int    `json:"latency,omitempty"`
	IsLive     bool   `json:"isLive,omitempty"`
}

// App is the main application struct exposed to the Wails frontend.
type App struct {
	ctx    context.Context
	mu     sync.RWMutex
	tasks  map[string]*Task
	runner *ffmpeg.Runner
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		tasks: make(map[string]*Task),
	}
}

// StartupCtx is called when the Wails application starts.
func (a *App) StartupCtx(ctx context.Context) {
	a.ctx = ctx
}

// GetFFmpegVersion returns the installed FFmpeg version string.
func (a *App) GetFFmpegVersion() string {
	version, err := ffmpeg.GetFFmpegVersion()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return version
}

// SelectFile opens a native file dialog and returns the selected file path.
func (a *App) SelectFile() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select File",
	})
	if err != nil {
		return ""
	}
	return path
}

// SelectDirectory opens a native directory dialog and returns the selected path.
func (a *App) SelectDirectory() string {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory",
	})
	if err != nil {
		return ""
	}
	return path
}

// StartTask begins a new FFmpeg task based on the provided JSON payload.
func (a *App) StartTask(payloadJSON string) (*Task, error) {
	var taskType string

	// Try to detect the task type from the payload
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payloadJSON), &raw); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}

	if _, hasURL := raw["url"]; hasURL {
		taskType = "stream"
	} else {
		taskType = "convert"
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	task := &Task{
		ID:        id,
		Type:      taskType,
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
	}

	a.mu.Lock()
	a.tasks[id] = task
	a.mu.Unlock()

	if taskType == "convert" {
		return a.startConvertTask(id, task, payloadJSON)
	}
	return a.startStreamTask(id, task, payloadJSON)
}

func (a *App) startConvertTask(id string, task *Task, payloadJSON string) (*Task, error) {
	var payload ConvertPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return nil, fmt.Errorf("invalid convert payload: %w", err)
	}

	args, err := ffmpeg.BuildConvertArgs(ffmpeg.ConvertOptions{
		Input:        payload.Input,
		Output:       payload.Output,
		VideoCodec:   payload.VideoCodec,
		AudioCodec:   payload.AudioCodec,
		Resolution:   payload.Resolution,
		FPS:          payload.FPS,
		CRF:          payload.CRF,
		Bitrate:      payload.Bitrate,
		AudioBitrate: payload.AudioBitrate,
		SubtitleFile: payload.SubtitleFile,
		Format:       payload.Format,
		ExtraArgs:    payload.ExtraArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build command: %w", err)
	}

	task.Input = payload.Input
	task.Output = payload.Output
	task.Command = fmt.Sprintf("ffmpeg %s", strings.Join(args, " "))
	task.Status = "running"

	// Probe duration for progress calculation
	duration, _ := ffmpeg.GetInputDuration(payload.Input)

	runner := ffmpeg.NewRunner()
	runner.Duration = duration

	runner.OnProgress = func(info ffmpeg.ProgressInfo) {
		a.mu.Lock()
		task.Progress = info.Progress
		task.Status = "running"
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "task:progress", map[string]interface{}{
			"id":       id,
			"progress": info.Progress,
			"fps":      info.FPS,
			"bitrate":  info.Bitrate,
			"time":     info.Time,
			"frame":    info.Frame,
			"speed":    info.Speed,
		})
	}

	runner.OnLog = func(line string) {
		runtime.EventsEmit(a.ctx, "task:log", map[string]interface{}{
			"id":   id,
			"line": line,
		})
	}

	runner.OnDone = func(runErr error) {
		a.mu.Lock()
		if runErr != nil {
			task.Status = "failed"
			task.Error = runErr.Error()
		} else {
			task.Status = "completed"
			task.Progress = 1.0
		}
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "task:done", map[string]interface{}{
			"id":     id,
			"status": task.Status,
			"error":  task.Error,
		})
	}

	a.mu.Lock()
	a.runner = runner
	a.mu.Unlock()

	go func() {
		ctx := context.Background()
		if runErr := runner.Run(ctx, args); runErr != nil {
			// Handled by OnDone callback.
			_ = runErr
		}
	}()

	return task, nil
}

func (a *App) startStreamTask(id string, task *Task, payloadJSON string) (*Task, error) {
	var payload StreamPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return nil, fmt.Errorf("invalid stream payload: %w", err)
	}

	args, err := ffmpeg.BuildStreamArgs(ffmpeg.StreamOptions{
		Source:     payload.Source,
		Protocol:   payload.Protocol,
		URL:        payload.URL,
		VideoCodec: payload.VideoCodec,
		AudioCodec: payload.AudioCodec,
		Bitrate:    payload.Bitrate,
		Preset:     payload.Preset,
		Latency:    payload.Latency,
		IsLive:     payload.IsLive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build stream command: %w", err)
	}

	task.Input = payload.Source
	task.Output = payload.URL
	task.Command = fmt.Sprintf("ffmpeg %s", strings.Join(args, " "))
	task.Status = "running"

	runner := ffmpeg.NewRunner()

	runner.OnProgress = func(info ffmpeg.ProgressInfo) {
		a.mu.Lock()
		task.Status = "running"
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "task:progress", map[string]interface{}{
			"id":      id,
			"fps":     info.FPS,
			"bitrate": info.Bitrate,
			"time":    info.Time,
			"frame":   info.Frame,
			"speed":   info.Speed,
		})
	}

	runner.OnLog = func(line string) {
		runtime.EventsEmit(a.ctx, "task:log", map[string]interface{}{
			"id":   id,
			"line": line,
		})
	}

	runner.OnDone = func(runErr error) {
		a.mu.Lock()
		if runErr != nil {
			task.Status = "failed"
			task.Error = runErr.Error()
		} else {
			task.Status = "completed"
		}
		a.mu.Unlock()
		runtime.EventsEmit(a.ctx, "task:done", map[string]interface{}{
			"id":     id,
			"status": task.Status,
			"error":  task.Error,
		})
	}

	a.mu.Lock()
	a.runner = runner
	a.mu.Unlock()

	go func() {
		ctx := context.Background()
		if runErr := runner.Run(ctx, args); runErr != nil {
			_ = runErr
		}
	}()

	return task, nil
}

// GetTaskStatus returns the current status of a task by its ID.
func (a *App) GetTaskStatus(id string) (*Task, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	task, ok := a.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

// CancelTask cancels a running task by its ID.
func (a *App) CancelTask(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	task, ok := a.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}

	if task.Status != "running" {
		return fmt.Errorf("task is not running: %s", id)
	}

	if a.runner != nil {
		a.runner.Cancel()
	}

	task.Status = "canceled"
	task.Error = "canceled by user"

	runtime.EventsEmit(a.ctx, "task:done", map[string]interface{}{
		"id":     id,
		"status": "canceled",
		"error":  "canceled by user",
	})

	return nil
}

// ListTasks returns all tasks.
func (a *App) ListTasks() []*Task {
	a.mu.RLock()
	defer a.mu.RUnlock()

	tasks := make([]*Task, 0, len(a.tasks))
	for _, task := range a.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
