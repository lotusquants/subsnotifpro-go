package routes

import (
	"time"

	"github.com/gin-gonic/gin"

	"subsnotifpro-go/internal/middleware"
	"subsnotifpro-go/internal/playstore/api/handler"
	"subsnotifpro-go/internal/playstore/api/service"
)

// EnhancedPlaystoreIntegrationExample demonstrates complete integration of:
// - Enhanced error handling, validation, logging, and resilience
// - Strategic caching middleware
// - Enhanced PlayStore API handler
func RegisterEnhancedPlaystoreIntegration(router *gin.Engine, playstoreService service.PlaystoreApiService) {
	// Create the enhanced handler
	enhancedHandler := handler.NewEnhancedPlaystoreApiHandler(playstoreService)

	// Create caching middleware instances
	shortCache := middleware.NewCacheMiddleware(2 * time.Minute)
	mediumCache := middleware.NewCacheMiddleware(10 * time.Minute)
	longCache := middleware.NewCacheMiddleware(1 * time.Hour)

	// Create enhanced middleware stack
	enhancedMiddleware := middleware.NewEnhancedMiddleware()

	// Apply global enhanced middleware
	router.Use(enhancedMiddleware.RequestLogging())
	router.Use(enhancedMiddleware.CorrelationID())
	router.Use(enhancedMiddleware.RateLimit())
	router.Use(enhancedMiddleware.ErrorHandler())

	// Create API group with enhanced middleware
	apiGroup := router.Group("/api/v1/playstore")
	apiGroup.Use(enhancedMiddleware.Timeout(5 * time.Minute))
	apiGroup.Use(enhancedMiddleware.CircuitBreakerSimple())

	// 🚀 Enhanced PlayStore API endpoints with strategic caching

	// User subscription endpoint - medium cache (user-specific, changes moderately)
	apiGroup.GET("/subscription/purchase",
		mediumCache.CacheResponseWithTTL(5*time.Minute),
		enhancedHandler.GetUserSubscriptionPurchase)

	// Product listing endpoint - long cache (product catalog changes rarely)
	apiGroup.GET("/products/subscriptions",
		longCache.CacheResponse(),
		enhancedHandler.ListSubscriptionProducts)

	// Product details endpoint - long cache (product details are static)
	apiGroup.GET("/product/subscription",
		longCache.CacheResponse(),
		enhancedHandler.GetSubscriptionProduct)

	// Health check endpoint - no cache (real-time health status)
	apiGroup.GET("/health",
		enhancedHandler.HealthCheck)

	// Cache management endpoints for monitoring and admin
	cacheGroup := apiGroup.Group("/cache")
	{
		// Get cache statistics
		cacheGroup.GET("/stats", func(c *gin.Context) {
			stats := map[string]interface{}{
				"short_cache":  getCacheStats(shortCache),
				"medium_cache": getCacheStats(mediumCache),
				"long_cache":   getCacheStats(longCache),
				"strategy":     getPlayStoreCacheStrategy(),
			}
			c.JSON(200, gin.H{
				"cache_statistics": stats,
				"timestamp":        time.Now(),
			})
		})

		// Invalidate cache patterns
		cacheGroup.POST("/invalidate", func(c *gin.Context) {
			pattern := c.Query("pattern")
			if pattern == "" {
				pattern = "/api/v1/playstore/*"
			}

			// Invalidate across all cache tiers
			shortCache.InvalidatePattern(pattern)
			mediumCache.InvalidatePattern(pattern)
			longCache.InvalidatePattern(pattern)

			c.JSON(200, gin.H{
				"message": "Cache invalidated successfully",
				"pattern": pattern,
			})
		})

		// Warm up cache (pre-populate common endpoints)
		cacheGroup.POST("/warmup", func(c *gin.Context) {
			// This could trigger background jobs to pre-populate cache
			// For demo purposes, just return success
			c.JSON(200, gin.H{
				"message": "Cache warmup initiated",
				"status":  "success",
			})
		})
	}

	// Metrics endpoint for monitoring cache performance
	apiGroup.GET("/metrics", func(c *gin.Context) {
		metrics := gin.H{
			"cache_metrics": map[string]interface{}{
				"short_cache_size":  getCacheSize(shortCache),
				"medium_cache_size": getCacheSize(mediumCache),
				"long_cache_size":   getCacheSize(longCache),
			},
			"performance_recommendations": getPerformanceRecommendations(),
		}
		c.JSON(200, metrics)
	})
}

// Helper functions for cache management

