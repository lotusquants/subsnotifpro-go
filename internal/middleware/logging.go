package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RequestLoggingConfig holds configuration for request logging
type RequestLoggingConfig struct {
	LogRequestBody      bool
	LogResponseBody     bool
	LogHeaders          bool
	MaxBodySize         int64
	SensitiveHeaders    []string
	SensitiveFields     []string
	SkipPaths          []string
	LogLevel           logrus.Level
	EnableCorrelationID bool
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() *RequestLoggingConfig {
	return &RequestLoggingConfig{
		LogRequestBody:      true,
		LogResponseBody:     true,
		LogHeaders:          true,
		MaxBodySize:         1024 * 10, // 10KB max body logging
		SensitiveHeaders:    []string{"Authorization", "Cookie", "Set-Cookie", "X-API-Key"},
		SensitiveFields:     []string{"password", "secret", "key", "token", "credential"},
		SkipPaths:          []string{"/api/health", "/metrics"},
		LogLevel:           logrus.InfoLevel,
		EnableCorrelationID: true,
	}
}

// ProductionLoggingConfig returns production logging configuration
func ProductionLoggingConfig() *RequestLoggingConfig {
	config := DefaultLoggingConfig()
	config.LogRequestBody = false  // Don't log request bodies in production
	config.LogResponseBody = false // Don't log response bodies in production
	config.LogHeaders = false      // Don't log headers in production
	config.LogLevel = logrus.WarnLevel
	return config
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// sanitizeHeaders removes sensitive headers from logging
func sanitizeHeaders(headers map[string][]string, sensitiveHeaders []string) map[string]interface{} {
	sanitized := make(map[string]interface{})
	
	for key, values := range headers {
		isSensitive := false
		keyLower := strings.ToLower(key)
		
		for _, sensitive := range sensitiveHeaders {
			if strings.ToLower(sensitive) == keyLower {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			sanitized[key] = "[REDACTED]"
		} else {
			if len(values) == 1 {
				sanitized[key] = values[0]
			} else {
				sanitized[key] = values
			}
		}
	}
	
	return sanitized
}

// sanitizeBody removes sensitive fields from JSON body
func sanitizeBody(body []byte, sensitiveFields []string) interface{} {
	if len(body) == 0 {
		return nil
	}
	
	// Try to parse as JSON
	var jsonBody interface{}
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		// If not JSON, return truncated string
		bodyStr := string(body)
		if len(bodyStr) > 200 {
			return bodyStr[:200] + "..."
		}
		return bodyStr
	}
	
	// Recursively sanitize JSON
	return sanitizeJSONValue(jsonBody, sensitiveFields)
}

// sanitizeJSONValue recursively sanitizes JSON values
func sanitizeJSONValue(value interface{}, sensitiveFields []string) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, val := range v {
			keyLower := strings.ToLower(key)
			isSensitive := false
			
			for _, sensitive := range sensitiveFields {
				if strings.Contains(keyLower, strings.ToLower(sensitive)) {
					isSensitive = true
					break
				}
			}
			
			if isSensitive {
				sanitized[key] = "[REDACTED]"
			} else {
				sanitized[key] = sanitizeJSONValue(val, sensitiveFields)
			}
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, val := range v {
			sanitized[i] = sanitizeJSONValue(val, sensitiveFields)
		}
		return sanitized
	default:
		return v
	}
}

// shouldSkipPath checks if the path should be skipped from logging
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// RequestLogger creates a request logging middleware
func RequestLogger(config *RequestLoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for certain paths
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}
		
		start := time.Now()
		
		// Generate correlation ID if enabled
		var correlationID string
		if config.EnableCorrelationID {
			correlationID = c.GetHeader("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.New().String()
				c.Header("X-Correlation-ID", correlationID)
			}
			c.Set("correlation_id", correlationID)
		}
		
		// Create base log fields
		logFields := logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
			"ip":     c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}
		
		if correlationID != "" {
			logFields["correlation_id"] = correlationID
		}
		
		// Log request headers if enabled
		if config.LogHeaders {
			headers := make(map[string][]string)
			for key, values := range c.Request.Header {
				headers[key] = values
			}
			logFields["request_headers"] = sanitizeHeaders(headers, config.SensitiveHeaders)
		}
		
		// Read and log request body if enabled
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			
			if len(requestBody) > 0 && int64(len(requestBody)) <= config.MaxBodySize {
				logFields["request_body"] = sanitizeBody(requestBody, config.SensitiveFields)
			}
		}
		
		// Wrap response writer to capture response body
		var respWriter *responseBodyWriter
		if config.LogResponseBody {
			respWriter = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:          bytes.NewBufferString(""),
			}
			c.Writer = respWriter
		}
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Add response details
		logFields["status"] = c.Writer.Status()
		logFields["duration_ms"] = duration.Milliseconds()
		logFields["response_size"] = c.Writer.Size()
		
		// Log response body if enabled
		if config.LogResponseBody && respWriter != nil {
			responseBody := respWriter.body.Bytes()
			if len(responseBody) > 0 && int64(len(responseBody)) <= config.MaxBodySize {
				logFields["response_body"] = sanitizeBody(responseBody, config.SensitiveFields)
			}
		}
		
		// Log errors if any
		if len(c.Errors) > 0 {
			logFields["errors"] = c.Errors.String()
		}
		
		// Choose log level based on status code
		level := config.LogLevel
		if c.Writer.Status() >= 500 {
			level = logrus.ErrorLevel
		} else if c.Writer.Status() >= 400 {
			level = logrus.WarnLevel
		}
		
		// Log the request
		message := fmt.Sprintf("%s %s %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
		logrus.WithFields(logFields).Log(level, message)
	}
}

// DefaultRequestLogger creates request logging middleware with default configuration
func DefaultRequestLogger() gin.HandlerFunc {
	return RequestLogger(DefaultLoggingConfig())
}

// ProductionRequestLogger creates request logging middleware with production configuration
func ProductionRequestLogger() gin.HandlerFunc {
	return RequestLogger(ProductionLoggingConfig())
}

// SmartRequestLogger creates request logging middleware based on environment
func SmartRequestLogger() gin.HandlerFunc {
	if gin.Mode() == gin.ReleaseMode {
		return ProductionRequestLogger()
	}
	return DefaultRequestLogger()
}
