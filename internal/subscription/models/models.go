package models

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type UnifiedSubscription struct {
	gorm.Model
	UserID         uuid.UUID `gorm:"type:uuid;index"`
	SubscriptionID uuid.UUID `gorm:"type:uuid;uniqueIndex"` // Subscription ID

	ActivePlatform PlatformType `gorm:"type:varchar(50);index"`  // APPLE_APPSTORE or GOOGLE_PLAYSTORE
	PlatformUserID *string      `gorm:"type:varchar(255);index"` // ObfuscatedExternalAccountID or AppAccountToken
	PurchaseToken  *string
	LatestOrderID  *string
	// Plan Type
	PlanType string `gorm:"type:varchar(20);index"`

	// Subscription State (Normalized)
	Status SubscriptionStatus `gorm:"type:varchar(50);index"`

	// Timing Information
	StartDate       time.Time `gorm:"index"`
	NextRenewalDate time.Time `gorm:"index"`
	// Expiration Information
	ExpirationDate time.Time `gorm:"index"`
	// Grace Period Information
	GracePeriodStartDate *time.Time `gorm:"index"`
	GracePeriodEndDate   *time.Time `gorm:"index"` // UTC version

	// Product Info
	ProductId string `gorm:"type:varchar(255);index"` // Product ID
	// Base Plan ID
	BasePlanID string `gorm:"type:varchar(255);index"` // Base plan ID
	// Add-ons
	AddOnID *string `gorm:"type:varchar(255);index"` // add-on ID
	// Offer Details
	ActiveOfferID *string `gorm:"type:varchar(255);index"`

	// Financials
	TotalAmount float64 `gorm:"type:decimal(10,2)"`
	Currency    string  `gorm:"type:varchar(3)"`
}

type PlatformType string

const (
	PlatformApple  PlatformType = "APPLE_APPSTORE"
	PlatformGoogle PlatformType = "GOOGLE_PLAYSTORE"
)

// Unified Subscription Status
type SubscriptionStatus string

const (
	StatusActive           SubscriptionStatus = "ACTIVE"
	StatusExpired          SubscriptionStatus = "EXPIRED"
	StatusBillingRetry     SubscriptionStatus = "BILLING_RETRY"
	StatusGracePeriod      SubscriptionStatus = "GRACE_PERIOD"
	StatusRevoked          SubscriptionStatus = "REVOKED"
	StatusPendingRenewal   SubscriptionStatus = "PENDING_RENEWAL"
	StatusPendingUpgrade   SubscriptionStatus = "PENDING_UPGRADE"
	StatusPendingDowngrade SubscriptionStatus = "PENDING_DOWNGRADE"
	StatusPaused           SubscriptionStatus = "PAUSED"
	StatusOnHold           SubscriptionStatus = "ON_HOLD"
	StatusCanceled         SubscriptionStatus = "CANCELED"
)
