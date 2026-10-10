package media

import (
	"fmt"
	"math"
	"strings"
)

type FilterOptions struct {
	Preset      string  `json:"preset,omitempty"`
	Brightness  float64 `json:"brightness,omitempty"`  // -100 to +100
	Contrast    float64 `json:"contrast,omitempty"`    // -100 to +100
	Saturation  float64 `json:"saturation,omitempty"`  // -100 to +100
	Temperature float64 `json:"temperature,omitempty"` // -100 to +100
}

type TrimOptions struct {
	InputPath   string
	OutputPath  string
	StartMs     int64
	EndMs       int64
	Quality     string
	Format      string
	Speed       float64
	FilterOpts  *FilterOptions
}

// BuildTrimArgs constructs safe FFmpeg arguments for trimming, speed adjustment, and color filtering.
// Uses stream copy for speed when quality is "original" and no speed or color filters are applied.
func BuildTrimArgs(opts TrimOptions) []string {
	startSec := float64(opts.StartMs) / 1000.0
	durationSec := float64(opts.EndMs-opts.StartMs) / 1000.0

	args := []string{
		"-y",
		"-ss", fmt.Sprintf("%.3f", startSec),
		"-i", opts.InputPath,
		"-t", fmt.Sprintf("%.3f", durationSec),
	}

	needsReencode := (opts.Quality != "original" && opts.Quality != "") ||
		(opts.Speed > 0 && math.Abs(opts.Speed-1.0) > 0.01) ||
		(opts.FilterOpts != nil && isFilterActive(*opts.FilterOpts))

	if !needsReencode {
		// Stream copy — fastest, no re-encoding
		args = append(args, "-c", "copy")
	} else {
		args = append(args, "-c:v", "libx264", "-preset", "fast", "-c:a", "aac")

		var videoFilters []string
		var audioFilters []string

		// Scale quality
		switch opts.Quality {
		case "1080p":
			videoFilters = append(videoFilters, "scale=-2:1080")
		case "720p":
			videoFilters = append(videoFilters, "scale=-2:720")
		case "480p":
			videoFilters = append(videoFilters, "scale=-2:480")
		}

		// Speed adjustment
		if opts.Speed > 0 && math.Abs(opts.Speed-1.0) > 0.01 {
			ptsFactor := 1.0 / opts.Speed
			videoFilters = append(videoFilters, fmt.Sprintf("setpts=%.4f*PTS", ptsFactor))

			// Chain atempo (0.5 to 2.0 each)
			currentSpeed := opts.Speed
			for currentSpeed > 2.0 {
				audioFilters = append(audioFilters, "atempo=2.0")
				currentSpeed /= 2.0
			}
			for currentSpeed < 0.5 {
				audioFilters = append(audioFilters, "atempo=0.5")
				currentSpeed /= 0.5
			}
			if math.Abs(currentSpeed-1.0) > 0.01 {
				audioFilters = append(audioFilters, fmt.Sprintf("atempo=%.4f", currentSpeed))
			}
		}

		// Color filters
		if opts.FilterOpts != nil {
			f := *opts.FilterOpts
			b := f.Brightness
			c := f.Contrast
			s := f.Saturation
			t := f.Temperature

			// Merge presets
			switch f.Preset {
			case "vivid":
				b += 10; c += 15; s += 30; t += 5
			case "warm":
				b += 5; c += 10; s += 15; t += 40
			case "cool":
				c += 10; s -= 5; t -= 40
			case "cinematic":
				b -= 5; c += 25; s -= 15; t -= 10
			case "vintage":
				b += 5; c -= 10; s -= 20; t += 25
			case "bw":
				b += 5; c += 20; s = -100
			}

			// Clamp
			b = math.Max(-100, math.Min(100, b))
			c = math.Max(-100, math.Min(100, c))
			s = math.Max(-100, math.Min(100, s))
			t = math.Max(-100, math.Min(100, t))

			var eqParts []string
			ffBrightness := b / 200.0
			ffContrast := 1.0 + c/100.0
			ffSaturation := math.Max(0, 1.0+s/100.0)
			if f.Preset == "bw" {
				ffSaturation = 0.0
			}

			if math.Abs(ffBrightness) > 0.01 {
				eqParts = append(eqParts, fmt.Sprintf("brightness=%.2f", ffBrightness))
			}
			if math.Abs(ffContrast-1.0) > 0.01 {
				eqParts = append(eqParts, fmt.Sprintf("contrast=%.2f", ffContrast))
			}
			if math.Abs(ffSaturation-1.0) > 0.01 || f.Preset == "bw" {
				eqParts = append(eqParts, fmt.Sprintf("saturation=%.2f", ffSaturation))
			}

			if len(eqParts) > 0 {
				videoFilters = append(videoFilters, fmt.Sprintf("eq=%s", strings.Join(eqParts, ":")))
			}

			// Temperature via colorbalance
			if math.Abs(t) > 5 {
				factor := (t / 100.0) * 0.3
				if factor > 0 {
					videoFilters = append(videoFilters, fmt.Sprintf("colorbalance=rm=%.2f:gm=0.0:bm=-%.2f", factor, factor*0.8))
				} else {
					absFactor := math.Abs(factor)
					videoFilters = append(videoFilters, fmt.Sprintf("colorbalance=rm=-%.2f:gm=0.0:bm=%.2f", absFactor*0.8, absFactor))
				}
			}
		}

		if len(videoFilters) > 0 {
			args = append(args, "-vf", strings.Join(videoFilters, ","))
		}
		if len(audioFilters) > 0 {
			args = append(args, "-af", strings.Join(audioFilters, ","))
		}
	}

	args = append(args, "-movflags", "+faststart", opts.OutputPath)
	return args
}

func isFilterActive(f FilterOptions) bool {
	return f.Preset != "" && f.Preset != "none" ||
		math.Abs(f.Brightness) > 0.01 ||
		math.Abs(f.Contrast) > 0.01 ||
		math.Abs(f.Saturation) > 0.01 ||
		math.Abs(f.Temperature) > 0.01
}
