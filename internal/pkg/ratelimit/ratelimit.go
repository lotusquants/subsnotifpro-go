package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	limiter *rate.Limiter
	name    string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(name string, requestsPerSecond int, burstSize int) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize),
		name:    name,
	}
}

// Wait blocks until the rate limiter allows the operation
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// Allow reports whether the operation may happen now
func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

// Reserve returns a Reservation that indicates how long the caller
// must wait before the operation is allowed to happen
func (rl *RateLimiter) Reserve() *rate.Reservation {
	return rl.limiter.Reserve()
}

// SetLimit changes the rate limit and burst size
func (rl *RateLimiter) SetLimit(newLimit rate.Limit, newBurst int) {
	rl.limiter.SetLimit(newLimit)
	rl.limiter.SetBurst(newBurst)
}

// Name returns the name of the rate limiter
func (rl *RateLimiter) Name() string {
	return rl.name
}

// MultiLimiter manages multiple rate limiters for different services
type MultiLimiter struct {
	limiters map[string]*RateLimiter
	mutex    sync.RWMutex
}

// NewMultiLimiter creates a new multi-service rate limiter
func NewMultiLimiter() *MultiLimiter {
	return &MultiLimiter{
		limiters: make(map[string]*RateLimiter),
	}
}

// AddLimiter adds a rate limiter for a specific service
func (ml *MultiLimiter) AddLimiter(serviceName string, requestsPerSecond int, burstSize int) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	ml.limiters[serviceName] = NewRateLimiter(serviceName, requestsPerSecond, burstSize)
}

// Wait waits for the specified service's rate limiter
func (ml *MultiLimiter) Wait(ctx context.Context, serviceName string) error {
	ml.mutex.RLock()
	limiter, exists := ml.limiters[serviceName]
	ml.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("rate limiter not found for service: %s", serviceName)
	}

	return limiter.Wait(ctx)
}

// Allow checks if the operation is allowed for the specified service
func (ml *MultiLimiter) Allow(serviceName string) bool {
	ml.mutex.RLock()
	limiter, exists := ml.limiters[serviceName]
	ml.mutex.RUnlock()

	if !exists {
		return false // Deny if no limiter configured
	}

	return limiter.Allow()
}

// GetLimiter returns the rate limiter for a specific service
func (ml *MultiLimiter) GetLimiter(serviceName string) *RateLimiter {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	return ml.limiters[serviceName]
}

// RemoveLimiter removes a rate limiter for a specific service
func (ml *MultiLimiter) RemoveLimiter(serviceName string) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	delete(ml.limiters, serviceName)
}

// ListServices returns all configured service names
func (ml *MultiLimiter) ListServices() []string {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	services := make([]string, 0, len(ml.limiters))
	for service := range ml.limiters {
		services = append(services, service)
	}

	return services
}

// RateLimitConfig holds configuration for different rate limits
type RateLimitConfig struct {
	GooglePlayAPI struct {
		RequestsPerSecond int
		BurstSize         int
	}
	AppStoreAPI struct {
		RequestsPerSecond int
		BurstSize         int
	}
	DatabaseAPI struct {
		RequestsPerSecond int
		BurstSize         int
	}
}

// DefaultRateLimitConfig returns sensible defaults for different services
func DefaultRateLimitConfig() RateLimitConfig {
	config := RateLimitConfig{}

	// Google Play API - Conservative limits to avoid hitting quotas
	config.GooglePlayAPI.RequestsPerSecond = 10
	config.GooglePlayAPI.BurstSize = 20

	// App Store API - Conservative limits
	config.AppStoreAPI.RequestsPerSecond = 5
	config.AppStoreAPI.BurstSize = 10

	// Database API - Higher limits for internal operations
	config.DatabaseAPI.RequestsPerSecond = 100
	config.DatabaseAPI.BurstSize = 200

	return config
}

