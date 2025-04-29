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
	if googleAccount == nil {
		return uuid.Nil, fmt.Errorf("google account cannot be nil")
	}

	// 1. Ensure ID is set
	if googleAccount.ID == uuid.Nil {
		googleAccount.ID = uuid.New()
	}

	// 2. Get or create AppUser
	appUserID, err := s.userService.CreateAppUserIfNotExists(ctx, tx, obfuscatedID, "GOOGLE_PLAYSTORE")
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get/create AppUser: %w", err)
	}

	// 3. Set AppUserID on GoogleAccount
	googleAccount.AppUserID = appUserID.String()

	// 4. FIRST save GoogleAccount and ensure it's committed
	persistedAccount, err := s.playstoreUserRepo.UpsertGoogleAccount(ctx, tx, googleAccount)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert GoogleAccount: %w", err)
	}

	// 5. Verify the GoogleAccount exists before linking
	exists, err := s.playstoreUserRepo.GoogleAccountExists(ctx, tx, persistedAccount.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to verify GoogleAccount: %w", err)
	}
	if !exists {
		return uuid.Nil, fmt.Errorf("GoogleAccount %s does not exist", persistedAccount.ID)
	}

	// 6. Now link the accounts
	if err := s.userService.LinkGoogleAccount(ctx, tx, appUserID, persistedAccount.ID); err != nil {
		return uuid.Nil, fmt.Errorf("failed to link GoogleAccount to user: %w", err)
	}

	return appUserID, nil
}
