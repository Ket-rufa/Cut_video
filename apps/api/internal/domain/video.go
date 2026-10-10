package domain

import "time"

type VideoPlatform string

const (
	PlatformYouTube   VideoPlatform = "youtube"
	PlatformTikTok    VideoPlatform = "tiktok"
	PlatformDouyin    VideoPlatform = "douyin"
	PlatformFacebook  VideoPlatform = "facebook"
	PlatformInstagram VideoPlatform = "instagram"
	PlatformTwitter   VideoPlatform = "twitter"
	PlatformGeneric   VideoPlatform = "generic"
	PlatformUnknown   VideoPlatform = "unknown"
)

type VideoStatus string

const (
	VideoStatusPending    VideoStatus = "pending"
	VideoStatusResolved   VideoStatus = "resolved"
	VideoStatusDownloaded VideoStatus = "downloaded"
	VideoStatusFailed     VideoStatus = "failed"
)

type Video struct {
	ID           string        `db:"id"            json:"id"`
	UserID       string        `db:"user_id"       json:"user_id"`
	SourceURL    string        `db:"source_url"    json:"source_url"`
	Platform     VideoPlatform `db:"platform"      json:"platform"`
	ExternalID   string        `db:"external_id"   json:"external_id"`
	Title        string        `db:"title"         json:"title"`
	Description  string        `db:"description"   json:"description"`
	ThumbnailURL string        `db:"thumbnail_url" json:"thumbnail_url"`
	DurationMs   int64         `db:"duration_ms"   json:"duration_ms"`
	Width        int           `db:"width"         json:"width"`
	Height       int           `db:"height"        json:"height"`
	SourcePath   string        `db:"source_path"   json:"source_path"`
	Status       VideoStatus   `db:"status"        json:"status"`
	FolderID     string        `db:"folder_id"     json:"folder_id"`
	FileSize     int64         `db:"file_size"     json:"file_size"`
	AuthorName   string        `db:"author_name"   json:"author_name"`
	AuthorAvatar string        `db:"author_avatar" json:"author_avatar"`
	Quality      string        `db:"quality"       json:"quality"`
	MimeType     string        `db:"mime_type"     json:"mime_type"`
	MetadataJSON string        `db:"metadata_json" json:"metadata_json,omitempty"`
	CreatedAt    time.Time     `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time     `db:"updated_at"    json:"updated_at"`
}

type VideoInfo struct {
	Platform     VideoPlatform `json:"platform"`
	ExternalID   string        `json:"external_id"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	ThumbnailURL string        `json:"thumbnail_url"`
	DurationMs   int64         `json:"duration_ms"`
	Width        int           `json:"width"`
	Height       int           `json:"height"`
	StreamURL    string        `json:"stream_url,omitempty"`
	Formats      []VideoFormat `json:"formats,omitempty"`
}

type VideoFormat struct {
	Quality  string `json:"quality"`
	URL      string `json:"url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	MimeType string `json:"mime_type"`
}

type MediaSource struct {
	URL     string            `json:"url"`
	Format  string            `json:"format"`
	Width   int               `json:"width"`
	Height  int               `json:"height"`
	Formats []VideoFormat     `json:"formats"`
	Headers map[string]string `json:"headers,omitempty"`
}
