// Package routes provides example implementations of observability integration
package routes

import (
	"context"
	"subsnotifpro-go/internal/observability"
	"time"

	"github.com/gin-gonic/gin"
)

// ObservabilityExampleHandler demonstrates how to integrate observability in handlers
type ObservabilityExampleHandler struct {
	observability *observability.ObservabilityMiddleware
}

// NewObservabilityExampleHandler creates a new example handler
func NewObservabilityExampleHandler(obs *observability.ObservabilityMiddleware) *ObservabilityExampleHandler {
	return &ObservabilityExampleHandler{
		observability: obs,
	}
}

// ProcessSubscriptionEvent demonstrates business operation tracing
func (h *ObservabilityExampleHandler) ProcessSubscriptionEvent(c *gin.Context) {
	// Extract business operation details
	platform := c.Query("platform")
	if platform == "" {
		platform = "unknown"
	}

	// Start business operation tracing
	ctx, finish := h.observability.BusinessOperationMiddleware("subscription.process", map[string]interface{}{
		"platform":    platform,
		"endpoint":    c.FullPath(),
		"user_agent": c.Request.UserAgent(),
	})(c.Request.Context())

	// Simulate business logic
	err := h.processBusinessLogic(ctx, platform)

	// Finish tracing (this will record metrics and close the span)
	finish(err)

	if err != nil {
		c.JSON(500, gin.H{"error": "processing failed"})
		return
	}

	c.JSON(200, gin.H{
		"status":   "success",
		"platform": platform,
		"message":  "subscription event processed successfully",
	})
}

// DatabaseExample demonstrates database operation tracing
func (h *ObservabilityExampleHandler) DatabaseExample(c *gin.Context) {
	// Start database operation tracing
	ctx, finish := h.observability.DatabaseMiddleware("SELECT", "subscriptions")(c.Request.Context())

	// Simulate database operation
	result, err := h.performDatabaseQuery(ctx)

	// Finish tracing
	finish()

	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}

	c.JSON(200, gin.H{
		"status": "success",
		"result": result,
	})
}

// MessageProcessingExample demonstrates message queue operation tracing
func (h *ObservabilityExampleHandler) MessageProcessingExample(c *gin.Context) {
	queueName := c.Param("queue")
	messageType := c.Param("type")

	// Start message processing tracing
	ctx, finish := h.observability.MessageProcessingMiddleware(queueName, messageType)(c.Request.Context())

	// Simulate message processing
	err := h.processMessage(ctx, queueName, messageType)

	// Finish tracing
	finish(err)

	if err != nil {
		c.JSON(500, gin.H{"error": "message processing failed"})
		return
	}

	c.JSON(200, gin.H{
		"status":       "success",
		"queue":        queueName,
		"message_type": messageType,
	})
}

// processBusinessLogic simulates business logic with tracing
func (h *ObservabilityExampleHandler) processBusinessLogic(ctx context.Context, platform string) error {
	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	// You could add custom tracing events here
	// span := trace.SpanFromContext(ctx)
	// span.AddEvent("validation_complete")

	return nil
}

// performDatabaseQuery simulates a database query
func (h *ObservabilityExampleHandler) performDatabaseQuery(ctx context.Context) (map[string]interface{}, error) {
	// Simulate database query time
	time.Sleep(20 * time.Millisecond)

	return map[string]interface{}{
		"count": 42,
		"rows":  []string{"row1", "row2", "row3"},
	}, nil
}

// processMessage simulates message processing
func (h *ObservabilityExampleHandler) processMessage(ctx context.Context, queueName, messageType string) error {
	// Simulate message processing time
	time.Sleep(30 * time.Millisecond)

	// Simulate conditional error
	if messageType == "error" {
		return context.DeadlineExceeded
	}

	return nil
}
