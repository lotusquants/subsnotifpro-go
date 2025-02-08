// internal/google_playstore/rtdn/queue/publisher.go
package queue

import (
	"context"
	"encoding/json"
	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/messaging"

	"github.com/streadway/amqp"
)

// PublishToQueue sends webhook data to RabbitMQ for background processing
func PublishToQueue(event interface{}) error {
	// ✅ Get a shared RabbitMQ channel from messaging package

	ch, err := messaging.GetChannel(context.Background()) // ✅ Ensure context is passed
	if err != nil {
		logger.Log.Error("❌ Failed to get RabbitMQ channel:", err)
		return err
	}

	// ✅ Ensure that the channel is not nil before proceeding
	if ch == nil {
		logger.Log.Error("❌ RabbitMQ channel is nil")
		return err
	}

	// ✅ Marshal the event data into JSON
	body, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("❌ Failed to marshal event:", err)
		return err
	}

	// ✅ Publish the message to the queue
	err = ch.Publish(
		"", constants.RTDNQueue, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers:     amqp.Table{"x-retry-count": 0}, // ✅ Set retry count to 0 initially
		},
	)
	if err != nil {
		logger.Log.Error("❌ Failed to publish message:", err)
		return err
	}

	logger.Log.Info("✅ Webhook event published to queue:", event)
	return nil
}
