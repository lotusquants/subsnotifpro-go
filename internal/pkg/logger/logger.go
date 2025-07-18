// internal/google_playstore/logger/logger.go
package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

// Log is the global structured logger
var Log = logrus.New()

func init() {
	Log.SetFormatter(&logrus.JSONFormatter{}) // ✅ JSON format for logs
	Log.SetOutput(os.Stdout)                  // ✅ Log to console
	Log.SetLevel(logrus.InfoLevel)            // ✅ Default level: INFO
}

// Enhanced logger with structured logging capabilities and context support
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithFields(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// F is a convenience function to create a Field
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// enhancedLogger implements the Logger interface with context support
type enhancedLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewEnhancedLogger creates a new enhanced logger with best practices
func NewEnhancedLogger() Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
	})
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.InfoLevel)
	l.SetReportCaller(true)

	return &enhancedLogger{logger: l, entry: l.WithFields(logrus.Fields{})}
}

// Debug logs a debug message with optional fields
func (l *enhancedLogger) Debug(msg string, fields ...Field) {
	l.entry.WithFields(l.convertFields(fields)).Debug(msg)
}

// Info logs an info message with optional fields
func (l *enhancedLogger) Info(msg string, fields ...Field) {
	l.entry.WithFields(l.convertFields(fields)).Info(msg)
}

// Warn logs a warning message with optional fields
func (l *enhancedLogger) Warn(msg string, fields ...Field) {
	l.entry.WithFields(l.convertFields(fields)).Warn(msg)
}

// Error logs an error message with optional fields
func (l *enhancedLogger) Error(msg string, fields ...Field) {
	l.entry.WithFields(l.convertFields(fields)).Error(msg)
}

// Fatal logs a fatal message with optional fields and exits
func (l *enhancedLogger) Fatal(msg string, fields ...Field) {
	l.entry.WithFields(l.convertFields(fields)).Fatal(msg)
}

// WithFields returns a new logger with the provided fields
func (l *enhancedLogger) WithFields(fields ...Field) Logger {
	return &enhancedLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(l.convertFields(fields)),
	}
}

// WithContext returns a new logger with context
func (l *enhancedLogger) WithContext(ctx context.Context) Logger {
	entry := l.entry.WithContext(ctx)
	
	// Add request ID if available
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		entry = entry.WithField("request_id", requestID)
	}
	
	return &enhancedLogger{
		logger: l.logger,
		entry:  entry,
	}
}

// convertFields converts Field slice to logrus.Fields
func (l *enhancedLogger) convertFields(fields []Field) logrus.Fields {
	logrusFields := make(logrus.Fields, len(fields))
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}
	return logrusFields
}

// Context utility functions
type contextKey string

const requestIDKey contextKey = "request_id"

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}
