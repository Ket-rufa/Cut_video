// Package tiktok handles TikTok URLs.
//
// LIMITATION: TikTok's API is restricted to approved partners.
// This implementation returns mock metadata. Replace with TikTok API when credentials available.
package tiktok

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
	return strings.Contains(host, "tiktok.com") || strings.Contains(host, "vm.tiktok.com")
}

func (p *Provider) Resolve(ctx context.Context, rawURL string) (*domain.VideoInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid TikTok URL: %w", err)
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	externalID := rawURL
	if len(parts) >= 3 {
		externalID = parts[2]
	}
	return &domain.VideoInfo{
		Platform:     domain.PlatformTikTok,
		ExternalID:   externalID,
		Title:        "TikTok Video",
		Description:  "TikTok video — full metadata requires TikTok API access",
		ThumbnailURL: "",
		DurationMs:   30000,
		Width:        1080,
		Height:       1920,
	}, nil
}

func (p *Provider) GetDownloadSource(ctx context.Context, rawURL string) (*domain.MediaSource, error) {
	return &domain.MediaSource{URL: rawURL, Format: "mp4"}, nil
}
