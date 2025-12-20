package routes

import (
	"subsnotifpro-go/internal/circuitbreaker"
	"subsnotifpro-go/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(deps *RouteDependencies) *gin.Engine {
	router := gin.Default()

	// Enhanced Security Middleware
	router.Use(middleware.SmartSecurity())
	
	// Circuit Breaker Middleware
	if deps.CircuitBreakerManager != nil {
		router.Use(circuitbreaker.Middleware(deps.CircuitBreakerManager))
	}
	
	// Enhanced CORS Middleware
	router.Use(middleware.SmartCORS())
	
	// Rate Limiting Middleware (Global)
	router.Use(middleware.DefaultRateLimiter())
	
	// Enhanced Request Logging
	router.Use(middleware.SmartRequestLogger())
	
	// Recovery Middleware
	router.Use(gin.Recovery())

	// Observability middleware (metrics and tracing)
	if deps.ObservabilityMiddleware != nil {
		router.Use(deps.ObservabilityMiddleware.HTTPMiddleware())
		router.Use(deps.ObservabilityMiddleware.CorrelationIDMiddleware())
	}

	// Health check endpoints
	if deps.HealthChecker != nil {
		router.GET("/health", gin.WrapH(deps.HealthChecker.HTTPHealthHandler()))
		router.GET("/health/ready", gin.WrapH(deps.HealthChecker.ReadinessHandler()))
		router.GET("/health/live", gin.WrapH(deps.HealthChecker.LivenessHandler()))
	}

	// Circuit breaker monitoring endpoints
	if deps.CircuitBreakerManager != nil {
		router.GET("/api/circuit-breaker/health", circuitbreaker.HealthCheckHandler(deps.CircuitBreakerManager))
		router.GET("/api/circuit-breaker/stats", circuitbreaker.StatsHandler(deps.CircuitBreakerManager))
	}

	// Legacy health check for backwards compatibility (bypasses rate limiting)
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "timestamp": gin.H{}})
	})

	// Apply strict rate limiting to sensitive endpoints
	sensitiveGroup := router.Group("/api/admin")
	sensitiveGroup.Use(middleware.StrictRateLimiter())
	
	// Register webhook routes with webhook-specific middleware
	webhookGroup := router.Group("/api/webhooks")
	webhookGroup.Use(middleware.WebhookRateLimiter())
	webhookGroup.Use(middleware.WebhookSecurity())

	// Register all grouped routes
	registerGooglePlayRoutes(router, deps)
	registerAppStoreRoutes(router, deps)
	registerAuthRoutes(router, deps)
	registerDashboardRoutes(router, deps)
	registerUnifiedSubscriptionRoutes(router, deps)

	return router
}
