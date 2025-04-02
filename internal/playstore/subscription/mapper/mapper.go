package mapper

import (
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
