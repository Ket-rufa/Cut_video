package platform

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"social-video-cutter/api/internal/domain"
	"social-video-cutter/api/internal/platform/douyin"
	"social-video-cutter/api/internal/platform/facebook"
	"social-video-cutter/api/internal/platform/generic"
	"social-video-cutter/api/internal/platform/instagram"
	"social-video-cutter/api/internal/platform/tiktok"
	"social-video-cutter/api/internal/platform/youtube"
)

// Resolver selects the right provider for a URL and resolves video info.
type Resolver struct {
	providers []domain.VideoProvider
}

func NewResolver() *Resolver {
	return &Resolver{
		providers: []domain.VideoProvider{
			youtube.NewProvider(),
			tiktok.NewProvider(),
			douyin.NewProvider(),
			facebook.NewProvider(),
			instagram.NewProvider(),
			generic.NewProvider(), // must be last (catch-all)
		},
	}
}

// ValidateURL checks that the URL is safe and well-formed (anti-SSRF).
func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported scheme %q (only http/https allowed)", u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return fmt.Errorf("URL has no host")
	}
	// Block SSRF: private IP ranges, localhost
	blocked := []string{
		"localhost", "127.", "0.", "::1",
		"192.168.", "10.", "172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.",
		"172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31.",
		"169.254.", "0.0.0.0",
	}
	for _, b := range blocked {
		if strings.HasPrefix(host, b) || host == strings.TrimSuffix(b, ".") {
			return fmt.Errorf("blocked host: %s", host)
		}
	}
	return nil
}

// Resolve validates URL, selects provider, and returns VideoInfo + platform.
func (r *Resolver) Resolve(ctx context.Context, rawURL string) (*domain.VideoInfo, domain.VideoPlatform, error) {
	if err := ValidateURL(rawURL); err != nil {
		return nil, domain.PlatformUnknown, err
	}
	for _, p := range r.providers {
		if p.Supports(rawURL) {
			info, err := p.Resolve(ctx, rawURL)
			if err != nil {
				return nil, domain.PlatformUnknown, fmt.Errorf("provider resolve: %w", err)
			}
			return info, info.Platform, nil
		}
	}
	return nil, domain.PlatformUnknown, fmt.Errorf("no provider for URL: %s", rawURL)
}

// GetDownloadSource returns the best download source for a URL.
func (r *Resolver) GetDownloadSource(ctx context.Context, rawURL string) (*domain.MediaSource, error) {
	if err := ValidateURL(rawURL); err != nil {
		return nil, err
	}
	for _, p := range r.providers {
		if p.Supports(rawURL) {
			return p.GetDownloadSource(ctx, rawURL)
		}
	}
	return nil, fmt.Errorf("no provider for URL: %s", rawURL)
}
