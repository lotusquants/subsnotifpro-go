// ---------------------------
// 🧩 Service: google playstore subscription service
// ---------------------------

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	playstoreCatalogService "subsnotifpro-go/internal/playstore/products/service"
	rtdnModels "subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/subscription/mapper"
	"subsnotifpro-go/internal/playstore/subscription/models"
	"subsnotifpro-go/internal/playstore/subscription/repository"
	playstoreUserService "subsnotifpro-go/internal/playstore/user/service"

	"google.golang.org/api/androidpublisher/v3"
)

type PlaystoreSubscriptionService interface {
	UpsertSubscription(ctx context.Context, subData *androidpublisher.SubscriptionPurchaseV2, event rtdnModels.GooglePlayWebhookEvent) error
}

type playstoreSubscriptionService struct {
	db                      *gorm.DB
	repo                    repository.PlaystoreSubscriptionRepository
	playstoreUserService    playstoreUserService.PlaystoreUserService
	playstoreApiService     playstoreApiService.PlaystoreApiService
	playstoreCatalogService playstoreCatalogService.SubscriptionCatalogService
}

func NewPlaystoreSubscriptionService(
	db *gorm.DB,
	repo repository.PlaystoreSubscriptionRepository,
	playstoreUserService playstoreUserService.PlaystoreUserService,
	playstoreApiService playstoreApiService.PlaystoreApiService,
	playstoreCatalogService playstoreCatalogService.SubscriptionCatalogService,

) PlaystoreSubscriptionService {
	return &playstoreSubscriptionService{
		db:                      db,
		repo:                    repo,
		playstoreUserService:    playstoreUserService,
		playstoreApiService:     playstoreApiService,
		playstoreCatalogService: playstoreCatalogService,
	}
}

