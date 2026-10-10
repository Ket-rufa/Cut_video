// Package instagram handles Instagram video URLs.
//
// LIMITATION: Instagram Basic Display API requires app approval.
// This implementation returns mock metadata.
package instagram

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
	return strings.Contains(host, "instagram.com") || strings.Contains(host, "instagr.am")
}

func (p *Provider) Resolve(ctx context.Context, rawURL string) (*domain.VideoInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Instagram URL: %w", err)
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	externalID := rawURL
	if len(parts) >= 2 {
		externalID = parts[1]
	}
	return &domain.VideoInfo{
		Platform:     domain.PlatformInstagram,
		ExternalID:   externalID,
		Title:        "Instagram Video",
		Description:  "Instagram video — full metadata requires Basic Display API access",
		ThumbnailURL: "",
		DurationMs:   60000,
		Width:        1080,
		Height:       1080,
	}, nil
}

func (p *Provider) GetDownloadSource(ctx context.Context, rawURL string) (*domain.MediaSource, error) {
	return &domain.MediaSource{URL: rawURL, Format: "mp4"}, nil
}
