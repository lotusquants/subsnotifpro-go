package google_play

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// StoreServiceAccountMetadata saves file details in DB
func StoreServiceAccountMetadata(fileName, filePath string) error {
	err := InsertServiceAccount(fileName, filePath)
	if err != nil {
		log.Println("❌ Error inserting service account metadata:", err)
		return err
	}
	log.Println("✅ Service account metadata stored successfully")
	return nil
}

// cleanupOldServiceAccounts removes all previous service account files
func CleanupOldServiceAccounts(directory string) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		filePath := filepath.Join(directory, file.Name())
		if err := os.Remove(filePath); err != nil {
			log.Println("⚠️ Warning: Failed to delete old file:", filePath, err)
		} else {
			log.Println("🗑️ Deleted old service account file:", filePath)
		}
	}

	return nil
}

// ValidateServiceAccount checks if the service account is valid
func ValidateServiceAccount(filePath string) error {
	ctx := context.Background()

	// Step 1: Validate JSON structure
	err := ValidateServiceAccountJSONStructure(filePath)
	if err != nil {
		log.Println("❌ JSON validation failed:", err)
		return err
	}

	// Step 2: Load credentials from JSON file
	service, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(filePath))
	if err != nil {
		log.Println("❌ Error creating Android Publisher Service:", err)
		return err
	}

	// Step 3: Ensure PACKAGE_NAME is set in .env
	packageName := os.Getenv("PACKAGE_NAME")
	if packageName == "" {
		return fmt.Errorf("❌ PACKAGE_NAME environment variable is not set")
	}

	// Step 4: Make API call to verify service account permissions
	_, err = service.Monetization.Subscriptions.List(packageName).Do()
	if err != nil {
		log.Println("❌ Google Play API call failed:", err)
		_ = UpdateServiceAccountValidationStatus(false)
		return err
	}

	// Step 5: Update validation status in DB
	log.Println("✅ Service account validated successfully!")
	_ = UpdateServiceAccountValidationStatus(true)
	return nil
}
