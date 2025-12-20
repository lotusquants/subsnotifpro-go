package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"subsnotifpro-go/internal/pkg/cache"
	"subsnotifpro-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
	CachedAt   time.Time           `json:"cached_at"`
}

// CacheMiddleware provides HTTP response caching
type CacheMiddleware struct {
	cacheManager *cache.CacheManager
	defaultTTL   time.Duration
}

// NewCacheMiddleware creates a new cache middleware instance
func NewCacheMiddleware(defaultTTL time.Duration) *CacheMiddleware {
	// Create cache manager with 5-minute default TTL and 1-minute cleanup interval
	cacheManager := cache.NewCacheManager(defaultTTL, time.Minute)

	return &CacheMiddleware{
		cacheManager: cacheManager,
		defaultTTL:   defaultTTL,
	}
}

// CacheResponse caches GET requests with the default TTL
func (cm *CacheMiddleware) CacheResponse() gin.HandlerFunc {
	return cm.CacheResponseWithTTL(cm.defaultTTL)
}

// CacheResponseWithTTL caches GET requests with a custom TTL
func (cm *CacheMiddleware) CacheResponseWithTTL(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Generate cache key based on the full URL and query parameters
		cacheKey := generateCacheKey(c)

		// Check if response is already cached
		if cached, found := cm.cacheManager.Get(cacheKey); found {
			if cachedResp, ok := cached.(*CachedResponse); ok {
				// Serve from cache
				cm.serveCachedResponse(c, cachedResp, cacheKey)
				return
			}
		}

		// Cache miss - capture the response
		cm.captureAndCacheResponse(c, cacheKey, ttl)
	}
}

// CacheInvalidate provides cache invalidation middleware
func (cm *CacheMiddleware) CacheInvalidate(patterns ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only invalidate on successful write operations
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			method := c.Request.Method
			if method == http.MethodPost || method == http.MethodPut ||
				method == http.MethodPatch || method == http.MethodDelete {

				for _, pattern := range patterns {
					cm.cacheManager.InvalidatePattern(pattern)
					logger.FromContext(c.Request.Context()).WithFields(logrus.Fields{
						"method":  method,
						"path":    c.Request.URL.Path,
						"pattern": pattern,
					}).Debug("Cache invalidated due to write operation")
				}
			}
		}
	}
}

// CacheStats returns cache statistics
func (cm *CacheMiddleware) CacheStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := cm.cacheManager.GetStats()
		c.JSON(http.StatusOK, gin.H{
			"cache_stats": stats,
		})
	}
}

// Close cleans up the cache middleware
func (cm *CacheMiddleware) Close() {
	if cm.cacheManager != nil {
		cm.cacheManager.Close()
	}
}

// InvalidatePattern invalidates cache entries matching a pattern
func (cm *CacheMiddleware) InvalidatePattern(pattern string) {
	cm.cacheManager.InvalidatePattern(pattern)
}

// Helper functions

func generateCacheKey(c *gin.Context) string {
	// Include method, path, and query parameters in the cache key
	key := fmt.Sprintf("%s:%s", c.Request.Method, c.Request.URL.Path)

	// Add query parameters to make cache key unique
	if c.Request.URL.RawQuery != "" {
		key += "?" + c.Request.URL.RawQuery
	}

	// Add user context if available (for user-specific caching)
	if userID := c.GetString("user_id"); userID != "" {
		key += ":user:" + userID
	}

	return key
}

func (cm *CacheMiddleware) serveCachedResponse(c *gin.Context, cachedResp *CachedResponse, cacheKey string) {
	// Set cached headers
	for name, values := range cachedResp.Headers {
		for _, value := range values {
			c.Header(name, value)
		}
	}

	// Add cache headers
	c.Header("X-Cache", "HIT")
	c.Header("X-Cache-Key", cacheKey)
	c.Header("X-Cached-At", cachedResp.CachedAt.Format(time.RFC3339))

	// Log cache hit
	logger.FromContext(c.Request.Context()).WithFields(logrus.Fields{
		"cache_key": cacheKey,
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
		"cached_at": cachedResp.CachedAt,
	}).Debug("Cache hit - serving cached response")

	// Write response
	c.Data(cachedResp.StatusCode, c.GetHeader("Content-Type"), cachedResp.Body)
	c.Abort()
}

func (cm *CacheMiddleware) captureAndCacheResponse(c *gin.Context, cacheKey string, ttl time.Duration) {
	// Create a response writer wrapper to capture the response
	writer := &responseWriter{
		ResponseWriter: c.Writer,
		body:           &bytes.Buffer{},
	}
	c.Writer = writer

	// Add cache miss header
	c.Header("X-Cache", "MISS")
	c.Header("X-Cache-Key", cacheKey)

	// Continue processing
	c.Next()

	// Only cache successful responses
	if writer.status >= 200 && writer.status < 300 {
		// Create cached response
		cachedResp := &CachedResponse{
			StatusCode: writer.status,
			Headers:    make(map[string][]string),
			Body:       writer.body.Bytes(),
			CachedAt:   time.Now(),
		}

		// Copy headers (excluding cache-specific headers)
		for name, values := range writer.Header() {
			if !isCacheSpecificHeader(name) {
				cachedResp.Headers[name] = values
			}
		}

		// Store in cache
		cm.cacheManager.SetWithTTL(cacheKey, cachedResp, ttl)

		// Log cache store
		logger.FromContext(c.Request.Context()).WithFields(logrus.Fields{
			"cache_key":   cacheKey,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": writer.status,
			"ttl":         ttl,
		}).Debug("Response cached")
	}
}

// responseWriter wraps gin.ResponseWriter to capture response data
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Status() int {
	return w.status
}

// Helper function to check if header should be excluded from caching
func isCacheSpecificHeader(name string) bool {
	cacheHeaders := map[string]bool{
		"X-Cache":     true,
		"X-Cache-Key": true,
		"X-Cached-At": true,
		"Date":        true,
		"Age":         true,
	}
	return cacheHeaders[name]
}

// Cache configuration for different endpoint types
var (
	// Short cache for dynamic data
	ShortCacheTTL = 2 * time.Minute

	// Medium cache for semi-static data
	MediumCacheTTL = 10 * time.Minute

	// Long cache for static data
	LongCacheTTL = 1 * time.Hour
)

// Predefined cache middleware instances
func NewShortCache() gin.HandlerFunc {
	middleware := NewCacheMiddleware(ShortCacheTTL)
	return middleware.CacheResponse()
}

func NewMediumCache() gin.HandlerFunc {
	middleware := NewCacheMiddleware(MediumCacheTTL)
	return middleware.CacheResponse()
}

func NewLongCache() gin.HandlerFunc {
	middleware := NewCacheMiddleware(LongCacheTTL)
	return middleware.CacheResponse()
}
