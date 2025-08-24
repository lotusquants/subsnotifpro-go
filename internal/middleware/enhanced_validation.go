package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// Enhanced validation middleware for Gin with go-playground/validator

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidationConfig holds validation middleware configuration
type ValidationConfig struct {
	MaxBodySize      int64
	RequireContentType bool
	AllowedContentTypes []string
	ValidateQueryParams bool
	ValidatePathParams  bool
	CustomValidators   map[string]validator.Func
}

// DefaultValidationConfig returns default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxBodySize:        10 * 1024 * 1024, // 10MB
		RequireContentType: true,
		AllowedContentTypes: []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
		},
		ValidateQueryParams: true,
		ValidatePathParams:  true,
		CustomValidators:   make(map[string]validator.Func),
	}
}

// EnhancedValidationMiddleware creates comprehensive validation middleware
func EnhancedValidationMiddleware(config *ValidationConfig) gin.HandlerFunc {
	// Register custom validators
	for tag, fn := range config.CustomValidators {
		validate.RegisterValidation(tag, fn)
	}

	return func(c *gin.Context) {
		// Set request size limit
		if config.MaxBodySize > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxBodySize)
		}

		// Validate content type for POST/PUT/PATCH requests
		if config.RequireContentType && isBodyMethod(c.Request.Method) {
			if err := validateContentType(c, config.AllowedContentTypes); err != nil {
				handleValidationError(c, http.StatusBadRequest, "content_type_validation", err.Error(), nil)
				return
			}
		}

		// Validate query parameters
		if config.ValidateQueryParams {
			if err := validateQueryParameters(c); err != nil {
				handleValidationError(c, http.StatusBadRequest, "query_validation", "Invalid query parameters", err)
				return
			}
		}

		// Validate path parameters
		if config.ValidatePathParams {
			if err := validatePathParameters(c); err != nil {
				handleValidationError(c, http.StatusBadRequest, "path_validation", "Invalid path parameters", err)
				return
			}
		}

		// Security validation
		if err := validateSecurityHeaders(c); err != nil {
			handleValidationError(c, http.StatusBadRequest, "security_validation", err.Error(), nil)
			return
		}

		c.Next()
	}
}

