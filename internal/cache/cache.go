// Package cache provides caching utilities for file system operations.
package cache

import (
	"os"
	"sync"
	"time"
)

// Cache provides a thread-safe cache for file system operations.
type Cache struct {
	mu         sync.RWMutex
	statCache  map[string]statEntry
	dirCache   map[string][]os.DirEntry
	sizeCache  map[string]int64
	createdAt  time.Time
	ttl        time.Duration
}

type statEntry struct {
	info  os.FileInfo
	err   error
}

// New creates a new Cache with the specified TTL.
// If ttl is 0, entries never expire.
func New(ttl time.Duration) *Cache {
	return &Cache{
		statCache: make(map[string]statEntry),
		dirCache:  make(map[string][]os.DirEntry),
		sizeCache: make(map[string]int64),
		createdAt: time.Now(),
		ttl:       ttl,
	}
}

// Stat returns cached os.Stat result or performs the operation and caches it.
func (c *Cache) Stat(path string) (os.FileInfo, error) {
	// Check cache first
	c.mu.RLock()
	if entry, ok := c.statCache[path]; ok {
		c.mu.RUnlock()
		return entry.info, entry.err
	}
	c.mu.RUnlock()

	// Perform actual stat
	info, err := os.Stat(path)

	// Cache result
	c.mu.Lock()
	c.statCache[path] = statEntry{info: info, err: err}
	c.mu.Unlock()

	return info, err
}

// ReadDir returns cached os.ReadDir result or performs the operation and caches it.
func (c *Cache) ReadDir(path string) ([]os.DirEntry, error) {
	// Check cache first
	c.mu.RLock()
	if entries, ok := c.dirCache[path]; ok {
		c.mu.RUnlock()
		return entries, nil
	}
	c.mu.RUnlock()

	// Perform actual ReadDir
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Cache result
	c.mu.Lock()
	c.dirCache[path] = entries
	c.mu.Unlock()

	return entries, nil
}

// DirSize returns cached directory size or calculates and caches it.
func (c *Cache) DirSize(path string, walkFn func(string) int64) int64 {
	// Check cache first
	c.mu.RLock()
	if size, ok := c.sizeCache[path]; ok {
		c.mu.RUnlock()
		return size
	}
	c.mu.RUnlock()

	// Calculate size
	size := walkFn(path)

	// Cache result
	c.mu.Lock()
	c.sizeCache[path] = size
	c.mu.Unlock()

	return size
}

// Invalidate removes all cached entries.
func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.statCache = make(map[string]statEntry)
	c.dirCache = make(map[string][]os.DirEntry)
	c.sizeCache = make(map[string]int64)
	c.createdAt = time.Now()
}

// InvalidatePath removes cached entries for a specific path.
func (c *Cache) InvalidatePath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.statCache, path)
	delete(c.dirCache, path)
	delete(c.sizeCache, path)
}

// IsExpired returns true if the cache has exceeded its TTL.
func (c *Cache) IsExpired() bool {
	if c.ttl == 0 {
		return false
	}
	return time.Since(c.createdAt) > c.ttl
}

// Stats returns cache statistics.
func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		StatEntries: len(c.statCache),
		DirEntries:  len(c.dirCache),
		SizeEntries: len(c.sizeCache),
		Age:         time.Since(c.createdAt),
	}
}

// CacheStats contains cache statistics.
type CacheStats struct {
	StatEntries int
	DirEntries  int
	SizeEntries int
	Age         time.Duration
}

// Global default cache instance
var (
	globalCache *Cache
	once        sync.Once
)

// Global returns the global cache instance, creating it on first use.
func Global() *Cache {
	once.Do(func() {
		globalCache = New(5 * time.Minute)
	})
	return globalCache
}

// ClearGlobal clears the global cache.
func ClearGlobal() {
	if globalCache != nil {
		globalCache.Invalidate()
	}
}
