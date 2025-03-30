// internal/google_playstore/rtdn/queue/publisher.go
package dispatch

import (
	"encoding/json"
	"fmt"
	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/logger"

	"github.com/streadway/amqp"
)

// PublishToQueue sends webhook data to RabbitMQ for background processing
func PublishToQueue(event interface{}, ch *amqp.Channel) error {
	// ✅ Ensure channel is not nil
	if ch == nil {
		logger.Log.Error("❌ RabbitMQ channel is nil")
		return fmt.Errorf("RabbitMQ channel is nil")
	}

	// ✅ Marshal the event data into JSON
	body, err := json.Marshal(event)
	if err != nil {
		logger.Log.Error("❌ Failed to marshal event:", err)
		return err
	}

	logger.Log.Info("✅ Successfully marshaled event. Publishing to RabbitMQ...")

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

	logger.Log.Info("✅ Webhook event published to RabbitMQ queue:", event)
	return nil
}
