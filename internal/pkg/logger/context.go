package logger

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ctxKey string

const (
	correlationIDKey ctxKey = "correlation_id"
	userIDKey        ctxKey = "user_id"
	// requestIDKey is already defined in logger.go
)

// CorrelationID helpers
func WithCorrelationID(ctx context.Context) context.Context {
	if GetCorrelationID(ctx) != "" {
		return ctx
	}
	return context.WithValue(ctx, correlationIDKey, uuid.New().String())
}

func WithCorrelationIDValue(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id
	}
	return ""
}

// UserID helpers
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// RequestID helpers - use existing functions from logger.go
func GetRequestIDFromCtx(ctx context.Context) string {
	return GetRequestIDFromContext(ctx)
}

func WithRequestIDCtx(ctx context.Context, requestID string) context.Context {
	return WithRequestID(ctx, requestID)
}

// Enhanced logger that extracts context information
func FromContext(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}

	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		fields["correlation_id"] = correlationID
	}

	if userID := GetUserID(ctx); userID != "" {
		fields["user_id"] = userID
	}

	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		fields["request_id"] = requestID
	}

	return Log.WithFields(fields)
}

// Gin middleware to add correlation ID to context
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add to context
		ctx := WithCorrelationIDValue(c.Request.Context(), correlationID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response header
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

// Request logging middleware
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Log request start
		FromContext(ctx).WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"user_agent": c.GetHeader("User-Agent"),
			"ip":         c.ClientIP(),
		}).Info("Request started")

		c.Next()

		// Log request completion
		FromContext(ctx).WithFields(logrus.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
			"size":   c.Writer.Size(),
		}).Info("Request completed")
	}
}

// Business context helpers
func WithPackageName(ctx context.Context, packageName string) context.Context {
	return context.WithValue(ctx, "package_name", packageName)
}

func WithProductID(ctx context.Context, productID string) context.Context {
	return context.WithValue(ctx, "product_id", productID)
}

func WithTransactionID(ctx context.Context, transactionID string) context.Context {
	return context.WithValue(ctx, "transaction_id", transactionID)
}

func WithBusinessContext(ctx context.Context) *logrus.Entry {
	entry := FromContext(ctx)

	if packageName, ok := ctx.Value("package_name").(string); ok && packageName != "" {
		entry = entry.WithField("package_name", packageName)
	}

	if productID, ok := ctx.Value("product_id").(string); ok && productID != "" {
		entry = entry.WithField("product_id", productID)
	}

	if transactionID, ok := ctx.Value("transaction_id").(string); ok && transactionID != "" {
		entry = entry.WithField("transaction_id", transactionID)
	}

	return entry
}

// Error logging helpers
func LogError(ctx context.Context, err error, message string, fields ...logrus.Fields) {
	entry := FromContext(ctx).WithError(err)

	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}

	entry.Error(message)
}

func LogWarning(ctx context.Context, message string, fields ...logrus.Fields) {
	entry := FromContext(ctx)

	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}

	entry.Warn(message)
}

func LogInfo(ctx context.Context, message string, fields ...logrus.Fields) {
	entry := FromContext(ctx)

	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}

	entry.Info(message)
}

func LogDebug(ctx context.Context, message string, fields ...logrus.Fields) {
	entry := FromContext(ctx)

	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}

	entry.Debug(message)
}

// Utility function to mask sensitive data for logging
func MaskSensitiveData(data string, visibleChars int) string {
	if len(data) <= visibleChars {
		return "***"
	}
	return data[:visibleChars] + "***"
}
