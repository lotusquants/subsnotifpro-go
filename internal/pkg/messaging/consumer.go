package messaging

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"subsnotifpro-go/internal/pkg/contextutil"
	"time"

	"github.com/streadway/amqp"
)

// Consumer defines a generic AMQP message consumer.
type Consumer struct {
	ch           *amqp.Channel
	exchangeName string
	routingKey   string
	queueName    string
	dlqName      string
	maxRetries   int
	workerCount  int
	handler      func(ctx context.Context, payload []byte) error
	publisher    MessagePublisher
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(
	ch *amqp.Channel,
	exchangeName string,
	routingKey string,
	queueName string,
	dlqName string,
	maxRetries int,
	workerCount int,
	handler func(ctx context.Context, payload []byte) error,
	publisher MessagePublisher,
) *Consumer {
	return &Consumer{
		ch:           ch,
		exchangeName: exchangeName,
		routingKey:   routingKey,
		queueName:    queueName,
		dlqName:      dlqName,
		maxRetries:   maxRetries,
		workerCount:  workerCount,
		handler:      handler,
		publisher:    publisher,
	}
}

// Start begins consuming messages from the queue.
func (c *Consumer) Start(ctx context.Context) {
	msgs, err := c.ch.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Failed to register consumer for queue %s: %v", c.queueName, err)
	}

	log.Printf("🔄 Consumer started for queue: %s", c.queueName)

	// Start worker pool
	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx, msgs)
	}

	<-ctx.Done() // Wait for shutdown
}

// worker processes messages concurrently.
func (c *Consumer) worker(ctx context.Context, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			log.Println("🚦 Shutdown signal received, stopping worker...")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("🚦 Channel closed, stopping worker...")
				return
			}
			c.processMessage(ctx, msg)
		}
	}
}

// processMessage handles a single message (retries, DLQ, etc.).
func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ Panic in message handler: %v", r)
			if err := msg.Nack(false, true); err != nil {
				log.Printf("⚠️ Failed to Nack message after panic: %v", err)
			}
		}
	}()

	// Add the delivery to the context before calling handler
	ctx = contextutil.ContextWithDelivery(ctx, &msg)

	retryCount := GetRetryCount(msg.Headers)

	if retryCount >= c.maxRetries {
		log.Printf("❌ Max retries reached for message, moving to DLQ")
		if err := msg.Nack(false, false); err != nil {
			log.Printf("⚠️ Failed to send message to DLQ: %v", err)
		}
		return
	}

	// Process message
	if err := c.handler(ctx, msg.Body); err != nil {
		log.Printf("❌ Handler failed, requeuing with delay (retry %d/%d)", retryCount+1, c.maxRetries)
		log.Printf("Error: %v", err)
		c.requeueWithDelay(msg, retryCount)
		return
	}

	// Success
	if err := msg.Ack(false); err != nil {
		log.Printf("⚠️ Failed to acknowledge message: %v", err)
	}
}

func (c *Consumer) requeueWithDelay(msg amqp.Delivery, retryCount int) {
	// Calculate delay with jitter
	baseDelay := 1 << retryCount
	jitter := rand.Intn(500)
	delay := time.Duration(baseDelay*1000+jitter) * time.Millisecond

	// Create new headers
	newHeaders := make(amqp.Table)
	for k, v := range msg.Headers {
		newHeaders[k] = v
	}
	newRetryCount := retryCount + 1
	newHeaders["x-retry-count"] = int32(newRetryCount)

	// Publish new message FIRST
	err := c.publisher.PublishWithDelay(
		context.Background(),
		c.exchangeName,
		c.routingKey,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg.Body,
			Headers:     newHeaders,
		},
		delay,
	)

	if err != nil {
		log.Printf("❌ Failed to requeue message: %v", err)
		// Don't ack if we failed to requeue
		if nackErr := msg.Nack(false, true); nackErr != nil {
			log.Printf("⚠️ Failed to Nack message: %v", nackErr)
		}
		return
	}

	// Only ack if requeue succeeded
	if ackErr := msg.Ack(false); ackErr != nil {
		log.Printf("⚠️ Failed to acknowledge original message: %v", ackErr)
		return
	}

	log.Printf("↩️ Requeued message with delay (retry %d/%d)", newRetryCount, c.maxRetries)
}

// getRetryCount extracts retry count from headers.
func GetRetryCount(headers amqp.Table) int {
	if retryHeader, exists := headers["x-retry-count"]; exists {
		switch v := retryHeader.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		default:
			// Fallback for any other type
			if strVal := fmt.Sprintf("%v", v); strVal != "" {
				if i, err := strconv.Atoi(strVal); err == nil {
					return i
				}
			}
		}
	}
	return 0
}
