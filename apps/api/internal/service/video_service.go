package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/platform"
	"social-video-cutter/api/internal/repository"
)

type VideoService struct {
	videoRepo repository.VideoRepository
	resolver  *platform.Resolver
}

func NewVideoService(videoRepo repository.VideoRepository, resolver *platform.Resolver) *VideoService {
	return &VideoService{videoRepo: videoRepo, resolver: resolver}
}

func (s *VideoService) Resolve(ctx context.Context, rawURL string, userID string) (*domain.Video, error) {
	info, plat, err := s.resolver.Resolve(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("resolve URL: %w", err)
	}
	now := time.Now()
	if userID == "" {
		userID = "default_user"
	}
	video := &domain.Video{
		ID:           uuid.NewString(),
		UserID:       userID,
		SourceURL:    rawURL,
		Platform:     plat,
		ExternalID:   info.ExternalID,
		Title:        info.Title,
		Description:  info.Description,
		ThumbnailURL: info.ThumbnailURL,
		DurationMs:   info.DurationMs,
		Width:        info.Width,
		Height:       info.Height,
		SourcePath:   info.StreamURL,
		Status:       domain.VideoStatusResolved,
		FolderID:     "f-all",
		FileSize:     45000000,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.videoRepo.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("save video: %w", err)
	}
	return video, nil
}

func (s *VideoService) GetByID(ctx context.Context, id string) (*domain.Video, error) {
	return s.videoRepo.GetByID(ctx, id)
}

func (s *VideoService) List(ctx context.Context, limit, offset int) ([]*domain.Video, error) {
	return s.videoRepo.List(ctx, limit, offset)
}

func (s *VideoService) ListFiltered(ctx context.Context, userID, platform, folderID, search string, limit, offset int) ([]*domain.Video, error) {
	return s.videoRepo.ListFiltered(ctx, userID, platform, folderID, search, limit, offset)
}

func (s *VideoService) Update(ctx context.Context, video *domain.Video) error {
	return s.videoRepo.Update(ctx, video)
}

func (s *VideoService) Delete(ctx context.Context, id string, userID string) error {
	return s.videoRepo.Delete(ctx, id, userID)
}
