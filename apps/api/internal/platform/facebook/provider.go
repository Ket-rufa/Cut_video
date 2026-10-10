// Package facebook handles Facebook video URLs.
//
// LIMITATION: Facebook Graph API requires app credentials.
// This implementation returns mock metadata.
package facebook

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
	return strings.Contains(host, "facebook.com") ||
		strings.Contains(host, "fb.com") ||
		strings.Contains(host, "fb.watch")
}

func (p *Provider) Resolve(ctx context.Context, rawURL string) (*domain.VideoInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Facebook URL: %w", err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	externalID := rawURL
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			externalID = part
		}
	}
	return &domain.VideoInfo{
		Platform:     domain.PlatformFacebook,
		ExternalID:   externalID,
		Title:        "Facebook Video",
		Description:  "Facebook video — full metadata requires Graph API credentials",
		ThumbnailURL: "",
		DurationMs:   60000,
		Width:        1280,
		Height:       720,
	}, nil
}

func (p *Provider) GetDownloadSource(ctx context.Context, rawURL string) (*domain.MediaSource, error) {
	return &domain.MediaSource{URL: rawURL, Format: "mp4"}, nil
}
