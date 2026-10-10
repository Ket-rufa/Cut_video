package domain

import (
	"database/sql"
	"encoding/json"
	"time"
)

type JobType string

const (
	JobTypeDownload  JobType = "download"
	JobTypeTrim      JobType = "trim"
	JobTypeCrop      JobType = "crop"
	JobTypeConvert   JobType = "convert"
	JobTypeExport    JobType = "export"
	JobTypeThumbnail JobType = "thumbnail"
)

type JobStatus string

const (
	JobStatusQueued     JobStatus = "queued"
	JobStatusAnalyzing  JobStatus = "analyzing"
	JobStatusDownloading JobStatus = "downloading"
	JobStatusMerging    JobStatus = "merging"
	JobStatusSaving     JobStatus = "saving"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

type VideoJob struct {
	ID              string          `db:"id"               json:"id"`
	UserID          string          `db:"user_id"          json:"user_id"`
	VideoID         sql.NullString  `db:"video_id"         json:"video_id,omitempty"`
	Platform        string          `db:"platform"         json:"platform,omitempty"`
	SourceURL       string          `db:"source_url"       json:"source_url,omitempty"`
	Quality         string          `db:"quality"          json:"quality,omitempty"`
	Type            JobType         `db:"type"             json:"type"`
	Status          JobStatus       `db:"status"           json:"status"`
	Progress        int             `db:"progress"         json:"progress"`
	DownloadedBytes int64           `db:"downloaded_bytes" json:"downloaded_bytes"`
	TotalBytes      int64           `db:"total_bytes"      json:"total_bytes"`
	Speed           int64           `db:"speed"            json:"speed"`
	ETA             int             `db:"eta"              json:"eta"`
	Payload         json.RawMessage `db:"payload"          json:"payload"`
	ErrorCode       string          `db:"error_code"       json:"error_code,omitempty"`
	ErrorMessage    string          `db:"error_message"    json:"error_message,omitempty"`
	StartedAt       *time.Time      `db:"started_at"       json:"started_at,omitempty"`
	CompletedAt     *time.Time      `db:"completed_at"     json:"completed_at,omitempty"`
	CreatedAt       time.Time       `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"       json:"updated_at"`
}

type TrimPayload struct {
	StartMs   int64  `json:"start_ms"`
	EndMs     int64  `json:"end_ms"`
	Format    string `json:"format"`
	Quality   string `json:"quality"`
	VideoPath string `json:"video_path,omitempty"`
}

type DownloadPayload struct {
	SourceURL string `json:"source_url"`
	Platform  string `json:"platform"`
	VideoID   string `json:"video_id"`
	FormatID  string `json:"format_id"`
	Quality   string `json:"quality"`
}

type JobMessage struct {
	JobID   string `json:"job_id"`
	VideoID string `json:"video_id"`
	Type    string `json:"type"`
}
