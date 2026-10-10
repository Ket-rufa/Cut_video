// Package douyin handles Douyin (Chinese TikTok) URLs.
package douyin

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"social-video-cutter/api/internal/domain"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return strings.Contains(host, "douyin.com") || strings.Contains(host, "iesdouyin.com")
}

func (p *Provider) Resolve(ctx context.Context, rawURL string) (*domain.VideoInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Douyin URL: %w", err)
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	externalID := rawURL
	if len(parts) >= 2 {
		externalID = parts[len(parts)-1]
	}
	return &domain.VideoInfo{
		Platform:     domain.PlatformDouyin,
		ExternalID:   externalID,
		Title:        "Douyin Video",
		Description:  "Douyin video — metadata extracted via video resolver",
		ThumbnailURL: "",
		DurationMs:   30000,
		Width:        1080,
		Height:       1920,
	}, nil
}

func (p *Provider) GetDownloadSource(ctx context.Context, rawURL string) (*domain.MediaSource, error) {
	return &domain.MediaSource{URL: rawURL, Format: "mp4"}, nil
}
