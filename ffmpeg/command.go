// Package ffmpeg provides FFmpeg command building, execution, and streaming capabilities.
package ffmpeg

import (
	"fmt"
	"strings"
)

// ConvertOptions defines parameters for video/audio conversion tasks.
type ConvertOptions struct {
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

// StreamOptions defines parameters for RTMP/SRT streaming tasks.
type StreamOptions struct {
	Source     string `json:"source"`
	Protocol   string `json:"protocol"` // "rtmp" or "srt"
	URL        string `json:"url"`
	VideoCodec string `json:"videoCodec,omitempty"`
	AudioCodec string `json:"audioCodec,omitempty"`
	Bitrate    string `json:"bitrate,omitempty"`
	Preset     string `json:"preset,omitempty"`
	Latency    int    `json:"latency,omitempty"` // SRT latency in ms
	IsLive     bool   `json:"isLive,omitempty"`
}

// BuildConvertArgs constructs FFmpeg arguments for a conversion task.
func BuildConvertArgs(opts ConvertOptions) ([]string, error) {
	if opts.Input == "" {
		return nil, fmt.Errorf("input file is required")
	}
	if opts.Output == "" {
		return nil, fmt.Errorf("output file is required")
	}

	args := []string{"-i", opts.Input}

	// Video codec
	if opts.VideoCodec != "" {
		args = append(args, "-c:v", opts.VideoCodec)
	}

	// CRF (constant rate factor)
	if opts.CRF > 0 {
		args = append(args, "-crf", fmt.Sprintf("%d", opts.CRF))
	}

	// Video bitrate
	if opts.Bitrate != "" {
		args = append(args, "-b:v", opts.Bitrate)
	}

	// Build combined -vf filter chain for subtitle and scale
	var vfFilters []string
	if opts.SubtitleFile != "" {
		vfFilters = append(vfFilters, fmt.Sprintf("subtitles=%s", escapeFilterPath(opts.SubtitleFile)))
	}
	if opts.Resolution != "" {
		vfFilters = append(vfFilters, fmt.Sprintf("scale=%s", opts.Resolution))
	}
	if len(vfFilters) > 0 {
		args = append(args, "-vf", strings.Join(vfFilters, ","))
	}

	// Frame rate
	if opts.FPS > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", opts.FPS))
	}

	// Audio codec
	if opts.AudioCodec != "" {
		args = append(args, "-c:a", opts.AudioCodec)
	}

	// Audio bitrate
	if opts.AudioBitrate != "" {
		args = append(args, "-b:a", opts.AudioBitrate)
	}

	// Output format
	if opts.Format != "" {
		args = append(args, "-f", opts.Format)
	}

	// Extra user-provided arguments (placed before output for safety)
	if opts.ExtraArgs != "" {
		extra := strings.Fields(opts.ExtraArgs)
		args = append(args, extra...)
	}

	// Overwrite output without asking
	args = append(args, "-y", opts.Output)

	return args, nil
}

// BuildStreamArgs constructs FFmpeg arguments for a streaming task.
func BuildStreamArgs(opts StreamOptions) ([]string, error) {
	if opts.Source == "" {
		return nil, fmt.Errorf("stream source is required")
	}
	if opts.URL == "" {
		return nil, fmt.Errorf("stream URL is required")
	}

	args := []string{}

	// For file sources, use -re to match realtime; for live sources, don't
	if !opts.IsLive {
		args = append(args, "-re")
	}

	args = append(args, "-i", opts.Source)

	// Video codec
	vcodec := opts.VideoCodec
	if vcodec == "" {
		vcodec = "libx264"
	}
	args = append(args, "-c:v", vcodec)

	// Encoding preset
	preset := opts.Preset
	if preset == "" {
		preset = "veryfast"
	}
	args = append(args, "-preset", preset)

	// Video bitrate
	bitrate := opts.Bitrate
	if bitrate == "" {
		bitrate = "3000k"
	}
	args = append(args, "-b:v", bitrate)

	// Audio codec
	acodec := opts.AudioCodec
	if acodec == "" {
		acodec = "aac"
	}
	args = append(args, "-c:a", acodec)
	args = append(args, "-b:a", "128k")

	switch strings.ToLower(opts.Protocol) {
	case "rtmp":
		args = append(args, "-f", "flv", opts.URL)
	case "srt":
		// FFmpeg SRT latency is set via -latency flag, not URL query param
		if opts.Latency > 0 {
			args = append(args, "-latency", fmt.Sprintf("%d", opts.Latency))
		}
		args = append(args, "-f", "mpegts", opts.URL)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s (use rtmp or srt)", opts.Protocol)
	}

	return args, nil
}

// escapeFilterPath escapes special characters in file paths for FFmpeg filter syntax.
func escapeFilterPath(path string) string {
	r := strings.NewReplacer(
		"'", "'\\''",
		":", "\\:",
		"[", "\\[",
		"]", "\\]",
	)
	return r.Replace(path)
}
