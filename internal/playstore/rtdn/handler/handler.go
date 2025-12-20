package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"subsnotifpro-go/internal/playstore/rtdn/dto"
	"subsnotifpro-go/internal/playstore/rtdn/parser"
	"subsnotifpro-go/internal/playstore/rtdn/service"
	"subsnotifpro-go/internal/playstore/rtdn/validator"

	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/utils"

	"github.com/gin-gonic/gin"
)

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

	logger.Log.Infof("--------------------------------------------------------------------------")
	logger.Log.Infof("--------------------------------------------------------------------------")
	logger.Log.Info("📩 Received RTDN webhook")

	defer c.Request.Body.Close()

	// 1. Validate Google Cloud Pub/Sub request
	if err := validator.ValidateGoogleCloudPubSubRequest(c); err != nil {
		logger.Log.WithError(err).Error("Google Cloud Pub/Sub request validation failed")
		utils.WriteGinErrorResponse(c, http.StatusForbidden, "Invalid request source")
		return
	}

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

	// 4. Initialize system fields for the received events (ID, received time, processing status, processed time)
	event.Init()

	// 5. Set idempotency context for middleware
	c.Set("eventID", event.ID.String())
	c.Set("eventType", "google_play_rtdn")

	// 6. Validate payload
	if err := validator.ValidateWebhookPayload(event); err != nil {
		logger.Log.WithError(err).Warn("Validation failed")
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 7. Process event
	if err := h.service.ProcessWebhookEventForPublish(c.Request.Context(), event); err != nil {
		logger.Log.WithError(err).Error("Processing failed")
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Event processing failed")
		return
	}

	// 8. Return success response
	utils.WriteGinJSONResponse(c, http.StatusAccepted, gin.H{
		"message": "Webhook processed successfully",
		"Id":      event.ID.String(),
		"type":    event.GetNotificationType(),
	})
}

func (h *RTDNHandler) parseRTDN(rawBody []byte) (*dto.GooglePlayWebhookEvent, error) {
	// Google Cloud Pub/Sub can send either wrapped or unwrapped messages
	// depending on the push subscription configuration

	// Try Pub/Sub wrapped format first (most common)
	var pubsub struct {
		Message struct {
			Data        string            `json:"data"`
			Attributes  map[string]string `json:"attributes,omitempty"`
			MessageId   string            `json:"messageId,omitempty"`
			PublishTime string            `json:"publishTime,omitempty"`
		} `json:"message"`
		Subscription string `json:"subscription,omitempty"`
	}

	if err := json.Unmarshal(rawBody, &pubsub); err == nil && pubsub.Message.Data != "" {
		logger.Log.Info("Parsing wrapped Pub/Sub RTDN format")
		return parser.ParseWrappedRTDN(pubsub.Message.Data)
	}

	// Fallback to unwrapped format (when push subscription is configured to unwrap)
	logger.Log.Info("Parsing unwrapped RTDN format")
	return parser.ParseUnwrappedRTDN(rawBody)
}

func (h *RTDNHandler) GetDLQSize(c *gin.Context) {
	size, err := h.service.GetDLQSize(c.Request.Context())
	if err != nil {
		logger.Log.WithError(err).Error("failed to get DLQ size")
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "failed to get DLQ size")
		return
	}

	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"dlq_size": size})
}

func (h *RTDNHandler) RetryDLQHandler(c *gin.Context) {
	if err := h.service.RetryMessages(c.Request.Context()); err != nil {
		logger.Log.WithError(err).Error("failed to start DLQ retry")
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "failed to start DLQ retry")
		return
	}

	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{
		"message": "DLQ reprocessing started",
		"status":  "pending",
	})
}
