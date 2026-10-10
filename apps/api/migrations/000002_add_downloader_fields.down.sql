-- 000002_add_downloader_fields.down.sql
ALTER TABLE video_jobs DROP COLUMN IF EXISTS error_code;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS eta;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS speed;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS total_bytes;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS downloaded_bytes;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS quality;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS source_url;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS platform;
ALTER TABLE video_jobs DROP COLUMN IF EXISTS user_id;

ALTER TABLE videos DROP COLUMN IF EXISTS metadata_json;
ALTER TABLE videos DROP COLUMN IF EXISTS mime_type;
ALTER TABLE videos DROP COLUMN IF EXISTS quality;
ALTER TABLE videos DROP COLUMN IF EXISTS author_avatar;
ALTER TABLE videos DROP COLUMN IF EXISTS author_name;
ALTER TABLE videos DROP COLUMN IF EXISTS user_id;
