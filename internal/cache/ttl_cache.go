// Package cache provides thread-safe caching utilities with time-based expiration.
package cache

import (
	"sync"
	"time"
)

// TTLCache is a thread-safe cache with time-based expiration.
// It stores key-value pairs and tracks a single timestamp for the entire cache.
// When the TTL expires, all entries are considered stale.
type TTLCache[K comparable, V any] struct {
	mu        sync.RWMutex
	data      map[K]V
	timestamp time.Time
	ttl       time.Duration
}

// New creates a new TTLCache with the given TTL duration.
// The cache starts empty and with a zero timestamp (expired).
func New[K comparable, V any](ttl time.Duration) *TTLCache[K, V] {
	return &TTLCache[K, V]{
		data: make(map[K]V),
		ttl:  ttl,
	}
}

// Get retrieves a value from the cache.
// Returns the value and ok=true if the key exists and cache is not expired.
// Returns zero value and ok=false if the key doesn't exist or cache is expired.
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.IsExpiredLocked() {
		var zero V
		return zero, false
	}

	value, ok := c.data[key]
	return value, ok
}

// Set stores a value in the cache and updates the timestamp to now.
// This resets the TTL timer for the entire cache.
func (c *TTLCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[K]V)
	}
	c.data[key] = value
	c.timestamp = time.Now()
}

// GetAll returns a copy of all cached values if the cache is not expired.
// Returns nil if the cache is expired.
// The returned map is a shallow copy and safe to modify.
func (c *TTLCache[K, V]) GetAll() map[K]V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.IsExpiredLocked() {
		return nil
	}

	// Return a copy to prevent external modification
	result := make(map[K]V, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// SetAll replaces all cached values with the provided data and updates the timestamp.
// This resets the TTL timer for the entire cache.
func (c *TTLCache[K, V]) SetAll(data map[K]V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[K]V, len(data))
	for k, v := range data {
		c.data[k] = v
	}
	c.timestamp = time.Now()
}

// IsExpired checks if the cache has expired based on TTL.
// A cache with a zero timestamp is considered expired.
func (c *TTLCache[K, V]) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsExpiredLocked()
}

// IsExpiredLocked checks if the cache has expired without acquiring locks.
// MUST be called with at least a read lock held.
func (c *TTLCache[K, V]) IsExpiredLocked() bool {
	return c.timestamp.IsZero() || time.Since(c.timestamp) >= c.ttl
}

// Invalidate clears all cached data and resets the timestamp.
// This forces the cache to be considered expired until new data is Set.
func (c *TTLCache[K, V]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[K]V)
	c.timestamp = time.Time{}
}

// Len returns the number of items currently in the cache.
// This does not check expiration - it returns the count even if expired.
func (c *TTLCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
