package handler

import (
	"net/http"
	"subsnotifpro-go/internal/app/models"
	"subsnotifpro-go/internal/app/service"

	"github.com/gin-gonic/gin"
)

type PlaystoreSetupHandlerInterface interface {
	SaveSetup(c *gin.Context)
	ValidateServiceAccount(c *gin.Context)
	GetSetupByAppID(c *gin.Context)
	UpdateBucketID(c *gin.Context)
}

type playstoreSetupHandler struct {
	service service.PlaystoreSetupServiceInterface
}

// NewPlaystoreSetupHandler returns the interface for route injection
func NewPlaystoreSetupHandler(service service.PlaystoreSetupServiceInterface) PlaystoreSetupHandlerInterface {
	return &playstoreSetupHandler{service: service}
}

func (h *playstoreSetupHandler) SaveSetup(c *gin.Context) {
	var setup models.PlaystoreSetup
	if err := c.ShouldBindJSON(&setup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.SaveSetup(c.Request.Context(), &setup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save setup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setup saved successfully"})
}

func (h *playstoreSetupHandler) ValidateServiceAccount(c *gin.Context) {
	var req struct {
		AppID string `json:"app_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id is required"})
		return
	}

	if err := h.service.ValidateServiceAccount(c.Request.Context(), req.AppID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service account validated successfully"})
}

func (h *playstoreSetupHandler) GetSetupByAppID(c *gin.Context) {
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id query param is required"})
		return
	}

	setup, err := h.service.GetByAppID(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setup not found"})
		return
	}

	c.JSON(http.StatusOK, setup)
}

func (h *playstoreSetupHandler) UpdateBucketID(c *gin.Context) {
	var req struct {
		AppID    string `json:"app_id" binding:"required"`
		BucketID string `json:"bucket_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id and bucket_id are required"})
		return
	}

	if err := h.service.UpdateBucketID(c.Request.Context(), req.AppID, req.BucketID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bucket ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bucket ID updated successfully"})
}
