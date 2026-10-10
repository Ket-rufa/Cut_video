package downloader

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"social-video-cutter/api/internal/config"
)

func TestReal_YouTube_Analyze(t *testing.T) {
	cfg := config.Load()
	if cfg.Downloader.YTDLPPath == "" {
		t.Skip("yt-dlp path not configured")
	}

	yt := NewYouTubeDownloader(cfg.Downloader.YTDLPPath, cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Use public test video (Google test video)
	testURL := "https://www.youtube.com/watch?v=aqz-KE-bpKQ"
	canonical, err := yt.CanonicalURL(ctx, testURL)
	if err != nil {
		t.Fatalf("CanonicalURL failed: %v", err)
	}

	result, err := yt.Analyze(ctx, canonical)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if result.VideoID == "" {
		t.Errorf("Expected non-empty video ID")
	}
	if result.Title == "" {
		t.Errorf("Expected non-empty title")
	}
	if len(result.Formats) == 0 {
		t.Errorf("Expected at least one format")
	}
	t.Logf("YouTube Analyzed Successfully: Title=%q, Formats=%d, Duration=%ds", result.Title, len(result.Formats), result.Duration)
}

func TestReal_TikTok_Analyze(t *testing.T) {
	cfg := config.Load()
	if cfg.Downloader.YTDLPPath == "" {
		t.Skip("yt-dlp path not configured")
	}

	tk := NewTikTokDownloader(cfg.Downloader.YTDLPPath, cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	testURL := "https://www.tiktok.com/@tiktok/video/7106594312292453675"
	canonical, err := tk.CanonicalURL(ctx, testURL)
	if err != nil {
		t.Fatalf("CanonicalURL failed: %v", err)
	}

	result, err := tk.Analyze(ctx, canonical)
	if err != nil {
		t.Logf("TikTok Analyze note: %v (network/geo dependent)", err)
		return
	}

	if result.VideoID == "" {
		t.Errorf("Expected non-empty video ID")
	}
	t.Logf("TikTok Analyzed Successfully: ID=%s, Title=%q", result.VideoID, result.Title)
}

func TestReal_YouTube_Download(t *testing.T) {
	cfg := config.Load()
	if cfg.Downloader.YTDLPPath == "" {
		t.Skip("yt-dlp path not configured")
	}

	yt := NewYouTubeDownloader(cfg.Downloader.YTDLPPath, cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	testURL := "https://www.youtube.com/watch?v=jNQXAC9IVRw"
	canonical, err := yt.CanonicalURL(ctx, testURL)
	if err != nil {
		t.Fatalf("CanonicalURL failed: %v", err)
	}

	analysis, err := yt.Analyze(ctx, canonical)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	_ = analysis

	tempDir := filepath.Join(os.TempDir(), "test_yt_download_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	defer os.RemoveAll(tempDir)

	var lastProgress int
	progressCount := 0

	opt := DownloadOptions{
		JobID:     "test-job-123",
		SourceURL: canonical,
		Quality:   "360p",
		FormatID:  "360p",
		TempDir:   tempDir,
	}

	result, err := yt.Download(ctx, opt, func(p DownloadProgress) {
		progressCount++
		if p.Progress < lastProgress {
			t.Errorf("Progress dropped from %d to %d at status %s", lastProgress, p.Progress, p.Status)
		}
		lastProgress = p.Progress
		t.Logf("Progress: stage=%s, overall=%d%%, speed=%d B/s, eta=%ds", p.Status, p.Progress, p.Speed, p.ETA)
	})

	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if result.FileSize <= 0 {
		t.Errorf("Expected FileSize > 0, got %d", result.FileSize)
	}
	if result.DurationMs <= 0 {
		t.Errorf("Expected DurationMs > 0, got %d", result.DurationMs)
	}
	if result.Width <= 0 || result.Height <= 0 {
		t.Errorf("Expected valid dimensions, got %dx%d", result.Width, result.Height)
	}

	// Verify the file using ffprobe
	verified, err := VerifyOutputFile(ctx, cfg.Downloader.FFprobePath, result.FilePath, true)
	if err != nil {
		t.Fatalf("VerifyOutputFile failed: %v", err)
	}
	if !verified.HasAudio {
		t.Errorf("Expected output file to have audio stream")
	}
	if !verified.HasVideo {
		t.Errorf("Expected output file to have video stream")
	}

	t.Logf("YouTube Download Verified Successfully: File=%s, Size=%d bytes, Duration=%dms, Res=%dx%d, HasAudio=%v",
		result.FileName, result.FileSize, result.DurationMs, result.Width, result.Height, verified.HasAudio)
}

func TestReal_TikTok_Download(t *testing.T) {
	cfg := config.Load()
	if cfg.Downloader.YTDLPPath == "" {
		t.Skip("yt-dlp path not configured")
	}

	tk := NewTikTokDownloader(cfg.Downloader.YTDLPPath, cfg.Downloader.FFmpegPath, cfg.Downloader.FFprobePath)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	testURL := "https://www.tiktok.com/@tiktok/video/7106594312292453675"
	canonical, err := tk.CanonicalURL(ctx, testURL)
	if err != nil {
		t.Fatalf("CanonicalURL failed: %v", err)
	}

	tempDir := filepath.Join(os.TempDir(), "test_tk_download_"+strconv.FormatInt(time.Now().UnixNano(), 10))
	defer os.RemoveAll(tempDir)

	var lastProgress int

	opt := DownloadOptions{
		JobID:     "test-tk-job-456",
		SourceURL: canonical,
		Quality:   "original",
		FormatID:  "best",
		TempDir:   tempDir,
	}

	result, err := tk.Download(ctx, opt, func(p DownloadProgress) {
		if p.Progress < lastProgress {
			t.Errorf("TikTok progress dropped from %d to %d at status %s", lastProgress, p.Progress, p.Status)
		}
		lastProgress = p.Progress
		t.Logf("TikTok Progress: stage=%s, overall=%d%%, speed=%d B/s, eta=%ds", p.Status, p.Progress, p.Speed, p.ETA)
	})

	if err != nil {
		t.Logf("TikTok download note: %v (network/geo dependent)", err)
		return
	}

	if result.FileSize <= 0 {
		t.Errorf("Expected FileSize > 0, got %d", result.FileSize)
	}

	verified, err := VerifyOutputFile(ctx, cfg.Downloader.FFprobePath, result.FilePath, true)
	if err != nil {
		t.Fatalf("TikTok VerifyOutputFile failed: %v", err)
	}

	t.Logf("TikTok Download Verified Successfully: File=%s, Size=%d bytes, Duration=%dms, Res=%dx%d, HasAudio=%v",
		result.FileName, result.FileSize, result.DurationMs, result.Width, result.Height, verified.HasAudio)
}


