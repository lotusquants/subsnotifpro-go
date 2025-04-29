package mapper

import (
	"subsnotifpro-go/internal/playstore/api/dto"
	"subsnotifpro-go/internal/playstore/subscription/models"
	playstoreModels "subsnotifpro-go/internal/playstore/user/models"
	"time"

	"github.com/google/uuid"
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

func BuildGoogleAccountModel(subData *dto.SubscriptionPurchaseV2) *playstoreModels.GoogleAccount {
	var external *dto.ExternalAccountIdentifiers
	var subscribeInfo *dto.SubscribeWithGoogleInfo

	if subData != nil {
		external = &subData.ExternalAccountIdentifiers
		subscribeInfo = subData.SubscribeWithGoogleInfo
	}

	if external == nil || external.ObfuscatedExternalAccountID == nil {
		return nil
	}

	account := &playstoreModels.GoogleAccount{
		ID:                          uuid.New(),
		ObfuscatedExternalAccountID: *external.ObfuscatedExternalAccountID,
		ExternalAccountID:           external.ExternalAccountID,
		ObfuscatedExternalProfileID: external.ObfuscatedExternalAccountID,
	}

	if subscribeInfo != nil {
		account.ProfileID = subscribeInfo.ProfileID
		account.ProfileName = subscribeInfo.ProfileName
		account.EmailAddress = subscribeInfo.EmailAddress
		account.GivenName = subscribeInfo.GivenName
		account.FamilyName = subscribeInfo.FamilyName
	}

	return account
}
