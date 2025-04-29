package handlers

import (
	"net/http"
	"subsnotifpro-go/internal/appstore/settings/models"
	"subsnotifpro-go/internal/appstore/settings/service"

	"github.com/gin-gonic/gin"
)

type AppStoreSettingsHandler struct {
	service service.AppStoreSettingsService
}

func NewAppStoreSettingsHandler(service service.AppStoreSettingsService) *AppStoreSettingsHandler {
	return &AppStoreSettingsHandler{service: service}
}

func (h *AppStoreSettingsHandler) GetSettings(c *gin.Context) {
	bundleID := c.Query("bundle_id")
	if bundleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bundle_id parameter is required"})
		return
	}

	settings, err := h.service.GetSettings(bundleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if settings == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "no settings found for this bundle ID"})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *AppStoreSettingsHandler) GetAllSettings(c *gin.Context) {
	settings, err := h.service.GetAllSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *AppStoreSettingsHandler) UpdateSettings(c *gin.Context) {
	var request models.AppStoreSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.UpdateSettings(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AppStoreSettingsHandler) DeleteSettings(c *gin.Context) {
	bundleID := c.Query("bundle_id")
	if bundleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bundle_id parameter is required"})
		return
	}

	err := h.service.DeleteSettings(bundleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings deleted successfully"})
}

func (h *AppStoreSettingsHandler) GenerateJWT(c *gin.Context) {
	bundleID := c.Query("bundle_id")
	if bundleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bundle_id parameter is required"})
		return
	}

	token, err := h.service.GenerateAppStoreJWT(bundleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to generate JWT",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"bundle_id":  bundleID,
		"expires_in": "20 minutes",
	})
}

func (h *AppStoreSettingsHandler) SendTestNotification(c *gin.Context) {
	// Get parameters from request
	bundleID := c.Query("bundle_id")
	if bundleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bundle_id parameter is required"})
		return
	}

	environment := c.Query("environment")
	if environment == "" {
		environment = "production" // default to production
	}

	// Determine if we're using sandbox
	sandbox := environment == "sandbox"

	// Call service
	result, err := h.service.SendTestNotification(c.Request.Context(), bundleID, sandbox)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to send test notification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":                 true,
		"test_notification_token": result.TestNotificationToken,
		"environment":             environment,
	})
}
