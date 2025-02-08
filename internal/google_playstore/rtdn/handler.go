// internal/google_playstore/rtdn/handler.go
package rtdn

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/google_playstore/models"
	"subsnotifpro-go/internal/google_playstore/rtdn/queue"
	"subsnotifpro-go/internal/google_playstore/rtdn/service"
	"subsnotifpro-go/internal/google_playstore/rtdn/validator"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/messaging"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// WebhookHandler handles both wrapped and unwrapped RTDN messages
func WebhookHandler(c *gin.Context, ch *amqp.Channel) {

	// // ✅ Extract and validate JWT
	// authHeader := c.GetHeader("Authorization")
	// expectedAudience := os.Getenv("GOOGLE_PLAY_PROJECT_ID")

	// if _, err := validator.ValidateJWT(authHeader, expectedAudience); err != nil {
	// 	logger.Log.Warnf("❌ Unauthorized RTDN request received: %v", err)
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized webhook"})
	// 	return
	// }

	// ✅ Read the raw request body
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Error("❌ Failed to read request body:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// ✅ Detect wrapped or unwrapped message
	var pubsubPayload struct {
		Message struct {
			Data string `json:"data"`
		} `json:"message"`
	}

	if err := json.Unmarshal(rawBody, &pubsubPayload); err == nil && pubsubPayload.Message.Data != "" {
		// ✅ Case 1: Wrapped Pub/Sub message
		log.Println("🔄 Detected wrapped RTDN Pub/Sub message.")

		// ✅ Decode the Base64-encoded message data
		event, err := validator.DecodeRTDNMessage(pubsubPayload.Message.Data)
		if err != nil {
			logger.Log.Errorf("❌ Failed to decode RTDN message: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid RTDN message"})
			return
		}

		processEvent(c, event, ch)
		return
	}

	// ✅ Case 2: Unwrapped Pub/Sub message (raw JSON)
	log.Println("🔄 Detected unwrapped RTDN message.")

	// ✅ Decode unwrapped message
	var tempPayload map[string]interface{}
	if err := json.Unmarshal(rawBody, &tempPayload); err != nil {
		logger.Log.Errorf("❌ Invalid unwrapped RTDN JSON format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// ✅ Convert `eventTimeMillis` safely to int64
	eventTimeInt, err := validator.ParseEventTimeMillis(tempPayload["eventTimeMillis"])
	if err != nil {
		logger.Log.Errorf("❌ Failed to convert eventTimeMillis: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid eventTimeMillis format"})
		return
	}

	// ✅ Convert temp map to event struct
	rawJSON, _ := json.Marshal(tempPayload)
	event := models.GooglePlayWebhookEvent{
		Version:         validator.SafeString(tempPayload["version"]),
		PackageName:     validator.SafeString(tempPayload["packageName"]),
		EventTimeMillis: eventTimeInt,
		RawPayload:      string(rawJSON),
		Status:          "pending",
		RetryCount:      0,
	}

	// ✅ Extract notification type (Only one should be present)
	if subNotification, exists := tempPayload["subscriptionNotification"]; exists {
		event.SubscriptionNotification = validator.SafeUnmarshal[models.SubscriptionNotification](subNotification)
	}

	if oneTimeNotification, exists := tempPayload["oneTimeProductNotification"]; exists {
		event.OneTimeProductNotification = validator.SafeUnmarshal[models.OneTimeProductNotification](oneTimeNotification)
	}

	if voidedNotification, exists := tempPayload["voidedPurchaseNotification"]; exists {
		event.VoidedPurchaseNotification = validator.SafeUnmarshal[models.VoidedPurchaseNotification](voidedNotification)
	}

	if testNotification, exists := tempPayload["testNotification"]; exists {
		event.TestNotification = validator.SafeUnmarshal[models.TestNotification](testNotification)
	}

	processEvent(c, event, ch)
}

// processEvent validates, stores, and queues the event
func processEvent(c *gin.Context, event models.GooglePlayWebhookEvent, ch *amqp.Channel) {
	// ✅ Validate webhook payload
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
	if err := queue.PublishToQueue(event, ch); err != nil {
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
