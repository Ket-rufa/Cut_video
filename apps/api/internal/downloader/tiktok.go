package downloader

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"social-video-cutter/api/internal/domain"
)

type TikTokDownloader struct {
	ytdlpPath   string
	ffmpegPath  string
	ffprobePath string
}

func NewTikTokDownloader(ytdlpPath, ffmpegPath, ffprobePath string) *TikTokDownloader {
	return &TikTokDownloader{
		ytdlpPath:   ytdlpPath,
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}
}

func (d *TikTokDownloader) Platform() domain.VideoPlatform {
	return domain.PlatformTikTok
}

func (d *TikTokDownloader) Supports(rawURL string) bool {
	u, err := ValidateTikTokURL(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	path := strings.ToLower(u.Path)

	if strings.Contains(host, "vm.tiktok.com") || strings.Contains(host, "vt.tiktok.com") {
		return strings.TrimPrefix(path, "/") != ""
	}
	if strings.Contains(host, "tiktok.com") {
		return strings.Contains(path, "/video/") || strings.Contains(path, "/v/")
	}
	return false
}

// CanonicalURL resolves short links (vm.tiktok.com, vt.tiktok.com) with hop-by-hop SSRF validation.
func (d *TikTokDownloader) CanonicalURL(ctx context.Context, rawURL string) (string, error) {
	u, err := ValidateTikTokURL(rawURL)
	if err != nil {
		return "", err
	}

	host := strings.ToLower(u.Hostname())
	// If short URL, resolve redirects with strict hop-by-hop check (Rule #6)
	if strings.Contains(host, "vm.tiktok.com") || strings.Contains(host, "vt.tiktok.com") {
		resolved, err := ResolveRedirectsSafely(ctx, rawURL, allowedTikTokHosts, 5)
		if err != nil {
			return "", fmt.Errorf("resolve TikTok short link: %w", err)
		}
		return resolved, nil
	}

	return u.String(), nil
}

func (d *TikTokDownloader) Analyze(ctx context.Context, canonicalURL string) (*VideoAnalysisResult, error) {
	if _, err := ValidateTikTokURL(canonicalURL); err != nil {
		return nil, err
	}

	args := []string{
		"--dump-json",
		"--no-playlist",
		"--skip-download",
		"--no-warnings",
		"--",
		canonicalURL,
	}

	cmd := exec.CommandContext(ctx, d.ytdlpPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, normalizeYtDlpError(err, out)
	}

	var meta ytDlpMetadata
	if err := json.Unmarshal(out, &meta); err != nil {
		return nil, fmt.Errorf("parse TikTok metadata JSON: %w", err)
	}

	authorName := meta.Uploader
	if authorName == "" {
		authorName = meta.Channel
	}

	// TikTok videos are typically standard MP4 vertical videos
	var formats []VideoFormatInfo
	formats = append(formats, VideoFormatInfo{
		FormatID:  "best",
		Container: "mp4",
		Quality:   "original",
		Width:     1080,
		Height:    1920,
		HasVideo:  true,
		HasAudio:  true,
	})

	// Add 720p option if available
	formats = append(formats, VideoFormatInfo{
		FormatID:  "720p",
		Container: "mp4",
		Quality:   "720p",
		Width:     720,
		Height:    1280,
		HasVideo:  true,
		HasAudio:  true,
	})

	// Audio only
	formats = append(formats, VideoFormatInfo{
		FormatID:  "bestaudio",
		Container: "m4a",
		Quality:   "audio_only",
		HasVideo:  false,
		HasAudio:  true,
		AudioOnly: true,
	})

	title := meta.Title
	if title == "" {
		title = meta.Description
	}
	if title == "" {
		title = fmt.Sprintf("TikTok Video (%s)", meta.ID)
	}

	return &VideoAnalysisResult{
		Platform:    domain.PlatformTikTok,
		SourceURL:   canonicalURL,
		VideoID:     meta.ID,
		Title:       title,
		Description: meta.Description,
		Thumbnail:   meta.Thumbnail,
		Duration:    int(meta.Duration),
		DurationMs:  int64(meta.Duration) * 1000,
		Author: AuthorInfo{
			Name: authorName,
		},
		Formats: formats,
	}, nil
}

func (d *TikTokDownloader) Download(ctx context.Context, opt DownloadOptions, progress ProgressCallback) (*DownloadResult, error) {
	if _, err := ValidateTikTokURL(opt.SourceURL); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(opt.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("create job temp dir: %w", err)
	}

	progress(DownloadProgress{
		Status:   "ANALYZING",
		Progress: 2,
	})

	outputTemplate := filepath.Join(opt.TempDir, "video.%(ext)s")

	// Standard format selector for TikTok
	formatSelector := "best[ext=mp4]/best"
	if opt.Quality == "audio_only" {
		formatSelector = "bestaudio/best"
	}

	args := []string{
		"--no-playlist",
		"--newline",
		"--concurrent-fragments", "8",
		"--buffer-size", "16M",
		"--retries", "10",
		"--fragment-retries", "10",
		"--file-access-retries", "5",
		"--no-mtime",
		"--format", formatSelector,
		"--output", outputTemplate,
	}

	if d.ffmpegPath != "" {
		ffmpegDir := filepath.Dir(d.ffmpegPath)
		args = append(args, "--ffmpeg-location", ffmpegDir)
	}

	args = append(args, "--", opt.SourceURL)

	cmd := exec.CommandContext(ctx, d.ytdlpPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start yt-dlp: %w", err)
	}

	lastProgressPct := 5
	progress(DownloadProgress{
		Status:   "DOWNLOADING",
		Progress: 5,
	})

	scanner := bufio.NewScanner(stdout)
	progressRegex := regexp.MustCompile(`\[download\]\s+([\d\.]+)%\s+of\s+~?([\d\.]+\w+)(?:\s+at\s+([\d\.]+\w+/s))?(?:\s+ETA\s+(\d+:\d+:\d+|\d+:\d+|\d+))?`)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			matches := progressRegex.FindStringSubmatch(line)
			if len(matches) >= 3 {
				if pctFloat, parseErr := strconv.ParseFloat(matches[1], 64); parseErr == nil {
					scaled := 5 + int(pctFloat*0.85)
					if scaled > lastProgressPct && scaled <= 90 {
						lastProgressPct = scaled
					}

					totalBytes := parseBytesString(matches[2])
					downloadedBytes := int64(float64(totalBytes) * (pctFloat / 100.0))
					var speed int64
					var eta int
					if len(matches) >= 4 && matches[3] != "" {
						speed = parseSpeedString(matches[3])
					}
					if len(matches) >= 5 && matches[4] != "" {
						eta = parseETAString(matches[4])
					}

					progress(DownloadProgress{
						Status:          "DOWNLOADING",
						Progress:        lastProgressPct,
						DownloadedBytes: downloadedBytes,
						TotalBytes:      totalBytes,
						Speed:           speed,
						ETA:             eta,
					})
				}
			}
		}
	}()

	var errLines []string
	errScanner := bufio.NewScanner(stderr)
	go func() {
		for errScanner.Scan() {
			errLines = append(errLines, errScanner.Text())
		}
	}()

	cmdErr := cmd.Wait()
	if cmdErr != nil {
		errDetail := strings.Join(errLines, "\n")
		return nil, fmt.Errorf("%w: %s", normalizeYtDlpError(cmdErr, []byte(errDetail)), errDetail)
	}

	// Locate downloaded file in temp dir
	expectedMP4 := filepath.Join(opt.TempDir, "video.mp4")
	if _, err := os.Stat(expectedMP4); os.IsNotExist(err) {
		entries, _ := os.ReadDir(opt.TempDir)
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".mp4") || strings.HasSuffix(entry.Name(), ".webm")) {
				expectedMP4 = filepath.Join(opt.TempDir, entry.Name())
				break
			}
		}
	}

	// Verify output
	verified, verifyErr := VerifyOutputFile(ctx, d.ffprobePath, expectedMP4, opt.Quality != "audio_only")
	if verifyErr != nil {
		return nil, fmt.Errorf("verify TikTok output: %w", verifyErr)
	}

	progress(DownloadProgress{
		Status:   "SAVING",
		Progress: 96,
	})

	return &DownloadResult{
		FilePath:   expectedMP4,
		FileName:   filepath.Base(expectedMP4),
		FileSize:   verified.FileSize,
		MimeType:   "video/mp4",
		Width:      verified.Width,
		Height:     verified.Height,
		DurationMs: verified.DurationMs,
	}, nil
}
