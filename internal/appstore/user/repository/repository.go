package repository

import (
	"context"
	"errors"
	"fmt"

	"subsnotifpro-go/internal/appstore/user/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppStoreUserRepository interface {
	FindByAppAccountToken(ctx context.Context, tx *gorm.DB, token uuid.UUID) (*models.AppleAccount, error)
	UpsertAppleAccount(
		ctx context.Context,
		tx *gorm.DB,
		appleAccount *models.AppleAccount,
	) error
}

type appStoreUserRepository struct{}

func NewAppStoreUserRepository() AppStoreUserRepository {
	return &appStoreUserRepository{}
}

func (r *appStoreUserRepository) FindByAppAccountToken(ctx context.Context, tx *gorm.DB, token uuid.UUID) (*models.AppleAccount, error) {
	var account models.AppleAccount
	err := tx.WithContext(ctx).
		Where("app_account_token = ?", token).
		First(&account).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *appStoreUserRepository) UpsertAppleAccount(
	ctx context.Context,
	tx *gorm.DB,
	appleAccount *models.AppleAccount,
) error {
	if appleAccount.AppAccountToken == "" {
		return fmt.Errorf("missing AppAccountToken")
	}

	var existing models.AppleAccount
	err := tx.WithContext(ctx).
		Where("app_account_token = ?", appleAccount.AppAccountToken).
		First(&existing).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new record
		return tx.WithContext(ctx).Create(appleAccount).Error
	}

	// Build update map only if values are different
	updates := map[string]interface{}{}

	if appleAccount.AppUserID != uuid.Nil && appleAccount.AppUserID != existing.AppUserID {
		updates["app_user_id"] = appleAccount.AppUserID
	}

	if appleAccount.BundleID != nil && (existing.BundleID == nil || *appleAccount.BundleID != *existing.BundleID) {
		updates["bundle_id"] = *appleAccount.BundleID
	}

	if len(updates) > 0 {
		return tx.WithContext(ctx).Model(&existing).Updates(updates).Error
	}

	// No differences — skip update
	return nil
}