func (s *playstoreSubscriptionService) UpsertSubscription(
	ctx context.Context,
	subData *androidpublisher.SubscriptionPurchaseV2,
	event rtdnModels.GooglePlayWebhookEvent,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1️⃣ Extract essentials
		purchaseToken := event.SubscriptionNotification.PurchaseToken
		subscriptionProductId := event.SubscriptionNotification.SubscriptionID
		notificationType := event.SubscriptionNotification.NotificationType
		packageName := event.PackageName

		if purchaseToken == "" || packageName == "" || subscriptionProductId == "" {
			return fmt.Errorf("missing purchaseToken or packageName or subscriptionProductId")
		}

		obfuscatedID := subData.ExternalAccountIdentifiers.ObfuscatedExternalAccountId
		if obfuscatedID == "" {
			return fmt.Errorf("missing obfuscatedExternalAccountId")
		}

		// 2️⃣ Resolve AppUser
		appUserID, err := s.playstoreUserService.GetOrCreateUserIDFromObfuscatedExternalAccountID(
			ctx, tx, obfuscatedID, mapper.BuildGoogleAccountModel(subData),
		)
		if err != nil {
			return fmt.Errorf("failed to resolve AppUser: %w", err)
		}

		// 3️⃣ Check for existing subscription
		existing, err := s.repo.GetSubscriptionByPurchaseToken(ctx, tx, purchaseToken)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check subscription existence: %w", err)
		}

		var existingSubID *uuid.UUID
		if existing != nil {
			id := existing.ID
			existingSubID = &id

		}

		var (
			subscriptionID uuid.UUID
			changeEventID  uuid.UUID
		)

		if existing != nil {
			// 🔁 Existing subscription: parse ID and create change event
			subscriptionID = existing.ID

			changeEventID, err = s.repo.CreateSubscriptionChangeEvent(ctx, tx, &models.SubscriptionChangeEvent{
				SubscriptionID:   subscriptionID,
				NotificationType: notificationType,
			})
			if err != nil {
				return fmt.Errorf("failed to create change event: %w", err)
			}
		} else {
			// 🆕 New subscription: generate ID, and create change event *before insert*
			subscriptionID = uuid.New()

			changeEventID, err = s.repo.CreateSubscriptionChangeEvent(ctx, tx, &models.SubscriptionChangeEvent{
				SubscriptionID: subscriptionID,
			})
			if err != nil {
				return fmt.Errorf("failed to create change event: %w", err)
			}
		}

		// 4️⃣ Resolve all dependent models with conditional change tracking
		regionID, err := s.repo.GetOrCreateRegionCodeID(ctx, tx, subData.RegionCode)
		if err != nil {
			return fmt.Errorf("failed to resolve region: %w", err)
		}

		subStateID, err := s.ResolveSubscriptionState(ctx, tx, *existingSubID, existing, models.SubscriptionState(subData.SubscriptionState), changeEventID)
		if err != nil {
			return fmt.Errorf("failed to resolve subscription state: %w", err)
		}

		latestOrderID, err := s.ResolveOrderID(ctx, tx, subscriptionID, existing, subData.LatestOrderId, changeEventID)
		if err != nil {
			return fmt.Errorf("failed to resolve order ID: %w", err)
		}

		// 🧩 If this subscription replaces another, handle the linked (old) subscription
		var linkedFromSubID *uuid.UUID

		if subData.LinkedPurchaseToken != "" {
			log.Printf("🔗 Linked purchase token detected: %s. Attempting to resolve and retire previous subscription...", subData.LinkedPurchaseToken)

			linkedFromSubID, err = s.ResolveLinkedPurchaseToken(ctx, tx, subData,
				subscriptionID,
				purchaseToken,
				changeEventID)
			if err != nil {
				return fmt.Errorf("failed to resolve linked purchase token: %w", err)
			}
		}

		ackStateID, err := s.ResolveAcknowledgementState(ctx, tx, existingSubID, existing, subData.AcknowledgementState, changeEventID, subscriptionProductId)
		if err != nil {
			return fmt.Errorf("failed to resolve acknowledgement state: %w", err)
		}

		pausedCtxID, err := s.ResolvePausedContext(ctx, tx, existingSubID, existing, subData, changeEventID, notificationType)
		if err != nil {
			return fmt.Errorf("failed to resolve paused context: %w", err)
		}

		cancelCtxID, err := s.ResolveCancellationContext(ctx, tx, existingSubID, existing, subData, changeEventID, notificationType)
		if err != nil {
			return fmt.Errorf("failed to resolve cancellation context: %w", err)
		}

		LineItems, err := s.ResolveLineItems(ctx, tx, existingSubID, existing, subData, changeEventID, notificationType)
		if err != nil {
			return fmt.Errorf("failed to resolve plan models: %w", err)
		}

		startTime, err := time.Parse(time.RFC3339, subData.StartTime)
		if err != nil {
			return fmt.Errorf("invalid startTime format: %w", err)
		}
		if existing != nil {
			// 🔁 Update existing
			updateFields, err := mapper.BuildSubscriptionUpdateFields(ctx, existing, &mapper.SubscriptionUpdateParams{
				AppUserID:                   appUserID,
				RegionID:                    regionID,
				SubscriptionStateModelID:    subStateID,
				AcknowledgementStateModelID: ackStateID,
				PausedContextID:             pausedCtxID,
				CancellationContextID:       cancelCtxID,
				LineItems:                   LineItems,
				StartTime:                   startTime,
				LatestOrderId:               latestOrderID,
				PackageName:                 packageName,
				LinkedFromSubscriptionID:    linkedFromSubID,
			})
			if err != nil {
				return fmt.Errorf("failed to build update fields: %w", err)
			}
			if len(updateFields) > 0 {
				if err := s.repo.UpdateSubscriptionFields(ctx, tx, existing.ID, updateFields); err != nil {
					return fmt.Errorf("failed to update subscription: %w", err)
				}
			}
		} else {
			// 🆕 Insert new
			newSub := models.SubscriptionPurchaseV2{
				ID:                                subscriptionID,
				PackageName:                       packageName,
				PurchaseToken:                     purchaseToken,
				UserID:                            appUserID,
				RegionCodeID:                      regionID,
				SubscriptionStateModelID:          subStateID,
				AcknowledgementStateModelID:       ackStateID,
				SubscriptionPausedContextID:       pausedCtxID,
				SubscriptionCancellationContextID: cancelCtxID,
				LineItems:                         LineItems,
				StartTime:                         startTime,
				LatestOrderId:                     latestOrderID,
				LinkedFromSubscriptionID:          linkedFromSubID,
			}
			if err := s.repo.InsertSubscription(ctx, tx, &newSub); err != nil {
				return fmt.Errorf("failed to insert subscription: %w", err)
			}
		}

		return nil
	})
}
