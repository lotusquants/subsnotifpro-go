// Package observability provides enhanced middleware that integrates metrics and tracing
package observability

import (
	"context"
	"fmt"
	"strconv"
	"subsnotifpro-go/internal/metrics"
	"subsnotifpro-go/internal/tracing"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ObservabilityMiddleware combines metrics and tracing functionality
type ObservabilityMiddleware struct {
	metrics *metrics.MetricsRegistry
	tracer  *tracing.TracerProvider
}

// NewObservabilityMiddleware creates a new observability middleware
func NewObservabilityMiddleware(metricsRegistry *metrics.MetricsRegistry, tracerProvider *tracing.TracerProvider) *ObservabilityMiddleware {
	return &ObservabilityMiddleware{
		metrics: metricsRegistry,
		tracer:  tracerProvider,
	}
}

// HTTPMiddleware provides comprehensive HTTP observability
func (om *ObservabilityMiddleware) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Start tracing span
		ctx, span := om.tracer.StartSpanWithAttributes(c.Request.Context(), "http.request", map[string]interface{}{
			"http.method":     c.Request.Method,
			"http.url":        c.Request.URL.String(),
			"http.route":      c.FullPath(),
			"http.user_agent": c.Request.UserAgent(),
			"component":       "http.server",
		})
		
		// Set context with span
		c.Request = c.Request.WithContext(ctx)
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		statusCodeStr := strconv.Itoa(statusCode)
		
		// Record metrics
		if om.metrics != nil {
			om.metrics.HTTPRequestsTotal.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
				statusCodeStr,
			).Inc()
			
			om.metrics.HTTPRequestDuration.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(duration.Seconds())
			
			om.metrics.HTTPResponseSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(c.Writer.Size()))
		}
		
		// Update span with response information
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
			attribute.Int("http.response_size", c.Writer.Size()),
		)
		
		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.SetStatus(codes.Error, "HTTP error")
			if statusCode >= 500 {
				span.RecordError(fmt.Errorf("server error"), trace.WithAttributes(attribute.String("error.type", "server_error")))
			} else {
				span.RecordError(fmt.Errorf("client error"), trace.WithAttributes(attribute.String("error.type", "client_error")))
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}
		
		span.End()
	}
}

// DatabaseMiddleware provides database operation observability
func (om *ObservabilityMiddleware) DatabaseMiddleware(operation, table string) func(context.Context) (context.Context, func()) {
	return func(ctx context.Context) (context.Context, func()) {
		start := time.Now()
		
		// Start tracing span for database operation
		ctx, span := om.tracer.TraceDBQuery(ctx, operation, table)
		
		// Return context and cleanup function
		return ctx, func() {
			duration := time.Since(start)
			
			// Record metrics
			if om.metrics != nil {
				om.metrics.DatabaseQueries.WithLabelValues(
					operation,
					table,
					"success", // This would be determined by error status
				).Inc()
				
				om.metrics.DatabaseQueryLatency.WithLabelValues(
					operation,
					table,
				).Observe(duration.Seconds())
			}
			
			// Update span
			span.SetAttributes(
				attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1e6),
			)
			
			span.End()
		}
	}
}

// MessageProcessingMiddleware provides message queue observability
func (om *ObservabilityMiddleware) MessageProcessingMiddleware(queueName, messageType string) func(context.Context) (context.Context, func(error)) {
	return func(ctx context.Context) (context.Context, func(error)) {
		start := time.Now()
		
		// Start tracing span for message processing
		ctx, span := om.tracer.TraceMessageProcessing(ctx, queueName, messageType)
		
		// Return context and cleanup function
		return ctx, func(err error) {
			duration := time.Since(start)
			status := "success"
			if err != nil {
				status = "error"
			}
			
			// Record metrics
			if om.metrics != nil {
				om.metrics.QueueMessages.WithLabelValues(
					queueName,
					status,
				).Inc()
				
				om.metrics.EventProcessingTime.WithLabelValues(
					messageType,
					"queue", // platform
				).Observe(duration.Seconds())
				
				if err != nil {
					om.metrics.FailedEvents.WithLabelValues(
						messageType,
						"queue",
						"processing_error",
					).Inc()
				} else {
					om.metrics.ProcessedEvents.WithLabelValues(
						messageType,
						"queue",
						"success",
					).Inc()
				}
			}
			
			// Update span
			span.SetAttributes(
				attribute.Float64("message.duration_ms", float64(duration.Nanoseconds())/1e6),
			)
			
			if err != nil {
				span.RecordError(err, trace.WithAttributes(attribute.String("error.description", "Message processing failed")))
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}
			
			span.End()
		}
	}
}

