// --------------------------
// 🧩 Repository: playstore user repository
// --------------------------
package repository

import (
	"context"
	"errors"
	"fmt"

	"subsnotifpro-go/internal/playstore/user/models"

	"gorm.io/gorm"
)

type PlaystoreUserRepository interface {
	FindByObfuscatedID(ctx context.Context, tx *gorm.DB, obfuscatedID string) (*models.GoogleAccount, error)
	UpsertGoogleAccount(
		ctx context.Context,
		tx *gorm.DB,
		googleAccount *models.GoogleAccount,
	) error
}

type playstoreUserRepository struct{}

func NewPlaystoreUserRepository() PlaystoreUserRepository {
	return &playstoreUserRepository{}
}

func (r *playstoreUserRepository) FindByObfuscatedID(ctx context.Context, tx *gorm.DB, obfuscatedID string) (*models.GoogleAccount, error) {
	var account models.GoogleAccount
	err := tx.WithContext(ctx).
		Where("obfuscated_external_account_id = ?", obfuscatedID).
		First(&account).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *playstoreUserRepository) UpsertGoogleAccount(
	ctx context.Context,
	tx *gorm.DB,
	googleAccount *models.GoogleAccount,
) error {
	if googleAccount.ObfuscatedExternalAccountID == "" {
		return fmt.Errorf("missing ObfuscatedExternalAccountID")
	}

	var existing models.GoogleAccount
	err := tx.WithContext(ctx).
		Where("obfuscated_external_account_id = ?", googleAccount.ObfuscatedExternalAccountID).
		First(&existing).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// ✅ No existing record: Create new
		return tx.WithContext(ctx).Create(googleAccount).Error
	}

	// ✅ Build update map only if values are different
	updates := map[string]interface{}{}

	if googleAccount.AppUserID != "" && googleAccount.AppUserID != existing.AppUserID {
		updates["app_user_id"] = googleAccount.AppUserID
	}
	if googleAccount.ExternalAccountID != nil && (existing.ExternalAccountID == nil || *googleAccount.ExternalAccountID != *existing.ExternalAccountID) {
		updates["external_account_id"] = *googleAccount.ExternalAccountID
	}
	if googleAccount.ObfuscatedExternalProfileID != nil && (existing.ObfuscatedExternalProfileID == nil || *googleAccount.ObfuscatedExternalProfileID != *existing.ObfuscatedExternalProfileID) {
		updates["obfuscated_external_profile_id"] = *googleAccount.ObfuscatedExternalProfileID
	}
	if googleAccount.ProfileID != nil && (existing.ProfileID == nil || *googleAccount.ProfileID != *existing.ProfileID) {
		updates["profile_id"] = *googleAccount.ProfileID
	}
	if googleAccount.ProfileName != nil && (existing.ProfileName == nil || *googleAccount.ProfileName != *existing.ProfileName) {
		updates["profile_name"] = *googleAccount.ProfileName
	}
	if googleAccount.EmailAddress != nil && (existing.EmailAddress == nil || *googleAccount.EmailAddress != *existing.EmailAddress) {
		updates["email_address"] = *googleAccount.EmailAddress
	}
	if googleAccount.GivenName != nil && (existing.GivenName == nil || *googleAccount.GivenName != *existing.GivenName) {
		updates["given_name"] = *googleAccount.GivenName
	}
	if googleAccount.FamilyName != nil && (existing.FamilyName == nil || *googleAccount.FamilyName != *existing.FamilyName) {
		updates["family_name"] = *googleAccount.FamilyName
	}

	if len(updates) > 0 {
		return tx.WithContext(ctx).Model(&existing).Updates(updates).Error
	}

	// ✅ No differences — skip update
	return nil
}
