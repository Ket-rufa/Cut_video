package downloader

import (
	"fmt"
	"sync"
)

// Registry manages the collection of platform adapters.
type Registry struct {
	mu          sync.RWMutex
	downloaders []VideoDownloader
}

func NewRegistry(adapters ...VideoDownloader) *Registry {
	return &Registry{
		downloaders: adapters,
	}
}

// Register adds a new platform adapter to the registry.
func (r *Registry) Register(d VideoDownloader) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.downloaders = append(r.downloaders, d)
}

// FindDownloader returns the first adapter that supports the given raw URL.
func (r *Registry) FindDownloader(rawURL string) (VideoDownloader, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, d := range r.downloaders {
		if d.Supports(rawURL) {
			return d, nil
		}
	}

	return nil, fmt.Errorf("%w: no downloader found for URL %s", ErrUnsupportedPlatform, rawURL)
}
