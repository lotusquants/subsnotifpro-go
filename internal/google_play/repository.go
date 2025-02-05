package google_play

import (
	"log"
	"time"

	"subsnotifpro-go/database"
	"subsnotifpro-go/models"
)

// InsertServiceAccount inserts file metadata into the database
func InsertServiceAccount(fileName, filePath string) error {
	serviceAccount := models.GooglePlayServiceAccount{
		FileName:    fileName,
		FilePath:    filePath,
		Validated:   false, // Default to false, validation happens later
		LastChecked: time.Now(),
	}

	return database.DB.Create(&serviceAccount).Error
}

// deleteOldServiceAccountFromDB removes the previous service account metadata
func DeleteOldServiceAccountFromDB() error {
	result := database.DB.Unscoped().Where("1 = 1").Delete(&models.GooglePlayServiceAccount{})
	if result.Error != nil {
		return result.Error
	}
	log.Println("🗑️ Deleted old service account entry from database")
	return nil
}

// UpdateServiceAccountValidationStatus updates validation status in DB
func UpdateServiceAccountValidationStatus(valid bool) error {
	var serviceAccount models.GooglePlayServiceAccount
	result := database.DB.First(&serviceAccount)
	if result.Error != nil {
		log.Println("❌ Error fetching service account:", result.Error)
		return result.Error
	}

	serviceAccount.Validated = valid
	serviceAccount.LastChecked = time.Now()
	result = database.DB.Save(&serviceAccount)
	if result.Error != nil {
		log.Println("❌ Error updating validation status:", result.Error)
		return result.Error
	}

	log.Println("✅ Updated validation status:", valid)
	return nil
}
