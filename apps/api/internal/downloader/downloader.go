package downloader

import (
	"context"
	"social-video-cutter/api/internal/domain"
)

// AuthorInfo holds details about the video creator.
type AuthorInfo struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

// VideoFormatInfo represents a normalized, user-friendly format.
type VideoFormatInfo struct {
	FormatID  string `json:"formatId"`
	Container string `json:"container"` // e.g. "mp4", "webm"
	Quality   string `json:"quality"`   // e.g. "1080p", "720p", "480p", "360p", "audio_only"
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	FPS       int    `json:"fps,omitempty"`
	HasVideo  bool   `json:"hasVideo"`
	HasAudio  bool   `json:"hasAudio"`
	AudioOnly bool   `json:"audioOnly"`
	FileSize  int64  `json:"fileSize,omitempty"` // in bytes
}

// VideoAnalysisResult contains structured metadata from video analysis.
type VideoAnalysisResult struct {
	AnalysisID  string            `json:"analysisId"`
	Platform    domain.VideoPlatform `json:"platform"`
	SourceURL   string            `json:"sourceUrl"`
	VideoID     string            `json:"videoId"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Thumbnail   string            `json:"thumbnail"`
	Duration    int               `json:"duration"` // duration in seconds
	DurationMs  int64             `json:"durationMs"`
	Author      AuthorInfo        `json:"author"`
	Formats     []VideoFormatInfo `json:"formats"`
}

// DownloadOptions holds options passed to download a video.
type DownloadOptions struct {
	JobID           string
	UserID          string
	SourceURL       string
	VideoID         string
	FormatID        string
	Quality         string
	DestinationPath string
	TempDir         string
}

// ProgressCallback reports real download & merge status updates.
type ProgressCallback func(p DownloadProgress)

// DownloadProgress represents the progress update event.
type DownloadProgress struct {
	Status          string `json:"status"` // QUEUED, ANALYZING, DOWNLOADING, MERGING, SAVING, COMPLETED, FAILED, CANCELLED
	Progress        int    `json:"progress"` // 0 to 100 monotonic
	DownloadedBytes int64  `json:"downloadedBytes"`
	TotalBytes      int64  `json:"totalBytes"`
	Speed           int64  `json:"speed"` // bytes/second
	ETA             int    `json:"eta"`   // remaining seconds
}

// DownloadResult contains the final verified video asset details.
type DownloadResult struct {
	FilePath     string
	FileName     string
	FileSize     int64
	MimeType     string
	Width        int
	Height       int
	DurationMs   int64
	ThumbnailURL string
	AuthorName   string
	AuthorAvatar string
	Title        string
	Description  string
}

// VideoDownloader is the interface every platform adapter must implement.
type VideoDownloader interface {
	// Supports checks if this downloader can handle the given raw URL.
	Supports(rawURL string) bool
	// Platform returns the domain platform enum.
	Platform() domain.VideoPlatform
	// CanonicalURL normalizes URL and resolves redirects if required (e.g. TikTok short URLs).
	CanonicalURL(ctx context.Context, rawURL string) (string, error)
	// Analyze fetches real metadata and formats without downloading media.
	Analyze(ctx context.Context, canonicalURL string) (*VideoAnalysisResult, error)
	// Download executes the real download, merges streams with FFmpeg if needed, and streams progress.
	Download(ctx context.Context, opt DownloadOptions, progress ProgressCallback) (*DownloadResult, error)
}
