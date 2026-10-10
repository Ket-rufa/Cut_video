CREATE TABLE IF NOT EXISTS video_subtitles (
    id TEXT PRIMARY KEY,
    video_id TEXT NOT NULL,
    language_code VARCHAR(10) NOT NULL DEFAULT 'vi',
    source_language VARCHAR(50) NOT NULL DEFAULT 'zh',
    format VARCHAR(10) NOT NULL DEFAULT 'srt',
    content TEXT NOT NULL DEFAULT '',
    cues_json TEXT NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_video_subtitles_vid ON video_subtitles(video_id);

CREATE TABLE IF NOT EXISTS app_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
