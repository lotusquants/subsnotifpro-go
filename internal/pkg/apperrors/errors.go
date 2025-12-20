package apperrors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a standardized application error
type AppError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Details  string `json:"details,omitempty"`
	Status   int    `json:"-"`
	Internal error  `json:"-"` // Internal error for logging
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Internal.Error())
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Internal
}

// Standard error types
var (
	ErrValidation         = &AppError{Code: "VALIDATION_ERROR", Status: http.StatusBadRequest}
	ErrNotFound           = &AppError{Code: "NOT_FOUND", Status: http.StatusNotFound}
	ErrExternalAPI        = &AppError{Code: "EXTERNAL_API_ERROR", Status: http.StatusBadGateway}
	ErrRateLimit          = &AppError{Code: "RATE_LIMIT", Status: http.StatusTooManyRequests}
	ErrInternal           = &AppError{Code: "INTERNAL_ERROR", Status: http.StatusInternalServerError}
	ErrUnauthorized       = &AppError{Code: "UNAUTHORIZED", Status: http.StatusUnauthorized}
	ErrForbidden          = &AppError{Code: "FORBIDDEN", Status: http.StatusForbidden}
	ErrTimeout            = &AppError{Code: "TIMEOUT", Status: http.StatusGatewayTimeout}
	ErrConflict           = &AppError{Code: "CONFLICT", Status: http.StatusConflict}
	ErrServiceUnavailable = &AppError{Code: "SERVICE_UNAVAILABLE", Status: http.StatusServiceUnavailable}
)

// Error constructors
func ValidationError(message string, details ...string) *AppError {
	err := &AppError{
		Code:    ErrValidation.Code,
		Message: message,
		Status:  ErrValidation.Status,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

func NotFoundError(message string, internal error) *AppError {
	return &AppError{
		Code:     ErrNotFound.Code,
		Message:  message,
		Status:   ErrNotFound.Status,
		Internal: internal,
	}
}

func ExternalAPIError(message string, internal error) *AppError {
	return &AppError{
		Code:     ErrExternalAPI.Code,
		Message:  message,
		Status:   ErrExternalAPI.Status,
		Internal: internal,
	}
}

func InternalError(message string, internal error) *AppError {
	return &AppError{
		Code:     ErrInternal.Code,
		Message:  message,
		Status:   ErrInternal.Status,
		Internal: internal,
	}
}

func TimeoutError(message string, internal error) *AppError {
	return &AppError{
		Code:     ErrTimeout.Code,
		Message:  message,
		Status:   ErrTimeout.Status,
		Internal: internal,
	}
}

func RateLimitError(message string) *AppError {
	return &AppError{
		Code:    ErrRateLimit.Code,
		Message: message,
		Status:  ErrRateLimit.Status,
	}
}

func ConflictError(message string, internal error) *AppError {
	return &AppError{
		Code:     ErrConflict.Code,
		Message:  message,
		Status:   ErrConflict.Status,
		Internal: internal,
	}
}

// Response structures
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Helper functions for standardized responses
func StandardErrorResponse(code, message string, details ...string) ErrorResponse {
	response := ErrorResponse{
		Success: false,
		Error:   message,
		Code:    code,
	}
	if len(details) > 0 {
		response.Details = details[0]
	}
	return response
}

func StandardSuccessResponse(data interface{}, message ...string) SuccessResponse {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}
	if len(message) > 0 {
		response.Message = message[0]
	}
	return response
}

// Gin helper function to handle AppError responses
func HandleAppError(c *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		c.JSON(appErr.Status, StandardErrorResponse(appErr.Code, appErr.Message, appErr.Details))
		return
	}

	// Fallback for unknown errors
	c.JSON(http.StatusInternalServerError, StandardErrorResponse(
		ErrInternal.Code,
		"An unexpected error occurred",
		"",
	))
}

// Validation helpers
func RequiredFieldError(fieldName string) *AppError {
	return ValidationError(
		fmt.Sprintf("Missing required parameter: %s", fieldName),
		fmt.Sprintf("The field '%s' is required and cannot be empty", fieldName),
	)
}

func InvalidFieldError(fieldName, reason string) *AppError {
	return ValidationError(
		fmt.Sprintf("Invalid parameter: %s", fieldName),
		reason,
	)
}

func MultipleFieldErrors(fields []string) *AppError {
	if len(fields) == 1 {
		return RequiredFieldError(fields[0])
	}
	return ValidationError(
		"Missing required parameters",
		fmt.Sprintf("The following fields are required: %v", fields),
	)
}
