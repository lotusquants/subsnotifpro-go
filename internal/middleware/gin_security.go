package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GinSecurityMiddleware adapts the existing SecurityMiddleware for Gin
func GinSecurityMiddleware(config SecurityConfig) gin.HandlerFunc {
	securityMiddleware := NewSecurityMiddleware(config)
	rateLimiter := NewRateLimiter(config)
	
	return func(c *gin.Context) {
		// Create a wrapper to adapt http.Handler to gin.HandlerFunc
		var handled bool
		
		// Rate limiting check
		if config.RateLimitEnabled {
			clientIP := GetClientIP(c.Request)
			if !rateLimiter.Allow(clientIP) {
				logrus.WithFields(logrus.Fields{
					"client_ip": clientIP,
					"path":     c.Request.URL.Path,
				}).Warn("Rate limit exceeded")
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "Rate limit exceeded",
					"retry_after": "60",
				})
				return
			}
		}
		
		// Wrap the gin context in an http handler
		handler := securityMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handled = true
		}))
		
		// Execute the security middleware
		handler.ServeHTTP(c.Writer, c.Request)
		
		// If the security middleware handled the request (e.g., OPTIONS), don't continue
		if !handled && c.Writer.Written() {
			return
		}
		
		c.Next()
	}
}

// SmartSecurity creates Gin security middleware based on environment
func SmartSecurity() gin.HandlerFunc {
	var config SecurityConfig
	if gin.Mode() == gin.ReleaseMode {
		config = DefaultSecurityConfig()
		// Production settings
		config.AllowedOrigins = []string{} // Must be configured via environment
		config.RateLimitEnabled = true
		config.RequestsPerMinute = 60
	} else {
		config = DefaultSecurityConfig()
		// Development settings
		config.AllowedOrigins = []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:8080"}
		config.RequestsPerMinute = 1000 // More lenient for development
	}
	
	return GinSecurityMiddleware(config)
}

// WebhookSecurity creates Gin security middleware specifically for webhook endpoints
func WebhookSecurity() gin.HandlerFunc {
	config := DefaultSecurityConfig()
	config.RateLimitEnabled = false // Webhooks have their own rate limiting
	config.AllowedOrigins = []string{"*"} // Webhooks come from external services
	
	return GinSecurityMiddleware(config)
}

// StrictSecurityMiddleware creates very strict security middleware for admin endpoints
func StrictSecurityMiddleware() gin.HandlerFunc {
	config := DefaultSecurityConfig()
	config.RequestsPerMinute = 30 // Very strict rate limiting
	config.BurstSize = 5
	config.AllowedOrigins = []string{} // Must be explicitly configured
	
	return GinSecurityMiddleware(config)
}
