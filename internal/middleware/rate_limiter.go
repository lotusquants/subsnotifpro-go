package middleware

import (
	"os"
	"strconv"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/didip/tollbooth_gin"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	TTL               time.Duration
	IPLookups         []string
	Methods           []string
	Headers           map[string][]string
	BasicAuthUsers    []string
	IgnoreURL         func(*limiter.Limiter, *gin.Context) bool
}

// DefaultRateLimiterConfig returns default rate limiting configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	// Load from environment variables or use defaults
	rps, _ := strconv.ParseFloat(getEnvOrDefault("RATE_LIMIT_RPS", "10.0"), 64)
	burst, _ := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_BURST", "20"))
	ttlMinutes, _ := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_TTL_MINUTES", "60"))

	return &RateLimiterConfig{
		RequestsPerSecond: rps,
		BurstSize:         burst,
		TTL:               time.Duration(ttlMinutes) * time.Minute,
		IPLookups:         []string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"},
		Methods:           []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		Headers:           make(map[string][]string),
		BasicAuthUsers:    []string{},
		IgnoreURL: func(lmt *limiter.Limiter, c *gin.Context) bool {
			// Ignore health check endpoints
			if c.Request.URL.Path == "/api/health" || c.Request.URL.Path == "/metrics" {
				return true
			}
			return false
		},
	}
}

// StrictRateLimiterConfig returns strict rate limiting for sensitive endpoints
func StrictRateLimiterConfig() *RateLimiterConfig {
	config := DefaultRateLimiterConfig()
	config.RequestsPerSecond = 2.0  // Much stricter for sensitive operations
	config.BurstSize = 5
	config.TTL = 5 * time.Minute
	return config
}

// CreateRateLimiter creates a new rate limiter with the given config
func CreateRateLimiter(config *RateLimiterConfig) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(config.RequestsPerSecond, &limiter.ExpirableOptions{
		DefaultExpirationTTL: config.TTL,
	})

	// Configure the limiter
	lmt.SetBurst(config.BurstSize)
	lmt.SetIPLookups(config.IPLookups)
	lmt.SetMethods(config.Methods)
	
	// Set headers for rate limiting info
	if len(config.Headers) > 0 {
		lmt.SetHeaders(config.Headers)
	}

	// Set basic auth users if any
	if len(config.BasicAuthUsers) > 0 {
		lmt.SetBasicAuthUsers(config.BasicAuthUsers)
	}

	// Note: tollbooth doesn't support custom ignore URL functions in this version
	// We'll handle ignoring URLs in the wrapper function

	// Log rate limiting configuration
	logrus.WithFields(logrus.Fields{
		"requests_per_second": config.RequestsPerSecond,
		"burst_size":         config.BurstSize,
		"ttl_minutes":        config.TTL.Minutes(),
	}).Info("Rate limiter configured")

	return tollbooth_gin.LimitHandler(lmt)
}

// DefaultRateLimiter creates a rate limiter with default configuration
func DefaultRateLimiter() gin.HandlerFunc {
	return CreateRateLimiter(DefaultRateLimiterConfig())
}

// StrictRateLimiter creates a rate limiter with strict configuration for sensitive endpoints
func StrictRateLimiter() gin.HandlerFunc {
	return CreateRateLimiter(StrictRateLimiterConfig())
}

// WebhookRateLimiter creates a rate limiter specifically for webhook endpoints
func WebhookRateLimiter() gin.HandlerFunc {
	config := &RateLimiterConfig{
		RequestsPerSecond: 50.0, // Higher limit for webhooks
		BurstSize:         100,
		TTL:               1 * time.Minute,
		IPLookups:         []string{"X-Forwarded-For", "X-Real-IP", "RemoteAddr"},
		Methods:           []string{"POST"},
		Headers:           make(map[string][]string),
		BasicAuthUsers:    []string{},
		IgnoreURL: func(lmt *limiter.Limiter, c *gin.Context) bool {
			return false // Don't ignore webhook endpoints
		},
	}

	logrus.WithFields(logrus.Fields{
		"type":               "webhook",
		"requests_per_second": config.RequestsPerSecond,
		"burst_size":         config.BurstSize,
	}).Info("Webhook rate limiter configured")

	return CreateRateLimiter(config)
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
