package utils

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"subsnotifpro-go/internal/pkg/logger"
	"time"
)

const (
	initialDelay = 500 * time.Millisecond // Start with 500ms
	maxDelay     = 30 * time.Second       // Max wait time between retries
	maxRetries   = 5                      // Total attempts including the first
)

// RetryWithBackoff executes a function with retry logic using exponential backoff.
// It will retry on failure up to maxRetries, with increasing wait time between attempts.
// Each operation call gets the provided context for handling deadlines/cancellations.
func RetryWithBackoff[T any](ctx context.Context, operation func(ctx context.Context) (T, error), operationName string) (T, error) {
	var lastErr error
	var result T

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Log.Infof("[RETRY] Attempt %d/%d for operation: %s", attempt, maxRetries, operationName)

		select {
		case <-ctx.Done():
			// Context cancelled before operation starts
			logger.Log.Warnf("[RETRY] Operation %s cancelled before attempt %d due to context timeout/cancellation", operationName, attempt)
			return result, fmt.Errorf("%s cancelled before attempt %d: %w", operationName, attempt, ctx.Err())
		default:
			// Proceed with operation
		}

		var err error
		result, err = operation(ctx)
		if err == nil {
			return result, nil // Success
		}

		lastErr = err
		logger.Log.Warnf("[RETRY] Operation %s failed (attempt %d/%d): %v", operationName, attempt, maxRetries, err)

		// Wait before retrying, with exponential backoff
		if attempt < maxRetries {
			waitTime := calculateBackoff(initialDelay, attempt)
			logger.Log.Infof("[RETRY] Waiting %s before next retry (operation: %s)", waitTime, operationName)

			select {
			case <-time.After(waitTime):
				// Wait completed, go to next iteration
			case <-ctx.Done():
				// Context cancelled during backoff
				logger.Log.Warnf("[RETRY] Operation %s cancelled during backoff wait before attempt %d", operationName, attempt+1)
				return result, fmt.Errorf("%s cancelled during backoff: %w", operationName, ctx.Err())
			}
		}
	}

	logger.Log.Errorf("[RETRY] Operation %s failed after %d attempts: %v", operationName, maxRetries, lastErr)
	return result, fmt.Errorf("%s failed after %d attempts: %w", operationName, maxRetries, lastErr)
}

// calculateBackoff calculates the wait time using exponential backoff with jitter.
func calculateBackoff(initialDelay time.Duration, attempt int) time.Duration {
	// Exponential backoff formula: initialDelay * 2^(attempt-1)
	expFactor := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(initialDelay) * expFactor)

	// Add random jitter up to 30% of the delay time to avoid thundering herd.
	jitter := time.Duration(rand.Int63n(int64(delay / 3)))

	// Final delay with jitter
	totalDelay := delay + jitter

	// Cap the delay at maxDelay
	if totalDelay > maxDelay {
		return maxDelay
	}
	return totalDelay
}
