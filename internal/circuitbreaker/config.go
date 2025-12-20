package circuitbreaker

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreakerConfig holds configuration for circuit breakers
type CircuitBreakerConfig struct {
	// Basic settings
	MaxRequests        uint32        `mapstructure:"max_requests" json:"max_requests"`
	Interval           time.Duration `mapstructure:"interval" json:"interval"`
	Timeout            time.Duration `mapstructure:"timeout" json:"timeout"`
	
	// Failure threshold settings
	FailureThreshold   uint32        `mapstructure:"failure_threshold" json:"failure_threshold"`
	SuccessThreshold   uint32        `mapstructure:"success_threshold" json:"success_threshold"`
	
	// Custom settings per service
	ReadyToTrip        func(counts gobreaker.Counts) bool `json:"-"`
	OnStateChange      func(name string, from gobreaker.State, to gobreaker.State) `json:"-"`
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxRequests:      10,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 3,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("🔄 Circuit breaker '%s' changed from %s to %s\n", name, from, to)
		},
	}
}

// ServiceConfigs holds circuit breaker configurations for different services
type ServiceConfigs struct {
	Database     *CircuitBreakerConfig `mapstructure:"database" json:"database"`
	PlayStoreAPI *CircuitBreakerConfig `mapstructure:"playstore_api" json:"playstore_api"`
	AppStoreAPI  *CircuitBreakerConfig `mapstructure:"appstore_api" json:"appstore_api"`
	RazorpayAPI  *CircuitBreakerConfig `mapstructure:"razorpay_api" json:"razorpay_api"`
	RabbitMQ     *CircuitBreakerConfig `mapstructure:"rabbitmq" json:"rabbitmq"`
	RedisCache   *CircuitBreakerConfig `mapstructure:"redis_cache" json:"redis_cache"`
}

// DefaultServiceConfigs returns default configurations for all services
func DefaultServiceConfigs() *ServiceConfigs {
	return &ServiceConfigs{
		Database:     DatabaseCircuitBreakerConfig(),
		PlayStoreAPI: ExternalAPICircuitBreakerConfig("PlayStore"),
		AppStoreAPI:  ExternalAPICircuitBreakerConfig("AppStore"),
		RazorpayAPI:  ExternalAPICircuitBreakerConfig("Razorpay"),
		RabbitMQ:     MessagingCircuitBreakerConfig(),
		RedisCache:   CacheCircuitBreakerConfig(),
	}
}

// DatabaseCircuitBreakerConfig returns optimized config for database operations
func DatabaseCircuitBreakerConfig() *CircuitBreakerConfig {
	config := DefaultCircuitBreakerConfig()
	config.MaxRequests = 20        // Higher for database
	config.Interval = 30 * time.Second
	config.Timeout = 60 * time.Second  // Longer timeout for DB
	config.FailureThreshold = 10
	config.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 10 && failureRatio >= 0.6
	}
	return config
}

// ExternalAPICircuitBreakerConfig returns config for external API calls
func ExternalAPICircuitBreakerConfig(serviceName string) *CircuitBreakerConfig {
	config := DefaultCircuitBreakerConfig()
	config.MaxRequests = 5         // Conservative for external APIs
	config.Interval = 120 * time.Second  // Longer recovery time
	config.Timeout = 15 * time.Second    // Shorter timeout for APIs
	config.FailureThreshold = 3
	config.ReadyToTrip = func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures >= 3
	}
	config.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		fmt.Printf("🌐 %s API circuit breaker '%s' changed from %s to %s\n", serviceName, name, from, to)
	}
	return config
}

// MessagingCircuitBreakerConfig returns config for messaging systems
func MessagingCircuitBreakerConfig() *CircuitBreakerConfig {
	config := DefaultCircuitBreakerConfig()
	config.MaxRequests = 15
	config.Interval = 45 * time.Second
	config.Timeout = 20 * time.Second
	config.FailureThreshold = 7
	config.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 5 && failureRatio >= 0.5
	}
	return config
}

// CacheCircuitBreakerConfig returns config for cache operations
func CacheCircuitBreakerConfig() *CircuitBreakerConfig {
	config := DefaultCircuitBreakerConfig()
	config.MaxRequests = 50        // High for cache
	config.Interval = 20 * time.Second  // Quick recovery
	config.Timeout = 5 * time.Second    // Very fast timeout
	config.FailureThreshold = 15
	config.ReadyToTrip = func(counts gobreaker.Counts) bool {
		// Cache failures should trip quickly but recover fast
		return counts.ConsecutiveFailures >= 5
	}
	return config
}

// ToBreakerSettings converts CircuitBreakerConfig to gobreaker.Settings
func (c *CircuitBreakerConfig) ToBreakerSettings(name string) gobreaker.Settings {
	return gobreaker.Settings{
		Name:         name,
		MaxRequests:  c.MaxRequests,
		Interval:     c.Interval,
		Timeout:      c.Timeout,
		ReadyToTrip:  c.ReadyToTrip,
		OnStateChange: c.OnStateChange,
	}
}
