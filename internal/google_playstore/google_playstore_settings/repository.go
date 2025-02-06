package playstoresettings

import (
	"errors"
	"log"
	"os"
	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/google_playstore/models"
	"time"
)

// SaveServiceAccount stores new service account metadata
func SaveServiceAccount(filePath, fileName string) error {
	serviceAccount := models.GooglePlayServiceAccount{
		FileName: fileName,
		FilePath: filePath,
	}
	return database.DB.Create(&serviceAccount).Error
}

// GetLatestServiceAccount fetches the latest stored service account
func GetLatestServiceAccount() (models.GooglePlayServiceAccount, error) {
	var serviceAccount models.GooglePlayServiceAccount
	result := database.DB.Order("created_at desc").First(&serviceAccount)
	if result.Error != nil {
		return models.GooglePlayServiceAccount{}, errors.New("no service account found")
	}
	return serviceAccount, nil
}

// UpdateServiceAccountValidationStatus updates the validation status in the database
func UpdateServiceAccountValidationStatus(isValid bool) error {
	var serviceAccount models.GooglePlayServiceAccount
	result := database.DB.Order("created_at desc").First(&serviceAccount)
	if result.Error != nil {
		log.Println("⚠️ No service account found to update validation status.")
		return errors.New("no service account found")
	}

	// Update validation status
	serviceAccount.Validated = isValid
	serviceAccount.LastChecked = time.Now()
	return database.DB.Save(&serviceAccount).Error
}

// DeleteExistingServiceAccount removes the old service account from DB & filesystem
func DeleteExistingServiceAccount() error {
	// Fetch latest service account
	serviceAccount, err := GetLatestServiceAccount()
	if err != nil {
		log.Println("⚠️ No existing service account found to delete")
		return err // ✅ Return error so handler can return 404
	}

	// Remove file from the file system
	if err := os.Remove(serviceAccount.FilePath); err != nil {
		log.Println("⚠️ Error deleting service account file:", err)
		return err
	}

	// Remove from database
	if err := database.DB.Delete(&serviceAccount).Error; err != nil {
		log.Println("⚠️ Error deleting service account record:", err)
		return err
	}

	log.Println("✅ Successfully deleted service account")
	return nil
}

// UpdatePackageName creates or updates the package name
func UpdatePackageName(packageName string) error {
	var settings models.GooglePlaySettings

	// Check if a settings record exists
	err := database.DB.First(&settings).Error
	if err != nil {
		// If not found, create a new settings record
		log.Println("ℹ️ No existing settings found. Creating new settings...")
		newSettings := models.GooglePlaySettings{
			PackageName: packageName,
		}
		return database.DB.Create(&newSettings).Error
	}

	// If settings exist, update the package name
	return database.DB.Model(&settings).Where("id = ?", settings.ID).Update("package_name", packageName).Error
}

// GetPackageName retrieves the package name from Google Play settings
func FetchPackageName() (string, error) {
	var settings models.GooglePlaySettings
	result := database.DB.First(&settings)
	if result.Error != nil {
		log.Println("⚠️ No package name found in database")
		return "", errors.New("package name not found")
	}
	return settings.PackageName, nil
}

// GetSettings fetches Google Play settings from the database
func GetSettings() (models.GooglePlaySettings, error) {
	var settings models.GooglePlaySettings
	result := database.DB.First(&settings)
	if result.Error != nil {
		log.Println("⚠️ No settings found in database")
		return models.GooglePlaySettings{}, result.Error
	}
	return settings, nil
}

// SaveSettings updates the Google Play settings
func SaveSettings(settings models.GooglePlaySettings) error {
	return database.DB.Save(&settings).Error
}

// DeletePackageName removes the package name from the database
func DeletePackageName() error {
	settings, err := GetSettings()
	if err != nil {
		return errors.New("no package name found")
	}

	// Clear the package name field
	settings.PackageName = ""
	return database.DB.Save(&settings).Error
}
