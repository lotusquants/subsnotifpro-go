package handler

import (
	"net/http"

	"subsnotifpro-go/internal/google_playstore/settings/service"

	"github.com/gin-gonic/gin"
)

type PlaystoreSettingsHandler struct {
	service service.PlaystoreSettingsService
}

func NewPlaystoreSettingsHandler(s service.PlaystoreSettingsService) *PlaystoreSettingsHandler {
	return &PlaystoreSettingsHandler{service: s}
}

func RegisterPlaystoreSettingsRoutes(rg *gin.RouterGroup, handler *PlaystoreSettingsHandler) {
	ps := rg.Group("/settings")
	ps.POST("/upload-service-account", handler.UploadServiceAccountHandler)
	ps.POST("/validate-service-account", handler.ValidateServiceAccountHandler)
	ps.GET("/service-account-status/:app_id", handler.GetServiceAccountStatusHandler)
}

// UploadServiceAccountHandler handles uploading service account for a specific app
func (h *PlaystoreSettingsHandler) UploadServiceAccountHandler(c *gin.Context) {
	var req struct {
		AppID    string `json:"app_id" binding:"required"`
		FilePath string `json:"file_path" binding:"required"`
		FileName string `json:"file_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UploadServiceAccount(c.Request.Context(), req.AppID, req.FilePath, req.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Service account uploaded successfully."})
}

// ValidateServiceAccountHandler triggers validation of the uploaded service account
func (h *PlaystoreSettingsHandler) ValidateServiceAccountHandler(c *gin.Context) {
	var req struct {
		AppID string `json:"app_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ValidateServiceAccount(c.Request.Context(), req.AppID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Service account validated successfully."})
}

// GetServiceAccountStatusHandler returns the validation status of the service account
func (h *PlaystoreSettingsHandler) GetServiceAccountStatusHandler(c *gin.Context) {
	appID := c.Param("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id is required"})
		return
	}

	status, err := h.service.GetServiceAccountStatus(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "status": status})
}
