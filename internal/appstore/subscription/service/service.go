package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/appstore/subscription/models"
	appStoreSubscriptionRepo "subsnotifpro-go/internal/appstore/subscription/repository"
	appStoreUserService "subsnotifpro-go/internal/appstore/user/service"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/pkg/contextutil"
	"subsnotifpro-go/internal/pkg/logger"

	unifiedSubscriptionService "subsnotifpro-go/internal/subscription/service"

	"gorm.io/gorm"
)

// AppStoreSubscriptionService handles App Store subscription business logic
type AppStoreSubscriptionService interface {
	ProcessSubscriptionNotification(ctx context.Context, notification *dto.AppStoreNotification) error
}

type appstoreSubscriptionService struct {
	db                         *gorm.DB
	repo                       appStoreSubscriptionRepo.AppStoreSubscriptionRepository
	consumptionRepo            appStoreSubscriptionRepo.ConsumptionRequestRepository
	appStoreUserService        appStoreUserService.AppStoreUserService
	unifiedSubscriptionService unifiedSubscriptionService.UnifiedSubscriptionService
}

func NewAppStoreSubscriptionService(db *gorm.DB,
	appStoreUserService appStoreUserService.AppStoreUserService,
	unifiedSubscriptionService unifiedSubscriptionService.UnifiedSubscriptionService,

) AppStoreSubscriptionService {
	return &appstoreSubscriptionService{
		db:                         db,
		repo:                       appStoreSubscriptionRepo.NewAppStoreSubscriptionRepository(),
		consumptionRepo:            appStoreSubscriptionRepo.NewConsumptionRequestRepository(db),
		appStoreUserService:        appStoreUserService,
		unifiedSubscriptionService: unifiedSubscriptionService,
	}
}

func (s *appstoreSubscriptionService) ProcessSubscriptionNotification(
	ctx context.Context,
	notification *dto.AppStoreNotification,
) error {
	logger.Log.Infof("Processing app store subscription for transaction ID %s", notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload.TransactionId)

	// Get transaction from context instead of creating a new one
	tx, ok := contextutil.TxFromContext(ctx)
	if !ok {
		return fmt.Errorf("no transaction found in context - transaction must be started at rtdn consumer level")
	}

	// Resolve user if AppAccountToken exists
	userID, err := s.ResolveAppUser(ctx, tx, notification)
	if err != nil {
		return fmt.Errorf("failed to resolve app user: %w", err)
	}

	// Process notification type
	var sub *models.AppStoreSubscription

	// Process based on notification type
	switch notification.ResponseBodyV2DecodedPayload.NotificationType {
	case dto.SUBSCRIBED:
		sub, err = s.handleSubscribed(ctx, tx, notification, userID)
	case dto.DID_RENEW:
		sub, err = s.handleDidRenew(ctx, tx, notification, userID)
	case dto.DID_CHANGE_RENEWAL_STATUS:
		sub, err = s.handleDidChangeRenewalStatus(ctx, tx, notification, userID)
	case dto.DID_CHANGE_RENEWAL_PREF:
		sub, err = s.handleDidChangeRenewalPref(ctx, tx, notification, userID)
	case dto.DID_FAIL_TO_RENEW:
		sub, err = s.handleDidFailToRenew(ctx, tx, notification, userID)
	case dto.EXPIRED:
		sub, err = s.handleExpired(ctx, tx, notification, userID)
	case dto.REFUND:
		sub, err = s.handleRefund(ctx, tx, notification, userID)
	case dto.REVOKE:
		sub, err = s.handleRevoke(ctx, tx, notification, userID)
	case dto.CONSUMPTION_REQUEST:
		sub, err = s.handleConsumptionRequest(ctx, tx, notification, userID)
	case dto.EXTERNAL_PURCHASE_TOKEN:
		sub, err = s.handleExternalPurchaseToken(ctx, tx, notification, userID)
	case dto.GRACE_PERIOD_EXPIRED:
		sub, err = s.handleGracePeriodExpired(ctx, tx, notification, userID)
	case dto.METADATA_UPDATE:
		sub, err = s.handleMetadataUpdate(ctx, tx, notification, userID)
	case dto.MIGRATION:
		sub, err = s.handleMigration(ctx, tx, notification, userID)
	case dto.OFFER_REDEEMED:
		sub, err = s.handleOfferRedeemed(ctx, tx, notification, userID)
	case dto.ONE_TIME_CHARGE:
		sub, err = s.handleOneTimeCharge(ctx, tx, notification, userID)
	case dto.PRICE_CHANGE:
		sub, err = s.handlePriceChange(ctx, tx, notification, userID)
	case dto.PRICE_INCREASE:
		sub, err = s.handlePriceIncrease(ctx, tx, notification, userID)
	case dto.REFUND_DECLINED:
		sub, err = s.handleRefundDeclined(ctx, tx, notification, userID)
	case dto.REFUND_REVERSED:
		sub, err = s.handleRefundReversed(ctx, tx, notification, userID)
	case dto.RENEWAL_EXTENDED:
		sub, err = s.handleRenewalExtended(ctx, tx, notification, userID)
	case dto.RENEWAL_EXTENSION:
		sub, err = s.handleRenewalExtension(ctx, tx, notification, userID)
	case dto.TEST:
		logger.Log.Info("Received test notification")
		return nil
	default:
		return fmt.Errorf("unhandled notification type: %s",
			notification.ResponseBodyV2DecodedPayload.NotificationType)
	}

	if err != nil {
		return err
	}

	logger.Log.Info("Processed the appstore subscription and delegating to unified subscription module")

	// Pass raw AppStore data to unified service
	notificationType := string(notification.ResponseBodyV2DecodedPayload.NotificationType)
	if err := s.unifiedSubscriptionService.CreateUnifiedSubscriptionFromAppStore(ctx, sub, &notificationType); err != nil {
		return fmt.Errorf("failed to process unified event: %w", err)
	}

	return nil

}
