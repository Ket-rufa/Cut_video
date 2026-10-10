// Package generic handles any http/https video URL as a fallback provider.
package generic

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"social-video-cutter/api/internal/domain"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

// Supports returns true for any http/https URL (catch-all).
func (p *Provider) Supports(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	s := strings.ToLower(u.Scheme)
	return s == "http" || s == "https"
}

func (p *Provider) Resolve(ctx context.Context, rawURL string) (*domain.VideoInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	base := filepath.Base(u.Path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if name == "" || name == "." {
		name = u.Hostname()
	}
	return &domain.VideoInfo{
		Platform:     domain.PlatformGeneric,
		ExternalID:   rawURL,
		Title:        fmt.Sprintf("Video: %s", name),
		Description:  "Generic video URL",
		ThumbnailURL: "",
		DurationMs:   0,
		Width:        0,
		Height:       0,
		StreamURL:    rawURL,
	}, nil
}

func (p *Provider) GetDownloadSource(ctx context.Context, rawURL string) (*domain.MediaSource, error) {
	return &domain.MediaSource{URL: rawURL, Format: "mp4"}, nil
}
