// ---------------------------
// 🧩 Service: google playstore subscription service
// ---------------------------

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"subsnotifpro-go/internal/pkg/contextutil"
	"subsnotifpro-go/internal/pkg/logger"

	apiDto "subsnotifpro-go/internal/playstore/api/dto"
	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	playstoreCatalogService "subsnotifpro-go/internal/playstore/products/service"
	rtdnDto "subsnotifpro-go/internal/playstore/rtdn/dto"
	"subsnotifpro-go/internal/playstore/subscription/mapper"
	"subsnotifpro-go/internal/playstore/subscription/models"
	"subsnotifpro-go/internal/playstore/subscription/repository"
	playstoreUserService "subsnotifpro-go/internal/playstore/user/service"
)

type PlaystoreSubscriptionService interface {
	UpsertSubscription(ctx context.Context, subData *apiDto.SubscriptionPurchaseV2, event *rtdnDto.GooglePlayWebhookEvent) error
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
	subData *apiDto.SubscriptionPurchaseV2,
	event *rtdnDto.GooglePlayWebhookEvent,
) error {

	logger.Log.Infof("Processing subscription for purchase token %s", event.Subscription.PurchaseToken)

	// Get transaction from context instead of creating a new one
	tx, ok := contextutil.TxFromContext(ctx)
	if !ok {
		return fmt.Errorf("no transaction found in context - transaction must be started at rtdn consumer level")
	}

	// 1️⃣ Validate and extract essential fields
	purchaseToken := event.Subscription.PurchaseToken
	subscriptionProductId := event.Subscription.SubscriptionID
	notificationType := event.Subscription.NotificationType
	packageName := event.PackageName

	if purchaseToken == "" || packageName == "" || subscriptionProductId == "" {
		return fmt.Errorf("missing required fields: purchaseToken, packageName or subscriptionProductId")
	}

	// 2️⃣ Resolve AppUser (Resove ObfuscatedExternalAccountID)
	if subData.ExternalAccountIdentifiers.ObfuscatedExternalAccountID == nil {
		return fmt.Errorf("missing obfuscatedExternalAccountId")
	}

	obfuscatedID := *subData.ExternalAccountIdentifiers.ObfuscatedExternalAccountID
	if obfuscatedID == "" {
		return fmt.Errorf("missing obfuscatedExternalAccountId")
	}

	// Get or create App User
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

	// 5️⃣ Branch based on whether subscription exists
	if existing == nil {
		// 🆕 CREATE NEW SUBSCRIPTION PATH
		return s.createNewSubscription(ctx, tx, appUserID, packageName, purchaseToken, subData, subscriptionProductId, notificationType, event.ID)
	} else {
		// 🔁 UPDATE EXISTING SUBSCRIPTION PATH
		return s.updateExistingSubscription(ctx, tx, existing, subData, notificationType, event.ID)
	}

}

