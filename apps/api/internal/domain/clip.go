package domain

import "time"

type Clip struct {
	ID          string    `db:"id"            json:"id"`
	VideoID     string    `db:"video_id"      json:"video_id"`
	Title       string    `db:"title"         json:"title"`
	StartTimeMs int64     `db:"start_time_ms" json:"start_time_ms"`
	EndTimeMs   int64     `db:"end_time_ms"   json:"end_time_ms"`
	DurationMs  int64     `db:"duration_ms"   json:"duration_ms"`
	FilePath    string    `db:"file_path"     json:"file_path,omitempty"`
	Thumbnail   string    `db:"thumbnail"     json:"thumbnail,omitempty"`
	Status      string    `db:"status"        json:"status"`
	CreatedAt   time.Time `db:"created_at"    json:"created_at"`
}

type Folder struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Icon      string    `db:"icon"       json:"icon"`
	Color     string    `db:"color"      json:"color"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Tag struct {
	ID        string    `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Color     string    `db:"color"      json:"color"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type HistoryLog struct {
	ID         string    `db:"id"          json:"id"`
	VideoID    string    `db:"video_id"    json:"video_id"`
	VideoTitle string    `db:"video_title" json:"video_title"`
	Action     string    `db:"action"      json:"action"`
	ClipName   string    `db:"clip_name"   json:"clip_name"`
	Status     string    `db:"status"      json:"status"`
	FileSize   int64     `db:"file_size"   json:"file_size"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}
