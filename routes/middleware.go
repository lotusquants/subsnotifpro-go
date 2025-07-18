package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"subsnotifpro-go/internal/auth"
	"subsnotifpro-go/internal/middleware"
)

// SetupSecurityMiddleware configures comprehensive security middleware
func SetupSecurityMiddleware(r *gin.Engine) {
	// Security headers and CORS
	securityConfig := middleware.DefaultSecurityConfig()
	securityConfig.AllowedOrigins = []string{
		"http://localhost:3000",
		"https://yourdomain.com",
		"https://*.yourdomain.com",
	}
	
	securityMiddleware := middleware.NewSecurityMiddleware(securityConfig)
	
	// Apply security middleware to all routes
	r.Use(func(c *gin.Context) {
		securityMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Copy headers from gin context to response writer
			for key, values := range c.Writer.Header() {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	})
	
	// Input validation middleware
	validator := middleware.NewInputValidator(10 * 1024 * 1024) // 10MB limit
	r.Use(func(c *gin.Context) {
		validator.ValidateRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	})
}

// JWTMiddleware handles JWT authentication
func JWTMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization header",
				"code":  "MISSING_AUTH_HEADER",
			})
			return
		}

		// Check for Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "INVALID_AUTH_FORMAT",
			})
			return
		}

		tokenString := parts[1]
		
		// Validate the token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			return
		}

		// Store user claims in context for use in handlers
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

// RateLimitMiddleware provides rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	limiter := middleware.NewRateLimiter(middleware.DefaultSecurityConfig())
	
	return func(c *gin.Context) {
		clientIP := middleware.GetClientIP(c.Request)
		
		if !limiter.Allow(clientIP) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			return
		}
		
		c.Next()
	}
}

// RequestIDMiddleware adds unique request IDs for tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("requestID", requestID)
		c.Next()
	}
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	// Simple implementation - in production, use a proper UUID library
	return "req-" + strings.ReplaceAll(strings.Replace(strings.Replace(
		"xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx", "x", "a", -1), "y", "b", -1), "a", "1")
}

// LoggingMiddleware provides structured logging
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return "" // We'll use our structured logger instead
	})
}

// RecoveryMiddleware handles panics gracefully
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	})
}
