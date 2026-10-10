package media

// BuildConvertArgs converts a video to the target format.
func BuildConvertArgs(inputPath, outputPath, format string) []string {
	args := []string{"-y", "-i", inputPath}
	switch format {
	case "webm":
		args = append(args, "-c:v", "libvpx-vp9", "-c:a", "libopus")
	default: // mp4
		args = append(args, "-c:v", "libx264", "-preset", "fast", "-c:a", "aac")
	}
	args = append(args, "-movflags", "+faststart", outputPath)
	return args
}