func (s *playstoreSubscriptionService) createNewSubscription(
	ctx context.Context,
	tx *gorm.DB,
	appUserID uuid.UUID,
	packageName string,
	purchaseToken string,
	subData *apiDto.SubscriptionPurchaseV2,
	subscriptionProductId string,
	notificationType rtdnDto.SubscriptionNotificationType,
	changeEventID uuid.UUID,
) error {
	// 1. Validate essential fields first
	if err := validateSubscriptionData(subData); err != nil {
		return fmt.Errorf("invalid subscription data: %w", err)
	}

	startTime := subData.StartTime

	newSubscriptionState := models.SubscriptionState(subData.SubscriptionState)
	newAckState := models.AcknowledgementState(subData.AcknowledgementState)

	// 3. Create minimal subscription record
	subscriptionID := uuid.New()
	newSub := models.SubscriptionPurchaseV2{
		ID:                   subscriptionID,
		PackageName:          packageName,
		PurchaseToken:        purchaseToken,
		UserID:               appUserID,
		StartTime:            startTime,
		LatestOrderID:        subData.LatestOrderID,
		SubscriptionState:    newSubscriptionState,
		AcknowledgementState: newAckState,
		RegionCode:           subData.RegionCode,
	}

	if err := s.repo.InsertSubscription(ctx, tx, &newSub); err != nil {
		return fmt.Errorf("failed to insert base subscription: %w", err)
	}

	// 4. Create change event
	err := s.repo.CreateSubscriptionEvent(ctx, tx, &models.SubscriptionEvent{
		SubscriptionID: subscriptionID,
		EventID:        changeEventID,
		EventType:      notificationType.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to create change event: %w", err)
	}

	// 5. Record initial state transitions
	if err := s.recordInitialStateTransitions(
		ctx, tx, subscriptionID, newSubscriptionState, newAckState,
		changeEventID, packageName, subscriptionProductId, purchaseToken,
	); err != nil {
		return err
	}

	// 6. Process linked purchase token if exists
	updateFields := make(map[string]interface{})
	if subData.LinkedPurchaseToken != nil {
		if err := s.processLinkedPurchaseToken(
			ctx, tx, subData, subscriptionID, purchaseToken,
			changeEventID, updateFields,
		); err != nil {
			return err
		}
	}

	// 7. Handle line items
	if _, err := s.ResolveLineItems(
		ctx, tx, &newSub, subData, changeEventID, notificationType,
	); err != nil {
		return fmt.Errorf("failed to resolve line items: %w", err)
	}

	// 8. Acknowledge with Play Store and update state
	if err := s.processAcknowledgement(
		ctx, tx, subscriptionID, newAckState, changeEventID,
		packageName, subscriptionProductId, purchaseToken, updateFields,
	); err != nil {
		return err
	}

	// 9. Final update if any fields need updating
	if len(updateFields) > 0 {
		if err := s.repo.UpdateSubscriptionFields(ctx, tx, subscriptionID, updateFields); err != nil {
			return fmt.Errorf("failed to update subscription relations: %w", err)
		}
	}

	return nil
}

// Helper functions for better organization:

func validateSubscriptionData(subData *apiDto.SubscriptionPurchaseV2) error {
	if subData.RegionCode == "" {
		return errors.New("region code cannot be empty")
	}
	if subData.SubscriptionState == "" {
		return errors.New("subscription state cannot be empty")
	}
	return nil
}

func (s *playstoreSubscriptionService) recordInitialStateTransitions(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	subState models.SubscriptionState,
	ackState models.AcknowledgementState,
	changeEventID uuid.UUID,
	packageName string,
	subscriptionProductId string,
	purchaseToken string,
) error {
	if err := s.RecordSubscriptionStateChange(
		ctx, tx, subscriptionID, nil, subState, changeEventID,
	); err != nil {
		return fmt.Errorf("failed to record subscription state change: %w", err)
	}

	if err := s.RecordAckStateTransitionForCreateBeforeAckApiCall(
		ctx, tx, subscriptionID, ackState, changeEventID,
		packageName, subscriptionProductId, purchaseToken,
	); err != nil {
		return fmt.Errorf("failed to record ack state change: %w", err)
	}
	return nil
}

func (s *playstoreSubscriptionService) processLinkedPurchaseToken(
	ctx context.Context,
	tx *gorm.DB,
	subData *apiDto.SubscriptionPurchaseV2,
	subscriptionID uuid.UUID,
	purchaseToken string,
	changeEventID uuid.UUID,
	updateFields map[string]interface{},
) error {
	// log.Printf("🔗 Resolving linked purchase token: %s", subData.LinkedPurchaseToken)
	linkedFromSubID, err := s.ResolveLinkedPurchaseToken(
		ctx, tx, subData, subscriptionID, purchaseToken, changeEventID,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve linked purchase token: %w", err)
	}
	updateFields["linked_from_subscription_id"] = linkedFromSubID
	updateFields["linked_purchase_token"] = subData.LinkedPurchaseToken
	return nil
}

func (s *playstoreSubscriptionService) processAcknowledgement(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	currentAckState models.AcknowledgementState,
	changeEventID uuid.UUID,
	packageName string,
	subscriptionProductId string,
	purchaseToken string,
	updateFields map[string]interface{},
) error {
	refreshedState, err := s.AcknowledgeSubscription(
		ctx, tx, subscriptionID, currentAckState,
		changeEventID, packageName, subscriptionProductId, purchaseToken,
	)
	if err != nil {
		return fmt.Errorf("failed to acknowledge subscription: %w", err)
	}

	if refreshedState != nil {
		updateFields["acknowledgement_state"] = refreshedState
	}

	return nil
}

func (s *playstoreSubscriptionService) updateExistingSubscription(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPurchaseV2,
	subData *apiDto.SubscriptionPurchaseV2,
	notificationType rtdnDto.SubscriptionNotificationType,
	changeEventID uuid.UUID,
) error {
	// // 1. Create change event first
	err := s.repo.CreateSubscriptionEvent(ctx, tx, &models.SubscriptionEvent{
		SubscriptionID: existing.ID,
		EventID:        changeEventID,
		EventType:      notificationType.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to create change event: %w", err)
	}

	// 3. Resolve all dependent models
	updateFields := make(map[string]interface{})

	updateFields["latest_order_id"] = subData.LatestOrderID
	updateFields["subscription_state"] = models.SubscriptionState(subData.SubscriptionState)

	// Handle state transitions
	if err := s.processStateTransitions(
		ctx, tx, existing, subData, changeEventID, updateFields,
	); err != nil {
		return err
	}

	// Handle context updates
	if err := s.processContextUpdates(
		ctx, tx, existing, subData, changeEventID, notificationType, updateFields,
	); err != nil {
		return err
	}

	// Handle linked purchase token
	if err := s.processLinkedPurchaseTokenUpdate(
		ctx, tx, existing, subData, changeEventID, updateFields,
	); err != nil {
		return err
	}

	// Handle line items
	if _, err := s.ResolveLineItems(
		ctx, tx, existing, subData, changeEventID, notificationType,
	); err != nil {
		return fmt.Errorf("failed to resolve line items: %w", err)
	}

	// 4. Perform final update if needed
	if len(updateFields) > 0 {
		if err := s.repo.UpdateSubscriptionFields(ctx, tx, existing.ID, updateFields); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	return nil
}

// Helper functions:

func (s *playstoreSubscriptionService) processStateTransitions(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPurchaseV2,
	subData *apiDto.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	updateFields map[string]interface{},
) error {
	if err := s.RecordSubscriptionStateChange(
		ctx, tx, existing.ID, existing, models.SubscriptionState(subData.SubscriptionState), changeEventID,
	); err != nil {
		return fmt.Errorf("failed to record subscription state change: %w", err)
	}

	// 2. Handle order ID transition
	newOrderID, err := s.RecordOrderIDChange(
		ctx, tx, existing.ID, existing, subData.LatestOrderID, changeEventID,
	)
	if err != nil {
		return fmt.Errorf("failed to record order ID change: %w", err)
	}

	// Update the order ID if changed
	if newOrderID != existing.LatestOrderID {
		updateFields["latest_order_id"] = newOrderID
	}

	return nil
}

func (s *playstoreSubscriptionService) processContextUpdates(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPurchaseV2,
	subData *apiDto.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	notificationType rtdnDto.SubscriptionNotificationType,
	updateFields map[string]interface{},
) error {
	// Paused context
	pausedCtxID, err := s.ResolvePausedContext(
		ctx, tx, existing.ID, existing,
		subData, changeEventID, notificationType,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve paused context: %w", err)
	}
	updateFields["subscription_paused_context_id"] = pausedCtxID

	// Cancellation context
	cancelCtxID, err := s.ResolveCancellationContext(
		ctx, tx, existing.ID, existing,
		subData, changeEventID, notificationType,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve cancellation context: %w", err)
	}
	updateFields["subscription_cancellation_context_id"] = cancelCtxID

	return nil
}

func (s *playstoreSubscriptionService) processLinkedPurchaseTokenUpdate(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPurchaseV2,
	subData *apiDto.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	updateFields map[string]interface{},
) error {
	if subData.LinkedPurchaseToken == nil {
		return nil
	}

	// log.Printf("🔗 Linked purchase token detected: %s", subData.LinkedPurchaseToken)
	linkedFromSubID, err := s.ResolveLinkedPurchaseToken(
		ctx, tx, subData, existing.ID, existing.PurchaseToken, changeEventID,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve linked purchase token: %w", err)
	}

	updateFields["linked_from_subscription_id"] = linkedFromSubID
	updateFields["linked_purchase_token"] = subData.LinkedPurchaseToken
	return nil
}
