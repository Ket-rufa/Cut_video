package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"social-video-cutter/api/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SubtitleRepository interface {
	SaveSubtitle(ctx context.Context, sub *domain.VideoSubtitle) error
	GetByVideoID(ctx context.Context, videoID string) (*domain.VideoSubtitle, error)
	DeleteByVideoID(ctx context.Context, videoID string) error
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, val string) error
}

type subtitleRepository struct {
	db *sqlx.DB
}

func NewSubtitleRepository(db *sqlx.DB) SubtitleRepository {
	return &subtitleRepository{db: db}
}

func (r *subtitleRepository) SaveSubtitle(ctx context.Context, sub *domain.VideoSubtitle) error {
	if sub.ID == "" {
		existing, _ := r.GetByVideoID(ctx, sub.VideoID)
		if existing != nil && existing.ID != "" {
			sub.ID = existing.ID
		} else {
			sub.ID = uuid.NewString()
		}
	}
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now()
	}
	sub.UpdatedAt = time.Now()

	cuesBytes, err := json.Marshal(sub.Cues)
	if err == nil {
		sub.CuesJSON = string(cuesBytes)
	}

	query := `INSERT INTO video_subtitles (
		id, video_id, language_code, source_language, format, content, cues_json, created_at, updated_at
	) VALUES (
		:id, :video_id, :language_code, :source_language, :format, :content, :cues_json, :created_at, :updated_at
	) ON CONFLICT (id) DO UPDATE SET
		content = EXCLUDED.content,
		cues_json = EXCLUDED.cues_json,
		source_language = EXCLUDED.source_language,
		language_code = EXCLUDED.language_code,
		updated_at = EXCLUDED.updated_at`

	_, err = r.db.NamedExecContext(ctx, query, sub)
	return err
}

func (r *subtitleRepository) GetByVideoID(ctx context.Context, videoID string) (*domain.VideoSubtitle, error) {
	var sub domain.VideoSubtitle
	query := "SELECT * FROM video_subtitles WHERE video_id = $1 ORDER BY updated_at DESC LIMIT 1"
	if r.db.DriverName() == "sqlite" {
		query = "SELECT * FROM video_subtitles WHERE video_id = ? ORDER BY updated_at DESC LIMIT 1"
	}
	err := r.db.GetContext(ctx, &sub, query, videoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if sub.CuesJSON != "" {
		_ = json.Unmarshal([]byte(sub.CuesJSON), &sub.Cues)
	}
	return &sub, nil
}

func (r *subtitleRepository) DeleteByVideoID(ctx context.Context, videoID string) error {
	query := "DELETE FROM video_subtitles WHERE video_id = $1"
	if r.db.DriverName() == "sqlite" {
		query = "DELETE FROM video_subtitles WHERE video_id = ?"
	}
	_, err := r.db.ExecContext(ctx, query, videoID)
	return err
}

func (r *subtitleRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var val string
	query := "SELECT value FROM app_settings WHERE key = $1"
	if r.db.DriverName() == "sqlite" {
		query = "SELECT value FROM app_settings WHERE key = ?"
	}
	err := r.db.GetContext(ctx, &val, query, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func (r *subtitleRepository) SetSetting(ctx context.Context, key, val string) error {
	now := time.Now()
	query := `INSERT INTO app_settings (key, value, updated_at) VALUES ($1, $2, $3)`
	if r.db.DriverName() != "sqlite" {
		query += ` ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`
	} else {
		query = `INSERT OR REPLACE INTO app_settings (key, value, updated_at) VALUES (?, ?, ?)`
	}
	_, err := r.db.ExecContext(ctx, query, key, val, now)
	return err
}
