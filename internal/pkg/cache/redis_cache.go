package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisCache provides a Redis-backed cache implementation
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
	keyPrefix  string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Pool     int
	Timeout  time.Duration
}

// DefaultRedisConfig returns default Redis configuration from environment
func DefaultRedisConfig() *RedisConfig {
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	pool, _ := strconv.Atoi(os.Getenv("REDIS_POOL_SIZE"))
	if pool == 0 {
		pool = 10
	}

	return &RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     port,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
		Pool:     pool,
		Timeout:  30 * time.Second,
	}
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(config *RedisConfig, defaultTTL time.Duration, keyPrefix string) (*RedisCache, error) {
	opts := &redis.Options{
		Addr:         config.Host + ":" + config.Port,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.Pool,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"host":       config.Host,
		"port":       config.Port,
		"db":         config.DB,
		"pool_size":  config.Pool,
		"key_prefix": keyPrefix,
	}).Info("Redis cache connected successfully")

	return &RedisCache{
		client:     client,
		defaultTTL: defaultTTL,
		keyPrefix:  keyPrefix,
	}, nil
}

// Get retrieves an item from Redis cache
func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	fullKey := r.buildKey(key)
	
	val, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		logrus.WithError(err).WithField("key", fullKey).Error("Redis get failed")
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		logrus.WithError(err).WithField("key", fullKey).Error("Failed to unmarshal cached value")
		return nil, false
	}

	return result, true
}

// Set stores an item in Redis cache with default TTL
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return r.SetWithTTL(ctx, key, value, r.defaultTTL)
}

// SetWithTTL stores an item in Redis cache with custom TTL
func (r *RedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.buildKey(key)
	
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := r.client.Set(ctx, fullKey, jsonValue, ttl).Err(); err != nil {
		logrus.WithError(err).WithField("key", fullKey).Error("Redis set failed")
		return err
	}

	return nil
}

// Delete removes an item from Redis cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.buildKey(key)
	return r.client.Del(ctx, fullKey).Err()
}

// Exists checks if a key exists in Redis cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)
	count, err := r.client.Exists(ctx, fullKey).Result()
	return count > 0, err
}

// Expire sets expiration time for a key
func (r *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := r.buildKey(key)
	return r.client.Expire(ctx, fullKey, ttl).Err()
}

// Clear removes all items with the cache prefix
func (r *RedisCache) Clear(ctx context.Context) error {
	pattern := r.buildKey("*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// InvalidatePattern removes all keys matching a pattern
func (r *RedisCache) InvalidatePattern(ctx context.Context, pattern string) error {
	fullPattern := r.buildKey(pattern)
	keys, err := r.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		deleted, err := r.client.Del(ctx, keys...).Result()
		if err != nil {
			return err
		}
		logrus.WithFields(logrus.Fields{
			"pattern": fullPattern,
			"deleted": deleted,
		}).Debug("Redis cache pattern invalidated")
	}

	return nil
}

// GetStats returns Redis cache statistics
func (r *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx, "memory", "stats").Result()
	if err != nil {
		return nil, err
	}

	// Get key count for our prefix
	pattern := r.buildKey("*")
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"redis_info":       info,
		"prefixed_keys":    len(keys),
		"key_prefix":       r.keyPrefix,
		"default_ttl":      r.defaultTTL.String(),
	}

	return stats, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// buildKey creates a full cache key with prefix
func (r *RedisCache) buildKey(key string) string {
	if r.keyPrefix == "" {
		return key
	}
	return r.keyPrefix + ":" + key
}

// DistributedCacheManager provides high-level operations for distributed caching
type DistributedCacheManager struct {
	redis    *RedisCache
	fallback *InMemoryCache
	enabled  bool
}

// NewDistributedCacheManager creates a new distributed cache manager
func NewDistributedCacheManager(redisConfig *RedisConfig, defaultTTL time.Duration, keyPrefix string) *DistributedCacheManager {
	var redisCache *RedisCache
	enabled := false

	// Try to create Redis cache
	if redisConfig != nil && redisConfig.Host != "" {
		var err error
		redisCache, err = NewRedisCache(redisConfig, defaultTTL, keyPrefix)
		if err != nil {
			logrus.WithError(err).Warn("Failed to connect to Redis, falling back to in-memory cache")
		} else {
			enabled = true
		}
	}

	// Always create fallback in-memory cache
	fallback := NewInMemoryCache(defaultTTL, 5*time.Minute)

	return &DistributedCacheManager{
		redis:    redisCache,
		fallback: fallback,
		enabled:  enabled,
	}
}

