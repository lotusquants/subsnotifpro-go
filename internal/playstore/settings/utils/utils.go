package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	GooglePlaySecretsDir = "secrets/google_play/"
)

// Helper function to generate a unique file name
func GenerateUniqueFileName(packageName string) string {
	return fmt.Sprintf("%s_%d.json", packageName, time.Now().UnixNano())
}

// Helper function to save the service account file
func SaveServiceAccountFile(filePath, fileName string) (string, error) {
	// Ensure the directory exists
	if err := os.MkdirAll(GooglePlaySecretsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate the new file path
	newFilePath := filepath.Join(GooglePlaySecretsDir, fileName)

	// Move the uploaded file to the new location
	if err := os.Rename(filePath, newFilePath); err != nil {
		return "", fmt.Errorf("failed to save service account file: %w", err)
	}

	return newFilePath, nil
}

// Helper function to delete the service account file
func DeleteServiceAccountFile(fileName string) error {
	filePath := filepath.Join(GooglePlaySecretsDir, fileName)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete service account file: %w", err)
	}
	return nil
}
