package circuitbreaker

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// Middleware creates a circuit breaker middleware for HTTP handlers
func Middleware(manager *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add circuit breaker manager to context
		c.Set("circuitbreaker", manager)
		c.Next()
	}
}

// WithCircuitBreaker wraps a handler with circuit breaker protection
func WithCircuitBreaker(manager *Manager, serviceName string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		
		result, err := manager.ExecuteWithBreaker(ctx, serviceName, func() (interface{}, error) {
			// Create a response recorder to capture the handler's response
			rec := &responseRecorder{
				ResponseWriter: c.Writer,
				statusCode:     http.StatusOK,
			}
			c.Writer = rec
			
			// Execute the handler
			handler(c)
			
			// Check if the response indicates an error
			if rec.statusCode >= 500 {
				return nil, &HTTPError{StatusCode: rec.statusCode}
			}
			
			return nil, nil
		})
		
		if err != nil {
			// Handle circuit breaker errors - check if it's a specific gobreaker error
			if err.Error() == "circuit breaker is open" {
				handleCircuitBreakerError(c, serviceName, err)
				return
			}
			
			// Handle HTTP errors
			if httpError, ok := err.(*HTTPError); ok {
				c.JSON(httpError.StatusCode, gin.H{
					"error":   "Service temporarily unavailable",
					"service": serviceName,
					"code":    httpError.StatusCode,
				})
				return
			}
			
			// Handle other errors
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"service": serviceName,
			})
			return
		}
		
		// Success case is handled by the original handler
		_ = result
	}
}

// GetFromContext retrieves the circuit breaker manager from Gin context
func GetFromContext(c *gin.Context) (*Manager, bool) {
	if manager, exists := c.Get("circuitbreaker"); exists {
		if cbManager, ok := manager.(*Manager); ok {
			return cbManager, true
		}
	}
	return nil, false
}

// handleCircuitBreakerError handles circuit breaker state errors
func handleCircuitBreakerError(c *gin.Context, serviceName string, err error) {
	var message string
	var statusCode int
	
	if err.Error() == "circuit breaker is open" {
		// Circuit breaker is open
		message = "Service temporarily unavailable - circuit breaker is open"
		statusCode = http.StatusServiceUnavailable
	} else {
		message = "Service temporarily unavailable"
		statusCode = http.StatusServiceUnavailable
	}
	
	c.Header("Retry-After", "60") // Suggest retry after 60 seconds
	
	response := CircuitBreakerErrorResponse{
		Error:       message,
		Service:     serviceName,
		StatusCode:  statusCode,
		Timestamp:   time.Now().Unix(),
		RetryAfter:  60,
	}
	
	c.JSON(statusCode, response)
}

// CircuitBreakerErrorResponse represents an error response when circuit breaker is open
type CircuitBreakerErrorResponse struct {
	Error      string `json:"error"`
	Service    string `json:"service"`
	StatusCode int    `json:"status_code"`
	Timestamp  int64  `json:"timestamp"`
	RetryAfter int    `json:"retry_after"`
}

// HTTPError represents an HTTP error
type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return http.StatusText(e.StatusCode)
}

// responseRecorder captures response status codes
type responseRecorder struct {
	gin.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// HealthCheckHandler returns a handler for circuit breaker health checks
func HealthCheckHandler(manager *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := manager.GetHealthStatus()
		
		// Determine overall health
		overallHealthy := true
		for _, status := range health {
			if status.State == gobreaker.StateOpen {
				overallHealthy = false
				break
			}
		}
		
		response := CircuitBreakerHealthResponse{
			Healthy:       overallHealthy,
			Timestamp:     time.Now().Unix(),
			CircuitBreakers: health,
		}
		
		statusCode := http.StatusOK
		if !overallHealthy {
			statusCode = http.StatusServiceUnavailable
		}
		
		c.JSON(statusCode, response)
	}
}

// CircuitBreakerHealthResponse represents the health status response
type CircuitBreakerHealthResponse struct {
	Healthy         bool                             `json:"healthy"`
	Timestamp       int64                            `json:"timestamp"`
	CircuitBreakers map[string]CircuitBreakerHealth  `json:"circuit_breakers"`
}

// StatsHandler returns a handler for circuit breaker statistics
func StatsHandler(manager *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := make(map[string]CircuitBreakerStats)
		
		for serviceName := range manager.breakers {
			if breaker, exists := manager.GetBreaker(serviceName); exists {
				counts := breaker.Counts()
				stats[serviceName] = CircuitBreakerStats{
					Service:              serviceName,
					State:                breaker.State(),
					Requests:             counts.Requests,
					TotalSuccesses:       counts.TotalSuccesses,
					TotalFailures:        counts.TotalFailures,
					ConsecutiveFailures:  counts.ConsecutiveFailures,
					ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
					Timestamp:            time.Now().Unix(),
				}
			}
		}
		
		c.JSON(http.StatusOK, CircuitBreakerStatsResponse{
			Stats:     stats,
			Timestamp: time.Now().Unix(),
		})
	}
}

// CircuitBreakerStats represents statistics for a circuit breaker
type CircuitBreakerStats struct {
	Service              string            `json:"service"`
	State                gobreaker.State   `json:"state"`
	Requests             uint32            `json:"requests"`
	TotalSuccesses       uint32            `json:"total_successes"`
	TotalFailures        uint32            `json:"total_failures"`
	ConsecutiveFailures  uint32            `json:"consecutive_failures"`
	ConsecutiveSuccesses uint32            `json:"consecutive_successes"`
	Timestamp            int64             `json:"timestamp"`
}

// CircuitBreakerStatsResponse represents the stats response
type CircuitBreakerStatsResponse struct {
	Stats     map[string]CircuitBreakerStats `json:"stats"`
	Timestamp int64                          `json:"timestamp"`
}

// ExecuteWithFallback executes a function with circuit breaker and fallback
func ExecuteWithFallback[T any](ctx context.Context, manager *Manager, serviceName string, 
	primaryFn func() (T, error), fallbackFn func() (T, error)) (T, error) {
	
	var zero T
	
	result, err := manager.ExecuteWithBreaker(ctx, serviceName, func() (interface{}, error) {
		return primaryFn()
	})
	
	if err != nil {
		// If circuit breaker is open or there's an error, try fallback
		if fallbackFn != nil {
			return fallbackFn()
		}
		return zero, err
	}
	
	if typedResult, ok := result.(T); ok {
		return typedResult, nil
	}
	
	return zero, nil
}

// AsyncExecuteWithBreaker executes a function asynchronously with circuit breaker
func AsyncExecuteWithBreaker(ctx context.Context, manager *Manager, serviceName string, 
	fn func() (interface{}, error)) <-chan AsyncResult {
	
	resultChan := make(chan AsyncResult, 1)
	
	go func() {
		defer close(resultChan)
		
		result, err := manager.ExecuteWithBreaker(ctx, serviceName, fn)
		resultChan <- AsyncResult{
			Result: result,
			Error:  err,
		}
	}()
	
	return resultChan
}

// AsyncResult represents the result of an async operation
type AsyncResult struct {
	Result interface{}
	Error  error
}
