// internal/google_playstore/rtdn/handler.go
package rtdn

import (
	"context"
	"net/http"
	"os"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/google_playstore/rtdn/queue"
	"subsnotifpro-go/internal/google_playstore/rtdn/service"
	"subsnotifpro-go/internal/google_playstore/rtdn/validator"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/messaging"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// WebhookHandler receives and stores Google Play events
func WebhookHandler(c *gin.Context) {

	// ✅ Extract and validate JWT
	authHeader := c.GetHeader("Authorization")
	expectedAudience := os.Getenv("GOOGLE_PLAY_PROJECT_ID")

	if _, err := validator.ValidateJWT(authHeader, expectedAudience); err != nil {
		logger.Log.Warnf("❌ Unauthorized RTDN request received: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized webhook"})
		return
	}

	// ✅ Parse incoming JSON request
	var requestPayload struct {
		Message struct {
			Data string `json:"data"`
		} `json:"message"`
	}

	// ✅ Bind JSON request into requestPayload
	if err := c.ShouldBindJSON(&requestPayload); err != nil {
		logger.Log.Errorf("❌ Invalid webhook JSON format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// ✅ Handle empty payload before decoding
	if requestPayload.Message.Data == "" {
		logger.Log.Warn("⚠️ Empty RTDN payload received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty RTDN payload"})
		return
	}

	// ✅ Decode the base64-encoded RTDN message
	event, err := validator.DecodeRTDNMessage(requestPayload.Message.Data)
	if err != nil {
		logger.Log.Errorf("❌ Failed to decode RTDN message: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid RTDN message"})
		return
	}

	// ✅ Validate webhook payload before storing
	if err := validator.ValidateWebhookPayload(&event); err != nil {
		logger.Log.Errorf("❌ Invalid RTDN payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ✅ Store webhook event in DB
	if err := service.SaveWebhookEvent(&event); err != nil {
		logger.Log.Errorf("❌ Database error: Failed to store event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store event"})
		return
	}

	// ✅ Push event to RabbitMQ for processing
	if err := queue.PublishToQueue(event); err != nil {
		logger.Log.Errorf("❌ RabbitMQ error: Failed to queue event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue event"})
		return
	}

	logger.Log.Info("✅ RTDN webhook received and queued successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received successfully"})
}

// GetDLQSize returns the number of messages in the Dead Letter Queue
func GetDLQSize(c *gin.Context) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Log.Error("❌ Failed to connect to RabbitMQ:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to RabbitMQ"})
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Log.Error("❌ Failed to open RabbitMQ channel:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open RabbitMQ channel"})
		return
	}
	defer ch.Close()

	queueInfo, err := ch.QueueInspect(constants.RTDNDLQ)
	if err != nil {
		logger.Log.Error("❌ Failed to inspect DLQ size:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get DLQ size"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dlq_size": queueInfo.Messages})
}

// RetryDLQHandler retries failed events from DLQ with a failure limit
func RetryDLQHandler(c *gin.Context) {
	ch, err := messaging.GetChannel(context.Background())
	if err != nil {
		logger.Log.Error("❌ Failed to get RabbitMQ channel:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to RabbitMQ"})
		return
	}

	msgs, err := ch.Consume(constants.RTDNDLQ, "", false, false, false, false, nil)
	if err != nil {
		logger.Log.Error("❌ Failed to consume DLQ:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to consume DLQ"})
		return
	}

	go func() {
		for msg := range msgs {
			// Extract retry count from headers
			retryCount := 0
			if val, ok := msg.Headers["x-retry-count"].(int32); ok {
				retryCount = int(val)
			}

			// If retry exceeds threshold, log permanently failed event
			if retryCount > constants.MaxRetries {
				logger.Log.Warnf("🚨 Permanently failed event: %s", string(msg.Body))
				msg.Ack(false) // Remove from DLQ permanently
				continue
			}

			// Retry event processing
			logger.Log.Warn("🔄 Retrying DLQ event:", string(msg.Body))
			ch.Publish("", constants.RTDNQueue, false, false, amqp.Publishing{
				ContentType: "application/json",
				Body:        msg.Body,
				Headers:     amqp.Table{"x-retry-count": retryCount + 1},
			})
			msg.Ack(false)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Retrying DLQ messages..."})
}
