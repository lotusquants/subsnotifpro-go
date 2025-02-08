// internal/google_playstore/rtdn/queue/consumer.go
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/google_playstore/models"
	"subsnotifpro-go/internal/google_playstore/rtdn/service"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/metrics"

	"github.com/streadway/amqp"
)

// StartQueueConsumer listens for RTDN events and processes them
func StartQueueConsumer(ctx context.Context, ch *amqp.Channel) {
	msgs, err := ch.Consume(constants.RTDNQueue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("❌ Failed to register consumer:", err)
		return
	}

	// ✅ Start message processing in a separate goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("🚨 Panic in Queue Consumer Goroutine:", r)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("🚦 Shutdown signal received! Stopping Queue Consumer...")
				if err := ch.Cancel("", false); err != nil {
					log.Println("⚠️ Warning: Failed to cancel queue consumer:", err)
				}
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("🚦 Queue channel closed, stopping consumer...")
					return
				}
				processMessage(ch, msg)
			}
		}
	}()

	<-ctx.Done() // 🛑 Wait for shutdown signal
}

// requeueWithDelay applies exponential backoff and jitter before requeueing the message
func requeueWithDelay(ch *amqp.Channel, msg amqp.Delivery, retryCount int) {
	// ✅ Check if max retries exceeded
	if retryCount >= constants.MaxRetries {
		log.Printf("❌ Max retries reached for event. Moving to DLQ.")
		if err := msg.Nack(false, false); err != nil {
			log.Printf("⚠️ Warning: Failed to send message to DLQ: %v", err)
		}
		metrics.FailedEvents.WithLabelValues("DLQ").Inc()
		return
	}

	metrics.FailedEvents.WithLabelValues("retry").Inc()

	// ✅ Exponential Backoff Calculation (2^retryCount) + Jitter
	baseDelay := 1 << retryCount
	jitter := rand.Intn(500)
	delay := time.Duration(baseDelay*1000+jitter) * time.Millisecond

	log.Printf("🔄 Retrying event in %v (Retry #%d)", delay, retryCount+1)

	// ✅ Use a goroutine to avoid blocking worker
	go func() {
		time.Sleep(delay)

		// ✅ Copy existing headers and update retry count
		newHeaders := amqp.Table{}
		for key, value := range msg.Headers {
			newHeaders[key] = value
		}
		newHeaders["x-retry-count"] = retryCount + 1

		// ✅ Requeue message with updated retry count
		err := ch.Publish(
			"", constants.RTDNQueue, false, false,
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

// processMessage handles an individual event message
func processMessage(ch *amqp.Channel, msg amqp.Delivery) {
	var event models.GooglePlayWebhookEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logger.Log.Warnf("❌ Failed to decode event: %v", err)
		if err := msg.Nack(false, false); err != nil {
			logger.Log.Errorf("⚠️ Warning: Failed to send message to DLQ: %v", err)
		}
		return
	}

	// ✅ Retrieve retry count safely
	msgRetryCount := 0
	if retryHeader, exists := msg.Headers["x-retry-count"]; exists {
		switch v := retryHeader.(type) {
		case int, int32, int64, float64:
			msgRetryCount = int(v.(int64)) // Convert to int
		case string:
			fmt.Sscanf(v, "%d", &msgRetryCount)
		}
	}

	// ✅ Move to DLQ if max retries reached
	if msgRetryCount >= constants.MaxRetries {
		logger.Log.Warnf("⚠️ Max retries reached. Moving event %s to DLQ", event.ID)
		if err := msg.Nack(false, false); err != nil {
			logger.Log.Errorf("⚠️ Warning: Failed to send message to DLQ: %v", err)
		}
		return
	}

	// ✅ Process event
	logger.Log.Infof("🔄 Processing RTDN event: %s (Retry: %d)", event.ID, msgRetryCount)
	err := service.ProcessWebhookEvent(event)
	if err != nil {
		logger.Log.Errorf("❌ Error processing event %s. Retrying...", event.ID)
		requeueWithDelay(ch, msg, msgRetryCount)
		return
	}

	if err := msg.Ack(false); err != nil {
		logger.Log.Errorf("⚠️ Warning: Failed to acknowledge message %s: %v", event.ID, err)
	}
}
