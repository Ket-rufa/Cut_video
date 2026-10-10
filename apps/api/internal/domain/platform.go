package domain

import "context"

// VideoProvider is the interface every platform must implement.
type VideoProvider interface {
	// Supports returns true if this provider handles the given URL.
	Supports(url string) bool
	// Resolve fetches video metadata for the given URL.
	Resolve(ctx context.Context, url string) (*VideoInfo, error)
	// GetDownloadSource returns the playable/downloadable media source.
	GetDownloadSource(ctx context.Context, url string) (*MediaSource, error)
}
