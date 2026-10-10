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

type YouTubeDownloader struct {
	ytdlpPath   string
	ffmpegPath  string
	ffprobePath string
}

func NewYouTubeDownloader(ytdlpPath, ffmpegPath, ffprobePath string) *YouTubeDownloader {
	return &YouTubeDownloader{
		ytdlpPath:   ytdlpPath,
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}
}

func (d *YouTubeDownloader) Platform() domain.VideoPlatform {
	return domain.PlatformYouTube
}

func (d *YouTubeDownloader) Supports(rawURL string) bool {
	u, err := ValidateYouTubeURL(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	path := strings.ToLower(u.Path)

	if strings.Contains(host, "youtu.be") {
		return strings.TrimPrefix(path, "/") != ""
	}
	if strings.Contains(host, "youtube.com") {
		return strings.Contains(path, "/watch") ||
			strings.Contains(path, "/shorts/") ||
			strings.Contains(path, "/live/")
	}
	return false
}

func (d *YouTubeDownloader) CanonicalURL(ctx context.Context, rawURL string) (string, error) {
	u, err := ValidateYouTubeURL(rawURL)
	if err != nil {
		return "", err
	}

	host := strings.ToLower(u.Hostname())
	path := u.Path

	// youtu.be/<id>
	if strings.Contains(host, "youtu.be") {
		id := strings.TrimPrefix(path, "/")
		if id == "" {
			return "", fmt.Errorf("%w: empty video ID in youtu.be URL", ErrInvalidURL)
		}
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s", id), nil
	}

	// youtube.com/shorts/<id>
	if strings.Contains(path, "/shorts/") {
		parts := strings.Split(strings.TrimPrefix(path, "/shorts/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			return fmt.Sprintf("https://www.youtube.com/watch?v=%s", parts[0]), nil
		}
	}

	// youtube.com/watch?v=<id>
	v := u.Query().Get("v")
	if v != "" {
		// Strip playlist/other params to avoid downloading whole playlist
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s", v), nil
	}

	return u.String(), nil
}

// ytDlpMetadata is the JSON schema returned by yt-dlp --dump-json
type ytDlpMetadata struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Thumbnail   string  `json:"thumbnail"`
	Duration    float64 `json:"duration"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	Uploader    string  `json:"uploader"`
	Channel     string  `json:"channel"`
	Formats     []struct {
		FormatID       string  `json:"format_id"`
		Ext            string  `json:"ext"`
		Width          int     `json:"width"`
		Height         int     `json:"height"`
		FPS            float64 `json:"fps"`
		VCodec         string  `json:"vcodec"`
		ACodec         string  `json:"acodec"`
		FileSize       int64   `json:"filesize"`
		FileSizeApprox int64   `json:"filesize_approx"`
		FormatNote     string  `json:"format_note"`
		Tbr            float64 `json:"tbr"`
	} `json:"formats"`
}

func (d *YouTubeDownloader) Analyze(ctx context.Context, canonicalURL string) (*VideoAnalysisResult, error) {
	if _, err := ValidateYouTubeURL(canonicalURL); err != nil {
		return nil, err
	}

	// Run yt-dlp --dump-json --no-playlist --skip-download
	args := []string{
		"--dump-json",
		"--no-playlist",
		"--skip-download",
		"--no-warnings",
		"--no-check-certificates",
		"--socket-timeout", "15",
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
		return nil, fmt.Errorf("parse video metadata JSON: %w", err)
	}

	authorName := meta.Uploader
	if authorName == "" {
		authorName = meta.Channel
	}

	// Normalize quality buckets
	qualityMap := make(map[string]VideoFormatInfo)
	hasAudioFormats := false

	for _, f := range meta.Formats {
		hasV := f.VCodec != "none" && f.VCodec != ""
		hasA := f.ACodec != "none" && f.ACodec != ""
		if hasA {
			hasAudioFormats = true
		}

		size := f.FileSize
		if size <= 0 {
			size = f.FileSizeApprox
		}

		// Video formats
		if hasV {
			var qLabel string
			switch {
			case f.Height >= 4320:
				qLabel = "4320p"
			case f.Height >= 2160:
				qLabel = "2160p"
			case f.Height >= 1440:
				qLabel = "1440p"
			case f.Height >= 1080:
				qLabel = "1080p"
			case f.Height >= 720:
				qLabel = "720p"
			case f.Height >= 480:
				qLabel = "480p"
			case f.Height >= 360:
				qLabel = "360p"
			case f.Height >= 240:
				qLabel = "240p"
			case f.Height >= 144:
				qLabel = "144p"
			default:
				if f.Height > 0 {
					qLabel = fmt.Sprintf("%dp", f.Height)
				} else {
					continue
				}
			}

			// Format scoring: prioritize resolution, then fps, then bitrate, then filesize
			candidateScore := int64(f.Height)*10000000 + int64(f.FPS)*100000 + int64(f.Tbr*10) + size/1024

			existing, exists := qualityMap[qLabel]
			existingScore := int64(0)
			if exists {
				existingScore = int64(existing.Height)*10000000 + int64(existing.FPS)*100000 + existing.FileSize/1024
			}

			if !exists || candidateScore > existingScore {
				container := f.Ext
				if container == "" || container == "webm" {
					container = "mp4"
				}
				qualityMap[qLabel] = VideoFormatInfo{
					FormatID:  f.FormatID,
					Container: container,
					Quality:   qLabel,
					Width:     f.Width,
					Height:    f.Height,
					FPS:       int(f.FPS),
					HasVideo:  true,
					HasAudio:  hasA,
					AudioOnly: false,
					FileSize:  size,
				}
			}
		}
	}

	// Convert quality map to sorted slice (highest resolution first)
	order := []string{"4320p", "2160p", "1440p", "1080p", "720p", "480p", "360p", "240p", "144p"}
	var formats []VideoFormatInfo
	for _, q := range order {
		if fi, ok := qualityMap[q]; ok {
			formats = append(formats, fi)
		}
	}
	// Add remaining non-standard resolutions if any
	for q, fi := range qualityMap {
		isStandard := false
		for _, std := range order {
			if q == std {
				isStandard = true
				break
			}
		}
		if !isStandard {
			formats = append(formats, fi)
		}
	}

	// Add audio only option if audio exists
	if hasAudioFormats {
		formats = append(formats, VideoFormatInfo{
			FormatID:  "bestaudio",
			Container: "m4a",
			Quality:   "audio_only",
			HasVideo:  false,
			HasAudio:  true,
			AudioOnly: true,
		})
	}

	// Fallback format if list empty
	if len(formats) == 0 {
		formats = append(formats, VideoFormatInfo{
			FormatID:  "best",
			Container: "mp4",
			Quality:   "1080p",
			Width:     1920,
			Height:    1080,
			HasVideo:  true,
			HasAudio:  true,
		})
	}

	return &VideoAnalysisResult{
		Platform:    domain.PlatformYouTube,
		SourceURL:   canonicalURL,
		VideoID:     meta.ID,
		Title:       meta.Title,
		Description: meta.Description,
		Thumbnail:   meta.Thumbnail,
		Duration:    int(meta.Duration),
		DurationMs:  int64(meta.Duration * 1000),
		Author: AuthorInfo{
			Name: authorName,
		},
		Formats: formats,
	}, nil
}

func (d *YouTubeDownloader) Download(ctx context.Context, opt DownloadOptions, progress ProgressCallback) (*DownloadResult, error) {
	if _, err := ValidateYouTubeURL(opt.SourceURL); err != nil {
		return nil, err
	}

	// Ensure job temp directory exists: storage/temp/{jobId}/
	if err := os.MkdirAll(opt.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("create job temp dir: %w", err)
	}

	progress(DownloadProgress{
		Status:   "ANALYZING",
		Progress: 2,
	})

	// Build safe format selector
	formatSelector := buildSafeFormatSelector(opt.Quality)

	outputTemplate := filepath.Join(opt.TempDir, "video.%(ext)s")

	args := []string{
		"--no-playlist",
		"--newline",
		"--concurrent-fragments", "32",
		"--http-chunk-size", "10M",
		"--buffer-size", "32M",
		"--retries", "10",
		"--fragment-retries", "10",
		"--file-access-retries", "5",
		"--retry-sleep", "1",
		"--no-mtime",
		"--format", formatSelector,
		"--merge-output-format", "mp4",
		"--postprocessor-args", "Merger:-c copy -threads 0",
		"--output", outputTemplate,
	}

	// Supply ffmpeg location
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

	// Scanner to parse stdout line-by-line
	scanner := bufio.NewScanner(stdout)
	progressRegex := regexp.MustCompile(`\[download\]\s+([\d\.]+)%\s+of\s+~?([\d\.]+\w+)(?:\s+at\s+([\d\.]+\w+/s))?(?:\s+ETA\s+(\d+:\d+:\d+|\d+:\d+|\d+))?`)
	mergingRegex := regexp.MustCompile(`\[(Merger|ffmpeg)\]`)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()

			// Check for merging stage
			if mergingRegex.MatchString(line) {
				if lastProgressPct < 88 {
					lastProgressPct = 88
				}
				progress(DownloadProgress{
					Status:   "MERGING",
					Progress: lastProgressPct,
				})
				continue
			}

			// Check for download progress
			matches := progressRegex.FindStringSubmatch(line)
			if len(matches) >= 3 {
				if pctFloat, parseErr := strconv.ParseFloat(matches[1], 64); parseErr == nil {
					// Scale yt-dlp 0-100% to 5% - 85% range for DOWNLOADING
					scaled := 5 + int(pctFloat*0.80)
					if scaled > lastProgressPct && scaled <= 85 {
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

	// Read stderr in background to capture technical details in case of failure
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

	// Locate final file in temp directory (expected .mp4)
	expectedMP4 := filepath.Join(opt.TempDir, "video.mp4")
	if _, err := os.Stat(expectedMP4); os.IsNotExist(err) {
		// Look for any media file created in temp dir
		entries, _ := os.ReadDir(opt.TempDir)
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".mp4") || strings.HasSuffix(entry.Name(), ".m4a") || strings.HasSuffix(entry.Name(), ".webm")) {
				expectedMP4 = filepath.Join(opt.TempDir, entry.Name())
				break
			}
		}
	}

	// Verify output with ffprobe (Rule #17)
	requireAudio := opt.Quality != "video_only" && !strings.Contains(opt.FormatID, "video_only")
	verified, verifyErr := VerifyOutputFile(ctx, d.ffprobePath, expectedMP4, requireAudio)
	if verifyErr != nil {
		return nil, fmt.Errorf("verify output file: %w", verifyErr)
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

func buildSafeFormatSelector(quality string) string {
	switch strings.ToLower(strings.TrimSpace(quality)) {
	case "4320p", "8k":
		return "bestvideo[height<=?4320][format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=?4320][format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[height<=?4320][format_note!*=Premium][format_note!*=premium]+bestaudio/best[height<=?4320]/best"
	case "2160p", "4k":
		return "bestvideo[height<=?2160][format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=?2160][format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[height<=?2160][format_note!*=Premium][format_note!*=premium]+bestaudio/best[height<=?2160]/best"
	case "1440p", "2k":
		return "bestvideo[height<=?1440][format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=?1440][format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[height<=?1440][format_note!*=Premium][format_note!*=premium]+bestaudio/best[height<=?1440]/best"
	case "1080p":
		return "bestvideo[height<=?1080][format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=?1080][format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[height<=?1080][format_note!*=Premium][format_note!*=premium]+bestaudio/best[height<=?1080]/best"
	case "720p":
		return "bestvideo[height<=?720][format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=?720][format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[height<=?720][format_note!*=Premium][format_note!*=premium]+bestaudio/best[height<=?720]/best"
	case "480p":
		return "bestvideo[height<=?480][format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=?480][format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[height<=?480][format_note!*=Premium][format_note!*=premium]+bestaudio/best[height<=?480]/best"
	case "360p":
		return "bestvideo[height<=?360][format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[height<=?360][format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[height<=?360][format_note!*=Premium][format_note!*=premium]+bestaudio/best[height<=?360]/best"
	case "audio_only", "audio":
		return "bestaudio[ext=m4a]/bestaudio/best"
	case "original", "best", "max", "highest", "":
		return "bestvideo[format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[format_note!*=Premium][format_note!*=premium]+bestaudio/best"
	default:
		return "bestvideo[format_note!*=Premium][format_note!*=premium][ext=mp4]+bestaudio[ext=m4a]/bestvideo[format_note!*=Premium][format_note!*=premium]+bestaudio[ext=m4a]/bestvideo[format_note!*=Premium][format_note!*=premium]+bestaudio/best"
	}
}

func parseBytesString(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "~")
	re := regexp.MustCompile(`(?i)^([\d\.]+)\s*([a-z]+)$`)
	m := re.FindStringSubmatch(s)
	if len(m) < 3 {
		return 0
	}
	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}
	unit := strings.ToUpper(m[2])
	switch unit {
	case "B", "BYTES":
		return int64(val)
	case "KIB", "KB", "K":
		return int64(val * 1024)
	case "MIB", "MB", "M":
		return int64(val * 1024 * 1024)
	case "GIB", "GB", "G":
		return int64(val * 1024 * 1024 * 1024)
	case "TIB", "TB", "T":
		return int64(val * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(val)
	}
}

func parseSpeedString(s string) int64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "/s")
	s = strings.TrimSuffix(s, "/S")
	return parseBytesString(s)
}

func parseETAString(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ":")
	if len(parts) == 3 {
		h, _ := strconv.Atoi(parts[0])
		m, _ := strconv.Atoi(parts[1])
		sec, _ := strconv.Atoi(parts[2])
		return h*3600 + m*60 + sec
	} else if len(parts) == 2 {
		m, _ := strconv.Atoi(parts[0])
		sec, _ := strconv.Atoi(parts[1])
		return m*60 + sec
	} else if len(parts) == 1 {
		sec, _ := strconv.Atoi(parts[0])
		return sec
	}
	return 0
}

func normalizeYtDlpError(err error, rawOutput []byte) error {
	msg := strings.ToLower(string(rawOutput))
	switch {
	case strings.Contains(msg, "private video") || strings.Contains(msg, "this video is private"):
		return fmt.Errorf("VIDEO_PRIVATE: Video này ở chế độ riêng tư")
	case strings.Contains(msg, "video unavailable") || strings.Contains(msg, "is not available"):
		return fmt.Errorf("VIDEO_UNAVAILABLE: Video này không khả dụng hoặc đã bị gỡ")
	case strings.Contains(msg, "sign in") || strings.Contains(msg, "age-restricted") || strings.Contains(msg, "confirm your age"):
		return fmt.Errorf("AGE_RESTRICTED: Video yêu cầu đăng nhập hoặc giới hạn độ tuổi")
	case strings.Contains(msg, "too many requests") || strings.Contains(msg, "http error 429"):
		return fmt.Errorf("RATE_LIMITED: YouTube chặn tạm thời do quá nhiều yêu cầu")
	case strings.Contains(msg, "country") || strings.Contains(msg, "not available in your country"):
		return fmt.Errorf("REGION_RESTRICTED: Video bị giới hạn theo khu vực địa lý")
	default:
		return fmt.Errorf("DOWNLOAD_FAILED: Không thể tải video: %v", err)
	}
}
