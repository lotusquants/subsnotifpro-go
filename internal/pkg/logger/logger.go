// internal/google_playstore/logger/logger.go
package logger

import (
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

// // internal/logger/logger.go
// package logger

// import (
// 	"context"
// 	"io"
// 	"os"
// 	"runtime"
// 	"strings"

// 	"github.com/sirupsen/logrus"
// )

// // Logger interface defines the contract for our logger
// type Logger interface {
// 	Debug(ctx context.Context, args ...interface{})
// 	Debugf(ctx context.Context, format string, args ...interface{})
// 	Info(ctx context.Context, args ...interface{})
// 	Infof(ctx context.Context, format string, args ...interface{})
// 	Warn(ctx context.Context, args ...interface{})
// 	Warnf(ctx context.Context, format string, args ...interface{})
// 	Error(ctx context.Context, args ...interface{})
// 	Errorf(ctx context.Context, format string, args ...interface{})
// 	Fatal(ctx context.Context, args ...interface{})
// 	Fatalf(ctx context.Context, format string, args ...interface{})
// 	Panic(ctx context.Context, args ...interface{})
// 	Panicf(ctx context.Context, format string, args ...interface{})

// 	WithFields(ctx context.Context, fields Fields) Logger
// 	SetLevel(level Level)
// 	GetLevel() Level
// }

// // Fields type for structured logging
// type Fields map[string]interface{}

// // Level represents log level
// type Level uint32

// const (
// 	PanicLevel Level = iota
// 	FatalLevel
// 	ErrorLevel
// 	WarnLevel
// 	InfoLevel
// 	DebugLevel
// 	TraceLevel
// )

// // logrusLogger wraps logrus.Logger and implements Logger interface
// type logrusLogger struct {
// 	logger *logrus.Logger
// }

// // New creates a new logger instance
// func New(options ...Option) Logger {
// 	// Default configuration
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.JSONFormatter{
// 		TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
// 		FieldMap: logrus.FieldMap{
// 			logrus.FieldKeyTime:  "timestamp",
// 			logrus.FieldKeyLevel: "severity",
// 			logrus.FieldKeyMsg:   "message",
// 		},
// 	})
// 	l.SetOutput(os.Stdout)
// 	l.SetLevel(logrus.InfoLevel)

// 	// Apply options
// 	logger := &logrusLogger{logger: l}
// 	for _, opt := range options {
// 		opt(logger)
// 	}

// 	return logger
// }

// // Option configures the logger
// type Option func(*logrusLogger)

// // WithOutput sets the output destination
// func WithOutput(w io.Writer) Option {
// 	return func(l *logrusLogger) {
// 		l.logger.SetOutput(w)
// 	}
// }

// // WithLevel sets the log level
// func WithLevel(level Level) Option {
// 	return func(l *logrusLogger) {
// 		l.logger.SetLevel(logrus.Level(level))
// 	}
// }

// // WithFormatter sets the formatter
// func WithFormatter(formatter logrus.Formatter) Option {
// 	return func(l *logrusLogger) {
// 		l.logger.SetFormatter(formatter)
// 	}
// }

// // WithCaller enables caller information in logs
// func WithCaller(enabled bool) Option {
// 	return func(l *logrusLogger) {
// 		if enabled {
// 			l.logger.SetReportCaller(true)
// 		}
// 	}
// }

// // Implementation of Logger interface methods
// func (l *logrusLogger) Debug(ctx context.Context, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Debug(args...)
// }

// func (l *logrusLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Debugf(format, args...)
// }

// func (l *logrusLogger) Info(ctx context.Context, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Info(args...)
// }

// func (l *logrusLogger) Infof(ctx context.Context, format string, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Infof(format, args...)
// }

// func (l *logrusLogger) Warn(ctx context.Context, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Warn(args...)
// }

// func (l *logrusLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Warnf(format, args...)
// }

// func (l *logrusLogger) Error(ctx context.Context, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Error(args...)
// }

// func (l *logrusLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Errorf(format, args...)
// }

// func (l *logrusLogger) Fatal(ctx context.Context, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Fatal(args...)
// }

// func (l *logrusLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Fatalf(format, args...)
// }

// func (l *logrusLogger) Panic(ctx context.Context, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Panic(args...)
// }

// func (l *logrusLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
// 	l.logger.WithContext(ctx).WithFields(l.getFields(ctx)).Panicf(format, args...)
// }

// func (l *logrusLogger) WithFields(ctx context.Context, fields Fields) Logger {
// 	return &logrusEntry{
// 		entry: l.logger.WithContext(ctx).WithFields(logrus.Fields(fields)),
// 	}
// }

// func (l *logrusLogger) SetLevel(level Level) {
// 	l.logger.SetLevel(logrus.Level(level))
// }

// func (l *logrusLogger) GetLevel() Level {
// 	return Level(l.logger.GetLevel())
// }

// // getFields extracts common fields from context
// func (l *logrusLogger) getFields(ctx context.Context) logrus.Fields {
// 	fields := logrus.Fields{}

// 	// Add request ID if available
// 	if requestID := GetRequestID(ctx); requestID != "" {
// 		fields["request_id"] = requestID
// 	}

// 	// Add caller information if enabled
// 	if l.logger.ReportCaller {
// 		if pc, file, line, ok := runtime.Caller(2); ok {
// 			funcName := runtime.FuncForPC(pc).Name()
// 			fields["caller"] = strings.TrimPrefix(file, os.Getenv("GOPATH")+"/src/")
// 			fields["line"] = line
// 			fields["func"] = funcName
// 		}
// 	}

// 	return fields
// }

// // logrusEntry wraps logrus.Entry and implements Logger interface
// type logrusEntry struct {
// 	entry *logrus.Entry
// }

// // Debug implements Logger.
// func (l *logrusEntry) Debug(ctx context.Context, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Debugf implements Logger.
// func (l *logrusEntry) Debugf(ctx context.Context, format string, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Error implements Logger.
// func (l *logrusEntry) Error(ctx context.Context, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Errorf implements Logger.
// func (l *logrusEntry) Errorf(ctx context.Context, format string, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Fatal implements Logger.
// func (l *logrusEntry) Fatal(ctx context.Context, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Fatalf implements Logger.
// func (l *logrusEntry) Fatalf(ctx context.Context, format string, args ...interface{}) {
// 	panic("unimplemented")
// }

// // GetLevel implements Logger.
// func (l *logrusEntry) GetLevel() Level {
// 	panic("unimplemented")
// }

// // Info implements Logger.
// func (l *logrusEntry) Info(ctx context.Context, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Infof implements Logger.
// func (l *logrusEntry) Infof(ctx context.Context, format string, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Panic implements Logger.
// func (l *logrusEntry) Panic(ctx context.Context, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Panicf implements Logger.
// func (l *logrusEntry) Panicf(ctx context.Context, format string, args ...interface{}) {
// 	panic("unimplemented")
// }

// // SetLevel implements Logger.
// func (l *logrusEntry) SetLevel(level Level) {
// 	panic("unimplemented")
// }

// // Warn implements Logger.
// func (l *logrusEntry) Warn(ctx context.Context, args ...interface{}) {
// 	panic("unimplemented")
// }

// // Warnf implements Logger.
// func (l *logrusEntry) Warnf(ctx context.Context, format string, args ...interface{}) {
// 	panic("unimplemented")
// }

// // WithFields implements Logger.
// func (l *logrusEntry) WithFields(ctx context.Context, fields Fields) Logger {
// 	panic("unimplemented")
// }

// // Implement Logger interface methods for logrusEntry...
// // (similar to logrusLogger methods but using entry instead of logger)

// // Context keys
// type contextKey string

// const requestIDKey contextKey = "request_id"

// // GetRequestID gets request ID from context
// func GetRequestID(ctx context.Context) string {
// 	if id, ok := ctx.Value(requestIDKey).(string); ok {
// 		return id
// 	}
// 	return ""
// }

// // WithRequestID adds request ID to context
// func WithRequestID(ctx context.Context, requestID string) context.Context {
// 	return context.WithValue(ctx, requestIDKey, requestID)
// }
