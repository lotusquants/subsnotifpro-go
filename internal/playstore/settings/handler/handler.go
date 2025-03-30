package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/playstore/settings/service"
)

type PlaystoreSettingsHandler struct {
	service service.PlaystoreSettingsService
}

func NewPlaystoreSettingsHandler(svc service.PlaystoreSettingsService) *PlaystoreSettingsHandler {
	return &PlaystoreSettingsHandler{service: svc}
}

func (h *PlaystoreSettingsHandler) SaveSettingsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	packageName := c.PostForm("package_name")
	file, err := c.FormFile("service_account")
	if packageName == "" || err != nil || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing package_name or service_account file"})
		return
	}

	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to open file"})
		return
	}
	defer fileReader.Close()

	if err := h.service.SaveOrUpdatePlayStoreSettings(ctx, packageName, fileReader, file.Filename); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"package": packageName,
			"error":   err.Error(),
		}).Error("❌ Failed to save settings")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to save settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Settings saved successfully"})
}

func (h *PlaystoreSettingsHandler) GetSettingsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	packageName := c.Query("package_name")
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	data, err := h.service.GetSettings(ctx, packageName)
	if err != nil {
		logger.Log.WithField("error", err.Error()).Error("❌ Failed to fetch settings")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (h *PlaystoreSettingsHandler) DeleteSettingsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	packageName := c.Query("package_name")
	if packageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "package_name is required"})
		return
	}

	if err := h.service.DeleteSettings(ctx, packageName); err != nil {
		logger.Log.WithField("error", err.Error()).Error("❌ Failed to delete settings")
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Settings deleted successfully"})
}
