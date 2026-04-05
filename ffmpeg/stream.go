package ffmpeg

import (
	"context"
	"fmt"
)

// StreamConfig holds the configuration for launching an FFmpeg stream.
type StreamConfig struct {
	Source     string
	Protocol   string // "rtmp" or "srt"
	URL        string
	VideoCodec string
	AudioCodec string
	Bitrate    string
	Preset     string
	Latency    int
	IsLive     bool
}

// StreamOutput contains runtime information about an active stream.
type StreamOutput struct {
	Args   []string
	Runner *Runner
}

// StartStream creates a Runner and starts an FFmpeg streaming session.
func StartStream(ctx context.Context, cfg StreamConfig) (*StreamOutput, error) {
	args, err := BuildStreamArgs(StreamOptions(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to build stream args: %w", err)
	}

	runner := NewRunner()

	go func() {
		if runErr := runner.Run(ctx, args); runErr != nil {
			// Error is handled by OnDone callback if set.
			_ = runErr
		}
	}()

	return &StreamOutput{
		Args:   args,
		Runner: runner,
	}, nil
}
