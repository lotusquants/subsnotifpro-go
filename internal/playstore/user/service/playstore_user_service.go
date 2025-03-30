package service

import (
	"context"
	"fmt"

	playstoreModels "subsnotifpro-go/internal/playstore/user/models"
	playstoreRepo "subsnotifpro-go/internal/playstore/user/repository"
	usersService "subsnotifpro-go/internal/users/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlaystoreUserService interface {
	GetOrCreateUserIDFromObfuscatedExternalAccountID(ctx context.Context, tx *gorm.DB, obfuscatedID string, googleAccount *playstoreModels.GoogleAccount) (uuid.UUID, error)
}

type playstoreUserService struct {
	userService       usersService.UserService
	playstoreUserRepo playstoreRepo.PlaystoreUserRepository
}

func NewPlaystoreUserService(userService usersService.UserService, playstoreUserRepo playstoreRepo.PlaystoreUserRepository) PlaystoreUserService {
	return &playstoreUserService{
		userService:       userService,
		playstoreUserRepo: playstoreUserRepo,
	}
}

// GetOrCreateUserIDFromObfuscatedExternalAccountID ensures an AppUser exists and associates a GoogleAccount with it.
func (s *playstoreUserService) GetOrCreateUserIDFromObfuscatedExternalAccountID(
	ctx context.Context,
	tx *gorm.DB,
	obfuscatedID string,
	googleAccount *playstoreModels.GoogleAccount,
) (uuid.UUID, error) {
	// 1️⃣ Get or create the AppUser using users service
	appUserID, err := s.userService.CreateAppUserIfNotExists(ctx, tx, obfuscatedID, "GOOGLE_PLAYSTORE")
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get/create AppUser: %w", err)
	}

	// 2️⃣ Assign AppUserID to the GoogleAccount model
	googleAccount.AppUserID = appUserID.String()

	// 3️⃣ Save or update GoogleAccount entry
	err = s.playstoreUserRepo.UpsertGoogleAccount(ctx, tx, googleAccount)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert GoogleAccount: %w", err)
	}

	// ✅ Done
	return appUserID, nil
}
