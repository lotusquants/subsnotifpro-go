package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/playstore/rtdn/dto"
	"subsnotifpro-go/internal/playstore/rtdn/parser"
	"subsnotifpro-go/internal/playstore/rtdn/service"
	"subsnotifpro-go/internal/playstore/rtdn/validator"

	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/utils"
	"subsnotifpro-go/queue"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// RegisterRTDNRoutes registers all RTDN webhook and DLQ management routes.
func RegisterRTDNRoutes(r *gin.RouterGroup, handler *RTDNHandler) {
	r.POST("/rtdn/webhooks", func(c *gin.Context) {
		handler.WebhookHandler(c)
	})
	r.GET("/rtdn/dlq/size", GetDLQSize)
	r.POST("/rtdn/dlq/retry", RetryDLQHandler)
}

// Define your handler struct with a field for the service
type RTDNHandler struct {
	service service.RTDNService
}

// NewWebhookHandler creates a new instance of WebhookHandler with the service injected
func NewRTDNHandler(service service.RTDNService) *RTDNHandler {
	return &RTDNHandler{service: service}
}

// WebhookHandler handles incoming RTDN requests (both wrapped and unwrapped).
func (h *RTDNHandler) WebhookHandler(c *gin.Context) {

	logger.Log.Info("📩 Received RTDN webhook request")
	defer c.Request.Body.Close()

	// 1. JWT Validation

	// 2. Read and parse request
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to read request body")
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 3. Parse based on format (wrapped/unwrapped)
	event, err := h.parseRTDN(rawBody)
	if err != nil {
		logger.Log.WithError(err).Error("Failed to parse RTDN")
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "Invalid RTDN format")
		return
	}

	// 4. Initialize system fields for the recived events (ID, recieved time, processing status, processed time)
	event.Init()

	// 5. Validate payload
	if err := validator.ValidateWebhookPayload(event); err != nil {
		logger.Log.WithError(err).Warn("Validation failed")
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 6. Process event
	if err := h.service.ProcessWebhookEventForPublish(c.Request.Context(), event); err != nil {
		logger.Log.WithError(err).Error("Processing failed")
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Event processing failed")
		return
	}

	// 7. Return success response
	utils.WriteGinJSONResponse(c, http.StatusAccepted, gin.H{
		"message": "Webhook processed successfully",
		"Id":      event.ID.String(),
		"type":    event.GetNotificationType(),
	})
}

func (h *RTDNHandler) parseRTDN(rawBody []byte) (*dto.GooglePlayWebhookEvent, error) {
	// Try Pub/Sub format first
	var pubsub struct {
		Message struct {
			Data string `json:"data"`
		} `json:"message"`
	}

	if json.Unmarshal(rawBody, &pubsub) == nil && pubsub.Message.Data != "" {
		return parser.ParseWrappedRTDN(pubsub.Message.Data)
	}

	// Fallback to direct format
	return parser.ParseUnwrappedRTDN(rawBody)
}

// // processEvent validates, stores, and enqueues RTDN messages.
// func (h *RTDNHandler) processEvent(c *gin.Context, event *dto.GooglePlayWebhookEvent, ch *amqp.Channel) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			logger.Log.Errorf("🚨 Panic recovered in processEvent: %v", r)
// 		}
// 	}()

// 	if err := h.service.SaveWebhookEvent(&event); err != nil {
// 		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to store event")
// 		return
// 	}

// 	if err := dispatch.PublishToQueue(event, ch); err != nil {
// 		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to queue event")
// 		return
// 	}

// 	logger.Log.Infof("✅ RTDN webhook received and queued successfully for product: %s", event.PackageName)
// 	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"message": "Webhook received successfully"})
// }

// GetDLQSize fetches the size of the RTDN Dead Letter Queue.
func GetDLQSize(c *gin.Context) {
	conn, ch, err := queue.GetAdminChannel()
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to connect to RabbitMQ")
		return
	}
	defer conn.Close()
	defer ch.Close()

	queueInfo, err := ch.QueueInspect(constants.RTDNDLQ)
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to check DLQ size")
		return
	}

	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"dlq_size": queueInfo.Messages})
}

// RetryDLQHandler retries all messages in the DLQ with retry limits.
func RetryDLQHandler(c *gin.Context) {
	conn, ch, err := queue.GetAdminChannel()
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to connect to RabbitMQ")
		return
	}
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(constants.RTDNDLQ, "", false, false, false, false, nil)
	if err != nil {
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to consume DLQ")
		return
	}

	go retryDLQMessages(ch, msgs)

	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"message": "DLQ retry process started"})
}

// retryDLQMessages handles processing messages from the DLQ.
func retryDLQMessages(ch *amqp.Channel, msgs <-chan amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("🚨 Recovered from panic in retryDLQMessages: %v", r)
		}
	}()

	for msg := range msgs {
		messageID := msg.MessageId
		if messageID == "" {
			messageID = "unknown"
		}
		retryCount := getRetryCount(msg.Headers)

		if retryCount > constants.MaxRetries {
			logger.Log.WithFields(map[string]interface{}{
				"messageID":  messageID,
				"retryCount": retryCount,
				"permanent":  true,
			}).Warn("🚨 Permanently discarding message after exceeding max retries")
			msg.Ack(false)
			continue
		}

		logger.Log.WithFields(map[string]interface{}{
			"messageID":  messageID,
			"retryCount": retryCount + 1,
		}).Info("🔄 Retrying message from DLQ")

		err := ch.Publish("", constants.RTDNQueue, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        msg.Body,
			Headers:     amqp.Table{"x-retry-count": retryCount + 1},
		})
		if err != nil {
			logger.Log.Errorf("❌ Failed to re-publish message from DLQ: %v", err)
			msg.Nack(false, true)              // Ensures message is requeued if processing fails
			time.Sleep(time.Millisecond * 500) // Exponential backoff
			continue
		}

		msg.Ack(false)
		time.Sleep(time.Millisecond * 250) // Prevents message flooding
	}
}

// getRetryCount safely extracts retry count from headers.
func getRetryCount(headers amqp.Table) int {
	if val, ok := headers["x-retry-count"]; ok {
		strVal := fmt.Sprintf("%v", val)
		if num, err := strconv.Atoi(strVal); err == nil {
			return num
		}
	}
	return 0
}
