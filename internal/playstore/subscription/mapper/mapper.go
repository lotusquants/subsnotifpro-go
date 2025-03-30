package mapper

import (
	"context"
	"subsnotifpro-go/internal/playstore/subscription/models"
	playstoreModels "subsnotifpro-go/internal/playstore/user/models"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
)

type SubscriptionUpdateParams struct {
	AppUserID                   uuid.UUID
	RegionID                    uuid.UUID
	SubscriptionStateModelID    uuid.UUID
	AcknowledgementStateModelID uuid.UUID
	PausedContextID             *uuid.UUID
	CancellationContextID       *uuid.UUID
	PlanType                    models.PlanType
	AutoRenewingPlanID          *uuid.UUID
	PrepaidPlanID               *uuid.UUID
	StartTime                   time.Time
	LatestOrderId               string
	PackageName                 string
	LinkedFromSubscriptionID    *uuid.UUID
	LineItems                   []models.SubscriptionLineItem
}

func BuildGoogleAccountModel(subData *androidpublisher.SubscriptionPurchaseV2) *playstoreModels.GoogleAccount {
	var external *androidpublisher.ExternalAccountIdentifiers
	var subscribeInfo *androidpublisher.SubscribeWithGoogleInfo

	if subData != nil {
		external = subData.ExternalAccountIdentifiers
		subscribeInfo = subData.SubscribeWithGoogleInfo
	}

	if external == nil {
		return nil // Cannot continue without obfuscatedExternalAccountId
	}

	account := &playstoreModels.GoogleAccount{
		ObfuscatedExternalAccountID: external.ObfuscatedExternalAccountId,
		ExternalAccountID:           nullableString(external.ExternalAccountId),
		ObfuscatedExternalProfileID: nullableString(external.ObfuscatedExternalProfileId),
	}

	// Add optional fields from SubscribeWithGoogleInfo if present
	if subscribeInfo != nil {
		account.ProfileID = nullableString(subscribeInfo.ProfileId)
		account.ProfileName = nullableString(subscribeInfo.ProfileName)
		account.EmailAddress = nullableString(subscribeInfo.EmailAddress)
		account.GivenName = nullableString(subscribeInfo.GivenName)
		account.FamilyName = nullableString(subscribeInfo.FamilyName)
	}

	return account
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// buildSubscriptionUpdateFields compares current vs new values and builds a map of changed fields for update.
func BuildSubscriptionUpdateFields(
	ctx context.Context,
	existing *models.SubscriptionPurchaseV2,
	updates *SubscriptionUpdateParams,
) (map[string]interface{}, error) {
	fields := make(map[string]interface{})

	// UUID fields
	if existing.UserID != updates.AppUserID {
		fields["user_id"] = updates.AppUserID
	}
	if existing.RegionCodeID != updates.RegionID {
		fields["region_code_id"] = updates.RegionID
	}
	if existing.SubscriptionStateModelID != updates.SubscriptionStateModelID {
		fields["subscription_state_model_id"] = updates.SubscriptionStateModelID
	}
	if existing.AcknowledgementStateModelID != updates.AcknowledgementStateModelID {
		fields["acknowledgement_state_model_id"] = updates.AcknowledgementStateModelID
	}
	if !uuidPtrEqual(existing.SubscriptionPausedContextID, updates.PausedContextID) {
		fields["subscription_paused_context_id"] = updates.PausedContextID
	}
	if !uuidPtrEqual(existing.SubscriptionCancellationContextID, updates.CancellationContextID) {
		fields["subscription_cancellation_context_id"] = updates.CancellationContextID
	}

	// Timestamps
	if !existing.StartTime.Equal(updates.StartTime) {
		fields["start_time"] = updates.StartTime
	}

	// Strings
	if existing.LatestOrderId != updates.LatestOrderId {
		fields["latest_order_id"] = updates.LatestOrderId
	}
	if existing.PackageName != updates.PackageName {
		fields["package_name"] = updates.PackageName
	}

	if existing.LinkedFromSubscriptionID != updates.LinkedFromSubscriptionID {
		fields["linked_from_subscription_id"] = updates.LinkedFromSubscriptionID
	}

	// Add line items comparison if needed
	if updates.LineItems != nil {
		// This assumes you want to replace all line items
		// Adjust logic if you need more sophisticated comparison
		fields["line_items"] = updates.LineItems
	}

	return fields, nil
}

func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && b != nil {
		return *a == *b
	}
	return false
}
