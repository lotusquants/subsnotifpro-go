package handler

import (
	"net/http"

	"subsnotifpro-go/internal/google_playstore/settings/service"
	settingsUtils "subsnotifpro-go/internal/google_playstore/settings/utils"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/utils"

	"github.com/gin-gonic/gin"
)

// GooglePlaySecretsDir defines the storage location for service accounts.
const GooglePlaySecretsDir = "secrets/google_play/"

// AllowedFileType ensures only JSON files are accepted.
const AllowedFileType = "application/json"

// Define your handler struct with a field for the service
type PlaystoreSettingsHandler struct {
	service service.PlaystoreSettingsService
}

// NewWebhookHandler creates a new instance of WebhookHandler with the service injected
func NewPlaystoreSettingsHandler(service service.PlaystoreSettingsService) *PlaystoreSettingsHandler {
	return &PlaystoreSettingsHandler{
		service: service,
	}
}

// RegisterRoutes registers all settings-related routes within the provided router group.
func RegisterPlaystoreSettingsRoutes(r *gin.RouterGroup, handler *PlaystoreSettingsHandler) {
	r.POST("/upload-service-account", handler.UploadServiceAccount)
	r.POST("/validate-service-account", handler.ValidateServiceAccount)
	r.GET("/service-account-status", handler.GetServiceAccountStatus)
	r.POST("/set-package-name", handler.SetPackageName)
	r.GET("/get-package-name", handler.GetPackageName)
}

// UploadServiceAccount handles uploading the service account JSON file.
func (h *PlaystoreSettingsHandler) UploadServiceAccount(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		logger.Log.Errorf("❌ File upload error: %v", err)
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "File is required")
		return
	}

	// Validate file type
	if !settingsUtils.IsValidJSONFile(file) {
		logger.Log.Errorf("❌ Invalid file type: %s", file.Filename)
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "Only JSON files are allowed")
		return
	}

	absolutePath, err := settingsUtils.SaveUploadedFile(c, file, GooglePlaySecretsDir)
	if err != nil {
		logger.Log.Errorf("❌ File save error (%s): %v", file.Filename, err)
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "File upload failed")
		return
	}

	if err := h.service.UploadServiceAccount(absolutePath, file.Filename); err != nil {
		logger.Log.Errorf("❌ Failed to process service account (%s): %v", file.Filename, err)
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to process service account")
		return
	}

	logger.Log.Infof("✅ Service account uploaded successfully: %s", file.Filename)
	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"message": "Service account uploaded successfully"})
}

// ValidateServiceAccount validates the stored service account.
func (h *PlaystoreSettingsHandler) ValidateServiceAccount(c *gin.Context) {
	if err := h.service.ValidateServiceAccount(); err != nil {
		logger.Log.Warnf("❌ Service account validation failed: %v", err)
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"message": "Service account validated successfully"})
}

// GetServiceAccountStatus retrieves the current service account status.
func (h *PlaystoreSettingsHandler) GetServiceAccountStatus(c *gin.Context) {
	status, err := h.service.GetServiceAccountStatus()
	if err != nil {
		logger.Log.Warnf("❌ No service account found: %v", err)
		utils.WriteGinErrorResponse(c, http.StatusNotFound, "No service account found")
		return
	}
	utils.WriteGinJSONResponse(c, http.StatusOK, status)
}

// SetPackageName sets the package name.
func (h *PlaystoreSettingsHandler) SetPackageName(c *gin.Context) {
	var request struct {
		PackageName string `json:"package_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Log.Warnf("❌ Package name binding failed: %v", err)
		utils.WriteGinErrorResponse(c, http.StatusBadRequest, "Package name is required")
		return
	}

	if err := h.service.SetPackageName(request.PackageName); err != nil {
		logger.Log.Errorf("❌ Failed to set package name: %v", err)
		utils.WriteGinErrorResponse(c, http.StatusInternalServerError, "Failed to set package name")
		return
	}

	logger.Log.Infof("✅ Package name updated: %s", request.PackageName)
	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"message": "Package name updated successfully"})
}

// GetPackageName retrieves the package name.
func (h *PlaystoreSettingsHandler) GetPackageName(c *gin.Context) {
	packageName, err := h.service.GetPackageName()
	if err != nil {
		logger.Log.Warnf("❌ No package name set: %v", err)
		utils.WriteGinErrorResponse(c, http.StatusNotFound, "No package name set")
		return
	}
	utils.WriteGinJSONResponse(c, http.StatusOK, gin.H{"package_name": packageName})
}
