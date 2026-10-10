package domain

import "time"

type ExportFormat string

const (
	FormatMP4  ExportFormat = "mp4"
	FormatWebM ExportFormat = "webm"
)

type ExportQuality string

const (
	QualityOriginal ExportQuality = "original"
	Quality1080p    ExportQuality = "1080p"
	Quality720p     ExportQuality = "720p"
	Quality480p     ExportQuality = "480p"
)

type VideoExport struct {
	ID         string        `db:"id"          json:"id"`
	VideoID    string        `db:"video_id"    json:"video_id"`
	JobID      string        `db:"job_id"      json:"job_id"`
	StartMs    int64         `db:"start_ms"    json:"start_ms"`
	EndMs      int64         `db:"end_ms"      json:"end_ms"`
	Width      int           `db:"width"       json:"width"`
	Height     int           `db:"height"      json:"height"`
	Format     ExportFormat  `db:"format"      json:"format"`
	Quality    ExportQuality `db:"quality"     json:"quality"`
	FileSize   int64         `db:"file_size"   json:"file_size"`
	OutputPath string        `db:"output_path" json:"output_path"`
	ExpiresAt  time.Time     `db:"expires_at"  json:"expires_at"`
	CreatedAt  time.Time     `db:"created_at"  json:"created_at"`
}
