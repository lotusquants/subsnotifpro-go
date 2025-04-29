package events

import (
	"time"

	"subsnotifpro-go/internal/subscription/models"

	"github.com/google/uuid"
)

// UnifiedEvent represents the normalized subscription event
type UnifiedEvent struct {
	EventID         uuid.UUID                 `json:"event_id"`
	EventType       *string                   `json:"event_type"` // "subscription.created", "subscription.updated"
	Timestamp       time.Time                 `json:"timestamp"`
	UserID          uuid.UUID                 `json:"user_id"`
	SubscriptionID  uuid.UUID                 `json:"subscription_id"`
	Platform        models.PlatformType       `json:"platform"`
	PlatformDetails PlatformDetails           `json:"platform_details"`
	Status          models.SubscriptionStatus `json:"status"`
	ProductInfo     ProductInfo               `json:"product_info"`
	Timing          TimingInfo                `json:"timing"`
	Financials      FinancialInfo             `json:"financials"`
}

type PlatformDetails struct {
	PlatformUserID *string `json:"platform_user_id"`
	PurchaseToken  *string `json:"purchase_token,omitempty"`  // Purchase token for p[lays tore/ original txn id for appstore]
	LatestOrderID  *string `json:"latest_order_id,omitempty"` // orderid/txn id

}

type ProductInfo struct {
	ProductID     string  `json:"product_id"`
	BasePlanID    string  `json:"base_plan_id"`
	AddOnID       *string `json:"add_on_id,omitempty"`
	ActiveOfferID *string `json:"active_offer_id,omitempty"`
	PlanType      string  `json:"plan_type"`
}

type TimingInfo struct {
	StartDate            time.Time  `json:"start_date"`
	NextRenewalDate      time.Time  `json:"next_renewal_date"`
	ExpirationDate       time.Time  `json:"expiration_date"`
	GracePeriodStartDate *time.Time `json:"grace_period_start_date,omitempty"`
	GracePeriodEndDate   *time.Time `json:"grace_period_end_date,omitempty"`
}

type FinancialInfo struct {
	Currency string `json:"currency"`
}
