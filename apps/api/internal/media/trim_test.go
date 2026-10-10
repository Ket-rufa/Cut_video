package media_test

import (
	"strings"
	"testing"

	"social-video-cutter/api/internal/media"
)

func TestBuildTrimArgs_StreamCopy(t *testing.T) {
	opts := media.TrimOptions{
		InputPath:  "/input.mp4",
		OutputPath: "/output.mp4",
		StartMs:    5000,
		EndMs:      15000,
		Quality:    "original",
		Speed:      1.0,
	}
	args := media.BuildTrimArgs(opts)

	if args[0] != "-y" {
		t.Errorf("expected -y at index 0, got %s", args[0])
	}
	hasCopy := false
	for i, a := range args {
		if a == "-c" && i+1 < len(args) && args[i+1] == "copy" {
			hasCopy = true
		}
	}
	if !hasCopy {
		t.Error("expected -c copy for original quality with 1x speed")
	}
}

func TestBuildTrimArgs_SpeedChange(t *testing.T) {
	opts := media.TrimOptions{
		InputPath:  "/input.mp4",
		OutputPath: "/output.mp4",
		StartMs:    0,
		EndMs:      10000,
		Quality:    "original",
		Speed:      2.0,
	}
	args := media.BuildTrimArgs(opts)

	fullArgs := strings.Join(args, " ")
	if !strings.Contains(fullArgs, "setpts=0.5000*PTS") {
		t.Errorf("expected setpts=0.5000*PTS for 2x speed, got %s", fullArgs)
	}
	if !strings.Contains(fullArgs, "atempo=2.0000") {
		t.Errorf("expected atempo=2.0000 for 2x speed, got %s", fullArgs)
	}
}

func TestBuildTrimArgs_ChainedAtempo(t *testing.T) {
	opts := media.TrimOptions{
		InputPath:  "/input.mp4",
		OutputPath: "/output.mp4",
		StartMs:    0,
		EndMs:      10000,
		Quality:    "original",
		Speed:      4.0,
	}
	args := media.BuildTrimArgs(opts)

	fullArgs := strings.Join(args, " ")
	if !strings.Contains(fullArgs, "atempo=2.0,atempo=2.0000") {
		t.Errorf("expected chained atempo=2.0,atempo=2.0000 for 4x speed, got %s", fullArgs)
	}
}

func TestBuildTrimArgs_ColorFilter(t *testing.T) {
	opts := media.TrimOptions{
		InputPath:  "/input.mp4",
		OutputPath: "/output.mp4",
		StartMs:    0,
		EndMs:      10000,
		Quality:    "original",
		Speed:      1.0,
		FilterOpts: &media.FilterOptions{
			Preset:     "vivid",
			Brightness: 10,
			Contrast:   20,
		},
	}
	args := media.BuildTrimArgs(opts)

	fullArgs := strings.Join(args, " ")
	if !strings.Contains(fullArgs, "-vf") || !strings.Contains(fullArgs, "eq=") {
		t.Errorf("expected -vf eq=... filter for color options, got %s", fullArgs)
	}
}

func TestBuildTrimArgs_Output(t *testing.T) {
	opts := media.TrimOptions{
		InputPath:  "/input.mp4",
		OutputPath: "/output/clip.mp4",
		StartMs:    1000,
		EndMs:      5000,
		Quality:    "original",
	}
	args := media.BuildTrimArgs(opts)
	last := args[len(args)-1]
	if last != "/output/clip.mp4" {
		t.Errorf("expected output path as last arg, got %s", last)
	}
}
