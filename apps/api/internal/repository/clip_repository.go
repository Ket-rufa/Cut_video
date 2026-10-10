package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"social-video-cutter/api/internal/domain"
)

type ClipRepository interface {
	Create(ctx context.Context, clip *domain.Clip) error
	GetByID(ctx context.Context, id string) (*domain.Clip, error)
	ListByVideoID(ctx context.Context, videoID string) ([]*domain.Clip, error)
	Delete(ctx context.Context, id string) error
}

type LibraryRepository interface {
	CreateFolder(ctx context.Context, folder *domain.Folder) error
	ListFolders(ctx context.Context) ([]*domain.Folder, error)
	DeleteFolder(ctx context.Context, id string) error

	CreateTag(ctx context.Context, tag *domain.Tag) error
	ListTags(ctx context.Context) ([]*domain.Tag, error)
	DeleteTag(ctx context.Context, id string) error

	CreateHistory(ctx context.Context, log *domain.HistoryLog) error
	ListHistory(ctx context.Context, limit, offset int) ([]*domain.HistoryLog, error)
	DeleteHistory(ctx context.Context, id string) error
}

type postgresClipRepository struct {
	db *sqlx.DB
}

func NewClipRepository(db *sqlx.DB) ClipRepository {
	return &postgresClipRepository{db: db}
}

func (r *postgresClipRepository) Create(ctx context.Context, clip *domain.Clip) error {
	query := `
		INSERT INTO clips (id, video_id, title, start_time_ms, end_time_ms, duration_ms, file_path, thumbnail, status, created_at)
		VALUES (:id, :video_id, :title, :start_time_ms, :end_time_ms, :duration_ms, :file_path, :thumbnail, :status, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, clip)
	return err
}

func (r *postgresClipRepository) GetByID(ctx context.Context, id string) (*domain.Clip, error) {
	clip := &domain.Clip{}
	err := r.db.GetContext(ctx, clip, "SELECT * FROM clips WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("clip not found: %s", id)
	}
	return clip, err
}

func (r *postgresClipRepository) ListByVideoID(ctx context.Context, videoID string) ([]*domain.Clip, error) {
	var clips []*domain.Clip
	err := r.db.SelectContext(ctx, &clips, "SELECT * FROM clips WHERE video_id = $1 ORDER BY start_time_ms ASC", videoID)
	return clips, err
}

func (r *postgresClipRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM clips WHERE id = $1", id)
	return err
}

type postgresLibraryRepository struct {
	db *sqlx.DB
}

func NewLibraryRepository(db *sqlx.DB) LibraryRepository {
	return &postgresLibraryRepository{db: db}
}

func (r *postgresLibraryRepository) CreateFolder(ctx context.Context, f *domain.Folder) error {
	query := `INSERT INTO folders (id, name, icon, color, created_at) VALUES (:id, :name, :icon, :color, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, f)
	return err
}

func (r *postgresLibraryRepository) ListFolders(ctx context.Context) ([]*domain.Folder, error) {
	var folders []*domain.Folder
	err := r.db.SelectContext(ctx, &folders, "SELECT * FROM folders ORDER BY created_at ASC")
	return folders, err
}

func (r *postgresLibraryRepository) DeleteFolder(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM folders WHERE id = $1", id)
	return err
}

func (r *postgresLibraryRepository) CreateTag(ctx context.Context, t *domain.Tag) error {
	query := `INSERT INTO tags (id, name, color, created_at) VALUES (:id, :name, :color, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, t)
	return err
}

func (r *postgresLibraryRepository) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	var tags []*domain.Tag
	err := r.db.SelectContext(ctx, &tags, "SELECT * FROM tags ORDER BY created_at ASC")
	return tags, err
}

func (r *postgresLibraryRepository) DeleteTag(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tags WHERE id = $1", id)
	return err
}

func (r *postgresLibraryRepository) CreateHistory(ctx context.Context, l *domain.HistoryLog) error {
	query := `INSERT INTO history_logs (id, video_id, video_title, action, clip_name, status, file_size, created_at)
		VALUES (:id, :video_id, :video_title, :action, :clip_name, :status, :file_size, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, l)
	return err
}

func (r *postgresLibraryRepository) ListHistory(ctx context.Context, limit, offset int) ([]*domain.HistoryLog, error) {
	var logs []*domain.HistoryLog
	err := r.db.SelectContext(ctx, &logs, "SELECT * FROM history_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return logs, err
}

func (r *postgresLibraryRepository) DeleteHistory(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM history_logs WHERE id = $1", id)
	return err
}
