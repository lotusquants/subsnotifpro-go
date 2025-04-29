package handler

import (
	"net/http"
	"subsnotifpro-go/internal/appstore/webhooks/converter"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/appstore/webhooks/service"

	"subsnotifpro-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AppStoreNotificationsHandler handles incoming App Store Server Notifications
type AppStoreNotificationsHandler struct {
	service   service.AppStoreNotificationsService
	converter converter.NotificationConverter
}

// NewAppStoreNotificationsHandler creates a new instance of AppStoreHandler
func NewAppStoreNotificationsHandler(service service.AppStoreNotificationsService) *AppStoreNotificationsHandler {
	return &AppStoreNotificationsHandler{
		service:   service,
		converter: *converter.NewNotificationConverter()}
}

// NotificationHandler handles incoming App Store notifications
func (h *AppStoreNotificationsHandler) NotificationHandler(c *gin.Context) {
	logger.Log.Info("📩 Received App Store notification")
	defer c.Request.Body.Close()

	// 1. Validate request
	if !h.validateRequest(c) {
		return
	}

	// 2. Parse notification
	// Read and parse the raw notification
	var rawNotification *dto.ResponseBodyV2
	if err := c.ShouldBindJSON(&rawNotification); err != nil {
		logger.Log.Errorf("Failed to parse notification: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification format"})
		return
	}

	// 3. Convert to Full DTO - Full JWS decoding
	notification, err := h.converter.ConvertToFullDTO(rawNotification)
	if err != nil {
		logger.Log.Errorf("JWS decoding failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signed payload"})
		return
	}

	// 3. Process notification
	if err := h.service.ProcessNotificationForPublish(c.Request.Context(), notification); err != nil {
		logger.Log.Errorf("Failed to process notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process notification"})
		return
	}

	// 4. Respond successfully
	c.JSON(http.StatusAccepted, gin.H{"status": "OK"})
}

// validateRequest performs basic request validation
func (h *AppStoreNotificationsHandler) validateRequest(c *gin.Context) bool {
	// Verify it's a POST request
	if c.Request.Method != http.MethodPost {
		logger.Log.Warn("Invalid HTTP method for App Store notification")
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
		return false
	}

	// Verify content type
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		logger.Log.Warnf("Invalid content type: %s", contentType)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content type"})
		return false
	}

	// Add IP whitelisting check here if needed
	// Apple publishes their IP ranges for verification

	return true
}
