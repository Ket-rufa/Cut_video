package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"social-video-cutter/api/internal/domain"
)

type VideoRepository interface {
	Create(ctx context.Context, video *domain.Video) error
	GetByID(ctx context.Context, id string) (*domain.Video, error)
	Update(ctx context.Context, video *domain.Video) error
	Delete(ctx context.Context, id string, userID string) error
	List(ctx context.Context, limit, offset int) ([]*domain.Video, error)
	ListFiltered(ctx context.Context, userID, platform, folderID, search string, limit, offset int) ([]*domain.Video, error)
}

type postgresVideoRepository struct {
	db *sqlx.DB
}

func NewVideoRepository(db *sqlx.DB) VideoRepository {
	return &postgresVideoRepository{db: db}
}

func (r *postgresVideoRepository) Create(ctx context.Context, video *domain.Video) error {
	if video.UserID == "" {
		video.UserID = "default_user"
	}
	query := `
		INSERT INTO videos (id, user_id, source_url, platform, external_id, title, description,
			thumbnail_url, duration_ms, width, height, source_path, status, folder_id, file_size,
			author_name, author_avatar, quality, mime_type, created_at, updated_at)
		VALUES (:id, :user_id, :source_url, :platform, :external_id, :title, :description,
			:thumbnail_url, :duration_ms, :width, :height, :source_path, :status, :folder_id, :file_size,
			:author_name, :author_avatar, :quality, :mime_type, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, video)
	return err
}

func (r *postgresVideoRepository) GetByID(ctx context.Context, id string) (*domain.Video, error) {
	video := &domain.Video{}
	err := r.db.GetContext(ctx, video, "SELECT * FROM videos WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("video not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (r *postgresVideoRepository) Update(ctx context.Context, video *domain.Video) error {
	video.UpdatedAt = time.Now()
	query := `
		UPDATE videos SET title=:title, description=:description, thumbnail_url=:thumbnail_url,
			duration_ms=:duration_ms, width=:width, height=:height,
			source_path=:source_path, status=:status, folder_id=:folder_id, file_size=:file_size,
			author_name=:author_name, author_avatar=:author_avatar, quality=:quality,
			mime_type=:mime_type, updated_at=:updated_at
		WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, query, video)
	return err
}

func (r *postgresVideoRepository) Delete(ctx context.Context, id string, userID string) error {
	if userID != "" {
		_, err := r.db.ExecContext(ctx, "DELETE FROM videos WHERE id = $1 AND user_id = $2", id, userID)
		return err
	}
	_, err := r.db.ExecContext(ctx, "DELETE FROM videos WHERE id = $1", id)
	return err
}

func (r *postgresVideoRepository) List(ctx context.Context, limit, offset int) ([]*domain.Video, error) {
	var videos []*domain.Video
	err := r.db.SelectContext(ctx, &videos,
		"SELECT * FROM videos ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return videos, err
}

func (r *postgresVideoRepository) ListFiltered(ctx context.Context, userID, platform, folderID, search string, limit, offset int) ([]*domain.Video, error) {
	query := "SELECT * FROM videos WHERE 1=1"
	var args []interface{}
	argIdx := 1

	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	}
	if platform != "" && platform != "all" {
		query += fmt.Sprintf(" AND platform = $%d", argIdx)
		args = append(args, platform)
		argIdx++
	}
	if folderID != "" && folderID != "all" && folderID != "f-all" {
		query += fmt.Sprintf(" AND folder_id = $%d", argIdx)
		args = append(args, folderID)
		argIdx++
	}
	if search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var videos []*domain.Video
	err := r.db.SelectContext(ctx, &videos, query, args...)
	return videos, err
}
