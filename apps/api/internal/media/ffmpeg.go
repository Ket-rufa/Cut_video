package media

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type FFmpegRunner struct {
	ffmpegPath string
}

func NewFFmpegRunner(path string) *FFmpegRunner {
	return &FFmpegRunner{ffmpegPath: path}
}

func (r *FFmpegRunner) Run(ctx context.Context, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, r.ffmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nstderr: %s", err, stderr.String())
	}
	return nil
}
