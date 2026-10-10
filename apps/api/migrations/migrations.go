package migrations

import (
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

//go:embed *.sql
var migrationFS embed.FS

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS videos (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL DEFAULT 'default_user',
    source_url TEXT NOT NULL,
    platform TEXT NOT NULL DEFAULT 'unknown',
    external_id TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    thumbnail_url TEXT NOT NULL DEFAULT '',
    duration_ms INTEGER NOT NULL DEFAULT 0,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    source_path TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    folder_id TEXT NOT NULL DEFAULT 'f-all',
    file_size INTEGER NOT NULL DEFAULT 0,
    author_name TEXT NOT NULL DEFAULT '',
    author_avatar TEXT NOT NULL DEFAULT '',
    quality TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    metadata_json TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS video_jobs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL DEFAULT 'default_user',
    video_id TEXT,
    platform TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    quality TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    progress INTEGER NOT NULL DEFAULT 0,
    downloaded_bytes INTEGER NOT NULL DEFAULT 0,
    total_bytes INTEGER NOT NULL DEFAULT 0,
    speed INTEGER NOT NULL DEFAULT 0,
    eta INTEGER NOT NULL DEFAULT 0,
    error_code TEXT NOT NULL DEFAULT '',
    payload TEXT NOT NULL DEFAULT '{}',
    error_message TEXT NOT NULL DEFAULT '',
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS video_exports (
    id TEXT PRIMARY KEY,
    video_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    start_ms INTEGER NOT NULL DEFAULT 0,
    end_ms INTEGER NOT NULL DEFAULT 0,
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    format TEXT NOT NULL DEFAULT 'mp4',
    quality TEXT NOT NULL DEFAULT 'original',
    file_size INTEGER NOT NULL DEFAULT 0,
    output_path TEXT NOT NULL DEFAULT '',
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS folders (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    icon TEXT NOT NULL DEFAULT '📁',
    color TEXT NOT NULL DEFAULT '#1677FF',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    color TEXT NOT NULL DEFAULT '#1677FF',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS clips (
    id TEXT PRIMARY KEY,
    video_id TEXT NOT NULL,
    title TEXT NOT NULL,
    start_time_ms INTEGER NOT NULL DEFAULT 0,
    end_time_ms INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    file_path TEXT NOT NULL DEFAULT '',
    thumbnail TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'ready',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS history_logs (
    id TEXT PRIMARY KEY,
    video_id TEXT NOT NULL,
    video_title TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL,
    clip_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'Đã cắt',
    file_size INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS video_subtitles (
    id TEXT PRIMARY KEY,
    video_id TEXT NOT NULL,
    language_code TEXT NOT NULL DEFAULT 'vi',
    source_language TEXT NOT NULL DEFAULT 'zh',
    format TEXT NOT NULL DEFAULT 'srt',
    content TEXT NOT NULL DEFAULT '',
    cues_json TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

// RunMigrations discovers and executes all .up.sql migration files or sets up SQLite schema.
func RunMigrations(db *sqlx.DB) error {
	driver := db.DriverName()
	if driver == "sqlite" {
		if _, err := db.Exec(sqliteSchema); err != nil {
			return fmt.Errorf("execute SQLite schema: %w", err)
		}
		log.Println("SQLite database schema initialized successfully")
		return nil
	}

	entries, err := migrationFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	// Ensure schema_migrations table exists to track applied versions
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, file := range upFiles {
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM schema_migrations WHERE version = $1", file)
		if err != nil {
			return fmt.Errorf("check migration status for %s: %w", file, err)
		}
		if count > 0 {
			continue // Already applied
		}

		content, err := migrationFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", file, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("execute migration %s failed: %w", file, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s failed: %w", file, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", file, err)
		}

		log.Printf("Applied migration: %s", file)
	}

	log.Println("Database migrations executed successfully")
	return nil
}