// BusinessOperationMiddleware provides business logic observability
func (om *ObservabilityMiddleware) BusinessOperationMiddleware(operation string, attrs map[string]interface{}) func(context.Context) (context.Context, func(error)) {
	return func(ctx context.Context) (context.Context, func(error)) {
		start := time.Now()
		
		// Start tracing span for business operation
		ctx, span := om.tracer.TraceBusinessOperation(ctx, operation, attrs)
		
		// Return context and cleanup function
		return ctx, func(err error) {
			duration := time.Since(start)
			
			// Record metrics for business operations
			if om.metrics != nil {
				// Record subscription events for subscription-related operations
				if operation == "subscription.create" || operation == "subscription.update" || operation == "subscription.cancel" {
					platform := "unknown"
					if p, ok := attrs["platform"]; ok {
						if ps, ok := p.(string); ok {
							platform = ps
						}
					}
					
					status := "success"
					if err != nil {
						status = "error"
					}
					
					om.metrics.SubscriptionEvents.WithLabelValues(
						operation,
						platform,
						status,
					).Inc()
				}
			}
			
			// Update span
			span.SetAttributes(
				attribute.Float64("business.duration_ms", float64(duration.Nanoseconds())/1e6),
			)
			
			if err != nil {
				span.RecordError(err, trace.WithAttributes(attribute.String("error.description", "Business operation failed")))
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "")
			}
			
			span.End()
		}
	}
}

// AuthenticationMiddleware provides authentication observability
func (om *ObservabilityMiddleware) AuthenticationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get authentication method from headers or context
		authMethod := "unknown"
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				authMethod = "bearer_token"
			} else if len(authHeader) > 6 && authHeader[:6] == "Basic " {
				authMethod = "basic_auth"
			}
		}
		
		// Start tracing span
		ctx, span := om.tracer.StartSpanWithAttributes(c.Request.Context(), "auth.authenticate", map[string]interface{}{
			"auth.method": authMethod,
			"component":   "authentication",
		})
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
		
		duration := time.Since(start)
		
		// Determine authentication status from response
		status := "success"
		if c.Writer.Status() == 401 {
			status = "unauthorized"
		} else if c.Writer.Status() == 403 {
			status = "forbidden"
		}
		
		// Record metrics
		if om.metrics != nil {
			om.metrics.AuthenticationTotal.WithLabelValues(
				authMethod,
				status,
			).Inc()
			
			om.metrics.AuthenticationLatency.WithLabelValues(
				authMethod,
			).Observe(duration.Seconds())
		}
		
		// Update span
		span.SetAttributes(
			attribute.String("auth.status", status),
			attribute.Float64("auth.duration_ms", float64(duration.Nanoseconds())/1e6),
		)
		
		if status != "success" {
			span.SetStatus(codes.Error, "Authentication failed")
		} else {
			span.SetStatus(codes.Ok, "")
		}
		
		span.End()
	}
}

// CorrelationIDMiddleware extracts or creates correlation IDs and adds them to tracing
func (om *ObservabilityMiddleware) CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			// Generate correlation ID from trace ID if available
			if om.tracer != nil {
				correlationID = om.tracer.GetTraceID(c.Request.Context())
			}
			if correlationID == "" {
				correlationID = generateCorrelationID()
			}
		}
		
		// Set correlation ID in response header
		c.Header("X-Correlation-ID", correlationID)
		
		// Add to tracing context
		if om.tracer != nil {
			om.tracer.SetAttributes(c.Request.Context(), map[string]interface{}{
				"correlation.id": correlationID,
			})
		}
		
		c.Next()
	}
}

// generateCorrelationID generates a simple correlation ID
func generateCorrelationID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
