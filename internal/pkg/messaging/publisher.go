package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

const (
	defaultPublishTimeout = 5 * time.Second
	retryHeader           = "x-retry-count"
	delayHeader           = "x-delay"
)

var (
	ErrNilChannel     = errors.New("rabbitmq channel is nil")
	ErrPublishTimeout = errors.New("publish timed out")
)

// MessagePublisher defines the interface for publishing messages
type MessagePublisher interface {
	PublishToQueue(ctx context.Context, queueName string, event interface{}) error
	PublishToExchange(ctx context.Context, exchange, routingKey string, event interface{}) error
	PublishWithDelay(ctx context.Context, exchange, routingKey string, event interface{}, delay time.Duration) error
}

type RabbitMQPublisher struct {
	ch *amqp.Channel
	// metrics PublisherMetrics // Optional metrics interface
}

type PublisherOptions struct {
	Channel *amqp.Channel
	// Metrics PublisherMetrics
}

// type PublisherMetrics interface {
// 	IncPublishSuccess(exchange, routingKey string)
// 	IncPublishFailure(exchange, routingKey string)
// 	ObservePublishLatency(exchange string, duration time.Duration)
// }

func NewRabbitMQPublisher(opts PublisherOptions) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		ch: opts.Channel,
		// metrics: opts.Metrics,
	}
}

// PublishToQueue publishes directly to a queue (legacy support)
func (p *RabbitMQPublisher) PublishToQueue(ctx context.Context, queueName string, event interface{}) error {
	return p.publish(ctx, "", queueName, event, 0)
}

// PublishToExchange publishes to an exchange with routing key
func (p *RabbitMQPublisher) PublishToExchange(
	ctx context.Context,
	exchange string,
	routingKey string,
	event interface{},
) error {
	return p.publish(ctx, exchange, routingKey, event, 0)
}

// PublishWithDelay publishes with a delay using delayed exchange plugin
func (p *RabbitMQPublisher) PublishWithDelay(
	ctx context.Context,
	exchange string,
	routingKey string,
	event interface{},
	delay time.Duration,
) error {
	return p.publish(ctx, exchange, routingKey, event, delay)
}

// publish handles the core publishing logic
func (p *RabbitMQPublisher) publish(
	ctx context.Context,
	exchange string,
	routingKey string,
	event interface{},
	delay time.Duration,
) error {
	if p.ch == nil {
		return ErrNilChannel
	}

	// startTime := time.Now()
	var body []byte
	var headers amqp.Table

	// Handle different input types
	switch v := event.(type) {
	case amqp.Publishing:
		body = v.Body
		headers = v.Headers
	case amqp.Delivery:
		body = v.Body
		headers = v.Headers
	default:
		var err error
		body, err = json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal failed: %w", err)
		}
	}

	if headers == nil {
		headers = make(amqp.Table)
	}

	// Prepare publishing
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Headers:      headers,
	}

	// Handle delayed messages
	if delay > 0 {
		publishing.Headers[delayHeader] = int64(delay / time.Millisecond)
	}

	// Set up context with timeout
	publishCtx, cancel := context.WithTimeout(ctx, defaultPublishTimeout)
	defer cancel()

	// Execute publish in goroutine to handle timeouts
	done := make(chan error, 1)
	go func() {
		err := p.ch.Publish(
			exchange,
			routingKey,
			false, // mandatory
			false, // immediate
			publishing,
		)
		select {
		case done <- err:
		case <-publishCtx.Done():
			return
		}
	}()

	// Wait for completion or timeout
	var err error
	select {
	case err = <-done:
	case <-publishCtx.Done():
		err = ErrPublishTimeout
	}

	// // Record metrics if available
	// if p.metrics != nil {
	// 	if err == nil {
	// 		p.metrics.IncPublishSuccess(exchange, routingKey)
	// 		p.metrics.ObservePublishLatency(exchange, time.Since(startTime))
	// 	} else {
	// 		p.metrics.IncPublishFailure(exchange, routingKey)
	// 	}
	// }

	if err != nil {
		return fmt.Errorf("publish to exchange %s with key %s failed: %w", exchange, routingKey, err)
	}

	log.Printf("Published to exchange %s with routing key %s", exchange, routingKey)
	return nil
}
