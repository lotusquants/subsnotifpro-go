// internal/google_playstore/rtdn/queue/dlq_consumer.go

package queue

import (
	"context"
	"log"
	"time"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/metrics"

	"github.com/streadway/amqp"
)

// StartDLQConsumer listens to the Dead Letter Queue (DLQ) and processes messages
func StartDLQConsumer(ctx context.Context, ch *amqp.Channel) {
	msgs, err := ch.Consume(constants.RTDNDLQ, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("❌ Failed to register DLQ consumer:", err)
		return
	}

	// ✅ Start DLQ monitoring in a separate goroutine
	go monitorDLQSize(ctx, ch)

	// ✅ Start message processing in a separate goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("🚨 Panic in DLQ Consumer Goroutine:", r)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("🚦 Shutdown signal received! Stopping DLQ Consumer...")
				if err := ch.Cancel("", false); err != nil {
					log.Println("⚠️ Warning: Failed to cancel DLQ consumer:", err)
				}
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("🚦 DLQ channel closed, stopping consumer...")
					return
				}
				processDLQMessage(msg)
			}
		}
	}()

	<-ctx.Done() // 🛑 Wait for shutdown signal
	log.Println("🚦 DLQ Consumer stopped.")
}

// processDLQMessage handles messages from the Dead Letter Queue
func processDLQMessage(msg amqp.Delivery) {
	log.Println("🚨 Dead Letter Event:", string(msg.Body))

	// ✅ Optionally requeue or process the message
	err := msg.Ack(false) // Acknowledge the message to remove it from the DLQ
	if err != nil {
		log.Printf("⚠️ Warning: Failed to acknowledge DLQ message: %v", err)
	}
}

// monitorDLQSize dynamically adjusts the check frequency based on queue size.
func monitorDLQSize(ctx context.Context, ch *amqp.Channel) {
	checkInterval := time.Duration(constants.DLQSizeCheckInterval) * time.Second
	lastSize := -1 // Stores the last recorded size to avoid redundant logs

	for {
		select {
		case <-ctx.Done(): // ✅ Stop monitoring when the consumer exits
			logger.Log.Warn("🚦 Stopping DLQ monitoring")
			return
		default:
			queueInfo, err := ch.QueueInspect(constants.RTDNDLQ)
			if err != nil {
				logger.Log.Error("❌ Failed to inspect DLQ size:", err)
			} else {
				dlqSize := queueInfo.Messages

				// ✅ Update Prometheus Gauge
				metrics.DLQSize.Set(float64(dlqSize))

				// ✅ Log only if DLQ size changes significantly
				if dlqSize != lastSize {
					logger.Log.Infof("📊 Current DLQ size: %d", dlqSize)
					lastSize = dlqSize
				}

				// Dynamically adjust the monitoring interval
				switch {
				case dlqSize == 0:
					checkInterval = time.Duration(constants.DLQCheckIntervalLow) * time.Second
				case dlqSize <= constants.DLQSizeMediumThreshold:
					checkInterval = time.Duration(constants.DLQCheckIntervalMedium) * time.Second
				default:
					checkInterval = time.Duration(constants.DLQCheckIntervalHigh) * time.Second
				}

				if dlqSize > constants.RTDNDLQThreshold {
					triggerDLQAlert(dlqSize)
				}
			}

			// ✅ Ensure `ctx.Done()` is checked before sleeping
			select {
			case <-ctx.Done():
				logger.Log.Warn("🚦 Stopping DLQ monitoring before sleep")
				return
			case <-time.After(checkInterval):
				// Wait for the next check interval (if not stopped)
			}
		}
	}
}

// triggerDLQAlert simulates an alert (this can be expanded to send an email, webhook, etc.)
func triggerDLQAlert(dlqSize int) {
	logger.Log.Errorf("🚨 ALERT! DLQ has %d messages. Immediate investigation required!", dlqSize)
}
