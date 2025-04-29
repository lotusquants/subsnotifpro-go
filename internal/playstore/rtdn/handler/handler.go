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
