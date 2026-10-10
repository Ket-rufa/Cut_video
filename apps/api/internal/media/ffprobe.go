package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type FFprobeInfo struct {
	Format  FFprobeFormat   `json:"format"`
	Streams []FFprobeStream `json:"streams"`
}

type FFprobeFormat struct {
	DurationStr string `json:"duration"`
	Size        string `json:"size"`
}

type FFprobeStream struct {
	CodecType string `json:"codec_type"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  string `json:"duration,omitempty"`
}

func Probe(ctx context.Context, ffprobePath, filePath string) (*FFprobeInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}
	cmd := exec.CommandContext(ctx, ffprobePath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w\nstderr: %s", err, stderr.String())
	}

	var info FFprobeInfo
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}
	return &info, nil
}
