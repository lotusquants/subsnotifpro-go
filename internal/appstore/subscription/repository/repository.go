package repository

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/appstore/subscription/models"
	"subsnotifpro-go/internal/appstore/webhooks/dto"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type AppStoreSubscriptionRepository interface {
	FindByOriginalTransactionID(ctx context.Context, tx *gorm.DB, originalTransactionID string) (*models.AppStoreSubscription, error)
	Create(ctx context.Context, tx *gorm.DB, subscription *models.AppStoreSubscription) error
	Update(ctx context.Context, tx *gorm.DB, subscription *models.AppStoreSubscription) error
	FindOrCreate(ctx context.Context, tx *gorm.DB, data *dto.Data, userID *uuid.UUID) (*models.AppStoreSubscription, error)
}

type appStoreSubscriptionRepository struct{}

func NewAppStoreSubscriptionRepository() AppStoreSubscriptionRepository {
	return &appStoreSubscriptionRepository{}
}

func (r *appStoreSubscriptionRepository) FindByOriginalTransactionID(
	ctx context.Context,
	tx *gorm.DB,
	originalTransactionID string,
) (*models.AppStoreSubscription, error) {
	var subscription models.AppStoreSubscription
	err := tx.WithContext(ctx).
		Where("original_transaction_id = ?", originalTransactionID).
		First(&subscription).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}
	return &subscription, nil
}

func (r *appStoreSubscriptionRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	subscription *models.AppStoreSubscription,
) error {
	if err := tx.WithContext(ctx).Create(subscription).Error; err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

func (r *appStoreSubscriptionRepository) Update(
	ctx context.Context,
	tx *gorm.DB,
	subscription *models.AppStoreSubscription,
) error {
	if err := tx.WithContext(ctx).Save(subscription).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}

func (r *appStoreSubscriptionRepository) FindOrCreate(
	ctx context.Context,
	tx *gorm.DB,
	data *dto.Data,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	// Try to find existing subscription
	var subscription models.AppStoreSubscription
	err := tx.WithContext(ctx).
		Where("original_transaction_id = ?", data.SignedTransactionInfo.JWSTransactionDecodedPayload.OriginalTransactionId).
		First(&subscription).Error

	if err == nil {
		return &subscription, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to query subscription: %w", err)
	}

	// Create new subscription
	subscription = models.AppStoreSubscription{
		OriginalTransactionID: data.SignedTransactionInfo.JWSTransactionDecodedPayload.OriginalTransactionId,
		CurrentTransactionID:  data.SignedTransactionInfo.JWSTransactionDecodedPayload.TransactionId,
		WebOrderLineItemID:    data.SignedTransactionInfo.JWSTransactionDecodedPayload.WebOrderLineItemId,
		ProductID:             data.SignedTransactionInfo.JWSTransactionDecodedPayload.ProductId,
		SubscriptionGroupID:   data.SignedTransactionInfo.JWSTransactionDecodedPayload.SubscriptionGroupIdentifier,
		Status:                models.SubscriptionStatusActive,
		AutoRenewStatus:       models.AutoRenewOn,
		Environment:           models.Environment(data.SignedTransactionInfo.JWSTransactionDecodedPayload.Environment),
		OriginalPurchaseDate:  data.SignedTransactionInfo.JWSTransactionDecodedPayload.OriginalPurchaseDate,
		PurchaseDate:          data.SignedTransactionInfo.JWSTransactionDecodedPayload.PurchaseDate,
		ExpiresDate:           data.SignedTransactionInfo.JWSTransactionDecodedPayload.ExpiresDate,
		InAppOwnershipType:    models.InAppOwnershipType(data.SignedTransactionInfo.JWSTransactionDecodedPayload.InAppOwnershipType),
		Currency:              data.SignedTransactionInfo.JWSTransactionDecodedPayload.Currency,
		Price:                 data.SignedTransactionInfo.JWSTransactionDecodedPayload.Price,
		CountryCode:           data.SignedTransactionInfo.JWSTransactionDecodedPayload.Storefront,
	}

	if userID != nil {
		subscription.UserID = *userID
	}

	if data.SignedTransactionInfo.JWSTransactionDecodedPayload.AppAccountToken != nil {
		subscription.AppAccountToken = data.SignedTransactionInfo.JWSTransactionDecodedPayload.AppAccountToken
	}

	if data.SignedRenewalInfo != nil {
		subscription.OfferType = &data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload.OfferType
		subscription.OfferIdentifier = &data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload.OfferIdentifier
		subscription.OfferDuration = &data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload.OfferPeriod
	}

	if err := tx.WithContext(ctx).Create(&subscription).Error; err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return &subscription, nil
}
