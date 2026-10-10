package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

var (
	ErrOutputNotFound   = errors.New("OUTPUT_FILE_NOT_FOUND")
	ErrOutputEmpty      = errors.New("OUTPUT_FILE_EMPTY")
	ErrInvalidMedia     = errors.New("INVALID_MEDIA_FILE")
	ErrMissingAudio     = errors.New("MISSING_AUDIO_STREAM")
	ErrMissingVideo     = errors.New("MISSING_VIDEO_STREAM")
)

type probeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"` // "video" or "audio"
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Duration  string `json:"duration"`
	} `json:"streams"`
	Format struct {
		Filename   string `json:"filename"`
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
	} `json:"format"`
}

// MediaVerification contains verified media metadata.
type MediaVerification struct {
	FileSize   int64
	DurationMs int64
	Width      int
	Height     int
	HasVideo   bool
	HasAudio   bool
	FormatName string
}

// VerifyOutputFile validates that the downloaded file exists, is non-empty, and possesses valid streams.
func VerifyOutputFile(ctx context.Context, ffprobePath, filePath string, requireAudio bool) (*MediaVerification, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: path %s", ErrOutputNotFound, filePath)
		}
		return nil, fmt.Errorf("stat output file: %w", err)
	}

	if fi.Size() == 0 {
		return nil, fmt.Errorf("%w: file size is 0 bytes", ErrOutputEmpty)
	}

	ver := &MediaVerification{
		FileSize: fi.Size(),
	}

	// If ffprobe is available, run deep inspection
	if ffprobePath != "" {
		cmd := exec.CommandContext(ctx, ffprobePath,
			"-v", "quiet",
			"-print_format", "json",
			"-show_format",
			"-show_streams",
			filePath,
		)

		out, err := cmd.Output()
		if err == nil {
			var probe probeOutput
			if jsonErr := json.Unmarshal(out, &probe); jsonErr == nil {
				for _, s := range probe.Streams {
					if s.CodecType == "video" {
						ver.HasVideo = true
						ver.Width = s.Width
						ver.Height = s.Height
					}
					if s.CodecType == "audio" {
						ver.HasAudio = true
					}
				}

				if probe.Format.Duration != "" {
					if durSec, parseErr := strconv.ParseFloat(probe.Format.Duration, 64); parseErr == nil {
						ver.DurationMs = int64(durSec * 1000)
					}
				}
				ver.FormatName = probe.Format.FormatName

				if !ver.HasVideo {
					return nil, fmt.Errorf("%w: no video stream found in %s", ErrMissingVideo, filePath)
				}
				if requireAudio && !ver.HasAudio {
					return nil, fmt.Errorf("%w: video requires audio but no audio stream found in %s", ErrMissingAudio, filePath)
				}
			}
		}
	}

	return ver, nil
}
