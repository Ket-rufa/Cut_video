package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/queue"
	"social-video-cutter/api/internal/repository"
)

type ExportService struct {
	jobRepo  repository.JobRepository
	enqueuer queue.JobEnqueuer
}

func NewExportService(jobRepo repository.JobRepository, enqueuer queue.JobEnqueuer) *ExportService {
	return &ExportService{jobRepo: jobRepo, enqueuer: enqueuer}
}

type TrimRequest struct {
	VideoID string
	StartMs int64
	EndMs   int64
	Format  string
	Quality string
}

func (s *ExportService) CreateTrimJob(ctx context.Context, req TrimRequest) (*domain.VideoJob, error) {
	if req.StartMs < 0 {
		return nil, fmt.Errorf("start_ms must be >= 0")
	}
	if req.EndMs <= req.StartMs {
		return nil, fmt.Errorf("end_ms must be greater than start_ms")
	}
	if req.Format == "" {
		req.Format = "mp4"
	}
	if req.Quality == "" {
		req.Quality = "original"
	}

	payload := domain.TrimPayload{
		StartMs: req.StartMs,
		EndMs:   req.EndMs,
		Format:  req.Format,
		Quality: req.Quality,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	now := time.Now()
	job := &domain.VideoJob{
		ID:        uuid.NewString(),
		VideoID:   sql.NullString{String: req.VideoID, Valid: req.VideoID != ""},
		Type:      domain.JobTypeTrim,
		Status:    domain.JobStatusQueued,
		Progress:  0,
		Payload:   json.RawMessage(payloadBytes),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}
	if err := s.enqueuer.EnqueueJob(ctx, job); err != nil {
		return nil, fmt.Errorf("enqueue job: %w", err)
	}
	return job, nil
}

func (s *ExportService) GetJob(ctx context.Context, id string) (*domain.VideoJob, error) {
	return s.jobRepo.GetByID(ctx, id)
}

func (s *ExportService) GetExport(ctx context.Context, id string) (*domain.VideoExport, error) {
	return s.jobRepo.GetExportByID(ctx, id)
}

func (s *ExportService) GetExportByJobID(ctx context.Context, jobID string) (*domain.VideoExport, error) {
	return s.jobRepo.GetExportByJobID(ctx, jobID)
}

func (s *ExportService) ListJobs(ctx context.Context, limit, offset int) ([]*domain.VideoJob, error) {
	return s.jobRepo.ListAll(ctx, limit, offset)
}