// SetupGlobalRateLimiters creates and configures rate limiters for all services
func SetupGlobalRateLimiters(config RateLimitConfig) *MultiLimiter {
	ml := NewMultiLimiter()

	// Add rate limiters for each service
	ml.AddLimiter("google_play_api", config.GooglePlayAPI.RequestsPerSecond, config.GooglePlayAPI.BurstSize)
	ml.AddLimiter("app_store_api", config.AppStoreAPI.RequestsPerSecond, config.AppStoreAPI.BurstSize)
	ml.AddLimiter("database_api", config.DatabaseAPI.RequestsPerSecond, config.DatabaseAPI.BurstSize)

	return ml
}

// AdaptiveRateLimiter automatically adjusts limits based on error rates
type AdaptiveRateLimiter struct {
	baseLimiter      *RateLimiter
	currentLimit     rate.Limit
	baseLimit        rate.Limit
	errorRate        float64
	successCount     int64
	errorCount       int64
	lastAdjustment   time.Time
	adjustmentWindow time.Duration
	mutex            sync.RWMutex
}

// NewAdaptiveRateLimiter creates a rate limiter that adjusts based on success/error rates
func NewAdaptiveRateLimiter(name string, baseRequestsPerSecond int, burstSize int) *AdaptiveRateLimiter {
	baseLimit := rate.Limit(baseRequestsPerSecond)
	return &AdaptiveRateLimiter{
		baseLimiter:      NewRateLimiter(name, baseRequestsPerSecond, burstSize),
		currentLimit:     baseLimit,
		baseLimit:        baseLimit,
		lastAdjustment:   time.Now(),
		adjustmentWindow: 1 * time.Minute,
	}
}

// Wait waits for the adaptive rate limiter to allow the operation
func (arl *AdaptiveRateLimiter) Wait(ctx context.Context) error {
	return arl.baseLimiter.Wait(ctx)
}

// Allow checks if the operation is allowed
func (arl *AdaptiveRateLimiter) Allow() bool {
	return arl.baseLimiter.Allow()
}

// RecordSuccess records a successful operation
func (arl *AdaptiveRateLimiter) RecordSuccess() {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()

	arl.successCount++
	arl.adjustIfNeeded()
}

// RecordError records a failed operation
func (arl *AdaptiveRateLimiter) RecordError() {
	arl.mutex.Lock()
	defer arl.mutex.Unlock()

	arl.errorCount++
	arl.adjustIfNeeded()
}

// adjustIfNeeded adjusts the rate limit based on error rates
func (arl *AdaptiveRateLimiter) adjustIfNeeded() {
	now := time.Now()
	if now.Sub(arl.lastAdjustment) < arl.adjustmentWindow {
		return
	}

	totalRequests := arl.successCount + arl.errorCount
	if totalRequests == 0 {
		return
	}

	arl.errorRate = float64(arl.errorCount) / float64(totalRequests)

	// Adjust limits based on error rate
	var newLimit rate.Limit
	if arl.errorRate > 0.1 { // High error rate - reduce limit
		newLimit = arl.baseLimit * 0.5
	} else if arl.errorRate < 0.01 { // Low error rate - increase limit
		newLimit = arl.baseLimit * 1.5
	} else {
		newLimit = arl.baseLimit // Normal error rate - keep base limit
	}

	arl.baseLimiter.SetLimit(newLimit, int(newLimit*2)) // Burst is 2x the limit
	arl.currentLimit = newLimit
	arl.lastAdjustment = now

	// Reset counters
	arl.successCount = 0
	arl.errorCount = 0
}

// GetCurrentLimit returns the current rate limit
func (arl *AdaptiveRateLimiter) GetCurrentLimit() rate.Limit {
	arl.mutex.RLock()
	defer arl.mutex.RUnlock()
	return arl.currentLimit
}

// GetErrorRate returns the current error rate
func (arl *AdaptiveRateLimiter) GetErrorRate() float64 {
	arl.mutex.RLock()
	defer arl.mutex.RUnlock()
	return arl.errorRate
}
