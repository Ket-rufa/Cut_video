-- 000001_init_schema.up.sql
CREATE TABLE IF NOT EXISTS videos (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL DEFAULT 'default_user',
    source_url TEXT NOT NULL,
    platform VARCHAR(50) NOT NULL DEFAULT 'unknown',
    external_id VARCHAR(255) NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    duration_ms BIGINT NOT NULL DEFAULT 0,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    source_path TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    folder_id VARCHAR(50) NOT NULL DEFAULT 'f-all',
    file_size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id);
CREATE INDEX IF NOT EXISTS idx_videos_platform ON videos(platform);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);
CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos(created_at DESC);

CREATE TABLE IF NOT EXISTS video_jobs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL DEFAULT 'default_user',
    video_id VARCHAR(36) NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'queued',
    progress INTEGER NOT NULL DEFAULT 0,
    payload JSONB NOT NULL DEFAULT '{}',
    error_message TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_video_jobs_user_id ON video_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_video_jobs_video_id ON video_jobs(video_id);
CREATE INDEX IF NOT EXISTS idx_video_jobs_status ON video_jobs(status);

CREATE TABLE IF NOT EXISTS video_exports (
    id VARCHAR(36) PRIMARY KEY,
    video_id VARCHAR(36) NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    job_id VARCHAR(36) NOT NULL REFERENCES video_jobs(id) ON DELETE CASCADE,
    start_ms BIGINT NOT NULL DEFAULT 0,
    end_ms BIGINT NOT NULL DEFAULT 0,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    format VARCHAR(20) NOT NULL DEFAULT 'mp4',
    quality VARCHAR(20) NOT NULL DEFAULT 'original',
    file_size BIGINT NOT NULL DEFAULT 0,
    output_path TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_video_exports_video_id ON video_exports(video_id);
CREATE INDEX IF NOT EXISTS idx_video_exports_job_id ON video_exports(job_id);
CREATE INDEX IF NOT EXISTS idx_video_exports_expires_at ON video_exports(expires_at);

CREATE TABLE IF NOT EXISTS folders (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    icon VARCHAR(50) NOT NULL DEFAULT '📁',
    color VARCHAR(50) NOT NULL DEFAULT '#1677FF',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tags (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(50) NOT NULL DEFAULT '#1677FF',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS clips (
    id VARCHAR(36) PRIMARY KEY,
    video_id VARCHAR(36) NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    start_time_ms BIGINT NOT NULL DEFAULT 0,
    end_time_ms BIGINT NOT NULL DEFAULT 0,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    file_path TEXT NOT NULL DEFAULT '',
    thumbnail TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'ready',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clips_video_id ON clips(video_id);

CREATE TABLE IF NOT EXISTS history_logs (
    id VARCHAR(36) PRIMARY KEY,
    video_id VARCHAR(36) NOT NULL,
    video_title TEXT NOT NULL DEFAULT '',
    action VARCHAR(100) NOT NULL,
    clip_name TEXT NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'Đã cắt',
    file_size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_history_logs_created_at ON history_logs(created_at DESC);
