package middleware

import (
	"context"
	"fmt"
	"time"

	"subsnotifpro-go/internal/pkg/apperrors"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/pkg/ratelimit"
	"subsnotifpro-go/internal/pkg/resilience"
	"subsnotifpro-go/internal/pkg/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EnhancedMiddleware provides centralized middleware with all our new packages
type EnhancedMiddleware struct {
	rateLimiter    *ratelimit.RateLimiter
	circuitBreaker *resilience.CircuitBreaker
}

// NewEnhancedMiddleware creates a new enhanced middleware instance
func NewEnhancedMiddleware() *EnhancedMiddleware {
	// Create rate limiter with different limits for different endpoints
	rateLimiter := ratelimit.NewRateLimiter("api", 100, 10) // 100 requests per second, burst of 10

	// Create circuit breaker with reasonable defaults
	config := resilience.CircuitBreakerConfig{
		Name:             "default",
		MaxRequests:      10,
		Interval:         60 * time.Second,
		Timeout:          10 * time.Second,
		FailureThreshold: 5,
	}
	circuitBreaker := resilience.NewCircuitBreaker(config)

	return &EnhancedMiddleware{
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
	}
}

// CorrelationID adds correlation ID to all requests
func (m *EnhancedMiddleware) CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add to context
		ctx := logger.WithCorrelationIDValue(c.Request.Context(), correlationID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response header
		c.Header("X-Correlation-ID", correlationID)

		// Log request start
		contextLogger := logger.FromContext(ctx)
		contextLogger.WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
		}).Info("Request started")

		c.Next()
	}
}

// RateLimit applies rate limiting based on client IP
func (m *EnhancedMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.rateLimiter.Allow() {
			ctx := c.Request.Context()
			contextLogger := logger.FromContext(ctx)
			contextLogger.WithFields(logrus.Fields{
				"client_ip": c.ClientIP(),
				"path":      c.Request.URL.Path,
			}).Warn("Rate limit exceeded")

			apperrors.HandleAppError(c, apperrors.RateLimitError("Rate limit exceeded"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdaptiveRateLimit applies endpoint-specific rate limiting
func (m *EnhancedMiddleware) AdaptiveRateLimit(endpointType string) gin.HandlerFunc {
	// Create adaptive rate limiter based on endpoint type
	var adaptiveRL *ratelimit.AdaptiveRateLimiter

	switch endpointType {
	case "webhook":
		adaptiveRL = ratelimit.NewAdaptiveRateLimiter("webhook", 50, 20)
	case "api":
		adaptiveRL = ratelimit.NewAdaptiveRateLimiter("api", 100, 30)
	case "admin":
		adaptiveRL = ratelimit.NewAdaptiveRateLimiter("admin", 200, 50)
	default:
		adaptiveRL = ratelimit.NewAdaptiveRateLimiter("default", 25, 10)
	}

	return func(c *gin.Context) {
		if !adaptiveRL.Allow() {
			ctx := c.Request.Context()
			contextLogger := logger.FromContext(ctx)
			contextLogger.WithFields(logrus.Fields{
				"client_ip":     c.ClientIP(),
				"path":          c.Request.URL.Path,
				"endpoint_type": endpointType,
			}).Warn("Adaptive rate limit exceeded")

			apperrors.HandleAppError(c, apperrors.RateLimitError(
				fmt.Sprintf("Rate limit exceeded for %s endpoints", endpointType)))
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateQuery validates query parameters using our validation package
func (m *EnhancedMiddleware) ValidateQuery(target interface{}) gin.HandlerFunc {
	return validation.ValidateQuery(target)
}

// ValidateJSON validates JSON request body using our validation package
func (m *EnhancedMiddleware) ValidateJSON(target interface{}) gin.HandlerFunc {
	return validation.ValidateJSON(target)
}

// Timeout adds timeout to requests
func (m *EnhancedMiddleware) Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// Channel to signal completion
		done := make(chan struct{})

		go func() {
			defer close(done)
			c.Next()
		}()

		select {
		case <-done:
			// Request completed normally
		case <-ctx.Done():
			// Request timed out
			contextLogger := logger.FromContext(ctx)
			contextLogger.WithFields(logrus.Fields{
				"timeout": duration.String(),
				"path":    c.Request.URL.Path,
			}).Warn("Request timed out")

			apperrors.HandleAppError(c, apperrors.TimeoutError("Request timeout", ctx.Err()))
			c.Abort()
		}
	}
}

// ErrorHandler handles panics and errors
func (m *EnhancedMiddleware) ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		ctx := c.Request.Context()

		if recovered != nil {
			contextLogger := logger.FromContext(ctx)
			contextLogger.WithFields(logrus.Fields{
				"panic": recovered,
				"path":  c.Request.URL.Path,
			}).Error("Panic recovered")

			apperrors.HandleAppError(c, apperrors.InternalError("Internal server error", fmt.Errorf("panic: %v", recovered)))
		}
	})
}

// RequestLogging logs all requests with correlation ID
func (m *EnhancedMiddleware) RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		ctx := c.Request.Context()

		fields := logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"duration":   duration,
			"user_agent": c.Request.UserAgent(),
		}

		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		contextLogger := logger.FromContext(ctx)
		contextLogger.WithFields(fields).Info("Request processed")
	}
}

// CircuitBreakerSimple wraps endpoint with simple circuit breaker protection
func (m *EnhancedMiddleware) CircuitBreakerSimple() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Simple circuit breaker check - just execute
		result, err := m.circuitBreaker.Execute(ctx, func() (interface{}, error) {
			c.Next()

			// Check if any errors occurred in the handlers
			if len(c.Errors) > 0 {
				return nil, c.Errors.Last().Err
			}

			// Check HTTP status
			if c.Writer.Status() >= 500 {
				return nil, fmt.Errorf("server error: %d", c.Writer.Status())
			}

			return nil, nil
		})

		_ = result // Ignore result for middleware

		if err != nil {
			contextLogger := logger.FromContext(ctx)
			contextLogger.WithFields(logrus.Fields{
				"circuit_breaker": "default",
				"error":           err.Error(),
			}).Error("Circuit breaker triggered")

			// Check if it's a circuit breaker open error
			if err.Error() == "circuit breaker is open" {
				apperrors.HandleAppError(c, apperrors.ExternalAPIError("Service temporarily unavailable", err))
			} else {
				apperrors.HandleAppError(c, apperrors.InternalError("Internal server error", err))
			}
			c.Abort()
		}
	}
}
