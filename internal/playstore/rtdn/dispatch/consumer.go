package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/metrics"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/rtdn/repository"
	"subsnotifpro-go/internal/playstore/rtdn/service"

	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

// Define worker count
const workerCount = 5

// Consumer struct to hold dependencies
type Consumer struct {
	ch        *amqp.Channel
	repo      repository.RTDNRepository
	service   service.RTDNService
	queueName string
}

// NewConsumer initializes the queue consumer with dependencies
func NewConsumer(ch *amqp.Channel, repo repository.RTDNRepository, svc service.RTDNService) *Consumer {
	return &Consumer{
		ch:        ch,
		repo:      repo,
		service:   svc,
		queueName: constants.RTDNQueue,
	}
}

// Start listens for RTDN events and processes them
func (c *Consumer) Start(ctx context.Context) {
	msgs, err := c.ch.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("❌ Failed to register consumer:", err)
		return
	}
	logger.Log.Infof("🔄 Queue consumer started successfully.")

	// Start worker pool for concurrent processing
	for i := 0; i < workerCount; i++ {
		go c.worker(ctx, msgs)
	}

	<-ctx.Done() // 🛑 Wait for shutdown signal
}

// Worker function to process messages
func (c *Consumer) worker(ctx context.Context, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			log.Println("🚦 Shutdown signal received! Stopping worker...")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("🚦 Queue channel closed, stopping worker...")
				return
			}
			c.processMessage(ctx, msg)
		}
	}
}

// processMessage handles an individual event message
func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	logger.Log.Infof("🔄 Received message: %s", string(msg.Body))

	var event models.GooglePlayWebhookEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.Log.Warnf("❌ Failed to decode event: %v", err)
		_ = msg.Nack(false, false)
		return
	}

	logger.Log.Infof("✅ Successfully unmarshalled event: %s", event.ID)

	msgRetryCount := getRetryCount(msg.Headers)

	// ✅ Handle transaction
	err := c.repo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// Check if retries exceeded
		if msgRetryCount >= constants.MaxRetries {
			logger.Log.Warnf("⚠️ Max retries reached for event %s. Moving to DLQ", event.ID)
			if err := c.repo.UpdateWebhookStatus(ctx, tx, event.ID, "movedToDLQ"); err != nil {
				return err
			}
			return msg.Nack(false, false)
		}

		// ✅ Process event
		if err := c.service.ProcessWebhookEvent(event); err != nil {
			logger.Log.Errorf("❌ Error processing event %s. Retrying...", event.ID)
			c.requeueWithDelay(msg, msgRetryCount)
			return err
		}

		// ✅ Mark as completed
		if err := c.repo.UpdateWebhookStatus(ctx, tx, event.ID, "completed"); err != nil {
			logger.Log.Errorf("⚠️ Failed to update status for event %s: %v", event.ID, err)
			return err
		}

		return nil
	})

	if err != nil {
		logger.Log.Errorf("❌ Final error processing event %s: %v", event.ID, err)
		return
	}

	// ✅ Acknowledge message after transaction success
	if err := msg.Ack(false); err != nil {
		logger.Log.Errorf("⚠️ Failed to acknowledge message %s: %v", event.ID, err)
	} else {
		logger.Log.Infof("✅ Message %s acknowledged successfully", event.ID)
	}
}

// requeueWithDelay applies exponential backoff and jitter before requeueing the message
func (c *Consumer) requeueWithDelay(msg amqp.Delivery, retryCount int) {
	if retryCount >= constants.MaxRetries {
		log.Printf("❌ Max retries reached for event. Moving to DLQ.")
		if err := msg.Nack(false, false); err != nil {
			log.Printf("⚠️ Warning: Failed to send message to DLQ: %v", err)
		}
		metrics.FailedEvents.WithLabelValues("DLQ").Inc()
		return
	}

	metrics.FailedEvents.WithLabelValues("retry").Inc()

	// Exponential Backoff Calculation (2^retryCount) + Jitter
	baseDelay := 1 << retryCount
	jitter := rand.Intn(500)
	delay := time.Duration(baseDelay*1000+jitter) * time.Millisecond

	log.Printf("🔄 Retrying event in %v (Retry #%d)", delay, retryCount+1)

	go func() {
		time.Sleep(delay)

		newHeaders := amqp.Table{}
		for key, value := range msg.Headers {
			newHeaders[key] = value
		}
		newHeaders["x-retry-count"] = retryCount + 1

		err := c.ch.Publish(
			"", c.queueName, false, false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        msg.Body,
				Headers:     newHeaders,
			},
		)
		if err != nil {
			log.Printf("❌ Failed to requeue event: %v", err)
		}
	}()
}

// Helper function to extract retry count from headers
func getRetryCount(headers amqp.Table) int {
	msgRetryCount := 0
	if retryHeader, exists := headers["x-retry-count"]; exists {
		switch v := retryHeader.(type) {
		case int:
			msgRetryCount = v
		case int32:
			msgRetryCount = int(v)
		case int64:
			msgRetryCount = int(v)
		case string:
			fmt.Sscanf(v, "%d", &msgRetryCount)
		}
	}
	return msgRetryCount
}
