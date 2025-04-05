package models

import (
	"subsnotifpro-go/internal/playstore/api/dto"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
)

type GooglePublishPayload struct {
	ID    uuid.UUID `json:"event_id"`
	Event *GooglePlayWebhookEvent

	SubscriptionPurchase *dto.SubscriptionPurchaseV2
	OneTimePurchase      *androidpublisher.ProductPurchase
	VoidedPurchase       *androidpublisher.VoidedPurchase
	TestPurchase         *androidpublisher.TestPurchase
}
