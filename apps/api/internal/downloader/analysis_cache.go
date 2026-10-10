package downloader

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type analysisEntry struct {
	result    *VideoAnalysisResult
	expiresAt time.Time
}

// AnalysisCache provides a thread-safe cache for analysis sessions with TTL.
type AnalysisCache struct {
	mu      sync.RWMutex
	entries map[string]analysisEntry
	ttl     time.Duration
}

func NewAnalysisCache(ttl time.Duration) *AnalysisCache {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	c := &AnalysisCache{
		entries: make(map[string]analysisEntry),
		ttl:     ttl,
	}

	// Background eviction ticker
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			c.mu.Lock()
			now := time.Now()
			for id, entry := range c.entries {
				if now.After(entry.expiresAt) {
					delete(c.entries, id)
				}
			}
			c.mu.Unlock()
		}
	}()

	return c
}

// Store saves the analysis result and returns a new unique analysisId.
func (c *AnalysisCache) Store(result *VideoAnalysisResult) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := uuid.NewString()
	result.AnalysisID = id
	c.entries[id] = analysisEntry{
		result:    result,
		expiresAt: time.Now().Add(c.ttl),
	}
	return id
}

// Get retrieves the cached analysis result by analysisId.
func (c *AnalysisCache) Get(analysisID string) (*VideoAnalysisResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[analysisID]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.result, true
}