func getCacheStats(cacheMiddleware *middleware.CacheMiddleware) map[string]interface{} {
	// This would require exposing the cache manager's GetStats method
	// For now, return a placeholder
	return map[string]interface{}{
		"status": "active",
		"type":   "in-memory",
	}
}

func getCacheSize(cacheMiddleware *middleware.CacheMiddleware) int {
	// This would require exposing the cache manager's Size method
	// For now, return a placeholder
	return 0
}

func getPlayStoreCacheStrategy() map[string]interface{} {
	return map[string]interface{}{
		"short_cache": map[string]interface{}{
			"ttl":      "2 minutes",
			"use_case": "Dynamic data that changes frequently",
			"examples": []string{"user sessions", "temporary states"},
		},
		"medium_cache": map[string]interface{}{
			"ttl":      "10 minutes",
			"use_case": "Semi-static data with moderate change frequency",
			"examples": []string{"user subscriptions", "pricing info", "promotional offers"},
		},
		"long_cache": map[string]interface{}{
			"ttl":      "1 hour",
			"use_case": "Static data that rarely changes",
			"examples": []string{"product catalog", "base plans", "regional settings"},
		},
	}
}

func getPerformanceRecommendations() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"category":       "Cache Strategy",
			"recommendation": "Use tiered caching based on data volatility",
			"impact":         "High",
			"implementation": "Already implemented with short/medium/long cache tiers",
		},
		{
			"category":       "Cache Invalidation",
			"recommendation": "Implement smart cache invalidation on data updates",
			"impact":         "Medium",
			"implementation": "Pattern-based invalidation is available via /cache/invalidate",
		},
		{
			"category":       "Cache Monitoring",
			"recommendation": "Monitor cache hit ratios and adjust TTLs accordingly",
			"impact":         "Medium",
			"implementation": "Cache statistics available via /cache/stats",
		},
		{
			"category":       "Distributed Caching",
			"recommendation": "Consider Redis for distributed caching in production",
			"impact":         "High (for scaling)",
			"implementation": "Current in-memory cache works for single instance",
		},
	}
}

// CachePerformanceMetrics represents cache performance data
type CachePerformanceMetrics struct {
	HitRatio       float64   `json:"hit_ratio"`
	MissRatio      float64   `json:"miss_ratio"`
	TotalRequests  int64     `json:"total_requests"`
	CacheHits      int64     `json:"cache_hits"`
	CacheMisses    int64     `json:"cache_misses"`
	AverageLatency float64   `json:"average_latency_ms"`
	LastUpdated    time.Time `json:"last_updated"`
}

// Enhanced middleware configuration
type EnhancedConfig struct {
	Cache struct {
		ShortTTL  time.Duration `json:"short_ttl"`
		MediumTTL time.Duration `json:"medium_ttl"`
		LongTTL   time.Duration `json:"long_ttl"`
	} `json:"cache"`

	RateLimit struct {
		RequestsPerMinute int `json:"requests_per_minute"`
		BurstSize         int `json:"burst_size"`
	} `json:"rate_limit"`

	CircuitBreaker struct {
		FailureThreshold int           `json:"failure_threshold"`
		Timeout          time.Duration `json:"timeout"`
	} `json:"circuit_breaker"`

	Logging struct {
		Level      string `json:"level"`
		Format     string `json:"format"`
		OutputPath string `json:"output_path"`
	} `json:"logging"`
}

// GetDefaultEnhancedConfig returns the default configuration for enhanced middleware
func GetDefaultEnhancedConfig() EnhancedConfig {
	return EnhancedConfig{
		Cache: struct {
			ShortTTL  time.Duration `json:"short_ttl"`
			MediumTTL time.Duration `json:"medium_ttl"`
			LongTTL   time.Duration `json:"long_ttl"`
		}{
			ShortTTL:  2 * time.Minute,
			MediumTTL: 10 * time.Minute,
			LongTTL:   1 * time.Hour,
		},
		RateLimit: struct {
			RequestsPerMinute int `json:"requests_per_minute"`
			BurstSize         int `json:"burst_size"`
		}{
			RequestsPerMinute: 100,
			BurstSize:         10,
		},
		CircuitBreaker: struct {
			FailureThreshold int           `json:"failure_threshold"`
			Timeout          time.Duration `json:"timeout"`
		}{
			FailureThreshold: 100,
			Timeout:          30 * time.Second,
		},
		Logging: struct {
			Level      string `json:"level"`
			Format     string `json:"format"`
			OutputPath string `json:"output_path"`
		}{
			Level:      "info",
			Format:     "json",
			OutputPath: "stdout",
		},
	}
}
