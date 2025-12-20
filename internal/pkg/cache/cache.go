package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired checks if the cache item has expired
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// InMemoryCache provides a simple in-memory cache implementation
type InMemoryCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex

	// Configuration
	defaultTTL    time.Duration
	cleanupTicker *time.Ticker
	done          chan bool
}

// NewInMemoryCache creates a new in-memory cache instance
func NewInMemoryCache(defaultTTL time.Duration, cleanupInterval time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		items:      make(map[string]*CacheItem),
		defaultTTL: defaultTTL,
		done:       make(chan bool),
	}

	// Start cleanup goroutine
	cache.cleanupTicker = time.NewTicker(cleanupInterval)
	go cache.startCleanup()

	return cache
}

// Get retrieves an item from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if item.IsExpired() {
		// Remove expired item
		delete(c.items, key)
		return nil, false
	}

	return item.Data, true
}

// Set stores an item in the cache with default TTL
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores an item in the cache with custom TTL
func (c *InMemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}
}

// Delete removes an item from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
}

// Size returns the number of items in the cache
func (c *InMemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// Close stops the cleanup goroutine
func (c *InMemoryCache) Close() {
	if c.cleanupTicker != nil {
		c.cleanupTicker.Stop()
	}
	close(c.done)
}

// startCleanup runs the cleanup process
func (c *InMemoryCache) startCleanup() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

// cleanup removes expired items from the cache
func (c *InMemoryCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.items, key)
	}

	if len(expiredKeys) > 0 {
		logrus.WithField("expired_count", len(expiredKeys)).Debug("Cleaned up expired cache items")
	}
}

// CacheManager provides high-level cache operations
type CacheManager struct {
	cache *InMemoryCache
}

// NewCacheManager creates a new cache manager
func NewCacheManager(defaultTTL, cleanupInterval time.Duration) *CacheManager {
	return &CacheManager{
		cache: NewInMemoryCache(defaultTTL, cleanupInterval),
	}
}

// GetOrSet retrieves from cache or executes the function and caches the result
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := cm.cache.Get(key); found {
		logrus.WithField("cache_key", key).Debug("Cache hit")
		return value, nil
	}

	// Cache miss - execute function
	logrus.WithField("cache_key", key).Debug("Cache miss - executing function")
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	cm.cache.Set(key, value)
	logrus.WithField("cache_key", key).Debug("Value cached")

	return value, nil
}

// GetOrSetWithTTL retrieves from cache or executes the function and caches the result with custom TTL
func (cm *CacheManager) GetOrSetWithTTL(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	if value, found := cm.cache.Get(key); found {
		logrus.WithField("cache_key", key).Debug("Cache hit")
		return value, nil
	}

	// Cache miss - execute function
	logrus.WithField("cache_key", key).Debug("Cache miss - executing function")
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache with custom TTL
	cm.cache.SetWithTTL(key, value, ttl)
	logrus.WithField("cache_key", key).Debug("Value cached with custom TTL")

	return value, nil
}

// Invalidate removes a specific key from cache
func (cm *CacheManager) Invalidate(key string) {
	cm.cache.Delete(key)
	logrus.WithField("cache_key", key).Debug("Cache invalidated")
}

// InvalidatePattern removes all keys matching a pattern (simple prefix matching)
func (cm *CacheManager) InvalidatePattern(pattern string) {
	cm.cache.mutex.Lock()
	defer cm.cache.mutex.Unlock()

	keysToDelete := make([]string, 0)
	for key := range cm.cache.items {
		if matchesPattern(key, pattern) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(cm.cache.items, key)
	}

	logrus.WithFields(logrus.Fields{
		"pattern":       pattern,
		"deleted_count": len(keysToDelete),
	}).Debug("Cache pattern invalidated")
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() map[string]interface{} {
	cm.cache.mutex.RLock()
	defer cm.cache.mutex.RUnlock()

	now := time.Now()
	totalItems := len(cm.cache.items)
	expiredItems := 0

	for _, item := range cm.cache.items {
		if now.After(item.ExpiresAt) {
			expiredItems++
		}
	}

	return map[string]interface{}{
		"total_items":   totalItems,
		"active_items":  totalItems - expiredItems,
		"expired_items": expiredItems,
		"hit_ratio":     "not_implemented", // Would need hit/miss counters
	}
}

// Close closes the cache manager
func (cm *CacheManager) Close() {
	cm.cache.Close()
}

// Get retrieves an item from the cache
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	return cm.cache.Get(key)
}

// SetWithTTL stores an item in the cache with custom TTL
func (cm *CacheManager) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	cm.cache.SetWithTTL(key, value, ttl)
}

// Set stores an item in the cache with default TTL
func (cm *CacheManager) Set(key string, value interface{}) {
	cm.cache.Set(key, value)
}

// Helper function for pattern matching (simple prefix matching)
func matchesPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple prefix matching for now
	// Can be enhanced with regex or glob patterns
	return len(key) >= len(pattern) && key[:len(pattern)] == pattern
}

// GenerateCacheKey creates a consistent cache key from multiple parameters
func GenerateCacheKey(prefix string, params ...interface{}) string {
	key := prefix
	for _, param := range params {
		key += fmt.Sprintf(":%v", param)
	}
	return key
}

// GenerateJSONCacheKey creates a cache key from JSON-serializable parameters
func GenerateJSONCacheKey(prefix string, params interface{}) (string, error) {
	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal params to JSON: %w", err)
	}
	return fmt.Sprintf("%s:%s", prefix, string(jsonBytes)), nil
}
