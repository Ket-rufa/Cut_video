package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type Runner struct {
	FFmpegPath  string
	FFprobePath string
}

func NewRunner(ffmpegPath, ffprobePath string) *Runner {
	return &Runner{FFmpegPath: ffmpegPath, FFprobePath: ffprobePath}
}

// Run executes FFmpeg with the given args.
// If progressFn is non-nil, it is called with real progress values (0-99)
// parsed from FFmpeg's stderr output. Progress reaches 100 only after
// the command exits successfully (set by the caller).
func (r *Runner) Run(ctx context.Context, args []string, progressFn func(int)) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.FFmpegPath, args...)

	var stderrBuf bytes.Buffer

	if progressFn != nil {
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("ffmpeg stderr pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("ffmpeg start: %w", err)
		}

		// Drain stderr: tee into buffer for error reporting, and parse real progress.
		parserDone := make(chan struct{})
		go func() {
			defer close(parserDone)
			tee := io.TeeReader(stderrPipe, &stderrBuf)
			ParseProgress(tee, 0, progressFn)
		}()

		cmdErr := cmd.Wait()
		<-parserDone // Ensure all stderr is consumed before reading stderrBuf

		if cmdErr != nil {
			return fmt.Errorf("ffmpeg: %w\nstderr: %s", cmdErr, stderrBuf.String())
		}
		return nil
	}

	// No progress tracking — simple synchronous execution.
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w\nstderr: %s", err, stderrBuf.String())
	}
	return nil
}

// Download downloads video from url to outputPath using ffmpeg.
func (r *Runner) Download(ctx context.Context, url, outputPath string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	args := []string{"-y", "-i", url, "-c", "copy", outputPath}
	cmd := exec.CommandContext(ctx, r.FFmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("download: %w\nstderr: %s", err, stderr.String())
	}
	return nil
}
