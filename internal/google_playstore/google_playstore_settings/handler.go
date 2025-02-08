package playstoresettings

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadServiceAccountHandler handles service account JSON uploads
func UploadServiceAccountHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("❌ Error retrieving file:", err)

		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Generate a unique file name with a UUID
	newFileName := file.Filename + "-" + uuid.New().String() + ".json"

	// Ensure secrets directory exists
	secretsDir := "secrets/google_play/"
	if err := os.MkdirAll(secretsDir, os.ModePerm); err != nil {
		log.Println("❌ Error creating secrets directory:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create storage directory"})
		return
	}

	// Define file path and save the file
	filePath := filepath.Join(secretsDir, newFileName)
	absolutePath, err := filepath.Abs(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error resolving file path"})
		return
	}
	if err := c.SaveUploadedFile(file, absolutePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File upload failed"})
		return
	}

	// Delegate to service layer
	err = UploadServiceAccount(absolutePath, newFileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process service account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service account uploaded successfully"})
}

// ValidateServiceAccountHandler checks if the stored service account is valid
func ValidateServiceAccountHandler(c *gin.Context) {
	if err := ValidateServiceAccount(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Service account validated successfully"})
}

// GetServiceAccountStatusHandler fetches service account status
func GetServiceAccountStatusHandler(c *gin.Context) {
	status, err := GetServiceAccountStatus()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No service account found"})
		return
	}
	c.JSON(http.StatusOK, status)
}

// DeleteServiceAccountHandler removes the current service account
func DeleteServiceAccountHandler(c *gin.Context) {
	err := DeleteExistingServiceAccount()
	if err != nil {
		log.Println("⚠️ Service account not found, returning 404")
		c.JSON(http.StatusNotFound, gin.H{"error": "No service account found to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Service account deleted successfully"})
}

// SetPackageNameHandler sets or updates the package name
func SetPackageNameHandler(c *gin.Context) {
	var request struct {
		PackageName string `json:"package_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Package name is required"})
		return
	}

	if err := SetPackageName(request.PackageName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set package name"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Package name updated successfully"})
}

// GetPackageNameHandler retrieves the package name
func GetPackageNameHandler(c *gin.Context) {
	packageName, err := GetPackageName()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No package name set"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"package_name": packageName})
}
