// Package youtube handles YouTube URLs.
//
// LIMITATION: Full YouTube Data API v3 integration requires an API key and
// OAuth credentials. This implementation returns mock/stub metadata so the
// entire architecture can run end-to-end. Replace Resolve() with a real
// YouTube Data API v3 call when credentials are available.
// Direct video download from YouTube is not supported due to platform ToS.
package youtube

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
	return strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be")
}

func extractYouTubeID(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, "youtu.be") {
		id := strings.TrimPrefix(u.Path, "/")
		if id == "" {
			return "", fmt.Errorf("empty video ID in youtu.be URL")
		}
		return id, nil
	}
	v := u.Query().Get("v")
	if v == "" {
		return "", fmt.Errorf("could not extract video ID from URL: %s", rawURL)
	}
	return v, nil
}

func (p *Provider) Resolve(ctx context.Context, rawURL string) (*domain.VideoInfo, error) {
	id, err := extractYouTubeID(rawURL)
	if err != nil {
		return nil, err
	}
	// MOCK: Replace with YouTube Data API v3 or yt-dlp integration.
	return &domain.VideoInfo{
		Platform:     domain.PlatformYouTube,
		ExternalID:   id,
		Title:        fmt.Sprintf("YouTube Video (%s)", id),
		Description:  "YouTube video — full metadata requires YouTube Data API v3 key",
		ThumbnailURL: fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", id),
		DurationMs:   60000,
		Width:        1920,
		Height:       1080,
	}, nil
}

func (p *Provider) GetDownloadSource(ctx context.Context, rawURL string) (*domain.MediaSource, error) {
	// LIMITATION: YouTube restricts programmatic download per ToS.
	// Use yt-dlp with appropriate permissions for actual download.
	return &domain.MediaSource{
		URL:    rawURL,
		Format: "mp4",
	}, nil
}
