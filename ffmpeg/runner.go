package ffmpeg

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProgressInfo contains parsed progress data from FFmpeg stderr.
type ProgressInfo struct {
	Frame    int64   `json:"frame"`
	FPS      float64 `json:"fps"`
	Bitrate  string  `json:"bitrate"`
	Time     string  `json:"time"`
	Progress float64 `json:"progress"`
	Speed    string  `json:"speed"`
}

// Runner executes FFmpeg commands and tracks progress.
type Runner struct {
	mu        sync.RWMutex
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	running   bool
	canceled  bool // set to true when Cancel() is called to suppress duplicate OnDone
	logs      []string

	// Duration of the input file in seconds, used for progress calculation.
	Duration float64

	// OnProgress is called with parsed progress information.
	OnProgress func(ProgressInfo)

	// OnLog is called for each line of FFmpeg output.
	OnLog func(string)

	// OnDone is called when the command finishes. The error is nil on success.
	OnDone func(error)
}

// NewRunner creates a new FFmpeg runner.
func NewRunner() *Runner {
	return &Runner{
		logs: make([]string, 0),
	}
}

// Run executes an FFmpeg command with the given arguments.
func (r *Runner) Run(ctx context.Context, args []string) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("a command is already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.cmd = exec.CommandContext(ctx, "ffmpeg", args...)
	r.running = true
	r.logs = make([]string, 0)
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		r.running = false
		r.mu.Unlock()
	}()

	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err = r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 0), 1024*1024) // 1MB buffer to avoid "token too long"
	for scanner.Scan() {
		line := scanner.Text()
		r.mu.Lock()
		r.logs = append(r.logs, line)
		r.mu.Unlock()

		if r.OnLog != nil {
			r.OnLog(line)
		}

		if info, ok := parseProgressLine(line); ok {
			if r.Duration > 0 {
				info.Progress = calcProgress(info.Time, r.Duration)
			}
			if r.OnProgress != nil {
				r.OnProgress(info)
			}
		}
	}

	if err = r.cmd.Wait(); err != nil {
		r.mu.Lock()
		wasCanceled := r.canceled
		r.mu.Unlock()
		if ctx.Err() == context.Canceled || wasCanceled {
			if r.OnDone != nil {
				r.OnDone(fmt.Errorf("canceled"))
			}
			return fmt.Errorf("canceled")
		}
		if r.OnDone != nil {
			r.OnDone(err)
		}
		return fmt.Errorf("ffmpeg exited with error: %w", err)
	}

	if r.OnProgress != nil {
		r.OnProgress(ProgressInfo{Progress: 1.0})
	}
	if r.OnDone != nil {
		r.OnDone(nil)
	}

	return nil
}

// Cancel terminates the running FFmpeg process.
func (r *Runner) Cancel() {
	r.mu.Lock()
	if r.cancel != nil {
		r.canceled = true
	}
	r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
}

// IsRunning reports whether a command is currently executing.
func (r *Runner) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.running
}

// Logs returns a copy of the captured log lines.
func (r *Runner) Logs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.logs))
	copy(out, r.logs)
	return out
}

// GetFFmpegVersion returns the installed FFmpeg version string.
func GetFFmpegVersion() (string, error) {
	cmd := exec.Command("ffmpeg", "-version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get ffmpeg version: %w", err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}
	return strings.TrimSpace(string(out)), nil
}

// GetInputDuration probes the duration of a media file using ffprobe.
func GetInputDuration(filePath string) (float64, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to probe file duration: %w", err)
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}
	return duration, nil
}

var (
	reFrame   = regexp.MustCompile(`frame=\s*(\d+)`)
	reFPS     = regexp.MustCompile(`fps=\s*([\d.]+)`)
	reBitrate = regexp.MustCompile(`bitrate=\s*([^\s]+)`)
	reTime    = regexp.MustCompile(`time=\s*(\d+:\d+:\d+\.\d+)`)
	reSpeed   = regexp.MustCompile(`speed=\s*([^\s]+)`)
)

// parseProgressLine extracts progress information from an FFmpeg status line.
func parseProgressLine(line string) (ProgressInfo, bool) {
	if !strings.Contains(line, "time=") {
		return ProgressInfo{}, false
	}

	var info ProgressInfo

	if m := reFrame.FindStringSubmatch(line); len(m) > 1 {
		info.Frame, _ = strconv.ParseInt(m[1], 10, 64)
	}
	if m := reFPS.FindStringSubmatch(line); len(m) > 1 {
		info.FPS, _ = strconv.ParseFloat(m[1], 64)
	}
	if m := reBitrate.FindStringSubmatch(line); len(m) > 1 {
		info.Bitrate = m[1]
	}
	if m := reTime.FindStringSubmatch(line); len(m) > 1 {
		info.Time = m[1]
	}
	if m := reSpeed.FindStringSubmatch(line); len(m) > 1 {
		info.Speed = m[1]
	}

	return info, true
}

// calcProgress converts a timestamp string to a percentage of total duration.
func calcProgress(timeStr string, totalDuration float64) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0
	}
	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0
	}

	current := hours*3600 + minutes*60 + seconds
	if totalDuration <= 0 {
		return 0
	}

	pct := current / totalDuration
	if pct > 1.0 {
		pct = 1.0
	}

	return pct
}

// FormatDuration formats a duration in seconds to HH:MM:SS.
func FormatDuration(d float64) string {
	if d <= 0 {
		return "00:00:00"
	}
	duration := time.Duration(d * float64(time.Second))
	h := int(duration.Hours())
	m := int(duration.Minutes()) % 60
	s := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
