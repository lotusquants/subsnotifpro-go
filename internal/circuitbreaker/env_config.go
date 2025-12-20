package circuitbreaker

import (
	"os"
	"strconv"
	"time"
	
	"github.com/sony/gobreaker"
)

// LoadConfigFromEnv loads circuit breaker configuration from environment variables
func LoadConfigFromEnv() *ServiceConfigs {
	configs := &ServiceConfigs{}
	
	// Check if circuit breakers are enabled
	if !getBoolEnv("CIRCUIT_BREAKER_ENABLED", true) {
		return nil
	}
	
	// Load database circuit breaker config
	configs.Database = &CircuitBreakerConfig{
		MaxRequests:      uint32(getIntEnv("DB_CB_MAX_REQUESTS", 20)),
		Interval:         getDurationEnv("DB_CB_INTERVAL", 30*time.Second),
		Timeout:          getDurationEnv("DB_CB_TIMEOUT", 60*time.Second),
		FailureThreshold: uint32(getIntEnv("DB_CB_FAILURE_THRESHOLD", 10)),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
	}
	
	// Load PlayStore API circuit breaker config
	configs.PlayStoreAPI = &CircuitBreakerConfig{
		MaxRequests:      uint32(getIntEnv("PLAYSTORE_CB_MAX_REQUESTS", 5)),
		Interval:         getDurationEnv("PLAYSTORE_CB_INTERVAL", 120*time.Second),
		Timeout:          getDurationEnv("PLAYSTORE_CB_TIMEOUT", 15*time.Second),
		FailureThreshold: uint32(getIntEnv("PLAYSTORE_CB_FAILURE_THRESHOLD", 3)),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logStateChange("PlayStore API", name, from, to)
		},
	}
	
	// Load AppStore API circuit breaker config
	configs.AppStoreAPI = &CircuitBreakerConfig{
		MaxRequests:      uint32(getIntEnv("APPSTORE_CB_MAX_REQUESTS", 5)),
		Interval:         getDurationEnv("APPSTORE_CB_INTERVAL", 120*time.Second),
		Timeout:          getDurationEnv("APPSTORE_CB_TIMEOUT", 15*time.Second),
		FailureThreshold: uint32(getIntEnv("APPSTORE_CB_FAILURE_THRESHOLD", 3)),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logStateChange("AppStore API", name, from, to)
		},
	}
	
	// Load Razorpay API circuit breaker config
	configs.RazorpayAPI = &CircuitBreakerConfig{
		MaxRequests:      uint32(getIntEnv("RAZORPAY_CB_MAX_REQUESTS", 5)),
		Interval:         getDurationEnv("RAZORPAY_CB_INTERVAL", 120*time.Second),
		Timeout:          getDurationEnv("RAZORPAY_CB_TIMEOUT", 15*time.Second),
		FailureThreshold: uint32(getIntEnv("RAZORPAY_CB_FAILURE_THRESHOLD", 3)),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logStateChange("Razorpay API", name, from, to)
		},
	}
	
	// Load RabbitMQ circuit breaker config
	configs.RabbitMQ = &CircuitBreakerConfig{
		MaxRequests:      uint32(getIntEnv("RABBITMQ_CB_MAX_REQUESTS", 15)),
		Interval:         getDurationEnv("RABBITMQ_CB_INTERVAL", 45*time.Second),
		Timeout:          getDurationEnv("RABBITMQ_CB_TIMEOUT", 20*time.Second),
		FailureThreshold: uint32(getIntEnv("RABBITMQ_CB_FAILURE_THRESHOLD", 7)),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.5
		},
	}
	
	// Load Redis cache circuit breaker config
	configs.RedisCache = &CircuitBreakerConfig{
		MaxRequests:      uint32(getIntEnv("REDIS_CB_MAX_REQUESTS", 50)),
		Interval:         getDurationEnv("REDIS_CB_INTERVAL", 20*time.Second),
		Timeout:          getDurationEnv("REDIS_CB_TIMEOUT", 5*time.Second),
		FailureThreshold: uint32(getIntEnv("REDIS_CB_FAILURE_THRESHOLD", 15)),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	}
	
	return configs
}

// Helper functions for environment variable parsing

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func logStateChange(serviceName, name string, from, to gobreaker.State) {
	// Use a simple print for now - in production, use proper logging
	println("🔄", serviceName, "circuit breaker", name, "changed from", from.String(), "to", to.String())
}
