package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Manager manages multiple circuit breakers for different services
type Manager struct {
	breakers map[string]*gobreaker.CircuitBreaker
	configs  *ServiceConfigs
	metrics  *Metrics
	mu       sync.RWMutex
}

// NewManager creates a new circuit breaker manager
func NewManager(configs *ServiceConfigs) *Manager {
	if configs == nil {
		configs = DefaultServiceConfigs()
	}

	manager := &Manager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		configs:  configs,
		metrics:  NewMetrics(),
		mu:       sync.RWMutex{},
	}

	// Initialize circuit breakers for all services
	manager.initializeBreakers()

	return manager
}

// initializeBreakers creates circuit breakers for all configured services
func (m *Manager) initializeBreakers() {
	// Database circuit breaker
	if m.configs.Database != nil {
		m.breakers["database"] = gobreaker.NewCircuitBreaker(
			m.configs.Database.ToBreakerSettings("database"),
		)
	}

	// External API circuit breakers
	if m.configs.PlayStoreAPI != nil {
		m.breakers["playstore_api"] = gobreaker.NewCircuitBreaker(
			m.configs.PlayStoreAPI.ToBreakerSettings("playstore_api"),
		)
	}

	if m.configs.AppStoreAPI != nil {
		m.breakers["appstore_api"] = gobreaker.NewCircuitBreaker(
			m.configs.AppStoreAPI.ToBreakerSettings("appstore_api"),
		)
	}

	if m.configs.RazorpayAPI != nil {
		m.breakers["razorpay_api"] = gobreaker.NewCircuitBreaker(
			m.configs.RazorpayAPI.ToBreakerSettings("razorpay_api"),
		)
	}

	// Infrastructure circuit breakers
	if m.configs.RabbitMQ != nil {
		m.breakers["rabbitmq"] = gobreaker.NewCircuitBreaker(
			m.configs.RabbitMQ.ToBreakerSettings("rabbitmq"),
		)
	}

	if m.configs.RedisCache != nil {
		m.breakers["redis_cache"] = gobreaker.NewCircuitBreaker(
			m.configs.RedisCache.ToBreakerSettings("redis_cache"),
		)
	}

	fmt.Printf("✅ Initialized %d circuit breakers\n", len(m.breakers))
}

// ExecuteWithBreaker executes a function with circuit breaker protection
func (m *Manager) ExecuteWithBreaker(ctx context.Context, serviceName string, fn func() (interface{}, error)) (interface{}, error) {
	m.mu.RLock()
	breaker, exists := m.breakers[serviceName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("circuit breaker not found for service: %s", serviceName)
	}

	// Record attempt
	m.metrics.RecordAttempt(serviceName)
	start := time.Now()

	// Execute with circuit breaker
	result, err := breaker.Execute(func() (interface{}, error) {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return fn()
	})

	// Record metrics
	duration := time.Since(start)
	state := breaker.State()

	if err != nil {
		m.metrics.RecordFailure(serviceName, state, duration)
	} else {
		m.metrics.RecordSuccess(serviceName, state, duration)
	}

	return result, err
}

// ExecuteDatabaseOperation executes database operations with circuit breaker
func (m *Manager) ExecuteDatabaseOperation(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return m.ExecuteWithBreaker(ctx, "database", fn)
}

// ExecutePlayStoreAPI executes PlayStore API calls with circuit breaker
func (m *Manager) ExecutePlayStoreAPI(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return m.ExecuteWithBreaker(ctx, "playstore_api", fn)
}

// ExecuteAppStoreAPI executes AppStore API calls with circuit breaker
func (m *Manager) ExecuteAppStoreAPI(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return m.ExecuteWithBreaker(ctx, "appstore_api", fn)
}

// ExecuteRazorpayAPI executes Razorpay API calls with circuit breaker
func (m *Manager) ExecuteRazorpayAPI(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return m.ExecuteWithBreaker(ctx, "razorpay_api", fn)
}

// ExecuteRabbitMQOperation executes RabbitMQ operations with circuit breaker
func (m *Manager) ExecuteRabbitMQOperation(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return m.ExecuteWithBreaker(ctx, "rabbitmq", fn)
}

// ExecuteCacheOperation executes cache operations with circuit breaker
func (m *Manager) ExecuteCacheOperation(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	return m.ExecuteWithBreaker(ctx, "redis_cache", fn)
}

// GetBreaker returns a specific circuit breaker by name
func (m *Manager) GetBreaker(serviceName string) (*gobreaker.CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	breaker, exists := m.breakers[serviceName]
	return breaker, exists
}

// GetAllStates returns the states of all circuit breakers
func (m *Manager) GetAllStates() map[string]gobreaker.State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]gobreaker.State)
	for name, breaker := range m.breakers {
		states[name] = breaker.State()
	}
	return states
}

// GetHealthStatus returns health status of all circuit breakers
func (m *Manager) GetHealthStatus() map[string]CircuitBreakerHealth {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]CircuitBreakerHealth)
	for name, breaker := range m.breakers {
		counts := breaker.Counts()
		health[name] = CircuitBreakerHealth{
			Name:               name,
			State:              breaker.State(),
			Requests:           counts.Requests,
			TotalSuccesses:     counts.TotalSuccesses,
			TotalFailures:      counts.TotalFailures,
			ConsecutiveFailures: counts.ConsecutiveFailures,
			ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		}
	}
	return health
}

