package google_play

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"subsnotifpro-go/database"
	"subsnotifpro-go/models"

	"github.com/gin-gonic/gin"
)

// UploadServiceAccountHandler handles file upload for Google Play service accounts
func UploadServiceAccountHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Ensure secrets directory exists
	secretsDir := "secrets/google_play/"
	if err := os.MkdirAll(secretsDir, os.ModePerm); err != nil {
		log.Println("❌ Error creating secrets directory:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create storage directory"})
		return
	}

	// Delete existing service account files from storage
	err = CleanupOldServiceAccounts(secretsDir)
	if err != nil {
		log.Println("❌ Error deleting old service account files:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clean up old files"})
		return
	}

	// Define new file path
	filePath := filepath.Join(secretsDir, file.Filename)

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Println("❌ Error saving file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "File upload failed"})
		return
	}

	// Delete previous service account entry in DB
	err = DeleteOldServiceAccountFromDB()
	if err != nil {
		log.Println("❌ Error deleting old service account record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove old metadata"})
		return
	}

	// Store metadata in the database via service layer
	serviceErr := StoreServiceAccountMetadata(file.Filename, filePath)
	if serviceErr != nil {
		log.Println("❌ Error storing metadata:", serviceErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store file metadata"})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Service account uploaded successfully",
		"file":    file.Filename,
		"path":    filePath,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// ValidateServiceAccountHandler handles service account validation
func ValidateServiceAccountHandler(c *gin.Context) {
	// Retrieve the latest stored service account file path
	var serviceAccount models.GooglePlayServiceAccount
	result := database.DB.First(&serviceAccount)
	if result.Error != nil {
		log.Println("❌ No service account found:", result.Error)
		c.JSON(http.StatusNotFound, gin.H{"error": "No service account found"})
		return
	}

	// Check if file exists
	filePath := serviceAccount.FilePath
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("❌ Service account file not found:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service account file not found"})
		return
	}

	// Validate JSON structure
	if err := ValidateServiceAccountJSONStructure(filePath); err != nil {
		log.Println("❌ Invalid JSON format:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate Service Account via Google Play API
	err := ValidateServiceAccount(filePath)
	if err != nil {
		log.Println("❌ Service account validation failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service account validated successfully"})
}
