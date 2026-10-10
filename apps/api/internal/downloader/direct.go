package downloader

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"social-video-cutter/api/internal/domain"
)

// DirectDownloader handles direct CDN / MP4 / stream links (e.g. CoCoCut captured URLs, douyinvod.com, etc.)
type DirectDownloader struct {
	ffmpegPath  string
	ffprobePath string
	httpClient  *http.Client
}

func NewDirectDownloader(ffmpegPath, ffprobePath string) *DirectDownloader {
	return &DirectDownloader{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (d *DirectDownloader) Platform() domain.VideoPlatform {
	return domain.PlatformGeneric
}

func (d *DirectDownloader) Supports(rawURL string) bool {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	lower := strings.ToLower(u.String())

	// Exclude social platform main web pages (handled by dedicated adapters)
	if (strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be")) && !strings.Contains(host, "googlevideo.com") {
		return false
	}
	if (strings.Contains(host, "tiktok.com") || strings.Contains(host, "douyin.com") || strings.Contains(host, "iesdouyin.com")) &&
		!strings.Contains(host, "douyinvod.com") && !strings.Contains(host, "tiktokcdn.com") {
		return false
	}

	// Known video CDNs and media extensions
	if strings.Contains(host, "zjcdn.com") ||
		strings.Contains(host, "douyinvod.com") ||
		strings.Contains(host, "bytegoofy.com") ||
		strings.Contains(host, "bytetos.com") ||
		strings.Contains(host, "bytecdn.com") ||
		strings.Contains(host, "bytescm.com") ||
		strings.Contains(host, "snssdk.com") ||
		strings.Contains(host, "amemv.com") ||
		strings.Contains(host, "pstatp.com") ||
		strings.Contains(host, "tiktokcdn.com") ||
		strings.Contains(host, "fbcdn.net") ||
		strings.Contains(host, "cdninstagram.com") ||
		strings.Contains(host, "googlevideo.com") ||
		strings.Contains(host, "cdn") ||
		strings.Contains(host, "vod") ||
		strings.Contains(host, "stream") ||
		strings.Contains(lower, ".mp4") ||
		strings.Contains(lower, ".m4v") ||
		strings.Contains(lower, ".mov") ||
		strings.Contains(lower, ".webm") ||
		strings.Contains(lower, ".m3u8") ||
		strings.Contains(lower, ".ts") ||
		strings.Contains(lower, ".m4a") ||
		strings.Contains(lower, "/video/") ||
		strings.Contains(lower, "/play/") {
		return true
	}

	// Any other direct URL with path
	return len(u.Path) > 5
}

func (d *DirectDownloader) CanonicalURL(ctx context.Context, rawURL string) (string, error) {
	return strings.TrimSpace(rawURL), nil
}

func (d *DirectDownloader) Analyze(ctx context.Context, canonicalURL string) (*VideoAnalysisResult, error) {
	hash := md5.Sum([]byte(canonicalURL))
	id := hex.EncodeToString(hash[:6])

	width := 1280
	height := 720
	duration := 60

	// Use ffprobe to inspect stream properties
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-headers", "Referer: https://www.douyin.com/\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36\r\n",
		canonicalURL,
	}
	cmd := exec.CommandContext(ctx, d.ffprobePath, args...)
	out, err := cmd.Output()
	if err == nil {
		var probeData struct {
			Format struct {
				Duration string `json:"duration"`
			} `json:"format"`
			Streams []struct {
				CodecType string `json:"codec_type"`
				Width     int    `json:"width"`
				Height    int    `json:"height"`
			} `json:"streams"`
		}
		if json.Unmarshal(out, &probeData) == nil {
			for _, s := range probeData.Streams {
				if s.CodecType == "video" && s.Width > 0 {
					width = s.Width
					height = s.Height
				}
			}
			if probeData.Format.Duration != "" {
				var durFloat float64
				fmt.Sscanf(probeData.Format.Duration, "%f", &durFloat)
				if durFloat > 0 {
					duration = int(durFloat)
				}
			}
		}
	}

	formats := []VideoFormatInfo{
		{
			FormatID:  "direct",
			Container: "mp4",
			Quality:   "original",
			Width:     width,
			Height:    height,
			HasVideo:  true,
			HasAudio:  true,
		},
	}

	return &VideoAnalysisResult{
		Platform:    domain.PlatformGeneric,
		SourceURL:   canonicalURL,
		VideoID:     id,
		Title:       fmt.Sprintf("Video trực tiếp (%s)", id),
		Description: "Tải từ liên kết trực tiếp CDN / CoCoCut",
		Duration:    duration,
		DurationMs:  int64(duration) * 1000,
		Formats:     formats,
	}, nil
}

func (d *DirectDownloader) Download(ctx context.Context, opt DownloadOptions, progress ProgressCallback) (*DownloadResult, error) {
	if err := os.MkdirAll(opt.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	progress(DownloadProgress{
		Status:   "DOWNLOADING",
		Progress: 10,
	})

	outputPath := filepath.Join(opt.TempDir, "video.mp4")

	// Try downloading with ffmpeg stream copy first
	args := []string{
		"-y",
		"-headers", "Referer: https://www.douyin.com/\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36\r\n",
		"-i", opt.SourceURL,
		"-c", "copy",
		outputPath,
	}
	cmd := exec.CommandContext(ctx, d.ffmpegPath, args...)
	if err := cmd.Run(); err != nil {
		// Fallback to HTTP download
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, opt.SourceURL, nil)
		if err != nil {
			return nil, fmt.Errorf("prepare direct download request: %w", err)
		}
		req.Header.Set("Referer", "https://www.douyin.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, httpErr := d.httpClient.Do(req)
		if httpErr != nil {
			return nil, fmt.Errorf("execute direct download: %w", httpErr)
		}
		defer resp.Body.Close()

		f, createErr := os.Create(outputPath)
		if createErr != nil {
			return nil, fmt.Errorf("create output file: %w", createErr)
		}
		defer f.Close()

		if _, copyErr := io.Copy(f, resp.Body); copyErr != nil {
			return nil, fmt.Errorf("stream direct video: %w", copyErr)
		}
	}

	progress(DownloadProgress{
		Status:   "COMPLETED",
		Progress: 100,
	})

	var fileSize int64
	if fi, statErr := os.Stat(outputPath); statErr == nil {
		fileSize = fi.Size()
	}

	return &DownloadResult{
		FilePath:   outputPath,
		FileName:   filepath.Base(outputPath),
		FileSize:   fileSize,
		MimeType:   "video/mp4",
		Width:      1280,
		Height:     720,
		DurationMs: 0,
		Title:      "Video trực tiếp",
	}, nil
}
