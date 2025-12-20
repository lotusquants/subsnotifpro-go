package circuitbreaker

import (
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// DatabaseOperations provides circuit breaker-protected database operations
type DatabaseOperations struct {
	manager *Manager
	db      *gorm.DB
}

// NewDatabaseOperations creates a new database operations wrapper
func NewDatabaseOperations(manager *Manager, db *gorm.DB) *DatabaseOperations {
	return &DatabaseOperations{
		manager: manager,
		db:      db,
	}
}

// QueryWithBreaker executes a database query with circuit breaker protection
func (d *DatabaseOperations) QueryWithBreaker(ctx context.Context, query func(*gorm.DB) error) error {
	_, err := d.manager.ExecuteDatabaseOperation(ctx, func() (interface{}, error) {
		return nil, query(d.db)
	})
	return err
}

// TransactionWithBreaker executes a database transaction with circuit breaker protection
func (d *DatabaseOperations) TransactionWithBreaker(ctx context.Context, txFunc func(*gorm.DB) error) error {
	_, err := d.manager.ExecuteDatabaseOperation(ctx, func() (interface{}, error) {
		return nil, d.db.Transaction(txFunc)
	})
	return err
}

// HTTPClient provides circuit breaker-protected HTTP operations
type HTTPClient struct {
	manager     *Manager
	client      *http.Client
	serviceName string
}

// NewHTTPClient creates a new HTTP client wrapper with circuit breaker
func NewHTTPClient(manager *Manager, serviceName string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		manager:     manager,
		serviceName: serviceName,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// DoWithBreaker executes an HTTP request with circuit breaker protection
func (h *HTTPClient) DoWithBreaker(ctx context.Context, req *http.Request) (*http.Response, error) {
	result, err := h.manager.ExecuteWithBreaker(ctx, h.serviceName, func() (interface{}, error) {
		return h.client.Do(req)
	})
	
	if err != nil {
		return nil, err
	}
	
	if resp, ok := result.(*http.Response); ok {
		return resp, nil
	}
	
	return nil, nil
}

// GetWithBreaker executes a GET request with circuit breaker protection
func (h *HTTPClient) GetWithBreaker(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	return h.DoWithBreaker(ctx, req)
}

// PlayStoreAPIClient provides circuit breaker-protected PlayStore API operations
type PlayStoreAPIClient struct {
	*HTTPClient
}

// NewPlayStoreAPIClient creates a new PlayStore API client with circuit breaker
func NewPlayStoreAPIClient(manager *Manager) *PlayStoreAPIClient {
	return &PlayStoreAPIClient{
		HTTPClient: NewHTTPClient(manager, "playstore_api", 15*time.Second),
	}
}

// ValidateReceipt validates a PlayStore receipt with circuit breaker protection
func (p *PlayStoreAPIClient) ValidateReceipt(ctx context.Context, receipt string) (interface{}, error) {
	return p.manager.ExecutePlayStoreAPI(ctx, func() (interface{}, error) {
		// Implement actual PlayStore API call here
		// This is a placeholder for the actual implementation
		time.Sleep(100 * time.Millisecond) // Simulate API call
		return map[string]interface{}{"valid": true}, nil
	})
}

// AppStoreAPIClient provides circuit breaker-protected AppStore API operations
type AppStoreAPIClient struct {
	*HTTPClient
}

// NewAppStoreAPIClient creates a new AppStore API client with circuit breaker
func NewAppStoreAPIClient(manager *Manager) *AppStoreAPIClient {
	return &AppStoreAPIClient{
		HTTPClient: NewHTTPClient(manager, "appstore_api", 15*time.Second),
	}
}

// ValidateReceipt validates an AppStore receipt with circuit breaker protection
func (a *AppStoreAPIClient) ValidateReceipt(ctx context.Context, receipt string) (interface{}, error) {
	return a.manager.ExecuteAppStoreAPI(ctx, func() (interface{}, error) {
		// Implement actual AppStore API call here
		// This is a placeholder for the actual implementation
		time.Sleep(100 * time.Millisecond) // Simulate API call
		return map[string]interface{}{"valid": true}, nil
	})
}

// RazorpayAPIClient provides circuit breaker-protected Razorpay API operations
type RazorpayAPIClient struct {
	*HTTPClient
}

// NewRazorpayAPIClient creates a new Razorpay API client with circuit breaker
func NewRazorpayAPIClient(manager *Manager) *RazorpayAPIClient {
	return &RazorpayAPIClient{
		HTTPClient: NewHTTPClient(manager, "razorpay_api", 15*time.Second),
	}
}

// CreateSubscription creates a subscription with circuit breaker protection
func (r *RazorpayAPIClient) CreateSubscription(ctx context.Context, subscription interface{}) (interface{}, error) {
	return r.manager.ExecuteRazorpayAPI(ctx, func() (interface{}, error) {
		// Implement actual Razorpay API call here
		// This is a placeholder for the actual implementation
		time.Sleep(100 * time.Millisecond) // Simulate API call
		return map[string]interface{}{"id": "sub_123", "status": "created"}, nil
	})
}

// CacheOperations provides circuit breaker-protected cache operations
type CacheOperations struct {
	manager *Manager
}

// NewCacheOperations creates a new cache operations wrapper
func NewCacheOperations(manager *Manager) *CacheOperations {
	return &CacheOperations{
		manager: manager,
	}
}

// GetWithBreaker gets a value from cache with circuit breaker protection
func (c *CacheOperations) GetWithBreaker(ctx context.Context, key string, getValue func(string) (interface{}, error)) (interface{}, error) {
	return c.manager.ExecuteCacheOperation(ctx, func() (interface{}, error) {
		return getValue(key)
	})
}

// SetWithBreaker sets a value in cache with circuit breaker protection
func (c *CacheOperations) SetWithBreaker(ctx context.Context, key string, value interface{}, setValue func(string, interface{}) error) error {
	_, err := c.manager.ExecuteCacheOperation(ctx, func() (interface{}, error) {
		return nil, setValue(key, value)
	})
	return err
}

// MessageOperations provides circuit breaker-protected messaging operations
type MessageOperations struct {
	manager *Manager
}

// NewMessageOperations creates a new message operations wrapper
func NewMessageOperations(manager *Manager) *MessageOperations {
	return &MessageOperations{
		manager: manager,
	}
}

// PublishWithBreaker publishes a message with circuit breaker protection
func (m *MessageOperations) PublishWithBreaker(ctx context.Context, message interface{}, publish func(interface{}) error) error {
	_, err := m.manager.ExecuteRabbitMQOperation(ctx, func() (interface{}, error) {
		return nil, publish(message)
	})
	return err
}

// ConsumeWithBreaker consumes a message with circuit breaker protection
func (m *MessageOperations) ConsumeWithBreaker(ctx context.Context, consume func() (interface{}, error)) (interface{}, error) {
	return m.manager.ExecuteRabbitMQOperation(ctx, consume)
}

// HealthChecker provides health check functionality for circuit breakers
type HealthChecker struct {
	manager *Manager
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(manager *Manager) *HealthChecker {
	return &HealthChecker{
		manager: manager,
	}
}

// CheckAllServices performs health checks on all services
func (h *HealthChecker) CheckAllServices(ctx context.Context) map[string]ServiceHealth {
	health := make(map[string]ServiceHealth)
	
	// Check database
	health["database"] = h.checkDatabase(ctx)
	
	// Check external APIs
	health["playstore_api"] = h.checkPlayStoreAPI(ctx)
	health["appstore_api"] = h.checkAppStoreAPI(ctx)
	health["razorpay_api"] = h.checkRazorpayAPI(ctx)
	
	// Check infrastructure
	health["rabbitmq"] = h.checkRabbitMQ(ctx)
	health["redis_cache"] = h.checkRedisCache(ctx)
	
	return health
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Service    string                     `json:"service"`
	Healthy    bool                       `json:"healthy"`
	Status     string                     `json:"status"`
	LastCheck  time.Time                  `json:"last_check"`
	Circuit    CircuitBreakerHealth       `json:"circuit_breaker"`
}

func (h *HealthChecker) checkDatabase(ctx context.Context) ServiceHealth {
	start := time.Now()
	healthy := true
	status := "healthy"
	
	_, err := h.manager.ExecuteDatabaseOperation(ctx, func() (interface{}, error) {
		// Simple ping operation
		time.Sleep(10 * time.Millisecond)
		return nil, nil
	})
	
	if err != nil {
		healthy = false
		status = err.Error()
	}
	
	breaker, _ := h.manager.GetBreaker("database")
	var cbHealth CircuitBreakerHealth
	if breaker != nil {
		counts := breaker.Counts()
		cbHealth = CircuitBreakerHealth{
			Name:                 "database",
			State:                breaker.State(),
			Requests:             counts.Requests,
			TotalSuccesses:       counts.TotalSuccesses,
			TotalFailures:        counts.TotalFailures,
			ConsecutiveFailures:  counts.ConsecutiveFailures,
			ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		}
	}
	
	return ServiceHealth{
		Service:   "database",
		Healthy:   healthy,
		Status:    status,
		LastCheck: start,
		Circuit:   cbHealth,
	}
}

func (h *HealthChecker) checkPlayStoreAPI(ctx context.Context) ServiceHealth {
	return h.checkExternalAPI(ctx, "playstore_api")
}

func (h *HealthChecker) checkAppStoreAPI(ctx context.Context) ServiceHealth {
	return h.checkExternalAPI(ctx, "appstore_api")
}

func (h *HealthChecker) checkRazorpayAPI(ctx context.Context) ServiceHealth {
	return h.checkExternalAPI(ctx, "razorpay_api")
}

func (h *HealthChecker) checkRabbitMQ(ctx context.Context) ServiceHealth {
	return h.checkInfrastructure(ctx, "rabbitmq")
}

func (h *HealthChecker) checkRedisCache(ctx context.Context) ServiceHealth {
	return h.checkInfrastructure(ctx, "redis_cache")
}

func (h *HealthChecker) checkExternalAPI(ctx context.Context, serviceName string) ServiceHealth {
	start := time.Now()
	healthy := true
	status := "healthy"
	
	_, err := h.manager.ExecuteWithBreaker(ctx, serviceName, func() (interface{}, error) {
		// Simulate API health check
		time.Sleep(50 * time.Millisecond)
		return nil, nil
	})
	
	if err != nil {
		healthy = false
		status = err.Error()
	}
	
	breaker, _ := h.manager.GetBreaker(serviceName)
	var cbHealth CircuitBreakerHealth
	if breaker != nil {
		counts := breaker.Counts()
		cbHealth = CircuitBreakerHealth{
			Name:                 serviceName,
			State:                breaker.State(),
			Requests:             counts.Requests,
			TotalSuccesses:       counts.TotalSuccesses,
			TotalFailures:        counts.TotalFailures,
			ConsecutiveFailures:  counts.ConsecutiveFailures,
			ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		}
	}
	
	return ServiceHealth{
		Service:   serviceName,
		Healthy:   healthy,
		Status:    status,
		LastCheck: start,
		Circuit:   cbHealth,
	}
}

func (h *HealthChecker) checkInfrastructure(ctx context.Context, serviceName string) ServiceHealth {
	start := time.Now()
	healthy := true
	status := "healthy"
	
	_, err := h.manager.ExecuteWithBreaker(ctx, serviceName, func() (interface{}, error) {
		// Simulate infrastructure health check
		time.Sleep(20 * time.Millisecond)
		return nil, nil
	})
	
	if err != nil {
		healthy = false
		status = err.Error()
	}
	
	breaker, _ := h.manager.GetBreaker(serviceName)
	var cbHealth CircuitBreakerHealth
	if breaker != nil {
		counts := breaker.Counts()
		cbHealth = CircuitBreakerHealth{
			Name:                 serviceName,
			State:                breaker.State(),
			Requests:             counts.Requests,
			TotalSuccesses:       counts.TotalSuccesses,
			TotalFailures:        counts.TotalFailures,
			ConsecutiveFailures:  counts.ConsecutiveFailures,
			ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		}
	}
	
	return ServiceHealth{
		Service:   serviceName,
		Healthy:   healthy,
		Status:    status,
		LastCheck: start,
		Circuit:   cbHealth,
	}
}
