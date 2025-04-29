package service

import (
	"context"
	"fmt"
	"time"

	appStoreModels "subsnotifpro-go/internal/appstore/subscription/models"
	"subsnotifpro-go/internal/pkg/logger"
	playStoreModels "subsnotifpro-go/internal/playstore/subscription/models"
	"subsnotifpro-go/internal/subscription/events"
	"subsnotifpro-go/internal/subscription/mapper"
	"subsnotifpro-go/internal/subscription/models"
	"subsnotifpro-go/internal/subscription/publisher"
	"subsnotifpro-go/internal/subscription/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UnifiedSubscriptionService interface
type UnifiedSubscriptionService interface {
	CreateUnifiedSubscriptionFromAppStore(ctx context.Context, sub *appStoreModels.AppStoreSubscription, eventType *string) error
	CreateUnifiedSubscriptionFromPlayStore(ctx context.Context, sub *playStoreModels.SubscriptionPurchaseV2, eventType *string) error

	ProcessUnifiedSubscriptionEvent(ctx context.Context, event events.UnifiedEvent) error
	UpsertUnifiedSubscription(ctx context.Context, sub *models.UnifiedSubscription) error
}

type unifiedSubscriptionService struct {
	publisher    *publisher.UnifiedEventPublisher
	statusMapper *mapper.StatusMapper
	repo         repository.SubscriptionRepository
	dashboardSvc DashboardService
}

func NewUnifiedSubscriptionService(db *gorm.DB, publisher *publisher.UnifiedEventPublisher, dashboardSvc DashboardService, repo repository.SubscriptionRepository) UnifiedSubscriptionService {
	return &unifiedSubscriptionService{
		publisher:    publisher,
		statusMapper: mapper.NewStatusMapper(),
		repo:         repo,
		dashboardSvc: dashboardSvc,
	}
}

func (s *unifiedSubscriptionService) CreateUnifiedSubscriptionFromAppStore(
	ctx context.Context,
	sub *appStoreModels.AppStoreSubscription,
	eventType *string,
) error {
	// Map status to unified model
	status := s.statusMapper.MapAppleStatus(sub.Status)

	// Build unified event
	event := events.UnifiedEvent{
		EventID:        uuid.New(),
		EventType:      eventType,
		Timestamp:      time.Now().UTC(),
		UserID:         sub.UserID,
		SubscriptionID: sub.ID,
		Platform:       models.PlatformApple,
		PlatformDetails: events.PlatformDetails{
			PlatformUserID: sub.AppAccountToken,
			LatestOrderID:  &sub.CurrentTransactionID,
			PurchaseToken:  &sub.OriginalTransactionID,
		},
		Status: status,
		ProductInfo: events.ProductInfo{
			ProductID:  sub.ProductID,
			BasePlanID: sub.ProductID,

			ActiveOfferID: sub.OfferIdentifier,
			PlanType:      "AUTO_RENEWING",
		},
		Timing: events.TimingInfo{
			StartDate:            sub.OriginalPurchaseDate,
			NextRenewalDate:      sub.ExpiresDate,
			ExpirationDate:       sub.ExpiresDate,
			GracePeriodStartDate: &sub.ExpiresDate,
			GracePeriodEndDate:   sub.GracePeriodExpiresDate,
		},
		Financials: events.FinancialInfo{
			Currency: sub.Currency,
		},
	}

	if err := s.publisher.Publish(ctx, event); err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"subscription_id": sub.ID,
			"platform":        "appstore",
		}).Error("Failed to publish unified subscription event")
		return err
	}

	logger.Log.WithFields(map[string]interface{}{
		"event_id":        event.EventID,
		"subscription_id": sub.ID,
		"user_id":         sub.UserID,
		"status":          status,
	}).Info("Published unified AppStore subscription event")

	return nil
}

