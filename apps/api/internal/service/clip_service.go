package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/repository"
)

type ClipService struct {
	clipRepo repository.ClipRepository
}

func NewClipService(clipRepo repository.ClipRepository) *ClipService {
	return &ClipService{clipRepo: clipRepo}
}

type CreateClipRequest struct {
	Title       string `json:"title" binding:"required"`
	StartTimeMs int64  `json:"start_time_ms"`
	EndTimeMs   int64  `json:"end_time_ms" binding:"required"`
	DurationMs  int64  `json:"duration_ms"`
	FilePath    string `json:"file_path"`
	Thumbnail   string `json:"thumbnail"`
}

func (s *ClipService) CreateClip(ctx context.Context, videoID string, req CreateClipRequest) (*domain.Clip, error) {
	duration := req.DurationMs
	if duration <= 0 {
		duration = req.EndTimeMs - req.StartTimeMs
	}
	clip := &domain.Clip{
		ID:          uuid.NewString(),
		VideoID:     videoID,
		Title:       req.Title,
		StartTimeMs: req.StartTimeMs,
		EndTimeMs:   req.EndTimeMs,
		DurationMs:  duration,
		FilePath:    req.FilePath,
		Thumbnail:   req.Thumbnail,
		Status:      "ready",
		CreatedAt:   time.Now(),
	}
	if err := s.clipRepo.Create(ctx, clip); err != nil {
		return nil, err
	}
	return clip, nil
}

func (s *ClipService) ListByVideoID(ctx context.Context, videoID string) ([]*domain.Clip, error) {
	return s.clipRepo.ListByVideoID(ctx, videoID)
}

func (s *ClipService) Delete(ctx context.Context, id string) error {
	return s.clipRepo.Delete(ctx, id)
}

type LibraryService struct {
	libRepo repository.LibraryRepository
}

func NewLibraryService(libRepo repository.LibraryRepository) *LibraryService {
	return &LibraryService{libRepo: libRepo}
}

func (s *LibraryService) ListFolders(ctx context.Context) ([]*domain.Folder, error) {
	return s.libRepo.ListFolders(ctx)
}

func (s *LibraryService) CreateFolder(ctx context.Context, name, icon, color string) (*domain.Folder, error) {
	f := &domain.Folder{
		ID:        uuid.NewString(),
		Name:      name,
		Icon:      icon,
		Color:     color,
		CreatedAt: time.Now(),
	}
	if err := s.libRepo.CreateFolder(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *LibraryService) DeleteFolder(ctx context.Context, id string) error {
	return s.libRepo.DeleteFolder(ctx, id)
}

func (s *LibraryService) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	return s.libRepo.ListTags(ctx)
}

func (s *LibraryService) CreateTag(ctx context.Context, name, color string) (*domain.Tag, error) {
	t := &domain.Tag{
		ID:        uuid.NewString(),
		Name:      name,
		Color:     color,
		CreatedAt: time.Now(),
	}
	if err := s.libRepo.CreateTag(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *LibraryService) DeleteTag(ctx context.Context, id string) error {
	return s.libRepo.DeleteTag(ctx, id)
}

func (s *LibraryService) ListHistory(ctx context.Context, limit, offset int) ([]*domain.HistoryLog, error) {
	return s.libRepo.ListHistory(ctx, limit, offset)
}

func (s *LibraryService) CreateHistory(ctx context.Context, log *domain.HistoryLog) error {
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	return s.libRepo.CreateHistory(ctx, log)
}

func (s *LibraryService) DeleteHistory(ctx context.Context, id string) error {
	return s.libRepo.DeleteHistory(ctx, id)
}
