package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	// Load allowed origins from environment
	originsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	
	if originsEnv != "" {
		allowedOrigins = strings.Split(originsEnv, ",")
		// Trim whitespace
		for i, origin := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		// Default development origins
		allowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:3001", 
			"http://localhost:8080",
			"https://localhost:3000",
			"https://localhost:3001",
		}
	}

	return &CORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "Accept", "Cache-Control",
			"X-Requested-With", "X-Request-ID", "X-Correlation-ID",
		},
		ExposedHeaders: []string{
			"Content-Length", "X-Request-ID", "X-Correlation-ID",
			"X-Rate-Limit-Limit", "X-Rate-Limit-Remaining", "X-Rate-Limit-Reset",
		},
		AllowCredentials: true,
		MaxAge:          12 * time.Hour,
	}
}

// ProductionCORSConfig returns production CORS configuration
func ProductionCORSConfig() *CORSConfig {
	config := DefaultCORSConfig()
	
	// In production, be more restrictive
	prodOrigins := os.Getenv("PROD_CORS_ALLOWED_ORIGINS")
	if prodOrigins != "" {
		config.AllowedOrigins = strings.Split(prodOrigins, ",")
		for i, origin := range config.AllowedOrigins {
			config.AllowedOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		// Default to empty for production - must be explicitly configured
		config.AllowedOrigins = []string{}
	}
	
	return config
}

// CreateCORSMiddleware creates CORS middleware with the given configuration
func CreateCORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     config.AllowedMethods,
		AllowHeaders:     config.AllowedHeaders,
		ExposeHeaders:    config.ExposedHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:          config.MaxAge,
	}

	// Log CORS configuration
	logrus.WithFields(logrus.Fields{
		"allowed_origins":  len(config.AllowedOrigins),
		"allowed_methods":  len(config.AllowedMethods),
		"allow_credentials": config.AllowCredentials,
		"max_age_hours":    config.MaxAge.Hours(),
	}).Info("CORS middleware configured")

	// Log origins in debug mode (don't log in production for security)
	if gin.Mode() != gin.ReleaseMode {
		logrus.WithField("origins", config.AllowedOrigins).Debug("CORS allowed origins")
	}

	return cors.New(corsConfig)
}

// DefaultCORS creates CORS middleware with default configuration
func DefaultCORS() gin.HandlerFunc {
	return CreateCORSMiddleware(DefaultCORSConfig())
}

// ProductionCORS creates CORS middleware with production configuration
func ProductionCORS() gin.HandlerFunc {
	return CreateCORSMiddleware(ProductionCORSConfig())
}

// SmartCORS creates CORS middleware based on environment
func SmartCORS() gin.HandlerFunc {
	if gin.Mode() == gin.ReleaseMode {
		return ProductionCORS()
	}
	return DefaultCORS()
}
