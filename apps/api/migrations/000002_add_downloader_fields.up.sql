-- 000002_add_downloader_fields.up.sql

-- Expand videos table
ALTER TABLE videos ADD COLUMN IF NOT EXISTS user_id VARCHAR(36) NOT NULL DEFAULT 'default_user';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS author_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS author_avatar TEXT NOT NULL DEFAULT '';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS quality VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS mime_type VARCHAR(100) NOT NULL DEFAULT 'video/mp4';
ALTER TABLE videos ADD COLUMN IF NOT EXISTS metadata_json JSONB NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_videos_user_external ON videos(user_id, platform, external_id);

-- Expand video_jobs table
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS user_id VARCHAR(36) NOT NULL DEFAULT 'default_user';
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS platform VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS source_url TEXT NOT NULL DEFAULT '';
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS quality VARCHAR(50) NOT NULL DEFAULT '';
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS downloaded_bytes BIGINT NOT NULL DEFAULT 0;
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS total_bytes BIGINT NOT NULL DEFAULT 0;
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS speed BIGINT NOT NULL DEFAULT 0;
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS eta INTEGER NOT NULL DEFAULT 0;
ALTER TABLE video_jobs ADD COLUMN IF NOT EXISTS error_code VARCHAR(100) NOT NULL DEFAULT '';

-- Allow video_id to be nullable for freshly queued download jobs before final video is saved
ALTER TABLE video_jobs ALTER COLUMN video_id DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_video_jobs_user_status ON video_jobs(user_id, status);
CREATE INDEX IF NOT EXISTS idx_video_jobs_created_at ON video_jobs(created_at DESC);
