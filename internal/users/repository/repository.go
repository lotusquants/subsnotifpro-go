// 📁 internal/users/repository/user_repository.go
package repository

import (
	"context"
	"errors"

	"subsnotifpro-go/internal/users/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateAppUserIfNotExists(ctx context.Context, tx *gorm.DB, obfuscatedID, platform string) (uuid.UUID, error)
	LogPlatformChange(ctx context.Context, tx *gorm.DB, userID uuid.UUID, oldPlatform, newPlatform string) error
}

type userRepository struct{}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

// ✅ Create or update AppUser with platform change detection
func (r *userRepository) CreateAppUserIfNotExists(ctx context.Context, tx *gorm.DB, obfuscatedID, platform string) (uuid.UUID, error) {
	var user models.AppUser
	err := tx.WithContext(ctx).
		Where("app_user_id = ?", obfuscatedID).
		First(&user).
		Error

	if err == nil {
		if user.ActivePlatform != platform {
			// 🔁 Update platform and log transition
			err := r.LogPlatformChange(ctx, tx, user.ID, user.ActivePlatform, platform)
			if err != nil {
				return uuid.Nil, err
			}

			if err := tx.WithContext(ctx).Model(&user).
				Update("active_platform", platform).Error; err != nil {
				return uuid.Nil, err
			}
		}
		return user.ID, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}

	// 🆕 Create new user
	newUser := models.AppUser{
		AppUserID:      obfuscatedID,
		ActivePlatform: platform,
	}
	if err := tx.WithContext(ctx).Create(&newUser).Error; err != nil {
		return uuid.Nil, err
	}
	return newUser.ID, nil
}

// 📦 Log platform transition
func (r *userRepository) LogPlatformChange(ctx context.Context, tx *gorm.DB, userID uuid.UUID, oldPlatform, newPlatform string) error {
	entry := models.AppUserPlatformChange{
		UserID:      userID,
		OldPlatform: oldPlatform,
		NewPlatform: newPlatform,
	}
	return tx.WithContext(ctx).Create(&entry).Error
}
