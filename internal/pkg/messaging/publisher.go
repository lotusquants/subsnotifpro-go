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
)

var (
	ErrNilChannel     = errors.New("rabbitmq channel is nil")
	ErrPublishTimeout = errors.New("publish timed out")
)

// MessagePublisher defines the interface for publishing messages
type MessagePublisher interface {
	Publish(ctx context.Context, queueName string, event interface{}) error
	PublishWithDelay(ctx context.Context, queueName string, event interface{}, delay time.Duration) error
}

type RabbitMQPublisher struct {
	ch *amqp.Channel
	// metrics PublisherMetrics // interface you define
}

type PublisherOptions struct {
	Channel *amqp.Channel
	// Metrics      PublisherMetrics
	DefaultQueue string
}

func NewRabbitMQPublisher(opts PublisherOptions) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		ch: opts.Channel,
		// metrics: opts.Metrics,
	}
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, queueName string, event interface{}) error {
	return p.publish(ctx, queueName, event, 0)
}

func (p *RabbitMQPublisher) PublishWithDelay(ctx context.Context, queueName string, event interface{}, delay time.Duration) error {
	return p.publish(ctx, queueName, event, delay)
}

func (p *RabbitMQPublisher) publish(ctx context.Context, queueName string, event interface{}, delay time.Duration) error {
	if p.ch == nil {
		return ErrNilChannel
	}

	var body []byte
	var headers amqp.Table

	switch v := event.(type) {
	case amqp.Publishing:
		body = v.Body
		headers = v.Headers
		if headers == nil {
			headers = make(amqp.Table)
		}
	case amqp.Delivery:
		body = v.Body
		headers = make(amqp.Table)
		for k, val := range v.Headers {
			headers[k] = val
		}
	default:
		var err error
		body, err = json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal failed: %w", err)
		}
		headers = make(amqp.Table)
	}

	// Debug log the headers before publishing
	log.Printf("Publishing with headers: %+v", headers)

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Headers:      headers,
	}

	if delay > 0 {
		publishing.Headers["x-delay"] = int64(delay / time.Millisecond)
	}

	publishCtx, cancel := context.WithTimeout(ctx, defaultPublishTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		select {
		case <-publishCtx.Done():
			done <- ErrPublishTimeout
		default:
			done <- p.ch.Publish(
				"", // exchange
				queueName,
				false,
				false,
				publishing,
			)
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("publish to queue %s failed: %w", queueName, err)
		}
	case <-publishCtx.Done():
		return ErrPublishTimeout
	}

	return nil
}
