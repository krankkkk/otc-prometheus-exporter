package provider

import (
	"sync"
	"time"
)

type cacheEntry struct {
	names     map[string]string
	updatedAt time.Time
}

// NameCache stores the last known good name map per namespace.
// Used as a stale-on-error fallback when service APIs fail.
type NameCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

// NewNameCache creates an empty NameCache.
func NewNameCache() *NameCache {
	return &NameCache{entries: make(map[string]cacheEntry)}
}

// Get returns a shallow copy of the cached name map for the given namespace, or nil.
// The copy prevents callers from accidentally mutating cached data.
func (c *NameCache) Get(namespace string) map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[namespace]
	if !ok {
		return nil
	}
	cp := make(map[string]string, len(e.names))
	for k, v := range e.names {
		cp[k] = v
	}
	return cp
}

// GetAge returns how long ago the cache entry for the given namespace was updated.
// Returns 0 if the namespace has no cache entry.
func (c *NameCache) GetAge(namespace string) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[namespace]
	if !ok {
		return 0
	}
	return time.Since(e.updatedAt)
}

// Put stores a name map for the given namespace, replacing any previous entry.
// A nil map is a no-op (does not clear the cache).
func (c *NameCache) Put(namespace string, names map[string]string) {
	if names == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[namespace] = cacheEntry{names: names, updatedAt: time.Now()}
}

// Cache is the package-level name cache shared by all providers.
var Cache = NewNameCache()
