package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"social-video-cutter/api/internal/domain"
)

type JobRepository interface {
	Create(ctx context.Context, job *domain.VideoJob) error
	GetByID(ctx context.Context, id string) (*domain.VideoJob, error)
	Update(ctx context.Context, job *domain.VideoJob) error
	ListByVideoID(ctx context.Context, videoID string) ([]*domain.VideoJob, error)
	ListAll(ctx context.Context, limit, offset int) ([]*domain.VideoJob, error)
	CreateExport(ctx context.Context, export *domain.VideoExport) error
	GetExportByJobID(ctx context.Context, jobID string) (*domain.VideoExport, error)
	GetExportByID(ctx context.Context, id string) (*domain.VideoExport, error)
}

type postgresJobRepository struct {
	db *sqlx.DB
}

func NewJobRepository(db *sqlx.DB) JobRepository {
	return &postgresJobRepository{db: db}
}

func (r *postgresJobRepository) Create(ctx context.Context, job *domain.VideoJob) error {
	query := `
		INSERT INTO video_jobs (id, video_id, type, status, progress, payload, error_message, created_at, updated_at)
		VALUES (:id, :video_id, :type, :status, :progress, :payload, :error_message, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, job)
	return err
}

func (r *postgresJobRepository) GetByID(ctx context.Context, id string) (*domain.VideoJob, error) {
	job := &domain.VideoJob{}
	err := r.db.GetContext(ctx, job, "SELECT * FROM video_jobs WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (r *postgresJobRepository) Update(ctx context.Context, job *domain.VideoJob) error {
	job.UpdatedAt = time.Now()
	query := `
		UPDATE video_jobs SET status=:status, progress=:progress, error_message=:error_message,
			started_at=:started_at, completed_at=:completed_at, updated_at=:updated_at
		WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, query, job)
	return err
}

func (r *postgresJobRepository) ListByVideoID(ctx context.Context, videoID string) ([]*domain.VideoJob, error) {
	var jobs []*domain.VideoJob
	err := r.db.SelectContext(ctx, &jobs,
		"SELECT * FROM video_jobs WHERE video_id = $1 ORDER BY created_at DESC", videoID)
	return jobs, err
}

func (r *postgresJobRepository) ListAll(ctx context.Context, limit, offset int) ([]*domain.VideoJob, error) {
	var jobs []*domain.VideoJob
	err := r.db.SelectContext(ctx, &jobs,
		"SELECT * FROM video_jobs ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return jobs, err
}

func (r *postgresJobRepository) CreateExport(ctx context.Context, export *domain.VideoExport) error {
	query := `
		INSERT INTO video_exports (id, video_id, job_id, start_ms, end_ms, width, height,
			format, quality, file_size, output_path, expires_at, created_at)
		VALUES (:id, :video_id, :job_id, :start_ms, :end_ms, :width, :height,
			:format, :quality, :file_size, :output_path, :expires_at, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, export)
	return err
}

func (r *postgresJobRepository) GetExportByJobID(ctx context.Context, jobID string) (*domain.VideoExport, error) {
	export := &domain.VideoExport{}
	err := r.db.GetContext(ctx, export, "SELECT * FROM video_exports WHERE job_id = $1", jobID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export not found for job: %s", jobID)
	}
	if err != nil {
		return nil, err
	}
	return export, nil
}

func (r *postgresJobRepository) GetExportByID(ctx context.Context, id string) (*domain.VideoExport, error) {
	export := &domain.VideoExport{}
	err := r.db.GetContext(ctx, export, "SELECT * FROM video_exports WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("export not found: %s", id)
	}
	if err != nil {
		return nil, err
	}
	return export, nil
}
