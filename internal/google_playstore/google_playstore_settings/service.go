package playstoresettings

import (
	"context"
	"errors"
	"log"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// UploadServiceAccount processes a new service account upload
func UploadServiceAccount(filePath, fileName string) error {
	if err := DeleteExistingServiceAccount(); err != nil {
		log.Println("⚠️ Error deleting old service account:", err)
	}
	return SaveServiceAccount(filePath, fileName)
}

// ValidateServiceAccount checks if the stored service account is valid
func ValidateServiceAccount() error {
	packageName, err := GetPackageName()
	if err != nil || packageName == "" {
		return errors.New("package name not set. Please configure it first")
	}

	serviceAccount, err := GetLatestServiceAccount()
	if err != nil {
		return errors.New("no service account found")
	}

	if err := ValidateServiceAccountJSONStructure(serviceAccount.FilePath); err != nil {
		return err
	}

	ctx := context.Background()
	service, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(serviceAccount.FilePath))
	if err != nil {
		return err
	}

	_, err = service.Monetization.Subscriptions.List(packageName).Do()
	if err != nil {
		UpdateServiceAccountValidationStatus(false)
		return err
	}

	UpdateServiceAccountValidationStatus(true)
	return nil
}

// GetServiceAccountStatus retrieves the latest service account validation status
func GetServiceAccountStatus() (map[string]interface{}, error) {
	serviceAccount, err := GetLatestServiceAccount()
	if err != nil {
		return nil, err
	}

	packageName, _ := GetPackageName()

	return map[string]interface{}{
		"validated":    serviceAccount.Validated,
		"last_checked": serviceAccount.LastChecked,
		"file_name":    serviceAccount.FileName,
		"package_name": packageName,
	}, nil
}

// DeleteServiceAccount deletes the current service account
func DeleteServiceAccount() error {
	return DeleteExistingServiceAccount()
}

// SetPackageName updates the package name
func SetPackageName(packageName string) error {
	return UpdatePackageName(packageName)
}

// GetPackageName retrieves the package name
func GetPackageName() (string, error) {
	return FetchPackageName()
}
