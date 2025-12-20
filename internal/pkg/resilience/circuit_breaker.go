package resilience

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker wraps gobreaker with additional functionality
type CircuitBreaker struct {
	cb   *gobreaker.CircuitBreaker
	name string
}

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	Name             string
	MaxRequests      uint32
	Interval         time.Duration
	Timeout          time.Duration
	FailureThreshold uint32
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:             name,
		MaxRequests:      3,
		Interval:         10 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= config.FailureThreshold
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes for monitoring
			fmt.Printf("Circuit breaker '%s' changed from %s to %s\n", name, from, to)
		},
	}

	return &CircuitBreaker{
		cb:   gobreaker.NewCircuitBreaker(settings),
		name: config.Name,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, req func() (interface{}, error)) (interface{}, error) {
	return cb.cb.Execute(req)
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() gobreaker.State {
	return cb.cb.State()
}

// Name returns the circuit breaker name
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns sensible defaults for retry logic
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}
}

// WithRetry executes a function with exponential backoff retry logic
func WithRetry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Don't delay after the last attempt
		if attempt < config.MaxAttempts-1 {
			delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt)))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// WithRetryAndResult executes a function with retry logic and returns a result
func WithRetryAndResult(ctx context.Context, config RetryConfig, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if result, err := fn(); err == nil {
			return result, nil
		} else {
			lastErr = err
		}

		// Don't delay after the last attempt
		if attempt < config.MaxAttempts-1 {
			delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt)))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// ResilienceWrapper combines circuit breaker and retry logic
type ResilienceWrapper struct {
	circuitBreaker *CircuitBreaker
	retryConfig    RetryConfig
}

// NewResilienceWrapper creates a new wrapper with both circuit breaker and retry
func NewResilienceWrapper(cbConfig CircuitBreakerConfig, retryConfig RetryConfig) *ResilienceWrapper {
	return &ResilienceWrapper{
		circuitBreaker: NewCircuitBreaker(cbConfig),
		retryConfig:    retryConfig,
	}
}

// Execute runs a function with both circuit breaker and retry protection
func (rw *ResilienceWrapper) Execute(ctx context.Context, fn func() error) error {
	return WithRetry(ctx, rw.retryConfig, func() error {
		_, err := rw.circuitBreaker.Execute(ctx, func() (interface{}, error) {
			return nil, fn()
		})
		return err
	})
}

// ExecuteWithResult runs a function with both circuit breaker and retry protection, returning a result
func (rw *ResilienceWrapper) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return WithRetryAndResult(ctx, rw.retryConfig, func() (interface{}, error) {
		result, err := rw.circuitBreaker.Execute(ctx, func() (interface{}, error) {
			return fn()
		})
		if err != nil {
			return nil, err
		}
		return result, nil
	})
}

// GetCircuitBreakerState returns the current state of the circuit breaker
func (rw *ResilienceWrapper) GetCircuitBreakerState() gobreaker.State {
	return rw.circuitBreaker.State()
}

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Add specific error checking logic here
	// For example, HTTP 5xx errors, timeout errors, etc.
	switch err {
	case context.DeadlineExceeded, context.Canceled:
		return false // Don't retry context cancellation
	default:
		return true // Retry other errors by default
	}
}

// ConditionalRetry allows custom retry logic based on error type
func ConditionalRetry(ctx context.Context, config RetryConfig, shouldRetry func(error) bool, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if !shouldRetry(err) {
				return err // Don't retry this error
			}
		}

		// Don't delay after the last attempt
		if attempt < config.MaxAttempts-1 {
			delay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt)))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}
