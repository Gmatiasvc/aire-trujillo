package cache

import (
	"sync"
	"time"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

type heatmapEntry struct {
	data      *domain.HeatmapResponse
	expiresAt time.Time
}

type playbackEntry struct {
	timestamps []string
	expiresAt  time.Time
}

// MemoryCache provides thread-safe caching for computed heatmap grids and playback timestamps.
type MemoryCache struct {
	mu       sync.RWMutex
	ttl      time.Duration
	maxItems int
	heatmaps map[string]heatmapEntry
	playback *playbackEntry
}

// NewMemoryCache creates a MemoryCache with the specified TTL.
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	return &MemoryCache{
		ttl:      ttl,
		maxItems: 256,
		heatmaps: make(map[string]heatmapEntry),
	}
}

// GetHeatmap returns a cached HeatmapResponse if present and not expired.
func (c *MemoryCache) GetHeatmap(key string) (*domain.HeatmapResponse, bool) {
	c.mu.RLock()
	entry, ok := c.heatmaps[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

// SetHeatmap stores a HeatmapResponse in cache, evicting expired or oldest entries if full.
func (c *MemoryCache) SetHeatmap(key string, resp *domain.HeatmapResponse) {
	if resp == nil {
		return
	}
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.heatmaps) >= c.maxItems {
		for k, v := range c.heatmaps {
			if now.After(v.expiresAt) {
				delete(c.heatmaps, k)
			}
		}
		if len(c.heatmaps) >= c.maxItems {
			// Clear map to keep free-tier memory bounded and predictable.
			c.heatmaps = make(map[string]heatmapEntry)
		}
	}

	c.heatmaps[key] = heatmapEntry{
		data:      resp,
		expiresAt: now.Add(c.ttl),
	}
}

// GetPlaybackTimestamps returns cached playback timestamps if present and valid.
func (c *MemoryCache) GetPlaybackTimestamps() ([]string, bool) {
	c.mu.RLock()
	entry := c.playback
	c.mu.RUnlock()
	if entry == nil || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	out := make([]string, len(entry.timestamps))
	copy(out, entry.timestamps)
	return out, true
}

// SetPlaybackTimestamps stores playback timestamps in cache.
func (c *MemoryCache) SetPlaybackTimestamps(timestamps []string) {
	copied := make([]string, len(timestamps))
	copy(copied, timestamps)

	c.mu.Lock()
	c.playback = &playbackEntry{
		timestamps: copied,
		expiresAt:  time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// InvalidateAll clears all cached heatmap grids and playback timestamps.
func (c *MemoryCache) InvalidateAll() {
	c.mu.Lock()
	c.heatmaps = make(map[string]heatmapEntry)
	c.playback = nil
	c.mu.Unlock()
}