// ValidationErrorDetail represents detailed validation error
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Value   string `json:"value,omitempty"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents validation error response
type ValidationErrorResponse struct {
	Error   string                  `json:"error"`
	Message string                  `json:"message"`
	Type    string                  `json:"type"`
	Details []ValidationErrorDetail `json:"details,omitempty"`
}

// isBodyMethod checks if HTTP method typically includes a body
func isBodyMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}

// validateContentType validates request content type
func validateContentType(c *gin.Context, allowedTypes []string) error {
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		return &ValidationError{
			Field:   "Content-Type",
			Message: "Content-Type header is required",
			Code:    "MISSING_CONTENT_TYPE",
		}
	}

	for _, allowedType := range allowedTypes {
		if strings.Contains(contentType, allowedType) {
			return nil
		}
	}

	return &ValidationError{
		Field:   "Content-Type",
		Message: "Invalid content type. Allowed: " + strings.Join(allowedTypes, ", "),
		Code:    "INVALID_CONTENT_TYPE",
	}
}

// validateQueryParameters validates query parameters for common issues
func validateQueryParameters(c *gin.Context) []ValidationErrorDetail {
	var errors []ValidationErrorDetail
	query := c.Request.URL.Query()

	for key, values := range query {
		for _, value := range values {
			// Check for dangerous characters
			if containsDangerousChars(value) {
				errors = append(errors, ValidationErrorDetail{
					Field:   key,
					Value:   value,
					Tag:     "security",
					Message: "Parameter contains potentially dangerous characters",
				})
				continue
			}

			// Check for SQL injection patterns
			if containsSQLInjection(value) {
				errors = append(errors, ValidationErrorDetail{
					Field:   key,
					Value:   "[REDACTED]",
					Tag:     "sql_injection",
					Message: "Parameter contains potential SQL injection patterns",
				})
				continue
			}

			// Validate specific common parameters
			switch key {
			case "page", "limit", "size", "count":
				if err := validatePositiveInteger(key, value, 1, 1000); err != nil {
					errors = append(errors, *err)
				}
			case "offset", "skip":
				if err := validateNonNegativeInteger(key, value, 0, 1000000); err != nil {
					errors = append(errors, *err)
				}
			case "id", "user_id", "app_id", "subscription_id":
				if err := validateID(key, value); err != nil {
					errors = append(errors, *err)
				}
			case "email":
				if err := validateEmail(key, value); err != nil {
					errors = append(errors, *err)
				}
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validatePathParameters validates path parameters
func validatePathParameters(c *gin.Context) []ValidationErrorDetail {
	var errors []ValidationErrorDetail
	
	// Get path parameters from Gin context
	for _, param := range c.Params {
		// Check for dangerous characters
		if containsDangerousChars(param.Value) {
			errors = append(errors, ValidationErrorDetail{
				Field:   param.Key,
				Value:   param.Value,
				Tag:     "security",
				Message: "Path parameter contains potentially dangerous characters",
			})
			continue
		}

		// Validate ID format for common ID parameters
		if strings.Contains(param.Key, "id") || strings.Contains(param.Key, "Id") {
			if err := validateID(param.Key, param.Value); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateSecurityHeaders validates security-related headers
func validateSecurityHeaders(c *gin.Context) error {
	// Check for suspicious User-Agent
	userAgent := c.GetHeader("User-Agent")
	if userAgent == "" {
		logrus.WithField("path", c.Request.URL.Path).Warn("Request with empty User-Agent")
	} else if isSuspiciousUserAgent(userAgent) {
		return &ValidationError{
			Field:   "User-Agent",
			Message: "Suspicious User-Agent detected",
			Code:    "SUSPICIOUS_USER_AGENT",
		}
	}

	// Check for dangerous headers
	dangerousHeaders := []string{
		"X-Forwarded-Host", "X-Original-URL", "X-Rewrite-URL",
	}
	
	for _, header := range dangerousHeaders {
		if value := c.GetHeader(header); value != "" {
			if containsDangerousChars(value) {
				return &ValidationError{
					Field:   header,
					Message: "Header contains potentially dangerous content",
					Code:    "DANGEROUS_HEADER",
				}
			}
		}
	}

	return nil
}

// Validation helper functions

func validatePositiveInteger(field, value string, min, max int) *ValidationErrorDetail {
	num, err := strconv.Atoi(value)
	if err != nil {
		return &ValidationErrorDetail{
			Field:   field,
			Value:   value,
			Tag:     "integer",
			Message: "Must be a valid integer",
		}
	}
	if num < min || num > max {
		return &ValidationErrorDetail{
			Field:   field,
			Value:   value,
			Tag:     "range",
			Message: "Must be between " + strconv.Itoa(min) + " and " + strconv.Itoa(max),
		}
	}
	return nil
}

func validateNonNegativeInteger(field, value string, min, max int) *ValidationErrorDetail {
	num, err := strconv.Atoi(value)
	if err != nil {
		return &ValidationErrorDetail{
			Field:   field,
			Value:   value,
			Tag:     "integer",
			Message: "Must be a valid integer",
		}
	}
	if num < min || num > max {
		return &ValidationErrorDetail{
			Field:   field,
			Value:   value,
			Tag:     "range",
			Message: "Must be between " + strconv.Itoa(min) + " and " + strconv.Itoa(max),
		}
	}
	return nil
}

func validateID(field, value string) *ValidationErrorDetail {
	if len(value) == 0 || len(value) > 255 {
		return &ValidationErrorDetail{
			Field:   field,
			Value:   value,
			Tag:     "length",
			Message: "ID must be between 1 and 255 characters",
		}
	}

	// Allow alphanumeric, hyphens, underscores, and UUIDs
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	for _, char := range value {
		if !strings.ContainsRune(validChars, char) {
			return &ValidationErrorDetail{
				Field:   field,
				Value:   value,
				Tag:     "format",
				Message: "ID contains invalid characters. Only alphanumeric, hyphens, and underscores allowed",
			}
		}
	}
	return nil
}

func validateEmail(field, value string) *ValidationErrorDetail {
	if err := validate.Var(value, "email"); err != nil {
		return &ValidationErrorDetail{
			Field:   field,
			Value:   value,
			Tag:     "email",
			Message: "Must be a valid email address",
		}
	}
	return nil
}

func containsDangerousChars(value string) bool {
	dangerous := []string{
		"<script", "</script", "javascript:", "data:", "vbscript:",
		"onload=", "onerror=", "onclick=", "eval(", "alert(",
		"../", "..\\", "%2e%2e%2f", "%2e%2e%5c",
		"<iframe", "</iframe", "<object", "</object",
		"<embed", "</embed", "<applet", "</applet",
	}

	lowerValue := strings.ToLower(value)
	for _, danger := range dangerous {
		if strings.Contains(lowerValue, danger) {
			return true
		}
	}
	return false
}

func containsSQLInjection(value string) bool {
	sqlPatterns := []string{
		"'", "\"", ";--", "/*", "*/", " or ", " and ", " union ", " select ",
		" insert ", " update ", " delete ", " drop ", " create ", " alter ",
		" exec ", " execute ", "xp_", "sp_", "0x", "char(", "cast(",
		"convert(", "substring(", "ascii(", "concat(",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerValue, pattern) {
			return true
		}
	}
	return false
}

func isSuspiciousUserAgent(userAgent string) bool {
	suspicious := []string{
		"sqlmap", "nikto", "masscan", "nmap", "whatweb", "dirb", "dirbuster",
		"gobuster", "wfuzz", "burp", "zap", "acunetix", "nessus", "openvas",
		"<script", "javascript:", "data:", "vbscript:", "python-requests",
		"curl", "wget", "bot", "crawler", "spider", "scraper",
	}

	lowerUA := strings.ToLower(userAgent)
	for _, pattern := range suspicious {
		if strings.Contains(lowerUA, pattern) {
			return true
		}
	}
	return false
}

// handleValidationError handles validation errors consistently
func handleValidationError(c *gin.Context, statusCode int, errorType, message string, details []ValidationErrorDetail) {
	response := ValidationErrorResponse{
		Error:   "validation_failed",
		Type:    errorType,
		Message: message,
		Details: details,
	}

	// Log validation error
	logrus.WithFields(logrus.Fields{
		"error_type": errorType,
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
	}).Warn("Validation error")

	c.JSON(statusCode, response)
	c.Abort()
}

// DefaultValidation creates validation middleware with default configuration
func DefaultValidation() gin.HandlerFunc {
	return EnhancedValidationMiddleware(DefaultValidationConfig())
}

// StrictValidation creates validation middleware with strict configuration
func StrictValidation() gin.HandlerFunc {
	config := DefaultValidationConfig()
	config.MaxBodySize = 1024 * 1024 // 1MB for strict validation
	return EnhancedValidationMiddleware(config)
}

// WebhookValidation creates validation middleware for webhook endpoints
func WebhookValidation() gin.HandlerFunc {
	config := DefaultValidationConfig()
	config.RequireContentType = true
	config.AllowedContentTypes = []string{"application/json"}
	config.MaxBodySize = 2 * 1024 * 1024 // 2MB for webhooks
	return EnhancedValidationMiddleware(config)
}
