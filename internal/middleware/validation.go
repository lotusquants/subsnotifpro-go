package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}

// ValidationResponse represents validation error response
type ValidationResponse struct {
	ErrorType string            `json:"error"`
	Message   string            `json:"message"`
	Details   []ValidationError `json:"details,omitempty"`
}

// Error implements the error interface
func (r *ValidationResponse) Error() string {
	return r.Message
}

// InputValidator provides request validation functionality
type InputValidator struct {
	maxBodySize int64
}

// NewInputValidator creates a new input validator
func NewInputValidator(maxBodySize int64) *InputValidator {
	if maxBodySize <= 0 {
		maxBodySize = 10 * 1024 * 1024 // 10MB default
	}
	return &InputValidator{
		maxBodySize: maxBodySize,
	}
}

// ValidateRequest validates incoming requests
func (v *InputValidator) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set request size limit
		r.Body = http.MaxBytesReader(w, r.Body, v.maxBodySize)

		// Validate content type for POST/PUT requests
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			if err := v.validateContentType(r); err != nil {
				v.writeValidationError(w, http.StatusBadRequest, err)
				return
			}
		}

		// Validate path parameters
		if err := v.validatePathParams(r); err != nil {
			v.writeValidationError(w, http.StatusBadRequest, err)
			return
		}

		// Validate query parameters
		if err := v.validateQueryParams(r); err != nil {
			v.writeValidationError(w, http.StatusBadRequest, err)
			return
		}

		// Add validation context
		ctx := context.WithValue(r.Context(), "validation", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateContentType validates request content type
func (v *InputValidator) validateContentType(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return &ValidationError{
			Field:   "Content-Type",
			Message: "Content-Type header is required",
			Code:    "MISSING_CONTENT_TYPE",
		}
	}

	// Allow common content types
	validTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
	}

	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			return nil
		}
	}

	return &ValidationError{
		Field:   "Content-Type",
		Message: "Invalid content type. Supported types: " + strings.Join(validTypes, ", "),
		Code:    "INVALID_CONTENT_TYPE",
	}
}

// validatePathParams validates path parameters
func (v *InputValidator) validatePathParams(r *http.Request) error {
	// Extract path parameters from URL path
	// This is a simplified implementation - in real usage, you would integrate with your router
	var errors []ValidationError

	// Check for dangerous characters in the path
	if v.containsDangerousChars(r.URL.Path) {
		errors = append(errors, ValidationError{
			Field:   "path",
			Message: "Path contains invalid characters",
			Code:    "INVALID_PATH",
		})
	}

	if len(errors) > 0 {
		return &ValidationResponse{
			ErrorType: "validation_failed",
			Message:   "Path parameter validation failed",
			Details:   errors,
		}
	}

	return nil
}

// validateQueryParams validates query parameters
func (v *InputValidator) validateQueryParams(r *http.Request) error {
	query := r.URL.Query()
	var errors []ValidationError

	for key, values := range query {
		// Check for SQL injection patterns
		for _, value := range values {
			if v.containsSQLInjection(value) {
				errors = append(errors, ValidationError{
					Field:   key,
					Message: "Query parameter contains potentially malicious content",
					Code:    "MALICIOUS_QUERY_PARAM",
				})
			}
		}

		// Validate specific parameters
		switch key {
		case "limit":
			if len(values) > 0 && !v.isValidNumber(values[0], 1, 1000) {
				errors = append(errors, ValidationError{
					Field:   key,
					Message: "Limit must be between 1 and 1000",
					Code:    "INVALID_LIMIT",
				})
			}
		case "offset":
			if len(values) > 0 && !v.isValidNumber(values[0], 0, 1000000) {
				errors = append(errors, ValidationError{
					Field:   key,
					Message: "Offset must be between 0 and 1000000",
					Code:    "INVALID_OFFSET",
				})
			}
		}
	}

	if len(errors) > 0 {
		return &ValidationResponse{
			ErrorType: "validation_failed",
			Message:   "Query parameter validation failed",
			Details:   errors,
		}
	}

	return nil
}

// containsDangerousChars checks for dangerous characters
func (v *InputValidator) containsDangerousChars(value string) bool {
	dangerous := []string{
		"<", ">", "script", "javascript:", "data:", "vbscript:",
		"onload", "onerror", "onclick", "../", "..\\", "eval(",
	}

	lowerValue := strings.ToLower(value)
	for _, danger := range dangerous {
		if strings.Contains(lowerValue, danger) {
			return true
		}
	}
	return false
}

// containsSQLInjection checks for common SQL injection patterns
func (v *InputValidator) containsSQLInjection(value string) bool {
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "union", "select",
		"insert", "update", "delete", "drop", "create", "alter",
		"exec", "execute", "xp_", "sp_", "0x", "char(",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerValue, pattern) {
			return true
		}
	}
	return false
}

// isValidID checks if value is a valid ID format
func (v *InputValidator) isValidID(value string) bool {
	if len(value) == 0 || len(value) > 255 {
		return false
	}
	
	// Allow alphanumeric, hyphens, and underscores
	for _, char := range value {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return false
		}
	}
	return true
}

// isValidNumber checks if value is a valid number within range
func (v *InputValidator) isValidNumber(value string, min, max int) bool {
	if len(value) == 0 {
		return false
	}
	
	num := 0
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
		num = num*10 + int(char-'0')
		if num > max {
			return false
		}
	}
	return num >= min
}

// writeValidationError writes validation error response
func (v *InputValidator) writeValidationError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var response ValidationResponse
	
	if valErr, ok := err.(*ValidationError); ok {
		response = ValidationResponse{
			ErrorType: "validation_failed",
			Message:   valErr.Message,
			Details:   []ValidationError{*valErr},
		}
	} else if valResp, ok := err.(*ValidationResponse); ok {
		response = *valResp
	} else {
		response = ValidationResponse{
			ErrorType: "validation_failed",
			Message:   err.Error(),
		}
	}

	json.NewEncoder(w).Encode(response)
}
