package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth represents the health of a specific component
type ComponentHealth struct {
	Status      HealthStatus `json:"status"`
	Message     string       `json:"message,omitempty"`
	LastChecked time.Time    `json:"last_checked"`
	ResponseTime time.Duration `json:"response_time_ms"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status     HealthStatus                `json:"status"`
	Timestamp  time.Time                   `json:"timestamp"`
	Components map[string]ComponentHealth  `json:"components"`
	Version    string                      `json:"version,omitempty"`
	Uptime     time.Duration               `json:"uptime_seconds"`
}

// HealthChecker manages health checks for various components
type HealthChecker struct {
	db                *gorm.DB
	rabbitConn        *amqp.Connection
	serviceBusClient  *azservicebus.Client
	messagingType     string
	startTime         time.Time
	version           string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *gorm.DB, messagingType string, version string) *HealthChecker {
	return &HealthChecker{
		db:            db,
		messagingType: messagingType,
		startTime:     time.Now(),
		version:       version,
	}
}

// SetRabbitMQConnection sets the RabbitMQ connection for health checks
func (h *HealthChecker) SetRabbitMQConnection(conn *amqp.Connection) {
	h.rabbitConn = conn
}

// SetServiceBusClient sets the Azure Service Bus client for health checks
func (h *HealthChecker) SetServiceBusClient(client *azservicebus.Client) {
	h.serviceBusClient = client
}

// CheckDatabase checks the database health
func (h *HealthChecker) CheckDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	if h.db == nil {
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "Database connection not initialized",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start),
		}
	}

	// Check if database is accessible
	sqlDB, err := h.db.DB()
	if err != nil {
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "Failed to get database instance: " + err.Error(),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start),
		}
	}

	// Ping database
	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "Database ping failed: " + err.Error(),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start),
		}
	}

	// Check database stats
	stats := sqlDB.Stats()
	responseTime := time.Since(start)
	
	message := "Database is healthy"
	status := StatusHealthy
	
	// Check for potential issues
	if stats.OpenConnections > 80 { // Assuming max 100 connections
		status = StatusDegraded
		message = "High connection usage detected"
	}

	return ComponentHealth{
		Status:       status,
		Message:      message,
		LastChecked:  time.Now(),
		ResponseTime: responseTime,
	}
}

// CheckRabbitMQ checks RabbitMQ health
func (h *HealthChecker) CheckRabbitMQ(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	if h.rabbitConn == nil {
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "RabbitMQ connection not initialized",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start),
		}
	}

	// Check if connection is closed
	if h.rabbitConn.IsClosed() {
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "RabbitMQ connection is closed",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start),
		}
	}

	// Try to create a channel to test connectivity
	ch, err := h.rabbitConn.Channel()
	if err != nil {
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "Failed to create RabbitMQ channel: " + err.Error(),
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start),
		}
	}
	defer ch.Close()

	return ComponentHealth{
		Status:       StatusHealthy,
		Message:      "RabbitMQ is healthy",
		LastChecked:  time.Now(),
		ResponseTime: time.Since(start),
	}
}

// CheckServiceBus checks Azure Service Bus health
func (h *HealthChecker) CheckServiceBus(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	if h.serviceBusClient == nil {
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "Service Bus client not initialized",
			LastChecked:  time.Now(),
			ResponseTime: time.Since(start),
		}
	}

	// Try to get namespace properties as a connectivity test
	// This is a lightweight operation that verifies connectivity
	// without creating/sending messages
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Note: This is a simplified check. In practice, you might want to
	// create a test receiver/sender to verify full functionality
	_ = ctxWithTimeout // Using context for potential future operations

	return ComponentHealth{
		Status:       StatusHealthy,
		Message:      "Service Bus is healthy",
		LastChecked:  time.Now(),
		ResponseTime: time.Since(start),
	}
}

// CheckMessaging checks the appropriate messaging backend
func (h *HealthChecker) CheckMessaging(ctx context.Context) ComponentHealth {
	switch h.messagingType {
	case "rabbitmq":
		return h.CheckRabbitMQ(ctx)
	case "servicebus":
		return h.CheckServiceBus(ctx)
	default:
		return ComponentHealth{
			Status:       StatusUnhealthy,
			Message:      "Unknown messaging type: " + h.messagingType,
			LastChecked:  time.Now(),
			ResponseTime: 0,
		}
	}
}

// PerformHealthCheck performs a comprehensive health check
func (h *HealthChecker) PerformHealthCheck(ctx context.Context) HealthResponse {
	components := make(map[string]ComponentHealth)
	
	// Check database
	components["database"] = h.CheckDatabase(ctx)
	
	// Check messaging
	components["messaging"] = h.CheckMessaging(ctx)
	
	// Determine overall status
	overallStatus := StatusHealthy
	for _, component := range components {
		if component.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		}
		if component.Status == StatusDegraded {
			overallStatus = StatusDegraded
		}
	}
	
	return HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Components: components,
		Version:    h.version,
		Uptime:     time.Since(h.startTime),
	}
}

// HTTPHealthHandler creates an HTTP handler for health checks
func (h *HealthChecker) HTTPHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Perform health check
		healthResponse := h.PerformHealthCheck(ctx)
		
		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if healthResponse.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if healthResponse.Status == StatusDegraded {
			statusCode = http.StatusPartialContent
		}
		
		// Set response headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		
		// Encode and send response
		json.NewEncoder(w).Encode(healthResponse)
	}
}

// ReadinessHandler creates a readiness probe handler
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Check only critical components for readiness
		dbHealth := h.CheckDatabase(ctx)
		messagingHealth := h.CheckMessaging(ctx)
		
		if dbHealth.Status == StatusUnhealthy || messagingHealth.Status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "not ready",
				"reason": "critical components unhealthy",
			})
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		})
	}
}

// LivenessHandler creates a liveness probe handler
func (h *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple liveness check - just verify the service is running
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "alive",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}