func (s *unifiedSubscriptionService) CreateUnifiedSubscriptionFromPlayStore(
	ctx context.Context,
	sub *playStoreModels.SubscriptionPurchaseV2,
	eventType *string,
) error {

	logger.Log.Infof("Creating unified subscription from PlayStore event: %v", sub)

	if sub == nil {
		return fmt.Errorf("subscription cannot be nil")
	}

	if eventType == nil {
		return fmt.Errorf("eventType cannot be nil")
	}

	// First validate the subscription has line items
	if len(sub.LineItems) == 0 {
		return fmt.Errorf("subscription has no line items")
	}

	firstLineItem := sub.LineItems[0]

	// Validate required fields
	if firstLineItem.ProductID == "" {
		return fmt.Errorf("missing product ID in line item")
	}

	// Map status to unified model
	status := s.statusMapper.MapGoogleStatus(sub.SubscriptionState)

	// Initialize event with safe defaults
	event := events.UnifiedEvent{
		EventID:        uuid.New(),
		EventType:      eventType,
		Timestamp:      time.Now().UTC(),
		UserID:         sub.UserID,
		SubscriptionID: sub.ID,
		Platform:       models.PlatformGoogle,
		Status:         status,
		ProductInfo: events.ProductInfo{
			ProductID: firstLineItem.ProductID,
			PlanType:  string(firstLineItem.PlanType),
		},
	}

	logger.Log.Infof("user details unified play store : %v", sub.User)

	// Safely add platform details
	if sub.User != nil && sub.User.GoogleAccount != nil {
		event.PlatformDetails.PlatformUserID = &sub.User.GoogleAccount.ObfuscatedExternalAccountID
		event.PlatformDetails.PurchaseToken = &sub.PurchaseToken
		event.PlatformDetails.LatestOrderID = &sub.LatestOrderID
	}
	event.PlatformDetails.PurchaseToken = &sub.PurchaseToken

	// Safely add offer details if available
	if firstLineItem.OfferDetails != nil {
		event.ProductInfo.BasePlanID = firstLineItem.OfferDetails.BasePlanID
		event.ProductInfo.ActiveOfferID = firstLineItem.OfferDetails.OfferID

		if firstLineItem.OfferDetails.BasePlanPrice != (playStoreModels.Money{}) {
			event.Financials.Currency = firstLineItem.OfferDetails.BasePlanPrice.CurrencyCode
		}
	}

	// Safely add timing info if available
	if firstLineItem.AutoRenewingPlan != nil {
		event.Timing.NextRenewalDate = firstLineItem.ExpiryTime
		event.Timing.ExpirationDate = firstLineItem.ExpiryTime
	}
	event.Timing.StartDate = sub.StartTime

	logger.Log.WithFields(map[string]interface{}{
		"subscription_id":    sub.ID,
		"user_id":            sub.UserID,
		"event_type":         *eventType,
		"line_items":         len(sub.LineItems),
		"has_user":           sub.User != nil,
		"has_google_account": sub.User != nil && sub.User.GoogleAccount != nil,
	}).Info("Publishing unified subscription event")

	// Publish event
	if err := s.publisher.Publish(ctx, event); err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"subscription_id": sub.ID,
			"platform":        "playstore",
		}).Error("Failed to publish unified subscription event")
		return err
	}

	logger.Log.WithFields(map[string]interface{}{
		"event_id":        event.EventID,
		"subscription_id": sub.ID,
		"user_id":         sub.UserID,
		"status":          status,
	}).Info("Published unified PlayStore subscription event")

	return nil
}

func (s *unifiedSubscriptionService) ProcessUnifiedSubscriptionEvent(
	ctx context.Context,
	event events.UnifiedEvent,
) error {

	logger.Log.Info("Processing unified subscription event", event)
	// Convert event to unified subscription model
	subscription := &models.UnifiedSubscription{
		UserID:               event.UserID,
		SubscriptionID:       event.SubscriptionID,
		ActivePlatform:       models.PlatformType(event.Platform),
		PlatformUserID:       event.PlatformDetails.PlatformUserID,
		LatestOrderID:        event.PlatformDetails.LatestOrderID,
		PurchaseToken:        event.PlatformDetails.PurchaseToken,
		PlanType:             event.ProductInfo.PlanType,
		Status:               models.SubscriptionStatus(event.Status),
		StartDate:            event.Timing.StartDate,
		NextRenewalDate:      event.Timing.NextRenewalDate,
		ExpirationDate:       event.Timing.ExpirationDate,
		GracePeriodStartDate: event.Timing.GracePeriodStartDate,
		GracePeriodEndDate:   event.Timing.GracePeriodEndDate,
		ProductId:            event.ProductInfo.ProductID,
		BasePlanID:           event.ProductInfo.BasePlanID,
		AddOnID:              event.ProductInfo.AddOnID,
		ActiveOfferID:        event.ProductInfo.ActiveOfferID,
		Currency:             event.Financials.Currency,
	}

	return s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.UpsertUnifiedSubscription(txCtx, subscription); err != nil {
			return err
		}

		// Refresh dashboard through the dashboard service
		return s.dashboardSvc.RefreshDashboard(txCtx)
	})
}

func (s *unifiedSubscriptionService) UpsertUnifiedSubscription(
	ctx context.Context,
	sub *models.UnifiedSubscription,
) error {
	return s.repo.UpsertSubscription(ctx, sub)
}
