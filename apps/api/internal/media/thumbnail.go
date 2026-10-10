package media

import "fmt"

// BuildThumbnailArgs constructs FFmpeg arguments to extract a thumbnail at offsetMs.
func BuildThumbnailArgs(inputPath, outputPath string, offsetMs int64) []string {
	offsetSec := float64(offsetMs) / 1000.0
	return []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", offsetSec),
		"-i", inputPath,
		"-vframes", "1",
		"-q:v", "2",
		outputPath,
	}
}
