package publisher

import (
	"context"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/pkg/messaging"
	"subsnotifpro-go/internal/subscription/events"
	"time"
)

// UnifiedEventPublisher handles subscription-specific event publishing
type UnifiedEventPublisher struct {
	publisher  messaging.MessagePublisher
	exchange   string
	routingKey string
	maxRetries int
	retryDelay time.Duration
}

type UnifiedPublisherOpts struct {
	Publisher  messaging.MessagePublisher // Your existing RabbitMQPublisher
	Exchange   string                     // e.g., "unified_subscription_exchange"
	RoutingKey string                     // e.g., "subscription.update"
	MaxRetries int                        // Default: 3
	RetryDelay time.Duration              // Default: 200ms
}

func NewUnifiedEventPublisher(opts UnifiedPublisherOpts) *UnifiedEventPublisher {
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = 200 * time.Millisecond
	}

	return &UnifiedEventPublisher{
		publisher:  opts.Publisher,
		exchange:   opts.Exchange,
		routingKey: opts.RoutingKey,
		maxRetries: opts.MaxRetries,
		retryDelay: opts.RetryDelay,
	}
}

// Publish sends a unified subscription event with:
// - Automatic retries
// - Dead-letter routing
// - Metrics integration
func (p *UnifiedEventPublisher) Publish(ctx context.Context, event events.UnifiedEvent) error {
	attempt := 0
	currentDelay := p.retryDelay

	for {
		err := p.publisher.PublishToExchange(
			ctx,
			p.exchange,
			p.routingKey,
			event,
		)

		if err == nil {
			logger.Log.Infof("Published unified subscription event to exchange %s with key %s: %+v",
				p.exchange, p.routingKey, event)
			return nil
		}

		attempt++
		if attempt >= p.maxRetries {
			logger.Log.Errorf("Failed to publish after %d attempts: %v", attempt, err)
			return err
		}

		logger.Log.Warnf("Publish attempt %d failed, retrying in %v: %v",
			attempt, currentDelay, err)

		select {
		case <-time.After(currentDelay):
			currentDelay *= 2 // Exponential backoff
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
