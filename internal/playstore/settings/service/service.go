package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	apiService "subsnotifpro-go/internal/playstore/api/service"
	"subsnotifpro-go/internal/playstore/settings/models"
	"subsnotifpro-go/internal/playstore/settings/repository"
	"subsnotifpro-go/internal/playstore/settings/validator"

	"gorm.io/gorm"
)

type PlaystoreSettingsService interface {
	SaveOrUpdatePlayStoreSettings(ctx context.Context, packageName string, fileReader io.Reader, fileName string) error
	GetSettings(ctx context.Context, packageName string) (map[string]interface{}, error)
	DeleteSettings(ctx context.Context, packageName string) error
	GetServiceAccountPath(ctx context.Context, packageName string) (string, error)
}

type playstoreSettingsService struct {
	repo       repository.PlaystoreSettingsRepository
	apiService apiService.PlaystoreApiService
	db         *gorm.DB
}

func NewPlaystoreSettingsService(repo repository.PlaystoreSettingsRepository, api apiService.PlaystoreApiService, db *gorm.DB) PlaystoreSettingsService {
	return &playstoreSettingsService{repo: repo, apiService: api, db: db}
}

func (s *playstoreSettingsService) SaveOrUpdatePlayStoreSettings(ctx context.Context, packageName string, fileReader io.Reader, fileName string) error {
	if packageName == "" || fileReader == nil || fileName == "" {
		return errors.New("invalid input")
	}

	// Read file once
	fileBytes, err := io.ReadAll(fileReader)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		oldFilePath, _ := s.getOldFilePath(ctx, tx, packageName) // ignore not found

		// 🔍 Validate structure and permissions
		if err := s.validateServiceAccount(ctx, fileBytes, packageName); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		// 📂 Save file
		uploadDir := "secrets/google_play"
		_ = os.MkdirAll(uploadDir, os.ModePerm)

		uniqueFileName := fmt.Sprintf("%s_%d_%s", packageName, time.Now().Unix(), fileName)
		filePath := filepath.Join(uploadDir, uniqueFileName)

		if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		// 🧹 Cleanup old file
		if oldFilePath != "" && oldFilePath != filePath {
			_ = os.Remove(oldFilePath)
		}

		// 💾 Save settings + service account
		settings := &models.GooglePlaySettings{PackageName: packageName}
		if err := s.repo.SavePackageSettings(ctx, tx, settings); err != nil {
			return fmt.Errorf("save settings: %w", err)
		}

		account := &models.GooglePlayServiceAccount{
			PackageName: packageName,
			FileName:    uniqueFileName,
			FilePath:    filePath,
			Validated:   true,
			LastChecked: ptrTime(time.Now()),
		}
		if err := s.repo.UploadServiceAccount(ctx, tx, account); err != nil {
			return fmt.Errorf("save service account: %w", err)
		}
		return nil
	})
}

func (s *playstoreSettingsService) validateServiceAccount(ctx context.Context, fileBytes []byte, packageName string) error {
	tmp, err := os.CreateTemp("", "sa-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(fileBytes); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	if err := validator.ValidateServiceAccountJSONStructure(tmp.Name()); err != nil {
		return err
	}

	ok, err := s.apiService.VerifyServiceAccountAccess(ctx, tmp.Name(), packageName)
	if err != nil {
		return fmt.Errorf("verify access: %w", err)
	}
	if !ok {
		return fmt.Errorf("no access to package: %s", packageName)
	}
	return nil
}

func (s *playstoreSettingsService) GetSettings(ctx context.Context, packageName string) (map[string]interface{}, error) {
	account, err := s.repo.GetServiceAccount(ctx, s.db, packageName)
	if err != nil {
		return nil, err
	}
	settings, err := s.repo.GetSettingsByPackageName(ctx, s.db, packageName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"validated":    account.Validated,
		"last_checked": account.LastChecked,
		"file_name":    account.FileName,
		"package_name": settings.PackageName,
	}, nil
}

func (s *playstoreSettingsService) DeleteSettings(ctx context.Context, packageName string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		account, err := s.repo.GetServiceAccount(ctx, tx, packageName)
		if err != nil {
			return err
		}
		if err := s.repo.DeletePackageSettings(ctx, tx, packageName); err != nil {
			return err
		}
		if err := os.Remove(account.FilePath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})
}

func (s *playstoreSettingsService) GetServiceAccountPath(ctx context.Context, packageName string) (string, error) {
	account, err := s.repo.GetServiceAccount(ctx, s.db, packageName)
	if err != nil {
		return "", err
	}
	return account.FilePath, nil
}

func (s *playstoreSettingsService) getOldFilePath(ctx context.Context, tx *gorm.DB, packageName string) (string, error) {
	account, err := s.repo.GetServiceAccount(ctx, tx, packageName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return account.FilePath, nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
