package utils

import (
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// saveUploadedFile is a helper function to save an uploaded file securely.
func SaveUploadedFile(c *gin.Context, file *multipart.FileHeader, dir string) (string, error) {
	newFileName := GenerateUniqueFileName()
	filePath := filepath.Join(dir, newFileName)

	// Ensure the directory exists
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	// Clean path to prevent traversal attacks
	absolutePath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}

	// Save the uploaded file
	if err := c.SaveUploadedFile(file, absolutePath); err != nil {
		return "", err
	}

	return absolutePath, nil
}

// generateUniqueFileName creates a unique file name to prevent conflicts.
func GenerateUniqueFileName() string {
	return uuid.NewString() + ".json"
}

// isValidJSONFile checks if the uploaded file is a valid JSON file.
func IsValidJSONFile(file *multipart.FileHeader) bool {
	return strings.HasSuffix(strings.ToLower(file.Filename), ".json")
}