// GetOrSet retrieves from cache or executes the function and caches the result
func (dcm *DistributedCacheManager) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	// Try Redis first if enabled
	if dcm.enabled {
		if value, found := dcm.redis.Get(ctx, key); found {
			logrus.WithField("cache_key", key).Debug("Redis cache hit")
			return value, nil
		}
	}

	// Try fallback cache
	if value, found := dcm.fallback.Get(key); found {
		logrus.WithField("cache_key", key).Debug("In-memory cache hit")
		return value, nil
	}

	// Cache miss - execute function
	logrus.WithField("cache_key", key).Debug("Cache miss - executing function")
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in both caches
	if dcm.enabled {
		if err := dcm.redis.Set(ctx, key, value); err != nil {
			logrus.WithError(err).WithField("cache_key", key).Warn("Failed to cache in Redis")
		}
	}
	dcm.fallback.Set(key, value)

	logrus.WithField("cache_key", key).Debug("Value cached")
	return value, nil
}

// GetOrSetWithTTL retrieves from cache or executes the function and caches the result with custom TTL
func (dcm *DistributedCacheManager) GetOrSetWithTTL(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try Redis first if enabled
	if dcm.enabled {
		if value, found := dcm.redis.Get(ctx, key); found {
			logrus.WithField("cache_key", key).Debug("Redis cache hit")
			return value, nil
		}
	}

	// Try fallback cache
	if value, found := dcm.fallback.Get(key); found {
		logrus.WithField("cache_key", key).Debug("In-memory cache hit")
		return value, nil
	}

	// Cache miss - execute function
	logrus.WithField("cache_key", key).Debug("Cache miss - executing function")
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in both caches with custom TTL
	if dcm.enabled {
		if err := dcm.redis.SetWithTTL(ctx, key, value, ttl); err != nil {
			logrus.WithError(err).WithField("cache_key", key).Warn("Failed to cache in Redis")
		}
	}
	dcm.fallback.SetWithTTL(key, value, ttl)

	logrus.WithField("cache_key", key).Debug("Value cached with custom TTL")
	return value, nil
}

// Invalidate removes a specific key from all caches
func (dcm *DistributedCacheManager) Invalidate(ctx context.Context, key string) {
	if dcm.enabled {
		if err := dcm.redis.Delete(ctx, key); err != nil {
			logrus.WithError(err).WithField("cache_key", key).Warn("Failed to invalidate Redis cache")
		}
	}
	dcm.fallback.Delete(key)
	logrus.WithField("cache_key", key).Debug("Cache invalidated")
}

// InvalidatePattern removes all keys matching a pattern from all caches
func (dcm *DistributedCacheManager) InvalidatePattern(ctx context.Context, pattern string) {
	if dcm.enabled {
		if err := dcm.redis.InvalidatePattern(ctx, pattern); err != nil {
			logrus.WithError(err).WithField("pattern", pattern).Warn("Failed to invalidate Redis cache pattern")
		}
	}
	
	// For in-memory cache, use the existing pattern invalidation
	dcm.fallback.mutex.Lock()
	defer dcm.fallback.mutex.Unlock()

	keysToDelete := make([]string, 0)
	for key := range dcm.fallback.items {
		if matchesPattern(key, pattern) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(dcm.fallback.items, key)
	}

	logrus.WithFields(logrus.Fields{
		"pattern":       pattern,
		"deleted_count": len(keysToDelete),
	}).Debug("Cache pattern invalidated")
}

// GetStats returns cache statistics from all caches
func (dcm *DistributedCacheManager) GetStats(ctx context.Context) map[string]interface{} {
	stats := map[string]interface{}{
		"redis_enabled": dcm.enabled,
		"fallback":      dcm.fallback.Size(),
	}

	if dcm.enabled {
		redisStats, err := dcm.redis.GetStats(ctx)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get Redis stats")
			stats["redis_error"] = err.Error()
		} else {
			stats["redis"] = redisStats
		}
	}

	return stats
}

// Close closes all cache connections
func (dcm *DistributedCacheManager) Close() {
	if dcm.enabled && dcm.redis != nil {
		if err := dcm.redis.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close Redis connection")
		}
	}
	if dcm.fallback != nil {
		dcm.fallback.Close()
	}
}

// IsRedisEnabled returns whether Redis is enabled
func (dcm *DistributedCacheManager) IsRedisEnabled() bool {
	return dcm.enabled
}

// Helper function
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
