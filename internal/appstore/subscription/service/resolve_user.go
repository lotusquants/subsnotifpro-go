package service

import (
	"context"
	"fmt"
	appStoreUserModels "subsnotifpro-go/internal/appstore/user/models"
	"subsnotifpro-go/internal/appstore/webhooks/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// resolveAppUser handles user resolution from AppAccountToken
func (s *appstoreSubscriptionService) ResolveAppUser(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
) (*uuid.UUID, error) {
	data := notification.ResponseBodyV2DecodedPayload.Data
	if data == nil || data.SignedTransactionInfo.JWSTransactionDecodedPayload.AppAccountToken == nil {
		return nil, nil
	}

	appAccountToken := *data.SignedTransactionInfo.JWSTransactionDecodedPayload.AppAccountToken
	appleAccount := &appStoreUserModels.AppleAccount{

		BundleID: &data.SignedTransactionInfo.JWSTransactionDecodedPayload.BundleId,
	}

	userID, err := s.appStoreUserService.GetOrCreateUserIDFromAppAccountToken(
		ctx,
		tx,
		appAccountToken,
		appleAccount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user from app account token: %w", err)
	}

	return &userID, nil
}