// CircuitBreakerHealth represents the health status of a circuit breaker
type CircuitBreakerHealth struct {
	Name                 string            `json:"name"`
	State                gobreaker.State   `json:"state"`
	Requests             uint32            `json:"requests"`
	TotalSuccesses       uint32            `json:"total_successes"`
	TotalFailures        uint32            `json:"total_failures"`
	ConsecutiveFailures  uint32            `json:"consecutive_failures"`
	ConsecutiveSuccesses uint32            `json:"consecutive_successes"`
}

// AddCustomBreaker adds a custom circuit breaker to the manager
func (m *Manager) AddCustomBreaker(name string, config *CircuitBreakerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	m.breakers[name] = gobreaker.NewCircuitBreaker(
		config.ToBreakerSettings(name),
	)

	fmt.Printf("✅ Added custom circuit breaker: %s\n", name)
}

// RemoveBreaker removes a circuit breaker from the manager
func (m *Manager) RemoveBreaker(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.breakers[name]; exists {
		delete(m.breakers, name)
		fmt.Printf("🗑️ Removed circuit breaker: %s\n", name)
	}
}

// Reset resets all circuit breakers to closed state
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name := range m.breakers {
		// gobreaker doesn't have a public Reset method, so we recreate the breaker
		var config *CircuitBreakerConfig
		switch name {
		case "database":
			config = m.configs.Database
		case "playstore_api":
			config = m.configs.PlayStoreAPI
		case "appstore_api":
			config = m.configs.AppStoreAPI
		case "razorpay_api":
			config = m.configs.RazorpayAPI
		case "rabbitmq":
			config = m.configs.RabbitMQ
		case "redis_cache":
			config = m.configs.RedisCache
		default:
			config = DefaultCircuitBreakerConfig()
		}
		
		// Recreate the breaker to reset its state
		m.breakers[name] = gobreaker.NewCircuitBreaker(
			config.ToBreakerSettings(name),
		)
		
		fmt.Printf("🔄 Reset circuit breaker: %s\n", name)
	}
}

// Metrics holds Prometheus metrics for circuit breakers
type Metrics struct {
	Attempts      *prometheus.CounterVec
	Successes     *prometheus.CounterVec
	Failures      *prometheus.CounterVec
	Duration      *prometheus.HistogramVec
	States        *prometheus.GaugeVec
}

// NewMetrics creates new Prometheus metrics for circuit breakers
func NewMetrics() *Metrics {
	return &Metrics{
		Attempts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "subsnotifpro",
				Subsystem: "circuitbreaker",
				Name:      "attempts_total",
				Help:      "Total number of circuit breaker attempts",
			},
			[]string{"service"},
		),
		Successes: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "subsnotifpro",
				Subsystem: "circuitbreaker",
				Name:      "successes_total",
				Help:      "Total number of successful circuit breaker operations",
			},
			[]string{"service", "state"},
		),
		Failures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "subsnotifpro",
				Subsystem: "circuitbreaker",
				Name:      "failures_total",
				Help:      "Total number of failed circuit breaker operations",
			},
			[]string{"service", "state"},
		),
		Duration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "subsnotifpro",
				Subsystem: "circuitbreaker",
				Name:      "operation_duration_seconds",
				Help:      "Duration of circuit breaker operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "state", "status"},
		),
		States: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "subsnotifpro",
				Subsystem: "circuitbreaker",
				Name:      "state",
				Help:      "Current state of circuit breakers (0=closed, 1=half-open, 2=open)",
			},
			[]string{"service"},
		),
	}
}

// RecordAttempt records a circuit breaker attempt
func (m *Metrics) RecordAttempt(service string) {
	m.Attempts.WithLabelValues(service).Inc()
}

// RecordSuccess records a successful operation
func (m *Metrics) RecordSuccess(service string, state gobreaker.State, duration time.Duration) {
	m.Successes.WithLabelValues(service, state.String()).Inc()
	m.Duration.WithLabelValues(service, state.String(), "success").Observe(duration.Seconds())
	m.updateStateMetric(service, state)
}

// RecordFailure records a failed operation
func (m *Metrics) RecordFailure(service string, state gobreaker.State, duration time.Duration) {
	m.Failures.WithLabelValues(service, state.String()).Inc()
	m.Duration.WithLabelValues(service, state.String(), "failure").Observe(duration.Seconds())
	m.updateStateMetric(service, state)
}

// updateStateMetric updates the state metric
func (m *Metrics) updateStateMetric(service string, state gobreaker.State) {
	var stateValue float64
	switch state {
	case gobreaker.StateClosed:
		stateValue = 0
	case gobreaker.StateHalfOpen:
		stateValue = 1
	case gobreaker.StateOpen:
		stateValue = 2
	}
	m.States.WithLabelValues(service).Set(stateValue)
}
