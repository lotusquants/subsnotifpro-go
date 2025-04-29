package service

import (
	"context"
	"fmt"

	appstoreModels "subsnotifpro-go/internal/appstore/user/models"
	appstoreRepo "subsnotifpro-go/internal/appstore/user/repository"
	usersService "subsnotifpro-go/internal/users/service"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppStoreUserService interface {
	GetOrCreateUserIDFromAppAccountToken(
		ctx context.Context,
		tx *gorm.DB,
		appAccountToken string,
		appleAccount *appstoreModels.AppleAccount,
	) (uuid.UUID, error)
}

type appStoreUserService struct {
	userService      usersService.UserService
	appStoreUserRepo appstoreRepo.AppStoreUserRepository
}

func NewAppStoreUserService(
	userService usersService.UserService,

) AppStoreUserService {
	return &appStoreUserService{
		userService:      userService,
		appStoreUserRepo: appstoreRepo.NewAppStoreUserRepository(),
	}
}

func (s *appStoreUserService) GetOrCreateUserIDFromAppAccountToken(
	ctx context.Context,
	tx *gorm.DB,
	appAccountToken string,
	appleAccount *appstoreModels.AppleAccount,
) (uuid.UUID, error) {
	// 1. Get or create the AppUser using users service
	appUserID, err := s.userService.CreateAppUserIfNotExists(
		ctx,
		tx,
		appAccountToken,
		"APPLE_APPSTORE",
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get/create AppUser: %w", err)
	}

	// 2. Assign AppUserID to the AppleAccount model
	appleAccount.AppUserID = appUserID
	appleAccount.AppAccountToken = appAccountToken

	// 3. Save or update AppleAccount entry
	err = s.appStoreUserRepo.UpsertAppleAccount(ctx, tx, appleAccount)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert AppleAccount: %w", err)
	}

	return appUserID, nil
}
