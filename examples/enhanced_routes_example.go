package routes

import (
	"time"

	"subsnotifpro-go/internal/middleware"
	"subsnotifpro-go/internal/pkg/validation"
	"subsnotifpro-go/internal/playstore/api/handler"

	"github.com/gin-gonic/gin"
)

// SetupEnhancedPlaystoreRoutes demonstrates how to use the new enhanced middleware
// with PlayStore API routes
func SetupEnhancedPlaystoreRoutes(r *gin.Engine, handler *handler.EnhancedPlaystoreApiHandler) {
	// Create enhanced middleware instance
	enhancedMW := middleware.NewEnhancedMiddleware()

	// PlayStore API group with enhanced middleware
	playstoreAPI := r.Group("/api/v1/playstore")
	{
		// Apply enhanced middleware stack
		playstoreAPI.Use(enhancedMW.CorrelationID())           // Add correlation IDs
		playstoreAPI.Use(enhancedMW.RequestLogging())          // Structured logging
		playstoreAPI.Use(enhancedMW.ErrorHandler())            // Panic recovery
		playstoreAPI.Use(enhancedMW.Timeout(30 * time.Second)) // Request timeout
		playstoreAPI.Use(enhancedMW.AdaptiveRateLimit("api"))  // Rate limiting

		// Health check endpoint (no additional middleware needed)
		playstoreAPI.GET("/health", handler.HealthCheck)

		// Subscription purchase endpoint with validation middleware
		playstoreAPI.GET("/subscription/purchase",
			enhancedMW.ValidateQuery(&validation.SubscriptionQueryParams{}), // Validate query params
			enhancedMW.CircuitBreakerSimple(),                               // Circuit breaker protection
			handler.GetUserSubscriptionPurchase,
		)

		// Subscription products endpoint
		playstoreAPI.GET("/subscription/products",
			enhancedMW.ValidateQuery(&validation.PlaystoreQueryParams{}), // Validate query params
			enhancedMW.CircuitBreakerSimple(),                            // Circuit breaker protection
			handler.ListSubscriptionProducts,
		)

		// Subscription product details endpoint
		playstoreAPI.GET("/subscription/product",
			enhancedMW.ValidateQuery(&validation.PlaystoreQueryParams{}), // Validate query params
			enhancedMW.CircuitBreakerSimple(),                            // Circuit breaker protection
			handler.GetSubscriptionProduct,
		)
	}

	// Webhook endpoints with different rate limiting
	playstoreWebhooks := r.Group("/webhooks/playstore")
	{
		playstoreWebhooks.Use(enhancedMW.CorrelationID())
		playstoreWebhooks.Use(enhancedMW.RequestLogging())
		playstoreWebhooks.Use(enhancedMW.ErrorHandler())
		playstoreWebhooks.Use(enhancedMW.Timeout(10 * time.Second))    // Shorter timeout for webhooks
		playstoreWebhooks.Use(enhancedMW.AdaptiveRateLimit("webhook")) // Different rate limit for webhooks

		// Webhook endpoints would go here
		// playstoreWebhooks.POST("/rtdn", handler.HandleRTDN)
	}

	// Admin endpoints with higher rate limits
	playstoreAdmin := r.Group("/admin/playstore")
	{
		playstoreAdmin.Use(enhancedMW.CorrelationID())
		playstoreAdmin.Use(enhancedMW.RequestLogging())
		playstoreAdmin.Use(enhancedMW.ErrorHandler())
		playstoreAdmin.Use(enhancedMW.Timeout(60 * time.Second))  // Longer timeout for admin operations
		playstoreAdmin.Use(enhancedMW.AdaptiveRateLimit("admin")) // Higher rate limits for admin

		// Admin endpoints would go here
		// playstoreAdmin.GET("/metrics", handler.GetMetrics)
		// playstoreAdmin.POST("/cache/clear", handler.ClearCache)
	}
}

// Example of how to setup routes with basic middleware
func SetupBasicPlaystoreRoutes(r *gin.Engine, handler *handler.PlaystoreApiHandler) {
	enhancedMW := middleware.NewEnhancedMiddleware()

	// Simple middleware stack for existing handlers
	playstoreAPI := r.Group("/api/v1/playstore/basic")
	{
		playstoreAPI.Use(enhancedMW.CorrelationID())  // Add correlation IDs
		playstoreAPI.Use(enhancedMW.RequestLogging()) // Structured logging
		playstoreAPI.Use(enhancedMW.ErrorHandler())   // Panic recovery
		playstoreAPI.Use(enhancedMW.RateLimit())      // Basic rate limiting

		// Use existing handlers (they will benefit from correlation IDs and logging)
		playstoreAPI.GET("/subscription/purchase", handler.GetUserSubscriptionPurchase)
		playstoreAPI.GET("/subscription/products", handler.ListSubscriptionProducts)
	}
}

// Example middleware configuration for different environments
func GetMiddlewareConfig(env string) *middleware.EnhancedMiddleware {
	// Different configurations based on environment
	switch env {
	case "production":
		return middleware.NewEnhancedMiddleware() // Use production defaults
	case "staging":
		return middleware.NewEnhancedMiddleware() // Use staging configuration
	case "development":
		return middleware.NewEnhancedMiddleware() // Use development configuration
	default:
		return middleware.NewEnhancedMiddleware()
	}
}
