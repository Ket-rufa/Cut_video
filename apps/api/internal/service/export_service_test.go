package service_test

import (
	"context"
	"fmt"
	"testing"

	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/service"
)

// --- Mock implementations ---

type mockJobRepo struct {
	jobs    []*domain.VideoJob
	exports []*domain.VideoExport
}

func (m *mockJobRepo) Create(ctx context.Context, job *domain.VideoJob) error {
	m.jobs = append(m.jobs, job)
	return nil
}
func (m *mockJobRepo) GetByID(ctx context.Context, id string) (*domain.VideoJob, error) {
	for _, j := range m.jobs {
		if j.ID == id {
			return j, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockJobRepo) Update(ctx context.Context, job *domain.VideoJob) error            { return nil }
func (m *mockJobRepo) ListByVideoID(ctx context.Context, videoID string) ([]*domain.VideoJob, error) {
	return nil, nil
}
func (m *mockJobRepo) ListAll(ctx context.Context, limit, offset int) ([]*domain.VideoJob, error) {
	return m.jobs, nil
}
func (m *mockJobRepo) CreateExport(ctx context.Context, e *domain.VideoExport) error {
	m.exports = append(m.exports, e)
	return nil
}
func (m *mockJobRepo) GetExportByJobID(ctx context.Context, jobID string) (*domain.VideoExport, error) {
	for _, e := range m.exports {
		if e.JobID == jobID {
			return e, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockJobRepo) GetExportByID(ctx context.Context, id string) (*domain.VideoExport, error) {
	for _, e := range m.exports {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

type mockEnqueuer struct {
	enqueued []*domain.VideoJob
}

func (m *mockEnqueuer) EnqueueJob(ctx context.Context, job *domain.VideoJob) error {
	m.enqueued = append(m.enqueued, job)
	return nil
}

// --- Tests ---

func TestCreateTrimJob_Success(t *testing.T) {
	repo := &mockJobRepo{}
	eq := &mockEnqueuer{}
	svc := service.NewExportService(repo, eq)

	job, err := svc.CreateTrimJob(context.Background(), service.TrimRequest{
		VideoID: "video-1",
		StartMs: 1000,
		EndMs:   5000,
		Format:  "mp4",
		Quality: "original",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.VideoID.String != "video-1" {
		t.Errorf("expected video-1, got %s", job.VideoID.String)
	}
	if job.Status != domain.JobStatusQueued {
		t.Errorf("expected queued, got %s", job.Status)
	}
	if len(eq.enqueued) != 1 {
		t.Errorf("expected 1 enqueued job, got %d", len(eq.enqueued))
	}
}

func TestCreateTrimJob_InvalidRange(t *testing.T) {
	repo := &mockJobRepo{}
	eq := &mockEnqueuer{}
	svc := service.NewExportService(repo, eq)

	_, err := svc.CreateTrimJob(context.Background(), service.TrimRequest{
		VideoID: "video-1",
		StartMs: 5000,
		EndMs:   3000, // end < start
	})
	if err == nil {
		t.Error("expected error for end < start")
	}
}

func TestCreateTrimJob_NegativeStart(t *testing.T) {
	repo := &mockJobRepo{}
	eq := &mockEnqueuer{}
	svc := service.NewExportService(repo, eq)

	_, err := svc.CreateTrimJob(context.Background(), service.TrimRequest{
		VideoID: "video-1",
		StartMs: -100,
		EndMs:   5000,
	})
	if err == nil {
		t.Error("expected error for negative start_ms")
	}
}
